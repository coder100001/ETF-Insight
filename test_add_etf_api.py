"""
测试添加ETF功能的脚本
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFConfig


def test_add_etf():
    """测试添加ETF功能"""
    print("="*60)
    print("测试：添加ETF功能")
    print("="*60)
    
    # 1. 查看当前ETF数量
    print("\n【1】当前ETF配置数量:")
    total = ETFConfig.objects.count()
    us_count = ETFConfig.objects.filter(market='US').count()
    cn_count = ETFConfig.objects.filter(market='CN').count()
    print(f"  总数: {total}")
    print(f"  美股: {us_count}")
    print(f"  A股: {cn_count}")
    
    # 2. 测试添加功能（模拟后端逻辑）
    print("\n【2】测试添加新ETF (模拟后端逻辑):")
    test_symbol = 'TEST_ETF'
    
    # 先删除可能存在的测试数据
    ETFConfig.objects.filter(symbol=test_symbol).delete()
    
    try:
        # 创建新ETF配置（模拟POST请求）
        new_etf = ETFConfig.objects.create(
            symbol=test_symbol,
            name='测试ETF',
            market='US',
            strategy='测试策略',
            description='这是一个测试ETF',
            focus='测试',
            expense_ratio=0.05,
            status=1,
            sort_order=999
        )
        
        print(f"  ✓ 成功创建: {new_etf.symbol} - {new_etf.name}")
        print(f"  ID: {new_etf.id}")
        print(f"  市场: {new_etf.market}")
        print(f"  状态: {'启用' if new_etf.status == 1 else '禁用'}")
        
        # 3. 验证是否添加成功
        print("\n【3】验证添加结果:")
        saved_etf = ETFConfig.objects.get(symbol=test_symbol)
        print(f"  ✓ 从数据库读取成功: {saved_etf.symbol}")
        
        # 4. 清理测试数据
        print("\n【4】清理测试数据:")
        saved_etf.delete()
        print(f"  ✓ 已删除测试ETF: {test_symbol}")
        
        print("\n" + "="*60)
        print("✓ 添加ETF功能测试通过！")
        print("="*60)
        
    except Exception as e:
        print(f"\n✗ 测试失败: {e}")
        import traceback
        traceback.print_exc()


def test_frontend_workflow():
    """演示前端添加流程"""
    print("\n" + "="*60)
    print("前端添加ETF的完整流程")
    print("="*60)
    
    print("""
1. 用户操作：
   - 访问页面: http://localhost:8000/workflow/etf-config/
   - 点击右上角"添加ETF"按钮
   - 弹出模态框

2. 填写表单字段：
   必填项：
   - ETF代码: 如 VTI
   - ETF名称: 如 Vanguard Total Stock Market ETF
   - 市场: 选择 美股(US) 或 A股(CN)
   
   可选项：
   - 策略类型: 如 全市场策略
   - 焦点领域: 如 全市场
   - 费率(%): 如 0.03
   - 描述: ETF详细描述
   - 排序: 数字，越小越靠前
   - 状态: 启用/禁用

3. 提交表单：
   - 点击"保存"按钮
   - JavaScript发送POST请求到 /workflow/etf-config/
   - 后端创建新记录
   - 返回成功消息
   - 页面自动刷新

4. 验证结果：
   - 新ETF应该出现在列表中
   - 如果选择启用，投资组合分析和对比分析页面会包含该ETF
    """)
    
    print("="*60)
    print("后端API接口")
    print("="*60)
    print("""
POST /workflow/etf-config/
Content-Type: multipart/form-data

参数：
- symbol: ETF代码（必填，会自动转大写）
- name: ETF名称（必填）
- market: 市场类型（必填，US或CN）
- strategy: 策略类型（可选）
- focus: 焦点领域（可选）
- expense_ratio: 费率（可选，数字）
- description: 描述（可选）
- sort_order: 排序（可选，数字，默认0）
- status: 状态（可选，1启用/0禁用，默认1）

返回（成功）：
{
    "success": true,
    "message": "成功添加 VTI",
    "data": {
        "id": 6,
        "symbol": "VTI",
        "name": "Vanguard Total Stock Market ETF"
    }
}

返回（失败）：
{
    "success": false,
    "message": "错误信息"
}
    """)


if __name__ == '__main__':
    test_add_etf()
    test_frontend_workflow()
