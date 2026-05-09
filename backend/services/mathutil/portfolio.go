package mathutil

import "math"

func PortfolioReturn(weights, returns []float64) float64 {
	var sum float64
	for i := range weights {
		sum += weights[i] * returns[i]
	}
	return sum
}

func PortfolioVolatility(weights []float64, covMatrix [][]float64) float64 {
	var variance float64
	n := len(weights)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			variance += weights[i] * weights[j] * covMatrix[i][j]
		}
	}
	if variance < 0 {
		return 0
	}
	return math.Sqrt(variance)
}

func DiversificationRatio(weights []float64, vols []float64, portfolioVol float64) float64 {
	if portfolioVol == 0 {
		return 1.0
	}
	var weightedSum float64
	for i := range weights {
		weightedSum += weights[i] * vols[i]
	}
	return weightedSum / portfolioVol
}
