package advanced

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// Package initialization marker
var _ = func() struct{} {
	log.Println("[advanced] Package loaded")
	return struct{}{}
}()

// AltitudeStatsAnalyzer implements altitude statistics analysis
// Skill: 28_altitude_dimension (Altitude Dimension - Statistics)
// Analyzes altitude distribution and vertical movement statistics
type AltitudeStatsAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewAltitudeStatsAnalyzer creates a new altitude stats analyzer
func NewAltitudeStatsAnalyzer(db *sql.DB) analysis.Analyzer {
	return &AltitudeStatsAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "altitude_stats", 10000),
	}
}

// Analyze performs altitude statistics analysis
func (a *AltitudeStatsAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[AltitudeStatsAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing stats (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM altitude_stats_bucketed"); err != nil {
			return fmt.Errorf("failed to clear altitude_stats_bucketed: %w", err)
		}
		log.Printf("[AltitudeStatsAnalyzer] Cleared existing altitude stats")
	}

	// Process different aggregation levels
	totalRecords := 0

	// 1. Global stats (ALL)
	if err := a.processAltitudeStats(ctx, "ALL", ""); err != nil {
		return fmt.Errorf("failed to process global stats: %w", err)
	}
	totalRecords++

	// 2. Province-level stats
	provinces, err := a.getDistinctValues(ctx, "province")
	if err != nil {
		return fmt.Errorf("failed to get provinces: %w", err)
	}
	for _, province := range provinces {
		if err := a.processAltitudeStats(ctx, "PROVINCE", province); err != nil {
			log.Printf("[AltitudeAnalyzer] Warning: failed to process province %s: %v", province, err)
			continue
		}
		totalRecords++
	}

	// 3. City-level stats
	cities, err := a.getDistinctValues(ctx, "city")
	if err != nil {
		return fmt.Errorf("failed to get cities: %w", err)
	}
	for _, city := range cities {
		if err := a.processAltitudeStats(ctx, "CITY", city); err != nil {
			log.Printf("[AltitudeAnalyzer] Warning: failed to process city %s: %v", city, err)
			continue
		}
		totalRecords++
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_records": totalRecords,
		"provinces":     len(provinces),
		"cities":        len(cities),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[AltitudeStatsAnalyzer] Analysis completed: %d records generated", totalRecords)
	return nil
}

// processAltitudeStats processes altitude statistics for a specific area
func (a *AltitudeStatsAnalyzer) processAltitudeStats(ctx context.Context, areaType, areaKey string) error {
	// Query track points with altitude data
	query := `
		SELECT
			id,
			altitude,
			distance
		FROM "一生足迹"
		WHERE altitude IS NOT NULL
		  AND altitude > 0
	`
	args := []interface{}{}

	if areaType != "ALL" {
		query += " AND " + areaType + " = ?"
		args = append(args, areaKey)
	}

	query += " ORDER BY dataTime"

	rows, err := a.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query track points: %w", err)
	}
	defer rows.Close()

	var altitudes []float64
	var totalAscent, totalDescent, totalDistance float64
	var prevAltitude float64
	pointCount := 0

	for rows.Next() {
		var id int64
		var altitude, distance float64

		if err := rows.Scan(&id, &altitude, &distance); err != nil {
			return fmt.Errorf("failed to scan track point: %w", err)
		}

		altitudes = append(altitudes, altitude)

		// Calculate ascent/descent from previous point
		if pointCount > 0 {
			altDiff := altitude - prevAltitude
			if altDiff > 0 {
				totalAscent += altDiff
			} else {
				totalDescent += math.Abs(altDiff)
			}
		}

		totalDistance += distance
		prevAltitude = altitude
		pointCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	if len(altitudes) == 0 {
		log.Printf("[AltitudeStatsAnalyzer] No altitude data for %s/%s", areaType, areaKey)
		return nil
	}

	// Calculate statistics
	stats := calculateAltitudeStats(altitudes, totalAscent, totalDescent, totalDistance, pointCount)

	// Insert into database
	if err := a.insertAltitudeStats(ctx, "all", "", areaType, areaKey, stats); err != nil {
		return fmt.Errorf("failed to insert altitude stats: %w", err)
	}

	return nil
}

// AltitudeStats holds altitude statistics
type AltitudeStats struct {
	MinAltitude       float64
	MaxAltitude       float64
	AvgAltitude       float64
	AltitudeSpan      float64
	P25Altitude       float64
	P50Altitude       float64
	P75Altitude       float64
	P90Altitude       float64
	TotalAscent       float64
	TotalDescent      float64
	VerticalIntensity float64
	PointCount        int
	SegmentCount      int
	TotalDistance     float64
}

// calculateAltitudeStats calculates altitude statistics from altitude values
func calculateAltitudeStats(altitudes []float64, totalAscent, totalDescent, totalDistance float64, pointCount int) AltitudeStats {
	if len(altitudes) == 0 {
		return AltitudeStats{}
	}

	// Sort for percentile calculation
	sorted := make([]float64, len(altitudes))
	copy(sorted, altitudes)
	sort.Float64s(sorted)

	// Calculate basic stats
	minAlt := sorted[0]
	maxAlt := sorted[len(sorted)-1]
	sum := 0.0
	for _, alt := range altitudes {
		sum += alt
	}
	avgAlt := sum / float64(len(altitudes))

	// Calculate percentiles
	p25 := percentile(sorted, 0.25)
	p50 := percentile(sorted, 0.50)
	p75 := percentile(sorted, 0.75)
	p90 := percentile(sorted, 0.90)

	// Calculate vertical intensity
	verticalIntensity := 0.0
	if totalDistance > 0 {
		verticalIntensity = (totalAscent + totalDescent) / totalDistance
	}

	return AltitudeStats{
		MinAltitude:       minAlt,
		MaxAltitude:       maxAlt,
		AvgAltitude:       avgAlt,
		AltitudeSpan:      maxAlt - minAlt,
		P25Altitude:       p25,
		P50Altitude:       p50,
		P75Altitude:       p75,
		P90Altitude:       p90,
		TotalAscent:       totalAscent,
		TotalDescent:      totalDescent,
		VerticalIntensity: verticalIntensity,
		PointCount:        len(altitudes),
		SegmentCount:      0, // Not applicable for point-based analysis
		TotalDistance:     totalDistance,
	}
}

// percentile calculates the percentile value from sorted data
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// insertAltitudeStats inserts altitude statistics into the database
func (a *AltitudeStatsAnalyzer) insertAltitudeStats(
	ctx context.Context,
	bucketType, bucketKey, areaType, areaKey string,
	stats AltitudeStats,
) error {
	query := `
		INSERT INTO altitude_stats_bucketed (
			bucket_type, bucket_key, area_type, area_key,
			min_altitude, max_altitude, avg_altitude, altitude_span,
			p25_altitude, p50_altitude, p75_altitude, p90_altitude,
			total_ascent, total_descent, vertical_intensity,
			point_count, segment_count, total_distance,
			algo_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1')
		ON CONFLICT(bucket_type, bucket_key, area_type, area_key) DO UPDATE SET
			min_altitude = excluded.min_altitude,
			max_altitude = excluded.max_altitude,
			avg_altitude = excluded.avg_altitude,
			altitude_span = excluded.altitude_span,
			p25_altitude = excluded.p25_altitude,
			p50_altitude = excluded.p50_altitude,
			p75_altitude = excluded.p75_altitude,
			p90_altitude = excluded.p90_altitude,
			total_ascent = excluded.total_ascent,
			total_descent = excluded.total_descent,
			vertical_intensity = excluded.vertical_intensity,
			point_count = excluded.point_count,
			segment_count = excluded.segment_count,
			total_distance = excluded.total_distance,
			updated_at = CAST(strftime('%s', 'now') AS INTEGER)
	`

	_, err := a.DB.ExecContext(ctx, query,
		bucketType, bucketKey, areaType, areaKey,
		stats.MinAltitude, stats.MaxAltitude, stats.AvgAltitude, stats.AltitudeSpan,
		stats.P25Altitude, stats.P50Altitude, stats.P75Altitude, stats.P90Altitude,
		stats.TotalAscent, stats.TotalDescent, stats.VerticalIntensity,
		stats.PointCount, stats.SegmentCount, stats.TotalDistance,
	)

	return err
}

// getDistinctValues gets distinct values for a column from track points
func (a *AltitudeStatsAnalyzer) getDistinctValues(ctx context.Context, column string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM "一生足迹"
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
	log.Println("[advanced] Registering altitude_stats analyzer")
	analysis.RegisterAnalyzer("altitude_stats", NewAltitudeStatsAnalyzer)
	log.Println("[advanced] altitude_stats analyzer registered")
}
