package behavior

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
			id, start_time, end_time, mode, distance_m, duration_s
		FROM segments
		ORDER BY start_time
	`

	staysQuery := `
		SELECT
			id, start_time, end_time, duration_s
		FROM stay_segments
		WHERE duration_s >= 1800
		ORDER BY start_time
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
	ID       int64
	StartTime int64
	EndTime   int64
	Mode     string
	Distance float64
	Duration int64
}

// Stay holds stay data
type Stay struct {
	ID       int64
	StartTime int64
	EndTime   int64
	Duration int64
}

// Trip holds trip data
type Trip struct {
	Date          string  // YYYY-MM-DD
	TripNumber    int     // 1, 2, 3... for the day
	OriginStayID  *int64  // Foreign key to stay_segments
	DestStayID    *int64  // Foreign key to stay_segments
	StartTime     int64
	EndTime       int64
	Duration      int64
	Distance      float64
	SegmentCount  int
	Modes         string  // JSON array of modes
	Metadata      string  // JSON object
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

		if err := rows.Scan(
			&seg.ID, &seg.StartTime, &seg.EndTime, &seg.Mode, &seg.Distance, &seg.Duration,
		); err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
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

		if err := rows.Scan(
			&stay.ID, &stay.StartTime, &stay.EndTime, &stay.Duration,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stay: %w", err)
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
	tripsByDate := make(map[string]int) // Track trip numbers per date

	// Simple approach: group consecutive segments with small gaps
	for _, seg := range segments {
		if currentTrip == nil {
			// Start new trip
			date := a.timestampToDate(seg.StartTime)
			tripsByDate[date]++

			currentTrip = &Trip{
				Date:       date,
				TripNumber: tripsByDate[date],
				StartTime:  seg.StartTime,
			}
			tripSegments = []Segment{seg}
		} else {
			// Check if this segment continues the current trip
			// Gap threshold: 2 hours
			gap := seg.StartTime - currentTrip.EndTime
			if gap <= 7200 {
				// Continue current trip
				tripSegments = append(tripSegments, seg)
			} else {
				// End current trip and start new one
				a.finalizeTrip(currentTrip, tripSegments, stays)
				trips = append(trips, *currentTrip)

				// Start new trip
				date := a.timestampToDate(seg.StartTime)
				tripsByDate[date]++

				currentTrip = &Trip{
					Date:       date,
					TripNumber: tripsByDate[date],
					StartTime:  seg.StartTime,
				}
				tripSegments = []Segment{seg}
			}
		}

		// Update trip end
		currentTrip.EndTime = seg.EndTime
	}

	// Finalize last trip
	if currentTrip != nil {
		a.finalizeTrip(currentTrip, tripSegments, stays)
		trips = append(trips, *currentTrip)
	}

	return trips
}

// timestampToDate converts Unix timestamp to YYYY-MM-DD format
func (a *TripConstructionAnalyzer) timestampToDate(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02")
}

// finalizeTrip calculates trip statistics
func (a *TripConstructionAnalyzer) finalizeTrip(trip *Trip, segments []Segment, stays []Stay) {
	// Calculate total distance
	totalDistance := 0.0
	for _, seg := range segments {
		totalDistance += seg.Distance
	}
	trip.Distance = totalDistance

	// Calculate duration
	trip.Duration = trip.EndTime - trip.StartTime

	// Count segments
	trip.SegmentCount = len(segments)

	// Create modes JSON array (list of unique modes used)
	modeSet := make(map[string]bool)
	for _, seg := range segments {
		modeSet[seg.Mode] = true
	}
	modes := []string{}
	for mode := range modeSet {
		modes = append(modes, mode)
	}
	modesJSON, _ := json.Marshal(modes)
	trip.Modes = string(modesJSON)

	// Try to link to origin and destination stays
	// Find stay that ends just before trip starts
	for _, stay := range stays {
		if stay.EndTime <= trip.StartTime && trip.StartTime-stay.EndTime < 3600 {
			trip.OriginStayID = &stay.ID
		}
		if stay.StartTime >= trip.EndTime && stay.StartTime-trip.EndTime < 3600 {
			trip.DestStayID = &stay.ID
		}
	}

	// Create metadata JSON object
	metadata := map[string]interface{}{
		"algorithm":      "simple_gap_based",
		"gap_threshold_s": 7200,
		"segment_ids":    []int64{},
	}
	for _, seg := range segments {
		metadata["segment_ids"] = append(metadata["segment_ids"].([]int64), seg.ID)
	}
	metadataJSON, _ := json.Marshal(metadata)
	trip.Metadata = string(metadataJSON)
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
			date, trip_number,
			origin_stay_id, dest_stay_id,
			start_time, end_time, duration_s,
			distance_m, segment_count, modes, metadata,
			algo_version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1',
		          CAST(strftime('%s', 'now') AS INTEGER),
		          CAST(strftime('%s', 'now') AS INTEGER))
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, trip := range trips {
		_, err := stmt.ExecContext(ctx,
			trip.Date, trip.TripNumber,
			trip.OriginStayID, trip.DestStayID,
			trip.StartTime, trip.EndTime, trip.Duration,
			trip.Distance, trip.SegmentCount, trip.Modes, trip.Metadata,
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