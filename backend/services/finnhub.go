package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// FinnhubClient Finnhub API客户端
type FinnhubClient struct {
	apiKey      string
	client      *http.Client
	baseURL     string
	rateLimiter chan struct{}
}

// FinnhubQuote Finnhub报价响应
type FinnhubQuote struct {
	Symbol    string  `json:"symbol"`
	Current   float64 `json:"c"`  // 当前价格
	Change    float64 `json:"d"`  // 价格变动
	Percent   float64 `json:"dp"` // 变动百分比
	High      float64 `json:"h"`  // 当日最高
	Low       float64 `json:"l"`  // 当日最低
	Open      float64 `json:"o"`  // 开盘价
	Previous  float64 `json:"pc"` // 前收盘价
	Timestamp int64   `json:"t"`  // 时间戳
}

// NewFinnhubClient 创建新的Finnhub客户端
func NewFinnhubClient() *FinnhubClient {
	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		// 未配置 Finnhub API Key，客户端将无法使用
		fmt.Println("⚠️ FINNHUB_API_KEY 环境变量未设置，Finnhub 数据源不可用")
	}

	// 创建HTTP传输配置，支持代理
	transport := &http.Transport{}
	proxyURL := os.Getenv("HTTP_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("http_proxy")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("https_proxy")
	}

	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(parsedURL)
		}
	}

	return &FinnhubClient{
		apiKey: apiKey,
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		baseURL:     "https://finnhub.io/api/v1",
		rateLimiter: make(chan struct{}, 60), // Finnhub免费版每秒60次调用
	}
}

// GetQuote 获取单只股票报价
func (c *FinnhubClient) GetQuote(symbol string) (*FinnhubQuote, error) {
	// 速率限制
	c.rateLimiter <- struct{}{}
	defer func() { <-c.rateLimiter }()

	url := fmt.Sprintf("%s/quote?symbol=%s&token=%s", c.baseURL, symbol, c.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var quote FinnhubQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	quote.Symbol = symbol

	// 添加延迟以遵守速率限制
	time.Sleep(20 * time.Millisecond)

	return &quote, nil
}

// GetQuotes 批量获取多只股票报价
func (c *FinnhubClient) GetQuotes(symbols []string) ([]FinnhubQuote, error) {
	var quotes []FinnhubQuote

	for _, symbol := range symbols {
		quote, err := c.GetQuote(symbol)
		if err != nil {
			fmt.Printf("Warning: failed to get quote for %s: %v\n", symbol, err)
			continue
		}
		quotes = append(quotes, *quote)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("failed to get any quotes")
	}

	return quotes, nil
}

// ToQuoteData 将FinnhubQuote转换为通用的QuoteData格式
func (q *FinnhubQuote) ToQuoteData() *QuoteData {
	return &QuoteData{
		Symbol:        q.Symbol,
		CurrentPrice:  q.Current,
		OpenPrice:     q.Open,
		DayHigh:       q.High,
		DayLow:        q.Low,
		PreviousClose: q.Previous,
		Change:        q.Change,
		ChangePercent: q.Percent,
		Volume:        0, // Finnhub免费版不提供成交量
	}
}
