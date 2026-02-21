package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// StatsHandler handles HTTP requests for statistics
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetFootprintStatistics handles GET /api/v1/tracks/statistics/footprint
func (h *StatsHandler) GetFootprintStatistics(c *gin.Context) {
	// Parse time range
	startTimeStr := c.DefaultQuery("startTime", "0")
	endTimeStr := c.DefaultQuery("endTime", "0")

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid startTime parameter")
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid endTime parameter")
		return
	}

	// Get statistics
	stats, err := h.statsService.GetFootprintStatistics(startTime, endTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// GetTimeDistribution handles GET /api/v1/tracks/statistics/time-distribution
func (h *StatsHandler) GetTimeDistribution(c *gin.Context) {
	// Parse time range
	startTimeStr := c.DefaultQuery("startTime", "0")
	endTimeStr := c.DefaultQuery("endTime", "0")

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid startTime parameter")
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid endTime parameter")
		return
	}

	// Get distribution
	distribution, err := h.statsService.GetTimeDistribution(startTime, endTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, distribution)
}

// GetSpeedDistribution handles GET /api/v1/tracks/statistics/speed-distribution
func (h *StatsHandler) GetSpeedDistribution(c *gin.Context) {
	// Parse time range
	startTimeStr := c.DefaultQuery("startTime", "0")
	endTimeStr := c.DefaultQuery("endTime", "0")

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid startTime parameter")
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid endTime parameter")
		return
	}

	// Get distribution
	distribution, err := h.statsService.GetSpeedDistribution(startTime, endTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, distribution)
}

// GetFootprintRankings handles GET /api/v1/stats/footprint/rankings
func (h *StatsHandler) GetFootprintRankings(c *gin.Context) {
	var filter models.StatsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Default values
	if filter.StatType == "" {
		filter.StatType = "PROVINCE"
	}
	if filter.TimeRange == "" {
		filter.TimeRange = "all"
	}
	if filter.OrderBy == "" {
		filter.OrderBy = "points"
	}
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	rankings, err := h.statsService.GetFootprintRankings(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get footprint rankings", err)
		return
	}

	response.Success(c, gin.H{
		"data":  rankings,
		"count": len(rankings),
	})
}

// GetStayRankings handles GET /api/v1/stats/stay/rankings
func (h *StatsHandler) GetStayRankings(c *gin.Context) {
	var filter models.StatsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Default values
	if filter.StatType == "" {
		filter.StatType = "PROVINCE"
	}
	if filter.TimeRange == "" {
		filter.TimeRange = "all"
	}
	if filter.OrderBy == "" {
		filter.OrderBy = "count"
	}
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	rankings, err := h.statsService.GetStayRankings(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get stay rankings", err)
		return
	}

	response.Success(c, gin.H{
		"data":  rankings,
		"count": len(rankings),
	})
}

// GetExtremeEvents handles GET /api/v1/stats/extreme-events
func (h *StatsHandler) GetExtremeEvents(c *gin.Context) {
	eventType := c.Query("eventType")
	eventCategory := c.Query("eventCategory")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	events, err := h.statsService.GetExtremeEvents(eventType, eventCategory, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get extreme events", err)
		return
	}

	response.Success(c, gin.H{
		"data":  events,
		"count": len(events),
	})
}

// GetAdminCrossings handles GET /api/v1/stats/admin-crossings
func (h *StatsHandler) GetAdminCrossings(c *gin.Context) {
	crossingType := c.Query("crossing_type")
	fromRegion := c.Query("from")
	toRegion := c.Query("to")
	startTimeStr := c.DefaultQuery("start_time", "0")
	endTimeStr := c.DefaultQuery("end_time", "0")
	limitStr := c.DefaultQuery("limit", "100")

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start_time parameter", err)
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end_time parameter", err)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	crossings, err := h.statsService.GetAdminCrossings(crossingType, fromRegion, toRegion, startTime, endTime, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get admin crossings", err)
		return
	}

	response.Success(c, gin.H{
		"data":  crossings,
		"count": len(crossings),
	})
}

// GetAdminView handles GET /api/v1/stats/admin-view
func (h *StatsHandler) GetAdminView(c *gin.Context) {
	adminLevel := c.Query("admin_level")
	adminName := c.Query("admin_name")
	parentName := c.Query("parent_name")
	sortBy := c.DefaultQuery("sort_by", "visit_count")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid limit parameter", err)
		return
	}

	stats, err := h.statsService.GetAdminStats(adminLevel, adminName, parentName, sortBy, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get admin stats", err)
		return
	}

	response.Success(c, gin.H{
		"data":  stats,
		"count": len(stats),
	})
}
