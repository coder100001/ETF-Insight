package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PriceType 价格类型
type PriceType string

const (
	PriceTypeDaily    PriceType = "daily"     // 日级价格
	PriceTypeWeekly   PriceType = "weekly"    // 周级价格
	PriceTypeMonthly  PriceType = "monthly"   // 月级价格
	PriceTypeMinute   PriceType = "minute"    // 分钟级价格
	PriceTypeHourly   PriceType = "hourly"    // 小时级价格
	PriceTypeRealTime PriceType = "real_time" // 实时价格
)

// Price 统一价格模型
// 所有资产类型的价格数据都存储在此表中
type Price struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AssetID   uint      `json:"asset_id" gorm:"index:idx_asset_price"`           // 资产ID
	Symbol    string    `json:"symbol" gorm:"size:20;index:idx_asset_price"`     // 资产代码（冗余，便于查询）
	Date      time.Time `json:"date" gorm:"index:idx_asset_price"`               // 价格日期/时间
	PriceType PriceType `json:"price_type" gorm:"size:20;index:idx_asset_price"` // 价格类型

	// OHLCV数据
	Open     decimal.Decimal `json:"open" gorm:"type:decimal(20,8)"`               // 开盘价
	High     decimal.Decimal `json:"high" gorm:"type:decimal(20,8)"`               // 最高价
	Low      decimal.Decimal `json:"low" gorm:"type:decimal(20,8)"`                // 最低价
	Close    decimal.Decimal `json:"close" gorm:"type:decimal(20,8)"`              // 收盘价
	Volume   int64           `json:"volume"`                                       // 成交量
	Turnover decimal.Decimal `json:"turnover,omitempty" gorm:"type:decimal(20,2)"` // 成交额
	VWAP     decimal.Decimal `json:"vwap,omitempty" gorm:"type:decimal(20,8)"`     // 成交量加权平均价

	// 调整后价格（用于计算收益率）
	AdjOpen     decimal.Decimal `json:"adj_open,omitempty" gorm:"type:decimal(20,8)"`     // 调整后开盘价
	AdjHigh     decimal.Decimal `json:"adj_high,omitempty" gorm:"type:decimal(20,8)"`     // 调整后最高价
	AdjLow      decimal.Decimal `json:"adj_low,omitempty" gorm:"type:decimal(20,8)"`      // 调整后最低价
	AdjClose    decimal.Decimal `json:"adj_close,omitempty" gorm:"type:decimal(20,8)"`    // 调整后收盘价
	AdjVolume   int64           `json:"adj_volume,omitempty"`                             // 调整后成交量
	SplitFactor decimal.Decimal `json:"split_factor,omitempty" gorm:"type:decimal(10,6)"` // 拆股因子

	// ETF特有字段
	NAV         decimal.Decimal `json:"nav,omitempty" gorm:"type:decimal(20,8)"`         // 净值（仅ETF）
	PremiumRate decimal.Decimal `json:"premium_rate,omitempty" gorm:"type:decimal(5,2)"` // 溢价率（仅ETF）

	// 衍生指标（可实时计算或缓存）
	ReturnDaily  decimal.Decimal `json:"return_daily,omitempty" gorm:"type:decimal(10,6)"`  // 日收益率
	Volatility20 decimal.Decimal `json:"volatility_20,omitempty" gorm:"type:decimal(10,6)"` // 20日波动率
	MA20         decimal.Decimal `json:"ma_20,omitempty" gorm:"type:decimal(20,8)"`         // 20日移动平均
	MA50         decimal.Decimal `json:"ma_50,omitempty" gorm:"type:decimal(20,8)"`         // 50日移动平均
	MA200        decimal.Decimal `json:"ma_200,omitempty" gorm:"type:decimal(20,8)"`        // 200日移动平均

	// 数据质量
	DataSource   string `json:"data_source" gorm:"size:50"`       // 数据来源
	IsValid      bool   `json:"is_valid" gorm:"default:true"`     // 数据是否有效
	IsImputed    bool   `json:"is_imputed" gorm:"default:false"`  // 是否为插值数据
	QualityScore int    `json:"quality_score" gorm:"default:100"` // 数据质量评分 0-100

	// 时间戳
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Asset Asset `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
}

// TableName 指定表名
func (Price) TableName() string {
	return "prices"
}

// PriceGap 价格缺口记录
type PriceGap struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AssetID     uint      `json:"asset_id" gorm:"index"`                   // 资产ID
	Symbol      string    `json:"symbol" gorm:"size:20"`                   // 资产代码
	StartDate   time.Time `json:"start_date"`                              // 缺失开始日期
	EndDate     time.Time `json:"end_date"`                                // 缺失结束日期
	MissingDays int       `json:"missing_days"`                            // 缺失天数
	PriceType   PriceType `json:"price_type" gorm:"size:20"`               // 价格类型
	Status      string    `json:"status" gorm:"size:20;default:'pending'"` // 状态：pending/filled/ignored
	FillMethod  string    `json:"fill_method,omitempty" gorm:"size:20"`    // 填充方法：interpolation/forward/backward
	FillDate    time.Time `json:"fill_date,omitempty"`                     // 填充日期

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PriceGap) TableName() string {
	return "price_gaps"
}

// PriceStats 价格统计（物化视图）
type PriceStats struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	AssetID  uint      `json:"asset_id" gorm:"uniqueIndex:idx_asset_stats"`       // 资产ID
	Period   string    `json:"period" gorm:"size:20;uniqueIndex:idx_asset_stats"` // 统计周期：daily/weekly/monthly
	AsOfDate time.Time `json:"as_of_date" gorm:"uniqueIndex:idx_asset_stats"`     // 截至日期

	// 基础统计
	Count     int       `json:"count"`      // 数据点数量
	StartDate time.Time `json:"start_date"` // 起始日期
	EndDate   time.Time `json:"end_date"`   // 结束日期

	// 价格统计
	MinPrice    decimal.Decimal `json:"min_price" gorm:"type:decimal(20,8)"`     // 最低价
	MaxPrice    decimal.Decimal `json:"max_price" gorm:"type:decimal(20,8)"`     // 最高价
	AvgPrice    decimal.Decimal `json:"avg_price" gorm:"type:decimal(20,8)"`     // 平均价
	MedianPrice decimal.Decimal `json:"median_price" gorm:"type:decimal(20,8)"`  // 中位数价格
	StdDevPrice decimal.Decimal `json:"std_dev_price" gorm:"type:decimal(20,8)"` // 价格标准差

	// 成交量统计
	TotalVolume  int64           `json:"total_volume"`                             // 总成交量
	AvgVolume    int64           `json:"avg_volume"`                               // 平均成交量
	MaxVolume    int64           `json:"max_volume"`                               // 最大成交量
	VolumeStdDev decimal.Decimal `json:"volume_std_dev" gorm:"type:decimal(20,2)"` // 成交量标准差

	// 收益率统计
	TotalReturn  decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`  // 总收益率
	AnnualReturn decimal.Decimal `json:"annual_return" gorm:"type:decimal(10,6)"` // 年化收益率
	Volatility   decimal.Decimal `json:"volatility" gorm:"type:decimal(10,6)"`    // 波动率
	SharpeRatio  decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,6)"`  // 夏普比率
	MaxDrawdown  decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,6)"`  // 最大回撤

	// 相关性统计（JSON存储）
	Correlations string `json:"correlations,omitempty" gorm:"type:json"` // 与其他资产的相关性

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PriceStats) TableName() string {
	return "price_stats"
}
