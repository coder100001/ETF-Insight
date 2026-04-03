// API服务 - 连接Go后端
import type { ETFData, PortfolioAnalysisResult, ExchangeRate, PortfolioConfig } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

// API响应类型
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

// 通用请求函数
async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// ETF相关API
export const etfAPI = {
  // 获取ETF列表
  getList: (market?: string) => {
    const params = market ? `?market=${market}` : '';
    return request<ApiResponse<ETFData[]>>(`/etf/list${params}`);
  },

  // 获取ETF对比数据
  getComparison: (period: string = '1y') => {
    return request<ApiResponse<ETFData[]>>(`/etf/comparison?period=${period}`);
  },

  // 获取投资组合分析
  getPortfolioAnalysis: (allocation: Record<string, number>, totalInvestment: number = 10000, taxRate: number = 0.10) => {
    return request<ApiResponse<PortfolioAnalysisResult>>(`/etf/portfolio`, {
      method: 'POST',
      body: JSON.stringify({ allocation, total_investment: totalInvestment, tax_rate: taxRate }),
    });
  },

  // 获取实时数据
  getRealtimeData: (symbol: string) => {
    return request<ApiResponse<ETFData>>(`/etf/${symbol}/realtime`);
  },

  // 获取指标数据
  getMetrics: (symbol: string, period: string = '1y') => {
    return request<ApiResponse<Record<string, number>>>(`/etf/${symbol}/metrics?period=${period}`);
  },

  // 获取历史数据
  getHistory: (symbol: string, period: string = '1y') => {
    return request<ApiResponse<{ date: string; price: number }[]>>(`/etf/${symbol}/history?period=${period}`);
  },

  // 获取收益预测
  getForecast: (symbol: string, initialInvestment: number = 10000, taxRate: number = 0.10) => {
    return request<ApiResponse<{ years: number; value: number }[]>>(`/etf/${symbol}/forecast?initial_investment=${initialInvestment}&tax_rate=${taxRate}`);
  },

  // 更新实时数据
  updateRealtimeData: () => {
    return request<ApiResponse<{ message: string; count: number }>>(`/etf/update-realtime`, {
      method: 'POST',
    });
  },
};

// 投资组合配置API
export const portfolioAPI = {
  // 获取配置列表
  getConfigs: () => {
    return request<ApiResponse<PortfolioConfig[]>>(`/portfolio-configs/`);
  },

  // 创建配置
  createConfig: (data: {
    name: string;
    description?: string;
    allocation: Record<string, number>;
    total_investment?: number;
    status?: number;
  }) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取配置详情
  getConfigDetail: (id: number) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/${id}`);
  },

  // 更新配置
  updateConfig: (id: number, data: Partial<{
    name: string;
    description: string;
    allocation: Record<string, number>;
    total_investment: number;
    status: number;
  }>) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除配置
  deleteConfig: (id: number) => {
    return request<ApiResponse<{ message: string }>>(`/portfolio-configs/${id}`, {
      method: 'DELETE',
    });
  },

  // 切换状态
  toggleStatus: (id: number) => {
    return request<ApiResponse<PortfolioConfig>>(`/portfolio-configs/${id}/toggle-status`, {
      method: 'POST',
    });
  },

  // 分析配置
  analyzeConfig: (id: number, taxRate: number = 0.10) => {
    return request<ApiResponse<PortfolioAnalysisResult>>(`/portfolio-configs/${id}/analyze`, {
      method: 'POST',
      body: JSON.stringify({ tax_rate: taxRate }),
    });
  },
};

// 汇率API
export const exchangeRateAPI = {
  // 获取汇率列表
  getRates: () => {
    return request<ApiResponse<ExchangeRate[]>>(`/exchange-rates/`);
  },

  // 获取汇率历史
  getHistory: (from: string = 'USD', to: string = 'CNY', days: number = 30) => {
    return request<ApiResponse<{ date: string; rate: number }[]>>(`/exchange-rates/history?from=${from}&to=${to}&days=${days}`);
  },

  // 货币转换
  convert: (from: string, to: string, amount: number) => {
    return request<ApiResponse<{ from: string; to: string; amount: number; result: number }>>(`/exchange-rates/convert?from=${from}&to=${to}&amount=${amount}`);
  },

  // 更新汇率
  updateRates: () => {
    return request<ApiResponse<{ message: string }>>(`/exchange-rates/update`, {
      method: 'POST',
    });
  },
};

// 工作流类型
interface Workflow {
  id: number;
  name: string;
  description: string;
  status: number;
  created_at: string;
  updated_at: string;
}

// 工作流API
export const workflowAPI = {
  // 获取工作流列表
  getWorkflows: () => {
    return request<ApiResponse<Workflow[]>>(`/workflows/`);
  },

  // 创建工作流
  createWorkflow: (data: Omit<Workflow, 'id' | 'created_at' | 'updated_at'>) => {
    return request<ApiResponse<Workflow>>(`/workflows/`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取工作流详情
  getWorkflow: (id: number) => {
    return request<ApiResponse<Workflow>>(`/workflows/${id}`);
  },

  // 更新工作流
  updateWorkflow: (id: number, data: Partial<Omit<Workflow, 'id' | 'created_at' | 'updated_at'>>) => {
    return request<ApiResponse<Workflow>>(`/workflows/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除工作流
  deleteWorkflow: (id: number) => {
    return request<ApiResponse<{ message: string }>>(`/workflows/${id}`, {
      method: 'DELETE',
    });
  },

  // 启动工作流
  startWorkflow: (id: number) => {
    return request<ApiResponse<Workflow>>(`/workflows/${id}/start`, {
      method: 'POST',
    });
  },
};

// 工作流实例类型
interface WorkflowInstance {
  id: number;
  workflow_id: number;
  status: string;
  created_at: string;
  updated_at: string;
}

// 工作流实例API
export const instanceAPI = {
  // 获取实例列表
  getInstances: () => {
    return request<ApiResponse<WorkflowInstance[]>>(`/instances/`);
  },

  // 获取实例详情
  getInstance: (id: number) => {
    return request<ApiResponse<WorkflowInstance>>(`/instances/${id}`);
  },

  // 重试实例
  retryInstance: (id: number) => {
    return request<ApiResponse<{ message: string }>>(`/instances/${id}/retry`, {
      method: 'POST',
    });
  },
};

// 调度器任务类型
interface SchedulerJob {
  id: string;
  name: string;
  schedule: string;
  last_run?: string;
  next_run?: string;
}

// 调度器API
export const schedulerAPI = {
  // 获取定时任务
  getJobs: () => {
    return request<ApiResponse<SchedulerJob[]>>(`/scheduler/jobs`);
  },

  // 立即执行一次
  runOnce: () => {
    return request<ApiResponse<{ message: string }>>(`/scheduler/run-once`, {
      method: 'POST',
    });
  },
};

// 系统统计类型
interface SystemStats {
  etf_count: number;
  workflow_count: number;
  instance_count: number;
  cache_size: number;
}

// 操作日志类型
interface OperationLog {
  id: number;
  operation_type: string;
  operation_name: string;
  status: number;
  created_at: string;
}

// 管理API
export const adminAPI = {
  // 获取统计信息
  getStats: () => {
    return request<ApiResponse<SystemStats>>(`/admin/stats`);
  },

  // 获取日志
  getLogs: () => {
    return request<ApiResponse<OperationLog[]>>(`/admin/logs`);
  },

  // 清除缓存
  clearCache: () => {
    return request<ApiResponse<{ message: string }>>(`/admin/clear-cache`, {
      method: 'POST',
    });
  },
};

// 健康检查
export const healthCheck = () => {
  return request<{ status: string; message: string }>(`/health`);
};
