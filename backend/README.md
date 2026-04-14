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
- **投资组合优化**: 马科维茨模型、有效前沿、夏普比率最大化
- **风险指标计算**: 波动率、最大回撤、夏普比率、Beta、Alpha
- **技术指标**: 支持扩展各类技术分析指标
- **多因子模型**: Fama-French 等因子分析框架

### 数据服务
- **多数据源支持**: Finage 主数据源 + 备用数据源故障转移
- **汇率服务**: 多数据源汇率，自动故障转移
- **A股ETF支持**: 中证红利、红利低波等主流红利ETF
- **数据完整性**: 所有字段入库校验，操作日志记录

### 安全架构
- **JWT认证**: 完整的认证中间件，支持角色控制
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
│   ├── exchange_rate.go   # 汇率相关模型
│   ├── a_share_dividend_etf.go  # A股ETF模型
│   ├── audit_log.go       # 审计日志模型
│   └── pagination.go      # 分页模型
├── handlers/              # HTTP 处理器
│   ├── etf_handler.go     # ETF 行情/历史/指标接口
│   ├── portfolio_handler.go     # 投资组合分析接口
│   ├── a_share_portfolio_handler.go  # A股ETF组合接口
│   ├── exchange_rate.go   # 汇率管理接口
│   ├── health_handler.go  # 健康检查
│   └── middleware.go      # 中间件
├── middleware/            # 中间件
│   ├── auth.go            # JWT认证中间件
│   ├── audit.go           # 审计日志中间件
│   ├── validation.go      # 数据验证中间件
│   ├── security.go        # 安全头 + 速率限制
│   └── security_test.go
├── services/              # 业务逻辑层
│   ├── datasource/        # 数据源微服务层
│   ├── exchange_rate/     # 汇率服务
│   ├── sync/              # 同步服务
│   ├── etf_analysis.go    # ETF分析服务
│   └── portfolio_optimizer.go  # 组合优化服务
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
# - JWT_SECRET_KEY (必需，用于认证)
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

## 📚 API 文档

### ETF 分析接口
- `GET /api/etf/list` - ETF列表（支持分页）
- `GET /api/etf/detail/:symbol` - ETF详情
- `GET /api/etf/history/:symbol` - 历史数据
- `GET /api/etf/metrics/:symbol` - 风险指标
- `GET /api/etf/compare` - ETF对比分析

### 投资组合接口
- `POST /api/portfolio/optimize` - 组合优化
- `POST /api/portfolio/efficient-frontier` - 有效前沿
- `GET /api/portfolio/analysis` - 组合分析

### A股ETF接口
- `GET /api/a-share/etfs` - A股ETF列表
- `GET /api/a-share/prices` - ETF价格
- `POST /api/a-share/prices/refresh` - 刷新价格

### 汇率接口
- `GET /api/exchange-rates` - 汇率列表
- `GET /api/exchange-rates/:currency` - 单币种汇率

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
// JWT认证
authMiddleware := middleware.NewAuthMiddleware(&cfg.JWT)
router.Use(authMiddleware.AuthRequired())

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

```bash
# 运行所有测试
go test -v ./...

# 运行特定包测试
go test -v ./services/...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 数据源策略

### ETF 数据源
- **主数据源**: Finage API
- **数据质量**: 实时数据、完整字段、入库校验
- **同步频率**: 定时任务自动更新

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
- [架构文档](../agents.md)
