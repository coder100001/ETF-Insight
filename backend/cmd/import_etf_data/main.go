package main

import (
	"fmt"
	"log"
	"time"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

func main() {
	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 需要导入的ETF列表
	etfs := []string{
		"VTI",  // Vanguard Total Stock Market
		"AGG",  // iShares Core U.S. Aggregate Bond
		"IWM",  // iShares Russell 2000
		"EFA",  // iShares MSCI EAFE
		"EEM",  // iShares Emerging Markets
		"TLT",  // iShares 20+ Year Treasury Bond
		"GLD",  // SPDR Gold Shares
		"VNQ",  // Vanguard Real Estate
		"BND",  // Vanguard Total Bond Market
		"VEA",  // Vanguard FTSE Developed Markets
		"VWO",  // Vanguard FTSE Emerging Markets
		"VIG",  // Vanguard Dividend Appreciation
		"VYM",  // Vanguard High Dividend Yield
		"VXUS", // Vanguard Total International Stock
	}

	// 创建Yahoo Finance客户端
	client := services.NewYahooFinanceClient()

	fmt.Println("开始导入ETF历史数据...")
	fmt.Printf("共需导入 %d 个ETF\n\n", len(etfs))

	for _, symbol := range etfs {
		fmt.Printf("正在导入 %s 的数据...\n", symbol)

		// 获取3年历史数据
		prices, err := client.GetHistoricalData(symbol, "3y", "1d")
		if err != nil {
			log.Printf("获取 %s 数据失败: %v\n", symbol, err)
			continue
		}

		if len(prices) == 0 {
			log.Printf("%s 没有返回数据\n", symbol)
			continue
		}

		// 准备批量插入的数据
		var etfDataList []models.ETFData
		for _, price := range prices {
			etfDataList = append(etfDataList, models.ETFData{
				Symbol:     symbol,
				Date:       price.Date,
				OpenPrice:  decimal.NewFromFloat(price.Open),
				ClosePrice: decimal.NewFromFloat(price.Close),
				HighPrice:  decimal.NewFromFloat(price.High),
				LowPrice:   decimal.NewFromFloat(price.Low),
				Volume:     price.Volume,
				DataSource: "yahoo",
			})
		}

		// 使用批量UPSERT操作，提高性能
		count := 0
		if len(etfDataList) > 0 {
			result := models.DB.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "symbol"},
					{Name: "date"},
					{Name: "data_source"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"open_price", "close_price", "high_price", "low_price", "volume",
				}),
			}).CreateInBatches(&etfDataList, 100)

			if result.Error != nil {
				log.Printf("批量保存 %s 数据失败: %v\n", symbol, result.Error)
			} else {
				count = len(etfDataList)
			}
		}

		fmt.Printf("✓ %s 导入完成: %d 条记录\n\n", symbol, count)

		// 添加延迟以避免触发速率限制
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n所有ETF数据导入完成!")

	// 验证导入结果
	fmt.Println("\n验证导入结果:")
	for _, symbol := range etfs {
		var count int64
		models.DB.Model(&models.ETFData{}).Where("symbol = ? AND data_source = ?", symbol, "yahoo").Count(&count)
		if count > 0 {
			fmt.Printf("✓ %s: %d 条记录\n", symbol, count)
		} else {
			fmt.Printf("✗ %s: 无数据\n", symbol)
		}
	}
}
