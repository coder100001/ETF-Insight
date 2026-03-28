#!/usr/bin/env python
"""
快速更新汇率脚本
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from datetime import date
from workflow.models import ExchangeRate

def update_exchange_rate(from_currency, to_currency, rate, data_source='system'):
    """
    更新或创建汇率记录
    """
    today = date.today()

    # 查找是否已存在今日的汇率记录
    rate_record = ExchangeRate.objects.filter(
        from_currency=from_currency,
        to_currency=to_currency,
        rate_date=today
    ).first()

    if rate_record:
        # 更新现有记录
        rate_record.rate = rate
        rate_record.data_source = data_source
        rate_record.save()
        print(f"✓ 更新: 1 {from_currency} = {rate:.6f} {to_currency}")
    else:
        # 创建新记录
        ExchangeRate.objects.create(
            from_currency=from_currency,
            to_currency=to_currency,
            rate=rate,
            rate_date=today,
            data_source=data_source
        )
        print(f"✓ 创建: 1 {from_currency} = {rate:.6f} {to_currency}")

def list_exchange_rates():
    """列出所有汇率"""
    today = date.today()
    rates = ExchangeRate.objects.filter(rate_date=today).order_by('from_currency', 'to_currency')

    print(f"\n今日汇率 ({today}):")
    print("-" * 50)
    for rate in rates:
        print(f"  1 {rate.from_currency:4s} = {rate.rate:10.6f} {rate.to_currency:4s}  ({rate.data_source})")
    print("-" * 50)

if __name__ == '__main__':
    print("更新今日汇率...")
    update_exchange_rate('USD', 'USD', 1.0, 'system')
    update_exchange_rate('CNY', 'CNY', 1.0, 'system')
    update_exchange_rate('HKD', 'HKD', 1.0, 'system')
    update_exchange_rate('CNY', 'USD', 0.138889, 'system')  # 1 CNY = 0.138889 USD (约7.2)
    update_exchange_rate('HKD', 'USD', 0.128205, 'system')  # 1 HKD = 0.128205 USD (约7.8)
    update_exchange_rate('USD', 'CNY', 7.2, 'system')       # 1 USD = 7.2 CNY
    update_exchange_rate('USD', 'HKD', 7.8, 'system')       # 1 USD = 7.8 HKD
    update_exchange_rate('CNY', 'HKD', 1.083333, 'system')  # 1 CNY = 1.083333 HKD
    update_exchange_rate('HKD', 'CNY', 0.923077, 'system')  # 1 HKD = 0.923077 CNY
    
    print("\n汇率更新完成！")
    list_exchange_rates()
