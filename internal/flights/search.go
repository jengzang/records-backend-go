package flights

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchFilters represents search and filter criteria for flights
type SearchFilters struct {
	FlightNumber string // Partial match
	Airline      string // Exact match
	DateFrom     string // YYYYMMDD format
	DateTo       string // YYYYMMDD format
	MinDistance  float64
	MaxDistance  float64
	SortBy       string // flight_date, distance, duration
	SortOrder    string // asc, desc
	Limit        int
	Offset       int
}

// SearchFlights searches flights with filters
func (r *Repository) SearchFlights(filters SearchFilters) ([]FlightWithStats, int, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}

	if filters.FlightNumber != "" {
		conditions = append(conditions, "f.flight_number LIKE ?")
		args = append(args, "%"+filters.FlightNumber+"%")
	}

	if filters.Airline != "" {
		conditions = append(conditions, "f.airline = ?")
		args = append(args, filters.Airline)
	}

	if filters.DateFrom != "" {
		conditions = append(conditions, "f.flight_date >= ?")
		args = append(args, filters.DateFrom)
	}

	if filters.DateTo != "" {
		conditions = append(conditions, "f.flight_date <= ?")
		args = append(args, filters.DateTo)
	}

	if filters.MinDistance > 0 {
		conditions = append(conditions, "fs.total_distance >= ?")
		args = append(args, filters.MinDistance)
	}

	if filters.MaxDistance > 0 {
		conditions = append(conditions, "fs.total_distance <= ?")
		args = append(args, filters.MaxDistance)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "f.flight_date DESC, f.flight_number"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "flight_date":
			orderBy = "f.flight_date"
		case "distance":
			orderBy = "fs.total_distance"
		case "duration":
			orderBy = "fs.duration_minutes"
		case "airline":
			orderBy = "f.airline"
		}

		if filters.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM flights f
		LEFT JOIN flight_statistics fs ON f.id = fs.flight_id
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count flights: %w", err)
	}

	// Get flights
	query := fmt.Sprintf(`
		SELECT
			f.id, f.flight_number, f.aircraft_number, f.airline,
			f.departure_airport, f.arrival_airport,
			f.departure_time, f.arrival_time,
			f.actual_departure, f.actual_arrival,
			f.flight_date, f.data_source, f.created_at,
			COALESCE(fs.total_distance, 0) as total_distance,
			COALESCE(fs.max_altitude, 0) as max_altitude,
			COALESCE(fs.max_speed, 0) as max_speed,
			COALESCE(fs.avg_speed, 0) as avg_speed,
			COALESCE(fs.duration_minutes, 0) as duration_minutes,
			COALESCE(fs.point_count, 0) as point_count
		FROM flights f
		LEFT JOIN flight_statistics fs ON f.id = fs.flight_id
		%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search flights: %w", err)
	}
	defer rows.Close()

	var flights []FlightWithStats
	for rows.Next() {
		var f FlightWithStats
		var stats FlightStatistics

		err := rows.Scan(
			&f.ID, &f.FlightNumber, &f.AircraftNumber, &f.Airline,
			&f.DepartureAirport, &f.ArrivalAirport,
			&f.DepartureTime, &f.ArrivalTime,
			&f.ActualDeparture, &f.ActualArrival,
			&f.FlightDate, &f.DataSource, &f.CreatedAt,
			&stats.TotalDistance, &stats.MaxAltitude, &stats.MaxSpeed,
			&stats.AvgSpeed, &stats.DurationMinutes, &stats.PointCount,
		)
		if err != nil {
			continue
		}

		if stats.PointCount > 0 {
			stats.FlightID = f.ID
			f.Statistics = &stats
		}
		f.PointCount = stats.PointCount

		flights = append(flights, f)
	}

	return flights, total, rows.Err()
}

// GetAirlines returns a list of all unique airlines
func (r *Repository) GetAirlines() ([]string, error) {
	query := `
		SELECT DISTINCT airline
		FROM flights
		WHERE airline != '' AND airline IS NOT NULL
		ORDER BY airline
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get airlines: %w", err)
	}
	defer rows.Close()

	var airlines []string
	for rows.Next() {
		var airline string
		if err := rows.Scan(&airline); err != nil {
			continue
		}
		airlines = append(airlines, airline)
	}

	return airlines, rows.Err()
}

// GetDateRange returns the earliest and latest flight dates
func (r *Repository) GetDateRange() (string, string, error) {
	query := `
		SELECT MIN(flight_date), MAX(flight_date)
		FROM flights
	`

	var minDate, maxDate sql.NullString
	err := r.db.QueryRow(query).Scan(&minDate, &maxDate)
	if err != nil {
		return "", "", fmt.Errorf("failed to get date range: %w", err)
	}

	return minDate.String, maxDate.String, nil
}

// GetFlightsByAirline returns flights grouped by airline with statistics
func (r *Repository) GetFlightsByAirline() (map[string]AirlineStats, error) {
	query := `
		SELECT
			f.airline,
			COUNT(*) as flight_count,
			COALESCE(SUM(fs.total_distance), 0) as total_distance,
			COALESCE(SUM(fs.duration_minutes), 0) as total_duration,
			COALESCE(AVG(fs.total_distance), 0) as avg_distance,
			COALESCE(AVG(fs.duration_minutes), 0) as avg_duration
		FROM flights f
		LEFT JOIN flight_statistics fs ON f.id = fs.flight_id
		WHERE f.airline != '' AND f.airline IS NOT NULL
		GROUP BY f.airline
		ORDER BY flight_count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get flights by airline: %w", err)
	}
	defer rows.Close()

	result := make(map[string]AirlineStats)
	for rows.Next() {
		var airline string
		var stats AirlineStats

		err := rows.Scan(
			&airline,
			&stats.FlightCount,
			&stats.TotalDistance,
			&stats.TotalDuration,
			&stats.AvgDistance,
			&stats.AvgDuration,
		)
		if err != nil {
			continue
		}

		result[airline] = stats
	}

	return result, rows.Err()
}

// AirlineStats represents statistics for an airline
type AirlineStats struct {
	FlightCount   int     `json:"flightCount"`
	TotalDistance float64 `json:"totalDistance"` // km
	TotalDuration int     `json:"totalDuration"` // minutes
	AvgDistance   float64 `json:"avgDistance"`   // km
	AvgDuration   int     `json:"avgDuration"`   // minutes
}
