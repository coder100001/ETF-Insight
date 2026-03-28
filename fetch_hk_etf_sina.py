"""
使用新浪财经API拉取港股ETF数据
新浪财经提供免费的港股行情API，不受Yahoo限流影响

港股代码格式说明:
- 新浪财经使用4位代码，如 3466（对应3466.HK）
- 查询参数: symbol=rt_hk&list=3466

API说明:
- 实时数据: http://hq.sinajs.cn/list=rt_hk,3466
- 历史数据: http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData

安装:
    pip install requests

用法:
    # 方式一：直接运行（不推荐，会报错）
    # python fetch_hk_etf_sina.py 3466
    # python fetch_hk_etf_sina.py 3466 3000 7000
    
    # 方式二：通过Django shell运行（推荐）
    # python manage.py shell < fetch_hk_etf_sina.py 3466
    # python manage.py shell < fetch_hk_etf_sina.py 3466 3000 7000
"""

import requests
import pandas as pd
from workflow.models import ETFData
from workflow.operation_service import operation_service
from datetime import datetime, timedelta
import time
import sys
import json


def get_hk_stock_realtime(sina_code):
    """
    获取港股实时数据

    参数:
        sina_code: 新浪4位代码（如3466）

    返回:
        dict: 实时数据
    """
    try:
        url = f"http://hq.sinajs.cn/list=rt_hk,{sina_code}"
        response = requests.get(url, timeout=10)

        if response.status_code != 200:
            print(f"获取实时数据失败: HTTP {response.status_code}")
            return None

        # 解析返回的JavaScript代码
        content = response.text
        if 'hq_str_' in content:
            # 提取数据
            start = content.index('"')
            end = content.rindex('"')
            data_str = content[start+1:end]
            data_list = data_str.split(',')

            # 解析字段
            # 格式: 代码,名称,开盘价,昨收价,最高价,最低价,买一价,卖一价,成交手数,成交金额,日期时间
            if len(data_list) >= 12:
                return {
                    'code': data_list[0],
                    'name': data_list[1],
                    'open': float(data_list[2]) if data_list[2] else 0,
                    'previous_close': float(data_list[3]) if data_list[3] else 0,
                    'high': float(data_list[4]) if data_list[4] else 0,
                    'low': float(data_list[5]) if data_list[5] else 0,
                    'current': float(data_list[6]) if data_list[6] else 0,
                    'volume': int(data_list[8]) if data_list[8] else 0,
                }

        return None

    except Exception as e:
        print(f"获取实时数据出错: {e}")
        return None


def get_hk_stock_history(sina_code, start_date, end_date):
    """
    获取港股历史数据（新浪财经）

    参数:
        sina_code: 新浪4位代码（如3466）
        start_date: 开始日期 (datetime.date)
        end_date: 结束日期 (datetime.date)

    返回:
        DataFrame: 历史数据
    """
    try:
        # 新浪财经历史数据API
        url = "http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData"
        params = {
            'symbol': f'hk{sina_code}',
            'scale': '240',  # 240=日线
            'ma': 'no',
            'datalen': '500',
        }

        print(f"\n请求URL: {url}")
        print(f"参数: {params}")

        response = requests.get(url, params=params, timeout=30)

        if response.status_code != 200:
            print(f"获取历史数据失败: HTTP {response.status_code}")
            return None

        # 解析JSON
        data = response.json()

        if 'result' not in data or not data['result']:
            print("返回数据为空")
            print(f"返回内容: {data}")
            return None

        # 提取K线数据
        kline_data = data['result']['data']

        if not kline_data:
            print("K线数据为空")
            return None

        # 转换为DataFrame
        df = pd.DataFrame(kline_data)

        # 重命名列
        column_map = {
            'd': 'date',
            'o': 'open',
            'h': 'high',
            'l': 'low',
            'c': 'close',
            'v': 'volume'
        }
        df = df.rename(columns=column_map)

        # 解析日期（新浪返回的是毫秒时间戳）
        df['date'] = pd.to_datetime(df['date'], unit='ms').dt.date

        # 筛选日期范围
        df = df[(df['date'] >= start_date) & (df['date'] <= end_date)]

        print(f"获取到 {len(df)} 条历史数据")
        print(f"日期范围: {df['date'].min()} 到 {df['date'].max()}")

        # 选择需要的列
        df = df[['date', 'open', 'high', 'low', 'close', 'volume']]

        return df

    except Exception as e:
        print(f"获取历史数据出错: {e}")
        import traceback
        traceback.print_exc()
        return None


def save_sina_data_to_db(data, symbol, realtime_data=None):
    """
    将新浪财经数据保存到数据库

    参数:
        data: 历史数据DataFrame
        symbol: ETF代码（如3466.HK）
        realtime_data: 实时数据dict

    返回:
        (成功, 保存记录数)
    """
    if data is None or data.empty:
        return False, 0

    try:
        # 创建工作流实例
        workflow_instance = operation_service.create_workflow_instance(
            workflow_name='港股ETF数据更新(新浪)',
            trigger_type=2,
            trigger_by='user'
        )

        print("\n正在保存到数据库...")
        saved_count = 0

        # 保存历史数据
        for idx in range(len(data)):
            row = data.iloc[idx]

            ETFData.objects.update_or_create(
                symbol=symbol,
                date=row['date'],
                defaults={
                    'open_price': float(row['open']),
                    'high_price': float(row['high']),
                    'low_price': float(row['low']),
                    'close_price': float(row['close']),
                    'volume': int(row['volume']),
                    'data_source': 'sina',
                    'fetch_instance': workflow_instance,
                }
            )
            saved_count += 1

            if saved_count % 100 == 0:
                print(f"  已保存 {saved_count} 条记录...")

        print(f"\n✓ 历史数据保存成功，共 {saved_count} 条记录")

        # 完成工作流实例
        operation_service.complete_workflow_instance(
            workflow_instance,
            success=True,
            context_data={'saved_count': saved_count, 'data_source': 'sina'}
        )

        return True, saved_count

    except Exception as e:
        print(f"\n✗ 保存到数据库失败: {str(e)}")
        import traceback
        traceback.print_exc()
        return False, 0


def fetch_hk_etf_sina(symbol_hk, retry=2):
    """
    使用新浪财经拉取港股ETF数据

    参数:
        symbol_hk: 港股代码（如3466.HK）
        retry: 重试次数

    返回:
        (成功, 数据记录数)
    """
    # 转换为新浪4位代码
    if '.HK' in symbol_hk:
        sina_code = symbol_hk.replace('.HK', '')
    else:
        sina_code = symbol_hk

    print(f"\n{'='*60}")
    print(f"使用新浪财经拉取港股ETF: {symbol_hk}")
    print(f"新浪代码: {sina_code}")
    print(f"{'='*60}")

    # 计算日期范围（2年）
    end_date = datetime.now().date()
    start_date = end_date - timedelta(days=730)

    for attempt in range(retry):
        try:
            print(f"\n[尝试 {attempt + 1}/{retry}]")

            # 获取历史数据
            print("\n获取历史数据...")
            data = get_hk_stock_history(sina_code, start_date, end_date)

            if data is None or data.empty:
                print("历史数据获取失败")
                return False, 0

            # 获取实时数据
            print("\n获取实时数据...")
            realtime = get_hk_stock_realtime(sina_code)

            if realtime:
                print(f"  当前价格: HKD{realtime['current']:.2f}")
                print(f"  涨跌: {realtime['current'] - realtime['previous_close']:+.2f}")
            else:
                print("  实时数据获取失败")

            # 保存到数据库
            success, count = save_sina_data_to_db(data, symbol_hk, realtime)

            if success:
                print(f"\n✓ {symbol_hk} 数据更新完成！")
                return True, count
            else:
                return False, 0

        except Exception as e:
            error_msg = str(e)
            print(f"\n✗ 尝试 {attempt + 1} 失败: {error_msg}")

            if attempt < retry - 1:
                wait_time = 5
                print(f"等待 {wait_time} 秒后重试...")
                time.sleep(wait_time)

    return False, 0


def main():
    """主函数"""
    print("="*60)
    print("港股ETF数据拉取工具 - 新浪财经API")
    print("="*60)

    # 检查命令行参数
    if len(sys.argv) < 2:
        print("\n用法: python manage.py shell < fetch_hk_etf_sina.py <港股代码1> [港股代码2] ...")
        print("\n代码格式:")
        print("  3466       - 表示3466.HK")
        print("  3466.HK    - 表示3466.HK")
        print("\n调用方式:")
        print("  python manage.py shell < fetch_hk_etf_sina.py 3466")
        print("\n示例:")
        print("  python manage.py shell < fetch_hk_etf_sina.py 3466              # 更新单个")
        print("  python manage.py shell < fetch_hk_etf_sina.py 3466 3000 7000  # 更新多个")
        print("\n说明:")
        print("  - 使用新浪财经免费API")
        print("  - 不受Yahoo Finance限流影响")
        print("  - 数据来源: 新浪财经")
        print("  - 必须通过Django shell运行（因为有Django模型导入）")
        print("="*60)
        return

    # 获取要更新的ETF列表
    symbols_input = sys.argv[1:]
    symbols = []
    for s in symbols_input:
        # 统一转换为.HK格式
        s_upper = s.upper()
        if not '.HK' in s_upper and s_upper.isdigit():
            s_upper = f"{s_upper}.HK"
        symbols.append(s_upper)

    print(f"\n准备更新的ETF: {', '.join(symbols)}")

    # 检查requests库
    try:
        import requests
        print(f"\nrequests版本: {requests.__version__}")
    except ImportError:
        print("\n❌ 未安装requests库！")
        print("\n请先安装:")
        print("  pip install requests")
        print("="*60)
        return

    # 分别更新每个ETF
    results = {}
    for i, symbol in enumerate(symbols, 1):
        print(f"\n{'='*60}")
        print(f"进度: {i}/{len(symbols)} - {symbol}")
        print(f"{'='*60}")

        # 拉取数据
        success, count = fetch_hk_etf_sina(symbol, retry=2)

        results[symbol] = {
            'success': success,
            'count': count
        }

        # 如果不是最后一个，添加延迟
        if i < len(symbols):
            print(f"\n等待3秒后更新下一个ETF...")
            time.sleep(3)

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
