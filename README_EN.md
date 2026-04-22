# ETF-Insight (v2.6.0) 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19.2.4-61DAFB)](https://reactjs.org/)
[![Test Coverage](https://img.shields.io/badge/coverage-55%25-yellowgreen)](https://github.com/coder100001/ETF-Insight)

**Open-Source Professional ETF Quantitative Analysis Platform**

ETF-Insight is an open-source ETF analysis platform for professional investors, quantitative researchers, and financial institutions. Built with Go + React stack, providing institutional-grade ETF data insights, multi-dimensional quantitative analysis, and portfolio optimization.

> 🎯 **Vision**: Become the most professional open-source ETF quantitative analysis tool, providing transparent and verifiable analysis capabilities

## 📢 Latest Updates

**v2.6 Data Layer Refactoring & Penetration Analysis**:
- ✅ **Unified Asset Model**: Asset base table supports stock/ETF/index and other asset types
- ✅ **ETF Holdings Penetration**: Underlying holdings details query, weight analysis
- ✅ **Overlap Calculation**: Two ETF holdings overlap analysis (minimum weight method)
- ✅ **Portfolio Penetration Analysis**: Portfolio underlying asset sector/geographic distribution
- ✅ **Concentration Metrics**: Top10/Top20 weight, Herfindahl Index, effective holdings count
- ✅ **Smart Caching**: Overlap calculation result caching (7-day TTL), auto-invalidation
- ✅ **Event-Driven**: Holdings update auto-triggers cache invalidation
- ✅ **Test Coverage**: portfolio 84.9%, event 73.2%

**v2.5 Real-time Data & Quantitative Analysis Upgrade**:
- ✅ **Real-time Data Acquisition**: Finage API integration, 3-year historical data sync (~388 days)
- ✅ **Portfolio Scenario Analysis**: Monte Carlo simulation (1000 runs), three market scenarios (optimistic/neutral/pessimistic), VaR/CVaR risk metrics
- ✅ **Dynamic ETF Selection**: Portfolio Analysis page supports fetching ETF list from API, displaying real-time prices
- ✅ **Flexible Weight Configuration**: Support adding/removing ETFs, real-time weight validation
- ✅ **Financial Calculation Formulas**: Standard financial formulas for portfolio metrics using real historical data
  - Portfolio Variance: $\sigma_p^2 = \sum_i w_i^2 \sigma_i^2 + 2 \sum_{i<j} w_i w_j \sigma_i \sigma_j \rho_{ij}$
  - Portfolio Maximum Drawdown: Calculated based on NAV series, not simple weighted
- ✅ **New Risk-Adjusted Metrics**: Sortino Ratio (downside risk), Calmar Ratio (return/drawdown)
- ✅ **Rolling Window Metrics**: 30/60/90/180/252-day dynamic windows, including win rate, profit/loss ratio
- ✅ **Statistical Metrics**: Skewness, kurtosis analysis of return distribution characteristics
- ✅ **Improved Dividend Reinvestment**: Quarterly/monthly reinvestment models, more accurate compound interest calculation
- ✅ **Default Portfolio Templates**: 6 preset portfolios (conservative/balanced/aggressive/income/dividend growth/tech focused)
- ✅ **Technical Indicators Library**: RSI, MACD, Bollinger Bands, Moving Averages
- ✅ **Risk Models**: VaR/CVaR (historical/parametric methods), portfolio risk analysis
- ✅ **Risk Metrics**: Sharpe Ratio, Sortino Ratio, Maximum Drawdown, Beta/Alpha
- ✅ **Frontend Analysis Pages**: Portfolio scenario analysis, technical analysis (radar chart), risk analysis (VaR visualization)
- ✅ **Test Coverage**: middleware 68.8%, utils 81.2%
- ✅ **CI/CD**: Coverage detection, Codecov integration

**v2.5 Quantitative Engine Enhancement**:
- ✅ **Backtest Engine**: Event-driven architecture, complete order system, slippage/commission models
- ✅ **Portfolio Optimization Enhancement**: Markowitz MPT, Risk Parity, Black-Litterman three models
- ✅ **Factor Analysis Module**: Fama-French 3-factor/5-factor models, attribution analysis
- ✅ **A-Share Data Source**: AKShare/TuShare integration, real-time market data sync
- ✅ **Cross-Asset Coverage**: Equity/Bond/Commodity/REIT/Currency/Multi-Asset full coverage

**v2.4 Security & API Documentation Upgrade**:
- ✅ **JWT Authentication**: Complete authentication middleware, supporting token generation/validation/role control
- ✅ **Audit Logging**: Asynchronous writing, automatic sensitive information masking, Request ID tracking
- ✅ **Data Validation**: Generic validation middleware, supporting string/number/email types
- ✅ **API Pagination**: Generic pagination response structure, supporting page/pageSize parameters
- ✅ **Rate Limiting**: IP-level request frequency limiting
- ✅ **Stock Symbol Validation**: Preventing illegal character injection
- ✅ **Swagger API Documentation**: OpenAPI 3.0 specification, interactive API testing

> **📚 For Developers**: This project uses **AGENTS.md** as the core context document, containing architecture design, data models, coding rules, and other critical information.
>
> 👉 [View AGENTS.md](./AGENTS.md) | 👉 [中文 README](./README.md)

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

### 🔄 Backtest Engine (v2.5 New)
- **Event-Driven Architecture** - High-performance backtest engine based on event bus
- **Order System** - Full support for market/limit/stop-loss/take-profit orders
- **Slippage Models** - Fixed slippage/percentage slippage/volatility slippage
- **Commission Models** - Fixed rate/tiered rate
- **Dividend Reinvestment** - Support automatic dividend reinvestment
- **Rebalancing Strategy** - Periodic rebalancing and threshold rebalancing
- **Backtest Analysis** - Return metrics, risk metrics, trade statistics

### 🎯 Portfolio Optimization (v2.5 Enhanced)
- **Markowitz MPT** - Mean-variance optimization, efficient frontier calculation
- **Risk Parity** - Equal Risk Contribution (ERC), inverse volatility weighting
- **Black-Litterman** - Bayesian view blending, posterior return distribution
- **Weight Constraints** - Support single asset upper/lower bound constraints
- **Optimization Objectives** - Max Sharpe/Min Volatility/Target Return

### 📈 Factor Analysis (v2.5 New)
- **Fama-French Models** - 3-factor/5-factor model support
- **Factor Exposure** - Portfolio exposure analysis on each factor
- **Attribution Analysis** - Return attribution, risk attribution decomposition
- **Active Return Analysis** - Alpha decomposition, excess return sources

### 🌏 A-Share ETF Data Source (v2.5 New)
- **AKShare Integration** - Python AKShare service, real-time A-share ETF data
- **TuShare Support** - Alternative data source, fund NAV data
- **Data Synchronization** - ETF list, prices, historical K-line automatic sync
- **Dividend Tracking** - Dividend yield calculation, dividend frequency statistics

### 🌍 Cross-Asset ETF (v2.5 New)
- **Full Asset Class Coverage** - Equity/Bond/Commodity/REIT/Currency/Multi-Asset/Alternative
- **Global Region Coverage** - US/China/Europe/Japan/Emerging/Asia-Pacific/Latin America
- **Rich ETF Types** - Index/Sector/Factor/Thematic/Active/Leveraged/Inverse
- **Multi-Dimensional Filtering** - Asset class/region/type/sector/expense ratio
- **Portfolio Allocation Suggestions** - Conservative/Balanced/Aggressive/Dividend strategy templates

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
├── AGENTS.md         # Project core context document
├── README.md         # Chinese documentation
└── README_EN.md      # English documentation
```

### Core Documentation
- **[AGENTS.md](./AGENTS.md)** - Project architecture, data models, development standards
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
