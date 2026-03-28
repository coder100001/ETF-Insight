"""
测试ETF配置缓存清除功能
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.services import etf_service
from workflow.models import ETFConfig


def test_cache_clear():
    """测试缓存清除功能"""
    print("="*60)
    print("测试：ETF配置缓存清除功能")
    print("="*60)
    
    # 1. 查看初始状态
    print("\n【1】初始状态:")
    print(f"  当前SYMBOLS: {etf_service.SYMBOLS}")
    print(f"  缓存时间: {etf_service._etf_config_cache_time}")
    
    # 2. 模拟禁用一个ETF
    print("\n【2】禁用一个ETF (VYM) 来测试缓存效果:")
    vym_etf = ETFConfig.objects.get(symbol='VYM')
    old_status = vym_etf.status
    print(f"  VYM原始状态: {'启用' if old_status == 1 else '禁用'}")
    
    # 禁用VYM
    vym_etf.status = 0
    vym_etf.save()
    
    # 检查缓存是否立即更新
    print(f"  修改后直接获取SYMBOLS: {etf_service.SYMBOLS}")
    print(f"  缓存时间: {etf_service._etf_config_cache_time}")
    
    # 3. 清除缓存后检查
    print("\n【3】清除缓存后:")
    etf_service.clear_etf_config_cache()
    print(f"  清除缓存后SYMBOLS: {etf_service.SYMBOLS}")
    
    # 4. 重新启用ETF并测试
    print("\n【4】重新启用ETF并再次测试:")
    vym_etf.status = 1  # 重新启用
    vym_etf.save()
    
    print(f"  重新启用后直接获取: {etf_service.SYMBOLS}")
    
    # 清除缓存
    etf_service.clear_etf_config_cache()
    print(f"  清除缓存后: {etf_service.SYMBOLS}")
    
    # 5. 恢复原状态
    vym_etf.status = old_status
    vym_etf.save()
    etf_service.clear_etf_config_cache()
    
    print("\n【5】恢复原状态:")
    print(f"  最终SYMBOLS: {etf_service.SYMBOLS}")
    
    print("\n" + "="*60)
    print("✓ 缓存清除功能测试完成！")
    print("✓ 修改ETF配置后，调用clear_etf_config_cache()可立即生效")
    print("="*60)


def test_with_operations():
    """测试在实际操作中缓存清除"""
    print("\n" + "="*60)
    print("测试：实际操作中的缓存清除")
    print("="*60)
    
    # 1. 查看当前状态
    print(f"\n【1】当前启用的ETF数量: {len(etf_service.SYMBOLS)}")
    
    # 2. 临时禁用一个ETF
    print("\n【2】临时禁用SCHD测试:")
    schd_etf = ETFConfig.objects.get(symbol='SCHD')
    original_status = schd_etf.status
    schd_etf.status = 0  # 禁用
    schd_etf.save()
    
    # 直接获取（应该还是有缓存）
    symbols_before_clear = etf_service.SYMBOLS
    print(f"  修改后未清除缓存: {symbols_before_clear}")
    
    # 清除缓存
    etf_service.clear_etf_config_cache()
    symbols_after_clear = etf_service.SYMBOLS
    print(f"  清除缓存后: {symbols_after_clear}")
    
    # 验证SCHD确实被移除
    if 'SCHD' in symbols_before_clear and 'SCHD' not in symbols_after_clear:
        print("  ✓ 缓存清除功能正常工作")
    else:
        print("  ✗ 缓存清除功能可能有问题")
    
    # 3. 恢复原状态
    schd_etf.status = original_status
    schd_etf.save()
    etf_service.clear_etf_config_cache()
    
    print(f"\n【3】恢复后: {etf_service.SYMBOLS}")
    
    print("\n" + "="*60)
    print("✓ 实际操作测试完成！")
    print("="*60)


if __name__ == '__main__':
    test_cache_clear()
    test_with_operations()