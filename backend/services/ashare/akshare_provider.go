package ashare

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// AKShareProvider AKShare数据源提供者
// AKShare是一个基于Python的开源财经数据接口库，这里通过HTTP API调用Python服务
type AKShareProvider struct {
	baseURL    string
	httpClient *http.Client
}

// NewAKShareProvider 创建AKShare数据源
func NewAKShareProvider(baseURL string) *AKShareProvider {
	if baseURL == "" {
		baseURL = "http://localhost:8000" // 默认Python服务地址
	}
	return &AKShareProvider{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ETFQuote ETF行情数据
type ETFQuote struct {
	Symbol         string          `json:"symbol"`
	Name           string          `json:"name"`
	CurrentPrice   decimal.Decimal `json:"current_price"`
	PreviousClose  decimal.Decimal `json:"previous_close"`
	Open           decimal.Decimal `json:"open"`
	High           decimal.Decimal `json:"high"`
	Low            decimal.Decimal `json:"low"`
	Volume         int64           `json:"volume"`
	Turnover       decimal.Decimal `json:"turnover"`
	PriceChange    decimal.Decimal `json:"price_change"`
	PriceChangePct decimal.Decimal `json:"price_change_pct"`
	BidPrice       decimal.Decimal `json:"bid_price"`
	AskPrice       decimal.Decimal `json:"ask_price"`
	BidVolume      int64           `json:"bid_volume"`
	AskVolume      int64           `json:"ask_volume"`
	UpdateTime     time.Time       `json:"update_time"`
}

// ETFHistoricalData ETF历史数据
type ETFHistoricalData struct {
	Symbol   string          `json:"symbol"`
	Date     time.Time       `json:"date"`
	Open     decimal.Decimal `json:"open"`
	High     decimal.Decimal `json:"high"`
	Low      decimal.Decimal `json:"low"`
	Close    decimal.Decimal `json:"close"`
	Volume   int64           `json:"volume"`
	Turnover decimal.Decimal `json:"turnover"`
	Amount   decimal.Decimal `json:"amount"` // 成交额
}

// ETFInfo ETF基本信息
type ETFInfo struct {
	Symbol        string          `json:"symbol"`
	Name          string          `json:"name"`
	FullName      string          `json:"full_name"`
	Exchange      string          `json:"exchange"`
	Benchmark     string          `json:"benchmark"`
	ManagementFee decimal.Decimal `json:"management_fee"`
	CustodyFee    decimal.Decimal `json:"custody_fee"`
	TotalScale    decimal.Decimal `json:"total_scale"`  // 总份额（亿份）
	NetValue      decimal.Decimal `json:"net_value"`    // 最新净值
	PremiumRate   decimal.Decimal `json:"premium_rate"` // 溢价率
}

// DividendData 分红数据
type DividendData struct {
	Symbol           string          `json:"symbol"`
	ExDividendDate   time.Time       `json:"ex_dividend_date"`   // 除息日
	DividendPerShare decimal.Decimal `json:"dividend_per_share"` // 每股分红
	DividendYield    decimal.Decimal `json:"dividend_yield"`     // 股息率
}

// GetETFList 获取A股ETF列表
func (p *AKShareProvider) GetETFList() ([]ETFInfo, error) {
	endpoint := fmt.Sprintf("%s/api/etf/list", p.baseURL)

	resp, err := p.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取ETF列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool      `json:"success"`
		Data    []ETFInfo `json:"data"`
		Error   string    `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API错误: %s", result.Error)
	}

	return result.Data, nil
}

// GetETFQuote 获取ETF实时行情
func (p *AKShareProvider) GetETFQuote(symbol string) (*ETFQuote, error) {
	endpoint := fmt.Sprintf("%s/api/etf/quote/%s", p.baseURL, symbol)

	resp, err := p.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取ETF行情失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool     `json:"success"`
		Data    ETFQuote `json:"data"`
		Error   string   `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API错误: %s", result.Error)
	}

	return &result.Data, nil
}

// GetETFQuotes 批量获取ETF行情
func (p *AKShareProvider) GetETFQuotes(symbols []string) (map[string]*ETFQuote, error) {
	endpoint := fmt.Sprintf("%s/api/etf/quotes", p.baseURL)

	// 构建查询参数
	params := url.Values{}
	for _, symbol := range symbols {
		params.Add("symbols", symbol)
	}

	resp, err := p.httpClient.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("批量获取ETF行情失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool                 `json:"success"`
		Data    map[string]*ETFQuote `json:"data"`
		Error   string               `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API错误: %s", result.Error)
	}

	return result.Data, nil
}

// GetHistoricalData 获取ETF历史数据
func (p *AKShareProvider) GetHistoricalData(symbol string, startDate, endDate time.Time) ([]ETFHistoricalData, error) {
	endpoint := fmt.Sprintf("%s/api/etf/historical/%s", p.baseURL, symbol)

	params := url.Values{}
	params.Set("start_date", startDate.Format("2006-01-02"))
	params.Set("end_date", endDate.Format("2006-01-02"))

	resp, err := p.httpClient.Get(endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("获取历史数据失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool                `json:"success"`
		Data    []ETFHistoricalData `json:"data"`
		Error   string              `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API错误: %s", result.Error)
	}

	return result.Data, nil
}

// GetDividendHistory 获取ETF分红历史
func (p *AKShareProvider) GetDividendHistory(symbol string) ([]DividendData, error) {
	endpoint := fmt.Sprintf("%s/api/etf/dividends/%s", p.baseURL, symbol)

	resp, err := p.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取分红历史失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool           `json:"success"`
		Data    []DividendData `json:"data"`
		Error   string         `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API错误: %s", result.Error)
	}

	return result.Data, nil
}

// ConvertToModel 将AKShare数据转换为模型
func (p *AKShareProvider) ConvertToModel(quote *ETFQuote) *models.AShareDividendETF {
	return &models.AShareDividendETF{
		Symbol:         quote.Symbol,
		Name:           quote.Name,
		CurrentPrice:   quote.CurrentPrice,
		PreviousClose:  quote.PreviousClose,
		PriceChange:    quote.PriceChange,
		PriceChangePct: quote.PriceChangePct,
		Volume:         quote.Volume,
		Turnover:       quote.Turnover,
		PriceUpdatedAt: quote.UpdateTime,
	}
}
