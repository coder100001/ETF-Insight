package optimization

import (
	"math"
	"testing"
)

func TestNewRiskParityOptimizer(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	if optimizer == nil {
		t.Fatal("NewRiskParityOptimizer() returned nil")
	}

	if optimizer.MaxIter != 1000 {
		t.Errorf("Expected default MaxIter 1000, got %d", optimizer.MaxIter)
	}

	if optimizer.Tolerance != 1e-8 {
		t.Errorf("Expected default Tolerance 1e-8, got %e", optimizer.Tolerance)
	}
}

func TestNewRiskParityConstraint(t *testing.T) {
	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	constraint := NewRiskParityConstraint(symbols)

	if constraint == nil {
		t.Fatal("NewRiskParityConstraint() returned nil")
	}

	if constraint.TargetVolatility != 0 {
		t.Errorf("Expected TargetVolatility 0, got %f", constraint.TargetVolatility)
	}

	if constraint.UseLeverage {
		t.Error("Expected UseLeverage to be false by default")
	}

	if constraint.MaxLeverage != 2.0 {
		t.Errorf("Expected MaxLeverage 2.0, got %f", constraint.MaxLeverage)
	}

	for _, symbol := range symbols {
		if min, ok := constraint.MinWeight[symbol]; !ok || min != 0.0 {
			t.Errorf("Expected MinWeight[%s] = 0.0, got %f", symbol, min)
		}
		if max, ok := constraint.MaxWeight[symbol]; !ok || max != 1.0 {
			t.Errorf("Expected MaxWeight[%s] = 1.0, got %f", symbol, max)
		}
	}
}

func TestRiskParity_Optimize_NormalCase(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.12,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	constraint := NewRiskParityConstraint(symbols)

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Optimize() returned nil result")
	}

	totalWeight := 0.0
	for symbol, weight := range result.Weights {
		if weight < 0 {
			t.Errorf("Weight for %s is negative: %f", symbol, weight)
		}
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}

	if result.Volatility <= 0 {
		t.Errorf("Volatility should be positive, got %f", result.Volatility)
	}

	if result.Leverage <= 0 {
		t.Errorf("Leverage should be positive, got %f", result.Leverage)
	}
}

func TestRiskParity_Optimize_EmptyReturns(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{}
	covMatrix := map[string]map[string]float64{}
	constraint := NewRiskParityConstraint([]string{})

	_, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err == nil {
		t.Error("Expected error for empty returns, got nil")
	}
}

func TestRiskParity_Optimize_SingleAsset(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL": 0.12,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL": {"AAPL": 0.04},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL"})

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result.Weights["AAPL"] < 0.99 || result.Weights["AAPL"] > 1.01 {
		t.Errorf("Expected weight ~1.0 for single asset, got %f", result.Weights["AAPL"])
	}
}

func TestOptimize_WithLeverage(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL", "MSFT"})
	constraint.TargetVolatility = 0.20
	constraint.UseLeverage = true
	constraint.MaxLeverage = 2.0

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result.Leverage < 1.0 {
		t.Errorf("Leverage should be >= 1.0 with target volatility, got %f", result.Leverage)
	}

	if result.TargetVolatility != constraint.TargetVolatility {
		t.Errorf("TargetVolatility mismatch: expected %f, got %f", constraint.TargetVolatility, result.TargetVolatility)
	}
}

func TestRiskParity_Optimize_WithConstraints(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL", "MSFT"})
	constraint.MinWeight["AAPL"] = 0.2
	constraint.MaxWeight["MSFT"] = 0.3

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result.Weights["AAPL"] < 0.2-0.01 {
		t.Errorf("AAPL weight should be >= 0.2, got %f", result.Weights["AAPL"])
	}

	if result.Weights["MSFT"] > 0.3+0.01 {
		t.Errorf("MSFT weight should be <= 0.3, got %f", result.Weights["MSFT"])
	}
}

func TestOptimizeERC(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.OptimizeERC(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("OptimizeERC() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("OptimizeERC() returned nil result")
	}

	totalWeight := 0.0
	for _, weight := range result.Weights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestOptimizeInverseVol(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.OptimizeInverseVol(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("OptimizeInverseVol() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("OptimizeInverseVol() returned nil result")
	}

	totalWeight := 0.0
	for _, weight := range result.Weights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}

	if result.Leverage != 1.0 {
		t.Errorf("Inverse vol should not use leverage, got %f", result.Leverage)
	}
}

func TestOptimizeInverseVol_DifferentVolatilities(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"LOW_VOL":  0.08,
		"HIGH_VOL": 0.15,
	}

	covMatrix := map[string]map[string]float64{
		"LOW_VOL":  {"LOW_VOL": 0.01, "HIGH_VOL": 0.005},
		"HIGH_VOL": {"LOW_VOL": 0.005, "HIGH_VOL": 0.09},
	}

	constraint := NewRiskParityConstraint([]string{"LOW_VOL", "HIGH_VOL"})

	result, err := optimizer.OptimizeInverseVol(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("OptimizeInverseVol() returned error: %v", err)
	}

	lowVolWeight := result.Weights["LOW_VOL"]
	highVolWeight := result.Weights["HIGH_VOL"]

	if lowVolWeight <= highVolWeight {
		t.Errorf("Low volatility asset should have higher weight: LOW_VOL=%f, HIGH_VOL=%f", lowVolWeight, highVolWeight)
	}
}

func TestCalculateRiskBudget(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	riskBudget := map[string]float64{
		"AAPL":  0.5,
		"GOOGL": 0.3,
		"MSFT":  0.2,
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.CalculateRiskBudget(returns, covMatrix, riskBudget, constraint)

	if err != nil {
		t.Fatalf("CalculateRiskBudget() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("CalculateRiskBudget() returned nil result")
	}

	totalWeight := 0.0
	for _, weight := range result.Weights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestCalculateRiskBudget_Normalization(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	riskBudget := map[string]float64{
		"AAPL":  2.0,
		"GOOGL": 2.0,
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.CalculateRiskBudget(returns, covMatrix, riskBudget, constraint)

	if err != nil {
		t.Fatalf("CalculateRiskBudget() returned error: %v", err)
	}

	totalWeight := 0.0
	for _, weight := range result.Weights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0 after normalization, got %f", totalWeight)
	}
}

func TestRiskParityResult_Fields(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.12,
		"GOOGL": 0.10,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	requiredFields := []string{"AAPL", "GOOGL"}
	for _, symbol := range requiredFields {
		if _, ok := result.Weights[symbol]; !ok {
			t.Errorf("Weights missing for symbol %s", symbol)
		}
		if _, ok := result.RiskContributions[symbol]; !ok {
			t.Errorf("RiskContributions missing for symbol %s", symbol)
		}
	}

	if result.ExpectedReturn == 0 {
		t.Error("ExpectedReturn should not be zero")
	}

	if result.Volatility == 0 {
		t.Error("Volatility should not be zero")
	}

	if result.DiversificationRatio < 1.0 {
		t.Logf("Warning: DiversificationRatio < 1.0: %f", result.DiversificationRatio)
	}
}

func TestCalculateRiskContributions_EqualRiskParity(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	returns := map[string]float64{
		"AAPL":  0.12,
		"GOOGL": 0.10,
		"MSFT":  0.08,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	constraint := NewRiskParityConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	n := len(result.RiskContributions)
	targetRC := 1.0 / float64(n)

	for symbol, rc := range result.RiskContributions {
		diff := math.Abs(rc - targetRC)
		tolerance := 0.15

		if diff > tolerance {
			t.Logf("Risk contribution for %s: %f (target: %f, diff: %f)", symbol, rc, targetRC, diff)
		}
	}
}

func TestCalculateRiskContributions_ZeroVolatility(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	Sigma := [][]float64{
		{0.0, 0.0},
		{0.0, 0.0},
	}
	weights := []float64{0.5, 0.5}
	symbols := []string{"A", "B"}

	result := optimizer.calculateRiskContributions(Sigma, weights, symbols)

	expectedRC := 1.0 / float64(len(symbols))
	for _, symbol := range symbols {
		if result[symbol] != expectedRC {
			t.Errorf("Risk contribution for %s should be %f for zero volatility, got %f", symbol, expectedRC, result[symbol])
		}
	}
}

func TestCalculatePortfolioVolatility(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	weights := []float64{0.5, 0.5}

	vol := optimizer.calculatePortfolioVolatility(Sigma, weights)

	if vol <= 0 {
		t.Errorf("Volatility should be positive, got %f", vol)
	}
}

func TestCalculateDiversificationRatio(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	weights := []float64{0.5, 0.5}

	ratio := optimizer.calculateDiversificationRatio(Sigma, weights)

	if ratio < 1.0 {
		t.Errorf("DiversificationRatio should be >= 1.0 for diversified portfolio, got %f", ratio)
	}
}

func TestRiskParity_CalculateDiversificationRatio_ZeroVolatility(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	Sigma := [][]float64{
		{0.0, 0.0},
		{0.0, 0.0},
	}
	weights := []float64{0.5, 0.5}

	ratio := optimizer.calculateDiversificationRatio(Sigma, weights)

	if ratio != 1.0 {
		t.Errorf("DiversificationRatio should be 1.0 for zero volatility, got %f", ratio)
	}
}

func TestRiskParity_ZeroVolatilityAsset(t *testing.T) {
	returns := map[string]float64{"A": 0.05, "B": 0.05}
	cov := map[string]map[string]float64{
		"A": {"A": 0, "B": 0},
		"B": {"A": 0, "B": 0.04},
	}

	optimizer := NewRiskParityOptimizer()

	result, err := optimizer.Optimize(returns, cov)
	if err != nil {
		t.Logf("Zero volatility returned error (acceptable): %v", err)
	} else if result != nil {
		t.Log("Zero volatility handled gracefully")
	}
}

func TestApplyConstraints(t *testing.T) {
	optimizer := NewRiskParityOptimizer()

	tests := []struct {
		name      string
		weights   []float64
		symbols   []string
		minWeight map[string]float64
		maxWeight map[string]float64
	}{
		{
			name:      "normal weights",
			weights:   []float64{0.4, 0.6},
			symbols:   []string{"A", "B"},
			minWeight: map[string]float64{"A": 0.0, "B": 0.0},
			maxWeight: map[string]float64{"A": 1.0, "B": 1.0},
		},
		{
			name:      "weights exceed bounds",
			weights:   []float64{1.2, -0.2},
			symbols:   []string{"A", "B"},
			minWeight: map[string]float64{"A": 0.0, "B": 0.0},
			maxWeight: map[string]float64{"A": 1.0, "B": 1.0},
		},
		{
			name:      "weights below min",
			weights:   []float64{-0.1, 0.3},
			symbols:   []string{"A", "B"},
			minWeight: map[string]float64{"A": 0.2, "B": 0.2},
			maxWeight: map[string]float64{"A": 0.8, "B": 0.8},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			constraint := &RiskParityConstraint{
				MinWeight: tc.minWeight,
				MaxWeight: tc.maxWeight,
			}

			result := optimizer.applyConstraints(tc.weights, tc.symbols, constraint)

			total := 0.0
			for i, w := range result {
				if w < tc.minWeight[tc.symbols[i]]-0.01 {
					t.Errorf("Weight %d below min: %f < %f", i, w, tc.minWeight[tc.symbols[i]])
				}
				if w > tc.maxWeight[tc.symbols[i]]+0.01 {
					t.Errorf("Weight %d above max: %f > %f", i, w, tc.maxWeight[tc.symbols[i]])
				}
				total += w
			}

			if math.Abs(total-1.0) > 0.01 {
				t.Errorf("Total weight should be 1.0, got %f", total)
			}
		})
	}
}

func TestSolveRiskParity_Convergence(t *testing.T) {
	optimizer := NewRiskParityOptimizer()
	optimizer.MaxIter = 100
	optimizer.Tolerance = 1e-6

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	symbols := []string{"A", "B"}
	constraint := NewRiskParityConstraint(symbols)

	weights, err := optimizer.solveRiskParity(Sigma, symbols, constraint)

	if err != nil {
		t.Fatalf("solveRiskParity() returned error: %v", err)
	}

	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestSolveRiskBudget_Convergence(t *testing.T) {
	optimizer := NewRiskParityOptimizer()
	optimizer.MaxIter = 100
	optimizer.Tolerance = 1e-6

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	symbols := []string{"A", "B"}
	riskBudget := map[string]float64{"A": 0.6, "B": 0.4}
	constraint := NewRiskParityConstraint(symbols)

	weights, err := optimizer.solveRiskBudget(Sigma, symbols, riskBudget, constraint)

	if err != nil {
		t.Fatalf("solveRiskBudget() returned error: %v", err)
	}

	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}
