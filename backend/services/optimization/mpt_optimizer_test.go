package optimization

import (
	"math"
	"testing"
)

func TestNewMPTOptimizer(t *testing.T) {
	optimizer := NewMPTOptimizer()

	if optimizer == nil {
		t.Fatal("NewMPTOptimizer() returned nil")
	}

	if optimizer.RiskFreeRate != 0.045 {
		t.Errorf("Expected default RiskFreeRate 0.045, got %f", optimizer.RiskFreeRate)
	}

	if optimizer.MaxIter != 1000 {
		t.Errorf("Expected default MaxIter 1000, got %d", optimizer.MaxIter)
	}

	if optimizer.Tolerance != 1e-8 {
		t.Errorf("Expected default Tolerance 1e-8, got %e", optimizer.Tolerance)
	}
}

func TestSetRiskFreeRate(t *testing.T) {
	optimizer := NewMPTOptimizer()
	newRate := 0.05
	optimizer.SetRiskFreeRate(newRate)

	if optimizer.RiskFreeRate != newRate {
		t.Errorf("Expected RiskFreeRate %f, got %f", newRate, optimizer.RiskFreeRate)
	}
}

func TestNewConstraint(t *testing.T) {
	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	constraint := NewConstraint(symbols)

	if constraint == nil {
		t.Fatal("NewConstraint() returned nil")
	}

	if constraint.TotalWeight != 1.0 {
		t.Errorf("Expected TotalWeight 1.0, got %f", constraint.TotalWeight)
	}

	if constraint.AllowShort {
		t.Error("Expected AllowShort to be false by default")
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

func TestConstraintSetters(t *testing.T) {
	symbols := []string{"AAPL", "GOOGL"}
	constraint := NewConstraint(symbols)

	constraint.SetMinWeight("AAPL", 0.1)
	if constraint.MinWeight["AAPL"] != 0.1 {
		t.Errorf("Expected MinWeight[AAPL] = 0.1, got %f", constraint.MinWeight["AAPL"])
	}

	constraint.SetMaxWeight("GOOGL", 0.5)
	if constraint.MaxWeight["GOOGL"] != 0.5 {
		t.Errorf("Expected MaxWeight[GOOGL] = 0.5, got %f", constraint.MaxWeight["GOOGL"])
	}

	constraint.SetSectorLimit("Technology", 0.4)
	if constraint.SectorLimits["Technology"] != 0.4 {
		t.Errorf("Expected SectorLimits[Technology] = 0.4, got %f", constraint.SectorLimits["Technology"])
	}
}

func TestOptimize_NormalCase(t *testing.T) {
	optimizer := NewMPTOptimizer()

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
	constraint := NewConstraint(symbols)

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

	if result.HerfindahlIndex < 0 || result.HerfindahlIndex > 1 {
		t.Errorf("HerfindahlIndex should be between 0 and 1, got %f", result.HerfindahlIndex)
	}

	if result.DiversificationRatio < 1.0 {
		t.Logf("Warning: DiversificationRatio < 1.0: %f", result.DiversificationRatio)
	}
}

func TestOptimize_EmptyReturns(t *testing.T) {
	optimizer := NewMPTOptimizer()

	returns := map[string]float64{}
	covMatrix := map[string]map[string]float64{}
	constraint := NewConstraint([]string{})

	_, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err == nil {
		t.Error("Expected error for empty returns, got nil")
	}
}

func TestOptimize_MissingCovarianceData(t *testing.T) {
	optimizer := NewMPTOptimizer()

	returns := map[string]float64{
		"AAPL":  0.12,
		"GOOGL": 0.10,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL": {"AAPL": 0.04},
	}

	constraint := NewConstraint([]string{"AAPL", "GOOGL"})

	_, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err == nil {
		t.Error("Expected error for missing covariance data, got nil")
	}
}

func TestOptimize_SingleAsset(t *testing.T) {
	optimizer := NewMPTOptimizer()

	returns := map[string]float64{
		"AAPL": 0.12,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL": {"AAPL": 0.04},
	}

	constraint := NewConstraint([]string{"AAPL"})

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result.Weights["AAPL"] < 0.99 || result.Weights["AAPL"] > 1.01 {
		t.Errorf("Expected weight ~1.0 for single asset, got %f", result.Weights["AAPL"])
	}
}

func TestOptimize_WithConstraints(t *testing.T) {
	optimizer := NewMPTOptimizer()

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

	constraint := NewConstraint([]string{"AAPL", "GOOGL", "MSFT"})
	constraint.SetMinWeight("AAPL", 0.2)
	constraint.SetMaxWeight("MSFT", 0.3)

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result.Weights["AAPL"] < 0.2 {
		t.Errorf("AAPL weight should be >= 0.2, got %f", result.Weights["AAPL"])
	}

	if result.Weights["MSFT"] > 0.3+0.01 {
		t.Errorf("MSFT weight should be <= 0.3, got %f", result.Weights["MSFT"])
	}
}

func TestOptimizeMaxSharpe(t *testing.T) {
	optimizer := NewMPTOptimizer()

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

	constraint := NewConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.OptimizeMaxSharpe(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("OptimizeMaxSharpe() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("OptimizeMaxSharpe() returned nil result")
	}

	if result.SharpeRatio <= 0 {
		t.Errorf("SharpeRatio should be positive, got %f", result.SharpeRatio)
	}
}

func TestOptimizeMinVolatility(t *testing.T) {
	optimizer := NewMPTOptimizer()

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

	constraint := NewConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.OptimizeMinVolatility(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("OptimizeMinVolatility() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("OptimizeMinVolatility() returned nil result")
	}

	if result.Volatility <= 0 {
		t.Errorf("Volatility should be positive, got %f", result.Volatility)
	}
}

func TestOptimizeForTargetReturn(t *testing.T) {
	optimizer := NewMPTOptimizer()

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

	constraint := NewConstraint([]string{"AAPL", "GOOGL", "MSFT"})
	targetReturn := 0.10

	result, err := optimizer.OptimizeForTargetReturn(returns, covMatrix, constraint, targetReturn)

	if err != nil {
		t.Fatalf("OptimizeForTargetReturn() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("OptimizeForTargetReturn() returned nil result")
	}

	returnDiff := math.Abs(result.ExpectedReturn - targetReturn)
	tolerance := 0.05
	if returnDiff > tolerance {
		t.Errorf("Expected return should be close to %f, got %f (diff: %f)", targetReturn, result.ExpectedReturn, returnDiff)
	}
}

func TestCalculateEfficientFrontier(t *testing.T) {
	optimizer := NewMPTOptimizer()

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

	constraint := NewConstraint([]string{"AAPL", "GOOGL", "MSFT"})
	numPoints := 10

	frontier, err := optimizer.CalculateEfficientFrontier(returns, covMatrix, constraint, numPoints)

	if err != nil {
		t.Fatalf("CalculateEfficientFrontier() returned error: %v", err)
	}

	if len(frontier) == 0 {
		t.Fatal("CalculateEfficientFrontier() returned empty frontier")
	}

	for i, point := range frontier {
		if point.MinVolatility <= 0 {
			t.Errorf("Point %d: Volatility should be positive, got %f", i, point.MinVolatility)
		}
	}

	for i := 1; i < len(frontier); i++ {
		if frontier[i].MinVolatility < frontier[i-1].MinVolatility {
			t.Errorf("Frontier should be sorted by volatility: point %d has lower vol than point %d", i, i-1)
		}
	}
}

func TestCalculateEfficientFrontier_MinimumPoints(t *testing.T) {
	optimizer := NewMPTOptimizer()

	returns := map[string]float64{
		"AAPL":  0.15,
		"GOOGL": 0.10,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	constraint := NewConstraint([]string{"AAPL", "GOOGL"})

	frontier, err := optimizer.CalculateEfficientFrontier(returns, covMatrix, constraint, 1)

	if err != nil {
		t.Fatalf("CalculateEfficientFrontier() returned error: %v", err)
	}

	if len(frontier) < 2 {
		t.Errorf("Expected at least 2 points, got %d", len(frontier))
	}
}

func TestPortfolioResult_Fields(t *testing.T) {
	optimizer := NewMPTOptimizer()

	returns := map[string]float64{
		"AAPL":  0.12,
		"GOOGL": 0.10,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	constraint := NewConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.Optimize(returns, covMatrix, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	requiredFields := []string{"AAPL", "GOOGL"}
	for _, symbol := range requiredFields {
		if _, ok := result.Weights[symbol]; !ok {
			t.Errorf("Weights missing for symbol %s", symbol)
		}
		if _, ok := result.RiskContribution[symbol]; !ok {
			t.Errorf("RiskContribution missing for symbol %s", symbol)
		}
	}

	if result.ExpectedReturn == 0 {
		t.Error("ExpectedReturn should not be zero")
	}

	if result.Volatility == 0 {
		t.Error("Volatility should not be zero")
	}

	if result.SharpeRatio == 0 {
		t.Error("SharpeRatio should not be zero")
	}
}

func TestDecimalConversions(t *testing.T) {
	testCases := []float64{0.0, 1.0, 0.5, 0.123456789, -0.5}

	for _, tc := range testCases {
		decimal := Float64ToDecimal(tc)
		back := DecimalToFloat64(decimal)

		diff := math.Abs(back - tc)
		tolerance := 1e-10

		if diff > tolerance {
			t.Errorf("Float64ToDecimal/DecimalToFloat64 conversion failed: input=%f, output=%f, diff=%e", tc, back, diff)
		}
	}
}

func TestSolveQuadraticProgramming_Convergence(t *testing.T) {
	optimizer := NewMPTOptimizer()
	optimizer.MaxIter = 100
	optimizer.Tolerance = 1e-6

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	mu := []float64{0.12, 0.10}
	minWeights := []float64{0.0, 0.0}
	maxWeights := []float64{1.0, 1.0}

	weights, err := optimizer.solveQuadraticProgramming(Sigma, mu, minWeights, maxWeights, 1.0)

	if err != nil {
		t.Fatalf("solveQuadraticProgramming() returned error: %v", err)
	}

	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestProjectToConstraints(t *testing.T) {
	optimizer := NewMPTOptimizer()

	tests := []struct {
		name       string
		weights    []float64
		minWeights []float64
		maxWeights []float64
		total      float64
	}{
		{
			name:       "normal weights",
			weights:    []float64{0.4, 0.6},
			minWeights: []float64{0.0, 0.0},
			maxWeights: []float64{1.0, 1.0},
			total:      1.0,
		},
		{
			name:       "weights exceed bounds",
			weights:    []float64{1.2, -0.2},
			minWeights: []float64{0.0, 0.0},
			maxWeights: []float64{1.0, 1.0},
			total:      1.0,
		},
		{
			name:       "weights below min",
			weights:    []float64{-0.1, 0.3},
			minWeights: []float64{0.2, 0.2},
			maxWeights: []float64{0.8, 0.8},
			total:      1.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := optimizer.projectToConstraints(tc.weights, tc.minWeights, tc.maxWeights, tc.total)

			total := 0.0
			for i, w := range result {
				if w < tc.minWeights[i]-0.01 {
					t.Errorf("Weight %d below min: %f < %f", i, w, tc.minWeights[i])
				}
				if w > tc.maxWeights[i]+0.01 {
					t.Errorf("Weight %d above max: %f > %f", i, w, tc.maxWeights[i])
				}
				total += w
			}

			if math.Abs(total-tc.total) > 0.01 {
				t.Errorf("Total weight should be %f, got %f", tc.total, total)
			}
		})
	}
}

func TestCalculateRiskContribution_ZeroVolatility(t *testing.T) {
	optimizer := NewMPTOptimizer()

	Sigma := [][]float64{
		{0.0, 0.0},
		{0.0, 0.0},
	}
	weights := []float64{0.5, 0.5}
	symbols := []string{"A", "B"}

	result := optimizer.calculateRiskContribution(Sigma, weights, symbols)

	for _, symbol := range symbols {
		if result[symbol] != 0 {
			t.Errorf("Risk contribution for %s should be 0 for zero volatility, got %f", symbol, result[symbol])
		}
	}
}

func TestCalculateDiversificationRatio_ZeroVolatility(t *testing.T) {
	optimizer := NewMPTOptimizer()

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
