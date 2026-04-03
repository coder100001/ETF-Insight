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
	fmt.Println("\nCalculating US Dollar Index...