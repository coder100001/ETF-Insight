"""
测试ETF配置操作记录功能
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFConfig, OperationLog
from workflow.operation_service import operation_service


def test_etf_operation_logs():
    """测试ETF配置操作记录"""
    print("="*60)
    print("测试：ETF配置操作记录功能")
    print("="*60)
    
    # 1. 查看当前操作记录数量
    print("\n【1】当前操作记录统计:")
    total_logs = OperationLog.objects.count()
    etf_logs = OperationLog.objects.filter(operation_type='etf_config').count()
    print(f"  总操作记录: {total_logs}")
    print(f"  ETF配置操作记录: {etf_logs}")
    
    # 2. 测试添加ETF操作记录
    print("\n【2】测试添加ETF操作记录:")
    test_symbol = 'TEST_LOG'
    
    # 先删除可能存在的测试数据
    ETFConfig.objects.filter(symbol=test_symbol).delete()
    
    # 模拟添加ETF操作（直接调用服务）
    op_log = operation_service.log_operation(
        operation_type='etf_config',
        operation_name=f'添加ETF配置: {test_symbol}',
        operator='test_user',
        input_params={
            'symbol': test_symbol,
            'name': '测试记录ETF',
            'market': 'US',
            'strategy': '测试策略',
            'status': '1',
        },
        ip_address='127.0.0.1',
        user_agent='test-agent'
    )
    
    # 创建ETF配置
    new_etf = ETFConfig.objects.create(
        symbol=test_symbol,
        name='测试记录ETF',
        market='US',
        strategy='测试策略',
        description='用于测试操作记录',
        focus='测试',
        expense_ratio=0.05,
        status=1,
        sort_order=999
    )
    
    # 完成操作记录
    operation_service.complete_operation(
        op_log,
        success=True,
        result={
            'id': new_etf.id,
            'symbol': new_etf.symbol,
            'name': new_etf.name,
            'market': new_etf.market
        }
    )
    
    print(f"  ✓ 成功记录添加操作: {new_etf.symbol}")
    
    # 3. 验证操作记录
    print("\n【3】验证操作记录:")
    recent_logs = OperationLog.objects.filter(
        operation_type='etf_config',
        operation_name__contains=test_symbol
    ).order_by('-start_time')[:5]
    
    for log in recent_logs:
        print(f"  - ID: {log.id}, 操作: {log.operation_name}")
        print(f"    类型: {log.operation_type}, 状态: {'成功' if log.status == 2 else '失败' if log.status == 3 else '进行中'}")
        print(f"    操作者: {log.operator}, IP: {log.ip_address}")
        print(f"    耗时: {log.duration_ms}ms, 时间: {log.start_time.strftime('%Y-%m-%d %H:%M:%S')}")
    
    # 4. 测试编辑操作记录
    print("\n【4】测试编辑ETF操作记录:")
    updated_etf = ETFConfig.objects.get(symbol=test_symbol)
    op_log = operation_service.log_operation(
        operation_type='etf_config',
        operation_name=f'编辑ETF配置: {updated_etf.symbol}',
        operator='test_user',
        input_params={
            'config_id': updated_etf.id,
            'symbol': updated_etf.symbol,
            'updates': {'name': '更新后的测试ETF', 'status': 0}
        },
        ip_address='127.0.0.1',
        user_agent='test-agent'
    )
    
    # 更新ETF
    updated_etf.name = '更新后的测试ETF'
    updated_etf.status = 0  # 禁用
    updated_etf.save()
    
    # 完成操作记录
    operation_service.complete_operation(
        op_log,
        success=True,
        result={'id': updated_etf.id, 'symbol': updated_etf.symbol}
    )
    
    print(f"  ✓ 成功记录编辑操作: {updated_etf.symbol}")
    
    # 5. 测试删除操作记录
    print("\n【5】测试删除ETF操作记录:")
    op_log = operation_service.log_operation(
        operation_type='etf_config',
        operation_name=f'删除ETF配置: {updated_etf.symbol}',
        operator='test_user',
        input_params={'config_id': updated_etf.id, 'symbol': updated_etf.symbol},
        ip_address='127.0.0.1',
        user_agent='test-agent'
    )
    
    # 删除ETF
    updated_etf.delete()
    
    # 完成操作记录
    operation_service.complete_operation(
        op_log,
        success=True,
        result={'symbol': test_symbol}
    )
    
    print(f"  ✓ 成功记录删除操作: {test_symbol}")
    
    # 6. 最终统计
    print("\n【6】最终操作记录统计:")
    new_total_logs = OperationLog.objects.count()
    new_etf_logs = OperationLog.objects.filter(operation_type='etf_config').count()
    print(f"  总操作记录: {new_total_logs} (+{new_total_logs - total_logs})")
    print(f"  ETF配置操作记录: {new_etf_logs} (+{new_etf_logs - etf_logs})")
    
    # 7. 清理测试数据
    ETFConfig.objects.filter(symbol=test_symbol).delete()
    
    print("\n" + "="*60)
    print("✓ ETF配置操作记录功能测试完成！")
    print("="*60)


def show_operation_types():
    """显示所有操作类型统计"""
    print("\n" + "="*60)
    print("操作类型统计")
    print("="*60)
    
    stats = operation_service.get_operation_stats(days=30)
    for stat in stats:
        print(f"  {stat['operation_type']}: {stat['count']} 次")
    
    print("\n最近的ETF配置操作:")
    recent_logs = OperationLog.objects.filter(
        operation_type='etf_config'
    ).order_by('-start_time')[:10]
    
    for log in recent_logs:
        status_text = {0: '进行中', 1: '完成', 2: '成功', 3: '失败'}.get(log.status, f'未知({log.status})')
        print(f"  [{log.start_time.strftime('%m-%d %H:%M:%S')}] {log.operation_name} - {status_text}")


if __name__ == '__main__':
    test_etf_operation_logs()
    show_operation_types()
