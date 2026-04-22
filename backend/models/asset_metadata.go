package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// AssetMetadata 资产元数据表
// 用于存储资产的扩展信息，如行业、国家、因子暴露等
type AssetMetadata struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	AssetID   uint   `json:"asset_id" gorm:"uniqueIndex"` // 资产ID
	Sector    string `json:"sector" gorm:"size:50"`       // 行业分类
	Industry  string `json:"industry" gorm:"size:50"`     // 细分行业
	Country   string `json:"country" gorm:"size:50"`      // 所属国家
	Region    string `json:"region" gorm:"size:50"`       // 所属地区
	Style     string `json:"style" gorm:"size:20"`        // 风格：value/growth/blend
	MarketCap string `json:"market_cap" gorm:"size:20"`   // 市值分类：large/mid/small

	// 因子暴露（JSON格式存储，灵活扩展）
	FactorExposure string `json:"factor_exposure" gorm:"type:text"` // JSON: {"momentum": 0.5, "value": -0.3}

	// ESG评分
	ESGScore         decimal.Decimal `json:"esg_score" gorm:"type:decimal(5,2)"`
	EnvironmentScore decimal.Decimal `json:"environment_score" gorm:"type:decimal(5,2)"`
	SocialScore      decimal.Decimal `json:"social_score" gorm:"type:decimal(5,2)"`
	GovernanceScore  decimal.Decimal `json:"governance_score" gorm:"type:decimal(5,2)"`

	// 基本面数据
	PERatio       decimal.Decimal `json:"pe_ratio" gorm:"type:decimal(10,2)"`
	PBRatio       decimal.Decimal `json:"pb_ratio" gorm:"type:decimal(10,2)"`
	DividendYield decimal.Decimal `json:"dividend_yield" gorm:"type:decimal(5,2)"`
	Beta          decimal.Decimal `json:"beta" gorm:"type:decimal(5,2)"`

	DataSource string    `json:"data_source" gorm:"size:50"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (AssetMetadata) TableName() string {
	return "asset_metadata"
}

// SectorAllocation 行业配置表
// 用于存储资产的行业分布
type SectorAllocation struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	AssetID   uint            `json:"asset_id" gorm:"index:idx_sector_date"` // 资产ID
	Date      time.Time       `json:"date" gorm:"index:idx_sector_date"`     // 日期
	Sector    string          `json:"sector" gorm:"size:50"`                 // 行业名称
	Weight    decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`       // 权重(%)
	SubSector string          `json:"sub_sector" gorm:"size:50"`             // 子行业

	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (SectorAllocation) TableName() string {
	return "sector_allocations"
}

// GeographicAllocation 地理配置表
// 用于存储资产的国家/地区分布
type GeographicAllocation struct {
	ID      uint            `json:"id" gorm:"primaryKey"`
	AssetID uint            `json:"asset_id" gorm:"index:idx_geo_date"` // 资产ID
	Date    time.Time       `json:"date" gorm:"index:idx_geo_date"`     // 日期
	Country string          `json:"country" gorm:"size:50"`             // 国家
	Region  string          `json:"region" gorm:"size:50"`              // 地区
	Weight  decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`    // 权重(%)

	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (GeographicAllocation) TableName() string {
	return "geographic_allocations"
}
