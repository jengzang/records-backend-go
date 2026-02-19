package models

import "time"

// Segment represents a behavior segment (transport mode classification result)
type Segment struct {
	ID int64 `json:"id" db:"id"`

	// Segment identification
	Mode         string `json:"mode" db:"mode"`                     // WALK, CAR, TRAIN, FLIGHT, STAY, UNKNOWN
	StartPointID int64  `json:"start_point_id" db:"start_point_id"` // Foreign key to track point
	EndPointID   int64  `json:"end_point_id" db:"end_point_id"`     // Foreign key to track point

	// Temporal info
	StartTime       int64 `json:"start_time" db:"start_time"`             // Unix timestamp
	EndTime         int64 `json:"end_time" db:"end_time"`                 // Unix timestamp
	DurationSeconds int64 `json:"duration_seconds" db:"duration_seconds"` // Duration in seconds

	// Spatial info
	DistanceMeters float64 `json:"distance_meters,omitempty" db:"distance_meters"`
	StartLat       float64 `json:"start_lat,omitempty" db:"start_lat"`
	StartLon       float64 `json:"start_lon,omitempty" db:"start_lon"`
	EndLat         float64 `json:"end_lat,omitempty" db:"end_lat"`
	EndLon         float64 `json:"end_lon,omitempty" db:"end_lon"`

	// Movement characteristics
	AvgSpeedKmh     float64 `json:"avg_speed_kmh,omitempty" db:"avg_speed_kmh"`
	MaxSpeedKmh     float64 `json:"max_speed_kmh,omitempty" db:"max_speed_kmh"`
	AvgHeading      float64 `json:"avg_heading,omitempty" db:"avg_heading"`
	HeadingVariance float64 `json:"heading_variance,omitempty" db:"heading_variance"`

	// Classification confidence
	Confidence  float64 `json:"confidence" db:"confidence"`       // 0~1
	ReasonCodes string  `json:"reason_codes" db:"reason_codes"`   // JSON array of reason codes

	// Administrative divisions
	Province string `json:"province,omitempty" db:"province"`
	City     string `json:"city,omitempty" db:"city"`
	County   string `json:"county,omitempty" db:"county"`

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TransportMode constants
const (
	ModeWalk    = "WALK"
	ModeCar     = "CAR"
	ModeTrain   = "TRAIN"
	ModeFlight  = "FLIGHT"
	ModeStay    = "STAY"
	ModeUnknown = "UNKNOWN"
)
