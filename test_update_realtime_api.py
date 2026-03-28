#!/usr/bin/env python
"""
测试实时数据更新API
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from django.test import Client
import json

# 创建测试客户端
client = Client()

# 准备测试数据
test_data = {
    'allocation': {
        'SCHD': 0.4,
        'SPYD': 0.3,
        'JEPQ': 0.3,
        'JEPI': 0.0,
        'VYM': 0.0
    },
    'total_investment': 10000
}

# 发送POST请求
response = client.post(
    '/workflow/api/update-realtime/',
    data=json.dumps(test_data),
    content_type='application/json',
    HTTP_X_CSRFTOKEN='test'  # 测试时可以使用简单的token
)

print("状态码:", response.status_code)
print("响应内容:", response.json() if response.status_code == 200 else response.content)
