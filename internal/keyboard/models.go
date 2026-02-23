package keyboard

import (
	"time"
)

// DailyStat represents daily keyboard/mouse usage statistics
type DailyStat struct {
	ID             int64     `json:"id"`
	Date           string    `json:"date"`
	Keystrokes     int       `json:"keystrokes"`
	LeftClicks     int       `json:"leftClicks"`
	RightClicks    int       `json:"rightClicks"`
	MiddleClicks   int       `json:"middleClicks"`
	ExtraClicks    int       `json:"extraClicks"`
	WheelScrolls   int       `json:"wheelScrolls"`
	HWheelScrolls  int       `json:"hWheelScrolls"`
	MouseDistanceM float64   `json:"mouseDistanceM"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ScancodeStat represents per-scancode daily statistics
type ScancodeStat struct {
	ID        int64     `json:"id"`
	Date      string    `json:"date"`
	Scancode  int       `json:"scancode"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"createdAt"`
}

// ScancodeMapping represents scancode to key name mapping
type ScancodeMapping struct {
	Scancode    int    `json:"scancode"`
	KeyName     string `json:"keyName"`
	KeyCategory string `json:"keyCategory"`
	Description string `json:"description"`
}

// SummaryStats represents overall summary statistics
type SummaryStats struct {
	TotalKeystrokes      int64   `json:"totalKeystrokes"`
	TotalClicks          int64   `json:"totalClicks"`
	TotalMouseDistance   float64 `json:"totalMouseDistance"`
	AvgKeystrokesPerDay  float64 `json:"avgKeystrokesPerDay"`
	AvgClicksPerDay      float64 `json:"avgClicksPerDay"`
	AvgMouseDistancePerDay float64 `json:"avgMouseDistancePerDay"`
	ActiveDays           int     `json:"activeDays"`
	PeakDay              *PeakDay `json:"peakDay"`
	DataRange            *DateRange `json:"dataRange"`
}

// PeakDay represents the day with highest usage
type PeakDay struct {
	Date       string `json:"date"`
	Keystrokes int    `json:"keystrokes"`
	Clicks     int    `json:"clicks"`
}

// DateRange represents the data date range
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TopKey represents a frequently used key
type TopKey struct {
	Scancode   int     `json:"scancode"`
	KeyName    string  `json:"keyName"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// TrendData represents time-series trend data
type TrendData struct {
	Date       string  `json:"date"`
	Keystrokes int     `json:"keystrokes"`
	Clicks     int     `json:"clicks"`
	Distance   float64 `json:"distance"`
}

// UsagePattern represents detected usage patterns
type UsagePattern struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
}
