// 优化服务层 - 处理数据转换和验证

import { optimizationAPI } from './api';
import { APIError, extractErrorMessage } from '../utils/errorHandler';
import type {
  OptimizationResult,
  EfficientFrontierPoint,
  RiskParityResult,
  BlackLittermanResult,
} from '../types';

export interface MPTOptimizeParams {
  symbols: string[];
  objective?: 'min_volatility' | 'max_sharpe' | 'target_return';
  targetReturn?: number;
  riskFreeRate?: number;
}

export interface BlackLittermanView {
  symbol: string;
  return: number;
  confidence: number;
}

/**
 * MPT优化服务
 * @param params - 优化参数
 * @returns 优化结果
 */
export const optimizeMPT = async (params: MPTOptimizeParams): Promise<OptimizationResult> => {
  const response = await optimizationAPI.mptOptimize(
    params.symbols,
    params.objective || 'max_sharpe',
    params.targetReturn,
    params.riskFreeRate
  );

  if (!response.success) {
    throw new APIError(extractErrorMessage(response), 400);
  }

  if (!response.data) {
    throw new APIError('优化结果为空', 500);
  }

  return {
    weights: response.data.weights ?? {},
    expected_return: response.data.expected_return,
    expected_risk: response.data.volatility,
    volatility: response.data.volatility,
    sharpe_ratio: response.data.sharpe_ratio,
    sortino_ratio: 0,
    diversification_ratio: 0,
    risk_contribution: {},
  };
};

/**
 * 计算有效前沿
 * @param symbols - ETF代码列表
 * @param points - 点数
 * @returns 有效前沿数据
 */
export const calculateEfficientFrontier = async (
  symbols: string[],
  points: number = 20
): Promise<EfficientFrontierPoint[]> => {
  const response = await optimizationAPI.efficientFrontier(symbols, points);

  if (!response.success) {
    throw new APIError(extractErrorMessage(response), 400);
  }

  if (!response.data) {
    throw new APIError('有效前沿数据为空', 500);
  }

  return response.data.map(point => ({
    target_return: point.target_return,
    min_volatility: point.min_volatility,
    optimal_weights: point.optimal_weights,
    sharpe_ratio: point.sharpe_ratio,
  }));
};

/**
 * 风险平价优化
 * @param symbols - ETF代码列表
 * @returns 风险平价结果
 */
export const optimizeRiskParity = async (symbols: string[]): Promise<RiskParityResult> => {
  const response = await optimizationAPI.riskParity(symbols);

  if (!response.success) {
    throw new APIError(extractErrorMessage(response), 400);
  }

  if (!response.data) {
    throw new APIError('风险平价结果为空', 500);
  }

  return {
    weights: response.data.weights,
    risk_contributions: response.data.risk_contributions,
    volatility: 0,
    diversification_ratio: 0,
  };
};

/**
 * Black-Litterman优化
 * @param symbols - ETF代码列表
 * @param views - 投资者观点
 * @returns Black-Litterman结果
 */
export const optimizeBlackLitterman = async (
  symbols: string[],
  views: BlackLittermanView[]
): Promise<BlackLittermanResult> => {
  const response = await optimizationAPI.blackLitterman(symbols, views);

  if (!response.success) {
    throw new APIError(extractErrorMessage(response), 400);
  }

  if (!response.data) {
    throw new APIError('Black-Litterman结果为空', 500);
  }

  return {
    posterior_returns: response.data.posterior_returns,
    optimal_weights: response.data.optimal_weights,
    expected_return: 0,
    expected_risk: 0,
    sharpe_ratio: 0,
  };
};

/**
 * 获取ETF统计数据
 * @param symbols - ETF代码列表
 * @returns 统计数据
 */
export const getETFStatistics = async (
  symbols: string[]
): Promise<Record<string, {
  mean_return: number;
  volatility: number;
  sharpe_ratio: number;
  max_drawdown: number;
}>> => {
  const response = await optimizationAPI.etfStatistics(symbols);

  if (!response.success) {
    throw new APIError(extractErrorMessage(response), 400);
  }

  if (!response.data) {
    throw new APIError('统计数据为空', 500);
  }

  return response.data;
};
