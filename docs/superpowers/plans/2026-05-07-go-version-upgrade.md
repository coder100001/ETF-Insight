# Go 版本升级实施计划 (1.24 → 1.26)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将项目中所有 Go 版本引用从 1.24 统一升级到 1.26，并运行 `go fix` 应用现代化改造。

**Architecture:** 纯配置/文档变更，不涉及业务逻辑修改。先改 go.mod，再改 CI/Docker/lint 配置，最后更新文档。每个 task 独立可验证。

**Tech Stack:** Go 1.26, Docker, GitHub Actions, golangci-lint

---

### Task 1: 升级 backend/go.mod

**Files:**
- Modify: `backend/go.mod:3`

- [ ] **Step 1: 修改 go.mod 版本声明**

将 `backend/go.mod` 第 3 行的 `go 1.24` 改为 `go 1.26`：

```go
module etf-insight

go 1.26
```

- [ ] **Step 2: 运行 go fix 应用 modernizers**

```bash
cd backend && go fix ./...
```

预期：无输出（无 modernizers 适用时静默退出）或显示应用的改动。

- [ ] **Step 3: 运行 go mod tidy**

```bash
cd backend && go mod tidy
```

预期：无错误，go.sum 可能更新。

- [ ] **Step 4: 验证编译**

```bash
cd backend && go build ./...
```

预期：无错误。

- [ ] **Step 5: 验证格式**

```bash
cd backend && gofmt -l .
```

预期：无输出（所有文件格式正确）。

---

### Task 2: 升级 tools/doccheck/go.mod

**Files:**
- Modify: `tools/doccheck/go.mod:3`

- [ ] **Step 1: 修改 go.mod 版本声明**

将 `tools/doccheck/go.mod` 的 `go 1.24` 改为 `go 1.26`：

```go
module github.com/coder100001/etf-insight/tools/doccheck

go 1.26
```

- [ ] **Step 2: 运行 go fix**

```bash
cd tools/doccheck && go fix ./...
```

- [ ] **Step 3: 运行 go mod tidy**

```bash
cd tools/doccheck && go mod tidy
```

- [ ] **Step 4: 验证编译**

```bash
cd tools/doccheck && go build -o doccheck .
```

预期：编译成功，生成 `doccheck` 二进制。

---

### Task 3: 升级 Dockerfile

**Files:**
- Modify: `Dockerfile:7`

- [ ] **Step 1: 修改基础镜像版本**

将 `Dockerfile` 第 7 行的 `golang:1.24-alpine` 改为 `golang:1.26-alpine`：

```dockerfile
FROM golang:1.26-alpine AS backend-builder
```

---

### Task 4: 升级 GitHub Actions CI 配置

**Files:**
- Modify: `.github/workflows/ci.yml:20`
- Modify: `.github/workflows/docs-consistency.yml:21`

- [ ] **Step 1: 修改 ci.yml 的 Go 版本**

将 `.github/workflows/ci.yml` 第 20 行的 `go-version: '1.24'` 改为 `go-version: '1.26'`：

```yaml
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
          cache-dependency-path: backend/go.sum
```

- [ ] **Step 2: 修改 docs-consistency.yml 的 Go 版本**

将 `.github/workflows/docs-consistency.yml` 第 21 行的 `go-version: '1.24'` 改为 `go-version: '1.26'`。同时将 `actions/setup-go@v4` 升级为 `actions/setup-go@v5`：

```yaml
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.26'
```

---

### Task 5: 升级 golangci-lint 配置

**Files:**
- Modify: `.golangci.yml:7`

- [ ] **Step 1: 修改 .golangci.yml 的 Go 版本**

将 `.golangci.yml` 第 7 行的 `go: "1.24"` 改为 `go: "1.26"`：

```yaml
run:
  timeout: 5m
  go: "1.26"
```

> 注意：`backend/.golangci.yml` 未指定 go 版本（继承项目级），无需修改。

---

### Task 6: 更新文档中的 Go 版本引用

**Files:**
- Modify: `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md:42`
- Modify: `docs/superpowers/plans/2026-05-06-quality-first-p0-plan.md:9`
- Modify: `design-docs/001-code-quality-analysis-and-improvement.md:31,46`

- [ ] **Step 1: 修改 FRONTEND_BACKEND_INTEGRATION_PLAN.md**

将第 42 行的 `Go 1.21+` 改为 `Go 1.26+`。

- [ ] **Step 2: 修改 quality-first-p0-plan.md**

将第 9 行的 `Go 1.24` 改为 `Go 1.26`。

- [ ] **Step 3: 修改 001-code-quality-analysis-and-improvement.md**

将第 31 行的 `Go 1.24` 改为 `Go 1.26`。
将第 46 行的 `Go 1.24` 改为 `Go 1.26`。

---

### Task 7: 综合验证

**Files:** 无新增修改

- [ ] **Step 1: 后端格式检查**

```bash
cd backend && gofmt -l .
```

预期：无输出。

- [ ] **Step 2: 后端静态分析**

```bash
cd backend && go vet ./...
```

预期：无错误。

- [ ] **Step 3: 后端编译**

```bash
cd backend && go build ./...
```

预期：无错误。

- [ ] **Step 4: 后端单元测试**

```bash
cd backend && go test -short ./...
```

预期：全部通过。

- [ ] **Step 5: doccheck 编译验证**

```bash
cd tools/doccheck && go build -o doccheck .
```

预期：编译成功。

- [ ] **Step 6: 前端构建验证**

```bash
cd frontend && npm run build
```

预期：构建成功。

- [ ] **Step 7: TypeScript 类型检查**

```bash
cd frontend && npx tsc --noEmit
```

预期：无错误。

- [ ] **Step 8: 确认所有版本引用已更新**

```bash
grep -r "1\.24" --include="*.mod" --include="*.yml" --include="*.yaml" --include="Dockerfile" backend/ tools/ .github/ .golangci.yml Dockerfile
```

预期：无输出（所有 1.24 引用已清除）。
