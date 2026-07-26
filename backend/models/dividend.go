package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ETFDividend ETF分红记录
type ETFDividend struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	ETFID            uint            `json:"etf_id"`
	Symbol           string          `json:"symbol" gorm:"size:20"`
	ExDividendDate   time.Time       `json:"ex_dividend_date"`
	RecordDate       time.Time       `json:"record_date"`
	PaymentDate      time.Time       `json:"payment_date"`
	DividendPerShare decimal.Decimal `json:"dividend_per_share" gorm:"type:decimal(10,4)"`
	DividendYield    decimal.Decimal `json:"dividend_yield" gorm:"type:decimal(5,2)"`

	CreatedAt time.Time `json:"created_at"`
}
