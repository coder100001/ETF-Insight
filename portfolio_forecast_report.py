"""
生成投资组合预测报告
包含所有要求的字段：税后股息、税前股息、年化收益率及资本所得、3/5/10年后的资产价值
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.services import etf_service


def generate_portfolio_forecast_report():
    """生成投资组合预测报告"""
    print("="*100)
    print("投资组合未来收益预测报告")
    print("="*100)
    
    # 创建一个示例投资组合
    allocation = {
        'SCHD': 0.4,  # 40% 投资SCHD
        'SPYD': 0.3,  # 30% 投资SPYD
        'JEPQ': 0.3   # 30% 投资JEPQ
    }
    
    print(f"投资组合配置:")
    for symbol, weight in allocation.items():
        print(f"  {symbol}: {weight*100:.1f}%")
    print(f"总投资金额: $100,000")
    print(f"股息税率: 10%")
    
    # 执行预测
    result = etf_service.forecast_portfolio_growth(
        allocation=allocation,
        total_investment=100000,
        tax_rate=0.10
    )
    
    print(f"\n当前组合加权股息率: {result['current_weighted_dividend_yield']:.2f}%")
    
    print("\n" + "="*100)
    print("未来收益预测详情")
    print("="*100)
    
    # 按场景展示
    for scenario_name, scenario_data in result['scenarios'].items():
        scenario_display = {
            'optimistic': '乐观情况',
            'neutral': '中性情况', 
            'pessimistic': '悲观情况'
        }.get(scenario_name, scenario_name)
        
        print(f"\n【{scenario_display}】年化收益率: {scenario_data['annual_return_rate']:.2f}%")
        print("-" * 80)
        
        # 展示所有要求的字段
        for years in [3, 5, 10]:
            data = scenario_data['years'][str(years)]
            print(f"  {years}年后预测:")
            print(f"    未来资产价值: ${data['future_value']:,.2f}")
            print(f"    资本增值: ${data['capital_appreciation']:,.2f}")
            print(f"    税前股息收入: ${data['total_dividend_before_tax']:,.2f}")
            print(f"    税后股息收入: ${data['total_dividend_after_tax']:,.2f}")
            print(f"    年化收益率及资本所得: {scenario_data['annual_return_rate']:.2f}% (年化收益率)")
            print(f"    年度税前股息: ${data['annual_dividend_before_tax']:,.2f}")
            print(f"    年度税后股息: ${data['annual_dividend_after_tax']:,.2f}")
            print(f"    股息税额: ${data['dividend_tax']:,.2f}")
            print(f"    税前总收益: ${data['total_return_before_tax']:,.2f}")
            print(f"    税后总收益: ${data['total_return_after_tax']:,.2f}")
            print()


def generate_comparison_table():
    """生成对比表格"""
    print("="*120)
    print("收益对比表格")
    print("="*120)
    
    allocation = {
        'SCHD': 0.4,
        'SPYD': 0.3,
        'JEPQ': 0.3
    }
    
    result = etf_service.forecast_portfolio_growth(
        allocation=allocation,
        total_investment=100000,
        tax_rate=0.10
    )
    
    # 创建表格头部
    print(f"{'':<12} {'':<15} {'未来资产价值':<15} {'资本增值':<15} {'税前股息收入':<15} {'税后股息收入':<15} {'年度税后股息':<15}")
    print("-" * 120)
    
    # 按年份和场景展示
    for years in [3, 5, 10]:
        print(f"\n{years}年后:")
        for scenario_name, scenario_data in result['scenarios'].items():
            scenario_display = {
                'optimistic': '乐观',
                'neutral': '中性', 
                'pessimistic': '悲观'
            }.get(scenario_name, scenario_name)
            
            data = scenario_data['years'][str(years)]
            print(f"  {scenario_display:<12} ({scenario_data['annual_return_rate']:.0f}%): "
                  f"{'$':>3}{data['future_value']:>11,.2f} "
                  f"{'$':>3}{data['capital_appreciation']:>11,.2f} "
                  f"{'$':>3}{data['total_dividend_before_tax']:>11,.2f} "
                  f"{'$':>3}{data['total_dividend_after_tax']:>11,.2f} "
                  f"{'$':>3}{data['annual_dividend_after_tax']:>11,.2f}")


def show_detailed_scenario_analysis():
    """展示详细的场景分析"""
    print("\n" + "="*100)
    print("详细场景分析")
    print("="*100)
    
    allocation = {
        'SCHD': 0.5,
        'SPYD': 0.5
    }
    
    # 使用自定义场景
    custom_scenarios = {
        'optimistic': 0.12,  # 乐观：年化12%
        'neutral': 0.08,     # 中性：年化8%
        'pessimistic': 0.03  # 悲观：年化3%
    }
    
    result = etf_service.forecast_portfolio_growth(
        allocation=allocation,
        total_investment=50000,
        tax_rate=0.10,
        scenarios=custom_scenarios
    )
    
    print(f"投资组合: {allocation}")
    print(f"初始投资: $50,000")
    print(f"当前加权股息率: {result['current_weighted_dividend_yield']:.2f}%")
    
    for scenario_name, scenario_data in result['scenarios'].items():
        scenario_display = {
            'optimistic': '乐观情况',
            'neutral': '中性情况', 
            'pessimistic': '悲观情况'
        }.get(scenario_name, scenario_name)
        
        print(f"\n{scenario_display} (年化{scenario_data['annual_return_rate']:.2f}%):")
        print(f"  - 3年后资产价值: ${scenario_data['years']['3']['future_value']:,.2f} (增长 {((scenario_data['years']['3']['future_value']/50000)-1)*100:.2f}%)")
        print(f"  - 5年后资产价值: ${scenario_data['years']['5']['future_value']:,.2f} (增长 {((scenario_data['years']['5']['future_value']/50000)-1)*100:.2f}%)")
        print(f"  - 10年后资产价值: ${scenario_data['years']['10']['future_value']:,.2f} (增长 {((scenario_data['years']['10']['future_value']/50000)-1)*100:.2f}%)")
        print(f"  - 税后年度股息: ${scenario_data['years']['10']['annual_dividend_after_tax']:,.2f}")
        print(f"  - 10年税后总股息: ${scenario_data['years']['10']['total_dividend_after_tax']:,.2f}")


if __name__ == '__main__':
    generate_portfolio_forecast_report()
    generate_comparison_table()
    show_detailed_scenario_analysis()
    
    print("\n" + "="*100)
    print("投资组合预测报告生成完成！")
    print("="*100)
    print("\n报告包含的字段:")
    print("✓ 税后股息 (Tax After Dividend)")
    print("✓ 税前股息 (Tax Before Dividend)")
    print("✓ 年化收益率及资本所得 (Annualized Return & Capital Gain)")
    print("✓ 3、5、10年后的资产价值 (Future Asset Value)")
    print("✓ 股息税额 (Dividend Tax)")
    print("✓ 总收益 (Total Return)")
    print("✓ 年度股息收入 (Annual Dividend Income)")