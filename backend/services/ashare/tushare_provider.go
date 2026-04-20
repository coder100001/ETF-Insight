package ashare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// TuShareProvider TuShare数据源提供者
// TuShare是一个提供中国股票、期货等金融数据的Python库
type TuShareProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewTuShareProvider 创建TuShare数据源
func NewTuShareProvider(apiKey, baseURL string) *TuShareProvider {
	if baseURL == "" {
		baseURL = "https://api.tushare.pro"
	}
	return &TuShareProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TuShareRequest TuShare API请求
type TuShareRequest struct {
	APIName string                 `json:"api_name"`
	Token   string                 `json:"token"`
	Params  map[string]interface{} `json:"params"`
	Fields  string                 `json:"fields,omitempty"`
}

// TuShareResponse TuShare API响应
type TuShareResponse struct {
	RequestID string      `json:"request_id"`
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      TuShareData `json:"data"`
}

// TuShareData TuShare数据
type TuShareData struct {
	Fields []string        `json:"fields"`
	Items  [][]interface{} `json:"items"`
}

// FundBasic 基金基础信息
type FundBasic struct {
	TSCode     string `json:"ts_code"`     // 基金代码
	Name       string `json:"name"`        // 基金名称
	Management string `json:"management"`  // 管理人
	Custodian  string `json:"custodian"`   // 托管人
	FundType   string `json:"fund_type"`   // 投资类型
	Benchmark  string `json:"benchmark"`   // 业绩比较基准
	Status     string `json:"status"`      // 状态
	ListDate   string `json:"list_date"`   // 上市日期
	DelistDate string `json:"delist_date"` // 退市日期
}

// FundNav 基金净值
type FundNav struct {
	TSCode   string          `json:"ts_code"`
	NavDate  string          `json:"nav_date"`
	UnitNav  decimal.Decimal `json:"unit_nav"`  // 单位净值
	AccumNav decimal.Decimal `json:"accum_nav"` // 累计净值
	AdjNav   decimal.Decimal `json:"adj_nav"`   // 复权净值
}

// FundDaily 基金日线行情
type FundDaily struct {
	TSCode    string          `json:"ts_code"`
	TradeDate string          `json:"trade_date"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	PreClose  decimal.Decimal `json:"pre_close"`
	Change    decimal.Decimal `json:"change"`
	PctChg    decimal.Decimal `json:"pct_chg"`
	Vol       decimal.Decimal `json:"vol"`
	Amount    decimal.Decimal `json:"amount"`
}

// callAPI 调用TuShare API
func (p *TuShareProvider) callAPI(apiName string, params map[string]interface{}) (*TuShareResponse, error) {
	req := TuShareRequest{
		APIName: apiName,
		Token:   p.apiKey,
		Params:  params,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := p.httpClient.Post(
		p.baseURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %w", err)
	}
	defer resp.Body.Close()

	var result TuShareResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API错误: %s", result.Msg)
	}

	return &result, nil
}

// GetFundBasic 获取基金基础信息
func (p *TuShareProvider) GetFundBasic(market string) ([]FundBasic, error) {
	params := map[string]interface{}{}
	if market != "" {
		params["market"] = market
	}

	resp, err := p.callAPI("fund_basic", params)
	if err != nil {
		return nil, err
	}

	var funds []FundBasic
	for _, item := range resp.Data.Items {
		if len(item) >= 8 {
			fund := FundBasic{
				TSCode:     getString(item, 0),
				Name:       getString(item, 1),
				Management: getString(item, 2),
				Custodian:  getString(item, 3),
				FundType:   getString(item, 4),
				Benchmark:  getString(item, 5),
				Status:     getString(item, 6),
				ListDate:   getString(item, 7),
			}
			funds = append(funds, fund)
		}
	}

	return funds, nil
}

// GetFundNav 获取基金净值
func (p *TuShareProvider) GetFundNav(tsCode string, startDate, endDate string) ([]FundNav, error) {
	params := map[string]interface{}{
		"ts_code": tsCode,
	}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}

	resp, err := p.callAPI("fund_nav", params)
	if err != nil {
		return nil, err
	}

	var navs []FundNav
	for _, item := range resp.Data.Items {
		if len(item) >= 4 {
			nav := FundNav{
				TSCode:  getString(item, 0),
				NavDate: getString(item, 1),
			}
			if val, ok := item[2].(float64); ok {
				nav.UnitNav = decimal.NewFromFloat(val)
			}
			if val, ok := item[3].(float64); ok {
				nav.AccumNav = decimal.NewFromFloat(val)
			}
			navs = append(navs, nav)
		}
	}

	return navs, nil
}

// GetFundDaily 获取基金日线行情
func (p *TuShareProvider) GetFundDaily(tsCode string, startDate, endDate string) ([]FundDaily, error) {
	params := map[string]interface{}{
		"ts_code": tsCode,
	}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}

	resp, err := p.callAPI("fund_daily", params)
	if err != nil {
		return nil, err
	}

	var dailies []FundDaily
	for _, item := range resp.Data.Items {
		if len(item) >= 11 {
			daily := FundDaily{
				TSCode:    getString(item, 0),
				TradeDate: getString(item, 1),
			}
			if val, ok := item[2].(float64); ok {
				daily.Open = decimal.NewFromFloat(val)
			}
			if val, ok := item[3].(float64); ok {
				daily.High = decimal.NewFromFloat(val)
			}
			if val, ok := item[4].(float64); ok {
				daily.Low = decimal.NewFromFloat(val)
			}
			if val, ok := item[5].(float64); ok {
				daily.Close = decimal.NewFromFloat(val)
			}
			if val, ok := item[6].(float64); ok {
				daily.PreClose = decimal.NewFromFloat(val)
			}
			if val, ok := item[7].(float64); ok {
				daily.Change = decimal.NewFromFloat(val)
			}
			if val, ok := item[8].(float64); ok {
				daily.PctChg = decimal.NewFromFloat(val)
			}
			if val, ok := item[9].(float64); ok {
				daily.Vol = decimal.NewFromFloat(val)
			}
			if val, ok := item[10].(float64); ok {
				daily.Amount = decimal.NewFromFloat(val)
			}
			dailies = append(dailies, daily)
		}
	}

	return dailies, nil
}

// getString 安全获取字符串
func getString(item []interface{}, index int) string {
	if index < len(item) {
		if val, ok := item[index].(string); ok {
			return val
		}
	}
	return ""
}
