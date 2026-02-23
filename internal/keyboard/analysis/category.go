package analysis

import (
	"database/sql"
	"fmt"
)

// CategoryAnalyzer handles key category analysis
type CategoryAnalyzer struct {
	db *sql.DB
}

// NewCategoryAnalyzer creates a new category analyzer
func NewCategoryAnalyzer(db *sql.DB) *CategoryAnalyzer {
	return &CategoryAnalyzer{db: db}
}

// CategoryDistribution represents distribution of a key category
type CategoryDistribution struct {
	Category   string  `json:"category"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
	UniqueKeys int     `json:"uniqueKeys"`
}

// TopKey represents a top key in a category
type TopKey struct {
	Scancode int    `json:"scancode"`
	KeyName  string `json:"keyName"`
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// ModifierUsage represents modifier key usage patterns
type ModifierUsage struct {
	Ctrl  int64 `json:"ctrl"`
	Shift int64 `json:"shift"`
	Alt   int64 `json:"alt"`
	Win   int64 `json:"win"`
}

// AnalyzeCategoryDistribution analyzes key usage by category
func (ca *CategoryAnalyzer) AnalyzeCategoryDistribution(startDate, endDate string) ([]CategoryDistribution, error) {
	query := `
		SELECT
			COALESCE(m.key_category, 'unknown') as category,
			SUM(s.count) as total_count,
			COUNT(DISTINCT s.scancode) as unique_keys
		FROM scancode_stats s
		LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND s.date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND s.date <= ?"
		args = append(args, endDate)
	}

	query += " GROUP BY m.key_category ORDER BY total_count DESC"

	rows, err := ca.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query category distribution: %w", err)
	}
	defer rows.Close()

	// First pass: collect data and calculate total
	var distributions []CategoryDistribution
	var totalKeystrokes int64

	for rows.Next() {
		var dist CategoryDistribution
		if err := rows.Scan(&dist.Category, &dist.Count, &dist.UniqueKeys); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		distributions = append(distributions, dist)
		totalKeystrokes += dist.Count
	}

	// Second pass: calculate percentages
	for i := range distributions {
		if totalKeystrokes > 0 {
			distributions[i].Percentage = float64(distributions[i].Count) / float64(totalKeystrokes) * 100
		}
	}

	return distributions, nil
}

// AnalyzeTopKeysByCategory gets top N keys for a specific category
func (ca *CategoryAnalyzer) AnalyzeTopKeysByCategory(category string, limit int, startDate, endDate string) ([]TopKey, error) {
	query := `
		SELECT
			s.scancode,
			m.key_name,
			m.key_category,
			SUM(s.count) as total_count
		FROM scancode_stats s
		LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
		WHERE m.key_category = ?
	`
	args := []interface{}{category}

	if startDate != "" {
		query += " AND s.date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND s.date <= ?"
		args = append(args, endDate)
	}

	query += " GROUP BY s.scancode, m.key_name, m.key_category ORDER BY total_count DESC LIMIT ?"
	args = append(args, limit)

	rows, err := ca.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top keys: %w", err)
	}
	defer rows.Close()

	var topKeys []TopKey
	for rows.Next() {
		var key TopKey
		if err := rows.Scan(&key.Scancode, &key.KeyName, &key.Category, &key.Count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		topKeys = append(topKeys, key)
	}

	return topKeys, nil
}

// AnalyzeAllTopKeys gets top keys for all categories
func (ca *CategoryAnalyzer) AnalyzeAllTopKeys(limit int, startDate, endDate string) (map[string][]TopKey, error) {
	categories := []string{"letter", "number", "function", "modifier", "special"}
	result := make(map[string][]TopKey)

	for _, category := range categories {
		topKeys, err := ca.AnalyzeTopKeysByCategory(category, limit, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze category %s: %w", category, err)
		}
		result[category] = topKeys
	}

	return result, nil
}

// AnalyzeModifierUsage analyzes modifier key usage patterns
func (ca *CategoryAnalyzer) AnalyzeModifierUsage(startDate, endDate string) (*ModifierUsage, error) {
	query := `
		SELECT
			m.key_name,
			SUM(s.count) as total_count
		FROM scancode_stats s
		LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
		WHERE m.key_category = 'modifier'
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND s.date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND s.date <= ?"
		args = append(args, endDate)
	}

	query += " GROUP BY m.key_name ORDER BY total_count DESC"

	rows, err := ca.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifier usage: %w", err)
	}
	defer rows.Close()

	usage := &ModifierUsage{}

	for rows.Next() {
		var keyName string
		var count int64

		if err := rows.Scan(&keyName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Aggregate by modifier type
		switch {
		case keyName == "Left Ctrl" || keyName == "Right Ctrl":
			usage.Ctrl += count
		case keyName == "Left Shift" || keyName == "Right Shift":
			usage.Shift += count
		case keyName == "Left Alt" || keyName == "Right Alt":
			usage.Alt += count
		case keyName == "Left Win" || keyName == "Right Win":
			usage.Win += count
		}
	}

	return usage, nil
}
