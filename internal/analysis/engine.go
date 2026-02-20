package analysis

import (
	"context"
	"database/sql"
)

// Analyzer is the interface that all analysis skills must implement
type Analyzer interface {
	// Analyze performs the analysis for a given task
	// taskID: the analysis task ID
	// mode: "incremental" or "full"
	Analyze(ctx context.Context, taskID int64, mode string) error

	// GetProgress returns the current progress of the analysis
	GetProgress(taskID int64) (*Progress, error)

	// GetName returns the name of the analyzer
	GetName() string
}

// Progress represents the progress of an analysis task
type Progress struct {
	Processed  int     // Number of records processed
	Total      int     // Total number of records to process
	Failed     int     // Number of failed records
	Percent    float64 // Progress percentage (0-100)
	ETASeconds int     // Estimated time to completion in seconds
	Message    string  // Optional progress message
}

// BaseAnalyzer provides common functionality for all analyzers
type BaseAnalyzer struct {
	DB   *sql.DB
	Name string
}

// NewBaseAnalyzer creates a new base analyzer
func NewBaseAnalyzer(db *sql.DB, name string) *BaseAnalyzer {
	return &BaseAnalyzer{
		DB:   db,
		Name: name,
	}
}

// GetName returns the analyzer name
func (a *BaseAnalyzer) GetName() string {
	return a.Name
}

// UpdateTaskProgress updates the progress of an analysis task in the database
func (a *BaseAnalyzer) UpdateTaskProgress(taskID int64, processed, total, failed int) error {
	percent := 0.0
	if total > 0 {
		percent = float64(processed) / float64(total) * 100.0
	}

	query := `
		UPDATE analysis_tasks
		SET processed_points = ?,
		    total_points = ?,
		    failed_points = ?,
		    progress_percent = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := a.DB.Exec(query, processed, total, failed, percent, taskID)
	return err
}

// MarkTaskAsRunning marks a task as running
func (a *BaseAnalyzer) MarkTaskAsRunning(taskID int64) error {
	query := `
		UPDATE analysis_tasks
		SET status = 'running',
		    started_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := a.DB.Exec(query, taskID)
	return err
}

// MarkTaskAsCompleted marks a task as completed
func (a *BaseAnalyzer) MarkTaskAsCompleted(taskID int64) error {
	query := `
		UPDATE analysis_tasks
		SET status = 'completed',
		    progress_percent = 100,
		    completed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := a.DB.Exec(query, taskID)
	return err
}

// MarkTaskAsFailed marks a task as failed with an error message
func (a *BaseAnalyzer) MarkTaskAsFailed(taskID int64, errorMsg string) error {
	query := `
		UPDATE analysis_tasks
		SET status = 'failed',
		    error_message = ?,
		    completed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := a.DB.Exec(query, errorMsg, taskID)
	return err
}

// GetTaskInfo retrieves task information from the database
func (a *BaseAnalyzer) GetTaskInfo(taskID int64) (*TaskInfo, error) {
	query := `
		SELECT id, skill_name, task_type, status, progress_percent,
		       total_points, processed_points, failed_points,
		       params_json, created_at, started_at, completed_at
		FROM analysis_tasks
		WHERE id = ?
	`

	var info TaskInfo
	var startedAt, completedAt sql.NullTime

	err := a.DB.QueryRow(query, taskID).Scan(
		&info.ID, &info.SkillName, &info.TaskType, &info.Status,
		&info.ProgressPercent, &info.TotalPoints, &info.ProcessedPoints,
		&info.FailedPoints, &info.ParamsJSON, &info.CreatedAt,
		&startedAt, &completedAt,
	)

	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		timeStr := startedAt.Time.Format("2006-01-02 15:04:05")
		info.StartedAt = &timeStr
	}
	if completedAt.Valid {
		timeStr := completedAt.Time.Format("2006-01-02 15:04:05")
		info.CompletedAt = &timeStr
	}

	return &info, nil
}

// TaskInfo contains information about an analysis task
type TaskInfo struct {
	ID              int64
	SkillName       string
	TaskType        string
	Status          string
	ProgressPercent float64
	TotalPoints     int
	ProcessedPoints int
	FailedPoints    int
	ParamsJSON      string
	CreatedAt       string
	StartedAt       *string
	CompletedAt     *string
}

// AnalyzerFactory is a function that creates an analyzer instance
type AnalyzerFactory func(db *sql.DB) Analyzer

// AnalyzerRegistry maps skill names to analyzer factories
var AnalyzerRegistry = make(map[string]AnalyzerFactory)

// RegisterAnalyzer registers an analyzer factory for a skill name
func RegisterAnalyzer(skillName string, factory AnalyzerFactory) {
	AnalyzerRegistry[skillName] = factory
}

// GetAnalyzer retrieves an analyzer instance for a skill name
func GetAnalyzer(skillName string, db *sql.DB) Analyzer {
	factory, ok := AnalyzerRegistry[skillName]
	if !ok {
		return nil
	}
	return factory(db)
}

// IsGoNativeSkill checks if a skill is implemented in Go (vs Python)
func IsGoNativeSkill(skillName string) bool {
	_, ok := AnalyzerRegistry[skillName]
	return ok
}
