"""
尝试更新单个A股ETF（510300）
"""
from workflow.models import ETFData
from workflow.operation_service import operation_service
import yfinance as yf
from datetime import datetime
import time

print("尝试更新A股ETF: 510300")
print("等待10秒避免API限流...")
time.sleep(10)

try:
    # 创建工作流实例
    workflow_instance = operation_service.create_workflow_instance(
        workflow_name='A股ETF数据更新',
        trigger_type=2,
        trigger_by='user'
    )

    print("\n正在下载 510300 的2年历史数据...")

    # 下载历史数据
    ticker = yf.Ticker('510300.SS')  # 尝试使用.SS后缀（上海证券交易所）
    data = ticker.history(period='2y', interval='1d')

    if data.empty:
        print("数据为空，尝试不带后缀...")
        ticker = yf.Ticker('510300')
        data = ticker.history(period='2y', interval='1d')

    if not data.empty:
        print(f"成功获取 {len(data)} 条数据记录")
        print(f"日期范围: {data.index[0].date()} 到 {data.index[-1].date()}")

        # 保存到数据库
        print("\n正在保存到数据库...")
        saved_count = 0
        for idx, row in data.iterrows():
            ETFData.objects.update_or_create(
                symbol='510300',
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
            if saved_count % 100 == 0:
                print(f"  已保存 {saved_count} 条记录...")

        print(f"\n✓ 成功保存 {saved_count} 条记录")

        # 获取实时数据
        print("\n正在获取实时数据...")
        info = ticker.info
        if info:
            current_price = info.get('currentPrice', info.get('regularMarketPrice', 0))
            print(f"  当前价格: ¥{current_price}")

        operation_service.complete_workflow_instance(
            workflow_instance,
            success=True,
            context_data={'saved_count': saved_count}
        )

        print("\n✓ 510300 数据更新完成！")
    else:
        print("✗ 无法获取数据")
        operation_service.complete_workflow_instance(
            workflow_instance,
            success=False,
            error_message="无法获取数据"
        )

except Exception as e:
    print(f"\n✗ 更新失败: {e}")
    import traceback
    traceback.print_exc()
