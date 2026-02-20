package stats

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// FootprintAnalyzer implements footprint statistics aggregation
// Skill: 足迹层统计与排行 (Footprint Analytics)
// Aggregates track points by administrative areas, time ranges, and grids
type FootprintAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewFootprintAnalyzer creates a new footprint statistics analyzer
func NewFootprintAnalyzer(db *sql.DB) analysis.Analyzer {
	return &FootprintAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "footprint_statistics", 1000),
	}
}

// Analyze performs footprint statistics aggregation
func (a *FootprintAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[FootprintAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Get last processed point ID for incremental mode
	var lastProcessedID int64
	if mode == "incremental" {
		lastProcessedID = a.GetLastProcessedID(taskID)
		log.Printf("[FootprintAnalyzer] Incremental mode: starting from point_id=%d", lastProcessedID)
	}

	// Count total points to process
	countQuery := `
		SELECT COUNT(*)
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND id > ?
	`
	var totalPoints int64
	if err := a.DB.QueryRowContext(ctx, countQuery, lastProcessedID).Scan(&totalPoints); err != nil {
		return fmt.Errorf("failed to count points: %w", err)
	}

	log.Printf("[FootprintAnalyzer] Total points to process: %d", totalPoints)

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, totalPoints, 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Process in batches
	batchSize := 10000
	processed := int64(0)

	for {
		// Query batch of points
		query := `
			SELECT
				id,
				dataTime,
				province,
				city,
				county,
				town,
				grid_id,
				distance,
				strftime('%Y', datetime(dataTime, 'unixepoch')) as year,
				strftime('%Y-%m', datetime(dataTime, 'unixepoch')) as month,
				strftime('%Y-%m-%d', datetime(dataTime, 'unixepoch')) as day
			FROM "一生足迹"
			WHERE outlier_flag = 0
				AND id > ?
			ORDER BY id
			LIMIT ?
		`

		rows, err := a.DB.QueryContext(ctx, query, lastProcessedID, batchSize)
		if err != nil {
			return fmt.Errorf("failed to query points: %w", err)
		}

		// Aggregate statistics
		stats := make(map[string]*FootprintStat)
		batchCount := 0
		maxID := lastProcessedID

		for rows.Next() {
			var (
				id                                     int64
				dataTime                               int64
				province, city, county, town, grid_id  sql.NullString
				distance                               sql.NullFloat64
				year, month, day                       string
			)

			if err := rows.Scan(&id, &dataTime, &province, &city, &county, &town, &grid_id, &distance, &year, &month, &day); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan row: %w", err)
			}

			batchCount++
			if id > maxID {
				maxID = id
			}

			// Aggregate by province
			if province.Valid && province.String != "" {
				a.aggregatePoint(stats, "PROVINCE", province.String, year, dataTime, distance.Float64)
				a.aggregatePoint(stats, "PROVINCE", province.String, month, dataTime, distance.Float64)
				a.aggregatePoint(stats, "PROVINCE", province.String, day, dataTime, distance.Float64)
				a.aggregatePoint(stats, "PROVINCE", province.String, "all", dataTime, distance.Float64)
			}

			// Aggregate by city
			if city.Valid && city.String != "" {
				a.aggregatePoint(stats, "CITY", city.String, year, dataTime, distance.Float64)
				a.aggregatePoint(stats, "CITY", city.String, month, dataTime, distance.Float64)
				a.aggregatePoint(stats, "CITY", city.String, day, dataTime, distance.Float64)
				a.aggregatePoint(stats, "CITY", city.String, "all", dataTime, distance.Float64)
			}

			// Aggregate by county
			if county.Valid && county.String != "" {
				a.aggregatePoint(stats, "COUNTY", county.String, year, dataTime, distance.Float64)
				a.aggregatePoint(stats, "COUNTY", county.String, month, dataTime, distance.Float64)
				a.aggregatePoint(stats, "COUNTY", county.String, day, dataTime, distance.Float64)
				a.aggregatePoint(stats, "COUNTY", county.String, "all", dataTime, distance.Float64)
			}

			// Aggregate by town
			if town.Valid && town.String != "" {
				a.aggregatePoint(stats, "TOWN", town.String, year, dataTime, distance.Float64)
				a.aggregatePoint(stats, "TOWN", town.String, month, dataTime, distance.Float64)
				a.aggregatePoint(stats, "TOWN", town.String, day, dataTime, distance.Float64)
				a.aggregatePoint(stats, "TOWN", town.String, "all", dataTime, distance.Float64)
			}

			// Aggregate by grid
			if grid_id.Valid && grid_id.String != "" {
				a.aggregatePoint(stats, "GRID", grid_id.String, year, dataTime, distance.Float64)
				a.aggregatePoint(stats, "GRID", grid_id.String, month, dataTime, distance.Float64)
				a.aggregatePoint(stats, "GRID", grid_id.String, day, dataTime, distance.Float64)
				a.aggregatePoint(stats, "GRID", grid_id.String, "all", dataTime, distance.Float64)
			}
		}
		rows.Close()

		// No more data
		if batchCount == 0 {
			break
		}

		// Insert/update statistics
		if err := a.upsertStatistics(ctx, stats); err != nil {
			return fmt.Errorf("failed to upsert statistics: %w", err)
		}

		// Update progress
		processed += int64(batchCount)
		lastProcessedID = maxID
		if err := a.UpdateTaskProgress(taskID, totalPoints, processed, 0); err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}

		log.Printf("[FootprintAnalyzer] Processed %d/%d points (%.1f%%)", processed, totalPoints, float64(processed)/float64(totalPoints)*100)

		// Check if we've processed all points
		if batchCount < batchSize {
			break
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points":     totalPoints,
		"processed_points": processed,
		"statistics_count": len(a.getStatisticsCount(ctx)),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[FootprintAnalyzer] Analysis completed: %d points processed", processed)
	return nil
}

// FootprintStat holds aggregated statistics for a specific stat_type + stat_key + time_range
type FootprintStat struct {
	StatType       string
	StatKey        string
	TimeRange      string
	PointCount     int64
	VisitDays      map[string]bool // Track unique days
	FirstVisit     int64
	LastVisit      int64
	TotalDistance  float64
	TotalDuration  int64
}

// aggregatePoint adds a point to the statistics
func (a *FootprintAnalyzer) aggregatePoint(stats map[string]*FootprintStat, statType, statKey, timeRange string, timestamp int64, distance float64) {
	key := fmt.Sprintf("%s|%s|%s", statType, statKey, timeRange)

	stat, exists := stats[key]
	if !exists {
		stat = &FootprintStat{
			StatType:   statType,
			StatKey:    statKey,
			TimeRange:  timeRange,
			VisitDays:  make(map[string]bool),
			FirstVisit: timestamp,
			LastVisit:  timestamp,
		}
		stats[key] = stat
	}

	// Update statistics
	stat.PointCount++
	stat.TotalDistance += distance

	// Track unique days
	day := time.Unix(timestamp, 0).Format("2006-01-02")
	stat.VisitDays[day] = true

	// Update first/last visit
	if timestamp < stat.FirstVisit {
		stat.FirstVisit = timestamp
	}
	if timestamp > stat.LastVisit {
		stat.LastVisit = timestamp
	}
}

// upsertStatistics inserts or updates statistics in the database
func (a *FootprintAnalyzer) upsertStatistics(ctx context.Context, stats map[string]*FootprintStat) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	upsertQuery := `
		INSERT INTO footprint_statistics (
			stat_type, stat_key, time_range,
			point_count, visit_count, first_visit, last_visit,
			total_distance_m, total_duration_s, metadata, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(stat_type, stat_key, time_range) DO UPDATE SET
			point_count = point_count + excluded.point_count,
			visit_count = visit_count + excluded.visit_count,
			first_visit = MIN(first_visit, excluded.first_visit),
			last_visit = MAX(last_visit, excluded.last_visit),
			total_distance_m = total_distance_m + excluded.total_distance_m,
			total_duration_s = total_duration_s + excluded.total_duration_s,
			metadata = excluded.metadata,
			updated_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, stat := range stats {
		visitCount := int64(len(stat.VisitDays))
		duration := stat.LastVisit - stat.FirstVisit

		// Create metadata JSON
		metadata := fmt.Sprintf(`{"visit_days":%d}`, visitCount)

		_, err := stmt.ExecContext(ctx,
			stat.StatType,
			stat.StatKey,
			stat.TimeRange,
			stat.PointCount,
			visitCount,
			stat.FirstVisit,
			stat.LastVisit,
			stat.TotalDistance,
			duration,
			metadata,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert statistic: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getStatisticsCount returns the count of statistics by type
func (a *FootprintAnalyzer) getStatisticsCount(ctx context.Context) map[string]int64 {
	query := `
		SELECT stat_type, COUNT(*)
		FROM footprint_statistics
		GROUP BY stat_type
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var statType string
		var count int64
		if err := rows.Scan(&statType, &count); err != nil {
			continue
		}
		counts[statType] = count
	}

	return counts
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("footprint_statistics", NewFootprintAnalyzer)
}
