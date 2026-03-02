package efficiency

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository handles database operations for efficiency data
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new efficiency repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetHourlyScores retrieves hourly efficiency scores for a date range
func (r *Repository) GetHourlyScores(startDate, endDate string) ([]HourlyEfficiencyScore, error) {
	query := `
		SELECT
			id, date, hour,
			typing_speed, typing_speed_normalized,
			work_app_ratio, entertainment_app_ratio, focus_session_count, app_switch_frequency,
			work_app_ratio_normalized, focus_normalized,
			avg_heart_rate, heart_rate_variability, step_count,
			hrv_normalized, activity_normalized,
			efficiency_score,
			has_keyboard_data, has_screentime_data, has_health_data, data_completeness,
			created_at, updated_at
		FROM hourly_efficiency_scores
		WHERE date >= ? AND date <= ?
		ORDER BY date, hour
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly scores: %w", err)
	}
	defer rows.Close()

	var scores []HourlyEfficiencyScore
	for rows.Next() {
		var score HourlyEfficiencyScore
		err := rows.Scan(
			&score.ID, &score.Date, &score.Hour,
			&score.TypingSpeed, &score.TypingSpeedNormalized,
			&score.WorkAppRatio, &score.EntertainmentAppRatio, &score.FocusSessionCount, &score.AppSwitchFrequency,
			&score.WorkAppRatioNormalized, &score.FocusNormalized,
			&score.AvgHeartRate, &score.HeartRateVariability, &score.StepCount,
			&score.HRVNormalized, &score.ActivityNormalized,
			&score.EfficiencyScore,
			&score.HasKeyboardData, &score.HasScreenTimeData, &score.HasHealthData, &score.DataCompleteness,
			&score.CreatedAt, &score.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hourly score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetProfile retrieves efficiency curve profile by type
func (r *Repository) GetProfile(profileType string) (*EfficiencyCurveProfile, error) {
	query := `
		SELECT
			id, profile_type,
			hour_0, hour_1, hour_2, hour_3, hour_4, hour_5, hour_6, hour_7,
			hour_8, hour_9, hour_10, hour_11, hour_12, hour_13, hour_14, hour_15,
			hour_16, hour_17, hour_18, hour_19, hour_20, hour_21, hour_22, hour_23,
			peak_hour, peak_score, peak_start_hour, peak_end_hour,
			low_hour, low_score,
			chronotype, chronotype_confidence,
			avg_efficiency, std_efficiency, total_samples,
			start_date, end_date,
			created_at, updated_at
		FROM efficiency_curve_profiles
		WHERE profile_type = ?
	`

	var profile EfficiencyCurveProfile
	err := r.db.QueryRow(query, profileType).Scan(
		&profile.ID, &profile.ProfileType,
		&profile.HourlyCurve[0], &profile.HourlyCurve[1], &profile.HourlyCurve[2], &profile.HourlyCurve[3],
		&profile.HourlyCurve[4], &profile.HourlyCurve[5], &profile.HourlyCurve[6], &profile.HourlyCurve[7],
		&profile.HourlyCurve[8], &profile.HourlyCurve[9], &profile.HourlyCurve[10], &profile.HourlyCurve[11],
		&profile.HourlyCurve[12], &profile.HourlyCurve[13], &profile.HourlyCurve[14], &profile.HourlyCurve[15],
		&profile.HourlyCurve[16], &profile.HourlyCurve[17], &profile.HourlyCurve[18], &profile.HourlyCurve[19],
		&profile.HourlyCurve[20], &profile.HourlyCurve[21], &profile.HourlyCurve[22], &profile.HourlyCurve[23],
		&profile.PeakHour, &profile.PeakScore, &profile.PeakStartHour, &profile.PeakEndHour,
		&profile.LowHour, &profile.LowScore,
		&profile.Chronotype, &profile.ChronotypeConfidence,
		&profile.AvgEfficiency, &profile.StdEfficiency, &profile.TotalSamples,
		&profile.StartDate, &profile.EndDate,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query profile: %w", err)
	}

	return &profile, nil
}

// GetInsights retrieves active efficiency insights
func (r *Repository) GetInsights() ([]EfficiencyInsight, error) {
	query := `
		SELECT
			id, insight_type, priority, title, description, recommendation,
			data, confidence, is_active, created_at, updated_at
		FROM efficiency_insights
		WHERE is_active = 1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query insights: %w", err)
	}
	defer rows.Close()

	var insights []EfficiencyInsight
	for rows.Next() {
		var insight EfficiencyInsight
		var recommendation, data sql.NullString
		err := rows.Scan(
			&insight.ID, &insight.InsightType, &insight.Priority,
			&insight.Title, &insight.Description, &recommendation,
			&data, &insight.Confidence, &insight.IsActive,
			&insight.CreatedAt, &insight.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan insight: %w", err)
		}
		if recommendation.Valid {
			insight.Recommendation = recommendation.String
		}
		if data.Valid {
			insight.Data = data.String
		}
		insights = append(insights, insight)
	}

	return insights, nil
}

// SaveHourlyScore inserts or updates an hourly efficiency score
func (r *Repository) SaveHourlyScore(score *HourlyEfficiencyScore) error {
	query := `
		INSERT INTO hourly_efficiency_scores (
			date, hour,
			typing_speed, typing_speed_normalized,
			work_app_ratio, entertainment_app_ratio, focus_session_count, app_switch_frequency,
			work_app_ratio_normalized, focus_normalized,
			avg_heart_rate, heart_rate_variability, step_count,
			hrv_normalized, activity_normalized,
			efficiency_score,
			has_keyboard_data, has_screentime_data, has_health_data, data_completeness,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date, hour) DO UPDATE SET
			typing_speed = excluded.typing_speed,
			typing_speed_normalized = excluded.typing_speed_normalized,
			work_app_ratio = excluded.work_app_ratio,
			entertainment_app_ratio = excluded.entertainment_app_ratio,
			focus_session_count = excluded.focus_session_count,
			app_switch_frequency = excluded.app_switch_frequency,
			work_app_ratio_normalized = excluded.work_app_ratio_normalized,
			focus_normalized = excluded.focus_normalized,
			avg_heart_rate = excluded.avg_heart_rate,
			heart_rate_variability = excluded.heart_rate_variability,
			step_count = excluded.step_count,
			hrv_normalized = excluded.hrv_normalized,
			activity_normalized = excluded.activity_normalized,
			efficiency_score = excluded.efficiency_score,
			has_keyboard_data = excluded.has_keyboard_data,
			has_screentime_data = excluded.has_screentime_data,
			has_health_data = excluded.has_health_data,
			data_completeness = excluded.data_completeness,
			updated_at = CURRENT_TIMESTAMP
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		score.Date, score.Hour,
		score.TypingSpeed, score.TypingSpeedNormalized,
		score.WorkAppRatio, score.EntertainmentAppRatio, score.FocusSessionCount, score.AppSwitchFrequency,
		score.WorkAppRatioNormalized, score.FocusNormalized,
		score.AvgHeartRate, score.HeartRateVariability, score.StepCount,
		score.HRVNormalized, score.ActivityNormalized,
		score.EfficiencyScore,
		score.HasKeyboardData, score.HasScreenTimeData, score.HasHealthData, score.DataCompleteness,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to save hourly score: %w", err)
	}

	return nil
}

// SaveProfile inserts or updates an efficiency curve profile
func (r *Repository) SaveProfile(profile *EfficiencyCurveProfile) error {
	query := `
		INSERT INTO efficiency_curve_profiles (
			profile_type,
			hour_0, hour_1, hour_2, hour_3, hour_4, hour_5, hour_6, hour_7,
			hour_8, hour_9, hour_10, hour_11, hour_12, hour_13, hour_14, hour_15,
			hour_16, hour_17, hour_18, hour_19, hour_20, hour_21, hour_22, hour_23,
			peak_hour, peak_score, peak_start_hour, peak_end_hour,
			low_hour, low_score,
			chronotype, chronotype_confidence,
			avg_efficiency, std_efficiency, total_samples,
			start_date, end_date,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_type) DO UPDATE SET
			hour_0 = excluded.hour_0, hour_1 = excluded.hour_1, hour_2 = excluded.hour_2, hour_3 = excluded.hour_3,
			hour_4 = excluded.hour_4, hour_5 = excluded.hour_5, hour_6 = excluded.hour_6, hour_7 = excluded.hour_7,
			hour_8 = excluded.hour_8, hour_9 = excluded.hour_9, hour_10 = excluded.hour_10, hour_11 = excluded.hour_11,
			hour_12 = excluded.hour_12, hour_13 = excluded.hour_13, hour_14 = excluded.hour_14, hour_15 = excluded.hour_15,
			hour_16 = excluded.hour_16, hour_17 = excluded.hour_17, hour_18 = excluded.hour_18, hour_19 = excluded.hour_19,
			hour_20 = excluded.hour_20, hour_21 = excluded.hour_21, hour_22 = excluded.hour_22, hour_23 = excluded.hour_23,
			peak_hour = excluded.peak_hour, peak_score = excluded.peak_score,
			peak_start_hour = excluded.peak_start_hour, peak_end_hour = excluded.peak_end_hour,
			low_hour = excluded.low_hour, low_score = excluded.low_score,
			chronotype = excluded.chronotype, chronotype_confidence = excluded.chronotype_confidence,
			avg_efficiency = excluded.avg_efficiency, std_efficiency = excluded.std_efficiency,
			total_samples = excluded.total_samples,
			start_date = excluded.start_date, end_date = excluded.end_date,
			updated_at = CURRENT_TIMESTAMP
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		profile.ProfileType,
		profile.HourlyCurve[0], profile.HourlyCurve[1], profile.HourlyCurve[2], profile.HourlyCurve[3],
		profile.HourlyCurve[4], profile.HourlyCurve[5], profile.HourlyCurve[6], profile.HourlyCurve[7],
		profile.HourlyCurve[8], profile.HourlyCurve[9], profile.HourlyCurve[10], profile.HourlyCurve[11],
		profile.HourlyCurve[12], profile.HourlyCurve[13], profile.HourlyCurve[14], profile.HourlyCurve[15],
		profile.HourlyCurve[16], profile.HourlyCurve[17], profile.HourlyCurve[18], profile.HourlyCurve[19],
		profile.HourlyCurve[20], profile.HourlyCurve[21], profile.HourlyCurve[22], profile.HourlyCurve[23],
		profile.PeakHour, profile.PeakScore, profile.PeakStartHour, profile.PeakEndHour,
		profile.LowHour, profile.LowScore,
		profile.Chronotype, profile.ChronotypeConfidence,
		profile.AvgEfficiency, profile.StdEfficiency, profile.TotalSamples,
		profile.StartDate, profile.EndDate,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}
