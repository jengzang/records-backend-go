package stats

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// AdminCrossingsAnalyzer implements administrative boundary crossing detection
// Skill: 行政区划穿越检测 (Admin Crossings Detection)
// Detects when trajectory crosses administrative boundaries
type AdminCrossingsAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewAdminCrossingsAnalyzer creates a new admin crossings analyzer
func NewAdminCrossingsAnalyzer(db *sql.DB) analysis.Analyzer {
	return &AdminCrossingsAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "admin_crossings", 10000),
	}
}

// Analyze performs admin crossing detection
func (a *AdminCrossingsAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[AdminCrossingsAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing crossings (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM admin_crossings"); err != nil {
			return fmt.Errorf("failed to clear admin_crossings: %w", err)
		}
		log.Printf("[AdminCrossingsAnalyzer] Cleared existing crossings")
	}

	// Query track points ordered by time
	query := `
		SELECT
			id, dataTime, latitude, longitude,
			province, city, county, town
		FROM "一生足迹"
		WHERE outlier_flag = 0
			AND province IS NOT NULL
		ORDER BY dataTime
	`

	rows, err := a.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query track points: %w", err)
	}
	defer rows.Close()

	// Process points and detect crossings
	var crossings []Crossing
	var prevPoint *TrackPoint
	totalPoints := 0

	for rows.Next() {
		var point TrackPoint
		var province, city, county, town sql.NullString

		if err := rows.Scan(
			&point.ID, &point.DataTime, &point.Latitude, &point.Longitude,
			&province, &city, &county, &town,
		); err != nil {
			return fmt.Errorf("failed to scan track point: %w", err)
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

		totalPoints++

		// Detect crossings by comparing with previous point
		if prevPoint != nil {
			crossing := a.detectCrossing(prevPoint, &point)
			if crossing != nil {
				crossings = append(crossings, *crossing)
			}
		}

		prevPoint = &point
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("[AdminCrossingsAnalyzer] Processed %d points, detected %d crossings", totalPoints, len(crossings))

	// Update task progress
	if err := a.UpdateTaskProgress(taskID, int64(totalPoints), int64(totalPoints), 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Insert crossings
	if err := a.insertCrossings(ctx, crossings); err != nil {
		return fmt.Errorf("failed to insert crossings: %w", err)
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_points": totalPoints,
		"crossings":    len(crossings),
		"province":     a.countCrossingsByType(crossings, "PROVINCE"),
		"city":         a.countCrossingsByType(crossings, "CITY"),
		"county":       a.countCrossingsByType(crossings, "COUNTY"),
		"town":         a.countCrossingsByType(crossings, "TOWN"),
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[AdminCrossingsAnalyzer] Analysis completed: %d crossings detected", len(crossings))
	return nil
}

// TrackPoint holds track point data
type TrackPoint struct {
	ID        int64
	DataTime  int64
	Latitude  float64
	Longitude float64
	Province  string
	City      string
	County    string
	Town      string
}

// Crossing holds crossing event data
type Crossing struct {
	CrossingTS   int64
	FromProvince string
	FromCity     string
	FromCounty   string
	FromTown     string
	ToProvince   string
	ToCity       string
	ToCounty     string
	ToTown       string
	CrossingType string
	Latitude     float64
	Longitude    float64
	Distance     float64
}

// detectCrossing detects if there's an admin boundary crossing between two points
func (a *AdminCrossingsAnalyzer) detectCrossing(prev, curr *TrackPoint) *Crossing {
	// Check for province crossing (highest priority)
	if prev.Province != curr.Province && curr.Province != "" {
		distance := haversineDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
		return &Crossing{
			CrossingTS:   curr.DataTime,
			FromProvince: prev.Province,
			FromCity:     prev.City,
			FromCounty:   prev.County,
			FromTown:     prev.Town,
			ToProvince:   curr.Province,
			ToCity:       curr.City,
			ToCounty:     curr.County,
			ToTown:       curr.Town,
			CrossingType: "PROVINCE",
			Latitude:     curr.Latitude,
			Longitude:    curr.Longitude,
			Distance:     distance,
		}
	}

	// Check for city crossing
	if prev.City != curr.City && curr.City != "" && prev.Province == curr.Province {
		distance := haversineDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
		return &Crossing{
			CrossingTS:   curr.DataTime,
			FromProvince: prev.Province,
			FromCity:     prev.City,
			FromCounty:   prev.County,
			FromTown:     prev.Town,
			ToProvince:   curr.Province,
			ToCity:       curr.City,
			ToCounty:     curr.County,
			ToTown:       curr.Town,
			CrossingType: "CITY",
			Latitude:     curr.Latitude,
			Longitude:    curr.Longitude,
			Distance:     distance,
		}
	}

	// Check for county crossing
	if prev.County != curr.County && curr.County != "" && prev.City == curr.City {
		distance := haversineDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
		return &Crossing{
			CrossingTS:   curr.DataTime,
			FromProvince: prev.Province,
			FromCity:     prev.City,
			FromCounty:   prev.County,
			FromTown:     prev.Town,
			ToProvince:   curr.Province,
			ToCity:       curr.City,
			ToCounty:     curr.County,
			ToTown:       curr.Town,
			CrossingType: "COUNTY",
			Latitude:     curr.Latitude,
			Longitude:    curr.Longitude,
			Distance:     distance,
		}
	}

	// Check for town crossing
	if prev.Town != curr.Town && curr.Town != "" && prev.County == curr.County {
		distance := haversineDistance(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
		return &Crossing{
			CrossingTS:   curr.DataTime,
			FromProvince: prev.Province,
			FromCity:     prev.City,
			FromCounty:   prev.County,
			FromTown:     prev.Town,
			ToProvince:   curr.Province,
			ToCity:       curr.City,
			ToCounty:     curr.County,
			ToTown:       curr.Town,
			CrossingType: "TOWN",
			Latitude:     curr.Latitude,
			Longitude:    curr.Longitude,
			Distance:     distance,
		}
	}

	return nil
}

// haversineDistance calculates the distance between two points in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// insertCrossings inserts crossings into the database
func (a *AdminCrossingsAnalyzer) insertCrossings(ctx context.Context, crossings []Crossing) error {
	if len(crossings) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO admin_crossings (
			crossing_ts, from_province, from_city, from_county, from_town,
			to_province, to_city, to_county, to_town,
			crossing_type, latitude, longitude, distance_from_prev_m,
			algo_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'v1', CURRENT_TIMESTAMP)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, crossing := range crossings {
		_, err := stmt.ExecContext(ctx,
			crossing.CrossingTS,
			crossing.FromProvince, crossing.FromCity, crossing.FromCounty, crossing.FromTown,
			crossing.ToProvince, crossing.ToCity, crossing.ToCounty, crossing.ToTown,
			crossing.CrossingType,
			crossing.Latitude, crossing.Longitude, crossing.Distance,
		)
		if err != nil {
			return fmt.Errorf("failed to insert crossing: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[AdminCrossingsAnalyzer] Inserted %d crossings", len(crossings))
	return nil
}

// countCrossingsByType counts crossings by type
func (a *AdminCrossingsAnalyzer) countCrossingsByType(crossings []Crossing, crossingType string) int {
	count := 0
	for _, crossing := range crossings {
		if crossing.CrossingType == crossingType {
			count++
		}
	}
	return count
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("admin_crossings", NewAdminCrossingsAnalyzer)
}
