package service

import (
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// SegmentService handles business logic for segments
type SegmentService struct {
	repo *repository.SegmentRepository
}

// NewSegmentService creates a new segment service
func NewSegmentService(repo *repository.SegmentRepository) *SegmentService {
	return &SegmentService{repo: repo}
}

// GetSegments retrieves segments with filtering and pagination
func (s *SegmentService) GetSegments(filter models.SegmentFilter) ([]models.Segment, int64, error) {
	return s.repo.GetSegments(filter)
}

// GetSegmentByID retrieves a single segment by ID
func (s *SegmentService) GetSegmentByID(id int64) (*models.Segment, error) {
	return s.repo.GetSegmentByID(id)
}
