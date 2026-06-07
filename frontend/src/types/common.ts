// 通用类型定义

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

export interface UserConfig {
  total_investment: number;
  allocation: {
    [symbol: string]: number;
  };
  tax_rate: number;
}

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
}

export interface ExchangeRate {
  from_currency: string;
  to_currency: string;
  rate: number;
  updated_at: string;
}

export interface ChartDataPoint {
  date: string;
  value: number;
  [key: string]: string | number | undefined;
}

export interface MenuItem {
  key: string;
  label: string;
  icon?: string;
  path: string;
  children?: MenuItem[];
}

export interface PageProps {
  title?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

export interface WorkflowStat {
  name: string;
  total: number;
  success: number;
  failed: number;
  success_rate: number;
  status: 'good' | 'warning' | 'danger';
}

export interface DailyStatItem {
  total: number;
  success: number;
  failed: number;
}

export interface DailyStat {
  [date: string]: DailyStatItem;
}

export interface FinancialConfig {
  risk_free_rate: number;
  trading_days_year: number;
  default_currency: string;
}
