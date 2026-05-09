package models

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/shopspring/decimal"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitDB 初始化数据库连接并自动迁移
func InitDB(dsn string) error {
	var err error
	var dialector gorm.Dialector

	// 根据 DSN 判断使用哪种数据库
	if dsn == "" || dsn == "etf_insight.db" || dsn == ":memory:" {
		// 使用 SQLite
		if dsn == "" {
			dsn = "etf_insight.db"
		}
		log.Printf("Using SQLite database: %s", dsn)
		dialector = sqlite.Open(dsn)
	} else {
		// 使用 PostgreSQL
		log.Printf("Using PostgreSQL database")
		dialector = postgres.Open(dsn)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return AutoMigrate()
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	return DB.AutoMigrate(
		// 原有模型
		&ETFConfig{},
		&ETFData{},
		&OperationLog{},
		&AuditLog{},
		&ExchangeRate{},
		&AShareDividendETF{},
		&AShareETFPortfolio{},
		&ASharePortfolioHolding{},
		&PortfolioConfig{},
		&UniversalETF{},
		&ETFClassification{},
		&ETFConstituent{},
		&ETFHistoricalData{},
		&ETFDividend{},
		&ETFAssetAllocation{},
		// 新增统一数据层模型
		&Asset{},
		&AssetMetadata{},
		&Holding{},
		&HoldingSnapshot{},
		&ETFHoldingsReport{},
		&HoldingChange{},
		&Price{},
		&Portfolio{},
		&PortfolioPosition{},
		// 新增因子数据层模型
		&FactorData{},
		&FactorTimingSignal{},
		// 新增Alpha观点层模型
		&AlphaView{},
		&AlphaViewPerformance{},
		&BlackLittermanConfig{},
		&BLPosteriorReturn{},
		// 新增风险预算层模型
		&RiskBudgetConfig{},
		&MonteCarloSimulation{},
		&RiskContribution{},
		&RiskBudgetExecution{},
		// 新增插件架构层模型
		&PluginRegistry{},
		&PluginConfiguration{},
		&PluginExecutionLog{},
		&ModelBenchmarkMatrix{},
		&StrategyExperiment{},
		// 新增报告系统模型
		&ReportTemplate{},
		&ReportSection{},
		&GeneratedReport{},
		&ReportParameter{},
		// 补充遗漏的模型
		&AssetPrice{},
		&AssetRelationship{},
		&SectorAllocation{},
		&GeographicAllocation{},
		&ETFHoldingSummary{},
		&PortfolioOverlap{},
		&PortfolioPerformance{},
		&PortfolioRebalance{},
		&PriceGap{},
		&PriceStats{},
		&ETFOverlapCache{},
		&CacheInvalidationLog{},
		&ExchangeRateSyncLog{},
	)
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	defaultETFs := []ETFConfig{
		{
			Symbol:       "QQQ",
			Name:         "Invesco QQQ Trust",
			Description:  "追踪纳斯达克100指数",
			Strategy:     "大盘成长",
			Focus:        "科技",
			ExpenseRatio: decimal.NewFromFloat(0.0020),
			Currency:     "USD",
			Exchange:     "NASDAQ",
			Category:     "大盘股",
			Provider:     "Invesco",
			Status:       1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Symbol:       "SCHD",
			Name:         "Schwab US Dividend Equity ETF",
			Description:  "美国股息股票ETF",
			Strategy:     "股息价值",
			Focus:        "高股息",
			ExpenseRatio: decimal.NewFromFloat(0.0006),
			Currency:     "USD",
			Exchange:     "NYSE",
			Category:     "股息",
			Provider:     "Charles Schwab",
			Status:       1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, etf := range defaultETFs {
		result := DB.Where("symbol = ?", etf.Symbol).First(&ETFConfig{})
		if result.Error == gorm.ErrRecordNotFound {
			if err := DB.Create(&etf).Error; err != nil {
				log.Printf("Failed to create default ETF %s: %v", etf.Symbol, err)
			}
		} else {
			// 记录已存在，更新关键字段（如 expense_ratio）
			if err := DB.Model(&ETFConfig{}).Where("symbol = ?", etf.Symbol).Updates(map[string]any{
				"expense_ratio": etf.ExpenseRatio,
				"name":          etf.Name,
				"description":   etf.Description,
				"strategy":      etf.Strategy,
				"focus":         etf.Focus,
				"category":      etf.Category,
				"provider":      etf.Provider,
				"updated_at":    time.Now(),
			}).Error; err != nil {
				log.Printf("Failed to update ETF %s: %v", etf.Symbol, err)
			}
		}
	}

	// 初始化默认投资组合配置
	if err := initDefaultPortfolioConfigs(); err != nil {
		log.Printf("Failed to init default portfolio configs: %v", err)
	}

	return nil
}

// initDefaultPortfolioConfigs 初始化默认投资组合配置
func initDefaultPortfolioConfigs() error {
	// 检查是否已有配置
	var count int64
	DB.Model(&PortfolioConfig{}).Count(&count)
	if count > 0 {
		return nil // 已有配置，跳过
	}

	defaultConfigs := []PortfolioConfig{
		{
			Name:            "稳健型组合",
			Description:     "低风险稳健投资，适合保守型投资者",
			Allocation:      `{"BND": 40, "VTI": 30, "VXUS": 20, "SCHD": 10}`,
			TotalInvestment: decimal.NewFromInt(100000),
			TaxRate:         decimal.NewFromFloat(0.10),
			Status:          1,
			IsDefault:       true,
		},
		{
			Name:            "平衡型组合",
			Description:     "股债平衡，适合中长期投资者",
			Allocation:      `{"VTI": 40, "QQQ": 20, "BND": 20, "VXUS": 20}`,
			TotalInvestment: decimal.NewFromInt(100000),
			TaxRate:         decimal.NewFromFloat(0.10),
			Status:          1,
			IsDefault:       false,
		},
		{
			Name:            "成长型组合",
			Description:     "高增长潜力，适合激进型投资者",
			Allocation:      `{"QQQ": 50, "VGT": 30, "ARKK": 20}`,
			TotalInvestment: decimal.NewFromInt(100000),
			TaxRate:         decimal.NewFromFloat(0.10),
			Status:          1,
			IsDefault:       false,
		},
	}

	for _, config := range defaultConfigs {
		if err := DB.Create(&config).Error; err != nil {
			log.Printf("Failed to create default portfolio config %s: %v", config.Name, err)
		}
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// IsSQLite 检查是否使用 SQLite
func IsSQLite() bool {
	// 简单检查：如果环境变量 DB_DSN 包含 .db 或为空，则认为是 SQLite
	dsn := os.Getenv("DB_DSN")
	return dsn == "" || dsn == "etf_insight.db" || dsn == ":memory:"
}
