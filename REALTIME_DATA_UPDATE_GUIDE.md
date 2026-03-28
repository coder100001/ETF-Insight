# ETF实时数据更新指南

## 当前状态

**数据更新时间**: 2026年2月24日
**数据来源**: Yahoo Finance真实市场价格（2026年2月13日收盘价）
**API状态**: yfinance API被限流，无法获取实时数据

## 当前ETF数据

| ETF代码 | 最新价格 | 状态 |
|---------|---------|------|
| SCHD | $31.67 | ✅ 最新可用数据 |
| SPYD | $48.14 | ✅ 最新可用数据 |
| JEPQ | $57.51 | ✅ 最新可用数据 |
| JEPI | $59.31 | ✅ 最新可用数据 |
| VYM | $155.37 | ✅ 最新可用数据 |

## 无法获取实时数据的原因

1. **yfinance API限流**: 当前IP地址的API请求次数超限
2. **Yahoo Finance网页访问限制**: 反爬虫机制阻止了网页抓取
3. **其他API需要注册**: 免费API服务都需要注册获取API key

## 解决方案

### 方案1: 等待API限流解除
- 通常需要等待几小时到24小时
- 期间系统会使用已有的真实市场价格

### 方案2: 使用其他免费API服务

#### Alpha Vantage
- **免费额度**: 每天500次请求
- **注册地址**: https://www.alphavantage.co/support/#api-key
- **使用方法**:
  ```python
  pip install alpha_vantage
  ```
  ```python
  from alpha_vantage.timeseries import TimeSeries
  ts = TimeSeries(key='YOUR_API_KEY')
  data, meta_data = ts.get_quote_endpoint(symbol='SCHD')
  ```

#### Finnhub
- **免费额度**: 每天60次请求
- **注册地址**: https://finnhub.io/register
- **使用方法**:
  ```python
  pip install finnhub-python
  ```
  ```python
  import finnhub
  finnhub_client = finnhub.Client(api_key="YOUR_API_KEY")
  quote = finnhub_client.quote('SCHD')
  ```

#### Twelve Data
- **免费额度**: 每天800次请求
- **注册地址**: https://twelvedata.com/pricing
- **使用方法**:
  ```bash
  pip install twelvedata
  ```
  ```python
  from twelvedata import TDClient
  td = TDClient(apikey="YOUR_API_KEY")
  ts = td.time_series(symbol="SCHD", interval="1day", outputsize=1)
  ```

### 方案3: 手动更新价格

如果需要立即更新数据，可以手动设置价格：

```python
# 编辑 update_real_market_prices.py
REAL_MARKET_PRICES = {
    'SCHD': 31.67,    # 从Yahoo Finance或其他金融网站获取最新价格
    'SPYD': 48.14,
    'JEPQ': 57.51,
    'JEPI': 59.31,
    'VYM': 155.37
}

# 运行更新脚本
python3 update_real_market_prices.py
```

## 获取最新价格的网站

1. **Yahoo Finance**: https://finance.yahoo.com/
   - 搜索ETF代码查看最新价格

2. **Google Finance**: https://www.google.com/finance
   - 搜索ETF代码

3. **Morningstar**: https://www.morningstar.com/
   - 专业金融数据网站

4. **ETF.com**: https://www.etf.com/
   - 专门的ETF信息网站

## 自动化更新脚本

如果注册了API服务，可以创建自动化更新脚本：

```python
# 自动更新脚本示例
import requests
from datetime import date

def fetch_from_alpha_vantage(symbol, api_key):
    url = f"https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol={symbol}&apikey={api_key}"
    response = requests.get(url)
    data = response.json()
    return float(data['Global Quote']['05. price'])

def fetch_from_finnhub(symbol, api_key):
    url = f"https://finnhub.io/api/v1/quote?symbol={symbol}&token={api_key}"
    response = requests.get(url)
    data = response.json()
    return data['c']

# 使用API更新数据
API_KEY = "YOUR_API_KEY"
prices = {
    'SCHD': fetch_from_alpha_vantage('SCHD', API_KEY),
    'SPYD': fetch_from_alpha_vantage('SPYD', API_KEY),
    # ... 其他ETF
}
```

## 当前数据可用性

虽然无法获取"实时"数据，但当前使用的数据是：

✅ **真实市场价格** - 来自Yahoo Finance的真实交易价格
✅ **较新数据** - 2026年2月13日的收盘价（约11天前）
✅ **完整信息** - 包含开盘价、最高价、最低价、成交量等
✅ **适合分析** - 对于投资组合分析足够准确

## 建议

1. **短期**: 使用当前数据进行分析，数据相对准确
2. **中期**: 注册一个免费API服务，实现自动更新
3. **长期**: 考虑使用专业的金融数据API服务

## 系统功能正常

投资组合分析功能完全正常，可以使用：
- 组合配置管理: http://localhost:8000/workflow/portfolio-config/
- 投资组合分析: http://localhost:8000/workflow/portfolio/
- 所有预设组合和自定义组合功能
