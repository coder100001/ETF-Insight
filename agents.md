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

# ========== JWT安全配置 ==========
JWT_SECRET_KEY=your-jwt-secret-key-min-32-characters-long
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168

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

## 🗺️ 演进路线图 (v2.4 → v3.0) - 开源专业分析平台

> **核心理念**: 坚持开源、专业、透明，打造学术界和业界认可的 ETF 量化分析基础设施

### 第一阶段：分析深度增强 (v2.4 - v2.5) - 当前

#### 技术基础完善
- ✅ **安全架构**: JWT认证、审计日志、数据验证、速率限制
- ✅ **数据质量**: 多数据源故障转移、数据完整性校验
- ✅ **性能优化**: 智能缓存、数据库索引优化、前端懒加载
- 🔄 **测试覆盖**: 单元测试覆盖率 >80%，核心算法 100% 覆盖

#### 量化分析能力
- ✅ **投资组合优化**: 马科维茨模型、有效前沿、夏普比率最大化
- 🔄 **技术指标**: RSI、MACD、布林带、均线系统
- 🔄 **风险模型**: VaR、CVaR、压力测试、情景分析
- 🔄 **因子分析**: Fama-French 三因子、Carhart 四因子模型

### 第二阶段：研究平台化 (v2.6 - v2.8) - 3-6个月

#### 学术研究支持
- 🔄 **回测引擎**: 事件驱动回测、滑点模拟、交易成本建模
- 🔄 **策略框架**: 策略模板、参数优化、 Walk-forward 分析
- 🔄 **论文复现**: 经典量化策略实现（动量、价值、低波动等）
- 🔄 **数据导出**: CSV/Excel/JSON 格式，支持学术研究

#### 数据源扩展
- 🔄 **插件架构**: 标准化数据源接口，支持自定义数据源
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

### v2.4 更新内容

#### 安全功能升级
- ✅ **JWT身份认证**: 完整的认证中间件，支持Token生成/验证/角色控制
  - `middleware/auth.go` - JWT认证中间件
  - 支持 `AuthRequired()` / `OptionalAuth()` / `RequireRole()` 三种模式
  - Token 过期时间可配置（默认24小时）
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
  - 防止暴力破解和DDoS攻击
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
- **更新同步**: 架构修改必须更新 agents.md

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
// JWT认证使用示例
authMiddleware := middleware.NewAuthMiddleware(&cfg.JWT)

// 需要认证的路由
router.Use(authMiddleware.AuthRequired())

// 可选认证（登录用户有额外功能）
router.Use(authMiddleware.OptionalAuth())

// 角色权限控制
router.Use(authMiddleware.RequireRole("admin", "editor"))

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
✅ API认证授权测试
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

### 3. 其他指标

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

## 🔒 安全要求

### 1. 安全边界

| 边界类型 | 要求 |
|----------|------|
| **输入验证** | 所有用户输入必须验证 |
| **SQL注入** | 使用参数化查询 |
| **XSS攻击** | React默认转义+ CSP头 |
| **CSRF** | Token验证 |
| **权限控制** | RBAC，最小权限原则 |

---

### 2. 数据加密标准

| 数据类型 | 加密方式 |
|----------|----------|
| **密码** | bcrypt, cost ≥ 12 |
| **API Key** | 环境变量，代码中不出现 |
| **敏感日志** | 脱敏处理 |
| **数据库** | TLS传输 |

---

### 3. 访问控制策略

```go
// 角色定义
const (
    RoleAdmin  Role = "admin"  // 管理员：全部权限
    RoleUser   Role = "user"   // 普通用户：自身数据
    RoleGuest  Role = "guest"  // 访客：只读
)

// 权限检查
func RequireRole(roles ...Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        for _, role := range roles {
            if Role(userRole) == role {
                c.Next()
                return
            }
        }
        c.JSON(403, gin.H{"error": "权限不足"})
        c.Abort()
    }
}
```

---

### 4. 安全审计

| 审计项 | 频率 | 记录 |
|--------|------|------|
| **登录日志** | 每次 | 用户ID、IP、时间 |
| **操作日志** | 每次 | 用户、动作、资源 |
| **错误日志** | 每次 | 错误类型、堆栈 |
| **安全扫描** | 每周 | 漏洞报告 |

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

## 🤖 AI 助手规范

### 1. 交互流程

```
用户请求 → 理解意图 → 查阅上下文 → 执行任务 → 验证结果 → 响应用户
    ↓           ↓            ↓           ↓           ↓           ↓
  自然语言    提取关键信息   agents.md   工具调用    测试验证    清晰反馈
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

*本文档最后更新: 2026-04-13 (v2.5 金融算法安全增强版)*
*强制上下文绑定版本: v2.0*
