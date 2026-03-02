package models

import (
	"time"
)

// Location represents a clustered location
type Location struct {
	ID               int       `json:"id"`
	Geohash          string    `json:"geohash"`
	CenterLat        float64   `json:"centerLat"`
	CenterLon        float64   `json:"centerLon"`
	Radius           float64   `json:"radius"`
	VisitCount       int       `json:"visitCount"`
	TotalDuration    int       `json:"totalDuration"`
	FirstVisit       string    `json:"firstVisit"`
	LastVisit        string    `json:"lastVisit"`
	Label            string    `json:"label"`
	LabelConfidence  float64   `json:"labelConfidence"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// LocationBehavior represents behavior data for a location visit
type LocationBehavior struct {
	ID                   int       `json:"id"`
	LocationID           int       `json:"locationId"`
	VisitDate            string    `json:"visitDate"`
	VisitStart           string    `json:"visitStart"`
	VisitEnd             string    `json:"visitEnd"`
	Duration             int       `json:"duration"`
	TypingSpeed          float64   `json:"typingSpeed"`
	WorkAppRatio         float64   `json:"workAppRatio"`
	EntertainmentRatio   float64   `json:"entertainmentRatio"`
	FocusDuration        int       `json:"focusDuration"`
	AppSwitchCount       int       `json:"appSwitchCount"`
	AvgHeartRate         float64   `json:"avgHeartRate"`
	HeartRateVariability float64   `json:"heartRateVariability"`
	Steps                int       `json:"steps"`
	CreatedAt            time.Time `json:"createdAt"`
}

// LocationEfficiencyScore represents efficiency scores for a location
type LocationEfficiencyScore struct {
	ID                int       `json:"id"`
	LocationID        int       `json:"locationId"`
	ProductivityScore float64   `json:"productivityScore"`
	HealthScore       float64   `json:"healthScore"`
	FocusScore        float64   `json:"focusScore"`
	EfficiencyScore   float64   `json:"efficiencyScore"`
	VisitCount        int       `json:"visitCount"`
	AvgDuration       int       `json:"avgDuration"`
	CalculatedAt      time.Time `json:"calculatedAt"`
}

// LocationHabit represents a detected habit for a location
type LocationHabit struct {
	ID                int       `json:"id"`
	LocationID        int       `json:"locationId"`
	HabitType         string    `json:"habitType"`
	HabitDescription  string    `json:"habitDescription"`
	Confidence        float64   `json:"confidence"`
	OccurrenceCount   int       `json:"occurrenceCount"`
	DetectedAt        time.Time `json:"detectedAt"`
}

// LocationWithEfficiency combines location with efficiency data
type LocationWithEfficiency struct {
	Location
	EfficiencyScore LocationEfficiencyScore `json:"efficiencyScore"`
	Habits          []LocationHabit         `json:"habits"`
}

// LocationBehaviorAnalysis represents the complete analysis result
type LocationBehaviorAnalysis struct {
	Locations       []LocationWithEfficiency `json:"locations"`
	TopEfficient    []LocationWithEfficiency `json:"topEfficient"`
	LeastEfficient  []LocationWithEfficiency `json:"leastEfficient"`
	Summary         LocationSummary          `json:"summary"`
}

// LocationSummary provides overall statistics
type LocationSummary struct {
	TotalLocations      int     `json:"totalLocations"`
	TotalVisits         int     `json:"totalVisits"`
	AvgEfficiencyScore  float64 `json:"avgEfficiencyScore"`
	MostEfficientLabel  string  `json:"mostEfficientLabel"`
	LeastEfficientLabel string  `json:"leastEfficientLabel"`
}
