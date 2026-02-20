package temporal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// TimeSpaceCompressionAnalyzer implements trajectory compression
// Skill: 时空压缩 (Time-Space Compression)
// Compresses trajectory data using Douglas-Peucker algorithm
type TimeSpaceCompressionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewTimeSpaceCompressionAnalyzer creates a new time-space compression analyzer
func NewTimeSpaceCompressionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &TimeSpaceCompressionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "time_space_compression", 10000),
	}
}

// Analyze performs trajectory compression
func (a *TimeSpaceCompressionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[TimeSpaceCompressionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing compressed trajectories (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM compressed_trajectories"); err != nil {
			return fmt.Errorf("failed to clear compressed_trajectories: %w", err)
		}
		log.Printf("[TimeSpaceCompressionAnalyzer] Cleared existing compressed trajectories")
	}

	// Load trajectory points
	points, err := a.loadTrajectoryPoints(ctx)
	if err != nil {
		return fmt.Errorf("failed to load trajectory points: %w", err)
	}

	if len(points) == 0 {
		log.Printf("[TimeSpaceCompressionAnalyzer] No points to compress")
		return a.MarkTaskAsCompleted(taskID, `{"compressed_trajectories": 0}`)
	}

	log.Printf("[TimeSpaceCompressionAnalyzer] Loaded %d points", len(points))

	// Compress trajectory using Douglas-Peucker algorithm
	// Try different epsilon values
	epsilons := []float64{0.0001, 0.0005, 0.001, 0.005} // degrees (~10m, ~50m, ~100m, ~500m)

	var compressedTrajectories []CompressedTrajectory

	for _, epsilon := range epsilons {
		compressed := a.douglasPeucker(points, epsilon)
		compressionRatio := float64(len(compressed)) / float64(len(points))

		// Convert to JSON
		pointsJSON, err := json.Marshal(compressed)
		if err != nil {
			return fmt.Errorf("failed to marshal compressed points: %w", err)
		}

		compressedTrajectories = append(compressedTrajectories, CompressedTrajectory{
			CompressionType:      "DOUGLAS_PEUCKER",
			Epsilon:              epsilon,
			OriginalPointCount:   len(points),
			CompressedPointCount: len(compressed),
			CompressionRatio:     compressionRatio,
			PointsJSON:           string(pointsJSON),
			StartTS:              points[0].DataTime,
			EndTS:                points[len(points)-1].DataTime,
		})

		log.Printf("[TimeSpaceCompressionAnalyzer] Epsilon %.4f: %d -> %d points (%.2f%%)",
			epsilon, len(points), len(compressed), compressionRatio*100)
	}

	// Insert compressed trajectories
	if err := a.insertCompressedTrajectories(ctx, compressedTrajectories); err != nil {
		return fmt.Errorf("failed to insert compressed trajectories: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"original_points":         len(points),
		"compressed_trajectories": len(compressedTrajectories),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[TimeSpaceCompressionAnalyzer] Analysis completed")
	return nil
}

// TrajectoryPoint holds trajectory point data
type TrajectoryPoint struct {
	ID        int64   `json:"id"`
	DataTime  int64   `json:"ts"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// CompressedTrajectory holds compressed trajectory data
type CompressedTrajectory struct {
	CompressionType      string
	Epsilon              float64
	OriginalPointCount   int
	CompressedPointCount int
	CompressionRatio     float64
	PointsJSON           string
	StartTS              int64
	EndTS                int64
}

// loadTrajectoryPoints loads trajectory points from database
func (a *TimeSpaceCompressionAnalyzer) loadTrajectoryPoints(ctx context.Context) ([]TrajectoryPoint, error) {
	query := `
		SELECT
			id, dataTime, latitude, longitude
		FROM "一生足迹"
		WHERE outlier_flag = 0
		ORDER BY dataTime
		LIMIT 50000
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query trajectory points: %w", err)
	}
	defer rows.Close()

	var points []TrajectoryPoint
	for rows.Next() {
		var point TrajectoryPoint
		if err := rows.Scan(&point.ID, &point.DataTime, &point.Latitude, &point.Longitude); err != nil {
			return nil, fmt.Errorf("failed to scan trajectory point: %w", err)
		}
		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return points, nil
}

// douglasPeucker implements the Douglas-Peucker line simplification algorithm
func (a *TimeSpaceCompressionAnalyzer) douglasPeucker(points []TrajectoryPoint, epsilon float64) []TrajectoryPoint {
	if len(points) < 3 {
		return points
	}

	// Find the point with the maximum distance from the line segment
	maxDist := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		dist := a.perpendicularDistance(
			points[i],
			points[0],
			points[len(points)-1],
		)
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}

	// If max distance is greater than epsilon, recursively simplify
	if maxDist > epsilon {
		// Recursive call
		left := a.douglasPeucker(points[:maxIndex+1], epsilon)
		right := a.douglasPeucker(points[maxIndex:], epsilon)

		// Combine results (remove duplicate middle point)
		result := append(left[:len(left)-1], right...)
		return result
	}

	// If max distance is less than epsilon, return endpoints only
	return []TrajectoryPoint{points[0], points[len(points)-1]}
}

// perpendicularDistance calculates perpendicular distance from point to line segment
func (a *TimeSpaceCompressionAnalyzer) perpendicularDistance(point, lineStart, lineEnd TrajectoryPoint) float64 {
	// Calculate distance using cross product method
	x := point.Longitude
	y := point.Latitude
	x1 := lineStart.Longitude
	y1 := lineStart.Latitude
	x2 := lineEnd.Longitude
	y2 := lineEnd.Latitude

	// Calculate line length
	lineLength := math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))

	if lineLength == 0 {
		// Line start and end are the same point
		return math.Sqrt((x-x1)*(x-x1) + (y-y1)*(y-y1))
	}

	// Calculate perpendicular distance
	distance := math.Abs((y2-y1)*x - (x2-x1)*y + x2*y1 - y2*x1) / lineLength

	return distance
}

// insertCompressedTrajectories inserts compressed trajectories into the database
func (a *TimeSpaceCompressionAnalyzer) insertCompressedTrajectories(ctx context.Context, trajectories []CompressedTrajectory) error {
	if len(trajectories) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO compressed_trajectories (
			compression_type, epsilon,
			original_point_count, compressed_point_count, compression_ratio,
			points_json, start_ts, end_ts,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, traj := range trajectories {
		_, err := stmt.ExecContext(ctx,
			traj.CompressionType, traj.Epsilon,
			traj.OriginalPointCount, traj.CompressedPointCount, traj.CompressionRatio,
			traj.PointsJSON, traj.StartTS, traj.EndTS,
		)
		if err != nil {
			return fmt.Errorf("failed to insert compressed trajectory: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[TimeSpaceCompressionAnalyzer] Inserted %d compressed trajectories", len(trajectories))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("time_space_compression", NewTimeSpaceCompressionAnalyzer)
}