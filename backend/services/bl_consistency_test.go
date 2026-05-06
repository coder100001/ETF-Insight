package services

import (
	"math"
	"testing"

	"etf-insight/services/optimization"
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

	optimizer := optimization.NewBlackLittermanOptimizer()
	optimizer.SetTau(0.025)
	optimizer.SetRiskFreeRate(0.04)

	result := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix, 2.5)

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

	optimizer := optimization.NewBlackLittermanOptimizer()
	optimizer.SetTau(0.025)
	optimizer.SetRiskFreeRate(0.04)

	result := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix, 2.5)

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

	optimizer := optimization.NewBlackLittermanOptimizer()
	optimizer.SetTau(0.025)
	optimizer.SetRiskFreeRate(0.04)

	result := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix, 2.5)

	for _, sym := range []string{"A", "B"} {
		ret, ok := result[sym]
		if !ok {
			t.Errorf("Missing result for %s", sym)
			continue
		}
		if math.IsNaN(ret) || math.IsInf(ret, 0) {
			t.Errorf("Invalid return for %s with degenerate covariance: %f", sym, ret)
		}
	}
}

func TestBLImplementations_NonSumToOneWeights(t *testing.T) {
	marketWeights := map[string]float64{"A": 0.3, "B": 0.3}
	covMatrix := map[string]map[string]float64{
		"A": {"A": 0.04, "B": 0.02},
		"B": {"A": 0.02, "B": 0.03},
	}

	optimizer := optimization.NewBlackLittermanOptimizer()
	optimizer.SetTau(0.025)
	optimizer.SetRiskFreeRate(0.04)

	result := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix, 2.5)

	t.Logf("Non-sum-to-one weights result: %+v", result)
}
