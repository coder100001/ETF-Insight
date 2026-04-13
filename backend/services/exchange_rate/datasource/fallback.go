package datasource

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// FallbackProvider 本地缓存后备数据源实现
// 当所有外部数据源都不可用时使用本地缓存或默认值
type FallbackProvider struct {
	name                string
	cache               sync.Map                   // 内存缓存 key=currency_pair, value= CachedRate
	lastUpdateTime      sync.Map                   // 最后更新时间
	defaultRates        map[string]decimal.Decimal // 默认汇率表
	requestCount        atomic.Int32               // 请求计数器
	successCount        atomic.Int32               // 成功计数器
	errorCount          atomic.Int32               // 错误计数器
	cacheTTL            time.Duration              // 缓存有效期
	supportedCurrencies []string                   // 支持的货币列表
}

// CachedRate 缓存汇率结构
type CachedRate struct {
	Rate        decimal.Decimal // 汇率值
	DataSource  string          // 数据来源
	LastUpdated time.Time       // 最后更新时间
	ExpiresAt   time.Time       // 过期时间
}

// NewFallbackProvider 创建本地缓存后备数据源提供者
func NewFallbackProvider() *FallbackProvider {
	provider := &FallbackProvider{
		name:     "fallback",
		cacheTTL: 24 * time.Hour, // 缓存24小时
	}

	// 初始化默认汇率表（常用货币对，基于USD）
	provider.defaultRates = map[string]decimal.Decimal{
		"USD/USD": decimal.NewFromFloat(1.0),
		"USD/CNY": decimal.NewFromFloat(7.2),
		"USD/HKD": decimal.NewFromFloat(7.8),
		"USD/EUR": decimal.NewFromFloat(0.92),
		"USD/GBP": decimal.NewFromFloat(0.79),
		"USD/JPY": decimal.NewFromFloat(150.0),
		"USD/CAD": decimal.NewFromFloat(1.35),
		"USD/AUD": decimal.NewFromFloat(1.5),
		"USD/CHF": decimal.NewFromFloat(0.88),
		"USD/SGD": decimal.NewFromFloat(1.35),
		"USD/KRW": decimal.NewFromFloat(1350.0),
		"USD/INR": decimal.NewFromFloat(83.0),
		"USD/RUB": decimal.NewFromFloat(90.0),
		"USD/BRL": decimal.NewFromFloat(5.0),
		"USD/MXN": decimal.NewFromFloat(17.0),
		"USD/TRY": decimal.NewFromFloat(32.0),
		"USD/ZAR": decimal.NewFromFloat(19.0),

		"CNY/CNY": decimal.NewFromFloat(1.0),
		"CNY/USD": decimal.NewFromFloat(0.1389),
		"CNY/HKD": decimal.NewFromFloat(1.0833),
		"CNY/EUR": decimal.NewFromFloat(0.1278),

		"HKD/HKD": decimal.NewFromFloat(1.0),
		"HKD/USD": decimal.NewFromFloat(0.1282),
		"HKD/CNY": decimal.NewFromFloat(0.9231),
	}

	provider.supportedCurrencies = []string{
		"USD", "CNY", "HKD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF",
		"SGD", "KRW", "INR", "RUB", "BRL", "MXN", "TRY", "ZAR",
	}

	provider.initializeCache()
	return provider
}

// initializeCache 初始化缓存
func (p *FallbackProvider) initializeCache() {
	now := time.Now()
	for pair, rate := range p.defaultRates {
		p.cache.Store(pair, &CachedRate{
			Rate:        rate,
			DataSource:  "default",
			LastUpdated: now,
			ExpiresAt:   now.Add(p.cacheTTL),
		})
	}
}

// GetName 获取数据源名称
func (p *FallbackProvider) GetName() string {
	return p.name
}

// GetBaseCurrency 获取基础货币
func (p *FallbackProvider) GetBaseCurrency() string {
	return "USD"
}

// GetRate 获取单个汇率
func (p *FallbackProvider) GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	p.requestCount.Add(1)

	fromCurrency = normalizeCurrency(fromCurrency)
	toCurrency = normalizeCurrency(toCurrency)

	if fromCurrency == toCurrency {
		p.successCount.Add(1)
		return decimal.NewFromFloat(1.0), nil
	}

	// 检查缓存
	rate, found := p.getFromCache(fromCurrency, toCurrency)
	if found {
		p.successCount.Add(1)
		return rate, nil
	}

	// 检查默认汇率表
	rate, found = p.getDefaultRate(fromCurrency, toCurrency)
	if found {
		p.cacheRate(fromCurrency, toCurrency, rate, "default")
		p.successCount.Add(1)
		return rate, nil
	}

	// 交叉汇率计算
	if fromCurrency != "USD" && toCurrency != "USD" {
		fromToUSD, found1 := p.getFromCacheOrDefault(fromCurrency, "USD")
		usdToTarget, found2 := p.getFromCacheOrDefault("USD", toCurrency)
		if found1 && found2 {
			crossRate := fromToUSD.Mul(usdToTarget)
			p.cacheRate(fromCurrency, toCurrency, crossRate, "calculated")
			p.successCount.Add(1)
			return crossRate, nil
		}
	}

	p.errorCount.Add(1)
	return decimal.Zero, &ProviderError{
		Code:    "RATE_NOT_AVAILABLE",
		Message: fmt.Sprintf("无法提供汇率: %s->%s", fromCurrency, toCurrency),
		Source:  p.name,
	}
}

// GetRates 批量获取汇率
func (p *FallbackProvider) GetRates(ctx context.Context, baseCurrency string) (*BatchRateResult, error) {
	startTime := time.Now()
	baseCurrency = normalizeCurrency(baseCurrency)

	rates := make(map[string]decimal.Decimal)
	rates[baseCurrency] = decimal.NewFromFloat(1.0)

	for _, target := range p.supportedCurrencies {
		if target == baseCurrency {
			continue
		}
		rate, err := p.GetRate(ctx, baseCurrency, target)
		if err == nil {
			rates[target] = rate
		}
	}

	p.successCount.Add(1)
	return &BatchRateResult{
		Data:        rates,
		From:        baseCurrency,
		BaseDate:    time.Now().Format("2006-01-02"),
		Success:     true,
		DataSource:  p.name,
		RequestTime: time.Since(startTime),
	}, nil
}

// IsAvailable 总是可用
func (p *FallbackProvider) IsAvailable(ctx context.Context) bool {
	return true
}

// GetRateLimit 无限制
func (p *FallbackProvider) GetRateLimit() int {
	return 1000
}

// GetResponseTime 本地极快
func (p *FallbackProvider) GetResponseTime() time.Duration {
	return 1 * time.Millisecond
}

// GetSuccessRate 获取成功率
func (p *FallbackProvider) GetSuccessRate() float64 {
	total := p.requestCount.Load()
	if total == 0 {
		return 1.0
	}
	return float64(p.successCount.Load()) / float64(total)
}

// GetAPIKey 不需要API Key
func (p *FallbackProvider) GetAPIKey() string {
	return "local_cache"
}

// GetSupportedCurrencies 获取支持的货币列表
func (p *FallbackProvider) GetSupportedCurrencies() []string {
	return p.supportedCurrencies
}

// ValidateAPIKey 总是有效
func (p *FallbackProvider) ValidateAPIKey(ctx context.Context) bool {
	return true
}

// UpdateCache 更新缓存（供外部调用，如数据同步时使用）
func (p *FallbackProvider) UpdateCache(fromCurrency, toCurrency string, rate decimal.Decimal, source string) {
	p.cacheRate(fromCurrency, toCurrency, rate, source)
	utils.Info("Fallback: 缓存已更新",
		"pair", fmt.Sprintf("%s/%s", fromCurrency, toCurrency),
		"rate", rate.String(),
		"source", source)
}

// BatchUpdateCache 批量更新缓存
func (p *FallbackProvider) BatchUpdateCache(baseCurrency string, rates map[string]decimal.Decimal, source string) {
	count := 0
	for target, rate := range rates {
		p.cacheRate(baseCurrency, target, rate, source)
		count++
	}
	utils.Info("Fallback: 批量缓存已更新",
		"base", baseCurrency,
		"count", count,
		"source", source)
}

// GetCacheStats 获取缓存统计
func (p *FallbackProvider) GetCacheStats() map[string]interface{} {
	cacheCount := 0
	expiredCount := 0
	now := time.Now()

	p.cache.Range(func(key, value interface{}) bool {
		cacheCount++
		if cached, ok := value.(*CachedRate); ok {
			if now.After(cached.ExpiresAt) {
				expiredCount++
			}
		}
		return true
	})

	return map[string]interface{}{
		"total_cached":   cacheCount,
		"expired_cached": expiredCount,
		"valid_cached":   cacheCount - expiredCount,
		"cache_ttl":      p.cacheTTL.String(),
	}
}

// CleanExpiredCache 清理过期缓存
func (p *FallbackProvider) CleanExpiredCache() int {
	now := time.Now()
	cleaned := 0

	p.cache.Range(func(key, value interface{}) bool {
		if cached, ok := value.(*CachedRate); ok {
			if now.After(cached.ExpiresAt) {
				p.cache.Delete(key)
				cleaned++
			}
		}
		return true
	})

	if cleaned > 0 {
		utils.Info("Fallback: 清理过期缓存", "cleaned", cleaned)
	}
	return cleaned
}

// --- 内部方法 ---

func (p *FallbackProvider) cacheRate(from, to string, rate decimal.Decimal, source string) {
	key := fmt.Sprintf("%s/%s", normalizeCurrency(from), normalizeCurrency(to))
	now := time.Now()
	p.cache.Store(key, &CachedRate{
		Rate:        rate,
		DataSource:  source,
		LastUpdated: now,
		ExpiresAt:   now.Add(p.cacheTTL),
	})
}

func (p *FallbackProvider) getFromCache(from, to string) (decimal.Decimal, bool) {
	key := fmt.Sprintf("%s/%s", normalizeCurrency(from), normalizeCurrency(to))
	if val, ok := p.cache.Load(key); ok {
		if cached, ok := val.(*CachedRate); ok {
			if time.Now().Before(cached.ExpiresAt) || cached.DataSource == "default" {
				// 默认值不过期
				return cached.Rate, true
			}
		}
	}
	return decimal.Zero, false
}

func (p *FallbackProvider) getDefaultRate(from, to string) (decimal.Decimal, bool) {
	key := fmt.Sprintf("%s/%s", normalizeCurrency(from), normalizeCurrency(to))
	if rate, ok := p.defaultRates[key]; ok {
		return rate, true
	}
	return decimal.Zero, false
}

func (p *FallbackProvider) getFromCacheOrDefault(from, to string) (decimal.Decimal, bool) {
	if rate, found := p.getFromCache(from, to); found {
		return rate, true
	}
	return p.getDefaultRate(from, to)
}

func (p *FallbackProvider) isCurrencySupported(currency string) bool {
	currency = normalizeCurrency(currency)
	for _, c := range p.supportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// normalizeCurrency 标准化货币代码（包级函数）
func normalizeCurrency(currency string) string {
	result := make([]byte, 0, len(currency))
	for _, r := range currency {
		if r >= 'a' && r <= 'z' {
			result = append(result, byte(r-32))
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}
