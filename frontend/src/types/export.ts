// 导出格式类型
export type ExportFormat = 'html' | 'pdf' | 'excel' | 'markdown';

// 导出请求接口
export interface ExportRequest {
  format: ExportFormat;
  title?: string;
  data: Record<string, unknown>;
}

// 导出响应接口
export interface ExportResponse {
  success: boolean;
  data?: ExportData;
  error?: string;
}

// 导出数据接口
export interface ExportData {
  content: string;      // base64编码的内容
  filename: string;     // 文件名
  mime_type: string;    // MIME类型
  file_size: number;    // 文件大小（字节）
}

// 导出元数据接口
export interface ExportMetadata {
  user_id: string;
  username: string;
  page_type: string;
  format: ExportFormat;
  title: string;
  data_size: number;
  timestamp: string;
}

// 支持格式响应接口
export interface ExportSupportedFormatsResponse {
  success: boolean;
  data?: ExportFormat[];
  error?: string;
}

// 支持类型响应接口
export interface ExportSupportedTypesResponse {
  success: boolean;
  data?: string[];
  error?: string;
}

// 导出错误类型
export interface ExportError {
  code: string;
  message: string;
  details?: string;
}

// 导出状态枚举
export enum ExportStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCESS = 'success',
  ERROR = 'error'
}

// 导出配置接口
export interface ExportConfig {
  maxDataSize: number;      // 最大数据大小（字节）
  timeout: number;          // 超时时间（毫秒）
  supportedFormats: ExportFormat[];
  defaultFormat: ExportFormat;
}

// 默认导出配置
export const DEFAULT_EXPORT_CONFIG: ExportConfig = {
  maxDataSize: 1024 * 1024, // 1MB
  timeout: 30000,           // 30秒
  supportedFormats: ['html', 'pdf', 'excel', 'markdown'],
  defaultFormat: 'html'
};
