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

// CurrencyAPIProvider CurrencyAPI数据源实现
// 免费额度：每月10000次请求，支持170+货币，每日更新
type CurrencyAPIProvider struct {
	client              *http.Client
	apiKey              string
	name                string
	baseURL             string
	rateLimit           int                      // 每分钟请求限制
	requestCount        atomic.Int32             // 请求计数器
	responseTime        atomic.Int64             // 平均响应时间（纳秒）
	successCount        atomic.Int32             // 成功计数器
	errorCount          atomic.Int32             // 错误计数器
	lastRequest         atomic.Pointer[time.Time] // 最后请求时间
	supportedCurrencies []string                 // 支持的货币列表
}

// CurrencyAPIResponse CurrencyAPI响应结构
type CurrencyAPIResponse struct {
	Meta struct {
		LastUpdatedAt string `json:"last_updated_at"`
	} `json:"meta"`
	Data map[string]struct {
		Code  string  `json:"code"`
		Value float64 `json:"value"`
	} `json:"data"`
	Error bool   `json:"error"`
	Code  string `json:"code"`
	Info  string `json:"info"`
}

// NewCurrencyAPIProvider 创建CurrencyAPI数据源提供者
func NewCurrencyAPIProvider(apiKey string) *CurrencyAPIProvider {
	provider := &CurrencyAPIProvider{
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
		name:      "currencyapi",
		baseURL:   "https://api.currencyapi.com/v3",
		rateLimit: 10000 / (30 * 24 * 60), // 每月10000次，转换为每分钟限制
	}

	// 初始化常用货币列表
	provider.supportedCurrencies = []string{
		"USD", "EUR", "GBP", "JPY", "CNY", "HKD", "CAD", "AUD", "CHF",
		"SGD", "KRW", "INR", "RUB", "BRL", "MXN", "TRY", "ZAR", "SEK",
		"NOK", "DKK", "PLN", "THB", "IDR", "MYR", "PHP", "VND", "AED",
		"SAR", "QAR", "KWD", "BHD", "OMR", "JOD", "LBP", "EGP", "MAD",
		"TND", "DZD", "LYD", "SDG", "TZS", "KES", "UGX", "GHS", "NGN",
		"XAF", "XOF", "XPF", "CLP", "COP", "PEN", "BOB", "PYG", "UYU",
		"ARS", "CRC", "DOP", "GTQ", "HNL", "NIO", "PAB", "SVC", "TTD",
		"VES", "BBD", "BMD", "BZD", "KYD", "XCD",
	}

	return provider
}

// GetName 获取数据源名称
func (p *CurrencyAPIProvider) GetName() string {
	return p.name
}

// GetBaseCurrency 获取基础货币
func (p *CurrencyAPIProvider) GetBaseCurrency() string {
	return "USD"
}

// GetRate 获取单个汇率
func (p *CurrencyAPIProvider) GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
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
func (p *CurrencyAPIProvider) GetRates(ctx context.Context, baseCurrency string) (*BatchRateResult, error) {
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
	url := fmt.Sprintf("%s/latest?apikey=%s&base_currency=%s", p.baseURL, p.apiKey, baseCurrency)

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
		utils.Warn("CurrencyAPI返回错误状态码",
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
	var apiResp CurrencyAPIResponse
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
			Error:       apiResp.Info,
			DataSource:  p.name,
			RequestTime: time.Since(startTime),
		}, &ProviderError{
			Code:    "API_ERROR",
			Message: apiResp.Info,
			Source:  p.name,
		}
	}

	// 转换响应数据
	rates := make(map[string]decimal.Decimal)
	for currencyCode, rateData := range apiResp.Data {
		rates[currencyCode] = decimal.NewFromFloat(rateData.Value)
	}

	// 添加基础货币自身的汇率
	rates[baseCurrency] = decimal.NewFromFloat(1.0)

	// 更新最后请求时间
	now := time.Now()
	p.lastRequest.Store(&now)
	p.successCount.Add(1)

	utils.Debug("CurrencyAPI请求成功",
		"base_currency", baseCurrency,
		"rates_count", len(rates),
		"response_time", time.Since(startTime).String(),
		"last_updated", apiResp.Meta.LastUpdatedAt)

	return &BatchRateResult{
		Data:        rates,
		From:        baseCurrency,
		BaseDate:    apiResp.Meta.LastUpdatedAt,
		Success:     true,
		Error:       "",
		DataSource:  p.name,
		RequestTime: time.Since(startTime),
	}, nil
}

// IsAvailable 检查数据源是否可用
func (p *CurrencyAPIProvider) IsAvailable(ctx context.Context) bool {
	// 检查API Key
	if p.apiKey == "" {
		return false
	}

	// 检查最近是否有成功请求
	if lastReq := p.lastRequest.Load(); lastReq != nil {
		// 如果最近10分钟内有成功请求，则认为可用
		if time.Since(*lastReq) < 10*time.Minute {
			return true
		}
	}

	// 尝试简单的测试请求
	_, err := p.GetRates(ctx, "USD")
	return err == nil
}

// GetRateLimit 获取速率限制
func (p *CurrencyAPIProvider) GetRateLimit() int {
	return p.rateLimit
}

// GetResponseTime 获取平均响应时间
func (p *CurrencyAPIProvider) GetResponseTime() time.Duration {
	responseTime := p.responseTime.Load()
	if responseTime == 0 {
		return 0
	}
	return time.Duration(responseTime / int64(p.requestCount.Load()))
}

// GetSuccessRate 获取成功率
func (p *CurrencyAPIProvider) GetSuccessRate() float64 {
	total := p.requestCount.Load()
	if total == 0 {
		return 0.0
	}
	success := p.successCount.Load()
	return float64(success) / float64(total)
}

// GetAPIKey 获取API Key标识
func (p *CurrencyAPIProvider) GetAPIKey() string {
	// 返回部分隐藏的API Key用于标识
	if len(p.apiKey) > 8 {
		return p.apiKey[:4] + "..." + p.apiKey[len(p.apiKey)-4:]
	}
	return "****"
}

// GetSupportedCurrencies 获取支持的货币列表
func (p *CurrencyAPIProvider) GetSupportedCurrencies() []string {
	return p.supportedCurrencies
}

// ValidateAPIKey 验证API Key是否有效
func (p *CurrencyAPIProvider) ValidateAPIKey(ctx context.Context) bool {
	if p.apiKey == "" {
		return false
	}

	// 发送测试请求
	result, err := p.GetRates(ctx, "USD")
	if err != nil {
		utils.Warn("CurrencyAPI API Key验证失败", "error", err.Error())
		return false
	}

	utils.Info("CurrencyAPI API Key验证成功",
		"base_currency", result.From,
		"rates_count", len(result.Data),
		"last_updated", result.BaseDate)

	return result.Success
}

// checkRateLimit 检查速率限制
func (p *CurrencyAPIProvider) checkRateLimit() bool {
	// 简单的速率限制检查
	// CurrencyAPI每月10000次，比较宽松
	requestCount := p.requestCount.Load()
	if requestCount > int32(p.rateLimit*60*24) { // 每天限制检查
		utils.Warn("CurrencyAPI速率限制接近上限",
			"request_count", requestCount,
			"rate_limit", p.rateLimit)
		return false
	}
	return true
}

// isCurrencySupported 检查货币是否支持
func (p *CurrencyAPIProvider) isCurrencySupported(currency string) bool {
	currency = p.normalizeCurrencyCode(currency)

	// 检查常用货币列表
	for _, supported := range p.supportedCurrencies {
		if supported == currency {
			return true
		}
	}

	// CurrencyAPI支持170+货币，这里放宽检查
	// 主要检查货币代码格式
	if len(currency) != 3 {
		return false
	}

	// 检查是否为有效的ISO 4217货币代码格式
	for _, r := range currency {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return false
		}
	}

	return true
}

// normalizeCurrencyCode 标准化货币代码
func (p *CurrencyAPIProvider) normalizeCurrencyCode(currency string) string {
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
func (p *CurrencyAPIProvider) recordRequest(startTime time.Time) {
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