import { useState, useCallback } from 'react';
import { message } from 'antd';
import { optimizationAPI } from '../services/api';
import type { OptimizationResult, EfficientFrontierPoint, RiskParityResult, BlackLittermanResult } from '../types';

interface UseOptimizationReturn {
  result: OptimizationResult | null;
  frontier: EfficientFrontierPoint[];
  rpResult: RiskParityResult | null;
  blResult: BlackLittermanResult | null;
  loading: boolean;
  error: string | null;
  optimize: (params: OptimizeParams) => Promise<void>;
  calculateFrontier: (symbols: string[], points?: number) => Promise<void>;
  calculateRiskParity: (symbols: string[]) => Promise<void>;
  calculateBlackLitterman: (symbols: string[], views: Array<{ symbol: string; return: number; confidence: number }>) => Promise<void>;
  reset: () => void;
}

export interface OptimizeParams {
  symbols: string[];
  objective: 'min_volatility' | 'max_sharpe' | 'target_return';
  targetReturn?: number;
  targetRisk?: number;
}

export const useOptimization = (): UseOptimizationReturn => {
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [frontier, setFrontier] = useState<EfficientFrontierPoint[]>([]);
  const [rpResult, setRpResult] = useState<RiskParityResult | null>(null);
  const [blResult, setBlResult] = useState<BlackLittermanResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const reset = useCallback(() => {
    setResult(null);
    setFrontier([]);
    setRpResult(null);
    setBlResult(null);
    setError(null);
  }, []);

  const optimize = useCallback(async (params: OptimizeParams) => {
    if (params.symbols.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await optimizationAPI.mptOptimize(
        params.symbols,
        params.objective === 'target_return' ? params.targetReturn : undefined,
        params.targetRisk
      );

      if (response.success && response.data) {
        const optResult: OptimizationResult = {
          weights: response.data.optimal_weights,
          expected_return: response.data.expected_return,
          expected_risk: response.data.expected_risk,
          volatility: response.data.expected_risk,
          sharpe_ratio: response.data.sharpe_ratio,
          sortino_ratio: response.data.sortino_ratio ?? 0,
          diversification_ratio: response.data.diversification_ratio ?? 0,
          risk_contribution: response.data.risk_contribution ?? {},
        };
        setResult(optResult);
        message.success('优化完成');
      } else {
        const errorMsg = response.message || '优化失败';
        setError(errorMsg);
        message.error(errorMsg);
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '优化请求失败';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  }, []);

  const calculateFrontier = useCallback(async (symbols: string[], points: number = 20) => {
    if (symbols.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await optimizationAPI.efficientFrontier(symbols, points);

      if (response.success && response.data) {
        const frontierData: EfficientFrontierPoint[] = response.data.map(point => ({
          target_return: point.return,
          min_volatility: point.risk,
          optimal_weights: point.weights,
          sharpe_ratio: point.return / point.risk,
        }));
        setFrontier(frontierData);
        message.success('有效前沿计算完成');
      } else {
        const errorMsg = response.message || '计算失败';
        setError(errorMsg);
        message.error(errorMsg);
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '计算请求失败';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  }, []);

  const calculateRiskParity = useCallback(async (symbols: string[]) => {
    if (symbols.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await optimizationAPI.riskParity(symbols);

      if (response.success && response.data) {
        const rpData: RiskParityResult = {
          weights: response.data.weights,
          risk_contributions: response.data.risk_contributions,
          volatility: response.data.volatility ?? 0,
          diversification_ratio: response.data.diversification_ratio ?? 0,
        };
        setRpResult(rpData);
        message.success('风险平价计算完成');
      } else {
        const errorMsg = response.message || '计算失败';
        setError(errorMsg);
        message.error(errorMsg);
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '计算请求失败';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  }, []);

  const calculateBlackLitterman = useCallback(async (
    symbols: string[],
    views: Array<{ symbol: string; return: number; confidence: number }>
  ) => {
    if (symbols.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await optimizationAPI.blackLitterman(symbols, views);

      if (response.success && response.data) {
        const blData: BlackLittermanResult = {
          posterior_returns: response.data.posterior_returns,
          optimal_weights: response.data.optimal_weights,
          expected_return: response.data.expected_return ?? 0,
          expected_risk: response.data.expected_risk ?? 0,
          sharpe_ratio: response.data.sharpe_ratio ?? 0,
        };
        setBlResult(blData);
        message.success('Black-Litterman计算完成');
      } else {
        const errorMsg = response.message || '计算失败';
        setError(errorMsg);
        message.error(errorMsg);
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '计算请求失败';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    result,
    frontier,
    rpResult,
    blResult,
    loading,
    error,
    optimize,
    calculateFrontier,
    calculateRiskParity,
    calculateBlackLitterman,
    reset,
  };
};
