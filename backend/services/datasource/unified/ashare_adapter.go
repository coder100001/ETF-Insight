package unified

import (
	"context"
	"time"
)

// adapterGetHealth 通用的 adapter 健康检查辅助函数
func adapterGetHealth(name string, pt ProviderType, isAvailable func(ctx context.Context) bool, ctx context.Context) *ProviderHealth {
	start := time.Now()
	available := isAvailable(ctx)
	latency := time.Since(start).Milliseconds()

	return &ProviderHealth{
		Name:      name,
		Type:      pt,
		TypeStr:   pt.String(),
		Available: available,
		LatencyMs: latency,
		LastCheck: time.Now(),
	}
}

type AShareProvider interface {
	GetName() string
	FetchETFPrice(ctx context.Context, symbol string) (float64, error)
	IsAvailable(ctx context.Context) bool
	GetRateLimit() int
}

type AShareAdapter struct {
	inner AShareProvider
}

func NewAShareAdapter(inner AShareProvider) *AShareAdapter {
	return &AShareAdapter{inner: inner}
}

func (a *AShareAdapter) GetName() string {
	return a.inner.GetName()
}

func (a *AShareAdapter) GetProviderType() ProviderType {
	return ProviderTypeAShare
}

func (a *AShareAdapter) IsAvailable(ctx context.Context) bool {
	return a.inner.IsAvailable(ctx)
}

func (a *AShareAdapter) GetRateLimit() int {
	return a.inner.GetRateLimit()
}

func (a *AShareAdapter) GetHealth(ctx context.Context) *ProviderHealth {
	return adapterGetHealth(a.inner.GetName(), ProviderTypeAShare, a.inner.IsAvailable, ctx)
}
