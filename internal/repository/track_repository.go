package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// TrackRepository handles database operations for track points
type TrackRepository struct {
	db *sql.DB
}

// NewTrackRepository creates a new track repository
func NewTrackRepository(db *sql.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

// GetTrackPoints retrieves track points with filtering and pagination
func (r *TrackRepository) GetTrackPoints(filter models.TrackPointFilter) ([]models.TrackPoint, int64, error) {
	// Build query
	query := `SELECT id, dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude,
		time_visually, time, province, city, county, town, village, created_at, updated_at, algo_version
		FROM "一生足迹"`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.StartTime > 0 {
		conditions = append(conditions, "dataTime >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime > 0 {
		conditions = append(conditions, "dataTime <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.Province != "" {
		conditions = append(conditions, "province = ?")
		args = append(args, filter.Province)
	}
	if filter.City != "" {
		conditions = append(conditions, "city = ?")
		args = append(args, filter.City)
	}
	if filter.County != "" {
		conditions = append(conditions, "county = ?")
		args = append(args, filter.County)
	}
	if filter.MinSpeed > 0 {
		conditions = append(conditions, "speed >= ?")
		args = append(args, filter.MinSpeed)
	}
	if filter.MaxSpeed > 0 {
		conditions = append(conditions, "speed <= ?")
		args = append(args, filter.MaxSpeed)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM \"一生足迹\""
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count track points: %w", err)
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
	query += " ORDER BY dataTime DESC LIMIT ? OFFSET ?"
	args = append(args, filter.PageSize, offset)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query track points: %w", err)
	}
	defer rows.Close()

	var points []models.TrackPoint
	for rows.Next() {
		var p models.TrackPoint
		err := rows.Scan(
			&p.ID, &p.DataTime, &p.Longitude, &p.Latitude, &p.Heading, &p.Accuracy,
			&p.Speed, &p.Distance, &p.Altitude, &p.TimeVisually, &p.Time,
			&p.Province, &p.City, &p.County, &p.Town, &p.Village,
			&p.CreatedAt, &p.UpdatedAt, &p.AlgoVersion,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan track point: %w", err)
		}
		points = append(points, p)
	}

	return points, total, nil
}

// GetTrackPointByID retrieves a single track point by ID
func (r *TrackRepository) GetTrackPointByID(id int64) (*models.TrackPoint, error) {
	query := `SELECT id, dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude,
		time_visually, time, province, city, county, town, village, created_at, updated_at, algo_version
		FROM "一生足迹" WHERE id = ?`

	var p models.TrackPoint
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.DataTime, &p.Longitude, &p.Latitude, &p.Heading, &p.Accuracy,
		&p.Speed, &p.Distance, &p.Altitude, &p.TimeVisually, &p.Time,
		&p.Province, &p.City, &p.County, &p.Town, &p.Village,
		&p.CreatedAt, &p.UpdatedAt, &p.AlgoVersion,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get track point: %w", err)
	}

	return &p, nil
}

// UpdateAdminDivisions updates administrative divisions for a track point
func (r *TrackRepository) UpdateAdminDivisions(id int64, province, city, county, town, village string) error {
	query := `UPDATE "一生足迹"
		SET province = ?, city = ?, county = ?, town = ?, village = ?, updated_at = datetime('now')
		WHERE id = ?`

	_, err := r.db.Exec(query, province, city, county, town, village, id)
	if err != nil {
		return fmt.Errorf("failed to update admin divisions: %w", err)
	}

	return nil
}

// BatchUpdateAdminDivisions updates administrative divisions for multiple track points
func (r *TrackRepository) BatchUpdateAdminDivisions(updates []struct {
	ID       int64
	Province string
	City     string
	County   string
	Town     string
	Village  string
}) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	stmt, err := tx.Prepare(`UPDATE "一生足迹"
		SET province = ?, city = ?, county = ?, town = ?, village = ?, updated_at = datetime('now')
		WHERE id = ?`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, update := range updates {
		_, err := stmt.Exec(update.Province, update.City, update.County, update.Town, update.Village, update.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update track point %d: %w", update.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUngeocodedPoints retrieves track points without administrative divisions
func (r *TrackRepository) GetUngeocodedPoints(limit int) ([]models.TrackPoint, error) {
	query := `SELECT id, dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude,
		time_visually, time, province, city, county, town, village, created_at, updated_at, algo_version
		FROM "一生足迹"
		WHERE province IS NULL OR province = ''
		ORDER BY dataTime ASC
		LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query ungeocoded points: %w", err)
	}
	defer rows.Close()

	var points []models.TrackPoint
	for rows.Next() {
		var p models.TrackPoint
		err := rows.Scan(
			&p.ID, &p.DataTime, &p.Longitude, &p.Latitude, &p.Heading, &p.Accuracy,
			&p.Speed, &p.Distance, &p.Altitude, &p.TimeVisually, &p.Time,
			&p.Province, &p.City, &p.County, &p.Town, &p.Village,
			&p.CreatedAt, &p.UpdatedAt, &p.AlgoVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track point: %w", err)
		}
		points = append(points, p)
	}

	return points, nil
}
