package service

import (
	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// GridService handles business logic for grid cells
type GridService struct {
	repo *repository.GridRepository
}

// NewGridService creates a new grid service
func NewGridService(repo *repository.GridRepository) *GridService {
	return &GridService{repo: repo}
}

// GetGridCells retrieves grid cells with filtering
func (s *GridService) GetGridCells(filter models.GridFilter) ([]models.GridCell, error) {
	return s.repo.GetGridCells(filter)
}

// GetGridCellByID retrieves a single grid cell by ID
func (s *GridService) GetGridCellByID(id int64) (*models.GridCell, error) {
	return s.repo.GetGridCellByID(id)
}
