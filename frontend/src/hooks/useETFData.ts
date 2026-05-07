import { useState, useEffect, useCallback, useRef } from 'react';
import { etfAPI, optimizationAPI } from '../services/api';
import type { ETFInfo, ETFStatistics } from '../types';

interface UseETFDataReturn {
  availableETFs: ETFInfo[];
  etfStatistics: Record<string, ETFStatistics>;
  loading: boolean;
  statsLoading: boolean;
  error: string | null;
  refreshETFList: () => Promise<void>;
  refreshStatistics: (symbols: string[]) => Promise<void>;
}

const DEFAULT_ETFS: ETFInfo[] = [
  { symbol: 'VTI', name: 'Vanguard Total Stock Market' },
  { symbol: 'VOO', name: 'Vanguard S&P 500' },
  { symbol: 'QQQ', name: 'Invesco QQQ Trust' },
  { symbol: 'IWM', name: 'iShares Russell 2000' },
  { symbol: 'EFA', name: 'iShares MSCI EAFE' },
  { symbol: 'EEM', name: 'iShares Emerging Markets' },
  { symbol: 'AGG', name: 'iShares Core U.S. Aggregate Bond' },
  { symbol: 'TLT', name: 'iShares 20+ Year Treasury Bond' },
  { symbol: 'GLD', name: 'SPDR Gold Shares' },
  { symbol: 'VNQ', name: 'Vanguard Real Estate' },
];

const validateETFStatistics = (stats: ETFStatistics): boolean => {
  if (stats.annualized < -0.5 || stats.annualized > 1.0) return false;
  if (stats.volatility < 0.001 || stats.volatility > 1.0) return false;
  if (stats.sharpe_ratio < -5.0 || stats.sharpe_ratio > 5.0) return false;
  if (stats.max_drawdown < 0 || stats.max_drawdown > 1.0) return false;
  return true;
};

export const useETFData = (): UseETFDataReturn => {
  const [availableETFs, setAvailableETFs] = useState<ETFInfo[]>(DEFAULT_ETFS);
  const [etfStatistics, setEtfStatistics] = useState<Record<string, ETFStatistics>>({});
  const [loading, setLoading] = useState(false);
  const [statsLoading, setStatsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isMountedRef = useRef(true);
  const abortControllerRef = useRef<AbortController | null>(null);
  const statsAbortControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      abortControllerRef.current?.abort();
      statsAbortControllerRef.current?.abort();
    };
  }, []);

  const refreshETFList = useCallback(async () => {
    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();

    setLoading(true);
    setError(null);
    try {
      const response = await etfAPI.getList();
      if (!isMountedRef.current) return;

      if (response.success && response.data && response.data.length > 0) {
        const etfList = response.data.map(etf => ({
          symbol: etf.symbol,
          name: etf.name,
        }));
        setAvailableETFs(etfList);
      } else {
        setAvailableETFs(DEFAULT_ETFS);
      }
    } catch (err) {
      if (!isMountedRef.current) return;
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }
      console.log('获取ETF列表失败，使用默认列表');
      setAvailableETFs(DEFAULT_ETFS);
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  }, []);

  const refreshStatistics = useCallback(async (symbols: string[]) => {
    if (symbols.length === 0) return;

    statsAbortControllerRef.current?.abort();
    statsAbortControllerRef.current = new AbortController();

    setStatsLoading(true);
    try {
      const response = await optimizationAPI.etfStatistics(symbols);
      if (!isMountedRef.current) return;

      if (response.success && response.data) {
        const validatedData: Record<string, ETFStatistics> = {};
        Object.entries(response.data).forEach(([symbol, stats]) => {
          const etfStats: ETFStatistics = {
            symbol: symbol,
            name: symbol,
            annualized: stats.mean_return || 0,
            volatility: stats.volatility || 0,
            sharpe_ratio: stats.sharpe_ratio || 0,
            max_drawdown: stats.max_drawdown || 0,
          };
          if (validateETFStatistics(etfStats)) {
            validatedData[symbol] = etfStats;
          } else {
            console.warn(`ETF ${symbol} 的数据未通过验证`);
          }
        });
        setEtfStatistics(validatedData);
      }
    } catch (err) {
      if (!isMountedRef.current) return;
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }
      console.log('获取ETF统计数据失败');
    } finally {
      if (isMountedRef.current) {
        setStatsLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    refreshETFList();
  }, [refreshETFList]);

  useEffect(() => {
    if (availableETFs.length > 0) {
      const symbols = availableETFs.map(etf => etf.symbol);
      refreshStatistics(symbols);
    }
  }, [availableETFs, refreshStatistics]);

  return {
    availableETFs,
    etfStatistics,
    loading,
    statsLoading,
    error,
    refreshETFList,
    refreshStatistics,
  };
};
