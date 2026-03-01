package analysis

import (
	"database/sql"
	"fmt"
	"time"
)

// WeekdayWeekendComparison represents weekday vs weekend keyboard usage comparison
type WeekdayWeekendComparison struct {
	Weekday struct {
		TotalKeystrokes    int                  `json:"totalKeystrokes"`
		AvgDailyKeystrokes float64              `json:"avgDailyKeystrokes"`
		CategoryDistribution map[string]int     `json:"categoryDistribution"`
		HandDistribution   map[string]int       `json:"handDistribution"`
		HourlyDistribution []HourlyKeystrokeStat `json:"hourlyDistribution"`
		TopKeys            []KeyStat             `json:"topKeys"`
		DayCount           int                   `json:"dayCount"`
	} `json:"weekday"`
	Weekend struct {
		TotalKeystrokes    int                  `json:"totalKeystrokes"`
		AvgDailyKeystrokes float64              `json:"avgDailyKeystrokes"`
		CategoryDistribution map[string]int     `json:"categoryDistribution"`
		HandDistribution   map[string]int       `json:"handDistribution"`
		HourlyDistribution []HourlyKeystrokeStat `json:"hourlyDistribution"`
		TopKeys            []KeyStat             `json:"topKeys"`
		DayCount           int                   `json:"dayCount"`
	} `json:"weekend"`
	Insights []string `json:"insights"`
}

// HourlyKeystrokeStat represents keystroke statistics for a specific hour
type HourlyKeystrokeStat struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// GetWeekdayWeekendComparison calculates weekday vs weekend keyboard usage comparison
func GetWeekdayWeekendComparison(db *sql.DB, getKeyCategory func(int) string, getKeyHand func(int) string, getKeyName func(int) string) (*WeekdayWeekendComparison, error) {
	comparison := &WeekdayWeekendComparison{}
	comparison.Weekday.CategoryDistribution = make(map[string]int)
	comparison.Weekday.HandDistribution = make(map[string]int)
	comparison.Weekday.HourlyDistribution = make([]HourlyKeystrokeStat, 24)
	comparison.Weekend.CategoryDistribution = make(map[string]int)
	comparison.Weekend.HandDistribution = make(map[string]int)
	comparison.Weekend.HourlyDistribution = make([]HourlyKeystrokeStat, 24)

	// Initialize hourly distribution
	for i := 0; i < 24; i++ {
		comparison.Weekday.HourlyDistribution[i].Hour = i
		comparison.Weekend.HourlyDistribution[i].Hour = i
	}

	// Get weekday data
	if err := getWeekdayData(db, &comparison.Weekday, false, getKeyCategory, getKeyHand, getKeyName); err != nil {
		return nil, fmt.Errorf("failed to get weekday data: %w", err)
	}

	// Get weekend data
	if err := getWeekdayData(db, &comparison.Weekend, true, getKeyCategory, getKeyHand, getKeyName); err != nil {
		return nil, fmt.Errorf("failed to get weekend data: %w", err)
	}

	// Generate insights
	comparison.Insights = generateWeekdayWeekendInsights(comparison)

	return comparison, nil
}

// getWeekdayData retrieves keyboard usage data for weekday or weekend
func getWeekdayData(db *sql.DB, data interface{}, isWeekend bool, getKeyCategory func(int) string, getKeyHand func(int) string, getKeyName func(int) string) error {
	// Query all keyboard data with date filtering
	query := `
	SELECT k.date, k.scancode, k.count
	FROM keyboard_data k
	WHERE 1=1
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query keyboard data: %w", err)
	}
	defer rows.Close()

	// Temporary storage for processing
	totalKeystrokes := 0
	categoryDist := make(map[string]int)
	handDist := make(map[string]int)
	hourlyDist := make([]int, 24)
	keyCount := make(map[string]int)
	dateSet := make(map[string]bool)

	for rows.Next() {
		var date string
		var scancode, count int
		if err := rows.Scan(&date, &scancode, &count); err != nil {
			continue
		}

		// Parse date to determine if it's weekday or weekend
		t, err := time.Parse("20060102", date)
		if err != nil {
			continue
		}

		weekday := t.Weekday()
		isWeekendDay := (weekday == time.Saturday || weekday == time.Sunday)

		// Skip if not matching the requested type
		if isWeekend != isWeekendDay {
			continue
		}

		// Track unique dates
		dateSet[date] = true

		// Accumulate statistics
		totalKeystrokes += count

		// Category distribution
		category := getKeyCategory(scancode)
		categoryDist[category] += count

		// Hand distribution
		hand := getKeyHand(scancode)
		handDist[hand] += count

		// Key count for top keys
		keyName := getKeyName(scancode)
		keyCount[keyName] += count

		// Note: We don't have hourly data in the current schema
		// This would require additional data collection
	}

	// Type assertion to set values
	switch v := data.(type) {
	case *struct {
		TotalKeystrokes      int
		AvgDailyKeystrokes   float64
		CategoryDistribution map[string]int
		HandDistribution     map[string]int
		HourlyDistribution   []HourlyKeystrokeStat
		TopKeys              []KeyStat
		DayCount             int
	}:
		v.TotalKeystrokes = totalKeystrokes
		v.DayCount = len(dateSet)
		if v.DayCount > 0 {
			v.AvgDailyKeystrokes = float64(totalKeystrokes) / float64(v.DayCount)
		}
		v.CategoryDistribution = categoryDist
		v.HandDistribution = handDist
		v.TopKeys = getTopKeys(keyCount, 10)

		// Set hourly distribution (currently all zeros without hourly data)
		for i := 0; i < 24; i++ {
			v.HourlyDistribution[i].Count = hourlyDist[i]
		}
	}

	return nil
}

// generateWeekdayWeekendInsights generates insights based on weekday vs weekend comparison
func generateWeekdayWeekendInsights(comparison *WeekdayWeekendComparison) []string {
	insights := []string{}

	// Compare total keystrokes
	if comparison.Weekend.AvgDailyKeystrokes > comparison.Weekday.AvgDailyKeystrokes*1.2 {
		insights = append(insights, fmt.Sprintf("周末平均按键数(%.0f)明显高于工作日(%.0f)，可能周末工作较多",
			comparison.Weekend.AvgDailyKeystrokes, comparison.Weekday.AvgDailyKeystrokes))
	} else if comparison.Weekday.AvgDailyKeystrokes > comparison.Weekend.AvgDailyKeystrokes*1.5 {
		insights = append(insights, fmt.Sprintf("工作日平均按键数(%.0f)显著高于周末(%.0f)，工作强度较大",
			comparison.Weekday.AvgDailyKeystrokes, comparison.Weekend.AvgDailyKeystrokes))
	} else {
		insights = append(insights, "工作日和周末的按键数较为接近，使用习惯稳定")
	}

	// Compare category distribution
	weekdayLetters := comparison.Weekday.CategoryDistribution["letter"]
	weekendLetters := comparison.Weekend.CategoryDistribution["letter"]
	weekdayNumbers := comparison.Weekday.CategoryDistribution["number"]
	weekendNumbers := comparison.Weekend.CategoryDistribution["number"]

	if weekdayLetters > 0 && weekendLetters > 0 {
		weekdayLetterRatio := float64(weekdayLetters) / float64(comparison.Weekday.TotalKeystrokes)
		weekendLetterRatio := float64(weekendLetters) / float64(comparison.Weekend.TotalKeystrokes)

		if weekdayLetterRatio > weekendLetterRatio*1.1 {
			insights = append(insights, "工作日字母键使用比例更高，可能涉及更多文字输入工作")
		} else if weekendLetterRatio > weekdayLetterRatio*1.1 {
			insights = append(insights, "周末字母键使用比例更高，可能周末有更多写作或编程活动")
		}
	}

	if weekdayNumbers > 0 && weekendNumbers > 0 {
		weekdayNumberRatio := float64(weekdayNumbers) / float64(comparison.Weekday.TotalKeystrokes)
		weekendNumberRatio := float64(weekendNumbers) / float64(comparison.Weekend.TotalKeystrokes)

		if weekdayNumberRatio > weekendNumberRatio*1.2 {
			insights = append(insights, "工作日数字键使用比例更高，可能涉及数据处理或财务工作")
		}
	}

	// Compare hand distribution
	weekdayLeftHand := comparison.Weekday.HandDistribution["left"]
	weekdayRightHand := comparison.Weekday.HandDistribution["right"]
	weekendLeftHand := comparison.Weekend.HandDistribution["left"]
	weekendRightHand := comparison.Weekend.HandDistribution["right"]

	if weekdayLeftHand+weekdayRightHand > 0 && weekendLeftHand+weekendRightHand > 0 {
		weekdayLeftRatio := float64(weekdayLeftHand) / float64(weekdayLeftHand+weekdayRightHand)
		weekendLeftRatio := float64(weekendLeftHand) / float64(weekendLeftHand+weekendRightHand)

		if weekdayLeftRatio > weekendLeftRatio+0.05 {
			insights = append(insights, "工作日左手使用比例更高，可能工作内容偏向左手区域按键")
		} else if weekendLeftRatio > weekdayLeftRatio+0.05 {
			insights = append(insights, "周末左手使用比例更高，使用习惯有所不同")
		}
	}

	// Top keys comparison
	if len(comparison.Weekday.TopKeys) > 0 && len(comparison.Weekend.TopKeys) > 0 {
		weekdayTop := comparison.Weekday.TopKeys[0].KeyName
		weekendTop := comparison.Weekend.TopKeys[0].KeyName
		insights = append(insights, fmt.Sprintf("工作日最常用按键: %s, 周末最常用按键: %s", weekdayTop, weekendTop))
	}

	// Data coverage
	insights = append(insights, fmt.Sprintf("数据覆盖: 工作日%d天, 周末%d天", comparison.Weekday.DayCount, comparison.Weekend.DayCount))

	return insights
}
