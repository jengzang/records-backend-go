package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jengzang/records-backend-go/internal/models"
)

// AnalysisTaskRepository handles database operations for analysis tasks
type AnalysisTaskRepository struct {
	db *sql.DB
}

// NewAnalysisTaskRepository creates a new analysis task repository
func NewAnalysisTaskRepository(db *sql.DB) *AnalysisTaskRepository {
	return &AnalysisTaskRepository{db: db}
}

// Create creates a new analysis task
func (r *AnalysisTaskRepository) Create(task *models.AnalysisTask) error {
	query := `
		INSERT INTO analysis_tasks (
			skill_name, task_type, status, progress_percent, eta_seconds,
			params_json, threshold_profile_id, total_points, processed_points,
			failed_points, start_time, end_time, result_summary, error_message,
			depends_on_task_ids, blocks_task_ids, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		task.SkillName,
		task.TaskType,
		task.Status,
		task.ProgressPercent,
		task.ETASeconds,
		task.ParamsJSON,
		task.ThresholdProfileID,
		task.TotalPoints,
		task.ProcessedPoints,
		task.FailedPoints,
		task.StartTime,
		task.EndTime,
		task.ResultSummary,
		task.ErrorMessage,
		task.DependsOnTaskIDs,
		task.BlocksTaskIDs,
		task.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create analysis task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = id
	return nil
}

// GetByID retrieves an analysis task by ID
func (r *AnalysisTaskRepository) GetByID(id int64) (*models.AnalysisTask, error) {
	query := `
		SELECT id, skill_name, task_type, status, progress_percent, eta_seconds,
			   params_json, threshold_profile_id, total_points, processed_points,
			   failed_points, start_time, end_time, result_summary, error_message,
			   depends_on_task_ids, blocks_task_ids, created_by, created_at, updated_at
		FROM analysis_tasks
		WHERE id = ?
	`

	task := &models.AnalysisTask{}
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.SkillName,
		&task.TaskType,
		&task.Status,
		&task.ProgressPercent,
		&task.ETASeconds,
		&task.ParamsJSON,
		&task.ThresholdProfileID,
		&task.TotalPoints,
		&task.ProcessedPoints,
		&task.FailedPoints,
		&task.StartTime,
		&task.EndTime,
		&task.ResultSummary,
		&task.ErrorMessage,
		&task.DependsOnTaskIDs,
		&task.BlocksTaskIDs,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("analysis task not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis task: %w", err)
	}

	return task, nil
}

// List retrieves analysis tasks with optional filters
func (r *AnalysisTaskRepository) List(skillName string, status string, limit int, offset int) ([]*models.AnalysisTask, error) {
	query := `
		SELECT id, skill_name, task_type, status, progress_percent, eta_seconds,
			   params_json, threshold_profile_id, total_points, processed_points,
			   failed_points, start_time, end_time, result_summary, error_message,
			   depends_on_task_ids, blocks_task_ids, created_by, created_at, updated_at
		FROM analysis_tasks
		WHERE 1=1
	`

	args := []interface{}{}
	if skillName != "" {
		query += " AND skill_name = ?"
		args = append(args, skillName)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list analysis tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.AnalysisTask
	for rows.Next() {
		task := &models.AnalysisTask{}
		err := rows.Scan(
			&task.ID,
			&task.SkillName,
			&task.TaskType,
			&task.Status,
			&task.ProgressPercent,
			&task.ETASeconds,
			&task.ParamsJSON,
			&task.ThresholdProfileID,
			&task.TotalPoints,
			&task.ProcessedPoints,
			&task.FailedPoints,
			&task.StartTime,
			&task.EndTime,
			&task.ResultSummary,
			&task.ErrorMessage,
			&task.DependsOnTaskIDs,
			&task.BlocksTaskIDs,
			&task.CreatedBy,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update updates an analysis task
func (r *AnalysisTaskRepository) Update(task *models.AnalysisTask) error {
	query := `
		UPDATE analysis_tasks
		SET status = ?, progress_percent = ?, eta_seconds = ?,
			processed_points = ?, failed_points = ?, start_time = ?,
			end_time = ?, result_summary = ?, error_message = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		task.Status,
		task.ProgressPercent,
		task.ETASeconds,
		task.ProcessedPoints,
		task.FailedPoints,
		task.StartTime,
		task.EndTime,
		task.ResultSummary,
		task.ErrorMessage,
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update analysis task: %w", err)
	}

	return nil
}

// UpdateProgress updates the progress of an analysis task
func (r *AnalysisTaskRepository) UpdateProgress(id int64, processedPoints int, failedPoints int, progressPercent int, etaSeconds int) error {
	query := `
		UPDATE analysis_tasks
		SET processed_points = ?, failed_points = ?, progress_percent = ?,
			eta_seconds = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, processedPoints, failedPoints, progressPercent, etaSeconds, id)
	if err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	return nil
}

// MarkAsRunning marks a task as running
func (r *AnalysisTaskRepository) MarkAsRunning(id int64) error {
	now := time.Now().Unix()
	query := `
		UPDATE analysis_tasks
		SET status = ?, start_time = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, models.TaskStatusRunning, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	return nil
}

// MarkAsCompleted marks a task as completed with result summary
func (r *AnalysisTaskRepository) MarkAsCompleted(id int64, resultSummary string) error {
	now := time.Now().Unix()
	query := `
		UPDATE analysis_tasks
		SET status = ?, end_time = ?, result_summary = ?,
			progress_percent = 100, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, models.TaskStatusCompleted, now, resultSummary, id)
	if err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	return nil
}

// MarkAsFailed marks a task as failed with an error message
func (r *AnalysisTaskRepository) MarkAsFailed(id int64, errorMessage string) error {
	now := time.Now().Unix()
	query := `
		UPDATE analysis_tasks
		SET status = ?, end_time = ?, error_message = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, models.TaskStatusFailed, now, errorMessage, id)
	if err != nil {
		return fmt.Errorf("failed to mark task as failed: %w", err)
	}

	return nil
}

// CountUnanalyzedPoints counts the number of points without analysis data
func (r *AnalysisTaskRepository) CountUnanalyzedPoints() (int, error) {
	query := `SELECT COUNT(*) FROM "一生足迹" WHERE segment_id IS NULL`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unanalyzed points: %w", err)
	}

	return count, nil
}

// CountAllPoints counts the total number of points
func (r *AnalysisTaskRepository) CountAllPoints() (int, error) {
	query := `SELECT COUNT(*) FROM "一生足迹"`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count all points: %w", err)
	}

	return count, nil
}
