#!/usr/bin/env python
"""
测试实时数据更新API
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

import json
from django.test import Client
from workflow.models import ETFConfig, ExchangeRate
from datetime import date

# 创建测试客户端
client = Client()

# 获取CSRF token
response = client.get('/workflow/portfolio/')
csrf_token = response.cookies.get('csrftoken')

print("=" * 60)
print("测试实时数据更新API")
print("=" * 60)

# 测试更新实时数据
print("\n1. 测试更新实时ETF数据...")
test_data = {
    'allocation': {
        'SCHD': 0.4,
        'SPYD': 0.3,
        'JEPQ': 0.3,
        'JEPI': 0.0,
        'VYM': 0.0
    },
    'total_investment': 10000
}

response = client.post(
    '/workflow/api/update-realtime/',
    data=json.dumps(test_data),
    content_type='application/json',
    HTTP_X_CSRFTOKEN=csrf_token
)

print(f"状态码: {response.status_code}")
if response.status_code == 200:
    data = response.json()
    print(f"成功: {data.get('success')}")
    print(f"更新时间: {data.get('update_time')}")
    if 'summary' in data:
        print(f"汇总: 成功 {data['summary']['success']}/{data['summary']['total']}, 失败 {data['summary']['failed']}")
    if 'update_results' in data:
        print("\nETF更新结果:")
        for result in data['update_results']:
            if result['success']:
                print(f"  ✓ {result['symbol']}: ${result['price']:.2f}")
            else:
                print(f"  ✗ {result['symbol']}: {result['error']}")
else:
    print(f"响应: {response.content.decode('utf-8')[:200]}")

# 测试更新汇率
print("\n" + "=" * 60)
print("2. 测试更新汇率...")
print("=" * 60)

response = client.post(
    '/workflow/api/update-exchange-rates/',
    data=json.dumps({}),
    content_type='application/json',
    HTTP_X_CSRFTOKEN=csrf_token
)

print(f"状态码: {response.status_code}")
if response.status_code == 200:
    data = response.json()
    print(f"成功: {data.get('success')}")
    print(f"更新时间: {data.get('update_time')}")
    if 'rates' in data:
        print("\n汇率更新结果:")
        for rate in data['rates']:
            print(f"  1 {rate['from_currency']} = {rate['rate']:.6f} {rate['to_currency']}")
else:
    print(f"响应: {response.content.decode('utf-8')[:200]}")

# 检查数据库中的数据
print("\n" + "=" * 60)
print("3. 检查数据库中的数据")
print("=" * 60)

# 检查ETF数据
today = date.today()
enabled_etfs = ETFConfig.objects.filter(status=1)
print(f"\n启用的ETF: {list(enabled_etfs.values_list('symbol', flat=True))}")

from workflow.models import ETFData
etf_data = ETFData.objects.filter(date=today)
print(f"\n今日ETF数据 ({today}):")
for etf in etf_data:
    print(f"  {etf.symbol}: 开盘 ${etf.open_price}, 收盘 ${etf.close_price}")

# 检查汇率
rates = ExchangeRate.objects.filter(rate_date=today)
print(f"\n今日汇率 ({today}) (前5条):")
for rate in rates[:5]:
    print(f"  1 {rate.from_currency} = {rate.rate:.6f} {rate.to_currency}")

print("\n" + "=" * 60)
print("测试完成")
print("=" * 60)
