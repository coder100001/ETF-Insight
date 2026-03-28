package handlers

import (
	"net/http"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// ETFHandler ETF处理器
type ETFHandler struct {
	cacheService *services.CacheService
	analysisSvc  *services.ETFAnalysisService
}

// NewETFHandler 创建新的ETF处理器
func NewETFHandler(cache *services.CacheService, analysis *services.ETFAnalysisService) *ETFHandler {
	return &ETFHandler{
		cacheService: cache,
		analysisSvc:  analysis,
	}
}

// GetETFList 获取ETF列表
func (h *ETFHandler) GetETFList(c *gin.Context) {
	market := c.Query("market")

	var configs []models.ETFConfig
	query := models.DB.Where("status = ?", 1)
	if market != "" {
		query = query.Where("market = ?", market)
	}

	if err := query.Order("sort_order, symbol").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// GetComparison 获取ETF对比数据
func (h *ETFHandler) GetComparison(c *gin.Context) {
	period := c.DefaultQuery("period", "1y")

	// 获取所有启用的ETF
	var configs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var symbols []string
	for _, cfg := range configs {
		symbols = append(symbols, cfg.Symbol)
	}

	data, err := h.analysisSvc.GetComparisonData(symbols, period)
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

// GetPortfolioAnalysis 获取投资组合分析
func (h *ETFHandler) GetPortfolioAnalysis(c *gin.Context) {
	var req struct {
		Allocation      map[string]float64 `json:"allocation" binding:"required"`
		TotalInvestment float64            `json:"total_investment"`
		TaxRate         float64            `json:"tax_rate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试从query参数获取
		req.TotalInvestment = 10000
		req.TaxRate = 0.10
	}

	if req.TotalInvestment == 0 {
		req.TotalInvestment = 10000
	}

	totalInvestment := decimal.NewFromFloat(req.TotalInvestment)
	taxRate := decimal.NewFromFloat(req.TaxRate)

	result, err := h.analysisSvc.AnalyzePortfolio(req.Allocation, totalInvestment, taxRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetRealtimeData 获取实时数据
func (h *ETFHandler) GetRealtimeData(c *gin.Context) {
	symbol := c.Param("symbol")

	data, err := h.cacheService.GetRealtimeData(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
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

// GetMetrics 获取ETF指标
func (h *ETFHandler) GetMetrics(c *gin.Context) {
	symbol := c.Param("symbol")
	period := c.DefaultQuery("period", "1y")

	// 获取历史数据
	var prices []models.ETFData
	if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Limit(252).Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(prices) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "no data found",
		})
		return
	}

	metrics, err := h.analysisSvc.CalculateMetrics(symbol, prices, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetHistory 获取历史数据
func (h *ETFHandler) GetHistory(c *gin.Context) {
	symbol := c.Param("symbol")
	period := c.DefaultQuery("period", "1y")

	// 获取历史数据
	var prices []models.ETFData
	if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 根据period过滤数据
	// 这里简化处理，实际应该根据period计算日期范围

	var chartData []map[string]interface{}
	for _, price := range prices {
		chartData = append(chartData, map[string]interface{}{
			"date":   price.Date.Format("2006-01-02"),
			"open":   price.OpenPrice,
			"high":   price.HighPrice,
			"low":    price.LowPrice,
			"close":  price.ClosePrice,
			"volume": price.Volume,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"symbol": symbol,
			"period": period,
			"prices": chartData,
		},
	})
}

// GetForecast 获取收益预测
func (h *ETFHandler) GetForecast(c *gin.Context) {
	symbol := c.Param("symbol")

	var req struct {
		InitialInvestment float64 `form:"initial_investment" json:"initial_investment"`
		AnnualReturnRate  float64 `form:"annual_return_rate" json:"annual_return_rate"`
		TaxRate           float64 `form:"tax_rate" json:"tax_rate"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		req.InitialInvestment = 10000
		req.TaxRate = 0.10
	}

	initialInvestment := decimal.NewFromFloat(req.InitialInvestment)
	taxRate := decimal.NewFromFloat(req.TaxRate)

	var annualReturnRate *decimal.Decimal
	if req.AnnualReturnRate > 0 {
		rate := decimal.NewFromFloat(req.AnnualReturnRate)
		annualReturnRate = &rate
	}

	result, err := h.analysisSvc.ForecastETFGrowth(symbol, initialInvestment, annualReturnRate, taxRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// UpdateRealtimeData 更新实时数据
func (h *ETFHandler) UpdateRealtimeData(c *gin.Context) {
	// 获取所有启用的ETF
	var configs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 创建Yahoo Finance客户端
	client := services.NewYahooFinanceClient()

	var symbols []string
	for _, cfg := range configs {
		symbols = append(symbols, cfg.Symbol)
	}

	// 获取实时数据
	quotes, err := client.GetQuotes(symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 更新缓存
	for _, quote := range quotes {
		realtimeData := &services.RealtimeData{
			Symbol:             quote.Symbol,
			Name:               quote.Name,
			CurrentPrice:       quote.CurrentPrice,
			PreviousClose:      quote.PreviousClose,
			OpenPrice:          quote.OpenPrice,
			DayHigh:            quote.DayHigh,
			DayLow:             quote.DayLow,
			Volume:             quote.Volume,
			Change:             quote.Change,
			ChangePercent:      quote.ChangePercent,
			MarketCap:          quote.MarketCap,
			DividendYield:      quote.DividendYield,
			FiftyTwoWeekHigh:   quote.FiftyTwoWeekHigh,
			FiftyTwoWeekLow:    quote.FiftyTwoWeekLow,
			AverageVolume:      quote.AverageVolume,
			Beta:               quote.Beta,
			PERatio:            quote.PERatio,
			Currency:           quote.Currency,
			DataSource:         "yahoo_finance",
		}
		h.cacheService.SetRealtimeData(quote.Symbol, realtimeData)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Realtime data updated",
		"count":   len(quotes),
	})
}
