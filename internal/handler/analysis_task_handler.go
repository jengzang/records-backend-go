package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/service"
	"github.com/jengzang/records-backend-go/pkg/response"
)

// AnalysisTaskHandler handles HTTP requests for analysis tasks
type AnalysisTaskHandler struct {
	service *service.AnalysisTaskService
}

// NewAnalysisTaskHandler creates a new analysis task handler
func NewAnalysisTaskHandler(service *service.AnalysisTaskService) *AnalysisTaskHandler {
	return &AnalysisTaskHandler{service: service}
}

// CreateTaskRequest represents the request body for creating an analysis task
type CreateTaskRequest struct {
	SkillName string                 `json:"skill_name" binding:"required"`
	TaskType  string                 `json:"task_type" binding:"required"` // INCREMENTAL or FULL_RECOMPUTE
	Params    map[string]interface{} `json:"params"`
}

// CreateTask creates a new analysis task
// POST /api/admin/analysis/tasks
func (h *AnalysisTaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user from context (set by auth middleware)
	createdBy := c.GetString("user")
	if createdBy == "" {
		createdBy = "admin" // Default for now
	}

	task, err := h.service.CreateTask(req.SkillName, req.TaskType, req.Params, createdBy)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, task)
}

// GetTask retrieves a task by ID
// GET /api/admin/analysis/tasks/:id
func (h *AnalysisTaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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
// GET /api/admin/analysis/tasks
func (h *AnalysisTaskHandler) ListTasks(c *gin.Context) {
	skillName := c.Query("skill_name")
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

	tasks, err := h.service.ListTasks(skillName, status, limit, offset)
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
// DELETE /api/admin/analysis/tasks/:id
func (h *AnalysisTaskHandler) CancelTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

// TriggerAnalysisChainRequest represents the request body for triggering an analysis chain
type TriggerAnalysisChainRequest struct {
	TaskType string `json:"task_type" binding:"required"` // INCREMENTAL or FULL_RECOMPUTE
}

// TriggerAnalysisChain triggers a complete analysis chain
// POST /api/admin/analysis/trigger-chain
func (h *AnalysisTaskHandler) TriggerAnalysisChain(c *gin.Context) {
	var req TriggerAnalysisChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user from context
	createdBy := c.GetString("user")
	if createdBy == "" {
		createdBy = "admin"
	}

	taskIDs, err := h.service.TriggerAnalysisChain(req.TaskType, createdBy)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message":  "Analysis chain triggered successfully",
		"task_ids": taskIDs,
	})
}
