package models

import (
	"fmt"
	"time"

	"etf-insight/config"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) error {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.GetDSN())
	case "sqlite":
		dialector = sqlite.Open(cfg.GetDSN())
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	logrus.Info("Database connected successfully")
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	return DB.AutoMigrate(
		&Workflow{},
		&WorkflowStep{},
		&WorkflowInstance{},
		&WorkflowInstanceStep{},
		&ETFConfig{},
		&ETFData{},
		&ETFBaseInfo{},
		&ETFPrice{},
		&ETFDividend{},
		&PortfolioConfig{},
		&ExchangeRate{},
		&OperationLog{},
		&SystemLog{},
		&Notification{},
		&AnalysisReport{},
	)
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 初始化默认ETF配置
	return initDefaultETFConfigs()
}

// initDefaultETFConfigs 初始化默认ETF配置
func initDefaultETFConfigs() error {
	defaultETFs := []ETFConfig{
		{
			Symbol:       "SCHD",
			Name:         "Schwab U.S. Dividend Equity ETF",
			Market:       "US",
			Strategy:     "质量股息策略",
			Description:  "追踪道琼斯美国股息100指数，投资高股息、财务稳健的美国公司",
			ExpenseRatio: decimal.NewFromFloat(0.06),
			Focus:        "质量+股息",
			Status:       1,
			SortOrder:    1,
		},
		{
			Symbol:       "SPYD",
			Name:         "SPDR Portfolio S&P 500 High Dividend ETF",
			Market:       "US",
			Strategy:     "高股息收益策略",
			Description:  "追踪S&P 500中股息收益率最高的80只股票",
			ExpenseRatio: decimal.NewFromFloat(0.07),
			Focus:        "高股息",
			Status:       1,
			SortOrder:    2,
		},
		{
			Symbol:       "JEPQ",
			Name:         "JPMorgan Nasdaq Equity Premium Income ETF",
			Market:       "US",
			Strategy:     "期权增强收益策略",
			Description:  "通过纳斯达克股票+卖出看涨期权获取增强收益",
			ExpenseRatio: decimal.NewFromFloat(0.35),
			Focus:        "增强收益",
			Status:       1,
			SortOrder:    3,
		},
		{
			Symbol:       "JEPI",
			Name:         "JPMorgan Equity Premium Income ETF",
			Market:       "US",
			Strategy:     "股息增强策略",
			Description:  "摩根大通股票溢价收益ETF，通过股票期权策略提供月度股息收益",
			ExpenseRatio: decimal.NewFromFloat(0.35),
			Focus:        "月度股息+收益增强",
			Status:       1,
			SortOrder:    4,
		},
		{
			Symbol:       "VYM",
			Name:         "Vanguard High Dividend Yield ETF",
			Market:       "US",
			Strategy:     "高股息宽基策略",
			Description:  "追踪FTSE高股息率指数，投资高股息的美国大盘股",
			ExpenseRatio: decimal.NewFromFloat(0.06),
			Focus:        "高股息+宽基",
			Status:       1,
			SortOrder:    5,
		},
	}

	for _, etf := range defaultETFs {
		var existing ETFConfig
		result := DB.Where("symbol = ?", etf.Symbol).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := DB.Create(&etf).Error; err != nil {
				logrus.WithError(err).Errorf("Failed to create ETF config: %s", etf.Symbol)
			}
		} else if result.Error != nil {
			return result.Error
		}
	}

	logrus.Info("Default ETF configs initialized")
	return nil
}
