#!/usr/bin/env python
"""
初始化ETF配置数据
添加默认的3个美股ETF：SCHD、SPYD、JEPQ
"""

import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFConfig


def init_default_etfs():
    """初始化默认ETF配置"""
    default_etfs = [
        {
            'symbol': 'SCHD',
            'name': 'Schwab U.S. Dividend Equity ETF',
            'market': 'US',
            'strategy': '质量股息策略',
            'description': '追踪道琼斯美国股息100指数，投资高股息、财务稳健的美国公司',
            'focus': '质量+股息',
            'expense_ratio': 0.06,
            'status': 1,
            'sort_order': 1,
        },
        {
            'symbol': 'SPYD',
            'name': 'SPDR Portfolio S&P 500 High Dividend ETF',
            'market': 'US',
            'strategy': '高股息收益策略',
            'description': '追踪S&P 500中股息收益率最高的80只股票',
            'focus': '高股息',
            'expense_ratio': 0.07,
            'status': 1,
            'sort_order': 2,
        },
        {
            'symbol': 'JEPQ',
            'name': 'JPMorgan Nasdaq Equity Premium Income ETF',
            'market': 'US',
            'strategy': '期权增强收益策略',
            'description': '通过纳斯达克股票+卖出看涨期权获取增强收益',
            'focus': '增强收益',
            'expense_ratio': 0.35,
            'status': 1,
            'sort_order': 3,
        },
    ]
    
    created_count = 0
    updated_count = 0
    
    for etf_data in default_etfs:
        obj, created = ETFConfig.objects.update_or_create(
            symbol=etf_data['symbol'],
            defaults=etf_data
        )
        if created:
            created_count += 1
            print(f'✓ 创建: {obj.symbol} - {obj.name}')
        else:
            updated_count += 1
            print(f'✓ 更新: {obj.symbol} - {obj.name}')
    
    print(f'\n初始化完成！创建 {created_count} 个，更新 {updated_count} 个')
    print(f'当前总数: {ETFConfig.objects.count()} 个ETF配置')


if __name__ == '__main__':
    print('开始初始化ETF配置...\n')
    init_default_etfs()
