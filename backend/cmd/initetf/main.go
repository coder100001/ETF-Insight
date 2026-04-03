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

	fmt.Println("Initializing ETF data...")

	// 初始化美股ETF实时数据
	initUSETFData()

	// 初始化A股ETF数据
	initAShareETFData()

	fmt.Println("\nETF data initialization completed!")
}

// initUSETFData 初始化美股ETF数据
func initUSETFData() {
	fmt.Println("\nInitializing US ETF data...")

	etfDataList := []struct {
		Symbol     string
		Name       string
		OpenPrice  float64
		ClosePrice float64
		HighPrice  float64
		LowPrice   float64
		Volume     int64
		Category   string
	}{
		{"QQQ", "Invesco QQQ Trust", 436.37, 438.52, 440.15, 435.20, 28500000, "大盘股"},
		{"SCHD", "Schwab US Dividend Equity ETF", 26.73, 26.85, 26.95, 26.68, 4200000, "股息"},
		{"VNQ", "Vanguard Real Estate ETF", 84.77, 84.32, 85.10, 84.15, 3200000, "REITs"},
		{"VYM", "Vanguard High Dividend Yield ETF", 108.07, 108.45, 108.80, 107.90, 1800000, "股息"},
		{"SPYD", "SPDR Portfolio S&P 500 High Dividend ETF", 27.34, 27.42, 27.55, 27.28, 890000, "股息"},
		{"JEPQ", "JPMorgan Nasdaq Equity Premium Income ETF", 48.53, 48.75, 48.95, 48.40, 1200000, "备兑认购"},
		{"JEPI", "JPMorgan Equity Premium Income ETF", 54.10, 54.28, 54.50, 54.00, 2100000, "备兑认购"},
	}

	for _, data := range etfDataList {
		// 检查ETF配置是否存在
		var config models.ETFConfig
		result := models.DB.Where("symbol = ?", data.Symbol).First(&config)

		if result.Error != nil {
			// 创建ETF配置
			config = models.ETFConfig{
				Symbol:          data.Symbol,
				Name:            data.Name,
				Category:        data.Category,
				Currency:        "USD",
				Exchange:        "NASDAQ",
				Status:          1,
				AutoUpdate:      true,
				UpdateFrequency: "每日",
				DataSource:      "Yahoo Finance",
			}
			models.DB.Create(&config)
		}

		// 创建或更新ETF实时数据
		etfData := &models.ETFData{
			Symbol:     data.Symbol,
			Date:       time.Now(),
			OpenPrice:  decimal.NewFromFloat(data.OpenPrice),
			ClosePrice: decimal.NewFromFloat(data.ClosePrice),
			HighPrice:  decimal.NewFromFloat(data.HighPrice),
			LowPrice:   decimal.NewFromFloat(data.LowPrice),
			Volume:     data.Volume,
			DataSource: "Yahoo Finance",
		}

		models.DB.Where("symbol = ? AND date = ?", data.Symbol, etfData.Date.Format("2006-01-02")).
			Assign(etfData).
			FirstOrCreate(etfData)

		changePercent := ((data.ClosePrice - data.OpenPrice) / data.OpenPrice) * 100
		fmt.Printf("  %s: $%.2f (%.2f%%)\n", data.Symbol, data.ClosePrice, changePercent)
	}
}

// initAShareETFData 初始化A股ETF数据
func initAShareETFData() {
	fmt.Println("\nInitializing A-Share ETF data...")

	etfs := []struct {
		Symbol           string
		Name             string
		Exchange         string
		DividendYieldMin float64
		DividendYieldMax float64
		Frequency        models.DividendFrequency
		Benchmark        string
		ManagementFee    float64
		Description      string
	}{
		{"515080", "中证红利ETF", "SSE", 4.8, 5.1, models.FrequencyQuarterly, "中证红利指数", 0.005, "追踪中证红利指数的高股息ETF"},
		{"515180", "红利ETF", "SSE", 4.2, 4.7, models.FrequencyYearly, "上证红利指数", 0.006, "追踪上证红利指数的ETF"},
		{"515300", "红利低波ETF", "SSE", 4.0, 4.8, models.FrequencyQuarterly, "中证红利低波动指数", 0.005, "红利+低波动双因子策略ETF"},
		{"510720", "红利国企ETF", "SSE", 3.5, 4.0, models.FrequencyMonthly, "中证国企红利指数", 0.004, "国企红利主题ETF，月度分红"},
		{"520900", "红利ETF易方达", "SSE", 4.0, 4.5, models.FrequencyQuarterly, "中证红利指数", 0.0015, "低费率红利ETF"},
		{"159545", "红利ETF南方", "SZSE", 4.0, 4.3, models.FrequencyQuarterly, "中证红利指数", 0.0015, "南方基金红利ETF"},
		{"520550", "红利质量ETF", "SSE", 3.5, 4.2, models.FrequencyQuarterly, "中证红利质量指数", 0.005, "红利+质量双因子ETF"},
		{"513820", "港股红利ETF", "SSE", 5.5, 6.8, models.FrequencyQuarterly, "中证港股通高股息指数", 0.006, "投资港股高股息股票的ETF"},
	}

	for _, etf := range etfs {
		// 创建A股ETF数据
		aShareETF := &models.AShareDividendETF{
			Symbol:            etf.Symbol,
			Name:              etf.Name,
			Exchange:          etf.Exchange,
			DividendYieldMin:  decimal.NewFromFloat(etf.DividendYieldMin),
			DividendYieldMax:  decimal.NewFromFloat(etf.DividendYieldMax),
			DividendFrequency: etf.Frequency,
			Benchmark:         etf.Benchmark,
			ManagementFee:     decimal.NewFromFloat(etf.ManagementFee),
			Description:       etf.Description,
			Status:            1,
		}

		models.DB.Where("symbol = ?", etf.Symbol).
			Assign(aShareETF).
			FirstOrCreate(aShareETF)

		fmt.Printf("  %s: %s (%.2f%%-%.2f%%)\n", etf.Symbol, etf.Name, etf.DividendYieldMin, etf.DividendYieldMax)
	}
}
