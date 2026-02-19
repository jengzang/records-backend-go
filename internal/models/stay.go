package models

// Stay represents a detected stay at a location
type Stay struct {
	ID        int64   `json:"id" db:"id"`
	StartTime int64   `json:"startTime" db:"start_time"`  // Unix timestamp
	EndTime   int64   `json:"endTime" db:"end_time"`      // Unix timestamp
	Duration  int64   `json:"duration" db:"duration"`     // Duration in seconds
	Longitude float64 `json:"longitude" db:"longitude"`
	Latitude  float64 `json:"latitude" db:"latitude"`

	// Administrative divisions
	Province string `json:"province,omitempty" db:"province"`
	City     string `json:"city,omitempty" db:"city"`
	County   string `json:"county,omitempty" db:"county"`
	Town     string `json:"town,omitempty" db:"town"`
	Village  string `json:"village,omitempty" db:"village"`

	// Stay classification
	StayType       string  `json:"stayType,omitempty" db:"stay_type"`             // e.g., "home", "work", "transit"
	Confidence     float64 `json:"confidence,omitempty" db:"confidence"`          // 0.0 to 1.0
	PointCount     int     `json:"pointCount,omitempty" db:"point_count"`         // Number of GPS points in this stay

	// Metadata
	CreatedAt   string `json:"createdAt,omitempty" db:"created_at"`
	AlgoVersion string `json:"algoVersion,omitempty" db:"algo_version"`
}

// StaysResponse represents a paginated response of stays
type StaysResponse struct {
	Data       []Stay `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}

// StayFilter represents filter parameters for querying stays
type StayFilter struct {
	StartTime   int64  `form:"startTime"`
	EndTime     int64  `form:"endTime"`
	Province    string `form:"province"`
	City        string `form:"city"`
	County      string `form:"county"`
	MinDuration int64  `form:"minDuration"`  // Minimum duration in seconds
	StayType    string `form:"stayType"`
	Page        int    `form:"page"`
	PageSize    int    `form:"pageSize"`
}
