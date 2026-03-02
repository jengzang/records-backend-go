package analysis

import (
	"database/sql"
	"fmt"
)

// HealthRankings represents health data rankings and personal bests
type HealthRankings struct {
	StepsRankings      []RankingEntry `json:"stepsRankings"`
	DistanceRankings   []RankingEntry `json:"distanceRankings"`
	CaloriesRankings   []RankingEntry `json:"caloriesRankings"`
	HeartRateRankings  []RankingEntry `json:"heartRateRankings"`
	SleepRankings      []RankingEntry `json:"sleepRankings"`
	WorkoutRankings    []RankingEntry `json:"workoutRankings"`
	PersonalBests      PersonalBests  `json:\"personalBests\"`
	Summary            RankingSummary `json:"summary"`
}

// RankingEntry represents a single ranking entry
type RankingEntry struct {
	Rank        int     `json:"rank"`
	Date        string  `json:"date"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Description string  `json:"description"`
	IsCurrent   bool    `json:"isCurrent"` // Is this from current month
}

// PersonalBests represents personal best records
type PersonalBests struct {
	MaxSteps          RankingEntry `json:"maxSteps"`
	MaxDistance       RankingEntry `json:"maxDistance"`
	MaxCalories       RankingEntry `json:"maxCalories"`
	LowestRestingHR   RankingEntry `json:"lowestRestingHR"`
	LongestSleep      RankingEntry `json:"longestSleep"`
	LongestWorkout    RankingEntry `json:"longestWorkout"`
	MostWorkoutsDay   RankingEntry `json:"mostWorkoutsDay"`
}

// RankingSummary provides overall ranking statistics
type RankingSummary struct {
	TotalDaysTracked  int     `json:"totalDaysTracked"`
	Top10DaysPercent  float64 `json:"top10DaysPercent"`  // Percentage of days in top 10
	CurrentStreak     int     `json:"currentStreak"`     // Days of consecutive activity
	LongestStreak     int     `json:"longestStreak"`
	AverageRank       float64 `json:"averageRank"`
	Improvement       string  `json:"improvement"`       // "improving", "stable", "declining"
}

// GetHealthRankings retrieves health data rankings and personal bests
func GetHealthRankings(db *sql.DB) (*HealthRankings, error) {
	rankings := &HealthRankings{}

	// Get steps rankings (top 20 days)
	stepsRankings, err := getStepsRankings(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get steps rankings: %w", err)
	}
	rankings.StepsRankings = stepsRankings

	// Get distance rankings
	distanceRankings, err := getDistanceRankings(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get distance rankings: %w", err)
	}
	rankings.DistanceRankings = distanceRankings

	// Get calories rankings
	caloriesRankings, err := getCaloriesRankings(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get calories rankings: %w", err)
	}
	rankings.CaloriesRankings = caloriesRankings

	// Get heart rate rankings (lowest resting HR)
	heartRateRankings, err := getHeartRateRankings(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get heart rate rankings: %w", err)
	}
	rankings.HeartRateRankings = heartRateRankings

	// Get sleep rankings (longest sleep duration)
	sleepRankings, err := getSleepRankings(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get sleep rankings: %w", err)
	}
	rankings.SleepRankings = sleepRankings

	// Get workout rankings
	workoutRankings, err := getWorkoutRankings(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout rankings: %w", err)
	}
	rankings.WorkoutRankings = workoutRankings

	// Get personal bests
	personalBests, err := getPersonalBests(db, stepsRankings, distanceRankings, caloriesRankings, heartRateRankings, sleepRankings, workoutRankings)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal bests: %w", err)
	}
	rankings.PersonalBests = *personalBests

	// Calculate summary statistics
	summary, err := calculateRankingSummary(db, stepsRankings)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate summary: %w", err)
	}
	rankings.Summary = *summary

	return rankings, nil
}

func getStepsRankings(db *sql.DB) ([]RankingEntry, error) {
	query := `
		SELECT
			DATE(start_date) as date,
			CAST(SUM(value) AS INTEGER) as total_steps
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		AND value IS NOT NULL
		GROUP BY DATE(start_date)
		HAVING total_steps > 0
		ORDER BY total_steps DESC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []RankingEntry{}
	rank := 1
	for rows.Next() {
		var entry RankingEntry
		var steps int64
		if err := rows.Scan(&entry.Date, &steps); err != nil {
			continue
		}

		entry.Rank = rank
		entry.Value = float64(steps)
		entry.Unit = "步"
		entry.Description = fmt.Sprintf("%d步", steps)
		entry.IsCurrent = isCurrentMonth(entry.Date)

		rankings = append(rankings, entry)
		rank++
	}

	return rankings, nil
}

func getDistanceRankings(db *sql.DB) ([]RankingEntry, error) {
	query := `
		SELECT
			DATE(start_date) as date,
			CAST(SUM(value) AS REAL) / 1000.0 as total_distance_km
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierDistanceWalkingRunning'
		AND value IS NOT NULL
		GROUP BY DATE(start_date)
		HAVING total_distance_km > 0
		ORDER BY total_distance_km DESC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []RankingEntry{}
	rank := 1
	for rows.Next() {
		var entry RankingEntry
		if err := rows.Scan(&entry.Date, &entry.Value); err != nil {
			continue
		}

		entry.Rank = rank
		entry.Unit = "km"
		entry.Description = fmt.Sprintf("%.2f km", entry.Value)
		entry.IsCurrent = isCurrentMonth(entry.Date)

		rankings = append(rankings, entry)
		rank++
	}

	return rankings, nil
}

func getCaloriesRankings(db *sql.DB) ([]RankingEntry, error) {
	query := `
		SELECT
			DATE(start_date) as date,
			CAST(SUM(value) AS REAL) as total_calories
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierActiveEnergyBurned'
		AND value IS NOT NULL
		GROUP BY DATE(start_date)
		HAVING total_calories > 0
		ORDER BY total_calories DESC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []RankingEntry{}
	rank := 1
	for rows.Next() {
		var entry RankingEntry
		if err := rows.Scan(&entry.Date, &entry.Value); err != nil {
			continue
		}

		entry.Rank = rank
		entry.Unit = "kcal"
		entry.Description = fmt.Sprintf("%.0f kcal", entry.Value)
		entry.IsCurrent = isCurrentMonth(entry.Date)

		rankings = append(rankings, entry)
		rank++
	}

	return rankings, nil
}

func getHeartRateRankings(db *sql.DB) ([]RankingEntry, error) {
	// Get lowest resting heart rate days
	query := `
		SELECT
			DATE(start_date) as date,
			CAST(AVG(value) AS REAL) as avg_hr
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierRestingHeartRate'
		AND value IS NOT NULL
		GROUP BY DATE(start_date)
		HAVING avg_hr > 0
		ORDER BY avg_hr ASC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []RankingEntry{}
	rank := 1
	for rows.Next() {
		var entry RankingEntry
		if err := rows.Scan(&entry.Date, &entry.Value); err != nil {
			continue
		}

		entry.Rank = rank
		entry.Unit = "bpm"
		entry.Description = fmt.Sprintf("%.0f bpm (静息心率)", entry.Value)
		entry.IsCurrent = isCurrentMonth(entry.Date)

		rankings = append(rankings, entry)
		rank++
	}

	return rankings, nil
}

func getSleepRankings(db *sql.DB) ([]RankingEntry, error) {
	// Get longest sleep duration days
	query := `
		SELECT
			DATE(start_date) as date,
			SUM((julianday(end_date) - julianday(start_date)) * 24) as sleep_hours
		FROM health_records
		WHERE type = 'HKCategoryTypeIdentifierSleepAnalysis'
		AND value IS NOT NULL
		GROUP BY DATE(start_date)
		HAVING sleep_hours > 0
		ORDER BY sleep_hours DESC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []RankingEntry{}
	rank := 1
	for rows.Next() {
		var entry RankingEntry
		if err := rows.Scan(&entry.Date, &entry.Value); err != nil {
			continue
		}

		entry.Rank = rank
		entry.Unit = "小时"
		entry.Description = fmt.Sprintf("%.1f 小时", entry.Value)
		entry.IsCurrent = isCurrentMonth(entry.Date)

		rankings = append(rankings, entry)
		rank++
	}

	return rankings, nil
}

func getWorkoutRankings(db *sql.DB) ([]RankingEntry, error) {
	// Get days with most workout duration
	query := `
		SELECT
			DATE(start_date) as date,
			SUM((julianday(end_date) - julianday(start_date)) * 24 * 60) as workout_minutes
		FROM workouts
		GROUP BY DATE(start_date)
		HAVING workout_minutes > 0
		ORDER BY workout_minutes DESC
		LIMIT 20
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []RankingEntry{}
	rank := 1
	for rows.Next() {
		var entry RankingEntry
		if err := rows.Scan(&entry.Date, &entry.Value); err != nil {
			continue
		}

		entry.Rank = rank
		entry.Unit = "分钟"
		entry.Description = fmt.Sprintf("%.0f 分钟", entry.Value)
		entry.IsCurrent = isCurrentMonth(entry.Date)

		rankings = append(rankings, entry)
		rank++
	}

	return rankings, nil
}

func getPersonalBests(db *sql.DB, steps, distance, calories, heartRate, sleep, workout []RankingEntry) (*PersonalBests, error) {
	bests := &PersonalBests{}

	if len(steps) > 0 {
		bests.MaxSteps = steps[0]
	}
	if len(distance) > 0 {
		bests.MaxDistance = distance[0]
	}
	if len(calories) > 0 {
		bests.MaxCalories = calories[0]
	}
	if len(heartRate) > 0 {
		bests.LowestRestingHR = heartRate[0]
	}
	if len(sleep) > 0 {
		bests.LongestSleep = sleep[0]
	}
	if len(workout) > 0 {
		bests.LongestWorkout = workout[0]
	}

	// Get most workouts in a single day
	var date string
	var count int
	err := db.QueryRow(`
		SELECT DATE(start_date), COUNT(*) as workout_count
		FROM workouts
		GROUP BY DATE(start_date)
		ORDER BY workout_count DESC
		LIMIT 1
	`).Scan(&date, &count)
	if err == nil {
		bests.MostWorkoutsDay = RankingEntry{
			Rank:        1,
			Date:        date,
			Value:       float64(count),
			Unit:        "次",
			Description: fmt.Sprintf("%d 次运动", count),
			IsCurrent:   isCurrentMonth(date),
		}
	}

	return bests, nil
}

func calculateRankingSummary(db *sql.DB, stepsRankings []RankingEntry) (*RankingSummary, error) {
	summary := &RankingSummary{}

	// Get total days tracked
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT DATE(start_date))
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
	`).Scan(&summary.TotalDaysTracked)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Calculate top 10 days percentage
	currentMonthCount := 0
	for _, entry := range stepsRankings {
		if entry.IsCurrent && entry.Rank <= 10 {
			currentMonthCount++
		}
	}
	if len(stepsRankings) > 0 {
		summary.Top10DaysPercent = float64(currentMonthCount) / 10.0 * 100.0
	}

	// Calculate current streak (consecutive days with >5000 steps)
	summary.CurrentStreak = calculateCurrentStreak(db)
	summary.LongestStreak = calculateLongestStreak(db)

	// Determine improvement trend
	summary.Improvement = determineImprovementTrend(db)

	return summary, nil
}

func calculateCurrentStreak(db *sql.DB) int {
	query := `
		SELECT DATE(start_date), SUM(CAST(value AS INTEGER)) as steps
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		GROUP BY DATE(start_date)
		ORDER BY DATE(start_date) DESC
		LIMIT 30
	`

	rows, err := db.Query(query)
	if err != nil {
		return 0
	}
	defer rows.Close()

	streak := 0
	for rows.Next() {
		var date string
		var steps int64
		if err := rows.Scan(&date, &steps); err != nil {
			break
		}

		if steps >= 5000 {
			streak++
		} else {
			break
		}
	}

	return streak
}

func calculateLongestStreak(db *sql.DB) int {
	query := `
		SELECT DATE(start_date), SUM(CAST(value AS INTEGER)) as steps
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		GROUP BY DATE(start_date)
		ORDER BY DATE(start_date) ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return 0
	}
	defer rows.Close()

	maxStreak := 0
	currentStreak := 0

	for rows.Next() {
		var date string
		var steps int64
		if err := rows.Scan(&date, &steps); err != nil {
			break
		}

		if steps >= 5000 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}

	return maxStreak
}

func determineImprovementTrend(db *sql.DB) string {
	// Compare last 30 days average with previous 30 days
	var recentAvg, previousAvg float64

	db.QueryRow(`
		SELECT AVG(CAST(value AS INTEGER))
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		AND DATE(start_date) >= DATE('now', '-30 days')
	`).Scan(&recentAvg)

	db.QueryRow(`
		SELECT AVG(CAST(value AS INTEGER))
		FROM health_records
		WHERE type = 'HKQuantityTypeIdentifierStepCount'
		AND DATE(start_date) >= DATE('now', '-60 days')
		AND DATE(start_date) < DATE('now', '-30 days')
	`).Scan(&previousAvg)

	if recentAvg > previousAvg*1.1 {
		return "improving"
	} else if recentAvg < previousAvg*0.9 {
		return "declining"
	}
	return "stable"
}

func isCurrentMonth(dateStr string) bool {
	// Simple check if date is in current month (YYYY-MM format)
	// This is a simplified implementation
	return len(dateStr) >= 7 && dateStr[:7] == "2026-03" // Adjust as needed
}
