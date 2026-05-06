package datasource

import (
	"context"
	"fmt"

	"etf-insight/utils"
)

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	OpenExchangeAPIKey string `yaml:"openexchange_api_key" json:"openexchange_api_key"`
	CurrencyAPIKey     string `yaml:"currencyapi_key" json:"currencyapi_key"`
	// Frankfurter不需要API Key
}

// InitDataSourceManager 初始化数据源管理器
// 根据配置创建所有数据源并注册到管理器
func InitDataSourceManager(config *DataSourceConfig) (*DataSourceManager, error) {
	utils.Info("初始化外汇数据源管理器")

	// 处理 nil 配置
	if config == nil {
		config = &DataSourceConfig{}
	}

	// 创建主数据源
	// 优先级：Open Exchange Rates > Frankfurter > Fallback
	var primary DataSourceProvider
	if config.OpenExchangeAPIKey != "" {
		primary = NewOpenExchangeProvider(config.OpenExchangeAPIKey)
		utils.Info("主数据源已配置: Open Exchange Rates",
			"api_key", primary.GetAPIKey())
	} else {
		// OpenExchange未配置时，使用Frankfurter作为主数据源（免费，无需API Key）
		primary = NewFrankfurterProvider()
		utils.Info("主数据源已配置: Frankfurter（免费API）")
	}

	// 创建备份数据源
	var backups []DataSourceProvider

	// 第一备份：Open Exchange Rates（如果已配置为备用）
	if config.OpenExchangeAPIKey != "" {
		backup1 := NewOpenExchangeProvider(config.OpenExchangeAPIKey)
		backups = append(backups, backup1)
		utils.Info("第一备份数据源已配置: Open Exchange Rates")
	}

	// 第二备份：CurrencyAPI
	if config.CurrencyAPIKey != "" {
		backup2 := NewCurrencyAPIProvider(config.CurrencyAPIKey)
		backups = append(backups, backup2)
		utils.Info("第二备份数据源已配置: CurrencyAPI",
			"api_key", backup2.GetAPIKey())
	} else {
		utils.Warn("第二备份数据源API Key未配置，CurrencyAPI不可用")
	}

	// 第三备份：Frankfurter（如果未作为主数据源）
	if _, isFrankfurter := primary.(*FrankfurterProvider); !isFrankfurter {
		backup3 := NewFrankfurterProvider()
		backups = append(backups, backup3)
		utils.Info("第三备份数据源已配置: Frankfurter")
	}

	// 最后添加Fallback作为最终后备
	fallback := NewFallbackProvider()
	backups = append(backups, fallback)
	utils.Info("最终后备数据源已配置: Fallback")

	// 创建数据源管理器
	manager := NewDataSourceManager(primary, backups...)

	utils.Info("数据源管理器初始化完成",
		"primary", primary.GetName(),
		"backup_count", len(backups))

	return manager, nil
}

// ValidateAllProviders 验证所有数据源
func ValidateAllProviders(manager *DataSourceManager) map[string]bool {
	ctx := context.Background()
	results := make(map[string]bool)

	// 验证主数据源
	primary := manager.GetPrimaryProvider()
	if primary != nil {
		results[primary.GetName()] = primary.ValidateAPIKey(ctx)
		utils.Info("主数据源验证",
			"name", primary.GetName(),
			"valid", results[primary.GetName()])
	}

	// 验证备份数据源
	for _, backup := range manager.GetBackupProviders() {
		results[backup.GetName()] = backup.ValidateAPIKey(ctx)
		utils.Info("备份数据源验证",
			"name", backup.GetName(),
			"valid", results[backup.GetName()])
	}

	return results
}

// GetProviderStatus 获取所有数据源状态摘要
func GetProviderStatus(manager *DataSourceManager) map[string]interface{} {
	ctx := context.Background()
	status := make(map[string]interface{})

	// 主数据源状态
	primary := manager.GetPrimaryProvider()
	if primary != nil {
		status["primary"] = map[string]interface{}{
			"name":          primary.GetName(),
			"available":     primary.IsAvailable(ctx),
			"rate_limit":    primary.GetRateLimit(),
			"response_time": primary.GetResponseTime().String(),
			"success_rate":  fmt.Sprintf("%.2f%%", primary.GetSuccessRate()*100),
			"api_key":       primary.GetAPIKey(),
		}
	}

	// 备份数据源状态
	var backupStatus []map[string]interface{}
	for _, backup := range manager.GetBackupProviders() {
		backupStatus = append(backupStatus, map[string]interface{}{
			"name":          backup.GetName(),
			"available":     backup.IsAvailable(ctx),
			"rate_limit":    backup.GetRateLimit(),
			"response_time": backup.GetResponseTime().String(),
			"success_rate":  fmt.Sprintf("%.2f%%", backup.GetSuccessRate()*100),
			"api_key":       backup.GetAPIKey(),
		})
	}
	status["backups"] = backupStatus

	// 当前使用的数据源
	current := manager.GetCurrentProvider()
	status["current"] = current.GetName()

	// 故障转移统计
	status["failover_stats"] = manager.GetFailoverStats()

	return status
}
