package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimalMean(t *testing.T) {
	tests := []struct {
		name     string
		values   []decimal.Decimal
		expected decimal.Decimal
	}{
		{
			"正常数据",
			Float64ToDecimalSlice([]float64{1, 2, 3, 4, 5}),
			decimal.NewFromFloat(3.0),
		},
		{
			"单个元素",
			Float64ToDecimalSlice([]float64{5}),
			decimal.NewFromFloat(5.0),
		},
		{
			"空切片",
			[]decimal.Decimal{},
			decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecimalMean(tt.values)
			assert.True(t, result.Equal(tt.expected))
		})
	}
}

func TestDecimalVariance(t *testing.T) {
	values := Float64ToDecimalSlice([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	mean := decimal.NewFromFloat(5.0)

	result := DecimalVariance(values, mean)
	expected := decimal.NewFromFloat(32.0 / 7.0)

	assert.True(t, result.Sub(expected).Abs().LessThan(decimal.NewFromFloat(1e-10)))
}

func TestDecimalStdDev(t *testing.T) {
	values := Float64ToDecimalSlice([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	mean := decimal.NewFromFloat(5.0)

	result := DecimalStdDev(values, mean)
	assert.True(t, result.IsPositive())
}

func TestDecimalLogReturns(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(110),
		decimal.NewFromFloat(105),
	}

	returns := DecimalLogReturns(prices)
	assert.Equal(t, 2, len(returns))

	// 验证第一个收益率: ln(110/100)
	expected1 := decimal.NewFromFloat(0.09531017980432486)
	assert.True(t, returns[0].Sub(expected1).Abs().LessThan(decimal.NewFromFloat(1e-10)))
}

func TestDecimalMaxDrawdown(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(110),
		decimal.NewFromFloat(90),
		decimal.NewFromFloat(95),
	}

	result := DecimalMaxDrawdown(prices)
	expected := decimal.NewFromFloat((110.0 - 90.0) / 110.0)

	assert.True(t, result.Sub(expected).Abs().LessThan(decimal.NewFromFloat(1e-10)))
}

func TestDecimalVaR(t *testing.T) {
	returns := Float64ToDecimalSlice([]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008})
	confidence := decimal.NewFromFloat(0.95)

	result := DecimalVaR(returns, confidence)
	assert.True(t, result.IsNegative()) // VaR 应该是负数
}

func TestDecimalCVaR(t *testing.T) {
	returns := Float64ToDecimalSlice([]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008})
	confidence := decimal.NewFromFloat(0.95)

	var95 := DecimalVaR(returns, confidence)
	cvar95 := DecimalCVaR(returns, confidence)

	// CVaR 应该比 VaR 更负
	assert.True(t, cvar95.LessThan(var95))
}

func TestDecimalSortinoRatio(t *testing.T) {
	returns := Float64ToDecimalSlice([]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008})
	riskFreeRate := decimal.NewFromFloat(0.02)

	result := DecimalSortinoRatio(returns, riskFreeRate)
	assert.False(t, result.IsZero())
}

func TestDecimalCalmarRatio(t *testing.T) {
	annualReturn := decimal.NewFromFloat(0.15)
	maxDrawdown := decimal.NewFromFloat(0.10)

	result := DecimalCalmarRatio(annualReturn, maxDrawdown)
	expected := decimal.NewFromFloat(1.5)

	assert.True(t, result.Equal(expected))
}

func TestDecimalSkewness(t *testing.T) {
	returns := Float64ToDecimalSlice([]float64{-0.03, -0.02, -0.01, 0, 0.01, 0.02, 0.03})
	result := DecimalSkewness(returns)

	// 对称分布偏度应该接近0
	assert.True(t, result.Abs().LessThan(decimal.NewFromFloat(0.5)))
}

func TestDecimalKurtosis(t *testing.T) {
	returns := Float64ToDecimalSlice([]float64{-0.03, -0.02, -0.01, 0, 0.01, 0.02, 0.03, 0.04})
	result := DecimalKurtosis(returns)

	// 正态分布峰度应该在合理范围内
	assert.True(t, result.GreaterThan(decimal.NewFromFloat(-3)))
}

func TestDecimalDownsideDeviation(t *testing.T) {
	returns := Float64ToDecimalSlice([]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005})
	targetReturn := decimal.Zero

	result := DecimalDownsideDeviation(returns, targetReturn)
	assert.True(t, result.IsPositive())
}

func TestDecimalToFloat64Slice(t *testing.T) {
	decimals := []decimal.Decimal{
		decimal.NewFromFloat(1.5),
		decimal.NewFromFloat(2.5),
		decimal.NewFromFloat(3.5),
	}

	floats := DecimalToFloat64Slice(decimals)
	assert.Equal(t, 3, len(floats))
	assert.InDelta(t, 1.5, floats[0], 1e-10)
	assert.InDelta(t, 2.5, floats[1], 1e-10)
	assert.InDelta(t, 3.5, floats[2], 1e-10)
}

func TestFloat64ToDecimalSlice(t *testing.T) {
	floats := []float64{1.5, 2.5, 3.5}

	decimals := Float64ToDecimalSlice(floats)
	assert.Equal(t, 3, len(decimals))
	assert.True(t, decimals[0].Equal(decimal.NewFromFloat(1.5)))
	assert.True(t, decimals[1].Equal(decimal.NewFromFloat(2.5)))
	assert.True(t, decimals[2].Equal(decimal.NewFromFloat(3.5)))
}
