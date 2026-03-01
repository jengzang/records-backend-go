package analysis

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// SleepAnalysis represents sleep quality analysis results
type SleepAnalysis struct {
	Summary           SleepSummary           `json:"summary"`
	DailySleep        []DailySleepStats      `json:"dailySleep"`
	SleepStages       SleepStageDistribution `json:"sleepStages"`
	SleepPattern      SleepPattern           `json:"sleepPattern"`
	QualityScore      SleepQualityScore      `json:"qualityScore"`
	HeartRateCorrelation HeartRateCorrelation `json:"heartRateCorrelation"`
	Recommendations   []string               `json:"recommendations"`
}

// SleepSummary contains overall sleep statistics
type SleepSummary struct {
	TotalSleepDays    int     `json:"totalSleepDays"`
	AvgSleepDuration  float64 `json:"avgSleepDuration"`  // hours
	AvgDeepSleep      float64 `json:"avgDeepSleep"`      // hours
	AvgLightSleep     float64 `json:"avgLightSleep"`     // hours
	AvgREMSleep       float64 `json:"avgREMSleep"`       // hours
	AvgBedTime        string  `json:"avgBedTime"`        // HH:MM
	AvgWakeTime       string  `json:"avgWakeTime"`       // HH:MM
	SleepEfficiency   float64 `json:"sleepEfficiency"`   // percentage
	BestSleepDate     string  `json:"bestSleepDate"`
	WorstSleepDate    string  `json:"worstSleepDate"`
}

// DailySleepStats contains daily sleep statistics
type DailySleepStats struct {
	Date             string  `json:"date"`
	TotalSleep       float64 `json:"totalSleep"`       // hours
	DeepSleep        float64 `json:"deepSleep"`        // hours
	LightSleep       float64 `json:"lightSleep"`       // hours
	REMSleep         float64 `json:"remSleep"`         // hours
	AwakeTime        float64 `json:"awakeTime"`        // hours
	BedTime          string  `json:"bedTime"`          // HH:MM
	WakeTime         string  `json:"wakeTime"`         // HH:MM
	SleepQuality     float64 `json:"sleepQuality"`     // 0-100
	HeartRateAvg     float64 `json:"heartRateAvg"`     // bpm
}

// SleepStageDistribution contains sleep stage distribution
type SleepStageDistribution struct {
	DeepSleepPercent  float64 `json:"deepSleepPercent"`
	LightSleepPercent float64 `json:"lightSleepPercent"`
	REMSleepPercent   float64 `json:"remSleepPercent"`
	AwakePercent      float64 `json:"awakePercent"`
	TotalMinutes      float64 `json:"totalMinutes"`
}

// SleepPattern contains sleep pattern analysis
type SleepPattern struct {
	BedTimeConsistency   float64 `json:"bedTimeConsistency"`   // 0-100
	WakeTimeConsistency  float64 `json:"wakeTimeConsistency"`  // 0-100
	WeekdayAvgSleep      float64 `json:"weekdayAvgSleep"`      // hours
	WeekendAvgSleep      float64 `json:"weekendAvgSleep"`      // hours
	SleepDebt            float64 `json:"sleepDebt"`            // hours
	OptimalBedTime       string  `json:"optimalBedTime"`       // HH:MM
	OptimalWakeTime      string  `json:"optimalWakeTime"`      // HH:MM
}

// SleepQualityScore contains sleep quality scoring
type SleepQualityScore struct {
	OverallScore      float64 `json:"overallScore"`      // 0-100
	DurationScore     float64 `json:"durationScore"`     // 0-100
	EfficiencyScore   float64 `json:"efficiencyScore"`   // 0-100
	ConsistencyScore  float64 `json:"consistencyScore"`  // 0-100
	DeepSleepScore    float64 `json:"deepSleepScore"`    // 0-100
	Grade             string  `json:"grade"`             // A+, A, B, C, D, F
}

// HeartRateCorrelation contains heart rate and sleep correlation
type HeartRateCorrelation struct {
	AvgSleepingHR     float64 `json:"avgSleepingHR"`     // bpm
	AvgWakingHR       float64 `json:"avgWakingHR"`       // bpm
	HRDrop            float64 `json:"hrDrop"`            // bpm
	HRDropPercent     float64 `json:"hrDropPercent"`     // percentage
	Correlation       float64 `json:"correlation"`       // -1 to 1
	Insight           string  `json:"insight"`
}

// GetSleepAnalysis analyzes sleep quality data
func GetSleepAnalysis(db *sql.DB) (*SleepAnalysis, error) {
	analysis := &SleepAnalysis{}
	analysis.DailySleep = []DailySleepStats{}
	analysis.Recommendations = []string{}

	// Get summary statistics
	summary, err := getSleepSummary(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get sleep summary: %w", err)
	}
	analysis.Summary = *summary

	// Get daily sleep statistics
	dailySleep, err := getDailySleepStats(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily sleep stats: %w", err)
	}
	analysis.DailySleep = dailySleep

	// Get sleep stage distribution
	sleepStages, err := getSleepStageDistribution(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get sleep stages: %w", err)
	}
	analysis.SleepStages = *sleepStages

	// Get sleep pattern analysis
	sleepPattern, err := getSleepPattern(db, dailySleep)
	if err != nil {
		return nil, fmt.Errorf("failed to get sleep pattern: %w", err)
	}
	analysis.SleepPattern = *sleepPattern

	// Calculate quality score
	qualityScore := calculateSleepQualityScore(summary, sleepStages, sleepPattern)
	analysis.QualityScore = *qualityScore

	// Get heart rate correlation
	hrCorrelation, err := getHeartRateCorrelation(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get heart rate correlation: %w", err)
	}
	analysis.HeartRateCorrelation = *hrCorrelation

	// Generate recommendations
	analysis.Recommendations = generateSleepRecommendations(summary, qualityScore, sleepPattern)

	return analysis, nil
}

func getSleepSummary(db *sql.DB) (*SleepSummary, error) {
	summary := &SleepSummary{}

	// Get sleep duration statistics
	err := db.QueryRow(`
		SELECT
			COUNT(DISTINCT DATE(start_date)) as sleep_days,
			COALESCE(AVG(CAST(value AS REAL) / 60.0), 0) as avg_sleep_hours
		FROM health_records
		WHERE type = 'HKCategoryTypeIdentifierSleepAnalysis'
		AND value IS NOT NULL
	`).Scan(&summary.TotalSleepDays, &summary.AvgSleepDuration)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get sleep stage averages (if available)
	// Note: This is simplified - real data would have separate records for each stage
	summary.AvgDeepSleep = summary.AvgSleepDuration * 0.20    // ~20% deep sleep
	summary.AvgLightSleep = summary.AvgSleepDuration * 0.50   // ~50% light sleep
	summary.AvgREMSleep = summary.AvgSleepDuration * 0.25     // ~25% REM sleep

	// Calculate sleep efficiency (simplified)
	summary.SleepEfficiency = 85.0 // Typical efficiency

	// Get best and worst sleep dates
	err = db.QueryRow(`
		SELECT DATE(start_date)
		FROM health_records
		WHERE type = 'HKCategoryTypeIdentifierSleepAnalysis'
		GROUP BY DATE(start_date)
		ORDER BY SUM(CAST(value AS REAL)) DESC
		LIMIT 1
	`).Scan(&summary.BestSleepDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = db.QueryRow(`
		SELECT DATE(start_date)
		FROM health_records
		WHERE type = 'HKCategoryTypeIdentifierSleepAnalysis'
		GROUP BY DATE(start_date)
		ORDER BY SUM(CAST(value AS REAL)) ASC
		LIMIT 1
	`).Scan(&summary.WorstSleepDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Estimate average bed time and wake time (simplified)
	summary.AvgBedTime = "23:30"
	summary.AvgWakeTime = "07:30"

	return summary, nil
}

func getDailySleepStats(db *sql.DB) ([]DailySleepStats, error) {
	rows, err := db.Query(`
		SELECT
			DATE(start_date) as date,
			COALESCE(SUM(CAST(value AS REAL)) / 60.0, 0) as total_sleep_hours,
			MIN(TIME(start_date)) as bed_time,
			MAX(TIME(end_date)) as wake_time
		FROM health_records
		WHERE type = 'HKCategoryTypeIdentifierSleepAnalysis'
		GROUP BY DATE(start_date)
		ORDER BY date DESC
		LIMIT 90
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []DailySleepStats{}
	for rows.Next() {
		var s DailySleepStats
		err := rows.Scan(&s.Date, &s.TotalSleep, &s.BedTime, &s.WakeTime)
		if err != nil {
			continue
		}

		// Estimate sleep stages (simplified distribution)
		s.DeepSleep = s.TotalSleep * 0.20
		s.LightSleep = s.TotalSleep * 0.50
		s.REMSleep = s.TotalSleep * 0.25
		s.AwakeTime = s.TotalSleep * 0.05

		// Calculate sleep quality score
		s.SleepQuality = calculateDailySleepQuality(s.TotalSleep, s.DeepSleep)

		stats = append(stats, s)
	}

	return stats, nil
}

func getSleepStageDistribution(db *sql.DB) (*SleepStageDistribution, error) {
	dist := &SleepStageDistribution{}

	// Get total sleep minutes
	err := db.QueryRow(`
		SELECT COALESCE(SUM(CAST(value AS REAL)), 0)
		FROM health_records
		WHERE type = 'HKCategoryTypeIdentifierSleepAnalysis'
	`).Scan(&dist.TotalMinutes)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Estimate stage distribution (simplified)
	dist.DeepSleepPercent = 20.0
	dist.LightSleepPercent = 50.0
	dist.REMSleepPercent = 25.0
	dist.AwakePercent = 5.0

	return dist, nil
}

func getSleepPattern(db *sql.DB, dailySleep []DailySleepStats) (*SleepPattern, error) {
	pattern := &SleepPattern{}

	if len(dailySleep) == 0 {
		return pattern, nil
	}

	// Calculate bed time and wake time consistency
	bedTimes := []time.Time{}
	wakeTimes := []time.Time{}
	weekdaySleep := []float64{}
	weekendSleep := []float64{}

	for _, day := range dailySleep {
		// Parse date to determine weekday/weekend
		date, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			continue
		}

		weekday := date.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			weekendSleep = append(weekendSleep, day.TotalSleep)
		} else {
			weekdaySleep = append(weekdaySleep, day.TotalSleep)
		}

		// Parse times for consistency calculation
		if bedTime, err := time.Parse("15:04:05", day.BedTime); err == nil {
			bedTimes = append(bedTimes, bedTime)
		}
		if wakeTime, err := time.Parse("15:04:05", day.WakeTime); err == nil {
			wakeTimes = append(wakeTimes, wakeTime)
		}
	}

	// Calculate consistency scores (based on standard deviation)
	pattern.BedTimeConsistency = calculateTimeConsistency(bedTimes)
	pattern.WakeTimeConsistency = calculateTimeConsistency(wakeTimes)

	// Calculate weekday vs weekend averages
	if len(weekdaySleep) > 0 {
		pattern.WeekdayAvgSleep = average(weekdaySleep)
	}
	if len(weekendSleep) > 0 {
		pattern.WeekendAvgSleep = average(weekendSleep)
	}

	// Calculate sleep debt (difference from recommended 8 hours)
	pattern.SleepDebt = math.Max(0, 8.0-pattern.WeekdayAvgSleep)

	// Set optimal times (simplified)
	pattern.OptimalBedTime = "23:00"
	pattern.OptimalWakeTime = "07:00"

	return pattern, nil
}

func calculateSleepQualityScore(summary *SleepSummary, stages *SleepStageDistribution, pattern *SleepPattern) *SleepQualityScore {
	score := &SleepQualityScore{}

	// Duration score (optimal: 7-9 hours)
	if summary.AvgSleepDuration >= 7.0 && summary.AvgSleepDuration <= 9.0 {
		score.DurationScore = 100.0
	} else if summary.AvgSleepDuration >= 6.0 && summary.AvgSleepDuration < 7.0 {
		score.DurationScore = 80.0
	} else if summary.AvgSleepDuration >= 5.0 && summary.AvgSleepDuration < 6.0 {
		score.DurationScore = 60.0
	} else {
		score.DurationScore = 40.0
	}

	// Efficiency score
	score.EfficiencyScore = summary.SleepEfficiency

	// Consistency score (average of bed time and wake time consistency)
	score.ConsistencyScore = (pattern.BedTimeConsistency + pattern.WakeTimeConsistency) / 2.0

	// Deep sleep score (optimal: 15-25%)
	if stages.DeepSleepPercent >= 15.0 && stages.DeepSleepPercent <= 25.0 {
		score.DeepSleepScore = 100.0
	} else if stages.DeepSleepPercent >= 10.0 && stages.DeepSleepPercent < 15.0 {
		score.DeepSleepScore = 80.0
	} else {
		score.DeepSleepScore = 60.0
	}

	// Overall score (weighted average)
	score.OverallScore = (score.DurationScore*0.3 +
		score.EfficiencyScore*0.25 +
		score.ConsistencyScore*0.25 +
		score.DeepSleepScore*0.20)

	// Assign grade
	if score.OverallScore >= 95 {
		score.Grade = "A+"
	} else if score.OverallScore >= 90 {
		score.Grade = "A"
	} else if score.OverallScore >= 80 {
		score.Grade = "B"
	} else if score.OverallScore >= 70 {
		score.Grade = "C"
	} else if score.OverallScore >= 60 {
		score.Grade = "D"
	} else {
		score.Grade = "F"
	}

	return score
}

func getHeartRateCorrelation(db *sql.DB) (*HeartRateCorrelation, error) {
	corr := &HeartRateCorrelation{}

	// Get average heart rate during sleep hours (simplified: 22:00-06:00)
	err := db.QueryRow(`
		SELECT COALESCE(AVG(CAST(value AS REAL)), 0)
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierHeartRate'
		AND CAST(strftime('%H', start_date) AS INTEGER) >= 22
		OR CAST(strftime('%H', start_date) AS INTEGER) <= 6
	`).Scan(&corr.AvgSleepingHR)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get average heart rate during waking hours
	err = db.QueryRow(`
		SELECT COALESCE(AVG(CAST(value AS REAL)), 0)
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierHeartRate'
		AND CAST(strftime('%H', start_date) AS INTEGER) > 6
		AND CAST(strftime('%H', start_date) AS INTEGER) < 22
	`).Scan(&corr.AvgWakingHR)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Calculate heart rate drop
	corr.HRDrop = corr.AvgWakingHR - corr.AvgSleepingHR
	if corr.AvgWakingHR > 0 {
		corr.HRDropPercent = (corr.HRDrop / corr.AvgWakingHR) * 100.0
	}

	// Simplified correlation (would need more complex calculation with real data)
	corr.Correlation = -0.65 // Negative correlation: lower HR = better sleep

	// Generate insight
	if corr.HRDropPercent >= 15.0 {
		corr.Insight = "睡眠期间心率下降良好，表明睡眠质量较高"
	} else if corr.HRDropPercent >= 10.0 {
		corr.Insight = "睡眠期间心率下降适中，睡眠质量正常"
	} else {
		corr.Insight = "睡眠期间心率下降较少，可能影响睡眠恢复效果"
	}

	return corr, nil
}

func generateSleepRecommendations(summary *SleepSummary, quality *SleepQualityScore, pattern *SleepPattern) []string {
	recommendations := []string{}

	// Duration recommendations
	if summary.AvgSleepDuration < 7.0 {
		recommendations = append(recommendations, "建议增加睡眠时长，成年人推荐每晚7-9小时睡眠")
	} else if summary.AvgSleepDuration > 9.0 {
		recommendations = append(recommendations, "睡眠时间过长可能影响睡眠质量，建议保持7-9小时")
	}

	// Consistency recommendations
	if quality.ConsistencyScore < 70 {
		recommendations = append(recommendations, "建议保持规律的作息时间，每天在相同时间入睡和起床")
	}

	// Deep sleep recommendations
	if quality.DeepSleepScore < 80 {
		recommendations = append(recommendations, "深度睡眠不足，建议睡前避免使用电子设备，保持卧室安静黑暗")
	}

	// Sleep debt recommendations
	if pattern.SleepDebt > 1.0 {
		recommendations = append(recommendations, fmt.Sprintf("存在睡眠债务(%.1f小时)，建议周末适当补充睡眠", pattern.SleepDebt))
	}

	// Weekday vs weekend recommendations
	if pattern.WeekendAvgSleep-pattern.WeekdayAvgSleep > 2.0 {
		recommendations = append(recommendations, "周末补觉过多，建议平时增加睡眠时间，保持一致的作息")
	}

	// General recommendations
	if quality.OverallScore < 70 {
		recommendations = append(recommendations, "睡眠质量有待提高，建议：规律作息、睡前放松、适度运动、避免咖啡因")
	}

	return recommendations
}

// Helper functions

func calculateDailySleepQuality(totalSleep, deepSleep float64) float64 {
	score := 0.0

	// Duration component (0-50 points)
	if totalSleep >= 7.0 && totalSleep <= 9.0 {
		score += 50.0
	} else if totalSleep >= 6.0 && totalSleep < 7.0 {
		score += 40.0
	} else if totalSleep >= 5.0 && totalSleep < 6.0 {
		score += 30.0
	} else {
		score += 20.0
	}

	// Deep sleep component (0-50 points)
	deepSleepPercent := (deepSleep / totalSleep) * 100.0
	if deepSleepPercent >= 15.0 && deepSleepPercent <= 25.0 {
		score += 50.0
	} else if deepSleepPercent >= 10.0 && deepSleepPercent < 15.0 {
		score += 40.0
	} else {
		score += 30.0
	}

	return score
}

func calculateTimeConsistency(times []time.Time) float64 {
	if len(times) < 2 {
		return 100.0
	}

	// Convert times to minutes since midnight
	minutes := make([]float64, len(times))
	for i, t := range times {
		minutes[i] = float64(t.Hour()*60 + t.Minute())
	}

	// Calculate standard deviation
	stdDev := standardDeviation(minutes)

	// Convert to consistency score (0-100)
	// Lower std dev = higher consistency
	// 0 min std dev = 100 score, 60 min std dev = 0 score
	score := math.Max(0, 100.0-(stdDev/60.0)*100.0)

	return score
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func standardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	avg := average(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}
