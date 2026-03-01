package analysis

import (
	"database/sql"
	"fmt"
)

// HandBalanceStats represents hand usage balance statistics
type HandBalanceStats struct {
	LeftHand struct {
		TotalPresses int     `json:"totalPresses"`
		Percentage   float64 `json:"percentage"`
		TopKeys      []KeyStat `json:"topKeys"`
	} `json:"leftHand"`
	RightHand struct {
		TotalPresses int     `json:"totalPresses"`
		Percentage   float64 `json:"percentage"`
		TopKeys      []KeyStat `json:"topKeys"`
	} `json:"rightHand"`
	BothHands struct {
		TotalPresses int     `json:"totalPresses"`
		Percentage   float64 `json:"percentage"`
		Keys         []string `json:"keys"`
	} `json:"bothHands"`
	Neutral struct {
		TotalPresses int     `json:"totalPresses"`
		Percentage   float64 `json:"percentage"`
		Keys         []string `json:"keys"`
	} `json:"neutral"`
	BalanceScore float64  `json:"balanceScore"` // 0-100, 100 = perfect balance
	Insights     []string `json:"insights"`
}

// KeyStat represents statistics for a single key
type KeyStat struct {
	KeyName string `json:"keyName"`
	Count   int    `json:"count"`
}

// GetHandBalanceStats calculates hand usage balance statistics
func GetHandBalanceStats(db *sql.DB, getKeyHand func(int) string, getKeyName func(int) string) (*HandBalanceStats, error) {
	stats := &HandBalanceStats{}
	stats.LeftHand.TopKeys = []KeyStat{}
	stats.RightHand.TopKeys = []KeyStat{}
	stats.BothHands.Keys = []string{}
	stats.Neutral.Keys = []string{}
	stats.Insights = []string{}

	// Query all keyboard data
	query := `
	SELECT scancode, SUM(count) as total_count
	FROM keyboard_data
	GROUP BY scancode
	ORDER BY total_count DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query keyboard data: %w", err)
	}
	defer rows.Close()

	var totalPresses int
	leftHandPresses := 0
	rightHandPresses := 0
	bothHandsPresses := 0
	neutralPresses := 0

	leftHandKeys := make(map[string]int)
	rightHandKeys := make(map[string]int)
	bothHandsKeys := make(map[string]int)
	neutralKeys := make(map[string]int)

	for rows.Next() {
		var scancode, count int
		if err := rows.Scan(&scancode, &count); err != nil {
			continue
		}

		totalPresses += count
		hand := getKeyHand(scancode)
		keyName := getKeyName(scancode)

		switch hand {
		case "left":
			leftHandPresses += count
			leftHandKeys[keyName] += count
		case "right":
			rightHandPresses += count
			rightHandKeys[keyName] += count
		case "both":
			bothHandsPresses += count
			bothHandsKeys[keyName] += count
		case "neutral":
			neutralPresses += count
			neutralKeys[keyName] += count
		}
	}

	// Calculate percentages
	if totalPresses > 0 {
		stats.LeftHand.TotalPresses = leftHandPresses
		stats.LeftHand.Percentage = float64(leftHandPresses) / float64(totalPresses) * 100

		stats.RightHand.TotalPresses = rightHandPresses
		stats.RightHand.Percentage = float64(rightHandPresses) / float64(totalPresses) * 100

		stats.BothHands.TotalPresses = bothHandsPresses
		stats.BothHands.Percentage = float64(bothHandsPresses) / float64(totalPresses) * 100

		stats.Neutral.TotalPresses = neutralPresses
		stats.Neutral.Percentage = float64(neutralPresses) / float64(totalPresses) * 100
	}

	// Get top keys for each hand
	stats.LeftHand.TopKeys = getTopKeys(leftHandKeys, 10)
	stats.RightHand.TopKeys = getTopKeys(rightHandKeys, 10)

	// Get keys for both hands and neutral
	for key := range bothHandsKeys {
		stats.BothHands.Keys = append(stats.BothHands.Keys, key)
	}
	for key := range neutralKeys {
		stats.Neutral.Keys = append(stats.Neutral.Keys, key)
	}

	// Calculate balance score (100 = perfect 50/50 balance)
	if leftHandPresses+rightHandPresses > 0 {
		leftRatio := float64(leftHandPresses) / float64(leftHandPresses+rightHandPresses)
		// Perfect balance is 0.5, so score = 100 - (deviation from 0.5) * 200
		deviation := leftRatio - 0.5
		if deviation < 0 {
			deviation = -deviation
		}
		stats.BalanceScore = 100 - (deviation * 200)
		if stats.BalanceScore < 0 {
			stats.BalanceScore = 0
		}
	}

	// Generate insights
	stats.Insights = generateHandBalanceInsights(stats)

	return stats, nil
}

// getTopKeys returns the top N keys by count
func getTopKeys(keyMap map[string]int, n int) []KeyStat {
	// Convert map to slice
	keys := make([]KeyStat, 0, len(keyMap))
	for key, count := range keyMap {
		keys = append(keys, KeyStat{KeyName: key, Count: count})
	}

	// Sort by count (bubble sort for simplicity)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j].Count > keys[i].Count {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Return top N
	if len(keys) > n {
		return keys[:n]
	}
	return keys
}

// generateHandBalanceInsights generates insights based on hand balance statistics
func generateHandBalanceInsights(stats *HandBalanceStats) []string {
	insights := []string{}

	// Balance score insights
	if stats.BalanceScore >= 90 {
		insights = append(insights, "左右手使用非常平衡，这是理想的打字习惯")
	} else if stats.BalanceScore >= 70 {
		insights = append(insights, "左右手使用较为平衡，保持良好的打字习惯")
	} else if stats.BalanceScore >= 50 {
		insights = append(insights, "左右手使用存在一定不平衡，可能导致单手疲劳")
	} else {
		insights = append(insights, "左右手使用严重不平衡，建议调整打字习惯以减少单手负担")
	}

	// Dominant hand insights
	if stats.LeftHand.Percentage > stats.RightHand.Percentage+10 {
		insights = append(insights, fmt.Sprintf("左手使用频率较高(%.1f%%)，注意左手休息", stats.LeftHand.Percentage))
	} else if stats.RightHand.Percentage > stats.LeftHand.Percentage+10 {
		insights = append(insights, fmt.Sprintf("右手使用频率较高(%.1f%%)，注意右手休息", stats.RightHand.Percentage))
	}

	// Top key insights
	if len(stats.LeftHand.TopKeys) > 0 {
		topKey := stats.LeftHand.TopKeys[0]
		insights = append(insights, fmt.Sprintf("左手最常用按键: %s (%d次)", topKey.KeyName, topKey.Count))
	}
	if len(stats.RightHand.TopKeys) > 0 {
		topKey := stats.RightHand.TopKeys[0]
		insights = append(insights, fmt.Sprintf("右手最常用按键: %s (%d次)", topKey.KeyName, topKey.Count))
	}

	return insights
}
