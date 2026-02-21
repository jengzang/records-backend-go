package foundation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// OutlierPoint represents a GPS point for outlier detection
type OutlierPoint struct {
	ID        int64
	Timestamp int64
	Lat       float64
	Lon       float64
	Speed     float64
	Accuracy  float64
}

// OutlierResult represents the result of outlier detection for a point
type OutlierResult struct {
	ID          int64
	IsOutlier   bool
	Reasons     []string
	QAStatus    string
}

// OutlierThresholds defines configurable thresholds for outlier detection
type OutlierThresholds struct {
	MaxSpeedMPS        float64 // 277.78 m/s (1000 km/h)
	MaxAccuracyM       float64 // 100 m
	JumpDistanceM      float64 // 1000 m
	JumpTimeS          int64   // 10 s
	BacktrackRadiusM   float64 // 50 m
	StaticDriftRadiusM float64 // 30 m
}

// DefaultThresholds provides default outlier detection thresholds
var DefaultThresholds = OutlierThresholds{
	MaxSpeedMPS:        277.78, // 1000 km/h (covers all commercial flights, max observed 970 km/h)
	MaxAccuracyM:       100.0,  // 100 meters
	JumpDistanceM:      1000.0, // 1 km
	JumpTimeS:          10,     // 10 seconds
	BacktrackRadiusM:   20.0,   // 20 meters (tightened from 50m)
	StaticDriftRadiusM: 50.0,   // 50 meters (increased from 30m)
}

// TrajectoryPoint represents a GPS point for trajectory completion
type TrajectoryPoint struct {
	ID        int64
	Timestamp int64
	Lat       float64
	Lon       float64
	Altitude  float64
	Speed     float64
}

// OutlierDetectionAnalyzer implements outlier detection
// Skill: 异常点检测 (Outlier Detection)
// Detects outliers using rule-based methods
type OutlierDetectionAnalyzer struct {
	*analysis.IncrementalAnalyzer
	Thresholds OutlierThresholds
}

// NewOutlierDetectionAnalyzer creates a new outlier detection analyzer
func NewOutlierDetectionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &OutlierDetectionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "outlier_detection", 10000),
		Thresholds:          DefaultThresholds,
	}
}

// Analyze performs outlier detection
func (a *OutlierDetectionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[OutlierDetectionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Reset outlier flags and reason codes (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "UPDATE \"一生足迹\" SET outlier_flag = 0, outlier_reason_codes = NULL, qa_status = NULL"); err != nil {
			return fmt.Errorf("failed to reset outlier flags: %w", err)
		}
		log.Printf("[OutlierDetectionAnalyzer] Reset outlier flags and reason codes")
	}

	// Get all track points with necessary fields for rule-based detection
	pointsQuery := `
		SELECT
			id,
			dataTime,
			latitude,
			longitude,
			speed,
			accuracy
		FROM "一生足迹"
		ORDER BY dataTime, id
	`

	rows, err := a.DB.QueryContext(ctx, pointsQuery)
	if err != nil {
		return fmt.Errorf("failed to query points: %w", err)
	}
	defer rows.Close()

	var points []OutlierPoint
	for rows.Next() {
		var point OutlierPoint
		var timestamp sql.NullInt64
		var lat, lon, speed, accuracy sql.NullFloat64

		if err := rows.Scan(&point.ID, &timestamp, &lat, &lon, &speed, &accuracy); err != nil {
			return fmt.Errorf("failed to scan point: %w", err)
		}

		if timestamp.Valid {
			point.Timestamp = timestamp.Int64
		}
		if lat.Valid {
			point.Lat = lat.Float64
		}
		if lon.Valid {
			point.Lon = lon.Float64
		}
		if speed.Valid {
			point.Speed = speed.Float64
		}
		if accuracy.Valid {
			point.Accuracy = accuracy.Float64
		}

		points = append(points, point)
	}

	if len(points) == 0 {
		log.Printf("[OutlierDetectionAnalyzer] No points to process")
		return a.MarkTaskAsCompleted(taskID, `{"outliers": 0}`)
	}

	log.Printf("[OutlierDetectionAnalyzer] Processing %d points", len(points))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(points)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Detect outliers with rule-based methods
	outlierResults := a.detectOutliers(points)

	// Update outlier flags and reason codes
	if err := a.updateOutlierResults(ctx, outlierResults); err != nil {
		return fmt.Errorf("failed to update outlier results: %w", err)
	}

	// Count outliers
	outlierCount := 0
	for _, result := range outlierResults {
		if result.IsOutlier {
			outlierCount++
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points": len(points),
		"outliers":     outlierCount,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[OutlierDetectionAnalyzer] Analysis completed: %d points processed, %d outliers detected", len(points), outlierCount)
	return nil
}

// detectOutliers detects outliers using rule-based methods
func (a *OutlierDetectionAnalyzer) detectOutliers(points []OutlierPoint) []OutlierResult {
	results := make([]OutlierResult, len(points))

	for i, point := range points {
		var reasons []string
		qaStatus := "PASS"

		// Rule 1: EXCESSIVE_SPEED - Speed > 1000 km/h (277.78 m/s)
		if point.Speed > a.Thresholds.MaxSpeedMPS {
			reasons = append(reasons, "EXCESSIVE_SPEED")
		}

		// Rule 2: LOW_ACCURACY - Accuracy > 100m
		if point.Accuracy > a.Thresholds.MaxAccuracyM {
			reasons = append(reasons, "LOW_ACCURACY")
		}

		// Rule 3: JUMP Detection (Teleportation)
		// Check if point jumps impossibly far in short time
		if i > 0 {
			prevPoint := points[i-1]
			timeDelta := point.Timestamp - prevPoint.Timestamp
			distance := haversineDistance(prevPoint.Lat, prevPoint.Lon, point.Lat, point.Lon)

			if timeDelta > 0 && timeDelta <= a.Thresholds.JumpTimeS && distance >= a.Thresholds.JumpDistanceM {
				reasons = append(reasons, "JUMP")
			}
		}

		// Rule 4: BACKTRACK Detection - DISABLED (too many false positives)
		// Rule 5: STATIC_DRIFT Detection - DISABLED (too many false positives)

		// Determine QA status
		if len(reasons) > 0 {
			qaStatus = "FAIL"
		} else if point.Accuracy >= 50 && point.Accuracy <= a.Thresholds.MaxAccuracyM {
			qaStatus = "WARNING" // Moderate accuracy, not an outlier but low quality
		}

		results[i] = OutlierResult{
			ID:        point.ID,
			IsOutlier: len(reasons) > 0,
			Reasons:   reasons,
			QAStatus:  qaStatus,
		}
	}

	return results
}

// haversineDistance calculates the distance between two GPS coordinates in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// isStaticDrift detects GPS drift in a window of points
// Returns true if points cluster but have high variance (indicating drift while stationary)
func isStaticDrift(window []OutlierPoint, radiusM float64) bool {
	if len(window) < 5 {
		return false
	}

	// Calculate centroid
	var sumLat, sumLon float64
	for _, p := range window {
		sumLat += p.Lat
		sumLon += p.Lon
	}
	centroidLat := sumLat / float64(len(window))
	centroidLon := sumLon / float64(len(window))

	// Check if all points are within radius of centroid
	allWithinRadius := true
	for _, p := range window {
		dist := haversineDistance(centroidLat, centroidLon, p.Lat, p.Lon)
		if dist > radiusM {
			allWithinRadius = false
			break
		}
	}

	// If all points cluster within radius, check for variance
	if allWithinRadius {
		// Calculate variance
		var sumSqDist float64
		for _, p := range window {
			dist := haversineDistance(centroidLat, centroidLon, p.Lat, p.Lon)
			sumSqDist += dist * dist
		}
		variance := sumSqDist / float64(len(window))

		// If variance is high (>400m²), it's likely drift
		// This means standard deviation > 20m, which is significant for stationary points
		return variance > 400 // 20m * 20m
	}

	return false
}

// updateOutlierResults updates outlier flags, reason codes, and QA status in the database
func (a *OutlierDetectionAnalyzer) updateOutlierResults(ctx context.Context, results []OutlierResult) error {
	if len(results) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `UPDATE "一生足迹" SET outlier_flag = ?, outlier_reason_codes = ?, qa_status = ? WHERE id = ?`

	stmt, err := tx.PrepareContext(ctx, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, result := range results {
		outlierFlag := 0
		if result.IsOutlier {
			outlierFlag = 1
		}

		var reasonCodesJSON string
		if len(result.Reasons) > 0 {
			reasonBytes, _ := json.Marshal(result.Reasons)
			reasonCodesJSON = string(reasonBytes)
		}

		if _, err := stmt.ExecContext(ctx, outlierFlag, reasonCodesJSON, result.QAStatus, result.ID); err != nil {
			return fmt.Errorf("failed to update outlier result for id %d: %w", result.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	outlierCount := 0
	for _, result := range results {
		if result.IsOutlier {
			outlierCount++
		}
	}

	log.Printf("[OutlierDetectionAnalyzer] Updated %d outlier flags (%d outliers)", len(results), outlierCount)
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("outlier_detection", NewOutlierDetectionAnalyzer)
}
