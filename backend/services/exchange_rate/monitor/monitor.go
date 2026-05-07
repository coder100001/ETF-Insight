package monitor

import (
	"context"
	"time"

	"etf-insight/services/exchange_rate/datasource"
	"etf-insight/utils"
)

// Monitor 数据源监控器
type Monitor struct {
	manager    *datasource.DataSourceManager
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewMonitor 创建监控器
func NewMonitor(manager *datasource.DataSourceManager) *Monitor {
	return &Monitor{
		manager: manager,
	}
}

// Start 启动监控
func (m *Monitor) Start(ctx context.Context) {
	m.ctx, m.cancelFunc = context.WithCancel(ctx)
	utils.Info("数据源监控器已启动")

	// 启动健康检查
	m.manager.StartHealthCheck(m.ctx)
}

// Stop 停止监控
func (m *Monitor) Stop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	m.manager.StopHealthCheck()
	utils.Info("数据源监控器已停止")
}

// GetStatusReport 获取状态报告
func (m *Monitor) GetStatusReport() *StatusReport {
	report := &StatusReport{
		GeneratedAt:    time.Now(),
		HealthStatus:   m.manager.GetHealthStatus(),
		FailoverStats:  m.manager.GetFailoverStats(),
		ProviderStatus: datasource.GetProviderStatus(m.manager),
	}
	return report
}

// StatusReport 状态报告
type StatusReport struct {
	GeneratedAt    time.Time                           `json:"generated_at"`
	HealthStatus   map[string]*datasource.HealthStatus `json:"health_status"`
	FailoverStats  map[string]any                      `json:"failover_stats"`
	ProviderStatus map[string]any                      `json:"provider_status"`
}
