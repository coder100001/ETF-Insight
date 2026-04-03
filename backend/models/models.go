package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ETFConfig ETF配置
type ETFConfig struct {
	ID              uint            `json:"id" yaml:"id"`
	Symbol          string          `json:"symbol" yaml:"symbol" gorm:"uniqueIndex"`
	Name            string          `json:"name" yaml:"name"`
	Description     string          `json:"description" yaml:"description"`
	Strategy        string          `json:"strategy" yaml:"strategy"`
	Focus           string          `json:"focus" yaml:"focus"`
	ExpenseRatio    decimal.Decimal `json:"expense_ratio" yaml:"expense_ratio" gorm:"type:decimal(10,4)"`
	Currency        string          `json:"currency" yaml:"currency"`
	Exchange        string          `json:"exchange" yaml:"exchange"`
	Category        string          `json:"category" yaml:"category"`
	Provider        string          `json:"provider" yaml:"provider"`
	Inception       string          `json:"inception" yaml:"inception"`
	AUM             decimal.Decimal `json:"aum" yaml:"aum" gorm:"type:decimal(20,2)"`
	Status          int             `json:"status" yaml:"status" gorm:"default:1"`
	AutoUpdate      bool            `json:"auto_update" yaml:"auto_update" gorm:"default:true"`
	UpdateFrequency string          `json:"update_frequency" yaml:"update_frequency" gorm:"default:'每日'"`
	DataSource      string          `json:"data_source" yaml:"data_source" gorm:"default:'Yahoo Finance'"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ETFData ETF数据
type ETFData struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	Symbol     string          `json:"symbol" gorm:"uniqueIndex:idx_symbol_date"`
	Date       time.Time       `json:"date" gorm:"uniqueIndex:idx_symbol_date"`
	OpenPrice  decimal.Decimal `json:"open_price" gorm:"type:decimal(20,8)"`
	ClosePrice decimal.Decimal `json:"close_price" gorm:"type:decimal(20,8)"`
	HighPrice  decimal.Decimal `json:"high_price" gorm:"type:decimal(20,8)"`
	LowPrice   decimal.Decimal `json:"low_price" gorm:"type:decimal(20,8)"`
	Volume     int64           `json:"volume"`
	DataSource string          `json:"data_source"`
	CreatedAt  time.Time       `json:"created_at"`
}

// OperationLog 操作日志
type OperationLog struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	OperationType string     `json:"operation_type"`
	OperationName string     `json:"operation_name"`
	Operator      string     `json:"operator"`
	Status        int        `json:"status"` // 0:进行中, 1:成功, 2:失败
	ErrorMessage  string     `json:"error_message"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	DurationMs    int        `json:"duration_ms"`
	Details       string     `json:"details" gorm:"type:text"`
}

// ETFDefinitions ETF配置列表
type ETFDefinitions struct {
	ETFs []ETFConfig `yaml:"etfs"`
}

// PortfolioConfig 投资组合配置
type PortfolioConfig struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	Name            string          `json:"name" gorm:"size:100;not null"`              // 组合名称
	Description     string          `json:"description" gorm:"size:500"`                // 组合描述
	Allocation      string          `json:"allocation" gorm:"type:text;not null"`       // 配置JSON: {"QQQ": 50, "SCHD": 50}
	TotalInvestment decimal.Decimal `json:"total_investment" gorm:"type:decimal(15,2)"` // 总投资金额
	TaxRate         decimal.Decimal `json:"tax_rate" gorm:"type:decimal(5,4)"`          // 税率(如0.10表示10%)
	Status          int             `json:"status" gorm:"default:1"`                    // 状态: 1-启用, 0-禁用
	IsDefault       bool            `json:"is_default" gorm:"default:false"`            // 是否为默认组合
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (PortfolioConfig) TableName() string {
	return "portfolio_configs"
}
