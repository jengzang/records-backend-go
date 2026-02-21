package behavior

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

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
		// Delete dependent rows first to avoid foreign key constraint violations
		// Order matters: delete child tables before parent tables
		// Ignore errors for non-existent tables (they may not be created yet)
		tablesToClear := []string{"speed_events", "render_segments_cache", "road_overlap_stats"}
		for _, table := range tablesToClear {
			if _, err := a.DB.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
				// Log warning but continue if table doesn't exist
				log.Printf("[TransportModeAnalyzer] Warning: failed to clear %s: %v (table may not exist yet)", table, err)
			}
		}

		// Delete segments table (this one must succeed)
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM segments"); err != nil {
			return fmt.Errorf("failed to clear segments: %w", err)
		}
		log.Printf("[TransportModeAnalyzer] Cleared existing segments and dependent tables")
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

// TransportSegment holds segment data for transport mode classification
type TransportSegment struct {
	Mode          string
	StartTime     int64
	EndTime       int64
	StartPointID  int64
	EndPointID    int64
	PointCount    int
	DistanceM     float64
	DurationS     int64
	AvgSpeedKmh   float64
	MaxSpeedKmh   float64
	Confidence    float64
	ReasonCodes   string // JSON array
	Metadata      string // JSON object
}

// classifySegments classifies points into transport mode segments
func (a *TransportModeAnalyzer) classifySegments(points []TrackPoint) []TransportSegment {
	if len(points) == 0 {
		return nil
	}

	var segments []TransportSegment
	var currentSegment *TransportSegment
	var segmentPoints []TrackPoint

	for i, point := range points {
		mode := a.classifyMode(point.Speed)

		if currentSegment == nil {
			// Start new segment
			currentSegment = &TransportSegment{
				Mode:         mode,
				StartTime:    point.Timestamp,
				StartPointID: point.ID,
				MaxSpeedKmh:  point.Speed * 3.6, // Convert m/s to km/h
				Confidence:   0.8,                // Default confidence
			}
			segmentPoints = []TrackPoint{point}
		} else if mode != currentSegment.Mode || i == len(points)-1 {
			// Mode changed or last point - end current segment
			if i == len(points)-1 && mode == currentSegment.Mode {
				segmentPoints = append(segmentPoints, point)
			}

			// Finalize segment
			lastPoint := segmentPoints[len(segmentPoints)-1]
			currentSegment.EndTime = lastPoint.Timestamp
			currentSegment.EndPointID = lastPoint.ID
			currentSegment.PointCount = len(segmentPoints)
			currentSegment.DurationS = currentSegment.EndTime - currentSegment.StartTime

			// Calculate distance and speeds
			totalSpeed := 0.0
			totalDistance := 0.0
			for j, p := range segmentPoints {
				speedKmh := p.Speed * 3.6
				totalSpeed += speedKmh
				if speedKmh > currentSegment.MaxSpeedKmh {
					currentSegment.MaxSpeedKmh = speedKmh
				}
				// Calculate distance between consecutive points
				if j > 0 {
					prev := segmentPoints[j-1]
					dist := haversineDistance(prev.Lat, prev.Lon, p.Lat, p.Lon)
					totalDistance += dist
				}
			}
			currentSegment.AvgSpeedKmh = totalSpeed / float64(len(segmentPoints))
			currentSegment.DistanceM = totalDistance

			// Set reason codes and metadata
			currentSegment.ReasonCodes = "[]" // Empty JSON array for now
			currentSegment.Metadata = "{}"    // Empty JSON object for now

			// Only add segments with duration > 10 seconds
			if currentSegment.DurationS >= 10 {
				segments = append(segments, *currentSegment)
			}

			// Start new segment if not last point
			if i < len(points)-1 {
				currentSegment = &TransportSegment{
					Mode:         mode,
					StartTime:    point.Timestamp,
					StartPointID: point.ID,
					MaxSpeedKmh:  point.Speed * 3.6,
					Confidence:   0.8,
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
func (a *TransportModeAnalyzer) insertSegments(ctx context.Context, segments []TransportSegment) error {
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
			mode, start_time, end_time, start_point_id, end_point_id,
			point_count, distance_m, duration_s, avg_speed_kmh, max_speed_kmh,
			confidence, reason_codes, metadata, algo_version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1.0', CAST(strftime('%s', 'now') AS INTEGER), CAST(strftime('%s', 'now') AS INTEGER))
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, seg := range segments {
		_, err := stmt.ExecContext(ctx,
			seg.Mode,
			seg.StartTime,
			seg.EndTime,
			seg.StartPointID,
			seg.EndPointID,
			seg.PointCount,
			seg.DistanceM,
			seg.DurationS,
			seg.AvgSpeedKmh,
			seg.MaxSpeedKmh,
			seg.Confidence,
			seg.ReasonCodes,
			seg.Metadata,
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

// haversineDistance calculates the distance between two points in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	lat1Rad := lat1 * 3.14159265359 / 180
	lat2Rad := lat2 * 3.14159265359 / 180
	deltaLat := (lat2 - lat1) * 3.14159265359 / 180
	deltaLon := (lon2 - lon1) * 3.14159265359 / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("transport_mode", NewTransportModeAnalyzer)
}
