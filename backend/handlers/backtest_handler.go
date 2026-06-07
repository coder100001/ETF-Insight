package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/services/backtest"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	DefaultSMBFactorReturn = 0.02
	DefaultHMLFactorReturn = 0.03
)

// BacktestHandler 回测处理器
type BacktestHandler struct{}

// NewBacktestHandler 创建回测处理器
func NewBacktestHandler() *BacktestHandler {
	return &BacktestHandler{}
}

// BacktestRequest 回测请求
type BacktestRequest struct {
	InitialCapital float64        `json:"initial_capital" binding:"required,min=1000"`
	StartDate      string         `json:"start_date" binding:"required"`
	EndDate        string         `json:"end_date" binding:"required"`
	Symbols        []string       `json:"symbols" binding:"required,min=1"`
	StrategyType   string         `json:"strategy_type" binding:"required"`
	StrategyParams map[string]any `json:"strategy_params"`
	SlippageRate   float64        `json:"slippage_rate"`   // 滑点率
	CommissionRate float64        `json:"commission_rate"` // 手续费率
	DividendTax    float64        `json:"dividend_tax"`    // 股息税率
}

// BacktestResponse 回测响应
type BacktestResponse struct {
	Success bool                     `json:"success"`
	Data    *backtest.BacktestResult `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// RunBacktest 运行回测
// @Summary 运行策略回测
// @Description 使用指定策略对投资组合进行历史回测
// @Tags backtest
// @Accept json
// @Produce json
// @Param request body BacktestRequest true "回测参数"
// @Success 200 {object} BacktestResponse
// @Router /api/backtest/run [post]
func (h *BacktestHandler) RunBacktest(c *gin.Context) {
	var req BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BacktestResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, BacktestResponse{
			Success: false,
			Error:   "开始日期格式错误: " + err.Error(),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, BacktestResponse{
			Success: false,
			Error:   "结束日期格式错误: " + err.Error(),
		})
		return
	}

	// 创建策略
	strategy, err := h.createStrategy(req.StrategyType, req.StrategyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, BacktestResponse{
			Success: false,
			Error:   "创建策略失败: " + err.Error(),
		})
		return
	}

	// 创建回测引擎
	engine := backtest.NewBacktestEngine(req.InitialCapital, strategy)

	// 设置滑点模型
	if req.SlippageRate > 0 {
		engine.SetSlippageModel(&backtest.DefaultSlippageModel{
			SlippageRate: decimal.NewFromFloat(req.SlippageRate),
		})
	}

	// 设置手续费模型
	if req.CommissionRate > 0 {
		engine.SetCommissionModel(&backtest.DefaultCommissionModel{
			CommissionRate: decimal.NewFromFloat(req.CommissionRate),
		})
	}

	// 设置股息模型
	if req.DividendTax > 0 {
		engine.SetDividendModel(&backtest.DefaultDividendModel{
			TaxRate: decimal.NewFromFloat(req.DividendTax),
		})
	}

	// 设置数据提供者
	dataProvider := backtest.NewDBProvider(models.DB, req.Symbols)
	engine.SetDataProvider(dataProvider)

	// 运行回测
	result, err := engine.Run(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BacktestResponse{
			Success: false,
			Error:   "回测执行失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BacktestResponse{
		Success: true,
		Data:    result,
	})
}

// FactorAnalysisRequest 因子分析请求
type FactorAnalysisRequest struct {
	Symbols     []string `json:"symbols" binding:"required,min=1"`
	StartDate   string   `json:"start_date" binding:"required"`
	EndDate     string   `json:"end_date" binding:"required"`
	FactorTypes []string `json:"factor_types"` // 要计算的因子类型
}

// FactorAnalysisResponse 因子分析响应
type FactorAnalysisResponse struct {
	Success bool                  `json:"success"`
	Data    *FactorAnalysisResult `json:"data,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// FactorAnalysisResult 因子分析结果
type FactorAnalysisResult struct {
	FactorExposures   map[string]map[string]float64 `json:"factor_exposures"` // symbol -> factor -> exposure
	FactorReturns     map[string]float64            `json:"factor_returns"`   // factor -> return
	FactorDefinitions []backtest.FactorDefinition   `json:"factor_definitions"`
}

// AnalyzeFactors 因子分析
// @Summary 因子分析
// @Description 分析投资组合的因子暴露
// @Tags backtest
// @Accept json
// @Produce json
// @Param request body FactorAnalysisRequest true "因子分析参数"
// @Success 200 {object} FactorAnalysisResponse
// @Router /api/backtest/factors [post]
func (h *BacktestHandler) AnalyzeFactors(c *gin.Context) {
	var req FactorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FactorAnalysisResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, FactorAnalysisResponse{
			Success: false,
			Error:   "开始日期格式错误: " + err.Error(),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, FactorAnalysisResponse{
			Success: false,
			Error:   "结束日期格式错误: " + err.Error(),
		})
		return
	}

	// 创建因子库
	factorLib := backtest.NewFactorLibrary(nil)

	// 获取因子定义
	definitions := factorLib.GetFactorDefinitions()

	// 获取历史数据
	dataProvider := backtest.NewDBProvider(models.DB, req.Symbols)
	data, err := dataProvider.GetData(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, FactorAnalysisResponse{
			Success: false,
			Error:   "获取历史数据失败: " + err.Error(),
		})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, FactorAnalysisResponse{
			Success: false,
			Error:   "没有可用的历史数据",
		})
		return
	}

	// 计算因子暴露
	factorExposures := make(map[string]map[string]float64)
	factorReturns := make(map[string]float64)

	// 按标的分组数据
	dataBySymbol := make(map[string][]*backtest.Bar)
	for _, bar := range data {
		dataBySymbol[bar.Symbol] = append(dataBySymbol[bar.Symbol], bar)
	}

	// 计算每个标的的因子暴露
	for symbol, bars := range dataBySymbol {
		if len(bars) < 20 {
			continue // 数据不足，跳过
		}

		// 提取价格序列
		prices := make([]decimal.Decimal, len(bars))
		for i, bar := range bars {
			prices[i] = bar.Close
		}

		// 计算收益率
		returns := make([]decimal.Decimal, len(prices)-1)
		for i := 1; i < len(prices); i++ {
			returns[i-1] = prices[i].Sub(prices[i-1]).Div(prices[i-1])
		}

		// 计算动量因子
		momentum := factorLib.CalculateMomentumFactor(prices, 12*20)

		// 计算低波因子
		lowVol := factorLib.CalculateLowVolFactor(returns)

		// 存储因子暴露
		factorExposures[symbol] = map[string]float64{
			string(backtest.FactorMarket):   1.0, // 市场因子默认为1
			string(backtest.FactorMomentum): momentum.InexactFloat64(),
			string(backtest.FactorLowVol):   lowVol.InexactFloat64(),
		}
	}

	// 计算市场平均因子收益
	if len(factorExposures) > 0 {
		avgMomentum := 0.0
		avgLowVol := 0.0
		count := 0

		for _, exposures := range factorExposures {
			if val, ok := exposures[string(backtest.FactorMomentum)]; ok {
				avgMomentum += val
			}
			if val, ok := exposures[string(backtest.FactorLowVol)]; ok {
				avgLowVol += val
			}
			count++
		}

		if count > 0 {
			// 计算市场因子收益（基于数据期间的实际收益）
			// 使用所有标的的等权平均收益率作为市场代理
			marketReturn := 0.0
			validSymbolCount := 0
			if len(dataBySymbol) > 0 {
				for _, bars := range dataBySymbol {
					if len(bars) >= 2 {
						firstPrice := bars[0].Close
						lastPrice := bars[len(bars)-1].Close
						symbolReturn := lastPrice.Sub(firstPrice).Div(firstPrice).InexactFloat64()
						marketReturn += symbolReturn
						validSymbolCount++
					}
				}
				if validSymbolCount > 0 {
					marketReturn /= float64(validSymbolCount)
				}
			}

			// 年化市场收益
			days := endDate.Sub(startDate).Hours() / 24
			annualizedMarketReturn := 0.0
			if days > 0 {
				annualizedMarketReturn = math.Pow(1+marketReturn, float64(config.GetFinancialConfig().TradingDaysYear)/days) - 1
			}

			// 无风险利率
			riskFreeRate := config.GetFinancialConfig().RiskFreeRate

			factorReturns[string(backtest.FactorMarket)] = annualizedMarketReturn - riskFreeRate
			factorReturns[string(backtest.FactorSMB)] = DefaultSMBFactorReturn
			factorReturns[string(backtest.FactorHML)] = DefaultHMLFactorReturn
			factorReturns[string(backtest.FactorMomentum)] = avgMomentum / float64(count)
			factorReturns[string(backtest.FactorLowVol)] = avgLowVol / float64(count)
		}
	}

	c.JSON(http.StatusOK, FactorAnalysisResponse{
		Success: true,
		Data: &FactorAnalysisResult{
			FactorExposures:   factorExposures,
			FactorReturns:     factorReturns,
			FactorDefinitions: definitions,
		},
	})
}

// StrategyListResponse 策略列表响应
type StrategyListResponse struct {
	Success bool           `json:"success"`
	Data    []StrategyInfo `json:"data,omitempty"`
}

// StrategyInfo 策略信息
type StrategyInfo struct {
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Params      []ParamInfo `json:"params"`
}

// ParamInfo 参数信息
type ParamInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     any    `json:"default"`
	Description string `json:"description"`
}

// ListStrategies 列出可用策略
// @Summary 列出可用策略
// @Description 获取所有可用的回测策略列表
// @Tags backtest
// @Produce json
// @Success 200 {object} StrategyListResponse
// @Router /api/backtest/strategies [get]
func (h *BacktestHandler) ListStrategies(c *gin.Context) {
	strategies := []StrategyInfo{
		{
			Type:        "ma_cross",
			Name:        "均线交叉策略",
			Description: "基于短期和长期移动平均线交叉生成买卖信号",
			Params: []ParamInfo{
				{Name: "short_period", Type: "int", Default: 5, Description: "短期均线周期"},
				{Name: "long_period", Type: "int", Default: 20, Description: "长期均线周期"},
			},
		},
		{
			Type:        "rsi",
			Name:        "RSI策略",
			Description: "基于RSI超买超卖信号生成交易信号",
			Params: []ParamInfo{
				{Name: "period", Type: "int", Default: 14, Description: "RSI计算周期"},
				{Name: "oversold", Type: "float", Default: 30, Description: "超卖阈值"},
				{Name: "overbought", Type: "float", Default: 70, Description: "超买阈值"},
			},
		},
		{
			Type:        "momentum",
			Name:        "动量策略",
			Description: "基于价格动量选择强势标的",
			Params: []ParamInfo{
				{Name: "lookback_period", Type: "int", Default: 60, Description: "动量计算周期"},
				{Name: "top_n", Type: "int", Default: 5, Description: "选择前N个标的"},
				{Name: "rebalance_freq", Type: "int", Default: 20, Description: "调仓频率(天)"},
			},
		},
		{
			Type:        "factor",
			Name:        "因子策略",
			Description: "基于多因子模型的量化策略",
			Params: []ParamInfo{
				{Name: "momentum_weight", Type: "float", Default: 0.3, Description: "动量因子权重"},
				{Name: "value_weight", Type: "float", Default: 0.3, Description: "价值因子权重"},
				{Name: "lowvol_weight", Type: "float", Default: 0.2, Description: "低波因子权重"},
				{Name: "quality_weight", Type: "float", Default: 0.2, Description: "质量因子权重"},
			},
		},
		{
			Type:        "buy_hold",
			Name:        "买入持有策略",
			Description: "简单的买入持有策略",
			Params:      []ParamInfo{},
		},
	}

	c.JSON(http.StatusOK, StrategyListResponse{
		Success: true,
		Data:    strategies,
	})
}

// createStrategy 创建策略
func (h *BacktestHandler) createStrategy(strategyType string, params map[string]any) (backtest.Strategy, error) {
	switch strategyType {
	case "ma_cross":
		shortPeriod := getIntParam(params, "short_period", 5)
		longPeriod := getIntParam(params, "long_period", 20)
		return backtest.NewMovingAverageCrossStrategy(shortPeriod, longPeriod), nil

	case "rsi":
		period := getIntParam(params, "period", 14)
		oversold := getFloatParam(params, "oversold", 30)
		overbought := getFloatParam(params, "overbought", 70)
		return backtest.NewRSIStrategy(period, oversold, overbought), nil

	case "momentum":
		lookback := getIntParam(params, "lookback_period", 60)
		topN := getIntParam(params, "top_n", 5)
		rebalance := getIntParam(params, "rebalance_freq", 20)
		return backtest.NewMomentumStrategy(lookback, topN, rebalance), nil

	case "factor":
		weights := make(map[backtest.FactorType]float64)
		weights[backtest.FactorMomentum] = getFloatParam(params, "momentum_weight", 0.3)
		weights[backtest.FactorValue] = getFloatParam(params, "value_weight", 0.3)
		weights[backtest.FactorLowVol] = getFloatParam(params, "lowvol_weight", 0.2)
		weights[backtest.FactorQuality] = getFloatParam(params, "quality_weight", 0.2)
		return backtest.NewFactorBasedStrategy(weights), nil

	case "buy_hold":
		return backtest.NewBuyAndHoldStrategy(), nil

	default:
		return nil, fmt.Errorf("未知的策略类型: %s", strategyType)
	}
}

// 辅助函数

func getIntParam(params map[string]any, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

func getFloatParam(params map[string]any, key string, defaultValue float64) float64 {
	if val, ok := params[key]; ok {
		if v, ok := val.(float64); ok {
			return v
		}
	}
	return defaultValue
}

// EventDrivenBacktestRequest 事件驱动回测请求
type EventDrivenBacktestRequest struct {
	InitialCapital    float64        `json:"initial_capital" binding:"required,min=1000"`
	StartDate         string         `json:"start_date" binding:"required"`
	EndDate           string         `json:"end_date" binding:"required"`
	Symbols           []string       `json:"symbols" binding:"required,min=1"`
	StrategyType      string         `json:"strategy_type" binding:"required"`
	StrategyParams    map[string]any `json:"strategy_params"`
	SlippageRate      float64        `json:"slippage_rate"`       // 滑点率
	CommissionRate    float64        `json:"commission_rate"`     // 手续费率
	DividendTax       float64        `json:"dividend_tax"`        // 股息税率
	StopLossEnabled   bool           `json:"stop_loss_enabled"`   // 是否启用止损
	StopLossPercent   float64        `json:"stop_loss_percent"`   // 止损百分比
	TakeProfitEnabled bool           `json:"take_profit_enabled"` // 是否启用止盈
	TakeProfitPercent float64        `json:"take_profit_percent"` // 止盈百分比
	RebalanceEnabled  bool           `json:"rebalance_enabled"`   // 是否启用再平衡
	RebalanceInterval int            `json:"rebalance_interval"`  // 再平衡间隔天数
}

// EventDrivenBacktestResponse 事件驱动回测响应
type EventDrivenBacktestResponse struct {
	Success bool                                `json:"success"`
	Data    *backtest.EventDrivenBacktestResult `json:"data,omitempty"`
	Error   string                              `json:"error,omitempty"`
}

// RunEventDrivenBacktest 运行事件驱动回测
// @Summary 运行事件驱动回测
// @Description 使用事件驱动架构运行策略回测，支持止损、止盈、再平衡等高级功能
// @Tags backtest
// @Accept json
// @Produce json
// @Param request body EventDrivenBacktestRequest true "回测参数"
// @Success 200 {object} EventDrivenBacktestResponse
// @Router /api/backtest/event-driven [post]
func (h *BacktestHandler) RunEventDrivenBacktest(c *gin.Context) {
	var req EventDrivenBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, EventDrivenBacktestResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, EventDrivenBacktestResponse{
			Success: false,
			Error:   "开始日期格式错误: " + err.Error(),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, EventDrivenBacktestResponse{
			Success: false,
			Error:   "结束日期格式错误: " + err.Error(),
		})
		return
	}

	// 创建策略
	strategy, err := h.createStrategy(req.StrategyType, req.StrategyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, EventDrivenBacktestResponse{
			Success: false,
			Error:   "创建策略失败: " + err.Error(),
		})
		return
	}

	// 创建事件驱动回测引擎
	engine := backtest.NewEventDrivenEngine(req.InitialCapital, strategy)

	// 设置滑点模型
	if req.SlippageRate > 0 {
		engine.SetSlippageModel(&backtest.DefaultSlippageModel{
			SlippageRate: decimal.NewFromFloat(req.SlippageRate),
		})
	}

	// 设置手续费模型
	if req.CommissionRate > 0 {
		engine.SetCommissionModel(&backtest.DefaultCommissionModel{
			CommissionRate: decimal.NewFromFloat(req.CommissionRate),
		})
	}

	// 设置股息模型
	if req.DividendTax > 0 {
		engine.SetDividendModel(&backtest.DefaultDividendModel{
			TaxRate: decimal.NewFromFloat(req.DividendTax),
		})
	}

	// 设置止损止盈
	engine.SetStopLoss(req.StopLossEnabled, req.StopLossPercent)
	engine.SetTakeProfit(req.TakeProfitEnabled, req.TakeProfitPercent)

	// 设置再平衡
	engine.SetRebalance(req.RebalanceEnabled, req.RebalanceInterval)

	// 设置数据提供者
	dataProvider := backtest.NewDBProvider(models.DB, req.Symbols)
	engine.SetDataProvider(dataProvider)

	// 运行回测
	result, err := engine.Run(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, EventDrivenBacktestResponse{
			Success: false,
			Error:   "回测执行失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, EventDrivenBacktestResponse{
		Success: true,
		Data:    result,
	})
}
