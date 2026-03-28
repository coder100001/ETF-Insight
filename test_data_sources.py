#!/usr/bin/env python
"""
测试多种数据源获取ETF实时价格
"""

import yfinance as yf
import requests
import time
from datetime import datetime
import json

print('=' * 80)
print('测试多种数据源获取ETF实时价格')
print(f'测试时间: {datetime.now()}')
print('=' * 80)

etfs = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

# 方法1: 使用yfinance快速获取
print('\n【方法1: yfinance快速获取（fast_info）】')
for etf in etfs:
    try:
        ticker = yf.Ticker(etf)
        info = ticker.fast_info
        price = info.get('last_price')
        if price:
            print(f'  {etf}: ${price:.2f}')
        else:
            print(f'  {etf}: 无法获取价格')
        time.sleep(0.3)
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)[:40]}')

# 方法2: 使用yfinance history
print('\n【方法2: yfinance history(1d)】')
for etf in etfs:
    try:
        ticker = yf.Ticker(etf)
        hist = ticker.history(period='1d', interval='1m')
        if not hist.empty:
            price = hist['Close'].iloc[-1]
            print(f'  {etf}: ${price:.2f}')
        else:
            print(f'  {etf}: 无数据')
        time.sleep(0.3)
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)[:40]}')

# 方法3: 使用yfinance info
print('\n【方法3: yfinance info (regularMarketPrice)】')
for etf in etfs:
    try:
        ticker = yf.Ticker(etf)
        info = ticker.info
        price = info.get('regularMarketPrice')
        if price:
            print(f'  {etf}: ${price:.2f}')
        else:
            print(f'  {etf}: 无法获取价格')
        time.sleep(0.5)
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)[:40]}')

print('\n' + '=' * 80)
