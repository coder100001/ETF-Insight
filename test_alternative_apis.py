#!/usr/bin/env python
"""
测试多种免费API获取ETF价格
"""

import requests
import time
from datetime import datetime
import json

print('=' * 80)
print('测试多种免费API获取ETF价格')
print(f'测试时间: {datetime.now()}')
print('=' * 80)

etfs = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

# 方法1: Twelve Data (免费API)
print('\n【方法1: Twelve Data API】')
API_KEY = "demo"  # 免费demo key
for etf in etfs:
    try:
        url = f"https://api.twelvedata.com/price?symbol={etf}&apikey={API_KEY}"
        response = requests.get(url, timeout=5)
        data = response.json()

        if 'price' in data:
            price = float(data['price'])
            print(f'  {etf}: ${price:.2f}')
        else:
            print(f'  {etf}: {data}')
        time.sleep(1)
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)}')

# 方法2: IEX Cloud (需要API key，使用sandbox测试)
print('\n【方法2: IEX Cloud (Sandbox)】')
for etf in etfs:
    try:
        url = f"https://cloud.iexapis.com/stable/stock/{etf}/quote"
        response = requests.get(url, timeout=5)
        data = response.json()

        if 'latestPrice' in data:
            price = data['latestPrice']
            print(f'  {etf}: ${price:.2f}')
        else:
            print(f'  {etf}: {data}')
        time.sleep(1)
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)}')

print('\n' + '=' * 80)
print('注意：大多数免费API需要注册获取API key')
print('建议注册以下免费API服务：')
print('  1. Finnhub - https://finnhub.io/ (每天60次免费请求)')
print('  2. Alpha Vantage - https://www.alphavantage.co/ (每天500次免费请求)')
print('  3. Twelve Data - https://twelvedata.com/ (每天800次免费请求)')
print('  4. Polygon.io - https://polygon.io/ (有限免费额度)')
print('=' * 80)
