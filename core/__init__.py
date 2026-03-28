"""
核心模块
包含 ETF 数据获取、处理和定时更新的核心逻辑

使用示例:
    # 快速更新所有 ETF 数据
    from core import quick_update
    result = quick_update()

    # 获取指定 ETF 数据
    from core import get_fetcher
    fetcher = get_fetcher()
    data = fetcher.fetch_historical_data('SCHD', period='1y')

    # 设置定时更新
    from core import get_scheduler
    scheduler = get_scheduler()
    scheduler.start()
    scheduler.schedule_market_close_update()
"""

from .data_fetcher import ETFDataFetcher, get_fetcher
from .data_storage import HybridStorage, get_storage
from .scheduler_service import (
    ETFUpdateScheduler,
    ETFUpdateTask,
    UpdateStatus,
    get_scheduler,
    quick_update,
    setup_default_schedule,
)
from .config import ETFConfig, DEFAULT_ETF_CONFIGS

__all__ = [
    # 数据获取
    'ETFDataFetcher',
    'get_fetcher',

    # 数据存储
    'HybridStorage',
    'get_storage',

    # 定时调度
    'ETFUpdateScheduler',
    'ETFUpdateTask',
    'UpdateStatus',
    'get_scheduler',
    'quick_update',
    'setup_default_schedule',

    # 配置
    'ETFConfig',
    'DEFAULT_ETF_CONFIGS',
]

__version__ = '1.0.0'
