package services

import (
	"testing"
	"time"

	"etf-insight/config"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzePortfolio_Basic(t *testing.T) {
	// 初始化日志
	utils.InitLogger("warn")

	// 创建测试数据
	cacheCfg := &config.CacheConfig{
		RealtimeTTL:   5 * time.Minute,
		HistoricalTTL: 24 * time.Hour,
		MetricsTTL:    1 * time.Hour,
		ComparisonTTL: 30 * time.Minute,
	}
	mockCache := NewCacheService(cacheCfg)
	defer mockCache.Close()

	// 添加模拟数据
	mockCache.SetRealtimeData("SCHD", &RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	})

	mockCache.SetRealtimeData("SPYD", &RealtimeData{
		Symbol:        "SPYD",
		Name:          "SPDR S&P 500 High Dividend ETF",
		CurrentPrice:  47.85,
		DividendYield: 4.12,
	})

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockCache, mockExchange)

	allocation := map[string]float64{
		"SCHD": 40,
		"SPYD": 60,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// 使用近似比较，允许浮点数精度误差
	assert.InDelta(t, 100000.0, result.TotalValue.InexactFloat64(), 0.01)
	assert.Equal(t, 2, len(result.Holdings))
}

func TestAnalyzePortfolio_TaxCalculation(t *testing.T) {
	utils.InitLogger("warn")

	cacheCfg := &config.CacheConfig{
		RealtimeTTL:   5 * time.Minute,
		HistoricalTTL: 24 * time.Hour,
		MetricsTTL:    1 * time.Hour,
		ComparisonTTL: 30 * time.Minute,
	}
	mockCache := NewCacheService(cacheCfg)
	defer mockCache.Close()

	mockCache.SetRealtimeData("SCHD", &RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	})

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockCache, mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.20)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	expectedDividendBeforeTax := decimal.NewFromFloat(3450.0)
	expectedDividendAfterTax := decimal.NewFromFloat(2760.0)

	assert.True(t, result.AnnualDividendBeforeTax.Equal(expectedDividendBeforeTax))
	assert.True(t, result.AnnualDividendAfterTax.Equal(expectedDividendAfterTax))
}

func TestAnalyzePortfolio_EmptyAllocation(t *testing.T) {
	utils.InitLogger("warn")

	cacheCfg := &config.CacheConfig{
		RealtimeTTL:   5 * time.Minute,
		HistoricalTTL: 24 * time.Hour,
		MetricsTTL:    1 * time.Hour,
		ComparisonTTL: 30 * time.Minute,
	}
	mockCache := NewCacheService(cacheCfg)
	defer mockCache.Close()

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockCache, mockExchange)

	allocation := map[string]float64{}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Holdings))
}

func TestAnalyzePortfolio_DefaultTaxRate(t *testing.T) {
	utils.InitLogger("warn")

	cacheCfg := &config.CacheConfig{
		RealtimeTTL:   5 * time.Minute,
		HistoricalTTL: 24 * time.Hour,
		MetricsTTL:    1 * time.Hour,
		ComparisonTTL: 30 * time.Minute,
	}
	mockCache := NewCacheService(cacheCfg)
	defer mockCache.Close()

	mockCache.SetRealtimeData("SCHD", &RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	})

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockCache, mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.Zero

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TaxRate.Equal(decimal.NewFromFloat(0.10)))
}

func TestAnalyzePortfolio_MultipleHoldings(t *testing.T) {
	utils.InitLogger("warn")

	cacheCfg := &config.CacheConfig{
		RealtimeTTL:   5 * time.Minute,
		HistoricalTTL: 24 * time.Hour,
		MetricsTTL:    1 * time.Hour,
		ComparisonTTL: 30 * time.Minute,
	}
	mockCache := NewCacheService(cacheCfg)
	defer mockCache.Close()

	mockCache.SetRealtimeData("SCHD", &RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	})

	mockCache.SetRealtimeData("SPYD", &RealtimeData{
		Symbol:        "SPYD",
		Name:          "SPDR S&P 500 High Dividend ETF",
		CurrentPrice:  47.85,
		DividendYield: 4.12,
	})

	mockCache.SetRealtimeData("JEPQ", &RealtimeData{
		Symbol:        "JEPQ",
		Name:          "JPMorgan Nasdaq Equity Premium Income ETF",
		CurrentPrice:  57.20,
		DividendYield: 11.2,
	})

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockCache, mockExchange)

	allocation := map[string]float64{
		"SCHD": 40,
		"SPYD": 30,
		"JEPQ": 30,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.Holdings))
	// 使用近似比较，允许浮点数精度误差
	assert.InDelta(t, 100000.0, result.TotalValue.InexactFloat64(), 0.01)

	// 验证加权股息率
	weightedYield := result.WeightedDividendYield.InexactFloat64()
	expectedYield := 3.45*0.4 + 4.12*0.3 + 11.2*0.3
	assert.InDelta(t, expectedYield, weightedYield, 0.01)
}
