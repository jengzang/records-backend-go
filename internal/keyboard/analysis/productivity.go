package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// ProductivityAnalyzer handles productivity metrics analysis
type ProductivityAnalyzer struct {
	db *sql.DB
}

// NewProductivityAnalyzer creates a new productivity analyzer
func NewProductivityAnalyzer(db *sql.DB) *ProductivityAnalyzer {
	return &ProductivityAnalyzer{db: db}
}

// ActivityMetrics represents activity-based productivity metrics
type ActivityMetrics struct {
	TotalDays            int     `json:"totalDays"`
	ActiveDays           int     `json:"activeDays"`
	InactiveDays         int     `json:"inactiveDays"`
	ActivityRate         float64 `json:"activityRate"`
	CurrentStreak        int     `json:"currentStreak"`
	LongestStreak        int     `json:"longestStreak"`
	AvgKeystrokesPerDay  float64 `json:"avgKeystrokesPerDay"`
	StdDevKeystrokes     float64 `json:"stdDevKeystrokes"`
	ConsistencyScore     float64 `json:"consistencyScore"`
	ThresholdUsed        int     `json:"thresholdUsed"`
}

// IntensityMetrics represents typing intensity metrics
type IntensityMetrics struct {
	AvgKeystrokes  float64 `json:"avgKeystrokes"`
	AvgClicks      float64 `json:"avgClicks"`
	AvgDistance    float64 `json:"avgDistance"`
	PeakKeystrokes int64   `json:"peakKeystrokes"`
	PeakClicks     int64   `json:"peakClicks"`
	PeakDistance   float64 `json:"peakDistance"`
	P50Keystrokes  int64   `json:"p50Keystrokes"`
	P75Keystrokes  int64   `json:"p75Keystrokes"`
	P95Keystrokes  int64   `json:"p95Keystrokes"`
	ActiveDays     int     `json:"activeDays"`
}

// PeakDay represents a peak usage day
type PeakDay struct {
	Date       string  `json:"date"`
	Keystrokes int64   `json:"keystrokes"`
	Clicks     int64   `json:"clicks"`
	Distance   float64 `json:"distance"`
}

// AnalyzeActivityMetrics calculates activity-based productivity metrics
func (pa *ProductivityAnalyzer) AnalyzeActivityMetrics(threshold int, startDate, endDate string) (*ActivityMetrics, error) {
	query := `
		SELECT
			date,
			keystrokes,
			left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks
		FROM daily_stats
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY date"

	rows, err := pa.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily stats: %w", err)
	}
	defer rows.Close()

	metrics := &ActivityMetrics{
		ThresholdUsed: threshold,
	}

	var dailyKeystrokes []int64
	var prevDate *time.Time
	var tempStreak int
	var longestStreak int

	for rows.Next() {
		var dateStr string
		var keystrokes, clicks int64

		if err := rows.Scan(&dateStr, &keystrokes, &clicks); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		metrics.TotalDays++
		dailyKeystrokes = append(dailyKeystrokes, keystrokes)

		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}

		if keystrokes >= int64(threshold) {
			metrics.ActiveDays++
			tempStreak++

			// Check if consecutive
			if prevDate != nil {
				expectedDate := prevDate.AddDate(0, 0, 1)
				if !date.Equal(expectedDate) {
					// Streak broken
					if tempStreak > longestStreak {
						longestStreak = tempStreak
					}
					tempStreak = 1
				}
			}

			prevDate = &date
		} else {
			// Inactive day, reset streak
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
			tempStreak = 0
			prevDate = nil
		}
	}

	// Check final streak
	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	// Current streak is the temp_streak if it extends to the last day
	if prevDate != nil {
		metrics.CurrentStreak = tempStreak
	}
	metrics.LongestStreak = longestStreak

	metrics.InactiveDays = metrics.TotalDays - metrics.ActiveDays

	// Calculate activity rate
	if metrics.TotalDays > 0 {
		metrics.ActivityRate = float64(metrics.ActiveDays) / float64(metrics.TotalDays)
	}

	// Calculate average and standard deviation
	if len(dailyKeystrokes) > 0 {
		var sum int64
		for _, k := range dailyKeystrokes {
			sum += k
		}
		metrics.AvgKeystrokesPerDay = float64(sum) / float64(len(dailyKeystrokes))

		// Calculate standard deviation
		var variance float64
		for _, k := range dailyKeystrokes {
			diff := float64(k) - metrics.AvgKeystrokesPerDay
			variance += diff * diff
		}
		variance /= float64(len(dailyKeystrokes))
		metrics.StdDevKeystrokes = math.Sqrt(variance)

		// Calculate consistency score (coefficient of variation)
		if metrics.AvgKeystrokesPerDay > 0 {
			metrics.ConsistencyScore = metrics.StdDevKeystrokes / metrics.AvgKeystrokesPerDay
		}
	}

	return metrics, nil
}

// AnalyzeTypingIntensity calculates typing intensity metrics
func (pa *ProductivityAnalyzer) AnalyzeTypingIntensity(startDate, endDate string) (*IntensityMetrics, error) {
	query := `
		SELECT
			keystrokes,
			left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks,
			mouse_distance_m
		FROM daily_stats
		WHERE keystrokes > 0
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY keystrokes DESC"

	rows, err := pa.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query intensity metrics: %w", err)
	}
	defer rows.Close()

	var keystrokesList []int64
	var clicksList []int64
	var distanceList []float64

	for rows.Next() {
		var keystrokes, clicks int64
		var distance float64

		if err := rows.Scan(&keystrokes, &clicks, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		keystrokesList = append(keystrokesList, keystrokes)
		clicksList = append(clicksList, clicks)
		distanceList = append(distanceList, distance)
	}

	if len(keystrokesList) == 0 {
		return &IntensityMetrics{}, nil
	}

	metrics := &IntensityMetrics{
		ActiveDays: len(keystrokesList),
	}

	// Calculate averages
	var sumKeystrokes, sumClicks int64
	var sumDistance float64
	for i := range keystrokesList {
		sumKeystrokes += keystrokesList[i]
		sumClicks += clicksList[i]
		sumDistance += distanceList[i]
	}

	metrics.AvgKeystrokes = float64(sumKeystrokes) / float64(len(keystrokesList))
	metrics.AvgClicks = float64(sumClicks) / float64(len(clicksList))
	metrics.AvgDistance = sumDistance / float64(len(distanceList))

	// Peak values
	metrics.PeakKeystrokes = keystrokesList[0] // Already sorted DESC
	metrics.PeakClicks = clicksList[0]
	metrics.PeakDistance = distanceList[0]

	// Calculate percentiles
	metrics.P50Keystrokes = percentile(keystrokesList, 50)
	metrics.P75Keystrokes = percentile(keystrokesList, 75)
	metrics.P95Keystrokes = percentile(keystrokesList, 95)

	return metrics, nil
}

// AnalyzePeakDays finds peak usage days
func (pa *ProductivityAnalyzer) AnalyzePeakDays(limit int, startDate, endDate string) ([]PeakDay, error) {
	query := `
		SELECT
			date,
			keystrokes,
			left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks,
			mouse_distance_m
		FROM daily_stats
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY keystrokes DESC LIMIT ?"
	args = append(args, limit)

	rows, err := pa.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query peak days: %w", err)
	}
	defer rows.Close()

	var result []PeakDay
	for rows.Next() {
		var day PeakDay
		if err := rows.Scan(&day.Date, &day.Keystrokes, &day.Clicks, &day.Distance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, day)
	}

	return result, nil
}

// percentile calculates the percentile of a sorted slice
func percentile(data []int64, p int) int64 {
	if len(data) == 0 {
		return 0
	}

	// Sort data in ascending order for percentile calculation
	sorted := make([]int64, len(data))
	copy(sorted, data)

	// Simple bubble sort (data is already sorted DESC, so reverse it)
	for i := 0; i < len(sorted)/2; i++ {
		sorted[i], sorted[len(sorted)-1-i] = sorted[len(sorted)-1-i], sorted[i]
	}

	index := int(float64(len(sorted)) * float64(p) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
