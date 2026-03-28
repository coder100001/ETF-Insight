#!/usr/bin/env python
"""
完整ETF数据分析报告
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData
from django.db.models import Min, Max, Avg
import json

print('\n' + '='*80)
print('完整ETF数据分析报告')
print('='*80 + '\n')

symbols = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

for symbol in symbols:
    latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
    earliest = ETFData.objects.filter(symbol=symbol).order_by('date').first()
    total = ETFData.objects.filter(symbol=symbol).count()

    if latest and earliest:
        total_return = ((latest.close_price - earliest.close_price) / earliest.close_price) * 100

        print(f'{symbol}:')
        print(f'  总记录数: {total}条')
        print(f'  数据范围: {earliest.date} 至 {latest.date}')
        print(f'  最新价格: ${latest.close_price:.2f}')
        print(f'  期间涨跌: {total_return:.2f}%')
        print(f'  最高价: ${ETFData.objects.filter(symbol=symbol).aggregate(max_price=Max("close_price"))["max_price"]:.2f}')
        print(f'  最低价: ${ETFData.objects.filter(symbol=symbol).aggregate(min_price=Min("close_price"))["min_price"]:.2f}')
        print(f'  平均价: ${ETFData.objects.filter(symbol=symbol).aggregate(avg_price=Avg("close_price"))["avg_price"]:.2f}')
        print()

print('='*80)
print('分析完成')
print('='*80 + '\n')
