"""
ETF 基础数据层服务
提供 ETF 基础信息、持仓数据、价格数据的获取和更新功能
"""

import yfinance as yf
import pandas as pd
from datetime import datetime, timedelta, date
from typing import Optional, List, Dict, Any, Tuple
import logging
from decimal import Decimal

from .data_fetcher import retry_on_error, DATA_FETCH_CONFIG

logger = logging.getLogger(__name__)


class ETFDataService:
    """
    ETF 基础数据服务
    封装 ETF 基础信息、价格、持仓等数据的获取逻辑
    """

    def __init__(self):
        self.data_dir = 'etf_data'
        self._last_request_time = 0
        self._min_request_interval = 1.0

    def _rate_limit(self):
        """简单的速率限制"""
        import time
        current_time = time.time()
        elapsed = current_time - self._last_request_time
        if elapsed < self._min_request_interval:
            time.sleep(self._min_request_interval - elapsed)
        self._last_request_time = time.time()

    @retry_on_error()
    def fetch_etf_base_info(self, symbol: str) -> Optional[Dict[str, Any]]:
        """
        获取 ETF 基础信息

        Args:
            symbol: ETF 代码

        Returns:
            ETF 基础信息字典
        """
        self._rate_limit()
        logger.info(f"获取 {symbol} 基础信息")

        try:
            ticker = yf.Ticker(symbol)
            info = ticker.info

            if not info:
                logger.warning(f"{symbol} 没有获取到基础信息")
                return None

            base_info = {
                'symbol': symbol,
                'name': info.get('longName', ''),
                'name_en': info.get('shortName', ''),
                'market': self._detect_market(symbol, info),
                'asset_class': self._detect_asset_class(info),
                'category': info.get('category', ''),

                # 发行方
                'issuer': info.get('fundFamily', info.get('company', '')),

                # 跟踪信息
                'tracking_index': info.get('indexName', ''),
                'tracking_method': info.get('legalType', ''),

                # 时间与规模
                'inception_date': self._parse_date(info.get('fundInceptionDate')),
                'aum': Decimal(str(info.get('totalAssets', 0))) if info.get('totalAssets') else None,
                'aum_currency': info.get('currency', 'USD'),
                'shares_outstanding': info.get('sharesOutstanding', 0),

                # 费率
                'expense_ratio': Decimal(str(info.get('annualReportExpenseRatio', 0))) * 100 if info.get('annualReportExpenseRatio') else None,

                # 交易信息
                'listing_exchange': info.get('exchange', ''),
                'trading_currency': info.get('currency', 'USD'),
                'is_leveraged': info.get('isLeveraged', False),
                'leverage_ratio': Decimal(str(info.get('leverage', 1))),

                # 策略
                'investment_strategy': info.get('investmentStrategy', ''),
                'investment_objective': info.get('investmentObjective', ''),
                'benchmark': info.get('benchmark', ''),

                # 其他
                'beta': info.get('beta', None),
                'dividend_yield': Decimal(str(info.get('yield', 0))) * 100 if info.get('yield') else None,
                'three_year_return': Decimal(str(info.get('threeYearAverageReturn', 0))) * 100 if info.get('threeYearAverageReturn') else None,
                'five_year_return': Decimal(str(info.get('fiveYearAverageReturn', 0))) * 100 if info.get('fiveYearAverageReturn') else None,
            }

            return base_info

        except Exception as e:
            logger.error(f"获取 {symbol} 基础信息失败: {e}")
            raise

    def _detect_market(self, symbol: str, info: Dict) -> str:
        """检测市场"""
        exchange = info.get('exchange', '')
        country = info.get('country', '')

        market_map = {
            'US': ['NYSE', 'NASDAQ', 'BATS', 'ARCA'],
            'CN': ['SSE', 'SZSE'],
            'HK': ['HKEX'],
            'JP': ['TSE'],
        }

        for market, exchanges in market_map.items():
            if exchange in exchanges:
                return market

        if country == 'United States':
            return 'US'
        elif country == 'China':
            return 'CN'
        elif country == 'Hong Kong':
            return 'HK'

        return 'US'  # 默认美股

    def _detect_asset_class(self, info: Dict) -> str:
        """检测资产类别"""
        fund_type = info.get('fundType', '').upper()
        category = info.get('category', '').upper()

        if 'BOND' in fund_type or 'BOND' in category:
            return 'BOND'
        elif 'COMMODITY' in fund_type or 'COMMODITY' in category:
            return 'COMMODITY'
        elif 'CURRENCY' in fund_type or 'CURRENCY' in category:
            return 'CURRENCY'
        elif 'ALTERNATIVE' in fund_type:
            return 'ALTERNATIVE'
        elif 'MULTI' in fund_type:
            return 'MULTI_ASSET'
        else:
            return 'EQUITY'

    def _parse_date(self, date_str) -> Optional[date]:
        """解析日期字符串"""
        if not date_str:
            return None
        try:
            if isinstance(date_str, int):
                return datetime.fromtimestamp(date_str).date()
            elif isinstance(date_str, str):
                return datetime.strptime(date_str, '%Y-%m-%d').date()
        except:
            pass
        return None

    @retry_on_error()
    def fetch_holdings(self, symbol: str) -> Optional[List[Dict[str, Any]]]:
        """
        获取 ETF 持仓数据

        Args:
            symbol: ETF 代码

        Returns:
            持仓列表
        """
        self._rate_limit()
        logger.info(f"获取 {symbol} 持仓数据")

        try:
            ticker = yf.Ticker(symbol)
            holdings = ticker.institutional_holders

            if holdings is None or holdings.empty:
                logger.warning(f"{symbol} 没有获取到持仓数据")
                return None

            holdings_list = []
            for index, row in holdings.iterrows():
                holding = {
                    'holding_symbol': row.get('Symbol', ''),
                    'holding_name': row.get('Holder', ''),
                    'shares': Decimal(str(row.get('Shares', 0))),
                    'market_value': Decimal(str(row.get('Value', 0))),
                    'weight': Decimal(str(row.get('pctHeld', 0))) * 100 if 'pctHeld' in row else None,
                    'report_date': self._parse_date(row.get('Date Reported')),
                }
                holdings_list.append(holding)

            return holdings_list

        except Exception as e:
            logger.error(f"获取 {symbol} 持仓数据失败: {e}")
            raise

    @retry_on_error()
    def fetch_sector_distribution(self, symbol: str) -> Optional[List[Dict[str, Any]]]:
        """
        获取行业分布

        Args:
            symbol: ETF 代码

        Returns:
            行业分布列表
        """
        self._rate_limit()
        logger.info(f"获取 {symbol} 行业分布")

        try:
            ticker = yf.Ticker(symbol)
            sectors = ticker.sector_weights

            if sectors is None:
                logger.warning(f"{symbol} 没有获取到行业分布")
                return None

            sector_list = []
            for sector_name, weight in sectors.items():
                sector_list.append({
                    'sector_name': sector_name,
                    'weight': Decimal(str(weight)) * 100,
                })

            return sector_list

        except Exception as e:
            logger.error(f"获取 {symbol} 行业分布失败: {e}")
            return None

    @retry_on_error()
    def fetch_country_distribution(self, symbol: str) -> Optional[List[Dict[str, Any]]]:
        """
        获取地区/国家分布

        Args:
            symbol: ETF 代码

        Returns:
            地区分布列表
        """
        self._rate_limit()
        logger.info(f"获取 {symbol} 地区分布")

        try:
            ticker = yf.Ticker(symbol)
            countries = ticker.country_weights

            if countries is None:
                logger.warning(f"{symbol} 没有获取到地区分布")
                return None

            country_list = []
            for country_name, weight in countries.items():
                country_list.append({
                    'region_name': country_name,
                    'country': country_name,
                    'weight': Decimal(str(weight)) * 100,
                })

            return country_list

        except Exception as e:
            logger.error(f"获取 {symbol} 地区分布失败: {e}")
            return None

    @retry_on_error()
    def fetch_dividends(self, symbol: str, period: str = '5y') -> Optional[List[Dict[str, Any]]]:
        """
        获取分红数据

        Args:
            symbol: ETF 代码
            period: 时间周期

        Returns:
            分红记录列表
        """
        self._rate_limit()
        logger.info(f"获取 {symbol} 分红数据")

        try:
            ticker = yf.Ticker(symbol)
            dividends = ticker.dividends

            if dividends is None or dividends.empty:
                logger.warning(f"{symbol} 没有获取到分红数据")
                return None

            dividend_list = []
            for date, amount in dividends.items():
                dividend_list.append({
                    'ex_dividend_date': date.date() if hasattr(date, 'date') else date,
                    'dividend_amount': Decimal(str(amount)),
                    'dividend_type': 'CASH',
                })

            return dividend_list

        except Exception as e:
            logger.error(f"获取 {symbol} 分红数据失败: {e}")
            raise

    def save_to_database(self, model_class, data: Dict, unique_fields: List[str]) -> bool:
        """
        保存数据到数据库

        Args:
            model_class: Django 模型类
            data: 数据字典
            unique_fields: 唯一字段列表

        Returns:
            是否成功
        """
        try:
            # 构建查询条件
            query = {field: data[field] for field in unique_fields if field in data}

            # 更新或创建
            obj, created = model_class.objects.update_or_create(
                defaults=data,
                **query
            )

            action = '创建' if created else '更新'
            logger.debug(f"{action} {model_class.__name__}: {query}")
            return True

        except Exception as e:
            logger.error(f"保存 {model_class.__name__} 失败: {e}")
            return False

    def sync_etf_base_info(self, symbol: str) -> bool:
        """
        同步 ETF 基础信息到数据库

        Args:
            symbol: ETF 代码

        Returns:
            是否成功
        """
        try:
            from workflow.models_etf_data_layer import ETFBaseInfo

            data = self.fetch_etf_base_info(symbol)
            if not data:
                return False

            return self.save_to_database(ETFBaseInfo, data, ['symbol'])

        except Exception as e:
            logger.error(f"同步 {symbol} 基础信息失败: {e}")
            return False

    def sync_etf_holdings(self, symbol: str, report_date: date = None) -> bool:
        """
        同步 ETF 持仓到数据库

        Args:
            symbol: ETF 代码
            report_date: 报告日期

        Returns:
            是否成功
        """
        try:
            from workflow.models_etf_data_layer import ETFHolding, ETFBaseInfo

            holdings = self.fetch_holdings(symbol)
            if not holdings:
                return False

            # 获取 ETF 对象
            etf = ETFBaseInfo.objects.filter(symbol=symbol).first()
            if not etf:
                logger.warning(f"ETF {symbol} 不存在，跳过持仓同步")
                return False

            if report_date is None:
                report_date = date.today()

            success_count = 0
            for holding_data in holdings:
                holding_data['etf'] = etf
                holding_data['symbol'] = symbol
                holding_data['report_date'] = report_date

                if self.save_to_database(ETFHolding, holding_data,
                                         ['symbol', 'holding_symbol', 'report_date']):
                    success_count += 1

            logger.info(f"同步 {symbol} 持仓完成: {success_count}/{len(holdings)}")
            return success_count > 0

        except Exception as e:
            logger.error(f"同步 {symbol} 持仓失败: {e}")
            return False

    def sync_etf_sectors(self, symbol: str, report_date: date = None) -> bool:
        """同步行业分布"""
        try:
            from workflow.models_etf_data_layer import ETFHoldingSector, ETFBaseInfo

            sectors = self.fetch_sector_distribution(symbol)
            if not sectors:
                return False

            etf = ETFBaseInfo.objects.filter(symbol=symbol).first()
            if not etf:
                return False

            if report_date is None:
                report_date = date.today()

            for sector_data in sectors:
                sector_data['etf'] = etf
                sector_data['symbol'] = symbol
                sector_data['report_date'] = report_date

                self.save_to_database(ETFHoldingSector, sector_data,
                                      ['symbol', 'sector_name', 'report_date'])

            return True

        except Exception as e:
            logger.error(f"同步 {symbol} 行业分布失败: {e}")
            return False

    def sync_etf_regions(self, symbol: str, report_date: date = None) -> bool:
        """同步地区分布"""
        try:
            from workflow.models_etf_data_layer import ETFHoldingRegion, ETFBaseInfo

            regions = self.fetch_country_distribution(symbol)
            if not regions:
                return False

            etf = ETFBaseInfo.objects.filter(symbol=symbol).first()
            if not etf:
                return False

            if report_date is None:
                report_date = date.today()

            for region_data in regions:
                region_data['etf'] = etf
                region_data['symbol'] = symbol
                region_data['report_date'] = report_date

                self.save_to_database(ETFHoldingRegion, region_data,
                                      ['symbol', 'region_name', 'report_date'])

            return True

        except Exception as e:
            logger.error(f"同步 {symbol} 地区分布失败: {e}")
            return False

    def sync_etf_dividends(self, symbol: str) -> bool:
        """同步分红数据"""
        try:
            from workflow.models_etf_data_layer import ETFDividend, ETFBaseInfo

            dividends = self.fetch_dividends(symbol)
            if not dividends:
                return False

            etf = ETFBaseInfo.objects.filter(symbol=symbol).first()
            if not etf:
                return False

            success_count = 0
            for div_data in dividends:
                div_data['etf'] = etf
                div_data['symbol'] = symbol

                if self.save_to_database(ETFDividend, div_data,
                                         ['symbol', 'ex_dividend_date']):
                    success_count += 1

            logger.info(f"同步 {symbol} 分红完成: {success_count}/{len(dividends)}")
            return success_count > 0

        except Exception as e:
            logger.error(f"同步 {symbol} 分红失败: {e}")
            return False

    def full_sync_etf(self, symbol: str) -> Dict[str, bool]:
        """
        完整同步 ETF 所有数据

        Args:
            symbol: ETF 代码

        Returns:
            各模块同步结果
        """
        logger.info(f"开始完整同步 {symbol}")

        results = {
            'base_info': self.sync_etf_base_info(symbol),
            'holdings': self.sync_etf_holdings(symbol),
            'sectors': self.sync_etf_sectors(symbol),
            'regions': self.sync_etf_regions(symbol),
            'dividends': self.sync_etf_dividends(symbol),
        }

        logger.info(f"{symbol} 同步结果: {results}")
        return results


# 全局单例
_etf_data_service = None


def get_etf_data_service() -> ETFDataService:
    """获取 ETFDataService 单例"""
    global _etf_data_service
    if _etf_data_service is None:
        _etf_data_service = ETFDataService()
    return _etf_data_service
