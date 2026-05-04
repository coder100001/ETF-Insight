// QuantLib API Types

export interface EuropeanOptionRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
  dividend_yield?: number;
}

export interface AmericanOptionRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
  steps?: number;
  dividend_yield?: number;
}

export interface GreeksRequest {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
}

export interface OptionResult {
  price: number;
  delta: number;
  gamma: number;
  theta: number;
  vega: number;
  rho: number;
}

export interface YieldCurveRequest {
  currency: string;
  calendar?: string;
  day_count?: string;
  tenors: string[];
  rates: number[];
  compounding?: string;
  frequency?: string;
}

export interface YieldCurveResult {
  currency: string;
  tenors: string[];
  rates: number[];
  zero_rates: number[];
  forward_rates: number[];
  discount_factors: number[];
}

export interface BondRequest {
  face_value: number;
  coupon_rate: number;
  frequency: number;
  maturity: string;
  yield_to_maturity: number;
  settlement_date?: string;
  day_count?: string;
}

export interface BondResult {
  dirty_price: number;
  clean_price: number;
  duration: number;
  modified_duration: number;
  convexity: number;
  yield_to_maturity: number;
  accrued_interest: number;
}

export interface VaRRequest {
  portfolio_value: number;
  returns: number[];
  confidence: number;
  holding_period?: number;
  method?: 'historical' | 'parametric' | 'monte_carlo';
}

export interface VaRResult {
  var: number;
  cvar: number;
  confidence: number;
  holding_period: number;
  method: string;
}
