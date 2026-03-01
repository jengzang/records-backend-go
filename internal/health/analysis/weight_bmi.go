package analysis

import (
	"database/sql"
	"fmt"
	"math"
)

// WeightBMIAnalysis represents weight and BMI analysis results
type WeightBMIAnalysis struct {
	CurrentWeight   float64           `json:"currentWeight"`   // kg
	CurrentBMI      float64           `json:"currentBMI"`
	CurrentHeight   float64           `json:"currentHeight"`   // cm
	BMICategory     string            `json:"bmiCategory"`     // underweight, normal, overweight, obese
	HealthStatus    string            `json:"healthStatus"`    // healthy, warning, risk
	WeightTrend     []WeightDataPoint `json:"weightTrend"`
	BMITrend        []BMIDataPoint    `json:"bmiTrend"`
	Statistics      WeightStatistics  `json:"statistics"`
	Prediction      WeightPrediction  `json:"prediction"`
	Recommendations []string          `json:"recommendations"`
}

// WeightDataPoint represents a single weight measurement
type WeightDataPoint struct {
	Date   string  `json:"date"`
	Weight float64 `json:"weight"` // kg
}

// BMIDataPoint represents a single BMI calculation
type BMIDataPoint struct {
	Date string  `json:"date"`
	BMI  float64 `json:"bmi"`
}

// WeightStatistics represents weight statistics
type WeightStatistics struct {
	MinWeight     float64 `json:"minWeight"`
	MaxWeight     float64 `json:"maxWeight"`
	AvgWeight     float64 `json:"avgWeight"`
	WeightChange  float64 `json:"weightChange"`  // kg (current - first)
	ChangePercent float64 `json:"changePercent"` // %
	DataPoints    int     `json:"dataPoints"`
	DateRange     struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"dateRange"`
}

// WeightPrediction represents weight prediction
type WeightPrediction struct {
	PredictedWeight float64 `json:"predictedWeight"` // kg (30 days from now)
	PredictedBMI    float64 `json:"predictedBMI"`
	Trend           string  `json:"trend"`           // increasing, decreasing, stable
	Confidence      string  `json:"confidence"`      // high, medium, low
}

// GetWeightBMIAnalysis analyzes weight and BMI data
func GetWeightBMIAnalysis(db *sql.DB) (*WeightBMIAnalysis, error) {
	analysis := &WeightBMIAnalysis{}
	analysis.WeightTrend = []WeightDataPoint{}
	analysis.BMITrend = []BMIDataPoint{}
	analysis.Recommendations = []string{}

	// Get current height (most recent measurement)
	heightQuery := `
	SELECT value
	FROM health_records
	WHERE type = 'HKQuantityTypeIdentifierHeight'
	ORDER BY start_date DESC
	LIMIT 1
	`
	var heightMeters float64
	if err := db.QueryRow(heightQuery).Scan(&heightMeters); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get height: %w", err)
		}
		// Default height if not found
		heightMeters = 1.75
	}
	analysis.CurrentHeight = heightMeters * 100 // Convert to cm

	// Get weight data
	weightQuery := `
	SELECT
		DATE(start_date) as date,
		value as weight
	FROM health_records
	WHERE type = 'HKQuantityTypeIdentifierBodyMass'
	ORDER BY start_date ASC
	`

	rows, err := db.Query(weightQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query weight data: %w", err)
	}
	defer rows.Close()

	var weights []WeightDataPoint
	var totalWeight float64
	var minWeight, maxWeight float64
	minWeight = math.MaxFloat64
	maxWeight = 0

	for rows.Next() {
		var date string
		var weight float64
		if err := rows.Scan(&date, &weight); err != nil {
			continue
		}

		weights = append(weights, WeightDataPoint{
			Date:   date,
			Weight: weight,
		})

		totalWeight += weight
		if weight < minWeight {
			minWeight = weight
		}
		if weight > maxWeight {
			maxWeight = weight
		}
	}

	if len(weights) == 0 {
		return nil, fmt.Errorf("no weight data found")
	}

	// Set current weight (most recent)
	analysis.CurrentWeight = weights[len(weights)-1].Weight

	// Calculate current BMI
	analysis.CurrentBMI = analysis.CurrentWeight / math.Pow(heightMeters, 2)

	// Determine BMI category and health status
	analysis.BMICategory, analysis.HealthStatus = getBMICategory(analysis.CurrentBMI)

	// Set weight trend (limit to last 90 days or all data if less)
	if len(weights) > 90 {
		analysis.WeightTrend = weights[len(weights)-90:]
	} else {
		analysis.WeightTrend = weights
	}

	// Calculate BMI trend
	for _, w := range analysis.WeightTrend {
		bmi := w.Weight / math.Pow(heightMeters, 2)
		analysis.BMITrend = append(analysis.BMITrend, BMIDataPoint{
			Date: w.Date,
			BMI:  bmi,
		})
	}

	// Calculate statistics
	analysis.Statistics.MinWeight = minWeight
	analysis.Statistics.MaxWeight = maxWeight
	analysis.Statistics.AvgWeight = totalWeight / float64(len(weights))
	analysis.Statistics.DataPoints = len(weights)
	analysis.Statistics.DateRange.Start = weights[0].Date
	analysis.Statistics.DateRange.End = weights[len(weights)-1].Date

	if len(weights) > 1 {
		analysis.Statistics.WeightChange = weights[len(weights)-1].Weight - weights[0].Weight
		analysis.Statistics.ChangePercent = (analysis.Statistics.WeightChange / weights[0].Weight) * 100
	}

	// Calculate prediction
	analysis.Prediction = predictWeight(weights, heightMeters)

	// Generate recommendations
	analysis.Recommendations = generateWeightRecommendations(analysis)

	return analysis, nil
}

// getBMICategory returns BMI category and health status
func getBMICategory(bmi float64) (string, string) {
	if bmi < 18.5 {
		return "underweight", "warning"
	} else if bmi < 24 {
		return "normal", "healthy"
	} else if bmi < 28 {
		return "overweight", "warning"
	} else {
		return "obese", "risk"
	}
}

// predictWeight predicts weight 30 days from now using linear regression
func predictWeight(weights []WeightDataPoint, heightMeters float64) WeightPrediction {
	prediction := WeightPrediction{}

	if len(weights) < 2 {
		prediction.Trend = "stable"
		prediction.Confidence = "low"
		prediction.PredictedWeight = weights[0].Weight
		prediction.PredictedBMI = weights[0].Weight / math.Pow(heightMeters, 2)
		return prediction
	}

	// Use last 30 days for prediction
	recentWeights := weights
	if len(weights) > 30 {
		recentWeights = weights[len(weights)-30:]
	}

	// Simple linear regression
	n := float64(len(recentWeights))
	var sumX, sumY, sumXY, sumX2 float64

	for i, w := range recentWeights {
		x := float64(i)
		y := w.Weight
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (trend)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Predict 30 days from last measurement
	predictedWeight := intercept + slope*float64(len(recentWeights)+30)
	prediction.PredictedWeight = math.Round(predictedWeight*10) / 10
	prediction.PredictedBMI = math.Round((predictedWeight/math.Pow(heightMeters, 2))*10) / 10

	// Determine trend
	if slope > 0.05 {
		prediction.Trend = "increasing"
	} else if slope < -0.05 {
		prediction.Trend = "decreasing"
	} else {
		prediction.Trend = "stable"
	}

	// Determine confidence based on data points
	if len(recentWeights) >= 20 {
		prediction.Confidence = "high"
	} else if len(recentWeights) >= 10 {
		prediction.Confidence = "medium"
	} else {
		prediction.Confidence = "low"
	}

	return prediction
}

// generateWeightRecommendations generates personalized recommendations
func generateWeightRecommendations(analysis *WeightBMIAnalysis) []string {
	recommendations := []string{}

	// BMI-based recommendations
	switch analysis.BMICategory {
	case "underweight":
		recommendations = append(recommendations, "您的BMI偏低，建议增加营养摄入，咨询营养师制定增重计划")
		recommendations = append(recommendations, "适当增加力量训练，帮助增加肌肉质量")
	case "normal":
		recommendations = append(recommendations, "您的BMI在健康范围内，继续保持良好的饮食和运动习惯")
		recommendations = append(recommendations, "建议每周进行150分钟中等强度有氧运动")
	case "overweight":
		recommendations = append(recommendations, "您的BMI偏高，建议适当控制饮食，增加运动量")
		recommendations = append(recommendations, "建议每周减重0.5-1kg，避免过快减重")
	case "obese":
		recommendations = append(recommendations, "您的BMI超标，建议咨询医生制定科学的减重计划")
		recommendations = append(recommendations, "建议结合饮食控制和规律运动，必要时寻求专业指导")
	}

	// Trend-based recommendations
	if analysis.Prediction.Trend == "increasing" && analysis.BMICategory != "underweight" {
		recommendations = append(recommendations, "体重呈上升趋势，建议注意饮食控制和增加运动")
	} else if analysis.Prediction.Trend == "decreasing" && analysis.BMICategory == "normal" {
		recommendations = append(recommendations, "体重呈下降趋势，注意保持营养均衡，避免过度减重")
	}

	// Weight change recommendations
	if math.Abs(analysis.Statistics.WeightChange) > 5 {
		if analysis.Statistics.WeightChange > 0 {
			recommendations = append(recommendations, fmt.Sprintf("体重增加了%.1fkg，建议关注体重变化原因", analysis.Statistics.WeightChange))
		} else {
			recommendations = append(recommendations, fmt.Sprintf("体重减少了%.1fkg，注意保持健康的减重速度", -analysis.Statistics.WeightChange))
		}
	}

	// General recommendations
	recommendations = append(recommendations, "建议定期测量体重，每周1-2次，早晨空腹测量最准确")
	recommendations = append(recommendations, "保持规律的作息和充足的睡眠，有助于体重管理")

	return recommendations
}
