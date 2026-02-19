package models

import "time"

// FootprintStatistics represents aggregated footprint statistics
type FootprintStatistics struct {
	ID int64 `json:"id" db:"id"`

	// Aggregation dimensions
	StatType  string `json:"stat_type" db:"stat_type"`   // PROVINCE, CITY, COUNTY, TOWN, GRID
	StatKey   string `json:"stat_key" db:"stat_key"`     // Province/city/county/town name or grid_id
	TimeRange string `json:"time_range,omitempty" db:"time_range"` // YYYY, YYYY-MM, YYYY-MM-DD, or ALL

	// Spatial info
	Province string `json:"province,omitempty" db:"province"`
	City     string `json:"city,omitempty" db:"city"`
	County   string `json:"county,omitempty" db:"county"`
	Town     string `json:"town,omitempty" db:"town"`

	// Statistics
	PointCount          int     `json:"point_count" db:"point_count"`
	VisitCount          int     `json:"visit_count" db:"visit_count"`
	TotalDistanceMeters float64 `json:"total_distance_meters" db:"total_distance_meters"`
	TotalDurationSeconds int64  `json:"total_duration_seconds" db:"total_duration_seconds"`
	FirstVisitTime      int64   `json:"first_visit_time,omitempty" db:"first_visit_time"` // Unix timestamp
	LastVisitTime       int64   `json:"last_visit_time,omitempty" db:"last_visit_time"`   // Unix timestamp

	// Rankings
	RankByPoints   int `json:"rank_by_points,omitempty" db:"rank_by_points"`
	RankByVisits   int `json:"rank_by_visits,omitempty" db:"rank_by_visits"`
	RankByDuration int `json:"rank_by_duration,omitempty" db:"rank_by_duration"`

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// StayStatistics represents aggregated stay statistics
type StayStatistics struct {
	ID int64 `json:"id" db:"id"`

	// Aggregation dimensions
	StatType  string `json:"stat_type" db:"stat_type"`   // PROVINCE, CITY, COUNTY, CATEGORY
	StatKey   string `json:"stat_key" db:"stat_key"`
	TimeRange string `json:"time_range,omitempty" db:"time_range"`

	// Spatial info
	Province string `json:"province,omitempty" db:"province"`
	City     string `json:"city,omitempty" db:"city"`
	County   string `json:"county,omitempty" db:"county"`

	// Statistics
	StayCount            int     `json:"stay_count" db:"stay_count"`
	TotalDurationSeconds int64   `json:"total_duration_seconds" db:"total_duration_seconds"`
	AvgDurationSeconds   float64 `json:"avg_duration_seconds,omitempty" db:"avg_duration_seconds"`
	MaxDurationSeconds   int64   `json:"max_duration_seconds,omitempty" db:"max_duration_seconds"`
	StayCategory         string  `json:"stay_category,omitempty" db:"stay_category"` // For CATEGORY stat_type

	// Rankings
	RankByCount    int `json:"rank_by_count,omitempty" db:"rank_by_count"`
	RankByDuration int `json:"rank_by_duration,omitempty" db:"rank_by_duration"`

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ExtremeEvent represents an extreme event (max altitude, speed, etc.)
type ExtremeEvent struct {
	ID int64 `json:"id" db:"id"`

	// Event type
	EventType     string `json:"event_type" db:"event_type"`         // MAX_ALTITUDE, MAX_SPEED, NORTHMOST, SOUTHMOST, EASTMOST, WESTMOST
	EventCategory string `json:"event_category" db:"event_category"` // SPATIAL, SPEED, ALTITUDE

	// Event details
	PointID    int64   `json:"point_id" db:"point_id"`       // Foreign key to track point
	EventTime  int64   `json:"event_time" db:"event_time"`   // Unix timestamp
	EventValue float64 `json:"event_value" db:"event_value"` // Altitude/speed/latitude/longitude

	// Location
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	Province  string  `json:"province,omitempty" db:"province"`
	City      string  `json:"city,omitempty" db:"city"`
	County    string  `json:"county,omitempty" db:"county"`

	// Context
	Mode      string `json:"mode,omitempty" db:"mode"`           // Transport mode at the time
	SegmentID int64  `json:"segment_id,omitempty" db:"segment_id"` // Foreign key to segments

	// Ranking
	Rank int `json:"rank,omitempty" db:"rank"` // Top N ranking

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// StatType constants
const (
	StatTypeProvince = "PROVINCE"
	StatTypeCity     = "CITY"
	StatTypeCounty   = "COUNTY"
	StatTypeTown     = "TOWN"
	StatTypeGrid     = "GRID"
	StatTypeCategory = "CATEGORY"
)

// EventType constants
const (
	EventTypeMaxAltitude = "MAX_ALTITUDE"
	EventTypeMaxSpeed    = "MAX_SPEED"
	EventTypeNorthmost   = "NORTHMOST"
	EventTypeSouthmost   = "SOUTHMOST"
	EventTypeEastmost    = "EASTMOST"
	EventTypeWestmost    = "WESTMOST"
)

// EventCategory constants
const (
	EventCategorySpatial  = "SPATIAL"
	EventCategorySpeed    = "SPEED"
	EventCategoryAltitude = "ALTITUDE"
)
