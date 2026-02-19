package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/spatial"
)

// GridSystemAnalyzer implements grid-based spatial indexing
// Skill: 网格系统 (Grid System)
// Creates geohash-based spatial index for trajectory points
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

	// Process multiple precision levels (4-7 characters)
	precisions := []int{4, 5, 6, 7}
	totalCells := 0

	for _, precision := range precisions {
		cells, err := a.processGridLevel(ctx, precision)
		if err != nil {
			return fmt.Errorf("failed to process grid level %d: %w", precision, err)
		}
		totalCells += len(cells)
		log.Printf("[GridSystemAnalyzer] Processed precision %d: %d cells", precision, len(cells))
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_cells": totalCells,
		"precisions":  precisions,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[GridSystemAnalyzer] Analysis completed: %d grid cells created", totalCells)
	return nil
}

// processGridLevel processes a single precision level
func (a *GridSystemAnalyzer) processGridLevel(ctx context.Context, precision int) ([]GridCell, error) {
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

		// Calculate geohash
		geohash := spatial.EncodeGeohash(lat, lon, precision)

		// Update or create cell
		if cell, exists := cellMap[geohash]; exists {
			cell.VisitCount++
			if timestamp < cell.FirstVisitTS {
				cell.FirstVisitTS = timestamp
			}
			if timestamp > cell.LastVisitTS {
				cell.LastVisitTS = timestamp
			}
		} else {
			cellMap[geohash] = &GridCell{
				GridID:       geohash,
				Precision:    precision,
				VisitCount:   1,
				FirstVisitTS: timestamp,
				LastVisitTS:  timestamp,
			}
		}
	}

	// Convert map to slice
	var cells []GridCell
	for _, cell := range cellMap {
		cell.TotalDurationS = cell.LastVisitTS - cell.FirstVisitTS
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
	Precision      int
	VisitCount     int64
	TotalDurationS int64
	FirstVisitTS   int64
	LastVisitTS    int64
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
			grid_id, precision, visit_count, total_duration_s,
			first_visit_ts, last_visit_ts, algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, cell := range cells {
		_, err := stmt.ExecContext(ctx,
			cell.GridID,
			cell.Precision,
			cell.VisitCount,
			cell.TotalDurationS,
			cell.FirstVisitTS,
			cell.LastVisitTS,
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

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("grid_system", NewGridSystemAnalyzer)
}
