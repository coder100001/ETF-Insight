# Design Doc 004: 关键缺陷修复与代码质量提升

## 元数据
- **编号**: 004
- **标题**: 关键缺陷修复与代码质量提升
- **状态**: draft
- **创建日期**: 2026-05-02
- **最后更新**: 2026-05-02
- **关联任务**: 代码审查发现的关键缺陷修复
- **复杂度级别**: L2
- **涉及端**: 双端（后端为主，含前端路由页面）
- **前置 Design Doc**: 001, 002, 003

## 1. 背景与动机

经过 Superpowers + Gstack 双维度代码审查，发现在以下严重缺陷：

1. **BL 公式虚假实现**: `blFormula()` 使用占位符逻辑代替真实算法
2. **梯度下降非确定性**: map 迭代顺序随机导致优化结果不可复现
3. **服务路由断裂**: 因子择时/Alpha 观点/BL 服务已实现但未注册路由
4. **金融计算精度违反规范**: 汇率服务使用 float64 而非 decimal.Decimal

这些缺陷直接影响系统的正确性、可复现性和前端可用性。

## 2. 调研与现状分析

### 2.1 BL 公式问题

当前实现 ([alpha_view_service.go:L324-L352](file:///Users/liunian/Desktop/dnmp/py_project/backend/services/alpha_view_service.go#L324-L352))：

```go
// 实际代码：占位符实现
for i := 0; i < n; i++ {
    posteriorReturns[i] = pi[i].Add(Q[0].Mul(decimal.NewFromFloat(0.1)))
}
```

标准 BL 后验收益公式：
```
Π_bl = [(τΣ)⁻¹ + P'Ω⁻¹P]⁻¹ × [(τΣ)⁻¹Π + P'Ω⁻¹Q]
```

项目中已实现矩阵运算函数（`calculateMatrixInverse`, `calculateMatrixMultiply`, `calculateMatrixTranspose`），但 BL 公式未调用它们。

### 2.2 梯度下降 map 迭代问题

[portfolio_optimizer.go:L333-L334](file:///Users/liunian/Desktop/dnmp/py_project/backend/services/portfolio_optimizer.go#L333-L334) 中每次调用 `calculateNegativeSharpeGradients` 时都重新从 map 构建 symbols 列表，顺序不可控。

### 2.3 路由注册缺失

`router.go` 中缺少以下路由注册：
- `POST /api/factor/timing/calculate`
- `GET /api/factor/timing/history`
- `POST /api/alpha-views`
- `GET /api/alpha-views/active`
- `POST /api/black-litterman/configs`
- `POST /api/black-litterman/calculate`

### 2.4 汇率精度问题

[exchange_rate.go:L39](file:///Users/liunian/Desktop/dnmp/py_project/backend/services/exchange_rate.go#L39) 中 `GetRate` 返回 float64，违反项目 Decimal 强制规范。

## 3. 可选方案

### 方案 A: 分批次修复（推荐）

按优先级分 P0/P1/P2 三个批次：
- **P0**: BL 公式、map 顺序（阻断性缺陷）
- **P1**: 路由注册、汇率类型、健康检查、测试
- **P2**: 风险预算服务、CORS、前端页面

优点：风险可控，每批可独立验证
缺点：周期较长

### 方案 B: 一次性全量修复

一次性修复所有问题。

优点：一步到位
缺点：变更面大，回滚困难，不易验证

## 4. 决策

**选定方案**: A（分批次修复）

决策理由：BL 公式和 map 顺序是阻断性缺陷，必须优先修复。路由注册和类型修正是 P1 级，可随后进行。P2 级功能需要较多工作量，可独立规划。

## 5. 后端修复设计

### 5.1 BL 公式重写

使用已有的矩阵运算函数实现完整 BL 公式：

```go
func (s *BlackLittermanService) blFormula(pi, cov, P, Q, Omega, tau) {
    // Step 1: τΣ
    tauCov := scaleMatrix(cov, tau)

    // Step 2: (τΣ)⁻¹
    tauCovInv := calculateMatrixInverse(tauCov)

    // Step 3: P'Ω⁻¹P
    omegaInv := calculateMatrixInverse(Omega)
    pT := calculateMatrixTranspose(P)
    pTOmegaInvP := calculateMatrixMultiply(calculateMatrixMultiply(pT, omegaInv), P)

    // Step 4: [(τΣ)⁻¹ + P'Ω⁻¹P]⁻¹
    sum := addMatrices(tauCovInv, pTOmegaInvP)
    sumInv := calculateMatrixInverse(sum)

    // Step 5: (τΣ)⁻¹Π + P'Ω⁻¹Q
    tauCovInvPi := matrixVectorMultiply(tauCovInv, pi)
    pTOmegaInvQ := matrixVectorMultiply(calculateMatrixMultiply(pT, omegaInv), Q)
    rhs := addVectors(tauCovInvPi, pTOmegaInvQ)

    // Step 6: posterior = sumInv × rhs
    posteriorReturns := matrixVectorMultiply(sumInv, rhs)
}
```

### 5.2 Map 顺序固定

在优化开始时对 symbols 排序，确保每次运行结果一致：

```go
symbols := make([]string, 0, n)
for symbol := range meanReturns {
    symbols = append(symbols, symbol)
}
sort.Strings(symbols) // 固定顺序
```

### 5.3 路由注册

新增 `registerAlphaRoutes()`, `registerBlackLittermanRoutes()`, `registerRiskBudgetRoutes()` 方法。

## 6. 前端修复设计

### 6.1 v2.7 页面实现

四个页面需要从 TODO 状态补全：
- `FactorTiming.tsx` - 因子择时信号分析
- `AlphaViews.tsx` - Alpha 观点管理
- `BlackLittermanConfig.tsx` - BL 模型配置
- `RiskBudget.tsx` - 风险预算管理

## 7. 影响范围

| 模块 | 变更 |
|------|------|
| `services/alpha_view_service.go` | BL 公式重写 |
| `services/portfolio_optimizer.go` | symbols 排序 |
| `router/router.go` | 新增路由注册 |
| `services/exchange_rate.go` | float64 → decimal |
| `core/app.go` | 健康检查适配 HTTPS |
| `services/ashare/akshare_server.py` | CORS 限制 |
| `frontend/src/pages/` | 4 个页面实现 |

## 8. 开放问题

- [ ] BL 公式重写后需要和现有 MPT 优化、风险平价集成测试
- [ ] 矩阵运算是否需要在数值稳定性上做额外处理（Tikhonov 正则化）？

---
**文档状态**: draft
