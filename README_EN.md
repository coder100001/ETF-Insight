# ETF-Insight

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26-blue)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19.2.4-61DAFB)](https://reactjs.org/)

**Open-source professional ETF quantitative analysis platform** — for professional investors, quantitative researchers, and financial institutions. Institutional-grade ETF data insights, multi-dimensional quantitative analysis, and portfolio optimization.

> Vision: Become the most professional open-source ETF quantitative analysis tool, providing transparent and verifiable analysis capabilities

## Core Features

| Module | Capabilities |
|--------|-------------|
| **Portfolio Analysis** | Monte Carlo simulation, scenario analysis, VaR/CVaR, Sharpe/Sortino/Calmar ratios |
| **Portfolio Optimization** | Markowitz MPT, Risk Parity, Black-Litterman, efficient frontier |
| **Factor Analysis** | Fama-French 3/5 factor models, attribution analysis |
| **Backtest Engine** | Event-driven architecture, order system, slippage/commission models, dividend reinvestment |
| **Technical Indicators** | RSI, MACD, Bollinger Bands, moving averages |
| **ETF Comparison** | Side-by-side multi-dimensional comparison, holdings overlap analysis |
| **Holdings Penetration** | Underlying asset details, sector/region/market cap distribution, concentration metrics |
| **QuantLib Integration** | Options pricing, yield curves, bond pricing, VaR calculation |
| **AI Agents** | 4 financial agents (Buffett/Graham/Bridgewater/Macro), team debate mode |
| **Data Sources** | Finage (ETF), AKShare/TuShare (A-Share), multi-source exchange rate failover |

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.26, Gin, GORM (SQLite/PostgreSQL) |
| **Frontend** | React 19, TypeScript, Vite, Ant Design, ECharts |
| **Microservices** | Python FastAPI (AI Agent port 8091, Data port 8092, Analytics port 8093) |
| **Data Sources** | Finage API, Open Exchange Rates, AKShare |

## Quick Start

```bash
git clone <repository-url>
cd py_project

# One-click startup
./start.sh        # macOS/Linux
start.bat         # Windows
```

**Manual startup:**

```bash
# Backend
cd backend && go mod tidy && go run main.go

# Frontend
cd frontend && npm install && npm run dev
```

**Requirements:** Go >= 1.26, Node.js >= 18, SQLite 3.35+

**Data source config:** Set `FINAGE_API_KEY` environment variable (see [.env.example](./backend/.env.example))

## Project Structure

```
py_project/
├── backend/              # Go backend service
├── frontend/             # React frontend application
├── services/
│   ├── agent/            # AI Agent microservice (port 8091)
│   ├── data/             # Data source microservice (port 8092)
│   └── analytics/        # Analytics microservice (port 8093)
├── tools/doccheck/       # Documentation consistency checker
├── AGENTS.md             # Project core context document
└── CHANGELOG.md          # Version changelog
```

## Documentation Index

| Document | Description |
|----------|-------------|
| [AGENTS.md](./AGENTS.md) | Architecture, data models, coding standards, financial algorithms (**required reading**) |
| [CHANGELOG.md](./CHANGELOG.md) | Version change history |
| [README.md](./README.md) | 中文文档 |
| [docs/](./docs/) | Documentation index | [docs/reference/](./docs/reference/) | Reference docs (roadmap, dev guides) |
| [docs/superpowers/](./docs/superpowers/) | Design docs & implementation plans | [docs/archive/](./docs/archive/) | Archived docs (reviews, security) |
| API Docs | `http://localhost:8080/swagger` (interactive) |

## Contributing

Welcome to Issues and Pull Requests. Please ensure:
1. Follow code standards (Go: `gofmt`, TypeScript: strict types)
2. New features include unit tests (coverage >= 80%)
3. Update relevant documentation

See [AGENTS.md Contributing Guide](./AGENTS.md#-贡献指南).

## License

[MIT](./LICENSE)
