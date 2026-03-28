"""
使用AKShare拉取港股ETF数据（备份方案）
AKShare是一个优秀的中文金融数据接口库，支持港股市场

安装:
    pip install akshare

用法:
    python fetch_hk_etf_akshare.py 3466.HK
    python fetch_hk_etf_akshare.py 3466.HK 3000.HK
"""

import akshare as ak
import pandas as pd
from workflow.models import ETFData
from workflow.operation_service import operation_service
from datetime import datetime, timedelta
import time
import sys


def fetch_hk_etf_akshare(symbol, retry=3):
    """
    使用AKShare拉取港股ETF数据

    参数:
        symbol: 港股代码（如 3466.HK）
        retry: 重试次数

    返回:
        DataFrame: 包含历史数据
    """
    # 提取股票代码（去掉.HK后缀）
    if '.HK' in symbol:
        stock_code = symbol.replace('.HK', '')
    else:
        stock_code = symbol

    print(f"\n使用AKShare拉取港股ETF: {symbol}")
    print(f"股票代码: {stock_code}")

    for attempt in range(retry):
        try:
            print(f"\n[尝试 {attempt + 1}/{retry}]")

            # AKShare港股历史数据接口
            # stock_hk_hist: 港股历史行情数据
            data = ak.stock_hk_hist(
                symbol=stock_code,
                period="daily",
                start_date="20230101",
                end_date=datetime.now().strftime('%Y%m%d'),
                adjust="qfq"  # 前复权
            )

            if data is None or data.empty:
                print(f"警告: {symbol} 没有获取到数据")
                return None

            print(f"成功获取 {len(data)} 条数据记录")
            print(f"日期范围: {data.iloc[0]['日期']} 到 {data.iloc[-1]['日期']}")

            # 显示数据列名
            print(f"数据列: {list(data.columns)}")
            print(f"\n前5条数据:")
            print(data.head())

            return data

        except Exception as e:
            error_msg = str(e)
            print(f"✗ 尝试 {attempt + 1} 失败: {error_msg}")

            if attempt < retry - 1:
                wait_time = 5
                print(f"等待 {wait_time} 秒后重试...")
                time.sleep(wait_time)
            else:
                print(f"\n已达到最大重试次数")
                return None

    return None


def save_akshare_data_to_db(data, symbol):
    """
    将AKShare数据保存到数据库

    参数:
        data: AKShare获取的DataFrame
        symbol: ETF代码

    返回:
        (成功, 保存记录数)
    """
    if data is None or data.empty:
        return False, 0

    try:
        # 创建工作流实例
        workflow_instance = operation_service.create_workflow_instance(
            workflow_name='港股ETF数据更新(AKShare)',
            trigger_type=2,
            trigger_by='user'
        )

        print("\n正在保存到数据库...")

        # 映射AKShare的列名到数据库字段
        # AKShare列: 日期, 开盘, 收盘, 最高, 最低, 成交量
        saved_count = 0
        for idx in range(len(data)):
            row = data.iloc[idx]

            # 解析日期
            date_obj = pd.to_datetime(row['日期']).date()

            # 保存到数据库
            ETFData.objects.update_or_create(
                symbol=symbol,
                date=date_obj,
                defaults={
                    'open_price': float(row['开盘']),
                    'high_price': float(row['最高']),
                    'low_price': float(row['最低']),
                    'close_price': float(row['收盘']),
                    'volume': int(row['成交量']),
                    'data_source': 'akshare',
                    'fetch_instance': workflow_instance,
                }
            )
            saved_count += 1

            if saved_count % 100 == 0:
                print(f"  已保存 {saved_count} 条记录...")

        # 完成工作流实例
        operation_service.complete_workflow_instance(
            workflow_instance,
            success=True,
            context_data={'saved_count': saved_count}
        )

        print(f"\n✓ 成功保存 {saved_count} 条记录")
        return True, saved_count

    except Exception as e:
        print(f"\n✗ 保存到数据库失败: {str(e)}")
        import traceback
        traceback.print_exc()
        return False, 0


def main():
    """主函数"""
    print("="*60)
    print("港股ETF数据拉取工具 - AKShare备份方案")
    print("="*60)

    # 检查命令行参数
    if len(sys.argv) < 2:
        print("\n用法: python fetch_hk_etf_akshare.py <港股代码>")
        print("\n示例:")
        print("  python fetch_hk_etf_akshare.py 3466.HK")
        print("  python fetch_hk_etf_akshare.py 3466.HK 3000.HK 7000.HK")
        print("\n说明:")
        print("  - AKShare是开源的中文金融数据接口")
        print("  - 支持港股、A股等多个市场")
        print("  - 不受Yahoo Finance API限流影响")
        print("="*60)
        return

    # 获取要更新的ETF列表
    symbols = [s.upper() for s in sys.argv[1:]]

    print(f"\n准备更新的ETF: {', '.join(symbols)}")

    # 检查是否安装了akshare
    try:
        import akshare
        print(f"\nAKShare版本: {ak.__version__}")
    except ImportError:
        print("\n❌ 未安装AKShare库！")
        print("\n请先安装:")
        print("  pip install akshare")
        print("\n或使用以下命令安装:")
        print("  pip install akshare -i https://pypi.tuna.tsinghua.edu.cn/simple")
        print("="*60)
        return

    print("\n注意:")
    print("  - 使用AKShare作为数据源")
    print("  - 数据来源: 新浪财经")
    print("  - 优势: 不受Yahoo Finance限流影响")

    # 分别更新每个ETF
    results = {}
    for i, symbol in enumerate(symbols, 1):
        print(f"\n{'='*60}")
        print(f"进度: {i}/{len(symbols)} - {symbol}")
        print(f"{'='*60}")

        # 拉取数据
        data = fetch_hk_etf_akshare(symbol, retry=3)

        if data is not None:
            # 保存到数据库
            success, count = save_akshare_data_to_db(data, symbol)
            results[symbol] = {
                'success': success,
                'count': count
            }
        else:
            results[symbol] = {
                'success': False,
                'count': 0
            }

        # 如果不是最后一个，添加延迟
        if i < len(symbols):
            print(f"\n等待5秒后更新下一个ETF...")
            time.sleep(5)

    # 显示结果总结
    print("\n" + "="*60)
    print("更新结果总结")
    print("="*60)

    for symbol, result in results.items():
        status = "✓ 成功" if result['success'] else "✗ 失败"
        count_info = f"({result['count']} 条记录)" if result['success'] else ""
        print(f"  {symbol}: {status} {count_info}")

    print("\n" + "="*60)

    # 检查数据库
    print("\n检查数据库最新数据:")
    print("-"*50)
    for symbol in symbols:
        latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
        if latest:
            print(f"  {symbol}: {latest.date} - HKD{latest.close_price:.2f}")
        else:
            print(f"  {symbol}: 无数据")
    print("-"*50)

    print("\n完成！")


if __name__ == '__main__':
    main()
