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

type FinageProvider struct {
	apiKey      string
	client      *http.Client
	baseURL     string
	rateLimiter chan struct{}
	mu          sync.RWMutex
	available   bool
}

// FinageQuote Finage 实时报价响应 (last/stock/{symbol})
type FinageQuote struct {
	Symbol    string  `json:"symbol"`
	Ask       float64 `json:"ask"`
	Bid       float64 `json:"bid"`
	AskSize   int     `json:"asize"`
	BidSize   int     `json:"bsize"`
	Timestamp int64   `json:"timestamp"`
}

// FinageAggResponse Finage 聚合数据响应 (agg/stock/{symbol}/1/day/{from}/{to})
type FinageAggResponse struct {
	Symbol       string `json:"symbol"`
	TotalResults int    `json:"totalResults"`
	Results      []struct {
		Open      float64 `json:"o"`
		High      float64 `json:"h"`
		Low       float64 `json:"l"`
		Close     float64 `json:"c"`
		Volume    int64   `json:"v"`
		Timestamp int64   `json:"t"`
	} `json:"results"`
}

type FinageConfig struct {
	APIKey    string
	Timeout   time.Duration
	RateLimit int
	ProxyURL  string
}

func NewFinageProvider(config ...FinageConfig) *FinageProvider {
	cfg := FinageConfig{
		APIKey:    os.Getenv("FINAGE_API_KEY"),
		Timeout:   30 * time.Second,
		RateLimit: 100,
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

	// 使用简单的 HTTP 客户端
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	// 如果设置了代理，则配置代理
	if cfg.ProxyURL != "" {
		if parsedURL, err := url.Parse(cfg.ProxyURL); err == nil {
			transport := &http.Transport{
				Proxy: http.ProxyURL(parsedURL),
			}
			client.Transport = transport
		}
	}

	return &FinageProvider{
		apiKey:      cfg.APIKey,
		client:      client,
		baseURL:     "https://api.finage.co.uk",
		rateLimiter: make(chan struct{}, cfg.RateLimit),
		available:   cfg.APIKey != "",
	}
}

func (f *FinageProvider) GetName() string {
	return "finage"
}

func (f *FinageProvider) IsAvailable(ctx context.Context) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.available || f.apiKey == "" {
		return false
	}

	// 简单检查：如果有 API key，就认为可用
	// 不进行实际的 API 调用，避免在初始化时消耗配额
	return true
}

func (f *FinageProvider) GetRateLimit() int {
	return cap(f.rateLimiter)
}

func (f *FinageProvider) GetQuote(ctx context.Context, symbol string) (*QuoteData, error) {
	if symbol == "" {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuote",
			Err:      ErrInvalidSymbol,
		}
	}

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

	// 优先使用聚合API获取完整OHLCV数据
	quoteData, err := f.getQuoteFromAgg(ctx, symbol)
	if err == nil && quoteData != nil {
		time.Sleep(time.Second / time.Duration(f.GetRateLimit()))
		return quoteData, nil
	}

	// 聚合API失败，回退到实时报价API（数据不完整但总比没有好）
	quoteData, err2 := f.getQuoteFromLast(ctx, symbol)
	if err2 != nil {
		// 两个API都失败，返回第一个错误
		return nil, err
	}

	time.Sleep(time.Second / time.Duration(f.GetRateLimit()))
	return quoteData, nil
}

// getQuoteFromAgg 使用聚合API获取完整OHLCV数据
func (f *FinageProvider) getQuoteFromAgg(ctx context.Context, symbol string) (*QuoteData, error) {
	// 获取最近3天的聚合数据，取最后一条
	now := time.Now()
	from := now.AddDate(0, 0, -5).Format("2006-01-02")
	to := now.Format("2006-01-02")

	reqURL := fmt.Sprintf("%s/agg/stock/%s/1/day/%s/%s?apikey=%s", f.baseURL, symbol, from, to, f.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromAgg",
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
			Op:       "GetQuoteFromAgg",
			Err:      fmt.Errorf("%w: %v", ErrNetwork, err),
			Symbol:   symbol,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromAgg",
			Err:      fmt.Errorf("%w: status %d, body: %s", ErrAPINotAvailable, resp.StatusCode, string(body)),
			Symbol:   symbol,
			Status:   resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromAgg",
			Err:      err,
			Symbol:   symbol,
		}
	}

	var aggResp FinageAggResponse
	if err := json.Unmarshal(body, &aggResp); err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromAgg",
			Err:      ErrInvalidResponse,
			Symbol:   symbol,
		}
	}

	if aggResp.TotalResults == 0 || len(aggResp.Results) == 0 {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromAgg",
			Err:      fmt.Errorf("no aggregate data for %s", symbol),
			Symbol:   symbol,
		}
	}

	// 取最新的数据点（最后一条）
	latest := aggResp.Results[len(aggResp.Results)-1]

	// 计算涨跌：如果有前一天数据，用前一天收盘价作为 previousClose
	var previousClose float64
	var change, changePercent float64
	if len(aggResp.Results) >= 2 {
		previousClose = aggResp.Results[len(aggResp.Results)-2].Close
		change = latest.Close - previousClose
		if previousClose > 0 {
			changePercent = (change / previousClose) * 100
		}
	}

	dataSource := f.GetName()

	return &QuoteData{
		Symbol:        symbol,
		CurrentPrice:  latest.Close,
		OpenPrice:     latest.Open,
		DayHigh:       latest.High,
		DayLow:        latest.Low,
		PreviousClose: previousClose,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        latest.Volume,
		Currency:      "USD",
		Exchange:      "NASDAQ",
		Timestamp:     time.Unix(latest.Timestamp/1000, 0),
		DataSource:    dataSource,
	}, nil
}

// getQuoteFromLast 使用实时报价API（仅ask/bid，无完整OHLCV）
func (f *FinageProvider) getQuoteFromLast(ctx context.Context, symbol string) (*QuoteData, error) {
	reqURL := fmt.Sprintf("%s/last/stock/%s?apikey=%s", f.baseURL, symbol, f.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromLast",
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
			Op:       "GetQuoteFromLast",
			Err:      fmt.Errorf("%w: %v", ErrNetwork, err),
			Symbol:   symbol,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromLast",
			Err:      fmt.Errorf("%w: status %d, body: %s", ErrAPINotAvailable, resp.StatusCode, string(body)),
			Symbol:   symbol,
			Status:   resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromLast",
			Err:      err,
			Symbol:   symbol,
		}
	}

	var quote FinageQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, &DataSourceError{
			Provider: f.GetName(),
			Op:       "GetQuoteFromLast",
			Err:      ErrInvalidResponse,
			Symbol:   symbol,
		}
	}

	return f.convertToQuoteData(symbol, &quote), nil
}

func (f *FinageProvider) GetQuotes(ctx context.Context, symbols []string) ([]*QuoteData, error) {
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

	workerCount := 10
	if len(symbols) < workerCount {
		workerCount = len(symbols)
	}

	symbolChan := make(chan string, len(symbols))
	errorChan := make(chan error, len(symbols))

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

	for _, symbol := range symbols {
		symbolChan <- symbol
	}
	close(symbolChan)

	wg.Wait()
	close(errorChan)

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

// convertToQuoteData 将Finage实时报价转换为标准格式
// 注意：last/stock API 只提供 ask/bid，OHLCV 数据不完整
// 优先使用 getQuoteFromAgg 获取完整数据
func (f *FinageProvider) convertToQuoteData(symbol string, quote *FinageQuote) *QuoteData {
	midPrice := (quote.Ask + quote.Bid) / 2

	// 当 ask 或 bid 为 0 时，使用非零的那个
	if quote.Ask == 0 && quote.Bid == 0 {
		midPrice = 0
	} else if quote.Ask == 0 {
		midPrice = quote.Bid
	} else if quote.Bid == 0 {
		midPrice = quote.Ask
	}

	return &QuoteData{
		Symbol:        symbol,
		CurrentPrice:  midPrice,
		OpenPrice:     0, // last API 不提供，需要聚合API
		DayHigh:       0,
		DayLow:        0,
		PreviousClose: 0,
		Change:        0,
		ChangePercent: 0,
		Volume:        0,
		Currency:      "USD",
		Exchange:      "NASDAQ",
		Timestamp:     time.Unix(quote.Timestamp/1000, 0),
		DataSource:    f.GetName() + "_last", // 标注数据来源为 last API（数据不完整）
	}
}

func (f *FinageProvider) SetAvailability(available bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.available = available
}
