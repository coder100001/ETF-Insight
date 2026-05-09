package handlers

import (
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// RiskBudgetHandler 风险预算处理器
type RiskBudgetHandler struct {
	riskBudgetService *services.RiskBudgetService
}

// NewRiskBudgetHandler 创建风险预算处理器
func NewRiskBudgetHandler() *RiskBudgetHandler {
	return &RiskBudgetHandler{
		riskBudgetService: services.NewRiskBudgetService(models.DB),
	}
}

// CreateRiskBudgetConfigRequest 创建风险预算配置请求
type CreateRiskBudgetConfigRequest struct {
	PortfolioID    uint    `json:"portfolio_id" binding:"required"`
	CVaRConfidence float64 `json:"cvar_confidence"`

	StockCVaRBudget     float64 `json:"stock_cvar_budget" binding:"required"`
	BondCVaRBudget      float64 `json:"bond_cvar_budget" binding:"required"`
	CommodityCVaRBudget float64 `json:"commodity_cvar_budget"`
	CashCVaRBudget      float64 `json:"cash_cvar_budget"`

	UseVaRConstraint bool    `json:"use_var_constraint"`
	StockVaRBudget   float64 `json:"stock_var_budget"`
	BondVaRBudget    float64 `json:"bond_var_budget"`

	MinSkewness   float64 `json:"min_skewness"`
	MaxDrawdown   float64 `json:"max_drawdown"`
	VaRConfidence float64 `json:"var_confidence"`
	EffectiveDate string  `json:"effective_date"`
}

// RiskBudgetConfigResponse 风险预算配置响应
type RiskBudgetConfigResponse struct {
	Success bool                     `json:"success"`
	Data    *models.RiskBudgetConfig `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// RiskBudgetConfigsResponse 配置列表响应
type RiskBudgetConfigsResponse struct {
	Success bool                      `json:"success"`
	Data    []models.RiskBudgetConfig `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// CalculateCVaRRequest 计算CVaR请求
type CalculateCVaRRequest struct {
	Returns       []float64 `json:"returns" binding:"required,min=10"`
	Confidence    float64   `json:"confidence" binding:"required"`
	UseParametric bool      `json:"use_parametric,omitempty"`
}

// CalculateCVaRResponse 计算CVaR响应
type CalculateCVaRResponse struct {
	Success bool        `json:"success"`
	Data    *CVaRResult `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CVaRResult CVaR计算结果
type CVaRResult struct {
	VaR        float64 `json:"var"`
	CVaR       float64 `json:"cvar"`
	Confidence float64 `json:"confidence"`
	Method     string  `json:"method"`
	SampleSize int     `json:"sample_size"`
}

// MonteCarloRequest 蒙特卡洛模拟请求
type MonteCarloRequest struct {
	PortfolioID    uint      `json:"portfolio_id" binding:"required"`
	NumSimulations int       `json:"num_simulations" binding:"required,min=100"`
	TimeSteps      int       `json:"time_steps" binding:"required,min=1"`
	Returns        []float64 `json:"returns" binding:"required,min=10"`
}

// MonteCarloResponse 蒙特卡洛模拟响应
type MonteCarloResponse struct {
	Success bool                         `json:"success"`
	Data    *models.MonteCarloSimulation `json:"data,omitempty"`
	Error   string                       `json:"error,omitempty"`
}

// CreateRiskBudgetConfig 创建风险预算配置
// @Summary 创建风险预算配置
// @Description 创建新的风险预算分析配置
// @Tags risk-budget
// @Accept json
// @Produce json
// @Param request body CreateRiskBudgetConfigRequest true "配置参数"
// @Success 200 {object} RiskBudgetConfigResponse
// @Router /api/risk-budget/configs [post]
func (h *RiskBudgetHandler) CreateRiskBudgetConfig(c *gin.Context) {
	var req CreateRiskBudgetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RiskBudgetConfigResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	cvarConfidence := req.CVaRConfidence
	if cvarConfidence == 0 {
		cvarConfidence = 0.95
	}

	var effectiveDate time.Time
	if req.EffectiveDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.EffectiveDate); err == nil {
			effectiveDate = parsed
		} else {
			effectiveDate = time.Now()
		}
	} else {
		effectiveDate = time.Now()
	}

	config := &models.RiskBudgetConfig{
		PortfolioID:         req.PortfolioID,
		CVaRConfidence:      decimal.NewFromFloat(cvarConfidence),
		StockCVaRBudget:     decimal.NewFromFloat(req.StockCVaRBudget),
		BondCVaRBudget:      decimal.NewFromFloat(req.BondCVaRBudget),
		CommodityCVaRBudget: decimal.NewFromFloat(req.CommodityCVaRBudget),
		CashCVaRBudget:      decimal.NewFromFloat(req.CashCVaRBudget),
		UseVaRConstraint:    req.UseVaRConstraint,
		StockVaRBudget:      decimal.NewFromFloat(req.StockVaRBudget),
		BondVaRBudget:       decimal.NewFromFloat(req.BondVaRBudget),
		MinSkewness:         decimal.NewFromFloat(req.MinSkewness),
		MaxDrawdown:         decimal.NewFromFloat(req.MaxDrawdown),
		VaRConfidence:       decimal.NewFromFloat(req.VaRConfidence),
		EffectiveDate:       effectiveDate,
	}

	if err := h.riskBudgetService.CreateConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, RiskBudgetConfigResponse{
			Success: false,
			Error:   "创建配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RiskBudgetConfigResponse{
		Success: true,
		Data:    config,
	})
}

// GetRiskBudgetConfigs 获取风险预算配置列表
// @Summary 获取风险预算配置列表
// @Description 获取所有风险预算分析配置
// @Tags risk-budget
// @Produce json
// @Success 200 {object} RiskBudgetConfigsResponse
// @Router /api/risk-budget/configs [get]
func (h *RiskBudgetHandler) GetRiskBudgetConfigs(c *gin.Context) {
	var configs []models.RiskBudgetConfig
	if err := models.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RiskBudgetConfigsResponse{
			Success: false,
			Error:   "查询配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RiskBudgetConfigsResponse{
		Success: true,
		Data:    configs,
	})
}

// CalculateCVaR 计算VaR和CVaR
// @Summary 计算VaR和CVaR
// @Description 基于历史收益率数据计算风险价值(VaR)和条件风险价值(CVaR)
// @Tags risk-budget
// @Accept json
// @Produce json
// @Param request body CalculateCVaRRequest true "计算参数"
// @Success 200 {object} CalculateCVaRResponse
// @Router /api/risk-budget/calculate-cvar [post]
func (h *RiskBudgetHandler) CalculateCVaR(c *gin.Context) {
	var req CalculateCVaRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CalculateCVaRResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	returns := make([]decimal.Decimal, len(req.Returns))
	for i, r := range req.Returns {
		returns[i] = decimal.NewFromFloat(r)
	}

	confidence := decimal.NewFromFloat(req.Confidence)
	varVaR, varCVaR, err := h.riskBudgetService.CalculateCVaR(returns, confidence, req.UseParametric)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CalculateCVaRResponse{
			Success: false,
			Error:   "计算失败: " + err.Error(),
		})
		return
	}

	method := "historical"
	if req.UseParametric {
		method = "parametric"
	}

	varVal, _ := varVaR.Float64()
	cvarVal, _ := varCVaR.Float64()

	c.JSON(http.StatusOK, CalculateCVaRResponse{
		Success: true,
		Data: &CVaRResult{
			VaR:        varVal,
			CVaR:       cvarVal,
			Confidence: req.Confidence,
			Method:     method,
			SampleSize: len(req.Returns),
		},
	})
}

// RunMonteCarlo 运行蒙特卡洛模拟
// @Summary 运行蒙特卡洛模拟
// @Description 使用蒙特卡洛方法模拟投资组合的未来表现
// @Tags risk-budget
// @Accept json
// @Produce json
// @Param request body MonteCarloRequest true "模拟参数"
// @Success 200 {object} MonteCarloResponse
// @Router /api/risk-budget/monte-carlo [post]
func (h *RiskBudgetHandler) RunMonteCarlo(c *gin.Context) {
	var req MonteCarloRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MonteCarloResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	returns := make([]decimal.Decimal, len(req.Returns))
	for i, r := range req.Returns {
		returns[i] = decimal.NewFromFloat(r)
	}

	result, err := h.riskBudgetService.RunMonteCarloSimulation(
		req.PortfolioID, req.NumSimulations, req.TimeSteps, returns,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MonteCarloResponse{
			Success: false,
			Error:   "模拟失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MonteCarloResponse{
		Success: true,
		Data:    result,
	})
}
