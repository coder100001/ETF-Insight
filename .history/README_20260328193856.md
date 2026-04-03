# ETF-Insight

一个专业的 ETF 数据管理与分析平台，提供完整的 ETF 基础信息、持仓数据、行情数据、技术指标等一站式解决方案。

## 🌟 核心特性

### 基础数据层（Data Layer）
- **ETF基础信息** - 发行方、费率、AUM、成立时间、跟踪指数
- **实时 & 历史行情** - K线数据（1分钟至月线）、成交量、换手率、买卖盘
- **持仓数据（核心）** - 前10大持仓、行业分布、地区分布、季度调仓
- **净值数据** - NAV、市价、溢价率
- **分红数据** - 分红金额、除息日、股息率
- **技术指标** - MA、RSI、MACD、夏普比率、最大回撤

### 智能数据管理
- 🔄 **自动定时更新** - 美股开盘前和收盘后自动更新数据
- ⚡ **并发获取** - 多线程并发拉取，提升效率
- 🛡️ **智能重试** - 指数退避重试机制，应对 API 限制
- 💾 **双存储架构** - MySQL 持久化 + Redis 缓存

### 动态配置管理
- 📋 **ETF动态配置** - 支持美股/A股/港股，灵活增删改查
- 📊 **策略管理** - 支持质量股息、高股息收益、期权增强等多种策略
- 🌐 **汇率管理** - 自动更新人民币、港币、美元汇率
- 📝 **操作日志** - 完整记录所有系统操作

### Web 可视化
- 📈 **仪表盘** - 实时数据概览、关键指标展示
- 📊 **K线图表** - 交互式价格图表，支持多周期切换
- 🔄 **持仓分析** - 持仓明细、行业分布、地区分布可视化
- 📉 **技术指标** - MA、RSI、MACD 等技术指标展示
- 📊 **对比分析** - 多 ETF 对比分析

## 📊 支持的 ETF 策略

| 策略类型 | 示例 ETF | 描述 |
|---------|---------|------|
| **质量股息** | SCHD | Schwab U.S. Dividend Equity ETF |
| **高股息收益** | SPYD | SPDR Portfolio S&P 500 High Dividend ETF |
| **期权增强收益** | JEPQ | JPMorgan Nasdaq Equity Premium Income ETF |
| **股息增强** | JEPI | JPMorgan Equity Premium Income ETF |
| **高股息宽基** | VYM | Vanguard High Dividend Yield ETF |
| **科技成长** | QQQ | Invesco QQQ Trust |
| **标普500** | SPY | SPDR S&P 500 ETF Trust |

## 🛠️ 技术栈

### 后端
- **Python 3.7+** - 核心语言
- **Django** - Web 框架
- **Django REST Framework** - API 框架
- **APScheduler** - 定时任务调度
- **yfinance** - 财经数据获取
- **Pandas / NumPy** - 数据分析

### 数据库
- **MySQL** - 主数据库，存储所有历史数据
- **Redis** - 缓存层，提升查询性能

### 前端
- **React 18** - 前端框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Ant Design** - UI 组件库
- **ECharts / Chart.js** - 数据可视化

## 🚀 快速开始

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

### 4. 初始化 ETF 配置

```bash
python init_etf_config.py
```

### 5. 启动服务

```bash
python manage.py runserver
```

访问 http://localhost:8000 查看 Web 界面

## 📖 使用指南

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

# 完整同步 ETF 数据（包含持仓、行业分布等）
python etf_manager.py sync SCHD
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

# 同步 ETF 基础数据
from core.etf_data_service import get_etf_data_service
service = get_etf_data_service()
service.sync_etf_base_info('SCHD')
service.sync_etf_holdings('SCHD')
service.sync_etf_sectors('SCHD')
service.sync_etf_regions('SCHD')
service.sync_etf_dividends('SCHD')
service.full_sync_etf('SCHD')
```

### Web 界面

| 页面 | URL | 功能 |
|------|-----|------|
| 仪表盘 | http://localhost:8000/dashboard/ | 实时数据概览 |
| ETF列表 | http://localhost:8000/etf/ | ETF配置管理 |
| ETF详情 | http://localhost:8000/etf/{symbol}/ | ETF详细信息 |
| 持仓分析 | http://localhost:8000/etf/{symbol}/holdings/ | 持仓明细 |
| 行业分布 | http://localhost:8000/etf/{symbol}/sectors/ | 行业分布 |
| 地区分布 | http://localhost:8000/etf/{symbol}/regions/ | 地区分布 |
| 对比分析 | http://localhost:8000/etf/comparison/ | 多 ETF 对比 |
| 操作日志 | http://localhost:8000/logs/ | 系统操作日志 |

## 📁 项目结构

```
ETF-Insight/
├── core/                          # 核心模块
│   ├── config.py                  # 配置管理
│   ├── data_fetcher.py            # 数据获取器
│   ├── data_storage.py            # 数据存储器
│   ├── etf_data_service.py        # ETF 数据服务
│   ├── scheduler_service.py       # 定时任务服务
│   └── cli.py                     # 命令行接口
├── workflow/                      # Django App
│   ├── models.py                  # 数据模型
│   ├── models_etf_data_layer.py   # ETF 基础数据层模型
│   ├── views.py                   # 视图
│   ├── services.py                # 业务逻辑
│   ├── scheduler.py               # 定时任务
│   └── urls.py                    # 路由配置
├── frontend/                      # 前端项目
│   ├── src/
│   │   ├── components/            # UI组件
│   │   ├── pages/                 # 页面
│   │   ├── styles/                # 样式
│   │   └── utils/                 # 工具函数
│   └── package.json
├── etf_manager.py                 # 统一入口
├── init_etf_config.py             # ETF配置初始化
├── manage.py                      # Django 管理
└── requirements.txt               # 依赖列表
```

## 🗄️ 数据模型

### ETF 基础信息 (ETFBaseInfo)
- ETF代码、名称、英文名称
- 市场、资产类别、分类
- 发行方、官网
- 跟踪指数、跟踪方式
- 成立日期、AUM、流通股数
- 管理费率、其他费用
- 上市交易所、交易货币
- 杠杆信息、反向信息
- 投资策略、投资目标、业绩基准

### 价格数据 (ETFPrice)
- 开高低收、昨收价
- 成交量、成交额
- 涨跌额、涨跌幅、换手率
- 买卖盘、买卖量
- 时间周期（1分钟至月线）

### 净值数据 (ETFNav)
- NAV、净值涨跌、净值涨跌幅
- 市价
- 溢价率 = (市价 - NAV) / NAV * 100

### 持仓数据 (ETFHolding)
- 持仓代码、名称、资产类型
- 持仓数量、市值、权重
- 权重变化、当前价格
- 报告日期、是否估算

### 行业分布 (ETFHoldingSector)
- 行业名称、行业代码
- 权重、权重变化
- 市值、股票数量

### 地区分布 (ETFHoldingRegion)
- 地区名称、地区代码、国家
- 权重、权重变化
- 市值、股票数量

### 调仓记录 (ETFRebalance)
- 持仓代码、名称
- 变化类型（新增/移除/增持/减持/不变）
- 原权重、新权重、权重变化
- 报告期、上期报告期

### 分红数据 (ETFDividend)
- 分红类型、每股分红金额
- 除息日、股权登记日、派息日
- 派息频率
- 股息率、年化股息率

### 技术指标 (ETFIndicator)
- 移动平均线（MA5/10/20/60/120）
- 波动率（20日、60日）
- RSI（6/12/24）
- MACD（DIF、DEA、柱状图）
- 布林带（上轨、中轨、下轨）
- 夏普比率、最大回撤、Beta、Alpha

## ⏰ 定时任务

| 任务 | 时间 | 说明 |
|------|------|------|
| 汇率更新 | 每天 10:30 | 更新人民币、港币、美元汇率 |
| ETF 盘前更新 | 每天 9:30 ET | 美股开盘前更新数据 |
| ETF 收盘更新 | 每天 16:30 ET | 美股收盘后更新数据 |
| 持仓同步 | 每周日 20:00 | 同步最新持仓数据 |
| 技术指标计算 | 每日 22:00 | 计算技术指标 |

## 🔧 配置管理

### ETF 动态配置

访问 http://localhost:8000/workflow/etf-config/ 管理 ETF

功能：
- ✅ 美股/A股/港股分 Tab 展示
- ✅ 添加/编辑/删除 ETF
- ✅ 启用/禁用 ETF
- ✅ 统计信息（总数、美股数、A股数、启用数）
- ✅ 排序显示

### 默认 ETF 配置

系统初始化时自动添加以下 ETF：

| 代码 | 名称 | 策略 | 市场 |
|------|------|------|------|
| SCHD | Schwab U.S. Dividend Equity ETF | 质量股息策略 | US |
| SPYD | SPDR Portfolio S&P 500 High Dividend ETF | 高股息收益策略 | US |
| JEPQ | JPMorgan Nasdaq Equity Premium Income ETF | 期权增强收益策略 | US |
| JEPI | JPMorgan Equity Premium Income ETF | 股息增强策略 | US |
| VYM | Vanguard High Dividend Yield ETF | 高股息宽基策略 | US |

## 📊 API 接口

### ETF 配置 API

```http
GET    /workflow/etf-config/          # 获取所有 ETF 配置
POST   /workflow/etf-config/          # 添加 ETF 配置
GET    /workflow/etf-config/{id}/     # 获取 ETF 详情
PUT    /workflow/etf-config/{id}/     # 更新 ETF
DELETE /workflow/etf-config/{id}/     # 删除 ETF
PATCH  /workflow/etf-config/{id}/     # 切换启用/禁用状态
```

### 数据获取 API

```http
GET    /workflow/etf-data/{symbol}/          # 获取 ETF 价格数据
GET    /workflow/etf-nav/{symbol}/           # 获取 ETF 净值数据
GET    /workflow/etf-holdings/{symbol}/      # 获取 ETF 持仓数据
GET    /workflow/etf-sectors/{symbol}/       # 获取 ETF 行业分布
GET    /workflow/etf-regions/{symbol}/       # 获取 ETF 地区分布
GET    /workflow/etf-rebalances/{symbol}/    # 获取 ETF 调仓记录
GET    /workflow/etf-dividends/{symbol}/     # 获取 ETF 分红数据
GET    /workflow/etf-indicators/{symbol}/    # 获取 ETF 技术指标
```

## 🧪 测试

```bash
# 运行测试
python test_etf_config.py

# 运行所有测试
python manage.py test
```

## 📝 开发计划

### 已完成 ✅
- [x] ETF 基础数据层模型设计
- [x] ETF 持仓数据结构
- [x] 实时 & 历史行情数据
- [x] ETF 基础信息获取逻辑
- [x] ETF 动态配置管理
- [x] ETF 配置初始化脚本
- [x] ETF 配置 Web 管理界面
- [x] ETF 配置服务层动态读取
- [x] 操作日志分页功能

### 进行中 🚧
- [ ] ETF 基础数据 Web 界面
- [ ] ETF 持仓数据可视化
- [ ] ETF 技术指标展示
- [ ] ETF 对比分析功能

### 待开发 📋
- [ ] 批量操作（批量启用/禁用、批量删除）
- [ ] ETF 配置导入导出（CSV/Excel）
- [ ] ETF 配置修改历史记录
- [ ] 用户权限控制
- [ ] 数据验证规则（费率范围检查等）
- [ ] ETF 搜索功能
- [ ] ETF 分析报告生成
- [ ] 邮件通知功能

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License

## 📧 联系

如有问题或建议，请通过以下方式联系：
- 提交 Issue
- 发送邮件至：coder100001@gmail.com
