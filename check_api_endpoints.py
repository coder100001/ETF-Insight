#!/usr/bin/env python
"""
检查API端点状态
"""
import os
import django

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from django.urls import get_resolver

print("=" * 60)
print("API端点检查")
print("=" * 60)

resolver = get_resolver()

# 检查关键API端点
api_endpoints = [
    'api_update_realtime',
    'api_update_exchange_rates',
    'portfolio_analysis',
]

print("\n关键API端点状态:")
for endpoint_name in api_endpoints:
    try:
        url_pattern = resolver.reverse_dict.get(endpoint_name)
        if url_pattern:
            print(f"  ✓ {endpoint_name}: 已配置")
        else:
            print(f"  ✗ {endpoint_name}: 未找到")
    except Exception as e:
        print(f"  ✗ {endpoint_name}: 错误 - {str(e)}")

# 检查URL配置
print("\n所有API URL:")
for pattern in resolver.url_patterns:
    if hasattr(pattern, 'url_patterns'):
        for sub_pattern in pattern.url_patterns:
            if 'api' in str(sub_pattern.pattern):
                print(f"  - {sub_pattern.pattern}")
    elif 'api' in str(pattern.pattern):
        print(f"  - {pattern.pattern}")

print("\n" + "=" * 60)
print("API端点检查完成")
print("=" * 60)
