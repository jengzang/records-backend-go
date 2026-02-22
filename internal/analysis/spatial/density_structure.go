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

// DensityStructureAnalyzer implements spatial density analysis (simplified)
// Skill: 密度结构分析 (Density Structure)
// Analyzes spatial density patterns using grid-based approach
type DensityStructureAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewDensityStructureAnalyzer creates a new density structure analyzer
func NewDensityStructureAnalyzer(db *sql.DB) analysis.Analyzer {
	return &DensityStructureAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "density_structure", 10000),
	}
}

// Analyze performs density structure analysis
func (a *DensityStructureAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[DensityStructureAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing zones (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM spatial_density_grid_stats"); err != nil {
			return fmt.Errorf("failed to clear spatial_density_grid_stats: %w", err)
		}
		log.Printf("[DensityStructureAnalyzer] Cleared existing density zones")
	}

	// Query grid cells with statistics
	// Note: grid_cells doesn't have admin columns, we'll leave them empty
	query := `
		SELECT
			grid_id, visit_count, point_count, total_duration_s,
			center_lat, center_lon
		FROM grid_cells
		WHERE visit_count > 0
		ORDER BY visit_count DESC
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query grid cells: %w", err)
	}
	defer rows.Close()

	var zones []DensityZone
	var allVisitCounts []int64

	for rows.Next() {
		var zone DensityZone

		if err := rows.Scan(
			&zone.GridID, &zone.VisitCount, &zone.PointCount, &zone.TotalDuration,
			&zone.CenterLat, &zone.CenterLon,
		); err != nil {
			return fmt.Errorf("failed to scan grid cell: %w", err)
		}

		// Calculate visit_days from grid metadata if available
		// For now, estimate as visit_count (simplified)
		zone.VisitDays = int(zone.VisitCount)

		zones = append(zones, zone)
		allVisitCounts = append(allVisitCounts, zone.VisitCount)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	if len(zones) == 0 {
		log.Printf("[DensityStructureAnalyzer] No grid cells to process")
		return a.MarkTaskAsCompleted(taskID, `{"zones": 0}`)
	}

	log.Printf("[DensityStructureAnalyzer] Processing %d grid cells", len(zones))

	// Calculate density scores and classify zones
	a.calculateDensityScores(zones, allVisitCounts)

	// Insert density zones
	if err := a.insertDensityZones(ctx, zones); err != nil {
		return fmt.Errorf("failed to insert density zones: %w", err)
	}

	// Count zones by density level
	coreCount := 0
	secondaryCount := 0
	activeCount := 0
	peripheralCount := 0
	rareCount := 0
	for _, zone := range zones {
		switch zone.DensityLevel {
		case "core":
			coreCount++
		case "secondary":
			secondaryCount++
		case "active":
			activeCount++
		case "peripheral":
			peripheralCount++
		case "rare":
			rareCount++
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_zones":      len(zones),
		"core_zones":       coreCount,
		"secondary_zones":  secondaryCount,
		"active_zones":     activeCount,
		"peripheral_zones": peripheralCount,
		"rare_zones":       rareCount,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[DensityStructureAnalyzer] Analysis completed: %d zones classified", len(zones))
	return nil
}

// DensityZone holds density zone data
type DensityZone struct {
	GridID        string
	DensityScore  float64
	PointCount    int64
	VisitCount    int64
	TotalDuration int64
	VisitDays     int
	DensityLevel  string // 'core', 'secondary', 'active', 'peripheral', 'rare'
	CenterLat     float64
	CenterLon     float64
	Province      string
	City          string
	County        string
}

// calculateDensityScores calculates density scores using weighted formula and classifies zones
func (a *DensityStructureAnalyzer) calculateDensityScores(zones []DensityZone, allVisitCounts []int64) {
	if len(zones) == 0 {
		return
	}

	// Calculate weighted density scores for all zones
	scores := make([]float64, len(zones))
	for i := range zones {
		scores[i] = calculateWeightedDensityScore(
			zones[i].TotalDuration,
			zones[i].VisitDays,
			int(zones[i].VisitCount),
		)
		zones[i].DensityScore = scores[i]
	}

	// Sort scores to find percentiles
	sortedScores := make([]float64, len(scores))
	copy(sortedScores, scores)

	// Simple bubble sort (sufficient for this use case)
	for i := 0; i < len(sortedScores); i++ {
		for j := i + 1; j < len(sortedScores); j++ {
			if sortedScores[i] < sortedScores[j] {
				sortedScores[i], sortedScores[j] = sortedScores[j], sortedScores[i]
			}
		}
	}

	// Calculate percentile thresholds
	// p90 (top 10%), p70 (top 30%), p30 (top 70%), p10 (top 90%)
	p90 := sortedScores[len(sortedScores)/10]
	p70 := sortedScores[len(sortedScores)*3/10]
	p30 := sortedScores[len(sortedScores)*7/10]
	p10 := sortedScores[len(sortedScores)*9/10]

	// Classify density levels
	for i := range zones {
		zones[i].DensityLevel = classifyDensityLevel(zones[i].DensityScore, p90, p70, p30, p10)
	}
}

// calculateWeightedDensityScore calculates weighted density score
// Formula: a*log(1+duration_hours) + b*log(1+days) + c*log(1+count)
func calculateWeightedDensityScore(stayDuration int64, visitDays int, stayCount int) float64 {
	a := 0.5 // duration weight
	b := 0.3 // days weight
	c := 0.2 // count weight

	durationHours := float64(stayDuration) / 3600.0
	score := a*log1p(durationHours) + b*log1p(float64(visitDays)) + c*log1p(float64(stayCount))
	return score
}

// classifyDensityLevel classifies density level based on percentiles
func classifyDensityLevel(score, p90, p70, p30, p10 float64) string {
	if score >= p90 {
		return "core" // Top 10%
	} else if score >= p70 {
		return "secondary" // 10-30%
	} else if score >= p30 {
		return "active" // 30-70%
	} else if score >= p10 {
		return "peripheral" // 70-90%
	}
	return "rare" // Bottom 10%
}

// log1p calculates log(1 + x) safely
func log1p(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Log(1 + x)
}

// insertDensityZones inserts density zones into the database
func (a *DensityStructureAnalyzer) insertDensityZones(ctx context.Context, zones []DensityZone) error {
	if len(zones) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO spatial_density_grid_stats (
			bucket_type, bucket_key, grid_id,
			center_lat, center_lon, province, city, county,
			density_score, density_level,
			stay_duration_s, stay_count, visit_days,
			algo_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1')
		ON CONFLICT(bucket_type, bucket_key, grid_id) DO UPDATE SET
			density_score = excluded.density_score,
			density_level = excluded.density_level,
			stay_duration_s = excluded.stay_duration_s,
			stay_count = excluded.stay_count,
			visit_days = excluded.visit_days,
			updated_at = CAST(strftime('%s', 'now') AS INTEGER)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, zone := range zones {
		_, err := stmt.ExecContext(ctx,
			"all", nil, zone.GridID,
			zone.CenterLat, zone.CenterLon, zone.Province, zone.City, zone.County,
			zone.DensityScore, zone.DensityLevel,
			zone.TotalDuration, zone.VisitCount, zone.VisitDays,
		)
		if err != nil {
			return fmt.Errorf("failed to insert density zone: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[DensityStructureAnalyzer] Inserted %d density zones", len(zones))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("density_structure", NewDensityStructureAnalyzer)
}