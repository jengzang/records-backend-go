package stats

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"records-backend/internal/analysis"
)

// StayAnalyzer implements stay statistics aggregation
// Skill: 停留层统计与排行 (Stay Analytics)
// Aggregates stay segments by administrative areas, time ranges, and stay types
type StayAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewStayAnalyzer creates a new stay statistics analyzer
func NewStayAnalyzer(db *sql.DB) analysis.Analyzer {
	return &StayAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "stay_statistics", 1000),
	}
}

// Analyze performs stay statistics aggregation
func (a *StayAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[StayAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Get last processed stay ID for incremental mode
	var lastProcessedID int64
	if mode == "incremental" {
		lastProcessedID = a.GetLastProcessedID(taskID)
		log.Printf("[StayAnalyzer] Incremental mode: starting from stay_id=%d", lastProcessedID)
	}

	// Count total stays to process
	countQuery := `
		SELECT COUNT(*)
		FROM stay_segments
		WHERE id > ?
	`
	var totalStays int64
	if err := a.DB.QueryRowContext(ctx, countQuery, lastProcessedID).Scan(&totalStays); err != nil {
		return fmt.Errorf("failed to count stays: %w", err)
	}

	log.Printf("[StayAnalyzer] Total stays to process: %d", totalStays)

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, totalStays, 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Process in batches
	batchSize := 1000
	processed := int64(0)

	for {
		// Query batch of stays
		query := `
			SELECT
				id,
				start_ts,
				end_ts,
				duration_s,
				stay_type,
				province,
				city,
				county,
				town,
				strftime('%Y', datetime(start_ts, 'unixepoch')) as year,
				strftime('%Y-%m', datetime(start_ts, 'unixepoch')) as month,
				strftime('%Y-%m-%d', datetime(start_ts, 'unixepoch')) as day,
				strftime('%H', datetime(start_ts, 'unixepoch')) as hour,
				strftime('%w', datetime(start_ts, 'unixepoch')) as weekday
			FROM stay_segments
			WHERE id > ?
			ORDER BY id
			LIMIT ?
		`

		rows, err := a.DB.QueryContext(ctx, query, lastProcessedID, batchSize)
		if err != nil {
			return fmt.Errorf("failed to query stays: %w", err)
		}

		// Aggregate statistics
		stats := make(map[string]*StayStat)
		batchCount := 0
		maxID := lastProcessedID

		for rows.Next() {
			var (
				id                                                int64
				startTS, endTS, durationS                         int64
				stayType                                          string
				province, city, county, town                      sql.NullString
				year, month, day, hour, weekday                   string
			)

			if err := rows.Scan(&id, &startTS, &endTS, &durationS, &stayType, &province, &city, &county, &town, &year, &month, &day, &hour, &weekday); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan row: %w", err)
			}

			batchCount++
			if id > maxID {
				maxID = id
			}

			// Aggregate by province
			if province.Valid && province.String != "" {
				a.aggregateStay(stats, "PROVINCE", province.String, year, day, durationS)
				a.aggregateStay(stats, "PROVINCE", province.String, month, day, durationS)
				a.aggregateStay(stats, "PROVINCE", province.String, day, day, durationS)
				a.aggregateStay(stats, "PROVINCE", province.String, "all", day, durationS)
			}

			// Aggregate by city
			if city.Valid && city.String != "" {
				a.aggregateStay(stats, "CITY", city.String, year, day, durationS)
				a.aggregateStay(stats, "CITY", city.String, month, day, durationS)
				a.aggregateStay(stats, "CITY", city.String, day, day, durationS)
				a.aggregateStay(stats, "CITY", city.String, "all", day, durationS)
			}

			// Aggregate by county
			if county.Valid && county.String != "" {
				a.aggregateStay(stats, "COUNTY", county.String, year, day, durationS)
				a.aggregateStay(stats, "COUNTY", county.String, month, day, durationS)
				a.aggregateStay(stats, "COUNTY", county.String, day, day, durationS)
				a.aggregateStay(stats, "COUNTY", county.String, "all", day, durationS)
			}

			// Aggregate by town
			if town.Valid && town.String != "" {
				a.aggregateStay(stats, "TOWN", town.String, year, day, durationS)
				a.aggregateStay(stats, "TOWN", town.String, month, day, durationS)
				a.aggregateStay(stats, "TOWN", town.String, day, day, durationS)
				a.aggregateStay(stats, "TOWN", town.String, "all", day, durationS)
			}

			// Aggregate by stay type
			a.aggregateStay(stats, "ACTIVITY_TYPE", stayType, year, day, durationS)
			a.aggregateStay(stats, "ACTIVITY_TYPE", stayType, month, day, durationS)
			a.aggregateStay(stats, "ACTIVITY_TYPE", stayType, day, day, durationS)
			a.aggregateStay(stats, "ACTIVITY_TYPE", stayType, "all", day, durationS)
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
		if err := a.UpdateTaskProgress(taskID, totalStays, processed, 0); err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}

		log.Printf("[StayAnalyzer] Processed %d/%d stays (%.1f%%)", processed, totalStays, float64(processed)/float64(totalStays)*100)

		// Check if we've processed all stays
		if batchCount < batchSize {
			break
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_stays":      totalStays,
		"processed_stays":  processed,
		"statistics_count": len(a.getStatisticsCount(ctx)),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[StayAnalyzer] Analysis completed: %d stays processed", processed)
	return nil
}

// StayStat holds aggregated statistics for a specific stat_type + stat_key + time_range
type StayStat struct {
	StatType      string
	StatKey       string
	TimeRange     string
	StayCount     int64
	VisitDays     map[string]bool // Track unique days
	TotalDuration int64
	MaxDuration   int64
}

// aggregateStay adds a stay to the statistics
func (a *StayAnalyzer) aggregateStay(stats map[string]*StayStat, statType, statKey, timeRange, day string, duration int64) {
	key := fmt.Sprintf("%s|%s|%s", statType, statKey, timeRange)

	stat, exists := stats[key]
	if !exists {
		stat = &StayStat{
			StatType:  statType,
			StatKey:   statKey,
			TimeRange: timeRange,
			VisitDays: make(map[string]bool),
		}
		stats[key] = stat
	}

	// Update statistics
	stat.StayCount++
	stat.TotalDuration += duration

	// Track unique days
	stat.VisitDays[day] = true

	// Update max duration
	if duration > stat.MaxDuration {
		stat.MaxDuration = duration
	}
}

// upsertStatistics inserts or updates statistics in the database
func (a *StayAnalyzer) upsertStatistics(ctx context.Context, stats map[string]*StayStat) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	upsertQuery := `
		INSERT INTO stay_statistics (
			stat_type, stat_key, time_range,
			stay_count, total_duration_s, avg_duration_s, max_duration_s,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(stat_type, stat_key, time_range) DO UPDATE SET
			stay_count = stay_count + excluded.stay_count,
			total_duration_s = total_duration_s + excluded.total_duration_s,
			avg_duration_s = (total_duration_s + excluded.total_duration_s) / (stay_count + excluded.stay_count),
			max_duration_s = MAX(max_duration_s, excluded.max_duration_s),
			updated_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, stat := range stats {
		avgDuration := float64(stat.TotalDuration) / float64(stat.StayCount)

		_, err := stmt.ExecContext(ctx,
			stat.StatType,
			stat.StatKey,
			stat.TimeRange,
			stat.StayCount,
			stat.TotalDuration,
			avgDuration,
			stat.MaxDuration,
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
func (a *StayAnalyzer) getStatisticsCount(ctx context.Context) map[string]int64 {
	query := `
		SELECT stat_type, COUNT(*)
		FROM stay_statistics
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
	analysis.RegisterAnalyzer("stay_statistics", NewStayAnalyzer)
}
