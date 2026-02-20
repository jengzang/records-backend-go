package stats

import (
	"math"
	"sort"
)

// PearsonCorrelation calculates the Pearson correlation coefficient between two variables
// Returns value between -1 and 1
func PearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	_ = float64(len(x)) // n is calculated but not used
	meanX := Mean(x)
	meanY := Mean(y)

	var sumXY, sumX2, sumY2 float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}

	if sumX2 == 0 || sumY2 == 0 {
		return 0
	}

	return sumXY / math.Sqrt(sumX2*sumY2)
}

// SpearmanCorrelation calculates the Spearman rank correlation coefficient
// Returns value between -1 and 1
func SpearmanCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	// Convert to ranks
	rankX := rank(x)
	rankY := rank(y)

	// Calculate Pearson correlation on ranks
	return PearsonCorrelation(rankX, rankY)
}

// rank converts values to ranks (average rank for ties)
func rank(values []float64) []float64 {
	n := len(values)
	if n == 0 {
		return nil
	}

	// Create index-value pairs
	type pair struct {
		index int
		value float64
	}
	pairs := make([]pair, n)
	for i, v := range values {
		pairs[i] = pair{i, v}
	}

	// Sort by value
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value < pairs[j].value
	})

	// Assign ranks (handle ties with average rank)
	ranks := make([]float64, n)
	i := 0
	for i < n {
		j := i
		// Find all values equal to current value
		for j < n && pairs[j].value == pairs[i].value {
			j++
		}

		// Average rank for ties
		avgRank := float64(i+j+1) / 2.0

		// Assign average rank to all tied values
		for k := i; k < j; k++ {
			ranks[pairs[k].index] = avgRank
		}

		i = j
	}

	return ranks
}

// Covariance calculates the sample covariance between two variables
func Covariance(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))
	meanX := Mean(x)
	meanY := Mean(y)

	var sumXY float64
	for i := 0; i < len(x); i++ {
		sumXY += (x[i] - meanX) * (y[i] - meanY)
	}

	return sumXY / (n - 1)
}

// RSquared calculates the coefficient of determination (R²)
// Measures the proportion of variance in y explained by x
func RSquared(x, y []float64) float64 {
	r := PearsonCorrelation(x, y)
	return r * r
}

// LinearRegression performs simple linear regression (y = a + bx)
// Returns slope (b) and intercept (a)
func LinearRegression(x, y []float64) (slope, intercept float64) {
	if len(x) != len(y) || len(x) < 2 {
		return 0, 0
	}

	meanX := Mean(x)
	meanY := Mean(y)

	var sumXY, sumX2 float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		sumXY += dx * (y[i] - meanY)
		sumX2 += dx * dx
	}

	if sumX2 == 0 {
		return 0, meanY
	}

	slope = sumXY / sumX2
	intercept = meanY - slope*meanX

	return slope, intercept
}

// Predict predicts y values using linear regression model
func Predict(x []float64, slope, intercept float64) []float64 {
	predictions := make([]float64, len(x))
	for i, xi := range x {
		predictions[i] = slope*xi + intercept
	}
	return predictions
}

// RMSE calculates the root mean squared error between predicted and actual values
func RMSE(actual, predicted []float64) float64 {
	if len(actual) != len(predicted) || len(actual) == 0 {
		return 0
	}

	var sumSquaredError float64
	for i := 0; i < len(actual); i++ {
		error := actual[i] - predicted[i]
		sumSquaredError += error * error
	}

	return math.Sqrt(sumSquaredError / float64(len(actual)))
}

// MAE calculates the mean absolute error between predicted and actual values
func MAE(actual, predicted []float64) float64 {
	if len(actual) != len(predicted) || len(actual) == 0 {
		return 0
	}

	var sumAbsError float64
	for i := 0; i < len(actual); i++ {
		sumAbsError += math.Abs(actual[i] - predicted[i])
	}

	return sumAbsError / float64(len(actual))
}

// CrossCorrelation calculates the cross-correlation at different lags
// Returns correlation coefficients for lags from -maxLag to +maxLag
func CrossCorrelation(x, y []float64, maxLag int) []float64 {
	if len(x) != len(y) || len(x) < 2 {
		return nil
	}

	n := len(x)
	if maxLag > n-1 {
		maxLag = n - 1
	}

	meanX := Mean(x)
	meanY := Mean(y)
	stdX := StdDev(x)
	stdY := StdDev(y)

	if stdX == 0 || stdY == 0 {
		return nil
	}

	result := make([]float64, 2*maxLag+1)

	for lag := -maxLag; lag <= maxLag; lag++ {
		var sum float64
		var count int

		for i := 0; i < n; i++ {
			j := i + lag
			if j >= 0 && j < n {
				sum += (x[i] - meanX) * (y[j] - meanY)
				count++
			}
		}

		if count > 0 {
			result[lag+maxLag] = sum / (float64(count) * stdX * stdY)
		}
	}

	return result
}

// AutoCorrelation calculates the autocorrelation at different lags
func AutoCorrelation(x []float64, maxLag int) []float64 {
	return CrossCorrelation(x, x, maxLag)
}

// PartialCorrelation calculates the partial correlation between x and y controlling for z
func PartialCorrelation(x, y, z []float64) float64 {
	if len(x) != len(y) || len(x) != len(z) || len(x) < 3 {
		return 0
	}

	rXY := PearsonCorrelation(x, y)
	rXZ := PearsonCorrelation(x, z)
	rYZ := PearsonCorrelation(y, z)

	numerator := rXY - rXZ*rYZ
	denominator := math.Sqrt((1 - rXZ*rXZ) * (1 - rYZ*rYZ))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
