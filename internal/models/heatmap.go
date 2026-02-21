package models

// HeatmapPoint represents a single point in the heatmap
type HeatmapPoint struct {
	Lat       float64 `json:"lat"`       // Latitude
	Lng       float64 `json:"lng"`       // Longitude
	Intensity float64 `json:"intensity"` // Normalized 0-1
	Value     int     `json:"value"`     // Raw value
	Metric    string  `json:"metric"`    // "point_count", "duration", "visit_count"
}

// HeatmapResponse represents the heatmap API response
type HeatmapResponse struct {
	Points    []HeatmapPoint `json:"points"`
	Count     int            `json:"count"`
	MaxValue  int            `json:"max_value"`
	MinValue  int            `json:"min_value"`
	Metric    string         `json:"metric"`
	GridLevel int            `json:"grid_level"`
}
