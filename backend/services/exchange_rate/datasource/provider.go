// Package datasource 提供外汇（汇率）数据源接口和实现
// 采用策略模式，支持多种汇率数据源切换和故障转移
package datasource

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// ExchangeRateData 标准化汇率数据
// 所有数据源实现都需要转换为这个通用格式
type ExchangeRateData struct {
	FromCurrency string          // 源货币代码
	ToCurrency   string          // 目标货币代码
	Rate         decimal.Decimal // 汇率
	Date         time.Time       // 汇率日期
	LastUpdated  time.Time       // 最后更新时间
	DataSource   string          // 数据来源标识
	ValidStatus  int             // 有效状态: 1有效, 0无效
	Priority     int             // 数据源优先级
}

// BatchRateResult 批量汇率获取结果
type BatchRateResult struct {
	Data        map[string]decimal.Decimal // 汇率数据 key=currency_code, value=rate
	From        string                     // 源货币
	BaseDate    string                     // 基准日期
	Success     bool                       // 是否成功
	Error       string                     // 错误信息
	DataSource  string                     // 使用的数据源
	RequestTime time.Duration              // 请求耗时
}

// DataSourceProvider 汇率数据源提供者接口
// 实现此接口即可接入汇率同步系统
type DataSourceProvider interface {
	// GetName 获取数据源名称
	GetName() string

	// GetBaseCurrency 获取基础货币（通常是USD）
	GetBaseCurrency() string

	// GetRate 获取单个汇率
	GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error)

	// GetRates 批量获取汇率（以基础货币为基准）
	GetRates(ctx context.Context, baseCurrency string) (*BatchRateResult, error)

	// IsAvailable 检查数据源是否可用
	IsAvailable(ctx context.Context) bool

	// GetRateLimit 获取速率限制（每分钟请求数）
	GetRateLimit() int

	// GetResponseTime 获取平均响应时间
	GetResponseTime() time.Duration

	// GetSuccessRate 获取成功率
	GetSuccessRate() float64

	// GetAPIKey 获取API Key标识（用于配额统计）
	GetAPIKey() string

	// GetSupportedCurrencies 获取支持的货币列表
	GetSupportedCurrencies() []string

	// ValidateAPIKey 验证API Key是否有效
	ValidateAPIKey(ctx context.Context) bool
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

// GetAvailableProvider 获取可用的数据源
// 按优先级返回第一个可用的数据源
func (f *ProviderFactory) GetAvailableProvider(ctx context.Context) (DataSourceProvider, error) {
	// 优先级顺序：openexchange -> currencyapi -> frankfurter -> fallback
	priorities := []string{"openexchange", "currencyapi", "frankfurter", "fallback"}

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
