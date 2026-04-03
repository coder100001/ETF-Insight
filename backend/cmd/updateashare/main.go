package main

import (
	"fmt"
	"log"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

func main() {
	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("开始更新A股红利ETF数据...")

	// 更新ETF基础数据（根据用户提供的真实数据）
	updateETFData()

	// 更新投资组合配置
	updatePortfolioConfig()

	fmt.Println("A股红利ETF数据更新完成!")
}

func updateETFData() {
	// 根据用户提供的数据和搜索结果验证后的准确数据
	etfData := []struct {
		Symbol            string
		Name              string
		DividendYieldMin  float64
		DividendYieldMax  float64
		DividendFrequency models.DividendFrequency
		Benchmark         string
		Exchange          string
		ManagementFee     float64
		Description       string
	}{
		{
			Symbol:            "515080",
			Name:              "中证红利ETF",
			DividendYieldMin:  4.8,
			DividendYieldMax:  5.1,
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利指数",
			Exchange:          "SSE",
			ManagementFee:     0.0020,
			Description:       "招商中证红利ETF，跟踪中证红利指数，2025年股息率约5.12%，季度分红",
		},
		{
			Symbol:            "515180",
			Name:              "红利ETF",
			DividendYieldMin:  4.4,
			DividendYieldMax:  4.5,
			DividendFrequency: models.FrequencyYearly,
			Benchmark:         "中证红利指数",
			Exchange:          "SSE",
			ManagementFee:     0.0015,
			Description:       "易方达中证红利ETF，2025年分红收益率约4.46%，年度分红",
		},
		{
			Symbol:            "515300",
			Name:              "红利低波ETF",
			DividendYieldMin:  3.5,
			DividendYieldMax:  3.7,
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "沪深300红利低波动指数",
			Exchange:          "SSE",
			ManagementFee:     0.0050,
			Description:       "嘉实沪深300红利低波动ETF，2025年12月股息率约5.2%，季度分红",
		},
		{
			Symbol:            "510720",
			Name:              "红利国企ETF",
			DividendYieldMin:  3.5,
			DividendYieldMax:  4.0,
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "上证国有企业红利指数",
			Exchange:          "SSE",
			ManagementFee:     0.0050,
			Description:       "国泰上证国有企业红利ETF，近12个月股息率约4.2%，月度分红",
		},
		{
			Symbol:            "520900",
			Name:              "港股红利ETF",
			DividendYieldMin:  5.5,
			DividendYieldMax:  6.0,
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证国新港股通央企红利指数",
			Exchange:          "SSE",
			ManagementFee:     0.0050,
			Description:       "广发中证国新港股通央企红利ETF，2025年股息率中枢5%-9%，季度分红",
		},
		{
			Symbol:            "159545",
			Name:              "港股低波ETF",
			DividendYieldMin:  3.5,
			DividendYieldMax:  3.6,
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "恒生港股通高股息低波动指数",
			Exchange:          "SHZ",
			ManagementFee:     0.0015,
			Description:       "易方达恒生港股通高股息低波动ETF，最新股息率约6.16%，月度分红",
		},
		{
			Symbol:            "520550",
			Name:              "恒生红利ETF",
			DividendYieldMin:  3.5,
			DividendYieldMax:  3.6,
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "恒生高股息率指数",
			Exchange:          "SHZ",
			ManagementFee:     0.0050,
			Description:       "跟踪恒生高股息率指数，月度分红",
		},
		{
			Symbol:            "513820",
			Name:              "港股通红利ETF",
			DividendYieldMin:  3.9,
			DividendYieldMax:  4.1,
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "港股通高股息指数",
			Exchange:          "SSE",
			ManagementFee:     0.0060,
			Description:       "港股红利ETF基金，最新股息率约8.12%，月度分红，同类领先",
		},
	}

	for _, data := range etfData {
		var etf models.AShareDividendETF
		result := models.DB.Where("symbol = ?", data.Symbol).First(&etf)

		if result.Error != nil {
			// 创建新记录
			etf = models.AShareDividendETF{
				Symbol:            data.Symbol,
				Name:              data.Name,
				DividendYieldMin:  decimal.NewFromFloat(data.DividendYieldMin),
				DividendYieldMax:  decimal.NewFromFloat(data.DividendYieldMax),
				DividendFrequency: data.DividendFrequency,
				Benchmark:         data.Benchmark,
				Exchange:          data.Exchange,
				ManagementFee:     decimal.NewFromFloat(data.ManagementFee),
				Description:       data.Description,
				Status:            1,
			}
			if err := models.DB.Create(&etf).Error; err != nil {
				log.Printf("创建ETF %s 失败: %v", data.Symbol, err)
			} else {
				fmt.Printf("✓ 创建ETF: %s (%s)\n", data.Symbol, data.Name)
			}
		} else {
			// 更新现有记录
			etf.Name = data.Name
			etf.DividendYieldMin = decimal.NewFromFloat(data.DividendYieldMin)
			etf.DividendYieldMax = decimal.NewFromFloat(data.DividendYieldMax)
			etf.DividendFrequency = data.DividendFrequency
			etf.Benchmark = data.Benchmark
			etf.Exchange = data.Exchange
			etf.ManagementFee = decimal.NewFromFloat(data.ManagementFee)
			etf.Description = data.Description
			if err := models.DB.Save(&etf).Error; err != nil {
				log.Printf("更新ETF %s 失败: %v", data.Symbol, err)
			} else {
				fmt.Printf("✓ 更新ETF: %s (%s)\n", data.Symbol, data.Name)
			}
		}
	}
}

func updatePortfolioConfig() {
	// 用户提供的投资组合配置（单位：万元）
	// 515080:15w, 515180:5w, 515300:5w, 510720:5w, 520900:5w, 159545:5w, 520550:5w, 513820:5w
	// 总计: 50w
	portfolioConfig := map[string]float64{
		"515080": 15.0, // 中证红利ETF - 季分
		"515180": 5.0,  // 红利ETF - 年分
		"515300": 5.0,  // 红利低波ETF - 季分
		"510720": 5.0,  // 红利国企ETF - 月分
		"520900": 5.0,  // 港股红利ETF - 季分
		"159545": 5.0,  // 港股低波ETF - 月分
		"520550": 5.0,  // 恒生红利ETF - 月分
		"513820": 5.0,  // 港股通红利ETF - 月分
	}

	// 查找或创建默认投资组合
	var portfolio models.AShareETFPortfolio
	result := models.DB.Where("is_default = ?", true).First(&portfolio)

	if result.Error != nil {
		// 创建新组合
		portfolio = models.AShareETFPortfolio{
			Name:            "A股红利ETF组合",
			Description:     "精选8只A股及港股红利ETF，总投入50万",
			TotalInvestment: decimal.NewFromFloat(500000),
			IsDefault:       true,
		}
		if err := models.DB.Create(&portfolio).Error; err != nil {
			log.Printf("创建投资组合失败: %v", err)
			return
		}
		fmt.Printf("✓ 创建投资组合: %s (ID: %d)\n", portfolio.Name, portfolio.ID)
	} else {
		// 更新现有组合
		portfolio.Name = "A股红利ETF组合"
		portfolio.Description = "精选8只A股及港股红利ETF，总投入50万"
		portfolio.TotalInvestment = decimal.NewFromFloat(500000)
		if err := models.DB.Save(&portfolio).Error; err != nil {
			log.Printf("更新投资组合失败: %v", err)
			return
		}
		fmt.Printf("✓ 更新投资组合: %s (ID: %d)\n", portfolio.Name, portfolio.ID)

		// 删除旧持仓
		models.DB.Where("portfolio_id = ?", portfolio.ID).Delete(&models.ASharePortfolioHolding{})
	}

	// 创建新的持仓记录
	for symbol, amount := range portfolioConfig {
		var etf models.AShareDividendETF
		if err := models.DB.Where("symbol = ?", symbol).First(&etf).Error; err != nil {
			log.Printf("找不到ETF %s: %v", symbol, err)
			continue
		}

		holding := models.ASharePortfolioHolding{
			PortfolioID: portfolio.ID,
			ETFID:       etf.ID,
			Investment:  decimal.NewFromFloat(amount * 10000), // 转换为元
		}
		if err := models.DB.Create(&holding).Error; err != nil {
			log.Printf("创建持仓 %s 失败: %v", symbol, err)
		} else {
			fmt.Printf("  ✓ 持仓: %s - %.1f万元\n", symbol, amount)
		}
	}

	// 计算并显示组合统计
	fmt.Println("\n组合配置统计:")
	fmt.Printf("  总投资: 50万元\n")
	fmt.Printf("  持仓数量: 8只ETF\n")

	// 计算预期分红
	var totalDividend decimal.Decimal
	for symbol, amount := range portfolioConfig {
		var etf models.AShareDividendETF
		models.DB.Where("symbol = ?", symbol).First(&etf)
		investment := decimal.NewFromFloat(amount * 10000)
		dividendYield := etf.DividendYieldMin.Add(etf.DividendYieldMax).Div(decimal.NewFromInt(2))
		dividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))
		totalDividend = totalDividend.Add(dividend)
	}

	avgYield := totalDividend.Div(decimal.NewFromFloat(500000)).Mul(decimal.NewFromInt(100))
	fmt.Printf("  预期年分红: %s元\n", totalDividend.StringFixed(2))
	fmt.Printf("  平均股息率: %s%%\n", avgYield.StringFixed(2))
}
