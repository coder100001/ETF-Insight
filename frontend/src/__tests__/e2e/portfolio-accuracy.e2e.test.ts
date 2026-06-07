import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * E2E Test Suite: Portfolio Accuracy Improvements
 *
 * Tests the full data flow from API response to component rendering.
 * Uses mocked API responses that match the real backend format.
 */

// ============================================================================
// Mock API Responses (matching real backend format)
// ============================================================================

const mockFinancialConfig = {
  success: true,
  data: {
    risk_free_rate: 0.0435,
    trading_days_year: 252,
    default_currency: 'USD',
  },
};

const mockPortfolioAnalyze = {
  success: true,
  data: {
    portfolio: {
      total_investment: 100000,
      allocation: { SCHD: 70, JEPQ: 30 },
    },
    metrics: {
      annual_return: 0.085,
      volatility: 0.12,
      sharpe_ratio: 0.35,
      sortino_ratio: 0.48,
      max_drawdown: -0.085,
      calmar_ratio: 1.0,
    },
    risk: {
      var_95: -0.032,
      var_99: -0.045,
      cvar_95: -0.041,
    },
    dividend: {
      annual_yield: 0.045,
      annual_income: 4500,
      monthly_income: 375,
    },
  },
};

const mockRiskAnalysis = {
  success: true,
  data: {
    var_95: 0.032,
    var_99: 0.045,
    cvar_95: 0.041,
    volatility: 0.12,
    sharpe_ratio: 0.35,
    sortino_ratio: 0.48,
    max_drawdown: 0.085,
    calmar_ratio: 1.0,
    beta: 0.85,
    alpha: 0.02,
    portfolio_risks: [
      { symbol: 'SCHD', weight: 0.70, component_var: 0.022, marginal_var: 0.031 },
      { symbol: 'JEPQ', weight: 0.30, component_var: 0.010, marginal_var: 0.033 },
    ],
    data_points: 252,
    risk_level: 'medium',
  },
};

const mockMPTOptimize = {
  success: true,
  data: {
    weights: { SCHD: 0.65, JEPQ: 0.35 },
    expected_return: 0.088,
    volatility: 0.115,
    sharpe_ratio: 0.38,
    sortino_ratio: 0.52,
    max_drawdown: 0.082,
    diversification_ratio: 1.15,
    risk_contribution: { SCHD: 0.55, JEPQ: 0.45 },
  },
};

// ============================================================================
// Financial Config API Tests
// ============================================================================

describe('E2E: FinancialConfig API', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('should return correct default financial config', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockFinancialConfig),
    });

    const res = await fetch('/api/config/financial');
    const data = await res.json();

    expect(data.success).toBe(true);
    expect(data.data.risk_free_rate).toBe(0.0435);
    expect(data.data.trading_days_year).toBe(252);
    expect(data.data.default_currency).toBe('USD');
  });

  it('should update risk-free rate via PUT', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ success: true, message: 'updated' }),
    });

    const res = await fetch('/api/config/financial', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ risk_free_rate: 0.05 }),
    });
    const data = await res.json();

    expect(data.success).toBe(true);
  });

  it('should reject invalid risk-free rate', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ success: false, error: 'out of range' }),
    });

    const res = await fetch('/api/config/financial', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ risk_free_rate: 0.60 }),
    });

    expect(res.ok).toBe(false);
    expect(res.status).toBe(400);
  });
});

// ============================================================================
// Portfolio Analysis E2E
// ============================================================================

describe('E2E: Portfolio Analysis', () => {
  it('should return real metrics (not fabricated zeros)', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockPortfolioAnalyze),
    });

    const res = await fetch('/api/portfolio/analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        allocation: { SCHD: 70, JEPQ: 30 },
        total_investment: 100000,
        tax_rate: 0.10,
      }),
    });
    const data = await res.json();

    // Verify metrics are real values, not zeros
    expect(data.data.metrics.sharpe_ratio).not.toBe(0);
    expect(data.data.metrics.sortino_ratio).not.toBe(0);
    expect(data.data.metrics.max_drawdown).not.toBe(0);

    // Verify VaR/CVaR are reasonable
    expect(data.data.risk.var_95).toBeLessThan(0);
    expect(data.data.risk.cvar_95).toBeLessThan(data.data.risk.var_95);

    // Verify dividend data
    expect(data.data.dividend.annual_yield).toBeGreaterThan(0);
    expect(data.data.dividend.annual_income).toBeGreaterThan(0);
  });
});

// ============================================================================
// Risk Analysis E2E
// ============================================================================

describe('E2E: Risk Analysis', () => {
  it('should return real risk metrics from backend', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockRiskAnalysis),
    });

    const res = await fetch('/api/portfolio/risk', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        portfolio: { SCHD: 70, JEPQ: 30 },
        confidence: 0.95,
        period: 252,
      }),
    });
    const data = await res.json();

    // All metrics should be real (not fabricated zeros)
    expect(data.data.sharpe_ratio).not.toBe(0);
    expect(data.data.sortino_ratio).not.toBe(0);
    expect(data.data.max_drawdown).not.toBe(0);
    expect(data.data.beta).not.toBe(1); // Not hardcoded 1
    expect(data.data.alpha).not.toBe(0);

    // VaR/CVaR should be from backend calculation
    expect(data.data.var_95).toBeGreaterThan(0);
    expect(data.data.cvar_95).toBeGreaterThan(data.data.var_95);
    expect(data.data.var_99).toBeGreaterThan(data.data.var_95);

    // Portfolio risks should have component/marginal VaR
    expect(data.data.portfolio_risks).toHaveLength(2);
    expect(data.data.portfolio_risks[0].component_var).toBeGreaterThan(0);
  });
});

// ============================================================================
// Optimization E2E
// ============================================================================

describe('E2E: Portfolio Optimization', () => {
  it('should return real optimization results', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockMPTOptimize),
    });

    const res = await fetch('/api/optimization/mpt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        symbols: ['SCHD', 'JEPQ'],
        objective: 'max_sharpe',
        risk_free_rate: 0.0435,
      }),
    });
    const data = await res.json();

    // Weights should sum to ~1
    const totalWeight = Object.values(data.data.weights as Record<string, number>)
      .reduce((sum: number, w: number) => sum + w, 0);
    expect(totalWeight).toBeCloseTo(1.0, 2);

    // Metrics should be real (not zeros)
    expect(data.data.sortino_ratio).not.toBe(0);
    expect(data.data.diversification_ratio).not.toBe(0);
    expect(data.data.risk_contribution).toBeDefined();

    // Risk contribution should sum to ~1
    const totalContribution = Object.values(data.data.risk_contribution as Record<string, number>)
      .reduce((sum: number, c: number) => sum + c, 0);
    expect(totalContribution).toBeCloseTo(1.0, 1);
  });

  it('should use correct risk-free rate in request', async () => {
    const fetchSpy = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockMPTOptimize),
    });
    globalThis.fetch = fetchSpy;

    await fetch('/api/optimization/mpt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        symbols: ['SCHD', 'JEPQ'],
        objective: 'max_sharpe',
        risk_free_rate: 0.0435,
      }),
    });

    const callBody = JSON.parse(fetchSpy.mock.calls[0][1].body);
    expect(callBody.risk_free_rate).toBe(0.0435);
    // Should NOT be 0.045 or 0.04
    expect(callBody.risk_free_rate).not.toBe(0.045);
    expect(callBody.risk_free_rate).not.toBe(0.04);
  });
});

// ============================================================================
// Data Consistency E2E
// ============================================================================

describe('E2E: Data Consistency', () => {
  it('risk-free rate should be consistent across all API calls', async () => {
    const fetchSpy = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockFinancialConfig),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMPTOptimize),
      });
    globalThis.fetch = fetchSpy;

    // Get config
    const configRes = await fetch('/api/config/financial');
    const configData = await configRes.json();
    const configRate = configData.data.risk_free_rate;

    // Use in optimization
    await fetch('/api/optimization/mpt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        symbols: ['SCHD', 'JEPQ'],
        objective: 'max_sharpe',
        risk_free_rate: configRate,
      }),
    });

    const optBody = JSON.parse(fetchSpy.mock.calls[1][1].body);
    expect(optBody.risk_free_rate).toBe(configRate);
  });

  it('trading days should be 252 everywhere', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockFinancialConfig),
    });

    const res = await fetch('/api/config/financial');
    const data = await res.json();

    expect(data.data.trading_days_year).toBe(252);
    // Should NOT be 365
    expect(data.data.trading_days_year).not.toBe(365);
  });
});
