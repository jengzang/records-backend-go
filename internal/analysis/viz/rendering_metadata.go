package viz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/stats"
)

// RenderingMetadataAnalyzer implements rendering metadata generation
// Skill: 渲染元数据生成 (Rendering Metadata)
// Generates visualization metadata for map rendering
type RenderingMetadataAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewRenderingMetadataAnalyzer creates a new rendering metadata analyzer
func NewRenderingMetadataAnalyzer(db *sql.DB) analysis.Analyzer {
	return &RenderingMetadataAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "rendering_metadata", 1000),
	}
}

// Analyze performs rendering metadata generation
func (a *RenderingMetadataAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[RenderingMetadataAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing render cache (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM render_segments_cache"); err != nil {
			return fmt.Errorf("failed to clear render cache: %w", err)
		}
		log.Printf("[RenderingMetadataAnalyzer] Cleared existing render cache")
	}

	// Step 1: Calculate global speed percentiles for bucketing
	speedPercentiles, err := a.calculateSpeedPercentiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to calculate speed percentiles: %w", err)
	}

	log.Printf("[RenderingMetadataAnalyzer] Speed percentiles: %v", speedPercentiles)

	// Step 2: Calculate overlap statistics using grid_id
	overlapStats, err := a.calculateOverlapStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to calculate overlap stats: %w", err)
	}

	log.Printf("[RenderingMetadataAnalyzer] Calculated overlap stats for %d grid cells", len(overlapStats))

	// Step 3: Get all segments
	segmentsQuery := `
		SELECT
			id,
			start_time,
			end_time
		FROM segments
		ORDER BY id
	`

	rows, err := a.DB.QueryContext(ctx, segmentsQuery)
	if err != nil {
		return fmt.Errorf("failed to query segments: %w", err)
	}

	var segments []SegmentInfo
	for rows.Next() {
		var seg SegmentInfo
		if err := rows.Scan(&seg.ID, &seg.StartTS, &seg.EndTS); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, seg)
	}
	rows.Close()

	log.Printf("[RenderingMetadataAnalyzer] Processing %d segments", len(segments))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(segments)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Step 4: Process each segment and generate rendering metadata
	processed := 0
	batchSize := 100
	var renderMetadata []RenderMetadata

	for _, seg := range segments {
		// Get points for this segment
		pointsQuery := `
			SELECT
				speed,
				grid_id
			FROM "一生足迹"
			WHERE dataTime BETWEEN ? AND ?
				AND outlier_flag = 0
			ORDER BY dataTime
		`

		pointRows, err := a.DB.QueryContext(ctx, pointsQuery, seg.StartTS, seg.EndTS)
		if err != nil {
			return fmt.Errorf("failed to query points for segment %d: %w", seg.ID, err)
		}

		var speeds []float64
		var gridIDs []string
		for pointRows.Next() {
			var speed sql.NullFloat64
			var gridID sql.NullString
			if err := pointRows.Scan(&speed, &gridID); err != nil {
				pointRows.Close()
				return fmt.Errorf("failed to scan speed: %w", err)
			}
			if speed.Valid && speed.Float64 > 0 {
				speeds = append(speeds, speed.Float64)
			}
			if gridID.Valid && gridID.String != "" {
				gridIDs = append(gridIDs, gridID.String)
			}
		}
		pointRows.Close()

		if len(speeds) == 0 {
			continue
		}

		// Calculate average speed for this segment
		avgSpeed := stats.Mean(speeds)

		// Determine speed bucket (0-5)
		speedBucket := a.getSpeedBucket(avgSpeed, speedPercentiles)

		// Get overlap rank from most common grid_id in this segment
		overlapRank := 0.0
		if len(gridIDs) > 0 {
			// Find most common grid_id
			gridCounts := make(map[string]int)
			for _, gid := range gridIDs {
				gridCounts[gid]++
			}
			maxCount := 0
			mostCommonGrid := ""
			for gid, count := range gridCounts {
				if count > maxCount {
					maxCount = count
					mostCommonGrid = gid
				}
			}
			if rank, ok := overlapStats[mostCommonGrid]; ok {
				overlapRank = rank
			}
		}

		// Calculate style hints
		lineWeight := a.calculateLineWeight(overlapRank)
		alphaHint := a.calculateAlphaHint(overlapRank)

		// Create render metadata for different LODs
		for lod := 0; lod <= 2; lod++ {
			metadata := RenderMetadata{
				SegmentID:   seg.ID,
				LOD:         lod,
				SpeedBucket: speedBucket,
				OverlapRank: overlapRank,
				LineWeight:  lineWeight,
				AlphaHint:   alphaHint,
			}
			renderMetadata = append(renderMetadata, metadata)
		}

		processed++
		if processed%batchSize == 0 {
			// Insert batch
			if err := a.insertRenderMetadata(ctx, renderMetadata); err != nil {
				return fmt.Errorf("failed to insert render metadata: %w", err)
			}
			renderMetadata = nil

			if err := a.UpdateTaskProgress(taskID, int64(len(segments)), int64(processed), 0); err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}
			log.Printf("[RenderingMetadataAnalyzer] Processed %d/%d segments", processed, len(segments))
		}
	}

	// Insert remaining metadata
	if len(renderMetadata) > 0 {
		if err := a.insertRenderMetadata(ctx, renderMetadata); err != nil {
			return fmt.Errorf("failed to insert render metadata: %w", err)
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_segments":     len(segments),
		"processed_segments": processed,
		"render_entries":     processed * 3, // 3 LODs per segment
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[RenderingMetadataAnalyzer] Analysis completed: %d segments processed", processed)
	return nil
}

// SegmentInfo holds segment information
type SegmentInfo struct {
	ID      int64
	StartTS int64
	EndTS   int64
}

// RenderMetadata holds rendering metadata
type RenderMetadata struct {
	SegmentID   int64
	LOD         int
	SpeedBucket int
	OverlapRank float64
	LineWeight  float64
	AlphaHint   float64
}

// calculateSpeedPercentiles calculates global speed percentiles
func (a *RenderingMetadataAnalyzer) calculateSpeedPercentiles(ctx context.Context) ([]float64, error) {
	query := `
		SELECT speed
		FROM "一生足迹"
		WHERE speed IS NOT NULL
			AND speed > 0
			AND outlier_flag = 0
		ORDER BY RANDOM()
		LIMIT 10000
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query speeds: %w", err)
	}
	defer rows.Close()

	var speeds []float64
	for rows.Next() {
		var speed float64
		if err := rows.Scan(&speed); err != nil {
			return nil, fmt.Errorf("failed to scan speed: %w", err)
		}
		speeds = append(speeds, speed)
	}

	if len(speeds) == 0 {
		return []float64{0, 2, 5, 10, 20, 30}, nil // Default percentiles
	}

	sort.Float64s(speeds)

	// Calculate percentiles for 6 buckets (0-5)
	percentiles := []float64{
		stats.Percentile(speeds, 0),
		stats.Percentile(speeds, 20),
		stats.Percentile(speeds, 40),
		stats.Percentile(speeds, 60),
		stats.Percentile(speeds, 80),
		stats.Percentile(speeds, 100),
	}

	return percentiles, nil
}

// calculateOverlapStats calculates overlap statistics by grid_id from track points
func (a *RenderingMetadataAnalyzer) calculateOverlapStats(ctx context.Context) (map[string]float64, error) {
	query := `
		SELECT
			grid_id,
			COUNT(*) as visit_count
		FROM "一生足迹"
		WHERE grid_id IS NOT NULL
			AND outlier_flag = 0
		GROUP BY grid_id
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query grid stats: %w", err)
	}
	defer rows.Close()

	gridCounts := make(map[string]int)
	var counts []int

	for rows.Next() {
		var gridID string
		var count int
		if err := rows.Scan(&gridID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan grid stat: %w", err)
		}
		gridCounts[gridID] = count
		counts = append(counts, count)
	}

	if len(counts) == 0 {
		return make(map[string]float64), nil
	}

	// Calculate percentile ranks
	sort.Ints(counts)
	overlapStats := make(map[string]float64)

	for gridID, count := range gridCounts {
		// Find percentile rank
		rank := 0.0
		for i, c := range counts {
			if c >= count {
				rank = float64(i) / float64(len(counts))
				break
			}
		}
		overlapStats[gridID] = rank
	}

	return overlapStats, nil
}

// getSpeedBucket determines speed bucket (0-5) based on percentiles
func (a *RenderingMetadataAnalyzer) getSpeedBucket(speed float64, percentiles []float64) int {
	for i := len(percentiles) - 1; i >= 0; i-- {
		if speed >= percentiles[i] {
			return i
		}
	}
	return 0
}

// calculateLineWeight calculates line weight hint based on overlap rank
func (a *RenderingMetadataAnalyzer) calculateLineWeight(overlapRank float64) float64 {
	// Higher overlap = thicker line
	// Range: 1.0 (low overlap) to 3.0 (high overlap)
	return 1.0 + (overlapRank * 2.0)
}

// calculateAlphaHint calculates alpha hint based on overlap rank
func (a *RenderingMetadataAnalyzer) calculateAlphaHint(overlapRank float64) float64 {
	// Higher overlap = more opaque
	// Range: 0.3 (low overlap) to 1.0 (high overlap)
	return 0.3 + (overlapRank * 0.7)
}

// insertRenderMetadata inserts render metadata into the database
func (a *RenderingMetadataAnalyzer) insertRenderMetadata(ctx context.Context, metadata []RenderMetadata) error {
	if len(metadata) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT OR REPLACE INTO render_segments_cache (
			segment_id, lod, speed_bucket, overlap_rank, line_weight_hint, alpha_hint, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range metadata {
		_, err := stmt.ExecContext(ctx,
			m.SegmentID,
			m.LOD,
			m.SpeedBucket,
			m.OverlapRank,
			m.LineWeight,
			m.AlphaHint,
		)
		if err != nil {
			return fmt.Errorf("failed to insert render metadata: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("rendering_metadata", NewRenderingMetadataAnalyzer)
}
