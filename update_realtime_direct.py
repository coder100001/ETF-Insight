#!/usr/bin/env python
"""
直接更新ETF实时数据
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

import yfinance as yf
from datetime import datetime, date
from workflow.models import ETFConfig, ETFData
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s [%(levelname)s] %(message)s')
logger = logging.getLogger(__name__)

def update_realtime_data():
    """更新所有启用ETF的实时数据"""
    logger.info("=" * 60)
    logger.info("开始更新ETF实时数据")
    logger.info("=" * 60)

    # 获取所有启用的ETF配置
    enabled_etfs = ETFConfig.objects.filter(status=1).values_list('symbol', flat=True)
    logger.info(f"启用的ETF: {list(enabled_etfs)}")
    logger.info("")

    success_count = 0
    error_count = 0
    today = date.today()

    for symbol in enabled_etfs:
        try:
            logger.info(f"更新 {symbol} 实时数据...")

            # 使用yfinance获取实时数据
            ticker = yf.Ticker(symbol)
            info = ticker.info

            # 提取所需数据
            current_price = info.get('regularMarketPrice') or info.get('previousClose')
            if not current_price:
                logger.error(f"无法获取 {symbol} 的价格")
                error_count += 1
                continue

            # 获取开盘价和最新价
            open_price = info.get('regularMarketOpen', current_price)
            high_price = info.get('regularMarketDayHigh', current_price)
            low_price = info.get('regularMarketDayLow', current_price)
            volume = info.get('regularMarketVolume')

            # 保存到数据库（更新今天的数据或创建新数据）
            ETFData.objects.update_or_create(
                symbol=symbol,
                date=today,
                defaults={
                    'open_price': open_price,
                    'close_price': current_price,
                    'high_price': high_price,
                    'low_price': low_price,
                    'volume': volume,
                    'data_source': 'yfinance_realtime'
                }
            )

            logger.info(f"  ✓ 开盘价: ${open_price}")
            logger.info(f"  ✓ 最新价: ${current_price}")
            logger.info(f"  ✓ 最高价: ${high_price}")
            logger.info(f"  ✓ 最低价: ${low_price}")
            logger.info(f"  ✓ 更新时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
            success_count += 1

        except Exception as e:
            logger.error(f"✗ 更新 {symbol} 失败: {str(e)}")
            error_count += 1

        logger.info("")

    logger.info("=" * 60)
    logger.info(f"更新完成")
    logger.info(f"成功: {success_count}/{len(enabled_etfs)}")
    logger.info(f"失败: {error_count}")
    logger.info("=" * 60)

if __name__ == '__main__':
    update_realtime_data()
