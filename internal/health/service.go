package health

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jengzang/records-backend-go/internal/health/analysis"
)

// Service handles business logic for health data
type Service struct {
	repo              *Repository
	hrAnalyzer        *analysis.HeartRateAnalyzer
	patternAnalyzer   *analysis.PatternAnalyzer
	healthScoreCalc   *analysis.HealthScoreCalculator
}

// NewService creates a new health service
func NewService(db *sql.DB) *Service {
	return &Service{
		repo:            NewRepository(db),
		hrAnalyzer:      analysis.NewHeartRateAnalyzer(db),
		patternAnalyzer: analysis.NewPatternAnalyzer(db),
		healthScoreCalc: analysis.NewHealthScoreCalculator(db),
	}
}

// GetSummary retrieves overall health data summary
func (s *Service) GetSummary() (*HealthSummary, error) {
	return s.repo.GetSummary()
}

// GetRecords retrieves health records with filters
func (s *Service) GetRecords(filter RecordFilter) ([]HealthRecord, error) {
	// Set default limit if not specified
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	// Validate limit
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	return s.repo.GetRecords(filter)
}

// GetRecordsByType retrieves health records for a specific type
func (s *Service) GetRecordsByType(recordType string, startDate, endDate time.Time, limit int) ([]HealthRecord, error) {
	filter := RecordFilter{
		Type:      recordType,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
	}
	return s.repo.GetRecords(filter)
}

// GetWorkouts retrieves workouts with filters
func (s *Service) GetWorkouts(filter WorkoutFilter) ([]Workout, error) {
	// Set default limit if not specified
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Validate limit
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	return s.repo.GetWorkouts(filter)
}

// GetWorkoutByID retrieves a single workout by ID
func (s *Service) GetWorkoutByID(id int) (*Workout, error) {
	return s.repo.GetWorkoutByID(id)
}

// GetWorkoutRoute retrieves GPS points for a workout
func (s *Service) GetWorkoutRoute(workoutID int) ([]WorkoutRoute, error) {
	// First verify the workout exists
	_, err := s.repo.GetWorkoutByID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("workout not found: %w", err)
	}

	return s.repo.GetWorkoutRoute(workoutID)
}

// GetDailyStatistics retrieves daily statistics for a metric
func (s *Service) GetDailyStatistics(metricType string, startDate, endDate time.Time) ([]HealthStatistics, error) {
	// Validate date range
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start date must be before end date")
	}

	// Limit date range to 1 year
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil, fmt.Errorf("date range cannot exceed 1 year")
	}

	return s.repo.GetDailyStatistics(metricType, startDate, endDate)
}

// GetWeeklyStatistics retrieves weekly statistics for a metric
func (s *Service) GetWeeklyStatistics(metricType string, startDate, endDate time.Time) ([]HealthStatistics, error) {
	// TODO: Implement weekly aggregation
	// For now, return daily statistics
	return s.repo.GetDailyStatistics(metricType, startDate, endDate)
}

// GetMonthlyStatistics retrieves monthly statistics for a metric
func (s *Service) GetMonthlyStatistics(metricType string, startDate, endDate time.Time) ([]HealthStatistics, error) {
	// TODO: Implement monthly aggregation
	// For now, return daily statistics
	return s.repo.GetDailyStatistics(metricType, startDate, endDate)
}

// CalculateHealthScore calculates an overall health score
func (s *Service) CalculateHealthScore() (float64, error) {
	// TODO: Implement health score calculation
	// This would consider:
	// - Daily step count
	// - Exercise frequency
	// - Sleep quality
	// - Heart rate variability
	// - Weight trends
	return 0.0, fmt.Errorf("not implemented")
}

// GetActivityPatterns analyzes activity patterns
func (s *Service) GetActivityPatterns() (map[string]interface{}, error) {
	// TODO: Implement activity pattern analysis
	// This would identify:
	// - Peak activity times
	// - Weekday vs weekend patterns
	// - Exercise consistency
	return nil, fmt.Errorf("not implemented")
}

// GetSleepAnalysis retrieves sleep analysis data
func (s *Service) GetSleepAnalysis(startDate, endDate time.Time) ([]SleepAnalysis, error) {
	// TODO: Implement sleep analysis query
	return nil, fmt.Errorf("not implemented")
}

// GetTrends analyzes trends for a specific metric
func (s *Service) GetTrends(metricType string, period string) (map[string]interface{}, error) {
	// TODO: Implement trend analysis
	// This would calculate:
	// - Moving averages
	// - Trend direction (increasing/decreasing)
	// - Rate of change
	// - Predictions
	return nil, fmt.Errorf("not implemented")
}

// Analysis methods

// GetHeartRateZones retrieves heart rate zone distribution
func (s *Service) GetHeartRateZones(startDate, endDate time.Time) (*analysis.HeartRateZones, error) {
	return s.hrAnalyzer.GetHeartRateZones(startDate, endDate)
}

// GetHeartRateAnomalies detects anomalous heart rate readings
func (s *Service) GetHeartRateAnomalies(startDate, endDate time.Time) ([]analysis.Anomaly, error) {
	return s.hrAnalyzer.DetectAnomalies(startDate, endDate)
}

// GetRestingHeartRate retrieves daily resting heart rate
func (s *Service) GetRestingHeartRate(startDate, endDate time.Time) ([]analysis.RestingHR, error) {
	return s.hrAnalyzer.GetRestingHeartRate(startDate, endDate)
}

// GetHeartRateVariability retrieves HRV metrics
func (s *Service) GetHeartRateVariability(startDate, endDate time.Time) (*analysis.HRVMetrics, error) {
	return s.hrAnalyzer.GetHeartRateVariability(startDate, endDate)
}

// GetDailyActivityPattern retrieves 24-hour activity pattern
func (s *Service) GetDailyActivityPattern() (*analysis.DailyPattern, error) {
	return s.patternAnalyzer.GetDailyPattern()
}

// GetWeeklyActivityPattern retrieves weekly activity pattern
func (s *Service) GetWeeklyActivityPattern() (*analysis.WeeklyPattern, error) {
	return s.patternAnalyzer.GetWeeklyPattern()
}

// GetActivityScore calculates daily activity score
func (s *Service) GetActivityScore(date time.Time) (float64, error) {
	return s.patternAnalyzer.GetActivityScore(date)
}

// GetHealthScoreForDate calculates health score for a specific date
func (s *Service) GetHealthScoreForDate(date time.Time) (*analysis.HealthScore, error) {
	return s.healthScoreCalc.CalculateHealthScore(date)
}

// GetHealthScoreTrend retrieves health score trend
func (s *Service) GetHealthScoreTrend(startDate, endDate time.Time) ([]analysis.HealthScorePoint, error) {
	return s.healthScoreCalc.GetHealthScoreTrend(startDate, endDate)
}
