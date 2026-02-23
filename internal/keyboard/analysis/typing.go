package analysis

import (
	"database/sql"
	"fmt"
)

// TypingBehaviorAnalyzer handles typing behavior analysis
type TypingBehaviorAnalyzer struct {
	db *sql.DB
}

// NewTypingBehaviorAnalyzer creates a new typing behavior analyzer
func NewTypingBehaviorAnalyzer(db *sql.DB) *TypingBehaviorAnalyzer {
	return &TypingBehaviorAnalyzer{db: db}
}

// TypingMetrics represents comprehensive typing behavior metrics
type TypingMetrics struct {
	TotalKeystrokes  int64   `json:"totalKeystrokes"`
	BackspaceCount   int64   `json:"backspaceCount"`
	EnterCount       int64   `json:"enterCount"`
	SpaceCount       int64   `json:"spaceCount"`
	DeleteCount      int64   `json:"deleteCount"`
	BackspaceRatio   float64 `json:"backspaceRatio"`
	DeleteRatio      float64 `json:"deleteRatio"`
	CorrectionRatio  float64 `json:"correctionRatio"`
	EstimatedWords   int64   `json:"estimatedWords"`
	EstimatedLines   int64   `json:"estimatedLines"`
	AvgWordLength    float64 `json:"avgWordLength"`
}

// SpecialKeyUsage represents usage of a special key
type SpecialKeyUsage struct {
	KeyName string `json:"keyName"`
	Count   int64  `json:"count"`
}

// LetterFrequency represents frequency of a letter
type LetterFrequency struct {
	Letter     string  `json:"letter"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// AnalyzeTypingMetrics calculates comprehensive typing behavior metrics
func (tba *TypingBehaviorAnalyzer) AnalyzeTypingMetrics(startDate, endDate string) (*TypingMetrics, error) {
	// Get total keystrokes
	totalQuery := `
		SELECT SUM(keystrokes) as total_keystrokes
		FROM daily_stats
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		totalQuery += " AND date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		totalQuery += " AND date <= ?"
		args = append(args, endDate)
	}

	var totalKeystrokes sql.NullInt64
	if err := tba.db.QueryRow(totalQuery, args...).Scan(&totalKeystrokes); err != nil {
		return nil, fmt.Errorf("failed to query total keystrokes: %w", err)
	}

	metrics := &TypingMetrics{
		TotalKeystrokes: totalKeystrokes.Int64,
	}

	// Get specific key counts
	keyQuery := `
		SELECT
			m.key_name,
			SUM(s.count) as total_count
		FROM scancode_stats s
		LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
		WHERE m.key_name IN ('Backspace', 'Enter', 'Space', 'Delete')
	`
	keyArgs := []interface{}{}

	if startDate != "" {
		keyQuery += " AND s.date >= ?"
		keyArgs = append(keyArgs, startDate)
	}
	if endDate != "" {
		keyQuery += " AND s.date <= ?"
		keyArgs = append(keyArgs, endDate)
	}

	keyQuery += " GROUP BY m.key_name"

	rows, err := tba.db.Query(keyQuery, keyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query key counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var keyName string
		var count int64

		if err := rows.Scan(&keyName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		switch keyName {
		case "Backspace":
			metrics.BackspaceCount = count
		case "Enter":
			metrics.EnterCount = count
		case "Space":
			metrics.SpaceCount = count
		case "Delete":
			metrics.DeleteCount = count
		}
	}

	// Calculate ratios
	if metrics.TotalKeystrokes > 0 {
		metrics.BackspaceRatio = float64(metrics.BackspaceCount) / float64(metrics.TotalKeystrokes)
		metrics.DeleteRatio = float64(metrics.DeleteCount) / float64(metrics.TotalKeystrokes)
		metrics.CorrectionRatio = float64(metrics.BackspaceCount+metrics.DeleteCount) / float64(metrics.TotalKeystrokes)
	}

	// Estimate words and lines
	if metrics.SpaceCount > 0 {
		metrics.EstimatedWords = metrics.SpaceCount + 1
	}
	if metrics.EnterCount > 0 {
		metrics.EstimatedLines = metrics.EnterCount + 1
	}

	// Calculate average word length
	if metrics.EstimatedWords > 0 {
		metrics.AvgWordLength = float64(metrics.TotalKeystrokes-metrics.SpaceCount) / float64(metrics.EstimatedWords)
	}

	return metrics, nil
}

// AnalyzeSpecialKeyUsage analyzes usage of special keys
func (tba *TypingBehaviorAnalyzer) AnalyzeSpecialKeyUsage(limit int, startDate, endDate string) ([]SpecialKeyUsage, error) {
	query := `
		SELECT
			m.key_name,
			SUM(s.count) as total_count
		FROM scancode_stats s
		LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
		WHERE m.key_category = 'special'
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

	query += " GROUP BY m.key_name ORDER BY total_count DESC LIMIT ?"
	args = append(args, limit)

	rows, err := tba.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query special key usage: %w", err)
	}
	defer rows.Close()

	var result []SpecialKeyUsage
	for rows.Next() {
		var usage SpecialKeyUsage
		if err := rows.Scan(&usage.KeyName, &usage.Count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, usage)
	}

	return result, nil
}

// AnalyzeLetterFrequency analyzes letter frequency distribution
func (tba *TypingBehaviorAnalyzer) AnalyzeLetterFrequency(startDate, endDate string) ([]LetterFrequency, error) {
	query := `
		SELECT
			m.key_name,
			SUM(s.count) as total_count
		FROM scancode_stats s
		LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
		WHERE m.key_category = 'letter'
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

	rows, err := tba.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query letter frequency: %w", err)
	}
	defer rows.Close()

	// First pass: collect data and calculate total
	var frequencies []LetterFrequency
	var totalLetters int64

	for rows.Next() {
		var freq LetterFrequency
		if err := rows.Scan(&freq.Letter, &freq.Count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		frequencies = append(frequencies, freq)
		totalLetters += freq.Count
	}

	// Second pass: calculate percentages
	for i := range frequencies {
		if totalLetters > 0 {
			frequencies[i].Percentage = float64(frequencies[i].Count) / float64(totalLetters) * 100
		}
	}

	return frequencies, nil
}
