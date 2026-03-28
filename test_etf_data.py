"""
测试ETF数据获取功能
"""

import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.services import etf_service


def test_realtime_data():
    """测试实时数据获取"""
    print("="*60)
    print("测试：ETF实时数据获取")
    print("="*60)
    
    for symbol in etf_service.SYMBOLS:
        print(f"\n获取 {symbol} 实时数据:")
        try:
            data = etf_service.fetch_realtime_data(symbol)
            if data and 'error' not in data:
                print(f"  名称: {data.get('name', 'N/A')}")
                print(f"  价格: ${data.get('current_price', 'N/A')}")
                print(f"  涨跌: {data.get('change_percent', 'N/A')}%")
                print(f"  股息率: {data.get('dividend_yield', 'N/A')}%")
                print(f"  数据来源: {data.get('data_source', 'N/A')}")
            else:
                print(f"  数据获取失败: {data.get('error', 'Unknown error') if data else 'No data'}")
        except Exception as e:
            print(f"  错误: {e}")


def test_historical_data():
    """测试历史数据获取"""
    print("\n" + "="*60)
    print("测试：ETF历史数据获取")
    print("="*60)
    
    for symbol in etf_service.SYMBOLS[:2]:  # 只测试前两个，避免API限制
        print(f"\n获取 {symbol} 历史数据 (1个月):")
        try:
            data = etf_service.fetch_historical_data(symbol, period='1mo')
            if data is not None and not data.empty:
                print(f"  数据点数: {len(data)}")
                print(f"  日期范围: {data.index[0]} 到 {data.index[-1]}")
                print(f"  最新价格: ${data['Close'][-1]:.2f}")
            else:
                print("  无历史数据")
        except Exception as e:
            print(f"  错误: {e}")


def test_metrics():
    """测试指标计算"""
    print("\n" + "="*60)
    print("测试：ETF指标计算")
    print("="*60)
    
    for symbol in etf_service.SYMBOLS[:2]:  # 只测试前两个
        print(f"\n计算 {symbol} 指标 (1年):")
        try:
            metrics = etf_service.calculate_metrics(symbol, '1y')
            if metrics:
                print(f"  年化收益率: {metrics.get('annual_return', 'N/A')}%")
                print(f"  波动率: {metrics.get('volatility', 'N/A')}%")
                print(f"  夏普比率: {metrics.get('sharpe_ratio', 'N/A')}")
                print(f"  最大回撤: {metrics.get('max_drawdown', 'N/A')}%")
            else:
                print("  无法计算指标")
        except Exception as e:
            print(f"  错误: {e}")


def test_comparison():
    """测试ETF对比数据"""
    print("\n" + "="*60)
    print("测试：ETF对比数据")
    print("="*60)
    
    try:
        comparison = etf_service.get_comparison_data('1y')
        if comparison:
            print(f"  获取到 {len(comparison)} 个ETF的对比数据")
            for etf in comparison:
                print(f"  {etf.get('symbol', 'N/A')}: 年化收益 {etf.get('annual_return', 'N/A')}%")
        else:
            print("  无对比数据")
    except Exception as e:
        print(f"  错误: {e}")


def test_portfolio_analysis():
    """测试投资组合分析"""
    print("\n" + "="*60)
    print("测试：投资组合分析")
    print("="*60)
    
    try:
        # 创建一个简单的分配比例
        allocation = {}
        symbols = etf_service.SYMBOLS[:3]  # 使用前3个ETF
        weight = 1.0 / len(symbols) if symbols else 0
        for symbol in symbols:
            allocation[symbol] = weight
        
        print(f"  配置: {allocation}")
        
        if allocation:
            result = etf_service.analyze_portfolio(allocation, 10000)  # 1万美元投资
            print(f"  总投资: ${result.get('total_investment', 0):,.2f}")
            print(f"  当前价值: ${result.get('total_value', 0):,.2f}")
            print(f"  总收益: ${result.get('total_return', 0):,.2f}")
            print(f"  年股息: ${result.get('annual_dividend_after_tax', 0):,.2f}")
        else:
            print("  无可用ETF进行组合分析")
            
    except Exception as e:
        print(f"  错误: {e}")


if __name__ == '__main__':
    test_realtime_data()
    test_historical_data()
    test_metrics()
    test_comparison()
    test_portfolio_analysis()
    
    print("\n" + "="*60)
    print("ETF数据获取功能测试完成！")
    print("="*60)