package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// DirectionalBiasAnalyzer implements directional movement pattern analysis
// Skill: 方向偏好分析 (Directional Bias)
// Analyzes heading distribution and identifies preferred directions
type DirectionalBiasAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewDirectionalBiasAnalyzer creates a new directional bias analyzer
func NewDirectionalBiasAnalyzer(db *sql.DB) analysis.Analyzer {
	return &DirectionalBiasAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "directional_bias", 10000),
	}
}

// Analyze performs directional bias analysis
func (a *DirectionalBiasAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[DirectionalBiasAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing stats (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM directional_stats_bucketed"); err != nil {
			return fmt.Errorf("failed to clear directional_stats_bucketed: %w", err)
		}
		log.Printf("[DirectionalBiasAnalyzer] Cleared existing directional stats")
	}

	// Query segments with coordinates and transport mode
	query := `
		SELECT
			s.id,
			s.start_time,
			s.end_time,
			s.distance_m,
			s.duration_s,
			s.mode,
			p1.latitude AS start_lat,
			p1.longitude AS start_lon,
			p2.latitude AS end_lat,
			p2.longitude AS end_lon,
			p1.province,
			p1.city,
			p1.county
		FROM segments s
		JOIN "一生足迹" p1 ON s.start_point_id = p1.id
		JOIN "一生足迹" p2 ON s.end_point_id = p2.id
		WHERE s.distance_m > 10
			AND p1.latitude IS NOT NULL
			AND p1.longitude IS NOT NULL
			AND p2.latitude IS NOT NULL
			AND p2.longitude IS NOT NULL
		ORDER BY s.start_time
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	// Aggregation map: (bucket_type, bucket_key, area_type, area_key, mode_filter) -> stats
	type AggKey struct {
		BucketType string
		BucketKey  string
		AreaType   string
		AreaKey    string
		ModeFilter string
	}
	aggMap := make(map[AggKey]*DirectionalAggregation)

	totalSegments := 0
	for rows.Next() {
		var seg Segment
		if err := rows.Scan(
			&seg.ID, &seg.StartTime, &seg.EndTime, &seg.Distance, &seg.Duration, &seg.Mode,
			&seg.StartLat, &seg.StartLon, &seg.EndLat, &seg.EndLon,
			&seg.Province, &seg.City, &seg.County,
		); err != nil {
			return fmt.Errorf("failed to scan segment: %w", err)
		}

		totalSegments++

		// Calculate bearing
		bearing := calculateBearing(seg.StartLat, seg.StartLon, seg.EndLat, seg.EndLon)
		bucket := bearingToBucket(bearing, 8)

		// Extract time dimensions
		startTime := time.Unix(seg.StartTime, 0)
		year := startTime.Format("2006")
		month := startTime.Format("2006-01")

		// Define aggregation keys
		areas := []struct {
			areaType string
			areaKey  string
		}{
			{"PROVINCE", seg.Province.String},
			{"CITY", seg.City.String},
			{"COUNTY", seg.County.String},
		}

		bucketTypes := []struct {
			bucketType string
			bucketKey  string
		}{
			{"all", "all"},
			{"year", year},
			{"month", month},
		}

		modeFilters := []string{"ALL", seg.Mode.String}

		// Aggregate across all dimensions
		for _, area := range areas {
			if area.areaKey == "" {
				continue
			}
			for _, bt := range bucketTypes {
				for _, mode := range modeFilters {
					if mode == "" {
						continue
					}

					key := AggKey{
						BucketType: bt.bucketType,
						BucketKey:  bt.bucketKey,
						AreaType:   area.areaType,
						AreaKey:    area.areaKey,
						ModeFilter: mode,
					}

					if aggMap[key] == nil {
						aggMap[key] = &DirectionalAggregation{
							Buckets: make([]float64, 8),
							Counts:  make([]int, 8),
						}
					}

					agg := aggMap[key]
					agg.Buckets[bucket] += seg.Distance
					agg.Counts[bucket]++
					agg.TotalDistance += seg.Distance
					agg.TotalDuration += seg.Duration
					agg.SegmentCount++
				}
			}
		}

		// Progress update every 1000 segments
		if totalSegments%1000 == 0 {
			if err := a.UpdateTaskProgress(taskID, int64(totalSegments), 0, 0); err != nil {
				log.Printf("[DirectionalBiasAnalyzer] Warning: failed to update progress: %v", err)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("[DirectionalBiasAnalyzer] Processed %d segments, generated %d aggregations", totalSegments, len(aggMap))

	// Calculate metrics and insert results
	insertedCount := 0
	for key, agg := range aggMap {
		// Calculate advanced metrics
		metrics := calculateDirectionalMetrics(agg.Buckets, agg.Counts)

		// Prepare histogram JSON
		histogram := make([]map[string]interface{}, 8)
		for i := 0; i < 8; i++ {
			histogram[i] = map[string]interface{}{
				"bin":      i,
				"distance": agg.Buckets[i],
				"count":    agg.Counts[i],
			}
		}
		histogramJSON, _ := json.Marshal(histogram)

		// Insert into database
		insertQuery := `
			INSERT INTO directional_stats_bucketed (
				bucket_type, bucket_key, area_type, area_key, mode_filter,
				direction_histogram_json, num_bins,
				dominant_direction_deg, directional_concentration,
				bidirectional_score, directional_entropy,
				total_distance, total_duration, segment_count,
				algo_version
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
			ON CONFLICT(bucket_type, bucket_key, area_type, area_key, mode_filter)
			DO UPDATE SET
				direction_histogram_json = excluded.direction_histogram_json,
				num_bins = excluded.num_bins,
				dominant_direction_deg = excluded.dominant_direction_deg,
				directional_concentration = excluded.directional_concentration,
				bidirectional_score = excluded.bidirectional_score,
				directional_entropy = excluded.directional_entropy,
				total_distance = excluded.total_distance,
				total_duration = excluded.total_duration,
				segment_count = excluded.segment_count,
				created_at = CURRENT_TIMESTAMP
		`

		_, err := a.DB.ExecContext(ctx, insertQuery,
			key.BucketType, key.BucketKey, key.AreaType, key.AreaKey, key.ModeFilter,
			string(histogramJSON), 8,
			metrics.DominantDirection, metrics.Concentration,
			metrics.BidirectionalScore, metrics.Entropy,
			agg.TotalDistance, agg.TotalDuration, agg.SegmentCount,
		)
		if err != nil {
			return fmt.Errorf("failed to insert directional stats: %w", err)
		}

		insertedCount++
		if insertedCount%100 == 0 {
			log.Printf("[DirectionalBiasAnalyzer] Inserted %d/%d records", insertedCount, len(aggMap))
		}
	}

	log.Printf("[DirectionalBiasAnalyzer] Inserted %d directional stats records", insertedCount)

	// Mark task as completed
	summary := map[string]interface{}{
		"total_segments":    totalSegments,
		"total_aggregations": len(aggMap),
		"inserted_records":  insertedCount,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[DirectionalBiasAnalyzer] Analysis completed")
	return nil
}

// Segment represents a trajectory segment with coordinates
type Segment struct {
	ID        int64
	StartTime int64
	EndTime   int64
	Distance  float64
	Duration  int64
	Mode      sql.NullString
	StartLat  float64
	StartLon  float64
	EndLat    float64
	EndLon    float64
	Province  sql.NullString
	City      sql.NullString
	County    sql.NullString
}

// DirectionalAggregation holds aggregated directional statistics
type DirectionalAggregation struct {
	Buckets       []float64 // Distance per bucket
	Counts        []int     // Segment count per bucket
	TotalDistance float64
	TotalDuration int64
	SegmentCount  int
}

// DirectionalMetrics holds calculated directional metrics
type DirectionalMetrics struct {
	DominantDirection   float64
	Concentration       float64
	BidirectionalScore  float64
	Entropy             float64
}

// calculateBearing calculates the initial bearing from point 1 to point 2
// using the great circle formula
func calculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	y := math.Sin(deltaLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLon)

	bearing := math.Atan2(y, x) * 180 / math.Pi
	return math.Mod(bearing+360, 360) // Normalize to [0, 360)
}

// bearingToBucket converts bearing (0-360) to bucket index
func bearingToBucket(bearing float64, numBins int) int {
	binSize := 360.0 / float64(numBins)
	bucket := int((bearing + binSize/2) / binSize)
	if bucket >= numBins {
		bucket = 0
	}
	return bucket
}

// calculateDirectionalMetrics calculates advanced directional metrics
func calculateDirectionalMetrics(buckets []float64, counts []int) DirectionalMetrics {
	numBins := len(buckets)
	totalDistance := 0.0
	for _, d := range buckets {
		totalDistance += d
	}

	if totalDistance == 0 {
		return DirectionalMetrics{}
	}

	// 1. Dominant direction (weighted average of dominant bin)
	dominantBin := 0
	maxDistance := 0.0
	for i, d := range buckets {
		if d > maxDistance {
			maxDistance = d
			dominantBin = i
		}
	}
	binSize := 360.0 / float64(numBins)
	dominantDirection := float64(dominantBin)*binSize + binSize/2

	// 2. Directional concentration (vector synthesis)
	var sumX, sumY float64
	for i, d := range buckets {
		angle := (float64(i)*binSize + binSize/2) * math.Pi / 180
		weight := d / totalDistance
		sumX += weight * math.Cos(angle)
		sumY += weight * math.Sin(angle)
	}
	concentration := math.Sqrt(sumX*sumX + sumY*sumY)

	// 3. Bidirectional score (opposite bin pairing)
	bidirectionalScore := 0.0
	for i := 0; i < numBins/2; i++ {
		opposite := i + numBins/2
		pairDistance := buckets[i] + buckets[opposite]
		if pairDistance > 0 {
			balance := math.Min(buckets[i], buckets[opposite]) / pairDistance
			weight := pairDistance / totalDistance
			bidirectionalScore += balance * weight
		}
	}

	// 4. Directional entropy (Shannon entropy, normalized)
	entropy := 0.0
	for _, d := range buckets {
		if d > 0 {
			p := d / totalDistance
			entropy -= p * math.Log2(p)
		}
	}
	maxEntropy := math.Log2(float64(numBins))
	if maxEntropy > 0 {
		entropy /= maxEntropy // Normalize to [0, 1]
	}

	return DirectionalMetrics{
		DominantDirection:  dominantDirection,
		Concentration:      concentration,
		BidirectionalScore: bidirectionalScore,
		Entropy:            entropy,
	}
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("directional_bias", NewDirectionalBiasAnalyzer)
}
