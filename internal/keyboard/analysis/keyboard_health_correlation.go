package analysis

import (
	"database/sql"
	"fmt"
	"math"
)

// KeyboardHealthCorrelation represents keyboard-health correlation analysis
type KeyboardHealthCorrelation struct {
	CorrelationScore    float64                    `json:"correlationScore"`    // Overall correlation score (0-100)
	TypingActivityCorr  *TypingActivityCorrelation `json:"typingActivityCorr"`  // Typing vs steps correlation
	TypingIntensityCorr *TypingIntensityCorr       `json:"typingIntensityCorr"` // Typing intensity vs heart rate
	SedentaryAnalysis   *SedentaryTypingAnalysis   `json:"sedentaryAnalysis"`   // Low typing days analysis
	WorkdayHealth       *WorkdayHealthPattern      `json:"workdayHealth"`       // Workday typing vs health
	Recommendations     []string                   `json:"recommendations"`     // Health recommendations
}

// TypingActivityCorrelation represents typing activity vs steps correlation
type TypingActivityCorrelation struct {
	CorrelationCoefficient float64              `json:"correlationCoefficient"` // Pearson correlation
	CorrelationType        string               `json:"correlationType"`        // "positive", "negative", "none"
	AvgKeystrokesPerDay    float64              `json:"avgKeystrokesPerDay"`
	AvgStepsPerDay         float64              `json:"avgStepsPerDay"`
	DataPoints             []ActivityDataPoint  `json:"dataPoints"`
}

// ActivityDataPoint represents a single day's data
type ActivityDataPoint struct {
	Date       string  `json:"date"`
	Keystrokes int64   `json:"keystrokes"`
	Steps      int64   `json:"steps"`
}

// TypingIntensityCorr represents typing intensity vs heart rate correlation
type TypingIntensityCorr struct {
	CorrelationCoefficient float64                  `json:"correlationCoefficient"`
	AvgKeystrokesPerHour   float64                  `json:"avgKeystrokesPerHour"`
	AvgHeartRate           float64                  `json:"avgHeartRate"`
	IntensityLevels        []IntensityLevel         `json:"intensityLevels"`
}

// IntensityLevel represents typing intensity level
type IntensityLevel struct {
	Level          string  `json:"level"`          // "low", "medium", "high"
	KeystrokeRange string  `json:"keystrokeRange"` // e.g., "0-1000"
	AvgHeartRate   float64 `json:"avgHeartRate"`
	DayCount       int     `json:"dayCount"`
}

// SedentaryTypingAnalysis represents sedentary typing days analysis
type SedentaryTypingAnalysis struct {
	SedentaryDays          int                    `json:"sedentaryDays"`          // Days with low typing (<5000 keystrokes)
	TotalDays              int                    `json:"totalDays"`
	SedentaryRate          float64                `json:"sedentaryRate"`          // Percentage
	AvgStepsOnSedentary    float64                `json:"avgStepsOnSedentary"`
	AvgStepsOnActive       float64                `json:"avgStepsOnActive"`
	HealthImpact           string                 `json:"healthImpact"`           // "high", "medium", "low"
	SedentaryDayDetails    []SedentaryDayDetail   `json:"sedentaryDayDetails"`
}

// SedentaryDayDetail represents a sedentary day detail
type SedentaryDayDetail struct {
	Date       string `json:"date"`
	Keystrokes int64  `json:"keystrokes"`
	Steps      int64  `json:"steps"`
	HeartRate  int    `json:"heartRate"`
}

// WorkdayHealthPattern represents workday typing vs health pattern
type WorkdayHealthPattern struct {
	WorkdayAvgKeystrokes float64 `json:"workdayAvgKeystrokes"`
	WeekendAvgKeystrokes float64 `json:"weekendAvgKeystrokes"`
	WorkdayAvgSteps      float64 `json:"workdayAvgSteps"`
	WeekendAvgSteps      float64 `json:"weekendAvgSteps"`
	WorkdayAvgHeartRate  float64 `json:"workdayAvgHeartRate"`
	WeekendAvgHeartRate  float64 `json:"weekendAvgHeartRate"`
	HealthBalance        string  `json:"healthBalance"` // "good", "moderate", "poor"
}

// GetKeyboardHealthCorrelation analyzes keyboard-health correlation
func GetKeyboardHealthCorrelation(keyboardDB, healthDB *sql.DB) (*KeyboardHealthCorrelation, error) {
	result := &KeyboardHealthCorrelation{
		TypingActivityCorr:  &TypingActivityCorrelation{DataPoints: []ActivityDataPoint{}},
		TypingIntensityCorr: &TypingIntensityCorr{IntensityLevels: []IntensityLevel{}},
		SedentaryAnalysis:   &SedentaryTypingAnalysis{SedentaryDayDetails: []SedentaryDayDetail{}},
		WorkdayHealth:       &WorkdayHealthPattern{},
		Recommendations:     []string{},
	}

	// 1. Analyze typing activity vs steps correlation
	if err := analyzeTypingActivityCorrelation(keyboardDB, healthDB, result); err != nil {
		return nil, fmt.Errorf("failed to analyze typing activity correlation: %w", err)
	}

	// 2. Analyze typing intensity vs heart rate
	if err := analyzeTypingIntensityCorrelation(keyboardDB, healthDB, result); err != nil {
		return nil, fmt.Errorf("failed to analyze typing intensity correlation: %w", err)
	}

	// 3. Analyze sedentary typing days
	if err := analyzeSedentaryTypingDays(keyboardDB, healthDB, result); err != nil {
		return nil, fmt.Errorf("failed to analyze sedentary typing days: %w", err)
	}

	// 4. Analyze workday vs weekend health pattern
	if err := analyzeWorkdayHealthPattern(keyboardDB, healthDB, result); err != nil {
		return nil, fmt.Errorf("failed to analyze workday health pattern: %w", err)
	}

	// 5. Calculate overall correlation score
	calculateOverallScore(result)

	// 6. Generate recommendations
	generateHealthRecommendations(result)

	return result, nil
}

func analyzeTypingActivityCorrelation(keyboardDB, healthDB *sql.DB, result *KeyboardHealthCorrelation) error {
	// Query daily keystrokes and steps
	query := `
		SELECT
			k.date,
			k.total_keystrokes,
			COALESCE(h.total_value, 0) as steps
		FROM keyboard_daily_stats k
		LEFT JOIN health_statistics h ON k.date = h.stat_date
			AND h.stat_type = 'daily'
			AND h.metric_type = 'StepCount'
		WHERE k.total_keystrokes > 0
		ORDER BY k.date
	`

	rows, err := keyboardDB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var dataPoints []ActivityDataPoint
	var sumKeystrokes, sumSteps float64
	var count int

	for rows.Next() {
		var dp ActivityDataPoint
		if err := rows.Scan(&dp.Date, &dp.Keystrokes, &dp.Steps); err != nil {
			continue
		}

		dataPoints = append(dataPoints, dp)
		sumKeystrokes += float64(dp.Keystrokes)
		sumSteps += float64(dp.Steps)
		count++
	}

	if count == 0 {
		return nil
	}

	// Calculate averages
	result.TypingActivityCorr.AvgKeystrokesPerDay = sumKeystrokes / float64(count)
	result.TypingActivityCorr.AvgStepsPerDay = sumSteps / float64(count)
	result.TypingActivityCorr.DataPoints = dataPoints

	// Calculate Pearson correlation coefficient
	if count > 1 {
		result.TypingActivityCorr.CorrelationCoefficient = calculatePearsonCorrelation(dataPoints)

		// Determine correlation type
		corr := result.TypingActivityCorr.CorrelationCoefficient
		if corr > 0.3 {
			result.TypingActivityCorr.CorrelationType = "正相关"
		} else if corr < -0.3 {
			result.TypingActivityCorr.CorrelationType = "负相关"
		} else {
			result.TypingActivityCorr.CorrelationType = "无明显相关"
		}
	}

	return nil
}

func analyzeTypingIntensityCorrelation(keyboardDB, healthDB *sql.DB, result *KeyboardHealthCorrelation) error {
	// Query typing intensity levels and corresponding heart rates
	query := `
		SELECT
			k.date,
			k.total_keystrokes,
			COALESCE(h.avg_value, 0) as avg_heart_rate
		FROM keyboard_daily_stats k
		LEFT JOIN health_statistics h ON k.date = h.stat_date
			AND h.stat_type = 'daily'
			AND h.metric_type = 'HeartRate'
		WHERE k.total_keystrokes > 0 AND h.avg_value > 0
		ORDER BY k.date
	`

	rows, err := keyboardDB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	type dataPoint struct {
		keystrokes int64
		heartRate  float64
	}

	var points []dataPoint
	var sumKeystrokes, sumHeartRate float64

	for rows.Next() {
		var date string
		var dp dataPoint
		if err := rows.Scan(&date, &dp.keystrokes, &dp.heartRate); err != nil {
			continue
		}

		points = append(points, dp)
		sumKeystrokes += float64(dp.keystrokes)
		sumHeartRate += dp.heartRate
	}

	if len(points) == 0 {
		return nil
	}

	result.TypingIntensityCorr.AvgKeystrokesPerHour = sumKeystrokes / float64(len(points)) / 24
	result.TypingIntensityCorr.AvgHeartRate = sumHeartRate / float64(len(points))

	// Categorize into intensity levels
	lowLevel := IntensityLevel{Level: "低强度", KeystrokeRange: "0-5000", DayCount: 0}
	medLevel := IntensityLevel{Level: "中强度", KeystrokeRange: "5000-15000", DayCount: 0}
	highLevel := IntensityLevel{Level: "高强度", KeystrokeRange: ">15000", DayCount: 0}

	var lowHR, medHR, highHR float64

	for _, p := range points {
		if p.keystrokes < 5000 {
			lowLevel.DayCount++
			lowHR += p.heartRate
		} else if p.keystrokes < 15000 {
			medLevel.DayCount++
			medHR += p.heartRate
		} else {
			highLevel.DayCount++
			highHR += p.heartRate
		}
	}

	if lowLevel.DayCount > 0 {
		lowLevel.AvgHeartRate = lowHR / float64(lowLevel.DayCount)
		result.TypingIntensityCorr.IntensityLevels = append(result.TypingIntensityCorr.IntensityLevels, lowLevel)
	}
	if medLevel.DayCount > 0 {
		medLevel.AvgHeartRate = medHR / float64(medLevel.DayCount)
		result.TypingIntensityCorr.IntensityLevels = append(result.TypingIntensityCorr.IntensityLevels, medLevel)
	}
	if highLevel.DayCount > 0 {
		highLevel.AvgHeartRate = highHR / float64(highLevel.DayCount)
		result.TypingIntensityCorr.IntensityLevels = append(result.TypingIntensityCorr.IntensityLevels, highLevel)
	}

	return nil
}

func analyzeSedentaryTypingDays(keyboardDB, healthDB *sql.DB, result *KeyboardHealthCorrelation) error {
	// Query days with low typing activity (<5000 keystrokes)
	query := `
		SELECT
			k.date,
			k.total_keystrokes,
			COALESCE(h_steps.total_value, 0) as steps,
			COALESCE(h_hr.avg_value, 0) as heart_rate
		FROM keyboard_daily_stats k
		LEFT JOIN health_statistics h_steps ON k.date = h_steps.stat_date
			AND h_steps.stat_type = 'daily'
			AND h_steps.metric_type = 'StepCount'
		LEFT JOIN health_statistics h_hr ON k.date = h_hr.stat_date
			AND h_hr.stat_type = 'daily'
			AND h_hr.metric_type = 'HeartRate'
		WHERE k.total_keystrokes > 0
		ORDER BY k.date
	`

	rows, err := keyboardDB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var sedentaryDays, totalDays int
	var sedentarySteps, activeSteps float64
	var sedentaryCount, activeCount int

	for rows.Next() {
		var detail SedentaryDayDetail
		if err := rows.Scan(&detail.Date, &detail.Keystrokes, &detail.Steps, &detail.HeartRate); err != nil {
			continue
		}

		totalDays++

		if detail.Keystrokes < 5000 {
			sedentaryDays++
			sedentarySteps += float64(detail.Steps)
			sedentaryCount++
			result.SedentaryAnalysis.SedentaryDayDetails = append(result.SedentaryAnalysis.SedentaryDayDetails, detail)
		} else {
			activeSteps += float64(detail.Steps)
			activeCount++
		}
	}

	result.SedentaryAnalysis.SedentaryDays = sedentaryDays
	result.SedentaryAnalysis.TotalDays = totalDays

	if totalDays > 0 {
		result.SedentaryAnalysis.SedentaryRate = float64(sedentaryDays) / float64(totalDays) * 100
	}

	if sedentaryCount > 0 {
		result.SedentaryAnalysis.AvgStepsOnSedentary = sedentarySteps / float64(sedentaryCount)
	}
	if activeCount > 0 {
		result.SedentaryAnalysis.AvgStepsOnActive = activeSteps / float64(activeCount)
	}

	// Determine health impact
	if result.SedentaryAnalysis.SedentaryRate > 50 {
		result.SedentaryAnalysis.HealthImpact = "高风险"
	} else if result.SedentaryAnalysis.SedentaryRate > 30 {
		result.SedentaryAnalysis.HealthImpact = "中风险"
	} else {
		result.SedentaryAnalysis.HealthImpact = "低风险"
	}

	return nil
}

func analyzeWorkdayHealthPattern(keyboardDB, healthDB *sql.DB, result *KeyboardHealthCorrelation) error {
	// Query workday vs weekend patterns
	query := `
		SELECT
			k.date,
			k.total_keystrokes,
			COALESCE(h_steps.total_value, 0) as steps,
			COALESCE(h_hr.avg_value, 0) as heart_rate,
			CASE
				WHEN CAST(strftime('%w', substr(k.date, 1, 4) || '-' || substr(k.date, 5, 2) || '-' || substr(k.date, 7, 2)) AS INTEGER) IN (0, 6)
				THEN 'weekend'
				ELSE 'workday'
			END as day_type
		FROM keyboard_daily_stats k
		LEFT JOIN health_statistics h_steps ON k.date = h_steps.stat_date
			AND h_steps.stat_type = 'daily'
			AND h_steps.metric_type = 'StepCount'
		LEFT JOIN health_statistics h_hr ON k.date = h_hr.stat_date
			AND h_hr.stat_type = 'daily'
			AND h_hr.metric_type = 'HeartRate'
		WHERE k.total_keystrokes > 0
	`

	rows, err := keyboardDB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var workdayKeystrokes, weekendKeystrokes float64
	var workdaySteps, weekendSteps float64
	var workdayHR, weekendHR float64
	var workdayCount, weekendCount int

	for rows.Next() {
		var date, dayType string
		var keystrokes, steps int64
		var heartRate float64

		if err := rows.Scan(&date, &keystrokes, &steps, &heartRate, &dayType); err != nil {
			continue
		}

		if dayType == "workday" {
			workdayKeystrokes += float64(keystrokes)
			workdaySteps += float64(steps)
			workdayHR += heartRate
			workdayCount++
		} else {
			weekendKeystrokes += float64(keystrokes)
			weekendSteps += float64(steps)
			weekendHR += heartRate
			weekendCount++
		}
	}

	if workdayCount > 0 {
		result.WorkdayHealth.WorkdayAvgKeystrokes = workdayKeystrokes / float64(workdayCount)
		result.WorkdayHealth.WorkdayAvgSteps = workdaySteps / float64(workdayCount)
		result.WorkdayHealth.WorkdayAvgHeartRate = workdayHR / float64(workdayCount)
	}

	if weekendCount > 0 {
		result.WorkdayHealth.WeekendAvgKeystrokes = weekendKeystrokes / float64(weekendCount)
		result.WorkdayHealth.WeekendAvgSteps = weekendSteps / float64(weekendCount)
		result.WorkdayHealth.WeekendAvgHeartRate = weekendHR / float64(weekendCount)
	}

	// Determine health balance
	stepsDiff := result.WorkdayHealth.WeekendAvgSteps - result.WorkdayHealth.WorkdayAvgSteps
	if stepsDiff > 2000 {
		result.WorkdayHealth.HealthBalance = "良好"
	} else if stepsDiff > 0 {
		result.WorkdayHealth.HealthBalance = "一般"
	} else {
		result.WorkdayHealth.HealthBalance = "需改善"
	}

	return nil
}

func calculatePearsonCorrelation(dataPoints []ActivityDataPoint) float64 {
	n := len(dataPoints)
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for _, dp := range dataPoints {
		x := float64(dp.Keystrokes)
		y := float64(dp.Steps)

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	numerator := float64(n)*sumXY - sumX*sumY
	denominator := math.Sqrt((float64(n)*sumX2 - sumX*sumX) * (float64(n)*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

func calculateOverallScore(result *KeyboardHealthCorrelation) {
	score := 100.0

	// Deduct for high sedentary rate
	score -= result.SedentaryAnalysis.SedentaryRate * 0.3

	// Deduct for poor workday health balance
	if result.WorkdayHealth.HealthBalance == "需改善" {
		score -= 20
	} else if result.WorkdayHealth.HealthBalance == "一般" {
		score -= 10
	}

	// Bonus for positive typing-activity correlation
	if result.TypingActivityCorr.CorrelationType == "正相关" {
		score += 10
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	result.CorrelationScore = score
}

func generateHealthRecommendations(result *KeyboardHealthCorrelation) {
	recommendations := []string{}

	// Sedentary recommendations
	if result.SedentaryAnalysis.SedentaryRate > 40 {
		recommendations = append(recommendations, fmt.Sprintf("久坐率为%.1f%%，建议每小时起身活动5-10分钟", result.SedentaryAnalysis.SedentaryRate))
	}

	// Activity correlation recommendations
	if result.TypingActivityCorr.CorrelationType == "负相关" {
		recommendations = append(recommendations, "打字活动与步数呈负相关，建议在工作间隙增加走动")
	}

	// Workday health recommendations
	if result.WorkdayHealth.HealthBalance == "需改善" {
		recommendations = append(recommendations, fmt.Sprintf("工作日平均步数(%.0f)低于周末，建议增加工作日活动量", result.WorkdayHealth.WorkdayAvgSteps))
	}

	// Steps recommendations
	if result.TypingActivityCorr.AvgStepsPerDay < 5000 {
		recommendations = append(recommendations, "日均步数不足5000步，建议设定每日步数目标")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "您的打字与健康习惯良好，继续保持!")
	}

	result.Recommendations = recommendations
}
