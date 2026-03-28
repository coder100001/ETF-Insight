#!/usr/bin/env python
"""
初始化预设的投资组合配置
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import PortfolioConfig

# 预设组合配置
PRESET_PORTFOLIOS = [
    {
        'name': '433组合',
        'description': 'SCHD 40%、SPYD 30%、JEPQ 30%的保守配置，适合稳健型投资者',
        'total_investment': 10000,
        'allocation': {
            'SCHD': 0.40,
            'SPYD': 0.30,
            'JEPQ': 0.30,
            'JEPI': 0.0,
            'VYM': 0.0
        },
        'status': 1
    },
    {
        'name': '442组合',
        'description': 'SCHD 40%、SPYD 40%、JEPQ 20%的平衡配置，兼顾收益与风险',
        'total_investment': 10000,
        'allocation': {
            'SCHD': 0.40,
            'SPYD': 0.40,
            'JEPQ': 0.20,
            'JEPI': 0.0,
            'VYM': 0.0
        },
        'status': 1
    },
    {
        'name': '平衡型组合',
        'description': '多ETF均衡配置，分散投资风险',
        'total_investment': 10000,
        'allocation': {
            'SCHD': 0.30,
            'SPYD': 0.20,
            'JEPQ': 0.15,
            'JEPI': 0.20,
            'VYM': 0.15
        },
        'status': 1
    },
    {
        'name': '稳健型组合',
        'description': '以高股息ETF为主的稳健投资策略',
        'total_investment': 10000,
        'allocation': {
            'SCHD': 0.40,
            'SPYD': 0.20,
            'JEPQ': 0.10,
            'JEPI': 0.20,
            'VYM': 0.10
        },
        'status': 1
    },
    {
        'name': '激进型组合',
        'description': '以JEPQ为主的高收益高风险配置',
        'total_investment': 10000,
        'allocation': {
            'SCHD': 0.20,
            'SPYD': 0.20,
            'JEPQ': 0.60,
            'JEPI': 0.0,
            'VYM': 0.0
        },
        'status': 1
    },
    {
        'name': '纯股息组合',
        'description': '专注于股息收益的组合配置',
        'total_investment': 10000,
        'allocation': {
            'SCHD': 0.35,
            'SPYD': 0.35,
            'JEPQ': 0.0,
            'JEPI': 0.0,
            'VYM': 0.30
        },
        'status': 1
    }
]


def init_portfolio_configs():
    """初始化预设组合配置"""
    print('=' * 60)
    print('初始化预设投资组合配置')
    print('=' * 60)
    
    created_count = 0
    updated_count = 0
    
    for portfolio_data in PRESET_PORTFOLIOS:
        name = portfolio_data['name']
        
        # 检查是否已存在
        existing = PortfolioConfig.objects.filter(name=name).first()
        
        if existing:
            # 更新现有配置
            existing.description = portfolio_data['description']
            existing.total_investment = portfolio_data['total_investment']
            existing.allocation = portfolio_data['allocation']
            existing.status = portfolio_data['status']
            existing.save()
            print(f'✓ 更新组合配置: {name}')
            updated_count += 1
        else:
            # 创建新配置
            PortfolioConfig.objects.create(**portfolio_data)
            print(f'✓ 创建组合配置: {name}')
            created_count += 1
    
    print('=' * 60)
    print(f'初始化完成！')
    print(f'创建新配置: {created_count}个')
    print(f'更新现有配置: {updated_count}个')
    print(f'总配置数: {PortfolioConfig.objects.count()}个')
    print('=' * 60)


if __name__ == '__main__':
    init_portfolio_configs()
