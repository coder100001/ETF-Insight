from workflow.models import ExchangeRate
from datetime import date

# 创建默认汇率数据（假设汇率：1 USD = 1 USD, 1 USD = 7.2 CNY, 1 USD = 7.8 HKD）
today = date.today()

rates_data = [
    ('USD', 'USD', 1.0),
    ('CNY', 'USD', 0.138889),  # 1 CNY = 0.138889 USD
    ('HKD', 'USD', 0.128205),  # 1 HKD = 0.128205 USD
    ('USD', 'CNY', 7.2),       # 1 USD = 7.2 CNY
    ('USD', 'HKD', 7.8),       # 1 USD = 7.8 HKD
    ('CNY', 'HKD', 1.083333),  # 1 CNY = 1.083333 HKD
]

for from_curr, to_curr, rate in rates_data:
    ExchangeRate.objects.get_or_create(
        from_currency=from_curr,
        to_currency=to_curr,
        rate_date=today,
        defaults={'rate': rate, 'data_source': 'system'}
    )

print('汇率数据初始化完成')
rates = ExchangeRate.objects.filter(rate_date=today)
print(f'已创建 {rates.count()} 条汇率记录')
for r in rates:
    print(f"  {r}")
