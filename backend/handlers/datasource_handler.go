package handlers

import (
	"context"
	"net/http"
	"time"

	"etf-insight/services/datasource/unified"

	"github.com/gin-gonic/gin"
)

type DataSourceHandler struct {
	registry *unified.UnifiedRegistry
}

func NewDataSourceHandler() *DataSourceHandler {
	return &DataSourceHandler{
		registry: unified.GetUnifiedRegistry(),
	}
}

type DataSourceProviderResponse struct {
	Success bool                     `json:"success"`
	Data    []unified.ProviderHealth `json:"data,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

type DataSourceProviderDetailResponse struct {
	Success bool                    `json:"success"`
	Data    *unified.ProviderHealth `json:"data,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// ListDataSources 列出所有已注册数据源
// @Summary 列出所有数据源
// @Description 获取所有已注册的数据源提供者列表及健康状态
// @Tags datasource
// @Produce json
// @Success 200 {object} DataSourceProviderResponse
// @Router /api/datasource/providers [get]
func (h *DataSourceHandler) ListDataSources(c *gin.Context) {
	checks := h.registry.HealthCheck(c.Request.Context())
	c.JSON(http.StatusOK, DataSourceProviderResponse{
		Success: true,
		Data:    checks,
	})
}

// GetDataSource 获取特定数据源信息
// @Summary 获取数据源详情
// @Description 根据名称获取特定数据源提供者的信息
// @Tags datasource
// @Produce json
// @Param name path string true "数据源名称"
// @Success 200 {object} DataSourceProviderDetailResponse
// @Router /api/datasource/providers/{name} [get]
func (h *DataSourceHandler) GetDataSource(c *gin.Context) {
	name := c.Param("name")
	provider, ok := h.registry.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, DataSourceProviderDetailResponse{
			Success: false,
			Error:   "数据源未找到: " + name,
		})
		return
	}

	health := provider.GetHealth(c.Request.Context())
	c.JSON(http.StatusOK, DataSourceProviderDetailResponse{
		Success: true,
		Data:    health,
	})
}

// HealthCheckDataSource 触发数据源健康检查
// @Summary 触发健康检查
// @Description 触发指定数据源的主动健康检查（5秒超时）并返回结果
// @Tags datasource
// @Produce json
// @Param name path string true "数据源名称"
// @Success 200 {object} DataSourceProviderDetailResponse
// @Router /api/datasource/providers/{name}/health [post]
func (h *DataSourceHandler) HealthCheckDataSource(c *gin.Context) {
	name := c.Param("name")
	provider, ok := h.registry.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, DataSourceProviderDetailResponse{
			Success: false,
			Error:   "数据源未找到: " + name,
		})
		return
	}

	// 使用短超时上下文触发主动探测，区别于普通的 GetHealth 读取
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := provider.GetHealth(ctx)
	c.JSON(http.StatusOK, DataSourceProviderDetailResponse{
		Success: true,
		Data:    health,
	})
}

// SetDataSourceStrategy 设置数据源选择策略
// @Summary 设置选择策略
// @Description 设置数据源提供者的选择策略（预留接口，暂未实现）
// @Tags datasource
// @Produce json
// @Success 501 {object} DataSourceProviderResponse
// @Router /api/datasource/strategy [put]
func (h *DataSourceHandler) SetDataSourceStrategy(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, DataSourceProviderResponse{
		Success: false,
		Error:   "数据源选择策略功能暂未实现",
	})
}
