#!/usr/bin/env python
"""
验证最终数据
"""
import os, django
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ExchangeRate
from datetime import date, timedelta

print('=' * 80)
print('验证最终数据')
print('=' * 80)

today = date.today()

# 1. 验证ETF数据
print('\n1. ETF数据验证')
etf_today = ETFData.objects.filter(date=today)
print(f'  今日ETF: {etf_today.count()} 条')

for etf in etf_today.order_by('symbol'):
    print(f'  {etf.symbol}: ${etf.close_price} ({etf.data_source})')

# 2. 验证汇率数据
print('\n2. 汇率数据验证')
rates_today = ExchangeRate.objects.filter(rate_date=today)
print(f'  今日汇率: {rates_today.count()} 条')

seven_days_ago = today - timedelta(days=6)
rates_7days = ExchangeRate.objects.filter(
    rate_date__gte=seven_days_ago,
    rate_date__lte=today
)

print(f'  7天汇率: {rates_7days.count()} 条')

# 按日期分组
rates_by_date = {}
for rate in rates_7days:
    date_str = str(rate.rate_date)
    if date_str not in rates_by_date:
        rates_by_date[date_str] = []
    rates_by_date[date_str].append(rate)

print(f'  日期分布: {len(rates_by_date)} 天')

for date_str in sorted(rates_by_date.keys()):
    count = len(rates_by_date[date_str])
    print(f'  {date_str}: {count} 条记录')

# 3. 诊断
print('\n3. 诊断问题')
if len(rates_by_date) < 7:
    print('  ⚠️  历史走势数据不完整（需要7天）')
else:
    print('  ✅ 历史走势数据完整（7天）')

if etf_today.count() == 5:
    print('  ✅ ETF数据完整（5个）')
else:
    print(f'  ⚠️  ETF数据不完整（应有5个，实际{etf_today.count()}个）')

if rates_today.count() == 9:
    print('  ✅ 今日汇率数据完整（9条）')
else:
    print(f'  ⚠️  今日汇率数据不完整（应有9条，实际{rates_today.count()}条）')

print('\n' + '=' * 80)
print('建议操作')
print('=' * 80)
print("""
1. 刷新浏览器页面（Ctrl+F5 或 Cmd+Shift+R）
2. 清除浏览器缓存（Ctrl+Shift+Delete）
3. 访问投资组合页面: http://127.0.0.1:8000/workflow/portfolio/
4. 访问汇率页面: http://127.0.0.1:8000/workflow/exchange-rates/
5. 点击"更新实时数据"按钮获取最新ETF价格
6. 点击"立即更新"按钮更新汇率数据

如果页面仍然显示旧数据：
- 检查Django服务器是否正在运行
- 重启Django服务器
- 检查浏览器是否有缓存问题
""")

print('\n' + '=' * 80)
print('完成')
print('=' * 80)
