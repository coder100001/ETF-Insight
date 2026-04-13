# ETF-Insight 前端

React + TypeScript + Vite 构建的现代化 ETF 分析与对比平台前端应用。

## 📋 项目概述

ETF-Insight 是一个专业的 ETF 分析与对比平台，提供实时行情、历史数据、投资组合分析等功能。前端采用 React 18 + TypeScript + Vite + Ant Design 技术栈，具有高性能、类型安全和优秀的用户体验。

### 主要功能

- **ETF 实时行情** - 展示 ETF 的实时价格、涨跌幅、成交量等数据
- **ETF 对比分析** - 支持多只 ETF 同时对比，包括费率、股息率、波动率等指标
- **投资组合分析** - 创建和分析投资组合，计算年化收益、夏普比率等
- **历史数据查询** - 查看 ETF 的历史价格走势和技术指标
- **A 股组合分析** - 支持 A 股投资组合的分红和收益分析
- **汇率查询** - 提供实时汇率数据和多数据源故障转移支持

## 🔧 环境要求

- **Node.js**: >= 18.0.0
- **npm**: >= 9.0.0
- **浏览器**: Chrome 90+ / Firefox 88+ / Safari 14+

## 📦 安装步骤

### 1. 克隆项目

```bash
cd /Users/liunian/Desktop/dnmp/py_project/frontend
```

### 2. 安装依赖

```bash
npm install
```

### 3. 配置环境变量（可选）

复制环境变量模板文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件配置后端 API 地址：

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=ETF-Insight
```

### 4. 启动开发服务器

```bash
npm run dev
```

应用将在 http://localhost:5173 启动

## 🏗️ 项目结构

```
frontend/
├── src/
│   ├── pages/              # 页面组件 (14个)
│   │   ├── Dashboard.tsx          # 仪表盘
│   │   ├── ETFDashboard.tsx       # ETF 市场总览
│   │   ├── ETFComparison.tsx      # ETF 对比分析
│   │   ├── ETFComparisonReport.tsx # ETF 对比报告
│   │   ├── ETFDetail.tsx          # ETF 详情页
│   │   ├── ETFConfig.tsx          # ETF 配置管理
│   │   ├── PortfolioAnalysis.tsx   # 投资组合分析
│   │   ├── PortfolioConfig.tsx     # 组合配置管理
│   │   ├── ASharePortfolio.tsx     # A股红利ETF组合
│   │   ├── ExchangeRate.tsx        # 汇率管理
│   │   ├── InvestmentStrategy.tsx  # 投资策略
│   │   ├── OperationLogs.tsx       # 操作日志
│   │   ├── WorkflowList.tsx        # 工作流列表
│   │   └── InstanceList.tsx        # 实例列表
│   ├── components/         # 公共组件
│   │   ├── Layout.tsx             # 布局
│   │   ├── PriceChart.tsx         # 价格图表
│   │   ├── ComparisonRadarChart.tsx # 对比雷达图
│   │   ├── ETFFilter.tsx          # ETF 筛选
│   │   ├── HoldingPieChart.tsx    # 持仓饼图
│   │   ├── SectorBarChart.tsx     # 行业柱状图
│   │   ├── StatCard.tsx           # 统计卡片
│   │   └── StockCard.tsx          # 股票卡片
│   ├── services/api.ts     # API 服务 (含请求合并+重试, 类型安全)
│   ├── utils/api.ts        # API 工具函数 (类型安全, ETFHistoryDataItem/ETFConfig)
│   ├── types/index.ts      # TypeScript 类型定义 (含ETFHistoryDataItem/ETFForecastResult)
│   └── styles/theme.ts     # 主题配置
├── package.json
└── vite.config.ts
```

## 🎨 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| **React** | ^19.2.4 | UI 框架 |
| **TypeScript** | ^5.x | 类型安全 |
| **Vite** | latest | 构建工具 |
| **Ant Design** | ^6.3.4 | UI 组件库 |
| **ECharts** | ^6.0.0 | 数据可视化 |
| **Recharts** | ^3.8.1 | 图表组件 |
| **React Router** | ^7.13.2 | 路由管理 |

## 📊 页面功能

### 仪表盘 (Dashboard)
- 系统概览和关键指标展示
- 快速访问常用功能

### ETF 市场总览 (ETF Dashboard)
- ETF 列表和实时行情
- 筛选和排序功能
- 快速查看关键指标

### ETF 对比分析 (ETF Comparison)
- 多只 ETF 并排对比
- 雷达图可视化对比
- 详细指标对比表格

### ETF 详情页 (ETF Detail)
- 详细 ETF 信息展示
- 历史价格图表
- 持仓分析和风险指标

### 投资组合分析 (Portfolio Analysis)
- 投资组合构建和管理
- 收益分析和风险评估
- 资产配置优化建议

### 汇率管理 (Exchange Rate)
- 实时汇率数据展示
- 多数据源状态监控
- 货币转换功能

## 🔧 开发指南

### 代码规范
- **TypeScript**: 严格类型检查，禁用 `any` 类型
- **React**: 函数式组件，Hooks 规范使用
- **命名规范**: 驼峰命名法，语义化命名
- **组件规范**: 单一职责原则，可复用性设计

### 状态管理
- 使用 React Hooks 进行状态管理
- 复杂状态使用 Context API
- 避免不必要的全局状态

### API 调用
- 统一使用 `services/api.ts` 进行 API 调用
- 支持请求合并和重试机制
- 完整的错误处理

### 样式规范
- 使用 Ant Design 组件库
- 自定义样式使用 CSS Modules
- 响应式设计支持

## 🚀 构建和部署

### 开发构建

```bash
# 开发模式
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

### 生产部署

构建后的文件位于 `dist/` 目录，可以部署到任何静态文件服务器。

```bash
# 构建生产版本
npm run build

# 部署到 Nginx
cp -r dist/* /usr/share/nginx/html/
```

## 📈 性能优化

### 代码分割
- 路由级别的代码分割
- 组件懒加载
- 按需引入第三方库

### 缓存策略
- API 响应缓存
- 组件 memoization
- 图片懒加载

### 构建优化
- Vite 构建优化
- Tree shaking
- 压缩和混淆

## 🐛 故障排除

### 常见问题

1. **API 连接失败**
   - 检查后端服务是否启动
   - 确认 API 地址配置正确
   - 检查网络连接

2. **类型错误**
   - 检查 TypeScript 类型定义
   - 确认 API 响应数据结构
   - 更新相关类型定义

3. **样式问题**
   - 检查 CSS Modules 导入
   - 确认 Ant Design 主题配置
   - 验证响应式设计

### 调试工具

- **React Developer Tools**: 组件状态调试
- **Redux DevTools**: 状态管理调试
- **浏览器开发者工具**: 网络请求和性能分析

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！请确保：

1. 遵循项目代码规范
2. 新功能包含单元测试
3. 更新相关文档
4. 通过代码审查

## 📄 许可证

本项目采用 MIT 许可证。

---

**立即体验**: [http://localhost:5173](http://localhost:5173)  
**后端 API**: [http://localhost:8080](http://localhost:8080)  
**项目文档**: [../README.md](../README.md)