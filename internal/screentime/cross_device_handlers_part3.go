package screentime

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProductivityDeep returns deep productivity analysis across devices
// GET /api/v1/screentime/cross-device/productivity-deep
func (h *MultiDeviceHandler) GetProductivityDeep(c *gin.Context) {
	type ProductivityApp struct {
		AppName    string  `json:"appName"`
		Device     string  `json:"device"`
		Duration   int64   `json:"duration"`
		Percentage float64 `json:"percentage"`
	}

	type ProductivityDeep struct {
		PhoneProductivity struct {
			TotalDuration int64             `json:"totalDuration"`
			TopApps       []ProductivityApp `json:"topApps"`
			Percentage    float64           `json:"percentage"`
		} `json:"phoneProductivity"`
		ComputerProductivity struct {
			TotalDuration int64             `json:"totalDuration"`
			TopApps       []ProductivityApp `json:"topApps"`
			Percentage    float64           `json:"percentage"`
		} `json:"computerProductivity"`
		TotalProductivity int64    `json:"totalProductivity"`
		ProductivityScore int      `json:"productivityScore"`
		BestDevice        string   `json:"bestDevice"`
		Insights          []string `json:"insights"`
	}

	var result ProductivityDeep
	result.PhoneProductivity.TopApps = []ProductivityApp{}
	result.ComputerProductivity.TopApps = []ProductivityApp{}
	result.Insights = []string{}

	// Get phone productivity apps
	phoneConn, err := h.deviceManager.GetDevice("phone_vivo_x90")
	if err == nil && phoneConn != nil && phoneConn.DB != nil {
		phoneQuery := `
		SELECT app_name, SUM(duration_ms) as total_duration
		FROM screentime_daily
		WHERE category IN ('Productivity', 'Tools', 'Office')
		AND package_id != 'ALL'
		GROUP BY app_name
		ORDER BY total_duration DESC
		LIMIT 5
		`
		rows, err := phoneConn.DB.Query(phoneQuery)
		if err == nil && rows != nil {
			defer rows.Close()
			var phoneTotal int64
			for rows.Next() {
				var app ProductivityApp
				if err := rows.Scan(&app.AppName, &app.Duration); err == nil {
					app.Device = "phone"
					phoneTotal += app.Duration
					result.PhoneProductivity.TopApps = append(result.PhoneProductivity.TopApps, app)
				}
			}
			result.PhoneProductivity.TotalDuration = phoneTotal
		}
	}

	// Get computer productivity apps
	computerConn, err := h.deviceManager.GetDevice("computer_main")
	if err == nil && computerConn != nil && computerConn.DB != nil {
		computerQuery := `
		SELECT application, SUM(duration_seconds * 1000) as total_duration
		FROM manictime_activities
		WHERE category IN ('Development', 'Office', 'Productivity')
		GROUP BY application
		ORDER BY total_duration DESC
		LIMIT 5
		`
		rows2, err := computerConn.DB.Query(computerQuery)
		if err == nil && rows2 != nil {
			defer rows2.Close()
			var computerTotal int64
			for rows2.Next() {
				var app ProductivityApp
				if err := rows2.Scan(&app.AppName, &app.Duration); err == nil {
					app.Device = "computer"
					computerTotal += app.Duration
					result.ComputerProductivity.TopApps = append(result.ComputerProductivity.TopApps, app)
				}
			}
			result.ComputerProductivity.TotalDuration = computerTotal
		}
	}

	// Calculate percentages
	result.TotalProductivity = result.PhoneProductivity.TotalDuration + result.ComputerProductivity.TotalDuration
	if result.TotalProductivity > 0 {
		result.PhoneProductivity.Percentage = float64(result.PhoneProductivity.TotalDuration) / float64(result.TotalProductivity) * 100
		result.ComputerProductivity.Percentage = float64(result.ComputerProductivity.TotalDuration) / float64(result.TotalProductivity) * 100
	}

	// Calculate productivity score (0-100)
	totalHours := float64(result.TotalProductivity) / 3600000
	if totalHours > 2000 {
		result.ProductivityScore = 100
	} else if totalHours > 1000 {
		result.ProductivityScore = 80
	} else if totalHours > 500 {
		result.ProductivityScore = 60
	} else {
		result.ProductivityScore = 40
	}

	// Determine best device
	if result.ComputerProductivity.TotalDuration > result.PhoneProductivity.TotalDuration*2 {
		result.BestDevice = "電腦"
	} else if result.PhoneProductivity.TotalDuration > result.ComputerProductivity.TotalDuration*2 {
		result.BestDevice = "手機"
	} else {
		result.BestDevice = "平衡使用"
	}

	// Generate insights
	result.Insights = append(result.Insights,
		fmt.Sprintf("電腦生產力時長占總生產力的%.1f%%", result.ComputerProductivity.Percentage),
		fmt.Sprintf("手機生產力時長占總生產力的%.1f%%", result.PhoneProductivity.Percentage),
		"最適合生產力工作的設備: "+result.BestDevice,
	)

	c.JSON(http.StatusOK, result)
}

// GetFocusAnalysis returns focus analysis across devices
// GET /api/v1/screentime/cross-device/focus-analysis
func (h *MultiDeviceHandler) GetFocusAnalysis(c *gin.Context) {
	type FocusSession struct {
		Date     string `json:"date"`
		Device   string `json:"device"`
		AppName  string `json:"appName"`
		Duration int64  `json:"duration"`
	}

	type FocusAnalysis struct {
		PhoneFocus struct {
			TotalFocusTime   int64          `json:"totalFocusTime"`
			FocusSessions    int            `json:"focusSessions"`
			AvgSessionLength int64          `json:"avgSessionLength"`
			TopFocusApps     []FocusSession `json:"topFocusApps"`
		} `json:"phoneFocus"`
		ComputerFocus struct {
			TotalFocusTime   int64          `json:"totalFocusTime"`
			FocusSessions    int            `json:"focusSessions"`
			AvgSessionLength int64          `json:"avgSessionLength"`
			TopFocusApps     []FocusSession `json:"topFocusApps"`
		} `json:"computerFocus"`
		TotalFocusTime  int64    `json:"totalFocusTime"`
		FocusScore      int      `json:"focusScore"`
		BestFocusDevice string   `json:"bestFocusDevice"`
		Insights        []string `json:"insights"`
	}

	var result FocusAnalysis
	result.PhoneFocus.TopFocusApps = []FocusSession{}
	result.ComputerFocus.TopFocusApps = []FocusSession{}
	result.Insights = []string{}

	// Phone focus: sessions > 15 minutes
	phoneConn, err := h.deviceManager.GetDevice("phone_vivo_x90")
	if err == nil && phoneConn != nil && phoneConn.DB != nil {
		phoneQuery := `
		SELECT date, app_name, duration_ms
		FROM screentime_daily
		WHERE duration_ms > 900000
		AND package_id != 'ALL'
		ORDER BY duration_ms DESC
		LIMIT 10
		`
		rows, err := phoneConn.DB.Query(phoneQuery)
		if err == nil && rows != nil {
			defer rows.Close()
			var phoneTotalFocus int64
			var phoneSessions int
			for rows.Next() {
				var session FocusSession
				if err := rows.Scan(&session.Date, &session.AppName, &session.Duration); err == nil {
					session.Device = "phone"
					phoneTotalFocus += session.Duration
					phoneSessions++
					if len(result.PhoneFocus.TopFocusApps) < 5 {
						result.PhoneFocus.TopFocusApps = append(result.PhoneFocus.TopFocusApps, session)
					}
				}
			}
			result.PhoneFocus.TotalFocusTime = phoneTotalFocus
			result.PhoneFocus.FocusSessions = phoneSessions
			if phoneSessions > 0 {
				result.PhoneFocus.AvgSessionLength = phoneTotalFocus / int64(phoneSessions)
			}
		}
	}

	// Computer focus: sessions > 30 minutes
	computerConn, err := h.deviceManager.GetDevice("computer_main")
	if err == nil && computerConn != nil && computerConn.DB != nil {
		computerQuery := `
		SELECT date, application, duration_seconds * 1000 as duration_ms
		FROM manictime_activities
		WHERE duration_seconds > 1800
		ORDER BY duration_seconds DESC
		LIMIT 10
		`
		rows2, err := computerConn.DB.Query(computerQuery)
		if err == nil && rows2 != nil {
			defer rows2.Close()
			var computerTotalFocus int64
			var computerSessions int
			for rows2.Next() {
				var session FocusSession
				if err := rows2.Scan(&session.Date, &session.AppName, &session.Duration); err == nil {
					session.Device = "computer"
					computerTotalFocus += session.Duration
					computerSessions++
					if len(result.ComputerFocus.TopFocusApps) < 5 {
						result.ComputerFocus.TopFocusApps = append(result.ComputerFocus.TopFocusApps, session)
					}
				}
			}
			result.ComputerFocus.TotalFocusTime = computerTotalFocus
			result.ComputerFocus.FocusSessions = computerSessions
			if computerSessions > 0 {
				result.ComputerFocus.AvgSessionLength = computerTotalFocus / int64(computerSessions)
			}
		}
	}

	// Calculate total focus
	result.TotalFocusTime = result.PhoneFocus.TotalFocusTime + result.ComputerFocus.TotalFocusTime

	// Calculate focus score (0-100)
	totalFocusHours := float64(result.TotalFocusTime) / 3600000
	if totalFocusHours > 1000 {
		result.FocusScore = 100
	} else if totalFocusHours > 500 {
		result.FocusScore = 80
	} else if totalFocusHours > 200 {
		result.FocusScore = 60
	} else {
		result.FocusScore = 40
	}

	// Determine best focus device
	if result.ComputerFocus.TotalFocusTime > result.PhoneFocus.TotalFocusTime*2 {
		result.BestFocusDevice = "電腦"
	} else if result.PhoneFocus.TotalFocusTime > result.ComputerFocus.TotalFocusTime*2 {
		result.BestFocusDevice = "手機"
	} else {
		result.BestFocusDevice = "兩者相當"
	}

	// Generate insights
	result.Insights = append(result.Insights,
		fmt.Sprintf("總專注時長: %.1f小時", float64(result.TotalFocusTime)/3600000),
		fmt.Sprintf("電腦專注會話: %d 次", result.ComputerFocus.FocusSessions),
		fmt.Sprintf("手機專注會話: %d 次", result.PhoneFocus.FocusSessions),
		"最佳專注設備: "+result.BestFocusDevice,
	)

	c.JSON(http.StatusOK, result)
}

// GetCrossDeviceRecommendations returns enhanced smart recommendations
// GET /api/v1/screentime/cross-device/recommendations
func (h *MultiDeviceHandler) GetCrossDeviceRecommendations(c *gin.Context) {
	type Recommendation struct {
		Category    string `json:"category"`
		Priority    string `json:"priority"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Action      string `json:"action"`
	}

	type Recommendations struct {
		HealthRecommendations       []Recommendation `json:"healthRecommendations"`
		ProductivityRecommendations []Recommendation `json:"productivityRecommendations"`
		BalanceRecommendations      []Recommendation `json:"balanceRecommendations"`
		TotalRecommendations        int              `json:"totalRecommendations"`
	}

	var result Recommendations
	result.HealthRecommendations = []Recommendation{}
	result.ProductivityRecommendations = []Recommendation{}
	result.BalanceRecommendations = []Recommendation{}

	// Get total screentime
	phoneConn, _ := h.deviceManager.GetDevice("phone_vivo_x90")
	computerConn, _ := h.deviceManager.GetDevice("computer_main")

	var phoneTotalMS, computerTotalMS int64
	if phoneConn != nil && phoneConn.DB != nil {
		phoneConn.DB.QueryRow("SELECT COALESCE(SUM(duration_ms), 0) FROM screentime_daily WHERE package_id = 'ALL'").Scan(&phoneTotalMS)
	}
	if computerConn != nil && computerConn.DB != nil {
		computerConn.DB.QueryRow("SELECT COALESCE(SUM(duration_seconds * 1000), 0) FROM manictime_activities").Scan(&computerTotalMS)
	}

	totalMS := phoneTotalMS + computerTotalMS
	if totalMS == 0 {
		c.JSON(http.StatusOK, result)
		return
	}

	totalHours := float64(totalMS) / 3600000
	avgDailyHours := totalHours / 365 // Approximate

	// Health recommendations
	if avgDailyHours > 10 {
		result.HealthRecommendations = append(result.HealthRecommendations, Recommendation{
			Category:    "健康",
			Priority:    "高",
			Title:       "總屏幕時間過長",
			Description: fmt.Sprintf("您的日均屏幕時間%.1f小時,建議減少使用", avgDailyHours),
			Action:      "設定每日使用目標,逐步減少至8小時以內",
		})
	}

	if avgDailyHours > 8 {
		result.HealthRecommendations = append(result.HealthRecommendations, Recommendation{
			Category:    "健康",
			Priority:    "中",
			Title:       "建議增加休息時間",
			Description: "長時間使用電子設備可能影響視力和睡眠",
			Action:      "每小時休息10分鐘,遠眺或活動身體",
		})
	}

	// Check late night usage
	if phoneConn != nil && phoneConn.DB != nil {
		var lateNightCount int
		phoneConn.DB.QueryRow(`
			SELECT COUNT(*) FROM screentime_unlocks
			WHERE unlock_time >= '23:00' OR unlock_time < '06:00'
		`).Scan(&lateNightCount)

		if lateNightCount > 100 {
			result.HealthRecommendations = append(result.HealthRecommendations, Recommendation{
				Category:    "健康",
				Priority:    "高",
				Title:       "深夜使用頻繁",
				Description: "檢測到大量深夜使用記錄,可能影響睡眠質量",
				Action:      "建議23:00後關閉電子設備,改善睡眠習慣",
			})
		}
	}

	// Productivity recommendations
	var productivityMS int64
	if phoneConn != nil && phoneConn.DB != nil {
		phoneConn.DB.QueryRow(`
			SELECT COALESCE(SUM(duration_ms), 0) FROM screentime_daily
			WHERE category IN ('Productivity', 'Tools', 'Office')
		`).Scan(&productivityMS)
	}

	var computerProductivityMS int64
	if computerConn != nil && computerConn.DB != nil {
		computerConn.DB.QueryRow(`
			SELECT COALESCE(SUM(duration_seconds * 1000), 0) FROM manictime_activities
			WHERE category IN ('Development', 'Office', 'Productivity')
		`).Scan(&computerProductivityMS)
	}

	productivityMS += computerProductivityMS
	productivityRatio := float64(productivityMS) / float64(totalMS) * 100

	if productivityRatio < 30 {
		result.ProductivityRecommendations = append(result.ProductivityRecommendations, Recommendation{
			Category:    "生產力",
			Priority:    "中",
			Title:       "生產力應用使用較少",
			Description: fmt.Sprintf("生產力應用僅占總使用時長的%.1f%%", productivityRatio),
			Action:      "建議增加工作/學習相關應用的使用時間",
		})
	}

	// Check device balance
	phonePercentage := float64(phoneTotalMS) / float64(totalMS) * 100
	if phonePercentage > 70 {
		result.ProductivityRecommendations = append(result.ProductivityRecommendations, Recommendation{
			Category:    "生產力",
			Priority:    "低",
			Title:       "手機使用占比過高",
			Description: fmt.Sprintf("手機使用占總時長的%.1f%%", phonePercentage),
			Action:      "對於需要專注的工作,建議優先使用電腦",
		})
	}

	// Balance recommendations
	if phonePercentage > 60 || phonePercentage < 40 {
		result.BalanceRecommendations = append(result.BalanceRecommendations, Recommendation{
			Category:    "平衡",
			Priority:    "低",
			Title:       "設備使用不均衡",
			Description: "手機和電腦使用時長差異較大",
			Action:      "根據任務特性選擇合適的設備,提高效率",
		})
	}

	// Check entertainment vs work
	var entertainmentMS int64
	if phoneConn != nil && phoneConn.DB != nil {
		phoneConn.DB.QueryRow(`
			SELECT COALESCE(SUM(duration_ms), 0) FROM screentime_daily
			WHERE category IN ('Entertainment', 'Social', 'Gaming')
		`).Scan(&entertainmentMS)
	}

	entertainmentRatio := float64(entertainmentMS) / float64(totalMS) * 100
	if entertainmentRatio > 50 {
		result.BalanceRecommendations = append(result.BalanceRecommendations, Recommendation{
			Category:    "平衡",
			Priority:    "中",
			Title:       "娛樂時間占比較高",
			Description: fmt.Sprintf("娛樂應用占總時長的%.1f%%", entertainmentRatio),
			Action:      "建議平衡工作與娛樂,提高時間利用效率",
		})
	}

	result.TotalRecommendations = len(result.HealthRecommendations) +
		len(result.ProductivityRecommendations) +
		len(result.BalanceRecommendations)

	c.JSON(http.StatusOK, result)
}
