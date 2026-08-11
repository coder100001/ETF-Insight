import { create } from 'zustand';
import { etfAPI, exchangeRateAPI } from '../services/api';
import type { ETFData, ETFStatistics } from '../types';

// ===== 默认数据 =====
const DEFAULT_ETFS: ETFData[] = [
  { symbol: 'VTI', name: 'Vanguard Total Stock Market', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'VOO', name: 'Vanguard S&P 500', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'QQQ', name: 'Invesco QQQ Trust', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'IWM', name: 'iShares Russell 2000', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'EFA', name: 'iShares MSCI EAFE', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'EEM', name: 'iShares Emerging Markets', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'AGG', name: 'iShares Core U.S. Aggregate Bond', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'TLT', name: 'iShares 20+ Year Treasury Bond', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'GLD', name: 'SPDR Gold Shares', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
  { symbol: 'VNQ', name: 'Vanguard Real Estate', current_price: 0, previous_close: 0, change: 0, change_percent: 0, open_price: 0, high_price: 0, low_price: 0, volume: 0 },
];

// ===== 数据验证 =====
const validateETFStatistics = (stats: ETFStatistics): boolean => {
  if (stats.annualized < -0.5 || stats.annualized > 1.0) return false;
  if (stats.volatility < 0.001 || stats.volatility > 1.0) return false;
  if (stats.sharpe_ratio < -5.0 || stats.sharpe_ratio > 5.0) return false;
  if (stats.max_drawdown < 0 || stats.max_drawdown > 1.0) return false;
  return true;
};

// 从 ETF 列表派生统计数据（后端 /etf/list 已含 volatility/sharpe 等指标，
// 无需额外统计端点；数据缺失的 symbol 保留为 null 供前端显示"-"）
type StatsMap = Record<string, ETFStatistics | null>;
const deriveStatistics = (etfs: ETFData[], symbols?: string[]): StatsMap => {
  const result: StatsMap = {};
  const wanted = symbols ? new Set(symbols) : null;
  etfs.forEach(etf => {
    if (wanted && !wanted.has(etf.symbol)) return;
    const candidate: ETFStatistics = {
      symbol: etf.symbol,
      name: etf.name,
      annualized: etf.total_return || 0,
      volatility: etf.volatility || 0,
      sharpe_ratio: etf.sharpe_ratio || 0,
      max_drawdown: etf.max_drawdown || 0,
    };
    // 数据缺失（全部为 0）时标记为 null，前端显示"-"，不展示虚假的 0%
    const hasData = etf.volatility !== undefined && etf.volatility > 0;
    result[etf.symbol] = hasData && validateETFStatistics(candidate) ? candidate : null;
  });
  return result;
};

// ===== Store 状态类型 =====
interface ETFStoreState {
  // 数据
  etfList: ETFData[];
  etfStatistics: StatsMap;
  exchangeRates: { from_currency: string; to_currency: string; rate: number; updated_at: string }[];

  // 加载状态
  loading: boolean;
  statsLoading: boolean;
  error: string | null;

  // 初始化标记（成功才置位；失败复位以允许重试）
  hasInitialized: boolean;

  // Actions
  // force=true 时取消进行中的请求并重新获取（用户手动刷新）；默认共享进行中请求（防重复）
  refreshETFList: (force?: boolean) => Promise<void>;
  refreshStatistics: (symbols: string[]) => Promise<void>;
  fetchExchangeRates: () => Promise<void>;
  initialize: () => void;
}

// ===== 模块级请求状态 =====
let listAbortController: AbortController | null = null;
let listRequest: Promise<void> | null = null;
let ratesRequest: Promise<void> | null = null;
let ratesFetchedAt = 0;
// 汇率缓存有效期：30 分钟（汇率由后端定时任务更新，无需每次实时请求）
const RATES_TTL = 30 * 60 * 1000;

export const useETFStore = create<ETFStoreState>((set, get) => ({
  etfList: DEFAULT_ETFS,
  etfStatistics: {},
  exchangeRates: [],
  loading: false,
  statsLoading: false,
  error: null,
  hasInitialized: false,

  initialize: () => {
    if (get().hasInitialized) return;
    set({ hasInitialized: true });
    get().refreshETFList();
    get().fetchExchangeRates(); // 一次性获取汇率（缓存到内存，Dashboard 直接读取，不再每次请求）
  },

  refreshETFList: (force = false) => {
    // 默认共享进行中的请求（防并发重复）；force（用户手动刷新）时取消旧请求重新获取
    if (!force && listRequest) return listRequest;

    if (listAbortController) listAbortController.abort();
    const controller = new AbortController();
    listAbortController = controller;

    set({ loading: true, error: null });

    listRequest = (async () => {
      try {
        const response = await etfAPI.getList();

        if (controller.signal.aborted) return;
        if (response.success && response.data && response.data.length > 0) {
          // 统计数据由列表派生，无需额外请求
          set({
            etfList: response.data,
            etfStatistics: deriveStatistics(response.data),
            loading: false,
            hasInitialized: true,
          });
        } else {
          set({ etfList: DEFAULT_ETFS, etfStatistics: {}, loading: false, error: 'ETF 列表为空' });
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') return;
        console.log('获取ETF列表失败，使用默认列表');
        // 失败复位初始化标记，下次挂载/刷新可重试，避免永久固化假数据
        set({ etfList: DEFAULT_ETFS, etfStatistics: {}, loading: false, error: '获取ETF列表失败', hasInitialized: false });
      } finally {
        listRequest = null;
      }
    })();

    return listRequest;
  },

  refreshStatistics: (symbols: string[]) => {
    // 统计数据直接从当前 ETF 列表派生（不再调用后端统计端点）
    const list = get().etfList;
    set({ etfStatistics: deriveStatistics(list, symbols), statsLoading: false });
    return Promise.resolve();
  },

  fetchExchangeRates: () => {
    if (ratesRequest) return ratesRequest;
    // TTL 内且已有数据则复用缓存
    if (get().exchangeRates.length > 0 && ratesFetchedAt && Date.now() - ratesFetchedAt < RATES_TTL) {
      return Promise.resolve();
    }
    ratesRequest = (async () => {
      try {
        const response = await exchangeRateAPI.getAll();
        if (response.success && response.data) {
          set({ exchangeRates: response.data });
          ratesFetchedAt = Date.now();
        }
      } catch {
        // 静默失败：不置 ratesFetchedAt，下次 initialize/刷新可重试
      } finally {
        ratesRequest = null;
      }
    })();
    return ratesRequest;
  },
}));
