package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// TripRepository handles database operations for trips
type TripRepository struct {
	db *sql.DB
}

// NewTripRepository creates a new trip repository
func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{db: db}
}

// GetTrips retrieves trips with filtering and pagination
func (r *TripRepository) GetTrips(filter models.TripFilter) ([]models.Trip, int64, error) {
	// Build query
	query := `SELECT id, date, start_time, end_time, duration_seconds,
		origin_stay_id, dest_stay_id, distance_meters,
		primary_mode, modes_json, segment_ids_json, trip_type,
		origin_province, origin_city, dest_province, dest_city,
		algo_version, created_at, updated_at
		FROM trips`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.StartTime > 0 {
		conditions = append(conditions, "start_time >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime > 0 {
		conditions = append(conditions, "end_time <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.OriginCity != "" {
		conditions = append(conditions, "origin_city = ?")
		args = append(args, filter.OriginCity)
	}
	if filter.DestCity != "" {
		conditions = append(conditions, "dest_city = ?")
		args = append(args, filter.DestCity)
	}
	if filter.MinDistance > 0 {
		conditions = append(conditions, "distance_meters >= ?")
		args = append(args, filter.MinDistance)
	}
	if filter.PrimaryMode != "" {
		conditions = append(conditions, "primary_mode = ?")
		args = append(args, filter.PrimaryMode)
	}
	if filter.TripType != "" {
		conditions = append(conditions, "trip_type = ?")
		args = append(args, filter.TripType)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM trips"
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count trips: %w", err)
	}

	// Add pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 100
	}
	if filter.PageSize > 1000 {
		filter.PageSize = 1000
	}

	offset := (filter.Page - 1) * filter.PageSize
	query += " ORDER BY start_time DESC LIMIT ? OFFSET ?"
	args = append(args, filter.PageSize, offset)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query trips: %w", err)
	}
	defer rows.Close()

	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		err := rows.Scan(
			&t.ID, &t.Date, &t.StartTime, &t.EndTime, &t.DurationSeconds,
			&t.OriginStayID, &t.DestStayID, &t.DistanceMeters,
			&t.PrimaryMode, &t.ModesJSON, &t.SegmentIDsJSON, &t.TripType,
			&t.OriginProvince, &t.OriginCity, &t.DestProvince, &t.DestCity,
			&t.AlgoVersion, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan trip: %w", err)
		}
		trips = append(trips, t)
	}

	return trips, total, nil
}

// GetTripByID retrieves a single trip by ID
func (r *TripRepository) GetTripByID(id int64) (*models.Trip, error) {
	query := `SELECT id, date, start_time, end_time, duration_seconds,
		origin_stay_id, dest_stay_id, distance_meters,
		primary_mode, modes_json, segment_ids_json, trip_type,
		origin_province, origin_city, dest_province, dest_city,
		algo_version, created_at, updated_at
		FROM trips WHERE id = ?`

	var t models.Trip
	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.Date, &t.StartTime, &t.EndTime, &t.DurationSeconds,
		&t.OriginStayID, &t.DestStayID, &t.DistanceMeters,
		&t.PrimaryMode, &t.ModesJSON, &t.SegmentIDsJSON, &t.TripType,
		&t.OriginProvince, &t.OriginCity, &t.DestProvince, &t.DestCity,
		&t.AlgoVersion, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	return &t, nil
}
