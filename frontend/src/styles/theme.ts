// Django 模板风格主题配置
export const theme = {
  colors: {
    // 主色调
    primary: '#3498db',
    primaryLight: '#ecf0f1',
    primaryDark: '#2980b9',

    // 侧边栏
    sidebarBg: '#2c3e50',
    sidebarHover: '#34495e',
    sidebarActive: '#34495e',
    sidebarBorder: '#3498db',

    // 涨跌色 - 兼容旧组件
    success: '#2ecc71',
    warning: '#f39c12',
    danger: '#e74c3c',
    info: '#3498db',
    up: '#e74c3c',        // 红色 - 涨
    down: '#2ecc71',      // 绿色 - 跌
    upBg: '#fdeaea',      // 涨背景
    downBg: '#eafaf1',    // 跌背景

    // 中性色
    background: '#f5f5f5',
    surface: '#ffffff',
    border: '#e0e0e0',
    divider: '#dee2e6',

    // 文字色
    textPrimary: '#333333',
    textSecondary: '#7f8c8d',
    textMuted: '#95a5a6',
    textTertiary: '#95a5a6',
    textInverse: '#ffffff',

    // 图表色
    chartColors: ['#3498db', '#2ecc71', '#e74c3c', '#f39c12', '#9b59b6', '#1abc9c'],
  },

  // 字体
  fonts: {
    family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    familyMono: '"SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, monospace',
    size: {
      xs: '11px',
      sm: '12px',
      base: '14px',
      lg: '16px',
      xl: '18px',
      '2xl': '24px',
      '3xl': '32px',
      '4xl': '48px',
    },
    weight: {
      normal: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
    },
  },

  // 间距
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '12px',
    lg: '16px',
    xl: '20px',
    '2xl': '24px',
    '3xl': '32px',
    '4xl': '40px',
  },

  // 圆角
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    full: '9999px',
  },

  // 阴影
  shadows: {
    sm: '0 1px 2px rgba(0, 0, 0, 0.05)',
    md: '0 2px 4px rgba(0, 0, 0, 0.1)',
    lg: '0 4px 8px rgba(0, 0, 0, 0.1)',
    card: '0 2px 4px rgba(0, 0, 0, 0.1)',
    hover: '0 4px 12px rgba(0, 0, 0, 0.15)',
  },

  // 过渡动画
  transitions: {
    fast: '0.15s ease',
    normal: '0.25s ease',
    slow: '0.35s ease',
  },

  // 布局
  layout: {
    sidebarWidth: '250px',
    headerHeight: '60px',
  },

  // 断点
  breakpoints: {
    xs: '480px',
    sm: '576px',
    md: '768px',
    lg: '992px',
    xl: '1200px',
    '2xl': '1600px',
  },
};

export type Theme = typeof theme;
