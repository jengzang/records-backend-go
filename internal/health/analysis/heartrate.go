package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// HeartRateZone represents a heart rate zone
type HeartRateZone struct {
	Name       string  `json:"name"`
	MinBPM     int     `json:"minBpm"`
	MaxBPM     int     `json:"maxBpm"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	TotalTime  float64 `json:"totalTime"` // in hours
}

// HeartRateZones represents distribution across zones
type HeartRateZones struct {
	StartDate time.Time        `json:"startDate"`
	EndDate   time.Time        `json:"endDate"`
	Zones     []HeartRateZone  `json:"zones"`
	TotalReadings int          `json:"totalReadings"`
}

// Anomaly represents an anomalous heart rate reading
type Anomaly struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Reason    string    `json:"reason"`
	Severity  string    `json:"severity"` // low, medium, high
}

// RestingHR represents daily resting heart rate
type RestingHR struct {
	Date       string  `json:"date"`
	RestingBPM float64 `json:"restingBpm"`
	MinBPM     float64 `json:"minBpm"`
	AvgBPM     float64 `json:"avgBpm"`
}

// HRVMetrics represents heart rate variability metrics
type HRVMetrics struct {
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	AvgHRV         float64   `json:"avgHrv"`
	StdDev         float64   `json:"stdDev"`
	RMSSD          float64   `json:"rmssd"` // Root mean square of successive differences
	Consistency    float64   `json:"consistency"` // 0-100 score
}

// HeartRateAnalyzer handles heart rate analysis
type HeartRateAnalyzer struct {
	db *sql.DB
}

// NewHeartRateAnalyzer creates a new heart rate analyzer
func NewHeartRateAnalyzer(db *sql.DB) *HeartRateAnalyzer {
	return &HeartRateAnalyzer{db: db}
}

// GetHeartRateZones calculates heart rate zone distribution
// Zones: Resting <60, Light 60-100, Moderate 100-140, Vigorous 140-170, Maximum >170
func (a *HeartRateAnalyzer) GetHeartRateZones(startDate, endDate time.Time) (*HeartRateZones, error) {
	zones := []HeartRateZone{
		{Name: "Resting", MinBPM: 0, MaxBPM: 60, Count: 0, TotalTime: 0},
		{Name: "Light", MinBPM: 60, MaxBPM: 100, Count: 0, TotalTime: 0},
		{Name: "Moderate", MinBPM: 100, MaxBPM: 140, Count: 0, TotalTime: 0},
		{Name: "Vigorous", MinBPM: 140, MaxBPM: 170, Count: 0, TotalTime: 0},
		{Name: "Maximum", MinBPM: 170, MaxBPM: 999, Count: 0, TotalTime: 0},
	}

	query := `
		SELECT value, start_date, end_date
		FROM health_records
		WHERE type = 'HeartRate'
		  AND start_date >= ?
		  AND end_date <= ?
		ORDER BY start_date
	`

	rows, err := a.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query heart rate data: %w", err)
	}
	defer rows.Close()

	totalReadings := 0
	for rows.Next() {
		var value float64
		var start, end time.Time
		if err := rows.Scan(&value, &start, &end); err != nil {
			continue
		}

		bpm := int(value)
		duration := end.Sub(start).Hours()

		// Assign to appropriate zone
		for i := range zones {
			if bpm >= zones[i].MinBPM && bpm < zones[i].MaxBPM {
				zones[i].Count++
				zones[i].TotalTime += duration
				break
			}
		}
		totalReadings++
	}

	// Calculate percentages
	for i := range zones {
		if totalReadings > 0 {
			zones[i].Percentage = float64(zones[i].Count) / float64(totalReadings) * 100
		}
	}

	return &HeartRateZones{
		StartDate:     startDate,
		EndDate:       endDate,
		Zones:         zones,
		TotalReadings: totalReadings,
	}, nil
}

// DetectAnomalies detects anomalous heart rate readings (>3 standard deviations)
func (a *HeartRateAnalyzer) DetectAnomalies(startDate, endDate time.Time) ([]Anomaly, error) {
	// First, calculate mean and standard deviation
	var mean, stdDev float64
	query := `
		SELECT AVG(value),
		       SQRT(AVG(value * value) - AVG(value) * AVG(value)) as stddev
		FROM health_records
		WHERE type = 'HeartRate'
		  AND start_date >= ?
		  AND end_date <= ?
	`

	err := a.db.QueryRow(query, startDate, endDate).Scan(&mean, &stdDev)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate statistics: %w", err)
	}

	// Find anomalies (>3 standard deviations from mean)
	threshold := 3.0
	lowerBound := mean - (threshold * stdDev)
	upperBound := mean + (threshold * stdDev)

	anomalyQuery := `
		SELECT id, start_date, value
		FROM health_records
		WHERE type = 'HeartRate'
		  AND start_date >= ?
		  AND end_date <= ?
		  AND (value < ? OR value > ?)
		ORDER BY start_date DESC
		LIMIT 100
	`

	rows, err := a.db.Query(anomalyQuery, startDate, endDate, lowerBound, upperBound)
	if err != nil {
		return nil, fmt.Errorf("failed to query anomalies: %w", err)
	}
	defer rows.Close()

	var anomalies []Anomaly
	for rows.Next() {
		var a Anomaly
		if err := rows.Scan(&a.ID, &a.Timestamp, &a.Value); err != nil {
			continue
		}

		// Determine reason and severity
		if a.Value < lowerBound {
			a.Reason = fmt.Sprintf("Unusually low (%.1f BPM, mean: %.1f)", a.Value, mean)
			if a.Value < 40 {
				a.Severity = "high"
			} else if a.Value < 50 {
				a.Severity = "medium"
			} else {
				a.Severity = "low"
			}
		} else {
			a.Reason = fmt.Sprintf("Unusually high (%.1f BPM, mean: %.1f)", a.Value, mean)
			if a.Value > 180 {
				a.Severity = "high"
			} else if a.Value > 160 {
				a.Severity = "medium"
			} else {
				a.Severity = "low"
			}
		}

		anomalies = append(anomalies, a)
	}

	return anomalies, nil
}

// GetRestingHeartRate calculates daily resting heart rate (using daily minimum as proxy)
func (a *HeartRateAnalyzer) GetRestingHeartRate(startDate, endDate time.Time) ([]RestingHR, error) {
	query := `
		SELECT DATE(start_date) as date,
		       MIN(value) as min_bpm,
		       AVG(value) as avg_bpm
		FROM health_records
		WHERE type = 'HeartRate'
		  AND start_date >= ?
		  AND end_date <= ?
		GROUP BY DATE(start_date)
		ORDER BY date
	`

	rows, err := a.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query resting heart rate: %w", err)
	}
	defer rows.Close()

	var results []RestingHR
	for rows.Next() {
		var r RestingHR
		if err := rows.Scan(&r.Date, &r.MinBPM, &r.AvgBPM); err != nil {
			continue
		}
		// Use minimum as proxy for resting heart rate
		r.RestingBPM = r.MinBPM
		results = append(results, r)
	}

	return results, nil
}

// GetHeartRateVariability calculates HRV metrics (using standard deviation as proxy)
func (a *HeartRateAnalyzer) GetHeartRateVariability(startDate, endDate time.Time) (*HRVMetrics, error) {
	// Calculate daily standard deviations
	query := `
		SELECT DATE(start_date) as date,
		       AVG(value) as avg_hr,
		       SQRT(AVG(value * value) - AVG(value) * AVG(value)) as stddev,
		       COUNT(*) as count
		FROM health_records
		WHERE type = 'HeartRate'
		  AND start_date >= ?
		  AND end_date <= ?
		GROUP BY DATE(start_date)
		HAVING count >= 10
	`

	rows, err := a.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query HRV data: %w", err)
	}
	defer rows.Close()

	var dailyStdDevs []float64
	var dailyCounts []int
	totalAvgHR := 0.0
	totalDays := 0

	for rows.Next() {
		var date string
		var avgHR, stdDev float64
		var count int
		if err := rows.Scan(&date, &avgHR, &stdDev, &count); err != nil {
			continue
		}
		dailyStdDevs = append(dailyStdDevs, stdDev)
		dailyCounts = append(dailyCounts, count)
		totalAvgHR += avgHR
		totalDays++
	}

	if totalDays == 0 {
		return &HRVMetrics{
			StartDate: startDate,
			EndDate:   endDate,
		}, nil
	}

	// Calculate average HRV (average of daily standard deviations)
	avgHRV := 0.0
	for _, stdDev := range dailyStdDevs {
		avgHRV += stdDev
	}
	avgHRV /= float64(len(dailyStdDevs))

	// Calculate standard deviation of HRV
	variance := 0.0
	for _, stdDev := range dailyStdDevs {
		variance += math.Pow(stdDev-avgHRV, 2)
	}
	stdDevOfHRV := math.Sqrt(variance / float64(len(dailyStdDevs)))

	// Calculate RMSSD (simplified version using daily stddevs)
	rmssd := avgHRV // Simplified proxy

	// Calculate consistency score (0-100)
	// Lower coefficient of variation = higher consistency
	cv := stdDevOfHRV / avgHRV
	consistency := math.Max(0, math.Min(100, 100*(1-cv)))

	return &HRVMetrics{
		StartDate:   startDate,
		EndDate:     endDate,
		AvgHRV:      avgHRV,
		StdDev:      stdDevOfHRV,
		RMSSD:       rmssd,
		Consistency: consistency,
	}, nil
}
