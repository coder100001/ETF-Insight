// QuantLib API Types
// Note: decimal fields are strings to preserve precision (shopspring/decimal serializes as JSON string)

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

// Result types use string for decimal precision (backend serializes decimal.Decimal as JSON string)
export interface OptionResult {
  price: string;
  delta: string;
  gamma: string;
  theta: string;
  vega: string;
  rho: string;
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
  rates: string[];
  zero_rates: string[];
  forward_rates: string[];
  discount_factors: string[];
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
  dirty_price: string;
  clean_price: string;
  duration: string;
  modified_duration: string;
  convexity: string;
  yield_to_maturity: string;
  accrued_interest: string;
}

export interface VaRRequest {
  portfolio_value: number;
  returns: number[];
  confidence: number;
  holding_period?: number;
  method?: 'historical' | 'parametric' | 'monte_carlo';
}

export interface VaRResult {
  var: string;
  cvar: string;
  confidence: string;
  holding_period: number;
  method: string;
}
