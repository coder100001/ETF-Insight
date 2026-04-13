# ETF-Insight 项目上下文文档（优化版）

> **⚠️ AI Agent 强制规则 | MANDATORY RULES FOR AI AGENTS**
>
> ## 核心原则
> ```
> [CRITICAL] 每次对话开始时，必须首先完整阅读本文档
> [CRITICAL] 任何代码修改前，必须查阅本文档相关章节
> [CRITICAL] 禁止在不了解上下文的情况下修改代码
> [CRITICAL] 修改架构/数据模型/API 后，必须同步更新本文档
> ```
>
> ## 违规后果
> - ❌ 违反架构设计原则
> - ❌ 破坏数据一致性
> - ❌ 引入技术债务
> - ❌ 代码无法通过审查

---

## 📋 项目概览

| 项目信息 | 详情 |
|---------|------|
| **项目名称** | ETF-Insight |
| **定位** | 专业的 ETF 分析与对比平台 |
| **对标产品** | Trackinsight、ETF Insider |
| **技术栈** | Go (Gin + GORM) + React (Vite + Ant Design) + SQLite |
| **当前版本** | v2.3.0 |
| **演进阶段** | 分析深度增强版 |

### 核心价值主张
- 🎯 **ETF 对比分析**: 多维度并排对比，发现最优投资标的
- 🔍 **持仓深度解构**: 穿透底层资产，了解真实风险敞口
- 📊 **风险指标评估**: 波动率、夏普比率、最大回撤等专业指标
- 💼 **投资组合优化**: 基于现代投资组合理论的资产配置

---

## 🏗️ 系统架构

### 架构图

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
│        │ (Primary) │                               │ (Emergency)│            │
│        └───────────┘                               └───────────┘              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 分层架构

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │  ← React 前端
├─────────────────────────────────────────┤
│         API Layer (Handlers)            │  ← Gin HTTP Handlers
├─────────────────────────────────────────┤
│         Business Logic (Services)       │  ← 业务逻辑层
├─────────────────────────────────────────┤
│         Data Access (Repository)        │  ← GORM 数据访问
├─────────────────────────────────────────┤
│         Infrastructure                  │  ← 数据源、缓存、日志
└─────────────────────────────────────────┘
```

### 目录结构（精简版）

```
ETF-Insight/
├── agents.md                   # 📌 本文件 - 项目核心上下文
├── .env.example                # 环境变量模板
├── start.sh / start.bat        # 一键启动脚本
│
├── backend/                    # Go 后端服务
│   ├── main.go                 # 入口文件
│   ├── config/                 # 配置管理
│   │   └── config.go           # 配置结构定义
│   ├── models/                 # 数据模型
│   │   ├── models.go           # ETF 核心模型
│   │   ├── exchange_rate.go    # 汇率模型
│   │   └── db.go               # 数据库初始化
│   ├── handlers/               # API 处理器
│   │   ├── etf_handler.go      # ETF 接口
│   │   ├── portfolio_handler.go # 组合接口
│   │   └── exchange_rate.go    # 汇率接口
│   ├── services/               # 业务逻辑层
│   │   ├── datasource/         # 数据源微服务
│   │   │   ├── provider.go     # 接口定义
│   │   │   ├── finage_provider.go  # Finage 实现
│   │   │   └── fallback_provider.go # 后备数据源
│   │   ├── exchange_rate/      # 汇率服务
│   │   │   ├── datasource/     # 汇率数据源管理
│   │   │   │   ├── manager.go  # 故障转移管理器
│   │   │   │   ├── openexchange.go # Open Exchange Rates
│   │   │   │   ├── currencyapi.go  # CurrencyAPI
│   │   │   │   └── frankfurter.go  # Frankfurter
│   │   │   └── service.go      # 汇率服务主逻辑
│   │   └── etf_analysis.go     # ETF 分析服务
│   ├── middleware/             # 中间件
│   │   └── security.go         # 安全中间件
│   ├── tasks/                  # 定时任务
│   │   ├── scheduler.go        # 主调度器
│   │   └── exchange_rate_task.go # 汇率同步任务
│   └── utils/                  # 工具包
│       └── logger.go           # 日志工具
│
├── frontend/                   # React 前端应用
│   ├── src/
│   │   ├── pages/              # 页面组件（14个）
│   │   ├── components/         # 公共组件
│   │   ├── services/api.ts     # API 服务
│   │   ├── types/index.ts      # TypeScript 类型定义
│   │   └── utils/api.ts        # API 工具函数
│   └── package.json
│
└── docs/
    └── openapi.yaml            # OpenAPI 3.0 接口文档
```

---

## 🔑 核心配置

### 环境变量配置

```bash
# ========== 必须配置 ==========
FINAGE_API_KEY=your_finage_api_key_here  # ETF 数据源（必须）

# ========== 汇率数据源（推荐配置） ==========
OPENEXCHANGE_API_KEY=your_key_here       # 主数据源
CURRENCYAPI_KEY=your_key_here            # 备份数据源
# Frankfurter 免费无需 API Key

# ========== 代理配置（可选） ==========
HTTP_PROXY=http://127.0.0.1:7897
HTTPS_PROXY=http://127.0.0.1:7897

# ========== 数据库配置 ==========
DB_DSN=etf_insight.db                    # SQLite 数据库文件

# ========== 服务器配置 ==========
SERVER_PORT=8080
SERVER_HOST=localhost
LOG_LEVEL=info
```

### 数据源策略

| 数据类型 | 主数据源 | 备用数据源 | 故障转移 |
|---------|---------|-----------|---------|
| **ETF 数据** | Finage API | Fallback Provider | ❌ 手动 |
| **汇率数据** | Open Exchange Rates | CurrencyAPI, Frankfurter | ✅ 自动 |

---

## 📊 数据模型

### 核心数据模型

#### 1. ETFConfig - ETF 配置
```go
type ETFConfig struct {
    ID              uint            `json:"id"`
    Symbol          string          `json:"symbol" gorm:"uniqueIndex"`
    Name            string          `json:"name"`
    Description     string          `json:"description"`
    Strategy        string          `json:"strategy"`
    Focus           string          `json:"focus"`
    ExpenseRatio    decimal.Decimal `json:"expense_ratio"`
    Currency        string          `json:"currency"`
    Exchange        string          `json:"exchange"`
    Category        string          `json:"category"`
    Provider        string          `json:"provider"`
    AUM             decimal.Decimal `json:"aum"`
    Status          int             `json:"status" gorm:"default:1"`
    AutoUpdate      bool            `json:"auto_update" gorm:"default:true"`
    DataSource      string          `json:"data_source" gorm:"default:'Finage'"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}
```

#### 2. ETFData - ETF 历史数据
```go
type ETFData struct {
    ID         uint            `json:"id" gorm:"primaryKey"`
    Symbol     string          `json:"symbol" gorm:"uniqueIndex:idx_symbol_date"`
    Date       time.Time       `json:"date" gorm:"uniqueIndex:idx_symbol_date"`
    OpenPrice  decimal.Decimal `json:"open_price"`
    ClosePrice decimal.Decimal `json:"close_price"`
    HighPrice  decimal.Decimal `json:"high_price"`
    LowPrice   decimal.Decimal `json:"low_price"`
    Volume     int64           `json:"volume"`
    DataSource string          `json:"data_source"`
    CreatedAt  time.Time       `json:"created_at"`
}
```

#### 3. ExchangeRate - 汇率数据
```go
type ExchangeRate struct {
    ID            uint            `gorm:"primaryKey"`
    FromCurrency  string          `gorm:"size:10;not null;index:idx_currency_pair,unique"`
    ToCurrency    string          `gorm:"size:10;not null;index:idx_currency_pair,unique"`
    Rate          decimal.Decimal `gorm:"type:decimal(20,8);not null"`
    PreviousRate  decimal.Decimal `gorm:"type:decimal(20,8)"`
    ChangePercent decimal.Decimal `gorm:"type:decimal(10,4)"`
    DataSource    string          `gorm:"size:50;not null"`
    ValidStatus   int             `gorm:"default:1"`
    SyncBatchID   string          `gorm:"size:50;index"`
    SyncedAt      *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### 4. PortfolioConfig - 投资组合配置
```go
type PortfolioConfig struct {
    ID              uint            `json:"id" gorm:"primaryKey"`
    Name            string          `json:"name" gorm:"size:100;not null"`
    Description     string          `json:"description" gorm:"size:500"`
    Allocation      string          `json:"allocation" gorm:"type:text;not null"` // JSON
    TotalInvestment decimal.Decimal `json:"total_investment"`
    TaxRate         decimal.Decimal `json:"tax_rate"`
    Status          int             `json:"status" gorm:"default:1"`
    IsDefault       bool            `json:"is_default" gorm:"default:false"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}
```

### 数据库索引策略

```sql
-- ETFData 表索引
CREATE UNIQUE INDEX idx_symbol_date ON etf_data(symbol, date);
CREATE INDEX idx_symbol ON etf_data(symbol);
CREATE INDEX idx_date ON etf_data(date);

-- ExchangeRate 表索引
CREATE UNIQUE INDEX idx_currency_pair ON exchange_rates(from_currency, to_currency);
CREATE INDEX idx_sync_batch_id ON exchange_rates(sync_batch_id);

-- ETFConfig 表索引
CREATE UNIQUE INDEX idx_symbol ON etf_configs(symbol);
CREATE INDEX idx_status ON etf_configs(status);
```

---

## 🔌 API 接口规范

### RESTful API 设计原则

1. **统一响应格式**:
```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "error": null
}
```

2. **HTTP 状态码**:
- `200 OK`: 成功
- `201 Created`: 创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未授权
- `403 Forbidden`: 禁止访问
- `404 Not Found`: 资源不存在
- `429 Too Many Requests`: 请求过多
- `500 Internal Server Error`: 服务器错误

### 核心 API 端点

#### ETF 数据接口
```
GET    /api/etf/list                    # 获取 ETF 列表
GET    /api/etf/comparison              # 获取对比数据
GET    /api/etf/:symbol/realtime        # 获取实时数据
GET    /api/etf/:symbol/history         # 获取历史数据
GET    /api/etf/:symbol/metrics         # 获取指标数据
GET    /api/etf/:symbol/forecast        # 获取收益预测
POST   /api/etf/update-realtime         # 更新实时数据
```

#### 投资组合接口
```
GET    /api/portfolio-configs/          # 获取组合列表
POST   /api/portfolio-configs/          # 创建组合
GET    /api/portfolio-configs/:id       # 获取组合详情
PUT    /api/portfolio-configs/:id       # 更新组合
DELETE /api/portfolio-configs/:id       # 删除组合
POST   /api/portfolio-configs/:id/analyze  # 分析组合
POST   /api/etf/portfolio               # 分析自定义组合
```

#### 汇率管理接口
```
GET    /api/exchange-rates              # 获取汇率列表
GET    /api/exchange-rates/:from/:to    # 获取特定汇率
POST   /api/exchange-rates/convert      # 货币转换
POST   /api/exchange-rates/sync         # 触发同步
GET    /api/exchange-rates/summary      # 获取汇率摘要
GET    /api/exchange-rates/datasource-status  # 数据源状态
```

#### 健康检查接口
```
GET    /health                          # 健康检查
GET    /ready                           # 就绪检查
GET    /live                            # 存活检查
```

---

## 🛠️ 开发规范

### Go 代码规范

#### 1. 命名规范
```go
// ✅ 正确：驼峰命名，首字母大写表示导出
type DataSourceProvider interface {}
func GetETFList() {}
var DefaultConfig Config

// ❌ 错误：下划线命名
type data_source_provider interface {}
func get_etf_list() {}
```

#### 2. 错误处理
```go
// ✅ 正确：返回自定义错误类型
func GetRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
    if from == "" || to == "" {
        return decimal.Zero, &ValidationError{
            Field: "currency",
            Message: "currency cannot be empty",
        }
    }
    // ...
}

// ❌ 错误：忽略错误
rate, _ := GetRate(ctx, "USD", "CNY")
```

#### 3. 并发安全
```go
// ✅ 正确：使用互斥锁保护共享资源
type Manager struct {
    current DataSourceProvider
    mu      sync.RWMutex
}

func (m *Manager) GetCurrent() DataSourceProvider {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.current
}

// ❌ 错误：直接访问共享资源
func (m *Manager) GetCurrent() DataSourceProvider {
    return m.current  // 竞态条件！
}
```

#### 4. Context 传递
```go
// ✅ 正确：第一个参数传递 context
func FetchData(ctx context.Context, symbol string) (*Data, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // 执行操作
    }
}

// ❌ 错误：不传递 context
func FetchData(symbol string) (*Data, error) {
    // 无法取消或超时
}
```

### TypeScript 代码规范

#### 1. 类型安全
```typescript
// ✅ 正确：明确的类型定义
interface ETFData {
  symbol: string;
  price: number;
  volume: number;
}

function getETF(symbol: string): Promise<ETFData> {
  return fetch(`/api/etf/${symbol}`).then(r => r.json());
}

// ❌ 错误：使用 any 类型
function getETF(symbol: string): Promise<any> {
  return fetch(`/api/etf/${symbol}`).then(r => r.json());
}
```

#### 2. React Hooks 规范
```typescript
// ✅ 正确：依赖数组完整
useEffect(() => {
  fetchData(symbol);
}, [symbol]);

// ❌ 错误：缺少依赖
useEffect(() => {
  fetchData(symbol);
}, []); // ESLint 警告
```

#### 3. 错误处理
```typescript
// ✅ 正确：完整的错误处理
try {
  const data = await fetchETF(symbol);
  setData(data);
} catch (error) {
  console.error('Failed to fetch ETF:', error);
  setError(error instanceof Error ? error.message : 'Unknown error');
}

// ❌ 错误：忽略错误
const data = await fetchETF(symbol);
setData(data);
```

### 数据库操作规范

#### 1. 使用事务
```go
// ✅ 正确：使用事务保证一致性
func UpdatePortfolio(portfolio *PortfolioConfig) error {
    return models.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Save(portfolio).Error; err != nil {
            return err
        }
        // 其他操作
        return nil
    })
}

// ❌ 错误：不使用事务
func UpdatePortfolio(portfolio *PortfolioConfig) error {
    return models.DB.Save(portfolio).Error
}
```

#### 2. 参数化查询
```go
// ✅ 正确：使用参数化查询
models.DB.Where("symbol = ?", symbol).First(&etf)

// ❌ 错误：字符串拼接（SQL 注入风险）
models.DB.Where(fmt.Sprintf("symbol = '%s'", symbol)).First(&etf)
```

---

## 🔒 安全规范

### 1. API Key 管理
```go
// ✅ 正确：从环境变量读取
apiKey := os.Getenv("FINAGE_API_KEY")

// ❌ 错误：硬编码
apiKey := "your_api_key_here"
```

### 2. 输入验证
```go
// ✅ 正确：验证输入
type CreateETFRequest struct {
    Symbol string `json:"symbol" binding:"required,min=1,max=10"`
    Name   string `json:"name" binding:"required,min=1,max=100"`
}

if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}

// ❌ 错误：不验证输入
var req CreateETFRequest
c.BindJSON(&req)
// 直接使用 req
```

### 3. 日志脱敏
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

### 4. CORS 配置
```go
// ✅ 正确：限制允许的域名
allowedOrigins := []string{
    "http://localhost:3000",
    "https://yourdomain.com",
}

// ❌ 错误：允许所有域名
c.Header("Access-Control-Allow-Origin", "*")
```

---

## 🎯 演进路线图

### v2.3 (当前版本) - 分析深度增强版
- ✅ 汇率服务多数据源故障转移
- ✅ 竞态条件修复和性能优化
- ✅ 代码质量全面优化（ESLint 零错误）
- ✅ TypeScript 类型安全（无 any 类型）

### v2.5 (计划中) - 智能分析版
- 🔄 技术指标计算（RSI/MACD/布林带）
- 🔄 智能分析引擎开发
- 🔄 回测框架集成
- 🔄 微服务架构准备

### v3.0 (长期规划) - 专业平台版
- 🔄 云原生转型（容器化部署）
- 🔄 开放平台建设（API 开放）
- 🔄 商业模式探索（付费版）

---

## 📈 成功指标

### 技术指标
| 指标 | 当前值 | 目标值 | 状态 |
|------|--------|--------|------|
| 系统可用性 | 99.9% | 99.9% | ✅ |
| API 响应时间 | <200ms | <200ms | ✅ |
| 测试覆盖率 | ~60% | >80% | 🔄 |
| 代码质量评分 | 4/5 | 4.5/5 | 🔄 |

### 业务指标
- 🔄 用户活跃度提升 30%
- 🔄 功能使用率提升 50%
- 🔄 用户满意度 > 4.5/5
- 🔄 留存率提升 20%

---

## 🚨 常见问题排查

### 1. 数据源连接失败
```bash
# 检查 API Key 配置
echo $FINAGE_API_KEY

# 检查网络连接
curl -I https://api.finage.co.uk

# 检查代理配置
echo $HTTP_PROXY
```

### 2. 汇率数据不一致
```bash
# 查看当前数据源
curl http://localhost:8080/api/exchange-rates/datasource-status

# 触发手动同步
curl -X POST http://localhost:8080/api/exchange-rates/sync

# 查看同步日志
tail -f backend/logs/exchange_rate.log
```

### 3. 数据库连接问题
```bash
# 检查数据库文件
ls -lh backend/etf_insight.db

# 检查数据库连接
sqlite3 backend/etf_insight.db "SELECT COUNT(*) FROM etf_configs;"
```

---

## 📞 技术支持

### 紧急处理流程
1. **数据源故障**: 系统自动故障转移，无需人工干预
2. **服务不可用**: 检查健康检查接口 `/health`
3. **数据不一致**: 使用数据验证工具检查数据完整性

### 日志查看
```bash
# 后端日志
tail -f backend/logs/app.log

# 汇率同步日志
tail -f backend/logs/exchange_rate.log

# 错误日志
grep "ERROR" backend/logs/app.log
```

---

## 🔄 文档维护规则

### 必须更新本文档的情况
1. ✅ 修改系统架构
2. ✅ 修改数据模型
3. ✅ 修改 API 接口
4. ✅ 修改核心配置
5. ✅ 修改编码规则
6. ✅ 添加新功能模块

### 文档更新流程
1. 修改代码
2. 更新 agents.md 相关章节
3. 提交 PR 时说明文档变更
4. Code Review 时检查文档同步

---

**文档版本**: v2.3.0  
**最后更新**: 2026-04-13  
**维护者**: ETF-Insight Team  
**下次审查**: 2026-05-13

> **重要提醒**: 本文档是项目的核心上下文，所有开发者和 AI Agent 必须遵守本文档的规范和约束。
