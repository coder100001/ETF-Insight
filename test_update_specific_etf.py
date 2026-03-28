"""
测试脚本：演示如何针对特定ETF更新数据

使用说明：
1. 使用 fetch_etf_data.py 脚本（独立运行）
   python fetch_etf_data.py SCHD         # 只更新SCHD
   python fetch_etf_data.py SCHD SPYD    # 更新SCHD和SPYD
   python fetch_etf_data.py              # 更新所有默认的ETF

2. 使用 Django 管理命令
   python manage.py update_etf           # 更新所有启用的ETF
   python manage.py update_etf SCHD      # 只更新SCHD
   python manage.py update_etf SCHD SPYD # 更新SCHD和SPYD
   python manage.py update_etf --all     # 明确指定更新所有

3. 在代码中调用
   见下面的示例代码
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.scheduler import update_etf_data
from workflow.services import etf_service


def test_update_specific_etf():
    """测试更新特定ETF"""
    print("="*60)
    print("测试：针对特定ETF更新数据")
    print("="*60)
    
    # 1. 查看当前启用的ETF
    print("\n【1】当前启用的ETF:")
    active_etfs = etf_service.get_active_etfs()
    for etf in active_etfs:
        print(f"  - {etf['symbol']}: {etf['name']} ({etf['market']})")
    
    # 2. 更新指定的单个ETF
    print("\n【2】测试：更新单个ETF (SCHD)")
    print("注意：由于可能遇到API频率限制，这里仅演示调用方式，不实际执行")
    print("实际调用代码：")
    print("  success = update_etf_data(symbols=['SCHD'])")
    
    # 3. 更新指定的多个ETF
    print("\n【3】测试：更新多个ETF (SCHD, SPYD)")
    print("实际调用代码：")
    print("  success = update_etf_data(symbols=['SCHD', 'SPYD'])")
    
    # 4. 更新所有启用的ETF
    print("\n【4】测试：更新所有启用的ETF")
    print("实际调用代码：")
    print("  success = update_etf_data()  # 或 update_etf_data(symbols=None)")
    
    print("\n" + "="*60)
    print("使用建议")
    print("="*60)
    print("""
1. 【fetch_etf_data.py】适合：
   - 不依赖Django的独立数据拉取
   - 导出CSV/Excel文件
   - 快速测试和验证

2. 【Django管理命令】适合：
   - 需要保存到数据库
   - 与工作流系统集成
   - 生产环境定时任务
   - 服务器上使用cron调度

3. 【代码直接调用】适合：
   - 自定义业务逻辑
   - API接口触发
   - 复杂的条件更新
    """)
    
    print("\n" + "="*60)
    print("命令示例")
    print("="*60)
    print("""
# 使用独立脚本（不依赖Django）
python fetch_etf_data.py SCHD
python fetch_etf_data.py SCHD SPYD JEPQ

# 使用Django管理命令（需要Django环境）
python manage.py update_etf SCHD
python manage.py update_etf SCHD SPYD
python manage.py update_etf --all

# 在Python代码中调用
from workflow.scheduler import update_etf_data
update_etf_data(symbols=['SCHD'])           # 更新单个
update_etf_data(symbols=['SCHD', 'SPYD'])   # 更新多个
update_etf_data()                           # 更新所有
    """)
    
    print("\n✓ 测试完成！")


if __name__ == '__main__':
    test_update_specific_etf()
