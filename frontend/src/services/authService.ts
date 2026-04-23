import { authRequest, request } from './api';

// API响应类型
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

// 登录请求
export interface LoginRequest {
  username: string;
  password: string;
}

// 登录响应
export interface LoginResponse {
  token: string;
  user: {
    id: string;
    username: string;
    role: string;
    permissions: string[];
  };
  expires_at: string;
}

// 用户信息
export interface UserInfo {
  id: string;
  username: string;
  role: string;
  permissions: string[];
}

// Token存储键
const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';
const TOKEN_EXPIRY_KEY = 'token_expiry';

/**
 * 认证服务
 */
export const authAPI = {
  /**
   * 用户登录
   * @param credentials 登录凭证
   * @returns 登录响应
   */
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await request<ApiResponse<LoginResponse>>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    }, {
      // 登录失败需要保留后端错误，不应触发全局 401 重定向。
      handleUnauthorized: false,
      maxRetries: 0,
    });

    if (!response.success || !response.data) {
      throw new Error(response.message || response.error || '登录失败');
    }

    // 保存Token和用户信息
    const { token, user, expires_at } = response.data;
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
    localStorage.setItem(TOKEN_EXPIRY_KEY, expires_at);

    return response.data;
  },

  /**
   * 用户登出
   */
  logout: async (): Promise<void> => {
    try {
      await request<ApiResponse<void>>('/auth/logout', {
        method: 'POST',
      });
    } catch {
      // 忽略登出API错误
    } finally {
      // 清除本地存储
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      localStorage.removeItem(TOKEN_EXPIRY_KEY);
    }
  },

  /**
   * 刷新Token
   * @returns 新Token
   */
  refreshToken: async (): Promise<string> => {
    const response = await authRequest<ApiResponse<{ token: string; expires_at: string }>>('/auth/refresh', {
      method: 'POST',
    });

    if (!response.success || !response.data) {
      // Token刷新失败，清除本地存储
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      localStorage.removeItem(TOKEN_EXPIRY_KEY);
      throw new Error('Token刷新失败');
    }

    localStorage.setItem(TOKEN_KEY, response.data.token);
    localStorage.setItem(TOKEN_EXPIRY_KEY, response.data.expires_at);

    return response.data.token;
  },

  /**
   * 获取当前用户信息
   * @returns 用户信息
   */
  getCurrentUser: (): UserInfo | null => {
    const userStr = localStorage.getItem(USER_KEY);
    if (!userStr) return null;

    try {
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  },

  /**
   * 获取Token
   * @returns Token字符串
   */
  getToken: (): string | null => {
    return localStorage.getItem(TOKEN_KEY);
  },

  /**
   * 检查是否已登录
   * @returns 是否已登录
   */
  isAuthenticated: (): boolean => {
    const token = localStorage.getItem(TOKEN_KEY);
    const expiry = localStorage.getItem(TOKEN_EXPIRY_KEY);

    if (!token || !expiry) return false;

    // 检查Token是否过期
    const expiryDate = new Date(expiry);
    return expiryDate > new Date();
  },

  /**
   * 检查是否有指定权限
   * @param permission 权限名称
   * @returns 是否有权限
   */
  hasPermission: (permission: string): boolean => {
    const user = authAPI.getCurrentUser();
    if (!user) return false;

    // 管理员拥有所有权限
    if (user.role === 'admin') return true;

    return user.permissions.includes(permission);
  },

  /**
   * 检查是否有指定角色
   * @param roles 角色列表
   * @returns 是否有角色
   */
  hasRole: (...roles: string[]): boolean => {
    const user = authAPI.getCurrentUser();
    if (!user) return false;

    return roles.includes(user.role);
  },
};

/**
 * 检查Token是否即将过期（提前5分钟刷新）
 * @returns 是否需要刷新
 */
export function shouldRefreshToken(): boolean {
  const expiry = localStorage.getItem(TOKEN_EXPIRY_KEY);
  if (!expiry) return false;

  const expiryDate = new Date(expiry);
  const now = new Date();
  const fiveMinutes = 5 * 60 * 1000;

  return expiryDate.getTime() - now.getTime() < fiveMinutes;
}
