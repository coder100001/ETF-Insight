// 优化分析相关类型定义

export interface OptimizationResult {
  weights: Record<string, number>;
  expected_return: number;
  expected_risk: number;
  volatility: number;
  sharpe_ratio: number;
  sortino_ratio: number;
  diversification_ratio: number;
  risk_contribution: Record<string, number>;
}

export interface EfficientFrontierPoint {
  target_return: number;
  min_volatility: number;
  optimal_weights: Record<string, number>;
  sharpe_ratio: number;
}

export interface RiskParityResult {
  weights: Record<string, number>;
  risk_contributions: Record<string, number>;
  volatility: number;
  diversification_ratio: number;
}

export interface BlackLittermanResult {
  posterior_returns: Record<string, number>;
  optimal_weights: Record<string, number>;
  expected_return: number;
  expected_risk: number;
  sharpe_ratio: number;
}

export interface RiskAnalysisResult {
  total_risk: number;
  systematic_risk: number;
  unsystematic_risk: number;
  diversification_ratio: number;
  concentration_risk: string;
  var_95?: number;
  var_99?: number;
  cvar_95?: number;
  max_drawdown?: number;
  holdings?: Array<{
    symbol: string;
    weight: number;
    componentVar: number;
    marginalVar: number;
  }>;
}
