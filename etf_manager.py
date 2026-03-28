#!/usr/bin/env python
"""
ETF 数据管理主脚本
整合所有功能的统一入口

用法:
    python etf_manager.py update              # 更新所有 ETF
    python etf_manager.py update SCHD SPYD    # 更新指定 ETF
    python etf_manager.py realtime            # 获取实时数据
    python etf_manager.py schedule start      # 启动定时更新
    python etf_manager.py status              # 查看状态
    python etf_manager.py export              # 导出数据
"""

import os
import sys

# 添加项目路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# 设置 Django 环境（用于数据库操作）
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')

try:
    import django
    django.setup()
except Exception as e:
    print(f"警告: Django 环境初始化失败: {e}")
    print("某些功能可能不可用")

from core.cli import main

if __name__ == '__main__':
    sys.exit(main())
