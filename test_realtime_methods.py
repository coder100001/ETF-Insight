#!/usr/bin/env python
"""
尝试使用不同方法获取实时ETF数据
"""

import yfinance as yf
import requests
from datetime import datetime
import time

print('=' * 80)
print('尝试获取实时ETF数据')
print('=' * 80)

etfs = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

# 方法1: 直接获取最新价格
print('\n【方法1: yfinance快速获取】')
for i, etf in enumerate(etfs):
    try:
        ticker = yf.Ticker(etf)
        # 使用fast_info获取快速数据
        info = ticker.fast_info
        price = info.get('last_price')
        
        if price:
            print(f'  {etf}: ${price:.2f}')
        else:
            # 尝试history
            hist = ticker.history(period='1d')
            if not hist.empty:
                price = hist['Close'].iloc[-1]
                print(f'  {etf}: ${price:.2f}')
            else:
                print(f'  {etf}: 无法获取数据')
        
        # 避免请求过快
        if i < len(etfs) - 1:
            time.sleep(0.5)
            
    except Exception as e:
        print(f'  {etf}: 错误 - {str(e)[:50]}')

print('\n' + '=' * 80)
