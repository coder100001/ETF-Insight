package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PriorType string

const (
	PriorTypeEqualWeight PriorType = "equal_weight"
	PriorTypeMinVariance PriorType = "min_variance"
	PriorTypeMarketCap   PriorType = "market_cap"
)

type OmegaMethod string

const (
	OmegaMethodIdzorek     OmegaMethod = "Idzorek"
	OmegaMethodHeLitterman OmegaMethod = "HeLitterman"
)

type BlackLittermanConfig struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	PortfolioID uint `json:"portfolio_id" gorm:"uniqueIndex"`

	RiskAversion decimal.Decimal `json:"risk_aversion" gorm:"type:decimal(10,6)"`
	PriorType    PriorType       `json:"prior_type" gorm:"size:20"`

	PriorWeights   string `json:"prior_weights" gorm:"type:json"`
	ImpliedReturns string `json:"implied_returns" gorm:"type:json"`

	OmegaMethod OmegaMethod `json:"omega_method" gorm:"size:20"`
	OmegaMatrix string      `json:"omega_matrix" gorm:"type:json"`

	IsActive       bool      `json:"is_active" gorm:"default:true"`
	LastCalculated time.Time `json:"last_calculated"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Portfolio Portfolio `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
}

func (BlackLittermanConfig) TableName() string {
	return "black_litterman_configs"
}

type BLPosteriorReturn struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	ConfigID        uint      `json:"config_id" gorm:"index"`
	CalculationDate time.Time `json:"calculation_date" gorm:"index"`

	PosteriorReturns string `json:"posterior_returns" gorm:"type:json"`
	PosteriorWeights string `json:"posterior_weights" gorm:"type:json"`
	PosteriorCov     string `json:"posterior_cov" gorm:"type:json"`

	NumViews   int             `json:"num_views"`
	ViewImpact decimal.Decimal `json:"view_impact" gorm:"type:decimal(10,6)"`

	CreatedAt time.Time `json:"created_at"`

	Config BlackLittermanConfig `json:"config,omitempty" gorm:"foreignKey:ConfigID"`
}

func (BLPosteriorReturn) TableName() string {
	return "bl_posterior_returns"
}

func (p PriorType) IsValid() bool {
	switch p {
	case PriorTypeEqualWeight, PriorTypeMinVariance, PriorTypeMarketCap:
		return true
	default:
		return false
	}
}

func (o OmegaMethod) IsValid() bool {
	return o == OmegaMethodIdzorek || o == OmegaMethodHeLitterman
}
