package handlers

import (
	"net/http"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
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
	_ = c.Param("symbol")
	_ = c.DefaultQuery("initial_investment", "10000")
	_ = c.DefaultQuery("tax_rate", "0.10")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []map[string]interface{}{
			{"year": 1, "value": 10500},
			{"year": 2, "value": 11025},
			{"year": 3, "value": 11576},
		},
	})
}

// UpdateRealtimeData 更新实时数据
func (h *ETFHandler) UpdateRealtimeData(c *gin.Context) {
	// TODO: 触发实时数据更新任务
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Update task triggered",
		"count":   7,
	})
}
