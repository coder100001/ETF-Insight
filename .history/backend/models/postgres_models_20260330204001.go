package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ETFConfig ETF配置模型
type ETFConfig struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	Symbol       string          `gorm:"uniqueIndex;size:20;not null" json:"symbol"`
	Name         string          `gorm:"size:200;not null" json:"name"`
	Currency     string          `gorm:"size:10;default:'USD'" json:"currency"`
	Exchange     string          `gorm:"size:20" json:"exchange"`
	Category     string          `gorm:"size:50" json:"category"`
	Provider     string          `gorm:"size:100" json:"provider"`
	ExpenseRatio decimal.Decimal `gorm:"type:decimal(10,4)" json:"expense_ratio"`
	AUM          decimal.Decimal `gorm:"type:decimal(20,2)" json:"aum"`
	Status       int             `gorm:"default:1" json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (ETFConfig) TableName() string {
	return "etf_configs"
}

// ETFPrice ETF价格数据
type ETFPrice struct {
	ID            uint64          `gorm:"primaryKey" json:"id"`
	Symbol        string          `gorm:"index;size:20;not null" json:"symbol"`
	Date          time.Time       `gorm:"index;type:date;not null" json:"date"`
	Open          decimal.Decimal `gorm:"type:decimal(18,4)" json:"open"`
	High          decimal.Decimal `gorm:"type:decimal(18,4)" json:"high"`
	Low           decimal.Decimal `gorm:"type:decimal(18,4)" json:"low"`
	Close         decimal.Decimal `gorm:"type:decimal(18,4);not null" json:"close"`
	Volume        int64           `json:"volume"`
	AdjustedClose decimal.Decimal `gorm:"type:decimal(18,4)" json:"adjusted_close"`
	CreatedAt     time.Time       `json:"created_at"`
}

// TableName 指定表名
func (ETFPrice) TableName() string {
	return "etf_prices"
}

// ETFHolding ETF持仓数据
type ETFHolding struct {
	ID          uint64          `gorm:"primaryKey" json:"id"`
	ETFSymbol   string          `gorm:"index;size:20;not null" json:"etf_symbol"`
	StockSymbol string          `gorm:"index;size:20;not null" json:"stock_symbol"`
	StockName   string          `gorm:"size:200" json:"stock_name"`
	Weight      decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"weight"`
	Sector      string          `gorm:"size:50" json:"sector"`
	Country     string          `gorm:"size:50" json:"country"`
	MarketCap   string          `gorm:"size:20" json:"market_cap"`
	ReportDate  time.Time       `gorm:"index;type:date;not null" json:"report_date"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TableName 指定表名
func (ETFHolding) TableName() string {
	return "etf_holdings"
}

// ETFRealtime ETF实时数据
type ETFRealtime struct {
	ID               uint64          `gorm:"primaryKey" json:"id"`
	Symbol           string          `gorm:"uniqueIndex;size:20;not null" json:"symbol"`
	CurrentPrice     decimal.Decimal `gorm:"type:decimal(18,4)" json:"current_price"`
	PreviousClose    decimal.Decimal `gorm:"type:decimal(18,4)" json:"previous_close"`
	OpenPrice        decimal.Decimal `gorm:"type:decimal(18,4)" json:"open_price"`
	DayHigh          decimal.Decimal `gorm:"type:decimal(18,4)" json:"day_high"`
	DayLow           decimal.Decimal `gorm:"type:decimal(18,4)" json:"day_low"`
	Volume           int64           `json:"volume"`
	Change           decimal.Decimal `gorm:"type:decimal(18,4)" json:"change"`
	ChangePercent    decimal.Decimal `gorm:"type:decimal(10,4)" json:"change_percent"`
	MarketCap        int64           `json:"market_cap"`
	DividendYield    decimal.Decimal `gorm:"type:decimal(10,4)" json:"dividend_yield"`
	FiftyTwoWeekHigh decimal.Decimal `gorm:"type:decimal(18,4)" json:"fifty_two_week_high"`
	FiftyTwoWeekLow  decimal.Decimal `gorm:"type:decimal(18,4)" json:"fifty_two_week_low"`
	AverageVolume    int64           `json:"average_volume"`
	Beta             decimal.Decimal `gorm:"type:decimal(10,4)" json:"beta"`
	PERatio          decimal.Decimal `gorm:"type:decimal(10,4)" json:"pe_ratio"`
	Currency         string          `gorm:"size:10" json:"currency"`
	DataSource       string          `gorm:"size:50" json:"data_source"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (ETFRealtime) TableName() string {
	return "etf_realtime"
}

// ExchangeRate 汇率数据
type ExchangeRate struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	FromCurrency string          `gorm:"index;size:10;not null" json:"from_currency"`
	ToCurrency   string          `gorm:"index;size:10;not null" json:"to_currency"`
	Rate         decimal.Decimal `gorm:"type:decimal(18,8);not null" json:"rate"`
	Date         time.Time       `gorm:"index;type:date;not null" json:"date"`
	CreatedAt    time.Time       `json:"created_at"`
}

// TableName 指定表名
func (ExchangeRate) TableName() string {
	return "exchange_rates"
}

// AnalysisTask 分析任务
type AnalysisTask struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Type         string          `gorm:"size:50;not null" json:"type"`
	Status       string          `gorm:"size:20;default:'pending'" json:"status"`
	Payload      json.RawMessage `gorm:"type:jsonb;not null" json:"payload"`
	Result       json.RawMessage `gorm:"type:jsonb" json:"result,omitempty"`
	ErrorMessage sql.NullString  `json:"error_message,omitempty"`
	WorkerID     sql.NullString  `json:"worker_id,omitempty"`
	StartedAt    sql.NullTime    `json:"started_at,omitempty"`
	CompletedAt  sql.NullTime    `json:"completed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// TableName 指定表名
func (AnalysisTask) TableName() string {
	return "analysis_tasks"
}

// BeforeCreate 创建前钩子
func (t *AnalysisTask) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// AnalysisCache 分析结果缓存
type AnalysisCache struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CacheKey  string          `gorm:"uniqueIndex;size:255;not null" json:"cache_key"`
	DataType  string          `gorm:"size:50;not null" json:"data_type"`
	Data      json.RawMessage `gorm:"type:jsonb;not null" json:"data"`
	ExpiresAt time.Time       `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// TableName 指定表名
func (AnalysisCache) TableName() string {
	return "analysis_cache"
}

// IsExpired 检查是否过期
func (c *AnalysisCache) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// OperationLog 操作日志
type OperationLog struct {
	ID           uint64          `gorm:"primaryKey" json:"id"`
	Operation    string          `gorm:"size:100;not null" json:"operation"`
	Details      json.RawMessage `gorm:"type:jsonb" json:"details,omitempty"`
	Status       string          `gorm:"size:20" json:"status"`
	ErrorMessage sql.NullString  `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// Repository 仓储接口
type Repository interface {
	// ETF配置
	GetETFConfig(symbol string) (*ETFConfig, error)
	GetAllETFConfigs() ([]ETFConfig, error)
	SaveETFConfig(config *ETFConfig) error

	// 价格数据
	GetETFPrices(symbol string, start, end time.Time) ([]ETFPrice, error)
	SaveETFPrice(price *ETFPrice) error
	SaveETFPrices(prices []ETFPrice) error

	// 持仓数据
	GetETFHoldings(etfSymbol string, reportDate time.Time) ([]ETFHolding, error)
	SaveETFHolding(holding *ETFHolding) error

	// 实时数据
	GetETFRealtime(symbol string) (*ETFRealtime, error)
	SaveETFRealtime(data *ETFRealtime) error

	// 汇率
	GetExchangeRate(from, to string, date time.Time) (*ExchangeRate, error)
	SaveExchangeRate(rate *ExchangeRate) error

	// 分析任务
	GetAnalysisTask(id uuid.UUID) (*AnalysisTask, error)
	CreateAnalysisTask(task *AnalysisTask) error
	UpdateAnalysisTask(task *AnalysisTask) error
	GetPendingTasks(limit int) ([]AnalysisTask, error)

	// 缓存
	GetAnalysisCache(key string) (*AnalysisCache, error)
	SaveAnalysisCache(cache *AnalysisCache) error
	DeleteExpiredCache() error
}
