package portfolio

import (
	"context"
	"testing"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Asset{},
		&models.ETFHolding{},
	)
	require.NoError(t, err)

	return db
}

// createTestETF 创建测试ETF
func createTestETF(db *gorm.DB, symbol, name string) (*models.Asset, error) {
	asset := &models.Asset{
		Symbol:   symbol,
		Name:     name,
		Type:     models.AssetTypeETF,
		Currency: "USD",
		Exchange: "NASDAQ",
		Status:   1,
	}
	if err := db.Create(asset).Error; err != nil {
		return nil, err
	}
	return asset, nil
}

// createTestStock 创建测试股票（作为底层资产）
func createTestStock(db *gorm.DB, symbol, name, sector, country string) (*models.Asset, error) {
	asset := &models.Asset{
		Symbol:   symbol,
		Name:     name,
		Type:     models.AssetTypeStock,
		Currency: "USD",
		Exchange: "NASDAQ",
		Sector:   sector,
		Country:  country,
		Status:   1,
	}
	if err := db.Create(asset).Error; err != nil {
		return nil, err
	}
	return asset, nil
}

// createTestHoldings 创建测试持仓
func createTestHoldings(db *gorm.DB, etfID uint, holdings []struct {
	Symbol string
	Name   string
	Weight float64
	Date   time.Time
}) error {
	for _, h := range holdings {
		holding := models.ETFHolding{
			ETFID:  etfID,
			Symbol: h.Symbol,
			Name:   h.Name,
			Weight: decimal.NewFromFloat(h.Weight),
			Date:   h.Date,
		}
		if err := db.Create(&holding).Error; err != nil {
			return err
		}
	}
	return nil
}

func TestAnalyzePortfolio(t *testing.T) {
	db := setupTestDB(t)
	service := NewPortfolioPenetrationService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 创建ETF
	qqq, err := createTestETF(db, "QQQ", "Invesco QQQ Trust")
	require.NoError(t, err)

	vgt, err := createTestETF(db, "VGT", "Vanguard Info Tech ETF")
	require.NoError(t, err)

	// 创建底层资产（股票）
	_, err = createTestStock(db, "AAPL", "Apple Inc.", "Technology", "US")
	require.NoError(t, err)
	_, err = createTestStock(db, "MSFT", "Microsoft Corp.", "Technology", "US")
	require.NoError(t, err)
	_, err = createTestStock(db, "GOOGL", "Alphabet Inc.", "Technology", "US")
	require.NoError(t, err)
	_, err = createTestStock(db, "AMZN", "Amazon.com Inc.", "Consumer Discretionary", "US")
	require.NoError(t, err)
	_, err = createTestStock(db, "NVDA", "NVIDIA Corp.", "Technology", "US")
	require.NoError(t, err)
	_, err = createTestStock(db, "TSM", "TSMC", "Technology", "Taiwan")
	require.NoError(t, err)

	// QQQ持仓
	qqqHoldings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 12.5, testDate},
		{"MSFT", "Microsoft Corp.", 11.2, testDate},
		{"GOOGL", "Alphabet Inc.", 8.7, testDate},
		{"AMZN", "Amazon.com Inc.", 7.3, testDate},
		{"NVDA", "NVIDIA Corp.", 6.8, testDate},
	}
	err = createTestHoldings(db, qqq.ID, qqqHoldings)
	require.NoError(t, err)

	// VGT持仓
	vgtHoldings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 15.8, testDate},
		{"MSFT", "Microsoft Corp.", 14.2, testDate},
		{"NVDA", "NVIDIA Corp.", 9.5, testDate},
		{"TSM", "TSMC", 5.1, testDate},
	}
	err = createTestHoldings(db, vgt.ID, vgtHoldings)
	require.NoError(t, err)

	t.Run("analyze portfolio with two ETFs", func(t *testing.T) {
		portfolio := []PortfolioHolding{
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Weight: decimal.NewFromInt(60)},
			{Symbol: "VGT", Name: "Vanguard Info Tech ETF", Weight: decimal.NewFromInt(40)},
		}

		result, err := service.AnalyzePortfolio(ctx, "test_portfolio", portfolio, testDate)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// 验证基本统计
		assert.Equal(t, "test_portfolio", result.PortfolioID)
		assert.Equal(t, 2, result.TotalETFs)
		assert.Equal(t, 9, result.TotalHoldings)  // 5 + 4
		assert.Equal(t, 6, result.UniqueHoldings) // AAPL, MSFT, GOOGL, AMZN, NVDA, TSM

		// 验证行业分布
		assert.Contains(t, result.SectorAllocation, "Technology")
		assert.Contains(t, result.SectorAllocation, "Consumer Discretionary")
		assert.True(t, result.SectorAllocation["Technology"].GreaterThan(decimal.Zero))

		// 验证地理分布
		assert.Contains(t, result.CountryAllocation, "US")
		assert.Contains(t, result.CountryAllocation, "Taiwan")

		// 验证前十大持仓
		assert.Len(t, result.TopHoldings, 6) // 只有6只底层资产
		assert.Equal(t, "AAPL", result.TopHoldings[0].Symbol)

		// 验证集中度指标
		assert.True(t, result.Concentration.Top10Weight.GreaterThan(decimal.Zero))
		assert.True(t, result.Concentration.HerfindahlIndex.GreaterThan(decimal.Zero))
		assert.True(t, result.Concentration.EffectiveHoldings.GreaterThan(decimal.Zero))

		t.Logf("Unique Holdings: %d", result.UniqueHoldings)
		t.Logf("Technology Sector: %s%%", result.SectorAllocation["Technology"].String())
		t.Logf("US Country: %s%%", result.CountryAllocation["US"].String())
		t.Logf("Top 10 Weight: %s%%", result.Concentration.Top10Weight.String())
	})

	t.Run("analyze portfolio with empty holdings", func(t *testing.T) {
		_, err := service.AnalyzePortfolio(ctx, "empty", []PortfolioHolding{}, testDate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestCalculateConcentration(t *testing.T) {
	db := setupTestDB(t)
	service := NewPortfolioPenetrationService(db)

	holdings := []*UnderlyingHolding{
		{Symbol: "AAPL", Weight: decimal.NewFromFloat(25.0)},
		{Symbol: "MSFT", Weight: decimal.NewFromFloat(20.0)},
		{Symbol: "GOOGL", Weight: decimal.NewFromFloat(15.0)},
		{Symbol: "AMZN", Weight: decimal.NewFromFloat(10.0)},
		{Symbol: "NVDA", Weight: decimal.NewFromFloat(8.0)},
		{Symbol: "META", Weight: decimal.NewFromFloat(7.0)},
		{Symbol: "TSLA", Weight: decimal.NewFromFloat(5.0)},
		{Symbol: "JPM", Weight: decimal.NewFromFloat(4.0)},
		{Symbol: "JNJ", Weight: decimal.NewFromFloat(3.0)},
		{Symbol: "V", Weight: decimal.NewFromFloat(2.0)},
		{Symbol: "WMT", Weight: decimal.NewFromFloat(1.0)},
	}

	concentration := service.calculateConcentration(holdings)

	// 前10大权重应该接近99%
	expectedTop10 := decimal.NewFromFloat(25.0 + 20.0 + 15.0 + 10.0 + 8.0 + 7.0 + 5.0 + 4.0 + 3.0 + 2.0)
	assert.True(t, concentration.Top10Weight.Equal(expectedTop10))

	// 前20大权重应该为100%（因为只有11只）
	expectedTop20 := decimal.NewFromFloat(100.0)
	assert.True(t, concentration.Top20Weight.Equal(expectedTop20))

	// 赫芬达尔指数
	assert.True(t, concentration.HerfindahlIndex.GreaterThan(decimal.Zero))

	// 有效持仓数
	assert.True(t, concentration.EffectiveHoldings.GreaterThan(decimal.Zero))

	t.Logf("Top 10 Weight: %s%%", concentration.Top10Weight.String())
	t.Logf("HHI: %s", concentration.HerfindahlIndex.String())
	t.Logf("Effective Holdings: %s", concentration.EffectiveHoldings.String())
}

func TestComparePortfolios(t *testing.T) {
	db := setupTestDB(t)
	service := NewPortfolioPenetrationService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 创建ETF和资产
	qqq, err := createTestETF(db, "QQQ", "Invesco QQQ Trust")
	require.NoError(t, err)

	spy, err := createTestETF(db, "SPY", "SPDR S&P 500")
	require.NoError(t, err)

	_, err = createTestStock(db, "AAPL", "Apple Inc.", "Technology", "US")
	require.NoError(t, err)
	_, err = createTestStock(db, "MSFT", "Microsoft Corp.", "Technology", "US")
	require.NoError(t, err)

	// QQQ持仓
	qqqHoldings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 12.5, testDate},
		{"MSFT", "Microsoft Corp.", 11.2, testDate},
	}
	err = createTestHoldings(db, qqq.ID, qqqHoldings)
	require.NoError(t, err)

	// SPY持仓
	spyHoldings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 7.0, testDate},
		{"MSFT", "Microsoft Corp.", 6.5, testDate},
	}
	err = createTestHoldings(db, spy.ID, spyHoldings)
	require.NoError(t, err)

	portfolioA := []PortfolioHolding{
		{Symbol: "QQQ", Name: "Invesco QQQ Trust", Weight: decimal.NewFromInt(100)},
	}

	portfolioB := []PortfolioHolding{
		{Symbol: "SPY", Name: "SPDR S&P 500", Weight: decimal.NewFromInt(100)},
	}

	comparison, err := service.ComparePortfolios(ctx, portfolioA, portfolioB, testDate)
	require.NoError(t, err)
	assert.NotNil(t, comparison)

	// 验证共同持仓
	assert.Equal(t, 2, comparison.CommonHoldings) // AAPL, MSFT

	// 验证行业差异
	assert.Contains(t, comparison.SectorDiff, "Technology")

	t.Logf("Common Holdings: %d", comparison.CommonHoldings)
}

func TestGetSectorExposure(t *testing.T) {
	db := setupTestDB(t)
	service := NewPortfolioPenetrationService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	qqq, err := createTestETF(db, "QQQ", "Invesco QQQ Trust")
	require.NoError(t, err)

	_, err = createTestStock(db, "AAPL", "Apple Inc.", "Technology", "US")
	require.NoError(t, err)

	qqqHoldings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 100.0, testDate},
	}
	err = createTestHoldings(db, qqq.ID, qqqHoldings)
	require.NoError(t, err)

	portfolio := []PortfolioHolding{
		{Symbol: "QQQ", Name: "Invesco QQQ Trust", Weight: decimal.NewFromInt(100)},
	}

	exposure, err := service.GetSectorExposure(ctx, portfolio, "Technology", testDate)
	require.NoError(t, err)
	assert.True(t, exposure.GreaterThan(decimal.Zero))

	// 不存在的行业
	noExposure, err := service.GetSectorExposure(ctx, portfolio, "Healthcare", testDate)
	require.NoError(t, err)
	assert.True(t, noExposure.Equal(decimal.Zero))
}
