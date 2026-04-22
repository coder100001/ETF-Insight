package etf

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

	// 自动迁移表结构
	err = db.AutoMigrate(
		&models.Asset{},
		&models.ETFHolding{},
	)
	require.NoError(t, err)

	return db
}

// createTestETF 创建测试ETF数据
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

// createTestHoldings 创建测试持仓数据
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

func TestCalculateOverlap(t *testing.T) {
	db := setupTestDB(t)
	service := NewHoldingsService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 创建测试ETF
	etfA, err := createTestETF(db, "QQQ", "Invesco QQQ Trust")
	require.NoError(t, err)

	etfB, err := createTestETF(db, "VGT", "Vanguard Info Tech ETF")
	require.NoError(t, err)

	// 创建ETF A持仓 (QQQ风格)
	holdingsA := []struct {
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
		{"META", "Meta Platforms Inc.", 4.2, testDate},
		{"TSLA", "Tesla Inc.", 3.5, testDate},
		{"AVGO", "Broadcom Inc.", 3.1, testDate},
		{"PEP", "PepsiCo Inc.", 2.8, testDate},
		{"COST", "Costco Wholesale", 2.5, testDate},
	}
	err = createTestHoldings(db, etfA.ID, holdingsA)
	require.NoError(t, err)

	// 创建ETF B持仓 (VGT风格 - 科技主题，与QQQ有大量重叠)
	holdingsB := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 15.8, testDate},
		{"MSFT", "Microsoft Corp.", 14.2, testDate},
		{"NVDA", "NVIDIA Corp.", 9.5, testDate},
		{"AVGO", "Broadcom Inc.", 5.1, testDate},
		{"CRM", "Salesforce Inc.", 4.8, testDate},
		{"AMD", "AMD Inc.", 4.3, testDate},
		{"GOOGL", "Alphabet Inc.", 4.0, testDate},
		{"INTC", "Intel Corp.", 3.9, testDate},
		{"ORCL", "Oracle Corp.", 3.7, testDate},
		{"ACN", "Accenture PLC", 3.2, testDate},
	}
	err = createTestHoldings(db, etfB.ID, holdingsB)
	require.NoError(t, err)

	t.Run("calculate overlap between QQQ and VGT", func(t *testing.T) {
		result, err := service.CalculateOverlap(ctx, "QQQ", "VGT", testDate)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "QQQ", result.ETFA)
		assert.Equal(t, "VGT", result.ETFB)

		// 验证共同持仓数量：AAPL, MSFT, GOOGL, NVDA, AVGO = 5只
		assert.Equal(t, 5, result.CommonHoldings)

		// 验证重叠度分数在合理范围内 (0-100)
		assert.True(t, result.OverlapScore.GreaterThanOrEqual(decimal.Zero))
		assert.True(t, result.OverlapScore.LessThanOrEqual(decimal.NewFromInt(100)))

		// 验证重叠度大于0（因为有共同持仓）
		assert.True(t, result.OverlapScore.GreaterThan(decimal.Zero))

		// 验证明细
		assert.Len(t, result.Details, 5)

		// 验证A中的重叠权重
		expectedWeightA := decimal.NewFromFloat(12.5 + 11.2 + 8.7 + 6.8 + 3.1)
		assert.True(t, result.TotalWeightA.Equal(expectedWeightA))

		// 验证B中的重叠权重
		expectedWeightB := decimal.NewFromFloat(15.8 + 14.2 + 4.0 + 9.5 + 5.1)
		assert.True(t, result.TotalWeightB.Equal(expectedWeightB))

		t.Logf("Overlap Score: %s%%", result.OverlapScore.String())
		t.Logf("Common Holdings: %d", result.CommonHoldings)
	})

	t.Run("calculate overlap with empty symbol", func(t *testing.T) {
		_, err := service.CalculateOverlap(ctx, "", "VGT", testDate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("calculate overlap with same ETF", func(t *testing.T) {
		_, err := service.CalculateOverlap(ctx, "QQQ", "QQQ", testDate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "same ETF")
	})
}

func TestCalculateOverlap_NoCommonHoldings(t *testing.T) {
	db := setupTestDB(t)
	service := NewHoldingsService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 创建两个完全不同的ETF
	etfA, err := createTestETF(db, "BND", "Vanguard Total Bond Market")
	require.NoError(t, err)

	etfB, err := createTestETF(db, "GLD", "SPDR Gold Shares")
	require.NoError(t, err)

	// ETF A - 债券
	holdingsA := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"GOVT", "iShares US Treasury", 25.0, testDate},
		{"LQD", "iShares iBoxx IG Corp", 20.0, testDate},
		{"HYG", "iShares High Yield", 15.0, testDate},
	}
	err = createTestHoldings(db, etfA.ID, holdingsA)
	require.NoError(t, err)

	// ETF B - 黄金
	holdingsB := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"GLD", "SPDR Gold Shares", 100.0, testDate},
	}
	err = createTestHoldings(db, etfB.ID, holdingsB)
	require.NoError(t, err)

	result, err := service.CalculateOverlap(ctx, "BND", "GLD", testDate)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// 没有共同持仓，重叠度应该为0
	assert.Equal(t, 0, result.CommonHoldings)
	assert.True(t, result.OverlapScore.Equal(decimal.Zero))
	assert.Len(t, result.Details, 0)
}

func TestCalculateOverlap_IdenticalHoldings(t *testing.T) {
	db := setupTestDB(t)
	service := NewHoldingsService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	etfA, err := createTestETF(db, "VOO", "Vanguard S&P 500")
	require.NoError(t, err)

	etfB, err := createTestETF(db, "SPY", "SPDR S&P 500")
	require.NoError(t, err)

	// 完全相同的持仓（权重不同）
	holdings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 7.0, testDate},
		{"MSFT", "Microsoft Corp.", 6.5, testDate},
		{"AMZN", "Amazon.com Inc.", 3.5, testDate},
	}

	err = createTestHoldings(db, etfA.ID, holdings)
	require.NoError(t, err)

	// B的权重略有不同
	holdingsB := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 7.2, testDate},
		{"MSFT", "Microsoft Corp.", 6.3, testDate},
		{"AMZN", "Amazon.com Inc.", 3.6, testDate},
	}
	err = createTestHoldings(db, etfB.ID, holdingsB)
	require.NoError(t, err)

	result, err := service.CalculateOverlap(ctx, "VOO", "SPY", testDate)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// 3只共同持仓
	assert.Equal(t, 3, result.CommonHoldings)

	// 重叠度应该很高（接近100%）
	assert.True(t, result.OverlapScore.GreaterThan(decimal.NewFromInt(80)))
	assert.True(t, result.OverlapScore.LessThanOrEqual(decimal.NewFromInt(100)))
}

func TestGetETFHoldings(t *testing.T) {
	db := setupTestDB(t)
	service := NewHoldingsService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	etf, err := createTestETF(db, "QQQ", "Invesco QQQ Trust")
	require.NoError(t, err)

	holdings := []struct {
		Symbol string
		Name   string
		Weight float64
		Date   time.Time
	}{
		{"AAPL", "Apple Inc.", 12.5, testDate},
		{"MSFT", "Microsoft Corp.", 11.2, testDate},
		{"GOOGL", "Alphabet Inc.", 8.7, testDate},
	}
	err = createTestHoldings(db, etf.ID, holdings)
	require.NoError(t, err)

	t.Run("get holdings with specific date", func(t *testing.T) {
		result, err := service.GetETFHoldings(ctx, "QQQ", testDate)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("get latest holdings", func(t *testing.T) {
		result, err := service.GetETFHoldings(ctx, "QQQ", time.Time{})
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("get holdings for non-existent ETF", func(t *testing.T) {
		_, err := service.GetETFHoldings(ctx, "NONEXISTENT", testDate)
		assert.Error(t, err)
	})
}

func TestGetTopHoldings(t *testing.T) {
	db := setupTestDB(t)
	service := NewHoldingsService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	etf, err := createTestETF(db, "QQQ", "Invesco QQQ Trust")
	require.NoError(t, err)

	holdings := []struct {
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
	err = createTestHoldings(db, etf.ID, holdings)
	require.NoError(t, err)

	t.Run("get top 3 holdings", func(t *testing.T) {
		result, err := service.GetTopHoldings(ctx, "QQQ", 3, testDate)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "AAPL", result[0].Symbol)
		assert.Equal(t, "MSFT", result[1].Symbol)
		assert.Equal(t, "GOOGL", result[2].Symbol)
	})

	t.Run("get all holdings when n > count", func(t *testing.T) {
		result, err := service.GetTopHoldings(ctx, "QQQ", 100, testDate)
		require.NoError(t, err)
		assert.Len(t, result, 5)
	})
}

func TestCalculateOverlapBatch(t *testing.T) {
	db := setupTestDB(t)
	service := NewHoldingsService(db)
	ctx := context.Background()
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 创建3个ETF
	etfA, err := createTestETF(db, "ETF1", "Test ETF 1")
	require.NoError(t, err)

	etfB, err := createTestETF(db, "ETF2", "Test ETF 2")
	require.NoError(t, err)

	etfC, err := createTestETF(db, "ETF3", "Test ETF 3")
	require.NoError(t, err)

	// 为每个ETF创建持仓
	for _, etf := range []*models.Asset{etfA, etfB, etfC} {
		holdings := []struct {
			Symbol string
			Name   string
			Weight float64
			Date   time.Time
		}{
			{"AAPL", "Apple Inc.", 10.0, testDate},
			{"MSFT", "Microsoft Corp.", 10.0, testDate},
		}
		err = createTestHoldings(db, etf.ID, holdings)
		require.NoError(t, err)
	}

	results, err := service.CalculateOverlapBatch(ctx, []string{"ETF1", "ETF2", "ETF3"}, testDate)
	require.NoError(t, err)

	// C(3,2) = 3 对组合
	assert.Len(t, results, 3)
	assert.Contains(t, results, "ETF1-ETF2")
	assert.Contains(t, results, "ETF1-ETF3")
	assert.Contains(t, results, "ETF2-ETF3")
}
