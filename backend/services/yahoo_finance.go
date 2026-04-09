package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"etf-insight/utils"
)

// YahooFinanceClient Yahoo Finance API客户端
type YahooFinanceClient struct {
	client      *http.Client
	baseURL     string
	rateLimiter chan struct{}
}

// NewYahooFinanceClient 创建新的Yahoo Finance客户端
func NewYahooFinanceClient() *YahooFinanceClient {
	// 创建HTTP传输配置
	transport := &http.Transport{}

	// 检查环境变量中的代理设置
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

	// 如果环境变量中有代理，使用它
	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(parsedURL)
			fmt.Printf("Using proxy for Yahoo Finance: %s\n", proxyURL)
		}
	}

	return &YahooFinanceClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		baseURL:     "https://query1.finance.yahoo.com",
		rateLimiter: make(chan struct{}, 3), // 限制并发数
	}
}

// QuoteResponse Yahoo Finance报价响应
type QuoteResponse struct {
	QuoteResponse struct {
		Result []struct {
			Symbol                     string  `json:"symbol"`
			ShortName                  string  `json:"shortName"`
			LongName                   string  `json:"longName"`
			RegularMarketPrice         float64 `json:"regularMarketPrice"`
			RegularMarketOpen          float64 `json:"regularMarketOpen"`
			RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
			RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
			RegularMarketVolume        int64   `json:"regularMarketVolume"`
			PreviousClose              float64 `json:"regularMarketPreviousClose"`
			MarketCap                  int64   `json:"marketCap"`
			FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
			FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
			AverageVolume              int64   `json:"averageVolume"`
			Beta                       float64 `json:"beta"`
			TrailingPE                 float64 `json:"trailingPE"`
			DividendYield              float64 `json:"dividendYield"`
			TrailingAnnualDividendRate float64 `json:"trailingAnnualDividendRate"`
			Currency                   string  `json:"currency"`
			Exchange                   string  `json:"exchange"`
			QuoteType                  string  `json:"quoteType"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"quoteResponse"`
}

// HistoricalData 历史数据响应
type HistoricalData struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency             string  `json:"currency"`
				Symbol               string  `json:"symbol"`
				ExchangeName         string  `json:"exchangeName"`
				InstrumentType       string  `json:"instrumentType"`
				FirstTradeDate       int64   `json:"firstTradeDate"`
				RegularMarketTime    int64   `json:"regularMarketTime"`
				Gmtoffset            int     `json:"gmtoffset"`
				Timezone             string  `json:"timezone"`
				ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
				RegularMarketPrice   float64 `json:"regularMarketPrice"`
				ChartPreviousClose   float64 `json:"chartPreviousClose"`
				PreviousClose        float64 `json:"previousClose"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
				Adjclose []struct {
					Adjclose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

// QuoteData 实时报价数据
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
}

// HistoricalPrice 历史价格数据
type HistoricalPrice struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// GetQuotes 获取多个股票的实时报价
func (c *YahooFinanceClient) GetQuotes(symbols []string) ([]QuoteData, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	symbolsStr := strings.Join(symbols, ",")
	encodedSymbols := url.QueryEscape(symbolsStr)

	url := fmt.Sprintf("%s/v7/finance/quote?symbols=%s", c.baseURL, encodedSymbols)

	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var quoteResp QuoteResponse
	if err := json.Unmarshal(body, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if quoteResp.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("API error: %v", quoteResp.QuoteResponse.Error)
	}

	var quotes []QuoteData
	for _, result := range quoteResp.QuoteResponse.Result {
		change := result.RegularMarketPrice - result.PreviousClose
		changePercent := 0.0
		if result.PreviousClose > 0 {
			changePercent = (change / result.PreviousClose) * 100
		}

		name := result.LongName
		if name == "" {
			name = result.ShortName
		}

		quotes = append(quotes, QuoteData{
			Symbol:           result.Symbol,
			Name:             name,
			CurrentPrice:     result.RegularMarketPrice,
			OpenPrice:        result.RegularMarketOpen,
			DayHigh:          result.RegularMarketDayHigh,
			DayLow:           result.RegularMarketDayLow,
			Volume:           result.RegularMarketVolume,
			PreviousClose:    result.PreviousClose,
			Change:           change,
			ChangePercent:    changePercent,
			MarketCap:        result.MarketCap,
			FiftyTwoWeekHigh: result.FiftyTwoWeekHigh,
			FiftyTwoWeekLow:  result.FiftyTwoWeekLow,
			AverageVolume:    result.AverageVolume,
			Beta:             result.Beta,
			PERatio:          result.TrailingPE,
			DividendYield:    result.DividendYield * 100, // 转换为百分比
			Currency:         result.Currency,
			Exchange:         result.Exchange,
		})
	}

	return quotes, nil
}

// GetQuote 获取单个股票的实时报价
func (c *YahooFinanceClient) GetQuote(symbol string) (*QuoteData, error) {
	quotes, err := c.GetQuotes([]string{symbol})
	if err != nil {
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	return &quotes[0], nil
}

// GetHistoricalData 获取历史数据
func (c *YahooFinanceClient) GetHistoricalData(symbol string, period string, interval string) ([]HistoricalPrice, error) {
	// 转换period为时间范围
	now := time.Now()
	var startTime time.Time

	switch period {
	case "1d":
		startTime = now.AddDate(0, 0, -1)
	case "5d":
		startTime = now.AddDate(0, 0, -5)
	case "1mo":
		startTime = now.AddDate(0, -1, 0)
	case "3mo":
		startTime = now.AddDate(0, -3, 0)
	case "6mo":
		startTime = now.AddDate(0, -6, 0)
	case "1y":
		startTime = now.AddDate(-1, 0, 0)
	case "2y":
		startTime = now.AddDate(-2, 0, 0)
	case "5y":
		startTime = now.AddDate(-5, 0, 0)
	case "10y":
		startTime = now.AddDate(-10, 0, 0)
	case "ytd":
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case "max":
		startTime = now.AddDate(-20, 0, 0)
	default:
		startTime = now.AddDate(-1, 0, 0)
	}

	// 默认使用日线
	if interval == "" {
		interval = "1d"
	}

	url := fmt.Sprintf("%s/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s&events=div",
		c.baseURL,
		symbol,
		startTime.Unix(),
		now.Unix(),
		interval,
	)

	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var histData HistoricalData
	if err := json.Unmarshal(body, &histData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if histData.Chart.Error != nil {
		return nil, fmt.Errorf("API error: %v", histData.Chart.Error)
	}

	if len(histData.Chart.Result) == 0 {
		return nil, fmt.Errorf("no historical data found for symbol: %s", symbol)
	}

	result := histData.Chart.Result[0]
	timestamps := result.Timestamp
	quote := result.Indicators.Quote[0]

	var prices []HistoricalPrice
	for i, ts := range timestamps {
		if i >= len(quote.Open) || i >= len(quote.High) || i >= len(quote.Low) || i >= len(quote.Close) || i >= len(quote.Volume) {
			continue
		}

		// 跳过无效数据
		if quote.Open[i] == 0 && quote.Close[i] == 0 {
			continue
		}

		prices = append(prices, HistoricalPrice{
			Date:   time.Unix(ts, 0),
			Open:   quote.Open[i],
			High:   quote.High[i],
			Low:    quote.Low[i],
			Close:  quote.Close[i],
			Volume: quote.Volume[i],
		})
	}

	return prices, nil
}

// makeRequest 发送HTTP请求
func (c *YahooFinanceClient) makeRequest(url string) (*http.Response, error) {
	// 速率限制
	c.rateLimiter <- struct{}{}
	defer func() { <-c.rateLimiter }()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头 - 模拟真实浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 添加延迟以避免触发限流
	time.Sleep(500 * time.Millisecond)

	return resp, nil
}

// RetryGetQuote 带重试的获取报价
func (c *YahooFinanceClient) RetryGetQuote(symbol string, maxRetries int) (*QuoteData, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		quote, err := c.GetQuote(symbol)
		if err == nil {
			return quote, nil
		}
		lastErr = err
		utils.Warn("Failed to get quote", err, "symbol", symbol, "attempt", i+1, "maxRetries", maxRetries)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// RetryGetHistoricalData 带重试的获取历史数据
func (c *YahooFinanceClient) RetryGetHistoricalData(symbol string, period string, interval string, maxRetries int) ([]HistoricalPrice, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		data, err := c.GetHistoricalData(symbol, period, interval)
		if err == nil {
			return data, nil
		}
		lastErr = err
		utils.Warn("Failed to get historical data", err, "symbol", symbol, "attempt", i+1, "maxRetries", maxRetries)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// ParsePeriodDays 将period转换为天数
func ParsePeriodDays(period string) int {
	switch period {
	case "1d":
		return 1
	case "5d":
		return 5
	case "1mo":
		return 30
	case "3mo":
		return 90
	case "6mo":
		return 180
	case "1y":
		return 365
	case "2y":
		return 730
	case "5y":
		return 1825
	case "10y":
		return 3650
	default:
		return 365
	}
}
