"""
ETF工作流系统 - Django View层（MTV中的V）- 汇率管理部分
"""

from django.shortcuts import render, get_object_or_404
from django.http import JsonResponse, HttpResponseRedirect
from django.views import View
from django.utils import timezone
from django.db.models import Count, Q
from django.core.serializers.json import DjangoJSONEncoder
from datetime import datetime, timedelta
import json
from rest_framework import viewsets, status
from rest_framework.decorators import action
from rest_framework.response import Response
import requests

from .models import (
    Workflow, WorkflowStep, WorkflowInstance,
    WorkflowInstanceStep, SystemLog, Notification,
    ETFData, PortfolioConfig, AnalysisReport, ETFConfig, ExchangeRate, OperationLog
)
from .services_exchange_rate import exchange_rate_service, update_exchange_rates_auto


# ===============================================================================
# 汇率管理视图
# ===============================================================================

class ExchangeRateListView(View):
    """汇率列表页面"""
    def get(self, request):
        # 获取所有货币对
        today = timezone.now().date()
        yesterday = today - timedelta(days=1)

        # 今日汇率 - 只获取data_source不是test_data的记录
        today_rates = ExchangeRate.objects.filter(
            rate_date=today
        ).exclude(data_source='test_data')
        
        # 昨日汇率（用于对比）
        yesterday_rates = ExchangeRate.objects.filter(rate_date=yesterday)
        
        # 构建汇率矩阵
        currency_pairs = [
            ('USD', 'CNY', '美元-人民币'),
            ('USD', 'HKD', '美元-港币'),
            ('CNY', 'HKD', '人民币-港币'),
            ('CNY', 'USD', '人民币-美元'),
            ('HKD', 'USD', '港币-美元'),
            ('HKD', 'CNY', '港币-人民币'),
        ]
        
        rates_data = []
        for from_curr, to_curr, description in currency_pairs:
            today_rate = today_rates.filter(
                from_currency=from_curr, 
                to_currency=to_curr
            ).first()
            
            yesterday_rate = yesterday_rates.filter(
                from_currency=from_curr, 
                to_currency=to_curr
            ).first()
            
            today_value = float(today_rate.rate) if today_rate else 0
            yesterday_value = float(yesterday_rate.rate) if yesterday_rate else 0
            
            # 计算变化
            if yesterday_value > 0:
                change = today_value - yesterday_value
                change_percent = (change / yesterday_value) * 100
            else:
                change = 0
                change_percent = 0
            
            rates_data.append({
                'from_currency': from_curr,
                'to_currency': to_curr,
                'description': description,
                'today_rate': today_value,
                'yesterday_rate': yesterday_value,
                'change': change,
                'change_percent': change_percent,
                'data_source': today_rate.data_source if today_rate else '-',
                'updated_at': today_rate.updated_at if today_rate else None,
            })
        
        # 最近7天的汇率历史（排除测试数据）
        seven_days_ago = today - timedelta(days=6)
        recent_rates = ExchangeRate.objects.filter(
            rate_date__gte=seven_days_ago,
            rate_date__lte=today
        ).exclude(data_source='test_data').order_by('rate_date', 'from_currency', 'to_currency')
        
        # 计算涨跌数量
        up_count = sum(1 for r in rates_data if r['change_percent'] > 0)
        down_count = sum(1 for r in rates_data if r['change_percent'] < 0)
        
        # 按日期分组，并为每个货币对创建快捷访问字典
        history_by_date = {}
        history_rates = {}  # 用于图表数据

        # 初始化所有货币对的字典
        currency_pairs = ['USD_CNY', 'USD_HKD', 'CNY_HKD', 'CNY_USD', 'HKD_USD', 'HKD_CNY']
        for pair in currency_pairs:
            history_rates[pair] = {}

        for rate in recent_rates:
            date_str = str(rate.rate_date)
            if date_str not in history_by_date:
                history_by_date[date_str] = []
            history_by_date[date_str].append({
                'from': rate.from_currency,
                'to': rate.to_currency,
                'rate': float(rate.rate),
            })

            # 创建快捷访问字典
            pair_key = f"{rate.from_currency}_{rate.to_currency}"
            if pair_key not in history_rates:
                history_rates[pair_key] = {}
            history_rates[pair_key][date_str] = float(rate.rate)
        
        context = {
            'rates_data': rates_data,
            'history_by_date': history_by_date,
            'history_rates': history_rates,
            'up_count': up_count,
            'down_count': down_count,
            'today': today,
            'yesterday': yesterday,
        }
        return render(request, 'workflow/exchange_rate_list.html', context)


class ExchangeRateUpdateView(View):
    """更新汇率页面"""
    def get(self, request):
        # 手动触发更新（从免费API获取真实汇率）
        try:
            result = exchange_rate_service.update_rates(source='free_api')
            message = f"成功更新 {result['updated_count']} 条汇率记录"
            success = True
        except Exception as e:
            message = f"更新失败: {str(e)}"
            success = False

        return JsonResponse({
            'success': success,
            'message': message,
            'result': result if success else None
        })


class ExchangeRateHistoryView(View):
    """汇率历史查询API"""
    def get(self, request):
        from_currency = request.GET.get('from', 'USD')
        to_currency = request.GET.get('to', 'CNY')
        days = int(request.GET.get('days', 30))

        end_date = timezone.now().date()
        start_date = end_date - timedelta(days=days)

        rates = ExchangeRate.objects.filter(
            from_currency=from_currency,
            to_currency=to_currency,
            rate_date__gte=start_date,
            rate_date__lte=end_date
        ).order_by('rate_date')

        data = [{
            'date': str(rate.rate_date),
            'rate': float(rate.rate),
            'source': rate.data_source,
        } for rate in rates]

        return JsonResponse({
            'success': True,
            'from_currency': from_currency,
            'to_currency': to_currency,
            'data': data
        })


class ExchangeRateConvertView(View):
    """货币转换API"""
    def get(self, request):
        from_currency = request.GET.get('from', 'USD')
        to_currency = request.GET.get('to', 'CNY')
        amount = float(request.GET.get('amount', 0))

        if amount <= 0:
            return JsonResponse({
                'success': False,
                'message': '金额必须大于0'
            })

        # 使用汇率服务进行转换
        rate = exchange_rate_service.get_latest_rate(from_currency, to_currency)
        result = exchange_rate_service.convert(amount, from_currency, to_currency)

        return JsonResponse({
            'success': True,
            'from_currency': from_currency,
            'to_currency': to_currency,
            'amount': amount,
            'rate': rate,
            'result': result
        })
