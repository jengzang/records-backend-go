package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/service"
)

type ImportHandler struct {
	importService *service.ImportService
}

func NewImportHandler(importService *service.ImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// ImportData handles file upload and import
// POST /api/v1/admin/tracks/import
func (h *ImportHandler) ImportData(c *gin.Context) {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file extension
	ext := filepath.Ext(file.Filename)
	if ext != ".csv" && ext != ".xlsx" && ext != ".xls" && ext != ".xlsm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Only CSV and Excel files are supported"})
		return
	}

	// Parse form parameters
	mode := c.DefaultPostForm("mode", "append")
	if mode != "append" && mode != "replace" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mode. Must be 'append' or 'replace'"})
		return
	}

	deduplicate := c.DefaultPostForm("deduplicate", "true") == "true"
	autoTrigger := c.DefaultPostForm("auto_trigger", "true") == "true"

	// Create import task
	task, err := h.importService.CreateImportTask(file.Filename, file.Size, mode, deduplicate, autoTrigger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create import task: %s", err.Error())})
		return
	}

	// Save uploaded file
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%d_%s_%s", task.ID, timestamp, file.Filename)
	filePath := h.importService.GetUploadPath(fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file: %s", err.Error())})
		return
	}

	// Execute import in background
	go func() {
		if err := h.importService.ExecuteImport(task.ID, filePath); err != nil {
			// Error is already logged in the task
			fmt.Printf("Import task %d failed: %s\n", task.ID, err.Error())
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"status":  task.Status,
		"message": "Import task created successfully",
	})
}

// GetImportStatus retrieves the status of an import task
// GET /api/v1/admin/tracks/import/:id
func (h *ImportHandler) GetImportStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.importService.GetImportTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Import task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
