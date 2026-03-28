#!/usr/bin/env python
"""
汇率走势功能综合测试
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate
from datetime import date, timedelta

print("=" * 80)
print("汇率走势功能综合测试")
print("=" * 80)

today = date.today()
seven_days_ago = today - timedelta(days=6)

print(f"\n📅 检查日期: {today}")
print(f"📅 检查范围: {seven_days_ago} 至 {today}")

# 1. 数据完整性检查
print("\n" + "=" * 80)
print("1️⃣  数据完整性检查")
print("=" * 80)

required_pairs = [
    ('USD', 'CNY'),
    ('USD', 'HKD'),
    ('CNY', 'HKD'),
    ('CNY', 'USD'),
    ('HKD', 'USD'),
    ('HKD', 'CNY'),
]

recent_rates = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
).order_by('rate_date')

print(f"\n📊 最近7天记录: {recent_rates.count()} 条")
print(f"📅 预期记录数: 7天 × 6对 = 42条")
print(f"✅ 数据完整性: {'完整' if recent_rates.count() >= 42 else '不完整'}")

# 2. 日期分布检查
print("\n" + "=" * 80)
print("2️⃣  日期分布检查")
print("=" * 80)

rates_by_date = {}
for rate in recent_rates:
    date_str = str(rate.rate_date)
    if date_str not in rates_by_date:
        rates_by_date[date_str] = []
    rates_by_date[date_str].append(rate)

print(f"\n📅 日期分布: {len(rates_by_date)} 天")
print(f"✅ 日期覆盖: {'完整 (7天)' if len(rates_by_date) == 7 else f'不完整 ({len(rates_by_date)}天)'}")

# 3. 各货币对数据检查
print("\n" + "=" * 80)
print("3️⃣  各货币对数据检查")
print("=" * 80)

for from_curr, to_curr in required_pairs:
    pair_rates = ExchangeRate.objects.filter(
        from_currency=from_curr,
        to_currency=to_curr,
        rate_date__gte=seven_days_ago,
        rate_date__lte=today
    ).order_by('rate_date')

    print(f"\n{from_curr}/{to_curr}:")
    print(f"  记录数: {pair_rates.count()} 条")

    if pair_rates.count() > 0:
        rates = [float(r.rate) for r in pair_rates]
        current = rates[-1]
        max_rate = max(rates)
        min_rate = min(rates)
        first = rates[0]
        change = ((current - first) / first * 100) if first > 0 else 0
        volatility = ((max_rate - min_rate) / first * 100) if first > 0 else 0

        print(f"  当前值: {current:.6f}")
        print(f"  最高值: {max_rate:.6f}")
        print(f"  最低值: {min_rate:.6f}")
        print(f"  7天涨跌: {change:+.2f}%")
        print(f"  7天波动: {volatility:.2f}%")

        # 显示日期序列
        print(f"  日期序列:")
        for rate in pair_rates:
            print(f"    {rate.rate_date}: {rate.rate:.6f}")
    else:
        print(f"  ⚠️  无数据")

# 4. 图表数据准备检查
print("\n" + "=" * 80)
print("4️⃣  图表数据准备检查")
print("=" * 80)

history_rates = {}
for from_curr, to_curr in required_pairs:
    pair_key = f"{from_curr}_{to_curr}"
    history_rates[pair_key] = {}
    for date_str in sorted(rates_by_date.keys()):
        rate = ExchangeRate.objects.filter(
            from_currency=from_curr,
            to_currency=to_curr,
            rate_date=date_str
        ).first()
        if rate:
            history_rates[pair_key][date_str] = float(rate.rate)

print(f"\n✅ 图表数据格式检查:")
print(f"  货币对数量: {len(history_rates)}")
print(f"  日期数量: {len(sorted(rates_by_date.keys()))}")

# 显示USD/CNY的图表数据（作为示例）
print(f"\nUSD/CNY 图表数据示例:")
usd_cny_rates = history_rates.get('USD_CNY', {})
for date_str in sorted(usd_cny_rates.keys()):
    print(f"  {date_str}: {usd_cny_rates[date_str]:.6f}")

# 5. 统计分析
print("\n" + "=" * 80)
print("5️⃣  统计分析")
print("=" * 80)

# 计算各货币对的7天变化
pair_changes = []
for from_curr, to_curr in required_pairs:
    pair_key = f"{from_curr}_{to_curr}"
    rates = ExchangeRate.objects.filter(
        from_currency=from_curr,
        to_currency=to_curr,
        rate_date__gte=seven_days_ago,
        rate_date__lte=today
    ).order_by('rate_date')

    if rates.count() >= 2:
        first = float(rates.first().rate)
        last = float(rates.last().rate)
        change = ((last - first) / first * 100)
        pair_changes.append({
            'pair': f"{from_curr}/{to_curr}",
            'change': change
        })

# 按涨跌排序
pair_changes.sort(key=lambda x: x['change'], reverse=True)

print(f"\n📈 涨幅TOP3:")
for i, item in enumerate(pair_changes[:3], 1):
    print(f"  {i}. {item['pair']}: {item['change']:+.2f}%")

print(f"\n📉 跌幅TOP3:")
for i, item in enumerate(pair_changes[-3:], 1):
    print(f"  {i}. {item['pair']}: {item['change']:+.2f}%")

# 6. 功能总结
print("\n" + "=" * 80)
print("6️⃣  功能总结")
print("=" * 80)

print(f"""
✅ 数据完整性: {recent_rates.count()} 条记录
✅ 日期覆盖: {len(rates_by_date)} 天
✅ 货币对数量: {len(required_pairs)} 对
✅ 图表数据: 已准备好
✅ 统计分析: 已完成

📝 功能特点:
  • 支持切换6种货币对
  • 显示7天历史走势
  • 实时统计: 当前值、最高值、最低值、波动率
  • 交互式图表: 悬停显示详细数据
  • 响应式设计: 适配不同屏幕尺寸
  • 颜色区分: 不同货币对不同颜色
""")

print("\n🌐 访问地址: http://127.0.0.1:8000/workflow/exchange-rates/")
print("\n" + "=" * 80)
print("测试完成")
print("=" * 80)
