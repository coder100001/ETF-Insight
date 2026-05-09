# Design Doc 001: 代码质量分析与改进方案

## 元数据
- **编号**: 001
- **标题**: Code Quality Analysis and Improvement
- **状态**: draft
- **创建日期**: 2026-04-28
- **最后更新**: 2026-04-28
- **关联任务**: 代码质量分析报告及改进建议
- **复杂度级别**: L2

## 1. 背景与动机

### 为什么需要这个分析？
- 项目已发展到 148 个 Go 文件，37,643 行后端代码
- 测试覆盖率不足（28 个测试文件 / 148 个文件 = 19%）
- 前端仅有 1 个测试文件
- 存在 8 处 TODO/FIXME 技术债
- 4 个超过 600 行的"大文件"需要重构

### 当前系统的痛点
1. **测试覆盖率低**: 核心模块缺乏单元测试，风险高
2. **网络配置问题**: `go test` 因 IPv6 超时无法运行
3. **代码复杂度**: 部分文件过大，圈复杂度高
4. **前端类型安全**: 使用 `any` 类型
5. **日志不规范**: `cmd/` 使用 `fmt.Println` 而非日志库

## 2. 调研与现状分析

### 2.1 现有实现
- **后端**: Go 1.26, Gin框架, GORM, 148个文件
- **前端**: React 19, TypeScript 5.9, Vite 8
- **测试**: `go test`, `vitest`
- **代码格式**: `gofmt` 通过，格式统一

### 2.2 业界实践
| 指标 | 业界标准 | 当前状态 | 差距 |
|-----|---------|---------|------|
| 测试覆盖率 | > 80% | ~19% | 🔴 严重不足 |
| 大文件阈值 | < 300 行 | 最大 1,117 行 | 🔴 超标 3.7 倍 |
| Lint 工具 | golangci-lint | 未配置 | 🟡 缺失 |
| 类型安全 | TypeScript strict | 使用 `any` | 🟡 部分缺失 |
| 日志规范 | 结构化日志 | fmt.Println | 🔴 不规范 |

### 2.3 技术约束
- Go 1.26 的 GOPROXY 需要配置（`https://goproxy.cn,direct`）
- 前端 `antd` 包名可能错误（应为 `antd` 而非 `antd`）
- SQLite/PostgreSQL 双数据库支持

## 3. 可选方案

### 方案 A: 渐进式改进（推荐）
- **描述**: 分阶段改进，先解决阻断问题（网络配置），再提升测试覆盖率
- **优点**: 风险低，可持续集成
- **缺点**: 耗时较长
- **工作量**: 中

### 方案 B: 全面重构
- **描述**: 一次性重构所有大文件，补充完整测试
- **优点**: 一步到位
- **缺点**: 风险高，可能影响现有功能
- **工作量**: 大

### 方案 C: 工具先行
- **描述**: 先配置 golangci-lint, pre-commit hooks，强制质量门禁
- **优点**: 防止质量进一步下降
- **缺点**: 存量问题仍需解决
- **工作量**: 小

## 4. 决策
- **选定方案**: 方案 A（渐进式改进）+ 方案 C（工具先行）
- **决策理由**:
  1. 先解决 `go test` 无法运行的问题（配置 GOPROXY）
  2. 配置 golangci-lint 防止新代码质量下降
  3. 分阶段补充单元测试（优先核心模块）
  4. 逐步重构大文件

## 5. 影响范围

### 5.1 后端改进
- `go.mod` / `go.sum` - 可能需添加 golangci-lint 配置
- `.golangci.yml` - 新增 Lint 配置
- `Makefile` - 添加 `make lint`, `make test-coverage` 命令
- 大文件重构: `event_engine.go`, `fama_french.go`, `etf_handler.go`, `portfolio_analytics.go`
- 测试补充: 28 个测试文件 → 目标 80% 覆盖率

### 5.2 前端改进
- `tsconfig.json` - 启用 `strict` 模式
- `eslint.config.js` - 添加 `@typescript-eslint/no-explicit-any` 规则
- 测试配置: `vitest` 已有配置，需补充测试文件

### 5.3 配置文件
- `.env.example` - 添加 `GOPROXY` 配置说明
- `.gitignore` - 添加 `coverage.out`, `lint-report.xml`

## 6. 开放问题

- [ ] 是否将 `antd` 依赖修正为 `antd`？
- [ ] 是否引入 `zap` 或 `logrus` 替代 `fmt.Println`？
- [ ] 重构大文件的优先级排序？
- [ ] 测试覆盖率目标是否设定为 80%？

---

**文档状态**: draft
**下一步**: 创建实施计划（已移至 docs/superpowers/plans/code-quality-fix.md），开始执行改进
