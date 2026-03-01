package flights

import "time"

// Flight represents a flight record
type Flight struct {
	ID               int       `json:"id"`
	FlightNumber     string    `json:"flightNumber"`
	AircraftNumber   string    `json:"aircraftNumber,omitempty"`
	Airline          string    `json:"airline,omitempty"`
	DepartureAirport string    `json:"departureAirport,omitempty"`
	ArrivalAirport   string    `json:"arrivalAirport,omitempty"`
	DepartureTime    time.Time `json:"departureTime,omitempty"`
	ArrivalTime      time.Time `json:"arrivalTime,omitempty"`
	ActualDeparture  time.Time `json:"actualDeparture,omitempty"`
	ActualArrival    time.Time `json:"actualArrival,omitempty"`
	FlightDate       string    `json:"flightDate"`
	DataSource       string    `json:"dataSource"`
	CreatedAt        time.Time `json:"createdAt"`
}

// FlightPoint represents a GPS tracking point during flight
type FlightPoint struct {
	ID         int       `json:"id"`
	FlightID   int       `json:"flightId"`
	UpdateTime int64     `json:"updateTime"` // Unix timestamp
	UTCTime    string    `json:"utcTime,omitempty"`
	Longitude  float64   `json:"longitude"`
	Latitude   float64   `json:"latitude"`
	Altitude   float64   `json:"altitude,omitempty"`   // meters
	Speed      float64   `json:"speed,omitempty"`      // km/h
	Heading    float64   `json:"heading,omitempty"`    // degrees
	CreatedAt  time.Time `json:"createdAt"`
}

// FlightStatistics represents pre-computed flight statistics
type FlightStatistics struct {
	ID              int       `json:"id"`
	FlightID        int       `json:"flightId"`
	TotalDistance   float64   `json:"totalDistance"`   // km
	MaxAltitude     float64   `json:"maxAltitude"`     // meters
	MaxSpeed        float64   `json:"maxSpeed"`        // km/h
	AvgSpeed        float64   `json:"avgSpeed"`        // km/h
	DurationMinutes int       `json:"durationMinutes"` // minutes
	PointCount      int       `json:"pointCount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// FlightTrackMatch represents a match between flight data and GPS tracks
type FlightTrackMatch struct {
	ID               int       `json:"id"`
	FlightID         int       `json:"flightId"`
	TrackPointID     int       `json:"trackPointId"`
	MatchConfidence  float64   `json:"matchConfidence"`  // 0.0-1.0
	TimeDiffSeconds  int       `json:"timeDiffSeconds"`  // seconds
	DistanceMeters   float64   `json:"distanceMeters"`   // meters
	MatchType        string    `json:"matchType"`        // exact, interpolated, estimated
	CreatedAt        time.Time `json:"createdAt"`
}

// FlightWithStats combines flight info with statistics
type FlightWithStats struct {
	Flight
	Statistics *FlightStatistics `json:"statistics,omitempty"`
	PointCount int               `json:"pointCount"`
}

// FlightSummary provides a summary view of all flights
type FlightSummary struct {
	TotalFlights    int     `json:"totalFlights"`
	TotalDistance   float64 `json:"totalDistance"`   // km
	TotalDuration   int     `json:"totalDuration"`   // minutes
	AverageDistance float64 `json:"averageDistance"` // km
	AverageDuration int     `json:"averageDuration"` // minutes
	Airlines        []string `json:"airlines"`
	MostFrequentRoute string `json:"mostFrequentRoute,omitempty"`
}

// VariflightData represents the JSON structure from Variflight exports
type VariflightData struct {
	FlightNumber string `json:"fnum"`
	AircraftNum  string `json:"anum"`
	Points       []struct {
		UpdateTime int64   `json:"updatetime"` // Unix timestamp
		UTCTime    string  `json:"UTC Time"`
		Longitude  float64 `json:"longitude"`
		Latitude   float64 `json:"latitude"`
		Height     float64 `json:"height"` // meters
		Speed      float64 `json:"speed"`  // km/h
		Angle      float64 `json:"angle"`  // heading
	} `json:"points"`
}
