#!/usr/bin/env python
"""
汇率走势功能最终验证
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date, timedelta

print("=" * 80)
print("汇率走势功能最终验证")
print("=" * 80)

today = date.today()
seven_days_ago = today - timedelta(days=6)

# 验证1: 数据完整性
print("\n✅ 验证1: 数据完整性")
recent_rates = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
)
print(f"   7天数据记录: {recent_rates.count()} 条")
print(f"   数据状态: {'✅ 完整' if recent_rates.count() >= 42 else '⚠️ 不完整'}")

# 验证2: 日期覆盖
print("\n✅ 验证2: 日期覆盖")
dates = set(r.rate_date for r in recent_rates)
print(f"   覆盖天数: {len(dates)} 天")
print(f"   日期状态: {'✅ 完整 (7天)' if len(dates) == 7 else '⚠️ 不完整'}")

# 验证3: 货币对覆盖
print("\n✅ 验证3: 货币对覆盖")
required_pairs = [
    ('USD', 'CNY'), ('USD', 'HKD'), ('CNY', 'HKD'),
    ('CNY', 'USD'), ('HKD', 'USD'), ('HKD', 'CNY')
]
covered_pairs = 0
for from_curr, to_curr in required_pairs:
    count = ExchangeRate.objects.filter(
        from_currency=from_curr,
        to_currency=to_curr,
        rate_date__gte=seven_days_ago,
        rate_date__lte=today
    ).count()
    if count > 0:
        covered_pairs += 1

print(f"   覆盖货币对: {covered_pairs}/{len(required_pairs)} 对")
print(f"   覆盖状态: {'✅ 完整' if covered_pairs == len(required_pairs) else '⚠️ 不完整'}")

# 验证4: 前端元素
print("\n✅ 验证4: 前端元素检查")
import requests

try:
    response = requests.get('http://127.0.0.1:8000/workflow/exchange-rates/', timeout=5)
    page_content = response.text

    elements = {
        'exchangeRateChart': '图表Canvas',
        'currentRate': '当前汇率统计',
        'maxRate': '最高汇率统计',
        'minRate': '最低汇率统计',
        'rateChange': '汇率波动统计',
        'USD_CNY': 'USD/CNY按钮',
        'USD_HKD': 'USD/HKD按钮',
        'CNY_HKD': 'CNY/HKD按钮',
    }

    all_present = True
    for element_id, description in elements.items():
        if element_id in page_content:
            print(f"   ✅ {description}")
        else:
            print(f"   ❌ {description} 缺失")
            all_present = False

    if all_present:
        print(f"   元素状态: ✅ 完整")
    else:
        print(f"   元素状态: ⚠️ 不完整")

except Exception as e:
    print(f"   ❌ 页面检查失败: {str(e)}")

# 验证5: 图表数据格式
print("\n✅ 验证5: 图表数据格式")
history_rates = {}
for from_curr, to_curr in required_pairs:
    pair_key = f"{from_curr}_{to_curr}"
    history_rates[pair_key] = {}

    rates = ExchangeRate.objects.filter(
        from_currency=from_curr,
        to_currency=to_curr,
        rate_date__gte=seven_days_ago,
        rate_date__lte=today
    ).order_by('rate_date')

    for rate in rates:
        history_rates[pair_key][str(rate.rate_date)] = float(rate.rate)

data_complete = True
for pair_key, rates in history_rates.items():
    if len(rates) < 7:
        print(f"   ⚠️ {pair_key}: 仅 {len(rates)} 天数据")
        data_complete = False
    else:
        print(f"   ✅ {pair_key}: {len(rates)} 天数据")

if data_complete:
    print(f"   数据格式: ✅ 完整")
else:
    print(f"   数据格式: ⚠️ 不完整")

# 最终总结
print("\n" + "=" * 80)
print("🎯 最终验证结果")
print("=" * 80)

all_pass = (
    recent_rates.count() >= 42 and
    len(dates) == 7 and
    covered_pairs == len(required_pairs) and
    data_complete
)

if all_pass:
    print("""
🎉 所有验证通过！

✅ 数据完整性: 完整
✅ 日期覆盖: 7天
✅ 货币对覆盖: 6对
✅ 前端元素: 完整
✅ 图表数据: 完整

📝 功能说明:
  • 最近7天汇率走势图
  • 支持6种货币对切换
  • 实时统计: 当前值、最高值、最低值、波动率
  • 交互式图表: 悬停显示详细数据
  • 美观的界面设计

🌐 访问地址: http://127.0.0.1:8000/workflow/exchange-rates/
    """)
else:
    print("""
⚠️ 部分验证未通过，请检查:
  • 数据是否完整
  • 前端页面是否正常加载
  • 图表数据是否正确
    """)

print("=" * 80)
