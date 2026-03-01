package railway

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Railway Lines

func (r *Repository) CreateLine(line *RailwayLine) error {
	query := `
		INSERT INTO railway_lines (line_name, line_code, line_type, total_distance,
			start_station, end_station, opened_date, max_speed, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, line.LineName, line.LineCode, line.LineType,
		line.TotalDistance, line.StartStation, line.EndStation, line.OpenedDate,
		line.MaxSpeed, line.Source)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	line.ID = int(id)
	return nil
}

func (r *Repository) GetLineByID(id int) (*RailwayLine, error) {
	query := `
		SELECT id, line_name, line_code, line_type, total_distance, start_station,
			end_station, opened_date, max_speed, source, created_at, updated_at
		FROM railway_lines WHERE id = ?
	`
	line := &RailwayLine{}
	err := r.db.QueryRow(query, id).Scan(
		&line.ID, &line.LineName, &line.LineCode, &line.LineType, &line.TotalDistance,
		&line.StartStation, &line.EndStation, &line.OpenedDate, &line.MaxSpeed,
		&line.Source, &line.CreatedAt, &line.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return line, nil
}

func (r *Repository) GetAllLines() ([]RailwayLine, error) {
	query := `
		SELECT id, line_name, line_code, line_type, total_distance, start_station,
			end_station, opened_date, max_speed, source, created_at, updated_at
		FROM railway_lines ORDER BY line_name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []RailwayLine
	for rows.Next() {
		var line RailwayLine
		err := rows.Scan(
			&line.ID, &line.LineName, &line.LineCode, &line.LineType, &line.TotalDistance,
			&line.StartStation, &line.EndStation, &line.OpenedDate, &line.MaxSpeed,
			&line.Source, &line.CreatedAt, &line.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// Railway Segments

func (r *Repository) CreateSegment(segment *RailwaySegment) error {
	query := `
		INSERT INTO railway_segments (line_id, segment_name, start_station,
			end_station, distance, sequence)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, segment.LineID, segment.SegmentName,
		segment.StartStation, segment.EndStation, segment.Distance, segment.Sequence)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	segment.ID = int(id)
	return nil
}

func (r *Repository) GetSegmentsByLineID(lineID int) ([]RailwaySegment, error) {
	query := `
		SELECT id, line_id, segment_name, start_station, end_station, distance,
			sequence, created_at
		FROM railway_segments WHERE line_id = ? ORDER BY sequence
	`
	rows, err := r.db.Query(query, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []RailwaySegment
	for rows.Next() {
		var segment RailwaySegment
		err := rows.Scan(
			&segment.ID, &segment.LineID, &segment.SegmentName, &segment.StartStation,
			&segment.EndStation, &segment.Distance, &segment.Sequence, &segment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

// Railway Points

func (r *Repository) CreatePoint(point *RailwayPoint) error {
	query := `
		INSERT INTO railway_points (segment_id, longitude, latitude, altitude, sequence)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, point.SegmentID, point.Longitude,
		point.Latitude, point.Altitude, point.Sequence)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	point.ID = int(id)
	return nil
}

func (r *Repository) CreatePointsBatch(points []RailwayPoint) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO railway_points (segment_id, longitude, latitude, altitude, sequence)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, point := range points {
		_, err := stmt.Exec(point.SegmentID, point.Longitude, point.Latitude,
			point.Altitude, point.Sequence)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetPointsBySegmentID(segmentID int) ([]RailwayPoint, error) {
	query := `
		SELECT id, segment_id, longitude, latitude, altitude, sequence, created_at
		FROM railway_points WHERE segment_id = ? ORDER BY sequence
	`
	rows, err := r.db.Query(query, segmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []RailwayPoint
	for rows.Next() {
		var point RailwayPoint
		err := rows.Scan(
			&point.ID, &point.SegmentID, &point.Longitude, &point.Latitude,
			&point.Altitude, &point.Sequence, &point.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	return points, nil
}

// Railway Trips

func (r *Repository) CreateTrip(trip *RailwayTrip) error {
	query := `
		INSERT INTO railway_trips (train_number, line_id, departure_station,
			arrival_station, departure_time, arrival_time, duration_minutes,
			distance, seat_type, ticket_price, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, trip.TrainNumber, trip.LineID,
		trip.DepartureStation, trip.ArrivalStation, trip.DepartureTime,
		trip.ArrivalTime, trip.DurationMinutes, trip.Distance, trip.SeatType,
		trip.TicketPrice, trip.Notes)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	trip.ID = int(id)
	return nil
}

func (r *Repository) GetAllTrips() ([]RailwayTrip, error) {
	query := `
		SELECT id, train_number, line_id, departure_station, arrival_station,
			departure_time, arrival_time, duration_minutes, distance, seat_type,
			ticket_price, notes, created_at, updated_at
		FROM railway_trips ORDER BY departure_time DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []RailwayTrip
	for rows.Next() {
		var trip RailwayTrip
		err := rows.Scan(
			&trip.ID, &trip.TrainNumber, &trip.LineID, &trip.DepartureStation,
			&trip.ArrivalStation, &trip.DepartureTime, &trip.ArrivalTime,
			&trip.DurationMinutes, &trip.Distance, &trip.SeatType, &trip.TicketPrice,
			&trip.Notes, &trip.CreatedAt, &trip.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	return trips, nil
}

func (r *Repository) GetTripByID(id int) (*RailwayTrip, error) {
	query := `
		SELECT id, train_number, line_id, departure_station, arrival_station,
			departure_time, arrival_time, duration_minutes, distance, seat_type,
			ticket_price, notes, created_at, updated_at
		FROM railway_trips WHERE id = ?
	`
	trip := &RailwayTrip{}
	err := r.db.QueryRow(query, id).Scan(
		&trip.ID, &trip.TrainNumber, &trip.LineID, &trip.DepartureStation,
		&trip.ArrivalStation, &trip.DepartureTime, &trip.ArrivalTime,
		&trip.DurationMinutes, &trip.Distance, &trip.SeatType, &trip.TicketPrice,
		&trip.Notes, &trip.CreatedAt, &trip.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return trip, nil
}

// Statistics

func (r *Repository) GetStatistics() (*RailwayStatistics, error) {
	// Calculate statistics
	var stats RailwayStatistics

	// Total lines
	err := r.db.QueryRow("SELECT COUNT(*) FROM railway_lines").Scan(&stats.TotalLines)
	if err != nil {
		return nil, err
	}

	// Total trips
	err = r.db.QueryRow("SELECT COUNT(*) FROM railway_trips").Scan(&stats.TotalTrips)
	if err != nil {
		return nil, err
	}

	// Total distance and duration
	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(distance), 0), COALESCE(SUM(duration_minutes), 0)
		FROM railway_trips
	`).Scan(&stats.TotalDistance, &stats.TotalDuration)
	if err != nil {
		return nil, err
	}

	// Unique trains
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT train_number) FROM railway_trips
	`).Scan(&stats.UniqueTrains)
	if err != nil {
		return nil, err
	}

	// Date range
	err = r.db.QueryRow(`
		SELECT MIN(departure_time), MAX(departure_time) FROM railway_trips
	`).Scan(&stats.DateRangeStart, &stats.DateRangeEnd)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &stats, nil
}

// Get line with segments and points
func (r *Repository) GetLineWithRoute(lineID int) (*RailwayLine, error) {
	line, err := r.GetLineByID(lineID)
	if err != nil {
		return nil, err
	}

	segments, err := r.GetSegmentsByLineID(lineID)
	if err != nil {
		return nil, err
	}

	for i := range segments {
		points, err := r.GetPointsBySegmentID(segments[i].ID)
		if err != nil {
			return nil, err
		}
		segments[i].Points = points
	}

	line.Segments = segments
	return line, nil
}
