import axios from 'axios';
import type { 
  ETFData, 
  PortfolioResult, 
  UserConfig, 
  RealtimeUpdateResponse,
  ExchangeRate,
  ApiResponse 
} from '../types';

// API基础配置 - 使用相对路径，让Django处理路由
const API_BASE_URL = '/api/workflow';

// 从Django cookie获取CSRF token
const getCSRFToken = (): string => {
  // 从cookie获取csrftoken
  const match = document.cookie.match(/csrftoken=([^;]+)/);
  return match ? match[1] : '';
};

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加CSRF Token
apiClient.interceptors.request.use((config) => {
  const csrfToken = getCSRFToken();
  if (csrfToken) {
    config.headers['X-CSRFToken'] = csrfToken;
  }
  return config;
});

// 响应拦截器 - 错误处理
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

// ETF相关API
export const etfApi = {
  // 获取所有ETF列表
  getETFList: async (): Promise<ETFData[]> => {
    const response = await apiClient.get('/etfs/');
    return response.data;
  },

  // 获取单个ETF详情
  getETFDetail: async (symbol: string): Promise<ETFData> => {
    const response = await apiClient.get(`/etfs/${symbol}/`);
    return response.data;
  },

  // 获取ETF历史数据
  getETFHistory: async (symbol: string, period: string = '1y'): Promise<any[]> => {
    const response = await apiClient.get(`/etfs/${symbol}/history/?period=${period}`);
    return response.data;
  },

  // 对比多个ETF
  compareETFs: async (symbols: string[]): Promise<ETFData[]> => {
    const response = await apiClient.post('/etfs/compare/', { symbols });
    return response.data;
  },
};

// 投资组合相关API
export const portfolioApi = {
  // 分析投资组合
  analyzePortfolio: async (config: UserConfig): Promise<PortfolioResult> => {
    const response = await apiClient.post('/portfolio/analyze/', config);
    return response.data;
  },

  // 预测投资组合收益
  forecastPortfolio: async (
    config: UserConfig, 
    years: number[] = [3, 5, 10]
  ): Promise<any> => {
    const response = await apiClient.post('/portfolio/forecast/', {
      ...config,
      forecast_years: years,
    });
    return response.data;
  },

  // 更新实时数据
  updateRealtime: async (config: UserConfig): Promise<RealtimeUpdateResponse> => {
    const response = await apiClient.post('/update-realtime/', {
      allocation: config.allocation,
      total_investment: config.total_investment,
    });
    return response.data;
  },
};

// 汇率相关API
export const exchangeRateApi = {
  // 获取所有汇率
  getRates: async (): Promise<ExchangeRate[]> => {
    const response = await apiClient.get('/exchange-rates/');
    return response.data;
  },

  // 更新汇率
  updateRates: async (): Promise<ApiResponse<ExchangeRate[]>> => {
    const response = await apiClient.post('/update-exchange-rates/');
    return response.data;
  },
};

// 配置相关API
export const configApi = {
  // 获取用户配置
  getConfig: async (): Promise<UserConfig> => {
    const response = await apiClient.get('/config/');
    return response.data;
  },

  // 保存用户配置
  saveConfig: async (config: UserConfig): Promise<ApiResponse<UserConfig>> => {
    const response = await apiClient.post('/config/', config);
    return response.data;
  },

  // 获取ETF配置列表
  getETFConfigs: async (): Promise<any[]> => {
    const response = await apiClient.get('/etf-configs/');
    return response.data;
  },
};

// 工具函数
export const formatPrice = (price: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(price);
};

export const formatPercent = (value: number, decimals: number = 2): string => {
  const sign = value >= 0 ? '+' : '';
  return `${sign}${value.toFixed(decimals)}%`;
};

export const formatNumber = (value: number, decimals: number = 0): string => {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value);
};

export const formatVolume = (volume: number): string => {
  if (volume >= 1000000) {
    return `${(volume / 1000000).toFixed(2)}M`;
  } else if (volume >= 1000) {
    return `${(volume / 1000).toFixed(2)}K`;
  }
  return volume.toString();
};

export default apiClient;