package screentime

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// TimeWasteDetection represents time waste analysis results
type TimeWasteDetection struct {
	TotalWastedTime   int64           `json:"totalWastedTime"`   // milliseconds
	WastePercentage   float64         `json:"wastePercentage"`
	WasteScenarios    []WasteScenario `json:"wasteScenarios"`
	Recommendations   []string        `json:"recommendations"`
	DateRange         DateRange       `json:"dateRange"`
}

// WasteScenario represents a specific time waste pattern
type WasteScenario struct {
	Type              string  `json:"type"`              // "work_hours_entertainment", "late_night_social", "fragmented"
	Description       string  `json:"description"`
	WastedTime        int64   `json:"wastedTime"`        // milliseconds
	Percentage        float64 `json:"percentage"`
	AffectedApps      []string `json:"affectedApps"`
	OccurrenceCount   int     `json:"occurrenceCount"`
	ImprovementTip    string  `json:"improvementTip"`
}

// DateRange represents a date range
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// GetTimeWasteDetection analyzes time waste patterns
// GET /api/v1/screentime/analysis/time-waste?start=20240101&end=20241231
func (h *Handler) GetTimeWasteDetection(c *gin.Context) {
	logger.Info("Time waste detection requested", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")

	var detection TimeWasteDetection
	detection.DateRange.Start = start
	detection.DateRange.End = end
	detection.WasteScenarios = []WasteScenario{}
	detection.Recommendations = []string{}

	// Get total screen time for percentage calculation
	var totalDuration int64
	totalQuery := `SELECT SUM(duration_ms) FROM screentime_daily WHERE package_id != 'ALL'`
	args := []interface{}{}

	if start != "" {
		totalQuery += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		totalQuery += " AND date <= ?"
		args = append(args, end)
	}

	h.db.QueryRow(totalQuery, args...).Scan(&totalDuration)

	// 1. Detect work hours entertainment usage (9:00-18:00)
	workHoursWaste := h.detectWorkHoursEntertainment(start, end)
	if workHoursWaste.WastedTime > 0 {
		detection.WasteScenarios = append(detection.WasteScenarios, workHoursWaste)
		detection.TotalWastedTime += workHoursWaste.WastedTime
	}

	// 2. Detect late night social media usage (23:00-02:00)
	lateNightWaste := h.detectLateNightSocial(start, end)
	if lateNightWaste.WastedTime > 0 {
		detection.WasteScenarios = append(detection.WasteScenarios, lateNightWaste)
		detection.TotalWastedTime += lateNightWaste.WastedTime
	}

	// 3. Detect fragmented usage (<5min sessions, frequent switching)
	fragmentedWaste := h.detectFragmentedUsage(start, end)
	if fragmentedWaste.WastedTime > 0 {
		detection.WasteScenarios = append(detection.WasteScenarios, fragmentedWaste)
		detection.TotalWastedTime += fragmentedWaste.WastedTime
	}

	// Calculate waste percentage
	if totalDuration > 0 {
		detection.WastePercentage = float64(detection.TotalWastedTime) / float64(totalDuration) * 100
	}

	// Generate recommendations
	detection.Recommendations = h.generateWasteRecommendations(detection.WasteScenarios)

	c.JSON(http.StatusOK, detection)
}

// detectWorkHoursEntertainment detects entertainment app usage during work hours
func (h *Handler) detectWorkHoursEntertainment(start, end string) WasteScenario {
	scenario := WasteScenario{
		Type:        "work_hours_entertainment",
		Description: "工作时间使用娱乐应用",
		AffectedApps: []string{},
	}

	// Query sessions during work hours (9:00-18:00) with entertainment categories
	query := `
	SELECT
		s.app_name,
		SUM(CAST((julianday(s.end_time) - julianday(s.start_time)) * 86400000 AS INTEGER)) as duration_ms,
		COUNT(*) as count
	FROM screentime_sessions s
	JOIN screentime_apps a ON s.package_id = a.package_id
	WHERE a.category IN ('Entertainment', 'Social', 'Games', 'Video')
		AND CAST(substr(s.start_time, 1, 2) AS INTEGER) >= 9
		AND CAST(substr(s.start_time, 1, 2) AS INTEGER) < 18
	`

	args := []interface{}{}
	if start != "" {
		query += " AND s.date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND s.date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY s.app_name ORDER BY duration_ms DESC LIMIT 5"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		logger.Error("Failed to detect work hours entertainment", err, nil)
		return scenario
	}
	defer rows.Close()

	for rows.Next() {
		var appName string
		var duration int64
		var count int
		if err := rows.Scan(&appName, &duration, &count); err == nil {
			scenario.AffectedApps = append(scenario.AffectedApps, appName)
			scenario.WastedTime += duration
			scenario.OccurrenceCount += count
		}
	}

	if scenario.WastedTime > 0 {
		scenario.ImprovementTip = "建议在工作时间专注于工作相关应用，可以设置应用使用限制或使用专注模式"
	}

	return scenario
}

// detectLateNightSocial detects social media usage late at night
func (h *Handler) detectLateNightSocial(start, end string) WasteScenario {
	scenario := WasteScenario{
		Type:        "late_night_social",
		Description: "深夜使用社交应用",
		AffectedApps: []string{},
	}

	// Query sessions during late night (23:00-02:00) with social categories
	query := `
	SELECT
		s.app_name,
		SUM(CAST((julianday(s.end_time) - julianday(s.start_time)) * 86400000 AS INTEGER)) as duration_ms,
		COUNT(*) as count
	FROM screentime_sessions s
	JOIN screentime_apps a ON s.package_id = a.package_id
	WHERE a.category IN ('Social', 'Entertainment', 'Communication')
		AND (CAST(substr(s.start_time, 1, 2) AS INTEGER) >= 23
		     OR CAST(substr(s.start_time, 1, 2) AS INTEGER) <= 2)
	`

	args := []interface{}{}
	if start != "" {
		query += " AND s.date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND s.date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY s.app_name ORDER BY duration_ms DESC LIMIT 5"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		logger.Error("Failed to detect late night social", err, nil)
		return scenario
	}
	defer rows.Close()

	for rows.Next() {
		var appName string
		var duration int64
		var count int
		if err := rows.Scan(&appName, &duration, &count); err == nil {
			scenario.AffectedApps = append(scenario.AffectedApps, appName)
			scenario.WastedTime += duration
			scenario.OccurrenceCount += count
		}
	}

	if scenario.WastedTime > 0 {
		scenario.ImprovementTip = "深夜使用社交应用会影响睡眠质量，建议在22:00后减少屏幕时间"
	}

	return scenario
}

// detectFragmentedUsage detects fragmented usage patterns
func (h *Handler) detectFragmentedUsage(start, end string) WasteScenario {
	scenario := WasteScenario{
		Type:        "fragmented",
		Description: "碎片化使用（频繁切换应用）",
		AffectedApps: []string{},
	}

	// Query short sessions (<5 minutes) with high frequency
	query := `
	SELECT
		s.app_name,
		COUNT(*) as session_count,
		SUM(CAST((julianday(s.end_time) - julianday(s.start_time)) * 86400000 AS INTEGER)) as total_duration
	FROM screentime_sessions s
	WHERE CAST((julianday(s.end_time) - julianday(s.start_time)) * 86400000 AS INTEGER) < 300000
	`

	args := []interface{}{}
	if start != "" {
		query += " AND s.date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND s.date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY s.app_name HAVING session_count > 10 ORDER BY session_count DESC LIMIT 5"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		logger.Error("Failed to detect fragmented usage", err, nil)
		return scenario
	}
	defer rows.Close()

	for rows.Next() {
		var appName string
		var count int
		var duration int64
		if err := rows.Scan(&appName, &count, &duration); err == nil {
			scenario.AffectedApps = append(scenario.AffectedApps, appName)
			scenario.WastedTime += duration
			scenario.OccurrenceCount += count
		}
	}

	if scenario.WastedTime > 0 {
		scenario.ImprovementTip = "频繁切换应用会降低专注力，建议减少多任务处理，一次专注于一个任务"
	}

	return scenario
}

// generateWasteRecommendations generates improvement recommendations
func (h *Handler) generateWasteRecommendations(scenarios []WasteScenario) []string {
	recommendations := []string{}

	for _, scenario := range scenarios {
		if scenario.WastedTime > 0 {
			recommendations = append(recommendations, scenario.ImprovementTip)
		}
	}

	// Add general recommendations
	if len(scenarios) > 0 {
		recommendations = append(recommendations, "使用应用定时器限制特定应用的使用时间")
		recommendations = append(recommendations, "设置专注模式，在特定时段屏蔽干扰应用")
	}

	return recommendations
}

// AppDependency represents app dependency analysis
type AppDependency struct {
	AppName           string  `json:"appName"`
	PackageID         string  `json:"packageId"`
	Category          string  `json:"category"`
	DependencyScore   int     `json:"dependencyScore"`   // 0-100
	DependencyType    string  `json:"dependencyType"`    // "social", "work", "entertainment"
	UsageFrequency    float64 `json:"usageFrequency"`    // launches per day
	UsagePercentage   float64 `json:"usagePercentage"`   // percentage of total time
	ConsecutiveDays   int     `json:"consecutiveDays"`   // current streak
	FirstOpenTime     string  `json:"firstOpenTime"`     // average first open time
	IsFirstMorningApp bool    `json:"isFirstMorningApp"` // is it the first app opened in morning
	Insights          []string `json:"insights"`
}

// GetAppDependencyAnalysis analyzes app dependency patterns
// GET /api/v1/screentime/analysis/app-dependency?start=20240101&end=20241231&limit=10
func (h *Handler) GetAppDependencyAnalysis(c *gin.Context) {
	logger.Info("App dependency analysis requested", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")
	limit := c.DefaultQuery("limit", "10")

	// Get total duration for percentage calculation
	var totalDuration int64
	totalQuery := `SELECT SUM(total_duration_ms) FROM screentime_apps WHERE package_id != 'ALL'`
	h.db.QueryRow(totalQuery).Scan(&totalDuration)

	// Get total active days
	var totalDays int
	daysQuery := `SELECT COUNT(DISTINCT date) FROM screentime_daily WHERE package_id != 'ALL'`
	args := []interface{}{}
	if start != "" {
		daysQuery += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		daysQuery += " AND date <= ?"
		args = append(args, end)
	}
	h.db.QueryRow(daysQuery, args...).Scan(&totalDays)

	// Query app statistics
	query := `
	SELECT
		a.app_name,
		a.package_id,
		a.category,
		a.total_duration_ms,
		a.total_launches,
		a.first_seen,
		a.last_seen
	FROM screentime_apps a
	WHERE a.package_id != 'ALL'
	ORDER BY a.total_duration_ms DESC
	LIMIT ?
	`

	rows, err := h.db.Query(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dependencies []AppDependency
	for rows.Next() {
		var dep AppDependency
		var totalDurationMS, totalLaunches int64
		var firstSeen, lastSeen string

		if err := rows.Scan(&dep.AppName, &dep.PackageID, &dep.Category, &totalDurationMS, &totalLaunches, &firstSeen, &lastSeen); err != nil {
			continue
		}

		// Calculate usage frequency (launches per day)
		if totalDays > 0 {
			dep.UsageFrequency = float64(totalLaunches) / float64(totalDays)
		}

		// Calculate usage percentage
		if totalDuration > 0 {
			dep.UsagePercentage = float64(totalDurationMS) / float64(totalDuration) * 100
		}

		// Calculate consecutive days streak
		dep.ConsecutiveDays = h.calculateStreak(dep.PackageID, start, end)

		// Get average first open time
		dep.FirstOpenTime = h.getAverageFirstOpenTime(dep.PackageID, start, end)

		// Check if it's the first morning app
		dep.IsFirstMorningApp = h.isFirstMorningApp(dep.PackageID, start, end)

		// Calculate dependency score (0-100)
		dep.DependencyScore = h.calculateDependencyScore(dep.UsageFrequency, dep.UsagePercentage, dep.ConsecutiveDays, dep.IsFirstMorningApp)

		// Determine dependency type
		dep.DependencyType = h.determineDependencyType(dep.Category, dep.UsageFrequency, dep.FirstOpenTime)

		// Generate insights
		dep.Insights = h.generateDependencyInsights(dep)

		dependencies = append(dependencies, dep)
	}

	c.JSON(http.StatusOK, dependencies)
}

// calculateStreak calculates consecutive usage days
func (h *Handler) calculateStreak(packageID, start, end string) int {
	query := `
	SELECT date
	FROM screentime_daily
	WHERE package_id = ?
	`
	args := []interface{}{packageID}

	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	query += " ORDER BY date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if rows.Scan(&date) == nil {
			dates = append(dates, date)
		}
	}

	if len(dates) == 0 {
		return 0
	}

	// Calculate consecutive days from most recent date
	streak := 1
	for i := 0; i < len(dates)-1; i++ {
		date1, _ := time.Parse("20060102", dates[i])
		date2, _ := time.Parse("20060102", dates[i+1])
		diff := date1.Sub(date2).Hours() / 24

		if diff == 1 {
			streak++
		} else {
			break
		}
	}

	return streak
}

// getAverageFirstOpenTime gets average first open time
func (h *Handler) getAverageFirstOpenTime(packageID, start, end string) string {
	query := `
	SELECT AVG(CAST(substr(start_time, 1, 2) AS INTEGER)) as avg_hour
	FROM (
		SELECT date, MIN(start_time) as start_time
		FROM screentime_sessions
		WHERE package_id = ?
	`
	args := []interface{}{packageID}

	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY date)"

	var avgHour sql.NullFloat64
	if err := h.db.QueryRow(query, args...).Scan(&avgHour); err != nil || !avgHour.Valid {
		return "N/A"
	}

	return fmt.Sprintf("%02d:00", int(avgHour.Float64))
}

// isFirstMorningApp checks if app is typically the first opened in morning
func (h *Handler) isFirstMorningApp(packageID, start, end string) bool {
	query := `
	SELECT COUNT(*) as first_count
	FROM (
		SELECT date, MIN(start_time) as first_time
		FROM screentime_sessions
		WHERE CAST(substr(start_time, 1, 2) AS INTEGER) >= 6
			AND CAST(substr(start_time, 1, 2) AS INTEGER) <= 10
	`
	args := []interface{}{}

	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	query += `
		GROUP BY date
	) morning_sessions
	JOIN screentime_sessions s ON morning_sessions.date = s.date AND morning_sessions.first_time = s.start_time
	WHERE s.package_id = ?
	`
	args = append(args, packageID)

	var firstCount int
	if err := h.db.QueryRow(query, args...).Scan(&firstCount); err != nil {
		return false
	}

	return firstCount > 5 // If it's the first app more than 5 times
}

// calculateDependencyScore calculates dependency score (0-100)
func (h *Handler) calculateDependencyScore(frequency, percentage float64, streak int, isFirstMorning bool) int {
	score := 0.0

	// Frequency component (0-30 points)
	if frequency >= 10 {
		score += 30
	} else {
		score += frequency * 3
	}

	// Percentage component (0-30 points)
	if percentage >= 30 {
		score += 30
	} else {
		score += percentage
	}

	// Streak component (0-30 points)
	if streak >= 30 {
		score += 30
	} else {
		score += float64(streak)
	}

	// First morning app bonus (0-10 points)
	if isFirstMorning {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return int(score)
}

// determineDependencyType determines the type of dependency
func (h *Handler) determineDependencyType(category string, frequency float64, firstOpenTime string) string {
	switch category {
	case "Social", "Communication":
		return "social"
	case "Productivity", "Business", "Tools":
		return "work"
	case "Entertainment", "Games", "Video":
		return "entertainment"
	default:
		// Determine by usage pattern
		if frequency > 10 {
			return "social"
		}
		return "entertainment"
	}
}

// generateDependencyInsights generates insights for app dependency
func (h *Handler) generateDependencyInsights(dep AppDependency) []string {
	insights := []string{}

	if dep.DependencyScore >= 80 {
		insights = append(insights, fmt.Sprintf("对%s高度依赖，建议适当减少使用时间", dep.AppName))
	} else if dep.DependencyScore >= 60 {
		insights = append(insights, fmt.Sprintf("对%s中度依赖，使用频率较高", dep.AppName))
	}

	if dep.IsFirstMorningApp {
		insights = append(insights, "这是你早晨最常打开的应用之一")
	}

	if dep.ConsecutiveDays >= 30 {
		insights = append(insights, fmt.Sprintf("已连续使用%d天，形成了稳定的使用习惯", dep.ConsecutiveDays))
	}

	if dep.UsagePercentage >= 20 {
		insights = append(insights, fmt.Sprintf("占总使用时间的%.1f%%，是主要时间消耗应用", dep.UsagePercentage))
	}

	return insights
}

// WeekdayWeekendComparison represents weekday vs weekend comparison
type WeekdayWeekendComparison struct {
	Weekday struct {
		TotalDuration     int64            `json:"totalDuration"`
		AvgDailyDuration  float64          `json:"avgDailyDuration"`
		TopApps           []string         `json:"topApps"`
		CategoryDistribution map[string]int64 `json:"categoryDistribution"`
		HourlyDistribution []HourlyStat   `json:"hourlyDistribution"`
		DevicePreference  string           `json:"devicePreference"`
	} `json:"weekday"`
	Weekend struct {
		TotalDuration     int64            `json:"totalDuration"`
		AvgDailyDuration  float64          `json:"avgDailyDuration"`
		TopApps           []string         `json:"topApps"`
		CategoryDistribution map[string]int64 `json:"categoryDistribution"`
		HourlyDistribution []HourlyStat   `json:"hourlyDistribution"`
		DevicePreference  string           `json:"devicePreference"`
	} `json:"weekend"`
	Insights []string `json:"insights"`
}

// GetWeekdayWeekendComparison compares weekday vs weekend usage
// GET /api/v1/screentime/analysis/weekday-weekend?start=20240101&end=20241231
func (h *Handler) GetWeekdayWeekendComparison(c *gin.Context) {
	logger.Info("Weekday vs weekend comparison requested", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")

	var comparison WeekdayWeekendComparison
	comparison.Weekday.CategoryDistribution = make(map[string]int64)
	comparison.Weekend.CategoryDistribution = make(map[string]int64)

	// Get weekday data
	h.getWeekdayData(&comparison.Weekday, start, end, false)

	// Get weekend data
	h.getWeekdayData(&comparison.Weekend, start, end, true)

	// Generate insights
	comparison.Insights = h.generateWeekdayWeekendInsights(comparison)

	c.JSON(http.StatusOK, comparison)
}

// getWeekdayData gets usage data for weekday or weekend
func (h *Handler) getWeekdayData(data interface{}, start, end string, isWeekend bool) {
	// Determine day of week filter
	var dayFilter string
	if isWeekend {
		dayFilter = "IN ('Saturday', 'Sunday')"
	} else {
		dayFilter = "NOT IN ('Saturday', 'Sunday')"
	}

	// Get total duration and average
	query := fmt.Sprintf(`
	SELECT
		SUM(duration_ms) as total_duration,
		COUNT(DISTINCT date) as day_count
	FROM screentime_daily
	WHERE package_id != 'ALL'
		AND strftime('%%w', substr(date, 1, 4) || '-' || substr(date, 5, 2) || '-' || substr(date, 7, 2)) %s
	`, dayFilter)

	args := []interface{}{}
	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	var totalDuration sql.NullInt64
	var dayCount sql.NullInt64
	h.db.QueryRow(query, args...).Scan(&totalDuration, &dayCount)

	// Type assertion to set values
	switch v := data.(type) {
	case *struct {
		TotalDuration        int64
		AvgDailyDuration     float64
		TopApps              []string
		CategoryDistribution map[string]int64
		HourlyDistribution   []HourlyStat
		DevicePreference     string
	}:
		v.TotalDuration = totalDuration.Int64
		if dayCount.Int64 > 0 {
			v.AvgDailyDuration = float64(totalDuration.Int64) / float64(dayCount.Int64)
		}

		// Get top apps
		v.TopApps = h.getTopAppsForPeriod(start, end, isWeekend, 5)

		// Get category distribution
		v.CategoryDistribution = h.getCategoryDistribution(start, end, isWeekend)

		// Get hourly distribution
		v.HourlyDistribution = h.getHourlyDistribution(start, end, isWeekend)
	}
}

// getTopAppsForPeriod gets top apps for weekday or weekend
func (h *Handler) getTopAppsForPeriod(start, end string, isWeekend bool, limit int) []string {
	var dayFilter string
	if isWeekend {
		dayFilter = "IN ('6', '0')" // Saturday=6, Sunday=0 in strftime %w
	} else {
		dayFilter = "NOT IN ('6', '0')"
	}

	query := fmt.Sprintf(`
	SELECT a.app_name
	FROM screentime_daily d
	JOIN screentime_apps a ON d.package_id = a.package_id
	WHERE d.package_id != 'ALL'
		AND strftime('%%w', substr(d.date, 1, 4) || '-' || substr(d.date, 5, 2) || '-' || substr(d.date, 7, 2)) %s
	`, dayFilter)

	args := []interface{}{}
	if start != "" {
		query += " AND d.date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND d.date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY a.app_name ORDER BY SUM(d.duration_ms) DESC LIMIT ?"
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var apps []string
	for rows.Next() {
		var appName string
		if rows.Scan(&appName) == nil {
			apps = append(apps, appName)
		}
	}

	return apps
}

// getCategoryDistribution gets category distribution for weekday or weekend
func (h *Handler) getCategoryDistribution(start, end string, isWeekend bool) map[string]int64 {
	var dayFilter string
	if isWeekend {
		dayFilter = "IN ('6', '0')"
	} else {
		dayFilter = "NOT IN ('6', '0')"
	}

	query := fmt.Sprintf(`
	SELECT a.category, SUM(d.duration_ms) as total_duration
	FROM screentime_daily d
	JOIN screentime_apps a ON d.package_id = a.package_id
	WHERE d.package_id != 'ALL'
		AND strftime('%%w', substr(d.date, 1, 4) || '-' || substr(d.date, 5, 2) || '-' || substr(d.date, 7, 2)) %s
	`, dayFilter)

	args := []interface{}{}
	if start != "" {
		query += " AND d.date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND d.date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY a.category"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return map[string]int64{}
	}
	defer rows.Close()

	distribution := make(map[string]int64)
	for rows.Next() {
		var category string
		var duration int64
		if rows.Scan(&category, &duration) == nil {
			distribution[category] = duration
		}
	}

	return distribution
}

// getHourlyDistribution gets hourly distribution for weekday or weekend
func (h *Handler) getHourlyDistribution(start, end string, isWeekend bool) []HourlyStat {
	var dayFilter string
	if isWeekend {
		dayFilter = "IN ('6', '0')"
	} else {
		dayFilter = "NOT IN ('6', '0')"
	}

	query := fmt.Sprintf(`
	SELECT
		CAST(substr(start_time, 1, 2) AS INTEGER) as hour,
		COUNT(*) as session_count
	FROM screentime_sessions
	WHERE strftime('%%w', substr(date, 1, 4) || '-' || substr(date, 5, 2) || '-' || substr(date, 7, 2)) %s
	`, dayFilter)

	args := []interface{}{}
	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY hour ORDER BY hour"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return []HourlyStat{}
	}
	defer rows.Close()

	hourlyStats := make([]HourlyStat, 24)
	for i := 0; i < 24; i++ {
		hourlyStats[i].Hour = i
	}

	for rows.Next() {
		var hour int
		var count int
		if rows.Scan(&hour, &count) == nil && hour >= 0 && hour < 24 {
			hourlyStats[hour].LaunchCount = count
		}
	}

	return hourlyStats
}

// generateWeekdayWeekendInsights generates insights for weekday vs weekend comparison
func (h *Handler) generateWeekdayWeekendInsights(comparison WeekdayWeekendComparison) []string {
	insights := []string{}

	// Compare total duration
	if comparison.Weekend.AvgDailyDuration > comparison.Weekday.AvgDailyDuration*1.2 {
		insights = append(insights, "周末平均使用时间明显高于工作日，建议适当控制周末屏幕时间")
	} else if comparison.Weekday.AvgDailyDuration > comparison.Weekend.AvgDailyDuration*1.2 {
		insights = append(insights, "工作日平均使用时间高于周末，可能工作压力较大")
	}

	// Compare category distribution
	weekdayEntertainment := comparison.Weekday.CategoryDistribution["Entertainment"] + comparison.Weekday.CategoryDistribution["Games"]
	weekendEntertainment := comparison.Weekend.CategoryDistribution["Entertainment"] + comparison.Weekend.CategoryDistribution["Games"]

	if weekendEntertainment > weekdayEntertainment*2 {
		insights = append(insights, "周末娱乐应用使用时间显著增加，这是正常的放松模式")
	}

	// Compare productivity
	weekdayWork := comparison.Weekday.CategoryDistribution["Productivity"] + comparison.Weekday.CategoryDistribution["Business"]
	weekendWork := comparison.Weekend.CategoryDistribution["Productivity"] + comparison.Weekend.CategoryDistribution["Business"]

	if weekendWork > weekdayWork*0.5 {
		insights = append(insights, "周末仍有较多工作相关应用使用，建议注意工作生活平衡")
	}

	return insights
}

// ProductivityEntertainmentTrend represents productivity vs entertainment ratio trends
type ProductivityEntertainmentTrend struct {
	Weekly    []TrendPoint `json:"weekly"`
	Monthly   []TrendPoint `json:"monthly"`
	Quarterly []TrendPoint `json:"quarterly"`
	Summary   struct {
		AvgProductivityRatio float64 `json:"avgProductivityRatio"`
		AvgEntertainmentRatio float64 `json:"avgEntertainmentRatio"`
		TrendDirection       string  `json:"trendDirection"` // "improving", "declining", "stable"
	} `json:"summary"`
}

// GetProductivityEntertainmentTrend gets productivity vs entertainment ratio trends
// GET /api/v1/screentime/analysis/productivity-entertainment-trend?start=20240101&end=20241231
func (h *Handler) GetProductivityEntertainmentTrend(c *gin.Context) {
	logger.Info("Productivity vs entertainment trend requested", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")

	var trend ProductivityEntertainmentTrend

	// Get weekly trends
	trend.Weekly = h.getProductivityTrends("weekly", start, end)

	// Get monthly trends
	trend.Monthly = h.getProductivityTrends("monthly", start, end)

	// Get quarterly trends
	trend.Quarterly = h.getProductivityTrends("quarterly", start, end)

	// Calculate summary
	h.calculateProductivitySummary(&trend)

	c.JSON(http.StatusOK, trend)
}

// getProductivityTrends gets productivity trends by granularity
func (h *Handler) getProductivityTrends(granularity, start, end string) []TrendPoint {
	var periodFormat string
	switch granularity {
	case "weekly":
		periodFormat = "strftime('%Y%W', substr(d.date, 1, 4) || '-' || substr(d.date, 5, 2) || '-' || substr(d.date, 7, 2))"
	case "monthly":
		periodFormat = "substr(d.date, 1, 6)"
	case "quarterly":
		periodFormat = "substr(d.date, 1, 4) || 'Q' || CAST((CAST(substr(d.date, 5, 2) AS INTEGER) - 1) / 3 + 1 AS TEXT)"
	default:
		periodFormat = "d.date"
	}

	query := fmt.Sprintf(`
	SELECT
		%s as period,
		SUM(CASE WHEN a.category IN ('Productivity', 'Business', 'Tools') THEN d.duration_ms ELSE 0 END) as productivity_duration,
		SUM(CASE WHEN a.category IN ('Entertainment', 'Games', 'Video', 'Social') THEN d.duration_ms ELSE 0 END) as entertainment_duration
	FROM screentime_daily d
	JOIN screentime_apps a ON d.package_id = a.package_id
	WHERE d.package_id != 'ALL'
	`, periodFormat)

	args := []interface{}{}
	if start != "" {
		query += " AND d.date >= ?"
		args = append(args, start)
	}
	if end != "" {
		query += " AND d.date <= ?"
		args = append(args, end)
	}

	query += " GROUP BY period ORDER BY period"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		logger.Error("Failed to get productivity trends", err, nil)
		return []TrendPoint{}
	}
	defer rows.Close()

	var trends []TrendPoint
	for rows.Next() {
		var period string
		var productivityDuration, entertainmentDuration int64
		if err := rows.Scan(&period, &productivityDuration, &entertainmentDuration); err == nil {
			total := productivityDuration + entertainmentDuration
			var ratio float64
			if total > 0 {
				ratio = float64(productivityDuration) / float64(total) * 100
			}

			trends = append(trends, TrendPoint{
				Date:  period,
				Value: ratio,
				Label: fmt.Sprintf("%.1f%%", ratio),
			})
		}
	}

	return trends
}

// calculateProductivitySummary calculates productivity summary
func (h *Handler) calculateProductivitySummary(trend *ProductivityEntertainmentTrend) {
	if len(trend.Monthly) == 0 {
		return
	}

	// Calculate average ratios
	var totalProductivity, totalEntertainment float64
	for _, point := range trend.Monthly {
		totalProductivity += point.Value
		totalEntertainment += (100 - point.Value)
	}

	trend.Summary.AvgProductivityRatio = totalProductivity / float64(len(trend.Monthly))
	trend.Summary.AvgEntertainmentRatio = totalEntertainment / float64(len(trend.Monthly))

	// Determine trend direction
	if len(trend.Monthly) >= 3 {
		recent := trend.Monthly[len(trend.Monthly)-3:]
		var recentAvg float64
		for _, point := range recent {
			recentAvg += point.Value
		}
		recentAvg /= float64(len(recent))

		if recentAvg > trend.Summary.AvgProductivityRatio+5 {
			trend.Summary.TrendDirection = "improving"
		} else if recentAvg < trend.Summary.AvgProductivityRatio-5 {
			trend.Summary.TrendDirection = "declining"
		} else {
			trend.Summary.TrendDirection = "stable"
		}
	}
}
