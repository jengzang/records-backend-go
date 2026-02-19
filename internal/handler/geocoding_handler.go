package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// GeocodingHandler handles HTTP requests for geocoding tasks
type GeocodingHandler struct {
	service *service.GeocodingService
}

// NewGeocodingHandler creates a new geocoding handler
func NewGeocodingHandler(service *service.GeocodingService) *GeocodingHandler {
	return &GeocodingHandler{service: service}
}

// CreateTask creates a new geocoding task
// POST /api/admin/geocoding/tasks
func (h *GeocodingHandler) CreateTask(c *gin.Context) {
	// Get user from context (set by auth middleware)
	createdBy := c.GetString("user")
	if createdBy == "" {
		createdBy = "admin" // Default for now
	}

	task, err := h.service.CreateTask(createdBy)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, task)
}

// GetTask retrieves a task by ID
// GET /api/admin/geocoding/tasks/:id
func (h *GeocodingHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.service.GetTask(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, task)
}

// ListTasks retrieves all tasks
// GET /api/admin/geocoding/tasks
func (h *GeocodingHandler) ListTasks(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	tasks, err := h.service.ListTasks(status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
	})
}

// CancelTask cancels a running task
// DELETE /api/admin/geocoding/tasks/:id
func (h *GeocodingHandler) CancelTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if err := h.service.CancelTask(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Task cancelled successfully"})
}
