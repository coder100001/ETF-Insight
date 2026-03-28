#!/usr/bin/env python
"""
检查汇率图表数据
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date, timedelta

print("=" * 80)
print("汇率图表数据检查")
print("=" * 80)

today = date.today()
seven_days_ago = today - timedelta(days=6)

print(f"\n📅 检查日期范围: {seven_days_ago} 至 {today}")

# 获取所有汇率数据
all_rates = ExchangeRate.objects.all()
print(f"\n📊 汇率记录总数: {all_rates.count()} 条")

# 获取最近7天的数据
recent_rates = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
).order_by('rate_date')

print(f"📊 最近7天记录: {recent_rates.count()} 条")

# 按日期分组
rates_by_date = {}
for rate in recent_rates:
    date_str = str(rate.rate_date)
    if date_str not in rates_by_date:
        rates_by_date[date_str] = []
    rates_by_date[date_str].append({
        'from': rate.from_currency,
        'to': rate.to_currency,
        'rate': float(rate.rate)
    })

print(f"\n📅 日期分布: {len(rates_by_date)} 天")

for date_str in sorted(rates_by_date.keys()):
    print(f"\n  {date_str}:")
    for rate in rates_by_date[date_str]:
        print(f"    {rate['from']}/{rate['to']}: {rate['rate']:.6f}")

print("\n" + "=" * 80)
print("检查完成")
print("=" * 80)
