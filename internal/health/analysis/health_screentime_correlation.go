package analysis

import (
	"database/sql"
	"fmt"
	"math"
)

// HealthScreentimeCorrelation 健康与屏幕时间关联分析结果
type HealthScreentimeCorrelation struct {
	SedentaryAnalysis      SedentaryAnalysis      `json:"sedentary_analysis"`       // 久坐分析
	ActivityCorrelation    ActivityCorrelation    `json:"activity_correlation"`     // 活动相关性
	SleepImpact            SleepImpact            `json:"sleep_impact"`             // 睡眠影响
	HealthScreentimeBalance HealthScreentimeBalance `json:"health_screentime_balance"` // 健康屏幕平衡
	Recommendations        []string               `json:"recommendations"`          // 建议
}

// SedentaryAnalysis 久坐分析
type SedentaryAnalysis struct {
	TotalSedentaryDays    int     `json:"total_sedentary_days"`    // 总久坐天数
	AvgSedentaryHours     float64 `json:"avg_sedentary_hours"`     // 平均久坐时长(小时)
	MaxSedentaryHours     float64 `json:"max_sedentary_hours"`     // 最大久坐时长(小时)
	SedentaryRate         float64 `json:"sedentary_rate"`          // 久坐率(%)
	HighRiskDays          int     `json:"high_risk_days"`          // 高风险天数(>6小时)
	SedentaryDayDetails   []SedentaryDay `json:"sedentary_day_details"` // 久坐日详情
}

// SedentaryDay 久坐日详情
type SedentaryDay struct {
	Date          string  `json:"date"`           // 日期
	ScreenHours   float64 `json:"screen_hours"`   // 屏幕时间(小时)
	Steps         int     `json:"steps"`          // 步数
	RiskLevel     string  `json:"risk_level"`     // 风险等级: low/medium/high
}

// ActivityCorrelation 活动相关性
type ActivityCorrelation struct {
	CorrelationCoefficient float64                `json:"correlation_coefficient"` // 相关系数(-1到1)
	CorrelationType        string                 `json:"correlation_type"`        // 相关类型: negative/none/positive
	DailyComparison        []DailyActivityScreen  `json:"daily_comparison"`        // 每日对比
	AverageStepsHighScreen float64                `json:"avg_steps_high_screen"`   // 高屏幕时间日平均步数
	AverageStepsLowScreen  float64                `json:"avg_steps_low_screen"`    // 低屏幕时间日平均步数
}

// DailyActivityScreen 每日活动与屏幕时间
type DailyActivityScreen struct {
	Date        string  `json:"date"`         // 日期
	Steps       int     `json:"steps"`        // 步数
	ScreenHours float64 `json:"screen_hours"` // 屏幕时间(小时)
}

// SleepImpact 睡眠影响
type SleepImpact struct {
	LateNightScreenDays     int     `json:"late_night_screen_days"`      // 深夜使用天数
	AvgSleepWithLateScreen  float64 `json:"avg_sleep_with_late_screen"`  // 深夜使用后平均睡眠(小时)
	AvgSleepWithoutLateScreen float64 `json:"avg_sleep_without_late_screen"` // 无深夜使用平均睡眠(小时)
	SleepQualityImpact      float64 `json:"sleep_quality_impact"`        // 睡眠质量影响(小时差)
	ImpactLevel             string  `json:"impact_level"`                // 影响等级: low/medium/high
}

// HealthScreentimeBalance 健康屏幕平衡
type HealthScreentimeBalance struct {
	BalanceScore       float64 `json:"balance_score"`        // 平衡评分(0-100)
	HealthyDays        int     `json:"healthy_days"`         // 健康天数
	UnhealthyDays      int     `json:"unhealthy_days"`       // 不健康天数
	RecommendedScreen  float64 `json:"recommended_screen"`   // 推荐屏幕时间(小时)
	CurrentAvgScreen   float64 `json:"current_avg_screen"`   // 当前平均屏幕时间(小时)
}

// GetHealthScreentimeCorrelation 获取健康与屏幕时间关联分析
func GetHealthScreentimeCorrelation(healthDB, screentimeDB *sql.DB) (*HealthScreentimeCorrelation, error) {
	correlation := &HealthScreentimeCorrelation{}

	// 1. 久坐分析
	sedentary, err := analyzeSedentaryBehavior(healthDB, screentimeDB)
	if err != nil {
		return nil, fmt.Errorf("sedentary analysis failed: %w", err)
	}
	correlation.SedentaryAnalysis = sedentary

	// 2. 活动相关性分析
	activityCorr, err := analyzeActivityCorrelation(healthDB, screentimeDB)
	if err != nil {
		return nil, fmt.Errorf("activity correlation failed: %w", err)
	}
	correlation.ActivityCorrelation = activityCorr

	// 3. 睡眠影响分析
	sleepImpact, err := analyzeSleepImpact(healthDB, screentimeDB)
	if err != nil {
		return nil, fmt.Errorf("sleep impact analysis failed: %w", err)
	}
	correlation.SleepImpact = sleepImpact

	// 4. 健康屏幕平衡
	balance := calculateHealthScreentimeBalance(sedentary, activityCorr)
	correlation.HealthScreentimeBalance = balance

	// 5. 生成建议
	correlation.Recommendations = generateHealthScreentimeRecommendations(correlation)

	return correlation, nil
}

// analyzeSedentaryBehavior 分析久坐行为
func analyzeSedentaryBehavior(healthDB, screentimeDB *sql.DB) (SedentaryAnalysis, error) {
	analysis := SedentaryAnalysis{
		SedentaryDayDetails: []SedentaryDay{},
	}

	// 查询每日屏幕时间和步数
	query := `
	SELECT
		date,
		screen_hours,
		steps
	FROM (
		SELECT
			h.date,
			COALESCE(SUM(CASE WHEN s.duration_ms IS NOT NULL THEN s.duration_ms END) / 3600000.0, 0) as screen_hours,
			COALESCE(AVG(CASE WHEN h.type = 'StepCount' THEN h.value END), 0) as steps
		FROM (
			SELECT DISTINCT date FROM health_records WHERE type = 'StepCount'
		) h
		LEFT JOIN (
			SELECT date, duration_ms FROM screentime_daily WHERE package_id != 'ALL'
		) s ON h.date = s.date
		GROUP BY h.date
	)
	WHERE screen_hours > 0 OR steps > 0
	ORDER BY date DESC
	LIMIT 90
	`

	rows, err := healthDB.Query(query)
	if err != nil {
		return analysis, err
	}
	defer rows.Close()

	var totalSedentaryHours float64
	var sedentaryCount int
	var maxHours float64

	for rows.Next() {
		var day SedentaryDay
		var steps float64

		if err := rows.Scan(&day.Date, &day.ScreenHours, &steps); err != nil {
			continue
		}

		day.Steps = int(steps)

		// 判断是否久坐：屏幕时间>2小时且步数<5000
		if day.ScreenHours > 2 && day.Steps < 5000 {
			sedentaryCount++
			totalSedentaryHours += day.ScreenHours

			if day.ScreenHours > maxHours {
				maxHours = day.ScreenHours
			}

			// 风险等级判断
			if day.ScreenHours > 6 {
				day.RiskLevel = "high"
				analysis.HighRiskDays++
			} else if day.ScreenHours > 4 {
				day.RiskLevel = "medium"
			} else {
				day.RiskLevel = "low"
			}

			analysis.SedentaryDayDetails = append(analysis.SedentaryDayDetails, day)
		}
	}

	analysis.TotalSedentaryDays = sedentaryCount
	if sedentaryCount > 0 {
		analysis.AvgSedentaryHours = totalSedentaryHours / float64(sedentaryCount)
	}
	analysis.MaxSedentaryHours = maxHours

	// 计算久坐率（假设总天数为90天）
	analysis.SedentaryRate = float64(sedentaryCount) / 90.0 * 100

	return analysis, nil
}

// analyzeActivityCorrelation 分析活动相关性
func analyzeActivityCorrelation(healthDB, screentimeDB *sql.DB) (ActivityCorrelation, error) {
	correlation := ActivityCorrelation{
		DailyComparison: []DailyActivityScreen{},
	}

	// 查询每日数据
	query := `
	SELECT
		date,
		steps,
		screen_hours
	FROM (
		SELECT
			h.date,
			COALESCE(AVG(CASE WHEN h.type = 'StepCount' THEN h.value END), 0) as steps,
			COALESCE(SUM(CASE WHEN s.duration_ms IS NOT NULL THEN s.duration_ms END) / 3600000.0, 0) as screen_hours
		FROM (
			SELECT DISTINCT date FROM health_records WHERE type = 'StepCount'
		) h
		LEFT JOIN (
			SELECT date, duration_ms FROM screentime_daily WHERE package_id != 'ALL'
		) s ON h.date = s.date
		GROUP BY h.date
	)
	WHERE steps > 0 AND screen_hours > 0
	ORDER BY date DESC
	LIMIT 60
	`

	rows, err := healthDB.Query(query)
	if err != nil {
		return correlation, err
	}
	defer rows.Close()

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	var n float64
	var highScreenSteps, lowScreenSteps []float64

	for rows.Next() {
		var day DailyActivityScreen
		var steps float64

		if err := rows.Scan(&day.Date, &steps, &day.ScreenHours); err != nil {
			continue
		}

		day.Steps = int(steps)
		correlation.DailyComparison = append(correlation.DailyComparison, day)

		// 计算相关系数
		x := day.ScreenHours
		y := steps

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
		n++

		// 分类高低屏幕时间
		if day.ScreenHours > 4 {
			highScreenSteps = append(highScreenSteps, steps)
		} else {
			lowScreenSteps = append(lowScreenSteps, steps)
		}
	}

	// 计算Pearson相关系数
	if n > 1 {
		numerator := n*sumXY - sumX*sumY
		denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
		if denominator != 0 {
			correlation.CorrelationCoefficient = numerator / denominator
		}
	}

	// 判断相关类型
	if correlation.CorrelationCoefficient < -0.3 {
		correlation.CorrelationType = "negative"
	} else if correlation.CorrelationCoefficient > 0.3 {
		correlation.CorrelationType = "positive"
	} else {
		correlation.CorrelationType = "none"
	}

	// 计算平均步数
	if len(highScreenSteps) > 0 {
		sum := 0.0
		for _, s := range highScreenSteps {
			sum += s
		}
		correlation.AverageStepsHighScreen = sum / float64(len(highScreenSteps))
	}

	if len(lowScreenSteps) > 0 {
		sum := 0.0
		for _, s := range lowScreenSteps {
			sum += s
		}
		correlation.AverageStepsLowScreen = sum / float64(len(lowScreenSteps))
	}

	return correlation, nil
}

// analyzeSleepImpact 分析睡眠影响
func analyzeSleepImpact(healthDB, screentimeDB *sql.DB) (SleepImpact, error) {
	impact := SleepImpact{}

	// 查询深夜屏幕使用（22:00-02:00）和睡眠数据
	// 注意：这里简化处理，实际需要更复杂的逻辑来关联深夜使用和次日睡眠
	query := `
	SELECT
		COUNT(DISTINCT CASE WHEN late_night_usage > 0 THEN date END) as late_night_days,
		AVG(CASE WHEN late_night_usage > 0 THEN sleep_hours END) as avg_sleep_with_late,
		AVG(CASE WHEN late_night_usage = 0 THEN sleep_hours END) as avg_sleep_without_late
	FROM (
		SELECT
			h.date,
			COALESCE(SUM(CASE
				WHEN CAST(substr(s.start_time, 1, 2) AS INTEGER) >= 22
				OR CAST(substr(s.start_time, 1, 2) AS INTEGER) <= 2
				THEN 1 ELSE 0 END), 0) as late_night_usage,
			COALESCE(AVG(CASE WHEN h.type = 'SleepAnalysis' THEN h.value / 3600.0 END), 0) as sleep_hours
		FROM (
			SELECT DISTINCT date FROM health_records WHERE type = 'SleepAnalysis'
		) h
		LEFT JOIN screentime_sessions s ON h.date = s.date
		GROUP BY h.date
	)
	WHERE sleep_hours > 0
	`

	err := healthDB.QueryRow(query).Scan(
		&impact.LateNightScreenDays,
		&impact.AvgSleepWithLateScreen,
		&impact.AvgSleepWithoutLateScreen,
	)

	if err != nil && err != sql.ErrNoRows {
		return impact, err
	}

	// 计算睡眠质量影响
	impact.SleepQualityImpact = impact.AvgSleepWithoutLateScreen - impact.AvgSleepWithLateScreen

	// 判断影响等级
	if impact.SleepQualityImpact > 1.0 {
		impact.ImpactLevel = "high"
	} else if impact.SleepQualityImpact > 0.5 {
		impact.ImpactLevel = "medium"
	} else {
		impact.ImpactLevel = "low"
	}

	return impact, nil
}

// calculateHealthScreentimeBalance 计算健康屏幕平衡
func calculateHealthScreentimeBalance(sedentary SedentaryAnalysis, activity ActivityCorrelation) HealthScreentimeBalance {
	balance := HealthScreentimeBalance{}

	// 计算健康天数和不健康天数
	balance.UnhealthyDays = sedentary.TotalSedentaryDays
	balance.HealthyDays = 90 - balance.UnhealthyDays // 假设90天总数

	// 计算平衡评分
	// 基础分100，久坐率每10%扣10分，相关系数为负加分
	balance.BalanceScore = 100.0
	balance.BalanceScore -= sedentary.SedentaryRate / 10.0 * 10.0

	if activity.CorrelationType == "negative" {
		balance.BalanceScore += 10.0 // 负相关是好事
	}

	if balance.BalanceScore < 0 {
		balance.BalanceScore = 0
	}
	if balance.BalanceScore > 100 {
		balance.BalanceScore = 100
	}

	// 推荐屏幕时间：基于WHO建议，成人每天不超过4小时
	balance.RecommendedScreen = 4.0

	// 当前平均屏幕时间
	if sedentary.TotalSedentaryDays > 0 {
		balance.CurrentAvgScreen = sedentary.AvgSedentaryHours
	}

	return balance
}

// generateHealthScreentimeRecommendations 生成健康屏幕建议
func generateHealthScreentimeRecommendations(correlation *HealthScreentimeCorrelation) []string {
	recommendations := []string{}

	// 久坐建议
	if correlation.SedentaryAnalysis.SedentaryRate > 50 {
		recommendations = append(recommendations,
			fmt.Sprintf("你有%.1f%%的天数处于久坐状态，建议每2小时起身活动10分钟",
				correlation.SedentaryAnalysis.SedentaryRate))
	}

	if correlation.SedentaryAnalysis.HighRiskDays > 10 {
		recommendations = append(recommendations,
			fmt.Sprintf("检测到%d天高风险久坐(>6小时)，建议设置屏幕时间提醒",
				correlation.SedentaryAnalysis.HighRiskDays))
	}

	// 活动相关性建议
	if correlation.ActivityCorrelation.CorrelationType == "negative" {
		recommendations = append(recommendations,
			"屏幕时间与活动量呈负相关，减少屏幕时间可以增加运动量")
	}

	stepDiff := correlation.ActivityCorrelation.AverageStepsLowScreen -
		correlation.ActivityCorrelation.AverageStepsHighScreen
	if stepDiff > 2000 {
		recommendations = append(recommendations,
			fmt.Sprintf("低屏幕时间日比高屏幕时间日平均多走%.0f步，建议控制屏幕使用", stepDiff))
	}

	// 睡眠影响建议
	if correlation.SleepImpact.ImpactLevel == "high" {
		recommendations = append(recommendations,
			fmt.Sprintf("深夜使用屏幕导致睡眠减少%.1f小时，建议22点后停止使用电子设备",
				correlation.SleepImpact.SleepQualityImpact))
	}

	// 平衡评分建议
	if correlation.HealthScreentimeBalance.BalanceScore < 60 {
		recommendations = append(recommendations,
			"健康屏幕平衡评分较低，建议制定数字健康计划")
	}

	if correlation.HealthScreentimeBalance.CurrentAvgScreen >
		correlation.HealthScreentimeBalance.RecommendedScreen {
		recommendations = append(recommendations,
			fmt.Sprintf("当前平均屏幕时间%.1f小时，超过推荐值%.1f小时",
				correlation.HealthScreentimeBalance.CurrentAvgScreen,
				correlation.HealthScreentimeBalance.RecommendedScreen))
	}

	// 通用建议
	recommendations = append(recommendations,
		"建议：每天至少30分钟中等强度运动，保持7-9小时睡眠")

	return recommendations
}
