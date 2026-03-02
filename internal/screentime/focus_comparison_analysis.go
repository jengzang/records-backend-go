package screentime

import (
	"fmt"
	"math"
	"time"
)

// FocusComparisonAnalysis represents comparative focus analysis results
type FocusComparisonAnalysis struct {
	WeekdayVsWeekend    FocusWeekdayWeekendComparison `json:"weekdayVsWeekend"`    // Weekday vs weekend patterns
	WorkVsLeisure       FocusWorkLeisureComparison    `json:"workVsLeisure"`       // Work hours vs leisure hours
	MonthlyTrends       []MonthlyFocusTrend           `json:"monthlyTrends"`       // Monthly focus trends
	DeviceComparison    *FocusDeviceComparison        `json:"deviceComparison"`    // Device-specific patterns (if available)
	Insights            []ComparisonInsight           `json:"insights"`            // Key findings
	OverallComparison   OverallComparison             `json:"overallComparison"`   // Summary comparison
}

// FocusWeekdayWeekendComparison compares weekday vs weekend focus
type FocusWeekdayWeekendComparison struct {
	Weekday FocusWeekdayWeekendMetrics `json:"weekday"`
	Weekend FocusWeekdayWeekendMetrics `json:"weekend"`
	Difference FocusComparisonDifference `json:"difference"`
}

// FocusWeekdayWeekendMetrics contains focus metrics for a period
type FocusWeekdayWeekendMetrics struct {
	AvgFocusScore      float64 `json:"avgFocusScore"`      // 0-100
	AvgSessionCount    float64 `json:"avgSessionCount"`    // Per day
	AvgSessionDuration int     `json:"avgSessionDuration"` // seconds
	DeepFocusRatio     float64 `json:"deepFocusRatio"`     // 0-1
	TotalFocusTime     int     `json:"totalFocusTime"`     // seconds
	DistractionCount   int     `json:"distractionCount"`   // Total interruptions
}

// FocusWorkLeisureComparison compares work hours (9-18) vs leisure hours
type FocusWorkLeisureComparison struct {
	WorkHours    FocusWorkLeisureMetrics   `json:"workHours"`
	LeisureHours FocusWorkLeisureMetrics   `json:"leisureHours"`
	Difference   FocusComparisonDifference `json:"difference"`
}

// FocusWorkLeisureMetrics contains focus metrics for work/leisure periods
type FocusWorkLeisureMetrics struct {
	AvgFocusScore      float64 `json:"avgFocusScore"`      // 0-100
	SessionCount       int     `json:"sessionCount"`       // Total sessions
	AvgSessionDuration int     `json:"avgSessionDuration"` // seconds
	DeepFocusRatio     float64 `json:"deepFocusRatio"`     // 0-1
	TotalFocusTime     int     `json:"totalFocusTime"`     // seconds
	TopApps            []string `json:"topApps"`           // Top 3 apps
}

// MonthlyFocusTrend represents focus metrics for a month
type MonthlyFocusTrend struct {
	Month              string  `json:"month"`              // YYYY-MM
	AvgFocusScore      float64 `json:"avgFocusScore"`      // 0-100
	TotalFocusTime     int     `json:"totalFocusTime"`     // seconds
	SessionCount       int     `json:"sessionCount"`       // Total sessions
	DeepFocusRatio     float64 `json:"deepFocusRatio"`     // 0-1
	Trend              string  `json:"trend"`              // improving/declining/stable
}

// FocusDeviceComparison compares focus patterns across devices
type FocusDeviceComparison struct {
	Phone    FocusDeviceMetrics        `json:"phone"`
	Computer FocusDeviceMetrics        `json:"computer"`
	Difference FocusComparisonDifference `json:"difference"`
}

// FocusDeviceMetrics contains focus metrics for a device
type FocusDeviceMetrics struct {
	AvgFocusScore      float64 `json:"avgFocusScore"`      // 0-100
	SessionCount       int     `json:"sessionCount"`       // Total sessions
	AvgSessionDuration int     `json:"avgSessionDuration"` // seconds
	DeepFocusRatio     float64 `json:"deepFocusRatio"`     // 0-1
	TotalFocusTime     int     `json:"totalFocusTime"`     // seconds
}

// FocusComparisonDifference represents the difference between two metrics
type FocusComparisonDifference struct {
	FocusScoreDiff      float64 `json:"focusScoreDiff"`      // Percentage difference
	SessionCountDiff    float64 `json:"sessionCountDiff"`    // Percentage difference
	SessionDurationDiff float64 `json:"sessionDurationDiff"` // Percentage difference
	DeepFocusRatioDiff  float64 `json:"deepFocusRatioDiff"`  // Percentage difference
	Winner              string  `json:"winner"`              // Which side is better
}

// ComparisonInsight represents a key finding from comparison
type ComparisonInsight struct {
	Type        string  `json:"type"`        // weekday_better/weekend_better/work_better/leisure_better/improving/declining
	Title       string  `json:"title"`       // Short title
	Description string  `json:"description"` // Detailed description
	Severity    string  `json:"severity"`    // high/medium/low
	Score       float64 `json:"score"`       // Relevance score 0-100
}

// OverallComparison provides summary comparison
type OverallComparison struct {
	BestPeriod      string  `json:"bestPeriod"`      // weekday/weekend/work/leisure
	BestMonth       string  `json:"bestMonth"`       // YYYY-MM
	WorstMonth      string  `json:"worstMonth"`      // YYYY-MM
	TrendDirection  string  `json:"trendDirection"`  // improving/declining/stable
	ConsistencyScore float64 `json:"consistencyScore"` // 0-100, higher = more consistent
}

// GetFocusComparisonAnalysis performs comprehensive focus comparison analysis
func (h *Handler) GetFocusComparisonAnalysis(startDate, endDate string) (*FocusComparisonAnalysis, error) {
	// Get all focus sessions for the period
	sessions, err := h.identifyFocusSessions(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to identify focus sessions: %w", err)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no focus sessions found in the specified period")
	}

	// Perform comparisons
	weekdayVsWeekend := h.compareWeekdayVsWeekend(sessions)
	workVsLeisure := h.compareWorkVsLeisure(sessions)
	monthlyTrends := h.calculateMonthlyTrends(sessions)
	deviceComparison := h.compareDevices(sessions)

	// Generate insights
	insights := h.generateComparisonInsights(weekdayVsWeekend, workVsLeisure, monthlyTrends)

	// Calculate overall comparison
	overallComparison := h.calculateOverallComparison(weekdayVsWeekend, workVsLeisure, monthlyTrends)

	return &FocusComparisonAnalysis{
		WeekdayVsWeekend:  weekdayVsWeekend,
		WorkVsLeisure:     workVsLeisure,
		MonthlyTrends:     monthlyTrends,
		DeviceComparison:  deviceComparison,
		Insights:          insights,
		OverallComparison: overallComparison,
	}, nil
}

// compareWeekdayVsWeekend compares focus patterns between weekdays and weekends
func (h *Handler) compareWeekdayVsWeekend(sessions []FocusSession) FocusWeekdayWeekendComparison {
	var weekdaySessions, weekendSessions []FocusSession

	for _, session := range sessions {
		date, err := time.Parse("2006-01-02", session.Date)
		if err != nil {
			continue
		}

		weekday := date.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			weekendSessions = append(weekendSessions, session)
		} else {
			weekdaySessions = append(weekdaySessions, session)
		}
	}

	weekdayMetrics := h.calculatePeriodMetrics(weekdaySessions, true)
	weekendMetrics := h.calculatePeriodMetrics(weekendSessions, false)
	difference := h.calculateDifference(weekdayMetrics, weekendMetrics, "weekday", "weekend")

	return FocusWeekdayWeekendComparison{
		Weekday:    weekdayMetrics,
		Weekend:    weekendMetrics,
		Difference: difference,
	}
}

// compareWorkVsLeisure compares focus during work hours (9-18) vs leisure hours
func (h *Handler) compareWorkVsLeisure(sessions []FocusSession) FocusWorkLeisureComparison {
	var workSessions, leisureSessions []FocusSession

	for _, session := range sessions {
		startTime, err := time.Parse("15:04:05", session.StartTime)
		if err != nil {
			continue
		}

		hour := startTime.Hour()
		if hour >= 9 && hour < 18 {
			workSessions = append(workSessions, session)
		} else {
			leisureSessions = append(leisureSessions, session)
		}
	}

	workMetrics := h.calculateWorkLeisureMetrics(workSessions)
	leisureMetrics := h.calculateWorkLeisureMetrics(leisureSessions)

	// Convert to FocusWeekdayWeekendMetrics for difference calculation
	workPeriodMetrics := FocusWeekdayWeekendMetrics{
		AvgFocusScore:      workMetrics.AvgFocusScore,
		AvgSessionCount:    float64(workMetrics.SessionCount),
		AvgSessionDuration: workMetrics.AvgSessionDuration,
		DeepFocusRatio:     workMetrics.DeepFocusRatio,
		TotalFocusTime:     workMetrics.TotalFocusTime,
	}
	leisurePeriodMetrics := FocusWeekdayWeekendMetrics{
		AvgFocusScore:      leisureMetrics.AvgFocusScore,
		AvgSessionCount:    float64(leisureMetrics.SessionCount),
		AvgSessionDuration: leisureMetrics.AvgSessionDuration,
		DeepFocusRatio:     leisureMetrics.DeepFocusRatio,
		TotalFocusTime:     leisureMetrics.TotalFocusTime,
	}

	difference := h.calculateDifference(workPeriodMetrics, leisurePeriodMetrics, "work", "leisure")

	return FocusWorkLeisureComparison{
		WorkHours:    workMetrics,
		LeisureHours: leisureMetrics,
		Difference:   difference,
	}
}

// calculateMonthlyTrends calculates focus trends by month
func (h *Handler) calculateMonthlyTrends(sessions []FocusSession) []MonthlyFocusTrend {
	monthlyData := make(map[string][]FocusSession)

	for _, session := range sessions {
		date, err := time.Parse("2006-01-02", session.Date)
		if err != nil {
			continue
		}

		month := date.Format("2006-01")
		monthlyData[month] = append(monthlyData[month], session)
	}

	var trends []MonthlyFocusTrend
	var prevScore float64

	for month, monthlySessions := range monthlyData {
		metrics := h.calculatePeriodMetrics(monthlySessions, false)

		trend := "stable"
		if prevScore > 0 {
			diff := metrics.AvgFocusScore - prevScore
			if diff > 5 {
				trend = "improving"
			} else if diff < -5 {
				trend = "declining"
			}
		}

		trends = append(trends, MonthlyFocusTrend{
			Month:          month,
			AvgFocusScore:  metrics.AvgFocusScore,
			TotalFocusTime: metrics.TotalFocusTime,
			SessionCount:   int(metrics.AvgSessionCount),
			DeepFocusRatio: metrics.DeepFocusRatio,
			Trend:          trend,
		})

		prevScore = metrics.AvgFocusScore
	}

	return trends
}

// compareDevices compares focus patterns across devices (if data available)
func (h *Handler) compareDevices(sessions []FocusSession) *FocusDeviceComparison {
	// This would require device information in the sessions
	// For now, return nil as device info is not available
	return nil
}

// calculatePeriodMetrics calculates metrics for a period
func (h *Handler) calculatePeriodMetrics(sessions []FocusSession, isWeekday bool) FocusWeekdayWeekendMetrics {
	if len(sessions) == 0 {
		return FocusWeekdayWeekendMetrics{}
	}

	var totalScore, totalDuration float64
	var deepCount, totalDistractions int
	uniqueDays := make(map[string]bool)

	for _, session := range sessions {
		totalScore += session.FocusScore
		totalDuration += float64(session.Duration)
		totalDistractions += session.InterruptionCount
		uniqueDays[session.Date] = true

		if session.Quality == "deep" {
			deepCount++
		}
	}

	dayCount := len(uniqueDays)
	if dayCount == 0 {
		dayCount = 1
	}

	return FocusWeekdayWeekendMetrics{
		AvgFocusScore:      totalScore / float64(len(sessions)),
		AvgSessionCount:    float64(len(sessions)) / float64(dayCount),
		AvgSessionDuration: int(totalDuration / float64(len(sessions))),
		DeepFocusRatio:     float64(deepCount) / float64(len(sessions)),
		TotalFocusTime:     int(totalDuration),
		DistractionCount:   totalDistractions,
	}
}

// calculateWorkLeisureMetrics calculates metrics for work/leisure periods
func (h *Handler) calculateWorkLeisureMetrics(sessions []FocusSession) FocusWorkLeisureMetrics {
	if len(sessions) == 0 {
		return FocusWorkLeisureMetrics{}
	}

	var totalScore, totalDuration float64
	var deepCount int
	appCounts := make(map[string]int)

	for _, session := range sessions {
		totalScore += session.FocusScore
		totalDuration += float64(session.Duration)
		appCounts[session.AppName]++

		if session.Quality == "deep" {
			deepCount++
		}
	}

	// Get top 3 apps
	type appCount struct {
		name  string
		count int
	}
	var appList []appCount
	for name, count := range appCounts {
		appList = append(appList, appCount{name, count})
	}

	// Simple bubble sort for top 3
	for i := 0; i < len(appList)-1; i++ {
		for j := 0; j < len(appList)-i-1; j++ {
			if appList[j].count < appList[j+1].count {
				appList[j], appList[j+1] = appList[j+1], appList[j]
			}
		}
	}

	var topApps []string
	for i := 0; i < 3 && i < len(appList); i++ {
		topApps = append(topApps, appList[i].name)
	}

	return FocusWorkLeisureMetrics{
		AvgFocusScore:      totalScore / float64(len(sessions)),
		SessionCount:       len(sessions),
		AvgSessionDuration: int(totalDuration / float64(len(sessions))),
		DeepFocusRatio:     float64(deepCount) / float64(len(sessions)),
		TotalFocusTime:     int(totalDuration),
		TopApps:            topApps,
	}
}

// calculateDifference calculates percentage differences between two periods
func (h *Handler) calculateDifference(a, b FocusWeekdayWeekendMetrics, aName, bName string) FocusComparisonDifference {
	focusScoreDiff := h.percentageDiff(a.AvgFocusScore, b.AvgFocusScore)
	sessionCountDiff := h.percentageDiff(a.AvgSessionCount, b.AvgSessionCount)
	sessionDurationDiff := h.percentageDiff(float64(a.AvgSessionDuration), float64(b.AvgSessionDuration))
	deepFocusRatioDiff := h.percentageDiff(a.DeepFocusRatio, b.DeepFocusRatio)

	// Determine winner based on focus score
	winner := aName
	if b.AvgFocusScore > a.AvgFocusScore {
		winner = bName
	}

	return FocusComparisonDifference{
		FocusScoreDiff:      focusScoreDiff,
		SessionCountDiff:    sessionCountDiff,
		SessionDurationDiff: sessionDurationDiff,
		DeepFocusRatioDiff:  deepFocusRatioDiff,
		Winner:              winner,
	}
}

// percentageDiff calculates percentage difference
func (h *Handler) percentageDiff(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return ((a - b) / b) * 100
}

// generateComparisonInsights generates key insights from comparisons
func (h *Handler) generateComparisonInsights(weekdayVsWeekend FocusWeekdayWeekendComparison, workVsLeisure FocusWorkLeisureComparison, monthlyTrends []MonthlyFocusTrend) []ComparisonInsight {
	var insights []ComparisonInsight

	// Weekday vs Weekend insights
	if math.Abs(weekdayVsWeekend.Difference.FocusScoreDiff) > 10 {
		severity := "medium"
		if math.Abs(weekdayVsWeekend.Difference.FocusScoreDiff) > 20 {
			severity = "high"
		}

		insights = append(insights, ComparisonInsight{
			Type:        weekdayVsWeekend.Difference.Winner + "_better",
			Title:       fmt.Sprintf("%s专注力明显更好", map[string]string{"weekday": "工作日", "weekend": "周末"}[weekdayVsWeekend.Difference.Winner]),
			Description: fmt.Sprintf("%s的平均专注分数比%s高%.1f%%",
				map[string]string{"weekday": "工作日", "weekend": "周末"}[weekdayVsWeekend.Difference.Winner],
				map[string]string{"weekday": "周末", "weekend": "工作日"}[weekdayVsWeekend.Difference.Winner],
				math.Abs(weekdayVsWeekend.Difference.FocusScoreDiff)),
			Severity:    severity,
			Score:       math.Abs(weekdayVsWeekend.Difference.FocusScoreDiff),
		})
	}

	// Work vs Leisure insights
	if math.Abs(workVsLeisure.Difference.FocusScoreDiff) > 10 {
		severity := "medium"
		if math.Abs(workVsLeisure.Difference.FocusScoreDiff) > 20 {
			severity = "high"
		}

		insights = append(insights, ComparisonInsight{
			Type:        workVsLeisure.Difference.Winner + "_better",
			Title:       fmt.Sprintf("%s时段专注力更高", map[string]string{"work": "工作", "leisure": "休闲"}[workVsLeisure.Difference.Winner]),
			Description: fmt.Sprintf("%s时段的平均专注分数比%s时段高%.1f%%",
				map[string]string{"work": "工作", "leisure": "休闲"}[workVsLeisure.Difference.Winner],
				map[string]string{"work": "休闲", "leisure": "工作"}[workVsLeisure.Difference.Winner],
				math.Abs(workVsLeisure.Difference.FocusScoreDiff)),
			Severity:    severity,
			Score:       math.Abs(workVsLeisure.Difference.FocusScoreDiff),
		})
	}

	// Monthly trend insights
	if len(monthlyTrends) >= 2 {
		improvingCount := 0
		decliningCount := 0

		for _, trend := range monthlyTrends {
			if trend.Trend == "improving" {
				improvingCount++
			} else if trend.Trend == "declining" {
				decliningCount++
			}
		}

		if improvingCount > decliningCount && improvingCount >= 2 {
			insights = append(insights, ComparisonInsight{
				Type:        "improving",
				Title:       "专注力呈上升趋势",
				Description: fmt.Sprintf("最近%d个月中有%d个月专注力在提升", len(monthlyTrends), improvingCount),
				Severity:    "low",
				Score:       float64(improvingCount) / float64(len(monthlyTrends)) * 100,
			})
		} else if decliningCount > improvingCount && decliningCount >= 2 {
			insights = append(insights, ComparisonInsight{
				Type:        "declining",
				Title:       "专注力呈下降趋势",
				Description: fmt.Sprintf("最近%d个月中有%d个月专注力在下降", len(monthlyTrends), decliningCount),
				Severity:    "high",
				Score:       float64(decliningCount) / float64(len(monthlyTrends)) * 100,
			})
		}
	}

	return insights
}

// calculateOverallComparison calculates overall comparison summary
func (h *Handler) calculateOverallComparison(weekdayVsWeekend FocusWeekdayWeekendComparison, workVsLeisure FocusWorkLeisureComparison, monthlyTrends []MonthlyFocusTrend) OverallComparison {
	// Determine best period
	bestPeriod := "weekday"
	if weekdayVsWeekend.Weekend.AvgFocusScore > weekdayVsWeekend.Weekday.AvgFocusScore {
		bestPeriod = "weekend"
	}
	if workVsLeisure.LeisureHours.AvgFocusScore > workVsLeisure.WorkHours.AvgFocusScore {
		bestPeriod = "leisure"
	}

	// Find best and worst months
	var bestMonth, worstMonth string
	var bestScore, worstScore float64 = -1, 101

	for _, trend := range monthlyTrends {
		if trend.AvgFocusScore > bestScore {
			bestScore = trend.AvgFocusScore
			bestMonth = trend.Month
		}
		if trend.AvgFocusScore < worstScore {
			worstScore = trend.AvgFocusScore
			worstMonth = trend.Month
		}
	}

	// Determine overall trend
	trendDirection := "stable"
	if len(monthlyTrends) >= 2 {
		firstHalf := monthlyTrends[:len(monthlyTrends)/2]
		secondHalf := monthlyTrends[len(monthlyTrends)/2:]

		var firstAvg, secondAvg float64
		for _, t := range firstHalf {
			firstAvg += t.AvgFocusScore
		}
		for _, t := range secondHalf {
			secondAvg += t.AvgFocusScore
		}
		firstAvg /= float64(len(firstHalf))
		secondAvg /= float64(len(secondHalf))

		diff := secondAvg - firstAvg
		if diff > 5 {
			trendDirection = "improving"
		} else if diff < -5 {
			trendDirection = "declining"
		}
	}

	// Calculate consistency score
	consistencyScore := h.calculateConsistencyScore(monthlyTrends)

	return OverallComparison{
		BestPeriod:       bestPeriod,
		BestMonth:        bestMonth,
		WorstMonth:       worstMonth,
		TrendDirection:   trendDirection,
		ConsistencyScore: consistencyScore,
	}
}

// calculateConsistencyScore calculates how consistent focus is over time
func (h *Handler) calculateConsistencyScore(monthlyTrends []MonthlyFocusTrend) float64 {
	if len(monthlyTrends) < 2 {
		return 100
	}

	var scores []float64
	for _, trend := range monthlyTrends {
		scores = append(scores, trend.AvgFocusScore)
	}

	// Calculate standard deviation
	var sum, mean float64
	for _, score := range scores {
		sum += score
	}
	mean = sum / float64(len(scores))

	var variance float64
	for _, score := range scores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= float64(len(scores))
	stdDev := math.Sqrt(variance)

	// Convert to consistency score (lower stdDev = higher consistency)
	// Assume stdDev of 0 = 100, stdDev of 20 = 0
	consistencyScore := math.Max(0, 100-(stdDev*5))

	return consistencyScore
}
