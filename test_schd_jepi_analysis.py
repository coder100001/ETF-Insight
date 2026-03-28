#!/usr/bin/env python
"""
测试SCHD和JEPI的投资组合分析
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.services import etf_service

print('=' * 80)
print('测试SCHD和JEPI投资组合分析')
print('=' * 80)

# 测试433组合：SCHD=40%、JEPI=30%、JEPQ=30%
allocation = {
    'SCHD': 0.40,
    'JEPI': 0.30,
    'JEPQ': 0.30,
    'SPYD': 0.00,
    'VYM': 0.00
}

total_investment = 10000
tax_rate = 0.10

print(f'\n投资组合配置：')
print(f'  SCHD: 40%')
print(f'  JEPI: 30%')
print(f'  JEPQ: 30%')
print(f'  总投资: ${total_investment}')
print(f'  税率: {tax_rate*100}%')

# 分析投资组合
print(f'\n开始分析...')
result = etf_service.analyze_portfolio(allocation, total_investment, tax_rate)

print(f'\n' + '=' * 80)
print('投资组合分析结果')
print('=' * 80)

print(f'\n总价值: ${result["total_value"]:.2f}')
print(f'总投资: ${result["total_investment"]:.2f}')
print(f'总收益: ${result["total_return"]:.2f} ({result["total_return_percent"]:.2f}%)')
print(f'加权股息率: {result["weighted_dividend_yield"]:.2f}%')
print(f'税前年股息: ${result["annual_dividend_before_tax"]:.2f}')
print(f'税后年股息: ${result["annual_dividend_after_tax"]:.2f}')
print(f'股息税: ${result["dividend_tax"]:.2f}')
print(f'综合总收益: ${result["total_return_with_dividend"]:.2f} ({result["total_return_with_dividend_percent"]:.2f}%)')

print(f'\n\n持仓明细：')
print(f'{"代码":<10} {"名称":<30} {"权重":<8} {"投资额":<12} {"当前价值":<12} {"股息率":<10} {"税后年股息":<15}')
print('-' * 100)

for holding in result['holdings']:
    symbol = holding['symbol']
    name = holding['name'][:28]
    weight = f"{holding['weight']:.1f}%"
    investment = f"${holding['investment']:.2f}"
    value = f"${holding['current_value']:.2f}"
    dividend_yield = f"{holding['dividend_yield']:.2f}%"
    after_tax = f"${holding['annual_dividend_after_tax']:.2f}"

    print(f'{symbol:<10} {name:<30} {weight:<8} {investment:<12} {value:<12} {dividend_yield:<10} {after_tax:<15}')

print('=' * 80)

# 单独检查SCHD和JEPI的数据
print('\n\nSCHD和JEPI的实时数据检查：')
print('=' * 80)

for symbol in ['SCHD', 'JEPI']:
    print(f'\n{symbol}:')
    data = etf_service.fetch_realtime_data(symbol)
    print(f'  价格: ${data.get("current_price", 0):.2f}')
    print(f'  股息率: {data.get("dividend_yield", 0):.2f}%')
    print(f'  开盘价: ${data.get("open_price", 0):.2f}')
    print(f'  最高价: ${data.get("day_high", 0):.2f}')
    print(f'  最低价: ${data.get("day_low", 0):.2f}')
    print(f'  成交量: {data.get("volume", 0):,}')
    print(f'  数据日期: {data.get("data_date", "N/A")}')
