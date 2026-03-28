"""
ETF数据拉取脚本
支持拉取SCHD和SPYD的历史数据和实时数据
"""

import yfinance as yf
import pandas as pd
from datetime import datetime, timedelta
import os
import time


class ETFDataFetcher:
    """ETF数据拉取器"""
    
    def __init__(self):
        self.symbols = ['SCHD', 'SPYD', 'VYMI', 'SCHY']
        self.data_dir = 'etf_data'
        
        # 创建数据存储目录
        if not os.path.exists(self.data_dir):
            os.makedirs(self.data_dir)
    
    def fetch_historical_data(self, symbol, period='1y', interval='1d', retry=3, delay=2):
        """
        拉取历史数据
        
        参数:
            symbol: 股票代码 (SCHD 或 SPYD)
            period: 时间周期，可选: 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, 10y, ytd, max
            interval: 数据间隔，可选: 1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo
            retry: 重试次数
            delay: 重试延迟（秒）
        
        返回:
            DataFrame: 包含历史数据的表格
        """
        print(f"正在拉取 {symbol} 的历史数据 (周期: {period}, 间隔: {interval})...")
        
        for attempt in range(retry):
            try:
                ticker = yf.Ticker(symbol)
                data = ticker.history(period=period, interval=interval)
                
                if data.empty:
                    print(f"警告: {symbol} 没有获取到数据")
                    return None
                
                print(f"成功获取 {symbol} 的 {len(data)} 条数据记录")
                return data
                
            except Exception as e:
                if "Rate limited" in str(e) and attempt < retry - 1:
                    wait_time = delay * (attempt + 1)
                    print(f"遇到频率限制，{wait_time}秒后重试...（第{attempt + 1}次重试）")
                    time.sleep(wait_time)
                else:
                    print(f"拉取 {symbol} 数据时出错: {str(e)}")
                    return None
        
        return None
    
    def fetch_date_range_data(self, symbol, start_date, end_date, interval='1d'):
        """
        拉取指定日期范围的数据
        
        参数:
            symbol: 股票代码
            start_date: 开始日期 (格式: 'YYYY-MM-DD')
            end_date: 结束日期 (格式: 'YYYY-MM-DD')
            interval: 数据间隔
        
        返回:
            DataFrame: 包含历史数据的表格
        """
        print(f"正在拉取 {symbol} 从 {start_date} 到 {end_date} 的数据...")
        
        try:
            ticker = yf.Ticker(symbol)
            data = ticker.history(start=start_date, end=end_date, interval=interval)
            
            if data.empty:
                print(f"警告: {symbol} 在指定日期范围内没有数据")
                return None
            
            print(f"成功获取 {symbol} 的 {len(data)} 条数据记录")
            return data
            
        except Exception as e:
            print(f"拉取 {symbol} 数据时出错: {str(e)}")
            return None
    
    def get_current_info(self, symbol, retry=3, delay=2):
        """
        获取当前实时信息
        
        参数:
            symbol: 股票代码
            retry: 重试次数
            delay: 重试延迟（秒）
        
        返回:
            dict: 包含实时信息的字典
        """
        print(f"正在获取 {symbol} 的实时信息...")
        
        for attempt in range(retry):
            try:
                time.sleep(1)  # 添加短暂延迟避免频率限制
                ticker = yf.Ticker(symbol)
                info = ticker.info
                
                # 提取关键信息
                key_info = {
                    '股票代码': symbol,
                    '名称': info.get('longName', 'N/A'),
                    '当前价格': info.get('currentPrice', info.get('regularMarketPrice', 'N/A')),
                    '前收盘价': info.get('previousClose', 'N/A'),
                    '开盘价': info.get('open', 'N/A'),
                    '日最高': info.get('dayHigh', 'N/A'),
                    '日最低': info.get('dayLow', 'N/A'),
                    '成交量': info.get('volume', 'N/A'),
                    '市值': info.get('marketCap', 'N/A'),
                    '股息率': info.get('dividendYield', 'N/A'),
                    '52周最高': info.get('fiftyTwoWeekHigh', 'N/A'),
                    '52周最低': info.get('fiftyTwoWeekLow', 'N/A'),
                }
                
                return key_info
                
            except Exception as e:
                if "Rate limited" in str(e) and attempt < retry - 1:
                    wait_time = delay * (attempt + 1)
                    print(f"遇到频率限制，{wait_time}秒后重试...（第{attempt + 1}次重试）")
                    time.sleep(wait_time)
                else:
                    print(f"获取 {symbol} 实时信息时出错: {str(e)}")
                    return None
        
        return None
    
    def save_to_csv(self, data, symbol, filename=None):
        """
        保存数据到CSV文件
        
        参数:
            data: DataFrame数据
            symbol: 股票代码
            filename: 文件名（可选，默认使用时间戳）
        """
        if data is None or data.empty:
            print("数据为空，无法保存")
            return
        
        if filename is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f"{symbol}_{timestamp}.csv"
        
        filepath = os.path.join(self.data_dir, filename)
        data.to_csv(filepath)
        print(f"数据已保存到: {filepath}")
    
    def save_to_excel(self, data_dict, filename=None):
        """
        保存多个数据到Excel文件（多个sheet）
        
        参数:
            data_dict: 字典，key为sheet名称，value为DataFrame
            filename: 文件名（可选）
        """
        if filename is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f"ETF_Data_{timestamp}.xlsx"
        
        filepath = os.path.join(self.data_dir, filename)
        
        with pd.ExcelWriter(filepath, engine='openpyxl') as writer:
            for sheet_name, data in data_dict.items():
                if data is not None and not data.empty:
                    # 创建数据副本并移除时区信息
                    data_copy = data.copy()
                    # 移除索引的时区信息（如果是DatetimeIndex）
                    if isinstance(data_copy.index, pd.DatetimeIndex):
                        data_copy.index = data_copy.index.tz_localize(None)
                    # 移除所有datetime列的时区信息
                    for col in data_copy.columns:
                        if pd.api.types.is_datetime64tz_dtype(data_copy[col]):
                            data_copy[col] = data_copy[col].dt.tz_localize(None)
                    data_copy.to_excel(writer, sheet_name=sheet_name)
        
        print(f"数据已保存到: {filepath}")
    
    def fetch_all_data(self, period='1y', save=True):
        """
        拉取所有ETF的数据
        
        参数:
            period: 时间周期
            save: 是否保存到文件
        
        返回:
            dict: 包含所有ETF数据的字典
        """
        all_data = {}
        
        for i, symbol in enumerate(self.symbols):
            if i > 0:
                time.sleep(2)  # 在请求之间添加延迟
            
            data = self.fetch_historical_data(symbol, period=period)
            if data is not None:
                all_data[symbol] = data
                
                if save:
                    self.save_to_csv(data, symbol)
        
        return all_data
    
    def display_summary(self, data):
        """
        显示数据摘要信息
        
        参数:
            data: DataFrame数据
        """
        if data is None or data.empty:
            print("数据为空")
            return
        
        print("\n数据摘要:")
        print(f"数据记录数: {len(data)}")
        print(f"日期范围: {data.index[0]} 到 {data.index[-1]}")
        print(f"\n最近5天数据:")
        print(data.tail())
        print(f"\n统计信息:")
        print(data.describe())


def main():
    """主函数 - 支持命令行参数指定ETF"""
    import sys
    
    fetcher = ETFDataFetcher()
    
    # 解析命令行参数
    symbols_to_fetch = []
    if len(sys.argv) > 1:
        # 命令行指定了特定的ETF符号
        symbols_to_fetch = [s.upper() for s in sys.argv[1:]]
        print("="* 60)
        print(f"ETF数据拉取工具 - 指定标的: {', '.join(symbols_to_fetch)}")
        print("="* 60)
    else:
        # 默认拉取所有标的
        symbols_to_fetch = fetcher.symbols
        print("="* 60)
        print(f"ETF数据拉取工具 - 全部标的: {', '.join(symbols_to_fetch)}")
        print("="* 60)
    
    print("\n提示: 如果遇到频率限制，请等待几分钟后再试")
    print("建议: 使用VPN或等待一段时间后再尝试")
    print(f"\n用法: python fetch_etf_data.py [ETF1] [ETF2] ...")
    print(f"示例: python fetch_etf_data.py SCHD SPYD\n")
    
    # 1. 拉取过去1年的历史数据（增加延迟）
    print("[1] 拉取过去1年的历史数据")
    print("正在拉取数据，请耐心等待...\n")
    
    historical_data = {}
    for i, symbol in enumerate(symbols_to_fetch):
        if i > 0:
            time.sleep(5)  # 两个请求之间等待
        data = fetcher.fetch_historical_data(symbol, period='1y', retry=5, delay=5)
        if data is not None:
            historical_data[symbol] = data
    
    # 2. 显示数据摘要
    for symbol in symbols_to_fetch:
        if symbol in historical_data:
            print(f"\n--- {symbol} 数据摘要 ---")
            fetcher.display_summary(historical_data[symbol])
    
    # 3. 获取实时信息
    print("\n[2] 获取实时信息")
    realtime_info = {}
    for i, symbol in enumerate(symbols_to_fetch):
        if i > 0:
            time.sleep(5)
        info = fetcher.get_current_info(symbol, retry=5, delay=5)
        if info:
            realtime_info[symbol] = info
            print(f"\n--- {symbol} 实时信息 ---")
            for key, value in info.items():
                print(f"{key}: {value}")
    
    # 4. 保存数据
    print("\n[3] 保存数据到文件")
    
    # 保存到CSV
    for symbol in symbols_to_fetch:
        if symbol in historical_data:
            fetcher.save_to_csv(historical_data[symbol], symbol)
    
    # 保存到Excel（一个文件包含多个sheet）
    if historical_data:
        fetcher.save_to_excel(historical_data)
    
    print("\n" + "=" * 60)
    print("数据拉取完成！")
    print("=" * 60)


if __name__ == "__main__":
    main()
