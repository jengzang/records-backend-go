package viz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// TimeAxisMapAnalyzer implements time-axis visualization metadata generation
// Skill: 时间轴地图 (Time Axis Map)
// Generates timeline markers for trajectory visualization
type TimeAxisMapAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewTimeAxisMapAnalyzer creates a new time axis map analyzer
func NewTimeAxisMapAnalyzer(db *sql.DB) analysis.Analyzer {
	return &TimeAxisMapAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "time_axis_map", 10000),
	}
}

// Analyze performs time axis map generation
func (a *TimeAxisMapAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[TimeAxisMapAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing markers (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM time_axis_markers"); err != nil {
			return fmt.Errorf("failed to clear time_axis_markers: %w", err)
		}
		log.Printf("[TimeAxisMapAnalyzer] Cleared existing time axis markers")
	}

	var markers []TimeAxisMarker

	// 1. Generate markers from segments
	segmentMarkers, err := a.generateSegmentMarkers(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate segment markers: %w", err)
	}
	markers = append(markers, segmentMarkers...)

	// 2. Generate markers from stays
	stayMarkers, err := a.generateStayMarkers(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate stay markers: %w", err)
	}
	markers = append(markers, stayMarkers...)

	// 3. Generate markers from speed events
	speedMarkers, err := a.generateSpeedEventMarkers(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate speed event markers: %w", err)
	}
	markers = append(markers, speedMarkers...)

	// 4. Generate markers from altitude events
	altitudeMarkers, err := a.generateAltitudeEventMarkers(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate altitude event markers: %w", err)
	}
	markers = append(markers, altitudeMarkers...)

	log.Printf("[TimeAxisMapAnalyzer] Generated %d markers", len(markers))

	// Insert markers
	if err := a.insertTimeAxisMarkers(ctx, markers); err != nil {
		return fmt.Errorf("failed to insert time axis markers: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_markers":    len(markers),
		"segment_markers":  len(segmentMarkers),
		"stay_markers":     len(stayMarkers),
		"speed_markers":    len(speedMarkers),
		"altitude_markers": len(altitudeMarkers),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[TimeAxisMapAnalyzer] Analysis completed")
	return nil
}

// TimeAxisMarker holds time axis marker data
type TimeAxisMarker struct {
	MarkerTS   int64
	MarkerType string
	EntityID   int64
	EntityType string
	Latitude   float64
	Longitude  float64
	Label      string
	Icon       string
	Color      string
}

// generateSegmentMarkers generates markers from segments
func (a *TimeAxisMapAnalyzer) generateSegmentMarkers(ctx context.Context) ([]TimeAxisMarker, error) {
	query := `
		SELECT
			id, start_ts, end_ts, mode,
			start_lat, start_lon, end_lat, end_lon
		FROM segments
		ORDER BY start_ts
		LIMIT 1000
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	var markers []TimeAxisMarker
	for rows.Next() {
		var id, startTS, endTS int64
		var mode string
		var startLat, startLon, endLat, endLon float64

		if err := rows.Scan(&id, &startTS, &endTS, &mode, &startLat, &startLon, &endLat, &endLon); err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}

		// Start marker
		markers = append(markers, TimeAxisMarker{
			MarkerTS:   startTS,
			MarkerType: "SEGMENT_START",
			EntityID:   id,
			EntityType: "SEGMENT",
			Latitude:   startLat,
			Longitude:  startLon,
			Label:      fmt.Sprintf("%s Start", mode),
			Icon:       a.getModeIcon(mode),
			Color:      a.getModeColor(mode),
		})

		// End marker
		markers = append(markers, TimeAxisMarker{
			MarkerTS:   endTS,
			MarkerType: "SEGMENT_END",
			EntityID:   id,
			EntityType: "SEGMENT",
			Latitude:   endLat,
			Longitude:  endLon,
			Label:      fmt.Sprintf("%s End", mode),
			Icon:       a.getModeIcon(mode),
			Color:      a.getModeColor(mode),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return markers, nil
}

// generateStayMarkers generates markers from stays
func (a *TimeAxisMapAnalyzer) generateStayMarkers(ctx context.Context) ([]TimeAxisMarker, error) {
	query := `
		SELECT
			id, start_ts, center_lat, center_lon, duration_s
		FROM stay_segments
		WHERE duration_s >= 7200
		ORDER BY start_ts
		LIMIT 500
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stays: %w", err)
	}
	defer rows.Close()

	var markers []TimeAxisMarker
	for rows.Next() {
		var id, startTS, duration int64
		var centerLat, centerLon float64

		if err := rows.Scan(&id, &startTS, &centerLat, &centerLon, &duration); err != nil {
			return nil, fmt.Errorf("failed to scan stay: %w", err)
		}

		durationHours := duration / 3600
		markers = append(markers, TimeAxisMarker{
			MarkerTS:   startTS,
			MarkerType: "STAY",
			EntityID:   id,
			EntityType: "STAY",
			Latitude:   centerLat,
			Longitude:  centerLon,
			Label:      fmt.Sprintf("Stay (%dh)", durationHours),
			Icon:       "pin",
			Color:      "#FF6B6B",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return markers, nil
}

// generateSpeedEventMarkers generates markers from speed events
func (a *TimeAxisMapAnalyzer) generateSpeedEventMarkers(ctx context.Context) ([]TimeAxisMarker, error) {
	query := `
		SELECT
			id, peak_ts, peak_lat, peak_lon, max_speed_mps
		FROM speed_events
		ORDER BY peak_ts
		LIMIT 200
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query speed events: %w", err)
	}
	defer rows.Close()

	var markers []TimeAxisMarker
	for rows.Next() {
		var id, peakTS int64
		var peakLat, peakLon, maxSpeed float64

		if err := rows.Scan(&id, &peakTS, &peakLat, &peakLon, &maxSpeed); err != nil {
			return nil, fmt.Errorf("failed to scan speed event: %w", err)
		}

		speedKmh := maxSpeed * 3.6
		markers = append(markers, TimeAxisMarker{
			MarkerTS:   peakTS,
			MarkerType: "EVENT",
			EntityID:   id,
			EntityType: "SPEED_EVENT",
			Latitude:   peakLat,
			Longitude:  peakLon,
			Label:      fmt.Sprintf("Speed %.0f km/h", speedKmh),
			Icon:       "flash",
			Color:      "#FFA500",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return markers, nil
}

// generateAltitudeEventMarkers generates markers from altitude events
func (a *TimeAxisMapAnalyzer) generateAltitudeEventMarkers(ctx context.Context) ([]TimeAxisMarker, error) {
	query := `
		SELECT
			id, start_ts, start_altitude, altitude_change, event_type
		FROM altitude_events
		WHERE ABS(altitude_change) >= 100
		ORDER BY start_ts
		LIMIT 200
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query altitude events: %w", err)
	}
	defer rows.Close()

	var markers []TimeAxisMarker
	for rows.Next() {
		var id, startTS int64
		var startAltitude, altitudeChange float64
		var eventType string

		if err := rows.Scan(&id, &startTS, &startAltitude, &altitudeChange, &eventType); err != nil {
			return nil, fmt.Errorf("failed to scan altitude event: %w", err)
		}

		icon := "mountain"
		if eventType == "DESCENT" {
			icon = "arrow-down"
		}

		markers = append(markers, TimeAxisMarker{
			MarkerTS:   startTS,
			MarkerType: "EVENT",
			EntityID:   id,
			EntityType: "ALTITUDE_EVENT",
			Latitude:   0, // Would need to query track points for exact location
			Longitude:  0,
			Label:      fmt.Sprintf("%s %.0fm", eventType, altitudeChange),
			Icon:       icon,
			Color:      "#4CAF50",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return markers, nil
}

// getModeIcon returns icon for transport mode
func (a *TimeAxisMapAnalyzer) getModeIcon(mode string) string {
	icons := map[string]string{
		"WALK":  "walk",
		"BIKE":  "bike",
		"CAR":   "car",
		"TRAIN": "train",
		"PLANE": "plane",
	}
	if icon, ok := icons[mode]; ok {
		return icon
	}
	return "circle"
}

// getModeColor returns color for transport mode
func (a *TimeAxisMapAnalyzer) getModeColor(mode string) string {
	colors := map[string]string{
		"WALK":  "#4CAF50",
		"BIKE":  "#2196F3",
		"CAR":   "#FF9800",
		"TRAIN": "#9C27B0",
		"PLANE": "#F44336",
	}
	if color, ok := colors[mode]; ok {
		return color
	}
	return "#757575"
}

// insertTimeAxisMarkers inserts time axis markers into the database
func (a *TimeAxisMapAnalyzer) insertTimeAxisMarkers(ctx context.Context, markers []TimeAxisMarker) error {
	if len(markers) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO time_axis_markers (
			marker_ts, marker_type, entity_id, entity_type,
			latitude, longitude, label, icon, color,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, marker := range markers {
		_, err := stmt.ExecContext(ctx,
			marker.MarkerTS, marker.MarkerType, marker.EntityID, marker.EntityType,
			marker.Latitude, marker.Longitude, marker.Label, marker.Icon, marker.Color,
		)
		if err != nil {
			return fmt.Errorf("failed to insert time axis marker: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[TimeAxisMapAnalyzer] Inserted %d time axis markers", len(markers))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("time_axis_map", NewTimeAxisMapAnalyzer)
}