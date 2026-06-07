package config

import "sync"

// FinancialConfig 统一金融常量配置
type FinancialConfig struct {
	RiskFreeRate    float64
	TradingDaysYear int
	DefaultCurrency string
}

var (
	financialConfig *FinancialConfig
	configOnce      sync.Once
	configMu        sync.RWMutex
)

// ensureInitialized 确保配置已初始化 (调用方不持锁)
func ensureInitialized() {
	configOnce.Do(func() {
		financialConfig = &FinancialConfig{
			RiskFreeRate:    0.0435,
			TradingDaysYear: 252,
			DefaultCurrency: "USD",
		}
	})
}

// GetFinancialConfig 获取金融配置单例 (读锁保护)
func GetFinancialConfig() *FinancialConfig {
	ensureInitialized()
	configMu.RLock()
	defer configMu.RUnlock()
	return financialConfig
}

// SetRiskFreeRate 设置无风险利率 (写锁保护)
func SetRiskFreeRate(rate float64) {
	ensureInitialized()
	configMu.Lock()
	defer configMu.Unlock()
	financialConfig.RiskFreeRate = rate
}

// SetTradingDaysYear 设置年交易日数 (写锁保护)
func SetTradingDaysYear(days int) {
	ensureInitialized()
	configMu.Lock()
	defer configMu.Unlock()
	financialConfig.TradingDaysYear = days
}
