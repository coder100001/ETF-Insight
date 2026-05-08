package quantlib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestEuropeanOptionBoundaryZeroSpot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.QuantLibAPIResponse{
			Success: false,
			Message: "validation error: spot must be greater than 0",
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.Zero,
		Strike:       decimal.NewFromFloat(100),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.PriceEuropeanOption(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation error")
}

func TestVaRMinimumReturns(t *testing.T) {
	err := models.ValidateVaRRequest(models.VaRRequest{
		PortfolioValue: decimal.NewFromFloat(1000000),
		Returns:        []decimal.Decimal{decimal.NewFromFloat(0.01)},
		Confidence:     decimal.NewFromFloat(0.95),
		HoldingPeriod:  1,
		Method:         "historical",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 elements")
}

func TestBondFrequencyBounds(t *testing.T) {
	err := models.ValidateBondRequest(models.BondRequest{
		FaceValue:       decimal.NewFromFloat(1000),
		CouponRate:      decimal.NewFromFloat(0.05),
		Frequency:       0,
		Maturity:        "2030-01-01",
		YieldToMaturity: decimal.NewFromFloat(0.04),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 12")

	err = models.ValidateBondRequest(models.BondRequest{
		FaceValue:       decimal.NewFromFloat(1000),
		CouponRate:      decimal.NewFromFloat(0.05),
		Frequency:       13,
		Maturity:        "2030-01-01",
		YieldToMaturity: decimal.NewFromFloat(0.04),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 12")
}

func TestYieldCurveEmptyTenors(t *testing.T) {
	err := models.ValidateYieldCurveRequest(models.YieldCurveRequest{
		Currency: "USD",
		Tenors:   []string{},
		Rates:    []decimal.Decimal{decimal.NewFromFloat(0.05)},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenors cannot be empty")
}

func TestYieldCurveMismatchedLengths(t *testing.T) {
	err := models.ValidateYieldCurveRequest(models.YieldCurveRequest{
		Currency: "USD",
		Tenors:   []string{"1M", "3M", "6M"},
		Rates:    []decimal.Decimal{decimal.NewFromFloat(0.04), decimal.NewFromFloat(0.045)},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same length")
}

func TestInvalidVaRMethod(t *testing.T) {
	err := models.ValidateVaRRequest(models.VaRRequest{
		PortfolioValue: decimal.NewFromFloat(1000000),
		Returns: []decimal.Decimal{
			decimal.NewFromFloat(0.01),
			decimal.NewFromFloat(-0.02),
		},
		Confidence:    decimal.NewFromFloat(0.95),
		HoldingPeriod: 1,
		Method:        "invalid_method",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 'historical', 'parametric', or 'monte_carlo'")
}
