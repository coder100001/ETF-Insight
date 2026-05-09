# ETF-Insight

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26-blue)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19.2.4-61DAFB)](https://reactjs.org/)

**开源专业的 ETF 量化分析平台** — 面向专业投资者、量化研究员和金融机构，提供机构级 ETF 数据洞察、多维度量化分析、投资组合优化。

> 愿景：成为开源社区最专业的 ETF 量化分析工具，为投资者提供透明、可验证的分析能力

## 核心特性

| 模块 | 能力 |
|------|------|
| **投资组合分析** | 蒙特卡洛模拟、情景分析、VaR/CVaR、夏普/索提诺/卡尔玛比率 |
| **组合优化** | 马科维茨 MPT、风险平价、Black-Litterman、有效前沿 |
| **因子分析** | Fama-French 三因子/五因子模型、归因分析 |
| **回测引擎** | 事件驱动架构、订单系统、滑点/手续费模型、分红再投资 |
| **技术指标** | RSI、MACD、布林带、移动平均线 |
| **ETF 对比** | 多维度并排对比、持仓重叠分析、业绩回测 |
| **持仓穿透** | 底层资产明细、行业/地区/市值分布、集中度指标 |
| **QuantLib 集成** | 期权定价、收益率曲线、债券定价、VaR 计算 |
| **AI Agent** | 4 个金融 Agent（Buffett/Graham/Bridgewater/Macro）、团队辩论模式 |
| **数据源** | Finage（ETF）、AKShare/TuShare（A 股）、多数据源汇率故障转移 |

## 技术栈

| 层级 | 技术 |
|------|------|
| **后端** | Go 1.26、Gin、GORM（SQLite/PostgreSQL） |
| **前端** | React 19、TypeScript、Vite、Ant Design、ECharts |
| **微服务** | Python FastAPI（AI Agent port 8091、数据源 port 8092、分析 port 8093） |
| **数据源** | Finage API、Open Exchange Rates、AKShare |

## 快速开始

```bash
git clone <repository-url>
cd py_project

# 一键启动
./start.sh        # macOS/Linux
start.bat         # Windows
```

**手动启动：**

```bash
# 后端
cd backend && go mod tidy && go run main.go

# 前端
cd frontend && npm install && npm run dev
```

**环境要求：** Go >= 1.26、Node.js >= 18、SQLite 3.35+

**数据源配置：** 设置 `FINAGE_API_KEY` 环境变量（[.env.example](./backend/.env.example)）

## 项目结构

```
py_project/
├── backend/              # Go 后端服务
├── frontend/             # React 前端应用
├── services/
│   ├── agent/            # AI Agent 微服务 (port 8091)
│   ├── data/             # 数据源微服务 (port 8092)
│   └── analytics/        # 分析微服务 (port 8093)
├── tools/doccheck/       # 文档一致性检查工具
├── AGENTS.md             # 项目核心上下文文档
└── CHANGELOG.md          # 版本变更日志
```

## 文档索引

| 文档 | 说明 |
|------|------|
| [AGENTS.md](./AGENTS.md) | 架构设计、数据模型、编码规范、金融算法标准（**开发者必读**） |
| [CHANGELOG.md](./CHANGELOG.md) | 版本变更记录 |
| [README_EN.md](./README_EN.md) | English documentation |
| [docs/](./docs/) | 文档入口与索引 | [docs/reference/](./docs/reference/) | 参考文档（路线图、开发指南等） |
| [docs/superpowers/](./docs/superpowers/) | 设计文档与实施计划 | [docs/archive/](./docs/archive/) | 归档文档（历史审查、安全文档） |
| API 文档 | `http://localhost:8080/swagger`（交互式） |

## 贡献

欢迎 Issue 和 Pull Request。请确保：
1. 遵循代码规范（Go: `gofmt`，TypeScript: 严格类型）
2. 新功能包含单元测试（覆盖率 >= 80%）
3. 更新相关文档

详见 [AGENTS.md 贡献指南](./AGENTS.md#-贡献指南)。

## 许可证

[MIT](./LICENSE)
