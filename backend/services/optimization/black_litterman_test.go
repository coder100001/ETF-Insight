package optimization

import (
	"math"
	"testing"
)

func TestNewBlackLittermanOptimizer(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	if optimizer == nil {
		t.Fatal("NewBlackLittermanOptimizer() returned nil")
	}

	if optimizer.Tau != 0.025 {
		t.Errorf("Expected default Tau 0.025, got %f", optimizer.Tau)
	}

	if optimizer.RiskFreeRate != 0.045 {
		t.Errorf("Expected default RiskFreeRate 0.045, got %f", optimizer.RiskFreeRate)
	}
}

func TestSetTau(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()
	newTau := 0.05
	optimizer.SetTau(newTau)

	if optimizer.Tau != newTau {
		t.Errorf("Expected Tau %f, got %f", newTau, optimizer.Tau)
	}
}

func TestBlackLitterman_SetRiskFreeRate(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()
	newRate := 0.05
	optimizer.SetRiskFreeRate(newRate)

	if optimizer.RiskFreeRate != newRate {
		t.Errorf("Expected RiskFreeRate %f, got %f", newRate, optimizer.RiskFreeRate)
	}
}

func TestNewBlackLittermanConstraint(t *testing.T) {
	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	constraint := NewBlackLittermanConstraint(symbols)

	if constraint == nil {
		t.Fatal("NewBlackLittermanConstraint() returned nil")
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

func TestBlackLitterman_Optimize_NormalCase(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.4,
		"GOOGL": 0.35,
		"MSFT":  0.25,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	views := []*InvestorView{
		{
			Type:       "absolute",
			Assets:     []string{"AAPL"},
			Return:     0.15,
			Confidence: 0.6,
		},
	}

	constraint := NewBlackLittermanConstraint([]string{"AAPL", "GOOGL", "MSFT"})

	result, err := optimizer.Optimize(marketWeights, covMatrix, views, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Optimize() returned nil result")
	}

	totalWeight := 0.0
	for symbol, weight := range result.OptimalWeights {
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
}

func TestBlackLitterman_Optimize_EmptyMarketWeights(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{}
	covMatrix := map[string]map[string]float64{}
	views := []*InvestorView{}
	constraint := NewBlackLittermanConstraint([]string{})

	_, err := optimizer.Optimize(marketWeights, covMatrix, views, constraint)

	if err == nil {
		t.Error("Expected error for empty market weights, got nil")
	}
}

func TestBlackLitterman_Optimize_NoViews(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.5,
		"GOOGL": 0.5,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	views := []*InvestorView{}
	constraint := NewBlackLittermanConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.Optimize(marketWeights, covMatrix, views, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Optimize() returned nil result")
	}

	for symbol := range marketWeights {
		if _, ok := result.PosteriorReturns[symbol]; !ok {
			t.Errorf("PosteriorReturns missing for symbol %s", symbol)
		}
	}
}

func TestBlackLitterman_Optimize_RelativeView(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.5,
		"GOOGL": 0.5,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	views := []*InvestorView{
		{
			Type:       "relative",
			Assets:     []string{"AAPL", "GOOGL"},
			Weights:    []float64{1.0, -1.0},
			Return:     0.03,
			Confidence: 0.7,
		},
	}

	constraint := NewBlackLittermanConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.Optimize(marketWeights, covMatrix, views, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Optimize() returned nil result")
	}

	totalWeight := 0.0
	for _, weight := range result.OptimalWeights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestBlackLitterman_Optimize_WithConstraints(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.4,
		"GOOGL": 0.35,
		"MSFT":  0.25,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02, "MSFT": 0.01},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03, "MSFT": 0.015},
		"MSFT":  {"AAPL": 0.01, "GOOGL": 0.015, "MSFT": 0.02},
	}

	views := []*InvestorView{
		{
			Type:       "absolute",
			Assets:     []string{"AAPL"},
			Return:     0.20,
			Confidence: 0.8,
		},
	}

	constraint := NewBlackLittermanConstraint([]string{"AAPL", "GOOGL", "MSFT"})
	constraint.MinWeight["AAPL"] = 0.3
	constraint.MaxWeight["MSFT"] = 0.2

	result, err := optimizer.Optimize(marketWeights, covMatrix, views, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	if result.OptimalWeights["AAPL"] < 0.3-0.01 {
		t.Errorf("AAPL weight should be >= 0.3, got %f", result.OptimalWeights["AAPL"])
	}

	if result.OptimalWeights["MSFT"] > 0.2+0.01 {
		t.Errorf("MSFT weight should be <= 0.2, got %f", result.OptimalWeights["MSFT"])
	}
}

func TestOptimizeWithViews(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.5,
		"GOOGL": 0.5,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	absoluteViews := map[string]float64{
		"AAPL": 0.15,
	}

	relativeViews := []*RelativeView{
		{
			Asset1:       "AAPL",
			Asset2:       "GOOGL",
			ExpectedDiff: 0.02,
			Confidence:   0.6,
		},
	}

	constraint := NewBlackLittermanConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.OptimizeWithViews(marketWeights, covMatrix, absoluteViews, relativeViews, constraint)

	if err != nil {
		t.Fatalf("OptimizeWithViews() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("OptimizeWithViews() returned nil result")
	}

	totalWeight := 0.0
	for _, weight := range result.OptimalWeights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestCalculateMarketImpliedReturns(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.5,
		"GOOGL": 0.5,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	riskAversion := 2.5

	impliedReturns := optimizer.CalculateMarketImpliedReturns(marketWeights, covMatrix, riskAversion)

	if impliedReturns == nil {
		t.Fatal("CalculateMarketImpliedReturns() returned nil")
	}

	for symbol := range marketWeights {
		if _, ok := impliedReturns[symbol]; !ok {
			t.Errorf("ImpliedReturns missing for symbol %s", symbol)
		}
	}
}

func TestBlackLittermanResult_Fields(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	marketWeights := map[string]float64{
		"AAPL":  0.5,
		"GOOGL": 0.5,
	}

	covMatrix := map[string]map[string]float64{
		"AAPL":  {"AAPL": 0.04, "GOOGL": 0.02},
		"GOOGL": {"AAPL": 0.02, "GOOGL": 0.03},
	}

	views := []*InvestorView{
		{
			Type:       "absolute",
			Assets:     []string{"AAPL"},
			Return:     0.15,
			Confidence: 0.6,
		},
	}

	constraint := NewBlackLittermanConstraint([]string{"AAPL", "GOOGL"})

	result, err := optimizer.Optimize(marketWeights, covMatrix, views, constraint)

	if err != nil {
		t.Fatalf("Optimize() returned error: %v", err)
	}

	requiredFields := []string{"AAPL", "GOOGL"}
	for _, symbol := range requiredFields {
		if _, ok := result.PriorReturns[symbol]; !ok {
			t.Errorf("PriorReturns missing for symbol %s", symbol)
		}
		if _, ok := result.PosteriorReturns[symbol]; !ok {
			t.Errorf("PosteriorReturns missing for symbol %s", symbol)
		}
		if _, ok := result.ImpliedReturns[symbol]; !ok {
			t.Errorf("ImpliedReturns missing for symbol %s", symbol)
		}
		if _, ok := result.OptimalWeights[symbol]; !ok {
			t.Errorf("OptimalWeights missing for symbol %s", symbol)
		}
	}

	if result.ExpectedReturn == 0 {
		t.Error("ExpectedReturn should not be zero")
	}

	if result.Volatility == 0 {
		t.Error("Volatility should not be zero")
	}
}

func TestCalculatePriorReturns(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	marketWeights := map[string]float64{"A": 0.5, "B": 0.5}
	symbols := []string{"A", "B"}
	delta := 2.5

	priorReturns := optimizer.calculatePriorReturns(Sigma, marketWeights, symbols, delta)

	if priorReturns == nil {
		t.Fatal("calculatePriorReturns() returned nil")
	}

	for _, symbol := range symbols {
		if _, ok := priorReturns[symbol]; !ok {
			t.Errorf("PriorReturns missing for symbol %s", symbol)
		}
	}
}

func TestBuildViewMatrices(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	views := []*InvestorView{
		{
			Type:       "absolute",
			Assets:     []string{"AAPL"},
			Return:     0.15,
			Confidence: 0.6,
		},
		{
			Type:       "relative",
			Assets:     []string{"AAPL", "GOOGL"},
			Weights:    []float64{1.0, -1.0},
			Return:     0.03,
			Confidence: 0.7,
		},
	}
	symbols := []string{"AAPL", "GOOGL", "MSFT"}

	P, Q, Omega := optimizer.buildViewMatrices(views, symbols)

	if P == nil {
		t.Fatal("buildViewMatrices() returned nil P matrix")
	}

	if Q == nil {
		t.Fatal("buildViewMatrices() returned nil Q vector")
	}

	if Omega == nil {
		t.Fatal("buildViewMatrices() returned nil Omega matrix")
	}

	if len(P) != len(views) {
		t.Errorf("Expected P matrix with %d rows, got %d", len(views), len(P))
	}

	if len(Q) != len(views) {
		t.Errorf("Expected Q vector with %d elements, got %d", len(views), len(Q))
	}
}

func TestBuildViewMatrices_EmptyViews(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	views := []*InvestorView{}
	symbols := []string{"AAPL", "GOOGL"}

	P, Q, Omega := optimizer.buildViewMatrices(views, symbols)

	if P != nil || Q != nil || Omega != nil {
		t.Error("Expected nil matrices for empty views")
	}
}

func TestCalculatePosteriorReturns(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	priorReturns := map[string]float64{"A": 0.10, "B": 0.08}
	P := [][]float64{{1.0, 0.0}}
	Q := []float64{0.15}
	Omega := [][]float64{{0.01}}
	symbols := []string{"A", "B"}

	posteriorReturns := optimizer.calculatePosteriorReturns(Sigma, priorReturns, P, Q, Omega, symbols)

	if posteriorReturns == nil {
		t.Fatal("calculatePosteriorReturns() returned nil")
	}

	for _, symbol := range symbols {
		if _, ok := posteriorReturns[symbol]; !ok {
			t.Errorf("PosteriorReturns missing for symbol %s", symbol)
		}
	}
}

func TestCalculatePosteriorReturns_NoViews(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	priorReturns := map[string]float64{"A": 0.10, "B": 0.08}
	symbols := []string{"A", "B"}

	posteriorReturns := optimizer.calculatePosteriorReturns(Sigma, priorReturns, nil, nil, nil, symbols)

	if posteriorReturns == nil {
		t.Fatal("calculatePosteriorReturns() returned nil")
	}

	for symbol, ret := range priorReturns {
		if posteriorReturns[symbol] != ret {
			t.Errorf("Without views, posterior should equal prior for %s: expected %f, got %f", symbol, ret, posteriorReturns[symbol])
		}
	}
}

func TestOptimizeWeights(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	Sigma := [][]float64{
		{0.04, 0.02},
		{0.02, 0.03},
	}
	posteriorReturns := map[string]float64{"A": 0.15, "B": 0.10}
	constraint := NewBlackLittermanConstraint([]string{"A", "B"})
	symbols := []string{"A", "B"}

	weights := optimizer.optimizeWeights(Sigma, posteriorReturns, constraint, symbols)

	if weights == nil {
		t.Fatal("optimizeWeights() returned nil")
	}

	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("Total weight should be ~1.0, got %f", totalWeight)
	}
}

func TestApplyConstraints_BlackLitterman(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

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
			constraint := &BlackLittermanConstraint{
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

func TestCalculateAverageConfidence(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()

	tests := []struct {
		name     string
		views    []*InvestorView
		expected float64
	}{
		{
			name: "single view",
			views: []*InvestorView{
				{Confidence: 0.6},
			},
			expected: 0.6,
		},
		{
			name: "multiple views",
			views: []*InvestorView{
				{Confidence: 0.6},
				{Confidence: 0.8},
			},
			expected: 0.7,
		},
		{
			name:     "empty views",
			views:    []*InvestorView{},
			expected: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := optimizer.calculateAverageConfidence(tc.views)

			if result != tc.expected {
				t.Errorf("Expected %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestViewsFromMatrices(t *testing.T) {
	P := [][]float64{
		{1.0, 0.0, 0.0},
		{1.0, -1.0, 0.0},
	}
	Q := []float64{0.15, 0.03}
	symbols := []string{"AAPL", "GOOGL", "MSFT"}

	views := viewsFromMatrices(P, Q, symbols)

	if len(views) != 2 {
		t.Fatalf("Expected 2 views, got %d", len(views))
	}

	if views[0].Type != "absolute" {
		t.Errorf("First view should be absolute, got %s", views[0].Type)
	}

	if views[1].Type != "relative" {
		t.Errorf("Second view should be relative, got %s", views[1].Type)
	}
}

func TestInvestorView_Fields(t *testing.T) {
	view := &InvestorView{
		Type:        "absolute",
		Assets:      []string{"AAPL"},
		Weights:     []float64{1.0},
		Return:      0.15,
		Confidence:  0.6,
		Description: "AAPL will outperform",
	}

	if view.Type != "absolute" {
		t.Errorf("Expected Type 'absolute', got %s", view.Type)
	}

	if len(view.Assets) != 1 || view.Assets[0] != "AAPL" {
		t.Errorf("Expected Assets ['AAPL'], got %v", view.Assets)
	}

	if view.Return != 0.15 {
		t.Errorf("Expected Return 0.15, got %f", view.Return)
	}

	if view.Confidence != 0.6 {
		t.Errorf("Expected Confidence 0.6, got %f", view.Confidence)
	}
}

func TestRelativeView_Fields(t *testing.T) {
	view := &RelativeView{
		Asset1:       "AAPL",
		Asset2:       "GOOGL",
		ExpectedDiff: 0.03,
		Confidence:   0.7,
	}

	if view.Asset1 != "AAPL" {
		t.Errorf("Expected Asset1 'AAPL', got %s", view.Asset1)
	}

	if view.Asset2 != "GOOGL" {
		t.Errorf("Expected Asset2 'GOOGL', got %s", view.Asset2)
	}

	if view.ExpectedDiff != 0.03 {
		t.Errorf("Expected ExpectedDiff 0.03, got %f", view.ExpectedDiff)
	}

	if view.Confidence != 0.7 {
		t.Errorf("Expected Confidence 0.7, got %f", view.Confidence)
	}
}

func TestBlackLitterman_Optimize_EmptySymbols(t *testing.T) {
	optimizer := NewBlackLittermanOptimizer()
	optimizer.SetTau(0.025)
	optimizer.SetRiskFreeRate(0.04)

	_, err := optimizer.Optimize(map[string]float64{}, map[string]map[string]float64{}, nil, nil)
	if err == nil {
		t.Error("Empty input should return error")
	}
}

func TestBlackLitterman_OptimizeWithViews_NoViews(t *testing.T) {
	returns := map[string]float64{"A": 0.08, "B": 0.06}
	cov := map[string]map[string]float64{
		"A": {"A": 0.04, "B": 0.02},
		"B": {"A": 0.02, "B": 0.03},
	}

	optimizer := NewBlackLittermanOptimizer()
	optimizer.SetTau(0.025)
	optimizer.SetRiskFreeRate(0.04)

	result, err := optimizer.OptimizeWithViews(returns, cov, nil, nil, nil)
	if err != nil {
		t.Fatalf("No views should work: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}
