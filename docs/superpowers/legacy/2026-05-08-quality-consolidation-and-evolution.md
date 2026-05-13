# Quality Consolidation & Evolution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate the quality of the existing codebase before adding new features — fix dead code, wire up unconnected services, add frontend test coverage, and clean up architecture debt.

**Architecture:** Three-phase approach: (1) immediate cleanup of dead code and broken endpoints, (2) frontend test coverage for core pages, (3) architecture convergence for the engine/ package and data source layer. Each phase produces independently shippable, testable software.

**Tech Stack:** Go 1.26, Gin, GORM, React 19, TypeScript, Vite, Ant Design, Vitest, pytest

---

## Phase 1: Dead Code Cleanup & Endpoint Wiring

### Task 1: Wire up risk-budget CVaR and Monte Carlo endpoints

The `RiskBudgetService` has `CalculateCVaR`, `CalculateMonteCarloCVaR`, and `RunMonteCarloSimulation` methods, but no handler exposes them. The `OptimizationHandler` already holds a `riskBudgetService` field.

**Files:**
- Modify: `backend/handlers/optimization_handler.go` (add 3 handler methods)
- Modify: `backend/router/router.go:196-202` (register new routes)

- [ ] **Step 1: Add CalculateCVaR handler method to optimization_handler.go**

```go
// CalculateCVaRRequest 计算CVaR请求
type CalculateCVaRRequest struct {
	Returns       []float64 `json:"returns" binding:"required,min=10"`
	Confidence    float64   `json:"confidence" binding:"required"`
	UseParametric bool      `json:"use_parametric,omitempty"`
}

// CalculateCVaRResponse 计算CVaR响应
type CalculateCVaRResponse struct {
	Success bool              `json:"success"`
	Data    *CVaRResult       `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// CVaRResult CVaR计算结果
type CVaRResult struct {
	VaR          float64 `json:"var"`
	CVaR         float64 `json:"cvar"`
	Confidence   float64 `json:"confidence"`
	Method       string  `json:"method"`
	SampleSize   int     `json:"sample_size"`
}

// CalculateCVaR 计算VaR和CVaR
func (h *OptimizationHandler) CalculateCVaR(c *gin.Context) {
	var req CalculateCVaRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CalculateCVaRResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	returns := make([]decimal.Decimal, len(req.Returns))
	for i, r := range req.Returns {
		returns[i] = decimal.NewFromFloat(r)
	}

	confidence := decimal.NewFromFloat(req.Confidence)
	varVaR, varCVaR, err := h.riskBudgetService.CalculateCVaR(returns, confidence, req.UseParametric)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CalculateCVaRResponse{
			Success: false,
			Error:   "计算失败: " + err.Error(),
		})
		return
	}

	method := "historical"
	if req.UseParametric {
		method = "parametric"
	}

	varVal, _ := varVaR.Float64()
	cvarVal, _ := varCVaR.Float64()

	c.JSON(http.StatusOK, CalculateCVaRResponse{
		Success: true,
		Data: &CVaRResult{
			VaR:        varVal,
			CVaR:       cvarVal,
			Confidence: req.Confidence,
			Method:     method,
			SampleSize: len(req.Returns),
		},
	})
}
```

- [ ] **Step 2: Add MonteCarlo simulation handler method**

```go
// MonteCarloRequest 蒙特卡洛模拟请求
type MonteCarloRequest struct {
	PortfolioID    uint      `json:"portfolio_id" binding:"required"`
	NumSimulations int       `json:"num_simulations" binding:"required,min=100"`
	TimeSteps      int       `json:"time_steps" binding:"required,min=1"`
	Returns        []float64 `json:"returns" binding:"required,min=10"`
}

// MonteCarloResponse 蒙特卡洛模拟响应
type MonteCarloResponse struct {
	Success bool                        `json:"success"`
	Data    *models.MonteCarloSimulation `json:"data,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

// RunMonteCarlo 运行蒙特卡洛模拟
func (h *OptimizationHandler) RunMonteCarlo(c *gin.Context) {
	var req MonteCarloRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MonteCarloResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	returns := make([]decimal.Decimal, len(req.Returns))
	for i, r := range req.Returns {
		returns[i] = decimal.NewFromFloat(r)
	}

	result, err := h.riskBudgetService.RunMonteCarloSimulation(
		req.PortfolioID, req.NumSimulations, req.TimeSteps, returns,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MonteCarloResponse{
			Success: false,
			Error:   "模拟失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MonteCarloResponse{
		Success: true,
		Data:    result,
	})
}
```

- [ ] **Step 3: Register new routes in router.go**

In `registerRiskBudgetRoutes()`, add after the existing routes:

```go
func (r *Router) registerRiskBudgetRoutes() {
	rb := r.engine.Group("/api/risk-budget")
	{
		rb.GET("/configs", r.handlers.Optimization.GetRiskBudgetConfigs)
		rb.POST("/configs", r.handlers.Optimization.CreateRiskBudgetConfig)
		rb.POST("/calculate-cvar", r.handlers.Optimization.CalculateCVaR)
		rb.POST("/monte-carlo", r.handlers.Optimization.RunMonteCarlo)
	}
}
```

- [ ] **Step 4: Verify Go build passes**

Run: `cd backend && go build -o /dev/null ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add backend/handlers/optimization_handler.go backend/router/router.go
git commit -m "feat(api): wire up risk-budget CVaR and Monte Carlo endpoints"
```

---

### Task 2: Remove dead `engine/` sub-packages

The `backend/services/engine/` directory has 4 empty sub-packages (`etf/`, `factor/`, `optimization/`, `portfolio/`) that were planned but never implemented. They are dead code.

**Files:**
- Delete: `backend/services/engine/etf/`
- Delete: `backend/services/engine/factor/`
- Delete: `backend/services/engine/optimization/`
- Delete: `backend/services/engine/portfolio/`
- Verify: `backend/services/engine/` only contains `cache_service.go`

- [ ] **Step 1: Check what's in each sub-package**

Run: `find backend/services/engine -type f -name "*.go" | sort`
Expected: Only `backend/services/engine/cache_service.go` has content; sub-packages have empty `*.go` files

- [ ] **Step 2: Delete empty sub-packages**

```bash
rm -rf backend/services/engine/etf
rm -rf backend/services/engine/factor
rm -rf backend/services/engine/optimization
rm -rf backend/services/engine/portfolio
```

- [ ] **Step 3: Verify no imports reference deleted packages**

Run: `grep -r "services/engine/etf\|services/engine/factor\|services/engine/optimization\|services/engine/portfolio" backend/ --include="*.go"`
Expected: No matches

- [ ] **Step 4: Verify Go build passes**

Run: `cd backend && go build -o /dev/null ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add -A backend/services/engine/
git commit -m "chore: remove empty engine sub-packages (etf, factor, optimization, portfolio)"
```

---

### Task 3: Fix API path mismatch in frontend `getSignalHistory`

The frontend `getSignalHistory` was calling `/factor/timing/history?factor_name=...` (query param) but the backend route is `/factor/timing/history/:factor_name` (path param). Already fixed in the previous session — verify the fix is correct.

**Files:**
- Verify: `frontend/src/services/api.ts:838-842`

- [ ] **Step 1: Read the current implementation**

Read `frontend/src/services/api.ts` lines 838-842. Confirm it uses path param format:
```typescript
getSignalHistory: async (factorName: string): Promise<ApiResponse<FactorTimingSignal[]>> => {
    return request<ApiResponse<FactorTimingSignal[]>>(
      `/factor/timing/history/${encodeURIComponent(factorName)}`
    );
},
```

- [ ] **Step 2: Verify backend route matches**

Read `backend/router/router.go` line 353. Confirm:
```go
timing.GET("/history/:factor_name", r.handlers.FactorTiming.GetFactorTimingHistory)
```

- [ ] **Step 3: Run TypeScript check**

Run: `cd frontend && npx tsc --noEmit`
Expected: Exit code 0

---

## Phase 2: Frontend Test Coverage

### Task 4: Add test for PortfolioOptimization page

The PortfolioOptimization page is the most complex page (MPT, risk parity, Black-Litterman). It has zero tests.

**Files:**
- Create: `frontend/src/pages/__tests__/PortfolioOptimization.test.tsx`
- Reference: `frontend/src/pages/PortfolioOptimization.tsx`

- [ ] **Step 1: Create test file with basic render test**

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import PortfolioOptimization from '../PortfolioOptimization';

// Mock API calls
vi.mock('../../services/api', () => ({
  optimizationAPI: {
    mptOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
    efficientFrontier: vi.fn().mockResolvedValue({ success: true, data: [] }),
    riskParityOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
    blackLittermanOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
  },
}));

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider>
      <AntdApp>
        <BrowserRouter>{ui}</BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  );
};

describe('PortfolioOptimization', () => {
  it('renders without crashing', () => {
    renderWithProviders(<PortfolioOptimization />);
    expect(screen.getByText(/组合优化|Portfolio Optimization/i)).toBeTruthy();
  });

  it('displays optimization type selector', () => {
    renderWithProviders(<PortfolioOptimization />);
    // Should have tabs or select for MPT, Risk Parity, BL
    expect(screen.getByText(/均值方差|MPT|最大夏普/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/pages/__tests__/PortfolioOptimization.test.tsx`
Expected: 2 tests pass

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/__tests__/PortfolioOptimization.test.tsx
git commit -m "test(frontend): add PortfolioOptimization page tests"
```

---

### Task 5: Add test for FactorTiming page

**Files:**
- Create: `frontend/src/pages/__tests__/FactorTiming.test.tsx`
- Reference: `frontend/src/pages/FactorTiming.tsx`

- [ ] **Step 1: Create test file**

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import FactorTiming from '../FactorTiming';

const mockCalculateSignal = vi.fn().mockResolvedValue({
  success: true,
  data: {
    id: 1,
    factor_name: 'Mkt-RF',
    signal_date: '2026-05-08',
    ma_slope_60: 0.0012,
    z_score: 1.5,
    percentile: 75.0,
    signal_strength: 'weak_positive',
    expected_return: 0.08,
    confidence: 65.0,
    signal_score: 0.6,
  },
});

const mockGetSignalHistory = vi.fn().mockResolvedValue({
  success: true,
  data: [],
});

vi.mock('../../services/api', () => ({
  factorTimingAPI: {
    calculateSignal: (...args: unknown[]) => mockCalculateSignal(...args),
    getSignalHistory: (...args: unknown[]) => mockGetSignalHistory(...args),
    generateView: vi.fn().mockResolvedValue({ success: true, data: null }),
  },
}));

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider>
      <AntdApp>
        <BrowserRouter>{ui}</BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  );
};

describe('FactorTiming', () => {
  it('renders without crashing', () => {
    renderWithProviders(<FactorTiming />);
    expect(screen.getByText(/因子择时/i)).toBeTruthy();
  });

  it('has a calculate button', () => {
    renderWithProviders(<FactorTiming />);
    expect(screen.getByText(/计算信号/i)).toBeTruthy();
  });

  it('calls calculateSignal API on button click', async () => {
    renderWithProviders(<FactorTiming />);
    const button = screen.getByText(/计算信号/i);
    fireEvent.click(button);
    await waitFor(() => {
      expect(mockCalculateSignal).toHaveBeenCalledWith('Mkt-RF', 60);
    });
  });
});
```

- [ ] **Step 2: Run test**

Run: `cd frontend && npx vitest run src/pages/__tests__/FactorTiming.test.tsx`
Expected: 3 tests pass

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/__tests__/FactorTiming.test.tsx
git commit -m "test(frontend): add FactorTiming page tests"
```

---

### Task 6: Add test for BlackLittermanConfig page

**Files:**
- Create: `frontend/src/pages/__tests__/BlackLittermanConfig.test.tsx`
- Reference: `frontend/src/pages/BlackLittermanConfig.tsx`

- [ ] **Step 1: Create test file**

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import BlackLittermanConfig from '../BlackLittermanConfig';

vi.mock('../../services/api', () => ({
  blackLittermanAPI: {
    createConfig: vi.fn().mockResolvedValue({ success: true, data: null }),
    calculate: vi.fn().mockResolvedValue({ success: true, data: null }),
    getPosteriorReturns: vi.fn().mockResolvedValue({ success: true, data: null }),
  },
}));

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider>
      <AntdApp>
        <BrowserRouter>{ui}</BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  );
};

describe('BlackLittermanConfig', () => {
  it('renders without crashing', () => {
    renderWithProviders(<BlackLittermanConfig />);
    expect(screen.getByText(/Black-Litterman|BL模型/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test**

Run: `cd frontend && npx vitest run src/pages/__tests__/BlackLittermanConfig.test.tsx`
Expected: 1 test passes

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/__tests__/BlackLittermanConfig.test.tsx
git commit -m "test(frontend): add BlackLittermanConfig page tests"
```

---

## Phase 3: Architecture Convergence

### Task 7: Mark `plugin_service.go` as experimental

The `PluginService` is fully implemented but has zero routes and zero handler. Rather than wiring it up (which would add untested functionality), mark it as experimental with a clear comment.

**Files:**
- Modify: `backend/services/plugin_service.go` (add experimental notice)

- [ ] **Step 1: Add experimental notice at top of file**

Add at the top of `backend/services/plugin_service.go`, after the package declaration:

```go
// EXPERIMENTAL: This service is fully implemented but not yet wired to any API routes.
// It provides plugin registration, configuration, execution, and benchmarking.
// To activate: create handlers/plugin_handler.go and register routes in router.go.
// See models/plugin.go for the data models.
```

- [ ] **Step 2: Verify Go build passes**

Run: `cd backend && go build -o /dev/null ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add backend/services/plugin_service.go
git commit -m "docs: mark plugin_service.go as experimental with activation instructions"
```

---

### Task 8: Document undocumented API routes

Several routes exist in `router.go` but are not documented in AGENTS.md: report system, data proxy, analytics proxy, portfolio configs CRUD, ETF configs CRUD, cache management.

**Files:**
- Modify: `AGENTS.md` (add missing endpoint documentation)

- [ ] **Step 1: Add report system endpoints to AGENTS.md**

Add after the existing QuantLib section:

```markdown
#### 报表系统 (v2.7+)
- `GET /api/reports/templates` - 获取报表模板列表
- `GET /api/reports/templates/default` - 获取默认模板
- `POST /api/reports/templates` - 创建报表模板
- `POST /api/reports/generate` - 生成报表
- `GET /api/reports/:id` - 获取报表详情
- `GET /api/reports/:id/download` - 下载报表
```

- [ ] **Step 2: Add data proxy endpoints**

```markdown
#### 数据代理 (v2.10+)
- `GET /api/data/health` - 数据服务健康检查
- `GET /api/data/fred/series/:series_id` - FRED 数据查询
- `GET /api/data/yfinance/quote/:symbol` - Yahoo Finance 报价
- `GET /api/data/akshare/stock/spot` - AKShare A股行情
```

- [ ] **Step 3: Add analytics proxy endpoints**

```markdown
#### 分析服务代理 (v2.11+)
- `GET /api/analytics/health` - 分析服务健康检查
- `POST /api/analytics/optimize` - 组合优化（外部引擎）
- `POST /api/analytics/var` - VaR 计算（外部引擎）
- `POST /api/analytics/capm` - CAPM 模型（外部引擎）
```

- [ ] **Step 4: Commit**

```bash
git add AGENTS.md
git commit -m "docs: add missing API endpoint documentation (reports, data proxy, analytics)"
```

---

## Verification Checklist

After all tasks complete, run the full verification suite:

- [ ] `cd backend && go build -o /dev/null ./...` — Go build passes
- [ ] `cd frontend && npx tsc --noEmit` — TypeScript check passes
- [ ] `cd frontend && npx vitest run` — All frontend tests pass
- [ ] `cd backend && go test ./...` — All backend tests pass
- [ ] `git log --oneline -10` — Clean commit history
- [ ] `grep -r "TODO\|FIXME\|HACK" backend/services/engine/` — No leftover TODOs in cleaned area

---

## NOT in Scope (Future Phases)

These items were identified in the evolution analysis but are explicitly excluded from this plan:

| Item | Reason | Future Phase |
|------|--------|-------------|
| WebSocket real-time推送 | New feature, not quality consolidation | Phase 4 |
| 策略实验台 UI | New feature | Phase 4 |
| 插件生态 handler + routes | Requires design decisions | Phase 3 |
| Go 版本降级/升级 | Environment issue, not code issue | Ops |
| 前端页面全覆盖测试 | Too large for one plan | Ongoing |
| 数据源统一框架 | Architecture refactor, needs separate plan | Phase 3 |
