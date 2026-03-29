package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// PortfolioHandler 投资组合处理器
type PortfolioHandler struct {
	analysisService *services.ETFAnalysisService
}

// NewPortfolioHandler 创建投资组合处理器
func NewPortfolioHandler(analysisService *services.ETFAnalysisService) *PortfolioHandler {
	return &PortfolioHandler{
		analysisService: analysisService,
	}
}

// PortfolioRequest 投资组合分析请求
type PortfolioRequest struct {
	Allocation      map[string]float64 `json:"allocation"`
	TotalInvestment float64            `json:"total_investment"`
	TaxRate         float64            `json:"tax_rate"`
}

// AnalyzePortfolio 分析投资组合
func (h *PortfolioHandler) AnalyzePortfolio(c *gin.Context) {
	var req PortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// 默认值
	if req.TotalInvestment == 0 {
		req.TotalInvestment = 10000
	}
	if req.TaxRate == 0 {
		req.TaxRate = 0.10
	}

	// 执行分析
	result, err := h.analysisService.AnalyzePortfolio(
		req.Allocation,
		decimal.NewFromFloat(req.TotalInvestment),
		decimal.NewFromFloat(req.TaxRate),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 转换结果为前端需要的格式
	response := map[string]interface{}{
		"total_value":              result.TotalValue.InexactFloat64(),
		"total_return":             result.TotalReturn.InexactFloat64(),
		"total_return_pct":         result.TotalReturnPercent.InexactFloat64(),
		"annual_dividend":          result.AnnualDividendAfterTax.InexactFloat64(),
		"dividend_yield":           result.WeightedDividendYield.InexactFloat64(),
		"tax_rate":                 result.TaxRate.InexactFloat64() * 100,
		"after_tax_return":         result.TotalReturnWithDividend.InexactFloat64(),
		"holdings":                 result.Holdings,
		"total_investment":         result.TotalInvestment.InexactFloat64(),
		"annual_dividend_before_tax": result.AnnualDividendBeforeTax.InexactFloat64(),
		"dividend_tax":             result.DividendTax.InexactFloat64(),
		"total_return_with_dividend": result.TotalReturnWithDividend.InexactFloat64(),
		"total_return_with_dividend_percent": result.TotalReturnWithDividendPercent.InexactFloat64(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetPortfolioConfigs 获取所有投资组合配置
func (h *PortfolioHandler) GetPortfolioConfigs(c *gin.Context) {
	var configs []models.PortfolioConfig
	if err := models.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// GetPortfolioConfig 获取单个投资组合配置
func (h *PortfolioHandler) GetPortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	
	var config models.PortfolioConfig
	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Config not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// CreatePortfolioConfig 创建投资组合配置
func (h *PortfolioHandler) CreatePortfolioConfig(c *gin.Context) {
	var config models.PortfolioConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if err := models.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdatePortfolioConfig 更新投资组合配置
func (h *PortfolioHandler) UpdatePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	
	var config models.PortfolioConfig
	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Config not found",
		})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if err := models.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// DeletePortfolioConfig 删除投资组合配置
func (h *PortfolioHandler) DeletePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	
	if err := models.DB.Delete(&models.PortfolioConfig{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Config deleted",
	})
}

// TogglePortfolioConfigStatus 切换投资组合配置状态
func (h *PortfolioHandler) TogglePortfolioConfigStatus(c *gin.Context) {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)
	
	var config models.PortfolioConfig
	if err := models.DB.First(&config, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Config not found",
		})
		return
	}

	config.Status = 1 - config.Status
	if err := models.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// AnalyzePortfolioConfig 分析投资组合配置
func (h *PortfolioHandler) AnalyzePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)
	
	var config models.PortfolioConfig
	if err := models.DB.First(&config, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Config not found",
		})
		return
	}

	var req struct {
		TaxRate float64 `json:"tax_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.TaxRate = 0.10
	}

	result, err := h.analysisService.AnalyzePortfolio(
		config.Allocation,
		decimal.NewFromFloat(config.TotalInvestment),
		decimal.NewFromFloat(req.TaxRate),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
