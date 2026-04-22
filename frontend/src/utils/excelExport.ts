import * as XLSX from 'xlsx';
import { UnifiedLog } from '../services/operationLogsService';
import { formatLogTime } from '../services/operationLogsService';

/**
 * 导出日志数据到Excel
 * @param logs 日志数据数组
 * @param filename 导出的文件名（不含扩展名）
 */
export function exportLogsToExcel(logs: UnifiedLog[], filename = '操作日志'): void {
  if (!logs || logs.length === 0) {
    console.warn('没有日志数据可导出');
    return;
  }

  try {
    // 准备数据 - 只导出关键字段
    const exportData = logs.map((log) => ({
      '时间': formatLogTime(log.timestamp),
      '用户': log.user || '系统',
      '日志类型': log.log_type === 'audit' ? 'API日志' : '操作日志',
      '操作类型': log.action_type,
      '操作模块': log.module,
      '状态': log.status === 'success' ? '成功' : '失败',
      'IP地址': log.ip || '',
      '状态码': log.status_code || '',
      '耗时(ms)': log.duration_ms || '',
      '详情': log.details || '',
      '错误信息': log.error_message || '',
    }));

    // 创建工作簿和工作表
    const workbook = XLSX.utils.book_new();
    const worksheet = XLSX.utils.json_to_sheet(exportData);

    // 设置列宽
    const colWidths = [
      { wch: 20 }, // 时间
      { wch: 10 }, // 用户
      { wch: 10 }, // 日志类型
      { wch: 15 }, // 操作类型
      { wch: 20 }, // 操作模块
      { wch: 8 },  // 状态
      { wch: 15 }, // IP地址
      { wch: 10 }, // 状态码
      { wch: 10 }, // 耗时
      { wch: 40 }, // 详情
      { wch: 40 }, // 错误信息
    ];
    worksheet['!cols'] = colWidths;

    // 添加工作表到工作簿
    XLSX.utils.book_append_sheet(workbook, worksheet, '操作日志');

    // 添加标题行样式（可选）
    const headerRow = Object.keys(exportData[0]);
    const range = XLSX.utils.decode_range(worksheet['!ref'] || 'A1:A1');
    for (let C = range.s.c; C <= range.e.c; ++C) {
      const cellAddress = XLSX.utils.encode_cell({ r: 0, c: C });
      if (!worksheet[cellAddress]) continue;
      worksheet[cellAddress].s = {
        font: { bold: true, color: { rgb: 'FFFFFF' } },
        fill: { fgColor: { rgb: '3498db' } },
        alignment: { horizontal: 'center', vertical: 'center' },
      };
    }

    // 生成文件并触发下载
    const excelFileName = `${filename}_${new Date().toISOString().split('T')[0]}.xlsx`;
    XLSX.writeFile(workbook, excelFileName);
  } catch (error) {
    console.error('导出Excel失败:', error);
    throw new Error(`导出Excel失败: ${error instanceof Error ? error.message : '未知错误'}`);
  }
}

/**
 * 从Blob数据导出Excel文件
 * @param blob Excel文件的Blob数据
 * @param filename 文件名（不含扩展名）
 */
export function exportBlobToExcel(blob: Blob, filename = '操作日志'): void {
  try {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${filename}_${new Date().toISOString().split('T')[0]}.xlsx`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (error) {
    console.error('下载Excel文件失败:', error);
    throw new Error(`下载Excel文件失败: ${error instanceof Error ? error.message : '未知错误'}`);
  }
}

/**
 * 格式化日志状态用于Excel显示
 * @param status 日志状态
 * @returns 格式化后的状态文本
 */
export function formatStatusForExcel(status: 'success' | 'failure'): string {
  return status === 'success' ? '成功' : '失败';
}

/**
 * 格式化日志类型用于Excel显示
 * @param logType 日志类型
 * @returns 格式化后的类型文本
 */
export function formatLogTypeForExcel(logType: 'audit' | 'operation'): string {
  return logType === 'audit' ? 'API日志' : '操作日志';
}

/**
 * 生成导出统计信息
 * @param logs 日志数据
 * @returns 统计信息对象
 */
export interface ExportStats {
  total: number;
  auditCount: number;
  operationCount: number;
  successCount: number;
  failureCount: number;
  timeRange: string;
}

export function generateExportStats(logs: UnifiedLog[]): ExportStats {
  const auditCount = logs.filter((log) => log.log_type === 'audit').length;
  const operationCount = logs.filter((log) => log.log_type === 'operation').length;
  const successCount = logs.filter((log) => log.status === 'success').length;
  const failureCount = logs.filter((log) => log.status === 'failure').length;

  const timestamps = logs.map((log) => new Date(log.timestamp).getTime());
  const minTime = timestamps.length > 0 ? Math.min(...timestamps) : Date.now();
  const maxTime = timestamps.length > 0 ? Math.max(...timestamps) : Date.now();

  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return {
    total: logs.length,
    auditCount,
    operationCount,
    successCount,
    failureCount,
    timeRange: `${formatDate(minTime)} - ${formatDate(maxTime)}`,
  };
}

/**
 * 添加统计信息到Excel工作表
 * @param workbook Excel工作簿
 * @param stats 统计信息
 */
export function addStatsToWorkbook(workbook: XLSX.WorkBook, stats: ExportStats): void {
  const statsData = [
    ['导出统计信息', ''],
    ['导出时间', new Date().toLocaleString('zh-CN')],
    ['总记录数', stats.total],
    ['API日志数', stats.auditCount],
    ['操作日志数', stats.operationCount],
    ['成功记录数', stats.successCount],
    ['失败记录数', stats.failureCount],
    ['时间范围', stats.timeRange],
    ['', ''],
  ];

  const statsWorksheet = XLSX.utils.aoa_to_sheet(statsData);

  // 设置统计信息列宽
  statsWorksheet['!cols'] = [{ wch: 15 }, { wch: 25 }];

  // 添加样式
  const range = XLSX.utils.decode_range(statsWorksheet['!ref'] || 'A1:A1');
  for (let R = 0; R <= range.e.r; ++R) {
    for (let C = range.s.c; C <= range.e.c; ++C) {
      const cellAddress = XLSX.utils.encode_cell({ r: R, c: C });
      if (!statsWorksheet[cellAddress]) continue;

      // 标题行样式
      if (R === 0) {
        statsWorksheet[cellAddress].s = {
          font: { bold: true, size: 14, color: { rgb: 'FFFFFF' } },
          fill: { fgColor: { rgb: '2980b9' } },
          alignment: { horizontal: 'center', vertical: 'center' },
        };
      } else {
        statsWorksheet[cellAddress].s = {
          font: { size: 11 },
          alignment: { vertical: 'center' },
        };
      }
    }
  }

  XLSX.utils.book_append_sheet(workbook, statsWorksheet, '导出统计');
}

// 默认导出
export default {
  exportLogsToExcel,
  exportBlobToExcel,
  formatStatusForExcel,
  formatLogTypeForExcel,
  generateExportStats,
  addStatsToWorkbook,
};