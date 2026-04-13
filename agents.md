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
CSRF_SECRET=your-random-secret-key-min-32-chars

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

## � 演进路线图 (v2.3 → v3.0)

### 第一阶段：短期增强 (1-3个月) - v2.3 分析深度增强版

#### 技术优化
- ✅ **性能提升**: 智能缓存策略、数据库查询优化、前端懒加载
- ✅ **代码质量**: 单元测试覆盖率 >80%、自动化代码检查、CI/CD流水线
- ✅ **竞态条件修复**: 汇率数据源管理器故障转移逻辑优化

#### 功能增强
- 🔄 **分析深度提升**: 技术指标计算(RSI/MACD/布林带)、趋势判断算法、基本面评分
- 🔄 **用户体验优化**: 个性化预警、移动端优化、数据导出功能
- 🔄 **数据丰富**: 更多ETF数据维度、宏观经济数据关联、数据质量监控

### 第二阶段：中期升级 (3-6个月) - v2.5 智能分析版

#### 技术架构升级
- 🔄 **微服务化准备**: 模块化重构、消息队列引入、服务发现
- 🔄 **数据架构优化**: 时序数据库、数据湖架构、数据血缘追踪

#### 智能功能开发
- ✅ **智能分析引擎**: 投资组合优化API、风险预警模型
- ✅ **专业工具集成**: 梯度下降优化算法、有效前沿生成
- 🔄 **回测框架**: 业绩归因分析、风险因子分析
- 🔄 **协作功能**: 团队协作、分析报告共享、评论讨论

### 第三阶段：长期战略 (6-12个月) - v3.0 专业平台版

#### 平台化架构
- 🔄 **云原生转型**: 容器化部署、多租户架构、弹性伸缩
- 🔄 **开放平台建设**: API开放平台、第三方集成、插件化架构

#### 商业模式探索
- 🔄 **产品矩阵**: 个人免费版、专业付费版、企业定制版
- 🔄 **生态建设**: 开发者社区、合作伙伴计划、数据市场

---

## 🔧 最新技术更新

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
- **更新同步**: 架构修改必须更新 agents.md

---

## 🔒 安全规范

### 1. API Key 管理
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
1. 检查 agents.md 文档是否包含相关信息
2. 查看日志文件定位问题
3. 验证数据源连接状态
4. 检查数据库连接和配置

### 紧急处理
- **数据源故障**: 自动故障转移机制已启用
- **服务不可用**: 检查健康检查接口 `/health`
- **数据不一致**: 使用数据验证工具检查数据完整性

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

*本文档最后更新: 2026-04-13 (v2.3.2 Pre-commit钩子版)*
*强制上下文绑定版本: v2.0*
