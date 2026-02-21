package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// GridRepository handles database operations for grid cells
type GridRepository struct {
	db *sql.DB
}

// NewGridRepository creates a new grid repository
func NewGridRepository(db *sql.DB) *GridRepository {
	return &GridRepository{db: db}
}

// GetGridCells retrieves grid cells with filtering
func (r *GridRepository) GetGridCells(filter models.GridFilter) ([]models.GridCell, error) {
	// Build query
	query := `SELECT grid_id, level, center_lat, center_lon,
		bbox_min_lat, bbox_max_lat, bbox_min_lon, bbox_max_lon,
		point_count, visit_count, first_visit, last_visit,
		total_duration_s, modes,
		created_at, updated_at
		FROM grid_cells`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.Level > 0 {
		conditions = append(conditions, "level = ?")
		args = append(args, filter.Level)
	}
	if filter.MinLat != 0 || filter.MaxLat != 0 || filter.MinLon != 0 || filter.MaxLon != 0 {
		// Filter by bounding box
		if filter.MinLat != 0 {
			conditions = append(conditions, "center_lat >= ?")
			args = append(args, filter.MinLat)
		}
		if filter.MaxLat != 0 {
			conditions = append(conditions, "center_lat <= ?")
			args = append(args, filter.MaxLat)
		}
		if filter.MinLon != 0 {
			conditions = append(conditions, "center_lon >= ?")
			args = append(args, filter.MinLon)
		}
		if filter.MaxLon != 0 {
			conditions = append(conditions, "center_lon <= ?")
			args = append(args, filter.MaxLon)
		}
	}
	if filter.MinDensity > 0 {
		conditions = append(conditions, "point_count >= ?")
		args = append(args, filter.MinDensity)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by point count descending (hottest cells first)
	query += " ORDER BY point_count DESC"

	// Limit to 10000 cells max for performance
	query += " LIMIT 10000"

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query grid cells: %w", err)
	}
	defer rows.Close()

	var cells []models.GridCell
	for rows.Next() {
		var c models.GridCell
		err := rows.Scan(
			&c.GridID, &c.Level, &c.CenterLat, &c.CenterLon,
			&c.MinLat, &c.MaxLat, &c.MinLon, &c.MaxLon,
			&c.PointCount, &c.VisitCount, &c.FirstVisit, &c.LastVisit,
			&c.TotalDurationSeconds, &c.ModesJSON,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grid cell: %w", err)
		}
		cells = append(cells, c)
	}

	return cells, nil
}

// GetGridCellByID retrieves a single grid cell by grid_id
func (r *GridRepository) GetGridCellByID(id int64) (*models.GridCell, error) {
	// Note: This method signature uses int64 for compatibility, but grid_id is actually TEXT
	// Convert int64 to string for the query
	gridID := fmt.Sprintf("%d", id)

	query := `SELECT grid_id, level, center_lat, center_lon,
		bbox_min_lat, bbox_max_lat, bbox_min_lon, bbox_max_lon,
		point_count, visit_count, first_visit, last_visit,
		total_duration_s, modes,
		created_at, updated_at
		FROM grid_cells WHERE grid_id = ?`

	var c models.GridCell
	err := r.db.QueryRow(query, gridID).Scan(
		&c.GridID, &c.Level, &c.CenterLat, &c.CenterLon,
		&c.MinLat, &c.MaxLat, &c.MinLon, &c.MaxLon,
		&c.PointCount, &c.VisitCount, &c.FirstVisit, &c.LastVisit,
		&c.TotalDurationSeconds, &c.ModesJSON,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get grid cell: %w", err)
	}

	return &c, nil
}
