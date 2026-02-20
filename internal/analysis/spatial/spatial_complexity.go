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

// SpatialComplexityAnalyzer implements spatial complexity metrics
// Skill: 空间复杂度分析 (Spatial Complexity)
// Calculates trajectory complexity, entropy, and tortuosity
type SpatialComplexityAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewSpatialComplexityAnalyzer creates a new spatial complexity analyzer
func NewSpatialComplexityAnalyzer(db *sql.DB) analysis.Analyzer {
	return &SpatialComplexityAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "spatial_complexity", 10000),
	}
}

// Analyze performs spatial complexity analysis
func (a *SpatialComplexityAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[SpatialComplexityAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing metrics (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM complexity_metrics"); err != nil {
			return fmt.Errorf("failed to clear complexity_metrics: %w", err)
		}
		log.Printf("[SpatialComplexityAnalyzer] Cleared existing complexity metrics")
	}

	// Compute all-time complexity metrics
	metrics, err := a.computeComplexityMetrics(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to compute complexity metrics: %w", err)
	}

	// Insert metrics
	if err := a.insertComplexityMetrics(ctx, metrics); err != nil {
		return fmt.Errorf("failed to insert complexity metrics: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"trajectory_complexity": metrics.TrajectoryComplexity,
		"direction_changes":     metrics.DirectionChanges,
		"spatial_entropy":       metrics.SpatialEntropy,
		"path_efficiency":       metrics.PathEfficiency,
		"tortuosity":            metrics.Tortuosity,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[SpatialComplexityAnalyzer] Analysis completed")
	return nil
}

// ComplexityMetrics holds spatial complexity metrics
type ComplexityMetrics struct {
	MetricDate           string
	TrajectoryComplexity float64
	DirectionChanges     int64
	AvgTurnAngle         float64
	SpatialEntropy       float64
	PathEfficiency       float64
	Tortuosity           float64
}

// computeComplexityMetrics computes complexity metrics
func (a *SpatialComplexityAnalyzer) computeComplexityMetrics(ctx context.Context, date string) (*ComplexityMetrics, error) {
	// Query track points with heading and distance
	query := `
		SELECT
			latitude, longitude, heading, distance
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND heading IS NOT NULL
			AND distance IS NOT NULL
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query track points: %w", err)
	}
	defer rows.Close()

	var points []ComplexityPoint
	for rows.Next() {
		var point ComplexityPoint
		if err := rows.Scan(&point.Latitude, &point.Longitude, &point.Heading, &point.Distance); err != nil {
			return nil, fmt.Errorf("failed to scan track point: %w", err)
		}
		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(points) < 2 {
		return &ComplexityMetrics{MetricDate: date}, nil
	}

	// Calculate metrics
	metrics := &ComplexityMetrics{MetricDate: date}

	// 1. Direction changes and turn angles
	var turnAngles []float64
	for i := 1; i < len(points); i++ {
		prevHeading := points[i-1].Heading
		currHeading := points[i].Heading

		// Calculate turn angle
		turnAngle := math.Abs(currHeading - prevHeading)
		if turnAngle > 180 {
			turnAngle = 360 - turnAngle
		}

		// Count as direction change if turn > 15 degrees
		if turnAngle > 15 {
			metrics.DirectionChanges++
			turnAngles = append(turnAngles, turnAngle)
		}
	}

	// Average turn angle
	if len(turnAngles) > 0 {
		sum := 0.0
		for _, angle := range turnAngles {
			sum += angle
		}
		metrics.AvgTurnAngle = sum / float64(len(turnAngles))
	}

	// 2. Path efficiency (actual distance / straight-line distance)
	actualDistance := 0.0
	for _, point := range points {
		actualDistance += point.Distance
	}

	firstPoint := points[0]
	lastPoint := points[len(points)-1]
	straightLineDistance := haversineDistance(
		firstPoint.Latitude, firstPoint.Longitude,
		lastPoint.Latitude, lastPoint.Longitude,
	)

	if straightLineDistance > 0 {
		metrics.PathEfficiency = straightLineDistance / actualDistance
	}

	// 3. Tortuosity (inverse of path efficiency)
	if metrics.PathEfficiency > 0 {
		metrics.Tortuosity = 1.0 / metrics.PathEfficiency
	}

	// 4. Spatial entropy (based on grid cell distribution)
	entropy, err := a.calculateSpatialEntropy(ctx)
	if err != nil {
		log.Printf("[SpatialComplexityAnalyzer] Warning: failed to calculate spatial entropy: %v", err)
	} else {
		metrics.SpatialEntropy = entropy
	}

	// 5. Trajectory complexity score (0-1, normalized)
	// Combines direction changes, tortuosity, and entropy
	complexityScore := 0.0

	// Normalize direction changes (assume max 1000 changes)
	directionScore := math.Min(float64(metrics.DirectionChanges)/1000.0, 1.0)

	// Normalize tortuosity (assume max 5.0)
	tortuosityScore := math.Min(metrics.Tortuosity/5.0, 1.0)

	// Normalize entropy (assume max 10.0)
	entropyScore := math.Min(metrics.SpatialEntropy/10.0, 1.0)

	// Weighted average
	complexityScore = (directionScore*0.4 + tortuosityScore*0.3 + entropyScore*0.3)
	metrics.TrajectoryComplexity = complexityScore

	return metrics, nil
}

// ComplexityPoint holds point data for complexity analysis
type ComplexityPoint struct {
	Latitude  float64
	Longitude float64
	Heading   float64
	Distance  float64
}

// calculateSpatialEntropy calculates spatial entropy based on grid cell distribution
func (a *SpatialComplexityAnalyzer) calculateSpatialEntropy(ctx context.Context) (float64, error) {
	query := `
		SELECT visit_count
		FROM grid_cells
		WHERE visit_count > 0
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to query grid cells: %w", err)
	}
	defer rows.Close()

	var visitCounts []int64
	totalVisits := int64(0)

	for rows.Next() {
		var count int64
		if err := rows.Scan(&count); err != nil {
			return 0, fmt.Errorf("failed to scan visit count: %w", err)
		}
		visitCounts = append(visitCounts, count)
		totalVisits += count
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating rows: %w", err)
	}

	if totalVisits == 0 {
		return 0, nil
	}

	// Calculate Shannon entropy
	entropy := 0.0
	for _, count := range visitCounts {
		if count > 0 {
			p := float64(count) / float64(totalVisits)
			entropy -= p * math.Log2(p)
		}
	}

	return entropy, nil
}

// insertComplexityMetrics inserts complexity metrics into the database
func (a *SpatialComplexityAnalyzer) insertComplexityMetrics(ctx context.Context, metrics *ComplexityMetrics) error {
	insertQuery := `
		INSERT INTO complexity_metrics (
			metric_date, trajectory_complexity, direction_changes,
			avg_turn_angle, spatial_entropy, path_efficiency, tortuosity,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	_, err := a.DB.ExecContext(ctx, insertQuery,
		metrics.MetricDate, metrics.TrajectoryComplexity, metrics.DirectionChanges,
		metrics.AvgTurnAngle, metrics.SpatialEntropy, metrics.PathEfficiency, metrics.Tortuosity,
	)
	if err != nil {
		return fmt.Errorf("failed to insert complexity metrics: %w", err)
	}

	log.Printf("[SpatialComplexityAnalyzer] Inserted complexity metrics")
	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("spatial_complexity", NewSpatialComplexityAnalyzer)
}