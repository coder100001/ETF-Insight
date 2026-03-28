"""
汇率服务 - 汇率数据获取和管理
"""

import requests
from datetime import datetime, timedelta
from django.utils import timezone
from .models import ExchangeRate
import logging
import yfinance as yf

logger = logging.getLogger(__name__)


class ExchangeRateService:
    """汇率服务"""

    # API端点（示例，需要替换为真实的API）
    API_ENDPOINTS = {
        'fixer': 'https://api.fixer.io/latest',
        'exchangerate': 'https://v6.exchangerate-api.com/v6/latest',
    }

    def __init__(self, api_key=None):
        self.api_key = api_key
        self.session = requests.Session()

    def fetch_from_api(self, source='system'):
        """
        从API获取汇率

        参数:
            source: 数据源（api, system, manual, yfinance, free_api）

        返回:
            dict: 汇率数据
        """
        if source == 'free_api':
            try:
                # 使用免费API获取真实汇率
                return self._fetch_free_api_rates()
            except Exception as e:
                logger.error(f"从免费API获取汇率失败: {e}")
                return None
        elif source == 'yfinance':
            try:
                # 使用yfinance获取真实汇率
                return self._fetch_yfinance_rates()
            except Exception as e:
                logger.error(f"从yfinance获取汇率失败: {e}")
                return None
        elif source == 'api' and self.api_key:
            try:
                # 这里可以调用真实的汇率API
                # 例如: response = requests.get(f"{API_ENDPOINT}?access_key={api_key}")
                pass
                pass
            except Exception as e:
                logger.error(f"从API获取汇率失败: {e}")
                return None
        elif source == 'system':
            # 使用系统默认汇率
            return self._get_default_rates()

        return None

    def _get_default_rates(self):
        """获取默认汇率"""
        return {
            'USD': {'CNY': 7.2, 'HKD': 7.8, 'USD': 1.0},
            'CNY': {'USD': 0.138889, 'HKD': 1.083333, 'CNY': 1.0},
            'HKD': {'USD': 0.128205, 'CNY': 0.923077, 'HKD': 1.0},
        }

    def _fetch_yfinance_rates(self):
        """
        使用yfinance获取真实汇率数据
        返回汇率字典
        """
        logger.info("开始从yfinance获取真实汇率数据...")

        try:
            # 获取USD到CNY的汇率 (USDCNY=X)
            usd_cny = yf.Ticker("USDCNY=X")
            usd_cny_rate = usd_cny.info.get('regularMarketPrice')

            # 获取USD到HKD的汇率 (USDHKD=X)
            usd_hkd = yf.Ticker("USDHKD=X")
            usd_hkd_rate = usd_hkd.info.get('regularMarketPrice')

            # 获取CNY到HKD的汇率 (CNYHKD=X)
            cny_hkd = yf.Ticker("CNYHKD=X")
            cny_hkd_rate = cny_hkd.info.get('regularMarketPrice')

            # 如果某些汇率获取失败，使用默认值
            if usd_cny_rate is None:
                logger.warning("USD/CNY汇率获取失败，使用默认值")
                usd_cny_rate = 7.2
            if usd_hkd_rate is None:
                logger.warning("USD/HKD汇率获取失败，使用默认值")
                usd_hkd_rate = 7.8
            if cny_hkd_rate is None:
                logger.warning("CNY/HKD汇率获取失败，使用默认值")
                cny_hkd_rate = 1.083333

            # 计算反向汇率
            rates = {
                'USD': {
                    'CNY': float(usd_cny_rate),
                    'HKD': float(usd_hkd_rate),
                    'USD': 1.0
                },
                'CNY': {
                    'USD': round(1.0 / usd_cny_rate, 6),
                    'HKD': float(cny_hkd_rate),
                    'CNY': 1.0
                },
                'HKD': {
                    'USD': round(1.0 / usd_hkd_rate, 6),
                    'CNY': round(1.0 / cny_hkd_rate, 6),
                    'HKD': 1.0
                }
            }

            logger.info(f"成功获取汇率数据: USD/CNY={usd_cny_rate}, USD/HKD={usd_hkd_rate}, CNY/HKD={cny_hkd_rate}")
            return rates

        except Exception as e:
            logger.error(f"从yfinance获取汇率失败: {e}")
            # 返回默认汇率
            logger.info("使用默认汇率数据")
            return self._get_default_rates()

    def _fetch_free_api_rates(self):
        """
        使用免费的汇率API获取数据
        使用 exchangerate-api.com 的免费端点
        """
        logger.info("开始从免费API获取真实汇率数据...")

        try:
            # 使用免费的汇率API
            url = "https://api.exchangerate-api.com/v4/latest/USD"
            response = requests.get(url, timeout=10)

            if response.status_code == 200:
                data = response.json()
                rates_data = data.get('rates', {})

                # 获取所需的汇率
                usd_cny_rate = rates_data.get('CNY', 7.2)
                usd_hkd_rate = rates_data.get('HKD', 7.8)

                # 计算CNY到HKD的汇率
                cny_hkd_rate = usd_hkd_rate / usd_cny_rate

                # 构建汇率字典
                rates = {
                    'USD': {
                        'CNY': float(usd_cny_rate),
                        'HKD': float(usd_hkd_rate),
                        'USD': 1.0
                    },
                    'CNY': {
                        'USD': round(1.0 / usd_cny_rate, 6),
                        'HKD': round(cny_hkd_rate, 6),
                        'CNY': 1.0
                    },
                    'HKD': {
                        'USD': round(1.0 / usd_hkd_rate, 6),
                        'CNY': round(1.0 / cny_hkd_rate, 6),
                        'HKD': 1.0
                    }
                }

                logger.info(f"成功获取汇率数据: USD/CNY={usd_cny_rate}, USD/HKD={usd_hkd_rate}, CNY/HKD={cny_hkd_rate}")
                return rates
            else:
                logger.warning(f"API返回状态码: {response.status_code}")
                return self._get_default_rates()

        except Exception as e:
            logger.error(f"从免费API获取汇率失败: {e}")
            # 返回默认汇率
            logger.info("使用默认汇率数据")
            return self._get_default_rates()

    def update_rates(self, rates=None, source='system'):
        """
        更新汇率到数据库

        参数:
            rates: 汇率数据字典（如果为None则从API获取）
            source: 数据源

        返回:
            dict: 更新结果
        """
        if rates is None:
            rates = self.fetch_from_api(source)

        if not rates:
            return {
                'success': False,
                'message': '无汇率数据可更新',
                'updated_count': 0
            }

        today = timezone.now().date()
        updated_count = 0
        errors = []

        for from_currency, to_rates in rates.items():
            for to_currency, rate in to_rates.items():
                try:
                    # 查询是否已存在今日汇率
                    existing = ExchangeRate.objects.filter(
                        from_currency=from_currency,
                        to_currency=to_currency,
                        rate_date=today
                    ).first()

                    if existing:
                        # 更新汇率
                        existing.rate = rate
                        existing.data_source = source
                        existing.save()
                        logger.info(f"更新汇率: 1 {from_currency} = {rate} {to_currency}")
                    else:
                        # 创建新汇率
                        ExchangeRate.objects.create(
                            from_currency=from_currency,
                            to_currency=to_currency,
                            rate=rate,
                            rate_date=today,
                            data_source=source
                        )
                        logger.info(f"创建汇率: 1 {from_currency} = {rate} {to_currency}")

                    updated_count += 1

                except Exception as e:
                    error_msg = f"更新汇率 {from_currency}->{to_currency} 失败: {str(e)}"
                    errors.append(error_msg)
                    logger.error(error_msg)

        return {
            'success': len(errors) == 0,
            'message': f"成功更新 {updated_count} 条汇率记录",
            'updated_count': updated_count,
            'errors': errors
        }

    def get_latest_rate(self, from_currency, to_currency):
        """
        获取最新汇率

        参数:
            from_currency: 源货币
            to_currency: 目标货币

        返回:
            float: 汇率
        """
        if from_currency == to_currency:
            return 1.0

        rate = ExchangeRate.objects.filter(
            from_currency=from_currency,
            to_currency=to_currency
        ).order_by('-rate_date').first()

        if rate:
            return float(rate.rate)
        else:
            logger.warning(f'未找到汇率: {from_currency} -> {to_currency}')
            # 尝试使用默认汇率
            default_rates = self._get_default_rates()
            if from_currency in default_rates and to_currency in default_rates[from_currency]:
                return default_rates[from_currency][to_currency]

            return 1.0

    def get_history(self, from_currency, to_currency, days=30):
        """
        获取汇率历史

        参数:
            from_currency: 源货币
            to_currency: 目标货币
            days: 天数

        返回:
            list: 历史汇率列表
        """
        end_date = timezone.now().date()
        start_date = end_date - timedelta(days=days)

        rates = ExchangeRate.objects.filter(
            from_currency=from_currency,
            to_currency=to_currency,
            rate_date__gte=start_date,
            rate_date__lte=end_date
        ).order_by('rate_date')

        return [
            {
                'date': str(rate.rate_date),
                'rate': float(rate.rate),
                'source': rate.data_source
            }
            for rate in rates
        ]

    def calculate_cross_rate(self, from_currency, to_currency):
        """
        计算交叉汇率（通过USD）

        参数:
            from_currency: 源货币
            to_currency: 目标货币

        返回:
            float: 交叉汇率
        """
        # 如果直接有汇率，直接返回
        direct_rate = self.get_latest_rate(from_currency, to_currency)
        if direct_rate != 1.0 or (from_currency == to_currency):
            return direct_rate

        # 通过USD计算交叉汇率
        from_to_usd = self.get_latest_rate(from_currency, 'USD')
        usd_to_target = self.get_latest_rate('USD', to_currency)

        cross_rate = from_to_usd * usd_to_target
        return cross_rate

    def convert(self, amount, from_currency, to_currency):
        """
        货币转换

        参数:
            amount: 金额
            from_currency: 源货币
            to_currency: 目标货币

        返回:
            float: 转换后的金额
        """
        if from_currency == to_currency:
            return amount

        rate = self.get_latest_rate(from_currency, to_currency)
        return amount * rate


# 全局汇率服务实例
exchange_rate_service = ExchangeRateService()


def update_exchange_rates_auto():
    """
    自动更新汇率（从免费API获取实时汇率）
    用于定时任务或手动触发
    """
    try:
        service = ExchangeRateService()
        # 尝试从免费API获取真实汇率
        result = service.update_rates(source='free_api')
        logger.info(f"汇率自动更新完成: {result['message']}")
        return result
    except Exception as e:
        logger.error(f"汇率自动更新失败: {e}")
        return {
            'success': False,
            'message': f"更新失败: {str(e)}",
            'updated_count': 0
        }


def update_exchange_rates_api(api_key):
    """
    从API更新汇率

    参数:
        api_key: API密钥

    返回:
        dict: 更新结果
    """
    try:
        service = ExchangeRateService(api_key=api_key)
        result = service.update_rates(source='api')
        return result
    except Exception as e:
        logger.error(f"从API更新汇率失败: {e}")
        return {
            'success': False,
            'message': f"更新失败: {str(e)}",
            'updated_count': 0
        }
