#!/usr/bin/env python
"""
测试组合配置管理功能
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import PortfolioConfig
from workflow.services import etf_service
from django.db.models import Avg
import json

print('=' * 60)
print('测试组合配置管理功能')
print('=' * 60)

# 1. 测试读取所有配置
print('\n1. 读取所有组合配置:')
configs = PortfolioConfig.objects.all()
for config in configs:
    print(f'   - {config.name}: ${config.total_investment} ({len(config.allocation)}个ETF)')

# 2. 测试获取单个配置
print('\n2. 获取单个配置详情:')
config = PortfolioConfig.objects.filter(name='433组合').first()
if config:
    print(f'   组合名称: {config.name}')
    print(f'   描述: {config.description}')
    print(f'   投资金额: ${config.total_investment}')
    print(f'   权重配置: {json.dumps(config.allocation, indent=6)}')

# 3. 测试组合分析
print('\n3. 测试组合分析:')
if config:
    try:
        result = etf_service.analyze_portfolio(
            allocation=config.allocation,
            total_investment=config.total_investment
        )
        
        print(f'   总投资: ${result.get("total_investment", 0):.2f}')
        print(f'   当前价值: ${result.get("total_value", 0):.2f}')
        print(f'   总收益: ${result.get("total_return", 0):.2f}')
        print(f'   收益率: {result.get("total_return_percent", 0):.2f}%')
        print(f'   年股息: ${result.get("total_dividend", 0):.2f}')
        print(f'   加权股息率: {result.get("weighted_dividend_yield", 0):.2f}%')
        
        print('\n   持仓明细:')
        for holding in result.get('holdings', []):
            print(f'   - {holding["symbol"]}: {holding["weight"]*100:.0f}% '
                  f'(${holding["current_value"]:.2f})')
    except Exception as e:
        print(f'   ✗ 分析失败: {e}')

# 4. 测试创建自定义配置
print('\n4. 测试创建自定义配置:')
try:
    custom_config = PortfolioConfig.objects.create(
        name='测试组合',
        description='用于测试的自定义组合',
        total_investment=50000,
        allocation={
            'SCHD': 0.25,
            'SPYD': 0.25,
            'JEPQ': 0.25,
            'JEPI': 0.25
        },
        status=1
    )
    print(f'   ✓ 创建成功: {custom_config.name} (ID: {custom_config.id})')
    
    # 分析自定义配置
    result = etf_service.analyze_portfolio(
        allocation=custom_config.allocation,
        total_investment=custom_config.total_investment
    )
    print(f'   组合价值: ${result.get("total_value", 0):.2f}')
    
    # 清理测试数据
    custom_config.delete()
    print(f'   ✓ 已删除测试配置')
    
except Exception as e:
    print(f'   ✗ 创建失败: {e}')

# 5. 测试配置状态切换
print('\n5. 测试配置状态切换:')
config = PortfolioConfig.objects.filter(name='激进型组合').first()
if config:
    original_status = config.status
    config.status = 0 if config.status == 1 else 1
    config.save()
    print(f'   ✓ {config.name} 状态已切换: {original_status} -> {config.status}')
    
    # 恢复状态
    config.status = 1
    config.save()

# 6. 测试统计功能
print('\n6. 统计功能:')
total_configs = PortfolioConfig.objects.count()
active_configs = PortfolioConfig.objects.filter(status=1).count()
print(f'   总配置数: {total_configs}')
print(f'   启用配置数: {active_configs}')
print(f'   禁用配置数: {total_configs - active_configs}')

# 计算平均投资金额
avg_investment = PortfolioConfig.objects.aggregate(
    avg_investment=Avg('total_investment')
)['avg_investment']
print(f'   平均投资金额: ${avg_investment:.2f}' if avg_investment else '   平均投资金额: $0.00')

print('\n' + '=' * 60)
print('✓ 所有测试完成！')
print('=' * 60)
