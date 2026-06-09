package services

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewScenarioAnalysisService(t *testing.T) {
	service := NewScenarioAnalysisService()
	assert.NotNil(t, service)
	assert.NotNil(t, service.analyticsService)
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

func TestGetDefaultPortfolioAnalytics(t *testing.T) {
	service := NewScenarioAnalysisService()

	t.Run("default analytics for SCHD/JEPQ portfolio", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.7,
			"JEPQ": 0.3,
		}
		analytics := service.getDefaultPortfolioAnalytics(portfolio)
		assert.NotNil(t, analytics)
		assert.Greater(t, analytics.ExpectedReturn, 0.0)
		assert.Greater(t, analytics.Volatility, 0.0)
		assert.NotNil(t, analytics.ETFMetrics)
	})

	t.Run("default analytics includes known ETFs", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.5,
			"JEPQ": 0.3,
			"QQQ":  0.2,
		}
		analytics := service.getDefaultPortfolioAnalytics(portfolio)
		assert.NotNil(t, analytics)
		assert.NotNil(t, analytics.ETFMetrics["SCHD"])
		assert.NotNil(t, analytics.ETFMetrics["JEPQ"])
		assert.NotNil(t, analytics.ETFMetrics["QQQ"])
	})
}

func TestAdjustAnalyticsForScenario(t *testing.T) {
	service := NewScenarioAnalysisService()

	baseAnalytics := &PortfolioAnalytics{
		ExpectedReturn: 0.12,
		Volatility:     0.18,
		SharpeRatio:    0.67,
		MaxDrawdown:    0.20,
	}

	t.Run("adjust for pessimistic scenario", func(t *testing.T) {
		adjusted := service.adjustAnalyticsForScenario(baseAnalytics, Pessimistic)
		assert.Less(t, adjusted.ExpectedReturn, baseAnalytics.ExpectedReturn)
		assert.Greater(t, adjusted.Volatility, baseAnalytics.Volatility)
		assert.Less(t, adjusted.SharpeRatio, baseAnalytics.SharpeRatio)
		assert.Greater(t, adjusted.MaxDrawdown, baseAnalytics.MaxDrawdown)
	})

	t.Run("adjust for optimistic scenario", func(t *testing.T) {
		adjusted := service.adjustAnalyticsForScenario(baseAnalytics, Optimistic)
		assert.Greater(t, adjusted.ExpectedReturn, baseAnalytics.ExpectedReturn)
		assert.Less(t, adjusted.Volatility, baseAnalytics.Volatility)
		assert.Greater(t, adjusted.SharpeRatio, baseAnalytics.SharpeRatio)
		assert.Less(t, adjusted.MaxDrawdown, baseAnalytics.MaxDrawdown)
	})

	t.Run("neutral scenario keeps base values", func(t *testing.T) {
		adjusted := service.adjustAnalyticsForScenario(baseAnalytics, Neutral)
		assert.Equal(t, baseAnalytics.ExpectedReturn, adjusted.ExpectedReturn)
		assert.Equal(t, baseAnalytics.Volatility, adjusted.Volatility)
		assert.Equal(t, baseAnalytics.SharpeRatio, adjusted.SharpeRatio)
		assert.Equal(t, baseAnalytics.MaxDrawdown, adjusted.MaxDrawdown)
	})
}

func TestGenerateScenario(t *testing.T) {
	service := NewScenarioAnalysisService()

	analytics := &PortfolioAnalytics{
		ExpectedReturn: 0.12,
		Volatility:     0.18,
		SharpeRatio:    0.67,
		MaxDrawdown:    0.20,
	}

	t.Run("generate neutral scenario", func(t *testing.T) {
		result := service.generateScenario(Neutral, analytics, 100000, 5, 1.0)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Projections))
		assert.NotNil(t, result.Assumptions)
		assert.Greater(t, result.FinalValue, 0.0)
	})

	t.Run("generate pessimistic scenario", func(t *testing.T) {
		pessimisticAnalytics := service.adjustAnalyticsForScenario(analytics, Pessimistic)
		result := service.generateScenario(Pessimistic, pessimisticAnalytics, 100000, 5, 0.85)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Projections))
		assert.Equal(t, Pessimistic, result.Assumptions.Scenario)
	})

	t.Run("generate optimistic scenario", func(t *testing.T) {
		optimisticAnalytics := service.adjustAnalyticsForScenario(analytics, Optimistic)
		result := service.generateScenario(Optimistic, optimisticAnalytics, 100000, 5, 1.15)
		assert.NotNil(t, result)
		assert.Equal(t, 5, len(result.Projections))
		assert.Equal(t, Optimistic, result.Assumptions.Scenario)
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

	t.Run("analyze with invalid weights", func(t *testing.T) {
		portfolio := map[string]float64{
			"SCHD": 0.5,
			"JEPQ": 0.3,
		}
		result, err := service.AnalyzePortfolio(portfolio, 100000, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCalculateComparison(t *testing.T) {
	service := NewScenarioAnalysisService()

	analytics := &PortfolioAnalytics{
		ExpectedReturn: 0.12,
		Volatility:     0.18,
		SharpeRatio:    0.67,
		MaxDrawdown:    0.20,
	}

	scenarios := make(map[MarketScenario]*ScenarioResult)
	scenarios[Neutral] = service.generateScenario(Neutral, analytics, 100000, 5, 1.0)
	scenarios[Pessimistic] = service.generateScenario(Pessimistic, service.adjustAnalyticsForScenario(analytics, Pessimistic), 100000, 5, 0.85)
	scenarios[Optimistic] = service.generateScenario(Optimistic, service.adjustAnalyticsForScenario(analytics, Optimistic), 100000, 5, 1.15)

	comparison := service.calculateComparison(scenarios)

	assert.NotNil(t, comparison)
	assert.NotEmpty(t, comparison.BestScenario)
	assert.NotEmpty(t, comparison.WorstScenario)
	assert.Greater(t, comparison.ValueDifference, 0.0)
	assert.NotEmpty(t, comparison.RiskAdjustedWinner)
}

func TestGenerateNormalRandom(t *testing.T) {
	service := NewScenarioAnalysisService()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成大量随机数并检查统计特性
	n := 10000
	sum := 0.0
	sumSq := 0.0

	for i := 0; i < n; i++ {
		z := service.generateNormalRandom(r)
		sum += z
		sumSq += z * z
	}

	mean := sum / float64(n)
	variance := sumSq/float64(n) - mean*mean

	// 均值应该接近0
	assert.InDelta(t, 0, mean, 0.1)
	// 方差应该接近1
	assert.InDelta(t, 1, variance, 0.2)
}

func TestCalculateCVaRFromSimulations(t *testing.T) {
	service := NewScenarioAnalysisService()

	// 生成模拟结果
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i) // 0, 1, 2, ..., 999
	}

	// 计算 95% CVaR
	cvar95 := service.calculateCVaRFromSimulations(values, 0.95)
	// 应该是最低 5% 的平均值 (0-49)
	assert.Less(t, cvar95, 50.0)

	// 计算 99% CVaR
	cvar99 := service.calculateCVaRFromSimulations(values, 0.99)
	// 应该是最低 1% 的平均值 (0-9)
	assert.Less(t, cvar99, 10.0)

	// 99% CVaR 应该比 95% CVaR 更小
	assert.Less(t, cvar99, cvar95)
}

func TestCalculateCVaRFromSimulations_SingleValue(t *testing.T) {
	service := NewScenarioAnalysisService()

	values := []float64{100}
	cvar := service.calculateCVaRFromSimulations(values, 0.95)
	assert.Equal(t, 100.0, cvar)
}
