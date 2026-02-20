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

// GridSystemAnalyzer implements grid-based spatial indexing
// Skill: 网格系统 (Grid System)
// Creates hierarchical grid system for trajectory points
type GridSystemAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewGridSystemAnalyzer creates a new grid system analyzer
func NewGridSystemAnalyzer(db *sql.DB) analysis.Analyzer {
	return &GridSystemAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "grid_system", 5000),
	}
}

// Analyze performs grid system analysis
func (a *GridSystemAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[GridSystemAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing grid cells (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM grid_cells"); err != nil {
			return fmt.Errorf("failed to clear grid cells: %w", err)
		}
		log.Printf("[GridSystemAnalyzer] Cleared existing grid cells")
	}

	// Process multiple zoom levels (8-15)
	// Level 8: ~150km cells (country/province level)
	// Level 10: ~40km cells (city level)
	// Level 12: ~10km cells (district level)
	// Level 15: ~1km cells (neighborhood level)
	levels := []int{8, 10, 12, 15}
	totalCells := 0

	for _, level := range levels {
		cells, err := a.processGridLevel(ctx, level)
		if err != nil {
			return fmt.Errorf("failed to process grid level %d: %w", level, err)
		}
		totalCells += len(cells)
		log.Printf("[GridSystemAnalyzer] Processed level %d: %d cells", level, len(cells))
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_cells": totalCells,
		"levels":      levels,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[GridSystemAnalyzer] Analysis completed: %d grid cells created", totalCells)
	return nil
}

// processGridLevel processes a single zoom level
func (a *GridSystemAnalyzer) processGridLevel(ctx context.Context, level int) ([]GridCell, error) {
	// Get all track points
	pointsQuery := `
		SELECT
			id,
			dataTime,
			latitude,
			longitude
		FROM "一生足迹"
		WHERE outlier_flag = 0
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, pointsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query points: %w", err)
	}
	defer rows.Close()

	// Aggregate points by grid cell
	cellMap := make(map[string]*GridCell)

	for rows.Next() {
		var id, timestamp int64
		var lat, lon float64

		if err := rows.Scan(&id, &timestamp, &lat, &lon); err != nil {
			return nil, fmt.Errorf("failed to scan point: %w", err)
		}

		// Calculate grid cell coordinates
		gridX, gridY := latLonToTile(lat, lon, level)
		gridID := fmt.Sprintf("L%d_%d_%d", level, gridX, gridY)

		// Calculate cell bounds
		minLat, minLon, maxLat, maxLon := tileToBounds(gridX, gridY, level)
		centerLat := (minLat + maxLat) / 2
		centerLon := (minLon + maxLon) / 2

		// Update or create cell
		if cell, exists := cellMap[gridID]; exists {
			cell.PointCount++
			cell.VisitCount++
			if timestamp < cell.FirstVisit {
				cell.FirstVisit = timestamp
			}
			if timestamp > cell.LastVisit {
				cell.LastVisit = timestamp
			}
		} else {
			cellMap[gridID] = &GridCell{
				GridID:     gridID,
				Level:      level,
				BBoxMinLat: minLat,
				BBoxMinLon: minLon,
				BBoxMaxLat: maxLat,
				BBoxMaxLon: maxLon,
				CenterLat:  centerLat,
				CenterLon:  centerLon,
				PointCount: 1,
				VisitCount: 1,
				FirstVisit: timestamp,
				LastVisit:  timestamp,
			}
		}
	}

	// Convert map to slice and calculate durations
	var cells []GridCell
	for _, cell := range cellMap {
		cell.TotalDurationS = cell.LastVisit - cell.FirstVisit
		cells = append(cells, *cell)
	}

	// Insert cells
	if err := a.insertGridCells(ctx, cells); err != nil {
		return nil, fmt.Errorf("failed to insert grid cells: %w", err)
	}

	return cells, nil
}

// GridCell holds grid cell data
type GridCell struct {
	GridID         string
	Level          int
	BBoxMinLat     float64
	BBoxMinLon     float64
	BBoxMaxLat     float64
	BBoxMaxLon     float64
	CenterLat      float64
	CenterLon      float64
	PointCount     int64
	VisitCount     int64
	FirstVisit     int64
	LastVisit      int64
	TotalDurationS int64
}

// insertGridCells inserts grid cells into the database
func (a *GridSystemAnalyzer) insertGridCells(ctx context.Context, cells []GridCell) error {
	if len(cells) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO grid_cells (
			grid_id, level, bbox_min_lat, bbox_min_lon, bbox_max_lat, bbox_max_lon,
			center_lat, center_lon, point_count, visit_count,
			first_visit, last_visit, total_duration_s, modes, metadata,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
				  CAST(strftime('%s', 'now') AS INTEGER),
				  CAST(strftime('%s', 'now') AS INTEGER))
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, cell := range cells {
		// Create empty JSON arrays/objects for modes and metadata
		modes := "[]"
		metadata := "{}"

		_, err := stmt.ExecContext(ctx,
			cell.GridID,
			cell.Level,
			cell.BBoxMinLat,
			cell.BBoxMinLon,
			cell.BBoxMaxLat,
			cell.BBoxMaxLon,
			cell.CenterLat,
			cell.CenterLon,
			cell.PointCount,
			cell.VisitCount,
			cell.FirstVisit,
			cell.LastVisit,
			cell.TotalDurationS,
			modes,
			metadata,
		)
		if err != nil {
			return fmt.Errorf("failed to insert grid cell: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[GridSystemAnalyzer] Inserted %d grid cells", len(cells))
	return nil
}

// latLonToTile converts lat/lon to tile coordinates at given zoom level
// Uses Web Mercator projection (EPSG:3857)
func latLonToTile(lat, lon float64, zoom int) (x, y int) {
	n := math.Pow(2, float64(zoom))
	x = int((lon + 180.0) / 360.0 * n)
	latRad := lat * math.Pi / 180.0
	y = int((1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n)
	return x, y
}

// tileToBounds converts tile coordinates to lat/lon bounds
func tileToBounds(x, y, zoom int) (minLat, minLon, maxLat, maxLon float64) {
	n := math.Pow(2, float64(zoom))
	minLon = float64(x)/n*360.0 - 180.0
	maxLon = float64(x+1)/n*360.0 - 180.0
	minLat = tileYToLat(y+1, zoom)
	maxLat = tileYToLat(y, zoom)
	return minLat, minLon, maxLat, maxLon
}

// tileYToLat converts tile Y coordinate to latitude
func tileYToLat(y, zoom int) float64 {
	n := math.Pow(2, float64(zoom))
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	return latRad * 180.0 / math.Pi
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("grid_system", NewGridSystemAnalyzer)
}
