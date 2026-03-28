#!/usr/bin/env python
"""
测试空参数的投资组合分析
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__))))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from django.test import RequestFactory
from workflow.views import PortfolioAnalysisView

print('=' * 80)
print('测试空参数的投资组合分析')
print('=' * 80)

# 创建请求工厂
factory = RequestFactory()

# 测试1: 空字符串参数（之前会报错的情况）
print('\n测试1: 空字符串参数')
print('URL: ?investment=30000&schd=70&spyd=&jepq=0&jepi=30&vym=0')

try:
    request = factory.get('/workflow/portfolio/', {
        'investment': '30000',
        'schd': '70',
        'spyd': '',  # 空字符串
        'jepq': '0',
        'jepi': '30',
        'vym': '0'
    })
    view = PortfolioAnalysisView.as_view()
    response = view(request)
    print(f'  状态码: {response.status_code}')
    print(f'  结果: ✅ 成功')
except Exception as e:
    print(f'  错误: ❌ {str(e)}')

# 测试2: 只有部分ETF有值
print('\n测试2: 只有部分ETF有值')
print('URL: ?investment=50000&schd=50&jepq=50')

try:
    request = factory.get('/workflow/portfolio/', {
        'investment': '50000',
        'schd': '50',
        'jepq': '50'
        # 其他ETF未提供
    })
    view = PortfolioAnalysisView.as_view()
    response = view(request)
    print(f'  状态码: {response.status_code}')
    print(f'  结果: ✅ 成功')
except Exception as e:
    print(f'  错误: ❌ {str(e)}')

# 测试3: 所有ETF都为空
print('\n测试3: 所有ETF都为空')
print('URL: ?investment=10000')

try:
    request = factory.get('/workflow/portfolio/', {
        'investment': '10000',
        'schd': '',
        'spyd': '',
        'jepq': '',
        'jepi': '',
        'vym': ''
    })
    view = PortfolioAnalysisView.as_view()
    response = view(request)
    print(f'  状态码: {response.status_code}')
    print(f'  结果: ✅ 成功（应该使用默认组合）')
except Exception as e:
    print(f'  错误: ❌ {str(e)}')

# 测试4: 正常参数
print('\n测试4: 正常参数')
print('URL: ?investment=10000&schd=40&spyd=30&jepq=30')

try:
    request = factory.get('/workflow/portfolio/', {
        'investment': '10000',
        'schd': '40',
        'spyd': '30',
        'jepq': '30'
    })
    view = PortfolioAnalysisView.as_view()
    response = view(request)
    print(f'  状态码: {response.status_code}')
    print(f'  结果: ✅ 成功')
except Exception as e:
    print(f'  错误: ❌ {str(e)}')

print('\n' + '=' * 80)
print('所有测试完成')
print('=' * 80)
