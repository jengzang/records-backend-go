package screentime

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// CrossDeviceComparison represents device usage comparison
type CrossDeviceComparison struct {
	Phone struct {
		TotalDuration   int64   `json:"totalDuration"`
		AvgDailyDuration float64 `json:"avgDailyDuration"`
		TotalApps       int     `json:"totalApps"`
		ActiveDays      int     `json:"activeDays"`
		TopApp          string  `json:"topApp"`
	} `json:"phone"`
	Computer struct {
		TotalDuration   int64   `json:"totalDuration"`
		AvgDailyDuration float64 `json:"avgDailyDuration"`
		TotalApps       int     `json:"totalApps"`
		ActiveDays      int     `json:"activeDays"`
		TopApp          string  `json:"topApp"`
	} `json:"computer"`
	Total struct {
		TotalDuration    int64   `json:"totalDuration"`
		AvgDailyDuration float64 `json:"avgDailyDuration"`
		PhonePercentage  float64 `json:"phonePercentage"`
		ComputerPercentage float64 `json:"computerPercentage"`
	} `json:"total"`
	Insights []string `json:"insights"`
}

// WorkLifeBalance represents work-life balance analysis
type WorkLifeBalance struct {
	WorkDuration      int64   `json:"workDuration"`
	LifeDuration      int64   `json:"lifeDuration"`
	BalanceScore      int     `json:"balanceScore"`
	WorkPercentage    float64 `json:"workPercentage"`
	LifePercentage    float64 `json:"lifePercentage"`
	Recommendation    string  `json:"recommendation"`
	WeekdayPattern    struct {
		WorkDuration int64 `json:"workDuration"`
		LifeDuration int64 `json:"lifeDuration"`
	} `json:"weekdayPattern"`
	WeekendPattern    struct {
		WorkDuration int64 `json:"workDuration"`
		LifeDuration int64 `json:"lifeDuration"`
	} `json:"weekendPattern"`
	Insights []string `json:"insights"`
}

// GetCrossDeviceComparison returns device usage comparison
// GET /api/v1/screentime/cross-device/comparison
func (h *MultiDeviceHandler) GetCrossDeviceComparison(c *gin.Context) {
	logger.Info("Cross-device comparison requested", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	var comparison CrossDeviceComparison

	// Get phone data
	phoneConn, err := h.deviceManager.GetDevice("phone_vivo_x90")
	if err == nil {
		query := `
		SELECT
			SUM(duration_ms) as total_duration,
			COUNT(DISTINCT date) as active_days,
			COUNT(DISTINCT package_id) as total_apps
		FROM screentime_daily
		WHERE package_id != 'ALL'
		`
		var totalDuration, activeDays, totalApps sql.NullInt64
		if err := phoneConn.DB.QueryRow(query).Scan(&totalDuration, &activeDays, &totalApps); err != nil {
			logger.Error("Failed to query phone data", err, logrus.Fields{
				"device": "phone_vivo_x90",
			})
		} else {
			comparison.Phone.TotalDuration = totalDuration.Int64
			comparison.Phone.ActiveDays = int(activeDays.Int64)
			comparison.Phone.TotalApps = int(totalApps.Int64)
			if activeDays.Int64 > 0 {
				comparison.Phone.AvgDailyDuration = float64(totalDuration.Int64) / float64(activeDays.Int64)
			}

			// Get top app
			phoneConn.DB.QueryRow("SELECT app_name FROM screentime_apps WHERE package_id != 'ALL' ORDER BY total_duration_ms DESC LIMIT 1").Scan(&comparison.Phone.TopApp)
		}
	} else {
		logger.Warn("Phone device not available", logrus.Fields{
			"error": err.Error(),
		})
	}

	// Get computer data using new query method
	computerSummary, err := h.deviceManager.GetComputerSummary("computer_main")
	if err == nil {
		comparison.Computer.TotalDuration = computerSummary["totalDurationMS"].(int64)
		comparison.Computer.TotalApps = computerSummary["totalApps"].(int)
		comparison.Computer.ActiveDays = computerSummary["activeDays"].(int)
		if avgDaily, ok := computerSummary["avgDailyDuration"].(float64); ok {
			comparison.Computer.AvgDailyDuration = avgDaily
		}
		if topApp, ok := computerSummary["topApp"].(string); ok {
			comparison.Computer.TopApp = topApp
		}
	} else {
		logger.Warn("Computer device not available", logrus.Fields{
			"error": err.Error(),
		})
	}

	// Calculate totals
	comparison.Total.TotalDuration = comparison.Phone.TotalDuration + comparison.Computer.TotalDuration
	if comparison.Phone.ActiveDays > comparison.Computer.ActiveDays {
		comparison.Total.AvgDailyDuration = float64(comparison.Total.TotalDuration) / float64(comparison.Phone.ActiveDays)
	} else {
		comparison.Total.AvgDailyDuration = float64(comparison.Total.TotalDuration) / float64(comparison.Computer.ActiveDays)
	}

	if comparison.Total.TotalDuration > 0 {
		comparison.Total.PhonePercentage = float64(comparison.Phone.TotalDuration) / float64(comparison.Total.TotalDuration) * 100
		comparison.Total.ComputerPercentage = float64(comparison.Computer.TotalDuration) / float64(comparison.Total.TotalDuration) * 100
	}

	// Generate insights
	comparison.Insights = []string{}
	if comparison.Total.ComputerPercentage > 60 {
		comparison.Insights = append(comparison.Insights, "电脑使用时长占总时长的60%以上，是主要使用设备")
	}
	if comparison.Total.AvgDailyDuration > 28800000 { // 8 hours
		comparison.Insights = append(comparison.Insights, "日均总屏幕时间超过8小时，建议减少使用")
	}
	comparison.Insights = append(comparison.Insights, "手机最常用应用是"+comparison.Phone.TopApp+"，电脑最常用是"+comparison.Computer.TopApp)

	logger.Info("Cross-device comparison completed", logrus.Fields{
		"phone_apps":    comparison.Phone.TotalApps,
		"computer_apps": comparison.Computer.TotalApps,
		"total_duration_hours": float64(comparison.Total.TotalDuration) / 3600000,
	})

	c.JSON(http.StatusOK, comparison)
}

// GetWorkLifeBalance returns work-life balance analysis
// GET /api/v1/screentime/cross-device/work-life-balance
func (h *MultiDeviceHandler) GetWorkLifeBalance(c *gin.Context) {
	var balance WorkLifeBalance

	// Get computer work duration using new query method
	computerSummary, err := h.deviceManager.GetComputerSummary("computer_main")
	if err == nil {
		if totalDuration, ok := computerSummary["totalDurationMS"].(int64); ok {
			balance.WorkDuration = totalDuration
		}
	}

	// Get phone entertainment duration
	phoneConn, err := h.deviceManager.GetDevice("phone_vivo_x90")
	if err == nil {
		// Simplified: assume most phone usage is life/entertainment
		var totalDuration sql.NullInt64
		phoneConn.DB.QueryRow("SELECT SUM(duration_ms) FROM screentime_daily WHERE package_id != 'ALL'").Scan(&totalDuration)
		balance.LifeDuration = totalDuration.Int64
	}

	// Calculate percentages
	total := balance.WorkDuration + balance.LifeDuration
	if total > 0 {
		balance.WorkPercentage = float64(balance.WorkDuration) / float64(total) * 100
		balance.LifePercentage = float64(balance.LifeDuration) / float64(total) * 100
	}

	// Calculate balance score (0-100, ideal is 60% work, 40% life)
	idealWorkRatio := 0.6
	workRatio := float64(balance.WorkDuration) / float64(total)
	deviation := workRatio - idealWorkRatio
	if deviation < 0 {
		deviation = -deviation
	}
	balance.BalanceScore = int(100 - (deviation * 200))
	if balance.BalanceScore < 0 {
		balance.BalanceScore = 0
	}
	if balance.BalanceScore > 100 {
		balance.BalanceScore = 100
	}

	// Generate recommendation
	if balance.WorkPercentage > 70 {
		balance.Recommendation = "工作时长占比较高，建议增加休闲娱乐时间"
	} else if balance.WorkPercentage < 50 {
		balance.Recommendation = "娱乐时长占比较高，建议增加工作学习时间"
	} else {
		balance.Recommendation = "工作生活平衡良好，继续保持"
	}

	// Generate insights
	balance.Insights = []string{
		"电脑使用主要用于工作，手机使用主要用于生活娱乐",
		"建议在工作时段减少手机使用，提高专注力",
	}

	c.JSON(http.StatusOK, balance)
}

// GetTotalScreentime returns total screentime across all devices
// GET /api/v1/screentime/cross-device/total-screentime
func (h *MultiDeviceHandler) GetTotalScreentime(c *gin.Context) {
	startDate := c.DefaultQuery("start", "")
	endDate := c.DefaultQuery("end", "")

	type DailyTotal struct {
		Date            string  `json:"date"`
		PhoneDuration   int64   `json:"phoneDuration"`
		ComputerDuration int64  `json:"computerDuration"`
		TotalDuration   int64   `json:"totalDuration"`
		TotalHours      float64 `json:"totalHours"`
	}

	var dailyTotals []DailyTotal

	// Get all unique dates from both devices
	dates := make(map[string]bool)

	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")
	computerConn, _ := h.deviceManager.GetDevice("computer_main")

	// Get phone dates
	phoneQuery := "SELECT DISTINCT date FROM screentime_daily WHERE package_id != 'ALL'"
	if startDate != "" {
		phoneQuery += " AND date >= '" + startDate + "'"
	}
	if endDate != "" {
		phoneQuery += " AND date <= '" + endDate + "'"
	}
	phoneQuery += " ORDER BY date DESC LIMIT 30"

	rows, _ := phoneConn.DB.Query(phoneQuery)
	for rows.Next() {
		var date string
		rows.Scan(&date)
		dates[date] = true
	}
	rows.Close()

	// For each date, get phone and computer duration
	for date := range dates {
		var daily DailyTotal
		daily.Date = date

		// Get phone duration
		phoneConn.DB.QueryRow("SELECT SUM(duration_ms) FROM screentime_daily WHERE date = ? AND package_id != 'ALL'", date).Scan(&daily.PhoneDuration)

		// Get computer duration
		computerConn.DB.QueryRow("SELECT SUM(total_duration_seconds) * 1000 FROM manictime_daily WHERE date = ?", date).Scan(&daily.ComputerDuration)

		daily.TotalDuration = daily.PhoneDuration + daily.ComputerDuration
		daily.TotalHours = float64(daily.TotalDuration) / 3600000.0

		dailyTotals = append(dailyTotals, daily)
	}

	c.JSON(http.StatusOK, dailyTotals)
}
