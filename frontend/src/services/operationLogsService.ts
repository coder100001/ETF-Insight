import { request } from './api';

// API响应类型
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

// 分页响应结构
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 日志列表筛选参数
export interface LogFilterParams {
  page: number;
  pageSize: number;
  startTime?: string;
  endTime?: string;
  user?: string;
  actionType?: string;
  status?: 'success' | 'failure';
  logType?: 'audit' | 'operation';
  search?: string;
}

// 统一日志类型
export interface UnifiedLog {
  id: number;
  log_type: 'audit' | 'operation';
  timestamp: string;
  user: string;
  module: string;
  action_type: string;
  details: string;
  ip: string;
  status: 'success' | 'failure';
  status_code: number;
  error_message: string;
  duration_ms: number;
}

// 日志列表响应元数据
export interface LogsMeta {
  pagination: PaginationMeta;
  summary: LogsSummary;
}

// 分页元数据
export interface PaginationMeta {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// 日志摘要信息
export interface LogsSummary {
  total_logs: number;
  total_audit: number;
  total_operation: number;
}

// 日志列表响应
export interface GetLogsResponse extends ApiResponse<UnifiedLog[]> {
  meta: LogsMeta;
}

// 日志详情响应
export interface GetLogDetailResponse extends ApiResponse<UnifiedLog> {
  data?: UnifiedLog;
}

// 日志类型列表响应
export interface LogTypesResponse {
  types: Array<{ value: string; label: string; count: number }>;
}

// 操作类型列表响应
export interface ActionTypesResponse {
  types: Array<{ value: string; label: string; count: number }>;
}

// 用户列表响应
export interface UsersResponse {
  users: string[];
}

// 导出日志请求参数
export interface ExportLogsRequest {
  startTime?: string;
  endTime?: string;
  user?: string;
  actionType?: string;
  status?: 'success' | 'failure';
  logType?: 'audit' | 'operation';
}

/**
 * 操作日志API服务
 */
export const operationLogsAPI = {
  /**
   * 获取日志列表
   * @param params 筛选参数
   * @returns 日志列表响应
   */
  getLogs: async (params: LogFilterParams): Promise<{ data: UnifiedLog[]; total: number }> => {
    const queryParams = new URLSearchParams();
    queryParams.append('page', params.page.toString());
    queryParams.append('page_size', params.pageSize.toString());

    if (params.startTime) queryParams.append('start_time', params.startTime);
    if (params.endTime) queryParams.append('end_time', params.endTime);
    if (params.user) queryParams.append('user', params.user);
    if (params.actionType) queryParams.append('action_type', params.actionType);
    if (params.status) queryParams.append('status', params.status);
    if (params.logType) queryParams.append('log_type', params.logType);
    if (params.search) queryParams.append('search', params.search);

    const response = await request<{
      success: boolean;
      data: UnifiedLog[];
      meta: {
        pagination: { page: number; page_size: number; total: number; total_pages: number };
        summary: { total_logs: number; total_audit: number; total_operation: number };
      };
    }>(`/logs?${queryParams.toString()}`);

    if (!response.success) {
      throw new Error('获取日志失败');
    }

    return {
      data: response.data || [],
      total: response.meta?.pagination?.total || 0,
    };
  },

  /**
   * 获取日志详情
   * @param type 日志类型 (audit/operation)
   * @param id 日志ID
   * @returns 日志详情响应
   */
  getLogDetail: async (type: string, id: number): Promise<UnifiedLog> => {
    const response = await request<GetLogDetailResponse>(`/logs/${type}/${id}`);

    if (!response.success || !response.data) {
      throw new Error(response.message || response.error || '获取日志详情失败');
    }

    return response.data;
  },

  /**
   * 获取日志类型列表
   * @returns 日志类型列表响应
   */
  getLogTypes: async (): Promise<LogTypesResponse> => {
    const response = await request<{
      success: boolean;
      types?: { audit: number; operation: number };
    }>('/logs/types');

    if (!response.success) {
      throw new Error('获取日志类型失败');
    }

    const types = response.types || { audit: 0, operation: 0 };
    return {
      types: [
        { value: 'audit', label: 'API日志', count: types.audit },
        { value: 'operation', label: '操作日志', count: types.operation },
      ],
    };
  },

  /**
   * 获取操作类型列表
   * @returns 操作类型列表响应
   */
  getActionTypes: async (): Promise<ActionTypesResponse> => {
    const response = await request<{ success: boolean; action_types?: string[] }>('/logs/action-types');

    if (!response.success) {
      throw new Error('获取操作类型失败');
    }

    return {
      types: (response.action_types || []).map((t) => ({
        value: t,
        label: t,
        count: 0,
      })),
    };
  },

  /**
   * 获取用户列表
   * @returns 用户列表响应
   */
  getUsers: async (): Promise<UsersResponse> => {
    const response = await request<{ success: boolean; users?: string[] }>('/logs/users');

    if (!response.success) {
      throw new Error('获取用户列表失败');
    }

    return { users: response.users || [] };
  },

  /**
   * 导出日志到Excel
   * @param params 导出筛选参数
   * @returns 文件下载响应
   */
  exportLogs: async (params: ExportLogsRequest): Promise<Blob> => {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'}/logs/export`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(params),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.blob();
  },
};

// 辅助函数：格式化日志时间
export function formatLogTime(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

// 辅助函数：获取状态标签样式
export function getStatusTagStyle(status: 'success' | 'failure'): {
  color: string;
  backgroundColor: string;
} {
  if (status === 'success') {
    return {
      color: '#52c41a',
      backgroundColor: '#f6ffed',
    };
  } else {
    return {
      color: '#ff4d4f',
      backgroundColor: '#fff2f0',
    };
  }
}

// 辅助函数：获取日志类型标签样式
export function getLogTypeTagStyle(logType: 'audit' | 'operation'): {
  color: string;
  backgroundColor: string;
} {
  if (logType === 'audit') {
    return {
      color: '#1890ff',
      backgroundColor: '#e6f7ff',
    };
  } else {
    return {
      color: '#722ed1',
      backgroundColor: '#f9f0ff',
    };
  }
}
