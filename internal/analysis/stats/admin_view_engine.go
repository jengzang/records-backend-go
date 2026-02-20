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

// AdminViewEngineAnalyzer implements multi-level administrative view statistics
// Skill: 行政区划视图引擎 (Admin View Engine)
// Generates hierarchical statistics by administrative level
type AdminViewEngineAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewAdminViewEngineAnalyzer creates a new admin view engine analyzer
func NewAdminViewEngineAnalyzer(db *sql.DB) analysis.Analyzer {
	return &AdminViewEngineAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "admin_view_engine", 10000),
	}
}

// Analyze performs admin view aggregation
func (a *AdminViewEngineAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[AdminViewEngineAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing stats (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM admin_stats"); err != nil {
			return fmt.Errorf("failed to clear admin_stats: %w", err)
		}
		log.Printf("[AdminViewEngineAnalyzer] Cleared existing admin stats")
	}

	// Process each admin level
	levels := []string{"PROVINCE", "CITY", "COUNTY", "TOWN"}
	totalStats := 0

	for _, level := range levels {
		stats, err := a.computeAdminStats(ctx, level)
		if err != nil {
			return fmt.Errorf("failed to compute stats for level %s: %w", level, err)
		}

		if err := a.insertAdminStats(ctx, stats); err != nil {
			return fmt.Errorf("failed to insert stats for level %s: %w", level, err)
		}

		totalStats += len(stats)
		log.Printf("[AdminViewEngineAnalyzer] Processed %s level: %d entries", level, len(stats))
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_stats": totalStats,
		"levels":      len(levels),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[AdminViewEngineAnalyzer] Analysis completed: %d admin stats generated", totalStats)
	return nil
}

// AdminStat holds administrative statistics
type AdminStat struct {
	AdminLevel     string
	AdminName      string
	ParentName     string
	VisitCount     int64
	TotalDuration  int64
	UniqueDays     int64
	FirstVisitTS   int64
	LastVisitTS    int64
	TotalDistance  float64
}

// computeAdminStats computes statistics for a specific admin level
func (a *AdminViewEngineAnalyzer) computeAdminStats(ctx context.Context, level string) ([]AdminStat, error) {
	var adminField, parentField string

	switch level {
	case "PROVINCE":
		adminField = "province"
		parentField = "NULL"
	case "CITY":
		adminField = "city"
		parentField = "province"
	case "COUNTY":
		adminField = "county"
		parentField = "city"
	case "TOWN":
		adminField = "town"
		parentField = "county"
	default:
		return nil, fmt.Errorf("invalid admin level: %s", level)
	}

	query := fmt.Sprintf(`
		SELECT
			%s as admin_name,
			%s as parent_name,
			COUNT(*) as visit_count,
			SUM(CASE WHEN distance IS NOT NULL THEN distance ELSE 0 END) as total_distance,
			COUNT(DISTINCT DATE(datetime(dataTime, 'unixepoch'))) as unique_days,
			MIN(dataTime) as first_visit_ts,
			MAX(dataTime) as last_visit_ts
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND %s IS NOT NULL
			AND %s != ''
		GROUP BY %s, %s
		ORDER BY visit_count DESC
	`, adminField, parentField, adminField, adminField, adminField, parentField)

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin stats: %w", err)
	}
	defer rows.Close()

	var stats []AdminStat
	for rows.Next() {
		var stat AdminStat
		var parentName sql.NullString
		var totalDistance sql.NullFloat64

		if err := rows.Scan(
			&stat.AdminName,
			&parentName,
			&stat.VisitCount,
			&totalDistance,
			&stat.UniqueDays,
			&stat.FirstVisitTS,
			&stat.LastVisitTS,
		); err != nil {
			return nil, fmt.Errorf("failed to scan admin stat: %w", err)
		}

		stat.AdminLevel = level
		if parentName.Valid {
			stat.ParentName = parentName.String
		}
		if totalDistance.Valid {
			stat.TotalDistance = totalDistance.Float64
		}

		// Calculate total duration (approximate based on point count and avg interval)
		// Assuming ~10 second intervals on average
		stat.TotalDuration = stat.VisitCount * 10

		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return stats, nil
}

// insertAdminStats inserts admin stats into the database
func (a *AdminViewEngineAnalyzer) insertAdminStats(ctx context.Context, stats []AdminStat) error {
	if len(stats) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO admin_stats (
			admin_level, admin_name, parent_name,
			visit_count, total_duration_s, unique_days,
			first_visit_ts, last_visit_ts, total_distance_m,
			algo_version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(admin_level, admin_name) DO UPDATE SET
			parent_name = excluded.parent_name,
			visit_count = excluded.visit_count,
			total_duration_s = excluded.total_duration_s,
			unique_days = excluded.unique_days,
			first_visit_ts = excluded.first_visit_ts,
			last_visit_ts = excluded.last_visit_ts,
			total_distance_m = excluded.total_distance_m,
			updated_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, stat := range stats {
		_, err := stmt.ExecContext(ctx,
			stat.AdminLevel, stat.AdminName, stat.ParentName,
			stat.VisitCount, stat.TotalDuration, stat.UniqueDays,
			stat.FirstVisitTS, stat.LastVisitTS, stat.TotalDistance,
		)
		if err != nil {
			return fmt.Errorf("failed to insert admin stat: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[AdminViewEngineAnalyzer] Inserted %d admin stats", len(stats))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("admin_view_engine", NewAdminViewEngineAnalyzer)
}