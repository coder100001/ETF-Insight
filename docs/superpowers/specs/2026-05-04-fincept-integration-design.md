# FinceptTerminal Integration Design

## Summary

Integrate FinceptTerminal's financial intelligence capabilities into ETF-Insight via a **hybrid architecture**: Phase 1 uses direct API integration (QuantLib cloud), Phase 2-4 use Python microservice bridge for extracted modules. Start with Phase 1 for immediate value, then progressively add AI Agents, data sources, and analytics.

## Architecture Decision

**Chosen approach**: Hybrid — Direct API + Python Microservice Bridge

### Decision Rationale

The brainstorming session recommended "方案 A: 插件化架构集成" (Plugin Architecture). We adopt a hybrid approach that aligns with the plugin philosophy while respecting FinceptTerminal's architecture constraints:

| Phase | Approach | Why |
|-------|----------|-----|
| Phase 1 | Direct API Client | QuantLib is already a cloud API — no code extraction needed, just HTTP client |
| Phase 2-4 | Python Microservice Bridge | FinceptTerminal's Python layer (250+ scripts) is cleanly separated from C++ UI via `PythonRunner` (QProcess-based IPC). We extract Python modules as FastAPI services without touching C++ code |

### Why not pure plugin architecture for Phase 1

- QuantLib API is a **cloud service** (`api.fincept.in`), not a local library
- Writing a Go HTTP client is simpler and more reliable than a plugin wrapper
- No AGPL license concerns — we're calling a public API, not using their code

### Alternatives considered

| Alternative | Verdict | Reason |
|-------------|---------|--------|
| Desktop engine + Web frontend | Rejected | Requires Qt6 dependency, too complex |
| C++ Shared Library (CGO) | Rejected | Integration difficulty too high, ABI compatibility issues |
| Docker + API Gateway | Deferred | Over-engineered for current scale, revisit if we exceed 3 microservices |
| Pure plugin architecture | Partially adopted | Used for Phase 2-4 where we control the code |

## Target Architecture

```
React Frontend (现有)
    │
Go Backend (现有 + QuantLib Client + Python Bridge Client)
    │
    ├── QuantLib API (cloud)          - api.fincept.in (Phase 1, direct HTTP)
    │
    └── Python Microservice Cluster (Phase 2-4, 新增)
        ├── Agent Service (port 8091)  - 30+ AI agents, multi-LLM
        ├── Data Service (port 8092)   - 60+ data sources
        └── Analytics Service (port 8093) - 80+ analytics modules
```

### Data Flow

```
Frontend → Go API Handler → QuantLib Client → api.fincept.in/quantlib/
                                    │
                                    ├─ Cache hit → return cached result
                                    ├─ API success → cache + return
                                    └─ API failure → fallback to local calc / return error
```

---

## Phase 1: QuantLib Direct Connection

### Goal

Connect to FinceptTerminal's existing QuantLib cloud API (`api.fincept.in/quantlib/`) to add institutional-grade quantitative analysis capabilities.

### What We Get

| Capability | Description | ETF-Insight Integration Point |
|-----------|-------------|-------------------------------|
| **Options Pricing** | Black-Scholes, Binomial, Monte Carlo | New: Options analysis tab |
| **Greeks Calculation** | Delta, Gamma, Theta, Vega, Rho | New: Greeks dashboard |
| **Yield Curves** | Construction, interpolation, zero rates | Enhance: bond/fixed income analysis |
| **Fixed Income** | Bond pricing, duration, convexity | New: Bond calculator |
| **Risk Metrics** | VaR, CVaR with QuantLib engine | Enhance: existing `risk_models.go` |
| **Stochastic Processes** | GBM, CIR, Hull-White | Enhance: scenario analysis |
| **Volatility Surfaces** | Implied vol, local vol | New: Vol surface visualization |

### API Design

FinceptTerminal's QuantLib API format (verified from `QuantLibClient.cpp` source code):

```
Base URL: https://api.fincept.in/quantlib/

POST Endpoints (request body as JSON):
- POST /options/european              - European option pricing
- POST /options/american              - American option pricing
- POST /bonds/fixed                   - Fixed bond pricing
- POST /yield-curve/build             - Yield curve construction
- POST /risk/var                      - Value at Risk calculation

GET Endpoints (cached, 1-hour TTL):
- GET  /core/types/currencies         - Supported currencies
- GET  /core/types/frequencies        - Payment frequencies
- GET  /scheduling/calendar/list      - Available calendars
- GET  /scheduling/daycount/conventions - Day count conventions
- GET  /scheduling/adjustment/methods - Adjustment methods

Query Param Endpoints (POST with empty body, params as query string):
- POST /core/types/spread/from-bps    - Spread from basis points
```

### Authentication

The QuantLib API uses `X-API-Key` header for authentication (verified from QuantLibClient.cpp):

```cpp
auto& auth_mgr = auth::AuthManager::instance();
if (auth_mgr.is_authenticated())
    req.setRawHeader("X-API-Key", auth_mgr.session().api_key.toUtf8());
```

**Important**: The API key is NOT optional. While some endpoints may work without authentication, production use requires a valid API key. We must:

1. Register for a FinceptTerminal API key
2. Store it securely via environment variable
3. Include it in all requests
4. Handle 401/403 responses gracefully

Request/Response format:
```json
// Request: European option pricing
{
  "spot": 100.0,
  "strike": 105.0,
  "rate": 0.05,
  "volatility": 0.20,
  "time_to_expiry": 1.0,
  "option_type": "call"
}

// Response (API envelope format)
{
  "success": true,
  "message": "Option priced successfully",
  "data": {
    "price": 8.02,
    "delta": 0.54,
    "gamma": 0.019,
    "theta": -0.015,
    "vega": 0.38,
    "rho": 0.46
  }
}

// Error response (FastAPI 422 validation)
{
  "detail": [
    {
      "loc": ["body", "spot"],
      "msg": "ensure this value is greater than 0",
      "type": "value_error.number.not_gt"
    }
  ]
}
```

### Precision Handling (Critical)

ETF-Insight mandates `decimal.Decimal` for all financial calculations. The QuantLib API returns `float64`. We must:

1. **API Response → Go**: Parse JSON floats into `decimal.Decimal` immediately upon deserialization
2. **Go → Frontend**: Serialize `decimal.Decimal` as strings (not floats) to avoid precision loss. Use `json:",string"` tag to force JSON string representation
3. **Frontend Display**: Use `toFixed()` or `Intl.NumberFormat` for display, never raw float arithmetic
4. **HKEX Compatibility**: All price fields must support HKEX tick sizes and lot sizes

```go
type OptionResult struct {
    Price  decimal.Decimal `json:"price,string"`
    Delta  decimal.Decimal `json:"delta,string"`
    Gamma  decimal.Decimal `json:"gamma,string"`
    Theta  decimal.Decimal `json:"theta,string"`
    Vega   decimal.Decimal `json:"vega,string"`
    Rho    decimal.Decimal `json:"rho,string"`
}
```

### Implementation Plan

#### 1. Go Backend - Data Models

New file: `models/quantlib.go`

> **WARNING**: Gin's binding validator does not support `gt` (greater than) on `decimal.Decimal` (it's a struct, not a primitive numeric type). All `decimal.Decimal` fields below use `binding:"required"` only. Range validation must be performed explicitly in handler code.

```go
type EuropeanOptionRequest struct {
    Spot         decimal.Decimal `json:"spot" binding:"required"`
    Strike       decimal.Decimal `json:"strike" binding:"required"`
    Rate         decimal.Decimal `json:"rate" binding:"required"`
    Volatility   decimal.Decimal `json:"volatility" binding:"required"`
    TimeToExpiry decimal.Decimal `json:"time_to_expiry" binding:"required"`
    OptionType   string          `json:"option_type" binding:"required,oneof=call put"`
}

type AmericanOptionRequest struct {
    Spot         decimal.Decimal `json:"spot" binding:"required"`
    Strike       decimal.Decimal `json:"strike" binding:"required"`
    Rate         decimal.Decimal `json:"rate" binding:"required"`
    Volatility   decimal.Decimal `json:"volatility" binding:"required"`
    TimeToExpiry decimal.Decimal `json:"time_to_expiry" binding:"required"`
    OptionType   string          `json:"option_type" binding:"required,oneof=call put"`
    Steps        int             `json:"steps" binding:"omitempty,min=10,max=1000"`
}

type OptionResult struct {
    Price  decimal.Decimal `json:"price,string"`
    Delta  decimal.Decimal `json:"delta,string"`
    Gamma  decimal.Decimal `json:"gamma,string"`
    Theta  decimal.Decimal `json:"theta,string"`
    Vega   decimal.Decimal `json:"vega,string"`
    Rho    decimal.Decimal `json:"rho,string"`
}

type YieldCurveRequest struct {
    CurveType  string          `json:"curve_type" binding:"required,oneof=flat zero forward"`
    Rate       decimal.Decimal `json:"rate,string" binding:"required"`
    DayCount   string          `json:"day_count" binding:"required"`
    Calendar   string          `json:"calendar" binding:"required"`
    Tenors     []string        `json:"tenors" binding:"required,min=1"`
}

type BondRequest struct {
    FaceValue    decimal.Decimal `json:"face_value,string" binding:"required"`
    CouponRate   decimal.Decimal `json:"coupon_rate,string" binding:"required"`
    Frequency    string          `json:"frequency" binding:"required"`
    Settlement   string          `json:"settlement" binding:"required"`
    Maturity     string          `json:"maturity" binding:"required"`
    Yield        decimal.Decimal `json:"yield,string" binding:"required"`
}

type QuantLibVaRRequest struct {
    PortfolioValue decimal.Decimal `json:"portfolio_value,string" binding:"required"`
    Confidence    decimal.Decimal `json:"confidence,string" binding:"required"`
    Horizon       int             `json:"horizon" binding:"required,gt=0"`
    Method        string          `json:"method" binding:"required,oneof=historical parametric monte_carlo"`
}
```

#### 2. Go Backend - QuantLib Client

New file: `services/quantlib/quantlib_client.go`

```go
type QuantLibClient struct {
    baseURL       string
    httpClient    *http.Client
    apiKey        string
    cache         engine.CacheService
    retryCount    int
    retryDelay    time.Duration
    requestTimeout time.Duration
}

func NewQuantLibClient(cfg config.QuantLibConfig, cache engine.CacheService) *QuantLibClient {
    return &QuantLibClient{
        baseURL: cfg.APIURL,
        httpClient: &http.Client{
            Timeout: cfg.RequestTimeout,
            Transport: &http.Transport{
                MaxIdleConns:        10,
                MaxConnsPerHost:     10,
                IdleConnTimeout:     90 * time.Second,
                DisableKeepAlives:   false,
            },
        },
        apiKey:         cfg.APIKey,
        cache:          cache,
        retryCount:     3,
        retryDelay:     100 * time.Millisecond,
        requestTimeout: cfg.RequestTimeout,
    }
}

// Required imports:
//   "etf-insight/services/engine"
//   "github.com/shopspring/decimal"

func (c *QuantLibClient) PriceEuropeanOption(ctx context.Context, req EuropeanOptionRequest) (*OptionResult, error)
func (c *QuantLibClient) PriceAmericanOption(ctx context.Context, req AmericanOptionRequest) (*OptionResult, error)
func (c *QuantLibClient) BuildYieldCurve(ctx context.Context, req YieldCurveRequest) (*YieldCurveResult, error)
func (c *QuantLibClient) PriceBond(ctx context.Context, req BondRequest) (*BondResult, error)
func (c *QuantLibClient) CalculateVaR(ctx context.Context, req QuantLibVaRRequest) (*VaRResult, error)
func (c *QuantLibClient) GetCurrencies(ctx context.Context) ([]Currency, error)
func (c *QuantLibClient) GetCalendars(ctx context.Context) ([]Calendar, error)
```

#### 3. Error Handling and Resilience

New file: `services/quantlib/resilience.go`

```go
type QuantLibError struct {
    StatusCode int
    Endpoint   string
    Message    string
    Retryable  bool
}

func (e *QuantLibError) Error() string {
    return fmt.Sprintf("QuantLib API error: %s (endpoint=%s, status=%d, retryable=%v)",
        e.Message, e.Endpoint, e.StatusCode, e.Retryable)
}

func classifyError(statusCode int) QuantLibError {
    switch {
    case statusCode == 401 || statusCode == 403:
        return QuantLibError{StatusCode: statusCode, Retryable: false, Message: "authentication failed"}
    case statusCode == 422:
        return QuantLibError{StatusCode: statusCode, Retryable: false, Message: "validation error"}
    case statusCode == 429:
        return QuantLibError{StatusCode: statusCode, Retryable: true, Message: "rate limited"}
    case statusCode >= 500:
        return QuantLibError{StatusCode: statusCode, Retryable: true, Message: "server error"}
    default:
        return QuantLibError{StatusCode: statusCode, Retryable: false, Message: "unexpected error"}
    }
}

func (c *QuantLibClient) doWithRetry(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
    var lastErr error
    for attempt := 0; attempt <= c.retryCount; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(time.Duration(float64(c.retryDelay) * float64(1<<attempt) * (0.5 + rand.Float64()*0.5))):
            }
        }
        err := c.doRequest(ctx, endpoint, body, result)
        if err == nil {
            return nil
        }
        var qlErr *QuantLibError
        if errors.As(err, &qlErr) && !qlErr.Retryable {
            return err
        }
        lastErr = err
    }
    return fmt.Errorf("QuantLib API failed after %d retries: %w", c.retryCount, lastErr)
}
```

#### 4. Fallback Strategy

When the QuantLib API is unavailable, we fall back to existing local calculations:

| Function | Primary | Fallback |
|----------|---------|----------|
| European option pricing | QuantLib API | Need to implement `localBlackScholes()` using Black-Scholes formula with `decimal.Decimal` (not yet in existing code) |
| VaR calculation | QuantLib API | Existing `CalculateVaR()` in `services/risk_models.go` |
| Yield curve | QuantLib API | Error with clear message (no local fallback) |
| Reference data | QuantLib API cache | Stale cache (allow expired entries on API failure) |

```go
func (c *QuantLibClient) PriceEuropeanOptionWithFallback(ctx context.Context, req EuropeanOptionRequest) (*OptionResult, error) {
    result, err := c.PriceEuropeanOption(ctx, req)
    if err != nil {
        utils.Warn("QuantLib API unavailable, falling back to local Black-Scholes", err)
        return c.localBlackScholes(req)  // ⚠️ Requires implementing localBlackScholes() — Black-Scholes formula with decimal.Decimal
    }
    return result, nil
}
```

#### 5. Go Backend - API Handlers

New file: `handlers/quantlib_handler.go`

```
POST /api/quantlib/options/european     - European option pricing
POST /api/quantlib/options/american     - American option pricing
POST /api/quantlib/options/greeks       - Greeks calculation
POST /api/quantlib/yield-curve/build    - Yield curve construction
POST /api/quantlib/bonds/price          - Bond pricing
POST /api/quantlib/risk/var             - VaR calculation
GET  /api/quantlib/types/currencies     - Supported currencies
GET  /api/quantlib/types/frequencies    - Payment frequencies
GET  /api/quantlib/calendars            - Available calendars
GET  /api/quantlib/daycount/conventions - Day count conventions
GET  /api/quantlib/adjustment/methods   - Adjustment methods
GET  /api/quantlib/health               - API connectivity check
```

Handler error handling follows ETF-Insight convention:
```go
func (h *QuantLibHandler) PriceEuropeanOption(c *gin.Context) {
    var req models.EuropeanOptionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error("Invalid request", err)
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid input parameters"})
        return
    }

    result, err := h.client.PriceEuropeanOptionWithFallback(c.Request.Context(), req)
    if err != nil {
        utils.Error("QuantLib API error", err)
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "success": false,
            "error":   "Quantitative analysis service temporarily unavailable",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
```

#### 6. Frontend - QuantLib Analysis Page

New file: `frontend/src/pages/QuantLibAnalysis.tsx`

Components:
- **OptionPricer**: Input spot/strike/rate/vol/expiry, get price + Greeks. Display Greeks as radar chart.
- **YieldCurveBuilder**: Visualize yield curves with ECharts line chart. Support curve type selection.
- **BondCalculator**: Bond pricing with duration/convexity display.
- **VaRCalculator**: Portfolio VaR with QuantLib engine. Compare with existing local VaR results.

Integration with existing pages:
- **PortfolioOptimization.tsx**: Add "QuantLib VaR" tab alongside existing risk metrics
- **RiskAnalysis.tsx**: Add "QuantLib Engine" option in VaR method selector
- **FactorAnalysis.tsx**: Add yield curve reference data display

New TypeScript types in `frontend/src/types/quantlib.ts`:
```typescript
export interface EuropeanOptionRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
}

export interface OptionResult {
  price: string;
  delta: string;
  gamma: string;
  theta: string;
  vega: string;
  rho: string;
}

export interface YieldCurvePoint {
  tenor: string;
  rate: string;
}

export interface BondResult {
  price: string;
  duration: string;
  convexity: string;
  yield_to_maturity: string;
}
```

New API service in `frontend/src/services/quantlibService.ts`:
```typescript
export const quantlibAPI = {
  priceEuropeanOption: (req: EuropeanOptionRequest) =>
    api.post<ApiResponse<OptionResult>>('/api/quantlib/options/european', req),
  priceAmericanOption: (req: AmericanOptionRequest) =>
    api.post<ApiResponse<OptionResult>>('/api/quantlib/options/american', req),
  buildYieldCurve: (req: YieldCurveRequest) =>
    api.post<ApiResponse<YieldCurveResult>>('/api/quantlib/yield-curve/build', req),
  priceBond: (req: BondRequest) =>
    api.post<ApiResponse<BondResult>>('/api/quantlib/bonds/price', req),
  calculateVaR: (req: QuantLibVaRRequest) =>
    api.post<ApiResponse<VaRResult>>('/api/quantlib/risk/var', req),
  getCurrencies: () =>
    api.get<ApiResponse<Currency[]>>('/api/quantlib/types/currencies'),
  getCalendars: () =>
    api.get<ApiResponse<Calendar[]>>('/api/quantlib/calendars'),
  healthCheck: () =>
    api.get<ApiResponse<{ status: string }>>('/api/quantlib/health'),
};
```

#### 7. Caching Strategy

| Data Type | TTL | Cache Key | Invalidation |
|-----------|-----|-----------|-------------|
| Reference data (currencies, calendars) | 1 hour | `quantlib:ref:{endpoint}` | TTL expiry |
| Option pricing results | 5 minutes | `quantlib:option:{hash(params)}` | TTL expiry |
| Yield curve data | 15 minutes | `quantlib:yieldcurve:{hash(params)}` | TTL expiry |
| Bond pricing results | 15 minutes | `quantlib:bond:{hash(params)}` | TTL expiry |

Stale cache fallback: On API failure, return expired cache entries with a warning header:
```go
c.Header("X-Cache-Status", "STALE")
c.Header("X-Cache-Age", fmt.Sprintf("%ds", ageSeconds))
```

#### 8. Configuration

New section in `config/config.go`:
```go
type QuantLibConfig struct {
    APIURL         string        `yaml:"api_url"`
    APIKey         string        `yaml:"api_key"`
    RequestTimeout time.Duration `yaml:"request_timeout"`
    RetryCount     int           `yaml:"retry_count"`
    EnableFallback bool          `yaml:"enable_fallback"`
}
```

**Note**: Struct tags do NOT auto-load env vars. Add to `DefaultConfig()` in `config.go`:
```go
QuantLib: QuantLibConfig{
    APIURL:         getEnv("QUANTLIB_API_URL", "https://api.fincept.in/quantlib"),
    APIKey:         getEnv("QUANTLIB_API_KEY", ""),
    RequestTimeout: getEnvAsDuration("QUANTLIB_REQUEST_TIMEOUT", 10*time.Second),
    RetryCount:     getEnvAsInt("QUANTLIB_RETRY_COUNT", 3),
    EnableFallback: getEnvAsBool("QUANTLIB_ENABLE_FALLBACK", true),
},
```
(Note: `getEnvAsDuration` may need to be implemented — current helpers only have getEnv, getEnvAsInt, getEnvAsBool)

Environment variables:
```bash
QUANTLIB_API_URL=https://api.fincept.in/quantlib
QUANTLIB_API_KEY=your_api_key_here
QUANTLIB_REQUEST_TIMEOUT=10s
QUANTLIB_RETRY_COUNT=3
QUANTLIB_ENABLE_FALLBACK=true
```

#### 9. Database

No new database tables needed for Phase 1. All QuantLib data is either:
- Cached in memory using existing `InMemoryCacheService` (sync.Map-based, in `services/engine/cache_service.go`) with TTL
- Computed on-the-fly from API responses

If we later need to persist QuantLib results, add a `quantlib_results` table in Phase 2.

#### 10. Monitoring and Observability

| Metric | Type | Description |
|--------|------|-------------|
| `quantlib_api_request_total` | Counter | Total API requests by endpoint |
| `quantlib_api_request_duration_seconds` | Histogram | Request latency by endpoint |
| `quantlib_api_error_total` | Counter | Errors by type (auth, timeout, server) |
| `quantlib_cache_hit_total` | Counter | Cache hits by endpoint |
| `quantlib_cache_miss_total` | Counter | Cache misses by endpoint |
| `quantlib_fallback_total` | Counter | Fallback to local calculation |

Logging:
```go
utils.Info("QuantLib API request", "endpoint", endpoint, "duration_ms", elapsed)
utils.Warn("QuantLib API fallback triggered", "endpoint", endpoint, "error", err)
utils.Error("QuantLib API authentication failed", nil)
```

#### 11. Security

| Concern | Mitigation |
|---------|-----------|
| API Key exposure | Store in environment variable, never log, redact from error messages |
| Input validation | Custom validation in handler layer using decimal.Decimal comparison functions (binding tags' gt=0 doesn't support decimal.Decimal struct type) |
| CORS | Existing CORS middleware applies, no special config needed |
| Rate limiting | Existing rate limiter applies (300 req/min per IP — configured in `middleware/security.go`) |
| HTTPS | API uses HTTPS, enforced in client |
| Error message leakage | Never expose internal errors to client, log server-side only |

#### 12. Rollback Strategy

Phase 1 is additive — no existing functionality is modified. Rollback steps:

1. Remove QuantLib routes from `router/router.go`
2. Remove `QUANTLIB_*` environment variables
3. Remove `frontend/src/pages/QuantLibAnalysis.tsx` and route entry
4. All existing features continue to work without QuantLib

Feature flag support:
```go
if cfg.QuantLibConfig.APIKey == "" {
    utils.Warn("QuantLib integration disabled: no API key configured")
    return
}
```

### Testing

| Test Type | Target | Coverage Goal |
|-----------|--------|---------------|
| Unit tests - QuantLibClient | Mocked HTTP responses, error handling, retry logic | > 90% |
| Unit tests - Data models | Validation, decimal precision, edge cases | > 90% |
| Unit tests - Handlers | Request parsing, error responses | > 80% |
| Integration tests - Live API | Skipped in CI, run manually with API key | Key scenarios |
| Frontend tests - OptionPricer | Component rendering, form validation | Key interactions |

Key test cases:
```go
func TestPriceEuropeanOption_Success(t *testing.T) {}
func TestPriceEuropeanOption_APIError(t *testing.T) {}
func TestPriceEuropeanOption_Timeout(t *testing.T) {}
func TestPriceEuropeanOption_RetryOn429(t *testing.T) {}
func TestPriceEuropeanOption_FallbackToLocal(t *testing.T) {}
func TestPriceEuropeanOption_DecimalPrecision(t *testing.T) {}
func TestPriceEuropeanOption_InvalidInput(t *testing.T) {}
func TestGetCurrencies_CacheHit(t *testing.T) {}
func TestGetCurrencies_StaleCacheOnFailure(t *testing.T) {}
```

---

## Phase 2: AI Agent Microservice

### Status: ✅ Completed

### Goal

Build a standalone FastAPI microservice (port 8091) providing AI agents for financial analysis, with Go backend integration and React frontend.

### License Resolution

**Decision**: Option B — Rewrite from scratch (零 AGPL 风险)

Agent 框架完全自主实现，仅参考 FinceptTerminal 的架构设计，不使用任何 AGPL 代码。

### Implementation Summary

- ✅ 4 个 Agent 已实现并测试通过 (Buffett, Graham, Bridgewater, Macro)
- ✅ FastAPI 服务运行在 port 8091
- ✅ 支持 OpenAI, DeepSeek, Ollama 三种 LLM Provider
- ✅ 19 个单元测试全部通过
- ✅ Go Backend 集成完成 (agent_client.go, agent_handler.go)
- ✅ Frontend 集成完成 (AIAgents.tsx, agent.ts, agentAPI)

### Architecture

```
React Frontend (/ai-agents)
    │
Go Backend (/api/agents/*)
    │
Python FastAPI Service (port 8091)
├── GET  /health           - 健康检查
├── GET  /agents/discover   - Agent 列表
├── POST /agents/run        - 单 Agent 执行
├── POST /agents/stream     - SSE 流式响应
└── POST /agents/team       - 多 Agent 团队辩论
```

### Implemented Agent Inventory (Phase 2 初始版本)

| Category | Count | Agents |
|----------|-------|--------|
| Legendary Investors | 2 | Warren Buffett, Benjamin Graham |
| Hedge Funds | 1 | Bridgewater Associates |
| Macro Economic | 1 | Macroeconomic Analyst |
| **Total** | **4** | (框架支持无限扩展) |

> 后续可通过 `agents/registry.py` 轻松添加更多 Agent（地缘政治、技术分析等）

### Key Components (from scratch)

| Component | File | Description |
|-----------|------|-------------|
| LLM Provider | `core/llm_provider.py` | 多模型抽象层 (OpenAI/Ollama/DeepSeek) |
| Base Agent | `core/base_agent.py` | Agent 抽象基类 |
| Tool Registry | `core/tool_registry.py` | 工具注册与调用机制 |
| Agent Manager | `core/agent_manager.py` | Agent 注册/发现/执行/团队协作 |
| FastAPI Server | `agent_server.py` | HTTP API 服务 (port 8091) |
| Pydantic Schemas | `models/schemas.py` | 请求/响应模型 (8 个类型) |
| Tests | `tests/` | 19 个单元测试全部通过 |
| Dockerfile | `Dockerfile` | 容器化部署 |

### Multi-LLM Support

| Provider | Base URL | Models |
|----------|----------|--------|
| OpenAI | `https://api.openai.com/v1` | gpt-4o, gpt-4o-mini, gpt-4-turbo |
| DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat, deepseek-reasoner |
| Ollama | `http://localhost:11434` | llama3.1, qwen2.5, deepseek-r1 |

通过 `get_provider(name)` 工厂函数切换，支持 OpenAI-compatible API 的任意提供商。

### Go Backend Integration

| Component | File |
|-----------|------|
| Client | `services/agent/agent_client.go` |
| Handler | `handlers/agent_handler.go` |
| Routes | `router/router.go` → `registerAgentRoutes()` |

### Frontend Integration

| Component | File |
|-----------|------|
| Types | `frontend/src/types/agent.ts` |
| API Service | `frontend/src/services/api.ts` → `agentAPI` |
| Page | `frontend/src/pages/AIAgents.tsx` |
| Route | `/ai-agents` |
| Sidebar | "AI Agent" (RobotOutlined icon) |

---

## Phase 3: Data Source Microservice

### Goal

直接引入 FinceptTerminal 的数据源脚本，构建统一数据服务微服务 (port 8092)。

### License Approach

**Decision**: 直接引入 FinceptTerminal 代码（不商业化，AGPL 合规可接受）

数据源脚本直接从 FinceptTerminal 仓库提取并适配，不重写。

### Actual Script Inventory (verified from source)

| Source | Actual Script | API Key | Status |
|--------|--------------|---------|--------|
| FRED | `fred_data.py` | `FRED_API_KEY` (required) | ✅ 可直接用 |
| World Bank | `worldbank_data.py` | None | ✅ 可直接用 |
| IMF | `imf_data.py` | None | ✅ 可直接用 |
| Yahoo Finance | `yfinance_data.py` | None | ✅ 可直接用 |
| AkShare | `akshare_*.py` (~20 modules) | None | ✅ 可直接用 |
| CoinGecko | `coingecko.py` | `COINGECKO_API_KEY` (optional) | ✅ 可直接用 |
| SEC EDGAR | ❌ 不存在 | — | 需从零实现 |
| DBnomics | ❌ 不存在 | — | 需从零实现 |
| Databento | ❌ 不存在 | — | 需从零实现 |

> 注: `world_bank_data.py` 实际名 `worldbank_data.py`, `coingecko_data.py` 实际名 `coingecko.py`

### Architecture

```
React Frontend (现有页面增强)
    │
Go Backend (/api/data/*)
    │
Python FastAPI Service (port 8092)
├── /api/fred/*           - FRED 经济数据
├── /api/worldbank/*      - World Bank 全球数据
├── /api/imf/*            - IMF 国际金融
├── /api/yfinance/*       - Yahoo Finance 市场数据
├── /api/akshare/*        - AkShare 中国市场
├── /api/coingecko/*      - CoinGecko 加密货币
└── /health               - 健康检查
```

### Integration with ETF-Insight

- FRED/IMF/World Bank → 增强 `services/exchange_rate/` 宏观指标
- Yahoo Finance → Finage 的备用数据源
- AkShare → 扩展现有 `services/ashare/` 集成
- CoinGecko → 新增加密货币分析能力

---

## Phase 4: Analytics Microservice

### Goal

Extract financial analytics modules.

### Actual Module Inventory (verified from source)

| Category | Modules | Key Capabilities |
|----------|---------|-----------------|
| Equity Investment | 9 | DCF, DDM, multiples, residual income, fundamental analysis |
| Portfolio Management | 11 | Optimization, risk management, ETF analytics, behavioral finance |
| Derivatives | 7 | Options, forwards, arbitrage |
| Economics | 11 | Growth, policy, currency, trade, capital flows |
| Financial Analysis | 11 | Balance sheet, income, cash flow, quality, tax |
| Alternative Investments | 10 | Real estate, hedge funds, private capital, crypto |
| Quantitative Methods | 4 | CFA quant models, rate calculations |
| ML for Trading | 3 | ML trading strategies |
| Technical Analysis | 2 | Momentum, chart patterns |
| Backtesting | 4 frameworks | LEAN, VectorBT, Backtrading.py, FastTrade |

### Integration with ETF-Insight

- Portfolio Management → Enhance existing `services/optimization/` with PyPortfolioOpt/RiskFolioLib
- Derivatives → Complement Phase 1 QuantLib options pricing
- Backtesting → Alternative to existing `services/backtest/` engine
- Economics → New macro analysis capabilities

### License Blocker

Same AGPL-3.0 concern as Phase 2.

---

## License Analysis

FinceptTerminal uses **AGPL-3.0 + Commercial dual license**.

### Phase 1: No License Risk

QuantLib API is a public cloud service. We are consuming it via HTTP, not using or distributing FinceptTerminal's code. This is clean API consumption with no AGPL obligations.

### Phase 2-4: AGPL Risk

AGPL-3.0 Section 13: "if you modify the Program, your modified version must prominently offer all users interacting with it remotely through a computer network [...] an opportunity to receive the Corresponding Source Code."

Extracting and running FinceptTerminal's Python scripts as a network service **triggers AGPL copyleft**. This means:

1. Our entire ETF-Insight backend would need to be made available under AGPL-3.0
2. This conflicts with ETF-Insight's current MIT license
3. We cannot proceed with Phase 2-4 without resolving this

### Resolution Options

| Option | Pros | Cons | Cost |
|--------|------|------|------|
| **A. Commercial License** | Clean legal standing, no copyleft | Recurring cost, dependency on Fincept Corp | Contact `support@fincept.in` |
| **B. Rewrite from scratch** | No license issues, full control | Significant development effort | 2-3 months per phase |
| **C. AGPL-compliant isolation** | Can use original code | Must open-source the microservice, clear boundary needed | Legal review required |
| **D. API-only approach** | No code extraction, no AGPL | Limited to what FinceptTerminal exposes as API | Depends on their API roadmap |

**Recommendation**: Start Phase 1 immediately (no license risk). For Phase 2-4, pursue Option A (commercial license) first. If cost is prohibitive, fall back to Option B (rewrite) for the most valuable modules only.

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| QuantLib API unavailable | Medium | High | Fallback to local calculations, stale cache |
| QuantLib API changes without notice | Low | High | Pin API version in URL, monitor for breaking changes |
| API key revoked or rate limited | Low | Medium | Implement rate limiting on our side, monitor usage |
| AGPL license blocks Phase 2-4 | High | High | Resolve before Phase 2, consider rewrite |
| Float precision loss in financial calculations | Medium | High | Use decimal.Decimal everywhere, string serialization |
| HKEX data format incompatibility | Low | Medium | Validate with HKEX tick sizes, add format adapters |
| Performance degradation from API latency | Medium | Medium | Connection pooling, caching, async requests |

---

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| QuantLib API response time | < 500ms (P95) | Prometheus histogram |
| Options pricing accuracy | Match Black-Scholes within 0.01 | Unit test comparison |
| API availability | > 99.5% | Health check monitoring |
| Fallback activation rate | < 5% of requests | Counter metric |
| Frontend load time | < 2s | Lighthouse score |
| Test coverage (new code) | > 80% | Go test -cover |
| Zero precision loss | 0 incidents | Decimal comparison tests |

---

## References

- [FinceptTerminal GitHub](https://github.com/Fincept-Corporation/FinceptTerminal)
- [QuantLibClient.cpp](https://github.com/Fincept-Corporation/FinceptTerminal/blob/main/fincept-qt/src/services/quantlib/QuantLibClient.cpp) — Verified API format and authentication
- [Python Scripts Library](https://github.com/Fincept-Corporation/FinceptTerminal/tree/main/fincept-qt/scripts) — 250+ scripts, 60+ data sources
- [Agents README](https://github.com/Fincept-Corporation/FinceptTerminal/tree/main/fincept-qt/scripts/agents) — 30+ agents inventory
- [Analytics README](https://github.com/Fincept-Corporation/FinceptTerminal/tree/main/fincept-qt/scripts/Analytics) — 80+ analytics modules

---

## Progress Tracking

### Phase 1: QuantLib Direct Connection

**Status**: ✅ Complete (2026-05-04)
**Commit**: `44fd3bb feat(quantlib): add QuantLib integration module`

| Component | File | Status |
|-----------|------|--------|
| Models | `backend/models/quantlib.go` | ✅ 13 types (6 request, 4 result, 2 enum, 1 wrapper) |
| Validator | `backend/models/quantlib_validator.go` | ✅ 6 validation functions |
| Client | `backend/services/quantlib/quantlib_client.go` | ✅ 10 methods, cache, cleanup |
| Tests | `backend/services/quantlib/quantlib_client_test.go` | ✅ 9 tests, httptest mock |
| Handler | `backend/handlers/quantlib_handler.go` | ✅ 7 handlers, validation, error masking |
| Router | `backend/router/router.go` | ✅ 7 endpoints registered |
| Frontend Types | `frontend/src/types/quantlib.ts` | ✅ 11 interfaces |
| Frontend API | `frontend/src/services/api.ts` | ✅ `quantlibAPI` module (7 methods) |
| Frontend Page | `frontend/src/pages/QuantLibAnalysis.tsx` | ✅ 4 tabs (options/bonds/yield curve/VaR) |
| Frontend Route | `frontend/src/App.tsx`, `Layout.tsx` | ✅ `/quantlib` route + sidebar nav |

**API Endpoints**:
```
POST /api/quantlib/options/european    — 欧式期权定价
POST /api/quantlib/options/american    — 美式期权定价
POST /api/quantlib/options/greeks      — Greeks 计算
POST /api/quantlib/yield-curve/build   — 收益率曲线构建
POST /api/quantlib/bonds/price         — 债券定价
POST /api/quantlib/risk/var            — VaR 计算
GET  /api/quantlib/reference/:type     — 参考数据 (currencies/frequencies/calendars/daycount)
```

**Code Review**: ✅ 通过 (2 轮审查，10 个问题已修复)

### Phase 2: AI Agent Microservice

**Status**: ✅ Complete (2026-05-05)
**License**: Option B — 从零重写，零 AGPL 风险

| Component | File | Status |
|-----------|------|--------|
| Schemas | `backend/services/agent/models/schemas.py` | ✅ 8 Pydantic v2 类型 |
| LLM Provider | `backend/services/agent/core/llm_provider.py` | ✅ OpenAI/Ollama/DeepSeek |
| Base Agent | `backend/services/agent/core/base_agent.py` | ✅ 抽象基类 + run() |
| Tool Registry | `backend/services/agent/core/tool_registry.py` | ✅ @decorator 注册 |
| Agent Manager | `backend/services/agent/core/agent_manager.py` | ✅ 注册/发现/执行/团队 |
| Buffett Agent | `backend/services/agent/agents/buffett.py` | ✅ 价值投资大师 |
| Graham Agent | `backend/services/agent/agents/graham.py` | ✅ 防御型投资 |
| Bridgewater Agent | `backend/services/agent/agents/bridgewater.py` | ✅ 宏观风险平价 |
| Macro Agent | `backend/services/agent/agents/macro.py` | ✅ 宏观经济分析 |
| Agent Registry | `backend/services/agent/agents/registry.py` | ✅ 自动注册 |
| FastAPI Server | `backend/services/agent/agent_server.py` | ✅ 6 endpoints |
| Tests | `backend/services/agent/tests/` | ✅ 19 tests passing |
| Dockerfile | `backend/services/agent/Dockerfile` | ✅ Python 3.11-slim |
| Go Client | `backend/services/agent/agent_client.go` | ✅ 4 methods |
| Go Handler | `backend/handlers/agent_handler.go` | ✅ 4 handlers |
| Go Router | `backend/router/router.go` | ✅ 4 endpoints |
| Frontend Types | `frontend/src/types/agent.ts` | ✅ 6 interfaces |
| Frontend API | `frontend/src/services/api.ts` | ✅ `agentAPI` (4 methods) |
| Frontend Page | `frontend/src/pages/AIAgents.tsx` | ✅ 单Agent/团队模式 |
| Frontend Route | `frontend/src/App.tsx`, `Layout.tsx` | ✅ `/ai-agents` + 侧边栏 |

**API Endpoints**:
```
GET  /api/agents/health    — 健康检查 (Agent数量/LLM提供商)
GET  /api/agents/discover  — Agent列表 (id/name/category/description)
POST /api/agents/run       — 单Agent执行 (支持4个LLM提供商)
POST /api/agents/team      — 多Agent团队辩论 (2-5个Agent, 1-3轮)
```

**Commits**:
```
c657028 feat(agent-service): scaffold Python agent service project structure
e82a4a8 fix(agent-service): improve schema type safety and add docstring
de8992c feat(agent-service): add LLM provider abstraction with OpenAI/Ollama support
3fefd27 feat(agent-service): add base agent class and tool registry
32955df feat(agent-service): add agent manager with discovery and team execution
eccaf0f fix(agent-service): truncate system prompt preview to exactly 200 chars
2d14804 feat(agent-service): add FastAPI server with discover/run/stream/team endpoints
1f70917 feat(agent-service): add 4 financial agents (Buffett, Graham, Bridgewater, Macro)
84fa2e3 feat(agent): add Go client, handler, and routes for agent service
5442592 feat(agent): add AI Agents page with single and team analysis modes
```

### Phase 3: Data Source Microservice

**Status**: ✅ Complete (2026-05-06)
**Plan**: `docs/superpowers/plans/2026-05-05-phase3-data-service.md`
**License**: 直接引入 FinceptTerminal 代码 (AGPL, 非商业)
**Approach**: One-clone 提取 16 个脚本 + FastAPI 包装 (port 8092)

| Component | File | Status |
|-----------|------|--------|
| Scripts (16 files) | `backend/services/data/sources/` | ✅ 已提取 |
| FastAPI Server | `backend/services/data/data_server.py` | ✅ 已实现 |
| Routers (6 modules) | `backend/services/data/routers/` | ✅ 已实现 |
| Tests | `backend/services/data/tests/` | ✅ 已编写 (3 tests) |
| Dockerfile | `backend/services/data/Dockerfile` | ✅ 已创建 |
| AGPL Notice | `backend/services/data/NOTICE` | ✅ 已创建 |
| Go Client | `backend/services/data/data_client.go` | ✅ 已实现 |
| Go Handler | `backend/handlers/data_handler.go` | ✅ 已实现 |
| Go Routes | `backend/router/router.go` | ✅ 已注册 |

**Data Sources**: FRED / World Bank / IMF / Yahoo Finance / AkShare (11 modules) / CoinGecko
**Skipped**: SEC EDGAR / DBnomics / Databento (FinceptTerminal 中不存在)

### Phase 4: Analytics Microservice

**Status**: ✅ Complete (2026-05-06)
**Plan**: `docs/superpowers/plans/2026-05-06-phase4-analytics-service.md`
**License**: 直接引入 FinceptTerminal 代码 (AGPL, 非商业)
**Approach**: One-clone 提取 16 个 Portfolio Management 模块 + FastAPI 包装 (port 8093)

| Component | File | Status |
|-----------|------|--------|
| Modules (16 files) | `backend/services/analytics/modules/` | ✅ 已提取 |
| FastAPI Server | `backend/services/analytics/analytics_server.py` | ✅ 已实现 |
| Routers (4 modules) | `backend/services/analytics/routers/` | ✅ 已实现 |
| Tests | `backend/services/analytics/tests/` | ✅ 已编写 (4 tests) |
| Dockerfile | `backend/services/analytics/Dockerfile` | ✅ 已创建 |
| AGPL Notice | `backend/services/analytics/NOTICE` | ✅ 已创建 |
| Go Client | `backend/services/analytics/analytics_client.go` | ✅ 已实现 |
| Go Handler | `backend/handlers/analytics_handler.go` | ✅ 已实现 |
| Go Routes | `backend/router/router.go` | ✅ 已注册 |

**Analytics Modules**: Portfolio Optimization / Risk Management / Portfolio Analytics / Portfolio Planning
**TDD Approach**: RED → GREEN → REFACTOR
