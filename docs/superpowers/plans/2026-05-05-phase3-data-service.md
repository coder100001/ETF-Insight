# Phase 3: Data Source Microservice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Directly integrate FinceptTerminal's data source scripts into a unified FastAPI data service (port 8092), with Go backend proxy.

**Architecture:** Extract Python data source scripts from FinceptTerminal (AGPL, non-commercial use), wrap them in a FastAPI service. Go backend acts as HTTP proxy. No rewriting — direct code reuse.

**Tech Stack:** Python 3.9+, FastAPI, uvicorn, requests, pandas, yfinance, akshare. Go: net/http, gin.

**License:** AGPL-3.0 (direct code reuse, non-commercial). A `NOTICE` file must be added to `backend/services/data/`.

---

## File Structure

```
backend/services/data/
├── data_server.py              # FastAPI entry point (port 8092)
├── requirements.txt            # Python dependencies
├── Dockerfile                  # Container
├── NOTICE                      # AGPL-3.0 compliance notice
├── sources/                    # Extracted data source scripts
│   ├── __init__.py
│   ├── fred_data.py            # From FinceptTerminal
│   ├── worldbank_data.py       # From FinceptTerminal (actual name, not world_bank_data.py)
│   ├── imf_data.py             # From FinceptTerminal
│   ├── yfinance_data.py        # From FinceptTerminal
│   ├── coingecko.py            # From FinceptTerminal (actual name, not coingecko_data.py)
│   ├── akshare_data.py         # From FinceptTerminal (orchestrator)
│   ├── akshare_stocks_realtime.py
│   ├── akshare_macro.py
│   ├── akshare_crypto.py
│   ├── akshare_bonds.py
│   ├── akshare_analysis.py
│   ├── akshare_derivatives.py
│   ├── akshare_economics_china.py
│   ├── akshare_economics_global.py
│   ├── akshare_alternative.py
│   └── akshare_funds_expanded.py
├── routers/                    # FastAPI route modules
│   ├── __init__.py
│   ├── fred.py
│   ├── worldbank.py
│   ├── imf.py
│   ├── yfinance.py
│   ├── akshare.py
│   └── coingecko.py
├── tests/
│   ├── __init__.py
│   ├── test_health.py
│   └── test_worldbank.py       # No API key needed
backend/services/data/data_client.go    # Go HTTP client
backend/handlers/data_handler.go        # Go Gin handler
backend/router/router.go                # MODIFY: add data routes
```

## Go Route Mapping

Go 后端 `/api/data/*` 代理转发到 Python 服务 `localhost:8092`：

| Go Route | Python Route | 说明 |
|----------|-------------|------|
| `GET /api/data/health` | `GET /health` | 健康检查 |
| `GET /api/data/fred/series/:id` | `GET /api/fred/series/:id` | FRED 时间序列 |
| `GET /api/data/fred/search` | `GET /api/fred/search` | FRED 搜索 |
| `GET /api/data/worldbank/indicators/:country/:indicator` | `GET /api/worldbank/indicators/:country/:indicator` | World Bank 指标 |
| `GET /api/data/worldbank/snapshot/:country` | `GET /api/worldbank/snapshot/:country` | 国家经济快照 |
| `GET /api/data/imf/economic-indicators` | `GET /api/imf/economic-indicators` | IMF 经济指标 |
| `GET /api/data/yfinance/quote/:symbol` | `GET /api/yfinance/quote/:symbol` | Yahoo 实时报价 |
| `GET /api/data/yfinance/historical/:symbol` | `GET /api/yfinance/historical/:symbol` | Yahoo 历史数据 |
| `GET /api/data/yfinance/info/:symbol` | `GET /api/yfinance/info/:symbol` | Yahoo 公司信息 |
| `POST /api/data/yfinance/batch-quotes` | `POST /api/yfinance/batch-quotes` | Yahoo 批量报价 |
| `GET /api/data/akshare/stock/spot` | `GET /api/akshare/stock/spot` | AkShare A股实时行情 |
| `GET /api/data/akshare/stock/daily/:symbol` | `GET /api/akshare/stock/daily/:symbol` | AkShare 日线数据 |
| `GET /api/data/akshare/macro/:indicator` | `GET /api/akshare/macro/:indicator` | AkShare 宏观经济指标 |
| `GET /api/data/coingecko/price` | `GET /api/coingecko/price` | CoinGecko 价格 |
| `GET /api/data/coingecko/markets` | `GET /api/coingecko/markets` | CoinGecko 市场数据 |

> Go 端使用通配路由 `r.engine.Group("/api/data")` 透传，不做路径重写。

---

### Task 1: Project Scaffolding + Requirements + AGPL Notice

**Files:**
- Create: `backend/services/data/requirements.txt`
- Create: `backend/services/data/sources/__init__.py`
- Create: `backend/services/data/routers/__init__.py`
- Create: `backend/services/data/tests/__init__.py`
- Create: `backend/services/data/NOTICE`

- [ ] **Step 1: Create requirements.txt**

```txt
fastapi==0.115.6
uvicorn[standard]==0.34.0
requests==2.32.3
pandas==2.0.3
yfinance==0.2.36
akshare==1.14.50
python-dotenv==1.0.1
pytest==8.3.4
pytest-asyncio==0.25.0
httpx==0.28.1
```

- [ ] **Step 2: Create package init files**

Create empty `__init__.py` in `sources/`, `routers/`, `tests/`.

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
git add backend/services/data/
git commit -m "feat(data-service): scaffold Python data service with AGPL notice"
```

---

### Task 2: Extract ALL FinceptTerminal Scripts (One Clone)

**Files:**
- Create: `backend/services/data/sources/fred_data.py`
- Create: `backend/services/data/sources/worldbank_data.py`
- Create: `backend/services/data/sources/imf_data.py`
- Create: `backend/services/data/sources/yfinance_data.py`
- Create: `backend/services/data/sources/coingecko.py`
- Create: `backend/services/data/sources/akshare_data.py`
- Create: `backend/services/data/sources/akshare_stocks_realtime.py`
- Create: `backend/services/data/sources/akshare_macro.py`
- Create: `backend/services/data/sources/akshare_crypto.py`
- Create: `backend/services/data/sources/akshare_bonds.py`
- Create: `backend/services/data/sources/akshare_analysis.py`
- Create: `backend/services/data/sources/akshare_derivatives.py`
- Create: `backend/services/data/sources/akshare_economics_china.py`
- Create: `backend/services/data/sources/akshare_economics_global.py`
- Create: `backend/services/data/sources/akshare_alternative.py`
- Create: `backend/services/data/sources/akshare_funds_expanded.py`

- [ ] **Step 1: Clone once and extract ALL scripts**

```bash
cd /tmp
git clone --depth 1 https://github.com/Fincept-Corporation/FinceptTerminal.git

DEST=/Users/liunian/Desktop/dnmp/py_project/backend/services/data/sources

# Core data sources
for f in fred_data.py worldbank_data.py imf_data.py yfinance_data.py coingecko.py; do
  cp "FinceptTerminal/fincept-qt/scripts/$f" "$DEST/" 2>/dev/null && echo "OK: $f" || echo "MISSING: $f"
done

# AkShare modules
for f in akshare_data.py akshare_stocks_realtime.py akshare_macro.py \
         akshare_crypto.py akshare_bonds.py akshare_analysis.py \
         akshare_derivatives.py akshare_economics_china.py \
         akshare_economics_global.py akshare_alternative.py \
         akshare_funds_expanded.py; do
  cp "FinceptTerminal/fincept-qt/scripts/$f" "$DEST/" 2>/dev/null && echo "OK: $f" || echo "MISSING: $f"
done

rm -rf FinceptTerminal
```

- [ ] **Step 2: Verify imports work**

```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend/services/data
python3.9 -c "
from sources.fred_data import get_series; print('FRED OK')
from sources.worldbank_data import get_indicators; print('WB OK')
from sources.imf_data import IMFDataWrapper; print('IMF OK')
from sources.coingecko import get_simple_price; print('CoinGecko OK')
"
```

> Note: `yfinance_data.py` depends on `yfinance` pip package. `akshare_*.py` depends on `akshare` package. These will be verified after Task 1 requirements are installed.

- [ ] **Step 3: Commit**

```bash
git add backend/services/data/sources/
git commit -m "feat(data-service): extract all FinceptTerminal data source scripts (16 files)"
```

---

### Task 3: FastAPI Server + Route Wrappers

**Files:**
- Create: `backend/services/data/data_server.py`
- Create: `backend/services/data/routers/fred.py`
- Create: `backend/services/data/routers/worldbank.py`
- Create: `backend/services/data/routers/imf.py`
- Create: `backend/services/data/routers/yfinance.py`
- Create: `backend/services/data/routers/akshare.py`
- Create: `backend/services/data/routers/coingecko.py`

- [ ] **Step 1: Create FastAPI server**

```python
#!/usr/bin/env python3
"""
Data Source Microservice - Unified data API (port 8092)
Directly wraps FinceptTerminal data source scripts.
"""

import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from routers import fred, worldbank, imf, yfinance, akshare, coingecko


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(
    title="ETF Insight Data Service",
    version="1.0.0",
    lifespan=lifespan,
)

allowed_origins = os.getenv(
    "DATA_CORS_ORIGINS",
    "http://localhost:5173,http://localhost:8080"
).split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

app.include_router(fred.router, prefix="/api/fred", tags=["FRED"])
app.include_router(worldbank.router, prefix="/api/worldbank", tags=["World Bank"])
app.include_router(imf.router, prefix="/api/imf", tags=["IMF"])
app.include_router(yfinance.router, prefix="/api/yfinance", tags=["Yahoo Finance"])
app.include_router(akshare.router, prefix="/api/akshare", tags=["AkShare"])
app.include_router(coingecko.router, prefix="/api/coingecko", tags=["CoinGecko"])


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "version": "1.0.0",
        "sources": ["fred", "worldbank", "imf", "yfinance", "akshare", "coingecko"],
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("DATA_SERVICE_PORT", "8092"))
    uvicorn.run(app, host="0.0.0.0", port=port)
```

- [ ] **Step 2: Create FRED router**

```python
from fastapi import APIRouter, Query
from sources.fred_data import get_series, search_series, get_categories

router = APIRouter()


@router.get("/series/{series_id}")
def fred_series(
    series_id: str,
    start_date: str = Query(None),
    end_date: str = Query(None),
    frequency: str = Query(None),
    transform: str = Query(None),
):
    return get_series(series_id, start_date, end_date, frequency, transform)


@router.get("/search")
def fred_search(q: str, limit: int = Query(10, ge=1, le=100)):
    return search_series(q, limit)


@router.get("/categories/{category_id}")
def fred_categories(category_id: str = "0"):
    return get_categories(category_id)
```

- [ ] **Step 3: Create World Bank router**

```python
from fastapi import APIRouter, Query
from sources.worldbank_data import get_indicators, get_countries, get_economic_snapshot

router = APIRouter()


@router.get("/indicators/{country_code}/{indicator}")
def wb_indicators(
    country_code: str,
    indicator: str,
    date_range: str = Query(None),
):
    return get_indicators(country_code, indicator, date_range)


@router.get("/countries")
def wb_countries(
    region: str = Query(None),
    income_level: str = Query(None),
):
    return get_countries(region, income_level)


@router.get("/snapshot/{country_code}")
def wb_snapshot(country_code: str):
    return get_economic_snapshot(country_code)
```

- [ ] **Step 4: Create IMF router**

```python
from fastapi import APIRouter, Query
from sources.imf_data import IMFDataWrapper

router = APIRouter()
wrapper = IMFDataWrapper()


@router.get("/economic-indicators")
def imf_indicators(
    countries: str = Query(None),
    symbols: str = Query(None),
    frequency: str = Query("A"),
):
    return wrapper.get_economic_indicators(countries, symbols, frequency)


@router.get("/trade")
def imf_trade(
    countries: str = Query(None),
    counterparts: str = Query(None),
    direction: str = Query("E"),
):
    return wrapper.get_direction_of_trade(countries, counterparts, direction)
```

- [ ] **Step 5: Create Yahoo Finance router**

```python
from fastapi import APIRouter, Query
from sources.yfinance_data import (
    get_quote, get_historical, get_info,
    get_financials, get_batch_quotes, search_symbols,
)

router = APIRouter()


@router.get("/quote/{symbol}")
def yf_quote(symbol: str):
    return get_quote(symbol)


@router.get("/historical/{symbol}")
def yf_historical(
    symbol: str,
    start_date: str = Query(None),
    end_date: str = Query(None),
    interval: str = Query("1d"),
):
    return get_historical(symbol, start_date, end_date, interval)


@router.get("/info/{symbol}")
def yf_info(symbol: str):
    return get_info(symbol)


@router.get("/financials/{symbol}")
def yf_financials(symbol: str):
    return get_financials(symbol)


@router.post("/batch-quotes")
def yf_batch_quotes(symbols: list[str]):
    return get_batch_quotes(symbols)


@router.get("/search")
def yf_search(q: str, limit: int = Query(10)):
    return search_symbols(q, limit)
```

- [ ] **Step 6: Create AkShare router (expanded)**

```python
from fastapi import APIRouter, Query
from sources.akshare_data import AKShareDataWrapper

router = APIRouter()
wrapper = AKShareDataWrapper()


# ---- Stock endpoints ----
@router.get("/stock/spot")
def ak_stock_spot():
    """A股实时行情"""
    return wrapper.get_stock_zh_a_spot()


@router.get("/stock/daily/{symbol}")
def ak_stock_daily(
    symbol: str,
    start_date: str = Query(None),
    end_date: str = Query(None),
):
    """A股日线数据"""
    return wrapper.get_stock_zh_a_daily(symbol, start_date, end_date)


@router.get("/stock/hk/spot")
def ak_stock_hk_spot():
    """港股实时行情"""
    from sources.akshare_stocks_realtime import get_stock_hk_spot
    return get_stock_hk_spot()


@router.get("/stock/us/spot")
def ak_stock_us_spot():
    """美股实时行情"""
    from sources.akshare_stocks_realtime import get_stock_us_spot
    return get_stock_us_spot()


# ---- Macro endpoints ----
@router.get("/macro/gdp")
def ak_macro_gdp():
    """中国GDP数据"""
    return wrapper.get_macro_china_gdp()


@router.get("/macro/{indicator}")
def ak_macro(indicator: str):
    """通用宏观经济指标"""
    func = getattr(wrapper, f"get_macro_china_{indicator}", None)
    if func:
        return func()
    return {"error": f"Unknown indicator: {indicator}"}


# ---- Bond endpoints ----
@router.get("/bonds/{bond_type}")
def ak_bonds(bond_type: str):
    """债券数据"""
    from sources.akshare_bonds import BondsWrapper
    bw = BondsWrapper()
    func = getattr(bw, f"get_{bond_type}", None)
    if func:
        return func()
    return {"error": f"Unknown bond type: {bond_type}"}


# ---- Crypto endpoints ----
@router.get("/crypto/spot")
def ak_crypto_spot():
    """加密货币现货"""
    from sources.akshare_crypto import get_crypto_js_spot
    return get_crypto_js_spot()
```

- [ ] **Step 7: Create CoinGecko router**

```python
from fastapi import APIRouter, Query
from sources.coingecko import (
    get_simple_price, get_coin_markets, get_coin_details,
    get_market_chart, get_trending_coins, get_global_data,
)

router = APIRouter()


@router.get("/price")
def cg_price(
    ids: str,
    vs_currencies: str = Query("usd"),
):
    return get_simple_price(ids, vs_currencies)


@router.get("/markets")
def cg_markets(
    vs_currency: str = Query("usd"),
    per_page: int = Query(20, ge=1, le=250),
):
    return get_coin_markets(vs_currency, per_page=per_page)


@router.get("/coin/{coin_id}")
def cg_coin(coin_id: str):
    return get_coin_details(coin_id)


@router.get("/chart/{coin_id}")
def cg_chart(
    coin_id: str,
    vs_currency: str = Query("usd"),
    days: str = Query("30"),
):
    return get_market_chart(coin_id, vs_currency, days)


@router.get("/trending")
def cg_trending():
    return get_trending_coins()


@router.get("/global")
def cg_global():
    return get_global_data()
```

- [ ] **Step 8: Verify server starts**

```bash
cd backend/services/data && timeout 5 python3.9 data_server.py || true
```

- [ ] **Step 9: Commit**

```bash
git add backend/services/data/data_server.py backend/services/data/routers/
git commit -m "feat(data-service): add FastAPI server with 6 data source routers"
```

---

### Task 4: Tests + Dockerfile

**Files:**
- Create: `backend/services/data/tests/test_health.py`
- Create: `backend/services/data/tests/test_worldbank.py`
- Create: `backend/services/data/Dockerfile`

- [ ] **Step 1: Create health test**

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from httpx import AsyncClient, ASGITransport
from data_server import app


@pytest.mark.asyncio
async def test_health():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "healthy"
        assert len(data["sources"]) == 6
```

- [ ] **Step 2: Create World Bank test (no API key needed)**

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from sources.worldbank_data import get_countries, get_indicators


def test_get_countries():
    result = get_countries()
    assert result is not None


def test_get_indicators_china_gdp():
    result = get_indicators("CHN", "NY.GDP.MKTP.CD", "2020:2023")
    assert result is not None
```

- [ ] **Step 3: Create Dockerfile**

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV DATA_SERVICE_PORT=8092
EXPOSE 8092

CMD ["python", "data_server.py"]
```

- [ ] **Step 4: Run tests**

```bash
cd backend/services/data && python3.9 -m pytest tests/ -v
```

- [ ] **Step 5: Commit**

```bash
git add backend/services/data/tests/ backend/services/data/Dockerfile
git commit -m "feat(data-service): add tests and Dockerfile"
```

---

### Task 5: Go Backend Integration

**Files:**
- Create: `backend/services/data/data_client.go`
- Create: `backend/handlers/data_handler.go`
- Modify: `backend/router/router.go`

- [ ] **Step 1: Create Go HTTP client**

```go
package data

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
	baseURL := os.Getenv("DATA_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8092"
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

func (c *Client) FredSeries(seriesID string, params map[string]string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/fred/series/%s", seriesID), params)
}

func (c *Client) WorldBankIndicators(country, indicator string, params map[string]string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/worldbank/indicators/%s/%s", country, indicator), params)
}

func (c *Client) YFinanceQuote(symbol string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/yfinance/quote/%s", symbol), nil)
}

func (c *Client) YFinanceHistorical(symbol string, params map[string]string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/yfinance/historical/%s", symbol), params)
}

func (c *Client) AkShareStockSpot() (interface{}, error) {
	return c.Get("/api/akshare/stock/spot", nil)
}

func (c *Client) CoinGeckoPrice(ids, vsCurrencies string) (interface{}, error) {
	return c.Get("/api/coingecko/price", map[string]string{"ids": ids, "vs_currencies": vsCurrencies})
}
```

- [ ] **Step 2: Create handler**

```go
package handlers

import (
	"net/http"
	"etf-insight/services/data"
	"etf-insight/utils"
	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	client *data.Client
}

func NewDataHandler() *DataHandler {
	return &DataHandler{client: data.NewClient()}
}

func (h *DataHandler) Health(c *gin.Context) {
	result, err := h.client.Health()
	if err != nil {
		utils.Error("Data service health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Data service unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *DataHandler) FredSeries(c *gin.Context) {
	seriesID := c.Param("series_id")
	params := map[string]string{}
	if v := c.Query("start_date"); v != "" { params["start_date"] = v }
	if v := c.Query("end_date"); v != "" { params["end_date"] = v }

	result, err := h.client.FredSeries(seriesID, params)
	if err != nil {
		utils.Error("FRED request failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *DataHandler) YFinanceQuote(c *gin.Context) {
	symbol := c.Param("symbol")
	result, err := h.client.YFinanceQuote(symbol)
	if err != nil {
		utils.Error("Yahoo Finance request failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *DataHandler) AkShareStockSpot(c *gin.Context) {
	result, err := h.client.AkShareStockSpot()
	if err != nil {
		utils.Error("AkShare request failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
```

- [ ] **Step 3: Add routes to router.go**

Follow the same pattern as Agent routes:
- Add `Data *handlers.DataHandler` to Handlers struct
- Initialize in NewRouter: `h.Data = handlers.NewDataHandler()`
- Add `r.registerDataRoutes()` with routes:

```go
func (r *Router) registerDataRoutes() {
	d := r.engine.Group("/api/data")
	{
		d.GET("/health", r.handlers.Data.Health)
		d.GET("/fred/series/:series_id", r.handlers.Data.FredSeries)
		d.GET("/yfinance/quote/:symbol", r.handlers.Data.YFinanceQuote)
		d.GET("/akshare/stock/spot", r.handlers.Data.AkShareStockSpot)
	}
}
```

- [ ] **Step 4: Verify Go compiles**

```bash
cd backend && go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add backend/services/data/data_client.go backend/handlers/data_handler.go backend/router/router.go
git commit -m "feat(data): add Go client, handler, and routes for data service"
```

---

### Task 6: Design Document + README Updates

- [ ] **Step 1: Update design document progress section**

Update `docs/superpowers/specs/2026-05-04-fincept-integration-design.md` Phase 3 status to ✅ Complete.

- [ ] **Step 2: Update README.md**

- v2.9.0 → v2.10.0 (if applicable) or add v2.10 update block
- Add Data Source Microservice feature section
- Add port 8092 to service list
- Update roadmap: Phase 3 ✅

- [ ] **Step 3: Update README_EN.md**

Same changes as README.md in English.

- [ ] **Step 4: Update backend/README.md**

- Add Data Service feature section
- Add `services/data/` to project structure
- Add Data API endpoints section
- Add port 8092

- [ ] **Step 5: Commit**

```bash
git add docs/ README.md README_EN.md backend/README.md
git commit -m "docs: update Phase 3 status to complete in all documentation"
```
