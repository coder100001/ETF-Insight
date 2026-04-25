// 因子择时、Alpha观点、Black-Litterman、风险预算相关类型定义

export type SignalStrength =
  | 'strong_positive'
  | 'weak_positive'
  | 'neutral'
  | 'weak_negative'
  | 'strong_negative';

export type ViewType = 'absolute' | 'relative';
export type ViewMethod = 'factor_timing' | 'momentum' | 'mean_reversion';
export type ViewStatus = 'active' | 'expired' | 'validated';

export type PriorType = 'equal_weight' | 'min_variance' | 'market_cap';
export type OmegaMethod = 'Idzorek' | 'HeLitterman';

export type RiskMethod = 'historical' | 'parametric' | 'monte_carlo';

export interface FactorTimingSignal {
  id: number;
  factor_name: string;
  signal_date: string;
  ma_slope_60: number;
  z_score: number;
  percentile: number;
  signal_strength: SignalStrength;
  signal_score: number;
  expected_return: number;
  confidence: number;
  created_at: string;
}

export interface AlphaView {
  id: number;
  portfolio_id: number;
  asset_symbol: string;
  view_return: number;
  confidence: number;
  view_type: ViewType;
  view_method: ViewMethod;
  generated_at: string;
  valid_until: string;
  status: ViewStatus;
  source_factor: string;
  factor_loading: number;
  created_at: string;
  updated_at: string;
  performance?: AlphaViewPerformance;
}

export interface AlphaViewPerformance {
  id: number;
  view_id: number;
  actual_return: number;
  prediction_error: number;
  is_validated: boolean;
  validation_date: string;
  is_correct: boolean;
  rolling_win_rate: number;
  created_at: string;
}

export interface BlackLittermanConfig {
  id: number;
  portfolio_id: number;
  risk_aversion: number;
  prior_type: PriorType;
  prior_weights: string;
  implied_returns: string;
  omega_method: OmegaMethod;
  omega_matrix: string;
  is_active: boolean;
  last_calculated: string;
  created_at: string;
  updated_at: string;
}

export interface BLPosteriorReturn {
  id: number;
  config_id: number;
  calculation_date: string;
  posterior_returns: string;
  posterior_weights: string;
  posterior_cov: string;
  num_views: number;
  view_impact: number;
  created_at: string;
}

export interface RiskBudgetConfig {
  id: number;
  portfolio_id: number;
  cvar_limit: number;
  confidence_level: number;
  time_horizon: number;
  method: RiskMethod;
  risk_budgets: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RiskContribution {
  id: number;
  config_id: number;
  calculation_date: string;
  asset_symbol: string;
  weight: number;
  marginal_risk: number;
  risk_contribution: number;
  percentage_contribution: number;
  created_at: string;
}

export interface MonteCarloSimulation {
  id: number;
  config_id: number;
  simulation_date: string;
  num_simulations: number;
  time_steps: number;
  mean_return: number;
  std_return: number;
  percentile_5: number;
  percentile_95: number;
  simulation_data: string;
  created_at: string;
}

export interface CVaRResult {
  cvar: number;
  var: number;
  risk_contributions: RiskContribution[];
}

export interface CreateAlphaViewRequest {
  portfolio_id: number;
  asset_symbol: string;
  view_return: number;
  confidence: number;
  view_type: ViewType;
  view_method: ViewMethod;
  valid_until: string;
  source_factor: string;
}

export interface UpdateAlphaViewRequest {
  view_return?: number;
  confidence?: number;
  valid_until?: string;
  status?: ViewStatus;
}

export interface CreateBLConfigRequest {
  portfolio_id: number;
  risk_aversion: number;
  prior_type: PriorType;
  prior_weights: Record<string, number>;
  omega_method: OmegaMethod;
}

export interface UpdateBLConfigRequest {
  risk_aversion?: number;
  prior_type?: PriorType;
  prior_weights?: Record<string, number>;
  omega_method?: OmegaMethod;
  is_active?: boolean;
}

export interface CreateRiskBudgetRequest {
  portfolio_id: number;
  cvar_limit: number;
  confidence_level: number;
  time_horizon: number;
  method: RiskMethod;
  risk_budgets: Record<string, number>;
}

export interface UpdateRiskBudgetRequest {
  cvar_limit?: number;
  confidence_level?: number;
  time_horizon?: number;
  method?: RiskMethod;
  risk_budgets?: Record<string, number>;
  is_active?: boolean;
}
