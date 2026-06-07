package handlers

import (
	"net/http"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// PortfolioRiskRequest 组合风险分析请求
type PortfolioRiskRequest struct {
	Portfolio  map[string]float64 `json:"portfolio" binding:"required"`
	Confidence float64            `json:"confidence"`
	Period     string             `json:"period"`
}

// AnalyzePortfolioRisk 分析投资组合风险
func (h *PortfolioHandler) AnalyzePortfolioRisk(c *gin.Context) {
	var req PortfolioRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if len(req.Portfolio) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Portfolio cannot be empty",
		})
		return
	}

	// 设置默认值
	confidence := req.Confidence
	if confidence <= 0 || confidence >= 1 {
		confidence = 0.95
	}

	period := req.Period
	if period == "" {
		period = "1y"
	}

	// 获取历史数据天数
	days := 252
	switch period {
	case "3m":
		days = 63
	case "6m":
		days = 126
	case "1y":
		days = 252
	case "3y":
		days = 756
	case "5y":
		days = 1260
	}

	// 获取各ETF的历史数据
	returns := make(map[string][]decimal.Decimal)
	etfPrices := make(map[string][]models.ETFData)

	for symbol := range req.Portfolio {
		var etfData []models.ETFData
		if err := models.DB.Where("symbol = ?", symbol).
			Order("date DESC").
			Limit(days).
			Find(&etfData).Error; err != nil || len(etfData) < 30 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Insufficient historical data for " + symbol,
			})
			return
		}
		etfPrices[symbol] = etfData

		// 计算收益率序列(使用公共函数)
		symbolReturns := utils.CalculateReturnsFromETFData(etfData)
		returns[symbol] = symbolReturns
	}

	// 计算组合收益率序列
	portfolioReturns := make([]decimal.Decimal, 0)
	numObservations := len(returns[getFirstKey(returns)])

	for i := 0; i < numObservations; i++ {
		portfolioReturn := decimal.Zero
		for symbol, symbolReturns := range returns {
			if i < len(symbolReturns) {
				weight := decimal.NewFromFloat(req.Portfolio[symbol])
				portfolioReturn = portfolioReturn.Add(symbolReturns[i].Mul(weight))
			}
		}
		portfolioReturns = append(portfolioReturns, portfolioReturn)
	}

	// 使用风险模型计算指标
	riskModels := services.NewRiskModels()

	// 计算组合VaR和CVaR
	varData, err := riskModels.CalculateHistoricalVaR(portfolioReturns, confidence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to calculate portfolio VaR: " + err.Error(),
		})
		return
	}

	// 计算综合风险指标
	riskFreeRate := decimal.NewFromFloat(config.GetFinancialConfig().RiskFreeRate / float64(config.GetFinancialConfig().TradingDaysYear))
	riskMetrics, err := riskModels.CalculateRiskMetrics(portfolioReturns, riskFreeRate, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to calculate risk metrics: " + err.Error(),
		})
		return
	}

	// 计算各资产的风险贡献
	portfolioRisks := make([]map[string]any, 0, len(req.Portfolio))
	for symbol, weight := range req.Portfolio {
		if assetReturns, ok := returns[symbol]; ok && len(assetReturns) > 0 {
			assetVarData, _ := riskModels.CalculateHistoricalVaR(assetReturns, confidence)
			componentVar := assetVarData.VaR.Mul(decimal.NewFromFloat(weight))
			marginalVar := decimal.Zero
			if varData.VaR.GreaterThan(decimal.Zero) {
				marginalVar = assetVarData.VaR.Div(varData.VaR)
			}

			portfolioRisks = append(portfolioRisks, map[string]any{
				"symbol":       symbol,
				"weight":       weight,
				"componentVar": componentVar.Mul(decimal.NewFromInt(100)).InexactFloat64(),
				"marginalVar":  marginalVar.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			})
		}
	}

	// 获取风险等级
	riskLevel := getRiskLevel(riskMetrics.Volatility.InexactFloat64() * 100)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]any{
			"portfolio":       req.Portfolio,
			"period":          period,
			"confidence":      confidence,
			"risk_level":      riskLevel,
			"var_95":          varData.VaR.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			"var_99":          varData.VaR.Mul(decimal.NewFromFloat(1.3)).Mul(decimal.NewFromInt(100)).InexactFloat64(), // 近似99% VaR
			"cvar_95":         varData.CVaR.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			"volatility":      riskMetrics.Volatility.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			"sharpe_ratio":    riskMetrics.SharpeRatio.InexactFloat64(),
			"sortino_ratio":   riskMetrics.SortinoRatio.InexactFloat64(),
			"max_drawdown":    riskMetrics.MaxDrawdown.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			"calmar_ratio":    riskMetrics.CalmarRatio.InexactFloat64(),
			"beta":            riskMetrics.Beta.InexactFloat64(),
			"alpha":           riskMetrics.Alpha.InexactFloat64(),
			"portfolio_risks": portfolioRisks,
			"data_points":     len(portfolioReturns),
		},
	})
}

// getFirstKey 获取map的第一个key
func getFirstKey(m map[string][]decimal.Decimal) string {
	for k := range m {
		return k
	}
	return ""
}

// getRiskLevel 根据波动率获取风险等级
func getRiskLevel(volatility float64) string {
	if volatility < 10 {
		return "low"
	} else if volatility < 15 {
		return "medium"
	}
	return "high"
}
