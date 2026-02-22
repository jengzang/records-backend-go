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

// UtilizationEfficiencyAnalyzer implements spatial utilization efficiency analysis
// Skill: 空间利用效率 (Utilization Efficiency)
// Distinguishes destinations (high stay) from transit corridors (high pass-through)
type UtilizationEfficiencyAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewUtilizationEfficiencyAnalyzer creates a new utilization efficiency analyzer
func NewUtilizationEfficiencyAnalyzer(db *sql.DB) analysis.Analyzer {
	return &UtilizationEfficiencyAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "utilization_efficiency", 10000),
	}
}

// Analyze performs utilization efficiency analysis
func (a *UtilizationEfficiencyAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[UtilizationEfficiencyAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing metrics (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM spatial_utilization_bucketed"); err != nil {
			return fmt.Errorf("failed to clear spatial_utilization_bucketed: %w", err)
		}
		log.Printf("[UtilizationEfficiencyAnalyzer] Cleared existing metrics")
	}

	// Process all-time bucket for all area types
	totalRecords := 0
	for _, areaType := range []string{"province", "city", "county", "town"} {
		count, err := a.processAreaType(ctx, "all", "", areaType)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", areaType, err)
		}
		totalRecords += count
		log.Printf("[UtilizationEfficiencyAnalyzer] Processed %s: %d records", areaType, count)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_records": totalRecords,
		"area_types":    []string{"province", "city", "county", "town"},
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[UtilizationEfficiencyAnalyzer] Analysis completed: %d total records", totalRecords)
	return nil
}

// processAreaType processes all areas of a given type
func (a *UtilizationEfficiencyAnalyzer) processAreaType(ctx context.Context, bucketType, bucketKey, areaType string) (int, error) {
	// Get distinct areas
	areas, err := a.getDistinctAreas(ctx, areaType)
	if err != nil {
		return 0, fmt.Errorf("failed to get distinct areas: %w", err)
	}

	count := 0
	for _, area := range areas {
		if area == "" {
			continue // Skip empty areas
		}

		// Calculate metrics for this area
		metrics, err := a.calculateAreaMetrics(ctx, areaType, area)
		if err != nil {
			log.Printf("[UtilizationEfficiencyAnalyzer] Warning: failed to calculate metrics for %s/%s: %v", areaType, area, err)
			continue
		}

		// Insert record
		if err := a.insertUtilizationRecord(ctx, bucketType, bucketKey, areaType, area, metrics); err != nil {
			log.Printf("[UtilizationEfficiencyAnalyzer] Warning: failed to insert record for %s/%s: %v", areaType, area, err)
			continue
		}

		count++
	}

	return count, nil
}

// getDistinctAreas retrieves distinct areas of a given type
func (a *UtilizationEfficiencyAnalyzer) getDistinctAreas(ctx context.Context, areaType string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM "一生足迹"
		WHERE %s IS NOT NULL AND %s != ''
		ORDER BY %s
	`, areaType, areaType, areaType, areaType)

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query distinct areas: %w", err)
	}
	defer rows.Close()

	var areas []string
	for rows.Next() {
		var area string
		if err := rows.Scan(&area); err != nil {
			return nil, fmt.Errorf("failed to scan area: %w", err)
		}
		areas = append(areas, area)
	}

	return areas, nil
}

// AreaMetrics holds calculated metrics for an area
type AreaMetrics struct {
	TransitIntensity   int
	StayDurationS      int64
	DistinctVisitDays  int
	DistinctGrids      int
	FirstVisit         int64
	LastVisit          int64
	UtilizationEff     float64
	TransitDominance   float64
	AreaDepth          float64
	CoverageEfficiency float64
}

// calculateAreaMetrics calculates all metrics for a given area
func (a *UtilizationEfficiencyAnalyzer) calculateAreaMetrics(ctx context.Context, areaType, areaKey string) (*AreaMetrics, error) {
	metrics := &AreaMetrics{}

	// 1. Calculate transit intensity (count of segments passing through)
	// Since segments table doesn't have trip_id, we count segments instead
	// Join segments with track points to get administrative region info
	transitQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT s.id)
		FROM segments s
		JOIN "一生足迹" p ON p.id = s.start_point_id
		WHERE p.%s = ?
	`, areaType)

	err := a.DB.QueryRowContext(ctx, transitQuery, areaKey).Scan(&metrics.TransitIntensity)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate transit intensity: %w", err)
	}

	// 2. Calculate stay intensity (total stay duration and distinct days)
	stayQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(duration_s), 0) as total_duration,
			COUNT(DISTINCT DATE(start_time, 'unixepoch')) as distinct_days,
			MIN(start_time) as first_visit,
			MAX(end_time) as last_visit
		FROM stay_segments
		WHERE %s = ?
	`, areaType)

	var firstVisit, lastVisit sql.NullInt64
	err = a.DB.QueryRowContext(ctx, stayQuery, areaKey).Scan(
		&metrics.StayDurationS,
		&metrics.DistinctVisitDays,
		&firstVisit,
		&lastVisit,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate stay intensity: %w", err)
	}

	if firstVisit.Valid {
		metrics.FirstVisit = firstVisit.Int64
	}
	if lastVisit.Valid {
		metrics.LastVisit = lastVisit.Int64
	}

	// 3. Calculate grid coverage (distinct grids visited in this area)
	gridQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT geohash6)
		FROM stay_segments
		WHERE %s = ?
	`, areaType)

	err = a.DB.QueryRowContext(ctx, gridQuery, areaKey).Scan(&metrics.DistinctGrids)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate grid coverage: %w", err)
	}

	// 4. Calculate derived metrics
	epsilon := 1.0

	// Utilization efficiency: stay / (transit + ε)
	metrics.UtilizationEff = float64(metrics.StayDurationS) / (float64(metrics.TransitIntensity) + epsilon)

	// Transit dominance: transit / (transit + stay_hours)
	stayHours := float64(metrics.StayDurationS) / 3600.0
	metrics.TransitDominance = float64(metrics.TransitIntensity) / (float64(metrics.TransitIntensity) + stayHours + epsilon)

	// Area depth: log(1 + stay) × log(1 + days)
	metrics.AreaDepth = math.Log(1+float64(metrics.StayDurationS)) * math.Log(1+float64(metrics.DistinctVisitDays))

	// Coverage efficiency: (distinct grids / total grids in area)
	// For now, set to 0 as we don't have total grid count per area
	metrics.CoverageEfficiency = 0

	return metrics, nil
}

// insertUtilizationRecord inserts a utilization record into the database
func (a *UtilizationEfficiencyAnalyzer) insertUtilizationRecord(
	ctx context.Context,
	bucketType, bucketKey, areaType, areaKey string,
	metrics *AreaMetrics,
) error {
	insertQuery := `
		INSERT INTO spatial_utilization_bucketed (
			bucket_type, bucket_key, area_type, area_key,
			transit_intensity, stay_duration_s,
			utilization_efficiency, transit_dominance, area_depth, coverage_efficiency,
			distinct_visit_days, distinct_grids, total_grids,
			first_visit, last_visit,
			algo_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1')
		ON CONFLICT(bucket_type, bucket_key, area_type, area_key) DO UPDATE SET
			transit_intensity = excluded.transit_intensity,
			stay_duration_s = excluded.stay_duration_s,
			utilization_efficiency = excluded.utilization_efficiency,
			transit_dominance = excluded.transit_dominance,
			area_depth = excluded.area_depth,
			coverage_efficiency = excluded.coverage_efficiency,
			distinct_visit_days = excluded.distinct_visit_days,
			distinct_grids = excluded.distinct_grids,
			first_visit = excluded.first_visit,
			last_visit = excluded.last_visit,
			updated_at = CAST(strftime('%s', 'now') AS INTEGER)
	`

	_, err := a.DB.ExecContext(ctx, insertQuery,
		bucketType, bucketKey, areaType, areaKey,
		metrics.TransitIntensity, metrics.StayDurationS,
		metrics.UtilizationEff, metrics.TransitDominance, metrics.AreaDepth, metrics.CoverageEfficiency,
		metrics.DistinctVisitDays, metrics.DistinctGrids, 0, // total_grids = 0 for now
		metrics.FirstVisit, metrics.LastVisit,
	)
	if err != nil {
		return fmt.Errorf("failed to insert utilization record: %w", err)
	}

	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("utilization_efficiency", NewUtilizationEfficiencyAnalyzer)
}
