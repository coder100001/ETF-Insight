package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"

	"github.com/gin-gonic/gin"
)

// ETFConfigHandler ETF配置处理器
type ETFConfigHandler struct{}

// NewETFConfigHandler 创建ETF配置处理器
func NewETFConfigHandler() *ETFConfigHandler {
	return &ETFConfigHandler{}
}

// GetETFConfigs 获取ETF配置列表
func (h *ETFConfigHandler) GetETFConfigs(c *gin.Context) {
	var configs []models.ETFConfig
	models.DB.Find(&configs)

	// 如果没有配置，初始化默认配置
	if len(configs) == 0 {
		configs = getDefaultETFConfigs()
		// 保存到数据库
		for i := range configs {
			models.DB.Create(&configs[i])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// GetETFConfig 获取单个ETF配置
func (h *ETFConfigHandler) GetETFConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID",
		})
		return
	}

	var config models.ETFConfig
	models.DB.First(&config, id)

	// 检查是否找到
	if config.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// CreateETFConfig 创建ETF配置
func (h *ETFConfigHandler) CreateETFConfig(c *gin.Context) {
	var config models.ETFConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	models.DB.Create(&config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
		"message": "ETF配置创建成功",
	})
}

// UpdateETFConfig 更新ETF配置
func (h *ETFConfigHandler) UpdateETFConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID",
		})
		return
	}

	var config models.ETFConfig
	models.DB.First(&config, id)

	// 检查是否找到
	if config.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF配置不存在",
		})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	models.DB.Model(&config).Updates(updateData)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
		"message": "ETF配置更新成功",
	})
}

// DeleteETFConfig 删除ETF配置
func (h *ETFConfigHandler) DeleteETFConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID",
		})
		return
	}

	models.DB.Delete(&models.ETFConfig{}, id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ETF配置删除成功",
	})
}

// ToggleETFConfigStatus 切换ETF配置状态
func (h *ETFConfigHandler) ToggleETFConfigStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID",
		})
		return
	}

	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	var config models.ETFConfig
	models.DB.First(&config, id)

	// 检查是否找到
	if config.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF配置不存在",
		})
		return
	}

	config.Status = req.Status
	models.DB.Save(&config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
		"message": "状态更新成功",
	})
}

// ToggleETFConfigAutoUpdate 切换ETF配置自动更新
func (h *ETFConfigHandler) ToggleETFConfigAutoUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID",
		})
		return
	}

	var req struct {
		AutoUpdate bool `json:"auto_update"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	var config models.ETFConfig
	models.DB.First(&config, id)

	// 检查是否找到
	if config.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "ETF配置不存在",
		})
		return
	}

	// 更新自动更新设置
	config.AutoUpdate = req.AutoUpdate
	models.DB.Save(&config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
		"message": "自动更新设置成功",
	})
}

// getDefaultETFConfigs 获取默认ETF配置
func getDefaultETFConfigs() []models.ETFConfig {
	return []models.ETFConfig{
		{
			ID:              1,
			Symbol:          "SCHD",
			Name:            "Schwab US Dividend Equity ETF",
			Status:          1,
			AutoUpdate:      true,
			UpdateFrequency: "每日",
			DataSource:      "Yahoo Finance",
			Currency:        "USD",
			Category:        "ETF",
		},
		{
			ID:              2,
			Symbol:          "SPYD",
			Name:            "SPDR S&P 500 High Dividend ETF",
			Status:          1,
			AutoUpdate:      true,
			UpdateFrequency: "每日",
			DataSource:      "Yahoo Finance",
			Currency:        "USD",
			Category:        "ETF",
		},
		{
			ID:              3,
			Symbol:          "JEPQ",
			Name:            "JPMorgan Nasdaq Equity Premium Income ETF",
			Status:          1,
			AutoUpdate:      true,
			UpdateFrequency: "每日",
			DataSource:      "Yahoo Finance",
			Currency:        "USD",
			Category:        "ETF",
		},
		{
			ID:              4,
			Symbol:          "JEPI",
			Name:            "JPMorgan Equity Premium Income ETF",
			Status:          1,
			AutoUpdate:      true,
			UpdateFrequency: "每日",
			DataSource:      "Yahoo Finance",
			Currency:        "USD",
			Category:        "ETF",
		},
		{
			ID:              5,
			Symbol:          "VYM",
			Name:            "Vanguard High Dividend Yield ETF",
			Status:          1,
			AutoUpdate:      true,
			UpdateFrequency: "每日",
			DataSource:      "Yahoo Finance",
			Currency:        "USD",
			Category:        "ETF",
		},
	}
}
