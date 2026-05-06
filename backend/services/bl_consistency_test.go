package services

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
)

func TestBLImplementations_ConsistentEquilibriumReturns(t *testing.T) {
	symbols := []string{"AAPL", "MSFT", "GOOG", "AMZN"}

	marketWeights := map[string]float64{
		"AAPL": 0.30, "MSFT": 0.25, "GOOG": 0.25, "AMZN": 0.20,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL": {"AAPL": 0.04, "MSFT": 0.02, "GOOG": 0.015, "AMZN": 0.025},
		"MSFT": {"AAPL": 0.02, "MSFT": 0.03, "GOOG": 0.01, "AMZN": 0.02},
		"GOOG": {"AAPL": 0.015, "MSFT": 0.01, "GOOG": 0.025, "AMZN": 0.015},
		"AMZN": {"AAPL": 0.025, "MSFT": 0.02, "GOOG": 0.015, "AMZN": 0.05},
	}

	optimizer := NewBlackLittermanOptimizer()
	optimizer.SetTau(decimal.NewFromFloat(0.025))
	optimizer.SetRiskFreeRate(decimal.NewFromFloat(0.04))

	result, err := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix)
	if err != nil {
		t.Fatalf("Optimization BL failed: %v", err)
	}

	for _, sym := range symbols {
		impliedReturn, exists := result[sym]
		if !exists {
			t.Errorf("Missing implied return for %s in optimizer result", sym)
			continue
		}
		if impliedReturn <= 0 || math.IsNaN(impliedReturn) || math.IsInf(impliedReturn, 0) {
			t.Errorf("Invalid implied return for %s: %f", sym, impliedReturn)
		}
	}
	t.Logf("Optimizer equilibrium returns: %+v", result)
}

func TestBLImplementations_SingleAsset(t *testing.T) {
	marketWeights := map[string]float64{"AAPL": 1.0}
	covMatrix := map[string]map[string]float64{"AAPL": {"AAPL": 0.04}}

	optimizer := NewBlackLittermanOptimizer()
	optimizer.SetTau(decimal.NewFromFloat(0.025))
	optimizer.SetRiskFreeRate(decimal.NewFromFloat(0.04))

	result, err := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix)
	if err != nil {
		t.Fatalf("Single asset should not fail: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}
}

func TestBLImplementations_DegenerateCovariance(t *testing.T) {
	marketWeights := map[string]float64{"A": 0.5, "B": 0.5}
	covMatrix := map[string]map[string]float64{
		"A": {"A": 0, "B": 0},
		"B": {"A": 0, "B": 0},
	}

	optimizer := NewBlackLittermanOptimizer()
	optimizer.SetTau(decimal.NewFromFloat(0.025))
	optimizer.SetRiskFreeRate(decimal.NewFromFloat(0.04))

	_, err := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix)
	if err == nil {
		t.Error("Degenerate covariance matrix should return error")
	}
}

func TestBLImplementations_NonSumToOneWeights(t *testing.T) {
	marketWeights := map[string]float64{"A": 0.3, "B": 0.3}
	covMatrix := map[string]map[string]float64{
		"A": {"A": 0.04, "B": 0.02},
		"B": {"A": 0.02, "B": 0.03},
	}

	optimizer := NewBlackLittermanOptimizer()
	optimizer.SetTau(decimal.NewFromFloat(0.025))
	optimizer.SetRiskFreeRate(decimal.NewFromFloat(0.04))

	result, err := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix)
	if err != nil {
		t.Logf("Non-sum-to-one weights returned error (may be expected): %v", err)
	} else {
		t.Logf("Non-sum-to-one weights handled: %+v", result)
	}
}
