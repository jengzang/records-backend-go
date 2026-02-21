package models

import "time"

// FootprintStatistics represents aggregated footprint statistics
type FootprintStatistics struct {
	ID int64 `json:"id" db:"id"`

	// Time range for query-based statistics
	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`

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
	TotalPoints         int     `json:"total_points"`
	PointCount          int     `json:"point_count" db:"point_count"`
	VisitCount          int     `json:"visit_count" db:"visit_count"`
	TotalDistanceMeters float64 `json:"total_distance_meters" db:"total_distance_meters"`
	TotalDurationSeconds int64  `json:"total_duration_seconds" db:"total_duration_seconds"`
	FirstVisitTime      int64   `json:"first_visit_time,omitempty" db:"first_visit_time"` // Unix timestamp
	LastVisitTime       int64   `json:"last_visit_time,omitempty" db:"last_visit_time"`   // Unix timestamp

	// Administrative division counts and lists
	ProvinceCount int      `json:"province_count"`
	Provinces     []string `json:"provinces,omitempty"`
	CityCount     int      `json:"city_count"`
	Cities        []string `json:"cities,omitempty"`
	CountyCount   int      `json:"county_count"`
	Counties      []string `json:"counties,omitempty"`
	TownCount     int      `json:"town_count"`
	VillageCount  int      `json:"village_count"`

	// Rankings
	RankByPoints   int `json:"rank_by_points,omitempty" db:"rank_by_points"`
	RankByVisits   int `json:"rank_by_visits,omitempty" db:"rank_by_visits"`
	RankByDuration int `json:"rank_by_duration,omitempty" db:"rank_by_duration"`

	// Metadata
	AlgoVersion string    `json:"algo_version,omitempty" db:"algo_version"`
	GeneratedAt string    `json:"generated_at,omitempty"`
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

// TimeDistribution represents time-based distribution statistics
type TimeDistribution struct {
	Hour     int   `json:"hour" db:"hour"`
	Weekday  int   `json:"weekday" db:"weekday"`
	Count    int   `json:"count" db:"count"`
	Duration int64 `json:"duration" db:"duration"`
}

// SpeedDistribution represents speed-based distribution statistics
type SpeedDistribution struct {
	SpeedRange string  `json:"speed_range" db:"speed_range"` // e.g., "0-10", "10-20"
	Count      int     `json:"count" db:"count"`
	Percentage float64 `json:"percentage" db:"percentage"`
}

// AdminCrossing represents an administrative boundary crossing event
type AdminCrossing struct {
	ID                 int64     `json:"id" db:"id"`
	CrossingTS         int64     `json:"crossing_ts" db:"crossing_ts"`
	FromProvince       string    `json:"from_province,omitempty" db:"from_province"`
	FromCity           string    `json:"from_city,omitempty" db:"from_city"`
	FromCounty         string    `json:"from_county,omitempty" db:"from_county"`
	FromTown           string    `json:"from_town,omitempty" db:"from_town"`
	ToProvince         string    `json:"to_province,omitempty" db:"to_province"`
	ToCity             string    `json:"to_city,omitempty" db:"to_city"`
	ToCounty           string    `json:"to_county,omitempty" db:"to_county"`
	ToTown             string    `json:"to_town,omitempty" db:"to_town"`
	CrossingType       string    `json:"crossing_type" db:"crossing_type"` // PROVINCE/CITY/COUNTY/TOWN
	Latitude           float64   `json:"latitude" db:"latitude"`
	Longitude          float64   `json:"longitude" db:"longitude"`
	DistanceFromPrevM  float64   `json:"distance_from_prev_m" db:"distance_from_prev_m"`
	AlgoVersion        string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// AdminStats represents administrative region statistics
type AdminStats struct {
	ID              int64     `json:"id" db:"id"`
	AdminLevel      string    `json:"admin_level" db:"admin_level"` // PROVINCE/CITY/COUNTY/TOWN
	AdminName       string    `json:"admin_name" db:"admin_name"`
	ParentName      string    `json:"parent_name,omitempty" db:"parent_name"`
	VisitCount      int       `json:"visit_count" db:"visit_count"`
	TotalDurationS  int64     `json:"total_duration_s" db:"total_duration_s"`
	UniqueDays      int       `json:"unique_days" db:"unique_days"`
	FirstVisitTS    int64     `json:"first_visit_ts,omitempty" db:"first_visit_ts"`
	LastVisitTS     int64     `json:"last_visit_ts,omitempty" db:"last_visit_ts"`
	TotalDistanceM  float64   `json:"total_distance_m" db:"total_distance_m"`
	AlgoVersion     string    `json:"algo_version,omitempty" db:"algo_version"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// CrossingType constants
const (
	CrossingTypeProvince = "PROVINCE"
	CrossingTypeCity     = "CITY"
	CrossingTypeCounty   = "COUNTY"
	CrossingTypeTown     = "TOWN"
)

// AdminLevel constants
const (
	AdminLevelProvince = "PROVINCE"
	AdminLevelCity     = "CITY"
	AdminLevelCounty   = "COUNTY"
	AdminLevelTown     = "TOWN"
)

// SpeedSpaceStats represents speed-space coupling statistics
type SpeedSpaceStats struct {
	ID             int64   `json:"id" db:"id"`
	BucketType     string  `json:"bucket_type" db:"bucket_type"`       // year, month, all
	BucketKey      string  `json:"bucket_key" db:"bucket_key"`         // 2024, 2024-01, all
	AreaType       string  `json:"area_type" db:"area_type"`           // PROVINCE, CITY, COUNTY
	AreaKey        string  `json:"area_key" db:"area_key"`             // Area name
	AvgSpeed       float64 `json:"avg_speed" db:"avg_speed"`           // km/h
	SpeedVariance  float64 `json:"speed_variance" db:"speed_variance"` // km/h²
	SpeedEntropy   float64 `json:"speed_entropy" db:"speed_entropy"`   // Shannon entropy
	TotalDistance  float64 `json:"total_distance" db:"total_distance"` // meters
	SegmentCount   int     `json:"segment_count" db:"segment_count"`
	IsHighSpeedZone bool   `json:"is_high_speed_zone" db:"is_high_speed_zone"`
	IsSlowLifeZone  bool   `json:"is_slow_life_zone" db:"is_slow_life_zone"`
	StayIntensity   float64 `json:"stay_intensity" db:"stay_intensity"`
	AlgoVersion     int     `json:"algo_version" db:"algo_version"`
	CreatedAt       string  `json:"created_at" db:"created_at"`
}
