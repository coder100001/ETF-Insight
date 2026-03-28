#!/usr/bin/env python
"""
综合测试脚本 - 汇率更新检查
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ExchangeRate, ETFData, ETFConfig
from datetime import date, datetime

print("=" * 80)
print(" " * 20 + "汇率立即更新检查报告")
print("=" * 80)

today = datetime.now().date()
print(f"\n📅 检查日期: {today}")
print(f"🕐 检查时间: {datetime.now().strftime('%H:%M:%S')}")

# 1. 汇率数据检查
print("\n" + "=" * 80)
print("1️⃣  汇率数据检查")
print("=" * 80)

rates = ExchangeRate.objects.filter(rate_date=today)
print(f"\n📊 汇率记录总数: {rates.count()} 条")

if rates.count() == 0:
    print("⚠️  警告: 今日没有汇率记录！")
else:
    print("\n✅ 完整汇率列表:")
    print("-" * 80)
    print(f"{'源货币':<8} {'目标货币':<8} {'汇率':<15} {'数据来源'}")
    print("-" * 80)

    for rate in rates.order_by('from_currency', 'to_currency'):
        print(f"{rate.from_currency:<8} {rate.to_currency:<8} {rate.rate:<15.6f} {rate.data_source}")

    print("-" * 80)

    # 检查关键汇率
    print("\n🔍 关键汇率检查:")
    key_rates = {
        ('USD', 'CNY'): '美元兑人民币',
        ('USD', 'HKD'): '美元兑港币',
        ('CNY', 'USD'): '人民币兑美元',
        ('HKD', 'USD'): '港币兑美元',
    }

    rate_dict = {(r.from_currency, r.to_currency): r for r in rates}

    for (from_curr, to_curr), desc in key_rates.items():
        if (from_curr, to_curr) in rate_dict:
            rate = rate_dict[(from_curr, to_curr)]
            print(f"  ✅ {desc}: 1 {from_curr} = {rate.rate:.6f} {to_curr}")
        else:
            print(f"  ❌ {desc}: 未找到数据")

# 2. ETF数据检查
print("\n" + "=" * 80)
print("2️⃣  ETF实时数据检查")
print("=" * 80)

etf_config = ETFConfig.objects.filter(status=1)
etf_data = ETFData.objects.filter(date=today)

print(f"\n📊 启用的ETF数量: {etf_config.count()} 个")
print(f"📊 今日数据记录: {etf_data.count()} 条")

if etf_config.count() != etf_data.count():
    print("⚠️  警告: ETF配置数量与数据记录数量不匹配！")

print("\n✅ ETF实时数据:")
print("-" * 80)
print(f"{'ETF代码':<8} {'开盘价':<12} {'收盘价':<12} {'最高价':<12} {'最低价':<12} {'数据源'}")
print("-" * 80)

etf_dict = {d.symbol: d for d in etf_data}

for etf in etf_config.order_by('symbol'):
    symbol = etf.symbol
    name = etf.name
    if symbol in etf_dict:
        data = etf_dict[symbol]
        print(f"{symbol:<8} ${data.open_price:<11.4f} ${data.close_price:<11.4f} ${data.high_price:<11.4f} ${data.low_price:<11.4f} {data.data_source}")
    else:
        print(f"{symbol:<8} {'无数据':<12} {'无数据':<12} {'无数据':<12} {'无数据':<12} {'无数据'}")

print("-" * 80)

# 3. 数据质量检查
print("\n" + "=" * 80)
print("3️⃣  数据质量检查")
print("=" * 80)

issues = []

# 检查汇率
if rates.count() == 0:
    issues.append("❌ 今日没有汇率记录")
else:
    for rate in rates:
        if rate.rate <= 0:
            issues.append(f"❌ 汇率异常: {rate.from_currency}->{rate.to_currency} = {rate.rate}")

# 检查ETF数据
for etf in etf_config:
    if etf.symbol not in etf_dict:
        issues.append(f"⚠️  ETF {etf.symbol} 没有今日数据")
    else:
        data = etf_dict[etf.symbol]
        if data.close_price <= 0:
            issues.append(f"❌ ETF {etf.symbol} 价格异常: {data.close_price}")

if not issues:
    print("\n✅ 所有数据质量检查通过！")
else:
    print("\n⚠️  发现问题:")
    for issue in issues:
        print(f"  {issue}")

# 4. 更新API状态
print("\n" + "=" * 80)
print("4️⃣  更新API状态")
print("=" * 80)

print("\n✅ 配置的API端点:")
print("  • POST /workflow/api/update-realtime/      - 更新实时ETF数据")
print("  • POST /workflow/api/update-exchange-rates/ - 更新汇率数据")

print("\n📝 使用说明:")
print("  1. 在浏览器访问: http://127.0.0.1:8000/workflow/portfolio/")
print("  2. 点击 '更新实时数据' 按钮获取最新ETF价格")
print("  3. 点击 '更新汇率' 按钮更新汇率数据")
print("  4. 系统会自动显示更新结果和统计数据")

# 5. 总结
print("\n" + "=" * 80)
print("📋 检查总结")
print("=" * 80)

print(f"\n✅ 汇率数据: {rates.count()} 条记录")
print(f"✅ ETF数据: {etf_data.count()} 条记录")
print(f"✅ 启用ETF: {etf_config.count()} 个")
print(f"✅ 数据问题: {len(issues)} 个")

if not issues:
    print("\n🎉 所有检查通过！系统数据完整，可以正常使用。")
else:
    print(f"\n⚠️  发现 {len(issues)} 个问题，请检查数据源或手动更新。")

print("\n" + "=" * 80)
print("检查完成 - " + datetime.now().strftime('%Y-%m-%d %H:%M:%S'))
print("=" * 80)
