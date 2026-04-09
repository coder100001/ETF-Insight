package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// FinnhubProvider Finnhub数据源实现
type FinnhubProvider struct {
	apiKey      string
	client      *http.Client
	baseURL     string
	rateLimiter chan struct{}
	mu          sync.RWMutex
	available   bool
}

// FinnhubQuote Finnhub API响应结构
type FinnhubQuote struct {
	Current   float64 `json:"c"`  // 当前价格
	Change    float64 `json:"d"`  // 价格变动
	Percent   float64 `json:"dp"` // 变动百分比
	High      float64 `json:"h"`  // 当日最高
	Low       float64 `json:"l"`  // 当日最低
	Open      float64 `json:"o"`  // 开盘价
	Previous  float64 `json:"pc"` // 前收盘价
	Timestamp int64   `json:"t"`  // 时间戳
}

// FinnhubConfig 配置选项
// 使用选项模式进行配置
type FinnhubConfig struct {
	APIKey    string
	Timeout   time.Duration
	RateLimit int // 每秒最大请求数
	ProxyURL  string
}

// NewFinnhubProvider 创建Finnhub提供者
// 支持通过选项模式自定义配置
func NewFinnhubProvider(config ...FinnhubConfig) *FinnhubProvider {
	cfg := FinnhubConfig{
		APIKey:    os.Getenv("FINNHUB_API_KEY"),
		Timeout:   30 * time.Second,
		RateLimit: 60,
		ProxyURL:  getProxyURL(),
	}

	if len(config) > 0 {
		if config[0].APIKey != "" {
			cfg.APIKey = config[0].APIKey
		}
		if config[0].Timeout > 0 {
			cfg.Timeout = config[0].Timeout
		}
		if config[0].RateLimit > 0 {
			cfg.RateLimit = config[0].RateLimit
		}
		if config[0].ProxyURL != "" {
			cfg.ProxyURL = config[0].ProxyURL
		}
	}

	// 创建HTTP客户端
	transport := &http.Transport{}
	if cfg.ProxyURL != "" {
		if parsedURL, err := url.Parse(cfg.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(parsedURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	return &FinnhubProvider{
		apiKey:      cfg.APIKey,
		client:      client,
		baseURL:     "https://finnhub.io/api/v1",
		rateLimiter: make(chan struct{}, cfg.RateLimit),
		available:   cfg.APIKey != "",
	}
}

// GetName 返回提供者名称
func (f *FinnhubProvider) GetName() string {
	return "finnhub"
}

// IsAvailable 检查提供者是否可用
func (f *FinnhubProvider) IsAvailable(ctx context.Context) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.available || f.apiKey == "" {
		return false
	}

	// 尝试获取一个已知股票的报价来验证API可用性
	_, err := f.GetQuote(ctx, "AAPL")
	return err == nil
}

// GetRateLimit 返回速率限制
func (f *FinnhubProvider) GetRateLimit() int {
	return cap(f.rateLimiter)
}

// GetQuote 获取单只股票报价
func (f *FinnhubProvider) GetQuote(ctx context.Context, symbol string) (*QuoteData, error) {
	if symbol == "" {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      ErrInvalidSymbol,
		}
	}

	// 速率限制
	select {
	case f.rateLimiter <- struct{}{}:
		defer func() { <-f.rateLimiter }()
	case <-ctx.Done():
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      ctx.Err(),
			Symbol:   symbol,
		}
	}

	reqURL := fmt.Sprintf("%s/quote?symbol=%s&token=%s", f.baseURL, symbol, f.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      err,
			Symbol:   symbol,
		}
	}

	req.Header.Set("User-Agent", "ETFInsight/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      ErrNetwork,
			Symbol:   symbol,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      fmt.Errorf("%w: status %d, body: %s", ErrAPINotAvailable, resp.StatusCode, string(body)),
			Symbol:   symbol,
			Status:   resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      err,
			Symbol:   symbol,
		}
	}

	var quote FinnhubQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      ErrInvalidResponse,
			Symbol:   symbol,
		}
	}

	// 添加延迟以遵守速率限制
	time.Sleep(time.Second / time.Duration(f.GetRateLimit()))

	return f.convertToQuoteData(symbol, &quote), nil
}

// GetQuotes 批量获取股票报价
func (f *FinnhubProvider) GetQuotes(ctx context.Context, symbols []string) ([]*QuoteData, error) {
	if len(symbols) == 0 {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuotes",
			Err:      ErrInvalidSymbol,
		}
	}

	var (
		results []*QuoteData
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	// 使用工作池限制并发数
	workerCount := 10
	if len(symbols) < workerCount {
		workerCount = len(symbols)
	}

	symbolChan := make(chan string, len(symbols))
	errorChan := make(chan error, len(symbols))

	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for symbol := range symbolChan {
				quote, err := f.GetQuote(ctx, symbol)
				if err != nil {
					errorChan <- err
					continue
				}
				mu.Lock()
				results = append(results, quote)
				mu.Unlock()
			}
		}()
	}

	// 发送任务
	for _, symbol := range symbols {
		symbolChan <- symbol
	}
	close(symbolChan)

	// 等待所有工作完成
	wg.Wait()
	close(errorChan)

	// 收集错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(results) == 0 {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuotes",
			Err:      fmt.Errorf("%w: no data retrieved, errors: %v", ErrAPINotAvailable, errors),
		}
	}

	return results, nil
}

// convertToQuoteData 将Finnhub响应转换为标准格式
func (f *FinnhubProvider) convertToQuoteData(symbol string, quote *FinnhubQuote) *QuoteData {
	return &QuoteData{
		Symbol:        symbol,
		CurrentPrice:  quote.Current,
		OpenPrice:     quote.Open,
		DayHigh:       quote.High,
		DayLow:        quote.Low,
		PreviousClose: quote.Previous,
		Change:        quote.Change,
		ChangePercent: quote.Percent,
		Volume:        0, // Finnhub免费版不提供成交量
		Currency:      "USD",
		Exchange:      "NASDAQ",
		Timestamp:     time.Unix(quote.Timestamp, 0),
		DataSource:    f.GetName(),
	}
}

// getProxyURL 获取代理URL
func getProxyURL() string {
	// 按优先级检查环境变量
	for _, env := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		if url := os.Getenv(env); url != "" {
			return url
		}
	}
	return ""
}

// SetAvailability 设置可用状态（用于测试）
func (f *FinnhubProvider) SetAvailability(available bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.available = available
}
