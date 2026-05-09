---
name: "gs3-hybrid"
description: "ETF-Insight 项目专用的 Superpowers + GStack 混合流程。Invoke when user starts new feature development, code review, or any engineering task requiring structured workflow with quality gates."
---

# GS3-Hybrid: ETF-Insight 项目专用流程

> 结合 Superpowers 逻辑规划与 gstack 工程标准的混合流程
> 针对前后端分离架构（Go + React/TypeScript）定制
> 版本: v2.0 | 基于 gs-hybrid-v3.1 优化

---

## 快速开始

### 启动指令

| 指令 | 流程 | 说明 |
|------|------|------|
| `/plan` 或 `gs3 plan` | 规划流程 | 新功能开发前的完整规划 |
| `/review` 或 `gs3 review` | 代码审查 | 代码完成后的质量审查 |
| `/test` 或 `gs3 test` | 测试驱动 | TDD 开发流程 |
| `/ship` 或 `gs3 ship` | 发布准备 | 提交前的最终检查 |
| `/qa` 或 `gs3 qa` | 质量保证 | 功能完成后的验证 |
| `/debug` 或 `gs3 debug` | 调试助手 | 问题诊断与修复 |
| `/refactor` 或 `gs3 refactor` | 重构建议 | 代码改进与优化 |

### 标准工作流

```
用户: "gs3 帮我开发新功能"

AI: 收到。我将按照 GS3-Hybrid 流程执行：

Step 0:   评估任务复杂度 (L1/L2/L3)
Step 0.5: L2+ 创建 Design Doc (编号设计文档)
Step 1:   生成详细 PLAN.md (前后端分离规划)
Step 2:   工程规范设计 (双端规范检查)
Step 2.5: 前端专项检验 (代码/性能/安全/A11y) [涉及前端时]
Step 3:   架构师评审 (SOLID 原则)
Step 4:   QA 评审 (测试策略)
Step 5:   CSO 安全评审 (安全扫描)
Step 6:   TDD 编码实现
Step 7:   验证交付
```

---

## 项目配置

### ETF-Insight 专用配置

```yaml
# ============================================================
# 后端 Go 配置
# ============================================================
backend:
  language: "Go 1.26"
  test_command: "go test ./... -race"
  coverage_command: "go test ./... -coverprofile=coverage.out"
  bench_command: "go test -bench=. -benchmem ./..."
  lint_command: "gofmt -l . && go vet ./..."
  security_scanner: "gosec ./..."
  build_command: "go build ./..."
  code_review_guide: "https://github.com/golang/go/wiki/CodeReviewComments"
  concurrency_model: "goroutine"
  injection_patterns:
    - "fmt\\.Sprintf.*%s.*query"
    - "exec\\.Command.*input"
  secret_patterns:
    - "password\\s*=\\s*[\"'][^\"']+[\"']"
    - "api_key\\s*=\\s*[\"'][^\"']+[\"']"

# ============================================================
# 前端 React/TypeScript 配置
# ============================================================
frontend:
  language: "TypeScript 5.9 + React 19"
  test_command: "npm run test"
  coverage_command: "npm run test:coverage"
  lint_command: "npm run lint"
  typecheck_command: "npx tsc --noEmit"
  build_command: "npm run build"
  code_review_guide: "https://github.com/airbnb/javascript"
  concurrency_model: "async/await"

# ============================================================
# 项目结构
# ============================================================
project:
  name: "ETF-Insight"
  type: "fullstack"
  backend_path: "./backend"
  frontend_path: "./frontend"
  docs_path: "./docs"
  design_docs_path: "./design-docs"
```

---

## 核心概念

| 概念 | 定义 |
|-----|------|
| **Design Doc** | 编号设计文档，存储在 `docs/superpowers/specs/` 目录，记录重大功能/改动的调研、设计决策和演进理由 |
| **变更文件数** | 新增、修改、删除的文件总数（不含自动生成文件） |
| **新增代码行** | 不含空行和注释的净增代码行数 |
| **架构影响** | 涉及模块间接口、数据流、部署方式的变更 |
| **风险等级** | 低：仅影响非核心功能；中：影响核心功能但可回滚；高：涉及数据迁移或不可逆变更 |
| **前后端联动** | 单端（仅后端或仅前端）/ 双端（前后端都需要修改） |

---

## 任务复杂度分级

### 判定标准

| 维度 | L1 简单 | L2 中等 | L3 复杂 |
|------|---------|---------|---------|
| 变更文件数 | < 3 | 3-8 | > 8 |
| 新增代码行 | < 100 | 100-500 | > 500 |
| 接口变更 | 无 | 新增 | 修改/删除 |
| 架构影响 | 无 | 局部 | 全局 |
| 依赖变更 | 无 | 新增依赖 | 替换核心依赖 |
| 前后端联动 | 单端 | 单端 | 双端 |
| 风险等级 | 低 | 中 | 高 |

### 复杂度判定决策树

```
开始
  │
  ├─ 是否涉及架构重构？
  │   ├─ 是 → L3
  │   └─ 否 ↓
  │
  ├─ 是否涉及安全或性能关键路径？
  │   ├─ 是 → L3
  │   └─ 否 ↓
  │
  ├─ 修改文件数 > 8？
  │   ├─ 是 → L3
  │   └─ 否 ↓
  │
  ├─ 新增代码 > 500 行？
  │   ├─ 是 → L3
  │   └─ 否 ↓
  │
  ├─ 修改文件数 >= 3？
  │   ├─ 是 ↓
  │   │   ├─ 新增接口或模块？ → L2
  │   │   └─ 影响现有功能？ → L2
  │   └─ 否 ↓
  │
  └─ L1 (简单任务)
```

### 流程适用矩阵

| Phase | 名称 | L1 | L2 | L3 | 说明 |
|:-----:|------|:--:|:--:|:--:|------|
| 0 | 复杂度评估 | ✅ | ✅ | ✅ | 所有任务必须 |
| 0.5 | Design Doc | ⚪ | 🔴 | 🔴 | L2+ 必须 |
| 1 | 逻辑规划 | ✅(简化) | ✅ | ✅ | 前后端分离规划 |
| 2 | 工程规范设计 | ⚪ | 🟡 | 🔴 | 双端规范检查 |
| 2.5 | 前端专项检验 | ⚪ | 🟡 | 🔴 | 前端代码/性能/安全/A11y |
| 3 | 架构师评审 | ⚪ | 🟡 | 🔴 | SOLID 原则 |
| 4 | QA 评审 | ⚪ | ⚪ | 🔴 | 测试策略 |
| 5 | CSO 安全评审 | ⚪ | ⚪ | 🔴 | 安全扫描 |
| 6 | 编码实现 | ✅ | ✅ | ✅ | TDD 流程 |
| 7 | 验证交付 | ✅(简化) | ✅ | ✅ | 质量门禁 |

> 图例: ✅ 必须 | 🟡 L2+ 必须 | 🔴 L3 必须 | ⚪ 可选

---

## Step 0: 复杂度评估

**触发**: 任何任务开始前

### 评估清单

```markdown
## 复杂度评估报告

### 变更统计
- 新增文件: __ 个
- 修改文件: __ 个
- 删除文件: __ 个
- 预估后端代码行: __ 行
- 预估前端代码行: __ 行

### 影响分析
- [ ] 影响后端 API 接口
- [ ] 影响前端页面/组件
- [ ] 影响数据库模型
- [ ] 需要前后端联调
- [ ] 涉及第三方依赖变更
- [ ] 影响配置文件
- [ ] 影响部署流程

### 评估结论
**复杂度级别**: L1 / L2 / L3
**涉及端**: 后端 / 前端 / 双端
**流程选择**: 简化流程 / 标准流程 / 完整流程
**预计耗时**: __ 小时
```

---

## Phase 0.5: Design Doc (研究 & 设计)

**触发**: Step 0 评估为 L2/L3 | **适用级别**: 🔴 L2+ 必须，⚪ L1 可选

### Design Doc 价值

| 价值 | 说明 |
|-----|------|
| **思考留痕** | 记录调研过程、方案对比、决策理由 |
| **可追溯性** | 编号递增 + Git 提交，可回溯设计决策 |
| **知识传承** | 新成员通过 docs/superpowers/specs/ 快速理解项目演进 |
| **减少返工** | 先思考后编码，避免方向性错误 |
| **评审依据** | Phase 2-5 评审的基础材料 |

### 文件命名规范

```
docs/superpowers/specs/
├── 001-feature-name.md
├── 002-another-feature.md
├── README.md          # 索引文件
└── archive/           # 归档文档
```

**命名规则**: `YYYY-MM-DD-kebab-case-title.md`
- `NNN`: 三位数字，从 001 开始递增
- `kebab-case-title`: 简短英文描述
- 编号必须连续，不允许跳号

### Design Doc 模板

```markdown
# Design Doc NNN: [标题]

## 元数据
- **编号**: NNN
- **标题**: [简短描述]
- **状态**: draft / approved / superseded / deprecated
- **创建日期**: YYYY-MM-DD
- **最后更新**: YYYY-MM-DD
- **关联任务**: [任务描述或 issue 编号]
- **复杂度级别**: L2 / L3
- **涉及端**: 后端 / 前端 / 双端
- **前置 Design Doc**: [如有关联，填写编号]

## 1. 背景与动机
- 为什么需要这个改动？
- 当前系统的痛点是什么？
- 业务/技术驱动因素

## 2. 调研与现状分析

### 2.1 现有实现
- 当前架构/代码如何工作
- 已有的限制和瓶颈

### 2.2 业界实践
- 同类系统的解决方案
- 开源参考实现
- 相关论文/文章

### 2.3 技术约束
- 语言/框架限制
- 兼容性要求
- 性能约束

## 3. 可选方案

### 方案 A: [名称]
- **描述**:
- **优点**:
- **缺点**:
- **工作量**: 小/中/大

### 方案 B: [名称]
- **描述**:
- **优点**:
- **缺点**:
- **工作量**: 小/中/大

### 方案 C: [名称] (如有)
- **描述**:
- **优点**:
- **缺点**:
- **工作量**: 小/中/大

## 4. 决策
- **选定方案**: [A/B/C]
- **决策理由**:
- **权衡取舍**:

## 5. 后端设计
### API 接口
```go
// 接口定义
```

### 数据模型
```go
// 模型定义
```

## 6. 前端设计
### 组件结构
```
Component/
├── index.tsx
├── types.ts
└── utils.ts
```

### API 调用
```typescript
// API 封装
```

## 7. 前后端交互
### 请求/响应格式
```json
{
  "success": true,
  "data": {},
  "message": ""
}
```

## 8. 影响范围
- 影响的模块/包
- 影响的接口
- 影响的配置
- 影响的部署

## 9. 开放问题
- [ ] 待解决问题 1
- [ ] 待解决问题 2

---
**文档状态**: draft
```

### README.md 索引模板

```markdown
# Design Docs

> 本目录记录项目的重大设计决策，类似 DB Migrations。

## 文档索引

| 编号 | 标题 | 状态 | 创建日期 | 涉及端 | 关联 |
|-----|------|------|---------|--------|------|
| 001 | [标题](001-title.md) | approved | YYYY-MM-DD | 后端 | - |
| 002 | [标题](002-title.md) | draft | YYYY-MM-DD | 双端 | 001 |

## 状态说明

| 状态 | 含义 |
|-----|------|
| draft | 草稿，尚未评审 |
| approved | 已通过评审，可进入实现 |
| superseded | 被后续文档取代 |
| deprecated | 已废弃，不再适用 |
```

### 强制检查清单
- [ ] 是否创建了 docs/superpowers/specs/ 目录？
- [ ] 是否确定了正确的编号 (NNN)？
- [ ] 是否使用了标准模板？
- [ ] 是否包含至少 2 个可选方案的对比？
- [ ] 是否记录了决策理由？
- [ ] 是否更新了 README.md 索引？
- [ ] 是否提交到了仓库？

**⚠️ 阻断规则**: L2+ 任务未完成 Design Doc，禁止进入 Phase 1

---

## Phase 1: 逻辑规划

**触发**: 任何任务（L2+ 需先完成 Phase 0.5）

**强制输出**: `PLAN.md`

### 后端规划模板

```markdown
# 后端实现计划

## 1. 需求概述
- **功能名称**:
- **API 路径**: `METHOD /api/xxx`
- **复杂度**: L1/L2/L3
- **Design Doc**: [YYYY-MM-DD-title](../docs/superpowers/specs/YYYY-MM-DD-title.md) (L2+ 必填)

## 2. 变更文件

### 新增文件
| 文件路径 | 用途 | 行数预估 |
|---------|------|---------|
| `services/xxx_service.go` | 业务逻辑 | ~200 |
| `services/xxx_service_test.go` | 单元测试 | ~150 |
| `handlers/xxx_handler.go` | HTTP 处理 | ~100 |
| `models/xxx.go` | 数据模型 | ~50 |

### 修改文件
| 文件路径 | 变更类型 | 影响范围 |
|---------|---------|---------|
| `router/router.go` | 添加路由 | 新增接口 |
| `models/db.go` | 注册模型 | 自动迁移 |

### 删除文件
| 文件路径 | 原因 |
|---------|------|
| 无 | - |

## 3. 接口设计
```go
type XXXRequest struct {
    Field string `json:"field" binding:"required"`
}

type XXXResponse struct {
    Data interface{} `json:"data"`
}
```

## 4. 核心逻辑
```
[流程图或伪代码]
```

## 5. 边界条件
| 场景 | 输入 | 预期输出 | 处理方式 |
|-----|------|---------|---------|
| 正常情况 | ... | ... | ... |
| 空输入 | ... | ... | 返回 400 错误 |
| 超大输入 | ... | ... | 分页处理 |
| 数据库错误 | ... | ... | 返回 500 错误 |
| 并发访问 | ... | ... | 加锁处理 |

## 6. 测试策略
- [ ] 单元测试覆盖率 >= 80%
- [ ] 边界条件测试
- [ ] 错误处理测试
- [ ] 并发测试（如需要）

## 7. 依赖分析
### 内部依赖
- `services/xxx` - 用途
- `models/xxx` - 用途

### 外部依赖
- 无新增 / `library/xxx` - 用途

### 循环依赖风险
- [ ] 无风险
- [ ] 有风险，解决方案：...

## 8. 风险评估
| 风险类型 | 概率 | 影响 | 缓解措施 |
|---------|------|------|---------|
| 性能瓶颈 | 中 | 高 | 提前基准测试 |
| 向后不兼容 | 低 | 高 | 保留旧接口 |

## 9. 验收标准
- [ ] 功能实现完整
- [ ] 测试覆盖率达标
- [ ] 性能指标满足
- [ ] 文档更新完成

## 10. 回滚策略
- [ ] 数据库迁移可回滚
- [ ] 配置变更可回滚
- [ ] 功能开关控制

---
**计划状态**: 待评审
**创建时间**: YYYY-MM-DD
**最后更新**: YYYY-MM-DD
```

### 前端规划模板

```markdown
# 前端实现计划

## 1. 需求概述
- **功能名称**:
- **页面路由**: `/xxx`
- **复杂度**: L1/L2/L3
- **Design Doc**: [YYYY-MM-DD-title](../docs/superpowers/specs/YYYY-MM-DD-title.md) (L2+ 必填)

## 2. 变更文件

### 新增文件
| 文件路径 | 用途 | 行数预估 |
|---------|------|---------|
| `src/pages/XXXPage.tsx` | 页面组件 | ~200 |
| `src/components/XXXComponent.tsx` | 子组件 | ~150 |
| `src/hooks/useXXX.ts` | 自定义 Hook | ~100 |
| `src/services/xxxAPI.ts` | API 封装 | ~80 |
| `src/types/xxx.ts` | 类型定义 | ~50 |

### 修改文件
| 文件路径 | 变更类型 |
|---------|---------|
| `src/App.tsx` | 添加路由 |
| `src/services/api.ts` | 添加 API |

## 3. 组件设计
```typescript
// 组件接口
interface XXXProps {
  data: XXXData;
  onSubmit: (values: XXXForm) => void;
}
```

## 4. 状态管理
- **本地状态**: useState
- **服务端状态**: React Query / SWR
- **全局状态**: Context / Redux (如需)

## 5. 边界条件
| 场景 | 处理方式 |
|-----|---------|
| 加载中 | Loading 组件 |
| 空数据 | Empty 状态 |
| 错误 | Error Boundary |
| 网络失败 | 重试机制 |

## 6. 测试策略
- [ ] 组件渲染测试
- [ ] 用户交互测试
- [ ] API Mock 测试

## 7. 性能考虑
- [ ] 组件懒加载
- [ ] 数据缓存
- [ ] 虚拟列表（大数据）

## 8. 风险评估
| 风险 | 概率 | 影响 | 缓解措施 |
|-----|------|------|---------|
| 浏览器兼容 | 低 | 中 | 提前测试 |

---
**计划状态**: 待评审
```

### 强制检查清单
- [ ] 是否生成了 PLAN.md?
- [ ] 是否列出了所有变更文件?
- [ ] 是否识别了边界条件?
- [ ] 是否评估了风险?
- [ ] 是否制定了回滚策略?
- [ ] L2+ 是否引用了 Design Doc 编号?
- [ ] 是否等待用户确认?

**⚠️ 阻断规则**: 未完成以上检查，禁止进入 Phase 2

---

## Phase 2: 工程规范设计（L2+）

### 后端规范检查

```markdown
## 后端工程规范检查

### 架构设计
- [ ] 符合三层架构（Handler-Service-Model）
- [ ] 无循环依赖
- [ ] 接口粒度适中（方法数 < 10）
- [ ] 依赖注入支持

### 代码规范
- [ ] 遵循 Go 官方代码规范
- [ ] 使用 `gofmt` 格式化
- [ ] 函数长度 < 50 行
- [ ] 圈复杂度 < 10
- [ ] 嵌套深度 < 3
- [ ] 命名清晰（动词+名词）
- [ ] 注释解释"为什么"而非"做什么"

### 金融计算规范
- [ ] 使用 `decimal.Decimal` 处理金额
- [ ] 收益率统一使用百分比
- [ ] 波动率使用年化值
- [ ] 除零保护

### 错误处理
- [ ] 所有错误必须处理
- [ ] 使用 `fmt.Errorf("context: %w", err)` 包装
- [ ] 不忽略错误（避免 `_ = xxx`）
- [ ] 返回有意义的错误信息

### 并发安全
- [ ] 共享数据加锁
- [ ] Goroutine 有退出机制
- [ ] 使用 `context` 控制超时
- [ ] 无 Goroutine 泄漏

### 性能评估
| 操作 | 复杂度 | 是否可接受 |
|-----|--------|:---------:|
| 读取 | O(n) | ✅ |
| 写入 | O(log n) | ✅ |
```

### 前端规范检查

```markdown
## 前端工程规范检查

### TypeScript 规范
- [ ] 严格类型检查（禁用 `any`）
- [ ] 接口定义完整
- [ ] 枚举使用常量
- [ ] 泛型合理使用

### React 规范
- [ ] 函数式组件
- [ ] Hooks 规范使用
- [ ] 组件粒度适中（< 300 行）
- [ ] Props 类型定义
- [ ] 依赖数组完整

### 代码组织
- [ ] 按功能组织目录
- [ ] 公共逻辑抽离 Hooks
- [ ] API 统一封装
- [ ] 类型集中管理

### 性能优化
- [ ] 使用 `React.memo`（如需要）
- [ ] 使用 `useMemo`/`useCallback`（如需要）
- [ ] 图片懒加载
- [ ] 代码分割

### 可访问性
- [ ] 语义化 HTML
- [ ] ARIA 标签
- [ ] 键盘导航
```

---

## Phase 2.5: 前端专项检验（L2+）

**触发**: Phase 2 通过后，涉及前端变更的任务 | **适用级别**: 🟡 L2+ 必须，⚪ L1 可选

> 专门针对前端代码的深度检验，包括代码质量、性能、安全性和可访问性

### 2.5.1 前端代码质量检验

```markdown
## 前端代码质量检验

### TypeScript 严格性检查
- [ ] **无 `any` 类型**: 所有变量/参数必须有明确类型
- [ ] **无 `@ts-ignore`**: 除非有详细注释说明原因
- [ ] **无 `@ts-expect-error`**: 临时禁用需标记 TODO
- [ ] **严格 null 检查**: 启用 `strictNullChecks`
- [ ] **严格模式**: 启用 `strict: true`

### 代码复杂度检查
| 指标 | 阈值 | 检查工具 |
|-----|------|---------|
| 组件行数 | < 300 行 | 人工检查 |
| 函数行数 | < 50 行 | 人工检查 |
| 圈复杂度 | < 10 | eslint-plugin-complexity |
| 嵌套深度 | < 4 层 | 人工检查 |
| Props 数量 | <= 7 个 | 人工检查 |

### 代码规范检查
- [ ] **命名规范**:
  - 组件: PascalCase (e.g., `UserProfile`)
  - Hook: camelCase 前缀 `use` (e.g., `useAuth`)
  - 工具函数: camelCase (e.g., `formatDate`)
  - 常量: SCREAMING_SNAKE_CASE
- [ ] **文件组织**: 每个文件一个主要导出
- [ ] **导入排序**: 第三方库 → 内部模块 → 相对路径
- [ ] **无未使用变量/导入**: ESLint `no-unused-vars`
- [ ] **无 console.log**: 生产代码移除或改用 logger

### 注释规范
- [ ] 复杂逻辑必须注释
- [ ] 公共 API 必须有 JSDoc
- [ ] 无注释掉的代码块
- [ ] TODO/FIXME 必须有 Issue 链接
```

### 2.5.2 前端性能检验

```markdown
## 前端性能检验

### 构建性能
| 指标 | 目标值 | 检查命令 |
|-----|--------|---------|
| 构建时间 | < 30s | `npm run build` |
| 包体积 | < 500KB (gzip) | `npm run analyze` |
| Chunk 数量 | < 20 个 | 构建输出 |

### 运行时性能
| 指标 | 目标值 | 检查工具 |
|-----|--------|---------|
| 首屏加载 (FCP) | < 1.8s | Lighthouse |
| 可交互时间 (TTI) | < 3.8s | Lighthouse |
| 累积布局偏移 (CLS) | < 0.1 | Lighthouse |
| 最大内容绘制 (LCP) | < 2.5s | Lighthouse |
| 总阻塞时间 (TBT) | < 200ms | Lighthouse |

### 代码性能检查
- [ ] **无内存泄漏**:
  - 组件卸载时清理定时器/订阅
  - 事件监听正确移除
  - 闭包不持有大对象引用
- [ ] **渲染优化**:
  - 大数据列表使用虚拟滚动
  - 图表使用 Canvas/WebGL (数据量大时)
  - 避免不必要的重渲染
- [ ] **网络优化**:
  - API 请求防抖/节流
  - 图片懒加载
  - 资源预加载关键资源
- [ ] **状态优化**:
  - 状态粒度适中
  - 避免深层嵌套状态
  - 使用 selector 减少重渲染

### 性能检测命令
```bash
# Lighthouse 检测
cd frontend && npx lighthouse http://localhost:5173 --preset=desktop

# 包体积分析
cd frontend && npm run build && npx vite-bundle-visualizer

# 性能分析（开发模式）
cd frontend && npm run dev
# Chrome DevTools → Performance → Record
```
```

### 2.5.3 前端安全检验

```markdown
## 前端安全检验

### XSS 防护检查
- [ ] **无 `dangerouslySetInnerHTML`**: 除非内容已净化
- [ ] **无 `innerHTML` 直接赋值**: 使用 textContent 替代
- [ ] **无 `eval()`**: 使用 JSON.parse 或 Function 构造器
- [ ] **无 `new Function()`**: 动态代码执行风险
- [ ] **URL 校验**: 跳转前校验 URL 协议
- [ ] **富文本净化**: 使用 DOMPurify 净化用户输入

### CSRF 防护检查
- [ ] **请求携带 Token**: 所有修改请求带 CSRF Token
- [ ] **SameSite Cookie**: Cookie 设置 SameSite 属性
- [ ] **验证 Origin/Referer**: 敏感操作验证来源

### 敏感数据处理
- [ ] **无 LocalStorage 存敏感信息**: token 存 httpOnly cookie
- [ ] **无 URL 传敏感参数**: 使用 POST body 或 Header
- [ ] **日志脱敏**: 错误日志不打敏感字段
- [ ] **表单自动完成**: 敏感字段关闭 autocomplete

### 依赖安全检查
```bash
# 检查已知漏洞
cd frontend && npm audit

# 检查过时依赖
cd frontend && npm outdated

# 许可证检查
cd frontend && npx license-checker --onlyAllow 'MIT;Apache-2.0;BSD-3-Clause'
```

### 安全扫描规则
```yaml
frontend_security_rules:
  xss_risk:
    - pattern: "dangerouslySetInnerHTML"
      severity: high
      message: "必须使用 DOMPurify 净化内容"
    - pattern: "innerHTML\\s*="
      severity: high
      message: "使用 textContent 或 React 的 JSX 替代"
    - pattern: "eval\\("
      severity: critical
      message: "禁止使用 eval"

  sensitive_storage:
    - pattern: "localStorage\\.setItem.*token"
      severity: high
      message: "Token 应存储在 httpOnly cookie"
    - pattern: "localStorage\\.setItem.*password"
      severity: critical
      message: "禁止在 LocalStorage 存储密码"

  insecure_api:
    - pattern: "axios\\.get.*params.*password"
      severity: high
      message: "敏感信息不应通过 URL 参数传递"
```
```

### 2.5.4 前端可访问性检验 (A11y)

```markdown
## 前端可访问性检验 (A11y)

### 语义化 HTML
- [ ] **正确使用标题层级**: h1 → h2 → h3，不跳过
- [ ] **表单关联标签**: `<label for="id">` 或包裹 input
- [ ] **按钮可识别**: 使用 `<button>` 而非 `<div onClick>`
- [ ] **链接可识别**: 使用 `<a href>` 而非 `<div onClick>`
- [ ] **列表结构**: 使用 `<ul>/<ol>/<li>` 而非 div 模拟
- [ ] **表格结构**: 使用 `<table>` 相关标签，含 `<th>`

### ARIA 标签
- [ ] **必要 ARIA 属性**:
  - 自定义控件: `role`, `aria-label`, `aria-describedby`
  - 动态内容: `aria-live`, `aria-atomic`
  - 状态指示: `aria-expanded`, `aria-selected`, `aria-hidden`
- [ ] **无冗余 ARIA**: 原生 HTML 语义足够时不加 ARIA
- [ ] **ARIA 状态同步**: 视觉状态与 ARIA 状态一致

### 键盘导航
- [ ] **Tab 顺序合理**: 符合视觉顺序
- [ ] **焦点可见**: 所有交互元素有焦点样式
- [ ] **焦点陷阱**: 弹窗/菜单内焦点循环
- [ ] **快捷键**: 提供键盘快捷键（如 Ctrl+K 搜索）
- [ ] **Esc 关闭**: 弹窗/菜单可用 Esc 关闭

### 视觉可访问性
- [ ] **颜色对比度**: 文本对比度 >= 4.5:1 (AA 级)
- [ ] **不依赖颜色**: 错误状态不只靠红色表示
- [ ] **字体大小**: 支持 200% 缩放不失真
- [ ] **动画控制**: 支持 `prefers-reduced-motion`

### 屏幕阅读器支持
- [ ] **图片替代文本**: 所有 `<img>` 有 `alt` 属性
- [ ] **图标说明**: 装饰性图标 `aria-hidden="true"`
- [ ] **状态通知**: 操作结果通过 `aria-live` 通知
- [ ] **跳过链接**: 提供跳转到主内容的链接

### A11y 检测工具
```bash
# axe-core 检测
cd frontend && npm install @axe-core/cli
npx axe http://localhost:5173

# Lighthouse A11y 检测
cd frontend && npx lighthouse http://localhost:5173 --only-categories=accessibility

# ESLint A11y 插件
# 已配置 eslint-plugin-jsx-a11y
npm run lint
```

### A11y 检查清单
| 检查项 | 工具 | 目标 |
|-------|------|------|
| 颜色对比度 | Lighthouse | 100% 通过 |
| ARIA 使用 | axe-core | 0 严重问题 |
| 键盘导航 | 人工测试 | 全部可访问 |
| 屏幕阅读器 | NVDA/VoiceOver | 可正常使用 |
```

### 2.5.5 前端检验报告模板

```markdown
## 前端专项检验报告

### 检验信息
- **检验日期**: YYYY-MM-DD
- **检验范围**: 前端代码 / 性能 / 安全 / A11y
- **检验人员**: AI / 人工

### 代码质量评分
| 维度 | 得分 | 状态 |
|-----|------|:----:|
| TypeScript 严格性 | 95/100 | ✅ |
| 代码复杂度 | 88/100 | ✅ |
| 代码规范 | 92/100 | ✅ |
| **综合评分** | **92/100** | ✅ |

### 性能评分
| 指标 | 实测值 | 目标值 | 状态 |
|-----|--------|--------|:----:|
| FCP | 1.2s | < 1.8s | ✅ |
| TTI | 2.5s | < 3.8s | ✅ |
| CLS | 0.05 | < 0.1 | ✅ |
| 包体积 | 420KB | < 500KB | ✅ |
| **性能评分** | **95/100** | | ✅ |

### 安全评分
| 检查项 | 发现问题 | 状态 |
|-------|---------|:----:|
| XSS 风险 | 0 | ✅ |
| CSRF 防护 | 通过 | ✅ |
| 依赖漏洞 | 0 高危 | ✅ |
| **安全评分** | **100/100** | ✅ |

### 可访问性评分
| 检查项 | 得分 | 状态 |
|-------|------|:----:|
| 语义化 HTML | 95/100 | ✅ |
| ARIA 标签 | 90/100 | ✅ |
| 键盘导航 | 100/100 | ✅ |
| 颜色对比度 | 100/100 | ✅ |
| **A11y 评分** | **96/100** | ✅ |

### 问题清单
| 严重程度 | 问题描述 | 位置 | 修复建议 |
|---------|---------|------|---------|
| 中 | 组件超过 300 行 | `pages/Dashboard.tsx:1` | 拆分为子组件 |
| 低 | 缺少 alt 文本 | `components/Chart.tsx:45` | 添加描述性 alt |

### 检验结论
- **整体状态**: ✅ 通过 / ⚠️ 有条件通过 / 🔴 不通过
- **是否可进入 Phase 3**: 是 / 否
- **备注**:
```

### 2.5.6 强制检查清单
- [ ] TypeScript 严格模式无错误？
- [ ] 组件/函数复杂度在阈值内？
- [ ] Lighthouse 性能评分 >= 90？
- [ ] 无 XSS/CSRF 安全风险？
- [ ] npm audit 无高危漏洞？
- [ ] 可访问性评分 >= 90？
- [ ] 键盘导航完整可用？

**⚠️ 阻断规则**:
- 发现 XSS/CSRF 高危漏洞 → 阻断
- Lighthouse 性能评分 < 70 → 阻断
- 可访问性评分 < 80 → 阻断
- npm audit 发现高危漏洞 → 阻断

---

## Phase 3: 架构师评审（L2+）

### SOLID 原则检查

```markdown
## 架构师评审报告

### SOLID 原则
- [ ] **S**ingle Responsibility: 每个函数/类职责单一
- [ ] **O**pen/Closed: 对扩展开放，对修改关闭
- [ ] **L**iskov Substitution: 子类可替换父类
- [ ] **I**nterface Segregation: 接口粒度适中
- [ ] **D**ependency Inversion: 依赖抽象而非具体

### 模块划分
```
┌─────────────────────────────────────┐
│           API 路由层                 │
├─────────────────────────────────────┤
│  handlers/  (HTTP 请求处理)          │
│  ├── 职责: 参数解析、调用 Service    │
│  └── 禁止: 业务逻辑                  │
├─────────────────────────────────────┤
│  services/  (业务逻辑层)             │
│  ├── 职责: 核心业务逻辑              │
│  ├── 允许: 调用其他 Service          │
│  └── 禁止: HTTP 相关操作             │
├─────────────────────────────────────┤
│  models/    (数据访问层)             │
│  ├── 职责: 数据模型、数据库操作      │
│  └── 禁止: 业务逻辑                  │
├─────────────────────────────────────┤
│  utils/     (工具函数)               │
│  └── 职责: 纯函数、无状态            │
└─────────────────────────────────────┘
```

### 依赖关系检查
- [ ] Handler → Service → Model（单向依赖）
- [ ] Service 之间可互相调用（避免循环）
- [ ] Utils 被所有层依赖（无反向依赖）

### 接口设计审查
- [ ] 接口粒度适中 (方法数 < 10)
- [ ] 方法命名清晰 (动词+名词)
- [ ] 参数数量合理 (<= 5)
- [ ] 返回值明确
- [ ] 无冗余接口

### 扩展性评估
- [ ] 是否支持水平扩展?
- [ ] 是否支持插件化?
- [ ] 配置是否外部化?
- [ ] 是否支持热更新?
```

---

## Phase 4: QA 评审（L3）

### 测试策略

```markdown
## QA 评审

### 后端测试矩阵
| 功能点 | 正常 Case | 异常 Case | 边界 Case |
|-------|:---------:|:---------:|:---------:|
| API 接口 | ✅ | ✅ | ✅ |
| 业务逻辑 | ✅ | ✅ | ✅ |
| 数据库操作 | ✅ | ✅ | ✅ |
| 并发场景 | ⚠️ | ⚠️ | ⚠️ |

### 前端测试矩阵
| 功能点 | 渲染测试 | 交互测试 | E2E 测试 |
|-------|:--------:|:--------:|:--------:|
| 组件 | ✅ | ✅ | ⚠️ |
| 页面 | ✅ | ⚠️ | ⚠️ |
| API 集成 | ⚠️ | ⚠️ | ✅ |

### 边界条件清单
- [ ] 空值处理（null/nil/undefined/""）
- [ ] 超大输入（max length, max items）
- [ ] 并发访问（race condition）
- [ ] 资源耗尽（OOM, disk full）
- [ ] 网络中断（timeout, retry）
- [ ] 编码/时区问题
- [ ] 浏览器兼容（Chrome, Firefox, Safari）

### 性能测试基准
| 指标 | 目标 | 测试方法 |
|-----|------|---------|
| API 响应时间 | P95 < 200ms | 压测工具 |
| 前端首屏加载 | < 3s | Lighthouse |
| 测试覆盖率 | >= 80% | 覆盖率工具 |
```

---

## Phase 5: CSO 安全评审（L3）

### 安全扫描清单

```markdown
## 安全评审

### 后端安全
- [ ] **无硬编码密钥**: API Key 全部环境变量
- [ ] **SQL 注入防护**: 使用参数化查询
- [ ] **输入验证**: 所有用户输入验证
- [ ] **错误信息**: 不暴露敏感信息
- [ ] **日志脱敏**: 敏感字段打码（password/token/api_key）
- [ ] **CORS 配置**: 限制允许域名
- [ ] **速率限制**: 防止暴力攻击

### 前端安全
- [ ] **XSS 防护**: 不直接使用 innerHTML/dangerouslySetInnerHTML
- [ ] **CSRF 防护**: 请求携带 Token
- [ ] **敏感数据**: 不存储在 LocalStorage
- [ ] **依赖安全**: 检查已知漏洞（npm audit）

### 敏感信息检测
```bash
# 搜索潜在风险
grep -r "password\|secret\|api_key\|token" \
  --include="*.go" --include="*.ts" --include="*.tsx" \
  --exclude-dir=node_modules --exclude-dir=vendor
```

### 安全扫描规则
```yaml
security_rules:
  secrets:
    - pattern: "password\\s*=\\s*[\"'][^\"']+[\"']"
      severity: critical
    - pattern: "api_key\\s*=\\s*[\"'][^\"']+[\"']"
      severity: critical

  injection:
    - pattern: "fmt\\.Sprintf.*%s.*query"
      severity: high

  weak_crypto:
    - pattern: "md5|DES|RC4"
      severity: critical
      message: "禁用弱加密算法，使用 AES/SHA256 等"
```

### 合规检查
- [ ] 符合 OWASP Top 10
- [ ] 数据加密（传输 + 静态）
```

---

## Phase 6: 编码实现

### TDD 流程

```
红 → 绿 → 重构
│     │      │
│     │      └── 优化代码结构
│     └───────── 编写最小实现
└─────────────── 编写失败测试
```

### 后端 TDD 示例

```markdown
## 后端 TDD 流程

### Step 1: 写测试（红）
```go
func TestCalculateSharpeRatio(t *testing.T) {
    tests := []struct {
        name     string
        input    decimal.Decimal
        expected decimal.Decimal
    }{
        {"正常情况", decimal.NewFromFloat(0.15), decimal.NewFromFloat(0.8)},
        {"零波动率", decimal.NewFromFloat(0), decimal.NewFromFloat(0)},
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

### Step 2: 运行测试（失败）
```bash
go test ./services/... -v -run TestCalculateSharpeRatio
# 预期: FAIL
```

### Step 3: 写实现（绿）
```go
func CalculateSharpeRatio(returns decimal.Decimal) decimal.Decimal {
    // 最小实现
    if returns.IsZero() {
        return decimal.Zero
    }
    return returns.Mul(decimal.NewFromFloat(5.33))
}
```

### Step 4: 运行测试（通过）
```bash
go test ./services/... -v -run TestCalculateSharpeRatio
# 预期: PASS
```

### Step 5: 重构
- 优化代码结构
- 提取公共函数
- 保持测试通过
```

### 前端编码流程

```markdown
## 前端编码流程

### Step 1: 类型定义
```typescript
// types/optimization.ts
export interface OptimizationRequest {
  symbols: string[];
  objective: 'min_volatility' | 'max_sharpe' | 'target_return';
}

export interface OptimizationResult {
  weights: Record<string, number>;
  expectedReturn: number;
  volatility: number;
}
```

### Step 2: API 封装
```typescript
// services/optimizationAPI.ts
import api from './api';

export const optimizationAPI = {
  optimize: (params: OptimizationRequest) =>
    api.post<OptimizationResult>('/api/portfolio/optimize', params),
};
```

### Step 3: Hook 实现
```typescript
// hooks/useOptimization.ts
import { useMutation } from '@tanstack/react-query';
import { optimizationAPI } from '../services/optimizationAPI';

export const useOptimization = () => {
  return useMutation({
    mutationFn: optimizationAPI.optimize,
  });
};
```

### Step 4: 组件实现
```typescript
// components/OptimizationForm.tsx
import { useOptimization } from '../hooks/useOptimization';

export const OptimizationForm: React.FC = () => {
  const { mutate, isPending, error } = useOptimization();
  // ...
};
```

### Step 5: 测试
```typescript
// components/OptimizationForm.test.tsx
import { render, screen } from '@testing-library/react';
import { OptimizationForm } from './OptimizationForm';

test('renders form', () => {
  render(<OptimizationForm />);
  expect(screen.getByText('优化')).toBeInTheDocument();
});
```

### 提交规范
```bash
# 提交信息格式
git commit -m "feat(optimization): 添加夏普比率计算

- 实现夏普比率计算公式
- 添加边界条件处理
- 单元测试覆盖率 100%

Closes #123"
```
```

### 编码规范

- [ ] 每次提交 < 200 行
- [ ] 提交信息符合 Conventional Commits
- [ ] 测试先行
- [ ] 频繁运行测试 (每 5-10 分钟)

---

## Phase 7: 验证交付

### 后端验证

```bash
# 1. 运行测试
cd backend && go test ./... -v

# 2. 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 3. 代码格式化检查
gofmt -l .

# 4. 静态分析
go vet ./...

# 5. 构建检查
go build ./...

# 6. 竞态检测
go test ./... -race
```

### 前端验证

```bash
# 1. 类型检查
cd frontend && npx tsc --noEmit

# 2. Lint 检查
npm run lint

# 3. 运行测试
npm run test

# 4. 生成覆盖率报告
npm run test:coverage

# 5. 构建检查
npm run build
```

### 覆盖率报告模板

```markdown
## 测试覆盖率报告

### 后端覆盖率
| 模块 | 覆盖率 | 状态 |
|------|--------|:----:|
| services/optimization | 85% | ✅ |
| services/factor | 72% | ⚠️ |
| handlers | 45% | 🔴 |

### 前端覆盖率
| 模块 | 覆盖率 | 状态 |
|------|--------|:----:|
| components | 60% | ⚠️ |
| pages | 40% | 🔴 |

### 未覆盖代码分析
- `services/xxx.go:45-50` - 错误处理分支，建议补充测试
```

### 交付检查清单

```markdown
## 交付检查清单

### 功能完整性
- [ ] 实现 PLAN.md 中所有功能点
- [ ] 边界条件处理完善
- [ ] 错误处理完善

### 代码质量
- [ ] 后端测试覆盖率 >= 80%
- [ ] 前端组件测试通过
- [ ] 无 Lint 错误
- [ ] 无 TypeScript 类型错误
- [ ] 代码格式化通过

### 文档更新
- [ ] API 文档更新（如需要）
- [ ] AGENTS.md 同步（如需要）
- [ ] CHANGELOG 更新

### 提交准备
- [ ] 提交信息符合规范
- [ ] 相关 Issue 已关联
```

---

## 专用流程指令

### `/plan` - 规划流程

```markdown
## 规划流程执行步骤

1. **阅读 AGENTS.md**: 了解项目背景和规范
2. **分析需求**: 与用户确认具体需求
3. **评估复杂度**: L1/L2/L3
4. **创建 Design Doc**（L2+）:
   - 确定编号 NNN
   - 调研现有代码
   - 方案对比
   - 记录决策
5. **生成 PLAN.md**:
   - 后端规划（如果是后端任务）
   - 前端规划（如果是前端任务）
   - 双端规划（如果是双端任务）
6. **等待用户确认**
7. **进入下一阶段**
```

### `/review` - 代码审查

```markdown
## 代码审查检查清单

### 后端审查
- [ ] 代码符合 Go 规范
- [ ] 错误处理完善
- [ ] 使用 decimal 处理金额
- [ ] 测试覆盖率达标
- [ ] 无循环依赖
- [ ] 函数长度合理

### 前端审查
- [ ] 无 any 类型
- [ ] 组件粒度适中
- [ ] Hooks 使用规范
- [ ] 测试覆盖关键路径
- [ ] 性能优化考虑

### 通用审查
- [ ] 命名清晰
- [ ] 注释恰当
- [ ] 无重复代码
- [ ] 安全无漏洞
```

### `/ship` - 发布准备

```markdown
## 发布准备流程

1. **最终测试**
   ```bash
   # 后端
   cd backend && go test ./... -race

   # 前端
   cd frontend && npm run test && npm run build
   ```

2. **代码格式化**
   ```bash
   # 后端
   gofmt -w .

   # 前端
   npm run lint -- --fix
   ```

3. **提交信息生成**
   ```
   <type>(<scope>): <subject>

   <body>

   <footer>
   ```

4. **最终检查**
   - [ ] 所有测试通过
   - [ ] 无 Lint 错误
   - [ ] 构建成功
   - [ ] 文档已更新
```

---

## 评审异常处理流程

当评审阶段未通过时：

```
评审结果
  │
  ├─ Phase 2 不通过
  │   ├─ 架构设计缺陷 → 修改设计方案 → 重新提交 Phase 2
  │   └─ 技术栈不合规 → 调整技术选型 → 重新提交 Phase 2
  │
  ├─ Phase 3 不通过
  │   ├─ 循环依赖 → 重构模块依赖 → 重新提交 Phase 3
  │   └─ 接口设计不合理 → 调整接口定义 → 重新提交 Phase 3
  │
  ├─ Phase 4 不通过
  │   ├─ 测试覆盖不足 → 补充测试用例 → 重新提交 Phase 4
  │   └─ 边界条件遗漏 → 补充边界处理 → 重新提交 Phase 4
  │
  ├─ Phase 5 不通过
  │   ├─ 安全漏洞 → 修复安全问题 → 重新提交 Phase 5
  │   └─ 敏感信息泄露 → 清理敏感数据 → 重新提交 Phase 5
  │
  └─ 连续 3 次不通过
      └─ 升级处理 → 召集相关人员讨论 → 更新 PLAN.md → 从 Phase 1 重新开始
```

### 异常处理原则

| 原则 | 说明 |
|-----|------|
| 小范围修复 | 仅修改不通过的部分，不扩大变更范围 |
| 记录原因 | 每次不通过必须记录原因和修改方案 |
| 限制重试 | 同一 Phase 连续 3 次不通过，升级处理 |
| 回溯检查 | 修改后需确认不影响已通过的 Phase |

---

## 冲突仲裁机制

当多角色评审意见冲突时：

| 冲突类型 | 示例 | 仲裁规则 |
|---------|------|---------|
| 架构 vs 性能 | 架构师要求分层，QA 担心性能 | 先满足架构，再优化性能 |
| 安全 vs 便利 | CSO 要求加密，产品经理要求快速 | 安全优先 |
| 质量 vs 进度 | QA 要求更多测试，项目要求按时交付 | 核心路径必须测试，非核心可后续补充 |

---

## 回滚机制

### 回滚触发条件
- [ ] 生产环境发现严重 bug
- [ ] 性能下降超过 20%
- [ ] 安全漏洞被利用
- [ ] 数据丢失或损坏

### 回滚策略
```markdown
## 回滚计划

### 数据库回滚
- [ ] 迁移脚本支持 down 操作
- [ ] 数据备份已创建
- [ ] 回滚时间窗口: < 5 分钟

### 代码回滚
- [ ] 上一个版本 tag 已创建
- [ ] 回滚命令可用
- [ ] 部署脚本支持快速回滚

### 配置回滚
- [ ] 配置版本控制
- [ ] 功能开关控制
- [ ] 灰度发布支持
```

---

## 强制阻断规则

以下情况必须阻断流程：

| 阶段 | 阻断条件 |
|-----|---------|
| Step 0 | 无法评估复杂度 |
| Phase 0.5 | L2+ 未创建 Design Doc / 未提交仓库 / 未包含方案对比 |
| Phase 1 | 未生成 PLAN.md / 未识别边界条件 / 无回滚策略 / L2+ 未引用 Design Doc |
| Phase 2 | 技术栈不合规 / 架构设计有缺陷 |
| Phase 2.5 | XSS/CSRF 高危漏洞 / Lighthouse 性能 < 70 / A11y 评分 < 80 / npm audit 高危漏洞 |
| Phase 3 | 存在循环依赖 / 接口设计不合理 |
| Phase 4 | 测试覆盖率目标未设定 / 边界条件未识别 |
| Phase 5 | 发现安全漏洞 / 敏感信息泄露风险 |
| Phase 6 | 测试未通过 / 代码规范检查失败 |
| Phase 7 | 覆盖率不达标 / 存在竞态/泄漏 / 性能下降 > 20% |

---

## 度量指标 (KPI)

### 流程效率
| 指标 | 目标 | 测量方法 |
|-----|------|---------|
| 计划准确率 | > 90% | 实际变更/计划变更 |
| 评审通过率 | > 80% | 首次通过/总评审数 |
| 返工率 | < 10% | 返工次数/总任务数 |

### 代码质量
| 指标 | 目标 | 测量方法 |
|-----|------|---------|
| 测试覆盖率 | > 80% | 项目覆盖率工具 |
| 圈复杂度 | < 10 | 项目静态分析工具 |
| 代码重复率 | < 5% | 项目重复检测工具 |
| 安全漏洞 | 0 高危 | 安全扫描工具 |

### 交付效率
| 指标 | 目标 | 测量方法 |
|-----|------|---------|
| 构建时间 | < 5 分钟 | CI/CD 记录 |
| 部署时间 | < 10 分钟 | CI/CD 记录 |
| 恢复时间 | < 30 分钟 | 事故记录 |

---

## 快速检查表

### 启动任务前
- [ ] 是否已创建 Todo 列表?
- [ ] 是否已阅读相关代码?
- [ ] 是否已理解需求?
- [ ] 是否已评估复杂度?
- [ ] 项目配置是否正确填充?

### 每个 Phase 完成后
- [ ] 是否完成当前阶段输出?
- [ ] 是否通过当前阶段检查?
- [ ] 是否更新 Todo 状态?
- [ ] 是否记录决策理由?

### 编码过程中
- [ ] 是否测试先行?
- [ ] 是否频繁验证?
- [ ] 是否遵循规范?
- [ ] 是否及时提交?

### 交付前
- [ ] 是否所有测试通过?
- [ ] 是否覆盖率达标?
- [ ] 是否性能基准满足?
- [ ] 是否文档更新?
- [ ] 是否 Changelog 生成?
- [ ] 是否无安全漏洞?

---

## 常用命令速查

### 后端开发
```bash
# 运行测试
cd backend && go test ./... -v

# 运行特定测试
cd backend && go test ./services/... -v -run TestRiskBudget

# 格式化代码
cd backend && go fmt ./...

# 构建
cd backend && go build ./...

# 竞态检测
cd backend && go test ./... -race

# 覆盖率
cd backend && go test ./... -coverprofile=coverage.out
cd backend && go tool cover -html=coverage.out -o coverage.html
```

### 前端开发
```bash
# 安装依赖
cd frontend && npm install

# 启动开发服务器
cd frontend && npm run dev

# 运行 lint
cd frontend && npm run lint

# 类型检查
cd frontend && npx tsc --noEmit

# 运行测试
cd frontend && npm run test

# 构建
cd frontend && npm run build
```

---

## 项目结构速查

```
ETF-Insight/
├── backend/                  # Go 后端
│   ├── handlers/             # HTTP 处理器
│   ├── services/             # 业务逻辑
│   │   ├── optimization/     # 优化算法
│   │   ├── factor/           # 因子分析
│   │   └── backtest/         # 回测引擎
│   ├── models/               # 数据模型
│   ├── middleware/           # 中间件
│   ├── utils/                # 工具函数
│   └── router/               # 路由配置
├── frontend/                 # React 前端
│   ├── src/
│   │   ├── pages/            # 页面组件
│   │   ├── components/       # 可复用组件
│   │   ├── hooks/            # 自定义 Hooks
│   │   ├── services/         # API 服务
│   │   └── types/            # TypeScript 类型
│   └── package.json
├── docs/superpowers/specs/   # 设计文档
├── docs/                     # 项目文档
└── .trae/
    └── skills/
        └── gs3-hybrid/       # 本 Skill
            └── SKILL.md
```

---

## 术语表

| 术语 | 全称 | 说明 |
|-----|------|------|
| Design Doc | Design Document | 编号设计文档，记录重大功能/改动的设计决策 |
| SRP | Single Responsibility Principle | 单一职责原则 |
| OCP | Open-Closed Principle | 开闭原则 |
| DIP | Dependency Inversion Principle | 依赖倒置原则 |
| ISP | Interface Segregation Principle | 接口隔离原则 |
| LSP | Liskov Substitution Principle | 里氏替换原则 |
| TDD | Test-Driven Development | 测试驱动开发 |
| DRY | Don't Repeat Yourself | 避免代码重复 |
| SOLID | SRP+OCP+LSP+ISP+DIP | 面向对象设计五大原则 |
| OWASP | Open Web Application Security Project | 开放式 Web 应用安全项目 |
| CSO | Chief Security Officer | 首席安全官 |
| QA | Quality Assurance | 质量保证 |
| CR | Code Review | 代码审查 |
| KPI | Key Performance Indicator | 关键绩效指标 |
| E2E | End-to-End | 端到端测试 |

---

## 版本历史

| 版本 | 日期 | 变更 |
|-----|------|------|
| v1.0 | 2026-04-28 | 初始版本，针对 ETF-Insight 项目定制 |
| v2.0 | 2026-04-28 | 基于 gs-hybrid-v3.1 优化，补充完整流程 |
| v2.1 | 2026-04-28 | 新增 Phase 2.5 前端专项检验层（代码质量/性能/安全/A11y） |

---

*本 Skill 专为 ETF-Insight 项目设计，结合 Superpowers 逻辑规划与 gstack 工程标准*
