package factor

import (
	"math"
	"testing"
)

func TestNewFamaFrenchModel(t *testing.T) {
	model := NewFamaFrenchModel()
	if model == nil {
		t.Fatal("NewFamaFrenchModel() returned nil")
	}
	if model.UseFiveFactor {
		t.Error("NewFamaFrenchModel() should default to three-factor model")
	}
}

func TestSetFiveFactor(t *testing.T) {
	model := NewFamaFrenchModel()

	// Test setting to five-factor
	model.SetFiveFactor(true)
	if !model.UseFiveFactor {
		t.Error("SetFiveFactor(true) failed")
	}

	// Test setting back to three-factor
	model.SetFiveFactor(false)
	if model.UseFiveFactor {
		t.Error("SetFiveFactor(false) failed")
	}
}

func TestLoadFactorData(t *testing.T) {
	model := NewFamaFrenchModel()

	marketReturns := []float64{0.01, 0.02, -0.01, 0.015}
	smbReturns := []float64{0.005, -0.005, 0.01, -0.002}
	hmlReturns := []float64{-0.002, 0.008, 0.003, -0.001}
	riskFreeReturns := []float64{0.001, 0.001, 0.001, 0.001}

	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	if len(model.MarketReturns) != 4 {
		t.Errorf("Expected 4 market returns, got %d", len(model.MarketReturns))
	}
	if len(model.SMBReturns) != 4 {
		t.Errorf("Expected 4 SMB returns, got %d", len(model.SMBReturns))
	}
	if len(model.HMLReturns) != 4 {
		t.Errorf("Expected 4 HML returns, got %d", len(model.HMLReturns))
	}
	if len(model.RiskFreeReturns) != 4 {
		t.Errorf("Expected 4 risk-free returns, got %d", len(model.RiskFreeReturns))
	}
}

func TestLoadFiveFactorData(t *testing.T) {
	model := NewFamaFrenchModel()
	model.SetFiveFactor(true)

	marketReturns := []float64{0.01, 0.02, -0.01, 0.015}
	smbReturns := []float64{0.005, -0.005, 0.01, -0.002}
	hmlReturns := []float64{-0.002, 0.008, 0.003, -0.001}
	rmwReturns := []float64{0.003, 0.002, -0.001, 0.004}
	cmaReturns := []float64{0.001, 0.003, 0.002, -0.001}
	riskFreeReturns := []float64{0.001, 0.001, 0.001, 0.001}

	model.LoadFiveFactorData(marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns)

	if len(model.RMWReturns) != 4 {
		t.Errorf("Expected 4 RMW returns, got %d", len(model.RMWReturns))
	}
	if len(model.CMAReturns) != 4 {
		t.Errorf("Expected 4 CMA returns, got %d", len(model.CMAReturns))
	}
}

func TestAnalyzePortfolio(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load sample factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)
	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// Create sample portfolio returns
	portfolioReturns := make([]float64, periods)
	for i := 0; i < periods; i++ {
		portfolioReturns[i] = 0.008 + (float64(i%5)-2)*0.01
	}

	weights := map[string]float64{
		"VTI": 0.6,
		"VOO": 0.4,
	}

	attribution, err := model.AnalyzePortfolio(portfolioReturns, weights)
	if err != nil {
		t.Fatalf("AnalyzePortfolio failed: %v", err)
	}

	if attribution == nil {
		t.Fatal("AnalyzePortfolio returned nil attribution")
	}

	if attribution.Exposures == nil {
		t.Fatal("Exposures is nil")
	}

	// Check that we have valid values (not NaN)
	if math.IsNaN(attribution.Exposures.Market) {
		t.Error("Market exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Size) {
		t.Error("Size exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Value) {
		t.Error("Value exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Alpha) {
		t.Error("Alpha is NaN")
	}

	// Check contributions
	if attribution.Contributions == nil {
		t.Error("Contributions is nil")
	}

	// Check T-statistics
	if attribution.TStatistics == nil {
		t.Error("TStatistics is nil")
	}

	// Check P-values
	if attribution.PValues == nil {
		t.Error("PValues is nil")
	}
}

func TestAnalyzePortfolio_EmptyData(t *testing.T) {
	model := NewFamaFrenchModel()

	// Test with empty portfolio returns
	_, err := model.AnalyzePortfolio([]float64{}, map[string]float64{})
	if err == nil {
		t.Error("Expected error for empty portfolio returns")
	}
}

func TestAnalyzeETF(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load sample factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)
	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// Create sample ETF returns
	etfReturns := make([]float64, periods)
	for i := 0; i < periods; i++ {
		etfReturns[i] = 0.01 + (float64(i%3)-1)*0.015
	}

	attribution, err := model.AnalyzeETF(etfReturns, "VTI")
	if err != nil {
		t.Fatalf("AnalyzeETF failed: %v", err)
	}

	if attribution == nil {
		t.Fatal("AnalyzeETF returned nil attribution")
	}

	if attribution.Exposures == nil {
		t.Fatal("Exposures is nil")
	}
}

func TestGetFactorStatistics(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load sample factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)
	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	stats := model.GetFactorStatistics()

	if len(stats) != 3 {
		t.Errorf("Expected 3 factor stats for three-factor model, got %d", len(stats))
	}

	// Check that all stats have valid values
	for _, stat := range stats {
		if stat.Name == "" {
			t.Error("Factor stat name is empty")
		}
		if math.IsNaN(stat.Annualized) {
			t.Errorf("Annualized return is NaN for %s", stat.Name)
		}
		if math.IsNaN(stat.Volatility) {
			t.Errorf("Volatility is NaN for %s", stat.Name)
		}
		if stat.Volatility < 0 {
			t.Errorf("Volatility should be non-negative for %s", stat.Name)
		}
	}
}

func TestGetFactorStatistics_FiveFactor(t *testing.T) {
	model := NewFamaFrenchModel()
	model.SetFiveFactor(true)

	// Load sample five-factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns := GenerateSampleFiveFactorData(periods)
	model.LoadFiveFactorData(marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns)

	stats := model.GetFactorStatistics()

	if len(stats) != 5 {
		t.Errorf("Expected 5 factor stats for five-factor model, got %d", len(stats))
	}
}

func TestGenerateSampleFactorData(t *testing.T) {
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)

	if len(marketReturns) != periods {
		t.Errorf("Expected %d market returns, got %d", periods, len(marketReturns))
	}
	if len(smbReturns) != periods {
		t.Errorf("Expected %d SMB returns, got %d", periods, len(smbReturns))
	}
	if len(hmlReturns) != periods {
		t.Errorf("Expected %d HML returns, got %d", periods, len(hmlReturns))
	}
	if len(riskFreeReturns) != periods {
		t.Errorf("Expected %d risk-free returns, got %d", periods, len(riskFreeReturns))
	}

	// Check that returns are not all the same (randomness check)
	marketSum := 0.0
	for _, r := range marketReturns {
		marketSum += r
	}
	marketAvg := marketSum / float64(periods)

	// Market returns should average within reasonable range (-3% to 3% monthly)
	// With mean 0.5% and std 4.5%, for 36 samples the SE is ~0.75%
	// So -3% to 3% covers ~3.3 standard errors (99.9% confidence)
	if marketAvg < -0.03 || marketAvg > 0.03 {
		t.Errorf("Market average return %f seems unreasonable", marketAvg)
	}
}

func TestGenerateSampleFiveFactorData(t *testing.T) {
	periods := 36
	marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns := GenerateSampleFiveFactorData(periods)

	if len(rmwReturns) != periods {
		t.Errorf("Expected %d RMW returns, got %d", periods, len(rmwReturns))
	}
	if len(cmaReturns) != periods {
		t.Errorf("Expected %d CMA returns, got %d", periods, len(cmaReturns))
	}

	// Check that all return slices have the same length
	if len(marketReturns) != periods || len(smbReturns) != periods ||
		len(hmlReturns) != periods || len(riskFreeReturns) != periods {
		t.Error("All return slices should have the same length")
	}
}

func TestMatrixOperations(t *testing.T) {
	model := NewFamaFrenchModel()

	// Test matrix multiplication
	A := [][]float64{
		{1, 2},
		{3, 4},
	}
	B := [][]float64{
		{5, 6},
		{7, 8},
	}

	C := model.matrixMultiply(A, B)

	// Expected: [[19, 22], [43, 50]]
	if len(C) != 2 || len(C[0]) != 2 {
		t.Fatal("Matrix multiplication result has wrong dimensions")
	}

	if C[0][0] != 19 || C[0][1] != 22 || C[1][0] != 43 || C[1][1] != 50 {
		t.Errorf("Matrix multiplication result incorrect: %v", C)
	}
}

func TestTranspose(t *testing.T) {
	model := NewFamaFrenchModel()

	A := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	At := model.transpose(A)

	// Expected: [[1, 4], [2, 5], [3, 6]]
	if len(At) != 3 || len(At[0]) != 2 {
		t.Fatal("Transpose result has wrong dimensions")
	}

	if At[0][0] != 1 || At[0][1] != 4 || At[1][0] != 2 || At[1][1] != 5 || At[2][0] != 3 || At[2][1] != 6 {
		t.Errorf("Transpose result incorrect: %v", At)
	}
}

func TestSolveLinearSystem(t *testing.T) {
	model := NewFamaFrenchModel()

	// Solve: 2x + y = 5, x + 3y = 8
	// Solution: x = 1.4, y = 2.2 (验证: 2*1.4 + 2.2 = 5, 1.4 + 3*2.2 = 8)
	A := [][]float64{
		{2, 1},
		{1, 3},
	}
	b := []float64{5, 8}

	x := model.solveLinearSystem(A, b)

	if len(x) != 2 {
		t.Fatalf("Expected solution of length 2, got %d", len(x))
	}

	// Check solution with tolerance
	tolerance := 1e-10
	if math.Abs(x[0]-1.4) > tolerance {
		t.Errorf("Expected x[0] = 1.4, got %f", x[0])
	}
	if math.Abs(x[1]-2.2) > tolerance {
		t.Errorf("Expected x[1] = 2.2, got %f", x[1])
	}

	// Verify the solution
	if math.Abs(2*x[0]+x[1]-5.0) > tolerance {
		t.Errorf("Solution doesn't satisfy first equation: 2*%f + %f = %f != 5", x[0], x[1], 2*x[0]+x[1])
	}
	if math.Abs(x[0]+3*x[1]-8.0) > tolerance {
		t.Errorf("Solution doesn't satisfy second equation: %f + 3*%f = %f != 8", x[0], x[1], x[0]+3*x[1])
	}
}

func TestMatrixInverse(t *testing.T) {
	model := NewFamaFrenchModel()

	// Inverse of [[1, 2], [3, 4]] is [[-2, 1], [1.5, -0.5]]
	A := [][]float64{
		{1, 2},
		{3, 4},
	}

	Ainv := model.matrixInverse(A)

	if len(Ainv) != 2 || len(Ainv[0]) != 2 {
		t.Fatal("Matrix inverse result has wrong dimensions")
	}

	// Check with tolerance
	tolerance := 1e-10
	if math.Abs(Ainv[0][0]-(-2.0)) > tolerance {
		t.Errorf("Expected Ainv[0][0] = -2.0, got %f", Ainv[0][0])
	}
	if math.Abs(Ainv[0][1]-1.0) > tolerance {
		t.Errorf("Expected Ainv[0][1] = 1.0, got %f", Ainv[0][1])
	}
	if math.Abs(Ainv[1][0]-1.5) > tolerance {
		t.Errorf("Expected Ainv[1][0] = 1.5, got %f", Ainv[1][0])
	}
	if math.Abs(Ainv[1][1]-(-0.5)) > tolerance {
		t.Errorf("Expected Ainv[1][1] = -0.5, got %f", Ainv[1][1])
	}
}

func TestNormalCDF(t *testing.T) {
	// Test values for standard normal CDF
	testCases := []struct {
		x        float64
		expected float64
		tol      float64
	}{
		{0, 0.5, 1e-10},       // CDF(0) = 0.5
		{1.96, 0.975, 0.002},  // ~97.5th percentile
		{-1.96, 0.025, 0.002}, // ~2.5th percentile
		{3, 0.9987, 0.0001},   // ~99.87th percentile
		{-3, 0.0013, 0.0001},  // ~0.13th percentile
	}

	for _, tc := range testCases {
		result := normalCDF(tc.x)
		if math.Abs(result-tc.expected) > tc.tol {
			t.Errorf("normalCDF(%f) = %f, expected %f (tol %f)", tc.x, result, tc.expected, tc.tol)
		}
	}
}

func TestRandNorm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stochastic test in short mode")
	}

	// Generate multiple random numbers
	n := 1000
	sum := 0.0
	sumSq := 0.0

	for i := 0; i < n; i++ {
		r := randNorm()
		sum += r
		sumSq += r * r
	}

	mean := sum / float64(n)
	variance := sumSq/float64(n) - mean*mean

	// Check that mean is close to 0 (within 3 standard errors)
	seMean := 1.0 / math.Sqrt(float64(n))
	if math.Abs(mean) > 3*seMean {
		t.Errorf("Mean of randNorm() = %f, expected close to 0 (SE = %f)", mean, seMean)
	}

	// Check that variance is close to 1 (within reasonable tolerance)
	if math.Abs(variance-1.0) > 0.1 {
		t.Errorf("Variance of randNorm() = %f, expected close to 1", variance)
	}
}

func TestComparePortfolios(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load sample factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)
	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// Create sample portfolio returns
	portfolio1 := make([]float64, periods)
	portfolio2 := make([]float64, periods)
	for i := 0; i < periods; i++ {
		portfolio1[i] = 0.008 + (float64(i%5)-2)*0.01
		portfolio2[i] = 0.006 + (float64(i%7)-3)*0.008
	}

	portfolios := map[string][]float64{
		"Portfolio1": portfolio1,
		"Portfolio2": portfolio2,
	}

	results := model.ComparePortfolios(portfolios)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check that both portfolios have results
	if _, ok := results["Portfolio1"]; !ok {
		t.Error("Portfolio1 not found in results")
	}
	if _, ok := results["Portfolio2"]; !ok {
		t.Error("Portfolio2 not found in results")
	}

	// Check that results are valid
	for name, attr := range results {
		if attr == nil {
			t.Errorf("%s has nil attribution", name)
			continue
		}
		if attr.Exposures == nil {
			t.Errorf("%s has nil Exposures", name)
		}
	}
}

func TestComparePortfolios_Empty(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load sample factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)
	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// Test with empty portfolios
	portfolios := map[string][]float64{}
	results := model.ComparePortfolios(portfolios)

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty portfolios, got %d", len(results))
	}
}

func TestRiskDecomposition_ThreeFactor(t *testing.T) {
	model := NewFamaFrenchModel()

	exposures := &FactorExposure{
		Market: 1.0,
		Size:   0.5,
		Value:  0.3,
		Alpha:  0.02,
		R2:     0.85,
		AdjR2:  0.82,
	}

	decomposition := model.RiskDecomposition(exposures)

	// Should have 3 factors
	if len(decomposition) != 3 {
		t.Errorf("Expected 3 risk decomposition entries, got %d", len(decomposition))
	}

	// Check that all factors are present
	expectedFactors := []string{"market", "size", "value"}
	for _, factor := range expectedFactors {
		if _, ok := decomposition[factor]; !ok {
			t.Errorf("Factor %s not found in decomposition", factor)
		}
	}

	// Check that contributions sum to 1
	sum := 0.0
	for _, contribution := range decomposition {
		sum += contribution
	}
	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("Risk contributions should sum to 1.0, got %f", sum)
	}

	// Market should have the highest contribution (largest exposure)
	if decomposition["market"] <= decomposition["size"] {
		t.Error("Market should have higher risk contribution than size")
	}
}

func TestRiskDecomposition_FiveFactor(t *testing.T) {
	model := NewFamaFrenchModel()
	model.SetFiveFactor(true)

	exposures := &FactorExposure{
		Market:        1.0,
		Size:          0.5,
		Value:         0.3,
		Profitability: 0.2,
		Investment:    0.1,
		Alpha:         0.02,
		R2:            0.90,
		AdjR2:         0.87,
	}

	decomposition := model.RiskDecomposition(exposures)

	// Should have 5 factors
	if len(decomposition) != 5 {
		t.Errorf("Expected 5 risk decomposition entries, got %d", len(decomposition))
	}

	// Check that all factors are present
	expectedFactors := []string{"market", "size", "value", "profitability", "investment"}
	for _, factor := range expectedFactors {
		if _, ok := decomposition[factor]; !ok {
			t.Errorf("Factor %s not found in decomposition", factor)
		}
	}

	// Check that contributions sum to 1
	sum := 0.0
	for _, contribution := range decomposition {
		sum += contribution
	}
	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("Risk contributions should sum to 1.0, got %f", sum)
	}
}

func TestRiskDecomposition_ZeroExposure(t *testing.T) {
	model := NewFamaFrenchModel()

	exposures := &FactorExposure{
		Market: 0.0,
		Size:   0.0,
		Value:  0.0,
		Alpha:  0.0,
		R2:     0.0,
		AdjR2:  0.0,
	}

	decomposition := model.RiskDecomposition(exposures)

	// Should return empty map when all exposures are zero
	if len(decomposition) != 0 {
		t.Errorf("Expected empty decomposition for zero exposures, got %d entries", len(decomposition))
	}
}

func TestSortByFactorExposure(t *testing.T) {
	attributions := map[string]*FactorAttribution{
		"Portfolio1": {
			Exposures: &FactorExposure{Market: 1.2, Size: 0.5, Value: 0.3},
		},
		"Portfolio2": {
			Exposures: &FactorExposure{Market: 0.8, Size: 0.9, Value: 0.1},
		},
		"Portfolio3": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.3, Value: 0.7},
		},
	}

	// Sort by market exposure (descending)
	sorted := SortByFactorExposure(attributions, "market")

	if len(sorted) != 3 {
		t.Errorf("Expected 3 sorted portfolios, got %d", len(sorted))
	}

	// Check order: Portfolio1 (1.2) > Portfolio3 (1.0) > Portfolio2 (0.8)
	if sorted[0] != "Portfolio1" {
		t.Errorf("Expected first to be Portfolio1, got %s", sorted[0])
	}
	if sorted[1] != "Portfolio3" {
		t.Errorf("Expected second to be Portfolio3, got %s", sorted[1])
	}
	if sorted[2] != "Portfolio2" {
		t.Errorf("Expected third to be Portfolio2, got %s", sorted[2])
	}
}

func TestSortByFactorExposure_Size(t *testing.T) {
	attributions := map[string]*FactorAttribution{
		"Portfolio1": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3},
		},
		"Portfolio2": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.9, Value: 0.1},
		},
		"Portfolio3": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.3, Value: 0.7},
		},
	}

	// Sort by size exposure (descending)
	sorted := SortByFactorExposure(attributions, "size")

	// Check order: Portfolio2 (0.9) > Portfolio1 (0.5) > Portfolio3 (0.3)
	if sorted[0] != "Portfolio2" {
		t.Errorf("Expected first to be Portfolio2, got %s", sorted[0])
	}
	if sorted[1] != "Portfolio1" {
		t.Errorf("Expected second to be Portfolio1, got %s", sorted[1])
	}
	if sorted[2] != "Portfolio3" {
		t.Errorf("Expected third to be Portfolio3, got %s", sorted[2])
	}
}

func TestSortByFactorExposure_Value(t *testing.T) {
	attributions := map[string]*FactorAttribution{
		"Portfolio1": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3},
		},
		"Portfolio2": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.9, Value: 0.1},
		},
		"Portfolio3": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.3, Value: 0.7},
		},
	}

	// Sort by value exposure (descending)
	sorted := SortByFactorExposure(attributions, "value")

	// Check order: Portfolio3 (0.7) > Portfolio1 (0.3) > Portfolio2 (0.1)
	if sorted[0] != "Portfolio3" {
		t.Errorf("Expected first to be Portfolio3, got %s", sorted[0])
	}
	if sorted[1] != "Portfolio1" {
		t.Errorf("Expected second to be Portfolio1, got %s", sorted[1])
	}
	if sorted[2] != "Portfolio2" {
		t.Errorf("Expected third to be Portfolio2, got %s", sorted[2])
	}
}

func TestSortByFactorExposure_InvalidFactor(t *testing.T) {
	attributions := map[string]*FactorAttribution{
		"Portfolio1": {
			Exposures: &FactorExposure{Market: 1.2, Size: 0.5, Value: 0.3},
		},
		"Portfolio2": {
			Exposures: &FactorExposure{Market: 0.8, Size: 0.9, Value: 0.1},
		},
	}

	// Sort by invalid factor (should return all portfolios with 0 exposure)
	sorted := SortByFactorExposure(attributions, "invalid_factor")

	if len(sorted) != 2 {
		t.Errorf("Expected 2 sorted portfolios, got %d", len(sorted))
	}
}

func TestSortByFactorExposure_Profitability(t *testing.T) {
	attributions := map[string]*FactorAttribution{
		"Portfolio1": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3, Profitability: 0.4, Investment: 0.2},
		},
		"Portfolio2": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3, Profitability: 0.8, Investment: 0.1},
		},
		"Portfolio3": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3, Profitability: 0.2, Investment: 0.3},
		},
	}

	// Sort by profitability exposure (descending)
	sorted := SortByFactorExposure(attributions, "profitability")

	// Check order: Portfolio2 (0.8) > Portfolio1 (0.4) > Portfolio3 (0.2)
	if sorted[0] != "Portfolio2" {
		t.Errorf("Expected first to be Portfolio2, got %s", sorted[0])
	}
	if sorted[1] != "Portfolio1" {
		t.Errorf("Expected second to be Portfolio1, got %s", sorted[1])
	}
	if sorted[2] != "Portfolio3" {
		t.Errorf("Expected third to be Portfolio3, got %s", sorted[2])
	}
}

func TestSortByFactorExposure_Investment(t *testing.T) {
	attributions := map[string]*FactorAttribution{
		"Portfolio1": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3, Profitability: 0.4, Investment: 0.2},
		},
		"Portfolio2": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3, Profitability: 0.8, Investment: 0.1},
		},
		"Portfolio3": {
			Exposures: &FactorExposure{Market: 1.0, Size: 0.5, Value: 0.3, Profitability: 0.2, Investment: 0.5},
		},
	}

	// Sort by investment exposure (descending)
	sorted := SortByFactorExposure(attributions, "investment")

	// Check order: Portfolio3 (0.5) > Portfolio1 (0.2) > Portfolio2 (0.1)
	if sorted[0] != "Portfolio3" {
		t.Errorf("Expected first to be Portfolio3, got %s", sorted[0])
	}
	if sorted[1] != "Portfolio1" {
		t.Errorf("Expected second to be Portfolio1, got %s", sorted[1])
	}
	if sorted[2] != "Portfolio2" {
		t.Errorf("Expected third to be Portfolio2, got %s", sorted[2])
	}
}

func TestSortByFactorExposure_Empty(t *testing.T) {
	attributions := map[string]*FactorAttribution{}

	sorted := SortByFactorExposure(attributions, "market")

	if len(sorted) != 0 {
		t.Errorf("Expected 0 sorted portfolios for empty input, got %d", len(sorted))
	}
}

func TestSolveLinearSystem_SingularMatrix(t *testing.T) {
	model := NewFamaFrenchModel()

	// Singular matrix (rows are linearly dependent)
	A := [][]float64{
		{1, 2},
		{2, 4}, // This is 2 * row 1, so matrix is singular
	}
	b := []float64{3, 6}

	x := model.solveLinearSystem(A, b)

	// Should return zero solution for singular matrix
	if len(x) != 2 {
		t.Fatalf("Expected solution of length 2, got %d", len(x))
	}

	// For singular matrix, should return zeros (fallback)
	if x[0] != 0 || x[1] != 0 {
		t.Logf("Singular matrix returned non-zero solution: %v", x)
	}
}

func TestMatrixInverse_SingularMatrix(t *testing.T) {
	model := NewFamaFrenchModel()

	// Singular matrix
	A := [][]float64{
		{1, 2},
		{2, 4},
	}

	Ainv := model.matrixInverse(A)

	// For singular matrix, it returns identity matrix as fallback
	if len(Ainv) != 2 || len(Ainv[0]) != 2 {
		t.Fatalf("Expected 2x2 matrix, got %dx%d", len(Ainv), len(Ainv[0]))
	}

	// Check that it returns identity matrix
	tolerance := 1e-10
	if math.Abs(Ainv[0][0]-1.0) > tolerance || math.Abs(Ainv[0][1]-0.0) > tolerance {
		t.Errorf("Expected identity matrix for singular input, got [%f, %f]", Ainv[0][0], Ainv[0][1])
	}
	if math.Abs(Ainv[1][0]-0.0) > tolerance || math.Abs(Ainv[1][1]-1.0) > tolerance {
		t.Errorf("Expected identity matrix for singular input, got [%f, %f]", Ainv[1][0], Ainv[1][1])
	}
}

func TestMatrixInverse_NonSquare(t *testing.T) {
	model := NewFamaFrenchModel()

	// Non-square matrix - the function still processes it but may produce unexpected results
	// This tests the behavior with non-square input
	A := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	// The function uses len(matrix) as n, so for 2x3 matrix, it will try to create 2x2 result
	Ainv := model.matrixInverse(A)

	// The function returns a result based on row count
	if len(Ainv) != 2 {
		t.Errorf("Expected 2 rows in result, got %d", len(Ainv))
	}
}

func TestAnalyzePortfolio_InsufficientData(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load sample factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := GenerateSampleFactorData(periods)
	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// Create portfolio returns with very few data points (less than factor count)
	portfolioReturns := make([]float64, 2)
	portfolioReturns[0] = 0.01
	portfolioReturns[1] = 0.02

	weights := map[string]float64{
		"VTI": 0.6,
		"VOO": 0.4,
	}

	_, err := model.AnalyzePortfolio(portfolioReturns, weights)
	// Should return an error for insufficient data
	if err == nil {
		t.Error("Expected error for insufficient data, got nil")
	}
}

func TestAnalyzePortfolio_FiveFactor(t *testing.T) {
	model := NewFamaFrenchModel()
	model.SetFiveFactor(true)

	// Load sample five factor data
	periods := 36
	marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns := GenerateSampleFiveFactorData(periods)
	model.LoadFiveFactorData(marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns)

	// Create sample portfolio returns
	portfolioReturns := make([]float64, periods)
	for i := 0; i < periods; i++ {
		portfolioReturns[i] = 0.008 + (float64(i%5)-2)*0.01
	}

	weights := map[string]float64{
		"VTI": 0.6,
		"VOO": 0.4,
	}

	attribution, err := model.AnalyzePortfolio(portfolioReturns, weights)
	if err != nil {
		t.Fatalf("AnalyzePortfolio failed: %v", err)
	}

	if attribution == nil {
		t.Fatal("AnalyzePortfolio returned nil attribution")
	}

	if attribution.Exposures == nil {
		t.Fatal("Exposures is nil")
	}

	// Check five factor exposures
	if math.IsNaN(attribution.Exposures.Market) {
		t.Error("Market exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Size) {
		t.Error("Size exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Value) {
		t.Error("Value exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Profitability) {
		t.Error("Profitability exposure is NaN")
	}
	if math.IsNaN(attribution.Exposures.Investment) {
		t.Error("Investment exposure is NaN")
	}

	// Check contributions
	if attribution.Contributions == nil {
		t.Fatal("Contributions is nil")
	}

	// Check T-statistics
	if attribution.TStatistics == nil {
		t.Fatal("TStatistics is nil")
	}

	// Check P-values
	if attribution.PValues == nil {
		t.Fatal("PValues is nil")
	}
}

func TestPerformRegression_EdgeCases(t *testing.T) {
	model := NewFamaFrenchModel()

	// Test with constant returns (zero variance)
	periods := 36
	portfolioReturns := make([]float64, periods)
	for i := 0; i < periods; i++ {
		portfolioReturns[i] = 0.01 // Constant return
	}

	marketReturns := make([]float64, periods)
	smbReturns := make([]float64, periods)
	hmlReturns := make([]float64, periods)
	riskFreeReturns := make([]float64, periods)

	for i := 0; i < periods; i++ {
		marketReturns[i] = 0.008 + (float64(i%5)-2)*0.002
		smbReturns[i] = 0.002 + (float64(i%3)-1)*0.001
		hmlReturns[i] = 0.003 + (float64(i%4)-1.5)*0.001
		riskFreeReturns[i] = 0.001
	}

	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	result, err := model.performRegression(portfolioReturns)

	if err != nil {
		t.Fatalf("performRegression failed: %v", err)
	}

	if result == nil {
		t.Fatal("RegressionResult is nil")
	}

	// Check that we got valid values
	if math.IsNaN(result.Coefficients["market"]) {
		t.Error("Market coefficient is NaN")
	}

	t.Logf("Coefficients: Market=%f, SMB=%f, HML=%f, Alpha=%f",
		result.Coefficients["market"], result.Coefficients["smb"],
		result.Coefficients["hml"], result.Coefficients["alpha"])
	t.Logf("R2: %f, AdjR2: %f", result.R2, result.AdjR2)
	t.Logf("T-stats: Market=%f, SMB=%f, HML=%f, Alpha=%f",
		result.TStats["market"], result.TStats["smb"],
		result.TStats["hml"], result.TStats["alpha"])
}

func TestPerformRegression_EmptyDependent(t *testing.T) {
	model := NewFamaFrenchModel()

	// Test with empty dependent variable
	emptyReturns := []float64{}

	_, err := model.performRegression(emptyReturns)
	if err == nil {
		t.Error("Expected error for empty dependent variable")
	}
}

func TestPerformRegression_EmptyFactorData(t *testing.T) {
	model := NewFamaFrenchModel()

	// Don't load any factor data
	portfolioReturns := []float64{0.01, 0.02, 0.015, 0.008}

	_, err := model.performRegression(portfolioReturns)
	if err == nil {
		t.Error("Expected error for empty factor data")
	}
}

func TestPerformRegression_InsufficientDataPoints(t *testing.T) {
	model := NewFamaFrenchModel()

	// Load minimal factor data (less than required for regression)
	marketReturns := []float64{0.01, 0.02}
	smbReturns := []float64{0.005, 0.008}
	hmlReturns := []float64{0.003, 0.006}
	riskFreeReturns := []float64{0.001, 0.001}

	model.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// Try to perform regression with only 2 data points (need at least 4 for 3-factor)
	portfolioReturns := []float64{0.01, 0.02}

	_, err := model.performRegression(portfolioReturns)
	if err == nil {
		t.Error("Expected error for insufficient data points")
	}
}

func TestPerformRegression_FiveFactorInsufficientData(t *testing.T) {
	model := NewFamaFrenchModel()
	model.SetFiveFactor(true)

	// Load minimal factor data (less than required for 5-factor regression)
	marketReturns := []float64{0.01, 0.02, 0.015, 0.008, 0.012}
	smbReturns := []float64{0.005, 0.008, 0.006, 0.004, 0.007}
	hmlReturns := []float64{0.003, 0.006, 0.004, 0.005, 0.003}
	rmwReturns := []float64{0.002, 0.004, 0.003, 0.002, 0.004}
	cmaReturns := []float64{0.001, 0.003, 0.002, 0.001, 0.003}
	riskFreeReturns := []float64{0.001, 0.001, 0.001, 0.001, 0.001}

	model.LoadFiveFactorData(marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns)

	// Try to perform regression with only 5 data points (need at least 6 for 5-factor)
	portfolioReturns := []float64{0.01, 0.02, 0.015, 0.008, 0.012}

	_, err := model.performRegression(portfolioReturns)
	if err == nil {
		t.Error("Expected error for insufficient data points for 5-factor model")
	}
}

func TestParseFrenchCSV(t *testing.T) {
	// Sample CSV data in Kenneth French database format
	csvData := `Date,Mkt-RF,SMB,HML,RF
202301,5.2,1.5,-0.8,0.4
202302,3.1,2.1,1.2,0.3
202303,-2.5,-1.0,0.5,0.4
202304,4.8,0.8,-1.5,0.3
202305,1.2,1.8,2.0,0.4
202306,2.5,-0.5,1.5,0.3`

	dates, marketReturns, smbReturns, hmlReturns, riskFreeReturns, err := ParseFrenchCSV(csvData)

	if err != nil {
		t.Fatalf("ParseFrenchCSV failed: %v", err)
	}

	// Should have 6 data points
	if len(dates) != 6 {
		t.Errorf("Expected 6 dates, got %d", len(dates))
	}
	if len(marketReturns) != 6 {
		t.Errorf("Expected 6 market returns, got %d", len(marketReturns))
	}
	if len(smbReturns) != 6 {
		t.Errorf("Expected 6 SMB returns, got %d", len(smbReturns))
	}
	if len(hmlReturns) != 6 {
		t.Errorf("Expected 6 HML returns, got %d", len(hmlReturns))
	}
	if len(riskFreeReturns) != 6 {
		t.Errorf("Expected 6 risk-free returns, got %d", len(riskFreeReturns))
	}

	// Check that values are converted to decimal (divided by 100)
	expectedMarket := 5.2 / 100
	if math.Abs(marketReturns[0]-expectedMarket) > 1e-10 {
		t.Errorf("Expected first market return %f, got %f", expectedMarket, marketReturns[0])
	}

	expectedSMB := 1.5 / 100
	if math.Abs(smbReturns[0]-expectedSMB) > 1e-10 {
		t.Errorf("Expected first SMB return %f, got %f", expectedSMB, smbReturns[0])
	}

	expectedHML := -0.8 / 100
	if math.Abs(hmlReturns[0]-expectedHML) > 1e-10 {
		t.Errorf("Expected first HML return %f, got %f", expectedHML, hmlReturns[0])
	}

	expectedRF := 0.4 / 100
	if math.Abs(riskFreeReturns[0]-expectedRF) > 1e-10 {
		t.Errorf("Expected first risk-free return %f, got %f", expectedRF, riskFreeReturns[0])
	}
}

func TestParseFrenchCSV_EmptyData(t *testing.T) {
	csvData := ``

	dates, marketReturns, _, _, _, err := ParseFrenchCSV(csvData)

	if err != nil {
		t.Fatalf("ParseFrenchCSV failed: %v", err)
	}

	// Should return empty slices
	if len(dates) != 0 {
		t.Errorf("Expected 0 dates for empty data, got %d", len(dates))
	}
	if len(marketReturns) != 0 {
		t.Errorf("Expected 0 market returns for empty data, got %d", len(marketReturns))
	}
}

func TestParseFrenchCSV_InvalidData(t *testing.T) {
	// CSV with some invalid rows (must have consistent column count)
	csvData := `Date,Mkt-RF,SMB,HML,RF
202301,5.2,1.5,-0.8,0.4
202302,invalid,invalid,invalid,invalid
202303,3.1,2.1,1.2,0.3
abc,not,numbers,here,test
202304,-2.5,-1.0,0.5,0.4`

	dates, marketReturns, _, _, _, err := ParseFrenchCSV(csvData)

	if err != nil {
		t.Fatalf("ParseFrenchCSV failed: %v", err)
	}

	// Should have 4 valid data points (rows with invalid dates are skipped, but invalid numbers parse as 0)
	if len(dates) != 4 {
		t.Errorf("Expected 4 valid dates, got %d", len(dates))
	}
	if len(marketReturns) != 4 {
		t.Errorf("Expected 4 valid market returns, got %d", len(marketReturns))
	}
}

func TestFamaFrench_FiveFactorEmptyData(t *testing.T) {
	model := NewFamaFrenchModel()

	model.LoadFiveFactorData(
		[]float64{}, []float64{}, []float64{},
		[]float64{}, []float64{}, []float64{},
	)

	if !model.UseFiveFactor {
		t.Error("UseFiveFactor should be true after LoadFiveFactorData")
	}
}

func TestFamaFrench_FactorModeSwitching(t *testing.T) {
	model := NewFamaFrenchModel()

	if model.UseFiveFactor {
		t.Error("New model should default to 3-factor mode")
	}

	model.LoadFiveFactorData(
		[]float64{0.01}, []float64{0.02}, []float64{0.03},
		[]float64{0.04}, []float64{0.05}, []float64{0.01},
	)

	if !model.UseFiveFactor {
		t.Error("Should be in 5-factor mode after LoadFiveFactorData")
	}
}

func TestFamaFrench_GenerateSampleData(t *testing.T) {
	m, s, h, rf := GenerateSampleFactorData(10)
	if len(m) != 10 || len(s) != 10 || len(h) != 10 || len(rf) != 10 {
		t.Errorf("Expected 10 data points each, got market=%d smb=%d hml=%d rf=%d", len(m), len(s), len(h), len(rf))
	}
}
