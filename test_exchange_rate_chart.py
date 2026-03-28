#!/usr/bin/env python
"""
测试汇率走势图表功能
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from django.utils import timezone
from datetime import timedelta, datetime
import json

print("=" * 80)
print("汇率走势图表数据检查")
print("=" * 80)

today = timezone.now().date()
seven_days_ago = today - timedelta(days=6)

print(f"\n📅 检查日期范围: {seven_days_ago} 至 {today}")

# 获取最近7天的汇率数据
recent_rates = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
).order_by('rate_date', 'from_currency', 'to_currency')

print(f"\n📊 汇率记录总数: {recent_rates.count()} 条")

# 按货币对分组
currency_pairs = {
    'USD_CNY': {'name': '美元/人民币', 'rates': []},
    'USD_HKD': {'name': '美元/港币', 'rates': []},
    'CNY_HKD': {'name': '人民币/港币', 'rates': []},
    'CNY_USD': {'name': '人民币/美元', 'rates': []},
    'HKD_USD': {'name': '港币/美元', 'rates': []},
}

for rate in recent_rates:
    pair_key = f"{rate.from_currency}_{rate.to_currency}"
    if pair_key in currency_pairs:
        currency_pairs[pair_key]['rates'].append({
            'date': str(rate.rate_date),
            'rate': float(rate.rate)
        })

# 显示每个货币对的数据
print("\n" + "=" * 80)
print("各货币对7天数据统计")
print("=" * 80)

for pair_key, data in currency_pairs.items():
    print(f"\n{pair_key}: {data['name']}")
    print(f"  记录数: {len(data['rates'])}")

    if data['rates']:
        rates = [r['rate'] for r in data['rates']]
        current = rates[-1]
        max_rate = max(rates)
        min_rate = min(rates)
        first = rates[0]
        change = ((current - first) / first * 100) if first > 0 else 0

        print(f"  当前值: {current:.6f}")
        print(f"  最高值: {max_rate:.6f}")
        print(f"  最低值: {min_rate:.6f}")
        print(f"  波动: {change:+.2f}%")

        print(f"  日期序列:")
        for rate in data['rates']:
            print(f"    {rate['date']}: {rate['rate']:.6f}")
    else:
        print("  ⚠️  无数据")

# 生成前端所需的数据格式
print("\n" + "=" * 80)
print("前端图表数据格式")
print("=" * 80)

history_rates = {}
for pair_key, data in currency_pairs.items():
    history_rates[pair_key] = {}
    for rate in data['rates']:
        history_rates[pair_key][rate['date']] = rate['rate']

print("\nJavaScript 变量 historyRates:")
print(json.dumps(history_rates, indent=2, ensure_ascii=False))

# 日期标签
date_labels = sorted(list(set([rate['rate_date'] for rate in recent_rates])), reverse=True)
date_labels_str = [str(d) for d in date_labels]

print(f"\nJavaScript 变量 dateLabels:")
print(json.dumps(date_labels_str, indent=2, ensure_ascii=False))

print("\n" + "=" * 80)
print("检查完成")
print("=" * 80)
