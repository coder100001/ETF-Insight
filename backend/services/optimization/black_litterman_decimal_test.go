package optimization

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBlackLitterman_CalculatePosteriorReturns_DecimalPrecision(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	priorReturns := map[string]float64{
		"AAPL":  0.08,
		"GOOGL": 0.10,
		"MSFT":  0.12,
	}

	Sigma := [][]float64{
		{0.04, 0.02, 0.01},
		{0.02, 0.03, 0.015},
		{0.01, 0.015, 0.02},
	}

	P := [][]float64{
		{1, 0, 0}, // AAPL 绝对观点
	}
	Q := []float64{0.15} // 预期收益 15%

	Omega := [][]float64{
		{0.01},
	}

	result := optimizer.calculatePosteriorReturns(Sigma, priorReturns, P, Q, Omega, symbols)

	if result == nil {
		t.Fatal("calculatePosteriorReturns() returned nil")
	}

	aaplReturn := result["AAPL"]

	expectedMin := decimal.NewFromFloat(0.115)
	expectedMax := decimal.NewFromFloat(0.15)

	actualDecimal := decimal.NewFromFloat(aaplReturn)

	if actualDecimal.LessThan(expectedMin) || actualDecimal.GreaterThan(expectedMax) {
		t.Errorf("AAPL posterior return %.6f should be between %s and %s (blended toward view)",
			aaplReturn, expectedMin.String(), expectedMax.String())
	}
}

func TestBlackLitterman_CalculatePosteriorReturns_EmptyViews(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	symbols := []string{"AAPL", "GOOGL"}
	priorReturns := map[string]float64{
		"AAPL":  0.08,
		"GOOGL": 0.10,
	}

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}

	result := optimizer.calculatePosteriorReturns(Sigma, priorReturns, nil, nil, nil, symbols)

	if result == nil {
		t.Fatal("calculatePosteriorReturns() returned nil for empty views")
	}

	if result["AAPL"] != priorReturns["AAPL"] {
		t.Errorf("With empty views, should return prior returns, got %f", result["AAPL"])
	}

	if result["GOOGL"] != priorReturns["GOOGL"] {
		t.Errorf("With empty views, should return prior returns, got %f", result["GOOGL"])
	}
}

func TestBlackLitterman_CalculatePosteriorReturns_MultipleViews(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	priorReturns := map[string]float64{
		"AAPL":  0.08,
		"GOOGL": 0.10,
		"MSFT":  0.12,
	}

	Sigma := [][]float64{
		{0.04, 0.02, 0.01},
		{0.02, 0.03, 0.015},
		{0.01, 0.015, 0.02},
	}

	P := [][]float64{
		{1, 0, 0}, // AAPL 观点
		{0, 1, 0}, // GOOGL 观点
	}
	Q := []float64{0.15, 0.14} // AAPL 15%, GOOGL 14%

	Omega := [][]float64{
		{0.01, 0},
		{0, 0.01},
	}

	result := optimizer.calculatePosteriorReturns(Sigma, priorReturns, P, Q, Omega, symbols)

	if result == nil {
		t.Fatal("calculatePosteriorReturns() returned nil")
	}

	for _, symbol := range symbols {
		if _, ok := result[symbol]; !ok {
			t.Errorf("Result missing symbol: %s", symbol)
		}
	}
}
