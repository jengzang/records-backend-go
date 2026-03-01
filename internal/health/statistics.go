package health

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// StatisticsGenerator 统计数据生成器
type StatisticsGenerator struct {
	db *sql.DB
}

// NewStatisticsGenerator 创建统计生成器
func NewStatisticsGenerator(db *sql.DB) *StatisticsGenerator {
	return &StatisticsGenerator{db: db}
}

// GenerateDailyStatistics 生成每日统计
func (g *StatisticsGenerator) GenerateDailyStatistics(metricType string, startDate, endDate time.Time) error {
	query := `
	INSERT OR REPLACE INTO health_statistics
	(stat_type, stat_date, metric_type, total_value, avg_value, min_value, max_value, data, created_at)
	SELECT
		'daily' as stat_type,
		DATE(start_date) as stat_date,
		? as metric_type,
		SUM(value) as total_value,
		AVG(value) as avg_value,
		MIN(value) as min_value,
		MAX(value) as max_value,
		json_object(
			'count', COUNT(*),
			'first_record', MIN(start_date),
			'last_record', MAX(start_date)
		) as data,
		CURRENT_TIMESTAMP as created_at
	FROM health_records
	WHERE type LIKE ?
		AND DATE(start_date) >= DATE(?)
		AND DATE(start_date) <= DATE(?)
	GROUP BY DATE(start_date)
	`

	typePattern := fmt.Sprintf("%%Identifier%s%%", metricType)
	_, err := g.db.Exec(query, metricType, typePattern, startDate, endDate)
	return err
}

// GenerateWeeklyStatistics 生成每周统计
func (g *StatisticsGenerator) GenerateWeeklyStatistics(metricType string, startDate, endDate time.Time) error {
	query := `
	INSERT OR REPLACE INTO health_statistics
	(stat_type, stat_date, metric_type, total_value, avg_value, min_value, max_value, data, created_at)
	SELECT
		'weekly' as stat_type,
		strftime('%Y-W%W', start_date) as stat_date,
		? as metric_type,
		SUM(value) as total_value,
		AVG(value) as avg_value,
		MIN(value) as min_value,
		MAX(value) as max_value,
		json_object(
			'count', COUNT(*),
			'week_start', MIN(DATE(start_date)),
			'week_end', MAX(DATE(start_date))
		) as data,
		CURRENT_TIMESTAMP as created_at
	FROM health_records
	WHERE type LIKE ?
		AND DATE(start_date) >= DATE(?)
		AND DATE(start_date) <= DATE(?)
	GROUP BY strftime('%Y-W%W', start_date)
	`

	typePattern := fmt.Sprintf("%%Identifier%s%%", metricType)
	_, err := g.db.Exec(query, metricType, typePattern, startDate, endDate)
	return err
}

// GenerateMonthlyStatistics 生成每月统计
func (g *StatisticsGenerator) GenerateMonthlyStatistics(metricType string, startDate, endDate time.Time) error {
	query := `
	INSERT OR REPLACE INTO health_statistics
	(stat_type, stat_date, metric_type, total_value, avg_value, min_value, max_value, data, created_at)
	SELECT
		'monthly' as stat_type,
		strftime('%Y-%m', start_date) as stat_date,
		? as metric_type,
		SUM(value) as total_value,
		AVG(value) as avg_value,
		MIN(value) as min_value,
		MAX(value) as max_value,
		json_object(
			'count', COUNT(*),
			'month_start', MIN(DATE(start_date)),
			'month_end', MAX(DATE(start_date))
		) as data,
		CURRENT_TIMESTAMP as created_at
	FROM health_records
	WHERE type LIKE ?
		AND DATE(start_date) >= DATE(?)
		AND DATE(start_date) <= DATE(?)
	GROUP BY strftime('%Y-%m', start_date)
	`

	typePattern := fmt.Sprintf("%%Identifier%s%%", metricType)
	_, err := g.db.Exec(query, metricType, typePattern, startDate, endDate)
	return err
}

// GenerateAllStatistics 生成所有统计数据
func (g *StatisticsGenerator) GenerateAllStatistics() error {
	// 获取数据的日期范围
	var minDateStr, maxDateStr string
	err := g.db.QueryRow(`
		SELECT MIN(start_date), MAX(start_date)
		FROM health_records
	`).Scan(&minDateStr, &maxDateStr)
	if err != nil {
		return fmt.Errorf("failed to get date range: %w", err)
	}

	// 解析日期字符串
	minDate, err := time.Parse("2006-01-02 15:04:05", minDateStr)
	if err != nil {
		return fmt.Errorf("failed to parse min date: %w", err)
	}
	maxDate, err := time.Parse("2006-01-02 15:04:05", maxDateStr)
	if err != nil {
		return fmt.Errorf("failed to parse max date: %w", err)
	}

	// 常见的健康指标类型
	metricTypes := []string{
		"HeartRate",
		"StepCount",
		"DistanceWalkingRunning",
		"ActiveEnergyBurned",
		"BasalEnergyBurned",
		"FlightsClimbed",
		"BodyMass",
		"Height",
		"BodyMassIndex",
	}

	// 为每个指标生成统计
	for _, metricType := range metricTypes {
		// 检查是否有该类型的数据
		var count int
		typePattern := fmt.Sprintf("%%Identifier%s%%", metricType)
		err := g.db.QueryRow(`
			SELECT COUNT(*) FROM health_records WHERE type LIKE ?
		`, typePattern).Scan(&count)

		if err != nil || count == 0 {
			continue // 跳过没有数据的指标
		}

		// 生成每日统计
		if err := g.GenerateDailyStatistics(metricType, minDate, maxDate); err != nil {
			return fmt.Errorf("failed to generate daily stats for %s: %w", metricType, err)
		}

		// 生成每周统计
		if err := g.GenerateWeeklyStatistics(metricType, minDate, maxDate); err != nil {
			return fmt.Errorf("failed to generate weekly stats for %s: %w", metricType, err)
		}

		// 生成每月统计
		if err := g.GenerateMonthlyStatistics(metricType, minDate, maxDate); err != nil {
			return fmt.Errorf("failed to generate monthly stats for %s: %w", metricType, err)
		}
	}

	return nil
}

// GetStatisticsSummary 获取统计摘要
func (g *StatisticsGenerator) GetStatisticsSummary() (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 统计总数
	var totalStats int
	err := g.db.QueryRow(`SELECT COUNT(*) FROM health_statistics`).Scan(&totalStats)
	if err != nil {
		return nil, err
	}
	summary["total_statistics"] = totalStats

	// 按类型统计
	rows, err := g.db.Query(`
		SELECT stat_type, COUNT(*) as count
		FROM health_statistics
		GROUP BY stat_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsByType := make(map[string]int)
	for rows.Next() {
		var statType string
		var count int
		if err := rows.Scan(&statType, &count); err != nil {
			continue
		}
		statsByType[statType] = count
	}
	summary["by_type"] = statsByType

	// 按指标统计
	rows, err = g.db.Query(`
		SELECT metric_type, COUNT(*) as count
		FROM health_statistics
		GROUP BY metric_type
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsByMetric := make(map[string]int)
	for rows.Next() {
		var metricType string
		var count int
		if err := rows.Scan(&metricType, &count); err != nil {
			continue
		}
		statsByMetric[metricType] = count
	}
	summary["by_metric"] = statsByMetric

	// 最后更新时间
	var lastUpdate sql.NullTime
	err = g.db.QueryRow(`
		SELECT MAX(created_at) FROM health_statistics
	`).Scan(&lastUpdate)
	if err == nil && lastUpdate.Valid {
		summary["last_update"] = lastUpdate.Time
	}

	return summary, nil
}

// RegenerateStatistics 重新生成统计(清除旧数据)
func (g *StatisticsGenerator) RegenerateStatistics() error {
	// 清除旧统计
	_, err := g.db.Exec(`DELETE FROM health_statistics`)
	if err != nil {
		return fmt.Errorf("failed to clear old statistics: %w", err)
	}

	// 生成新统计
	return g.GenerateAllStatistics()
}

// IncrementalUpdate 增量更新统计(仅更新最近的数据)
func (g *StatisticsGenerator) IncrementalUpdate(days int) error {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	metricTypes := []string{
		"HeartRate",
		"StepCount",
		"DistanceWalkingRunning",
		"ActiveEnergyBurned",
		"BasalEnergyBurned",
		"FlightsClimbed",
		"BodyMass",
		"Height",
		"BodyMassIndex",
	}

	for _, metricType := range metricTypes {
		// 检查是否有该类型的数据
		var count int
		typePattern := fmt.Sprintf("%%Identifier%s%%", metricType)
		err := g.db.QueryRow(`
			SELECT COUNT(*) FROM health_records
			WHERE type LIKE ?
			AND DATE(start_date) >= DATE(?)
		`, typePattern, startDate).Scan(&count)

		if err != nil || count == 0 {
			continue
		}

		// 更新统计
		if err := g.GenerateDailyStatistics(metricType, startDate, endDate); err != nil {
			return err
		}
		if err := g.GenerateWeeklyStatistics(metricType, startDate, endDate); err != nil {
			return err
		}
		if err := g.GenerateMonthlyStatistics(metricType, startDate, endDate); err != nil {
			return err
		}
	}

	return nil
}

// StatisticsData 统计数据结构
type StatisticsData struct {
	Count       int       `json:"count"`
	FirstRecord time.Time `json:"first_record,omitempty"`
	LastRecord  time.Time `json:"last_record,omitempty"`
	WeekStart   string    `json:"week_start,omitempty"`
	WeekEnd     string    `json:"week_end,omitempty"`
	MonthStart  string    `json:"month_start,omitempty"`
	MonthEnd    string    `json:"month_end,omitempty"`
}

// ParseStatisticsData 解析统计数据JSON
func ParseStatisticsData(dataJSON string) (*StatisticsData, error) {
	var data StatisticsData
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return nil, err
	}
	return &data, nil
}
