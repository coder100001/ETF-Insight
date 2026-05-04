package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// ExchangeRateService 汇率服务
type ExchangeRateService struct {
	client *http.Client
}

// NewExchangeRateService 创建新的汇率服务
func NewExchangeRateService() *ExchangeRateService {
	return &ExchangeRateService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ExchangeRateAPIResponse 汇率API响应
type ExchangeRateAPIResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
	Date  string             `json:"date"`
}

// GetRate 获取汇率 (返回 decimal.Decimal 以满足金融计算精度要求)
func (s *ExchangeRateService) GetRate(fromCurrency, toCurrency string) decimal.Decimal {
	if fromCurrency == toCurrency {
		return decimal.NewFromInt(1)
	}

	var rate models.ExchangeRate
	result := models.DB.Where(
		"from_currency = ? AND to_currency = ?",
		fromCurrency, toCurrency,
	).Order("rate_date DESC").First(&rate)

	if result.Error == nil {
		return rate.Rate
	}

	return s.getDefaultRate(fromCurrency, toCurrency)
}

// Convert 货币转换
func (s *ExchangeRateService) Convert(amount decimal.Decimal, fromCurrency, toCurrency string) decimal.Decimal {
	if fromCurrency == toCurrency {
		return amount
	}

	rate := s.GetRate(fromCurrency, toCurrency)
	return amount.Mul(rate)
}

// UpdateRates 更新汇率
func (s *ExchangeRateService) UpdateRates() error {
	utils.Info("Starting exchange rate update...")

	// 从免费API获取汇率
	rates, err := s.fetchFromFreeAPI()
	if err != nil {
		utils.Warn("Failed to fetch from free API, using default rates", err)
		rates = s.getDefaultRates()
	}

	// 保存到数据库
	for fromCurrency, toRates := range rates {
		for toCurrency, rate := range toRates {
			exchangeRate := models.ExchangeRate{
				FromCurrency: fromCurrency,
				ToCurrency:   toCurrency,
				Rate:         rate,
				DataSource:   "api",
			}
			models.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "from_currency"}, {Name: "to_currency"}, {Name: "rate_date"}},
				DoUpdates: clause.AssignmentColumns([]string{"rate"}),
			}).Create(&exchangeRate)

			utils.Info("Updated rate", "from", fromCurrency, "to", toCurrency, "rate", rate)
		}
	}

	utils.Info("Exchange rate update completed")
	return nil
}

// fetchFromFreeAPI 从免费API获取汇率
func (s *ExchangeRateService) fetchFromFreeAPI() (map[string]map[string]decimal.Decimal, error) {
	url := "https://api.exchangerate-api.com/v4/latest/USD"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp ExchangeRateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	rates := make(map[string]map[string]decimal.Decimal)

	usdCny := decimal.NewFromFloat(apiResp.Rates["CNY"])
	usdHkd := decimal.NewFromFloat(apiResp.Rates["HKD"])
	one := decimal.NewFromInt(1)

	rates["USD"] = map[string]decimal.Decimal{
		"USD": one,
		"CNY": usdCny,
		"HKD": usdHkd,
	}

	rates["CNY"] = map[string]decimal.Decimal{
		"CNY": one,
		"USD": one.Div(usdCny),
		"HKD": usdHkd.Div(usdCny),
	}

	rates["HKD"] = map[string]decimal.Decimal{
		"HKD": one,
		"USD": one.Div(usdHkd),
		"CNY": usdCny.Div(usdHkd),
	}

	return rates, nil
}

// getDefaultRates 获取默认汇率
func (s *ExchangeRateService) getDefaultRates() map[string]map[string]decimal.Decimal {
	usdCny := decimal.NewFromFloat(7.25)
	usdHkd := decimal.NewFromFloat(7.83)

	return map[string]map[string]decimal.Decimal{
		"USD": {
			"USD": decimal.NewFromInt(1),
			"CNY": usdCny,
			"HKD": usdHkd,
		},
		"CNY": {
			"CNY": decimal.NewFromInt(1),
			"USD": decimal.NewFromInt(1).Div(usdCny),
			"HKD": usdHkd.Div(usdCny),
		},
		"HKD": {
			"HKD": decimal.NewFromInt(1),
			"USD": decimal.NewFromInt(1).Div(usdHkd),
			"CNY": usdCny.Div(usdHkd),
		},
	}
}

// getDefaultRate 获取默认汇率
func (s *ExchangeRateService) getDefaultRate(fromCurrency, toCurrency string) decimal.Decimal {
	rates := s.getDefaultRates()
	if fromRates, ok := rates[fromCurrency]; ok {
		if rate, ok := fromRates[toCurrency]; ok {
			return rate
		}
	}
	return decimal.NewFromInt(1)
}

// GetHistory 获取汇率历史
func (s *ExchangeRateService) GetHistory(fromCurrency, toCurrency string, days int) ([]map[string]interface{}, error) {
	// 简化处理，返回空历史
	return []map[string]interface{}{}, nil
}

// CalculateCrossRate 计算交叉汇率
func (s *ExchangeRateService) CalculateCrossRate(fromCurrency, toCurrency string) decimal.Decimal {
	if fromCurrency == toCurrency {
		return decimal.NewFromInt(1)
	}

	directRate := s.GetRate(fromCurrency, toCurrency)
	one := decimal.NewFromInt(1)
	if !directRate.Equal(one) {
		return directRate
	}

	fromToUSD := s.GetRate(fromCurrency, "USD")
	usdToTarget := s.GetRate("USD", toCurrency)

	return fromToUSD.Mul(usdToTarget)
}
