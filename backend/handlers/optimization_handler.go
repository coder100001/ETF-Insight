package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/services/optimization"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// OptimizationHandler 组合优化处理器
type OptimizationHandler struct {
	mptOptimizer            *optimization.MPTOptimizer
	riskParityOptimizer     *optimization.RiskParityOptimizer
	blackLittermanOptimizer *optimization.BlackLittermanOptimizer
}

// NewOptimizationHandler 创建优化处理器
func NewOptimizationHandler() *OptimizationHandler {
	return &OptimizationHandler{
		mptOptimizer:            optimization.NewMPTOptimizer(),
		riskParityOptimizer:     optimization.NewRiskParityOptimizer(),
		blackLittermanOptimizer: optimization.NewBlackLittermanOptimizer(),
	}
}

// MPTOptimizeRequest 均值-方差优化请求
type MPTOptimizeRequest struct {
	Symbols      []string                      `json:"symbols" binding:"required,min=2"`
	Returns      map[string]float64            `json:"returns,omitempty"`
	CovMatrix    map[string]map[string]float64 `json:"cov_matrix,omitempty"`
	Constraints  *ConstraintConfig             `json:"constraints,omitempty"`
	Objective    string                        `json:"objective" binding:"required,oneof=min_volatility max_sharpe target_return"`
	TargetReturn float64                       `json:"target_return,omitempty"`
	RiskFreeRate float64                       `json:"risk_free_rate,omitempty"`
}

// ConstraintConfig 约束配置
type ConstraintConfig struct {
	MinWeights     map[string]float64 `json:"min_weights,omitempty"`
	MaxWeights     map[string]float64 `json:"max_weights,omitempty"`
	AllowShort     bool               `json:"allow_short,omitempty"`
	MaxShortWeight float64            `json:"max_short_weight,omitempty"`
}

// MPTOptimizeResponse 优化响应
type MPTOptimizeResponse struct {
	Success bool                          `json:"success"`
	Data    *optimization.PortfolioResult `json:"data,omitempty"`
	Error   string                        `json:"error,omitempty"`
}

// EfficientFrontierRequest 有效前沿请求
type EfficientFrontierRequest struct {
	Symbols     []string                      `json:"symbols" binding:"required,min=2"`
	Returns     map[string]float64            `json:"returns,omitempty"`
	CovMatrix   map[string]map[string]float64 `json:"cov_matrix,omitempty"`
	Constraints *ConstraintConfig             `json:"constraints,omitempty"`
	NumPoints   int                           `json:"num_points,omitempty"`
}

// EfficientFrontierResponse 有效前沿响应
type EfficientFrontierResponse struct {
	Success bool                                   `json:"success"`
	Data    []*optimization.EfficientFrontierPoint `json:"data,omitempty"`
	Error   string                                 `json:"error,omitempty"`
}

// MPTOptimize 均值-方差优化
// @Summary 执行均值-方差优化
// @Description 基于现代投资组合理论(MPT)优化资产配置
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body MPTOptimizeRequest true "优化参数"
// @Success 200 {object} MPTOptimizeResponse
// @Router /api/optimization/mpt [post]
func (h *OptimizationHandler) MPTOptimize(c *gin.Context) {
	var req MPTOptimizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MPTOptimizeResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 如果未提供 Returns 和 CovMatrix，自动从历史数据计算
	if req.Returns == nil || req.CovMatrix == nil {
		returns, covMatrix, err := h.calculateReturnsAndCovMatrix(req.Symbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MPTOptimizeResponse{
				Success: false,
				Error:   "无法获取历史数据: " + err.Error(),
			})
			return
		}
		if req.Returns == nil {
			req.Returns = returns
		}
		if req.CovMatrix == nil {
			req.CovMatrix = covMatrix
		}
	}

	// 设置无风险利率
	if req.RiskFreeRate > 0 {
		h.mptOptimizer.SetRiskFreeRate(req.RiskFreeRate)
	}

	// 构建约束条件
	constraint := optimization.NewConstraint(req.Symbols)
	if req.Constraints != nil {
		if req.Constraints.MinWeights != nil {
			for symbol, weight := range req.Constraints.MinWeights {
				constraint.SetMinWeight(symbol, weight)
			}
		}
		if req.Constraints.MaxWeights != nil {
			for symbol, weight := range req.Constraints.MaxWeights {
				constraint.SetMaxWeight(symbol, weight)
			}
		}
		constraint.AllowShort = req.Constraints.AllowShort
		constraint.MaxShortWeight = req.Constraints.MaxShortWeight
	}

	// 执行优化
	var result *optimization.PortfolioResult
	var err error

	switch req.Objective {
	case "min_volatility":
		result, err = h.mptOptimizer.OptimizeMinVolatility(req.Returns, req.CovMatrix, constraint)
	case "max_sharpe":
		result, err = h.mptOptimizer.OptimizeMaxSharpe(req.Returns, req.CovMatrix, constraint)
	case "target_return":
		if req.TargetReturn == 0 {
			c.JSON(http.StatusBadRequest, MPTOptimizeResponse{
				Success: false,
				Error:   "目标收益率模式下必须提供 target_return 参数",
			})
			return
		}
		result, err = h.mptOptimizer.OptimizeForTargetReturn(req.Returns, req.CovMatrix, constraint, req.TargetReturn)
	default:
		result, err = h.mptOptimizer.Optimize(req.Returns, req.CovMatrix, constraint)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, MPTOptimizeResponse{
			Success: false,
			Error:   "优化失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MPTOptimizeResponse{
		Success: true,
		Data:    result,
	})
}

// EfficientFrontier 计算有效前沿
// @Summary 计算有效前沿
// @Description 生成投资组合的有效前沿曲线
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body EfficientFrontierRequest true "有效前沿参数"
// @Success 200 {object} EfficientFrontierResponse
// @Router /api/optimization/efficient-frontier [post]
func (h *OptimizationHandler) EfficientFrontier(c *gin.Context) {
	var req EfficientFrontierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, EfficientFrontierResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认点数
	if req.NumPoints <= 0 {
		req.NumPoints = 20
	}
	if req.NumPoints > 100 {
		req.NumPoints = 100
	}

	// 如果未提供 Returns 和 CovMatrix，自动从历史数据计算
	if req.Returns == nil || req.CovMatrix == nil {
		returns, covMatrix, err := h.calculateReturnsAndCovMatrix(req.Symbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, EfficientFrontierResponse{
				Success: false,
				Error:   "无法获取历史数据: " + err.Error(),
			})
			return
		}
		if req.Returns == nil {
			req.Returns = returns
		}
		if req.CovMatrix == nil {
			req.CovMatrix = covMatrix
		}
	}

	// 构建约束条件
	constraint := optimization.NewConstraint(req.Symbols)
	if req.Constraints != nil {
		if req.Constraints.MinWeights != nil {
			for symbol, weight := range req.Constraints.MinWeights {
				constraint.SetMinWeight(symbol, weight)
			}
		}
		if req.Constraints.MaxWeights != nil {
			for symbol, weight := range req.Constraints.MaxWeights {
				constraint.SetMaxWeight(symbol, weight)
			}
		}
	}

	// 计算有效前沿
	frontier, err := h.mptOptimizer.CalculateEfficientFrontier(req.Returns, req.CovMatrix, constraint, req.NumPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, EfficientFrontierResponse{
			Success: false,
			Error:   "计算失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, EfficientFrontierResponse{
		Success: true,
		Data:    frontier,
	})
}

// CalculateCovarianceMatrix 计算协方差矩阵
// @Summary 计算协方差矩阵
// @Description 基于历史收益率数据计算协方差矩阵
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body CovarianceRequest true "协方差计算参数"
// @Success 200 {object} CovarianceResponse
// @Router /api/optimization/covariance [post]
func (h *OptimizationHandler) CalculateCovarianceMatrix(c *gin.Context) {
	var req CovarianceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CovarianceResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 计算协方差矩阵
	covMatrix := calculateCovarianceMatrix(req.Returns)

	c.JSON(http.StatusOK, CovarianceResponse{
		Success: true,
		Data:    covMatrix,
	})
}

// CovarianceRequest 协方差计算请求
type CovarianceRequest struct {
	Returns map[string][]float64 `json:"returns" binding:"required"` // symbol -> returns array
}

// CovarianceResponse 协方差计算响应
type CovarianceResponse struct {
	Success bool                          `json:"success"`
	Data    map[string]map[string]float64 `json:"data,omitempty"`
	Error   string                        `json:"error,omitempty"`
}

// calculateCovarianceMatrix 计算协方差矩阵
func calculateCovarianceMatrix(returns map[string][]float64) map[string]map[string]float64 {
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}

	n := len(symbols)
	if n == 0 {
		return nil
	}

	// 计算均值
	means := make(map[string]float64)
	for _, symbol := range symbols {
		data := returns[symbol]
		if len(data) == 0 {
			continue
		}
		sum := 0.0
		for _, r := range data {
			sum += r
		}
		means[symbol] = sum / float64(len(data))
	}

	// 计算协方差
	covMatrix := make(map[string]map[string]float64)
	for _, s1 := range symbols {
		covMatrix[s1] = make(map[string]float64)
		for _, s2 := range symbols {
			data1 := returns[s1]
			data2 := returns[s2]

			minLen := len(data1)
			if len(data2) < minLen {
				minLen = len(data2)
			}

			if minLen == 0 {
				covMatrix[s1][s2] = 0
				continue
			}

			cov := 0.0
			for i := 0; i < minLen; i++ {
				cov += (data1[i] - means[s1]) * (data2[i] - means[s2])
			}

			covMatrix[s1][s2] = cov / float64(minLen)
		}
	}

	return covMatrix
}

// ==================== 风险平价优化 API ====================

// RiskParityRequest 风险平价优化请求
type RiskParityRequest struct {
	Symbols     []string                      `json:"symbols" binding:"required,min=2"`
	Returns     map[string]float64            `json:"returns"`    // 各资产预期收益率（可选，不提供则使用示例数据）
	CovMatrix   map[string]map[string]float64 `json:"cov_matrix"` // 协方差矩阵（可选，不提供则使用示例数据）
	Constraints *RiskParityConstraintConfig   `json:"constraints,omitempty"`
	Method      string                        `json:"method" binding:"omitempty,oneof=parity inverse_vol budget"`
	RiskBudget  map[string]float64            `json:"risk_budget,omitempty"`
}

// RiskParityConstraintConfig 风险平价约束配置
type RiskParityConstraintConfig struct {
	MinWeights       map[string]float64 `json:"min_weights,omitempty"`
	MaxWeights       map[string]float64 `json:"max_weights,omitempty"`
	TargetVolatility float64            `json:"target_volatility,omitempty"`
	UseLeverage      bool               `json:"use_leverage,omitempty"`
	MaxLeverage      float64            `json:"max_leverage,omitempty"`
}

// RiskParityResponse 风险平价优化响应
type RiskParityResponse struct {
	Success bool                           `json:"success"`
	Data    *optimization.RiskParityResult `json:"data,omitempty"`
	Error   string                         `json:"error,omitempty"`
}

// RiskParityOptimize 风险平价优化
// @Summary 执行风险平价优化
// @Description 基于风险贡献的资产配置，使各资产对组合风险的贡献相等
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body RiskParityRequest true "优化参数"
// @Success 200 {object} RiskParityResponse
// @Router /api/optimization/risk-parity [post]
func (h *OptimizationHandler) RiskParityOptimize(c *gin.Context) {
	var req RiskParityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RiskParityResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 如果没有提供收益率和协方差矩阵，生成示例数据
	returns := req.Returns
	covMatrix := req.CovMatrix

	if returns == nil || covMatrix == nil {
		returns, covMatrix = generateSampleRiskParityData(req.Symbols)
	}

	// 构建约束条件
	constraint := optimization.NewRiskParityConstraint(req.Symbols)
	if req.Constraints != nil {
		if req.Constraints.MinWeights != nil {
			for symbol, weight := range req.Constraints.MinWeights {
				constraint.MinWeight[symbol] = weight
			}
		}
		if req.Constraints.MaxWeights != nil {
			for symbol, weight := range req.Constraints.MaxWeights {
				constraint.MaxWeight[symbol] = weight
			}
		}
		constraint.TargetVolatility = req.Constraints.TargetVolatility
		constraint.UseLeverage = req.Constraints.UseLeverage
		constraint.MaxLeverage = req.Constraints.MaxLeverage
	}

	// 执行优化
	var result *optimization.RiskParityResult
	var err error

	switch req.Method {
	case "inverse_vol":
		result, err = h.riskParityOptimizer.OptimizeInverseVol(returns, covMatrix, constraint)
	case "budget":
		if len(req.RiskBudget) == 0 {
			c.JSON(http.StatusBadRequest, RiskParityResponse{
				Success: false,
				Error:   "风险预算方法需要提供 risk_budget 参数",
			})
			return
		}
		result, err = h.riskParityOptimizer.CalculateRiskBudget(returns, covMatrix, req.RiskBudget, constraint)
	default:
		result, err = h.riskParityOptimizer.Optimize(returns, covMatrix, constraint)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, RiskParityResponse{
			Success: false,
			Error:   "优化失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RiskParityResponse{
		Success: true,
		Data:    result,
	})
}

// generateSampleRiskParityData 生成示例风险平价数据
func generateSampleRiskParityData(symbols []string) (map[string]float64, map[string]map[string]float64) {
	returns := make(map[string]float64)
	covMatrix := make(map[string]map[string]float64)

	// 为每个资产生成示例预期收益率和波动率
	// 在实际应用中，这些数据应该从数据库或市场数据API获取
	sampleData := map[string]struct {
		expectedReturn float64
		volatility     float64
	}{
		"SPY": {expectedReturn: 0.08, volatility: 0.20}, // 标普500: 8%收益, 20%波动率
		"TLT": {expectedReturn: 0.04, volatility: 0.15}, // 长期国债: 4%收益, 15%波动率
		"GLD": {expectedReturn: 0.05, volatility: 0.18}, // 黄金: 5%收益, 18%波动率
		"VNQ": {expectedReturn: 0.06, volatility: 0.22}, // 房地产: 6%收益, 22%波动率
		"EFA": {expectedReturn: 0.07, volatility: 0.21}, // 发达市场: 7%收益, 21%波动率
		"EEM": {expectedReturn: 0.09, volatility: 0.28}, // 新兴市场: 9%收益, 28%波动率
		"AGG": {expectedReturn: 0.03, volatility: 0.05}, // 综合债券: 3%收益, 5%波动率
		"LQD": {expectedReturn: 0.04, volatility: 0.08}, // 公司债: 4%收益, 8%波动率
		"HYG": {expectedReturn: 0.05, volatility: 0.12}, // 高收益债: 5%收益, 12%波动率
		"DBC": {expectedReturn: 0.03, volatility: 0.16}, // 大宗商品: 3%收益, 16%波动率
	}

	// 生成收益率
	for _, symbol := range symbols {
		if data, exists := sampleData[symbol]; exists {
			returns[symbol] = data.expectedReturn
		} else {
			// 对于未知资产，使用默认值
			returns[symbol] = 0.06
		}
	}

	// 生成协方差矩阵
	for i, symbol1 := range symbols {
		covMatrix[symbol1] = make(map[string]float64)
		vol1 := 0.20 // 默认波动率
		if data, exists := sampleData[symbol1]; exists {
			vol1 = data.volatility
		}

		for j, symbol2 := range symbols {
			vol2 := 0.20 // 默认波动率
			if data, exists := sampleData[symbol2]; exists {
				vol2 = data.volatility
			}

			if i == j {
				// 对角线元素：方差
				covMatrix[symbol1][symbol2] = vol1 * vol1
			} else {
				// 非对角线元素：协方差（假设相关系数为0.3）
				correlation := 0.3
				covMatrix[symbol1][symbol2] = correlation * vol1 * vol2
			}
		}
	}

	return returns, covMatrix
}

// ==================== Black-Litterman 优化 API ====================

// BlackLittermanRequest Black-Litterman优化请求
type BlackLittermanRequest struct {
	Symbols       []string                        `json:"symbols,omitempty"`        // 资产列表（可选，用于生成示例数据）
	MarketWeights map[string]float64              `json:"market_weights"`           // 市场均衡权重（可选，不提供则使用等权重）
	CovMatrix     map[string]map[string]float64   `json:"cov_matrix"`               // 协方差矩阵（可选，不提供则使用示例数据）
	Views         []BLViewInput                   `json:"views,omitempty"`          // 投资者观点（前端格式）
	AbsoluteViews map[string]float64              `json:"absolute_views,omitempty"` // 绝对观点（直接指定资产收益）
	RelativeViews []*RelativeViewConfig           `json:"relative_views,omitempty"` // 相对观点
	RiskAversion  float64                         `json:"risk_aversion,omitempty"`
	Tau           float64                         `json:"tau,omitempty"`
	RiskFreeRate  float64                         `json:"risk_free_rate,omitempty"`
	Constraints   *BlackLittermanConstraintConfig `json:"constraints,omitempty"`
}

// BLViewInput 前端传入的观点格式
type BLViewInput struct {
	Symbol     string  `json:"symbol" binding:"required"`
	Return     float64 `json:"return" binding:"required"`
	Confidence float64 `json:"confidence,omitempty"`
}

// RelativeViewConfig 相对观点配置
type RelativeViewConfig struct {
	Asset1       string  `json:"asset1" binding:"required"`
	Asset2       string  `json:"asset2" binding:"required"`
	ExpectedDiff float64 `json:"expected_diff" binding:"required"`
	Confidence   float64 `json:"confidence,omitempty"`
}

// BlackLittermanConstraintConfig Black-Litterman约束配置
type BlackLittermanConstraintConfig struct {
	MinWeights map[string]float64 `json:"min_weights,omitempty"`
	MaxWeights map[string]float64 `json:"max_weights,omitempty"`
}

// BlackLittermanResponse Black-Litterman优化响应
type BlackLittermanResponse struct {
	Success bool                               `json:"success"`
	Data    *optimization.BlackLittermanResult `json:"data,omitempty"`
	Error   string                             `json:"error,omitempty"`
}

// BlackLittermanOptimize Black-Litterman优化
// @Summary 执行Black-Litterman优化
// @Description 融合市场均衡收益和投资者观点，生成后验收益估计并优化
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body BlackLittermanRequest true "优化参数"
// @Success 200 {object} BlackLittermanResponse
// @Router /api/optimization/black-litterman [post]
func (h *OptimizationHandler) BlackLittermanOptimize(c *gin.Context) {
	var req BlackLittermanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BlackLittermanResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 如果没有提供市场权重和协方差矩阵，生成示例数据
	marketWeights := req.MarketWeights
	covMatrix := req.CovMatrix
	symbols := req.Symbols

	if len(symbols) == 0 {
		// 从市场权重中提取资产列表
		symbols = make([]string, 0, len(marketWeights))
		for symbol := range marketWeights {
			symbols = append(symbols, symbol)
		}
	}

	if len(symbols) == 0 {
		c.JSON(http.StatusBadRequest, BlackLittermanResponse{
			Success: false,
			Error:   "请提供资产列表(symbols)或市场权重(market_weights)",
		})
		return
	}

	// 生成示例市场权重（等权重）
	if marketWeights == nil {
		marketWeights = make(map[string]float64)
		equalWeight := 1.0 / float64(len(symbols))
		for _, symbol := range symbols {
			marketWeights[symbol] = equalWeight
		}
	}

	// 生成示例协方差矩阵
	if covMatrix == nil {
		covMatrix = generateSampleCovMatrix(symbols)
	}

	// 转换前端观点格式为绝对观点
	absoluteViews := req.AbsoluteViews
	if absoluteViews == nil {
		absoluteViews = make(map[string]float64)
	}

	// 将前端 views 转换为 absolute_views
	for _, view := range req.Views {
		absoluteViews[view.Symbol] = view.Return
	}

	// 设置参数
	if req.Tau > 0 {
		h.blackLittermanOptimizer.SetTau(req.Tau)
	}
	if req.RiskFreeRate > 0 {
		h.blackLittermanOptimizer.SetRiskFreeRate(req.RiskFreeRate)
	}

	// 构建约束条件
	constraint := optimization.NewBlackLittermanConstraint(symbols)
	if req.Constraints != nil {
		if req.Constraints.MinWeights != nil {
			for symbol, weight := range req.Constraints.MinWeights {
				constraint.MinWeight[symbol] = weight
			}
		}
		if req.Constraints.MaxWeights != nil {
			for symbol, weight := range req.Constraints.MaxWeights {
				constraint.MaxWeight[symbol] = weight
			}
		}
	}

	// 转换相对观点
	relativeViews := make([]*optimization.RelativeView, 0)
	for _, rv := range req.RelativeViews {
		confidence := rv.Confidence
		if confidence == 0 {
			confidence = 0.5
		}
		relativeViews = append(relativeViews, &optimization.RelativeView{
			Asset1:       rv.Asset1,
			Asset2:       rv.Asset2,
			ExpectedDiff: rv.ExpectedDiff,
			Confidence:   confidence,
		})
	}

	// 执行优化
	result, err := h.blackLittermanOptimizer.OptimizeWithViews(
		marketWeights,
		covMatrix,
		absoluteViews,
		relativeViews,
		constraint,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, BlackLittermanResponse{
			Success: false,
			Error:   "优化失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BlackLittermanResponse{
		Success: true,
		Data:    result,
	})
}

// generateSampleCovMatrix 生成示例协方差矩阵
func generateSampleCovMatrix(symbols []string) map[string]map[string]float64 {
	covMatrix := make(map[string]map[string]float64)

	// 为常见资产提供示例波动率
	volatilities := map[string]float64{
		"SPY": 0.20, "VTI": 0.20, "VOO": 0.20, "QQQ": 0.25,
		"TLT": 0.15, "AGG": 0.05, "BND": 0.05, "LQD": 0.08,
		"GLD": 0.18, "SLV": 0.25, "USO": 0.30, "DBA": 0.20,
		"EFA": 0.21, "EEM": 0.28, "IWM": 0.24, "VUG": 0.22,
		"VTV": 0.18, "VIG": 0.17, "VYM": 0.18, "VNQ": 0.22,
	}

	for i, symbol1 := range symbols {
		covMatrix[symbol1] = make(map[string]float64)
		vol1 := 0.20 // 默认波动率
		if v, exists := volatilities[symbol1]; exists {
			vol1 = v
		}

		for j, symbol2 := range symbols {
			vol2 := 0.20 // 默认波动率
			if v, exists := volatilities[symbol2]; exists {
				vol2 = v
			}

			if i == j {
				// 对角线元素：方差
				covMatrix[symbol1][symbol2] = vol1 * vol1
			} else {
				// 非对角线元素：协方差（假设相关系数为0.3）
				correlation := 0.3
				covMatrix[symbol1][symbol2] = correlation * vol1 * vol2
			}
		}
	}

	return covMatrix
}

// MarketImpliedReturnsRequest 市场隐含收益请求
type MarketImpliedReturnsRequest struct {
	MarketWeights map[string]float64            `json:"market_weights" binding:"required"`
	CovMatrix     map[string]map[string]float64 `json:"cov_matrix" binding:"required"`
	RiskAversion  float64                       `json:"risk_aversion,omitempty"`
}

// MarketImpliedReturnsResponse 市场隐含收益响应
type MarketImpliedReturnsResponse struct {
	Success bool               `json:"success"`
	Data    map[string]float64 `json:"data,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// MarketImpliedReturns 计算市场隐含收益
// @Summary 计算市场隐含收益（反向优化）
// @Description 基于市场权重和协方差矩阵，反推出市场隐含的收益预期
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body MarketImpliedReturnsRequest true "计算参数"
// @Success 200 {object} MarketImpliedReturnsResponse
// @Router /api/optimization/market-implied-returns [post]
func (h *OptimizationHandler) MarketImpliedReturns(c *gin.Context) {
	var req MarketImpliedReturnsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MarketImpliedReturnsResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	riskAversion := req.RiskAversion
	if riskAversion == 0 {
		riskAversion = 2.5
	}

	result := h.blackLittermanOptimizer.CalculateMarketImpliedReturns(
		req.MarketWeights,
		req.CovMatrix,
		riskAversion,
	)

	c.JSON(http.StatusOK, MarketImpliedReturnsResponse{
		Success: true,
		Data:    result,
	})
}

// ==================== ETF历史数据统计 API ====================

// ETFStatistics ETF统计数据
type ETFStatistics struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Annualized  float64 `json:"annualized"`   // 年化收益率
	Volatility  float64 `json:"volatility"`   // 年化波动率
	Sharpe      float64 `json:"sharpe"`       // 夏普比率
	MaxDrawdown float64 `json:"max_drawdown"` // 最大回撤
	DataPoints  int     `json:"data_points"`  // 数据点数
}

// GetETFStatisticsRequest 获取ETF统计请求
type GetETFStatisticsRequest struct {
	Symbols []string `json:"symbols" binding:"required,min=1"`
	Period  string   `json:"period,omitempty"`   // 数据周期: 1y, 3y, 5y, 10y
	EndDate string   `json:"end_date,omitempty"` // 结束日期，格式: YYYY-MM-DD
}

// GetETFStatisticsResponse 获取ETF统计响应
type GetETFStatisticsResponse struct {
	Success bool                     `json:"success"`
	Data    map[string]ETFStatistics `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// GetETFStatistics 获取ETF历史统计数据
// @Summary 获取ETF历史统计数据
// @Description 基于历史价格数据计算ETF的年化收益、波动率、夏普比率等统计指标
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body GetETFStatisticsRequest true "请求参数"
// @Success 200 {object} GetETFStatisticsResponse
// @Router /api/optimization/etf-statistics [post]
func (h *OptimizationHandler) GetETFStatistics(c *gin.Context) {
	var req GetETFStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GetETFStatisticsResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 解析结束日期
	endDate := time.Now()
	if req.EndDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, GetETFStatisticsResponse{
				Success: false,
				Error:   "结束日期格式错误，请使用 YYYY-MM-DD 格式",
			})
			return
		}
		endDate = parsedDate
	}

	// 解析数据周期
	periodDays := 365 * 3 // 默认3年
	switch req.Period {
	case "1y":
		periodDays = 365
	case "3y":
		periodDays = 365 * 3
	case "5y":
		periodDays = 365 * 5
	case "10y":
		periodDays = 365 * 10
	}

	startDate := endDate.AddDate(0, 0, -periodDays)

	// 获取每个ETF的统计数据
	result := make(map[string]ETFStatistics)
	for _, symbol := range req.Symbols {
		stats := calculateETFStatistics(symbol, startDate, endDate)
		if stats != nil {
			result[symbol] = *stats
		}
	}

	c.JSON(http.StatusOK, GetETFStatisticsResponse{
		Success: true,
		Data:    result,
	})
}

// calculateETFStatistics 计算单个ETF的统计数据
func calculateETFStatistics(symbol string, startDate, endDate time.Time) *ETFStatistics {
	// 从数据库获取ETF历史数据
	var etfData []models.ETFData
	if err := models.DB.Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("date ASC").
		Find(&etfData).Error; err != nil {
		return nil
	}

	if len(etfData) < 30 { // 至少需要30个数据点
		return nil
	}

	// 计算收益率序列
	returns := utils.CalculateReturnsFromETFData(etfData)
	if len(returns) < 2 {
		return nil
	}

	// 转换为float64
	returnsFloat := make([]float64, len(returns))
	for i, r := range returns {
		returnsFloat[i], _ = r.Float64()
	}

	// 计算年化收益率
	avgReturn := 0.0
	for _, r := range returnsFloat {
		avgReturn += r
	}
	avgReturn /= float64(len(returnsFloat))
	// 使用几何平均法计算年化收益率: (1 + 平均日收益率)^252 - 1
	annualizedReturn := math.Pow(1+avgReturn, 252) - 1

	// 计算年化波动率 (样本标准差)
	variance := 0.0
	for _, r := range returnsFloat {
		variance += (r - avgReturn) * (r - avgReturn)
	}
	stdDev := math.Sqrt(variance / float64(len(returnsFloat)-1))
	annualizedVol := stdDev * math.Sqrt(252)

	// 计算夏普比率 (假设无风险利率为2%)
	riskFreeRate := 0.02
	sharpe := 0.0
	if annualizedVol > 0.001 {
		sharpe = (annualizedReturn - riskFreeRate) / annualizedVol
	}

	// 计算最大回撤
	maxDrawdown := calculateMaxDrawdownFromPrices(etfData)

	// 获取ETF名称
	name := symbol
	var etfConfig models.ETFConfig
	if err := models.DB.Where("symbol = ?", symbol).First(&etfConfig).Error; err == nil {
		name = etfConfig.Name
	}

	// 数据验证：确保数值在合理范围内
	stats := &ETFStatistics{
		Symbol:      symbol,
		Name:        name,
		Annualized:  annualizedReturn,
		Volatility:  annualizedVol,
		Sharpe:      sharpe,
		MaxDrawdown: maxDrawdown,
		DataPoints:  len(etfData),
	}

	// 验证数据合理性
	if !validateETFStatistics(stats) {
		return nil
	}

	return stats
}

// calculateMaxDrawdownFromPrices 从价格数据计算最大回撤
func calculateMaxDrawdownFromPrices(etfData []models.ETFData) float64 {
	if len(etfData) == 0 {
		return 0
	}

	// 数据已从数据库按日期排序，无需再次排序
	peak := decimal.Zero
	maxDrawdown := decimal.Zero

	for _, data := range etfData {
		price := data.ClosePrice
		if price.GreaterThan(peak) {
			peak = price
		}
		if peak.GreaterThan(decimal.Zero) {
			drawdown := peak.Sub(price).Div(peak)
			if drawdown.GreaterThan(maxDrawdown) {
				maxDrawdown = drawdown
			}
		}
	}

	md, _ := maxDrawdown.Float64()
	return md
}

// validateETFStatistics 验证ETF统计数据是否在合理范围内
func validateETFStatistics(stats *ETFStatistics) bool {
	// 检查数据点数是否足够
	if stats.DataPoints < 30 {
		return false
	}

	// 检查年化收益率是否在合理范围 (-50% 到 +100%)
	if stats.Annualized < -0.5 || stats.Annualized > 1.0 {
		return false
	}

	// 检查波动率是否在合理范围 (0.1% 到 100%)
	if stats.Volatility < 0.001 || stats.Volatility > 1.0 {
		return false
	}

	// 检查夏普比率是否在合理范围 (-5 到 +5)
	if stats.Sharpe < -5.0 || stats.Sharpe > 5.0 {
		return false
	}

	// 检查最大回撤是否在合理范围 (0 到 100%)
	if stats.MaxDrawdown < 0 || stats.MaxDrawdown > 1.0 {
		return false
	}

	return true
}

// calculateReturnsAndCovMatrix 从历史数据计算预期收益率和协方差矩阵
func (h *OptimizationHandler) calculateReturnsAndCovMatrix(symbols []string) (map[string]float64, map[string]map[string]float64, error) {
	// 获取所有ETF的历史价格数据
	priceData := make(map[string][]models.ETFData)
	for _, symbol := range symbols {
		var prices []models.ETFData
		if err := models.DB.Where("symbol = ?", symbol).Order("date ASC").Limit(252).Find(&prices).Error; err != nil {
			return nil, nil, err
		}
		if len(prices) < 30 {
			return nil, nil, fmt.Errorf("insufficient data for symbol %s: got %d prices, need at least 30", symbol, len(prices))
		}
		priceData[symbol] = prices
	}

	// 计算日收益率
	returnsData := make(map[string][]float64)
	expectedReturns := make(map[string]float64)

	for symbol, prices := range priceData {
		returns := make([]float64, 0, len(prices)-1)
		for i := 1; i < len(prices); i++ {
			if prices[i-1].ClosePrice.GreaterThan(decimal.Zero) {
				dailyReturn := prices[i].ClosePrice.Sub(prices[i-1].ClosePrice).Div(prices[i-1].ClosePrice)
				ret, _ := dailyReturn.Float64()
				returns = append(returns, ret)
			}
		}
		returnsData[symbol] = returns

		// 计算年化预期收益率
		if len(returns) > 0 {
			avgReturn := 0.0
			for _, r := range returns {
				avgReturn += r
			}
			avgReturn /= float64(len(returns))
			// 年化收益率: (1 + 平均日收益率)^252 - 1
			expectedReturns[symbol] = math.Pow(1+avgReturn, 252) - 1
		}
	}

	// 计算协方差矩阵
	covMatrix := calculateCovarianceMatrix(returnsData)

	return expectedReturns, covMatrix, nil
}
