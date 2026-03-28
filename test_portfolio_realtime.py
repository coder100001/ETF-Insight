#!/usr/bin/env python
"""
测试投资组合分析功能使用最新ETF数据
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.services import etf_service
from workflow.models import PortfolioConfig, ETFData
from datetime import datetime

print('\n' + '='*80)
print('投资组合分析 - 最新ETF数据测试')
print('='*80 + '\n')

# 1. 检查最新ETF数据
print('1. 检查最新ETF数据:')
print('-' * 80)
symbols = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']
for symbol in symbols:
    latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
    if latest:
        print(f'  {symbol}: ${latest.close_price:.2f} ({latest.date})')
    else:
        print(f'  {symbol}: 无数据')

# 2. 测试433组合分析
print('\n2. 测试433组合分析 (SCHD 40%, SPYD 30%, JEPQ 30%):')
print('-' * 80)

try:
    result = etf_service.analyze_portfolio(
        allocation={
            'SCHD': 0.40,
            'SPYD': 0.30,
            'JEPQ': 0.30
        },
        total_investment=10000
    )

    print(f'  投资金额: ${result.get("total_investment", 0):,.2f}')
    print(f'  当前价值: ${result.get("total_value", 0):,.2f}')
    print(f'  总收益: ${result.get("total_return", 0):,.2f}')
    print(f'  收益率: {result.get("total_return_percent", 0):+.2f}%')
    print(f'  年股息: ${result.get("total_dividend", 0):,.2f}')
    print(f'  加权股息率: {result.get("weighted_dividend_yield", 0):.2f}%')

    print('\n  持仓明细:')
    for holding in result.get('holdings', []):
        symbol = holding['symbol']
        weight = holding['weight'] * 100
        value = holding.get('current_value', 0)
        shares = holding.get('shares', 0)
        price = holding.get('current_price', 0)
        dividend = holding.get('annual_dividend', 0)

        print(f'    {symbol}: {weight:.0f}% (${value:,.2f}) - '
              f'{shares:.2f}股 @ ${price:.2f} - 股息 ${dividend:.2f}')

except Exception as e:
    print(f'  ✗ 分析失败: {e}')
    import traceback
    traceback.print_exc()

# 3. 测试平衡型组合分析
print('\n3. 测试平衡型组合分析 (SCHD 30%, SPYD 20%, JEPQ 15%, JEPI 20%, VYM 15%):')
print('-' * 80)

try:
    result = etf_service.analyze_portfolio(
        allocation={
            'SCHD': 0.30,
            'SPYD': 0.20,
            'JEPQ': 0.15,
            'JEPI': 0.20,
            'VYM': 0.15
        },
        total_investment=50000
    )

    print(f'  投资金额: ${result.get("total_investment", 0):,.2f}')
    print(f'  当前价值: ${result.get("total_value", 0):,.2f}')
    print(f'  总收益: ${result.get("total_return", 0):,.2f}')
    print(f'  收益率: {result.get("total_return_percent", 0):+.2f}%')
    print(f'  年股息: ${result.get("total_dividend", 0):,.2f}')

    print('\n  持仓明细:')
    for holding in result.get('holdings', []):
        print(f'    {holding["symbol"]}: {holding["weight"]*100:.0f}% - '
              f'${holding.get("current_value", 0):,.2f}')

except Exception as e:
    print(f'  ✗ 分析失败: {e}')

# 4. 测试预设配置
print('\n4. 测试预设配置:')
print('-' * 80)

configs = PortfolioConfig.objects.filter(status=1)[:3]
for config in configs:
    try:
        result = etf_service.analyze_portfolio(
            allocation=config.allocation,
            total_investment=config.total_investment
        )

        print(f'  {config.name}:')
        print(f'    价值: ${result.get("total_value", 0):,.2f} | '
              f'收益: {result.get("total_return_percent", 0):+.2f}% | '
              f'股息: ${result.get("total_dividend", 0):,.2f}')
    except Exception as e:
        print(f'  {config.name}: ✗ 分析失败 - {e}')

# 5. 检查数据时效性
print('\n5. 数据时效性检查:')
print('-' * 80)
today = datetime.now().date()
for symbol in symbols:
    latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
    if latest:
        days_old = (today - latest.date).days
        status = '✓' if days_old <= 1 else '⚠️'
        print(f'  {status} {symbol}: {latest.date} ({days_old}天前)')

print('\n' + '='*80)
print('✓ 投资组合分析功能测试完成！')
print('='*80 + '\n')

# 提供访问建议
print('💡 使用建议:')
print('  1. 访问组合配置管理: http://localhost:8000/workflow/portfolio-config/')
print('  2. 查看投资组合分析: http://localhost:8000/workflow/portfolio/')
print('  3. 可以使用预设配置或创建自定义组合')
print('  4. 定期更新ETF数据以保持数据新鲜度')
print()
