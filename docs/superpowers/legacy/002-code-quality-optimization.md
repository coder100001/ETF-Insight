# Design Doc 002: 代码质量优化与修复

## 元数据
- **编号**: 002
- **标题**: 代码质量优化与修复
- **状态**: draft
- **创建日期**: 2026-04-28
- **最后更新**: 2026-04-28
- **关联任务**: 基于质量报告的代码优化
- **复杂度级别**: L3
- **涉及端**: 双端（后端 + 前端）
- **前置 Design Doc**: 001-code-quality-analysis-and-improvement.md

## 1. 背景与动机

### 为什么需要这个改动？
根据代码质量分析报告，项目存在以下关键问题：
- **P0**: 优化算法零测试、前端 `any` 类型
- **P1**: 测试覆盖率不达标、前端组件过长、可访问性不足
- **P2**: 前端测试薄弱

### 当前系统的痛点
- 核心优化算法（MPT、风险平价、Black-Litterman）无单元测试，质量无保障
- 前端 TypeScript 类型不严格，存在 `any` 类型
- 测试覆盖率仅 35-40%，远低于 80% 目标
- 前端组件过长，可维护性差
- 可访问性评分低，影响无障碍使用

### 业务/技术驱动因素
- 提升代码质量和可维护性
- 保障核心算法正确性
- 提升前端类型安全
- 改善用户体验（可访问性）

## 2. 调研与现状分析

### 2.1 现有实现
- 后端：`services/optimization/` 目录下有 MPT、风险平价、Black-Litterman 优化器
- 前端：`api.ts`、`PortfolioOptimization.tsx` 等文件使用 `any` 类型
- 测试：`services/technical_indicators_test.go` 等有良好测试示例

### 2.2 业界实践
- Go 测试：使用 `testing` 包，表驱动测试
- TypeScript：严格模式，禁用 `any`
- React：组件拆分，单一职责

### 2.3 技术约束
- 不修改核心业务逻辑
- 保持向后兼容
- 测试必须覆盖边界条件

## 3. 可选方案

### 方案 A: 分阶段修复（推荐）
- **描述**: 按 P0 → P1 → P2 优先级分阶段修复
- **优点**: 风险可控，可逐步验证
- **缺点**: 周期较长
- **工作量**: 大

### 方案 B: 并行修复
- **描述**: 同时处理所有问题
- **优点**: 快速完成
- **缺点**: 风险高，可能引入新问题
- **工作量**: 大

### 方案 C: 仅修复 P0
- **描述**: 只处理最高优先级问题
- **优点**: 快速见效
- **缺点**: 技术债遗留
- **工作量**: 中

## 4. 决策
- **选定方案**: 方案 A（分阶段修复）
- **决策理由**: 风险可控，质量有保障
- **权衡取舍**: 时间换质量

## 5. 后端设计

### 新增测试文件
```
backend/services/optimization/
├── mpt_optimizer_test.go      # MPT 优化器测试
├── risk_parity_test.go        # 风险平价测试
└── black_litterman_test.go    # Black-Litterman 测试
```

### 测试用例设计
```go
func TestMPTOptimizer(t *testing.T) {
    tests := []struct {
        name     string
        input    OptimizationInput
        expected OptimizationResult
        wantErr  bool
    }{
        {"正常情况", normalInput, normalResult, false},
        {"空输入", emptyInput, OptimizationResult{}, true},
        {"边界条件", boundaryInput, boundaryResult, false},
    }
    // ...
}
```

## 6. 前端设计

### 类型修复
```typescript
// 修复前
const data: any = response.data;

// 修复后
interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}
const data: ApiResponse<OptimizationResult> = response;
```

### 组件拆分
```
PortfolioOptimization.tsx (拆分前: 500+ 行)
├── PortfolioOptimization.tsx (主组件: ~150 行)
├── OptimizationForm.tsx (表单组件: ~100 行)
├── OptimizationResult.tsx (结果组件: ~100 行)
└── types.ts (类型定义: ~50 行)
```

### 可访问性改进
```tsx
// 修复前
<div onClick={handleClick}>按钮</div>

// 修复后
<button onClick={handleClick} aria-label="优化按钮">按钮</button>
```

## 7. 前后端交互
无 API 变更，仅内部优化。

## 8. 影响范围
- 影响的模块: `services/optimization/`, `frontend/src/`
- 影响的接口: 无
- 影响的配置: 无
- 影响的部署: 无

## 9. 开放问题
- [ ] 是否需要集成测试？
- [ ] 是否需要性能基准测试？

---
**文档状态**: draft
