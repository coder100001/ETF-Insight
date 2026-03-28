"""
检查ETF实时数据获取功能
"""

import os
import sys
import django
import time

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.services import etf_service


def check_realtime_data():
    """检查实时数据获取"""
    print("="*60)
    print("检查ETF实时数据获取功能")
    print("="*60)
    
    print(f"当前启用的ETF: {etf_service.SYMBOLS}")
    
    for symbol in etf_service.SYMBOLS:
        print(f"\n获取 {symbol} 实时数据:")
        try:
            start_time = time.time()
            data = etf_service.fetch_realtime_data(symbol)
            end_time = time.time()
            
            if data and 'error' not in data:
                print(f"  ✓ 价格: ${data.get('current_price', 'N/A')}")
                print(f"  ✓ 股息率: {data.get('dividend_yield', 'N/A')}%")
                print(f"  ✓ 涨跌: {data.get('change_percent', 'N/A')}%")
                print(f"  ✓ 数据来源: {data.get('data_source', 'N/A')}")
                print(f"  ✓ 数据日期: {data.get('data_date', 'N/A')}")
                print(f"  ✓ 耗时: {(end_time - start_time)*1000:.2f}ms")
            else:
                print(f"  ✗ 数据获取失败: {data.get('error', 'Unknown error') if data else 'No data'}")
        except Exception as e:
            print(f"  ✗ 错误: {e}")
    
    print("\n" + "="*60)
    print("检查完成")
    print("="*60)


def check_cache_status():
    """检查缓存状态"""
    print("\n" + "="*60)
    print("检查缓存状态")
    print("="*60)
    
    from workflow.cache_manager import etf_cache
    
    stats = etf_cache.get_cache_stats()
    if stats:
        print(f"缓存统计:")
        print(f"  实时数据缓存: {stats.get('realtime_count', 0)}")
        print(f"  历史数据缓存: {stats.get('historical_count', 0)}")
        print(f"  指标数据缓存: {stats.get('metrics_count', 0)}")
        print(f"  对比数据缓存: {stats.get('comparison_count', 0)}")
        print(f"  总计: {stats.get('total_count', 0)}")
    else:
        print("无法获取缓存统计信息")


def check_database_data():
    """检查数据库中的数据"""
    print("\n" + "="*60)
    print("检查数据库中的ETF数据")
    print("="*60)
    
    from workflow.models import ETFData
    
    for symbol in etf_service.SYMBOLS:
        latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
        if latest:
            print(f"{symbol}: 最新数据日期 {latest.date}, 价格 ${latest.close_price}")
        else:
            print(f"{symbol}: 无数据库记录")


if __name__ == '__main__':
    check_realtime_data()
    check_cache_status()
    check_database_data()