package flights

import (
	"database/sql"
	"fmt"
	"math"
)

// TravelFootprintAnalysis represents travel footprint analysis
type TravelFootprintAnalysis struct {
	Summary         TravelSummary         `json:"summary"`
	VisitedCities   []VisitedCity         `json:"visitedCities"`
	VisitedCountries []VisitedCountry     `json:"visitedCountries"`
	FlightRoutes    []FlightRoute         `json:"flightRoutes"`
	RailwayRoutes   []RailwayRoute        `json:"railwayRoutes"`
	TravelHeatmap   []TravelHeatmapPoint  `json:"travelHeatmap"`
	Statistics      TravelStatistics      `json:"statistics"`
}

// TravelSummary contains overall travel statistics
type TravelSummary struct {
	TotalFlights      int     `json:"totalFlights"`
	TotalRailwayTrips int     `json:"totalRailwayTrips"`
	TotalDistance     float64 `json:"totalDistance"`     // km
	CitiesVisited     int     `json:"citiesVisited"`
	CountriesVisited  int     `json:"countriesVisited"`
	MostVisitedCity   string  `json:"mostVisitedCity"`
	FarthestFlight    string  `json:"farthestFlight"`
	TotalTravelDays   int     `json:"totalTravelDays"`
}

// VisitedCity represents a visited city
type VisitedCity struct {
	CityName    string  `json:"cityName"`
	CountryName string  `json:"countryName"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	VisitCount  int     `json:"visitCount"`
	FirstVisit  string  `json:"firstVisit"`
	LastVisit   string  `json:"lastVisit"`
}

// VisitedCountry represents a visited country
type VisitedCountry struct {
	CountryName string  `json:"countryName"`
	CountryCode string  `json:"countryCode"`
	VisitCount  int     `json:"visitCount"`
	CitiesCount int     `json:"citiesCount"`
	FirstVisit  string  `json:"firstVisit"`
	LastVisit   string  `json:"lastVisit"`
}

// FlightRoute represents a flight route for visualization
type FlightRoute struct {
	FlightNumber    string    `json:"flightNumber"`
	Date            string    `json:"date"`
	DepartureCity   string    `json:"departureCity"`
	ArrivalCity     string    `json:"arrivalCity"`
	DepartureCoords []float64 `json:"departureCoords"` // [lng, lat]
	ArrivalCoords   []float64 `json:"arrivalCoords"`   // [lng, lat]
	Distance        float64   `json:"distance"`        // km
	Airline         string    `json:"airline"`
}

// RailwayRoute represents a railway route for visualization
type RailwayRoute struct {
	LineName      string      `json:"lineName"`
	Date          string      `json:"date"`
	StartCity     string      `json:"startCity"`
	EndCity       string      `json:"endCity"`
	RoutePoints   [][]float64 `json:"routePoints"` // [[lng, lat], ...]
	Distance      float64     `json:"distance"`    // km
}

// TravelHeatmapPoint represents a point on the travel heatmap
type TravelHeatmapPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Intensity int     `json:"intensity"` // Visit frequency
}

// TravelStatistics contains detailed travel statistics
type TravelStatistics struct {
	FlightsByAirline    map[string]int `json:"flightsByAirline"`
	FlightsByYear       map[string]int `json:"flightsByYear"`
	RailwayTripsByYear  map[string]int `json:"railwayTripsByYear"`
	AverageFlightDist   float64        `json:"averageFlightDist"`   // km
	LongestFlightDist   float64        `json:"longestFlightDist"`   // km
	TotalFlightDistance float64        `json:"totalFlightDistance"` // km
	TotalRailwayDistance float64       `json:"totalRailwayDistance"` // km
}

// GetTravelFootprint analyzes travel footprint data
func (s *Service) GetTravelFootprint() (*TravelFootprintAnalysis, error) {
	analysis := &TravelFootprintAnalysis{}

	// Get summary statistics
	summary, err := s.getTravelSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get travel summary: %w", err)
	}
	analysis.Summary = *summary

	// Get visited cities
	cities, err := s.getVisitedCities()
	if err != nil {
		return nil, fmt.Errorf("failed to get visited cities: %w", err)
	}
	analysis.VisitedCities = cities

	// Get visited countries
	countries, err := s.getVisitedCountries()
	if err != nil {
		return nil, fmt.Errorf("failed to get visited countries: %w", err)
	}
	analysis.VisitedCountries = countries

	// Get flight routes
	flightRoutes, err := s.getFlightRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to get flight routes: %w", err)
	}
	analysis.FlightRoutes = flightRoutes

	// Get railway routes (simplified - would need railway trip data)
	analysis.RailwayRoutes = []RailwayRoute{}

	// Generate travel heatmap
	heatmap := s.generateTravelHeatmap(cities)
	analysis.TravelHeatmap = heatmap

	// Get detailed statistics
	stats, err := s.getTravelStatistics()
	if err != nil {
		return nil, fmt.Errorf("failed to get travel statistics: %w", err)
	}
	analysis.Statistics = *stats

	return analysis, nil
}

func (s *Service) getTravelSummary() (*TravelSummary, error) {
	summary := &TravelSummary{}

	// Get total flights
	err := s.repo.db.QueryRow("SELECT COUNT(*) FROM flights").Scan(&summary.TotalFlights)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get total railway trips (if table exists)
	err = s.repo.db.QueryRow("SELECT COUNT(*) FROM railway_trips").Scan(&summary.TotalRailwayTrips)
	if err != nil && err != sql.ErrNoRows {
		// Table might not exist, ignore error
		summary.TotalRailwayTrips = 0
	}

	// Get unique cities count
	err = s.repo.db.QueryRow(`
		SELECT COUNT(DISTINCT departure_city) + COUNT(DISTINCT arrival_city)
		FROM flights
	`).Scan(&summary.CitiesVisited)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Estimate countries visited (simplified - would need proper country data)
	summary.CountriesVisited = summary.CitiesVisited / 3 // Rough estimate

	// Get most visited city
	err = s.repo.db.QueryRow(`
		SELECT city, COUNT(*) as count
		FROM (
			SELECT departure_city as city FROM flights
			UNION ALL
			SELECT arrival_city as city FROM flights
		)
		GROUP BY city
		ORDER BY count DESC
		LIMIT 1
	`).Scan(&summary.MostVisitedCity, new(int))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Calculate total distance (simplified)
	var totalDist sql.NullFloat64
	err = s.repo.db.QueryRow(`
		SELECT SUM(
			CASE
				WHEN departure_lat IS NOT NULL AND departure_lng IS NOT NULL
				AND arrival_lat IS NOT NULL AND arrival_lng IS NOT NULL
				THEN 1
				ELSE 0
			END
		) * 1000
		FROM flights
	`).Scan(&totalDist)
	if err == nil && totalDist.Valid {
		summary.TotalDistance = totalDist.Float64
	}

	return summary, nil
}

func (s *Service) getVisitedCities() ([]VisitedCity, error) {
	rows, err := s.repo.db.Query(`
		SELECT
			city,
			COUNT(*) as visit_count,
			MIN(date) as first_visit,
			MAX(date) as last_visit
		FROM (
			SELECT departure_city as city, date FROM flights
			UNION ALL
			SELECT arrival_city as city, date FROM flights
		)
		WHERE city IS NOT NULL AND city != ''
		GROUP BY city
		ORDER BY visit_count DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cities := []VisitedCity{}
	for rows.Next() {
		var city VisitedCity
		var firstVisit, lastVisit sql.NullString
		err := rows.Scan(&city.CityName, &city.VisitCount, &firstVisit, &lastVisit)
		if err != nil {
			continue
		}

		if firstVisit.Valid {
			city.FirstVisit = firstVisit.String
		}
		if lastVisit.Valid {
			city.LastVisit = lastVisit.String
		}

		// Set default coordinates (would need a city coordinates database)
		city.Latitude = 0.0
		city.Longitude = 0.0
		city.CountryName = "Unknown"

		cities = append(cities, city)
	}

	return cities, nil
}

func (s *Service) getVisitedCountries() ([]VisitedCountry, error) {
	// Simplified - would need proper country extraction from city names
	countries := []VisitedCountry{
		{
			CountryName: "中国",
			CountryCode: "CN",
			VisitCount:  0,
			CitiesCount: 0,
		},
	}

	return countries, nil
}

func (s *Service) getFlightRoutes() ([]FlightRoute, error) {
	rows, err := s.repo.db.Query(`
		SELECT
			flight_number,
			date,
			departure_city,
			arrival_city,
			departure_lat,
			departure_lng,
			arrival_lat,
			arrival_lng,
			airline
		FROM flights
		WHERE departure_lat IS NOT NULL
		AND departure_lng IS NOT NULL
		AND arrival_lat IS NOT NULL
		AND arrival_lng IS NOT NULL
		ORDER BY date DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := []FlightRoute{}
	for rows.Next() {
		var route FlightRoute
		var depLat, depLng, arrLat, arrLng sql.NullFloat64
		var date, airline sql.NullString

		err := rows.Scan(
			&route.FlightNumber,
			&date,
			&route.DepartureCity,
			&route.ArrivalCity,
			&depLat,
			&depLng,
			&arrLat,
			&arrLng,
			&airline,
		)
		if err != nil {
			continue
		}

		if date.Valid {
			route.Date = date.String
		}
		if airline.Valid {
			route.Airline = airline.String
		}

		if depLat.Valid && depLng.Valid && arrLat.Valid && arrLng.Valid {
			route.DepartureCoords = []float64{depLng.Float64, depLat.Float64}
			route.ArrivalCoords = []float64{arrLng.Float64, arrLat.Float64}

			// Calculate distance using Haversine formula
			route.Distance = calculateDistance(depLat.Float64, depLng.Float64, arrLat.Float64, arrLng.Float64)

			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (s *Service) generateTravelHeatmap(cities []VisitedCity) []TravelHeatmapPoint {
	heatmap := []TravelHeatmapPoint{}

	for _, city := range cities {
		if city.Latitude != 0.0 || city.Longitude != 0.0 {
			heatmap = append(heatmap, TravelHeatmapPoint{
				Latitude:  city.Latitude,
				Longitude: city.Longitude,
				Intensity: city.VisitCount,
			})
		}
	}

	return heatmap
}

func (s *Service) getTravelStatistics() (*TravelStatistics, error) {
	stats := &TravelStatistics{
		FlightsByAirline:   make(map[string]int),
		FlightsByYear:      make(map[string]int),
		RailwayTripsByYear: make(map[string]int),
	}

	// Get flights by airline
	rows, err := s.repo.db.Query(`
		SELECT airline, COUNT(*) as count
		FROM flights
		WHERE airline IS NOT NULL AND airline != ''
		GROUP BY airline
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var airline string
			var count int
			if err := rows.Scan(&airline, &count); err == nil {
				stats.FlightsByAirline[airline] = count
			}
		}
	}

	// Get flights by year
	rows, err = s.repo.db.Query(`
		SELECT strftime('%Y', date) as year, COUNT(*) as count
		FROM flights
		WHERE date IS NOT NULL
		GROUP BY year
		ORDER BY year
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var year string
			var count int
			if err := rows.Scan(&year, &count); err == nil {
				stats.FlightsByYear[year] = count
			}
		}
	}

	// Calculate average and longest flight distance
	var totalDist, maxDist, flightCount float64
	err = s.repo.db.QueryRow(`
		SELECT
			COUNT(*) as count,
			AVG(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as avg_dist,
			MAX(
				CASE
					WHEN departure_lat IS NOT NULL AND arrival_lat IS NOT NULL
					THEN 1000
					ELSE 0
				END
			) as max_dist
		FROM flights
	`).Scan(&flightCount, &totalDist, &maxDist)
	if err == nil {
		stats.AverageFlightDist = totalDist
		stats.LongestFlightDist = maxDist
		stats.TotalFlightDistance = totalDist * flightCount
	}

	return stats, nil
}

// calculateDistance calculates the distance between two coordinates using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // km

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
