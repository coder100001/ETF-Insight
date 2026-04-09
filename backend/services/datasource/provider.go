// Package datasource 提供ETF数据源接口和实现
// 采用策略模式，支持多种数据源切换
package datasource

import (
	"context"
	"time"
)

// QuoteData 标准化报价数据
// 所有数据源实现都需要转换为这个通用格式
type QuoteData struct {
	Symbol           string
	Name             string
	CurrentPrice     float64
	OpenPrice        float64
	DayHigh          float64
	DayLow           float64
	Volume           int64
	PreviousClose    float64
	Change           float64
	ChangePercent    float64
	MarketCap        int64
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	AverageVolume    int64
	Beta             float64
	PERatio          float64
	DividendYield    float64
	Currency         string
	Exchange         string
	Timestamp        time.Time
	DataSource       string // 数据来源标识
}

// ETFInfo ETF基本信息
type ETFInfo struct {
	Symbol       string
	Name         string
	Category     string
	Description  string
	Provider     string
	ExpenseRatio float64
}

// DataSourceProvider 数据源提供者接口
// 实现此接口即可接入同步系统
type DataSourceProvider interface {
	// GetName 获取数据源名称
	GetName() string

	// GetQuote 获取单只股票报价
	GetQuote(ctx context.Context, symbol string) (*QuoteData, error)

	// GetQuotes 批量获取股票报价
	GetQuotes(ctx context.Context, symbols []string) ([]*QuoteData, error)

	// IsAvailable 检查数据源是否可用
	IsAvailable(ctx context.Context) bool

	// GetRateLimit 获取速率限制（每秒请求数）
	GetRateLimit() int
}

// ProviderFactory 数据源工厂
// 用于创建和管理不同的数据源实例
type ProviderFactory struct {
	providers map[string]DataSourceProvider
}

// NewProviderFactory 创建数据源工厂
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]DataSourceProvider),
	}
}

// Register 注册数据源提供者
func (f *ProviderFactory) Register(name string, provider DataSourceProvider) {
	f.providers[name] = provider
}

// Get 获取指定名称的数据源
func (f *ProviderFactory) Get(name string) (DataSourceProvider, bool) {
	provider, ok := f.providers[name]
	return provider, ok
}

// GetDefault 获取默认数据源
// 按优先级返回第一个可用的数据源
func (f *ProviderFactory) GetDefault(ctx context.Context) (DataSourceProvider, error) {
	// 优先级顺序：Finage -> Finnhub -> Yahoo -> Fallback
	priorities := []string{"finage", "finnhub", "yahoo", "fallback"}

	for _, name := range priorities {
		if provider, ok := f.providers[name]; ok && provider.IsAvailable(ctx) {
			return provider, nil
		}
	}

	return nil, ErrNoAvailableProvider
}

// ListProviders 列出所有已注册的提供者
func (f *ProviderFactory) ListProviders() []string {
	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
