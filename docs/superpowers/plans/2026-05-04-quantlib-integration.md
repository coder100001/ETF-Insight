# QuantLib Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect to FinceptTerminal's QuantLib cloud API (`api.fincept.in/quantlib/`) to add institutional-grade quantitative analysis: options pricing, Greeks, yield curves, bond pricing, and VaR calculation.

**Architecture:** Go backend acts as a proxy/middleware to the QuantLib cloud API. Frontend provides interactive analysis pages. No local QuantLib installation needed.

**Tech Stack:** Go (Gin HTTP client), React (Ant Design), TypeScript

---

## File Structure

| File | Responsibility |
|------|---------------|
| `backend/models/quantlib.go` | Request/response structs for QuantLib API |
| `backend/services/quantlib/quantlib_client.go` | HTTP client for QuantLib cloud API |
| `backend/handlers/quantlib_handler.go` | Gin handlers for QuantLib endpoints |
| `backend/router/router.go` | Route registration (modify existing) |
| `frontend/src/types/quantlib.ts` | TypeScript types for QuantLib |
| `frontend/src/services/api.ts` | Add quantlibAPI module (modify existing) |
| `frontend/src/pages/QuantLibAnalysis.tsx` | Interactive QuantLib analysis page |
| `frontend/src/App.tsx` | Add route (modify existing) |
| `backend/services/quantlib/quantlib_client_test.go` | Unit tests |

---

### Task 1: QuantLib Models

**Files:**
- Create: `backend/models/quantlib.go`

- [ ] **Step 1: Create QuantLib request/response models**

```go
package models

import "github.com/shopspring/decimal"

// OptionType represents call or put
type OptionType string

const (
	OptionTypeCall OptionType = "call"
	OptionTypePut  OptionType = "put"
)

// ExerciseStyle represents European or American
type ExerciseStyle string

const (
	ExerciseStyleEuropean ExerciseStyle = "european"
	ExerciseStyleAmerican ExerciseStyle = "american"
)

// EuropeanOptionRequest is the request for European option pricing
type EuropeanOptionRequest struct {
	Spot        decimal.Decimal `json:"spot" binding:"required"`
	Strike      decimal.Decimal `json:"strike" binding:"required"`
	Rate        decimal.Decimal `json:"rate" binding:"required"`
	Volatility  decimal.Decimal `json:"volatility" binding:"required"`
	TimeToExpiry decimal.Decimal `json:"time_to_expiry" binding:"required"`
	OptionType  OptionType      `json:"option_type" binding:"required"`
	DividendYield decimal.Decimal `json:"dividend_yield,omitempty"`
}

// AmericanOptionRequest is the request for American option pricing
type AmericanOptionRequest struct {
	Spot         decimal.Decimal `json:"spot" binding:"required"`
	Strike       decimal.Decimal `json:"strike" binding:"required"`
	Rate         decimal.Decimal `json:"rate" binding:"required"`
	Volatility   decimal.Decimal `json:"volatility" binding:"required"`
	TimeToExpiry decimal.Decimal `json:"time_to_expiry" binding:"required"`
	OptionType   OptionType      `json:"option_type" binding:"required"`
	Steps        int             `json:"steps,omitempty"`
	DividendYield decimal.Decimal `json:"dividend_yield,omitempty"`
}

// OptionResult is the response from option pricing
type OptionResult struct {
	Price   decimal.Decimal `json:"price"`
	Delta   decimal.Decimal `json:"delta"`
	Gamma   decimal.Decimal `json:"gamma"`
	Theta   decimal.Decimal `json:"theta"`
	Vega    decimal.Decimal `json:"vega"`
	Rho     decimal.Decimal `json:"rho"`
}

// GreeksRequest is the request for Greeks calculation
type GreeksRequest struct {
	Spot         decimal.Decimal `json:"spot" binding:"required"`
	Strike       decimal.Decimal `json:"strike" binding:"required"`
	Rate         decimal.Decimal `json:"rate" binding:"required"`
	Volatility   decimal.Decimal `json:"volatility" binding:"required"`
	TimeToExpiry decimal.Decimal `json:"time_to_expiry" binding:"required"`
	OptionType   OptionType      `json:"option_type" binding:"required"`
}

// YieldCurveRequest is the request for yield curve construction
type YieldCurveRequest struct {
	Currency    string              `json:"currency" binding:"required"`
	Calendar    string              `json:"calendar,omitempty"`
	DayCount    string              `json:"day_count,omitempty"`
	Tenors      []string            `json:"tenors" binding:"required"`
	Rates       []decimal.Decimal   `json:"rates" binding:"required"`
	Compounding string              `json:"compounding,omitempty"`
	Frequency   string              `json:"frequency,omitempty"`
}

// YieldCurveResult is the response from yield curve construction
type YieldCurveResult struct {
	Currency    string                   `json:"currency"`
	Tenors      []string                 `json:"tenors"`
	Rates       []decimal.Decimal        `json:"rates"`
	ZeroRates   []decimal.Decimal        `json:"zero_rates"`
	ForwardRates []decimal.Decimal       `json:"forward_rates"`
	DiscountFactors []decimal.Decimal    `json:"discount_factors"`
}

// BondRequest is the request for bond pricing
type BondRequest struct {
	FaceValue     decimal.Decimal `json:"face_value" binding:"required"`
	CouponRate    decimal.Decimal `json:"coupon_rate" binding:"required"`
	Frequency     int             `json:"frequency" binding:"required"`
	Maturity      string          `json:"maturity" binding:"required"`
	YieldToMaturity decimal.Decimal `json:"yield_to_maturity" binding:"required"`
	SettlementDate string         `json:"settlement_date,omitempty"`
	DayCount      string          `json:"day_count,omitempty"`
}

// BondResult is the response from bond pricing
type BondResult struct {
	DirtyPrice  decimal.Decimal `json:"dirty_price"`
	CleanPrice  decimal.Decimal `json:"clean_price"`
	Duration    decimal.Decimal `json:"duration"`
	ModifiedDuration decimal.Decimal `json:"modified_duration"`
	Convexity   decimal.Decimal `json:"convexity"`
	YieldToMaturity decimal.Decimal `json:"yield_to_maturity"`
	AccruedInterest decimal.Decimal `json:"accrued_interest"`
}

// VaRRequest is the request for Value at Risk calculation
type VaRRequest struct {
	PortfolioValue decimal.Decimal   `json:"portfolio_value" binding:"required"`
	Returns        []decimal.Decimal `json:"returns" binding:"required"`
	Confidence     decimal.Decimal   `json:"confidence" binding:"required"`
	HoldingPeriod int              `json:"holding_period,omitempty"`
	Method         string           `json:"method,omitempty"`
}

// VaRResult is the response from VaR calculation
type VaRResult struct {
	VaR          decimal.Decimal `json:"var"`
	CVaR         decimal.Decimal `json:"cvar"`
	Confidence   decimal.Decimal `json:"confidence"`
	HoldingPeriod int            `json:"holding_period"`
	Method       string          `json:"method"`
}

// QuantLibAPIResponse is the generic response wrapper from QuantLib API
type QuantLibAPIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go build ./models/...`
Expected: No errors

---

### Task 2: QuantLib HTTP Client

**Files:**
- Create: `backend/services/quantlib/quantlib_client.go`

- [ ] **Step 1: Create QuantLib client with HTTP methods**

```go
package quantlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

const (
	defaultBaseURL = "https://api.fincept.in/quantlib"
	defaultTimeout = 30 * time.Second
)

// Client communicates with the QuantLib cloud API
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewClient creates a new QuantLib API client
func NewClient() *Client {
	baseURL := os.Getenv("QUANTLIB_API_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	apiKey := os.Getenv("QUANTLIB_API_KEY")

	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		apiKey:     apiKey,
	}
}

// doRequest performs an HTTP request and decodes the response
func (c *Client) doRequest(method, endpoint string, body interface{}, result interface{}) error {
	url := c.baseURL + "/" + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ETF-Insight/1.0")

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// Try to parse as QuantLibAPIResponse wrapper
	var apiResp models.QuantLibAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		// Try direct parse if wrapper fails
		return json.Unmarshal(respBody, result)
	}

	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Message)
	}

	// Marshal data back to JSON then unmarshal into result
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal API data: %w", err)
	}

	return json.Unmarshal(dataBytes, result)
}

// PriceEuropeanOption prices a European option
func (c *Client) PriceEuropeanOption(req models.EuropeanOptionRequest) (*models.OptionResult, error) {
	var result models.OptionResult
	if err := c.doRequest("POST", "options/european", req, &result); err != nil {
		return nil, fmt.Errorf("failed to price European option: %w", err)
	}
	return &result, nil
}

// PriceAmericanOption prices an American option
func (c *Client) PriceAmericanOption(req models.AmericanOptionRequest) (*models.OptionResult, error) {
	var result models.OptionResult
	if err := c.doRequest("POST", "options/american", req, &result); err != nil {
		return nil, fmt.Errorf("failed to price American option: %w", err)
	}
	return &result, nil
}

// CalculateGreeks calculates option Greeks
func (c *Client) CalculateGreeks(req models.GreeksRequest) (*models.OptionResult, error) {
	var result models.OptionResult
	if err := c.doRequest("POST", "options/greeks", req, &result); err != nil {
		return nil, fmt.Errorf("failed to calculate Greeks: %w", err)
	}
	return &result, nil
}

// BuildYieldCurve constructs a yield curve
func (c *Client) BuildYieldCurve(req models.YieldCurveRequest) (*models.YieldCurveResult, error) {
	var result models.YieldCurveResult
	if err := c.doRequest("POST", "yield-curve/build", req, &result); err != nil {
		return nil, fmt.Errorf("failed to build yield curve: %w", err)
	}
	return &result, nil
}

// PriceBond prices a fixed-income bond
func (c *Client) PriceBond(req models.BondRequest) (*models.BondResult, error) {
	var result models.BondResult
	if err := c.doRequest("POST", "bonds/fixed", req, &result); err != nil {
		return nil, fmt.Errorf("failed to price bond: %w", err)
	}
	return &result, nil
}

// CalculateVaR calculates Value at Risk
func (c *Client) CalculateVaR(req models.VaRRequest) (*models.VaRResult, error) {
	// Apply defaults
	if req.Confidence.IsZero() {
		req.Confidence = decimal.NewFromFloat(0.95)
	}
	if req.HoldingPeriod == 0 {
		req.HoldingPeriod = 1
	}
	if req.Method == "" {
		req.Method = "historical"
	}

	var result models.VaRResult
	if err := c.doRequest("POST", "risk/var", req, &result); err != nil {
		return nil, fmt.Errorf("failed to calculate VaR: %w", err)
	}
	return &result, nil
}

// GetSupportedCurrencies returns supported currencies
func (c *Client) GetSupportedCurrencies() (interface{}, error) {
	var result interface{}
	if err := c.doRequest("GET", "core/types/currencies", nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get currencies: %w", err)
	}
	return result, nil
}

// GetFrequencies returns payment frequencies
func (c *Client) GetFrequencies() (interface{}, error) {
	var result interface{}
	if err := c.doRequest("GET", "core/types/frequencies", nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get frequencies: %w", err)
	}
	return result, nil
}

// GetCalendars returns available calendars
func (c *Client) GetCalendars() (interface{}, error) {
	var result interface{}
	if err := c.doRequest("GET", "scheduling/calendar/list", nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get calendars: %w", err)
	}
	return result, nil
}

// GetDayCountConventions returns day count conventions
func (c *Client) GetDayCountConventions() (interface{}, error) {
	var result interface{}
	if err := c.doRequest("GET", "scheduling/daycount/conventions", nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get day count conventions: %w", err)
	}
	return result, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go build ./services/quantlib/...`
Expected: No errors

---

### Task 3: QuantLib Handler

**Files:**
- Create: `backend/handlers/quantlib_handler.go`

- [ ] **Step 1: Create QuantLib handler**

```go
package handlers

import (
	"net/http"

	"etf-insight/models"
	"etf-insight/services/quantlib"

	"github.com/gin-gonic/gin"
)

// QuantLibHandler handles QuantLib API requests
type QuantLibHandler struct {
	client *quantlib.Client
}

// NewQuantLibHandler creates a new QuantLib handler
func NewQuantLibHandler() *QuantLibHandler {
	return &QuantLibHandler{
		client: quantlib.NewClient(),
	}
}

// PriceEuropeanOption handles European option pricing
func (h *QuantLibHandler) PriceEuropeanOption(c *gin.Context) {
	var req models.EuropeanOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.PriceEuropeanOption(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// PriceAmericanOption handles American option pricing
func (h *QuantLibHandler) PriceAmericanOption(c *gin.Context) {
	var req models.AmericanOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.PriceAmericanOption(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// CalculateGreeks handles Greeks calculation
func (h *QuantLibHandler) CalculateGreeks(c *gin.Context) {
	var req models.GreeksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateGreeks(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// BuildYieldCurve handles yield curve construction
func (h *QuantLibHandler) BuildYieldCurve(c *gin.Context) {
	var req models.YieldCurveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.BuildYieldCurve(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// PriceBond handles bond pricing
func (h *QuantLibHandler) PriceBond(c *gin.Context) {
	var req models.BondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.PriceBond(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// CalculateVaR handles VaR calculation
func (h *QuantLibHandler) CalculateVaR(c *gin.Context) {
	var req models.VaRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateVaR(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetReferenceData returns QuantLib reference data
func (h *QuantLibHandler) GetReferenceData(c *gin.Context) {
	dataType := c.Param("type")

	var result interface{}
	var err error

	switch dataType {
	case "currencies":
		result, err = h.client.GetSupportedCurrencies()
	case "frequencies":
		result, err = h.client.GetFrequencies()
	case "calendars":
		result, err = h.client.GetCalendars()
	case "daycount":
		result, err = h.client.GetDayCountConventions()
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "unknown reference data type: " + dataType})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go build ./handlers/...`
Expected: No errors

---

### Task 4: Router Registration

**Files:**
- Modify: `backend/router/router.go`

- [ ] **Step 1: Add QuantLib handler to Handlers struct**

Add to the `Handlers` struct (after line 35):
```go
QuantLib        *handlers.QuantLibHandler
```

- [ ] **Step 2: Initialize QuantLib handler in NewRouter**

Add after the `h.BlackLitterman` assignment (after line 86):
```go
h.QuantLib = handlers.NewQuantLibHandler()
```

- [ ] **Step 3: Add route registration call**

Add to `RegisterRoutes()` (after line 111):
```go
r.registerQuantLibRoutes()
```

- [ ] **Step 4: Add route registration method**

Add at the end of the file:
```go
func (r *Router) registerQuantLibRoutes() {
	ql := r.engine.Group("/api/quantlib")
	{
		ql.POST("/options/european", r.handlers.QuantLib.PriceEuropeanOption)
		ql.POST("/options/american", r.handlers.QuantLib.PriceAmericanOption)
		ql.POST("/options/greeks", r.handlers.QuantLib.CalculateGreeks)
		ql.POST("/yield-curve/build", r.handlers.QuantLib.BuildYieldCurve)
		ql.POST("/bonds/price", r.handlers.QuantLib.PriceBond)
		ql.POST("/risk/var", r.handlers.QuantLib.CalculateVaR)
		ql.GET("/reference/:type", r.handlers.QuantLib.GetReferenceData)
	}
}
```

- [ ] **Step 5: Verify compilation**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go build ./...`
Expected: No errors

---

### Task 5: Frontend TypeScript Types

**Files:**
- Create: `frontend/src/types/quantlib.ts`

- [ ] **Step 1: Create QuantLib TypeScript types**

```typescript
// QuantLib API Types

export interface EuropeanOptionRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
  dividend_yield?: number;
}

export interface AmericanOptionRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
  steps?: number;
  dividend_yield?: number;
}

export interface GreeksRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
}

export interface OptionResult {
  price: number;
  delta: number;
  gamma: number;
  theta: number;
  vega: number;
  rho: number;
}

export interface YieldCurveRequest {
  currency: string;
  calendar?: string;
  day_count?: string;
  tenors: string[];
  rates: number[];
  compounding?: string;
  frequency?: string;
}

export interface YieldCurveResult {
  currency: string;
  tenors: string[];
  rates: number[];
  zero_rates: number[];
  forward_rates: number[];
  discount_factors: number[];
}

export interface BondRequest {
  face_value: number;
  coupon_rate: number;
  frequency: number;
  maturity: string;
  yield_to_maturity: number;
  settlement_date?: string;
  day_count?: string;
}

export interface BondResult {
  dirty_price: number;
  clean_price: number;
  duration: number;
  modified_duration: number;
  convexity: number;
  yield_to_maturity: number;
  accrued_interest: number;
}

export interface VaRRequest {
  portfolio_value: number;
  returns: number[];
  confidence: number;
  holding_period?: number;
  method?: 'historical' | 'parametric' | 'monte_carlo';
}

export interface VaRResult {
  var: number;
  cvar: number;
  confidence: number;
  holding_period: number;
  method: string;
}
```

- [ ] **Step 2: Add to types barrel export**

Add to `frontend/src/types/index.ts`:
```typescript
export * from './quantlib';
```

---

### Task 6: Frontend API Service

**Files:**
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: Add QuantLib API module**

Add at the end of the file (before any closing exports):
```typescript
import type {
  EuropeanOptionRequest, AmericanOptionRequest, GreeksRequest,
  OptionResult, YieldCurveRequest, YieldCurveResult,
  BondRequest, BondResult, VaRRequest, VaRResult
} from '../types/quantlib';

export const quantlibAPI = {
  priceEuropeanOption: async (req: EuropeanOptionRequest): Promise<ApiResponse<OptionResult>> => {
    return request<ApiResponse<OptionResult>>('/quantlib/options/european', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  priceAmericanOption: async (req: AmericanOptionRequest): Promise<ApiResponse<OptionResult>> => {
    return request<ApiResponse<OptionResult>>('/quantlib/options/american', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  calculateGreeks: async (req: GreeksRequest): Promise<ApiResponse<OptionResult>> => {
    return request<ApiResponse<OptionResult>>('/quantlib/options/greeks', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  buildYieldCurve: async (req: YieldCurveRequest): Promise<ApiResponse<YieldCurveResult>> => {
    return request<ApiResponse<YieldCurveResult>>('/quantlib/yield-curve/build', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  priceBond: async (req: BondRequest): Promise<ApiResponse<BondResult>> => {
    return request<ApiResponse<BondResult>>('/quantlib/bonds/price', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  calculateVaR: async (req: VaRRequest): Promise<ApiResponse<VaRResult>> => {
    return request<ApiResponse<VaRResult>>('/quantlib/risk/var', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  getReferenceData: async (type: string): Promise<ApiResponse<unknown>> => {
    return request<ApiResponse<unknown>>(`/quantlib/reference/${type}`);
  },
};
```

---

### Task 7: Frontend QuantLib Page

**Files:**
- Create: `frontend/src/pages/QuantLibAnalysis.tsx`

- [ ] **Step 1: Create QuantLib Analysis page**

```tsx
import React, { useState } from 'react';
import {
  Card, Tabs, Form, InputNumber, Select, Button, Row, Col,
  Statistic, Table, message, Spin, Input
} from 'antd';
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend
} from 'recharts';
import { quantlibAPI } from '../services/api';
import type { OptionResult, BondResult, YieldCurveResult, VaRResult } from '../types/quantlib';

const { TabPane } = Tabs;
const { Option } = Select;

const QuantLibAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [optionResult, setOptionResult] = useState<OptionResult | null>(null);
  const [bondResult, setBondResult] = useState<BondResult | null>(null);
  const [yieldCurveData, setYieldCurveData] = useState<YieldCurveResult | null>(null);
  const [varResult, setVarResult] = useState<VaRResult | null>(null);

  const handleOptionPrice = async (values: Record<string, number>) => {
    setLoading(true);
    try {
      const result = await quantlibAPI.priceEuropeanOption({
        spot: values.spot,
        strike: values.strike,
        rate: values.rate / 100,
        volatility: values.volatility / 100,
        time_to_expiry: values.time_to_expiry,
        option_type: values.option_type as 'call' | 'put',
      });
      if (result.success && result.data) {
        setOptionResult(result.data);
        message.success('期权定价完成');
      } else {
        message.error(result.error || '定价失败');
      }
    } catch (error) {
      message.error('请求失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleBondPrice = async (values: Record<string, number>) => {
    setLoading(true);
    try {
      const result = await quantlibAPI.priceBond({
        face_value: values.face_value,
        coupon_rate: values.coupon_rate / 100,
        frequency: values.frequency,
        maturity: `${values.maturity_years}Y`,
        yield_to_maturity: values.ytm / 100,
      });
      if (result.success && result.data) {
        setBondResult(result.data);
        message.success('债券定价完成');
      } else {
        message.error(result.error || '定价失败');
      }
    } catch (error) {
      message.error('请求失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleYieldCurve = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const tenorsStr = (values.tenors as string) || '1M,3M,6M,1Y,2Y,5Y,10Y,30Y';
      const ratesStr = (values.rates as string) || '4.5,4.6,4.7,4.8,4.9,5.0,5.1,5.2';
      const tenors = tenorsStr.split(',').map(t => t.trim());
      const rates = ratesStr.split(',').map(r => parseFloat(r.trim()) / 100);

      const result = await quantlibAPI.buildYieldCurve({
        currency: (values.currency as string) || 'USD',
        tenors,
        rates,
      });
      if (result.success && result.data) {
        setYieldCurveData(result.data);
        message.success('收益率曲线构建完成');
      } else {
        message.error(result.error || '构建失败');
      }
    } catch (error) {
      message.error('请求失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleVaR = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const returnsStr = (values.returns as string) || '';
      const returns = returnsStr.split(',').map(r => parseFloat(r.trim()) / 100);

      const result = await quantlibAPI.calculateVaR({
        portfolio_value: values.portfolio_value as number,
        returns,
        confidence: (values.confidence as number) / 100,
        holding_period: values.holding_period as number,
        method: values.method as 'historical' | 'parametric',
      });
      if (result.success && result.data) {
        setVarResult(result.data);
        message.success('VaR 计算完成');
      } else {
        message.error(result.error || '计算失败');
      }
    } catch (error) {
      message.error('请求失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const optionColumns = [
    { title: '指标', dataIndex: 'label', key: 'label' },
    { title: '值', dataIndex: 'value', key: 'value' },
  ];

  const optionData = optionResult
    ? [
        { key: '1', label: '期权价格', value: optionResult.price.toFixed(4) },
        { key: '2', label: 'Delta', value: optionResult.delta.toFixed(4) },
        { key: '3', label: 'Gamma', value: optionResult.gamma.toFixed(4) },
        { key: '4', label: 'Theta', value: optionResult.theta.toFixed(4) },
        { key: '5', label: 'Vega', value: optionResult.vega.toFixed(4) },
        { key: '6', label: 'Rho', value: optionResult.rho.toFixed(4) },
      ]
    : [];

  return (
    <div style={{ padding: 24 }}>
      <h1>QuantLib 量化分析</h1>
      <p style={{ color: '#666', marginBottom: 24 }}>
        基于 QuantLib 引擎的机构级量化分析工具
      </p>

      <Spin spinning={loading}>
        <Tabs defaultActiveKey="options">
          <TabPane tab="期权定价" key="options">
            <Row gutter={24}>
              <Col span={12}>
                <Card title="欧式期权参数">
                  <Form onFinish={handleOptionPrice} layout="vertical">
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item label="标的价格" name="spot" initialValue={100}>
                          <InputNumber style={{ width: '100%' }} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item label="行权价" name="strike" initialValue={105}>
                          <InputNumber style={{ width: '100%' }} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item label="无风险利率 (%)" name="rate" initialValue={5}>
                          <InputNumber style={{ width: '100%' }} step={0.1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item label="波动率 (%)" name="volatility" initialValue={20}>
                          <InputNumber style={{ width: '100%' }} step={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item label="到期时间 (年)" name="time_to_expiry" initialValue={1}>
                          <InputNumber style={{ width: '100%' }} step={0.1} min={0.01} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Form.Item label="期权类型" name="option_type" initialValue="call">
                      <Select>
                        <Option value="call">看涨 (Call)</Option>
                        <Option value="put">看跌 (Put)</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算价格
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="定价结果">
                  {optionResult ? (
                    <>
                      <Row gutter={16} style={{ marginBottom: 24 }}>
                        <Col span={8}>
                          <Statistic title="期权价格" value={optionResult.price} precision={4} />
                        </Col>
                        <Col span={8}>
                          <Statistic title="Delta" value={optionResult.delta} precision={4} />
                        </Col>
                        <Col span={8}>
                          <Statistic title="Gamma" value={optionResult.gamma} precision={4} />
                        </Col>
                      </Row>
                      <Table
                        columns={optionColumns}
                        dataSource={optionData}
                        pagination={false}
                        size="small"
                      />
                    </>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                      请在左侧输入参数并点击计算
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="债券定价" key="bonds">
            <Row gutter={24}>
              <Col span={12}>
                <Card title="债券参数">
                  <Form onFinish={handleBondPrice} layout="vertical">
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item label="面值" name="face_value" initialValue={1000}>
                          <InputNumber style={{ width: '100%' }} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item label="票息率 (%)" name="coupon_rate" initialValue={5}>
                          <InputNumber style={{ width: '100%' }} step={0.1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item label="付息频率" name="frequency" initialValue={2}>
                          <Select>
                            <Option value={1}>年付</Option>
                            <Option value={2}>半年付</Option>
                            <Option value={4}>季付</Option>
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item label="期限 (年)" name="maturity_years" initialValue={10}>
                          <InputNumber style={{ width: '100%' }} min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item label="到期收益率 (%)" name="ytm" initialValue={5}>
                          <InputNumber style={{ width: '100%' }} step={0.1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算价格
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="定价结果">
                  {bondResult ? (
                    <Row gutter={[16, 16]}>
                      <Col span={8}>
                        <Statistic title="净价" value={bondResult.clean_price} precision={4} />
                      </Col>
                      <Col span={8}>
                        <Statistic title="全价" value={bondResult.dirty_price} precision={4} />
                      </Col>
                      <Col span={8}>
                        <Statistic title="久期" value={bondResult.duration} precision={4} suffix="年" />
                      </Col>
                      <Col span={8}>
                        <Statistic title="修正久期" value={bondResult.modified_duration} precision={4} />
                      </Col>
                      <Col span={8}>
                        <Statistic title="凸性" value={bondResult.convexity} precision={4} />
                      </Col>
                      <Col span={8}>
                        <Statistic title="应计利息" value={bondResult.accrued_interest} precision={4} />
                      </Col>
                    </Row>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                      请在左侧输入参数并点击计算
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="收益率曲线" key="yield-curve">
            <Row gutter={24}>
              <Col span={8}>
                <Card title="曲线参数">
                  <Form onFinish={handleYieldCurve} layout="vertical">
                    <Form.Item label="货币" name="currency" initialValue="USD">
                      <Select>
                        <Option value="USD">USD</Option>
                        <Option value="EUR">EUR</Option>
                        <Option value="CNY">CNY</Option>
                        <Option value="GBP">GBP</Option>
                        <Option value="JPY">JPY</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item label="期限" name="tenors" initialValue="1M,3M,6M,1Y,2Y,5Y,10Y,30Y">
                      <Input.TextArea rows={2} placeholder="1M,3M,6M,1Y,2Y,5Y,10Y,30Y" />
                    </Form.Item>
                    <Form.Item label="利率 (%)" name="rates" initialValue="4.5,4.6,4.7,4.8,4.9,5.0,5.1,5.2">
                      <Input.TextArea rows={2} placeholder="4.5,4.6,4.7,4.8,4.9,5.0,5.1,5.2" />
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        构建曲线
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={16}>
                <Card title="收益率曲线">
                  {yieldCurveData ? (
                    <ResponsiveContainer width="100%" height={400}>
                      <LineChart
                        data={yieldCurveData.tenors.map((t, i) => ({
                          tenor: t,
                          rate: yieldCurveData.rates[i] * 100,
                          zeroRate: yieldCurveData.zero_rates[i] * 100,
                        }))}
                        margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="tenor" />
                        <YAxis label={{ value: '利率 (%)', angle: -90, position: 'insideLeft' }} />
                        <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} />
                        <Legend />
                        <Line type="monotone" dataKey="rate" stroke="#1890ff" name="即期利率" />
                        <Line type="monotone" dataKey="zeroRate" stroke="#52c41a" name="零息利率" />
                      </LineChart>
                    </ResponsiveContainer>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                      请在左侧输入参数并点击构建
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="VaR 计算" key="var">
            <Row gutter={24}>
              <Col span={12}>
                <Card title="VaR 参数">
                  <Form onFinish={handleVaR} layout="vertical">
                    <Form.Item label="组合价值" name="portfolio_value" initialValue={1000000}>
                      <InputNumber style={{ width: '100%' }} min={0} />
                    </Form.Item>
                    <Form.Item label="历史收益率 (%)" name="returns">
                      <Input.TextArea
                        rows={3}
                        placeholder="输入历史收益率，用逗号分隔，如: -2.1,1.5,-0.8,3.2,-1.5,0.9,-3.1,2.8"
                      />
                    </Form.Item>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item label="置信水平 (%)" name="confidence" initialValue={95}>
                          <Select>
                            <Option value={90}>90%</Option>
                            <Option value={95}>95%</Option>
                            <Option value={99}>99%</Option>
                          </Select>
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item label="持有期 (天)" name="holding_period" initialValue={1}>
                          <InputNumber style={{ width: '100%' }} min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item label="方法" name="method" initialValue="historical">
                          <Select>
                            <Option value="historical">历史模拟法</Option>
                            <Option value="parametric">参数法</Option>
                          </Select>
                        </Form.Item>
                      </Col>
                    </Row>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算 VaR
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="VaR 结果">
                  {varResult ? (
                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Statistic
                          title="VaR (风险价值)"
                          value={varResult.var}
                          precision={2}
                          prefix="$"
                          valueStyle={{ color: '#cf1322' }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="CVaR (条件风险价值)"
                          value={varResult.cvar}
                          precision={2}
                          prefix="$"
                          valueStyle={{ color: '#cf1322' }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic title="置信水平" value={varResult.confidence * 100} suffix="%" />
                      </Col>
                      <Col span={12}>
                        <Statistic title="持有期" value={varResult.holding_period} suffix="天" />
                      </Col>
                      <Col span={24}>
                        <Statistic title="计算方法" value={varResult.method} />
                      </Col>
                    </Row>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
                      请在左侧输入参数并点击计算
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>
        </Tabs>
      </Spin>
    </div>
  );
};

export default QuantLibAnalysis;
```

---

### Task 8: Frontend Route Registration

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Add QuantLib route**

Add import at the top:
```tsx
import QuantLibAnalysis from './pages/QuantLibAnalysis';
```

Add route in the `<Routes>` section (after existing routes):
```tsx
<Route path="/quantlib" element={<QuantLibAnalysis />} />
```

- [ ] **Step 2: Add navigation menu item**

Add to the menu items array (wherever the sidebar/navigation is defined):
```tsx
{
  key: 'quantlib',
  icon: <CalculatorOutlined />,
  label: <Link to="/quantlib">QuantLib 分析</Link>,
}
```

Import the icon:
```tsx
import { CalculatorOutlined } from '@ant-design/icons';
```

---

### Task 9: Unit Tests

**Files:**
- Create: `backend/services/quantlib/quantlib_client_test.go`

- [ ] **Step 1: Write unit tests for QuantLib client**

```go
package quantlib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceEuropeanOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/options/european")

		var req models.EuropeanOptionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.True(t, req.Spot.Equal(decimal.NewFromInt(100)))
		assert.True(t, req.Strike.Equal(decimal.NewFromInt(105)))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"price": 8.02,
				"delta": 0.54,
				"gamma": 0.019,
				"theta": -0.015,
				"vega":  0.38,
				"rho":   0.46,
			},
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	result, err := client.PriceEuropeanOption(models.EuropeanOptionRequest{
		Spot:         decimal.NewFromInt(100),
		Strike:       decimal.NewFromInt(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.20),
		TimeToExpiry: decimal.NewFromInt(1),
		OptionType:   models.OptionTypeCall,
	})

	require.NoError(t, err)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(8.02)))
	assert.True(t, result.Delta.Equal(decimal.NewFromFloat(0.54)))
}

func TestPriceEuropeanOptionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "message": "internal error"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	_, err := client.PriceEuropeanOption(models.EuropeanOptionRequest{
		Spot:         decimal.NewFromInt(100),
		Strike:       decimal.NewFromInt(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.20),
		TimeToExpiry: decimal.NewFromInt(1),
		OptionType:   models.OptionTypeCall,
	})

	assert.Error(t, err)
}

func TestPriceBond(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/bonds/fixed")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"dirty_price":       100.5,
				"clean_price":       100.0,
				"duration":          7.5,
				"modified_duration": 7.2,
				"convexity":         65.3,
				"yield_to_maturity": 0.05,
				"accrued_interest":  0.5,
			},
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	result, err := client.PriceBond(models.BondRequest{
		FaceValue:       decimal.NewFromInt(1000),
		CouponRate:      decimal.NewFromFloat(0.05),
		Frequency:       2,
		Maturity:        "10Y",
		YieldToMaturity: decimal.NewFromFloat(0.05),
	})

	require.NoError(t, err)
	assert.True(t, result.CleanPrice.Equal(decimal.NewFromFloat(100.0)))
	assert.True(t, result.Duration.Equal(decimal.NewFromFloat(7.5)))
}

func TestCalculateVaR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/risk/var")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"var":            -25000,
				"cvar":           -35000,
				"confidence":     0.95,
				"holding_period": 1,
				"method":         "historical",
			},
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	result, err := client.CalculateVaR(models.VaRRequest{
		PortfolioValue: decimal.NewFromInt(1000000),
		Returns: []decimal.Decimal{
			decimal.NewFromFloat(-0.021),
			decimal.NewFromFloat(0.015),
			decimal.NewFromFloat(-0.008),
		},
		Confidence: decimal.NewFromFloat(0.95),
	})

	require.NoError(t, err)
	assert.True(t, result.VaR.Equal(decimal.NewFromInt(-25000)))
	assert.True(t, result.CVaR.Equal(decimal.NewFromInt(-35000)))
}

func TestNewClientDefaultURL(t *testing.T) {
	client := NewClient()
	assert.Equal(t, defaultBaseURL, client.baseURL)
}
```

- [ ] **Step 2: Run tests**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go test ./services/quantlib/... -v`
Expected: All tests PASS

---

### Task 10: Integration Verification

- [ ] **Step 1: Build entire backend**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go build ./...`
Expected: No errors

- [ ] **Step 2: Run all backend tests**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/backend && go test ./... -count=1`
Expected: All tests pass

- [ ] **Step 3: Build frontend**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/frontend && npm run build`
Expected: Build succeeds with no TypeScript errors

- [ ] **Step 4: Run frontend lint**

Run: `cd /Users/liunian/Desktop/dnmp/py_project/frontend && npm run lint`
Expected: No errors

- [ ] **Step 5: Commit all changes**

```bash
cd /Users/liunian/Desktop/dnmp/py_project
git add backend/models/quantlib.go
git add backend/services/quantlib/
git add backend/handlers/quantlib_handler.go
git add backend/router/router.go
git add frontend/src/types/quantlib.ts
git add frontend/src/services/api.ts
git add frontend/src/pages/QuantLibAnalysis.tsx
git add frontend/src/App.tsx
git commit -m "feat(quantlib): integrate QuantLib cloud API for options, bonds, yield curves, VaR

- Add Go HTTP client for api.fincept.in/quantlib/ cloud service
- Add models for European/American options, bonds, yield curves, VaR
- Add Gin handlers with /api/quantlib/* endpoints
- Add frontend QuantLib Analysis page with interactive forms and charts
- Add unit tests with mock HTTP server
- Phase 1 of FinceptTerminal integration plan"
```
