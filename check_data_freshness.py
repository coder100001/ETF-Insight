#!/usr/bin/env python
"""
检查实时数据和历史走势的时效性
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ExchangeRate
from datetime import date, timedelta, datetime

print("=" * 80)
print("检查实时数据和历史走势的时效性")
print("=" * 80)

today = date.today()
seven_days_ago = today - timedelta(days=6)

print(f"\n📅 当前日期: {today}")
print(f"📅 检查范围: {seven_days_ago} 至 {today}")
print(f"🕐 检查时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

# 1. 检查ETF数据
print("\n" + "=" * 80)
print("1️⃣  ETF实时数据检查")
print("=" * 80)

etf_today = ETFData.objects.filter(date=today)
print(f"\n📊 今日ETF数据 ({today}): {etf_today.count()} 条")

for etf in etf_today.order_by('symbol'):
    print(f"\n  {etf.symbol}:")
    print(f"    收盘价: ${etf.close_price}")
    print(f"    数据来源: {etf.data_source}")
    print(f"    创建时间: {etf.created_at}")
    print(f"    更新时间: {etf.updated_at}")

    # 检查数据是否是最新的
    time_diff = datetime.now() - etf.updated_at.replace(tzinfo=None)
    hours_ago = time_diff.total_seconds() / 3600
    if hours_ago < 1:
        print(f"    ✅ 数据新鲜度: {hours_ago:.1f} 小时前 (最新)")
    else:
        print(f"    ⚠️  数据新鲜度: {hours_ago:.1f} 小时前 (非最新)")

# 2. 检查汇率数据
print("\n" + "=" * 80)
print("2️⃣  汇率数据检查")
print("=" * 80)

rates_today = ExchangeRate.objects.filter(rate_date=today)
print(f"\n📊 今日汇率数据 ({today}): {rates_today.count()} 条")

for rate in rates_today.order_by('from_currency', 'to_currency'):
    print(f"\n  {rate.from_currency}/{rate.to_currency}:")
    print(f"    汇率: {rate.rate:.6f}")
    print(f"    数据来源: {rate.data_source}")
    print(f"    创建时间: {rate.created_at}")
    print(f"    更新时间: {rate.updated_at}")

    # 检查数据是否是最新的
    time_diff = datetime.now() - rate.updated_at.replace(tzinfo=None)
    hours_ago = time_diff.total_seconds() / 3600
    if hours_ago < 1:
        print(f"    ✅ 数据新鲜度: {hours_ago:.1f} 小时前 (最新)")
    else:
        print(f"    ⚠️  数据新鲜度: {hours_ago:.1f} 小时前 (非最新)")

# 3. 检查历史走势数据
print("\n" + "=" * 80)
print("3️⃣  历史走势数据检查")
print("=" * 80)

rates_7days = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
).order_by('rate_date')

print(f"\n📊 最近7天汇率数据: {rates_7days.count()} 条")

# 按日期分组
rates_by_date = {}
for rate in rates_7days:
    date_str = str(rate.rate_date)
    if date_str not in rates_by_date:
        rates_by_date[date_str] = []
    rates_by_date[date_str].append(rate)

print(f"\n📅 日期分布: {len(rates_by_date)} 天")

for date_str in sorted(rates_by_date.keys(), reverse=True):
    print(f"\n  {date_str}: {len(rates_by_date[date_str])} 条记录")

    # 显示该日期最新的汇率更新时间
    latest_update = max(rates_by_date[date_str], key=lambda r: r.updated_at)
    time_diff = datetime.now() - latest_update.updated_at.replace(tzinfo=None)
    hours_ago = time_diff.total_seconds() / 3600

    print(f"    最新更新: {latest_update.updated_at.strftime('%H:%M:%S')}")
    print(f"    数据新鲜度: {hours_ago:.1f} 小时前")

# 4. 检查缓存问题
print("\n" + "=" * 80)
print("4️⃣  可能的问题诊断")
print("=" * 80)

issues = []

# 检查ETF数据是否包含今日
if etf_today.count() == 0:
    issues.append("❌ 今日没有ETF数据")
else:
    # 检查数据来源
    for etf in etf_today:
        if 'realtime' not in etf.data_source:
            issues.append(f"⚠️  {etf.symbol} 的数据来源不是实时数据: {etf.data_source}")

# 检查汇率数据是否包含今日
if rates_today.count() == 0:
    issues.append("❌ 今日没有汇率数据")
else:
    # 检查数据来源
    for rate in rates_today:
        if rate.data_source == 'free_api':
            issues.append(f"⚠️  {rate.from_currency}/{rate.to_currency} 使用的是旧API数据")

# 检查历史走势数据
if len(rates_by_date) < 7:
    issues.append(f"⚠️  历史走势数据不完整: {len(rates_by_date)} 天，需要7天")

if not issues:
    print("\n✅ 未发现明显问题，数据看起来是新鲜的")
    print("\n💡 可能的原因:")
    print("  1. 页面缓存问题 - 尝试硬刷新（Ctrl+F5）")
    print("  2. 浏览器缓存 - 清除浏览器缓存")
    print("  3. Django模板缓存 - 重启服务器")
    print("  4. 数据库连接问题 - 检查数据库连接")
else:
    print("\n🔍 发现的问题:")
    for issue in issues:
        print(f"  {issue}")

# 5. 建议
print("\n" + "=" * 80)
print("5️⃣  建议解决方案")
print("=" * 80)

print("""
📝 检查步骤:
  1. 检查服务器是否正在运行
  2. 检查数据库连接是否正常
  3. 清除浏览器缓存（Ctrl+Shift+Delete）
  4. 硬刷新页面（Ctrl+F5 或 Cmd+Shift+R）
  5. 重启Django服务器

🔧 如果问题持续:
  1. 检查Django DEBUG设置（应=True）
  2. 检查模板缓存是否启用
  3. 检查数据库查询是否使用了正确的日期
  4. 检查前端JavaScript是否有缓存问题
""")

print("\n" + "=" * 80)
print("检查完成")
print("=" * 80)
