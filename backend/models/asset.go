package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// AssetType 资产类型
type AssetType string

const (
	AssetTypeStock AssetType = "stock" // 股票
	AssetTypeETF   AssetType = "etf"   // ETF
	AssetTypeIndex AssetType = "index" // 指数
	AssetTypeBond  AssetType = "bond"  // 债券
	AssetTypeCash  AssetType = "cash"  // 现金
	AssetTypeOther AssetType = "other" // 其他
)

// Asset 统一资产基表
// 作为所有可交易资产的基础抽象，支持股票、ETF、指数等多种资产类型
type Asset struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Symbol   string    `json:"symbol" gorm:"uniqueIndex;size:20"` // 资产代码
	Name     string    `json:"name" gorm:"size:100"`              // 资产名称
	Type     AssetType `json:"type" gorm:"size:20;index"`         // 资产类型
	Currency string    `json:"currency" gorm:"size:10"`           // 计价货币
	Exchange string    `json:"exchange" gorm:"size:20"`           // 交易所
	ISIN     string    `json:"isin" gorm:"size:20"`               // ISIN代码
	Country  string    `json:"country" gorm:"size:50"`            // 所属国家/地区
	Sector   string    `json:"sector" gorm:"size:50"`             // 行业分类
	Status   int       `json:"status" gorm:"default:1"`           // 1-正常，0-停用

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Asset) TableName() string {
	return "assets"
}

// AssetPrice 资产价格历史数据
// 统一存储所有资产类型的价格数据
type AssetPrice struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	AssetID    uint            `json:"asset_id" gorm:"index:idx_asset_date"`
	Symbol     string          `json:"symbol" gorm:"size:20"` // 冗余存储，便于查询
	Date       time.Time       `json:"date" gorm:"index:idx_asset_date"`
	OpenPrice  decimal.Decimal `json:"open_price" gorm:"type:decimal(20,8)"`
	ClosePrice decimal.Decimal `json:"close_price" gorm:"type:decimal(20,8)"`
	HighPrice  decimal.Decimal `json:"high_price" gorm:"type:decimal(20,8)"`
	LowPrice   decimal.Decimal `json:"low_price" gorm:"type:decimal(20,8)"`
	Volume     int64           `json:"volume"`
	DataSource string          `json:"data_source" gorm:"size:50"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (AssetPrice) TableName() string {
	return "asset_prices"
}

// AssetRelationship 资产关系表
// 用于存储资产之间的关联关系，如 ETF 与其跟踪指数的关系
type AssetRelationship struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	AssetID      uint            `json:"asset_id" gorm:"index"`           // 主资产ID
	RelatedID    uint            `json:"related_id" gorm:"index"`         // 关联资产ID
	RelationType string          `json:"relation_type" gorm:"size:20"`    // 关系类型：benchmark, constituent, etc.
	Weight       decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"` // 权重（如成分股权重）

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AssetRelationship) TableName() string {
	return "asset_relationships"
}
