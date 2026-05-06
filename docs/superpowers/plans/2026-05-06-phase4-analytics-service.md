# Phase 4: Analytics Microservice - TDD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract and integrate Portfolio Management analytics modules from FinceptTerminal using Test-Driven Development (TDD).

**Architecture:** Extract Python analytics modules from FinceptTerminal (AGPL, non-commercial use), wrap them in a FastAPI service. Go backend acts as HTTP proxy. Follow TDD: write failing tests first, then implement minimal code to pass.

**Tech Stack:** Python 3.9+, FastAPI, uvicorn, pandas, numpy, scipy. Go: net/http, gin.

**License:** AGPL-3.0 (direct code reuse, non-commercial). A `NOTICE` file must be added to `backend/services/analytics/`.

---

## TDD Workflow

```
RED → GREEN → REFACTOR
  │       │         │
  │       │         └── Clean up code, keep tests green
  │       └──────────── Write minimal code to pass test
  └──────────────────── Write failing test first
```

### Iron Rule
**NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST**

---

## Module Inventory

### Portfolio Management Modules (11 files)

| Module | Key Classes | Priority | Status |
|--------|-------------|----------|--------|
| `portfolio_optimization.py` | `optimize_portfolio()`, `fetch_returns()` | P0 | 📋 待提取 |
| `risk_management.py` | `RiskBudget`, `VaRCalculations`, `ScenarioAnalysis` | P0 | 📋 待提取 |
| `portfolio_analytics.py` | `AssetClassAnalysis`, `CAPMAnalysis`, `EfficientFrontierAnalysis` | P0 | 📋 待提取 |
| `portfolio_management.py` | Core portfolio operations | P1 | 📋 待提取 |
| `portfolio_planning.py` | Planning and allocation | P1 | 📋 待提取 |
| `active_management.py` | Active management strategies | P1 | 📋 待提取 |
| `economics_markets.py` | Economic analysis | P2 | 📋 待提取 |
| `ffn_analysis.py` | Financial function analysis | P2 | 📋 待提取 |
| `math_engine.py` | Mathematical operations | P2 | 📋 待提取 |
| `behavioral_finance.py` | Behavioral analysis | P2 | 📋 待提取 |
| `etf_analytics.py` | ETF-specific analytics | P2 | 📋 待提取 |

### Supporting Files

| File | Purpose | Status |
|------|---------|--------|
| `config.py` | Configuration | 📋 待提取 |
| `data_manager.py` | Data management | 📋 待提取 |
| `fetch_historical.py` | Historical data | 📋 待提取 |
| `fetch_quotes.py` | Quote data | 📋 待提取 |
| `quantstats_analysis.py` | QuantStats integration | 📋 待提取 |

---

## File Structure

```
backend/services/analytics/
├── analytics_server.py              # FastAPI entry point (port 8093)
├── requirements.txt                 # Python dependencies
├── Dockerfile                       # Container
├── NOTICE                           # AGPL-3.0 compliance notice
├── modules/                         # Extracted analytics modules
│   ├── __init__.py
│   ├── portfolio_optimization.py    # From FinceptTerminal
│   ├── risk_management.py           # From FinceptTerminal
│   ├── portfolio_analytics.py       # From FinceptTerminal
│   ├── portfolio_management.py      # From FinceptTerminal
│   ├── portfolio_planning.py        # From FinceptTerminal
│   ├── active_management.py         # From FinceptTerminal
│   ├── economics_markets.py         # From FinceptTerminal
│   ├── ffn_analysis.py              # From FinceptTerminal
│   ├── math_engine.py               # From FinceptTerminal
│   ├── behavioral_finance.py        # From FinceptTerminal
│   ├── etf_analytics.py             # From FinceptTerminal
│   ├── config.py                    # From FinceptTerminal
│   ├── data_manager.py              # From FinceptTerminal
│   ├── fetch_historical.py          # From FinceptTerminal
│   ├── fetch_quotes.py              # From FinceptTerminal
│   └── quantstats_analysis.py       # From FinceptTerminal
├── routers/                         # FastAPI route modules
│   ├── __init__.py
│   ├── optimization.py
│   ├── risk.py
│   ├── analytics.py
│   └── planning.py
├── tests/
│   ├── __init__.py
│   ├── test_health.py
│   ├── test_optimization.py
│   ├── test_risk.py
│   └── test_analytics.py
backend/services/analytics/analytics_client.go    # Go HTTP client
backend/handlers/analytics_handler.go             # Go Gin handler
backend/router/router.go                          # MODIFY: add analytics routes
```

---

## TDD Implementation Tasks

### Task 1: Project Scaffolding + Requirements + AGPL Notice

**Files:**
- Create: `backend/services/analytics/requirements.txt`
- Create: `backend/services/analytics/modules/__init__.py`
- Create: `backend/services/analytics/routers/__init__.py`
- Create: `backend/services/analytics/tests/__init__.py`
- Create: `backend/services/analytics/NOTICE`

- [ ] **Step 1: Create requirements.txt**

```txt
fastapi==0.115.6
uvicorn[standard]==0.34.0
pandas==2.0.3
numpy==1.26.4
scipy==1.11.4
python-dotenv==1.0.1
pytest==8.3.4
pytest-asyncio==0.25.0
httpx==0.28.1
```

- [ ] **Step 2: Create package init files**

Create empty `__init__.py` in `modules/`, `routers/`, `tests/`.

- [ ] **Step 3: Create AGPL compliance NOTICE**

```
This directory contains code derived from FinceptTerminal
(https://github.com/Fincept-Corporation/FinceptTerminal)

Original code is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).
See: https://www.gnu.org/licenses/agpl-3.0.html

This project uses the code for non-commercial purposes only.
```

- [ ] **Step 4: Commit**

```bash
git add backend/services/analytics/
git commit -m "feat(analytics-service): scaffold Python analytics service with AGPL notice"
```

---

### Task 2: Extract Portfolio Management Scripts (One Clone)

**Files:**
- Create: `backend/services/analytics/modules/portfolio_optimization.py`
- Create: `backend/services/analytics/modules/risk_management.py`
- Create: `backend/services/analytics/modules/portfolio_analytics.py`
- Create: `backend/services/analytics/modules/portfolio_management.py`
- Create: `backend/services/analytics/modules/portfolio_planning.py`
- Create: `backend/services/analytics/modules/active_management.py`
- Create: `backend/services/analytics/modules/economics_markets.py`
- Create: `backend/services/analytics/modules/ffn_analysis.py`
- Create: `backend/services/analytics/modules/math_engine.py`
- Create: `backend/services/analytics/modules/behavioral_finance.py`
- Create: `backend/services/analytics/modules/etf_analytics.py`
- Create: `backend/services/analytics/modules/config.py`
- Create: `backend/services/analytics/modules/data_manager.py`
- Create: `backend/services/analytics/modules/fetch_historical.py`
- Create: `backend/services/analytics/modules/fetch_quotes.py`
- Create: `backend/services/analytics/modules/quantstats_analysis.py`

- [ ] **Step 1: Clone once and extract ALL scripts**

```bash
cd /tmp
git clone --depth 1 https://github.com/Fincept-Corporation/FinceptTerminal.git

DEST=/Users/liunian/Desktop/dnmp/py_project/backend/services/analytics/modules

# Portfolio Management modules
for f in portfolio_optimization.py risk_management.py portfolio_analytics.py \
         portfolio_management.py portfolio_planning.py active_management.py \
         economics_markets.py ffn_analysis.py math_engine.py \
         behavioral_finance.py etf_analytics.py config.py \
         data_manager.py fetch_historical.py fetch_quotes.py \
         quantstats_analysis.py; do
  cp "FinceptTerminal/fincept-qt/scripts/Analytics/portfolioManagement/$f" "$DEST/" 2>/dev/null && echo "OK: $f" || echo "MISSING: $f"
done

rm -rf FinceptTerminal
```

- [ ] **Step 2: Verify imports work**

```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend/services/analytics
python3.9 -c "
from modules.portfolio_optimization import optimize_portfolio; print('Portfolio Optimization OK')
from modules.risk_management import RiskManagement; print('Risk Management OK')
from modules.portfolio_analytics import PortfolioAnalytics; print('Portfolio Analytics OK')
"
```

- [ ] **Step 3: Commit**

```bash
git add backend/services/analytics/modules/
git commit -m "feat(analytics-service): extract Portfolio Management modules (16 files)"
```

---

### Task 3: TDD - Write Failing Tests First

**Files:**
- Create: `backend/services/analytics/tests/test_health.py`
- Create: `backend/services/analytics/tests/test_optimization.py`
- Create: `backend/services/analytics/tests/test_risk.py`
- Create: `backend/services/analytics/tests/test_analytics.py`

- [ ] **Step 1: Create health test (RED)**

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from httpx import AsyncClient, ASGITransport
from analytics_server import app


@pytest.mark.asyncio
async def test_health():
    """Test health endpoint returns 200 and expected data."""
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "healthy"
        assert "modules" in data
        assert len(data["modules"]) > 0
```

- [ ] **Step 2: Create optimization test (RED)**

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from modules.portfolio_optimization import optimize_portfolio, fetch_returns


def test_fetch_returns():
    """Test fetching returns for given symbols."""
    symbols = ["AAPL", "MSFT", "GOOGL"]
    returns = fetch_returns(symbols, period="1mo")
    assert returns is not None
    assert len(returns) > 0
    assert all(s in returns.columns for s in symbols)


def test_optimize_portfolio_max_sharpe():
    """Test portfolio optimization with max Sharpe ratio strategy."""
    symbols = ["AAPL", "MSFT", "GOOGL"]
    result = optimize_portfolio(symbols, strategy="max_sharpe")
    assert result is not None
    assert "weights" in result
    assert "return" in result
    assert "volatility" in result
    assert "sharpe_ratio" in result
    # Weights should sum to approximately 1
    assert abs(sum(result["weights"]) - 1.0) < 0.01


def test_optimize_portfolio_min_volatility():
    """Test portfolio optimization with min volatility strategy."""
    symbols = ["AAPL", "MSFT", "GOOGL"]
    result = optimize_portfolio(symbols, strategy="min_volatility")
    assert result is not None
    assert "weights" in result
    assert "return" in result
    assert "volatility" in result
    # Min volatility should have lower volatility than max sharpe
    max_sharpe_result = optimize_portfolio(symbols, strategy="max_sharpe")
    assert result["volatility"] <= max_sharpe_result["volatility"]
```

- [ ] **Step 3: Create risk management test (RED)**

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from modules.risk_management import RiskManagement, VaRCalculations


def test_risk_management_initialization():
    """Test RiskManagement class can be initialized."""
    rm = RiskManagement()
    assert rm is not None


def test_var_calculations():
    """Test VaR calculations."""
    var_calc = VaRCalculations()
    assert var_calc is not None

    # Test with sample returns
    import numpy as np
    sample_returns = np.random.normal(0.001, 0.02, 100)

    var_95 = var_calc.calculate_var(sample_returns, confidence=0.95)
    assert var_95 is not None
    assert var_95 < 0  # VaR should be negative (loss)

    var_99 = var_calc.calculate_var(sample_returns, confidence=0.99)
    assert var_99 is not None
    assert var_99 < var_95  # 99% VaR should be more negative than 95% VaR
```

- [ ] **Step 4: Create analytics test (RED)**

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from modules.portfolio_analytics import PortfolioAnalytics, CAPMAnalysis


def test_portfolio_analytics_initialization():
    """Test PortfolioAnalytics class can be initialized."""
    pa = PortfolioAnalytics()
    assert pa is not None


def test_capm_analysis():
    """Test CAPM analysis."""
    capm = CAPMAnalysis()
    assert capm is not None

    # Test CAPM calculation
    risk_free_rate = 0.04
    market_return = 0.10
    beta = 1.2

    expected_return = capm.calculate_expected_return(risk_free_rate, market_return, beta)
    assert expected_return is not None
    # CAPM: E(R) = Rf + β(Rm - Rf)
    expected = risk_free_rate + beta * (market_return - risk_free_rate)
    assert abs(expected_return - expected) < 0.001
```

- [ ] **Step 5: Run tests (should fail)**

```bash
cd backend/services/analytics && python3.9 -m pytest tests/ -v
```

- [ ] **Step 6: Commit**

```bash
git add backend/services/analytics/tests/
git commit -m "test(analytics-service): add failing tests for Portfolio Management modules"
```

---

### Task 4: FastAPI Server + Route Wrappers (GREEN)

**Files:**
- Create: `backend/services/analytics/analytics_server.py`
- Create: `backend/services/analytics/routers/optimization.py`
- Create: `backend/services/analytics/routers/risk.py`
- Create: `backend/services/analytics/routers/analytics.py`
- Create: `backend/services/analytics/routers/planning.py`

- [ ] **Step 1: Create FastAPI server**

```python
#!/usr/bin/env python3
"""
Analytics Microservice - Unified analytics API (port 8093)
Directly wraps FinceptTerminal analytics modules.
"""

import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from routers import optimization, risk, analytics, planning


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(
    title="ETF Insight Analytics Service",
    version="1.0.0",
    lifespan=lifespan,
)

allowed_origins = os.getenv(
    "ANALYTICS_CORS_ORIGINS",
    "http://localhost:5173,http://localhost:8080"
).split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

app.include_router(optimization.router, prefix="/api/optimization", tags=["Portfolio Optimization"])
app.include_router(risk.router, prefix="/api/risk", tags=["Risk Management"])
app.include_router(analytics.router, prefix="/api/analytics", tags=["Portfolio Analytics"])
app.include_router(planning.router, prefix="/api/planning", tags=["Portfolio Planning"])


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "version": "1.0.0",
        "modules": [
            "portfolio_optimization",
            "risk_management",
            "portfolio_analytics",
            "portfolio_management",
            "portfolio_planning",
            "active_management",
            "economics_markets",
            "ffn_analysis",
            "math_engine",
            "behavioral_finance",
            "etf_analytics",
        ],
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("ANALYTICS_SERVICE_PORT", "8093"))
    uvicorn.run(app, host="0.0.0.0", port=port)
```

- [ ] **Step 2: Create optimization router**

```python
from fastapi import APIRouter, Query
from modules.portfolio_optimization import optimize_portfolio, fetch_returns

router = APIRouter()


@router.post("/optimize")
def optimize(
    symbols: list[str],
    strategy: str = Query("max_sharpe", enum=["max_sharpe", "min_volatility", "equal_weight"]),
    period: str = Query("1y"),
):
    """Optimize portfolio with given strategy."""
    try:
        result = optimize_portfolio(symbols, strategy, {"period": period})
        return {"success": True, "data": result}
    except Exception as e:
        return {"success": False, "error": str(e)}


@router.post("/returns")
def get_returns(
    symbols: list[str],
    period: str = Query("1y"),
):
    """Fetch historical returns for given symbols."""
    try:
        returns = fetch_returns(symbols, period)
        return {"success": True, "data": returns.to_dict()}
    except Exception as e:
        return {"success": False, "error": str(e)}
```

- [ ] **Step 3: Create risk router**

```python
from fastapi import APIRouter, Query
from modules.risk_management import RiskManagement, VaRCalculations

router = APIRouter()
risk_mgmt = RiskManagement()
var_calc = VaRCalculations()


@router.post("/var")
def calculate_var(
    returns: list[float],
    confidence: float = Query(0.95, ge=0.9, le=0.99),
):
    """Calculate Value at Risk."""
    try:
        import numpy as np
        returns_array = np.array(returns)
        var = var_calc.calculate_var(returns_array, confidence)
        return {"success": True, "data": {"var": var, "confidence": confidence}}
    except Exception as e:
        return {"success": False, "error": str(e)}


@router.post("/cvar")
def calculate_cvar(
    returns: list[float],
    confidence: float = Query(0.95, ge=0.9, le=0.99),
):
    """Calculate Conditional Value at Risk."""
    try:
        import numpy as np
        returns_array = np.array(returns)
        cvar = var_calc.calculate_cvar(returns_array, confidence)
        return {"success": True, "data": {"cvar": cvar, "confidence": confidence}}
    except Exception as e:
        return {"success": False, "error": str(e)}
```

- [ ] **Step 4: Create analytics router**

```python
from fastapi import APIRouter, Query
from modules.portfolio_analytics import PortfolioAnalytics, CAPMAnalysis

router = APIRouter()
pa = PortfolioAnalytics()
capm = CAPMAnalysis()


@router.post("/capm")
def calculate_capm(
    risk_free_rate: float = Query(..., ge=0, le=1),
    market_return: float = Query(..., ge=0, le=1),
    beta: float = Query(..., ge=0, le=5),
):
    """Calculate expected return using CAPM."""
    try:
        expected_return = capm.calculate_expected_return(risk_free_rate, market_return, beta)
        return {
            "success": True,
            "data": {
                "expected_return": expected_return,
                "risk_free_rate": risk_free_rate,
                "market_return": market_return,
                "beta": beta,
            },
        }
    except Exception as e:
        return {"success": False, "error": str(e)}


@router.post("/metrics")
def calculate_metrics(
    symbols: list[str],
    period: str = Query("1y"),
):
    """Calculate portfolio metrics."""
    try:
        metrics = pa.calculate_portfolio_metrics(symbols, period)
        return {"success": True, "data": metrics}
    except Exception as e:
        return {"success": False, "error": str(e)}
```

- [ ] **Step 5: Create planning router**

```python
from fastapi import APIRouter, Query
from modules.portfolio_planning import PortfolioPlanning

router = APIRouter()
planning = PortfolioPlanning()


@router.post("/allocate")
def allocate_portfolio(
    symbols: list[str],
    risk_tolerance: str = Query("moderate", enum=["conservative", "moderate", "aggressive"]),
):
    """Allocate portfolio based on risk tolerance."""
    try:
        allocation = planning.allocate(symbols, risk_tolerance)
        return {"success": True, "data": allocation}
    except Exception as e:
        return {"success": False, "error": str(e)}
```

- [ ] **Step 6: Verify server starts**

```bash
cd backend/services/analytics && timeout 5 python3.9 analytics_server.py || true
```

- [ ] **Step 7: Commit**

```bash
git add backend/services/analytics/analytics_server.py backend/services/analytics/routers/
git commit -m "feat(analytics-service): add FastAPI server with 4 analytics routers"
```

---

### Task 5: Tests + Dockerfile

**Files:**
- Create: `backend/services/analytics/Dockerfile`

- [ ] **Step 1: Create Dockerfile**

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV ANALYTICS_SERVICE_PORT=8093
EXPOSE 8093

CMD ["python", "analytics_server.py"]
```

- [ ] **Step 2: Run tests**

```bash
cd backend/services/analytics && python3.9 -m pytest tests/ -v
```

- [ ] **Step 3: Commit**

```bash
git add backend/services/analytics/Dockerfile
git commit -m "feat(analytics-service): add Dockerfile"
```

---

### Task 6: Go Backend Integration

**Files:**
- Create: `backend/services/analytics/analytics_client.go`
- Create: `backend/handlers/analytics_handler.go`
- Modify: `backend/router/router.go`

- [ ] **Step 1: Create Go HTTP client**

```go
package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("ANALYTICS_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8093"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Get(path string, params map[string]string) (interface{}, error) {
	u, _ := url.Parse(c.baseURL + path)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return result, nil
}

func (c *Client) Health() (interface{}, error) {
	return c.Get("/health", nil)
}

func (c *Client) OptimizePortfolio(symbols []string, strategy string) (interface{}, error) {
	params := map[string]string{
		"strategy": strategy,
	}
	for i, s := range symbols {
		params[fmt.Sprintf("symbols[%d]", i)] = s
	}
	return c.Get("/api/optimization/optimize", params)
}

func (c *Client) CalculateVaR(returns []float64, confidence float64) (interface{}, error) {
	params := map[string]string{
		"confidence": fmt.Sprintf("%f", confidence),
	}
	return c.Get("/api/risk/var", params)
}

func (c *Client) CalculateCAPM(riskFreeRate, marketReturn, beta float64) (interface{}, error) {
	params := map[string]string{
		"risk_free_rate": fmt.Sprintf("%f", riskFreeRate),
		"market_return":  fmt.Sprintf("%f", marketReturn),
		"beta":           fmt.Sprintf("%f", beta),
	}
	return c.Get("/api/analytics/capm", params)
}
```

- [ ] **Step 2: Create handler**

```go
package handlers

import (
	"net/http"

	"etf-insight/services/analytics"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	client *analytics.Client
}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		client: analytics.NewClient(),
	}
}

func (h *AnalyticsHandler) Health(c *gin.Context) {
	result, err := h.client.Health()
	if err != nil {
		utils.Error("Analytics service health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Analytics service unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) OptimizePortfolio(c *gin.Context) {
	var req struct {
		Symbols  []string `json:"symbols" binding:"required"`
		Strategy string   `json:"strategy" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.OptimizePortfolio(req.Symbols, req.Strategy)
	if err != nil {
		utils.Error("Portfolio optimization failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) CalculateVaR(c *gin.Context) {
	var req struct {
		Returns    []float64 `json:"returns" binding:"required"`
		Confidence float64   `json:"confidence" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateVaR(req.Returns, req.Confidence)
	if err != nil {
		utils.Error("VaR calculation failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) CalculateCAPM(c *gin.Context) {
	var req struct {
		RiskFreeRate float64 `json:"risk_free_rate" binding:"required"`
		MarketReturn float64 `json:"market_return" binding:"required"`
		Beta         float64 `json:"beta" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateCAPM(req.RiskFreeRate, req.MarketReturn, req.Beta)
	if err != nil {
		utils.Error("CAPM calculation failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
```

- [ ] **Step 3: Add routes to router.go**

Follow the same pattern as Agent routes:
- Add `Analytics *handlers.AnalyticsHandler` to Handlers struct
- Initialize in NewRouter: `h.Analytics = handlers.NewAnalyticsHandler()`
- Add `r.registerAnalyticsRoutes()` with routes:

```go
func (r *Router) registerAnalyticsRoutes() {
	a := r.engine.Group("/api/analytics")
	{
		a.GET("/health", r.handlers.Analytics.Health)
		a.POST("/optimize", r.handlers.Analytics.OptimizePortfolio)
		a.POST("/var", r.handlers.Analytics.CalculateVaR)
		a.POST("/capm", r.handlers.Analytics.CalculateCAPM)
	}
}
```

- [ ] **Step 4: Verify Go compiles**

```bash
cd backend && go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add backend/services/analytics/analytics_client.go backend/handlers/analytics_handler.go backend/router/router.go
git commit -m "feat(analytics): add Go client, handler, and routes for analytics service"
```

---

### Task 7: Design Document + README Updates

- [ ] **Step 1: Update design document progress section**

Update `docs/superpowers/specs/2026-05-04-fincept-integration-design.md` Phase 4 status to ✅ Complete.

- [ ] **Step 2: Update README.md**

- Add Analytics Microservice feature section
- Add port 8093 to service list
- Update roadmap: Phase 4 ✅

- [ ] **Step 3: Update backend/README.md**

- Add Analytics Service feature section
- Add `services/analytics/` to project structure
- Add Analytics API endpoints section
- Add port 8093

- [ ] **Step 4: Commit**

```bash
git add docs/ README.md backend/README.md
git commit -m "docs: update Phase 4 status to complete in all documentation"
```

---

## TDD Cycle Summary

```
Task 3: RED (Write failing tests)
    ↓
Task 4: GREEN (Implement minimal code to pass)
    ↓
Task 5: REFACTOR (Clean up, add Dockerfile)
    ↓
Task 6: INTEGRATE (Go backend)
    ↓
Task 7: DOCUMENT (Update docs)
```

---

## Success Criteria

- [ ] All tests pass (100% pass rate)
- [ ] Go code compiles without errors
- [ ] FastAPI server starts on port 8093
- [ ] Health endpoint returns expected data
- [ ] Portfolio optimization works with max_sharpe strategy
- [ ] VaR calculation works with sample data
- [ ] CAPM calculation returns correct expected return
- [ ] Go proxy routes work correctly
- [ ] Documentation updated

---

*Plan created: 2026-05-06*
*TDD approach: RED → GREEN → REFACTOR*
