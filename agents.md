## 📚 文档导航

### 快速入口

| 文档 | 描述 | 适用人群 |
|------|------|----------|
| [README.md](../README.md) | 项目总览和快速开始 | 所有人 |
| [EVOLUTION_ROADMAP_2026.md](./roadmap/EVOLUTION_ROADMAP_2026.md) | 2026年演进路线图 | 项目管理者 |
| [TEST_COVERAGE_PLAN.md](./TEST_COVERAGE_PLAN.md) | 测试覆盖提升计划 | 开发者 |
| [MONITORING_ALERTING_PLAN.md](./MONITORING_ALERTING_PLAN.md) | 监控告警方案 | DevOps |
| [DOCUMENT_INDEX.md](./DOCUMENT_INDEX.md) | 完整文档索引 | 所有人 |

---

## 🔑 核心配置

### API Keys (重要!)

| 服务 | 环境变量 | 状态 |
|------|---------|------|
| **Finage** | `FINAGE_API_KEY` | ✅ **唯一数据源 (必须配置)** |
| **Finnhub** | `FINNHUB_API_KEY` | 🚫 **已废弃** (代码保留但不使用) |

> **⚠️ 安全提醒**: API Key 不得硬编码到代码中，统一通过环境变量配置。参考 `.env.example`。

### 环境变量

```bash
# ========== 代理配置 (Clash VPN) ==========
HTTP_PROXY=http://127.0.0.1:7897
HTTPS_PROXY=http://127.0.0.1:7897

# ========== 数据库配置 ==========
DB_DRIVER=sqlite
DB_DSN=etf_insight.db

# ========== Finage API Key (必须配置) ==========
FINAGE_API_KEY=your_finage_api_key_here

# ========== 汇率数据源配置 ==========
OPENEXCHANGE_API_KEY=your_key_here       # 主数据源
CURRENCYAPI_KEY=your_key_here            # 备份数据源
# Frankfurter 免费无需 API Key

# ========== 安全配置 ==========
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# ========== 后端服务配置 ==========
SERVER_PORT=8080
SERVER_HOST=localhost

# ========== 缓存配置 ==========
REDIS_ENABLED=false
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# ========== 日志级别 ==========
LOG_LEVEL=info
```

---

## 📊 数据源策略

### ETF 数据源
- **主数据源**: Finage API (唯一真实数据源)
- **数据质量**: 实时数据、完整字段、入库校验
- **同步频率**: 定时任务自动更新

### 汇率数据源
- **主数据源**: Open Exchange Rates
- **备用数据源**: CurrencyAPI、Frankfurter
- **故障转移**: 自动切换、健康检查
- **同步策略**: 5分钟间隔、数据一致性保证

### 数据验证
- ✅ **字段完整性**: 所有字段必须入库
- ✅ **数据准确性**: 实时数据校验
- ✅ **一致性检查**: 多数据源对比验证

---

## 🗺️ 演进路线图 (v2.4 → v3.0) - 开源专业分析平台

> **核心理念**: 坚持开源、专业、透明，打造学术界和业界认可的 ETF 量化分析基础设施

### 第一阶段：分析深度增强 (v2.4 - v2.5) ✅ 完成

#### 技术基础完善
- ✅ **数据质量**: 多数据源故障转移、数据完整性校验
- ✅ **性能优化**: 智能缓存、数据库索引优化、前端懒加载
- 🔄 **测试覆盖**: 单元测试覆盖率 >80%，核心算法 100% 覆盖

#### 量化分析能力
- ✅ **投资组合优化**: 马科维茨模型、有效前沿、夏普比率最大化
- ✅ **投资组合情景分析**: 蒙特卡洛模拟、三种市场情景（乐观/中性/悲观）、VaR/CVaR风险指标
- ✅ **技术指标**: RSI、MACD、布林带、均线系统 (测试覆盖率 100%)
- ✅ **风险模型**: VaR、CVaR（历史法/参数法）、压力测试、情景分析 (测试覆盖率 100%)
- ✅ **因子分析**: Fama-French 三因子/五因子模型、归因分析 (测试覆盖率 80.8%)
  - Tikhonov正则化解决数值稳定性
  - 组合风险分解
  - 多组合对比分析
- ✅ **回测引擎**: 事件驱动架构、订单系统、滑点/手续费模型
- ✅ **组合优化增强**: 风险平价模型、Black-Litterman模型

#### 数据扩展
- ✅ **A股数据源**: AKShare/TuShare接入、实时行情同步
- ✅ **跨资产类别**: 股票/债券/商品/REIT/货币全覆盖

### 第二阶段：研究平台化 (v2.6 - v2.8) - 进行中

#### 数据层重构 (v2.6 已完成)
- ✅ **统一资产模型**: Asset 基表支持股票/ETF/指数等多种资产类型
- ✅ **ETF持仓穿透**: 底层持仓明细查询、权重分析
- ✅ **重叠度计算**: 两只ETF持仓重叠度分析(最小权重法)
- ✅ **组合穿透分析**: 投资组合底层资产行业/地理分布
- ✅ **集中度指标**: Top10/Top20权重、Herfindahl指数、有效持仓数
- ✅ **智能缓存**: 重叠度计算结果缓存(7天TTL)、自动失效
- ✅ **事件驱动**: 持仓更新自动触发缓存失效
- ✅ **测试覆盖**: portfolio 84.9%, event 73.2%

#### 学术研究支持
- ✅ **回测引擎**: 事件驱动回测、滑点模拟、交易成本建模
- 🔄 **策略框架**: 策略模板、参数优化、 Walk-forward 分析
- 🔄 **论文复现**: 经典量化策略实现（动量、价值、低波动等）
- 🔄 **数据导出**: CSV/Excel/JSON 格式，支持学术研究

#### 模型深度融合 (v2.7 规划)
- 📋 **Alpha+BL闭环**: Fama-French因子择时 → Black-Litterman观点融合
  - 因子择时信号生成（60日MA斜率、Z-score）
  - 观点量化与信心水平映射
  - BL模型后验收益优化
- 📋 **CVaR风险预算**: 蒙特卡洛模拟 → CVaR约束风险平价
  - 尾部风险预算设定
  - 风险贡献分解与优化
  - 极端行情回撤改善验证

#### 微内核架构 (v2.8 - v2.9)
- ✅ **QuantLib 云 API 集成**: 直接调用 `api.fincept.in/quantlib/` 云服务
  - 期权定价: 欧式/美式 Black-Scholes，完整 Greeks
  - 收益率曲线: 多货币构建和可视化
  - 债券定价: 久期、修正久期、凸性
  - VaR 计算: 历史模拟法/参数法
  - 参考数据: 缓存 1 小时 TTL
- ✅ **前端量化分析页面**: `QuantLibAnalysis.tsx` - 4 Tab 交互式界面
- ✅ **AI Agent 微服务**: 从零重写 Agent 框架，4 个金融 Agent，多 LLM 支持
  - 路由: `/ai-agents`
  - 支持单 Agent 执行和多 Agent 团队辩论
- 📋 **插件接口标准化**:
  - Alpha Generator插件（输入行情/因子，输出观点）
  - Portfolio Optimizer插件（输入观点，输出权重）
  - Risk Model插件（输入权重，输出风险指标）
- 📋 **策略实验台**: 同平台公平竞赛，模型基准对比矩阵

#### 数据源扩展
- ✅ **插件架构**: 标准化数据源接口，支持自定义数据源
- 🔄 **另类数据**: 情绪指标、资金流向、宏观经济数据
- 🔄 **历史数据**: 更长时间序列，支持长期研究

#### 社区协作
- 🔄 **策略分享**: 社区策略仓库、版本控制、回测验证
- 🔄 **Notebook 支持**: Jupyter 集成，支持交互式研究
- 🔄 **文档完善**: API 文档、算法说明、使用教程

### 第三阶段：开源生态 (v3.0+) - 6-12个月

#### 平台基础设施
- 🔄 **容器化部署**: Docker Compose、Kubernetes 支持
- 🔄 **插件市场**: 官方插件 + 社区插件，可扩展分析能力
- 🔄 **API 开放**: RESTful API 完整开放，支持第三方集成
- 🔄 **Webhook 支持**: 实时数据推送、事件通知

#### 学术合作
- 🔄 **论文引用**: 成为学术研究的基础设施，提供引用支持
- 🔄 **数据集发布**: 定期发布清洗后的 ETF 数据集
- 🔄 **基准指数**: 构建开源 ETF 策略基准指数

#### 社区治理
- 🔄 **开源治理**: 明确的贡献指南、代码审查流程
- 🔄 **技术委员会**: 核心贡献者决策机制
- 🔄 **资金透明**: 捐赠和赞助使用公开透明

---

## 🤝 贡献指南

### 如何贡献

我们欢迎各种形式的贡献：

| 贡献类型 | 说明 | 示例 |
|---------|------|------|
| **代码贡献** | 新功能、Bug修复、性能优化 | 实现新的技术指标 |
| **策略贡献** | 量化策略实现和回测 | 动量策略、均值回归策略 |
| **数据源** | 新增数据源适配器 | 接入新的数据提供商 |
| **文档** | 使用文档、API文档、教程 | 编写策略开发指南 |
| **测试** | 单元测试、集成测试 | 增加算法测试覆盖 |
| **反馈** | Bug报告、功能建议 | 提交 Issue |

### 贡献流程

```bash
# 1. Fork 项目
# 2. 创建功能分支
git checkout -b feature/your-feature-name

# 3. 提交代码（遵循 conventional commits）
git commit -m "feat(analysis): 添加布林带指标计算"

# 4. 推送到你的 Fork
git push origin feature/your-feature-name

# 5. 创建 Pull Request
# 详细说明你的改动，关联相关 Issue
```

### 代码规范

- **Go**: 遵循官方规范，使用 `gofmt` 格式化
- **TypeScript**: 严格类型检查，禁用 `any` 类型
- **测试**: 新功能必须包含单元测试，覆盖率 >80%
- **文档**: 算法实现必须包含公式说明和参考文献

---

## 🔧 最新技术更新

### v2.8 更新内容 (2026-05-04)

#### FinceptTerminal QuantLib 集成 (Phase 1 完成)
- ✅ **QuantLib 云 API 对接**: 直接调用 `api.fincept.in/quantlib/` 云服务
  - `services/quantlib/quantlib_client.go` - HTTP 客户端 (10 个方法 + 缓存)
  - `services/quantlib/quantlib_client_test.go` - 9 个单元测试 (httptest mock)
  - `models/quantlib.go` - 13 个请求/响应模型
  - `models/quantlib_validator.go` - 6 个输入验证函数
  - `handlers/quantlib_handler.go` - 7 个 Gin handler

- ✅ **期权定价**: 欧式/美式期权 Black-Scholes 定价
  - API: `POST /api/quantlib/options/european`
  - API: `POST /api/quantlib/options/american`
  - 包含完整 Greeks: Delta, Gamma, Theta, Vega, Rho

- ✅ **收益率曲线**: 构建和可视化收益率曲线
  - API: `POST /api/quantlib/yield-curve/build`
  - 支持多货币: USD, EUR, CNY, GBP, JPY

- ✅ **债券定价**: 固定收益债券定价分析
  - API: `POST /api/quantlib/bonds/price`
  - 包含久期、修正久期、凸性

- ✅ **VaR 计算**: QuantLib 引擎驱动的风险价值计算
  - API: `POST /api/quantlib/risk/var`
  - 支持历史模拟法、参数法

- ✅ **参考数据**: 缓存 1 小时 TTL
  - API: `GET /api/quantlib/reference/:type` (currencies/frequencies/calendars/daycount)

- ✅ **前端页面**: `QuantLibAnalysis.tsx` - 4 个 Tab (期权/债券/收益率曲线/VaR)
  - 交互式表单 + 实时结果展示
  - Recharts 收益率曲线图表
  - 路由: `/quantlib`

- ✅ **代码审查**: 2 轮审查，10 个问题已修复 (P0×3, P1×4, P2×3)

#### AI Agent 微服务 (v2.9 新增)
- ✅ **从零重写 Agent 框架**: 零 AGPL 风险，完全自主实现
  - `services/agent/core/base_agent.py` - Agent 抽象基类
  - `services/agent/core/llm_provider.py` - 多模型抽象层 (OpenAI/Ollama/DeepSeek)
  - `services/agent/core/tool_registry.py` - 工具注册与调用机制
  - `services/agent/core/agent_manager.py` - Agent 注册/发现/执行/团队协作
- ✅ **4 个金融 Agent**:
  - Warren Buffett - 价值投资大师 (经济护城河、内在价值)
  - Benjamin Graham - 防御型投资 (安全边际、净净值分析)
  - Bridgewater Associates - 宏观风险平价 (经济机器模型)
  - Macroeconomic Analyst - 自上而下宏观分析
- ✅ **FastAPI 服务** (port 8091): 6 个端点
  - `GET /health` - 健康检查
  - `GET /agents/discover` - Agent 列表
  - `POST /agents/run` - 单 Agent 执行
  - `POST /agents/stream` - SSE 流式响应
  - `POST /agents/team` - 多 Agent 团队辩论 (2-5 Agent, 1-3 轮)
- ✅ **Go 后端集成**:
  - `services/agent/agent_client.go` - HTTP 客户端
  - `handlers/agent_handler.go` - 4 个 Gin handler
  - `router/router.go` - `/api/agents/*` 路由
- ✅ **前端页面**: `AIAgents.tsx` - 单 Agent/团队分析模式
  - 路由: `/ai-agents`
  - 侧边栏: "AI Agent" (RobotOutlined)
  - 支持 4 个 LLM 模型选择
- ✅ **测试覆盖**: 19 个 Python 单元测试全部通过
- ✅ **容器化**: Dockerfile (Python 3.11-slim)

- 📋 **下一步**: 添加更多 Agent (地缘政治/技术分析/对冲基金)

### v2.5 更新内容

#### 投资组合情景分析 (新增)
- ✅ **蒙特卡洛模拟**: 1000次模拟路径，基于几何布朗运动模型
  - `services/scenario_analysis.go` - 情景分析服务
  - 使用真实历史数据计算收益率和波动率
  - 三种市场情景：乐观(+20%)、中性(0%)、悲观(-20%)
- ✅ **金融计算公式**: 标准金融公式计算组合指标
  - `services/portfolio_analytics.go` - 组合分析服务
  - 组合收益率: $R_p = \sum w_i R_i$
  - 组合方差: $\sigma_p^2 = \sum_i w_i^2 \sigma_i^2 + 2 \sum_{i<j} w_i w_j \sigma_i \sigma_j \rho_{ij}$
  - 年化计算: $\times \sqrt{252}$
- ✅ **风险指标**: VaR 95%/99%、CVaR 95%、最大回撤、夏普比率
  - 参数法计算VaR/CVaR，基于正态分布假设
  - 蒙特卡洛模拟计算置信区间
- ✅ **新增风险调整指标**:
  - **索提诺比率 (Sortino Ratio)**: $(R_p - R_f) / \sigma_d$，只考虑下行风险
  - **卡尔玛比率 (Calmar Ratio)**: $R_p / \text{最大回撤}$，收益与回撤的比值
  - **下行偏差 (Downside Deviation)**: 只考虑低于目标收益率的波动
  - **偏度 (Skewness)**: 收益分布的不对称性
  - **峰度 (Kurtosis)**: 收益分布的尾部厚度
- ✅ **滚动窗口指标**: 动态计算近期表现
  - 支持 30日/60日/90日/180日/1年(252日) 窗口
  - 包含年化收益率、波动率、夏普比率、最大回撤等
  - 新增交易指标: 胜率、平均盈亏、盈亏比
- ✅ **改进股息再投资模型**:
  - 季度再投资: 每季度支付并立即再投资
  - 月度再投资: 模拟实际季度支付模式(3/6/9/12月)
  - 更精确的复利计算，考虑时间价值
- ✅ **默认投资组合模板**: 6种预设组合
  - 保守型、平衡型、进取型、收入型、股息增长型、科技聚焦型
  - API: `GET /api/portfolio/default-templates`
- ✅ **前端分析页面**: 交互式情景分析可视化
  - `PortfolioAnalysis.tsx` - 投资组合情景分析页面
  - **动态 ETF 选择**: 从 API 获取可用 ETF 列表，显示实时价格
  - **灵活配置**: 支持添加/删除 ETF，动态调整权重
  - **权重验证**: 实时显示总配比，100% 时才可分析
  - 折线图展示三种情景下的资产增长路径
  - 对比表格展示各情景关键指标

#### 实时数据获取 (v2.5 增强)
- ✅ **Finage API 集成**: 唯一真实数据源
  - `services/datasource/finage_provider.go` - Finage 数据源提供者
  - 环境变量: `FINAGE_API_KEY` (必须配置)
  - 支持实时报价和聚合历史数据
- ✅ **历史数据同步**: 3 年历史数据入库
  - `cmd/sync_etf_history/main.go` - 历史数据同步脚本
  - 同步 6 个 ETF 的日级别 OHLCV 数据
  - 数据时间范围: 2024-11 至 2026-04 (约 388 天)
- ✅ **实时数据更新**: 定时任务自动更新
  - `POST /api/etf/update-realtime` - 手动刷新实时数据
  - 数据入库: ETFData 表存储完整 OHLCV 数据
  - 冲突处理: 使用 UPSERT 避免重复数据
- ✅ **数据质量保证**:
  - 字段完整性: Open, High, Low, Close, Volume 全部入库
  - 数据源标识: 记录数据来源 (finage)
  - 历史数据计算: 基于真实历史数据计算年化收益率、波动率等指标

#### 技术指标库 (v2.5 新增)
- ✅ **RSI (相对强弱指标)**: 基于价格变动的动量指标
  - `services/technical_indicators.go`
  - 支持 6/12/24 日周期
  - 超买(>70)/超卖(<30)信号识别
- ✅ **MACD (指数平滑异同平均线)**: 趋势跟踪动量指标
  - DIF线、DEA线、MACD柱状图
  - 金叉/死叉信号识别
- ✅ **布林带 (Bollinger Bands)**: 波动性通道指标
  - 中轨(SMA20)、上轨(+2σ)、下轨(-2σ)
  - 价格位置百分比计算
- ✅ **移动平均线**: SMA简单移动平均、EMA指数移动平均
  - 支持自定义周期
  - 多周期均线对比

#### 风险模型 (v2.5 新增)
- ✅ **VaR/CVaR计算**: `services/risk_models.go`
  - 历史模拟法: 基于历史收益率分位数
  - 参数法: 基于正态分布假设
  - 支持 95% 和 99% 置信水平
- ✅ **组合风险分解**: 成分VaR、边际VaR分析
- ✅ **风险调整收益**: 夏普比率、索提诺比率、卡尔玛比率
- ✅ **市场风险指标**: Beta、Alpha、最大回撤

#### 回测引擎 (v2.5 新增)
- ✅ **事件驱动架构**: `services/backtest/event_engine.go`
  - 事件总线(EventBus)实现
  - 多种事件类型: 市场事件、订单事件、成交事件、持仓事件
  - 事件处理器链式调用
- ✅ **订单系统**: 完整订单生命周期管理
  - 订单类型: 市价单、限价单、止损单、止盈单
  - 订单状态: 待提交/已提交/部分成交/全部成交/已取消/已拒绝
  - 订单验证和风险控制
- ✅ **高级回测功能**:
  - 滑点模型: 固定滑点/百分比滑点/波动率滑点
  - 手续费模型: 固定费率/阶梯费率
  - 分红再投资: 支持股息自动再投资
  - 再平衡: 定期再平衡和阈值再平衡
  - 止损止盈: 自动触发订单
- ✅ **回测结果分析**:
  - 收益指标: 总收益、年化收益、超额收益
  - 风险指标: 最大回撤、波动率、夏普比率、索提诺比率
  - 交易统计: 胜率、盈亏比、平均持仓时间
  - 净值曲线和回撤曲线
- ✅ **API端点**:
  - `POST /api/backtest/run` - 运行回测
  - `POST /api/backtest/event-driven` - 事件驱动回测
  - `GET /api/backtest/result/:id` - 获取回测结果
  - `GET /api/backtest/compare` - 对比多个策略

#### 组合优化增强 (v2.5 新增)
- ✅ **马科维茨均值-方差优化**: `services/optimization/mpt_optimizer.go`
  - 有效前沿计算
  - 三种优化目标: 最大夏普比率/最小波动率/目标收益
  - 权重约束: 单资产上限/下限
  - 前沿曲线生成
- ✅ **风险平价模型**: `services/optimization/risk_parity.go`
  - 等风险贡献(ERC)
  - 反向波动率加权
  - 风险预算配置
  - 迭代优化算法
- ✅ **Black-Litterman模型**: `services/optimization/black_litterman.go`
  - 市场均衡收益
  - 投资者观点融合
  - 绝对观点/相对观点支持
  - 后验收益分布
- ✅ **API端点**:
  - `POST /api/portfolio/mpt-optimize` - MPT优化
  - `POST /api/portfolio/efficient-frontier` - 有效前沿
  - `POST /api/portfolio/risk-parity` - 风险平价
  - `POST /api/portfolio/black-litterman` - Black-Litterman优化

#### 因子分析模块 (v2.5 新增)
- ✅ **Fama-French模型**: `services/factor/fama_french.go`
  - 三因子模型: 市场因子、规模因子、价值因子
  - 五因子模型: 增加盈利因子、投资因子
  - 因子收益率计算
  - 组合因子暴露度分析
- ✅ **归因分析**:
  - 收益归因: 因子贡献分解
  - 风险归因: 因子风险贡献
  - 主动收益分析
- ✅ **API端点**:
  - `POST /api/factor/fama-french` - Fama-French分析
  - `GET /api/factor/models` - 获取可用模型
  - `GET /api/factor/factors` - 获取因子定义
- ✅ **前端页面**: `FactorAnalysis.tsx`
  - 因子暴露度可视化
  - 归因分析图表
  - 模型选择界面

#### A股ETF数据源扩展 (v2.5 新增)
- ✅ **AKShare数据源**: `services/ashare/akshare_provider.go`
  - Python AKShare服务: `services/ashare/akshare_server.py`
  - ETF列表获取
  - 实时行情数据
  - 历史K线数据
  - 分红历史数据
- ✅ **TuShare数据源**: `services/ashare/tushare_provider.go`
  - 基金基础信息
  - 基金净值数据
  - 日线行情数据
- ✅ **ETF数据服务**: `services/ashare/etf_data_service.go`
  - 数据同步管理
  - 多数据源集成
  - 股息率计算
  - 搜索和筛选
- ✅ **API端点**:
  - `POST /api/a-share/enable-akshare` - 启用AKShare
  - `POST /api/a-share/sync-etf-list` - 同步ETF列表
  - `POST /api/a-share/sync-prices` - 同步价格
  - `GET /api/a-share/search` - 搜索ETF
  - `GET /api/a-share/by-frequency/:frequency` - 按分红频率筛选

#### 跨资产类别ETF支持 (v2.5 新增)
- ✅ **通用ETF模型**: `models/universal_etf.go`
  - 资产类别: 股票/债券/商品/REIT/货币/多资产/另类
  - 地区覆盖: 全球/美国/中国/欧洲/日本/新兴市场/亚太/拉美
  - ETF类型: 指数/行业/因子/主题/主动/杠杆/反向
  - 完整的风险收益指标
- ✅ **ETF数据服务**: `services/etf/universal_etf_service.go`
  - 预置ETF数据: 美股+A股主流ETF
  - 多维度筛选: 资产类别/地区/类型/行业
  - 分布统计: 资产类别分布/地区分布
  - 组合配置建议: 保守/平衡/激进/股息策略
- ✅ **API端点** (`handlers/universal_etf_handler.go`):
  - `POST /api/universal-etf/initialize` - 初始化ETF数据
  - `GET /api/universal-etf` - 获取所有ETF
  - `GET /api/universal-etf/asset-class/:asset_class` - 按资产类别筛选
  - `GET /api/universal-etf/region/:region` - 按地区筛选
  - `GET /api/universal-etf/type/:etf_type` - 按类型筛选
  - `POST /api/universal-etf/filter` - 多条件筛选
  - `POST /api/universal-etf/compare` - ETF对比
  - `GET /api/universal-etf/portfolio-allocation` - 组合配置建议
  - `GET /api/universal-etf/categories` - 获取分类列表

### v2.7 实施进度 (2026-04-25)

#### 数据层模型定义（已完成 100%）

##### 因子数据层模型
- ✅ `models/factor.go` - 因子数据模型
  - `FactorData` - 因子历史数据（factor_name, date, value, data_source）
  - `FactorTimingSignal` - 因子择时信号（MA斜率、Z-score、百分位数、信号强度）
  - 支持Fama-French五因子模型（Mkt-RF, SMB, HML, RMW, CMA）
  - 信号强度枚举：strong_positive, weak_positive, neutral, weak_negative, strong_negative

##### Alpha观点层模型
- ✅ `models/alpha_view.go` - Alpha观点模型
  - `AlphaView` - Alpha观点（资产、观点收益、置信度、类型、方法、有效期）
  - `AlphaViewPerformance` - 观点表现追踪（实际收益、预测误差、验证状态）
  - 观点类型：absolute（绝对）、relative（相对）
  - 观点方法：factor_timing（因子择时）、momentum（动量）、mean_reversion（均值回归）
  - 观点状态：active（活跃）、expired（过期）、validated（已验证）

- ✅ `models/black_litterman.go` - Black-Litterman模型
  - `BlackLittermanConfig` - BL模型配置（风险厌恶、先验类型、Omega方法）
  - `BLPosteriorReturn` - BL后验收益（后验收益、后验权重、后验协方差）
  - 先验类型：equal_weight（等权）、min_variance（最小方差）、market_cap（市值加权）
  - Omega方法：Idzorek、HeLitterman

##### 风险预算层模型
- ✅ `models/risk_budget.go` - 风险预算模型
  - `RiskBudgetConfig` - CVaR预算配置（CVaR限制、置信水平、时间范围、方法）
  - `MonteCarloSimulation` - 蒙特卡洛模拟（模拟次数、时间步数、统计指标）
  - `RiskContribution` - 风险贡献分解（边际风险、风险贡献、百分比贡献）
  - `RiskBudgetExecution` - 预算执行记录（实际CVaR、预算偏差、调整建议）
  - 风险方法：historical（历史法）、parametric（参数法）、monte_carlo（蒙特卡洛）

##### 插件架构层模型
- ✅ `models/plugin.go` - 插件架构模型
  - `PluginRegistry` - 插件注册表（名称、类型、版本、状态、配置Schema）
  - `PluginConfiguration` - 插件配置（参数、优先级、触发条件）
  - `PluginExecutionLog` - 执行日志（输入、输出、耗时、状态）
  - `ModelBenchmarkMatrix` - 模型基准对比（模型组合、回测期间、性能指标）
  - `StrategyExperiment` - 策略实验（实验名称、策略配置、结果、状态）

##### 数据库迁移
- ✅ `migrations/001_add_factor_tables.sql` - 因子数据层迁移脚本
- ✅ `migrations/002_add_alpha_view_tables.sql` - Alpha观点层迁移脚本
- ✅ `migrations/003_add_risk_budget_tables.sql` - 风险预算层迁移脚本
- ✅ `migrations/004_add_plugin_tables.sql` - 插件架构层迁移脚本
- ✅ `migrations/README.md` - 迁移指南文档
- ✅ `models/db.go` - AutoMigrate函数更新，注册所有新模型

#### 服务层实现（进行中 50%）

##### 因子数据服务层（已完成 100%）
- ✅ `services/factor_data_service.go` - 因子数据服务
  - `CreateFactorData()` - 创建因子数据
  - `GetFactorData()` - 获取因子数据（按时间范围）
  - `GetLatestFactorData()` - 获取最新因子数据
  - `BatchCreateFactorData()` - 批量创建因子数据
  - `CalculateTimingSignal()` - 计算因子择时信号
    - 计算60日移动平均斜率
    - 计算当前Z-score
    - 计算历史百分位数
    - 判断信号强度（强/弱 正/负/中性）
    - 计算预期收益和置信度
  - `CreateTimingSignal()` - 创建择时信号
  - `GetTimingSignals()` - 获取择时信号历史
  - `GetLatestTimingSignal()` - 获取最新择时信号

- ✅ `services/factor_data_service_test.go` - 因子数据服务测试
  - 12个测试用例，全部通过 ✅
  - 测试覆盖率：CRUD操作、信号计算、边界条件
  - 测试辅助函数：setupTestDB、cleanupTestDB

##### Alpha观点服务层（已完成 100%）
- ✅ `services/alpha_view_service.go` - Alpha观点服务
  - `AlphaViewService` - Alpha观点管理服务
    - `CreateAlphaView()` - 创建Alpha观点
    - `GetAlphaView()` - 获取Alpha观点
    - `GetActiveAlphaViews()` - 获取活跃观点列表
    - `UpdateAlphaView()` - 更新Alpha观点
    - `DeactivateView()` - 停用观点
    - `GenerateViewFromFactorTiming()` - 从因子择时信号生成观点
    - `RecordViewPerformance()` - 记录观点表现
    - `GetViewPerformance()` - 获取观点表现
    - `validateView()` - 观点验证（类型、置信度、有效期）

  - `BlackLittermanService` - Black-Litterman服务
    - `CreateConfig()` - 创建BL配置
    - `GetConfig()` - 获取BL配置
    - `UpdateConfig()` - 更新BL配置
    - `CalculatePosteriorReturns()` - 计算后验收益
    - `GetPosteriorReturns()` - 获取后验收益
    - `validateConfig()` - 配置验证
    - `parseMarketWeights()` - 解析市场权重JSON
    - `parseCovarianceMatrix()` - 解析协方差矩阵JSON
    - `calculateEquilibriumReturns()` - 计算均衡收益
    - `buildViewMatrices()` - 构建观点矩阵（P, Q, Omega）
    - `blFormula()` - BL公式计算
    - `calculateMatrixInverse()` - 矩阵求逆
    - `calculateMatrixMultiply()` - 矩阵乘法
    - `calculateMatrixTranspose()` - 矩阵转置

##### 风险预算服务层（待实现 0%）
- 📋 `services/risk_budget_service.go` - 风险预算服务（待实现）
  - CVaR计算（历史法、参数法、蒙特卡洛）
  - 风险贡献分解
  - 风险预算优化
  - 蒙特卡洛模拟

##### 插件管理服务层（待实现 0%）
- 📋 `services/plugin_service.go` - 插件管理服务（待实现）
  - 插件注册机制
  - 插件配置管理
  - 插件执行引擎

#### 前端实施方案（已完成文档）

##### 完整实施方案文档
- ✅ `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md` - 前后端一体化实施方案
  - **前端页面改造方案**：4个新页面 + 2个现有页面增强
    - `FactorTiming.tsx` - 因子择时信号分析页面
    - `AlphaViews.tsx` - Alpha观点管理页面
    - `BlackLittermanConfig.tsx` - BL模型配置页面
    - `RiskBudget.tsx` - 风险预算管理页面
    - `PortfolioOptimization.tsx` 增强 - 集成BL优化器
    - `FactorAnalysis.tsx` 增强 - 添加择时信号展示

  - **后端API接口规范**：15个API端点完整定义
    - `POST /api/factor/timing/calculate` - 计算择时信号
    - `GET /api/factor/timing/history` - 获取信号历史
    - `POST /api/alpha-views` - 创建Alpha观点
    - `GET /api/alpha-views/active` - 获取活跃观点
    - `POST /api/alpha-views/generate-from-signal` - 从信号生成观点
    - `POST /api/black-litterman/configs` - 创建BL配置
    - `POST /api/black-litterman/calculate` - 计算后验收益
    - `POST /api/risk-budget/configs` - 创建风险预算配置
    - `POST /api/risk-budget/calculate-cvar` - 计算CVaR
    - `POST /api/risk-budget/monte-carlo` - 运行蒙特卡洛模拟
    - `POST /api/kimi/market-analysis` - AI市场分析
    - `POST /api/kimi/alpha-view` - AI观点生成

  - **数据交互协议**：统一请求/响应格式
    - 成功响应：`{ success: true, data: {...}, message: "..." }`
    - 错误响应：`{ success: false, error: "...", error_code: "...", details: {...} }`
    - 分页格式：`{ success: true, data: [...], pagination: {...} }`

  - **TypeScript类型定义**：完整类型系统
    - `FactorTimingSignal` - 因子择时信号类型
    - `AlphaView` - Alpha观点类型
    - `BlackLittermanConfig` - BL配置类型
    - `RiskBudgetConfig` - 风险预算配置类型
    - `MonteCarloSimulation` - 蒙特卡洛模拟类型

  - **API服务封装**：前端API调用封装
    - `factorTimingAPI` - 因子择时API
    - `alphaViewAPI` - Alpha观点API
    - `blackLittermanAPI` - Black-Litterman API
    - `riskBudgetAPI` - 风险预算API
    - `kimiAPI` - Kimi AI API

  - **前后端联调方案**：开发环境配置、测试流程
    - 后端配置：SERVER_PORT, CORS_ALLOWED_ORIGINS, KIMI_API_KEY
    - 前端配置：VITE_API_BASE_URL, VITE_KIMI_API_KEY
    - 单元测试：后端Handler测试、前端组件测试
    - 集成测试：端到端测试脚本
    - API测试：HTTP测试用例

  - **兼容性处理策略**：向后兼容、渐进式增强
    - API版本管理：`/api/v1` 和 `/api/v2` 并存
    - 数据库迁移：渐进式迁移，保留旧表结构
    - 前端特性检测：动态检测API功能支持
    - 浏览器兼容：Polyfill配置、CSS前缀

  - **分阶段实施计划**：6个阶段，共6周
    - 阶段一：基础设施准备（1周）
    - 阶段二：因子择时模块（1周）
    - 阶段三：Alpha观点模块（1周）
    - 阶段四：Black-Litterman模块（1.5周）
    - 阶段五：风险预算模块（1.5周）
    - 阶段六：集成测试与优化（1周）

  - **Kimi2.6接口接入指南**：完整的AI集成方案
    - `KimiService` - Kimi API调用封装
    - `GenerateMarketAnalysis()` - 市场分析生成
    - `GenerateAlphaView()` - Alpha观点生成
    - 前端集成：AI分析按钮、实时结果展示

#### 进度统计

| 模块 | 完成度 | 详情 |
|------|--------|------|
| 数据模型 | 100% | 15个模型已定义 |
| 数据库迁移 | 100% | 4个迁移脚本已创建 |
| 服务层接口 | 50% | 2/4个服务层已实现 |
| 测试用例 | 25% | 12个测试用例已编写 |
| 前端方案 | 100% | 完整实施方案文档 |
| **总体进度** | **37.5%** | Phase 1 进行中 |

#### 关键成果

##### 闭环一：Alpha模型 + Black-Litterman（已实现后端）
1. **因子择时信号生成** ✅
   - 计算60日移动平均斜率
   - 计算当前Z-score
   - 生成预期收益和信心水平

2. **Alpha观点格式** ✅
   - `[资产, 观点收益, 信心水平]`
   - 支持绝对和相对观点

3. **Black-Litterman集成** ✅
   - 市场隐含均衡收益计算
   - 观点误差矩阵Ω构建
   - 后验收益计算

##### 闭环二：风险预算 + 风险平价（待实现）
1. **CVaR计算** 📋 待实现
   - 蒙特卡洛模拟
   - 历史模拟法
   - 参数法

2. **风险贡献分解** 📋 待实现
   - 欧拉分配：`RC_i = w_i × ∂CVaR/∂w_i`
   - 预算约束优化

3. **风险预算优化** 📋 待实现
   - 最小化预算偏差
   - 偏度约束

#### 下一步计划

**本周目标（2026-04-25 ~ 2026-05-01）**：✅ 已完成
- ✅ 实现因子数据服务层
- ✅ 实现Alpha观点服务层
- ✅ 编写测试用例
- ✅ 创建前后端一体化实施方案

**下周目标（2026-05-02 ~ 2026-05-08）**：
- 编写Alpha观点服务层测试用例
- 实现风险预算服务层接口
- 编写风险预算模型测试用例
- 开始前端页面开发（Kimi负责）

#### 文档更新
- ✅ `docs/DOCUMENT_INDEX.md` - 添加前后端实施方案文档
- ✅ `docs/roadmap/EVOLUTION_ROADMAP_2026.md` - 标记v2.7进度
- ✅ `docs/development/v2.7_phase1_progress.md` - Phase 1进度文档
- ✅ `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md` - 前后端实施方案

### v2.6.1 更新内容 (2026-04-25)

#### 前后端接口一致性修复
- ✅ **MPT优化API**: Returns/CovMatrix参数改为可选，支持自动从历史数据计算
- ✅ **有效前沿API**: 同样优化，Returns/CovMatrix可选
- ✅ **字段名对齐**: 前端类型定义与后端响应完全一致
  - `weights` (非 `optimal_weights`)
  - `volatility` (非 `expected_risk`)
  - `target_return`, `min_volatility`, `optimal_weights`, `sharpe_ratio`
- ✅ **有效前沿图表修复**:
  - 动态坐标轴domain计算
  - tickFormatter格式化显示
  - 图表边距优化（bottom:50, left:60）
  - tooltip格式化函数改进

#### 文档一致性维护
- ✅ **移除JWT残留**: 所有文档中已移除登录/JWT认证相关描述
  - REVIEW_SUMMARY.md, CODE_REVIEW_REPORT.md
  - SECURITY_AUDIT.md, EVOLUTION_ROADMAP_2026.md
  - PROFESSIONAL_ENHANCEMENT.md, TEST_COVERAGE_PLAN.md
  - archive/overview.md

### v2.4 更新内容

#### v2.4 更新内容

#### 基础设施升级
- ✅ **审计日志**: 异步写入，敏感信息自动脱敏
  - `middleware/audit.go` - 审计日志中间件
  - 自动记录所有API请求（method/path/IP/statusCode）
  - 敏感字段脱敏（password/token/secret/api_key）
  - Request ID 追踪，支持分布式日志追踪
- ✅ **数据验证**: 通用验证中间件，支持多种类型
  - `middleware/validation.go` - 输入验证中间件
  - 支持 string/number/email 类型验证
  - 支持 Min/Max/Pattern/Enum 约束
  - `ValidateSymbol()` - 股票代码格式验证，防止注入攻击
- ✅ **速率限制**: IP级别的请求频率限制
  - `RateLimiterHandler()` - 滑动窗口限流算法
  - 防止滥用和DDoS攻击
- ✅ **API分页**: 通用分页响应结构
  - `models/pagination.go` - 分页模型
  - 支持 page/pageSize 参数，最大100条/页
  - 返回 total/totalPages/hasNext/hasPrev

#### A股红利ETF价格功能
- ✅ **实时价格获取**: 支持A股ETF实时价格查询
- ✅ **价格刷新接口**: 手动刷新ETF价格数据
- ✅ **前端价格展示**: 当前价格、涨跌幅、成交量

### v2.3 更新内容

#### 汇率服务优化
- ✅ **多数据源故障转移**: 支持 Open Exchange Rates、CurrencyAPI、Frankfurter 三个数据源
- ✅ **竞态条件修复**: 数据源管理器中的并发访问问题已解决
- ✅ **健康监控**: 自动数据源可用性检查和故障切换

#### 投资组合优化API
- ✅ **投资组合优化接口**: POST /api/portfolio/optimize
- ✅ **有效前沿生成**: POST /api/portfolio/efficient-frontier
- ✅ **支持三种优化类型**: max_sharpe, min_volatility, equal_weight
- ✅ **梯度下降算法**: 高效的组合优化计算

#### 金融算法修复
- ✅ **夏普比率计算**: 修复单位混用问题，正确实现公式
- ✅ **最大回撤注释**: 添加明确的返回值说明
- ✅ **股息率TODO**: 添加TODO注释说明需从数据库获取

#### 代码质量提升
- ✅ **ESLint 零错误**: 修复全部 ESLint 问题
- ✅ **TypeScript 类型安全**: 消除所有 `any` 类型
- ✅ **Go 代码格式化**: 统一代码风格
- ✅ **React Hooks 规范**: 修复 `exhaustive-deps` 警告

#### 性能优化
- ✅ **缓存策略优化**: 分级缓存、智能过期策略
- ✅ **数据库优化**: 索引优化、查询性能提升
- ✅ **并发处理**: 竞态条件修复、线程安全保证

#### Pre-commit钩子
- ✅ **TypeScript类型检查**: `tsc --noEmit` 提交前检查
- ✅ **ESLint代码检查**: 前端代码质量保证
- ✅ **Go格式化检查**: `gofmt` 代码风格统一
- ✅ **失败阻止提交**: 语法错误无法提交

---

## 🛠️ 开发规范

### 代码规范
- **Go**: 遵循 Go 官方代码规范，使用 `gofmt` 格式化
- **TypeScript**: 严格类型检查，禁用 `any` 类型
- **React**: 函数式组件，Hooks 规范使用
- **命名规范**: 驼峰命名法，语义化命名

### 提交规范
- **提交信息**: 遵循 conventional commits 规范
- **代码审查**: 所有修改必须经过代码审查
- **测试要求**: 新功能必须包含单元测试

### 文档要求
- **API 文档**: OpenAPI 3.0 规范
- **代码注释**: 重要函数和复杂逻辑必须注释
- **更新同步**: 架构修改必须更新 AGENTS.md

---

## � 完整开发流程

### 阶段概览

```
需求分析 → 系统设计 → 编码实现 → 单元测试 → 集成测试 → 系统测试 → 部署上线 → 后期维护
   ↓           ↓           ↓           ↓           ↓           ↓           ↓
  PRD        架构设计     代码规范   覆盖率≥80%   接口测试    UAT测试    CI/CD    监控告警
```

---

### 阶段一：需求分析与规划

#### 1.1 需求采集
| 项目 | 内容 |
|------|------|
| **输入** | 用户反馈、业务需求、市场分析 |
| **输出** | 需求列表 (Issue/Ticket) |
| **责任人** | 产品经理 |

#### 1.2 需求评审
| 项目 | 内容 |
|------|------|
| **输入** | 需求列表 |
| **输出** | 评审通过的需求文档 |
| **标准** | 需求明确、可测试、可实现 |
| **责任人** | 技术负责人 + 产品 + 开发 |

#### 1.3 任务拆分
| 项目 | 内容 |
|------|------|
| **输入** | 评审通过的需求 |
| **输出** | 技术任务列表 (GitHub Issues) |
| **标准** | 每个任务 ≤ 2天工作量 |
| **责任人** | 技术负责人 |

---

### 阶段二：系统设计

#### 2.1 架构设计
| 项目 | 内容 |
|------|------|
| **输入** | 技术任务 |
| **输出** | 架构设计文档、数据库设计、API设计 |
| **标准** | 符合项目技术栈和编码规范 |
| **责任人** | 高级工程师/架构师 |

#### 2.2 设计评审
| 项目 | 内容 |
|------|------|
| **输入** | 架构设计文档 |
| **输出** | 评审通过的设计 |
| **标准** | 可行性、可扩展性、安全性 |
| **责任人** | 技术负责人 + 团队评审 |

#### 2.3 输出标准
```
✅ 数据库设计文档 (ER图)
✅ API接口文档 (OpenAPI 3.0)
✅ 详细设计文档 (接口定义、数据结构)
```

---

### 阶段三：编码实现

#### 3.1 分支管理
```bash
# 功能分支命名
feature/功能描述-日期
bugfix/问题描述-日期

# 示例
feature/portfolio-optimizer-20260413
bugfix/fix-sharpe-ratio-20260413
```

#### 3.2 编码规范
| 语言 | 规范 |
|------|------|
| **Go** | gofmt格式化、错误处理、注释完整 |
| **TypeScript** | 严格类型、禁用any、Hooks规范 |
| **React** | 函数组件、Hooks规范、组件分离 |

#### 3.2.1 中间件使用规范
```go
// 审计日志（自动记录所有请求）
router.Use(middleware.AuditLogger())

// 数据验证
router.POST("/api/etf",
    middleware.ValidateInput([]middleware.ValidationRule{
        {Field: "symbol", Type: "string", Required: true, Min: 1, Max: 20},
        {Field: "name", Type: "string", Required: true, Min: 1, Max: 100},
        {Field: "price", Type: "number", Required: true, Min: 0},
    }),
    handler.CreateETF,
)

// 速率限制（每IP每分钟100请求）
router.Use(middleware.RateLimiterHandler(100, time.Minute))

// 股票代码验证（防止注入）
router.GET("/api/etf/:symbol", middleware.ValidateSymbol(), handler.GetETF)
```

#### 3.3 代码审查 (Code Review)
| 项目 | 内容 |
|------|------|
| **触发条件** | Pull Request创建 |
| **审查清单** | 代码规范、安全漏洞、逻辑错误、测试覆盖 |
| **通过条件** | 至少1人Approval + 所有CI通过 |
| **审查内容** | [CODE MODIFICATION CHECKLIST](#代码修改确认) |

#### 3.4 提交规范
```bash
# 格式
<type>(<scope>): <subject>

# 示例
feat(portfolio): 添加投资组合优化API
fix(etf): 修复夏普比率计算单位问题
docs(readme): 更新README文档
```

---

### 阶段四：单元测试

#### 4.1 测试覆盖率要求
| 模块 | 覆盖率目标 |
|------|-----------|
| **核心业务逻辑** | ≥ 80% |
| **工具函数** | ≥ 90% |
| **Handlers/API层** | ≥ 70% |
| **前端组件** | 快照测试+关键交互测试 |

#### 4.2 测试规范
```go
// Go单元测试示例
func TestCalculateSharpeRatio(t *testing.T) {
    tests := []struct {
        name     string
        input    decimal.Decimal
        expected decimal.Decimal
    }{
        {"正常情况", decimal.NewFromFloat(0.15), decimal.NewFromFloat(0.8)},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateSharpeRatio(tt.input)
            if !result.Equal(tt.expected) {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

#### 4.3 运行测试
```bash
# Go测试
go test -v -cover ./...

# 前端测试
npm run test
```

---

### 阶段五：集成测试

#### 5.1 接口测试
| 项目 | 内容 |
|------|------|
| **测试范围** | API接口、数据流、模块间调用 |
| **工具** | Postman/curl/自动化测试脚本 |
| **通过标准** | 所有用例通过、响应时间<200ms |

#### 5.2 集成测试清单
```
✅ 后端API接口测试
✅ 数据库读写测试
✅ 缓存功能测试
✅ 多数据源故障转移测试
✅ 前端API调用测试
```

---

### 阶段六：系统测试

#### 6.1 功能测试
| 项目 | 内容 |
|------|------|
| **测试方式** | 手动测试 + 自动化E2E测试 |
| **测试用例** | 覆盖所有用户场景 |
| **通过标准** | 零严重/高优先级Bug |

#### 6.2 性能测试
```
指标要求:
- API响应时间 < 200ms (P95)
- 系统可用性 ≥ 99.9%
- 并发用户数 ≥ 100
```

#### 6.3 安全测试
```
✅ SQL注入防护测试
✅ XSS攻击防护测试
✅ 敏感数据加密测试
```

---

### 阶段七：部署上线

#### 7.1 部署前检查
```bash
# CI/CD检查清单
✅ 代码规范检查通过 (pre-commit)
✅ 单元测试覆盖率 ≥ 80%
✅ 所有CI构建成功
✅ 安全扫描无高危漏洞
✅ 功能测试通过
```

#### 7.2 部署流程
```bash
# 1. 合并到main分支
git checkout main
git merge feature/xxx

# 2. CI/CD自动流程
make build    # 编译打包
make test     # 运行测试
make deploy   # 部署到环境

# 3. 部署验证
curl http://localhost:8080/health
```

#### 7.3 回滚机制
```bash
# 回滚命令
kubectl rollout undo deployment/etf-insight
# 或
docker-compose down && docker-compose -f backup.yml up
```

---

### 阶段八：后期维护

#### 8.1 监控告警
| 监控项 | 告警阈值 |
|--------|----------|
| **API错误率** | > 1% |
| **响应时间** | > 500ms |
| **CPU使用率** | > 80% |
| **内存使用率** | > 85% |

#### 8.2 问题处理
```
问题等级:
P0 - 系统宕机 → 立即处理，15分钟内响应
P1 - 核心功能故障 → 4小时内响应
P2 - 非核心功能异常 → 24小时内响应
P3 - 优化改进 → 纳入迭代计划
```

#### 8.3 变更管理
```
变更申请 → 技术评审 → 变更审批 → 实施变更 → 变更验证
    ↓           ↓           ↓           ↓           ↓
  Issue    代码审查    技术负责人   CI/CD部署   监控验证
```

---

### 阶段衔接机制

#### 流程检查点
| 检查点 | 触发条件 | 负责人 |
|--------|----------|--------|
| **需求冻结** | 需求评审通过 | 产品经理 |
| **设计冻结** | 设计评审通过 | 技术负责人 |
| **代码冻结** | CR通过+测试通过 | 开发负责人 |
| **发布评审** | 系统测试通过 | 技术负责人 |
| **上线确认** | 部署验证通过 | 运维负责人 |

#### 文档交付清单
```
阶段          交付文档
─────────────────────────────────
需求分析      PRD、用户故事
系统设计      架构设计、数据库设计、API文档
编码实现      代码、单元测试、代码审查记录
集成测试      测试用例、测试报告
系统测试      测试报告、Bug修复记录
部署上线      部署手册、配置清单
维护          运维手册、监控配置、故障复盘
```

---

### 质量控制门禁

| 门禁 | 检查项 | 通过标准 |
|------|--------|----------|
| **代码门禁** | ESLint/tsc/gofmt | 零错误 |
| **测试门禁** | 单元测试覆盖率 | ≥ 80% |
| **安全门禁** | 安全扫描 | 无高危漏洞 |
| **构建门禁** | CI构建 | 全部通过 |
| **部署门禁** | 功能验证 | 核心功能正常 |

---

## 📊 金融算法标准

### 算法规范总则

| 项目 | 要求 |
|------|------|
| **精度** | 使用decimal.Decimal，避免浮点数精度问题 |
| **单位** | 收益率统一使用百分比，波动率使用年化值 |
| **边界** | 除零保护，返回零值或明确错误 |
| **数据来源** | 优先使用数据库真实历史数据，不足时使用默认值 |

---

### 0. 投资组合分析核心公式

#### 0.1 对数收益率计算
```
r_t = ln(P_t / P_{t-1})

其中:
- P_t: 第t日收盘价
- P_{t-1}: 第t-1日收盘价
- r_t: 第t日对数收益率
```

#### 0.2 组合预期收益率
```
R_p = Σ w_i * R_i

其中:
- R_p: 组合预期年化收益率
- w_i: 第i个资产的权重
- R_i: 第i个资产的预期年化收益率
```

#### 0.3 组合方差 (考虑相关性)
```
σ_p² = Σ Σ w_i * w_j * σ_i * σ_j * ρ_ij

其中:
- σ_p²: 组合方差
- w_i, w_j: 第i,j个资产的权重
- σ_i, σ_j: 第i,j个资产的年化波动率
- ρ_ij: 第i,j个资产的相关系数
```

#### 0.4 蒙特卡洛模拟 (几何布朗运动)
```
S_t = S_0 * exp((μ - σ²/2) * t + σ * √t * Z)

其中:
- S_t: t时刻的资产价格
- S_0: 初始资产价格
- μ: 预期年化收益率
- σ: 年化波动率
- t: 时间(年)
- Z: 标准正态分布随机变量
```

#### 0.5 VaR (参数法)
```
VaR_α = μ + Z_α * σ

其中:
- VaR_α: 置信水平α下的风险价值
- μ: 收益率均值
- Z_α: 标准正态分布的分位数 (95%: -1.645, 99%: -2.326)
- σ: 收益率标准差
```

#### 0.6 CVaR (参数法)
```
CVaR_α = μ - σ * φ(Z_α) / Φ(Z_α)

其中:
- φ: 标准正态分布PDF
- Φ: 标准正态分布CDF
```

---

### 1. 夏普比率 (Sharpe Ratio)

#### 标准公式
```
SR = (Rp - Rf) / σp

其中:
- Rp: 投资组合年化收益率
- Rf: 年化无风险利率 (默认4%)
- σp: 年化波动率
```

#### 实现规范
```go
// ✅ 正确：完整注释说明
// SharpeRatio 计算夏普比率
// 公式: (年化收益率 - 无风险利率) / 年化波动率
// 参数:
//   - avgDailyReturn: 平均日收益率（百分比形式，如5表示5%）
//   - volatility: 年化波动率（百分比形式，如15表示15%）
//   - riskFreeRate: 年化无风险利率（默认4%）
// 返回: 夏普比率（无纲量）
// 注意: 当波动率为0时返回0
func CalculateSharpeRatio(avgDailyReturn, volatility, riskFreeRate decimal.Decimal) decimal.Decimal {
    if volatility.IsZero() {
        return decimal.Zero
    }
    // avgDailyReturn 是百分比，需要转换为小数
    avgDailyReturnDecimal := avgDailyReturn.Div(decimal.NewFromInt(100))
    // 年化收益率 = 日均收益率 * 252（交易天数）
    annualizedReturn := avgDailyReturnDecimal.Mul(decimal.NewFromInt(252)).Mul(decimal.NewFromInt(100))
    // 计算超额收益
    excessReturn := annualizedReturn.Sub(riskFreeRate)
    // 波动率转换为小数
    volatilityDecimal := volatility.Div(decimal.NewFromInt(100))
    return excessReturn.Div(volatilityDecimal)
}
```

#### 验证方法
| 测试用例 | 输入 | 预期输出 |
|----------|------|----------|
| 正常情况 | avgReturn=5%, vol=15%, rf=4% | SR ≈ 0.42 |
| 零波动率 | vol=0 | 返回0 |
| 负超额收益 | avgReturn=2%, vol=15%, rf=4% | SR < 0 |

---

### 2. 最大回撤 (Maximum Drawdown)

#### 标准公式
```
MDD = (Trough - Peak) / Peak × 100%

返回负数百分比，表示从峰值到谷底的最大跌幅
```

#### 实现规范
```go
// ✅ 正确：清晰注释
// calculateMaxDrawdown 计算最大回撤
// 返回负数百分比，表示从峰值到谷底的下跌幅度
// 例如：返回 -8.5 表示从峰值下跌了 8.5%
func calculateMaxDrawdown(prices []models.ETFData) decimal.Decimal {
    if len(prices) == 0 {
        return decimal.Zero
    }

    maxDrawdown := decimal.Zero
    peak := prices[0].ClosePrice

    for _, price := range prices {
        // 更新峰值
        if price.ClosePrice.GreaterThan(peak) {
            peak = price.ClosePrice
        }

        // 计算回撤
        if peak.IsPositive() {
            drawdown := peak.Sub(price.ClosePrice).Div(peak).Mul(decimal.NewFromInt(100))
            if drawdown.GreaterThan(maxDrawdown) {
                maxDrawdown = drawdown
            }
        }
    }

    return maxDrawdown.Neg() // 返回负值表示回撤
}
```

#### 验证方法
| 测试用例 | 数据特征 | 预期输出 |
|----------|----------|----------|
| 正常回撤 | 10%跌幅后回升 | MDD ≈ -10% |
| 无回撤 | 持续上涨 | MDD = 0 |
| 完全回撤 | 跌至0 | MDD = -100% |

---

### 3. 索提诺比率 (Sortino Ratio)

#### 标准公式
```
Sortino = (Rp - Rf) / σd

其中:
- Rp: 投资组合年化收益率
- Rf: 年化无风险利率
- σd: 下行标准差 (只考虑负收益)
```

#### 下行标准差计算
```
σd = sqrt( Σ(min(ri - τ, 0)²) / n )

其中:
- ri: 第i期收益率
- τ: 目标收益率 (通常设为0)
- n: 总期数
```

#### 实现规范
```go
// CalculateSortinoRatio 计算索提诺比率
// 只考虑下行风险，比夏普比率更适合评估不对称收益
func (s *PortfolioAnalyticsService) CalculateSortinoRatio(returns []float64, riskFreeRate float64) float64 {
    if len(returns) == 0 {
        return 0
    }

    meanReturn := s.mean(returns)
    targetReturn := 0.0 // 最小可接受收益率

    // 计算下行偏差
    downsideDeviations := make([]float64, 0)
    for _, r := range returns {
        if r < targetReturn {
            downsideDeviations = append(downsideDeviations, (r-targetReturn)*(r-targetReturn))
        }
    }

    if len(downsideDeviations) == 0 {
        return 0 // 没有下行风险
    }

    downsideVariance := 0.0
    for _, d := range downsideDeviations {
        downsideVariance += d
    }
    downsideStd := math.Sqrt(downsideVariance / float64(len(returns)))

    if downsideStd == 0 {
        return 0
    }

    // 年化处理
    annualReturn := meanReturn * 252
    annualDownsideStd := downsideStd * math.Sqrt(252)

    return (annualReturn - riskFreeRate) / annualDownsideStd
}
```

---

### 4. 卡尔玛比率 (Calmar Ratio)

#### 标准公式
```
Calmar = Rp / |MDD|

其中:
- Rp: 投资组合年化收益率
- MDD: 最大回撤 (取绝对值)
```

#### 实现规范
```go
// CalculateCalmarRatio 计算卡尔玛比率
// 衡量收益与最大回撤的比值，适合评估风险调整后的收益
func (s *PortfolioAnalyticsService) CalculateCalmarRatio(annualReturn, maxDrawdown float64) float64 {
    if maxDrawdown == 0 {
        return 0
    }
    return annualReturn / maxDrawdown
}
```

---

### 5. 偏度与峰度

#### 偏度 (Skewness)
```
Skewness = [n / ((n-1)(n-2))] * Σ((xi - μ) / σ)³

衡量收益分布的不对称性:
- 正偏度: 右尾较长，极端正收益概率高
- 负偏度: 左尾较长，极端负收益概率高
```

#### 峰度 (Kurtosis)
```
Excess Kurtosis = [Σ((xi - μ) / σ)⁴ / n] - 3

衡量收益分布的尾部厚度:
- 正超额峰度: 厚尾，极端事件概率高
- 负超额峰度: 薄尾，收益更集中在均值附近
```

---

### 6. 滚动窗口指标

#### 滚动窗口计算
```
对于窗口期 n (如30日、60日、90日、180日、252日):

年化收益率 = (Σ日收益率) × (252/n)
年化波动率 = σ_日 × √252
夏普比率 = (年化收益率 - Rf) / 年化波动率
卡尔玛比率 = 年化收益率 / 最大回撤
索提诺比率 = (年化收益率 - Rf) / 下行标准差
```

#### 交易统计指标
```
胜率 = 盈利交易数 / 总交易数
平均盈利 = 总盈利 / 盈利交易数
平均亏损 = 总亏损 / 亏损交易数
盈亏比 = 总盈利 / 总亏损
```

#### 实现规范
```go
// RollingWindowMetrics 滚动窗口指标
type RollingWindowMetrics struct {
    WindowDays   int     `json:"window_days"`   // 窗口天数
    AnnualReturn float64 `json:"annual_return"` // 年化收益率
    Volatility   float64 `json:"volatility"`    // 年化波动率
    SharpeRatio  float64 `json:"sharpe_ratio"`  // 夏普比率
    MaxDrawdown  float64 `json:"max_drawdown"`  // 最大回撤
    CalmarRatio  float64 `json:"calmar_ratio"`  // 卡尔玛比率
    SortinoRatio float64 `json:"sortino_ratio"` // 索提诺比率
    VaR95        float64 `json:"var_95"`        // 95% VaR
    WinRate      float64 `json:"win_rate"`      // 胜率
    AvgGain      float64 `json:"avg_gain"`      // 平均盈利
    AvgLoss      float64 `json:"avg_loss"`      // 平均亏损
    ProfitFactor float64 `json:"profit_factor"` // 盈亏比
}

// CalculateAllRollingWindows 计算所有常用滚动窗口
func (s *PortfolioAnalyticsService) CalculateAllRollingWindows(
    returns []float64,
    prices []decimal.Decimal,
) map[int]*RollingWindowMetrics {
    windows := []int{30, 60, 90, 180, 252} // 252个交易日≈1年
    // ... 实现
}
```

---

### 7. 股息再投资模型

#### 季度再投资模型
```
每季度末:
1. 计算季度收益: S_t = S_{t-1} × exp((μ - σ²/2)×Δt + σ×√Δt×Z)
2. 支付季度股息: D_t = S_t × (年化股息率 / 4)
3. 再投资: S_t' = S_t + D_t

其中:
- Δt = 0.25 (1个季度)
- 股息立即再投资，产生复利效应
```

#### 月度再投资模型 (更精确)
```
每月末:
1. 计算月度收益: S_t = S_{t-1} × exp((μ - σ²/2)×Δt + σ×√Δt×Z)
2. 判断是否为股息月 (通常3/6/9/12月)
3. 如果是股息月: D_t = S_t × (年化股息率 / 4)
4. 再投资: S_t' = S_t + D_t

其中:
- Δt = 1/12 (1个月)
- 季度股息累积到支付月一次性支付
```

#### 复利效应
```
再投资 vs 不复投资的差异:

不复投资: 终值 = P × (1 + r)^n + Σ(股息)
再投资:   终值 = P × (1 + r + d)^n

其中:
- P: 初始投资
- r: 年化收益率
- d: 年化股息率
- n: 年数

长期持有(10年以上)，再投资的复利效应显著
```

---

### 8. 其他指标

| 指标 | 公式 | 精度要求 |
|------|------|----------|
| **Calmar** | 年化收益 / \|MDD\| | 保留4位小数 |
| **Sortino** | (Rp-Rf) / 下行标准差 | 保留4位小数 |
| **Profit Factor** | 总盈利 / 总亏损 | 保留2位小数 |
| **Win Rate** | 盈利交易数 / 总交易数 | 保留2位小数 |

---

### 算法测试覆盖率要求

| 模块 | 覆盖率目标 |
|------|-----------|
| **夏普比率** | 100% (包括边界) |
| **最大回撤** | 100% (包括边界) |
| **其他指标** | ≥ 90% |

---

## 🔒 安全规范

### 1. 安全边界

| 边界类型 | 要求 |
|----------|------|
| **输入验证** | 所有用户输入必须验证 |
| **SQL注入** | 使用参数化查询 |
| **XSS攻击** | React默认转义+ CSP头 |
| **权限控制** | 最小权限原则 |

### 2. 数据加密标准

| 数据类型 | 加密方式 |
|----------|----------|
| **API Key** | 环境变量，代码中不出现 |
| **敏感日志** | 脱敏处理 |
| **数据库** | TLS传输 |

### 3. 安全审计

| 审计项 | 频率 | 记录 |
|--------|------|------|
| **操作日志** | 每次 | 动作、资源 |
| **错误日志** | 每次 | 错误类型、堆栈 |
| **安全扫描** | 每周 | 漏洞报告 |

### 4. API Key 管理
```go
// ✅ 正确：从环境变量读取
apiKey := os.Getenv("FINAGE_API_KEY")

// ❌ 错误：硬编码
apiKey := "your_api_key_here"
```

### 2. CORS 配置
```go
// ✅ 正确：限制允许的域名
allowedOrigins := []string{
    "http://localhost:3000",
    "http://localhost:8080",
    "https://yourdomain.com",
}

// ❌ 错误：允许所有域名
c.Header("Access-Control-Allow-Origin", "*")
```

### 3. 输入验证
```go
// ✅ 正确：验证输入
type CreateETFRequest struct {
    Symbol       string  `json:"symbol" binding:"required,min=1,max=10"`
    Name         string  `json:"name" binding:"required,min=1,max=100"`
    ExpenseRatio float64 `json:"expense_ratio" binding:"gte=0,lte=10"`
}

if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
```

### 4. 日志脱敏
```go
// ✅ 正确：脱敏敏感信息
utils.Info("API request", "url", sanitizeURL(reqURL))

func sanitizeURL(url string) string {
    re := regexp.MustCompile(`apikey=[^&\s]+`)
    return re.ReplaceAllString(url, "apikey=***")
}

// ❌ 错误：记录完整 URL（包含 API Key）
utils.Info("API request", "url", reqURL)
```

---

## 🗄️ 数据操作日志规范

### 设计原则
- **不删除数据**: 所有数据操作采用软删除或更新，保留历史记录
- **完整审计**: 记录所有数据变更的操作人、时间、前后值
- **可追溯**: 支持按时间、表名、操作类型查询

### 数据操作日志表 (data_operation_logs)

```go
// DataOperationLog 数据操作日志模型
type DataOperationLog struct {
    ID            uint           `json:"id" gorm:"primaryKey"`
    TableName     string         `json:"table_name" gorm:"index;size:64"`      // 表名: etf_configs, portfolio_configs等
    OperationType string         `json:"operation_type" gorm:"index;size:20"`  // 操作类型: CREATE, UPDATE, DELETE
    RecordID      uint           `json:"record_id" gorm:"index"`               // 被操作记录ID
    OldValues     datatypes.JSON `json:"old_values" gorm:"type:json"`          // 操作前的数据(JSON)
    NewValues     datatypes.JSON `json:"new_values" gorm:"type:json"`          // 操作后的数据(JSON)
    Diff          datatypes.JSON `json:"diff" gorm:"type:json"`                // 变更字段对比
    OperatorID    uint           `json:"operator_id" gorm:"index"`             // 操作人ID
    OperatorName  string         `json:"operator_name" gorm:"size:64"`         // 操作人名称
    OperatorIP    string         `json:"operator_ip" gorm:"size:64"`           // 操作人IP
    RequestID     string         `json:"request_id" gorm:"size:64"`            // 请求ID(关联审计日志)
    Reason        string         `json:"reason" gorm:"size:255"`               // 操作原因/备注
    CreatedAt     time.Time      `json:"created_at"`
}
```

### 自动记录机制

#### 1. GORM Hook 自动记录
```go
// 在模型定义中启用审计
func (e *ETFConfig) AfterCreate(tx *gorm.DB) error {
    return logDataChange(tx, "etf_configs", "CREATE", e.ID, nil, e, "")
}

func (e *ETFConfig) AfterUpdate(tx *gorm.DB) error {
    // 获取旧值
    var oldETF ETFConfig
    tx.Unscoped().First(&oldETF, e.ID)
    return logDataChange(tx, "etf_configs", "UPDATE", e.ID, &oldETF, e, "")
}

func (e *ETFConfig) AfterDelete(tx *gorm.DB) error {
    return logDataChange(tx, "etf_configs", "DELETE", e.ID, e, nil, "")
}
```

#### 2. 手动记录辅助函数
```go
// LogDataChange 记录数据变更
func LogDataChange(
    db *gorm.DB,
    tableName string,
    operationType string,
    recordID uint,
    oldValues interface{},
    newValues interface{},
    reason string,
) error {
    // 从context获取操作人信息
    operatorID, _ := db.Statement.Context.Value("operator_id").(uint)
    operatorName, _ := db.Statement.Context.Value("operator_name").(string)
    operatorIP, _ := db.Statement.Context.Value("operator_ip").(string)
    requestID, _ := db.Statement.Context.Value("request_id").(string)

    // 计算差异
    diff := calculateDiff(oldValues, newValues)

    log := DataOperationLog{
        TableName:     tableName,
        OperationType: operationType,
        RecordID:      recordID,
        OldValues:     toJSON(oldValues),
        NewValues:     toJSON(newValues),
        Diff:          toJSON(diff),
        OperatorID:    operatorID,
        OperatorName:  operatorName,
        OperatorIP:    operatorIP,
        RequestID:     requestID,
        Reason:        reason,
    }

    return db.Create(&log).Error
}
```

### 查询API

```go
// 查询数据操作日志
GET /api/data-logs?table=etf_configs&operation=UPDATE&start_date=2026-01-01&end_date=2026-12-31

// 查询单条记录的历史
GET /api/data-logs/record?table=etf_configs&record_id=1

// 数据回滚(管理员权限)
POST /api/data-logs/rollback
{
    "log_id": 123,
    "reason": "误操作恢复"
}
```

### 应用场景

| 场景 | 功能 |
|------|------|
| **数据追溯** | 查看某条记录的所有历史变更 |
| **误操作恢复** | 通过日志回滚到之前的状态 |
| **审计合规** | 记录谁在什么时间做了什么修改 |
| **数据分析** | 统计各表的操作频率和趋势 |

---

## 🎯 成功指标

### 技术指标
| 指标 | 当前值 | 目标值 | 状态 |
|------|--------|--------|------|
| **系统可用性** | 99.9% | 99.9% | ✅ |
| **API响应时间** | <200ms | <200ms | ✅ |
| **测试覆盖率** | ~60% | >80% | 🔄 |
| **代码质量评分** | 4/5 | 4.5/5 | 🔄 |

### 业务指标
- 🔄 用户活跃度提升 30%
- 🔄 功能使用率提升 50%
- 🔄 用户满意度 > 4.5/5
- 🔄 留存率提升 20%

---

## 📞 技术支持

### 问题排查
1. 检查 AGENTS.md 文档是否包含相关信息
2. 查看日志文件定位问题
3. 验证数据源连接状态
4. 检查数据库连接和配置

### 紧急处理
- **数据源故障**: 自动故障转移机制已启用
- **服务不可用**: 检查健康检查接口 `/health`
- **数据不一致**: 使用数据验证工具检查数据完整性

---

## 🤖 AI 助手规范

### 1. 交互流程

```
用户请求 → 理解意图 → 查阅上下文 → 执行任务 → 验证结果 → 响应用户
    ↓           ↓            ↓           ↓           ↓           ↓
  自然语言    提取关键信息   AGENTS.md   工具调用    测试验证    清晰反馈
```

### 2. 响应格式

| 场景 | 响应格式 |
|------|----------|
| **代码修改** | 说明 + 代码引用 + 验证结果 |
| **任务完成** | 完成状态 + 关键结果 + 后续建议 |
| **问题诊断** | 原因分析 + 解决方案 + 预防措施 |
| **进度汇报** | 当前状态 + 完成项 + 待办项 |

### 3. 错误处理

```go
// 错误处理优先级
1. 立即修复: 语法错误、类型错误、明显bug
2. 记录问题: 非关键问题，记录待后续处理
3. 忽略忽略: 与任务无关的警告（不阻断执行）
```

### 4. 用户数据保护

| 保护项 | 要求 |
|--------|------|
| **API Key** | 绝不记录或暴露，仅使用环境变量 |
| **密码** | 不记录，不在日志中输出 |
| **敏感配置** | 脱敏处理后记录 |
| **用户数据** | 仅在必要时访问，不存储副本 |

### 5. 任务执行标准

| 标准 | 要求 |
|------|------|
| **完整性** | 一次请求完成全部相关任务 |
| **准确性** | 验证后再报告成功 |
| **可追溯性** | 保留修改记录，关联issue |
| **文档同步** | 代码修改同步更新相关文档 |

---

## 🔒 强制上下文确认 (MANDATORY CONTEXT CONFIRMATION)

### 对话开始确认
```
[AI AGENT CHECKLIST - 每次对话开始时必须勾选]

□ 已完整阅读本文档开头的"强制上下文绑定"章节
□ 已查阅本文档中与当前任务相关的章节
□ 已了解项目架构设计原则
□ 已了解数据模型定义
□ 已了解编码规则和约束
□ 已了解安全规范要求
□ 已确认不会违反本文档中的任何规则

[确认声明]
我确认已阅读并理解本文档内容，将严格遵守文档中的架构设计、
编码规则、安全规范和数据约束。如有违反，愿意承担相应后果。
```

### 代码修改确认
```
[CODE MODIFICATION CHECKLIST - 每次修改代码前必须勾选]

□ 已查阅本文档相关章节，了解上下文
□ 修改符合架构设计原则
□ 修改符合数据模型定义
□ 修改符合编码规则
□ 修改符合安全规范
□ 修改后已同步更新本文档相关章节
□ 已通过本地测试验证

[确认声明]
我确认本次代码修改符合本文档所有规范，并已同步更新文档。
```

---

## 📚 相关文档链接

| 文档 | 路径 | 说明 |
|------|------|------|
| **项目 README (中文)** | `/README.md` | 项目介绍、快速开始、使用指南 |
| **项目 README (英文)** | `/README_EN.md` | English version of README |
| **API 文档** | `/docs/openapi.yaml` | OpenAPI 3.0 接口规范 |
| **环境变量模板** | `/.env.example` | 环境变量配置模板 |
| **后端配置** | `/backend/config.yaml` | 后端服务配置 |
| **审查总结** | `/REVIEW_SUMMARY.md` | 代码审查总结报告 |
| **安全改进** | `/SECURITY_IMPROVEMENTS.md` | 安全改进指南 |
| **代码审查报告** | `/CODE_REVIEW_REPORT.md` | 详细代码审查报告 |

---

*本文档最后更新: 2026-05-05 (v2.9 AI Agent 微服务版)*
*强制上下文绑定版本: v2.0*
