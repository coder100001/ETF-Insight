"""
从CSV文件导入ETF数据到数据库
用于手动导入港股、A股等ETF数据
"""
import pandas as pd
from workflow.models import ETFData
from workflow.operation_service import operation_service
import os


def import_etf_from_csv(csv_file, symbol):
    """
    从CSV文件导入ETF数据

    参数:
        csv_file: CSV文件路径
        symbol: ETF代码

    返回:
        (成功, 导入记录数, 错误信息)
    """
    if not os.path.exists(csv_file):
        return False, 0, f"文件不存在: {csv_file}"

    print(f"正在读取CSV文件: {csv_file}")

    try:
        # 创建工作流实例
        workflow_instance = operation_service.create_workflow_instance(
            workflow_name=f'CSV导入ETF数据',
            trigger_type=2,
            trigger_by='user'
        )

        # 读取CSV文件
        df = pd.read_csv(csv_file)

        print(f"CSV文件包含 {len(df)} 条记录")

        # 检查必要的列
        required_columns = ['Date', 'Open', 'High', 'Low', 'Close', 'Volume']
        missing_columns = [col for col in required_columns if col not in df.columns]

        if missing_columns:
            return False, 0, f"CSV文件缺少必要的列: {missing_columns}"

        # 导入数据
        saved_count = 0
        for idx, row in df.iterrows():
            try:
                # 解析日期（支持多种格式）
                date_str = row['Date']
                if isinstance(date_str, str):
                    # 尝试不同的日期格式
                    for fmt in ['%Y-%m-%d', '%Y-%m-%d %H:%M:%S', '%Y/%m/%d', '%Y/%m/%d %H:%M:%S']:
                        try:
                            date_obj = pd.to_datetime(date_str, format=fmt).date()
                            break
                        except:
                            continue
                    else:
                        # 如果所有格式都不匹配，使用pandas自动解析
                        date_obj = pd.to_datetime(date_str).date()
                else:
                    date_obj = pd.to_datetime(date_str).date()

                # 保存到数据库
                ETFData.objects.update_or_create(
                    symbol=symbol,
                    date=date_obj,
                    defaults={
                        'open_price': float(row['Open']),
                        'high_price': float(row['High']),
                        'low_price': float(row['Low']),
                        'close_price': float(row['Close']),
                        'volume': int(row['Volume']),
                        'data_source': 'csv_import',
                        'fetch_instance': workflow_instance,
                    }
                )
                saved_count += 1

                if saved_count % 100 == 0:
                    print(f"  已导入 {saved_count} 条记录...")

            except Exception as e:
                print(f"  警告: 第{idx}行导入失败: {e}")
                continue

        # 完成工作流实例
        operation_service.complete_workflow_instance(
            workflow_instance,
            success=True,
            context_data={'saved_count': saved_count, 'csv_file': csv_file}
        )

        print(f"\n✓ 成功导入 {saved_count} 条记录")
        return True, saved_count, ""

    except Exception as e:
        error_msg = f"导入失败: {str(e)}"
        print(f"\n✗ {error_msg}")
        import traceback
        traceback.print_exc()
        return False, 0, error_msg


def main():
    """主函数"""
    print("="*60)
    print("ETF数据CSV导入工具")
    print("="*60)

    # 显示当前数据库中的ETF数据统计
    print("\n当前数据库ETF数据统计:")
    print("-"*50)
    from workflow.models import ETFData
    stats = ETFData.objects.values('symbol').distinct()
    existing_symbols = [s['symbol'] for s in stats]
    for symbol in sorted(existing_symbols):
        count = ETFData.objects.filter(symbol=symbol).count()
        print(f"  {symbol}: {count} 条记录")
    print("-"*50)

    print("\n用法:")
    print("  python import_etf_csv.py <CSV文件路径> <ETF代码>")
    print("\n示例:")
    print("  python import_etf_csv.py 3466.HK_data.csv 3466.HK")
    print("  python import_etf_csv.py 510300_data.csv 510300")
    print()

    # 检查命令行参数
    import sys
    if len(sys.argv) < 3:
        print("\n未提供参数，演示导入功能...")
        print("\n提示: 请提供CSV文件路径和ETF代码")
        print("="*60)
        return

    csv_file = sys.argv[1]
    symbol = sys.argv[2].upper()

    print(f"准备导入: {symbol}")
    print(f"CSV文件: {csv_file}")
    print()

    # 执行导入
    success, count, error = import_etf_from_csv(csv_file, symbol)

    if success:
        print(f"\n✓ {symbol} 导入成功！共导入 {count} 条记录")

        # 显示最新数据
        latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
        if latest:
            print(f"最新数据日期: {latest.date}")
            print(f"最新价格: {latest.close_price}")
    else:
        print(f"\n✗ 导入失败: {error}")

    print("="*60)


if __name__ == '__main__':
    main()
