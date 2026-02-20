package behavior

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// StreakDetectionAnalyzer implements streak detection
// Skill: 连续活动检测 (Streak Detection)
// Detects consecutive days with activity
type StreakDetectionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewStreakDetectionAnalyzer creates a new streak detection analyzer
func NewStreakDetectionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &StreakDetectionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "streak_detection", 50000),
	}
}

// Analyze performs streak detection
func (a *StreakDetectionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[StreakDetectionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing streaks (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM streaks"); err != nil {
			return fmt.Errorf("failed to clear streaks: %w", err)
		}
		log.Printf("[StreakDetectionAnalyzer] Cleared existing streaks")
	}

	// Get daily activity statistics
	dailyStatsQuery := `
		SELECT
			DATE(datetime(dataTime, 'unixepoch')) as date,
			SUM(distance) as total_distance,
			COUNT(*) as point_count,
			MAX(dataTime) - MIN(dataTime) as duration
		FROM "一生足迹"
		WHERE outlier_flag = 0
		GROUP BY date
		ORDER BY date
	`

	rows, err := a.DB.QueryContext(ctx, dailyStatsQuery)
	if err != nil {
		return fmt.Errorf("failed to query daily stats: %w", err)
	}
	defer rows.Close()

	var dailyStats []DailyStats
	for rows.Next() {
		var stats DailyStats
		var distance sql.NullFloat64
		var duration sql.NullInt64

		if err := rows.Scan(&stats.Date, &distance, &stats.PointCount, &duration); err != nil {
			return fmt.Errorf("failed to scan daily stats: %w", err)
		}

		if distance.Valid {
			stats.TotalDistance = distance.Float64
		}
		if duration.Valid {
			stats.Duration = duration.Int64
		}

		dailyStats = append(dailyStats, stats)
	}

	if len(dailyStats) == 0 {
		log.Printf("[StreakDetectionAnalyzer] No daily stats to process")
		return a.MarkTaskAsCompleted(taskID, `{"streaks": 0}`)
	}

	log.Printf("[StreakDetectionAnalyzer] Processing %d days", len(dailyStats))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(dailyStats)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Detect streaks
	minActivityDistance := 1000.0 // 1 km minimum activity
	streaks := a.detectStreaks(dailyStats, minActivityDistance)

	// Insert streaks
	if err := a.insertStreaks(ctx, streaks); err != nil {
		return fmt.Errorf("failed to insert streaks: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_days": len(dailyStats),
		"streaks":    len(streaks),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[StreakDetectionAnalyzer] Analysis completed: %d days processed, %d streaks detected", len(dailyStats), len(streaks))
	return nil
}

// DailyStats holds daily statistics
type DailyStats struct {
	Date          string
	TotalDistance float64
	PointCount    int64
	Duration      int64
}

// Streak holds streak data
type Streak struct {
	StartDate     string
	EndDate       string
	DaysCount     int
	TotalDistance float64
	TotalDuration int64
}

// detectStreaks detects consecutive day streaks
func (a *StreakDetectionAnalyzer) detectStreaks(dailyStats []DailyStats, minActivityDistance float64) []Streak {
	if len(dailyStats) == 0 {
		return nil
	}

	var streaks []Streak
	var currentStreak *Streak
	var streakDays []DailyStats

	for i, stats := range dailyStats {
		// Check if day has sufficient activity
		if stats.TotalDistance < minActivityDistance {
			// End current streak if exists
			if currentStreak != nil {
				if currentStreak.DaysCount >= 2 { // Only keep streaks of 2+ days
					streaks = append(streaks, *currentStreak)
				}
				currentStreak = nil
				streakDays = nil
			}
			continue
		}

		// Parse date
		currentDate, err := time.Parse("2006-01-02", stats.Date)
		if err != nil {
			continue
		}

		if currentStreak == nil {
			// Start new streak
			currentStreak = &Streak{
				StartDate:     stats.Date,
				EndDate:       stats.Date,
				DaysCount:     1,
				TotalDistance: stats.TotalDistance,
				TotalDuration: stats.Duration,
			}
			streakDays = []DailyStats{stats}
		} else {
			// Check if consecutive day
			prevDate, err := time.Parse("2006-01-02", streakDays[len(streakDays)-1].Date)
			if err != nil {
				continue
			}

			daysDiff := int(currentDate.Sub(prevDate).Hours() / 24)

			if daysDiff == 1 {
				// Continue streak
				currentStreak.EndDate = stats.Date
				currentStreak.DaysCount++
				currentStreak.TotalDistance += stats.TotalDistance
				currentStreak.TotalDuration += stats.Duration
				streakDays = append(streakDays, stats)
			} else {
				// Gap detected - end current streak
				if currentStreak.DaysCount >= 2 {
					streaks = append(streaks, *currentStreak)
				}

				// Start new streak
				currentStreak = &Streak{
					StartDate:     stats.Date,
					EndDate:       stats.Date,
					DaysCount:     1,
					TotalDistance: stats.TotalDistance,
					TotalDuration: stats.Duration,
				}
				streakDays = []DailyStats{stats}
			}
		}

		// Handle last day
		if i == len(dailyStats)-1 && currentStreak != nil && currentStreak.DaysCount >= 2 {
			streaks = append(streaks, *currentStreak)
		}
	}

	return streaks
}

// insertStreaks inserts streaks into the database
func (a *StreakDetectionAnalyzer) insertStreaks(ctx context.Context, streaks []Streak) error {
	if len(streaks) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO streaks (
			start_date, end_date, days_count, total_distance_m, total_duration_s,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, streak := range streaks {
		_, err := stmt.ExecContext(ctx,
			streak.StartDate,
			streak.EndDate,
			streak.DaysCount,
			streak.TotalDistance,
			streak.TotalDuration,
		)
		if err != nil {
			return fmt.Errorf("failed to insert streak: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[StreakDetectionAnalyzer] Inserted %d streaks", len(streaks))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("streak_detection", NewStreakDetectionAnalyzer)
}
