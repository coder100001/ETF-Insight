package handlers

import (
	"encoding/json"
	"net/http"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PortfolioHandler struct {
	analysisService *services.ETFAnalysisService
}

func NewPortfolioHandler(analysisService *services.ETFAnalysisService) *PortfolioHandler {
	return &PortfolioHandler{
		analysisService: analysisService,
	}
}

type PortfolioRequest struct {
	Allocation      map[string]float64 `json:"allocation"`
	TotalInvestment float64            `json:"total_investment"`
	TaxRate         float64            `json:"tax_rate"`
}

type CreatePortfolioConfigRequest struct {
	Name            string             `json:"name" binding:"required"`
	Description     string             `json:"description"`
	Allocation      map[string]float64 `json:"allocation" binding:"required"`
	TotalInvestment float64            `json:"total_investment"`
	TaxRate         float64            `json:"tax_rate"`
	IsDefault       bool               `json:"is_default"`
}

type UpdatePortfolioConfigRequest struct {
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Allocation      map[string]float64 `json:"allocation"`
	TotalInvestment float64            `json:"total_investment"`
	TaxRate         float64            `json:"tax_rate"`
	IsDefault       bool               `json:"is_default"`
}

func (h *PortfolioHandler) AnalyzePortfolio(c *gin.Context) {
	var req PortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.TotalInvestment == 0 {
		req.TotalInvestment = 10000
	}
	if req.TaxRate == 0 {
		req.TaxRate = 0.10
	}

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

	response := map[string]interface{}{
		"total_value":                        result.TotalValue.InexactFloat64(),
		"total_return":                       result.TotalReturn.InexactFloat64(),
		"total_return_pct":                   result.TotalReturnPercent.InexactFloat64(),
		"annual_dividend_before_tax":         result.AnnualDividendBeforeTax.InexactFloat64(),
		"annual_dividend_after_tax":          result.AnnualDividendAfterTax.InexactFloat64(),
		"dividend_yield":                     result.WeightedDividendYield.InexactFloat64(),
		"tax_rate":                           result.TaxRate.InexactFloat64() * 100,
		"after_tax_return":                   result.TotalReturnWithDividend.InexactFloat64(),
		"holdings":                           result.Holdings,
		"total_investment":                   result.TotalInvestment.InexactFloat64(),
		"dividend_tax":                       result.DividendTax.InexactFloat64(),
		"total_return_with_dividend":         result.TotalReturnWithDividend.InexactFloat64(),
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
	result := models.DB.Order("id desc").Find(&configs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch portfolio configs",
		})
		return
	}

	var response []map[string]interface{}
	for _, config := range configs {
		var allocation map[string]float64
		if err := json.Unmarshal([]byte(config.Allocation), &allocation); err != nil {
			allocation = make(map[string]float64)
		}

		response = append(response, map[string]interface{}{
			"id":               config.ID,
			"name":             config.Name,
			"description":      config.Description,
			"allocation":       allocation,
			"total_investment": config.TotalInvestment.InexactFloat64(),
			"tax_rate":         config.TaxRate.InexactFloat64(),
			"status":           config.Status,
			"is_default":       config.IsDefault,
			"created_at":       config.CreatedAt,
			"updated_at":       config.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetPortfolioConfig 获取单个投资组合配置
func (h *PortfolioHandler) GetPortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	var config models.PortfolioConfig

	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Portfolio config not found",
		})
		return
	}

	var allocation map[string]float64
	if err := json.Unmarshal([]byte(config.Allocation), &allocation); err != nil {
		allocation = make(map[string]float64)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":               config.ID,
			"name":             config.Name,
			"description":      config.Description,
			"allocation":       allocation,
			"total_investment": config.TotalInvestment.InexactFloat64(),
			"tax_rate":         config.TaxRate.InexactFloat64(),
			"status":           config.Status,
			"is_default":       config.IsDefault,
			"created_at":       config.CreatedAt,
			"updated_at":       config.UpdatedAt,
		},
	})
}

// CreatePortfolioConfig 创建投资组合配置
func (h *PortfolioHandler) CreatePortfolioConfig(c *gin.Context) {
	var req CreatePortfolioConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// 将allocation转换为JSON字符串
	allocationJSON, err := json.Marshal(req.Allocation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid allocation format",
		})
		return
	}

	// 设置默认值
	totalInvestment := decimal.NewFromFloat(10000)
	if req.TotalInvestment > 0 {
		totalInvestment = decimal.NewFromFloat(req.TotalInvestment)
	}

	taxRate := decimal.NewFromFloat(0.10)
	if req.TaxRate > 0 {
		taxRate = decimal.NewFromFloat(req.TaxRate)
	}

	// 如果设置为默认组合，取消其他默认组合
	if req.IsDefault {
		models.DB.Model(&models.PortfolioConfig{}).Update("is_default", false)
	}

	config := models.PortfolioConfig{
		Name:            req.Name,
		Description:     req.Description,
		Allocation:      string(allocationJSON),
		TotalInvestment: totalInvestment,
		TaxRate:         taxRate,
		Status:          1,
		IsDefault:       req.IsDefault,
	}

	if err := models.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create portfolio config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":               config.ID,
			"name":             config.Name,
			"description":      config.Description,
			"allocation":       req.Allocation,
			"total_investment": config.TotalInvestment.InexactFloat64(),
			"tax_rate":         config.TaxRate.InexactFloat64(),
			"status":           config.Status,
			"is_default":       config.IsDefault,
			"created_at":       config.CreatedAt,
			"updated_at":       config.UpdatedAt,
		},
	})
}

// UpdatePortfolioConfig 更新投资组合配置
func (h *PortfolioHandler) UpdatePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	var config models.PortfolioConfig

	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Portfolio config not found",
		})
		return
	}

	var req UpdatePortfolioConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// 更新字段
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	if req.Allocation != nil && len(req.Allocation) > 0 {
		allocationJSON, err := json.Marshal(req.Allocation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid allocation format",
			})
			return
		}
		config.Allocation = string(allocationJSON)
	}
	if req.TotalInvestment > 0 {
		config.TotalInvestment = decimal.NewFromFloat(req.TotalInvestment)
	}
	if req.TaxRate >= 0 {
		config.TaxRate = decimal.NewFromFloat(req.TaxRate)
	}

	// 如果设置为默认组合，取消其他默认组合
	if req.IsDefault && !config.IsDefault {
		models.DB.Model(&models.PortfolioConfig{}).Where("id != ?", config.ID).Update("is_default", false)
		config.IsDefault = true
	} else if req.IsDefault != config.IsDefault {
		config.IsDefault = req.IsDefault
	}

	if err := models.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update portfolio config: " + err.Error(),
		})
		return
	}

	var allocation map[string]float64
	json.Unmarshal([]byte(config.Allocation), &allocation)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":               config.ID,
			"name":             config.Name,
			"description":      config.Description,
			"allocation":       allocation,
			"total_investment": config.TotalInvestment.InexactFloat64(),
			"tax_rate":         config.TaxRate.InexactFloat64(),
			"status":           config.Status,
			"is_default":       config.IsDefault,
			"created_at":       config.CreatedAt,
			"updated_at":       config.UpdatedAt,
		},
	})
}

// DeletePortfolioConfig 删除投资组合配置
func (h *PortfolioHandler) DeletePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	var config models.PortfolioConfig

	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Portfolio config not found",
		})
		return
	}

	if err := models.DB.Delete(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete portfolio config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Portfolio config deleted successfully",
	})
}

// TogglePortfolioConfigStatus 切换投资组合配置状态
func (h *PortfolioHandler) TogglePortfolioConfigStatus(c *gin.Context) {
	id := c.Param("id")
	var config models.PortfolioConfig

	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Portfolio config not found",
		})
		return
	}

	// 切换状态
	newStatus := 0
	if config.Status == 0 {
		newStatus = 1
	}
	config.Status = newStatus

	if err := models.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update portfolio config status: " + err.Error(),
		})
		return
	}

	var allocation map[string]float64
	json.Unmarshal([]byte(config.Allocation), &allocation)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":               config.ID,
			"name":             config.Name,
			"allocation":       allocation,
			"total_investment": config.TotalInvestment.InexactFloat64(),
			"status":           config.Status,
			"is_default":       config.IsDefault,
		},
	})
}

// AnalyzePortfolioConfig 分析投资组合配置
func (h *PortfolioHandler) AnalyzePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	var config models.PortfolioConfig

	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Portfolio config not found",
		})
		return
	}

	var req struct {
		TaxRate float64 `json:"tax_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.TaxRate = config.TaxRate.InexactFloat64()
	}
	if req.TaxRate == 0 {
		req.TaxRate = 0.10
	}

	// 解析allocation
	var allocation map[string]float64
	if err := json.Unmarshal([]byte(config.Allocation), &allocation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid allocation data",
		})
		return
	}

	result, err := h.analysisService.AnalyzePortfolio(
		allocation,
		config.TotalInvestment,
		decimal.NewFromFloat(req.TaxRate),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"config_id":                          config.ID,
		"config_name":                        config.Name,
		"total_value":                        result.TotalValue.InexactFloat64(),
		"total_return":                       result.TotalReturn.InexactFloat64(),
		"total_return_pct":                   result.TotalReturnPercent.InexactFloat64(),
		"annual_dividend_before_tax":         result.AnnualDividendBeforeTax.InexactFloat64(),
		"annual_dividend_after_tax":          result.AnnualDividendAfterTax.InexactFloat64(),
		"dividend_yield":                     result.WeightedDividendYield.InexactFloat64(),
		"tax_rate":                           result.TaxRate.InexactFloat64() * 100,
		"after_tax_return":                   result.TotalReturnWithDividend.InexactFloat64(),
		"holdings":                           result.Holdings,
		"total_investment":                   result.TotalInvestment.InexactFloat64(),
		"dividend_tax":                       result.DividendTax.InexactFloat64(),
		"total_return_with_dividend":         result.TotalReturnWithDividend.InexactFloat64(),
		"total_return_with_dividend_percent": result.TotalReturnWithDividendPercent.InexactFloat64(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
