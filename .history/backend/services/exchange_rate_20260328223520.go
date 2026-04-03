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

// GetRate 获取汇率
func (s *ExchangeRateService) GetRate(fromCurrency, toCurrency string) float64 {
	if fromCurrency == toCurrency {
		return 1.0
	}

	// 先尝试从数据库获取
	var rate models.ExchangeRate
	result := models.DB.Where(
		"from_currency = ? AND to_currency = ?",
		fromCurrency, toCurrency,
	).Order("rate_date DESC").First(&rate)

	if result.Error == nil {
		return rate.Rate.InexactFloat64()
	}

	// 使用默认汇率
	return s.getDefaultRate(fromCurrency, toCurrency)
}

// Convert 货币转换
func (s *ExchangeRateService) Convert(amount decimal.Decimal, fromCurrency, toCurrency string) decimal.Decimal {
	if fromCurrency == toCurrency {
		return amount
	}

	rate := s.GetRate(fromCurrency, toCurrency)
	return amount.Mul(decimal.NewFromFloat(rate))
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
	now := time.Now()
	for fromCurrency, toRates := range rates {
		for toCurrency, rate := range toRates {
			// 查找是否已存在今日汇率
			var existing models.ExchangeRate
			result := models.DB.Where(
				"from_currency = ? AND to_currency = ? AND rate_date = ?",
				fromCurrency, toCurrency, now.Format("2006-01-02"),
			).First(&existing)

			if result.Error == nil {
				// 更新现有记录
				existing.Rate = decimal.NewFromFloat(rate)
				existing.DataSource = "api"
				models.DB.Save(&existing)
			} else {
				// 创建新记录
				newRate := models.ExchangeRate{
					FromCurrency: fromCurrency,
					ToCurrency:   toCurrency,
					Rate:         decimal.NewFromFloat(rate),
					RateDate:     now.Format("2006-01-02"),
					DataSource:   "api",
				}
				models.DB.Create(&newRate)
			}

			utils.Info("Updated rate", "from", fromCurrency, "to", toCurrency, "rate", rate)
		}
	}

	utils.Info("Exchange rate update completed")
	return nil
}

// fetchFromFreeAPI 从免费API获取汇率
func (s *ExchangeRateService) fetchFromFreeAPI() (map[string]map[string]float64, error) {
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

	// 构建汇率矩阵
	rates := make(map[string]map[string]float64)

	// USD基础汇率
	usdCny := apiResp.Rates["CNY"]
	usdHkd := apiResp.Rates["HKD"]

	rates["USD"] = map[string]float64{
		"USD": 1.0,
		"CNY": usdCny,
		"HKD": usdHkd,
	}

	rates["CNY"] = map[string]float64{
		"CNY": 1.0,
		"USD": 1.0 / usdCny,
		"HKD": usdHkd / usdCny,
	}

	rates["HKD"] = map[string]float64{
		"HKD": 1.0,
		"USD": 1.0 / usdHkd,
		"CNY": usdCny / usdHkd,
	}

	return rates, nil
}

// getDefaultRates 获取默认汇率
func (s *ExchangeRateService) getDefaultRates() map[string]map[string]float64 {
	return map[string]map[string]float64{
		"USD": {
			"USD": 1.0,
			"CNY": 7.2,
			"HKD": 7.8,
		},
		"CNY": {
			"CNY": 1.0,
			"USD": 0.138889,
			"HKD": 1.083333,
		},
		"HKD": {
			"HKD": 1.0,
			"USD": 0.128205,
			"CNY": 0.923077,
		},
	}
}

// getDefaultRate 获取默认汇率
func (s *ExchangeRateService) getDefaultRate(fromCurrency, toCurrency string) float64 {
	rates := s.getDefaultRates()
	if fromRates, ok := rates[fromCurrency]; ok {
		if rate, ok := fromRates[toCurrency]; ok {
			return rate
		}
	}
	return 1.0
}

// GetHistory 获取汇率历史
func (s *ExchangeRateService) GetHistory(fromCurrency, toCurrency string, days int) ([]map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	var rates []models.ExchangeRate
	result := models.DB.Where(
		"from_currency = ? AND to_currency = ? AND rate_date >= ? AND rate_date <= ?",
		fromCurrency, toCurrency, startDate, endDate,
	).Order("rate_date ASC").Find(&rates)

	if result.Error() != nil {
		return nil, result.Error()
	}

	var history []map[string]interface{}
	for _, rate := range rates {
		history = append(history, map[string]interface{}{
			"date":   rate.RateDate,
			"rate":   rate.Rate.InexactFloat64(),
			"source": rate.DataSource,
		})
	}

	return history, nil
}

// CalculateCrossRate 计算交叉汇率
func (s *ExchangeRateService) CalculateCrossRate(fromCurrency, toCurrency string) float64 {
	if fromCurrency == toCurrency {
		return 1.0
	}

	// 尝试直接获取
	directRate := s.GetRate(fromCurrency, toCurrency)
	if directRate != 1.0 {
		return directRate
	}

	// 通过USD计算交叉汇率
	fromToUSD := s.GetRate(fromCurrency, "USD")
	usdToTarget := s.GetRate("USD", toCurrency)

	return fromToUSD * usdToTarget
}
