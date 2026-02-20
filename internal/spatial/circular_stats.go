package spatial

import (
	"math"
)

// CircularMean calculates the mean of circular data (angles in radians)
// weights: optional weights for each angle (can be nil for equal weights)
// Returns mean angle in radians
func CircularMean(angles []float64, weights []float64) float64 {
	if len(angles) == 0 {
		return 0
	}

	var sumSin, sumCos float64
	if weights == nil {
		// Equal weights
		for _, angle := range angles {
			sumSin += math.Sin(angle)
			sumCos += math.Cos(angle)
		}
	} else {
		// Weighted
		for i, angle := range angles {
			w := 1.0
			if i < len(weights) {
				w = weights[i]
			}
			sumSin += w * math.Sin(angle)
			sumCos += w * math.Cos(angle)
		}
	}

	return math.Atan2(sumSin, sumCos)
}

// CircularMeanDegrees calculates the mean of circular data in degrees
func CircularMeanDegrees(angles []float64, weights []float64) float64 {
	radians := make([]float64, len(angles))
	for i, angle := range angles {
		radians[i] = angle * math.Pi / 180
	}
	meanRad := CircularMean(radians, weights)
	meanDeg := meanRad * 180 / math.Pi
	if meanDeg < 0 {
		meanDeg += 360
	}
	return meanDeg
}

// CircularVariance calculates the circular variance (1 - R)
// where R is the mean resultant length
func CircularVariance(angles []float64, weights []float64) float64 {
	r := MeanResultantLength(angles, weights)
	return 1 - r
}

// CircularStdDev calculates the circular standard deviation
func CircularStdDev(angles []float64, weights []float64) float64 {
	r := MeanResultantLength(angles, weights)
	return math.Sqrt(-2 * math.Log(r))
}

// MeanResultantLength calculates the mean resultant length (R)
// R ranges from 0 (uniform distribution) to 1 (all angles identical)
func MeanResultantLength(angles []float64, weights []float64) float64 {
	if len(angles) == 0 {
		return 0
	}

	var sumSin, sumCos, sumWeights float64
	if weights == nil {
		// Equal weights
		for _, angle := range angles {
			sumSin += math.Sin(angle)
			sumCos += math.Cos(angle)
		}
		sumWeights = float64(len(angles))
	} else {
		// Weighted
		for i, angle := range angles {
			w := 1.0
			if i < len(weights) {
				w = weights[i]
			}
			sumSin += w * math.Sin(angle)
			sumCos += w * math.Cos(angle)
			sumWeights += w
		}
	}

	if sumWeights == 0 {
		return 0
	}

	r := math.Sqrt(sumSin*sumSin+sumCos*sumCos) / sumWeights
	return r
}

// CircularConcentration calculates the concentration parameter (kappa)
// Higher kappa means more concentrated around the mean
func CircularConcentration(angles []float64, weights []float64) float64 {
	r := MeanResultantLength(angles, weights)
	_ = float64(len(angles)) // n is not used in the approximation

	// Approximation for kappa (von Mises distribution)
	if r < 0.53 {
		return 2*r + r*r*r + 5*r*r*r*r*r/6
	} else if r < 0.85 {
		return -0.4 + 1.39*r + 0.43/(1-r)
	} else {
		return 1 / (r*r*r - 4*r*r + 3*r)
	}
}

// CircularEntropy calculates the Shannon entropy of a circular distribution
// weights: frequency counts for each direction
func CircularEntropy(weights []float64) float64 {
	if len(weights) == 0 {
		return 0
	}

	// Normalize weights to probabilities
	var sum float64
	for _, w := range weights {
		sum += w
	}

	if sum == 0 {
		return 0
	}

	var entropy float64
	for _, w := range weights {
		if w > 0 {
			p := w / sum
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// AngularDifference calculates the smallest difference between two angles (radians)
// Result is in range [-π, π]
func AngularDifference(angle1, angle2 float64) float64 {
	diff := angle2 - angle1
	// Normalize to [-π, π]
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}
	return diff
}

// AngularDifferenceDegrees calculates the smallest difference between two angles (degrees)
// Result is in range [-180, 180]
func AngularDifferenceDegrees(angle1, angle2 float64) float64 {
	diff := angle2 - angle1
	// Normalize to [-180, 180]
	for diff > 180 {
		diff -= 360
	}
	for diff < -180 {
		diff += 360
	}
	return diff
}

// IsCircularUniform tests if angles are uniformly distributed
// Returns true if Rayleigh test p-value > 0.05
func IsCircularUniform(angles []float64) bool {
	r := MeanResultantLength(angles, nil)
	n := float64(len(angles))

	// Rayleigh test statistic
	z := n * r * r

	// Approximate p-value (for large n)
	pValue := math.Exp(-z)

	return pValue > 0.05
}
