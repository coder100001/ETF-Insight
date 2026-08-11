import { useEffect, useMemo } from 'react';
import { useETFStore } from '../stores/etfStore';
import type { ETFInfo, ETFStatistics } from '../types';

interface UseETFDataReturn {
  availableETFs: ETFInfo[];
  etfStatistics: Record<string, ETFStatistics | null>;
  loading: boolean;
  statsLoading: boolean;
  error: string | null;
  refreshETFList: (force?: boolean) => Promise<void>;
  refreshStatistics: (symbols: string[]) => Promise<void>;
}

/**
 * useETFData — ETF 列表与统计数据 Hook
 *
 * 底层使用 Zustand store (useETFStore) 实现全局共享状态：
 * - 多个组件使用此 Hook 不会重复请求（请求去重）
 * - 数据在页面间共享，切换路由不丢失
 * - 保持原有 API 不变（availableETFs 返回 ETFInfo[]），现有组件无需修改
 */
export const useETFData = (): UseETFDataReturn => {
  const etfList = useETFStore(state => state.etfList);
  const etfStatistics = useETFStore(state => state.etfStatistics);
  const loading = useETFStore(state => state.loading);
  const statsLoading = useETFStore(state => state.statsLoading);
  const error = useETFStore(state => state.error);
  const refreshETFList = useETFStore(state => state.refreshETFList);
  const refreshStatistics = useETFStore(state => state.refreshStatistics);
  const initialize = useETFStore(state => state.initialize);

  // 将 ETFData[] 映射为 ETFInfo[] 以保持向后兼容
  const availableETFs = useMemo<ETFInfo[]>(
    () => etfList.map(etf => ({ symbol: etf.symbol, name: etf.name })),
    [etfList]
  );

  // 组件挂载时初始化（仅首次有效）
  useEffect(() => {
    initialize();
  }, [initialize]);

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
