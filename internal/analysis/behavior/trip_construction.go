package behavior

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// TripConstructionAnalyzer implements trip construction from segments and stays
// Skill: 行程构建 (Trip Construction)
// Constructs complete trips by combining segments and stays
type TripConstructionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewTripConstructionAnalyzer creates a new trip construction analyzer
func NewTripConstructionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &TripConstructionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "trip_construction", 10000),
	}
}

// Analyze performs trip construction
func (a *TripConstructionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[TripConstructionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing trips (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM trips"); err != nil {
			return fmt.Errorf("failed to clear trips: %w", err)
		}
		log.Printf("[TripConstructionAnalyzer] Cleared existing trips")
	}

	// Query segments and stays ordered by time
	segmentsQuery := `
		SELECT
			id, start_ts, end_ts, mode, distance_m,
			start_lat, start_lon, end_lat, end_lon,
			start_province, start_city, start_county,
			end_province, end_city, end_county
		FROM segments
		ORDER BY start_ts
	`

	staysQuery := `
		SELECT
			id, start_ts, end_ts, duration_s,
			center_lat, center_lon, province, city, county
		FROM stay_segments
		WHERE duration_s >= 1800
		ORDER BY start_ts
	`

	// Load segments
	segments, err := a.loadSegments(ctx, segmentsQuery)
	if err != nil {
		return fmt.Errorf("failed to load segments: %w", err)
	}

	// Load stays
	stays, err := a.loadStays(ctx, staysQuery)
	if err != nil {
		return fmt.Errorf("failed to load stays: %w", err)
	}

	log.Printf("[TripConstructionAnalyzer] Loaded %d segments and %d stays", len(segments), len(stays))

	// Construct trips
	trips := a.constructTrips(segments, stays)

	log.Printf("[TripConstructionAnalyzer] Constructed %d trips", len(trips))

	// Insert trips
	if err := a.insertTrips(ctx, trips); err != nil {
		return fmt.Errorf("failed to insert trips: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_trips":    len(trips),
		"total_segments": len(segments),
		"total_stays":    len(stays),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[TripConstructionAnalyzer] Analysis completed")
	return nil
}

// Segment holds segment data
type Segment struct {
	ID            int64
	StartTS       int64
	EndTS         int64
	Mode          string
	Distance      float64
	StartLat      float64
	StartLon      float64
	EndLat        float64
	EndLon        float64
	StartProvince string
	StartCity     string
	StartCounty   string
	EndProvince   string
	EndCity       string
	EndCounty     string
}

// Stay holds stay data
type Stay struct {
	ID        int64
	StartTS   int64
	EndTS     int64
	Duration  int64
	CenterLat float64
	CenterLon float64
	Province  string
	City      string
	County    string
}

// Trip holds trip data
type Trip struct {
	StartTS         int64
	EndTS           int64
	OriginLat       float64
	OriginLon       float64
	OriginProvince  string
	OriginCity      string
	OriginCounty    string
	DestLat         float64
	DestLon         float64
	DestProvince    string
	DestCity        string
	DestCounty      string
	TotalDistance   float64
	Duration        int64
	PrimaryMode     string
	SegmentCount    int
	StayCount       int
	Purpose         string
	Confidence      float64
}

// loadSegments loads segments from database
func (a *TripConstructionAnalyzer) loadSegments(ctx context.Context, query string) ([]Segment, error) {
	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	var segments []Segment
	for rows.Next() {
		var seg Segment
		var startProvince, startCity, startCounty sql.NullString
		var endProvince, endCity, endCounty sql.NullString

		if err := rows.Scan(
			&seg.ID, &seg.StartTS, &seg.EndTS, &seg.Mode, &seg.Distance,
			&seg.StartLat, &seg.StartLon, &seg.EndLat, &seg.EndLon,
			&startProvince, &startCity, &startCounty,
			&endProvince, &endCity, &endCounty,
		); err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}

		if startProvince.Valid {
			seg.StartProvince = startProvince.String
		}
		if startCity.Valid {
			seg.StartCity = startCity.String
		}
		if startCounty.Valid {
			seg.StartCounty = startCounty.String
		}
		if endProvince.Valid {
			seg.EndProvince = endProvince.String
		}
		if endCity.Valid {
			seg.EndCity = endCity.String
		}
		if endCounty.Valid {
			seg.EndCounty = endCounty.String
		}

		segments = append(segments, seg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return segments, nil
}

// loadStays loads stays from database
func (a *TripConstructionAnalyzer) loadStays(ctx context.Context, query string) ([]Stay, error) {
	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stays: %w", err)
	}
	defer rows.Close()

	var stays []Stay
	for rows.Next() {
		var stay Stay
		var province, city, county sql.NullString

		if err := rows.Scan(
			&stay.ID, &stay.StartTS, &stay.EndTS, &stay.Duration,
			&stay.CenterLat, &stay.CenterLon,
			&province, &city, &county,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stay: %w", err)
		}

		if province.Valid {
			stay.Province = province.String
		}
		if city.Valid {
			stay.City = city.String
		}
		if county.Valid {
			stay.County = county.String
		}

		stays = append(stays, stay)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return stays, nil
}

// constructTrips constructs trips from segments and stays
func (a *TripConstructionAnalyzer) constructTrips(segments []Segment, stays []Stay) []Trip {
	if len(segments) == 0 {
		return nil
	}

	var trips []Trip
	var currentTrip *Trip
	var tripSegments []Segment
	var tripStays []Stay

	// Merge segments and stays into a timeline
	// Simple approach: group consecutive segments between significant stays

	for _, seg := range segments {
		if currentTrip == nil {
			// Start new trip
			currentTrip = &Trip{
				StartTS:        seg.StartTS,
				OriginLat:      seg.StartLat,
				OriginLon:      seg.StartLon,
				OriginProvince: seg.StartProvince,
				OriginCity:     seg.StartCity,
				OriginCounty:   seg.StartCounty,
			}
			tripSegments = []Segment{seg}
		} else {
			// Check if this segment continues the current trip
			// Gap threshold: 2 hours
			gap := seg.StartTS - currentTrip.EndTS
			if gap <= 7200 {
				// Continue current trip
				tripSegments = append(tripSegments, seg)
			} else {
				// End current trip and start new one
				a.finalizeTrip(currentTrip, tripSegments, tripStays)
				trips = append(trips, *currentTrip)

				// Start new trip
				currentTrip = &Trip{
					StartTS:        seg.StartTS,
					OriginLat:      seg.StartLat,
					OriginLon:      seg.StartLon,
					OriginProvince: seg.StartProvince,
					OriginCity:     seg.StartCity,
					OriginCounty:   seg.StartCounty,
				}
				tripSegments = []Segment{seg}
				tripStays = nil
			}
		}

		// Update trip end
		currentTrip.EndTS = seg.EndTS
		currentTrip.DestLat = seg.EndLat
		currentTrip.DestLon = seg.EndLon
		currentTrip.DestProvince = seg.EndProvince
		currentTrip.DestCity = seg.EndCity
		currentTrip.DestCounty = seg.EndCounty
	}

	// Finalize last trip
	if currentTrip != nil {
		a.finalizeTrip(currentTrip, tripSegments, tripStays)
		trips = append(trips, *currentTrip)
	}

	return trips
}

// finalizeTrip calculates trip statistics
func (a *TripConstructionAnalyzer) finalizeTrip(trip *Trip, segments []Segment, stays []Stay) {
	// Calculate total distance
	for _, seg := range segments {
		trip.TotalDistance += seg.Distance
	}

	// Calculate duration
	trip.Duration = trip.EndTS - trip.StartTS

	// Count segments and stays
	trip.SegmentCount = len(segments)
	trip.StayCount = len(stays)

	// Determine primary mode (most common mode by distance)
	modeDistances := make(map[string]float64)
	for _, seg := range segments {
		modeDistances[seg.Mode] += seg.Distance
	}

	maxDistance := 0.0
	for mode, distance := range modeDistances {
		if distance > maxDistance {
			maxDistance = distance
			trip.PrimaryMode = mode
		}
	}

	// Infer trip purpose (simplified heuristics)
	trip.Purpose = "UNKNOWN"
	trip.Confidence = 0.5

	// If trip is short and within same city, likely commute
	if trip.Duration < 7200 && trip.OriginCity == trip.DestCity {
		trip.Purpose = "COMMUTE"
		trip.Confidence = 0.7
	}

	// If trip crosses provinces, likely travel
	if trip.OriginProvince != trip.DestProvince {
		trip.Purpose = "TRAVEL"
		trip.Confidence = 0.8
	}

	// If trip is on weekend, likely leisure
	// (Would need to check day of week from timestamp)
}

// insertTrips inserts trips into the database
func (a *TripConstructionAnalyzer) insertTrips(ctx context.Context, trips []Trip) error {
	if len(trips) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO trips (
			start_ts, end_ts,
			origin_lat, origin_lon, origin_province, origin_city, origin_county,
			dest_lat, dest_lon, dest_province, dest_city, dest_county,
			total_distance_m, duration_s, primary_mode,
			segment_count, stay_count, purpose, confidence,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, trip := range trips {
		_, err := stmt.ExecContext(ctx,
			trip.StartTS, trip.EndTS,
			trip.OriginLat, trip.OriginLon, trip.OriginProvince, trip.OriginCity, trip.OriginCounty,
			trip.DestLat, trip.DestLon, trip.DestProvince, trip.DestCity, trip.DestCounty,
			trip.TotalDistance, trip.Duration, trip.PrimaryMode,
			trip.SegmentCount, trip.StayCount, trip.Purpose, trip.Confidence,
		)
		if err != nil {
			return fmt.Errorf("failed to insert trip: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[TripConstructionAnalyzer] Inserted %d trips", len(trips))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("trip_construction", NewTripConstructionAnalyzer)
}