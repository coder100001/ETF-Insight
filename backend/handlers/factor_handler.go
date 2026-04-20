package handlers

import (
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/services/factor"

	"github.com/gin-gonic/gin"
)

// FactorHandler 因子分析处理器
type FactorHandler struct {
	ffModel *factor.FamaFrenchModel
}

// NewFactorHandler 创建因子分析处理器
func NewFactorHandler() *FactorHandler {
	return &FactorHandler{
		ffModel: factor.NewFamaFrenchModel(),
	}
}

// ==================== Fama-French 因子分析 API ====================

// FamaFrenchAnalysisRequest Fama-French因子分析请求
type FamaFrenchAnalysisRequest struct {
	Returns       []float64 `json:"returns" binding:"required"` // 资产收益率序列
	Symbol        string    `json:"symbol,omitempty"`           // 资产代码
	UseFiveFactor bool      `json:"use_five_factor,omitempty"`  // 是否使用五因子
	Periods       int       `json:"periods,omitempty"`          // 数据周期数
}

// FamaFrenchAnalysisResponse Fama-French因子分析响应
type FamaFrenchAnalysisResponse struct {
	Success bool                      `json:"success"`
	Data    *factor.FactorAttribution `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// AnalyzeFactorExposure 分析因子暴露
// @Summary 分析资产或组合的Fama-French因子暴露
// @Description 使用Fama-French三因子或五因子模型分析收益率来源
// @Tags factor
// @Accept json
// @Produce json
// @Param request body FamaFrenchAnalysisRequest true "分析参数"
// @Success 200 {object} FamaFrenchAnalysisResponse
// @Router /api/factor/analyze [post]
func (h *FactorHandler) AnalyzeFactorExposure(c *gin.Context) {
	var req FamaFrenchAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FamaFrenchAnalysisResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 设置模型
	h.ffModel.SetFiveFactor(req.UseFiveFactor)

	// 加载因子数据
	periods := req.Periods
	if periods == 0 {
		periods = len(req.Returns)
	}

	// 尝试从数据库加载真实因子数据
	var marketReturns, smbReturns, hmlReturns, riskFreeReturns []float64
	var factorErr error

	// 计算日期范围 (假设使用月度数据，periods为月数)
	endDate := time.Now()
	startDate := endDate.AddDate(0, -periods, 0)

	// 尝试从数据库加载
	marketReturns, smbReturns, hmlReturns, riskFreeReturns, factorErr = factor.LoadFactorDataFromDB(
		models.DB, startDate, endDate,
	)

	// 如果数据库加载失败，使用ETF代理计算
	if factorErr != nil {
		marketReturns, smbReturns, hmlReturns, factorErr = factor.CalculateFactorFromETFs(
			"SPY", // 市场ETF
			"IWM", // 小盘ETF
			"VV",  // 大盘ETF
			"VTV", // 价值ETF
			"VUG", // 成长ETF
			startDate, endDate,
		)
		if factorErr != nil {
			// 如果ETF计算也失败，最后使用模拟数据
			marketReturns, smbReturns, hmlReturns, riskFreeReturns = factor.GenerateSampleFactorData(periods)
		} else {
			// ETF计算成功，但需要无风险利率
			riskFreeReturns = make([]float64, len(marketReturns))
			for i := range riskFreeReturns {
				riskFreeReturns[i] = 0.0015 // 默认月化无风险利率
			}
		}
	}

	h.ffModel.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// 执行分析
	var result *factor.FactorAttribution
	var err error

	if req.Symbol != "" {
		result, err = h.ffModel.AnalyzeETF(req.Returns, req.Symbol)
	} else {
		result, err = h.ffModel.AnalyzePortfolio(req.Returns, nil)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, FamaFrenchAnalysisResponse{
			Success: false,
			Error:   "分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, FamaFrenchAnalysisResponse{
		Success: true,
		Data:    result,
	})
}

// PortfolioFactorAnalysisRequest 组合因子分析请求
type PortfolioFactorAnalysisRequest struct {
	PortfolioReturns []float64          `json:"portfolio_returns" binding:"required"` // 组合收益率
	Weights          map[string]float64 `json:"weights,omitempty"`                    // 组合权重
	UseFiveFactor    bool               `json:"use_five_factor,omitempty"`            // 是否使用五因子
}

// PortfolioFactorAnalysisResponse 组合因子分析响应
type PortfolioFactorAnalysisResponse struct {
	Success bool                      `json:"success"`
	Data    *factor.FactorAttribution `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// AnalyzePortfolioFactors 分析投资组合因子
// @Summary 分析投资组合的Fama-French因子暴露
// @Description 分析投资组合的因子暴露和收益归因
// @Tags factor
// @Accept json
// @Produce json
// @Param request body PortfolioFactorAnalysisRequest true "分析参数"
// @Success 200 {object} PortfolioFactorAnalysisResponse
// @Router /api/factor/portfolio [post]
func (h *FactorHandler) AnalyzePortfolioFactors(c *gin.Context) {
	var req PortfolioFactorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PortfolioFactorAnalysisResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 设置模型
	h.ffModel.SetFiveFactor(req.UseFiveFactor)

	// 加载示例因子数据
	periods := len(req.PortfolioReturns)
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := factor.GenerateSampleFactorData(periods)
	h.ffModel.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// 执行分析
	result, err := h.ffModel.AnalyzePortfolio(req.PortfolioReturns, req.Weights)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PortfolioFactorAnalysisResponse{
			Success: false,
			Error:   "分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PortfolioFactorAnalysisResponse{
		Success: true,
		Data:    result,
	})
}

// MultiAssetFactorAnalysisRequest 多资产因子分析请求
type MultiAssetFactorAnalysisRequest struct {
	Assets        map[string][]float64 `json:"assets" binding:"required"` // 资产收益率映射
	UseFiveFactor bool                 `json:"use_five_factor,omitempty"` // 是否使用五因子
}

// MultiAssetFactorAnalysisResponse 多资产因子分析响应
type MultiAssetFactorAnalysisResponse struct {
	Success bool                                 `json:"success"`
	Data    map[string]*factor.FactorAttribution `json:"data,omitempty"`
	Error   string                               `json:"error,omitempty"`
}

// AnalyzeMultipleAssets 分析多个资产的因子暴露
// @Summary 批量分析多个资产的因子暴露
// @Description 同时分析多个资产的Fama-French因子特征
// @Tags factor
// @Accept json
// @Produce json
// @Param request body MultiAssetFactorAnalysisRequest true "分析参数"
// @Success 200 {object} MultiAssetFactorAnalysisResponse
// @Router /api/factor/multi-asset [post]
func (h *FactorHandler) AnalyzeMultipleAssets(c *gin.Context) {
	var req MultiAssetFactorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MultiAssetFactorAnalysisResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 设置模型
	h.ffModel.SetFiveFactor(req.UseFiveFactor)

	// 获取数据长度
	maxLen := 0
	for _, returns := range req.Assets {
		if len(returns) > maxLen {
			maxLen = len(returns)
		}
	}

	// 加载示例因子数据
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := factor.GenerateSampleFactorData(maxLen)
	h.ffModel.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// 批量分析
	results := h.ffModel.ComparePortfolios(req.Assets)

	c.JSON(http.StatusOK, MultiAssetFactorAnalysisResponse{
		Success: true,
		Data:    results,
	})
}

// FactorStatisticsResponse 因子统计响应
type FactorStatisticsResponse struct {
	Success bool                   `json:"success"`
	Data    []*factor.FactorReturn `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// GetFactorStatistics 获取因子统计信息
// @Summary 获取Fama-French因子统计信息
// @Description 获取各因子的年化收益、波动率、夏普比率等统计指标
// @Tags factor
// @Accept json
// @Produce json
// @Param five_factor query bool false "是否使用五因子"
// @Success 200 {object} FactorStatisticsResponse
// @Router /api/factor/statistics [get]
func (h *FactorHandler) GetFactorStatistics(c *gin.Context) {
	useFiveFactor := c.Query("five_factor") == "true"
	h.ffModel.SetFiveFactor(useFiveFactor)

	// 生成示例数据
	periods := 120 // 10年月度数据
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := factor.GenerateSampleFactorData(periods)

	if useFiveFactor {
		rmwReturns := make([]float64, periods)
		cmaReturns := make([]float64, periods)
		for i := 0; i < periods; i++ {
			rmwReturns[i] = 0.002
			cmaReturns[i] = 0.001
		}
		h.ffModel.LoadFiveFactorData(marketReturns, smbReturns, hmlReturns, rmwReturns, cmaReturns, riskFreeReturns)
	} else {
		h.ffModel.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)
	}

	stats := h.ffModel.GetFactorStatistics()

	c.JSON(http.StatusOK, FactorStatisticsResponse{
		Success: true,
		Data:    stats,
	})
}

// RiskDecompositionRequest 风险分解请求
type RiskDecompositionRequest struct {
	Exposures     *factor.FactorExposure `json:"exposures" binding:"required"` // 因子暴露
	UseFiveFactor bool                   `json:"use_five_factor,omitempty"`    // 是否使用五因子
}

// RiskDecompositionResponse 风险分解响应
type RiskDecompositionResponse struct {
	Success bool               `json:"success"`
	Data    map[string]float64 `json:"data,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// DecomposeRisk 风险分解
// @Summary 基于因子暴露进行风险分解
// @Description 计算各因子对组合风险的贡献比例
// @Tags factor
// @Accept json
// @Produce json
// @Param request body RiskDecompositionRequest true "分解参数"
// @Success 200 {object} RiskDecompositionResponse
// @Router /api/factor/risk-decomposition [post]
func (h *FactorHandler) DecomposeRisk(c *gin.Context) {
	var req RiskDecompositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RiskDecompositionResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	h.ffModel.SetFiveFactor(req.UseFiveFactor)
	decomposition := h.ffModel.RiskDecomposition(req.Exposures)

	c.JSON(http.StatusOK, RiskDecompositionResponse{
		Success: true,
		Data:    decomposition,
	})
}

// FactorAttributionComparisonRequest 因子归因对比请求
type FactorAttributionComparisonRequest struct {
	Portfolios    map[string][]float64 `json:"portfolios" binding:"required"` // 多个组合的收益率
	Factor        string               `json:"factor" binding:"required"`     // 对比因子
	UseFiveFactor bool                 `json:"use_five_factor,omitempty"`     // 是否使用五因子
}

// FactorAttributionComparisonResponse 因子归因对比响应
type FactorAttributionComparisonResponse struct {
	Success bool     `json:"success"`
	Data    []string `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// CompareFactorAttribution 对比因子归因
// @Summary 对比多个组合的因子暴露
// @Description 按指定因子对多个组合进行排序
// @Tags factor
// @Accept json
// @Produce json
// @Param request body FactorAttributionComparisonRequest true "对比参数"
// @Success 200 {object} FactorAttributionComparisonResponse
// @Router /api/factor/compare [post]
func (h *FactorHandler) CompareFactorAttribution(c *gin.Context) {
	var req FactorAttributionComparisonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FactorAttributionComparisonResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// 设置模型
	h.ffModel.SetFiveFactor(req.UseFiveFactor)

	// 获取数据长度
	maxLen := 0
	for _, returns := range req.Portfolios {
		if len(returns) > maxLen {
			maxLen = len(returns)
		}
	}

	// 加载示例因子数据
	marketReturns, smbReturns, hmlReturns, riskFreeReturns := factor.GenerateSampleFactorData(maxLen)
	h.ffModel.LoadFactorData(marketReturns, smbReturns, hmlReturns, riskFreeReturns)

	// 分析所有组合
	attributions := h.ffModel.ComparePortfolios(req.Portfolios)

	// 按指定因子排序
	sorted := factor.SortByFactorExposure(attributions, req.Factor)

	c.JSON(http.StatusOK, FactorAttributionComparisonResponse{
		Success: true,
		Data:    sorted,
	})
}
