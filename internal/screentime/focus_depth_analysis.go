package screentime

import (
	"fmt"
	"math"
	"time"
)

// FocusDepthAnalysis represents deep focus analysis results
type FocusDepthAnalysis struct {
	OverallFocusScore    float64                `json:"overallFocusScore"`    // 0-100
	FocusLevel           string                 `json:"focusLevel"`           // excellent/good/fair/poor
	FocusSessions        []FocusSession         `json:"focusSessions"`        // Individual focus sessions
	DistractionSources   []DistractionSource    `json:"distractionSources"`   // What breaks focus
	BestFocusTimeSlots   []TimeSlot             `json:"bestFocusTimeSlots"`   // When focus is best
	FocusTrends          []FocusTrendPoint      `json:"focusTrends"`          // Daily focus trends
	FocusRecommendations []string               `json:"focusRecommendations"` // Personalized tips
	Statistics           FocusStatistics        `json:"statistics"`           // Overall stats
}

// FocusSession represents a single focus session
type FocusSession struct {
	Date              string  `json:"date"`
	StartTime         string  `json:"startTime"`
	EndTime           string  `json:"endTime"`
	Duration          int     `json:"duration"`          // seconds
	AppName           string  `json:"appName"`
	FocusScore        float64 `json:"focusScore"`        // 0-100
	InterruptionCount int     `json:"interruptionCount"` // Number of switches
	Quality           string  `json:"quality"`           // deep/moderate/shallow
}

// DistractionSource represents what breaks focus
type DistractionSource struct {
	AppName           string  `json:"appName"`
	Category          string  `json:"category"`
	InterruptionCount int     `json:"interruptionCount"`
	TotalDuration     int     `json:"totalDuration"` // seconds
	AvgDuration       int     `json:"avgDuration"`   // seconds per interruption
	ImpactScore       float64 `json:"impactScore"`   // 0-100, higher = worse
}

// TimeSlot represents a time period with focus metrics
type TimeSlot struct {
	Hour              int     `json:"hour"`              // 0-23
	AvgFocusScore     float64 `json:"avgFocusScore"`     // 0-100
	SessionCount      int     `json:"sessionCount"`      // Number of focus sessions
	AvgSessionLength  int     `json:"avgSessionLength"`  // seconds
	Recommendation    string  `json:"recommendation"`    // Best for deep work / etc
}

// FocusTrendPoint represents focus score over time
type FocusTrendPoint struct {
	Date       string  `json:"date"`
	FocusScore float64 `json:"focusScore"` // 0-100
	SessionCount int   `json:"sessionCount"`
	TotalFocusTime int `json:"totalFocusTime"` // seconds
}

// FocusStatistics represents overall focus statistics
type FocusStatistics struct {
	TotalFocusSessions    int     `json:"totalFocusSessions"`
	TotalFocusTime        int     `json:"totalFocusTime"`        // seconds
	AvgSessionLength      int     `json:"avgSessionLength"`      // seconds
	LongestSession        int     `json:"longestSession"`        // seconds
	DeepFocusSessions     int     `json:"deepFocusSessions"`     // >30 min, <3 switches
	ModerateFocusSessions int     `json:"moderateFocusSessions"` // 15-30 min
	ShallowFocusSessions  int     `json:"shallowFocusSessions"`  // <15 min
	AvgInterruptionsPerHour float64 `json:"avgInterruptionsPerHour"`
	BestFocusDay          string  `json:"bestFocusDay"`
	BestFocusScore        float64 `json:"bestFocusScore"`
}

// GetFocusDepthAnalysis performs deep focus analysis
func (h *Handler) GetFocusDepthAnalysis(startDate, endDate string) (*FocusDepthAnalysis, error) {
	// Step 1: Identify focus sessions (sessions > 10 minutes with minimal switching)
	focusSessions, err := h.identifyFocusSessions(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to identify focus sessions: %w", err)
	}

	// Step 2: Analyze distraction sources
	distractionSources, err := h.analyzeDistractionSources(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze distractions: %w", err)
	}

	// Step 3: Identify best focus time slots
	bestTimeSlots, err := h.identifyBestFocusTimeSlots(focusSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to identify best time slots: %w", err)
	}

	// Step 4: Calculate focus trends
	focusTrends, err := h.calculateFocusTrends(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate trends: %w", err)
	}

	// Step 5: Calculate statistics
	stats := h.calculateFocusStatistics(focusSessions)

	// Step 6: Calculate overall focus score
	overallScore := h.calculateOverallFocusScore(focusSessions, distractionSources, stats)

	// Step 7: Determine focus level
	focusLevel := h.determineFocusLevel(overallScore)

	// Step 8: Generate recommendations
	recommendations := h.generateFocusRecommendations(overallScore, distractionSources, bestTimeSlots, stats)

	return &FocusDepthAnalysis{
		OverallFocusScore:    overallScore,
		FocusLevel:           focusLevel,
		FocusSessions:        focusSessions,
		DistractionSources:   distractionSources,
		BestFocusTimeSlots:   bestTimeSlots,
		FocusTrends:          focusTrends,
		FocusRecommendations: recommendations,
		Statistics:           stats,
	}, nil
}

// identifyFocusSessions identifies focus sessions from phone sessions
func (h *Handler) identifyFocusSessions(startDate, endDate string) ([]FocusSession, error) {
	query := `
		SELECT
			date,
			start_time,
			end_time,
			duration_ms,
			app_name
		FROM phone_sessions
		WHERE date BETWEEN ? AND ?
		  AND duration_ms >= 600000  -- At least 10 minutes
		ORDER BY date, start_time
	`

	rows, err := h.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []FocusSession
	var currentDate string
	var sessionCount int

	for rows.Next() {
		var date, startTime, endTime, appName string
		var durationMs int64

		if err := rows.Scan(&date, &startTime, &endTime, &durationMs, &appName); err != nil {
			continue
		}

		// Count sessions per day for interruption calculation
		if date != currentDate {
			currentDate = date
			sessionCount = 0
		}
		sessionCount++

		duration := int(durationMs / 1000) // Convert to seconds

		// Calculate focus score based on duration and app switches
		focusScore := h.calculateSessionFocusScore(duration, sessionCount)

		// Determine quality
		quality := "shallow"
		if duration >= 1800 && sessionCount <= 3 { // 30+ min, <=3 switches
			quality = "deep"
		} else if duration >= 900 { // 15+ min
			quality = "moderate"
		}

		sessions = append(sessions, FocusSession{
			Date:              date,
			StartTime:         startTime,
			EndTime:           endTime,
			Duration:          duration,
			AppName:           appName,
			FocusScore:        focusScore,
			InterruptionCount: sessionCount - 1,
			Quality:           quality,
		})
	}

	return sessions, nil
}

// calculateSessionFocusScore calculates focus score for a single session
func (h *Handler) calculateSessionFocusScore(duration, interruptionCount int) float64 {
	// Base score from duration (0-50 points)
	durationScore := math.Min(float64(duration)/3600*50, 50) // Max at 1 hour

	// Penalty for interruptions (0-50 points)
	interruptionPenalty := float64(interruptionCount) * 5
	interruptionScore := math.Max(50-interruptionPenalty, 0)

	return durationScore + interruptionScore
}

// analyzeDistractionSources analyzes what breaks focus
func (h *Handler) analyzeDistractionSources(startDate, endDate string) ([]DistractionSource, error) {
	// Identify short sessions (<5 min) as distractions
	query := `
		SELECT
			app_name,
			category,
			COUNT(*) as interruption_count,
			SUM(duration_ms) as total_duration_ms,
			AVG(duration_ms) as avg_duration_ms
		FROM phone_sessions
		WHERE date BETWEEN ? AND ?
		  AND duration_ms < 300000  -- Less than 5 minutes
		GROUP BY app_name, category
		ORDER BY interruption_count DESC
		LIMIT 20
	`

	rows, err := h.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []DistractionSource
	for rows.Next() {
		var appName, category string
		var interruptionCount int
		var totalDurationMs, avgDurationMs int64

		if err := rows.Scan(&appName, &category, &interruptionCount, &totalDurationMs, &avgDurationMs); err != nil {
			continue
		}

		// Calculate impact score (frequency * brevity)
		impactScore := float64(interruptionCount) * (1.0 - float64(avgDurationMs)/300000.0) * 100
		impactScore = math.Min(impactScore, 100)

		sources = append(sources, DistractionSource{
			AppName:           appName,
			Category:          category,
			InterruptionCount: interruptionCount,
			TotalDuration:     int(totalDurationMs / 1000),
			AvgDuration:       int(avgDurationMs / 1000),
			ImpactScore:       impactScore,
		})
	}

	return sources, nil
}

// identifyBestFocusTimeSlots identifies when focus is best
func (h *Handler) identifyBestFocusTimeSlots(sessions []FocusSession) ([]TimeSlot, error) {
	// Group sessions by hour
	hourlyData := make(map[int][]FocusSession)
	for _, session := range sessions {
		t, err := time.Parse("15:04:05", session.StartTime)
		if err != nil {
			continue
		}
		hour := t.Hour()
		hourlyData[hour] = append(hourlyData[hour], session)
	}

	// Calculate metrics for each hour
	var timeSlots []TimeSlot
	for hour := 0; hour < 24; hour++ {
		sessions := hourlyData[hour]
		if len(sessions) == 0 {
			continue
		}

		var totalScore, totalDuration float64
		for _, s := range sessions {
			totalScore += s.FocusScore
			totalDuration += float64(s.Duration)
		}

		avgScore := totalScore / float64(len(sessions))
		avgLength := int(totalDuration / float64(len(sessions)))

		// Generate recommendation
		recommendation := ""
		if avgScore >= 70 && avgLength >= 1200 {
			recommendation = "最佳深度工作时段"
		} else if avgScore >= 50 {
			recommendation = "适合专注任务"
		} else {
			recommendation = "容易分心时段"
		}

		timeSlots = append(timeSlots, TimeSlot{
			Hour:             hour,
			AvgFocusScore:    avgScore,
			SessionCount:     len(sessions),
			AvgSessionLength: avgLength,
			Recommendation:   recommendation,
		})
	}

	return timeSlots, nil
}

// calculateFocusTrends calculates daily focus trends
func (h *Handler) calculateFocusTrends(startDate, endDate string) ([]FocusTrendPoint, error) {
	query := `
		SELECT
			date,
			COUNT(*) as session_count,
			SUM(duration_ms) as total_duration_ms
		FROM phone_sessions
		WHERE date BETWEEN ? AND ?
		  AND duration_ms >= 600000  -- Focus sessions only
		GROUP BY date
		ORDER BY date
	`

	rows, err := h.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []FocusTrendPoint
	for rows.Next() {
		var date string
		var sessionCount int
		var totalDurationMs int64

		if err := rows.Scan(&date, &sessionCount, &totalDurationMs); err != nil {
			continue
		}

		// Simple focus score: based on total focus time and session count
		totalFocusTime := int(totalDurationMs / 1000)
		focusScore := math.Min(float64(totalFocusTime)/14400*100, 100) // Max at 4 hours

		trends = append(trends, FocusTrendPoint{
			Date:           date,
			FocusScore:     focusScore,
			SessionCount:   sessionCount,
			TotalFocusTime: totalFocusTime,
		})
	}

	return trends, nil
}

// calculateFocusStatistics calculates overall statistics
func (h *Handler) calculateFocusStatistics(sessions []FocusSession) FocusStatistics {
	if len(sessions) == 0 {
		return FocusStatistics{}
	}

	var totalFocusTime, longestSession int
	var deepCount, moderateCount, shallowCount int
	var totalInterruptions int
	bestScore := 0.0
	bestDay := ""

	// Daily scores for finding best day
	dailyScores := make(map[string]float64)
	dailyCounts := make(map[string]int)

	for _, s := range sessions {
		totalFocusTime += s.Duration
		if s.Duration > longestSession {
			longestSession = s.Duration
		}

		switch s.Quality {
		case "deep":
			deepCount++
		case "moderate":
			moderateCount++
		case "shallow":
			shallowCount++
		}

		totalInterruptions += s.InterruptionCount

		// Track daily scores
		dailyScores[s.Date] += s.FocusScore
		dailyCounts[s.Date]++
	}

	// Find best day
	for date, score := range dailyScores {
		avgScore := score / float64(dailyCounts[date])
		if avgScore > bestScore {
			bestScore = avgScore
			bestDay = date
		}
	}

	avgSessionLength := totalFocusTime / len(sessions)
	avgInterruptionsPerHour := float64(totalInterruptions) / (float64(totalFocusTime) / 3600.0)

	return FocusStatistics{
		TotalFocusSessions:      len(sessions),
		TotalFocusTime:          totalFocusTime,
		AvgSessionLength:        avgSessionLength,
		LongestSession:          longestSession,
		DeepFocusSessions:       deepCount,
		ModerateFocusSessions:   moderateCount,
		ShallowFocusSessions:    shallowCount,
		AvgInterruptionsPerHour: avgInterruptionsPerHour,
		BestFocusDay:            bestDay,
		BestFocusScore:          bestScore,
	}
}

// calculateOverallFocusScore calculates overall focus score
func (h *Handler) calculateOverallFocusScore(sessions []FocusSession, distractions []DistractionSource, stats FocusStatistics) float64 {
	if len(sessions) == 0 {
		return 0
	}

	// Component 1: Average session focus score (0-40 points)
	var totalScore float64
	for _, s := range sessions {
		totalScore += s.FocusScore
	}
	avgSessionScore := (totalScore / float64(len(sessions))) * 0.4

	// Component 2: Deep focus ratio (0-30 points)
	deepRatio := float64(stats.DeepFocusSessions) / float64(stats.TotalFocusSessions)
	deepScore := deepRatio * 30

	// Component 3: Distraction penalty (0-30 points)
	var totalDistractionImpact float64
	for _, d := range distractions {
		totalDistractionImpact += d.ImpactScore
	}
	avgDistractionImpact := 0.0
	if len(distractions) > 0 {
		avgDistractionImpact = totalDistractionImpact / float64(len(distractions))
	}
	distractionScore := math.Max(30-(avgDistractionImpact*0.3), 0)

	return avgSessionScore + deepScore + distractionScore
}

// determineFocusLevel determines focus level from score
func (h *Handler) determineFocusLevel(score float64) string {
	if score >= 80 {
		return "excellent"
	} else if score >= 60 {
		return "good"
	} else if score >= 40 {
		return "fair"
	}
	return "poor"
}

// generateFocusRecommendations generates personalized recommendations
func (h *Handler) generateFocusRecommendations(score float64, distractions []DistractionSource, timeSlots []TimeSlot, stats FocusStatistics) []string {
	var recommendations []string

	// Recommendation based on overall score
	if score < 60 {
		recommendations = append(recommendations, "您的专注力有较大提升空间，建议尝试番茄工作法（25分钟专注+5分钟休息）")
	}

	// Recommendation based on deep focus sessions
	if stats.DeepFocusSessions < stats.TotalFocusSessions/3 {
		recommendations = append(recommendations, "深度专注时段较少，建议每天安排至少2个30分钟以上的专注时段")
	}

	// Recommendation based on distractions
	if len(distractions) > 0 && distractions[0].InterruptionCount > 20 {
		recommendations = append(recommendations, fmt.Sprintf("「%s」是您最大的干扰源，建议在专注时段关闭通知", distractions[0].AppName))
	}

	// Recommendation based on best time slots
	if len(timeSlots) > 0 {
		var bestSlot TimeSlot
		for _, slot := range timeSlots {
			if slot.AvgFocusScore > bestSlot.AvgFocusScore {
				bestSlot = slot
			}
		}
		if bestSlot.Hour > 0 {
			recommendations = append(recommendations, fmt.Sprintf("您在%d:00-%d:00专注力最佳，建议在此时段安排重要任务", bestSlot.Hour, bestSlot.Hour+1))
		}
	}

	// Recommendation based on interruptions
	if stats.AvgInterruptionsPerHour > 10 {
		recommendations = append(recommendations, "每小时切换应用过于频繁，建议使用专注模式或应用限制功能")
	}

	// Recommendation based on session length
	if stats.AvgSessionLength < 900 { // Less than 15 minutes
		recommendations = append(recommendations, "平均专注时长较短，建议逐步延长单次专注时间至20-30分钟")
	}

	return recommendations
}
