"""
ETF数据分析服务 - SCHD, SPYD, JEPQ, JEPI, VYM
"""

import yfinance as yf
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
from decimal import Decimal
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
import logging

logger = logging.getLogger(__name__)

# 延迟导入避免循环依赖
def get_cache_manager():
    from .cache_manager import etf_cache
    return etf_cache


class ETFAnalysisService:
    """ETF数据分析服务"""

    # 保留作为后备（如果数据库为空）
    _FALLBACK_SYMBOLS = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']

    # 市场货币映射
    _MARKET_CURRENCY = {
        'US': 'USD',
        'CN': 'CNY',
        'HK': 'HKD'
    }
    _FALLBACK_ETF_INFO = {
        'SCHD': {
            'name': 'Schwab U.S. Dividend Equity ETF',
            'strategy': '质量股息策略',
            'description': '追踪道琼斯美国股息100指数，投资高股息、财务稳健的美国公司',
            'expense_ratio': 0.06,
            'focus': '质量+股息'
        },
        'SPYD': {
            'name': 'SPDR Portfolio S&P 500 High Dividend ETF',
            'strategy': '高股息收益策略',
            'description': '追踪S&P 500中股息收益率最高的80只股票',
            'expense_ratio': 0.07,
            'focus': '高股息'
        },
        'JEPQ': {
            'name': 'JPMorgan Nasdaq Equity Premium Income ETF',
            'strategy': '期权增强收益策略',
            'description': '通过纳斯达克股票+卖出看涨期权获取增强收益',
            'expense_ratio': 0.35,
            'focus': '增强收益'
        },
        'JEPI': {
            'name': 'JPMorgan Equity Premium Income ETF',
            'strategy': '股息增强策略',
            'description': '摩根大通股票溢价收益ETF，通过股票期权策略提供月度股息收益',
            'expense_ratio': 0.35,
            'focus': '月度股息+收益增强'
        },
        'VYM': {
            'name': 'Vanguard High Dividend Yield ETF',
            'strategy': '高股息宽基策略',
            'description': '追踪FTSE高股息率指数，投资高股息的美国大盘股',
            'expense_ratio': 0.06,
            'focus': '高股息+宽基'
        }
    }
    
    def __init__(self):
        self.redis_cache = None  # 延迟初始化
        self.memory_cache = {}  # 内存缓存（备用）
        self.cache_time = {}
        self.memory_cache_duration = 300  # 内存缓存5分钟
        self._etf_config_cache = None  # ETF配置缓存
        self._etf_config_cache_time = 0
        self._etf_config_cache_duration = 60  # ETF配置缓存1分钟
    
    @property
    def SYMBOLS(self):
        """动态获取启用的ETF代码列表"""
        configs = self._get_active_etf_configs()
        if configs:
            return [config['symbol'] for config in configs]
        return self._FALLBACK_SYMBOLS
    
    @property
    def ETF_INFO(self):
        """动态获取ETF信息字典"""
        configs = self._get_active_etf_configs()
        if configs:
            return {config['symbol']: {
                'name': config['name'],
                'strategy': config['strategy'],
                'description': config['description'],
                'expense_ratio': config['expense_ratio'],
                'focus': config['focus']
            } for config in configs}
        return self._FALLBACK_ETF_INFO
    
    def _get_active_etf_configs(self, market=None):
        """获取启用的ETF配置（带缓存）"""
        # 检查缓存
        now = time.time()
        if self._etf_config_cache and (now - self._etf_config_cache_time < self._etf_config_cache_duration):
            configs = self._etf_config_cache
        else:
            # 从数据库读取
            try:
                from .models import ETFConfig
                queryset = ETFConfig.objects.filter(status=1).order_by('sort_order', 'symbol')
                if market:
                    queryset = queryset.filter(market=market)
                
                configs = list(queryset.values(
                    'symbol', 'name', 'market', 'strategy', 
                    'description', 'focus', 'expense_ratio'
                ))
                
                # 转换expense_ratio为float
                for config in configs:
                    if config['expense_ratio']:
                        config['expense_ratio'] = float(config['expense_ratio'])
                    else:
                        config['expense_ratio'] = 0
                
                # 更新缓存
                self._etf_config_cache = configs
                self._etf_config_cache_time = now
                
                logger.info(f'从数据库加载 {len(configs)} 个启用的ETF配置')
            except Exception as e:
                logger.warning(f'读取ETF配置失败，使用后备数据: {e}')
                configs = []
        
        if market:
            return [c for c in configs if c.get('market') == market]
        return configs
    
    def clear_etf_config_cache(self):
        """清除ETF配置缓存"""
        self._etf_config_cache = None
        self._etf_config_cache_time = 0
        logger.info('ETF配置缓存已清除')
    
    def get_active_etfs(self, market=None):
        """获取启用的ETF列表（供外部调用）"""
        return self._get_active_etf_configs(market)

    def get_exchange_rate(self, from_currency, to_currency, rate_date=None):
        """
        获取最新汇率

        参数:
            from_currency: 源货币（USD, CNY, HKD）
            to_currency: 目标货币（USD, CNY, HKD）
            rate_date: 汇率日期（如果为None则使用最新汇率）

        返回:
            汇率数值（1单位from_currency = ?单位to_currency）
        """
        try:
            from .models import ExchangeRate

            if from_currency == to_currency:
                return 1.0

            # 查询最新汇率
            queryset = ExchangeRate.objects.filter(
                from_currency=from_currency,
                to_currency=to_currency
            )

            if rate_date:
                queryset = queryset.filter(rate_date=rate_date)

            rate_record = queryset.order_by('-rate_date').first()

            if rate_record:
                return float(rate_record.rate)
            else:
                logger.warning(f'未找到 {from_currency} -> {to_currency} 的汇率，使用默认值1.0')
                return 1.0
        except Exception as e:
            logger.error(f'获取汇率失败: {e}')
            return 1.0

    def convert_to_usd(self, amount, currency):
        """
        将任意货币转换为美元

        参数:
            amount: 金额
            currency: 货币（USD, CNY, HKD）

        返回:
            转换后的美元金额
        """
        if currency == 'USD':
            return amount

        rate = self.get_exchange_rate(currency, 'USD')
        return amount * rate

    def get_etf_currency(self, symbol):
        """
        获取ETF的计价货币

        参数:
            symbol: ETF代码

        返回:
            货币代码（USD, CNY, HKD）
        """
        try:
            from .models import ETFConfig
            config = ETFConfig.objects.filter(symbol=symbol).first()
            if config:
                return self._MARKET_CURRENCY.get(config.market, 'USD')
            else:
                # 默认为美元
                return 'USD'
        except Exception as e:
            logger.warning(f'获取ETF货币失败: {e}')
            return 'USD'

    def _get_redis_cache(self):
        """获取Redis缓存管理器（延迟初始化）"""
        if self.redis_cache is None:
            try:
                self.redis_cache = get_cache_manager()
            except Exception as e:
                logger.warning(f'Redis缓存初始化失败，使用内存缓存: {e}')
        return self.redis_cache
    
    def _get_cached(self, key):
        """获取缓存数据（优先Redis，降级内存）"""
        # 先尝试Redis
        redis_cache = self._get_redis_cache()
        if redis_cache:
            try:
                cached = redis_cache.cache.get(key)
                if cached is not None:
                    logger.debug(f'Redis缓存命中: {key}')
                    return cached
            except Exception as e:
                logger.warning(f'Redis读取失败: {e}')
        
        # 降级到内存缓存
        if key in self.memory_cache:
            if time.time() - self.cache_time.get(key, 0) < self.memory_cache_duration:
                logger.debug(f'内存缓存命中: {key}')
                return self.memory_cache[key]
        
        return None
    
    def _set_cache(self, key, value, timeout=None):
        """设置缓存（双写Redis+内存）"""
        # 写入Redis
        redis_cache = self._get_redis_cache()
        if redis_cache:
            try:
                redis_cache.cache.set(key, value, timeout=timeout)
                logger.debug(f'Redis缓存写入: {key}')
            except Exception as e:
                logger.warning(f'Redis写入失败: {e}')
        
        # 同时写入内存缓存（备用）
        self.memory_cache[key] = value
        self.cache_time[key] = time.time()

    def fetch_realtime_data(self, symbol):
        """获取实时数据（直接从数据库读取，不使用缓存）"""
        # 直接从数据库获取最新记录
        try:
            from .models import ETFData
            from datetime import date

            # 首先尝试获取今天的最新数据
            today = date.today()
            latest = ETFData.objects.filter(
                symbol=symbol,
                date=today
            ).first()

            # 如果今天没有数据，获取最新的历史数据
            if not latest:
                latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
                if latest:
                    logger.info(f'使用历史数据: {symbol} ({latest.date})')
            else:
                logger.info(f'使用数据库数据: {symbol} ({today})')

            # 预设股息率（基于实际市场股息率）
            default_yields = {
                'SCHD': 3.51,    # Schwab 实际股息率
                'SPYD': 4.32,    # SPDR 实际股息率
                'JEPQ': 9.25,    # JPMorgan Nasdaq 实际股息率
                'JEPI': 7.50,    # JPMorgan Equity 实际股息率
                'VYM': 2.80      # Vanguard 实际股息率
            }

            data = {
                'symbol': symbol,
                'name': self.ETF_INFO.get(symbol, {}).get('name', symbol),
                'current_price': float(latest.close_price or 0),
                'previous_close': float(latest.open_price or 0),
                'open_price': float(latest.open_price or 0),
                'day_high': float(latest.high_price or 0),
                'day_low': float(latest.low_price or 0),
                'volume': int(latest.volume or 0),
                'market_cap': 0,
                'dividend_yield': default_yields.get(symbol, 3.5),
                'fifty_two_week_high': 0,
                'fifty_two_week_low': 0,
                'avg_volume': 0,
                'beta': 0,
                'pe_ratio': 0,
                'expense_ratio': self.ETF_INFO.get(symbol, {}).get('expense_ratio', 0),
                'strategy': self.ETF_INFO.get(symbol, {}).get('strategy', 'N/A'),
                'description': self.ETF_INFO.get(symbol, {}).get('description', 'N/A'),
                'data_source': 'database',
                'data_date': str(latest.date),
            }

            if data['previous_close'] and data['current_price']:
                data['change'] = data['current_price'] - data['previous_close']
                data['change_percent'] = (data['change'] / data['previous_close']) * 100
            else:
                data['change'] = 0
                data['change_percent'] = 0

            return data
        except Exception as e:
            logger.error(f'数据库读取失败 {symbol}: {e}')

        # 返回错误信息
        return {
            'symbol': symbol,
            'error': '数据库中无数据',
            'name': self.ETF_INFO.get(symbol, {}).get('name', symbol),
        }
    
    def fetch_historical_data(self, symbol, period='1y'):
        """获取历史数据（Redis持久化 -> 数据库降级 -> API）"""
        # 1. 先从Redis获取
        redis_cache = self._get_redis_cache()
        if redis_cache:
            cached = redis_cache.get_historical(symbol, period)
            if cached is not None:
                logger.info(f'Redis命中历史数据: {symbol}_{period}')
                return cached
        
        # 2. Redis无数据，尝试从数据库读取
        try:
            from .models import ETFData
            from datetime import datetime, timedelta
            
            # 根据period计算日期范围
            end_date = datetime.now().date()
            period_map = {
                '1mo': 30, '3mo': 90, '6mo': 180,
                '1y': 365, '2y': 730, '5y': 1825, '10y': 3650
            }
            days = period_map.get(period, 365)
            start_date = end_date - timedelta(days=days)
            
            # 查询数据库
            db_data = ETFData.objects.filter(
                symbol=symbol,
                date__gte=start_date,
                date__lte=end_date
            ).order_by('date')
            
            if db_data.exists():
                # 转换为pandas DataFrame
                import pandas as pd
                data_list = []
                for item in db_data:
                    data_list.append({
                        'Date': item.date,
                        'Open': float(item.open_price or 0),
                        'High': float(item.high_price or 0),
                        'Low': float(item.low_price or 0),
                        'Close': float(item.close_price or 0),
                        'Volume': int(item.volume or 0),
                    })
                
                df = pd.DataFrame(data_list)
                df.set_index('Date', inplace=True)
                df.index = pd.to_datetime(df.index)
                
                logger.info(f'数据库命中历史数据: {symbol}_{period} ({len(df)}条)')
                
                # 写入Redis缓存
                if redis_cache:
                    redis_cache.set_historical(symbol, period, df)
                
                return df
        except Exception as e:
            logger.warning(f'从数据库读取失败: {e}')
        
        # 3. 数据库也无数据，从API获取
        try:
            ticker = yf.Ticker(symbol)
            data = ticker.history(period=period)
            
            if data.empty:
                return None
            
            logger.info(f'API获取历史数据: {symbol}_{period}')
            
            # 写入Redis（持久化）
            if redis_cache:
                redis_cache.set_historical(symbol, period, data)
            
            # 同时写入数据库
            try:
                from .models import ETFData
                for date, row in data.iterrows():
                    ETFData.objects.update_or_create(
                        symbol=symbol,
                        date=date.date(),
                        defaults={
                            'open_price': float(row['Open']),
                            'high_price': float(row['High']),
                            'low_price': float(row['Low']),
                            'close_price': float(row['Close']),
                            'volume': int(row['Volume']),
                            'data_source': 'yfinance',
                        }
                    )
                logger.info(f'同步写入数据库: {symbol} {len(data)}条')
            except Exception as e:
                logger.warning(f'写入数据库失败: {e}')
            
            return data
            
        except Exception as e:
            logger.error(f"Error fetching {symbol}: {e}")
            return None
    
    def calculate_metrics(self, symbol, period='1y'):
        """计算ETF指标（Redis持久化）"""
        # 先从Redis获取计算好的指标
        redis_cache = self._get_redis_cache()
        if redis_cache:
            cached_metrics = redis_cache.get_metrics(symbol, period)
            if cached_metrics:
                return cached_metrics
        
        data = self.fetch_historical_data(symbol, period)
        
        if data is None or data.empty:
            return None
        
        # 计算收益率
        returns = data['Close'].pct_change().dropna()
        
        # 计算各项指标
        metrics = {
            'symbol': symbol,
            'period': period,
            'start_price': float(data['Close'].iloc[0]),
            'end_price': float(data['Close'].iloc[-1]),
            'total_return': float((data['Close'].iloc[-1] / data['Close'].iloc[0] - 1) * 100),
            'avg_daily_return': float(returns.mean() * 100),
            'volatility': float(returns.std() * np.sqrt(252) * 100),  # 年化波动率
            'max_price': float(data['Close'].max()),
            'min_price': float(data['Close'].min()),
            'avg_volume': int(data['Volume'].mean()),
            'trading_days': len(data),
        }
        
        # 计算最大回撤
        cumulative = (1 + returns).cumprod()
        running_max = cumulative.cummax()
        drawdown = (cumulative - running_max) / running_max
        metrics['max_drawdown'] = float(drawdown.min() * 100)
        
        # 计算夏普比率（假设无风险利率4%）
        risk_free_rate = 0.04
        excess_returns = returns.mean() * 252 - risk_free_rate
        metrics['sharpe_ratio'] = float(excess_returns / (returns.std() * np.sqrt(252))) if returns.std() > 0 else 0
        
        # 写入Redis（持久化）
        if redis_cache:
            redis_cache.set_metrics(symbol, period, metrics)
        
        return metrics
    
    def _batch_fetch_historical(self, period='1y'):
        """批量下载历史数据（Redis -> 数据库 -> API）"""
        cache_key = f'batch_history_{period}'
        cached = self._get_cached(cache_key)
        if cached is not None:
            logger.info(f'批量历史数据命中缓存: {period}')
            return cached
        
        result = {}
        missing_symbols = []
        
        # 1. 先尝试从单个Redis缓存或数据库获取
        for symbol in self.SYMBOLS:
            data = self.fetch_historical_data(symbol, period)
            if data is not None and not data.empty:
                result[symbol] = data
            else:
                missing_symbols.append(symbol)
        
        # 2. 如果有缺失的，批量从API获取
        if missing_symbols:
            try:
                logger.info(f'批量API获取: {missing_symbols}')
                data = yf.download(missing_symbols, period=period, group_by='ticker', progress=False, threads=True)
                
                for symbol in missing_symbols:
                    if symbol in data.columns.get_level_values(0):
                        symbol_data = data[symbol].dropna()
                        if not symbol_data.empty:
                            result[symbol] = symbol_data
                            # 更新Redis缓存
                            redis_cache = self._get_redis_cache()
                            if redis_cache:
                                redis_cache.set_historical(symbol, period, symbol_data)
                            
                            # 同步写入数据库
                            try:
                                from .models import ETFData
                                for date, row in symbol_data.iterrows():
                                    ETFData.objects.update_or_create(
                                        symbol=symbol,
                                        date=date.date(),
                                        defaults={
                                            'open_price': float(row['Open']),
                                            'high_price': float(row['High']),
                                            'low_price': float(row['Low']),
                                            'close_price': float(row['Close']),
                                            'volume': int(row['Volume']),
                                            'data_source': 'yfinance',
                                        }
                                    )
                                logger.info(f'批量同步数据库: {symbol} {len(symbol_data)}条')
                            except Exception as e:
                                logger.warning(f'批量写入DB失败 {symbol}: {e}')
            except Exception as e:
                logger.error(f"Batch download error: {e}")
        
        # 3. 缓存批量结果
        if result:
            self._set_cache(cache_key, result, timeout=3600)
        
        return result
    
    def _fetch_single_realtime(self, symbol):
        """获取单个ETF实时数据（用于并行）"""
        return (symbol, self.fetch_realtime_data(symbol))
    
    def get_comparison_data(self, period='1y'):
        """获取所有ETF对比数据 - 优化版（并行+批量）"""
        comparison = []
        
        # 1. 批量获取历史数据（单次API调用）
        batch_history = self._batch_fetch_historical(period)
        
        # 2. 并行获取实时数据
        realtime_data = {}
        with ThreadPoolExecutor(max_workers=3) as executor:
            futures = {executor.submit(self._fetch_single_realtime, sym): sym for sym in self.SYMBOLS}
            for future in as_completed(futures):
                try:
                    symbol, data = future.result()
                    realtime_data[symbol] = data
                except Exception as e:
                    print(f"Error fetching realtime: {e}")
        
        # 3. 组合数据
        for symbol in self.SYMBOLS:
            realtime = realtime_data.get(symbol, {'symbol': symbol, 'error': 'Failed to fetch'})
            
            # 从批量数据计算指标
            hist_data = batch_history.get(symbol)
            if hist_data is not None and not hist_data.empty:
                metrics = self._calculate_metrics_from_data(symbol, hist_data, period)
            else:
                metrics = self.calculate_metrics(symbol, period)
            
            if metrics:
                data = {
                    **realtime,
                    **metrics,
                    'info': self.ETF_INFO.get(symbol, {})
                }
                comparison.append(data)
            else:
                comparison.append(realtime)
        
        return comparison
    
    def _calculate_metrics_from_data(self, symbol, data, period):
        """从已有数据计算指标（避免重复请求）"""
        if data is None or data.empty:
            return None
        
        returns = data['Close'].pct_change().dropna()
        
        metrics = {
            'symbol': symbol,
            'period': period,
            'start_price': float(data['Close'].iloc[0]),
            'end_price': float(data['Close'].iloc[-1]),
            'total_return': float((data['Close'].iloc[-1] / data['Close'].iloc[0] - 1) * 100),
            'avg_daily_return': float(returns.mean() * 100),
            'volatility': float(returns.std() * np.sqrt(252) * 100),
            'max_price': float(data['Close'].max()),
            'min_price': float(data['Close'].min()),
            'avg_volume': int(data['Volume'].mean()),
            'trading_days': len(data),
        }
        
        cumulative = (1 + returns).cumprod()
        running_max = cumulative.cummax()
        drawdown = (cumulative - running_max) / running_max
        metrics['max_drawdown'] = float(drawdown.min() * 100)
        
        risk_free_rate = 0.04
        excess_returns = returns.mean() * 252 - risk_free_rate
        metrics['sharpe_ratio'] = float(excess_returns / (returns.std() * np.sqrt(252))) if returns.std() > 0 else 0
        
        return metrics
    
    def get_historical_chart_data(self, symbol, period='1y'):
        """获取图表数据"""
        data = self.fetch_historical_data(symbol, period)
        
        if data is None or data.empty:
            return None
        
        # 转换为图表格式
        chart_data = {
            'dates': [d.strftime('%Y-%m-%d') for d in data.index],
            'close': data['Close'].round(2).tolist(),
            'volume': data['Volume'].tolist(),
            'high': data['High'].round(2).tolist(),
            'low': data['Low'].round(2).tolist(),
        }
        
        return chart_data
    
    def get_comparison_chart_data(self, period='1y'):
        """获取对比图表数据（归一化累计收益率）"""
        # 批量获取历史数据
        batch_history = self._batch_fetch_historical(period)

        if not batch_history:
            return None

        # 找到所有ETF的共同日期范围（只包含所有ETF都有数据的日期）
        # 先收集所有日期
        all_dates_set = set()
        for symbol, data in batch_history.items():
            if data is not None and not data.empty:
                all_dates_set.update(data.index.date)

        if not all_dates_set:
            return None

        # 排序日期
        sorted_dates = sorted(all_dates_set)

        # 找到所有ETF都有的日期（交集）
        common_dates = None
        for symbol, data in batch_history.items():
            if data is not None and not data.empty:
                symbol_dates = set(data.index.date)
                if common_dates is None:
                    common_dates = symbol_dates
                else:
                    common_dates = common_dates.intersection(symbol_dates)

        if not common_dates:
            # 如果没有共同日期，使用所有日期
            common_dates = all_dates_set

        sorted_common_dates = sorted(common_dates)
        min_common_date = min(sorted_common_dates)
        max_common_date = max(sorted_common_dates)

        # 准备图表数据
        chart_data = {
            'dates': [d.strftime('%Y-%m-%d') for d in sorted_common_dates],
            'series': {}
        }

        # 为每个ETF计算归一化累计收益率
        for symbol in self.SYMBOLS:
            data = batch_history.get(symbol)
            if data is None or data.empty:
                continue

            # 筛选共同日期范围内的数据
            filtered_data = data[(data.index.date >= min_common_date) & (data.index.date <= max_common_date)]

            if filtered_data.empty:
                continue

            # 计算归一化累计收益率（以第一天为基准0%）
            initial_price = filtered_data['Close'].iloc[0]
            normalized_returns = ((filtered_data['Close'] / initial_price - 1) * 100).round(2)

            # 构建与所有共同日期对应的数据序列
            return_series = []
            for date in sorted_common_dates:
                # 查找该日期的数据
                date_mask = filtered_data.index.date == date
                if date_mask.any():
                    # 找到匹配的索引
                    matching_indices = filtered_data.index[date_mask]
                    if not matching_indices.empty:
                        idx = matching_indices[0]
                        # 在normalized_returns中找到对应位置
                        if idx in normalized_returns.index:
                            return_series.append(float(normalized_returns.loc[idx]))
                        else:
                            # 如果索引不匹配，尝试使用最接近的值
                            return_series.append(float(normalized_returns.iloc[-1]) if len(normalized_returns) > 0 else 0.0)
                    else:
                        # 使用前一个有效值（前向填充）
                        return_series.append(float(return_series[-1]) if return_series else 0.0)
                else:
                    # 该ETF在这个日期没有数据，使用前一个有效值（前向填充）
                    return_series.append(float(return_series[-1]) if return_series else 0.0)

            chart_data['series'][symbol] = return_series

        return chart_data
    
    def analyze_portfolio(self, allocation, total_investment=10000, tax_rate=0.10, base_currency='USD'):
        """
        分析投资组合（支持多货币ETF，自动汇率换算）

        参数:
            allocation: 配置比例，如 {'SCHD': 0.4, 'SPYD': 0.3, '3466.HK': 0.3}
            total_investment: 总投资金额（美元）
            tax_rate: 股息税率（默认10%，中国大陆居民W-8BEN表格税率）
            base_currency: 基础货币（默认为USD）
        """
        result = {
            'total_investment': total_investment,
            'base_currency': base_currency,
            'allocation': allocation,
            'holdings': [],
            'total_value': 0,
            'total_value_usd': 0,
            'total_return': 0,
            'total_return_usd': 0,
            'weighted_dividend_yield': 0,
            'portfolio_metrics': {},
            'tax_rate': tax_rate * 100,  # 转换为百分比
            'exchange_rates': {},  # 使用的汇率信息
        }

        all_returns = []
        weights = []
        total_annual_dividend_before_tax = 0  # 税前股息
        total_annual_dividend_after_tax = 0   # 税后股息

        for symbol, weight in allocation.items():
            if symbol not in self.SYMBOLS:
                continue

            # 跳过权重为 0 的 ETF
            if weight <= 0:
                continue

            realtime = self.fetch_realtime_data(symbol)
            metrics = self.calculate_metrics(symbol, '1y')

            # 获取ETF的计价货币
            etf_currency = self.get_etf_currency(symbol)

            # 计算投资金额（美元）
            investment_amount_usd = total_investment * weight

            # 根据ETF的计价货币计算实际投资金额
            if etf_currency == 'USD':
                investment_amount = investment_amount_usd
            elif etf_currency == 'CNY':
                investment_amount = investment_amount_usd * self.get_exchange_rate('USD', 'CNY')
            elif etf_currency == 'HKD':
                investment_amount = investment_amount_usd * self.get_exchange_rate('USD', 'HKD')
            else:
                investment_amount = investment_amount_usd

            current_price = realtime.get('current_price', 0)

            if current_price and current_price > 0:
                shares = investment_amount / current_price
                current_value = shares * current_price

                # 将当前价值转换为美元
                current_value_usd = self.convert_to_usd(current_value, etf_currency)

                # 计算股息
                dividend_yield_decimal = realtime.get('dividend_yield', 0) / 100  # 转回小数
                annual_dividend_before_tax = current_value * dividend_yield_decimal
                annual_dividend_before_tax_usd = self.convert_to_usd(annual_dividend_before_tax, etf_currency)
                annual_dividend_after_tax = annual_dividend_before_tax * (1 - tax_rate)
                annual_dividend_after_tax_usd = self.convert_to_usd(annual_dividend_after_tax, etf_currency)

                # 计算资本利得（基于历史区间收益率预估）
                historical_return_rate = metrics.get('total_return', 0) / 100 if metrics else 0
                capital_gain = investment_amount_usd * historical_return_rate  # 基于美元计算
                capital_gain_percent = historical_return_rate * 100

                holding = {
                    'symbol': symbol,
                    'name': realtime.get('name', symbol),
                    'currency': etf_currency,  # ETF的计价货币
                    'weight': float(weight * 100),  # 转换为百分比数值（17），用于模板显示
                    'investment': round(investment_amount_usd, 2),  # 美元投资金额（统一使用USD显示）
                    'investment_usd': round(investment_amount_usd, 2),  # 美元投资金额
                    'shares': round(shares, 4),
                    'current_price': current_price,
                    'current_value': round(current_value_usd, 2),  # 美元当前价值（统一使用USD显示）
                    'current_value_usd': round(current_value_usd, 2),  # 美元当前价值
                    'dividend_yield': realtime.get('dividend_yield', 0),  # 已经是百分比数值
                    'annual_dividend_before_tax': round(annual_dividend_after_tax_usd, 2),  # 美元税前股息（统一使用USD）
                    'annual_dividend_before_tax_usd': round(annual_dividend_before_tax_usd, 2),  # 美元税前股息
                    'annual_dividend_after_tax': round(annual_dividend_after_tax_usd, 2),  # 美元税后股息（统一使用USD）
                    'annual_dividend_after_tax_usd': round(annual_dividend_after_tax_usd, 2),  # 美元税后股息
                    'capital_gain': round(capital_gain, 2),  # 预期资本利得（基于历史收益率，美元）
                    'capital_gain_percent': round(capital_gain_percent, 2),  # 预期资本利得率
                }

                # 保存使用的汇率信息
                if etf_currency != 'USD':
                    rate = self.get_exchange_rate(etf_currency, 'USD')
                    result['exchange_rates'][f'{etf_currency}_to_USD'] = rate

                if metrics:
                    holding['total_return'] = metrics.get('total_return', 0)  # 历史区间收益
                    holding['volatility'] = metrics.get('volatility', 0)
                    all_returns.append(metrics.get('total_return', 0) / 100)
                    weights.append(weight)

                result['holdings'].append(holding)
                result['total_value'] += current_value_usd  # 累加美元当前价值
                result['total_value_usd'] += current_value_usd
                result['weighted_dividend_yield'] += (realtime.get('dividend_yield', 0) or 0) * weight
                total_annual_dividend_before_tax += annual_dividend_before_tax_usd
                total_annual_dividend_after_tax += annual_dividend_after_tax_usd

        # 计算组合预期资本利得（基于历史加权收益率，美元）
        total_capital_gain = sum(h.get('capital_gain', 0) for h in result['holdings'])
        result['total_return'] = round(total_capital_gain, 2)
        result['total_return_percent'] = round((total_capital_gain / total_investment) * 100, 2) if total_investment > 0 else 0

        # 股息收入（美元）
        result['annual_dividend_before_tax'] = round(total_annual_dividend_before_tax, 2)
        result['annual_dividend_after_tax'] = round(total_annual_dividend_after_tax, 2)
        result['dividend_tax'] = round(total_annual_dividend_before_tax * tax_rate, 2)

        # 综合收益（资本利得 + 税后股息，美元）
        result['total_return_with_dividend'] = result['total_return'] + result['annual_dividend_after_tax']
        result['total_return_with_dividend_percent'] = (result['total_return_with_dividend'] / total_investment) * 100 if total_investment > 0 else 0

        # 计算组合加权收益
        if all_returns and weights:
            result['portfolio_metrics']['weighted_return'] = sum(r * w for r, w in zip(all_returns, weights)) * 100

        return result

    def forecast_etf_growth(self, symbol, initial_investment=10000, annual_return_rate=None, tax_rate=0.10):
        """
        根据当前年化收益率预测ETF未来3、5、10年收益
        
        参数:
            symbol: ETF代码
            initial_investment: 初始投资金额
            annual_return_rate: 年化收益率（如果为None则使用当前ETF的历史年化收益率）
            tax_rate: 股息税率（默认10%）
        """
        # 获取ETF实时数据
        realtime_data = self.fetch_realtime_data(symbol)
        
        if 'error' in realtime_data:
            return {
                'error': f'无法获取{symbol}的实时数据',
                'symbol': symbol
            }
        
        # 如果没有提供年化收益率，则使用当前ETF的历史年化收益率
        if annual_return_rate is None:
            metrics = self.calculate_metrics(symbol, '1y')
            if metrics:
                # 使用平均日年化收益率（avg_daily_return * 252）作为年化收益率的参考
                avg_daily_return = metrics.get('avg_daily_return', 0) / 100
                annual_return_rate = avg_daily_return * 252

                # 检查数据时间跨度（trading_days）
                trading_days = metrics.get('trading_days', 0)
                # 如果数据不足180个交易日（约9个月），对年化收益率进行调整
                if trading_days < 180:
                    # 使用总收益率作为参考，并进行保守调整
                    total_return = metrics.get('total_return', 0) / 100
                    # 对于短期数据，使用年化后的总收益率（保守估计）
                    annual_return_rate = total_return / (trading_days / 252) * 0.7  # 0.7是保守因子

                # 限制合理的年化收益率范围（-30% 到 50%）
                annual_return_rate = max(-0.30, min(0.50, annual_return_rate))
            else:
                # 如果无法获取历史数据，使用预设值
                default_returns = {
                    'SCHD': 0.08, 'SPYD': 0.09, 'JEPQ': 0.12, 'JEPI': 0.08,
                    '3466.HK': 0.08, '510300': 0.08, 'VYM': 0.08
                }
                annual_return_rate = default_returns.get(symbol, 0.08)

        # 获取当前股息率
        dividend_yield = realtime_data.get('dividend_yield', 3.5) / 100  # 转换为小数

        # 预测年份
        forecast_years = [3, 5, 10]

        result = {
            'symbol': symbol,
            'initial_investment': initial_investment,
            'annual_return_rate': annual_return_rate * 100,  # 转换为百分比
            'dividend_yield': dividend_yield * 100,  # 转换为百分比
            'tax_rate': tax_rate * 100,  # 转换为百分比
            'forecasts': {}
        }

        for years in forecast_years:
            # 计算未来资产价值（复利增长，不包括股息再投资）
            future_value = initial_investment * (1 + annual_return_rate) ** years

            # 计算资本增值
            capital_appreciation = future_value - initial_investment

            # 计算每年的股息（基于当年的资产价值）
            total_dividend_before_tax = 0
            total_dividend_after_tax = 0

            for year in range(1, years + 1):
                # 当年的资产价值
                year_value = initial_investment * (1 + annual_return_rate) ** year
                # 当年的股息
                annual_dividend_before_tax = year_value * dividend_yield
                # 当年的税后股息
                annual_dividend_after_tax = annual_dividend_before_tax * (1 - tax_rate)

                total_dividend_before_tax += annual_dividend_before_tax
                total_dividend_after_tax += annual_dividend_after_tax

            # 计算股息税额
            dividend_tax = total_dividend_before_tax - total_dividend_after_tax

            # 计算总收益（资本增值 + 税后股息）
            total_return_after_tax = capital_appreciation + total_dividend_after_tax

            # 计算有效年化收益率（包含资本增值和股息，但假设股息不 reinvest）
            # (initial + capital_gain + dividends) / initial ** (1/years) - 1
            total_value = initial_investment + total_return_after_tax
            effective_annual_return = (total_value / initial_investment) ** (1 / years) - 1

            # 计算第一年的股息（用于显示）
            first_year_dividend_before_tax = (initial_investment * (1 + annual_return_rate)) * dividend_yield
            first_year_dividend_after_tax = first_year_dividend_before_tax * (1 - tax_rate)

            result['forecasts'][str(years)] = {
                'years': years,
                'future_value': round(future_value, 2),
                'capital_appreciation': round(capital_appreciation, 2),
                'total_dividend_before_tax': round(total_dividend_before_tax, 2),
                'total_dividend_after_tax': round(total_dividend_after_tax, 2),
                'annual_dividend_before_tax': round(first_year_dividend_before_tax, 2),
                'annual_dividend_after_tax': round(first_year_dividend_after_tax, 2),
                'dividend_tax': round(dividend_tax, 2),
                'total_return_after_tax': round(total_return_after_tax, 2),
                'effective_annual_return_rate': round(effective_annual_return * 100, 2)
            }

        return result

    def forecast_portfolio_growth(self, allocation, total_investment=10000, tax_rate=0.10, scenarios=None):
        """
        预测投资组合未来增长（乐观、中性、悲观情况）- 支持多货币ETF

        参数:
            allocation: 配置比例，如 {'SCHD': 0.4, 'SPYD': 0.3, '3466.HK': 0.3}
            total_investment: 总投资金额（美元）
            tax_rate: 股息税率（默认10%）
            scenarios: 场景收益率，如 {'optimistic': 0.12, 'neutral': 0.08, 'pessimistic': 0.04}
        """
        if scenarios is None:
            # 默认场景收益率（基于历史数据和市场预期）
            scenarios = {
                'optimistic': 0.12,  # 乐观：年化12%
                'neutral': 0.08,     # 中性：年化8%
                'pessimistic': 0.04  # 悲观：年化4%
            }

        result = {
            'total_investment': total_investment,
            'allocation': allocation,
            'tax_rate': tax_rate * 100,  # 转换为百分比
            'base_currency': 'USD',  # 以美元为基准
            'scenarios': {}
        }

        # 计算当前组合的加权股息率
        weighted_dividend_yield = 0
        for symbol, weight in allocation.items():
            if symbol in self.SYMBOLS:
                realtime = self.fetch_realtime_data(symbol)
                dividend_yield = realtime.get('dividend_yield', 0) / 100  # 转换为小数
                weighted_dividend_yield += dividend_yield * weight

        result['current_weighted_dividend_yield'] = weighted_dividend_yield * 100  # 转换为百分比

        # 为每个场景计算未来3、5、10年的收益
        for scenario_name, annual_return in scenarios.items():
            scenario_result = {
                'annual_return_rate': annual_return * 100,  # 转换为百分比
                'years': {}
            }

            # 计算未来3、5、10年的数据
            for years in [3, 5, 10]:
                # 复合增长计算（包含股息再投资）- 以美元计算
                future_value_with_reinvestment = total_investment * (1 + annual_return) ** years

                # 计算税前年化股息收入（美元）
                annual_dividend_before_tax = total_investment * weighted_dividend_yield
                total_dividend_before_tax = annual_dividend_before_tax * years

                # 计算税后年化股息收入（美元）
                annual_dividend_after_tax = annual_dividend_before_tax * (1 - tax_rate)
                total_dividend_after_tax = annual_dividend_after_tax * years

                # 计算资本增值（假设价格增长，美元）
                capital_appreciation = future_value_with_reinvestment - total_investment

                # 计算未来资产价值（考虑股息再投资，美元）
                future_value = total_investment * (1 + annual_return) ** years

                scenario_result['years'][str(years)] = {
                    'future_value': round(future_value, 2),  # 美元
                    'capital_appreciation': round(capital_appreciation, 2),  # 美元
                    'total_dividend_before_tax': round(total_dividend_before_tax, 2),  # 美元
                    'total_dividend_after_tax': round(total_dividend_after_tax, 2),  # 美元
                    'annual_dividend_before_tax': round(annual_dividend_before_tax, 2),  # 美元
                    'annual_dividend_after_tax': round(annual_dividend_after_tax, 2),  # 美元
                    'total_return_before_tax': round(capital_appreciation + total_dividend_before_tax, 2),  # 美元
                    'total_return_after_tax': round(capital_appreciation + total_dividend_after_tax, 2),  # 美元
                    'annual_return_rate': annual_return * 100,
                    'dividend_tax': round(total_dividend_before_tax * tax_rate, 2)  # 美元
                }

            result['scenarios'][scenario_name] = scenario_result

        return result




# 全局服务实例
etf_service = ETFAnalysisService()
