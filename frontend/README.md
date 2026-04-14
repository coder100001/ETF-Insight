# ETF-Insight 前端

[![React Version](https://img.shields.io/badge/React-19.2.4-61DAFB)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-blue)](https://www.typescriptlang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

开源专业的 ETF 量化分析平台前端应用，基于 React 19 + TypeScript + Vite 构建。

## 🎯 项目定位

ETF-Insight 前端致力于提供：

- **专业可视化**: 机构级的数据可视化能力，支持复杂金融图表
- **高性能**: 优化的渲染性能，流畅的大数据量展示
- **类型安全**: TypeScript 严格模式，零 `any` 类型
- **可扩展**: 组件化架构，支持自定义分析视图

## ✨ 核心功能

### 量化分析视图
- **ETF对比分析**: 多维度并排对比，雷达图、柱状图可视化
- **投资组合分析**: 饼图展示配置，收益曲线，风险指标
- **技术指标展示**: 支持各类技术分析指标可视化
- **有效前沿图**: 马科维茨投资组合优化的有效前沿展示

### A股ETF支持
- **实时价格展示**: 当前价格、涨跌幅、成交量
- **分红分析**: 股息率、分红频率、预期分红计算
- **组合配置**: 可视化配置投资占比

### 数据可视化
- **ECharts/Recharts**: 专业金融图表库
- **响应式设计**: 适配桌面和移动端
- **主题定制**: 支持亮色/暗色主题

## 🔧 环境要求

- **Node.js**: >= 18.0.0
- **npm**: >= 9.0.0
- **浏览器**: Chrome 90+ / Firefox 88+ / Safari 14+

## 📦 安装步骤

### 1. 安装依赖

```bash
cd frontend
npm install
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=ETF-Insight
```

### 3. 启动开发服务器

```bash
npm run dev
```

应用将在 http://localhost:5173 启动

## 🏗️ 项目结构

```
frontend/
├── src/
│   ├── pages/              # 页面组件
│   │   ├── Dashboard.tsx          # 仪表盘
│   │   ├── ETFDashboard.tsx       # ETF市场总览
│   │   ├── ETFComparison.tsx      # ETF对比分析
│   │   ├── ETFDetail.tsx          # ETF详情
│   │   ├── PortfolioAnalysis.tsx   # 投资组合分析
│   │   ├── ASharePortfolio.tsx     # A股红利ETF组合
│   │   ├── ExchangeRate.tsx        # 汇率管理
│   │   └── ...
│   ├── components/         # 公共组件
│   │   ├── Layout.tsx             # 布局组件
│   │   ├── PriceChart.tsx         # 价格图表
│   │   ├── ComparisonRadarChart.tsx # 对比雷达图
│   │   ├── HoldingPieChart.tsx    # 持仓饼图
│   │   └── ...
│   ├── services/           # API 服务
│   │   └── api.ts                 # API调用封装
│   ├── types/              # TypeScript 类型
│   │   └── index.ts               # 全局类型定义
│   ├── utils/              # 工具函数
│   │   └── api.ts                 # API工具函数
│   └── styles/             # 样式
│       └── theme.ts               # 主题配置
├── public/                 # 静态资源
├── package.json
└── vite.config.ts
```

## 🎨 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| React | ^19.2.4 | UI 框架 |
| TypeScript | ^5.x | 类型安全 |
| Vite | latest | 构建工具 |
| Ant Design | ^6.3.4 | UI 组件库 |
| ECharts | ^6.0.0 | 数据可视化 |
| Recharts | ^3.8.1 | 图表组件 |
| React Router | ^7.13.2 | 路由管理 |

## 🚀 开发指南

### 代码规范

- **TypeScript**: 严格类型检查，禁用 `any` 类型
- **React**: 函数式组件，Hooks 规范使用
- **样式**: 使用 styled-components 或 CSS Modules
- **组件**: 单一职责原则，可复用组件抽离

### 组件开发示例

```tsx
// 组件结构
import React from 'react';
import styled from 'styled-components';
import { Card } from 'antd';

interface Props {
  title: string;
  data: ETFData;
}

const StyledCard = styled(Card)`
  margin-bottom: 16px;
`;

export const ETFCard: React.FC<Props> = ({ title, data }) => {
  return (
    <StyledCard title={title}>
      {/* 组件内容 */}
    </StyledCard>
  );
};
```

### API 调用

```typescript
import { etfAPI } from '@/services/api';

// 获取ETF列表
const etfs = await etfAPI.getList();

// 获取ETF详情
const detail = await etfAPI.getDetail('SPY');
```

## 🧪 测试

```bash
# 运行测试
npm run test

# 运行测试并生成覆盖率报告
npm run test:coverage

# 运行 ESLint
npm run lint

# 运行 TypeScript 类型检查
npm run type-check
```

## 📦 构建部署

```bash
# 开发构建
npm run build

# 生产构建
npm run build:prod

# 预览生产构建
npm run preview
```

## 🤝 贡献指南

欢迎贡献代码！请遵循以下流程：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 贡献类型
- **UI组件**: 新的可视化组件或图表类型
- **页面功能**: 新的分析页面或功能模块
- **性能优化**: 渲染性能优化、代码分割
- **Bug修复**: 修复界面问题或交互bug

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE) 文件

## 🔗 相关链接

- [项目主页](https://github.com/coder100001/ETF-Insight)
- [后端文档](../backend/README.md)
- [架构文档](../agents.md)
