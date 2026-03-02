package screentime

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jengzang/records-backend-go/internal/logger"
	"github.com/sirupsen/logrus"
)

// HeatmapCell represents a single cell in the heatmap
type HeatmapCell struct {
	Date     string  `json:"date"`     // YYYYMMDD format
	Hour     int     `json:"hour"`     // 0-23
	Value    int64   `json:"value"`    // Duration in milliseconds
	Intensity float64 `json:"intensity"` // Normalized 0-1
}

// UsageHeatmap represents the complete heatmap data
type UsageHeatmap struct {
	Cells       []HeatmapCell `json:"cells"`
	MaxValue    int64         `json:"maxValue"`
	MinValue    int64         `json:"minValue"`
	AvgValue    float64       `json:"avgValue"`
	DateRange   DateRange     `json:"dateRange"`
	TotalCells  int           `json:"totalCells"`
	Description string        `json:"description"`
}

// GetUsageHeatmap returns usage heatmap data (date × hour)
// GET /api/v1/screentime/analysis/usage-heatmap?start=20240101&end=20241231
func (h *Handler) GetUsageHeatmap(c *gin.Context) {
	logger.Info("Usage heatmap requested", logrus.Fields{
		"client_ip": c.ClientIP(),
	})

	start := c.DefaultQuery("start", "")
	end := c.DefaultQuery("end", "")

	// If no date range specified, use last 30 days
	if start == "" || end == "" {
		now := time.Now()
		end = now.Format("20060102")
		start = now.AddDate(0, 0, -30).Format("20060102")
	}

	heatmap := UsageHeatmap{
		Cells: []HeatmapCell{},
		DateRange: DateRange{
			Start: start,
			End:   end,
		},
		Description: "App usage intensity heatmap (date × hour)",
	}

	// Query to get date × hour aggregated data
	query := `
	SELECT
		s.date,
		CAST(substr(s.start_time, 1, 2) AS INTEGER) as hour,
		SUM(s.duration_ms) as total_duration
	FROM screentime_sessions s
	WHERE s.date >= ? AND s.date <= ?
	GROUP BY s.date, hour
	ORDER BY s.date, hour
	`

	rows, err := h.db.Query(query, start, end)
	if err != nil {
		logger.Error("Failed to query heatmap data", err, logrus.Fields{
			"start": start,
			"end":   end,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var maxValue int64 = 0
	var minValue int64 = -1
	var totalValue int64 = 0
	var cellCount int = 0

	for rows.Next() {
		var cell HeatmapCell
		var hour sql.NullInt64

		err := rows.Scan(&cell.Date, &hour, &cell.Value)
		if err != nil {
			logger.Warn("Failed to scan heatmap cell", logrus.Fields{
				"error": err.Error(),
			})
			continue
		}

		// Handle NULL hour (invalid time format)
		if !hour.Valid {
			continue
		}
		cell.Hour = int(hour.Int64)

		// Skip invalid hours
		if cell.Hour < 0 || cell.Hour >= 24 {
			continue
		}

		// Update statistics
		if cell.Value > maxValue {
			maxValue = cell.Value
		}
		if minValue == -1 || cell.Value < minValue {
			minValue = cell.Value
		}
		totalValue += cell.Value
		cellCount++

		heatmap.Cells = append(heatmap.Cells, cell)
	}

	// Calculate average
	if cellCount > 0 {
		heatmap.AvgValue = float64(totalValue) / float64(cellCount)
	}

	heatmap.MaxValue = maxValue
	heatmap.MinValue = minValue
	heatmap.TotalCells = cellCount

	// Normalize intensity values (0-1 scale)
	if maxValue > 0 {
		for i := range heatmap.Cells {
			heatmap.Cells[i].Intensity = float64(heatmap.Cells[i].Value) / float64(maxValue)
		}
	}

	logger.Info("Usage heatmap generated", logrus.Fields{
		"total_cells": cellCount,
		"max_value":   maxValue,
		"min_value":   minValue,
		"avg_value":   heatmap.AvgValue,
	})

	c.JSON(http.StatusOK, heatmap)
}
