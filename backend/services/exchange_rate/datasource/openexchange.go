package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// OpenExchangeProvider Open Exchange Rates数据源实现
// 免费额度：每月1000次请求，支持200+货币，每小时更新
type OpenExchangeProvider struct {
	client         *http.Client
	apiKey         string
	name           string
	baseURL        string
	rateLimit      int                      // 每分钟请求限制
	requestCount   atomic.Int32             // 请求计数器
	responseTime   atomic.Int64             // 平均响应时间（纳秒）
	successCount   atomic.Int32             // 成功计数器
	errorCount     atomic.Int32             // 错误计数器
	lastRequest    atomic.Pointer[time.Time] // 最后请求时间
	supportedCurrencies []string             // 支持的货币列表
}

// OpenExchangeResponse Open Exchange Rates API响应结构
type OpenExchangeResponse struct {
	Disclaimer  string             `json:"disclaimer"`
	License     string             `json:"license"`
	Timestamp   int64              `json:"timestamp"`
	Base        string             `json:"base"`
	Rates       map[string]float64 `json:"rates"`
	Error       bool               `json:"error"`
	Status      int                `json:"status"`
	Message     string             `json:"message"`
	Description string             `json:"description"`
}

// NewOpenExchangeProvider 创建Open Exchange Rates数据源提供者
func NewOpenExchangeProvider(apiKey string) *OpenExchangeProvider {
	provider := &OpenExchangeProvider{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		},
		apiKey:    apiKey,
		name:      "openexchange",
		baseURL:   "https://openexchangerates.org/api",
		rateLimit: 1000 / (30 * 24 * 60), // 每月1000次，转换为每分钟限制
	}

	// 初始化常用货币列表
	provider.supportedCurrencies = []string{
		"USD", "EUR", "GBP", "JPY", "CNY", "HKD", "CAD", "AUD", "CHF",
		"SGD", "KRW", "INR", "RUB", "BRL", "MXN", "TRY", "ZAR", "SEK",
		"NOK", "DKK", "PLN", "THB", "IDR", "MYR", "PHP", "VND", "AED",
	}

	return provider
}

// GetName 获取数据源名称
func (p *OpenExchangeProvider) GetName() string {
	return p.name
}

// GetBaseCurrency 获取基础货币
func (p *OpenExchangeProvider) GetBaseCurrency() string {
	return "USD"
}

// GetRate 获取单个汇率
func (p *OpenExchangeProvider) GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	startTime := time.Now()
	defer p.recordRequest(startTime)

	// 检查是否支持该货币对
	if !p.isCurrencySupported(fromCurrency) || !p.isCurrencySupported(toCurrency) {
		p.errorCount.Add(1)
		return decimal.Zero, &ProviderError{
			Code:    "CURRENCY_NOT_SUPPORTED",
			Message: fmt.Sprintf("货币不支持: %s->%s", fromCurrency, toCurrency),
			Source:  p.name,
		}
	}

	// 如果是相同货币，直接返回1
	if fromCurrency == toCurrency {
		p.successCount.Add(1)
		return decimal.NewFromFloat(1.0), nil
	}

	// 获取汇率数据
	result, err := p.GetRates(ctx, fromCurrency)
	if err != nil {
		p.errorCount.Add(1)
		return decimal.Zero, err
	}

	// 查找目标货币汇率
	if rate, ok := result.Data[toCurrency]; ok {
		p.successCount.Add(1)
		return rate, nil
	}

	// 如果找不到直接汇率，尝试通过USD计算交叉汇率
	if fromCurrency != "USD" && toCurrency != "USD" {
		fromToUSD, err1 := p.GetRate(ctx, fromCurrency, "USD")
		usdToTarget, err2 := p.GetRate(ctx, "USD", toCurrency)

		if err1 == nil && err2 == nil {
			crossRate := fromToUSD.Mul(usdToTarget)
			p.successCount.Add(1)
			return crossRate, nil
		}
	}

	p.errorCount.Add(1)
	return decimal.Zero, &ProviderError{
		Code:    "RATE_NOT_FOUND",
		Message: fmt.Sprintf("未找到汇率: %s->%s", fromCurrency, toCurrency),
		Source:  p.name,
	}
}

// GetRates 批量获取汇率（以基础货币为基准）
func (p *OpenExchangeProvider) GetRates(ctx context.Context, baseCurrency string) (*BatchRateResult, error) {
	startTime := time.Now()
	defer p.recordRequest(startTime)

	// 检查速率限制
	if !p.checkRateLimit() {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       "速率限制超出",
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, ErrRateLimitExceeded
	}

	// 检查API Key
	if p.apiKey == "" {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       "API密钥未配置",
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, ErrInvalidAPIKey
	}

	// 构建API URL
	url := fmt.Sprintf("%s/latest.json?app_id=%s&base=%s", p.baseURL, p.apiKey, baseCurrency)

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       fmt.Sprintf("创建请求失败: %v", err),
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, Wrap(err, p.name, "REQUEST_CREATE_ERROR", "创建HTTP请求失败")
	}

	req.Header.Set("User-Agent", "ETF-Insight/2.2.0")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       fmt.Sprintf("网络请求失败: %v", err),
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, Wrap(err, p.name, "NETWORK_ERROR", "网络请求失败")
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       fmt.Sprintf("读取响应失败: %v", err),
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, Wrap(err, p.name, "RESPONSE_READ_ERROR", "读取响应失败")
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		p.errorCount.Add(1)
		utils.Warn("OpenExchange API返回错误状态码",
			"status_code", resp.StatusCode,
			"url", url)

		return &BatchRateResult{
			Success:     false,
			Error:       fmt.Sprintf("API返回状态码: %d", resp.StatusCode),
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, &ProviderError{
			Code:    "API_ERROR",
			Message: fmt.Sprintf("API返回错误: %d", resp.StatusCode),
			Source:  p.name,
		}
	}

	// 解析JSON响应
	var apiResp OpenExchangeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       fmt.Sprintf("解析JSON失败: %v", err),
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, Wrap(err, p.name, "DATA_PARSE_ERROR", "解析JSON数据失败")
	}

	// 检查API错误
	if apiResp.Error {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       apiResp.Description,
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, &ProviderError{
			Code:    "API_ERROR",
			Message: apiResp.Description,
			Source:  p.name,
		}
	}

	// 转换响应数据
	rates := make(map[string]decimal.Decimal)
	for currency, rate := range apiResp.Rates {
		rates[currency] = decimal.NewFromFloat(rate)
	}

	// 添加基础货币自身的汇率
	rates[baseCurrency] = decimal.NewFromFloat(1.0)

	// 更新最后请求时间
	now := time.Now()
	p.lastRequest.Store(&now)
	p.successCount.Add(1)

	utils.Debug("OpenExchange API请求成功",
		"base_currency", baseCurrency,
		"rates_count", len(rates),
		"response_time", time.Since(startTime).String())

	return &BatchRateResult{
		Data:        rates,
		From:        baseCurrency,
		BaseDate:    time.Unix(apiResp.Timestamp, 0).Format("2006-01-02"),
		Success:     true,
		Error:       "",
		DataSource:  p.name,
		RequestTime: time.Since(startTime),
	}, nil
}

// IsAvailable 检查数据源是否可用
func (p *OpenExchangeProvider) IsAvailable(ctx context.Context) bool {
	// 检查API Key
	if p.apiKey == "" {
		return false
	}

	// 检查最近是否有成功请求
	if lastReq := p.lastRequest.Load(); lastReq != nil {
		// 如果最近5分钟内有成功请求，则认为可用
		if time.Since(*lastReq) < 5*time.Minute {
			return true
		}
	}

	// 尝试简单的测试请求
	_, err := p.GetRates(ctx, "USD")
	return err == nil
}

// GetRateLimit 获取速率限制
func (p *OpenExchangeProvider) GetRateLimit() int {
	return p.rateLimit
}

// GetResponseTime 获取平均响应时间
func (p *OpenExchangeProvider) GetResponseTime() time.Duration {
	responseTime := p.responseTime.Load()
	if responseTime == 0 {
		return 0
	}
	return time.Duration(responseTime / int64(p.requestCount.Load()))
}

// GetSuccessRate 获取成功率
func (p *OpenExchangeProvider) GetSuccessRate() float64 {
	total := p.requestCount.Load()
	if total == 0 {
		return 0.0
	}
	success := p.successCount.Load()
	return float64(success) / float64(total)
}

// GetAPIKey 获取API Key标识
func (p *OpenExchangeProvider) GetAPIKey() string {
	// 返回部分隐藏的API Key用于标识
	if len(p.apiKey) > 8 {
		return p.apiKey[:4] + "..." + p.apiKey[len(p.apiKey)-4:]
	}
	return "****"
}

// GetSupportedCurrencies 获取支持的货币列表
func (p *OpenExchangeProvider) GetSupportedCurrencies() []string {
	return p.supportedCurrencies
}

// ValidateAPIKey 验证API Key是否有效
func (p *OpenExchangeProvider) ValidateAPIKey(ctx context.Context) bool {
	if p.apiKey == "" {
		return false
	}

	// 发送测试请求
	result, err := p.GetRates(ctx, "USD")
	if err != nil {
		utils.Warn("OpenExchange API Key验证失败", "error", err.Error())
		return false
	}

	utils.Info("OpenExchange API Key验证成功",
		"base_currency", result.From,
		"rates_count", len(result.Data))

	return result.Success
}

// checkRateLimit 检查速率限制
func (p *OpenExchangeProvider) checkRateLimit() bool {
	// 简单的速率限制检查
	// 实际应该使用更精确的令牌桶算法
	requestCount := p.requestCount.Load()
	if requestCount > int32(p.rateLimit*60) { // 每小时限制
		return false
	}
	return true
}

// isCurrencySupported 检查货币是否支持
func (p *OpenExchangeProvider) isCurrencySupported(currency string) bool {
	currency = p.normalizeCurrencyCode(currency)

	// 检查常用货币列表
	for _, supported := range p.supportedCurrencies {
		if supported == currency {
			return true
		}
	}

	// 默认支持主要货币
	majorCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true,
		"CNY": true, "HKD": true, "CAD": true, "AUD": true,
	}

	return majorCurrencies[currency]
}

// normalizeCurrencyCode 标准化货币代码
func (p *OpenExchangeProvider) normalizeCurrencyCode(currency string) string {
	// 转换为大写
	normalized := ""
	for _, r := range currency {
		if r >= 'a' && r <= 'z' {
			normalized += string(r - 32)
		} else {
			normalized += string(r)
		}
	}
	return normalized
}

// recordRequest 记录请求统计
func (p *OpenExchangeProvider) recordRequest(startTime time.Time) {
	duration := time.Since(startTime)
	p.requestCount.Add(1)

	// 更新平均响应时间
	oldAvg := p.responseTime.Load()
	count := p.requestCount.Load()
	if count > 1 {
		newAvg := (oldAvg*int64(count-1) + duration.Nanoseconds()) / int64(count)
		p.responseTime.Store(newAvg)
	} else {
		p.responseTime.Store(duration.Nanoseconds())
	}
}