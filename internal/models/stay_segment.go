package models

import "time"

// StaySegment represents a stay detection result
type StaySegment struct {
	ID int64 `json:"id" db:"id"`

	// Stay identification
	StayType     string `json:"stay_type" db:"stay_type"`           // SPATIAL, ADMIN
	FirstPointID int64  `json:"first_point_id" db:"first_point_id"` // Foreign key to track point
	LastPointID  int64  `json:"last_point_id" db:"last_point_id"`   // Foreign key to track point

	// Temporal info
	StartTime       int64 `json:"start_time" db:"start_time"`             // Unix timestamp
	EndTime         int64 `json:"end_time" db:"end_time"`                 // Unix timestamp
	DurationSeconds int64 `json:"duration_seconds" db:"duration_seconds"` // Duration in seconds

	// Spatial info (center point)
	CenterLat    float64 `json:"center_lat" db:"center_lat"`
	CenterLon    float64 `json:"center_lon" db:"center_lon"`
	RadiusMeters float64 `json:"radius_meters,omitempty" db:"radius_meters"`

	// Administrative divisions
	Province string `json:"province,omitempty" db:"province"`
	City     string `json:"city,omitempty" db:"city"`
	County   string `json:"county,omitempty" db:"county"`
	Town     string `json:"town,omitempty" db:"town"`
	Village  string `json:"village,omitempty" db:"village"`

	// Stay characteristics
	PointCount              int     `json:"point_count,omitempty" db:"point_count"`
	AvgAccuracy             float64 `json:"avg_accuracy,omitempty" db:"avg_accuracy"`
	MaxDistanceFromCenter   float64 `json:"max_distance_from_center,omitempty" db:"max_distance_from_center"`

	// Semantic annotation
	StayLabel    string  `json:"stay_label,omitempty" db:"stay_label"`       // User-defined or system-generated label
	StayCategory string  `json:"stay_category,omitempty" db:"stay_category"` // HOME, WORK, TRANSIT, LEISURE, etc.
	Confidence   float64 `json:"confidence,omitempty" db:"confidence"`       // 0~1

	// Metadata
	Metadata    string    `json:"metadata,omitempty" db:"metadata"`         // JSON metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// StayType constants
const (
	StayTypeSpatial = "SPATIAL" // Spatial clustering-based stay
	StayTypeAdmin   = "ADMIN"   // Administrative division-based stay
)

// StayCategory constants
const (
	StayCategoryHome    = "HOME"
	StayCategoryWork    = "WORK"
	StayCategoryTransit = "TRANSIT"
	StayCategoryLeisure = "LEISURE"
	StayCategoryUnknown = "UNKNOWN"
)
