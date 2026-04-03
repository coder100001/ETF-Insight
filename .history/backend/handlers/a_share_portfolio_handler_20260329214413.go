package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ASharePortfolioHandler A股ETF组合处理器
type ASharePortfolioHandler struct{}

// NewASharePortfolioHandler 创建新的处理器
func NewASharePortfolioHandler() *ASharePortfolioHandler {
	return &ASharePortfolioHandler{}
}

// GetDefaultETFs 获取默认A股红利ETF列表
func (h *ASharePortfolioHandler) GetDefaultETFs(c *gin.Context) {
	etfs := getDefaultAShareETFs()
	
	// 保存到数据库（如果不存在）
	for i := range etfs {
		var existing models.AShareDividendETF
		models.DB.First(&existing, "symbol = ?", etfs[i].Symbol)
		if existing.ID == 0 {
			models.DB.Create(&etfs[i])
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    etfs,
	})
}

// GetDefaultPortfolio 获取默认组合配置
func (h *ASharePortfolioHandler) GetDefaultPortfolio(c *gin.Context) {
	// 获取默认ETF
	etfs := getDefaultAShareETFs()
	
	// 默认投资金额配置（单位：万元）
	defaultInvestments := map[string]float64{
		"515080": 12.5,  // 中证红利ETF
		"515180": 10.0,  // 红利ETF
		"515300": 15.0,  // 中证红利低波动
		"510720": 8.0,   // 红利国企ETF
		"520900": 10.0,  // 红利低波ETF
		"159545": 7.5,   // 红利ETF易方达
		"520550": 5.0,   // 红利质量ETF
		"513820": 5.0,   // 港股红利ETF
	}
	
	// 计算总投资
	var totalInvestment decimal.Decimal
	for _, amount := range defaultInvestments {
		totalInvestment = totalInvestment.Add(decimal.NewFromFloat(amount * 10000)) // 转换为元
	}
	
	// 构建持仓明细
	var holdings []models.AShareHoldingDetail
	for _, etf := range etfs {
		investment := decimal.NewFromFloat(defaultInvestments[etf.Symbol] * 10000)
		weight := investment.Div(totalInvestment).Mul(decimal.NewFromInt(100))
		
		// 计算股息率（取中间值）
		dividendYield := etf.DividendYieldMin.Add(etf.DividendYieldMax).Div(decimal.NewFromInt(2))
		
		// 计算预期年分红：投资金额 × 股息率 / 100
		expectedDividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))
		
		holdings = append(holdings, models.AShareHoldingDetail{
			Symbol:            etf.Symbol,
			Name:              etf.Name,
			Investment:        investment,
			Weight:            weight,
			DividendYield:     dividendYield,
			DividendFrequency: string(etf.DividendFrequency),
			ExpectedDividend:  expectedDividend,
		})
	}
	
	// 计算总预期分红
	var totalDividend decimal.Decimal
	for _, h := range holdings {
		totalDividend = totalDividend.Add(h.ExpectedDividend)
	}
	
	// 计算各持仓分红贡献占比
	for i := range holdings {
		if totalDividend.IsPositive() {
			holdings[i].DividendContribution = holdings[i].ExpectedDividend.Div(totalDividend).Mul(decimal.NewFromInt(100))
		}
	}
	
	// 计算平均股息率
	avgDividendYield := decimal.Zero
	if totalInvestment.IsPositive() {
		avgDividendYield = totalDividend.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	}
	
	result := models.AShareDividendCalculation{
		TotalInvestment:        totalInvestment,
		ExpectedAnnualDividend: totalDividend,
		AverageDividendYield:   avgDividendYield,
		MonthlyDividend:        totalDividend.Div(decimal.NewFromInt(12)),
		QuarterlyDividend:      totalDividend.Div(decimal.NewFromInt(4)),
		Holdings:               holdings,
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// AnalyzePortfolio 分析组合收益
func (h *ASharePortfolioHandler) AnalyzePortfolio(c *gin.Context) {
	var req struct {
		Investments map[string]float64 `json:"investments"` // ETF代码:投资金额(元)
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}
	
	// 获取ETF信息
	etfs := getDefaultAShareETFs()
	etfMap := make(map[string]models.AShareDividendETF)
	for _, etf := range etfs {
		etfMap[etf.Symbol] = etf
	}
	
	// 计算总投资
	var totalInvestment decimal.Decimal
	for _, amount := range req.Investments {
		totalInvestment = totalInvestment.Add(decimal.NewFromFloat(amount))
	}
	
	// 构建持仓明细
	var holdings []models.AShareHoldingDetail
	for symbol, amount := range req.Investments {
		etf, exists := etfMap[symbol]
		if !exists {
			continue
		}
		
		investment := decimal.NewFromFloat(amount)
		weight := decimal.Zero
		if totalInvestment.IsPositive() {
			weight = investment.Div(totalInvestment).Mul(decimal.NewFromInt(100))
		}
		
		// 计算股息率（取中间值）
		dividendYield := etf.DividendYieldMin.Add(etf.DividendYieldMax).Div(decimal.NewFromInt(2))
		
		// 计算预期年分红
		expectedDividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))
		
		holdings = append(holdings, models.AShareHoldingDetail{
			Symbol:            symbol,
			Name:              etf.Name,
			Investment:        investment,
			Weight:            weight,
			DividendYield:     dividendYield,
			DividendFrequency: string(etf.DividendFrequency),
			ExpectedDividend:  expectedDividend,
		})
	}
	
	// 计算总预期分红
	var totalDividend decimal.Decimal
	for _, h := range holdings {
		totalDividend = totalDividend.Add(h.ExpectedDividend)
	}
	
	// 计算各持仓分红贡献占比
	for i := range holdings {
		if totalDividend.IsPositive() {
			holdings[i].DividendContribution = holdings[i].ExpectedDividend.Div(totalDividend).Mul(decimal.NewFromInt(100))
		}
	}
	
	// 计算平均股息率
	avgDividendYield := decimal.Zero
	if totalInvestment.IsPositive() {
		avgDividendYield = totalDividend.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	}
	
	result := models.AShareDividendCalculation{
		TotalInvestment:        totalInvestment,
		ExpectedAnnualDividend: totalDividend,
		AverageDividendYield:   avgDividendYield,
		MonthlyDividend:        totalDividend.Div(decimal.NewFromInt(12)),
		QuarterlyDividend:      totalDividend.Div(decimal.NewFromInt(4)),
		Holdings:               holdings,
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// UpdateHolding 更新持仓金额
func (h *ASharePortfolioHandler) UpdateHolding(c *gin.Context) {
	symbol := c.Param("symbol")
	
	var req struct {
		Investment float64 `json:"investment"` // 投资金额(元)
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}
	
	// 获取ETF信息
	etfs := getDefaultAShareETFs()
	var targetETF *models.AShareDividendETF
	for i := range etfs {
		if etfs[i].Symbol == symbol {
			targetETF = &etfs[i]
			break
		}
	}
	
	if targetETF == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF不存在",
		})
		return
	}
	
	investment := decimal.NewFromFloat(req.Investment)
	dividendYield := targetETF.DividendYieldMin.Add(targetETF.DividendYieldMax).Div(decimal.NewFromInt(2))
	expectedDividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":           symbol,
			"name":             targetETF.Name,
			"investment":       investment,
			"dividend_yield":   dividendYield,
			"expected_dividend": expectedDividend,
		},
	})
}

// getDefaultAShareETFs 获取默认A股红利ETF列表
func getDefaultAShareETFs() []models.AShareDividendETF {
	return []models.AShareDividendETF{
		{
			Symbol:            "515080",
			Name:              "中证红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.8),
			DividendYieldMax:  decimal.NewFromFloat(5.1),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "跟踪中证红利指数，选取沪深两市股息率较高的100只股票",
		},
		{
			Symbol:            "515180",
			Name:              "红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.5),
			DividendYieldMax:  decimal.NewFromFloat(5.0),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "上证红利指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.006),
			Description:       "跟踪上证红利指数，选取上海市场股息率较高的50只股票",
		},
		{
			Symbol:            "515300",
			Name:              "中证红利低波动",
			DividendYieldMin:  decimal.NewFromFloat(4.2),
			DividendYieldMax:  decimal.NewFromFloat(4.8),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利低波动指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "结合红利和低波动因子，选取低波动的高股息股票",
		},
		{
			Symbol:            "510720",
			Name:              "红利国企ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.0),
			DividendYieldMax:  decimal.NewFromFloat(4.6),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证国企红利指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "聚焦国企红利，选取高分红的国有企业",
		},
		{
			Symbol:            "520900",
			Name:              "红利低波ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.3),
			DividendYieldMax:  decimal.NewFromFloat(4.9),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利低波动指数",
			Exchange:          "SHZ",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "红利低波动策略，适合稳健型投资者",
		},
		{
			Symbol:            "159545",
			Name:              "红利ETF易方达",
			DividendYieldMin:  decimal.NewFromFloat(4.4),
			DividendYieldMax:  decimal.NewFromFloat(5.0),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利指数",
			Exchange:          "SHZ",
			ManagementFee:     decimal.NewFromFloat(0.0015),
			Description:       "低费率红利ETF，跟踪中证红利指数",
		},
		{
			Symbol:            "520550",
			Name:              "红利质量ETF",
			DividendYieldMin:  decimal.NewFromFloat(3.8),
			DividendYieldMax:  decimal.NewFromFloat(4.5),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利质量指数",
			Exchange:          "SHZ",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "结合红利和质量因子，选取优质高分红股票",
		},
		{
			Symbol:            "513820",
			Name:              "港股红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(5.5),
			DividendYieldMax:  decimal.NewFromFloat(6.5),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证港股通高股息指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.006),
			Description:       "投资港股高股息标的，分散A股单一市场风险",
		},
	}
}

// CalculateDividendByFrequency 按频率计算分红
func (h *ASharePortfolioHandler) CalculateDividendByFrequency(c *gin.Context) {
	frequency := c.Param("frequency") // monthly/quarterly/yearly
	
	// 获取默认组合
	etfs := getDefaultAShareETFs()
	defaultInvestments := map[string]float64{
		"515080": 125000,
		"515180": 100000,
		"515300": 150000,
		"510720": 80000,
		"520900": 100000,
		"159545": 75000,
		"520550": 50000,
		"513820": 50000,
	}
	
	var result []gin.H
	
	for _, etf := range etfs {
		investment := decimal.NewFromFloat(defaultInvestments[etf.Symbol])
		dividendYield := etf.DividendYieldMin.Add(etf.DividendYieldMax).Div(decimal.NewFromInt(2))
		annualDividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))
		
		var periodDividend decimal.Decimal
		switch frequency {
		case "monthly":
			if etf.DividendFrequency == models.FrequencyMonthly {
				periodDividend = annualDividend.Div(decimal.NewFromInt(12))
			}
		case "quarterly":
			if etf.DividendFrequency == models.FrequencyQuarterly {
				periodDividend = annualDividend.Div(decimal.NewFromInt(4))
			}
		case "yearly":
			periodDividend = annualDividend
		default:
			periodDividend = annualDividend
		}
		
		result = append(result, gin.H{
			"symbol":          etf.Symbol,
			"name":            etf.Name,
			"investment":      investment,
			"period_dividend": periodDividend,
			"annual_dividend": annualDividend,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
