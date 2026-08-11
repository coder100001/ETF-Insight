package services

import (
	"testing"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzePortfolio_BasicAllocation(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 50,
		"VTI":  50,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Holdings))
	assert.True(t, result.TotalInvestment.Equal(totalInvestment))
}

func TestAnalyzePortfolio_ZeroWeight(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 0,
		"VTI":  100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Holdings))
}

func TestAnalyzePortfolio_SingleHolding(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(50000)
	taxRate := decimal.NewFromFloat(0.15)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Holdings))

	holding := result.Holdings[0]
	assert.Equal(t, "SCHD", holding.Symbol)
	assert.InDelta(t, 50000.0, holding.Investment, 0.01)
	assert.True(t, holding.Weight > 0)
}

func TestAnalyzePortfolio_MultipleHoldings(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 30,
		"VTI":  40,
		"BND":  30,
	}
	totalInvestment := decimal.NewFromInt(200000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.Holdings))

	totalWeight := 0.0
	for _, h := range result.Holdings {
		totalWeight += h.Weight
	}
	assert.InDelta(t, 100.0, totalWeight, 0.01)
}

func TestAnalyzePortfolio_WeightedDividendYield(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 50,
		"VTI":  50,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.WeightedDividendYield.GreaterThan(decimal.Zero))
}

func TestAnalyzePortfolio_TotalValueCalculation(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 50,
		"VTI":  50,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalValue.GreaterThan(decimal.Zero))

	var holdingsTotalValue decimal.Decimal
	for _, h := range result.Holdings {
		holdingsTotalValue = holdingsTotalValue.Add(decimal.NewFromFloat(h.CurrentValueUSD))
	}
	assert.True(t, result.TotalValue.Equal(holdingsTotalValue))
}

func TestAnalyzePortfolio_DividendCalculation(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.AnnualDividendBeforeTax.GreaterThan(decimal.Zero))
	assert.True(t, result.AnnualDividendAfterTax.LessThan(result.AnnualDividendBeforeTax))
	assert.True(t, result.DividendTax.GreaterThan(decimal.Zero))
}

func TestAnalyzePortfolio_TotalReturnWithDividend(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 50,
		"VTI":  50,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalReturnWithDividend.Equal(result.TotalReturn.Add(result.AnnualDividendAfterTax)))
}

func TestAnalyzePortfolio_InvalidTaxRate(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(-0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAnalyzePortfolio_ExchangeRatesInitialized(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ExchangeRates)
}

func TestAnalyzePortfolio_HoldingsContainRequiredFields(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	if len(result.Holdings) > 0 {
		h := result.Holdings[0]
		assert.NotEmpty(t, h.Symbol)
		assert.NotEmpty(t, h.Name)
		assert.NotEmpty(t, h.Currency)
		assert.True(t, h.Weight >= 0)
		assert.True(t, h.Investment >= 0)
		assert.True(t, h.Shares >= 0)
		assert.True(t, h.CurrentPrice >= 0)
		assert.True(t, h.CurrentValue >= 0)
	}
}

func TestAnalyzePortfolio_LargeInvestment(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 50,
		"VTI":  50,
	}
	totalInvestment := decimal.NewFromFloat(10000000.00)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalValue.GreaterThan(decimal.NewFromFloat(9000000)))
}

func TestAnalyzePortfolio_EqualWeights(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	holdings := []string{"SCHD", "VTI", "BND", "VNQ"}
	allocation := make(map[string]float64)
	for _, h := range holdings {
		allocation[h] = 100.0 / float64(len(holdings))
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(holdings), len(result.Holdings))
}

func TestAnalyzePortfolio_TotalReturnPercent(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalReturnPercent.Equal(result.TotalReturn.Div(totalInvestment).Mul(decimal.NewFromInt(100))))
}

func TestAnalyzePortfolio_TotalReturnWithDividendPercent(t *testing.T) {
	utils.InitLogger("warn")

	if err := models.InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := newTestExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	expected := result.TotalReturnWithDividend.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	assert.True(t, result.TotalReturnWithDividendPercent.Equal(expected))
}
