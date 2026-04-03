# ETF-Insight

一个专业的 ETF 数据管理与分析平台，基于 Go + React 技术栈，提供完整的 ETF 基础信息、持仓数据、行情数据、技术指标等一站式解决方案。

## 🌟 核心特性

### 基础数据层（Data Layer）
- **ETF基础信息** - 发行方、费率、AUM、成立时间、跟踪指数
- **实时 & 历史行情** - K线数据（1分钟至月线）、成交量、换手率、买卖盘
- **持仓数据（核心）** - 前10大持仓、行业分布、地区分布、季度调仓
- **净值数据** - NAV、市价、溢价率
- **分红数据** - 分红金额、除息日、股息率
- **技术指标** - MA、RSI、MACD、夏普比率、最大回撤

### 智能数据管理
- 🔄 **自动定时更新** - 美股开盘前和收盘后自动更新数据
- ⚡ **并发获取** - 多线程并发拉取，提升效率
- 🛡️ **智能重试** - 指数退避重试机制，应对 API 限制
- 💾 **双存储架构** - MySQL 持久化 + Redis 缓存

### 动态配置管理
- 📋 **ETF动态配置** - 支持美股/A股/港股，灵活增删改查
- 📊 **策略管理** - 支持质量股息、高股息收益、期权增强等多种策略
- 🌐 **汇率管理** - 自动更新人民币、港币、美元汇率
- 📝 **操作日志** - 完整记录所有系统操作

### Web 可视化
- 📈 **仪表盘** - 实时数据概览、关键指标展示
- 📊 **K线图表** - 交互式价格图表，支持多周期切换
- 🔄 **持仓分析** - 持仓明细、行业分布、地区分布可视化
- 📉 **技术指标** - MA、RSI、MACD 等技术指标展示
- 📊 **对比分析** - 多 ETF 对比分析

## 📊 支持的 ETF 策略

| 策略类型 | 示例 ETF | 描述 |
|---------|---------|------|
| **质量股息** | SCHD | Schwab U.S. Dividend Equity ETF |
| **高股息收益** | SPYD | SPDR Portfolio S&P 500 High Dividend ETF |
| **期权增强收益** | JEPQ | JPMorgan Nasdaq Equity Premium Income ETF |
| **股息增强** | JEPI | JPMorgan Equity Premium Income ETF |
| **高股息宽基** | VYM | Vanguard High Dividend Yield ETF |
| **科技成长** | QQQ | Invesco QQQ Trust |
| **标普500** | SPY | SPDR S&P 500 ETF Trust |

## 🛠️ 技术栈

### 后端 (Go)
- **Go 1.21+** - 核心语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **cron** - 定时任务调度
- **logrus** - 日志框架
- **go-cache** - 内存缓存

### 数据库
- **MySQL** - 主数据库，存储所有历史数据
- **Redis** - 缓存层，提升查询性能
- **SQLite** - 开发环境轻量级数据库

### 前端 (React)
- **React 18** - 前端框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Ant Design** - UI 组件库
- **ECharts** - 数据可视化

## 🚀 快速开始

### 方式一：使用 Docker（推荐）

```bash
# 克隆项目
git clone git@github.com:coder100001/ETF-Insight.git
cd ETF-Insight

# 启动所有服务
docker-compose up -d

# 访问 http://localhost:8080
```

### 方式二：本地开发

#### 1. 启动后端服务

```bash
cd backend

# 配置 Go 代理（国内用户）
go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖
go mod tidy

# 启动服务
go run main.go

# 或带配置文件启动
go run main.go -config=config.yaml

# 初始化数据库
go run main.go -init-db
```

后端服务默认运行在 http://localhost:8080

#### 2. 启动前端服务

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问 http://localhost:5173
```

## 📁 项目结构

```
ETF-Insight/
├── backend/                 # Go 后端服务
│   ├── config/             # 配置管理
│   │   └── config.go
│   ├── handlers/           # HTTP 处理器
│   │   ├── admin.go
│   │   ├── etf.go
│   │   ├── exchange_rate.go
│   │   ├── portfolio.go
│   │   ├── scheduler.go
│   │   └── workflow.go
│   ├── models/             # 数据模型
│   │   ├── db.go
│   │   └── models.go
│   ├── routers/            # 路由配置
│   │   └── router.go
│   ├── services/           # 业务逻辑
│   │   ├── cache.go
│   │   ├── etf_analysis.go
│   │   ├── exchange_rate.go
│   │   └── yahoo_finance.go
│   ├── tasks/              # 定时任务
│   │   └── scheduler.go
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── README.md
├── frontend/               # React 前端
│   ├── src/
│   │   ├── components/     # UI 组件
│   │   ├── pages/          # 页面
│   │   ├── services/       # API 服务
│   │   ├── styles/         # 样式
│   │   ├── types/          # TypeScript 类型
│   │   └── utils/          # 工具函数
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile              # Docker 构建文件
├── docker-compose.yml      # Docker Compose 配置
├── Makefile               # 构建脚本
└── README.md
```

## 📖 API 接口

### ETF 配置 API

```http
GET    /api/etf-configs          # 获取所有 ETF 配置
POST   /api/etf-configs          # 添加 ETF 配置
GET    /api/etf-configs/:id      # 获取 ETF 详情
PUT    /api/etf-configs/:id      # 更新 ETF
DELETE /api/etf-configs/:id      # 删除 ETF
PATCH  /api/etf-configs/:id      # 切换启用/禁用状态
```

### ETF 数据 API

```http
GET    /api/etf-data/:symbol          # 获取 ETF 价格数据
GET    /api/etf-nav/:symbol           # 获取 ETF 净值数据
GET    /api/etf-holdings/:symbol      # 获取 ETF 持仓数据
GET    /api/etf-sectors/:symbol       # 获取 ETF 行业分布
GET    /api/etf-regions/:symbol       # 获取 ETF 地区分布
GET    /api/etf-rebalances/:symbol    # 获取 ETF 调仓记录
GET    /api/etf-dividends/:symbol     # 获取 ETF 分红数据
GET    /api/etf-indicators/:symbol    # 获取 ETF 技术指标
```

### 汇率 API

```http
GET    /api/exchange-rates           # 获取所有汇率
POST   /api/exchange-rates           # 添加汇率
GET    /api/exchange-rates/latest    # 获取最新汇率
PUT    /api/exchange-rates/:id       # 更新汇率
```

## 🗄️ 数据模型

### ETF 基础信息 (ETFConfig)
- ETF代码、名称、英文名称
- 市场（US/CN/HK）、资产类别
- 发行方、费率、AUM
- 跟踪指数、成立日期
- 启用/禁用状态

### ETF 价格数据 (ETFPrice)
- 开高低收、昨收价
- 成交量、成交额
- 涨跌额、涨跌幅、换手率
- 时间周期（1分钟至月线）

### 持仓数据 (ETFHolding)
- 持仓代码、名称、资产类型
- 持仓数量、市值、权重
- 报告日期

### 行业分布 (ETFHoldingSector)
- 行业名称、权重
- 市值、股票数量

### 地区分布 (ETFHoldingRegion)
- 地区名称、国家、权重
- 市值、股票数量

## ⏰ 定时任务

| 任务 | 时间 | 说明 |
|------|------|------|
| 汇率更新 | 每天 10:30 | 更新人民币、港币、美元汇率 |
| ETF 盘前更新 | 每天 9:30 ET | 美股开盘前更新数据 |
| ETF 收盘更新 | 每天 16:30 ET | 美股收盘后更新数据 |
| 持仓同步 | 每周日 20:00 | 同步最新持仓数据 |
| 技术指标计算 | 每日 22:00 | 计算技术指标 |

## 🔧 配置管理

### 后端配置 (config.yaml)

```yaml
server:
  port: 8080
  mode: debug  # debug/release

database:
  driver: mysql  # mysql/sqlite
  host: localhost
  port: 3306
  user: root
  password: password
  name: etf_insight
  charset: utf8mb4

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

scheduler:
  enabled: true
  timezone: "America/New_York"
```

### ETF 动态配置

访问 http://localhost:8080/api/etf-configs 管理 ETF

功能：
- ✅ 美股/A股/港股分 Tab 展示
- ✅ 添加/编辑/删除 ETF
- ✅ 启用/禁用 ETF
- ✅ 统计信息（总数、美股数、A股数、启用数）
- ✅ 排序显示

## 📝 开发计划

### 已完成 ✅
- [x] Go 后端框架搭建
- [x] ETF 基础数据模型
- [x] ETF 配置管理 API
- [x] 汇率管理功能
- [x] 定时任务调度
- [x] React 前端框架
- [x] ETF 配置管理页面
- [x] Docker 容器化

### 进行中 🚧
- [ ] ETF 持仓数据获取
- [ ] ETF 行情数据展示
- [ ] K线图表组件
- [ ] 技术指标计算

### 待开发 📋
- [ ] ETF 持仓数据可视化
- [ ] 行业分布图表
- [ ] 地区分布图表
- [ ] ETF 对比分析
- [ ] 投资组合分析
- [ ] 数据导出功能
- [ ] 用户权限控制
- [ ] 邮件通知功能

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License

## 📧 联系

如有问题或建议，请通过以下方式联系：
- 提交 Issue
- 发送邮件至：coder100001@gmail.com
