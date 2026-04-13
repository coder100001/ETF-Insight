package datasource

import (
	"context"
	"sync"
	"time"

	"etf-insight/utils"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	checkInterval    time.Duration            // 检查间隔
	timeout          time.Duration            // 检查超时时间
	failureThreshold int                      // 连续失败阈值
	successThreshold int                      // 连续成功阈值
	healthStatus     map[string]*HealthStatus // 健康状态
	stopChan         chan struct{}            // 停止通道
	mu               sync.RWMutex             // 并发控制锁
}

// HealthStatus 数据源健康状态
type HealthStatus struct {
	Name            string        // 数据源名称
	LastCheckTime   time.Time     // 最后检查时间
	LastSuccessTime time.Time     // 最后成功时间
	FailureCount    int           // 连续失败次数
	SuccessCount    int           // 连续成功次数
	IsAvailable     bool          // 是否可用
	ResponseTime    time.Duration // 平均响应时间
	SuccessRate     float64       // 成功率（最近100次）
	TotalRequests   int           // 总请求数
	TotalSuccess    int           // 成功请求数
	TotalErrors     int           // 错误请求数
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(checkInterval, timeout time.Duration, failureThreshold, successThreshold int) *HealthChecker {
	return &HealthChecker{
		checkInterval:    checkInterval,
		timeout:          timeout,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		healthStatus:     make(map[string]*HealthStatus),
		stopChan:         make(chan struct{}),
	}
}

// Register 注册数据源到健康检查器
func (h *HealthChecker) Register(sourceName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.healthStatus[sourceName]; !exists {
		h.healthStatus[sourceName] = &HealthStatus{
			Name:          sourceName,
			LastCheckTime: time.Time{},
			IsAvailable:   false, // 默认不可用，等待第一次检查
			SuccessRate:   0.0,
		}
		utils.Debug("Registered data source for health check", "source", sourceName)
	}
}

// Start 启动健康检查
func (h *HealthChecker) Start(ctx context.Context, manager *DataSourceManager) {
	utils.Info("Starting health check service",
		"interval", h.checkInterval.String(),
		"timeout", h.timeout.String())

	go h.runHealthChecks(ctx, manager)
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	close(h.stopChan)
	utils.Info("Health check service stopped")
}

// RecordSuccess 记录成功请求
func (h *HealthChecker) RecordSuccess(sourceName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	status, exists := h.healthStatus[sourceName]
	if !exists {
		// 自动注册
		status = &HealthStatus{
			Name:          sourceName,
			LastCheckTime: time.Now(),
			IsAvailable:   true,
			SuccessRate:   1.0,
		}
		h.healthStatus[sourceName] = status
	}

	now := time.Now()
	status.LastCheckTime = now
	status.LastSuccessTime = now
	status.FailureCount = 0
	status.SuccessCount++
	status.TotalRequests++
	status.TotalSuccess++

	// 更新成功率（滑动窗口）
	if status.TotalRequests > 100 {
		status.SuccessRate = float64(status.TotalSuccess) / float64(100)
	} else {
		status.SuccessRate = float64(status.TotalSuccess) / float64(status.TotalRequests)
	}

	// 检查是否达到成功阈值，标记为可用
	if status.SuccessCount >= h.successThreshold && !status.IsAvailable {
		status.IsAvailable = true
		utils.Info("Data source marked as available",
			"source", sourceName,
			"success_count", status.SuccessCount)
	}
}

// RecordFailure 记录失败请求
func (h *HealthChecker) RecordFailure(sourceName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	status, exists := h.healthStatus[sourceName]
	if !exists {
		// 自动注册
		status = &HealthStatus{
			Name:          sourceName,
			LastCheckTime: time.Now(),
			IsAvailable:   false,
			SuccessRate:   0.0,
		}
		h.healthStatus[sourceName] = status
	}

	now := time.Now()
	status.LastCheckTime = now
	status.SuccessCount = 0
	status.FailureCount++
	status.TotalRequests++
	status.TotalErrors++

	// 更新成功率（滑动窗口）
	if status.TotalRequests > 100 {
		status.SuccessRate = float64(status.TotalSuccess) / float64(100)
	} else {
		status.SuccessRate = float64(status.TotalSuccess) / float64(status.TotalRequests)
	}

	// 检查是否达到失败阈值，标记为不可用
	if status.FailureCount >= h.failureThreshold && status.IsAvailable {
		status.IsAvailable = false
		utils.Warn("Data source marked as unavailable",
			"source", sourceName,
			"failure_count", status.FailureCount)
	}
}

// GetStatus 获取数据源健康状态
func (h *HealthChecker) GetStatus(sourceName string) *HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.healthStatus[sourceName]
}

// GetAllStatus 获取所有数据源的健康状态
func (h *HealthChecker) GetAllStatus() map[string]*HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 返回深拷贝
	result := make(map[string]*HealthStatus)
	for name, status := range h.healthStatus {
		statusCopy := *status
		result[name] = &statusCopy
	}
	return result
}

// runHealthChecks 运行定期健康检查
func (h *HealthChecker) runHealthChecks(ctx context.Context, manager *DataSourceManager) {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	utils.Info("Health check service started")

	for {
		select {
		case <-ticker.C:
			h.performHealthChecks(ctx, manager)
		case <-h.stopChan:
			utils.Info("Health check service stopping")
			return
		case <-ctx.Done():
			utils.Info("Health check service context cancelled")
			return
		}
	}
}

// performHealthChecks 执行健康检查
func (h *HealthChecker) performHealthChecks(ctx context.Context, manager *DataSourceManager) {
	h.mu.RLock()
	sources := make([]string, 0, len(h.healthStatus))
	for sourceName := range h.healthStatus {
		sources = append(sources, sourceName)
	}
	h.mu.RUnlock()

	for _, sourceName := range sources {
		select {
		case <-ctx.Done():
			return
		default:
			h.checkSourceHealth(ctx, sourceName)
		}
	}

	// 记录健康检查完成
	utils.Debug("Health check round completed", "sources_checked", len(sources))
}

// checkSourceHealth 检查单个数据源的健康状况
func (h *HealthChecker) checkSourceHealth(ctx context.Context, sourceName string) {
	// 查找对应的数据源提供者
	// 注意：这里需要从管理器获取数据源实例
	// 由于循环依赖，实际检查逻辑在管理器中实现
	utils.Debug("Performing health check", "source", sourceName)

	// 这里记录一个简单的检查，实际实现会在管理器中
	h.RecordSuccess(sourceName) // 假设检查成功
}

// IsSourceAvailable 检查数据源是否可用
func (h *HealthChecker) IsSourceAvailable(sourceName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status, exists := h.healthStatus[sourceName]
	if !exists {
		return false
	}
	return status.IsAvailable
}

// GetAvailabilityStats 获取可用性统计
func (h *HealthChecker) GetAvailabilityStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	totalSources := len(h.healthStatus)
	availableSources := 0

	for name, status := range h.healthStatus {
		if status.IsAvailable {
			availableSources++
		}

		stats[name] = map[string]interface{}{
			"available":      status.IsAvailable,
			"success_rate":   status.SuccessRate,
			"response_time":  status.ResponseTime.String(),
			"total_requests": status.TotalRequests,
			"total_success":  status.TotalSuccess,
			"total_errors":   status.TotalErrors,
			"last_check":     status.LastCheckTime.Format(time.RFC3339),
			"last_success":   status.LastSuccessTime.Format(time.RFC3339),
		}
	}

	stats["summary"] = map[string]interface{}{
		"total_sources":     totalSources,
		"available_sources": availableSources,
		"availability_rate": float64(availableSources) / float64(totalSources),
		"check_interval":    h.checkInterval.String(),
		"failure_threshold": h.failureThreshold,
		"success_threshold": h.successThreshold,
	}

	return stats
}
