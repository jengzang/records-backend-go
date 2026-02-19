package models

// FootprintStatistics represents footprint statistics at different administrative levels
type FootprintStatistics struct {
	// Time range
	StartTime int64 `json:"startTime" db:"start_time"`
	EndTime   int64 `json:"endTime" db:"end_time"`

	// Province level statistics
	ProvinceCount int      `json:"provinceCount" db:"province_count"`
	Provinces     []string `json:"provinces,omitempty"`

	// City level statistics
	CityCount int      `json:"cityCount" db:"city_count"`
	Cities    []string `json:"cities,omitempty"`

	// County level statistics
	CountyCount int      `json:"countyCount" db:"county_count"`
	Counties    []string `json:"counties,omitempty"`

	// Town level statistics
	TownCount int      `json:"townCount" db:"town_count"`
	Towns     []string `json:"towns,omitempty"`

	// Village level statistics
	VillageCount int      `json:"villageCount" db:"village_count"`
	Villages     []string `json:"villages,omitempty"`

	// Total points
	TotalPoints int64 `json:"totalPoints" db:"total_points"`

	// Metadata
	GeneratedAt string `json:"generatedAt" db:"generated_at"`
}

// StayStatistics represents stay statistics at different administrative levels
type StayStatistics struct {
	// Time range
	StartTime int64 `json:"startTime" db:"start_time"`
	EndTime   int64 `json:"endTime" db:"end_time"`

	// Administrative level
	AdminLevel string `json:"adminLevel" db:"admin_level"`  // "province", "city", "county", "town", "village"
	AdminName  string `json:"adminName" db:"admin_name"`

	// Stay statistics
	TotalStays      int   `json:"totalStays" db:"total_stays"`
	TotalDuration   int64 `json:"totalDuration" db:"total_duration"`      // Total duration in seconds
	AverageDuration int64 `json:"averageDuration" db:"average_duration"`  // Average duration in seconds
	LongestStay     int64 `json:"longestStay" db:"longest_stay"`          // Longest stay in seconds

	// Metadata
	GeneratedAt string `json:"generatedAt" db:"generated_at"`
}

// TimeDistribution represents time-based distribution statistics
type TimeDistribution struct {
	Hour  int   `json:"hour" db:"hour"`    // 0-23
	Count int64 `json:"count" db:"count"`
}

// SpeedDistribution represents speed-based distribution statistics
type SpeedDistribution struct {
	SpeedRange string `json:"speedRange" db:"speed_range"`  // e.g., "0-10", "10-30", "30-60", "60-120", "120+"
	Count      int64  `json:"count" db:"count"`
}

// AdminCrossing represents a crossing between administrative divisions
type AdminCrossing struct {
	FromProvince string `json:"fromProvince" db:"from_province"`
	FromCity     string `json:"fromCity" db:"from_city"`
	ToProvince   string `json:"toProvince" db:"to_province"`
	ToCity       string `json:"toCity" db:"to_city"`
	CrossingTime int64  `json:"crossingTime" db:"crossing_time"`  // Unix timestamp
	Count        int    `json:"count" db:"count"`                 // Number of times this crossing occurred
}
