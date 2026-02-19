package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jengzang/records-backend-go/internal/models"
)

// GeocodingRepository handles database operations for geocoding tasks
type GeocodingRepository struct {
	db *sql.DB
}

// NewGeocodingRepository creates a new geocoding repository
func NewGeocodingRepository(db *sql.DB) *GeocodingRepository {
	return &GeocodingRepository{db: db}
}

// Create creates a new geocoding task
func (r *GeocodingRepository) Create(task *models.GeocodingTask) error {
	query := `
		INSERT INTO geocoding_tasks (
			status, total_points, processed_points, failed_points,
			start_time, end_time, eta_seconds, error_message, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		task.Status,
		task.TotalPoints,
		task.ProcessedPoints,
		task.FailedPoints,
		task.StartTime,
		task.EndTime,
		task.ETASeconds,
		task.ErrorMessage,
		task.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create geocoding task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = int(id)
	return nil
}

// GetByID retrieves a geocoding task by ID
func (r *GeocodingRepository) GetByID(id int) (*models.GeocodingTask, error) {
	query := `
		SELECT id, status, total_points, processed_points, failed_points,
			   start_time, end_time, eta_seconds, error_message, created_by,
			   created_at, updated_at
		FROM geocoding_tasks
		WHERE id = ?
	`

	task := &models.GeocodingTask{}
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.Status,
		&task.TotalPoints,
		&task.ProcessedPoints,
		&task.FailedPoints,
		&task.StartTime,
		&task.EndTime,
		&task.ETASeconds,
		&task.ErrorMessage,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("geocoding task not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get geocoding task: %w", err)
	}

	return task, nil
}

// List retrieves all geocoding tasks with optional status filter
func (r *GeocodingRepository) List(status string, limit int, offset int) ([]*models.GeocodingTask, error) {
	query := `
		SELECT id, status, total_points, processed_points, failed_points,
			   start_time, end_time, eta_seconds, error_message, created_by,
			   created_at, updated_at
		FROM geocoding_tasks
	`

	args := []interface{}{}
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list geocoding tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.GeocodingTask
	for rows.Next() {
		task := &models.GeocodingTask{}
		err := rows.Scan(
			&task.ID,
			&task.Status,
			&task.TotalPoints,
			&task.ProcessedPoints,
			&task.FailedPoints,
			&task.StartTime,
			&task.EndTime,
			&task.ETASeconds,
			&task.ErrorMessage,
			&task.CreatedBy,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan geocoding task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update updates a geocoding task
func (r *GeocodingRepository) Update(task *models.GeocodingTask) error {
	query := `
		UPDATE geocoding_tasks
		SET status = ?, processed_points = ?, failed_points = ?,
			start_time = ?, end_time = ?, eta_seconds = ?, error_message = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		task.Status,
		task.ProcessedPoints,
		task.FailedPoints,
		task.StartTime,
		task.EndTime,
		task.ETASeconds,
		task.ErrorMessage,
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update geocoding task: %w", err)
	}

	return nil
}

// UpdateProgress updates the progress of a geocoding task
func (r *GeocodingRepository) UpdateProgress(id int, processedPoints int, failedPoints int, etaSeconds *int) error {
	query := `
		UPDATE geocoding_tasks
		SET processed_points = ?, failed_points = ?, eta_seconds = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, processedPoints, failedPoints, etaSeconds, id)
	if err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	return nil
}

// MarkAsRunning marks a task as running
func (r *GeocodingRepository) MarkAsRunning(id int) error {
	now := time.Now()
	query := `
		UPDATE geocoding_tasks
		SET status = ?, start_time = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, models.TaskStatusRunning, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	return nil
}

// MarkAsCompleted marks a task as completed
func (r *GeocodingRepository) MarkAsCompleted(id int) error {
	now := time.Now()
	query := `
		UPDATE geocoding_tasks
		SET status = ?, end_time = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, models.TaskStatusCompleted, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	return nil
}

// MarkAsFailed marks a task as failed with an error message
func (r *GeocodingRepository) MarkAsFailed(id int, errorMessage string) error {
	now := time.Now()
	query := `
		UPDATE geocoding_tasks
		SET status = ?, end_time = ?, error_message = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, models.TaskStatusFailed, now, errorMessage, id)
	if err != nil {
		return fmt.Errorf("failed to mark task as failed: %w", err)
	}

	return nil
}

// CountUngeocodedPoints counts the number of points without geocoding data
func (r *GeocodingRepository) CountUngeocodedPoints() (int, error) {
	query := `SELECT COUNT(*) FROM "一生足迹" WHERE province IS NULL`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ungeocoded points: %w", err)
	}

	return count, nil
}
