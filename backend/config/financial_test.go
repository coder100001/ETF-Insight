package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFinancialConfig(t *testing.T) {
	cfg := GetFinancialConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 0.0435, cfg.RiskFreeRate)
	assert.Equal(t, 252, cfg.TradingDaysYear)
	assert.Equal(t, "USD", cfg.DefaultCurrency)
}

func TestSetRiskFreeRate(t *testing.T) {
	SetRiskFreeRate(0.05)
	assert.Equal(t, 0.05, GetFinancialConfig().RiskFreeRate)
	SetRiskFreeRate(0.0435)
}

func TestSetTradingDaysYear(t *testing.T) {
	SetTradingDaysYear(365)
	assert.Equal(t, 365, GetFinancialConfig().TradingDaysYear)
	SetTradingDaysYear(252)
}

func TestSingleton(t *testing.T) {
	a := GetFinancialConfig()
	b := GetFinancialConfig()
	assert.Same(t, a, b)
}
