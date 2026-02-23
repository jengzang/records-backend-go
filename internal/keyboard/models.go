package keyboard

// KeyboardData represents raw keyboard data from keyboard_data table
type KeyboardData struct {
	Date       string `json:"date"`
	Keystrokes int64  `json:"keystrokes"`
}

// MouseData represents raw mouse data from mouse_data table
type MouseData struct {
	Date          string  `json:"date"`
	LeftClicks    int64   `json:"leftClicks"`    // lbcount
	RightClicks   int64   `json:"rightClicks"`   // rbcount
	MiddleClicks  int64   `json:"middleClicks"`  // mbcount
	ExtraClicks   int64   `json:"extraClicks"`   // xbcount
	WheelScrolls  int64   `json:"wheelScrolls"`  // wheel
	HWheelScrolls int64   `json:"hWheelScrolls"` // hwheel
	MouseDistance float64 `json:"mouseDistance"` // move
}

// DailyStat represents combined daily keyboard/mouse usage statistics (from JOIN)
type DailyStat struct {
	Date           string  `json:"date"`
	Keystrokes     int64   `json:"keystrokes"`
	LeftClicks     int64   `json:"leftClicks"`
	RightClicks    int64   `json:"rightClicks"`
	MiddleClicks   int64   `json:"middleClicks"`
	ExtraClicks    int64   `json:"extraClicks"`
	WheelScrolls   int64   `json:"wheelScrolls"`
	HWheelScrolls  int64   `json:"hWheelScrolls"`
	MouseDistanceM float64 `json:"mouseDistanceM"`
	TotalClicks    int64   `json:"totalClicks"` // Computed field
}

// ScancodeStat represents per-scancode daily statistics from scan_codes table
type ScancodeStat struct {
	Date     string `json:"date"`
	ScanCode int    `json:"scanCode"` // Column name is scan_code in database
	Count    int64  `json:"count"`
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
