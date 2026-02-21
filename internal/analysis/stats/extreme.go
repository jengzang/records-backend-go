package stats

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/jengzang/records-backend-go/internal/analysis"
	"github.com/jengzang/records-backend-go/internal/stats"
)

// ExtremeEventsAnalyzer implements extreme events detection
// Skill: 极值旅行事件 (Extreme Events)
// Finds highest altitude, furthest east/west/north/south trips
type ExtremeEventsAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewExtremeEventsAnalyzer creates a new extreme events analyzer
func NewExtremeEventsAnalyzer(db *sql.DB) analysis.Analyzer {
	return &ExtremeEventsAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "extreme_events", 1000),
	}
}

// Analyze performs extreme events detection
func (a *ExtremeEventsAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[ExtremeEventsAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing extreme events (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM extreme_events"); err != nil {
			return fmt.Errorf("failed to clear extreme events: %w", err)
		}
		log.Printf("[ExtremeEventsAnalyzer] Cleared existing extreme events")
	}

	// Get all trips
	tripsQuery := `
		SELECT
			id,
			start_time,
			end_time
		FROM trips
		ORDER BY id
	`

	rows, err := a.DB.QueryContext(ctx, tripsQuery)
	if err != nil {
		return fmt.Errorf("failed to query trips: %w", err)
	}

	var trips []TripInfo
	for rows.Next() {
		var trip TripInfo
		if err := rows.Scan(&trip.ID, &trip.StartTS, &trip.EndTS); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan trip: %w", err)
		}
		trips = append(trips, trip)
	}
	rows.Close()

	log.Printf("[ExtremeEventsAnalyzer] Processing %d trips", len(trips))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(trips)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Process each trip to find extremes
	extremes := make(map[string]*ExtremeEvent)
	processed := 0

	for _, trip := range trips {
		// Get points for this trip
		pointsQuery := `
			SELECT
				id,
				dataTime,
				latitude,
				longitude,
				altitude,
				province,
				city,
				county
			FROM "一生足迹"
			WHERE dataTime BETWEEN ? AND ?
				AND outlier_flag = 0
			ORDER BY dataTime
		`

		pointRows, err := a.DB.QueryContext(ctx, pointsQuery, trip.StartTS, trip.EndTS)
		if err != nil {
			return fmt.Errorf("failed to query points for trip %d: %w", trip.ID, err)
		}

		var points []PointData
		for pointRows.Next() {
			var point PointData
			var altitude sql.NullFloat64
			var province, city, county sql.NullString
			if err := pointRows.Scan(&point.ID, &point.Timestamp, &point.Lat, &point.Lon, &altitude, &province, &city, &county); err != nil {
				pointRows.Close()
				return fmt.Errorf("failed to scan point: %w", err)
			}
			if altitude.Valid {
				point.Alt = altitude.Float64
			}
			if province.Valid {
				point.Province = province.String
			}
			if city.Valid {
				point.City = city.String
			}
			if county.Valid {
				point.County = county.String
			}
			points = append(points, point)
		}
		pointRows.Close()

		if len(points) == 0 {
			continue
		}

		// Calculate extremes for this trip
		a.calculateTripExtremes(extremes, trip, points)

		processed++
		if processed%100 == 0 {
			if err := a.UpdateTaskProgress(taskID, int64(len(trips)), int64(processed), 0); err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}
			log.Printf("[ExtremeEventsAnalyzer] Processed %d/%d trips", processed, len(trips))
		}
	}

	// Insert extreme events
	if err := a.insertExtremeEvents(ctx, extremes); err != nil {
		return fmt.Errorf("failed to insert extreme events: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_trips":     len(trips),
		"processed_trips": processed,
		"extreme_events":  len(extremes),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[ExtremeEventsAnalyzer] Analysis completed: %d trips processed, %d extreme events found", processed, len(extremes))
	return nil
}

// TripInfo holds trip information
type TripInfo struct {
	ID      int64
	StartTS int64
	EndTS   int64
}

// PointData holds point data
type PointData struct {
	ID        int64
	Timestamp int64
	Lat       float64
	Lon       float64
	Alt       float64
	Province  string
	City      string
	County    string
}

// ExtremeEvent holds extreme event data
type ExtremeEvent struct {
	EventType string
	TripID    int64
	PointID   int64
	Value     float64
	Latitude  float64
	Longitude float64
	Timestamp int64
	Province  string
	City      string
	County    string
}

// calculateTripExtremes calculates extreme events for a trip
func (a *ExtremeEventsAnalyzer) calculateTripExtremes(extremes map[string]*ExtremeEvent, trip TripInfo, points []PointData) {
	if len(points) == 0 {
		return
	}

	// Extract values for percentile calculation
	var altitudes, latitudes, longitudes []float64
	for _, p := range points {
		if p.Alt != 0 {
			altitudes = append(altitudes, p.Alt)
		}
		latitudes = append(latitudes, p.Lat)
		longitudes = append(longitudes, p.Lon)
	}

	// MAX_ALTITUDE (use p99 for robustness)
	if len(altitudes) > 0 {
		sort.Float64s(altitudes)
		maxAlt := stats.Percentile(altitudes, 99)

		// Find the point with this altitude
		for _, p := range points {
			if p.Alt >= maxAlt {
				key := fmt.Sprintf("MAX_ALTITUDE_%d", trip.ID)
				event := &ExtremeEvent{
					EventType: "MAX_ALTITUDE",
					TripID:    trip.ID,
					PointID:   p.ID,
					Value:     p.Alt,
					Latitude:  p.Lat,
					Longitude: p.Lon,
					Timestamp: p.Timestamp,
					Province:  p.Province,
					City:      p.City,
					County:    p.County,
				}
				extremes[key] = event
				break
			}
		}
	}

	// EASTMOST (use p99 longitude)
	if len(longitudes) > 0 {
		sort.Float64s(longitudes)
		eastmost := stats.Percentile(longitudes, 99)

		for _, p := range points {
			if p.Lon >= eastmost {
				key := fmt.Sprintf("EASTMOST_%d", trip.ID)
				event := &ExtremeEvent{
					EventType: "EASTMOST",
					TripID:    trip.ID,
					PointID:   p.ID,
					Value:     p.Lon,
					Latitude:  p.Lat,
					Longitude: p.Lon,
					Timestamp: p.Timestamp,
					Province:  p.Province,
					City:      p.City,
					County:    p.County,
				}
				extremes[key] = event
				break
			}
		}
	}

	// WESTMOST (use p01 longitude)
	if len(longitudes) > 0 {
		westmost := stats.Percentile(longitudes, 1)

		for _, p := range points {
			if p.Lon <= westmost {
				key := fmt.Sprintf("WESTMOST_%d", trip.ID)
				event := &ExtremeEvent{
					EventType: "WESTMOST",
					TripID:    trip.ID,
					PointID:   p.ID,
					Value:     p.Lon,
					Latitude:  p.Lat,
					Longitude: p.Lon,
					Timestamp: p.Timestamp,
					Province:  p.Province,
					City:      p.City,
					County:    p.County,
				}
				extremes[key] = event
				break
			}
		}
	}

	// NORTHMOST (use p99 latitude)
	if len(latitudes) > 0 {
		sort.Float64s(latitudes)
		northmost := stats.Percentile(latitudes, 99)

		for _, p := range points {
			if p.Lat >= northmost {
				key := fmt.Sprintf("NORTHMOST_%d", trip.ID)
				event := &ExtremeEvent{
					EventType: "NORTHMOST",
					TripID:    trip.ID,
					PointID:   p.ID,
					Value:     p.Lat,
					Latitude:  p.Lat,
					Longitude: p.Lon,
					Timestamp: p.Timestamp,
					Province:  p.Province,
					City:      p.City,
					County:    p.County,
				}
				extremes[key] = event
				break
			}
		}
	}

	// SOUTHMOST (use p01 latitude)
	if len(latitudes) > 0 {
		southmost := stats.Percentile(latitudes, 1)

		for _, p := range points {
			if p.Lat <= southmost {
				key := fmt.Sprintf("SOUTHMOST_%d", trip.ID)
				event := &ExtremeEvent{
					EventType: "SOUTHMOST",
					TripID:    trip.ID,
					PointID:   p.ID,
					Value:     p.Lat,
					Latitude:  p.Lat,
					Longitude: p.Lon,
					Timestamp: p.Timestamp,
					Province:  p.Province,
					City:      p.City,
					County:    p.County,
				}
				extremes[key] = event
				break
			}
		}
	}
}

// insertExtremeEvents inserts extreme events into the database
func (a *ExtremeEventsAnalyzer) insertExtremeEvents(ctx context.Context, extremes map[string]*ExtremeEvent) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO extreme_events (
			event_type, point_id, value, latitude, longitude, timestamp, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range extremes {
		metadata := map[string]interface{}{
			"trip_id":  event.TripID,
			"province": event.Province,
			"city":     event.City,
			"county":   event.County,
		}
		metadataJSON, _ := json.Marshal(metadata)

		_, err := stmt.ExecContext(ctx,
			event.EventType,
			event.PointID,
			event.Value,
			event.Latitude,
			event.Longitude,
			event.Timestamp,
			string(metadataJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert extreme event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("extreme_events", NewExtremeEventsAnalyzer)
}
