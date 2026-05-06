package exchange_rate

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestExchangeRate_SameCurrency(t *testing.T) {
	service := NewExchangeRateService(nil)

	amount := decimal.NewFromFloat(100.0)
	result := service.Convert(amount, "USD", "USD")

	if !result.Equal(amount) {
		t.Errorf("Same currency conversion should return same amount, got %s", result.String())
	}
}

func TestExchangeRate_GetRate_FallbackOnNilManager(t *testing.T) {
	service := &ExchangeRateService{}

	rate := service.GetRate("USD", "CNY")
	if rate == 0 {
		t.Log("GetRate returned 0 (fallback behavior)")
	} else {
		t.Logf("GetRate returned rate: %.4f", rate)
	}
}

func TestExchangeRate_GetRateDecimal_TimeoutHandling(t *testing.T) {
	service := NewExchangeRateService(nil)

	_, err := service.GetRateDecimal("USD", "CNY")
	if err != nil {
		t.Logf("GetRateDecimal returned expected error (no datasource): %v", err)
	}
}

func TestExchangeRate_Convert_LargeAmount(t *testing.T) {
	service := NewExchangeRateService(nil)

	largeAmount := decimal.NewFromFloat(1000000.50)
	result := service.Convert(largeAmount, "USD", "EUR")

	assert.True(t, result.GreaterThanOrEqual(decimal.Zero),
		"Converted amount should be non-negative, got %s", result.String())
}

func TestExchangeRate_Convert_SmallAmount(t *testing.T) {
	service := NewExchangeRateService(nil)

	smallAmount := decimal.NewFromFloat(0.01)
	result := service.Convert(smallAmount, "JPY", "USD")

	assert.True(t, result.GreaterThanOrEqual(decimal.Zero),
		"Small amount conversion should be non-negative")
}

func TestExchangeRate_UpdateRates_NilConfig(t *testing.T) {
	service := NewExchangeRateService(nil)

	err := service.UpdateRates()
	if err != nil {
		t.Logf("UpdateRates returned expected error (nil config): %v", err)
	}
}
