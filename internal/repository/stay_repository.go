package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// StayRepository handles database operations for stay segments
type StayRepository struct {
	db *sql.DB
}

// NewStayRepository creates a new stay repository
func NewStayRepository(db *sql.DB) *StayRepository {
	return &StayRepository{db: db}
}

// GetStays retrieves stay segments with filtering and pagination
func (r *StayRepository) GetStays(filter models.StayFilter) ([]models.StaySegment, int64, error) {
	// Build query
	query := `SELECT id, stay_type, stay_category, start_time, end_time, duration_seconds,
		center_lat, center_lon, radius_meters, point_count,
		province, city, county, town, village,
		confidence, metadata, algo_version, created_at, updated_at
		FROM stay_segments`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.StayType != "" {
		conditions = append(conditions, "stay_type = ?")
		args = append(args, filter.StayType)
	}
	if filter.StayCategory != "" {
		conditions = append(conditions, "stay_category = ?")
		args = append(args, filter.StayCategory)
	}
	if filter.MinDuration > 0 {
		conditions = append(conditions, "duration_seconds >= ?")
		args = append(args, filter.MinDuration)
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
	if filter.StartTime > 0 {
		conditions = append(conditions, "start_time >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime > 0 {
		conditions = append(conditions, "end_time <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.MinConfidence > 0 {
		conditions = append(conditions, "confidence >= ?")
		args = append(args, filter.MinConfidence)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM stay_segments"
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count stay segments: %w", err)
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
		return nil, 0, fmt.Errorf("failed to query stay segments: %w", err)
	}
	defer rows.Close()

	var stays []models.StaySegment
	for rows.Next() {
		var s models.StaySegment
		err := rows.Scan(
			&s.ID, &s.StayType, &s.StayCategory, &s.StartTime, &s.EndTime, &s.DurationSeconds,
			&s.CenterLat, &s.CenterLon, &s.RadiusMeters, &s.PointCount,
			&s.Province, &s.City, &s.County, &s.Town, &s.Village,
			&s.Confidence, &s.Metadata, &s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stay segment: %w", err)
		}
		stays = append(stays, s)
	}

	return stays, total, nil
}

// GetStayByID retrieves a single stay segment by ID
func (r *StayRepository) GetStayByID(id int64) (*models.StaySegment, error) {
	query := `SELECT id, stay_type, stay_category, start_time, end_time, duration_seconds,
		center_lat, center_lon, radius_meters, point_count,
		province, city, county, town, village,
		confidence, metadata, algo_version, created_at, updated_at
		FROM stay_segments WHERE id = ?`

	var s models.StaySegment
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.StayType, &s.StayCategory, &s.StartTime, &s.EndTime, &s.DurationSeconds,
		&s.CenterLat, &s.CenterLon, &s.RadiusMeters, &s.PointCount,
		&s.Province, &s.City, &s.County, &s.Town, &s.Village,
		&s.Confidence, &s.Metadata, &s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stay segment: %w", err)
	}

	return &s, nil
}
