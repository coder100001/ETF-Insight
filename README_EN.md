# ETF-Insight (v2.3.2)

A professional ETF analysis and comparison platform, benchmarking against international tools like Trackinsight and ETF Insider. Built with Go + React stack, providing in-depth ETF data insights, multi-dimensional comparison analysis, holdings decomposition, risk assessment, and portfolio optimization.

**v2.3 Update**: Comprehensive exchange rate service upgrade - multi-data source failover, race condition fixes, performance optimization.

**v2.2 Update**: Comprehensive code quality optimization - Fixed all ESLint issues, achieved TypeScript type safety, unified code style.

**v2.1 Update**: Fixed dividend yield display and capital gain calculations, optimized data accuracy.

**v2.0 Architecture Update**: Fully rely on Finage real data, removed all hard-coded mock data, all fields must be persisted.

**v2.3 Exchange Rate Service Upgrade**:
- ✅ **Multi-data source failover**: Support for Open Exchange Rates, CurrencyAPI, Frankfurter three data sources
- ✅ **Race condition fixes**: Resolved concurrent access issues in data source manager
- ✅ **Health monitoring**: Automatic data source availability checks and failover switching
- ✅ **Performance optimization**: Smart caching strategies, database query optimization

**v2.3.2 Pre-commit Hooks**:
- ✅ **TypeScript type check**: Automatic check before commit
- ✅ **ESLint check**: Frontend code quality assurance
- ✅ **Fail blocks commit**: Syntax errors cannot be committed, avoiding CI build failures

> **📚 For Developers**: This project uses **agents.md** as the core context document, containing architecture design, data models, coding rules, and other critical information.
>
> 👉 [View agents.md](./agents.md) | 👉 [中文 README](./README.md)

---

## 🎯 Product Positioning

ETF-Insight aims to become a powerful ETF analysis tool for professional investors and institutional users:

- **ETF Comparison Analysis** - Side-by-side multi-dimensional comparison to discover optimal investment targets
- **Holdings Deep Decomposition** - Penetrate underlying assets to understand real risk exposure
- **Risk Indicator Assessment** - Professional metrics including volatility, Sharpe ratio, maximum drawdown, Beta
- **Portfolio Optimization** - Build optimal asset allocation based on Modern Portfolio Theory

---

## ✨ Core Features

### 📊 ETF Comparison Analysis
- **Side-by-Side Comparison** - Support up to 5 ETFs simultaneously
- **Multi-Dimensional Metrics** - Expense ratio, AUM, dividend yield, performance, risk indicators
- **Smart Dividend Yield** - Automatically set reasonable dividend yields based on ETF type (High Dividend 3.5%, Covered Call 7%, Bonds 4%)
- **Holdings Overlap Analysis** - Identify holdings overlap between ETFs to avoid over-concentration
- **Performance Backtest Comparison** - Compare returns across different time periods

### 🔍 Holdings Deep Decomposition
- **Top 10 Holdings** - Penetrate underlying assets to understand core holdings
- **Sector Distribution** - Sector weight distribution and trend analysis
- **Regional Distribution** - Country/region allocation ratios
- **Market Cap Distribution** - Large/mid/small-cap allocation ratios
- **Style Analysis** - Value/growth style exposure

### 💼 A-Share Dividend ETF Portfolio
- **A-Share ETF Management** - Support mainstream dividend ETFs like CSI Dividend, Dividend Low Volatility
- **Investment Allocation Distribution** - Pie chart visualization of portfolio allocation
- **Dividend Data Tracking** - Key indicators including dividend yield and dividend frequency

### 💱 Exchange Rate Data Management
- **Real-Time Exchange Rates** - Major currency pairs like USD/CNY, USD/HKD
- **Multi-Data Source Support** - Open Exchange Rates, CurrencyAPI, Frankfurter three data sources
- **Automatic Failover** - Automatic switch to backup data sources when primary is unavailable
- **Health Monitoring** - Real-time monitoring of data source availability
- **Automatic Synchronization** - Scheduled tasks automatically update exchange rate data (every 5 minutes)
- **Currency Conversion** - Support for conversion between multiple currencies
- **Sync Logging** - Complete exchange rate synchronization batch records and detailed tracking

### ⚙️ ETF Configuration Management
- **CRUD Operations** - Create, read, update, delete ETF configuration information
- **Status Management** - Enable/disable automatic ETF data updates
- **Data Source Configuration** - **Finage as the only real data source** (v2.0 architecture)

### 📈 Portfolio Configuration
- **Portfolio Construction** - Custom investment portfolios and weight allocation
- **Return Analysis** - Portfolio return simulation based on historical data
- **Capital Gain Calculation** - Calculate capital gains and returns based on real historical data
- **Preset Portfolios** - Built-in various investment strategy portfolio templates

---

## 🛠️ Technology Stack

### Backend (Go)
| Technology | Version | Purpose |
|------------|---------|---------|
| Go | >= 1.21 | Core language |
| Gin | v1.12.0 | Web framework |
| GORM | v1.30.0 | ORM framework (SQLite/PostgreSQL) |
| go-cache | v2.1.0 | In-memory caching |
| cron/v3 | v3.0.1 | Scheduled task scheduling |

### Frontend (React)
| Technology | Version | Purpose |
|------------|---------|---------|
| React | ^19.2.4 | UI framework |
| TypeScript | ^5.x | Type safety |
| Vite | latest | Build tool |
| Ant Design | ^6.3.4 | UI component library |
| ECharts | ^6.0.0 | Data visualization |
| Recharts | ^3.8.1 | Chart components |
| React Router | ^7.13.2 | Routing management |

### Data Storage
- **SQLite** - Default local database (development environment)
- **PostgreSQL** - Production database support

---

## 🚀 Quick Start

### Method 1: One-Click Startup (Recommended)

```bash
# Clone the project
git clone <repository-url>
cd py_project

# One-click startup (macOS/Linux)
./start.sh

# One-click startup (Windows)
start.bat
```

### Method 2: Manual Startup

#### 1. Backend Service Startup
```bash
cd backend

# Install dependencies
go mod tidy

# Configure environment variables
cp .env.example .env
# Edit .env file, configure Finage API Key

# Start backend service
go run main.go
```

#### 2. Frontend Service Startup
```bash
cd frontend

# Install dependencies
npm install

# Start frontend service
npm run dev
```

### Environment Requirements
- **Go**: >= 1.21
- **Node.js**: >= 18.0.0
- **npm**: >= 9.0.0
- **SQLite**: 3.35.0+

---

## 📊 Data Source Configuration

### ETF Data Source
- **Primary Data Source**: Finage API (must be configured)
- **Environment Variable**: `FINAGE_API_KEY=your_api_key_here`

### Exchange Rate Data Sources
- **Primary Data Source**: Open Exchange Rates
- **Backup Data Sources**: CurrencyAPI, Frankfurter
- **Failover**: Automatic switching, no manual configuration required

---

## 🔧 Development Guide

### Code Standards
- **Go**: Follow official code standards, use `gofmt` for formatting
- **TypeScript**: Strict type checking, disable `any` type
- **React**: Functional components, proper Hooks usage

### Project Structure
```
ETF-Insight/
├── backend/          # Go backend service
├── frontend/         # React frontend application
├── agents.md         # Project core context document
├── README.md         # Chinese documentation
└── README_EN.md      # English documentation
```

### Core Documentation
- **[agents.md](./agents.md)** - Project architecture, data models, development standards
- **[docs/openapi.yaml](./docs/openapi.yaml)** - API interface documentation

---

## 🎯 Evolution Roadmap

### v2.3 (Current Version)
- ✅ Exchange rate service multi-data source failover
- ✅ Race condition fixes and performance optimization
- ✅ Comprehensive code quality optimization

### v2.5 (Planned)
- 🔄 Intelligent analysis engine development
- 🔄 Backtesting framework integration
- 🔄 Microservices architecture preparation

### v3.0 (Long-term Planning)
- 🔄 Cloud-native transformation
- 🔄 Open platform development
- 🔄 Business model exploration

---

## 📞 Technical Support

### Common Issues
1. **Data source connection failure**: Check network connection and API Key configuration
2. **Exchange rate data inconsistency**: System automatically fails over, check logs to confirm current data source
3. **Performance issues**: Enable Redis cache to improve performance

### Log Viewing
```bash
# View backend logs
tail -f backend/logs/app.log

# View exchange rate sync logs
tail -f backend/logs/exchange_rate.log
```

---

## 🤝 Contribution Guidelines

Welcome to submit Issues and Pull Requests! Please ensure:
1. Follow project code standards
2. New features include unit tests
3. Update relevant documentation
4. Pass code review

---

## 📄 License

This project is licensed under the MIT License.

---

**Experience Now**: [http://localhost:8080](http://localhost:8080)
**API Documentation**: [http://localhost:8080/docs](http://localhost:8080/docs)
