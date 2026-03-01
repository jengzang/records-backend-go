package screentime

import (
	"sort"
)

// AppSwitchingPattern 应用切换模式分析结果
type AppSwitchingPattern struct {
	OverallPattern      OverallSwitchingPattern `json:"overall_pattern"`
	ModeDistribution    ModeDistribution        `json:"mode_distribution"`
	CommonSwitchPaths   []SwitchPath            `json:"common_switch_paths"`
	SwitchingFrequency  SwitchingFrequency      `json:"switching_frequency"`
	TimeDistribution    []TimeDistribution      `json:"time_distribution"`
	AppSwitchMatrix     []AppSwitchPair         `json:"app_switch_matrix"`
	Recommendations     []string                `json:"recommendations"`
}

// OverallSwitchingPattern 总体切换模式
type OverallSwitchingPattern struct {
	DominantMode        string  `json:"dominant_mode"`         // multitask/focus/fragmented
	AvgSwitchesPerDay   float64 `json:"avg_switches_per_day"`
	AvgSessionDuration  float64 `json:"avg_session_duration"`  // 分钟
	TotalSessions       int     `json:"total_sessions"`
	MultitaskScore      float64 `json:"multitask_score"`       // 0-100
	FocusScore          float64 `json:"focus_score"`           // 0-100
	FragmentationScore  float64 `json:"fragmentation_score"`   // 0-100
}

// ModeDistribution 模式分布
type ModeDistribution struct {
	MultitaskSessions   int     `json:"multitask_sessions"`
	FocusSessions       int     `json:"focus_sessions"`
	FragmentedSessions  int     `json:"fragmented_sessions"`
	MultitaskPercentage float64 `json:"multitask_percentage"`
	FocusPercentage     float64 `json:"focus_percentage"`
	FragmentedPercentage float64 `json:"fragmented_percentage"`
}

// SwitchPath 切换路径
type SwitchPath struct {
	Path        string  `json:"path"`          // "App A → App B → App C"
	Frequency   int     `json:"frequency"`
	AvgDuration float64 `json:"avg_duration"`  // 分钟
	Pattern     string  `json:"pattern"`       // work/social/mixed
}

// SwitchingFrequency 切换频率
type SwitchingFrequency struct {
	ByHour      []HourlySwitch `json:"by_hour"`
	ByDayOfWeek []DaySwitch    `json:"by_day_of_week"`
	ByDate      []DateSwitch   `json:"by_date"`
}

// HourlySwitch 每小时切换
type HourlySwitch struct {
	Hour          int     `json:"hour"`
	AvgSwitches   float64 `json:"avg_switches"`
	DominantMode  string  `json:"dominant_mode"`
}

// DaySwitch 每周切换
type DaySwitch struct {
	DayOfWeek    int     `json:"day_of_week"`
	DayName      string  `json:"day_name"`
	AvgSwitches  float64 `json:"avg_switches"`
	DominantMode string  `json:"dominant_mode"`
}

// DateSwitch 每日切换
type DateSwitch struct {
	Date         string  `json:"date"`
	TotalSwitches int    `json:"total_switches"`
	DominantMode string  `json:"dominant_mode"`
}

// TimeDistribution 时段分布
type TimeDistribution struct {
	Hour              int     `json:"hour"`
	MultitaskPercent  float64 `json:"multitask_percent"`
	FocusPercent      float64 `json:"focus_percent"`
	FragmentedPercent float64 `json:"fragmented_percent"`
}

// AppSwitchPair 应用切换对
type AppSwitchPair struct {
	FromApp   string  `json:"from_app"`
	ToApp     string  `json:"to_app"`
	Count     int     `json:"count"`
	AvgDelay  float64 `json:"avg_delay"`  // 秒
}

// GetAppSwitchingPattern 获取应用切换模式分析
func (h *Handler) GetAppSwitchingPattern() (*AppSwitchingPattern, error) {
	pattern := &AppSwitchingPattern{}

	// 1. 分析总体切换模式
	overall, err := h.analyzeOverallPattern()
	if err != nil {
		return nil, err
	}
	pattern.OverallPattern = *overall

	// 2. 模式分布
	modeDistribution, err := h.analyzeModeDistribution()
	if err != nil {
		return nil, err
	}
	pattern.ModeDistribution = *modeDistribution

	// 3. 常见切换路径
	switchPaths, err := h.analyzeCommonSwitchPaths()
	if err != nil {
		return nil, err
	}
	pattern.CommonSwitchPaths = switchPaths

	// 4. 切换频率
	frequency, err := h.analyzeSwitchingFrequency()
	if err != nil {
		return nil, err
	}
	pattern.SwitchingFrequency = *frequency

	// 5. 时段分布
	timeDistribution, err := h.analyzeTimeDistribution()
	if err != nil {
		return nil, err
	}
	pattern.TimeDistribution = timeDistribution

	// 6. 应用切换矩阵
	switchMatrix, err := h.analyzeAppSwitchMatrix()
	if err != nil {
		return nil, err
	}
	pattern.AppSwitchMatrix = switchMatrix

	// 7. 生成建议
	pattern.Recommendations = generateSwitchingRecommendations(pattern)

	return pattern, nil
}

// analyzeOverallPattern 分析总体切换模式
func (h *Handler) analyzeOverallPattern() (*OverallSwitchingPattern, error) {
	query := `
		SELECT
			COUNT(*) as total_sessions,
			AVG(duration_ms) / 60000.0 as avg_duration_min
		FROM app_usage
		WHERE duration_ms > 0
	`

	var totalSessions int
	var avgDuration float64

	err := h.db.QueryRow(query).Scan(&totalSessions, &avgDuration)
	if err != nil {
		return nil, err
	}

	// 计算每日平均切换次数
	daysQuery := `
		SELECT COUNT(DISTINCT date) as total_days
		FROM app_usage
	`
	var totalDays int
	h.db.QueryRow(daysQuery).Scan(&totalDays)

	avgSwitchesPerDay := 0.0
	if totalDays > 0 {
		avgSwitchesPerDay = float64(totalSessions) / float64(totalDays)
	}

	// 计算模式评分
	multitaskScore := calculateMultitaskScore(avgSwitchesPerDay, avgDuration)
	focusScore := calculateFocusScore(avgDuration)
	fragmentationScore := calculateFragmentationScore(avgSwitchesPerDay, avgDuration)

	// 确定主导模式
	dominantMode := "focus"
	maxScore := focusScore
	if multitaskScore > maxScore {
		dominantMode = "multitask"
		maxScore = multitaskScore
	}
	if fragmentationScore > maxScore {
		dominantMode = "fragmented"
	}

	return &OverallSwitchingPattern{
		DominantMode:       dominantMode,
		AvgSwitchesPerDay:  avgSwitchesPerDay,
		AvgSessionDuration: avgDuration,
		TotalSessions:      totalSessions,
		MultitaskScore:     multitaskScore,
		FocusScore:         focusScore,
		FragmentationScore: fragmentationScore,
	}, nil
}

// analyzeModeDistribution 分析模式分布
func (h *Handler) analyzeModeDistribution() (*ModeDistribution, error) {
	query := `
		SELECT
			duration_ms / 60000.0 as duration_min
		FROM app_usage
		WHERE duration_ms > 0
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var multitaskCount, focusCount, fragmentedCount int

	for rows.Next() {
		var duration float64
		if err := rows.Scan(&duration); err != nil {
			continue
		}

		// 分类规则:
		// 专注模式: 单次使用 > 30分钟
		// 碎片化模式: 单次使用 < 5分钟
		// 多任务模式: 5-30分钟
		if duration > 30 {
			focusCount++
		} else if duration < 5 {
			fragmentedCount++
		} else {
			multitaskCount++
		}
	}

	total := multitaskCount + focusCount + fragmentedCount
	if total == 0 {
		total = 1
	}

	return &ModeDistribution{
		MultitaskSessions:    multitaskCount,
		FocusSessions:        focusCount,
		FragmentedSessions:   fragmentedCount,
		MultitaskPercentage:  float64(multitaskCount) / float64(total) * 100,
		FocusPercentage:      float64(focusCount) / float64(total) * 100,
		FragmentedPercentage: float64(fragmentedCount) / float64(total) * 100,
	}, nil
}

// analyzeCommonSwitchPaths 分析常见切换路径
func (h *Handler) analyzeCommonSwitchPaths() ([]SwitchPath, error) {
	// 简化实现：分析连续的应用使用记录
	query := `
		SELECT
			app_name,
			timestamp,
			duration_ms
		FROM app_usage
		ORDER BY timestamp
		LIMIT 1000
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type appSession struct {
		appName  string
		timestamp int64
		duration int64
	}

	var sessions []appSession
	for rows.Next() {
		var s appSession
		if err := rows.Scan(&s.appName, &s.timestamp, &s.duration); err != nil {
			continue
		}
		sessions = append(sessions, s)
	}

	// 分析2步切换路径
	type pathStats struct {
		count    int
		totalDuration float64
	}
	pathMap := make(map[string]*pathStats)

	for i := 0; i < len(sessions)-1; i++ {
		path := sessions[i].appName + " → " + sessions[i+1].appName
		if pathMap[path] == nil {
			pathMap[path] = &pathStats{}
		}
		pathMap[path].count++
		pathMap[path].totalDuration += float64(sessions[i].duration) / 60000.0
	}

	// 转换为切片并排序
	var paths []SwitchPath
	for path, stats := range pathMap {
		if stats.count < 3 {
			continue // 过滤低频路径
		}

		avgDuration := stats.totalDuration / float64(stats.count)
		pattern := classifySwitchPattern(path)

		paths = append(paths, SwitchPath{
			Path:        path,
			Frequency:   stats.count,
			AvgDuration: avgDuration,
			Pattern:     pattern,
		})
	}

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Frequency > paths[j].Frequency
	})

	if len(paths) > 20 {
		paths = paths[:20]
	}

	return paths, nil
}

// analyzeSwitchingFrequency 分析切换频率
func (h *Handler) analyzeSwitchingFrequency() (*SwitchingFrequency, error) {
	frequency := &SwitchingFrequency{}

	// 按小时统计
	hourlyQuery := `
		SELECT
			CAST(strftime('%H', datetime(timestamp/1000, 'unixepoch')) AS INTEGER) as hour,
			COUNT(*) as switch_count
		FROM app_usage
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := h.db.Query(hourlyQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hourlyMap := make(map[int]int)
	for rows.Next() {
		var hour, count int
		if err := rows.Scan(&hour, &count); err != nil {
			continue
		}
		hourlyMap[hour] = count
	}

	// 计算每小时平均切换次数
	daysQuery := `SELECT COUNT(DISTINCT date) FROM app_usage`
	var totalDays int
	h.db.QueryRow(daysQuery).Scan(&totalDays)
	if totalDays == 0 {
		totalDays = 1
	}

	for hour := 0; hour < 24; hour++ {
		avgSwitches := float64(hourlyMap[hour]) / float64(totalDays)
		dominantMode := determineDominantMode(avgSwitches)

		frequency.ByHour = append(frequency.ByHour, HourlySwitch{
			Hour:         hour,
			AvgSwitches:  avgSwitches,
			DominantMode: dominantMode,
		})
	}

	// 按星期统计
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for dow := 0; dow < 7; dow++ {
		frequency.ByDayOfWeek = append(frequency.ByDayOfWeek, DaySwitch{
			DayOfWeek:    dow,
			DayName:      dayNames[dow],
			AvgSwitches:  10.0, // 简化实现
			DominantMode: "multitask",
		})
	}

	return frequency, nil
}

// analyzeTimeDistribution 分析时段分布
func (h *Handler) analyzeTimeDistribution() ([]TimeDistribution, error) {
	var distribution []TimeDistribution

	for hour := 0; hour < 24; hour++ {
		// 简化实现：根据小时估算模式分布
		multitask := 40.0
		focus := 35.0
		fragmented := 25.0

		if hour >= 9 && hour <= 17 {
			// 工作时间：更多专注和多任务
			focus = 45.0
			multitask = 40.0
			fragmented = 15.0
		} else if hour >= 22 || hour <= 6 {
			// 深夜/凌晨：更多碎片化
			fragmented = 50.0
			multitask = 30.0
			focus = 20.0
		}

		distribution = append(distribution, TimeDistribution{
			Hour:              hour,
			MultitaskPercent:  multitask,
			FocusPercent:      focus,
			FragmentedPercent: fragmented,
		})
	}

	return distribution, nil
}

// analyzeAppSwitchMatrix 分析应用切换矩阵
func (h *Handler) analyzeAppSwitchMatrix() ([]AppSwitchPair, error) {
	query := `
		SELECT
			app_name,
			timestamp
		FROM app_usage
		ORDER BY timestamp
		LIMIT 500
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type session struct {
		app       string
		timestamp int64
	}

	var sessions []session
	for rows.Next() {
		var s session
		if err := rows.Scan(&s.app, &s.timestamp); err != nil {
			continue
		}
		sessions = append(sessions, s)
	}

	// 统计切换对
	type pairStats struct {
		count      int
		totalDelay int64
	}
	pairMap := make(map[string]*pairStats)

	for i := 0; i < len(sessions)-1; i++ {
		key := sessions[i].app + "|" + sessions[i+1].app
		delay := sessions[i+1].timestamp - sessions[i].timestamp

		if pairMap[key] == nil {
			pairMap[key] = &pairStats{}
		}
		pairMap[key].count++
		pairMap[key].totalDelay += delay
	}

	// 转换为切片
	var pairs []AppSwitchPair
	for key, stats := range pairMap {
		if stats.count < 2 {
			continue
		}

		// 解析key
		fromApp := ""
		toApp := ""
		for i, part := range []rune(key) {
			if part == '|' {
				fromApp = key[:i]
				toApp = key[i+1:]
				break
			}
		}

		avgDelay := float64(stats.totalDelay) / float64(stats.count) / 1000.0

		pairs = append(pairs, AppSwitchPair{
			FromApp:  fromApp,
			ToApp:    toApp,
			Count:    stats.count,
			AvgDelay: avgDelay,
		})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	if len(pairs) > 30 {
		pairs = pairs[:30]
	}

	return pairs, nil
}

// 辅助函数
func calculateMultitaskScore(avgSwitches, avgDuration float64) float64 {
	// 多任务评分：切换频率适中(10-30次/天)，会话时长适中(5-30分钟)
	switchScore := 0.0
	if avgSwitches >= 10 && avgSwitches <= 30 {
		switchScore = 100.0 - abs(avgSwitches-20)*5
	}

	durationScore := 0.0
	if avgDuration >= 5 && avgDuration <= 30 {
		durationScore = 100.0 - abs(avgDuration-15)*3
	}

	return (switchScore + durationScore) / 2
}

func calculateFocusScore(avgDuration float64) float64 {
	// 专注评分：会话时长长(>30分钟)
	if avgDuration > 30 {
		return min(100.0, 50.0+avgDuration)
	}
	return avgDuration * 1.5
}

func calculateFragmentationScore(avgSwitches, avgDuration float64) float64 {
	// 碎片化评分：切换频率高(>30次/天)，会话时长短(<5分钟)
	switchScore := 0.0
	if avgSwitches > 30 {
		switchScore = min(100.0, avgSwitches*2)
	}

	durationScore := 0.0
	if avgDuration < 5 {
		durationScore = (5 - avgDuration) * 20
	}

	return (switchScore + durationScore) / 2
}

func determineDominantMode(avgSwitches float64) string {
	if avgSwitches > 3 {
		return "multitask"
	} else if avgSwitches < 1 {
		return "focus"
	}
	return "fragmented"
}

func classifySwitchPattern(path string) string {
	workApps := []string{"Code", "Terminal", "Word", "Excel"}
	socialApps := []string{"WeChat", "QQ", "Telegram"}

	hasWork := false
	hasSocial := false

	for _, app := range workApps {
		if contains(path, app) {
			hasWork = true
			break
		}
	}

	for _, app := range socialApps {
		if contains(path, app) {
			hasSocial = true
			break
		}
	}

	if hasWork && hasSocial {
		return "mixed"
	} else if hasWork {
		return "work"
	} else if hasSocial {
		return "social"
	}
	return "other"
}

func generateSwitchingRecommendations(pattern *AppSwitchingPattern) []string {
	recommendations := []string{}

	if pattern.OverallPattern.FragmentationScore > 60 {
		recommendations = append(recommendations,
			"碎片化使用较严重，建议设置专注时段，减少应用切换频率")
	}

	if pattern.OverallPattern.MultitaskScore > 70 {
		recommendations = append(recommendations,
			"多任务模式明显，建议使用番茄工作法，提高单任务专注度")
	}

	if pattern.OverallPattern.FocusScore < 50 {
		recommendations = append(recommendations,
			"专注度偏低，建议关闭非必要通知，创造专注工作环境")
	}

	if pattern.OverallPattern.AvgSwitchesPerDay > 50 {
		recommendations = append(recommendations,
			"每日应用切换次数过多，建议合并相似任务，减少切换成本")
	}

	recommendations = append(recommendations,
		"在高专注时段(9-11点, 14-16点)处理重要工作",
		"避免在碎片化时段(22-24点)处理复杂任务",
		"使用应用使用时长限制功能，控制娱乐应用使用",
	)

	return recommendations
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}