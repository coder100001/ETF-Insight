package unified

import (
	"context"
)

type PythonDataService interface {
	GetName() string
	FetchData(ctx context.Context, endpoint string, params map[string]string) ([]byte, error)
	IsAvailable(ctx context.Context) bool
	GetRateLimit() int
}

type PythonProxyAdapter struct {
	inner PythonDataService
}

func NewPythonProxyAdapter(inner PythonDataService) *PythonProxyAdapter {
	return &PythonProxyAdapter{inner: inner}
}

func (a *PythonProxyAdapter) GetName() string {
	return a.inner.GetName()
}

func (a *PythonProxyAdapter) GetProviderType() ProviderType {
	return ProviderTypePython
}

func (a *PythonProxyAdapter) IsAvailable(ctx context.Context) bool {
	return a.inner.IsAvailable(ctx)
}

func (a *PythonProxyAdapter) GetRateLimit() int {
	return a.inner.GetRateLimit()
}

func (a *PythonProxyAdapter) GetHealth(ctx context.Context) *ProviderHealth {
	return adapterGetHealth(a.inner.GetName(), ProviderTypePython, a.inner.IsAvailable, ctx)
}
