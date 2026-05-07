// 投资组合相关类型定义

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

export interface PortfolioAnalysisResult {
  total_investment: number;
  total_value: number;
  total_return: number;
  total_return_pct: number;
  annual_dividend_before_tax: number;
  annual_dividend_after_tax: number;
  dividend_yield: number;
  tax_rate: number;
  after_tax_return: number;
  dividend_tax: number;
  total_return_with_dividend: number;
  total_return_with_dividend_percent: number;
  holdings: PortfolioHolding[];
}

export interface PortfolioConfig {
  id: number;
  name: string;
  description?: string;
  allocation: string | Record<string, number>;
  total_investment: number;
  status: number;
  created_at: string;
  updated_at: string;
  is_default?: boolean;
}

export interface AShareHoldingDetail {
  symbol: string;
  name: string;
  current_price?: number;
  previous_close?: number;
  price_change?: number;
  price_change_pct?: number;
  volume?: number;
  turnover?: number;
  investment: number;
  weight: number;
  dividend_yield: number;
  dividend_frequency: string;
  expected_dividend: number;
  dividend_contribution: number;
}

export interface AShareDividendCalculation {
  total_investment: number;
  expected_annual_dividend: number;
  average_dividend_yield: number;
  monthly_dividend: number;
  quarterly_dividend: number;
  holdings: AShareHoldingDetail[];
}

export interface PortfolioRisk {
  symbol: string;
  weight: number;
  componentVar: number;
  marginalVar: number;
}
