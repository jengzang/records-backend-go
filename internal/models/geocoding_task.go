package models

import "time"

// GeocodingTask represents a geocoding batch processing task
type GeocodingTask struct {
	ID              int       `json:"id" db:"id"`
	Status          string    `json:"status" db:"status"` // pending, running, completed, failed
	TotalPoints     int       `json:"total_points" db:"total_points"`
	ProcessedPoints int       `json:"processed_points" db:"processed_points"`
	FailedPoints    int       `json:"failed_points" db:"failed_points"`
	StartTime       *time.Time `json:"start_time,omitempty" db:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty" db:"end_time"`
	ETASeconds      *int      `json:"eta_seconds,omitempty" db:"eta_seconds"`
	ErrorMessage    *string   `json:"error_message,omitempty" db:"error_message"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// TaskStatus constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// IsTerminal returns true if the task is in a terminal state
func (t *GeocodingTask) IsTerminal() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// Progress returns the completion percentage (0-100)
func (t *GeocodingTask) Progress() float64 {
	if t.TotalPoints == 0 {
		return 0
	}
	return float64(t.ProcessedPoints) / float64(t.TotalPoints) * 100
}
