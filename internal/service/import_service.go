package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jengzang/records-backend-go/internal/models"
)

type ImportService struct {
	db                   *sql.DB
	uploadDir            string
	pythonScriptPath     string
	geocodingService     *GeocodingService
	analysisTaskService  *AnalysisTaskService
}

func NewImportService(db *sql.DB) *ImportService {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/tmp/uploads"
	}

	// Ensure upload directory exists
	os.MkdirAll(uploadDir, 0755)

	return &ImportService{
		db:                   db,
		uploadDir:            uploadDir,
		pythonScriptPath:     "scripts/tracks/import/write2sql.py",
		geocodingService:     nil, // Will be set via SetGeocodingService
		analysisTaskService:  nil, // Will be set via SetAnalysisTaskService
	}
}

// SetGeocodingService sets the geocoding service (for dependency injection)
func (s *ImportService) SetGeocodingService(service *GeocodingService) {
	s.geocodingService = service
}

// SetAnalysisTaskService sets the analysis task service (for dependency injection)
func (s *ImportService) SetAnalysisTaskService(service *AnalysisTaskService) {
	s.analysisTaskService = service
}

// CreateImportTask creates a new import task record
func (s *ImportService) CreateImportTask(fileName string, fileSize int64, mode string, deduplicate bool, autoTrigger bool) (*models.ImportTask, error) {
	task := &models.ImportTask{
		Status:      "pending",
		FileName:    fileName,
		FileSize:    fileSize,
		Mode:        mode,
		Deduplicate: deduplicate,
		AutoTrigger: autoTrigger,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO import_tasks (status, file_name, file_size, mode, deduplicate, auto_trigger, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(query, task.Status, task.FileName, task.FileSize, task.Mode, task.Deduplicate, task.AutoTrigger, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create import task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	task.ID = id

	return task, nil
}

// GetImportTask retrieves an import task by ID
func (s *ImportService) GetImportTask(id int64) (*models.ImportTask, error) {
	task := &models.ImportTask{}
	query := `
		SELECT id, status, file_path, file_name, file_size, mode, deduplicate, auto_trigger,
		       total_records, new_records, duplicate_records, error_message,
		       created_at, updated_at, completed_at
		FROM import_tasks
		WHERE id = ?
	`
	err := s.db.QueryRow(query, id).Scan(
		&task.ID, &task.Status, &task.FilePath, &task.FileName, &task.FileSize,
		&task.Mode, &task.Deduplicate, &task.AutoTrigger,
		&task.TotalRecords, &task.NewRecords, &task.DuplicateRecords, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get import task: %w", err)
	}
	return task, nil
}

// ListImportTasks retrieves all import tasks ordered by creation time (newest first)
func (s *ImportService) ListImportTasks(limit int, offset int) ([]*models.ImportTask, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	query := `
		SELECT id, status, file_path, file_name, file_size, mode, deduplicate, auto_trigger,
		       total_records, new_records, duplicate_records, error_message,
		       created_at, updated_at, completed_at
		FROM import_tasks
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list import tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.ImportTask
	for rows.Next() {
		task := &models.ImportTask{}
		err := rows.Scan(
			&task.ID, &task.Status, &task.FilePath, &task.FileName, &task.FileSize,
			&task.Mode, &task.Deduplicate, &task.AutoTrigger,
			&task.TotalRecords, &task.NewRecords, &task.DuplicateRecords, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating import tasks: %w", err)
	}

	return tasks, nil
}

// UpdateImportTask updates an import task
func (s *ImportService) UpdateImportTask(task *models.ImportTask) error {
	task.UpdatedAt = time.Now()
	query := `
		UPDATE import_tasks
		SET status = ?, file_path = ?, total_records = ?, new_records = ?, duplicate_records = ?,
		    error_message = ?, updated_at = ?, completed_at = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, task.Status, task.FilePath, task.TotalRecords, task.NewRecords,
		task.DuplicateRecords, task.ErrorMessage, task.UpdatedAt, task.CompletedAt, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update import task: %w", err)
	}
	return nil
}

// ExecuteImport executes the Python import script
func (s *ImportService) ExecuteImport(taskID int64, filePath string) error {
	task, err := s.GetImportTask(taskID)
	if err != nil {
		return err
	}

	// Update status to running
	task.Status = "running"
	task.FilePath = filePath
	if err := s.UpdateImportTask(task); err != nil {
		return err
	}

	// Build command
	dbPath := os.Getenv("TRACKS_DB_PATH")
	if dbPath == "" {
		dbPath = "data/tracks.db"
	}

	deduplicateStr := "false"
	if task.Deduplicate {
		deduplicateStr = "true"
	}

	cmd := exec.Command("python3",
		s.pythonScriptPath,
		"--file", filePath,
		"--db", dbPath,
		"--table", "一生足迹",
		"--mode", task.Mode,
		"--deduplicate", deduplicateStr,
	)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = fmt.Sprintf("Python script error: %s\nOutput: %s", err.Error(), string(output))
		s.UpdateImportTask(task)
		return fmt.Errorf("failed to execute import: %w", err)
	}

	// Parse JSON output from Python script
	outputStr := string(output)
	jsonStart := -1
	for i := len(outputStr) - 1; i >= 0; i-- {
		if outputStr[i] == '{' {
			jsonStart = i
			break
		}
	}

	if jsonStart == -1 {
		task.Status = "failed"
		task.ErrorMessage = "Failed to parse Python output"
		s.UpdateImportTask(task)
		return fmt.Errorf("no JSON output found")
	}

	var result struct {
		TotalRecords     int `json:"total_records"`
		NewRecords       int `json:"new_records"`
		DuplicateRecords int `json:"duplicate_records"`
	}

	if err := json.Unmarshal([]byte(outputStr[jsonStart:]), &result); err != nil {
		task.Status = "failed"
		task.ErrorMessage = fmt.Sprintf("Failed to parse JSON: %s", err.Error())
		s.UpdateImportTask(task)
		return fmt.Errorf("failed to parse JSON output: %w", err)
	}

	// Update task with results
	now := time.Now()
	task.Status = "completed"
	task.TotalRecords = result.TotalRecords
	task.NewRecords = result.NewRecords
	task.DuplicateRecords = result.DuplicateRecords
	task.CompletedAt = &now

	if err := s.UpdateImportTask(task); err != nil {
		return err
	}

	// Auto-trigger geocoding and analysis if enabled
	if task.AutoTrigger && task.NewRecords > 0 {
		go s.triggerPipeline(task.ID)
	}

	return nil
}

// triggerPipeline triggers geocoding and analysis chain after import
func (s *ImportService) triggerPipeline(importTaskID int64) {
	fmt.Printf("Auto-triggering pipeline for import task %d\n", importTaskID)

	// Step 1: Trigger geocoding
	if s.geocodingService == nil {
		fmt.Printf("Geocoding service not available, skipping auto-trigger\n")
		return
	}

	geocodingTask, err := s.geocodingService.CreateTask(fmt.Sprintf("import_task_%d", importTaskID))
	if err != nil {
		fmt.Printf("Failed to create geocoding task: %v\n", err)
		return
	}

	fmt.Printf("Created geocoding task %d\n", geocodingTask.ID)

	// Step 2: Wait for geocoding to complete, then trigger analysis
	// Poll geocoding task status every 10 seconds
	go s.waitAndTriggerAnalysis(geocodingTask.ID, importTaskID)
}

// waitAndTriggerAnalysis waits for geocoding to complete, then triggers analysis chain
func (s *ImportService) waitAndTriggerAnalysis(geocodingTaskID int, importTaskID int64) {
	fmt.Printf("Waiting for geocoding task %d to complete before triggering analysis\n", geocodingTaskID)

	maxWaitTime := 30 * time.Minute
	pollInterval := 10 * time.Second
	startTime := time.Now()

	for {
		// Check if max wait time exceeded
		if time.Since(startTime) > maxWaitTime {
			fmt.Printf("Geocoding task %d exceeded max wait time, aborting auto-trigger\n", geocodingTaskID)
			return
		}

		// Get geocoding task status
		task, err := s.geocodingService.GetTask(geocodingTaskID)
		if err != nil {
			fmt.Printf("Failed to get geocoding task status: %v\n", err)
			time.Sleep(pollInterval)
			continue
		}

		// Check if completed
		if task.Status == "completed" {
			fmt.Printf("Geocoding task %d completed, triggering analysis chain\n", geocodingTaskID)
			s.triggerAnalysisChain(importTaskID)
			return
		}

		// Check if failed
		if task.Status == "failed" {
			fmt.Printf("Geocoding task %d failed, aborting auto-trigger\n", geocodingTaskID)
			return
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}

// triggerAnalysisChain triggers the complete analysis chain
func (s *ImportService) triggerAnalysisChain(importTaskID int64) {
	if s.analysisTaskService == nil {
		fmt.Printf("Analysis task service not available, skipping analysis chain\n")
		return
	}

	taskIDs, err := s.analysisTaskService.TriggerAnalysisChain("incremental", fmt.Sprintf("import_task_%d", importTaskID))
	if err != nil {
		fmt.Printf("Failed to trigger analysis chain: %v\n", err)
		return
	}

	fmt.Printf("Triggered analysis chain with %d tasks: %v\n", len(taskIDs), taskIDs)
}

// GetUploadPath returns the full path for an uploaded file
func (s *ImportService) GetUploadPath(fileName string) string {
	return filepath.Join(s.uploadDir, fileName)
}

