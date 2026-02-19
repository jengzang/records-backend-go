package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// SegmentRepository handles database operations for segments
type SegmentRepository struct {
	db *sql.DB
}

// NewSegmentRepository creates a new segment repository
func NewSegmentRepository(db *sql.DB) *SegmentRepository {
	return &SegmentRepository{db: db}
}

// GetSegments retrieves segments with filtering and pagination
func (r *SegmentRepository) GetSegments(filter models.SegmentFilter) ([]models.Segment, int64, error) {
	// Build query
	query := `SELECT id, mode, start_point_id, end_point_id, start_time, end_time, duration_seconds,
		distance_meters, start_lat, start_lon, end_lat, end_lon,
		avg_speed_kmh, max_speed_kmh, avg_heading, heading_variance,
		confidence, reason_codes, province, city, county,
		algo_version, created_at, updated_at
		FROM segments`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.Mode != "" {
		conditions = append(conditions, "mode = ?")
		args = append(args, filter.Mode)
	}
	if filter.StartTime > 0 {
		conditions = append(conditions, "start_time >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime > 0 {
		conditions = append(conditions, "end_time <= ?")
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
	if filter.MinDistance > 0 {
		conditions = append(conditions, "distance_meters >= ?")
		args = append(args, filter.MinDistance)
	}
	if filter.MinDuration > 0 {
		conditions = append(conditions, "duration_seconds >= ?")
		args = append(args, filter.MinDuration)
	}
	if filter.MinConfidence > 0 {
		conditions = append(conditions, "confidence >= ?")
		args = append(args, filter.MinConfidence)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM segments"
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count segments: %w", err)
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
		return nil, 0, fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	var segments []models.Segment
	for rows.Next() {
		var s models.Segment
		err := rows.Scan(
			&s.ID, &s.Mode, &s.StartPointID, &s.EndPointID, &s.StartTime, &s.EndTime, &s.DurationSeconds,
			&s.DistanceMeters, &s.StartLat, &s.StartLon, &s.EndLat, &s.EndLon,
			&s.AvgSpeedKmh, &s.MaxSpeedKmh, &s.AvgHeading, &s.HeadingVariance,
			&s.Confidence, &s.ReasonCodes, &s.Province, &s.City, &s.County,
			&s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, s)
	}

	return segments, total, nil
}

// GetSegmentByID retrieves a single segment by ID
func (r *SegmentRepository) GetSegmentByID(id int64) (*models.Segment, error) {
	query := `SELECT id, mode, start_point_id, end_point_id, start_time, end_time, duration_seconds,
		distance_meters, start_lat, start_lon, end_lat, end_lon,
		avg_speed_kmh, max_speed_kmh, avg_heading, heading_variance,
		confidence, reason_codes, province, city, county,
		algo_version, created_at, updated_at
		FROM segments WHERE id = ?`

	var s models.Segment
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.Mode, &s.StartPointID, &s.EndPointID, &s.StartTime, &s.EndTime, &s.DurationSeconds,
		&s.DistanceMeters, &s.StartLat, &s.StartLon, &s.EndLat, &s.EndLon,
		&s.AvgSpeedKmh, &s.MaxSpeedKmh, &s.AvgHeading, &s.HeadingVariance,
		&s.Confidence, &s.ReasonCodes, &s.Province, &s.City, &s.County,
		&s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get segment: %w", err)
	}

	return &s, nil
}
