package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// UtilizationEfficiencyAnalyzer implements spatial utilization efficiency analysis
// Skill: 空间利用效率 (Utilization Efficiency)
// Calculates spatial coverage and revisit efficiency metrics
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
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM utilization_metrics"); err != nil {
			return fmt.Errorf("failed to clear utilization_metrics: %w", err)
		}
		log.Printf("[UtilizationEfficiencyAnalyzer] Cleared existing metrics")
	}

	// Compute all-time metrics
	allTimeMetrics, err := a.computeUtilizationMetrics(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to compute all-time metrics: %w", err)
	}

	// Insert metrics
	if err := a.insertMetrics(ctx, allTimeMetrics); err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"utilization_ratio":  allTimeMetrics.UtilizationRatio,
		"revisit_efficiency": allTimeMetrics.RevisitEfficiency,
		"unique_grids":       allTimeMetrics.UniqueGrids,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[UtilizationEfficiencyAnalyzer] Analysis completed")
	return nil
}

// UtilizationMetrics holds utilization efficiency metrics
type UtilizationMetrics struct {
	MetricDate        string
	TotalAreaKm2      float64
	VisitedAreaKm2    float64
	UtilizationRatio  float64
	RevisitEfficiency float64
	UniqueGrids       int64
	TotalVisits       int64
	AvgVisitsPerGrid  float64
}

// computeUtilizationMetrics computes utilization metrics
func (a *UtilizationEfficiencyAnalyzer) computeUtilizationMetrics(ctx context.Context, date string) (*UtilizationMetrics, error) {
	// Query grid statistics
	query := `
		SELECT
			COUNT(DISTINCT grid_id) as unique_grids,
			SUM(visit_count) as total_visits,
			AVG(visit_count) as avg_visits_per_grid
		FROM grid_cells
	`

	var metrics UtilizationMetrics
	var avgVisits sql.NullFloat64

	err := a.DB.QueryRowContext(ctx, query).Scan(
		&metrics.UniqueGrids,
		&metrics.TotalVisits,
		&avgVisits,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query grid stats: %w", err)
	}

	if avgVisits.Valid {
		metrics.AvgVisitsPerGrid = avgVisits.Float64
	}

	metrics.MetricDate = date

	// Calculate area metrics
	// Geohash precision 6 = ~1.2km x 0.6km = ~0.72 km²
	gridAreaKm2 := 0.72
	metrics.VisitedAreaKm2 = float64(metrics.UniqueGrids) * gridAreaKm2

	// Estimate total area from bounding box
	boundingBoxQuery := `
		SELECT
			MIN(latitude) as min_lat,
			MAX(latitude) as max_lat,
			MIN(longitude) as min_lon,
			MAX(longitude) as max_lon
		FROM "一生足迹"
		WHERE outlier_flag = 0
	`

	var minLat, maxLat, minLon, maxLon float64
	err = a.DB.QueryRowContext(ctx, boundingBoxQuery).Scan(&minLat, &maxLat, &minLon, &maxLon)
	if err != nil {
		return nil, fmt.Errorf("failed to query bounding box: %w", err)
	}

	// Approximate bounding box area (simplified)
	latDiff := maxLat - minLat
	lonDiff := maxLon - minLon
	// At mid-latitudes, 1 degree lat ≈ 111 km, 1 degree lon ≈ 85 km
	metrics.TotalAreaKm2 = latDiff * 111 * lonDiff * 85

	// Calculate utilization ratio
	if metrics.TotalAreaKm2 > 0 {
		metrics.UtilizationRatio = metrics.VisitedAreaKm2 / metrics.TotalAreaKm2
	}

	// Calculate revisit efficiency (how efficiently we revisit areas)
	// Higher value = more revisits per unique location
	if metrics.UniqueGrids > 0 {
		metrics.RevisitEfficiency = float64(metrics.TotalVisits) / float64(metrics.UniqueGrids)
	}

	return &metrics, nil
}

// insertMetrics inserts utilization metrics into the database
func (a *UtilizationEfficiencyAnalyzer) insertMetrics(ctx context.Context, metrics *UtilizationMetrics) error {
	insertQuery := `
		INSERT INTO utilization_metrics (
			metric_date, total_area_km2, visited_area_km2,
			utilization_ratio, revisit_efficiency,
			unique_grids, total_visits, avg_visits_per_grid,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	_, err := a.DB.ExecContext(ctx, insertQuery,
		metrics.MetricDate, metrics.TotalAreaKm2, metrics.VisitedAreaKm2,
		metrics.UtilizationRatio, metrics.RevisitEfficiency,
		metrics.UniqueGrids, metrics.TotalVisits, metrics.AvgVisitsPerGrid,
	)
	if err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	log.Printf("[UtilizationEfficiencyAnalyzer] Inserted utilization metrics")
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("utilization_efficiency", NewUtilizationEfficiencyAnalyzer)
}