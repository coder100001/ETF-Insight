package handlers

import (
	"math"
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ETFStatisticsHandler ETF历史数据统计处理器
type ETFStatisticsHandler struct{}

// NewETFStatisticsHandler 创建ETF统计处理器
func NewETFStatisticsHandler() *ETFStatisticsHandler {
	return &ETFStatisticsHandler{}
}

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
func (h *ETFStatisticsHandler) GetETFStatistics(c *gin.Context) {
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
