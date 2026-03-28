"""
更新汇率管理脚本
支持从API获取最新汇率或手动设置汇率
"""
from workflow.models import ExchangeRate
from datetime import date
import sys

def update_exchange_rate(from_currency, to_currency, rate, data_source='manual'):
    """
    更新或创建汇率记录

    参数:
        from_currency: 源货币（USD, CNY, HKD）
        to_currency: 目标货币（USD, CNY, HKD）
        rate: 汇率（1单位from_currency = ?单位to_currency）
        data_source: 数据来源（manual, api, system）
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
        print(f"✓ 更新汇率: 1 {from_currency} = {rate} {to_currency}")
    else:
        # 创建新记录
        ExchangeRate.objects.create(
            from_currency=from_currency,
            to_currency=to_currency,
            rate=rate,
            rate_date=today,
            data_source=data_source
        )
        print(f"✓ 创建汇率: 1 {from_currency} = {rate} {to_currency}")

def list_exchange_rates():
    """列出所有汇率"""
    today = date.today()
    rates = ExchangeRate.objects.filter(rate_date=today).order_by('from_currency', 'to_currency')

    print(f"\n今日汇率 ({today}):")
    print("-" * 40)
    for rate in rates:
        print(f"  1 {rate.from_currency:4s} = {rate.rate:.6f} {rate.to_currency:4s} ({rate.data_source})")
    print("-" * 40)

def set_default_rates():
    """设置默认汇率（示例）"""
    print("设置默认汇率...")
    update_exchange_rate('USD', 'USD', 1.0, 'system')
    update_exchange_rate('CNY', 'USD', 0.138889, 'system')  # 1 CNY = 0.138889 USD
    update_exchange_rate('HKD', 'USD', 0.128205, 'system')  # 1 HKD = 0.128205 USD
    update_exchange_rate('USD', 'CNY', 7.2, 'system')       # 1 USD = 7.2 CNY
    update_exchange_rate('USD', 'HKD', 7.8, 'system')       # 1 USD = 7.8 HKD
    update_exchange_rate('CNY', 'HKD', 1.083333, 'system')  # 1 CNY = 1.083333 HKD
    print("默认汇率设置完成！")

if __name__ == '__main__':
    if len(sys.argv) > 1 and sys.argv[1] == 'list':
        list_exchange_rates()
    elif len(sys.argv) > 1 and sys.argv[1] == 'default':
        set_default_rates()
        list_exchange_rates()
    elif len(sys.argv) == 4:
        # 手动设置汇率：python update_exchange_rates.py USD CNY 7.2
        from_currency = sys.argv[1].upper()
        to_currency = sys.argv[2].upper()
        rate = float(sys.argv[3])
        update_exchange_rate(from_currency, to_currency, rate, 'manual')
        list_exchange_rates()
    else:
        print("用法:")
        print("  列出今日汇率: python update_exchange_rates.py list")
        print("  设置默认汇率: python update_exchange_rates.py default")
        print("  手动设置汇率: python update_exchange_rates.py USD CNY 7.2")
        print("\n示例:")
        print("  python update_exchange_rates.py USD CNY 7.2    # 设置 1 USD = 7.2 CNY")
        print("  python update_exchange_rates.py CNY USD 0.138889  # 设置 1 CNY = 0.138889 USD")
