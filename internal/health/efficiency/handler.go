package efficiency

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for efficiency endpoints
type Handler struct {
	service *Service
}

// NewHandler creates a new efficiency handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers efficiency routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	efficiency := router.Group("/efficiency-curve")
	{
		efficiency.GET("/hourly", h.GetHourlyCurve)
		efficiency.GET("/profile", h.GetProfile)
		efficiency.GET("/comparison", h.GetComparison)
		efficiency.GET("/insights", h.GetInsights)
		efficiency.POST("/analyze", h.AnalyzeEfficiency)
	}
}

// GetHourlyCurve godoc
// @Summary Get hourly efficiency curve
// @Description Retrieves hourly efficiency scores for a date range
// @Tags Efficiency
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)" default(30 days ago)
// @Param end_date query string false "End date (YYYY-MM-DD)" default(today)
// @Success 200 {object} EfficiencyCurveResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cross-module/efficiency-curve/hourly [get]
func (h *Handler) GetHourlyCurve(c *gin.Context) {
	// Parse query parameters
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))

	// Validate dates
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, expected YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, expected YYYY-MM-DD"})
		return
	}

	// Get hourly curve
	response, err := h.service.GetHourlyCurve(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile godoc
// @Summary Get efficiency curve profile
// @Description Retrieves aggregated efficiency profile (workday or weekend)
// @Tags Efficiency
// @Accept json
// @Produce json
// @Param profile_type query string true "Profile type (workday or weekend)" default(workday)
// @Success 200 {object} EfficiencyCurveProfile
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cross-module/efficiency-curve/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	profileType := c.DefaultQuery("profile_type", "workday")

	// Validate profile type
	if profileType != "workday" && profileType != "weekend" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile_type, expected 'workday' or 'weekend'"})
		return
	}

	// Get profile
	profile, err := h.service.GetProfile(profileType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetComparison godoc
// @Summary Get workday vs weekend comparison
// @Description Retrieves comparison between workday and weekend efficiency profiles
// @Tags Efficiency
// @Accept json
// @Produce json
// @Success 200 {object} ProfileComparisonResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cross-module/efficiency-curve/comparison [get]
func (h *Handler) GetComparison(c *gin.Context) {
	comparison, err := h.service.GetComparison()
	if err != nil {
		if err.Error() == "workday profile not found" || err.Error() == "weekend profile not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// GetInsights godoc
// @Summary Get efficiency insights
// @Description Retrieves actionable insights and recommendations
// @Tags Efficiency
// @Accept json
// @Produce json
// @Success 200 {array} EfficiencyInsight
// @Failure 500 {object} map[string]string
// @Router /api/v1/cross-module/efficiency-curve/insights [get]
func (h *Handler) GetInsights(c *gin.Context) {
	insights, err := h.service.GetInsights()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// AnalyzeEfficiency godoc
// @Summary Trigger efficiency analysis
// @Description Analyzes efficiency for a date range and generates profiles
// @Tags Efficiency
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)" default(30 days ago)
// @Param end_date query string false "End date (YYYY-MM-DD)" default(today)
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/cross-module/efficiency-curve/analyze [post]
func (h *Handler) AnalyzeEfficiency(c *gin.Context) {
	// Parse query parameters
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))

	// Validate dates
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, expected YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, expected YYYY-MM-DD"})
		return
	}

	// Trigger analysis
	if err := h.service.AnalyzeEfficiency(startDate, endDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Efficiency analysis completed successfully",
		"start_date": startDate,
		"end_date":   endDate,
	})
}
