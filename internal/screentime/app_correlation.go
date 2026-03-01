package screentime

import (
	"fmt"
	"math"
)

// AppCorrelationAnalysis 应用使用相关性分析结果
type AppCorrelationAnalysis struct {
	CorrelationPairs []AppCorrelationPair `json:"correlation_pairs"` // 应用关联对
	FrequentPatterns []FrequentPattern    `json:"frequent_patterns"` // 频繁模式
	TimeWindowStats  TimeWindowStats      `json:"time_window_stats"` // 时间窗口统计
	NetworkGraph     NetworkGraph         `json:"network_graph"`     // 关系网络图
	Insights         []string             `json:"insights"`          // 洞察
}

// AppCorrelationPair 应用关联对
type AppCorrelationPair struct {
	App1            string  `json:"app1"`             // 应用1
	App2            string  `json:"app2"`             // 应用2
	CoOccurrence    int     `json:"co_occurrence"`    // 共现次数
	CorrelationScore float64 `json:"correlation_score"` // 关联评分(0-1)
	Confidence      float64 `json:"confidence"`       // 置信度
	Lift            float64 `json:"lift"`             // 提升度
	Pattern         string  `json:"pattern"`          // 模式: sequential/concurrent/alternating
}

// FrequentPattern 频繁模式
type FrequentPattern struct {
	Apps       []string `json:"apps"`       // 应用列表
	Frequency  int      `json:"frequency"`  // 出现频率
	Support    float64  `json:"support"`    // 支持度
	AvgGap     float64  `json:"avg_gap"`    // 平均间隔(分钟)
	Typical    string   `json:"typical"`    // 典型场景
}

// TimeWindowStats 时间窗口统计
type TimeWindowStats struct {
	WindowSize      int     `json:"window_size"`      // 窗口大小(分钟)
	TotalWindows    int     `json:"total_windows"`    // 总窗口数
	AvgAppsPerWindow float64 `json:"avg_apps_per_window"` // 平均每窗口应用数
	MaxAppsPerWindow int     `json:"max_apps_per_window"` // 最大每窗口应用数
}

// NetworkGraph 关系网络图
type NetworkGraph struct {
	Nodes []NetworkNode `json:"nodes"` // 节点
	Edges []NetworkEdge `json:"edges"` // 边
}

// NetworkNode 网络节点
type NetworkNode struct {
	ID       string  `json:"id"`       // 应用ID
	Label    string  `json:"label"`    // 应用名称
	Category string  `json:"category"` // 类别
	Size     float64 `json:"size"`     // 节点大小(基于使用时长)
}

// NetworkEdge 网络边
type NetworkEdge struct {
	Source string  `json:"source"` // 源节点
	Target string  `json:"target"` // 目标节点
	Weight float64 `json:"weight"` // 权重(关联强度)
}

// GetAppCorrelationAnalysis 获取应用使用相关性分析
func (h *Handler) GetAppCorrelationAnalysis() (*AppCorrelationAnalysis, error) {
	windowSize := 30 // 30分钟时间窗口

	// 1. 分析应用共现
	pairs, err := h.analyzeAppCoOccurrence(windowSize)
	if err != nil {
		return nil, err
	}

	// 2. 识别频繁模式
	patterns, err := h.identifyFrequentPatterns(windowSize)
	if err != nil {
		return nil, err
	}

	// 3. 计算时间窗口统计
	windowStats, err := h.calculateTimeWindowStats(windowSize)
	if err != nil {
		return nil, err
	}

	// 4. 构建网络图
	graph := h.buildNetworkGraph(pairs)

	// 5. 生成洞察
	insights := h.generateCorrelationInsights(pairs, patterns)

	return &AppCorrelationAnalysis{
		CorrelationPairs: pairs,
		FrequentPatterns: patterns,
		TimeWindowStats:  windowStats,
		NetworkGraph:     graph,
		Insights:         insights,
	}, nil
}

// analyzeAppCoOccurrence 分析应用共现
func (h *Handler) analyzeAppCoOccurrence(windowSize int) ([]AppCorrelationPair, error) {
	// 查询会话数据，按时间窗口分组
	query := `
	SELECT
		s1.package_id as app1,
		s1.app_name as app1_name,
		s2.package_id as app2,
		s2.app_name as app2_name,
		COUNT(*) as co_occurrence
	FROM screentime_sessions s1
	JOIN screentime_sessions s2
		ON s1.date = s2.date
		AND s1.package_id < s2.package_id
		AND ABS(
			(CAST(substr(s1.start_time, 1, 2) AS INTEGER) * 60 + CAST(substr(s1.start_time, 4, 2) AS INTEGER)) -
			(CAST(substr(s2.start_time, 1, 2) AS INTEGER) * 60 + CAST(substr(s2.start_time, 4, 2) AS INTEGER))
		) <= ?
	GROUP BY s1.package_id, s2.package_id
	HAVING co_occurrence >= 3
	ORDER BY co_occurrence DESC
	LIMIT 50
	`

	rows, err := h.db.Query(query, windowSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取每个应用的总出现次数
	appCounts := make(map[string]int)
	countQuery := `
	SELECT package_id, COUNT(*) as count
	FROM screentime_sessions
	GROUP BY package_id
	`
	countRows, err := h.db.Query(countQuery)
	if err != nil {
		return nil, err
	}
	defer countRows.Close()

	for countRows.Next() {
		var packageID string
		var count int
		if err := countRows.Scan(&packageID, &count); err != nil {
			continue
		}
		appCounts[packageID] = count
	}

	totalSessions := 0
	for _, count := range appCounts {
		totalSessions += count
	}

	var pairs []AppCorrelationPair
	for rows.Next() {
		var app1, app1Name, app2, app2Name string
		var coOccurrence int

		if err := rows.Scan(&app1, &app1Name, &app2, &app2Name, &coOccurrence); err != nil {
			continue
		}

		// 计算关联指标
		count1 := float64(appCounts[app1])
		count2 := float64(appCounts[app2])
		coOcc := float64(coOccurrence)
		total := float64(totalSessions)

		// 置信度: P(B|A) = P(A∩B) / P(A)
		confidence := coOcc / count1

		// 提升度: Lift = P(A∩B) / (P(A) * P(B))
		lift := (coOcc / total) / ((count1 / total) * (count2 / total))

		// 关联评分 (0-1)
		correlationScore := math.Min(confidence*lift/2, 1.0)

		// 判断模式
		pattern := h.determineCorrelationPattern(app1, app2)

		pairs = append(pairs, AppCorrelationPair{
			App1:            app1Name,
			App2:            app2Name,
			CoOccurrence:    coOccurrence,
			CorrelationScore: correlationScore,
			Confidence:      confidence,
			Lift:            lift,
			Pattern:         pattern,
		})
	}

	return pairs, nil
}

// identifyFrequentPatterns 识别频繁模式
func (h *Handler) identifyFrequentPatterns(windowSize int) ([]FrequentPattern, error) {
	// 查询频繁的3应用组合
	query := `
	WITH app_sequences AS (
		SELECT
			date,
			package_id,
			app_name,
			start_time,
			CAST(substr(start_time, 1, 2) AS INTEGER) * 60 + CAST(substr(start_time, 4, 2) AS INTEGER) as start_minutes
		FROM screentime_sessions
		ORDER BY date, start_minutes
	),
	app_groups AS (
		SELECT
			a1.app_name as app1,
			a2.app_name as app2,
			a3.app_name as app3,
			COUNT(*) as frequency
		FROM app_sequences a1
		JOIN app_sequences a2
			ON a1.date = a2.date
			AND a2.start_minutes BETWEEN a1.start_minutes AND a1.start_minutes + ?
			AND a1.package_id != a2.package_id
		JOIN app_sequences a3
			ON a2.date = a3.date
			AND a3.start_minutes BETWEEN a2.start_minutes AND a2.start_minutes + ?
			AND a2.package_id != a3.package_id
			AND a1.package_id != a3.package_id
		GROUP BY a1.app_name, a2.app_name, a3.app_name
		HAVING frequency >= 2
		ORDER BY frequency DESC
		LIMIT 20
	)
	SELECT * FROM app_groups
	`

	rows, err := h.db.Query(query, windowSize, windowSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取总会话数
	var totalSessions int
	h.db.QueryRow("SELECT COUNT(*) FROM screentime_sessions").Scan(&totalSessions)

	var patterns []FrequentPattern
	for rows.Next() {
		var app1, app2, app3 string
		var frequency int

		if err := rows.Scan(&app1, &app2, &app3, &frequency); err != nil {
			continue
		}

		support := float64(frequency) / float64(totalSessions)
		typical := h.determineTypicalScenario([]string{app1, app2, app3})

		patterns = append(patterns, FrequentPattern{
			Apps:      []string{app1, app2, app3},
			Frequency: frequency,
			Support:   support,
			AvgGap:    float64(windowSize) / 2, // 简化估算
			Typical:   typical,
		})
	}

	return patterns, nil
}

// calculateTimeWindowStats 计算时间窗口统计
func (h *Handler) calculateTimeWindowStats(windowSize int) (TimeWindowStats, error) {
	// 统计每个时间窗口的应用数
	query := `
	WITH time_windows AS (
		SELECT
			date,
			(CAST(substr(start_time, 1, 2) AS INTEGER) * 60 + CAST(substr(start_time, 4, 2) AS INTEGER)) / ? as window_id,
			COUNT(DISTINCT package_id) as app_count
		FROM screentime_sessions
		GROUP BY date, window_id
	)
	SELECT
		COUNT(*) as total_windows,
		AVG(app_count) as avg_apps,
		MAX(app_count) as max_apps
	FROM time_windows
	`

	var stats TimeWindowStats
	stats.WindowSize = windowSize

	err := h.db.QueryRow(query, windowSize).Scan(
		&stats.TotalWindows,
		&stats.AvgAppsPerWindow,
		&stats.MaxAppsPerWindow,
	)

	return stats, err
}

// buildNetworkGraph 构建网络图
func (h *Handler) buildNetworkGraph(pairs []AppCorrelationPair) NetworkGraph {
	graph := NetworkGraph{
		Nodes: []NetworkNode{},
		Edges: []NetworkEdge{},
	}

	// 收集所有应用
	appSet := make(map[string]bool)
	for _, pair := range pairs {
		appSet[pair.App1] = true
		appSet[pair.App2] = true
	}

	// 获取应用使用时长和类别
	mapper := NewAppCategoryMapper()
	for app := range appSet {
		var totalDuration int64
		h.db.QueryRow(`
			SELECT COALESCE(SUM(duration_ms), 0)
			FROM screentime_daily
			WHERE app_name = ?
		`, app).Scan(&totalDuration)

		size := math.Log10(float64(totalDuration)/1000 + 1) * 10 // 对数缩放

		graph.Nodes = append(graph.Nodes, NetworkNode{
			ID:       app,
			Label:    app,
			Category: mapper.GetCategory(app),
			Size:     size,
		})
	}

	// 添加边
	for _, pair := range pairs {
		if pair.CorrelationScore > 0.3 { // 只显示强关联
			graph.Edges = append(graph.Edges, NetworkEdge{
				Source: pair.App1,
				Target: pair.App2,
				Weight: pair.CorrelationScore,
			})
		}
	}

	return graph
}

// determineCorrelationPattern 判断关联模式
func (h *Handler) determineCorrelationPattern(app1, app2 string) string {
	// 查询两个应用的时间关系
	query := `
	SELECT
		AVG(ABS(
			(CAST(substr(s1.start_time, 1, 2) AS INTEGER) * 60 + CAST(substr(s1.start_time, 4, 2) AS INTEGER)) -
			(CAST(substr(s2.start_time, 1, 2) AS INTEGER) * 60 + CAST(substr(s2.start_time, 4, 2) AS INTEGER))
		)) as avg_gap
	FROM screentime_sessions s1
	JOIN screentime_sessions s2
		ON s1.date = s2.date
		AND s1.package_id = (SELECT package_id FROM screentime_apps WHERE app_name = ? LIMIT 1)
		AND s2.package_id = (SELECT package_id FROM screentime_apps WHERE app_name = ? LIMIT 1)
	`

	var avgGap float64
	err := h.db.QueryRow(query, app1, app2).Scan(&avgGap)
	if err != nil {
		return "concurrent"
	}

	if avgGap < 5 {
		return "concurrent" // 并发使用
	} else if avgGap < 30 {
		return "sequential" // 顺序使用
	} else {
		return "alternating" // 交替使用
	}
}

// determineTypicalScenario 判断典型场景
func (h *Handler) determineTypicalScenario(apps []string) string {
	mapper := NewAppCategoryMapper()
	categories := make(map[string]int)

	for _, app := range apps {
		category := mapper.GetCategory(app)
		categories[category]++
	}

	// 判断主要类别
	if categories["Social"] >= 2 || categories["Communication"] >= 2 {
		return "社交沟通场景"
	} else if categories["Entertainment"] >= 2 || categories["Video"] >= 2 {
		return "娱乐休闲场景"
	} else if categories["Productivity"] >= 2 || categories["Development"] >= 2 {
		return "工作学习场景"
	} else if categories["Shopping"] >= 1 && categories["Payment"] >= 1 {
		return "购物支付场景"
	} else {
		return "混合使用场景"
	}
}

// generateCorrelationInsights 生成关联洞察
func (h *Handler) generateCorrelationInsights(pairs []AppCorrelationPair, patterns []FrequentPattern) []string {
	var insights []string

	// 1. 最强关联
	if len(pairs) > 0 {
		strongest := pairs[0]
		insights = append(insights,
			"最强关联: "+strongest.App1+" 和 "+strongest.App2+
			" 经常一起使用，关联评分 "+formatFloat(strongest.CorrelationScore))
	}

	// 2. 高提升度关联
	for _, pair := range pairs {
		if pair.Lift > 2.0 {
			insights = append(insights,
				pair.App1+" 和 "+pair.App2+" 的关联提升度达 "+
				formatFloat(pair.Lift)+"，远高于随机组合")
			break
		}
	}

	// 3. 频繁模式
	if len(patterns) > 0 {
		pattern := patterns[0]
		insights = append(insights,
			"最频繁的使用模式: "+pattern.Apps[0]+" → "+
			pattern.Apps[1]+" → "+pattern.Apps[2]+
			" ("+pattern.Typical+")")
	}

	// 4. 模式类型统计
	patternTypes := make(map[string]int)
	for _, pair := range pairs {
		patternTypes[pair.Pattern]++
	}
	if patternTypes["concurrent"] > len(pairs)/2 {
		insights = append(insights, "你倾向于同时使用多个应用(多任务模式)")
	} else if patternTypes["sequential"] > len(pairs)/2 {
		insights = append(insights, "你倾向于按顺序切换应用(专注模式)")
	}

	// 5. 场景识别
	scenarioCounts := make(map[string]int)
	for _, pattern := range patterns {
		scenarioCounts[pattern.Typical]++
	}
	var maxScenario string
	var maxCount int
	for scenario, count := range scenarioCounts {
		if count > maxCount {
			maxScenario = scenario
			maxCount = count
		}
	}
	if maxScenario != "" {
		insights = append(insights, "你的主要使用场景是: "+maxScenario)
	}

	return insights
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
