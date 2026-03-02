package screentime

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Handler struct {
	db *sql.DB
}

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

// GetSummary returns overall usage summary
// GET /api/v1/screentime/summary
func (h *Handler) GetSummary(c *gin.Context) {
	query := `
	SELECT
		COUNT(DISTINCT package_id) as total_apps,
		SUM(duration_ms) as total_duration,
		COUNT(DISTINCT date) as active_days,
		SUM(launch_count) as total_launches,
		SUM(notification_count) as total_notifications,
		MIN(date) as start_date,
		MAX(date) as end_date
	FROM screentime_daily
	WHERE package_id != 'ALL'
	`

	var summary Summary
	var totalDuration sql.NullInt64
	var totalLaunches, totalNotifications sql.NullInt64

	err := h.db.QueryRow(query).Scan(
		&summary.TotalApps,
		&totalDuration,
		&summary.ActiveDays,
		&totalLaunches,
		&totalNotifications,
		&summary.DateRange.Start,
		&summary.DateRange.End,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	summary.TotalDurationMS = totalDuration.Int64
	summary.TotalLaunches = int(totalLaunches.Int64)
	summary.TotalNotifications = int(totalNotifications.Int64)

	if summary.ActiveDays > 0 {
		summary.AvgDailyDuration = float64(summary.TotalDurationMS) / float64(summary.ActiveDays)
	}

	// Get top app
	topAppQuery := `
	SELECT app_name, package_id, total_duration_ms
	FROM screentime_apps
	WHERE package_id != 'ALL'
	ORDER BY total_duration_ms DESC
	LIMIT 1
	`

	err = h.db.QueryRow(topAppQuery).Scan(
		&summary.TopApp,
		&summary.TopAppPackage,
		&summary.TopAppDurationMS,
	)

	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetDailyStats returns daily statistics
// GET /api/v1/screentime/daily?start=20240101&end=20241231&limit=30
func (h *Handler) GetDailyStats(c *gin.Context) {
	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")
	limitStr := c.DefaultQuery("limit", "30")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 30
	}

	query := `
	SELECT
		d.date,
		SUM(d.duration_ms) as total_duration,
		COUNT(DISTINCT d.package_id) as unique_apps,
		SUM(d.launch_count) as launch_count,
		SUM(d.notification_count) as notification_count
	FROM screentime_daily d
	WHERE d.package_id != 'ALL'
	`

	args := []interface{}{}

	if start != "" {
		query += " AND d.date >= ?"
		args = append(args, start)
	}

	if end != "" {
		query += " AND d.date <= ?"
		args = append(args, end)
	}

	query += `
	GROUP BY d.date
	ORDER BY d.date DESC
	LIMIT ?
	`
	args = append(args, limit)

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
			&stat.Date,
			&stat.TotalDurationMS,
			&stat.UniqueApps,
			&stat.LaunchCount,
			&stat.NotificationCount,
		)
		if err != nil {
			continue
		}

		// Get top app for this day
		topAppQuery := `
		SELECT app_name, duration_ms
		FROM screentime_daily
		WHERE date = ? AND package_id != 'ALL'
		ORDER BY duration_ms DESC
		LIMIT 1
		`
		h.db.QueryRow(topAppQuery, stat.Date).Scan(&stat.TopApp, &stat.TopAppDurationMS)

		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, stats)
}

// GetRankings returns app rankings
// GET /api/v1/screentime/rankings?limit=20&orderBy=duration&category=Social
func (h *Handler) GetRankings(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	orderBy := c.DefaultQuery("orderBy", "duration")
	category := c.Query("category")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	query := `
	SELECT
		app_name,
		package_id,
		category,
		total_duration_ms,
		total_launches,
		total_notifications,
		(julianday(last_seen) - julianday(first_seen) + 1) as active_days
	FROM screentime_apps
	WHERE package_id != 'ALL'
	`

	args := []interface{}{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	switch orderBy {
	case "launches":
		query += " ORDER BY total_launches DESC"
	case "notifications":
		query += " ORDER BY total_notifications DESC"
	default:
		query += " ORDER BY total_duration_ms DESC"
	}

	query += " LIMIT ?"
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get total duration for percentage calculation
	var totalDuration int64
	h.db.QueryRow("SELECT SUM(total_duration_ms) FROM screentime_apps WHERE package_id != 'ALL'").Scan(&totalDuration)

	var rankings []AppRanking
	rank := 1
	for rows.Next() {
		var r AppRanking
		var activeDays float64
		err := rows.Scan(
			&r.AppName,
			&r.PackageID,
			&r.Category,
			&r.TotalDurationMS,
			&r.LaunchCount,
			&r.NotificationCount,
			&activeDays,
		)
		if err != nil {
			continue
		}

		r.Rank = rank
		if totalDuration > 0 {
			r.Percentage = float64(r.TotalDurationMS) / float64(totalDuration) * 100
		}
		r.ActiveDays = int(activeDays)
		if r.ActiveDays > 0 {
			r.AvgDailyDuration = float64(r.TotalDurationMS) / float64(r.ActiveDays)
		}

		rankings = append(rankings, r)
		rank++
	}

	c.JSON(http.StatusOK, rankings)
}

// GetCategories returns category statistics
// GET /api/v1/screentime/categories
func (h *Handler) GetCategories(c *gin.Context) {
	query := `
	SELECT
		category,
		COUNT(*) as app_count,
		SUM(total_duration_ms) as total_duration,
		SUM(total_launches) as total_launches,
		SUM(total_notifications) as total_notifications
	FROM screentime_apps
	WHERE package_id != 'ALL'
	GROUP BY category
	ORDER BY total_duration DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get total duration for percentage
	var totalDuration int64
	h.db.QueryRow("SELECT SUM(total_duration_ms) FROM screentime_apps WHERE package_id != 'ALL'").Scan(&totalDuration)

	var categories []CategoryStat
	for rows.Next() {
		var cat CategoryStat
		err := rows.Scan(
			&cat.Category,
			&cat.AppCount,
			&cat.TotalDurationMS,
			&cat.LaunchCount,
			&cat.NotificationCount,
		)
		if err != nil {
			continue
		}

		if totalDuration > 0 {
			cat.Percentage = float64(cat.TotalDurationMS) / float64(totalDuration) * 100
		}

		// Get app names for this category
		appQuery := `
		SELECT app_name
		FROM screentime_apps
		WHERE category = ? AND package_id != 'ALL'
		ORDER BY total_duration_ms DESC
		LIMIT 10
		`
		appRows, err := h.db.Query(appQuery, cat.Category)
		if err == nil {
			for appRows.Next() {
				var appName string
				if appRows.Scan(&appName) == nil {
					cat.Apps = append(cat.Apps, appName)
				}
			}
			appRows.Close()
		}

		categories = append(categories, cat)
	}

	c.JSON(http.StatusOK, categories)
}

// GetHourlyStats returns hourly usage distribution
// GET /api/v1/screentime/hourly?start=20240101&end=20241231
func (h *Handler) GetHourlyStats(c *gin.Context) {
	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")

	query := `
	SELECT
		CAST(substr(start_time, 1, 2) AS INTEGER) as hour,
		COUNT(*) as session_count,
		COUNT(DISTINCT package_id) as unique_apps
	FROM screentime_sessions
	WHERE 1=1
	`

	args := []interface{}{}

	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}

	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	query += `
	GROUP BY hour
	ORDER BY hour
	`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	hourlyStats := make([]HourlyStat, 24)
	for i := 0; i < 24; i++ {
		hourlyStats[i].Hour = i
	}

	for rows.Next() {
		var hour int
		var sessionCount, uniqueApps int
		err := rows.Scan(&hour, &sessionCount, &uniqueApps)
		if err != nil || hour < 0 || hour >= 24 {
			continue
		}

		hourlyStats[hour].LaunchCount = sessionCount
		hourlyStats[hour].UniqueApps = uniqueApps
	}

	c.JSON(http.StatusOK, hourlyStats)
}

// GetTrends returns usage trends
// GET /api/v1/screentime/trends?granularity=daily&start=20240101&end=20241231
func (h *Handler) GetTrends(c *gin.Context) {
	granularity := c.DefaultQuery("granularity", "daily")
	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")

	var query string
	args := []interface{}{}

	switch granularity {
	case "weekly":
		query = `
		SELECT
			strftime('%Y%W', date) as period,
			SUM(duration_ms) as total_duration
		FROM screentime_daily
		WHERE package_id != 'ALL'
		`
	case "monthly":
		query = `
		SELECT
			substr(date, 1, 6) as period,
			SUM(duration_ms) as total_duration
		FROM screentime_daily
		WHERE package_id != 'ALL'
		`
	default: // daily
		query = `
		SELECT
			date as period,
			SUM(duration_ms) as total_duration
		FROM screentime_daily
		WHERE package_id != 'ALL'
		`
	}

	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}

	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	query += `
	GROUP BY period
	ORDER BY period
	`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var trends []TrendPoint
	for rows.Next() {
		var point TrendPoint
		var duration int64
		err := rows.Scan(&point.Date, &duration)
		if err != nil {
			continue
		}

		point.Value = float64(duration) / 3600000.0 // Convert to hours
		trends = append(trends, point)
	}

	c.JSON(http.StatusOK, trends)
}

// GetAppDetail returns detailed statistics for a specific app
// GET /api/v1/screentime/app/:packageId
func (h *Handler) GetAppDetail(c *gin.Context) {
	packageID := c.Param("packageId")

	// Get app metadata
	var app App
	query := `
	SELECT
		id, package_id, app_name, category, first_seen, last_seen,
		total_duration_ms, total_launches, total_notifications,
		created_at, updated_at
	FROM screentime_apps
	WHERE package_id = ?
	`

	err := h.db.QueryRow(query, packageID).Scan(
		&app.ID,
		&app.PackageID,
		&app.AppName,
		&app.Category,
		&app.FirstSeen,
		&app.LastSeen,
		&app.TotalDurationMS,
		&app.TotalLaunches,
		&app.TotalNotifications,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get daily usage trend (last 30 days)
	trendQuery := `
	SELECT date, duration_ms, launch_count
	FROM screentime_daily
	WHERE package_id = ?
	ORDER BY date DESC
	LIMIT 30
	`

	rows, err := h.db.Query(trendQuery, packageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dailyTrend []map[string]interface{}
	for rows.Next() {
		var date string
		var duration, launches int64
		if rows.Scan(&date, &duration, &launches) == nil {
			dailyTrend = append(dailyTrend, map[string]interface{}{
				"date":        date,
				"duration":    duration,
				"launches":    launches,
			})
		}
	}

	response := map[string]interface{}{
		"app":        app,
		"dailyTrend": dailyTrend,
	}

	c.JSON(http.StatusOK, response)
}

// GetAppSwitchingPatternHandler returns app switching pattern analysis
// GET /api/v1/screentime/analysis/switching-pattern
func (h *Handler) GetAppSwitchingPatternHandler(c *gin.Context) {
	pattern, err := h.GetAppSwitchingPattern()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pattern)
}

// GetLateNightUsageAnalysisHandler returns late night usage analysis
// GET /api/v1/screentime/analysis/late-night
func (h *Handler) GetLateNightUsageAnalysisHandler(c *gin.Context) {
	analysis, err := h.GetLateNightUsageAnalysis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetAppCorrelationAnalysisHandler returns app correlation analysis
// GET /api/v1/screentime/analysis/app-correlation
func (h *Handler) GetAppCorrelationAnalysisHandler(c *gin.Context) {
	analysis, err := h.GetAppCorrelationAnalysis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

