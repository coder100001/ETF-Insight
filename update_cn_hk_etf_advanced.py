"""
改进的港股和A股ETF数据更新脚本
- 分别更新每个ETF（避免批量请求限流）
- 增加重试次数和延迟
- 提供详细的进度信息
"""
from workflow.models import ETFData
from workflow.operation_service import operation_service
import yfinance as yf
from datetime import datetime
import time
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)


def update_single_etf(symbol, market_name, retry=3, delay=30):
    """
    更新单个ETF数据

    参数:
        symbol: ETF代码
        market_name: 市场名称（如"港股"、"A股"）
        retry: 重试次数
        delay: 重试延迟（秒）
    """
    print(f"\n{'='*60}")
    print(f"开始更新 {market_name} ETF: {symbol}")
    print(f"{'='*60}")

    for attempt in range(retry):
        try:
            print(f"\n[尝试 {attempt + 1}/{retry}]")

            # 创建工作流实例
            workflow_instance = operation_service.create_workflow_instance(
                workflow_name=f'{market_name}ETF数据更新',
                trigger_type=2,  # 手动
                trigger_by='user'
            )

            print(f"正在下载 {symbol} 的2年历史数据...")

            # 下载历史数据
            ticker = yf.Ticker(symbol)
            data = ticker.history(period='2y', interval='1d')

            if data.empty:
                print(f"警告: {symbol} 没有获取到数据")
                return False

            print(f"成功获取 {len(data)} 条数据记录")
            print(f"日期范围: {data.index[0].date()} 到 {data.index[-1].date()}")

            # 保存到数据库
            print("\n正在保存到数据库...")
            saved_count = 0
            for idx, row in data.iterrows():
                ETFData.objects.update_or_create(
                    symbol=symbol,
                    date=idx.date(),
                    defaults={
                        'open_price': float(row['Open']),
                        'high_price': float(row['High']),
                        'low_price': float(row['Low']),
                        'close_price': float(row['Close']),
                        'volume': int(row['Volume']),
                        'data_source': 'yfinance',
                        'fetch_instance': workflow_instance,
                    }
                )
                saved_count += 1

                # 每100条记录显示一次进度
                if saved_count % 100 == 0:
                    print(f"  已保存 {saved_count} 条记录...")

            print(f"\n✓ 成功保存 {saved_count} 条记录")

            # 获取实时数据
            print("\n正在获取实时数据...")
            info = ticker.info

            if info:
                raw_yield = info.get('dividendYield', 0) or 0
                dividend_yield = raw_yield * 100 if raw_yield < 1 else raw_yield

                print(f"  当前价格: ${info.get('currentPrice', 0):.2f}")
                print(f"  股息率: {dividend_yield:.2f}%")

            # 完成工作流实例
            operation_service.complete_workflow_instance(
                workflow_instance,
                success=True,
                context_data={'saved_count': saved_count}
            )

            print(f"\n✓ {symbol} 数据更新完成！")
            return True

        except Exception as e:
            error_msg = str(e)
            print(f"\n✗ 尝试 {attempt + 1} 失败: {error_msg}")

            # 如果是限流错误且还有重试机会
            if "Rate limited" in error_msg or "Too Many Requests" in error_msg:
                if attempt < retry - 1:
                    wait_time = delay * (attempt + 1)
                    print(f"\n遇到API限流，等待 {wait_time} 秒后重试...")
                    print(f"建议：等待更长时间或使用VPN")
                    time.sleep(wait_time)
                else:
                    print(f"\n已达到最大重试次数，更新失败")
                    print(f"建议：稍后再试或使用手动导入CSV文件的方式")
            else:
                # 其他错误，不再重试
                print(f"\n遇到其他错误，停止重试")
                return False

    return False


def main():
    """主函数"""
    print("="*60)
    print("港股和A股ETF数据更新工具（改进版）")
    print("="*60)

    # 定义要更新的ETF
    etfs_to_update = [
        {
            'symbol': '510300',
            'name': '华泰柏瑞沪深300ETF',
            'market': 'A股'
        },
        {
            'symbol': '3466.HK',
            'name': '南方恒生科技ETF',
            'market': '港股'
        }
    ]

    print("\n准备更新的ETF:")
    for i, etf in enumerate(etfs_to_update, 1):
        print(f"  {i}. {etf['symbol']} - {etf['name']} ({etf['market']})")

    print("\n⚠️  注意:")
    print("  - 由于API限流，会分别更新每个ETF")
    print("  - 每次请求之间会有30秒以上的延迟")
    print("  - 如果仍然失败，请稍后再试或使用VPN")
    print("\n按Ctrl+C可以随时中断...")
    print()

    results = {}

    # 分别更新每个ETF
    for i, etf in enumerate(etfs_to_update, 1):
        print(f"\n进度: {i}/{len(etfs_to_update)}")

        # 如果不是第一个，添加延迟
        if i > 1:
            wait_time = 60
            print(f"\n等待 {wait_time} 秒后再更新下一个ETF...")
            print(f"({datetime.now().strftime('%H:%M:%S')} 等待中...)")
            time.sleep(wait_time)

        # 更新单个ETF
        success = update_single_etf(
            symbol=etf['symbol'],
            market_name=etf['market'],
            retry=3,
            delay=30
        )

        results[etf['symbol']] = {
            'name': etf['name'],
            'market': etf['market'],
            'success': success
        }

    # 显示结果总结
    print("\n" + "="*60)
    print("更新结果总结")
    print("="*60)

    for symbol, result in results.items():
        status = "✓ 成功" if result['success'] else "✗ 失败"
        print(f"  {symbol} - {result['name']} ({result['market']}): {status}")

    print("\n" + "="*60)

    # 检查数据库
    print("\n检查数据库最新数据:")
    print("-"*50)
    for etf in etfs_to_update:
        latest = ETFData.objects.filter(symbol=etf['symbol']).order_by('-date').first()
        if latest:
            print(f"  {etf['symbol']}: {latest.date} - ${latest.close_price:.2f}")
        else:
            print(f"  {etf['symbol']}: 无数据")
    print("-"*50)


if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n用户中断了更新过程")
    except Exception as e:
        print(f"\n\n发生错误: {e}")
        import traceback
        traceback.print_exc()
