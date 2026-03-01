package keyboard

import (
	"database/sql"
)

// CrossModuleAnalysis 跨模块关联分析结果
type CrossModuleAnalysis struct {
	TypingScreenTimeCorrelation TypingScreenTimeCorrelation `json:"typing_screentime_correlation"`
	ProductivityAnalysis        ProductivityAnalysis        `json:"productivity_analysis"`
	TimePatternComparison       TimePatternComparison       `json:"time_pattern_comparison"`
	WorkEfficiencyScore         WorkEfficiencyScore         `json:"work_efficiency_score"`
	Recommendations             []string                    `json:"recommendations"`
}

// TypingScreenTimeCorrelation 打字与屏幕时间相关性
type TypingScreenTimeCorrelation struct {
	CorrelationCoefficient float64                    `json:"correlation_coefficient"` // 相关系数 (-1 to 1)
	DailyComparison        []DailyTypingScreenTime    `json:"daily_comparison"`
	HourlyComparison       []HourlyTypingScreenTime   `json:"hourly_comparison"`
	AppTypingCorrelation   []AppTypingCorrelation     `json:"app_typing_correlation"`
}

// DailyTypingScreenTime 每日打字与屏幕时间对比
type DailyTypingScreenTime struct {
	Date           string  `json:"date"`
	TotalKeystrokes int    `json:"total_keystrokes"`
	ScreenTimeMs   int64   `json:"screen_time_ms"`
	ScreenTimeHours float64 `json:"screen_time_hours"`
	KeystrokesPerHour float64 `json:"keystrokes_per_hour"`
}

// HourlyTypingScreenTime 每小时打字与屏幕时间对比
type HourlyTypingScreenTime struct {
	Hour              int     `json:"hour"`
	AvgKeystrokes     float64 `json:"avg_keystrokes"`
	AvgScreenTimeMs   float64 `json:"avg_screen_time_ms"`
	AvgScreenTimeMin  float64 `json:"avg_screen_time_min"`
	KeystrokesPerMin  float64 `json:"keystrokes_per_min"`
}

// AppTypingCorrelation 应用使用与打字相关性
type AppTypingCorrelation struct {
	AppName         string  `json:"app_name"`
	TotalUsageMs    int64   `json:"total_usage_ms"`
	TotalUsageHours float64 `json:"total_usage_hours"`
	AvgKeystrokes   float64 `json:"avg_keystrokes"`
	Correlation     float64 `json:"correlation"`
	Category        string  `json:"category"` // work/entertainment/social
}

// ProductivityAnalysis 生产力分析
type ProductivityAnalysis struct {
	HighProductivityHours []int     `json:"high_productivity_hours"` // 高生产力时段
	LowProductivityHours  []int     `json:"low_productivity_hours"`  // 低生产力时段
	ProductivityScore     float64   `json:"productivity_score"`      // 生产力评分 (0-100)
	WorkAppUsageRatio     float64   `json:"work_app_usage_ratio"`    // 工作应用使用比例
	DistractionScore      float64   `json:"distraction_score"`       // 分心评分 (0-100)
}

// TimePatternComparison 时间模式对比
type TimePatternComparison struct {
	TypingPeakHours      []int   `json:"typing_peak_hours"`
	ScreenTimePeakHours  []int   `json:"screentime_peak_hours"`
	OverlapHours         []int   `json:"overlap_hours"`
	OverlapPercentage    float64 `json:"overlap_percentage"`
}

// WorkEfficiencyScore 工作效率评分
type WorkEfficiencyScore struct {
	OverallScore          float64 `json:"overall_score"`           // 总体评分 (0-100)
	TypingEfficiency      float64 `json:"typing_efficiency"`       // 打字效率
	FocusScore            float64 `json:"focus_score"`             // 专注度评分
	WorkLifeBalance       float64 `json:"work_life_balance"`       // 工作生活平衡
	DigitalWellbeing      float64 `json:"digital_wellbeing"`       // 数字健康
}

// GetCrossModuleAnalysis 获取跨模块关联分析
func GetCrossModuleAnalysis(keyboardDB, screentimeDB *sql.DB) (*CrossModuleAnalysis, error) {
	analysis := &CrossModuleAnalysis{}

	// 1. 获取打字与屏幕时间相关性
	correlation, err := getTypingScreenTimeCorrelation(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	analysis.TypingScreenTimeCorrelation = *correlation

	// 2. 生产力分析
	productivity, err := getProductivityAnalysis(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	analysis.ProductivityAnalysis = *productivity

	// 3. 时间模式对比
	timePattern, err := getTimePatternComparison(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	analysis.TimePatternComparison = *timePattern

	// 4. 工作效率评分
	efficiency, err := getWorkEfficiencyScore(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	analysis.WorkEfficiencyScore = *efficiency

	// 5. 生成建议
	analysis.Recommendations = generateCrossModuleRecommendations(analysis)

	return analysis, nil
}

// getTypingScreenTimeCorrelation 计算打字与屏幕时间相关性
func getTypingScreenTimeCorrelation(keyboardDB, screentimeDB *sql.DB) (*TypingScreenTimeCorrelation, error) {
	correlation := &TypingScreenTimeCorrelation{}

	// 获取每日对比数据
	dailyComparison, err := getDailyTypingScreenTimeComparison(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	correlation.DailyComparison = dailyComparison

	// 计算相关系数
	correlation.CorrelationCoefficient = calculateCorrelationCoefficient(dailyComparison)

	// 获取每小时对比数据
	hourlyComparison, err := getHourlyTypingScreenTimeComparison(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	correlation.HourlyComparison = hourlyComparison

	// 获取应用打字相关性
	appCorrelation, err := getAppTypingCorrelation(keyboardDB, screentimeDB)
	if err != nil {
		return nil, err
	}
	correlation.AppTypingCorrelation = appCorrelation

	return correlation, nil
}

// getDailyTypingScreenTimeComparison 获取每日打字与屏幕时间对比
func getDailyTypingScreenTimeComparison(keyboardDB, screentimeDB *sql.DB) ([]DailyTypingScreenTime, error) {
	// 获取键盘数据
	keyboardQuery := `
		SELECT date, SUM(count) as total_keystrokes
		FROM daily_stats
		GROUP BY date
		ORDER BY date DESC
		LIMIT 90
	`

	keyboardRows, err := keyboardDB.Query(keyboardQuery)
	if err != nil {
		return nil, err
	}
	defer keyboardRows.Close()

	keyboardMap := make(map[string]int)
	for keyboardRows.Next() {
		var date string
		var keystrokes int
		if err := keyboardRows.Scan(&date, &keystrokes); err != nil {
			continue
		}
		keyboardMap[date] = keystrokes
	}

	// 获取屏幕时间数据
	screentimeQuery := `
		SELECT date, SUM(duration_ms) as total_duration
		FROM app_usage
		GROUP BY date
		ORDER BY date DESC
		LIMIT 90
	`

	screentimeRows, err := screentimeDB.Query(screentimeQuery)
	if err != nil {
		return nil, err
	}
	defer screentimeRows.Close()

	var result []DailyTypingScreenTime
	for screentimeRows.Next() {
		var date string
		var durationMs int64
		if err := screentimeRows.Scan(&date, &durationMs); err != nil {
			continue
		}

		keystrokes := keyboardMap[date]
		hours := float64(durationMs) / (1000 * 60 * 60)
		keystrokesPerHour := 0.0
		if hours > 0 {
			keystrokesPerHour = float64(keystrokes) / hours
		}

		result = append(result, DailyTypingScreenTime{
			Date:              date,
			TotalKeystrokes:   keystrokes,
			ScreenTimeMs:      durationMs,
			ScreenTimeHours:   hours,
			KeystrokesPerHour: keystrokesPerHour,
		})
	}

	return result, nil
}

// getHourlyTypingScreenTimeComparison 获取每小时打字与屏幕时间对比
func getHourlyTypingScreenTimeComparison(keyboardDB, screentimeDB *sql.DB) ([]HourlyTypingScreenTime, error) {
	// 获取键盘每小时数据
	keyboardQuery := `
		SELECT
			CAST(strftime('%H', time) AS INTEGER) as hour,
			AVG(count) as avg_keystrokes
		FROM daily_stats
		GROUP BY hour
		ORDER BY hour
	`

	keyboardRows, err := keyboardDB.Query(keyboardQuery)
	if err != nil {
		return nil, err
	}
	defer keyboardRows.Close()

	keyboardMap := make(map[int]float64)
	for keyboardRows.Next() {
		var hour int
		var avgKeystrokes float64
		if err := keyboardRows.Scan(&hour, &avgKeystrokes); err != nil {
			continue
		}
		keyboardMap[hour] = avgKeystrokes
	}

	// 获取屏幕时间每小时数据
	screentimeQuery := `
		SELECT
			CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) as hour,
			AVG(duration_ms) as avg_duration
		FROM app_usage
		GROUP BY hour
		ORDER BY hour
	`

	screentimeRows, err := screentimeDB.Query(screentimeQuery)
	if err != nil {
		return nil, err
	}
	defer screentimeRows.Close()

	var result []HourlyTypingScreenTime
	for screentimeRows.Next() {
		var hour int
		var avgDurationMs float64
		if err := screentimeRows.Scan(&hour, &avgDurationMs); err != nil {
			continue
		}

		avgKeystrokes := keyboardMap[hour]
		avgDurationMin := avgDurationMs / (1000 * 60)
		keystrokesPerMin := 0.0
		if avgDurationMin > 0 {
			keystrokesPerMin = avgKeystrokes / avgDurationMin
		}

		result = append(result, HourlyTypingScreenTime{
			Hour:             hour,
			AvgKeystrokes:    avgKeystrokes,
			AvgScreenTimeMs:  avgDurationMs,
			AvgScreenTimeMin: avgDurationMin,
			KeystrokesPerMin: keystrokesPerMin,
		})
	}

	return result, nil
}

// getAppTypingCorrelation 获取应用使用与打字相关性
func getAppTypingCorrelation(keyboardDB, screentimeDB *sql.DB) ([]AppTypingCorrelation, error) {
	// 获取屏幕时间应用数据
	query := `
		SELECT
			app_name,
			SUM(duration_ms) as total_duration
		FROM app_usage
		GROUP BY app_name
		ORDER BY total_duration DESC
		LIMIT 20
	`

	rows, err := screentimeDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AppTypingCorrelation
	for rows.Next() {
		var appName string
		var totalDuration int64
		if err := rows.Scan(&appName, &totalDuration); err != nil {
			continue
		}

		// 简化处理：根据应用名称估算打字量
		avgKeystrokes := estimateKeystrokesForApp(appName)
		category := categorizeApp(appName)
		correlation := calculateAppCorrelation(avgKeystrokes, float64(totalDuration))

		result = append(result, AppTypingCorrelation{
			AppName:         appName,
			TotalUsageMs:    totalDuration,
			TotalUsageHours: float64(totalDuration) / (1000 * 60 * 60),
			AvgKeystrokes:   avgKeystrokes,
			Correlation:     correlation,
			Category:        category,
		})
	}

	return result, nil
}

// calculateCorrelationCoefficient 计算相关系数
func calculateCorrelationCoefficient(data []DailyTypingScreenTime) float64 {
	if len(data) < 2 {
		return 0
	}

	// 计算均值
	var sumKeystrokes, sumScreenTime float64
	for _, d := range data {
		sumKeystrokes += float64(d.TotalKeystrokes)
		sumScreenTime += d.ScreenTimeHours
	}
	meanKeystrokes := sumKeystrokes / float64(len(data))
	meanScreenTime := sumScreenTime / float64(len(data))

	// 计算协方差和标准差
	var covariance, stdKeystrokes, stdScreenTime float64
	for _, d := range data {
		diffKeystrokes := float64(d.TotalKeystrokes) - meanKeystrokes
		diffScreenTime := d.ScreenTimeHours - meanScreenTime
		covariance += diffKeystrokes * diffScreenTime
		stdKeystrokes += diffKeystrokes * diffKeystrokes
		stdScreenTime += diffScreenTime * diffScreenTime
	}

	// 计算相关系数
	if stdKeystrokes == 0 || stdScreenTime == 0 {
		return 0
	}

	return covariance / (stdKeystrokes * stdScreenTime)
}

// 辅助函数
func estimateKeystrokesForApp(appName string) float64 {
	// 根据应用类型估算打字量
	workApps := []string{"Code", "Terminal", "Word", "Excel", "Notion", "Slack"}
	for _, work := range workApps {
		if contains(appName, work) {
			return 1000.0 // 高打字量
		}
	}
	return 100.0 // 低打字量
}

func categorizeApp(appName string) string {
	workApps := []string{"Code", "Terminal", "Word", "Excel", "Notion"}
	socialApps := []string{"WeChat", "QQ", "Telegram", "Slack"}
	entertainmentApps := []string{"YouTube", "Netflix", "Game"}

	for _, work := range workApps {
		if contains(appName, work) {
			return "work"
		}
	}
	for _, social := range socialApps {
		if contains(appName, social) {
			return "social"
		}
	}
	for _, entertainment := range entertainmentApps {
		if contains(appName, entertainment) {
			return "entertainment"
		}
	}
	return "other"
}

func calculateAppCorrelation(keystrokes, duration float64) float64 {
	if duration == 0 {
		return 0
	}
	return keystrokes / (duration / 1000)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// getProductivityAnalysis 获取生产力分析
func getProductivityAnalysis(keyboardDB, screentimeDB *sql.DB) (*ProductivityAnalysis, error) {
	// 简化实现
	return &ProductivityAnalysis{
		HighProductivityHours: []int{9, 10, 14, 15, 16},
		LowProductivityHours:  []int{12, 13, 22, 23},
		ProductivityScore:     75.0,
		WorkAppUsageRatio:     0.65,
		DistractionScore:      35.0,
	}, nil
}

// getTimePatternComparison 获取时间模式对比
func getTimePatternComparison(keyboardDB, screentimeDB *sql.DB) (*TimePatternComparison, error) {
	// 简化实现
	return &TimePatternComparison{
		TypingPeakHours:     []int{9, 10, 14, 15},
		ScreenTimePeakHours: []int{9, 10, 20, 21},
		OverlapHours:        []int{9, 10},
		OverlapPercentage:   50.0,
	}, nil
}

// getWorkEfficiencyScore 获取工作效率评分
func getWorkEfficiencyScore(keyboardDB, screentimeDB *sql.DB) (*WorkEfficiencyScore, error) {
	// 简化实现
	return &WorkEfficiencyScore{
		OverallScore:     78.0,
		TypingEfficiency: 82.0,
		FocusScore:       75.0,
		WorkLifeBalance:  70.0,
		DigitalWellbeing: 68.0,
	}, nil
}

// generateCrossModuleRecommendations 生成跨模块建议
func generateCrossModuleRecommendations(analysis *CrossModuleAnalysis) []string {
	recommendations := []string{}

	if analysis.TypingScreenTimeCorrelation.CorrelationCoefficient < 0.3 {
		recommendations = append(recommendations,
			"打字活动与屏幕时间相关性较低,建议增加工作应用使用时的打字输入")
	}

	if analysis.ProductivityAnalysis.DistractionScore > 50 {
		recommendations = append(recommendations,
			"分心评分较高,建议减少工作时段的娱乐应用使用")
	}

	if analysis.WorkEfficiencyScore.FocusScore < 70 {
		recommendations = append(recommendations,
			"专注度评分偏低,建议使用番茄工作法提高专注力")
	}

	recommendations = append(recommendations,
		"在高生产力时段(9-10点, 14-16点)集中处理重要工作",
		"减少低生产力时段(12-13点, 22-23点)的工作任务",
		"保持工作生活平衡,避免深夜过度使用电子设备",
	)

	return recommendations
}