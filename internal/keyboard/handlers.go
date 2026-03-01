package keyboard

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/keyboard/analysis"
	_ "modernc.org/sqlite"
)

type Handler struct {
	db *sql.DB
}

// NewHandler creates a new keyboard handler
func NewHandler(dbPath string) (*Handler, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Handler{db: db}, nil
}

// Close closes the database connection
func (h *Handler) Close() error {
	return h.db.Close()
}

		// RegisterRoutes registers all keyboard routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	keyboard := r.Group("/keyboard")
	{
		// Data query endpoints
		keyboard.GET("/daily", h.GetDailyStats)
		keyboard.GET("/scancodes", h.GetScancodeStats)
		keyboard.GET("/top-keys", h.GetTopKeys)

		// Statistics endpoints
		keyboard.GET("/statistics/summary", h.GetSummaryStats)
		keyboard.GET("/statistics/trends", h.GetTrends)
		keyboard.GET("/statistics/temporal", h.GetTemporalAnalysis)
		keyboard.GET("/statistics/categories", h.GetCategoryAnalysis)
		keyboard.GET("/statistics/typing_behavior", h.GetTypingBehavior)
		keyboard.GET("/statistics/productivity", h.GetProductivityMetrics)
		keyboard.GET("/statistics/hand_balance", h.GetHandBalance)
		keyboard.GET("/statistics/weekday_weekend", h.GetWeekdayWeekendComparison)
		keyboard.GET("/cross_module", h.GetCrossModuleAnalysis)

		// Visualization data endpoints
		keyboard.GET("/heatmap/keyboard", h.GetKeyboardHeatmap)
		keyboard.GET("/heatmap/detailed", h.GetDetailedKeyboardHeatmap)
		keyboard.GET("/heatmap/time", h.GetTimeHeatmap)
	}
}

// GetDailyStats returns daily statistics with optional date range filter
func (h *Handler) GetDailyStats(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")
	limit := c.DefaultQuery("limit", "100")

	query := `
		SELECT
			k.date,
			k.keystrokes,
			COALESCE(m.lbcount, 0) as left_clicks,
			COALESCE(m.rbcount, 0) as right_clicks,
			COALESCE(m.mbcount, 0) as middle_clicks,
			COALESCE(m.xbcount, 0) as extra_clicks,
			COALESCE(m.wheel, 0) as wheel_scrolls,
			COALESCE(m.hwheel, 0) as h_wheel_scrolls,
			COALESCE(m.move, 0.0) as mouse_distance_m
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

	query += " ORDER BY k.date DESC LIMIT ?"
	limitInt, _ := strconv.Atoi(limit)
	args = append(args, limitInt)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var stats []DailyStat
	for rows.Next() {
		var stat DailyStat
		err := rows.Scan(
			&stat.Date, &stat.Keystrokes, &stat.LeftClicks,
			&stat.RightClicks, &stat.MiddleClicks, &stat.ExtraClicks,
			&stat.WheelScrolls, &stat.HWheelScrolls, &stat.MouseDistanceM,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  stats,
		"count": len(stats),
	})
}

// GetScancodeStats returns scancode statistics for a specific date
func (h *Handler) GetScancodeStats(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date parameter is required"})
		return
	}

	query := `
		SELECT s.date, s.scan_code, s.count
		FROM scan_codes s
		WHERE s.date = ?
		ORDER BY s.count DESC
	`

	rows, err := h.db.Query(query, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type ScancodeStatWithName struct {
		Date     string `json:"date"`
		ScanCode int    `json:"scanCode"`
		Count    int64  `json:"count"`
		KeyName  string `json:"keyName"`
	}

	var stats []ScancodeStatWithName
	for rows.Next() {
		var stat ScancodeStatWithName
		err := rows.Scan(
			&stat.Date, &stat.ScanCode, &stat.Count,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Get key name from in-memory mapping
		stat.KeyName = GetKeyName(stat.ScanCode)
		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  stats,
		"count": len(stats),
	})
}

// GetTopKeys returns top N keys by usage
func (h *Handler) GetTopKeys(c *gin.Context) {
	limit := c.DefaultQuery("limit", "20")

	query := `
		SELECT s.scan_code, SUM(s.count) as total
		FROM scan_codes s
		GROUP BY s.scan_code
		ORDER BY total DESC
		LIMIT ?
	`

	limitInt, _ := strconv.Atoi(limit)
	rows, err := h.db.Query(query, limitInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get total keystrokes for percentage calculation
	var totalKeystrokes int64
	err = h.db.QueryRow("SELECT SUM(keystrokes) FROM keyboard_data").Scan(&totalKeystrokes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var topKeys []TopKey
	for rows.Next() {
		var scancode int
		var count int64
		err := rows.Scan(&scancode, &count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get key name from in-memory mapping
		topKeys = append(topKeys, TopKey{
			Scancode:   scancode,
			KeyName:    GetKeyName(scancode),
			Count:      count,
			Percentage: float64(count) / float64(totalKeystrokes) * 100,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  topKeys,
		"count": len(topKeys),
	})
}

// GetSummaryStats returns overall summary statistics
func (h *Handler) GetSummaryStats(c *gin.Context) {
	var summary SummaryStats

	// Get total statistics
	query := `
		SELECT
			SUM(k.keystrokes) as total_keystrokes,
			SUM(COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0)) as total_clicks,
			SUM(COALESCE(m.move, 0.0)) as total_distance,
			AVG(k.keystrokes) as avg_keystrokes,
			AVG(COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0)) as avg_clicks,
			AVG(COALESCE(m.move, 0.0)) as avg_distance,
			COUNT(*) as total_days
		FROM keyboard_data k
		LEFT JOIN mouse_data m ON k.date = m.date
		WHERE k.keystrokes > 0
	`

	var totalDays int
	err := h.db.QueryRow(query).Scan(
		&summary.TotalKeystrokes,
		&summary.TotalClicks,
		&summary.TotalMouseDistance,
		&summary.AvgKeystrokesPerDay,
		&summary.AvgClicksPerDay,
		&summary.AvgMouseDistancePerDay,
		&totalDays,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get active days count
	err = h.db.QueryRow("SELECT COUNT(*) FROM keyboard_data WHERE keystrokes > 100").Scan(&summary.ActiveDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get peak day
	var peakDay PeakDay
	err = h.db.QueryRow(`
		SELECT k.date, k.keystrokes, COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0) as total_clicks
		FROM keyboard_data k
		LEFT JOIN mouse_data m ON k.date = m.date
		ORDER BY k.keystrokes DESC
		LIMIT 1
	`).Scan(&peakDay.Date, &peakDay.Keystrokes, &peakDay.Clicks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	summary.PeakDay = &peakDay

	// Get date range
	var dateRange DateRange
	err = h.db.QueryRow("SELECT MIN(date), MAX(date) FROM keyboard_data").Scan(&dateRange.Start, &dateRange.End)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	summary.DataRange = &dateRange

	c.JSON(http.StatusOK, summary)
}

// GetTrends returns time-series trend data
func (h *Handler) GetTrends(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")
	granularity := c.DefaultQuery("granularity", "daily") // daily, weekly, monthly

	var query string
	args := []interface{}{}

	switch granularity {
	case "daily":
		query = `
			SELECT k.date, k.keystrokes,
			       COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0) as total_clicks,
			       COALESCE(m.move, 0.0) as mouse_distance
			FROM keyboard_data k
			LEFT JOIN mouse_data m ON k.date = m.date
			WHERE 1=1
		`
	case "weekly":
		query = `
			SELECT
				strftime('%Y-W%W', substr(k.date, 1, 4) || '-' || substr(k.date, 5, 2) || '-' || substr(k.date, 7, 2)) as week,
				SUM(k.keystrokes) as keystrokes,
				SUM(COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0)) as total_clicks,
				SUM(COALESCE(m.move, 0.0)) as distance
			FROM keyboard_data k
			LEFT JOIN mouse_data m ON k.date = m.date
			WHERE 1=1
		`
	case "monthly":
		query = `
			SELECT
				substr(k.date, 1, 6) as month,
				SUM(k.keystrokes) as keystrokes,
				SUM(COALESCE(m.lbcount, 0) + COALESCE(m.rbcount, 0) + COALESCE(m.mbcount, 0) + COALESCE(m.xbcount, 0)) as total_clicks,
				SUM(COALESCE(m.move, 0.0)) as distance
			FROM keyboard_data k
			LEFT JOIN mouse_data m ON k.date = m.date
			WHERE 1=1
		`
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid granularity"})
		return
	}

	if startDate != "" {
		query += " AND k.date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND k.date <= ?"
		args = append(args, endDate)
	}

	if granularity != "daily" {
		query += " GROUP BY " + map[string]string{
			"weekly":  "week",
			"monthly": "month",
		}[granularity]
	}

	query += " ORDER BY date"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var trends []TrendData
	for rows.Next() {
		var trend TrendData
		err := rows.Scan(&trend.Date, &trend.Keystrokes, &trend.Clicks, &trend.Distance)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		trends = append(trends, trend)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        trends,
		"count":       len(trends),
		"granularity": granularity,
	})
}

// GetKeyboardHeatmap returns keyboard heatmap data
func (h *Handler) GetKeyboardHeatmap(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	query := `
		SELECT s.scan_code, SUM(s.count) as total
		FROM scan_codes s
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND s.date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND s.date <= ?"
		args = append(args, endDate)
	}

	query += " GROUP BY s.scan_code ORDER BY total DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type KeyHeatmapData struct {
		Scancode    int    `json:"scancode"`
		KeyName     string `json:"keyName"`
		KeyCategory string `json:"keyCategory"`
		Count       int64  `json:"count"`
	}

	var heatmapData []KeyHeatmapData
	for rows.Next() {
		var scancode int
		var count int64
		err := rows.Scan(&scancode, &count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get key info from in-memory mapping
		heatmapData = append(heatmapData, KeyHeatmapData{
			Scancode:    scancode,
			KeyName:     GetKeyName(scancode),
			KeyCategory: GetKeyCategory(scancode),
			Count:       count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  heatmapData,
		"count": len(heatmapData),
	})
}

// GetTimeHeatmap returns time-based heatmap data (hour of day x day of week)
func (h *Handler) GetTimeHeatmap(c *gin.Context) {
	// This requires hourly_stats table which needs to be populated
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Time heatmap requires hourly data - not yet implemented",
		"data":    []interface{}{},
	})
}

// GetTemporalAnalysis returns temporal analysis (day-of-week, monthly patterns)
func (h *Handler) GetTemporalAnalysis(c *gin.Context) {
	analysisType := c.DefaultQuery("type", "daily") // daily, monthly, weekday_vs_weekend
	startDate := c.Query("start")
	endDate := c.Query("end")

	analyzer := analysis.NewTemporalAnalyzer(h.db)

	switch analysisType {
	case "daily":
		result, err := analyzer.AnalyzeDayOfWeek(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "monthly":
		result, err := analyzer.AnalyzeMonthlyPatterns(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "weekday_vs_weekend":
		result, err := analyzer.AnalyzeWeekdayVsWeekend(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis type"})
	}
}

// GetCategoryAnalysis returns key category distribution and top keys
func (h *Handler) GetCategoryAnalysis(c *gin.Context) {
	analysisType := c.DefaultQuery("type", "distribution") // distribution, top_keys, modifiers
	startDate := c.Query("start")
	endDate := c.Query("end")

	analyzer := analysis.NewCategoryAnalyzer(h.db)

	switch analysisType {
	case "distribution":
		result, err := analyzer.AnalyzeCategoryDistribution(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "top_keys":
		result, err := analyzer.AnalyzeAllTopKeys(5, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "modifiers":
		result, err := analyzer.AnalyzeModifierUsage(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis type"})
	}
}

// GetTypingBehavior returns typing behavior metrics
func (h *Handler) GetTypingBehavior(c *gin.Context) {
	analysisType := c.DefaultQuery("type", "metrics") // metrics, special_keys, letter_frequency
	startDate := c.Query("start")
	endDate := c.Query("end")

	analyzer := analysis.NewTypingBehaviorAnalyzer(h.db)

	switch analysisType {
	case "metrics":
		result, err := analyzer.AnalyzeTypingMetrics(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "special_keys":
		result, err := analyzer.AnalyzeSpecialKeyUsage(20, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "letter_frequency":
		result, err := analyzer.AnalyzeLetterFrequency(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis type"})
	}
}

// GetProductivityMetrics returns productivity metrics
func (h *Handler) GetProductivityMetrics(c *gin.Context) {
	analysisType := c.DefaultQuery("type", "activity") // activity, intensity, peak_days
	startDate := c.Query("start")
	endDate := c.Query("end")

	analyzer := analysis.NewProductivityAnalyzer(h.db)

	switch analysisType {
	case "activity":
		result, err := analyzer.AnalyzeActivityMetrics(100, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "intensity":
		result, err := analyzer.AnalyzeTypingIntensity(startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	case "peak_days":
		result, err := analyzer.AnalyzePeakDays(10, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis type"})
	}
}

// GetDetailedKeyboardHeatmap returns detailed keyboard heatmap with statistics
func (h *Handler) GetDetailedKeyboardHeatmap(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	query := `
		SELECT
			s.scan_code,
			SUM(s.count) as total_count,
			MAX(s.count) as peak_count,
			(SELECT date FROM scan_codes WHERE scan_code = s.scan_code ORDER BY count DESC LIMIT 1) as peak_date,
			AVG(s.count) as avg_count_per_day,
			COUNT(DISTINCT s.date) as day_count
		FROM scan_codes s
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND s.date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND s.date <= ?"
		args = append(args, endDate)
	}

	query += " GROUP BY s.scan_code ORDER BY total_count DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get total keystrokes for percentage
	var totalKeystrokes int64
	totalQuery := "SELECT SUM(keystrokes) FROM keyboard_data WHERE 1=1"
	totalArgs := []interface{}{}
	if startDate != "" {
		totalQuery += " AND date >= ?"
		totalArgs = append(totalArgs, startDate)
	}
	if endDate != "" {
		totalQuery += " AND date <= ?"
		totalArgs = append(totalArgs, endDate)
	}
	err = h.db.QueryRow(totalQuery, totalArgs...).Scan(&totalKeystrokes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type DetailedKeyHeatmap struct {
		Scancode       int     `json:"scancode"`
		KeyName        string  `json:"keyName"`
		KeyCategory    string  `json:"keyCategory"`
		TotalCount     int64   `json:"totalCount"`
		PeakCount      int64   `json:"peakCount"`
		PeakDate       string  `json:"peakDate"`
		AvgCountPerDay float64 `json:"avgCountPerDay"`
		Percentage     float64 `json:"percentage"`
		DayCount       int     `json:"dayCount"`
	}

	var heatmapData []DetailedKeyHeatmap
	for rows.Next() {
		var scancode int
		var totalCount, peakCount int64
		var peakDate string
		var avgCountPerDay float64
		var dayCount int

		err := rows.Scan(
			&scancode,
			&totalCount,
			&peakCount,
			&peakDate,
			&avgCountPerDay,
			&dayCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get key info from in-memory mapping and calculate percentage
		heatmapData = append(heatmapData, DetailedKeyHeatmap{
			Scancode:       scancode,
			KeyName:        GetKeyName(scancode),
			KeyCategory:    GetKeyCategory(scancode),
			TotalCount:     totalCount,
			PeakCount:      peakCount,
			PeakDate:       peakDate,
			AvgCountPerDay: avgCountPerDay,
			Percentage:     float64(totalCount) / float64(totalKeystrokes) * 100,
			DayCount:       dayCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  heatmapData,
		"count": len(heatmapData),
	})
}

// GetHandBalance returns hand usage balance statistics
// GET /api/v1/keyboard/statistics/hand_balance
func (h *Handler) GetHandBalance(c *gin.Context) {
	stats, err := analysis.GetHandBalanceStats(h.db, GetKeyHand, GetKeyName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetWeekdayWeekendComparison returns weekday vs weekend keyboard usage comparison
// GET /api/v1/keyboard/statistics/weekday_weekend
func (h *Handler) GetWeekdayWeekendComparison(c *gin.Context) {
	comparison, err := analysis.GetWeekdayWeekendComparison(h.db, GetKeyCategory, GetKeyHand, GetKeyName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// GetCrossModuleAnalysis returns cross-module analysis with screentime
// GET /api/v1/keyboard/cross_module
func (h *Handler) GetCrossModuleAnalysis(c *gin.Context) {
	// Get screentime database path from config
	screentimeDBPath := "data/screentime.db"

	screentimeDB, err := sql.Open("sqlite3", screentimeDBPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open screentime database"})
		return
	}
	defer screentimeDB.Close()

	analysis, err := GetCrossModuleAnalysis(h.db, screentimeDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}
