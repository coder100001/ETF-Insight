package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ExchangeRate 汇率数据表
type ExchangeRate struct {
	ID            uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	FromCurrency  string          `gorm:"size:10;not null;index:idx_currency_pair,unique" json:"from_currency"`
	ToCurrency    string          `gorm:"size:10;not null;index:idx_currency_pair,unique" json:"to_currency"`
	Rate          decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"rate"`
	PreviousRate  decimal.Decimal `gorm:"type:decimal(20,8);default:0" json:"previous_rate"`
	ChangePercent decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"change_percent"`
	DataSource    string          `gorm:"size:50;not null" json:"data_source"`
	SourceType    string          `gorm:"size:20;default:'api'" json:"source_type"` // api, manual, calculated
	ValidStatus   int             `gorm:"default:1" json:"valid_status"`             // 1:有效, 0:无效
	Priority      int             `gorm:"default:0" json:"priority"`                 // 优先级
	SyncBatchID   string          `gorm:"size:50;index" json:"sync_batch_id"`        // 同步批次ID
	SyncedAt      *time.Time      `json:"synced_at"`
	ExpiresAt     *time.Time      `json:"expires_at"` // 数据过期时间
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ExchangeRate) TableName() string {
	return "exchange_rates"
}

// ExchangeRateSyncLog 汇率同步日志表
type ExchangeRateSyncLog struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchID       string         `gorm:"size:50;not null;index" json:"batch_id"`
	SyncType      string         `gorm:"size:20;not null" json:"sync_type"`      // full:全量, incremental:增量
	DataSource    string         `gorm:"size:50;not null" json:"data_source"`
	Status        string         `gorm:"size:20;not null" json:"status"`         // success:成功, failed:失败, partial:部分成功
	TotalCount    int            `json:"total_count"`                            // 总数据量
	SuccessCount  int            `json:"success_count"`                          // 成功数量
	FailedCount   int            `json:"failed_count"`                           // 失败数量
	SkippedCount  int            `json:"skipped_count"`                          // 跳过数量
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`
	RetryCount    int            `gorm:"default:0" json:"retry_count"`
	MaxRetryCount int            `gorm:"default:3" json:"max_retry_count"`
	StartedAt     time.Time      `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	DurationMs    int64          `json:"duration_ms"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ExchangeRateSyncLog) TableName() string {
	return "exchange_rate_sync_logs"
}

// ExchangeRateSyncDetail 汇率同步明细表
type ExchangeRateSyncDetail struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	SyncLogID    uint           `gorm:"not null;index" json:"sync_log_id"`
	FromCurrency string         `gorm:"size:10;not null" json:"from_currency"`
	ToCurrency   string         `gorm:"size:10;not null" json:"to_currency"`
	OldRate      decimal.Decimal `gorm:"type:decimal(20,8)" json:"old_rate"`
	NewRate      decimal.Decimal `gorm:"type:decimal(20,8)" json:"new_rate"`
	ChangePercent decimal.Decimal `gorm:"type:decimal(10,4)" json:"change_percent"`
	Status       string         `gorm:"size:20;not null" json:"status"` // success:成功, failed:失败, skipped:跳过
	ErrorMessage string         `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (ExchangeRateSyncDetail) TableName() string {
	return "exchange_rate_sync_details"
}

// CurrencyPair 货币对配置表
type CurrencyPair struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FromCurrency string         `gorm:"size:10;not null;index:idx_pair,unique" json:"from_currency"`
	ToCurrency   string         `gorm:"size:10;not null;index:idx_pair,unique" json:"to_currency"`
	IsActive     int            `gorm:"default:1" json:"is_active"` // 1:启用, 0:禁用
	Priority     int            `gorm:"default:0" json:"priority"`  // 优先级
	Description  string         `gorm:"size:100" json:"description"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (CurrencyPair) TableName() string {
	return "currency_pairs"
}

// InitExchangeRateTables 初始化汇率相关表
func InitExchangeRateTables() error {
	return DB.AutoMigrate(
		&ExchangeRate{},
		&ExchangeRateSyncLog{},
		&ExchangeRateSyncDetail{},
		&CurrencyPair{},
	)
}

// InitDefaultCurrencyPairs 初始化默认货币对
func InitDefaultCurrencyPairs() error {
	defaultPairs := []CurrencyPair{
		{FromCurrency: "USD", ToCurrency: "CNY", Priority: 1, Description: "美元兑人民币"},
		{FromCurrency: "USD", ToCurrency: "HKD", Priority: 2, Description: "美元兑港币"},
		{FromCurrency: "EUR", ToCurrency: "CNY", Priority: 3, Description: "欧元兑人民币"},
		{FromCurrency: "GBP", ToCurrency: "CNY", Priority: 4, Description: "英镑兑人民币"},
		{FromCurrency: "JPY", ToCurrency: "CNY", Priority: 5, Description: "日元兑人民币"},
		{FromCurrency: "CNY", ToCurrency: "USD", Priority: 6, Description: "人民币兑美元"},
		{FromCurrency: "HKD", ToCurrency: "USD", Priority: 7, Description: "港币兑美元"},
	}

	for _, pair := range defaultPairs {
		result := DB.Where("from_currency = ? AND to_currency = ?", pair.FromCurrency, pair.ToCurrency).First(&CurrencyPair{})
		if result.Error == gorm.ErrRecordNotFound {
			if err := DB.Create(&pair).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
