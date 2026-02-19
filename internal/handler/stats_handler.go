package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
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
