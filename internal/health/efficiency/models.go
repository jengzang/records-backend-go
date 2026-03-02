package efficiency

import "time"

// HourlyEfficiencyScore represents efficiency score for a specific hour
type HourlyEfficiencyScore struct {
	ID   int64  `json:"id"`
	Date string `json:"date"` // YYYY-MM-DD
	Hour int    `json:"hour"` // 0-23

	// Keyboard metrics
	TypingSpeed           *float64 `json:"typing_speed,omitempty"`
	TypingSpeedNormalized *float64 `json:"typing_speed_normalized,omitempty"`

	// ScreenTime metrics
	WorkAppRatio              *float64 `json:"work_app_ratio,omitempty"`
	EntertainmentAppRatio     *float64 `json:"entertainment_app_ratio,omitempty"`
	FocusSessionCount         *int     `json:"focus_session_count,omitempty"`
	AppSwitchFrequency        *float64 `json:"app_switch_frequency,omitempty"`
	WorkAppRatioNormalized    *float64 `json:"work_app_ratio_normalized,omitempty"`
	FocusNormalized           *float64 `json:"focus_normalized,omitempty"`

	// Health metrics
	AvgHeartRate          *float64 `json:"avg_heart_rate,omitempty"`
	HeartRateVariability  *float64 `json:"heart_rate_variability,omitempty"`
	StepCount             *int     `json:"step_count,omitempty"`
	HRVNormalized         *float64 `json:"hrv_normalized,omitempty"`
	ActivityNormalized    *float64 `json:"activity_normalized,omitempty"`

	// Composite score
	EfficiencyScore float64 `json:"efficiency_score"`

	// Data quality
	HasKeyboardData    bool    `json:"has_keyboard_data"`
	HasScreenTimeData  bool    `json:"has_screentime_data"`
	HasHealthData      bool    `json:"has_health_data"`
	DataCompleteness   float64 `json:"data_completeness"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EfficiencyCurveProfile represents aggregated efficiency pattern
type EfficiencyCurveProfile struct {
	ID          int64  `json:"id"`
	ProfileType string `json:"profile_type"` // 'workday' or 'weekend'

	// 24-hour curve
	HourlyCurve [24]float64 `json:"hourly_curve"`

	// Peak analysis
	PeakHour      int     `json:"peak_hour"`
	PeakScore     float64 `json:"peak_score"`
	PeakStartHour int     `json:"peak_start_hour"`
	PeakEndHour   int     `json:"peak_end_hour"`

	// Low analysis
	LowHour  int     `json:"low_hour"`
	LowScore float64 `json:"low_score"`

	// Chronotype
	Chronotype           string  `json:"chronotype"` // 'morning', 'evening', 'intermediate'
	ChronotypeConfidence float64 `json:"chronotype_confidence"`

	// Statistics
	AvgEfficiency float64 `json:"avg_efficiency"`
	StdEfficiency float64 `json:"std_efficiency"`
	TotalSamples  int     `json:"total_samples"`

	// Date range
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EfficiencyInsight represents an actionable insight
type EfficiencyInsight struct {
	ID           int64  `json:"id"`
	InsightType  string `json:"insight_type"`
	Priority     int    `json:"priority"` // 0=low, 1=medium, 2=high
	Title        string `json:"title"`
	Description  string `json:"description"`
	Recommendation string `json:"recommendation,omitempty"`
	Data         string `json:"data,omitempty"` // JSON
	Confidence   float64 `json:"confidence"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EfficiencyCurveRequest represents request for efficiency curve
type EfficiencyCurveRequest struct {
	StartDate string `json:"start_date" form:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date" form:"end_date"`     // YYYY-MM-DD
}

// EfficiencyCurveResponse represents response with hourly data
type EfficiencyCurveResponse struct {
	Scores []HourlyEfficiencyScore `json:"scores"`
	Stats  struct {
		TotalHours       int     `json:"total_hours"`
		AvgEfficiency    float64 `json:"avg_efficiency"`
		MaxEfficiency    float64 `json:"max_efficiency"`
		MinEfficiency    float64 `json:"min_efficiency"`
		DataCompleteness float64 `json:"data_completeness"`
	} `json:"stats"`
}

// ProfileComparisonResponse represents workday vs weekend comparison
type ProfileComparisonResponse struct {
	Workday EfficiencyCurveProfile `json:"workday"`
	Weekend EfficiencyCurveProfile `json:"weekend"`
	Diff    struct {
		AvgDiff       float64     `json:"avg_diff"`
		PeakHourDiff  int         `json:"peak_hour_diff"`
		HourlyDiff    [24]float64 `json:"hourly_diff"`
		Interpretation string     `json:"interpretation"`
	} `json:"diff"`
}
