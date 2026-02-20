package foundation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// TrajectoryCompletionAnalyzer implements trajectory completion
// Skill: 轨迹补全 (Trajectory Completion)
// Fills gaps in trajectory using linear interpolation
type TrajectoryCompletionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewTrajectoryCompletionAnalyzer creates a new trajectory completion analyzer
func NewTrajectoryCompletionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &TrajectoryCompletionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "trajectory_completion", 5000),
	}
}

// Analyze performs trajectory completion
func (a *TrajectoryCompletionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[TrajectoryCompletionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Remove existing interpolated points (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM \"一生足迹\" WHERE qa_status = 'interpolated'"); err != nil {
			return fmt.Errorf("failed to remove interpolated points: %w", err)
		}
		log.Printf("[TrajectoryCompletionAnalyzer] Removed existing interpolated points")
	}

	// Get all track points ordered by time
	pointsQuery := `
		SELECT
			id,
			dataTime,
			latitude,
			longitude,
			altitude,
			speed
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND (qa_status IS NULL OR qa_status != 'interpolated')
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, pointsQuery)
	if err != nil {
		return fmt.Errorf("failed to query points: %w", err)
	}
	defer rows.Close()

	var points []TrajectoryPoint
	for rows.Next() {
		var point TrajectoryPoint
		var altitude, speed sql.NullFloat64

		if err := rows.Scan(&point.ID, &point.Timestamp, &point.Lat, &point.Lon, &altitude, &speed); err != nil {
			return fmt.Errorf("failed to scan point: %w", err)
		}

		if altitude.Valid {
			point.Altitude = altitude.Float64
		}
		if speed.Valid {
			point.Speed = speed.Float64
		}

		points = append(points, point)
	}

	if len(points) < 2 {
		log.Printf("[TrajectoryCompletionAnalyzer] Not enough points to process")
		return a.MarkTaskAsCompleted(taskID, `{"interpolated_points": 0}`)
	}

	log.Printf("[TrajectoryCompletionAnalyzer] Processing %d points", len(points))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(points)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Detect gaps and interpolate
	gapThreshold := int64(300)  // 5 minutes
	maxGap := int64(1800)       // 30 minutes
	interpolatedPoints := a.detectAndInterpolate(points, gapThreshold, maxGap)

	// Insert interpolated points
	if err := a.insertInterpolatedPoints(ctx, interpolatedPoints); err != nil {
		return fmt.Errorf("failed to insert interpolated points: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points":        len(points),
		"interpolated_points": len(interpolatedPoints),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[TrajectoryCompletionAnalyzer] Analysis completed: %d points processed, %d points interpolated", len(points), len(interpolatedPoints))
	return nil
}

// InterpolatedPoint holds interpolated point data
type InterpolatedPoint struct {
	Timestamp int64
	Lat       float64
	Lon       float64
	Altitude  float64
	Speed     float64
}

// detectAndInterpolate detects gaps and creates interpolated points
func (a *TrajectoryCompletionAnalyzer) detectAndInterpolate(points []TrajectoryPoint, gapThreshold, maxGap int64) []InterpolatedPoint {
	var interpolated []InterpolatedPoint

	for i := 0; i < len(points)-1; i++ {
		p1 := points[i]
		p2 := points[i+1]

		timeDiff := p2.Timestamp - p1.Timestamp

		// Check if gap exists and is within max gap
		if timeDiff > gapThreshold && timeDiff <= maxGap {
			// Calculate number of points to interpolate (one point every 60 seconds)
			numPoints := int(timeDiff / 60)
			if numPoints > 30 {
				numPoints = 30 // Limit to 30 points per gap
			}

			// Linear interpolation
			for j := 1; j <= numPoints; j++ {
				ratio := float64(j) / float64(numPoints+1)
				timestamp := p1.Timestamp + int64(float64(timeDiff)*ratio)

				interpolated = append(interpolated, InterpolatedPoint{
					Timestamp: timestamp,
					Lat:       p1.Lat + (p2.Lat-p1.Lat)*ratio,
					Lon:       p1.Lon + (p2.Lon-p1.Lon)*ratio,
					Altitude:  p1.Altitude + (p2.Altitude-p1.Altitude)*ratio,
					Speed:     p1.Speed + (p2.Speed-p1.Speed)*ratio,
				})
			}
		}
	}

	return interpolated
}

// insertInterpolatedPoints inserts interpolated points into the database
func (a *TrajectoryCompletionAnalyzer) insertInterpolatedPoints(ctx context.Context, points []InterpolatedPoint) error {
	if len(points) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO "一生足迹" (
			dataTime, latitude, longitude, altitude, speed,
			qa_status, outlier_flag
		) VALUES (?, ?, ?, ?, ?, 'interpolated', 0)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, point := range points {
		_, err := stmt.ExecContext(ctx,
			point.Timestamp,
			point.Lat,
			point.Lon,
			point.Altitude,
			point.Speed,
		)
		if err != nil {
			return fmt.Errorf("failed to insert interpolated point: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[TrajectoryCompletionAnalyzer] Inserted %d interpolated points", len(points))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("trajectory_completion", NewTrajectoryCompletionAnalyzer)
}
