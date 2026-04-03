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
		syncRate("CNY", "USD", 1.0/usd