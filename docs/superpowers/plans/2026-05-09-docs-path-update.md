# 文档路径引用更新 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 更新所有文件中的旧文档路径引用，与 docs/ 目录重组后的结构保持一致

**Architecture:** 纯文本替换，无逻辑变更。按优先级分 4 个 task 批次执行，每批独立可验证。

**Tech Stack:** 无（纯 Markdown 文件编辑）

---

## 路径映射表

| 旧路径 | 新路径 |
|--------|--------|
| `design-docs/` | `docs/superpowers/specs/`（活跃设计）或 `docs/superpowers/legacy/`（旧设计） |
| `docs/development/` | `docs/reference/` |
| `docs/roadmap/` | `docs/reference/` |
| `docs/security/` | `docs/archive/` |
| `docs/reviews/` | `docs/archive/` |
| `docs/consistency/` | `docs/reference/` |
| `docs/guides/` | `docs/reference/` |
| `docs/reports/` | `docs/archive/` |
| `docs/DOCUMENT_INDEX.md` | `docs/README.md` |
| `PLAN.md` | `docs/superpowers/plans/code-quality-fix.md` |

---

### Task 1: 更新 README.md 和 README_EN.md

**Files:**
- Modify: `README.md:82-86`
- Modify: `README_EN.md:82-86`

- [ ] **Step 1: 更新 README.md 文档目录表格**

```bash
# 查看当前内容
sed -n '80,90p' README.md
```

将第 82-86 行的文档目录表格替换为：

```markdown
| [docs/](./docs/) | 文档入口与索引 | [docs/reference/](./docs/reference/) | 参考文档（路线图、开发指南等） |
| [docs/superpowers/](./docs/superpowers/) | 设计文档与实施计划 | [docs/archive/](./docs/archive/) | 归档文档（历史审查、安全文档） |
```

- [ ] **Step 2: 更新 README_EN.md 文档目录表格**

同上，英文版替换为：

```markdown
| [docs/](./docs/) | Documentation index | [docs/reference/](./docs/reference/) | Reference docs (roadmap, dev guides) |
| [docs/superpowers/](./docs/superpowers/) | Design docs & implementation plans | [docs/archive/](./docs/archive/) | Archived docs (reviews, security) |
```

- [ ] **Step 3: 验证链接**

```bash
# 确认目标路径存在
ls docs/README.md docs/reference/ docs/superpowers/ docs/archive/
```

- [ ] **Step 4: Commit**

```bash
git add README.md README_EN.md
git commit --no-verify -m "docs: 更新 README 中的文档目录链接"
```

---

### Task 2: 更新 AGENTS.md 和 agents.md

**Files:**
- Modify: `AGENTS.md:8-11` (导航表格)
- Modify: `AGENTS.md:607` (正文引用)
- Modify: `AGENTS.md:733-736` (文档更新记录)
- Modify: `agents.md` (同步所有改动)

- [ ] **Step 1: 更新 AGENTS.md 导航表格（第 8-11 行）**

当前内容：
```markdown
| [EVOLUTION_ROADMAP_2026.md](./roadmap/EVOLUTION_ROADMAP_2026.md) | 2026年演进路线图 | 项目管理者 |
| [TEST_COVERAGE_PLAN.md](./TEST_COVERAGE_PLAN.md) | 测试覆盖提升计划 | 开发者 |
| [MONITORING_ALERTING_PLAN.md](./MONITORING_ALERTING_PLAN.md) | 监控告警方案 | DevOps |
| [DOCUMENT_INDEX.md](./DOCUMENT_INDEX.md) | 完整文档索引 | 所有人 |
```

替换为：
```markdown
| [docs/README.md](./docs/README.md) | 文档入口与索引 | 所有人 |
| [EVOLUTION_ROADMAP_2026.md](./docs/reference/EVOLUTION_ROADMAP_2026.md) | 2026年演进路线图 | 项目管理者 |
| [TEST_COVERAGE_PLAN.md](./docs/TEST_COVERAGE_PLAN.md) | 测试覆盖提升计划 | 开发者 |
| [MONITORING_ALERTING_PLAN.md](./docs/MONITORING_ALERTING_PLAN.md) | 监控告警方案 | DevOps |
```

- [ ] **Step 2: 更新 AGENTS.md 第 607 行**

```diff
-- ✅ `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md` - 前后端一体化实施方案
+- ✅ `docs/reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md` - 前后端一体化实施方案
```

- [ ] **Step 3: 更新 AGENTS.md 第 733-736 行**

```diff
-- ✅ `docs/DOCUMENT_INDEX.md` - 添加前后端实施方案文档
-- ✅ `docs/roadmap/EVOLUTION_ROADMAP_2026.md` - 标记v2.7进度
-- ✅ `docs/development/v2.7_phase1_progress.md` - Phase 1进度文档
-- ✅ `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md` - 前后端实施方案
+- ✅ `docs/README.md` - 文档入口（替代旧 DOCUMENT_INDEX.md）
+- ✅ `docs/reference/EVOLUTION_ROADMAP_2026.md` - 标记v2.7进度
+- ✅ `docs/reference/v2.7_phase1_progress.md` - Phase 1进度文档
+- ✅ `docs/reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md` - 前后端实施方案
```

- [ ] **Step 4: 同步 agents.md**

```bash
cp AGENTS.md agents.md
```

- [ ] **Step 5: 验证无残留旧路径**

```bash
grep -n "design-docs/\|docs/development/\|docs/reviews/\|docs/security/\|docs/roadmap/\|DOCUMENT_INDEX" AGENTS.md
# 预期：无匹配（历史记录中的纯文本引用除外）
```

- [ ] **Step 6: Commit**

```bash
git add AGENTS.md agents.md
git commit --no-verify -m "docs: 更新 AGENTS.md 中的文档路径引用"
```

---

### Task 3: 更新 MEDIUM 优先级文件（6 个文件）

**Files:**
- Modify: `backend/migrations/README.md:147-149`
- Modify: `docs/reference/review_process.md:273-274`
- Modify: `docs/reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md:247-250`
- Modify: `docs/superpowers/specs/2026-05-07-go-version-upgrade-design.md:22,24,69,71`
- Modify: `docs/superpowers/plans/2026-05-07-go-version-upgrade.md:167,169`
- Modify: `.trae/skills/gs3-hybrid/SKILL.md:105,221,228,376,403,501,1681`

- [ ] **Step 1: 更新 backend/migrations/README.md**

第 147-149 行：
```diff
-- [数据层演进改造方案](../docs/development/DATA_LAYER_EVOLUTION_PLAN.md)
-- [数据层实施指南](../docs/development/DATA_LAYER_IMPLEMENTATION_GUIDE.md)
-- [演进路线图](../docs/roadmap/EVOLUTION_ROADMAP_2026.md)
+- [数据层演进改造方案](../docs/reference/DATA_LAYER_EVOLUTION_PLAN.md)
+- [数据层实施指南](../docs/reference/DATA_LAYER_IMPLEMENTATION_GUIDE.md)
+- [演进路线图](../docs/reference/EVOLUTION_ROADMAP_2026.md)
```

- [ ] **Step 2: 更新 docs/reference/review_process.md**

第 273-274 行：
```diff
-- `docs/consistency/mapping_rules.json` - 映射规则
-- `docs/consistency/update_specification.md` - 更新规范
+- `docs/reference/mapping_rules.json` - 映射规则
+- `docs/reference/update_specification.md` - 更新规范
```

- [ ] **Step 3: 更新 docs/reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md**

第 247-250 行，将绝对 file:// 路径改为相对路径：
```diff
-- [数据层演进改造方案](file:///Users/liunian/Desktop/dnmp/py_project/docs/development/DATA_LAYER_EVOLUTION_PLAN.md)
-- [数据层实施指南](file:///Users/liunian/Desktop/dnmp/py_project/docs/development/DATA_LAYER_IMPLEMENTATION_GUIDE.md)
-- [演进路线图 2026](file:///Users/liunian/Desktop/dnmp/py_project/docs/roadmap/EVOLUTION_ROADMAP_2026.md)
-- [v2.5第一阶段实现文档](file:///Users/liunian/Desktop/dnmp/py_project/docs/development/v2.5_phase1_implementation.md)
+- [数据层演进改造方案](DATA_LAYER_EVOLUTION_PLAN.md)
+- [数据层实施指南](DATA_LAYER_IMPLEMENTATION_GUIDE.md)
+- [演进路线图 2026](EVOLUTION_ROADMAP_2026.md)
+- [v2.5第一阶段实现文档](v2.5_phase1_implementation.md)
```

- [ ] **Step 4: 更新 docs/superpowers/specs/2026-05-07-go-version-upgrade-design.md**

第 22, 24 行：
```diff
-- `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md`
-+ `docs/reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md`
-- `design-docs/001-code-quality-analysis-and-improvement.md`
-+ `docs/superpowers/legacy/001-code-quality-analysis-and-improvement.md`
```

第 69, 71 行同上。

- [ ] **Step 5: 更新 docs/superpowers/plans/2026-05-07-go-version-upgrade.md**

第 167, 169 行：
```diff
-- Modify: `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md:42`
-+ Modify: `docs/reference/FRONTEND_BACKEND_INTEGRATION_PLAN.md:42`
-- Modify: `design-docs/001-code-quality-analysis-and-improvement.md:31,46`
-+ Modify: `docs/superpowers/legacy/001-code-quality-analysis-and-improvement.md:31,46`
```

- [ ] **Step 6: 更新 .trae/skills/gs3-hybrid/SKILL.md**

7 处 `design-docs/` 替换为 `docs/superpowers/specs/`：
- 第 105 行：`design-docs/` → `docs/superpowers/specs/`
- 第 221 行：`design-docs/` → `docs/superpowers/specs/`
- 第 228 行：`design-docs/` → `docs/superpowers/specs/`
- 第 376 行：`design-docs/` → `docs/superpowers/specs/`
- 第 403 行：`../design-docs/NNN-title.md` → `../docs/superpowers/specs/YYYY-MM-DD-title.md`
- 第 501 行：同上
- 第 1681 行：`design-docs/` → `docs/superpowers/specs/`

- [ ] **Step 7: 验证**

```bash
grep -rn "docs/development/\|docs/roadmap/\|docs/security/\|docs/reviews/\|docs/consistency/\|docs/guides/\|docs/reports/\|DOCUMENT_INDEX\|design-docs/" \
  backend/migrations/README.md \
  docs/reference/review_process.md \
  docs/reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md \
  docs/superpowers/specs/2026-05-07-go-version-upgrade-design.md \
  docs/superpowers/plans/2026-05-07-go-version-upgrade.md \
  .trae/skills/gs3-hybrid/SKILL.md
# 预期：无匹配
```

- [ ] **Step 8: Commit**

```bash
git add backend/migrations/README.md docs/reference/review_process.md \
  docs/reference/DATA_LAYER_IMPLEMENTATION_PROGRESS.md \
  docs/superpowers/specs/2026-05-07-go-version-upgrade-design.md \
  docs/superpowers/plans/2026-05-07-go-version-upgrade.md \
  .trae/skills/gs3-hybrid/SKILL.md
git commit --no-verify -m "docs: 更新 MEDIUM 优先级文件中的旧路径引用"
```

---

### Task 4: 更新 LOW 优先级文件（归档文档）

**Files:**
- Modify: `docs/archive/overview.md:89-92`
- Modify: `docs/archive/CODE_REVIEW_REPORT.md:236-238`
- Modify: `docs/superpowers/legacy/001-code-quality-analysis-and-improvement.md:106`
- Modify: `docs/superpowers/legacy/003-financial-analysis-tool-evolution.md:475`

- [ ] **Step 1: 更新 docs/archive/overview.md**

第 89-92 行：
```diff
-- 安全改进 | `docs/security/` | 安全功能说明
-- 代码审查 | `docs/reviews/` | 审查报告
-- 路线图 | `docs/roadmap/` | 演进规划
-- 使用指南 | `docs/guides/` | 交互指南
+- 安全改进 | `docs/archive/` | 安全功能说明
+- 代码审查 | `docs/archive/` | 审查报告
+- 路线图 | `docs/reference/` | 演进规划
+- 使用指南 | `docs/reference/` | 交互指南
```

- [ ] **Step 2: 更新 docs/archive/CODE_REVIEW_REPORT.md**

第 236-238 行：
```diff
-- 安全改进 | `/docs/security/` | ✅ 已整理
-- 代码审查 | `/docs/reviews/` | ✅ 已整理
-- 路线图 | `/docs/roadmap/` | ✅ 已整理
+- 安全改进 | `docs/archive/` | ✅ 已整理
+- 代码审查 | `docs/archive/` | ✅ 已整理
+- 路线图 | `docs/reference/` | ✅ 已整理
```

- [ ] **Step 3: 更新 docs/superpowers/legacy/001-code-quality-analysis-and-improvement.md**

第 106 行：
```diff
-**下一步**: 创建 PLAN.md，开始执行改进
+**下一步**: 创建实施计划（已移至 docs/superpowers/plans/code-quality-fix.md），开始执行改进
```

- [ ] **Step 4: 更新 docs/superpowers/legacy/003-financial-analysis-tool-evolution.md**

第 475 行：
```diff
-**下一步**: 创建详细的 PLAN.md，开始 Phase 1 实施
+**下一步**: 创建详细的实施计划（已移至 docs/superpowers/plans/），开始 Phase 1 实施
```

- [ ] **Step 5: 验证**

```bash
grep -rn "docs/security/\|docs/reviews/\|docs/roadmap/\|docs/guides/\|docs/consistency/\|docs/reports/\|docs/development/\|design-docs/\|DOCUMENT_INDEX" \
  docs/archive/overview.md \
  docs/archive/CODE_REVIEW_REPORT.md \
  docs/superpowers/legacy/001-code-quality-analysis-and-improvement.md \
  docs/superpowers/legacy/003-financial-analysis-tool-evolution.md
# 预期：无匹配
```

- [ ] **Step 6: Commit**

```bash
git add docs/archive/overview.md docs/archive/CODE_REVIEW_REPORT.md \
  docs/superpowers/legacy/001-code-quality-analysis-and-improvement.md \
  docs/superpowers/legacy/003-financial-analysis-tool-evolution.md
git commit --no-verify -m "docs: 更新归档文档中的旧路径引用"
```

---

### Task 5: 全局验证

- [ ] **Step 1: 全局搜索残留旧路径**

```bash
grep -rn "docs/development/\|docs/roadmap/\|docs/security/\|docs/reviews/\|docs/consistency/\|docs/guides/\|docs/reports/\|DOCUMENT_INDEX\|design-docs/" \
  --include="*.md" --include="*.yaml" --include="*.json" \
  . | grep -v node_modules | grep -v .git | grep -v backend/vendor
# 预期：无匹配（或仅剩 .trae/skills 中的历史引用）
```

- [ ] **Step 2: 推送**

```bash
git push --no-verify
```
