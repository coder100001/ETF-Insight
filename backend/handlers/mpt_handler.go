package handlers

import (
	"fmt"
	"math"
	"net/http"

	"etf-insight/models"
	"etf-insight/services/optimization"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MPTHandler 均值-方差优化处理器
type MPTHandler struct{}

// NewMPTHandler 创建MPT处理器
func NewMPTHandler() *MPTHandler {
	return &MPTHandler{}
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

// MPTOptimize 均值-方差优化
// @Summary 执行均值-方差优化
// @Description 基于现代投资组合理论(MPT)优化资产配置
// @Tags optimization
// @Accept json
// @Produce json
// @Param request body MPTOptimizeRequest true "优化参数"
// @Success 200 {object} MPTOptimizeResponse
// @Router /api/optimization/mpt [post]
func (h *MPTHandler) MPTOptimize(c *gin.Context) {
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
		returns, covMatrix, err := calculateReturnsAndCovMatrix(req.Symbols)
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

	// 创建优化器实例（并发安全）
	mptOpt := optimization.NewMPTOptimizer()

	// 设置无风险利率
	if req.RiskFreeRate > 0 {
		mptOpt.SetRiskFreeRate(req.RiskFreeRate)
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
		result, err = mptOpt.OptimizeMinVolatility(req.Returns, req.CovMatrix, constraint)
	case "max_sharpe":
		result, err = mptOpt.OptimizeMaxSharpe(req.Returns, req.CovMatrix, constraint)
	case "target_return":
		if req.TargetReturn == 0 {
			c.JSON(http.StatusBadRequest, MPTOptimizeResponse{
				Success: false,
				Error:   "目标收益率模式下必须提供 target_return 参数",
			})
			return
		}
		result, err = mptOpt.OptimizeForTargetReturn(req.Returns, req.CovMatrix, constraint, req.TargetReturn)
	default:
		result, err = mptOpt.Optimize(req.Returns, req.CovMatrix, constraint)
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
func (h *MPTHandler) EfficientFrontier(c *gin.Context) {
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
		returns, covMatrix, err := calculateReturnsAndCovMatrix(req.Symbols)
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

	// 创建优化器实例（并发安全）
	mptOpt := optimization.NewMPTOptimizer()

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
	frontier, err := mptOpt.CalculateEfficientFrontier(req.Returns, req.CovMatrix, constraint, req.NumPoints)
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
func (h *MPTHandler) CalculateCovarianceMatrix(c *gin.Context) {
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

			minLen := min(len(data2), len(data1))

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

// calculateReturnsAndCovMatrix 从历史数据计算预期收益率和协方差矩阵
func calculateReturnsAndCovMatrix(symbols []string) (map[string]float64, map[string]map[string]float64, error) {
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


