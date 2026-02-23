package models

import "time"

// ImportTask represents a data import task
type ImportTask struct {
	ID               int64      `json:"id" db:"id"`
	Status           string     `json:"status" db:"status"` // pending, running, completed, failed
	FilePath         *string    `json:"file_path,omitempty" db:"file_path"`
	FileName         string     `json:"file_name" db:"file_name"`
	FileSize         int64      `json:"file_size" db:"file_size"`
	Mode             string     `json:"mode" db:"mode"`                         // append, replace
	Deduplicate      bool       `json:"deduplicate" db:"deduplicate"`
	AutoTrigger      bool       `json:"auto_trigger" db:"auto_trigger"`         // Auto trigger geocoding
	TotalRecords     int        `json:"total_records" db:"total_records"`
	NewRecords       int        `json:"new_records" db:"new_records"`
	DuplicateRecords int        `json:"duplicate_records" db:"duplicate_records"`
	ErrorMessage     *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}
