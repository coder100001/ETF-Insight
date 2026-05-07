package datasource

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ProviderRegistry 数据源注册表
// 全局单例，负责管理所有已注册的数据源提供者
type ProviderRegistry struct {
	providers map[string]DataSourceProvider
	mu        sync.RWMutex
}

var (
	globalRegistry *ProviderRegistry
	once           sync.Once
)

// GetRegistry 获取全局注册表实例
func GetRegistry() *ProviderRegistry {
	once.Do(func() {
		globalRegistry = &ProviderRegistry{
			providers: make(map[string]DataSourceProvider),
		}
	})
	return globalRegistry
}

// RegisterProvider 注册数据源提供者
func (r *ProviderRegistry) RegisterProvider(name string, provider DataSourceProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
	fmt.Printf("Registered provider: %s\n", name)
}

// UnregisterProvider 取消注册数据源提供者
func (r *ProviderRegistry) UnregisterProvider(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
	fmt.Printf("Unregistered provider: %s\n", name)
}

// GetProvider 获取指定名称的数据源
func (r *ProviderRegistry) GetProvider(name string) (DataSourceProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}

// ListProviders 列出所有已注册的提供者
func (r *ProviderRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetProviderPriority 获取提供者优先级（数值越小优先级越高）
func (r *ProviderRegistry) GetProviderPriority(name string) int {
	provider, ok := r.GetProvider(name)
	if !ok {
		return 100 // 最低优先级
	}
	// 优先级基于速率限制，低速率限制（更宽松）的优先级更高
	// 如果速率限制是10请求/秒，优先级是10
	// 如果速率限制是100请求/秒，优先级是100
	return provider.GetRateLimit()
}

// SelectBestProvider 选择最佳数据源
func (r *ProviderRegistry) SelectBestProvider(ctx context.Context, strategy ProviderSelectionStrategy) (DataSourceProvider, error) {
	switch strategy {
	case PriorityStrategy:
		return r.selectByPriority(ctx)
	case WeightedRoundRobinStrategy:
		return r.selectByWeightedRoundRobin(ctx)
	case LeastLoadedStrategy:
		return r.selectByLeastLoaded(ctx)
	default:
		return r.selectByPriority(ctx)
	}
}

// ProviderSelectionStrategy 数据源选择策略
type ProviderSelectionStrategy int

const (
	PriorityStrategy           ProviderSelectionStrategy = iota // 按优先级选择
	WeightedRoundRobinStrategy                                  // 加权轮询
	LeastLoadedStrategy                                         // 最少负载
)

// selectByPriority 按优先级选择数据源
func (r *ProviderRegistry) selectByPriority(ctx context.Context) (DataSourceProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 获取所有可用数据源
	var providers []DataSourceProvider
	for _, provider := range r.providers {
		if provider.IsAvailable(ctx) {
			providers = append(providers, provider)
		}
	}

	if len(providers) == 0 {
		return nil, ErrNoAvailableProvider
	}

	// 按照优先级排序（优先级数字越小优先级越高）
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].GetRateLimit() < providers[j].GetRateLimit()
	})

	return providers[0], nil
}

// weightedProvider 加权提供者
type weightedProvider struct {
	provider DataSourceProvider
	weight   int
}

// selectByWeightedRoundRobin 加权轮询选择数据源
func (r *ProviderRegistry) selectByWeightedRoundRobin(ctx context.Context) (DataSourceProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 计算加权提供者列表
	var providers []weightedProvider
	for _, provider := range r.providers {
		if provider.IsAvailable(ctx) {
			// 权重基于速率限制，速率限制越高权重越大
			weight := provider.GetRateLimit()
			providers = append(providers, weightedProvider{
				provider: provider,
				weight:   weight,
			})
		}
	}

	if len(providers) == 0 {
		return nil, ErrNoAvailableProvider
	}

	// 简单的加权轮询实现
	// TODO: 实现更复杂的加权轮询算法
	return providers[0].provider, nil
}

// selectByLeastLoaded 最少负载选择数据源
func (r *ProviderRegistry) selectByLeastLoaded(ctx context.Context) (DataSourceProvider, error) {
	// 这个实现需要额外的负载监控功能
	// 目前使用优先级选择
	return r.selectByPriority(ctx)
}

// ProviderHealthCheck 提供者健康检查
type ProviderHealthCheck struct {
	Name        string    `json:"name"`
	Available   bool      `json:"available"`
	LatencyMs   int64     `json:"latency_ms,omitempty"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
}

// HealthCheck 执行健康检查
func (r *ProviderRegistry) HealthCheck(ctx context.Context) []ProviderHealthCheck {
	r.mu.RLock()
	providers := make([]DataSourceProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	r.mu.RUnlock()

	checks := make([]ProviderHealthCheck, 0, len(providers))
	for _, provider := range providers {
		start := time.Now()
		available := provider.IsAvailable(ctx)
		latency := time.Since(start).Milliseconds()
		check := ProviderHealthCheck{
			Name:        provider.GetName(),
			Available:   available,
			LatencyMs:   latency,
			LastChecked: time.Now(),
		}
		checks = append(checks, check)
	}
	return checks
}

// IsProviderAvailable 检查指定提供者是否可用
func (r *ProviderRegistry) IsProviderAvailable(ctx context.Context, name string) bool {
	provider, ok := r.GetProvider(name)
	if !ok {
		return false
	}
	return provider.IsAvailable(ctx)
}

// GetDefaultProvider 获取默认数据源
func (r *ProviderRegistry) GetDefaultProvider(ctx context.Context) (DataSourceProvider, error) {
	return r.SelectBestProvider(ctx, PriorityStrategy)
}

// GetProviderInfo 获取提供者信息
func (r *ProviderRegistry) GetProviderInfo(name string) map[string]any {
	provider, ok := r.GetProvider(name)
	if !ok {
		return nil
	}

	info := make(map[string]any)
	info["name"] = provider.GetName()
	info["rate_limit"] = provider.GetRateLimit()
	info["priority"] = r.GetProviderPriority(name)
	return info
}
