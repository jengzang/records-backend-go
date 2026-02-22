package spatial

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/stats"
)

// RevisitAnalyzer implements the Revisit Patterns analysis skill
type RevisitAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// RevisitLocation represents a location with revisit statistics
type RevisitLocation struct {
	Geohash         string
	Lat             float64
	Lon             float64
	Province        string
	City            string
	County          string
	VisitCount      int
	FirstVisit      int64
	LastVisit       int64
	TotalDuration   int
	AvgInterval     float64
	StdInterval     float64
	MinInterval     float64
	MaxInterval     float64
	RegularityScore float64
	IsPeriodic      bool
	IsHabitual      bool
	RevisitStrength float64
}

// NewRevisitAnalyzer creates a new revisit patterns analyzer
func NewRevisitAnalyzer(db *sql.DB) analysis.Analyzer {
	return &RevisitAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "revisit_pattern", 1000),
	}
}

// Analyze performs the revisit patterns analysis
func (a *RevisitAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("Starting revisit patterns analysis (task: %d, mode: %s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Query stay_segments grouped by geohash6
	query := `
		SELECT
			geohash6,
			AVG(center_lat) as lat,
			AVG(center_lon) as lon,
			COUNT(*) as visit_count,
			MIN(start_time) as first_visit,
			MAX(end_time) as last_visit,
			SUM(duration_s) as total_duration
		FROM stay_segments
		WHERE geohash6 IS NOT NULL AND geohash6 != ''
		GROUP BY geohash6
		HAVING visit_count >= 2
		ORDER BY visit_count DESC
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		a.MarkTaskAsFailed(taskID, fmt.Sprintf("Query failed: %v", err))
		return fmt.Errorf("failed to query stay_segments: %w", err)
	}
	defer rows.Close()

	locations := []RevisitLocation{}
	processed := 0

	for rows.Next() {
		var loc RevisitLocation
		err := rows.Scan(&loc.Geohash, &loc.Lat, &loc.Lon, &loc.VisitCount,
			&loc.FirstVisit, &loc.LastVisit, &loc.TotalDuration)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Get admin info from first stay in this location
		adminQuery := `
			SELECT province, city, county
			FROM stay_segments
			WHERE geohash6 = ?
			LIMIT 1
		`
		err = a.DB.QueryRow(adminQuery, loc.Geohash).Scan(&loc.Province, &loc.City, &loc.County)
		if err != nil {
			// Admin info is optional, just log and continue
			log.Printf("Failed to get admin info for %s: %v", loc.Geohash, err)
		}

		// Calculate revisit intervals
		intervals, minInterval, maxInterval, err := a.getIntervals(loc.Geohash)
		if err != nil {
			log.Printf("Failed to get intervals for %s: %v", loc.Geohash, err)
			continue
		}

		if len(intervals) > 0 {
			loc.AvgInterval = stats.Mean(intervals)
			loc.StdInterval = stats.StdDev(intervals)
			loc.MinInterval = minInterval
			loc.MaxInterval = maxInterval

			// Calculate regularity score (0-1, higher = more regular)
			if loc.AvgInterval > 0 {
				cv := loc.StdInterval / loc.AvgInterval
				loc.RegularityScore = 1.0 / (1.0 + cv)
			}

			// Mark as periodic if highly regular (low coefficient of variation)
			loc.IsPeriodic = loc.RegularityScore > 0.8 && loc.VisitCount >= 3

			// Mark as habitual if visited frequently and regularly
			loc.IsHabitual = loc.VisitCount >= 5 && loc.RegularityScore > 0.7
		}

		// Calculate revisit strength: log(1 + visits) × log(1 + duration)
		loc.RevisitStrength = math.Log(1+float64(loc.VisitCount)) * math.Log(1+float64(loc.TotalDuration))

		locations = append(locations, loc)
		processed++

		// Update progress periodically
		if processed%100 == 0 {
			a.UpdateTaskProgress(taskID, int64(processed), int64(len(locations)), 0)
		}
	}

	if err := rows.Err(); err != nil {
		a.MarkTaskAsFailed(taskID, fmt.Sprintf("Row iteration failed: %v", err))
		return fmt.Errorf("row iteration failed: %w", err)
	}

	log.Printf("Processed %d revisit locations", len(locations))

	// Convert locations to []interface{} for insertResults
	locationsInterface := make([]interface{}, len(locations))
	for i, loc := range locations {
		locationsInterface[i] = loc
	}

	// Insert results into database
	if err := a.insertResults(locationsInterface); err != nil {
		a.MarkTaskAsFailed(taskID, fmt.Sprintf("Insert failed: %v", err))
		return fmt.Errorf("failed to insert results: %w", err)
	}

	// Mark task as completed
	if err := a.MarkTaskAsCompleted(taskID); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("Revisit patterns analysis completed (task: %d)", taskID)
	return nil
}

// getIntervals calculates the time intervals between visits to a location
func (a *RevisitAnalyzer) getIntervals(geohash string) ([]float64, float64, float64, error) {
	query := `
		SELECT start_time
		FROM stay_segments
		WHERE geohash6 = ?
		ORDER BY start_time
	`

	rows, err := a.DB.Query(query, geohash)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	var times []int64
	for rows.Next() {
		var t int64
		if err := rows.Scan(&t); err != nil {
			return nil, 0, 0, err
		}
		times = append(times, t)
	}

	if len(times) < 2 {
		return nil, 0, 0, nil
	}

	// Calculate intervals in days
	intervals := make([]float64, len(times)-1)
	minInterval := math.MaxFloat64
	maxInterval := 0.0

	for i := 1; i < len(times); i++ {
		intervalSeconds := times[i] - times[i-1]
		intervalDays := float64(intervalSeconds) / 86400.0
		intervals[i-1] = intervalDays

		if intervalDays < minInterval {
			minInterval = intervalDays
		}
		if intervalDays > maxInterval {
			maxInterval = intervalDays
		}
	}

	return intervals, minInterval, maxInterval, nil
}

// insertResults inserts the analysis results into the database
func (a *RevisitAnalyzer) insertResults(locations []interface{}) error {
	// Clear existing results (for full recompute)
	_, err := a.DB.Exec("DELETE FROM revisit_patterns")
	if err != nil {
		return fmt.Errorf("failed to clear existing results: %w", err)
	}

	// Prepare insert statement
	stmt, err := a.DB.Prepare(`
		INSERT INTO revisit_patterns (
			geohash6, center_lat, center_lon,
			province, city, county,
			visit_count, first_visit, last_visit, total_duration_seconds,
			avg_interval_days, std_interval_days, min_interval_days, max_interval_days,
			regularity_score, is_periodic, is_habitual, revisit_strength,
			algo_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1')
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert each location
	for _, locInterface := range locations {
		loc, ok := locInterface.(RevisitLocation)
		if !ok {
			log.Printf("Failed to cast location to RevisitLocation")
			continue
		}

		isPeriodic := 0
		if loc.IsPeriodic {
			isPeriodic = 1
		}
		isHabitual := 0
		if loc.IsHabitual {
			isHabitual = 1
		}

		_, err := stmt.Exec(
			loc.Geohash, loc.Lat, loc.Lon,
			loc.Province, loc.City, loc.County,
			loc.VisitCount, loc.FirstVisit, loc.LastVisit, loc.TotalDuration,
			loc.AvgInterval, loc.StdInterval, loc.MinInterval, loc.MaxInterval,
			loc.RegularityScore, isPeriodic, isHabitual, loc.RevisitStrength,
		)
		if err != nil {
			log.Printf("Failed to insert location %s: %v", loc.Geohash, err)
			continue
		}
	}

	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("revisit_pattern", NewRevisitAnalyzer)
}

// calculateIntervals calculates intervals between visit times
func calculateIntervals(times []int64) ([]float64, float64, float64) {
	if len(times) < 2 {
		return nil, 0, 0
	}

	intervals := make([]float64, len(times)-1)
	minInterval := math.MaxFloat64
	maxInterval := 0.0

	for i := 1; i < len(times); i++ {
		intervalSeconds := times[i] - times[i-1]
		intervalDays := float64(intervalSeconds) / 86400.0
		intervals[i-1] = intervalDays

		if intervalDays < minInterval {
			minInterval = intervalDays
		}
		if intervalDays > maxInterval {
			maxInterval = intervalDays
		}
	}

	return intervals, minInterval, maxInterval
}
