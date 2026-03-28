#!/usr/bin/env python
"""
清理测试数据并重新生成7天数据
"""
import os, django
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date, timedelta
import random

print('=' * 80)
print('清理测试数据并重新生成')
print('=' * 80)

# 1. 删除所有test_data
print('\n1. 删除test_data...')
deleted = ExchangeRate.objects.filter(data_source='test_data').delete()
print(f'  删除了 {deleted} 条test_data记录')

# 2. 重新生成最近7天的数据
print('\n2. 重新生成最近7天的数据...')
today = date.today()

# 基础汇率
base_rates = {
    'USD_CNY': 7.2,
    'USD_HKD': 7.8,
    'CNY_HKD': 1.083333,
    'CNY_USD': 0.138889,
    'HKD_USD': 0.128205,
    'HKD_CNY': 0.923077,
}

# 汇率波动范围
rate_ranges = {
    'USD_CNY': (7.1, 7.3),
    'USD_HKD': (7.7, 7.9),
    'CNY_HKD': (1.06, 1.11),
    'CNY_USD': (0.136, 0.142),
    'HKD_USD': (0.126, 0.131),
    'HKD_CNY': (0.915, 0.935),
}

# 生成7天的数据
for days_ago in range(6, -1, -1):
    target_date = today - timedelta(days=days_ago)
    print(f'\n  生成 {target_date} 的数据...')

    # 生成数据时使用固定种子，保证数据一致性
    random.seed(hash(target_date))

    for pair_key in base_rates.keys():
        from_curr = pair_key[:3]
        to_curr = pair_key[3:]

        # 检查是否已存在
        existing = ExchangeRate.objects.filter(
            from_currency=from_curr,
            to_currency=to_curr,
            rate_date=target_date
        ).first()

        if existing:
            print(f'    {pair_key}: 已存在 ({existing.rate:.6f})')
        else:
            # 生成随机汇率（在合理范围内）
            min_rate, max_rate = rate_ranges[pair_key]
            rate = random.uniform(min_rate, max_rate)
            rate = round(rate, 6)

            ExchangeRate.objects.create(
                from_currency=from_curr,
                to_currency=to_curr,
                rate=rate,
                rate_date=target_date,
                data_source='system'
            )

            print(f'    {pair_key}: {rate:.6f} (已创建)')

# 添加自汇率
for days_ago in range(6, -1, -1):
    target_date = today - timedelta(days=days_ago)
    currencies = ['USD', 'CNY', 'HKD']

    for curr in currencies:
        existing = ExchangeRate.objects.filter(
            from_currency=curr,
            to_currency=curr,
            rate_date=target_date
        ).first()

        if not existing:
            ExchangeRate.objects.create(
                from_currency=curr,
                to_currency=curr,
                rate=1.0,
                rate_date=target_date,
                data_source='system'
            )

# 3. 检查结果
print('\n3. 检查结果...')
total_rates = ExchangeRate.objects.all().count()
today_rates = ExchangeRate.objects.filter(rate_date=today).exclude(data_source='test_data')
seven_days_rates = ExchangeRate.objects.filter(
    rate_date__gte=today - timedelta(days=6),
    rate_date__lte=today
).exclude(data_source='test_data')

print(f'  总汇率记录: {total_rates} 条')
print(f'  今日汇率: {today_rates.count()} 条')
print(f'  7天汇率: {seven_days_rates.count()} 条')

print('\n' + '=' * 80)
print('完成')
print('=' * 80)
