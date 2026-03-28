"""
ETF投资组合构造器
支持SPYD、SCHD、VYMI的多种配置方案
"""

import yfinance as yf
import pandas as pd
from datetime import datetime
import time


class PortfolioConstructor:
    """投资组合构造器"""
    
    def __init__(self):
        self.symbols = ['SPYD', 'SCHD', 'VYMI']
        self.etf_info = {}
        self.portfolio_strategies = {}
    
    def fetch_basic_info(self, symbol, retry=3, delay=3):
        """获取ETF基本信息"""
        print(f"正在获取 {symbol} 信息...")
        
        for attempt in range(retry):
            try:
                time.sleep(2)
                ticker = yf.Ticker(symbol)
                info = ticker.info
                hist = ticker.history(period="1y")
                
                return {
                    '名称': info.get('longName', 'N/A'),
                    '当前价格': info.get('currentPrice', info.get('regularMarketPrice', 'N/A')),
                    '股息率': info.get('yield', info.get('dividendYield', 'N/A')),
                    '费用率': info.get('annualReportExpenseRatio', 'N/A'),
                    '总资产': info.get('totalAssets', 'N/A'),
                    '历史数据': hist
                }
            except Exception as e:
                if attempt < retry - 1:
                    time.sleep(delay * (attempt + 1))
                else:
                    print(f"  获取失败: {str(e)}")
                    return None
        return None
    
    def define_portfolio_strategies(self):
        """定义多种组合配置策略"""
        
        self.portfolio_strategies = {
            '均衡型组合': {
                'SCHD': 50,
                'SPYD': 30,
                'VYMI': 20,
                '特点': '平衡美国市场和国际市场，SCHD作为核心',
                '适合': '追求稳定收益和适度国际分散的投资者',
                '风险等级': '中等',
                '预期股息率': '3.5-4.5%'
            },
            
            '高股息型组合': {
                'SPYD': 40,
                'VYMI': 35,
                'SCHD': 25,
                '特点': '最大化当前股息收入',
                '适合': '追求高现金流的退休投资者',
                '风险等级': '中高',
                '预期股息率': '4.5-5.5%'
            },
            
            '质量优先组合': {
                'SCHD': 60,
                'SPYD': 25,
                'VYMI': 15,
                '特点': '注重股息质量和可持续增长',
                '适合': '长期持有、注重股息增长的投资者',
                '风险等级': '中低',
                '预期股息率': '3.0-4.0%'
            },
            
            '国际分散组合': {
                'VYMI': 40,
                'SCHD': 35,
                'SPYD': 25,
                '特点': '增加国际市场暴露，分散美国市场风险',
                '适合': '看好国际市场复苏的投资者',
                '风险等级': '中等',
                '预期股息率': '4.0-5.0%'
            },
            
            '等权配置组合': {
                'SCHD': 33.33,
                'SPYD': 33.33,
                'VYMI': 33.34,
                '特点': '简单的三等分配置',
                '适合': '不想深度研究、追求简单的投资者',
                '风险等级': '中等',
                '预期股息率': '4.0-4.5%'
            },
            
            '美国核心组合': {
                'SCHD': 55,
                'SPYD': 35,
                'VYMI': 10,
                '特点': '主要配置美国市场，少量国际分散',
                '适合': '看好美国市场的投资者',
                '风险等级': '中等',
                '预期股息率': '3.5-4.5%'
            },
            
            '保守型组合': {
                'SCHD': 70,
                'VYMI': 20,
                'SPYD': 10,
                '特点': '以质量股为主，降低波动',
                '适合': '风险厌恶型投资者',
                '风险等级': '低',
                '预期股息率': '3.0-3.5%'
            },
            
            '激进型组合': {
                'SPYD': 45,
                'VYMI': 35,
                'SCHD': 20,
                '特点': '追求最高股息率，接受更高波动',
                '适合': '追求最大现金流、风险承受力强',
                '风险等级': '高',
                '预期股息率': '5.0-6.0%'
            }
        }
    
    def calculate_portfolio_metrics(self, allocation, investment_amount=10000):
        """计算投资组合指标"""
        results = {
            '总投资金额': investment_amount,
            '各ETF投资金额': {},
            '各ETF购买股数': {},
            '预计年股息收入': 0,
            '组合加权股息率': 0
        }
        
        total_dividend = 0
        
        for symbol, percentage in allocation.items():
            if symbol in self.etf_info and self.etf_info[symbol]:
                info = self.etf_info[symbol]
                amount = investment_amount * (percentage / 100)
                results['各ETF投资金额'][symbol] = round(amount, 2)
                
                price = info.get('当前价格', 0)
                if price and price != 'N/A':
                    shares = amount / price
                    results['各ETF购买股数'][symbol] = int(shares)
                    
                    dividend_yield = info.get('股息率', 0)
                    if dividend_yield and dividend_yield != 'N/A':
                        annual_dividend = amount * dividend_yield
                        total_dividend += annual_dividend
        
        results['预计年股息收入'] = round(total_dividend, 2)
        if investment_amount > 0:
            results['组合加权股息率'] = round((total_dividend / investment_amount) * 100, 2)
        
        return results
    
    def generate_portfolio_report(self):
        """生成投资组合配置报告"""
        report = []
        report.append("=" * 80)
        report.append("ETF投资组合配置方案报告")
        report.append("SPYD + SCHD + VYMI 组合配置")
        report.append("=" * 80)
        report.append(f"\n生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        
        # ETF简介
        report.append("\n【一、ETF基本介绍】")
        report.append("-" * 80)
        
        etf_intro = {
            'SCHD': """
SCHD - Schwab U.S. Dividend Equity ETF
- 发行商: Charles Schwab
- 跟踪指数: 道琼斯美国股息100指数
- 特点: 高质量、持续分红的美国大盘股
- 筛选标准: 10年连续分红、财务健康、股息增长
- 分红频率: 季度
- 费用率: 0.06%（极低）
- 核心优势: 质量优先、股息增长稳定、费用极低
            """,
            'SPYD': """
SPYD - SPDR Portfolio S&P 500 High Dividend ETF
- 发行商: State Street
- 跟踪指数: S&P 500高股息指数
- 特点: 配置80只S&P 500中股息率最高的股票
- 权重方式: 等权重配置
- 分红频率: 季度
- 费用率: 0.07%（很低）
- 核心优势: 高当前股息率、等权重分散风险
            """,
            'VYMI': """
VYMI - Vanguard International High Dividend Yield ETF
- 发行商: Vanguard
- 投资范围: 美国以外的发达和新兴市场
- 特点: 专注高股息国际股票
- 地域分散: 欧洲、亚太、新兴市场
- 分红频率: 季度
- 费用率: 0.22%（较低）
- 核心优势: 国际分散、货币多元化、估值相对较低
            """
        }
        
        for symbol in self.symbols:
            report.append(etf_intro.get(symbol, ''))
        
        # 三者对比分析
        report.append("\n【二、三者核心差异对比】")
        report.append("-" * 80)
        
        report.append("""
1. 市场覆盖：
   - SCHD: 美国市场，大盘优质股
   - SPYD: 美国市场，S&P 500高股息股
   - VYMI: 国际市场（非美国）

2. 股息策略：
   - SCHD: 质量 + 增长（股息贵族）
   - SPYD: 高当前股息率（可能牺牲增长）
   - VYMI: 国际高股息（估值优势）

3. 风险特征：
   - SCHD: 低波动，质量防御
   - SPYD: 中高波动，周期性暴露
   - VYMI: 中等波动，货币风险

4. 行业偏好：
   - SCHD: 消费、医疗、工业、金融
   - SPYD: 房地产、公用事业、能源
   - VYMI: 金融、能源、消费品

5. 费用对比：
   - SCHD: 0.06%（最低）
   - SPYD: 0.07%（很低）
   - VYMI: 0.22%（相对较高但合理）

6. 股息增长潜力：
   - SCHD: 高（历史年均增长5-7%）
   - SPYD: 中等（波动较大）
   - VYMI: 中等（取决于国际市场）
        """)
        
        # 组合配置方案
        report.append("\n【三、推荐组合配置方案】")
        report.append("-" * 80)
        
        for strategy_name, strategy_detail in self.portfolio_strategies.items():
            report.append(f"\n>>> {strategy_name}")
            report.append("-" * 40)
            
            # 配置比例
            report.append("配置比例:")
            for symbol, percentage in strategy_detail.items():
                if symbol in self.symbols:
                    report.append(f"  - {symbol}: {percentage}%")
            
            # 策略特点
            report.append(f"\n特点: {strategy_detail.get('特点', '')}")
            report.append(f"适合: {strategy_detail.get('适合', '')}")
            report.append(f"风险等级: {strategy_detail.get('风险等级', '')}")
            report.append(f"预期股息率: {strategy_detail.get('预期股息率', '')}")
            report.append("")
        
        # 投资金额示例
        report.append("\n【四、投资金额配置示例】")
        report.append("-" * 80)
        report.append("\n以投资10,000美元为例，各方案具体配置：\n")
        
        if self.etf_info:
            for strategy_name, strategy_detail in self.portfolio_strategies.items():
                allocation = {k: v for k, v in strategy_detail.items() if k in self.symbols}
                if allocation:
                    report.append(f"\n{strategy_name}:")
                    report.append("-" * 40)
                    
                    for symbol, percentage in allocation.items():
                        amount = 10000 * (percentage / 100)
                        report.append(f"{symbol}: ${amount:,.2f} ({percentage}%)")
                    report.append("")
        
        # 组合构建建议
        report.append("\n【五、组合构建实施建议】")
        report.append("-" * 80)
        
        report.append("""
1. 入场策略：
   - 建议分批建仓，每月定投
   - 避免追高，等待市场调整机会
   - 首次投资可以一次性建立基础仓位（50%），其余分批

2. 再平衡：
   - 每季度检查一次配置偏离
   - 偏离超过5%时考虑再平衡
   - 利用新增资金进行再平衡，减少卖出

3. 税务优化：
   - 优先在退休账户（IRA、401k）中持有
   - 应税账户关注合格股息税率
   - VYMI的外国税收抵免

4. 股息再投资：
   - 开启DRIP（股息再投资计划）
   - 利用复利效应加速增长
   - 小额投资者尤其重要

5. 监控指标：
   - 每只ETF的股息支付稳定性
   - 费用率变化
   - 持仓行业变化
   - 整体组合波动率

6. 调整时机：
   - 个人风险承受能力变化
   - 市场环境重大变化
   - 退休接近时降低SPYD和VYMI比例
        """)
        
        # 风险提示
        report.append("\n【六、风险提示与注意事项】")
        report.append("-" * 80)
        
        report.append("""
1. 市场风险：
   - 所有ETF受市场波动影响
   - 经济衰退期股息可能削减
   - 利率上升影响高股息股估值

2. 行业集中风险：
   - SPYD在房地产、公用事业集中度高
   - 特定行业危机影响较大

3. 国际市场风险（VYMI）：
   - 货币汇率波动
   - 地缘政治风险
   - 新兴市场波动
   - 外国税收抵扣复杂性

4. 股息削减风险：
   - SPYD追求高股息率，可能选入削减股息的公司
   - 经济下行期风险增大

5. 再投资风险：
   - 股息率高不等于总回报高
   - 需关注总回报表现

6. 过度集中：
   - 不要全部资产投入股息ETF
   - 建议股息ETF占股票仓位30-60%
   - 配合成长型ETF平衡

建议：
- 投资前充分了解各ETF特性
- 根据自身情况选择合适配置
- 长期持有，不要频繁交易
- 定期审视，适时调整
        """)
        
        # 不同投资目标配置建议
        report.append("\n【七、根据投资目标选择配置】")
        report.append("-" * 80)
        
        report.append("""
1. 目标：退休收入最大化（60岁+）
   推荐方案: 高股息型组合
   配置: SPYD 40% + VYMI 35% + SCHD 25%
   
2. 目标：长期财富积累（30-50岁）
   推荐方案: 质量优先组合
   配置: SCHD 60% + SPYD 25% + VYMI 15%

3. 目标：国际分散
   推荐方案: 国际分散组合
   配置: VYMI 40% + SCHD 35% + SPYD 25%

4. 目标：平衡收益与增长
   推荐方案: 均衡型组合
   配置: SCHD 50% + SPYD 30% + VYMI 20%

5. 目标：简单易管理
   推荐方案: 等权配置组合
   配置: 各 33.33%

6. 目标：风险最小化
   推荐方案: 保守型组合
   配置: SCHD 70% + VYMI 20% + SPYD 10%

7. 目标：现金流最大化
   推荐方案: 激进型组合
   配置: SPYD 45% + VYMI 35% + SCHD 20%
        """)
        
        report.append("\n" + "=" * 80)
        report.append("报告结束 - 投资有风险，入市需谨慎")
        report.append("=" * 80)
        
        return "\n".join(report)
    
    def save_report(self, report_text, filename=None):
        """保存报告"""
        if filename is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f'Portfolio_Configuration_{timestamp}.txt'
        
        filepath = f'etf_data/{filename}'
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(report_text)
        
        print(f"\n报告已保存到: {filepath}")
        return filepath


def main():
    """主函数"""
    print("=" * 80)
    print("ETF投资组合构造器")
    print("SPYD + SCHD + VYMI 组合配置方案")
    print("=" * 80)
    print()
    
    constructor = PortfolioConstructor()
    
    # 获取ETF信息（可选，如果API可用）
    print("正在获取ETF信息...\n")
    for symbol in constructor.symbols:
        info = constructor.fetch_basic_info(symbol)
        if info:
            constructor.etf_info[symbol] = info
            print(f"✓ {symbol} 信息获取成功")
        else:
            print(f"✗ {symbol} 信息获取失败（将使用默认配置）")
        time.sleep(2)
    
    # 定义策略
    constructor.define_portfolio_strategies()
    
    # 生成报告
    print("\n正在生成配置报告...\n")
    report = constructor.generate_portfolio_report()
    
    # 显示报告
    print(report)
    
    # 保存报告
    constructor.save_report(report)
    
    print("\n配置方案生成完成！")


if __name__ == "__main__":
    main()
