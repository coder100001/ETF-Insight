package handlers

import (
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
	Returns      map[string]float64            `json:"returns" binding:"required"`
	CovMatrix    map[string]map[string]float64 `json:"cov_matrix" binding:"required"`
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
	Returns     map[string]float64            `json:"returns" binding:"required"`
	CovMatrix   map[string]map[string]float64 `json:"cov_matrix" binding:"required"`
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
	Returns     map[string]float64            `json:"returns" binding:"required"`
	CovMatrix   map[string]map[string]float64 `json:"cov_matrix" binding:"required"`
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
		result, err = h.riskParityOptimizer.OptimizeInverseVol(req.Returns, req.CovMatrix, constraint)
	case "budget":
		if len(req.RiskBudget) == 0 {
			c.JSON(http.StatusBadRequest, RiskParityResponse{
				Success: false,
				Error:   "风险预算方法需要提供 risk_budget 参数",
			})
			return
		}
		result, err = h.riskParityOptimizer.CalculateRiskBudget(req.Returns, req.CovMatrix, req.RiskBudget, constraint)
	default:
		result, err = h.riskParityOptimizer.Optimize(req.Returns, req.CovMatrix, constraint)
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

// ==================== Black-Litterman 优化 API ====================

// BlackLittermanRequest Black-Litterman优化请求
type BlackLittermanRequest struct {
	MarketWeights map[string]float64              `json:"market_weights" binding:"required"`
	CovMatrix     map[string]map[string]float64   `json:"cov_matrix" binding:"required"`
	AbsoluteViews map[string]float64              `json:"absolute_views,omitempty"`
	RelativeViews []*RelativeViewConfig           `json:"relative_views,omitempty"`
	RiskAversion  float64                         `json:"risk_aversion,omitempty"`
	Tau           float64                         `json:"tau,omitempty"`
	RiskFreeRate  float64                         `json:"risk_free_rate,omitempty"`
	Constraints   *BlackLittermanConstraintConfig `json:"constraints,omitempty"`
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

	// 设置参数
	if req.Tau > 0 {
		h.blackLittermanOptimizer.SetTau(req.Tau)
	}
	if req.RiskFreeRate > 0 {
		h.blackLittermanOptimizer.SetRiskFreeRate(req.RiskFreeRate)
	}

	// 构建约束条件
	symbols := make([]string, 0, len(req.MarketWeights))
	for symbol := range req.MarketWeights {
		symbols = append(symbols, symbol)
	}
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
		req.MarketWeights,
		req.CovMatrix,
		req.AbsoluteViews,
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
