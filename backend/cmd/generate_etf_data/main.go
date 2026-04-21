package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// ETF预设配置
type ETFConfig struct {
	Symbol     string
	Name       string
	BasePrice  float64
	Return     float64 // 年化收益率
	Volatility float64 // 年化波动率
}

var etfConfigs = []ETFConfig{
	{Symbol: "VTI", Name: "Vanguard Total Stock Market", BasePrice: 275.0, Return: 0.10, Volatility: 0.16},
	{Symbol: "AGG", Name: "iShares Core U.S. Aggregate Bond", BasePrice: 98.0, Return: 0.04, Volatility: 0.05},
	{Symbol: "IWM", Name: "iShares Russell 2000", BasePrice: 220.0, Return: 0.09, Volatility: 0.19},
	{Symbol: "EFA", Name: "iShares MSCI EAFE", BasePrice: 75.0, Return: 0.08, Volatility: 0.17},
	{Symbol: "EEM", Name: "iShares Emerging Markets", BasePrice: 42.0, Return: 0.08, Volatility: 0.22},
	{Symbol: "TLT", Name: "iShares 20+ Year Treasury Bond", BasePrice: 90.0, Return: 0.03, Volatility: 0.14},
	{Symbol: "GLD", Name: "SPDR Gold Shares", BasePrice: 290.0, Return: 0.06, Volatility: 0.15},
	{Symbol: "VNQ", Name: "Vanguard Real Estate", BasePrice: 85.0, Return: 0.08, Volatility: 0.18},
	{Symbol: "BND", Name: "Vanguard Total Bond Market", BasePrice: 72.0, Return: 0.04, Volatility: 0.05},
	{Symbol: "VEA", Name: "Vanguard FTSE Developed Markets", BasePrice: 47.0, Return: 0.08, Volatility: 0.17},
	{Symbol: "VWO", Name: "Vanguard FTSE Emerging Markets", BasePrice: 43.0, Return: 0.08, Volatility: 0.22},
	{Symbol: "VIG", Name: "Vanguard Dividend Appreciation", BasePrice: 185.0, Return: 0.09, Volatility: 0.15},
	{Symbol: "VYM", Name: "Vanguard High Dividend Yield", BasePrice: 130.0, Return: 0.09, Volatility: 0.15},
	{Symbol: "VXUS", Name: "Vanguard Total International Stock", BasePrice: 58.0, Return: 0.08, Volatility: 0.17},
}

func main() {
	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	fmt.Println("开始生成ETF历史数据...")
	fmt.Printf("共需生成 %d 个ETF的数据\n\n", len(etfConfigs))

	// 生成3年历史数据
	endDate := time.Now()
	startDate := endDate.AddDate(-3, 0, 0)
	days := int(endDate.Sub(startDate).Hours() / 24)

	for _, config := range etfConfigs {
		fmt.Printf("正在生成 %s 的数据...\n", config.Symbol)

		// 检查是否已有数据
		var existingCount int64
		models.DB.Model(&models.ETFData{}).
			Where("symbol = ?", config.Symbol).
			Count(&existingCount)

		if existingCount > 100 {
			fmt.Printf("✓ %s 已有 %d 条记录，跳过\n\n", config.Symbol, existingCount)
			continue
		}

		// 生成模拟数据
		count := generateHistoricalData(config, startDate, days)
		fmt.Printf("✓ %s 生成完成: %d 条记录\n\n", config.Symbol, count)
	}

	fmt.Println("\n所有ETF数据生成完成!")

	// 验证结果
	fmt.Println("\n验证生成结果:")
	for _, config := range etfConfigs {
		var count int64
		models.DB.Model(&models.ETFData{}).Where("symbol = ?", config.Symbol).Count(&count)
		if count > 0 {
			fmt.Printf("✓ %s: %d 条记录\n", config.Symbol, count)
		} else {
			fmt.Printf("✗ %s: 无数据\n", config.Symbol)
		}
	}
}

// generateHistoricalData 生成历史数据
func generateHistoricalData(config ETFConfig, startDate time.Time, days int) int {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 日收益率和波动率
	dailyReturn := config.Return / 365.0
	dailyVol := config.Volatility / math.Sqrt(365.0)

	price := config.BasePrice
	count := 0

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)

		// 跳过周末
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		// 生成随机收益率 (几何布朗运动)
		z := rnd.NormFloat64()
		returnRate := dailyReturn + dailyVol*z

		// 计算OHLC
		openPrice := price
		closePrice := price * (1 + returnRate)

		// 生成日内高低点
		highPrice := math.Max(openPrice, closePrice) * (1 + rnd.Float64()*0.01)
		lowPrice := math.Min(openPrice, closePrice) * (1 - rnd.Float64()*0.01)

		// 生成成交量 (随机)
		volume := int64(1000000 + rnd.Int63n(49000000))

		etfData := models.ETFData{
			Symbol:     config.Symbol,
			Date:       date,
			OpenPrice:  decimal.NewFromFloat(openPrice),
			ClosePrice: decimal.NewFromFloat(closePrice),
			HighPrice:  decimal.NewFromFloat(highPrice),
			LowPrice:   decimal.NewFromFloat(lowPrice),
			Volume:     volume,
			DataSource: "generated",
		}

		// 使用UPSERT避免重复
		models.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"open_price", "close_price", "high_price", "low_price", "volume", "data_source"}),
		}).Create(&etfData)

		price = closePrice
		count++
	}

	return count
}
