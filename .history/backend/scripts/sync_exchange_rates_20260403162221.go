package main

import (
	"fmt"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

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

	fmt.Println("Syncing exchange rates...")

	// 模拟汇率数据（实际应用中应该从API获取）
	rates := []struct {
		From string
		To   string
		Rate float64
	}{
		{"USD", "CNY", 7.2450},  // 美元兑人民币
		{"CNY", "USD", 0.1380},  // 人民币兑美元
		{"USD", "HKD", 7.7980},  // 美元兑港币
		{"HKD", "USD", 0.1282},  // 港币兑美元
		{"USD", "EUR", 0.9230},  // 美元兑欧元
		{"EUR", "USD", 1.0835},  // 欧元兑美元
		{"USD", "GBP", 0.7850},  // 美元兑英镑
		{"GBP", "USD", 1.2739},  // 英镑兑美元
		{"USD", "JPY", 151.4500}, // 美元兑日元
		{"JPY", "USD", 0.0066},  // 日元兑美元
	}

	for _, r := range rates {
		syncRate(r.From, r.To, r.Rate)
	}

	// 计算美元指数
	fmt.Println("\nCalculating US Dollar Index...")
	calculateUSDIndex()

	fmt.Println("\nSync completed!")
	fmt.Println("\nCurrent rates in database:")
	showRates()
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
		DataSource:    "demo",
		SourceType:    "manual",
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
func calculateUSDIndex() {
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

	// 获取 USD 对这些货币的汇率
	rates := make(map[string]float64)
	var rate models.ExchangeRate
	
	currencies := []string{"EUR", "JPY", "GBP", "CAD", "SEK", "CHF"}
	for _, currency := range currencies {
		if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
			"USD", currency, 1).First(&rate).Error; err == nil {
			rates[currency] = rate.Rate.InexactFloat64()
		}
	}

	// 计算美元指数
	index := 50.14348112
	for currency, weight := range weights {
		r, ok := rates[currency]
		if !ok || r <= 0 {
			continue
		}
		// 使用 math.Pow
		index *= pow(r, -weight)
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
		SyncedAt:      &[]time.Time{