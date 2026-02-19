package models

import "time"

// ThresholdProfile represents a threshold configuration for analysis skills
type ThresholdProfile struct {
	ID int64 `json:"id" db:"id"`

	// Profile identification
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	IsDefault   bool   `json:"is_default" db:"is_default"`

	// Skill name
	SkillName string `json:"skill_name" db:"skill_name"` // e.g., "stay_detection", "transport_mode"

	// Parameters (JSON)
	ParamsJSON string `json:"params_json" db:"params_json"` // JSON object with all parameters

	// Metadata
	CreatedBy string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
