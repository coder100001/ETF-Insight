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
	