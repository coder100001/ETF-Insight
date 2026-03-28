"""
核心配置模块
集中管理所有配置参数
"""
import os
from dataclasses import dataclass
from typing import List, Dict


@dataclass
class ETFConfig:
    """ETF 配置"""
    symbol: str
    name: str
    market: str  # US, CN, HK
    strategy: str
    description: str
    expense_ratio: float
    focus: str
    dividend_yield: float = 0.0


# 默认 ETF 配置
DEFAULT_ETF_CONFIGS: List[ETFConfig] = [
    ETFConfig(
        symbol='SCHD',
        name='Schwab U.S. Dividend Equity ETF',
        market='US',
        strategy='质量股息策略',
        description='追踪道琼斯美国股息100指数，投资高股息、财务稳健的美国公司',
        expense_ratio=0.06,
        focus='质量+股息',
        dividend_yield=3.51
    ),
    ETFConfig(
        symbol='SPYD',
        name='SPDR Portfolio S&P 500 High Dividend ETF',
        market='US',
        strategy='高股息收益策略',
        description='追踪S&P 500中股息收益率最高的80只股票',
        expense_ratio=0.07,
        focus='高股息',
        dividend_yield=4.32
    ),
    ETFConfig(
        symbol='JEPQ',
        name='JPMorgan Nasdaq Equity Premium Income ETF',
        market='US',
        strategy='期权增强收益策略',
        description='通过纳斯达克股票+卖出看涨期权获取增强收益',
        expense_ratio=0.35,
        focus='增强收益',
        dividend_yield=9.25
    ),
    ETFConfig(
        symbol='JEPI',
        name='JPMorgan Equity Premium Income ETF',
        market='US',
        strategy='股息增强策略',
        description='摩根大通股票溢价收益ETF，通过股票期权策略提供月度股息收益',
        expense_ratio=0.35,
        focus='月度股息+收益增强',
        dividend_yield=7.50
    ),
    ETFConfig(
        symbol='VYM',
        name='Vanguard High Dividend Yield ETF',
        market='US',
        strategy='高股息宽基策略',
        description='追踪FTSE高股息率指数，投资高股息的美国大盘股',
        expense_ratio=0.06,
        focus='高股息+宽基',
        dividend_yield=2.80
    ),
]

# 数据获取配置
DATA_FETCH_CONFIG = {
    'retry_times': 3,
    'retry_delay': 2,
    'request_timeout': 30,
    'rate_limit_delay': 5,
    'max_workers': 3,  # 并发线程数
}

# 缓存配置
CACHE_CONFIG = {
    'redis_ttl': {
        'realtime': 300,      # 实时数据 5 分钟
        'historical': 86400,  # 历史数据 1 天
        'metrics': 3600,      # 指标数据 1 小时
        'comparison': 1800,   # 对比数据 30 分钟
    },
    'memory_ttl': 300,  # 内存缓存 5 分钟
}

# 定时任务配置
SCHEDULER_CONFIG = {
    'daily_update_time': '09:30',  # 每日更新时间（美股开盘前）
    'market_close_update': '16:30',  # 美股收盘后更新
    'timezone': 'America/New_York',
}

# 数据库配置
DB_CONFIG = {
    'batch_size': 1000,  # 批量插入大小
}

# 市场货币映射
MARKET_CURRENCY = {
    'US': 'USD',
    'CN': 'CNY',
    'HK': 'HKD',
}

# 周期映射（天数）
PERIOD_DAYS_MAP = {
    '1d': 1,
    '5d': 5,
    '1mo': 30,
    '3mo': 90,
    '6mo': 180,
    '1y': 365,
    '2y': 730,
    '5y': 1825,
    '10y': 3650,
    'ytd': None,  # 特殊处理
    'max': 36500,
}
