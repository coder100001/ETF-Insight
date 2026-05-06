# Pre-commit 与 CI/CD 检查统一化设计文档

> **日期**: 2026-05-06
> **状态**: 已批准
> **关联**: 质量优先演进路线 P0 阶段

---

## 1. 问题陈述

当前 pre-commit 配置与 CI/CD 存在严重不一致，导致：
- **本地 pre-commit 通过，CI 失败**：pre-commit 只跑格式检查，CI 跑 vet/build/test/lint
- **反馈延迟**：开发者提交后才发现问题，需反复 amend/rebase
- **ESLint 和 golangci-lint 被注释**：已知兼容性问题未解决
- **Doccheck 不在 pre-commit 中**：文档一致性只在 CI 中检查

## 2. 设计目标

1. **Pre-commit 拦截 95%+ 的 CI 失败**：本地提交前就能发现绝大多数问题
2. **提交体验 < 15 秒**：通过增量检查、缓存、并行化保持速度
3. **单一配置源**：pre-commit、CI、Makefile 共用同一套检查清单
4. **强制不可跳过**：所有检查必须通过，不允许 `--no-verify`

## 3. 架构设计

### 3.1 统一检查清单

```yaml
# scripts/check-manifest.yaml
stages:
  format:
    description: "格式检查（<2s）"
    commands:
      - name: gofmt
        cmd: "cd backend && gofmt -l ."
        fail_on_output: true
      - name: goimports
        cmd: "cd backend && goimports -l ."
        fail_on_output: true
      - name: prettier
        cmd: "cd frontend && npx prettier --check ."
        optional: true

  static:
    description: "静态分析（<5s）"
    commands:
      - name: go-vet
        cmd: "cd backend && go vet ./..."
        env:
          GOPROXY: "https://goproxy.cn,direct"
      - name: tsc
        cmd: "cd frontend && npx tsc --noEmit"
      - name: eslint
        cmd: "cd frontend && npx eslint --cache ."
        condition: "node_version >= 20"

  build:
    description: "编译检查（<10s）"
    commands:
      - name: go-build
        cmd: "cd backend && go build ./..."
        env:
          GOPROXY: "https://goproxy.cn,direct"
      - name: npm-build
        cmd: "cd frontend && npm run build"

  test:
    description: "单元测试（<30s）"
    commands:
      - name: go-test-short
        cmd: "cd backend && go test -short -count=1 ./..."
        env:
          GOPROXY: "https://goproxy.cn,direct"
      - name: frontend-test
        cmd: "cd frontend && npm run test:run"

  docs:
    description: "文档一致性（<5s）"
    commands:
      - name: doccheck
        cmd: "./tools/doccheck/doccheck --quick --strict"
```

### 3.2 Doccheck 快速模式设计

**核心策略：变更驱动 + 全量规则**

快速模式不是减少检查规则，而是**减少扫描范围**：

```
输入: Git 变更文件列表 (git diff --cached --name-only)

处理流程:
  1. 提取受影响的代码元素
     - 从 .go 文件提取 struct/func 名
     - 从 .ts/.tsx 提取 component/hook 名

  2. 双向一致性检查（只检查受影响的部分）
     a) 代码→文档: 新增/修改的代码元素是否在文档中提及？
     b) 文档→代码: 文档中提到的功能是否在代码中实现？
        - 只检查与变更文件相关的功能映射

  3. 版本一致性（每次必查）
     - README.md 版本 vs package.json 版本
     - 无论文件是否变更都检查

  4. 必需文档存在性（快速检查）
     - AGENTS.md, README.md, docs/ 目录
```

**快速模式 vs 完整模式：**

| 维度 | 快速模式 (pre-commit) | 完整模式 (CI/scheduled) |
|------|---------------------|------------------------|
| 扫描范围 | 变更文件关联的元素 | 所有代码元素 |
| 检查规则 | 全量规则 | 全量规则 |
| 版本检查 | 每次必查 | 每次必查 |
| 功能映射 | 仅相关功能 | 所有功能 |
| 执行时间 | <5秒 | 10-30秒 |
| 退出码 | strict（high severity 失败） | strict（high severity 失败） |

### 3.3 Pre-commit 配置重构

```yaml
# .pre-commit-config.yaml
repos:
  # 基础检查（保留）
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-json
      - id: check-merge-conflict
      - id: check-case-conflict
      - id: mixed-line-ending
      - id: detect-private-key

  # Go 格式（保留）
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-imports

  # 统一检查脚本（新增）
  - repo: local
    hooks:
      - id: unified-checks
        name: Unified Pre-commit Checks
        entry: bash scripts/pre-commit-checks.sh
        language: system
        pass_filenames: false
        always_run: true
        verbose: true
```

### 3.4 CI/CD 配置重构

```yaml
# .github/workflows/ci.yml
jobs:
  backend:
    steps:
      # ... setup ...
      - name: Run unified checks (backend)
        run: ./scripts/run-checks.sh --stage=format,static,build,test --backend
        env:
          GOPROXY: https://goproxy.cn,direct

  frontend:
    steps:
      # ... setup ...
      - name: Run unified checks (frontend)
        run: ./scripts/run-checks.sh --stage=format,static,build,test --frontend

  docs:
    steps:
      - name: Run doccheck (full)
        run: ./tools/doccheck/doccheck --strict
```

### 3.5 Makefile 更新

```makefile
lint:
	@./scripts/run-checks.sh --stage=format,static

test:
	@./scripts/run-checks.sh --stage=test

check:
	@./scripts/run-checks.sh --stage=all
```

## 4. 关键问题解决

### 4.1 ESLint 9.x + Node.js v24 兼容性

**方案**：降级 ESLint 到 8.x 稳定版

```bash
cd frontend
npm uninstall eslint @eslint/js
npm install eslint@^8.57.0
```

或升级 Node 到 v20 LTS（如果项目允许）。

### 4.2 Go 网络超时（D3）

**方案**：所有 Go 命令统一设置 GOPROXY

```bash
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off  # 可选，加速
```

### 4.3 golangci-lint v2 配置迁移

**方案**：暂不启用，先解决基础问题。后续独立任务迁移。

### 4.4 测试速度优化

**方案**：
- `go test -short`：跳过网络依赖测试
- `go test -count=1`：禁用缓存（pre-commit 中）
- 标记慢测试：使用 `t.Skip("skipping slow test in short mode")`

## 5. 文件清单

| 文件 | 操作 | 职责 |
|------|------|------|
| `.pre-commit-config.yaml` | 修改 | 简化配置，调用统一脚本 |
| `.github/workflows/ci.yml` | 修改 | 调用统一脚本替代分散命令 |
| `.github/workflows/docs-consistency.yml` | 修改 | 调用 doccheck 完整模式 |
| `Makefile` | 修改 | 调用统一脚本 |
| `scripts/pre-commit-checks.sh` | 创建 | pre-commit 统一入口 |
| `scripts/run-checks.sh` | 创建 | CI/本地统一入口 |
| `scripts/check-manifest.yaml` | 创建 | 检查清单配置 |
| `tools/doccheck/main.go` | 修改 | 添加 `--quick` 和 `--changed-files` |
| `tools/doccheck/checker/checker.go` | 修改 | 实现快速模式逻辑 |
| `tools/doccheck/checker/quick.go` | 创建 | 快速模式专用检查器 |
| `frontend/package.json` | 可能修改 | ESLint 降级 |

## 6. 验收标准

- [ ] `git commit` 时自动运行完整检查（format + static + build + test + docs）
- [ ] 检查总耗时 < 15 秒（M1 Mac, warm cache）
- [ ] 不允许 `--no-verify` 跳过（通过团队约定 + CI 兜底）
- [ ] CI 使用与 pre-commit 相同的检查清单
- [ ] `make check` 等价于 pre-commit 完整检查
- [ ] Doccheck 快速模式 < 5 秒
- [ ] 新增代码元素无文档时 pre-commit 失败
- [ ] 文档提到未实现功能时 pre-commit 失败
- [ ] 版本不一致时 pre-commit 失败
