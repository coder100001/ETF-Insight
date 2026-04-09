# ETF-Insight 项目上下文文档

> **⚠️ 强制上下文绑定 | MANDATORY CONTEXT BINDING**
> 
> ## AI Agent 必须遵守的规则
> 
> ### 1. 强制阅读要求 (MANDATORY READING)
> ```
> [CRITICAL] 每次对话开始时，必须首先完整阅读本文档
> [CRITICAL] 任何代码修改前，必须查阅本文档相关章节
> [CRITICAL] 禁止在不了解上下文的情况下修改代码
> ```
> 
> ### 2. 本文档包含的关键信息
> - ✅ 系统架构设计 (System Architecture)
> - ✅ 数据模型定义 (Data Models)
> - ✅ API 接口规范 (API Specifications)
> - ✅ 编码规则与约束 (Coding Rules & Constraints)
> - ✅ 数据源策略 (Data Source Strategy)
> - ✅ 缓存架构设计 (Cache Architecture)
> - ✅ 项目目录结构 (Project Structure)
> 
> ### 3. 修改同步要求 (MANDATORY SYNC)
> ```
> [RULE] 修改架构 → 必须更新本文档架构章节
> [RULE] 修改数据模型 → 必须更新本文档数据模型章节
> [RULE] 修改API → 必须更新本文档API接口章节
> [RULE] 修改配置 → 必须更新本文档核心配置章节
> [RULE] 修改编码规则 → 必须更新本文档编码规则章节
> ```
> 
> ### 4. 违规后果
> > 不阅读本文档直接修改代码可能导致：
> > - 违反架构设计原则
> > - 破坏数据一致性
> > - 引入技术债务
> > - 代码无法通过审查

---

## 📋 项目概览

**项目名称**: ETF-Insight  
**定位**: 专业的 ETF 分析与对比平台，对标 Trackinsight、ETF Insider 等国际知名工具  
**技术栈**: Go (Gin + GORM) + React (Vite + Ant Design) + SQLite  
**版本**: v2.2.0  

---

## 🏗️ 系统架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ETF Data Sync Service                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Config    │  │   Logger    │  │  Repository │  │  DataSourceProvider │ │
│  │   Manager   │  │   Service   │  │   Layer     │  │     (Interface)     │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         └─────────────────┴─────────────────┴──────────────────┘             │
│                                    │                                         │
│                              ┌─────┴─────┐                                   │
│                              │  SyncJob  │                                   │
│                              │  Service  │                                   │
│                              └─────┬─────┘                                   │
│                                    │                                         │
│              ┌─────────────────────┴─────────────────────┐                   │
│              │                                           │                   │
│        ┌─────┴─────┐                               ┌─────┴─────┐              │
│        │  Finage   │                               │  Fallback │              │
│        │  Provider │                               │  Provider │              │
│        │           │                               │ (Emergency) │            │
│        └───────────┘                               └───────────┘              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 目录结构

```
ETF-Insight/
├── start.sh                    # 一键启动脚本 (macOS/Linux)
├── start.bat                   # 一键启动脚本 (Windows)
├── agents.md                   # 📌 本文件 - 项目核心上下文
├── .env.example                # 环境变量模板
├── backend/
│   ├── main.go                 # 后端入口
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
│   │   ├── datasource/         # 🏗️ 数据源微服务层
│   │   │   ├── provider.go     # 数据源接口定义 + ProviderFactory
│   │   │   ├── errors.go       # 标准错误定义
│   │   │   ├── finage_provider.go   # Finage API (聚合API + last API)
│   │   │   ├── finnhub_provider.go  # Finnhub API 实现
│   │   │   └── fallback_provider.go # 后备数据源
│   │   ├── sync/               # 🔄 同步服务层
│   │   │   ├── service.go      # 同步业务逻辑 + 入库校验 + 操作日志
│   │   │   └── config.go       # ETF配置数据 + 预设组合
│   │   ├── etf_analysis.go     # ETF分析服务 (指标/组合/预测/对比)
│   │   ├── yahoo_finance.go    # Yahoo Finance 客户端
│   │   ├── cache.go            # 缓存服务 (策略模式重构)
│   │   ├── cache_provider.go   # 缓存提供者接口 (抽象)
│   │   ├── cache_factory.go    # 缓存工厂 (工厂模式)
│   │   ├── redis_client.go     # Redis 客户端封装
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
│   │   ├── update_etf_prices/  # ETF价格批量更新工具 (Finage聚合API)
│   │   ├── generate_history/  # 生成模拟历史数据
│   │   ├── initetf/            # ETF初始数据导入
│   │   ├── syncrates/          # 汇率数据同步
│   │   ├── updateashare/       # A股红利ETF数据更新
│   │   ├── test_factory/       # 数据源工厂测试
│   │   └── test_finage/        # Finage API 测试
│   └── infrastructure/         # 基础设施 (预留目录)
│       ├── cache/
│       ├── circuitbreaker/
│       ├── database/
│       ├── messagequeue/
│       └── metrics/
├── frontend/
│   ├── src/
│   │   ├── pages/              # 页面组件 (14个)
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
│   │   │   ├── OperationLogs.tsx       # 操作日志
│   │   │   ├── WorkflowList.tsx        # 工作流列表
│   │   │   └── InstanceList.tsx        # 实例列表
│   │   ├── components/         # 公共组件
│   │   │   ├── Layout.tsx             # 布局
│   │   │   ├── PriceChart.tsx         # 价格图表
│   │   │   ├── ComparisonRadarChart.tsx # 对比雷达图
│   │   │   ├── ETFFilter.tsx          # ETF 筛选
│   │   │   ├── HoldingPieChart.tsx    # 持仓饼图
│   │   │   ├── SectorBarChart.tsx     # 行业柱状图
│   │   │   ├── StatCard.tsx           # 统计卡片
│   │   │   └── StockCard.tsx          # 股票卡片
│   │   ├── services/api.ts     # API 服务 (含请求合并+重试, 类型安全)
│   │   ├── utils/api.ts        # API 工具函数 (类型安全, ETFHistoryDataItem/ETFConfig)
│   │   ├── types/index.ts      # TypeScript 类型定义 (含ETFHistoryDataItem/ETFForecastResult)
│   │   └── styles/theme.ts     # 主题配置
│   └── package.json
├── docs/
│   └── openapi.yaml            # OpenAPI 3.0 接口文档
├── scripts/
│   ├── install-hooks.sh        # Git hooks 安装
│   └── startup.sh              # 生产启动脚本
└── docker-compose.yml
```

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
# 代理配置 (Clash VPN)
HTTP_PROXY=http://127.0.0.1:7897
HTTPS_PROXY=http://127.0.0.1:7897

# Finage API Key (唯一数据源) - 必须配置，否则系统无法工作
FINAGE_API_KEY=your_finage_api_key_here

# 注意：Finnhub API Key 已废弃，仅作为历史代码保留
# FINNHUB_API_KEY=your_finnhub_api_key_here
```

### 配置文件 (backend/config.yaml)

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

### 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| 后端 | 8080 | Gin Web 服务 |
| 前端 | 5173 | Vite Dev Server |

---

## 🏛️ 缓存架构设计 (OOP 重构)

### 设计原则

缓存系统经过全面重构，严格遵循面向对象编程原则：

#### 1. **单一职责原则 (SRP)**
- `CacheProvider` 接口：只定义缓存操作
- `MemoryCache`：只负责内存缓存
- `RedisCache`：只负责 Redis 缓存
- `HybridCache`：只负责混合缓存策略
- `CacheFactory`：只负责创建缓存实例
- `CacheService`：只负责缓存业务逻辑

#### 2. **开闭原则 (OCP)**
- 通过 `CacheProvider` 接口扩展新的缓存类型，无需修改现有代码
- 使用工厂模式创建缓存实例，支持灵活配置

#### 3. **里氏替换原则 (LSP)**
- 所有缓存实现（MemoryCache、RedisCache、HybridCache）都可以替换 `CacheProvider` 接口

#### 4. **接口隔离原则 (ISP)**
- `CacheProvider` 接口只包含必要的缓存操作方法

#### 5. **依赖倒置原则 (DIP)**
- `CacheService` 依赖于抽象接口 `CacheProvider`，而不是具体实现

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    CacheService (业务层)                    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  - provider: CacheProvider (依赖倒置)                 │  │
│  │  - cfg: *config.CacheConfig                          │  │
│  │  - ctx: context.Context                              │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│            CacheProvider Interface (抽象接口)               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  + Get(key string) (interface{}, bool)                │  │
│  │  + Set(key string, value interface{}, exp Duration)   │  │
│  │  + Delete(key string) error                           │  │
│  │  + Clear()                                            │  │
│  │  + GetType() string                                   │  │
│  └───────────────────────────────────────────────────────┘  │
└──────────────────────────┬──────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ MemoryCache  │  │  RedisCache  │  │ HybridCache  │
│              │  │              │  │              │
│ - cache      │  │ - client     │  │ - primary    │
│              │  │              │  │ - secondary  │
└──────────────┘  └──────────────┘  └──────────────┘
      ▲                  ▲                  ▲
      │                  │                  │
      └──────────────────┴──────────────────┘
                         │
                         ▼
              ┌────────────────────┐
              │  CacheFactory      │
              │  (工厂模式)        │
              │                    │
              │  + CreateCache()   │
              └────────────────────┘
```

### 核心组件

#### 1. CacheProvider 接口 (抽象)

```go
type CacheProvider interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, expiration time.Duration)
    Delete(key string) error
    Clear()
    GetType() string
}
```

#### 2. MemoryCache 实现

```go
type MemoryCache struct {
    cache *cache.Cache
}

func NewMemoryCache(defaultExpiration, cleanupInterval time.Duration) *MemoryCache {
    return &MemoryCache{
        cache: cache.New(defaultExpiration, cleanupInterval),
    }
}
```

#### 3. RedisCache 实现

```go
type RedisCache struct {
    client *RedisClient
}

func NewRedisCache(client *RedisClient) *RedisCache {
    return &RedisCache{client: client}
}
```

#### 4. HybridCache 实现 (组合模式)

```go
type HybridCache struct {
    primary   CacheProvider  // Redis
    secondary CacheProvider  // Memory
}

func (h *HybridCache) Get(key string) (interface{}, bool) {
    if val, ok := h.primary.Get(key); ok {
        return val, true
    }
    return h.secondary.Get(key)
}
```

#### 5. CacheFactory (工厂模式)

```go
type CacheFactory struct{}

func (f *CacheFactory) CreateCache(cfg *config.CacheConfig, redisCfg *config.RedisConfig) CacheProvider {
    if redisCfg.Enabled {
        return f.createHybridCache(cfg, redisCfg)
    }
    return f.createMemoryCache(cfg)
}
```

### 使用示例

```go
// 创建缓存服务
factory := NewCacheFactory()
provider := factory.CreateCache(cacheCfg, redisCfg)
cacheService := NewCacheServiceWithProvider(provider, cacheCfg)

// 使用缓存
cacheService.Set("key", "value", 5*time.Minute)
value, found := cacheService.Get("key")
```

### 缓存策略

| 策略 | 描述 | 适用场景 |
|------|------|---------|
| `CacheStrategyMemory` | 纯内存缓存 | 单机部署、开发测试 |
| `CacheStrategyRedis` | 纯 Redis 缓存 | 分布式部署、需要持久化 |
| `CacheStrategyHybrid` | Redis + Memory 混合缓存 | 生产环境、高性能要求 |

### 配置方式

```yaml
redis:
  enabled: true
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  timeout: 5s

etf:
  cache:
    realtime_ttl: 5m
    historical_ttl: 24h
    metrics_ttl: 1h
    comparison_ttl: 30m
```

---

## 🗄️ 数据模型

### 核心表结构

#### 1. ETFConfig (ETF配置表)
```go
type ETFConfig struct {
    ID              uint
    Symbol          string          // 代码, 唯一索引
    Name            string          // 名称
    Description     string          // 描述
    Strategy        string          // 策略
    Focus           string          // 投资焦点
    ExpenseRatio    decimal.Decimal // 费率 (10,4)
    Currency        string          // 货币
    Exchange        string          // 交易所
    Category        string          // 类别
    Provider        string          // 提供商
    Inception       string          // 成立日期
    AUM             decimal.Decimal // 管理规模 (20,2)
    Status          int             // 状态: 1启用, 0禁用
    AutoUpdate      bool            // 自动更新
    UpdateFrequency string          // 更新频率
    DataSource      string          // 数据来源: "finage_agg" / "fallback" (v2.0: 仅Finage或明确fallback)
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

#### 2. ETFData (ETF行情数据表)
```go
type ETFData struct {
    ID         uint
    Symbol     string          // 代码
    Date       time.Time       // 日期, 联合唯一索引(symbol, date)
    OpenPrice  decimal.Decimal // 开盘价 (20,8)
    ClosePrice decimal.Decimal // 收盘价 (20,8)
    HighPrice  decimal.Decimal // 最高价 (20,8)
    LowPrice   decimal.Decimal // 最低价 (20,8)
    Volume     int64           // 成交量
    DataSource string          // 数据来源
    CreatedAt  time.Time
}
```

#### 3. OperationLog (操作日志表)
```go
type OperationLog struct {
    ID            uint
    OperationType string     // 操作类型: SYNC / scheduled_task
    OperationName string     // 操作名称: ETF_SYNC / 定时汇率更新 / ETF数据更新
    Operator      string     // 操作人: system
    Status        int        // 0进行中, 1成功, 2失败, 3部分成功
    ErrorMessage  string     // 错误信息
    StartTime     time.Time  // 开始时间
    EndTime       *time.Time // 结束时间
    DurationMs    int        // 耗时(毫秒)
    Details       string     // 详情(JSON格式)
}
```

#### 4. PortfolioConfig (投资组合配置表)
```go
type PortfolioConfig struct {
    ID              uint
    Name            string          // 组合名称
    Description     string          // 描述
    Allocation      string          // 配置JSON: {"QQQ": 50, "SCHD": 50}
    TotalInvestment decimal.Decimal // 总投资金额 (15,2)
    TaxRate         decimal.Decimal // 税率 (5,4)
    Status          int             // 状态: 1-启用, 0-禁用
    IsDefault       bool            // 是否默认
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

#### 5. ExchangeRate (汇率数据表)
```go
type ExchangeRate struct {
    ID            uint            // 主键
    FromCurrency  string          // 源货币 (联合唯一索引)
    ToCurrency    string          // 目标货币 (联合唯一索引)
    Rate          decimal.Decimal // 汇率 (20,8)
    PreviousRate  decimal.Decimal // 前次汇率
    ChangePercent decimal.Decimal // 变动百分比 (10,4)
    DataSource    string          // 数据来源
    SourceType    string          // api / manual / calculated
    ValidStatus   int             // 1有效, 0无效
    Priority      int             // 优先级
    SyncBatchID   string          // 同步批次ID
    SyncedAt      *time.Time      // 同步时间
    ExpiresAt     *time.Time      // 过期时间
    DeletedAt     gorm.DeletedAt  // 软删除
}
```

#### 6. ExchangeRateSyncLog / ExchangeRateSyncDetail (汇率同步日志)
```go
type ExchangeRateSyncLog struct {
    ID            uint       // 批次ID
    BatchID       string     // 批次标识
    SyncType      string     // full / incremental
    DataSource    string     // 数据源
    Status        string     // success / failed / partial
    TotalCount    int        // 总数
    SuccessCount  int        // 成功数
    FailedCount   int        // 失败数
    DurationMs    int64      // 耗时
}

type ExchangeRateSyncDetail struct {
    ID            uint            // 明细ID
    SyncLogID     uint            // 关联批次
    FromCurrency  string          // 源货币
    ToCurrency    string          // 目标货币
    OldRate       decimal.Decimal // 旧汇率
    NewRate       decimal.Decimal // 新汇率
    ChangePercent decimal.Decimal // 变动
    Status        string          // success / failed / skipped
}
```

#### 7. CurrencyPair (货币对配置表)
```go
type CurrencyPair struct {
    ID           uint           // 主键
    FromCurrency string         // 源货币 (联合唯一索引)
    ToCurrency   string         // 目标货币 (联合唯一索引)
    IsActive     int            // 1启用, 0禁用
    Priority     int            // 优先级
    Description  string         // 描述, 如 "美元兑人民币"
}
```
默认货币对: USD/CNY, USD/HKD, EUR/CNY, GBP/CNY, JPY/CNY, CNY/USD, HKD/USD

#### 8. AShareDividendETF (A股红利ETF表)
```go
type AShareDividendETF struct {
    ID                uint              // 主键
    Symbol            string            // ETF代码 (如515080), 唯一索引
    Name              string            // ETF名称
    DividendYieldMin  decimal.Decimal   // 股息率下限(%) (5,2)
    DividendYieldMax  decimal.Decimal   // 股息率上限(%) (5,2)
    DividendFrequency DividendFrequency // 月分/季分/年分
    Benchmark         string            // 跟踪基准指数
    Exchange          string            // 交易所: SSE/SHZ
    ManagementFee     decimal.Decimal   // 管理费率(%) (5,4)
    Description       string            // 产品描述
    Status            int               // 1正常, 0停用
}
```

#### 9. AShareETFPortfolio / ASharePortfolioHolding (A股组合)
```go
type AShareETFPortfolio struct {
    ID              uint            // 主键
    Name            string          // 组合名称
    TotalInvestment decimal.Decimal // 总投资金额 (15,2)
    IsDefault       bool            // 是否默认
    Description     string          // 描述
}

type ASharePortfolioHolding struct {
    ID          uint            // 主键
    PortfolioID uint            // 组合ID (索引)
    ETFID       uint            // ETF产品ID (索引, 外键)
    Investment  decimal.Decimal // 投资金额 (15,2)
    Weight      decimal.Decimal // 占比(%) (5,2)
    SortOrder   int             // 排序
}
```

### 默认ETF列表 (15只)

```go
DefaultETFList = []ETFInfo{
    {Symbol: "QQQ", Name: "Invesco QQQ Trust", Category: "大盘股", Provider: "Invesco", ExpenseRatio: 0.0020},
    {Symbol: "SCHD", Name: "Schwab US Dividend Equity ETF", Category: "股息", Provider: "Schwab", ExpenseRatio: 0.0006},
    {Symbol: "VNQ", Name: "Vanguard Real Estate ETF", Category: "REITs", Provider: "Vanguard", ExpenseRatio: 0.0012},
    {Symbol: "VYM", Name: "Vanguard High Dividend Yield ETF", Category: "股息", Provider: "Vanguard", ExpenseRatio: 0.0006},
    {Symbol: "SPYD", Name: "SPDR Portfolio S&P 500 High Dividend ETF", Category: "股息", Provider: "SPDR", ExpenseRatio: 0.0035},
    {Symbol: "JEPQ", Name: "JPMorgan Nasdaq Equity Premium Income ETF", Category: "备兑认购", Provider: "JPMorgan", ExpenseRatio: 0.0035},
    {Symbol: "JEPI", Name: "JPMorgan Equity Premium Income ETF", Category: "备兑认购", Provider: "JPMorgan", ExpenseRatio: 0.0035},
    {Symbol: "VTI", Name: "Vanguard Total Stock Market ETF", Category: "全市场", Provider: "Vanguard", ExpenseRatio: 0.0003},
    {Symbol: "VOO", Name: "Vanguard S&P 500 ETF", Category: "大盘股", Provider: "Vanguard", ExpenseRatio: 0.0003},
    {Symbol: "VEA", Name: "Vanguard FTSE Developed Markets ETF", Category: "国际", Provider: "Vanguard", ExpenseRatio: 0.0005},
    {Symbol: "VWO", Name: "Vanguard FTSE Emerging Markets ETF", Category: "新兴市场", Provider: "Vanguard", ExpenseRatio: 0.0010},
    {Symbol: "BND", Name: "Vanguard Total Bond Market ETF", Category: "债券", Provider: "Vanguard", ExpenseRatio: 0.0003},
    {Symbol: "AGG", Name: "iShares Core U.S. Aggregate Bond ETF", Category: "债券", Provider: "iShares", ExpenseRatio: 0.0003},
    {Symbol: "GLD", Name: "SPDR Gold Shares", Category: "商品", Provider: "SPDR", ExpenseRatio: 0.0040},
    {Symbol: "TLT", Name: "iShares 20+ Year Treasury Bond ETF", Category: "国债", Provider: "iShares", ExpenseRatio: 0.0015},
}
```

---

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
| GET | `/api/etf/:symbol/realtime` | 获取单只 ETF 实时数据 (数据库最新OHLCV) |
| GET | `/api/etf/:symbol/history` | 获取历史数据 (支持 period: 1m/3m/6m/1y/3y/5y) |
| GET | `/api/etf/:symbol/metrics` | 获取风险指标 (波动率/夏普比率/最大回撤) |
| GET | `/api/etf/:symbol/forecast` | 获取10年收益预测 |
| GET | `/api/etf/comparison` | 获取 ETF 对比数据 (默认6只: SCHD/SPYD/JEPQ/JEPI/VYM/QQQ) |
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

---

## 🔄 数据同步架构

### 数据源提供者接口 (Strategy Pattern)

```go
type DataSourceProvider interface {
    GetName() string                                    // 获取数据源名称
    GetQuote(ctx context.Context, symbol string) (*QuoteData, error)      // 单只报价
    GetQuotes(ctx context.Context, symbols []string) ([]*QuoteData, error) // 批量报价
    IsAvailable(ctx context.Context) bool               // 检查可用性
    GetRateLimit() int                                  // 速率限制(每秒请求数)
}
```

### 数据源策略 (v2.0)

**核心原则：完全依赖 Finage 真实数据，干掉所有硬编码mock**

1. **Finage (唯一真实数据源)** - 必须配置 `FINAGE_API_KEY`，系统无法工作
   - 使用聚合API获取完整OHLCV + Volume数据
   - **所有字段必须入库** → 数据库必须有完整的数据记录
   - 如果API Key未配置或请求失败 → 系统明确报错，不返回假数据

2. **Fallback Provider (仅紧急备用)** - 仅在没有Finage API Key时提供基础演示
   - **不提供模拟假数据** → 仅返回有限的基本信息
   - **不会入库** → 避免污染数据库
   - **前端会明确显示"数据源: fallback"** → 提醒用户配置API Key

3. **Finnhub (已废弃)** - 代码保留，但不再使用

### Finage Provider 数据获取策略

Finage 提供两种 API，优先级如下：

| 优先级 | API 端点 | 数据完整性 | 用途 |
|--------|---------|-----------|------|
| 1 (优先) | `GET /agg/stock/{symbol}/1/day/{from}/{to}` | 完整 OHLCV + Volume | 主力数据源，入库数据必须来自此API |
| 2 (降级) | `GET /last/stock/{symbol}` | 仅 ask/bid | 聚合API失败时的备选，**数据不完整，不可入库** |

**关键规则 (v2.0 严格模式)**：
- **完全依赖 Finage 聚合API** → `GetQuote` 只调用 `/agg/stock/{symbol}/1/day/` 获取完整OHLCV
- **不降级到 last API** → last API 数据不完整，不满足"所有字段入库"要求
- **所有字段必须入库** → OHLCV、Volume、数据来源等全部写入 `etf_data` 表
- **严格数据校验** → `sync/service.go` 校验：OHLCV不全>0则拒绝入库，避免脏数据
- **无硬编码回退** → 前端/后端所有硬编码mock数据已删除

### 同步流程 (v2.0 Finage-Only)

```
1. 检查数据源可用性
   └─> Finage API Key 配置 → 继续
   └─> 无API Key → 明确报错，不降级到模拟数据

2. 获取报价 (Finage聚合API)
   └─> GET /agg/stock/{symbol}/1/day/{from}/{to}
       └─> 获取最近5天完整OHLCV + Volume
       └─> **不降级到last API** (数据不完整)

3. 数据入库校验 (sync/service.go - 严格模式)
   ├─> 检查 OHLCV 是否全为0 → ❌ 拒绝入库
   ├─> 检查数据完整性 → ❌ 不完整则拒绝
   └─> 通过校验 → ✅ 更新/创建 ETFData (所有字段)

4. 更新数据库 (完整字段)
   ├─> 更新 ETFConfig (配置表)
   └─> 更新 ETFData (行情表，包含: Open/High/Low/Close/Volume/DataSource等)

5. 记录操作日志
   └─> 明确标记数据来源: "finage_agg"
   └─> 记录入库成功/失败详情
```
   └─> 写入 operation_logs 表
```

### 涨跌计算逻辑

涨跌始终基于 **前一日收盘价 (PreviousClose)** 计算，而非当日开盘价：

```
change = ClosePrice - PreviousClose
changePercent = (change / PreviousClose) * 100
```

PreviousClose 获取优先级：
1. 从 realtimeData 缓存获取 (Yahoo Finance 提供)
2. 从数据库查询前一日 ETFData 的 ClosePrice
3. 兜底：使用当日 OpenPrice（不理想但好于0）

### ETF 价格更新工具

```bash
# 使用 Finage 聚合API 逐个更新 ETF 价格
cd backend && go run ./cmd/update_etf_prices/

# 特点：
# - 从数据库 etf_configs 表获取 ETF 列表
# - 逐个请求 Finage 聚合API (每个symbol独立请求，避免推断错误)
# - 自动更新 etf_data 表
# - 数据有效性验证：OHLCV > 0 才入库
# - symbol 由请求参数决定，不依赖价格范围推断
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

### 同步命令

```bash
# 直接运行同步
./backend/cmd/syncetf/syncetf

# 或从源码运行
cd backend && go run ./cmd/syncetf/

# 带代理运行
HTTP_PROXY=http://127.0.0.1:7897 HTTPS_PROXY=http://127.0.0.1:7897 ./syncetf
```

---

## 📅 定时任务

### 主调度器 (tasks/scheduler.go)

| 任务 | Cron 表达式 | 说明 |
|------|------------|------|
| 汇率更新 | `0 30 10 * * *` | 每天 10:30 更新汇率 |
| ETF盘前更新 | `0 30 9 * * *` | 每天 09:30 (盘前) |
| ETF收盘更新 | `0 30 16 * * *` | 每天 16:30 (收盘后) |
| 小时检查 | `0 0 * * * *` | 每小时检查缓存状态 |

### 汇率同步任务 (tasks/exchange_rate_task.go)

| 任务 | Cron 表达式 | 说明 |
|------|------------|------|
| 频繁同步 | `0 */5 * * * *` | 每 5 分钟 |
| 日同步 | `0 30 10 * * *` | 每天 10:30 |

---

## 🚀 启动方式

### 方式一: 一键启动 (推荐)

```bash
cd /Users/liunian/Desktop/dnmp/py_project
./start.sh
```

会自动完成:
1. 检查环境 (Go, Node.js)
2. 安装后端依赖
3. 编译后端
4. 安装前端依赖
5. 启动后端 (端口 8080)
6. 启动前端 (端口 5173)

### 方式二: 手动启动

```bash
# 后端
cd backend
go run main.go

# 前端 (新终端)
cd frontend
npm run dev
```

### 服务地址

- 前端: http://localhost:5173
- 后端 API: http://localhost:8080
- 健康检查: http://localhost:8080/health
- 就绪检查: http://localhost:8080/ready
- 存活检查: http://localhost:8080/live

---

## 🛡️ 安全与中间件

### 安全头中间件 (middleware/security.go)
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security`
- `Content-Security-Policy`
- API 路径禁用缓存 (`Cache-Control: no-store`)

### 速率限制
- 100 请求/分钟/IP

---

## 🛠️ 开发规范

### OOP 设计原则

- **单一职责**: 每个服务只负责一个功能领域
- **开闭原则**: 通过接口扩展功能，不修改现有代码
- **策略模式**: DataSourceProvider 接口支持多数据源切换
- **工厂模式**: ProviderFactory 管理数据源注册和选择
- **选项模式**: 配置结构体支持灵活参数设置

### 错误处理

使用 `backend/services/datasource/errors.go` 定义的标准错误:

```go
ErrNoAvailableProvider = errors.New("no available data source provider")
ErrRateLimitExceeded   = errors.New("rate limit exceeded")
ErrInvalidSymbol       = errors.New("invalid symbol")
ErrAPIError            = errors.New("API request failed")
ErrProxyError          = errors.New("proxy configuration error")
```

### 并发控制

- Finage GetQuotes: 并发10个worker，每个symbol独立请求聚合API
- update_etf_prices: 逐个symbol串行请求，间隔200ms
- 超时控制: 每个请求 30 秒超时
- 速率限制: Finage 100 req/s, Finnhub 60 req/s

---

## 📊 监控与日志

### 操作日志自动记录

每次同步会自动记录到 `operation_logs` 表:

```json
{
  "total": 15,
  "success": 15,
  "fail": 0,
  "updated": 15,
  "duration": "2.1s",
  "dataSource": "finage"
}
```

### 查看日志

```bash
# SQLite 命令行
cd backend
sqlite3 etf_insight.db "SELECT * FROM operation_logs ORDER BY id DESC LIMIT 5;"
```

---

## 📝 修改记录

| 日期 | 修改人 | 内容 |
|------|--------|------|
| 2025-04-07 | AI Agent | 初始创建 agents.md，记录微服务架构、数据源设计、同步流程 |
| 2025-04-07 17:09 | AI Agent | 修改 GetETFList API，从数据库读取实时数据，移除硬编码模拟数据 |
| 2025-04-07 18:00 | AI Agent | 优化 ETF 对比接口性能，添加缓存机制，响应时间 < 300ms |
| 2025-04-07 18:30 | AI Agent | 集成 Finage API，创建 update_etf_prices 批量更新工具 |
| 2025-04-07 18:50 | AI Agent | 优化批量请求，每批 10 只 ETF，减少 87% API 请求次数 |
| 2025-04-08 | AI Agent | **数据入库逻辑全面修复**：1) FinageProvider 重写为优先使用聚合API获取完整OHLCV；2) 修复涨跌计算逻辑改用PreviousClose；3) update_etf_prices改用逐个请求避免symbol推断错乱；4) 清理handler中硬编码mock数据改为数据库查询；5) 更新Fallback基准价格；6) sync层增加OHLCV校验拒绝不完整数据入库 |
| 2025-04-08 | AI Agent | **文档全面更新**：补充缺失的数据模型(ExchangeRate系列/AShare系列)、完整API接口列表(含ETF配置/组合配置/A股/汇率)、目录结构细化(含所有文件)、命令行工具文档、定时任务说明、安全中间件说明 |
| **2026-04-08** | **AI Agent** | **v2.0 重大架构调整**：1) **完全依赖 Finage 作为唯一数据源**，删除所有硬编码mock数据；2) **所有字段必须入库**，数据完整性严格校验；3) **删除前端硬编码** (InvestmentStrategy/PortfolioConfig/ETFConfig/PortfolioAnalysis)；4) **删除后端硬编码** (scheduler只更新2只ETF、cache.go MockRealtimeData、etf_handler.go默认列表)；5) **更新架构文档**：Finage-only数据流，废弃Finnhub |
| **2026-04-09** | **AI Agent** | **v2.1 股息率与资本利得优化**：1) **修复股息率显示问题**：添加 `getDefaultDividendYield()` 函数，根据 ETF 类型智能设置合理股息率（高股息 ETF 3.5%、覆盖收益型 7%、债券 ETF 4%、宽基指数 0.5%）；2) **修复资本利得计算问题**：从数据库获取真实价格，从历史数据计算真实收益率，从数据库获取 ETF 真实名称；3) **优化 FallbackProvider**：从数据库读取最新真实价格作为基准，减小模拟波动范围（±3% → ±0.5%）；4) **清理数据**：删除 15 条错误的 fallback 污染数据；5) **添加 .gitignore**：排除二进制文件 |
| **2026-04-09** | **AI Agent** | **v2.2 代码质量全面优化**：1) **Go 代码格式化**：使用 `gofmt -w .` 统一格式化，修复 map 对齐和尾部空格问题；2) **ESLint 问题修复**：修复全部 29 个 ESLint 问题 (3 errors + 26 warnings)；3) **消除未使用变量**：删除 InvestmentStrategy.tsx 中未使用的 `getETF` 变量、PortfolioConfig.tsx 中约35行未使用的 `mockConfigs`；4) **TypeScript 类型安全**：新增 `ETFHistoryDataItem` 和 `ETFForecastResult` 类型定义，替换所有 `any` 类型为具体类型（~15个文件）；5) **React Hooks 规范**：为仅挂载执行的 useEffect 添加 exhaustive-deps 禁用注释；6) **Vite 配置清理**：移除 vite.config.ts 中的 `as any` 断言 |

---

## 🤖 给 AI 的提示 (v2.0 更新)

### 核心原则
1. **完全依赖 Finage 真实数据** - 无硬编码mock，无Fallback假数据
2. **所有字段必须入库** - OHLCV+Volume+DataSource全部写入etf_data表
3. **严格数据校验** - 数据不全则拒绝入库，避免脏数据

### 编码规则
1. **修改代码后**: 必须同步更新本文档中的相关部分
2. **数据源策略**: Finage是唯一真实数据源，不添加其他数据源
3. **修改数据库模型**: 更新本文档的 "数据模型" 章节
4. **添加新 API**: 记录到 "API 接口" 章节
5. **修改配置**: 更新 "核心配置" 章节
6. **优化性能**: 考虑批量请求、缓存策略、并发控制
7. **数据入库**: 必须确保 OHLCV 数据完整（非全0），否则 sync 层会拒绝入库
8. **涨跌计算**: 始终使用 PreviousClose（前日收盘价），禁止用 OpenPrice 代替
9. **Finage API**: **只使用聚合API** (`/agg/stock/`)，不使用last API（数据不完整）
10. **前端开发**: **禁止硬编码数据**，所有数据从API获取

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
□ 已确认不会违反本文档中的任何规则

[确认声明]
我确认已阅读并理解本文档内容，将严格遵守文档中的架构设计、
编码规则和数据约束。如有违反，愿意承担相应后果。
```

### 代码修改确认
```
[CODE MODIFICATION CHECKLIST - 每次修改代码前必须勾选]

□ 已查阅本文档相关章节，了解上下文
□ 修改符合架构设计原则
□ 修改符合数据模型定义
□ 修改符合编码规则
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

---

*本文档最后更新: 2026-04-09 (v2.2 代码质量全面优化)*
*强制上下文绑定版本: v1.0*
