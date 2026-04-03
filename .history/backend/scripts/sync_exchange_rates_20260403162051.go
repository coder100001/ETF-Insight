package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// FrankfurterAPIResponse Frankfurter API响应
type FrankfurterAPIResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
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

	client := &http.Client{Timeout: 30 * time.Second}

	// 获取 USD 为基础的所有汇率
	fmt.Println("Fetching USD base rates...")
	usdRates, err := fetchUSDRates(client)
	if err != nil {
		fmt.Printf("Failed to fetch USD rates: %v\n", err)
		return
	}

	// 同步美元兑主要货币
	syncRate("USD", "CNY", usdRates["CNY"])
	syncRate("USD", "HKD", usdRates["HKD"])
	syncRate("USD", "EUR", usdRates["EUR"])
	syncRate("USD", "GBP", usdRates["GBP"])
	syncRate("USD", "JPY", usdRates["JPY"])
	syncRate("USD", "CAD", usdRates["CAD"])
	syncRate("USD", "AUD", usdRates["AUD"])
	syncRate("USD", "CHF", usdRates["CHF"])
	syncRate("USD", "SGD", usdRates["SGD"])

	// 计算反向汇率
	if usdCNY, ok := usdRates["CNY"]; ok && usdCNY > 0 {
		syncRate("CNY", "USD", 1.0/usdCNY)
	}
	if usdHKD, ok := usdRates["HKD"]; ok && usdHKD > 0 {
		syncRate("HKD", "USD", 1.0/usdHKD)
	}

	// 计算美元指数
	fmt.Println("\nCalculating US Dollar Index...")
	calculateUSDIndex(usdRates)

	fmt.Println("\nSync completed!")
}

// fetchUSDRates 获取 USD 为基础的汇率
func fetchUSDRates(client *http.Client) (map[string]float64, error) {
	url := "https://api.frankfurter.app/latest?from=USD&to=CNY,HKD,EUR,GBP,JPY,CAD,AUD,CHF,SGD"

	resp, err := client.Get(url)
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

	var apiResp FrankfurterAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.Rates, nil
}

// syncRate 同步单个汇率
func syncRate(from, to string, rate float64) {
	if rate <= 0 {
		fmt.Printf("  Skipping %s/%s: invalid rate\n", from, to)
		return
	}

	// 获取旧汇率
	var oldRate decimal.Decimal
	var existing models.ExchangeRate
	if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
		from, to, 1).
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
		FromCurrency:  from,
		ToCurrency:    to,
		Rate:          newRate,
		PreviousRate:  oldRate,
		ChangePercent: changePercent,
		DataSource:    "frankfurter",
		SourceType:    "api",
		ValidStatus:   1,
		SyncedAt:      &[]time.Time{time.Now()}[0],
		ExpiresAt:     &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
	}

	if err := models.DB.Where("from_currency = ? AND to_currency = ?", from, to).
		Assign(exchangeRate).
		FirstOrCreate(exchangeRate).Error; err != nil {
		fmt.Printf("  Failed to save %s/%s: %v\n", from, to, err)
		return
	}

	fmt.Printf("  %s/%s = %.4f (change: %s%%)\n", from, to, rate, changePercent.StringFixed(2))
}

// calculateUSDIndex 计算美元指数
func calculateUSDIndex(rates map[string]float64) {
	// 美元指数权重
	weights := map[string]float64{
		"EUR": 0.576,
		"JPY": 0.136,
		"GBP": 0.119,
		"CAD": 0.091,
		"SEK": 0.042,
		"CHF": 0.036,
	}

	// 计算美元指数
	index := 50.14348112
	for currency, weight := range weights {
		rate, ok := rates[currency]
		if !ok || rate <= 0 {
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

	fmt.Printf("  USD Index: %.2f\n", index)
}

// pow 计算幂
func pow(x, y float64) float64 {
	if x <= 0 {
		return 0
	}
	// 使用自然对数计算幂
	return exp(y * ln(x))
}

// exp 计算 e^x
func exp(x float64) float64 {
	result := 1.0
	term := 1.0
	for i := 1; i < 20