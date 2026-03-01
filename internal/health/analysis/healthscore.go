package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// HealthScore represents comprehensive health score
type HealthScore struct {
	Date                string  `json:"date"`
	OverallScore        float64 `json:"overallScore"`
	RestingHRScore      float64 `json:"restingHrScore"`
	VariabilityScore    float64 `json:"variabilityScore"`
	ConsistencyScore    float64 `json:"consistencyScore"`
	RestingHR           float64 `json:"restingHr"`
	AvgHR               float64 `json:"avgHr"`
	MeasurementCount    int     `json:"measurementCount"`
	Grade               string  `json:"grade"` // A+, A, B+, B, C+, C, D, F
}

// HealthScorePoint represents a point in health score trend
type HealthScorePoint struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
	Grade string  `json:"grade"`
}

// HealthScoreCalculator handles health score calculation
type HealthScoreCalculator struct {
	db *sql.DB
}

// NewHealthScoreCalculator creates a new health score calculator
func NewHealthScoreCalculator(db *sql.DB) *HealthScoreCalculator {
	return &HealthScoreCalculator{db: db}
}

// CalculateHealthScore calculates comprehensive health score (0-100)
// Weights: Resting HR 40% + Variability 30% + Consistency 30%
func (c *HealthScoreCalculator) CalculateHealthScore(date time.Time) (*HealthScore, error) {
	dateStr := date.Format("2006-01-02")

	// Get daily statistics
	query := `
		SELECT MIN(value) as min_hr,
		       AVG(value) as avg_hr,
		       MAX(value) as max_hr,
		       COUNT(*) as count,
		       SQRT(AVG(value * value) - AVG(value) * AVG(value)) as stddev
		FROM health_records
		WHERE type = 'HeartRate'
		  AND DATE(start_date) = ?
	`

	var minHR, avgHR, maxHR, stdDev float64
	var count int
	err := c.db.QueryRow(query, dateStr).Scan(&minHR, &avgHR, &maxHR, &count, &stdDev)
	if err != nil {
		return nil, fmt.Errorf("failed to query health data: %w", err)
	}

	if count == 0 {
		return &HealthScore{
			Date:         dateStr,
			OverallScore: 0,
			Grade:        "N/A",
		}, nil
	}

	// 1. Resting HR Score (40 points)
	// Optimal resting HR: 50-70 BPM
	// Score decreases as HR moves away from optimal range
	restingHRScore := calculateRestingHRScore(minHR)

	// 2. Variability Score (30 points)
	// Healthy HRV (using stddev as proxy): 10-30 BPM
	// Higher variability = better cardiovascular health
	variabilityScore := calculateVariabilityScore(stdDev)

	// 3. Consistency Score (30 points)
	// Based on measurement frequency
	// More measurements = more consistent tracking
	consistencyScore := calculateConsistencyScore(count)

	// Calculate overall score
	overallScore := restingHRScore + variabilityScore + consistencyScore

	// Determine grade
	grade := calculateGrade(overallScore)

	return &HealthScore{
		Date:             dateStr,
		OverallScore:     overallScore,
		RestingHRScore:   restingHRScore,
		VariabilityScore: variabilityScore,
		ConsistencyScore: consistencyScore,
		RestingHR:        minHR,
		AvgHR:            avgHR,
		MeasurementCount: count,
		Grade:            grade,
	}, nil
}

// GetHealthScoreTrend retrieves health score trend over time
func (c *HealthScoreCalculator) GetHealthScoreTrend(startDate, endDate time.Time) ([]HealthScorePoint, error) {
	var points []HealthScorePoint

	// Iterate through each day
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		score, err := c.CalculateHealthScore(currentDate)
		if err != nil {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		if score.MeasurementCount > 0 {
			points = append(points, HealthScorePoint{
				Date:  score.Date,
				Score: score.OverallScore,
				Grade: score.Grade,
			})
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return points, nil
}

// Helper functions

func calculateRestingHRScore(restingHR float64) float64 {
	// Optimal range: 50-70 BPM
	// Score: 40 points max
	if restingHR >= 50 && restingHR <= 70 {
		return 40.0
	}

	// Calculate distance from optimal range
	var distance float64
	if restingHR < 50 {
		distance = 50 - restingHR
	} else {
		distance = restingHR - 70
	}

	// Decrease score by 2 points per BPM away from optimal
	score := 40.0 - (distance * 2.0)
	if score < 0 {
		score = 0
	}

	return score
}

func calculateVariabilityScore(stdDev float64) float64 {
	// Optimal HRV (stddev): 10-30 BPM
	// Score: 30 points max
	if stdDev >= 10 && stdDev <= 30 {
		return 30.0
	}

	if stdDev < 10 {
		// Low variability (not ideal)
		score := stdDev / 10.0 * 30.0
		return score
	}

	// Very high variability (might indicate irregular rhythm)
	if stdDev > 50 {
		return 10.0
	}

	// Gradually decrease score for high variability
	score := 30.0 - ((stdDev - 30.0) / 20.0 * 20.0)
	if score < 10 {
		score = 10
	}

	return score
}

func calculateConsistencyScore(measurementCount int) float64 {
	// Optimal: 50+ measurements per day
	// Score: 30 points max
	if measurementCount >= 50 {
		return 30.0
	}

	// Linear scale: 0-50 measurements = 0-30 points
	score := float64(measurementCount) / 50.0 * 30.0
	return math.Min(score, 30.0)
}

func calculateGrade(score float64) string {
	switch {
	case score >= 95:
		return "A+"
	case score >= 90:
		return "A"
	case score >= 85:
		return "B+"
	case score >= 80:
		return "B"
	case score >= 75:
		return "C+"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
