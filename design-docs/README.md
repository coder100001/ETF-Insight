# Design Docs

> 本目录记录项目的重大设计决策，类似 DB Migrations。
> 每份文档记录一次设计思考的完整过程：调研 → 方案对比 → 决策。

## 文档索引

| 编号 | 标题 | 状态 | 创建日期 | 涉及端 | 关联 |
|-----|------|------|---------|--------|------|
| 001 | [代码质量分析与改进](001-code-quality-analysis-and-improvement.md) | approved | 2026-04-28 | 双端 | - |
| 002 | [代码质量优化与修复](002-code-quality-optimization.md) | draft | 2026-04-28 | 双端 | 001 |
| 003 | [项目向金融分析工具演进规划](003-financial-analysis-tool-evolution.md) | draft | 2026-04-30 | 双端 | 001 |
| 004 | [关键缺陷修复与代码质量提升](004-critical-code-fixes.md) | draft | 2026-05-02 | 双端 | 001, 002, 003 |

## 状态说明

| 状态 | 含义 |
|-----|------|
| draft | 草稿，尚未评审 |
| approved | 已通过评审，可进入实现 |
| superseded | 被后续文档取代 |
| deprecated | 已废弃，不再适用 |
