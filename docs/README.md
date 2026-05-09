# ETF-Insight 文档

**更新日期**: 2026-05-09

---

## 设计与计划

核心设计决策和实施计划由 [superpowers](https://github.com/obra/superpowers) 工具链管理。

| 目录 | 说明 | 文件数 |
|------|------|--------|
| [specs/](superpowers/specs/) | 设计文档 — 当前活跃的设计决策 | 3 |
| [plans/](superpowers/plans/) | 实施计划 — 进行中或待实施 | 4 |
| [legacy/](superpowers/legacy/) | 历史设计文档 — 早期手写设计文档 | 1 |

---

## 参考文档

开发参考、路线图等。

| 文档 | 说明 | 状态 |
|------|------|------|
| [EVOLUTION_ROADMAP_2026.md](reference/EVOLUTION_ROADMAP_2026.md) | 2026 年演进路线图 | ⚠️ 需更新 |
| [FRONTEND_BACKEND_INTEGRATION_PLAN.md](reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md) | v2.7 前后端一体化实施方案 | ✅ 待实施 |
| [DATA_LAYER_EVOLUTION_PLAN.md](reference/DATA_LAYER_EVOLUTION_PLAN.md) | 数据层演进改造方案 | ✅ 活跃 |
| [DATA_LAYER_IMPLEMENTATION_GUIDE.md](reference/DATA_LAYER_IMPLEMENTATION_GUIDE.md) | 数据层改造实施指南 | ✅ 活跃 |
| [DATA_LAYER_IMPLEMENTATION_PROGRESS.md](reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md) | 数据层改造进度跟踪 | ⚠️ 需更新 |
| [v2.5_phase1_implementation.md](reference/v2.5_phase1_implementation.md) | v2.5 Phase 1 实施记录 | 📦 历史 |
| [v2.7_phase1_progress.md](reference/v2.7_phase1_progress.md) | v2.7 Phase 1 进度 | ⚠️ 需更新 |
| [consistency-README.md](reference/consistency-README.md) | 文档一致性管理机制 | ✅ 活跃 |
| [mapping_rules.json](reference/mapping_rules.json) | 文档与代码映射规则 | ⚠️ 需更新 |
| [review_process.md](reference/review_process.md) | 文档评审流程 | ✅ 活跃 |

---

## 活跃规划

| 文档 | 说明 | 状态 |
|------|------|------|
| [MONITORING_ALERTING_PLAN.md](MONITORING_ALERTING_PLAN.md) | 监控告警系统计划 | ⚠️ 待实施 |
| [TEST_COVERAGE_PLAN.md](TEST_COVERAGE_PLAN.md) | 测试覆盖率提升计划 | ⚠️ 需更新 |

---

## 归档

历史审查报告、安全文档等。详见 [archive/](archive/)（8 个文件）。

---

## 清理记录

**2026-05-09**: 删除 16 个废弃文档：
- 8 个已完成的实施计划（AI Agent、QuantLib、Phase3/4、precommit-ci、go-version-upgrade、code-quality-fix ×2）
- 3 个已废弃的 legacy 设计文档（001、002、004）
- 2 个过时的 reference 文档（OPTIMIZED_PROMPTS、PROFESSIONAL_ENHANCEMENT）
- 1 个过时的 spec（go-version-upgrade-design）
- 1 个严重过时的 openapi.yaml（仅覆盖 7 个端点，实际有 100+）
- 1 个 legacy README（索引已空）

---

## 目录结构

```
docs/
├── README.md                 ← 本文档
├── superpowers/              ← 设计 + 计划
│   ├── specs/                ← 3 个设计文档
│   ├── plans/                ← 4 个实施计划
│   └── legacy/               ← 1 个历史设计文档
├── reference/                ← 10 个参考文档
├── archive/                  ← 8 个归档文档
├── MONITORING_ALERTING_PLAN.md
└── TEST_COVERAGE_PLAN.md
```
