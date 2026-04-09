# ETF-Insight (v2.2)

A professional ETF analysis and comparison platform, benchmarking against international tools like Trackinsight and ETF Insider. Built with Go + React stack, providing in-depth ETF data insights, multi-dimensional comparison analysis, holdings decomposition, risk assessment, and portfolio optimization.

**v2.2 Update**: Comprehensive code quality optimization - Fixed all ESLint issues, achieved TypeScript type safety, unified code style.

**v2.1 Update**: Fixed dividend yield display and capital gain calculations, optimized data accuracy.

**v2.0 Architecture Update**: Fully rely on Finage real data, removed all hard-coded mock data, all fields must be persisted.

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
- **Auto Sync** - Scheduled tasks automatically update exchange rate data (every 5 minutes)
- **Currency Conversion** - Support conversion between multiple currencies
- **Sync Logs** - Complete exchange rate sync batch records and detailed tracking

### ⚙️ ETF Configuration Management
- **CRUD Operations** - Create, read, update, delete ETF configuration information
- **Status Management** - Enable/disable ETF data auto-update
- **Data Source Configuration** - **Finage as the only real data source** (v2.0 architecture)

### 📈 Portfolio Configuration
- **Portfolio Construction** - Custom portfolio and weight allocation
- **Return Analysis** - Portfolio return simulation based on historical data
- **Capital Gain Calculation** - Calculate capital gains and returns based on real historical data
- **Preset Portfolios** - Built-in multiple investment strategy portfolio templates

---

## 🛠️ Tech Stack

### Backend (Go)
| Technology | Version | Purpose |
|------------|---------|---------|
| Go | >= 1.21 | Core language |
| Gin | v1.12.0 | Web framework |
| GORM | v1.30.0 | ORM framework (SQLite/PostgreSQL) |
| go-cache | v2.1.0 | In-memory cache |
| cron/v3 | v3.0.1 | Scheduled task scheduler |

### Frontend (React)
| Technology | Version | Purpose |
|------------|---------|---------|
| React | ^19.2.4 | UI framework |
| TypeScript | ^5.x | Type safety |
| Vite | latest | Build tool |
| Ant Design | ^6.3.4 | UI component library |
| ECharts | ^6.0.0 | Data visualization |
| Recharts | ^3.8.1 | Chart components |
| React Router | ^7.13.2 | Route management |

### Data Storage
- **SQLite** - Default local database (development environment)
- **PostgreSQL** - Production database support

---

## 🚀 Quick Start

### Method 1: One-Click Start (Recommended)

```bash
# Clone project
git clone https://github.com/coder100001/ETF-Insight.git
cd ETF-Insight

# macOS / Linux
chmod +x start.sh
./start.sh

# Windows
start.bat
```

The startup script will automatically:
1. ✅ Check runtime environment (Go, Node.js)
2. ✅ Install backend dependencies (go mod download)
3. ✅ Compile backend project
4. ✅ Install frontend dependencies (npm install)
5. ✅ Start backend service (port 8080)
6. ✅ Start frontend service (port 5173)

### Method 2: Docker Deployment

```bash
git clone https://github.com/coder100001/ETF-Insight.git
cd ETF-Insight
docker-compose up -d
```

Visit http://localhost:8080

### Method 3: Manual Start

```bash
# Backend
cd backend
go mod download
go build -o etf-insight .
./etf-insight

# New terminal - Frontend
cd frontend
npm install
npm run dev
```

---

## 📋 Environment Requirements

| Tool | Minimum Version | Recommended Version |
|------|-----------------|---------------------|
| Go | 1.21+ | 1.25+ |
| Node.js | 18+ | 20+ |
| npm | 9+ | 10+ |

---

## 🔧 Configuration

### Environment Variables

Copy `.env.example` and configure:

```bash
# Proxy configuration (if needed to access API)
HTTP_PROXY=http://127.0.0.1:7897
HTTPS_PROXY=http://127.0.0.1:7897

# Finage API Key (only data source) - Must configure, otherwise system won't work
FINAGE_API_KEY=your_finage_api_key_here

# Note: Finnhub API Key is deprecated, kept only for historical code
FINNHUB_API_KEY=your_finnhub_api_key_here
```

> **⚠️ Security Reminder**: API Keys must not be hard-coded in code, configure uniformly through environment variables.
> **v2.0 Architecture Update**: Fully rely on Finage real data, all hard-coded mock data has been removed.

### Backend Configuration File

Located at `backend/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  dsn: "etf_insight.db"     # SQLite / PostgreSQL DSN

etf:
  cache:
    type: "memory"
    ttl: 3600

schedule:
  update_interval: "1h"

log:
  level: "info"
```

---

## 🏛️ Cache Architecture (OOP Design)

### Design Principles

The cache system adopts object-oriented design, strictly following SOLID principles:

- **Single Responsibility Principle (SRP)** - Each component is responsible for only one function
- **Open/Closed Principle (OCP)** - Extend through interfaces without modifying existing code
- **Liskov Substitution Principle (LSP)** - All cache implementations can be substituted for each other
- **Interface Segregation Principle (ISP)** - Interfaces contain only necessary methods
- **Dependency Inversion Principle (DIP)** - Depend on abstract interfaces, not concrete implementations

### Cache Strategies

| Strategy | Description | Applicable Scenarios |
|----------|-------------|---------------------|
| **Memory** | Pure in-memory cache | Single-machine deployment, development testing |
| **Redis** | Pure Redis cache | Distributed deployment, need persistence |
| **Hybrid** | Redis + Memory hybrid | Production environment, high performance requirements |

### Core Components

```
CacheService (Business Layer)
    ↓
CacheProvider Interface (Abstract Interface)
    ↓
┌─────────────┬─────────────┬─────────────┐
│ MemoryCache │ RedisCache  │ HybridCache │
└─────────────┴─────────────┴─────────────┘
    ↓
CacheFactory (Factory Pattern)
```

### Configuration Example

```yaml
redis:
  enabled: true          # Enable Redis
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  timeout: 5s

etf:
  cache:
    realtime_ttl: 5m     # Real-time data cache time
    historical_ttl: 24h  # Historical data cache time
    metrics_ttl: 1h      # Metrics data cache time
    comparison_ttl: 30m  # Comparison data cache time
```

### Usage Example

```go
// Create cache service
cacheService := services.NewCacheService(cacheCfg, redisCfg)

// Use cache
cacheService.Set("etf:SCHD", data, 5*time.Minute)
value, found := cacheService.Get("etf:SCHD")

// Get cache statistics
stats := cacheService.GetCacheStats()
// Output: {"provider_type": "hybrid"}
```

---

## 📁 Project Structure

```
ETF-Insight/
├── start.sh                    # One-click start script (macOS/Linux)
├── start.bat                   # One-click start script (Windows)
├── .env.example                # Environment variables template
├── backend/
│   ├── main.go                 # Backend entry + route registration
│   ├── config.yaml             # Configuration file
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Configuration structure definition and loading
│   │   └── config_test.go      # Configuration tests
│   ├── models/                 # Data models
│   │   ├── models.go           # ETFConfig, ETFData, OperationLog, PortfolioConfig
│   │   ├── db.go               # Database initialization and migration
│   │   ├── exchange_rate.go    # ExchangeRate, ExchangeRateSyncLog, CurrencyPair, etc.
│   │   └── a_share_dividend_etf.go  # AShareDividendETF, AShareETFPortfolio, etc.
│   ├── handlers/               # API handlers
│   │   ├── etf_handler.go      # ETF quotes/history/metrics/forecast interfaces
│   │   ├── etf_config_handler.go    # ETF configuration CRUD interfaces
│   │   ├── portfolio_handler.go     # Portfolio analysis/configuration interfaces
│   │   ├── a_share_portfolio_handler.go  # A-share dividend ETF portfolio interfaces
│   │   ├── exchange_rate.go    # Exchange rate management interfaces
│   │   ├── health_handler.go   # Health checks (health/ready/live)
│   │   └── middleware.go       # Logging and CORS middleware
│   ├── services/               # Business logic layer
│   │   ├── datasource/         # Data source microservice layer (strategy pattern)
│   │   │   ├── provider.go     # Data source interface definition + ProviderFactory
│   │   │   ├── errors.go       # Standard error definitions
│   │   │   ├── finage_provider.go   # Finage API (aggregate API + last API)
│   │   │   ├── finnhub_provider.go  # Finnhub API implementation
│   │   │   └── fallback_provider.go # Fallback data source
│   │   ├── sync/               # Data sync service
│   │   │   ├── service.go      # Sync business logic + persistence validation + operation logs
│   │   │   └── config.go       # ETF configuration data + preset portfolios
│   │   ├── etf_analysis.go     # ETF analysis service (metrics/portfolio/forecast/comparison)
│   │   ├── yahoo_finance.go    # Yahoo Finance client
│   │   ├── cache.go            # Cache service + RealtimeData model
│   │   ├── exchange_rate.go    # Exchange rate service
│   │   └── finnhub.go          # Finnhub standalone client
│   ├── middleware/             # Middleware
│   │   ├── security.go         # Security headers + rate limiting (100/min)
│   │   └── security_test.go
│   ├── tasks/                  # Scheduled tasks
│   │   ├── scheduler.go        # Main scheduler (ETF update/exchange rate update/hourly check)
│   │   └── exchange_rate_task.go   # Exchange rate sync task (5min/10:30daily)
│   ├── utils/                  # Utilities
│   │   ├── logger.go           # Logging utility
│   │   └── logger_test.go
│   ├── cmd/                    # CLI tools
│   │   ├── syncetf/            # ETF data sync tool
│   │   ├── update_etf_prices/  # ETF price batch update (Finage aggregate API)
│   │   ├── generate_history/   # Generate simulated historical data
│   │   ├── initetf/            # ETF initial data import
│   │   ├── syncrates/          # Exchange rate data sync
│   │   ├── updateashare/       # A-share dividend ETF data update
│   │   ├── test_factory/       # Data source factory test
│   │   └── test_finage/        # Finage API test
│   └── infrastructure/         # Infrastructure (reserved directory)
├── frontend/
│   ├── src/
│   │   ├── pages/              # Page components
│   │   │   ├── Dashboard.tsx          # Dashboard
│   │   │   ├── ETFDashboard.tsx       # ETF market overview
│   │   │   ├── ETFComparison.tsx      # ETF comparison analysis
│   │   │   ├── ETFComparisonReport.tsx # ETF comparison report
│   │   │   ├── ETFDetail.tsx          # ETF detail page
│   │   │   ├── ETFConfig.tsx          # ETF configuration management
│   │   │   ├── PortfolioAnalysis.tsx   # Portfolio analysis
│   │   │   ├── PortfolioConfig.tsx     # Portfolio configuration management
│   │   │   ├── ASharePortfolio.tsx     # A-share dividend ETF portfolio
│   │   │   ├── ExchangeRate.tsx        # Exchange rate management
│   │   │   ├── InvestmentStrategy.tsx  # Investment strategy
│   │   │   └── OperationLogs.tsx       # Operation logs
│   │   ├── components/         # Common components
│   │   │   ├── Layout.tsx             # Layout
│   │   │   ├── PriceChart.tsx         # Price chart
│   │   │   ├── ComparisonRadarChart.tsx # Comparison radar chart
│   │   │   ├── ETFFilter.tsx          # ETF filter
│   │   │   ├── HoldingPieChart.tsx    # Holdings pie chart
│   │   │   ├── SectorBarChart.tsx     # Sector bar chart
│   │   │   ├── StatCard.tsx           # Statistic card
│   │   │   └── StockCard.tsx          # Stock card
│   │   ├── services/api.ts     # API service (with request merging + retry)
│   │   ├── types/index.ts      # TypeScript type definitions
│   │   └── styles/theme.ts     # Theme configuration
│   └── package.json
├── docs/
│   └── openapi.yaml            # OpenAPI 3.0 interface documentation
├── scripts/
│   ├── install-hooks.sh        # Git hooks installation
│   └── startup.sh              # Production startup script
└── docker-compose.yml
```

---

## 🌐 API Interfaces

### Health Checks
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Basic health check |
| GET | `/ready` | Readiness check (includes database/service status) |
| GET | `/live` | Liveness check (includes uptime) |

### ETF Quotes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/etf/list` | Get ETF list (includes quotes + metrics, 5min cache) |
| GET | `/api/etf/:symbol/realtime` | Get single ETF real-time data |
| GET | `/api/etf/:symbol/history` | Get historical data (supports period: 1m/3m/6m/1y/3y/5y) |
| GET | `/api/etf/:symbol/metrics` | Get risk metrics (volatility/sharpe ratio/max drawdown) |
| GET | `/api/etf/:symbol/forecast` | Get 10-year return forecast |
| GET | `/api/etf/comparison` | Get ETF comparison data |
| POST | `/api/etf/update-realtime` | Update all ETF real-time data (Yahoo Finance) |
| POST | `/api/etf/portfolio` | Analyze portfolio |

### ETF Configuration Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/etf-configs/` | Get ETF configuration list |
| POST | `/api/etf-configs/` | Create ETF configuration |
| GET | `/api/etf-configs/:id` | Get single ETF configuration |
| PUT | `/api/etf-configs/:id` | Update ETF configuration |
| DELETE | `/api/etf-configs/:id` | Delete ETF configuration |
| POST | `/api/etf-configs/:id/toggle-status` | Toggle enable/disable status |
| POST | `/api/etf-configs/:id/auto-update` | Toggle auto-update setting |

### Portfolio Configuration
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/portfolio-configs/` | Get portfolio configuration list |
| POST | `/api/portfolio-configs/` | Create portfolio configuration |
| GET | `/api/portfolio-configs/:id` | Get single portfolio configuration |
| PUT | `/api/portfolio-configs/:id` | Update portfolio configuration |
| DELETE | `/api/portfolio-configs/:id` | Delete portfolio configuration |
| POST | `/api/portfolio-configs/:id/toggle-status` | Toggle status |
| POST | `/api/portfolio-configs/:id/analyze` | Analyze portfolio returns |

### A-Share Dividend ETF Portfolio
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/a-share/etfs` | Get A-share dividend ETF list |
| GET | `/api/a-share/portfolio/default` | Get default portfolio configuration |
| POST | `/api/a-share/portfolio/analyze` | Analyze custom portfolio |
| POST | `/api/a-share/portfolio/holding/:symbol` | Update holding amount |
| GET | `/api/a-share/dividend/:frequency` | Calculate dividends by frequency (monthly/quarterly/yearly) |

### Exchange Rate Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/exchange-rates` | Get exchange rate list (supports from/to/source filtering) |
| GET | `/api/exchange-rates/:from/:to` | Get specified currency pair exchange rate |
| POST | `/api/exchange-rates/convert` | Currency conversion |
| POST | `/api/exchange-rates/sync` | Trigger exchange rate sync (async) |
| GET | `/api/exchange-rates/summary` | Get exchange rate summary |
| GET | `/api/exchange-rates/currencies` | Get supported currency list |
| GET | `/api/currency-pairs` | Get currency pair configuration |

---

## 📖 User Guide

### Start Project

```bash
# One-click start (recommended)
./start.sh

# Or manual start
cd backend && ./etf-insight &
cd frontend && npm run dev
```

### Access URLs

- **Frontend Interface**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Readiness Check**: http://localhost:8080/ready
- **Liveness Check**: http://localhost:8080/live

### ETF Data Update (v2.0 Finage-Only)

```bash
# Use Finage aggregate API to update ETF prices one by one (complete OHLCV)
cd backend && go run ./cmd/update_etf_prices/

# Trigger real-time data update via API (Finage aggregate API)
curl -X POST http://localhost:8080/api/etf/update-realtime

# Data sync CLI tool
cd backend && go run ./cmd/syncetf/

# Run with proxy
HTTP_PROXY=http://127.0.0.1:7897 HTTPS_PROXY=http://127.0.0.1:7897 go run ./cmd/update_etf_prices/
```

### Other CLI Tools

```bash
# Generate simulated historical data (90 days)
cd backend && go run ./cmd/generate_history/

# ETF initial data import
cd backend && go run ./cmd/initetf/

# Exchange rate data sync
cd backend && go run ./cmd/syncrates/

# A-share dividend ETF data update
cd backend && go run ./cmd/updateashare/

# Data source factory test
cd backend && go run ./cmd/test_factory/

# Finage API test
cd backend && go run ./cmd/test_finage/
```

### Data Acquisition Strategy (v2.0 Finage-Only)

- **Finage Aggregate API** (`/agg/stock/{symbol}/1/day`) is the only data source, providing complete OHLCV + Volume
- **No downgrade to last API** - last API data is incomplete, does not meet "all fields must be persisted" requirement
- **No Yahoo Finance dependency** - All data is obtained from Finage and fully persisted
- **Fallback Provider** only provides basic demonstration when no Finage API Key is available, does not return simulated fake data
- **All fields must be persisted** - OHLCV+Volume+DataSource are all written to etf_data table
- **Price change calculation** is based on previous day's closing price (PreviousClose), querying real data from database
- **Strict validation** - Data will be rejected if incomplete, avoiding dirty data pollution of database

### Dividend Yield Standards (v2.1)

| ETF Type | Representative ETFs | Default Dividend Yield |
|----------|---------------------|------------------------|
| High Dividend ETFs | SCHD, VYM, SPYD, HDV, DGRO | **3.5%** |
| Covered Call | JEPI, JEPQ, QYLD, XYLD | **7.0%** |
| Bond ETFs | BND, AGG, TLT | **4.0%** |
| Real Estate ETFs | VNQ | **4.0%** |
| Gold ETFs | GLD | **0.0%** |
| Broad Market Index | QQQ, VOO, VTI, SPY | **0.5%** |
| International Markets | VEA, VWO, VXUS | **3.0%** |
| Default | Others | **1.0%** |

### FAQ

**Q: What if the port is occupied?**

Modify the port configuration in `backend/config.yaml`, or stop the process occupying the port:

```bash
# macOS/Linux
lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill

# Windows
netstat -ano | findstr :8080
taskkill /PID <process ID> /F
```

**Q: Dependency installation failed?**

Users in China can set proxy:
```bash
export GOPROXY=https://goproxy.cn,direct
npm config set registry https://registry.npmmirror.com
```

---

## 📅 Scheduled Tasks

| Task | Cron Expression | Description |
|------|-----------------|-------------|
| ETF Pre-market Update | `0 30 9 * * *` | Daily 09:30 |
| ETF Close Update | `0 30 16 * * *` | Daily 16:30 |
| Exchange Rate Frequent Sync | `0 */5 * * * *` | Every 5 minutes |
| Exchange Rate Daily Sync | `0 30 10 * * *` | Daily 10:30 |
| Hourly Check | `0 0 * * * *` | Check cache status every hour |

---

## 🗺️ Development Roadmap

### Phase 1: Basic Features ✅
- [x] ETF basic information management
- [x] Real-time quote data display
- [x] Exchange rate data management
- [x] A-share dividend ETF portfolio
- [x] Investment allocation pie chart
- [x] ETF comparison analysis
- [x] ETF configuration management (CRUD)
- [x] Portfolio configuration management

### Phase 2: Deep Analysis 🚧
- [x] Risk indicator calculation (volatility, sharpe ratio, max drawdown)
- [x] ETF return forecast and portfolio analysis
- [x] Multi-data source strategy pattern (Finage/Finnhub/Fallback)
- [x] Security middleware (rate limiting/security headers)
- [ ] Holdings overlap analysis
- [ ] Industry/region distribution visualization
- [ ] Correlation matrix

### Phase 3: Portfolio Optimization 📋
- [ ] Portfolio builder
- [ ] Efficient frontier analysis
- [ ] Rebalancing strategy recommendations

### Phase 4: Advanced Features 📋
- [ ] Intelligent recommendation system
- [ ] Historical backtesting functionality
- [ ] Investment report export
- [ ] Mobile adaptation

---

## 🤝 Contributing Guide

Welcome to submit Issues and Pull Requests!

1. Fork this repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Submit Pull Request

---

## 📄 License

MIT License
