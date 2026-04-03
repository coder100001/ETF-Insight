# ETF-Insight

一个智能的 ETF 数据管理与分析平台，专注于高股息策略 ETF（SCHD、SPYD、JEPI 等）的实时数据追踪、历史数据分析和投资组合管理。

## 功能特性

- 📊 **多 ETF 数据追踪** - 支持 SCHD、SPYD、JEPQ、JEPI、VYM 等高股息 ETF
- 🔄 **自动定时更新** - 美股开盘前和收盘后自动更新数据
- 📈 **历史数据分析** - 支持多周期历史数据获取和对比分析
- 💾 **数据持久化** - MySQL 数据库 + Redis 缓存双存储
- 🌐 **Web 可视化** - Django 驱动的数据展示界面
- ⚡ **并发获取** - 多线程并发拉取，提升效率
- 🛡️ **智能重试** - 指数退避重试机制，应对 API 限制

## 技术栈

- **后端**: Python 3.7+, Django, Django REST Framework
- **数据库**: MySQL, Redis
- **任务调度**: APScheduler
- **数据获取**: yfinance
- **数据分析**: Pandas, NumPy

## 快速开始

### 1. 克隆项目

```bash
git clone git@github.com:coder100001/ETF-Insight.git
cd ETF-Insight
```

### 2. 安装依赖

```bash
pip install -r requirements.txt
```

### 3. 配置数据库

```bash
python manage.py migrate
```

### 4. 启动服务

```bash
python manage.py runserver
```

## 使用方法

### 命令行工具

```bash
# 更新所有 ETF 数据
python etf_manager.py update

# 更新指定 ETF
python etf_manager.py update SCHD SPYD

# 获取实时数据
python etf_manager.py realtime

# 启动定时更新服务
python etf_manager.py schedule start

# 查看数据状态
python etf_manager.py status
```

### Python API

```python
from core import quick_update, get_fetcher, get_scheduler

# 快速更新
result = quick_update()

# 获取数据
fetcher = get_fetcher()
data = fetcher.fetch_historical_data('SCHD', period='1y')

# 设置定时更新
scheduler = get_scheduler()
scheduler.schedule_market_close_update()
```

## 项目结构

```
ETF-Insight/
├── core/                    # 核心模块
│   ├── data_fetcher.py      # 数据获取
│   ├── data_storage.py      # 数据存储
│   ├── scheduler_service.py # 定时任务
│   └── cli.py               # 命令行接口
├── workflow/                # Django App
│   ├── models.py            # 数据模型
│   ├── views.py             # 视图
│   ├── services.py          # 业务逻辑
│   └── scheduler.py         # 定时任务
├── examples/                # 使用示例
├── etf_manager.py           # 统一入口
└── manage.py                # Django 管理
```

## 定时任务

默认定时任务配置：

| 任务 | 时间 | 说明 |
|------|------|------|
| 汇率更新 | 每天 10:30 | 更新汇率数据 |
| ETF 盘前更新 | 每天 9:30 ET | 美股开盘前 |
| ETF 收盘更新 | 每天 16:30 ET | 美股收盘后 |

## 支持的 ETF

| 代码 | 名称 | 策略 |
|------|------|------|
| SCHD | Schwab U.S. Dividend Equity ETF | 质量股息策略 |
| SPYD | SPDR Portfolio S&P 500 High Dividend ETF | 高股息收益策略 |
| JEPQ | JPMorgan Nasdaq Equity Premium Income ETF | 期权增强收益策略 |
| JEPI | JPMorgan Equity Premium Income ETF | 股息增强策略 |
| VYM | Vanguard High Dividend Yield ETF | 高股息宽基策略 |

## License

MIT License
