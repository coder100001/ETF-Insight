package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// AssetClass 资产类别
type AssetClass string

const (
	AssetClassEquity      AssetClass = "equity"      // 股票
	AssetClassBond        AssetClass = "bond"        // 债券
	AssetClassCommodity   AssetClass = "commodity"   // 商品
	AssetClassREIT        AssetClass = "reit"        // 房地产信托
	AssetClassCurrency    AssetClass = "currency"    // 货币
	AssetClassMultiAsset  AssetClass = "multi_asset" // 多资产
	AssetClassAlternative AssetClass = "alternative" // 另类投资
)

// Region 地区
type Region string

const (
	RegionGlobal       Region = "global"        // 全球
	RegionUS           Region = "us"            // 美国
	RegionChina        Region = "china"         // 中国
	RegionEurope       Region = "europe"        // 欧洲
	RegionJapan        Region = "japan"         // 日本
	RegionEmerging     Region = "emerging"      // 新兴市场
	RegionAsiaPacific  Region = "asia_pacific"  // 亚太
	RegionLatinAmerica Region = "latin_america" // 拉美
)

// ETFType ETF类型
type ETFType string

const (
	ETFTypeIndex     ETFType = "index"     // 指数基金
	ETFTypeSector    ETFType = "sector"    // 行业ETF
	ETFTypeFactor    ETFType = "factor"    // 因子ETF
	ETFTypeThematic  ETFType = "thematic"  // 主题ETF
	ETFTypeActive    ETFType = "active"    // 主动管理ETF
	ETFTypeLeveraged ETFType = "leveraged" // 杠杆ETF
	ETFTypeInverse   ETFType = "inverse"   // 反向ETF
)

// UniversalETF 通用ETF模型
type UniversalETF struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	AssetID  uint   `json:"asset_id" gorm:"index"`             // 关联Asset表
	Symbol   string `json:"symbol" gorm:"uniqueIndex;size:20"` // ETF代码
	Name     string `json:"name" gorm:"size:100"`              // ETF名称
	FullName string `json:"full_name" gorm:"size:200"`         // 完整名称

	// 分类信息
	AssetClass AssetClass `json:"asset_class" gorm:"size:20"` // 资产类别
	Region     Region     `json:"region" gorm:"size:20"`      // 地区
	ETFType    ETFType    `json:"etf_type" gorm:"size:20"`    // ETF类型
	Sector     string     `json:"sector" gorm:"size:50"`      // 行业板块

	// 基本信息
	Exchange  string `json:"exchange" gorm:"size:20"`   // 交易所
	Currency  string `json:"currency" gorm:"size:10"`   // 计价货币
	ISIN      string `json:"isin" gorm:"size:20"`       // ISIN代码
	Benchmark string `json:"benchmark" gorm:"size:100"` // 跟踪指数
	Provider  string `json:"provider" gorm:"size:50"`   // 发行商

	// 费率信息
	ExpenseRatio  decimal.Decimal `json:"expense_ratio" gorm:"type:decimal(5,4)"`  // 管理费率
	TrackingError decimal.Decimal `json:"tracking_error" gorm:"type:decimal(5,4)"` // 跟踪误差

	// 规模信息
	AUM               decimal.Decimal `json:"aum" gorm:"type:decimal(20,2)"` // 资产管理规模
	SharesOutstanding int64           `json:"shares_outstanding"`            // 流通份额

	// 交易信息
	CurrentPrice   decimal.Decimal `json:"current_price" gorm:"type:decimal(10,4)"`
	PreviousClose  decimal.Decimal `json:"previous_close" gorm:"type:decimal(10,4)"`
	PriceChange    decimal.Decimal `json:"price_change" gorm:"type:decimal(10,4)"`
	PriceChangePct decimal.Decimal `json:"price_change_pct" gorm:"type:decimal(5,2)"`
	Volume         int64           `json:"volume"`
	Turnover       decimal.Decimal `json:"turnover" gorm:"type:decimal(15,2)"`
	NAV            decimal.Decimal `json:"nav" gorm:"type:decimal(10,4)"`         // 净值
	PremiumRate    decimal.Decimal `json:"premium_rate" gorm:"type:decimal(5,2)"` // 溢价率

	// 风险指标
	Volatility3M  decimal.Decimal `json:"volatility_3m" gorm:"type:decimal(5,2)"`   // 3月波动率
	Volatility1Y  decimal.Decimal `json:"volatility_1y" gorm:"type:decimal(5,2)"`   // 1年波动率
	SharpeRatio1Y decimal.Decimal `json:"sharpe_ratio_1y" gorm:"type:decimal(5,2)"` // 1年夏普比率
	MaxDrawdown1Y decimal.Decimal `json:"max_drawdown_1y" gorm:"type:decimal(5,2)"` // 1年最大回撤

	// 收益指标
	Return1M  decimal.Decimal `json:"return_1m" gorm:"type:decimal(5,2)"`  // 1月收益
	Return3M  decimal.Decimal `json:"return_3m" gorm:"type:decimal(5,2)"`  // 3月收益
	Return6M  decimal.Decimal `json:"return_6m" gorm:"type:decimal(5,2)"`  // 6月收益
	Return1Y  decimal.Decimal `json:"return_1y" gorm:"type:decimal(5,2)"`  // 1年收益
	Return3Y  decimal.Decimal `json:"return_3y" gorm:"type:decimal(5,2)"`  // 3年收益
	Return5Y  decimal.Decimal `json:"return_5y" gorm:"type:decimal(5,2)"`  // 5年收益
	ReturnYTD decimal.Decimal `json:"return_ytd" gorm:"type:decimal(5,2)"` // 年初至今收益

	// 分红信息
	DividendYield     decimal.Decimal `json:"dividend_yield" gorm:"type:decimal(5,2)"` // 股息率
	DividendFrequency string          `json:"dividend_frequency" gorm:"size:10"`       // 分红频率

	// 状态
	Status      int       `json:"status" gorm:"default:1"`    // 1-正常，0-停用
	DataSource  string    `json:"data_source" gorm:"size:50"` // 数据源
	LastUpdated time.Time `json:"last_updated"`               // 最后更新时间

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UniversalETF) TableName() string {
	return "universal_etfs"
}

// ETFClassification ETF分类
type ETFClassification struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	ETFID       uint   `json:"etf_id" gorm:"index"`
	Category    string `json:"category" gorm:"size:50"`     // 细分类别
	SubCategory string `json:"sub_category" gorm:"size:50"` // 子类别
	Tags        string `json:"tags" gorm:"size:200"`        // 标签，逗号分隔
	RiskLevel   int    `json:"risk_level"`                  // 风险等级 1-5

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ETFConstituent ETF成分股
type ETFConstituent struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	ETFID       uint            `json:"etf_id" gorm:"index:idx_etf_constituent"`
	Symbol      string          `json:"symbol" gorm:"size:20;index:idx_etf_constituent"`
	Name        string          `json:"name" gorm:"size:100"`
	Weight      decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"` // 权重(%)
	Shares      int64           `json:"shares"`                          // 持股数量
	MarketValue decimal.Decimal `json:"market_value" gorm:"type:decimal(15,2)"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ETFHistoricalData ETF历史数据
type ETFHistoricalData struct {
	ID       uint            `json:"id" gorm:"primaryKey"`
	ETFID    uint            `json:"etf_id" gorm:"index:idx_etf_historical"`
	Symbol   string          `json:"symbol" gorm:"size:20;index:idx_etf_historical"`
	Date     time.Time       `json:"date" gorm:"index:idx_etf_historical"`
	Open     decimal.Decimal `json:"open" gorm:"type:decimal(10,4)"`
	High     decimal.Decimal `json:"high" gorm:"type:decimal(10,4)"`
	Low      decimal.Decimal `json:"low" gorm:"type:decimal(10,4)"`
	Close    decimal.Decimal `json:"close" gorm:"type:decimal(10,4)"`
	Volume   int64           `json:"volume"`
	Turnover decimal.Decimal `json:"turnover" gorm:"type:decimal(15,2)"`
	NAV      decimal.Decimal `json:"nav" gorm:"type:decimal(10,4)"`

	CreatedAt time.Time `json:"created_at"`
}

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

// ETFAssetAllocation ETF资产配置
type ETFAssetAllocation struct {
	ID    uint `json:"id" gorm:"primaryKey"`
	ETFID uint `json:"etf_id" gorm:"uniqueIndex"`

	// 资产类别配置
	Stocks decimal.Decimal `json:"stocks" gorm:"type:decimal(5,2)"` // 股票占比
	Bonds  decimal.Decimal `json:"bonds" gorm:"type:decimal(5,2)"`  // 债券占比
	Cash   decimal.Decimal `json:"cash" gorm:"type:decimal(5,2)"`   // 现金占比
	Others decimal.Decimal `json:"others" gorm:"type:decimal(5,2)"` // 其他占比

	// 地区配置
	US           decimal.Decimal `json:"us" gorm:"type:decimal(5,2)"`            // 美国
	China        decimal.Decimal `json:"china" gorm:"type:decimal(5,2)"`         // 中国
	Europe       decimal.Decimal `json:"europe" gorm:"type:decimal(5,2)"`        // 欧洲
	Japan        decimal.Decimal `json:"japan" gorm:"type:decimal(5,2)"`         // 日本
	Emerging     decimal.Decimal `json:"emerging" gorm:"type:decimal(5,2)"`      // 新兴市场
	OtherRegions decimal.Decimal `json:"other_regions" gorm:"type:decimal(5,2)"` // 其他地区

	// 行业配置
	Technology   decimal.Decimal `json:"technology" gorm:"type:decimal(5,2)"`    // 科技
	Healthcare   decimal.Decimal `json:"healthcare" gorm:"type:decimal(5,2)"`    // 医疗
	Financials   decimal.Decimal `json:"financials" gorm:"type:decimal(5,2)"`    // 金融
	Consumer     decimal.Decimal `json:"consumer" gorm:"type:decimal(5,2)"`      // 消费
	Energy       decimal.Decimal `json:"energy" gorm:"type:decimal(5,2)"`        // 能源
	Industrials  decimal.Decimal `json:"industrials" gorm:"type:decimal(5,2)"`   // 工业
	OtherSectors decimal.Decimal `json:"other_sectors" gorm:"type:decimal(5,2)"` // 其他行业

	UpdatedAt time.Time `json:"updated_at"`
}

// ETFComparison ETF对比
type ETFComparison struct {
	ETF1ID     uint   `json:"etf1_id"`
	ETF2ID     uint   `json:"etf2_id"`
	ETF1Symbol string `json:"etf1_symbol"`
	ETF2Symbol string `json:"etf2_symbol"`

	// 对比指标
	Correlation  decimal.Decimal `json:"correlation" gorm:"type:decimal(5,2)"`    // 相关性
	Beta         decimal.Decimal `json:"beta" gorm:"type:decimal(5,2)"`           // Beta值
	TrackingDiff decimal.Decimal `json:"tracking_diff" gorm:"type:decimal(5,2)"`  // 跟踪差异
	ExpenseDiff  decimal.Decimal `json:"expense_diff" gorm:"type:decimal(5,2)"`   // 费率差异
	ReturnDiff1Y decimal.Decimal `json:"return_diff_1y" gorm:"type:decimal(5,2)"` // 1年收益差异

	CreatedAt time.Time `json:"created_at"`
}
