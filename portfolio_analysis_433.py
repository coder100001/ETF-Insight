"""
投资组合收益分析 - SCHD:SPYD:JEPQ = 4:3:3
投资金额：10,000美元
"""

import yfinance as yf
import pandas as pd
from datetime import datetime, timedelta
import time
import numpy as np


class PortfolioAnalyzer433:
    """4:3:3组合分析器"""
    
    def __init__(self, total_investment=10000):
        self.total_investment = total_investment
        self.symbols = ['SCHD', 'SPYD', 'JEPQ']
        self.allocation = {
            'SCHD': 0.40,  # 40%
            'SPYD': 0.30,  # 30%
            'JEPQ': 0.30   # 30%
        }
        self.etf_data = {}
        
    def fetch_data(self, symbol, retry=3, delay=3):
        """获取ETF数据"""
        print(f"正在获取 {symbol} 数据...")
        
        for attempt in range(retry):
            try:
                time.sleep(2)
                ticker = yf.Ticker(symbol)
                
                # 获取基本信息
                info = ticker.info
                
                # 获取历史数据（1年）
                hist_1y = ticker.history(period="1y")
                
                # 获取历史数据（3个月）用于短期分析
                hist_3m = ticker.history(period="3mo")
                
                # 获取股息历史
                dividends = ticker.dividends
                
                return {
                    'info': info,
                    'history_1y': hist_1y,
                    'history_3m': hist_3m,
                    'dividends': dividends
                }
                
            except Exception as e:
                if "Rate limited" in str(e) and attempt < retry - 1:
                    wait_time = delay * (attempt + 1)
                    print(f"  遇到频率限制，{wait_time}秒后重试...")
                    time.sleep(wait_time)
                else:
                    print(f"  获取失败: {str(e)}")
                    return None
        
        return None
    
    def calculate_portfolio_metrics(self):
        """计算组合指标"""
        results = {
            '投资配置': {},
            '当前价值': {},
            '收益分析': {},
            '股息分析': {},
            '风险指标': {}
        }
        
        total_current_value = 0
        total_annual_dividend = 0
        portfolio_returns = []
        
        for symbol in self.symbols:
            if symbol not in self.etf_data or not self.etf_data[symbol]:
                continue
                
            data = self.etf_data[symbol]
            info = data['info']
            hist = data['history_1y']
            
            if hist.empty:
                continue
            
            # 投资金额
            investment = self.total_investment * self.allocation[symbol]
            results['投资配置'][symbol] = {
                '投资金额': f"${investment:,.2f}",
                '配置比例': f"{self.allocation[symbol]*100:.0f}%"
            }
            
            # 当前价格和股数
            current_price = info.get('currentPrice', info.get('regularMarketPrice', 0))
            if current_price:
                shares = investment / current_price
                current_value = shares * current_price
                total_current_value += current_value
                
                results['当前价值'][symbol] = {
                    '当前价格': f"${current_price:.2f}",
                    '持有股数': f"{shares:.2f}",
                    '当前价值': f"${current_value:,.2f}"
                }
            
            # 收益率计算
            if len(hist) > 20:
                start_price = hist['Close'].iloc[0]
                end_price = hist['Close'].iloc[-1]
                total_return = ((end_price - start_price) / start_price) * 100
                
                # 计算持有期收益
                holding_return = (current_value - investment) if current_price else 0
                
                results['收益分析'][symbol] = {
                    '年化收益率': f"{total_return:.2f}%",
                    '持有期收益': f"${holding_return:,.2f}",
                    '收益率': f"{(holding_return/investment)*100:.2f}%" if investment > 0 else "0%"
                }
                
                # 保存日收益率用于组合计算
                daily_returns = hist['Close'].pct_change().dropna()
                weighted_returns = daily_returns * self.allocation[symbol]
                portfolio_returns.append(weighted_returns)
            
            # 股息分析
            dividend_yield = info.get('yield', info.get('dividendYield', 0))
            if dividend_yield:
                annual_dividend = investment * dividend_yield
                total_annual_dividend += annual_dividend
                
                results['股息分析'][symbol] = {
                    '股息率': f"{dividend_yield*100:.2f}%",
                    '年股息收入': f"${annual_dividend:,.2f}",
                    '月股息收入': f"${annual_dividend/12:,.2f}"
                }
        
        # 组合整体指标
        if portfolio_returns:
            # 合并组合收益
            portfolio_daily_returns = pd.concat(portfolio_returns, axis=1).sum(axis=1)
            
            # 计算组合波动率
            portfolio_volatility = portfolio_daily_returns.std() * np.sqrt(252) * 100
            
            # 计算最大回撤
            cumulative = (1 + portfolio_daily_returns).cumprod()
            running_max = cumulative.expanding().max()
            drawdown = (cumulative - running_max) / running_max
            max_drawdown = drawdown.min() * 100
            
            # 计算夏普比率
            annual_return = ((total_current_value - self.total_investment) / self.total_investment) * 100
            risk_free_rate = 4.0  # 假设无风险利率4%
            sharpe_ratio = (annual_return - risk_free_rate) / portfolio_volatility if portfolio_volatility > 0 else 0
            
            results['风险指标'] = {
                '组合波动率': f"{portfolio_volatility:.2f}%",
                '最大回撤': f"{max_drawdown:.2f}%",
                '夏普比率': f"{sharpe_ratio:.2f}"
            }
        
        # 组合总结
        results['组合总结'] = {
            '总投资金额': f"${self.total_investment:,.2f}",
            '当前总价值': f"${total_current_value:,.2f}",
            '总收益': f"${total_current_value - self.total_investment:,.2f}",
            '总收益率': f"{((total_current_value - self.total_investment)/self.total_investment)*100:.2f}%",
            '年股息总收入': f"${total_annual_dividend:,.2f}",
            '组合股息率': f"{(total_annual_dividend/self.total_investment)*100:.2f}%"
        }
        
        return results
    
    def generate_report(self, metrics):
        """生成研究报表"""
        report = []
        report.append("=" * 80)
        report.append("投资组合收益分析报告")
        report.append("配置方案: SCHD:SPYD:JEPQ = 4:3:3")
        report.append(f"投资金额: ${self.total_investment:,.2f}")
        report.append("=" * 80)
        report.append(f"\n生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        
        # 一、投资配置
        report.append("\n【一、投资配置明细】")
        report.append("-" * 80)
        for symbol, detail in metrics['投资配置'].items():
            report.append(f"\n{symbol}:")
            for key, value in detail.items():
                report.append(f"  {key}: {value}")
        
        # 二、当前持仓价值
        report.append("\n\n【二、当前持仓价值】")
        report.append("-" * 80)
        for symbol, detail in metrics['当前价值'].items():
            report.append(f"\n{symbol}:")
            for key, value in detail.items():
                report.append(f"  {key}: {value}")
        
        # 三、收益分析
        report.append("\n\n【三、收益分析】")
        report.append("-" * 80)
        for symbol, detail in metrics['收益分析'].items():
            report.append(f"\n{symbol}:")
            for key, value in detail.items():
                report.append(f"  {key}: {value}")
        
        # 四、股息收入分析
        report.append("\n\n【四、股息收入分析】")
        report.append("-" * 80)
        for symbol, detail in metrics['股息分析'].items():
            report.append(f"\n{symbol}:")
            for key, value in detail.items():
                report.append(f"  {key}: {value}")
        
        # 五、风险指标
        if metrics['风险指标']:
            report.append("\n\n【五、组合风险指标】")
            report.append("-" * 80)
            for key, value in metrics['风险指标'].items():
                report.append(f"{key}: {value}")
        
        # 六、组合总结
        report.append("\n\n【六、组合总结】")
        report.append("-" * 80)
        for key, value in metrics['组合总结'].items():
            report.append(f"{key}: {value}")
        
        # 七、策略分析
        report.append("\n\n【七、策略分析】")
        report.append("-" * 80)
        report.append("""
4:3:3配置策略特点：

1. 配置逻辑：
   - SCHD 40%: 作为核心底仓，提供稳定的质量股息
   - SPYD 30%: 增强当前股息率，提升现金流
   - JEPQ 30%: 通过期权策略增加收益，提供月度分红

2. 优势：
   - 平衡了质量（SCHD）和高股息（SPYD/JEPQ）
   - 兼顾美国大盘（SCHD/SPYD）和纳斯达克科技（JEPQ）
   - 月度现金流（JEPQ）+ 季度现金流（SCHD/SPYD）
   - 适度分散，降低单一策略风险

3. 风险点：
   - JEPQ的期权策略在牛市中可能限制上涨空间
   - SPYD周期性行业暴露较高
   - 整体偏向股息策略，成长性相对有限

4. 适合人群：
   - 追求稳定现金流的投资者
   - 能接受中等波动的投资者
   - 看好美股长期表现
   - 希望每月都有分红收入
        """)
        
        # 八、优化建议
        report.append("\n【八、优化建议】")
        report.append("-" * 80)
        report.append("""
1. 再平衡频率：
   - 建议每季度检查一次配置偏离
   - 偏离超过5%时考虑再平衡
   - 利用新增资金进行再平衡，减少卖出

2. 风险控制：
   - 设置止损位：单只ETF跌幅超过-15%时重新评估
   - 关注JEPQ的月度分红变化
   - 监控SPYD的行业集中度变化

3. 进阶策略：
   - 可考虑加入5-10%的国际股息ETF（如VYMI）增加分散
   - 牛市环境下可降低JEPQ比例，增加SCHD
   - 熊市环境下可增加SCHD比例，降低SPYD

4. 税务优化：
   - 优先在税延账户（IRA/401k）中持有
   - JEPQ的期权收入可能有不同税务处理
   - 注意合格股息vs非合格股息的税率差异

5. 长期持有建议：
   - 开启DRIP（股息再投资计划）
   - 坚持定期定额投资
   - 至少持有3-5年以发挥复利效应
        """)
        
        report.append("\n" + "=" * 80)
        report.append("报告结束")
        report.append("=" * 80)
        report.append("\n免责声明：本报告仅供参考，不构成投资建议。投资有风险，入市需谨慎。")
        
        return "\n".join(report)
    
    def save_report(self, report_text, filename=None):
        """保存报告"""
        if filename is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f'Portfolio_433_Analysis_{timestamp}.txt'
        
        filepath = f'etf_data/{filename}'
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(report_text)
        
        print(f"\n报告已保存到: {filepath}")
        return filepath


def main():
    """主函数"""
    print("=" * 80)
    print("投资组合收益分析工具")
    print("配置方案: SCHD:SPYD:JEPQ = 4:3:3")
    print("投资金额: $10,000")
    print("=" * 80)
    print("\n提示: 数据获取需要一些时间，请耐心等待...\n")
    
    # 创建分析器
    analyzer = PortfolioAnalyzer433(total_investment=10000)
    
    # 获取数据
    print("正在获取ETF数据...\n")
    for symbol in analyzer.symbols:
        data = analyzer.fetch_data(symbol)
        if data:
            analyzer.etf_data[symbol] = data
            print(f"✓ {symbol} 数据获取成功")
        else:
            print(f"✗ {symbol} 数据获取失败")
        time.sleep(3)
    
    # 计算指标
    print("\n正在计算组合指标...")
    metrics = analyzer.calculate_portfolio_metrics()
    
    # 生成报告
    print("正在生成分析报告...\n")
    report = analyzer.generate_report(metrics)
    
    # 显示报告
    print(report)
    
    # 保存报告
    analyzer.save_report(report)
    
    print("\n分析完成！")


if __name__ == "__main__":
    main()
