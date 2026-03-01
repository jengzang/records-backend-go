package analysis

import (
	"database/sql"
	"fmt"
	"time"
)

// HourlyPattern represents activity pattern for an hour
type HourlyPattern struct {
	Hour         int     `json:"hour"`
	AvgHeartRate float64 `json:"avgHeartRate"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
}

// DailyPattern represents 24-hour activity pattern
type DailyPattern struct {
	Hours          []HourlyPattern `json:"hours"`
	PeakHour       int             `json:"peakHour"`
	QuietestHour   int             `json:"quietestHour"`
	TotalReadings  int             `json:"totalReadings"`
}

// WeekdayPattern represents activity pattern for a day of week
type WeekdayPattern struct {
	Weekday      string  `json:"weekday"`
	AvgHeartRate float64 `json:"avgHeartRate"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
}

// WeeklyPattern represents weekly activity pattern
type WeeklyPattern struct {
	Days           []WeekdayPattern `json:"days"`
	MostActiveDay  string           `json:"mostActiveDay"`
	LeastActiveDay string           `json:"leastActiveDay"`
	TotalReadings  int              `json:"totalReadings"`
}

// PatternAnalyzer handles activity pattern detection
type PatternAnalyzer struct {
	db *sql.DB
}

// NewPatternAnalyzer creates a new pattern analyzer
func NewPatternAnalyzer(db *sql.DB) *PatternAnalyzer {
	return &PatternAnalyzer{db: db}
}

// GetDailyPattern analyzes hourly activity patterns
func (a *PatternAnalyzer) GetDailyPattern() (*DailyPattern, error) {
	query := `
		SELECT strftime('%H', start_date) as hour,
		       AVG(value) as avg_hr,
		       COUNT(*) as count
		FROM health_records
		WHERE type = 'HeartRate'
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily pattern: %w", err)
	}
	defer rows.Close()

	var hours []HourlyPattern
	totalReadings := 0
	maxCount := 0
	minCount := int(^uint(0) >> 1) // Max int
	peakHour := 0
	quietestHour := 0

	for rows.Next() {
		var h HourlyPattern
		var hourStr string
		if err := rows.Scan(&hourStr, &h.AvgHeartRate, &h.Count); err != nil {
			continue
		}

		fmt.Sscanf(hourStr, "%d", &h.Hour)
		totalReadings += h.Count

		if h.Count > maxCount {
			maxCount = h.Count
			peakHour = h.Hour
		}
		if h.Count < minCount {
			minCount = h.Count
			quietestHour = h.Hour
		}

		hours = append(hours, h)
	}

	// Calculate percentages
	for i := range hours {
		if totalReadings > 0 {
			hours[i].Percentage = float64(hours[i].Count) / float64(totalReadings) * 100
		}
	}

	return &DailyPattern{
		Hours:         hours,
		PeakHour:      peakHour,
		QuietestHour:  quietestHour,
		TotalReadings: totalReadings,
	}, nil
}

// GetWeeklyPattern analyzes weekly activity patterns
func (a *PatternAnalyzer) GetWeeklyPattern() (*WeeklyPattern, error) {
	query := `
		SELECT
			CASE CAST(strftime('%w', start_date) AS INTEGER)
				WHEN 0 THEN 'Sunday'
				WHEN 1 THEN 'Monday'
				WHEN 2 THEN 'Tuesday'
				WHEN 3 THEN 'Wednesday'
				WHEN 4 THEN 'Thursday'
				WHEN 5 THEN 'Friday'
				WHEN 6 THEN 'Saturday'
			END as weekday,
			AVG(value) as avg_hr,
			COUNT(*) as count
		FROM health_records
		WHERE type = 'HeartRate'
		GROUP BY strftime('%w', start_date)
		ORDER BY strftime('%w', start_date)
	`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query weekly pattern: %w", err)
	}
	defer rows.Close()

	var days []WeekdayPattern
	totalReadings := 0
	maxCount := 0
	minCount := int(^uint(0) >> 1) // Max int
	mostActiveDay := ""
	leastActiveDay := ""

	for rows.Next() {
		var d WeekdayPattern
		if err := rows.Scan(&d.Weekday, &d.AvgHeartRate, &d.Count); err != nil {
			continue
		}

		totalReadings += d.Count

		if d.Count > maxCount {
			maxCount = d.Count
			mostActiveDay = d.Weekday
		}
		if d.Count < minCount {
			minCount = d.Count
			leastActiveDay = d.Weekday
		}

		days = append(days, d)
	}

	// Calculate percentages
	for i := range days {
		if totalReadings > 0 {
			days[i].Percentage = float64(days[i].Count) / float64(totalReadings) * 100
		}
	}

	return &WeeklyPattern{
		Days:           days,
		MostActiveDay:  mostActiveDay,
		LeastActiveDay: leastActiveDay,
		TotalReadings:  totalReadings,
	}, nil
}

// GetActivityScore calculates daily activity score (0-100)
func (a *PatternAnalyzer) GetActivityScore(date time.Time) (float64, error) {
	dateStr := date.Format("2006-01-02")

	query := `
		SELECT COUNT(*) as measurement_count,
		       AVG(value) as avg_hr,
		       MAX(value) as max_hr,
		       MIN(value) as min_hr
		FROM health_records
		WHERE type = 'HeartRate'
		  AND DATE(start_date) = ?
	`

	var count int
	var avgHR, maxHR, minHR float64
	err := a.db.QueryRow(query, dateStr).Scan(&count, &avgHR, &maxHR, &minHR)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate activity score: %w", err)
	}

	if count == 0 {
		return 0, nil
	}

	// Calculate score based on:
	// 1. Measurement frequency (40 points): more measurements = more active
	// 2. Heart rate range (30 points): wider range = more varied activity
	// 3. Average heart rate (30 points): higher average = more active

	// Frequency score (normalize to 0-40, assuming 100+ measurements is excellent)
	frequencyScore := float64(count) / 100.0 * 40.0
	if frequencyScore > 40 {
		frequencyScore = 40
	}

	// Range score (normalize to 0-30, assuming 80+ BPM range is excellent)
	hrRange := maxHR - minHR
	rangeScore := hrRange / 80.0 * 30.0
	if rangeScore > 30 {
		rangeScore = 30
	}

	// Average HR score (normalize to 0-30, assuming 80+ BPM is active)
	avgScore := (avgHR - 60) / 40.0 * 30.0
	if avgScore < 0 {
		avgScore = 0
	}
	if avgScore > 30 {
		avgScore = 30
	}

	totalScore := frequencyScore + rangeScore + avgScore
	if totalScore > 100 {
		totalScore = 100
	}

	return totalScore, nil
}
