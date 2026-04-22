package handlers

import (
	"net/http"
	"time"

	"etf-insight/services/portfolio"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PortfolioPenetrationHandler 投资组合穿透处理器
type PortfolioPenetrationHandler struct {
	db                 *gorm.DB
	penetrationService *portfolio.PortfolioPenetrationService
}

// NewPortfolioPenetrationHandler 创建投资组合穿透处理器
func NewPortfolioPenetrationHandler(db *gorm.DB) *PortfolioPenetrationHandler {
	return &PortfolioPenetrationHandler{
		db:                 db,
		penetrationService: portfolio.NewPortfolioPenetrationService(db),
	}
}

// AnalyzePortfolioPenetration 分析投资组合持仓穿透
// POST /api/portfolio/penetration
func (h *PortfolioPenetrationHandler) AnalyzePortfolioPenetration(c *gin.Context) {
	var req struct {
		PortfolioID string `json:"portfolio_id"`
		Holdings    []struct {
			Symbol string          `json:"symbol" binding:"required"`
			Name   string          `json:"name"`
			Weight decimal.Decimal `json:"weight" binding:"required"`
		} `json:"holdings" binding:"required,min=1"`
		Date string `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
			"code":    "INVALID_REQUEST",
		})
		return
	}

	// 解析日期
	var date time.Time
	if req.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效，请使用 YYYY-MM-DD 格式",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	// 转换持仓数据
	var portfolioHoldings []portfolio.PortfolioHolding
	for _, h := range req.Holdings {
		portfolioHoldings = append(portfolioHoldings, portfolio.PortfolioHolding{
			Symbol: h.Symbol,
			Name:   h.Name,
			Weight: h.Weight,
		})
	}

	// 执行穿透分析
	result, err := h.penetrationService.AnalyzePortfolio(c.Request.Context(), req.PortfolioID, portfolioHoldings, date)
	if err != nil {
		utils.Error("投资组合穿透分析失败", err, "portfolio_id", req.PortfolioID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "穿透分析失败: " + err.Error(),
			"code":    "PENETRATION_ANALYSIS_ERROR",
		})
		return
	}

	// 转换行业分布
	var sectors []gin.H
	for sector, weight := range result.SectorAllocation {
		sectors = append(sectors, gin.H{
			"sector": sector,
			"weight": weight,
		})
	}

	// 转换地理分布
	var countries []gin.H
	for country, weight := range result.CountryAllocation {
		countries = append(countries, gin.H{
			"country": country,
			"weight":  weight,
		})
	}

	// 转换前十大持仓
	var topHoldings []gin.H
	for _, h := range result.TopHoldings {
		topHoldings = append(topHoldings, gin.H{
			"symbol":      h.Symbol,
			"name":        h.Name,
			"weight":      h.Weight,
			"sector":      h.Sector,
			"country":     h.Country,
			"source_etfs": h.SourceETFs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"portfolio_id":       result.PortfolioID,
			"total_etfs":         result.TotalETFs,
			"total_holdings":     result.TotalHoldings,
			"unique_holdings":    result.UniqueHoldings,
			"sector_allocation":  sectors,
			"country_allocation": countries,
			"top_holdings":       topHoldings,
			"concentration": gin.H{
				"top10_weight":       result.Concentration.Top10Weight,
				"top20_weight":       result.Concentration.Top20Weight,
				"herfindahl_index":   result.Concentration.HerfindahlIndex,
				"effective_holdings": result.Concentration.EffectiveHoldings,
			},
			"calculated_at": result.CalculatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// ComparePortfolios 对比两个投资组合
// POST /api/portfolio/compare
func (h *PortfolioPenetrationHandler) ComparePortfolios(c *gin.Context) {
	var req struct {
		PortfolioA []struct {
			Symbol string          `json:"symbol" binding:"required"`
			Name   string          `json:"name"`
			Weight decimal.Decimal `json:"weight" binding:"required"`
		} `json:"portfolio_a" binding:"required,min=1"`
		PortfolioB []struct {
			Symbol string          `json:"symbol" binding:"required"`
			Name   string          `json:"name"`
			Weight decimal.Decimal `json:"weight" binding:"required"`
		} `json:"portfolio_b" binding:"required,min=1"`
		Date string `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
			"code":    "INVALID_REQUEST",
		})
		return
	}

	// 解析日期
	var date time.Time
	if req.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	// 转换持仓数据
	var holdingsA, holdingsB []portfolio.PortfolioHolding
	for _, h := range req.PortfolioA {
		holdingsA = append(holdingsA, portfolio.PortfolioHolding{
			Symbol: h.Symbol,
			Name:   h.Name,
			Weight: h.Weight,
		})
	}
	for _, h := range req.PortfolioB {
		holdingsB = append(holdingsB, portfolio.PortfolioHolding{
			Symbol: h.Symbol,
			Name:   h.Name,
			Weight: h.Weight,
		})
	}

	// 执行对比
	comparison, err := h.penetrationService.ComparePortfolios(c.Request.Context(), holdingsA, holdingsB, date)
	if err != nil {
		utils.Error("投资组合对比失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "对比失败: " + err.Error(),
			"code":    "COMPARE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"portfolio_a":       formatPenetrationResult(comparison.PortfolioA),
			"portfolio_b":       formatPenetrationResult(comparison.PortfolioB),
			"common_holdings":   comparison.CommonHoldings,
			"unique_holdings_a": comparison.UniqueHoldingsA,
			"unique_holdings_b": comparison.UniqueHoldingsB,
		},
	})
}

// GetSectorExposure 获取投资组合的行业暴露
// POST /api/portfolio/sector-exposure
func (h *PortfolioPenetrationHandler) GetSectorExposure(c *gin.Context) {
	var req struct {
		Holdings []struct {
			Symbol string          `json:"symbol" binding:"required"`
			Weight decimal.Decimal `json:"weight" binding:"required"`
		} `json:"holdings" binding:"required,min=1"`
		Sector string `json:"sector" binding:"required"`
		Date   string `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
			"code":    "INVALID_REQUEST",
		})
		return
	}

	var date time.Time
	if req.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	var holdings []portfolio.PortfolioHolding
	for _, h := range req.Holdings {
		holdings = append(holdings, portfolio.PortfolioHolding{
			Symbol: h.Symbol,
			Weight: h.Weight,
		})
	}

	exposure, err := h.penetrationService.GetSectorExposure(c.Request.Context(), holdings, req.Sector, date)
	if err != nil {
		utils.Error("获取行业暴露失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取行业暴露失败",
			"code":    "SECTOR_EXPOSURE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sector":   req.Sector,
			"exposure": exposure,
		},
	})
}

// formatPenetrationResult 格式化穿透结果为响应格式
func formatPenetrationResult(result *portfolio.PenetrationResult) gin.H {
	if result == nil {
		return nil
	}

	var sectors []gin.H
	for sector, weight := range result.SectorAllocation {
		sectors = append(sectors, gin.H{
			"sector": sector,
			"weight": weight,
		})
	}

	var countries []gin.H
	for country, weight := range result.CountryAllocation {
		countries = append(countries, gin.H{
			"country": country,
			"weight":  weight,
		})
	}

	var topHoldings []gin.H
	for _, h := range result.TopHoldings {
		topHoldings = append(topHoldings, gin.H{
			"symbol":  h.Symbol,
			"name":    h.Name,
			"weight":  h.Weight,
			"sector":  h.Sector,
			"country": h.Country,
		})
	}

	return gin.H{
		"portfolio_id":       result.PortfolioID,
		"total_etfs":         result.TotalETFs,
		"unique_holdings":    result.UniqueHoldings,
		"sector_allocation":  sectors,
		"country_allocation": countries,
		"top_holdings":       topHoldings,
		"concentration": gin.H{
			"top10_weight":       result.Concentration.Top10Weight,
			"top20_weight":       result.Concentration.Top20Weight,
			"herfindahl_index":   result.Concentration.HerfindahlIndex,
			"effective_holdings": result.Concentration.EffectiveHoldings,
		},
	}
}
