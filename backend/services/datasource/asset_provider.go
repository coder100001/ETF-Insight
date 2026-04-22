package datasource

import (
	"context"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// AssetType 资产类型（与models一致）
type AssetType string

// AssetProvider 统一资产数据源接口
// 这是数据层抽象的核心，支持多种数据源接入
type AssetProvider interface {
	// ProviderInfo 提供者信息
	Name() string                         // 数据源名称
	Description() string                  // 描述
	IsAvailable(ctx context.Context) bool // 是否可用
	RateLimit() int                       // 速率限制（每秒请求数）
	Priority() int                        // 优先级（数值越小优先级越高）

	// 资产基本信息
	GetAsset(ctx context.Context, symbol string) (*models.Asset, error)                 // 获取单个资产信息
	SearchAssets(ctx context.Context, query string, limit int) ([]*models.Asset, error) // 搜索资产
	BatchGetAssets(ctx context.Context, symbols []string) ([]*models.Asset, error)      // 批量获取资产信息

	// 价格数据
	GetPrice(ctx context.Context, symbol string, date time.Time) (*models.Price, error)                  // 获取指定日期价格
	GetPrices(ctx context.Context, symbol string, startDate, endDate time.Time) ([]*models.Price, error) // 获取时间序列价格
	GetLatestPrice(ctx context.Context, symbol string) (*models.Price, error)                            // 获取最新价格
	BatchGetLatestPrices(ctx context.Context, symbols []string) ([]*models.Price, error)                 // 批量获取最新价格

	// 持仓数据（ETF特有）
	GetHoldings(ctx context.Context, etfSymbol string, date time.Time) ([]*models.Holding, error)                     // 获取ETF持仓
	GetLatestHoldings(ctx context.Context, etfSymbol string) ([]*models.Holding, error)                               // 获取最新持仓
	GetHoldingHistory(ctx context.Context, etfSymbol string, startDate, endDate time.Time) ([]*models.Holding, error) // 获取持仓历史

	// 元数据
	GetAssetMetadata(ctx context.Context, symbol string) (*models.AssetMetadata, error)                              // 获取资产元数据
	GetDividendInfo(ctx context.Context, symbol string) (*models.ETFDividend, error)                                 // 获取分红信息
	GetCorporateActions(ctx context.Context, symbol string, startDate, endDate time.Time) ([]CorporateAction, error) // 获取公司行动
}

// CorporateAction 公司行动（拆股、分红、并购等）
type CorporateAction struct {
	Type        string          // 行动类型：dividend/split/merger/spinoff
	Symbol      string          // 资产代码
	Date        time.Time       // 生效日期
	RecordDate  time.Time       // 登记日期
	ExDate      time.Time       // 除权除息日
	PaymentDate time.Time       // 支付日期
	Details     string          // 详情描述
	Ratio       decimal.Decimal // 比例（拆股比例等）
	Amount      decimal.Decimal // 金额（分红金额等）
}

// ETFProvider ETF特定数据源接口
type ETFProvider interface {
	AssetProvider

	// ETF特定功能
	GetETFList(ctx context.Context, filter ETFListFilter) ([]*models.Asset, error)                                    // 获取ETF列表
	GetETFOverview(ctx context.Context, symbol string) (*ETFOverview, error)                                          // 获取ETF概览
	GetETFHoldingsReport(ctx context.Context, symbol string, reportDate time.Time) (*models.ETFHoldingsReport, error) // 获取持仓报告
	GetETFPerformance(ctx context.Context, symbol string, period string) (*ETFPerformance, error)                     // 获取表现数据
}

// ETFListFilter ETF列表筛选条件
type ETFListFilter struct {
	AssetClass      models.AssetClass
	Region          models.Region
	ETFType         models.ETFType
	Sector          string
	Provider        string
	MinExpenseRatio decimal.Decimal
	MaxExpenseRatio decimal.Decimal
	MinAUM          decimal.Decimal
	HasHoldingsData bool
	Limit           int
}

// ETFOverview ETF概览信息
type ETFOverview struct {
	Asset              *models.Asset
	Metadata           *models.AssetMetadata
	LatestHoldings     []*models.Holding
	LatestPrice        *models.Price
	PerformanceMetrics ETFPerformance
	RiskMetrics        RiskMetrics
	HoldingsStats      HoldingsStats
}

// ETFPerformance ETF表现数据
type ETFPerformance struct {
	Period          string
	Return          decimal.Decimal
	BenchmarkReturn decimal.Decimal
	TrackingError   decimal.Decimal
	SharpeRatio     decimal.Decimal
	Volatility      decimal.Decimal
	MaxDrawdown     decimal.Decimal
	Alpha           decimal.Decimal
	Beta            decimal.Decimal
}

// RiskMetrics 风险指标
type RiskMetrics struct {
	VaR95             decimal.Decimal // 95%置信度VaR
	CVaR95            decimal.Decimal // 95%置信度CVaR
	StandardDeviation decimal.Decimal // 标准差
	Skewness          decimal.Decimal // 偏度
	Kurtosis          decimal.Decimal // 峰度
	DownsideDeviation decimal.Decimal // 下行偏差
}

// HoldingsStats 持仓统计
type HoldingsStats struct {
	TotalHoldings       int
	Top10Weight         decimal.Decimal
	Concentration       decimal.Decimal // 赫芬达尔指数
	SectorConcentration decimal.Decimal // 行业集中度
	TurnoverRate        decimal.Decimal // 换手率
	AvgMarketCap        decimal.Decimal // 平均市值
}

// AssetProviderFactory 资产数据源工厂
// 负责管理和选择合适的数据源
type AssetProviderFactory struct {
	providers map[string]AssetProvider
}

// NewAssetProviderFactory 创建新的资产数据源工厂
func NewAssetProviderFactory() *AssetProviderFactory {
	return &AssetProviderFactory{
		providers: make(map[string]AssetProvider),
	}
}

// Register 注册数据源
func (f *AssetProviderFactory) Register(provider AssetProvider) {
	f.providers[provider.Name()] = provider
}

// Get 获取指定名称的数据源
func (f *AssetProviderFactory) Get(name string) (AssetProvider, bool) {
	provider, ok := f.providers[name]
	return provider, ok
}

// GetAvailableProvider 获取可用的数据源（按优先级）
func (f *AssetProviderFactory) GetAvailableProvider(ctx context.Context) (AssetProvider, error) {
	// 按优先级排序
	providers := make([]AssetProvider, 0, len(f.providers))
	for _, p := range f.providers {
		providers = append(providers, p)
	}

	// 按优先级排序
	sortProvidersByPriority(providers)

	// 返回第一个可用的
	for _, provider := range providers {
		if provider.IsAvailable(ctx) {
			return provider, nil
		}
	}

	return nil, ErrNoAvailableProvider
}

// GetAllProviders 获取所有数据源
func (f *AssetProviderFactory) GetAllProviders() []AssetProvider {
	providers := make([]AssetProvider, 0, len(f.providers))
	for _, p := range f.providers {
		providers = append(providers, p)
	}
	return providers
}

// GetETFProvider 获取ETF数据源（优先选择ETFProvider）
func (f *AssetProviderFactory) GetETFProvider(ctx context.Context) (ETFProvider, error) {
	provider, err := f.GetAvailableProvider(ctx)
	if err != nil {
		return nil, err
	}

	if etfProvider, ok := provider.(ETFProvider); ok {
		return etfProvider, nil
	}

	// 如果没有专门的ETFProvider，使用通用AssetProvider
	return nil, ErrNoETFProvider
}

// sortProvidersByPriority 按优先级排序数据源
func sortProvidersByPriority(providers []AssetProvider) {
	// 使用稳定的排序算法
	for i := 0; i < len(providers)-1; i++ {
		for j := 0; j < len(providers)-i-1; j++ {
			if providers[j].Priority() > providers[j+1].Priority() {
				providers[j], providers[j+1] = providers[j+1], providers[j]
			}
		}
	}
}

// MultiSourceProvider 多数据源提供者（聚合模式）
type MultiSourceProvider struct {
	factory *AssetProviderFactory
}

// NewMultiSourceProvider 创建多数据源提供者
func NewMultiSourceProvider(factory *AssetProviderFactory) *MultiSourceProvider {
	return &MultiSourceProvider{
		factory: factory,
	}
}

// GetAsset 从多个数据源获取资产信息（使用第一个可用的）
func (m *MultiSourceProvider) GetAsset(ctx context.Context, symbol string) (*models.Asset, error) {
	provider, err := m.factory.GetAvailableProvider(ctx)
	if err != nil {
		return nil, err
	}
	return provider.GetAsset(ctx, symbol)
}

// GetPrice 从多个数据源获取价格（使用第一个可用的）
func (m *MultiSourceProvider) GetPrice(ctx context.Context, symbol string, date time.Time) (*models.Price, error) {
	provider, err := m.factory.GetAvailableProvider(ctx)
	if err != nil {
		return nil, err
	}
	return provider.GetPrice(ctx, symbol, date)
}

// GetLatestPrice 从多个数据源获取最新价格（使用第一个可用的）
func (m *MultiSourceProvider) GetLatestPrice(ctx context.Context, symbol string) (*models.Price, error) {
	provider, err := m.factory.GetAvailableProvider(ctx)
	if err != nil {
		return nil, err
	}
	return provider.GetLatestPrice(ctx, symbol)
}

// FallbackProvider 降级数据源
// 当所有主要数据源都不可用时使用
type FallbackProvider struct {
	baseProvider AssetProvider
}

// NewFallbackProvider 创建降级数据源
func NewFallbackProvider(baseProvider AssetProvider) *FallbackProvider {
	return &FallbackProvider{
		baseProvider: baseProvider,
	}
}

// Name 数据源名称
func (f *FallbackProvider) Name() string {
	return "fallback"
}

// Description 描述
func (f *FallbackProvider) Description() string {
	return "降级数据源，当所有主要数据源不可用时使用"
}

// IsAvailable 总是可用
func (f *FallbackProvider) IsAvailable(ctx context.Context) bool {
	return true
}

// RateLimit 速率限制较低
func (f *FallbackProvider) RateLimit() int {
	return 1 // 每秒1个请求
}

// Priority 优先级最低
func (f *FallbackProvider) Priority() int {
	return 100
}

// GetAsset 实现（可能返回有限信息）
func (f *FallbackProvider) GetAsset(ctx context.Context, symbol string) (*models.Asset, error) {
	// 尝试使用基础提供者
	if f.baseProvider.IsAvailable(ctx) {
		return f.baseProvider.GetAsset(ctx, symbol)
	}

	// 否则返回错误或基础信息
	return nil, ErrProviderUnavailable
}

// 其他方法类似实现...
