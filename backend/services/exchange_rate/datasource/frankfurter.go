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

// FrankfurterProvider Frankfurter数据源实现
// 免费额度：无限请求，支持30+货币，每日更新
type FrankfurterProvider struct {
	client              *http.Client
	name                string
	baseURL             string
	rateLimit           int                       // 每分钟请求限制（理论上无限）
	requestCount        atomic.Int32              // 请求计数器
	responseTime        atomic.Int64              // 平均响应时间（纳秒）
	successCount        atomic.Int32              // 成功计数器
	errorCount          atomic.Int32              // 错误计数器
	lastRequest         atomic.Pointer[time.Time] // 最后请求时间
	supportedCurrencies []string                  // 支持的货币列表
}

// FrankfurterResponse Frankfurter API响应结构
type FrankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
	Error  string             `json:"error,omitempty"`
}

// createProxyTransport 创建带代理支持的 Transport
func createProxyTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 10,
		Proxy:               http.ProxyFromEnvironment, // 自动使用 HTTP_PROXY/HTTPS_PROXY 环境变量
	}
}

// NewFrankfurterProvider 创建Frankfurter数据源提供者
func NewFrankfurterProvider() *FrankfurterProvider {
	provider := &FrankfurterProvider{
		client: &http.Client{
			Timeout:   30 * time.Second, // 增加超时时间
			Transport: createProxyTransport(),
		},
		name:      "frankfurter",
		baseURL:   "https://api.frankfurter.app",
		rateLimit: 100, // 保守限制，避免滥用
	}

	// Frankfurter支持的主要货币（30+）
	provider.supportedCurrencies = []string{
		"USD", "EUR", "GBP", "JPY", "CHF", "CAD", "AUD", "NZD",
		"CNY", "HKD", "SGD", "KRW", "INR", "BRL", "RUB", "ZAR",
		"TRY", "MXN", "SEK", "NOK", "DKK", "PLN", "THB", "IDR",
		"MYR", "PHP", "CZK", "HUF", "ILS", "CLP", "ARS", "PEN",
		"COP", "UYU", "VES",
	}

	return provider
}

// GetName 获取数据源名称
func (p *FrankfurterProvider) GetName() string {
	return p.name
}

// GetBaseCurrency 获取基础货币
func (p *FrankfurterProvider) GetBaseCurrency() string {
	return "EUR" // Frankfurter默认以EUR为基础
}

// GetRate 获取单个汇率
func (p *FrankfurterProvider) GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
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

	// 如果找不到直接汇率，尝试通过EUR计算交叉汇率
	if fromCurrency != "EUR" && toCurrency != "EUR" {
		fromToEUR, err1 := p.GetRate(ctx, fromCurrency, "EUR")
		eurToTarget, err2 := p.GetRate(ctx, "EUR", toCurrency)

		if err1 == nil && err2 == nil {
			crossRate := fromToEUR.Mul(eurToTarget)
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
func (p *FrankfurterProvider) GetRates(ctx context.Context, baseCurrency string) (*BatchRateResult, error) {
	startTime := time.Now()
	defer p.recordRequest(startTime)

	// 检查速率限制（Frankfurter无限，但需要避免滥用）
	if !p.checkRateLimit() {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       "本地速率限制超出",
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, ErrRateLimitExceeded
	}

	// 构建API URL
	url := fmt.Sprintf("%s/latest?from=%s", p.baseURL, baseCurrency)

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
		utils.Warn("Frankfurter API返回错误状态码",
			"status_code", resp.StatusCode,
			"url", url)

		// 尝试解析错误信息
		var errorResp FrankfurterResponse
		if jsonErr := json.Unmarshal(body, &errorResp); jsonErr == nil && errorResp.Error != "" {
			return &BatchRateResult{
					Success:     false,
					Error:       errorResp.Error,
					DataSource:  p.name,
					RequestTime: time.Since(startTime),
				}, &ProviderError{
					Code:    "API_ERROR",
					Message: errorResp.Error,
					Source:  p.name,
				}
		}

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
	var apiResp FrankfurterResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		p.errorCount.Add(1)
		return &BatchRateResult{
			Success:     false,
			Error:       fmt.Sprintf("解析JSON失败: %v", err),
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, Wrap(err, p.name, "DATA_PARSE_ERROR", "解析JSON数据失败")
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

	utils.Debug("Frankfurter API请求成功",
		"base_currency", baseCurrency,
		"rates_count", len(rates),
		"response_time", time.Since(startTime).String(),
		"date", apiResp.Date)

	return &BatchRateResult{
		Data:        rates,
		From:        baseCurrency,
		BaseDate:    apiResp.Date,
		Success:     true,
		Error:       "",
		DataSource:  p.name,
		RequestTime: time.Since(startTime),
	}, nil
}

// IsAvailable 检查数据源是否可用
func (p *FrankfurterProvider) IsAvailable(ctx context.Context) bool {
	// Frankfurter是免费的公共服务，通常可用
	// 但需要检查最近是否有成功请求

	// 检查最近是否有成功请求
	if lastReq := p.lastRequest.Load(); lastReq != nil {
		// 如果最近15分钟内有成功请求，则认为可用
		if time.Since(*lastReq) < 15*time.Minute {
			return true
		}
	}

	// 尝试简单的测试请求
	_, err := p.GetRates(ctx, "EUR")
	return err == nil
}

// GetRateLimit 获取速率限制
func (p *FrankfurterProvider) GetRateLimit() int {
	return p.rateLimit
}

// GetResponseTime 获取平均响应时间
func (p *FrankfurterProvider) GetResponseTime() time.Duration {
	responseTime := p.responseTime.Load()
	if responseTime == 0 {
		return 0
	}
	return time.Duration(responseTime / int64(p.requestCount.Load()))
}

// GetSuccessRate 获取成功率
func (p *FrankfurterProvider) GetSuccessRate() float64 {
	total := p.requestCount.Load()
	if total == 0 {
		return 0.0
	}
	success := p.successCount.Load()
	return float64(success) / float64(total)
}

// GetAPIKey 获取API Key标识
func (p *FrankfurterProvider) GetAPIKey() string {
	// Frankfurter是免费的，不需要API Key
	return "free_public_api"
}

// GetSupportedCurrencies 获取支持的货币列表
func (p *FrankfurterProvider) GetSupportedCurrencies() []string {
	return p.supportedCurrencies
}

// ValidateAPIKey 验证API Key是否有效
func (p *FrankfurterProvider) ValidateAPIKey(ctx context.Context) bool {
	// Frankfurter不需要API Key，总是返回true
	return true
}

// checkRateLimit 检查速率限制
func (p *FrankfurterProvider) checkRateLimit() bool {
	// Frankfurter是免费公共服务，需要避免滥用
	// 设置本地限制：每分钟最多100次请求
	requestCount := p.requestCount.Load()

	// 检查最近1分钟的请求频率
	if lastReq := p.lastRequest.Load(); lastReq != nil {
		if time.Since(*lastReq) < time.Minute {
			// 如果最近1分钟请求太多，限制一下
			if requestCount > int32(p.rateLimit) {
				utils.Warn("Frankfurter本地速率限制超出",
					"request_count", requestCount,
					"rate_limit", p.rateLimit)
				return false
			}
		}
	}
	return true
}

// isCurrencySupported 检查货币是否支持
func (p *FrankfurterProvider) isCurrencySupported(currency string) bool {
	currency = p.normalizeCurrencyCode(currency)

	// 检查支持的货币列表
	for _, supported := range p.supportedCurrencies {
		if supported == currency {
			return true
		}
	}

	// Frankfurter支持30+货币，这里检查是否是主要货币
	majorCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true,
		"CHF": true, "CAD": true, "AUD": true, "NZD": true,
		"CNY": true, "HKD": true, "SGD": true, "KRW": true,
	}

	return majorCurrencies[currency]
}

// normalizeCurrencyCode 标准化货币代码
func (p *FrankfurterProvider) normalizeCurrencyCode(currency string) string {
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
func (p *FrankfurterProvider) recordRequest(startTime time.Time) {
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
