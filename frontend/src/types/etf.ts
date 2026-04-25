// ETF相关类型定义

export interface ETFInfo {
  symbol: string;
  name: string;
}

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

export interface ETFConfig {
  id: number;
  symbol: string;
  name: string;
  description?: string;
  strategy?: string;
  focus?: string;
  expense_ratio?: number;
  currency?: string;
  exchange?: string;
  category?: string;
  provider?: string;
  inception?: string;
  aum?: number;
  status: number;
  is_active?: boolean;
  auto_update?: boolean;
  update_frequency?: string;
  last_updated?: string;
  data_source?: string;
  created_at?: string;
  updated_at?: string;
}

export interface ETFStatistics {
  symbol: string;
  name: string;
  annualized: number;
  volatility: number;
  sharpe_ratio: number;
  max_drawdown: number;
  mean_return?: number;
}

export interface ETFHistoryDataItem {
  date: string;
  close_price?: number;
  volume?: number;
  open_price?: number;
  high_price?: number;
  low_price?: number;
  price?: number;
}

export interface ETFForecastResult {
  years: number;
  future_value: number;
  capital_appreciation: number;
  total_dividend_before_tax: number;
  total_dividend_after_tax: number;
  dividend_tax: number;
  annual_return_rate: number;
  effective_annual_return_rate: number;
}

export interface AShareDividendETF {
  id: number;
  symbol: string;
  name: string;
  dividend_yield_min: number;
  dividend_yield_max: number;
  dividend_frequency: '月分' | '季分' | '年分';
  benchmark: string;
  exchange: string;
  management_fee: number;
  description: string;
  status: number;
}

export interface AShareETFPrice {
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
}
