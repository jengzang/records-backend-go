package railway

import "time"

// RailwayLine represents a railway line
type RailwayLine struct {
	ID            int       `json:"id"`
	LineName      string    `json:"line_name"`
	LineCode      string    `json:"line_code,omitempty"`
	LineType      string    `json:"line_type"` // 高速/普速/城际
	TotalDistance float64   `json:"total_distance,omitempty"`
	StartStation  string    `json:"start_station,omitempty"`
	EndStation    string    `json:"end_station,omitempty"`
	OpenedDate    string    `json:"opened_date,omitempty"`
	MaxSpeed      int       `json:"max_speed,omitempty"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Segments      []RailwaySegment `json:"segments,omitempty"`
}

// RailwaySegment represents a segment of a railway line
type RailwaySegment struct {
	ID           int       `json:"id"`
	LineID       int       `json:"line_id"`
	SegmentName  string    `json:"segment_name,omitempty"`
	StartStation string    `json:"start_station"`
	EndStation   string    `json:"end_station"`
	Distance     float64   `json:"distance,omitempty"`
	Sequence     int       `json:"sequence"`
	CreatedAt    time.Time `json:"created_at"`
	Points       []RailwayPoint `json:"points,omitempty"`
}

// RailwayPoint represents a coordinate point on a railway segment
type RailwayPoint struct {
	ID        int       `json:"id"`
	SegmentID int       `json:"segment_id"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	Altitude  float64   `json:"altitude,omitempty"`
	Sequence  int       `json:"sequence"`
	CreatedAt time.Time `json:"created_at"`
}

// RailwayTrip represents a train trip record
type RailwayTrip struct {
	ID               int       `json:"id"`
	TrainNumber      string    `json:"train_number"`
	LineID           *int      `json:"line_id,omitempty"`
	DepartureStation string    `json:"departure_station"`
	ArrivalStation   string    `json:"arrival_station"`
	DepartureTime    time.Time `json:"departure_time"`
	ArrivalTime      time.Time `json:"arrival_time"`
	DurationMinutes  int       `json:"duration_minutes"`
	Distance         float64   `json:"distance,omitempty"`
	SeatType         string    `json:"seat_type,omitempty"`
	TicketPrice      float64   `json:"ticket_price,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Line             *RailwayLine `json:"line,omitempty"`
}

// RailwayStatistics represents overall railway statistics
type RailwayStatistics struct {
	ID             int       `json:"id"`
	TotalLines     int       `json:"total_lines"`
	TotalTrips     int       `json:"total_trips"`
	TotalDistance  float64   `json:"total_distance"`
	TotalDuration  int       `json:"total_duration"`
	UniqueTrains   int       `json:"unique_trains"`
	DateRangeStart string    `json:"date_range_start,omitempty"`
	DateRangeEnd   string    `json:"date_range_end,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// KMLData represents parsed KML file data
type KMLData struct {
	Name        string
	Description string
	Coordinates [][]float64 // [longitude, latitude, altitude]
}
