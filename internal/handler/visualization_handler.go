package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// VisualizationHandler handles HTTP requests for visualization data
type VisualizationHandler struct {
	service *service.VisualizationService
}

// NewVisualizationHandler creates a new visualization handler
func NewVisualizationHandler(service *service.VisualizationService) *VisualizationHandler {
	return &VisualizationHandler{service: service}
}

// GetRenderingMetadata handles GET /api/v1/viz/rendering
func (h *VisualizationHandler) GetRenderingMetadata(c *gin.Context) {
	var filter models.RenderFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Default LOD level
	if filter.LODLevel == 0 {
		filter.LODLevel = 3
	}

	// Default limit
	if filter.Limit == 0 {
		filter.Limit = 10000
	}

	points, err := h.service.GetRenderingMetadata(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get rendering metadata", err)
		return
	}

	response.Success(c, gin.H{
		"data":  points,
		"count": len(points),
	})
}

// GetTimeSliceData handles GET /api/v1/viz/time-slices
func (h *VisualizationHandler) GetTimeSliceData(c *gin.Context) {
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	granularity := c.DefaultQuery("granularity", "day")

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid startTime parameter", err)
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid endTime parameter", err)
		return
	}

	data, err := h.service.GetTimeSliceData(startTime, endTime, granularity)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get time slice data", err)
		return
	}

	response.Success(c, data)
}
