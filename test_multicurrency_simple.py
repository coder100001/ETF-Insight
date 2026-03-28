"""
测试多货币投资组合分析（通过Django shell执行）
"""
from workflow.services import etf_service

# 创建一个包含美股、A股、港股的投资组合
allocation = {
    'SCHD': 0.3,      # 美股
    'SPYD': 0.2,      # 美股
    '510300': 0.2,    # A股
    '3466.HK': 0.3,   # 港股
}

total_investment = 10000  # 1万美元投资

print("投资组合配置:")
for symbol, weight in allocation.items():
    print(f"  {symbol}: {weight * 100:.0f}%")
print(f"总投资金额: ${total_investment:,.2f} USD")
print()

# 测试汇率功能
print("测试汇率功能:")
print(f"  1 CNY = {etf_service.get_exchange_rate('CNY', 'USD'):.6f} USD")
print(f"  1 HKD = {etf_service.get_exchange_rate('HKD', 'USD'):.6f} USD")
print(f"  1 USD = {etf_service.get_exchange_rate('USD', 'CNY'):.6f} CNY")
print(f"  1 USD = {etf_service.get_exchange_rate('USD', 'HKD'):.6f} HKD")
print()

# 获取每个ETF的货币
print("ETF计价货币:")
for symbol in allocation.keys():
    currency = etf_service.get_etf_currency(symbol)
    print(f"  {symbol}: {currency}")
print()

# 分析投资组合
result = etf_service.analyze_portfolio(allocation, total_investment)

if result.get('holdings'):
    print("投资组合分析结果:")
    for holding in result['holdings']:
        print(f"\n{holding['symbol']} - {holding['name']}")
        print(f"  计价货币: {holding['currency']}")
        print(f"  投资金额: {holding['investment']:,.2f} {holding['currency']} = ${holding['investment_usd']:,.2f} USD")
        print(f"  当前价值: ${holding['current_value_usd']:,.2f} USD")
        print(f"  年股息(税后): ${holding['annual_dividend_after_tax_usd']:,.2f} USD")

    print(f"\n总投资价值(美元): ${result['total_value_usd']:,.2f} USD")
    print(f"加权股息率: {result['weighted_dividend_yield']:.2f}%")
    print(f"年股息收入(税后): ${result['annual_dividend_after_tax']:,.2f} USD")

    if result.get('exchange_rates'):
        print("\n使用的汇率:")
        for currency_pair, rate in result['exchange_rates'].items():
            print(f"  1 {currency_pair.replace('_to_USD', '')} = {rate:.6f} USD")

    print(f"\n预期资本利得: ${result['total_return']:,.2f} USD")
    print(f"综合收益(含股息): ${result['total_return_with_dividend']:,.2f} USD")
