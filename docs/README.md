# ETF-Insight 文档

**更新日期**: 2026-05-10

---

## 设计与计划

核心设计决策和实施计划由 [superpowers](https://github.com/obra/superpowers) 工具链管理。

| 目录 | 说明 | 文件数 |
|------|------|--------|
| [specs/](superpowers/specs/) | 设计文档 — 当前活跃的设计决策 | 4 |
| [plans/](superpowers/plans/) | 实施计划 — 进行中或待实施 | 2 |
| [legacy/](superpowers/legacy/) | 历史设计文档 | 1 |

---

## 参考文档

| 文档 | 说明 | 状态 |
|------|------|------|
| [consistency-README.md](reference/consistency-README.md) | 文档一致性管理机制 | ✅ 活跃 |
| [review_process.md](reference/review_process.md) | 文档评审流程 | ✅ 活跃 |

---

## 最新功能

### 报告导出功能 (2026-05-10)

实现了多格式报告导出功能，支持 HTML、PDF、Excel、Markdown 四种格式。

**相关文档：**
- [报告导出功能设计文档](superpowers/specs/2026-05-10-export-report-design.md) - 完整的设计规格
- [实施分析报告](superpowers/reports/2026-05-10-implementation-analysis-report.md) - 实施过程分析

**API 端点：**
- `POST /api/export/:type` - 导出数据
- `GET /api/export/formats` - 获取支持的格式
- `GET /api/export/types` - 获取支持的数据类型

**支持的格式：**
- HTML - 网页格式，支持样式和交互
- PDF - PDF 文档，适合打印和分享
- Excel - Excel 表格（CSV 格式），支持数据分析
- Markdown - Markdown 格式，适合文档编辑

---

## 归档

历史审查报告、安全文档、过期计划等。详见 [archive/](archive/)。

---

## 目录结构

```
docs/
├── README.md                 ← 本文档
├── superpowers/              ← 设计 + 计划
│   ├── specs/                ← 4 个设计文档
│   ├── plans/                ← 2 个实施计划
│   ├── reports/              ← 实施报告
│   └── legacy/               ← 1 个历史设计文档
├── reference/                ← 2 个参考文档
└── archive/                  ← 归档文档
```