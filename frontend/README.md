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

### ETF持仓穿透分析 (v2.6 新增)
- **底层持仓明细**: 查看ETF底层资产及权重分布
- **重叠度可视化**: 两只ETF持仓重叠度图表展示
- **组合穿透分析**: 投资组合底层资产行业/地理分布
- **集中度指标**: Top10权重、Herfindahl指数可视化

### 量化分析视图
- **ETF对比分析**: 多维度并排对比，雷达图、柱状图可视化
- **投资组合情景分析**:
  - 蒙特卡洛模拟可视化，三种市场情景对比
  - **动态 ETF 选择**: 从 API 获取可用 ETF 列表，显示实时价格
  - **灵活权重配置**: 支持添加/删除 ETF，实时权重验证
  - **基于真实数据**: 使用历史数据计算年化收益率、波动率等指标
- **投资组合分析**: 饼图展示配置，收益曲线，风险指标
- **技术指标展示**: RSI、MACD、布林带雷达图可视化
- **风险分析**: VaR/CVaR风险指标展示
- **有效前沿图**: 马科维茨投资组合优化的有效前沿展示

### QuantLib 量化分析 (v2.8 新增)
- **期权定价**: 欧式/美式 Black-Scholes 定价，完整 Greeks 计算
- **收益率曲线**: 构建和可视化，支持多货币 (USD/EUR/CNY/GBP/JPY)
- **债券定价**: 固定收益分析，久期/凸性计算
- **VaR 计算**: QuantLib 引擎驱动的历史模拟法/参数法
- **交互式界面**: 4 个 Tab 的专业分析页面
- **实时图表**: Recharts 收益率曲线可视化

### 🤖 AI Agent 分析 (v2.9 新增)
- **单 Agent 分析**: 选择投资大师进行独立分析
- **团队辩论模式**: 多 Agent 同时分析并综合观点
- **4 个金融 Agent**: Buffett、Graham、Bridgewater、Macro
- **多 LLM 模型**: 支持 OpenAI/Ollama/DeepSeek 切换
- **SSE 流式响应**: 实时展示分析过程
- **中文界面**: 全中文标签和提示

### 回测引擎 (v2.5 新增)
- **策略回测**: 事件驱动回测可视化
- **订单管理**: 市价单/限价单/止损单模拟
- **回测结果**: 收益曲线、回撤曲线、交易统计
- **策略对比**: 多策略回测结果对比分析

### 组合优化 (v2.5 增强)
- **MPT优化**: 均值-方差优化配置界面
- **风险平价**: 等风险贡献权重配置
- **Black-Litterman**: 投资者观点输入与优化
- **优化结果**: 有效前沿曲线、权重分配建议

### 因子分析 (v2.5 新增)
- **因子暴露度**: Fama-French因子雷达图展示
- **归因分析**: 收益归因、风险归因可视化
- **模型选择**: 三因子/五因子模型切换
- **主动收益**: Alpha分解图表

### A股ETF支持
- **实时价格展示**: 当前价格、涨跌幅、成交量
- **分红分析**: 股息率、分红频率、预期分红计算
- **组合配置**: 可视化配置投资占比
- **数据同步**: AKShare数据源状态管理

### 跨资产类别ETF (v2.5 新增)
- **资产类别筛选**: 股票/债券/商品/REIT/货币/多资产/另类
- **地区筛选**: 美国/中国/欧洲/日本/新兴市场/亚太/拉美
- **ETF类型筛选**: 指数/行业/因子/主题/主动/杠杆/反向
- **组合配置建议**: 保守/平衡/激进/股息策略模板

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
│   │   ├── ETFComparisonReport.tsx # ETF对比报告
│   │   ├── ETFDetail.tsx          # ETF详情
│   │   ├── ETFConfig.tsx          # ETF配置管理
│   │   ├── PortfolioAnalysis.tsx   # 投资组合情景分析(动态ETF选择+实时价格)
│   │   ├── PortfolioConfig.tsx     # 投资组合配置
│   │   ├── PortfolioOptimization.tsx # 组合优化(MPT/风险平价/BL)
│   │   ├── TechnicalAnalysis.tsx   # 技术分析(RSI/MACD/布林带)
│   │   ├── RiskAnalysis.tsx        # 风险分析(VaR/CVaR)
│   │   ├── FactorAnalysis.tsx      # 因子分析
│   │   ├── QuantLibAnalysis.tsx    # QuantLib 量化分析 (v2.8)
│   │   ├── AIAgents.tsx            # AI Agent 分析页面 (v2.9)
│   │   ├── ASharePortfolio.tsx     # A股红利ETF组合
│   │   ├── InvestmentStrategy.tsx  # 投资策略
│   │   ├── ExchangeRate.tsx        # 汇率管理
│   │   ├── OperationLogs.tsx       # 操作日志
│   │   └── ...
│   ├── components/         # 公共组件
│   │   ├── Layout.tsx             # 布局组件（含侧边栏展开/收起、closeSidebar）
│   │   ├── PriceChart.tsx         # 价格图表
│   │   ├── ComparisonRadarChart.tsx # 对比雷达图
│   │   ├── HoldingPieChart.tsx    # 持仓饼图
│   │   └── ...
│   ├── services/           # API 服务
│   │   ├── api.ts                 # API调用封装 (含 agentAPI 模块 v2.9)
│   │   └── portfolio.ts           # 投资组合API
│   ├── types/              # TypeScript 类型
│   │   ├── index.ts               # 全局类型定义
│   │   ├── quantlib.ts            # QuantLib 类型 (v2.8)
│   │   └── agent.ts               # AI Agent 类型定义 (v2.9)
│   ├── utils/              # 工具函数
│   │   └── api.ts                 # API工具函数
│   └── styles/             # 样式
│       └── theme.ts               # 主题配置
├── public/                 # 静态资源
├── package.json
└── vite.config.ts
```

## 🧭 路由与导航

### 路由表

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | Dashboard | 仪表盘 |
| `/etf-dashboard` | ETFDashboard | ETF市场总览 |
| `/etf-comparison` | ETFComparison | ETF对比分析 |
| `/etf-detail/:symbol` | ETFDetail | ETF详情 |
| `/etf-config` | ETFConfig | ETF配置管理 |
| `/portfolio-analysis` | PortfolioAnalysis | 投资组合情景分析 |
| `/portfolio-config` | PortfolioConfig | 投资组合配置 |
| `/portfolio-optimization` | PortfolioOptimization | 组合优化 |
| `/a-share-portfolio` | ASharePortfolio | A股红利ETF组合 |
| `/technical-analysis` | TechnicalAnalysis | 技术分析 |
| `/risk-analysis` | RiskAnalysis | 风险分析 |
| `/factor-analysis` | FactorAnalysis | 因子分析 |
| `/factor-timing` | FactorTiming | 因子择时 |
| `/alpha-views` | AlphaViews | Alpha观点管理 |
| `/black-litterman` | BlackLittermanConfig | Black-Litterman配置 |
| `/risk-budget` | RiskBudget | 风险预算 |
| `/quantlib` | QuantLibAnalysis | QuantLib 量化分析 |
| `/ai-agents` | AIAgents | AI Agent 分析页面 |

### 侧边栏导航

| 图标 | 菜单项 | 说明 |
|------|--------|------|
| DashboardOutlined | 仪表盘 | 系统概览 |
| FundOutlined | ETF 管理 | ETF 列表和详情 |
| PieChartOutlined | 投资组合 | 组合分析和优化 |
| LineChartOutlined | 量化分析 | QuantLib 分析 |
| RobotOutlined | AI Agent | AI 投资分析助手 |
| ExperimentOutlined | 因子分析 | Fama-French 分析 |
| SafetyOutlined | 风险分析 | VaR/CVaR 分析 |

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

### 前端测试状态

| 类型 | 状态 | 说明 |
|------|------|------|
| 组件测试 | 🔄 进行中 | StatCard 等组件测试示例 |
| E2E测试 | 📋 计划中 | 端到端测试覆盖主要用户流程 |
| 类型检查 | ✅ 通过 | TypeScript 严格模式 |

### 运行测试

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
- [架构文档](../AGENTS.md)
