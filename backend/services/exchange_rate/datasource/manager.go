package datasource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// DataSourceManager 数据源管理器
// 负责管理多个数据源的故障转移和健康检查
type DataSourceManager struct {
	primary       DataSourceProvider   // 主数据源
	backups       []DataSourceProvider // 备份数据源列表
	current       DataSourceProvider   // 当前使用的数据源
	healthCheck   *HealthChecker       // 健康检查器
	mu            sync.RWMutex         // 并发控制锁
	lastFailover  time.Time            // 最后一次故障转移时间
	failoverCount int                  // 故障转移次数
}

// NewDataSourceManager 创建新的数据源管理器
func NewDataSourceManager(primary DataSourceProvider, backups ...DataSourceProvider) *DataSourceManager {
	manager := &DataSourceManager{
		primary:       primary,
		backups:       backups,
		current:       primary,
		lastFailover:  time.Time{},
		failoverCount: 0,
	}

	// 创建健康检查器
	manager.healthCheck = NewHealthChecker(
		1*time.Minute,  // 每分钟检查一次
		10*time.Second, // 10秒超时
		3,              // 连续3次失败标记为不可用
		3,              // 连续3次成功标记为可用
	)

	// 初始化健康状态
	manager.initializeHealthStatus()

	return manager
}

// GetCurrentProvider 获取当前使用的数据源
func (m *DataSourceManager) GetCurrentProvider() DataSourceProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// GetPrimaryProvider 获取主数据源
func (m *DataSourceManager) GetPrimaryProvider() DataSourceProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primary
}

// GetBackupProviders 获取备份数据源列表
func (m *DataSourceManager) GetBackupProviders() []DataSourceProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.backups
}

// GetRate 获取汇率（自动故障转移）
func (m *DataSourceManager) GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	// 获取当前数据源（避免竞态条件）
	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	// 尝试使用当前数据源
	rate, err := current.GetRate(ctx, fromCurrency, toCurrency)
	if err == nil {
		// 记录成功请求
		m.healthCheck.RecordSuccess(current.GetName())
		return rate, nil
	}

	// 当前数据源失败，尝试故障转移
	utils.Warn("Current data source failed, attempting failover",
		"source", current.GetName(),
		"error", err.Error(),
		"currencies", fmt.Sprintf("%s->%s", fromCurrency, toCurrency))

	// 尝试切换到下一个可用的数据源
	if err := m.performFailover(ctx, err); err != nil {
		// 所有数据源都不可用
		return decimal.Zero, err
	}

	// 重试使用新的数据源
	return m.current.GetRate(ctx, fromCurrency, toCurrency)
}

// GetRates 批量获取汇率（自动故障转移）
func (m *DataSourceManager) GetRates(ctx context.Context, baseCurrency string) (*BatchRateResult, error) {
	// 获取当前数据源（避免竞态条件）
	m.mu.RLock()
	current := m.current
	m.mu.RUnlock()

	// 尝试使用当前数据源
	result, err := current.GetRates(ctx, baseCurrency)
	if err == nil && result.Success {
		// 记录成功请求
		m.healthCheck.RecordSuccess(current.GetName())
		return result, nil
	}

	// 当前数据源失败，尝试故障转移
	utils.Warn("Current data source failed for batch rates, attempting failover",
		"source", current.GetName(),
		"error", err.Error(),
		"baseCurrency", baseCurrency)

	// 尝试切换到下一个可用的数据源
	if err := m.performFailover(ctx, err); err != nil {
		// 所有数据源都不可用
		return &BatchRateResult{
			Success:    false,
			Error:      err.Error(),
			DataSource: "none",
		}, err
	}

	// 重试使用新的数据源
	return m.current.GetRates(ctx, baseCurrency)
}

// performFailover 执行故障转移
func (m *DataSourceManager) performFailover(ctx context.Context, originalError error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 记录故障转移
	m.lastFailover = time.Now()
	m.failoverCount++

	// 标记当前数据源为不可用
	m.healthCheck.RecordFailure(m.current.GetName())

	// 查找下一个可用的数据源
	var newProvider DataSourceProvider
	var failoverReason string

	// 优先检查主数据源是否可用
	if m.current != m.primary && m.primary.IsAvailable(ctx) {
		newProvider = m.primary
		failoverReason = "fallback_to_primary"
		utils.Info("Failover: Switching back to primary data source",
			"from", m.current.GetName(),
			"to", newProvider.GetName())
	} else {
		// 检查备份数据源
		for _, backup := range m.backups {
			if backup != m.current && backup.IsAvailable(ctx) {
				newProvider = backup
				failoverReason = "failover_to_backup"
				utils.Info("Failover: Switching to backup data source",
					"from", m.current.GetName(),
					"to", newProvider.GetName(),
					"reason", originalError.Error())
				break
			}
		}
	}

	if newProvider == nil {
		// 所有数据源都不可用
		utils.Error("All data sources are unavailable", originalError,
			"current", m.current.GetName())

		return &ProviderError{
			Code:      "ALL_SOURCES_UNAVAILABLE",
			Message:   "所有数据源都不可用",
			Source:    m.current.GetName(),
			InnerErr:  originalError,
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// 切换到新的数据源
	oldSource := m.current.GetName()
	m.current = newProvider

	// 记录故障转移事件
	utils.Info("Failover completed",
		"old_source", oldSource,
		"new_source", newProvider.GetName(),
		"reason", failoverReason,
		"failover_count", m.failoverCount)

	return nil
}

// RestoreToPrimary 尝试恢复到主数据源
func (m *DataSourceManager) RestoreToPrimary(ctx context.Context) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == m.primary {
		// 已经使用主数据源
		return true
	}

	if !m.primary.IsAvailable(ctx) {
		// 主数据源不可用
		return false
	}

	// 切换到主数据源
	oldSource := m.current.GetName()
	m.current = m.primary

	utils.Info("Restored to primary data source",
		"old_source", oldSource,
		"new_source", m.primary.GetName())

	return true
}

// GetHealthStatus 获取健康状态报告
func (m *DataSourceManager) GetHealthStatus() map[string]*HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]*HealthStatus)

	// 主数据源状态
	primaryStatus := m.healthCheck.GetStatus(m.primary.GetName())
	if primaryStatus != nil {
		status[m.primary.GetName()] = primaryStatus
	}

	// 备份数据源状态
	for _, backup := range m.backups {
		backupStatus := m.healthCheck.GetStatus(backup.GetName())
		if backupStatus != nil {
			status[backup.GetName()] = backupStatus
		}
	}

	// 当前数据源状态
	currentStatus := m.healthCheck.GetStatus(m.current.GetName())
	if currentStatus != nil {
		status["current"] = currentStatus
	}

	// 系统状态
	status["system"] = &HealthStatus{
		Name:          "system",
		LastCheckTime: time.Now(),
		IsAvailable:   m.current.IsAvailable(context.Background()),
		ResponseTime:  m.current.GetResponseTime(),
		SuccessRate:   m.current.GetSuccessRate(),
	}

	return status
}

// StartHealthCheck 启动健康检查
func (m *DataSourceManager) StartHealthCheck(ctx context.Context) {
	m.healthCheck.Start(ctx, m)
}

// StopHealthCheck 停止健康检查
func (m *DataSourceManager) StopHealthCheck() {
	m.healthCheck.Stop()
}

// GetFailoverStats 获取故障转移统计
func (m *DataSourceManager) GetFailoverStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"last_failover_time": m.lastFailover,
		"failover_count":     m.failoverCount,
		"current_source":     m.current.GetName(),
		"primary_source":     m.primary.GetName(),
		"backup_count":       len(m.backups),
	}
}

// initializeHealthStatus 初始化健康状态
func (m *DataSourceManager) initializeHealthStatus() {
	// 注册所有数据源到健康检查器
	m.healthCheck.Register(m.primary.GetName())

	for _, backup := range m.backups {
		m.healthCheck.Register(backup.GetName())
	}
}
