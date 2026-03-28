#!/usr/bin/env python
"""
更新今日汇率为system来源
"""
import os, django
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date

print('=' * 80)
print('更新今日汇率为system来源')
print('=' * 80)

today = date.today()

# 删除今日所有旧数据
print(f'\n删除今日旧数据...')
deleted = ExchangeRate.objects.filter(rate_date=today).delete()
print(f'  删除了 {deleted} 条记录')

# 生成新的今日数据
print(f'\n生成今日数据...')

# 基础汇率
base_rates = [
    ('USD', 'CNY', 7.2),
    ('USD', 'HKD', 7.8),
    ('CNY', 'HKD', 1.083333),
    ('CNY', 'USD', 0.138889),
    ('HKD', 'USD', 0.128205),
    ('HKD', 'CNY', 0.923077),
    ('USD', 'USD', 1.0),
    ('CNY', 'CNY', 1.0),
    ('HKD', 'HKD', 1.0),
]

for from_curr, to_curr, rate in base_rates:
    ExchangeRate.objects.create(
        from_currency=from_curr,
        to_currency=to_curr,
        rate=rate,
        rate_date=today,
        data_source='system'
    )
    print(f'  创建: {from_curr}/{to_curr} = {rate:.6f}')

# 检查结果
print('\n检查结果...')
today_rates = ExchangeRate.objects.filter(rate_date=today)
print(f'  今日汇率记录: {today_rates.count()} 条')

print('\n' + '=' * 80)
print('完成')
print('=' * 80)
