#!/usr/bin/env python
"""
检查当前最新的ETF数据
显示真正的实时数据对比
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
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

def check_current_data():
    """检查当前最新的ETF数据"""
    today = date.today()
    
    logger.info('=' * 80)
    logger.info('当前最新的ETF数据检查')
    logger.info(f'检查日期: {today}')
    logger.info('=' * 80)
    
    # 获取今天的ETF数据
    etf_data = ETFData.objects.filter(date=today)
    
    if not etf_data.exists():
        logger.warning(f"今天({today})没有ETF数据")
        return
    
    # 获取昨天的数据作为对比
    from datetime import timedelta
    yesterday = today - timedelta(days=1)
    yesterday_data = ETFData.objects.filter(date=yesterday)
    
    logger.info(f"\n当前({today})最新数据:")
    logger.info("-" * 40)
    
    for etf in etf_data.order_by('symbol'):
        # 查找昨天的数据用于对比
        yest_record = yesterday_data.filter(symbol=etf.symbol).first()
        
        if yest_record:
            price_change = etf.close_price - yest_record.close_price
            price_change_pct = (price_change / yest_record.close_price) * 100
            change_str = f"({price_change:+.2f} / {price_change_pct:+.2f}%)"
        else:
            change_str = "(无昨日对比数据)"
        
        logger.info(f"  {etf.symbol}:")
        logger.info(f"    价格: ${etf.close_price:.2f} {change_str}")
        logger.info(f"    开盘: ${etf.open_price:.2f}")
        logger.info(f"    最高: ${etf.high_price:.2f}")
        logger.info(f"    最低: ${etf.low_price:.2f}")
        logger.info(f"    成交量: {etf.volume:,}")
        logger.info(f"    数据来源: {etf.data_source}")
    
    # 显示所有日期的数据记录
    logger.info(f"\n所有ETF数据记录（按日期降序）:")
    logger.info("-" * 40)
    
    all_etf_data = ETFData.objects.order_by('-date', 'symbol')
    unique_dates = set()
    
    for etf in all_etf_data[:20]:  # 显示最近20条
        unique_dates.add(etf.date)
        
        if len(unique_dates) > 3:  # 只显示最近3天的数据
            break
            
        logger.info(f"  {etf.date}: {etf.symbol} = ${etf.close_price:.2f} ({etf.data_source})")
    
    # 检查数据时效性
    latest_date = all_etf_data.first().date if all_etf_data.exists() else None
    if latest_date:
        days_diff = (today - latest_date).days
        logger.info(f"\n数据时效性分析:")
        logger.info(f"  最新数据日期: {latest_date}")
        logger.info(f"  与今天相差: {days_diff} 天")
        
        if days_diff == 0:
            logger.info("  ✅ 数据是今天的，非常新鲜")
        elif days_diff <= 1:
            logger.info("  ✅ 数据是最近1天的，比较新鲜")
        elif days_diff <= 3:
            logger.info("  ⚠️  数据是最近3天的，还可以接受")
        else:
            logger.info(f"  ⚠️  数据是{days_diff}天前的，可能需要更新")
    
    logger.info('=' * 80)
    logger.info('数据检查完成')
    logger.info('=' * 80)

if __name__ == '__main__':
    check_current_data()