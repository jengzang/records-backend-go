package efficiency

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// Service handles business logic for efficiency analysis
type Service struct {
	repo       *Repository
	keyboardDB *sql.DB
	screentimeDB *sql.DB
	healthDB   *sql.DB
}

// NewService creates a new efficiency service
func NewService(healthDB, keyboardDB, screentimeDB *sql.DB) *Service {
	return &Service{
		repo:         NewRepository(healthDB),
		keyboardDB:   keyboardDB,
		screentimeDB: screentimeDB,
		healthDB:     healthDB,
	}
}

// GetHourlyCurve retrieves hourly efficiency curve for date range
func (s *Service) GetHourlyCurve(startDate, endDate string) (*EfficiencyCurveResponse, error) {
	scores, err := s.repo.GetHourlyScores(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	var totalEfficiency, maxEfficiency, minEfficiency, totalCompleteness float64
	minEfficiency = 100.0
	totalHours := len(scores)

	for _, score := range scores {
		totalEfficiency += score.EfficiencyScore
		totalCompleteness += score.DataCompleteness
		if score.EfficiencyScore > maxEfficiency {
			maxEfficiency = score.EfficiencyScore
		}
		if score.EfficiencyScore < minEfficiency {
			minEfficiency = score.EfficiencyScore
		}
	}

	response := &EfficiencyCurveResponse{
		Scores: scores,
	}
	response.Stats.TotalHours = totalHours
	if totalHours > 0 {
		response.Stats.AvgEfficiency = totalEfficiency / float64(totalHours)
		response.Stats.DataCompleteness = totalCompleteness / float64(totalHours)
	}
	response.Stats.MaxEfficiency = maxEfficiency
	response.Stats.MinEfficiency = minEfficiency

	return response, nil
}

// GetProfile retrieves efficiency curve profile
func (s *Service) GetProfile(profileType string) (*EfficiencyCurveProfile, error) {
	return s.repo.GetProfile(profileType)
}

// GetComparison retrieves workday vs weekend comparison
func (s *Service) GetComparison() (*ProfileComparisonResponse, error) {
	workday, err := s.repo.GetProfile("workday")
	if err != nil {
		return nil, err
	}
	if workday == nil {
		return nil, fmt.Errorf("workday profile not found")
	}

	weekend, err := s.repo.GetProfile("weekend")
	if err != nil {
		return nil, err
	}
	if weekend == nil {
		return nil, fmt.Errorf("weekend profile not found")
	}

	response := &ProfileComparisonResponse{
		Workday: *workday,
		Weekend: *weekend,
	}

	// Calculate differences
	response.Diff.AvgDiff = workday.AvgEfficiency - weekend.AvgEfficiency
	response.Diff.PeakHourDiff = workday.PeakHour - weekend.PeakHour

	for i := 0; i < 24; i++ {
		response.Diff.HourlyDiff[i] = workday.HourlyCurve[i] - weekend.HourlyCurve[i]
	}

	// Generate interpretation
	if response.Diff.AvgDiff > 10 {
		response.Diff.Interpretation = "工作日效率显著高于周末，建议保持工作日的良好习惯"
	} else if response.Diff.AvgDiff < -10 {
		response.Diff.Interpretation = "周末效率高于工作日，可能工作日压力过大或时间管理需要优化"
	} else {
		response.Diff.Interpretation = "工作日与周末效率相近，生活节奏较为稳定"
	}

	return response, nil
}

// GetInsights retrieves active efficiency insights
func (s *Service) GetInsights() ([]EfficiencyInsight, error) {
	return s.repo.GetInsights()
}

// AnalyzeEfficiency triggers efficiency analysis for a date range
// This will aggregate data from keyboard, screentime, and health modules
func (s *Service) AnalyzeEfficiency(startDate, endDate string) error {
	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	// Iterate through each day
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		// Analyze each hour of the day
		for hour := 0; hour < 24; hour++ {
			score, err := s.analyzeHour(dateStr, hour)
			if err != nil {
				return fmt.Errorf("failed to analyze %s hour %d: %w", dateStr, hour, err)
			}

			// Save score
			if err := s.repo.SaveHourlyScore(score); err != nil {
				return fmt.Errorf("failed to save score for %s hour %d: %w", dateStr, hour, err)
			}
		}
	}

	// Generate profiles
	if err := s.generateProfiles(startDate, endDate); err != nil {
		return fmt.Errorf("failed to generate profiles: %w", err)
	}

	return nil
}

// analyzeHour analyzes efficiency for a specific hour
func (s *Service) analyzeHour(date string, hour int) (*HourlyEfficiencyScore, error) {
	score := &HourlyEfficiencyScore{
		Date: date,
		Hour: hour,
	}

	// Fetch keyboard data
	keyboardData, err := s.fetchKeyboardData(date, hour)
	if err == nil && keyboardData != nil {
		score.TypingSpeed = &keyboardData.TypingSpeed
		score.TypingSpeedNormalized = &keyboardData.TypingSpeedNormalized
		score.HasKeyboardData = true
	}

	// Fetch screentime data
	screentimeData, err := s.fetchScreenTimeData(date, hour)
	if err == nil && screentimeData != nil {
		score.WorkAppRatio = &screentimeData.WorkAppRatio
		score.EntertainmentAppRatio = &screentimeData.EntertainmentAppRatio
		score.FocusSessionCount = &screentimeData.FocusSessionCount
		score.AppSwitchFrequency = &screentimeData.AppSwitchFrequency
		score.WorkAppRatioNormalized = &screentimeData.WorkAppRatioNormalized
		score.FocusNormalized = &screentimeData.FocusNormalized
		score.HasScreenTimeData = true
	}

	// Fetch health data
	healthData, err := s.fetchHealthData(date, hour)
	if err == nil && healthData != nil {
		score.AvgHeartRate = &healthData.AvgHeartRate
		score.HeartRateVariability = &healthData.HeartRateVariability
		score.StepCount = &healthData.StepCount
		score.HRVNormalized = &healthData.HRVNormalized
		score.ActivityNormalized = &healthData.ActivityNormalized
		score.HasHealthData = true
	}

	// Calculate composite efficiency score
	score.EfficiencyScore = s.calculateEfficiencyScore(score)

	// Calculate data completeness
	completeness := 0.0
	if score.HasKeyboardData {
		completeness += 0.33
	}
	if score.HasScreenTimeData {
		completeness += 0.33
	}
	if score.HasHealthData {
		completeness += 0.34
	}
	score.DataCompleteness = completeness

	return score, nil
}

// calculateEfficiencyScore calculates weighted efficiency score
// Weights: typing_speed(30%) + work_app_ratio(20%) + hrv(20%) + focus(15%) + activity(15%)
func (s *Service) calculateEfficiencyScore(score *HourlyEfficiencyScore) float64 {
	var totalScore, totalWeight float64

	// Typing speed (30%)
	if score.TypingSpeedNormalized != nil {
		totalScore += *score.TypingSpeedNormalized * 0.30
		totalWeight += 0.30
	}

	// Work app ratio (20%)
	if score.WorkAppRatioNormalized != nil {
		totalScore += *score.WorkAppRatioNormalized * 0.20
		totalWeight += 0.20
	}

	// HRV (20%)
	if score.HRVNormalized != nil {
		totalScore += *score.HRVNormalized * 0.20
		totalWeight += 0.20
	}

	// Focus (15%)
	if score.FocusNormalized != nil {
		totalScore += *score.FocusNormalized * 0.15
		totalWeight += 0.15
	}

	// Activity (15%)
	if score.ActivityNormalized != nil {
		totalScore += *score.ActivityNormalized * 0.15
		totalWeight += 0.15
	}

	// Normalize to 0-100 based on available data
	if totalWeight > 0 {
		return totalScore / totalWeight * 100
	}

	return 0
}

// generateProfiles generates workday and weekend profiles
func (s *Service) generateProfiles(startDate, endDate string) error {
	// Fetch all scores
	scores, err := s.repo.GetHourlyScores(startDate, endDate)
	if err != nil {
		return err
	}

	// Separate workday and weekend scores
	workdayScores := make(map[int][]float64) // hour -> []scores
	weekendScores := make(map[int][]float64)

	for _, score := range scores {
		date, _ := time.Parse("2006-01-02", score.Date)
		weekday := date.Weekday()

		if weekday == time.Saturday || weekday == time.Sunday {
			weekendScores[score.Hour] = append(weekendScores[score.Hour], score.EfficiencyScore)
		} else {
			workdayScores[score.Hour] = append(workdayScores[score.Hour], score.EfficiencyScore)
		}
	}

	// Generate workday profile
	workdayProfile := s.buildProfile("workday", workdayScores, startDate, endDate)
	if err := s.repo.SaveProfile(workdayProfile); err != nil {
		return err
	}

	// Generate weekend profile
	weekendProfile := s.buildProfile("weekend", weekendScores, startDate, endDate)
	if err := s.repo.SaveProfile(weekendProfile); err != nil {
		return err
	}

	return nil
}

// buildProfile builds an efficiency curve profile from hourly scores
func (s *Service) buildProfile(profileType string, hourlyScores map[int][]float64, startDate, endDate string) *EfficiencyCurveProfile {
	profile := &EfficiencyCurveProfile{
		ProfileType: profileType,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Calculate average for each hour
	var totalEfficiency float64
	var totalSamples int
	var maxScore float64
	var minScore float64 = 100.0

	for hour := 0; hour < 24; hour++ {
		scores := hourlyScores[hour]
		if len(scores) > 0 {
			sum := 0.0
			for _, score := range scores {
				sum += score
			}
			avg := sum / float64(len(scores))
			profile.HourlyCurve[hour] = avg
			totalEfficiency += avg
			totalSamples += len(scores)

			if avg > maxScore {
				maxScore = avg
				profile.PeakHour = hour
			}
			if avg < minScore {
				minScore = avg
				profile.LowHour = hour
			}
		}
	}

	profile.PeakScore = maxScore
	profile.LowScore = minScore
	profile.TotalSamples = totalSamples

	if totalSamples > 0 {
		profile.AvgEfficiency = totalEfficiency / 24.0
	}

	// Calculate standard deviation
	var variance float64
	for hour := 0; hour < 24; hour++ {
		diff := profile.HourlyCurve[hour] - profile.AvgEfficiency
		variance += diff * diff
	}
	profile.StdEfficiency = math.Sqrt(variance / 24.0)

	// Detect peak period
	profile.PeakStartHour, profile.PeakEndHour = s.detectPeakPeriod(profile.HourlyCurve[:])

	// Classify chronotype
	profile.Chronotype, profile.ChronotypeConfidence = s.classifyChronotype(profile.HourlyCurve[:])

	return profile
}

// detectPeakPeriod detects continuous peak efficiency period
func (s *Service) detectPeakPeriod(curve []float64) (int, int) {
	threshold := 0.0
	for _, v := range curve {
		threshold += v
	}
	threshold = threshold / float64(len(curve)) * 1.1 // 10% above average

	start, end := -1, -1
	maxDuration := 0
	currentStart := -1

	for i, score := range curve {
		if score >= threshold {
			if currentStart == -1 {
				currentStart = i
			}
		} else {
			if currentStart != -1 {
				duration := i - currentStart
				if duration > maxDuration {
					maxDuration = duration
					start = currentStart
					end = i - 1
				}
				currentStart = -1
			}
		}
	}

	// Check if peak period extends to end of day
	if currentStart != -1 {
		duration := len(curve) - currentStart
		if duration > maxDuration {
			start = currentStart
			end = len(curve) - 1
		}
	}

	return start, end
}

// classifyChronotype classifies user's chronotype based on efficiency curve
func (s *Service) classifyChronotype(curve []float64) (string, float64) {
	// Calculate average efficiency for different periods
	morningAvg := (curve[6] + curve[7] + curve[8] + curve[9] + curve[10]) / 5.0  // 6-10am
	eveningAvg := (curve[20] + curve[21] + curve[22] + curve[23]) / 4.0          // 8pm-12am
	middayAvg := (curve[10] + curve[11] + curve[12] + curve[13] + curve[14] + curve[15] + curve[16] + curve[17] + curve[18]) / 9.0 // 10am-6pm

	// Determine chronotype
	if morningAvg > eveningAvg+10 && morningAvg > middayAvg {
		confidence := math.Min((morningAvg-eveningAvg)/morningAvg, 1.0)
		return "morning", confidence
	} else if eveningAvg > morningAvg+10 && eveningAvg > middayAvg {
		confidence := math.Min((eveningAvg-morningAvg)/eveningAvg, 1.0)
		return "evening", confidence
	} else {
		confidence := 1.0 - math.Abs(morningAvg-eveningAvg)/math.Max(morningAvg, eveningAvg)
		return "intermediate", confidence
	}
}

// Placeholder methods for data fetching (to be implemented)
type KeyboardMetrics struct {
	TypingSpeed           float64
	TypingSpeedNormalized float64
}

type ScreenTimeMetrics struct {
	WorkAppRatio              float64
	EntertainmentAppRatio     float64
	FocusSessionCount         int
	AppSwitchFrequency        float64
	WorkAppRatioNormalized    float64
	FocusNormalized           float64
}

type HealthMetrics struct {
	AvgHeartRate          float64
	HeartRateVariability  float64
	StepCount             int
	HRVNormalized         float64
	ActivityNormalized    float64
}

