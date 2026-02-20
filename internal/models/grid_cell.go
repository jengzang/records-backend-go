package models

import "time"

// GridCell represents a spatial grid cell for heatmap and spatial analysis
type GridCell struct {
	ID int64 `json:"id" db:"id"`

	// Grid identification
	GridID string `json:"grid_id" db:"grid_id"` // Format: "L{level}_{x}_{y}"
	Level  int    `json:"level" db:"level"`     // 1-15 (zoom level)
	X      int    `json:"x" db:"x"`             // Grid X coordinate
	Y      int    `json:"y" db:"y"`             // Grid Y coordinate

	// Center point
	CenterLat float64 `json:"center_lat" db:"center_lat"`
	CenterLon float64 `json:"center_lon" db:"center_lon"`

	// Bounding box
	MinLat float64 `json:"min_lat" db:"min_lat"`
	MaxLat float64 `json:"max_lat" db:"max_lat"`
	MinLon float64 `json:"min_lon" db:"min_lon"`
	MaxLon float64 `json:"max_lon" db:"max_lon"`

	// Statistics
	PointCount          int    `json:"point_count" db:"point_count"`
	VisitCount          int    `json:"visit_count" db:"visit_count"`                     // Number of distinct visits
	TotalDurationSeconds int64  `json:"total_duration_seconds" db:"total_duration_seconds"`
	FirstVisit          int64  `json:"first_visit,omitempty" db:"first_visit"`           // Unix timestamp (alias for FirstVisitTime)
	LastVisit           int64  `json:"last_visit,omitempty" db:"last_visit"`             // Unix timestamp (alias for LastVisitTime)
	FirstVisitTime      int64  `json:"first_visit_time,omitempty" db:"first_visit_time"` // Unix timestamp
	LastVisitTime       int64  `json:"last_visit_time,omitempty" db:"last_visit_time"`   // Unix timestamp
	ModesJSON           string `json:"modes_json,omitempty" db:"modes_json"`             // JSON array of transport modes

	// Movement characteristics
	AvgSpeedKmh  float64 `json:"avg_speed_kmh,omitempty" db:"avg_speed_kmh"`
	MaxSpeedKmh  float64 `json:"max_speed_kmh,omitempty" db:"max_speed_kmh"`
	DominantMode string  `json:"dominant_mode,omitempty" db:"dominant_mode"` // Most common transport mode

	// Administrative division
	Province string `json:"province,omitempty" db:"province"`
	City     string `json:"city,omitempty" db:"city"`
	County   string `json:"county,omitempty" db:"county"`

	// Density metrics
	DensityScore float64 `json:"density_score,omitempty" db:"density_score"` // Normalized 0~1
	RevisitScore float64 `json:"revisit_score,omitempty" db:"revisit_score"` // Normalized 0~1

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
