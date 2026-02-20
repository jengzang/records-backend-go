package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// DirectionalBiasAnalyzer implements directional movement pattern analysis
// Skill: 方向偏好分析 (Directional Bias)
// Analyzes heading distribution and identifies preferred directions
type DirectionalBiasAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewDirectionalBiasAnalyzer creates a new directional bias analyzer
func NewDirectionalBiasAnalyzer(db *sql.DB) analysis.Analyzer {
	return &DirectionalBiasAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "directional_bias", 10000),
	}
}

// Analyze performs directional bias analysis
func (a *DirectionalBiasAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[DirectionalBiasAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing stats (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM directional_stats"); err != nil {
			return fmt.Errorf("failed to clear directional_stats: %w", err)
		}
		log.Printf("[DirectionalBiasAnalyzer] Cleared existing directional stats")
	}

	// Query track points with heading data
	query := `
		SELECT
			heading, distance, dataTime
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND heading IS NOT NULL
			AND heading >= 0
			AND heading < 360
			AND distance IS NOT NULL
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query track points: %w", err)
	}
	defer rows.Close()

	// Initialize direction buckets (8 directions: N, NE, E, SE, S, SW, W, NW)
	buckets := make([]DirectionBucket, 8)
	for i := 0; i < 8; i++ {
		buckets[i].Bucket = i
	}

	totalPoints := 0
	totalDistance := 0.0
	var prevTime int64

	for rows.Next() {
		var heading, distance float64
		var dataTime int64

		if err := rows.Scan(&heading, &distance, &dataTime); err != nil {
			return fmt.Errorf("failed to scan track point: %w", err)
		}

		totalPoints++
		totalDistance += distance

		// Calculate direction bucket (0=N, 1=NE, 2=E, ..., 7=NW)
		bucket := a.headingToBucket(heading)
		buckets[bucket].Distance += distance
		buckets[bucket].PointCount++

		if prevTime > 0 {
			duration := dataTime - prevTime
			buckets[bucket].Duration += duration
		}

		prevTime = dataTime
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	// Calculate percentages
	for i := range buckets {
		if totalDistance > 0 {
			buckets[i].Percentage = (buckets[i].Distance / totalDistance) * 100
		}
	}

	log.Printf("[DirectionalBiasAnalyzer] Processed %d points, total distance: %.2f m", totalPoints, totalDistance)

	// Update task progress
	if err := a.UpdateTaskProgress(taskID, int64(totalPoints), int64(totalPoints), 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Insert directional stats
	if err := a.insertDirectionalStats(ctx, buckets); err != nil {
		return fmt.Errorf("failed to insert directional stats: %w", err)
	}

	// Find dominant direction
	dominantBucket := 0
	maxPercentage := 0.0
	for i, bucket := range buckets {
		if bucket.Percentage > maxPercentage {
			maxPercentage = bucket.Percentage
			dominantBucket = i
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points":      totalPoints,
		"total_distance":    totalDistance,
		"dominant_direction": a.bucketToDirection(dominantBucket),
		"dominant_percent":   maxPercentage,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[DirectionalBiasAnalyzer] Analysis completed")
	return nil
}

// DirectionBucket holds statistics for a direction bucket
type DirectionBucket struct {
	Bucket     int
	Distance   float64
	Duration   int64
	PointCount int64
	Percentage float64
}

// headingToBucket converts heading (0-360) to bucket (0-7)
// 0=N, 1=NE, 2=E, 3=SE, 4=S, 5=SW, 6=W, 7=NW
func (a *DirectionalBiasAnalyzer) headingToBucket(heading float64) int {
	// Normalize heading to 0-360
	heading = math.Mod(heading, 360)
	if heading < 0 {
		heading += 360
	}

	// Each bucket covers 45 degrees
	// N: 337.5-22.5, NE: 22.5-67.5, E: 67.5-112.5, etc.
	bucket := int((heading + 22.5) / 45.0)
	if bucket >= 8 {
		bucket = 0
	}

	return bucket
}

// bucketToDirection converts bucket number to direction name
func (a *DirectionalBiasAnalyzer) bucketToDirection(bucket int) string {
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	if bucket >= 0 && bucket < 8 {
		return directions[bucket]
	}
	return "UNKNOWN"
}

// insertDirectionalStats inserts directional stats into the database
func (a *DirectionalBiasAnalyzer) insertDirectionalStats(ctx context.Context, buckets []DirectionBucket) error {
	if len(buckets) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO directional_stats (
			metric_date, direction_bucket, distance_m, duration_s,
			point_count, percentage,
			algo_version, created_at
		) VALUES (NULL, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, bucket := range buckets {
		_, err := stmt.ExecContext(ctx,
			bucket.Bucket, bucket.Distance, bucket.Duration,
			bucket.PointCount, bucket.Percentage,
		)
		if err != nil {
			return fmt.Errorf("failed to insert directional stat: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[DirectionalBiasAnalyzer] Inserted %d directional stats", len(buckets))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("directional_bias", NewDirectionalBiasAnalyzer)
}