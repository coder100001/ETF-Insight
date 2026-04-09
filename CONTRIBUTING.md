# Contributing to ETF-Insight

感谢您对 ETF-Insight 项目的关注！我们欢迎所有形式的贡献，包括但不限于代码贡献、文档改进、问题报告和功能建议。

## 目录

1. [行为准则](#行为准则)
2. [环境搭建](#环境搭建)
3. [开发流程](#开发流程)
4. [代码规范](#代码规范)
5. [提交信息格式](#提交信息格式)
6. [Pull Request 流程](#pull-request-流程)
7. [代码审查标准](#代码审查标准)

---

## 行为准则

### 我们的承诺

为了营造开放和友好的社区环境，我们承诺：
- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 关注对社区最有利的事情
- 对其他社区成员表示同理心

### 不可接受的行为

- 使用性化的语言或图像
- 恶意评论或人身攻击
- 公开或私下骚扰
- 未经许可发布他人的私人信息
- 其他不专业或不恰当的行为

---

## 环境搭建

### 前置要求

在开始贡献之前，请确保您已安装以下工具：

| 工具 | 版本要求 | 说明 |
|------|---------|------|
| Go | >= 1.21 | 后端开发语言 |
| Node.js | >= 18.x | 前端开发语言 |
| npm | >= 9.x | Node.js 包管理器 |
| Git | 最新版 | 版本控制 |
| SQLite3 | 最新版 | 本地数据库（可选）|

### 克隆仓库

```bash
# Fork 项目到您的 GitHub 账户
# 然后克隆您的 fork
git clone https://github.com/YOUR_USERNAME/ETF-Insight.git
cd ETF-Insight

# 添加上游仓库
git remote add upstream https://github.com/coder100001/ETF-Insight.git
```

### 安装依赖

#### 后端

```bash
cd backend
go mod download
go vet ./...
```

#### 前端

```bash
cd frontend
npm install
npm run lint
```

### 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，填入您的配置
# 特别是 FINAGE_API_KEY
```

### 验证环境

```bash
# 启动后端（新终端）
cd backend
go run main.go

# 启动前端（新终端）
cd frontend
npm run dev
```

访问以下地址验证：
- 前端: http://localhost:5173
- 后端 API: http://localhost:8080/health

---

## 开发流程

### 1. 同步上游代码

在开始开发前，请确保您的代码与上游仓库同步：

```bash
git checkout main
git fetch upstream
git merge upstream/main
```

### 2. 创建功能分支

```bash
# 使用清晰的分支名称
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
# 或
git checkout -b docs/your-doc-update
```

### 3. 开发代码

#### 后端开发

- 所有 Go 代码必须通过 `gofmt` 格式化
- 所有 Go 代码必须通过 `golangci-lint` 检查
- 添加或修改代码时，请同步更新 `agents.md` 中的相关章节
- 为新功能编写单元测试

```bash
# 格式化代码
cd backend
gofmt -w .

# 运行代码检查
golangci-lint run

# 运行测试
go test ./... -v
```

#### 前端开发

- 所有 TypeScript/JavaScript 代码必须通过 ESLint 检查
- 遵循 TypeScript 类型安全原则，避免使用 `any` 类型
- 遵循 React Hooks 最佳实践
- 使用 Ant Design 组件库

```bash
# 运行代码检查
cd frontend
npm run lint

# 运行类型检查
npm run type-check

# 启动开发服务器
npm run dev
```

### 4. 运行测试

#### 后端测试

```bash
cd backend
go test ./... -v -cover
```

#### 前端测试

```bash
cd frontend
npm run test
```

### 5. 提交代码

在提交前，请确保：
- 代码已格式化
- 所有测试通过
- 没有 lint 错误
- 更新了相关文档（如 `agents.md`）

---

## 代码规范

### Go 代码规范

#### 格式化

使用 `gofmt` 统一格式化：

```bash
cd backend
gofmt -w .
```

#### 导入顺序

导入顺序应为：
1. 标准库
2. 第三方库
3. 本地包

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "gorm.io/gorm"

    "etf-insight/models"
    "etf-insight/services"
    "etf-insight/utils"
)
```

#### 错误处理

- 所有错误都应该被处理
- 使用 `fmt.Errorf` 包装错误
- 使用项目标准错误（`services/datasource/errors.go`）

```go
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}
```

#### 命名规范

- 包名：小写，简洁
- 导出函数/类型：大驼峰
- 私有函数/类型：小驼峰
- 常量：大写下划线分隔

### TypeScript/React 代码规范

#### 类型安全

- 避免使用 `any` 类型
- 为所有 props、state、API 响应定义类型
- 使用 TypeScript 严格模式

```typescript
// Good
interface ETFData {
    symbol: string;
    name: string;
    currentPrice: number;
}

// Bad
const data: any = await fetchData();
```

#### React Hooks

- 只在组件顶层调用 Hooks
- 使用 ESLint 检查 `exhaustive-deps`
- 为仅挂载执行的 effect 添加注释

```typescript
useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
}, []);
```

#### 组件结构

- 函数式组件优先
- 使用 TypeScript 定义组件 props
- 组件文件使用 `.tsx` 扩展名

---

## 提交信息格式

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型

| 类型 | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复 bug |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响代码运行） |
| `refactor` | 重构（既不是新增功能，也不是修改bug） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `build` | 构建系统或外部依赖变更 |
| `ci` | CI/CD 相关 |
| `chore` | 其他不修改 src 或 test 的更改 |
| `revert` | 回滚提交 |

### 示例

```
feat(etf): 添加持仓重叠分析功能

- 实现 ETF 间持仓重叠计算
- 添加可视化图表展示重叠比例
- 新增 API 端点 /api/etf/overlap

Closes #123
```

```
fix(analysis): 修复资本利得计算错误

- 从数据库获取真实价格而非硬编码
- 添加最小数据点验证
- 更新相关测试

Fixes #456
```

```
docs(readme): 更新 CONTRIBUTING.md

- 添加环境搭建说明
- 补充代码规范
- 添加 PR 流程指南
```

---

## Pull Request 流程

### 1. 准备 PR

在创建 PR 之前，请确保：
- [ ] 您的代码与上游 `main` 分支同步
- [ ] 所有测试通过
- [ ] 没有 lint 错误
- [ ] 更新了相关文档（如 `agents.md`）
- [ ] 添加了必要的测试
- [ ] 提交信息符合规范

### 2. 创建 PR

1. 访问您的 fork 仓库
2. 点击 "Compare & pull request"
3. 填写 PR 模板（如下）
4. 创建 PR

### 3. PR 模板

```markdown
## 变更类型

- [ ] Bug 修复
- [ ] 新功能
- [ ] 性能优化
- [ ] 文档更新
- [ ] 代码重构
- [ ] 其他（请说明）

## 变更描述

请简要描述您的变更内容。

## 相关 Issue

Closes #<issue_number>

## 测试

- [ ] 后端测试通过
- [ ] 前端测试通过
- [ ] 手动测试完成
- [ ] 添加了新测试

## 检查清单

- [ ] 代码已格式化（gofmt / eslint）
- [ ] 所有 lint 检查通过
- [ ] 已更新相关文档（agents.md）
- [ ] 提交信息符合 Conventional Commits 规范
- [ ] 没有引入安全漏洞

## 截图（如适用）

如果有 UI 变更，请附上截图。
```

### 4. 代码审查

PR 创建后，项目维护者将进行代码审查。请：
- 及时回应审查意见
- 根据反馈进行修改
- 修改后追加提交，不要强制推送

---

## 代码审查标准

### 必须通过的检查

- [ ] 所有 CI 检查通过（测试、lint、类型检查）
- [ ] 代码符合项目编码规范
- [ ] 变更范围合理，一次只做一件事
- [ ] 没有引入安全漏洞
- [ ] 文档已同步更新

### 代码质量标准

| 维度 | 标准 |
|------|------|
| 可读性 | 代码清晰，有必要的注释 |
| 可维护性 | 遵循单一职责原则，易于修改 |
| 可测试性 | 代码易于测试，有单元测试覆盖 |
| 性能 | 没有明显的性能问题 |
| 安全性 | 没有安全漏洞（SQL注入、XSS等） |

---

## 其他资源

- [README.md](./README.md) - 项目介绍和快速开始
- [README_EN.md](./README_EN.md) - 英文文档
- [agents.md](./agents.md) - 项目核心上下文（必读！）
- [.golangci.yml](./.golangci.yml) - Go 代码检查配置
- [.pre-commit-config.yaml](./.pre-commit-config.yaml) - Pre-commit 钩子配置

---

## 获得帮助

如果您有任何问题或需要帮助：
- 查看 [agents.md](./agents.md) 了解项目上下文
- 查看已有 [Issues](../../issues) 是否有相关问题
- 创建新的 Issue 提问
- 在 PR 中 @ 项目维护者

---

## 许可证

通过贡献代码，您同意您的贡献将根据项目的 [LICENSE](./LICENSE) 文件进行许可。

---

感谢您为 ETF-Insight 做出贡献！🎉
