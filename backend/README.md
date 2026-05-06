# ETF-Insight Go Backend

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

开源专业的 ETF 量化分析平台后端服务，基于 Go/Gin 框架构建，提供高性能、可扩展的 RESTful API。

## 🎯 项目定位

ETF-Insight 后端坚持以下设计理念：

- **开源透明**: MIT 协议，算法实现完全可审计
- **专业准确**: 机构级量化算法，支持学术研究
- **高性能**: Go 语言 + Gin 框架，低延迟高并发
- **可扩展**: 插件化架构，支持自定义数据源和分析模块

## ✨ 核心特性

### 量化分析引擎
- **投资组合优化**: 马科维茨MPT、风险平价、Black-Litterman三种模型
- **投资组合情景分析**: 蒙特卡洛模拟、三种市场情景(乐观/中性/悲观)、VaR/CVaR风险指标
- **风险指标计算**: 波动率、最大回撤、夏普比率、索提诺比率、卡尔玛比率、Beta、Alpha
- **技术指标**: RSI、MACD、布林带、移动平均线
- **风险模型**: VaR/CVaR (历史法/参数法)、组合风险分解
- **多因子模型**: Fama-French 三因子/五因子模型、归因分析
- **回测引擎**: 事件驱动架构、完整订单系统、滑点/手续费模型

### QuantLib 量化引擎 (v2.8 新增)
- **期权定价**: 欧式/美式 Black-Scholes 定价，完整 Greeks (Delta/Gamma/Theta/Vega/Rho)
- **收益率曲线**: 构建和可视化，支持多货币 (USD/EUR/CNY/GBP/JPY)
- **债券定价**: 固定收益分析，久期、修正久期、凸性
- **VaR 计算**: QuantLib 引擎驱动的历史模拟法/参数法
- **参考数据**: 缓存 1 小时 TTL，自动清理过期缓存

### AI Agent 微服务 (v2.9 新增)
- **投资大师分析**: Warren Buffett (价值投资)、Benjamin Graham (防御型投资)
- **对冲基金视角**: Bridgewater Associates (宏观风险平价)
- **宏观经济分析**: 自上而下宏观经济分析
- **多 Agent 团队辩论**: 2-5 个 Agent 同时分析，综合观点
- **多 LLM 支持**: OpenAI/Ollama/DeepSeek，通过工厂函数切换
- **FastAPI 服务**: port 8091，19 个单元测试

### 数据源微服务 (v2.10 新增)
- **统一数据 API**: FastAPI 服务 (port 8092)，封装 6 大数据源
- **FRED**: 美联储经济数据，支持时间序列和搜索
- **World Bank**: 世界银行经济指标，国家经济快照
- **IMF**: 国际货币基金组织数据，经济指标和贸易数据
- **Yahoo Finance**: 股票实时报价、历史数据、公司信息
- **AkShare**: A股实时行情、宏观经济、债券、加密货币 (11 个模块)
- **CoinGecko**: 加密货币价格、市场数据、趋势
- **AGPL 合规**: 直接使用 FinceptTerminal 代码，NOTICE 文件声明

### 分析微服务 (v2.11 新增)
- **TDD 开发**: RED → GREEN → REFACTOR 流程
- **Portfolio Management**: 16 个分析模块从 FinceptTerminal 提取
- **Portfolio Optimization**: 组合优化 (最大夏普/最小波动率/等权重)
- **Risk Management**: VaR/CVaR 计算、风险预算、情景分析
- **Portfolio Analytics**: CAPM 分析、有效前沿、组合指标
- **Portfolio Planning**: 资产配置、风险容忍度分析
- **FastAPI 服务**: port 8093，4 个路由模块
- **AGPL 合规**: 直接使用 FinceptTerminal 代码，NOTICE 文件声明

### 数据服务 (v2.6 重构)
- **统一资产模型**: Asset 基表支持股票/ETF/指数等多种资产类型
- **ETF持仓穿透**: 底层持仓明细查询、权重分析
- **重叠度计算**: 两只ETF持仓重叠度分析(最小权重法)
- **组合穿透分析**: 投资组合底层资产行业/地理分布
- **集中度指标**: Top10/Top20权重、Herfindahl指数、有效持仓数
- **智能缓存**: 重叠度计算结果缓存(7天TTL)、自动失效
- **事件驱动**: 持仓更新自动触发缓存失效

### 数据源服务
- **多数据源支持**: Finage 主数据源 + 备用数据源故障转移
- **汇率服务**: 多数据源汇率，自动故障转移
- **A股ETF支持**: AKShare/TuShare接入、中证红利、红利低波等主流红利ETF
- **跨资产类别**: 股票/债券/商品/REIT/货币/多资产/另类全覆盖
- **数据完整性**: 所有字段入库校验，操作日志记录

### 基础设施
- **审计日志**: 异步写入，敏感信息自动脱敏
- **数据验证**: 通用验证中间件，防止注入攻击
- **速率限制**: IP级别限流，防止滥用

## 📁 项目结构

```
backend/
├── config/                 # 配置管理
│   ├── config.go          # 配置结构定义与加载
│   └── config_test.go     # 配置测试
├── models/                # 数据模型 (GORM)
│   ├── models.go          # ETFConfig, ETFData, OperationLog
│   ├── db.go              # 数据库初始化与迁移
│   ├── asset.go           # 统一资产模型 (v2.6)
│   ├── etf_holding.go     # ETF持仓穿透模型 (v2.6)
│   ├── asset_metadata.go  # 资产元数据模型 (v2.6)
│   ├── cache.go           # 缓存模型 (v2.6)
│   ├── exchange_rate.go   # 汇率相关模型
│   ├── a_share_dividend_etf.go  # A股ETF模型
│   ├── audit_log.go       # 审计日志模型
│   ├── pagination.go      # 分页模型
│   ├── quantlib.go           # QuantLib 请求/响应模型 (v2.8)
│   └── quantlib_validator.go # QuantLib 输入验证 (v2.8)
├── handlers/              # HTTP 处理器
│   ├── etf_handler.go     # ETF 行情/历史/指标接口
│   ├── etf_holding_handler.go  # ETF持仓穿透接口 (v2.6)
│   ├── portfolio_handler.go     # 投资组合分析接口
│   ├── portfolio_penetration.go # 组合穿透分析接口 (v2.6)
│   ├── a_share_portfolio_handler.go  # A股ETF组合接口
│   ├── exchange_rate.go   # 汇率管理接口
│   ├── health_handler.go  # 健康检查
│   ├── middleware.go      # 中间件
│   └── quantlib_handler.go   # QuantLib API 接口 (v2.8)
│   └── agent_handler.go      # Agent 服务 Handler (v2.9)
│   └── data_handler.go       # 数据源服务 Handler (v2.10)
├── middleware/            # 中间件
│   ├── audit.go           # 审计日志中间件
│   ├── validation.go      # 数据验证中间件
│   ├── security.go        # 安全头 + 速率限制
│   └── security_test.go
├── services/              # 业务逻辑层
│   ├── datasource/        # 数据源微服务层
│   ├── exchange_rate/     # 汇率服务
│   ├── sync/              # 同步服务
│   ├── backtest/          # 回测引擎(事件驱动)
│   ├── optimization/      # 组合优化(MPT/风险平价/BL)
│   ├── factor/            # 因子分析(Fama-French)
│   ├── ashare/            # A股数据源(AKShare/TuShare)
│   ├── etf/               # ETF服务(跨资产类别)
│   │   ├── holdings_service.go  # 持仓服务 (v2.6)
│   │   └── cache_service.go     # 缓存服务 (v2.6)
│   ├── portfolio/         # 组合服务
│   │   └── penetration.go       # 穿透分析服务 (v2.6)
│   ├── event/             # 事件服务 (v2.6)
│   │   └── trigger.go           # 事件总线
│   ├── etf_analysis.go    # ETF分析服务
│   ├── portfolio_optimizer.go  # 组合优化服务
│   ├── portfolio_analytics.go  # 组合分析服务(金融公式计算)
│   ├── scenario_analysis.go    # 情景分析服务(蒙特卡洛模拟)
│   ├── technical_indicators.go # 技术指标服务(RSI/MACD/布林带)
│   ├── risk_models.go          # 风险模型服务(VaR/CVaR)
│   └── quantlib/              # QuantLib 云 API 客户端 (v2.8)
│       ├── quantlib_client.go      # HTTP 客户端 (10 方法 + 缓存)
│       └── quantlib_client_test.go # 单元测试 (httptest mock)
│   └── agent/                # AI Agent 微服务 (v2.9)
│       ├── core/
│       │   ├── base_agent.py        # Agent 抽象基类
│       │   ├── llm_provider.py      # 多模型抽象层 (OpenAI/Ollama/DeepSeek)
│       │   ├── tool_registry.py     # 工具注册机制
│       │   └── agent_manager.py     # Agent 管理器
│       ├── agents/
│       │   ├── buffett.py           # Warren Buffett Agent
│       │   ├── graham.py            # Benjamin Graham Agent
│       │   ├── bridgewater.py       # Bridgewater Agent
│       │   ├── macro.py             # 宏观经济分析 Agent
│       │   └── registry.py          # Agent 自动注册
│       ├── models/schemas.py        # Pydantic v2 请求/响应模型
│       ├── agent_server.py          # FastAPI 入口 (port 8091)
│       ├── Dockerfile               # 容器化部署
│       └── tests/                   # 19 个单元测试
│   └── agent/                # Agent 服务 Go 客户端
│       └── agent_client.go       # HTTP 客户端 (4 方法)
│   └── data/                 # 数据源微服务 (v2.10)
│       ├── sources/              # 16 个数据源脚本 (FinceptTerminal)
│       ├── routers/              # 6 个 FastAPI 路由模块
│       ├── tests/                # 单元测试
│       ├── data_server.py        # FastAPI 入口 (port 8092)
│       ├── data_client.go        # Go HTTP 客户端
│       ├── Dockerfile            # 容器化部署
│       └── NOTICE                # AGPL 合规声明
├── tasks/                 # 定时任务
│   ├── scheduler.go       # 主调度器
│   ├── exchange_rate_task.go
│   └── ashare_price_scheduler.go
├── utils/                 # 工具包
│   ├── logger.go          # 日志工具
│   └── logger_test.go
├── cmd/                   # 命令行工具
│   ├── syncetf/           # ETF数据同步
│   ├── updateashare/      # A股ETF更新
│   └── syncrates/         # 汇率同步
├── main.go                # 入口文件
└── go.mod                 # Go 模块配置
```

## 🚀 快速开始

### 环境要求
- Go >= 1.21
- SQLite (开发) 或 PostgreSQL (生产)
- Finage API Key (必需)

### 1. 安装依赖

```bash
cd backend
go mod tidy
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，配置以下关键变量：
# - FINAGE_API_KEY (必需)
# - 数据库配置
# - 汇率数据源配置
```

### 3. 启动服务

```bash
# 开发模式
go run main.go

# 生产模式
go build -o etf-insight
./etf-insight
```

服务将在 http://localhost:8080 启动

AI Agent 服务: http://localhost:8091 (Python FastAPI)

## 📚 API 文档

### ETF 分析接口
- `GET /api/etf/list` - ETF列表（支持分页）
- `GET /api/etf/comparison` - ETF对比分析
- `GET /api/etf/:symbol/realtime` - ETF实时数据
- `GET /api/etf/:symbol/history` - 历史数据
- `GET /api/etf/:symbol/metrics` - 风险指标
- `GET /api/etf/:symbol/forecast` - ETF预测
- `GET /api/etf/:symbol/risk` - ETF风险指标(VaR/CVaR)
- `POST /api/etf/update-realtime` - 手动刷新实时数据

### ETF持仓穿透接口 (v2.6 新增)
- `GET /api/etf/:symbol/holdings` - 获取ETF底层持仓明细
- `GET /api/etf/overlap?sym1=&sym2=` - 计算两只ETF持仓重叠度
- `GET /api/etf/:symbol/top-holdings` - 获取前N大持仓
- `GET /api/etf/:symbol/sector-allocation` - 获取行业配置
- `POST /api/etf/holdings/comparison` - 多ETF持仓对比
- `POST /api/etf/:symbol/holdings` - 保存持仓数据

### 投资组合穿透接口 (v2.6 新增)
- `POST /api/portfolio/penetration` - 组合穿透分析(行业/地理分布)
- `POST /api/portfolio/compare` - 组合对比分析
- `POST /api/portfolio/sector-exposure` - 行业暴露查询

### 缓存管理接口 (v2.6 新增)
- `GET /api/cache/overlap/stats` - 缓存统计信息
- `POST /api/cache/overlap/invalidate` - 手动失效缓存
- `POST /api/cache/overlap/clean` - 清理过期缓存

### 投资组合接口
- `POST /api/etf/portfolio` - 组合分析
- `POST /api/portfolio/scenarios` - 投资组合情景分析(蒙特卡洛模拟)
- `GET /api/portfolio/default-templates` - 获取默认投资组合模板
- `POST /api/portfolio/risk` - 组合风险分析
- `POST /api/portfolio/optimize` - 组合优化(旧接口)
- `POST /api/portfolio/efficient-frontier` - 有效前沿(旧接口)

### 投资组合配置接口
- `GET /api/portfolio-configs/` - 组合配置列表
- `POST /api/portfolio-configs/` - 创建组合配置
- `GET /api/portfolio-configs/:id` - 获取组合配置
- `PUT /api/portfolio-configs/:id` - 更新组合配置
- `DELETE /api/portfolio-configs/:id` - 删除组合配置
- `POST /api/portfolio-configs/:id/toggle-status` - 切换状态
- `POST /api/portfolio-configs/:id/analyze` - 分析组合配置

### 组合优化接口
- `POST /api/optimization/mpt` - MPT均值-方差优化
- `POST /api/optimization/efficient-frontier` - 有效前沿计算
- `POST /api/optimization/covariance` - 协方差矩阵计算
- `POST /api/optimization/etf-statistics` - ETF统计信息
- `POST /api/optimization/risk-parity` - 风险平价优化
- `POST /api/optimization/black-litterman` - Black-Litterman优化
- `POST /api/optimization/market-implied-returns` - 市场隐含收益

### 因子分析接口
- `POST /api/factor/analyze` - 因子暴露度分析
- `POST /api/factor/portfolio` - 组合因子分析
- `POST /api/factor/multi-asset` - 多资产因子分析
- `GET /api/factor/statistics` - 因子统计信息
- `POST /api/factor/risk-decomposition` - 风险分解
- `POST /api/factor/compare` - 因子归因对比

### QuantLib 量化分析接口 (v2.8 新增)
- `POST /api/quantlib/options/european` - 欧式期权定价
- `POST /api/quantlib/options/american` - 美式期权定价
- `POST /api/quantlib/options/greeks` - Greeks 计算
- `POST /api/quantlib/yield-curve/build` - 收益率曲线构建
- `POST /api/quantlib/bonds/price` - 债券定价
- `POST /api/quantlib/risk/var` - VaR 计算
- `GET /api/quantlib/reference/:type` - 参考数据 (currencies/frequencies/calendars/daycount)

### AI Agent 微服务接口 (v2.9 新增)
- `GET /api/agents/health` - 健康检查 (Agent数量/LLM提供商)
- `GET /api/agents/discover` - Agent 列表发现
- `POST /api/agents/run` - 单 Agent 执行
- `POST /api/agents/stream` - SSE 流式响应
- `POST /api/agents/team` - 多 Agent 团队辩论

### 回测引擎接口
- `POST /api/backtest/run` - 运行回测
- `POST /api/backtest/event-driven` - 事件驱动回测
- `GET /api/backtest/strategies` - 策略列表
- `POST /api/backtest/factors` - 因子分析

### ETF配置管理接口
- `GET /api/etf-configs/` - ETF配置列表
- `POST /api/etf-configs/` - 创建ETF配置
- `GET /api/etf-configs/:id` - 获取ETF配置
- `PUT /api/etf-configs/:id` - 更新ETF配置
- `DELETE /api/etf-configs/:id` - 删除ETF配置
- `POST /api/etf-configs/:id/toggle-status` - 切换状态
- `POST /api/etf-configs/:id/auto-update` - 切换自动更新

### A股ETF接口
- `GET /api/a-share/etfs` - A股ETF列表
- `GET /api/a-share/portfolio/default` - 默认组合
- `POST /api/a-share/portfolio/analyze` - 分析组合
- `POST /api/a-share/portfolio/holding/:symbol` - 更新持仓
- `GET /api/a-share/dividend/:frequency` - 按频率计算分红
- `GET /api/a-share/prices` - ETF价格列表
- `GET /api/a-share/prices/:symbol` - 单只ETF价格
- `POST /api/a-share/prices/refresh` - 刷新价格
- `POST /api/a-share/enable-akshare` - 启用AKShare数据源
- `POST /api/a-share/sync-etf-list` - 同步ETF列表
- `POST /api/a-share/sync-prices` - 同步价格
- `POST /api/a-share/refresh-all` - 刷新所有数据
- `GET /api/a-share/price/:symbol` - 单只ETF价格(数据服务)
- `GET /api/a-share/all-prices` - 所有ETF价格(数据服务)
- `POST /api/a-share/historical/:symbol` - 历史数据
- `GET /api/a-share/search` - 搜索ETF
- `GET /api/a-share/by-frequency/:frequency` - 按分红频率筛选
- `GET /api/a-share/dividend-yield/:symbol` - 计算股息率
- `GET /api/a-share/data-source-status` - 数据源状态

### 跨资产类别ETF接口
- `POST /api/universal-etf/initialize` - 初始化ETF数据
- `GET /api/universal-etf` - 获取所有ETF
- `GET /api/universal-etf/:symbol` - 获取单个ETF
- `GET /api/universal-etf/asset-class/:asset_class` - 按资产类别筛选
- `GET /api/universal-etf/region/:region` - 按地区筛选
- `GET /api/universal-etf/type/:etf_type` - 按类型筛选
- `GET /api/universal-etf/search` - 搜索ETF
- `POST /api/universal-etf/filter` - 多条件筛选
- `GET /api/universal-etf/distribution/asset-class` - 资产类别分布
- `GET /api/universal-etf/distribution/region` - 地区分布
- `POST /api/universal-etf/compare` - ETF对比
- `GET /api/universal-etf/portfolio-allocation` - 组合配置建议
- `GET /api/universal-etf/categories` - 获取分类列表
- `GET /api/universal-etf/top-performers` - 获取表现最佳ETF

### 汇率接口
- `GET /api/exchange-rates` - 汇率列表
- `GET /api/exchange-rates/:from/:to` - 货币对汇率
- `POST /api/exchange-rates/convert` - 货币转换
- `POST /api/exchange-rates/sync` - 手动同步
- `GET /api/exchange-rates/summary` - 汇率摘要
- `GET /api/exchange-rates/currencies` - 支持的货币列表
- `GET /api/exchange-rates/datasource-status` - 数据源状态
- `GET /api/currency-pairs` - 货币对列表

### 操作日志接口
- `GET /api/logs/` - 日志列表(支持筛选)
- `GET /api/logs/types` - 日志类型统计
- `GET /api/logs/action-types` - 操作类型列表
- `GET /api/logs/users` - 用户列表
- `POST /api/logs/export` - 导出日志
- `GET /api/logs/:type/:id` - 日志详情

### 健康检查
- `GET /health` - 服务健康状态
- `GET /ready` - 服务就绪状态
- `GET /live` - 服务存活状态

## 🔧 开发指南

### 代码规范
- 遵循 Go 官方代码规范，使用 `gofmt` 格式化
- 重要函数必须包含完整注释（输入、输出、边界情况）
- 新功能必须包含单元测试，覆盖率 >80%
- 算法实现必须包含公式说明和参考文献

### 中间件使用

```go
// 审计日志
router.Use(middleware.AuditLogger())

// 数据验证
router.POST("/api/etf", middleware.ValidateInput(rules), handler)

// 速率限制
router.Use(middleware.RateLimiterHandler(100, time.Minute))
```

### 添加新数据源

```go
// 1. 实现 Provider 接口
type MyDataProvider struct{}

func (p *MyDataProvider) GetETFData(symbol string) (*ETFData, error) {
    // 实现数据获取逻辑
}

// 2. 注册到 ProviderFactory
func init() {
    datasource.RegisterProvider("myprovider", &MyDataProvider{})
}
```

## 🧪 测试

### 测试覆盖率 (2026-04-22)

| 模块 | 覆盖率 | 关键功能 |
|------|--------|----------|
| services/portfolio | **84.9%** | 组合穿透分析、集中度指标 |
| services/event | **73.2%** | 事件总线、缓存失效 |
| services/factor | **80.8%** | Fama-French 三因子/五因子模型 |
| services/technical | 100% | RSI、MACD、布林带 |
| services/risk | 100% | VaR、CVaR、风险指标 |
| services/quantlib | 9 tests | QuantLib 云 API 客户端、缓存、错误处理 |
| services/agent | 19 tests | Agent 框架、LLM Provider、Agent Manager、API 端点 |
| middleware | 68.8% | 审计、安全 |
| utils | 81.2% | 工具函数 |

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包测试
go test -v ./services/...

# 运行组合穿透分析测试
go test -v ./services/portfolio

# 运行事件服务测试
go test -v ./services/event

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 数据源策略

### ETF 数据源
- **主数据源**: Finage API (唯一真实数据源)
- **数据质量**: 实时数据、完整OHLCV字段、入库校验
- **历史数据**: 3年历史数据同步(约388天)
- **同步频率**: 定时任务自动更新 + 手动刷新接口

### 历史数据同步
```bash
# 同步3年历史数据到数据库
cd backend
FINAGE_API_KEY=your_api_key go run cmd/sync_etf_history/main.go

# 支持的ETF: SCHD, JEPQ, JEPI, SPYD, VOO, QQQ
```

### 实时数据更新
```bash
# 手动刷新实时数据
curl -X POST http://localhost:8080/api/etf/update-realtime

# 返回示例:
# {"success": true, "count": 6, "source": "finage"}
```

### 汇率数据源
- **主数据源**: Open Exchange Rates
- **备用数据源**: CurrencyAPI、Frankfurter
- **故障转移**: 自动切换、健康检查

## 🤝 贡献指南

欢迎贡献代码！请遵循以下流程：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 贡献类型
- **算法贡献**: 实现新的量化指标或策略
- **数据源**: 添加新的数据提供商支持
- **Bug修复**: 修复问题和改进性能
- **文档**: 完善文档和注释

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE) 文件

## 🔗 相关链接

- [项目主页](https://github.com/coder100001/ETF-Insight)
- [前端文档](../frontend/README.md)
- [架构文档](../AGENTS.md)
