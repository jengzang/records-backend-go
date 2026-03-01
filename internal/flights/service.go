package flights

import (
	"database/sql"
	"fmt"
	"math"
)

// Service handles flight business logic
type Service struct {
	repo *Repository
}

// NewService creates a new flight service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ImportFlight imports a flight with its tracking points
func (s *Service) ImportFlight(flight *Flight, points []FlightPoint) error {
	// Create flight record
	flightID, err := s.repo.CreateFlight(flight)
	if err != nil {
		return fmt.Errorf("failed to import flight: %w", err)
	}

	// Set flight ID for all points
	for i := range points {
		points[i].FlightID = int(flightID)
	}

	// Insert points in batch
	if len(points) > 0 {
		if err := s.repo.CreateFlightPointsBatch(points); err != nil {
			return fmt.Errorf("failed to import flight points: %w", err)
		}
	}

	// Calculate and store statistics
	if len(points) > 0 {
		stats := s.calculateStatistics(int(flightID), points)
		if err := s.repo.UpdateFlightStatistics(stats); err != nil {
			return fmt.Errorf("failed to update statistics: %w", err)
		}
	}

	return nil
}

// GetFlight retrieves a flight with its points and statistics
func (s *Service) GetFlight(id int) (*FlightWithStats, []FlightPoint, error) {
	flight, err := s.repo.GetFlightByID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get flight: %w", err)
	}

	points, err := s.repo.GetFlightPoints(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get flight points: %w", err)
	}

	flightWithStats := &FlightWithStats{
		Flight:     *flight,
		PointCount: len(points),
	}

	return flightWithStats, points, nil
}

// ListFlights retrieves all flights with pagination
func (s *Service) ListFlights(page, pageSize int) ([]FlightWithStats, int, error) {
	offset := (page - 1) * pageSize
	flights, err := s.repo.ListFlights(pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list flights: %w", err)
	}

	// Get total count (simplified - in production, use a separate count query)
	total := len(flights)
	if len(flights) == pageSize {
		total = page * pageSize // Estimate
	}

	return flights, total, nil
}

// GetSummary retrieves flight summary statistics
func (s *Service) GetSummary() (*FlightSummary, error) {
	return s.repo.GetFlightSummary()
}

// SearchFlights searches flights with filters
func (s *Service) SearchFlights(filters SearchFilters) ([]FlightWithStats, int, error) {
	// Set default limit if not specified
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return s.repo.SearchFlights(filters)
}

// GetAirlines returns all unique airlines
func (s *Service) GetAirlines() ([]string, error) {
	return s.repo.GetAirlines()
}

// GetDateRange returns the date range of all flights
func (s *Service) GetDateRange() (string, string, error) {
	return s.repo.GetDateRange()
}

// GetAirlineStatistics returns statistics grouped by airline
func (s *Service) GetAirlineStatistics() (map[string]AirlineStats, error) {
	return s.repo.GetFlightsByAirline()
}

// calculateStatistics computes statistics from flight points
func (s *Service) calculateStatistics(flightID int, points []FlightPoint) *FlightStatistics {
	if len(points) == 0 {
		return &FlightStatistics{FlightID: flightID}
	}

	stats := &FlightStatistics{
		FlightID:   flightID,
		PointCount: len(points),
	}

	var totalDistance float64
	var totalSpeed float64
	var speedCount int

	for i, point := range points {
		// Track max altitude
		if point.Altitude > stats.MaxAltitude {
			stats.MaxAltitude = point.Altitude
		}

		// Track max speed
		if point.Speed > stats.MaxSpeed {
			stats.MaxSpeed = point.Speed
		}

		// Calculate average speed
		if point.Speed > 0 {
			totalSpeed += point.Speed
			speedCount++
		}

		// Calculate distance between consecutive points
		if i > 0 {
			dist := haversineDistance(
				points[i-1].Latitude, points[i-1].Longitude,
				point.Latitude, point.Longitude,
			)
			totalDistance += dist
		}
	}

	stats.TotalDistance = totalDistance

	if speedCount > 0 {
		stats.AvgSpeed = totalSpeed / float64(speedCount)
	}

	// Calculate duration from first to last point
	if len(points) > 1 {
		durationSeconds := points[len(points)-1].UpdateTime - points[0].UpdateTime
		stats.DurationMinutes = int(durationSeconds / 60)
	}

	return stats
}

// haversineDistance calculates the distance between two GPS coordinates in kilometers
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // km

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// MatchFlightToTracks attempts to match flight points with GPS track points
func (s *Service) MatchFlightToTracks(flightID int, tracksDB *sql.DB) ([]FlightTrackMatch, error) {
	// Get flight points
	points, err := s.repo.GetFlightPoints(flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight points: %w", err)
	}

	var matches []FlightTrackMatch

	// For each flight point, find nearby track points
	for _, point := range points {
		// Query tracks database for points within time window and spatial proximity
		query := `
		SELECT id, dataTime, longitude, latitude
		FROM "一生足迹"
		WHERE dataTime BETWEEN ? AND ?
		AND longitude BETWEEN ? AND ?
		AND latitude BETWEEN ? AND ?
		LIMIT 10
		`

		// Time window: ±5 minutes
		timeWindow := int64(300)
		startTime := point.UpdateTime - timeWindow
		endTime := point.UpdateTime + timeWindow

		// Spatial window: ±0.1 degrees (~11km)
		spatialWindow := 0.1

		rows, err := tracksDB.Query(query,
			startTime, endTime,
			point.Longitude-spatialWindow, point.Longitude+spatialWindow,
			point.Latitude-spatialWindow, point.Latitude+spatialWindow,
		)
		if err != nil {
			continue
		}

		// Find best match
		var bestMatch *FlightTrackMatch
		minDistance := math.MaxFloat64

		for rows.Next() {
			var trackID int
			var trackTime int64
			var trackLon, trackLat float64

			if err := rows.Scan(&trackID, &trackTime, &trackLon, &trackLat); err != nil {
				continue
			}

			// Calculate distance
			distance := haversineDistance(point.Latitude, point.Longitude, trackLat, trackLon)
			timeDiff := int(math.Abs(float64(point.UpdateTime - trackTime)))

			// Calculate confidence score (0.0-1.0)
			// Based on spatial and temporal proximity
			spatialScore := 1.0 - math.Min(distance/10.0, 1.0) // 10km max
			temporalScore := 1.0 - math.Min(float64(timeDiff)/300.0, 1.0) // 5min max
			confidence := (spatialScore + temporalScore) / 2.0

			if distance < minDistance && confidence > 0.5 {
				minDistance = distance
				matchType := "estimated"
				if distance < 0.1 && timeDiff < 60 {
					matchType = "exact"
				} else if distance < 1.0 && timeDiff < 180 {
					matchType = "interpolated"
				}

				bestMatch = &FlightTrackMatch{
					FlightID:        flightID,
					TrackPointID:    trackID,
					MatchConfidence: confidence,
					TimeDiffSeconds: timeDiff,
					DistanceMeters:  distance * 1000,
					MatchType:       matchType,
				}
			}
		}
		rows.Close()

		if bestMatch != nil {
			matches = append(matches, *bestMatch)
		}
	}

	return matches, nil
}
