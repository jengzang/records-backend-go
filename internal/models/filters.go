package models

// SegmentFilter represents filter parameters for querying segments
type SegmentFilter struct {
	Mode         string  `form:"mode"`         // WALK, CAR, TRAIN, FLIGHT, STAY, UNKNOWN
	StartTime    int64   `form:"startTime"`    // Unix timestamp
	EndTime      int64   `form:"endTime"`      // Unix timestamp
	Province     string  `form:"province"`
	City         string  `form:"city"`
	County       string  `form:"county"`
	MinDistance  float64 `form:"minDistance"`  // Meters
	MinDuration  int64   `form:"minDuration"`  // Seconds
	MinConfidence float64 `form:"minConfidence"` // 0-1
	Page         int     `form:"page"`
	PageSize     int     `form:"pageSize"`
}

// StayFilter represents filter parameters for querying stay segments
type StayFilter struct {
	StayType     string  `form:"stayType"`     // SPATIAL, ADMIN
	StayCategory string  `form:"stayCategory"` // HOME, WORK, TRANSIT, VISIT, UNKNOWN
	MinDuration  int64   `form:"minDuration"`  // Seconds
	Province     string  `form:"province"`
	City         string  `form:"city"`
	County       string  `form:"county"`
	StartTime    int64   `form:"startTime"`    // Unix timestamp
	EndTime      int64   `form:"endTime"`      // Unix timestamp
	MinConfidence float64 `form:"minConfidence"` // 0-1
	Page         int     `form:"page"`
	PageSize     int     `form:"pageSize"`
}

// TripFilter represents filter parameters for querying trips
type TripFilter struct {
	StartTime   int64   `form:"startTime"`   // Unix timestamp
	EndTime     int64   `form:"endTime"`     // Unix timestamp
	OriginCity  string  `form:"originCity"`
	DestCity    string  `form:"destCity"`
	MinDistance float64 `form:"minDistance"` // Meters
	PrimaryMode string  `form:"primaryMode"` // WALK, CAR, TRAIN, FLIGHT
	TripType    string  `form:"tripType"`    // COMMUTE, ROUND_TRIP, ONE_WAY, MULTI_STOP
	Page        int     `form:"page"`
	PageSize    int     `form:"pageSize"`
}

// GridFilter represents filter parameters for querying grid cells
type GridFilter struct {
	Level      int     `form:"level"`      // 1-5
	MinLat     float64 `form:"minLat"`
	MaxLat     float64 `form:"maxLat"`
	MinLon     float64 `form:"minLon"`
	MaxLon     float64 `form:"maxLon"`
	MinDensity int     `form:"minDensity"` // Minimum point count
}

// RenderFilter represents filter parameters for rendering metadata
type RenderFilter struct {
	MinLat    float64 `form:"minLat"`
	MaxLat    float64 `form:"maxLat"`
	MinLon    float64 `form:"minLon"`
	MaxLon    float64 `form:"maxLon"`
	LODLevel  int     `form:"lodLevel"`  // Level of detail 1-5
	StartTime int64   `form:"startTime"` // Unix timestamp
	EndTime   int64   `form:"endTime"`   // Unix timestamp
	Mode      string  `form:"mode"`      // Filter by transport mode
	Limit     int     `form:"limit"`     // Max points to return
}

// StatsFilter represents filter parameters for statistics queries
type StatsFilter struct {
	StatType  string `form:"statType"`  // PROVINCE, CITY, COUNTY, TOWN, GRID, ACTIVITY_TYPE
	TimeRange string `form:"timeRange"` // all, YYYY, YYYY-MM, YYYY-MM-DD
	OrderBy   string `form:"orderBy"`   // points, visits, duration, distance, count
	Limit     int    `form:"limit"`     // Max results
}
