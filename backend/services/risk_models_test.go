package services

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewRiskModels(t *testing.T) {
	rm := NewRiskModels()
	assert.NotNil(t, rm)
}

func TestCalculateHistoricalVaR(t *testing.T) {
	rm := NewRiskModels()

	// 生成测试收益率数据（模拟一些负收益）
	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(-0.02),
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(-0.03),
		decimal.NewFromFloat(0.005),
		decimal.NewFromFloat(-0.025),
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(-0.015),
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(-0.035),
		decimal.NewFromFloat(0.025),
		decimal.NewFromFloat(-0.01),
		decimal.NewFromFloat(0.008),
		decimal.NewFromFloat(-0.028),
		decimal.NewFromFloat(0.012),
		decimal.NewFromFloat(-0.018),
		decimal.NewFromFloat(0.022),
		decimal.NewFromFloat(-0.008),
		decimal.NewFromFloat(0.005),
		decimal.NewFromFloat(-0.032),
	}

	result, err := rm.CalculateHistoricalVaR(returns, 0.95)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.VaR.LessThanOrEqual(decimal.Zero))
	assert.True(t, result.CVaR.LessThanOrEqual(decimal.Zero))
	assert.Equal(t, "historical", result.Method)
	assert.Equal(t, 20, result.PeriodDays)
}

func TestCalculateHistoricalVaR_InsufficientData(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{}

	result, err := rm.CalculateHistoricalVaR(returns, 0.95)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidReturns, err)
	assert.Nil(t, result)
}

func TestCalculateHistoricalVaR_InvalidConfidence(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(-0.02),
	}

	result, err := rm.CalculateHistoricalVaR(returns, 1.5)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidConfidenceLevel, err)
	assert.Nil(t, result)
}

func TestCalculateParametricVaR(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(-0.02),
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(-0.03),
		decimal.NewFromFloat(0.005),
		decimal.NewFromFloat(-0.025),
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(-0.015),
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(-0.035),
		decimal.NewFromFloat(0.025),
		decimal.NewFromFloat(-0.01),
		decimal.NewFromFloat(0.008),
		decimal.NewFromFloat(-0.028),
		decimal.NewFromFloat(0.012),
		decimal.NewFromFloat(-0.018),
		decimal.NewFromFloat(0.022),
		decimal.NewFromFloat(-0.008),
		decimal.NewFromFloat(0.005),
		decimal.NewFromFloat(-0.032),
	}

	result, err := rm.CalculateParametricVaR(returns, 0.95)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.VaR.GreaterThan(decimal.Zero))
	assert.Equal(t, "parametric", result.Method)
}

func TestCalculatePortfolioVaR(t *testing.T) {
	rm := NewRiskModels()

	weights := map[string]decimal.Decimal{
		"SPY": decimal.NewFromFloat(0.6),
		"BND": decimal.NewFromFloat(0.4),
	}

	returns := map[string][]decimal.Decimal{
		"SPY": {
			decimal.NewFromFloat(0.02),
			decimal.NewFromFloat(-0.01),
			decimal.NewFromFloat(0.015),
			decimal.NewFromFloat(-0.025),
			decimal.NewFromFloat(0.01),
		},
		"BND": {
			decimal.NewFromFloat(0.005),
			decimal.NewFromFloat(0.002),
			decimal.NewFromFloat(-0.003),
			decimal.NewFromFloat(0.004),
			decimal.NewFromFloat(0.001),
		},
	}

	result, err := rm.CalculatePortfolioVaR(weights, returns, 0.95)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// PortfolioVaR 可以是负数（表示损失）或零
	assert.NotNil(t, result.PortfolioVaR)
	assert.NotNil(t, result.ComponentVaR)
	assert.NotNil(t, result.MarginalVaR)
}

func TestCalculateRiskMetrics(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.001),
		decimal.NewFromFloat(-0.002),
		decimal.NewFromFloat(0.0015),
		decimal.NewFromFloat(-0.001),
		decimal.NewFromFloat(0.002),
		decimal.NewFromFloat(-0.0015),
		decimal.NewFromFloat(0.001),
		decimal.NewFromFloat(-0.0005),
		decimal.NewFromFloat(0.0025),
		decimal.NewFromFloat(-0.002),
	}

	riskFreeRate := decimal.NewFromFloat(0.0001) // 日无风险利率约 0.01%

	result, err := rm.CalculateRiskMetrics(returns, riskFreeRate, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Volatility.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, result.MaxDrawdown.GreaterThanOrEqual(decimal.Zero))
}

func TestCalculateRiskMetrics_WithBenchmark(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.001),
		decimal.NewFromFloat(-0.002),
		decimal.NewFromFloat(0.0015),
		decimal.NewFromFloat(-0.001),
		decimal.NewFromFloat(0.002),
	}

	benchmarkReturns := []decimal.Decimal{
		decimal.NewFromFloat(0.0008),
		decimal.NewFromFloat(-0.0015),
		decimal.NewFromFloat(0.0012),
		decimal.NewFromFloat(-0.0008),
		decimal.NewFromFloat(0.0018),
	}

	riskFreeRate := decimal.NewFromFloat(0.0001)

	result, err := rm.CalculateRiskMetrics(returns, riskFreeRate, benchmarkReturns)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Beta 和 Alpha 应该被计算
	assert.True(t, result.Beta.GreaterThan(decimal.Zero) || result.Beta.Equal(decimal.Zero))
}

func TestCalculateRiskMetrics_InsufficientData(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{}
	riskFreeRate := decimal.NewFromFloat(0.0001)

	result, err := rm.CalculateRiskMetrics(returns, riskFreeRate, nil)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidReturns, err)
	assert.Nil(t, result)
}

func TestGetZScore(t *testing.T) {
	// 测试已知置信水平
	assert.InDelta(t, 1.645, getZScore(0.95), 0.001)
	assert.InDelta(t, 2.326, getZScore(0.99), 0.001)
	assert.InDelta(t, 1.282, getZScore(0.90), 0.001)

	// 测试插值
	z95 := getZScore(0.95)
	z99 := getZScore(0.99)
	assert.True(t, z95 < z99)
}

func TestCalculateMean(t *testing.T) {
	values := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(20),
		decimal.NewFromFloat(30),
	}

	mean := calculateMean(values)
	assert.InDelta(t, 20.0, mean.InexactFloat64(), 0.0001)
}

func TestCalculateMean_Empty(t *testing.T) {
	values := []decimal.Decimal{}
	mean := calculateMean(values)
	assert.Equal(t, decimal.Zero, mean)
}

func TestCalculateStdDev(t *testing.T) {
	values := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(20),
		decimal.NewFromFloat(30),
	}
	mean := decimal.NewFromFloat(20)

	stdDev := calculateStdDev(values, mean)
	assert.True(t, stdDev.GreaterThan(decimal.Zero))
}

func TestCalculateMaxDrawdownFromReturns(t *testing.T) {
	// 模拟一个先涨后跌的收益率序列
	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.05),  // 涨 5%
		decimal.NewFromFloat(0.03),  // 涨 3%
		decimal.NewFromFloat(-0.02), // 跌 2%
		decimal.NewFromFloat(-0.04), // 跌 4%
		decimal.NewFromFloat(0.01),  // 涨 1%
	}

	maxDrawdown := calculateMaxDrawdownFromReturns(returns)
	assert.True(t, maxDrawdown.GreaterThan(decimal.Zero))
}

func TestCalculateMaxDrawdownFromReturns_NoDrawdown(t *testing.T) {
	// 一直上涨的序列
	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(0.01),
	}

	maxDrawdown := calculateMaxDrawdownFromReturns(returns)
	assert.Equal(t, decimal.Zero, maxDrawdown)
}

func TestCalculateMaxDrawdownFromReturns_Empty(t *testing.T) {
	returns := []decimal.Decimal{}
	maxDrawdown := calculateMaxDrawdownFromReturns(returns)
	assert.Equal(t, decimal.Zero, maxDrawdown)
}

func TestCalculateBetaAlpha(t *testing.T) {
	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(-0.01),
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(-0.005),
		decimal.NewFromFloat(0.01),
	}

	benchmarkReturns := []decimal.Decimal{
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(-0.008),
		decimal.NewFromFloat(0.012),
		decimal.NewFromFloat(-0.003),
		decimal.NewFromFloat(0.008),
	}

	riskFreeRate := decimal.NewFromFloat(0.0001)

	beta, alpha := calculateBetaAlpha(returns, benchmarkReturns, riskFreeRate)

	// Beta 应该为正数（资产与市场正相关）
	assert.True(t, beta.GreaterThan(decimal.Zero) || beta.Equal(decimal.Zero))
	// Alpha 可以为正或负
	assert.NotNil(t, alpha)
}

func TestCalculateTrackingError(t *testing.T) {
	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(-0.01),
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(-0.005),
		decimal.NewFromFloat(0.01),
	}

	benchmarkReturns := []decimal.Decimal{
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(-0.008),
		decimal.NewFromFloat(0.012),
		decimal.NewFromFloat(-0.003),
		decimal.NewFromFloat(0.008),
	}

	trackingError := calculateTrackingError(returns, benchmarkReturns)
	assert.True(t, trackingError.GreaterThanOrEqual(decimal.Zero))
}

func TestGetFirstKey(t *testing.T) {
	m := map[string][]decimal.Decimal{
		"SPY": {decimal.NewFromFloat(0.01)},
		"BND": {decimal.NewFromFloat(0.005)},
	}

	key := getFirstKey(m)
	assert.NotEmpty(t, key)
	_, exists := m[key]
	assert.True(t, exists)
}

func TestGetFirstKey_Empty(t *testing.T) {
	m := map[string][]decimal.Decimal{}
	key := getFirstKey(m)
	assert.Empty(t, key)
}

func TestRiskModels_ParametricVsHistoricalConsistency(t *testing.T) {
	rm := NewRiskModels()
	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01), decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(-0.01), decimal.NewFromFloat(0.03),
		decimal.NewFromFloat(-0.02), decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(0.008), decimal.NewFromFloat(-0.005),
		decimal.NewFromFloat(0.012), decimal.NewFromFloat(0.025),
	}
	confidence := 0.95

	parametricResult, _ := rm.CalculateParametricVaR(returns, confidence)
	historicalResult, _ := rm.CalculateHistoricalVaR(returns, confidence)

	if parametricResult == nil || historicalResult == nil {
		t.Skip("One or both methods returned nil")
	}

	pVar := parametricResult.VaR.InexactFloat64()
	hVar := historicalResult.VaR.InexactFloat64()

	if pVar == 0 && hVar == 0 {
		t.Skip("Both methods returned zero VaR")
	}

	ratio := math.Abs(pVar-hVar) / math.Max(math.Abs(hVar), 0.0001)
	if ratio > 0.5 {
		t.Errorf("VaR methods diverge too much: parametric=%.6f, historical=%.6f, ratio=%.2f",
			pVar, hVar, ratio)
	}
}

func TestRiskModels_EmptyReturns(t *testing.T) {
	rm := NewRiskModels()
	_, err := rm.CalculateHistoricalVaR([]decimal.Decimal{}, 0.95)
	if err == nil {
		t.Error("Empty returns should return error")
	}
}

func TestRiskModels_ConstantReturns(t *testing.T) {
	rm := NewRiskModels()
	constantReturns := make([]decimal.Decimal, 100)
	for i := range constantReturns {
		constantReturns[i] = decimal.NewFromFloat(0.01)
	}
	result, err := rm.CalculateHistoricalVaR(constantReturns, 0.95)
	if err != nil {
		t.Fatalf("Constant returns should not error: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}
