package models

import "time"

// AnalysisTask represents an analysis task for trajectory processing
type AnalysisTask struct {
	ID int64 `json:"id" db:"id"`

	// Task identification
	SkillName string `json:"skill_name" db:"skill_name"` // Which skill to run
	TaskType  string `json:"task_type" db:"mode"`        // INCREMENTAL, FULL_RECOMPUTE

	// Status
	Status          string `json:"status" db:"status"`                       // pending, running, completed, failed
	ProgressPercent int    `json:"progress_percent" db:"progress_percent"`
	ETASeconds      int    `json:"eta_seconds,omitempty" db:"eta_seconds"`

	// Input parameters
	ParamsJSON         string `json:"params_json,omitempty" db:"params_json"`
	ThresholdProfileID int64  `json:"threshold_profile_id,omitempty" db:"threshold_profile_id"`

	// Execution info
	TotalPoints     int   `json:"total_points,omitempty" db:"total_points"`
	ProcessedPoints int   `json:"processed_points" db:"processed_points"`
	FailedPoints    int   `json:"failed_points" db:"failed_points"`
	StartTime       int64 `json:"start_time,omitempty" db:"start_time"`     // Unix timestamp
	EndTime         int64 `json:"end_time,omitempty" db:"end_time"`         // Unix timestamp

	// Results
	ResultSummary string `json:"result_summary,omitempty" db:"result_summary"` // JSON object with summary statistics
	ErrorMessage  string `json:"error_message,omitempty" db:"error_message"`

	// Dependencies
	DependsOnTaskIDs string `json:"depends_on_task_ids,omitempty" db:"depends_on_task_ids"` // JSON array of task IDs
	BlocksTaskIDs    string `json:"blocks_task_ids,omitempty" db:"blocks_task_ids"`         // JSON array of task IDs

	// Metadata
	CreatedBy string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TaskType constants
const (
	TaskTypeIncremental   = "INCREMENTAL"
	TaskTypeFullRecompute = "FULL_RECOMPUTE"
)

// TaskStatus constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)
