#!/usr/bin/env python
"""
基础使用示例
展示 core 模块的基本用法
"""

import os
import sys

# 添加项目路径
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# 设置 Django 环境
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'etf_workflow_project.settings')
import django
django.setup()

from core import get_fetcher, get_storage, get_scheduler, quick_update


def example_1_fetch_single():
    """示例1: 获取单个 ETF 数据"""
    print("=" * 60)
    print("示例1: 获取单个 ETF 历史数据")
    print("=" * 60)

    fetcher = get_fetcher()

    # 获取 SCHD 近一年数据
    data = fetcher.fetch_historical_data('SCHD', period='1y')

    if data is not None:
        print(f"获取到 {len(data)} 条记录")
        print(f"日期范围: {data.index[0]} 到 {data.index[-1]}")
        print("\n最近5条数据:")
        print(data.tail())
    else:
        print("获取数据失败")


def example_2_fetch_multiple():
    """示例2: 并发获取多个 ETF 数据"""
    print("\n" + "=" * 60)
    print("示例2: 并发获取多个 ETF 数据")
    print("=" * 60)

    fetcher = get_fetcher()

    symbols = ['SCHD', 'SPYD', 'VYM']

    def progress_callback(completed, total):
        print(f"进度: {completed}/{total}")

    results = fetcher.fetch_multiple_symbols(
        symbols,
        period='6mo',
        progress_callback=progress_callback
    )

    for symbol, data in results.items():
        if data is not None:
            print(f"{symbol}: {len(data)} 条记录")
        else:
            print(f"{symbol}: 获取失败")


def example_3_incremental_update():
    """示例3: 增量更新"""
    print("\n" + "=" * 60)
    print("示例3: 增量更新数据")
    print("=" * 60)

    fetcher = get_fetcher()
    storage = get_storage()

    symbol = 'SCHD'

    # 获取数据库中最新日期
    latest_date = storage.get_latest_date(symbol)
    print(f"数据库最新日期: {latest_date}")

    # 获取增量数据
    new_data = fetcher.fetch_incremental_data(symbol, latest_date)

    if new_data is not None:
        if new_data.empty:
            print("数据已是最新，无需更新")
        else:
            print(f"获取到 {len(new_data)} 条新记录")
            print("新数据:")
            print(new_data)

            # 保存到数据库
            success = storage.save_historical_data(symbol, new_data)
            print(f"保存结果: {'成功' if success else '失败'}")


def example_4_quick_update():
    """示例4: 快速更新所有 ETF"""
    print("\n" + "=" * 60)
    print("示例4: 快速更新所有 ETF")
    print("=" * 60)

    result = quick_update()

    print(f"\n更新完成!")
    print(f"总计: {result['total_symbols']} 个 ETF")
    print(f"成功: {result['success_count']} 个")
    print(f"失败: {result['failed_count']} 个")
    print(f"跳过: {result['skipped_count']} 个")
    print(f"更新记录: {result['total_records_updated']} 条")
    print(f"耗时: {result['duration_seconds']:.2f} 秒")


def example_5_realtime_data():
    """示例5: 获取实时数据"""
    print("\n" + "=" * 60)
    print("示例5: 获取实时数据")
    print("=" * 60)

    fetcher = get_fetcher()

    symbols = ['SCHD', 'SPYD']
    results = fetcher.fetch_multiple_realtime(symbols)

    for symbol, data in results.items():
        if data:
            print(f"\n{data['name']} ({symbol}):")
            print(f"  当前价格: ${data['current_price']}")
            print(f"  涨跌: {data['change']:.2f} ({data['change_percent']:.2f}%)")
            print(f"  成交量: {data['volume']:,}")
        else:
            print(f"{symbol}: 获取失败")


def example_6_scheduler():
    """示例6: 设置定时更新"""
    print("\n" + "=" * 60)
    print("示例6: 设置定时更新")
    print("=" * 60)

    scheduler = get_scheduler()

    # 启动调度器
    scheduler.start()
    print("调度器已启动")

    # 设置每日 9:30 更新
    scheduler.schedule_daily_update(hour=9, minute=30)
    print("已设置每日 9:30 更新")

    # 设置收盘后更新
    scheduler.schedule_market_close_update()
    print("已设置收盘后更新 (16:30 ET)")

    # 查看已计划的任务
    jobs = scheduler.get_scheduled_jobs()
    print(f"\n已计划的任务 ({len(jobs)} 个):")
    for job in jobs:
        print(f"  - {job['name']}")
        print(f"    下次执行: {job['next_run_time']}")

    # 停止调度器（实际使用时不要立即停止）
    scheduler.stop()
    print("\n调度器已停止")


if __name__ == '__main__':
    print("ETF 数据管理 - 基础使用示例")
    print("=" * 60)

    # 运行示例
    try:
        example_1_fetch_single()
    except Exception as e:
        print(f"示例1失败: {e}")

    try:
        example_2_fetch_multiple()
    except Exception as e:
        print(f"示例2失败: {e}")

    try:
        example_3_incremental_update()
    except Exception as e:
        print(f"示例3失败: {e}")

    try:
        example_4_quick_update()
    except Exception as e:
        print(f"示例4失败: {e}")

    try:
        example_5_realtime_data()
    except Exception as e:
        print(f"示例5失败: {e}")

    try:
        example_6_scheduler()
    except Exception as e:
        print(f"示例6失败: {e}")

    print("\n" + "=" * 60)
    print("所有示例运行完成")
