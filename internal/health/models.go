package health

import "time"

// HealthRecord represents a health data record
type HealthRecord struct {
	ID             int       `json:"id"`
	Type           string    `json:"type"`
	Value          float64   `json:"value"`
	Unit           string    `json:"unit"`
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	SourceName     string    `json:"sourceName"`
	SourceVersion  string    `json:"sourceVersion,omitempty"`
	Device         string    `json:"device,omitempty"`
	CreationDate   time.Time `json:"creationDate,omitempty"`
	Metadata       string    `json:"metadata,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Workout represents a workout session
type Workout struct {
	ID              int       `json:"id"`
	WorkoutType     string    `json:"workoutType"`
	DurationSeconds float64   `json:"durationSeconds"`
	DurationUnit    string    `json:"durationUnit,omitempty"`
	DistanceMeters  float64   `json:"distanceMeters"`
	DistanceUnit    string    `json:"distanceUnit,omitempty"`
	Calories        float64   `json:"calories"`
	CaloriesUnit    string    `json:"caloriesUnit,omitempty"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	SourceName      string    `json:"sourceName"`
	SourceVersion   string    `json:"sourceVersion,omitempty"`
	Device          string    `json:"device,omitempty"`
	CreationDate    time.Time `json:"creationDate,omitempty"`
	RouteFile       string    `json:"routeFile,omitempty"`
	Metadata        string    `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// WorkoutRoute represents a GPS point in a workout route
type WorkoutRoute struct {
	ID                   int       `json:"id"`
	WorkoutID            int       `json:"workoutId"`
	Timestamp            time.Time `json:"timestamp"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	Altitude             float64   `json:"altitude,omitempty"`
	Speed                float64   `json:"speed,omitempty"`
	HorizontalAccuracy   float64   `json:"horizontalAccuracy,omitempty"`
	VerticalAccuracy     float64   `json:"verticalAccuracy,omitempty"`
	Course               float64   `json:"course,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
}

// SleepAnalysis represents sleep data
type SleepAnalysis struct {
	ID              int       `json:"id"`
	SleepType       string    `json:"sleepType"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	DurationSeconds int       `json:"durationSeconds"`
	SourceName      string    `json:"sourceName"`
	Device          string    `json:"device,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// ActivitySummary represents daily activity summary
type ActivitySummary struct {
	ID                      int       `json:"id"`
	Date                    string    `json:"date"`
	ActiveEnergyBurned      float64   `json:"activeEnergyBurned"`
	ActiveEnergyBurnedUnit  string    `json:"activeEnergyBurnedUnit,omitempty"`
	AppleExerciseTime       float64   `json:"appleExerciseTime"`
	AppleExerciseTimeUnit   string    `json:"appleExerciseTimeUnit,omitempty"`
	AppleStandHours         int       `json:"appleStandHours"`
	AppleStandHoursGoal     int       `json:"appleStandHoursGoal"`
	CreatedAt               time.Time `json:"createdAt"`
}

// HealthStatistics represents cached statistics
type HealthStatistics struct {
	ID         int       `json:"id"`
	StatType   string    `json:"statType"`   // daily, weekly, monthly, yearly
	StatDate   string    `json:"statDate"`   // YYYY-MM-DD, YYYY-Www, YYYY-MM, YYYY
	MetricType string    `json:"metricType"` // steps, heart_rate, sleep, etc.
	TotalValue float64   `json:"totalValue"`
	AvgValue   float64   `json:"avgValue"`
	MinValue   float64   `json:"minValue"`
	MaxValue   float64   `json:"maxValue"`
	Count      int       `json:"count"`
	Data       string    `json:"data,omitempty"` // JSON format
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// HealthSummary represents overall health data summary
type HealthSummary struct {
	TotalRecords        int       `json:"totalRecords"`
	TotalWorkouts       int       `json:"totalWorkouts"`
	TotalSleepRecords   int       `json:"totalSleepRecords"`
	DateRangeStart      string    `json:"dateRangeStart"`
	DateRangeEnd        string    `json:"dateRangeEnd"`
	LastImportDate      time.Time `json:"lastImportDate"`
	AvailableMetrics    []string  `json:"availableMetrics"`
}

// RecordFilter represents filters for querying health records
type RecordFilter struct {
	Type      string    `json:"type"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// WorkoutFilter represents filters for querying workouts
type WorkoutFilter struct {
	WorkoutType string    `json:"workoutType"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

// StatisticsFilter represents filters for statistics queries
type StatisticsFilter struct {
	StatType   string `json:"statType"`   // daily, weekly, monthly
	MetricType string `json:"metricType"` // steps, heart_rate, etc.
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
	Limit      int    `json:"limit"`
}
