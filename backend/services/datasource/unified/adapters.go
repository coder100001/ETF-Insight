package unified

import (
	"context"

	"etf-insight/services/datasource"
	erdatasource "etf-insight/services/exchange_rate/datasource"
)

type ETFAdapter struct {
	inner datasource.DataSourceProvider
}

func NewETFAdapter(inner datasource.DataSourceProvider) *ETFAdapter {
	return &ETFAdapter{inner: inner}
}

func (a *ETFAdapter) GetName() string {
	return a.inner.GetName()
}

func (a *ETFAdapter) GetProviderType() ProviderType {
	return ProviderTypeETF
}

func (a *ETFAdapter) IsAvailable(ctx context.Context) bool {
	return a.inner.IsAvailable(ctx)
}

func (a *ETFAdapter) GetRateLimit() int {
	return a.inner.GetRateLimit()
}

func (a *ETFAdapter) GetHealth(ctx context.Context) *ProviderHealth {
	return adapterGetHealth(a.inner.GetName(), ProviderTypeETF, a.inner.IsAvailable, ctx)
}

type FXAdapter struct {
	inner erdatasource.DataSourceProvider
}

func NewFXAdapter(inner erdatasource.DataSourceProvider) *FXAdapter {
	return &FXAdapter{inner: inner}
}

func (a *FXAdapter) GetName() string {
	return a.inner.GetName()
}

func (a *FXAdapter) GetProviderType() ProviderType {
	return ProviderTypeFX
}

func (a *FXAdapter) IsAvailable(ctx context.Context) bool {
	return a.inner.IsAvailable(ctx)
}

func (a *FXAdapter) GetRateLimit() int {
	return a.inner.GetRateLimit()
}

func (a *FXAdapter) GetHealth(ctx context.Context) *ProviderHealth {
	return adapterGetHealth(a.inner.GetName(), ProviderTypeFX, a.inner.IsAvailable, ctx)
}
