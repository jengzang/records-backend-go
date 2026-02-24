package screentime

import "time"

// DailyUsage represents daily app usage statistics
type DailyUsage struct {
	ID                int       `json:"id" db:"id"`
	Date              string    `json:"date" db:"date"`                             // YYYYMMDD
	AppName           string    `json:"appName" db:"app_name"`
	PackageID         string    `json:"packageId" db:"package_id"`
	DurationMS        int64     `json:"durationMs" db:"duration_ms"`
	LaunchCount       int       `json:"launchCount" db:"launch_count"`
	NotificationCount int       `json:"notificationCount" db:"notification_count"`
	SplitScreenMS     int64     `json:"splitScreenMs" db:"split_screen_ms"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
}

// Session represents a single app usage session
type Session struct {
	ID            int       `json:"id" db:"id"`
	Date          string    `json:"date" db:"date"`
	StartTimeMS   int64     `json:"startTimeMs" db:"start_time_ms"`
	EndTimeMS     int64     `json:"endTimeMs" db:"end_time_ms"`
	StartTime     string    `json:"startTime" db:"start_time"`
	EndTime       string    `json:"endTime" db:"end_time"`
	AppName       string    `json:"appName" db:"app_name"`
	PackageID     string    `json:"packageId" db:"package_id"`
	DurationText  string    `json:"durationText" db:"duration_text"`
	IsStreaming   bool      `json:"isStreaming" db:"is_streaming"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// Unlock represents an unlock event with app sequence
type Unlock struct {
	ID              int       `json:"id" db:"id"`
	Date            string    `json:"date" db:"date"`
	UnlockTime      string    `json:"unlockTime" db:"unlock_time"`
	SessionDuration string    `json:"sessionDuration" db:"session_duration"`
	AppSequence     string    `json:"appSequence" db:"app_sequence"` // JSON array
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// App represents app metadata
type App struct {
	ID                   int       `json:"id" db:"id"`
	PackageID            string    `json:"packageId" db:"package_id"`
	AppName              string    `json:"appName" db:"app_name"`
	Category             string    `json:"category" db:"category"`
	FirstSeen            string    `json:"firstSeen" db:"first_seen"`
	LastSeen             string    `json:"lastSeen" db:"last_seen"`
	TotalDurationMS      int64     `json:"totalDurationMs" db:"total_duration_ms"`
	TotalLaunches        int       `json:"totalLaunches" db:"total_launches"`
	TotalNotifications   int       `json:"totalNotifications" db:"total_notifications"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at"`
}

// Statistics represents cached statistics
type Statistics struct {
	ID                 int       `json:"id" db:"id"`
	StatType           string    `json:"statType" db:"stat_type"`
	StatDate           string    `json:"statDate" db:"stat_date"`
	TotalDurationMS    int64     `json:"totalDurationMs" db:"total_duration_ms"`
	TotalUnlocks       int       `json:"totalUnlocks" db:"total_unlocks"`
	TotalLaunches      int       `json:"totalLaunches" db:"total_launches"`
	TotalNotifications int       `json:"totalNotifications" db:"total_notifications"`
	UniqueApps         int       `json:"uniqueApps" db:"unique_apps"`
	TopAppPackage      string    `json:"topAppPackage" db:"top_app_package"`
	TopAppDurationMS   int64     `json:"topAppDurationMs" db:"top_app_duration_ms"`
	Data               string    `json:"data" db:"data"` // JSON blob
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Summary represents overall usage summary
type Summary struct {
	TotalApps         int     `json:"totalApps"`
	TotalDurationMS   int64   `json:"totalDurationMs"`
	AvgDailyDuration  float64 `json:"avgDailyDuration"`
	TopApp            string  `json:"topApp"`
	TopAppPackage     string  `json:"topAppPackage"`
	TopAppDurationMS  int64   `json:"topAppDurationMs"`
	ActiveDays        int     `json:"activeDays"`
	TotalLaunches     int     `json:"totalLaunches"`
	TotalNotifications int    `json:"totalNotifications"`
	DateRange         struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"dateRange"`
}

// DailyStat represents daily aggregated statistics
type DailyStat struct {
	Date               string  `json:"date"`
	TotalDurationMS    int64   `json:"totalDurationMs"`
	UnlockCount        int     `json:"unlockCount"`
	LaunchCount        int     `json:"launchCount"`
	NotificationCount  int     `json:"notificationCount"`
	UniqueApps         int     `json:"uniqueApps"`
	TopApp             string  `json:"topApp"`
	TopAppDurationMS   int64   `json:"topAppDurationMs"`
}

// AppRanking represents app ranking with statistics
type AppRanking struct {
	Rank              int     `json:"rank"`
	AppName           string  `json:"appName"`
	PackageID         string  `json:"packageId"`
	Category          string  `json:"category"`
	TotalDurationMS   int64   `json:"totalDurationMs"`
	Percentage        float64 `json:"percentage"`
	LaunchCount       int     `json:"launchCount"`
	NotificationCount int     `json:"notificationCount"`
	ActiveDays        int     `json:"activeDays"`
	AvgDailyDuration  float64 `json:"avgDailyDuration"`
}

// CategoryStat represents category-level statistics
type CategoryStat struct {
	Category          string   `json:"category"`
	Apps              []string `json:"apps"`
	AppCount          int      `json:"appCount"`
	TotalDurationMS   int64    `json:"totalDurationMs"`
	Percentage        float64  `json:"percentage"`
	LaunchCount       int      `json:"launchCount"`
	NotificationCount int      `json:"notificationCount"`
}

// HourlyStat represents hourly usage distribution
type HourlyStat struct {
	Hour              int   `json:"hour"` // 0-23
	TotalDurationMS   int64 `json:"totalDurationMs"`
	UnlockCount       int   `json:"unlockCount"`
	LaunchCount       int   `json:"launchCount"`
	UniqueApps        int   `json:"uniqueApps"`
}

// TrendPoint represents a point in time series
type TrendPoint struct {
	Date            string  `json:"date"`
	Value           float64 `json:"value"`
	Label           string  `json:"label,omitempty"`
}
