package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// FinageAggResponse Finage 聚合数据响应
type FinageAggResponse struct {
	Symbol       string `json:"symbol"`
	TotalResults int    `json:"totalResults"`
	Results      []struct {
		Open      float64 `json:"o"`
		High      float64 `json:"h"`
		Low       float64 `json:"l"`
		Close     float64 `json:"c"`
		Volume    int64   `json:"v"`
		Timestamp int64   `json:"t"`
	} `json:"results"`
}

func main() {
	apiKey := os.Getenv("FINAGE_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ FINAGE_API_KEY 环境变量未设置")
		os.Exit(1)
	}

	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		fmt.Printf("❌ 数据库初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 获取所有 ETF 配置
	var etfConfigs []models.ETFConfig
	models.DB.Where("status = ?", 1).Find(&etfConfigs)

	if len(etfConfigs) == 0 {
		fmt.Println("❌ 没有找到启用的 ETF 配置")
		os.Exit(1)
	}

	fmt.Printf("📊 开始同步 %d 个 ETF 的历史数据...\n\n", len(etfConfigs))

	// 获取过去 3 年的数据
	endDate := time.Now()
	startDate := endDate.AddDate(-3, 0, 0)

	for _, cfg := range etfConfigs {
		fmt.Printf("🔄 同步 %s...", cfg.Symbol)
		count := syncETFHistory(cfg.Symbol, apiKey, startDate, endDate)
		fmt.Printf(" ✓ 同步了 %d 条数据\n", count)
		time.Sleep(1 * time.Second) // 避免触发速率限制
	}

	fmt.Println("\n✅ 历史数据同步完成!")
}

func syncETFHistory(symbol, apiKey string, startDate, endDate time.Time) int {
	from := startDate.Format("2006-01-02")
	to := endDate.Format("2006-01-02")

	url := fmt.Sprintf("https://api.finage.co.uk/agg/stock/%s/1/day/%s/%s?apikey=%s",
		symbol, from, to, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf(" (错误: %v)", err)
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(" (读取错误: %v)", err)
		return 0
	}

	var aggResp FinageAggResponse
	if err := json.Unmarshal(body, &aggResp); err != nil {
		fmt.Printf(" (解析错误: %v)", err)
		return 0
	}

	if aggResp.TotalResults == 0 {
		return 0
	}

	// 保存到数据库
	count := 0
	for _, r := range aggResp.Results {
		date := time.Unix(r.Timestamp/1000, 0)

		etfData := models.ETFData{
			Symbol:     symbol,
			Date:       date,
			OpenPrice:  decimal.NewFromFloat(r.Open),
			ClosePrice: decimal.NewFromFloat(r.Close),
			HighPrice:  decimal.NewFromFloat(r.High),
			LowPrice:   decimal.NewFromFloat(r.Low),
			Volume:     r.Volume,
			DataSource: "finage",
		}

		models.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"open_price", "close_price", "high_price", "low_price", "volume", "data_source"}),
		}).Create(&etfData)

		count++
	}

	return count
}
