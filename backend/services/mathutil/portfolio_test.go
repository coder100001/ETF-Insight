package mathutil

import (
	"math"
	"testing"
)

func TestPortfolioReturn(t *testing.T) {
	weights := []float64{0.5, 0.3, 0.2}
	returns := []float64{0.12, 0.10, 0.08}
	result := PortfolioReturn(weights, returns)
	expected := 0.5*0.12 + 0.3*0.10 + 0.2*0.08
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("PortfolioReturn = %f, expected %f", result, expected)
	}
}

func TestPortfolioVolatility(t *testing.T) {
	weights := []float64{0.5, 0.5}
	covMatrix := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	result := PortfolioVolatility(weights, covMatrix)
	if result <= 0 {
		t.Errorf("PortfolioVolatility should be positive, got %f", result)
	}
	variance := 0.5*0.5*0.04 + 2*0.5*0.5*0.02 + 0.5*0.5*0.03
	expected := math.Sqrt(variance)
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("PortfolioVolatility = %f, expected %f", result, expected)
	}
}

func TestPortfolioVolatility_ZeroVariance(t *testing.T) {
	weights := []float64{0.5, 0.5}
	covMatrix := [][]float64{
		{0, 0},
		{0, 0},
	}
	result := PortfolioVolatility(weights, covMatrix)
	if result != 0 {
		t.Errorf("PortfolioVolatility should be 0 for zero covariance, got %f", result)
	}
}

func TestPortfolioVolatility_NegativeVariance(t *testing.T) {
	weights := []float64{1.0}
	covMatrix := [][]float64{{-0.01}}
	result := PortfolioVolatility(weights, covMatrix)
	if result != 0 {
		t.Errorf("PortfolioVolatility should be 0 for negative variance, got %f", result)
	}
}

func TestDiversificationRatio(t *testing.T) {
	weights := []float64{0.5, 0.5}
	vols := []float64{0.2, 0.15}
	portfolioVol := 0.15
	result := DiversificationRatio(weights, vols, portfolioVol)
	expected := (0.5*0.2 + 0.5*0.15) / 0.15
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("DiversificationRatio = %f, expected %f", result, expected)
	}
}

func TestDiversificationRatio_ZeroPortfolioVol(t *testing.T) {
	weights := []float64{0.5, 0.5}
	vols := []float64{0.2, 0.15}
	result := DiversificationRatio(weights, vols, 0)
	if result != 1.0 {
		t.Errorf("DiversificationRatio should be 1.0 for zero portfolio vol, got %f", result)
	}
}
