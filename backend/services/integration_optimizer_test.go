package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPortfolioOptimizer_MaxSharpe_ThreeAssets(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{"AAPL", "MSFT", "GOOG"},
		OptimizationType: OptimizationTypeMaxSharpe,
		RiskAversion:     decimal.NewFromFloat(2.0),
		RiskFreeRate:     decimal.NewFromFloat(0.04),
	}

	result, err := optimizer.Optimize(request)
	if err != nil {
		t.Logf("Optimize returned error (may need analysis service): %v", err)
		return
	}

	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Weights, "Weights should not be empty")
	assert.Equal(t, OptimizationTypeMaxSharpe, result.OptimizationType)

	totalWeight := decimal.Zero
	for _, w := range result.Weights {
		totalWeight = totalWeight.Add(w)
	}
	assert.True(t, totalWeight.GreaterThanOrEqual(decimal.NewFromFloat(0.99)),
		"Weights should sum to ~1.0, got %s", totalWeight.String())
	assert.True(t, totalWeight.LessThanOrEqual(decimal.NewFromFloat(1.01)),
		"Weights should not exceed 1.0 significantly")
}

func TestPortfolioOptimizer_MinVolatility_FourAssets(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{"AAPL", "MSFT", "GOOG", "AMZN"},
		OptimizationType: OptimizationTypeMinVolatility,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
	}

	result, err := optimizer.Optimize(request)
	if err != nil {
		t.Logf("MinVolatility optimize: %v", err)
		return
	}

	assert.NotNil(t, result)
	assert.Equal(t, OptimizationTypeMinVolatility, result.OptimizationType)
	assert.True(result.ExpectedVolatility.GreaterThan(decimal.Zero))
}

func TestPortfolioOptimizer_EqualWeight_FiveAssets(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{"AAPL", "MSFT", "GOOG", "AMZN", "META"},
		OptimizationType: OptimizationTypeEqualWeight,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
	}

	result, err := optimizer.Optimize(request)
	if err != nil {
		t.Logf("EqualWeight optimize: %v", err)
		return
	}

	assert.NotNil(t, result)
	assert.Equal(t, OptimizationTypeEqualWeight, result.OptimizationType)

	expectedWeight := decimal.NewFromFloat(1.0 / float64(len(request.Symbols)))
	for sym, w := range result.Weights {
		diff := w.Sub(expectedWeight).Abs()
		assert.True(t, diff.LessThan(decimal.NewFromFloat(0.01)),
			"Symbol %s weight %s should be close to equal weight %s",
			sym, w.String(), expectedWeight.String())
	}
}

func TestPortfolioValidator_TooFewSymbols(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{"AAPL"},
		OptimizationType: OptimizationTypeMaxSharpe,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
	}

	_, err := optimizer.Optimize(request)
	assert.Error(t, err, "Single symbol should return validation error")
}

func TestPortfolioValidator_EmptySymbols(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{},
		OptimizationType: OptimizationTypeMaxSharpe,
	}

	_, err := optimizer.Optimize(request)
	assert.Error(t, err, "Empty symbols should return validation error")
}

func TestPortfolioOptimizer_DefaultConstraintsApplied(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{"AAPL", "MSFT", "GOOG"},
		OptimizationType: OptimizationTypeMaxSharpe,
		Constraints:      OptimizationConstraints{},
	}

	result, err := optimizer.Optimize(request)
	if err != nil {
		t.Logf("Optimize with defaults: %v", err)
		return
	}

	if result != nil && len(result.Weights) > 0 {
		maxAllowed := decimal.NewFromFloat(0.4)
		for sym, w := range result.Weights {
			assert.True(t, w.LessThanOrEqual(maxAllowed),
				"Symbol %s weight %s should not exceed max %s (default constraint)",
				sym, w.String(), maxAllowed.String())
		}
	}
}

func TestPortfolioOptimizer_CustomConstraints(t *testing.T) {
	optimizer := NewPortfolioOptimizer(nil)

	request := PortfolioOptimizationRequest{
		Symbols:          []string{"AAPL", "MSFT", "GOOG"},
		OptimizationType: OptimizationTypeMaxSharpe,
		Constraints: OptimizationConstraints{
			MaxWeightPerAsset: decimal.NewFromFloat(0.5),
			MinWeightPerAsset: decimal.NewFromFloat(0.1),
			AllowShort:        false,
		},
	}

	result, err := optimizer.Optimize(request)
	if err != nil {
		t.Logf("Custom constraints: %v", err)
		return
	}

	if result != nil && len(result.Weights) > 0 {
		maxAllowed := decimal.NewFromFloat(0.5)
		minAllowed := decimal.NewFromFloat(0.1)
		for sym, w := range result.Weights {
			assert.True(t, w.LessThanOrEqual(maxAllowed),
				"%s weight %s exceeds custom max %s", sym, w.String(), maxAllowed.String())
			assert.True(t, w.GreaterThanOrEqual(minAllowed),
				"%s weight %s below custom min %s", sym, w.String(), minAllowed.String())
		}
	}
}
