package services

import (
	"math"
	"testing"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// PortfolioAnalyticsTestSuite 组合分析测试套件
type PortfolioAnalyticsTestSuite struct {
	suite.Suite
	service *PortfolioAnalyticsService
	db      *gorm.DB
}

func (suite *PortfolioAnalyticsTestSuite) SetupSuite() {
	// 使用 SQLite 内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(&models.ETFData{})
	suite.Require().NoError(err)

	suite.db = db
	suite.service = NewPortfolioAnalyticsService()
	suite.service.SetDB(db)
}

func (suite *PortfolioAnalyticsTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func TestPortfolioAnalyticsSuite(t *testing.T) {
	suite.Run(t, new(PortfolioAnalyticsTestSuite))
}

// ========== 基础统计函数测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestMean() {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"正常数据", []float64{1, 2, 3, 4, 5}, 3.0},
		{"单个元素", []float64{5}, 5.0},
		{"空切片", []float64{}, 0},
		{"负数", []float64{-1, -2, -3}, -2.0},
		{"混合正负数", []float64{-1, 1}, 0.0},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.mean(tt.values)
			assert.InDelta(suite.T(), tt.expected, result, 1e-10)
		})
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestVariance() {
	tests := []struct {
		name     string
		values   []float64
		mean     float64
		expected float64
	}{
		// 样本方差 = sum((x-mean)^2) / (N-1)
		// 数据: {2,4,4,4,5,5,7,9}, mean=5
		// 偏差平方和 = 9+1+1+1+0+0+4+16 = 32
		// 样本方差 = 32/7 ≈ 4.5714
		{"正常数据", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 5.0, 32.0 / 7.0},
		{"单个元素", []float64{5}, 5.0, 0},
		{"空切片", []float64{}, 0, 0},
		{"相同值", []float64{3, 3, 3}, 3.0, 0},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.variance(tt.values, tt.mean)
			assert.InDelta(suite.T(), tt.expected, result, 1e-10)
		})
	}
}

// ========== 对数收益率测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateLogReturns() {
	tests := []struct {
		name     string
		prices   []decimal.Decimal
		expected int
	}{
		{
			"正常价格序列",
			[]decimal.Decimal{
				decimal.NewFromFloat(100),
				decimal.NewFromFloat(110),
				decimal.NewFromFloat(105),
			},
			2,
		},
		{
			"单个价格",
			[]decimal.Decimal{decimal.NewFromFloat(100)},
			0,
		},
		{
			"价格不变",
			[]decimal.Decimal{
				decimal.NewFromFloat(100),
				decimal.NewFromFloat(100),
			},
			1,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.CalculateLogReturns(tt.prices)
			assert.Equal(suite.T(), tt.expected, len(result))
		})
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateLogReturns_Values() {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(110),
		decimal.NewFromFloat(105),
	}

	result := suite.service.CalculateLogReturns(prices)
	assert.Equal(suite.T(), 2, len(result))
	assert.InDelta(suite.T(), math.Log(110.0/100.0), result[0], 1e-10)
	assert.InDelta(suite.T(), math.Log(105.0/110.0), result[1], 1e-10)
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateLogReturns_ZeroPrice() {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(100),
	}

	result := suite.service.CalculateLogReturns(prices)
	assert.Equal(suite.T(), 1, len(result))
	assert.Equal(suite.T(), float64(0), result[0]) // 零价格时返回0
}

// ========== VaR 测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateVaR() {
	tests := []struct {
		name       string
		returns    []float64
		confidence float64
		checkFunc  func(float64) bool
	}{
		{
			"正常收益率",
			[]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008},
			0.95,
			func(v float64) bool { return v < 0 }, // VaR 应该是负数
		},
		{
			"空收益率",
			[]float64{},
			0.95,
			func(v float64) bool { return v == 0 },
		},
		{
			"99%置信度",
			[]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008},
			0.99,
			func(v float64) bool { return v < 0 },
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.CalculateVaR(tt.returns, tt.confidence)
			assert.True(suite.T(), tt.checkFunc(result))
		})
	}
}

// ========== CVaR 测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateCVaR() {
	returns := []float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008}

	var95 := suite.service.CalculateVaR(returns, 0.95)
	cvar95 := suite.service.CalculateCVaR(returns, 0.95)

	// CVaR 应该比 VaR 更负 (更保守)
	assert.Less(suite.T(), cvar95, var95)

	// 空收益率返回0
	cvarEmpty := suite.service.CalculateCVaR([]float64{}, 0.95)
	assert.Equal(suite.T(), float64(0), cvarEmpty)
}

// ========== 索提诺比率测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateSortinoRatio() {
	tests := []struct {
		name         string
		returns      []float64
		riskFreeRate float64
		checkFunc    func(float64) bool
	}{
		{
			"有下行风险",
			[]float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008},
			0.02,
			func(v float64) bool { return v != 0 },
		},
		{
			"全部正收益",
			[]float64{0.01, 0.02, 0.015, 0.01, 0.005},
			0.02,
			func(v float64) bool { return v == 0 }, // 没有下行风险
		},
		{
			"空收益率",
			[]float64{},
			0.02,
			func(v float64) bool { return v == 0 },
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.CalculateSortinoRatio(tt.returns, tt.riskFreeRate)
			assert.True(suite.T(), tt.checkFunc(result))
		})
	}
}

// ========== 卡尔玛比率测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateCalmarRatio() {
	tests := []struct {
		name         string
		annualReturn float64
		maxDrawdown  float64
		expected     float64
	}{
		{"正常计算", 0.15, 0.10, 1.5},
		{"零回撤", 0.15, 0, 0},
		{"负收益", -0.05, 0.10, -0.5},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.CalculateCalmarRatio(tt.annualReturn, tt.maxDrawdown)
			assert.InDelta(suite.T(), tt.expected, result, 1e-10)
		})
	}
}

// ========== 下行偏差测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateDownsideDeviation() {
	returns := []float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005}
	targetReturn := 0.0

	result := suite.service.CalculateDownsideDeviation(returns, targetReturn)
	assert.Greater(suite.T(), result, float64(0))

	// 空收益率返回0
	resultEmpty := suite.service.CalculateDownsideDeviation([]float64{}, targetReturn)
	assert.Equal(suite.T(), float64(0), resultEmpty)

	// 所有收益高于目标返回0
	resultAllAbove := suite.service.CalculateDownsideDeviation([]float64{0.01, 0.02}, targetReturn)
	assert.Equal(suite.T(), float64(0), resultAllAbove)
}

// ========== 偏度测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateSkewness() {
	tests := []struct {
		name     string
		returns  []float64
		expected float64
	}{
		{
			"正态分布近似",
			[]float64{-0.03, -0.02, -0.01, 0, 0.01, 0.02, 0.03},
			0, // 对称分布偏度接近0
		},
		{
			"右偏分布",
			[]float64{-0.01, 0, 0.01, 0.02, 0.03, 0.04, 0.05},
			0, // 需要更多数据点才能有明显偏度
		},
		{
			"数据不足",
			[]float64{0.01, 0.02},
			0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.CalculateSkewness(tt.returns)
			// 对称分布偏度应该接近0
			assert.InDelta(suite.T(), tt.expected, result, 0.5)
		})
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateSkewness_ConstantReturns() {
	returns := []float64{0.01, 0.01, 0.01, 0.01, 0.01}
	result := suite.service.CalculateSkewness(returns)
	assert.Equal(suite.T(), float64(0), result) // 常数收益率偏度为0
}

// ========== 峰度测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateKurtosis() {
	tests := []struct {
		name     string
		returns  []float64
		expected float64
		tol      float64
	}{
		{
			"正态分布近似",
			[]float64{-0.03, -0.02, -0.01, 0, 0.01, 0.02, 0.03, 0.04},
			-1.2, // 样本峰度可能偏离理论值
			1.0,
		},
		{
			"数据不足",
			[]float64{0.01, 0.02, 0.03},
			0,
			1e-10,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.CalculateKurtosis(tt.returns)
			assert.InDelta(suite.T(), tt.expected, result, tt.tol)
		})
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateKurtosis_ConstantReturns() {
	returns := []float64{0.01, 0.01, 0.01, 0.01, 0.01}
	result := suite.service.CalculateKurtosis(returns)
	assert.Equal(suite.T(), float64(0), result) // 常数收益率峰度为0
}

// ========== 最大回撤测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateMaxDrawdown() {
	tests := []struct {
		name     string
		prices   []decimal.Decimal
		expected float64
	}{
		{
			"有回撤",
			[]decimal.Decimal{
				decimal.NewFromFloat(100),
				decimal.NewFromFloat(110),
				decimal.NewFromFloat(90),
				decimal.NewFromFloat(95),
			},
			(110.0 - 90.0) / 110.0, // 约 18.18%
		},
		{
			"无回撤",
			[]decimal.Decimal{
				decimal.NewFromFloat(100),
				decimal.NewFromFloat(110),
				decimal.NewFromFloat(120),
			},
			0,
		},
		{
			"空价格",
			[]decimal.Decimal{},
			0,
		},
		{
			"单个价格",
			[]decimal.Decimal{decimal.NewFromFloat(100)},
			0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.calculateMaxDrawdown(tt.prices)
			assert.InDelta(suite.T(), tt.expected, result, 1e-10)
		})
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateMaxDrawdownFromValues() {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			"有回撤",
			[]float64{100, 110, 90, 95},
			(110.0 - 90.0) / 110.0,
		},
		{
			"无回撤",
			[]float64{100, 110, 120},
			0,
		},
		{
			"空净值",
			[]float64{},
			0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.calculateMaxDrawdownFromValues(tt.values)
			assert.InDelta(suite.T(), tt.expected, result, 1e-10)
		})
	}
}

// ========== 滚动窗口测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateRollingWindowMetrics() {
	// 生成足够的测试数据
	n := 60
	returns := make([]float64, n)
	prices := make([]decimal.Decimal, n+1)
	price := 100.0
	prices[0] = decimal.NewFromFloat(price)

	for i := 0; i < n; i++ {
		returns[i] = 0.001 // 0.1% 日收益
		price *= (1 + returns[i])
		prices[i+1] = decimal.NewFromFloat(price)
	}

	// 测试 30 日窗口
	metrics30 := suite.service.CalculateRollingWindowMetrics(returns, prices, 30)
	assert.NotNil(suite.T(), metrics30)
	assert.Equal(suite.T(), 30, metrics30.WindowDays)
	assert.NotZero(suite.T(), metrics30.AnnualReturn)
	assert.NotZero(suite.T(), metrics30.Volatility)

	// 测试数据不足的情况
	metrics100 := suite.service.CalculateRollingWindowMetrics(returns, prices, 100)
	assert.Nil(suite.T(), metrics100)

	// 测试窗口太小
	metricsSmall := suite.service.CalculateRollingWindowMetrics(returns, prices, 10)
	assert.Nil(suite.T(), metricsSmall)
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateAllRollingWindows() {
	// 生成足够的测试数据
	n := 300
	returns := make([]float64, n)
	prices := make([]decimal.Decimal, n+1)
	price := 100.0
	prices[0] = decimal.NewFromFloat(price)

	for i := 0; i < n; i++ {
		returns[i] = 0.001 * float64(i%3-1) // 有正有负
		price *= (1 + returns[i])
		prices[i+1] = decimal.NewFromFloat(price)
	}

	result := suite.service.CalculateAllRollingWindows(returns, prices)

	// 应该有 30, 60, 90, 180, 252 五个窗口
	assert.Equal(suite.T(), 5, len(result))
	assert.Contains(suite.T(), result, 30)
	assert.Contains(suite.T(), result, 60)
	assert.Contains(suite.T(), result, 90)
	assert.Contains(suite.T(), result, 180)
	assert.Contains(suite.T(), result, 252)
}

// ========== ETF 指标测试 (需要数据库) ==========

func (suite *PortfolioAnalyticsTestSuite) insertTestData(symbol string, days int) {
	now := time.Now()
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -days+i)
		price := 100.0 + float64(i)*0.5
		suite.db.Create(&models.ETFData{
			Symbol:     symbol,
			Date:       date,
			ClosePrice: decimal.NewFromFloat(price),
		})
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestGetETFHistoricalData() {
	// 插入测试数据
	suite.insertTestData("TEST", 100)

	// 测试正常获取
	data, err := suite.service.GetETFHistoricalData("TEST", 50)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(data), 30)

	// 测试数据不足
	_, err = suite.service.GetETFHistoricalData("NONEXIST", 50)
	assert.Error(suite.T(), err)
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateETFMetrics() {
	// 插入测试数据
	suite.insertTestData("SPY", 252)

	metrics, err := suite.service.CalculateETFMetrics("SPY", 252)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), metrics)
	assert.Equal(suite.T(), "SPY", metrics.Symbol)
	assert.NotZero(suite.T(), metrics.AnnualReturn)
	assert.NotZero(suite.T(), metrics.Volatility)
	assert.NotZero(suite.T(), metrics.SharpeRatio)
	assert.NotNil(suite.T(), metrics.RollingMetrics)
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateETFMetrics_InsufficientData() {
	// 插入少量数据
	suite.insertTestData("SHORT", 20)

	_, err := suite.service.CalculateETFMetrics("SHORT", 252)
	assert.Error(suite.T(), err)
}

// ========== 相关系数测试 (需要数据库) ==========

func (suite *PortfolioAnalyticsTestSuite) TestCalculateCorrelation() {
	// 插入两只 ETF 的测试数据
	now := time.Now()
	for i := 0; i < 100; i++ {
		date := now.AddDate(0, 0, -100+i)
		suite.db.Create(&models.ETFData{
			Symbol:     "ETF_A",
			Date:       date,
			ClosePrice: decimal.NewFromFloat(100 + float64(i)*0.5),
		})
		suite.db.Create(&models.ETFData{
			Symbol:     "ETF_B",
			Date:       date,
			ClosePrice: decimal.NewFromFloat(200 + float64(i)*0.3),
		})
	}

	corr, err := suite.service.CalculateCorrelation("ETF_A", "ETF_B", 50)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), corr, -1.0)
	assert.LessOrEqual(suite.T(), corr, 1.0)
}

// ========== 组合分析测试 (需要数据库) ==========

func (suite *PortfolioAnalyticsTestSuite) TestAnalyzePortfolio() {
	// 插入多只 ETF 的测试数据
	symbols := []string{"PORT_A", "PORT_B", "PORT_C"}
	now := time.Now()

	for _, symbol := range symbols {
		for i := 0; i < 100; i++ {
			date := now.AddDate(0, 0, -100+i)
			price := 100.0 + float64(i)*0.5 + float64(len(symbol))*10
			suite.db.Create(&models.ETFData{
				Symbol:     symbol,
				Date:       date,
				ClosePrice: decimal.NewFromFloat(price),
			})
		}
	}

	portfolio := map[string]float64{
		"PORT_A": 0.4,
		"PORT_B": 0.3,
		"PORT_C": 0.3,
	}

	result, err := suite.service.AnalyzePortfolio(portfolio, 50)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotZero(suite.T(), result.ExpectedReturn)
	assert.NotZero(suite.T(), result.Volatility)
	assert.NotNil(suite.T(), result.CorrelationMatrix)
	assert.NotNil(suite.T(), result.ETFMetrics)
}

// ========== 边界条件测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestGetETFHistoricalData_NilDB() {
	service := &PortfolioAnalyticsService{db: nil}
	_, err := service.GetETFHistoricalData("TEST", 50)
	assert.Error(suite.T(), err)
}

func (suite *PortfolioAnalyticsTestSuite) TestCalculateLogReturns_AllZeroPrices() {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(0),
	}

	result := suite.service.CalculateLogReturns(prices)
	assert.Equal(suite.T(), 2, len(result))
	for _, r := range result {
		assert.Equal(suite.T(), float64(0), r)
	}
}

func (suite *PortfolioAnalyticsTestSuite) TestVaR_ConfidenceLevels() {
	returns := []float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008}

	var90 := suite.service.CalculateVaR(returns, 0.90)
	var95 := suite.service.CalculateVaR(returns, 0.95)
	var99 := suite.service.CalculateVaR(returns, 0.99)

	// 更高的置信度应该有更负的 VaR
	assert.Less(suite.T(), var99, var95)
	assert.Less(suite.T(), var95, var90)
}

func (suite *PortfolioAnalyticsTestSuite) TestCVaR_ConfidenceLevels() {
	returns := []float64{0.01, -0.02, 0.015, -0.01, 0.005, -0.005, 0.02, -0.015, 0.01, -0.008}

	cvar90 := suite.service.CalculateCVaR(returns, 0.90)
	cvar95 := suite.service.CalculateCVaR(returns, 0.95)
	cvar99 := suite.service.CalculateCVaR(returns, 0.99)

	// 更高的置信度应该有更负的 CVaR
	assert.Less(suite.T(), cvar99, cvar95)
	assert.Less(suite.T(), cvar95, cvar90)
}

// ========== 辅助函数测试 ==========

func (suite *PortfolioAnalyticsTestSuite) TestInverseNormalCDF() {
	tests := []struct {
		name     string
		p        float64
		expected float64
	}{
		{"95%置信度", 0.95, 1.645},
		{"99%置信度", 0.99, 2.326},
		{"50%置信度", 0.5, 0},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := inverseNormalCDF(tt.p)
			assert.InDelta(suite.T(), tt.expected, result, 0.01)
		})
	}

	// 边界条件
	assert.True(suite.T(), math.IsInf(inverseNormalCDF(0), -1))
	assert.True(suite.T(), math.IsInf(inverseNormalCDF(1), 1))
}

func (suite *PortfolioAnalyticsTestSuite) TestFindCommonDates() {
	now := time.Now()

	etfData := map[string][]models.ETFData{
		"A": {
			{Date: now, ClosePrice: decimal.NewFromFloat(100)},
			{Date: now.AddDate(0, 0, -1), ClosePrice: decimal.NewFromFloat(101)},
			{Date: now.AddDate(0, 0, -2), ClosePrice: decimal.NewFromFloat(102)},
		},
		"B": {
			{Date: now, ClosePrice: decimal.NewFromFloat(200)},
			{Date: now.AddDate(0, 0, -1), ClosePrice: decimal.NewFromFloat(201)},
		},
	}

	dates := suite.service.findCommonDates(etfData)
	assert.Equal(suite.T(), 2, len(dates)) // 只有2个共同日期
}

func (suite *PortfolioAnalyticsTestSuite) TestFindCommonDates_Empty() {
	dates := suite.service.findCommonDates(map[string][]models.ETFData{})
	assert.Equal(suite.T(), 0, len(dates))
}

// ========== 基准测试 ==========

func BenchmarkCalculateVaR(b *testing.B) {
	service := &PortfolioAnalyticsService{}
	returns := make([]float64, 252)
	for i := range returns {
		returns[i] = 0.001 * float64(i%10-5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateVaR(returns, 0.95)
	}
}

func BenchmarkCalculateCVaR(b *testing.B) {
	service := &PortfolioAnalyticsService{}
	returns := make([]float64, 252)
	for i := range returns {
		returns[i] = 0.001 * float64(i%10-5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateCVaR(returns, 0.95)
	}
}
