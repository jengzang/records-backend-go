package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// SegmentHandler handles HTTP requests for segments
type SegmentHandler struct {
	service *service.SegmentService
}

// NewSegmentHandler creates a new segment handler
func NewSegmentHandler(service *service.SegmentService) *SegmentHandler {
	return &SegmentHandler{service: service}
}

// GetSegments handles GET /api/v1/tracks/segments
func (h *SegmentHandler) GetSegments(c *gin.Context) {
	var filter models.SegmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	segments, total, err := h.service.GetSegments(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get segments", err)
		return
	}

	// Calculate pagination info
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 100
	}
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"data":       segments,
		"total":      total,
		"page":       filter.Page,
		"pageSize":   filter.PageSize,
		"totalPages": totalPages,
	})
}

// GetSegmentByID handles GET /api/v1/tracks/segments/:id
func (h *SegmentHandler) GetSegmentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid segment ID", err)
		return
	}

	segment, err := h.service.GetSegmentByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get segment", err)
		return
	}

	if segment == nil {
		response.Error(c, http.StatusNotFound, "Segment not found", nil)
		return
	}

	response.Success(c, segment)
}
