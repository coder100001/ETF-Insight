# 文档与代码实现一致性分析实施计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 全面分析 AGENTS.md 中所有功能点的实现状态，识别未实现、部分实现和存在风险的功能。

**架构：** 采用只读分析方法，通过 Grep/Read 工具验证代码存在性和质量，通过 RunCommand 运行测试验证，最后生成结构化报告。

**技术栈：** Go 1.26, React 19, Python FastAPI, SQLite

---

## 文件结构

**将创建的文件：**
- `docs/superpowers/reports/2026-05-10-implementation-analysis-report.md` - 分析报告

**将读取的文件：**
- `AGENTS.md` - 功能定义源
- `backend/services/**/*.go` - 服务实现
- `backend/handlers/*.go` - API 处理器
- `backend/**/*_test.go` - 测试文件
- `frontend/src/pages/*.tsx` - 前端页面

---

## 任务 1：提取 AGENTS.md 功能点

**目标：** 从 AGENTS.md 提取所有功能点并按模块分类

**文件：**
- 读取：`AGENTS.md`
- 输出：功能点清单（内存数据结构）

- [ ] **步骤 1：读取 AGENTS.md 文件**

使用 Read 工具读取完整文件内容。

- [ ] **步骤 2：提取功能点**

从文档中提取所有标记为 ✅/🔄/📋 的功能描述，记录：
- 功能名称
- 文档状态（✅/🔄/📋）
- 所属模块
- 描述信息

- [ ] **步骤 3：按模块分类**

将功能点按以下模块分类：
1. 投资组合分析
2. 组合优化
3. 因子分析
4. 回测引擎
5. 技术指标
6. QuantLib 集成
7. AI Agent
8. 数据源
9. 数据层重构 (v2.6)
10. 插件架构 (v2.8)

---

## 任务 2：验证投资组合分析模块

**目标：** 验证投资组合分析模块的所有功能实现状态

**文件：**
- 检查：`backend/services/scenario_analysis.go`
- 检查：`backend/services/portfolio_analytics.go`
- 检查：`backend/services/risk_models.go`
- 检查：`backend/handlers/portfolio_handler.go`
- 检查：`backend/handlers/portfolio_handler_test.go`
- 检查：`frontend/src/pages/PortfolioAnalysis.tsx`

- [ ] **步骤 1：验证蒙特卡洛模拟**

使用 Grep 搜索关键函数：
```bash
grep -r "MonteCarlo\|monte_carlo" backend/services/
```

检查：
- 是否有 `scenario_analysis.go` 文件
- 是否有蒙特卡洛模拟函数实现
- 函数体是否非空
- 是否有 TODO/FIXME 注释

- [ ] **步骤 2：验证情景分析**

使用 Grep 搜索：
```bash
grep -r "ScenarioAnalysis\|scenario_analysis" backend/services/
```

检查 API 端点：
```bash
grep -r "scenario" backend/handlers/
grep -r "/api/portfolio/scenario" backend/router/
```

- [ ] **步骤 3：验证 VaR/CVaR 计算**

使用 Grep 搜索：
```bash
grep -r "VaR\|CVaR\|ValueAtRisk" backend/services/risk_models.go
```

检查实现完整性：
- 历史模拟法
- 参数法
- 蒙特卡洛法

- [ ] **步骤 4：验证夏普/索提诺/卡尔玛比率**

使用 Grep 搜索：
```bash
grep -r "SharpeRatio\|SortinoRatio\|CalmarRatio" backend/services/
```

- [ ] **步骤 5：验证滚动窗口指标**

使用 Grep 搜索：
```bash
grep -r "RollingWindow\|rolling_window" backend/services/
```

- [ ] **步骤 6：验证股息再投资模型**

使用 Grep 搜索：
```bash
grep -r "DividendReinvest\|dividend_reinvest" backend/services/
```

- [ ] **步骤 7：检查测试文件**

使用 Grep 搜索测试文件：
```bash
find backend/services -name "*_test.go" -exec grep -l "portfolio\|scenario\|risk" {} \;
```

运行测试：
```bash
cd backend && go test ./services/... -run "Portfolio\|Scenario\|Risk" -v
```

- [ ] **步骤 8：检查前端页面**

验证 `frontend/src/pages/PortfolioAnalysis.tsx` 是否存在：
```bash
ls -la frontend/src/pages/PortfolioAnalysis.tsx
```

检查页面功能：
```bash
grep -r "MonteCarlo\|Scenario" frontend/src/pages/PortfolioAnalysis.tsx
```

- [ ] **步骤 9：记录分析结果**

记录每个功能的实现状态：
- 代码状态：✅ 完整 / ⚠️ 部分 / ❌ 未实现
- 测试状态：✅ 通过 / ⚠️ 部分测试 / ❌ 无测试
- 风险等级：高/中/低
- 备注：具体问题

---

## 任务 3：验证组合优化模块

**目标：** 验证组合优化模块的所有功能实现状态

**文件：**
- 检查：`backend/services/optimization/mpt_optimizer.go`
- 检查：`backend/services/optimization/risk_parity.go`
- 检查：`backend/services/optimization/black_litterman.go`
- 检查：`backend/handlers/portfolio_optimizer_handler.go`
- 检查：`backend/handlers/mpt_handler.go`
- 检查：`backend/handlers/risk_parity_handler.go`
- 检查：`backend/handlers/black_litterman_handler.go`
- 检查：`frontend/src/pages/PortfolioOptimization.tsx`

- [ ] **步骤 1：验证马科维茨 MPT**

使用 Grep 搜索：
```bash
grep -r "MPT\|MeanVariance\|EfficientFrontier" backend/services/optimization/
```

检查 API 端点：
```bash
grep -r "/api/optimization/mpt" backend/router/
```

- [ ] **步骤 2：验证风险平价**

使用 Grep 搜索：
```bash
grep -r "RiskParity\|risk_parity" backend/services/optimization/
```

检查实现：
- 等风险贡献 (ERC)
- 反向波动率加权
- 风险预算配置

- [ ] **步骤 3：验证 Black-Litterman**

使用 Grep 搜索：
```bash
grep -r "BlackLitterman\|black_litterman" backend/services/optimization/
```

检查实现：
- 市场均衡收益计算
- 观点融合
- 后验收益计算

- [ ] **步骤 4：验证有效前沿**

使用 Grep 搜索：
```bash
grep -r "EfficientFrontier\|efficient_frontier" backend/services/optimization/
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services/optimization -name "*_test.go"
cd backend && go test ./services/optimization/... -v
```

- [ ] **步骤 6：检查前端页面**

```bash
ls -la frontend/src/pages/PortfolioOptimization.tsx
grep -r "BlackLitterman\|RiskParity" frontend/src/pages/PortfolioOptimization.tsx
```

- [ ] **步骤 7：记录分析结果**

---

## 任务 4：验证因子分析模块

**目标：** 验证因子分析模块的所有功能实现状态

**文件：**
- 检查：`backend/services/factor/fama_french.go`
- 检查：`backend/services/factor_data_service.go`
- 检查：`backend/services/alpha_view_service.go`
- 检查：`backend/handlers/factor_handler.go`
- 检查：`backend/handlers/factor_timing_handler.go`
- 检查：`backend/handlers/alpha_view_handler.go`
- 检查：`frontend/src/pages/FactorAnalysis.tsx`
- 检查：`frontend/src/pages/FactorTiming.tsx`
- 检查：`frontend/src/pages/AlphaViews.tsx`

- [ ] **步骤 1：验证 Fama-French 模型**

使用 Grep 搜索：
```bash
grep -r "FamaFrench\|fama_french" backend/services/factor/
```

检查实现：
- 三因子模型
- 五因子模型
- 因子收益率计算

- [ ] **步骤 2：验证因子择时信号**

使用 Grep 搜索：
```bash
grep -r "FactorTiming\|factor_timing" backend/services/
```

检查实现：
- 60 日 MA 斜率
- Z-score 计算
- 信号强度判断

- [ ] **步骤 3：验证 Alpha 观点管理**

使用 Grep 搜索：
```bash
grep -r "AlphaView\|alpha_view" backend/services/
```

检查实现：
- 观点创建
- 观点验证
- 观点表现追踪

- [ ] **步骤 4：验证归因分析**

使用 Grep 搜索：
```bash
grep -r "Attribution\|attribution" backend/services/factor/
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services -name "*_test.go" -exec grep -l "factor\|alpha" {} \;
cd backend && go test ./services/factor/... -v
cd backend && go test ./services/alpha_view_service_test.go -v
```

- [ ] **步骤 6：检查前端页面**

```bash
ls -la frontend/src/pages/FactorAnalysis.tsx
ls -la frontend/src/pages/FactorTiming.tsx
ls -la frontend/src/pages/AlphaViews.tsx
```

- [ ] **步骤 7：记录分析结果**

---

## 任务 5：验证回测引擎模块

**目标：** 验证回测引擎模块的所有功能实现状态

**文件：**
- 检查：`backend/services/backtest/engine.go`
- 检查：`backend/services/backtest/event_engine.go`
- 检查：`backend/services/backtest/dataprovider.go`
- 检查：`backend/services/backtest/strategies.go`
- 检查：`backend/handlers/backtest_handler.go`

- [ ] **步骤 1：验证事件驱动架构**

使用 Grep 搜索：
```bash
grep -r "EventEngine\|event_engine\|EventBus" backend/services/backtest/
```

检查实现：
- 事件总线
- 事件处理器
- 事件类型定义

- [ ] **步骤 2：验证订单系统**

使用 Grep 搜索：
```bash
grep -r "Order\|order" backend/services/backtest/
```

检查实现：
- 订单类型（市价单、限价单、止损单）
- 订单状态管理
- 订单验证

- [ ] **步骤 3：验证滑点/手续费模型**

使用 Grep 搜索：
```bash
grep -r "Slippage\|Commission\|slippage\|commission" backend/services/backtest/
```

- [ ] **步骤 4：验证分红再投资**

使用 Grep 搜索：
```bash
grep -r "Dividend\|dividend" backend/services/backtest/
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services/backtest -name "*_test.go"
cd backend && go test ./services/backtest/... -v
```

- [ ] **步骤 6：记录分析结果**

---

## 任务 6：验证技术指标模块

**目标：** 验证技术指标模块的所有功能实现状态

**文件：**
- 检查：`backend/services/technical_indicators.go`
- 检查：`backend/services/technical_indicators_test.go`

- [ ] **步骤 1：验证 RSI**

使用 Grep 搜索：
```bash
grep -r "RSI\|RelativeStrengthIndex" backend/services/technical_indicators.go
```

- [ ] **步骤 2：验证 MACD**

使用 Grep 搜索：
```bash
grep -r "MACD\|MovingAverageConvergenceDivergence" backend/services/technical_indicators.go
```

- [ ] **步骤 3：验证布林带**

使用 Grep 搜索：
```bash
grep -r "Bollinger\|bollinger" backend/services/technical_indicators.go
```

- [ ] **步骤 4：验证移动平均线**

使用 Grep 搜索：
```bash
grep -r "SMA\|EMA\|MovingAverage" backend/services/technical_indicators.go
```

- [ ] **步骤 5：检查测试文件**

```bash
cd backend && go test ./services/technical_indicators_test.go -v
```

- [ ] **步骤 6：记录分析结果**

---

## 任务 7：验证 QuantLib 集成模块

**目标：** 验证 QuantLib 集成模块的所有功能实现状态

**文件：**
- 检查：`backend/services/quantlib/quantlib_client.go`
- 检查：`backend/handlers/quantlib_handler.go`
- 检查：`backend/models/quantlib.go`
- 检查：`frontend/src/pages/QuantLibAnalysis.tsx`

- [ ] **步骤 1：验证期权定价**

使用 Grep 搜索：
```bash
grep -r "Option\|option\|BlackScholes" backend/services/quantlib/
```

检查 API 端点：
```bash
grep -r "/api/quantlib/options" backend/router/
```

- [ ] **步骤 2：验证收益率曲线**

使用 Grep 搜索：
```bash
grep -r "YieldCurve\|yield_curve" backend/services/quantlib/
```

- [ ] **步骤 3：验证债券定价**

使用 Grep 搜索：
```bash
grep -r "Bond\|bond" backend/services/quantlib/
```

- [ ] **步骤 4：验证 VaR 计算**

使用 Grep 搜索：
```bash
grep -r "VaR\|ValueAtRisk" backend/services/quantlib/
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services/quantlib -name "*_test.go"
cd backend && go test ./services/quantlib/... -v
```

- [ ] **步骤 6：检查前端页面**

```bash
ls -la frontend/src/pages/QuantLibAnalysis.tsx
```

- [ ] **步骤 7：记录分析结果**

---

## 任务 8：验证 AI Agent 模块

**目标：** 验证 AI Agent 模块的所有功能实现状态

**文件：**
- 检查：`backend/services/agent/agent_server.py`
- 检查：`backend/services/agent/agents/*.py`
- 检查：`backend/services/agent/core/*.py`
- 检查：`backend/services/agent/agent_client.go`
- 检查：`backend/handlers/agent_handler.go`
- 检查：`frontend/src/pages/AIAgents.tsx`

- [ ] **步骤 1：验证 4 个金融 Agent**

使用 Grep 搜索：
```bash
ls -la backend/services/agent/agents/
grep -r "Buffett\|Graham\|Bridgewater\|Macro" backend/services/agent/agents/
```

- [ ] **步骤 2：验证团队辩论模式**

使用 Grep 搜索：
```bash
grep -r "team\|debate\|辩论" backend/services/agent/
```

- [ ] **步骤 3：验证多 LLM 支持**

使用 Grep 搜索：
```bash
grep -r "OpenAI\|Ollama\|DeepSeek\|LLM" backend/services/agent/core/
```

- [ ] **步骤 4：检查 API 端点**

使用 Grep 搜索：
```bash
grep -r "/api/agents" backend/router/
grep -r "/agents/run\|/agents/team" backend/handlers/agent_handler.go
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services/agent -name "test_*.py"
cd backend/services/agent && python -m pytest tests/ -v
```

- [ ] **步骤 6：检查前端页面**

```bash
ls -la frontend/src/pages/AIAgents.tsx
```

- [ ] **步骤 7：记录分析结果**

---

## 任务 9：验证数据源模块

**目标：** 验证数据源模块的所有功能实现状态

**文件：**
- 检查：`backend/services/datasource/finage_provider.go`
- 检查：`backend/services/ashare/akshare_provider.go`
- 检查：`backend/services/ashare/tushare_provider.go`
- 检查：`backend/services/exchange_rate/datasource/*.go`
- 检查：`backend/handlers/exchange_rate.go`

- [ ] **步骤 1：验证 Finage API**

使用 Grep 搜索：
```bash
grep -r "Finage\|finage" backend/services/datasource/
```

检查 API Key 配置：
```bash
grep -r "FINAGE_API_KEY" backend/
```

- [ ] **步骤 2：验证 A 股数据源**

使用 Grep 搜索：
```bash
grep -r "AKShare\|TuShare" backend/services/ashare/
```

检查 Python 服务：
```bash
ls -la backend/services/ashare/akshare_server.py
```

- [ ] **步骤 3：验证汇率数据源**

使用 Grep 搜索：
```bash
ls -la backend/services/exchange_rate/datasource/
grep -r "OpenExchange\|CurrencyAPI\|Frankfurter" backend/services/exchange_rate/datasource/
```

- [ ] **步骤 4：验证故障转移机制**

使用 Grep 搜索：
```bash
grep -r "fallback\|Fallback\|Failover" backend/services/exchange_rate/
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services -name "*_test.go" -exec grep -l "datasource\|exchange_rate" {} \;
cd backend && go test ./services/datasource/... -v
cd backend && go test ./services/exchange_rate/... -v
```

- [ ] **步骤 6：记录分析结果**

---

## 任务 10：验证数据层重构模块 (v2.6)

**目标：** 验证数据层重构模块的所有功能实现状态

**文件：**
- 检查：`backend/models/asset.go`
- 检查：`backend/models/etf_holding.go`
- 检查：`backend/services/portfolio/penetration.go`
- 检查：`backend/services/etf/holdings_service.go`
- 检查：`backend/handlers/etf_holding_handler.go`
- 检查：`backend/handlers/portfolio_penetration.go`

- [ ] **步骤 1：验证统一资产模型**

使用 Grep 搜索：
```bash
grep -r "Asset\|asset" backend/models/asset.go
```

- [ ] **步骤 2：验证 ETF 持仓穿透**

使用 Grep 搜索：
```bash
grep -r "Holding\|holding" backend/models/etf_holding.go
grep -r "Penetration\|penetration" backend/services/portfolio/
```

- [ ] **步骤 3：验证重叠度计算**

使用 Grep 搜索：
```bash
grep -r "Overlap\|overlap" backend/services/etf/
```

- [ ] **步骤 4：验证集中度指标**

使用 Grep 搜索：
```bash
grep -r "Concentration\|Herfindahl" backend/services/
```

- [ ] **步骤 5：检查测试文件**

```bash
find backend/services -name "*_test.go" -exec grep -l "penetration\|holding" {} \;
cd backend && go test ./services/portfolio/... -v
cd backend && go test ./services/etf/... -v
```

- [ ] **步骤 6：记录分析结果**

---

## 任务 11：验证插件架构模块 (v2.8)

**目标：** 验证插件架构模块的所有功能实现状态

**文件：**
- 检查：`backend/models/plugin.go`
- 检查：`backend/services/plugin_service.go`
- 检查：`backend/handlers/` (检查是否有插件相关 handler)

- [ ] **步骤 1：验证插件注册机制**

使用 Grep 搜索：
```bash
grep -r "Plugin\|plugin" backend/models/plugin.go
grep -r "Register\|register" backend/services/plugin_service.go
```

- [ ] **步骤 2：验证插件配置管理**

使用 Grep 搜索：
```bash
grep -r "Configuration\|configuration" backend/models/plugin.go
```

- [ ] **步骤 3：验证插件执行引擎**

使用 Grep 搜索：
```bash
grep -r "Execute\|execute" backend/services/plugin_service.go
```

- [ ] **步骤 4：检查测试文件**

```bash
ls -la backend/services/plugin_service_test.go
cd backend && go test ./services/plugin_service_test.go -v
```

- [ ] **步骤 5：检查 API 端点**

使用 Grep 搜索：
```bash
grep -r "/api/plugins" backend/router/
```

- [ ] **步骤 6：记录分析结果**

---

## 任务 12：搜索 TODO/FIXME 注释

**目标：** 搜索代码中的所有 TODO/FIXME 注释，识别未完成的工作

**文件：**
- 搜索：`backend/**/*.go`
- 搜索：`frontend/src/**/*.{ts,tsx}`

- [ ] **步骤 1：搜索 Go 代码中的 TODO**

使用 Grep 搜索：
```bash
grep -r "TODO\|FIXME\|XXX\|HACK" backend/ --include="*.go" > /tmp/go_todos.txt
```

- [ ] **步骤 2：搜索前端代码中的 TODO**

使用 Grep 搜索：
```bash
grep -r "TODO\|FIXME\|XXX\|HACK" frontend/src/ --include="*.ts" --include="*.tsx" > /tmp/frontend_todos.txt
```

- [ ] **步骤 3：分类整理 TODO**

按模块分类整理：
- 投资组合分析相关
- 组合优化相关
- 因子分析相关
- 其他模块

- [ ] **步骤 4：评估 TODO 影响**

评估每个 TODO 的风险等级：
- 高风险：影响核心功能
- 中风险：影响非核心功能
- 低风险：优化改进项

---

## 任务 13：运行测试套件

**目标：** 运行所有测试，检查测试通过率和覆盖率

**文件：**
- 运行：`backend/**/*_test.go`
- 运行：`backend/services/agent/tests/*.py`

- [ ] **步骤 1：运行 Go 后端测试**

```bash
cd backend && go test ./... -v -coverprofile=coverage.out
```

- [ ] **步骤 2：检查测试覆盖率**

```bash
cd backend && go tool cover -func=coverage.out | tail -n 1
```

- [ ] **步骤 3：运行 Python Agent 测试**

```bash
cd backend/services/agent && python -m pytest tests/ -v --cov=.
```

- [ ] **步骤 4：记录测试结果**

记录：
- 总测试数
- 通过数
- 失败数
- 覆盖率

---

## 任务 14：生成分析报告

**目标：** 汇总所有分析结果，生成结构化报告

**文件：**
- 创建：`docs/superpowers/reports/2026-05-10-implementation-analysis-report.md`

- [ ] **步骤 1：汇总统计数据**

计算：
- 总功能数
- 已完全实现数
- 部分实现数
- 未实现数
- 高风险项数

- [ ] **步骤 2：按模块生成表格**

为每个模块生成表格：
| 功能 | 文档状态 | 代码状态 | 测试状态 | 风险等级 | 备注 |
|------|---------|---------|---------|---------|------|

- [ ] **步骤 3：编写关键发现**

总结：
1. 高风险项列表
2. 文档不一致项
3. 测试缺口
4. TODO/FIXME 汇总

- [ ] **步骤 4：编写建议**

提供：
1. 优先修复建议
2. 测试补充建议
3. 文档更新建议

- [ ] **步骤 5：生成报告文件**

使用 Write 工具创建报告文件。

- [ ] **步骤 6：Commit 报告**

```bash
git add docs/superpowers/reports/2026-05-10-implementation-analysis-report.md
git commit -m "docs: add documentation and implementation analysis report"
```

---

## 自检清单

### 1. 规格覆盖度 ✅
- [x] 覆盖所有 10 个模块
- [x] 包含 TODO/FIXME 搜索
- [x] 包含测试验证
- [x] 包含报告生成

### 2. 占位符扫描 ✅
- [x] 无"待定"或"TODO"占位符
- [x] 所有步骤都有具体的命令或代码
- [x] 所有文件路径都是精确的

### 3. 类型一致性 ✅
- [x] 所有 Grep 命令语法正确
- [x] 所有文件路径一致
- [x] 所有命令可执行

---

**计划已完成并保存到 `docs/superpowers/plans/2026-05-10-documentation-implementation-analysis.md`。**

**两种执行方式：**

**1. 子代理驱动（推荐）** - 每个任务调度一个新的子代理，任务间进行审查，快速迭代

**2. 内联执行** - 在当前会话中使用 executing-plans 执行任务，批量执行并设有检查点

**选哪种方式？**
