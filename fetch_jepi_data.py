#!/usr/bin/env python
"""
获取JEPI ETF的完整历史数据
"""
import os
import sys
import django
import yfinance as yf
from datetime import datetime

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData
from workflow.operation_service import operation_service

print("=" * 60)
print("获取JEPI ETF的完整历史数据")
print("=" * 60)

try:
    # 创建工作流实例
    workflow_instance = operation_service.create_workflow_instance(
        workflow_name='JEPI数据更新',
        trigger_type=2,
        trigger_by='user'
    )

    print("\n正在下载JEPI的2年历史数据...")

    # 下载历史数据
    ticker = yf.Ticker('JEPI')
    data = ticker.history(period='2y', interval='1d')

    if data.empty:
        print("✗ 无法获取数据")
        operation_service.complete_workflow_instance(
            workflow_instance,
            success=False,
            error_message="无法获取数据"
        )
        sys.exit(1)

    print(f"✓ 成功获取 {len(data)} 条数据记录")
    print(f"  日期范围: {data.index[0].date()} 到 {data.index[-1].date()}")

    # 保存到数据库
    print("\n正在保存到数据库...")
    saved_count = 0
    for idx, row in data.iterrows():
        ETFData.objects.update_or_create(
            symbol='JEPI',
            date=idx.date(),
            defaults={
                'open_price': float(row['Open']),
                'high_price': float(row['High']),
                'low_price': float(row['Low']),
                'close_price': float(row['Close']),
                'volume': int(row['Volume']),
                'data_source': 'yfinance',
                'fetch_instance': workflow_instance,
            }
        )
        saved_count += 1
        if saved_count % 50 == 0:
            print(f"  已保存 {saved_count} 条记录...")

    print(f"\n✓ 成功保存 {saved_count} 条记录")

    # 获取实时数据
    print("\n正在获取实时数据...")
    info = ticker.info
    if info:
        current_price = info.get('currentPrice', info.get('regularMarketPrice', 'N/A'))
        print(f"  当前价格: ${current_price}")

    operation_service.complete_workflow_instance(
        workflow_instance,
        success=True,
        context_data={'saved_count': saved_count}
    )

    print("\n" + "=" * 60)
    print("✓ JEPI数据更新完成！")
    print("=" * 60)

except Exception as e:
    print(f"\n✗ 更新失败: {e}")
    import traceback
    traceback.print_exc()
    sys.exit(1)
