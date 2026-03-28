"""
ETF 数据获取模块
优化后的数据获取逻辑，支持增量更新、智能重试、并发处理
"""

import yfinance as yf
import pandas as pd
from datetime import datetime, timedelta, date
from typing import Optional, Dict, List, Tuple, Callable
import time
import logging
from concurrent.futures import ThreadPoolExecutor, as_completed
from functools import wraps
import random

from .config import DATA_FETCH_CONFIG, DEFAULT_ETF_CONFIGS, PERIOD_DAYS_MAP

logger = logging.getLogger(__name__)


def retry_on_error(max_retries: int = None, delay: float = None, backoff: float = 2.0):
    """
    重试装饰器
    
    Args:
        max_retries: 最大重试次数
        delay: 初始延迟时间（秒）
        backoff: 退避系数
    """
    if max_retries is None:
        max_retries = DATA_FETCH_CONFIG['retry_times']
    if delay is None:
        delay = DATA_FETCH_CONFIG['retry_delay']
    
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            current_delay = delay
            last_exception = None
            
            for attempt in range(max_retries):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    last_exception = e
                    error_msg = str(e).lower()
                    
                    # 判断是否可重试的错误
                    is_retryable = any([
                        'rate limit' in error_msg,
                        'too many requests' in error_msg,
                        'timeout' in error_msg,
                        'connection' in error_msg,
                        'temporary' in error_msg,
                    ])
                    
                    if not is_retryable or attempt == max_retries - 1:
                        raise last_exception
                    
                    # 添加随机抖动避免雪崩
                    jitter = random.uniform(0, 1)
                    wait_time = current_delay * (backoff ** attempt) + jitter
                    
                    logger.warning(
                        f"{func.__name__} 第 {attempt + 1} 次重试，"
                        f"等待 {wait_time:.1f} 秒... 错误: {e}"
                    )
                    time.sleep(wait_time)
            
            raise last_exception
        return wrapper
    return decorator


class ETFDataFetcher:
    """
    优化的 ETF 数据获取器
    
    特性：
    1. 智能重试机制（指数退避 + 抖动）
    2. 增量更新支持
    3. 并发数据获取
    4. 自动限流保护
    """
    
    def __init__(self):
        self.data_dir = 'etf_data'
        self._last_request_time = 0
        self._min_request_interval = 1.0  # 最小请求间隔（秒）
        
        # 创建数据目录
        import os
        os.makedirs(self.data_dir, exist_ok=True)
    
    def _rate_limit(self):
        """简单的速率限制"""
        current_time = time.time()
        elapsed = current_time - self._last_request_time
        if elapsed < self._min_request_interval:
            time.sleep(self._min_request_interval - elapsed)
        self._last_request_time = time.time()
    
    @retry_on_error()
    def fetch_historical_data(
        self,
        symbol: str,
        period: str = '1y',
        interval: str = '1d',
        start_date: Optional[date] = None,
        end_date: Optional[date] = None,
    ) -> Optional[pd.DataFrame]:
        """
        获取历史数据
        
        Args:
            symbol: ETF 代码
            period: 时间周期 (1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, 10y, ytd, max)
            interval: 数据间隔 (1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo)
            start_date: 开始日期（优先于 period）
            end_date: 结束日期（优先于 period）
        
        Returns:
            DataFrame 或 None
        """
        self._rate_limit()
        
        logger.info(f"获取 {symbol} 历史数据 (period={period}, interval={interval})")
        
        try:
            ticker = yf.Ticker(symbol)
            
            # 优先使用日期范围
            if start_date and end_date:
                data = ticker.history(
                    start=start_date,
                    end=end_date,
                    interval=interval,
                    timeout=DATA_FETCH_CONFIG['request_timeout']
                )
            else:
                data = ticker.history(
                    period=period,
                    interval=interval,
                    timeout=DATA_FETCH_CONFIG['request_timeout']
                )
            
            if data.empty:
                logger.warning(f"{symbol} 没有获取到数据")
                return None
            
            # 标准化列名
            data = self._standardize_columns(data)
            
            logger.info(f"成功获取 {symbol} 的 {len(data)} 条记录")
            return data
            
        except Exception as e:
            logger.error(f"获取 {symbol} 历史数据失败: {e}")
            raise
    
    def _standardize_columns(self, df: pd.DataFrame) -> pd.DataFrame:
        """标准化 DataFrame 列名"""
        column_mapping = {
            'Open': 'open',
            'High': 'high',
            'Low': 'low',
            'Close': 'close',
            'Adj Close': 'adj_close',
            'Volume': 'volume',
        }
        df = df.rename(columns=column_mapping)
        
        # 确保索引是日期类型
        if not isinstance(df.index, pd.DatetimeIndex):
            df.index = pd.to_datetime(df.index)
        
        return df
    
    @retry_on_error()
    def fetch_realtime_data(self, symbol: str) -> Optional[Dict]:
        """
        获取实时数据
        
        Args:
            symbol: ETF 代码
        
        Returns:
            包含实时数据的字典
        """
        self._rate_limit()
        
        logger.debug(f"获取 {symbol} 实时数据")
        
        try:
            ticker = yf.Ticker(symbol)
            info = ticker.info
            
            # 获取默认股息率
            default_yields = {cfg.symbol: cfg.dividend_yield for cfg in DEFAULT_ETF_CONFIGS}
            
            data = {
                'symbol': symbol,
                'name': info.get('longName', symbol),
                'current_price': info.get('currentPrice') or info.get('regularMarketPrice', 0),
                'previous_close': info.get('previousClose', 0),
                'open_price': info.get('open', 0),
                'day_high': info.get('dayHigh', 0),
                'day_low': info.get('dayLow', 0),
                'volume': info.get('volume', 0),
                'market_cap': info.get('marketCap', 0),
                'dividend_yield': info.get('dividendYield', default_yields.get(symbol, 0)),
                'fifty_two_week_high': info.get('fiftyTwoWeekHigh', 0),
                'fifty_two_week_low': info.get('fiftyTwoWeekLow', 0),
                'avg_volume': info.get('averageVolume', 0),
                'beta': info.get('beta', 0),
                'pe_ratio': info.get('trailingPE', 0),
                'timestamp': datetime.now().isoformat(),
            }
            
            # 计算涨跌幅
            if data['previous_close'] and data['current_price']:
                data['change'] = data['current_price'] - data['previous_close']
                data['change_percent'] = (data['change'] / data['previous_close']) * 100
            else:
                data['change'] = 0
                data['change_percent'] = 0
            
            return data
            
        except Exception as e:
            logger.error(f"获取 {symbol} 实时数据失败: {e}")
            raise
    
    def fetch_incremental_data(
        self,
        symbol: str,
        last_date: Optional[date] = None,
        interval: str = '1d'
    ) -> Optional[pd.DataFrame]:
        """
        获取增量数据
        
        Args:
            symbol: ETF 代码
            last_date: 数据库中最新日期
            interval: 数据间隔
        
        Returns:
            新增的数据
        """
        if last_date is None:
            # 如果没有历史数据，获取最近一年
            return self.fetch_historical_data(symbol, period='1y', interval=interval)
        
        # 计算需要获取的起始日期
        today = date.today()
        if last_date >= today:
            logger.info(f"{symbol} 数据已是最新")
            return pd.DataFrame()
        
        # 多获取一天以覆盖可能的边界问题
        start_date = last_date - timedelta(days=1)
        
        logger.info(f"获取 {symbol} 增量数据: {start_date} 到 {today}")
        
        data = self.fetch_historical_data(
            symbol=symbol,
            start_date=start_date,
            end_date=today,
            interval=interval
        )
        
        if data is not None and not data.empty:
            # 过滤掉已存在的数据
            data = data[data.index.date > last_date]
            logger.info(f"{symbol} 新增 {len(data)} 条记录")
        
        return data
    
    def fetch_multiple_symbols(
        self,
        symbols: List[str],
        period: str = '1y',
        max_workers: int = None,
        progress_callback: Optional[Callable] = None
    ) -> Dict[str, Optional[pd.DataFrame]]:
        """
        并发获取多个 ETF 数据
        
        Args:
            symbols: ETF 代码列表
            period: 时间周期
            max_workers: 最大并发数
            progress_callback: 进度回调函数 (completed, total)
        
        Returns:
            字典 {symbol: DataFrame}
        """
        if max_workers is None:
            max_workers = DATA_FETCH_CONFIG['max_workers']
        
        results = {}
        completed = 0
        
        logger.info(f"并发获取 {len(symbols)} 个 ETF 数据 (max_workers={max_workers})")
        
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            # 提交所有任务
            future_to_symbol = {
                executor.submit(self.fetch_historical_data, symbol, period): symbol
                for symbol in symbols
            }
            
            # 收集结果
            for future in as_completed(future_to_symbol):
                symbol = future_to_symbol[future]
                try:
                    results[symbol] = future.result()
                except Exception as e:
                    logger.error(f"获取 {symbol} 失败: {e}")
                    results[symbol] = None
                
                completed += 1
                if progress_callback:
                    progress_callback(completed, len(symbols))
        
        return results
    
    def fetch_multiple_realtime(
        self,
        symbols: List[str],
        max_workers: int = None
    ) -> Dict[str, Optional[Dict]]:
        """
        并发获取多个 ETF 实时数据
        
        Args:
            symbols: ETF 代码列表
            max_workers: 最大并发数
        
        Returns:
            字典 {symbol: data}
        """
        if max_workers is None:
            max_workers = DATA_FETCH_CONFIG['max_workers']
        
        results = {}
        
        logger.info(f"并发获取 {len(symbols)} 个 ETF 实时数据")
        
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            future_to_symbol = {
                executor.submit(self.fetch_realtime_data, symbol): symbol
                for symbol in symbols
            }
            
            for future in as_completed(future_to_symbol):
                symbol = future_to_symbol[future]
                try:
                    results[symbol] = future.result()
                except Exception as e:
                    logger.error(f"获取 {symbol} 实时数据失败: {e}")
                    results[symbol] = None
        
        return results
    
    def get_latest_trading_day(self, symbol: str) -> Optional[date]:
        """获取最新交易日"""
        try:
            data = self.fetch_historical_data(symbol, period='5d', interval='1d')
            if data is not None and not data.empty:
                return data.index[-1].date()
        except Exception as e:
            logger.error(f"获取 {symbol} 最新交易日失败: {e}")
        return None


# 全局单例
_etf_fetcher = None


def get_fetcher() -> ETFDataFetcher:
    """获取 ETFDataFetcher 单例"""
    global _etf_fetcher
    if _etf_fetcher is None:
        _etf_fetcher = ETFDataFetcher()
    return _etf_fetcher
