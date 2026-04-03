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
		&ETFConfig{},
		&ETFData{},
		&OperationLog{},
		&ExchangeRate{},
		&AShareDividendETF{},
		&AShareETFPortfolio{},
		&ASharePortfolioHolding{},
		&PortfolioConfig{},
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
