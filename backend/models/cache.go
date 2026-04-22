package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ETFOverlapCache ETF重叠度缓存表
// 用于存储预计算的ETF重叠结果，避免重复计算
type ETFOverlapCache struct {
	ID             uint            `json:"id" gorm:"primaryKey"`
	ETFAID         uint            `json:"etf_a_id" gorm:"index:idx_overlap_cache_pair"`                // ETF A ID
	ETFBID         uint            `json:"etf_b_id" gorm:"index:idx_overlap_cache_pair"`                // ETF B ID
	ETFASymbol     string          `json:"etf_a_symbol" gorm:"size:20;index:idx_overlap_cache_symbols"` // ETF A 代码
	ETFBSymbol     string          `json:"etf_b_symbol" gorm:"size:20;index:idx_overlap_cache_symbols"` // ETF B 代码
	OverlapScore   decimal.Decimal `json:"overlap_score" gorm:"type:decimal(5,2)"`                      // 重叠度分数(0-100)
	CommonHoldings int             `json:"common_holdings"`                                             // 共同持仓数量
	TotalWeightA   decimal.Decimal `json:"total_weight_a" gorm:"type:decimal(5,2)"`                     // A中重叠权重
	TotalWeightB   decimal.Decimal `json:"total_weight_b" gorm:"type:decimal(5,2)"`                     // B中重叠权重
	HoldingsDate   time.Time       `json:"holdings_date" gorm:"index"`                                  // 持仓日期
	CalculatedAt   time.Time       `json:"calculated_at"`                                               // 计算时间
	ExpiresAt      time.Time       `json:"expires_at" gorm:"index"`                                     // 缓存过期时间
	DataVersion    int             `json:"data_version" gorm:"default:1"`                               // 数据版本号

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ETFOverlapCache) TableName() string {
	return "etf_overlap_cache"
}

// IsExpired 检查缓存是否过期
func (c *ETFOverlapCache) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// CacheInvalidationLog 缓存失效日志
// 记录缓存失效事件，用于审计和调试
type CacheInvalidationLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CacheType   string    `json:"cache_type" gorm:"size:50"`   // 缓存类型：overlap, penetration, etc.
	CacheKey    string    `json:"cache_key" gorm:"size:100"`   // 缓存键
	Reason      string    `json:"reason" gorm:"size:200"`      // 失效原因
	TriggeredBy string    `json:"triggered_by" gorm:"size:50"` // 触发源：api, event, manual
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (CacheInvalidationLog) TableName() string {
	return "cache_invalidation_logs"
}
