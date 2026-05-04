package quantlib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientDefaultURL(t *testing.T) {
	os.Unsetenv("QUANTLIB_API_URL")
	os.Unsetenv("QUANTLIB_API_KEY")

	client := NewClient()

	assert.Equal(t, "https://api.fincept.in/quantlib", client.baseURL)
	assert.Empty(t, client.apiKey)
	assert.NotNil(t, client.httpClient)
}

func TestNewClientCustomURL(t *testing.T) {
	os.Setenv("QUANTLIB_API_URL", "http://localhost:9090/quantlib")
	os.Setenv("QUANTLIB_API_KEY", "test-key-123")
	defer os.Unsetenv("QUANTLIB_API_URL")
	defer os.Unsetenv("QUANTLIB_API_KEY")

	client := NewClient()

	assert.Equal(t, "http://localhost:9090/quantlib", client.baseURL)
	assert.Equal(t, "test-key-123", client.apiKey)
}

func TestPriceEuropeanOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/options/european", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "ETF-Insight/1.0", r.Header.Get("User-Agent"))

		var req models.EuropeanOptionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.Spot.Equal(decimal.NewFromFloat(100)))

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]interface{}{
				"price": "10.4506",
				"delta": "0.5948",
				"gamma": "0.0188",
				"theta": "-0.0128",
				"vega":  "0.3756",
				"rho":   "0.4502",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.PriceEuropeanOption(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(10.4506)))
	assert.True(t, result.Delta.Equal(decimal.NewFromFloat(0.5948)))
	assert.True(t, result.Gamma.Equal(decimal.NewFromFloat(0.0188)))
	assert.True(t, result.Theta.Equal(decimal.NewFromFloat(-0.0128)))
	assert.True(t, result.Vega.Equal(decimal.NewFromFloat(0.3756)))
	assert.True(t, result.Rho.Equal(decimal.NewFromFloat(0.4502)))
}

func TestPriceEuropeanOptionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.QuantLibAPIResponse{
			Success: false,
			Message: "internal server error",
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.PriceEuropeanOption(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "500")
}

func TestPriceEuropeanOptionAPIFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.QuantLibAPIResponse{
			Success: false,
			Message: "invalid parameter: spot must be positive",
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.NewFromFloat(-1),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.PriceEuropeanOption(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid parameter")
}

func TestPriceBond(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/bonds/fixed", r.URL.Path)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]interface{}{
				"dirty_price":       "103.2500",
				"clean_price":       "102.5000",
				"duration":          "4.5230",
				"modified_duration": "4.3510",
				"convexity":         "22.1500",
				"yield_to_maturity": "0.0450",
				"accrued_interest":  "0.7500",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.BondRequest{
		FaceValue:       decimal.NewFromFloat(100),
		CouponRate:      decimal.NewFromFloat(0.05),
		Frequency:       2,
		Maturity:        "2030-01-01",
		YieldToMaturity: decimal.NewFromFloat(0.045),
	}

	result, err := client.PriceBond(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.DirtyPrice.Equal(decimal.NewFromFloat(103.25)))
	assert.True(t, result.CleanPrice.Equal(decimal.NewFromFloat(102.5)))
	assert.True(t, result.Duration.Equal(decimal.NewFromFloat(4.523)))
	assert.True(t, result.ModifiedDuration.Equal(decimal.NewFromFloat(4.351)))
	assert.True(t, result.Convexity.Equal(decimal.NewFromFloat(22.15)))
	assert.True(t, result.YieldToMaturity.Equal(decimal.NewFromFloat(0.045)))
	assert.True(t, result.AccruedInterest.Equal(decimal.NewFromFloat(0.75)))
}

func TestCalculateVaR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/risk/var", r.URL.Path)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]interface{}{
				"var":            "15230.45",
				"cvar":           "18950.20",
				"confidence":     "0.95",
				"holding_period": 10,
				"method":         "historical",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.VaRRequest{
		PortfolioValue: decimal.NewFromFloat(1000000),
		Returns: []decimal.Decimal{
			decimal.NewFromFloat(0.01),
			decimal.NewFromFloat(-0.02),
			decimal.NewFromFloat(0.005),
		},
		Confidence:    decimal.NewFromFloat(0.95),
		HoldingPeriod: 10,
		Method:        "historical",
	}

	result, err := client.CalculateVaR(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.VaR.Equal(decimal.NewFromFloat(15230.45)))
	assert.True(t, result.CVaR.Equal(decimal.NewFromFloat(18950.20)))
	assert.True(t, result.Confidence.Equal(decimal.NewFromFloat(0.95)))
	assert.Equal(t, 10, result.HoldingPeriod)
	assert.Equal(t, "historical", result.Method)
}

func TestPriceEuropeanOptionDirectParse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"price": "5.50",
			"delta": "0.45",
			"gamma": "0.02",
			"theta": "-0.01",
			"vega":  "0.30",
			"rho":   "0.25",
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(100),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(0.5),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.PriceEuropeanOption(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(5.50)))
}

func TestAPIKeyHeader(t *testing.T) {
	var receivedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.Header.Get("X-API-Key")

		resp := models.QuantLibAPIResponse{
			Success: true,
			Data: map[string]interface{}{
				"price": "1.0", "delta": "0.5", "gamma": "0.01",
				"theta": "-0.01", "vega": "0.1", "rho": "0.05",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		apiKey:     "my-secret-key",
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(100),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	_, err := client.PriceEuropeanOption(req)
	require.NoError(t, err)
	assert.Equal(t, "my-secret-key", receivedKey)
}
