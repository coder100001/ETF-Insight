import React, { useState, useCallback, useRef, useEffect } from 'react';
import styled from 'styled-components';
import { theme } from '../styles/theme';
import { exportApi } from '../services/exportApi';
import type { ExportFormat, ExportStatus } from '../types/export';
import { DEFAULT_EXPORT_CONFIG } from '../types/export';

// 样式组件
const Container = styled.div`
  position: relative;
  display: inline-block;
`;

const Button = styled.button<{ $variant?: 'primary' | 'secondary' | 'outline' }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: ${theme.spacing.sm};
  padding: ${theme.spacing.sm} ${theme.spacing.md};
  border-radius: ${theme.borderRadius.md};
  font-size: ${theme.fonts.size.sm};
  font-weight: ${theme.fonts.weight.medium};
  cursor: pointer;
  transition: all ${theme.transitions.normal};
  border: 1px solid transparent;

  ${({ $variant = 'primary' }) => {
    switch ($variant) {
      case 'primary':
        return `
          background: ${theme.colors.primary};
          color: white;
          &:hover {
            background: ${theme.colors.primaryDark};
          }
        `;
      case 'secondary':
        return `
          background: ${theme.colors.info};
          color: white;
          &:hover {
            background: ${theme.colors.primaryDark};
          }
        `;
      case 'outline':
        return `
          background: transparent;
          color: ${theme.colors.primary};
          border-color: ${theme.colors.primary};
          &:hover {
            background: ${theme.colors.primaryLight};
          }
        `;
    }
  }}

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const DropdownMenu = styled.div<{ $isOpen: boolean }>`
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.lg};
  border: 1px solid ${theme.colors.border};
  z-index: 1000;
  margin-top: ${theme.spacing.xs};
  display: ${({ $isOpen }) => ($isOpen ? 'block' : 'none')};
`;

const DropdownItem = styled.button<{ $isSelected?: boolean }>`
  width: 100%;
  padding: ${theme.spacing.sm} ${theme.spacing.md};
  text-align: left;
  background: ${({ $isSelected }) =>
    $isSelected ? theme.colors.primaryLight : 'transparent'};
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: ${theme.spacing.sm};
  font-size: ${theme.fonts.size.sm};
  color: ${theme.colors.textPrimary};
  transition: background ${theme.transitions.fast};

  &:hover {
    background: ${theme.colors.primaryLight};
  }

  &:first-child {
    border-radius: ${theme.borderRadius.md} ${theme.borderRadius.md} 0 0;
  }

  &:last-child {
    border-radius: 0 0 ${theme.borderRadius.md} ${theme.borderRadius.md};
  }
`;

const FormatIcon = styled.span`
  font-size: 16px;
  width: 20px;
  text-align: center;
`;

const LoadingSpinner = styled.div`
  width: 16px;
  height: 16px;
  border: 2px solid ${theme.colors.border};
  border-top: 2px solid ${theme.colors.primary};
  border-radius: 50%;
  animation: spin 1s linear infinite;

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

const StatusMessage = styled.div<{ $type: 'success' | 'error' }>`
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  padding: ${theme.spacing.sm};
  margin-top: ${theme.spacing.xs};
  border-radius: ${theme.borderRadius.md};
  font-size: ${theme.fonts.size.xs};
  text-align: center;
  background: ${({ $type }) =>
    $type === 'success' ? theme.colors.upBg : theme.colors.downBg};
  color: ${({ $type }) =>
    $type === 'success' ? theme.colors.success : theme.colors.danger};
  border: 1px solid ${({ $type }) =>
    $type === 'success' ? theme.colors.success : theme.colors.danger};
`;

// 格式配置
const FORMAT_CONFIG: Record<ExportFormat, { label: string; icon: string; description: string }> = {
  html: { label: 'HTML', icon: '🌐', description: '网页格式，支持样式和交互' },
  pdf: { label: 'PDF', icon: '📄', description: 'PDF文档，适合打印和分享' },
  excel: { label: 'Excel', icon: '📊', description: 'Excel表格，支持数据分析' },
  markdown: { label: 'Markdown', icon: '📝', description: 'Markdown格式，适合文档编辑' }
};

// 组件属性
interface ExportButtonProps {
  pageType: string;
  data: Record<string, unknown>;
  title?: string;
  variant?: 'primary' | 'secondary' | 'outline';
  buttonText?: string;
  showFormatLabel?: boolean;
  onExportStart?: () => void;
  onExportSuccess?: (filename: string) => void;
  onExportError?: (error: string) => void;
  disabled?: boolean;
  supportedFormats?: ExportFormat[];
}

// 组件
export const ExportButton: React.FC<ExportButtonProps> = ({
  pageType,
  data,
  title,
  variant = 'primary',
  buttonText = '导出',
  showFormatLabel = true,
  onExportStart,
  onExportSuccess,
  onExportError,
  disabled = false,
  supportedFormats = DEFAULT_EXPORT_CONFIG.supportedFormats
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [status, setStatus] = useState<ExportStatus>('idle');
  const [statusMessage, setStatusMessage] = useState('');
  const [selectedFormat, setSelectedFormat] = useState<ExportFormat | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // 清理 timer
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  // 处理格式选择
  const handleFormatSelect = useCallback(async (format: ExportFormat) => {
    if (disabled || status === 'loading') return;

    // 清除之前的 timer
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }

    setSelectedFormat(format);
    setIsOpen(false);
    setStatus('loading');
    setStatusMessage('');

    onExportStart?.();

    try {
      const request = {
        format,
        title: title || `${pageType}报告`,
        data
      };

      await exportApi.exportAndDownload(pageType, request);

      setStatus('success');
      setStatusMessage('导出成功！');
      onExportSuccess?.(`${title || pageType}_report.${format === 'excel' ? 'csv' : format}`);

      // 3秒后清除状态消息
      timeoutRef.current = setTimeout(() => {
        setStatus('idle');
        setStatusMessage('');
        timeoutRef.current = null;
      }, 3000);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '导出失败';
      setStatus('error');
      setStatusMessage(errorMessage);
      onExportError?.(errorMessage);

      // 5秒后清除错误消息
      timeoutRef.current = setTimeout(() => {
        setStatus('idle');
        setStatusMessage('');
        timeoutRef.current = null;
      }, 5000);
    }
  }, [pageType, data, title, disabled, status, onExportStart, onExportSuccess, onExportError]);

  // 切换下拉菜单
  const toggleDropdown = useCallback(() => {
    if (!disabled && status !== 'loading') {
      setIsOpen(prev => !prev);
    }
  }, [disabled, status]);

  // 点击外部关闭下拉菜单
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest('[data-export-button]')) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // 获取可用的格式
  const availableFormats = supportedFormats.filter(format =>
    FORMAT_CONFIG[format] !== undefined
  );

  return (
    <Container data-export-button>
      <Button
        $variant={variant}
        onClick={toggleDropdown}
        disabled={disabled || status === 'loading'}
      >
        {status === 'loading' ? (
          <LoadingSpinner />
        ) : (
          <span>📥</span>
        )}
        {buttonText}
        {showFormatLabel && selectedFormat && (
          <span>({FORMAT_CONFIG[selectedFormat].label})</span>
        )}
      </Button>

      <DropdownMenu $isOpen={isOpen}>
        {availableFormats.map(format => (
          <DropdownItem
            key={format}
            onClick={() => handleFormatSelect(format)}
            $isSelected={selectedFormat === format}
          >
            <FormatIcon>{FORMAT_CONFIG[format].icon}</FormatIcon>
            <div>
              <div style={{ fontWeight: theme.fonts.weight.medium }}>
                {FORMAT_CONFIG[format].label}
              </div>
              <div style={{
                fontSize: theme.fonts.size.xs,
                color: theme.colors.textSecondary
              }}>
                {FORMAT_CONFIG[format].description}
              </div>
            </div>
          </DropdownItem>
        ))}
      </DropdownMenu>

      {statusMessage && (
        <StatusMessage $type={status === 'success' ? 'success' : 'error'}>
          {statusMessage}
        </StatusMessage>
      )}
    </Container>
  );
};

// 默认导出
export default ExportButton;
