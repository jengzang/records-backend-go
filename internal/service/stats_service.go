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

// GetAdminCrossings retrieves administrative boundary crossing events
func (s *StatsService) GetAdminCrossings(crossingType, fromRegion, toRegion string, startTime, endTime int64, limit int) ([]models.AdminCrossing, error) {
	// Validate time range
	if startTime < 0 {
		startTime = 0
	}
	if endTime < 0 {
		endTime = time.Now().Unix()
	}
	if startTime > 0 && endTime > 0 && startTime > endTime {
		return nil, fmt.Errorf("start time must be before end time")
	}

	return s.statsRepo.GetAdminCrossings(crossingType, fromRegion, toRegion, startTime, endTime, limit)
}

// GetAdminStats retrieves administrative region statistics
func (s *StatsService) GetAdminStats(adminLevel, adminName, parentName, sortBy string, limit int) ([]models.AdminStats, error) {
	return s.statsRepo.GetAdminStats(adminLevel, adminName, parentName, sortBy, limit)
}
// GetSpeedSpaceStats retrieves speed-space coupling statistics
func (s *StatsService) GetSpeedSpaceStats(bucketType, areaType, areaName string, limit int) ([]models.SpeedSpaceStats, error) {
	return s.statsRepo.GetSpeedSpaceStats(bucketType, areaType, areaName, limit)
}

// GetHighSpeedZones retrieves high-speed zones
func (s *StatsService) GetHighSpeedZones(bucketType, areaType string, limit int) ([]models.SpeedSpaceStats, error) {
	return s.statsRepo.GetHighSpeedZones(bucketType, areaType, limit)
}

// GetSlowLifeZones retrieves slow-life zones
func (s *StatsService) GetSlowLifeZones(bucketType, areaType string, limit int) ([]models.SpeedSpaceStats, error) {
	return s.statsRepo.GetSlowLifeZones(bucketType, areaType, limit)
}

// GetDirectionalBiasStats retrieves directional bias statistics
func (s *StatsService) GetDirectionalBiasStats(bucketType, areaType, areaKey, modeFilter string, limit int) ([]models.DirectionalBiasStats, error) {
	return s.statsRepo.GetDirectionalBiasStats(bucketType, areaType, areaKey, modeFilter, limit)
}

// GetTopDirectionalAreas retrieves areas with highest directional concentration
func (s *StatsService) GetTopDirectionalAreas(bucketType string, limit int) ([]models.DirectionalBiasStats, error) {
	return s.statsRepo.GetTopDirectionalAreas(bucketType, limit)
}

// GetBidirectionalPatterns retrieves areas with strong bidirectional patterns
func (s *StatsService) GetBidirectionalPatterns(bucketType string, limit int) ([]models.DirectionalBiasStats, error) {
	return s.statsRepo.GetBidirectionalPatterns(bucketType, limit)
}

// GetRevisitPatterns retrieves revisit patterns with filters
func (s *StatsService) GetRevisitPatterns(minVisits int, habitualOnly, periodicOnly bool, limit int) ([]models.RevisitPattern, error) {
	return s.statsRepo.GetRevisitPatterns(minVisits, habitualOnly, periodicOnly, limit)
}

// GetTopRevisitLocations retrieves locations with highest revisit strength
func (s *StatsService) GetTopRevisitLocations(limit int) ([]models.RevisitPattern, error) {
	return s.statsRepo.GetTopRevisitLocations(limit)
}

// GetHabitualLocations retrieves habitual locations
func (s *StatsService) GetHabitualLocations(limit int) ([]models.RevisitPattern, error) {
	return s.statsRepo.GetHabitualLocations(limit)
}

// GetPeriodicLocations retrieves locations with periodic visit patterns
func (s *StatsService) GetPeriodicLocations(limit int) ([]models.RevisitPattern, error) {
	return s.statsRepo.GetPeriodicLocations(limit)
}

// GetSpatialUtilization retrieves utilization stats with filters
func (s *StatsService) GetSpatialUtilization(
	bucketType string,
	areaType string,
	areaKey string,
	limit int,
) ([]models.SpatialUtilization, error) {
	return s.statsRepo.GetSpatialUtilization(bucketType, areaType, areaKey, limit)
}

// GetDestinationAreas retrieves areas with high utilization efficiency
func (s *StatsService) GetDestinationAreas(
	bucketType string,
	areaType string,
	limit int,
) ([]models.SpatialUtilization, error) {
	return s.statsRepo.GetDestinationAreas(bucketType, areaType, limit)
}

// GetTransitCorridors retrieves areas with high transit dominance
func (s *StatsService) GetTransitCorridors(
	bucketType string,
	areaType string,
	limit int,
) ([]models.SpatialUtilization, error) {
	return s.statsRepo.GetTransitCorridors(bucketType, areaType, limit)
}

// GetDeepEngagementAreas retrieves areas with high area depth
func (s *StatsService) GetDeepEngagementAreas(
	bucketType string,
	areaType string,
	limit int,
) ([]models.SpatialUtilization, error) {
	return s.statsRepo.GetDeepEngagementAreas(bucketType, areaType, limit)
}

// GetDensityGrids retrieves density grids with filters
func (s *StatsService) GetDensityGrids(
	bucketType string,
	densityLevel string,
	limit int,
) ([]models.SpatialDensityGrid, error) {
	return s.statsRepo.GetDensityGrids(bucketType, densityLevel, limit)
}

// GetCoreAreas retrieves core density areas
func (s *StatsService) GetCoreAreas(
	bucketType string,
	limit int,
) ([]models.SpatialDensityGrid, error) {
	return s.statsRepo.GetCoreAreas(bucketType, limit)
}

// GetRareVisits retrieves rare visit locations
func (s *StatsService) GetRareVisits(
	bucketType string,
	limit int,
) ([]models.SpatialDensityGrid, error) {
	return s.statsRepo.GetRareVisits(bucketType, limit)
}

// GetDensityClusters retrieves density clusters
func (s *StatsService) GetDensityClusters(
	bucketType string,
	limit int,
) ([]models.SpatialDensityGrid, error) {
	return s.statsRepo.GetDensityClusters(bucketType, limit)
}

// GetAltitudeStats retrieves altitude statistics with filters
func (s *StatsService) GetAltitudeStats(
	bucketType string,
	areaType string,
	areaKey string,
	limit int,
) ([]models.AltitudeStats, error) {
	return s.statsRepo.GetAltitudeStats(bucketType, areaType, areaKey, limit)
}

// GetHighestAltitudeSpans retrieves areas with highest altitude spans
func (s *StatsService) GetHighestAltitudeSpans(
	bucketType string,
	limit int,
) ([]models.AltitudeStats, error) {
	return s.statsRepo.GetHighestAltitudeSpans(bucketType, limit)
}

// GetHighestVerticalIntensity retrieves areas with highest vertical intensity
func (s *StatsService) GetHighestVerticalIntensity(
	bucketType string,
	limit int,
) ([]models.AltitudeStats, error) {
	return s.statsRepo.GetHighestVerticalIntensity(bucketType, limit)
}

// GetTimeSpaceCompression retrieves time-space compression stats with filters
func (s *StatsService) GetTimeSpaceCompression(
	bucketType string,
	areaType string,
	areaKey string,
	limit int,
) ([]models.TimeSpaceCompression, error) {
	return s.statsRepo.GetTimeSpaceCompression(bucketType, areaType, areaKey, limit)
}

// GetHighestMovementIntensity retrieves areas with highest movement intensity
func (s *StatsService) GetHighestMovementIntensity(
	bucketType string,
	limit int,
) ([]models.TimeSpaceCompression, error) {
	return s.statsRepo.GetHighestMovementIntensity(bucketType, limit)
}

// GetBurstPeriods retrieves areas with most burst periods
func (s *StatsService) GetBurstPeriods(
	bucketType string,
	limit int,
) ([]models.TimeSpaceCompression, error) {
	return s.statsRepo.GetBurstPeriods(bucketType, limit)
}

// GetTimeSpaceSlices retrieves time-space slices with filters
func (s *StatsService) GetTimeSpaceSlices(
	sliceType string,
	limit int,
) ([]models.TimeSpaceSlice, error) {
	return s.statsRepo.GetTimeSpaceSlices(sliceType, limit)
}

// GetWeeklyPattern retrieves weekly-hourly pattern
func (s *StatsService) GetWeeklyPattern() ([]models.TimeSpaceSlice, error) {
	return s.statsRepo.GetWeeklyPattern()
}

// GetHourlyPattern retrieves hourly pattern
func (s *StatsService) GetHourlyPattern() ([]models.TimeSpaceSlice, error) {
	return s.statsRepo.GetHourlyPattern()
}
