package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRSI(t *testing.T) {
	ti := NewTechnicalIndicators()

	// 测试数据：模拟一个上涨后下跌的价格序列
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(102),
		decimal.NewFromFloat(101),
		decimal.NewFromFloat(103),
		decimal.NewFromFloat(105),
		decimal.NewFromFloat(104),
		decimal.NewFromFloat(106),
		decimal.NewFromFloat(108),
		decimal.NewFromFloat(107),
		decimal.NewFromFloat(109),
		decimal.NewFromFloat(110),
		decimal.NewFromFloat(108),
		decimal.NewFromFloat(111),
		decimal.NewFromFloat(113),
		decimal.NewFromFloat(112),
	}

	result, err := ti.CalculateRSI(prices, 14)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.RSI.GreaterThan(decimal.Zero))
	assert.True(t, result.RSI.LessThanOrEqual(decimal.NewFromInt(100)))
}

func TestCalculateRSI_InsufficientData(t *testing.T) {
	ti := NewTechnicalIndicators()

	prices := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(102),
	}

	result, err := ti.CalculateRSI(prices, 14)

	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientData, err)
	assert.Nil(t, result)
}

func TestCalculateRSI_Overbought(t *testing.T) {
	ti := NewTechnicalIndicators()

	// 模拟持续上涨的价格序列
	prices := make([]decimal.Decimal, 20)
	prices[0] = decimal.NewFromFloat(100)
	for i := 1; i < 20; i++ {
		prices[i] = prices[i-1].Add(decimal.NewFromFloat(2))
	}

	result, err := ti.CalculateRSI(prices, 14)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.RSI.GreaterThan(decimal.NewFromInt(70)))
	assert.True(t, result.Overbought)
}

func TestCalculateRSI_Oversold(t *testing.T) {
	ti := NewTechnicalIndicators()

	// 模拟持续下跌的价格序列
	prices := make([]decimal.Decimal, 20)
	prices[0] = decimal.NewFromFloat(100)
	for i := 1; i < 20; i++ {
		prices[i] = prices[i-1].Sub(decimal.NewFromFloat(2))
	}

	result, err := ti.CalculateRSI(prices, 14)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.RSI.LessThan(decimal.NewFromInt(30)))
	assert.True(t, result.Oversold)
}

func TestCalculateMACD(t *testing.T) {
	ti := NewTechnicalIndicators()

	// 生成测试价格数据
	prices := make([]decimal.Decimal, 50)
	prices[0] = decimal.NewFromFloat(100)
	for i := 1; i < 50; i++ {
		// 添加一些随机波动
		change := decimal.NewFromFloat(float64(i%5-2) * 0.5)
		prices[i] = prices[i-1].Add(change)
	}

	result, err := ti.CalculateMACD(prices, 12, 26, 9)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCalculateMACD_InsufficientData(t *testing.T) {
	ti := NewTechnicalIndicators()

	prices := make([]decimal.Decimal, 20)
	for i := 0; i < 20; i++ {
		prices[i] = decimal.NewFromFloat(100 + float64(i))
	}

	result, err := ti.CalculateMACD(prices, 12, 26, 9)

	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientData, err)
	assert.Nil(t, result)
}

func TestCalculateBollingerBands(t *testing.T) {
	ti := NewTechnicalIndicators()

	// 生成测试价格数据
	prices := make([]decimal.Decimal, 30)
	prices[0] = decimal.NewFromFloat(100)
	for i := 1; i < 30; i++ {
		prices[i] = prices[i-1].Add(decimal.NewFromFloat(float64(i%3-1) * 2))
	}

	result, err := ti.CalculateBollingerBands(prices, 20, 2)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Upper.GreaterThan(result.Middle))
	assert.True(t, result.Lower.LessThan(result.Middle))
	assert.True(t, result.Bandwidth.GreaterThan(decimal.Zero))
}

func TestCalculateBollingerBands_InsufficientData(t *testing.T) {
	ti := NewTechnicalIndicators()

	prices := make([]decimal.Decimal, 10)
	for i := 0; i < 10; i++ {
		prices[i] = decimal.NewFromFloat(100)
	}

	result, err := ti.CalculateBollingerBands(prices, 20, 2)

	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientData, err)
	assert.Nil(t, result)
}

func TestCalculateMovingAverages(t *testing.T) {
	ti := NewTechnicalIndicators()

	prices := make([]decimal.Decimal, 50)
	prices[0] = decimal.NewFromFloat(100)
	for i := 1; i < 50; i++ {
		prices[i] = prices[i-1].Add(decimal.NewFromFloat(float64(i%5 - 2)))
	}

	periods := []int{5, 10, 20}
	result, err := ti.CalculateMovingAverages(prices, periods)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.SMA))
	assert.Equal(t, 3, len(result.EMA))

	for _, period := range periods {
		assert.True(t, result.SMA[period].GreaterThan(decimal.Zero))
		assert.True(t, result.EMA[period].GreaterThan(decimal.Zero))
	}
}

func TestCalculateMovingAverages_InsufficientData(t *testing.T) {
	ti := NewTechnicalIndicators()

	prices := make([]decimal.Decimal, 10)
	for i := 0; i < 10; i++ {
		prices[i] = decimal.NewFromFloat(100)
	}

	periods := []int{5, 20}
	result, err := ti.CalculateMovingAverages(prices, periods)

	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientData, err)
	assert.Nil(t, result)
}

func TestCalculateAverage(t *testing.T) {
	values := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(20),
		decimal.NewFromFloat(30),
	}

	avg := calculateAverage(values)
	// 使用 InDelta 比较浮点数
	assert.InDelta(t, 20.0, avg.InexactFloat64(), 0.0001)
}

func TestCalculateAverage_Empty(t *testing.T) {
	values := []decimal.Decimal{}
	avg := calculateAverage(values)
	assert.Equal(t, decimal.Zero, avg)
}

func TestCalculateSMA(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(20),
		decimal.NewFromFloat(30),
	}

	sma := calculateSMA(prices)
	assert.InDelta(t, 20.0, sma.InexactFloat64(), 0.0001)
}

func TestCalculateEMA(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(12),
		decimal.NewFromFloat(14),
		decimal.NewFromFloat(16),
		decimal.NewFromFloat(18),
	}

	ema := calculateEMA(prices, 3)
	assert.NotNil(t, ema)
	assert.True(t, len(ema) > 0)
}

func TestCalculateVariance(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(20),
		decimal.NewFromFloat(30),
	}
	mean := decimal.NewFromFloat(20)

	variance := calculateVariance(prices, mean)
	assert.True(t, variance.GreaterThan(decimal.Zero))
}
