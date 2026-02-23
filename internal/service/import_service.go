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
	db              *sql.DB
	uploadDir       string
	pythonScriptPath string
}

func NewImportService(db *sql.DB) *ImportService {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "/tmp/uploads"
	}

	// Ensure upload directory exists
	os.MkdirAll(uploadDir, 0755)

	return &ImportService{
		db:              db,
		uploadDir:       uploadDir,
		pythonScriptPath: "scripts/tracks/import/write2sql.py",
	}
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

	return nil
}

// GetUploadPath returns the full path for an uploaded file
func (s *ImportService) GetUploadPath(fileName string) string {
	return filepath.Join(s.uploadDir, fileName)
}

