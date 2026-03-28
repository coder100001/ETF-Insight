"""
更新港股和A股ETF数据
"""
from workflow.scheduler import update_etf_data
import time

print("="*60)
print("港股和A股ETF数据更新")
print("="*60)
print()

print("准备更新的ETF:")
print("  1. 3466.HK - 南方恒生科技ETF（港股）")
print("  2. 510300 - 华泰柏瑞沪深300ETF（A股）")
print()

print("等待10秒避免API限流...")
time.sleep(10)

print()
print("开始更新数据...")
print("-"*60)

success = update_etf_data(symbols=['3466.HK', '510300'])

print("-"*60)
print()
if success:
    print("✓ 数据更新成功！")
else:
    print("✗ 数据更新失败，请查看日志")
print()
print("="*60)
