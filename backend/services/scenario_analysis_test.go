package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScenarioAnalysisService(t *testing.T) {
	service := NewScenarioAnalysisService()
	assert.NotNil(t, service)
	assert.NotNil(t, service.historicalData)
}

func TestValidatePortfolio(t *testing.T) {
	service := NewScenarioAnalysisService()

	t.Run("valid portfolio", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.7,
			"JEPQ": 0.3,
		}
		err := service.validatePortfolio(portfolio)
		assert.NoError(t, err)
	})

	t.Run("empty portfolio", func(t *testing.T) {
		portfolio := map[string]float64{}
		err := service.validatePortfolio(portfolio)
		assert.Error(t, err)
	})

	t.Run("invalid weight - zero", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0,
			"JEPQ": 1,
		}
		err := service.validatePortfolio(portfolio)
		assert.Error(t, err)
	})

	t.Run("invalid weight - greater than 1", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 1.5,
			"JEPQ": 0.3,
		}
		err := service.validatePortfolio(portfolio)
		assert.Error(t, err)
	})

	t.Run("weights not sum to 1", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.5,
			"JEPQ": 0.3,
		}
		err := service.validatePortfolio(portfolio)
		assert.Error(t, err)
	})
}

func TestCalculateWeightedMetrics(t *testing.T) {
	service := NewScenarioAnalysisService()

	t.Run("calculate metrics for SCHD/JEPQ portfolio", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.7,
			"JEPQ": 0.3,
		}
		metrics := service.calculateWeightedMetrics(portfolio)
		assert.NotNil(t, metrics)
		assert.Greater(t, metrics.AnnualReturn, 0.0)
		assert.Greater(t, metrics.Volatility, 0.0)
		assert.Greater(t, metrics.DividendYield, 0.0)
	})
}

func TestAdjustForPessimistic(t *testing.T) {
	service := NewScenarioAnalysisService()

	metrics := &WeightedMetrics{
		AnnualReturn:  0.12,
		Volatility:    0.18,
		SharpeRatio:   0.67,
		MaxDrawdown:   0.20,
		DividendYield: 0.05,
	}

	pessimistic := service.adjustForPessimistic(metrics)

	assert.Less(t, pessimistic.AnnualReturn, metrics.AnnualReturn)
	assert.Greater(t, pessimistic.Volatility, metrics.Volatility)
	assert.Less(t, pessimistic.SharpeRatio, metrics.SharpeRatio)
	assert.Greater(t, pessimistic.MaxDrawdown, metrics.MaxDrawdown)
}

func TestAdjustForOptimistic(t *testing.T) {
	service := NewScenarioAnalysisService()

	metrics := &WeightedMetrics{
		AnnualReturn:  0.12,
		Volatility:    0.18,
		SharpeRatio:   0.67,
		MaxDrawdown:   0.20,
		DividendYield: 0.05,
	}

	optimistic := service.adjustForOptimistic(metrics)

	assert.Greater(t, optimistic.AnnualReturn, metrics.AnnualReturn)
	assert.Less(t, optimistic.Volatility, metrics.Volatility)
	assert.Greater(t, optimistic.SharpeRatio, metrics.SharpeRatio)
	assert.Less(t, optimistic.MaxDrawdown, metrics.MaxDrawdown)
}

func TestGenerateScenario(t *testing.T) {
	service := NewScenarioAnalysisService()

	metrics := &WeightedMetrics{
		AnnualReturn:  0.12,
		Volatility:    0.18,
		SharpeRatio:   0.67,
		MaxDrawdown:   0.20,
		DividendYield: 0.05,
	}

	t.Run("generate neutral scenario", func(t *testing.T) {
		result := service.generateScenario(Neutral, metrics, 100000, 5)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Projections))
		assert.Greater(t, result.FinalValue, 100000.0)
	})

	t.Run("generate pessimistic scenario", func(t *testing.T) {
		pessimistic := service.adjustForPessimistic(metrics)
		result := service.generateScenario(Pessimistic, pessimistic, 100000, 5)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Projections))
	})

	t.Run("generate optimistic scenario", func(t *testing.T) {
		optimistic := service.adjustForOptimistic(metrics)
		result := service.generateScenario(Optimistic, optimistic, 100000, 5)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Projections))
	})
}

func TestAnalyzePortfolio(t *testing.T) {
	service := NewScenarioAnalysisService()

	t.Run("analyze default portfolio", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.7,
			"JEPQ": 0.3,
		}
		result, err := service.AnalyzePortfolio(portfolio, 100000, 10)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 10, result.TimeHorizonYears)
		assert.Equal(t, 100000.0, result.TotalInvestment)
		assert.NotNil(t, result.Scenarios[Neutral])
		assert.NotNil(t, result.Scenarios[Pessimistic])
		assert.NotNil(t, result.Scenarios[Optimistic])
		assert.NotNil(t, result.ComparisonMetrics)
		assert.NotEmpty(t, result.Methodology)
		assert.NotEmpty(t, result.Assumptions)
		assert.NotEmpty(t, result.Limitations)
	})

	t.Run("analyze with invalid portfolio", func(t *testing.T) {
		portfolio := map[string]float64{}
		result, err := service.AnalyzePortfolio(portfolio, 100000, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCalculateComparison(t *testing.T) {
	service := NewScenarioAnalysisService()

	metrics := &WeightedMetrics{
		AnnualReturn:  0.12,
		Volatility:    0.18,
		SharpeRatio:   0.67,
		MaxDrawdown:   0.20,
		DividendYield: 0.05,
	}

	scenarios := make(map[MarketScenario]*ScenarioResult)
	scenarios[Neutral] = service.generateScenario(Neutral, metrics, 100000, 5)
	scenarios[Pessimistic] = service.generateScenario(Pessimistic, service.adjustForPessimistic(metrics), 100000, 5)
	scenarios[Optimistic] = service.generateScenario(Optimistic, service.adjustForOptimistic(metrics), 100000, 5)

	comparison := service.calculateComparison(scenarios)

	assert.NotNil(t, comparison)
	assert.NotEmpty(t, comparison.BestScenario)
	assert.NotEmpty(t, comparison.WorstScenario)
	assert.Greater(t, comparison.ValueDifference, 0.0)
}
