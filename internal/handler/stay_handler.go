package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// StayHandler handles HTTP requests for stay segments
type StayHandler struct {
	service *service.StayService
}

// NewStayHandler creates a new stay handler
func NewStayHandler(service *service.StayService) *StayHandler {
	return &StayHandler{service: service}
}

// GetStays handles GET /api/v1/tracks/stays
func (h *StayHandler) GetStays(c *gin.Context) {
	var filter models.StayFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	stays, total, err := h.service.GetStays(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get stay segments", err)
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
		"data":       stays,
		"total":      total,
		"page":       filter.Page,
		"pageSize":   filter.PageSize,
		"totalPages": totalPages,
	})
}

// GetStayByID handles GET /api/v1/tracks/stays/:id
func (h *StayHandler) GetStayByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stay segment ID", err)
		return
	}

	stay, err := h.service.GetStayByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get stay segment", err)
		return
	}

	if stay == nil {
		response.Error(c, http.StatusNotFound, "Stay segment not found", nil)
		return
	}

	response.Success(c, stay)
}
