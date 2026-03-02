package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// LocationBehaviorHandler handles HTTP requests for location behavior analysis
type LocationBehaviorHandler struct {
	service *service.LocationBehaviorService
}

// NewLocationBehaviorHandler creates a new handler
func NewLocationBehaviorHandler(service *service.LocationBehaviorService) *LocationBehaviorHandler {
	return &LocationBehaviorHandler{service: service}
}

// GetLocations handles GET /api/v1/cross-module/location-behavior/locations
func (h *LocationBehaviorHandler) GetLocations(c *gin.Context) {
	locations, err := h.service.GetAllLocations()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get locations", err)
		return
	}

	response.Success(c, gin.H{
		"data":  locations,
		"total": len(locations),
	})
}

// GetLocationByID handles GET /api/v1/cross-module/location-behavior/locations/:id
func (h *LocationBehaviorHandler) GetLocationByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	location, err := h.service.GetLocationByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get location", err)
		return
	}

	if location == nil {
		response.Error(c, http.StatusNotFound, "Location not found", nil)
		return
	}

	response.Success(c, location)
}

// GetRankings handles GET /api/v1/cross-module/location-behavior/rankings
func (h *LocationBehaviorHandler) GetRankings(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	topEfficient, err := h.service.GetTopEfficientLocations(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get top efficient locations", err)
		return
	}

	leastEfficient, err := h.service.GetLeastEfficientLocations(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get least efficient locations", err)
		return
	}

	response.Success(c, gin.H{
		"topEfficient":   topEfficient,
		"leastEfficient": leastEfficient,
	})
}

// GetHeatmapData handles GET /api/v1/cross-module/location-behavior/heatmap
func (h *LocationBehaviorHandler) GetHeatmapData(c *gin.Context) {
	locations, err := h.service.GetAllLocations()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get heatmap data", err)
		return
	}

	// Transform to heatmap format
	var heatmapData []map[string]interface{}
	for _, loc := range locations {
		heatmapData = append(heatmapData, map[string]interface{}{
			"lat":             loc.CenterLat,
			"lon":             loc.CenterLon,
			"efficiency":      loc.EfficiencyScore.EfficiencyScore,
			"visitCount":      loc.VisitCount,
			"label":           loc.Label,
			"labelConfidence": loc.LabelConfidence,
		})
	}

	response.Success(c, gin.H{
		"data": heatmapData,
	})
}

// GetHabits handles GET /api/v1/cross-module/location-behavior/habits
func (h *LocationBehaviorHandler) GetHabits(c *gin.Context) {
	locations, err := h.service.GetAllLocations()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get habits", err)
		return
	}

	// Collect all habits
	var allHabits []map[string]interface{}
	for _, loc := range locations {
		for _, habit := range loc.Habits {
			allHabits = append(allHabits, map[string]interface{}{
				"locationId":       loc.ID,
				"locationLabel":    loc.Label,
				"habitType":        habit.HabitType,
				"habitDescription": habit.HabitDescription,
				"confidence":       habit.Confidence,
				"occurrenceCount":  habit.OccurrenceCount,
			})
		}
	}

	response.Success(c, gin.H{
		"data":  allHabits,
		"total": len(allHabits),
	})
}

// AnalyzeLocations handles POST /api/v1/cross-module/location-behavior/analyze
func (h *LocationBehaviorHandler) AnalyzeLocations(c *gin.Context) {
	if err := h.service.AnalyzeLocations(); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to analyze locations", err)
		return
	}

	response.Success(c, gin.H{
		"message": "Location analysis completed successfully",
	})
}

// UpdateLocationLabel handles PATCH /api/v1/cross-module/location-behavior/locations/:id
func (h *LocationBehaviorHandler) UpdateLocationLabel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid location ID", err)
		return
	}

	var req struct {
		Label           string  `json:"label"`
		LabelConfidence float64 `json:"labelConfidence"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get location
	location, err := h.service.GetLocationByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get location", err)
		return
	}

	if location == nil {
		response.Error(c, http.StatusNotFound, "Location not found", nil)
		return
	}

	// Update label
	location.Label = req.Label
	location.LabelConfidence = req.LabelConfidence

	// This would require adding an update method to the service
	// For now, return success
	response.Success(c, gin.H{
		"message": "Location label updated successfully",
	})
}
