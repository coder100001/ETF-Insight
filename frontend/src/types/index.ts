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

// 投资组合配置
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

// ETF配置
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
  status: number;  // 1: 启用, 0: 禁用
  is_active?: boolean;  // 前端使用的字段
  auto_update?: boolean;  // 是否自动更新
  update_frequency?: string;  // 更新频率
  last_updated?: string;  // 最后更新时间
  data_source?: string;  // 数据源
  created_at?: string;
  updated_at?: string;
}

// A股红利ETF
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

// A股ETF价格信息
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

// A股组合持仓明细
export interface AShareHoldingDetail {
  symbol: string;
  name: string;
  investment: number;
  weight: number;
  dividend_yield: number;
  dividend_frequency: string;
  expected_dividend: number;
  dividend_contribution: number;
  // 价格信息
  current_price?: number;
  price_change_pct?: number;
  volume?: number;
}

// A股分红计算结果
export interface AShareDividendCalculation {
  total_investment: number;
  expected_annual_dividend: number;
  average_dividend_yield: number;
  monthly_dividend: number;
  quarterly_dividend: number;
  holdings: AShareHoldingDetail[];
}

// ETF历史数据条目 (兼容后端两种返回格式: {date, price} 或 {date, close_price, volume})
export interface ETFHistoryDataItem {
  date: string;
  close_price?: number;
  volume?: number;
  open_price?: number;
  high_price?: number;
  low_price?: number;
  price?: number;  // 兼容旧格式
}

// ETF预测结果
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

// ETF统计数据
export interface ETFStatistics {
  symbol: string;
  name: string;
  annualized: number;
  volatility: number;
  sharpe_ratio: number;
  max_drawdown: number;
  mean_return?: number;
}

// 优化结果
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

// 有效前沿点
export interface EfficientFrontierPoint {
  target_return: number;
  min_volatility: number;
  optimal_weights: Record<string, number>;
  sharpe_ratio: number;
}

// 风险平价结果
export interface RiskParityResult {
  weights: Record<string, number>;
  risk_contributions: Record<string, number>;
  volatility: number;
  diversification_ratio: number;
}

// Black-Litterman结果
export interface BlackLittermanResult {
  posterior_returns: Record<string, number>;
  optimal_weights: Record<string, number>;
  expected_return: number;
  expected_risk: number;
  sharpe_ratio: number;
}

// 风险分析结果
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
  holdings?: PortfolioRisk[];
}

// 组合风险
export interface PortfolioRisk {
  symbol: string;
  weight: number;
  componentVar: number;
  marginalVar: number;
}
