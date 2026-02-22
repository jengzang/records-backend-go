package advanced

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// MovementIntensityAnalyzer implements time-space compression analysis
// Skill: 27_movement_intensity (Time-Space Compression)
// Analyzes movement intensity and time-space efficiency patterns
type MovementIntensityAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewMovementIntensityAnalyzer creates a new time-space compression analyzer
func NewMovementIntensityAnalyzer(db *sql.DB) analysis.Analyzer {
	return &MovementIntensityAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "movement_intensity", 10000),
	}
}

// Analyze performs time-space compression analysis
func (a *MovementIntensityAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[MovementIntensityAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing stats (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM time_space_compression_bucketed"); err != nil {
			return fmt.Errorf("failed to clear time_space_compression_bucketed: %w", err)
		}
		log.Printf("[MovementIntensityAnalyzer] Cleared existing time-space compression stats")
	}

	// Process global stats only (segments table doesn't have admin columns)
	totalRecords := 0

	// Global stats (ALL)
	if err := a.processCompressionStats(ctx, "ALL", ""); err != nil {
		return fmt.Errorf("failed to process global stats: %w", err)
	}
	totalRecords++

	// Mark task as completed
	summary := map[string]interface{}{
		"total_records": totalRecords,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[MovementIntensityAnalyzer] Analysis completed: %d records generated", totalRecords)
	return nil
}

// processCompressionStats processes time-space compression statistics for a specific area
func (a *MovementIntensityAnalyzer) processCompressionStats(ctx context.Context, areaType, areaKey string) error {
	// Query segments with movement data
	query := `
		SELECT
			start_time,
			end_time,
			duration_s,
			distance_m,
			avg_speed_kmh,
			max_speed_kmh,
			mode
		FROM segments
		WHERE duration_s > 0
		  AND distance_m > 0
	`
	args := []interface{}{}

	if areaType != "ALL" {
		query += " AND " + areaType + " = ?"
		args = append(args, areaKey)
	}

	query += " ORDER BY start_time"

	rows, err := a.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	var segments []SegmentData
	for rows.Next() {
		var seg SegmentData
		var mode sql.NullString

		if err := rows.Scan(
			&seg.StartTime, &seg.EndTime, &seg.Duration,
			&seg.Distance, &seg.AvgSpeed, &seg.MaxSpeed, &mode,
		); err != nil {
			return fmt.Errorf("failed to scan segment: %w", err)
		}

		if mode.Valid {
			seg.Mode = mode.String
		}

		segments = append(segments, seg)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	if len(segments) == 0 {
		log.Printf("[MovementIntensityAnalyzer] No segment data for %s/%s", areaType, areaKey)
		return nil
	}

	// Calculate compression statistics
	stats := calculateCompressionStats(segments)

	// Insert into database
	if err := a.insertCompressionStats(ctx, "all", "", areaType, areaKey, stats); err != nil {
		return fmt.Errorf("failed to insert compression stats: %w", err)
	}

	return nil
}

// SegmentData holds segment information
type SegmentData struct {
	StartTime int64
	EndTime   int64
	Duration  int64
	Distance  float64
	AvgSpeed  float64
	MaxSpeed  float64
	Mode      string
}

// CompressionStats holds time-space compression statistics
type CompressionStats struct {
	MovementIntensity       float64
	BurstIntensity          float64
	BurstCount              int
	BurstDuration           int64
	ActiveTime              int64
	InactiveTime            int64
	ActivityRatio           float64
	EffectiveMovementRatio  float64
	AvgSpeedKmh             float64
	MaxSpeedKmh             float64
	DistancePerDay          float64
	TimeCompressionIndex    float64
	TotalDistance           float64
	TotalDuration           int64
	TripCount               int
	DistinctDays            int
}

// calculateCompressionStats calculates time-space compression statistics
func calculateCompressionStats(segments []SegmentData) CompressionStats {
	if len(segments) == 0 {
		return CompressionStats{}
	}

	// Speed threshold for "active" movement (km/h)
	const activeSpeedThreshold = 5.0
	const burstSpeedThreshold = 50.0 // High-speed burst threshold

	var totalDistance float64
	var totalDuration int64
	var activeTime int64
	var maxSpeed float64
	var burstCount int
	var burstDuration int64
	var inBurst bool

	// Track distinct days
	daySet := make(map[string]bool)

	for _, seg := range segments {
		totalDistance += seg.Distance
		totalDuration += seg.Duration

		// Track distinct days
		dayKey := fmt.Sprintf("%d", seg.StartTime/(86400))
		daySet[dayKey] = true

		// Count active time (speed > threshold)
		if seg.AvgSpeed > activeSpeedThreshold {
			activeTime += seg.Duration
		}

		// Track max speed
		if seg.MaxSpeed > maxSpeed {
			maxSpeed = seg.MaxSpeed
		}

		// Detect burst periods (high-speed movement)
		if seg.AvgSpeed > burstSpeedThreshold {
			if !inBurst {
				burstCount++
				inBurst = true
			}
			burstDuration += seg.Duration
		} else {
			inBurst = false
		}
	}

	inactiveTime := totalDuration - activeTime
	distinctDays := len(daySet)

	// Calculate metrics
	movementIntensity := 0.0
	if totalDuration > 0 {
		movementIntensity = (totalDistance / 1000.0) / (float64(totalDuration) / 3600.0) // km/h
	}

	activityRatio := 0.0
	if totalDuration > 0 {
		activityRatio = float64(activeTime) / float64(totalDuration)
	}

	effectiveMovementRatio := 1.0 // Assume all distance is effective

	avgSpeedKmh := 0.0
	if activeTime > 0 {
		avgSpeedKmh = (totalDistance / 1000.0) / (float64(activeTime) / 3600.0)
	}

	distancePerDay := 0.0
	if distinctDays > 0 {
		distancePerDay = (totalDistance / 1000.0) / float64(distinctDays)
	}

	// Calculate burst intensity (average speed during burst periods)
	burstIntensity := 0.0
	if burstDuration > 0 {
		// Estimate distance during bursts (simplified)
		burstIntensity = burstSpeedThreshold * 1.5 // Approximate
	}

	// Time compression index: composite metric
	// Higher values indicate more efficient time-space usage
	timeCompressionIndex := movementIntensity * activityRatio * math.Log(1+float64(distinctDays))

	return CompressionStats{
		MovementIntensity:      movementIntensity,
		BurstIntensity:         burstIntensity,
		BurstCount:             burstCount,
		BurstDuration:          burstDuration,
		ActiveTime:             activeTime,
		InactiveTime:           inactiveTime,
		ActivityRatio:          activityRatio,
		EffectiveMovementRatio: effectiveMovementRatio,
		AvgSpeedKmh:            avgSpeedKmh,
		MaxSpeedKmh:            maxSpeed,
		DistancePerDay:         distancePerDay,
		TimeCompressionIndex:   timeCompressionIndex,
		TotalDistance:          totalDistance,
		TotalDuration:          totalDuration,
		TripCount:              len(segments),
		DistinctDays:           distinctDays,
	}
}

// insertCompressionStats inserts compression statistics into the database
func (a *MovementIntensityAnalyzer) insertCompressionStats(
	ctx context.Context,
	bucketType, bucketKey, areaType, areaKey string,
	stats CompressionStats,
) error {
	query := `
		INSERT INTO time_space_compression_bucketed (
			bucket_type, bucket_key, area_type, area_key,
			movement_intensity, burst_intensity, burst_count, burst_duration_s,
			active_time_s, inactive_time_s, activity_ratio, effective_movement_ratio,
			avg_speed_kmh, max_speed_kmh, distance_per_day, time_compression_index,
			total_distance_m, total_duration_s, trip_count, distinct_days,
			algo_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1')
		ON CONFLICT(bucket_type, bucket_key, area_type, area_key) DO UPDATE SET
			movement_intensity = excluded.movement_intensity,
			burst_intensity = excluded.burst_intensity,
			burst_count = excluded.burst_count,
			burst_duration_s = excluded.burst_duration_s,
			active_time_s = excluded.active_time_s,
			inactive_time_s = excluded.inactive_time_s,
			activity_ratio = excluded.activity_ratio,
			effective_movement_ratio = excluded.effective_movement_ratio,
			avg_speed_kmh = excluded.avg_speed_kmh,
			max_speed_kmh = excluded.max_speed_kmh,
			distance_per_day = excluded.distance_per_day,
			time_compression_index = excluded.time_compression_index,
			total_distance_m = excluded.total_distance_m,
			total_duration_s = excluded.total_duration_s,
			trip_count = excluded.trip_count,
			distinct_days = excluded.distinct_days,
			updated_at = CAST(strftime('%s', 'now') AS INTEGER)
	`

	_, err := a.DB.ExecContext(ctx, query,
		bucketType, bucketKey, areaType, areaKey,
		stats.MovementIntensity, stats.BurstIntensity, stats.BurstCount, stats.BurstDuration,
		stats.ActiveTime, stats.InactiveTime, stats.ActivityRatio, stats.EffectiveMovementRatio,
		stats.AvgSpeedKmh, stats.MaxSpeedKmh, stats.DistancePerDay, stats.TimeCompressionIndex,
		stats.TotalDistance, stats.TotalDuration, stats.TripCount, stats.DistinctDays,
	)

	return err
}

// getDistinctValues gets distinct values for a column from segments
func (a *MovementIntensityAnalyzer) getDistinctValues(ctx context.Context, column string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM segments
		WHERE %s IS NOT NULL AND %s != ''
		ORDER BY %s
	`, column, column, column, column)

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, rows.Err()
}

// Register the analyzer
func init() {
	log.Println("[advanced] Registering movement_intensity analyzer")
	analysis.RegisterAnalyzer("movement_intensity", NewMovementIntensityAnalyzer)
	log.Println("[advanced] movement_intensity analyzer registered")
}
