package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

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
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM density_zones"); err != nil {
			return fmt.Errorf("failed to clear density_zones: %w", err)
		}
		log.Printf("[DensityStructureAnalyzer] Cleared existing density zones")
	}

	// Query grid cells with statistics
	query := `
		SELECT
			grid_id, visit_count, point_count, total_duration_s,
			center_lat, center_lon, province, city, county
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
		var province, city, county sql.NullString

		if err := rows.Scan(
			&zone.GridID, &zone.VisitCount, &zone.PointCount, &zone.TotalDuration,
			&zone.CenterLat, &zone.CenterLon,
			&province, &city, &county,
		); err != nil {
			return fmt.Errorf("failed to scan grid cell: %w", err)
		}

		if province.Valid {
			zone.Province = province.String
		}
		if city.Valid {
			zone.City = city.String
		}
		if county.Valid {
			zone.County = county.String
		}

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

	// Count zones by type
	hotCount := 0
	warmCount := 0
	coldCount := 0
	for _, zone := range zones {
		switch zone.ZoneType {
		case "HOT":
			hotCount++
		case "WARM":
			warmCount++
		case "COLD":
			coldCount++
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_zones": len(zones),
		"hot_zones":   hotCount,
		"warm_zones":  warmCount,
		"cold_zones":  coldCount,
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
	ZoneType      string
	CenterLat     float64
	CenterLon     float64
	Province      string
	City          string
	County        string
}

// calculateDensityScores calculates density scores and classifies zones
func (a *DensityStructureAnalyzer) calculateDensityScores(zones []DensityZone, allVisitCounts []int64) {
	if len(allVisitCounts) == 0 {
		return
	}

	// Calculate percentiles for classification
	// Sort visit counts to find percentiles
	sortedCounts := make([]int64, len(allVisitCounts))
	copy(sortedCounts, allVisitCounts)

	// Simple bubble sort (sufficient for this use case)
	for i := 0; i < len(sortedCounts); i++ {
		for j := i + 1; j < len(sortedCounts); j++ {
			if sortedCounts[i] < sortedCounts[j] {
				sortedCounts[i], sortedCounts[j] = sortedCounts[j], sortedCounts[i]
			}
		}
	}

	// Find thresholds (top 10% = HOT, 10-30% = WARM, rest = COLD)
	hotThreshold := sortedCounts[len(sortedCounts)/10]
	warmThreshold := sortedCounts[len(sortedCounts)*3/10]

	// Calculate max visit count for normalization
	maxVisitCount := sortedCounts[0]

	// Assign density scores and zone types
	for i := range zones {
		// Normalize density score (0-1)
		if maxVisitCount > 0 {
			zones[i].DensityScore = float64(zones[i].VisitCount) / float64(maxVisitCount)
		}

		// Classify zone type
		if zones[i].VisitCount >= hotThreshold {
			zones[i].ZoneType = "HOT"
		} else if zones[i].VisitCount >= warmThreshold {
			zones[i].ZoneType = "WARM"
		} else {
			zones[i].ZoneType = "COLD"
		}
	}
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
		INSERT INTO density_zones (
			grid_id, density_score, point_count, visit_count,
			total_duration_s, zone_type,
			center_lat, center_lon, province, city, county,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, zone := range zones {
		_, err := stmt.ExecContext(ctx,
			zone.GridID, zone.DensityScore, zone.PointCount, zone.VisitCount,
			zone.TotalDuration, zone.ZoneType,
			zone.CenterLat, zone.CenterLon, zone.Province, zone.City, zone.County,
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