package health

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository handles database operations for health data
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new health repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetSummary retrieves overall health data summary
func (r *Repository) GetSummary() (*HealthSummary, error) {
	summary := &HealthSummary{}

	// Get total records
	err := r.db.QueryRow("SELECT COUNT(*) FROM health_records").Scan(&summary.TotalRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to count health records: %w", err)
	}

	// Get total workouts
	err = r.db.QueryRow("SELECT COUNT(*) FROM workouts").Scan(&summary.TotalWorkouts)
	if err != nil {
		return nil, fmt.Errorf("failed to count workouts: %w", err)
	}

	// Get total sleep records
	err = r.db.QueryRow("SELECT COUNT(*) FROM sleep_analysis").Scan(&summary.TotalSleepRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to count sleep records: %w", err)
	}

	// Get date range
	err = r.db.QueryRow(`
		SELECT MIN(DATE(start_date)), MAX(DATE(start_date))
		FROM health_records
	`).Scan(&summary.DateRangeStart, &summary.DateRangeEnd)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get date range: %w", err)
	}

	// Get last import date
	err = r.db.QueryRow(`
		SELECT import_date FROM import_metadata
		WHERE status = 'success'
		ORDER BY import_date DESC LIMIT 1
	`).Scan(&summary.LastImportDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get last import date: %w", err)
	}

	// Get available metrics
	rows, err := r.db.Query(`
		SELECT DISTINCT type FROM health_records
		ORDER BY type
		LIMIT 50
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get available metrics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var metricType string
		if err := rows.Scan(&metricType); err != nil {
			continue
		}
		summary.AvailableMetrics = append(summary.AvailableMetrics, metricType)
	}

	return summary, nil
}

// GetRecords retrieves health records with filters
func (r *Repository) GetRecords(filter RecordFilter) ([]HealthRecord, error) {
	query := `
		SELECT id, type, value, unit, start_date, end_date,
		       source_name, source_version, device, creation_date,
		       metadata, created_at
		FROM health_records
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.Type != "" {
		query += " AND type = ?"
		args = append(args, filter.Type)
	}

	if !filter.StartDate.IsZero() {
		query += " AND start_date >= ?"
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query += " AND end_date <= ?"
		args = append(args, filter.EndDate)
	}

	query += " ORDER BY start_date DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	} else {
		query += " LIMIT 1000" // Default limit
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query health records: %w", err)
	}
	defer rows.Close()

	var records []HealthRecord
	for rows.Next() {
		var r HealthRecord
		var sourceVersion, device, metadata sql.NullString
		var creationDate sql.NullTime

		err := rows.Scan(
			&r.ID, &r.Type, &r.Value, &r.Unit, &r.StartDate, &r.EndDate,
			&r.SourceName, &sourceVersion, &device, &creationDate,
			&metadata, &r.CreatedAt,
		)
		if err != nil {
			continue
		}

		if sourceVersion.Valid {
			r.SourceVersion = sourceVersion.String
		}
		if device.Valid {
			r.Device = device.String
		}
		if creationDate.Valid {
			r.CreationDate = creationDate.Time
		}
		if metadata.Valid {
			r.Metadata = metadata.String
		}

		records = append(records, r)
	}

	return records, nil
}

// GetWorkouts retrieves workouts with filters
func (r *Repository) GetWorkouts(filter WorkoutFilter) ([]Workout, error) {
	query := `
		SELECT id, workout_type, duration_seconds, duration_unit,
		       distance_meters, distance_unit, calories, calories_unit,
		       start_date, end_date, source_name, source_version,
		       device, creation_date, route_file, metadata, created_at
		FROM workouts
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.WorkoutType != "" {
		query += " AND workout_type = ?"
		args = append(args, filter.WorkoutType)
	}

	if !filter.StartDate.IsZero() {
		query += " AND start_date >= ?"
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query += " AND end_date <= ?"
		args = append(args, filter.EndDate)
	}

	query += " ORDER BY start_date DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	} else {
		query += " LIMIT 100" // Default limit
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query workouts: %w", err)
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		var durationUnit, distanceUnit, caloriesUnit sql.NullString
		var sourceVersion, device, routeFile, metadata sql.NullString
		var creationDate sql.NullTime
		var durationSeconds, distanceMeters, calories sql.NullFloat64

		err := rows.Scan(
			&w.ID, &w.WorkoutType, &durationSeconds, &durationUnit,
			&distanceMeters, &distanceUnit, &calories, &caloriesUnit,
			&w.StartDate, &w.EndDate, &w.SourceName, &sourceVersion,
			&device, &creationDate, &routeFile, &metadata, &w.CreatedAt,
		)
		if err != nil {
			continue
		}

		if durationSeconds.Valid {
			w.DurationSeconds = durationSeconds.Float64
		}
		if durationUnit.Valid {
			w.DurationUnit = durationUnit.String
		}
		if distanceMeters.Valid {
			w.DistanceMeters = distanceMeters.Float64
		}
		if distanceUnit.Valid {
			w.DistanceUnit = distanceUnit.String
		}
		if calories.Valid {
			w.Calories = calories.Float64
		}
		if caloriesUnit.Valid {
			w.CaloriesUnit = caloriesUnit.String
		}
		if sourceVersion.Valid {
			w.SourceVersion = sourceVersion.String
		}
		if device.Valid {
			w.Device = device.String
		}
		if creationDate.Valid {
			w.CreationDate = creationDate.Time
		}
		if routeFile.Valid {
			w.RouteFile = routeFile.String
		}
		if metadata.Valid {
			w.Metadata = metadata.String
		}

		workouts = append(workouts, w)
	}

	return workouts, nil
}

// GetWorkoutByID retrieves a single workout by ID
func (r *Repository) GetWorkoutByID(id int) (*Workout, error) {
	query := `
		SELECT id, workout_type, duration_seconds, duration_unit,
		       distance_meters, distance_unit, calories, calories_unit,
		       start_date, end_date, source_name, source_version,
		       device, creation_date, route_file, metadata, created_at
		FROM workouts
		WHERE id = ?
	`

	var w Workout
	var durationUnit, distanceUnit, caloriesUnit sql.NullString
	var sourceVersion, device, routeFile, metadata sql.NullString
	var creationDate sql.NullTime
	var durationSeconds, distanceMeters, calories sql.NullFloat64

	err := r.db.QueryRow(query, id).Scan(
		&w.ID, &w.WorkoutType, &durationSeconds, &durationUnit,
		&distanceMeters, &distanceUnit, &calories, &caloriesUnit,
		&w.StartDate, &w.EndDate, &w.SourceName, &sourceVersion,
		&device, &creationDate, &routeFile, &metadata, &w.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workout not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	if durationSeconds.Valid {
		w.DurationSeconds = durationSeconds.Float64
	}
	if durationUnit.Valid {
		w.DurationUnit = durationUnit.String
	}
	if distanceMeters.Valid {
		w.DistanceMeters = distanceMeters.Float64
	}
	if distanceUnit.Valid {
		w.DistanceUnit = distanceUnit.String
	}
	if calories.Valid {
		w.Calories = calories.Float64
	}
	if caloriesUnit.Valid {
		w.CaloriesUnit = caloriesUnit.String
	}
	if sourceVersion.Valid {
		w.SourceVersion = sourceVersion.String
	}
	if device.Valid {
		w.Device = device.String
	}
	if creationDate.Valid {
		w.CreationDate = creationDate.Time
	}
	if routeFile.Valid {
		w.RouteFile = routeFile.String
	}
	if metadata.Valid {
		w.Metadata = metadata.String
	}

	return &w, nil
}

// GetWorkoutRoute retrieves GPS points for a workout
func (r *Repository) GetWorkoutRoute(workoutID int) ([]WorkoutRoute, error) {
	query := `
		SELECT id, workout_id, timestamp, latitude, longitude,
		       altitude, speed, horizontal_accuracy, vertical_accuracy,
		       course, created_at
		FROM workout_routes
		WHERE workout_id = ?
		ORDER BY timestamp
	`

	rows, err := r.db.Query(query, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout route: %w", err)
	}
	defer rows.Close()

	var routes []WorkoutRoute
	for rows.Next() {
		var wr WorkoutRoute
		var altitude, speed, hAccuracy, vAccuracy, course sql.NullFloat64

		err := rows.Scan(
			&wr.ID, &wr.WorkoutID, &wr.Timestamp, &wr.Latitude, &wr.Longitude,
			&altitude, &speed, &hAccuracy, &vAccuracy, &course, &wr.CreatedAt,
		)
		if err != nil {
			continue
		}

		if altitude.Valid {
			wr.Altitude = altitude.Float64
		}
		if speed.Valid {
			wr.Speed = speed.Float64
		}
		if hAccuracy.Valid {
			wr.HorizontalAccuracy = hAccuracy.Float64
		}
		if vAccuracy.Valid {
			wr.VerticalAccuracy = vAccuracy.Float64
		}
		if course.Valid {
			wr.Course = course.Float64
		}

		routes = append(routes, wr)
	}

	return routes, nil
}

// GetDailyStatistics retrieves daily statistics for a metric
func (r *Repository) GetDailyStatistics(metricType string, startDate, endDate time.Time) ([]HealthStatistics, error) {
	query := `
		SELECT id, stat_type, stat_date, metric_type,
		       total_value, avg_value, min_value, max_value,
		       count, data, created_at, updated_at
		FROM health_statistics
		WHERE stat_type = 'daily'
		  AND metric_type = ?
		  AND stat_date >= ?
		  AND stat_date <= ?
		ORDER BY stat_date
	`

	rows, err := r.db.Query(query, metricType, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query statistics: %w", err)
	}
	defer rows.Close()

	var stats []HealthStatistics
	for rows.Next() {
		var s HealthStatistics
		var data sql.NullString

		err := rows.Scan(
			&s.ID, &s.StatType, &s.StatDate, &s.MetricType,
			&s.TotalValue, &s.AvgValue, &s.MinValue, &s.MaxValue,
			&s.Count, &data, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if data.Valid {
			s.Data = data.String
		}

		stats = append(stats, s)
	}

	return stats, nil
}
