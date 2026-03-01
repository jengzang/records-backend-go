package analysis

import (
	"database/sql"
	"fmt"
)

// SeasonalTrends 季节性健康趋势分析结果
type SeasonalTrends struct {
	SeasonalStats    []SeasonalStat    `json:"seasonal_stats"`    // 季节统计
	MonthlyStats     []MonthlyStat     `json:"monthly_stats"`     // 月度统计
	YearlyComparison []YearlyComparison `json:"yearly_comparison"` // 年度对比
	SeasonalPatterns SeasonalPatterns  `json:"seasonal_patterns"` // 季节模式
	AnnualReport     AnnualReport      `json:"annual_report"`     // 年度报告
	Insights         []string          `json:"insights"`          // 洞察
}

// SeasonalStat 季节统计
type SeasonalStat struct {
	Season          string  `json:"season"`           // 季节: spring/summer/autumn/winter
	SeasonName      string  `json:"season_name"`      // 季节名称
	AvgHeartRate    float64 `json:"avg_heart_rate"`   // 平均心率
	AvgRestingHR    float64 `json:"avg_resting_hr"`   // 平均静息心率
	AvgSteps        float64 `json:"avg_steps"`        // 平均步数
	AvgDistance     float64 `json:"avg_distance"`     // 平均距离(km)
	AvgCalories     float64 `json:"avg_calories"`     // 平均卡路里
	ActiveDays      int     `json:"active_days"`      // 活跃天数
	TotalRecords    int     `json:"total_records"`    // 总记录数
}

// MonthlyStat 月度统计
type MonthlyStat struct {
	Month        int     `json:"month"`         // 月份(1-12)
	MonthName    string  `json:"month_name"`    // 月份名称
	Year         int     `json:"year"`          // 年份
	AvgHeartRate float64 `json:"avg_heart_rate"` // 平均心率
	AvgSteps     float64 `json:"avg_steps"`     // 平均步数
	ActiveDays   int     `json:"active_days"`   // 活跃天数
}

// YearlyComparison 年度对比
type YearlyComparison struct {
	Year         int     `json:"year"`          // 年份
	AvgHeartRate float64 `json:"avg_heart_rate"` // 平均心率
	AvgSteps     float64 `json:"avg_steps"`     // 平均步数
	TotalRecords int     `json:"total_records"` // 总记录数
	ActiveDays   int     `json:"active_days"`   // 活跃天数
}

// SeasonalPatterns 季节模式
type SeasonalPatterns struct {
	MostActiveSeasons    string  `json:"most_active_season"`     // 最活跃季节
	LeastActiveSeasons   string  `json:"least_active_season"`    // 最不活跃季节
	HighestHRSeason      string  `json:"highest_hr_season"`      // 心率最高季节
	LowestHRSeason       string  `json:"lowest_hr_season"`       // 心率最低季节
	SeasonalVariation    float64 `json:"seasonal_variation"`     // 季节变异系数
	ConsistencyScore     float64 `json:"consistency_score"`      // 一致性评分(0-100)
}

// AnnualReport 年度报告
type AnnualReport struct {
	TotalYears       int     `json:"total_years"`        // 总年数
	TotalRecords     int     `json:"total_records"`      // 总记录数
	AvgHeartRate     float64 `json:"avg_heart_rate"`     // 平均心率
	AvgSteps         float64 `json:"avg_steps"`          // 平均步数
	TotalDistance    float64 `json:"total_distance"`     // 总距离(km)
	TotalCalories    float64 `json:"total_calories"`     // 总卡路里
	HealthTrend      string  `json:"health_trend"`       // 健康趋势: improving/stable/declining
	BestMonth        string  `json:"best_month"`         // 最佳月份
	WorstMonth       string  `json:"worst_month"`        // 最差月份
}

// GetSeasonalTrends 获取季节性健康趋势分析
func GetSeasonalTrends(db *sql.DB) (*SeasonalTrends, error) {
	trends := &SeasonalTrends{}

	// 1. 计算季节统计
	seasonalStats, err := calculateSeasonalStats(db)
	if err != nil {
		return nil, err
	}
	trends.SeasonalStats = seasonalStats

	// 2. 计算月度统计
	monthlyStats, err := calculateMonthlyStats(db)
	if err != nil {
		return nil, err
	}
	trends.MonthlyStats = monthlyStats

	// 3. 计算年度对比
	yearlyComparison, err := calculateYearlyComparison(db)
	if err != nil {
		return nil, err
	}
	trends.YearlyComparison = yearlyComparison

	// 4. 识别季节模式
	trends.SeasonalPatterns = identifySeasonalPatterns(seasonalStats)

	// 5. 生成年度报告
	trends.AnnualReport = generateAnnualReport(db, monthlyStats, yearlyComparison)

	// 6. 生成洞察
	trends.Insights = generateSeasonalInsights(trends)

	return trends, nil
}

// calculateSeasonalStats 计算季节统计
func calculateSeasonalStats(db *sql.DB) ([]SeasonalStat, error) {
	query := `
	SELECT
		CASE
			WHEN CAST(strftime('%m', date) AS INTEGER) IN (3, 4, 5) THEN 'spring'
			WHEN CAST(strftime('%m', date) AS INTEGER) IN (6, 7, 8) THEN 'summer'
			WHEN CAST(strftime('%m', date) AS INTEGER) IN (9, 10, 11) THEN 'autumn'
			ELSE 'winter'
		END as season,
		AVG(CASE WHEN type = 'HeartRate' THEN value END) as avg_heart_rate,
		AVG(CASE WHEN type = 'RestingHeartRate' THEN value END) as avg_resting_hr,
		AVG(CASE WHEN type = 'StepCount' THEN value END) as avg_steps,
		AVG(CASE WHEN type = 'DistanceWalkingRunning' THEN value END) as avg_distance,
		AVG(CASE WHEN type = 'ActiveEnergyBurned' THEN value END) as avg_calories,
		COUNT(DISTINCT date) as active_days,
		COUNT(*) as total_records
	FROM health_records
	GROUP BY season
	ORDER BY
		CASE season
			WHEN 'spring' THEN 1
			WHEN 'summer' THEN 2
			WHEN 'autumn' THEN 3
			WHEN 'winter' THEN 4
		END
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []SeasonalStat
	for rows.Next() {
		var stat SeasonalStat
		var avgHR, avgRestingHR, avgSteps, avgDistance, avgCalories sql.NullFloat64

		err := rows.Scan(
			&stat.Season,
			&avgHR,
			&avgRestingHR,
			&avgSteps,
			&avgDistance,
			&avgCalories,
			&stat.ActiveDays,
			&stat.TotalRecords,
		)
		if err != nil {
			continue
		}

		stat.AvgHeartRate = avgHR.Float64
		stat.AvgRestingHR = avgRestingHR.Float64
		stat.AvgSteps = avgSteps.Float64
		stat.AvgDistance = avgDistance.Float64
		stat.AvgCalories = avgCalories.Float64
		stat.SeasonName = getSeasonName(stat.Season)

		stats = append(stats, stat)
	}

	return stats, nil
}

// calculateMonthlyStats 计算月度统计
func calculateMonthlyStats(db *sql.DB) ([]MonthlyStat, error) {
	query := `
	SELECT
		CAST(strftime('%m', date) AS INTEGER) as month,
		CAST(strftime('%Y', date) AS INTEGER) as year,
		AVG(CASE WHEN type = 'HeartRate' THEN value END) as avg_heart_rate,
		AVG(CASE WHEN type = 'StepCount' THEN value END) as avg_steps,
		COUNT(DISTINCT date) as active_days
	FROM health_records
	GROUP BY year, month
	ORDER BY year DESC, month DESC
	LIMIT 24
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []MonthlyStat
	for rows.Next() {
		var stat MonthlyStat
		var avgHR, avgSteps sql.NullFloat64

		err := rows.Scan(
			&stat.Month,
			&stat.Year,
			&avgHR,
			&avgSteps,
			&stat.ActiveDays,
		)
		if err != nil {
			continue
		}

		stat.AvgHeartRate = avgHR.Float64
		stat.AvgSteps = avgSteps.Float64
		stat.MonthName = getMonthName(stat.Month)

		stats = append(stats, stat)
	}

	return stats, nil
}

// calculateYearlyComparison 计算年度对比
func calculateYearlyComparison(db *sql.DB) ([]YearlyComparison, error) {
	query := `
	SELECT
		CAST(strftime('%Y', date) AS INTEGER) as year,
		AVG(CASE WHEN type = 'HeartRate' THEN value END) as avg_heart_rate,
		AVG(CASE WHEN type = 'StepCount' THEN value END) as avg_steps,
		COUNT(*) as total_records,
		COUNT(DISTINCT date) as active_days
	FROM health_records
	GROUP BY year
	ORDER BY year DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comparisons []YearlyComparison
	for rows.Next() {
		var comp YearlyComparison
		var avgHR, avgSteps sql.NullFloat64

		err := rows.Scan(
			&comp.Year,
			&avgHR,
			&avgSteps,
			&comp.TotalRecords,
			&comp.ActiveDays,
		)
		if err != nil {
			continue
		}

		comp.AvgHeartRate = avgHR.Float64
		comp.AvgSteps = avgSteps.Float64

		comparisons = append(comparisons, comp)
	}

	return comparisons, nil
}

// identifySeasonalPatterns 识别季节模式
func identifySeasonalPatterns(stats []SeasonalStat) SeasonalPatterns {
	patterns := SeasonalPatterns{}

	if len(stats) == 0 {
		return patterns
	}

	// 找出最活跃和最不活跃的季节
	maxSteps := 0.0
	minSteps := 999999999.0
	maxHR := 0.0
	minHR := 999999999.0

	for _, stat := range stats {
		if stat.AvgSteps > maxSteps {
			maxSteps = stat.AvgSteps
			patterns.MostActiveSeasons = stat.SeasonName
		}
		if stat.AvgSteps < minSteps && stat.AvgSteps > 0 {
			minSteps = stat.AvgSteps
			patterns.LeastActiveSeasons = stat.SeasonName
		}
		if stat.AvgHeartRate > maxHR {
			maxHR = stat.AvgHeartRate
			patterns.HighestHRSeason = stat.SeasonName
		}
		if stat.AvgHeartRate < minHR && stat.AvgHeartRate > 0 {
			minHR = stat.AvgHeartRate
			patterns.LowestHRSeason = stat.SeasonName
		}
	}

	// 计算季节变异系数
	if maxSteps > 0 && minSteps > 0 {
		patterns.SeasonalVariation = (maxSteps - minSteps) / maxSteps * 100
	}

	// 计算一致性评分 (变异越小，一致性越高)
	patterns.ConsistencyScore = 100 - patterns.SeasonalVariation
	if patterns.ConsistencyScore < 0 {
		patterns.ConsistencyScore = 0
	}

	return patterns
}

// generateAnnualReport 生成年度报告
func generateAnnualReport(db *sql.DB, monthlyStats []MonthlyStat, yearlyComparison []YearlyComparison) AnnualReport {
	report := AnnualReport{}

	// 基础统计
	var totalRecords int
	var avgHR, avgSteps, totalDistance, totalCalories sql.NullFloat64

	db.QueryRow(`
		SELECT
			COUNT(*) as total_records,
			AVG(CASE WHEN type = 'HeartRate' THEN value END) as avg_heart_rate,
			AVG(CASE WHEN type = 'StepCount' THEN value END) as avg_steps,
			SUM(CASE WHEN type = 'DistanceWalkingRunning' THEN value END) as total_distance,
			SUM(CASE WHEN type = 'ActiveEnergyBurned' THEN value END) as total_calories
		FROM health_records
	`).Scan(&totalRecords, &avgHR, &avgSteps, &totalDistance, &totalCalories)

	report.TotalRecords = totalRecords
	report.AvgHeartRate = avgHR.Float64
	report.AvgSteps = avgSteps.Float64
	report.TotalDistance = totalDistance.Float64
	report.TotalCalories = totalCalories.Float64
	report.TotalYears = len(yearlyComparison)

	// 健康趋势判断
	if len(yearlyComparison) >= 2 {
		recent := yearlyComparison[0]
		previous := yearlyComparison[1]

		if recent.AvgSteps > previous.AvgSteps*1.1 {
			report.HealthTrend = "improving"
		} else if recent.AvgSteps < previous.AvgSteps*0.9 {
			report.HealthTrend = "declining"
		} else {
			report.HealthTrend = "stable"
		}
	} else {
		report.HealthTrend = "stable"
	}

	// 找出最佳和最差月份
	if len(monthlyStats) > 0 {
		maxSteps := 0.0
		minSteps := 999999999.0

		for _, stat := range monthlyStats {
			if stat.AvgSteps > maxSteps {
				maxSteps = stat.AvgSteps
				report.BestMonth = fmt.Sprintf("%d年%s", stat.Year, stat.MonthName)
			}
			if stat.AvgSteps < minSteps && stat.AvgSteps > 0 {
				minSteps = stat.AvgSteps
				report.WorstMonth = fmt.Sprintf("%d年%s", stat.Year, stat.MonthName)
			}
		}
	}

	return report
}

// generateSeasonalInsights 生成季节洞察
func generateSeasonalInsights(trends *SeasonalTrends) []string {
	insights := []string{}

	// 季节活动模式
	if trends.SeasonalPatterns.MostActiveSeasons != "" {
		insights = append(insights,
			fmt.Sprintf("你在%s最活跃，平均步数最高", trends.SeasonalPatterns.MostActiveSeasons))
	}

	if trends.SeasonalPatterns.LeastActiveSeasons != "" {
		insights = append(insights,
			fmt.Sprintf("你在%s活动量最少，建议增加户外活动", trends.SeasonalPatterns.LeastActiveSeasons))
	}

	// 一致性评分
	if trends.SeasonalPatterns.ConsistencyScore >= 80 {
		insights = append(insights, "你的活动量在各季节保持高度一致，这是良好的健康习惯")
	} else if trends.SeasonalPatterns.ConsistencyScore < 50 {
		insights = append(insights, "你的活动量季节性波动较大，建议在淡季保持运动习惯")
	}

	// 健康趋势
	switch trends.AnnualReport.HealthTrend {
	case "improving":
		insights = append(insights, "你的健康状况呈上升趋势，继续保持！")
	case "declining":
		insights = append(insights, "你的活动量有所下降，建议增加运动频率")
	case "stable":
		insights = append(insights, "你的健康状况保持稳定")
	}

	// 年度成就
	if trends.AnnualReport.TotalDistance > 1000 {
		insights = append(insights,
			fmt.Sprintf("年度总里程达到%.0f公里，相当于从北京到上海的距离！", trends.AnnualReport.TotalDistance))
	}

	if trends.AnnualReport.TotalCalories > 100000 {
		insights = append(insights,
			fmt.Sprintf("年度消耗%.0f千卡，相当于减掉%.1f公斤体重！",
				trends.AnnualReport.TotalCalories,
				trends.AnnualReport.TotalCalories/7700))
	}

	return insights
}

// 辅助函数
func getSeasonName(season string) string {
	names := map[string]string{
		"spring": "春季",
		"summer": "夏季",
		"autumn": "秋季",
		"winter": "冬季",
	}
	return names[season]
}

func getMonthName(month int) string {
	names := []string{"", "1月", "2月", "3月", "4月", "5月", "6月",
		"7月", "8月", "9月", "10月", "11月", "12月"}
	if month >= 1 && month <= 12 {
		return names[month]
	}
	return ""
}
