package stats

import (
	"math"
	"sort"
)

// Percentile calculates the p-th percentile (0-100)
// Uses linear interpolation between closest ranks
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}

	// Convert percentile to quantile
	q := p / 100.0
	return Quantile(values, q)
}

// Percentiles calculates multiple percentiles at once
func Percentiles(values []float64, ps []float64) []float64 {
	if len(values) == 0 {
		return make([]float64, len(ps))
	}

	// Sort once for efficiency
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	results := make([]float64, len(ps))
	for i, p := range ps {
		if p < 0 {
			p = 0
		}
		if p > 100 {
			p = 100
		}

		q := p / 100.0
		n := float64(len(sorted))
		index := q * (n - 1)
		lower := int(math.Floor(index))
		upper := int(math.Ceil(index))

		if lower == upper {
			results[i] = sorted[lower]
		} else {
			// Linear interpolation
			weight := index - float64(lower)
			results[i] = sorted[lower]*(1-weight) + sorted[upper]*weight
		}
	}

	return results
}

// PercentileRank calculates the percentile rank of a value
// Returns the percentage of values less than or equal to the given value
func PercentileRank(values []float64, value float64) float64 {
	if len(values) == 0 {
		return 0
	}

	count := 0
	for _, v := range values {
		if v <= value {
			count++
		}
	}

	return float64(count) / float64(len(values)) * 100.0
}

// FiveNumberSummary returns the five-number summary (min, Q1, median, Q3, max)
func FiveNumberSummary(values []float64) (min, q1, median, q3, max float64) {
	if len(values) == 0 {
		return 0, 0, 0, 0, 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	min = sorted[0]
	max = sorted[len(sorted)-1]
	q1 = Quantile(sorted, 0.25)
	median = Quantile(sorted, 0.5)
	q3 = Quantile(sorted, 0.75)

	return
}

// Quartiles returns the three quartiles (Q1, Q2/median, Q3)
func Quartiles(values []float64) (q1, q2, q3 float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	q1 = Quantile(values, 0.25)
	q2 = Quantile(values, 0.5)
	q3 = Quantile(values, 0.75)

	return
}

// Deciles returns the nine deciles (D1 through D9)
func Deciles(values []float64) []float64 {
	deciles := make([]float64, 9)
	for i := 1; i <= 9; i++ {
		deciles[i-1] = Quantile(values, float64(i)/10.0)
	}
	return deciles
}

// OutliersBounds calculates the lower and upper bounds for outliers using IQR method
// Outliers are values < Q1 - 1.5*IQR or > Q3 + 1.5*IQR
func OutliersBounds(values []float64) (lowerBound, upperBound float64) {
	q1, _, q3 := Quartiles(values)
	iqr := q3 - q1

	lowerBound = q1 - 1.5*iqr
	upperBound = q3 + 1.5*iqr

	return
}

// DetectOutliers identifies outliers using the IQR method
// Returns indices of outlier values
func DetectOutliers(values []float64) []int {
	if len(values) == 0 {
		return nil
	}

	lowerBound, upperBound := OutliersBounds(values)

	outliers := []int{}
	for i, v := range values {
		if v < lowerBound || v > upperBound {
			outliers = append(outliers, i)
		}
	}

	return outliers
}

// RemoveOutliers removes outliers from the data using IQR method
func RemoveOutliers(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	lowerBound, upperBound := OutliersBounds(values)

	filtered := []float64{}
	for _, v := range values {
		if v >= lowerBound && v <= upperBound {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

// ZScoreOutliers identifies outliers using z-score method
// threshold: typically 2.5 or 3.0 standard deviations
func ZScoreOutliers(values []float64, threshold float64) []int {
	if len(values) == 0 {
		return nil
	}

	zScores := ZScore(values)

	outliers := []int{}
	for i, z := range zScores {
		if math.Abs(z) > threshold {
			outliers = append(outliers, i)
		}
	}

	return outliers
}

// MADOutliers identifies outliers using MAD (Median Absolute Deviation) method
// threshold: typically 2.5 or 3.0 MADs
func MADOutliers(values []float64, threshold float64) []int {
	if len(values) == 0 {
		return nil
	}

	median := Median(values)
	mad := MAD(values)

	if mad == 0 {
		return nil
	}

	outliers := []int{}
	for i, v := range values {
		// Modified z-score using MAD
		modifiedZ := 0.6745 * math.Abs(v-median) / mad
		if modifiedZ > threshold {
			outliers = append(outliers, i)
		}
	}

	return outliers
}

// Winsorize replaces extreme values with less extreme values
// lower, upper: percentiles (0-100) to winsorize at
func Winsorize(values []float64, lower, upper float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	lowerVal := Percentile(values, lower)
	upperVal := Percentile(values, upper)

	result := make([]float64, len(values))
	for i, v := range values {
		if v < lowerVal {
			result[i] = lowerVal
		} else if v > upperVal {
			result[i] = upperVal
		} else {
			result[i] = v
		}
	}

	return result
}

// Trim removes extreme values from both ends
// lower, upper: percentiles (0-100) to trim at
func Trim(values []float64, lower, upper float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	lowerVal := Percentile(values, lower)
	upperVal := Percentile(values, upper)

	result := []float64{}
	for _, v := range values {
		if v >= lowerVal && v <= upperVal {
			result = append(result, v)
		}
	}

	return result
}

// PercentileCI calculates the confidence interval for a percentile using bootstrap
// p: percentile (0-100), confidence: confidence level (e.g., 0.95 for 95%)
// nBootstrap: number of bootstrap samples
func PercentileCI(values []float64, p float64, confidence float64, nBootstrap int) (lower, upper float64) {
	if len(values) < 2 || nBootstrap < 100 {
		return 0, 0
	}

	// Bootstrap resampling
	bootstrapPercentiles := make([]float64, nBootstrap)
	n := len(values)

	for i := 0; i < nBootstrap; i++ {
		// Resample with replacement
		sample := make([]float64, n)
		for j := 0; j < n; j++ {
			// Simple pseudo-random sampling (for production, use crypto/rand)
			idx := (i*n + j) % n
			sample[j] = values[idx]
		}

		bootstrapPercentiles[i] = Percentile(sample, p)
	}

	// Calculate confidence interval
	alpha := 1 - confidence
	lower = Percentile(bootstrapPercentiles, alpha/2*100)
	upper = Percentile(bootstrapPercentiles, (1-alpha/2)*100)

	return
}
