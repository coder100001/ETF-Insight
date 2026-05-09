# ETF-Insight 文档

**更新日期**: 2026-05-09

---

## 设计与计划

核心设计决策和实施计划由 [superpowers](https://github.com/obra/superpowers) 工具链管理。

| 目录 | 说明 | 文件数 |
|------|------|--------|
| [specs/](superpowers/specs/) | 设计文档 — 当前活跃的设计决策 | 3 |
| [plans/](superpowers/plans/) | 实施计划 — 进行中或待实施 | 2 |
| [legacy/](superpowers/legacy/) | 历史设计文档 | 1 |

---

## 参考文档

| 文档 | 说明 | 状态 |
|------|------|------|
| [FRONTEND_BACKEND_INTEGRATION_PLAN.md](reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md) | v2.7 前后端一体化实施方案 | ✅ 待实施 |
| [DATA_LAYER_EVOLUTION_PLAN.md](reference/DATA_LAYER_EVOLUTION_PLAN.md) | 数据层演进改造方案 | ✅ 活跃 |
| [DATA_LAYER_IMPLEMENTATION_GUIDE.md](reference/DATA_LAYER_IMPLEMENTATION_GUIDE.md) | 数据层改造实施指南 | ✅ 活跃 |
| [DATA_LAYER_IMPLEMENTATION_PROGRESS.md](reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md) | 数据层改造进度跟踪 | ⚠️ 需更新 |
| [consistency-README.md](reference/consistency-README.md) | 文档一致性管理机制 | ✅ 活跃 |
| [review_process.md](reference/review_process.md) | 文档评审流程 | ✅ 活跃 |

---

## 归档

历史审查报告、安全文档、过期计划等。详见 [archive/](archive/)。

---

## 目录结构

```
docs/
├── README.md                 ← 本文档
├── superpowers/              ← 设计 + 计划
│   ├── specs/                ← 3 个设计文档
│   ├── plans/                ← 2 个实施计划
│   └── legacy/               ← 1 个历史设计文档
├── reference/                ← 6 个参考文档
└── archive/                  ← 归档文档
```
