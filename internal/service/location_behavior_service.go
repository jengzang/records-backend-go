package service

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// LocationBehaviorService handles business logic for location behavior analysis
type LocationBehaviorService struct {
	tracksDB     *sql.DB
	keyboardDB   *sql.DB
	screentimeDB *sql.DB
	healthDB     *sql.DB
	repo         *repository.LocationBehaviorRepository
}

// NewLocationBehaviorService creates a new service
func NewLocationBehaviorService(tracksDB, keyboardDB, screentimeDB, healthDB *sql.DB) *LocationBehaviorService {
	return &LocationBehaviorService{
		tracksDB:     tracksDB,
		keyboardDB:   keyboardDB,
		screentimeDB: screentimeDB,
		healthDB:     healthDB,
		repo:         repository.NewLocationBehaviorRepository(tracksDB),
	}
}

// AnalyzeLocations performs complete location behavior analysis
func (s *LocationBehaviorService) AnalyzeLocations() error {
	// Step 1: Cluster locations from stay_segments using GeoHash6
	locations, err := s.clusterLocations()
	if err != nil {
		return fmt.Errorf("failed to cluster locations: %w", err)
	}

	// Step 2: For each location, collect behavior data from all visits
	for i := range locations {
		loc := &locations[i]

		// Save location first to get ID
		if err := s.repo.SaveLocation(loc); err != nil {
			return fmt.Errorf("failed to save location: %w", err)
		}

		// Collect behaviors for each visit
		behaviors, err := s.collectLocationBehaviors(loc)
		if err != nil {
			return fmt.Errorf("failed to collect behaviors for location %d: %w", loc.ID, err)
		}

		// Save behaviors
		for _, behavior := range behaviors {
			if err := s.repo.SaveLocationBehavior(&behavior); err != nil {
				return fmt.Errorf("failed to save behavior: %w", err)
			}
		}

		// Step 3: Calculate efficiency scores
		efficiencyScore, err := s.calculateEfficiencyScore(loc.ID, behaviors)
		if err != nil {
			return fmt.Errorf("failed to calculate efficiency: %w", err)
		}
		if err := s.repo.SaveLocationEfficiencyScore(efficiencyScore); err != nil {
			return fmt.Errorf("failed to save efficiency score: %w", err)
		}

		// Step 4: Detect habits
		habits := s.detectHabits(loc, behaviors)
		for _, habit := range habits {
			if err := s.repo.SaveLocationHabit(&habit); err != nil {
				return fmt.Errorf("failed to save habit: %w", err)
			}
		}

		// Step 5: Auto-label location
		label, confidence := s.autoLabelLocation(loc, behaviors, efficiencyScore)
		loc.Label = label
		loc.LabelConfidence = confidence
		if err := s.repo.SaveLocation(loc); err != nil {
			return fmt.Errorf("failed to update location label: %w", err)
		}
	}

	return nil
}

// clusterLocations clusters stay_segments by GeoHash6
func (s *LocationBehaviorService) clusterLocations() ([]models.Location, error) {
	// Query significant stays (≥2 hours, ≥10 points, confidence ≥0.7)
	query := `
		SELECT id, start_time, end_time, duration_s, center_lat, center_lon,
		       radius_m, geohash6, point_count, confidence
		FROM stay_segments
		WHERE duration_s >= 7200
		  AND point_count >= 10
		  AND confidence >= 0.7
		  AND geohash6 IS NOT NULL
		ORDER BY start_time
	`

	rows, err := s.tracksDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Group stays by GeoHash6
	geohashMap := make(map[string]*models.Location)
	stayMap := make(map[string][]stayInfo)

	for rows.Next() {
		var stay stayInfo
		err := rows.Scan(&stay.ID, &stay.StartTime, &stay.EndTime, &stay.Duration,
			&stay.CenterLat, &stay.CenterLon, &stay.Radius, &stay.Geohash6,
			&stay.PointCount, &stay.Confidence)
		if err != nil {
			continue
		}

		if _, exists := geohashMap[stay.Geohash6]; !exists {
			geohashMap[stay.Geohash6] = &models.Location{
				Geohash:      stay.Geohash6,
				CenterLat:    stay.CenterLat,
				CenterLon:    stay.CenterLon,
				Radius:       stay.Radius,
				VisitCount:   0,
				TotalDuration: 0,
				FirstVisit:   time.Unix(stay.StartTime, 0).Format("2006-01-02 15:04:05"),
				LastVisit:    time.Unix(stay.EndTime, 0).Format("2006-01-02 15:04:05"),
			}
		}

		loc := geohashMap[stay.Geohash6]
		loc.VisitCount++
		loc.TotalDuration += stay.Duration
		loc.LastVisit = time.Unix(stay.EndTime, 0).Format("2006-01-02 15:04:05")

		// Update center (weighted average by duration)
		totalWeight := float64(loc.TotalDuration)
		loc.CenterLat = (loc.CenterLat*float64(loc.TotalDuration-stay.Duration) + stay.CenterLat*float64(stay.Duration)) / totalWeight
		loc.CenterLon = (loc.CenterLon*float64(loc.TotalDuration-stay.Duration) + stay.CenterLon*float64(stay.Duration)) / totalWeight

		stayMap[stay.Geohash6] = append(stayMap[stay.Geohash6], stay)
	}

	// Convert map to slice
	var locations []models.Location
	for _, loc := range geohashMap {
		locations = append(locations, *loc)
	}

	return locations, nil
}

// stayInfo holds stay segment data for clustering
type stayInfo struct {
	ID         int
	StartTime  int64
	EndTime    int64
	Duration   int
	CenterLat  float64
	CenterLon  float64
	Radius     float64
	Geohash6   string
	PointCount int
	Confidence float64
}

// collectLocationBehaviors collects behavior data for all visits to a location
func (s *LocationBehaviorService) collectLocationBehaviors(loc *models.Location) ([]models.LocationBehavior, error) {
	// Get all stays for this location
	query := `
		SELECT id, start_time, end_time, duration_s
		FROM stay_segments
		WHERE geohash6 = ?
		  AND duration_s >= 7200
		  AND point_count >= 10
		  AND confidence >= 0.7
		ORDER BY start_time
	`

	rows, err := s.tracksDB.Query(query, loc.Geohash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var behaviors []models.LocationBehavior
	for rows.Next() {
		var stayID int
		var startTime, endTime int64
		var duration int
		if err := rows.Scan(&stayID, &startTime, &endTime, &duration); err != nil {
			continue
		}

		visitDate := time.Unix(startTime, 0).Format("2006-01-02")
		visitStart := time.Unix(startTime, 0).Format("2006-01-02 15:04:05")
		visitEnd := time.Unix(endTime, 0).Format("2006-01-02 15:04:05")

		behavior := models.LocationBehavior{
			LocationID: loc.ID,
			VisitDate:  visitDate,
			VisitStart: visitStart,
			VisitEnd:   visitEnd,
			Duration:   duration,
		}

		// Collect keyboard data
		typingSpeed := s.getTypingSpeed(visitDate)
		behavior.TypingSpeed = typingSpeed

		// Collect screentime data
		workRatio, entertainmentRatio, focusDuration, switchCount := s.getScreentimeMetrics(startTime, endTime)
		behavior.WorkAppRatio = workRatio
		behavior.EntertainmentRatio = entertainmentRatio
		behavior.FocusDuration = focusDuration
		behavior.AppSwitchCount = switchCount

		// Collect health data
		avgHR, hrv, steps := s.getHealthMetrics(visitDate)
		behavior.AvgHeartRate = avgHR
		behavior.HeartRateVariability = hrv
		behavior.Steps = steps

		behaviors = append(behaviors, behavior)
	}

	return behaviors, nil
}

// getTypingSpeed retrieves typing speed for a date
func (s *LocationBehaviorService) getTypingSpeed(date string) float64 {
	query := `SELECT total_keys FROM daily_stats WHERE date = ?`
	var totalKeys int
	if err := s.keyboardDB.QueryRow(query, date).Scan(&totalKeys); err != nil {
		return 0
	}
	// Assume 8 hours of work per day
	return float64(totalKeys) / 8.0
}

// getScreentimeMetrics retrieves screentime metrics for a time range
func (s *LocationBehaviorService) getScreentimeMetrics(startTime, endTime int64) (workRatio, entertainmentRatio float64, focusDuration, switchCount int) {
	// Query phone sessions within time range
	query := `
		SELECT app_name, start_time, end_time, duration_ms
		FROM phone_sessions
		WHERE start_time >= ? AND end_time <= ?
		ORDER BY start_time
	`

	rows, err := s.screentimeDB.Query(query, startTime, endTime)
	if err != nil {
		return 0, 0, 0, 0
	}
	defer rows.Close()

	var totalDuration, workDuration, entertainmentDuration int64
	sessionCount := 0

	workApps := map[string]bool{
		"com.microsoft.office": true,
		"com.google.docs":      true,
		"com.jetbrains":        true,
		"com.vscode":           true,
	}

	entertainmentApps := map[string]bool{
		"com.tiktok":    true,
		"com.youtube":   true,
		"com.netflix":   true,
		"com.bilibili":  true,
		"com.douyin":    true,
	}

	for rows.Next() {
		var appName string
		var start, end, duration int64
		if err := rows.Scan(&appName, &start, &end, &duration); err != nil {
			continue
		}

		totalDuration += duration
		sessionCount++

		// Check if work app
		isWork := false
		for workApp := range workApps {
			if strings.Contains(appName, workApp) {
				workDuration += duration
				isWork = true
				break
			}
		}

		// Check if entertainment app
		if !isWork {
			for entApp := range entertainmentApps {
				if strings.Contains(appName, entApp) {
					entertainmentDuration += duration
					break
				}
			}
		}

		// Calculate focus duration (sessions > 10 minutes)
		if duration > 600000 { // 10 minutes in ms
			focusDuration += int(duration / 1000)
		}
	}

	if totalDuration > 0 {
		workRatio = float64(workDuration) / float64(totalDuration)
		entertainmentRatio = float64(entertainmentDuration) / float64(totalDuration)
	}

	switchCount = sessionCount

	return workRatio, entertainmentRatio, focusDuration, switchCount
}

// getHealthMetrics retrieves health metrics for a date
func (s *LocationBehaviorService) getHealthMetrics(date string) (avgHR, hrv float64, steps int) {
	// Query heart rate
	hrQuery := `
		SELECT AVG(value) as avg_hr
		FROM health_records
		WHERE type = 'HeartRate'
		  AND DATE(startDate) = ?
	`
	s.healthDB.QueryRow(hrQuery, date).Scan(&avgHR)

	// Query HRV
	hrvQuery := `
		SELECT AVG(value) as avg_hrv
		FROM health_records
		WHERE type = 'HeartRateVariability'
		  AND DATE(startDate) = ?
	`
	s.healthDB.QueryRow(hrvQuery, date).Scan(&hrv)

	// Query steps
	stepsQuery := `
		SELECT SUM(value) as total_steps
		FROM health_records
		WHERE type = 'StepCount'
		  AND DATE(startDate) = ?
	`
	s.healthDB.QueryRow(stepsQuery, date).Scan(&steps)

	return avgHR, hrv, steps
}

// calculateEfficiencyScore calculates efficiency score for a location
func (s *LocationBehaviorService) calculateEfficiencyScore(locationID int, behaviors []models.LocationBehavior) (*models.LocationEfficiencyScore, error) {
	if len(behaviors) == 0 {
		return &models.LocationEfficiencyScore{
			LocationID:        locationID,
			ProductivityScore: 0,
			HealthScore:       0,
			FocusScore:        0,
			EfficiencyScore:   0,
			VisitCount:        0,
			AvgDuration:       0,
		}, nil
	}

	// Calculate average metrics
	var totalTypingSpeed, totalWorkRatio, totalFocusDuration float64
	var totalEntertainmentRatio, totalSwitchFreq float64
	var totalHRV, totalHR, totalSteps float64
	var totalDuration int

	for _, b := range behaviors {
		totalTypingSpeed += b.TypingSpeed
		totalWorkRatio += b.WorkAppRatio
		totalFocusDuration += float64(b.FocusDuration)
		totalEntertainmentRatio += b.EntertainmentRatio
		if b.Duration > 0 {
			totalSwitchFreq += float64(b.AppSwitchCount) / float64(b.Duration/3600) // switches per hour
		}
		totalHRV += b.HeartRateVariability
		totalHR += b.AvgHeartRate
		totalSteps += float64(b.Steps)
		totalDuration += b.Duration
	}

	count := float64(len(behaviors))
	avgTypingSpeed := totalTypingSpeed / count
	avgWorkRatio := totalWorkRatio / count
	avgFocusDuration := totalFocusDuration / count
	avgEntertainmentRatio := totalEntertainmentRatio / count
	avgSwitchFreq := totalSwitchFreq / count
	avgHRV := totalHRV / count
	avgHR := totalHR / count
	avgSteps := totalSteps / count
	avgDuration := totalDuration / len(behaviors)

	// Normalize and calculate scores
	// Productivity Score = 0.4 * typing_speed + 0.3 * work_ratio + 0.3 * focus_duration
	productivityScore := 0.4*normalize(avgTypingSpeed, 0, 500, 100) +
		0.3*normalize(avgWorkRatio, 0, 1, 100) +
		0.3*normalize(avgFocusDuration, 0, 7200, 100)

	// Health Score = 0.5 * HRV + 0.3 * (1/HR) + 0.2 * steps
	healthScore := 0.5*normalize(avgHRV, 0, 100, 100) +
		0.3*normalize(1/(avgHR/100), 0, 2, 100) +
		0.2*normalize(avgSteps, 0, 10000, 100)

	// Focus Score = 0.5 * (1 - entertainment_ratio) + 0.5 * (1 - switch_freq)
	focusScore := 0.5*normalize(1-avgEntertainmentRatio, 0, 1, 100) +
		0.5*normalize(1-avgSwitchFreq/10, 0, 1, 100)

	// Total Efficiency = 0.50 * productivity + 0.25 * health + 0.25 * focus
	efficiencyScore := 0.50*productivityScore + 0.25*healthScore + 0.25*focusScore

	return &models.LocationEfficiencyScore{
		LocationID:        locationID,
		ProductivityScore: productivityScore,
		HealthScore:       healthScore,
		FocusScore:        focusScore,
		EfficiencyScore:   efficiencyScore,
		VisitCount:        len(behaviors),
		AvgDuration:       avgDuration,
	}, nil
}

// normalize normalizes a value to 0-100 range
func normalize(value, min, max, scale float64) float64 {
	if max == min {
		return 0
	}
	normalized := (value - min) / (max - min) * scale
	if normalized < 0 {
		return 0
	}
	if normalized > scale {
		return scale
	}
	return normalized
}

// detectHabits detects location-specific habits
func (s *LocationBehaviorService) detectHabits(loc *models.Location, behaviors []models.LocationBehavior) []models.LocationHabit {
	var habits []models.LocationHabit

	if len(behaviors) < 3 {
		return habits
	}

	// Habit 1: High entertainment usage
	entertainmentCount := 0
	for _, b := range behaviors {
		if b.EntertainmentRatio > 0.5 {
			entertainmentCount++
		}
	}
	if float64(entertainmentCount)/float64(len(behaviors)) > 0.6 {
		habits = append(habits, models.LocationHabit{
			LocationID:       loc.ID,
			HabitType:        "HIGH_ENTERTAINMENT",
			HabitDescription: fmt.Sprintf("在此地点经常使用娱乐应用 (%.1f%%的访问)", float64(entertainmentCount)/float64(len(behaviors))*100),
			Confidence:       float64(entertainmentCount) / float64(len(behaviors)),
			OccurrenceCount:  entertainmentCount,
		})
	}

	// Habit 2: High productivity
	productivityCount := 0
	for _, b := range behaviors {
		if b.WorkAppRatio > 0.6 && b.TypingSpeed > 200 {
			productivityCount++
		}
	}
	if float64(productivityCount)/float64(len(behaviors)) > 0.6 {
		habits = append(habits, models.LocationHabit{
			LocationID:       loc.ID,
			HabitType:        "HIGH_PRODUCTIVITY",
			HabitDescription: fmt.Sprintf("在此地点工作效率高 (%.1f%%的访问)", float64(productivityCount)/float64(len(behaviors))*100),
			Confidence:       float64(productivityCount) / float64(len(behaviors)),
			OccurrenceCount:  productivityCount,
		})
	}

	// Habit 3: Frequent app switching
	switchCount := 0
	for _, b := range behaviors {
		if b.Duration > 0 && float64(b.AppSwitchCount)/float64(b.Duration/3600) > 20 {
			switchCount++
		}
	}
	if float64(switchCount)/float64(len(behaviors)) > 0.5 {
		habits = append(habits, models.LocationHabit{
			LocationID:       loc.ID,
			HabitType:        "FREQUENT_SWITCHING",
			HabitDescription: fmt.Sprintf("在此地点频繁切换应用 (%.1f%%的访问)", float64(switchCount)/float64(len(behaviors))*100),
			Confidence:       float64(switchCount) / float64(len(behaviors)),
			OccurrenceCount:  switchCount,
		})
	}

	// Habit 4: Long focus sessions
	focusCount := 0
	for _, b := range behaviors {
		if b.FocusDuration > 3600 { // > 1 hour
			focusCount++
		}
	}
	if float64(focusCount)/float64(len(behaviors)) > 0.5 {
		habits = append(habits, models.LocationHabit{
			LocationID:       loc.ID,
			HabitType:        "LONG_FOCUS",
			HabitDescription: fmt.Sprintf("在此地点能够长时间专注 (%.1f%%的访问)", float64(focusCount)/float64(len(behaviors))*100),
			Confidence:       float64(focusCount) / float64(len(behaviors)),
			OccurrenceCount:  focusCount,
		})
	}

	return habits
}

// autoLabelLocation automatically labels a location based on behavior patterns
func (s *LocationBehaviorService) autoLabelLocation(loc *models.Location, behaviors []models.LocationBehavior, score *models.LocationEfficiencyScore) (string, float64) {
	if len(behaviors) == 0 {
		return "UNKNOWN", 0.0
	}

	// Calculate time distribution
	nightCount := 0 // 22:00 - 06:00
	workHourCount := 0 // 09:00 - 18:00
	weekdayCount := 0

	for _, b := range behaviors {
		visitTime, _ := time.Parse("2006-01-02 15:04:05", b.VisitStart)
		hour := visitTime.Hour()
		weekday := visitTime.Weekday()

		if hour >= 22 || hour < 6 {
			nightCount++
		}
		if hour >= 9 && hour <= 18 {
			workHourCount++
		}
		if weekday >= 1 && weekday <= 5 {
			weekdayCount++
		}
	}

	nightRatio := float64(nightCount) / float64(len(behaviors))
	workHourRatio := float64(workHourCount) / float64(len(behaviors))
	weekdayRatio := float64(weekdayCount) / float64(len(behaviors))

	// HOME: Most frequent, high night ratio, low productivity
	if loc.VisitCount >= 10 && nightRatio > 0.5 && score.ProductivityScore < 40 {
		return "HOME", math.Min(nightRatio+0.2, 1.0)
	}

	// OFFICE: High productivity, work hours, weekdays
	if score.ProductivityScore > 70 && workHourRatio > 0.7 && weekdayRatio > 0.8 {
		return "OFFICE", math.Min(score.ProductivityScore/100+0.2, 1.0)
	}

	// CAFE: Medium duration (2-4h), high productivity, public place
	avgDuration := float64(score.AvgDuration) / 3600.0
	if avgDuration >= 2 && avgDuration <= 4 && score.ProductivityScore > 60 {
		return "CAFE", 0.7
	}

	// GYM: High steps, short duration (<2h)
	avgSteps := 0.0
	for _, b := range behaviors {
		avgSteps += float64(b.Steps)
	}
	avgSteps /= float64(len(behaviors))
	if avgSteps > 5000 && avgDuration < 2 {
		return "GYM", 0.6
	}

	// TRANSIT: Very short duration, low productivity
	if avgDuration < 0.5 && score.ProductivityScore < 30 {
		return "TRANSIT", 0.5
	}

	// LEISURE: High entertainment, weekends
	if score.FocusScore < 40 && weekdayRatio < 0.3 {
		return "LEISURE", 0.6
	}

	return "OTHER", 0.3
}

// GetAllLocations retrieves all analyzed locations
func (s *LocationBehaviorService) GetAllLocations() ([]models.LocationWithEfficiency, error) {
	locations, err := s.repo.GetAllLocations()
	if err != nil {
		return nil, err
	}

	var result []models.LocationWithEfficiency
	for _, loc := range locations {
		score, err := s.repo.GetLocationEfficiencyScore(loc.ID)
		if err != nil {
			continue
		}
		if score == nil {
			score = &models.LocationEfficiencyScore{}
		}

		habits, err := s.repo.GetLocationHabits(loc.ID)
		if err != nil {
			habits = []models.LocationHabit{}
		}

		result = append(result, models.LocationWithEfficiency{
			Location:        loc,
			EfficiencyScore: *score,
			Habits:          habits,
		})
	}

	return result, nil
}

// GetLocationByID retrieves a single location with full details
func (s *LocationBehaviorService) GetLocationByID(id int) (*models.LocationWithEfficiency, error) {
	loc, err := s.repo.GetLocationByID(id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, nil
	}

	score, err := s.repo.GetLocationEfficiencyScore(id)
	if err != nil {
		return nil, err
	}
	if score == nil {
		score = &models.LocationEfficiencyScore{}
	}

	habits, err := s.repo.GetLocationHabits(id)
	if err != nil {
		habits = []models.LocationHabit{}
	}

	_, err = s.repo.GetLocationBehaviors(id)
	if err != nil {
		// behaviors not used in response, just checking for errors
	}

	return &models.LocationWithEfficiency{
		Location:        *loc,
		EfficiencyScore: *score,
		Habits:          habits,
	}, nil
}

// GetTopEfficientLocations retrieves top N most efficient locations
func (s *LocationBehaviorService) GetTopEfficientLocations(limit int) ([]models.LocationWithEfficiency, error) {
	locationIDs, err := s.repo.GetTopEfficientLocations(limit)
	if err != nil {
		return nil, err
	}

	var result []models.LocationWithEfficiency
	for _, id := range locationIDs {
		loc, err := s.GetLocationByID(id)
		if err != nil || loc == nil {
			continue
		}
		result = append(result, *loc)
	}

	return result, nil
}

// GetLeastEfficientLocations retrieves bottom N least efficient locations
func (s *LocationBehaviorService) GetLeastEfficientLocations(limit int) ([]models.LocationWithEfficiency, error) {
	locationIDs, err := s.repo.GetLeastEfficientLocations(limit)
	if err != nil {
		return nil, err
	}

	var result []models.LocationWithEfficiency
	for _, id := range locationIDs {
		loc, err := s.GetLocationByID(id)
		if err != nil || loc == nil {
			continue
		}
		result = append(result, *loc)
	}

	return result, nil
}
