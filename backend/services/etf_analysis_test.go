package services

import (
	"testing"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzePortfolio_EmptyAllocation(t *testing.T) {
	utils.InitLogger("warn")

	// 初始化内存数据库
	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

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

	// 初始化内存数据库
	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	// 初始化默认 ETF 数据
	models.InitDefaultData()

	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

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
