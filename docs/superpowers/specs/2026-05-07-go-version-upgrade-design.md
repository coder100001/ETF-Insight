# Go 版本升级设计文档：1.24 → 1.26

> **日期**: 2026-05-07
> **状态**: 已批准
> **关联**: 质量优先演进路线

---

## 1. 问题陈述

当前项目 Go 版本声明不一致，且落后于最新稳定版：

| 位置 | 当前版本 | 目标版本 |
|------|----------|----------|
| `backend/go.mod` | `go 1.24` | `go 1.26` |
| `tools/doccheck/go.mod` | `go 1.24` | `go 1.26` |
| `Dockerfile` | `golang:1.24-alpine` | `golang:1.26-alpine` |
| `.github/workflows/ci.yml` | `go-version: '1.24'` | `go-version: '1.26'` |
| `.github/workflows/docs-consistency.yml` | `go-version: '1.24'` | `go-version: '1.26'` |
| `.golangci.yml` | `go: "1.24"` | `go: "1.26"` |
| `backend/.golangci.yml` | 未指定 | 保持不指定（继承项目级） |
| `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md` | `Go 1.21+` | `Go 1.26+` |
| `docs/superpowers/plans/2026-05-06-quality-first-p0-plan.md` | `Go 1.24` | `Go 1.26` |
| `design-docs/001-code-quality-analysis-and-improvement.md` | `Go 1.24` | `Go 1.26` |

---

## 2. 设计目标

1. **版本统一**：所有代码和文档中的 Go 版本引用统一为 `1.26`
2. **向后兼容**：`go.mod` 使用 `go 1.26`（非 `toolchain` 指令），保持模块兼容性
3. **利用新特性**：运行 `go fix` 应用 Go 1.26 modernizers
4. **性能收益**：Green Tea GC 默认启用，无需代码改动

---

## 3. Go 1.26 关键变化（与本项目相关）

### 3.1 语言变化
- `new(expr)` 支持表达式作为初始值（可选使用）
- 泛型类型可在自身类型参数列表中自引用（可选使用）

### 3.2 工具链
- `go fix` 重写，包含 modernizers 分析器
- `go mod init` 默认使用更低版本（不影响现有模块）

### 3.3 运行时
- **Green Tea GC 默认启用**：预计 GC 开销降低 10-40%
- cgo 调用开销降低约 30%

### 3.4 标准库
- 新增 `crypto/hpke`、`crypto/mlkem/mlkemtest`、`testing/cryptotest`
- `math/rand/v2` 可用（可选迁移）

---

## 4. 实施方案

### 4.1 文件清单与修改

| # | 文件 | 操作 | 修改内容 |
|---|------|------|----------|
| 1 | `backend/go.mod` | 修改 | `go 1.24` → `go 1.26` |
| 2 | `tools/doccheck/go.mod` | 修改 | `go 1.24` → `go 1.26` |
| 3 | `Dockerfile` | 修改 | `golang:1.24-alpine` → `golang:1.26-alpine` |
| 4 | `.github/workflows/ci.yml` | 修改 | `go-version: '1.24'` → `go-version: '1.26'` |
| 5 | `.github/workflows/docs-consistency.yml` | 修改 | `go-version: '1.24'` → `go-version: '1.26'` |
| 6 | `.golangci.yml` | 修改 | `go: "1.24"` → `go: "1.26"` |
| 7 | `docs/development/FRONTEND_BACKEND_INTEGRATION_PLAN.md` | 修改 | `Go 1.21+` → `Go 1.26+` |
| 8 | `docs/superpowers/plans/2026-05-06-quality-first-p0-plan.md` | 修改 | `Go 1.24` → `Go 1.26` |
| 9 | `design-docs/001-code-quality-analysis-and-improvement.md` | 修改 | `Go 1.24` → `Go 1.26` |
| 10 | `AGENTS.md` | 检查 | 如有 Go 版本引用则更新 |

### 4.2 `go fix` 现代化步骤

```bash
# 1. 升级 Go 版本声明
cd backend && go mod edit -go=1.26
cd ../tools/doccheck && go mod edit -go=1.26

# 2. 运行 go fix 应用 modernizers
cd backend && go fix ./...
cd ../tools/doccheck && go fix ./...

# 3. 整理依赖
cd backend && go mod tidy
cd ../tools/doccheck && go mod tidy
```

### 4.3 验证步骤

| 步骤 | 命令 | 预期结果 |
|------|------|----------|
| 格式检查 | `cd backend && gofmt -l .` | 无输出 |
| 静态分析 | `cd backend && go vet ./...` | 无错误 |
| 编译检查 | `cd backend && go build ./...` | 无错误 |
| 单元测试 | `cd backend && go test -short ./...` | 全部通过 |
| 前端构建 | `cd frontend && npm run build` | 成功 |
| TypeScript | `cd frontend && npx tsc --noEmit` | 无错误 |

---

## 5. 风险评估

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| `go fix` 引入不兼容改动 | 低 | 中 | 在独立分支执行，CI 验证通过后再合并 |
| 依赖库不支持 Go 1.26 | 低 | 高 | `go mod tidy` 会自动处理；如失败则回滚 |
| Docker 镜像拉取失败 | 低 | 高 | `golang:1.26-alpine` 已发布，确认可用 |
| CI 环境 Go 1.26 不可用 | 低 | 高 | `actions/setup-go@v5` 支持 1.26 |

---

## 6. 回滚方案

如升级后出现问题：

```bash
# 回滚 go.mod
cd backend && go mod edit -go=1.24
cd ../tools/doccheck && go mod edit -go=1.24

# 回滚 Dockerfile
git checkout Dockerfile

# 回滚 CI 配置
git checkout .github/workflows/

# 回滚 lint 配置
git checkout .golangci.yml
```

---

## 7. 验收标准

- [ ] `backend/go.mod` 显示 `go 1.26`
- [ ] `tools/doccheck/go.mod` 显示 `go 1.26`
- [ ] `Dockerfile` 使用 `golang:1.26-alpine`
- [ ] CI workflows 使用 `go-version: '1.26'`
- [ ] `.golangci.yml` 使用 `go: "1.26"`
- [ ] 所有文档中的 Go 版本引用更新为 1.26
- [ ] `go vet ./...` 无错误
- [ ] `go build ./...` 成功
- [ ] `go test -short ./...` 全部通过
- [ ] `npm run build` 成功
