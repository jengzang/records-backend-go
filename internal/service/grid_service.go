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

// GetHeatmapData retrieves heatmap data with normalized intensity scores
func (s *GridService) GetHeatmapData(filter models.GridFilter, metric string) (*models.HeatmapResponse, error) {
	// 1. Get grid cells using existing method
	cells, err := s.repo.GetGridCells(filter)
	if err != nil {
		return nil, err
	}

	if len(cells) == 0 {
		return &models.HeatmapResponse{
			Points:    []models.HeatmapPoint{},
			Count:     0,
			MaxValue:  0,
			MinValue:  0,
			Metric:    metric,
			GridLevel: filter.Level,
		}, nil
	}

	// 2. Extract metric values and find min/max
	points := make([]models.HeatmapPoint, 0, len(cells))
	maxValue := 0
	minValue := int(^uint(0) >> 1) // Max int value

	for _, cell := range cells {
		var value int
		switch metric {
		case "point_count":
			value = cell.PointCount
		case "duration":
			value = int(cell.TotalDurationSeconds)
		case "visit_count":
			value = cell.VisitCount
		default:
			value = cell.PointCount
		}

		if value > maxValue {
			maxValue = value
		}
		if value < minValue {
			minValue = value
		}

		points = append(points, models.HeatmapPoint{
			Lat:    cell.CenterLat,
			Lng:    cell.CenterLon,
			Value:  value,
			Metric: metric,
		})
	}

	// 3. Normalize intensity scores (0-1)
	valueRange := float64(maxValue - minValue)
	if valueRange > 0 {
		for i := range points {
			points[i].Intensity = float64(points[i].Value-minValue) / valueRange
		}
	} else {
		// All values are the same, set intensity to 1.0
		for i := range points {
			points[i].Intensity = 1.0
		}
	}

	return &models.HeatmapResponse{
		Points:    points,
		Count:     len(points),
		MaxValue:  maxValue,
		MinValue:  minValue,
		Metric:    metric,
		GridLevel: filter.Level,
	}, nil
}
