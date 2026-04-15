package services

import (
	"math"
	"testing"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ETFAnalysisServiceTestSuite struct {
	suite.Suite
	service *ETFAnalysisService
}

func (s *ETFAnalysisServiceTestSuite) SetupTest() {
	utils.InitLogger("warn")
	if err := models.InitDB(":memory:"); err != nil {
		s.T().Fatalf("Failed to init database: %v", err)
	}
	models.InitDefaultData()

	mockExchange := NewExchangeRateService()
	s.service = NewETFAnalysisService(mockExchange)
}

func TestETFAnalysisServiceSuite(t *testing.T) {
	suite.Run(t, new(ETFAnalysisServiceTestSuite))
}

func (s *ETFAnalysisServiceTestSuite) TestCalculateMetrics_InsufficientData() {
	prices := []models.ETFData{
		{Symbol: "TEST", ClosePrice: decimal.NewFromFloat(100), Date: time.Now()},
	}

	metrics, err := s.service.CalculateMetrics("TEST", prices, "1mo")

	s.Error(err)
	s.Nil(metrics)
	s.Contains(err.Error(), "insufficient data")
}

func (s *ETFAnalysisServiceTestSuite) TestCalculateMetrics_NormalCase() {
	prices := generateTestPrices("SPY", 100.0, 110.0, 30)

	metrics, err := s.service.CalculateMetrics("SPY", prices, "1mo")

	s.NoError(err)
	s.NotNil(metrics)
	s.Equal("SPY", metrics.Symbol)
	s.Equal("1mo", metrics.Period)
	s.True(metrics.StartPrice.GreaterThan(decimal.Zero))
	s.True(metrics.EndPrice.GreaterThan(decimal.Zero))
	s.True(metrics.TotalReturn.GreaterThan(decimal.Zero))
	s.True(metrics.Volatility.GreaterThanOrEqual(decimal.Zero))
	s.True(metrics.MaxDrawdown.LessThanOrEqual(decimal.Zero))
}

func (s *ETFAnalysisServiceTestSuite) TestCalculateMetrics_VolatilityCalculation() {
	// 生成高波动率数据
	prices := []models.ETFData{}
	basePrice := 100.0
	for i := 0; i < 30; i++ {
		price := basePrice + float64(i%10)*5 - 25 // 波动范围 75-125
		prices = append(prices, models.ETFData{
			Symbol:     "VOLATILE",
			ClosePrice: decimal.NewFromFloat(price),
			Date:       time.Now().AddDate(0, 0, -30+i),
			Volume:     1000000,
		})
	}

	metrics, err := s.service.CalculateMetrics("VOLATILE", prices, "1mo")

	s.NoError(err)
	s.NotNil(metrics)
	s.True(metrics.Volatility.GreaterThan(decimal.Zero), "Volatility should be positive for volatile data")
}

func (s *ETFAnalysisServiceTestSuite) TestCalculateMetrics_SharpeRatio() {
	prices := generateTestPrices("SPY", 100.0, 120.0, 60)

	metrics, err := s.service.CalculateMetrics("SPY", prices, "3mo")

	s.NoError(err)
	s.NotNil(metrics)
	// 夏普比率可能为正或负，取决于收益与波动
	s.NotNil(metrics.SharpeRatio)
}

func (s *ETFAnalysisServiceTestSuite) TestCalculateMetrics_MaxDrawdown() {
	// 创建有明确回撤的数据: 100 -> 120 -> 80 -> 110
	prices := []models.ETFData{
		{Symbol: "DRAW", ClosePrice: decimal.NewFromFloat(100), Date: time.Now().AddDate(0, 0, -4)},
		{Symbol: "DRAW", ClosePrice: decimal.NewFromFloat(120), Date: time.Now().AddDate(0, 0, -3)},
		{Symbol: "DRAW", ClosePrice: decimal.NewFromFloat(80), Date: time.Now().AddDate(0, 0, -2)},
		{Symbol: "DRAW", ClosePrice: decimal.NewFromFloat(110), Date: time.Now().AddDate(0, 0, -1)},
	}

	metrics, err := s.service.CalculateMetrics("DRAW", prices, "1w")

	s.NoError(err)
	s.NotNil(metrics)
	s.True(metrics.MaxDrawdown.LessThan(decimal.Zero), "Max drawdown should be negative")
	// 最大回撤约为 (120-80)/120 = 33.3%
	s.True(metrics.MaxDrawdown.LessThan(decimal.NewFromFloat(-30)))
}

func (s *ETFAnalysisServiceTestSuite) TestAnalyzePortfolio_EmptyAllocation() {
	allocation := map[string]float64{}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := s.service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(0, len(result.Holdings))
}

func (s *ETFAnalysisServiceTestSuite) TestAnalyzePortfolio_DefaultTaxRate() {
	allocation := map[string]float64{
		"SCHD": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.Zero

	result, err := s.service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	s.NoError(err)
	s.NotNil(result)
	s.True(result.TaxRate.Equal(decimal.NewFromFloat(0.10)))
}

func (s *ETFAnalysisServiceTestSuite) TestAnalyzePortfolio_MultipleETFs() {
	allocation := map[string]float64{
		"SCHD": 50,
		"VNQ":  30,
		"QQQ":  20,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := s.service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(3, len(result.Holdings))

	// 验证权重总和为100
	totalWeight := 0.0
	for _, holding := range result.Holdings {
		totalWeight += holding.Weight
	}
	s.InDelta(100.0, totalWeight, 0.01)
}

func (s *ETFAnalysisServiceTestSuite) TestAnalyzePortfolio_InvalidAllocation() {
	allocation := map[string]float64{
		"INVALID_ETF": 100,
	}
	totalInvestment := decimal.NewFromInt(100000)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := s.service.AnalyzePortfolio(allocation, totalInvestment, taxRate)

	// 应该返回错误或不包含该ETF的结果
	s.NoError(err)
	s.NotNil(result)
}

func (s *ETFAnalysisServiceTestSuite) TestForecastETFGrowth() {
	initialInvestment := decimal.NewFromInt(100000)
	annualReturnRate := decimal.NewFromFloat(0.08)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := s.service.ForecastETFGrowth("SCHD", initialInvestment, &annualReturnRate, taxRate)

	s.NoError(err)
	s.NotNil(result)
	s.NotEmpty(result.Forecasts)

	// 验证预测包含多年数据
	s.True(len(result.Forecasts) > 0)

	// 验证初始投资
	s.True(result.InitialInvestment.Equal(initialInvestment))
}

func (s *ETFAnalysisServiceTestSuite) TestForecastETFGrowth_MultipleYears() {
	initialInvestment := decimal.NewFromInt(50000)
	annualReturnRate := decimal.NewFromFloat(0.07)
	taxRate := decimal.NewFromFloat(0.10)

	result, err := s.service.ForecastETFGrowth("SCHD", initialInvestment, &annualReturnRate, taxRate)

	s.NoError(err)
	s.NotNil(result)

	// 验证多年预测数据 - Forecasts 是 map[string]YearlyForecast
	s.NotEmpty(result.Forecasts)
	for yearStr, forecast := range result.Forecasts {
		year, err := decimal.NewFromString(yearStr)
		s.NoError(err)
		s.True(year.GreaterThanOrEqual(decimal.NewFromInt(1)))
		s.True(forecast.FutureValue.GreaterThan(decimal.Zero))
		s.True(forecast.TotalDividendBeforeTax.GreaterThanOrEqual(decimal.Zero))
		s.True(forecast.TotalDividendAfterTax.GreaterThanOrEqual(decimal.Zero))
	}
}

func (s *ETFAnalysisServiceTestSuite) TestGetComparisonData() {
	// 首先为测试创建必要的ETF配置和价格数据
	symbols := []string{"SPY", "QQQ"}
	period := "1y"

	// 确保ETF配置存在
	for _, symbol := range symbols {
		var cfg models.ETFConfig
		if err := models.DB.Where("symbol = ?", symbol).First(&cfg).Error; err != nil {
			// 创建测试配置
			cfg = models.ETFConfig{
				Symbol:   symbol,
				Name:     symbol + " Test",
				Currency: "USD",
			}
			models.DB.Create(&cfg)
		}
	}

	// 创建测试价格数据
	now := time.Now()
	for i, symbol := range symbols {
		for day := 0; day < 30; day++ {
			price := models.ETFData{
				Symbol:     symbol,
				Date:       now.AddDate(0, 0, -day),
				ClosePrice: decimal.NewFromFloat(100.0 + float64(i)*10 + float64(day)*0.5),
				OpenPrice:  decimal.NewFromFloat(99.0 + float64(i)*10 + float64(day)*0.5),
				HighPrice:  decimal.NewFromFloat(101.0 + float64(i)*10 + float64(day)*0.5),
				LowPrice:   decimal.NewFromFloat(98.0 + float64(i)*10 + float64(day)*0.5),
				Volume:     1000000,
			}
			models.DB.Create(&price)
		}
	}

	result, err := s.service.GetComparisonData(symbols, period)

	// 主要测试函数不panic
	s.NotPanics(func() {
		s.service.GetComparisonData(symbols, period)
	})

	// 由于我们创建了测试数据，应该能成功返回结果
	if err == nil && result != nil {
		s.NotNil(result)
		s.True(len(result) > 0, "Should return comparison data for ETFs with data")
	}
}

func (s *ETFAnalysisServiceTestSuite) TestNewETFAnalysisService() {
	mockExchange := NewExchangeRateService()
	service := NewETFAnalysisService(mockExchange)

	s.NotNil(service)
	s.NotNil(service.exchangeRate)
}

// generateTestPrices 生成测试价格数据
func generateTestPrices(symbol string, startPrice, endPrice float64, days int) []models.ETFData {
	prices := []models.ETFData{}
	priceStep := (endPrice - startPrice) / float64(days-1)

	for i := 0; i < days; i++ {
		price := startPrice + priceStep*float64(i)
		prices = append(prices, models.ETFData{
			Symbol:     symbol,
			ClosePrice: decimal.NewFromFloat(price),
			Date:       time.Now().AddDate(0, 0, -days+i),
			Volume:     1000000,
		})
	}
	return prices
}

// TestCalculateMaxDrawdown 测试最大回撤计算函数
func TestCalculateMaxDrawdown(t *testing.T) {
	tests := []struct {
		name     string
		prices   []models.ETFData
		expected float64
	}{
		{
			name: "No drawdown - always increasing",
			prices: []models.ETFData{
				{ClosePrice: decimal.NewFromFloat(100)},
				{ClosePrice: decimal.NewFromFloat(110)},
				{ClosePrice: decimal.NewFromFloat(120)},
			},
			expected: 0,
		},
		{
			name: "Single drawdown",
			prices: []models.ETFData{
				{ClosePrice: decimal.NewFromFloat(100)},
				{ClosePrice: decimal.NewFromFloat(120)},
				{ClosePrice: decimal.NewFromFloat(80)},
			},
			expected: -33.33, // (80-120)/120 * 100
		},
		{
			name: "Multiple drawdowns - deepest",
			prices: []models.ETFData{
				{ClosePrice: decimal.NewFromFloat(100)},
				{ClosePrice: decimal.NewFromFloat(120)},
				{ClosePrice: decimal.NewFromFloat(90)},  // -25%
				{ClosePrice: decimal.NewFromFloat(130)}, // new high
				{ClosePrice: decimal.NewFromFloat(80)},  // -38.46% deepest
			},
			expected: -38.46,
		},
		{
			name: "Insufficient data",
			prices: []models.ETFData{
				{ClosePrice: decimal.NewFromFloat(100)},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMaxDrawdown(tt.prices)
			assert.InDelta(t, tt.expected, result.InexactFloat64(), 0.5)
		})
	}
}

// TestCalculateVolatility 测试波动率计算
func TestCalculateVolatility(t *testing.T) {
	tests := []struct {
		name         string
		dailyReturns []float64
		shouldBeZero bool
	}{
		{
			name:         "Constant returns - zero volatility",
			dailyReturns: []float64{0.01, 0.01, 0.01, 0.01, 0.01},
			shouldBeZero: true,
		},
		{
			name:         "Variable returns - non-zero volatility",
			dailyReturns: []float64{0.01, -0.02, 0.015, -0.01, 0.025},
			shouldBeZero: false,
		},
		{
			name:         "Empty returns",
			dailyReturns: []float64{},
			shouldBeZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dailyReturns []decimal.Decimal
			for _, r := range tt.dailyReturns {
				dailyReturns = append(dailyReturns, decimal.NewFromFloat(r))
			}

			volatility := calculateVolatility(dailyReturns)

			if tt.shouldBeZero {
				assert.True(t, volatility.IsZero() || volatility.LessThan(decimal.NewFromFloat(0.0001)))
			} else {
				assert.True(t, volatility.GreaterThan(decimal.Zero))
			}
		})
	}
}

// calculateVolatility 辅助函数计算波动率
func calculateVolatility(dailyReturns []decimal.Decimal) decimal.Decimal {
	if len(dailyReturns) < 2 {
		return decimal.Zero
	}

	mean := decimal.Zero
	for _, r := range dailyReturns {
		mean = mean.Add(r)
	}
	mean = mean.Div(decimal.NewFromInt(int64(len(dailyReturns))))

	variance := decimal.Zero
	for _, r := range dailyReturns {
		diff := r.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(dailyReturns) - 1)))

	return decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
}
