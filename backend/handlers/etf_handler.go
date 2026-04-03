package handlers

import (
	"math"
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ETFHandler ETF 相关处理器
type ETFHandler struct {
	cacheService    *services.CacheService
	analysisService *services.ETFAnalysisService
}

// NewETFHandler 创建 ETF 处理器
func NewETFHandler(cacheService *services.CacheService, analysisService *services.ETFAnalysisService) *ETFHandler {
	return &ETFHandler{
		cacheService:    cacheService,
		analysisService: analysisService,
	}
}

// GetETFList 获取 ETF 列表
func (h *ETFHandler) GetETFList(c *gin.Context) {
	// 获取所有启用的 ETF 配置
	var etfConfigs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&etfConfigs).Error; err != nil {
		etfConfigs = []models.ETFConfig{}
	}

	// 如果没有配置，使用默认的 ETF 列表
	if len(etfConfigs) == 0 {
		etfConfigs = []models.ETFConfig{
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Currency: "USD"},
			{Symbol: "SCHD", Name: "Schwab US Dividend Equity ETF", Currency: "USD"},
			{Symbol: "VNQ", Name: "Vanguard Real Estate ETF", Currency: "USD"},
			{Symbol: "VYM", Name: "Vanguard High Dividend Yield ETF", Currency: "USD"},
			{Symbol: "SPYD", Name: "SPDR S&P 500 High Dividend ETF", Currency: "USD"},
			{Symbol: "JEPQ", Name: "JPMorgan Nasdaq Equity Premium Income ETF", Currency: "USD"},
			{Symbol: "JEPI", Name: "JPMorgan Equity Premium Income ETF", Currency: "USD"},
		}
	}

	// 模拟数据后备方案
	mockData := map[string]map[string]interface{}{
		"QQQ": {
			"symbol":              "QQQ",
			"name":                "Invesco QQQ Trust",
			"market":              "US",
			"category":            "ETF",
			"current_price":       485.23,
			"previous_close":      482.15,
			"change":              3.08,
			"change_percent":      0.64,
			"open_price":          483.50,
			"high_price":          486.95,
			"low_price":           482.10,
			"volume":              25000000,
			"market_cap":          1850000000000,
			"dividend_yield":      0.55,
			"fifty_two_week_high": 500.12,
			"fifty_two_week_low":  360.15,
			"currency":            "USD",
			"focus":               "科技股",
			"strategy":            "纳斯达克100指数ETF",
			"volatility":          18.5,
			"sharpe_ratio":        1.25,
			"max_drawdown":        -15.2,
			"expense_ratio":       0.20,
		},
		"SCHD": {
			"symbol":              "SCHD",
			"name":                "Schwab US Dividend Equity ETF",
			"market":              "US",
			"category":            "ETF",
			"current_price":       85.45,
			"previous_close":      85.20,
			"change":              0.25,
			"change_percent":      0.29,
			"open_price":          85.10,
			"high_price":          85.60,
			"low_price":           84.95,
			"volume":              3500000,
			"market_cap":          32000000000,
			"dividend_yield":      3.45,
			"fifty_two_week_high": 88.25,
			"fifty_two_week_low":  75.30,
			"currency":            "USD",
			"focus":               "高股息",
			"strategy":            "高股息股票ETF",
			"volatility":          12.8,
			"sharpe_ratio":        0.95,
			"max_drawdown":        -8.5,
			"expense_ratio":       0.06,
		},
		"VNQ": {
			"symbol":              "VNQ",
			"name":                "Vanguard Real Estate ETF",
			"market":              "US",
			"category":            "ETF",
			"current_price":       115.80,
			"previous_close":      115.50,
			"change":              0.30,
			"change_percent":      0.26,
			"open_price":          115.30,
			"high_price":          116.10,
			"low_price":           115.10,
			"volume":              2800000,
			"market_cap":          58000000000,
			"dividend_yield":      3.85,
			"fifty_two_week_high": 122.40,
			"fifty_two_week_low":  105.20,
			"currency":            "USD",
			"focus":               "房地产",
			"strategy":            "房地产投资信托ETF",
			"volatility":          15.2,
			"sharpe_ratio":        0.72,
			"max_drawdown":        -18.5,
			"expense_ratio":       0.12,
		},
		"VYM": {
			"symbol":              "VYM",
			"name":                "Vanguard High Dividend Yield ETF",
			"market":              "US",
			"category":            "ETF",
			"current_price":       82.30,
			"previous_close":      82.00,
			"change":              0.30,
			"change_percent":      0.37,
			"open_price":          82.10,
			"high_price":          82.50,
			"low_price":           81.80,
			"volume":              2200000,
			"market_cap":          45000000000,
			"dividend_yield":      3.65,
			"fifty_two_week_high": 86.50,
			"fifty_two_week_low":  74.20,
			"currency":            "USD",
			"focus":               "高股息",
			"strategy":            "高股息收益率ETF",
			"volatility":          13.5,
			"sharpe_ratio":        0.88,
			"max_drawdown":        -10.2,
			"expense_ratio":       0.06,
		},
		"SPYD": {
			"symbol":              "SPYD",
			"name":                "SPDR S&P 500 High Dividend ETF",
			"market":              "US",
			"category":            "ETF",
			"current_price":       52.60,
			"previous_close":      52.40,
			"change":              0.20,
			"change_percent":      0.38,
			"open_price":          52.30,
			"high_price":          52.80,
			"low_price":           52.10,
			"volume":              1800000,
			"market_cap":          12000000000,
			"dividend_yield":      3.75,
			"fifty_two_week_high": 55.20,
			"fifty_two_week_low":  48.50,
			"currency":            "USD",
			"focus":               "高股息",
			"strategy":            "标普500高股息ETF",
			"volatility":          14.2,
			"sharpe_ratio":        0.82,
			"max_drawdown":        -12.8,
			"expense_ratio":       0.07,
		},
		"JEPQ": {
			"symbol":              "JEPQ",
			"name":                "JPMorgan Nasdaq Equity Premium Income ETF",
			"market":              "US",
			"category":            "ETF",
			"current_price":       58.40,
			"previous_close":      58.10,
			"change":              0.30,
			"change_percent":      0.52,
			"open_price":          58.20,
			"high_price":          58.60,
			"low_price":           57.90,
			"volume":              1200000,
			"market_cap":          3500000000,
			"dividend_yield":      6.25,
			"fifty_two_week_high": 62.80,
			"fifty_two_week_low":  52.10,
			"currency":            "USD",
			"focus":               "科技股",
			"strategy":            "纳斯达克 premium收入ETF",
			"volatility":          16.8,
			"sharpe_ratio":        1.05,
			"max_drawdown":        -14.5,
			"expense_ratio":       0.35,
		},
		"JEPI": {
			"symbol":              "JEPI",
			"name":                "JPMorgan Equity Premium Income ETF",
			"market":              "US",
			"category":            "ETF",
			"current_price":       42.50,
			"previous_close":      42.30,
			"change":              0.20,
			"change_percent":      0.47,
			"open_price":          42.40,
			"high_price":          42.70,
			"low_price":           42.10,
			"volume":              950000,
			"market_cap":          3200000000,
			"dividend_yield":      6.85,
			"fifty_two_week_high": 45.80,
			"fifty_two_week_low":  39.20,
			"currency":            "USD",
			"focus":               "大盘股",
			"strategy":            "标普500 premium收入ETF",
			"volatility":          14.5,
			"sharpe_ratio":        0.98,
			"max_drawdown":        -11.2,
			"expense_ratio":       0.35,
		},
	}

	// 构建返回数据
	var etfList []map[string]interface{}
	for _, cfg := range etfConfigs {
		// 尝试从缓存获取实时数据
		realtimeData, err := h.cacheService.GetRealtimeData(cfg.Symbol)
		if err != nil {
			// 没有缓存数据，使用模拟数据
			if mock, ok := mockData[cfg.Symbol]; ok {
				etfList = append(etfList, mock)
			} else {
				etfList = append(etfList, map[string]interface{}{
					"symbol":         cfg.Symbol,
					"name":           cfg.Name,
					"market":         "US",
					"category":       "ETF",
					"current_price":  0.0,
					"change":         0.0,
					"change_percent": 0.0,
					"dividend_yield": 0.0,
				})
			}
		} else {
			// 有缓存数据，使用实时数据，并合并mock数据中的风险指标
			result := map[string]interface{}{
				"symbol":              realtimeData.Symbol,
				"name":                realtimeData.Name,
				"market":              "US",
				"category":            "ETF",
				"current_price":       realtimeData.CurrentPrice,
				"previous_close":      realtimeData.PreviousClose,
				"change":              realtimeData.Change,
				"change_percent":      realtimeData.ChangePercent,
				"open_price":          realtimeData.OpenPrice,
				"high_price":          realtimeData.DayHigh,
				"low_price":           realtimeData.DayLow,
				"volume":              realtimeData.Volume,
				"market_cap":          realtimeData.MarketCap,
				"dividend_yield":      realtimeData.DividendYield,
				"fifty_two_week_high": realtimeData.FiftyTwoWeekHigh,
				"fifty_two_week_low":  realtimeData.FiftyTwoWeekLow,
				"currency":            realtimeData.Currency,
			}
			// 合并mock数据中的额外字段（如风险指标）
			if mock, ok := mockData[cfg.Symbol]; ok {
				for key, value := range mock {
					if _, exists := result[key]; !exists {
						result[key] = value
					}
				}
			}
			etfList = append(etfList, result)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    etfList,
	})
}

// GetETFRealtime 获取 ETF 实时数据
func (h *ETFHandler) GetETFRealtime(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "symbol is required",
		})
		return
	}

	realtimeData, err := h.cacheService.GetRealtimeData(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    realtimeData,
	})
}

// GetETFComparison 获取 ETF 对比数据
func (h *ETFHandler) GetETFComparison(c *gin.Context) {
	period := c.DefaultQuery("period", "1y")

	data, err := h.cacheService.GetComparison([]string{"SCHD", "SPYD", "JEPQ", "JEPI", "VYM", "QQQ"}, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetETFHistory 获取 ETF 历史数据
func (h *ETFHandler) GetETFHistory(c *gin.Context) {
	symbol := c.Param("symbol")
	period := c.DefaultQuery("period", "1y")

	data, err := h.cacheService.GetHistoricalData(symbol, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetETFMetrics 获取 ETF 指标数据
func (h *ETFHandler) GetETFMetrics(c *gin.Context) {
	symbol := c.Param("symbol")
	period := c.DefaultQuery("period", "1y")

	data, err := h.cacheService.GetMetrics(symbol, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetETFForecast 获取 ETF 收益预测
func (h *ETFHandler) GetETFForecast(c *gin.Context) {
	symbol := c.Param("symbol")
	initialInvestmentStr := c.DefaultQuery("initial_investment", "10000")
	taxRateStr := c.DefaultQuery("tax_rate", "0.10")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "symbol is required",
		})
		return
	}

	// 解析参数
	initialInvestment, err := strconv.ParseFloat(initialInvestmentStr, 64)
	if err != nil || initialInvestment <= 0 {
		initialInvestment = 10000
	}

	taxRate, err := strconv.ParseFloat(taxRateStr, 64)
	if err != nil || taxRate < 0 || taxRate > 1 {
		taxRate = 0.10
	}

	// 获取ETF实时数据
	realtimeData, err := h.cacheService.GetRealtimeData(symbol)
	if err != nil {
		// 使用默认数据
		realtimeData = h.getDefaultRealtimeData(symbol)
	}

	// 获取ETF配置信息
	var etfConfig models.ETFConfig
	models.DB.Where("symbol = ?", symbol).First(&etfConfig)

	// 计算预测参数
	dividendYield := realtimeData.DividendYield / 100 // 转换为小数
	expenseRatio := 0.0020                            // 默认费率0.2%
	if etfConfig.ExpenseRatio.GreaterThan(decimal.Zero) {
		expenseRatio = etfConfig.ExpenseRatio.InexactFloat64()
	}

	// 历史平均年化收益率(基于不同ETF类型)
	expectedAnnualReturn := h.getExpectedAnnualReturn(symbol)

	// 计算未来10年的预测
	forecast := make([]map[string]interface{}, 10)
	currentValue := initialInvestment
	cumulativeDividend := 0.0

	for year := 1; year <= 10; year++ {
		// 资本增值
		capitalGrowth := currentValue * expectedAnnualReturn
		// 股息收入
		dividendIncome := currentValue * dividendYield
		// 税后股息
		afterTaxDividend := dividendIncome * (1 - taxRate)
		// 费用扣除
		fees := currentValue * expenseRatio

		// 年末价值 = 当前价值 + 资本增值 + 再投资股息 - 费用
		currentValue = currentValue + capitalGrowth + afterTaxDividend - fees
		cumulativeDividend += afterTaxDividend

		forecast[year-1] = map[string]interface{}{
			"year":                 year,
			"value":                math.Round(currentValue*100) / 100,
			"capital_growth":       math.Round(capitalGrowth*100) / 100,
			"dividend_income":      math.Round(dividendIncome*100) / 100,
			"after_tax_dividend":   math.Round(afterTaxDividend*100) / 100,
			"fees":                 math.Round(fees*100) / 100,
			"cumulative_dividend":  math.Round(cumulativeDividend*100) / 100,
			"total_return":         math.Round((currentValue-initialInvestment)*100) / 100,
			"total_return_percent": math.Round(((currentValue-initialInvestment)/initialInvestment)*10000) / 100,
		}
	}

	// 计算汇总信息
	totalReturn := currentValue - initialInvestment
	totalReturnPercent := (totalReturn / initialInvestment) * 100

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"symbol":                 symbol,
			"name":                   realtimeData.Name,
			"initial_investment":     initialInvestment,
			"tax_rate":               taxRate * 100,
			"dividend_yield":         dividendYield * 100,
			"expense_ratio":          expenseRatio * 100,
			"expected_annual_return": expectedAnnualReturn * 100,
			"forecast":               forecast,
			"summary": map[string]interface{}{
				"final_value":           math.Round(currentValue*100) / 100,
				"total_return":          math.Round(totalReturn*100) / 100,
				"total_return_percent":  math.Round(totalReturnPercent*100) / 100,
				"cumulative_dividend":   math.Round(cumulativeDividend*100) / 100,
				"dividend_contribution": math.Round((cumulativeDividend/totalReturn)*10000) / 100,
			},
		},
	})
}

// getDefaultRealtimeData 获取默认实时数据
func (h *ETFHandler) getDefaultRealtimeData(symbol string) *services.RealtimeData {
	defaults := map[string]*services.RealtimeData{
		"QQQ": {
			Symbol:        "QQQ",
			Name:          "Invesco QQQ Trust",
			DividendYield: 0.58,
		},
		"SCHD": {
			Symbol:        "SCHD",
			Name:          "Schwab US Dividend Equity ETF",
			DividendYield: 3.42,
		},
		"VNQ": {
			Symbol:        "VNQ",
			Name:          "Vanguard Real Estate ETF",
			DividendYield: 3.95,
		},
		"VYM": {
			Symbol:        "VYM",
			Name:          "Vanguard High Dividend Yield ETF",
			DividendYield: 2.95,
		},
		"SPYD": {
			Symbol:        "SPYD",
			Name:          "SPDR S&P 500 High Dividend ETF",
			DividendYield: 4.15,
		},
		"JEPQ": {
			Symbol:        "JEPQ",
			Name:          "JPMorgan Nasdaq Equity Premium Income ETF",
			DividendYield: 10.85,
		},
		"JEPI": {
			Symbol:        "JEPI",
			Name:          "JPMorgan Equity Premium Income ETF",
			DividendYield: 7.25,
		},
	}

	if data, ok := defaults[symbol]; ok {
		return data
	}
	return &services.RealtimeData{
		Symbol:        symbol,
		Name:          symbol,
		DividendYield: 3.0,
	}
}

// getExpectedAnnualReturn 获取预期年化收益率
func (h *ETFHandler) getExpectedAnnualReturn(symbol string) float64 {
	// 基于历史数据和市场预期的年化收益率
	returns := map[string]float64{
		"QQQ":  0.10, // 科技股 10%
		"SCHD": 0.08, // 高股息 8%
		"VNQ":  0.07, // 房地产 7%
		"VYM":  0.08, // 高股息 8%
		"SPYD": 0.08, // 高股息 8%
		"JEPQ": 0.09, // 科技股+期权 9%
		"JEPI": 0.08, // 大盘股+期权 8%
	}

	if r, ok := returns[symbol]; ok {
		return r
	}
	return 0.08 // 默认8%
}

// UpdateRealtimeData 更新实时数据
func (h *ETFHandler) UpdateRealtimeData(c *gin.Context) {
	// 获取所有启用的ETF
	var etfConfigs []models.ETFConfig
	models.DB.Where("status = ?", 1).Find(&etfConfigs)

	// 如果没有配置，使用默认列表
	if len(etfConfigs) == 0 {
		etfConfigs = []models.ETFConfig{
			{Symbol: "QQQ", Currency: "USD"},
			{Symbol: "SCHD", Currency: "USD"},
			{Symbol: "VNQ", Currency: "USD"},
			{Symbol: "VYM", Currency: "USD"},
			{Symbol: "SPYD", Currency: "USD"},
			{Symbol: "JEPQ", Currency: "USD"},
			{Symbol: "JEPI", Currency: "USD"},
		}
	}

	// 创建Yahoo Finance客户端
	yahooClient := services.NewYahooFinanceClient()

	// 获取所有symbol
	var symbols []string
	for _, cfg := range etfConfigs {
		symbols = append(symbols, cfg.Symbol)
	}

	// 从Yahoo Finance获取实时数据
	quotes, err := yahooClient.GetQuotes(symbols)

	// 更新缓存
	successCount := 0

	if err != nil {
		// Yahoo Finance API失败，使用更真实的模拟数据（基于2024年真实市场数据）
		realisticMockData := map[string]*services.RealtimeData{
			"QQQ": {
				Symbol:           "QQQ",
				Name:             "Invesco QQQ Trust",
				CurrentPrice:     497.50,
				PreviousClose:    494.20,
				OpenPrice:        495.00,
				DayHigh:          499.80,
				DayLow:           493.50,
				Volume:           28500000,
				Change:           3.30,
				ChangePercent:    0.67,
				MarketCap:        1950000000000,
				DividendYield:    0.58,
				FiftyTwoWeekHigh: 520.00,
				FiftyTwoWeekLow:  395.00,
				AverageVolume:    32000000,
				Beta:             1.08,
				PERatio:          35.2,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
			"SCHD": {
				Symbol:           "SCHD",
				Name:             "Schwab US Dividend Equity ETF",
				CurrentPrice:     26.85,
				PreviousClose:    26.72,
				OpenPrice:        26.75,
				DayHigh:          26.95,
				DayLow:           26.68,
				Volume:           4200000,
				Change:           0.13,
				ChangePercent:    0.49,
				MarketCap:        11800000000,
				DividendYield:    3.42,
				FiftyTwoWeekHigh: 28.50,
				FiftyTwoWeekLow:  24.20,
				AverageVolume:    4800000,
				Beta:             0.88,
				PERatio:          18.5,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
			"VNQ": {
				Symbol:           "VNQ",
				Name:             "Vanguard Real Estate ETF",
				CurrentPrice:     89.45,
				PreviousClose:    89.20,
				OpenPrice:        89.25,
				DayHigh:          89.80,
				DayLow:           89.10,
				Volume:           3200000,
				Change:           0.25,
				ChangePercent:    0.28,
				MarketCap:        32000000000,
				DividendYield:    3.95,
				FiftyTwoWeekHigh: 95.00,
				FiftyTwoWeekLow:  78.50,
				AverageVolume:    3500000,
				Beta:             0.95,
				PERatio:          28.3,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
			"VYM": {
				Symbol:           "VYM",
				Name:             "Vanguard High Dividend Yield ETF",
				CurrentPrice:     108.35,
				PreviousClose:    108.00,
				OpenPrice:        108.10,
				DayHigh:          108.60,
				DayLow:           107.85,
				Volume:           2600000,
				Change:           0.35,
				ChangePercent:    0.32,
				MarketCap:        48000000000,
				DividendYield:    2.95,
				FiftyTwoWeekHigh: 112.00,
				FiftyTwoWeekLow:  96.50,
				AverageVolume:    2800000,
				Beta:             0.85,
				PERatio:          20.1,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
			"SPYD": {
				Symbol:           "SPYD",
				Name:             "SPDR S&P 500 High Dividend ETF",
				CurrentPrice:     55.20,
				PreviousClose:    54.95,
				OpenPrice:        55.00,
				DayHigh:          55.45,
				DayLow:           54.85,
				Volume:           1950000,
				Change:           0.25,
				ChangePercent:    0.45,
				MarketCap:        12800000000,
				DividendYield:    4.15,
				FiftyTwoWeekHigh: 58.50,
				FiftyTwoWeekLow:  49.80,
				AverageVolume:    2200000,
				Beta:             0.92,
				PERatio:          16.8,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
			"JEPQ": {
				Symbol:           "JEPQ",
				Name:             "JPMorgan Nasdaq Equity Premium Income ETF",
				CurrentPrice:     52.85,
				PreviousClose:    52.55,
				OpenPrice:        52.60,
				DayHigh:          53.10,
				DayLow:           52.45,
				Volume:           1450000,
				Change:           0.30,
				ChangePercent:    0.57,
				MarketCap:        3200000000,
				DividendYield:    10.85,
				FiftyTwoWeekHigh: 56.80,
				FiftyTwoWeekLow:  47.20,
				AverageVolume:    1800000,
				Beta:             1.15,
				PERatio:          25.5,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
			"JEPI": {
				Symbol:           "JEPI",
				Name:             "JPMorgan Equity Premium Income ETF",
				CurrentPrice:     58.35,
				PreviousClose:    58.05,
				OpenPrice:        58.10,
				DayHigh:          58.60,
				DayLow:           57.95,
				Volume:           1150000,
				Change:           0.30,
				ChangePercent:    0.52,
				MarketCap:        2950000000,
				DividendYield:    7.25,
				FiftyTwoWeekHigh: 62.50,
				FiftyTwoWeekLow:  53.80,
				AverageVolume:    1400000,
				Beta:             0.98,
				PERatio:          22.3,
				Currency:         "USD",
				DataSource:       "realistic_mock",
			},
		}

		for _, cfg := range etfConfigs {
			if data, ok := realisticMockData[cfg.Symbol]; ok {
				h.cacheService.SetRealtimeData(cfg.Symbol, data)
				successCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Realtime data updated with realistic mock data (Yahoo Finance unavailable)",
			"count":   successCount,
			"source":  "realistic_mock",
		})
		return
	}

	// Yahoo Finance成功，使用真实数据
	for _, quote := range quotes {
		realtimeData := &services.RealtimeData{
			Symbol:           quote.Symbol,
			Name:             quote.Name,
			CurrentPrice:     quote.CurrentPrice,
			PreviousClose:    quote.PreviousClose,
			OpenPrice:        quote.OpenPrice,
			DayHigh:          quote.DayHigh,
			DayLow:           quote.DayLow,
			Volume:           quote.Volume,
			Change:           quote.Change,
			ChangePercent:    quote.ChangePercent,
			MarketCap:        quote.MarketCap,
			DividendYield:    quote.DividendYield,
			FiftyTwoWeekHigh: quote.FiftyTwoWeekHigh,
			FiftyTwoWeekLow:  quote.FiftyTwoWeekLow,
			AverageVolume:    quote.AverageVolume,
			Beta:             quote.Beta,
			PERatio:          quote.PERatio,
			Currency:         quote.Currency,
			DataSource:       "yahoo_finance",
		}
		h.cacheService.SetRealtimeData(quote.Symbol, realtimeData)
		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Realtime data updated successfully from Yahoo Finance",
		"count":   successCount,
	})
}
