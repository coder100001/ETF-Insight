#!/usr/bin/env python
"""
实时汇率更新脚本
使用yfinance获取USD, CNY, HKD之间的实时汇率
"""
import os
import sys
import django

# 设置Django环境
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
django.setup()

import yfinance as yf
from datetime import date
from workflow.models import ExchangeRate
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s [%(levelname)s] %(message)s')
logger = logging.getLogger(__name__)

# 货币对到Yahoo Finance ticker的映射
CURRENCY_PAIRS = {
    ('USD', 'CNY'): 'USDCNY=X',  # 1 USD = ? CNY
    ('USD', 'HKD'): 'USDHKD=X',  # 1 USD = ? HKD
    ('CNY', 'USD'): 'CNY=X',     # 1 CNY = ? USD
    ('HKD', 'USD'): 'HKD=X',     # 1 HKD = ? USD
    ('CNY', 'HKD'): 'CNYHKD=X',  # 1 CNY = ? HKD
    ('HKD', 'CNY'): 'HKDCNY=X',  # 1 HKD = ? CNY
}

def update_exchange_rate(from_currency, to_currency, rate, data_source='yfinance'):
    """
    更新或创建汇率记录
    
    参数:
        from_currency: 源货币（USD, CNY, HKD）
        to_currency: 目标货币（USD, CNY, HKD）
        rate: 汇率（1单位from_currency = ?单位to_currency）
        data_source: 数据来源
    """
    today = date.today()

    try:
        # 查找是否已存在今日的汇率记录
        rate_record = ExchangeRate.objects.filter(
            from_currency=from_currency,
            to_currency=to_currency,
            rate_date=today
        ).first()

        if rate_record:
            # 更新现有记录
            rate_record.rate = rate
            rate_record.data_source = data_source
            rate_record.save()
            logger.info(f"  ✓ 更新: 1 {from_currency} = {rate:.6f} {to_currency}")
        else:
            # 创建新记录
            ExchangeRate.objects.create(
                from_currency=from_currency,
                to_currency=to_currency,
                rate=rate,
                rate_date=today,
                data_source=data_source
            )
            logger.info(f"  ✓ 创建: 1 {from_currency} = {rate:.6f} {to_currency}")
        return True
    except Exception as e:
        logger.error(f"  ✗ 更新 {from_currency} → {to_currency} 失败: {str(e)}")
        return False

def fetch_realtime_rates():
    """获取实时汇率"""
    logger.info("=" * 60)
    logger.info("开始获取实时汇率")
    logger.info("=" * 60)
    logger.info("")

    today = date.today()
    success_count = 0
    total_count = len(CURRENCY_PAIRS)

    for (from_currency, to_currency), ticker_symbol in CURRENCY_PAIRS.items():
        try:
            logger.info(f"获取 {from_currency} → {to_currency} 汇率...")
            
            # 使用yfinance获取汇率
            ticker = yf.Ticker(ticker_symbol)
            info = ticker.info
            
            # 获取当前汇率
            current_rate = info.get('regularMarketPrice')
            
            if not current_rate:
                # 尝试使用历史数据中的最新价格
                hist = ticker.history(period='1d')
                if not hist.empty:
                    current_rate = hist['Close'].iloc[-1]
                else:
                    logger.warning(f"  ! 无法获取 {ticker_symbol} 的汇率，跳过")
                    continue
            
            # 保存到数据库
            if update_exchange_rate(from_currency, to_currency, current_rate, 'yfinance'):
                success_count += 1
            
            logger.info("")
            
        except Exception as e:
            logger.error(f"✗ 获取 {from_currency} → {to_currency} 失败: {str(e)}")
            logger.error(f"  Ticker: {ticker_symbol}")
            logger.info("")

    logger.info("=" * 60)
    logger.info(f"更新完成")
    logger.info(f"成功: {success_count}/{total_count}")
    logger.info(f"失败: {total_count - success_count}")
    logger.info("=" * 60)
    logger.info("")

    list_exchange_rates()

def list_exchange_rates():
    """列出所有今日汇率"""
    today = date.today()
    rates = ExchangeRate.objects.filter(rate_date=today).order_by('from_currency', 'to_currency')

    logger.info(f"今日汇率 ({today}):")
    logger.info("-" * 60)
    for rate in rates:
        logger.info(f"  1 {rate.from_currency:4s} = {rate.rate:10.6f} {rate.to_currency:4s}  ({rate.data_source})")
    logger.info("-" * 60)

if __name__ == '__main__':
    fetch_realtime_rates()
