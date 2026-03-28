"""
Redis缓存管理器 - ETF数据持久化缓存
"""

from django.core.cache import caches
import json
import logging

logger = logging.getLogger(__name__)


class ETFCacheManager:
    """ETF数据Redis缓存管理器"""
    
    def __init__(self):
        # 使用专门的etf_data缓存（持久化）
        self.cache = caches['etf_data']
        self.prefix = 'etf'
    
    def _make_key(self, category, identifier):
        """生成缓存键"""
        return f'{self.prefix}:{category}:{identifier}'
    
    def get_realtime(self, symbol):
        """获取实时数据缓存"""
        key = self._make_key('realtime', symbol)
        return self.cache.get(key)
    
    def set_realtime(self, symbol, data):
        """设置实时数据缓存（1小时过期）"""
        key = self._make_key('realtime', symbol)
        self.cache.set(key, data, timeout=3600)
        logger.info(f'Redis缓存实时数据: {symbol}')
    
    def get_historical(self, symbol, period):
        """获取历史数据缓存"""
        key = self._make_key('historical', f'{symbol}_{period}')
        cached = self.cache.get(key)
        if cached is not None:  # 使用is not None避免pandas DataFrame判断问题
            logger.info(f'Redis命中历史数据: {symbol}_{period}')
        return cached
    
    def set_historical(self, symbol, period, data):
        """设置历史数据缓存（持久化，不过期）"""
        key = self._make_key('historical', f'{symbol}_{period}')
        self.cache.set(key, data, timeout=None)
        logger.info(f'Redis缓存历史数据: {symbol}_{period}')
    
    def get_metrics(self, symbol, period):
        """获取指标数据缓存"""
        key = self._make_key('metrics', f'{symbol}_{period}')
        return self.cache.get(key)
    
    def set_metrics(self, symbol, period, metrics):
        """设置指标数据缓存（持久化）"""
        key = self._make_key('metrics', f'{symbol}_{period}')
        self.cache.set(key, metrics, timeout=None)
        logger.info(f'Redis缓存指标数据: {symbol}_{period}')
    
    def get_comparison(self, period):
        """获取对比数据缓存"""
        key = self._make_key('comparison', period)
        return self.cache.get(key)
    
    def set_comparison(self, period, data):
        """设置对比数据缓存（1小时过期）"""
        key = self._make_key('comparison', period)
        self.cache.set(key, data, timeout=3600)
        logger.info(f'Redis缓存对比数据: {period}')
    
    def clear_symbol(self, symbol):
        """清除指定ETF的所有缓存"""
        patterns = [
            f'{self.prefix}:realtime:{symbol}',
            f'{self.prefix}:historical:{symbol}_*',
            f'{self.prefix}:metrics:{symbol}_*',
        ]
        
        cleared = 0
        for pattern in patterns:
            try:
                # 使用Redis的keys命令查找匹配的键
                keys = self.cache.keys(pattern)
                for key in keys:
                    self.cache.delete(key)
                    cleared += 1
            except Exception as e:
                logger.warning(f'清除缓存失败 {pattern}: {e}')
        
        logger.info(f'清除{symbol}缓存: {cleared}个键')
        return cleared
    
    def clear_all(self):
        """清除所有ETF缓存"""
        try:
            self.cache.clear()
            logger.info('清除所有ETF缓存')
            return True
        except Exception as e:
            logger.error(f'清除所有缓存失败: {e}')
            return False
    
    def clear_comparison(self):
        """清除对比数据缓存"""
        pattern = f'{self.prefix}:comparison:*'
        try:
            keys = self.cache.keys(pattern)
            for key in keys:
                self.cache.delete(key)
            logger.info(f'清除对比数据缓存: {len(keys)}个键')
            return len(keys)
        except Exception as e:
            logger.warning(f'清除对比缓存失败: {e}')
            return 0
    
    def get_cache_stats(self):
        """获取缓存统计信息"""
        try:
            stats = {
                'realtime_count': len(self.cache.keys(f'{self.prefix}:realtime:*')),
                'historical_count': len(self.cache.keys(f'{self.prefix}:historical:*')),
                'metrics_count': len(self.cache.keys(f'{self.prefix}:metrics:*')),
                'comparison_count': len(self.cache.keys(f'{self.prefix}:comparison:*')),
            }
            stats['total_count'] = sum(stats.values())
            return stats
        except Exception as e:
            logger.error(f'获取缓存统计失败: {e}')
            return None


# 全局缓存管理器实例
etf_cache = ETFCacheManager()
