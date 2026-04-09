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
- **汇率查询** - 提供实时汇率数据

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

## 🚀 启动项目

### 开发模式

```bash
npm run dev
```

启动后访问：http://localhost:5173

### 生产构建

```bash
npm run build
```

构建产物输出到 `dist/` 目录

### 预览生产构建

```bash
npm run preview
```

### 代码检查

```bash
# ESLint 检查
npm run lint

# 类型检查
npm run type-check
```

## 📁 项目结构

```
frontend/
├── src/
│   ├── pages/              # 页面组件
│   │   ├── ETFComparison.tsx      # ETF 对比页面
│   │   ├── ETFDetail.tsx          # ETF 详情页面
│   │   ├── PortfolioAnalysis.tsx  # 投资组合分析
│   │   ├── ASharePortfolio.tsx    # A 股组合页面
│   │   ├── ExchangeRate.tsx       # 汇率页面
│   │   └── InvestmentStrategy.tsx # 投资策略页面
│   ├── components/         # 通用组件
│   ├── services/           # API 服务
│   ├── hooks/              # 自定义 Hooks
│   ├── utils/              # 工具函数
│   ├── types/              # TypeScript 类型定义
│   ├── App.tsx             # 应用根组件
│   └── main.tsx            # 应用入口
├── public/                 # 静态资源
├── package.json            # 依赖配置
├── tsconfig.json           # TypeScript 配置
├── vite.config.ts          # Vite 配置
└── .eslintrc.js            # ESLint 配置
```

## 🎯 页面说明

### ETF 对比页面 (`/etf-comparison`)

展示多只 ETF 的对比数据，包括：
- 基本信息（名称、代码、费率、交易所）
- 实时行情（价格、涨跌幅、成交量）
- 技术指标（波动率、夏普比率、最大回撤）
- 股息数据（股息率、年度收益）

### ETF 详情页面 (`/etf-detail/:symbol`)

单只 ETF 的详细信息页面，包含：
- 实时价格图表
- 历史走势分析
- 基本信息卡片
- 技术指标展示

### 投资组合分析 (`/portfolio`)

创建和分析投资组合：
- 自定义 ETF 配置比例
- 计算总投资价值
- 分析年度股息收益
- 计算税后收益

### A 股组合页面 (`/a-share-portfolio`)

A 股投资组合分析：
- 持仓列表展示
- 分红数据统计
- 收益率计算

### 汇率页面 (`/exchange-rate`)

汇率查询功能：
- 主要货币对汇率
- 汇率走势图表
- 实时汇率更新

## 🔌 API 集成

项目使用 Axios 进行 HTTP 请求，所有 API 服务位于 `src/services/` 目录。

### 主要 API 端点

```typescript
// ETF 相关
GET  /api/etf/list              // 获取 ETF 列表
GET  /api/etf/comparison        // 获取 ETF 对比数据
GET  /api/etf/:symbol/realtime  // 获取实时行情
GET  /api/etf/:symbol/history   // 获取历史数据
POST /api/etf/update-realtime   // 更新实时数据

// 投资组合
POST /api/etf/portfolio         // 分析投资组合

// A 股组合
GET  /api/a-share-portfolio/list     // 获取组合列表
POST /api/a-share-portfolio          // 创建组合
PUT  /api/a-share-portfolio/:id      // 更新组合
DELETE /api/a-share-portfolio/:id    // 删除组合

// 汇率
GET /api/exchange-rate/pairs    // 获取货币对列表
GET /api/exchange-rate/latest   // 获取最新汇率
```

### 使用示例

```typescript
import { etfService } from '@/services/etf';

// 获取 ETF 列表
const etfList = await etfService.getETFList();

// 获取实时行情
const realtimeData = await etfService.getETFRealtime('SCHD');

// 获取对比数据
const comparison = await etfService.getETFComparison(['SCHD', 'SPYD', 'VYM']);
```

## 🎨 技术栈

- **React 18** - 前端框架
- **TypeScript 5.x** - 类型系统
- **Vite 6.x** - 构建工具
- **Ant Design 5.x** - UI 组件库
- **Axios** - HTTP 客户端
- **Recharts** - 图表库
- **Day.js** - 日期处理
- **ESLint + Prettier** - 代码规范

## 📝 开发规范

### 代码风格

- 使用 TypeScript 编写所有代码
- 组件采用函数式编程风格
- 使用 ESLint 和 Prettier 保持代码一致性
- 遵循 React Hooks 最佳实践

### 命名规范

- 组件文件：PascalCase (如 `ETFComparison.tsx`)
- 工具函数：camelCase (如 `formatPrice.ts`)
- 类型定义：PascalCase (如 `ETFData.ts`)
- 常量：UPPER_SNAKE_CASE (如 `API_BASE_URL`)

### Git 提交规范

```bash
# 功能开发
git commit -m "feat: add ETF comparison feature"

# Bug 修复
git commit -m "fix: resolve ETF price display issue"

# 文档更新
git commit -m "docs: update README.md"

# 代码重构
git commit -m "refactor: optimize API service structure"
```

## 🔍 常见问题

### Q: 开发服务器启动失败？

**A**: 检查 Node.js 版本是否满足要求，并确认端口 5173 未被占用。

```bash
node -v  # 检查 Node.js 版本
lsof -ti:5173 | xargs kill  # 释放端口
```

### Q: API 请求失败？

**A**: 确认后端服务已启动（端口 8080），并检查 `.env` 中的 `VITE_API_BASE_URL` 配置。

```bash
# 检查后端服务
curl http://localhost:8080/health
```

### Q: TypeScript 类型错误？

**A**: 运行类型检查查看详细错误，并确保类型定义完整。

```bash
npm run type-check
```

### Q: 构建产物过大？

**A**: 检查是否有不必要的依赖，使用动态导入优化代码分割。

```bash
npm run build -- --sourcemap
npx vite-bundle-visualizer
```

### Q: 缓存导致数据不更新？

**A**: 清除浏览器缓存或使用强制刷新（Ctrl+Shift+R / Cmd+Shift+R）。

## 📄 许可证

MIT License

## 🤝 贡献指南

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📞 联系方式

如有问题或建议，请提交 Issue 或联系开发团队。

---

**最后更新**: 2026-04-08
