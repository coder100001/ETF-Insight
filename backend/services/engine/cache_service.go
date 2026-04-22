package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// ================================
// 缓存服务
// ================================

// CacheKey 缓存键类型
type CacheKey string

// CacheValue 缓存值类型
type CacheValue interface{}

// CacheService 缓存服务接口
type CacheService interface {
	Engine
	
	// 基本操作
	Get(ctx context.Context, key CacheKey) (CacheValue, error)
	Set(ctx context.Context, key CacheKey, value CacheValue, ttl time.Duration) error
	Delete(ctx context.Context, key CacheKey) error
	Exists(ctx context.Context, key CacheKey) (bool, error)
	Clear(ctx context.Context) error

	// 批量操作
	MGet(ctx context.Context, keys []CacheKey) (map[CacheKey]CacheValue, error)
	MSet(ctx context.Context, values map[CacheKey]CacheValue, ttl time.Duration) error
	MDelete(ctx context.Context, keys []CacheKey) error

	// 统计信息
	Stats(ctx context.Context) (*CacheStats, error)
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits          int64     // 命中次数
	Misses        int64     // 未命中次数
	HitRate       float64   // 命中率
	TotalItems    int64     // 总缓存项数
	TotalSize     int64     // 总缓存大小（字节）
	Evictions     int64     // 驱逐次数
	MemoryUsage   float64   // 内存使用率
	AvgLoadTimeMs float64   // 平均加载时间
}

// ================================
// 物化视图管理器
// ================================

// MaterializedView 物化视图接口
type MaterializedView interface {
	Name() string                    // 视图名称
	Description() string             // 视图描述
	Schema() string                  // 视图模式
	Refresh(ctx context.Context) error // 刷新视图
	IsValid() bool                   // 视图是否有效
	LastRefreshed() time.Time        // 最后刷新时间
	NextRefresh() time.Time          // 下次刷新时间
	Size() int64                     // 视图大小
}

// MaterializedViewManager 物化视图管理器
type MaterializedViewManager interface {
	Engine
	
	// 视图管理
	CreateView(ctx context.Context, view MaterializedView) error
	UpdateView(ctx context.Context, view MaterializedView) error
	DeleteView(ctx context.Context, name string) error
	GetView(ctx context.Context, name string) (MaterializedView, error)
	ListViews(ctx context.Context) ([]MaterializedView, error)

	// 视图操作
	RefreshView(ctx context.Context, name string) error
	RefreshAllViews(ctx context.Context) error
	ScheduleRefresh(ctx context.Context, name string, schedule string) error

	// 视图状态
	GetViewStats(ctx context.Context, name string) (*ViewStats, error)
	GetViewData(ctx context.Context, name string, query string) (interface{}, error)
}

// ViewStats 视图统计
type ViewStats struct {
	Name          string        // 视图名称
	Rows          int64         // 行数
	SizeBytes     int64         // 大小（字节）
	LastRefresh   time.Time     // 最后刷新时间
	RefreshTimeMs int64         // 刷新耗时（毫秒）
	IsValid       bool          // 是否有效
	Error         string        // 错误信息
}

// ================================
// 常用物化视图定义
// ================================

// ETFOverlapView ETF重叠度物化视图
type ETFOverlapView struct {
	ETF1ID       uint            `json:"etf1_id"`
	ETF2ID       uint            `json:"etf2_id"`
	ETF1Symbol   string          `json:"etf1_symbol"`
	ETF2Symbol   string          `json:"etf2_symbol"`
	AsOfDate     time.Time       `json:"as_of_date"`
	TotalOverlap decimal.Decimal `json:"total_overlap"`
	WeightedOverlap decimal.Decimal `json:"weighted_overlap"`
	JaccardIndex decimal.Decimal `json:"jaccard_index"`
	SectorOverlap decimal.Decimal `json:"sector_overlap"`
	CountryOverlap decimal.Decimal `json:"country_overlap"`
	CommonHoldings int           `json:"common_holdings"`
	CreatedAt    time.Time       `json:"created_at"`
}

// PortfolioExposureView 组合暴露度物化视图
type PortfolioExposureView struct {
	PortfolioID  uint            `json:"portfolio_id"`
	AsOfDate     time.Time       `json:"as_of_date"`
	Sector       string          `json:"sector"`
	Weight       decimal.Decimal `json:"weight"`
	MarketValue  decimal.Decimal `json:"market_value"`
	NumAssets    int             `json:"num_assets"`
	CreatedAt    time.Time       `json:"created_at"`
}

// FactorExposureView 因子暴露度物化视图
type FactorExposureView struct {
	PortfolioID  uint            `json:"portfolio_id"`
	FactorName   string          `json:"factor_name"`
	AsOfDate     time.Time       `json:"as_of_date"`
	Exposure     decimal.Decimal `json:"exposure"`
	RiskContribution decimal.Decimal `json:"risk_contribution"`
	ActiveExposure decimal.Decimal `json:"active_exposure"`
	CreatedAt    time.Time       `json:"created_at"`
}

// PriceStatsView 价格统计物化视图
type PriceStatsView struct {
	AssetID     uint            `json:"asset_id"`
	Symbol      string          `json:"symbol"`
	Period      string          `json:"period"` // daily/weekly/monthly
	AsOfDate    time.Time       `json:"as_of_date"`
	Return      decimal.Decimal `json:"return"`
	Volatility  decimal.Decimal `json:"volatility"`
	SharpeRatio decimal.Decimal `json:"sharpe_ratio"`
	MaxDrawdown decimal.Decimal `json:"max_drawdown"`
	Correlations string         `json:"correlations"` // JSON格式
	CreatedAt   time.Time       `json:"created_at"`
}

// ================================
// 缓存管理器实现
// ================================

// InMemoryCacheService 内存缓存服务
type InMemoryCacheService struct {
	cache     sync.Map
	stats     CacheStats
	statsLock sync.RWMutex
}

// NewInMemoryCacheService 创建内存缓存服务
func NewInMemoryCacheService() *InMemoryCacheService {
	return &InMemoryCacheService{
		cache: sync.Map{},
		stats: CacheStats{},
	}
}

// Name 缓存服务名称
func (c *InMemoryCacheService) Name() string {
	return "InMemoryCacheService"
}

// Description 描述
func (c *InMemoryCacheService) Description() string {
	return "基于sync.Map的内存缓存服务"
}

// Version 版本
func (c *InMemoryCacheService) Version() string {
	return "1.0.0"
}

// IsAvailable 是否可用
func (c *InMemoryCacheService) IsAvailable() bool {
	return true
}

// HealthCheck 健康检查
func (c *InMemoryCacheService) HealthCheck() error {
	return nil
}

// Get 获取缓存值
func (c *InMemoryCacheService) Get(ctx context.Context, key CacheKey) (CacheValue, error) {
	c.updateStats(false)

	value, ok := c.cache.Load(key)
	if !ok {
		c.updateStats(true)
		return nil, fmt.Errorf("cache key not found: %s", key)
	}

	cacheItem, ok := value.(*cacheItem)
	if !ok {
		c.deleteKey(key)
		c.updateStats(true)
		return nil, fmt.Errorf("invalid cache item: %s", key)
	}

	// 检查是否过期
	if cacheItem.expiresAt.Before(time.Now()) {
		c.deleteKey(key)
		c.updateStats(true)
		return nil, fmt.Errorf("cache expired: %s", key)
	}

	return cacheItem.value, nil
}

// Set 设置缓存值
func (c *InMemoryCacheService) Set(ctx context.Context, key CacheKey, value CacheValue, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	cacheItem := &cacheItem{
		value:     value,
		expiresAt: expiresAt,
		createdAt: time.Now(),
	}
	c.cache.Store(key, cacheItem)
	return nil
}

// Delete 删除缓存值
func (c *InMemoryCacheService) Delete(ctx context.Context, key CacheKey) error {
	c.cache.Delete(key)
	return nil
}

// Exists 检查缓存是否存在
func (c *InMemoryCacheService) Exists(ctx context.Context, key CacheKey) (bool, error) {
	value, ok := c.cache.Load(key)
	if !ok {
		return false, nil
	}

	cacheItem, ok := value.(*cacheItem)
	if !ok {
		return false, nil
	}

	// 检查是否过期
	if cacheItem.expiresAt.Before(time.Now()) {
		c.deleteKey(key)
		return false, nil
	}

	return true, nil
}

// Clear 清空缓存
func (c *InMemoryCacheService) Clear(ctx context.Context) error {
	c.cache = sync.Map{}
	c.resetStats()
	return nil
}

// MGet 批量获取
func (c *InMemoryCacheService) MGet(ctx context.Context, keys []CacheKey) (map[CacheKey]CacheValue, error) {
	result := make(map[CacheKey]CacheValue)
	for _, key := range keys {
		if value, err := c.Get(ctx, key); err == nil {
			result[key] = value
		}
	}
	return result, nil
}

// MSet 批量设置
func (c *InMemoryCacheService) MSet(ctx context.Context, values map[CacheKey]CacheValue, ttl time.Duration) error {
	for key, value := range values {
		if err := c.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// MDelete 批量删除
func (c *InMemoryCacheService) MDelete(ctx context.Context, keys []CacheKey) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Stats 获取统计信息
func (c *InMemoryCacheService) Stats(ctx context.Context) (*CacheStats, error) {
	c.statsLock.RLock()
	defer c.statsLock.RUnlock()

	// 计算命中率
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	} else {
		c.stats.HitRate = 0
	}

	// 计算缓存项数
	var count int64
	c.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	c.stats.TotalItems = count

	stats := c.stats
	return &stats, nil
}

// cacheItem 缓存项
type cacheItem struct {
	value     CacheValue
	expiresAt time.Time
	createdAt time.Time
}

// deleteKey 删除缓存键
func (c *InMemoryCacheService) deleteKey(key CacheKey) {
	c.cache.Delete(key)
}

// updateStats 更新统计信息
func (c *InMemoryCacheService) updateStats(isMiss bool) {
	c.statsLock.Lock()
	defer c.statsLock.Unlock()

	if isMiss {
		c.stats.Misses++
	} else {
		c.stats.Hits++
	}
}

// resetStats 重置统计信息
func (c *InMemoryCacheService) resetStats() {
	c.statsLock.Lock()
	defer c.statsLock.Unlock()

	c.stats = CacheStats{}
}

// ================================
// 缓存键生成器
// ================================

// CacheKeyGenerator 缓存键生成器
type CacheKeyGenerator struct {
	prefix string
}

// NewCacheKeyGenerator 创建缓存键生成器
func NewCacheKeyGenerator(prefix string) *CacheKeyGenerator {
	return &CacheKeyGenerator{prefix: prefix}
}

// GenerateKey 生成缓存键
func (g *CacheKeyGenerator) GenerateKey(parts ...string) CacheKey {
	key := g.prefix
	for _, part := range parts {
		key += ":" + part
	}
	return CacheKey(key)
}

// 常用缓存键模式
var (
	// 资产相关
	AssetKeyPrefix       = "asset"
	PriceKeyPrefix       = "price"
	HoldingsKeyPrefix    = "holdings"
	
	// 组合相关
	PortfolioKeyPrefix   = "portfolio"
	PositionKeyPrefix    = "position"
	PerformanceKeyPrefix = "performance"
	
	// 分析相关
	AnalysisKeyPrefix    = "analysis"
	OverlapKeyPrefix     = "overlap"
	ExposureKeyPrefix    = "exposure"
	FactorKeyPrefix      = "factor"
	
	// 计算相关
	CalculationKeyPrefix = "calc"
	StatsKeyPrefix       = "stats"
)

// GetAssetKey 获取资产缓存键
func GetAssetKey(symbol string) CacheKey {
	return CacheKey(fmt.Sprintf("%s:%s", AssetKeyPrefix, symbol))
}

// GetPriceKey 获取价格缓存键
func GetPriceKey(symbol string, date time.Time) CacheKey {
	dateStr := date.Format("2006-01-02")
	return CacheKey(fmt.Sprintf("%s:%s:%s", PriceKeyPrefix, symbol, dateStr))
}

// GetHoldingsKey 获取持仓缓存键
func GetHoldingsKey(parentID uint, date time.Time) CacheKey {
	dateStr := date.Format("2006-01-02")
	return CacheKey(fmt.Sprintf("%s:%d:%s", HoldingsKeyPrefix, parentID, dateStr))
}

// GetPortfolioKey 获取组合缓存键
func GetPortfolioKey(portfolioID uint) CacheKey {
	return CacheKey(fmt.Sprintf("%s:%d", PortfolioKeyPrefix, portfolioID))
}

// GetPositionKey 获取持仓缓存键
func GetPositionKey(portfolioID, assetID uint) CacheKey {
	return CacheKey(fmt.Sprintf("%s:%d:%d", PositionKeyPrefix, portfolioID, assetID))
}

// GetOverlapKey 获取重叠度缓存键
func GetOverlapKey(etf1ID, etf2ID uint, date time.Time) CacheKey {
	dateStr := date.Format("2006-01-02")
	return CacheKey(fmt.Sprintf("%s:%d:%d:%s", OverlapKeyPrefix, etf1ID, etf2ID, dateStr))
}

// GetExposureKey 获取暴露度缓存键
func GetExposureKey(portfolioID uint, dimension string, date time.Time) CacheKey {
	dateStr := date.Format("2006-01-02")
	return CacheKey(fmt.Sprintf("%s:%d:%s:%s", ExposureKeyPrefix, portfolioID, dimension, dateStr))
}

// GetFactorKey 获取因子缓存键
func GetFactorKey(portfolioID uint, factorName string, date time.Time) CacheKey {
	dateStr := date.Format("2006-01-02")
	return CacheKey(fmt.Sprintf("%s:%d:%s:%s", FactorKeyPrefix, portfolioID, factorName, dateStr))
}

// GetCalculationKey 获取计算缓存键
func GetCalculationKey(calcType string, params string) CacheKey {
	return CacheKey(fmt.Sprintf("%s:%s:%s", CalculationKeyPrefix, calcType, params))
}

// GetStatsKey 获取统计缓存键
func GetStatsKey(assetID uint, period string, date time.Time) CacheKey {
	dateStr := date.Format("2006-01-02")
	return CacheKey(fmt.Sprintf("%s:%d:%s:%s", StatsKeyPrefix, assetID, period, dateStr))
}