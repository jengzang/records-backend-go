package behavior

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// SpeedEventsAnalyzer implements speed event detection
// Skill: 速度事件检测 (Speed Events)
// Detects high-speed events from CAR segments
type SpeedEventsAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewSpeedEventsAnalyzer creates a new speed events analyzer
func NewSpeedEventsAnalyzer(db *sql.DB) analysis.Analyzer {
	return &SpeedEventsAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "speed_events", 1000),
	}
}

// Analyze performs speed event detection
func (a *SpeedEventsAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[SpeedEventsAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing speed events (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM speed_events"); err != nil {
			return fmt.Errorf("failed to clear speed events: %w", err)
		}
		log.Printf("[SpeedEventsAnalyzer] Cleared existing speed events")
	}

	// Get all CAR segments
	segmentsQuery := `
		SELECT
			id,
			start_ts,
			end_ts,
			province,
			city,
			county,
			town,
			grid_id
		FROM segments
		WHERE mode = 'CAR'
			AND outlier_flag = 0
		ORDER BY id
	`

	rows, err := a.DB.QueryContext(ctx, segmentsQuery)
	if err != nil {
		return fmt.Errorf("failed to query segments: %w", err)
	}

	var segments []SegmentInfo
	for rows.Next() {
		var seg SegmentInfo
		if err := rows.Scan(&seg.ID, &seg.StartTS, &seg.EndTS, &seg.Province, &seg.City, &seg.County, &seg.Town, &seg.GridID); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan segment: %w", err)
		}
		segments = append(segments, seg)
	}
	rows.Close()

	log.Printf("[SpeedEventsAnalyzer] Processing %d CAR segments", len(segments))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(segments)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Process each segment to detect speed events
	var speedEvents []SpeedEvent
	processed := 0

	// Default thresholds (should come from threshold_profiles)
	minEventSpeed := 33.33  // 120 km/h = 33.33 m/s
	minEventDuration := 60  // 60 seconds
	allowedGap := 10        // 10 seconds

	for _, seg := range segments {
		// Get points for this segment
		pointsQuery := `
			SELECT
				id,
				dataTime,
				latitude,
				longitude,
				speed
			FROM "一生足迹"
			WHERE dataTime BETWEEN ? AND ?
				AND outlier_flag = 0
			ORDER BY dataTime
		`

		pointRows, err := a.DB.QueryContext(ctx, pointsQuery, seg.StartTS, seg.EndTS)
		if err != nil {
			return fmt.Errorf("failed to query points for segment %d: %w", seg.ID, err)
		}

		var points []PointData
		for pointRows.Next() {
			var point PointData
			var speed sql.NullFloat64
			if err := pointRows.Scan(&point.ID, &point.Timestamp, &point.Lat, &point.Lon, &speed); err != nil {
				pointRows.Close()
				return fmt.Errorf("failed to scan point: %w", err)
			}
			if speed.Valid {
				point.Speed = speed.Float64
			}
			points = append(points, point)
		}
		pointRows.Close()

		if len(points) == 0 {
			continue
		}

		// Detect speed events using state machine
		events := a.detectSpeedEvents(seg, points, minEventSpeed, float64(minEventDuration), float64(allowedGap))
		speedEvents = append(speedEvents, events...)

		processed++
		if processed%100 == 0 {
			if err := a.UpdateTaskProgress(taskID, int64(len(segments)), int64(processed), 0); err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}
			log.Printf("[SpeedEventsAnalyzer] Processed %d/%d segments", processed, len(segments))
		}
	}

	// Insert speed events
	if err := a.insertSpeedEvents(ctx, speedEvents); err != nil {
		return fmt.Errorf("failed to insert speed events: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_segments": len(segments),
		"processed_segments": processed,
		"speed_events": len(speedEvents),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[SpeedEventsAnalyzer] Analysis completed: %d segments processed, %d speed events found", processed, len(speedEvents))
	return nil
}

// SegmentInfo holds segment information
type SegmentInfo struct {
	ID       int64
	StartTS  int64
	EndTS    int64
	Province sql.NullString
	City     sql.NullString
	County   sql.NullString
	Town     sql.NullString
	GridID   sql.NullString
}

// PointData holds point data
type PointData struct {
	ID        int64
	Timestamp int64
	Lat       float64
	Lon       float64
	Speed     float64
}

// SpeedEvent holds speed event data
type SpeedEvent struct {
	SegmentID  int64
	StartTS    int64
	EndTS      int64
	DurationS  int64
	MaxSpeed   float64
	AvgSpeed   float64
	PeakTS     int64
	PeakLat    float64
	PeakLon    float64
	Province   string
	City       string
	County     string
	Town       string
	GridID     string
	Confidence float64
	Reasons    []string
}

// detectSpeedEvents detects speed events using state machine
func (a *SpeedEventsAnalyzer) detectSpeedEvents(seg SegmentInfo, points []PointData, minSpeed, minDuration, allowedGap float64) []SpeedEvent {
	var events []SpeedEvent

	if len(points) == 0 {
		return events
	}

	var currentEvent *SpeedEvent
	var eventPoints []PointData
	lastHighSpeedTS := int64(0)

	for _, point := range points {
		if point.Speed >= minSpeed {
			// High speed point
			if currentEvent == nil {
				// Start new event
				currentEvent = &SpeedEvent{
					SegmentID: seg.ID,
					StartTS:   point.Timestamp,
					MaxSpeed:  point.Speed,
					PeakTS:    point.Timestamp,
					PeakLat:   point.Lat,
					PeakLon:   point.Lon,
				}
				eventPoints = []PointData{point}
			} else {
				// Continue event
				eventPoints = append(eventPoints, point)
				if point.Speed > currentEvent.MaxSpeed {
					currentEvent.MaxSpeed = point.Speed
					currentEvent.PeakTS = point.Timestamp
					currentEvent.PeakLat = point.Lat
					currentEvent.PeakLon = point.Lon
				}
			}
			lastHighSpeedTS = point.Timestamp
		} else {
			// Low speed point
			if currentEvent != nil {
				gap := point.Timestamp - lastHighSpeedTS
				if float64(gap) > allowedGap {
					// End current event
					currentEvent.EndTS = lastHighSpeedTS
					currentEvent.DurationS = currentEvent.EndTS - currentEvent.StartTS

					if float64(currentEvent.DurationS) >= minDuration {
						// Calculate average speed
						totalSpeed := 0.0
						for _, p := range eventPoints {
							totalSpeed += p.Speed
						}
						currentEvent.AvgSpeed = totalSpeed / float64(len(eventPoints))

						// Set location info
						if seg.Province.Valid {
							currentEvent.Province = seg.Province.String
						}
						if seg.City.Valid {
							currentEvent.City = seg.City.String
						}
						if seg.County.Valid {
							currentEvent.County = seg.County.String
						}
						if seg.Town.Valid {
							currentEvent.Town = seg.Town.String
						}
						if seg.GridID.Valid {
							currentEvent.GridID = seg.GridID.String
						}

						// Calculate confidence
						currentEvent.Confidence = a.calculateConfidence(currentEvent, eventPoints)
						currentEvent.Reasons = a.generateReasons(currentEvent, eventPoints)

						events = append(events, *currentEvent)
					}

					// Reset
					currentEvent = nil
					eventPoints = nil
				}
			}
		}
	}

	// Handle event at end of segment
	if currentEvent != nil {
		currentEvent.EndTS = lastHighSpeedTS
		currentEvent.DurationS = currentEvent.EndTS - currentEvent.StartTS

		if float64(currentEvent.DurationS) >= minDuration {
			totalSpeed := 0.0
			for _, p := range eventPoints {
				totalSpeed += p.Speed
			}
			currentEvent.AvgSpeed = totalSpeed / float64(len(eventPoints))

			if seg.Province.Valid {
				currentEvent.Province = seg.Province.String
			}
			if seg.City.Valid {
				currentEvent.City = seg.City.String
			}
			if seg.County.Valid {
				currentEvent.County = seg.County.String
			}
			if seg.Town.Valid {
				currentEvent.Town = seg.Town.String
			}
			if seg.GridID.Valid {
				currentEvent.GridID = seg.GridID.String
			}

			currentEvent.Confidence = a.calculateConfidence(currentEvent, eventPoints)
			currentEvent.Reasons = a.generateReasons(currentEvent, eventPoints)

			events = append(events, *currentEvent)
		}
	}

	return events
}

// calculateConfidence calculates confidence score for speed event
func (a *SpeedEventsAnalyzer) calculateConfidence(event *SpeedEvent, points []PointData) float64 {
	confidence := 1.0

	// Reduce confidence if duration is short
	if event.DurationS < 120 {
		confidence *= 0.8
	}

	// Reduce confidence if few points
	if len(points) < 5 {
		confidence *= 0.7
	}

	// Reduce confidence if max speed is not much higher than threshold
	if event.MaxSpeed < 40 { // 144 km/h
		confidence *= 0.9
	}

	return confidence
}

// generateReasons generates reason codes for speed event
func (a *SpeedEventsAnalyzer) generateReasons(event *SpeedEvent, points []PointData) []string {
	var reasons []string

	if event.MaxSpeed >= 50 { // 180 km/h
		reasons = append(reasons, "VERY_HIGH_SPEED")
	} else if event.MaxSpeed >= 40 { // 144 km/h
		reasons = append(reasons, "HIGH_SPEED")
	} else {
		reasons = append(reasons, "MODERATE_SPEED")
	}

	if event.DurationS >= 300 {
		reasons = append(reasons, "LONG_DURATION")
	} else if event.DurationS >= 120 {
		reasons = append(reasons, "MEDIUM_DURATION")
	} else {
		reasons = append(reasons, "SHORT_DURATION")
	}

	if len(points) >= 10 {
		reasons = append(reasons, "MANY_POINTS")
	} else if len(points) >= 5 {
		reasons = append(reasons, "MODERATE_POINTS")
	} else {
		reasons = append(reasons, "FEW_POINTS")
	}

	return reasons
}

// insertSpeedEvents inserts speed events into the database
func (a *SpeedEventsAnalyzer) insertSpeedEvents(ctx context.Context, events []SpeedEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO speed_events (
			segment_id, start_ts, end_ts, duration_s, max_speed_mps, avg_speed_mps,
			peak_ts, peak_lat, peak_lon, province, city, county, town, grid_id,
			confidence, reason_codes, algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		reasonsJSON, _ := json.Marshal(event.Reasons)

		_, err := stmt.ExecContext(ctx,
			event.SegmentID,
			event.StartTS,
			event.EndTS,
			event.DurationS,
			event.MaxSpeed,
			event.AvgSpeed,
			event.PeakTS,
			event.PeakLat,
			event.PeakLon,
			event.Province,
			event.City,
			event.County,
			event.Town,
			event.GridID,
			event.Confidence,
			string(reasonsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert speed event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[SpeedEventsAnalyzer] Inserted %d speed events", len(events))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("speed_events", NewSpeedEventsAnalyzer)
}
