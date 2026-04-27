# Trae IDE 项目规则 - ETF-Insight

## 开发工作流（类似 gstack）

### 1. 需求分析阶段
**触发**: 用户提出新功能需求

**必须执行**:
- [ ] 阅读 AGENTS.md 了解项目背景
- [ ] 查看相关代码文件和测试
- [ ] 询问用户具体需求和预期行为
- [ ] 确认技术可行性

**输出**: 需求理解摘要

---

### 2. 规划阶段
**触发**: 需求明确后

**必须执行**:
- [ ] 创建或更新 TODO 列表
- [ ] 设计数据模型（如需要）
- [ ] 设计 API 接口（如需要）
- [ ] 规划测试用例

**输出**: 实施计划

---

### 3. 编码阶段
**触发**: 规划完成后

**必须执行**:
- [ ] 先写测试（TDD）
- [ ] 实现功能代码
- [ ] 运行测试验证
- [ ] 代码自检（见下方检查清单）

**代码检查清单**:
- [ ] 错误处理完善
- [ ] 使用 decimal 处理金额
- [ ] 添加必要注释
- [ ] 遵循命名规范

---

### 4. 审查阶段
**触发**: 编码完成后

**必须执行**:
- [ ] 运行所有测试
- [ ] 运行 lint 检查
- [ ] 检查类型错误
- [ ] 验证边界条件

**Go 后端检查**:
```bash
cd backend
go test ./...
go fmt ./...
```

**前端检查**:
```bash
cd frontend
npm run lint
npm run typecheck
```

---

### 5. 提交阶段
**触发**: 审查通过后

**必须执行**:
- [ ] 遵循 conventional commits 格式
- [ ] 提交信息包含变更摘要
- [ ] 关联相关 issue（如适用）

**提交格式**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**:
- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `test`: 测试
- `refactor`: 重构
- `perf`: 性能优化

---

## 常用命令速查

### 后端开发
```bash
# 运行测试
cd backend && go test ./... -v

# 运行特定测试
cd backend && go test ./services/... -v -run TestRiskBudget

# 格式化代码
cd backend && go fmt ./...

# 构建
cd backend && go build ./...
```

### 前端开发
```bash
# 安装依赖
cd frontend && npm install

# 启动开发服务器
cd frontend && npm run dev

# 运行 lint
cd frontend && npm run lint

# 类型检查
cd frontend && npm run typecheck

# 构建
cd frontend && npm run build
```

### 数据库
```bash
# 后端目录下运行迁移
cd backend
go run cmd/migrate/main.go
```

---

## 项目结构速查

```
ETF-Insight/
├── backend/           # Go 后端
│   ├── services/      # 业务逻辑
│   ├── handlers/      # HTTP 处理器
│   ├── models/        # 数据模型
│   └── migrations/    # 数据库迁移
├── frontend/          # React 前端
│   ├── src/pages/     # 页面组件
│   ├── src/services/  # API 服务
│   └── src/types/     # TypeScript 类型
└── docs/              # 文档
```

---

## 关键规范

### 错误处理
- 所有错误必须处理，不能忽略
- 使用 `fmt.Errorf("context: %w", err)` 包装错误
- 返回有意义的错误信息

### 金额计算
- 使用 `decimal.Decimal` 类型
- 避免浮点数精度问题
- 百分比和小数要明确区分

### API 设计
- RESTful 风格
- 统一响应格式: `{ success, data, error }`
- 使用适当的 HTTP 状态码

### 测试要求
- 新功能必须有单元测试
- 测试覆盖率目标 > 80%
- 包含边界条件测试

---

## 参考资料

- [AGENTS.md](../AGENTS.md) - 项目完整文档
- [API 文档](../docs/openapi.yaml) - OpenAPI 规范
- [README](../../README.md) - 项目介绍
