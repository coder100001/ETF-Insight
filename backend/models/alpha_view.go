package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ViewType string

const (
	ViewTypeAbsolute ViewType = "absolute"
	ViewTypeRelative ViewType = "relative"
)

type ViewMethod string

const (
	ViewMethodFactorTiming  ViewMethod = "factor_timing"
	ViewMethodMomentum      ViewMethod = "momentum"
	ViewMethodMeanReversion ViewMethod = "mean_reversion"
)

type ViewStatus string

const (
	ViewStatusActive    ViewStatus = "active"
	ViewStatusExpired   ViewStatus = "expired"
	ViewStatusValidated ViewStatus = "validated"
)

type AlphaView struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	PortfolioID uint `json:"portfolio_id" gorm:"index"`

	AssetSymbol string          `json:"asset_symbol" gorm:"size:20;index:idx_view_asset"`
	ViewReturn  decimal.Decimal `json:"view_return" gorm:"type:decimal(10,6)"`
	Confidence  decimal.Decimal `json:"confidence" gorm:"type:decimal(5,2)"`

	ViewType   ViewType   `json:"view_type" gorm:"size:20"`
	ViewMethod ViewMethod `json:"view_method" gorm:"size:50"`

	GeneratedAt time.Time  `json:"generated_at"`
	ValidUntil  time.Time  `json:"valid_until"`
	Status      ViewStatus `json:"status" gorm:"size:20;default:'active'"`

	SourceFactor  string          `json:"source_factor" gorm:"size:20"`
	FactorLoading decimal.Decimal `json:"factor_loading" gorm:"type:decimal(10,6)"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Portfolio   Portfolio             `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
	Performance *AlphaViewPerformance `json:"performance,omitempty" gorm:"foreignKey:ViewID"`
}

func (AlphaView) TableName() string {
	return "alpha_views"
}

type AlphaViewPerformance struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	ViewID uint `json:"view_id" gorm:"index"`

	ActualReturn    decimal.Decimal `json:"actual_return" gorm:"type:decimal(10,6)"`
	PredictionError decimal.Decimal `json:"prediction_error" gorm:"type:decimal(10,6)"`

	IsValidated    bool      `json:"is_validated"`
	ValidationDate time.Time `json:"validation_date"`
	IsCorrect      bool      `json:"is_correct"`

	RollingWinRate decimal.Decimal `json:"rolling_win_rate" gorm:"type:decimal(5,2)"`

	CreatedAt time.Time `json:"created_at"`

	View AlphaView `json:"view,omitempty" gorm:"foreignKey:ViewID"`
}

func (AlphaViewPerformance) TableName() string {
	return "alpha_view_performances"
}

func (v ViewType) IsValid() bool {
	return v == ViewTypeAbsolute || v == ViewTypeRelative
}

func (m ViewMethod) IsValid() bool {
	switch m {
	case ViewMethodFactorTiming, ViewMethodMomentum, ViewMethodMeanReversion:
		return true
	default:
		return false
	}
}

func (s ViewStatus) IsValid() bool {
	switch s {
	case ViewStatusActive, ViewStatusExpired, ViewStatusValidated:
		return true
	default:
		return false
	}
}
