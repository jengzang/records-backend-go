package service

import (
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// StayService handles business logic for stay segments
type StayService struct {
	repo *repository.StayRepository
}

// NewStayService creates a new stay service
func NewStayService(repo *repository.StayRepository) *StayService {
	return &StayService{repo: repo}
}

// GetStays retrieves stay segments with filtering and pagination
func (s *StayService) GetStays(filter models.StayFilter) ([]models.StaySegment, int64, error) {
	return s.repo.GetStays(filter)
}

// GetStayByID retrieves a single stay segment by ID
func (s *StayService) GetStayByID(id int64) (*models.StaySegment, error) {
	return s.repo.GetStayByID(id)
}
