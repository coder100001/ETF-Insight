#!/usr/bin/env python
"""
ETF数据更新脚本
功能：先清除Redis缓存，再更新MySQL数据库和Redis持久化缓存
"""

import os
import sys
import django
from datetime import datetime

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.cache_manager import etf_cache
from workflow.services import etf_service
from workflow.models import ETFData
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)


class ETFDataUpdater:
    """ETF数据更新器"""
    
    def __init__(self):
        self.symbols = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']
        self.periods = ['1mo', '3mo', '6mo', '1y', '2y', '5y']
    
    def clear_cache(self, symbol=None):
        """清除缓存"""
        logger.info('='*60)
        logger.info('步骤1：清除缓存')
        logger.info('='*60)
        
        if symbol:
            # 清除指定ETF缓存
            count = etf_cache.clear_symbol(symbol)
            logger.info(f'清除{symbol}缓存: {count}个键')
        else:
            # 清除所有ETF缓存
            etf_cache.clear_all()
            logger.info('清除所有ETF缓存完成')
    
    def update_mysql(self, symbol, period='1y'):
        """更新MySQL数据库"""
        logger.info(f'\n更新MySQL数据: {symbol} ({period})')
        
        try:
            # 获取历史数据
            data = etf_service.fetch_historical_data(symbol, period)
            if data is None or data.empty:
                logger.warning(f'{symbol} 无数据')
                return False
            
            # 保存到数据库
            saved_count = 0
            for index, row in data.iterrows():
                etf_data, created = ETFData.objects.update_or_create(
                    symbol=symbol,
                    date=index.date(),
                    defaults={
                        'open_price': float(row['Open']),
                        'high_price': float(row['High']),
                        'low_price': float(row['Low']),
                        'close_price': float(row['Close']),
                        'volume': int(row['Volume']),
                    }
                )
                saved_count += 1
            
            logger.info(f'  MySQL保存: {saved_count}条记录')
            return True
            
        except Exception as e:
            logger.error(f'MySQL更新失败: {e}')
            return False
    
    def update_redis(self, symbol, period='1y'):
        """更新Redis持久化缓存"""
        logger.info(f'更新Redis缓存: {symbol} ({period})')
        
        try:
            # 获取实时数据
            realtime = etf_service.fetch_realtime_data(symbol)
            logger.info(f'  实时数据: {realtime.get("current_price", "N/A")}')
            
            # 获取历史数据
            historical = etf_service.fetch_historical_data(symbol, period)
            if historical is not None:
                logger.info(f'  历史数据: {len(historical)}条')
            
            # 计算指标
            metrics = etf_service.calculate_metrics(symbol, period)
            if metrics:
                logger.info(f'  指标计算: 收益率{metrics["total_return"]:.2f}%')
            
            return True
            
        except Exception as e:
            logger.error(f'Redis更新失败: {e}')
            return False
    
    def update_symbol(self, symbol):
        """更新单个ETF的所有数据"""
        logger.info('\n' + '='*60)
        logger.info(f'更新 {symbol} 数据')
        logger.info('='*60)
        
        # 1. 清除该ETF的缓存
        self.clear_cache(symbol)
        
        # 2. 更新MySQL和Redis
        logger.info('\n步骤2：更新数据')
        logger.info('-'*60)
        
        success = True
        for period in self.periods:
            # 更新MySQL
            if not self.update_mysql(symbol, period):
                success = False
            
            # 更新Redis
            if not self.update_redis(symbol, period):
                success = False
        
        # 3. 清除对比数据缓存
        etf_cache.clear_comparison()
        
        return success
    
    def update_all(self):
        """更新所有ETF数据"""
        logger.info('\n' + '#'*60)
        logger.info('开始更新所有ETF数据')
        logger.info('#'*60)
        
        start_time = datetime.now()
        
        # 1. 清除所有缓存
        self.clear_cache()
        
        # 2. 更新每个ETF
        logger.info('\n步骤2：更新所有ETF数据')
        logger.info('='*60)
        
        results = {}
        for symbol in self.symbols:
            logger.info(f'\n处理 {symbol}...')
            results[symbol] = self.update_symbol(symbol)
        
        # 3. 输出统计
        end_time = datetime.now()
        duration = (end_time - start_time).total_seconds()
        
        logger.info('\n' + '#'*60)
        logger.info('更新完成')
        logger.info('#'*60)
        logger.info(f'总耗时: {duration:.2f}秒')
        
        success_count = sum(1 for v in results.values() if v)
        logger.info(f'成功: {success_count}/{len(self.symbols)}')
        
        for symbol, success in results.items():
            status = '✓' if success else '✗'
            logger.info(f'  {status} {symbol}')
        
        # 4. 缓存统计
        stats = etf_cache.get_cache_stats()
        if stats:
            logger.info('\nRedis缓存统计:')
            logger.info(f'  实时数据: {stats["realtime_count"]}')
            logger.info(f'  历史数据: {stats["historical_count"]}')
            logger.info(f'  指标数据: {stats["metrics_count"]}')
            logger.info(f'  对比数据: {stats["comparison_count"]}')
            logger.info(f'  总计: {stats["total_count"]}个键')
        
        return all(results.values())


def main():
    """主函数"""
    import argparse
    
    parser = argparse.ArgumentParser(description='ETF数据更新脚本')
    parser.add_argument('--symbol', type=str, help='指定ETF符号 (SCHD/SPYD/JEPQ)')
    parser.add_argument('--all', action='store_true', help='更新所有ETF')
    parser.add_argument('--clear-only', action='store_true', help='仅清除缓存')
    parser.add_argument('--stats', action='store_true', help='查看缓存统计')
    
    args = parser.parse_args()
    
    updater = ETFDataUpdater()
    
    if args.stats:
        # 查看缓存统计
        stats = etf_cache.get_cache_stats()
        if stats:
            print('\nRedis缓存统计:')
            print(f'  实时数据: {stats["realtime_count"]}')
            print(f'  历史数据: {stats["historical_count"]}')
            print(f'  指标数据: {stats["metrics_count"]}')
            print(f'  对比数据: {stats["comparison_count"]}')
            print(f'  总计: {stats["total_count"]}个键\n')
    
    elif args.clear_only:
        # 仅清除缓存
        if args.symbol:
            updater.clear_cache(args.symbol.upper())
        else:
            updater.clear_cache()
    
    elif args.symbol:
        # 更新指定ETF
        symbol = args.symbol.upper()
        if symbol not in updater.symbols:
            logger.error(f'无效的ETF符号: {symbol}')
            logger.info(f'支持的符号: {", ".join(updater.symbols)}')
            sys.exit(1)
        
        success = updater.update_symbol(symbol)
        sys.exit(0 if success else 1)
    
    elif args.all:
        # 更新所有ETF
        success = updater.update_all()
        sys.exit(0 if success else 1)
    
    else:
        parser.print_help()


if __name__ == '__main__':
    main()
