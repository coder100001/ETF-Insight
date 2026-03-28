#!/usr/bin/env python
"""
测试汇率更新API
"""
import os
import django
import requests
import json

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date

print("=" * 60)
print("汇率立即更新测试")
print("=" * 60)

# 获取当前汇率
today = date.today()
print(f"\n1. 更新前汇率检查 ({today}):")
before_rates = list(ExchangeRate.objects.filter(rate_date=today))
print(f"   总计: {len(before_rates)} 条记录")

for rate in before_rates[:5]:
    print(f"   1 {rate.from_currency} = {rate.rate:.6f} {rate.to_currency} ({rate.data_source})")

# 发送更新请求
print(f"\n2. 调用汇率更新API...")
print("   URL: http://127.0.0.1:8000/workflow/api/update-exchange-rates/")

try:
    # 先获取页面以获取CSRF token
    session = requests.Session()
    response = session.get('http://127.0.0.1:8000/workflow/portfolio/')

    if 'csrftoken' in session.cookies:
        csrf_token = session.cookies['csrftoken']
        print(f"   CSRF Token: {csrf_token[:20]}...")

        # 发送更新请求
        response = session.post(
            'http://127.0.0.1:8000/workflow/api/update-exchange-rates/',
            json={},
            headers={
                'X-CSRFToken': csrf_token,
                'Content-Type': 'application/json'
            }
        )

        print(f"   响应状态码: {response.status_code}")

        if response.status_code == 200:
            data = response.json()
            print(f"   更新成功: {data.get('success')}")
            print(f"   更新时间: {data.get('update_time')}")
            print(f"   更新的汇率数量: {len(data.get('rates', []))}")

            print(f"\n   更新的汇率:")
            for rate in data.get('rates', [])[:5]:
                print(f"     1 {rate['from_currency']} = {rate['rate']:.6f} {rate['to_currency']}")
        else:
            print(f"   错误: {response.text[:200]}")
    else:
        print("   错误: 无法获取CSRF token")

except Exception as e:
    print(f"   请求失败: {str(e)}")

# 检查更新后的汇率
print(f"\n3. 更新后汇率检查 ({today}):")
after_rates = list(ExchangeRate.objects.filter(rate_date=today))
print(f"   总计: {len(after_rates)} 条记录")

# 找出新增或更新的汇率
before_dict = {(r.from_currency, r.to_currency): r for r in before_rates}
after_dict = {(r.from_currency, r.to_currency): r for r in after_rates}

changed_rates = []
for key, after_rate in after_dict.items():
    before_rate = before_dict.get(key)
    if not before_rate or before_rate.rate != after_rate.rate:
        changed_rates.append(after_rate)

print(f"   新增或更新的汇率: {len(changed_rates)} 条")

for rate in after_rates[:9]:
    if rate in changed_rates:
        print(f"   [更新] 1 {rate.from_currency} = {rate.rate:.6f} {rate.to_currency} ({rate.data_source})")
    else:
        print(f"   1 {rate.from_currency} = {rate.rate:.6f} {rate.to_currency} ({rate.data_source})")

print("\n" + "=" * 60)
print("汇率更新检查完成")
print("=" * 60)
