package stats

import (
	"math"
	"sort"
)

// Mean calculates the arithmetic mean of a slice of float64 values
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// WeightedMean calculates the weighted mean
func WeightedMean(values, weights []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sumWeighted, sumWeights float64
	for i, v := range values {
		w := 1.0
		if i < len(weights) {
			w = weights[i]
		}
		sumWeighted += v * w
		sumWeights += w
	}

	if sumWeights == 0 {
		return Mean(values)
	}

	return sumWeighted / sumWeights
}

// Variance calculates the sample variance
func Variance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := Mean(values)
	var sumSquaredDiff float64
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(values)-1)
}

// WeightedVariance calculates the weighted variance
func WeightedVariance(values, weights []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := WeightedMean(values, weights)
	var sumWeightedSquaredDiff, sumWeights float64

	for i, v := range values {
		w := 1.0
		if i < len(weights) {
			w = weights[i]
		}
		diff := v - mean
		sumWeightedSquaredDiff += w * diff * diff
		sumWeights += w
	}

	if sumWeights == 0 {
		return Variance(values)
	}

	return sumWeightedSquaredDiff / sumWeights
}

// StdDev calculates the sample standard deviation
func StdDev(values []float64) float64 {
	return math.Sqrt(Variance(values))
}

// WeightedStdDev calculates the weighted standard deviation
func WeightedStdDev(values, weights []float64) float64 {
	return math.Sqrt(WeightedVariance(values, weights))
}

// Median calculates the median value
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// Min returns the minimum value
func Min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Max returns the maximum value
func Max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// Sum returns the sum of all values
func Sum(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}

// Range returns the range (max - min)
func Range(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return Max(values) - Min(values)
}

// Mode returns the most frequent value (for discrete data)
// For continuous data, consider binning first
func Mode(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	freq := make(map[float64]int)
	for _, v := range values {
		freq[v]++
	}

	maxFreq := 0
	var mode float64
	for v, f := range freq {
		if f > maxFreq {
			maxFreq = f
			mode = v
		}
	}

	return mode
}

// CoefficientOfVariation calculates the coefficient of variation (CV = stddev / mean)
func CoefficientOfVariation(values []float64) float64 {
	mean := Mean(values)
	if mean == 0 {
		return 0
	}
	return StdDev(values) / mean
}

// Skewness calculates the sample skewness (Fisher-Pearson coefficient)
func Skewness(values []float64) float64 {
	n := len(values)
	if n < 3 {
		return 0
	}

	mean := Mean(values)
	stddev := StdDev(values)
	if stddev == 0 {
		return 0
	}

	var sumCubedDiff float64
	for _, v := range values {
		diff := (v - mean) / stddev
		sumCubedDiff += diff * diff * diff
	}

	return sumCubedDiff * float64(n) / float64((n-1)*(n-2))
}

// Kurtosis calculates the sample excess kurtosis
func Kurtosis(values []float64) float64 {
	n := len(values)
	if n < 4 {
		return 0
	}

	mean := Mean(values)
	stddev := StdDev(values)
	if stddev == 0 {
		return 0
	}

	var sumQuadDiff float64
	for _, v := range values {
		diff := (v - mean) / stddev
		sumQuadDiff += diff * diff * diff * diff
	}

	kurtosis := sumQuadDiff * float64(n*(n+1)) / float64((n-1)*(n-2)*(n-3))
	kurtosis -= 3.0 * float64((n-1)*(n-1)) / float64((n-2)*(n-3))

	return kurtosis
}

// Quantile calculates the q-th quantile (0 <= q <= 1)
func Quantile(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if q < 0 {
		q = 0
	}
	if q > 1 {
		q = 1
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := float64(len(sorted))
	index := q * (n - 1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// IQR calculates the interquartile range (Q3 - Q1)
func IQR(values []float64) float64 {
	q1 := Quantile(values, 0.25)
	q3 := Quantile(values, 0.75)
	return q3 - q1
}

// MAD calculates the median absolute deviation
func MAD(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	median := Median(values)
	deviations := make([]float64, len(values))
	for i, v := range values {
		deviations[i] = math.Abs(v - median)
	}

	return Median(deviations)
}

// ZScore calculates the z-score for each value
func ZScore(values []float64) []float64 {
	mean := Mean(values)
	stddev := StdDev(values)

	if stddev == 0 {
		result := make([]float64, len(values))
		return result
	}

	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = (v - mean) / stddev
	}

	return result
}

// Normalize normalizes values to [0, 1] range
func Normalize(values []float64) []float64 {
	min := Min(values)
	max := Max(values)
	rangeVal := max - min

	if rangeVal == 0 {
		result := make([]float64, len(values))
		return result
	}

	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = (v - min) / rangeVal
	}

	return result
}

// Standardize standardizes values to mean=0, stddev=1
func Standardize(values []float64) []float64 {
	return ZScore(values)
}
