package efficiency

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// fetchKeyboardData retrieves keyboard metrics for a specific hour
func (s *Service) fetchKeyboardData(date string, hour int) (*KeyboardMetrics, error) {
	// Convert date format from YYYY-MM-DD to YYYYMMDD
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	dateKey := t.Format("20060102")

	// Query daily keystrokes
	query := `SELECT keystrokes FROM keyboard_data WHERE date = ?`
	var dailyKeystrokes int
	err = s.keyboardDB.QueryRow(query, dateKey).Scan(&dailyKeystrokes)
	if err == sql.ErrNoRows {
		return nil, nil // No data for this date
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query keyboard data: %w", err)
	}

	// Since we only have daily data, estimate hourly distribution
	// Assume active hours are 8am-11pm (16 hours), distribute keystrokes across these hours
	// Other hours get 0
	var hourlyKeystrokes float64
	if hour >= 8 && hour <= 23 {
		hourlyKeystrokes = float64(dailyKeystrokes) / 16.0
	} else {
		hourlyKeystrokes = 0
	}

	// Normalize typing speed (0-100 scale)
	// Assume average typing speed is 1000 keystrokes/hour, max is 3000
	normalized := normalizeValue(hourlyKeystrokes, 0, 3000, 1000)

	return &KeyboardMetrics{
		TypingSpeed:           hourlyKeystrokes,
		TypingSpeedNormalized: normalized,
	}, nil
}

// fetchScreenTimeData retrieves screentime metrics for a specific hour
func (s *Service) fetchScreenTimeData(date string, hour int) (*ScreenTimeMetrics, error) {
	// Convert date format from YYYY-MM-DD to YYYYMMDD
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	dateKey := t.Format("20060102")

	// Calculate hour boundaries in Unix milliseconds
	hourStart := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	hourEnd := hourStart.Add(time.Hour)
	startMs := hourStart.UnixMilli()
	endMs := hourEnd.UnixMilli()

	// Query sessions that overlap with this hour
	query := `
		SELECT
			s.start_time_ms,
			s.end_time_ms,
			s.package_id,
			COALESCE(a.category, 'Other') as category
		FROM screentime_sessions s
		LEFT JOIN screentime_apps a ON s.package_id = a.package_id
		WHERE s.date = ?
		AND s.start_time_ms < ?
		AND s.end_time_ms > ?
	`

	rows, err := s.screentimeDB.Query(query, dateKey, endMs, startMs)
	if err != nil {
		return nil, fmt.Errorf("failed to query screentime data: %w", err)
	}
	defer rows.Close()

	var totalDuration int64
	var workDuration int64
	var entertainmentDuration int64
	var sessionCount int
	var focusSessions int
	var lastEndTime int64

	for rows.Next() {
		var sessionStart, sessionEnd int64
		var packageID, category string
		if err := rows.Scan(&sessionStart, &sessionEnd, &packageID, &category); err != nil {
			return nil, fmt.Errorf("failed to scan screentime row: %w", err)
		}

		// Calculate overlap duration with this hour
		overlapStart := max64(sessionStart, startMs)
		overlapEnd := min64(sessionEnd, endMs)
		overlapDuration := overlapEnd - overlapStart

		if overlapDuration > 0 {
			totalDuration += overlapDuration
			sessionCount++

			// Categorize as work or entertainment
			if isWorkCategory(category) {
				workDuration += overlapDuration
			} else if isEntertainmentCategory(category) {
				entertainmentDuration += overlapDuration
			}

			// Detect focus sessions (>30 minutes continuous)
			if overlapDuration > 30*60*1000 {
				focusSessions++
			}

			lastEndTime = sessionEnd
		}
	}

	if totalDuration == 0 {
		return nil, nil // No screentime data for this hour
	}

	// Calculate metrics
	workRatio := float64(workDuration) / float64(totalDuration)
	entertainmentRatio := float64(entertainmentDuration) / float64(totalDuration)
	appSwitchFreq := float64(sessionCount) // switches per hour

	// Normalize metrics
	workRatioNormalized := workRatio * 100 // Already 0-1, scale to 0-100
	focusNormalized := normalizeValue(float64(focusSessions), 0, 5, 2)

	return &ScreenTimeMetrics{
		WorkAppRatio:              workRatio,
		EntertainmentAppRatio:     entertainmentRatio,
		FocusSessionCount:         focusSessions,
		AppSwitchFrequency:        appSwitchFreq,
		WorkAppRatioNormalized:    workRatioNormalized,
		FocusNormalized:           focusNormalized,
	}, nil
}

// fetchHealthData retrieves health metrics for a specific hour
func (s *Service) fetchHealthData(date string, hour int) (*HealthMetrics, error) {
	// Parse date
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Calculate hour boundaries
	hourStart := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	hourEnd := hourStart.Add(time.Hour)

	// Query heart rate data
	hrQuery := `
		SELECT AVG(value) as avg_hr
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierHeartRate'
		AND start_date >= ?
		AND start_date < ?
	`
	var avgHeartRate sql.NullFloat64
	err = s.healthDB.QueryRow(hrQuery, hourStart.Format("2006-01-02 15:04:05"), hourEnd.Format("2006-01-02 15:04:05")).Scan(&avgHeartRate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query heart rate: %w", err)
	}

	// Query HRV data (if available)
	hrvQuery := `
		SELECT AVG(value) as avg_hrv
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierHeartRateVariabilitySDNN'
		AND start_date >= ?
		AND start_date < ?
	`
	var avgHRV sql.NullFloat64
	err = s.healthDB.QueryRow(hrvQuery, hourStart.Format("2006-01-02 15:04:05"), hourEnd.Format("2006-01-02 15:04:05")).Scan(&avgHRV)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query HRV: %w", err)
	}

	// Query step count
	stepsQuery := `
		SELECT SUM(value) as total_steps
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		AND start_date >= ?
		AND start_date < ?
	`
	var totalSteps sql.NullInt64
	err = s.healthDB.QueryRow(stepsQuery, hourStart.Format("2006-01-02 15:04:05"), hourEnd.Format("2006-01-02 15:04:05")).Scan(&totalSteps)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query steps: %w", err)
	}

	// If no health data at all, return nil
	if !avgHeartRate.Valid && !avgHRV.Valid && !totalSteps.Valid {
		return nil, nil
	}

	// Normalize metrics
	var hrvNormalized float64
	if avgHRV.Valid {
		// Higher HRV is better (indicates lower stress)
		// Typical range: 20-100ms, optimal: 50-100ms
		hrvNormalized = normalizeValue(avgHRV.Float64, 20, 100, 50)
	}

	var activityNormalized float64
	if totalSteps.Valid {
		// Normalize steps (0-1000 steps per hour, optimal: 500)
		activityNormalized = normalizeValue(float64(totalSteps.Int64), 0, 1000, 500)
	}

	return &HealthMetrics{
		AvgHeartRate:          avgHeartRate.Float64,
		HeartRateVariability:  avgHRV.Float64,
		StepCount:             int(totalSteps.Int64),
		HRVNormalized:         hrvNormalized,
		ActivityNormalized:    activityNormalized,
	}, nil
}

// Helper functions

// normalizeValue normalizes a value to 0-100 scale using a sigmoid-like function
// min: minimum expected value
// max: maximum expected value
// optimal: optimal value (gets score of 75)
func normalizeValue(value, min, max, optimal float64) float64 {
	if value <= min {
		return 0
	}
	if value >= max {
		return 100
	}

	// Linear normalization with optimal point
	if value <= optimal {
		// 0 to optimal maps to 0 to 75
		return (value - min) / (optimal - min) * 75
	} else {
		// optimal to max maps to 75 to 100
		return 75 + (value-optimal)/(max-optimal)*25
	}
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// isWorkCategory checks if an app category is work-related
func isWorkCategory(category string) bool {
	workCategories := map[string]bool{
		"Productivity": true,
		"Business":     true,
		"Education":    true,
		"Tools":        true,
		"Development":  true,
	}
	return workCategories[category]
}

// isEntertainmentCategory checks if an app category is entertainment-related
func isEntertainmentCategory(category string) bool {
	entertainmentCategories := map[string]bool{
		"Entertainment": true,
		"Games":         true,
		"Social":        true,
		"Video":         true,
		"Music":         true,
		"Shopping":      true,
	}
	return entertainmentCategories[category]
}
