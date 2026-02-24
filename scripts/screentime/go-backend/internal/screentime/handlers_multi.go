package screentime

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

// MultiDeviceHandler handles requests across multiple devices
type MultiDeviceHandler struct {
	deviceManager *DeviceManager
}

// NewMultiDeviceHandler creates a new multi-device handler
func NewMultiDeviceHandler(deviceManager *DeviceManager) *MultiDeviceHandler {
	return &MultiDeviceHandler{
		deviceManager: deviceManager,
	}
}

// getDeviceDB returns the database connection for the specified device
// If device is "all", returns nil (caller should aggregate across all devices)
func (h *MultiDeviceHandler) getDeviceDB(c *gin.Context) (*sql.DB, string, error) {
	deviceID := c.DefaultQuery("device", "phone_vivo_x90") // Default to phone

	if deviceID == "all" {
		return nil, "all", nil
	}

	conn, err := h.deviceManager.GetDevice(deviceID)
	if err != nil {
		return nil, "", fmt.Errorf("device not found: %s", deviceID)
	}

	return conn.DB, deviceID, nil
}

// GetSummary returns overall usage summary
// GET /api/v1/screentime/summary?device=phone_vivo_x90|computer_main|all
func (h *MultiDeviceHandler) GetSummary(c *gin.Context) {
	db, deviceID, err := h.getDeviceDB(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if deviceID == "all" {
		// Aggregate across all devices
		h.getAggregatedSummary(c)
		return
	}

	// Single device summary
	h.getSingleDeviceSummary(c, db, deviceID)
}

func (h *MultiDeviceHandler) getSingleDeviceSummary(c *gin.Context, db *sql.DB, deviceID string) {
	conn, _ := h.deviceManager.GetDevice(deviceID)

	var query string
	if conn.Type == "phone" {
		query = `
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
	} else { // computer
		query = `
		SELECT
			COUNT(DISTINCT application) as total_apps,
			SUM(duration_seconds) * 1000 as total_duration,
			COUNT(DISTINCT date) as active_days,
			0 as total_launches,
			0 as total_notifications,
			MIN(date) as start_date,
			MAX(date) as end_date
		FROM manictime_daily
		`
	}

	var summary Summary
	var totalDuration sql.NullInt64
	var totalLaunches, totalNotifications sql.NullInt64

	err := db.QueryRow(query).Scan(
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
	if conn.Type == "phone" {
		topAppQuery := `
		SELECT app_name, package_id, total_duration_ms
		FROM screentime_apps
		WHERE package_id != 'ALL'
		ORDER BY total_duration_ms DESC
		LIMIT 1
		`
		db.QueryRow(topAppQuery).Scan(
			&summary.TopApp,
			&summary.TopAppPackage,
			&summary.TopAppDurationMS,
		)
	} else {
		topAppQuery := `
		SELECT application, application, total_duration_seconds * 1000
		FROM manictime_apps
		ORDER BY total_duration_seconds DESC
		LIMIT 1
		`
		db.QueryRow(topAppQuery).Scan(
			&summary.TopApp,
			&summary.TopAppPackage,
			&summary.TopAppDurationMS,
		)
	}

	c.JSON(http.StatusOK, summary)
}

func (h *MultiDeviceHandler) getAggregatedSummary(c *gin.Context) {
	connections := h.deviceManager.GetAllActiveConnections()

	var totalApps int
	var totalDuration int64
	var totalLaunches int
	var totalNotifications int
	var activeDays int
	var earliestDate, latestDate string

	for _, conn := range connections {
		var query string
		if conn.Type == "phone" {
			query = `
			SELECT
				COUNT(DISTINCT package_id),
				SUM(duration_ms),
				COUNT(DISTINCT date),
				SUM(launch_count),
				SUM(notification_count),
				MIN(date),
				MAX(date)
			FROM screentime_daily
			WHERE package_id != 'ALL'
			`
		} else {
			query = `
			SELECT
				COUNT(DISTINCT application),
				SUM(duration_seconds) * 1000,
				COUNT(DISTINCT date),
				0,
				0,
				MIN(date),
				MAX(date)
			FROM manictime_daily
			`
		}

		var apps int
		var duration, launches, notifications sql.NullInt64
		var days int
		var startDate, endDate string

		err := conn.DB.QueryRow(query).Scan(&apps, &duration, &days, &launches, &notifications, &startDate, &endDate)
		if err != nil {
			continue
		}

		totalApps += apps
		totalDuration += duration.Int64
		totalLaunches += int(launches.Int64)
		totalNotifications += int(notifications.Int64)
		if days > activeDays {
			activeDays = days
		}

		if earliestDate == "" || startDate < earliestDate {
			earliestDate = startDate
		}
		if latestDate == "" || endDate > latestDate {
			latestDate = endDate
		}
	}

	summary := Summary{
		TotalApps:          totalApps,
		TotalDurationMS:    totalDuration,
		TotalLaunches:      totalLaunches,
		TotalNotifications: totalNotifications,
		ActiveDays:         activeDays,
	}

	if activeDays > 0 {
		summary.AvgDailyDuration = float64(totalDuration) / float64(activeDays)
	}

	summary.DateRange.Start = earliestDate
	summary.DateRange.End = latestDate

	c.JSON(http.StatusOK, summary)
}

// GetDailyStats returns daily statistics
// GET /api/v1/screentime/daily?device=phone_vivo_x90&start=20240101&end=20241231&limit=30
func (h *MultiDeviceHandler) GetDailyStats(c *gin.Context) {
	db, deviceID, err := h.getDeviceDB(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if deviceID == "all" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device=all not supported for daily stats, please specify a device"})
		return
	}

	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")
	limitStr := c.DefaultQuery("limit", "30")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 30
	}

	conn, _ := h.deviceManager.GetDevice(deviceID)

	var query string
	if conn.Type == "phone" {
		query = `
		SELECT
			d.date,
			SUM(d.duration_ms) as total_duration,
			COUNT(DISTINCT d.package_id) as unique_apps,
			SUM(d.launch_count) as launch_count,
			SUM(d.notification_count) as notification_count
		FROM screentime_daily d
		WHERE d.package_id != 'ALL'
		`
	} else {
		query = `
		SELECT
			date,
			SUM(duration_seconds) * 1000 as total_duration,
			COUNT(DISTINCT application) as unique_apps,
			0 as launch_count,
			0 as notification_count
		FROM manictime_daily
		WHERE 1=1
		`
	}

	args := []interface{}{}

	if start != "" {
		query += " AND date >= ?"
		args = append(args, start)
	}

	if end != "" {
		query += " AND date <= ?"
		args = append(args, end)
	}

	if conn.Type == "phone" {
		query += `
		GROUP BY d.date
		ORDER BY d.date DESC
		LIMIT ?
		`
	} else {
		query += `
		GROUP BY date
		ORDER BY date DESC
		LIMIT ?
		`
	}
	args = append(args, limit)

	rows, err := db.Query(query, args...)
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

		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, stats)
}

// GetRankings returns app rankings
// GET /api/v1/screentime/rankings?device=phone_vivo_x90&limit=20&orderBy=duration
func (h *MultiDeviceHandler) GetRankings(c *gin.Context) {
	db, deviceID, err := h.getDeviceDB(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if deviceID == "all" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device=all not supported for rankings, please specify a device"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	orderBy := c.DefaultQuery("orderBy", "duration")
	category := c.Query("category")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	conn, _ := h.deviceManager.GetDevice(deviceID)

	var query string
	if conn.Type == "phone" {
		query = `
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
	} else {
		query = `
		SELECT
			application,
			application,
			category,
			total_duration_seconds * 1000,
			0,
			0,
			(julianday(last_seen) - julianday(first_seen) + 1) as active_days
		FROM manictime_apps
		WHERE 1=1
		`
	}

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
		if conn.Type == "phone" {
			query += " ORDER BY total_duration_ms DESC"
		} else {
			query += " ORDER BY total_duration_seconds DESC"
		}
	}

	query += " LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get total duration for percentage calculation
	var totalDuration int64
	if conn.Type == "phone" {
		db.QueryRow("SELECT SUM(total_duration_ms) FROM screentime_apps WHERE package_id != 'ALL'").Scan(&totalDuration)
	} else {
		var totalSeconds int64
		db.QueryRow("SELECT SUM(total_duration_seconds) FROM manictime_apps").Scan(&totalSeconds)
		totalDuration = totalSeconds * 1000
	}

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

// ListDevices returns all registered devices
// GET /api/v1/screentime/devices
func (h *MultiDeviceHandler) ListDevices(c *gin.Context) {
	devices, err := h.deviceManager.ListDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}
