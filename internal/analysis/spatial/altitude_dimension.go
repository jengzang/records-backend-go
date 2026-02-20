package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// AltitudeDimensionAnalyzer implements altitude/elevation analysis
// Skill: 海拔维度分析 (Altitude Dimension)
// Analyzes elevation patterns and detects climbs/descents
type AltitudeDimensionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewAltitudeDimensionAnalyzer creates a new altitude dimension analyzer
func NewAltitudeDimensionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &AltitudeDimensionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "altitude_dimension", 10000),
	}
}

// Analyze performs altitude analysis
func (a *AltitudeDimensionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[AltitudeDimensionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing events (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM altitude_events"); err != nil {
			return fmt.Errorf("failed to clear altitude_events: %w", err)
		}
		log.Printf("[AltitudeDimensionAnalyzer] Cleared existing altitude events")
	}

	// Query track points with altitude data
	query := `
		SELECT
			id, dataTime, latitude, longitude, altitude, distance,
			province, city, county
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND altitude IS NOT NULL
			AND altitude > -500
			AND altitude < 9000
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query track points: %w", err)
	}
	defer rows.Close()

	// Process points and detect altitude events
	var events []AltitudeEvent
	var currentEvent *AltitudeEvent
	var prevPoint *AltitudePoint
	totalPoints := 0

	// Thresholds
	const minAltitudeChange = 50.0  // meters
	const minDuration = 300          // 5 minutes
	const plateauThreshold = 10.0    // meters

	for rows.Next() {
		var point AltitudePoint
		var distance sql.NullFloat64
		var province, city, county sql.NullString

		if err := rows.Scan(
			&point.ID, &point.DataTime, &point.Latitude, &point.Longitude,
			&point.Altitude, &distance,
			&province, &city, &county,
		); err != nil {
			return fmt.Errorf("failed to scan track point: %w", err)
		}

		if distance.Valid {
			point.Distance = distance.Float64
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

		totalPoints++

		if prevPoint != nil {
			altChange := point.Altitude - prevPoint.Altitude
			_ = point.DataTime - prevPoint.DataTime // duration unused for now

			// Detect event start
			if currentEvent == nil {
				if math.Abs(altChange) > plateauThreshold {
					eventType := "CLIMB"
					if altChange < 0 {
						eventType = "DESCENT"
					}

					currentEvent = &AltitudeEvent{
						EventType:     eventType,
						StartTS:       prevPoint.DataTime,
						StartAltitude: prevPoint.Altitude,
						StartLat:      prevPoint.Latitude,
						StartLon:      prevPoint.Longitude,
						Province:      prevPoint.Province,
						City:          prevPoint.City,
						County:        prevPoint.County,
						TotalDistance: 0,
					}
				}
			} else {
				// Continue or end current event
				expectedType := "CLIMB"
				if altChange < 0 {
					expectedType = "DESCENT"
				}

				// Check if event continues
				if expectedType == currentEvent.EventType && math.Abs(altChange) > plateauThreshold {
					// Continue event
					currentEvent.TotalDistance += point.Distance
				} else {
					// End current event
					currentEvent.EndTS = prevPoint.DataTime
					currentEvent.EndAltitude = prevPoint.Altitude
					currentEvent.Duration = currentEvent.EndTS - currentEvent.StartTS
					currentEvent.AltitudeChange = currentEvent.EndAltitude - currentEvent.StartAltitude

					// Calculate grade
					if currentEvent.TotalDistance > 0 {
						currentEvent.AvgGrade = (currentEvent.AltitudeChange / currentEvent.TotalDistance) * 100
					}

					// Only keep significant events
					if math.Abs(currentEvent.AltitudeChange) >= minAltitudeChange &&
						currentEvent.Duration >= minDuration {
						events = append(events, *currentEvent)
					}

					// Start new event if needed
					if math.Abs(altChange) > plateauThreshold {
						currentEvent = &AltitudeEvent{
							EventType:     expectedType,
							StartTS:       prevPoint.DataTime,
							StartAltitude: prevPoint.Altitude,
							StartLat:      prevPoint.Latitude,
							StartLon:      prevPoint.Longitude,
							Province:      prevPoint.Province,
							City:          prevPoint.City,
							County:        prevPoint.County,
							TotalDistance: point.Distance,
						}
					} else {
						currentEvent = nil
					}
				}
			}
		}

		prevPoint = &point
	}

	// Handle last event
	if currentEvent != nil && prevPoint != nil {
		currentEvent.EndTS = prevPoint.DataTime
		currentEvent.EndAltitude = prevPoint.Altitude
		currentEvent.Duration = currentEvent.EndTS - currentEvent.StartTS
		currentEvent.AltitudeChange = currentEvent.EndAltitude - currentEvent.StartAltitude

		if currentEvent.TotalDistance > 0 {
			currentEvent.AvgGrade = (currentEvent.AltitudeChange / currentEvent.TotalDistance) * 100
		}

		if math.Abs(currentEvent.AltitudeChange) >= minAltitudeChange &&
			currentEvent.Duration >= minDuration {
			events = append(events, *currentEvent)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("[AltitudeDimensionAnalyzer] Processed %d points, detected %d altitude events", totalPoints, len(events))

	// Update task progress
	if err := a.UpdateTaskProgress(taskID, int64(totalPoints), int64(totalPoints), 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Insert events
	if err := a.insertAltitudeEvents(ctx, events); err != nil {
		return fmt.Errorf("failed to insert altitude events: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points": totalPoints,
		"events":       len(events),
		"climbs":       a.countEventsByType(events, "CLIMB"),
		"descents":     a.countEventsByType(events, "DESCENT"),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[AltitudeDimensionAnalyzer] Analysis completed: %d altitude events detected", len(events))
	return nil
}

// AltitudePoint holds altitude point data
type AltitudePoint struct {
	ID        int64
	DataTime  int64
	Latitude  float64
	Longitude float64
	Altitude  float64
	Distance  float64
	Province  string
	City      string
	County    string
}

// AltitudeEvent holds altitude event data
type AltitudeEvent struct {
	EventType      string
	StartTS        int64
	EndTS          int64
	StartAltitude  float64
	EndAltitude    float64
	AltitudeChange float64
	Duration       int64
	AvgGrade       float64
	MaxGrade       float64
	TotalDistance  float64
	StartLat       float64
	StartLon       float64
	Province       string
	City           string
	County         string
}

// insertAltitudeEvents inserts altitude events into the database
func (a *AltitudeDimensionAnalyzer) insertAltitudeEvents(ctx context.Context, events []AltitudeEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO altitude_events (
			event_type, start_ts, end_ts,
			start_altitude, end_altitude, altitude_change,
			duration_s, avg_grade, distance_m,
			province, city, county,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		_, err := stmt.ExecContext(ctx,
			event.EventType, event.StartTS, event.EndTS,
			event.StartAltitude, event.EndAltitude, event.AltitudeChange,
			event.Duration, event.AvgGrade, event.TotalDistance,
			event.Province, event.City, event.County,
		)
		if err != nil {
			return fmt.Errorf("failed to insert altitude event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[AltitudeDimensionAnalyzer] Inserted %d altitude events", len(events))
	return nil
}

// countEventsByType counts events by type
func (a *AltitudeDimensionAnalyzer) countEventsByType(events []AltitudeEvent, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("altitude_dimension", NewAltitudeDimensionAnalyzer)
}