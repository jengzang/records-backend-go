package screentime

import (
	"database/sql"
)

// LateNightUsageAnalysis 深夜使用习惯分析结果
type LateNightUsageAnalysis struct {
	OverallPattern      LateNightOverallPattern `json:"overall_pattern"`
	FrequencyAnalysis   FrequencyAnalysis       `json:"frequency_analysis"`
	AppUsageBreakdown   []LateNightApp          `json:"app_usage_breakdown"`
	TimeDistribution    []HourDistribution      `json:"time_distribution"`
	SleepImpact         SleepImpactAssessment   `json:"sleep_impact"`
	WeekdayComparison   WeekdayComparison       `json:"weekday_comparison"`
	Recommendations     []string                `json:"recommendations"`
}

// LateNightOverallPattern 深夜使用总体模式
type LateNightOverallPattern struct {
	TotalLateNightHours   float64 `json:"total_late_night_hours"`    // 总深夜使用时长(小时)
	AvgPerNight           float64 `json:"avg_per_night"`             // 平均每晚使用时长(小时)
	LateNightDays         int     `json:"late_night_days"`           // 深夜使用天数
	TotalDays             int     `json:"total_days"`                // 总天数
	LateNightFrequency    float64 `json:"late_night_frequency"`      // 深夜使用频率(%)
	MostActiveLateHour    int     `json:"most_active_late_hour"`     // 最活跃深夜时段
	AvgSessionDuration    float64 `json:"avg_session_duration"`      // 平均会话时长(分钟)
	LongestSession        float64 `json:"longest_session"`           // 最长会话(分钟)
}

// FrequencyAnalysis 频率分析
type FrequencyAnalysis struct {
	NeverUseLateNight    int     `json:"never_use_late_night"`     // 从不深夜使用天数
	OccasionalUse        int     `json:"occasional_use"`           // 偶尔使用(<30分钟)
	ModerateUse          int     `json:"moderate_use"`             // 中度使用(30-60分钟)
	HeavyUse             int     `json:"heavy_use"`                // 重度使用(>60分钟)
	OccasionalPercentage float64 `json:"occasional_percentage"`
	ModeratePercentage   float64 `json:"moderate_percentage"`
	HeavyPercentage      float64 `json:"heavy_percentage"`
}

// LateNightApp 深夜使用应用
type LateNightApp struct {
	AppName         string  `json:"app_name"`
	Category        string  `json:"category"`
	TotalDuration   float64 `json:"total_duration"`    // 小时
	UsageCount      int     `json:"usage_count"`       // 使用次数
	AvgDuration     float64 `json:"avg_duration"`      // 平均时长(分钟)
	Percentage      float64 `json:"percentage"`        // 占比
	ImpactLevel     string  `json:"impact_level"`      // 影响等级: high/medium/low
}

// HourDistribution 小时分布
type HourDistribution struct {
	Hour            int     `json:"hour"`
	AvgDuration     float64 `json:"avg_duration"`      // 分钟
	UsageCount      int     `json:"usage_count"`
	MostUsedApp     string  `json:"most_used_app"`
}

// SleepImpactAssessment 睡眠影响评估
type SleepImpactAssessment struct {
	ImpactScore         float64  `json:"impact_score"`          // 影响评分(0-100, 越高影响越大)
	EstimatedSleepLoss  float64  `json:"estimated_sleep_loss"`  // 估算睡眠损失(小时/周)
	BlueLightExposure   float64  `json:"blue_light_exposure"`   // 蓝光暴露时长(小时)
	RiskLevel           string   `json:"risk_level"`            // 风险等级: low/medium/high
	ImpactFactors       []string `json:"impact_factors"`        // 影响因素
}

// WeekdayComparison 工作日周末对比
type WeekdayComparison struct {
	WeekdayAvg  float64 `json:"weekday_avg"`   // 工作日平均(分钟)
	WeekendAvg  float64 `json:"weekend_avg"`   // 周末平均(分钟)
	Difference  float64 `json:"difference"`    // 差异(分钟)
	Pattern     string  `json:"pattern"`       // 模式: weekday_heavy/weekend_heavy/balanced
}

// GetLateNightUsageAnalysis 获取深夜使用习惯分析
func (h *Handler) GetLateNightUsageAnalysis() (*LateNightUsageAnalysis, error) {
	analysis := &LateNightUsageAnalysis{}

	// 1. 总体模式分析
	overall, err := h.analyzeLateNightOverallPattern()
	if err != nil {
		return nil, err
	}
	analysis.OverallPattern = *overall

	// 2. 频率分析
	frequency, err := h.analyzeLateNightFrequency()
	if err != nil {
		return nil, err
	}
	analysis.FrequencyAnalysis = *frequency

	// 3. 应用使用分解
	apps, err := h.analyzeLateNightApps()
	if err != nil {
		return nil, err
	}
	analysis.AppUsageBreakdown = apps

	// 4. 时间分布
	timeDistribution, err := h.analyzeLateNightTimeDistribution()
	if err != nil {
		return nil, err
	}
	analysis.TimeDistribution = timeDistribution

	// 5. 睡眠影响评估
	sleepImpact, err := h.assessSleepImpact(overall, apps)
	if err != nil {
		return nil, err
	}
	analysis.SleepImpact = *sleepImpact

	// 6. 工作日周末对比
	weekdayComp, err := h.compareLateNightWeekdayWeekend()
	if err != nil {
		return nil, err
	}
	analysis.WeekdayComparison = *weekdayComp

	// 7. 生成建议
	analysis.Recommendations = generateLateNightRecommendations(analysis)

	return analysis, nil
}

// analyzeLateNightOverallPattern 分析深夜使用总体模式
func (h *Handler) analyzeLateNightOverallPattern() (*LateNightOverallPattern, error) {
	// 深夜定义: 22:00-02:00 (晚上10点到凌晨2点)
	query := `
		SELECT
			date,
			CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) as hour,
			SUM(duration_ms) as total_duration
		FROM app_usage
		WHERE (hour >= 22 OR hour < 2)
		GROUP BY date, hour
		ORDER BY date, hour
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyUsage := make(map[string]float64)
	hourlyUsage := make(map[int]float64)
	var longestSession float64

	for rows.Next() {
		var date string
		var hour int
		var duration int64

		if err := rows.Scan(&date, &hour, &duration); err != nil {
			continue
		}

		durationHours := float64(duration) / (1000 * 60 * 60)
		dailyUsage[date] += durationHours
		hourlyUsage[hour] += durationHours

		sessionMinutes := float64(duration) / (1000 * 60)
		if sessionMinutes > longestSession {
			longestSession = sessionMinutes
		}
	}

	// 计算统计数据
	totalLateNightHours := 0.0
	lateNightDays := 0
	for _, hours := range dailyUsage {
		totalLateNightHours += hours
		if hours > 0 {
			lateNightDays++
		}
	}

	// 获取总天数
	totalDaysQuery := `SELECT COUNT(DISTINCT date) FROM app_usage`
	var totalDays int
	h.db.QueryRow(totalDaysQuery).Scan(&totalDays)
	if totalDays == 0 {
		totalDays = 1
	}

	avgPerNight := 0.0
	if lateNightDays > 0 {
		avgPerNight = totalLateNightHours / float64(lateNightDays)
	}

	frequency := float64(lateNightDays) / float64(totalDays) * 100

	// 找出最活跃的深夜时段
	mostActiveHour := 22
	maxUsage := 0.0
	for hour, usage := range hourlyUsage {
		if usage > maxUsage {
			maxUsage = usage
			mostActiveHour = hour
		}
	}

	avgSessionDuration := 0.0
	if lateNightDays > 0 {
		avgSessionDuration = (totalLateNightHours * 60) / float64(lateNightDays)
	}

	return &LateNightOverallPattern{
		TotalLateNightHours: totalLateNightHours,
		AvgPerNight:         avgPerNight,
		LateNightDays:       lateNightDays,
		TotalDays:           totalDays,
		LateNightFrequency:  frequency,
		MostActiveLateHour:  mostActiveHour,
		AvgSessionDuration:  avgSessionDuration,
		LongestSession:      longestSession,
	}, nil
}

// analyzeLateNightFrequency 分析深夜使用频率
func (h *Handler) analyzeLateNightFrequency() (*FrequencyAnalysis, error) {
	query := `
		SELECT
			date,
			SUM(duration_ms) / (1000.0 * 60) as total_minutes
		FROM app_usage
		WHERE (CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) >= 22
		   OR CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) < 2)
		GROUP BY date
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var neverUse, occasional, moderate, heavy int

	for rows.Next() {
		var date string
		var minutes float64

		if err := rows.Scan(&date, &minutes); err != nil {
			continue
		}

		if minutes == 0 {
			neverUse++
		} else if minutes < 30 {
			occasional++
		} else if minutes < 60 {
			moderate++
		} else {
			heavy++
		}
	}

	total := neverUse + occasional + moderate + heavy
	if total == 0 {
		total = 1
	}

	return &FrequencyAnalysis{
		NeverUseLateNight:    neverUse,
		OccasionalUse:        occasional,
		ModerateUse:          moderate,
		HeavyUse:             heavy,
		OccasionalPercentage: float64(occasional) / float64(total) * 100,
		ModeratePercentage:   float64(moderate) / float64(total) * 100,
		HeavyPercentage:      float64(heavy) / float64(total) * 100,
	}, nil
}

// analyzeLateNightApps 分析深夜使用应用
func (h *Handler) analyzeLateNightApps() ([]LateNightApp, error) {
	query := `
		SELECT
			app_name,
			SUM(duration_ms) as total_duration,
			COUNT(*) as usage_count
		FROM app_usage
		WHERE (CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) >= 22
		   OR CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) < 2)
		GROUP BY app_name
		ORDER BY total_duration DESC
		LIMIT 20
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []LateNightApp
	var totalDuration float64

	for rows.Next() {
		var appName string
		var duration int64
		var count int

		if err := rows.Scan(&appName, &duration, &count); err != nil {
			continue
		}

		durationHours := float64(duration) / (1000 * 60 * 60)
		totalDuration += durationHours

		apps = append(apps, LateNightApp{
			AppName:       appName,
			TotalDuration: durationHours,
			UsageCount:    count,
			AvgDuration:   float64(duration) / float64(count) / (1000 * 60),
		})
	}

	// 计算百分比和影响等级
	mapper := NewAppCategoryMapper()
	for i := range apps {
		apps[i].Percentage = apps[i].TotalDuration / totalDuration * 100
		apps[i].Category = mapper.GetCategory(apps[i].AppName)
		apps[i].ImpactLevel = assessAppImpactLevel(apps[i].AppName, apps[i].Category)
	}

	return apps, nil
}

// analyzeLateNightTimeDistribution 分析深夜时间分布
func (h *Handler) analyzeLateNightTimeDistribution() ([]HourDistribution, error) {
	lateNightHours := []int{22, 23, 0, 1}
	var distribution []HourDistribution

	for _, hour := range lateNightHours {
		query := `
			SELECT
				COUNT(*) as usage_count,
				AVG(duration_ms) / (1000.0 * 60) as avg_duration,
				app_name
			FROM app_usage
			WHERE CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) = ?
			GROUP BY app_name
			ORDER BY SUM(duration_ms) DESC
			LIMIT 1
		`

		var count int
		var avgDuration float64
		var mostUsedApp string

		err := h.db.QueryRow(query, hour).Scan(&count, &avgDuration, &mostUsedApp)
		if err != nil && err != sql.ErrNoRows {
			continue
		}

		distribution = append(distribution, HourDistribution{
			Hour:        hour,
			AvgDuration: avgDuration,
			UsageCount:  count,
			MostUsedApp: mostUsedApp,
		})
	}

	return distribution, nil
}

// assessSleepImpact 评估睡眠影响
func (h *Handler) assessSleepImpact(overall *LateNightOverallPattern, apps []LateNightApp) (*SleepImpactAssessment, error) {
	// 计算影响评分 (0-100)
	impactScore := 0.0

	// 基于使用频率
	if overall.LateNightFrequency > 70 {
		impactScore += 40
	} else if overall.LateNightFrequency > 40 {
		impactScore += 25
	} else if overall.LateNightFrequency > 20 {
		impactScore += 10
	}

	// 基于平均使用时长
	if overall.AvgPerNight > 2 {
		impactScore += 40
	} else if overall.AvgPerNight > 1 {
		impactScore += 25
	} else if overall.AvgPerNight > 0.5 {
		impactScore += 15
	}

	// 基于应用类型
	highImpactApps := 0
	for _, app := range apps {
		if app.ImpactLevel == "high" {
			highImpactApps++
		}
	}
	if highImpactApps > 5 {
		impactScore += 20
	} else if highImpactApps > 2 {
		impactScore += 10
	}

	// 估算睡眠损失 (每周)
	estimatedSleepLoss := overall.AvgPerNight * 7 * 0.8 // 假设80%的深夜使用时间影响睡眠

	// 蓝光暴露
	blueLightExposure := overall.TotalLateNightHours

	// 风险等级
	riskLevel := "low"
	if impactScore > 70 {
		riskLevel = "high"
	} else if impactScore > 40 {
		riskLevel = "medium"
	}

	// 影响因素
	impactFactors := []string{}
	if overall.LateNightFrequency > 50 {
		impactFactors = append(impactFactors, "深夜使用频率过高")
	}
	if overall.AvgPerNight > 1 {
		impactFactors = append(impactFactors, "深夜使用时长过长")
	}
	if highImpactApps > 2 {
		impactFactors = append(impactFactors, "使用高刺激性应用")
	}
	if overall.MostActiveLateHour >= 0 && overall.MostActiveLateHour < 2 {
		impactFactors = append(impactFactors, "凌晨时段活跃")
	}

	return &SleepImpactAssessment{
		ImpactScore:        impactScore,
		EstimatedSleepLoss: estimatedSleepLoss,
		BlueLightExposure:  blueLightExposure,
		RiskLevel:          riskLevel,
		ImpactFactors:      impactFactors,
	}, nil
}

// compareLateNightWeekdayWeekend 对比工作日周末深夜使用
func (h *Handler) compareLateNightWeekdayWeekend() (*WeekdayComparison, error) {
	query := `
		SELECT
			CASE
				WHEN CAST(strftime('%w', date) AS INTEGER) IN (0, 6) THEN 'weekend'
				ELSE 'weekday'
			END as day_type,
			AVG(duration_ms) / (1000.0 * 60) as avg_minutes
		FROM (
			SELECT
				date,
				SUM(duration_ms) as duration_ms
			FROM app_usage
			WHERE (CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) >= 22
			   OR CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) < 2)
			GROUP BY date
		)
		GROUP BY day_type
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weekdayAvg, weekendAvg float64

	for rows.Next() {
		var dayType string
		var avgMinutes float64

		if err := rows.Scan(&dayType, &avgMinutes); err != nil {
			continue
		}

		if dayType == "weekday" {
			weekdayAvg = avgMinutes
		} else {
			weekendAvg = avgMinutes
		}
	}

	difference := weekendAvg - weekdayAvg
	pattern := "balanced"
	if difference > 20 {
		pattern = "weekend_heavy"
	} else if difference < -20 {
		pattern = "weekday_heavy"
	}

	return &WeekdayComparison{
		WeekdayAvg: weekdayAvg,
		WeekendAvg: weekendAvg,
		Difference: difference,
		Pattern:    pattern,
	}, nil
}

// assessAppImpactLevel 评估应用影响等级
func assessAppImpactLevel(appName, category string) string {
	// 高影响: 社交、游戏、视频
	highImpactCategories := []string{"social", "game", "video", "entertainment"}
	for _, cat := range highImpactCategories {
		if category == cat {
			return "high"
		}
	}

	// 中影响: 新闻、购物
	mediumImpactCategories := []string{"news", "shopping"}
	for _, cat := range mediumImpactCategories {
		if category == cat {
			return "medium"
		}
	}

	// 低影响: 工具、阅读
	return "low"
}

// generateLateNightRecommendations 生成深夜使用建议
func generateLateNightRecommendations(analysis *LateNightUsageAnalysis) []string {
	recommendations := []string{}

	// 基于风险等级
	if analysis.SleepImpact.RiskLevel == "high" {
		recommendations = append(recommendations,
			"深夜使用严重影响睡眠质量，建议立即调整使用习惯")
	}

	// 基于使用频率
	if analysis.OverallPattern.LateNightFrequency > 50 {
		recommendations = append(recommendations,
			"深夜使用频率过高，建议设置22:00后的应用使用限制")
	}

	// 基于使用时长
	if analysis.OverallPattern.AvgPerNight > 1 {
		recommendations = append(recommendations,
			"深夜平均使用时长超过1小时，建议逐步减少至30分钟以内")
	}

	// 基于应用类型
	highImpactCount := 0
	for _, app := range analysis.AppUsageBreakdown {
		if app.ImpactLevel == "high" {
			highImpactCount++
		}
	}
	if highImpactCount > 3 {
		recommendations = append(recommendations,
			"深夜使用过多高刺激性应用(社交/游戏/视频)，建议改用阅读类应用")
	}

	// 基于最活跃时段
	if analysis.OverallPattern.MostActiveLateHour >= 0 && analysis.OverallPattern.MostActiveLateHour < 2 {
		recommendations = append(recommendations,
			"凌晨时段仍在使用手机，严重影响睡眠，建议23:00前结束使用")
	}

	// 通用建议
	recommendations = append(recommendations,
		"开启夜间模式或蓝光过滤功能，减少蓝光对睡眠的影响",
		"睡前1小时避免使用电子设备，改为阅读纸质书籍或冥想",
		"设置固定的睡眠时间，培养良好的睡眠习惯",
	)

	return recommendations
}