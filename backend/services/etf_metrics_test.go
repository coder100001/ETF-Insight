package services

import (
	"testing"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateMetrics_InsufficientData(t *testing.T) {
	service := NewETFAnalysisService(nil)

	_, err := service.CalculateMetrics("SPY", []models.ETFData{}, "1y")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestCalculateMetrics_SingleDataPoint(t *testing.T) {
	service := NewETFAnalysisService(nil)

	_, err := service.CalculateMetrics("SPY", []models.ETFData{
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(100)},
	}, "1y")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestCalculateMetrics_Normal(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, 0, -20), ClosePrice: decimal.NewFromFloat(102), Volume: 1100000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, 0, -10), ClosePrice: decimal.NewFromFloat(105), Volume: 900000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(108), Volume: 1200000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SPY", result.Symbol)
	assert.Equal(t, "1y", result.Period)
	assert.True(t, result.TotalReturn.GreaterThan(decimal.Zero))
	assert.Equal(t, 4, result.TradingDays)
}

func TestCalculateMetrics_ZeroVolatility(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.SharpeRatio.Equal(decimal.Zero))
	assert.True(t, result.Volatility.Equal(decimal.Zero))
}

func TestCalculateMetrics_NegativeReturn(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(90), Volume: 1100000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalReturn.LessThan(decimal.Zero))
}

func TestCalculateMetrics_MaxDrawdown_Normal(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -4, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(110), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(105), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(95), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.MaxDrawdown.LessThan(decimal.Zero))
	assert.True(t, result.MaxDrawdown.GreaterThan(decimal.NewFromFloat(-20)))
}

func TestCalculateMetrics_MaxDrawdown_NoDrawdown(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -4, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(105), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(110), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(115), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(120), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.MaxDrawdown.Equal(decimal.Zero))
}

func TestCalculateMetrics_MaxDrawdown_CompleteDrawdown(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(50), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(50), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.MaxDrawdown.Equal(decimal.NewFromFloat(-50)))
}

func TestCalculateMetrics_VolumeAverage(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -4, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(105), Volume: 2000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(110), Volume: 3000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(115), Volume: 4000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(120), Volume: 5000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2800000), result.AvgVolume)
}

func TestCalculateMetrics_HighVolatility(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -5, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -4, 0), ClosePrice: decimal.NewFromFloat(150), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(75), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(112), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(56), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(84), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Volatility.GreaterThan(decimal.NewFromFloat(50)))
}

func TestCalculateMetrics_SortByDate(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(108), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(105), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(102), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.StartPrice.Equal(decimal.NewFromFloat(100)))
	assert.True(t, result.EndPrice.Equal(decimal.NewFromFloat(108)))
	assert.True(t, result.TotalReturn.GreaterThan(decimal.Zero))
}

func TestCalculateMetrics_MaxMinPrice(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(95), Volume: 1000000, HighPrice: decimal.NewFromFloat(100), LowPrice: decimal.NewFromFloat(92)},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(110), Volume: 1000000, HighPrice: decimal.NewFromFloat(115), LowPrice: decimal.NewFromFloat(108)},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(105), Volume: 1000000, HighPrice: decimal.NewFromFloat(110), LowPrice: decimal.NewFromFloat(103)},
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(108), Volume: 1000000, HighPrice: decimal.NewFromFloat(112), LowPrice: decimal.NewFromFloat(106)},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.MaxPrice.Equal(decimal.NewFromFloat(110)))
	assert.True(t, result.MinPrice.Equal(decimal.NewFromFloat(95)))
}

func TestCalculateMetrics_SharpeRatioValidation(t *testing.T) {
	service := NewETFAnalysisService(nil)

	testCases := []struct {
		name         string
		startPrice   float64
		endPrice     float64
		expectedSign int
		description  string
	}{
		{"Positive excess return", 100, 112, 1, "Return > RiskFree"},
		{"Negative excess return", 100, 99, -1, "Return < RiskFree"},
		{"Zero volatility case", 100, 100, 0, "No price change"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var prices []models.ETFData
			if tc.name == "Zero volatility case" {
				prices = []models.ETFData{
					{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(tc.startPrice), Volume: 1000000},
					{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(tc.endPrice), Volume: 1000000},
				}
			} else {
				baseDate := time.Now().AddDate(0, 0, -10)
				for i := range 11 {
					progress := float64(i) / 10.0
					price := tc.startPrice + (tc.endPrice-tc.startPrice)*progress
					noise := (float64(i%3) - 1.0) * 0.5
					price = price + noise
					prices = append(prices, models.ETFData{
						Symbol:     "SPY",
						Date:       baseDate.AddDate(0, 0, i),
						ClosePrice: decimal.NewFromFloat(price),
						Volume:     1000000,
					})
				}
			}

			result, err := service.CalculateMetrics("SPY", prices, "1y")

			assert.NoError(t, err)
			assert.NotNil(t, result)

			switch tc.expectedSign {
			case 1:
				assert.True(t, result.SharpeRatio.GreaterThan(decimal.Zero), tc.description)
			case -1:
				assert.True(t, result.SharpeRatio.LessThan(decimal.Zero), tc.description)
			case 0:
				assert.True(t, result.SharpeRatio.Equal(decimal.Zero), tc.description)
			}
		})
	}
}

func TestCalculateMetrics_LargeDataset(t *testing.T) {
	service := NewETFAnalysisService(nil)

	var prices []models.ETFData
	basePrice := 100.0
	for i := 252; i >= 0; i-- {
		price := basePrice + float64(i%20)*0.5
		prices = append(prices, models.ETFData{
			Symbol:     "SPY",
			Date:       time.Now().AddDate(0, 0, -i),
			ClosePrice: decimal.NewFromFloat(price),
			Volume:     1000000,
		})
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 253, result.TradingDays)
	assert.True(t, result.Volatility.GreaterThan(decimal.Zero))
}

func TestCalculateMetrics_UnsortedDates(t *testing.T) {
	service := NewETFAnalysisService(nil)

	prices := []models.ETFData{
		{Symbol: "SPY", Date: time.Now(), ClosePrice: decimal.NewFromFloat(108), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -2, 0), ClosePrice: decimal.NewFromFloat(102), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -1, 0), ClosePrice: decimal.NewFromFloat(105), Volume: 1000000},
		{Symbol: "SPY", Date: time.Now().AddDate(0, -3, 0), ClosePrice: decimal.NewFromFloat(100), Volume: 1000000},
	}

	result, err := service.CalculateMetrics("SPY", prices, "1y")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.StartPrice.Equal(decimal.NewFromFloat(100)))
	assert.True(t, result.EndPrice.Equal(decimal.NewFromFloat(108)))
}
