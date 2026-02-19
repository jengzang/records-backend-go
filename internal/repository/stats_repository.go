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
