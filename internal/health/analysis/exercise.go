package analysis

import (
	"database/sql"
	"fmt"
	"math"
)

// ExerciseAnalysis represents exercise and activity data analysis
type ExerciseAnalysis struct {
	Summary          ExerciseSummary          `json:"summary"`
	DailyStats       []DailyExerciseStats     `json:"dailyStats"`
	WorkoutTypes     []WorkoutTypeStats       `json:"workoutTypes"`
	CalorieTrend     []CalorieDataPoint       `json:"calorieTrend"`
	IntensityAnalysis IntensityAnalysis       `json:"intensityAnalysis"`
	Achievements     []Achievement            `json:"achievements"`
	Recommendations  []string                 `json:"recommendations"`
}

// ExerciseSummary contains overall exercise statistics
type ExerciseSummary struct {
	TotalSteps          int64   `json:"totalSteps"`
	TotalDistance       float64 `json:"totalDistance"`       // km
	TotalCalories       float64 `json:"totalCalories"`       // kcal
	TotalWorkouts       int     `json:"totalWorkouts"`
	AvgDailySteps       int64   `json:"avgDailySteps"`
	AvgDailyDistance    float64 `json:"avgDailyDistance"`    // km
	AvgDailyCalories    float64 `json:"avgDailyCalories"`    // kcal
	ActiveDays          int     `json:"activeDays"`
	MostActiveDay       string  `json:"mostActiveDay"`
	LongestWorkout      float64 `json:"longestWorkout"`      // minutes
	TotalExerciseTime   float64 `json:"totalExerciseTime"`   // hours
}

// DailyExerciseStats contains daily exercise statistics
type DailyExerciseStats struct {
	Date              string  `json:"date"`
	Steps             int64   `json:"steps"`
	Distance          float64 `json:"distance"`          // km
	Calories          float64 `json:"calories"`          // kcal
	FlightsClimbed    int     `json:"flightsClimbed"`
	ExerciseMinutes   int     `json:"exerciseMinutes"`
	WorkoutCount      int     `json:"workoutCount"`
}

// WorkoutTypeStats contains statistics for each workout type
type WorkoutTypeStats struct {
	WorkoutType       string  `json:"workoutType"`
	Count             int     `json:"count"`
	TotalDuration     float64 `json:"totalDuration"`     // minutes
	TotalDistance     float64 `json:"totalDistance"`     // km
	TotalCalories     float64 `json:"totalCalories"`     // kcal
	AvgDuration       float64 `json:"avgDuration"`       // minutes
	AvgDistance       float64 `json:"avgDistance"`       // km
	AvgCalories       float64 `json:"avgCalories"`       // kcal
	Percentage        float64 `json:"percentage"`
}

// CalorieDataPoint represents calorie data for a specific date
type CalorieDataPoint struct {
	Date              string  `json:"date"`
	ActiveCalories    float64 `json:"activeCalories"`    // kcal
	RestingCalories   float64 `json:"restingCalories"`   // kcal
	TotalCalories     float64 `json:"totalCalories"`     // kcal
}

// IntensityAnalysis contains exercise intensity analysis
type IntensityAnalysis struct {
	LowIntensity      IntensityStats `json:"lowIntensity"`
	ModerateIntensity IntensityStats `json:"moderateIntensity"`
	HighIntensity     IntensityStats `json:"highIntensity"`
	AvgMETs           float64        `json:"avgMETs"`
	IntensityScore    float64        `json:"intensityScore"`    // 0-100
}

// IntensityStats contains statistics for an intensity level
type IntensityStats struct {
	Minutes           int     `json:"minutes"`
	Percentage        float64 `json:"percentage"`
	Calories          float64 `json:"calories"`
}

// Achievement represents an exercise achievement
type Achievement struct {
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	Value             float64 `json:"value"`
	Unit              string  `json:"unit"`
	Date              string  `json:"date"`
	Type              string  `json:"type"`              // steps, distance, calories, workout
}

// GetExerciseAnalysis analyzes exercise and activity data
func GetExerciseAnalysis(db *sql.DB) (*ExerciseAnalysis, error) {
	analysis := &ExerciseAnalysis{}
	analysis.DailyStats = []DailyExerciseStats{}
	analysis.WorkoutTypes = []WorkoutTypeStats{}
	analysis.CalorieTrend = []CalorieDataPoint{}
	analysis.Achievements = []Achievement{}
	analysis.Recommendations = []string{}

	// Get summary statistics
	summary, err := getExerciseSummary(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise summary: %w", err)
	}
	analysis.Summary = *summary

	// Get daily statistics
	dailyStats, err := getDailyExerciseStats(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	analysis.DailyStats = dailyStats

	// Get workout type statistics
	workoutTypes, err := getWorkoutTypeStats(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout types: %w", err)
	}
	analysis.WorkoutTypes = workoutTypes

	// Get calorie trend
	calorieTrend, err := getCalorieTrend(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get calorie trend: %w", err)
	}
	analysis.CalorieTrend = calorieTrend

	// Get intensity analysis
	intensityAnalysis, err := getIntensityAnalysis(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get intensity analysis: %w", err)
	}
	analysis.IntensityAnalysis = *intensityAnalysis

	// Detect achievements
	analysis.Achievements = detectAchievements(dailyStats, workoutTypes)

	// Generate recommendations
	analysis.Recommendations = generateExerciseRecommendations(summary, intensityAnalysis)

	return analysis, nil
}

func getExerciseSummary(db *sql.DB) (*ExerciseSummary, error) {
	summary := &ExerciseSummary{}

	// Get step count statistics
	err := db.QueryRow(`
		SELECT
			COALESCE(SUM(CAST(value AS INTEGER)), 0) as total_steps,
			COALESCE(AVG(CAST(value AS INTEGER)), 0) as avg_steps,
			COUNT(DISTINCT DATE(start_date)) as active_days
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		AND value IS NOT NULL
	`).Scan(&summary.TotalSteps, &summary.AvgDailySteps, &summary.ActiveDays)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get distance statistics (convert meters to km)
	var totalDistanceMeters float64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(CAST(value AS REAL)), 0)
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierDistanceWalkingRunning'
		AND value IS NOT NULL
	`).Scan(&totalDistanceMeters)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	summary.TotalDistance = totalDistanceMeters / 1000.0
	if summary.ActiveDays > 0 {
		summary.AvgDailyDistance = summary.TotalDistance / float64(summary.ActiveDays)
	}

	// Get calorie statistics
	err = db.QueryRow(`
		SELECT COALESCE(SUM(CAST(value AS REAL)), 0)
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierActiveEnergyBurned'
		AND value IS NOT NULL
	`).Scan(&summary.TotalCalories)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if summary.ActiveDays > 0 {
		summary.AvgDailyCalories = summary.TotalCalories / float64(summary.ActiveDays)
	}

	// Get workout statistics
	err = db.QueryRow(`
		SELECT
			COUNT(*) as total_workouts,
			COALESCE(MAX((julianday(end_date) - julianday(start_date)) * 24 * 60), 0) as longest_workout,
			COALESCE(SUM((julianday(end_date) - julianday(start_date)) * 24), 0) as total_hours
		FROM workouts
	`).Scan(&summary.TotalWorkouts, &summary.LongestWorkout, &summary.TotalExerciseTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get most active day
	err = db.QueryRow(`
		SELECT DATE(start_date)
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		GROUP BY DATE(start_date)
		ORDER BY SUM(CAST(value AS INTEGER)) DESC
		LIMIT 1
	`).Scan(&summary.MostActiveDay)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return summary, nil
}

func getDailyExerciseStats(db *sql.DB) ([]DailyExerciseStats, error) {
	rows, err := db.Query(`
		SELECT
			DATE(start_date) as date,
			COALESCE(SUM(CASE WHEN type = 'HKQuantityTypeIdentifierStepCount' THEN CAST(value AS INTEGER) ELSE 0 END), 0) as steps,
			COALESCE(SUM(CASE WHEN type = 'HKQuantityTypeIdentifierDistanceWalkingRunning' THEN CAST(value AS REAL) ELSE 0 END) / 1000.0, 0) as distance_km,
			COALESCE(SUM(CASE WHEN type = 'HKQuantityTypeIdentifierActiveEnergyBurned' THEN CAST(value AS REAL) ELSE 0 END), 0) as calories,
			COALESCE(SUM(CASE WHEN type = 'HKQuantityTypeIdentifierFlightsClimbed' THEN CAST(value AS INTEGER) ELSE 0 END), 0) as flights
		FROM health_records
		WHERE type IN (
			'HKQuantityTypeIdentifierStepCount',
			'HKQuantityTypeIdentifierDistanceWalkingRunning',
			'HKQuantityTypeIdentifierActiveEnergyBurned',
			'HKQuantityTypeIdentifierFlightsClimbed'
		)
		GROUP BY DATE(start_date)
		ORDER BY date DESC
		LIMIT 90
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []DailyExerciseStats{}
	for rows.Next() {
		var s DailyExerciseStats
		err := rows.Scan(&s.Date, &s.Steps, &s.Distance, &s.Calories, &s.FlightsClimbed)
		if err != nil {
			continue
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func getWorkoutTypeStats(db *sql.DB) ([]WorkoutTypeStats, error) {
	rows, err := db.Query(`
		SELECT
			workout_type,
			COUNT(*) as count,
			SUM((julianday(end_date) - julianday(start_date)) * 24 * 60) as total_duration_minutes,
			COALESCE(SUM(CAST(total_distance AS REAL)), 0) / 1000.0 as total_distance_km,
			COALESCE(SUM(CAST(total_energy_burned AS REAL)), 0) as total_calories
		FROM workouts
		WHERE workout_type IS NOT NULL
		GROUP BY workout_type
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []WorkoutTypeStats{}
	totalWorkouts := 0

	// First pass: collect data and count total
	tempStats := []WorkoutTypeStats{}
	for rows.Next() {
		var s WorkoutTypeStats
		err := rows.Scan(&s.WorkoutType, &s.Count, &s.TotalDuration, &s.TotalDistance, &s.TotalCalories)
		if err != nil {
			continue
		}

		if s.Count > 0 {
			s.AvgDuration = s.TotalDuration / float64(s.Count)
			s.AvgDistance = s.TotalDistance / float64(s.Count)
			s.AvgCalories = s.TotalCalories / float64(s.Count)
		}

		totalWorkouts += s.Count
		tempStats = append(tempStats, s)
	}

	// Second pass: calculate percentages
	for _, s := range tempStats {
		if totalWorkouts > 0 {
			s.Percentage = float64(s.Count) / float64(totalWorkouts) * 100.0
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func getCalorieTrend(db *sql.DB) ([]CalorieDataPoint, error) {
	rows, err := db.Query(`
		SELECT
			DATE(start_date) as date,
			COALESCE(SUM(CASE WHEN type = 'HKQuantityTypeIdentifierActiveEnergyBurned' THEN CAST(value AS REAL) ELSE 0 END), 0) as active_calories,
			COALESCE(SUM(CASE WHEN type = 'HKQuantityTypeIdentifierBasalEnergyBurned' THEN CAST(value AS REAL) ELSE 0 END), 0) as resting_calories
		FROM health_records
		WHERE type IN ('HKQuantityTypeIdentifierActiveEnergyBurned', 'HKQuantityTypeIdentifierBasalEnergyBurned')
		GROUP BY DATE(start_date)
		ORDER BY date DESC
		LIMIT 90
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trend := []CalorieDataPoint{}
	for rows.Next() {
		var point CalorieDataPoint
		err := rows.Scan(&point.Date, &point.ActiveCalories, &point.RestingCalories)
		if err != nil {
			continue
		}
		point.TotalCalories = point.ActiveCalories + point.RestingCalories
		trend = append(trend, point)
	}

	return trend, nil
}

func getIntensityAnalysis(db *sql.DB) (*IntensityAnalysis, error) {
	analysis := &IntensityAnalysis{}

	// Get exercise minutes by intensity (using heart rate zones as proxy)
	// Low: <64% max HR, Moderate: 64-76%, High: >76%
	// This is a simplified approach - real METs would require more detailed data

	// For now, use workout duration as a proxy
	var totalMinutes float64
	err := db.QueryRow(`
		SELECT COALESCE(SUM((julianday(end_date) - julianday(start_date)) * 24 * 60), 0)
		FROM workouts
	`).Scan(&totalMinutes)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Estimate intensity distribution (simplified)
	// Assume 30% low, 50% moderate, 20% high intensity
	analysis.LowIntensity.Minutes = int(totalMinutes * 0.3)
	analysis.LowIntensity.Percentage = 30.0
	analysis.ModerateIntensity.Minutes = int(totalMinutes * 0.5)
	analysis.ModerateIntensity.Percentage = 50.0
	analysis.HighIntensity.Minutes = int(totalMinutes * 0.2)
	analysis.HighIntensity.Percentage = 20.0

	// Estimate average METs (3.0 for low, 5.0 for moderate, 8.0 for high)
	analysis.AvgMETs = (3.0*0.3 + 5.0*0.5 + 8.0*0.2)

	// Calculate intensity score (0-100)
	// Based on WHO recommendations: 150 min moderate or 75 min vigorous per week
	weeklyModerate := float64(analysis.ModerateIntensity.Minutes) / 4.0 // Approximate weekly
	weeklyVigorous := float64(analysis.HighIntensity.Minutes) / 4.0

	score := (weeklyModerate/150.0 + weeklyVigorous/75.0) * 50.0
	analysis.IntensityScore = math.Min(100.0, score)

	return analysis, nil
}

func detectAchievements(dailyStats []DailyExerciseStats, workoutTypes []WorkoutTypeStats) []Achievement {
	achievements := []Achievement{}

	if len(dailyStats) == 0 {
		return achievements
	}

	// Find max steps day
	maxSteps := int64(0)
	maxStepsDate := ""
	for _, day := range dailyStats {
		if day.Steps > maxSteps {
			maxSteps = day.Steps
			maxStepsDate = day.Date
		}
	}
	if maxSteps >= 10000 {
		achievements = append(achievements, Achievement{
			Title:       "步数之王",
			Description: "单日步数最高记录",
			Value:       float64(maxSteps),
			Unit:        "步",
			Date:        maxStepsDate,
			Type:        "steps",
		})
	}

	// Find max distance day
	maxDistance := 0.0
	maxDistanceDate := ""
	for _, day := range dailyStats {
		if day.Distance > maxDistance {
			maxDistance = day.Distance
			maxDistanceDate = day.Date
		}
	}
	if maxDistance >= 5.0 {
		achievements = append(achievements, Achievement{
			Title:       "长距离挑战",
			Description: "单日行走距离最远",
			Value:       maxDistance,
			Unit:        "km",
			Date:        maxDistanceDate,
			Type:        "distance",
		})
	}

	// Find max calories day
	maxCalories := 0.0
	maxCaloriesDate := ""
	for _, day := range dailyStats {
		if day.Calories > maxCalories {
			maxCalories = day.Calories
			maxCaloriesDate = day.Date
		}
	}
	if maxCalories >= 500.0 {
		achievements = append(achievements, Achievement{
			Title:       "燃脂达人",
			Description: "单日消耗卡路里最高",
			Value:       maxCalories,
			Unit:        "kcal",
			Date:        maxCaloriesDate,
			Type:        "calories",
		})
	}

	// Most frequent workout type
	if len(workoutTypes) > 0 {
		topWorkout := workoutTypes[0]
		if topWorkout.Count >= 5 {
			achievements = append(achievements, Achievement{
				Title:       "运动专家",
				Description: fmt.Sprintf("最常进行的运动: %s", topWorkout.WorkoutType),
				Value:       float64(topWorkout.Count),
				Unit:        "次",
				Date:        "",
				Type:        "workout",
			})
		}
	}

	return achievements
}

func generateExerciseRecommendations(summary *ExerciseSummary, intensity *IntensityAnalysis) []string {
	recommendations := []string{}

	// Step count recommendations
	if summary.AvgDailySteps < 5000 {
		recommendations = append(recommendations, "建议增加日常步数，目标每天至少5000步")
	} else if summary.AvgDailySteps < 10000 {
		recommendations = append(recommendations, "步数表现良好，可以尝试达到每天10000步的目标")
	} else {
		recommendations = append(recommendations, "步数表现优秀！保持当前的活动水平")
	}

	// Intensity recommendations
	if intensity.IntensityScore < 50 {
		recommendations = append(recommendations, "运动强度偏低，建议增加中高强度运动")
		recommendations = append(recommendations, "WHO建议：每周至少150分钟中等强度或75分钟高强度运动")
	} else if intensity.IntensityScore < 80 {
		recommendations = append(recommendations, "运动强度适中，可以适当增加高强度间歇训练")
	} else {
		recommendations = append(recommendations, "运动强度很好！注意运动后的恢复和休息")
	}

	// Workout variety
	if summary.TotalWorkouts < 10 {
		recommendations = append(recommendations, "建议增加运动频率，尝试多样化的运动类型")
	}

	// Calorie recommendations
	if summary.AvgDailyCalories < 200 {
		recommendations = append(recommendations, "日均消耗卡路里较低，建议增加有氧运动")
	}

	return recommendations
}
