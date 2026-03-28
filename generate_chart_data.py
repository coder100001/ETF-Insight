#!/usr/bin/env python
"""
生成最近7天的汇率测试数据
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date, timedelta
import random

print("=" * 80)
print("生成最近7天的汇率测试数据")
print("=" * 80)

today = date.today()

# 基础汇率（2026-01-04的实际数据）
base_rates = {
    'USD_CNY': 7.2,
    'USD_HKD': 7.8,
    'CNY_HKD': 1.083333,
    'CNY_USD': 0.138889,
    'HKD_USD': 0.128205,
    'HKD_CNY': 0.923077,
}

# 为每个货币对生成波动范围
rate_ranges = {
    'USD_CNY': (7.0, 7.5),
    'USD_HKD': (7.6, 8.0),
    'CNY_HKD': (1.05, 1.12),
    'CNY_USD': (0.13, 0.145),
    'HKD_USD': (0.125, 0.132),
    'HKD_CNY': (0.90, 0.95),
}

# 生成7天的数据
for days_ago in range(6, -1, -1):
    target_date = today - timedelta(days=days_ago)

    print(f"\n生成 {target_date} 的数据...")

    for pair_key in base_rates.keys():
        from_curr = pair_key[:3]
        to_curr = pair_key[3:]

        # 检查是否已存在该日期的数据
        existing = ExchangeRate.objects.filter(
            from_currency=from_curr,
            to_currency=to_curr,
            rate_date=target_date
        ).first()

        if existing:
            print(f"  {pair_key}: 已存在 ({existing.rate:.6f})")
        else:
            # 生成随机汇率（在合理范围内波动）
            min_rate, max_rate = rate_ranges[pair_key]
            rate = random.uniform(min_rate, max_rate)
            rate = round(rate, 6)

            # 创建汇率记录
            ExchangeRate.objects.create(
                from_currency=from_curr,
                to_currency=to_curr,
                rate=rate,
                rate_date=target_date,
                data_source='test_data'
            )

            print(f"  {pair_key}: {rate:.6f} (已创建)")

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
                data_source='test_data'
            )
            print(f"  {curr}/{curr}: 1.000000 (已创建)")

print("\n" + "=" * 80)
print("数据生成完成")
print("=" * 80)
