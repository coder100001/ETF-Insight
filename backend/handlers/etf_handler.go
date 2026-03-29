package handlers

import (
	"net/http"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
)

// ETFHandler ETF 相关处理器
type ETFHandler struct {
	cacheService  *services.CacheService
	analysisService *services.ETFAnalysisService
}

// NewETFHandler 创建 ETF 处理器
func NewETFHandler(cacheService *services.CacheService, analysisService *services.ETFAnalysisService) *ETFHandler {
	return &ETFHandler{
		cacheService:  cacheService,
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

	// 构建返回数据
	var etfList []map[string]interface{}
	for _, cfg := range etfConfigs {
		// 尝试从缓存获取实时数据
		realtimeData, err := h.cacheService.GetRealtimeData(cfg.Symbol)
		if err != nil {
			// 没有缓存数据，使用基本信息
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
		} else {
			// 有缓存数据，使用实时数据
			etfList = append(etfList, map[string]interface{}{
				"symbol":              realtimeData.Symbol,
				"name":                realtimeData.Name,
				"market":              "US",
				"category":            "ETF",
				"current_price":       realtimeData.CurrentPrice,
				"previous_close":      realtimeData.PreviousClose,
				"change":              realtimeData.Change,
				"change_percent":      realtimeData.ChangePercent,
				"open_price":          realtimeData.OpenPrice,
				"day_high":            realtimeData.DayHigh,
				"day_low":             realtimeData.DayLow,
				"volume":              realtimeData.Volume,
				"market_cap":          realtimeData.MarketCap,
				"dividend_yield":      realtimeData.DividendYield,
				"fifty_two_week_high": realtimeData.FiftyTwoWeekHigh,
				"fifty_two_week_low":  realtimeData.FiftyTwoWeekLow,
				"currency":            realtimeData.Currency,
			})
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
	initialInvestment := c.DefaultQuery("initial_investment", "10000")
	taxRate := c.DefaultQuery("tax_rate", "0.10")

	// TODO: 实现收益预测逻辑
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
