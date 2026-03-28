#!/usr/bin/env python
"""
ETF对比分析报告
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData, ETFConfig
from django.db.models import Min, Max, Avg

print('\n' + '='*100)
print('ETF对比分析报告'.center(100))
print('='*100 + '\n')

# 获取ETF配置信息
configs = {config.symbol: config for config in ETFConfig.objects.filter(status=1)}

symbols = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

# 打印表头
header = f"{'ETF':<8} {'名称':<40} {'最新价格':<10} {'期间涨跌':<10} {'最高价':<10} {'最低价':<10}"
print(header)
print('-' * 100)

# 打印每个ETF的数据
for symbol in symbols:
    config = configs.get(symbol)
    name = config.name if config else 'N/A'

    latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
    earliest = ETFData.objects.filter(symbol=symbol).order_by('date').first()

    if latest and earliest:
        total_return = ((latest.close_price - earliest.close_price) / earliest.close_price) * 100

        row = f"{symbol:<8} {name:<40} ${latest.close_price:<9.2f} {total_return:>9.2f}% ${ETFData.objects.filter(symbol=symbol).aggregate(max_price=Max('close_price'))['max_price']:<9.2f} ${ETFData.objects.filter(symbol=symbol).aggregate(min_price=Min('close_price'))['min_price']:<9.2f}"
        print(row)

print('\n' + '='*100)

# 性能排名
print('\n【按期间涨跌排名】')
print('-' * 100)

performance_data = []
for symbol in symbols:
    latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
    earliest = ETFData.objects.filter(symbol=symbol).order_by('date').first()

    if latest and earliest:
        total_return = ((latest.close_price - earliest.close_price) / earliest.close_price) * 100
        config = configs.get(symbol)
        name = config.name if config else 'N/A'
        performance_data.append({
            'symbol': symbol,
            'name': name,
            'return': total_return,
            'latest_price': latest.close_price
        })

# 按收益率排序
performance_data.sort(key=lambda x: x['return'], reverse=True)

for idx, data in enumerate(performance_data, 1):
    print(f"第{idx}名: {data['symbol']} - {data['name']}")
    print(f"      收益率: {data['return']:.2f}% | 最新价格: ${data['latest_price']:.2f}")

print('\n' + '='*100)
print('分析完成')
print('='*100 + '\n')
