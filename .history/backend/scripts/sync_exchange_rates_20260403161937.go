package main

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

// ExchangeRateAPIResponse 汇率API响应
type ExchangeRateAPIResponse struct {
	Result   string             `json:"result"`
	BaseCode string             `json:"base_code"`
	Rates    map[string]float64 `json:"conversion_rates"`
}

func main() {
	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}

	// 初始化汇率表
	if err := models.InitExchangeRateTables(); err != nil {
		fmt.Printf("Failed to initialize exchange rate tables: %v\n", err)
		return
	}

	// 初始化货币对
	if err := models.InitDefaultCurrencyPairs(); err != nil {
		fmt.Printf("Failed to initialize currency pairs: %v\n", err)
		return
	}

	// 定义需要同步的货币对
	pairs := []struct {
		From string
		To   string
	}{
		{"USD", "CNY"},  // 美元兑人民币
		{"CNY", "USD"},  // 人民币兑美元
		{"USD", "HKD"},  // 美元兑港币
		{"HKD", "USD"},  // 港币兑美元
		{"USD", "CNH"},  // 美元兑离岸人民币
	}

	client := &http.Client{Timeout: 30 * time.Second}

	for _, pair := range pairs {
		fmt.Printf("Syncing %s/%s...\n", pair.From, pair.To)

		// 从API获取汇率
		rate, err := fetchRate(client, pair.From, pair.To)
		if err != nil {
			fmt.Printf("  Failed to fetch rate: %v\n", err)
			continue
		}

		// 获取旧汇率
		var oldRate decimal.Decimal
		var existing models.ExchangeRate
		if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
			pair.From, pair.To, 1).
			First(&existing).Error; err == nil {
			oldRate = existing.Rate
		}

		newRate := decimal.NewFromFloat(rate)
		changePercent := decimal.Zero
		if !oldRate.IsZero() {
			changePercent = newRate.Sub(oldRate).Div(oldRate).Mul(decimal.NewFromInt(100))
		}

		// 保存到数据库
		exchangeRate := &models.ExchangeRate{
			FromCurrency:  pair.From,
			ToCurrency:    pair.To,
			Rate:          newRate,
			PreviousRate:  oldRate,
			ChangePercent: changePercent,
			DataSource:    "exchangerate-api",
			SourceType:    "api",
			ValidStatus:   1,
			SyncedAt:      &[]time.Time{time.Now()}[0],
			ExpiresAt:     &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
		}

		if err := models.DB.Where("from_currency = ? AND to_currency = ?", pair.From, pair.To).
			Assign(exchangeRate).
			FirstOrCreate(exchangeRate).Error; err != nil {
			fmt.Printf("  Failed to save rate: %v\n", err)
			continue
		}

		fmt.Printf("  Success: %s/%s = %f (change: %s%%)\n", pair.From, pair.To, rate, changePercent.StringFixed(2))
	}

	// 获取美元指数 (使用 USD/EUR, USD/GBP, USD/JPY 等计算)
	fmt.Println("\nCalculating US Dollar Index...")
	calculateUSDIndex(client)

	fmt.Println("\nSync completed!")
}

// fetchRate 从API获取汇率
func fetchRate(client *http.Client, fromCurrency, toCurrency string) (float64, error) {
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", fromCurrency)

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var apiResp ExchangeRateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return 0, err
	}

	rate, ok := apiResp.Rates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("currency %s not found", toCurrency)
	}

	return rate, nil
}

// calculateUSDIndex 计算美元指数
func calculateUSDIndex(client *http.Client) {
	// 美元指数通常基于 EUR, JPY, GBP, CAD, SEK, CHF
	// 这里使用简化的计算方式
	weights := map[string]float64{
		"EUR": 0.576,
		"JPY": 0.136,
		"GBP": 0.119,
		"CAD": 0.091,
		"SEK": 0.042,
		"CHF": 0.036,
	}

	url := "https://api.exchangerate-api.com/v4/latest/USD"
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("  Failed to fetch USD rates: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("  Failed to read response: %v\n", err)
		return
	}

	var apiResp ExchangeRateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("  Failed to parse response: %v\n", err)
		return
	}

	// 计算美元指数 (几何加权平均)
	index := 50.14348112
	for currency, weight := range weights {
		rate, ok := apiResp.Rates[currency]
		if !ok {
			continue
		}
		index *= pow(rate, -weight)
	}

	// 保存美元指数
	usdIndex := &models.ExchangeRate{
		FromCurrency:  "USD",
		ToCurrency:    "INDEX",
		Rate:          decimal.NewFromFloat(index),
		PreviousRate:  decimal.Zero,
		ChangePercent: decimal.Zero,
		DataSource:    "calculated",
		SourceType:    "calculated",
		ValidStatus:   1,
		SyncedAt:      &[]time.Time{time.Now()}[0],
		ExpiresAt:     &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
	}

	if err := models.DB.Where("from_currency = ? AND to_currency = ?", "USD", "INDEX").
		Assign(usdIndex).
		FirstOrCreate(usdIndex).Error; err != nil {
		fmt.Printf("  Failed to save USD index: %v\n", err)
		return
	}

	fmt.Printf("  USD Index: %f\n", index)
}

// pow 计算幂
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y*1000); i++ {
		result *= x
	}
	return result
}
