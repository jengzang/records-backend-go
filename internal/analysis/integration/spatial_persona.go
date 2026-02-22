package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// SpatialPersonaAnalyzer implements spatial persona engine
// Skill: 空间画像引擎 (Spatial Persona Engine)
// Integrates all spatial analysis features into a comprehensive profile
type SpatialPersonaAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewSpatialPersonaAnalyzer creates a new spatial persona analyzer
func NewSpatialPersonaAnalyzer(db *sql.DB) analysis.Analyzer {
	return &SpatialPersonaAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "spatial_persona", 10000),
	}
}

// Analyze performs spatial persona analysis
func (a *SpatialPersonaAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[SpatialPersonaAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	log.Printf("[SpatialPersonaAnalyzer] Marking task as running...")
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}
	log.Printf("[SpatialPersonaAnalyzer] Task marked as running")

	// Clear existing persona (full recompute)
	if mode == "full" {
		log.Printf("[SpatialPersonaAnalyzer] Clearing existing persona...")
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM spatial_persona"); err != nil {
			return fmt.Errorf("failed to clear spatial_persona: %w", err)
		}
		log.Printf("[SpatialPersonaAnalyzer] Cleared existing spatial persona")
	}

	// Collect features from all analysis tables
	log.Printf("[SpatialPersonaAnalyzer] Collecting features...")
	features, err := a.collectFeatures(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect features: %w", err)
	}
	log.Printf("[SpatialPersonaAnalyzer] Features collected successfully")

	// Calculate comprehensive spatial index (PSI)
	log.Printf("[SpatialPersonaAnalyzer] Calculating PSI...")
	psi := a.calculatePSI(features)
	log.Printf("[SpatialPersonaAnalyzer] PSI calculated: %.2f", psi)

	// Classify behavior type
	log.Printf("[SpatialPersonaAnalyzer] Classifying behavior type...")
	behaviorType := a.classifyBehaviorType(features)
	log.Printf("[SpatialPersonaAnalyzer] Behavior type: %s", behaviorType)

	// Calculate profile stability
	log.Printf("[SpatialPersonaAnalyzer] Calculating stability...")
	stability := a.calculateStability(features)
	log.Printf("[SpatialPersonaAnalyzer] Stability: %.2f", stability)

	// Create persona
	persona := &SpatialPersona{
		Features:     features,
		PSI:          psi,
		BehaviorType: behaviorType,
		Stability:    stability,
	}

	// Insert persona
	log.Printf("[SpatialPersonaAnalyzer] Inserting persona...")
	if err := a.insertPersona(ctx, persona); err != nil {
		return fmt.Errorf("failed to insert persona: %w", err)
	}
	log.Printf("[SpatialPersonaAnalyzer] Persona inserted successfully")

	// Mark task as completed
	summary := map[string]interface{}{
		"psi":           psi,
		"behavior_type": behaviorType,
		"stability":     stability,
	}
	summaryJSON, _ := json.Marshal(summary)

	log.Printf("[SpatialPersonaAnalyzer] Marking task as completed...")
	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[SpatialPersonaAnalyzer] Analysis completed: PSI=%.2f, Type=%s", psi, behaviorType)
	return nil
}

// SpatialPersona holds the comprehensive spatial profile
type SpatialPersona struct {
	Features     *FeatureVector
	PSI          float64
	BehaviorType string
	Stability    float64
}

// FeatureVector holds 9-dimensional feature vector
type FeatureVector struct {
	// Footprint features
	FootprintDiversity float64
	FootprintSpread    float64

	// Movement features
	MovementIntensity float64
	MovementBurst     float64

	// Spatial features
	SpatialComplexity float64
	SpatialEntropy    float64

	// Temporal features
	TemporalRegularity float64
	TemporalCoverage   float64

	// Road features
	RoadOverlap float64
}

// collectFeatures collects features from all analysis tables
func (a *SpatialPersonaAnalyzer) collectFeatures(ctx context.Context) (*FeatureVector, error) {
	features := &FeatureVector{}

	// 1. Footprint features
	var uniqueGrids, totalPoints int64
	err := a.DB.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT grid_id), COUNT(*)
		FROM "一生足迹"
		WHERE outlier_flag = 0
	`).Scan(&uniqueGrids, &totalPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to query footprint: %w", err)
	}
	features.FootprintDiversity = math.Log(1 + float64(uniqueGrids))
	features.FootprintSpread = math.Log(1 + float64(totalPoints))

	// 2. Movement features
	err = a.DB.QueryRowContext(ctx, `
		SELECT movement_intensity, burst_intensity
		FROM time_space_compression_bucketed
		WHERE bucket_type = 'all'
		LIMIT 1
	`).Scan(&features.MovementIntensity, &features.MovementBurst)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query movement: %w", err)
	}

	// 3. Spatial features
	err = a.DB.QueryRowContext(ctx, `
		SELECT trajectory_complexity, spatial_entropy
		FROM complexity_metrics
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&features.SpatialComplexity, &features.SpatialEntropy)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query complexity: %w", err)
	}

	// 4. Temporal features (from time-space slices)
	var sliceCount int64
	err = a.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM time_space_slices
		WHERE slice_type = 'WEEKLY_HOURLY' AND point_count > 0
	`).Scan(&sliceCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query temporal: %w", err)
	}
	features.TemporalRegularity = float64(sliceCount) / 168.0
	features.TemporalCoverage = math.Min(features.TemporalRegularity*2, 1.0)

	// 5. Road features
	var onRoad, offRoad float64
	err = a.DB.QueryRowContext(ctx, `
		SELECT SUM(on_road_distance_m), SUM(off_road_distance_m)
		FROM road_overlap_stats
	`).Scan(&onRoad, &offRoad)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query road overlap: %w", err)
	}
	if onRoad+offRoad > 0 {
		features.RoadOverlap = onRoad / (onRoad + offRoad)
	}

	return features, nil
}

// calculatePSI calculates Personal Spatial Index (0-100)
func (a *SpatialPersonaAnalyzer) calculatePSI(features *FeatureVector) float64 {
	weights := map[string]float64{
		"footprint_diversity":  0.15,
		"footprint_spread":     0.10,
		"movement_intensity":   0.15,
		"movement_burst":       0.10,
		"spatial_complexity":   0.15,
		"spatial_entropy":      0.10,
		"temporal_regularity":  0.10,
		"temporal_coverage":    0.05,
		"road_overlap":         0.10,
	}

	normalized := map[string]float64{
		"footprint_diversity":  math.Min(features.FootprintDiversity/10.0, 1.0),
		"footprint_spread":     math.Min(features.FootprintSpread/15.0, 1.0),
		"movement_intensity":   math.Min(features.MovementIntensity/50.0, 1.0),
		"movement_burst":       math.Min(features.MovementBurst/100.0, 1.0),
		"spatial_complexity":   features.SpatialComplexity,
		"spatial_entropy":      math.Min(features.SpatialEntropy/10.0, 1.0),
		"temporal_regularity":  features.TemporalRegularity,
		"temporal_coverage":    features.TemporalCoverage,
		"road_overlap":         features.RoadOverlap,
	}

	psi := 0.0
	for key, weight := range weights {
		psi += normalized[key] * weight
	}

	return psi * 100.0
}

// classifyBehaviorType classifies spatial behavior type
func (a *SpatialPersonaAnalyzer) classifyBehaviorType(features *FeatureVector) string {
	if features.FootprintDiversity > 8.0 && features.MovementIntensity > 30.0 {
		return "EXPLORER"
	} else if features.TemporalRegularity > 0.7 && features.RoadOverlap > 0.85 {
		return "COMMUTER"
	} else if features.FootprintDiversity < 5.0 && features.SpatialComplexity < 0.5 {
		return "HOMEBODY"
	} else if features.MovementBurst > 70.0 {
		return "TRAVELER"
	} else if features.SpatialEntropy > 7.0 {
		return "WANDERER"
	} else {
		return "BALANCED"
	}
}

// calculateStability calculates profile stability (0-1)
func (a *SpatialPersonaAnalyzer) calculateStability(features *FeatureVector) float64 {
	temporalStability := features.TemporalRegularity
	spatialStability := 1.0 - (features.SpatialComplexity * 0.5)
	return (temporalStability + spatialStability) / 2.0
}

// insertPersona inserts spatial persona into the database
func (a *SpatialPersonaAnalyzer) insertPersona(ctx context.Context, persona *SpatialPersona) error {
	featuresJSON, _ := json.Marshal(persona.Features)

	insertQuery := `
		INSERT INTO spatial_persona (
			psi, behavior_type, stability,
			footprint_diversity, footprint_spread,
			movement_intensity, movement_burst,
			spatial_complexity, spatial_entropy,
			temporal_regularity, temporal_coverage,
			road_overlap, features_json,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	_, err := a.DB.ExecContext(ctx, insertQuery,
		persona.PSI, persona.BehaviorType, persona.Stability,
		persona.Features.FootprintDiversity, persona.Features.FootprintSpread,
		persona.Features.MovementIntensity, persona.Features.MovementBurst,
		persona.Features.SpatialComplexity, persona.Features.SpatialEntropy,
		persona.Features.TemporalRegularity, persona.Features.TemporalCoverage,
		persona.Features.RoadOverlap, string(featuresJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert spatial persona: %w", err)
	}

	log.Printf("[SpatialPersonaAnalyzer] Inserted spatial persona")
	return nil
}

// Register the analyzer
func init() {
	log.Println("[integration] Package loaded")
	log.Println("[integration] Registering spatial_persona analyzer")
	analysis.RegisterAnalyzer("spatial_persona", NewSpatialPersonaAnalyzer)
	log.Println("[integration] spatial_persona analyzer registered")
}
