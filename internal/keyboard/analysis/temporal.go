package analysis

import (
	"database/sql"
	"fmt"
	"time"
)

// TemporalAnalyzer handles temporal pattern analysis
type TemporalAnalyzer struct {
	db *sql.DB
}

// NewTemporalAnalyzer creates a new temporal analyzer
func NewTemporalAnalyzer(db *sql.DB) *TemporalAnalyzer {
	return &TemporalAnalyzer{db: db}
}

// DayOfWeekStat represents statistics for a day of the week
type DayOfWeekStat struct {
	Day              int     `json:"day"`
	DayName          string  `json:"dayName"`
	IsWeekend        bool    `json:"isWeekend"`
	TotalKeystrokes  int64   `json:"totalKeystrokes"`
	TotalClicks      int64   `json:"totalClicks"`
	TotalDistance    float64 `json:"totalDistance"`
	AvgKeystrokes    float64 `json:"avgKeystrokes"`
	AvgClicks        float64 `json:"avgClicks"`
	AvgDistance      float64 `json:"avgDistance"`
	DayCount         int     `json:"dayCount"`
}

// MonthlyPattern represents statistics for a month
type MonthlyPattern struct {
	Year             int     `json:"year"`
	Month            int     `json:"month"`
	MonthName        string  `json:"monthName"`
	TotalKeystrokes  int64   `json:"totalKeystrokes"`
	TotalClicks      int64   `json:"totalClicks"`
	TotalDistance    float64 `json:"totalDistance"`
	AvgKeystrokes    float64 `json:"avgKeystrokes"`
	AvgClicks        float64 `json:"avgClicks"`
	DayCount         int     `json:"dayCount"`
}

// WeekdayVsWeekend represents weekday vs weekend comparison
type WeekdayVsWeekend struct {
	Weekday struct {
		TotalKeystrokes int64   `json:"totalKeystrokes"`
		TotalClicks     int64   `json:"totalClicks"`
		TotalDistance   float64 `json:"totalDistance"`
		AvgKeystrokes   float64 `json:"avgKeystrokes"`
		AvgClicks       float64 `json:"avgClicks"`
		DayCount        int     `json:"dayCount"`
	} `json:"weekday"`
	Weekend struct {
		TotalKeystrokes int64   `json:"totalKeystrokes"`
		TotalClicks     int64   `json:"totalClicks"`
		TotalDistance   float64 `json:"totalDistance"`
		AvgKeystrokes   float64 `json:"avgKeystrokes"`
		AvgClicks       float64 `json:"avgClicks"`
		DayCount        int     `json:"dayCount"`
	} `json:"weekend"`
	WeekdayToWeekendRatio float64 `json:"weekdayToWeekendRatio"`
}

// AnalyzeDayOfWeek analyzes usage patterns by day of week
func (ta *TemporalAnalyzer) AnalyzeDayOfWeek(startDate, endDate string) ([]DayOfWeekStat, error) {
	query := `
		SELECT
			k.date,
			k.keystrokes,
			COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0) as total_clicks,
			COALESCE(m.move, 0.0) as mouse_distance
		FROM keyboard_data k
		LEFT JOIN mouse_data m ON k.date = m.date
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}

	rows, err := ta.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily stats: %w", err)
	}
	defer rows.Close()

	// Group by day of week
	dayStats := make(map[int]*DayOfWeekStat)
	for i := 0; i < 7; i++ {
		dayStats[i] = &DayOfWeekStat{
			Day:       i,
			IsWeekend: i >= 5,
		}
	}

	dayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

	for rows.Next() {
		var dateStr string
		var keystrokes, clicks int64
		var distance float64

		if err := rows.Scan(&dateStr, &keystrokes, &clicks, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse date (YYYYMMDD format)
		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}

		dayOfWeek := int(date.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 6 // Sunday -> 6
		} else {
			dayOfWeek-- // Monday=0, Tuesday=1, etc.
		}

		stat := dayStats[dayOfWeek]
		stat.TotalKeystrokes += keystrokes
		stat.TotalClicks += clicks
		stat.TotalDistance += distance
		stat.DayCount++
	}

	// Calculate averages
	result := make([]DayOfWeekStat, 0, 7)
	for i := 0; i < 7; i++ {
		stat := dayStats[i]
		stat.DayName = dayNames[i]
		if stat.DayCount > 0 {
			stat.AvgKeystrokes = float64(stat.TotalKeystrokes) / float64(stat.DayCount)
			stat.AvgClicks = float64(stat.TotalClicks) / float64(stat.DayCount)
			stat.AvgDistance = stat.TotalDistance / float64(stat.DayCount)
		}
		result = append(result, *stat)
	}

	return result, nil
}

// AnalyzeMonthlyPatterns analyzes usage patterns by month
func (ta *TemporalAnalyzer) AnalyzeMonthlyPatterns(startDate, endDate string) ([]MonthlyPattern, error) {
	query := `
		SELECT
			SUBSTR(k.date, 1, 4) as year,
			SUBSTR(k.date, 5, 2) as month,
			SUM(k.keystrokes) as total_keystrokes,
			SUM(COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0)) as total_clicks,
			SUM(COALESCE(m.move, 0.0)) as total_distance,
			AVG(k.keystrokes) as avg_keystrokes,
			AVG(COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0)) as avg_clicks,
			COUNT(*) as day_count
		FROM keyboard_data k
		LEFT JOIN mouse_data m ON k.date = m.date
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}

	query += " GROUP BY year, month ORDER BY year, month"

	rows, err := ta.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly patterns: %w", err)
	}
	defer rows.Close()

	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	var result []MonthlyPattern
	for rows.Next() {
		var pattern MonthlyPattern
		var yearStr, monthStr string

		if err := rows.Scan(&yearStr, &monthStr, &pattern.TotalKeystrokes,
			&pattern.TotalClicks, &pattern.TotalDistance, &pattern.AvgKeystrokes,
			&pattern.AvgClicks, &pattern.DayCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Sscanf(yearStr, "%d", &pattern.Year)
		fmt.Sscanf(monthStr, "%d", &pattern.Month)
		pattern.MonthName = monthNames[pattern.Month]

		result = append(result, pattern)
	}

	return result, nil
}

// AnalyzeWeekdayVsWeekend compares weekday vs weekend usage
func (ta *TemporalAnalyzer) AnalyzeWeekdayVsWeekend(startDate, endDate string) (*WeekdayVsWeekend, error) {
	query := `
		SELECT
			k.date,
			k.keystrokes,
			COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0) as total_clicks,
			COALESCE(m.move, 0.0) as mouse_distance
		FROM keyboard_data k
		LEFT JOIN mouse_data m ON k.date = m.date
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND date <= ?"
		args = append(args, endDate)
	}

	rows, err := ta.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily stats: %w", err)
	}
	defer rows.Close()

	result := &WeekdayVsWeekend{}

	for rows.Next() {
		var dateStr string
		var keystrokes, clicks int64
		var distance float64

		if err := rows.Scan(&dateStr, &keystrokes, &clicks, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}

		dayOfWeek := date.Weekday()
		if dayOfWeek == 0 || dayOfWeek == 6 { // Sunday or Saturday
			result.Weekend.TotalKeystrokes += keystrokes
			result.Weekend.TotalClicks += clicks
			result.Weekend.TotalDistance += distance
			result.Weekend.DayCount++
		} else {
			result.Weekday.TotalKeystrokes += keystrokes
			result.Weekday.TotalClicks += clicks
			result.Weekday.TotalDistance += distance
			result.Weekday.DayCount++
		}
	}

	// Calculate averages
	if result.Weekday.DayCount > 0 {
		result.Weekday.AvgKeystrokes = float64(result.Weekday.TotalKeystrokes) / float64(result.Weekday.DayCount)
		result.Weekday.AvgClicks = float64(result.Weekday.TotalClicks) / float64(result.Weekday.DayCount)
	}

	if result.Weekend.DayCount > 0 {
		result.Weekend.AvgKeystrokes = float64(result.Weekend.TotalKeystrokes) / float64(result.Weekend.DayCount)
		result.Weekend.AvgClicks = float64(result.Weekend.TotalClicks) / float64(result.Weekend.DayCount)
	}

	// Calculate ratio
	if result.Weekend.TotalKeystrokes > 0 {
		result.WeekdayToWeekendRatio = float64(result.Weekday.TotalKeystrokes) / float64(result.Weekend.TotalKeystrokes)
	}

	return result, nil
}
