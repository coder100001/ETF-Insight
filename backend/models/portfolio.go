package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PortfolioType 组合类型
type PortfolioType string

const (
	PortfolioTypePersonal  PortfolioType = "personal"  // 个人投资组合
	PortfolioTypeModel     PortfolioType = "model"     // 模型组合
	PortfolioTypeBenchmark PortfolioType = "benchmark" // 基准组合
	PortfolioTypeWatchlist PortfolioType = "watchlist" // 观察列表
)

// PortfolioStatus 组合状态
type PortfolioStatus string

const (
	PortfolioStatusActive   PortfolioStatus = "active"   // 活跃
	PortfolioStatusArchived PortfolioStatus = "archived" // 归档
	PortfolioStatusPaused   PortfolioStatus = "paused"   // 暂停
)

// RebalanceStrategy 再平衡策略
type RebalanceStrategy string

const (
	RebalanceStrategyNone      RebalanceStrategy = "none"      // 不再平衡
	RebalanceStrategyTime      RebalanceStrategy = "time"      // 定时再平衡
	RebalanceStrategyThreshold RebalanceStrategy = "threshold" // 阈值再平衡
	RebalanceStrategyHybrid    RebalanceStrategy = "hybrid"    // 混合再平衡
)

// Portfolio 投资组合
type Portfolio struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Name          string          `json:"name" gorm:"size:100;not null"`          // 组合名称
	Description   string          `json:"description" gorm:"size:500"`            // 组合描述
	PortfolioType PortfolioType   `json:"portfolio_type" gorm:"size:20;index"`    // 组合类型
	Status        PortfolioStatus `json:"status" gorm:"size:20;default:'active'"` // 组合状态

	// 所有者信息
	UserID    string `json:"user_id,omitempty" gorm:"size:100;index"` // 用户ID（可选，支持多用户）
	IsPublic  bool   `json:"is_public" gorm:"default:false"`          // 是否公开
	IsDefault bool   `json:"is_default" gorm:"default:false"`         // 是否为默认组合

	// 货币和基准
	BaseCurrency    string `json:"base_currency" gorm:"size:10;default:'USD'"` // 基准货币
	BenchmarkID     uint   `json:"benchmark_id,omitempty" gorm:"index"`        // 基准组合ID
	BenchmarkSymbol string `json:"benchmark_symbol,omitempty" gorm:"size:20"`  // 基准代码（冗余）

	// 资金管理
	InitialCapital decimal.Decimal `json:"initial_capital" gorm:"type:decimal(15,2)"` // 初始资金
	CurrentValue   decimal.Decimal `json:"current_value" gorm:"type:decimal(15,2)"`   // 当前市值
	CashBalance    decimal.Decimal `json:"cash_balance" gorm:"type:decimal(15,2)"`    // 现金余额

	// 再平衡配置
	RebalanceStrategy  RebalanceStrategy `json:"rebalance_strategy" gorm:"size:20"`            // 再平衡策略
	RebalanceFrequency string            `json:"rebalance_frequency,omitempty" gorm:"size:20"` // 再平衡频率：monthly/quarterly/yearly
	ThresholdPercent   decimal.Decimal   `json:"threshold_percent" gorm:"type:decimal(5,2)"`   // 再平衡阈值百分比
	LastRebalance      time.Time         `json:"last_rebalance"`                               // 上次再平衡时间
	NextRebalance      time.Time         `json:"next_rebalance"`                               // 下次再平衡时间

	// 目标配置
	TargetAllocation JSONMap `json:"target_allocation,omitempty" gorm:"type:json"` // 目标配置（JSON格式）
	RiskTolerance    string `json:"risk_tolerance,omitempty" gorm:"size:20"`      // 风险承受能力：conservative/moderate/aggressive
	TimeHorizon      string `json:"time_horizon,omitempty" gorm:"size:20"`        // 投资期限：short/medium/long

	// 表现指标（缓存）
	TotalReturn  decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`  // 总收益率
	AnnualReturn decimal.Decimal `json:"annual_return" gorm:"type:decimal(10,6)"` // 年化收益率
	Volatility   decimal.Decimal `json:"volatility" gorm:"type:decimal(10,6)"`    // 波动率
	SharpeRatio  decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,6)"`  // 夏普比率
	MaxDrawdown  decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,6)"`  // 最大回撤
	Alpha        decimal.Decimal `json:"alpha" gorm:"type:decimal(10,6)"`         // Alpha值
	Beta         decimal.Decimal `json:"beta" gorm:"type:decimal(10,6)"`          // Beta值

	// 更新时间
	LastCalculated time.Time `json:"last_calculated"` // 最后计算时间

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Positions []PortfolioPosition `json:"positions,omitempty" gorm:"foreignKey:PortfolioID"`
	Benchmark *Portfolio          `json:"benchmark,omitempty" gorm:"foreignKey:BenchmarkID"`
}

// TableName 指定表名
func (Portfolio) TableName() string {
	return "portfolios"
}

// PortfolioPosition 组合持仓
type PortfolioPosition struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	PortfolioID uint   `json:"portfolio_id" gorm:"index:idx_portfolio_asset"` // 组合ID
	AssetID     uint   `json:"asset_id" gorm:"index:idx_portfolio_asset"`     // 资产ID
	Symbol      string `json:"symbol" gorm:"size:20"`                         // 资产代码（冗余）

	// 持仓信息
	Quantity    decimal.Decimal `json:"quantity" gorm:"type:decimal(15,6)"`     // 持仓数量
	CostBasis   decimal.Decimal `json:"cost_basis" gorm:"type:decimal(15,2)"`   // 成本基础
	AverageCost decimal.Decimal `json:"average_cost" gorm:"type:decimal(15,6)"` // 平均成本
	MarketValue decimal.Decimal `json:"market_value" gorm:"type:decimal(15,2)"` // 当前市值
	Weight      decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`        // 权重百分比

	// 目标配置
	TargetWeight decimal.Decimal `json:"target_weight" gorm:"type:decimal(5,2)"` // 目标权重
	MinWeight    decimal.Decimal `json:"min_weight" gorm:"type:decimal(5,2)"`    // 最小权重
	MaxWeight    decimal.Decimal `json:"max_weight" gorm:"type:decimal(5,2)"`    // 最大权重

	// 交易信息
	FirstPurchase time.Time `json:"first_purchase"` // 首次购买日期
	LastPurchase  time.Time `json:"last_purchase"`  // 最后购买日期
	LastSale      time.Time `json:"last_sale"`      // 最后卖出日期

	// 表现指标
	UnrealizedGain decimal.Decimal `json:"unrealized_gain" gorm:"type:decimal(15,2)"` // 未实现收益
	RealizedGain   decimal.Decimal `json:"realized_gain" gorm:"type:decimal(15,2)"`   // 已实现收益
	TotalReturn    decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`    // 总收益率

	// 状态
	IsActive bool   `json:"is_active" gorm:"default:true"`   // 是否活跃持仓
	Notes    string `json:"notes,omitempty" gorm:"size:200"` // 备注

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Portfolio Portfolio `json:"portfolio" gorm:"foreignKey:PortfolioID"`
	Asset     Asset     `json:"asset" gorm:"foreignKey:AssetID"`
}

// TableName 指定表名
func (PortfolioPosition) TableName() string {
	return "portfolio_positions"
}

// PortfolioPerformance 组合表现快照（按日/周/月记录）
type PortfolioPerformance struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PortfolioID uint      `json:"portfolio_id" gorm:"index:idx_portfolio_date"` // 组合ID
	Date        time.Time `json:"date" gorm:"index:idx_portfolio_date"`         // 日期
	Period      string    `json:"period" gorm:"size:20"`                        // 周期：daily/weekly/monthly

	// 价值指标
	TotalValue  decimal.Decimal `json:"total_value" gorm:"type:decimal(15,2)"`  // 总市值
	CashBalance decimal.Decimal `json:"cash_balance" gorm:"type:decimal(15,2)"` // 现金余额
	NetDeposits decimal.Decimal `json:"net_deposits" gorm:"type:decimal(15,2)"` // 净存款

	// 收益指标
	DailyReturn   decimal.Decimal `json:"daily_return" gorm:"type:decimal(10,6)"`   // 日收益率
	WeeklyReturn  decimal.Decimal `json:"weekly_return" gorm:"type:decimal(10,6)"`  // 周收益率
	MonthlyReturn decimal.Decimal `json:"monthly_return" gorm:"type:decimal(10,6)"` // 月收益率
	YTDReturn     decimal.Decimal `json:"ytd_return" gorm:"type:decimal(10,6)"`     // 年初至今收益率
	TotalReturn   decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`   // 总收益率

	// 风险指标
	Volatility   decimal.Decimal `json:"volatility" gorm:"type:decimal(10,6)"`    // 波动率
	MaxDrawdown  decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,6)"`  // 最大回撤
	SharpeRatio  decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,6)"`  // 夏普比率
	SortinoRatio decimal.Decimal `json:"sortino_ratio" gorm:"type:decimal(10,6)"` // 索提诺比率

	// 相对于基准
	Alpha            decimal.Decimal `json:"alpha" gorm:"type:decimal(10,6)"`             // Alpha值
	Beta             decimal.Decimal `json:"beta" gorm:"type:decimal(10,6)"`              // Beta值
	TrackingError    decimal.Decimal `json:"tracking_error" gorm:"type:decimal(10,6)"`    // 跟踪误差
	InformationRatio decimal.Decimal `json:"information_ratio" gorm:"type:decimal(10,6)"` // 信息比率

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PortfolioPerformance) TableName() string {
	return "portfolio_performance"
}

// PortfolioRebalance 再平衡记录
type PortfolioRebalance struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	PortfolioID   uint            `json:"portfolio_id" gorm:"index"`            // 组合ID
	RebalanceDate time.Time       `json:"rebalance_date" gorm:"index"`          // 再平衡日期
	Reason        string          `json:"reason" gorm:"size:100"`               // 再平衡原因：scheduled/threshold/manual
	PreValue      decimal.Decimal `json:"pre_value" gorm:"type:decimal(15,2)"`  // 再平衡前市值
	PostValue     decimal.Decimal `json:"post_value" gorm:"type:decimal(15,2)"` // 再平衡后市值
	Cost          decimal.Decimal `json:"cost" gorm:"type:decimal(15,2)"`       // 再平衡成本（交易费用）
	TaxImpact     decimal.Decimal `json:"tax_impact" gorm:"type:decimal(15,2)"` // 税务影响

	// 变更明细（JSON存储）
	Changes JSONMap `json:"changes" gorm:"type:json"` // 持仓变更明细

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (PortfolioRebalance) TableName() string {
	return "portfolio_rebalances"
}
