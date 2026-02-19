package service

import (
	"fmt"
	"math"

	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// TrackService handles business logic for track points
type TrackService struct {
	trackRepo *repository.TrackRepository
}

// NewTrackService creates a new track service
func NewTrackService(trackRepo *repository.TrackRepository) *TrackService {
	return &TrackService{
		trackRepo: trackRepo,
	}
}

// GetTrackPoints retrieves track points with filtering and pagination
func (s *TrackService) GetTrackPoints(filter models.TrackPointFilter) (*models.TrackPointsResponse, error) {
	// Validate filter
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 100
	}
	if filter.PageSize > 1000 {
		filter.PageSize = 1000
	}

	// Get track points from repository
	points, total, err := s.trackRepo.GetTrackPoints(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get track points: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	return &models.TrackPointsResponse{
		Data:       points,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTrackPointByID retrieves a single track point by ID
func (s *TrackService) GetTrackPointByID(id int64) (*models.TrackPoint, error) {
	point, err := s.trackRepo.GetTrackPointByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get track point: %w", err)
	}
	if point == nil {
		return nil, fmt.Errorf("track point not found")
	}
	return point, nil
}

// GetUngeocodedPoints retrieves track points without administrative divisions
func (s *TrackService) GetUngeocodedPoints(limit int) ([]models.TrackPoint, error) {
	if limit < 1 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	points, err := s.trackRepo.GetUngeocodedPoints(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get ungeocoded points: %w", err)
	}

	return points, nil
}
