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

// ExchangeRateService 汇率服务（简化版，用于兼容）
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

// SimpleExchangeRateResponse 简化版汇率API响应
type SimpleExchangeRateResponse struct {
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
	).Order("updated_at DESC").First(&rate)

	if result.Error == nil {
		return rate.Rate.InexactFloat64()
	}

	// 从API获取
	return s.fetchRateFromAPI(fromCurrency, toCurrency)
}

// fetchRateFromAPI 从API获取汇率
func (s *ExchangeRateService) fetchRateFromAPI(fromCurrency, toCurrency string) float64 {
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", fromCurrency)

	resp, err := s.client.Get(url)
	if err != nil {
		utils.Error("Failed to fetch exchange rate", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Error("Exchange rate API returned non-200 status", fmt.Errorf("status: %d", resp.StatusCode))
		return 0
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Error("Failed to read response body", err)
		return 0
	}

	var apiResp SimpleExchangeRateResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		utils.Error("Failed to unmarshal response", err)
		return 0
	}

	rate, ok := apiResp.Rates[toCurrency]
	if !ok {
		utils.Error("Currency not found in rates", fmt.Errorf("currency: %s", toCurrency))
		return 0
	}

	return rate
}

// UpdateRates 更新汇率（用于定时任务，兼容旧接口）
func (s *ExchangeRateService) UpdateRates() error {
	return s.UpdateExchangeRates()
}

// UpdateExchangeRates 更新汇率（用于定时任务）
func (s *ExchangeRateService) UpdateExchangeRates() error {
	// 获取所有启用的货币对
	var pairs []models.CurrencyPair
	if err := models.DB.Where("is_active = ?", 1).Find(&pairs).Error; err != nil {
		return err
	}

	for _, pair := range pairs {
		rate := s.fetchRateFromAPI(pair.FromCurrency, pair.ToCurrency)
		if rate == 0 {
			continue
		}

		// 获取旧汇率
		var oldRate decimal.Decimal
		var existing models.ExchangeRate
		if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
			pair.FromCurrency, pair.ToCurrency, 1).
			First(&existing).Error; err == nil {
			oldRate = existing.Rate
		}

		newRate := decimal.NewFromFloat(rate)
		changePercent := decimal.Zero
		if !oldRate.IsZero() {
			changePercent = newRate.Sub(oldRate).Div(oldRate).Mul(decimal.NewFromInt(100))
		}

		// 保存汇率
		exchangeRate := &models.ExchangeRate{
			FromCurrency:  pair.FromCurrency,
			ToCurrency:    pair.ToCurrency,
			Rate:          newRate,
			PreviousRate:  oldRate,
			ChangePercent: changePercent,
			DataSource:    "exchangerate-api",
			SourceType:    "api",
			ValidStatus:   1,
		}

		models.DB.Where("from_currency = ? AND to_currency = ?", pair.FromCurrency, pair.ToCurrency).
			Assign(exchangeRate).
			FirstOrCreate(exchangeRate)
	}

	return nil
}
