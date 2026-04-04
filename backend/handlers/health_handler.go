package handlers

import (
	"net/http"
	"time"

	"etf-insight/models"

	"github.com/gin-gonic/gin"
)

// HealthStatus 服务健康状态
type HealthStatus struct {
	Status    string          `json:"status"`
	Timestamp time.Time       `json:"timestamp"`
	Version   string          `json:"version"`
	Services  map[string]bool `json:"services"`
	Uptime    string          `json:"uptime"`
}

var startTime = time.Now()

// HealthHandler 健康检查处理器 - 基础健康检查
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "ETF-Insight API is running",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ReadyHandler 就绪检查处理器 - 详细服务状态
func ReadyHandler(c *gin.Context) {
	status := HealthStatus{
		Status:    "ready",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  make(map[string]bool),
		Uptime:    time.Since(startTime).String(),
	}

	// 检查数据库连接
	if models.DB != nil {
		sqlDB, err := models.DB.DB()
		if err == nil {
			err = sqlDB.Ping()
			status.Services["database"] = err == nil
		} else {
			status.Services["database"] = false
		}
	} else {
		status.Services["database"] = false
	}

	// 检查核心服务状态
	status.Services["api"] = true
	status.Services["scheduler"] = true

	// 如果有任何服务未就绪，返回 503
	allReady := true
	for _, ready := range status.Services {
		if !ready {
			allReady = false
			break
		}
	}

	if !allReady {
		status.Status = "not_ready"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// LiveHandler 存活检查处理器
func LiveHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}
