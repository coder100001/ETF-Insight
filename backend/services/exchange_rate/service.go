// Package exchange_rate 提供汇率数据服务
// 包含多数据源故障转移、健康检查和数据同步功能
package exchange_rate

import (
	"context"
	"fmt"
	"time"

	"etf-insight/models"
	"etf-insight/services/exchange_rate/datasource"
	syncpkg "etf-insight/services/exchange_rate/sync"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// ExchangeRateService 汇率服务
// 集成了多数据源故障转移、健康检查和数据同步功能
type ExchangeRateService struct {
	manager      *datasource.DataSourceManager
	synchronizer *syncpkg.Synchronizer
}

// NewExchangeRateService 创建新的汇率服务
func NewExchangeRateService(config *datasource.DataSourceConfig) *ExchangeRateService {
	// 初始化数据源管理器
	manager, err := datasource.InitDataSourceManager(config)
	if err != nil {
		utils.Error("初始化数据源管理器失败", err)
		// 降级：使用Fallback
		manager = datasource.NewDataSourceManager(datasource.NewFallbackProvider())
	}

	// 初始化数据同步器
	synchronizer := syncpkg.NewSynchronizer(manager)

	service := &ExchangeRateService{
		manager:      manager,
		synchronizer: synchronizer,
	}

	utils.Info("汇率服务初始化完成",
		"primary_source", manager.GetPrimaryProvider().GetName(),
		"current_source", manager.GetCurrentProvider().GetName())

	return service
}

// GetRate 获取汇率（自动故障转移）
func (s *ExchangeRateService) GetRate(fromCurrency, toCurrency string) float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if s.manager == nil {
		return s.getDefaultRate(fromCurrency, toCurrency)
	}

	rate, err := s.manager.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		utils.Warn("获取汇率失败，使用默认值",
			"from", fromCurrency,
			"to", toCurrency,
			"error", err.Error(),
			"source", s.manager.GetCurrentProvider().GetName())
		return s.getDefaultRate(fromCurrency, toCurrency)
	}

	return rate.InexactFloat64()
}

// GetRateDecimal 获取汇率（decimal.Decimal类型，无精度损失）
// 推荐在需要高精度计算时使用此方法
func (s *ExchangeRateService) GetRateDecimal(fromCurrency, toCurrency string) (decimal.Decimal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if s.manager == nil {
		return decimal.Zero, fmt.Errorf("exchange rate manager is nil")
	}

	return s.manager.GetRate(ctx, fromCurrency, toCurrency)
}

// Convert 货币转换
func (s *ExchangeRateService) Convert(amount decimal.Decimal, fromCurrency, toCurrency string) decimal.Decimal {
	if fromCurrency == toCurrency {
		return amount
	}

	rate := s.GetRate(fromCurrency, toCurrency)
	return amount.Mul(decimal.NewFromFloat(rate))
}

// UpdateRates 更新汇率（全量同步）
func (s *ExchangeRateService) UpdateRates() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	utils.Info("开始更新汇率...")

	result, err := s.synchronizer.SyncAllRates(ctx)
	if err != nil {
		utils.Error("汇率更新失败", err)
		return err
	}

	utils.Info("汇率更新完成",
		"status", result.Status,
		"total", result.TotalCount,
		"success", result.SuccessCount,
		"failed", result.FailedCount,
		"source", result.DataSource,
		"duration_ms", result.DurationMs)

	if result.Status == "failed" {
		return fmt.Errorf("汇率更新失败: %d个货币对同步失败", result.FailedCount)
	}

	return nil
}

// SyncFromPrimary 从主数据源同步数据（恢复后同步）
func (s *ExchangeRateService) SyncFromPrimary() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := s.synchronizer.SyncFromPrimary(ctx)
	if err != nil {
		utils.Error("从主数据源同步失败", err)
		return err
	}

	utils.Info("主数据源同步完成",
		"status", result.Status,
		"total", result.TotalCount,
		"success", result.SuccessCount)

	return nil
}

// ConsistencyCheck 数据一致性检查
func (s *ExchangeRateService) ConsistencyCheck() (*syncpkg.ConsistencyReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return s.synchronizer.ConsistencyCheck(ctx)
}

// GetHistory 获取汇率历史
func (s *ExchangeRateService) GetHistory(fromCurrency, toCurrency string, days int) ([]map[string]any, error) {
	// 从数据库获取历史汇率数据
	var rates []models.ExchangeRate
	if models.DB != nil {
		result := models.DB.Where(
			"from_currency = ? AND to_currency = ?",
			fromCurrency, toCurrency,
		).Order("updated_at DESC").Limit(days).Find(&rates)

		if result.Error != nil {
			return nil, result.Error
		}
	}

	history := make([]map[string]any, 0, len(rates))
	for _, rate := range rates {
		history = append(history, map[string]any{
			"date": rate.UpdatedAt.Format("2006-01-02"),
			"rate": rate.Rate.InexactFloat64(),
		})
	}

	return history, nil
}

// CalculateCrossRate 计算交叉汇率
func (s *ExchangeRateService) CalculateCrossRate(fromCurrency, toCurrency string) float64 {
	if fromCurrency == toCurrency {
		return 1.0
	}

	// 尝试直接获取
	directRate := s.GetRate(fromCurrency, toCurrency)
	if directRate != 1.0 {
		return directRate
	}

	// 通过USD计算交叉汇率
	fromToUSD := s.GetRate(fromCurrency, "USD")
	usdToTarget := s.GetRate("USD", toCurrency)

	return fromToUSD * usdToTarget
}

// GetDataSourceStatus 获取数据源状态
func (s *ExchangeRateService) GetDataSourceStatus() map[string]any {
	return datasource.GetProviderStatus(s.manager)
}

// GetFailoverStats 获取故障转移统计
func (s *ExchangeRateService) GetFailoverStats() map[string]any {
	return s.manager.GetFailoverStats()
}

// RestoreToPrimary 尝试恢复到主数据源
func (s *ExchangeRateService) RestoreToPrimary() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return s.manager.RestoreToPrimary(ctx)
}

// ValidateAllProviders 验证所有数据源
func (s *ExchangeRateService) ValidateAllProviders() map[string]bool {
	return datasource.ValidateAllProviders(s.manager)
}

// getDefaultRate 获取默认汇率（所有数据源都不可用时的后备值）
func (s *ExchangeRateService) getDefaultRate(fromCurrency, toCurrency string) float64 {
	// 先尝试从数据库获取（DB 未初始化时跳过，避免 nil panic）
	if models.DB != nil {
		var rate models.ExchangeRate
		result := models.DB.Where(
			"from_currency = ? AND to_currency = ?",
			fromCurrency, toCurrency,
		).Order("updated_at DESC").First(&rate)

		if result.Error == nil {
			return rate.Rate.InexactFloat64()
		}
	}

	// 硬编码默认汇率
	defaultRates := map[string]map[string]float64{
		"USD": {"CNY": 7.2, "HKD": 7.8, "EUR": 0.92, "GBP": 0.79, "JPY": 150.0},
		"CNY": {"USD": 0.1389, "HKD": 1.0833},
		"HKD": {"USD": 0.1282, "CNY": 0.9231},
	}

	if fromRates, ok := defaultRates[fromCurrency]; ok {
		if rate, ok := fromRates[toCurrency]; ok {
			return rate
		}
	}

	return 1.0
}
