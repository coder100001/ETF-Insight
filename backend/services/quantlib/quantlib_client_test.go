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
			Data: map[string]any{
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
		assert.Equal(t, "/bonds/price", r.URL.Path)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
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
			Data: map[string]any{
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
		json.NewEncoder(w).Encode(map[string]any{
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
			Data: map[string]any{
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

func TestDecimalSerializationPrecision(t *testing.T) {
	// TDD RED: This test verifies that decimal fields are serialized as strings
	// in JSON to prevent JavaScript precision loss.
	//
	// Current behavior: decimal.Decimal serializes as JSON number: {"price": 10.4506}
	// Desired behavior: decimal.Decimal serializes as JSON string: {"price": "10.4506"}
	//
	// Why: JavaScript's number type (IEEE 754 float64) loses precision for
	// values with more than 15-17 significant digits. Financial calculations
	// require exact decimal representation.

	result := models.OptionResult{
		Price: decimal.RequireFromString("10.4506"),
		Delta: decimal.RequireFromString("0.594823456789012345"),
		Gamma: decimal.RequireFromString("0.0188"),
		Theta: decimal.RequireFromString("-0.0128"),
		Vega:  decimal.RequireFromString("0.3756"),
		Rho:   decimal.RequireFromString("0.4502"),
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)

	// Verify decimal values are quoted strings, not raw numbers
	// Current: {"price":10.4506,...} (number - FAILS)
	// Expected: {"price":"10.4506",...} (string - PASSES)
	assert.Contains(t, jsonStr, `"price":"10.4506"`, "Price should be serialized as JSON string")
	assert.Contains(t, jsonStr, `"delta":"0.594823456789012345"`, "Delta should preserve full precision as string")
	assert.NotContains(t, jsonStr, `"price":10.4506`, "Price should NOT be a raw JSON number")

	// Verify round-trip: deserialize back should preserve exact value
	var parsed models.OptionResult
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)
	assert.True(t, parsed.Price.Equal(decimal.RequireFromString("10.4506")), "Round-trip should preserve price")
	assert.True(t, parsed.Delta.Equal(decimal.RequireFromString("0.594823456789012345")), "Round-trip should preserve full precision delta")
}

func TestDecimalSerializationBondResult(t *testing.T) {
	result := models.BondResult{
		DirtyPrice:       decimal.RequireFromString("103.25"),
		CleanPrice:       decimal.RequireFromString("102.50"),
		Duration:         decimal.RequireFromString("4.5230"),
		ModifiedDuration: decimal.RequireFromString("4.3510"),
		Convexity:        decimal.RequireFromString("22.1500"),
		YieldToMaturity:  decimal.RequireFromString("0.0450"),
		AccruedInterest:  decimal.RequireFromString("0.7500"),
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"dirty_price":"103.25"`, "Bond dirty_price should be string")
	assert.Contains(t, jsonStr, `"duration":"4.523"`, "Bond duration should be string")
	assert.NotContains(t, jsonStr, `"dirty_price":103.25`, "Bond dirty_price should NOT be raw number")
}

func TestDecimalSerializationVaRResult(t *testing.T) {
	result := models.VaRResult{
		VaR:           decimal.RequireFromString("-25000.50"),
		CVaR:          decimal.RequireFromString("-35000.75"),
		Confidence:    decimal.RequireFromString("0.95"),
		HoldingPeriod: 10,
		Method:        "historical",
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"var":"-25000.5"`, "VaR should be string")
	assert.Contains(t, jsonStr, `"cvar":"-35000.75"`, "CVaR should be string")
	assert.Contains(t, jsonStr, `"confidence":"0.95"`, "Confidence should be string")
	assert.Contains(t, jsonStr, `"holding_period":10`, "HoldingPeriod (int) should remain number")
	assert.Contains(t, jsonStr, `"method":"historical"`, "Method (string) should remain string")
}

func TestPriceAmericanOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/options/american", r.URL.Path)

		var req models.AmericanOptionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.Spot.Equal(decimal.NewFromFloat(100)))
		assert.Equal(t, 200, req.Steps)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
				"price": "11.2345",
				"delta": "0.6123",
				"gamma": "0.0210",
				"theta": "-0.0145",
				"vega":  "0.3890",
				"rho":   "0.4890",
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

	req := models.AmericanOptionRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.25),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
		Steps:        200,
	}

	result, err := client.PriceAmericanOption(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(11.2345)))
	assert.True(t, result.Delta.Equal(decimal.NewFromFloat(0.6123)))
}

func TestPriceAmericanOptionValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.AmericanOptionRequest
		wantErr bool
	}{
		{
			name: "steps below minimum",
			req: models.AmericanOptionRequest{
				Spot: decimal.NewFromFloat(100), Strike: decimal.NewFromFloat(100),
				Rate: decimal.NewFromFloat(0.05), Volatility: decimal.NewFromFloat(0.2),
				TimeToExpiry: decimal.NewFromFloat(1), OptionType: models.OptionTypeCall,
				Steps: 5,
			},
			wantErr: true,
		},
		{
			name: "steps above maximum",
			req: models.AmericanOptionRequest{
				Spot: decimal.NewFromFloat(100), Strike: decimal.NewFromFloat(100),
				Rate: decimal.NewFromFloat(0.05), Volatility: decimal.NewFromFloat(0.2),
				TimeToExpiry: decimal.NewFromFloat(1), OptionType: models.OptionTypeCall,
				Steps: 1500,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateAmericanOptionRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateGreeks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/options/greeks", r.URL.Path)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
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

	req := models.GreeksRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.CalculateGreeks(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(10.4506)))
	assert.True(t, result.Delta.Equal(decimal.NewFromFloat(0.5948)))
}

func TestBuildYieldCurve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/yield-curve/build", r.URL.Path)

		var req models.YieldCurveRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "USD", req.Currency)
		assert.Len(t, req.Tenors, 8)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
				"currency":         "USD",
				"tenors":           []any{"1M", "3M", "6M", "1Y", "2Y", "5Y", "10Y", "30Y"},
				"rates":            []any{"0.043", "0.044", "0.045", "0.046", "0.047", "0.048", "0.049", "0.05"},
				"zero_rates":       []any{"0.0425", "0.0435", "0.0445", "0.0455", "0.0465", "0.0475", "0.0485", "0.0495"},
				"forward_rates":    []any{"0.0435", "0.0445", "0.0465", "0.0475", "0.0495", "0.0515", "0.0535", "0.0555"},
				"discount_factors": []any{"0.9965", "0.9891", "0.9780", "0.9558", "0.9137", "0.7896", "0.6139", "0.2237"},
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

	req := models.YieldCurveRequest{
		Currency: "USD",
		Tenors:   []string{"1M", "3M", "6M", "1Y", "2Y", "5Y", "10Y", "30Y"},
		Rates: []decimal.Decimal{
			decimal.NewFromFloat(0.043),
			decimal.NewFromFloat(0.044),
			decimal.NewFromFloat(0.045),
			decimal.NewFromFloat(0.046),
			decimal.NewFromFloat(0.047),
			decimal.NewFromFloat(0.048),
			decimal.NewFromFloat(0.049),
			decimal.NewFromFloat(0.05),
		},
	}

	result, err := client.BuildYieldCurve(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "USD", result.Currency)
	assert.Len(t, result.Tenors, 8)
	assert.Len(t, result.Rates, 8)
	assert.Len(t, result.ZeroRates, 8)
	assert.Len(t, result.ForwardRates, 8)
	assert.Len(t, result.DiscountFactors, 8)
}

func TestGetSupportedCurrencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/core/types/currencies", r.URL.Path)

		resp := map[string]any{
			"currencies": []any{"USD", "EUR", "GBP", "JPY", "CNY", "AUD", "CAD", "CHF"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		cache:      newCache(),
	}

	result, err := client.GetSupportedCurrencies()

	require.NoError(t, err)
	require.NotNil(t, result)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "Result should be a map")
	currencies, exists := resultMap["currencies"]
	require.True(t, exists, "Result should contain 'currencies' key")
	currencyList, ok := currencies.([]any)
	require.True(t, ok, "Currencies should be a slice")
	assert.Len(t, currencyList, 8)

	result2, err := client.GetSupportedCurrencies()
	require.NoError(t, err)
	assert.Equal(t, result, result2)
}
