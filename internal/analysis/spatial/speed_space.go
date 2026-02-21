package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/stats"
)

// SpeedSpaceAnalyzer implements speed-space coupling analysis
// Skill: 速度-空间耦合层 (Speed-Space Coupling)
// Analyzes the coupling between speed and spatial structure
type SpeedSpaceAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewSpeedSpaceAnalyzer creates a new speed-space coupling analyzer
func NewSpeedSpaceAnalyzer(db *sql.DB) analysis.Analyzer {
	return &SpeedSpaceAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "speed_space_coupling", 1000),
	}
}

// Analyze performs speed-space coupling analysis
func (a *SpeedSpaceAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[SpeedSpaceAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Get all segments with speed and location data
	segmentsQuery := `
		SELECT
			s.id,
			s.start_time,
			s.end_time,
			s.distance_m,
			s.avg_speed_kmh,
			s.mode,
			p.province,
			p.city,
			p.county,
			strftime('%Y', datetime(s.start_time, 'unixepoch')) as year,
			strftime('%Y-%m', datetime(s.start_time, 'unixepoch')) as month
		FROM segments s
		LEFT JOIN "一生足迹" p ON s.start_point_id = p.id
		WHERE s.avg_speed_kmh IS NOT NULL
			AND s.distance_m > 0
		ORDER BY s.id
	`

	rows, err := a.DB.QueryContext(ctx, segmentsQuery)
	if err != nil {
		return fmt.Errorf("failed to query segments: %w", err)
	}

	// Aggregate speed statistics by area
	areaStats := make(map[string]*AreaSpeedStat)
	var allSpeeds []float64
	totalSegments := 0

	for rows.Next() {
		var (
			id                                      int64
			startTS, endTS                          int64
			distance, avgSpeed                      float64
			mode                                    string
			province, city, county                  sql.NullString
			year, month                             string
		)

		if err := rows.Scan(&id, &startTS, &endTS, &distance, &avgSpeed, &mode, &province, &city, &county, &year, &month); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan segment: %w", err)
		}

		totalSegments++
		allSpeeds = append(allSpeeds, avgSpeed)

		// Aggregate by province
		if province.Valid && province.String != "" {
			a.aggregateAreaSpeed(areaStats, "PROVINCE", province.String, year, avgSpeed, distance)
			a.aggregateAreaSpeed(areaStats, "PROVINCE", province.String, month, avgSpeed, distance)
			a.aggregateAreaSpeed(areaStats, "PROVINCE", province.String, "all", avgSpeed, distance)
		}

		// Aggregate by city
		if city.Valid && city.String != "" {
			a.aggregateAreaSpeed(areaStats, "CITY", city.String, year, avgSpeed, distance)
			a.aggregateAreaSpeed(areaStats, "CITY", city.String, month, avgSpeed, distance)
			a.aggregateAreaSpeed(areaStats, "CITY", city.String, "all", avgSpeed, distance)
		}

		// Aggregate by county
		if county.Valid && county.String != "" {
			a.aggregateAreaSpeed(areaStats, "COUNTY", county.String, year, avgSpeed, distance)
			a.aggregateAreaSpeed(areaStats, "COUNTY", county.String, month, avgSpeed, distance)
			a.aggregateAreaSpeed(areaStats, "COUNTY", county.String, "all", avgSpeed, distance)
		}
	}
	rows.Close()

	log.Printf("[SpeedSpaceAnalyzer] Processed %d segments", totalSegments)

	// Calculate global speed thresholds
	highSpeedThreshold := stats.Percentile(allSpeeds, 90)
	lowSpeedThreshold := stats.Percentile(allSpeeds, 25)

	log.Printf("[SpeedSpaceAnalyzer] Speed thresholds: high=%.2f km/h, low=%.2f km/h", highSpeedThreshold, lowSpeedThreshold)

	// Calculate final statistics and classify zones
	for _, stat := range areaStats {
		// Calculate weighted average speed
		stat.AvgSpeed = stat.TotalWeightedSpeed / stat.TotalDistance

		// Calculate weighted variance
		stat.SpeedVariance = stat.TotalWeightedVariance / stat.TotalDistance

		// Classify zones
		if stat.AvgSpeed > highSpeedThreshold {
			stat.IsHighSpeedZone = true
		}
		if stat.AvgSpeed < lowSpeedThreshold {
			stat.IsSlowLifeZone = true
		}

		// Calculate speed entropy
		stat.SpeedEntropy = a.calculateSpeedEntropy(stat.SpeedBins)
	}

	// Insert results into spatial_analysis table
	if err := a.insertSpeedSpaceResults(ctx, areaStats); err != nil {
		return fmt.Errorf("failed to insert results: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_segments":     totalSegments,
		"areas_analyzed":     len(areaStats),
		"high_speed_zones":   a.countHighSpeedZones(areaStats),
		"slow_life_zones":    a.countSlowLifeZones(areaStats),
		"global_avg_speed":   stats.Mean(allSpeeds),
		"high_speed_threshold": highSpeedThreshold,
		"low_speed_threshold":  lowSpeedThreshold,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[SpeedSpaceAnalyzer] Analysis completed: %d areas analyzed", len(areaStats))
	return nil
}

// AreaSpeedStat holds speed statistics for an area
type AreaSpeedStat struct {
	AreaType             string
	AreaKey              string
	TimeRange            string
	TotalDistance        float64
	TotalWeightedSpeed   float64
	TotalWeightedVariance float64
	AvgSpeed             float64
	SpeedVariance        float64
	SpeedEntropy         float64
	SegmentCount         int
	IsHighSpeedZone      bool
	IsSlowLifeZone       bool
	SpeedBins            map[int]float64 // Speed bins for entropy calculation
}

// aggregateAreaSpeed aggregates speed data for an area
func (a *SpeedSpaceAnalyzer) aggregateAreaSpeed(stats map[string]*AreaSpeedStat, areaType, areaKey, timeRange string, speed, distance float64) {
	key := fmt.Sprintf("%s|%s|%s", areaType, areaKey, timeRange)

	stat, exists := stats[key]
	if !exists {
		stat = &AreaSpeedStat{
			AreaType:  areaType,
			AreaKey:   areaKey,
			TimeRange: timeRange,
			SpeedBins: make(map[int]float64),
		}
		stats[key] = stat
	}

	// Update weighted statistics
	stat.TotalDistance += distance
	stat.TotalWeightedSpeed += speed * distance
	stat.SegmentCount++

	// Update speed bins for entropy calculation (10 km/h bins)
	bin := int(speed / 10)
	stat.SpeedBins[bin] += distance

	// Note: Variance will be calculated in a second pass
}

// calculateSpeedEntropy calculates Shannon entropy of speed distribution
func (a *SpeedSpaceAnalyzer) calculateSpeedEntropy(speedBins map[int]float64) float64 {
	if len(speedBins) == 0 {
		return 0
	}

	// Calculate total distance
	var totalDistance float64
	for _, distance := range speedBins {
		totalDistance += distance
	}

	// Calculate entropy
	var entropy float64
	for _, distance := range speedBins {
		if distance > 0 {
			p := distance / totalDistance
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// insertSpeedSpaceResults inserts speed-space coupling results
func (a *SpeedSpaceAnalyzer) insertSpeedSpaceResults(ctx context.Context, stats map[string]*AreaSpeedStat) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO speed_space_stats_bucketed (
			bucket_type, bucket_key, area_type, area_key,
			avg_speed, speed_variance, speed_entropy,
			total_distance, segment_count,
			is_high_speed_zone, is_slow_life_zone,
			stay_intensity, algo_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
		ON CONFLICT(bucket_type, bucket_key, area_type, area_key)
		DO UPDATE SET
			avg_speed = excluded.avg_speed,
			speed_variance = excluded.speed_variance,
			speed_entropy = excluded.speed_entropy,
			total_distance = excluded.total_distance,
			segment_count = excluded.segment_count,
			is_high_speed_zone = excluded.is_high_speed_zone,
			is_slow_life_zone = excluded.is_slow_life_zone,
			stay_intensity = excluded.stay_intensity,
			created_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, stat := range stats {
		// Determine bucket type and key from time range
		bucketType, bucketKey := a.parseBucketInfo(stat.TimeRange)

		_, err := stmt.ExecContext(ctx,
			bucketType,
			bucketKey,
			stat.AreaType,
			stat.AreaKey,
			stat.AvgSpeed,
			stat.SpeedVariance,
			stat.SpeedEntropy,
			stat.TotalDistance,
			stat.SegmentCount,
			stat.IsHighSpeedZone,
			stat.IsSlowLifeZone,
			0.0, // stay_intensity - will be calculated later if needed
		)
		if err != nil {
			return fmt.Errorf("failed to insert result: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// parseBucketInfo parses time range into bucket type and key
func (a *SpeedSpaceAnalyzer) parseBucketInfo(timeRange string) (string, string) {
	if timeRange == "all" {
		return "all", "all"
	}
	if len(timeRange) == 4 { // Year: "2024"
		return "year", timeRange
	}
	if len(timeRange) == 7 { // Month: "2024-01"
		return "month", timeRange
	}
	return "all", "all"
}

// countHighSpeedZones counts the number of high-speed zones
func (a *SpeedSpaceAnalyzer) countHighSpeedZones(stats map[string]*AreaSpeedStat) int {
	count := 0
	for _, stat := range stats {
		if stat.IsHighSpeedZone {
			count++
		}
	}
	return count
}

// countSlowLifeZones counts the number of slow-life zones
func (a *SpeedSpaceAnalyzer) countSlowLifeZones(stats map[string]*AreaSpeedStat) int {
	count := 0
	for _, stat := range stats {
		if stat.IsSlowLifeZone {
			count++
		}
	}
	return count
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("speed_space_coupling", NewSpeedSpaceAnalyzer)
}
