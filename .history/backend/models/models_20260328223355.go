package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ETFConfig ETF配置
type ETFConfig struct {
	ID           uint            `json:"id" yaml:"id"`
	Symbol       string          `json:"symbol" yaml:"symbol" gorm:"uniqueIndex"`
	Name         string          `json:"name" yaml:"name"`
	Description  string          `json:"description" yaml:"description"`
	Strategy     string          `json:"strategy" yaml:"strategy"`
	Focus        string          `json:"focus" yaml:"focus"`
	ExpenseRatio decimal.Decimal `json:"expense_ratio" yaml:"expense_ratio" gorm:"type:decimal(10,4)"`
	Currency     string          `json:"currency" yaml:"currency"`
	Exchange     string          `json:"exchange" yaml:"exchange"`
	Category     string          `json:"category" yaml:"category"`
	Provider     string          `json:"provider" yaml:"provider"`
	Inception    string          `json:"inception" yaml:"inception"`
	AUM          decimal.Decimal `json:"aum" yaml:"aum" gorm:"type:decimal(20,2)"`
	Status       int             `json:"status" yaml:"status" gorm:"default:1"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ETFData ETF数据
type ETFData struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	Symbol     string          `json:"symbol" gorm:"index"`
	Date       time.Time       `json:"date" gorm:"index"`
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

// ExchangeRate 汇率
type ExchangeRate struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	FromCurrency string          `json:"from_currency" gorm:"index"`
	ToCurrency   string          `json:"to_currency" gorm:"index"`
	Rate         decimal.Decimal `json:"rate" gorm:"type:decimal(20,8)"`
	RateDate     string          `json:"rate_date" gorm:"index"`
	DataSource   string          `json:"data_source"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
