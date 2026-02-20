package models

import "time"

// Trip represents a trip construction result (origin-destination pair)
type Trip struct {
	ID int64 `json:"id" db:"id"`

	// Trip identification
	Date       string `json:"date" db:"date"`               // YYYY-MM-DD
	TripNumber int    `json:"trip_number" db:"trip_number"` // 1st, 2nd, 3rd trip of the day

	// Temporal info
	StartTime       int64 `json:"start_time" db:"start_time"`             // Unix timestamp
	EndTime         int64 `json:"end_time" db:"end_time"`                 // Unix timestamp
	DurationSeconds int64 `json:"duration_seconds" db:"duration_seconds"` // Duration in seconds

	// Origin and destination
	OriginStayID int64   `json:"origin_stay_id,omitempty" db:"origin_stay_id"` // Foreign key to stay_segments
	DestStayID   int64   `json:"dest_stay_id,omitempty" db:"dest_stay_id"`     // Foreign key to stay_segments
	OriginLat    float64 `json:"origin_lat,omitempty" db:"origin_lat"`
	OriginLon    float64 `json:"origin_lon,omitempty" db:"origin_lon"`
	DestLat      float64 `json:"dest_lat,omitempty" db:"dest_lat"`
	DestLon      float64 `json:"dest_lon,omitempty" db:"dest_lon"`

	// Administrative divisions
	OriginProvince string `json:"origin_province,omitempty" db:"origin_province"`
	OriginCity     string `json:"origin_city,omitempty" db:"origin_city"`
	OriginCounty   string `json:"origin_county,omitempty" db:"origin_county"`
	DestProvince   string `json:"dest_province,omitempty" db:"dest_province"`
	DestCity       string `json:"dest_city,omitempty" db:"dest_city"`
	DestCounty     string `json:"dest_county,omitempty" db:"dest_county"`

	// Trip characteristics
	DistanceMeters float64 `json:"distance_meters,omitempty" db:"distance_meters"`
	AvgSpeedKmh    float64 `json:"avg_speed_kmh,omitempty" db:"avg_speed_kmh"`
	MaxSpeedKmh    float64 `json:"max_speed_kmh,omitempty" db:"max_speed_kmh"`
	PrimaryMode    string  `json:"primary_mode,omitempty" db:"primary_mode"` // Dominant transport mode

	// Segments involved
	ModesJSON      string `json:"modes_json,omitempty" db:"modes_json"`           // JSON array of transport modes
	SegmentIDsJSON string `json:"segment_ids_json,omitempty" db:"segment_ids_json"` // JSON array of segment IDs

	// Trip type
	TripType    string `json:"trip_type,omitempty" db:"trip_type"` // INTRA_CITY, INTER_CITY, INTER_PROVINCE
	IsRoundTrip bool   `json:"is_round_trip" db:"is_round_trip"`

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TripType constants
const (
	TripTypeIntraCity     = "INTRA_CITY"
	TripTypeInterCity     = "INTER_CITY"
	TripTypeInterProvince = "INTER_PROVINCE"
)

// TripsResponse represents a paginated response of trips
type TripsResponse struct {
	Data       []Trip `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}

// TripFilter is defined in filters.go
