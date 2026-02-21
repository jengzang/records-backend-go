package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// AnalysisTaskService handles analysis task business logic
type AnalysisTaskService struct {
	repo *repository.AnalysisTaskRepository
	db   *sql.DB
}

// NewAnalysisTaskService creates a new analysis task service
func NewAnalysisTaskService(repo *repository.AnalysisTaskRepository, db *sql.DB) *AnalysisTaskService {
	return &AnalysisTaskService{
		repo: repo,
		db:   db,
	}
}

// CreateTask creates a new analysis task and starts the Python worker
func (s *AnalysisTaskService) CreateTask(skillName string, taskType string, params map[string]interface{}, createdBy string) (*models.AnalysisTask, error) {
	// Validate skill name
	if !isValidSkillName(skillName) {
		return nil, fmt.Errorf("invalid skill name: %s", skillName)
	}

	// Validate task type
	if taskType != models.TaskTypeIncremental && taskType != models.TaskTypeFullRecompute {
		return nil, fmt.Errorf("invalid task type: %s", taskType)
	}

	// Count points to analyze
	var count int
	var err error
	if taskType == models.TaskTypeIncremental {
		count, err = s.repo.CountUnanalyzedPoints()
	} else {
		count, err = s.repo.CountAllPoints()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to count points: %w", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("no points to analyze")
	}

	// Serialize params to JSON
	var paramsJSON *string
	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize params: %w", err)
		}
		jsonStr := string(paramsBytes)
		paramsJSON = &jsonStr
	}

	// Create task record
	task := &models.AnalysisTask{
		SkillName:       skillName,
		TaskType:        taskType,
		Status:          models.TaskStatusPending,
		ProgressPercent: 0,
		TotalPoints:     count,
		ProcessedPoints: 0,
		FailedPoints:    0,
		ParamsJSON:      paramsJSON,
		CreatedBy:       createdBy,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Start analysis worker asynchronously (Go or Python)
	go s.startAnalysisWorker(task.ID, skillName, taskType)

	return task, nil
}

// startAnalysisWorker starts the analysis worker (Go or Python)
func (s *AnalysisTaskService) startAnalysisWorker(taskID int64, skillName string, taskType string) {
	log.Printf("Starting analysis worker for task %d (skill: %s, type: %s)", taskID, skillName, taskType)

	// Check if skill is implemented in Go
	if analysis.IsGoNativeSkill(skillName) {
		// Execute in Go (in-process)
		s.executeGoAnalysis(taskID, skillName, taskType)
	} else {
		// Execute in Python Docker container
		s.executePythonWorker(taskID, skillName, taskType)
	}
}

// executeGoAnalysis executes a Go-native analysis skill
func (s *AnalysisTaskService) executeGoAnalysis(taskID int64, skillName string, taskType string) {
	log.Printf("Executing Go analysis for task %d (skill: %s)", taskID, skillName)

	// Get analyzer instance
	analyzer := analysis.GetAnalyzer(skillName, s.db)
	if analyzer == nil {
		log.Printf("Failed to get analyzer for skill: %s", skillName)
		s.repo.MarkAsFailed(taskID, fmt.Sprintf("Unknown skill: %s", skillName))
		return
	}

	// Execute analysis
	mode := "incremental"
	if taskType == models.TaskTypeFullRecompute {
		mode = "full"
	}

	ctx := context.Background()
	err := analyzer.Analyze(ctx, taskID, mode)
	if err != nil {
		log.Printf("Go analysis failed for task %d: %v", taskID, err)
		s.repo.MarkAsFailed(taskID, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	log.Printf("Go analysis completed for task %d", taskID)
}

// executePythonWorker starts the Python analysis worker in a Docker container
func (s *AnalysisTaskService) executePythonWorker(taskID int64, skillName string, taskType string) {
	log.Printf("Executing Python worker for task %d (skill: %s)", taskID, skillName)

	// Docker run command
	// docker run --rm \
	//   -v /data/tracks:/data \
	//   records-analysis-<skillName>:latest \
	//   python /app/worker.py --task-id <taskID> --mode <taskType>

	imageName := fmt.Sprintf("records-analysis-%s:latest", skillName)
	mode := "incremental"
	if taskType == models.TaskTypeFullRecompute {
		mode = "full"
	}

	cmd := exec.Command("docker", "run", "--rm",
		"-v", "C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks:/data",
		imageName,
		"python", "/app/worker.py",
		"--task-id", strconv.FormatInt(taskID, 10),
		"--mode", mode)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Analysis worker failed for task %d: %v\nOutput: %s", taskID, err, string(output))
		// Mark task as failed
		s.repo.MarkAsFailed(taskID, fmt.Sprintf("Worker failed: %v", err))
		return
	}

	log.Printf("Analysis worker completed for task %d", taskID)
}

// GetTask retrieves a task by ID
func (s *AnalysisTaskService) GetTask(id int64) (*models.AnalysisTask, error) {
	return s.repo.GetByID(id)
}

// ListTasks retrieves all tasks with optional filters
func (s *AnalysisTaskService) ListTasks(skillName string, status string, limit int, offset int) ([]*models.AnalysisTask, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(skillName, status, limit, offset)
}

// CancelTask cancels a running task
func (s *AnalysisTaskService) CancelTask(id int64) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.Status != models.TaskStatusPending && task.Status != models.TaskStatusRunning {
		return fmt.Errorf("task is not running (status: %s)", task.Status)
	}

	// TODO: Implement Docker container stop logic
	// For now, just mark as failed
	return s.repo.MarkAsFailed(id, "Task cancelled by user")
}

// TriggerAnalysisChain triggers a complete analysis chain with dependencies
func (s *AnalysisTaskService) TriggerAnalysisChain(taskType string, createdBy string) ([]int64, error) {
	// Define skill execution order based on dependencies
	skillOrder := []string{
		"outlier_detection",
		"transport_mode",
		"stay_detection",
		"trip_construction",
		"grid_system",
		"footprint_statistics",
		"stay_statistics",
		"rendering_metadata",
	}

	taskIDs := []int64{}

	for _, skillName := range skillOrder {
		task, err := s.CreateTask(skillName, taskType, nil, createdBy)
		if err != nil {
			return taskIDs, fmt.Errorf("failed to create task for %s: %w", skillName, err)
		}
		taskIDs = append(taskIDs, task.ID)

		// Wait for task to complete before starting next one
		// TODO: Implement proper task dependency management
		// For now, tasks will run sequentially
	}

	return taskIDs, nil
}

// isValidSkillName validates if a skill name is supported
func isValidSkillName(skillName string) bool {
	validSkills := map[string]bool{
		"outlier_detection":    true,
		"trajectory_completion": true,
		"transport_mode":       true,
		"stay_detection":       true,
		"trip_construction":    true,
		"streak_detection":     true,
		"speed_events":         true,
		"grid_system":          true,
		"road_overlap":         true,
		"density_structure":    true,
		"speed_space_coupling": true,
		"revisit_pattern":      true,
		"utilization_efficiency": true,
		"spatial_complexity":   true,
		"directional_bias":     true,
		"footprint_statistics": true,
		"stay_statistics":      true,
		"extreme_events":       true,
		"admin_crossings":      true,
		"admin_view_engine":    true,
		"time_space_slicing":   true,
		"time_space_compression": true,
		"altitude_dimension":   true,
		"rendering_metadata":   true,
		"time_axis_map":        true,
		"stay_annotation":      true,
		"spatial_persona":      true,
	}

	return validSkills[skillName]
}
