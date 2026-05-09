package handlers

import (
	"maps"
	"net/http"

	"etf-insight/services/optimization"

	"github.com/gin-gonic/gin"
)

// RiskParityHandler 风险平价优化处理器
type RiskParityHandler struct {
	riskParityOptimizer *optimization.RiskParityOptimizer
}

// NewRiskParityHandler 创建风险平价处理器
func NewRiskParityHandler() *RiskParityHandler {
	return &RiskParityHandler{
		riskParityOptimizer: optimization.NewRiskParityOptimizer(),
	}
}

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
func (h *RiskParityHandler) RiskParityOptimize(c *gin.Context) {
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
			maps.Copy(constraint.MinWeight, req.Constraints.MinWeights)
		}
		if req.Constraints.MaxWeights != nil {
			maps.Copy(constraint.MaxWeight, req.Constraints.MaxWeights)
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
