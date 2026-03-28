#!/usr/bin/env python
"""
测试修复后的代码
"""
import os, django
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ExchangeRate
from datetime import date

print('=' * 80)
print('测试修复后的代码')
print('=' * 80)

today = date.today()

# 1. 测试ETF数据查询
print('\n1. 测试ETF数据查询（今天的数据）')
etf_today = ETFData.objects.filter(date=today)
print(f'  今日ETF记录数: {etf_today.count()}')

for etf in etf_today.order_by('symbol'):
    print(f'  {etf.symbol}: ${etf.close_price} ({etf.data_source})')

# 2. 测试汇率数据查询
print('\n2. 测试汇率数据查询（今天的数据，排除test_data）')
rates_today = ExchangeRate.objects.filter(
    rate_date=today
).exclude(data_source='test_data')

print(f'  今日汇率记录数: {rates_today.count()}')

for rate in rates_today.order_by('from_currency', 'to_currency'):
    print(f'  {rate.from_currency}/{rate.to_currency}: {rate.rate:.6f} ({rate.data_source})')

# 3. 测试历史走势数据
print('\n3. 测试历史走势数据（最近7天，排除test_data）')
from datetime import timedelta
seven_days_ago = today - timedelta(days=6)

rates_7days = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
).exclude(data_source='test_data')

print(f'  7天汇率记录数: {rates_7days.count()}')

# 按日期分组
rates_by_date = {}
for rate in rates_7days:
    date_str = str(rate.rate_date)
    if date_str not in rates_by_date:
        rates_by_date[date_str] = []
    rates_by_date[date_str].append(rate)

for date_str in sorted(rates_by_date.keys()):
    print(f'  {date_str}: {len(rates_by_date[date_str])} 条记录')

# 4. 测试services.py中的fetch_realtime_data
print('\n4. 测试services.py的fetch_realtime_data方法')
from workflow.services import etf_service

for symbol in ['SCHD', 'SPYD', 'JEPQ']:
    data = etf_service.fetch_realtime_data(symbol)
    print(f'  {symbol}: ${data.get("current_price", 0)} (日期: {data.get("date", "N/A")})')

print('\n' + '=' * 80)
print('测试完成')
print('=' * 80)
