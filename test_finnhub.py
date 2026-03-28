#!/usr/bin/env python
"""
测试使用Finnhub API获取ETF价格
Finnhub提供免费的股票API
"""

import requests
import json

# Finnhub API (免费注册获取API key)
# 这里使用公开的测试端点
FINNHUB_API_KEY = "demo"  # 注册后替换为真实API key

print('=' * 80)
print('测试Finnhub API获取ETF价格')
print('=' * 80)

etfs = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

print('\n【Finnhub API - Quote】')
for etf in etfs:
    try:
        url = f"https://finnhub.io/api/v1/quote?symbol={etf}&token={FINNHUB_API_KEY}"
        response = requests.get(url, timeout=5)
        data = response.json()

        if 'c' in data and data['c'] != 0:
            price = data['c']
            change = data['d']
            change_pct = data['dp']
            print(f'  {etf}: ${price:.2f} (变化: {change:+.2f}, {change_pct:+.2f}%)')
        else:
            print(f'  {etf}: {data}')
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)}')

print('\n' + '=' * 80)
