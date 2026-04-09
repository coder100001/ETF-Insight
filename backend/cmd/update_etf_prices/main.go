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
)

const (
	finageBaseURL = "https://api.finage.co.uk"
)

// getFinageAPIKey 获取 Finage API Key，优先从环境变量读取
func getFinageAPIKey() string {
	key := os.Getenv("FINAGE_API_KEY")
	if key == "" {
		fmt.Println("⚠️ FINAGE_API_KEY 环境变量未设置，请先配置：")
		fmt.Println("  export FINAGE_API_KEY=your_api_key_here")
		os.Exit(1)
	}
	return key
}

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
	if err := initDatabase(); err != nil {
		fmt.Printf("❌ 数据库初始化失败: %v\n", err)
		os.Exit(1)
	}

	etfSymbols, err := fetchETFSymbolsFromDB()
	if err != nil {
		fmt.Printf("❌ 获取ETF列表失败: %v\n", err)
		os.Exit(1)
	}

	if len(etfSymbols) == 0 {
		fmt.Println("⚠️ 数据库中没有ETF配置")
		os.Exit(0)
	}

	fmt.Printf("🚀 开始逐个更新 %d 只ETF的完整OHLCV数据...\n\n", len(etfSymbols))

	successCount := 0
	for i, symbol := range etfSymbols {
		fmt.Printf("📦 [%d/%d] 获取 %s 聚合数据...\n", i+1, len(etfSymbols), symbol)

		dataPoint, err := fetchAggregateData(symbol)
		if err != nil {
			fmt.Printf("❌ %s: 获取数据失败 - %v\n", symbol, err)
			continue
		}

		if dataPoint == nil {
			fmt.Printf("⚠️ %s: 无数据返回\n", symbol)
			continue
		}

		if err := saveDataToDB(*dataPoint); err != nil {
			fmt.Printf("❌ %s: 保存数据失败 - %v\n", symbol, err)
			continue
		}

		fmt.Printf("✅ %s: O:%.2f H:%.2f L:%.2f C:%.2f V:%d\n",
			dataPoint.Symbol, dataPoint.Open, dataPoint.High, dataPoint.Low, dataPoint.Close, dataPoint.Volume)
		successCount++

		// 请求间隔，避免触发速率限制
		if i < len(etfSymbols)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Printf("\n✅ 更新完成: %d/%d 成功\n", successCount, len(etfSymbols))
}

func initDatabase() error {
	if err := models.InitDB("etf_insight.db"); err != nil {
		return err
	}
	fmt.Println("✅ 数据库连接成功")
	return nil
}

func fetchETFSymbolsFromDB() ([]string, error) {
	var configs []models.ETFConfig
	result := models.DB.Where("status = ?", 1).Find(&configs)
	if result.Error != nil {
		return nil, result.Error
	}

	symbols := make([]string, 0, len(configs))
	for _, cfg := range configs {
		if cfg.Symbol != "" {
			symbols = append(symbols, cfg.Symbol)
		}
	}

	return symbols, nil
}

// ETFDataPoint 存储单个ETF的聚合数据
type ETFDataPoint struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Timestamp int64
}

// fetchAggregateData 逐个symbol请求聚合API，避免批量推断symbol的问题
func fetchAggregateData(symbol string) (*ETFDataPoint, error) {
	// 获取最近7天的数据，取最后一条作为最新数据
	now := time.Now()
	from := now.AddDate(0, 0, -7).Format("2006-01-02")
	to := now.Format("2006-01-02")

	url := fmt.Sprintf("%s/agg/stock/%s/1/day/%s/%s?apikey=%s",
		finageBaseURL, symbol, from, to, getFinageAPIKey())

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var aggResp FinageAggResponse
	if err := json.Unmarshal(body, &aggResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if aggResp.TotalResults == 0 || len(aggResp.Results) == 0 {
		return nil, nil // 无数据
	}

	// 取最新的数据点（最后一条）
	latest := aggResp.Results[len(aggResp.Results)-1]

	// 验证 symbol 是否匹配
	if aggResp.Symbol != "" && aggResp.Symbol != symbol {
		fmt.Printf("⚠️ %s: API返回symbol不匹配 (期望: %s, 实际: %s)，使用请求symbol\n",
			symbol, symbol, aggResp.Symbol)
	}

	return &ETFDataPoint{
		Symbol:    symbol, // 始终使用请求的symbol，不依赖API返回
		Open:      latest.Open,
		High:      latest.High,
		Low:       latest.Low,
		Close:     latest.Close,
		Volume:    latest.Volume,
		Timestamp: latest.Timestamp,
	}, nil
}

func saveDataToDB(dp ETFDataPoint) error {
	// 将时间戳转换为时间
	date := time.Unix(dp.Timestamp/1000, 0)

	// 验证数据有效性
	if dp.Open <= 0 || dp.Close <= 0 || dp.High <= 0 || dp.Low <= 0 {
		return fmt.Errorf("无效的价格数据: O=%.2f H=%.2f L=%.2f C=%.2f", dp.Open, dp.High, dp.Low, dp.Close)
	}

	// 检查是否已存在该日期的记录
	var existing models.ETFData
	result := models.DB.Where("symbol = ? AND date(date) = date(?)", dp.Symbol, date).First(&existing)

	if result.Error != nil {
		// 创建新记录
		etfData := models.ETFData{
			Symbol:     dp.Symbol,
			Date:       date,
			OpenPrice:  decimal.NewFromFloat(dp.Open),
			ClosePrice: decimal.NewFromFloat(dp.Close),
			HighPrice:  decimal.NewFromFloat(dp.High),
			LowPrice:   decimal.NewFromFloat(dp.Low),
			Volume:     dp.Volume,
			DataSource: "finage",
		}
		return models.DB.Create(&etfData).Error
	}

	// 更新现有记录
	updates := map[string]interface{}{
		"open_price":  decimal.NewFromFloat(dp.Open),
		"high_price":  decimal.NewFromFloat(dp.High),
		"low_price":   decimal.NewFromFloat(dp.Low),
		"close_price": decimal.NewFromFloat(dp.Close),
		"volume":      dp.Volume,
		"data_source": "finage",
	}

	return models.DB.Model(&models.ETFData{}).
		Where("id = ?", existing.ID).
		Updates(updates).Error
}
