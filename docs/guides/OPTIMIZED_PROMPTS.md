# AI 助手交互指南

**版本**: v1.0
**更新日期**: 2026-04-14
**项目**: ETF-Insight

---

## 🎯 核心原则

| 原则 | 说明 | 示例 |
|------|------|------|
| **明确性** | 清晰表达需求 | ✅ "修复夏普比率计算" vs ❌ "修一下" |
| **完整性** | 提供上下文 | 错误日志、相关代码、文件路径 |
| **可执行性** | 具体可操作 | 预期结果、验收标准 |
| **渐进性** | 复杂问题分步 | 先原理，后实现 |

---

## 📝 提示词模板

### 1. Bug 修复

```markdown
## Bug 修复

### 问题
[描述现象]

### 错误信息
```
[完整错误日志]
```

### 相关代码
- 文件: [路径]
- 函数: [函数名]

### 复现步骤
1. [步骤1]
2. [步骤2]

### 预期
[正确行为]

### 实际
[错误行为]
```

**示例**:
```markdown
## Bug 修复

### 问题
投资组合分析页面收益率显示 NaN

### 错误信息
```
TypeError: Cannot read property 'total_return' of undefined
at PortfolioAnalysis.tsx:232
```

### 相关代码
- 文件: frontend/src/pages/PortfolioAnalysis.tsx
- 函数: calculatePortfolioReturn

### 复现步骤
1. 打开 /portfolio-analysis
2. 选择投资组合
3. 点击分析

### 预期
显示正确收益率

### 实际
显示 NaN，Console 报错
```

---

### 2. 新功能开发

```markdown
## 新功能

### 名称
[功能名]

### 描述
[详细需求]

### 输入
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| param1 | string | 是 | 说明 |

### 输出
| 字段 | 类型 | 说明 |
|------|------|------|
| field1 | number | 说明 |

### 约束
- [约束1]
- [约束2]

### 验收标准
1. [标准1]
2. [标准2]
```

---

### 3. 代码审查

```markdown
## 代码审查

### 范围
- 文件: [路径]
- 函数: [函数名]

### 代码
```go
[代码片段]
```

### 关注点
- [关注点1]
- [关注点2]
```

---

## 🔧 常用场景

### 启动项目
```markdown
启动项目:
- 前端: http://localhost:3000
- 后端: http://localhost:8080
```

### 代码修改
```markdown
修改 CalculateMetrics:
- symbol 不能为空验证
- 日期范围验证
- 返回标准 API 格式
```

### 测试验证
```markdown
测试 GetETFList 分页:
- page=1&pageSize=10 → 前10条
- page=2&pageSize=10 → 11-20条
- pageSize>100 → 错误
```

### 提交代码
```markdown
提交修改:
- 信息: feat(portfolio): 添加优化API
- 文件: handlers/portfolio.go, services/portfolio.go
```

---

## ⚠️ 常见错误

### ❌ 上下文不足
```markdown
# 错误
修一下这个bug

# 正确
修复 services/etf_analysis.go 中 CalculateMetrics 的
除零错误。当 volatility=0 时返回 0。
```

### ❌ 期望不明确
```markdown
# 错误
优化性能

# 正确
优化 CalculateMetrics:
- 当前: 500ms
- 目标: <100ms
- 方案: 添加缓存
```

### ❌ 多任务混杂
```markdown
# 错误
项目跑不起来，优化代码，更新文档

# 正确
分三个请求:
1. [Bug] 启动失败，port already in use
2. [优化] CalculateMetrics 性能
3. [文档] 更新 agents.md
```

---

## ✅ 检查清单

### 提交前
- [ ] 需求描述清晰
- [ ] 提供错误信息
- [ ] 指定文件路径
- [ ] 说明预期行为

### 响应后
- [ ] 理解响应内容
- [ ] 执行建议操作
- [ ] 验证修复结果
- [ ] 测试通过

---

## 🎓 高级技巧

### 链式思考
```markdown
分析问题:
1. 可能原因
2. 解决方案
3. 实施步骤
```

### 边界测试
```markdown
补充边界测试:
- 零值输入
- 极端值
- 空数据
```

---

**维护者**: ETF-Insight Team
