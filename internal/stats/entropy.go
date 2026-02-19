package stats

import (
	"math"
)

// ShannonEntropy calculates the Shannon entropy of a probability distribution
// values: frequency counts or probabilities
// Returns entropy in bits (log base 2)
func ShannonEntropy(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Normalize to probabilities
	sum := Sum(values)
	if sum == 0 {
		return 0
	}

	var entropy float64
	for _, v := range values {
		if v > 0 {
			p := v / sum
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// ShannonEntropyNats calculates Shannon entropy in nats (log base e)
func ShannonEntropyNats(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := Sum(values)
	if sum == 0 {
		return 0
	}

	var entropy float64
	for _, v := range values {
		if v > 0 {
			p := v / sum
			entropy -= p * math.Log(p)
		}
	}

	return entropy
}

// NormalizedEntropy calculates the normalized Shannon entropy (0 to 1)
// Divides by log2(n) where n is the number of categories
func NormalizedEntropy(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	entropy := ShannonEntropy(values)
	maxEntropy := math.Log2(float64(len(values)))

	if maxEntropy == 0 {
		return 0
	}

	return entropy / maxEntropy
}

// GiniImpurity calculates the Gini impurity (used in decision trees)
// Returns value between 0 (pure) and 0.5 (maximum impurity for binary)
func GiniImpurity(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := Sum(values)
	if sum == 0 {
		return 0
	}

	var gini float64
	for _, v := range values {
		if v > 0 {
			p := v / sum
			gini += p * (1 - p)
		}
	}

	return gini
}

// CrossEntropy calculates the cross-entropy between two distributions
// p: true distribution, q: predicted distribution
func CrossEntropy(p, q []float64) float64 {
	if len(p) != len(q) || len(p) == 0 {
		return 0
	}

	// Normalize distributions
	sumP := Sum(p)
	sumQ := Sum(q)
	if sumP == 0 || sumQ == 0 {
		return 0
	}

	var crossEntropy float64
	for i := 0; i < len(p); i++ {
		if p[i] > 0 && q[i] > 0 {
			pi := p[i] / sumP
			qi := q[i] / sumQ
			crossEntropy -= pi * math.Log2(qi)
		} else if p[i] > 0 && q[i] == 0 {
			// Infinite cross-entropy if q[i] = 0 but p[i] > 0
			return math.Inf(1)
		}
	}

	return crossEntropy
}

// KLDivergence calculates the Kullback-Leibler divergence from q to p
// D_KL(P || Q) = sum(p[i] * log(p[i] / q[i]))
func KLDivergence(p, q []float64) float64 {
	if len(p) != len(q) || len(p) == 0 {
		return 0
	}

	// Normalize distributions
	sumP := Sum(p)
	sumQ := Sum(q)
	if sumP == 0 || sumQ == 0 {
		return 0
	}

	var kl float64
	for i := 0; i < len(p); i++ {
		if p[i] > 0 && q[i] > 0 {
			pi := p[i] / sumP
			qi := q[i] / sumQ
			kl += pi * math.Log2(pi/qi)
		} else if p[i] > 0 && q[i] == 0 {
			// Infinite KL divergence if q[i] = 0 but p[i] > 0
			return math.Inf(1)
		}
	}

	return kl
}

// JSDivergence calculates the Jensen-Shannon divergence (symmetric version of KL)
// JSD(P || Q) = 0.5 * KL(P || M) + 0.5 * KL(Q || M), where M = 0.5 * (P + Q)
func JSDivergence(p, q []float64) float64 {
	if len(p) != len(q) || len(p) == 0 {
		return 0
	}

	// Calculate mixture distribution M
	m := make([]float64, len(p))
	for i := 0; i < len(p); i++ {
		m[i] = (p[i] + q[i]) / 2
	}

	// Calculate KL divergences
	klPM := KLDivergence(p, m)
	klQM := KLDivergence(q, m)

	// Check for infinite values
	if math.IsInf(klPM, 1) || math.IsInf(klQM, 1) {
		return math.Inf(1)
	}

	return (klPM + klQM) / 2
}

// MutualInformation calculates the mutual information between two variables
// x, y: discrete values (will be binned if continuous)
func MutualInformation(x, y []float64, bins int) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	// Create joint frequency table
	jointFreq := make(map[[2]int]int)
	xFreq := make(map[int]int)
	yFreq := make(map[int]int)

	// Bin the data
	xBinned := binData(x, bins)
	yBinned := binData(y, bins)

	n := len(x)
	for i := 0; i < n; i++ {
		xBin := xBinned[i]
		yBin := yBinned[i]
		jointFreq[[2]int{xBin, yBin}]++
		xFreq[xBin]++
		yFreq[yBin]++
	}

	// Calculate mutual information
	var mi float64
	for joint, count := range jointFreq {
		pXY := float64(count) / float64(n)
		pX := float64(xFreq[joint[0]]) / float64(n)
		pY := float64(yFreq[joint[1]]) / float64(n)

		if pXY > 0 && pX > 0 && pY > 0 {
			mi += pXY * math.Log2(pXY/(pX*pY))
		}
	}

	return mi
}

// binData bins continuous data into discrete bins
func binData(data []float64, bins int) []int {
	if len(data) == 0 || bins <= 0 {
		return nil
	}

	min := Min(data)
	max := Max(data)
	rangeVal := max - min

	if rangeVal == 0 {
		result := make([]int, len(data))
		return result
	}

	binned := make([]int, len(data))
	for i, v := range data {
		bin := int((v - min) / rangeVal * float64(bins))
		if bin >= bins {
			bin = bins - 1
		}
		binned[i] = bin
	}

	return binned
}

// ConditionalEntropy calculates H(Y|X) - entropy of Y given X
func ConditionalEntropy(x, y []float64, bins int) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	// H(Y|X) = H(X,Y) - H(X)
	jointEntropy := JointEntropy(x, y, bins)
	xEntropy := EntropyFromData(x, bins)

	return jointEntropy - xEntropy
}

// JointEntropy calculates the joint entropy H(X,Y)
func JointEntropy(x, y []float64, bins int) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	// Create joint frequency table
	jointFreq := make(map[[2]int]int)

	xBinned := binData(x, bins)
	yBinned := binData(y, bins)

	n := len(x)
	for i := 0; i < n; i++ {
		jointFreq[[2]int{xBinned[i], yBinned[i]}]++
	}

	// Calculate joint entropy
	var entropy float64
	for _, count := range jointFreq {
		if count > 0 {
			p := float64(count) / float64(n)
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// EntropyFromData calculates entropy from continuous data by binning
func EntropyFromData(data []float64, bins int) float64 {
	if len(data) == 0 || bins <= 0 {
		return 0
	}

	binned := binData(data, bins)

	// Count frequencies
	freq := make(map[int]int)
	for _, bin := range binned {
		freq[bin]++
	}

	// Convert to slice
	freqSlice := make([]float64, 0, len(freq))
	for _, count := range freq {
		freqSlice = append(freqSlice, float64(count))
	}

	return ShannonEntropy(freqSlice)
}
