package di

import (
	"context"

	"etf-insight/config"
	"etf-insight/services/datasource"
	"etf-insight/services/datasource/unified"
	erdatasource "etf-insight/services/exchange_rate/datasource"
	"etf-insight/utils"
)

// ProvideScheduleConfig 从 Config 提取定时任务配置
func ProvideScheduleConfig(cfg *config.Config) *config.ScheduleConfig {
	return &cfg.Schedule
}

// ProvideExchangeRateConfig 从 Config 提取汇率数据源配置
func ProvideExchangeRateConfig(cfg *config.Config) *erdatasource.DataSourceConfig {
	return &erdatasource.DataSourceConfig{
		OpenExchangeAPIKey: cfg.ExchangeRate.OpenExchangeAPIKey,
		CurrencyAPIKey:     cfg.ExchangeRate.CurrencyAPIKey,
	}
}

// ProvideDataSourceProvider 创建并选择默认 ETF 数据源
// 封装了 ProviderFactory 注册、默认选择、失败回退到 Mock 的逻辑
func ProvideDataSourceProvider() (datasource.DataSourceProvider, error) {
	ctx := context.Background()

	finageProvider := datasource.NewFinageProvider()
	utils.Info("Finage provider initialized",
		"available", finageProvider.IsAvailable(ctx))

	providerFactory := datasource.NewProviderFactory()
	providerFactory.Register("finage", finageProvider)
	providerFactory.Register("fallback", datasource.NewMockDataProvider())

	defaultProvider, err := providerFactory.GetDefault(ctx)
	if err != nil {
		utils.Warn("No data source available, using mock provider", "error", err)
		defaultProvider = datasource.NewMockDataProvider()
	} else {
		utils.Info("Using data source", "provider", defaultProvider.GetName())
	}

	// 初始化统一数据源注册表
	initUnifiedRegistry(defaultProvider)

	return defaultProvider, nil
}

// initUnifiedRegistry 将 ETF 和汇率数据源注册到统一注册表
func initUnifiedRegistry(etfProvider datasource.DataSourceProvider) {
	registry := unified.GetUnifiedRegistry()
	ctx := context.Background()

	if etfProvider != nil {
		adapter := unified.NewETFAdapter(etfProvider)
		registry.Register(etfProvider.GetName(), adapter)
		utils.Info("ETF data source registered to unified registry",
			"name", etfProvider.GetName(),
			"available", etfProvider.IsAvailable(ctx))
	}

	fxProvider := erdatasource.NewFallbackProvider()
	if fxProvider != nil {
		adapter := unified.NewFXAdapter(fxProvider)
		registry.Register(fxProvider.GetName(), adapter)
		utils.Info("FX data source registered to unified registry",
			"name", fxProvider.GetName(),
			"available", fxProvider.IsAvailable(ctx))
	}

	utils.Info("Unified data source registry initialized",
		"total_providers", registry.Count())
}
