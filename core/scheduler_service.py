"""
定时任务服务模块
优化的定时更新逻辑，支持多种调度策略
"""

import logging
from datetime import datetime, time as dt_time
from typing import Optional, List, Callable, Dict, Any
from enum import Enum
import threading
import time

try:
    from apscheduler.schedulers.background import BackgroundScheduler
    from apscheduler.triggers.cron import CronTrigger
    from apscheduler.triggers.interval import IntervalTrigger
    AP_SCHEDULER_AVAILABLE = True
except ImportError:
    AP_SCHEDULER_AVAILABLE = False

from .config import SCHEDULER_CONFIG, DEFAULT_ETF_CONFIGS
from .data_fetcher import get_fetcher
from .data_storage import get_storage

logger = logging.getLogger(__name__)


class UpdateStatus(Enum):
    """更新状态"""
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    SKIPPED = "skipped"


class ETFUpdateTask:
    """ETF 更新任务"""
    
    def __init__(self, symbol: str):
        self.symbol = symbol
        self.status = UpdateStatus.PENDING
        self.start_time: Optional[datetime] = None
        self.end_time: Optional[datetime] = None
        self.records_updated = 0
        self.error_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            'symbol': self.symbol,
            'status': self.status.value,
            'start_time': self.start_time.isoformat() if self.start_time else None,
            'end_time': self.end_time.isoformat() if self.end_time else None,
            'duration_seconds': (
                (self.end_time - self.start_time).total_seconds()
                if self.start_time and self.end_time else None
            ),
            'records_updated': self.records_updated,
            'error_message': self.error_message,
        }


class ETFUpdateScheduler:
    """
    ETF 数据更新调度器
    
    特性：
    1. 支持 APScheduler 和简单定时器两种模式
    2. 增量更新策略
    3. 详细的任务状态跟踪
    4. 失败重试机制
    """
    
    def __init__(self):
        self.scheduler = None
        self.is_running = False
        self.tasks: Dict[str, ETFUpdateTask] = {}
        self.update_history: List[Dict] = []
        self.max_history = 100
        self._lock = threading.Lock()
        
        # 初始化 APScheduler
        if AP_SCHEDULER_AVAILABLE:
            self.scheduler = BackgroundScheduler()
    
    def start(self):
        """启动调度器"""
        if self.is_running:
            logger.warning("调度器已在运行")
            return
        
        if self.scheduler:
            self.scheduler.start()
            logger.info("APScheduler 已启动")
        
        self.is_running = True
    
    def stop(self):
        """停止调度器"""
        if not self.is_running:
            return
        
        if self.scheduler:
            self.scheduler.shutdown()
            logger.info("APScheduler 已停止")
        
        self.is_running = False
    
    def schedule_daily_update(
        self,
        hour: int = 9,
        minute: int = 30,
        symbols: Optional[List[str]] = None,
        job_id: str = 'daily_etf_update'
    ):
        """
        设置每日定时更新
        
        Args:
            hour: 小时 (0-23)
            minute: 分钟 (0-59)
            symbols: 要更新的 ETF 列表，None 表示全部
            job_id: 任务 ID
        """
        if not self.scheduler:
            logger.error("APScheduler 不可用")
            return
        
        def job_func():
            self.run_update(symbols)
        
        # 移除已存在的任务
        try:
            self.scheduler.remove_job(job_id)
        except:
            pass
        
        # 添加新任务
        trigger = CronTrigger(hour=hour, minute=minute)
        self.scheduler.add_job(
            func=job_func,
            trigger=trigger,
            id=job_id,
            name=f'每日ETF更新 ({hour:02d}:{minute:02d})',
            replace_existing=True,
            max_instances=1,  # 防止任务重叠
        )
        
        logger.info(f"已设置每日定时更新: {hour:02d}:{minute:02d}")
    
    def schedule_market_close_update(
        self,
        symbols: Optional[List[str]] = None,
        timezone: str = 'America/New_York'
    ):
        """
        设置美股收盘后更新
        美股收盘时间：16:00 (EST/EDT)
        """
        # 收盘后 30 分钟更新
        self.schedule_daily_update(
            hour=16,
            minute=30,
            symbols=symbols,
            job_id='market_close_update'
        )
        logger.info("已设置美股收盘后自动更新 (16:30 ET)")
    
    def schedule_interval_update(
        self,
        minutes: int = 60,
        symbols: Optional[List[str]] = None,
        job_id: str = 'interval_etf_update'
    ):
        """
        设置间隔更新
        
        Args:
            minutes: 间隔分钟数
            symbols: 要更新的 ETF 列表
            job_id: 任务 ID
        """
        if not self.scheduler:
            logger.error("APScheduler 不可用")
            return
        
        def job_func():
            self.run_update(symbols)
        
        try:
            self.scheduler.remove_job(job_id)
        except:
            pass
        
        trigger = IntervalTrigger(minutes=minutes)
        self.scheduler.add_job(
            func=job_func,
            trigger=trigger,
            id=job_id,
            name=f'定时ETF更新 ({minutes}分钟)',
            replace_existing=True,
            max_instances=1,
        )
        
        logger.info(f"已设置间隔更新: 每 {minutes} 分钟")
    
    def run_update(
        self,
        symbols: Optional[List[str]] = None,
        incremental: bool = True,
        progress_callback: Optional[Callable[[str, UpdateStatus], None]] = None
    ) -> Dict[str, Any]:
        """
        执行 ETF 数据更新
        
        Args:
            symbols: 要更新的 ETF 列表，None 表示全部
            incremental: 是否使用增量更新
            progress_callback: 进度回调函数 (symbol, status)
        
        Returns:
            更新结果统计
        """
        if symbols is None:
            symbols = [cfg.symbol for cfg in DEFAULT_ETF_CONFIGS]
        
        logger.info(f"开始更新 {len(symbols)} 个 ETF 数据")
        
        fetcher = get_fetcher()
        storage = get_storage()
        
        start_time = datetime.now()
        results = []
        
        for symbol in symbols:
            task = ETFUpdateTask(symbol)
            self.tasks[symbol] = task
            
            try:
                task.start_time = datetime.now()
                task.status = UpdateStatus.RUNNING
                
                if progress_callback:
                    progress_callback(symbol, task.status)
                
                # 获取最新日期
                if incremental:
                    latest_date = storage.get_latest_date(symbol)
                    data = fetcher.fetch_incremental_data(symbol, latest_date)
                else:
                    data = fetcher.fetch_historical_data(symbol, period='1y')
                
                if data is None:
                    task.status = UpdateStatus.FAILED
                    task.error_message = "获取数据失败"
                elif data.empty:
                    task.status = UpdateStatus.SKIPPED
                    task.error_message = "数据已是最新"
                else:
                    # 保存数据
                    success = storage.save_historical_data(symbol, data)
                    if success:
                        task.status = UpdateStatus.SUCCESS
                        task.records_updated = len(data)
                    else:
                        task.status = UpdateStatus.FAILED
                        task.error_message = "保存数据失败"
                
            except Exception as e:
                task.status = UpdateStatus.FAILED
                task.error_message = str(e)
                logger.error(f"更新 {symbol} 失败: {e}")
            
            finally:
                task.end_time = datetime.now()
                results.append(task.to_dict())
                
                if progress_callback:
                    progress_callback(symbol, task.status)
        
        end_time = datetime.now()
        duration = (end_time - start_time).total_seconds()
        
        # 统计结果
        success_count = sum(1 for r in results if r['status'] == 'success')
        failed_count = sum(1 for r in results if r['status'] == 'failed')
        skipped_count = sum(1 for r in results if r['status'] == 'skipped')
        total_records = sum(r['records_updated'] for r in results)
        
        summary = {
            'start_time': start_time.isoformat(),
            'end_time': end_time.isoformat(),
            'duration_seconds': duration,
            'total_symbols': len(symbols),
            'success_count': success_count,
            'failed_count': failed_count,
            'skipped_count': skipped_count,
            'total_records_updated': total_records,
            'details': results,
        }
        
        # 保存到历史记录
        with self._lock:
            self.update_history.append(summary)
            if len(self.update_history) > self.max_history:
                self.update_history.pop(0)
        
        logger.info(
            f"更新完成: 成功 {success_count}, 失败 {failed_count}, "
            f"跳过 {skipped_count}, 总记录 {total_records}, 耗时 {duration:.1f}秒"
        )
        
        return summary
    
    def run_realtime_update(
        self,
        symbols: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        获取并保存实时数据
        
        Args:
            symbols: ETF 代码列表
        
        Returns:
            实时数据字典
        """
        if symbols is None:
            symbols = [cfg.symbol for cfg in DEFAULT_ETF_CONFIGS]
        
        fetcher = get_fetcher()
        
        logger.info(f"获取 {len(symbols)} 个 ETF 实时数据")
        
        results = fetcher.fetch_multiple_realtime(symbols)
        
        # 统计
        success_count = sum(1 for v in results.values() if v is not None)
        
        return {
            'timestamp': datetime.now().isoformat(),
            'total': len(symbols),
            'success': success_count,
            'failed': len(symbols) - success_count,
            'data': results,
        }
    
    def get_update_history(self, limit: int = 10) -> List[Dict]:
        """获取更新历史"""
        with self._lock:
            return self.update_history[-limit:]
    
    def get_scheduled_jobs(self) -> List[Dict]:
        """获取已计划的任务列表"""
        if not self.scheduler:
            return []
        
        jobs = []
        for job in self.scheduler.get_jobs():
            jobs.append({
                'id': job.id,
                'name': job.name,
                'next_run_time': job.next_run_time.isoformat() if job.next_run_time else None,
                'trigger': str(job.trigger),
            })
        return jobs
    
    def remove_job(self, job_id: str):
        """移除指定任务"""
        if self.scheduler:
            try:
                self.scheduler.remove_job(job_id)
                logger.info(f"已移除任务: {job_id}")
            except Exception as e:
                logger.error(f"移除任务失败: {e}")


# 全局单例
_scheduler = None


def get_scheduler() -> ETFUpdateScheduler:
    """获取 ETFUpdateScheduler 单例"""
    global _scheduler
    if _scheduler is None:
        _scheduler = ETFUpdateScheduler()
    return _scheduler


def quick_update(symbols: Optional[List[str]] = None) -> Dict[str, Any]:
    """
    快速更新函数（用于命令行或脚本调用）
    
    Args:
        symbols: ETF 代码列表，None 表示全部
    
    Returns:
        更新结果
    """
    scheduler = get_scheduler()
    return scheduler.run_update(symbols=symbols)


def setup_default_schedule():
    """设置默认的定时更新计划"""
    scheduler = get_scheduler()
    scheduler.start()
    
    # 美股开盘前更新
    scheduler.schedule_daily_update(hour=9, minute=30, job_id='pre_market_update')
    
    # 美股收盘后更新
    scheduler.schedule_daily_update(hour=16, minute=30, job_id='post_market_update')
    
    logger.info("默认定时计划已设置")
