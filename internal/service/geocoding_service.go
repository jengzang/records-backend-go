package service

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// GeocodingService handles geocoding business logic
type GeocodingService struct {
	repo *repository.GeocodingRepository
}

// NewGeocodingService creates a new geocoding service
func NewGeocodingService(repo *repository.GeocodingRepository) *GeocodingService {
	return &GeocodingService{repo: repo}
}

// CreateTask creates a new geocoding task and starts the Python worker
func (s *GeocodingService) CreateTask(createdBy string) (*models.GeocodingTask, error) {
	// Count ungeocoded points
	count, err := s.repo.CountUngeocodedPoints()
	if err != nil {
		return nil, fmt.Errorf("failed to count ungeocoded points: %w", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("no ungeocoded points found")
	}

	// Create task record
	task := &models.GeocodingTask{
		Status:          models.TaskStatusPending,
		TotalPoints:     count,
		ProcessedPoints: 0,
		FailedPoints:    0,
		CreatedBy:       createdBy,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Start Python Docker container asynchronously
	go s.startGeocodingWorker(task.ID)

	return task, nil
}

// startGeocodingWorker starts the Python geocoding worker in a Docker container
func (s *GeocodingService) startGeocodingWorker(taskID int) {
	log.Printf("Starting geocoding worker for task %d", taskID)

	// Docker run command
	// docker run --rm \
	//   -v /data/tracks:/data \
	//   -v /data/geo:/geo \
	//   records-geocoding:latest \
	//   python /app/geocode_worker.py --task-id <taskID>

	cmd := exec.Command("docker", "run", "--rm",
		"-v", "C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks:/data",
		"-v", "C:/Users/joengzaang/CodeProject/records/go-backend/data/geo:/geo",
		"records-geocoding:latest",
		"python", "/app/geocode_worker.py", "--task-id", strconv.Itoa(taskID))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Geocoding worker failed for task %d: %v\nOutput: %s", taskID, err, string(output))
		// Mark task as failed
		s.repo.MarkAsFailed(taskID, fmt.Sprintf("Worker failed: %v", err))
		return
	}

	log.Printf("Geocoding worker completed for task %d", taskID)
}

// GetTask retrieves a task by ID
func (s *GeocodingService) GetTask(id int) (*models.GeocodingTask, error) {
	return s.repo.GetByID(id)
}

// ListTasks retrieves all tasks with optional status filter
func (s *GeocodingService) ListTasks(status string, limit int, offset int) ([]*models.GeocodingTask, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(status, limit, offset)
}

// CancelTask cancels a running task
func (s *GeocodingService) CancelTask(id int) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if task.IsTerminal() {
		return fmt.Errorf("task is already in terminal state: %s", task.Status)
	}

	// TODO: Implement Docker container stop logic
	// For now, just mark as failed
	return s.repo.MarkAsFailed(id, "Task cancelled by user")
}
