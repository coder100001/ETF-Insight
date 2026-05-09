// API服务 - 连接Go后端
import type {
  ETFData, ETFConfig, PortfolioAnalysisResult, ExchangeRate, PortfolioConfig, ETFHistoryDataItem,
  FactorTimingSignal, AlphaView, AlphaViewPerformance, ViewMethod,
  CreateAlphaViewRequest, UpdateAlphaViewRequest,
  BlackLittermanConfig, BLPosteriorReturn, CreateBLConfigRequest,
  RiskBudgetConfig, MonteCarloSimulation, CVaRResult, CreateRiskBudgetRequest
} from '../types';
import type {
  EuropeanOptionRequest, AmericanOptionRequest, GreeksRequest, OptionResult,
  YieldCurveRequest, YieldCurveResult, BondRequest, BondResult, VaRRequest, VaRResult
} from '../types/quantlib';
import type {
  AgentInfo, AgentRunRequest, AgentRunResponse,
  AgentTeamRequest, AgentTeamResponse
} from '../types/agent';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

// API响应类型
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

// 请求重试配置
interface RetryConfig {
  maxRetries?: number;
  baseDelay?: number;
  maxDelay?: number;
  timeout?: number;
}

// 请求合并器 - 避免重复请求
class RequestCoalescer {
  private pendingRequests = new Map<string, Promise<unknown>>();

  async getOrSet<T>(key: string, fetcher: () => Promise<T>): Promise<T> {
    if (this.pendingRequests.has(key)) {
      return this.pendingRequests.get(key) as Promise<T>;
    }

    const promise = fetcher().finally(() => {
      this.pendingRequests.delete(key);
    });

    this.pendingRequests.set(key, promise);
    return promise;
  }

  clear(key?: string) {
    if (key) {
      this.pendingRequests.delete(key);
    } else {
      this.pendingRequests.clear();
    }
  }
}

const requestCoalescer = new RequestCoalescer();

// 创建带超时的 fetch
function fetchWithTimeout(
  url: string,
  options: RequestInit,
  timeout: number
): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  return fetch(url, {
    ...options,
    signal: controller.signal,
  }).finally(() => clearTimeout(timeoutId));
}

// 判断是否应该重试
function shouldRetry(statusCode?: number): boolean {
  // 429 Too Many Requests - 重试
  if (statusCode === 429) return true;
  // 5xx 服务器错误 - 重试
  if (statusCode && statusCode >= 500 && statusCode < 600) return true;
  // 网络错误 (statusCode undefined) - 重试
  if (!statusCode) return true;
  // 4xx 客户端错误 - 不重试
  return false;
}

// 重试请求函数
async function requestWithRetry<T>(
  url: string,
  options?: RequestInit,
  config: RetryConfig = {}
): Promise<T> {
  const {
    maxRetries = 3,
    baseDelay = 1000,
    maxDelay = 10000,
    timeout = 30000,
  } = config;

  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetchWithTimeout(`${API_BASE_URL}${url}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      }, timeout);

      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        const err = new Error(error.error || `HTTP ${response.status}`);
        (err as Error & { statusCode: number }).statusCode = response.status;
        throw err;
      }

      return await response.json();
    } catch (error) {
      lastError = error as Error;

      // 用户取消请求，直接抛出不再重试
      if (error instanceof Error && error.name === 'AbortError') {
        throw error;
      }

      const statusCode = (error as Error & { statusCode?: number }).statusCode;

      if (attempt < maxRetries && shouldRetry(statusCode)) {
        const isRateLimit = statusCode === 429;
        // 429错误使用更长的退避时间
        const multiplier = isRateLimit ? 4 : 2;
        const delay = Math.min(baseDelay * Math.pow(multiplier, attempt), maxDelay);
        await new Promise(resolve => setTimeout(resolve, delay));
      } else {
        // 不可重试的错误，直接抛出
        throw error;
      }
    }
  }

  throw new Error(`Request failed after ${maxRetries + 1} attempts: ${lastError?.message}`);
}

// 通用请求函数（带合并和重试）
export async function request<T>(url: string, options?: RequestInit, config?: RetryConfig): Promise<T> {
  const cacheKey = `${url}-${JSON.stringify(options)}`;

  return requestCoalescer.getOrSet(cacheKey, () =>
    requestWithRetry<T>(url, options, config)
  );
}

// ETF相关API
export const etfAPI = {
  getList: (market?: string) => {
    const params = market ? `?market=${market}` : '';
    return request<ApiResponse<ETFData[]>>(`/etf/list${params}`);
  },

  getComparison: (period: string = '1y') => {
    return request<ApiResponse<ETFData[]>>(`/etf/comparison?period=${period}`);
  },

  getPortfolioAnalysis: (allocation: Record<string, number>, totalInvestment: number = 10000, taxRate: number = 0.10) => {
    return request<ApiResponse<PortfolioAnalysisResult>>(`/etf/portfolio`, {
      method: 'POST',
      body: JSON.stringify({ allocation, total_investment: totalInvestment, tax_rate: taxRate }),
    });
  },

  getRealtimeData: (symbol: string) => {
    return request<ApiResponse<ETFData>>(`/etf/${symbol}/realtime`);
  },

  getMetrics: (symbol: string, period: string = '1y') => {
    return request<ApiResponse<Record<string, number>>>(`/etf/${symbol}/metrics?period=${period}`);
  },

  getHistory: (symbol: string, period: string = '1y') => {
    return request<ApiResponse<ETFHistoryDataItem[]>>(`/etf/${symbol}/history?period=${period}`);
  },

  getForecast: (symbol: string, initialInvestment: number = 10000, taxRate: number = 0.10) => {
    return request<ApiResponse<{ years: number; value: number }[]>>(`/etf/${symbol}/forecast?initial_investment=${initialInvestment}&tax_rate=${taxRate}`);
  },

  updateRealtimeData: () => {
    return request<ApiResponse<{ message: string; count: number }>>(`/etf/update-realtime`, {
      method: 'POST',
    });
  },

  getRisk: (symbol: string, period: string = '1y', confidence: number = 0.95) => {
    return request<ApiResponse<{
      symbol: string;
      period: string;
      current_price: number;
      period_high: number;
      period_low: number;
      volatility: number;
      var_95: number;
      var_99: number;
      cvar_95: number;
      max_drawdown: number;
      sharpe_ratio: number;
      beta: number;
    }>>(`/etf/${symbol}/risk?period=${period}&confidence=${confidence}`);
  },
};

// 投资组合相关API
export const portfolioAPI = {
  analyzeScenarios: (allocation: Record<string, number>, scenarios: Array<{ name: string; shock: Record<string, number> }>) => {
    return request<ApiResponse<Record<string, unknown>>>(`/portfolio/scenarios`, {
      method: 'POST',
      body: JSON.stringify({ allocation, scenarios }),
    });
  },

  getDefaultTemplates: () => {
    return request<ApiResponse<Array<{ name: string; description: string; allocation: Record<string, number> }>>>(`/portfolio/default-templates`);
  },

  analyzeRisk: (allocation: Record<string, number>, totalInvestment: number = 10000) => {
    return request<ApiResponse<{
      total_risk: number;
      systematic_risk: number;
      unsystematic_risk: number;
      diversification_ratio: number;
      concentration_risk: string;
    }>>(`/portfolio/risk`, {
      method: 'POST',
      body: JSON.stringify({ allocation, total_investment: totalInvestment }),
    });
  },

  optimize: (allocation: Record<string, number>, objective: string = 'sharpe', constraints?: Record<string, unknown>) => {
    return request<ApiResponse<{
      optimal_allocation: Record<string, number>;
      expected_return: number;
      expected_risk: number;
      sharpe_ratio: number;
    }>>(`/portfolio/optimize`, {
      method: 'POST',
      body: JSON.stringify({ allocation, objective, constraints }),
    });
  },

  getEfficientFrontier: (symbols: string[], points: number = 20) => {
    return request<ApiResponse<Array<{ return: number; risk: number; allocation: Record<string, number> }>>>(`/portfolio/efficient-frontier`, {
      method: 'POST',
      body: JSON.stringify({ symbols, points }),
    });
  },
};

// 投资组合配置API
export const portfolioConfigAPI = {
  getAll: () => {
    return request<ApiResponse<PortfolioConfig[]>>(`/portfolio-configs/`);
  },

  getById: (id: string) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/${id}`);
  },

  create: (config: Omit<PortfolioConfig, 'id' | 'created_at' | 'updated_at'>) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/`, {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  update: (id: string, config: Partial<PortfolioConfig>) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  delete: (id: string) => {
    return request<ApiResponse<void>>(`/portfolio-configs/${id}`, {
      method: 'DELETE',
    });
  },

  toggleStatus: (id: string) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/${id}/toggle-status`, {
      method: 'POST',
    });
  },

  analyze: (id: string) => {
    return request<ApiResponse<PortfolioAnalysisResult>>(`/portfolio-configs/${id}/analyze`, {
      method: 'POST',
    });
  },
};

// 优化相关API
export const optimizationAPI = {
  mptOptimize: (symbols: string[], objective: string = 'max_sharpe', targetReturn?: number, riskFreeRate?: number) => {
    return request<ApiResponse<{
      weights: Record<string, number>;
      expected_return: number;
      volatility: number;
      sharpe_ratio: number;
      sortino_ratio?: number;
      diversification_ratio?: number;
      risk_contribution?: Record<string, number>;
    }>>(`/optimization/mpt`, {
      method: 'POST',
      body: JSON.stringify({
        symbols,
        objective,
        target_return: objective === 'target_return' ? targetReturn : undefined,
        risk_free_rate: riskFreeRate,
      }),
    });
  },

  efficientFrontier: (symbols: string[], points: number = 20) => {
    return request<ApiResponse<Array<{ target_return: number; min_volatility: number; optimal_weights: Record<string, number>; sharpe_ratio: number }>>>(`/optimization/efficient-frontier`, {
      method: 'POST',
      body: JSON.stringify({ symbols, num_points: points }),
    });
  },

  covarianceMatrix: (symbols: string[]) => {
    return request<ApiResponse<Record<string, Record<string, number>>>>(`/optimization/covariance`, {
      method: 'POST',
      body: JSON.stringify({ symbols }),
    });
  },

  etfStatistics: (symbols: string[]) => {
    return request<ApiResponse<Record<string, {
      mean_return: number;
      volatility: number;
      sharpe_ratio: number;
      max_drawdown: number;
    }>>>(`/optimization/etf-statistics`, {
      method: 'POST',
      body: JSON.stringify({ symbols }),
    });
  },

  riskParity: (symbols: string[]) => {
    return request<ApiResponse<{
      weights: Record<string, number>;
      risk_contributions: Record<string, number>;
      volatility?: number;
      diversification_ratio?: number;
    }>>(`/optimization/risk-parity`, {
      method: 'POST',
      body: JSON.stringify({ symbols }),
    });
  },

  blackLitterman: (symbols: string[], views: Array<{ symbol: string; return: number; confidence: number }>) => {
    return request<ApiResponse<{
      posterior_returns: Record<string, number>;
      optimal_weights: Record<string, number>;
      expected_return?: number;
      expected_risk?: number;
      sharpe_ratio?: number;
    }>>(`/optimization/black-litterman`, {
      method: 'POST',
      body: JSON.stringify({ symbols, views }),
    });
  },

  marketImpliedReturns: (symbols: string[], marketPortfolio?: Record<string, number>) => {
    return request<ApiResponse<Record<string, number>>>(`/optimization/market-implied-returns`, {
      method: 'POST',
      body: JSON.stringify({ symbols, market_portfolio: marketPortfolio }),
    });
  },
};

// 因子分析API
export const factorAPI = {
  analyzeExposure: (symbol: string, factors: string[] = ['market', 'size', 'value', 'momentum', 'quality']) => {
    return request<ApiResponse<{
      exposures: Record<string, number>;
      r_squared: number;
    }>>(`/factor/analyze`, {
      method: 'POST',
      body: JSON.stringify({ symbol, factors }),
    });
  },

  analyzePortfolio: (allocation: Record<string, number>) => {
    return request<ApiResponse<{
      portfolio_exposures: Record<string, number>;
      factor_contributions: Record<string, number>;
    }>>(`/factor/portfolio`, {
      method: 'POST',
      body: JSON.stringify({ allocation }),
    });
  },

  analyzeMultipleAssets: (symbols: string[]) => {
    return request<ApiResponse<Record<string, Record<string, number>>>>(`/factor/multi-asset`, {
      method: 'POST',
      body: JSON.stringify({ symbols }),
    });
  },

  getStatistics: () => {
    return request<ApiResponse<{
      factors: Array<{ name: string; description: string; annualized_return: number; volatility: number }>;
      correlations: Record<string, Record<string, number>>;
    }>>(`/factor/statistics`);
  },

  decomposeRisk: (allocation: Record<string, number>) => {
    return request<ApiResponse<{
      total_risk: number;
      factor_risks: Record<string, number>;
      idiosyncratic_risk: number;
    }>>(`/factor/risk-decomposition`, {
      method: 'POST',
      body: JSON.stringify({ allocation }),
    });
  },

  compareAttribution: (allocation1: Record<string, number>, allocation2: Record<string, number>) => {
    return request<ApiResponse<{
      allocation1_exposures: Record<string, number>;
      allocation2_exposures: Record<string, number>;
      differences: Record<string, number>;
    }>>(`/factor/compare`, {
      method: 'POST',
      body: JSON.stringify({ allocation1, allocation2 }),
    });
  },
};

// ETF持仓相关API
export const etfHoldingAPI = {
  getHoldings: (symbol: string) => {
    return request<ApiResponse<{
      symbol: string;
      holdings: Array<{ symbol: string; name: string; weight: number; sector: string }>;
      updated_at: string;
    }>>(`/etf/${symbol}/holdings`);
  },

  getOverlap: (symbols: string[]) => {
    return request<ApiResponse<{
      overlap_matrix: Record<string, Record<string, number>>;
      average_overlap: number;
    }>>(`/etf/overlap?symbols=${symbols.join(',')}`);
  },

  getTopHoldings: (symbol: string, limit: number = 10) => {
    return request<ApiResponse<Array<{ symbol: string; name: string; weight: number; sector: string }>>>(`/etf/${symbol}/top-holdings?limit=${limit}`);
  },

  getSectorAllocation: (symbol: string) => {
    return request<ApiResponse<Record<string, number>>>(`/etf/${symbol}/sector-allocation`);
  },

  compareHoldings: (symbols: string[]) => {
    return request<ApiResponse<{
      common_holdings: Array<{ symbol: string; name: string; weights: Record<string, number> }>;
      total_overlap: number;
    }>>(`/etf/holdings/comparison`, {
      method: 'POST',
      body: JSON.stringify({ symbols }),
    });
  },

  saveHoldings: (symbol: string, holdings: Array<{ symbol: string; weight: number }>) => {
    return request<ApiResponse<void>>(`/etf/${symbol}/holdings`, {
      method: 'POST',
      body: JSON.stringify({ holdings }),
    });
  },
};

// 回测相关API
export const backtestAPI = {
  run: (allocation: Record<string, number>, startDate: string, endDate: string, rebalanceFrequency: string = 'monthly') => {
    return request<ApiResponse<{
      total_return: number;
      annualized_return: number;
      annualized_volatility: number;
      sharpe_ratio: number;
      max_drawdown: number;
      monthly_returns: Array<{ date: string; return: number }>;
    }>>(`/backtest/run`, {
      method: 'POST',
      body: JSON.stringify({ allocation, start_date: startDate, end_date: endDate, rebalance_frequency: rebalanceFrequency }),
    });
  },

  runEventDriven: (allocation: Record<string, number>, events: Array<{ date: string; type: string; impact: Record<string, number> }>) => {
    return request<ApiResponse<{
      total_return: number;
      event_impacts: Array<{ date: string; type: string; portfolio_impact: number }>;
    }>>(`/backtest/event-driven`, {
      method: 'POST',
      body: JSON.stringify({ allocation, events }),
    });
  },

  listStrategies: () => {
    return request<ApiResponse<Array<{ id: string; name: string; description: string; parameters: Record<string, unknown> }>>>(`/backtest/strategies`);
  },

  analyzeFactors: (allocation: Record<string, number>, period: string = '1y') => {
    return request<ApiResponse<{
      factor_returns: Record<string, number>;
      factor_attributions: Record<string, number>;
    }>>(`/backtest/factors`, {
      method: 'POST',
      body: JSON.stringify({ allocation, period }),
    });
  },
};

// ETF配置API
export const etfConfigAPI = {
  getAll: () => {
    return request<ApiResponse<ETFConfig[]>>(`/etf-configs/`);
  },

  getById: (id: string) => {
    return request<ApiResponse<ETFConfig>>(`/etf-configs/${id}`);
  },

  create: (config: Omit<ETFConfig, 'id' | 'created_at' | 'updated_at'>) => {
    return request<ApiResponse<ETFConfig>>(`/etf-configs/`, {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  update: (id: string, config: Partial<ETFConfig>) => {
    return request<ApiResponse<ETFConfig>>(`/etf-configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  delete: (id: string) => {
    return request<ApiResponse<void>>(`/etf-configs/${id}`, {
      method: 'DELETE',
    });
  },

  toggleStatus: (id: string) => {
    return request<ApiResponse<ETFConfig>>(`/etf-configs/${id}/toggle-status`, {
      method: 'POST',
    });
  },

  toggleAutoUpdate: (id: string) => {
    return request<ApiResponse<ETFConfig>>(`/etf-configs/${id}/auto-update`, {
      method: 'POST',
    });
  },
};

// A股相关API
export const aShareAPI = {
  getDefaultETFs: () => {
    return request<ApiResponse<Array<{ symbol: string; name: string; dividend_yield: number; price: number }>>>(`/a-share/etfs`);
  },

  getDefaultPortfolio: () => {
    return request<ApiResponse<{
      portfolio_id: number;
      total_investment: number;
      expected_annual_dividend: number;
      average_dividend_yield: number;
      monthly_dividend: number;
      quarterly_dividend: number;
      holdings: Array<{
        symbol: string;
        name: string;
        current_price: number;
        previous_close: number;
        price_change: number;
        price_change_pct: number;
        volume: number;
        turnover: number;
        investment: number;
        weight: number;
        dividend_yield: number;
        dividend_frequency: string;
        expected_dividend: number;
        dividend_contribution: number;
      }>;
    }>>(`/a-share/portfolio/default`);
  },

  analyzePortfolio: (etfs: Array<{ symbol: string; weight: number }>, totalInvestment: number = 100000) => {
    return request<ApiResponse<{
      expected_dividend: number;
      dividend_yield: number;
      risk_metrics: Record<string, number>;
    }>>(`/a-share/portfolio/analyze`, {
      method: 'POST',
      body: JSON.stringify({ etfs, total_investment: totalInvestment }),
    });
  },

  getPrices: () => {
    return request<ApiResponse<Array<{
      symbol: string;
      name: string;
      current_price: number;
      previous_close: number;
      price_change: number;
      price_change_pct: number;
      volume: number;
      turnover: number;
      nav: number;
      premium_rate: number;
      price_updated_at: string;
    }>>>(`/a-share/prices`);
  },

  refreshPrices: () => {
    return request<ApiResponse<{ updated: number; failed: number }>>(`/a-share/prices/refresh`, {
      method: 'POST',
    });
  },
};

// 汇率相关API
export const exchangeRateAPI = {
  getAll: () => {
    return request<ApiResponse<ExchangeRate[]>>(`/exchange-rates/`);
  },

  getRate: (from: string, to: string) => {
    return request<ApiResponse<ExchangeRate>>(`/exchange-rates/${from}/${to}`);
  },

  convert: (amount: number, from: string, to: string) => {
    return request<ApiResponse<{ amount: number; from: string; to: string; result: number; rate: number }>>(`/exchange-rates/convert`, {
      method: 'POST',
      body: JSON.stringify({ amount, from, to }),
    });
  },

  sync: () => {
    return request<ApiResponse<{ message: string; updated: number }>>(`/exchange-rates/sync`, {
      method: 'POST',
    });
  },

  getSummary: () => {
    return request<ApiResponse<{
      total_pairs: number;
      last_update: string;
      currencies: string[];
    }>>(`/exchange-rates/summary`);
  },

  getCurrencies: () => {
    return request<ApiResponse<string[]>>(`/exchange-rates/currencies`);
  },
};

// 通用ETF API
export const universalETFAPI = {
  initialize: () => {
    return request<ApiResponse<{ message: string; count: number }>>(`/universal-etf/initialize`, {
      method: 'POST',
    });
  },

  getAll: () => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/`);
  },

  getBySymbol: (symbol: string) => {
    return request<ApiResponse<ETFData>>(`/universal-etf/${symbol}`);
  },

  getByAssetClass: (assetClass: string) => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/asset-class/${assetClass}`);
  },

  getByRegion: (region: string) => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/region/${region}`);
  },

  getByType: (etfType: string) => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/type/${etfType}`);
  },

  search: (query: string) => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/search?q=${encodeURIComponent(query)}`);
  },

  filter: (filters: Record<string, string | number | boolean>) => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/filter`, {
      method: 'POST',
      body: JSON.stringify(filters),
    });
  },

  getAssetClassDistribution: () => {
    return request<ApiResponse<Record<string, number>>>(`/universal-etf/distribution/asset-class`);
  },

  getRegionDistribution: () => {
    return request<ApiResponse<Record<string, number>>>(`/universal-etf/distribution/region`);
  },

  compare: (symbols: string[]) => {
    return request<ApiResponse<{
      comparison: Record<string, Record<string, number | string>>;
      correlations: Record<string, Record<string, number>>;
    }>>(`/universal-etf/compare`, {
      method: 'POST',
      body: JSON.stringify({ symbols }),
    });
  },

  getPortfolioAllocation: () => {
    return request<ApiResponse<Record<string, number>>>(`/universal-etf/portfolio-allocation`);
  },

  getCategories: () => {
    return request<ApiResponse<Array<{ id: string; name: string; count: number }>>>(`/universal-etf/categories`);
  },

  getTopPerformers: (period: string = '1y', limit: number = 10) => {
    return request<ApiResponse<ETFData[]>>(`/universal-etf/top-performers?period=${period}&limit=${limit}`);
  },
};

// 操作日志API
export const operationLogsAPI = {
  getLogs: (params?: {
    page?: number;
    page_size?: number;
    log_type?: string;
    action_type?: string;
    user?: string;
    module?: string;
    start_date?: string;
    end_date?: string;
    status?: string;
  }) => {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          queryParams.append(key, String(value));
        }
      });
    }
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    return request<ApiResponse<{
      data: Array<{
        id: number;
        log_type: string;
        timestamp: string;
        user: string;
        module: string;
        action_type: string;
        details: string;
        ip: string;
        status: string;
        status_code: number;
        error_message: string;
        duration_ms: number;
      }>;
      meta: {
        pagination: {
          page: number;
          page_size: number;
          total: number;
          total_pages: number;
          has_next: boolean;
          has_prev: boolean;
        };
        summary: {
          total_logs: number;
          total_audit: number;
          total_operation: number;
        };
      };
    }>>(`/logs${query}`);
  },

  getLogTypes: () => {
    return request<ApiResponse<string[]>>(`/logs/types`);
  },

  getActionTypes: () => {
    return request<ApiResponse<string[]>>(`/logs/action-types`);
  },

  getUsers: () => {
    return request<ApiResponse<string[]>>(`/logs/users`);
  },

  exportLogs: (params?: {
    format?: 'csv' | 'json';
    start_date?: string;
    end_date?: string;
    log_type?: string;
  }) => {
    return request<ApiResponse<{ download_url: string; filename: string }>>(`/logs/export`, {
      method: 'POST',
      body: JSON.stringify(params || {}),
    });
  },
};

// 因子择时API
export const factorTimingAPI = {
  calculateSignal: async (factorName: string, lookbackDays: number): Promise<ApiResponse<FactorTimingSignal>> => {
    return request<ApiResponse<FactorTimingSignal>>('/factor/timing/calculate', {
      method: 'POST',
      body: JSON.stringify({
        factor_name: factorName,
        lookback_days: lookbackDays,
      }),
    });
  },

  getSignalHistory: async (factorName: string, startDate?: string, endDate?: string): Promise<ApiResponse<FactorTimingSignal[]>> => {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);

    const queryString = params.toString();
    const url = queryString
      ? `/factor/timing/history/${encodeURIComponent(factorName)}?${queryString}`
      : `/factor/timing/history/${encodeURIComponent(factorName)}`;

    return request<ApiResponse<FactorTimingSignal[]>>(url);
  },

  generateView: async (signal: FactorTimingSignal, targetAsset: string): Promise<ApiResponse<AlphaView>> => {
    return request<ApiResponse<AlphaView>>('/alpha-views/generate-from-factor', {
      method: 'POST',
      body: JSON.stringify({
        factor_name: signal.factor_name,
        asset_symbol: targetAsset,
      }),
    });
  },
};

// Alpha观点API
export const alphaViewAPI = {
  create: async (view: CreateAlphaViewRequest): Promise<ApiResponse<AlphaView>> => {
    return request<ApiResponse<AlphaView>>('/alpha-views', {
      method: 'POST',
      body: JSON.stringify(view),
    });
  },

  getActive: async (assetSymbol?: string, method?: ViewMethod): Promise<ApiResponse<AlphaView[]>> => {
    const params = new URLSearchParams();
    if (assetSymbol) params.append('asset_symbol', assetSymbol);
    if (method) params.append('method', method);

    return request<ApiResponse<AlphaView[]>>(`/alpha-views/active?${params.toString()}`);
  },

  getById: async (id: number): Promise<ApiResponse<AlphaView>> => {
    return request<ApiResponse<AlphaView>>(`/alpha-views/${id}`);
  },

  update: async (id: number, view: UpdateAlphaViewRequest): Promise<ApiResponse<AlphaView>> => {
    return request<ApiResponse<AlphaView>>(`/alpha-views/${id}`, {
      method: 'PUT',
      body: JSON.stringify(view),
    });
  },

  deactivate: async (id: number): Promise<ApiResponse<void>> => {
    return request<ApiResponse<void>>(`/alpha-views/${id}/deactivate`, {
      method: 'POST',
    });
  },

  validate: async (id: number, actualReturn: number): Promise<ApiResponse<AlphaViewPerformance>> => {
    return request<ApiResponse<AlphaViewPerformance>>(`/alpha-views/${id}/validate`, {
      method: 'POST',
      body: JSON.stringify({ actual_return: actualReturn }),
    });
  },
};

// Black-Litterman API
export const blackLittermanAPI = {
  createConfig: async (config: CreateBLConfigRequest): Promise<ApiResponse<BlackLittermanConfig>> => {
    return request<ApiResponse<BlackLittermanConfig>>('/black-litterman/configs', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  calculate: async (configId: number, viewIds: number[]): Promise<ApiResponse<BLPosteriorReturn>> => {
    return request<ApiResponse<BLPosteriorReturn>>('/black-litterman/calculate', {
      method: 'POST',
      body: JSON.stringify({
        config_id: configId,
        view_ids: viewIds,
      }),
    });
  },

  getPosteriorReturns: async (configId: number): Promise<ApiResponse<BLPosteriorReturn>> => {
    return request<ApiResponse<BLPosteriorReturn>>(`/black-litterman/results/${configId}`);
  },
};

// 风险预算API
export const riskBudgetAPI = {
  createConfig: async (config: CreateRiskBudgetRequest): Promise<ApiResponse<RiskBudgetConfig>> => {
    return request<ApiResponse<RiskBudgetConfig>>('/risk-budget/configs', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  calculateCVaR: async (configId: number, weights: number[]): Promise<ApiResponse<CVaRResult>> => {
    return request<ApiResponse<CVaRResult>>('/risk-budget/calculate-cvar', {
      method: 'POST',
      body: JSON.stringify({
        config_id: configId,
        weights: weights,
      }),
    });
  },

  runMonteCarlo: async (configId: number, simulations: number, timeSteps: number): Promise<ApiResponse<MonteCarloSimulation>> => {
    return request<ApiResponse<MonteCarloSimulation>>('/risk-budget/monte-carlo', {
      method: 'POST',
      body: JSON.stringify({
        config_id: configId,
        simulations: simulations,
        time_steps: timeSteps,
      }),
    });
  },
};

// QuantLib API
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
    return request<ApiResponse<unknown>>(`/quantlib/reference/${encodeURIComponent(type)}`);
  },
};

// Agent Service API
export const agentAPI = {
  health: async (): Promise<ApiResponse<Record<string, unknown>>> => {
    return request<ApiResponse<Record<string, unknown>>>('/agents/health');
  },

  discover: async (): Promise<ApiResponse<AgentInfo[]>> => {
    return request<ApiResponse<AgentInfo[]>>('/agents/discover');
  },

  run: async (req: AgentRunRequest): Promise<ApiResponse<AgentRunResponse>> => {
    return request<ApiResponse<AgentRunResponse>>('/agents/run', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  runTeam: async (req: AgentTeamRequest): Promise<ApiResponse<AgentTeamResponse>> => {
    return request<ApiResponse<AgentTeamResponse>>('/agents/team', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },
};
