package annotation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jengzang/records-backend-go/internal/analysis"
)

// StayAnnotationAnalyzer implements stay annotation and label suggestion
// Skill: 停留标注与建议 (Stay Annotation)
// Generates context cards and label suggestions for stays
type StayAnnotationAnalyzer struct {
	*analysis.IncrementalAnalyzer
}

// NewStayAnnotationAnalyzer creates a new stay annotation analyzer
func NewStayAnnotationAnalyzer(db *sql.DB) analysis.Analyzer {
	return &StayAnnotationAnalyzer{
		IncrementalAnalyzer: analysis.NewIncrementalAnalyzer(db, "stay_annotation", 1000),
	}
}

// Analyze performs stay annotation and label suggestion
func (a *StayAnnotationAnalyzer) Analyze(ctx context.Context, taskID int64, mode string) error {
	log.Printf("[StayAnnotationAnalyzer] Starting analysis (task_id=%d, mode=%s)", taskID, mode)

	// Mark task as running
	if err := a.MarkTaskAsRunning(taskID); err != nil {
		return fmt.Errorf("failed to mark task as running: %w", err)
	}

	// Clear existing context cache (full recompute)
	if mode == "full" {
		if _, err := a.DB.ExecContext(ctx, "DELETE FROM stay_context_cache"); err != nil {
			return fmt.Errorf("failed to clear context cache: %w", err)
		}
		log.Printf("[StayAnnotationAnalyzer] Cleared existing context cache")
	}

	// Get all stay segments
	staysQuery := `
		SELECT
			id,
			start_ts,
			end_ts,
			duration_s,
			center_lat,
			center_lon,
			province,
			city,
			county,
			town,
			grid_id
		FROM stay_segments
		WHERE outlier_flag = 0
		ORDER BY id
	`

	rows, err := a.DB.QueryContext(ctx, staysQuery)
	if err != nil {
		return fmt.Errorf("failed to query stays: %w", err)
	}

	var stays []StayInfo
	for rows.Next() {
		var stay StayInfo
		if err := rows.Scan(&stay.ID, &stay.StartTS, &stay.EndTS, &stay.DurationS,
			&stay.CenterLat, &stay.CenterLon, &stay.Province, &stay.City,
			&stay.County, &stay.Town, &stay.GridID); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan stay: %w", err)
		}
		stays = append(stays, stay)
	}
	rows.Close()

	log.Printf("[StayAnnotationAnalyzer] Processing %d stays", len(stays))

	// Update task with total count
	if err := a.UpdateTaskProgress(taskID, int64(len(stays)), 0, 0); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Load place anchors (HOME, WORK, etc.)
	anchors, err := a.loadPlaceAnchors(ctx)
	if err != nil {
		return fmt.Errorf("failed to load place anchors: %w", err)
	}

	log.Printf("[StayAnnotationAnalyzer] Loaded %d place anchors", len(anchors))

	// Process each stay
	processed := 0
	batchSize := 100
	var contextCache []StayContext

	for _, stay := range stays {
		// Extract time features
		timeFeatures := a.extractTimeFeatures(stay)

		// Extract arrival/departure context
		arrivalContext, departureContext := a.extractMovementContext(ctx, stay)

		// Extract location features
		locationFeatures := a.extractLocationFeatures(stay)

		// Query historical annotations for this location
		historicalLabel := a.queryHistoricalLabel(ctx, stay)

		// Generate label suggestions using rule engine
		suggestions := a.generateLabelSuggestions(stay, timeFeatures, arrivalContext, departureContext, locationFeatures, historicalLabel, anchors)

		// Create context card
		contextCard := ContextCard{
			StayID:            stay.ID,
			TimeFeatures:      timeFeatures,
			LocationFeatures:  locationFeatures,
			ArrivalContext:    arrivalContext,
			DepartureContext:  departureContext,
			HistoricalLabel:   historicalLabel,
		}

		contextJSON, _ := json.Marshal(contextCard)
		suggestionsJSON, _ := json.Marshal(suggestions)

		contextCache = append(contextCache, StayContext{
			StayID:      stay.ID,
			ContextJSON: string(contextJSON),
			SuggestionsJSON: string(suggestionsJSON),
		})

		processed++
		if processed%batchSize == 0 {
			// Insert batch
			if err := a.insertContextCache(ctx, contextCache); err != nil {
				return fmt.Errorf("failed to insert context cache: %w", err)
			}
			contextCache = nil

			if err := a.UpdateTaskProgress(taskID, int64(len(stays)), int64(processed), 0); err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}
			log.Printf("[StayAnnotationAnalyzer] Processed %d/%d stays", processed, len(stays))
		}
	}

	// Insert remaining context cache
	if len(contextCache) > 0 {
		if err := a.insertContextCache(ctx, contextCache); err != nil {
			return fmt.Errorf("failed to insert context cache: %w", err)
		}
	}

	// Mark task as completed
	summary := map[string]interface{}{
		"total_stays":     len(stays),
		"processed_stays": processed,
	}
	summaryJSON, _ := json.Marshal(summary)

	if err := a.MarkTaskAsCompleted(taskID, string(summaryJSON)); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}

	log.Printf("[StayAnnotationAnalyzer] Analysis completed: %d stays processed", processed)
	return nil
}

// StayInfo holds stay information
type StayInfo struct {
	ID        int64
	StartTS   int64
	EndTS     int64
	DurationS int64
	CenterLat float64
	CenterLon float64
	Province  sql.NullString
	City      sql.NullString
	County    sql.NullString
	Town      sql.NullString
	GridID    sql.NullString
}

// TimeFeatures holds time-related features
type TimeFeatures struct {
	HourOfDay    int
	Weekday      int
	IsWeekend    bool
	IsNight      bool
	IsOvernight  bool
	DurationHours float64
}

// LocationFeatures holds location-related features
type LocationFeatures struct {
	Province string
	City     string
	County   string
	Town     string
	GridID   string
}

// MovementContext holds arrival/departure context
type MovementContext struct {
	Mode     string
	Distance float64
	Duration int64
}

// ContextCard holds complete context for a stay
type ContextCard struct {
	StayID            int64
	TimeFeatures      TimeFeatures
	LocationFeatures  LocationFeatures
	ArrivalContext    MovementContext
	DepartureContext  MovementContext
	HistoricalLabel   string
}

// LabelSuggestion holds a label suggestion with confidence
type LabelSuggestion struct {
	Label      string
	Confidence float64
	Reasons    []string
}

// PlaceAnchor holds a known place anchor
type PlaceAnchor struct {
	Type   string
	GridID string
}

// StayContext holds stay context cache entry
type StayContext struct {
	StayID          int64
	ContextJSON     string
	SuggestionsJSON string
}

// extractTimeFeatures extracts time-related features
func (a *StayAnnotationAnalyzer) extractTimeFeatures(stay StayInfo) TimeFeatures {
	startTime := time.Unix(stay.StartTS, 0)
	endTime := time.Unix(stay.EndTS, 0)

	hourOfDay := startTime.Hour()
	weekday := int(startTime.Weekday())
	isWeekend := weekday == 0 || weekday == 6
	isNight := hourOfDay >= 22 || hourOfDay < 6
	isOvernight := startTime.Day() != endTime.Day()
	durationHours := float64(stay.DurationS) / 3600.0

	return TimeFeatures{
		HourOfDay:     hourOfDay,
		Weekday:       weekday,
		IsWeekend:     isWeekend,
		IsNight:       isNight,
		IsOvernight:   isOvernight,
		DurationHours: durationHours,
	}
}

// extractLocationFeatures extracts location-related features
func (a *StayAnnotationAnalyzer) extractLocationFeatures(stay StayInfo) LocationFeatures {
	features := LocationFeatures{}

	if stay.Province.Valid {
		features.Province = stay.Province.String
	}
	if stay.City.Valid {
		features.City = stay.City.String
	}
	if stay.County.Valid {
		features.County = stay.County.String
	}
	if stay.Town.Valid {
		features.Town = stay.Town.String
	}
	if stay.GridID.Valid {
		features.GridID = stay.GridID.String
	}

	return features
}

// extractMovementContext extracts arrival/departure context
func (a *StayAnnotationAnalyzer) extractMovementContext(ctx context.Context, stay StayInfo) (MovementContext, MovementContext) {
	// Query segment before stay (arrival)
	arrivalQuery := `
		SELECT mode, distance_m, duration_s
		FROM segments
		WHERE end_ts <= ?
		ORDER BY end_ts DESC
		LIMIT 1
	`

	var arrivalContext MovementContext
	var mode sql.NullString
	var distance, duration sql.NullFloat64

	err := a.DB.QueryRowContext(ctx, arrivalQuery, stay.StartTS).Scan(&mode, &distance, &duration)
	if err == nil {
		if mode.Valid {
			arrivalContext.Mode = mode.String
		}
		if distance.Valid {
			arrivalContext.Distance = distance.Float64
		}
		if duration.Valid {
			arrivalContext.Duration = int64(duration.Float64)
		}
	}

	// Query segment after stay (departure)
	departureQuery := `
		SELECT mode, distance_m, duration_s
		FROM segments
		WHERE start_ts >= ?
		ORDER BY start_ts ASC
		LIMIT 1
	`

	var departureContext MovementContext
	err = a.DB.QueryRowContext(ctx, departureQuery, stay.EndTS).Scan(&mode, &distance, &duration)
	if err == nil {
		if mode.Valid {
			departureContext.Mode = mode.String
		}
		if distance.Valid {
			departureContext.Distance = distance.Float64
		}
		if duration.Valid {
			departureContext.Duration = int64(duration.Float64)
		}
	}

	return arrivalContext, departureContext
}

// queryHistoricalLabel queries historical label for this location
func (a *StayAnnotationAnalyzer) queryHistoricalLabel(ctx context.Context, stay StayInfo) string {
	if !stay.GridID.Valid {
		return ""
	}

	query := `
		SELECT sa.label
		FROM stay_annotations sa
		JOIN stay_segments ss ON sa.stay_id = ss.id
		WHERE ss.grid_id = ?
			AND sa.confirmed = 1
		ORDER BY sa.updated_at DESC
		LIMIT 1
	`

	var label string
	err := a.DB.QueryRowContext(ctx, query, stay.GridID.String).Scan(&label)
	if err != nil {
		return ""
	}

	return label
}

// loadPlaceAnchors loads known place anchors
func (a *StayAnnotationAnalyzer) loadPlaceAnchors(ctx context.Context) ([]PlaceAnchor, error) {
	query := `
		SELECT type, grid_id
		FROM place_anchors
		WHERE active_to_ts IS NULL OR active_to_ts > ?
	`

	rows, err := a.DB.QueryContext(ctx, query, time.Now().Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query place anchors: %w", err)
	}
	defer rows.Close()

	var anchors []PlaceAnchor
	for rows.Next() {
		var anchor PlaceAnchor
		if err := rows.Scan(&anchor.Type, &anchor.GridID); err != nil {
			return nil, fmt.Errorf("failed to scan anchor: %w", err)
		}
		anchors = append(anchors, anchor)
	}

	return anchors, nil
}

// generateLabelSuggestions generates label suggestions using rule engine
func (a *StayAnnotationAnalyzer) generateLabelSuggestions(
	stay StayInfo,
	timeFeatures TimeFeatures,
	arrivalContext MovementContext,
	departureContext MovementContext,
	locationFeatures LocationFeatures,
	historicalLabel string,
	anchors []PlaceAnchor,
) []LabelSuggestion {
	var suggestions []LabelSuggestion

	// Check if this location matches a known anchor
	if stay.GridID.Valid {
		for _, anchor := range anchors {
			if anchor.GridID == stay.GridID.String {
				suggestions = append(suggestions, LabelSuggestion{
					Label:      anchor.Type,
					Confidence: 0.95,
					Reasons:    []string{"KNOWN_ANCHOR"},
				})
				return suggestions
			}
		}
	}

	// Use historical label if available
	if historicalLabel != "" {
		suggestions = append(suggestions, LabelSuggestion{
			Label:      historicalLabel,
			Confidence: 0.85,
			Reasons:    []string{"HISTORICAL_LABEL"},
		})
	}

	// Rule-based suggestions

	// HOME: overnight stay, night hours, long duration
	if timeFeatures.IsOvernight && timeFeatures.IsNight && timeFeatures.DurationHours >= 6 {
		suggestions = append(suggestions, LabelSuggestion{
			Label:      "HOME",
			Confidence: 0.8,
			Reasons:    []string{"OVERNIGHT", "NIGHT_HOURS", "LONG_DURATION"},
		})
	}

	// WORK: weekday daytime, long duration
	if !timeFeatures.IsWeekend && timeFeatures.HourOfDay >= 8 && timeFeatures.HourOfDay <= 18 && timeFeatures.DurationHours >= 4 {
		suggestions = append(suggestions, LabelSuggestion{
			Label:      "WORK",
			Confidence: 0.7,
			Reasons:    []string{"WEEKDAY", "DAYTIME", "LONG_DURATION"},
		})
	}

	// EAT: meal hours, short duration
	if (timeFeatures.HourOfDay >= 11 && timeFeatures.HourOfDay <= 14) || (timeFeatures.HourOfDay >= 17 && timeFeatures.HourOfDay <= 20) {
		if timeFeatures.DurationHours >= 0.5 && timeFeatures.DurationHours <= 2 {
			suggestions = append(suggestions, LabelSuggestion{
				Label:      "EAT",
				Confidence: 0.6,
				Reasons:    []string{"MEAL_HOURS", "SHORT_DURATION"},
			})
		}
	}

	// SLEEP: night hours, medium duration
	if timeFeatures.IsNight && timeFeatures.DurationHours >= 4 && timeFeatures.DurationHours <= 10 {
		suggestions = append(suggestions, LabelSuggestion{
			Label:      "SLEEP",
			Confidence: 0.65,
			Reasons:    []string{"NIGHT_HOURS", "MEDIUM_DURATION"},
		})
	}

	// TRANSIT: short duration, between movements
	if timeFeatures.DurationHours < 1 && arrivalContext.Mode != "" && departureContext.Mode != "" {
		suggestions = append(suggestions, LabelSuggestion{
			Label:      "TRANSIT",
			Confidence: 0.5,
			Reasons:    []string{"SHORT_DURATION", "BETWEEN_MOVEMENTS"},
		})
	}

	return suggestions
}

// insertContextCache inserts context cache into the database
func (a *StayAnnotationAnalyzer) insertContextCache(ctx context.Context, cache []StayContext) error {
	if len(cache) == 0 {
		return nil
	}

	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT OR REPLACE INTO stay_context_cache (
			stay_id, context_json, suggestions_json, computed_at, algo_version
		) VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'v1')
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range cache {
		_, err := stmt.ExecContext(ctx,
			c.StayID,
			c.ContextJSON,
			c.SuggestionsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert context cache: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Register the analyzer
func init() {
	analysis.RegisterAnalyzer("stay_annotation", NewStayAnnotationAnalyzer)
}
