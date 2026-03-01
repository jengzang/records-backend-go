package health

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for health data
type Handler struct {
	service *Service
}

// NewHandler creates a new health handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		service: NewService(db),
	}
}

// GetSummary handles GET /api/v1/health/summary
func (h *Handler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetRecords handles GET /api/v1/health/records
func (h *Handler) GetRecords(c *gin.Context) {
	// Parse query parameters
	recordType := c.Query("type")
	startDateStr := c.Query("start")
	endDateStr := c.Query("end")
	limitStr := c.DefaultQuery("limit", "100")

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
			return
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use YYYY-MM-DD"})
			return
		}
	}

	// Create filter
	filter := RecordFilter{
		Type:      recordType,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
	}

	// Get records
	records, err := h.service.GetRecords(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"count":   len(records),
	})
}

// GetWorkouts handles GET /api/v1/health/workouts
func (h *Handler) GetWorkouts(c *gin.Context) {
	// Parse query parameters
	workoutType := c.Query("type")
	startDateStr := c.Query("start")
	endDateStr := c.Query("end")
	limitStr := c.DefaultQuery("limit", "50")

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
			return
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use YYYY-MM-DD"})
			return
		}
	}

	// Create filter
	filter := WorkoutFilter{
		WorkoutType: workoutType,
		StartDate:   startDate,
		EndDate:     endDate,
		Limit:       limit,
	}

	// Get workouts
	workouts, err := h.service.GetWorkouts(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workouts": workouts,
		"count":    len(workouts),
	})
}

// GetWorkout handles GET /api/v1/health/workouts/:id
func (h *Handler) GetWorkout(c *gin.Context) {
	// Parse workout ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout ID"})
		return
	}

	// Get workout
	workout, err := h.service.GetWorkoutByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		return
	}

	c.JSON(http.StatusOK, workout)
}

// GetWorkoutRoute handles GET /api/v1/health/workouts/:id/route
func (h *Handler) GetWorkoutRoute(c *gin.Context) {
	// Parse workout ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout ID"})
		return
	}

	// Get route
	route, err := h.service.GetWorkoutRoute(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workoutId": id,
		"points":    route,
		"count":     len(route),
	})
}

// GetDailyStatistics handles GET /api/v1/health/statistics/daily
func (h *Handler) GetDailyStatistics(c *gin.Context) {
	// Parse query parameters
	metricType := c.Query("metric")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric parameter is required"})
		return
	}

	startDateStr := c.Query("start")
	endDateStr := c.Query("end")

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to today
		endDate = time.Now()
	}

	// Get statistics
	stats, err := h.service.GetDailyStatistics(metricType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metric":     metricType,
		"startDate":  startDate.Format("2006-01-02"),
		"endDate":    endDate.Format("2006-01-02"),
		"statistics": stats,
		"count":      len(stats),
	})
}

// GetWeeklyStatistics handles GET /api/v1/health/statistics/weekly
func (h *Handler) GetWeeklyStatistics(c *gin.Context) {
	// Parse query parameters
	metricType := c.Query("metric")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric parameter is required"})
		return
	}

	startDateStr := c.DefaultQuery("start", time.Now().AddDate(0, -3, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	// Get statistics
	stats, err := h.service.GetWeeklyStatistics(metricType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metric":     metricType,
		"startDate":  startDate.Format("2006-01-02"),
		"endDate":    endDate.Format("2006-01-02"),
		"statistics": stats,
		"count":      len(stats),
	})
}

// GetMonthlyStatistics handles GET /api/v1/health/statistics/monthly
func (h *Handler) GetMonthlyStatistics(c *gin.Context) {
	// Parse query parameters
	metricType := c.Query("metric")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric parameter is required"})
		return
	}

	startDateStr := c.DefaultQuery("start", time.Now().AddDate(-1, 0, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	// Get statistics
	stats, err := h.service.GetMonthlyStatistics(metricType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metric":     metricType,
		"startDate":  startDate.Format("2006-01-02"),
		"endDate":    endDate.Format("2006-01-02"),
		"statistics": stats,
		"count":      len(stats),
	})
}

// GetTrends handles GET /api/v1/health/statistics/trends
func (h *Handler) GetTrends(c *gin.Context) {
	metricType := c.Query("metric")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric parameter is required"})
		return
	}

	period := c.DefaultQuery("period", "monthly")

	trends, err := h.service.GetTrends(metricType, period)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "trends analysis not yet implemented"})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// GetSleepStatistics handles GET /api/v1/health/statistics/sleep
func (h *Handler) GetSleepStatistics(c *gin.Context) {
	startDateStr := c.DefaultQuery("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	sleepData, err := h.service.GetSleepAnalysis(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "sleep analysis not yet implemented"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
		"data":      sleepData,
		"count":     len(sleepData),
	})
}

// GetActivityPatterns handles GET /api/v1/health/analysis/activity-patterns
func (h *Handler) GetActivityPatterns(c *gin.Context) {
	patterns, err := h.service.GetActivityPatterns()
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "activity pattern analysis not yet implemented"})
		return
	}

	c.JSON(http.StatusOK, patterns)
}

// GetHealthScore handles GET /api/v1/health/analysis/health-score
func (h *Handler) GetHealthScore(c *gin.Context) {
	score, err := h.service.CalculateHealthScore()
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "health score calculation not yet implemented"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"score":       score,
		"maxScore":    100,
		"description": "Overall health score based on activity, sleep, and vital signs",
	})
}

// Analysis endpoint handlers

// GetHeartRateZones handles GET /api/v1/health/analysis/heartrate/zones
func (h *Handler) GetHeartRateZones(c *gin.Context) {
	startDateStr := c.DefaultQuery("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	zones, err := h.service.GetHeartRateZones(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zones)
}

// GetHeartRateAnomalies handles GET /api/v1/health/analysis/heartrate/anomalies
func (h *Handler) GetHeartRateAnomalies(c *gin.Context) {
	startDateStr := c.DefaultQuery("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	anomalies, err := h.service.GetHeartRateAnomalies(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

// GetRestingHeartRate handles GET /api/v1/health/analysis/heartrate/resting
func (h *Handler) GetRestingHeartRate(c *gin.Context) {
	startDateStr := c.DefaultQuery("start", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	restingHR, err := h.service.GetRestingHeartRate(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  restingHR,
		"count": len(restingHR),
	})
}

// GetDailyActivityPattern handles GET /api/v1/health/analysis/patterns/daily
func (h *Handler) GetDailyActivityPattern(c *gin.Context) {
	pattern, err := h.service.GetDailyActivityPattern()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pattern)
}

// GetWeeklyActivityPattern handles GET /api/v1/health/analysis/patterns/weekly
func (h *Handler) GetWeeklyActivityPattern(c *gin.Context) {
	pattern, err := h.service.GetWeeklyActivityPattern()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pattern)
}

// GetHealthScoreForDate handles GET /api/v1/health/analysis/health-score
func (h *Handler) GetHealthScoreForDate(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	score, err := h.service.GetHealthScoreForDate(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, score)
}
