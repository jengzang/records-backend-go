package models

import "time"

// TrackPoint represents a GPS track point with administrative division information
type TrackPoint struct {
	ID           int64   `json:"id" db:"id"`
	DataTime     int64   `json:"dataTime" db:"dataTime"`           // Unix timestamp in seconds
	Longitude    float64 `json:"longitude" db:"longitude"`
	Latitude     float64 `json:"latitude" db:"latitude"`
	Heading      float64 `json:"heading" db:"heading"`
	Accuracy     float64 `json:"accuracy" db:"accuracy"`
	Speed        float64 `json:"speed" db:"speed"`
	Distance     float64 `json:"distance" db:"distance"`
	Altitude     float64 `json:"altitude" db:"altitude"`
	TimeVisually string  `json:"timeVisually" db:"time_visually"`  // Format: 2025/01/22 21:42:18.000
	Time         string  `json:"time" db:"time"`                   // Format: 20250122214218

	// Administrative divisions
	Province string `json:"province,omitempty" db:"province"`     // 省级
	City     string `json:"city,omitempty" db:"city"`             // 市级
	County   string `json:"county,omitempty" db:"county"`         // 区县级
	Town     string `json:"town,omitempty" db:"town"`             // 乡镇级
	Village  string `json:"village,omitempty" db:"village"`       // 村级/街道级

	// Metadata
	CreatedAt   *string `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt   *string `json:"updatedAt,omitempty" db:"updated_at"`
	AlgoVersion *string `json:"algoVersion,omitempty" db:"algo_version"`
}

// TrackPointsResponse represents a paginated response of track points
type TrackPointsResponse struct {
	Data       []TrackPoint `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	TotalPages int          `json:"totalPages"`
}

// TrackPointFilter represents filter parameters for querying track points
type TrackPointFilter struct {
	StartTime int64   `form:"startTime"`  // Unix timestamp
	EndTime   int64   `form:"endTime"`    // Unix timestamp
	Province  string  `form:"province"`
	City      string  `form:"city"`
	County    string  `form:"county"`
	MinSpeed  float64 `form:"minSpeed"`
	MaxSpeed  float64 `form:"maxSpeed"`
	Page      int     `form:"page"`
	PageSize  int     `form:"pageSize"`
}
