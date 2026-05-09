# ETF-Insight 文档

**更新日期**: 2026-05-09

---

## 设计与计划

核心设计决策和实施计划由 [superpowers](https://github.com/obra/superpowers) 工具链管理。

| 目录 | 说明 | 文件数 |
|------|------|--------|
| [specs/](superpowers/specs/) | 设计文档 — 当前活跃的设计决策 | 4 |
| [plans/](superpowers/plans/) | 实施计划 — 当前活跃的实施计划 | 10 |
| [legacy/](superpowers/legacy/) | 历史设计文档 — 早期手写设计文档 | 5 |

---

## 参考文档

开发参考、路线图、API 规范等。

| 文档 | 说明 |
|------|------|
| [EVOLUTION_ROADMAP_2026.md](reference/EVOLUTION_ROADMAP_2026.md) | 2026 年演进路线图 |
| [PROFESSIONAL_ENHANCEMENT.md](reference/PROFESSIONAL_ENHANCEMENT.md) | 专业功能增强计划 |
| [FRONTEND_BACKEND_INTEGRATION_PLAN.md](reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md) | v2.7 前后端一体化实施方案 |
| [DATA_LAYER_EVOLUTION_PLAN.md](reference/DATA_LAYER_EVOLUTION_PLAN.md) | 数据层演进改造方案 |
| [DATA_LAYER_IMPLEMENTATION_GUIDE.md](reference/DATA_LAYER_IMPLEMENTATION_GUIDE.md) | 数据层改造实施指南 |
| [DATA_LAYER_IMPLEMENTATION_PROGRESS.md](reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md) | 数据层改造进度跟踪 |
| [v2.5_phase1_implementation.md](reference/v2.5_phase1_implementation.md) | v2.5 Phase 1 实施记录 |
| [v2.7_phase1_progress.md](reference/v2.7_phase1_progress.md) | v2.7 Phase 1 进度 |
| [OPTIMIZED_PROMPTS.md](reference/OPTIMIZED_PROMPTS.md) | AI 优化提示指南 |
| [openapi.yaml](openapi.yaml) | OpenAPI 3.0 接口规范 |

---

## 活跃规划

| 文档 | 说明 |
|------|------|
| [MONITORING_ALERTING_PLAN.md](MONITORING_ALERTING_PLAN.md) | 监控告警系统计划 |
| [TEST_COVERAGE_PLAN.md](TEST_COVERAGE_PLAN.md) | 测试覆盖率提升计划 |

---

## 归档

历史审查报告、安全文档等。详见 [archive/](archive/)。

---

## 目录结构

```
docs/
├── README.md                 ← 本文档
├── superpowers/              ← 设计 + 计划（主力）
│   ├── specs/                ← 设计文档（日期制）
│   ├── plans/                ← 实施计划（日期制）
│   └── legacy/               ← 历史设计文档（编号制）
├── reference/                ← 参考文档
├── archive/                  ← 归档文档
├── MONITORING_ALERTING_PLAN.md
├── TEST_COVERAGE_PLAN.md
└── openapi.yaml
```
