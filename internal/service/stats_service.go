package service

import (
	"fmt"
	"time"

	"github.com/jengzang/records-backend-go/internal/models"
	"github.com/jengzang/records-backend-go/internal/repository"
)

// StatsService handles business logic for statistics
type StatsService struct {
	statsRepo *repository.StatsRepository
}

// NewStatsService creates a new stats service
func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
	}
}

// GetFootprintStatistics retrieves footprint statistics for a time range
func (s *StatsService) GetFootprintStatistics(startTime, endTime int64) (*models.FootprintStatistics, error) {
	// Validate time range
	if startTime < 0 {
		startTime = 0
	}
	if endTime < 0 {
		endTime = time.Now().Unix()
	}
	if startTime > endTime {
		return nil, fmt.Errorf("start time must be before end time")
	}

	stats, err := s.statsRepo.GetFootprintStatistics(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get footprint statistics: %w", err)
	}

	stats.GeneratedAt = time.Now().Format(time.RFC3339)
	return stats, nil
}

// GetTimeDistribution retrieves time distribution statistics
func (s *StatsService) GetTimeDistribution(startTime, endTime int64) ([]models.TimeDistribution, error) {
	// Validate time range
	if startTime < 0 {
		startTime = 0
	}
	if endTime < 0 {
		endTime = time.Now().Unix()
	}
	if startTime > endTime {
		return nil, fmt.Errorf("start time must be before end time")
	}

	distribution, err := s.statsRepo.GetTimeDistribution(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get time distribution: %w", err)
	}

	return distribution, nil
}

// GetSpeedDistribution retrieves speed distribution statistics
func (s *StatsService) GetSpeedDistribution(startTime, endTime int64) ([]models.SpeedDistribution, error) {
	// Validate time range
	if startTime < 0 {
		startTime = 0
	}
	if endTime < 0 {
		endTime = time.Now().Unix()
	}
	if startTime > endTime {
		return nil, fmt.Errorf("start time must be before end time")
	}

	distribution, err := s.statsRepo.GetSpeedDistribution(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get speed distribution: %w", err)
	}

	return distribution, nil
}

// GetFootprintRankings retrieves footprint statistics with rankings
func (s *StatsService) GetFootprintRankings(filter models.StatsFilter) ([]models.FootprintStatistics, error) {
	return s.statsRepo.GetFootprintRankings(filter)
}

// GetStayRankings retrieves stay statistics with rankings
func (s *StatsService) GetStayRankings(filter models.StatsFilter) ([]models.StayStatistics, error) {
	return s.statsRepo.GetStayRankings(filter)
}

// GetExtremeEvents retrieves extreme events
func (s *StatsService) GetExtremeEvents(eventType, eventCategory string, limit int) ([]models.ExtremeEvent, error) {
	return s.statsRepo.GetExtremeEvents(eventType, eventCategory, limit)
}
