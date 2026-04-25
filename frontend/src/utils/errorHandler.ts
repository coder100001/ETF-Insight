// 统一错误处理工具

export class APIError extends Error {
  statusCode: number;
  originalError?: unknown;

  constructor(
    message: string,
    statusCode: number = 500,
    originalError?: unknown
  ) {
    super(message);
    this.name = 'APIError';
    this.statusCode = statusCode;
    this.originalError = originalError;
  }
}

export class ValidationError extends Error {
  field?: string;

  constructor(
    message: string,
    field?: string
  ) {
    super(message);
    this.name = 'ValidationError';
    this.field = field;
  }
}

export class NetworkError extends Error {
  constructor(message: string = '网络连接失败') {
    super(message);
    this.name = 'NetworkError';
  }
}

/**
 * 统一处理API错误
 * @param error - 错误对象
 * @returns 用户友好的错误信息
 */
export const handleAPIError = (error: unknown): string => {
  if (error instanceof APIError) {
    return `API错误 (${error.statusCode}): ${error.message}`;
  }
  if (error instanceof ValidationError) {
    return `验证错误: ${error.message}`;
  }
  if (error instanceof NetworkError) {
    return `网络错误: ${error.message}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return '未知错误';
};

/**
 * 从API响应中提取错误信息
 * @param response - API响应对象
 * @returns 错误信息
 */
export const extractErrorMessage = (response: { message?: string; error?: string }): string => {
  return response.message || response.error || '请求失败';
};

/**
 * 安全地执行异步操作
 * @param fn - 异步函数
 * @param fallbackValue - 失败时的默认值
 * @returns 执行结果或默认值
 */
export const safeExecute = async <T>(
  fn: () => Promise<T>,
  fallbackValue: T,
  errorHandler?: (error: unknown) => void
): Promise<T> => {
  try {
    return await fn();
  } catch (error) {
    if (errorHandler) {
      errorHandler(error);
    }
    return fallbackValue;
  }
};

/**
 * 验证ETF统计数据
 * @param stats - ETF统计数据
 * @returns 是否通过验证
 */
export const validateETFStatistics = (stats: {
  annualized: number;
  volatility: number;
  sharpe_ratio: number;
  max_drawdown: number;
}): boolean => {
  if (stats.annualized < -0.5 || stats.annualized > 1.0) return false;
  if (stats.volatility < 0.001 || stats.volatility > 1.0) return false;
  if (stats.sharpe_ratio < -5.0 || stats.sharpe_ratio > 5.0) return false;
  if (stats.max_drawdown < 0 || stats.max_drawdown > 1.0) return false;
  return true;
};

/**
 * 验证投资组合权重
 * @param weights - 权重对象
 * @returns 是否通过验证
 */
export const validateWeights = (weights: Record<string, number>): boolean => {
  const values = Object.values(weights);
  const total = values.reduce((sum, w) => sum + w, 0);
  return Math.abs(total - 1.0) < 0.001 && values.every(w => w >= 0);
};
