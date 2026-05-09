package handlers

import (
	"maps"
	"net/http"

	"etf-insight/services/optimization"

	"github.com/gin-gonic/gin"
)

// BLHandler Black-Litterman优化处理器
type BLHandler struct {
	blackLittermanOptimizer *optimization.BlackLittermanOptimizer
}

// NewBLHandler 创建Black-Litterman处理器
func NewBLHandler() *BLHandler {
	return &BLHandler{
		blackLittermanOptimizer: optimization.NewBlackLittermanOptimizer(),
	}
}

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

// BlackLittermanOptimize Black-Litterman优化
// @Summary 执行Black-Litterman优化
// @Description 融合市场均衡收益和投资者观点，生成后验收益估计并优化
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body BlackLittermanRequest true "优化参数"
// @Success 200 {object} BlackLittermanResponse
// @Router /api/optimization/black-litterman [post]
func (h *BLHandler) BlackLittermanOptimize(c *gin.Context) {
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
			maps.Copy(constraint.MinWeight, req.Constraints.MinWeights)
		}
		if req.Constraints.MaxWeights != nil {
			maps.Copy(constraint.MaxWeight, req.Constraints.MaxWeights)
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

// MarketImpliedReturns 计算市场隐含收益
// @Summary 计算市场隐含收益（反向优化）
// @Description 基于市场权重和协方差矩阵，反推出市场隐含的收益预期
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body MarketImpliedReturnsRequest true "计算参数"
// @Success 200 {object} MarketImpliedReturnsResponse
// @Router /api/optimization/market-implied-returns [post]
func (h *BLHandler) MarketImpliedReturns(c *gin.Context) {
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
