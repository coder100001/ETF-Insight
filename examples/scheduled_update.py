#!/usr/bin/env python
"""
定时更新示例
展示如何设置自动定时更新 ETF 数据
"""

import os
import sys
import time
import logging

# 添加项目路径
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# 设置 Django 环境
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
import django
django.setup()

from core import get_scheduler, quick_update

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)


def on_update_progress(symbol, status):
    """更新进度回调"""
    status_icons = {
        'running': '⏳',
        'success': '✅',
        'failed': '❌',
        'skipped': '⏭️',
    }
    icon = status_icons.get(status.value, '❓')
    logger.info(f"{icon} {symbol}: {status.value}")


def run_scheduled_update():
    """
    运行定时更新服务

    功能:
    1. 美股开盘前更新 (9:30 ET)
    2. 美股收盘后更新 (16:30 ET)
    3. 每小时检查一次实时数据
    """
    scheduler = get_scheduler()

    logger.info("=" * 60)
    logger.info("启动 ETF 定时更新服务")
    logger.info("=" * 60)

    # 启动调度器
    scheduler.start()

    # 设置定时任务
    # 1. 美股开盘前更新
    scheduler.schedule_daily_update(
        hour=9, minute=30,
        job_id='pre_market_update'
    )
    logger.info("✓ 已设置开盘前更新 (9:30 ET)")

    # 2. 美股收盘后更新
    scheduler.schedule_market_close_update()
    logger.info("✓ 已设置收盘后更新 (16:30 ET)")

    # 3. 每小时检查实时数据
    scheduler.schedule_interval_update(
        minutes=60,
        job_id='hourly_check'
    )
    logger.info("✓ 已设置每小时检查")

    # 显示所有任务
    jobs = scheduler.get_scheduled_jobs()
    logger.info(f"\n已配置 {len(jobs)} 个定时任务:")
    for job in jobs:
        logger.info(f"  - {job['name']}")
        logger.info(f"    下次执行: {job['next_run_time']}")

    logger.info("\n" + "=" * 60)
    logger.info("服务运行中，按 Ctrl+C 停止")
    logger.info("=" * 60)

    try:
        # 保持运行
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logger.info("\n正在停止服务...")
        scheduler.stop()
        logger.info("服务已停止")


def run_once():
    """立即执行一次更新"""
    logger.info("执行单次更新...")

    result = quick_update()

    logger.info(f"\n更新完成!")
    logger.info(f"成功: {result['success_count']}/{result['total_symbols']}")
    logger.info(f"更新记录: {result['total_records_updated']} 条")
    logger.info(f"耗时: {result['duration_seconds']:.2f} 秒")


if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='ETF 定时更新服务')
    parser.add_argument(
        '--once',
        action='store_true',
        help='立即执行一次更新，不启动定时服务'
    )

    args = parser.parse_args()

    if args.once:
        run_once()
    else:
        run_scheduled_update()
