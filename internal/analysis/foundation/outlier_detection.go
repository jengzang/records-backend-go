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

// OutlierDetectionAnalyzer implements outlier detection
// Skill: 异常点检测 (Outlier Detection)
// Detects outliers using Z-score and IQR methods
type OutlierDetectionAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewOutlierDetectionAnalyzer creates a new outlier detection analyzer
func NewOutlierDetectionAnalyzer(db *sql.DB) analysis.Analyzer {
	return &OutlierDetectionAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "outlier_detection", 10000),
	}
}

// Analyze performs outlier detection
func (a *OutlierDetectionAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[OutlierDetectionAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Reset outlier flags (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "UPDATE \"一生足迹\" SET outlier_flag = 0"); err != nil {
			return fmt.Errorf("failed to reset outlier flags: %w", err)
		}
		log.Printf("[OutlierDetectionAnalyzer] Reset outlier flags")
	}

	// Get all track points
	pointsQuery := `
		SELECT
			id,
			speed,
			accuracy
		FROM "一生足迹"
		ORDER BY id
	`

	rows, err := a.DB.QueryContext(ctx, pointsQuery)
	if err != nil {
		return fmt.Errorf("failed to query points: %w", err)
	}
	defer rows.Close()

	type Point struct {
		ID       int64
		Speed    float64
		Accuracy float64
	}

	var points []Point
	for rows.Next() {
		var point Point
		var speed, accuracy sql.NullFloat64

		if err := rows.Scan(&point.ID, &speed, &accuracy); err != nil {
			return fmt.Errorf("failed to scan point: %w", err)
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

	// Detect outliers
	outlierIDs := a.detectOutliers(points)

	// Update outlier flags
	if err := a.updateOutlierFlags(ctx, outlierIDs); err != nil {
		return fmt.Errorf("failed to update outlier flags: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points": len(points),
		"outliers":     len(outlierIDs),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[OutlierDetectionAnalyzer] Analysis completed: %d points processed, %d outliers detected", len(points), len(outlierIDs))
	return nil
}

// detectOutliers detects outliers using multiple methods
func (a *OutlierDetectionAnalyzer) detectOutliers(points []Point) []int64 {
	outlierMap := make(map[int64]bool)

	// Method 1: Speed outliers (>200 km/h = 55.56 m/s)
	maxSpeed := 55.56
	for _, point := range points {
		if point.Speed > maxSpeed {
			outlierMap[point.ID] = true
		}
	}

	// Method 2: Accuracy outliers (>1000m)
	maxAccuracy := 1000.0
	for _, point := range points {
		if point.Accuracy > maxAccuracy {
			outlierMap[point.ID] = true
		}
	}

	// Method 3: Z-score method for speed
	speedOutliers := a.detectZScoreOutliers(points, func(p Point) float64 { return p.Speed })
	for _, id := range speedOutliers {
		outlierMap[id] = true
	}

	// Method 4: IQR method for speed
	speedIQROutliers := a.detectIQROutliers(points, func(p Point) float64 { return p.Speed })
	for _, id := range speedIQROutliers {
		outlierMap[id] = true
	}

	// Convert map to slice
	var outlierIDs []int64
	for id := range outlierMap {
		outlierIDs = append(outlierIDs, id)
	}

	return outlierIDs
}

// detectZScoreOutliers detects outliers using Z-score method (|z| > 3)
func (a *OutlierDetectionAnalyzer) detectZScoreOutliers(points []Point, getValue func(Point) float64) []int64 {
	if len(points) == 0 {
		return nil
	}

	// Calculate mean
	sum := 0.0
	for _, point := range points {
		sum += getValue(point)
	}
	mean := sum / float64(len(points))

	// Calculate standard deviation
	sumSquares := 0.0
	for _, point := range points {
		diff := getValue(point) - mean
		sumSquares += diff * diff
	}
	stddev := math.Sqrt(sumSquares / float64(len(points)))

	// Detect outliers (|z| > 3)
	var outlierIDs []int64
	threshold := 3.0
	for _, point := range points {
		if stddev > 0 {
			z := math.Abs((getValue(point) - mean) / stddev)
			if z > threshold {
				outlierIDs = append(outlierIDs, point.ID)
			}
		}
	}

	return outlierIDs
}

// detectIQROutliers detects outliers using IQR method
func (a *OutlierDetectionAnalyzer) detectIQROutliers(points []Point, getValue func(Point) float64) []int64 {
	if len(points) == 0 {
		return nil
	}

	// Extract values and sort
	values := make([]float64, len(points))
	for i, point := range points {
		values[i] = getValue(point)
	}

	// Sort values
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	// Calculate Q1, Q3, IQR
	q1 := a.percentile(values, 25)
	q3 := a.percentile(values, 75)
	iqr := q3 - q1

	// Calculate bounds
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	// Detect outliers
	var outlierIDs []int64
	for _, point := range points {
		value := getValue(point)
		if value < lowerBound || value > upperBound {
			outlierIDs = append(outlierIDs, point.ID)
		}
	}

	return outlierIDs
}

// percentile calculates the percentile of sorted values
func (a *OutlierDetectionAnalyzer) percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := (p / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

// updateOutlierFlags updates outlier flags in the database
func (a *OutlierDetectionAnalyzer) updateOutlierFlags(ctx context.Context, outlierIDs []int64) error {
	if len(outlierIDs) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `UPDATE "一生足迹" SET outlier_flag = 1 WHERE id = ?`

	stmt, err := tx.PrepareContext(ctx, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, id := range outlierIDs {
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			return fmt.Errorf("failed to update outlier flag for id %d: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[OutlierDetectionAnalyzer] Updated %d outlier flags", len(outlierIDs))
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("outlier_detection", NewOutlierDetectionAnalyzer)
}
