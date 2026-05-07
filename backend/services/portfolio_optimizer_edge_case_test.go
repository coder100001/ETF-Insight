package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PortfolioOptimizerEdgeCaseTestSuite struct {
	suite.Suite
	optimizer *PortfolioOptimizer
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) SetupTest() {
	analysisService := NewETFAnalysisService(nil)
	s.optimizer = NewPortfolioOptimizer(analysisService)
}

func TestPortfolioOptimizerEdgeCaseSuite(t *testing.T) {
	suite.Run(t, new(PortfolioOptimizerEdgeCaseTestSuite))
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) TestCalculatePortfolioReturn_EmptyWeights() {
	weights := map[string]decimal.Decimal{}
	meanReturns := map[string]decimal.Decimal{
		"SPY": decimal.NewFromFloat(0.10),
	}

	result := s.optimizer.calculatePortfolioReturn(weights, meanReturns)

	s.NotNil(result)
	s.True(result.IsZero(), "Empty weights should return zero return")
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) TestCalculatePortfolioVolatility_EmptyWeights() {
	weights := map[string]decimal.Decimal{}
	covMatrix := map[string]map[string]decimal.Decimal{
		"SPY": {"SPY": decimal.NewFromFloat(0.04)},
	}

	result := s.optimizer.calculatePortfolioVolatility(weights, covMatrix)

	s.NotNil(result)
	s.True(result.IsZero(), "Empty weights should return zero volatility")
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) TestCalculateSharpeRatio_ZeroVolatility() {
	expectedReturn := decimal.NewFromFloat(0.10)
	volatility := decimal.Zero
	riskFreeRate := decimal.NewFromFloat(0.04)

	result := s.optimizer.calculateSharpeRatio(expectedReturn, volatility, riskFreeRate)

	s.NotNil(result)
	s.True(result.IsZero(), "Zero volatility should return zero Sharpe ratio")
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) TestCalculateSharpeRatio_NegativeVolatility() {
	expectedReturn := decimal.NewFromFloat(0.10)
	volatility := decimal.NewFromFloat(-0.01)
	riskFreeRate := decimal.NewFromFloat(0.04)

	result := s.optimizer.calculateSharpeRatio(expectedReturn, volatility, riskFreeRate)

	s.NotNil(result)
	s.True(result.IsZero(), "Negative volatility should return zero Sharpe ratio")
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) TestMaximizeSharpeRatio_ZeroWeightSymbols() {
	request := PortfolioOptimizationRequest{
		Symbols:          []string{},
		OptimizationType: OptimizationTypeMaxSharpe,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
	}

	s.NotPanics(func() {
		_, err := s.optimizer.Optimize(request)
		assert.Error(s.T(), err, "Empty symbols should return error")
	})
}

func (s *PortfolioOptimizerEdgeCaseTestSuite) TestMaximizeSharpeRatio_SingleSymbol() {
	request := PortfolioOptimizationRequest{
		Symbols:          []string{"SPY"},
		OptimizationType: OptimizationTypeMaxSharpe,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
	}

	s.NotPanics(func() {
		result, err := s.optimizer.Optimize(request)
		if err == nil && result != nil {
			s.NotNil(result.Weights)
			if weight, ok := result.Weights["SPY"]; ok {
				s.Equal(decimal.NewFromFloat(1.0), weight, "Single symbol should have weight 1.0")
			}
		}
	})
}
