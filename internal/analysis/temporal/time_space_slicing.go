package temporal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// TimeSpaceSlicingAnalyzer implements time-space slicing analysis
// Skill: 时空切片 (Time-Space Slicing)
// Slices trajectory by time and space dimensions for aggregation
type TimeSpaceSlicingAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewTimeSpaceSlicingAnalyzer creates a new time-space slicing analyzer
func NewTimeSpaceSlicingAnalyzer(db *sql.DB) analysis.Analyzer {
	return &TimeSpaceSlicingAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "time_space_slicing", 10000),
	}
}

// Analyze performs time-space slicing
func (a *TimeSpaceSlicingAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[TimeSpaceSlicingAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing slices (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM time_space_slices"); err != nil {
			return fmt.Errorf("failed to clear time_space_slices: %w", err)
		}
		log.Printf("[TimeSpaceSlicingAnalyzer] Cleared existing time-space slices")
	}

	var allSlices []TimeSpaceSlice

	// 1. Hourly slices
	hourlySlices, err := a.computeHourlySlices(ctx)
	if err != nil {
		return fmt.Errorf("failed to compute hourly slices: %w", err)
	}
	allSlices = append(allSlices, hourlySlices...)

	// 2. Daily slices
	dailySlices, err := a.computeDailySlices(ctx)
	if err != nil {
		return fmt.Errorf("failed to compute daily slices: %w", err)
	}
	allSlices = append(allSlices, dailySlices...)

	log.Printf("[TimeSpaceSlicingAnalyzer] Generated %d slices", len(allSlices))

	// Insert slices
	if err := a.insertTimeSpaceSlices(ctx, allSlices); err != nil {
		return fmt.Errorf("failed to insert time-space slices: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_slices":  len(allSlices),
		"hourly_slices": len(hourlySlices),
		"daily_slices":  len(dailySlices),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[TimeSpaceSlicingAnalyzer] Analysis completed")
	return nil
}

// TimeSpaceSlice holds time-space slice data
type TimeSpaceSlice struct {
	SliceType       string
	SliceKey        string
	AdminLevel      string
	AdminName       string
	GridID          string
	PointCount      int64
	Distance        float64
	Duration        int64
	UniqueLocations int64
}

// computeHourlySlices computes hourly time slices
func (a *TimeSpaceSlicingAnalyzer) computeHourlySlices(ctx context.Context) ([]TimeSpaceSlice, error) {
	query := `
		SELECT
			CAST(strftime('%H', datetime(dataTime, 'unixepoch')) AS INTEGER) as hour,
			COUNT(*) as point_count,
			SUM(CASE WHEN distance IS NOT NULL THEN distance ELSE 0 END) as total_distance,
			COUNT(DISTINCT grid_id) as unique_locations
		FROM "一生足迹"
		WHERE outlier_flag = 0
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly slices: %w", err)
	}
	defer rows.Close()

	var slices []TimeSpaceSlice
	for rows.Next() {
		var slice TimeSpaceSlice
		var hour int
		var distance sql.NullFloat64

		if err := rows.Scan(&hour, &slice.PointCount, &distance, &slice.UniqueLocations); err != nil {
			return nil, fmt.Errorf("failed to scan hourly slice: %w", err)
		}

		if distance.Valid {
			slice.Distance = distance.Float64
		}

		slice.SliceType = "HOURLY"
		slice.SliceKey = fmt.Sprintf("%02d", hour)
		slice.Duration = slice.PointCount * 10 // Approximate

		slices = append(slices, slice)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return slices, nil
}

// computeDailySlices computes daily time slices
func (a *TimeSpaceSlicingAnalyzer) computeDailySlices(ctx context.Context) ([]TimeSpaceSlice, error) {
	query := `
		SELECT
			DATE(datetime(dataTime, 'unixepoch')) as date,
			COUNT(*) as point_count,
			SUM(CASE WHEN distance IS NOT NULL THEN distance ELSE 0 END) as total_distance,
			COUNT(DISTINCT grid_id) as unique_locations
		FROM "一生足迹"
		WHERE outlier_flag = 0
		GROUP BY date
		ORDER BY date
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily slices: %w", err)
	}
	defer rows.Close()

	var slices []TimeSpaceSlice
	for rows.Next() {
		var slice TimeSpaceSlice
		var date string
		var distance sql.NullFloat64

		if err := rows.Scan(&date, &slice.PointCount, &distance, &slice.UniqueLocations); err != nil {
			return nil, fmt.Errorf("failed to scan daily slice: %w", err)
		}

		if distance.Valid {
			slice.Distance = distance.Float64
		}

		slice.SliceType = "DAILY"
		slice.SliceKey = date
		slice.Duration = slice.PointCount * 10

		slices = append(slices, slice)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return slices, nil
}

// insertTimeSpaceSlices inserts time-space slices into the database
func (a *TimeSpaceSlicingAnalyzer) insertTimeSpaceSlices(ctx context.Context, slices []TimeSpaceSlice) error {
	if len(slices) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO time_space_slices (
			slice_type, slice_key, admin_level, admin_name, grid_id,
			point_count, distance_m, duration_s, unique_locations,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
		ON CONFLICT(slice_type, slice_key, admin_level, admin_name, grid_id) DO UPDATE SET
			point_count = excluded.point_count,
			distance_m = excluded.distance_m,
			duration_s = excluded.duration_s,
			unique_locations = excluded.unique_locations
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, slice := range slices {
		_, err := stmt.ExecContext(ctx,
			slice.SliceType, slice.SliceKey, slice.AdminLevel, slice.AdminName, slice.GridID,
			slice.PointCount, slice.Distance, slice.Duration, slice.UniqueLocations,
		)
		if err != nil {
			return fmt.Errorf("failed to insert time-space slice: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[TimeSpaceSlicingAnalyzer] Inserted %d time-space slices", len(slices))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("time_space_slicing", NewTimeSpaceSlicingAnalyzer)
}
