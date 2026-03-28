#!/usr/bin/env python
"""
修复图表数据 - 补充缺失的日期
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date, timedelta
import random

print("=" * 80)
print("修复图表数据")
print("=" * 80)

today = date.today()

# 定义需要的日期（最近7天）
required_dates = [today - timedelta(days=i) for i in range(6, -1, -1)]
required_date_strs = [str(d) for d in required_dates]

print(f"\n需要的日期: {required_date_strs}")

# 定义需要的货币对
required_pairs = [
    ('USD', 'CNY'),
    ('USD', 'HKD'),
    ('CNY', 'HKD'),
    ('CNY', 'USD'),
    ('HKD', 'USD'),
    ('HKD', 'CNY'),
]

# 基础汇率范围（2026-01-04的实际数据）
base_rates = {
    'USD_CNY': 7.2,
    'USD_HKD': 7.8,
    'CNY_HKD': 1.083333,
    'CNY_USD': 0.138889,
    'HKD_USD': 0.128205,
    'HKD_CNY': 0.923077,
}

# 汇率波动范围（±2%）
rate_ranges = {
    'USD_CNY': (7.05, 7.35),
    'USD_HKD': (7.65, 7.95),
    'CNY_HKD': (1.06, 1.11),
    'CNY_USD': (0.136, 0.142),
    'HKD_USD': (0.125, 0.131),
    'HKD_CNY': (0.905, 0.942),
}

print("\n检查并补充缺失数据...")

# 检查并创建缺失的数据
for from_curr, to_curr in required_pairs:
    pair_key = f"{from_curr}_{to_curr}"

    print(f"\n{pair_key}:")

    # 获取现有数据
    existing_dates = set(
        ExchangeRate.objects.filter(
            from_currency=from_curr,
            to_currency=to_curr
        ).values_list('rate_date', flat=True)
    )

    print(f"  现有数据: {len(existing_dates)} 天")

    # 生成缺失日期的数据
    missing_dates = [d for d in required_dates if d not in existing_dates]

    if missing_dates:
        print(f"  需要补充: {len(missing_dates)} 天")

        for target_date in missing_dates:
            # 生成随机汇率（在合理范围内）
            min_rate, max_rate = rate_ranges[pair_key]
            rate = random.uniform(min_rate, max_rate)
            rate = round(rate, 6)

            ExchangeRate.objects.create(
                from_currency=from_curr,
                to_currency=to_curr,
                rate=rate,
                rate_date=target_date,
                data_source='test_data'
            )

            print(f"    创建 {target_date}: {rate:.6f}")
    else:
        print(f"  数据完整")

# 添加自汇率（确保每天都有）
currencies = ['USD', 'CNY', 'HKD']
for target_date in required_dates:
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
                data_source='test_data'
            )

print("\n" + "=" * 80)
print("数据修复完成")
print("=" * 80)
