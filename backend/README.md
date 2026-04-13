# ETF-Insight Go Backend

高性能的 Go/Gin 后端服务，替代原有的 Python/Django 实现。

## 特性

- **高性能**: 使用 Go 语言和 Gin 框架，性能大幅提升
- **完整功能迁移**: 保留原有 Python 项目的所有功能
- **RESTful API**: 统一的 API 设计，与前端完美对接
- **数据持久化**: 支持 MySQL 和 SQLite
- **Redis 缓存**: 高性能缓存支持
- **定时任务**: 内置定时任务调度器
- **汇率服务**: 支持多货币转换和多数据源故障转移

## 项目结构

```
backend/
├── config/                 # 配置管理
│   ├── config.go          # 配置结构定义与加载
│   └── config_test.go     # 配置测试
├── models/                # 数据模型 (GORM)
│   ├── models.go          # ETFConfig, ETFData, OperationLog, PortfolioConfig
│   ├── db.go              # 数据库初始化与迁移
│   ├── exchange_rate.go   # ExchangeRate, ExchangeRateSyncLog, CurrencyPair
│   └── a_share_dividend_etf.go  # AShareDividendETF, AShareETFPortfolio
├── handlers/              # HTTP 处理器
│   ├── etf_handler.go     # ETF 行情/历史/指标/预测接口
│   ├── etf_config_handler.go    # ETF 配置 CRUD 接口
│   ├── portfolio_handler.go     # 投资组合分析/配置接口
│   ├── a_share_portfolio_handler.go  # A股红利ETF组合接口
│   ├── exchange_rate.go   # 汇率管理接口
│   ├── health_handler.go  # 健康检查 (health/ready/live)
│   └── middleware.go      # 日志与 CORS 中间件
├── services/              # 业务逻辑层
│   ├── datasource/        # 数据源微服务层
│   │   ├── provider.go    # 数据源接口定义 + ProviderFactory
│   │   ├── errors.go      # 标准错误定义
│   │   ├── finage_provider.go   # Finage API 实现
│   │   ├── finnhub_provider.go  # Finnhub API 实现
│   │   └── fallback_provider.go # 后备数据源
│   ├── exchange_rate/     # 汇率服务微服务层
│   │   ├── datasource/    # 汇率数据源管理
│   │   │   ├── provider.go # 汇率数据源接口
│   │   │   ├── manager.go  # 数据源管理器 (含故障转移)
│   │   │   ├── openexchange.go # Open Exchange Rates 实现
│   │   │   ├── currencyapi.go  # CurrencyAPI 实现
│   │   │   ├── frankfurter.go  # Frankfurter 实现
│   │   │   └── fallback.go     # 后备数据源
│   │   ├── monitor/       # 健康监控
│   │   ├── sync/          # 数据同步
│   │   └── service.go     # 汇率服务主逻辑
│   ├── sync/              # 同步服务层
│   │   ├── service.go     # 同步业务逻辑 + 入库校验 + 操作日志
│   │   └── config.go      # ETF配置数据 + 预设组合
│   ├── etf_analysis.go    # ETF分析服务 (指标/组合/预测/对比)
│   ├── yahoo_finance.go   # Yahoo Finance 客户端
│   └── finnhub.go         # Finnhub 独立客户端
├── middleware/            # 中间件
│   ├── security.go        # 安全头 + 速率限制 (100/min)
│   └── security_test.go
├── tasks/                 # 定时任务
│   ├── scheduler.go       # 主调度器 (ETF更新/汇率更新)
│   └── exchange_rate_task.go  # 汇率同步任务 (5min/10:30daily)
├── utils/                 # 工具包
│   ├── logger.go          # 日志工具
│   └── logger_test.go
├── cmd/                   # 命令行工具
│   ├── syncetf/           # ETF数据同步工具
│   ├── update_etf_prices/ # ETF价格批量更新工具 (Finage聚合API)
│   ├── generate_history/  # 生成模拟历史数据
│   ├── initetf/           # ETF初始数据导入
│   ├── syncrates/         # 汇率数据同步
│   ├── updateashare/      # A股红利ETF数据更新
│   ├── test_factory/      # 数据源工厂测试
│   └── test_finage/       # Finage API 测试
├── main.go                # 入口文件
└── go.mod                 # Go 模块配置
```

## 快速开始

### 1. 安装依赖

```bash
cd backend
go mod tidy
```

### 2. 配置环境变量

```bash
# 数据库配置
export DB_DRIVER=sqlite  # 或 mysql
export DB_DSN=etf_insight.db

# Finage API Key (必须配置)
export FINAGE_API_KEY=your_finage_api_key_here

# 后端服务配置
export SERVER_PORT=8080
export SERVER_HOST=localhost

# 缓存配置
export REDIS_ENABLED=false
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=

# 日志级别
export LOG_LEVEL=info
```

### 3. 启动服务

```bash
# 开发模式
go run main.go

# 生产模式
go build -o etf-insight
./etf-insight
```

## API 接口

### ETF 相关接口
- `GET /api/etf/list` - 获取 ETF 列表
- `GET /api/etf/detail/:symbol` - 获取 ETF 详情
- `GET /api/etf/history/:symbol` - 获取 ETF 历史数据
- `GET /api/etf/compare` - ETF 对比分析

### 投资组合接口
- `GET /api/portfolio/analysis` - 投资组合分析
- `POST /api/portfolio/config` - 配置投资组合
- `GET /api/portfolio/returns` - 组合收益率计算

### 汇率接口
- `GET /api/exchange-rate` - 获取汇率数据
- `GET /api/exchange-rate/sync-log` - 获取同步日志

### 健康检查
- `GET /health` - 服务健康状态
- `GET /ready` - 服务就绪状态
- `GET /live` - 服务存活状态

## 数据源配置

### ETF 数据源
- **主数据源**: Finage API (唯一真实数据源)
- **同步频率**: 定时任务自动更新
- **数据质量**: 实时数据、完整字段、入库校验

### 汇率数据源
- **主数据源**: Open Exchange Rates
- **备用数据源**: CurrencyAPI、Frankfurter
- **故障转移**: 自动切换、健康检查
- **同步策略**: 5分钟间隔、数据一致性保证

## 开发指南

### 代码规范
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 重要函数和复杂逻辑必须注释
- 新功能必须包含单元测试

### 数据模型
- 使用 GORM ORM 框架
- 所有字段必须入库
- 数据一致性校验
- 操作日志记录

### 错误处理
- 统一的错误码和错误信息
- 详细的日志记录
- 友好的错误提示

## 部署说明

### 开发环境
```bash
# 使用 SQLite 数据库
go run main.go
```

### 生产环境
```bash
# 使用 PostgreSQL 数据库
export DB_DRIVER=postgres
export DB_DSN="host=localhost user=etf dbname=etf_insight sslmode=disable"

# 启用 Redis 缓存
export REDIS_ENABLED=true
export REDIS_HOST=redis-server
export REDIS_PORT=6379

# 构建并运行
go build -o etf-insight
./etf-insight
```

## 监控和日志

### 日志配置
- 日志级别: DEBUG, INFO, WARN, ERROR
- 日志文件: `logs/app.log`
- 汇率同步日志: `logs/exchange_rate.log`

### 健康监控
- 服务健康状态: `GET /health`
- 数据源可用性检查
- 数据库连接状态

## 故障排除

### 常见问题
1. **数据源连接失败**: 检查网络连接和 API Key 配置
2. **数据库连接失败**: 检查数据库配置和连接状态
3. **汇率数据不一致**: 系统会自动故障转移，检查日志确认当前数据源

### 日志查看
```bash
# 查看应用日志
tail -f logs/app.log

# 查看汇率同步日志
tail -f logs/exchange_rate.log

# 查看错误日志
grep "ERROR" logs/app.log
```

## 版本更新

### v2.3 更新内容
- ✅ 汇率服务多数据源故障转移
- ✅ 竞态条件修复和性能优化
- ✅ 代码质量全面优化

### 未来计划
- 🔄 智能分析引擎开发
- 🔄 回测框架集成
- 🔄 微服务架构准备

---

**更多信息请参考**: [项目主文档](../README.md) | [核心上下文文档](../agents.md)