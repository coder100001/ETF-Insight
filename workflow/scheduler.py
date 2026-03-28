"""
定时任务服务
使用APScheduler实现定时任务

集成优化后的 core 模块，提供 ETF 数据定时更新功能
"""

from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
import logging
from django.utils import timezone

logger = logging.getLogger(__name__)

scheduler = None


def get_scheduler():
    """获取调度器实例（单例模式）"""
    global scheduler
    if scheduler is None:
        scheduler = BackgroundScheduler()
        scheduler.start()
        logger.info("定时任务调度器已启动")
    return scheduler


def start_exchange_rate_job():
    """启动汇率更新定时任务"""
    sched = get_scheduler()
    
    # 添加每天10:30执行的任务
    sched.add_job(
        func=update_exchange_rates_job,
        trigger=CronTrigger(hour=10, minute=30),
        id='exchange_rate_update',
        name='汇率更新任务',
        replace_existing=True,
        max_instances=1
    )
    
    logger.info("汇率更新定时任务已配置（每天10:30执行）")


def update_exchange_rates_job():
    """汇率更新任务"""
    from .services_exchange_rate import update_exchange_rates_auto
    from .models import OperationLog

    try:
        logger.info("=" * 60)
        logger.info("开始执行汇率更新任务")
        logger.info("=" * 60)

        # 创建操作记录
        op_log = OperationLog.objects.create(
            operation_type='scheduled_task',
            operation_name='定时汇率更新',
            operator='system',
            status=0  # 进行中
        )

        # 执行更新（从免费API获取真实汇率）
        result = update_exchange_rates_auto()

        # 更新操作记录
        op_log.complete(
            success=True,
            result={'updated_count': result['updated_count']}
        )

        logger.info(f"汇率更新任务执行成功: 更新了 {result['updated_count']} 条记录")
        logger.info("=" * 60)

    except Exception as e:
        logger.error(f"汇率更新任务执行失败: {e}", exc_info=True)

        # 记录失败
        try:
            op_log.complete(
                success=False,
                error=str(e)
            )
        except:
            pass


def update_etf_data(symbols=None):
    """
    更新ETF数据任务
    
    Args:
        symbols: ETF代码列表，None表示更新所有
    
    Returns:
        bool: 是否成功
    """
    from .models import OperationLog
    
    try:
        logger.info("=" * 60)
        logger.info("开始执行ETF数据更新任务")
        logger.info("=" * 60)
        
        # 导入 core 模块的更新功能
        try:
            from core import quick_update
            use_core = True
        except ImportError:
            use_core = False
            logger.warning("core 模块不可用，使用 fallback 更新")
        
        # 创建操作记录
        op_log = OperationLog.objects.create(
            operation_type='scheduled_task',
            operation_name='ETF数据更新',
            operator='system',
            status=0
        )
        
        if use_core:
            # 使用优化的 core 模块
            result = quick_update(symbols=symbols)
            
            success = result['failed_count'] == 0
            message = (
                f"成功: {result['success_count']}, "
                f"失败: {result['failed_count']}, "
                f"跳过: {result['skipped_count']}, "
                f"记录: {result['total_records_updated']}"
            )
        else:
            # Fallback: 使用原有的 services
            from .services import etf_service
            from .models import ETFData
            from datetime import datetime, timedelta
            
            if symbols is None:
                symbols = etf_service.SYMBOLS
            
            success_count = 0
            failed_count = 0
            
            for symbol in symbols:
                try:
                    # 获取历史数据
                    data = etf_service.fetch_historical_data(symbol, '1y')
                    if data is not None and not data.empty:
                        # 保存到数据库
                        for index, row in data.iterrows():
                            ETFData.objects.update_or_create(
                                symbol=symbol,
                                date=index.date(),
                                defaults={
                                    'open_price': float(row.get('Open', 0)),
                                    'high_price': float(row.get('High', 0)),
                                    'low_price': float(row.get('Low', 0)),
                                    'close_price': float(row.get('Close', 0)),
                                    'volume': int(row.get('Volume', 0)),
                                }
                            )
                        success_count += 1
                    else:
                        failed_count += 1
                except Exception as e:
                    logger.error(f"更新 {symbol} 失败: {e}")
                    failed_count += 1
            
            success = failed_count == 0
            message = f"成功: {success_count}, 失败: {failed_count}"
        
        # 更新操作记录
        op_log.complete(
            success=success,
            result={'message': message}
        )
        
        logger.info(f"ETF数据更新完成: {message}")
        logger.info("=" * 60)
        
        return success
        
    except Exception as e:
        logger.error(f"ETF数据更新任务失败: {e}", exc_info=True)
        
        try:
            op_log.complete(success=False, error=str(e))
        except:
            pass
        
        return False


def start_etf_update_job(hour=16, minute=30):
    """
    启动ETF数据定时更新任务
    
    Args:
        hour: 小时 (默认16，美股收盘后)
        minute: 分钟 (默认30)
    """
    sched = get_scheduler()
    
    sched.add_job(
        func=update_etf_data,
        trigger=CronTrigger(hour=hour, minute=minute),
        id='etf_daily_update',
        name=f'ETF每日更新 ({hour:02d}:{minute:02d})',
        replace_existing=True,
        max_instances=1
    )
    
    logger.info(f"ETF数据定时任务已配置（每天{hour:02d}:{minute:02d}执行）")


def start_pre_market_etf_job(hour=9, minute=30):
    """
    启动盘前ETF数据更新任务
    
    Args:
        hour: 小时 (默认9，美股开盘前)
        minute: 分钟 (默认30)
    """
    sched = get_scheduler()
    
    sched.add_job(
        func=update_etf_data,
        trigger=CronTrigger(hour=hour, minute=minute),
        id='etf_pre_market_update',
        name=f'ETF盘前更新 ({hour:02d}:{minute:02d})',
        replace_existing=True,
        max_instances=1
    )
    
    logger.info(f"ETF盘前定时任务已配置（每天{hour:02d}:{minute:02d}执行）")


def start_scheduler():
    """启动调度器并注册所有定时任务"""
    try:
        sched = get_scheduler()
        
        # 启动汇率更新定时任务
        start_exchange_rate_job()
        
        # 启动ETF数据更新任务（收盘后）
        start_etf_update_job(hour=16, minute=30)
        
        # 启动ETF盘前更新任务
        start_pre_market_etf_job(hour=9, minute=30)
        
        logger.info("所有定时任务已启动")
        
    except Exception as e:
        logger.error(f"启动调度器失败: {e}", exc_info=True)


def stop_scheduler():
    """停止调度器"""
    global scheduler
    if scheduler:
        scheduler.shutdown()
        scheduler = None
        logger.info("定时任务调度器已停止")


def list_jobs():
    """列出所有定时任务"""
    sched = get_scheduler()
    jobs = sched.get_jobs()
    
    job_list = []
    for job in jobs:
        job_list.append({
            'id': job.id,
            'name': job.name,
            'next_run_time': job.next_run_time,
            'trigger': str(job.trigger)
        })
    
    return job_list
