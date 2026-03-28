#!/usr/bin/env python
import os, django
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ExchangeRate
from datetime import date, timedelta, datetime

print('=' * 80)
print('数据新鲜度快速检查')
print('=' * 80)

today = date.today()
now = datetime.now()

print(f'\n当前时间: {now.strftime("%Y-%m-%d %H:%M:%S")}')

# 1. 检查ETF数据
etf_today = ETFData.objects.filter(date=today)
print(f'\n今日ETF数据 ({today}): {etf_today.count()} 条')

if etf_today.count() > 0:
    for etf in etf_today.order_by('symbol'):
        print(f'  {etf.symbol}: ${etf.close_price} (来源: {etf.data_source})')
else:
    print('  ⚠️  今日没有ETF数据')

# 2. 检查汇率数据
rates_today = ExchangeRate.objects.filter(rate_date=today)
print(f'\n今日汇率数据 ({today}): {rates_today.count()} 条')

if rates_today.count() > 0:
    for rate in rates_today.order_by('from_currency', 'to_currency'):
        print(f'  {rate.from_currency}/{rate.to_currency}: {rate.rate:.6f} (来源: {rate.data_source})')
else:
    print('  ⚠️  今日没有汇率数据')

# 3. 检查最新ETF数据
latest_etf = ETFData.objects.order_by('-created_at').first()
print(f'\n最新ETF记录: {latest_etf.symbol} ({latest_etf.date}) ${latest_etf.close_price}')
print(f'  创建时间: {latest_etf.created_at}')

# 4. 检查最新汇率数据
latest_rate = ExchangeRate.objects.order_by('-created_at').first()
print(f'\n最新汇率记录: {latest_rate.from_currency}/{latest_rate.to_currency} ({latest_rate.rate_date}) {latest_rate.rate:.6f}')
print(f'  创建时间: {latest_rate.created_at}')

# 5. 诊断
print('\n' + '=' * 80)
print('诊断结果:')
print('=' * 80)

issues = []

if etf_today.count() == 0:
    issues.append('❌ 今日没有ETF数据，需要执行实时更新')
else:
    has_realtime = any('realtime' in etf.data_source for etf in etf_today)
    if has_realtime:
        print('✅ ETF数据包含实时来源')
    else:
        print('⚠️  ETF数据可能不是最新的')

if rates_today.count() == 0:
    issues.append('❌ 今日没有汇率数据，需要执行汇率更新')
else:
    print('✅ 汇率数据存在')

if issues:
    print('\n发现问题:')
    for issue in issues:
        print(f'  {issue}')
else:
    print('\n✅ 数据看起来是新鲜的')

print('\n建议:')
print('  1. 访问投资组合页面并点击"更新实时数据"按钮')
print('  2. 访问汇率页面并点击"立即更新"按钮')
print('  3. 硬刷新浏览器页面（Ctrl+F5）')

print('\n' + '=' * 80)
