#!/usr/bin/env python
"""
命令行接口
提供便捷的 ETF 数据管理命令
"""

import argparse
import sys
import logging
from datetime import datetime
from typing import List, Optional

from .scheduler_service import get_scheduler, quick_update
from .data_fetcher import get_fetcher
from .data_storage import get_storage
from .config import DEFAULT_ETF_CONFIGS

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)


def print_progress(symbol: str, status):
    """打印进度"""
    status_icons = {
        'running': '⏳',
        'success': '✅',
        'failed': '❌',
        'skipped': '⏭️',
    }
    icon = status_icons.get(status.value, '❓')
    print(f"{icon} {symbol}: {status.value}")


def cmd_update(args):
    """更新命令"""
    symbols = args.symbols if args.symbols else None
    
    if args.all:
        symbols = None
        print("将更新所有 ETF")
    elif symbols:
        print(f"将更新指定 ETF: {', '.join(symbols)}")
    else:
        # 默认更新所有
        symbols = None
        print("将更新所有 ETF")
    
    # 执行更新
    result = quick_update(symbols=symbols)
    
    # 打印结果
    print("\n" + "=" * 60)
    print("更新结果")
    print("=" * 60)
    print(f"总计: {result['total_symbols']} 个 ETF")
    print(f"成功: {result['success_count']} 个")
    print(f"失败: {result['failed_count']} 个")
    print(f"跳过: {result['skipped_count']} 个")
    print(f"更新记录: {result['total_records_updated']} 条")
    print(f"耗时: {result['duration_seconds']:.2f} 秒")
    
    # 打印详细信息
    if args.verbose:
        print("\n详细信息:")
        for detail in result['details']:
            status_icon = '✅' if detail['status'] == 'success' else '❌' if detail['status'] == 'failed' else '⏭️'
            print(f"  {status_icon} {detail['symbol']}: {detail['status']}")
            if detail['error_message']:
                print(f"     错误: {detail['error_message']}")
            if detail['records_updated'] > 0:
                print(f"     更新: {detail['records_updated']} 条记录")
    
    # 返回码
    return 0 if result['failed_count'] == 0 else 1


def cmd_fetch(args):
    """获取数据命令（不保存到数据库）"""
    fetcher = get_fetcher()
    
    symbols = args.symbols
    period = args.period
    
    print(f"获取 {', '.join(symbols)} 的 {period} 数据...")
    
    for symbol in symbols:
        print(f"\n{'='*60}")
        print(f"ETF: {symbol}")
        print('='*60)
        
        data = fetcher.fetch_historical_data(symbol, period=period)
        
        if data is not None:
            print(f"记录数: {len(data)}")
            print(f"日期范围: {data.index[0]} 到 {data.index[-1]}")
            print("\n最近5条数据:")
            print(data.tail())
            print("\n统计信息:")
            print(data.describe())
        else:
            print("获取数据失败")
    
    return 0


def cmd_realtime(args):
    """获取实时数据命令"""
    fetcher = get_fetcher()
    
    symbols = args.symbols if args.symbols else [cfg.symbol for cfg in DEFAULT_ETF_CONFIGS]
    
    print(f"获取 {', '.join(symbols)} 的实时数据...\n")
    
    results = fetcher.fetch_multiple_realtime(symbols)
    
    for symbol, data in results.items():
        if data:
            print(f"{'='*60}")
            print(f"{data.get('name', symbol)} ({symbol})")
            print('='*60)
            print(f"当前价格: ${data.get('current_price', 'N/A')}")
            print(f"涨跌: {data.get('change', 0):.2f} ({data.get('change_percent', 0):.2f}%)")
            print(f"开盘价: ${data.get('open_price', 'N/A')}")
            print(f"最高: ${data.get('day_high', 'N/A')}")
            print(f"最低: ${data.get('day_low', 'N/A')}")
            print(f"成交量: {data.get('volume', 'N/A'):,}")
            print(f"股息率: {data.get('dividend_yield', 'N/A')}%")
            print()
        else:
            print(f"{symbol}: 获取失败\n")
    
    return 0


def cmd_schedule(args):
    """定时任务命令"""
    scheduler = get_scheduler()
    
    if args.action == 'start':
        scheduler.start()
        
        # 设置定时任务
        if args.daily:
            hour, minute = map(int, args.daily.split(':'))
            scheduler.schedule_daily_update(hour=hour, minute=minute)
        
        if args.market_close:
            scheduler.schedule_market_close_update()
        
        if args.interval:
            scheduler.schedule_interval_update(minutes=args.interval)
        
        # 如果没有指定任何任务，使用默认设置
        if not any([args.daily, args.market_close, args.interval]):
            scheduler.schedule_market_close_update()
            print("已设置默认收盘后更新任务 (16:30 ET)")
        
        print("调度器已启动，按 Ctrl+C 停止")
        
        try:
            while True:
                import time
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n停止调度器...")
            scheduler.stop()
    
    elif args.action == 'list':
        jobs = scheduler.get_scheduled_jobs()
        if jobs:
            print("已计划的定时任务:")
            for job in jobs:
                print(f"  - {job['name']} (下次执行: {job['next_run_time']})")
        else:
            print("没有已计划的定时任务")
    
    elif args.action == 'stop':
        scheduler.stop()
        print("调度器已停止")
    
    return 0


def cmd_status(args):
    """查看状态命令"""
    scheduler = get_scheduler()
    storage = get_storage()
    
    print("=" * 60)
    print("ETF 数据状态")
    print("=" * 60)
    
    # 检查每个 ETF 的最新数据
    print("\n各 ETF 最新数据日期:")
    for cfg in DEFAULT_ETF_CONFIGS:
        latest = storage.get_latest_date(cfg.symbol)
        if latest:
            days_ago = (datetime.now().date() - latest).days
            status = "✅ 最新" if days_ago <= 1 else f"⚠️ {days_ago}天前"
            print(f"  {cfg.symbol}: {latest} {status}")
        else:
            print(f"  {cfg.symbol}: ❌ 无数据")
    
    # 更新历史
    print("\n最近更新历史:")
    history = scheduler.get_update_history(limit=5)
    if history:
        for record in history:
            print(f"  {record['start_time']}: "
                  f"成功{record['success_count']}/"
                  f"失败{record['failed_count']}/"
                  f"跳过{record['skipped_count']} "
                  f"({record['duration_seconds']:.1f}s)")
    else:
        print("  暂无更新记录")
    
    return 0


def cmd_export(args):
    """导出数据命令"""
    storage = get_storage()
    
    symbols = args.symbols if args.symbols else [cfg.symbol for cfg in DEFAULT_ETF_CONFIGS]
    
    print(f"导出 {', '.join(symbols)} 数据到 Excel...")
    
    filepath = storage.export_to_excel(symbols, args.output)
    
    print(f"导出完成: {filepath}")
    
    return 0


def main(argv: Optional[List[str]] = None) -> int:
    """主入口"""
    parser = argparse.ArgumentParser(
        prog='etf',
        description='ETF 数据管理工具',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  etf update                    # 更新所有 ETF 数据
  etf update SCHD SPYD          # 更新指定 ETF
  etf fetch SCHD --period 1y    # 获取历史数据
  etf realtime                  # 获取实时数据
  etf schedule start            # 启动定时更新
  etf status                    # 查看数据状态
  etf export --output data.xlsx # 导出数据
        """
    )
    
    subparsers = parser.add_subparsers(dest='command', help='可用命令')
    
    # update 命令
    update_parser = subparsers.add_parser('update', help='更新 ETF 数据')
    update_parser.add_argument('symbols', nargs='*', help='ETF 代码（可选）')
    update_parser.add_argument('--all', action='store_true', help='更新所有 ETF')
    update_parser.add_argument('-v', '--verbose', action='store_true', help='详细输出')
    
    # fetch 命令
    fetch_parser = subparsers.add_parser('fetch', help='获取历史数据（不保存）')
    fetch_parser.add_argument('symbols', nargs='+', help='ETF 代码')
    fetch_parser.add_argument('--period', default='1y', help='时间周期 (1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y)')
    
    # realtime 命令
    realtime_parser = subparsers.add_parser('realtime', help='获取实时数据')
    realtime_parser.add_argument('symbols', nargs='*', help='ETF 代码（可选）')
    
    # schedule 命令
    schedule_parser = subparsers.add_parser('schedule', help='定时任务管理')
    schedule_parser.add_argument('action', choices=['start', 'stop', 'list'], help='操作')
    schedule_parser.add_argument('--daily', metavar='HH:MM', help='每日更新时间')
    schedule_parser.add_argument('--market-close', action='store_true', help='收盘后更新')
    schedule_parser.add_argument('--interval', type=int, metavar='MINUTES', help='间隔分钟数')
    
    # status 命令
    subparsers.add_parser('status', help='查看数据状态')
    
    # export 命令
    export_parser = subparsers.add_parser('export', help='导出数据')
    export_parser.add_argument('symbols', nargs='*', help='ETF 代码（可选，默认全部）')
    export_parser.add_argument('-o', '--output', help='输出文件名')
    
    args = parser.parse_args(argv)
    
    if not args.command:
        parser.print_help()
        return 1
    
    # 路由到对应命令
    commands = {
        'update': cmd_update,
        'fetch': cmd_fetch,
        'realtime': cmd_realtime,
        'schedule': cmd_schedule,
        'status': cmd_status,
        'export': cmd_export,
    }
    
    return commands[args.command](args)


if __name__ == '__main__':
    sys.exit(main())
