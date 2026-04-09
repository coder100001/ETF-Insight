package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/datasource"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// SimpleRealtimeData 简单的实时数据占位符类型（缓存移除后使用）
type SimpleRealtimeData struct {
	Symbol           string
	Name             string
	CurrentPrice     float64
	PreviousClose    float64
	OpenPrice        float64
	DayHigh          float64
	DayLow           float64
	Volume           int64
	Change           float64
	ChangePercent    float64
	MarketCap        int64
	DividendYield    float64
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	AverageVolume    int64
	Beta             float64
	PERatio          float64
	Currency         string
	DataSource       string
	UpdatedAt        time.Time
}

// ETFHandler ETF 相关处理器
type ETFHandler struct {
	analysisService *services.ETFAnalysisService
	provider        datasource.DataSourceProvider
}

// NewETFHandler 创建 ETF 处理器
func NewETFHandler(analysisService *services.ETFAnalysisService, provider datasource.DataSourceProvider) *ETFHandler {
	return &ETFHandler{
		analysisService: analysisService,
		provider:        provider,
	}
}

// GetETFList 获取 ETF 列表 - 数据全部从数据库/Finage获取，不使用缓存
func (h *ETFHandler) GetETFList(c *gin.Context) {
	// 获取所有启用的 ETF 配置
	var etfConfigs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&etfConfigs).Error; err != nil || len(etfConfigs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	// 并行获取数据
	type ETFResult struct {
		Symbol string
		Data   map[string]interface{}
		Error  error
	}

	results := make(chan ETFResult, len(etfConfigs))

	for _, cfg := range etfConfigs {
		go func(cfg models.ETFConfig) {
			result := ETFResult{Symbol: cfg.Symbol}
			result.Data, result.Error = h.getETFDetailData(cfg)
			results <- result
		}(cfg)
	}

	etfList := make([]map[string]interface{}, 0, len(etfConfigs))
	for i := 0; i < len(etfConfigs); i++ {
		result := <-results
		if result.Error == nil && result.Data != nil {
			etfList = append(etfList, result.Data)
		}
	}

	response := gin.H{
		"success": true,
		"data":    etfList,
	}

	c.JSON(http.StatusOK, response)
}

// getETFDetailData 获取单个 ETF 的详细数据
func (h *ETFHandler) getETFDetailData(cfg models.ETFConfig) (map[string]interface{}, error) {
	// 从数据库获取最新的行情数据
	var etfData models.ETFData
	err := models.DB.Where("symbol = ?", cfg.Symbol).
		Order("date DESC").
		First(&etfData).Error

	// 实时数据（移除缓存后，需要重新设计数据获取方式）
	var realtimeData *SimpleRealtimeData
	// TODO: 如果需要实时数据，可直接调用provider获取
	// quote, err := h.provider.GetRealtimeQuote(cfg.Symbol)
	// if err == nil {
	// 	realtimeData = &services.RealtimeData{
	// 		Symbol:           quote.Symbol,
	// 		CurrentPrice:     quote.CurrentPrice,
	// 		PreviousClose:    quote.PreviousClose,
	// 		OpenPrice:        quote.OpenPrice,
	// 		DayHigh:          quote.DayHigh,
	// 		DayLow:           quote.DayLow,
	// 		Volume:           quote.Volume,
	// 		Change:           quote.Change,
	// 		ChangePercent:    quote.ChangePercent,
	// 		DividendYield:    quote.DividendYield,
	// 		FiftyTwoWeekHigh: quote.FiftyTwoWeekHigh,
	// 		FiftyTwoWeekLow:  quote.FiftyTwoWeekLow,
	// 		DataSource:       quote.DataSource,
	// 		UpdatedAt:        quote.UpdatedAt,
	// 	}
	// }

	// 从数据库获取历史价格数据计算指标
	var prices []models.ETFData
	models.DB.Where("symbol = ?", cfg.Symbol).Order("date DESC").Limit(252).Find(&prices)
	metrics := calculateMetricsFromPrices(prices, "1y")

	return h.buildETFResult(cfg, etfData, realtimeData, metrics, err == nil && etfData.ID > 0), nil
}

// buildETFResult 构建 ETF 结果数据
func (h *ETFHandler) buildETFResult(cfg models.ETFConfig, etfData models.ETFData, realtimeData *SimpleRealtimeData, metrics *HandlerMetrics, hasData bool) map[string]interface{} {
	if hasData {
		// 涨跌计算：基于前一日收盘价
		// 优先从 realtimeData 获取 previousClose
		previousClose := 0.0
		if realtimeData != nil && realtimeData.PreviousClose > 0 {
			previousClose = realtimeData.PreviousClose
		} else {
			// 从数据库获取前一日数据作为 previousClose
			var prevData models.ETFData
			if err := models.DB.Where("symbol = ? AND date < ?", cfg.Symbol, etfData.Date).
				Order("date DESC").First(&prevData).Error; err == nil && prevData.ID > 0 {
				previousClose = prevData.ClosePrice.InexactFloat64()
			}
		}

		// 如果仍无 previousClose，使用 OpenPrice 作为近似值（不理想但好于0）
		if previousClose == 0 {
			previousClose = etfData.OpenPrice.InexactFloat64()
		}

		change := etfData.ClosePrice.InexactFloat64() - previousClose
		changePercent := 0.0
		if previousClose > 0 {
			changePercent = (change / previousClose) * 100
		}

		// 根据 ETF 类型设置合理的默认股息率
		defaultDividendYield := getDefaultDividendYield(cfg.Symbol)

		result := map[string]interface{}{
			"symbol":              cfg.Symbol,
			"name":                cfg.Name,
			"market":              "US",
			"category":            cfg.Category,
			"current_price":       etfData.ClosePrice.InexactFloat64(),
			"previous_close":      previousClose,
			"change":              math.Round(change*100) / 100,
			"change_percent":      math.Round(changePercent*100) / 100,
			"open_price":          etfData.OpenPrice.InexactFloat64(),
			"high_price":          etfData.HighPrice.InexactFloat64(),
			"low_price":           etfData.LowPrice.InexactFloat64(),
			"volume":              etfData.Volume,
			"market_cap":          cfg.AUM.InexactFloat64(),
			"dividend_yield":      defaultDividendYield,
			"fifty_two_week_high": 0.0,
			"fifty_two_week_low":  0.0,
			"currency":            cfg.Currency,
			"focus":               cfg.Focus,
			"strategy":            cfg.Strategy,
			"data_source":         etfData.DataSource,
			"volatility":          metrics.Volatility,
			"total_return":        metrics.TotalReturn,
			"max_drawdown":        metrics.MaxDrawdown,
			"sharpe_ratio":        metrics.SharpeRatio,
			"expense_ratio":       cfg.ExpenseRatio.InexactFloat64() * 100,
		}

		// 合并缓存数据中的额外字段
		if realtimeData != nil {
			result["market_cap"] = realtimeData.MarketCap
			result["dividend_yield"] = realtimeData.DividendYield
			result["fifty_two_week_high"] = realtimeData.FiftyTwoWeekHigh
			result["fifty_two_week_low"] = realtimeData.FiftyTwoWeekLow
		}

		return result
	}

	// 没有数据库数据，返回基本信息
	return map[string]interface{}{
		"symbol":         cfg.Symbol,
		"name":           cfg.Name,
		"market":         "US",
		"category":       cfg.Category,
		"current_price":  0.0,
		"change":         0.0,
		"change_percent": 0.0,
		"dividend_yield": 0.0,
		"currency":       cfg.Currency,
		"focus":          cfg.Focus,
		"strategy":       cfg.Strategy,
		"volatility":     metrics.Volatility,
		"total_return":   metrics.TotalReturn,
		"max_drawdown":   metrics.MaxDrawdown,
		"sharpe_ratio":   metrics.SharpeRatio,
		"expense_ratio":  cfg.ExpenseRatio.InexactFloat64() * 100,
	}
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

	// 从数据库获取最新的 OHLCV 数据
	var etfData models.ETFData
	result := models.DB.Where("symbol = ?", symbol).Order("date DESC").First(&etfData)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF data not found",
		})
		return
	}

	// 获取 ETF 配置信息
	var etfConfig models.ETFConfig
	models.DB.Where("symbol = ?", symbol).First(&etfConfig)

	// 计算涨跌幅 - 基于前一日收盘价
	var prevData models.ETFData
	previousClose := etfData.OpenPrice.InexactFloat64() // 默认使用开盘价作为近似
	if err := models.DB.Where("symbol = ? AND date < ?", symbol, etfData.Date).
		Order("date DESC").First(&prevData).Error; err == nil && prevData.ID > 0 {
		previousClose = prevData.ClosePrice.InexactFloat64()
	}

	change := etfData.ClosePrice.Sub(decimal.NewFromFloat(previousClose))
	changePercent := decimal.Zero
	if previousClose > 0 {
		changePercent = change.Div(decimal.NewFromFloat(previousClose)).Mul(decimal.NewFromInt(100))
	}

	// 根据 ETF 类型设置合理的默认股息率
	defaultDividendYield := getDefaultDividendYield(symbol)

	data := map[string]interface{}{
		"symbol":         symbol,
		"name":           etfConfig.Name,
		"current_price":  etfData.ClosePrice.InexactFloat64(),
		"previous_close": previousClose,
		"open_price":     etfData.OpenPrice.InexactFloat64(),
		"high_price":     etfData.HighPrice.InexactFloat64(),
		"low_price":      etfData.LowPrice.InexactFloat64(),
		"day_high":       etfData.HighPrice.InexactFloat64(),
		"day_low":        etfData.LowPrice.InexactFloat64(),
		"volume":         etfData.Volume,
		"change":         change.InexactFloat64(),
		"change_percent": changePercent.InexactFloat64(),
		"market_cap":     etfConfig.AUM.InexactFloat64(),
		"dividend_yield": defaultDividendYield,
		"expense_ratio":  etfConfig.ExpenseRatio.InexactFloat64() * 100,
		"focus":          etfConfig.Focus,
		"strategy":       etfConfig.Strategy,
		"category":       etfConfig.Category,
		"provider":       etfConfig.Provider,
		"currency":       "USD",
		"data_source":    etfData.DataSource,
		"updated_at":     etfData.Date,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetETFComparison 获取 ETF 对比数据
func (h *ETFHandler) GetETFComparison(c *gin.Context) {
	period := c.DefaultQuery("period", "1y")

	// 从数据库获取所有启用的ETF列表，不硬编码
	var etfConfigs []models.ETFConfig
	models.DB.Where("status = ?", 1).Find(&etfConfigs)

	symbols := make([]string, 0, len(etfConfigs))
	for _, cfg := range etfConfigs {
		symbols = append(symbols, cfg.Symbol)
	}

	if len(symbols) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	data, err := h.analysisService.GetComparisonData(symbols, period)
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

	var startDate time.Time
	switch period {
	case "1m":
		startDate = time.Now().AddDate(0, -1, 0)
	case "3m":
		startDate = time.Now().AddDate(0, -3, 0)
	case "6m":
		startDate = time.Now().AddDate(0, -6, 0)
	case "1y":
		startDate = time.Now().AddDate(-1, 0, 0)
	case "3y":
		startDate = time.Now().AddDate(-3, 0, 0)
	case "5y":
		startDate = time.Now().AddDate(-5, 0, 0)
	default:
		startDate = time.Now().AddDate(-1, 0, 0)
	}

	var prices []models.ETFData
	if err := models.DB.Where("symbol = ? AND date >= ?", symbol, startDate).
		Order("date ASC").
		Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch historical data",
		})
		return
	}

	var data []map[string]interface{}
	for _, price := range prices {
		data = append(data, map[string]interface{}{
			"date":        price.Date.Format("2006-01-02"),
			"open_price":  price.OpenPrice.InexactFloat64(),
			"close_price": price.ClosePrice.InexactFloat64(),
			"high_price":  price.HighPrice.InexactFloat64(),
			"low_price":   price.LowPrice.InexactFloat64(),
			"volume":      price.Volume,
			"price":       price.ClosePrice.InexactFloat64(),
		})
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

	var etfConfig models.ETFConfig
	if result := models.DB.Where("symbol = ?", symbol).First(&etfConfig); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF not found",
		})
		return
	}

	var prices []models.ETFData
	if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Limit(252).Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch price data",
		})
		return
	}

	if len(prices) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No price data available",
		})
		return
	}

	metrics := calculateMetricsFromPrices(prices, period)

	data := map[string]interface{}{
		"symbol":         symbol,
		"name":           etfConfig.Name,
		"expense_ratio":  etfConfig.ExpenseRatio.InexactFloat64() * 100,
		"dividend_yield": 0.0, // 从数据库/Finage获取，不硬编码
		"strategy":       etfConfig.Strategy,
		"focus":          etfConfig.Focus,
		"category":       etfConfig.Category,
		"provider":       etfConfig.Provider,
		"aum":            etfConfig.AUM.InexactFloat64(),
		"volatility":     metrics.Volatility,
		"total_return":   metrics.TotalReturn,
		"max_drawdown":   metrics.MaxDrawdown,
		"sharpe_ratio":   metrics.SharpeRatio,
		"period":         period,
		"data_source":    "database",
		"updated_at":     time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// HandlerMetrics 指标数据结构（用于 handlers 包）
type HandlerMetrics struct {
	Volatility  float64
	TotalReturn float64
	MaxDrawdown float64
	SharpeRatio float64
}

// calculateMetricsFromPrices 从历史价格计算指标
func calculateMetricsFromPrices(prices []models.ETFData, period string) *HandlerMetrics {
	if len(prices) < 2 {
		return &HandlerMetrics{}
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		prevPrice := prices[i].ClosePrice.InexactFloat64()
		currPrice := prices[i-1].ClosePrice.InexactFloat64()
		if prevPrice > 0 {
			returns[i-1] = (currPrice - prevPrice) / prevPrice
		}
	}

	firstPrice := prices[len(prices)-1].ClosePrice.InexactFloat64()
	lastPrice := prices[0].ClosePrice.InexactFloat64()
	totalReturn := 0.0
	if firstPrice > 0 {
		totalReturn = (lastPrice - firstPrice) / firstPrice * 100
	}

	volatility := calculateVolatility(returns)
	maxDrawdown := calculateMaxDrawdown(prices)
	sharpeRatio := calculateSharpeRatio(returns, 0.02)

	return &HandlerMetrics{
		Volatility:  volatility,
		TotalReturn: totalReturn,
		MaxDrawdown: maxDrawdown,
		SharpeRatio: sharpeRatio,
	}
}

func calculateVolatility(returns []float64) float64 {
	if len(returns) < 10 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	return stdDev * math.Sqrt(252) * 100
}

func calculateMaxDrawdown(prices []models.ETFData) float64 {
	if len(prices) < 10 {
		return 0
	}

	maxDrawdown := 0.0
	peak := prices[len(prices)-1].ClosePrice.InexactFloat64()

	for i := len(prices) - 1; i >= 0; i-- {
		price := prices[i].ClosePrice.InexactFloat64()
		if price > peak {
			peak = price
		}
		drawdown := (peak - price) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown * 100
}

func calculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) < 10 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	meanReturn := sum / float64(len(returns))

	annualizedReturn := meanReturn * 252

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-meanReturn, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))
	annualizedStdDev := stdDev * math.Sqrt(252)

	if annualizedStdDev == 0 {
		return 0
	}

	return (annualizedReturn - riskFreeRate) / annualizedStdDev
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

	initialInvestment, err := strconv.ParseFloat(initialInvestmentStr, 64)
	if err != nil || initialInvestment <= 0 {
		initialInvestment = 10000
	}

	taxRate, err := strconv.ParseFloat(taxRateStr, 64)
	if err != nil || taxRate < 0 || taxRate > 1 {
		taxRate = 0.10
	}

	// 获取ETF配置
	var etfConfig models.ETFConfig
	models.DB.Where("symbol = ?", symbol).First(&etfConfig)

	// 从数据库获取历史数据计算预期收益率
	dividendYield := 0.0
	expenseRatio := 0.0020
	if etfConfig.ExpenseRatio.GreaterThan(decimal.Zero) {
		expenseRatio = etfConfig.ExpenseRatio.InexactFloat64()
	}

	// 从历史数据计算年化收益率
	expectedAnnualReturn := 0.08 // 默认8%
	var prices []models.ETFData
	if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Limit(252).Find(&prices).Error; err == nil && len(prices) >= 30 {
		firstPrice := prices[len(prices)-1].ClosePrice.InexactFloat64()
		lastPrice := prices[0].ClosePrice.InexactFloat64()
		if firstPrice > 0 {
			// 简单年化收益率计算
			days := len(prices)
			totalReturn := (lastPrice - firstPrice) / firstPrice
			expectedAnnualReturn = totalReturn * (252.0 / float64(days))
			if expectedAnnualReturn < -0.5 {
				expectedAnnualReturn = -0.5
			}
			if expectedAnnualReturn > 0.5 {
				expectedAnnualReturn = 0.5
			}
		}
	}

	// 获取实时数据中的dividend_yield（移除缓存后需要重新设计）
	// TODO: 如果需要实时股息率，可直接调用provider获取
	// realtimeData, _ := h.provider.GetRealtimeQuote(symbol)
	// if realtimeData != nil && realtimeData.DividendYield > 0 {
	// 	dividendYield = realtimeData.DividendYield / 100
	// }

	// 计算预测
	forecast := make([]map[string]interface{}, 10)
	currentValue := initialInvestment
	cumulativeDividend := 0.0

	for year := 1; year <= 10; year++ {
		capitalGrowth := currentValue * expectedAnnualReturn
		dividendIncome := currentValue * dividendYield
		afterTaxDividend := dividendIncome * (1 - taxRate)
		fees := currentValue * expenseRatio

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

	totalReturn := currentValue - initialInvestment
	totalReturnPercent := (totalReturn / initialInvestment) * 100

	name := symbol
	if etfConfig.Name != "" {
		name = etfConfig.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"symbol":                 symbol,
			"name":                   name,
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

// UpdateRealtimeData 更新实时数据 - 使用Finage数据源获取完整OHLCV并入库
func (h *ETFHandler) UpdateRealtimeData(c *gin.Context) {
	// 获取所有启用的ETF
	var etfConfigs []models.ETFConfig
	models.DB.Where("status = ?", 1).Find(&etfConfigs)

	if len(etfConfigs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No enabled ETF configs found",
			"count":   0,
		})
		return
	}

	symbols := make([]string, 0, len(etfConfigs))
	for _, cfg := range etfConfigs {
		symbols = append(symbols, cfg.Symbol)
	}

	// 使用 Finage 数据源获取数据
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if h.provider == nil || !h.provider.IsAvailable(ctx) {
		providerName := "nil"
		if h.provider != nil {
			providerName = h.provider.GetName()
		}
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Data source not available: " + providerName,
		})
		return
	}

	quotes, err := h.provider.GetQuotes(ctx, symbols)
	if err != nil {
		utils.Error("Failed to get quotes", err, "provider", h.provider.GetName())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "Failed to get data: " + err.Error(),
		})
		return
	}

	// 更新缓存和数据库
	successCount := 0
	for _, quote := range quotes {
		// 缓存已移除，不再创建RealtimeData对象
		// realtimeData := &SimpleRealtimeData{
		//	Symbol:           quote.Symbol,
		//	CurrentPrice:     quote.CurrentPrice,
		//	PreviousClose:    quote.PreviousClose,
		//	OpenPrice:        quote.OpenPrice,
		//	DayHigh:          quote.DayHigh,
		//	DayLow:           quote.DayLow,
		//	Volume:           quote.Volume,
		//	Change:           quote.Change,
		//	ChangePercent:    quote.ChangePercent,
		//	DividendYield:    quote.DividendYield,
		//	DataSource:       quote.DataSource,
		//	UpdatedAt:        time.Now(),
		// }
		// 缓存已移除：h.cacheService.SetRealtimeData(quote.Symbol, realtimeData)

		// 入库完整OHLCV数据
		if quote.OpenPrice > 0 || quote.DayHigh > 0 || quote.DayLow > 0 || quote.Volume > 0 {
			date := quote.Timestamp
			if date.IsZero() {
				date = time.Now().Truncate(24 * time.Hour)
			}

			etfData := models.ETFData{
				Symbol:     quote.Symbol,
				Date:       date,
				OpenPrice:  decimal.NewFromFloat(quote.OpenPrice),
				ClosePrice: decimal.NewFromFloat(quote.CurrentPrice),
				HighPrice:  decimal.NewFromFloat(quote.DayHigh),
				LowPrice:   decimal.NewFromFloat(quote.DayLow),
				Volume:     quote.Volume,
				DataSource: quote.DataSource,
			}

			models.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "symbol"}, {Name: "date"}},
				DoUpdates: clause.AssignmentColumns([]string{"open_price", "close_price", "high_price", "low_price", "volume", "data_source"}),
			}).Create(&etfData)
		}

		successCount++
	}

	// 缓存已移除：h.cacheService.ClearCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Realtime data updated successfully from " + h.provider.GetName(),
		"count":   successCount,
		"source":  h.provider.GetName(),
	})
}

// getDefaultDividendYield 根据 ETF 类型返回合理的默认股息率（百分比）
func getDefaultDividendYield(symbol string) float64 {
	// 高股息 ETF
	if symbol == "SCHD" || symbol == "VYM" || symbol == "SPYD" || symbol == "HDV" || symbol == "DGRO" {
		return 3.5 // 3.5%
	}
	// 覆盖收益型 ETF
	if symbol == "JEPI" || symbol == "JEPQ" || symbol == "QYLD" || symbol == "XYLD" {
		return 7.0 // 7%
	}
	// 债券 ETF
	if symbol == "BND" || symbol == "AGG" || symbol == "TLT" {
		return 4.0 // 4%
	}
	// 房地产 ETF
	if symbol == "VNQ" {
		return 4.0 // 4%
	}
	// 黄金 ETF
	if symbol == "GLD" {
		return 0.0 // 0%
	}
	// 宽基指数 ETF
	if symbol == "QQQ" || symbol == "VOO" || symbol == "VTI" || symbol == "SPY" {
		return 0.5 // 0.5%
	}
	// 国际市场 ETF
	if symbol == "VEA" || symbol == "VWO" || symbol == "VXUS" {
		return 3.0 // 3%
	}
	// 默认
	return 1.0 // 1%
}
