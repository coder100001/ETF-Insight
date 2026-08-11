# 架构重构实施计划

> **日期**: 2026-08-06
> **状态**: 评审修订版 v2，待执行
> **范围**: 前端图表库统一 + 后端依赖注入 + 前端状态管理
> **修订说明 (v2)**: 修复 v1 评审发现的循环导入、虚构风险、验收标准与范围矛盾问题；Phase 3 范围扩展为统一 4 个页面的 ETF 列表数据源；执行顺序改为前后端并行。

---

## 决策摘要 (ADR)

### ADR-001: 统一图表库为 ECharts

| 项目 | 内容 |
|------|------|
| **决策** | 移除 Recharts 和 @ant-design/charts，统一使用 ECharts |
| **驱动力** | 项目当前混用三个图表库（ECharts、Recharts、@ant-design/charts），维护成本高；@ant-design/charts 间接引入 @ant-design/plots/G2，包体积大（移除它是减包大头）；ECharts 在金融图表（K线图、热力图、桑基图）能力更强 |
| **替代方案** | A) 保持三库混用（被拒：维护成本高） B) 全部迁移到 Recharts（被拒：金融图表能力弱） C) 迁移到 @ant-design/charts（被拒：底层依赖 G2，包体积更大且灵活性不如原生 ECharts） |
| **影响范围** | 8 个前端页面文件（7 个使用 Recharts + 1 个使用 @ant-design/charts）及其测试 |
| **可逆性** | 两向门 — 可回退，但迁移后不应回退 |
| **负责人** | 开发者 + 本地 `npm run build && npm run test` 验证 |
| **备注** | 统一后 echarts 为全量 import（`import * as echarts`），与现有 6 个 ECharts 文件一致。按需注册（`echarts/core` + `use()`）可进一步减包，但不在 Phase 1 范围，作为后续优化项 |

### ADR-002: 引入 google/wire 依赖注入

| 项目 | 内容 |
|------|------|
| **决策** | 引入 google/wire 编译时依赖注入，替代手动构造 |
| **驱动力** | router.go 手动创建所有 handler/service，构造函数参数列表随功能增长而膨胀 |
| **替代方案** | A) uber/fx（被拒：运行时 DI 增加启动开销和调试复杂度） B) 保持手动（被拒：不可持续） |
| **影响范围** | core/app.go, router/router.go, 新增 di/ 包, main.go, CI |
| **可逆性** | 两向门 — wire 生成代码可手动维护 |
| **负责人** | 开发者 + 本地 `go build ./... && go test ./...` 验证 |
| **前置约束** | 必须按 2.2 Step 3 的方案打破 core ↔ di 循环导入（main.go 直连 di），并新增 core 导出构造函数 |

### ADR-003: 引入 Zustand 前端状态管理

| 项目 | 内容 |
|------|------|
| **决策** | 引入 Zustand 作为全局状态管理方案，统一 4 个页面的 ETF 列表数据源 |
| **驱动力** | Dashboard / ETFDashboard / PortfolioAnalysis / PortfolioOptimization 各自向 `/etf/list` 发重复请求；未来用户体系和告警系统需要全局状态 |
| **替代方案** | A) Redux Toolkit（被拒：样板代码多，对小项目过重） B) Jotai（被拒：原子粒度管理对中大型项目不如 store 直观） C) Context API（被拒：性能问题 — 全局 context 更新触发所有消费者重渲染） |
| **影响范围** | 新增 stores/ 目录；迁移 useETFData.ts；接入 Dashboard / ETFDashboard / PortfolioAnalysis / PortfolioOptimization |
| **可逆性** | 两向门 — Zustand store 可逐步引入，与现有 useState 并存 |
| **负责人** | 开发者 + 本地 `npm run build && npm run test` 验证 |
| **范围界定** | 仅统一 ETF 列表/统计数据；useFinancialConfig、useOptimization 不迁移；portfolioStore/notificationStore 不创建（延后到有真实需求时） |

---

## Phase 1: ECharts 统一迁移 (预计 2-3 天)

### 1.1 现状分析（已核查验证）

**使用 Recharts 的文件 (7个)**:
- `pages/PortfolioAnalysis.tsx` — LineChart (多线), BarChart (分组柱状) (组合分析)
- `pages/ETFComparison.tsx` — RadarChart (雷达), BarChart (横向柱状) (ETF对比)
- `pages/RiskAnalysis.tsx` — RadarChart (雷达), BarChart (柱状) (风险分析)
- `pages/PortfolioOptimization.tsx` — PieChart (饼图+环形), ScatterChart (散点) (组合优化)
- `pages/ASharePortfolio.tsx` — PieChart (环形), BarChart (分组柱状) (A股组合)
- `pages/DCACalculator.tsx` — AreaChart (面积图) (DCA计算器)
- `pages/InvestmentStrategy.tsx` — AreaChart + Line + ReferenceLine (面积图叠加折线+参考线) (投资策略)

**使用 @ant-design/charts 的文件 (1个)**:
- `pages/TechnicalAnalysis.tsx` — `import { Radar, Line } from '@ant-design/charts'` (技术分析)
- 对应测试: `pages/__tests__/TechnicalAnalysis.test.tsx:7` — 已 `vi.mock('@ant-design/charts')`（返回 null 组件）

**使用 ECharts 的文件 (5个)**:
- `pages/Dashboard.tsx`、`components/HoldingPieChart.tsx`、`components/PriceChart.tsx`、`components/ComparisonRadarChart.tsx`、`components/SectorBarChart.tsx` — 均直接 `echarts.init`（echarts 6.0.0 下 `echarts.graphic.LinearGradient`、`echarts.init` 均验证存在）

> **注意**: 以上 5 个已有 ECharts 组件均直接使用 `echarts.init`，未使用统一封装。迁移完成后可选择性重构为使用 `ReactEChart` 封装组件，但不作为 Phase 1 的必须项。

**测试现状（已核查）**: 迁移涉及的 8 个页面中 7 个有测试文件（DCACalculator 无测试）；现有测试**只断言文本/loading 状态，不断言图表 DOM**，迁移后断言无需重写。

### 1.2 迁移步骤

#### Step 1: ECharts React 封装组件 (已完成)

`components/ReactEChart.tsx` 已创建并就绪（单一通用封装，option 式 API）：
- 自动 init/dispose 生命周期管理
- `window.resize` + `ResizeObserver` 双重 resize 监听
- Safari 兼容性处理（自动禁用动画）
- 支持 canvas / svg 渲染器
- TypeScript 类型安全（`EChartsOption` 类型约束）
- option 变更时自动 `setOption(option, { notMerge: true })`

**设计理由**: ECharts 采用统一的 `EChartsOption` 配置模式，通过 `series[].type` 区分图表类型，无需为每种图表类型创建独立组件。

#### Step 2: 逐页面迁移 (1.5天)

按复杂度从低到高排序迁移：

| 顺序 | 文件 | 原图表库及组件 | 迁移复杂度 | 预计时间 |
|------|------|--------------|-----------|---------|
| 1 | DCACalculator.tsx | Recharts: AreaChart | 低 | 0.5h |
| 2 | InvestmentStrategy.tsx | Recharts: AreaChart + Line + ReferenceLine | 低 | 0.5h |
| 3 | ASharePortfolio.tsx | Recharts: PieChart (环形) + BarChart | 中 | 1h |
| 4 | RiskAnalysis.tsx | Recharts: RadarChart + BarChart | 中 | 1h |
| 5 | ETFComparison.tsx | Recharts: RadarChart + BarChart (横向) | 中 | 1h |
| 6 | PortfolioOptimization.tsx | Recharts: PieChart + ScatterChart | 高 | 1.5h |
| 7 | PortfolioAnalysis.tsx | Recharts: LineChart (多线) + BarChart (分组) | 高 | 1.5h |
| 8 | TechnicalAnalysis.tsx | @ant-design/charts: Radar + Line | 中 | 1h |

#### Step 3: 测试与清理 (1天)

1. **测试 mock 更新（必须，否则 npm run test 红）**:
   - 6 个迁移页面测试（PortfolioAnalysis / ETFComparison / RiskAnalysis / PortfolioOptimization / ASharePortfolio / InvestmentStrategy）补充 `vi.mock('echarts')`，复用 `Dashboard.test.tsx:7` 的既有 mock 模式（`init` 返回带 setOption/resize/dispose 的 stub）
   - `TechnicalAnalysis.test.tsx` 将 `vi.mock('@ant-design/charts')` 替换为同一 echarts mock 模式
   - 若迁移代码路径触发 `echarts.graphic.LinearGradient`（如渐变填充），需在 mock 中补充 `graphic: { LinearGradient: vi.fn() }`（现有 Dashboard mock 未含此字段）
2. 从 `package.json` 移除 `recharts` 和 `@ant-design/charts` 依赖
3. 执行 `npm install` 更新 lockfile，确认 @ant-design/plots 等间接依赖一并移除
4. 运行 `npm run build` / `npm run test` / `npm run lint` 全部通过
5. 手动验证 8 个迁移页面图表渲染正常

### 1.3 风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| ECharts API 差异导致图表样式不一致 | 中 | 中 | 参考现有 ECharts 组件风格，保持一致 |
| Recharts 的 ResponsiveContainer 在 ECharts 中无等价 | 低 | 低 | ReactEChart 封装中通过 ResizeObserver 处理 |
| 页面测试未补 echarts mock 导致测试失败 | 高 | 低 | Step 3 明确定义 mock 模式（复用 Dashboard.test.tsx），逐页核对 |
| `@ant-design/charts` 间接依赖移除后 lockfile 残留 | 中 | 低 | 移除后执行 `npm install`，grep `@ant-design/plots` 确认清除 |

---

## Phase 2: google/wire 依赖注入 (预计 2 天)

### 2.1 现状分析（已核查验证）

**当前手动 DI 位置**:
- `core/app.go` New() 函数 — 创建所有 service 和 provider
- `router/router.go` NewRouter() — 接收 7 个参数创建所有 handler（`Handlers` 结构体已存在，router.go:17-29，含 11 个 handler 字段）

**New() 的完整副作用序列（wire 化后必须保持此顺序）**:
```
LoadConfig → InitLogger → InitDB(设置全局 models.DB) → InitDefaultData
→ InitEventBus(全局 event.GlobalEventBus) → InitExchangeRateTables
→ InitDefaultCurrencyPairs → 构造 ExchangeRateService/ETFAnalysisService/PortfolioOptimizer
→ NewFinageProvider → ProviderFactory 注册 finage/fallback → GetDefault
→ initUnifiedRegistry(写入全局 unified singleton) → NewScheduler → 构造 exchangeRateConfig
→ NewExchangeRateTask → NewRouter(中间件+handler) → RegisterRoutes → http.Server(TLS) → App
```

**关键事实（已核查）**:
- `New()` 中**不启动任何 goroutine**；scheduler/exchangeRateTask 的 cron 在 `App.Start()` 中启动，汇率同步刷新的是数据库而非 config 结构体
- 全局可变状态共 3 处：`models.DB` (models/db.go:16)、`event.GlobalEventBus` (services/event/trigger.go:202)、unified 全局注册表
- `main.go` 是 `core.New` 的唯一调用者（cmd/ 下 11 个子程序均不经过 core，不受本次重构影响）
- `core.App` 6 个字段全部未导出（config/router/scheduler/exchangeRateTask/server/provider）

### 2.2 实施步骤

#### Step 1: 引入 wire 依赖 (0.5天)

```bash
cd backend
go get github.com/google/wire@latest
go install github.com/google/wire/cmd/wire@latest
```

#### Step 2: 创建 DI Provider 定义 (1天)

创建 `di/` 目录：

```
di/
├── wire.go          # Wire provider 定义 (//go:build wireinject)
├── providers.go     # Provider 函数实现
└── wire_gen.go      # 生成的代码 (wire 自动生成，提交到仓库)
```

**providers.go** — 将 app.go 和 router.go 中的构造逻辑提取为独立 provider 函数：
- `ProvideConfig` — 加载配置
- `ProvideLogger` — `utils.InitLogger`（副作用，无返回）
- `ProvideDB` — 依次执行 `models.InitDB` → `models.InitDefaultData` → `event.InitEventBus(models.DB)` → `models.InitExchangeRateTables` → `models.InitDefaultCurrencyPairs`（**顺序固定，与 app.go 一致**）
- `ProvideExchangeRateService` / `ProvideETFAnalysisService` / `ProvidePortfolioOptimizer` — 服务链
- `ProvideDataSourceProvider` — Finage provider + ProviderFactory 注册 finage/fallback + `GetDefault` + `initUnifiedRegistry`
- `ProvideScheduler` / `ProvideExchangeRateTask` — 任务（含 exchangeRateConfig 构造）
- `ProvideRouter` — 接收 `*Handlers` 结构体（替代 7 散参），保留中间件挂载逻辑
- `ProvideServer` — 构造 http.Server（含 TLS）
- `ProvideApp` — 调用 core 新增的导出构造函数组装 App

**wire.go** — 定义注入关系：
```go
//go:build wireinject

package di

import "etf-insight/core"

func InitializeApp(configPath string) (*core.App, error) {
    wire.Build(
        ProvideConfig, ProvideLogger, ProvideDB, ProvideExchangeRateService,
        ProvideETFAnalysisService, ProvidePortfolioOptimizer,
        ProvideDataSourceProvider, ProvideScheduler,
        ProvideExchangeRateTask, ProvideRouter, ProvideServer,
        ProvideApp,
    )
    return nil, nil
}
```

#### Step 3: 重构入口代码 (0.5天) — 打破循环导入

> **v2 修正**: v1 方案"core.New 调用 di.InitializeApp"构成 core ↔ di 循环导入（di 构造 *core.App 需 import core），编译不过。修正方案：

1. **`core/app.go`**: 新增导出构造函数供 di 调用（App 字段未导出，di 无法直接构造）：
   ```go
   // NewApp 由 wire provider 调用，字段在 di 包组装完成后注入
   func NewApp(cfg *config.Config, r *router.Router, s *tasks.Scheduler,
       erTask *tasks.ExchangeRateTask, srv *http.Server, p datasource.DataSourceProvider) *App
   ```
   删除 `core.New()`（已核查：main.go 是唯一调用者，cmd/ 子程序均不依赖）
2. **`main.go`**: 改为 `di.InitializeApp(*configPath)`，不再 import core（`App.RunOnce()/Start()` 方法仍可用）
3. **`router/router.go`**: `NewRouter` 签名从 7 散参简化为接收 `*Handlers` 结构体（结构体已存在，仅改签名）
4. 生成 wire_gen.go: `wire ./di/`
5. 验证 `go build ./...` 通过

#### Step 4: CI 集成 (0.5天) — fitness functions 落地

> **v2 修正**: v1 的 fitness functions 声称"每次 CI"检查 wire 可再生，但 ci.yml 未装 wire。补充：

1. 修改 `scripts/run-check.sh`（backend 阶段）或 `.github/workflows/ci.yml`：
   - backend 检查步骤中追加 `go install github.com/google/wire/cmd/wire@latest`
   - 追加 `cd backend/di && wire ./... && git diff --exit-code`（生成代码一致性校验）
2. 前端阶段追加图表库单一性 grep 检查（recharts / @ant-design/charts import 数 = 0）
3. 验证 CI 全绿

### 2.3 风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| wire 生成失败 | 中 | 高 | 先确保所有 provider 函数签名正确；wire 原生支持 error 返回，`InitializeApp` 返回 `(*core.App, error)` 自动传播 |
| 全局副作用顺序被破坏（InitDB 系列 + initUnifiedRegistry） | 低 | 高 | ProvideDB/ProvideDataSourceProvider 内部按 2.1 序列固定顺序执行；Provider 链拓扑序为 ProvideConfig → ProvideLogger → ProvideDB → ProvideServices → ProvideRouter → ProvideApp |
| 现有测试依赖 NewRouter 签名 | 低 | 低 | **已排查**: 后端 `*_test.go` 中无直接引用 `NewRouter` 或 `core.New()` 的测试 |
| 删除 core.New 后遗漏调用点 | 低 | 中 | 已核查 main.go 为唯一调用者；删除后 `grep -rn "core.New" --include="*.go"` 确认无残留 |
| CI 未装 wire 导致生成一致性检查失败 | 中 | 低 | Step 4 明确 CI 安装步骤，与 Phase 2 同 PR 落地 |

---

## Phase 3: Zustand 状态管理 (预计 3 天)

> **v2 范围修正**: v1 仅迁移 useETFData（单一使用方，无共享收益），且验收标准"Dashboard/ETFDashboard 共享数据"与步骤矛盾。v2 扩展为统一 4 个页面的 ETF 列表数据源 —— 已核查 4 个页面的 ETF 列表**全部来自同一 endpoint `/etf/list`**（Dashboard/ETFDashboard/PortfolioOptimization 走 `etfAPI.getList()`，PortfolioAnalysis 为同一 endpoint 的裸 `request` 调用），统一进 store 后重复请求消除，验收标准成立。

### 3.1 现状分析（已核查验证）

**ETF 列表请求现状（重复请求问题）**:
| 页面 | 请求方式 | 数据用途 |
|------|---------|---------|
| Dashboard | `etfAPI.getList()` (Promise.allSettled 之一) | 前 10 条 + totalETFs 计数 |
| ETFDashboard | `etfAPI.getList()` | 完整列表 + 字段转换 (ETFApiItem → ETFData) |
| PortfolioAnalysis | 裸 `request('/etf/list?pageSize=100')` | availableETFs (symbol/name) |
| PortfolioOptimization | `useETFData` → `etfAPI.getList()` + `optimizationAPI.etfStatistics()` | ETF 列表 + 统计数据 |

**其他 hooks（本次不迁移）**:
- `useFinancialConfig.ts` — 被 InvestmentStrategy.tsx、PortfolioOptimization.tsx 使用
- `useOptimization.ts` — 仅被 PortfolioOptimization.tsx 使用

### 3.2 实施步骤

#### Step 1: 安装 Zustand (0.1天)

```bash
cd frontend
npm install zustand
```

#### Step 2: 创建 etfStore (0.5天)

```
stores/
├── index.ts              # 统一导出
└── etfStore.ts           # ETF 列表 + 统计数据 store
```

**etfStore.ts** 设计：
- `etfs: ETFData[]` — 原始列表（页面自行转换/裁剪，不做展示层格式）
- `etfStatistics: Record<string, ETFStatistics>` — 统计数据（供 PortfolioOptimization）
- `loading / statsLoading / error`
- `fetchETFList()` — 统一走 `etfAPI.getList()`；**请求去重**：进行中请求由 in-flight Promise 共享，并发调用者复用同一请求
- `fetchStatistics(symbols)` — 走 `optimizationAPI.etfStatistics()`
- **AbortController 保留**: 每次新请求取消上一次未完成请求，catch 中检查 `err.name === 'AbortError'` 跳过状态更新（替代原 hook 的 isMountedRef 模式）

#### Step 3: 接入 4 个页面 (2天)

1. **PortfolioOptimization.tsx** — `useETFData` 替换为 `useETFDataStore`（兼容层，见 Step 4）
2. **ETFDashboard.tsx** — 删除本地 `fetchETFData` 的列表部分，改用 store 数据 + 保留字段转换
3. **Dashboard.tsx** — 从 store 取列表做 slice(0,10) + 计数；`fetchDashboardData` 中移除 `etfAPI.getList()` 调用
4. **PortfolioAnalysis.tsx** — 移除裸 `request('/etf/list?pageSize=100')`，改用 store 数据
5. 清理各页面重复的 loading/error 本地状态，接入 store 的 loading/error

#### Step 4: 兼容层与清理 (0.4天)

1. `useETFData.ts` 保留为兼容包装（内部委托 etfStore），避免遗漏的调用方（已核查当前仅 PortfolioOptimization，但保留成本低）
2. `npm run build` / `npm run test` / `npm run lint` 全部通过
3. 更新受影响的页面测试（Dashboard/ETFDashboard/PortfolioAnalysis 测试若 mock 了 `etfAPI.getList` 的调用次数断言，需相应调整）

### 3.3 风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| Store 状态在测试中难以 mock | 中 | 中 | Zustand 支持 `setState` 直接设置状态用于测试 |
| 页面测试断言 `etfAPI.getList` 调用次数 | 中 | 低 | 迁移时同步调整测试（Dashboard/ETFDashboard 测试可保留各自 mock，store 的 fetch 在页面测试中直接 setState 预置） |
| 请求去重逻辑引入竞态 | 低 | 中 | in-flight Promise 共享 + AbortController，单测覆盖并发调用场景 |
| 页面重渲染性能 | 低 | 低 | Zustand 选择器精确订阅，避免不必要重渲染 |
| useFinancialConfig/useOptimization 未迁移导致状态不一致 | 低 | 低 | Phase 3 明确不迁移，保留原 Hook 行为；后续迭代处理 |

---

## 执行顺序

```
Phase 1: ECharts 迁移 (2-3天) ────────────┐
  ├── Step 1: ReactEChart 封装 ✅ 已完成    ├─ 前后端独立目录，并行执行
  ├── Step 2: 迁移 8 个页面                │
  └── Step 3: 测试 mock + 移除依赖 + 验证   │
                                          │
Phase 2: google/wire DI (2-2.5天) ────────┘
  ├── Step 1: 安装 wire
  ├── Step 2: 创建 providers.go + wire.go
  ├── Step 3: main.go 直连 + core.NewApp + NewRouter 简化
  └── Step 4: CI 集成
                                          │
Phase 3: Zustand (3天) ── 紧随 Phase 1 串行（同目录，避免文件冲突）
  ├── Step 1: 安装 zustand
  ├── Step 2: 创建 etfStore
  ├── Step 3: 接入 4 个页面
  └── Step 4: 兼容层 + 验证
```

**总预计**: 并行后 max(Phase 1, Phase 2) + Phase 3 ≈ **5-5.5 天**（串行为 7-8.5 天）

---

## 验收标准

### Phase 1 验收
- [ ] `package.json` 中无 `recharts`、`@ant-design/charts` 依赖（lockfile 同步清理）
- [ ] 6 个迁移页面测试 + TechnicalAnalysis 测试均补充 echarts mock，`npm run test` 全部通过
- [ ] `npm run build` / `npm run lint` 成功
- [ ] 8 个迁移页面的图表视觉验证通过

### Phase 2 验收
- [ ] `go build ./...` 成功
- [ ] `go test ./...` 全部通过
- [ ] `main.go` 调用 `di.InitializeApp`（不再经 `core.New`），`grep -rn "core.New" --include="*.go"` 无残留
- [ ] `core/app.go` 新增导出构造函数 `NewApp`，原 `New()` 删除
- [ ] `router/router.go` NewRouter 接收 `*Handlers` 结构体
- [ ] `wire ./di/` 可重复生成一致代码（CI 已集成，非仅本地）
- [ ] CI 全绿（含 wire 安装 + 再生成校验步骤）

### Phase 3 验收
- [ ] `package.json` 中有 `zustand` 依赖
- [ ] Dashboard / ETFDashboard / PortfolioAnalysis / PortfolioOptimization 的 ETF 列表均来自 etfStore
- [ ] 首次进入任一页面只发一次 `/etf/list` 请求；二次进入复用 store 数据不重复请求（Network 面板验证）
- [ ] `useETFData` 兼容 Hook 仍可用
- [ ] `npm run build` / `npm run test` / `npm run lint` 全部通过

---

## 架构适配度函数 (Fitness Functions)

| 属性 | 指标 | 阈值 | 测量方式 | 频率 |
|------|------|------|---------|------|
| 图表库单一性 | recharts import 数 | 0 | `grep -r "from 'recharts'" src/` | 每次 CI |
| 图表库单一性 | @ant-design/charts import 数 | 0 | `grep -r "@ant-design/charts" src/` | 每次 CI |
| DI 完整性 | wire_gen.go 可再生 | 生成代码一致 | CI 安装 wire 后 `wire ./di/ && git diff --exit-code` | 每次 CI |
| 状态管理有效性 | etfStore 使用页面数 | ≥ 3 | `grep -rl "etfStore" src/pages/ \| wc -l` | 每次 CI |
| 构建通过 | npm run build / go build | exit 0 | CI | 每次提交 |
| 测试通过 | go test ./... / npm run test | exit 0 | CI | 每次提交 |
