package handlers

import (
	"net/http"

	"etf-insight/services/export"
	"etf-insight/services/export/service"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	exportService *service.ExportService
}

// NewExportHandler 创建导出处理器
func NewExportHandler() *ExportHandler {
	return &ExportHandler{
		exportService: service.NewExportService(),
	}
}

// Export 导出API
// @Summary 导出数据
// @Description 将数据导出为指定格式（HTML、PDF、Excel、Markdown）
// @Tags export
// @Accept json
// @Produce json
// @Param type path string true "页面类型（portfolio、risk等）"
// @Param request body export.ExportRequest true "导出请求"
// @Success 200 {object} export.ExportResponse
// @Failure 400 {object} export.ExportResponse
// @Failure 500 {object} export.ExportResponse
// @Router /api/export/{type} [post]
func (h *ExportHandler) Export(c *gin.Context) {
	pageType := c.Param("type")
	if pageType == "" {
		c.JSON(http.StatusBadRequest, export.ExportResponse{
			Success: false,
			Error:   "页面类型不能为空",
		})
		return
	}

	var request export.ExportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, export.ExportResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取用户信息（从context中获取，如果没有则使用默认值）
	userID := GetUserID(c)
	username := GetUsername(c)

	// 执行导出
	response, err := h.exportService.Export(userID, username, pageType, &request)
	if err != nil {
		utils.Error("Export failed", err)
		c.JSON(http.StatusInternalServerError, export.ExportResponse{
			Success: false,
			Error:   "导出失败: " + err.Error(),
		})
		return
	}

	if !response.Success {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSupportedFormats 获取支持的导出格式
// @Summary 获取支持的导出格式
// @Description 获取系统支持的所有导出格式
// @Tags export
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/export/formats [get]
func (h *ExportHandler) GetSupportedFormats(c *gin.Context) {
	formats := h.exportService.GetSupportedFormats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    formats,
	})
}

// GetSupportedTypes 获取支持的数据类型
// @Summary 获取支持的数据类型
// @Description 获取系统支持的所有数据类型
// @Tags export
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/export/types [get]
func (h *ExportHandler) GetSupportedTypes(c *gin.Context) {
	types := h.exportService.GetSupportedTypes()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    types,
	})
}

// GetUserID 从context获取用户ID
func GetUserID(c *gin.Context) string {
	// 尝试从JWT token中获取
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	// 返回默认值
	return "anonymous"
}

// GetUsername 从context获取用户名
func GetUsername(c *gin.Context) string {
	// 尝试从JWT token中获取
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	// 返回默认值
	return "Anonymous"
}
