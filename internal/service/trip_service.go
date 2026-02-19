package service

import (
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// TripService handles business logic for trips
type TripService struct {
	repo *repository.TripRepository
}

// NewTripService creates a new trip service
func NewTripService(repo *repository.TripRepository) *TripService {
	return &TripService{repo: repo}
}

// GetTrips retrieves trips with filtering and pagination
func (s *TripService) GetTrips(filter models.TripFilter) ([]models.Trip, int64, error) {
	return s.repo.GetTrips(filter)
}

// GetTripByID retrieves a single trip by ID
func (s *TripService) GetTripByID(id int64) (*models.Trip, error) {
	return s.repo.GetTripByID(id)
}
