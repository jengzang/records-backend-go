package models

// Trip represents a trip constructed from stays
type Trip struct {
	ID        int64  `json:"id" db:"id"`
	StartTime int64  `json:"startTime" db:"start_time"`  // Unix timestamp
	EndTime   int64  `json:"endTime" db:"end_time"`      // Unix timestamp
	Duration  int64  `json:"duration" db:"duration"`     // Duration in seconds
	Distance  float64 `json:"distance" db:"distance"`    // Total distance in meters

	// Origin
	OriginLongitude float64 `json:"originLongitude" db:"origin_longitude"`
	OriginLatitude  float64 `json:"originLatitude" db:"origin_latitude"`
	OriginProvince  string  `json:"originProvince,omitempty" db:"origin_province"`
	OriginCity      string  `json:"originCity,omitempty" db:"origin_city"`
	OriginCounty    string  `json:"originCounty,omitempty" db:"origin_county"`

	// Destination
	DestLongitude float64 `json:"destLongitude" db:"dest_longitude"`
	DestLatitude  float64 `json:"destLatitude" db:"dest_latitude"`
	DestProvince  string  `json:"destProvince,omitempty" db:"dest_province"`
	DestCity      string  `json:"destCity,omitempty" db:"dest_city"`
	DestCounty    string  `json:"destCounty,omitempty" db:"dest_county"`

	// Trip classification
	TripType       string  `json:"tripType,omitempty" db:"trip_type"`             // e.g., "commute", "travel", "errand"
	TransportMode  string  `json:"transportMode,omitempty" db:"transport_mode"`   // e.g., "walk", "bike", "car", "train", "plane"
	Confidence     float64 `json:"confidence,omitempty" db:"confidence"`          // 0.0 to 1.0
	PointCount     int     `json:"pointCount,omitempty" db:"point_count"`         // Number of GPS points in this trip

	// Metadata
	CreatedAt   string `json:"createdAt,omitempty" db:"created_at"`
	AlgoVersion string `json:"algoVersion,omitempty" db:"algo_version"`
}

// TripsResponse represents a paginated response of trips
type TripsResponse struct {
	Data       []Trip `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}

// TripFilter represents filter parameters for querying trips
type TripFilter struct {
	StartTime     int64   `form:"startTime"`
	EndTime       int64   `form:"endTime"`
	OriginCity    string  `form:"originCity"`
	DestCity      string  `form:"destCity"`
	MinDistance   float64 `form:"minDistance"`
	TransportMode string  `form:"transportMode"`
	TripType      string  `form:"tripType"`
	Page          int     `form:"page"`
	PageSize      int     `form:"pageSize"`
}
