// ETF数据类型
export interface ETFData {
  symbol: string;
  name: string;
  current_price: number;
  previous_close: number;
  change: number;
  change_percent: number;
  open_price: number;
  high_price: number;
  low_price: number;
  volume: number;
  dividend_yield?: number;
  volatility?: number;
  total_return?: number;
  max_drawdown?: number;
  sharpe_ratio?: number;
  expense_ratio?: number;
  focus?: string;
  strategy?: string;
  description?: string;
  info?: {
    focus: string;
    strategy: string;
    description?: string;
  };
}

// 投资组合持仓
export interface PortfolioHolding {
  symbol: string;
  name: string;
  weight: number;
  investment: number;
  current_price: number;
  shares: number;
  current_value: number;
  capital_gain: number;
  capital_gain_percent: number;
  total_return: number;
  volatility: number;
  dividend_yield?: number;
  annual_dividend_before_tax: number;
  annual_dividend_after_tax: number;
}

// 投资组合结果
export interface PortfolioResult {
  total_investment: number;
  total_value: number;
  total_return: number;
  total_return_percent: number;
  annual_dividend_before_tax: number;
  annual_dividend_after_tax: number;
  dividend_tax: number;
  tax_rate: number;
  weighted_dividend_yield: number;
  total_return_with_dividend: number;
  total_return_with_dividend_percent: number;
  holdings: PortfolioHolding[];
  portfolio_metrics?: {
    weighted_return: number;
    volatility: number;
    sharpe_ratio: number;
  };
}

// 投资组合分析（API返回类型）
export interface PortfolioAnalysisResult {
  total_value: number;
  total_return: number;
  total_return_pct: number;
  annual_dividend_before_tax: number;
  annual_dividend: number;
  dividend_yield: number;
  tax_rate: number;
  after_tax_return: number;
  dividend_tax: number;
  holdings: PortfolioHolding[];
}

// 投资组合配置
export interface PortfolioConfig {
  id: number;
  name: string;
  description?: string;
  allocation: Record<string, number>;
  total_investment: number;
  status: number;
  created_at: string;
  updated_at: string;
}

// 预测数据
export interface ForecastData {
  years: number;
  future_value: number;
  capital_appreciation: number;
  total_dividend_before_tax: number;
  total_dividend_after_tax: number;
  dividend_tax: number;
  annual_return_rate: number;
  effective_annual_return_rate: number;
}

export interface ETFForecast {
  [symbol: string]: {
    [year: number]: ForecastData;
  };
}

// 场景预测
export interface ScenarioForecast {
  years: {
    [year: number]: ForecastData;
  };
}

export interface ScenarioForecasts {
  pessimistic: ScenarioForecast;
  conservative: ScenarioForecast;
  neutral: ScenarioForecast;
  optimistic: ScenarioForecast;
}

// 用户配置
export interface UserConfig {
  total_investment: number;
  allocation: {
    [symbol: string]: number;
  };
  tax_rate: number;
}

// 实时更新结果
export interface UpdateResult {
  symbol: string;
  success: boolean;
  price?: number;
  open?: number;
  high?: number;
  low?: number;
  volume?: number;
  error?: string;
}

export interface RealtimeUpdateResponse {
  success: boolean;
  update_time: string;
  summary: {
    total: number;
    success: number;
    failed: number;
  };
  update_results: UpdateResult[];
  portfolio?: PortfolioResult;
}

// 汇率数据
export interface ExchangeRate {
  from_currency: string;
  to_currency: string;
  rate: number;
  updated_at: string;
}

// 图表数据
export interface ChartDataPoint {
  date: string;
  value: number;
  [key: string]: string | number | undefined;
}

// 菜单项
export interface MenuItem {
  key: string;
  label: string;
  icon?: string;
  path: string;
  children?: MenuItem[];
}

// 页面Props
export interface PageProps {
  title?: string;
}

// 通用响应
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

// 工作流统计
export interface WorkflowStat {
  name: string;
  total: number;
  success: number;
  failed: number;
  success_rate: number;
  status: 'good' | 'warning' | 'danger';
}

// 每日统计
export interface DailyStatItem {
  total: number;
  success: number;
  failed: number;
}

export interface DailyStat {
  [date: string]: DailyStatItem;
}