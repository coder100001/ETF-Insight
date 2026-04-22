package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// HoldingType 持仓类型
type HoldingType string

const (
	HoldingTypeETFConstituent HoldingType = "etf_constituent" // ETF成分股持仓
	HoldingTypePortfolio      HoldingType = "portfolio"       // 投资组合持仓
	HoldingTypeIndex          HoldingType = "index"           // 指数成分股
	HoldingTypeFund           HoldingType = "fund"            // 基金持仓
)

// Holding 资产持仓关系
// 核心表：记录资产间的包含关系（如ETF持有股票、组合持有ETF）
type Holding struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	ParentID  uint        `json:"parent_id" gorm:"index:idx_parent_child"`         // 父资产ID（ETF/组合/指数）
	ChildID   uint        `json:"child_id" gorm:"index:idx_parent_child"`          // 子资产ID（股票/ETF/债券）
	ParentType HoldingType `json:"parent_type" gorm:"size:20;index:idx_parent_child"` // 父资产类型
	ChildType  AssetType  `json:"child_type" gorm:"size:20"`                       // 子资产类型

	// 持仓权重信息
	Weight     decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`           // 权重（百分比）
	Shares     int64           `json:"shares,omitempty"`                          // 持股数量
	MarketValue decimal.Decimal `json:"market_value,omitempty" gorm:"type:decimal(20,2)"` // 持仓市值
	NotionalValue decimal.Decimal `json:"notional_value,omitempty" gorm:"type:decimal(20,2)"` // 名义价值

	// 时间维度
	EffectiveDate time.Time `json:"effective_date" gorm:"index:idx_effective_date"` // 生效日期（持仓数据发布日期）
	ReportDate    time.Time `json:"report_date"`                                    // 报告日期（季度报告日期）
	IsLatest      bool      `json:"is_latest" gorm:"default:false"`                 // 是否为最新持仓
	IsHistorical  bool      `json:"is_historical" gorm:"default:false"`             // 是否为历史持仓

	// 来源和验证
	Source       string `json:"source" gorm:"size:50"`                            // 数据来源
	IsEstimated  bool   `json:"is_estimated" gorm:"default:false"`                // 是否为估算数据
	Confidence   int    `json:"confidence" gorm:"default:80"`                     // 数据置信度 0-100

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联（可选）
	ParentAsset Asset `json:"parent_asset,omitempty" gorm:"foreignKey:ParentID"`
	ChildAsset  Asset `json:"child_asset,omitempty" gorm:"foreignKey:ChildID"`
}

// TableName 指定表名
func (Holding) TableName() string {
	return "holdings"
}

// HoldingSnapshot 持仓快照表（物化视图）
// 用于快速查询特定日期的持仓结构
type HoldingSnapshot struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	ParentID     uint            `json:"parent_id" gorm:"index:idx_snapshot_parent"`
	ParentType   HoldingType     `json:"parent_type" gorm:"size:20"`
	AsOfDate     time.Time       `json:"as_of_date" gorm:"index:idx_snapshot_date"` // 快照日期
	TotalHoldings int            `json:"total_holdings"`                            // 总持仓数量
	TotalWeight   decimal.Decimal `json:"total_weight" gorm:"type:decimal(5,2)"`     // 总权重（应为100%）
	TotalValue    decimal.Decimal `json:"total_value" gorm:"type:decimal(20,2)"`     // 总市值

	// 分类统计（物化字段）
	StockWeight  decimal.Decimal `json:"stock_weight" gorm:"type:decimal(5,2)"`      // 股票权重
	BondWeight   decimal.Decimal `json:"bond_weight" gorm:"type:decimal(5,2)"`       // 债券权重
	CashWeight   decimal.Decimal `json:"cash_weight" gorm:"type:decimal(5,2)"`       // 现金权重
	OtherWeight  decimal.Decimal `json:"other_weight" gorm:"type:decimal(5,2)"`      // 其他权重

	// 地区分布
	USWeight       decimal.Decimal `json:"us_weight" gorm:"type:decimal(5,2)"`       // 美国权重
	ChinaWeight    decimal.Decimal `json:"china_weight" gorm:"type:decimal(5,2)"`    // 中国权重
	EuropeWeight   decimal.Decimal `json:"europe_weight" gorm:"type:decimal(5,2)"`   // 欧洲权重
	JapanWeight    decimal.Decimal `json:"japan_weight" gorm:"type:decimal(5,2)"`    // 日本权重
	EmergingWeight decimal.Decimal `json:"emerging_weight" gorm:"type:decimal(5,2)"` // 新兴市场权重

	// 行业分布
	TechWeight     decimal.Decimal `json:"tech_weight" gorm:"type:decimal(5,2)"`     // 科技行业权重
	FinanceWeight  decimal.Decimal `json:"finance_weight" gorm:"type:decimal(5,2)"`  // 金融行业权重
	HealthcareWeight decimal.Decimal `json:"healthcare_weight" gorm:"type:decimal(5,2)"` // 医疗行业权重
	ConsumerWeight decimal.Decimal `json:"consumer_weight" gorm:"type:decimal(5,2)"` // 消费行业权重
	IndustrialWeight decimal.Decimal `json:"industrial_weight" gorm:"type:decimal(5,2)"` // 工业行业权重
	EnergyWeight   decimal.Decimal `json:"energy_weight" gorm:"type:decimal(5,2)"`   // 能源行业权重

	// 风险指标
	AvgBeta        decimal.Decimal `json:"avg_beta" gorm:"type:decimal(5,2)"`        // 平均Beta值
	AvgPERatio     decimal.Decimal `json:"avg_pe_ratio" gorm:"type:decimal(10,2)"`   // 平均市盈率
	AvgDividendYield decimal.Decimal `json:"avg_dividend_yield" gorm:"type:decimal(5,2)"` // 平均股息率

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (HoldingSnapshot) TableName() string {
	return "holding_snapshots"
}

// ETFHoldingsReport ETF持仓报告（按日期归档）
type ETFHoldingsReport struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	ETFID      uint            `json:"etf_id" gorm:"index:idx_etf_report"`          // ETF资产ID
	ReportDate time.Time       `json:"report_date" gorm:"index:idx_etf_report"`     // 报告日期
	ReportType string          `json:"report_type" gorm:"size:20"`                  // 报告类型：quarterly/annually
	HoldingsCount int          `json:"holdings_count"`                              // 持仓数量
	Top10Weight  decimal.Decimal `json:"top10_weight" gorm:"type:decimal(5,2)"`     // 前十大持仓权重
	TurnoverRate decimal.Decimal `json:"turnover_rate" gorm:"type:decimal(5,2)"`    // 换手率
	TrackingError decimal.Decimal `json:"tracking_error" gorm:"type:decimal(5,2)"`  // 跟踪误差

	// 文件信息（如果从PDF/HTML解析）
	SourceURL   string `json:"source_url" gorm:"size:200"`                        // 来源URL
	FilePath    string `json:"file_path" gorm:"size:200"`                         // 本地文件路径
	ParseStatus string `json:"parse_status" gorm:"size:20;default:'pending'"`     // 解析状态：pending/processing/completed/failed
	ParseError  string `json:"parse_error,omitempty" gorm:"size:500"`             // 解析错误信息

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ETFHoldingsReport) TableName() string {
	return "etf_holdings_reports"
}

// HoldingChange 持仓变动记录
type HoldingChange struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	ParentID      uint            `json:"parent_id" gorm:"index"`                 // 父资产ID
	ChildID       uint            `json:"child_id" gorm:"index"`                  // 子资产ID
	PeriodStart   time.Time       `json:"period_start"`                           // 期间开始
	PeriodEnd     time.Time       `json:"period_end"`                             // 期间结束
	PrevWeight    decimal.Decimal `json:"prev_weight" gorm:"type:decimal(5,2)"`   // 上期权重
	CurrentWeight decimal.Decimal `json:"current_weight" gorm:"type:decimal(5,2)"` // 本期权重
	WeightChange  decimal.Decimal `json:"weight_change" gorm:"type:decimal(5,2)"` // 权重变化
	ChangeType    string          `json:"change_type" gorm:"size:20"`             // 变化类型：new/addition/reduction/removal

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (HoldingChange) TableName() string {
	return "holding_changes"
}