#!/usr/bin/env python
"""
检查实时数据完整性
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ExchangeRate, ETFConfig
from datetime import date

print('=' * 60)
print('检查实时数据')
print('=' * 60)

today = date.today()
print(f'\n今日日期: {today}\n')

# 检查ETF数据
print('1. 启用的ETF配置:')
enabled_etfs = ETFConfig.objects.filter(status=1)
for etf in enabled_etfs:
    print(f'   {etf.symbol} - {etf.name}')

print('\n2. 今日ETF实时数据:')
etf_data = ETFData.objects.filter(date=today)
for data in etf_data:
    print(f'   {data.symbol}:')
    print(f'     开盘: ${data.open_price}')
    print(f'     收盘: ${data.close_price}')
    print(f'     最高: ${data.high_price}')
    print(f'     最低: ${data.low_price}')
    print(f'     成交量: {data.volume}')

# 检查汇率
print('\n3. 今日汇率:')
rates = ExchangeRate.objects.filter(rate_date=today)
print(f'   总计: {rates.count()} 条汇率记录')
for rate in rates.order_by('from_currency', 'to_currency'):
    print(f'   1 {rate.from_currency} = {rate.rate:.6f} {rate.to_currency} ({rate.data_source})')

print('\n' + '=' * 60)
print('检查完成')
print('=' * 60)
