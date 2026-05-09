package unified

import (
	"context"
	"sort"
	"sync"
	"time"
)

type ProviderType int

const (
	ProviderTypeETF ProviderType = iota
	ProviderTypeFX
	ProviderTypeAShare
	ProviderTypePython
)

func (t ProviderType) String() string {
	switch t {
	case ProviderTypeETF:
		return "etf"
	case ProviderTypeFX:
		return "fx"
	case ProviderTypeAShare:
		return "ashare"
	case ProviderTypePython:
		return "python"
	default:
		return "unknown"
	}
}

type ProviderHealth struct {
	Name      string       `json:"name"`
	Type      ProviderType `json:"type"`
	TypeStr   string       `json:"type_str"`
	Available bool         `json:"available"`
	LatencyMs int64        `json:"latency_ms,omitempty"`
	LastCheck time.Time    `json:"last_check"`
	Error     string       `json:"error,omitempty"`
}

type UnifiedDataProvider interface {
	GetName() string
	GetProviderType() ProviderType
	IsAvailable(ctx context.Context) bool
	GetRateLimit() int
	GetHealth(ctx context.Context) *ProviderHealth
}

type UnifiedRegistry struct {
	providers map[string]UnifiedDataProvider
	mu        sync.RWMutex
}

var (
	globalUnifiedRegistry *UnifiedRegistry
	unifiedOnce           sync.Once
)

func GetUnifiedRegistry() *UnifiedRegistry {
	unifiedOnce.Do(func() {
		globalUnifiedRegistry = &UnifiedRegistry{
			providers: make(map[string]UnifiedDataProvider),
		}
	})
	return globalUnifiedRegistry
}

func (r *UnifiedRegistry) Register(name string, provider UnifiedDataProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

func (r *UnifiedRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}

func (r *UnifiedRegistry) Get(name string) (UnifiedDataProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

func (r *UnifiedRegistry) List() []UnifiedDataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]UnifiedDataProvider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].GetName() < result[j].GetName()
	})
	return result
}

func (r *UnifiedRegistry) GetByType(t ProviderType) []UnifiedDataProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []UnifiedDataProvider
	for _, p := range r.providers {
		if p.GetProviderType() == t {
			result = append(result, p)
		}
	}
	return result
}

func (r *UnifiedRegistry) HealthCheck(ctx context.Context) []ProviderHealth {
	r.mu.RLock()
	providers := make([]UnifiedDataProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	r.mu.RUnlock()

	checks := make([]ProviderHealth, 0, len(providers))
	for _, p := range providers {
		health := p.GetHealth(ctx)
		checks = append(checks, *health)
	}
	return checks
}

func (r *UnifiedRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// ResetForTesting 清空所有注册的 provider，仅供测试使用
func (r *UnifiedRegistry) ResetForTesting() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = make(map[string]UnifiedDataProvider)
}
