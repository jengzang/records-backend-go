package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// RoadOverlapAnalyzer implements road network overlap analysis (simplified)
// Skill: 道路重叠分析 (Road Overlap)
// Analyzes trajectory overlap with road networks using speed-based heuristics
type RoadOverlapAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewRoadOverlapAnalyzer creates a new road overlap analyzer
func NewRoadOverlapAnalyzer(db *sql.DB) analysis.Analyzer {
	return &RoadOverlapAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "road_overlap", 10000),
	}
}

// Analyze performs road overlap analysis
func (a *RoadOverlapAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[RoadOverlapAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing stats (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM road_overlap_stats"); err != nil {
			return fmt.Errorf("failed to clear road_overlap_stats: %w", err)
		}
		log.Printf("[RoadOverlapAnalyzer] Cleared existing road overlap stats")
	}

	// Query segments
	query := `
		SELECT
			id, mode, distance_m, avg_speed_kmh, max_speed_kmh
		FROM segments
		WHERE mode IN ('CAR', 'BIKE', 'WALK')
			AND distance_m > 0
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	var stats []RoadOverlapStat
	totalSegments := 0

	for rows.Next() {
		var segmentID int64
		var mode string
		var distance, avgSpeedKmh, maxSpeedKmh float64

		if err := rows.Scan(&segmentID, &mode, &distance, &avgSpeedKmh, &maxSpeedKmh); err != nil {
			return fmt.Errorf("failed to scan segment: %w", err)
		}

		totalSegments++

		// Convert km/h to m/s
		avgSpeed := avgSpeedKmh / 3.6
		maxSpeed := maxSpeedKmh / 3.6

		// Estimate road overlap based on mode and speed
		stat := a.estimateRoadOverlap(segmentID, mode, distance, avgSpeed, maxSpeed)
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("[RoadOverlapAnalyzer] Processed %d segments", totalSegments)

	// Update task progress
	if err := a.UpdateTaskProgress(taskID, int64(totalSegments), int64(totalSegments), 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Insert road overlap stats
	if err := a.insertRoadOverlapStats(ctx, stats); err != nil {
		return fmt.Errorf("failed to insert road overlap stats: %w", err)
	}

	// Calculate summary statistics
	totalOnRoad := 0.0
	totalOffRoad := 0.0
	for _, stat := range stats {
		totalOnRoad += stat.OnRoadDistance
		totalOffRoad += stat.OffRoadDistance
	}

	overallRatio := 0.0
	if totalOnRoad+totalOffRoad > 0 {
		overallRatio = totalOnRoad / (totalOnRoad + totalOffRoad)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_segments":    totalSegments,
		"on_road_distance":  totalOnRoad,
		"off_road_distance": totalOffRoad,
		"overlap_ratio":     overallRatio,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[RoadOverlapAnalyzer] Analysis completed")
	return nil
}

// RoadOverlapStat holds road overlap statistics
type RoadOverlapStat struct {
	SegmentID      int64
	OnRoadDistance float64
	OffRoadDistance float64
	OverlapRatio   float64
	RoadType       string
	Confidence     float64
}

// estimateRoadOverlap estimates road overlap using speed-based heuristics
func (a *RoadOverlapAnalyzer) estimateRoadOverlap(segmentID int64, mode string, distance, avgSpeed, maxSpeed float64) RoadOverlapStat {
	stat := RoadOverlapStat{
		SegmentID: segmentID,
	}

	// Speed thresholds (m/s)
	const (
		walkSpeed    = 2.0   // 7.2 km/h
		bikeSpeed    = 5.0   // 18 km/h
		carCitySpeed = 13.9  // 50 km/h
		carHighway   = 27.8  // 100 km/h
	)

	// Estimate road overlap based on mode and speed
	switch mode {
	case "CAR":
		// Cars are almost always on roads
		stat.OnRoadDistance = distance
		stat.OffRoadDistance = 0
		stat.OverlapRatio = 1.0
		stat.Confidence = 0.95

		// Classify road type by speed
		if maxSpeed >= carHighway {
			stat.RoadType = "HIGHWAY"
		} else if avgSpeed >= carCitySpeed {
			stat.RoadType = "ARTERIAL"
		} else {
			stat.RoadType = "LOCAL"
		}

	case "BIKE":
		// Bikes are usually on roads or bike paths
		// Assume 90% on road, 10% off road (parks, trails)
		stat.OnRoadDistance = distance * 0.9
		stat.OffRoadDistance = distance * 0.1
		stat.OverlapRatio = 0.9
		stat.RoadType = "LOCAL"
		stat.Confidence = 0.7

	case "WALK":
		// Walking can be on roads, sidewalks, or off-road
		// Estimate based on speed
		if avgSpeed > walkSpeed*1.5 {
			// Fast walking, likely on roads/sidewalks
			stat.OnRoadDistance = distance * 0.8
			stat.OffRoadDistance = distance * 0.2
			stat.OverlapRatio = 0.8
			stat.Confidence = 0.6
		} else {
			// Slow walking, more likely off-road (parks, trails)
			stat.OnRoadDistance = distance * 0.5
			stat.OffRoadDistance = distance * 0.5
			stat.OverlapRatio = 0.5
			stat.Confidence = 0.5
		}
		stat.RoadType = "LOCAL"

	default:
		// Unknown mode, assume mixed
		stat.OnRoadDistance = distance * 0.7
		stat.OffRoadDistance = distance * 0.3
		stat.OverlapRatio = 0.7
		stat.RoadType = "UNKNOWN"
		stat.Confidence = 0.3
	}

	return stat
}

// insertRoadOverlapStats inserts road overlap stats into the database
func (a *RoadOverlapAnalyzer) insertRoadOverlapStats(ctx context.Context, stats []RoadOverlapStat) error {
	if len(stats) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO road_overlap_stats (
			segment_id, on_road_distance_m, off_road_distance_m,
			overlap_ratio, road_type, confidence,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.SegmentID, stat.OnRoadDistance, stat.OffRoadDistance,
			stat.OverlapRatio, stat.RoadType, stat.Confidence,
		)
		if err != nil {
			return fmt.Errorf("failed to insert road overlap stat: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[RoadOverlapAnalyzer] Inserted %d road overlap stats", len(stats))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("road_overlap", NewRoadOverlapAnalyzer)
}