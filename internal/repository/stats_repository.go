package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jengzang/records-backend-go/internal/models"
)

// StatsRepository handles database operations for statistics
type StatsRepository struct {
	db *sql.DB
}

// NewStatsRepository creates a new stats repository
func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetFootprintStatistics retrieves footprint statistics for a time range
func (r *StatsRepository) GetFootprintStatistics(startTime, endTime int64) (*models.FootprintStatistics, error) {
	stats := &models.FootprintStatistics{
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}

	if startTime > 0 {
		conditions = append(conditions, "dataTime >= ?")
		args = append(args, startTime)
	}
	if endTime > 0 {
		conditions = append(conditions, "dataTime <= ?")
		args = append(args, endTime)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total points
	query := `SELECT COUNT(*) FROM "一生足迹"` + whereClause
	err := r.db.QueryRow(query, args...).Scan(&stats.TotalPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to count total points: %w", err)
	}

	// Get province count and list
	query = `SELECT COUNT(DISTINCT province), GROUP_CONCAT(DISTINCT province)
		FROM "一生足迹"` + whereClause + ` AND province IS NOT NULL AND province != ''`
	var provinceList sql.NullString
	err = r.db.QueryRow(query, args...).Scan(&stats.ProvinceCount, &provinceList)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get province stats: %w", err)
	}
	if provinceList.Valid {
		stats.Provinces = strings.Split(provinceList.String, ",")
	}

	// Get city count and list
	query = `SELECT COUNT(DISTINCT city), GROUP_CONCAT(DISTINCT city)
		FROM "一生足迹"` + whereClause + ` AND city IS NOT NULL AND city != ''`
	var cityList sql.NullString
	err = r.db.QueryRow(query, args...).Scan(&stats.CityCount, &cityList)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get city stats: %w", err)
	}
	if cityList.Valid {
		stats.Cities = strings.Split(cityList.String, ",")
	}

	// Get county count and list
	query = `SELECT COUNT(DISTINCT county), GROUP_CONCAT(DISTINCT county)
		FROM "一生足迹"` + whereClause + ` AND county IS NOT NULL AND county != ''`
	var countyList sql.NullString
	err = r.db.QueryRow(query, args...).Scan(&stats.CountyCount, &countyList)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get county stats: %w", err)
	}
	if countyList.Valid {
		stats.Counties = strings.Split(countyList.String, ",")
	}

	// Get town count
	query = `SELECT COUNT(DISTINCT town) FROM "一生足迹"` + whereClause + ` AND town IS NOT NULL AND town != ''`
	err = r.db.QueryRow(query, args...).Scan(&stats.TownCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get town count: %w", err)
	}

	// Get village count
	query = `SELECT COUNT(DISTINCT village) FROM "一生足迹"` + whereClause + ` AND village IS NOT NULL AND village != ''`
	err = r.db.QueryRow(query, args...).Scan(&stats.VillageCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get village count: %w", err)
	}

	return stats, nil
}

// GetTimeDistribution retrieves time distribution statistics
func (r *StatsRepository) GetTimeDistribution(startTime, endTime int64) ([]models.TimeDistribution, error) {
	query := `SELECT
		CAST(strftime('%H', datetime(dataTime, 'unixepoch')) AS INTEGER) as hour,
		COUNT(*) as count
		FROM "一生足迹"
		WHERE dataTime >= ? AND dataTime <= ?
		GROUP BY hour
		ORDER BY hour`

	rows, err := r.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query time distribution: %w", err)
	}
	defer rows.Close()

	var distribution []models.TimeDistribution
	for rows.Next() {
		var td models.TimeDistribution
		if err := rows.Scan(&td.Hour, &td.Count); err != nil {
			return nil, fmt.Errorf("failed to scan time distribution: %w", err)
		}
		distribution = append(distribution, td)
	}

	return distribution, nil
}

// GetSpeedDistribution retrieves speed distribution statistics
func (r *StatsRepository) GetSpeedDistribution(startTime, endTime int64) ([]models.SpeedDistribution, error) {
	query := `SELECT
		CASE
			WHEN speed < 10 THEN '0-10'
			WHEN speed < 30 THEN '10-30'
			WHEN speed < 60 THEN '30-60'
			WHEN speed < 120 THEN '60-120'
			ELSE '120+'
		END as speed_range,
		COUNT(*) as count
		FROM "一生足迹"
		WHERE dataTime >= ? AND dataTime <= ? AND speed > 0
		GROUP BY speed_range
		ORDER BY
			CASE speed_range
				WHEN '0-10' THEN 1
				WHEN '10-30' THEN 2
				WHEN '30-60' THEN 3
				WHEN '60-120' THEN 4
				WHEN '120+' THEN 5
			END`

	rows, err := r.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query speed distribution: %w", err)
	}
	defer rows.Close()

	var distribution []models.SpeedDistribution
	for rows.Next() {
		var sd models.SpeedDistribution
		if err := rows.Scan(&sd.SpeedRange, &sd.Count); err != nil {
			return nil, fmt.Errorf("failed to scan speed distribution: %w", err)
		}
		distribution = append(distribution, sd)
	}

	return distribution, nil
}

// GetFootprintRankings retrieves footprint statistics with rankings
func (r *StatsRepository) GetFootprintRankings(filter models.StatsFilter) ([]models.FootprintStatistics, error) {
	// Build query
	query := `SELECT id, stat_type, stat_key, time_range,
		province, city, county, town,
		point_count, visit_count, total_distance_meters, total_duration_seconds,
		first_visit_time, last_visit_time,
		rank_by_points, rank_by_visits, rank_by_duration,
		algo_version, created_at, updated_at
		FROM footprint_statistics`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.StatType != "" {
		conditions = append(conditions, "stat_type = ?")
		args = append(args, filter.StatType)
	}
	if filter.TimeRange != "" {
		conditions = append(conditions, "time_range = ?")
		args = append(args, filter.TimeRange)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by
	orderBy := "point_count DESC"
	if filter.OrderBy == "visits" {
		orderBy = "visit_count DESC"
	} else if filter.OrderBy == "duration" {
		orderBy = "total_duration_seconds DESC"
	} else if filter.OrderBy == "distance" {
		orderBy = "total_distance_meters DESC"
	}
	query += " ORDER BY " + orderBy

	// Limit
	limit := 100
	if filter.Limit > 0 && filter.Limit <= 1000 {
		limit = filter.Limit
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query footprint rankings: %w", err)
	}
	defer rows.Close()

	var stats []models.FootprintStatistics
	for rows.Next() {
		var s models.FootprintStatistics
		err := rows.Scan(
			&s.ID, &s.StatType, &s.StatKey, &s.TimeRange,
			&s.Province, &s.City, &s.County, &s.Town,
			&s.PointCount, &s.VisitCount, &s.TotalDistanceMeters, &s.TotalDurationSeconds,
			&s.FirstVisitTime, &s.LastVisitTime,
			&s.RankByPoints, &s.RankByVisits, &s.RankByDuration,
			&s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan footprint statistics: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetStayRankings retrieves stay statistics with rankings
func (r *StatsRepository) GetStayRankings(filter models.StatsFilter) ([]models.StayStatistics, error) {
	// Build query
	query := `SELECT id, stat_type, stat_key, time_range,
		province, city, county,
		stay_count, total_duration_seconds, avg_duration_seconds, max_duration_seconds,
		stay_category, rank_by_count, rank_by_duration,
		algo_version, created_at, updated_at
		FROM stay_statistics`

	var conditions []string
	var args []interface{}

	// Add filters
	if filter.StatType != "" {
		conditions = append(conditions, "stat_type = ?")
		args = append(args, filter.StatType)
	}
	if filter.TimeRange != "" {
		conditions = append(conditions, "time_range = ?")
		args = append(args, filter.TimeRange)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by
	orderBy := "stay_count DESC"
	if filter.OrderBy == "duration" {
		orderBy = "total_duration_seconds DESC"
	}
	query += " ORDER BY " + orderBy

	// Limit
	limit := 100
	if filter.Limit > 0 && filter.Limit <= 1000 {
		limit = filter.Limit
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stay rankings: %w", err)
	}
	defer rows.Close()

	var stats []models.StayStatistics
	for rows.Next() {
		var s models.StayStatistics
		err := rows.Scan(
			&s.ID, &s.StatType, &s.StatKey, &s.TimeRange,
			&s.Province, &s.City, &s.County,
			&s.StayCount, &s.TotalDurationSeconds, &s.AvgDurationSeconds, &s.MaxDurationSeconds,
			&s.StayCategory, &s.RankByCount, &s.RankByDuration,
			&s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stay statistics: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetExtremeEvents retrieves extreme events
func (r *StatsRepository) GetExtremeEvents(eventType, eventCategory string, limit int) ([]models.ExtremeEvent, error) {
	// Build query - use actual column names from database
	query := `SELECT id, event_type,
		COALESCE(event_category, '') as event_category,
		point_id,
		timestamp as event_time,
		value as event_value,
		latitude, longitude,
		COALESCE(province, '') as province,
		COALESCE(city, '') as city,
		COALESCE(county, '') as county,
		COALESCE(mode, '') as mode,
		COALESCE(segment_id, 0) as segment_id,
		COALESCE(rank, 0) as rank,
		COALESCE(algo_version, 'v1') as algo_version,
		created_at, updated_at
		FROM extreme_events`

	var conditions []string
	var args []interface{}

	// Add filters
	if eventType != "" {
		conditions = append(conditions, "event_type = ?")
		args = append(args, eventType)
	}
	if eventCategory != "" {
		conditions = append(conditions, "event_category = ?")
		args = append(args, eventCategory)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by rank (or value if rank is not set)
	query += " ORDER BY COALESCE(rank, 999999) ASC, value DESC"

	// Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query extreme events: %w", err)
	}
	defer rows.Close()

	var events []models.ExtremeEvent
	for rows.Next() {
		var e models.ExtremeEvent
		err := rows.Scan(
			&e.ID, &e.EventType, &e.EventCategory, &e.PointID, &e.EventTime, &e.EventValue,
			&e.Latitude, &e.Longitude, &e.Province, &e.City, &e.County,
			&e.Mode, &e.SegmentID, &e.Rank,
			&e.AlgoVersion, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan extreme event: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

// GetAdminCrossings retrieves administrative boundary crossing events
func (r *StatsRepository) GetAdminCrossings(crossingType, fromRegion, toRegion string, startTime, endTime int64, limit int) ([]models.AdminCrossing, error) {
	// Build query
	query := `SELECT id, crossing_ts, from_province, from_city, from_county, from_town,
		to_province, to_city, to_county, to_town, crossing_type,
		latitude, longitude, distance_from_prev_m, algo_version, created_at
		FROM admin_crossings`

	var conditions []string
	var args []interface{}

	// Add filters
	if crossingType != "" {
		conditions = append(conditions, "crossing_type = ?")
		args = append(args, crossingType)
	}
	if startTime > 0 {
		conditions = append(conditions, "crossing_ts >= ?")
		args = append(args, startTime)
	}
	if endTime > 0 {
		conditions = append(conditions, "crossing_ts <= ?")
		args = append(args, endTime)
	}
	if fromRegion != "" {
		conditions = append(conditions, "(from_province = ? OR from_city = ? OR from_county = ? OR from_town = ?)")
		args = append(args, fromRegion, fromRegion, fromRegion, fromRegion)
	}
	if toRegion != "" {
		conditions = append(conditions, "(to_province = ? OR to_city = ? OR to_county = ? OR to_town = ?)")
		args = append(args, toRegion, toRegion, toRegion, toRegion)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by timestamp
	query += " ORDER BY crossing_ts DESC"

	// Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin crossings: %w", err)
	}
	defer rows.Close()

	var crossings []models.AdminCrossing
	for rows.Next() {
		var c models.AdminCrossing
		err := rows.Scan(
			&c.ID, &c.CrossingTS, &c.FromProvince, &c.FromCity, &c.FromCounty, &c.FromTown,
			&c.ToProvince, &c.ToCity, &c.ToCounty, &c.ToTown, &c.CrossingType,
			&c.Latitude, &c.Longitude, &c.DistanceFromPrevM, &c.AlgoVersion, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin crossing: %w", err)
		}
		crossings = append(crossings, c)
	}

	return crossings, nil
}

// GetAdminStats retrieves administrative region statistics
func (r *StatsRepository) GetAdminStats(adminLevel, adminName, parentName, sortBy string, limit int) ([]models.AdminStats, error) {
	// Build query
	query := `SELECT id, admin_level, admin_name, parent_name, visit_count,
		total_duration_s, unique_days, first_visit_ts, last_visit_ts,
		total_distance_m, algo_version, created_at, updated_at
		FROM admin_stats`

	var conditions []string
	var args []interface{}

	// Add filters
	if adminLevel != "" {
		conditions = append(conditions, "admin_level = ?")
		args = append(args, adminLevel)
	}
	if adminName != "" {
		conditions = append(conditions, "admin_name = ?")
		args = append(args, adminName)
	}
	if parentName != "" {
		conditions = append(conditions, "parent_name = ?")
		args = append(args, parentName)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by
	orderBy := "visit_count DESC"
	if sortBy == "duration" {
		orderBy = "total_duration_s DESC"
	} else if sortBy == "unique_days" {
		orderBy = "unique_days DESC"
	} else if sortBy == "distance" {
		orderBy = "total_distance_m DESC"
	}
	query += " ORDER BY " + orderBy

	// Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin stats: %w", err)
	}
	defer rows.Close()

	var stats []models.AdminStats
	for rows.Next() {
		var s models.AdminStats
		err := rows.Scan(
			&s.ID, &s.AdminLevel, &s.AdminName, &s.ParentName, &s.VisitCount,
			&s.TotalDurationS, &s.UniqueDays, &s.FirstVisitTS, &s.LastVisitTS,
			&s.TotalDistanceM, &s.AlgoVersion, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}
// GetSpeedSpaceStats retrieves speed-space coupling statistics
func (r *StatsRepository) GetSpeedSpaceStats(bucketType, areaType, areaName string, limit int) ([]models.SpeedSpaceStats, error) {
	query := `SELECT id, bucket_type, bucket_key, area_type, area_key,
		avg_speed, speed_variance, speed_entropy, total_distance, segment_count,
		is_high_speed_zone, is_slow_life_zone, stay_intensity,
		algo_version, created_at
		FROM speed_space_stats_bucketed`

	var conditions []string
	var args []interface{}

	// Add filters
	if bucketType != "" {
		conditions = append(conditions, "bucket_type = ?")
		args = append(args, bucketType)
	}
	if areaType != "" {
		conditions = append(conditions, "area_type = ?")
		args = append(args, areaType)
	}
	if areaName != "" {
		conditions = append(conditions, "area_key = ?")
		args = append(args, areaName)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by average speed descending
	query += " ORDER BY avg_speed DESC"

	// Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query speed-space stats: %w", err)
	}
	defer rows.Close()

	var stats []models.SpeedSpaceStats
	for rows.Next() {
		var s models.SpeedSpaceStats
		err := rows.Scan(
			&s.ID, &s.BucketType, &s.BucketKey, &s.AreaType, &s.AreaKey,
			&s.AvgSpeed, &s.SpeedVariance, &s.SpeedEntropy, &s.TotalDistance, &s.SegmentCount,
			&s.IsHighSpeedZone, &s.IsSlowLifeZone, &s.StayIntensity,
			&s.AlgoVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan speed-space stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetHighSpeedZones retrieves high-speed zones
func (r *StatsRepository) GetHighSpeedZones(bucketType, areaType string, limit int) ([]models.SpeedSpaceStats, error) {
	query := `SELECT id, bucket_type, bucket_key, area_type, area_key,
		avg_speed, speed_variance, speed_entropy, total_distance, segment_count,
		is_high_speed_zone, is_slow_life_zone, stay_intensity,
		algo_version, created_at
		FROM speed_space_stats_bucketed
		WHERE is_high_speed_zone = 1`

	var args []interface{}

	// Add filters
	if bucketType != "" {
		query += " AND bucket_type = ?"
		args = append(args, bucketType)
	}
	if areaType != "" {
		query += " AND area_type = ?"
		args = append(args, areaType)
	}

	// Order by average speed descending
	query += " ORDER BY avg_speed DESC"

	// Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query high-speed zones: %w", err)
	}
	defer rows.Close()

	var zones []models.SpeedSpaceStats
	for rows.Next() {
		var s models.SpeedSpaceStats
		err := rows.Scan(
			&s.ID, &s.BucketType, &s.BucketKey, &s.AreaType, &s.AreaKey,
			&s.AvgSpeed, &s.SpeedVariance, &s.SpeedEntropy, &s.TotalDistance, &s.SegmentCount,
			&s.IsHighSpeedZone, &s.IsSlowLifeZone, &s.StayIntensity,
			&s.AlgoVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan high-speed zone: %w", err)
		}
		zones = append(zones, s)
	}

	return zones, nil
}

// GetSlowLifeZones retrieves slow-life zones
func (r *StatsRepository) GetSlowLifeZones(bucketType, areaType string, limit int) ([]models.SpeedSpaceStats, error) {
	query := `SELECT id, bucket_type, bucket_key, area_type, area_key,
		avg_speed, speed_variance, speed_entropy, total_distance, segment_count,
		is_high_speed_zone, is_slow_life_zone, stay_intensity,
		algo_version, created_at
		FROM speed_space_stats_bucketed
		WHERE is_slow_life_zone = 1`

	var args []interface{}

	// Add filters
	if bucketType != "" {
		query += " AND bucket_type = ?"
		args = append(args, bucketType)
	}
	if areaType != "" {
		query += " AND area_type = ?"
		args = append(args, areaType)
	}

	// Order by average speed ascending (slowest first)
	query += " ORDER BY avg_speed ASC"

	// Limit
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query slow-life zones: %w", err)
	}
	defer rows.Close()

	var zones []models.SpeedSpaceStats
	for rows.Next() {
		var s models.SpeedSpaceStats
		err := rows.Scan(
			&s.ID, &s.BucketType, &s.BucketKey, &s.AreaType, &s.AreaKey,
			&s.AvgSpeed, &s.SpeedVariance, &s.SpeedEntropy, &s.TotalDistance, &s.SegmentCount,
			&s.IsHighSpeedZone, &s.IsSlowLifeZone, &s.StayIntensity,
			&s.AlgoVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slow-life zone: %w", err)
		}
		zones = append(zones, s)
	}

	return zones, nil
}

// GetDirectionalBiasStats retrieves directional bias statistics
func (r *StatsRepository) GetDirectionalBiasStats(
	bucketType, areaType, areaKey, modeFilter string,
	limit int,
) ([]models.DirectionalBiasStats, error) {
	query := `
		SELECT
			id, bucket_type, bucket_key, area_type, area_key, mode_filter,
			direction_histogram_json, num_bins,
			dominant_direction_deg, directional_concentration,
			bidirectional_score, directional_entropy,
			total_distance, total_duration, segment_count,
			algo_version, created_at
		FROM directional_stats_bucketed
		WHERE 1=1
	`

	var args []interface{}
	if bucketType != "" {
		query += " AND bucket_type = ?"
		args = append(args, bucketType)
	}
	if areaType != "" {
		query += " AND area_type = ?"
		args = append(args, areaType)
	}
	if areaKey != "" {
		query += " AND area_key = ?"
		args = append(args, areaKey)
	}
	if modeFilter != "" {
		query += " AND mode_filter = ?"
		args = append(args, modeFilter)
	}

	query += " ORDER BY total_distance DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query directional bias stats: %w", err)
	}
	defer rows.Close()

	var stats []models.DirectionalBiasStats
	for rows.Next() {
		var s models.DirectionalBiasStats
		err := rows.Scan(
			&s.ID, &s.BucketType, &s.BucketKey, &s.AreaType, &s.AreaKey, &s.ModeFilter,
			&s.DirectionHistogramJSON, &s.NumBins,
			&s.DominantDirectionDeg, &s.DirectionalConcentration,
			&s.BidirectionalScore, &s.DirectionalEntropy,
			&s.TotalDistance, &s.TotalDuration, &s.SegmentCount,
			&s.AlgoVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan directional bias stat: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetTopDirectionalAreas retrieves areas with highest directional concentration
func (r *StatsRepository) GetTopDirectionalAreas(
	bucketType string,
	limit int,
) ([]models.DirectionalBiasStats, error) {
	query := `
		SELECT
			id, bucket_type, bucket_key, area_type, area_key, mode_filter,
			direction_histogram_json, num_bins,
			dominant_direction_deg, directional_concentration,
			bidirectional_score, directional_entropy,
			total_distance, total_duration, segment_count,
			algo_version, created_at
		FROM directional_stats_bucketed
		WHERE bucket_type = ?
			AND mode_filter = 'ALL'
		ORDER BY directional_concentration DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, bucketType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top directional areas: %w", err)
	}
	defer rows.Close()

	var stats []models.DirectionalBiasStats
	for rows.Next() {
		var s models.DirectionalBiasStats
		err := rows.Scan(
			&s.ID, &s.BucketType, &s.BucketKey, &s.AreaType, &s.AreaKey, &s.ModeFilter,
			&s.DirectionHistogramJSON, &s.NumBins,
			&s.DominantDirectionDeg, &s.DirectionalConcentration,
			&s.BidirectionalScore, &s.DirectionalEntropy,
			&s.TotalDistance, &s.TotalDuration, &s.SegmentCount,
			&s.AlgoVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top directional area: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetBidirectionalPatterns retrieves areas with strong bidirectional patterns
func (r *StatsRepository) GetBidirectionalPatterns(
	bucketType string,
	limit int,
) ([]models.DirectionalBiasStats, error) {
	query := `
		SELECT
			id, bucket_type, bucket_key, area_type, area_key, mode_filter,
			direction_histogram_json, num_bins,
			dominant_direction_deg, directional_concentration,
			bidirectional_score, directional_entropy,
			total_distance, total_duration, segment_count,
			algo_version, created_at
		FROM directional_stats_bucketed
		WHERE bucket_type = ?
			AND mode_filter = 'ALL'
		ORDER BY bidirectional_score DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, bucketType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query bidirectional patterns: %w", err)
	}
	defer rows.Close()

	var stats []models.DirectionalBiasStats
	for rows.Next() {
		var s models.DirectionalBiasStats
		err := rows.Scan(
			&s.ID, &s.BucketType, &s.BucketKey, &s.AreaType, &s.AreaKey, &s.ModeFilter,
			&s.DirectionHistogramJSON, &s.NumBins,
			&s.DominantDirectionDeg, &s.DirectionalConcentration,
			&s.BidirectionalScore, &s.DirectionalEntropy,
			&s.TotalDistance, &s.TotalDuration, &s.SegmentCount,
			&s.AlgoVersion, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bidirectional pattern: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetRevisitPatterns retrieves revisit patterns with filters
func (r *StatsRepository) GetRevisitPatterns(
	minVisits int,
	habitualOnly bool,
	periodicOnly bool,
	limit int,
) ([]models.RevisitPattern, error) {
	query := `
		SELECT
			id, geohash6, center_lat, center_lon,
			province, city, county,
			visit_count, first_visit, last_visit, total_duration_seconds,
			avg_interval_days, std_interval_days, min_interval_days, max_interval_days,
			regularity_score, is_periodic, is_habitual, revisit_strength,
			algo_version, created_at, updated_at
		FROM revisit_patterns
		WHERE visit_count >= ?
	`

	var args []interface{}
	args = append(args, minVisits)

	if habitualOnly {
		query += " AND is_habitual = 1"
	}
	if periodicOnly {
		query += " AND is_periodic = 1"
	}

	query += " ORDER BY revisit_strength DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query revisit patterns: %w", err)
	}
	defer rows.Close()

	var patterns []models.RevisitPattern
	for rows.Next() {
		var p models.RevisitPattern
		var isPeriodic, isHabitual int
		err := rows.Scan(
			&p.ID, &p.Geohash6, &p.CenterLat, &p.CenterLon,
			&p.Province, &p.City, &p.County,
			&p.VisitCount, &p.FirstVisit, &p.LastVisit, &p.TotalDurationSeconds,
			&p.AvgIntervalDays, &p.StdIntervalDays, &p.MinIntervalDays, &p.MaxIntervalDays,
			&p.RegularityScore, &isPeriodic, &isHabitual, &p.RevisitStrength,
			&p.AlgoVersion, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan revisit pattern: %w", err)
		}
		p.IsPeriodic = isPeriodic == 1
		p.IsHabitual = isHabitual == 1
		patterns = append(patterns, p)
	}

	return patterns, nil
}

// GetTopRevisitLocations retrieves locations with highest revisit strength
func (r *StatsRepository) GetTopRevisitLocations(limit int) ([]models.RevisitPattern, error) {
	return r.GetRevisitPatterns(2, false, false, limit)
}

// GetHabitualLocations retrieves habitual locations (≥5 visits + high regularity)
func (r *StatsRepository) GetHabitualLocations(limit int) ([]models.RevisitPattern, error) {
	return r.GetRevisitPatterns(5, true, false, limit)
}

// GetPeriodicLocations retrieves locations with periodic visit patterns
func (r *StatsRepository) GetPeriodicLocations(limit int) ([]models.RevisitPattern, error) {
	return r.GetRevisitPatterns(3, false, true, limit)
}
