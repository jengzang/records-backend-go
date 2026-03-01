package screentime

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDeviceSwitchingPatterns returns device switching patterns
// GET /api/v1/screentime/cross-device/switching-patterns
func (h *MultiDeviceHandler) GetDeviceSwitchingPatterns(c *gin.Context) {
	type SwitchingPattern struct {
		Date              string  `json:"date"`
		PhoneSessions     int     `json:"phoneSessions"`
		ComputerSessions  int     `json:"computerSessions"`
		EstimatedSwitches int     `json:"estimatedSwitches"`
		DominantDevice    string  `json:"dominantDevice"`
	}

	var patterns []SwitchingPattern

	// Get dates with both phone and computer usage
	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")
	computerConn, _ := h.deviceManager.GetDevice("computer_main")

	// Get recent 30 days
	query := `
	SELECT DISTINCT date
	FROM screentime_daily
	WHERE package_id != 'ALL'
	ORDER BY date DESC
	LIMIT 30
	`
	rows, _ := phoneConn.DB.Query(query)
	defer rows.Close()

	for rows.Next() {
		var date string
		rows.Scan(&date)

		var pattern SwitchingPattern
		pattern.Date = date

		// Count phone sessions (unlocks)
		phoneConn.DB.QueryRow("SELECT COUNT(*) FROM screentime_unlocks WHERE date = ?", date).Scan(&pattern.PhoneSessions)

		// Count computer sessions (approximate from activity changes)
		computerConn.DB.QueryRow("SELECT COUNT(*) FROM manictime_activities WHERE date = ?", date).Scan(&pattern.ComputerSessions)

		// Estimate switches (simplified: assume alternating usage)
		pattern.EstimatedSwitches = (pattern.PhoneSessions + pattern.ComputerSessions) / 2

		// Determine dominant device
		var phoneDuration, computerDuration int64
		phoneConn.DB.QueryRow("SELECT SUM(duration_ms) FROM screentime_daily WHERE date = ? AND package_id != 'ALL'", date).Scan(&phoneDuration)
		computerConn.DB.QueryRow("SELECT SUM(total_duration_seconds) * 1000 FROM manictime_daily WHERE date = ?", date).Scan(&computerDuration)

		if phoneDuration > computerDuration {
			pattern.DominantDevice = "phone"
		} else {
			pattern.DominantDevice = "computer"
		}

		patterns = append(patterns, pattern)
	}

	c.JSON(http.StatusOK, patterns)
}

// GetAppEcosystem returns application ecosystem analysis
// GET /api/v1/screentime/cross-device/app-ecosystem
func (h *MultiDeviceHandler) GetAppEcosystem(c *gin.Context) {
	type AppEcosystem struct {
		CrossPlatformApps []string `json:"crossPlatformApps"`
		PhoneOnlyApps     []string `json:"phoneOnlyApps"`
		ComputerOnlyApps  []string `json:"computerOnlyApps"`
		TotalApps         int      `json:"totalApps"`
		CrossPlatformCount int     `json:"crossPlatformCount"`
		Insights          []string `json:"insights"`
	}

	var ecosystem AppEcosystem
	ecosystem.CrossPlatformApps = []string{}
	ecosystem.PhoneOnlyApps = []string{}
	ecosystem.ComputerOnlyApps = []string{}
	ecosystem.Insights = []string{}

	// Get phone apps
	phoneConn, err := h.deviceManager.GetDevice("phone_vivo_x90")
	if err != nil || phoneConn == nil || phoneConn.DB == nil {
		c.JSON(http.StatusOK, ecosystem)
		return
	}

	phoneAppsList := []string{}
	rows, err := phoneConn.DB.Query("SELECT DISTINCT app_name FROM screentime_apps WHERE package_id != 'ALL' ORDER BY app_name LIMIT 500")
	if err == nil && rows != nil {
		defer rows.Close()
		for rows.Next() {
			var appName string
			if err := rows.Scan(&appName); err == nil && appName != "" {
				phoneAppsList = append(phoneAppsList, appName)
			}
		}
	}

	// Get computer apps
	computerConn, err := h.deviceManager.GetDevice("computer_main")
	if err != nil || computerConn == nil || computerConn.DB == nil {
		c.JSON(http.StatusOK, ecosystem)
		return
	}

	computerAppsList := []string{}
	rows2, err := computerConn.DB.Query("SELECT DISTINCT application FROM manictime_apps WHERE application != '' ORDER BY application LIMIT 500")
	if err == nil && rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var appName string
			if err := rows2.Scan(&appName); err == nil && appName != "" && appName != " " {
				computerAppsList = append(computerAppsList, appName)
			}
		}
	}

	// Use app normalizer to find cross-platform apps
	normalizer := NewAppNameNormalizer()
	ecosystem.CrossPlatformApps = normalizer.FindCrossPlatformApps(phoneAppsList, computerAppsList)
	ecosystem.PhoneOnlyApps = normalizer.FilterPhoneOnlyApps(phoneAppsList, computerAppsList, 20)
	ecosystem.ComputerOnlyApps = normalizer.FilterComputerOnlyApps(phoneAppsList, computerAppsList, 20)

	ecosystem.TotalApps = len(phoneAppsList) + len(computerAppsList)
	ecosystem.CrossPlatformCount = len(ecosystem.CrossPlatformApps)

	// Generate insights
	phoneOnlyCount := len(ecosystem.PhoneOnlyApps)
	computerOnlyCount := len(ecosystem.ComputerOnlyApps)
	crossPlatformCount := ecosystem.CrossPlatformCount

	ecosystem.Insights = append(ecosystem.Insights,
		fmt.Sprintf("跨平台应用数量: %d 个", crossPlatformCount),
		fmt.Sprintf("手机专属应用: %d 个，主要用于社交和娱乐", phoneOnlyCount),
		fmt.Sprintf("电脑专属应用: %d 个，主要用于开发和办公", computerOnlyCount),
	)

	if crossPlatformCount > 0 {
		ecosystem.Insights = append(ecosystem.Insights,
			fmt.Sprintf("跨平台应用占比: %.1f%%", float64(crossPlatformCount)/float64(ecosystem.TotalApps)*100),
		)
	}

	c.JSON(http.StatusOK, ecosystem)
}

// GetTimeAllocation returns time allocation analysis
// GET /api/v1/screentime/cross-device/time-allocation
func (h *MultiDeviceHandler) GetTimeAllocation(c *gin.Context) {
	type HourlyAllocation struct {
		Hour             int     `json:"hour"`
		PhoneDuration    int64   `json:"phoneDuration"`
		ComputerDuration int64   `json:"computerDuration"`
		TotalDuration    int64   `json:"totalDuration"`
		PhonePercentage  float64 `json:"phonePercentage"`
		ComputerPercentage float64 `json:"computerPercentage"`
	}

	allocations := make([]HourlyAllocation, 24)
	for i := 0; i < 24; i++ {
		allocations[i].Hour = i
	}

	// Get phone hourly data
	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")
	rows, _ := phoneConn.DB.Query(`
		SELECT
			CAST(substr(start_time, 1, 2) AS INTEGER) as hour,
			COUNT(*) * 60000 as estimated_duration
		FROM screentime_sessions
		GROUP BY hour
	`)
	for rows.Next() {
		var hour int
		var duration int64
		rows.Scan(&hour, &duration)
		if hour >= 0 && hour < 24 {
			allocations[hour].PhoneDuration = duration
		}
	}
	rows.Close()

	// Get computer hourly data (simplified: distribute evenly across active hours)
	computerConn, _ := h.deviceManager.GetDevice("computer_main")
	var totalComputerDuration int64
	computerConn.DB.QueryRow("SELECT SUM(total_duration_seconds) * 1000 FROM manictime_daily LIMIT 30").Scan(&totalComputerDuration)

	// Distribute computer usage across typical work hours (9-18)
	workHours := 9
	durationPerHour := totalComputerDuration / int64(workHours) / 30 // Average per day
	for i := 9; i < 18; i++ {
		allocations[i].ComputerDuration = durationPerHour
	}

	// Calculate totals and percentages
	for i := 0; i < 24; i++ {
		allocations[i].TotalDuration = allocations[i].PhoneDuration + allocations[i].ComputerDuration
		if allocations[i].TotalDuration > 0 {
			allocations[i].PhonePercentage = float64(allocations[i].PhoneDuration) / float64(allocations[i].TotalDuration) * 100
			allocations[i].ComputerPercentage = float64(allocations[i].ComputerDuration) / float64(allocations[i].TotalDuration) * 100
		}
	}

	c.JSON(http.StatusOK, allocations)
}

// GetUserProfile returns cross-device user profile
// GET /api/v1/screentime/cross-device/user-profile
func (h *MultiDeviceHandler) GetUserProfile(c *gin.Context) {
	type UserProfile struct {
		DeviceDependency  string   `json:"deviceDependency"`  // phone-dependent, computer-dependent, balanced
		WorkMode          string   `json:"workMode"`          // remote, office, hybrid
		EntertainmentPref string   `json:"entertainmentPref"` // phone-entertainment, computer-entertainment
		ProductivityType  string   `json:"productivityType"`  // high, medium, low
		HealthStatus      string   `json:"healthStatus"`      // healthy, warning, danger
		TotalScreentime   float64  `json:"totalScreentime"`   // hours per day
		Recommendations   []string `json:"recommendations"`
	}

	var profile UserProfile

	// Get basic stats using new query methods
	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")

	var phoneDuration int64
	var phoneActiveDays int
	phoneConn.DB.QueryRow("SELECT SUM(duration_ms), COUNT(DISTINCT date) FROM screentime_daily WHERE package_id != 'ALL'").Scan(&phoneDuration, &phoneActiveDays)

	// Use GetComputerSummary for computer data
	computerSummary, err := h.deviceManager.GetComputerSummary("computer_main")
	var computerDuration int64
	var computerActiveDays int
	if err == nil {
		if totalDuration, ok := computerSummary["totalDurationMS"].(int64); ok {
			computerDuration = totalDuration
		}
		if activeDays, ok := computerSummary["activeDays"].(int); ok {
			computerActiveDays = activeDays
		}
	}

	totalDuration := phoneDuration + computerDuration
	avgDays := phoneActiveDays
	if computerActiveDays > avgDays {
		avgDays = computerActiveDays
	}

	profile.TotalScreentime = float64(totalDuration) / float64(avgDays) / 3600000.0

	// Determine device dependency
	phonePercentage := float64(phoneDuration) / float64(totalDuration) * 100
	if phonePercentage > 60 {
		profile.DeviceDependency = "phone-dependent"
	} else if phonePercentage < 40 {
		profile.DeviceDependency = "computer-dependent"
	} else {
		profile.DeviceDependency = "balanced"
	}

	// Determine work mode (based on computer usage)
	if computerDuration > phoneDuration*2 {
		profile.WorkMode = "remote-work"
	} else {
		profile.WorkMode = "hybrid"
	}

	// Entertainment preference
	profile.EntertainmentPref = "phone-entertainment"

	// Productivity type
	if profile.TotalScreentime > 10 {
		profile.ProductivityType = "high"
	} else if profile.TotalScreentime > 6 {
		profile.ProductivityType = "medium"
	} else {
		profile.ProductivityType = "low"
	}

	// Health status
	if profile.TotalScreentime > 12 {
		profile.HealthStatus = "warning"
	} else if profile.TotalScreentime > 8 {
		profile.HealthStatus = "moderate"
	} else {
		profile.HealthStatus = "healthy"
	}

	// Generate recommendations
	profile.Recommendations = []string{}
	if profile.TotalScreentime > 10 {
		profile.Recommendations = append(profile.Recommendations, "建议减少总屏幕时间，每天控制在8小时以内")
	}
	if profile.DeviceDependency == "computer-dependent" {
		profile.Recommendations = append(profile.Recommendations, "电脑使用时间较长，建议每小时休息10分钟")
	}
	profile.Recommendations = append(profile.Recommendations, "保持工作生活平衡，适当增加户外活动")

	c.JSON(http.StatusOK, profile)
}

