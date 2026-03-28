"""
尝试再次更新港股ETF（3466.HK）
使用更长的等待时间
"""
from workflow.models import ETFData
from workflow.operation_service import operation_service
import yfinance as yf
from datetime import datetime
import time
import random

print("="*60)
print("尝试更新港股ETF: 3466.HK")
print("="*60)

# 等待30秒
print("\n等待30秒避免API限流...")
for i in range(30, 0, -5):
    print(f"  剩余等待时间: {i} 秒", end='\r')
    time.sleep(5)
print()

try:
    # 创建工作流实例
    workflow_instance = operation_service.create_workflow_instance(
        workflow_name='港股ETF数据更新',
        trigger_type=2,
        trigger_by='user'
    )

    print("\n正在下载 3466.HK 的2年历史数据...")

    # 尝试下载历史数据
    ticker = yf.Ticker('3466.HK')
    data = ticker.history(period='2y', interval='1d')

    if not data.empty:
        print(f"成功获取 {len(data)} 条数据记录")
        print(f"日期范围: {data.index[0].date()} 到 {data.index[-1].date()}")

        # 保存到数据库
        print("\n正在保存到数据库...")
        saved_count = 0
        for idx, row in data.iterrows():
            ETFData.objects.update_or_create(
                symbol='3466.HK',
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
            print(f"  当前价格: HKD{current_price}")

        operation_service.complete_workflow_instance(
            workflow_instance,
            success=True,
            context_data={'saved_count': saved_count}
        )

        print("\n✓ 3466.HK 数据更新完成！")

    else:
        print("✗ 无法获取数据（API限流或数据不可用）")
        print("\n建议方案:")
        print("  1. 使用CSV导入工具（import_etf_csv.py）")
        print("  2. 等待1小时后再次尝试")
        print("  3. 使用VPN切换IP后重试")

        operation_service.complete_workflow_instance(
            workflow_instance,
            success=False,
            error_message="API限流：无法获取数据"
        )

except Exception as e:
    error_msg = str(e)
    print(f"\n✗ 更新失败: {error_msg}")

    if "Rate limited" in error_msg or "Too Many Requests" in error_msg:
        print("\n遇到API限流！")
        print("\n解决方案:")
        print("  1. 使用CSV导入工具")
        print("     - 从其他数据源下载CSV文件")
        print("     - 运行: python import_etf_csv.py <CSV文件> 3466.HK")
        print()
        print("  2. 等待1小时后再次尝试")
        print("     - 运行: python manage.py shell < update_hk_etf.py")
        print()
        print("  3. 创建示例数据用于测试")
        print("     - 运行: python manage.py shell < create_hk_sample_data.py")

        operation_service.complete_workflow_instance(
            workflow_instance,
            success=False,
            error_message="API限流"
        )
    else:
        import traceback
        traceback.print_exc()
        operation_service.complete_workflow_instance(
            workflow_instance,
            success=False,
            error_message=error_msg
        )

print("="*60)
