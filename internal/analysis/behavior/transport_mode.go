package behavior

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// TransportModeAnalyzer implements transport mode classification
// Skill: 交通方式识别 (Transport Mode Classification)
// Classifies trajectory segments by transport mode based on speed
type TransportModeAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewTransportModeAnalyzer creates a new transport mode analyzer
func NewTransportModeAnalyzer(db *sql.DB) analysis.Analyzer {
	return &TransportModeAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "transport_mode", 1000),
	}
}

// Analyze performs transport mode classification
func (a *TransportModeAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[TransportModeAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing segments (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM segments"); err != nil {
			return fmt.Errorf("failed to clear segments: %w", err)
		}
		log.Printf("[TransportModeAnalyzer] Cleared existing segments")
	}

	// Get all track points ordered by time
	pointsQuery := `
		SELECT
			id,
			dataTime,
			latitude,
			longitude,
			speed,
			province,
			city,
			county,
			town,
			grid_id
		FROM "一生足迹"
		WHERE outlier_flag = 0
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, pointsQuery)
	if err != nil {
		return fmt.Errorf("failed to query points: %w", err)
	}
	defer rows.Close()

	var points []TrackPoint
	for rows.Next() {
		var point TrackPoint
		var speed sql.NullFloat64
		var province, city, county, town, gridID sql.NullString

		if err := rows.Scan(&point.ID, &point.Timestamp, &point.Lat, &point.Lon,
			&speed, &province, &city, &county, &town, &gridID); err != nil {
			return fmt.Errorf("failed to scan point: %w", err)
		}

		if speed.Valid {
			point.Speed = speed.Float64
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
		if town.Valid {
			point.Town = town.String
		}
		if gridID.Valid {
			point.GridID = gridID.String
		}

		points = append(points, point)
	}

	if len(points) == 0 {
		log.Printf("[TransportModeAnalyzer] No points to process")
		return a.MarkTaskAsCompleted(taskID, `{"segments": 0}`)
	}

	log.Printf("[TransportModeAnalyzer] Processing %d points", len(points))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(points)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Classify segments by transport mode
	segments := a.classifySegments(points)

	// Insert segments
	if err := a.insertSegments(ctx, segments); err != nil {
		return fmt.Errorf("failed to insert segments: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points": len(points),
		"segments":     len(segments),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[TransportModeAnalyzer] Analysis completed: %d points processed, %d segments created", len(points), len(segments))
	return nil
}

// TrackPoint holds track point data
type TrackPoint struct {
	ID        int64
	Timestamp int64
	Lat       float64
	Lon       float64
	Speed     float64
	Province  string
	City      string
	County    string
	Town      string
	GridID    string
}

// Segment holds segment data
type Segment struct {
	Mode      string
	StartTS   int64
	EndTS     int64
	StartLat  float64
	StartLon  float64
	EndLat    float64
	EndLon    float64
	AvgSpeed  float64
	MaxSpeed  float64
	Province  string
	City      string
	County    string
	Town      string
	GridID    string
}

// classifySegments classifies points into transport mode segments
func (a *TransportModeAnalyzer) classifySegments(points []TrackPoint) []Segment {
	if len(points) == 0 {
		return nil
	}

	var segments []Segment
	var currentSegment *Segment
	var segmentPoints []TrackPoint

	for i, point := range points {
		mode := a.classifyMode(point.Speed)

		if currentSegment == nil {
			// Start new segment
			currentSegment = &Segment{
				Mode:     mode,
				StartTS:  point.Timestamp,
				StartLat: point.Lat,
				StartLon: point.Lon,
				MaxSpeed: point.Speed,
				Province: point.Province,
				City:     point.City,
				County:   point.County,
				Town:     point.Town,
				GridID:   point.GridID,
			}
			segmentPoints = []TrackPoint{point}
		} else if mode != currentSegment.Mode || i == len(points)-1 {
			// Mode changed or last point - end current segment
			if i == len(points)-1 && mode == currentSegment.Mode {
				segmentPoints = append(segmentPoints, point)
			}

			// Finalize segment
			currentSegment.EndTS = segmentPoints[len(segmentPoints)-1].Timestamp
			currentSegment.EndLat = segmentPoints[len(segmentPoints)-1].Lat
			currentSegment.EndLon = segmentPoints[len(segmentPoints)-1].Lon

			// Calculate average speed
			totalSpeed := 0.0
			for _, p := range segmentPoints {
				totalSpeed += p.Speed
				if p.Speed > currentSegment.MaxSpeed {
					currentSegment.MaxSpeed = p.Speed
				}
			}
			currentSegment.AvgSpeed = totalSpeed / float64(len(segmentPoints))

			// Only add segments with duration > 10 seconds
			if currentSegment.EndTS-currentSegment.StartTS >= 10 {
				segments = append(segments, *currentSegment)
			}

			// Start new segment if not last point
			if i < len(points)-1 {
				currentSegment = &Segment{
					Mode:     mode,
					StartTS:  point.Timestamp,
					StartLat: point.Lat,
					StartLon: point.Lon,
					MaxSpeed: point.Speed,
					Province: point.Province,
					City:     point.City,
					County:   point.County,
					Town:     point.Town,
					GridID:   point.GridID,
				}
				segmentPoints = []TrackPoint{point}
			}
		} else {
			// Continue current segment
			segmentPoints = append(segmentPoints, point)
		}
	}

	return segments
}

// classifyMode classifies transport mode based on speed
func (a *TransportModeAnalyzer) classifyMode(speed float64) string {
	// Speed thresholds (m/s):
	// WALK: 0-2 m/s (0-7.2 km/h)
	// BIKE: 2-8 m/s (7.2-28.8 km/h)
	// CAR: 8-40 m/s (28.8-144 km/h)
	// TRAIN: 40-60 m/s (144-216 km/h)
	// PLANE: >60 m/s (>216 km/h)

	if speed < 2.0 {
		return "WALK"
	} else if speed < 8.0 {
		return "BIKE"
	} else if speed < 40.0 {
		return "CAR"
	} else if speed < 60.0 {
		return "TRAIN"
	} else {
		return "PLANE"
	}
}

// insertSegments inserts segments into the database
func (a *TransportModeAnalyzer) insertSegments(ctx context.Context, segments []Segment) error {
	if len(segments) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO segments (
			mode, start_ts, end_ts, start_lat, start_lon, end_lat, end_lon,
			avg_speed_mps, max_speed_mps, province, city, county, town, grid_id,
			outlier_flag, algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, seg := range segments {
		_, err := stmt.ExecContext(ctx,
			seg.Mode,
			seg.StartTS,
			seg.EndTS,
			seg.StartLat,
			seg.StartLon,
			seg.EndLat,
			seg.EndLon,
			seg.AvgSpeed,
			seg.MaxSpeed,
			seg.Province,
			seg.City,
			seg.County,
			seg.Town,
			seg.GridID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert segment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[TransportModeAnalyzer] Inserted %d segments", len(segments))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("transport_mode", NewTransportModeAnalyzer)
}
