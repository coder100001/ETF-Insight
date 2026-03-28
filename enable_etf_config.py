#!/usr/bin/env python
"""
启用JEPI和VYM ETF配置
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFConfig

symbols = ['JEPI', 'VYM']

for symbol in symbols:
    config = ETFConfig.objects.filter(symbol=symbol).first()
    if config:
        config.status = 1
        config.save()
        print(f'✓ 已启用: {config.symbol} - {config.name}')
    else:
        print(f'✗ 未找到配置: {symbol}')
