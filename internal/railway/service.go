package railway

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Lines

func (s *Service) GetAllLines() ([]RailwayLine, error) {
	return s.repo.GetAllLines()
}

func (s *Service) GetLineByID(id int) (*RailwayLine, error) {
	return s.repo.GetLineByID(id)
}

func (s *Service) GetLineWithRoute(id int) (*RailwayLine, error) {
	return s.repo.GetLineWithRoute(id)
}

func (s *Service) CreateLine(line *RailwayLine) error {
	return s.repo.CreateLine(line)
}

// Trips

func (s *Service) GetAllTrips() ([]RailwayTrip, error) {
	return s.repo.GetAllTrips()
}

func (s *Service) GetTripByID(id int) (*RailwayTrip, error) {
	trip, err := s.repo.GetTripByID(id)
	if err != nil {
		return nil, err
	}

	// Load associated line if exists
	if trip.LineID != nil {
		line, err := s.repo.GetLineByID(*trip.LineID)
		if err == nil {
			trip.Line = line
		}
	}

	return trip, nil
}

func (s *Service) CreateTrip(trip *RailwayTrip) error {
	return s.repo.CreateTrip(trip)
}

// Statistics

func (s *Service) GetStatistics() (*RailwayStatistics, error) {
	return s.repo.GetStatistics()
}

// KML Parsing

// KML structure for parsing
type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

type Document struct {
	Placemarks []Placemark `xml:"Placemark"`
}

type Placemark struct {
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	LineString  LineString  `xml:"LineString"`
}

type LineString struct {
	Coordinates string `xml:"coordinates"`
}

// ParseKML parses a KML file and returns KMLData
func (s *Service) ParseKML(reader io.Reader) (*KMLData, error) {
	var kml KML
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&kml); err != nil {
		return nil, fmt.Errorf("failed to decode KML: %w", err)
	}

	if len(kml.Document.Placemarks) == 0 {
		return nil, fmt.Errorf("no placemarks found in KML")
	}

	placemark := kml.Document.Placemarks[0]
	coords, err := parseCoordinates(placemark.LineString.Coordinates)
	if err != nil {
		return nil, err
	}

	return &KMLData{
		Name:        placemark.Name,
		Description: placemark.Description,
		Coordinates: coords,
	}, nil
}

// parseCoordinates parses KML coordinate string
// Format: "lon,lat,alt lon,lat,alt ..."
func parseCoordinates(coordStr string) ([][]float64, error) {
	coordStr = strings.TrimSpace(coordStr)
	points := strings.Fields(coordStr)

	var coords [][]float64
	for _, point := range points {
		parts := strings.Split(point, ",")
		if len(parts) < 2 {
			continue
		}

		lon, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude: %s", parts[0])
		}

		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude: %s", parts[1])
		}

		alt := 0.0
		if len(parts) >= 3 {
			alt, _ = strconv.ParseFloat(parts[2], 64)
		}

		coords = append(coords, []float64{lon, lat, alt})
	}

	return coords, nil
}

// ImportKMLLine imports a railway line from KML data
func (s *Service) ImportKMLLine(kmlData *KMLData, lineType string) (*RailwayLine, error) {
	// Extract line name and stations from KML name
	// Format: "线路名(起点-终点), 段名"
	lineName, startStation, endStation := parseLineName(kmlData.Name)

	// Create line
	line := &RailwayLine{
		LineName:     lineName,
		LineType:     lineType,
		StartStation: startStation,
		EndStation:   endStation,
		Source:       "openrailmap",
	}

	if err := s.repo.CreateLine(line); err != nil {
		return nil, err
	}

	// Create segment
	segment := &RailwaySegment{
		LineID:       line.ID,
		SegmentName:  kmlData.Name,
		StartStation: startStation,
		EndStation:   endStation,
		Sequence:     1,
	}

	if err := s.repo.CreateSegment(segment); err != nil {
		return nil, err
	}

	// Create points
	var points []RailwayPoint
	for i, coord := range kmlData.Coordinates {
		points = append(points, RailwayPoint{
			SegmentID: segment.ID,
			Longitude: coord[0],
			Latitude:  coord[1],
			Altitude:  coord[2],
			Sequence:  i + 1,
		})
	}

	if err := s.repo.CreatePointsBatch(points); err != nil {
		return nil, err
	}

	return line, nil
}

// parseLineName extracts line name and stations from KML name
// Example: "京广线(广州-北京), 广州-韶关东" -> "京广线", "广州", "韶关东"
func parseLineName(name string) (lineName, startStation, endStation string) {
	// Split by comma
	parts := strings.Split(name, ",")
	if len(parts) == 0 {
		return name, "", ""
	}

	// Extract line name
	firstPart := strings.TrimSpace(parts[0])
	if idx := strings.Index(firstPart, "("); idx != -1 {
		lineName = firstPart[:idx]
	} else {
		lineName = firstPart
	}

	// Extract stations from segment name
	if len(parts) >= 2 {
		segmentName := strings.TrimSpace(parts[1])
		stations := strings.Split(segmentName, "-")
		if len(stations) >= 2 {
			startStation = strings.TrimSpace(stations[0])
			endStation = strings.TrimSpace(stations[1])
		}
	}

	return
}
