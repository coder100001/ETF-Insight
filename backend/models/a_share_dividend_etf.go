package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// DividendFrequency 分红频率
type DividendFrequency string

const (
	FrequencyMonthly   DividendFrequency = "月分" // 月分
	FrequencyQuarterly DividendFrequency = "季分" // 季分
	FrequencyYearly    DividendFrequency = "年分" // 年分
)

// AShareDividendETF A股红利ETF产品信息
type AShareDividendETF struct {
	ID                uint              `json:"id" gorm:"primaryKey"`
	Symbol            string            `json:"symbol" gorm:"uniqueIndex;size:20"`           // ETF代码，如515080
	Name              string            `json:"name" gorm:"size:100"`                        // ETF名称
	DividendYieldMin  decimal.Decimal   `json:"dividend_yield_min" gorm:"type:decimal(5,2)"` // 股息率下限(%)
	DividendYieldMax  decimal.Decimal   `json:"dividend_yield_max" gorm:"type:decimal(5,2)"` // 股息率上限(%)
	DividendFrequency DividendFrequency `json:"dividend_frequency" gorm:"size:10"`           // 分红频率：月分/季分/年分
	Benchmark         string            `json:"benchmark" gorm:"size:100"`                   // 跟踪基准指数
	Exchange          string            `json:"exchange" gorm:"size:20;default:'SSE'"`       // 交易所：SSE/SHZ
	ManagementFee     decimal.Decimal   `json:"management_fee" gorm:"type:decimal(5,4)"`     // 管理费率(%)
	Description       string            `json:"description" gorm:"size:500"`                 // 产品描述
	Status            int               `json:"status" gorm:"default:1"`                     // 状态：1-正常，0-停用
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// AShareETFPortfolio A股ETF组合配置
type AShareETFPortfolio struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	Name            string          `json:"name" gorm:"size:100;default:'默认组合'"`        // 组合名称
	TotalInvestment decimal.Decimal `json:"total_investment" gorm:"type:decimal(15,2)"` // 总投资金额
	IsDefault       bool            `json:"is_default" gorm:"default:false"`            // 是否为默认组合
	Description     string          `json:"description" gorm:"size:500"`                // 组合描述
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ASharePortfolioHolding 组合持仓明细
type ASharePortfolioHolding struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	PortfolioID uint            `json:"portfolio_id" gorm:"index"`            // 组合ID
	ETFID       uint            `json:"etf_id" gorm:"index"`                  // ETF产品ID
	Investment  decimal.Decimal `json:"investment" gorm:"type:decimal(15,2)"` // 投资金额
	Weight      decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`      // 占比(%)
	SortOrder   int             `json:"sort_order" gorm:"default:0"`          // 排序
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	// 关联
	ETF AShareDividendETF `json:"etf,omitempty" gorm:"foreignKey:ETFID"`
}

// AShareDividendCalculation 分红收益计算
type AShareDividendCalculation struct {
	PortfolioID            uint                  `json:"portfolio_id"`
	TotalInvestment        decimal.Decimal       `json:"total_investment"`         // 总投资金额
	ExpectedAnnualDividend decimal.Decimal       `json:"expected_annual_dividend"` // 预期年分红总额
	AverageDividendYield   decimal.Decimal       `json:"average_dividend_yield"`   // 平均股息率
	MonthlyDividend        decimal.Decimal       `json:"monthly_dividend"`         // 月均分红
	QuarterlyDividend      decimal.Decimal       `json:"quarterly_dividend"`       // 季均分红
	Holdings               []AShareHoldingDetail `json:"holdings"`                 // 持仓明细
}

// AShareHoldingDetail 持仓分红明细
type AShareHoldingDetail struct {
	Symbol               string          `json:"symbol"`                // ETF代码
	Name                 string          `json:"name"`                  // ETF名称
	Investment           decimal.Decimal `json:"investment"`            // 投资金额
	Weight               decimal.Decimal `json:"weight"`                // 占比
	DividendYield        decimal.Decimal `json:"dividend_yield"`        // 股息率(取中间值)
	DividendFrequency    string          `json:"dividend_frequency"`    // 分红频率
	ExpectedDividend     decimal.Decimal `json:"expected_dividend"`     // 预期年分红
	DividendContribution decimal.Decimal `json:"dividend_contribution"` // 分红贡献占比
}

// TableName 指定表名
func (AShareDividendETF) TableName() string {
	return "a_share_dividend_etfs"
}

func (AShareETFPortfolio) TableName() string {
	return "a_share_etf_portfolios"
}

func (ASharePortfolioHolding) TableName() string {
	return "a_share_portfolio_holdings"
}
