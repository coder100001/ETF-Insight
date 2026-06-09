package unified

import (
	"context"

	"etf-insight/services/datasource"
)

// AssetProviderAdapter 将新版 AssetProvider 适配为 UnifiedDataProvider
type AssetProviderAdapter struct {
	inner datasource.AssetProvider
}

// NewAssetProviderAdapter 创建 AssetProvider 适配器
func NewAssetProviderAdapter(inner datasource.AssetProvider) *AssetProviderAdapter {
	return &AssetProviderAdapter{inner: inner}
}

func (a *AssetProviderAdapter) GetName() string {
	return a.inner.Name()
}

func (a *AssetProviderAdapter) GetProviderType() ProviderType {
	return ProviderTypeETF
}

func (a *AssetProviderAdapter) IsAvailable(ctx context.Context) bool {
	return a.inner.IsAvailable(ctx)
}

func (a *AssetProviderAdapter) GetRateLimit() int {
	return a.inner.RateLimit()
}

func (a *AssetProviderAdapter) GetHealth(ctx context.Context) *ProviderHealth {
	return adapterGetHealth(a.inner.Name(), ProviderTypeETF, a.inner.IsAvailable, ctx)
}

// GetAssetProvider 获取底层的 AssetProvider
func (a *AssetProviderAdapter) GetAssetProvider() datasource.AssetProvider {
	return a.inner
}
