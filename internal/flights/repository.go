package flights

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository handles flight data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new flight repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateFlight inserts a new flight record
func (r *Repository) CreateFlight(flight *Flight) (int64, error) {
	query := `
	INSERT INTO flights (
		flight_number, aircraft_number, airline,
		departure_airport, arrival_airport,
		departure_time, arrival_time,
		actual_departure, actual_arrival,
		flight_date, data_source
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(flight_number, flight_date) DO UPDATE SET
		aircraft_number = excluded.aircraft_number,
		airline = excluded.airline,
		departure_airport = excluded.departure_airport,
		arrival_airport = excluded.arrival_airport
	`

	result, err := r.db.Exec(query,
		flight.FlightNumber, flight.AircraftNumber, flight.Airline,
		flight.DepartureAirport, flight.ArrivalAirport,
		flight.DepartureTime, flight.ArrivalTime,
		flight.ActualDeparture, flight.ActualArrival,
		flight.FlightDate, flight.DataSource,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create flight: %w", err)
	}

	return result.LastInsertId()
}

// CreateFlightPoint inserts a flight tracking point
func (r *Repository) CreateFlightPoint(point *FlightPoint) error {
	query := `
	INSERT INTO flight_points (
		flight_id, update_time, utc_time,
		longitude, latitude, altitude, speed, heading
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		point.FlightID, point.UpdateTime, point.UTCTime,
		point.Longitude, point.Latitude, point.Altitude,
		point.Speed, point.Heading,
	)
	if err != nil {
		return fmt.Errorf("failed to create flight point: %w", err)
	}

	return nil
}

// CreateFlightPointsBatch inserts multiple flight points in a transaction
func (r *Repository) CreateFlightPointsBatch(points []FlightPoint) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO flight_points (
			flight_id, update_time, utc_time,
			longitude, latitude, altitude, speed, heading
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, point := range points {
		_, err := stmt.Exec(
			point.FlightID, point.UpdateTime, point.UTCTime,
			point.Longitude, point.Latitude, point.Altitude,
			point.Speed, point.Heading,
		)
		if err != nil {
			return fmt.Errorf("failed to insert point: %w", err)
		}
	}

	return tx.Commit()
}

// GetFlightByID retrieves a flight by ID
func (r *Repository) GetFlightByID(id int) (*Flight, error) {
	query := `
	SELECT id, flight_number, aircraft_number, airline,
		departure_airport, arrival_airport,
		departure_time, arrival_time,
		actual_departure, actual_arrival,
		flight_date, data_source, created_at
	FROM flights
	WHERE id = ?
	`

	var flight Flight
	err := r.db.QueryRow(query, id).Scan(
		&flight.ID, &flight.FlightNumber, &flight.AircraftNumber, &flight.Airline,
		&flight.DepartureAirport, &flight.ArrivalAirport,
		&flight.DepartureTime, &flight.ArrivalTime,
		&flight.ActualDeparture, &flight.ActualArrival,
		&flight.FlightDate, &flight.DataSource, &flight.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight: %w", err)
	}

	return &flight, nil
}

// ListFlights retrieves all flights with optional filters
func (r *Repository) ListFlights(limit, offset int) ([]FlightWithStats, error) {
	query := `
	SELECT
		f.id, f.flight_number, f.aircraft_number, f.airline,
		f.departure_airport, f.arrival_airport,
		f.departure_time, f.arrival_time,
		f.actual_departure, f.actual_arrival,
		f.flight_date, f.data_source, f.created_at,
		fs.total_distance, fs.max_altitude, fs.max_speed,
		fs.avg_speed, fs.duration_minutes, fs.point_count
	FROM flights f
	LEFT JOIN flight_statistics fs ON f.id = fs.flight_id
	ORDER BY f.flight_date DESC, f.flight_number
	LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list flights: %w", err)
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
			f.Statistics = &stats
		}
		flights = append(flights, f)
	}

	return flights, rows.Err()
}

// GetFlightPoints retrieves all tracking points for a flight
func (r *Repository) GetFlightPoints(flightID int) ([]FlightPoint, error) {
	query := `
	SELECT id, flight_id, update_time, utc_time,
		longitude, latitude, altitude, speed, heading, created_at
	FROM flight_points
	WHERE flight_id = ?
	ORDER BY update_time ASC
	`

	rows, err := r.db.Query(query, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight points: %w", err)
	}
	defer rows.Close()

	var points []FlightPoint
	for rows.Next() {
		var p FlightPoint
		err := rows.Scan(
			&p.ID, &p.FlightID, &p.UpdateTime, &p.UTCTime,
			&p.Longitude, &p.Latitude, &p.Altitude,
			&p.Speed, &p.Heading, &p.CreatedAt,
		)
		if err != nil {
			continue
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// GetFlightSummary retrieves summary statistics for all flights
func (r *Repository) GetFlightSummary() (*FlightSummary, error) {
	query := `
	SELECT
		COUNT(*) as total_flights,
		COALESCE(SUM(fs.total_distance), 0) as total_distance,
		COALESCE(SUM(fs.duration_minutes), 0) as total_duration,
		COALESCE(AVG(fs.total_distance), 0) as avg_distance,
		COALESCE(AVG(fs.duration_minutes), 0) as avg_duration
	FROM flights f
	LEFT JOIN flight_statistics fs ON f.id = fs.flight_id
	`

	var summary FlightSummary
	err := r.db.QueryRow(query).Scan(
		&summary.TotalFlights,
		&summary.TotalDistance,
		&summary.TotalDuration,
		&summary.AverageDistance,
		&summary.AverageDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight summary: %w", err)
	}

	// Get unique airlines
	airlinesQuery := `SELECT DISTINCT airline FROM flights WHERE airline != '' ORDER BY airline`
	rows, err := r.db.Query(airlinesQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var airline string
			if rows.Scan(&airline) == nil {
				summary.Airlines = append(summary.Airlines, airline)
			}
		}
	}

	return &summary, nil
}

// UpdateFlightStatistics updates or creates flight statistics
func (r *Repository) UpdateFlightStatistics(stats *FlightStatistics) error {
	query := `
	INSERT INTO flight_statistics (
		flight_id, total_distance, max_altitude, max_speed,
		avg_speed, duration_minutes, point_count, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(flight_id) DO UPDATE SET
		total_distance = excluded.total_distance,
		max_altitude = excluded.max_altitude,
		max_speed = excluded.max_speed,
		avg_speed = excluded.avg_speed,
		duration_minutes = excluded.duration_minutes,
		point_count = excluded.point_count,
		updated_at = excluded.updated_at
	`

	_, err := r.db.Exec(query,
		stats.FlightID, stats.TotalDistance, stats.MaxAltitude,
		stats.MaxSpeed, stats.AvgSpeed, stats.DurationMinutes,
		stats.PointCount, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update flight statistics: %w", err)
	}

	return nil
}
