package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ETFData struct {
	ID         uint
	Symbol     string
	Date       time.Time
	OpenPrice  decimal.Decimal
	ClosePrice decimal.Decimal
	HighPrice  decimal.Decimal
	LowPrice   decimal.Decimal
	Volume     int64
	DataSource string
	CreatedAt  time.Time
}

// 基准价格映射
var basePrices = map[string]float64{
	"QQQ":   450.0,
	"SCHD":  30.0,
	"VNQ":   90.0,
	"VYM":   110.0,
	"SPYD":  45.0,
	"JEPQ":  55.0,
	"JEPI":  55.0,
	"VTI":   280.0,
	"VOO":   480.0,
	"VEA":   50.0,
	"VWO":   45.0,
	"BND":   75.0,
	"AGG":   100.0,
	"GLD":   220.0,
	"TLT":   90.0,
}

func main() {
	db, err := gorm.Open(sqlite.Open("etf_insight.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	fmt.Printf("Generating 90 days of historical data for %d ETFs...\n", len(basePrices))

	// 为每个ETF生成90天的历史数据
	for symbol, basePrice := range basePrices {
		fmt.Printf("Generating data for %s (base: $%.2f)...\n", symbol, basePrice)
		generateHistoricalData(db, symbol, basePrice)
	}

	fmt.Println("Done!")
}

func generateHistoricalData(db *gorm.DB, symbol string, currentPrice float64) {
	// 删除该ETF的旧数据
	db.Where("symbol = ?", symbol).Delete(&ETFData{})

	// 生成90天的历史数据
	days := 90
	basePrice := currentPrice * 0.95 // 从当前价格的95%开始
	volatility := 0.015              // 1.5%日波动率

	now := time.Now()
	for i := days; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)

		// 跳过周末
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		// 生成随机价格变动
		change := (rand.Float64() - 0.48) * volatility
		basePrice = basePrice * (1 + change)

		// 生成OHLC数据
		open := basePrice * (1 + (rand.Float64()-0.5)*0.005)
		close := basePrice
		high := math.Max(open, close) * (1 + rand.Float64()*0.008)
		low := math.Min(open, close) * (1 - rand.Float64()*0.008)
		volume := int64(1000000 + rand.Int63n(5000000))

		etfData := ETFData{
			Symbol:     symbol,
			Date:       date,
			OpenPrice:  decimal.NewFromFloat(open),
			ClosePrice: decimal.NewFromFloat(close),
			HighPrice:  decimal.NewFromFloat(high),
			LowPrice:   decimal.NewFromFloat(low),
			Volume:     volume,
			DataSource: "generated",
		}

		db.Create(&etfData)
	}

	fmt.Printf("  Generated ~%d records for %s (final price: $%.2f)\n", days*5/7, symbol, basePrice)
}
