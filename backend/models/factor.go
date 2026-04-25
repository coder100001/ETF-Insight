package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type FactorType string

const (
	FactorTypeMarket FactorType = "Mkt-RF"
	FactorTypeSMB    FactorType = "SMB"
	FactorTypeHML    FactorType = "HML"
	FactorTypeRMW    FactorType = "RMW"
	FactorTypeCMA    FactorType = "CMA"
)

type FactorData struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	FactorName string          `json:"factor_name" gorm:"size:20;index:idx_factor_date"`
	Date       time.Time       `json:"date" gorm:"index:idx_factor_date"`
	Value      decimal.Decimal `json:"value" gorm:"type:decimal(10,6)"`
	DataSource string          `json:"data_source" gorm:"size:50"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (FactorData) TableName() string {
	return "factor_data"
}

type SignalStrength string

const (
	SignalStrengthStrongPositive SignalStrength = "strong_positive"
	SignalStrengthWeakPositive   SignalStrength = "weak_positive"
	SignalStrengthNeutral        SignalStrength = "neutral"
	SignalStrengthWeakNegative   SignalStrength = "weak_negative"
	SignalStrengthStrongNegative SignalStrength = "strong_negative"
)

type FactorTimingSignal struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FactorName string    `json:"factor_name" gorm:"size:20;index:idx_signal_date"`
	SignalDate time.Time `json:"signal_date" gorm:"index:idx_signal_date"`

	MASlope60  decimal.Decimal `json:"ma_slope_60" gorm:"type:decimal(10,6)"`
	ZScore     decimal.Decimal `json:"z_score" gorm:"type:decimal(10,6)"`
	Percentile decimal.Decimal `json:"percentile" gorm:"type:decimal(5,2)"`

	SignalStrength SignalStrength `json:"signal_strength" gorm:"size:20"`
	SignalScore    int            `json:"signal_score"`

	ExpectedReturn decimal.Decimal `json:"expected_return" gorm:"type:decimal(10,6)"`
	Confidence     decimal.Decimal `json:"confidence" gorm:"type:decimal(5,2)"`

	CreatedAt time.Time `json:"created_at"`
}

func (FactorTimingSignal) TableName() string {
	return "factor_timing_signals"
}

func (s SignalStrength) IsValid() bool {
	switch s {
	case SignalStrengthStrongPositive, SignalStrengthWeakPositive,
		SignalStrengthNeutral, SignalStrengthWeakNegative,
		SignalStrengthStrongNegative:
		return true
	default:
		return false
	}
}

func (s SignalStrength) ToScore() int {
	switch s {
	case SignalStrengthStrongPositive:
		return 2
	case SignalStrengthWeakPositive:
		return 1
	case SignalStrengthNeutral:
		return 0
	case SignalStrengthWeakNegative:
		return -1
	case SignalStrengthStrongNegative:
		return -2
	default:
		return 0
	}
}

func ScoreToSignalStrength(score int) SignalStrength {
	switch {
	case score >= 2:
		return SignalStrengthStrongPositive
	case score == 1:
		return SignalStrengthWeakPositive
	case score == 0:
		return SignalStrengthNeutral
	case score == -1:
		return SignalStrengthWeakNegative
	default:
		return SignalStrengthStrongNegative
	}
}
