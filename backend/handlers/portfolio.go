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
	analysisSvc *services.ETFAnalysisService
}

// NewPortfolioHandler 创建新的投资组合处理器
func NewPortfolioHandler(analysis *services.ETFAnalysisService) *PortfolioHandler {
	return &PortfolioHandler{
		analysisSvc: analysis,
	}
}

// GetConfigs 获取投资组合配置列表
func (h *PortfolioHandler) GetConfigs(c *gin.Context) {
	var configs []models.PortfolioConfig
	if err := models.DB.Order("created_at DESC").Find(&configs).Error; err != nil {
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

// CreateConfig 创建投资组合配置
func (h *PortfolioHandler) CreateConfig(c *gin.Context) {
	var req struct {
		Name            string             `json:"name" binding:"required"`
		Description     string             `json:"description"`
		Allocation      map[string]float64 `json:"allocation" binding:"required"`
		TotalInvestment float64            `json:"total_investment"`
		Status          int                `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 验证权重总和
	totalWeight := 0.0
	for _, weight := range req.Allocation {
		totalWeight += weight
	}
	if totalWeight < 0.99 || totalWeight > 1.01 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "allocation weights must sum to 1.0",
		})
		return
	}

	if req.TotalInvestment == 0 {
		req.TotalInvestment = 10000
	}

	config := models.PortfolioConfig{
		Name:            req.Name,
		Description:     req.Description,
		Allocation:      req.Allocation,
		TotalInvestment: decimal.NewFromFloat(req.TotalInvestment),
		Status:          req.Status,
	}

	if config.Status == 0 {
		config.Status = 1
	}

	if err := models.DB.Create(&config).Error; err != nil {
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

// GetConfigDetail 获取配置详情
func (h *PortfolioHandler) GetConfigDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	var config models.PortfolioConfig
	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "config not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateConfig 更新配置
func (h *PortfolioHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	var config models.PortfolioConfig
	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "config not found",
		})
		return
	}

	var req struct {
		Name            string             `json:"name"`
		Description     string             `json:"description"`
		Allocation      map[string]float64 `json:"allocation"`
		TotalInvestment float64            `json:"total_investment"`
		Status          int                `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	if req.Allocation != nil {
		// 验证权重总和
		totalWeight := 0.0
		for _, weight := range req.Allocation {
			totalWeight += weight
		}
		if totalWeight < 0.99 || totalWeight > 1.01 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "allocation weights must sum to 1.0",
			})
			return
		}
		config.Allocation = req.Allocation
	}
	if req.TotalInvestment > 0 {
		config.TotalInvestment = decimal.NewFromFloat(req.TotalInvestment)
	}
	if req.Status != 0 || c.Request.Body != nil {
		config.Status = req.Status
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

// DeleteConfig 删除配置
func (h *PortfolioHandler) DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	if err := models.DB.Delete(&models.PortfolioConfig{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "config deleted",
	})
}

// ToggleStatus 切换配置状态
func (h *PortfolioHandler) ToggleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	var config models.PortfolioConfig
	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "config not found",
		})
		return
	}

	if config.Status == 1 {
		config.Status = 0
	} else {
		config.Status = 1
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
		"data": map[string]interface{}{
			"id":          config.ID,
			"status":      config.Status,
			"status_text": map[int]string{0: "禁用", 1: "启用"}[config.Status],
		},
	})
}

// AnalyzeConfig 分析配置
func (h *PortfolioHandler) AnalyzeConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	var config models.PortfolioConfig
	if err := models.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "config not found",
		})
		return
	}

	var req struct {
		TaxRate float64 `json:"tax_rate"`
	}
	c.ShouldBindJSON(&req)
	if req.TaxRate == 0 {
		req.TaxRate = 0.10
	}

	result, err := h.analysisSvc.AnalyzePortfolio(
		config.Allocation,
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
