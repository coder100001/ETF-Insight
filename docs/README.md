# ETF-Insight 文档

**更新日期**: 2026-05-13

---

## 设计文档

核心设计决策由 [superpowers](https://github.com/obra/superpowers) 工具链管理。

| 目录 | 说明 | 文件数 |
|------|------|--------|
| [specs/](superpowers/specs/) | 设计文档 — 当前活跃的设计决策 | 6 |
| [legacy/](superpowers/legacy/) | 历史设计文档、旧计划和报告 | 5 |

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

## 目录结构

```
docs/
├── README.md                 ← 本文档
├── superpowers/              ← 设计文档
│   ├── specs/                ← 6 个设计文档
│   └── legacy/               ← 5 个历史文档
└── reference/                ← 2 个参考文档
```

---

## 历史清理记录

- **2026-05-13**: 清理了 archive/ 目录下的所有历史审查报告和旧计划，将过时的计划和报告移到 superpowers/legacy/，大幅精简了文档结构。
