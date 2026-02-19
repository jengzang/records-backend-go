package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// GridHandler handles HTTP requests for grid cells
type GridHandler struct {
	service *service.GridService
}

// NewGridHandler creates a new grid handler
func NewGridHandler(service *service.GridService) *GridHandler {
	return &GridHandler{service: service}
}

// GetGridCells handles GET /api/v1/viz/grid-cells
func (h *GridHandler) GetGridCells(c *gin.Context) {
	var filter models.GridFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Default to level 3 (district level) if not specified
	if filter.Level == 0 {
		filter.Level = 3
	}

	cells, err := h.service.GetGridCells(filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get grid cells", err)
		return
	}

	response.Success(c, gin.H{
		"data":  cells,
		"count": len(cells),
	})
}
