#!/usr/bin/env python
"""
测试ETF实时数据更新API
"""
import os
import django
import requests
import json

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ETFConfig
from datetime import date

print("=" * 60)
print("ETF实时数据立即更新测试")
print("=" * 60)

# 获取当前ETF数据
today = date.today()
print(f"\n1. 更新前ETF数据检查 ({today}):")
before_data = list(ETFData.objects.filter(date=today))
print(f"   总计: {len(before_data)} 条记录")

for etf in before_data:
    print(f"   {etf.symbol}: 收盘 ${etf.close_price}")

# 发送更新请求
print(f"\n2. 调用ETF实时数据更新API...")
print("   URL: http://127.0.0.1:8000/workflow/api/update-realtime/")
print("   注意: 此操作可能需要10-20秒，因为要从yfinance获取数据")

try:
    # 先获取页面以获取CSRF token
    session = requests.Session()
    response = session.get('http://127.0.0.1:8000/workflow/portfolio/')

    if 'csrftoken' in session.cookies:
        csrf_token = session.cookies['csrftoken']
        print(f"   CSRF Token: {csrf_token[:20]}...")

        # 发送更新请求
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

        print("   正在发送请求...")
        response = session.post(
            'http://127.0.0.1:8000/workflow/api/update-realtime/',
            json=test_data,
            headers={
                'X-CSRFToken': csrf_token,
                'Content-Type': 'application/json'
            },
            timeout=60  # 设置60秒超时
        )

        print(f"   响应状态码: {response.status_code}")

        if response.status_code == 200:
            data = response.json()
            print(f"   更新成功: {data.get('success')}")
            print(f"   更新时间: {data.get('update_time')}")

            if 'summary' in data:
                summary = data['summary']
                print(f"\n   汇总:")
                print(f"     总计: {summary['total']}")
                print(f"     成功: {summary['success']}")
                print(f"     失败: {summary['failed']}")

            if 'update_results' in data:
                print(f"\n   ETF更新结果:")
                for result in data['update_results']:
                    if result['success']:
                        print(f"     ✓ {result['symbol']}: ${result['price']:.2f} (开盘${result['open']:.2f}, 最高${result['high']:.2f}, 最低${result['low']:.2f})")
                    else:
                        print(f"     ✗ {result['symbol']}: {result['error']}")
        else:
            print(f"   错误: {response.text[:200]}")
    else:
        print("   错误: 无法获取CSRF token")

except requests.exceptions.Timeout:
    print("   错误: 请求超时（超过60秒）")
except Exception as e:
    print(f"   请求失败: {str(e)}")

# 检查更新后的ETF数据
print(f"\n3. 更新后ETF数据检查 ({today}):")
after_data = list(ETFData.objects.filter(date=today))
print(f"   总计: {len(after_data)} 条记录")

for etf in after_data:
    print(f"   {etf.symbol}:")
    print(f"     开盘: ${etf.open_price}")
    print(f"     收盘: ${etf.close_price}")
    print(f"     最高: ${etf.high_price}")
    print(f"     最低: ${etf.low_price}")
    print(f"     成交量: {etf.volume}")
    print(f"     来源: {etf.data_source}")

print("\n" + "=" * 60)
print("ETF实时数据更新检查完成")
print("=" * 60)
