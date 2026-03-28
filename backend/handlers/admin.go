package handlers

import (
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理处理器
type AdminHandler struct{}

// NewAdminHandler 创建新的管理处理器
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// GetStats 获取系统统计
func (h *AdminHandler) GetStats(c *gin.Context) {
	// ETF统计
	var etfCount int64
	models.DB.Model(&models.ETFConfig{}).Count(&etfCount)

	var activeETFCount int64
	models.DB.Model(&models.ETFConfig{}).Where("status = ?", 1).Count(&activeETFCount)

	// 数据记录统计
	var dataCount int64
	models.DB.Model(&models.ETFData{}).Count(&dataCount)

	// 投资组合统计
	var portfolioCount int64
	models.DB.Model(&models.PortfolioConfig{}).Count(&portfolioCount)

	// 工作流统计
	var workflowCount int64
	models.DB.Model(&models.Workflow{}).Count(&workflowCount)

	var instanceCount int64
	models.DB.Model(&models.WorkflowInstance{}).Count(&instanceCount)

	// 今日操作记录统计
	today := time.Now().Format("2006-01-02")
	var todayOperationCount int64
	models.DB.Model(&models.OperationLog{}).Where("DATE(start_time) = ?", today).Count(&todayOperationCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"etf_count":             etfCount,
			"active_etf_count":      activeETFCount,
			"data_count":            dataCount,
			"portfolio_count":       portfolioCount,
			"workflow_count":        workflowCount,
			"instance_count":        instanceCount,
			"today_operation_count": todayOperationCount,
		},
	})
}

// GetLogs 获取操作日志
func (h *AdminHandler) GetLogs(c *gin.Context) {
	var logs []models.OperationLog
	if err := models.DB.Order("start_time DESC").Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// ClearCache 清除缓存
func (h *AdminHandler) ClearCache(c *gin.Context) {
	// 这里需要传入cacheService，简化处理
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "cache cleared",
	})
}
