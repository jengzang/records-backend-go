package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// TrackHandler handles HTTP requests for track points
type TrackHandler struct {
	trackService *service.TrackService
}

// NewTrackHandler creates a new track handler
func NewTrackHandler(trackService *service.TrackService) *TrackHandler {
	return &TrackHandler{
		trackService: trackService,
	}
}

// GetTrackPoints handles GET /api/v1/tracks/points
func (h *TrackHandler) GetTrackPoints(c *gin.Context) {
	var filter models.TrackPointFilter

	// Parse query parameters
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// Get track points
	result, err := h.trackService.GetTrackPoints(filter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetTrackPointByID handles GET /api/v1/tracks/points/:id
func (h *TrackHandler) GetTrackPointByID(c *gin.Context) {
	// Parse ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid track point ID")
		return
	}

	// Get track point
	point, err := h.trackService.GetTrackPointByID(id)
	if err != nil {
		response.NotFound(c, "Track point not found")
		return
	}

	response.Success(c, point)
}

// GetUngeocodedPoints handles GET /api/v1/tracks/ungeocoded
func (h *TrackHandler) GetUngeocodedPoints(c *gin.Context) {
	// Parse limit
	limitStr := c.DefaultQuery("limit", "1000")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequest(c, "Invalid limit parameter")
		return
	}

	// Get ungeocoded points
	points, err := h.trackService.GetUngeocodedPoints(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":  points,
		"count": len(points),
	})
}
