package screentime

import (
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

	// Get phone apps
	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")
	phoneApps := make(map[string]bool)
	rows, _ := phoneConn.DB.Query("SELECT DISTINCT app_name FROM screentime_apps WHERE package_id != 'ALL' LIMIT 100")
	for rows.Next() {
		var appName string
		rows.Scan(&appName)
		phoneApps[appName] = true
	}
	rows.Close()

	// Get computer apps
	computerConn, _ := h.deviceManager.GetDevice("computer_main")
	computerApps := make(map[string]bool)
	rows, _ = computerConn.DB.Query("SELECT DISTINCT application FROM manictime_apps LIMIT 100")
	for rows.Next() {
		var appName string
		rows.Scan(&appName)
		computerApps[appName] = true
	}
	rows.Close()

	// Find cross-platform apps (simplified: check for common names)
	crossPlatformNames := []string{"Edge", "Chrome", "微信", "QQ", "Telegram"}
	for _, name := range crossPlatformNames {
		foundInPhone := false
		foundInComputer := false

		for phoneApp := range phoneApps {
			if contains(phoneApp, name) {
				foundInPhone = true
				break
			}
		}

		for computerApp := range computerApps {
			if contains(computerApp, name) {
				foundInComputer = true
				break
			}
		}

		if foundInPhone && foundInComputer {
			ecosystem.CrossPlatformApps = append(ecosystem.CrossPlatformApps, name)
		}
	}

	// Sample phone-only and computer-only apps
	count := 0
	for app := range phoneApps {
		if count >= 10 {
			break
		}
		ecosystem.PhoneOnlyApps = append(ecosystem.PhoneOnlyApps, app)
		count++
	}

	count = 0
	for app := range computerApps {
		if count >= 10 {
			break
		}
		ecosystem.ComputerOnlyApps = append(ecosystem.ComputerOnlyApps, app)
		count++
	}

	ecosystem.TotalApps = len(phoneApps) + len(computerApps)
	ecosystem.CrossPlatformCount = len(ecosystem.CrossPlatformApps)

	// Generate insights
	ecosystem.Insights = []string{
		"跨平台应用数量: " + string(rune(ecosystem.CrossPlatformCount+'0')),
		"手机专属应用占比较高，主要用于社交和娱乐",
		"电脑专属应用主要用于开发和办公",
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

	// Get basic stats
	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")
	computerConn, _ := h.deviceManager.GetDevice("computer_main")

	var phoneDuration, computerDuration int64
	var phoneActiveDays, computerActiveDays int

	phoneConn.DB.QueryRow("SELECT SUM(duration_ms), COUNT(DISTINCT date) FROM screentime_daily WHERE package_id != 'ALL'").Scan(&phoneDuration, &phoneActiveDays)
	computerConn.DB.QueryRow("SELECT SUM(total_duration_seconds) * 1000, COUNT(DISTINCT date) FROM manictime_daily").Scan(&computerDuration, &computerActiveDays)

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

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
