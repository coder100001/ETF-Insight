package handlers

import (
	"etf-insight/constants"
	"net/http"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/ashare"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ASharePortfolioHandler A股ETF组合处理器
type ASharePortfolioHandler struct {
	etfService *ashare.ETFDataService
}

// NewASharePortfolioHandler 创建新的处理器
func NewASharePortfolioHandler() *ASharePortfolioHandler {
	return &ASharePortfolioHandler{
		etfService: ashare.NewETFDataService(models.DB),
	}
}

// AShareHoldingDetailResponse 持仓明细响应（使用float64）
type AShareHoldingDetailResponse struct {
	Symbol               string  `json:"symbol"`                // ETF代码
	Name                 string  `json:"name"`                  // ETF名称
	CurrentPrice         float64 `json:"current_price"`         // 当前价格
	PreviousClose        float64 `json:"previous_close"`        // 昨日收盘价
	PriceChange          float64 `json:"price_change"`          // 价格变动
	PriceChangePct       float64 `json:"price_change_pct"`      // 涨跌幅(%)
	Volume               int64   `json:"volume"`                // 成交量
	Turnover             float64 `json:"turnover"`              // 成交额
	Investment           float64 `json:"investment"`            // 投资金额
	Weight               float64 `json:"weight"`                // 占比
	DividendYield        float64 `json:"dividend_yield"`        // 股息率(取中间值)
	DividendFrequency    string  `json:"dividend_frequency"`    // 分红频率
	ExpectedDividend     float64 `json:"expected_dividend"`     // 预期年分红
	DividendContribution float64 `json:"dividend_contribution"` // 分红贡献占比
}

// AShareDividendCalculationResponse 分红收益计算响应（使用float64）
type AShareDividendCalculationResponse struct {
	PortfolioID            uint                          `json:"portfolio_id"`
	TotalInvestment        float64                       `json:"total_investment"`         // 总投资金额
	ExpectedAnnualDividend float64                       `json:"expected_annual_dividend"` // 预期年分红总额
	AverageDividendYield   float64                       `json:"average_dividend_yield"`   // 平均股息率
	MonthlyDividend        float64                       `json:"monthly_dividend"`         // 月均分红
	QuarterlyDividend      float64                       `json:"quarterly_dividend"`       // 季均分红
	Holdings               []AShareHoldingDetailResponse `json:"holdings"`                 // 持仓明细
}

// getOrCreateDefaultETFs 从数据库获取核心ETF列表（只返回组合内8个ETF）
func (h *ASharePortfolioHandler) getOrCreateDefaultETFs() []models.AShareDividendETF {
	var etfs []models.AShareDividendETF
	result := models.DB.Where("symbol IN ?", constants.CoreETFSymbols).Find(&etfs)
	if result.Error == nil && len(etfs) == len(constants.CoreETFSymbols) {
		return etfs
	}

	defaultETFs := []models.AShareDividendETF{
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
			Status:            1,
		},
		{
			Symbol:            "515180",
			Name:              "红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.4),
			DividendYieldMax:  decimal.NewFromFloat(4.5),
			DividendFrequency: models.FrequencyYearly,
			Benchmark:         "上证红利指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.006),
			Description:       "跟踪上证红利指数，选取上海市场股息率较高的50只股票",
			Status:            1,
		},
		{
			Symbol:            "515300",
			Name:              "红利低波ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.4),
			DividendYieldMax:  decimal.NewFromFloat(4.5),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证红利低波动指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "结合红利和低波动因子，选取低波动的高股息股票",
			Status:            1,
		},
		{
			Symbol:            "510720",
			Name:              "红利国企ETF",
			DividendYieldMin:  decimal.NewFromFloat(3.5),
			DividendYieldMax:  decimal.NewFromFloat(4.0),
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "中证国企红利指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "聚焦国企红利，选取高分红的国有企业",
			Status:            1,
		},
		{
			Symbol:            "520900",
			Name:              "港股红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(5.7),
			DividendYieldMax:  decimal.NewFromFloat(5.7),
			DividendFrequency: models.FrequencyQuarterly,
			Benchmark:         "中证港股通高股息指数",
			Exchange:          "SHZ",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "投资港股高股息标的，分散A股单一市场风险",
			Status:            1,
		},
		{
			Symbol:            "159545",
			Name:              "港股低波ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.0),
			DividendYieldMax:  decimal.NewFromFloat(4.0),
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "中证港股通低波动指数",
			Exchange:          "SHZ",
			ManagementFee:     decimal.NewFromFloat(0.0015),
			Description:       "港股低波动策略，选取波动率较低的港股",
			Status:            1,
		},
		{
			Symbol:            "520550",
			Name:              "恒生红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(4.0),
			DividendYieldMax:  decimal.NewFromFloat(4.0),
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "恒生高股息率指数",
			Exchange:          "SHZ",
			ManagementFee:     decimal.NewFromFloat(0.005),
			Description:       "跟踪恒生高股息率指数，投资港股高分红股票",
			Status:            1,
		},
		{
			Symbol:            "513820",
			Name:              "港股通红利ETF",
			DividendYieldMin:  decimal.NewFromFloat(5.0),
			DividendYieldMax:  decimal.NewFromFloat(5.0),
			DividendFrequency: models.FrequencyMonthly,
			Benchmark:         "中证港股通高股息指数",
			Exchange:          "SSE",
			ManagementFee:     decimal.NewFromFloat(0.006),
			Description:       "通过港股通投资港股高股息标的",
			Status:            1,
		},
	}

	// 保存到数据库
	for i := range defaultETFs {
		models.DB.Create(&defaultETFs[i])
	}

	return defaultETFs
}

// GetDefaultETFs 获取默认A股红利ETF列表
func (h *ASharePortfolioHandler) GetDefaultETFs(c *gin.Context) {
	etfs := h.getOrCreateDefaultETFs()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    etfs,
	})
}

// GetDefaultPortfolio 获取默认组合配置
func (h *ASharePortfolioHandler) GetDefaultPortfolio(c *gin.Context) {
	// 从数据库获取ETF列表
	etfs := h.getOrCreateDefaultETFs()

	// 从数据库获取默认投资组合配置
	var portfolio models.AShareETFPortfolio
	result := models.DB.Where("is_default = ?", true).First(&portfolio)

	// 默认投资金额配置（单位：万元）
	defaultInvestments := map[string]float64{
		"515080": 12.5, // 中证红利ETF - 季分 股息4.8-5.1
		"515180": 5.0,  // 红利ETF - 年分 股息4.4-4.5
		"515300": 7.5,  // 红利低波ETF - 季分 股息4.4-4.5
		"510720": 10.0, // 红利国企ETF - 月分 股息3.5-4
		"520900": 7.5,  // 港股红利ETF - 季分 股息5.7
		"159545": 2.5,  // 港股低波ETF - 月分 股息4
		"520550": 2.5,  // 恒生红利ETF - 月分 股息4
		"513820": 2.5,  // 港股通红利ETF - 月分 股息5
	}

	// 如果数据库中有投资组合配置，使用数据库中的投资金额
	if result.Error == nil {
		var holdings []models.ASharePortfolioHolding
		models.DB.Where("portfolio_id = ?", portfolio.ID).Find(&holdings)
		for _, h := range holdings {
			// 查找对应的ETF symbol
			for _, etf := range etfs {
				if etf.ID == h.ETFID {
					defaultInvestments[etf.Symbol] = h.Investment.InexactFloat64() / 10000 // 转换为万元
					break
				}
			}
		}
	} else {
		// 创建默认投资组合
		portfolio = models.AShareETFPortfolio{
			Name:            "默认A股红利组合",
			Description:     "精选8只A股市场优质红利ETF",
			TotalInvestment: decimal.NewFromFloat(500000),
			IsDefault:       true,
		}
		models.DB.Create(&portfolio)

		// 创建持仓记录
		for _, etf := range etfs {
			investment := decimal.NewFromFloat(defaultInvestments[etf.Symbol] * 10000)
			holding := models.ASharePortfolioHolding{
				PortfolioID: portfolio.ID,
				ETFID:       etf.ID,
				Investment:  investment,
			}
			models.DB.Create(&holding)
		}
	}

	// 计算总投资
	var totalInvestment decimal.Decimal
	for _, amount := range defaultInvestments {
		totalInvestment = totalInvestment.Add(decimal.NewFromFloat(amount * 10000)) // 转换为元
	}

	// 构建持仓明细
	var holdings []AShareHoldingDetailResponse
	for _, etf := range etfs {
		investment := decimal.NewFromFloat(defaultInvestments[etf.Symbol] * 10000)
		weight := investment.Div(totalInvestment).Mul(decimal.NewFromInt(100))

		// 计算股息率（取中间值）
		dividendYield := etf.DividendYieldMin.Add(etf.DividendYieldMax).Div(decimal.NewFromInt(2))

		// 计算预期年分红：投资金额 × 股息率 / 100
		expectedDividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))

		holdings = append(holdings, AShareHoldingDetailResponse{
			Symbol:            etf.Symbol,
			Name:              etf.Name,
			CurrentPrice:      etf.CurrentPrice.InexactFloat64(),
			PreviousClose:     etf.PreviousClose.InexactFloat64(),
			PriceChange:       etf.PriceChange.InexactFloat64(),
			PriceChangePct:    etf.PriceChangePct.InexactFloat64(),
			Volume:            etf.Volume,
			Turnover:          etf.Turnover.InexactFloat64(),
			Investment:        investment.InexactFloat64(),
			Weight:            weight.InexactFloat64(),
			DividendYield:     dividendYield.InexactFloat64(),
			DividendFrequency: string(etf.DividendFrequency),
			ExpectedDividend:  expectedDividend.InexactFloat64(),
		})
	}

	// 计算总预期分红
	var totalDividend decimal.Decimal
	for _, h := range holdings {
		totalDividend = totalDividend.Add(decimal.NewFromFloat(h.ExpectedDividend))
	}

	// 计算各持仓分红贡献占比
	for i := range holdings {
		if totalDividend.IsPositive() {
			contribution := decimal.NewFromFloat(holdings[i].ExpectedDividend).Div(totalDividend).Mul(decimal.NewFromInt(100))
			holdings[i].DividendContribution = contribution.InexactFloat64()
		}
	}

	// 计算平均股息率
	avgDividendYield := decimal.Zero
	if totalInvestment.IsPositive() {
		avgDividendYield = totalDividend.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	}

	response := AShareDividendCalculationResponse{
		PortfolioID:            portfolio.ID,
		TotalInvestment:        totalInvestment.InexactFloat64(),
		ExpectedAnnualDividend: totalDividend.InexactFloat64(),
		AverageDividendYield:   avgDividendYield.InexactFloat64(),
		MonthlyDividend:        totalDividend.Div(decimal.NewFromInt(12)).InexactFloat64(),
		QuarterlyDividend:      totalDividend.Div(decimal.NewFromInt(4)).InexactFloat64(),
		Holdings:               holdings,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
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

	// 从数据库获取ETF信息
	etfs := h.getOrCreateDefaultETFs()
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
	var holdings []AShareHoldingDetailResponse
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

		holdings = append(holdings, AShareHoldingDetailResponse{
			Symbol:            symbol,
			Name:              etf.Name,
			Investment:        investment.InexactFloat64(),
			Weight:            weight.InexactFloat64(),
			DividendYield:     dividendYield.InexactFloat64(),
			DividendFrequency: string(etf.DividendFrequency),
			ExpectedDividend:  expectedDividend.InexactFloat64(),
		})
	}

	// 计算总预期分红
	var totalDividend decimal.Decimal
	for _, h := range holdings {
		totalDividend = totalDividend.Add(decimal.NewFromFloat(h.ExpectedDividend))
	}

	// 计算各持仓分红贡献占比
	for i := range holdings {
		if totalDividend.IsPositive() {
			contribution := decimal.NewFromFloat(holdings[i].ExpectedDividend).Div(totalDividend).Mul(decimal.NewFromInt(100))
			holdings[i].DividendContribution = contribution.InexactFloat64()
		}
	}

	// 计算平均股息率
	avgDividendYield := decimal.Zero
	if totalInvestment.IsPositive() {
		avgDividendYield = totalDividend.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	}

	result := AShareDividendCalculationResponse{
		TotalInvestment:        totalInvestment.InexactFloat64(),
		ExpectedAnnualDividend: totalDividend.InexactFloat64(),
		AverageDividendYield:   avgDividendYield.InexactFloat64(),
		MonthlyDividend:        totalDividend.Div(decimal.NewFromInt(12)).InexactFloat64(),
		QuarterlyDividend:      totalDividend.Div(decimal.NewFromInt(4)).InexactFloat64(),
		Holdings:               holdings,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ETFPriceResponse ETF价格响应
type ETFPriceResponse struct {
	Symbol         string  `json:"symbol"`           // ETF代码
	Name           string  `json:"name"`             // ETF名称
	CurrentPrice   float64 `json:"current_price"`    // 当前价格
	PreviousClose  float64 `json:"previous_close"`   // 昨日收盘价
	PriceChange    float64 `json:"price_change"`     // 价格变动
	PriceChangePct float64 `json:"price_change_pct"` // 涨跌幅(%)
	Volume         int64   `json:"volume"`           // 成交量
	Turnover       float64 `json:"turnover"`         // 成交额
	NAV            float64 `json:"nav"`              // 净值
	PremiumRate    float64 `json:"premium_rate"`     // 溢价率(%)
	PriceUpdatedAt string  `json:"price_updated_at"` // 价格更新时间
}

// GetETFPrices 获取核心ETF价格（通过Service层）
func (h *ASharePortfolioHandler) GetETFPrices(c *gin.Context) {
	etfs, err := h.etfService.GetCoreETFPrices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取ETF数据失败",
		})
		return
	}

	var response []ETFPriceResponse
	for _, etf := range etfs {
		response = append(response, ETFPriceResponse{
			Symbol:         etf.Symbol,
			Name:           etf.Name,
			CurrentPrice:   etf.CurrentPrice.InexactFloat64(),
			PreviousClose:  etf.PreviousClose.InexactFloat64(),
			PriceChange:    etf.PriceChange.InexactFloat64(),
			PriceChangePct: etf.PriceChangePct.InexactFloat64(),
			Volume:         etf.Volume,
			Turnover:       etf.Turnover.InexactFloat64(),
			NAV:            etf.NAV.InexactFloat64(),
			PremiumRate:    etf.PremiumRate.InexactFloat64(),
			PriceUpdatedAt: etf.PriceUpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetETFPriceBySymbol 获取单个ETF价格
func (h *ASharePortfolioHandler) GetETFPriceBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")

	var etf models.AShareDividendETF
	result := models.DB.Where("symbol = ? AND status = ?", symbol, 1).First(&etf)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF不存在",
		})
		return
	}

	response := ETFPriceResponse{
		Symbol:         etf.Symbol,
		Name:           etf.Name,
		CurrentPrice:   etf.CurrentPrice.InexactFloat64(),
		PreviousClose:  etf.PreviousClose.InexactFloat64(),
		PriceChange:    etf.PriceChange.InexactFloat64(),
		PriceChangePct: etf.PriceChangePct.InexactFloat64(),
		Volume:         etf.Volume,
		Turnover:       etf.Turnover.InexactFloat64(),
		NAV:            etf.NAV.InexactFloat64(),
		PremiumRate:    etf.PremiumRate.InexactFloat64(),
		PriceUpdatedAt: etf.PriceUpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// RefreshETFPrices 刷新ETF价格
func (h *ASharePortfolioHandler) RefreshETFPrices(c *gin.Context) {
	priceService := services.NewASharePriceService(models.DB)
	if err := priceService.UpdateAllETFPrices(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "刷新价格失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "价格刷新成功",
	})
}

// GetProviderStatus 获取数据源状态
func (h *ASharePortfolioHandler) GetProviderStatus(c *gin.Context) {
	priceService := services.NewASharePriceService(models.DB)
	statuses := priceService.GetProviderStatus()

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"providers": statuses,
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

	// 从数据库获取ETF信息
	var etf models.AShareDividendETF
	result := models.DB.Where("symbol = ?", symbol).First(&etf)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF不存在",
		})
		return
	}

	investment := decimal.NewFromFloat(req.Investment)
	dividendYield := etf.DividendYieldMin.Add(etf.DividendYieldMax).Div(decimal.NewFromInt(2))
	expectedDividend := investment.Mul(dividendYield).Div(decimal.NewFromInt(100))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":            symbol,
			"name":              etf.Name,
			"investment":        investment.InexactFloat64(),
			"dividend_yield":    dividendYield.InexactFloat64(),
			"expected_dividend": expectedDividend.InexactFloat64(),
		},
	})
}

// CalculateDividendByFrequency 按频率计算分红
func (h *ASharePortfolioHandler) CalculateDividendByFrequency(c *gin.Context) {
	frequency := c.Param("frequency") // monthly/quarterly/yearly

	// 从数据库获取ETF列表
	etfs := h.getOrCreateDefaultETFs()

	// 从数据库获取默认投资组合
	var portfolio models.AShareETFPortfolio
	models.DB.Where("is_default = ?", true).First(&portfolio)

	// 默认投资金额
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

	// 如果数据库中有配置，使用数据库中的值
	if portfolio.ID > 0 {
		var holdings []models.ASharePortfolioHolding
		models.DB.Where("portfolio_id = ?", portfolio.ID).Find(&holdings)
		for _, h := range holdings {
			for _, etf := range etfs {
				if etf.ID == h.ETFID {
					defaultInvestments[etf.Symbol] = h.Investment.InexactFloat64()
					break
				}
			}
		}
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
			"investment":      investment.InexactFloat64(),
			"period_dividend": periodDividend.InexactFloat64(),
			"annual_dividend": annualDividend.InexactFloat64(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
