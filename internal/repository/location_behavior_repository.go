package repository

import (
	"database/sql"

	"github.com/jengzang/records-backend-go/internal/models"
)

// LocationBehaviorRepository handles database operations for location behavior analysis
type LocationBehaviorRepository struct {
	db *sql.DB
}

// NewLocationBehaviorRepository creates a new repository
func NewLocationBehaviorRepository(db *sql.DB) *LocationBehaviorRepository {
	return &LocationBehaviorRepository{db: db}
}

// GetAllLocations retrieves all locations
func (r *LocationBehaviorRepository) GetAllLocations() ([]models.Location, error) {
	query := `
		SELECT id, geohash, center_lat, center_lon, radius, visit_count,
		       total_duration, first_visit, last_visit, label, label_confidence,
		       created_at, updated_at
		FROM locations
		ORDER BY visit_count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var loc models.Location
		err := rows.Scan(
			&loc.ID, &loc.Geohash, &loc.CenterLat, &loc.CenterLon, &loc.Radius,
			&loc.VisitCount, &loc.TotalDuration, &loc.FirstVisit, &loc.LastVisit,
			&loc.Label, &loc.LabelConfidence, &loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			continue
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// GetLocationByID retrieves a location by ID
func (r *LocationBehaviorRepository) GetLocationByID(id int) (*models.Location, error) {
	query := `
		SELECT id, geohash, center_lat, center_lon, radius, visit_count,
		       total_duration, first_visit, last_visit, label, label_confidence,
		       created_at, updated_at
		FROM locations
		WHERE id = ?
	`

	var loc models.Location
	err := r.db.QueryRow(query, id).Scan(
		&loc.ID, &loc.Geohash, &loc.CenterLat, &loc.CenterLon, &loc.Radius,
		&loc.VisitCount, &loc.TotalDuration, &loc.FirstVisit, &loc.LastVisit,
		&loc.Label, &loc.LabelConfidence, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &loc, nil
}

// GetLocationEfficiencyScore retrieves efficiency score for a location
func (r *LocationBehaviorRepository) GetLocationEfficiencyScore(locationID int) (*models.LocationEfficiencyScore, error) {
	query := `
		SELECT id, location_id, productivity_score, health_score, focus_score,
		       efficiency_score, visit_count, avg_duration, calculated_at
		FROM location_efficiency_scores
		WHERE location_id = ?
	`

	var score models.LocationEfficiencyScore
	err := r.db.QueryRow(query, locationID).Scan(
		&score.ID, &score.LocationID, &score.ProductivityScore, &score.HealthScore,
		&score.FocusScore, &score.EfficiencyScore, &score.VisitCount,
		&score.AvgDuration, &score.CalculatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &score, nil
}

// GetLocationHabits retrieves habits for a location
func (r *LocationBehaviorRepository) GetLocationHabits(locationID int) ([]models.LocationHabit, error) {
	query := `
		SELECT id, location_id, habit_type, habit_description, confidence,
		       occurrence_count, detected_at
		FROM location_habits
		WHERE location_id = ?
		ORDER BY confidence DESC
	`

	rows, err := r.db.Query(query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []models.LocationHabit
	for rows.Next() {
		var habit models.LocationHabit
		err := rows.Scan(
			&habit.ID, &habit.LocationID, &habit.HabitType, &habit.HabitDescription,
			&habit.Confidence, &habit.OccurrenceCount, &habit.DetectedAt,
		)
		if err != nil {
			continue
		}
		habits = append(habits, habit)
	}

	return habits, nil
}

// SaveLocation saves or updates a location
func (r *LocationBehaviorRepository) SaveLocation(loc *models.Location) error {
	query := `
		INSERT INTO locations (geohash, center_lat, center_lon, radius, visit_count,
		                       total_duration, first_visit, last_visit, label, label_confidence)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			visit_count = excluded.visit_count,
			total_duration = excluded.total_duration,
			last_visit = excluded.last_visit,
			label = excluded.label,
			label_confidence = excluded.label_confidence,
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := r.db.Exec(query,
		loc.Geohash, loc.CenterLat, loc.CenterLon, loc.Radius, loc.VisitCount,
		loc.TotalDuration, loc.FirstVisit, loc.LastVisit, loc.Label, loc.LabelConfidence,
	)
	if err != nil {
		return err
	}

	if loc.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		loc.ID = int(id)
	}

	return nil
}

// SaveLocationBehavior saves behavior data for a location visit
func (r *LocationBehaviorRepository) SaveLocationBehavior(behavior *models.LocationBehavior) error {
	query := `
		INSERT INTO location_behaviors (
			location_id, visit_date, visit_start, visit_end, duration,
			typing_speed, work_app_ratio, entertainment_ratio, focus_duration,
			app_switch_count, avg_heart_rate, heart_rate_variability, steps
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		behavior.LocationID, behavior.VisitDate, behavior.VisitStart, behavior.VisitEnd,
		behavior.Duration, behavior.TypingSpeed, behavior.WorkAppRatio,
		behavior.EntertainmentRatio, behavior.FocusDuration, behavior.AppSwitchCount,
		behavior.AvgHeartRate, behavior.HeartRateVariability, behavior.Steps,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	behavior.ID = int(id)

	return nil
}

// SaveLocationEfficiencyScore saves efficiency score for a location
func (r *LocationBehaviorRepository) SaveLocationEfficiencyScore(score *models.LocationEfficiencyScore) error {
	query := `
		INSERT OR REPLACE INTO location_efficiency_scores (
			location_id, productivity_score, health_score, focus_score,
			efficiency_score, visit_count, avg_duration
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		score.LocationID, score.ProductivityScore, score.HealthScore,
		score.FocusScore, score.EfficiencyScore, score.VisitCount, score.AvgDuration,
	)
	if err != nil {
		return err
	}

	if score.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		score.ID = int(id)
	}

	return nil
}

// SaveLocationHabit saves a detected habit
func (r *LocationBehaviorRepository) SaveLocationHabit(habit *models.LocationHabit) error {
	query := `
		INSERT INTO location_habits (
			location_id, habit_type, habit_description, confidence, occurrence_count
		) VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		habit.LocationID, habit.HabitType, habit.HabitDescription,
		habit.Confidence, habit.OccurrenceCount,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	habit.ID = int(id)

	return nil
}

// GetTopEfficientLocations retrieves top N most efficient locations
func (r *LocationBehaviorRepository) GetTopEfficientLocations(limit int) ([]int, error) {
	query := `
		SELECT location_id
		FROM location_efficiency_scores
		ORDER BY efficiency_score DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locationIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		locationIDs = append(locationIDs, id)
	}

	return locationIDs, nil
}

// GetLeastEfficientLocations retrieves bottom N least efficient locations
func (r *LocationBehaviorRepository) GetLeastEfficientLocations(limit int) ([]int, error) {
	query := `
		SELECT location_id
		FROM location_efficiency_scores
		ORDER BY efficiency_score ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locationIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		locationIDs = append(locationIDs, id)
	}

	return locationIDs, nil
}

// GetLocationBehaviors retrieves all behaviors for a location
func (r *LocationBehaviorRepository) GetLocationBehaviors(locationID int) ([]models.LocationBehavior, error) {
	query := `
		SELECT id, location_id, visit_date, visit_start, visit_end, duration,
		       typing_speed, work_app_ratio, entertainment_ratio, focus_duration,
		       app_switch_count, avg_heart_rate, heart_rate_variability, steps, created_at
		FROM location_behaviors
		WHERE location_id = ?
		ORDER BY visit_date DESC
	`

	rows, err := r.db.Query(query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var behaviors []models.LocationBehavior
	for rows.Next() {
		var b models.LocationBehavior
		err := rows.Scan(
			&b.ID, &b.LocationID, &b.VisitDate, &b.VisitStart, &b.VisitEnd, &b.Duration,
			&b.TypingSpeed, &b.WorkAppRatio, &b.EntertainmentRatio, &b.FocusDuration,
			&b.AppSwitchCount, &b.AvgHeartRate, &b.HeartRateVariability, &b.Steps, &b.CreatedAt,
		)
		if err != nil {
			continue
		}
		behaviors = append(behaviors, b)
	}

	return behaviors, nil
}
