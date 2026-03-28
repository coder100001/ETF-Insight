"""
ETF对比分析报告生成器
分析SCHD、SPYD、VYMI、SCHY四个ETF的差异和投资特征
"""

import yfinance as yf
import pandas as pd
from datetime import datetime, timedelta
import time


class ETFAnalyzer:
    """ETF分析器，用于对比分析多个ETF"""
    
    def __init__(self, symbols):
        self.symbols = symbols
        self.etf_data = {}
        self.analysis_results = {}
    
    def fetch_etf_info(self, symbol, retry=3, delay=3):
        """获取ETF的详细信息"""
        print(f"正在获取 {symbol} 的详细信息...")
        
        for attempt in range(retry):
            try:
                time.sleep(2)  # 添加延迟避免频率限制
                ticker = yf.Ticker(symbol)
                info = ticker.info
                
                # 获取历史数据用于计算收益率
                hist = ticker.history(period="1y")
                
                return {
                    'info': info,
                    'history': hist
                }
                
            except Exception as e:
                if "Rate limited" in str(e) and attempt < retry - 1:
                    wait_time = delay * (attempt + 1)
                    print(f"遇到频率限制，{wait_time}秒后重试...")
                    time.sleep(wait_time)
                else:
                    print(f"获取 {symbol} 信息时出错: {str(e)}")
                    return None
        
        return None
    
    def calculate_performance_metrics(self, history_data):
        """计算性能指标"""
        if history_data is None or history_data.empty:
            return {}
        
        # 计算收益率
        start_price = history_data['Close'].iloc[0]
        end_price = history_data['Close'].iloc[-1]
        total_return = ((end_price - start_price) / start_price) * 100
        
        # 计算波动率（年化）
        daily_returns = history_data['Close'].pct_change().dropna()
        volatility = daily_returns.std() * (252 ** 0.5) * 100  # 年化波动率
        
        # 计算最大回撤
        cumulative = (1 + daily_returns).cumprod()
        running_max = cumulative.expanding().max()
        drawdown = (cumulative - running_max) / running_max
        max_drawdown = drawdown.min() * 100
        
        # 计算夏普比率（假设无风险利率4%）
        risk_free_rate = 0.04
        excess_return = (total_return / 100) - risk_free_rate
        sharpe_ratio = excess_return / (volatility / 100) if volatility != 0 else 0
        
        return {
            '年化收益率': round(total_return, 2),
            '年化波动率': round(volatility, 2),
            '最大回撤': round(max_drawdown, 2),
            '夏普比率': round(sharpe_ratio, 2)
        }
    
    def analyze_all_etfs(self):
        """分析所有ETF"""
        print("\n" + "="*60)
        print("开始分析ETF...")
        print("="*60 + "\n")
        
        for symbol in self.symbols:
            data = self.fetch_etf_info(symbol)
            if data:
                self.etf_data[symbol] = data
                
                # 计算性能指标
                performance = self.calculate_performance_metrics(data['history'])
                
                info = data['info']
                self.analysis_results[symbol] = {
                    '基本信息': {
                        '名称': info.get('longName', 'N/A'),
                        '股票代码': symbol,
                        '当前价格': info.get('currentPrice', info.get('regularMarketPrice', 'N/A')),
                        '总资产规模': info.get('totalAssets', 'N/A'),
                        '费用率': info.get('annualReportExpenseRatio', 'N/A'),
                        '成立日期': info.get('fundInceptionDate', 'N/A'),
                    },
                    '股息信息': {
                        '股息率': info.get('yield', info.get('dividendYield', 'N/A')),
                        '分红频率': info.get('dividendRate', 'N/A'),
                        '最近分红': info.get('lastDividendValue', 'N/A'),
                    },
                    '性能指标': performance,
                    '持仓信息': {
                        '持仓数量': info.get('holdings', 'N/A'),
                        '前十大持仓占比': info.get('top10Holdings', 'N/A'),
                        '行业集中度': info.get('sectorWeightings', 'N/A'),
                    }
                }
                
                print(f"✓ {symbol} 分析完成")
                time.sleep(3)  # 请求间延迟
        
        print("\n所有ETF分析完成！\n")
    
    def generate_comparison_report(self):
        """生成对比报告"""
        report = []
        report.append("="*80)
        report.append("ETF对比分析报告 - SCHD vs SPYD vs VYMI vs SCHY")
        report.append("="*80)
        report.append(f"\n生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        
        # 1. ETF概述
        report.append("\n【一、ETF基本概述】")
        report.append("-"*80)
        
        etf_descriptions = {
            'SCHD': """
SCHD (Schwab U.S. Dividend Equity ETF)
- 发行商: Charles Schwab
- 策略: 被动跟踪道琼斯美国股息100指数
- 特点: 专注高质量、持续分红的美国大盘股
- 目标: 长期资本增值和稳定股息收入
- 风险等级: 中等风险
            """,
            'SPYD': """
SPYD (SPDR Portfolio S&P 500 High Dividend ETF)
- 发行商: State Street
- 策略: 被动跟踪S&P 500高股息指数
- 特点: 平均权重配置80只S&P 500中股息率最高的股票
- 目标: 追求高股息收益率
- 风险等级: 中等风险
            """,
            'VYMI': """
VYMI (Vanguard International High Dividend Yield ETF)
- 发行商: Vanguard
- 投资范围: 美国以外的发达和新兴市场
- 特点: 专注高股息国际股票
- 地域分散: 欧洲、亚太、新兴市场
- 分红频率: 季度
- 风险等级: 中等偏高（受汇率和海外市场影响）
            """,
            'SCHY': """
SCHY (Schwab International Dividend Equity ETF)
- 发行商: Charles Schwab
- 投资范围: 美国以外发达市场高质量股息股
- 特点: 强调质量筛选 + 稳定分红
- 风格: 略偏防御的国际价值股
- 风险等级: 中等
            """
        }
        
        for symbol in self.symbols:
            if symbol in etf_descriptions:
                report.append(etf_descriptions[symbol])
        
        # 2. 详细数据对比
        report.append("\n【二、详细数据对比】")
        report.append("-"*80)
        
        if self.analysis_results:
            # 基本信息对比
            report.append("\n1. 基本信息对比")
            report.append("-"*40)
            for symbol, analysis in self.analysis_results.items():
                report.append(f"\n{symbol}:")
                for key, value in analysis['基本信息'].items():
                    report.append(f"  {key}: {value}")
            
            # 股息信息对比
            report.append("\n\n2. 股息信息对比")
            report.append("-"*40)
            for symbol, analysis in self.analysis_results.items():
                report.append(f"\n{symbol}:")
                for key, value in analysis['股息信息'].items():
                    if key == '股息率' and value != 'N/A':
                        try:
                            value = f"{float(value)*100:.2f}%"
                        except:
                            pass
                    report.append(f"  {key}: {value}")
            
            # 性能指标对比
            report.append("\n\n3. 性能指标对比（过去一年）")
            report.append("-"*40)
            for symbol, analysis in self.analysis_results.items():
                report.append(f"\n{symbol}:")
                for key, value in analysis['性能指标'].items():
                    if '%' not in str(value):
                        report.append(f"  {key}: {value}%")
                    else:
                        report.append(f"  {key}: {value}")
        
        # 3. 核心差异分析
        report.append("\n\n【三、核心差异分析】")
        report.append("-"*80)
        
        report.append("""
1. 投资策略差异：
   - SCHD: 美国高质量股息股票，偏质量+股息增长
   - SPYD: 美国高股息股票，偏当前股息率
   - VYMI: 国际高股息股票，强调估值和股息
   - SCHY: 国际高质量股息股票，偏质量+稳定

2. 收益来源：
   - SCHD: 股票增值 + 稳定增长的股息
   - SPYD: 高股息收入 + 一定的资本增值
   - VYMI: 海外股息 + 估值修复机会
   - SCHY: 海外质量股 + 稳定股息

3. 持仓风格：
   - SCHD: 偏大盘价值 + 质量股
   - SPYD: 周期性行业权重较高（如房地产、公用事业等）
   - VYMI: 覆盖欧洲、亚太、新兴市场高股息股
   - SCHY: 偏发达市场高质量股息股，行业更均衡

4. 分红频率：
   - 四只ETF目前均为季度分红

5. 波动性：
   - SCHD: 中等波动，质量护城河较强
   - SPYD: 中高波动，受周期行业和经济环境影响更大
   - VYMI: 受海外市场和汇率影响，波动中等偏高
   - SCHY: 相对VYMI略稳健一些，偏质量防御
        """)
        
        # 4. 适用投资者分析
        report.append("\n\n【四、适用投资者类型】")
        report.append("-"*80)
        
        report.append("""
SCHD 适合：
- 追求长期稳定增长的投资者
- 注重股息质量和可持续性
- 偏好蓝筹股和价值投资
- 适合作为美元资产核心底仓

SPYD 适合：
- 追求较高当前股息率
- 能接受一定周期性波动
- 希望强化现金流的投资者

VYMI 适合：
- 希望增加美国以外市场暴露
- 看重高股息 + 估值优势
- 接受汇率波动和海外市场风险

SCHY 适合：
- 希望配置海外资产但更偏质量和防御
- 不想过度暴露于高风险新兴市场
- 作为SCHD的国际质量股息补充
        """)
        
        # 5. 风险提示
        report.append("\n\n【五、风险提示】")
        report.append("-"*80)
        
        report.append("""
SCHD 风险：
- 股息率相对不算最高，但质量更高
- 可能错过部分高成长科技股机会
- 风格轮动时可能阶段性落后大盘

SPYD 风险：
- 周期性行业暴露大（房地产、公用事业、能源等）
- 高股息可能来自基本面承压公司
- 经济衰退期分红可能被削减

VYMI 风险：
- 货币汇率波动影响实际回报
- 地缘政治与海外政策风险
- 新兴市场波动较大

SCHY 风险：
- 仍受海外市场系统性风险影响
- 行业与地域集中度需要关注
- 相比VYMI，可能牺牲部分高股息以换取质量
        """)
        
        # 6. 投资建议
        report.append("\n\n【六、组合配置建议】")
        report.append("-"*80)
        
        report.append("""
均衡型投资者：
- SCHD: 40% （核心美股质量股息）
- SPYD: 25% （抬高整体股息率）
- VYMI: 20% （国际高股息）
- SCHY: 15% （国际质量股息补充）

保守型投资者：
- SCHD: 55% （主要配置）
- SCHY: 25% （防御型海外股息）
- VYMI: 10%
- SPYD: 10%

激进型投资者：
- SPYD: 40% （高股息+更高波动）
- VYMI: 30% （国际高股息）
- SCHD: 20%
- SCHY: 10%

注意：
1. 以上配置仅供参考，需根据个人风险承受能力调整
2. 建议定期再平衡，保持目标配置比例
3. 关注宏观经济和利率环境变化
4. 不要过度集中于单一ETF
        """)
        
        report.append("\n" + "="*80)
        report.append("报告结束")
        report.append("="*80)
        
        return "\n".join(report)
    
    def save_report(self, report_text, filename=None):
        """保存报告到文件"""
        if filename is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f'ETF_Comparison_Report_{timestamp}.txt'
        
        filepath = f'etf_data/{filename}'
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(report_text)
        
        print(f"\n报告已保存到: {filepath}")
        return filepath


def main():
    """主函数"""
    symbols = ['SCHD', 'SPYD', 'VYMI', 'SCHY']
    
    print("="*80)
    print("ETF对比分析报告生成器")
    print("="*80)
    print("\n提示: 数据获取需要一些时间，请耐心等待...")
    print("建议: 如遇频率限制，请等待几分钟后重试\n")
    
    # 创建分析器
    analyzer = ETFAnalyzer(symbols)
    
    # 分析所有ETF
    analyzer.analyze_all_etfs()
    
    # 生成报告
    print("正在生成对比报告...")
    report = analyzer.generate_comparison_report()
    
    # 显示报告
    print("\n" + report)
    
    # 保存报告
    analyzer.save_report(report)
    
    print("\n分析完成！")


if __name__ == "__main__":
    main()
