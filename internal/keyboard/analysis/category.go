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
			s.scan_code,
			SUM(s.count) as total_count
		FROM scan_codes s
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

	query += " GROUP BY s.scan_code ORDER BY total_count DESC"

	rows, err := ca.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query category distribution: %w", err)
	}
	defer rows.Close()

	// Collect data and group by category using in-memory mapping
	categoryMap := make(map[string]*CategoryDistribution)
	var totalKeystrokes int64

	for rows.Next() {
		var scancode int
		var count int64

		if err := rows.Scan(&scancode, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Get category from in-memory mapping
		category := GetKeyCategory(scancode)

		if _, exists := categoryMap[category]; !exists {
			categoryMap[category] = &CategoryDistribution{
				Category:   category,
				Count:      0,
				UniqueKeys: 0,
			}
		}

		categoryMap[category].Count += count
		categoryMap[category].UniqueKeys++
		totalKeystrokes += count
	}

	// Convert map to slice and calculate percentages
	var distributions []CategoryDistribution
	for _, dist := range categoryMap {
		if totalKeystrokes > 0 {
			dist.Percentage = float64(dist.Count) / float64(totalKeystrokes) * 100
		}
		distributions = append(distributions, *dist)
	}

	// Sort by count descending
	for i := 0; i < len(distributions)-1; i++ {
		for j := i + 1; j < len(distributions); j++ {
			if distributions[j].Count > distributions[i].Count {
				distributions[i], distributions[j] = distributions[j], distributions[i]
			}
		}
	}

	return distributions, nil
}

// AnalyzeTopKeysByCategory gets top N keys for a specific category
func (ca *CategoryAnalyzer) AnalyzeTopKeysByCategory(category string, limit int, startDate, endDate string) ([]TopKey, error) {
	query := `
		SELECT
			s.scan_code,
			SUM(s.count) as total_count
		FROM scan_codes s
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

	query += " GROUP BY s.scan_code ORDER BY total_count DESC"

	rows, err := ca.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top keys: %w", err)
	}
	defer rows.Close()

	var topKeys []TopKey
	for rows.Next() {
		var scancode int
		var count int64

		if err := rows.Scan(&scancode, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Get key info from in-memory mapping
		keyCategory := GetKeyCategory(scancode)

		// Filter by category
		if keyCategory == category {
			topKeys = append(topKeys, TopKey{
				Scancode: scancode,
				KeyName:  GetKeyName(scancode),
				Category: keyCategory,
				Count:    count,
			})

			// Stop when we reach the limit
			if len(topKeys) >= limit {
				break
			}
		}
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
			s.scan_code,
			SUM(s.count) as total_count
		FROM scan_codes s
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

	query += " GROUP BY s.scan_code ORDER BY total_count DESC"

	rows, err := ca.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifier usage: %w", err)
	}
	defer rows.Close()

	usage := &ModifierUsage{}

	for rows.Next() {
		var scancode int
		var count int64

		if err := rows.Scan(&scancode, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Get key info from in-memory mapping
		keyName := GetKeyName(scancode)
		keyCategory := GetKeyCategory(scancode)

		// Only process modifier keys
		if keyCategory != "modifier" {
			continue
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
