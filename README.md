# ETF-Insight (v2.0 Finage-Only)

一个专业的 ETF 分析与对比平台，对标 Trackinsight、ETF Insider 等国际知名 ETF 分析工具。基于 Go + React 技术栈，提供深度的 ETF 数据洞察、多维度对比分析、持仓解构、风险评估和投资组合优化等一站式解决方案。

**v2.0 架构更新**: 完全依赖 Finage 真实数据，删除所有硬编码mock，所有字段必须入库。

## 🎯 产品定位

ETF-Insight 致力于成为专业投资者和机构用户的 ETF 分析利器：

- **ETF 对比分析** - 多维度并排对比，发现最优投资标的
- **持仓深度解构** - 穿透底层资产，了解真实风险敞口
- **风险指标评估** - 波动率、夏普比率、最大回撤、Beta 等专业指标
- **投资组合优化** - 基于现代投资组合理论，构建最优资产配置

## ✨ 核心特性

### 📊 ETF 对比分析（ETF Comparison）
- **并排对比** - 最多支持 5 只 ETF 同时对比
- **多维度指标** - 费率、AUM、股息率、业绩表现、风险指标
- **持仓重叠分析** - 识别 ETF 间的持仓重合度，避免过度集中
- **业绩回测对比** - 不同时间周期的收益表现对比

### 🔍 持仓深度解构（Holdings Analysis）
- **前十大持仓** - 穿透底层资产，了解核心持仓
- **行业分布** - sector 权重分布及变化趋势
- **地区分布** - 国家/地区配置比例
- **市值分布** - 大/中/小盘股配置比例
- **风格分析** - 价值/成长风格暴露度

### 💼 A股红利ETF投资组合
- **A股ETF管理** - 支持中证红利、红利低波等主流红利ETF
- **投资占比分布** - 饼状图可视化展示投资组合配置
- **分红数据追踪** - 股息率、分红频率等关键指标

### 💱 汇率数据管理
- **实时汇率** - USD/CNY、USD/HKD 等主要货币对
- **自动同步** - 定时任务自动更新汇率数据（每5分钟）
- **货币转换** - 支持多种货币间的换算功能
- **同步日志** - 完整的汇率同步批次记录与明细追踪

### ⚙️ ETF 配置管理
- **CRUD 操作** - 增删改查 ETF 配置信息
- **状态管理** - 启用/禁用 ETF 数据自动更新
- **数据源配置** - **Finage 唯一真实数据源** (v2.0 架构)

### 📈 投资组合配置
- **组合构建** - 自定义投资组合及权重分配
- **收益分析** - 基于历史数据的组合收益模拟
- **预设组合** - 内置多种投资策略组合模板

## 🛠️ 技术栈

### 后端 (Go)
| 技术 | 版本 | 用途 |
|------|------|------|
| Go | >= 1.21 | 核心语言 |
| Gin | v1.12.0 | Web 框架 |
| GORM | v1.30.0 | ORM 框架 (SQLite/PostgreSQL) |
| go-cache | v2.1.0 | 内存缓存 |
| cron/v3 | v3.0.1 | 定时任务调度 |

### 前端 (React)
| 技术 | 版本 | 用途 |
|------|------|------|
| React | ^19.2.4 | UI 框架 |
| TypeScript | ^5.x | 类型安全 |
| Vite | latest | 构建工具 |
| Ant Design | ^6.3.4 | UI 组件库 |
| ECharts | ^6.0.0 | 数据可视化 |
| Recharts | ^3.8.1 | 图表组件 |
| React Router | ^7.13.2 | 路由管理 |

### 数据存储
- **SQLite** - 默认本地数据库（开发环境）
- **PostgreSQL** - 生产数据库支持

## 🚀 快速开始

### 方式一：一键启动（推荐）

```bash
# 克隆项目
git clone https://github.com/coder100001/ETF-Insight.git
cd ETF-Insight

# macOS / Linux
chmod +x start.sh
./start.sh

# Windows
start.bat
```

启动脚本会自动完成以下操作：
1. ✅ 检查运行环境（Go、Node.js）
2. ✅ 安装后端依赖（go mod download）
3. ✅ 编译后端项目
4. ✅ 安装前端依赖（npm install）
5. ✅ 启动后端服务（端口 8080）
6. ✅ 启动前端服务（端口 5173）

### 方式二：Docker 部署

```bash
git clone https://github.com/coder100001/ETF-Insight.git
cd ETF-Insight
docker-compose up -d
```

访问 http://localhost:8080

### 方式三：手动启动

```bash
# 后端
cd backend
go mod download
go build -o etf-insight .
./etf-insight

# 新终端 - 前端
cd frontend
npm install
npm run dev
```

## 📋 环境要求

| 工具 | 最低版本 | 推荐版本 |
|------|----------|----------|
| Go | 1.21+ | 1.25+ |
| Node.js | 18+ | 20+ |
| npm | 9+ | 10+ |

## 🔧 配置说明

### 环境变量

复制 `.env.example` 并配置：

```bash
# 代理配置 (如需翻墙访问 API)
HTTP_PROXY=http://127.0.0.1:7897
HTTPS_PROXY=http://127.0.0.1:7897

# Finage API Key (唯一数据源) - 必须配置，否则系统无法工作
FINAGE_API_KEY=your_finage_api_key_here

# 注意：Finnhub API Key 已废弃，仅作为历史代码保留
FINNHUB_API_KEY=your_finnhub_api_key_here
```

> **⚠️ 安全提醒**: API Key 不得硬编码到代码中，统一通过环境变量配置。
> **⚠️ 安全提醒**: API Key 不得硬编码到代码中，统一通过环境变量配置。
> **v2.0 架构更新**: 完全依赖 Finage 真实数据，所有硬编码mock数据已删除。

### 后端配置文件

位于 `backend/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  dsn: "etf_insight.db"     # SQLite / PostgreSQL DSN

etf:
  cache:
    type: "memory"
    ttl: 3600

schedule:
  update_interval: "1h"

log:
  level: "info"
```

## 🏛️ 缓存架构 (OOP 设计)

### 设计原则

缓存系统采用面向对象设计，严格遵循 SOLID 原则：

- **单一职责原则 (SRP)** - 每个组件只负责一个功能
- **开闭原则 (OCP)** - 通过接口扩展，无需修改现有代码
- **里氏替换原则 (LSP)** - 所有缓存实现可互相替换
- **接口隔离原则 (ISP)** - 接口只包含必要的方法
- **依赖倒置原则 (DIP)** - 依赖于抽象接口，而非具体实现

### 缓存策略

| 策略 | 描述 | 适用场景 |
|------|------|---------|
| **Memory** | 纯内存缓存 | 单机部署、开发测试 |
| **Redis** | 纯 Redis 缓存 | 分布式部署、需要持久化 |
| **Hybrid** | Redis + Memory 混合 | 生产环境、高性能要求 |

### 核心组件

```
CacheService (业务层)
    ↓
CacheProvider Interface (抽象接口)
    ↓
┌─────────────┬─────────────┬─────────────┐
│ MemoryCache │ RedisCache  │ HybridCache │
└─────────────┴─────────────┴─────────────┘
    ↓
CacheFactory (工厂模式)
```

### 配置示例

```yaml
redis:
  enabled: true          # 启用 Redis
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  timeout: 5s

etf:
  cache:
    realtime_ttl: 5m     # 实时数据缓存时间
    historical_ttl: 24h  # 历史数据缓存时间
    metrics_ttl: 1h      # 指标数据缓存时间
    comparison_ttl: 30m  # 对比数据缓存时间
```

### 使用示例

```go
// 创建缓存服务
cacheService := services.NewCacheService(cacheCfg, redisCfg)

// 使用缓存
cacheService.Set("etf:SCHD", data, 5*time.Minute)
value, found := cacheService.Get("etf:SCHD")

// 获取缓存统计
stats := cacheService.GetCacheStats()
// Output: {"provider_type": "hybrid"}
```

## 📁 项目结构

```
ETF-Insight/
├── start.sh                    # 一键启动脚本 (macOS/Linux)
├── start.bat                   # 一键启动脚本 (Windows)
├── .env.example                # 环境变量模板
├── backend/
│   ├── main.go                 # 后端入口 + 路由注册
│   ├── config.yaml             # 配置文件
│   ├── config/                 # 配置管理
│   │   ├── config.go           # 配置结构定义与加载
│   │   └── config_test.go      # 配置测试
│   ├── models/                 # 数据模型
│   │   ├── models.go           # ETFConfig, ETFData, OperationLog, PortfolioConfig
│   │   ├── db.go               # 数据库初始化与迁移
│   │   ├── exchange_rate.go    # ExchangeRate, ExchangeRateSyncLog, CurrencyPair 等
│   │   └── a_share_dividend_etf.go  # AShareDividendETF, AShareETFPortfolio 等
│   ├── handlers/               # API 处理器
│   │   ├── etf_handler.go      # ETF 行情/历史/指标/预测接口
│   │   ├── etf_config_handler.go    # ETF 配置 CRUD 接口
│   │   ├── portfolio_handler.go     # 投资组合分析/配置接口
│   │   ├── a_share_portfolio_handler.go  # A股红利ETF组合接口
│   │   ├── exchange_rate.go    # 汇率管理接口
│   │   ├── health_handler.go   # 健康检查 (health/ready/live)
│   │   └── middleware.go       # 日志与 CORS 中间件
│   ├── services/               # 业务逻辑层
│   │   ├── datasource/         # 数据源微服务层 (策略模式)
│   │   │   ├── provider.go     # 数据源接口定义 + ProviderFactory
│   │   │   ├── errors.go       # 标准错误定义
│   │   │   ├── finage_provider.go   # Finage API (聚合API + last API)
│   │   │   ├── finnhub_provider.go  # Finnhub API 实现
│   │   │   └── fallback_provider.go # 后备数据源
│   │   ├── sync/               # 数据同步服务
│   │   │   ├── service.go      # 同步业务逻辑 + 入库校验 + 操作日志
│   │   │   └── config.go       # ETF配置数据 + 预设组合
│   │   ├── etf_analysis.go     # ETF分析服务 (指标/组合/预测/对比)
│   │   ├── yahoo_finance.go    # Yahoo Finance 客户端
│   │   ├── cache.go            # 缓存服务 + RealtimeData 模型
│   │   ├── exchange_rate.go    # 汇率服务
│   │   └── finnhub.go          # Finnhub 独立客户端
│   ├── middleware/             # 中间件
│   │   ├── security.go         # 安全头 + 速率限制 (100/min)
│   │   └── security_test.go
│   ├── tasks/                  # 定时任务
│   │   ├── scheduler.go        # 主调度器 (ETF更新/汇率更新/小时检查)
│   │   └── exchange_rate_task.go   # 汇率同步任务 (5min/10:30daily)
│   ├── utils/                  # 工具包
│   │   ├── logger.go           # 日志工具
│   │   └── logger_test.go
│   ├── cmd/                    # 命令行工具
│   │   ├── syncetf/            # ETF数据同步工具
│   │   ├── update_etf_prices/  # ETF价格批量更新 (Finage聚合API)
│   │   ├── generate_history/   # 生成模拟历史数据
│   │   ├── initetf/            # ETF初始数据导入
│   │   ├── syncrates/          # 汇率数据同步
│   │   ├── updateashare/       # A股红利ETF数据更新
│   │   ├── test_factory/       # 数据源工厂测试
│   │   └── test_finage/        # Finage API 测试
│   └── infrastructure/         # 基础设施 (预留目录)
├── frontend/
│   ├── src/
│   │   ├── pages/              # 页面组件
│   │   │   ├── Dashboard.tsx          # 仪表盘
│   │   │   ├── ETFDashboard.tsx       # ETF 市场总览
│   │   │   ├── ETFComparison.tsx      # ETF 对比分析
│   │   │   ├── ETFComparisonReport.tsx # ETF 对比报告
│   │   │   ├── ETFDetail.tsx          # ETF 详情页
│   │   │   ├── ETFConfig.tsx          # ETF 配置管理
│   │   │   ├── PortfolioAnalysis.tsx   # 投资组合分析
│   │   │   ├── PortfolioConfig.tsx     # 组合配置管理
│   │   │   ├── ASharePortfolio.tsx     # A股红利ETF组合
│   │   │   ├── ExchangeRate.tsx        # 汇率管理
│   │   │   ├── InvestmentStrategy.tsx  # 投资策略
│   │   │   └── OperationLogs.tsx       # 操作日志
│   │   ├── components/         # 公共组件
│   │   │   ├── Layout.tsx             # 布局
│   │   │   ├── PriceChart.tsx         # 价格图表
│   │   │   ├── ComparisonRadarChart.tsx # 对比雷达图
│   │   │   ├── ETFFilter.tsx          # ETF 筛选
│   │   │   ├── HoldingPieChart.tsx    # 持仓饼图
│   │   │   ├── SectorBarChart.tsx     # 行业柱状图
│   │   │   ├── StatCard.tsx           # 统计卡片
│   │   │   └── StockCard.tsx          # 股票卡片
│   │   ├── services/api.ts     # API 服务 (含请求合并+重试)
│   │   ├── types/index.ts      # TypeScript 类型定义
│   │   └── styles/theme.ts     # 主题配置
│   └── package.json
├── docs/
│   └── openapi.yaml            # OpenAPI 3.0 接口文档
├── scripts/
│   ├── install-hooks.sh        # Git hooks 安装
│   └── startup.sh              # 生产启动脚本
└── docker-compose.yml
```

## 🌐 API 接口

### 健康检查
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 基础健康检查 |
| GET | `/ready` | 就绪检查 (含数据库/服务状态) |
| GET | `/live` | 存活检查 (含运行时间) |

### ETF 行情相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/etf/list` | 获取 ETF 列表 (含行情+指标, 5min缓存) |
| GET | `/api/etf/:symbol/realtime` | 获取单只 ETF 实时数据 |
| GET | `/api/etf/:symbol/history` | 获取历史数据 (支持 period: 1m/3m/6m/1y/3y/5y) |
| GET | `/api/etf/:symbol/metrics` | 获取风险指标 (波动率/夏普比率/最大回撤) |
| GET | `/api/etf/:symbol/forecast` | 获取10年收益预测 |
| GET | `/api/etf/comparison` | 获取 ETF 对比数据 |
| POST | `/api/etf/update-realtime` | 更新所有 ETF 实时数据 (Yahoo Finance) |
| POST | `/api/etf/portfolio` | 分析投资组合 |

### ETF 配置管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/etf-configs/` | 获取 ETF 配置列表 |
| POST | `/api/etf-configs/` | 创建 ETF 配置 |
| GET | `/api/etf-configs/:id` | 获取单个 ETF 配置 |
| PUT | `/api/etf-configs/:id` | 更新 ETF 配置 |
| DELETE | `/api/etf-configs/:id` | 删除 ETF 配置 |
| POST | `/api/etf-configs/:id/toggle-status` | 切换启用/禁用状态 |
| POST | `/api/etf-configs/:id/auto-update` | 切换自动更新设置 |

### 投资组合配置
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/portfolio-configs/` | 获取组合配置列表 |
| POST | `/api/portfolio-configs/` | 创建组合配置 |
| GET | `/api/portfolio-configs/:id` | 获取单个组合配置 |
| PUT | `/api/portfolio-configs/:id` | 更新组合配置 |
| DELETE | `/api/portfolio-configs/:id` | 删除组合配置 |
| POST | `/api/portfolio-configs/:id/toggle-status` | 切换状态 |
| POST | `/api/portfolio-configs/:id/analyze` | 分析组合收益 |

### A股红利ETF组合
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/a-share/etfs` | 获取A股红利ETF列表 |
| GET | `/api/a-share/portfolio/default` | 获取默认组合配置 |
| POST | `/api/a-share/portfolio/analyze` | 分析自定义组合 |
| POST | `/api/a-share/portfolio/holding/:symbol` | 更新持仓金额 |
| GET | `/api/a-share/dividend/:frequency` | 按频率计算分红 (monthly/quarterly/yearly) |

### 汇率管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/exchange-rates` | 获取汇率列表 (支持 from/to/source 筛选) |
| GET | `/api/exchange-rates/:from/:to` | 获取指定货币对汇率 |
| POST | `/api/exchange-rates/convert` | 货币转换 |
| POST | `/api/exchange-rates/sync` | 触发汇率同步 (异步) |
| GET | `/api/exchange-rates/summary` | 获取汇率摘要 |
| GET | `/api/exchange-rates/currencies` | 获取支持货币列表 |
| GET | `/api/currency-pairs` | 获取货币对配置 |

## 📖 使用指南

### 启动项目

```bash
# 一键启动（推荐）
./start.sh

# 或手动启动
cd backend && ./etf-insight &
cd frontend && npm run dev
```

### 访问地址

- **前端界面**: http://localhost:5173
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **就绪检查**: http://localhost:8080/ready
- **存活检查**: http://localhost:8080/live

### ETF 数据更新 (v2.0 Finage-Only)

```bash
# 使用 Finage 聚合API 逐个更新 ETF 价格 (完整 OHLCV)
cd backend && go run ./cmd/update_etf_prices/

# 通过 API 触发实时数据更新 (Finage 聚合API)
curl -X POST http://localhost:8080/api/etf/update-realtime

# 数据同步命令行工具
cd backend && go run ./cmd/syncetf/

# 带代理运行
HTTP_PROXY=http://127.0.0.1:7897 HTTPS_PROXY=http://127.0.0.1:7897 go run ./cmd/update_etf_prices/
```

### 其他命令行工具

```bash
# 生成模拟历史数据 (90天)
cd backend && go run ./cmd/generate_history/

# ETF初始数据导入
cd backend && go run ./cmd/initetf/

# 汇率数据同步
cd backend && go run ./cmd/syncrates/

# A股红利ETF数据更新
cd backend && go run ./cmd/updateashare/

# 数据源工厂测试
cd backend && go run ./cmd/test_factory/

# Finage API 测试
cd backend && go run ./cmd/test_finage/
```

### 数据获取策略 (v2.0 Finage-Only)

- **Finage 聚合API** (`/agg/stock/{symbol}/1/day`) 唯一数据源，提供完整 OHLCV + Volume
- **不降级到 last API** - last API 数据不完整，不符合"所有字段入库"要求
- **无 Yahoo Finance 依赖** - 所有数据从 Finage 获取并完整入库
- **Fallback Provider** 仅在没有 Finage API Key 时提供基础演示，不返回模拟假数据
- **所有字段必须入库** - OHLCV+Volume+DataSource 全部写入 etf_data 表
- **涨跌计算** 基于前一日收盘价 (PreviousClose)，从数据库查询真实数据
- **严格校验** - 数据不全则拒绝入库，避免脏数据污染数据库

### 常见问题

**Q: 端口被占用怎么办？**

修改 `backend/config.yaml` 中的端口配置，或停止占用端口的进程：

```bash
# macOS/Linux
lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill

# Windows
netstat -ano | findstr :8080
taskkill /PID <进程ID> /F
```

**Q: 依赖安装失败？**

国内用户可设置代理：
```bash
export GOPROXY=https://goproxy.cn,direct
npm config set registry https://registry.npmmirror.com
```

## 📅 定时任务

| 任务 | Cron 表达式 | 说明 |
|------|------------|------|
| ETF盘前更新 | `0 30 9 * * *` | 每天 09:30 |
| ETF收盘更新 | `0 30 16 * * *` | 每天 16:30 |
| 汇率频繁同步 | `0 */5 * * * *` | 每 5 分钟 |
| 汇率日同步 | `0 30 10 * * *` | 每天 10:30 |
| 小时检查 | `0 0 * * * *` | 每小时检查缓存状态 |

## 🗺️ 开发路线图

### Phase 1: 基础功能 ✅
- [x] ETF 基础信息管理
- [x] 实时行情数据展示
- [x] 汇率数据管理
- [x] A股红利ETF组合
- [x] 投资占比饼状图
- [x] ETF 对比分析
- [x] ETF 配置管理 (CRUD)
- [x] 投资组合配置管理

### Phase 2: 深度分析 🚧
- [x] 风险指标计算 (波动率、夏普比率、最大回撤)
- [x] ETF 收益预测与投资组合分析
- [x] 多数据源策略模式 (Finage/Finnub/Fallback)
- [x] 安全中间件 (速率限制/安全头)
- [ ] 持仓重叠分析
- [ ] 行业/地区分布可视化
- [ ] 相关性矩阵

### Phase 3: 组合优化 📋
- [ ] 投资组合构建器
- [ ] 有效前沿分析
- [ ] 再平衡策略建议

### Phase 4: 高级功能 📋
- [ ] 智能推荐系统
- [ ] 历史回测功能
- [ ] 投资报告导出
- [ ] 移动端适配

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 License

MIT License
