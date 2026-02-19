package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// TripHandler handles HTTP requests for trips
type TripHandler struct {
	service *service.TripService
}

// NewTripHandler creates a new trip handler
func NewTripHandler(service *service.TripService) *TripHandler {
	return &TripHandler{service: service}
}

// GetTrips handles GET /api/v1/tracks/trips
func (h *TripHandler) GetTrips(c *gin.Context) {
	var filter models.TripFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	trips, total, err := h.service.GetTrips(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get trips", err)
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
		"data":       trips,
		"total":      total,
		"page":       filter.Page,
		"pageSize":   filter.PageSize,
		"totalPages": totalPages,
	})
}

// GetTripByID handles GET /api/v1/tracks/trips/:id
func (h *TripHandler) GetTripByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid trip ID", err)
		return
	}

	trip, err := h.service.GetTripByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get trip", err)
		return
	}

	if trip == nil {
		response.Error(c, http.StatusNotFound, "Trip not found", nil)
		return
	}

	response.Success(c, trip)
}
