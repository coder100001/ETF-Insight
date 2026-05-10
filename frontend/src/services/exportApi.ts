import type {
  ExportFormat,
  ExportRequest,
  ExportResponse,
  ExportData,
  ExportSupportedFormatsResponse,
  ExportSupportedTypesResponse
} from '../types/export';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

// 导出API服务
export const exportApi = {
  // 导出数据
  async exportData(pageType: string, request: ExportRequest): Promise<ExportResponse> {
    const response = await fetch(`${API_BASE_URL}/export/${pageType}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  },

  // 获取支持的格式
  async getSupportedFormats(): Promise<ExportFormat[]> {
    const response = await fetch(`${API_BASE_URL}/export/formats`);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    const result: ExportSupportedFormatsResponse = await response.json();
    return result.data || [];
  },

  // 获取支持的数据类型
  async getSupportedTypes(): Promise<string[]> {
    const response = await fetch(`${API_BASE_URL}/export/types`);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    const result: ExportSupportedTypesResponse = await response.json();
    return result.data || [];
  },

  // 下载导出文件
  downloadExportFile(exportData: ExportData): void {
    const byteCharacters = atob(exportData.content);
    const byteNumbers = new Array(byteCharacters.length);
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i);
    }
    const byteArray = new Uint8Array(byteNumbers);
    const blob = new Blob([byteArray], { type: exportData.mime_type });

    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = exportData.filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  },

  // 导出并下载
  async exportAndDownload(pageType: string, request: ExportRequest): Promise<void> {
    const response = await this.exportData(pageType, request);

    if (!response.success || !response.data) {
      throw new Error(response.error || '导出失败');
    }

    this.downloadExportFile(response.data);
  }
};

// 默认导出
export default exportApi;
