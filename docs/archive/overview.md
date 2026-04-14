# 项目架构概述

**版本**: v2.4
**更新日期**: 2026-04-14
**状态**: 已归档（历史文档）

---

## 📁 项目结构

```
py_project/
├── backend/              # Go 后端服务
│   ├── config/          # 配置管理
│   ├── docs/            # Swagger API 文档
│   ├── handlers/        # HTTP 处理器
│   ├── middleware/      # 中间件
│   │   ├── auth.go      # JWT 认证
│   │   ├── audit.go     # 审计日志
│   │   ├── ratelimit.go # 速率限制
│   │   ├── validation.go # 数据验证
│   │   └── security.go  # 安全头
│   ├── models/          # GORM 数据模型
│   ├── services/        # 业务服务
│   │   ├── etf_analysis.go      # ETF 分析
│   │   ├── portfolio_optimizer.go # 组合优化
│   │   ├── exchange_rate/       # 汇率服务
│   │   └── datasource/          # 数据源管理
│   └── main.go
├── frontend/            # React 前端
│   └── src/
│       ├── components/  # 组件
│       ├── pages/       # 页面
│       └── services/    # API 服务
└── docs/                # 项目文档
    ├── security/        # 安全文档
    ├── reviews/         # 审查报告
    ├── roadmap/         # 路线图
    └── guides/          # 使用指南
```

---

## 🚀 核心功能

### 后端服务

| 模块 | 功能 | 文件 |
|------|------|------|
| **ETF 分析** | 收益率、波动率、最大回撤 | `services/etf_analysis.go` |
| **投资组合** | 马科维茨优化、有效前沿 | `services/portfolio_optimizer.go` |
| **汇率服务** | 多数据源故障转移 | `services/exchange_rate/` |
| **安全** | JWT、审计、限流 | `middleware/` |
| **API 文档** | Swagger/OpenAPI 3.0 | `docs/` |

### 前端应用

| 模块 | 功能 |
|------|------|
| **ETF 分析** | 列表、详情、对比 |
| **投资组合** | 配置、分析、优化 |
| **A股红利** | ETF价格、分红计算 |
| **汇率** | 实时汇率查询 |

---

## 🔒 安全架构

```
请求流程:
Client → Rate Limiter → CORS → Security Headers → Auth → Audit Log → Handler
```

**已实现**:
- ✅ JWT 身份认证
- ✅ 审计日志（自动脱敏）
- ✅ 速率限制（滑动窗口）
- ✅ 数据验证
- ✅ CORS 配置
- ✅ 安全响应头

---

## 📚 文档索引

| 文档 | 路径 | 说明 |
|------|------|------|
| API 文档 | `/swagger` | OpenAPI 3.0 |
| 安全改进 | `docs/security/` | 安全功能说明 |
| 代码审查 | `docs/reviews/` | 审查报告 |
| 路线图 | `docs/roadmap/` | 演进规划 |
| 使用指南 | `docs/guides/` | 交互指南 |

---

## 🏛️ 开源定位

> 坚持**开源、专业、透明**，打造学术界和业界认可的 ETF 量化分析基础设施

- 🔓 **完全开源**: MIT 协议
- 📊 **专业分析**: 机构级量化指标
- 🔧 **可扩展**: 插件化架构
- 🤝 **社区驱动**: 欢迎贡献

---

**注意**: 本文档为历史归档文档，最新信息请参考 [README.md](../../README.md) 和 [agents.md](../../agents.md)
