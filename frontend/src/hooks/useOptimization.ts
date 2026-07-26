import { useState, useCallback, useRef, useEffect } from 'react';
import { message } from 'antd';
import { optimizationAPI } from '../services/api';
import type { OptimizationResult, EfficientFrontierPoint } from '../types';

interface UseOptimizationReturn {
  result: OptimizationResult | null;
  frontier: EfficientFrontierPoint[];
  loading: boolean;
  error: string | null;
  optimize: (params: OptimizeParams) => Promise<void>;
  calculateFrontier: (symbols: string[], points?: number) => Promise<void>;
  reset: () => void;
}

export interface OptimizeParams {
  symbols: string[];
  objective: 'min_volatility' | 'max_sharpe' | 'target_return';
  targetReturn?: number;
  riskFreeRate?: number;
}

export const useOptimization = (): UseOptimizationReturn => {
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [frontier, setFrontier] = useState<EfficientFrontierPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isMountedRef = useRef(true);
  const abortControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      abortControllerRef.current?.abort();
    };
  }, []);

  const cancelPreviousRequest = useCallback(() => {
    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();
  }, []);

  const reset = useCallback(() => {
    setResult(null);
    setFrontier([]);
    setError(null);
  }, []);

  const optimize = useCallback(async (params: OptimizeParams) => {
    if (params.symbols.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    cancelPreviousRequest();
    setLoading(true);
    setError(null);
    try {
      const response = await optimizationAPI.mptOptimize(
        params.symbols,
        params.objective,
        params.targetReturn,
        params.riskFreeRate
      );

      if (!isMountedRef.current) return;

      if (response.success && response.data) {
        const optResult: OptimizationResult = {
          weights: response.data.weights ?? {},
          expected_return: response.data.expected_return,
          expected_risk: response.data.volatility,
          volatility: response.data.volatility,
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
      if (!isMountedRef.current) return;
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }
      const errorMsg = err instanceof Error ? err.message : '优化请求失败';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  }, [cancelPreviousRequest]);

  const calculateFrontier = useCallback(async (symbols: string[], points: number = 20) => {
    if (symbols.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    cancelPreviousRequest();
    setLoading(true);
    setError(null);
    try {
      const response = await optimizationAPI.efficientFrontier(symbols, points);

      if (!isMountedRef.current) return;

      if (response.success && response.data) {
        const frontierData: EfficientFrontierPoint[] = response.data.map(point => ({
          target_return: point.target_return,
          min_volatility: point.min_volatility,
          optimal_weights: point.optimal_weights,
          sharpe_ratio: point.sharpe_ratio,
        }));
        setFrontier(frontierData);
        message.success('有效前沿计算完成');
      } else {
        const errorMsg = response.message || '计算失败';
        setError(errorMsg);
        message.error(errorMsg);
      }
    } catch (err) {
      if (!isMountedRef.current) return;
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }
      const errorMsg = err instanceof Error ? err.message : '计算请求失败';
      setError(errorMsg);
      message.error(errorMsg);
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  }, [cancelPreviousRequest]);

  return {
    result,
    frontier,
    loading,
    error,
    optimize,
    calculateFrontier,
    reset,
  };
};
