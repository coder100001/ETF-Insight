#!/usr/bin/env python
"""
实时ETF数据更新脚本
使用yfinance API获取最新实时价格并更新到数据库
"""

import os
import sys
import django
from datetime import date

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

from workflow.models import ETFData
import yfinance as yf
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

def update_realtime_etf_data():
    """更新所有ETF的实时数据"""
    symbols = ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM']
    today = date.today()
    
    logger.info('=' * 60)
    logger.info('开始更新实时ETF数据')
    logger.info(f'今天日期: {today}')
    logger.info('=' * 60)
    
    success_count = 0
    failed_count = 0
    
    for symbol in symbols:
        try:
            logger.info(f'获取 {symbol} 实时数据...')
            
            # 使用yfinance获取实时数据
            ticker = yf.Ticker(symbol)
            hist = ticker.history(period='1d')
            
            if not hist.empty:
                # 获取最新价格数据
                latest = hist.iloc[-1]
                current_price = latest['Close']
                open_price = latest['Open']
                high_price = latest['High'] 
                low_price = latest['Low']
                volume = latest['Volume']
                
                logger.info(f'  - 当前价格: ${current_price:.2f}')
                logger.info(f'  - 开盘价: ${open_price:.2f}')
                logger.info(f'  - 最高价: ${high_price:.2f}')
                logger.info(f'  - 最低价: ${low_price:.2f}')
                logger.info(f'  - 成交量: {volume:,}')
                
                # 更新或创建数据库记录
                etf_data, created = ETFData.objects.update_or_create(
                    symbol=symbol,
                    date=today,
                    defaults={
                        'close_price': current_price,
                        'open_price': open_price,
                        'high_price': high_price,
                        'low_price': low_price,
                        'volume': volume,
                        'data_source': 'yfinance_realtime'
                    }
                )
                
                if created:
                    logger.info(f'  ✓ {symbol} 创建新记录')
                else:
                    logger.info(f'  ✓ {symbol} 更新现有记录')
                
                success_count += 1
                
            else:
                logger.error(f'  ✗ {symbol} 无可用数据')
                failed_count += 1
                
        except Exception as e:
            logger.error(f'  ✗ {symbol} 更新失败: {e}')
            failed_count += 1
    
    logger.info('=' * 60)
    logger.info('实时数据更新完成')
    logger.info(f'成功: {success_count}/{len(symbols)}')
    logger.info(f'失败: {failed_count}/{len(symbols)}')
    logger.info('=' * 60)

if __name__ == '__main__':
    update_realtime_etf_data()