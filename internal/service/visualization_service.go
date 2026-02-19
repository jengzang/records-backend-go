package service

import (
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// VisualizationService handles business logic for visualization data
type VisualizationService struct {
	repo *repository.VisualizationRepository
}

// NewVisualizationService creates a new visualization service
func NewVisualizationService(repo *repository.VisualizationRepository) *VisualizationService {
	return &VisualizationService{repo: repo}
}

// GetRenderingMetadata retrieves track points with rendering properties
func (s *VisualizationService) GetRenderingMetadata(filter models.RenderFilter) ([]models.TrackPoint, error) {
	return s.repo.GetRenderingMetadata(filter)
}

// GetTimeSliceData retrieves aggregated data for time axis
func (s *VisualizationService) GetTimeSliceData(startTime, endTime int64, granularity string) (map[string]interface{}, error) {
	return s.repo.GetTimeSliceData(startTime, endTime, granularity)
}
