package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) GetRealtimeData(symbol string) (*RealtimeData, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RealtimeData), args.Error(1)
}

func (m *MockCacheService) SetRealtimeData(symbol string, data *RealtimeData) {
	m.Called(symbol, data)
}

func (m *MockCacheService) Close() {
	m.Called()
}

func (m *MockCacheService) GetCacheStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockCacheService) GetRealtimeDataJSON(symbol string) ([]byte, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheService) GetHistoricalData(symbol string, period string) ([]byte, error) {
	args := m.Called(symbol, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheService) SetHistoricalData(symbol string, period string, data []byte) {
	m.Called(symbol, period, data)
}

func (m *MockCacheService) GetMetrics(symbol string, period string) ([]byte, error) {
	args := m.Called(symbol, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheService) SetMetrics(symbol string, period string, data []byte) {
	m.Called(symbol, period, data)
}

func (m *MockCacheService) GetComparison(symbols []string, period string) ([]byte, error) {
	args := m.Called(symbols, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheService) SetComparison(symbols []string, period string, data []byte) {
	m.Called(symbols, period, data)
}

type MockExchangeRateService struct {
	mock.Mock
}

func (m *MockExchangeRateService) GetRate(from, to string) float64 {
	args := m.Called(from, to)
	return args.Get(0).(float64)
}

func TestAnalyzePortfolio(t *testing.T) {
	mockCache := new(MockCacheService)
	mockExchange := new(MockExchangeRateService)

	service := NewETFAnalysisService(mockCache, mockExchange)

	mockCache.On("GetRealtimeData", "SCHD").Return(&RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	}, nil)

	mockCache.On("GetRealtimeData", "SPYD").Return(&RealtimeData{
		Symbol:        "SPYD",
		Name:          "SPDR S&P 500 High Dividend ETF",
		CurrentPrice:  47.85,
		DividendYield: 4.12,
	}, nil)

	mockCache.On("Close").Return()

	allocation := map[string]float64{
		"SCHD": 40,
		"SPYD": 60,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalValue.Equal(decimal.NewFromInt(100000)))
	assert.Equal(t, 2, len(result.Holdings))

	mockCache.AssertExpectations(t)
}

func TestAnalyzePortfolio_TaxCalculation(t *testing.T) {
	mockCache := new(MockCacheService)
	mockExchange := new(MockExchangeRateService)

	service := NewETFAnalysisService(mockCache, mockExchange)

	mockCache.On("GetRealtimeData", "SCHD").Return(&RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	}, nil)

	mockCache.On("Close").Return()

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

	mockCache.AssertExpectations(t)
}

func TestAnalyzePortfolio_EmptyAllocation(t *testing.T) {
	mockCache := new(MockCacheService)
	mockExchange := new(MockExchangeRateService)

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
	mockCache := new(MockCacheService)
	mockExchange := new(MockExchangeRateService)

	service := NewETFAnalysisService(mockCache, mockExchange)

	mockCache.On("GetRealtimeData", "SCHD").Return(&RealtimeData{
		Symbol:        "SCHD",
		Name:          "Schwab US Dividend Equity ETF",
		CurrentPrice:  30.44,
		DividendYield: 3.45,
	}, nil)

	mockCache.On("Close").Return()

	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.Zero

	result, err := service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TaxRate.Equal(decimal.NewFromFloat(0.10)))

	mockCache.AssertExpectations(t)
}
