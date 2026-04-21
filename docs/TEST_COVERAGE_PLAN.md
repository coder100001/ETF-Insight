# ETF-Insight 测试覆盖提升计划

**版本**: v1.0
**制定日期**: 2026-04-21
**执行周期**: 2026-04-21 至 2026-06-30 (10周)
**目标覆盖率**: 整体 70%，核心业务 90%

---

## 📊 当前测试覆盖现状

### 后端覆盖率统计 (2026-04-21)

| 模块 | 当前覆盖率 | 目标覆盖率 | 缺口 | 优先级 |
|------|-----------|-----------|------|--------|
| **services/factor** | 80.8% | 90% | +9.2% | P1 |
| **services/technical** | 100% | 100% | 0% | ✅ |
| **services/risk** | 100% | 100% | 0% | ✅ |
| **services/optimization** | 0% | 70% | +70% | P0 |
| **services/backtest** | 0% | 70% | +70% | P0 |
| **services/ashare** | 0% | 60% | +60% | P1 |
| **services/exchange_rate** | 27.3% | 70% | +42.7% | P1 |
| **handlers** | 37.1% | 70% | +32.9% | P1 |
| **middleware** | 68.8% | 80% | +11.2% | P2 |
| **models** | 56.7% | 70% | +13.3% | P2 |
| **utils** | 81.2% | 90% | +8.8% | P2 |
| **整体** | ~55% | 70% | +15% | - |

### 前端覆盖率统计

| 模块 | 当前覆盖率 | 目标覆盖率 | 缺口 | 优先级 |
|------|-----------|-----------|------|--------|
| **components** | 0% | 60% | +60% | P1 |
| **pages** | 0% | 50% | +50% | P2 |
| **services** | 0% | 70% | +70% | P1 |
| **utils** | 0% | 80% | +80% | P1 |
| **整体** | 0% | 60% | +60% | - |

---

## 🎯 测试策略

### 测试金字塔

```
        /\
       /  \
      / E2E \          端到端测试 (10%)
     /________\
    /          \
   / Integration \     集成测试 (30%)
  /______________\
 /                \
/    Unit Tests    \  单元测试 (60%)
/____________________\
```

### 测试类型定义

| 测试类型 | 定义 | 覆盖范围 | 工具 |
|----------|------|----------|------|
| **单元测试** | 测试单个函数/方法 | 所有业务逻辑 | Go testing, Jest |
| **集成测试** | 测试模块间交互 | API接口、数据库 | Go testing, Supertest |
| **E2E测试** | 测试完整用户流程 | 关键业务流程 | Cypress, Playwright |
| **性能测试** | 测试系统性能 | API响应时间 | k6, Apache Bench |
| **安全测试** | 测试安全漏洞 | 认证、输入验证 | gosec, npm audit |

---

## 📅 实施时间表

### Phase 1: 核心服务测试 (Week 1-4)

#### Week 1: optimization 模块 (P0)

**目标**: 达到 50% 覆盖率

| 文件 | 当前 | 目标 | 测试用例数 |
|------|------|------|-----------|
| `mpt_optimizer.go` | 0% | 60% | 15+ |
| `risk_parity.go` | 0% | 60% | 12+ |
| `black_litterman.go` | 0% | 40% | 10+ |

**测试重点**:
- 马科维茨优化算法
- 有效前沿计算
- 风险平价权重计算
- Black-Litterman 观点融合

**负责人**: Backend Team
**交付物**: `optimization/*_test.go`

#### Week 2: backtest 模块 (P0)

**目标**: 达到 50% 覆盖率

| 文件 | 当前 | 目标 | 测试用例数 |
|------|------|------|-----------|
| `event_engine.go` | 0% | 70% | 20+ |
| `engine.go` | 0% | 60% | 15+ |
| `strategies.go` | 0% | 50% | 10+ |

**测试重点**:
- 事件驱动架构
- 订单生命周期
- 滑点模型
- 策略执行

**负责人**: Backend Team
**交付物**: `backtest/*_test.go`

#### Week 3: handlers 提升 (P1)

**目标**: 从 37.1% 提升到 55%

| 处理器 | 当前 | 目标 | 测试用例数 |
|--------|------|------|-----------|
| `etf_handler.go` | 部分 | 70% | 20+ |
| `portfolio_handler.go` | 部分 | 70% | 15+ |
| `backtest_handler.go` | 0% | 60% | 12+ |
| `factor_handler.go` | 0% | 70% | 10+ |

**测试重点**:
- HTTP 请求处理
- 参数验证
- 错误处理
- 响应格式

**负责人**: Backend Team
**交付物**: `handlers/*_test.go`

#### Week 4: exchange_rate 模块 (P1)

**目标**: 从 27.3% 提升到 60%

| 文件 | 当前 | 目标 | 测试用例数 |
|------|------|------|-----------|
| `datasource/*.go` | 27.3% | 70% | 25+ |
| `sync/*.go` | 0% | 60% | 10+ |
| `monitor/*.go` | 0% | 50% | 8+ |

**测试重点**:
- 多数据源故障转移
- 健康检查
- 同步逻辑

**负责人**: Backend Team
**交付物**: `exchange_rate/**/*_test.go`

### Phase 2: 扩展服务测试 (Week 5-7)

#### Week 5: ashare 模块 (P1)

**目标**: 达到 50% 覆盖率

| 文件 | 当前 | 目标 | 测试用例数 |
|------|------|------|-----------|
| `akshare_provider.go` | 0% | 60% | 12+ |
| `tushare_provider.go` | 0% | 60% | 10+ |
| `etf_data_service.go` | 0% | 50% | 10+ |

**测试重点**:
- AKShare/TuShare 数据获取
- 数据转换
- 错误处理

**负责人**: Backend Team
**交付物**: `ashare/*_test.go`

#### Week 6: models 和 utils (P2)

**目标**:
- models: 56.7% → 70%
- utils: 81.2% → 90%

| 文件 | 当前 | 目标 | 测试用例数 |
|------|------|------|-----------|
| `models/*.go` | 56.7% | 70% | 15+ |
| `utils/*.go` | 81.2% | 90% | 10+ |

**测试重点**:
- 数据模型验证
- 工具函数边界情况
- 错误处理

**负责人**: Backend Team
**交付物**: `models/*_test.go`, `utils/*_test.go`

#### Week 7: middleware 完善 (P2)

**目标**: 68.8% → 80%

| 中间件 | 当前 | 目标 | 测试用例数 |
|--------|------|------|-----------|
| `auth.go` | 部分 | 90% | 15+ |
| `audit.go` | 部分 | 80% | 10+ |
| `validation.go` | 部分 | 85% | 12+ |

**测试重点**:
- JWT 验证逻辑
- 审计日志记录
- 输入验证规则

**负责人**: Backend Team
**交付物**: `middleware/*_test.go`

### Phase 3: 前端测试 (Week 8-9)

#### Week 8: 组件测试 (P1)

**目标**: 核心组件 60% 覆盖

| 组件 | 测试类型 | 测试用例数 |
|------|----------|-----------|
| `StatCard` | 单元测试 | 5+ |
| `PriceChart` | 单元测试 | 8+ |
| `RadarChart` | 单元测试 | 5+ |
| `Layout` | 单元测试 | 5+ |

**测试重点**:
- 组件渲染
- Props 传递
- 事件处理
- 状态管理

**负责人**: Frontend Team
**交付物**: `components/__tests__/*.test.tsx`

#### Week 9: 服务和工具测试 (P1)

**目标**:
- services: 70% 覆盖
- utils: 80% 覆盖

| 模块 | 测试类型 | 测试用例数 |
|------|----------|-----------|
| `api.ts` | 单元测试 | 15+ |
| `portfolio.ts` | 单元测试 | 10+ |
| `utils/*.ts` | 单元测试 | 15+ |

**测试重点**:
- API 调用
- 数据处理
- 错误处理

**负责人**: Frontend Team
**交付物**: `services/__tests__/*.test.ts`, `utils/__tests__/*.test.ts`

### Phase 4: 集成与 E2E 测试 (Week 10)

#### Week 10: 集成测试和 E2E 测试

**集成测试**:

| 场景 | 测试类型 | 测试用例数 |
|------|----------|-----------|
| ETF 数据流 | 集成测试 | 5+ |
| 组合优化流程 | 集成测试 | 5+ |
| 回测流程 | 集成测试 | 3+ |
| 用户认证流程 | 集成测试 | 5+ |

**E2E 测试**:

| 场景 | 测试工具 | 测试用例数 |
|------|----------|-----------|
| 用户登录 | Cypress | 3+ |
| ETF 分析 | Cypress | 5+ |
| 组合优化 | Cypress | 3+ |
| 回测执行 | Cypress | 3+ |

**负责人**: Full Stack Team
**交付物**:
- `tests/integration/*.test.go`
- `frontend/cypress/e2e/*.cy.ts`

---

## 🛠️ 测试工具与配置

### 后端测试工具

```go
// 测试依赖
go get github.com/stretchr/testify
go get github.com/stretchr/testify/mock

// 覆盖率工具
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

// 性能测试
go test -bench=. -benchmem
```

### 前端测试工具

```bash
# 测试框架
npm install --save-dev vitest
npm install --save-dev @testing-library/react
npm install --save-dev @testing-library/jest-dom

# E2E 测试
npm install --save-dev cypress
# 或
npm install --save-dev @playwright/test

# 覆盖率
npm run test:coverage
```

### CI/CD 集成

```yaml
# .github/workflows/ci.yml
- name: Run Backend Tests
  run: |
    cd backend
    go test -v -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

- name: Run Frontend Tests
  run: |
    cd frontend
    npm run test:coverage

- name: Coverage Threshold Check
  run: |
    # 检查覆盖率是否达到阈值
    ./scripts/check-coverage.sh
```

---

## 📋 测试规范

### 单元测试规范

```go
// 命名规范
func Test<FunctionName>_<Scenario>(t *testing.T)

// 示例
func TestCalculateSharpeRatio_NormalCase(t *testing.T)
func TestCalculateSharpeRatio_ZeroVolatility(t *testing.T)
func TestCalculateSharpeRatio_NegativeReturns(t *testing.T)

// 结构规范
tests := []struct {
    name     string
    input    InputType
    expected ExpectedType
    wantErr  bool
}{
    // 测试用例
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := Function(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            return
        }
        if !reflect.DeepEqual(result, tt.expected) {
            t.Errorf("got %v, want %v", result, tt.expected)
        }
    })
}
```

### 测试数据管理

```go
// testdata/fixtures.go
package testdata

// 生成测试用的ETF数据
func GenerateMockETFData(days int) []models.ETFData

// 生成测试用的价格序列
func GenerateMockPrices(count int) []decimal.Decimal

// 生成测试用的收益率序列
func GenerateMockReturns(count int) []float64
```

### Mock 和 Stub

```go
// 接口定义
type DataProvider interface {
    GetQuote(symbol string) (*Quote, error)
    GetHistoricalData(symbol string, days int) ([]DataPoint, error)
}

// Mock 实现
type MockDataProvider struct {
    mock.Mock
}

func (m *MockDataProvider) GetQuote(symbol string) (*Quote, error) {
    args := m.Called(symbol)
    return args.Get(0).(*Quote), args.Error(1)
}
```

---

## 👥 责任人与资源

### 团队分工

| 角色 | 职责 | 人员 |
|------|------|------|
| **测试负责人** | 整体测试策略、进度跟踪 | TBD |
| **Backend 测试** | 后端单元测试、集成测试 | Backend Team |
| **Frontend 测试** | 前端单元测试、组件测试 | Frontend Team |
| **E2E 测试** | 端到端测试、性能测试 | QA Team |
| **DevOps** | CI/CD 集成、测试环境 | DevOps Team |

### 资源需求

| 资源类型 | 需求 | 备注 |
|----------|------|------|
| **人力资源** | 2-3 人全职 | 10周周期 |
| **测试环境** | 独立测试数据库 | 避免污染生产数据 |
| **CI/CD 资源** | GitHub Actions | 已配置 |
| **测试数据** | 历史数据样本 | 用于集成测试 |

---

## 📈 进度跟踪

### 周报模板

```markdown
## Week X 测试覆盖报告

### 完成情况
- [ ] optimization 模块: X% → Y%
- [ ] backtest 模块: X% → Y%
- [ ] handlers: X% → Y%

### 新增测试
- 单元测试: X 个
- 集成测试: X 个
- E2E 测试: X 个

### 遇到的问题
1. 问题描述
2. 解决方案

### 下周计划
- 重点任务
- 目标覆盖率
```

### 里程碑检查点

| 检查点 | 日期 | 目标 | 检查项 |
|--------|------|------|--------|
| Phase 1 完成 | 2026-05-19 | 核心服务 50%+ | 代码审查、覆盖率报告 |
| Phase 2 完成 | 2026-06-09 | 扩展服务 60%+ | 集成测试通过 |
| Phase 3 完成 | 2026-06-23 | 前端 50%+ | 组件测试覆盖 |
| 最终验收 | 2026-06-30 | 整体 70%+ | 全量测试通过 |

---

## 🎯 成功标准

### 覆盖率目标

- [ ] 整体测试覆盖率 ≥ 70%
- [ ] 核心业务逻辑覆盖率 ≥ 90%
- [ ] 所有 P0/P1 模块覆盖率 ≥ 60%
- [ ] 关键算法覆盖率 = 100%

### 质量目标

- [ ] 所有测试通过 CI/CD
- [ ] 无 flaky tests
- [ ] 测试执行时间 < 5分钟
- [ ] 代码审查 100% 通过

### 文档目标

- [ ] 测试规范文档
- [ ] 测试用例文档
- [ ] 覆盖率报告自动化

---

## 🚨 风险管理

### 风险识别

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| 时间不足 | 中 | 高 | 优先级管理、分阶段交付 |
| 测试环境不稳定 | 低 | 中 | 独立测试环境、数据隔离 |
| 历史代码难以测试 | 高 | 中 | 渐进式重构、接口抽象 |
| 人员变动 | 低 | 高 | 知识文档化、代码审查 |

### 应急预案

1. **进度滞后**: 调整优先级，确保核心模块先完成
2. **测试失败**: 立即修复，不累积技术债务
3. **环境问题**: 快速切换到本地测试环境

---

## 📚 参考资源

### 测试最佳实践

- [Go Testing Best Practices](https://golang.org/doc/code.html#Testing)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Cypress Best Practices](https://docs.cypress.io/guides/references/best-practices)

### 内部文档

- [agents.md](../agents.md) - 开发规范
- [v2.5_phase1_implementation.md](./development/v2.5_phase1_implementation.md) - 测试示例
- [backend/README.md](../backend/README.md) - 后端测试说明

---

**制定**: ETF-Insight Team
**审批**: Tech Lead
**执行**: Development Team

**最后更新**: 2026-04-21
**下次评审**: 2026-04-28
