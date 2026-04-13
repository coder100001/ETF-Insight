package datasource

import (
	"context"
	"testing"
	"time"

	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 初始化日志，避免nil pointer错误
	utils.InitLogger("warn")
}

// --- FallbackProvider 测试 ---

func TestFallbackProvider_GetName(t *testing.T) {
	provider := NewFallbackProvider()
	assert.Equal(t, "fallback", provider.GetName())
}

func TestFallbackProvider_GetBaseCurrency(t *testing.T) {
	provider := NewFallbackProvider()
	assert.Equal(t, "USD", provider.GetBaseCurrency())
}

func TestFallbackProvider_GetRate_SameCurrency(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	rate, err := provider.GetRate(ctx, "USD", "USD")
	assert.NoError(t, err)
	assert.True(t, rate.Equal(decimal.NewFromFloat(1.0)))
}

func TestFallbackProvider_GetRate_DefaultRates(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	// 测试默认汇率
	rate, err := provider.GetRate(ctx, "USD", "CNY")
	assert.NoError(t, err)
	assert.True(t, rate.GreaterThan(decimal.Zero))

	rate, err = provider.GetRate(ctx, "USD", "HKD")
	assert.NoError(t, err)
	assert.True(t, rate.GreaterThan(decimal.Zero))
}

func TestFallbackProvider_GetRate_UnsupportedCurrency(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	_, err := provider.GetRate(ctx, "XYZ", "ABC")
	assert.Error(t, err)
}

func TestFallbackProvider_GetRates(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	result, err := provider.GetRates(ctx, "USD")
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "USD", result.From)
	assert.Equal(t, "fallback", result.DataSource)
	assert.Contains(t, result.Data, "CNY")
	assert.Contains(t, result.Data, "HKD")
}

func TestFallbackProvider_IsAvailable(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	// Fallback总是可用
	assert.True(t, provider.IsAvailable(ctx))
}

func TestFallbackProvider_ValidateAPIKey(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	// Fallback不需要API Key
	assert.True(t, provider.ValidateAPIKey(ctx))
}

func TestFallbackProvider_UpdateCache(t *testing.T) {
	provider := NewFallbackProvider()
	ctx := context.Background()

	// 更新缓存
	newRate := decimal.NewFromFloat(7.5)
	provider.UpdateCache("USD", "CNY", newRate, "test")

	// 从缓存获取
	rate, err := provider.GetRate(ctx, "USD", "CNY")
	assert.NoError(t, err)
	assert.True(t, rate.Equal(newRate))
}

func TestFallbackProvider_BatchUpdateCache(t *testing.T) {
	provider := NewFallbackProvider()

	rates := map[string]decimal.Decimal{
		"CNY": decimal.NewFromFloat(7.3),
		"EUR": decimal.NewFromFloat(0.93),
	}
	provider.BatchUpdateCache("USD", rates, "test_batch")

	ctx := context.Background()
	cnyRate, err := provider.GetRate(ctx, "USD", "CNY")
	assert.NoError(t, err)
	assert.True(t, cnyRate.Equal(decimal.NewFromFloat(7.3)))
}

func TestFallbackProvider_GetCacheStats(t *testing.T) {
	provider := NewFallbackProvider()

	stats := provider.GetCacheStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_cached")
}

// --- DataSourceManager 测试 ---

func TestNewDataSourceManager(t *testing.T) {
	primary := NewFallbackProvider()
	backup1 := NewFrankfurterProvider()

	manager := NewDataSourceManager(primary, backup1)
	assert.NotNil(t, manager)
	assert.Equal(t, primary, manager.GetCurrentProvider())
	assert.Equal(t, primary, manager.GetPrimaryProvider())
	// backup1 + 内部fallback
	assert.GreaterOrEqual(t, len(manager.GetBackupProviders()), 1)
}

func TestDataSourceManager_GetRate_SameCurrency(t *testing.T) {
	primary := NewFallbackProvider()
	manager := NewDataSourceManager(primary)

	ctx := context.Background()
	rate, err := manager.GetRate(ctx, "USD", "USD")
	assert.NoError(t, err)
	assert.True(t, rate.Equal(decimal.NewFromFloat(1.0)))
}

func TestDataSourceManager_GetRate_Failover(t *testing.T) {
	// 创建一个会失败的主数据源和一个正常的备份数据源
	primary := &mockFailingProvider{name: "failing_primary"}
	backup := NewFallbackProvider()

	manager := NewDataSourceManager(primary, backup)

	ctx := context.Background()
	rate, err := manager.GetRate(ctx, "USD", "CNY")
	assert.NoError(t, err)
	assert.True(t, rate.GreaterThan(decimal.Zero))
	// 应该切换到备份数据源
	assert.Equal(t, "fallback", manager.GetCurrentProvider().GetName())
}

func TestDataSourceManager_GetFailoverStats(t *testing.T) {
	primary := NewFallbackProvider()
	manager := NewDataSourceManager(primary)

	stats := manager.GetFailoverStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "failover_count")
	assert.Contains(t, stats, "current_source")
}

func TestDataSourceManager_RestoreToPrimary(t *testing.T) {
	primary := NewFallbackProvider()
	backup := NewFallbackProvider()
	manager := NewDataSourceManager(primary, backup)

	ctx := context.Background()
	// 已经是主数据源，应该返回true
	assert.True(t, manager.RestoreToPrimary(ctx))
}

// --- HealthChecker 测试 ---

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker(
		1*time.Minute,
		10*time.Second,
		3,
		3,
	)
	assert.NotNil(t, checker)
}

func TestHealthChecker_Register(t *testing.T) {
	checker := NewHealthChecker(1*time.Minute, 10*time.Second, 3, 3)
	checker.Register("test_source")

	status := checker.GetStatus("test_source")
	assert.NotNil(t, status)
	assert.Equal(t, "test_source", status.Name)
}

func TestHealthChecker_RecordSuccess(t *testing.T) {
	checker := NewHealthChecker(1*time.Minute, 10*time.Second, 3, 3)
	checker.Register("test_source")

	checker.RecordSuccess("test_source")

	status := checker.GetStatus("test_source")
	assert.NotNil(t, status)
	assert.Equal(t, 1, status.SuccessCount)
	assert.Equal(t, 0, status.FailureCount)
}

func TestHealthChecker_RecordFailure(t *testing.T) {
	checker := NewHealthChecker(1*time.Minute, 10*time.Second, 3, 3)
	checker.Register("test_source")

	checker.RecordFailure("test_source")

	status := checker.GetStatus("test_source")
	assert.NotNil(t, status)
	assert.Equal(t, 0, status.SuccessCount)
	assert.Equal(t, 1, status.FailureCount)
}

func TestHealthChecker_FailureThreshold(t *testing.T) {
	checker := NewHealthChecker(1*time.Minute, 10*time.Second, 3, 3)
	checker.Register("test_source")

	// 先标记为可用
	status := checker.GetStatus("test_source")
	status.IsAvailable = true

	// 连续3次失败，应该标记为不可用
	for i := 0; i < 3; i++ {
		checker.RecordFailure("test_source")
	}

	status = checker.GetStatus("test_source")
	assert.False(t, status.IsAvailable)
}

func TestHealthChecker_GetAvailabilityStats(t *testing.T) {
	checker := NewHealthChecker(1*time.Minute, 10*time.Second, 3, 3)
	checker.Register("source1")
	checker.Register("source2")
	checker.RecordSuccess("source1")

	stats := checker.GetAvailabilityStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "summary")
}

// --- ProviderFactory 测试 ---

func TestProviderFactory_RegisterAndGet(t *testing.T) {
	factory := NewProviderFactory()
	provider := NewFallbackProvider()

	factory.Register("fallback", provider)

	retrieved, ok := factory.Get("fallback")
	assert.True(t, ok)
	assert.Equal(t, provider, retrieved)
}

func TestProviderFactory_ListProviders(t *testing.T) {
	factory := NewProviderFactory()
	factory.Register("fallback", NewFallbackProvider())
	factory.Register("frankfurter", NewFrankfurterProvider())

	providers := factory.ListProviders()
	assert.Len(t, providers, 2)
}

// --- Error 测试 ---

func TestProviderError_Error(t *testing.T) {
	err := &ProviderError{
		Code:    "TEST_ERROR",
		Message: "测试错误",
		Source:  "test_source",
	}

	errorStr := err.Error()
	assert.Contains(t, errorStr, "test_source")
	assert.Contains(t, errorStr, "TEST_ERROR")
}

func TestProviderError_Wrap(t *testing.T) {
	innerErr := &ProviderError{Code: "INNER", Message: "内部错误"}
	wrappedErr := Wrap(innerErr, "test_source", "WRAPPED", "包装错误")

	assert.Equal(t, "WRAPPED", wrappedErr.Code)
	assert.Equal(t, "test_source", wrappedErr.Source)
	assert.Equal(t, innerErr, wrappedErr.Unwrap())
}

func TestIsProviderError(t *testing.T) {
	providerErr := &ProviderError{Code: "TEST", Message: "test"}
	otherErr := context.Canceled

	assert.True(t, IsProviderError(providerErr))
	assert.False(t, IsProviderError(otherErr))
}

func TestIsTemporaryError(t *testing.T) {
	networkErr := &ProviderError{Code: "NETWORK_ERROR", Message: "网络错误"}
	authErr := &ProviderError{Code: "INVALID_API_KEY", Message: "API密钥无效"}
	otherErr := &ProviderError{Code: "DATA_PARSE_ERROR", Message: "数据解析错误"}

	assert.True(t, IsTemporaryError(networkErr))
	assert.False(t, IsTemporaryError(authErr))
	assert.False(t, IsTemporaryError(otherErr))
}

func TestIsFatalError(t *testing.T) {
	authErr := &ProviderError{Code: "INVALID_API_KEY", Message: "API密钥无效"}
	networkErr := &ProviderError{Code: "NETWORK_ERROR", Message: "网络错误"}

	assert.True(t, IsFatalError(authErr))
	assert.False(t, IsFatalError(networkErr))
}

// --- 辅助类型 ---

// mockFailingProvider 模拟失败的数据源
type mockFailingProvider struct {
	name string
}

func (m *mockFailingProvider) GetName() string                                          { return m.name }
func (m *mockFailingProvider) GetBaseCurrency() string                                  { return "USD" }
func (m *mockFailingProvider) GetRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	return decimal.Zero, &ProviderError{Code: "MOCK_ERROR", Message: "模拟失败", Source: m.name}
}
func (m *mockFailingProvider) GetRates(ctx context.Context, base string) (*BatchRateResult, error) {
	return &BatchRateResult{Success: false, Error: "模拟失败", DataSource: m.name},
		&ProviderError{Code: "MOCK_ERROR", Message: "模拟失败", Source: m.name}
}
func (m *mockFailingProvider) IsAvailable(ctx context.Context) bool  { return false }
func (m *mockFailingProvider) GetRateLimit() int                     { return 0 }
func (m *mockFailingProvider) GetResponseTime() time.Duration        { return 0 }
func (m *mockFailingProvider) GetSuccessRate() float64               { return 0 }
func (m *mockFailingProvider) GetAPIKey() string                     { return "" }
func (m *mockFailingProvider) GetSupportedCurrencies() []string      { return nil }
func (m *mockFailingProvider) ValidateAPIKey(ctx context.Context) bool { return false }
