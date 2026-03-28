"""
测试多货币投资组合分析
"""
from workflow.services import etf_service

def test_multicurrency_portfolio():
    """测试多货币投资组合分析"""
    print("测试：多货币投资组合分析（美元计价）")
    print("=" * 60)

    # 创建一个包含美股、A股、港股的投资组合
    allocation = {
        'SCHD': 0.3,      # 美股
        'SPYD': 0.2,      # 美股
        '510300': 0.2,    # A股
        '3466.HK': 0.3,   # 港股
    }

    total_investment = 10000  # 1万美元投资

    print(f"投资组合配置:")
    for symbol, weight in allocation.items():
        print(f"  {symbol}: {weight * 100:.0f}%")
    print(f"总投资金额: ${total_investment:,.2f} USD")
    print()

    # 分析投资组合
    result = etf_service.analyze_portfolio(allocation, total_investment)

    if result.get('holdings'):
        print("投资组合分析结果:")
        print("-" * 60)

        # 显示每个持仓
        for holding in result['holdings']:
            print(f"\n{holding['symbol']} - {holding['name']}")
            print(f"  计价货币: {holding['currency']}")
            print(f"  权重: {holding['weight']:.1f}%")
            print(f"  投资金额: {holding['investment']:,.2f} {holding['currency']} = ${holding['investment_usd']:,.2f} USD")
            print(f"  当前价值: {holding['current_value']:,.2f} {holding['currency']} = ${holding['current_value_usd']:,.2f} USD")
            print(f"  股息率: {holding['dividend_yield']:.2f}%")
            print(f"  年股息(税前): {holding['annual_dividend_before_tax']:,.2f} {holding['currency']} = ${holding['annual_dividend_before_tax_usd']:,.2f} USD")
            print(f"  年股息(税后): {holding['annual_dividend_after_tax']:,.2f} {holding['currency']} = ${holding['annual_dividend_after_tax_usd']:,.2f} USD")

        print("\n" + "-" * 60)
        print(f"总投资价值: ${result['total_value_usd']:,.2f} USD")
        print(f"加权股息率: {result['weighted_dividend_yield']:.2f}%")
        print(f"年股息收入(税前): ${result['annual_dividend_before_tax']:,.2f} USD")
        print(f"年股息收入(税后): ${result['annual_dividend_after_tax']:,.2f} USD")
        print(f"股息税额: ${result['dividend_tax']:,.2f} USD")

        # 测试汇率功能
        if result.get('exchange_rates'):
            print("\n使用的汇率:")
            for currency_pair, rate in result['exchange_rates'].items():
                print(f"  1 {currency_pair.replace('_to_USD', '')} = {rate:.6f} USD")

        print(f"\n预期资本利得: ${result['total_return']:,.2f} USD")
        print(f"预期年化收益率: {result['total_return_percent']:.2f}%")
        print(f"综合收益(含股息): ${result['total_return_with_dividend']:,.2f} USD")
        print(f"综合年化收益率: {result['total_return_with_dividend_percent']:.2f}%")
    else:
        print("  无可用ETF进行组合分析")

    print("\n" + "=" * 60)

    # 测试组合预测功能
    print("\n测试：多货币投资组合预测（美元计价）")
    print("=" * 60)

    forecast_result = etf_service.forecast_portfolio_growth(allocation, total_investment)

    print(f"总投资金额: ${forecast_result['total_investment']:,.2f} USD")
    print(f"加权股息率: {forecast_result['current_weighted_dividend_yield']:.2f}%")
    print(f"基础货币: {forecast_result['base_currency']}")
    print()

    # 显示每个场景的预测结果
    for scenario_name, scenario_data in forecast_result['scenarios'].items():
        print(f"{scenario_name.upper()}情况 (年化收益率: {scenario_data['annual_return_rate']:.1f}%):")
        for years, year_data in scenario_data['years'].items():
            print(f"  {years}年后:")
            print(f"    资产价值: ${year_data['future_value']:,.2f} USD")
            print(f"    资本增值: ${year_data['capital_appreciation']:,.2f} USD")
            print(f"    总股息(税前): ${year_data['total_dividend_before_tax']:,.2f} USD")
            print(f"    总股息(税后): ${year_data['total_dividend_after_tax']:,.2f} USD")
            print(f"    股息税: ${year_data['dividend_tax']:,.2f} USD")
            print(f"    综合收益(税后): ${year_data['total_return_after_tax']:,.2f} USD")
        print()

    print("=" * 60)

if __name__ == '__main__':
    test_multicurrency_portfolio()
