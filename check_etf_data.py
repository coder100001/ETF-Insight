#!/usr/bin/env python
"""
检查数据库中的ETF数据
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData
from datetime import date

print('=' * 80)
print('数据库中ETF数据检查')
print('=' * 80)

today = date.today()
print(f'当前日期: {today}\n')

etfs = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

for etf in etfs:
    print(f'\n【{etf}】')
    
    # 获取最新的记录
    latest = ETFData.objects.filter(symbol=etf).order_by('-date').first()
    
    if latest:
        print(f'  最新日期: {latest.date}')
        print(f'  收盘价: ${latest.close_price:.2f}')
        print(f'  开盘价: ${latest.open_price:.2f}')
        print(f'  最高价: ${latest.high_price:.2f}')
        print(f'  最低价: ${latest.low_price:.2f}')
        print(f'  成交量: {latest.volume:,}')
        print(f'  数据来源: {latest.data_source}')
    else:
        print(f'  无数据')

# 检查今天的数据
print(f'\n\n{"="*80}')
print(f'今日数据汇总 ({today})')
print(f'{"="*80}\n')

for etf in etfs:
    today_data = ETFData.objects.filter(symbol=etf, date=today).first()
    if today_data:
        print(f'{etf}: ${today_data.close_price:.2f}')
    else:
        print(f'{etf}: 今日无数据')

print(f'\n{"="*80}')
