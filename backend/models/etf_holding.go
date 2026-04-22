package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ETFHolding ETF持仓穿透表
// 用于存储ETF的底层持仓明细，支持持仓穿透分析
type ETFHolding struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	ETFID           uint            `json:"etf_id" gorm:"index:idx_etf_holding_date"`        // ETF资产ID
	UnderlyingAssetID uint          `json:"underlying_asset_id" gorm:"index:idx_etf_holding_date"` // 底层资产ID
	Symbol          string          `json:"symbol" gorm:"size:20"`                           // 底层资产代码（冗余存储）
	Name            string          `json:"name" gorm:"size:100"`                            // 底层资产名称
	Weight          decimal.Decimal `json:"weight" gorm:"type:decimal(8,4)"`                 // 权重(%)
	Shares          int64           `json:"shares"`                                          // 持股数量
	MarketValue     decimal.Decimal `json:"market_value" gorm:"type:decimal(15,2)"`          // 市值
	Date            time.Time       `json:"date" gorm:"index:idx_etf_holding_date"`          // 持仓日期
	DataSource      string          `json:"data_source" gorm:"size:50"`                      // 数据来源

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ETFHolding) TableName() string {
	return "etf_holdings"
}

// ETFHoldingSummary ETF持仓汇总
// 用于快速查询ETF的持仓概览
type ETFHoldingSummary struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	ETFID           uint            `json:"etf_id" gorm:"uniqueIndex:idx_etf_summary_date"` // ETF资产ID
	Date            time.Time       `json:"date" gorm:"uniqueIndex:idx_etf_summary_date"`   // 持仓日期
	TotalHoldings   int             `json:"total_holdings"`                                 // 总持仓数量
	Top10Weight     decimal.Decimal `json:"top10_weight" gorm:"type:decimal(5,2)"`          // 前十大持仓权重
	SectorCount     int             `json:"sector_count"`                                   // 行业数量
	CountryCount    int             `json:"country_count"`                                  // 国家数量
	Concentration   decimal.Decimal `json:"concentration" gorm:"type:decimal(5,2)"`         // 集中度（赫芬达尔指数）
	DataSource      string          `json:"data_source" gorm:"size:50"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ETFHoldingSummary) TableName() string {
	return "etf_holding_summaries"
}

// PortfolioOverlap 组合重叠度计算结果
type PortfolioOverlap struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	ETFAID        uint            `json:"etf_a_id" gorm:"index:idx_overlap"`      // ETF A ID
	ETFBID        uint            `json:"etf_b_id" gorm:"index:idx_overlap"`      // ETF B ID
	OverlapScore  decimal.Decimal `json:"overlap_score" gorm:"type:decimal(5,2)"` // 重叠度分数(0-100)
	CommonHoldings int            `json:"common_holdings"`                         // 共同持仓数量
	TotalWeightA  decimal.Decimal `json:"total_weight_a" gorm:"type:decimal(5,2)"` // A中重叠权重
	TotalWeightB  decimal.Decimal `json:"total_weight_b" gorm:"type:decimal(5,2)"` // B中重叠权重
	CalculatedAt  time.Time       `json:"calculated_at"`                           // 计算时间

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PortfolioOverlap) TableName() string {
	return "portfolio_overlaps"
}
