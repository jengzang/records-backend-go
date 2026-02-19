package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// VisualizationRepository handles database operations for visualization data
type VisualizationRepository struct {
	db *sql.DB
}

// NewVisualizationRepository creates a new visualization repository
func NewVisualizationRepository(db *sql.DB) *VisualizationRepository {
	return &VisualizationRepository{db: db}
}

// GetRenderingMetadata retrieves track points with rendering properties for map display
func (r *VisualizationRepository) GetRenderingMetadata(filter models.RenderFilter) ([]models.TrackPoint, error) {
	// Build query - select only fields needed for rendering
	query := `SELECT id, dataTime, longitude, latitude, heading, speed, altitude,
		mode, render_color, render_width, render_opacity, lod_level
		FROM "一生足迹"`

	var conditions []string
	var args []interface{}

	// Add bounding box filter
	if filter.MinLat != 0 || filter.MaxLat != 0 || filter.MinLon != 0 || filter.MaxLon != 0 {
		if filter.MinLat != 0 {
			conditions = append(conditions, "latitude >= ?")
			args = append(args, filter.MinLat)
		}
		if filter.MaxLat != 0 {
			conditions = append(conditions, "latitude <= ?")
			args = append(args, filter.MaxLat)
		}
		if filter.MinLon != 0 {
			conditions = append(conditions, "longitude >= ?")
			args = append(args, filter.MinLon)
		}
		if filter.MaxLon != 0 {
			conditions = append(conditions, "longitude <= ?")
			args = append(args, filter.MaxLon)
		}
	}

	// Add time range filter
	if filter.StartTime > 0 {
		conditions = append(conditions, "dataTime >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime > 0 {
		conditions = append(conditions, "dataTime <= ?")
		args = append(args, filter.EndTime)
	}

	// Add mode filter
	if filter.Mode != "" {
		conditions = append(conditions, "mode = ?")
		args = append(args, filter.Mode)
	}

	// Add LOD level filter
	if filter.LODLevel > 0 {
		conditions = append(conditions, "lod_level <= ?")
		args = append(args, filter.LODLevel)
	}

	// Exclude outliers
	conditions = append(conditions, "(outlier_flag IS NULL OR outlier_flag = 0)")

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by time
	query += " ORDER BY dataTime ASC"

	// Limit results
	limit := 10000
	if filter.Limit > 0 && filter.Limit <= 50000 {
		limit = filter.Limit
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rendering metadata: %w", err)
	}
	defer rows.Close()

	var points []models.TrackPoint
	for rows.Next() {
		var p models.TrackPoint
		var mode, renderColor, renderWidth, renderOpacity, lodLevel sql.NullString

		err := rows.Scan(
			&p.ID, &p.DataTime, &p.Longitude, &p.Latitude, &p.Heading, &p.Speed, &p.Altitude,
			&mode, &renderColor, &renderWidth, &renderOpacity, &lodLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track point: %w", err)
		}

		// Handle nullable fields
		if mode.Valid {
			// Store mode in a custom field if needed
		}

		points = append(points, p)
	}

	return points, nil
}

// GetTimeSliceData retrieves aggregated data for time axis filtering
func (r *VisualizationRepository) GetTimeSliceData(startTime, endTime int64, granularity string) (map[string]interface{}, error) {
	var timeFormat string
	switch granularity {
	case "day":
		timeFormat = "%Y-%m-%d"
	case "month":
		timeFormat = "%Y-%m"
	case "year":
		timeFormat = "%Y"
	default:
		timeFormat = "%Y-%m-%d"
	}

	query := fmt.Sprintf(`SELECT
		strftime('%s', datetime(dataTime, 'unixepoch')) as time_slice,
		COUNT(*) as point_count,
		COUNT(DISTINCT mode) as mode_count,
		MIN(dataTime) as min_time,
		MAX(dataTime) as max_time
		FROM "一生足迹"
		WHERE dataTime >= ? AND dataTime <= ?
		AND (outlier_flag IS NULL OR outlier_flag = 0)
		GROUP BY time_slice
		ORDER BY time_slice`, timeFormat)

	rows, err := r.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query time slice data: %w", err)
	}
	defer rows.Close()

	var slices []map[string]interface{}
	for rows.Next() {
		var timeSlice string
		var pointCount, modeCount int
		var minTime, maxTime int64

		err := rows.Scan(&timeSlice, &pointCount, &modeCount, &minTime, &maxTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time slice: %w", err)
		}

		slices = append(slices, map[string]interface{}{
			"time_slice":  timeSlice,
			"point_count": pointCount,
			"mode_count":  modeCount,
			"min_time":    minTime,
			"max_time":    maxTime,
		})
	}

	return map[string]interface{}{
		"granularity": granularity,
		"slices":      slices,
		"count":       len(slices),
	}, nil
}
