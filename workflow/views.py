"""
ETF工作流系统 - Django View层（MTV中的V）
"""

import logging
logger = logging.getLogger(__name__)

from django.shortcuts import render, get_object_or_404
from django.http import JsonResponse
from django.views import View
from django.utils import timezone
from django.db.models import Count, Q
from django.core.serializers.json import DjangoJSONEncoder
from datetime import datetime, timedelta, date
import json
from rest_framework import viewsets, status
from rest_framework.decorators import action
from rest_framework.response import Response

from .models import (
    Workflow, WorkflowStep, WorkflowInstance,
    WorkflowInstanceStep, SystemLog, Notification,
    ETFData, PortfolioConfig, AnalysisReport, ETFConfig, ExchangeRate, OperationLog
)
from .serializers import (
    WorkflowSerializer, WorkflowStepSerializer,
    WorkflowInstanceSerializer, WorkflowInstanceStepSerializer,
    SystemLogSerializer
)
from .engine import workflow_engine
from .services import etf_service
from .services_exchange_rate import update_exchange_rates_auto


# ===============================================================================
# 页面视图（Template Views）
# ===============================================================================

class DashboardView(View):
    """仪表板页面"""
    def get(self, request):
        # 今日统计
        today_start = timezone.now().replace(hour=0, minute=0, second=0, microsecond=0)
        today_instances = WorkflowInstance.objects.filter(created_at__gte=today_start)
        
        today_total = today_instances.count()
        today_success = today_instances.filter(status=2).count()
        today_failed = today_instances.filter(status=3).count()
        today_running = today_instances.filter(status=1).count()
        
        # 近7天趋势
        seven_days_ago = timezone.now() - timedelta(days=7)
        recent_instances = WorkflowInstance.objects.filter(
            created_at__gte=seven_days_ago
        ).order_by('created_at')
        
        # 计算每日统计数据（用于图表）
        daily_stats = {}
        from django.db.models import Count
        from django.db.models.functions import TruncDate
        
        for i in range(7):
            date = (timezone.now() - timedelta(days=i)).date()
            daily_instances = recent_instances.filter(created_at__date=date)
            daily_stats[str(date)] = {
                'total': daily_instances.count(),
                'success': daily_instances.filter(status=2).count(),
                'failed': daily_instances.filter(status=3).count(),
                'running': daily_instances.filter(status=1).count(),
            }
        
        # 计算工作流统计
        workflow_stats = []
        workflows = Workflow.objects.all()
        for wf in workflows:
            wf_instances = WorkflowInstance.objects.filter(workflow=wf)
            total = wf_instances.count()
            success = wf_instances.filter(status=2).count()
            success_rate = (success / total * 100) if total > 0 else 0
            
            # 只添加有实例数据的工作流
            if total > 0:
                workflow_stats.append({
                    'name': wf.name,
                    'total': total,
                    'success': success,
                    'success_rate': success_rate,
                    'status': wf.status,
                })
        
        context = {
            'today_total': today_total,
            'today_success': today_success,
            'today_failed': today_failed,
            'today_running': today_running,
            'recent_instances': recent_instances,
            'daily_stats': json.dumps(daily_stats),
            'workflow_stats': workflow_stats,
        }
        return render(request, 'workflow/dashboard.html', context)


class WorkflowListView(View):
    """工作流列表页面"""
    def get(self, request):
        workflows = Workflow.objects.all()
        context = {
            'workflows': workflows
        }
        return render(request, 'workflow/workflow_list.html', context)


class WorkflowDetailView(View):
    """工作流详情页面"""
    def get(self, request, workflow_id):
        workflow = get_object_or_404(Workflow, id=workflow_id)
        steps = WorkflowStep.objects.filter(workflow=workflow).order_by('order_index')
        
        context = {
            'workflow': workflow,
            'steps': steps
        }
        return render(request, 'workflow/workflow_detail.html', context)


class InstanceListView(View):
    """实例列表页面"""
    def get(self, request):
        # 获取筛选参数
        status_filter = request.GET.get('status', '')

        # 获取分页参数
        page = int(request.GET.get('page', 1))
        per_page = 20  # 每页显示20条

        # 构建查询集
        all_instances = WorkflowInstance.objects.all()

        # 应用状态筛选
        if status_filter:
            all_instances = all_instances.filter(status=status_filter)

        all_instances = all_instances.order_by('-created_at')

        # 计算分页
        total_count = all_instances.count()
        total_pages = (total_count + per_page - 1) // per_page if total_count > 0 else 1

        # 确保页码在有效范围内
        page = max(1, min(page, total_pages))

        # 获取当前页的数据
        start = (page - 1) * per_page
        end = start + per_page
        instances = all_instances[start:end]

        context = {
            'instances': instances,
            'page': page,
            'total_pages': total_pages,
            'total_count': total_count,
            'per_page': per_page,
            'status_filter': status_filter,
            'status_choices': WorkflowInstance.STATUS_CHOICES
        }
        return render(request, 'workflow/instance_list.html', context)


class InstanceDetailView(View):
    """实例详情页面"""
    def get(self, request, instance_id):
        instance = get_object_or_404(WorkflowInstance, id=instance_id)
        steps = WorkflowInstanceStep.objects.filter(workflow_instance=instance).order_by('workflow_step__order_index')
        logs = SystemLog.objects.filter(workflow_instance=instance).order_by('created_at')
        
        context = {
            'instance': instance,
            'steps': steps,
            'logs': logs
        }
        return render(request, 'workflow/instance_detail.html', context)


class ETFDashboardView(View):
    """ETF仪表板页面"""
    def get(self, request):
        # 获取ETF对比数据
        etf_list = etf_service.get_comparison_data()
        
        context = {
            'etf_list': etf_list
        }
        return render(request, 'workflow/etf_dashboard.html', context)


class ETFDetailView(View):
    """ETF详情页面"""
    def get(self, request, symbol):
        # 获取实时数据
        realtime = etf_service.fetch_realtime_data(symbol)
        # 获取ETF信息
        etf_info = etf_service.ETF_INFO.get(symbol, {})
        # 获取关键指标
        metrics = etf_service.calculate_metrics(symbol, period='1y')
        # 获取图表数据
        chart_data = etf_service.get_historical_chart_data(symbol, period='1y')

        context = {
            'symbol': symbol,
            'realtime': realtime,
            'info': etf_info,
            'etf_info': etf_info,  # 保持原有变量，以防其他模板需要
            'metrics': metrics,
            'chart_data': chart_data,
        }
        return render(request, 'workflow/etf_detail.html', context)


class ETFComparisonView(View):
    """ETF对比页面"""
    def get(self, request):
        # 获取查询参数
        period = request.GET.get('period', '1y')
        
        # 获取ETF对比数据
        comparison_data = etf_service.get_comparison_data(period)
        
        # 获取图表数据
        chart_data = etf_service.get_comparison_chart_data(period)
        
        context = {
            'etfs': etf_service.SYMBOLS,
            'symbols': comparison_data,      # 提供模板所需的symbols变量
            'comparison': comparison_data,  # 提供模板所需的comparison变量
            'chart_data': chart_data,      # 提供图表数据
            'period': period
        }
        return render(request, 'workflow/etf_comparison.html', context)


class ETFForecastView(View):
    """ETF收益预测页面"""
    def get(self, request):
        context = {
            'etfs': etf_service.SYMBOLS
        }
        return render(request, 'workflow/etf_forecast.html', context)


class PortfolioAnalysisView(View):
    """投资组合分析页面"""
    def get(self, request):
        # 获取投资金额参数（如果提供）
        total_investment = float(request.GET.get('investment', 10000))

        # 获取ETF权重（使用预设值或默认值）
        etf_weights = []
        has_user_weights = False

        # 首先检查是否有用户提供的权重
        for symbol in etf_service.SYMBOLS:
            weight_param = request.GET.get(symbol.lower())
            if weight_param is not None:
                has_user_weights = True
                break

        if has_user_weights:
            # 使用用户提供的权重
            for symbol in etf_service.SYMBOLS:
                weight_param = request.GET.get(symbol.lower())
                weight = float(weight_param) if weight_param and weight_param.strip() else 0

                etf_weights.append({
                    'symbol': symbol,
                    'weight': weight  # 保留为百分比数值（17），不转换为小数
                })
        else:
            # 默认使用433组合：SCHD=40%、SPYD=30%、JEPQ=30%、JEPI=0%、VYM=0%
            default_weights = {
                'SCHD': 40,
                'SPYD': 30,
                'JEPQ': 30,
                'JEPI': 0,
                'VYM': 0
            }

            for symbol in etf_service.SYMBOLS:
                etf_weights.append({
                    'symbol': symbol,
                    'weight': float(default_weights.get(symbol, 0))  # 使用默认权重，未配置的ETF为0
                })

        # 计算分配比例（转换为小数用于计算）
        allocation = {}
        for weight_info in etf_weights:
            # 将百分比转换为小数（17 -> 0.17）
            allocation[weight_info['symbol']] = weight_info['weight'] / 100.0

        # 计算投资组合分析
        portfolio = etf_service.analyze_portfolio(allocation, total_investment)

        context = {
            'etfs': etf_service.SYMBOLS,
            'etf_weights': etf_weights,
            'total_investment': total_investment,
            'result': portfolio,
            'symbols': etf_service.SYMBOLS
        }
        return render(request, 'workflow/portfolio_analysis.html', context)
    
    def post(self, request):
        """处理投资组合预测请求"""
        # 获取POST参数
        allocation_str = request.POST.get('allocation', '{}')
        try:
            allocation = json.loads(allocation_str)
        except:
            allocation = {}
        
        total_investment = float(request.POST.get('total_investment', 10000))
        tax_rate = float(request.POST.get('tax_rate', 0.10))
        
        # 获取预测年份
        forecast_years = request.POST.get('forecast_years', '3,5,10')
        years = [int(y) for y in forecast_years.split(',')]
        
        # 计算投资组合分析
        portfolio = etf_service.analyze_portfolio(allocation, total_investment, tax_rate)
        
        # 为每个ETF计算预测收益（中性情况）
        forecast_data = {}
        for symbol in allocation.keys():
            if symbol in etf_service.SYMBOLS and allocation[symbol] > 0:
                forecast_data[symbol] = {}
                for year in years:
                    forecast_result = etf_service.forecast_etf_growth(
                        symbol,
                        total_investment * allocation[symbol],
                        None,
                        tax_rate
                    )
                    if 'error' not in forecast_result:
                        forecast_data[symbol][year] = forecast_result['forecasts'][str(year)]
        
        # 计算不同场景的组合预测（保守、中性、乐观）
        scenarios = {
            'pessimistic': 0.04,    # 悲观：年化4%
            'conservative': 0.06,   # 保守：年化6%
            'neutral': 0.08,        # 中性：年化8%
            'optimistic': 0.12      # 乐观：年化12%
        }
        
        scenario_forecasts = {}
        for scenario_name, annual_return in scenarios.items():
            scenario_forecast = etf_service.forecast_portfolio_growth(
                allocation, 
                total_investment, 
                tax_rate,
                {scenario_name: annual_return}
            )
            scenario_forecasts[scenario_name] = scenario_forecast['scenarios'][scenario_name]
        
        # 组合结果
        result = {
            'portfolio': portfolio,
            'forecast_data': forecast_data,
            'scenario_forecasts': scenario_forecasts
        }
        
        return JsonResponse(result)


class UpdateExchangeRatesView(View):
    """更新汇率API"""
    def post(self, request):
        """更新汇率数据"""
        try:
            # 使用预设的汇率值（避免API限流）
            today = date.today()

            # 先删除今日的旧数据
            ExchangeRate.objects.filter(rate_date=today, data_source__in=['test_data', 'free_api']).delete()

            # 定义预设汇率
            preset_rates = [
                ('USD', 'USD', 1.0),
                ('CNY', 'CNY', 1.0),
                ('HKD', 'HKD', 1.0),
                ('CNY', 'USD', 0.138889),  # 1 CNY = 0.138889 USD
                ('HKD', 'USD', 0.128205),  # 1 HKD = 0.128205 USD
                ('USD', 'CNY', 7.2),       # 1 USD = 7.2 CNY
                ('USD', 'HKD', 7.8),       # 1 USD = 7.8 HKD
                ('CNY', 'HKD', 1.083333),  # 1 CNY = 1.083333 HKD
                ('HKD', 'CNY', 0.923077),  # 1 HKD = 0.923077 CNY
            ]

            updated_rates = []

            for from_curr, to_curr, rate in preset_rates:
                # 先删除今日该货币对的旧数据
                ExchangeRate.objects.filter(
                    from_currency=from_curr,
                    to_currency=to_curr,
                    rate_date=today
                ).delete()

                # 创建新记录
                rate_record = ExchangeRate.objects.create(
                    from_currency=from_curr,
                    to_currency=to_curr,
                    rate=rate,
                    rate_date=today,
                    data_source='system'
                )

                logger.info(f"创建汇率: 1 {from_curr} = {rate} {to_curr}")

                updated_rates.append({
                    'from_currency': from_curr,
                    'to_currency': to_curr,
                    'rate': float(rate),
                    'date': str(today)
                })

            return JsonResponse({
                'success': True,
                'rates': updated_rates,
                'update_time': timezone.now().strftime('%Y-%m-%d %H:%M:%S')
            })

        except Exception as e:
            logger.error(f"更新汇率失败: {e}", exc_info=True)
            return JsonResponse({
                'success': False,
                'message': str(e)
            }, status=500)


class UpdateRealtimeDataView(View):
    """更新实时数据API - 从yfinance获取最新数据并更新数据库"""
    def post(self, request):
        """更新实时数据并返回新的组合分析结果"""
        from .cache_manager import etf_cache
        from .operation_service import operation_service

        try:
            import yfinance as yf
            import time

            # 解析请求参数
            data = json.loads(request.body)
            allocation = data.get('allocation', {})
            total_investment = float(data.get('total_investment', 10000))

            # 获取所有启用的ETF
            enabled_etfs = ETFConfig.objects.filter(status=1).values_list('symbol', flat=True)

            # 更新每个ETF的实时数据
            update_results = []
            today = date.today()

            for symbol in enabled_etfs:
                try:
                    logger.info(f"正在获取 {symbol} 的实时数据...")

                    # 使用yfinance获取实时数据
                    ticker = yf.Ticker(symbol)
                    info = ticker.info

                    # 提取价格数据
                    current_price = info.get('regularMarketPrice') or info.get('previousClose')
                    if not current_price:
                        logger.warning(f"{symbol} 无法获取当前价格，跳过")
                        continue

                    # 获取其他价格数据
                    open_price = info.get('regularMarketOpen', current_price)
                    high_price = info.get('regularMarketDayHigh', current_price)
                    low_price = info.get('regularMarketDayLow', current_price)
                    volume = info.get('regularMarketVolume')

                    # 先删除今日旧数据
                    ETFData.objects.filter(symbol=symbol, date=today).delete()

                    # 保存到数据库中的今日数据
                    etf_data, created = ETFData.objects.create(
                        symbol=symbol,
                        date=today,
                        open_price=open_price,
                        close_price=current_price,
                        high_price=high_price,
                        low_price=low_price,
                        volume=volume,
                        data_source='yfinance_realtime'
                    )

                    update_results.append({
                        'symbol': symbol,
                        'price': float(current_price),
                        'open': float(open_price),
                        'high': float(high_price),
                        'low': float(low_price),
                        'volume': volume,
                        'success': True
                    })

                    logger.info(f"{symbol} 实时数据更新成功: ${current_price}")

                    # 添加延迟避免API限流
                    time.sleep(0.5)

                except Exception as e:
                    logger.error(f"获取 {symbol} 实时数据失败: {e}")
                    update_results.append({
                        'symbol': symbol,
                        'success': False,
                        'error': str(e)
                    })

            # 清除缓存以强制使用最新数据
            etf_cache.clear_all()

            # 重新计算投资组合分析
            portfolio = etf_service.analyze_portfolio(allocation, total_investment)

            # 统计更新结果
            success_count = sum(1 for r in update_results if r['success'])
            failed_count = len(update_results) - success_count

            # 记录操作日志
            operation_service.create_workflow_instance(
                workflow_name='实时数据更新',
                trigger_type=2,
                trigger_by='user'
            )

            return JsonResponse({
                'success': True,
                'portfolio': portfolio,
                'update_time': timezone.now().strftime('%Y-%m-%d %H:%M:%S'),
                'update_results': update_results,
                'summary': {
                    'total': len(update_results),
                    'success': success_count,
                    'failed': failed_count
                }
            })

        except Exception as e:
            logger.error(f"更新实时数据失败: {e}", exc_info=True)
            return JsonResponse({
                'success': False,
                'message': str(e),
                'update_time': timezone.now().strftime('%Y-%m-%d %H:%M:%S')
            }, status=500)


class OperationLogView(View):
    """操作记录页面"""
    def get(self, request):
        # 获取筛选参数
        selected_type = request.GET.get('type', '')
        selected_status = request.GET.get('status', '')
        per_page = int(request.GET.get('per_page', 20))
        page = int(request.GET.get('page', 1))
        
        # 构建查询
        logs_query = OperationLog.objects.all()
        
        if selected_type:
            logs_query = logs_query.filter(operation_type=selected_type)
        if selected_status:
            logs_query = logs_query.filter(status=int(selected_status))
        
        # 分页
        from django.core.paginator import Paginator
        paginator = Paginator(logs_query, per_page)
        logs = paginator.get_page(page)
        
        # 获取统计数据（近7天）
        from django.utils import timezone
        from datetime import timedelta
        seven_days_ago = timezone.now() - timedelta(days=7)
        
        stats = OperationLog.objects.filter(
            start_time__gte=seven_days_ago
        ).values('operation_type').annotate(
            count=Count('id')
        ).order_by('-count')
        
        context = {
            'logs': logs,
            'paginator': paginator,
            'stats': stats,
            'operation_types': OperationLog.OPERATION_TYPE_CHOICES,
            'status_choices': OperationLog.STATUS_CHOICES,
            'selected_type': selected_type,
            'selected_status': int(selected_status) if selected_status else '',
            'per_page': per_page,
        }
        return render(request, 'workflow/operation_logs.html', context)


class ETFConfigListView(View):
    """ETF配置列表页面"""
    def get(self, request):
        # 获取当前市场参数
        current_market = request.GET.get('market', 'US')
        
        # 获取所有配置
        all_configs = ETFConfig.objects.all()
        
        # 计算统计数据
        total_count = all_configs.count()
        us_count = all_configs.filter(market='US').count()
        cn_count = all_configs.filter(market='CN').count()
        hk_count = all_configs.filter(market='HK').count()
        active_count = all_configs.filter(status=1).count()
        
        # 根据市场筛选
        if current_market in ['US', 'CN', 'HK']:
            etf_configs = all_configs.filter(market=current_market).order_by('sort_order', 'symbol')
        else:
            etf_configs = all_configs.order_by('sort_order', 'symbol')
        
        context = {
            'total_count': total_count,
            'us_count': us_count,
            'cn_count': cn_count,
            'hk_count': hk_count,
            'active_count': active_count,
            'current_market': current_market,
            'etf_configs': etf_configs,
            'configs': etf_configs  # 保持向后兼容
        }
        return render(request, 'workflow/etf_config_list.html', context)
    
    def post(self, request):
        """添加ETF配置"""
        try:
            # 解析表单数据
            data = request.POST.dict()
            
            # 创建ETF配置
            config = ETFConfig.objects.create(
                symbol=data.get('symbol', '').upper(),
                name=data.get('name', ''),
                market=data.get('market', 'US'),
                strategy=data.get('strategy', ''),
                focus=data.get('focus', ''),
                expense_ratio=float(data.get('expense_ratio')) if data.get('expense_ratio') else None,
                description=data.get('description', ''),
                sort_order=int(data.get('sort_order', 0)),
                status=int(data.get('status', 1)),
            )
            
            # 清除ETF配置缓存
            etf_service.clear_etf_config_cache()
            
            return JsonResponse({
                'success': True,
                'message': f'ETF配置 {config.symbol} 添加成功'
            })
        except Exception as e:
            return JsonResponse({
                'success': False,
                'message': f'添加失败: {str(e)}'
            }, status=400)


class AddStepView(View):
    """添加步骤视图"""
    def post(self, request):
        """添加步骤到工作流"""
        try:
            import json
            data = json.loads(request.body)

            # 验证工作流是否存在
            workflow = get_object_or_404(Workflow, id=data.get('workflow_id'))

            # 创建步骤
            step = WorkflowStep.objects.create(
                workflow=workflow,
                name=data.get('name'),
                step_type=data.get('step_type'),
                handler_type=data.get('handler_type'),
                order_index=data.get('order_index'),
                retry_times=data.get('retry_times', 3),
                retry_interval=data.get('retry_interval', 5),
                timeout=data.get('timeout', 300),
                is_critical=data.get('is_critical', False),
                handler_config=data.get('handler_config', {})
            )

            # 记录操作日志
            from .models import OperationLog
            OperationLog.objects.create(
                workflow_instance=None,
                operation_type='manual_trigger',
                operation_name=f'添加步骤: {step.name} 到工作流 {workflow.name}',
                operator=request.META.get('REMOTE_ADDR', 'system'),
                status=1,
                output_result={'step_id': step.id}
            )

            return JsonResponse({
                'success': True,
                'message': f'步骤 {step.name} 添加成功',
                'step_id': step.id
            })

        except Workflow.DoesNotExist:
            return JsonResponse({
                'success': False,
                'message': '工作流不存在'
            }, status=404)
        except Exception as e:
            return JsonResponse({
                'success': False,
                'message': f'添加失败: {str(e)}'
            }, status=500)


class StepRetryView(View):
    """步骤重试视图"""
    def post(self, request, instance_id, step_id):
        """重试步骤"""
        try:
            # 获取步骤实例
            step_instance = get_object_or_404(WorkflowInstanceStep, id=step_id, workflow_instance_id=instance_id)

            # 只能重试失败的步骤
            if step_instance.status != 3:
                return JsonResponse({
                    'success': False,
                    'message': '只能重试失败的步骤'
                }, status=400)

            # 重置步骤状态
            step_instance.status = 0  # 等待中
            step_instance.retry_count = 0
            step_instance.error_message = None
            step_instance.start_time = None
            step_instance.end_time = None
            step_instance.duration = None
            step_instance.save()

            # 记录操作日志
            from .models import OperationLog
            op_log = OperationLog.objects.create(
                workflow_instance_id=instance_id,
                operation_type='manual_trigger',
                operation_name=f'手动重试步骤: {step_instance.step_name}',
                operator=request.META.get('REMOTE_ADDR', 'system'),
                status=0
            )

            # 异步执行步骤（这里简化为同步）
            from .engine import workflow_engine
            workflow_engine._execute_step(step_instance, step_instance.input_data.get('context', {}))

            # 完成操作日志
            op_log.complete(success=True)

            return JsonResponse({
                'success': True,
                'message': f'步骤 {step_instance.step_name} 重试成功'
            })

        except Exception as e:
            return JsonResponse({
                'success': False,
                'message': f'重试失败: {str(e)}'
            }, status=500)


class ETFConfigDetailView(View):
    """ETF配置详情页面"""
    def get(self, request, config_id):
        config = get_object_or_404(ETFConfig, id=config_id)
        
        # 如果请求的是JSON数据（API调用）
        accept_header = request.META.get('HTTP_ACCEPT', '')
        if 'application/json' in accept_header:
            return JsonResponse({
                'success': True,
                'data': {
                    'id': config.id,
                    'symbol': config.symbol,
                    'name': config.name,
                    'market': config.market,
                    'strategy': config.strategy,
                    'focus': config.focus,
                    'expense_ratio': float(config.expense_ratio) if config.expense_ratio else None,
                    'description': config.description,
                    'sort_order': config.sort_order,
                    'status': config.status,
                }
            })
        
        # 否则返回HTML页面
        context = {
            'config': config
        }
        return render(request, 'workflow/etf_detail.html', context)
    
    def put(self, request, config_id):
        """更新ETF配置"""
        try:
            config = get_object_or_404(ETFConfig, id=config_id)
            
            # 解析JSON数据
            data = json.loads(request.body)
            
            # 更新字段
            config.name = data.get('name', config.name)
            config.market = data.get('market', config.market)
            config.strategy = data.get('strategy', config.strategy)
            config.focus = data.get('focus', config.focus)
            if data.get('expense_ratio') is not None:
                config.expense_ratio = float(data['expense_ratio'])
            config.description = data.get('description', config.description)
            config.sort_order = int(data.get('sort_order', config.sort_order))
            config.status = int(data.get('status', config.status))
            config.save()
            
            # 清除ETF配置缓存
            etf_service.clear_etf_config_cache()
            
            return JsonResponse({
                'success': True,
                'message': f'ETF配置 {config.symbol} 更新成功'
            })
        except Exception as e:
            return JsonResponse({
                'success': False,
                'message': f'更新失败: {str(e)}'
            }, status=400)
    
    def patch(self, request, config_id):
        """切换ETF配置状态"""
        try:
            config = get_object_or_404(ETFConfig, id=config_id)
            
            # 切换状态
            config.status = 0 if config.status == 1 else 1
            config.save()
            
            # 清除ETF配置缓存
            etf_service.clear_etf_config_cache()
            
            return JsonResponse({
                'success': True,
                'message': f'ETF配置 {config.symbol} 状态已切换为 {"启用" if config.status == 1 else "禁用"}'
            })
        except Exception as e:
            return JsonResponse({
                'success': False,
                'message': f'操作失败: {str(e)}'
            }, status=400)
    
    def delete(self, request, config_id):
        """删除ETF配置"""
        try:
            config = get_object_or_404(ETFConfig, id=config_id)
            symbol = config.symbol
            config.delete()
            
            # 清除ETF配置缓存
            etf_service.clear_etf_config_cache()
            
            return JsonResponse({
                'success': True,
                'message': f'ETF配置 {symbol} 已删除'
            })
        except Exception as e:
            return JsonResponse({
                'success': False,
                'message': f'删除失败: {str(e)}'
            }, status=400)


# ===============================================================================
# API视图（REST API Views）
# ===============================================================================

class ETFApiView(View):
    """ETF API视图"""
    def get(self, request, *args, **kwargs):
        action = kwargs.get('action', request.GET.get('action', ''))
        
        if action == 'list':
            return self.list_etfs(request)
        elif action == 'comparison':
            return self.comparison_data(request)
        elif action == 'portfolio':
            return self.portfolio_data(request)
        elif action == 'realtime':
            symbol = kwargs.get('symbol', request.GET.get('symbol', ''))
            return self.realtime_data(request, symbol)
        elif action == 'metrics':
            symbol = kwargs.get('symbol', request.GET.get('symbol', ''))
            period = request.GET.get('period', '1y')
            return self.metrics_data(request, symbol, period)
        elif action == 'history':
            symbol = kwargs.get('symbol', request.GET.get('symbol', ''))
            period = request.GET.get('period', '1y')
            return self.history_data(request, symbol, period)
        elif action == 'forecast':
            symbol = kwargs.get('symbol', request.GET.get('symbol', ''))
            return self.forecast_data(request, symbol)
        else:
            return JsonResponse({'error': 'Invalid action'}, status=400)
    
    def list_etfs(self, request):
        """获取ETF列表"""
        etfs = etf_service.get_active_etfs()
        return JsonResponse({'etfs': etfs})
    
    def comparison_data(self, request):
        """获取ETF对比数据"""
        period = request.GET.get('period', '1y')
        comparison = etf_service.get_comparison_data(period)
        return JsonResponse({'comparison': comparison})
    
    def portfolio_data(self, request):
        """获取投资组合数据"""
        allocation_str = request.GET.get('allocation', '{}')
        try:
            allocation = json.loads(allocation_str)
        except:
            allocation = {}
        
        total_investment = float(request.GET.get('total_investment', 10000))
        tax_rate = float(request.GET.get('tax_rate', 0.10))
        
        portfolio = etf_service.analyze_portfolio(allocation, total_investment, tax_rate)
        return JsonResponse({'portfolio': portfolio})
    
    def realtime_data(self, request, symbol):
        """获取实时数据"""
        data = etf_service.fetch_realtime_data(symbol)
        return JsonResponse({'data': data})
    
    def metrics_data(self, request, symbol, period):
        """获取指标数据"""
        metrics = etf_service.calculate_metrics(symbol, period)
        return JsonResponse({'metrics': metrics})
    
    def history_data(self, request, symbol, period):
        """获取历史数据"""
        chart_data = etf_service.get_historical_chart_data(symbol, period)
        return JsonResponse({'chart_data': chart_data})
    
    def forecast_data(self, request, symbol):
        """获取收益预测数据"""
        initial_investment = float(request.GET.get('initial_investment', 10000))
        annual_return_str = request.GET.get('annual_return_rate', 'None')
        annual_return_rate = float(annual_return_str) if annual_return_str != 'None' else None
        tax_rate = float(request.GET.get('tax_rate', 0.10))
        
        forecast = etf_service.forecast_etf_growth(
            symbol, 
            initial_investment, 
            annual_return_rate, 
            tax_rate
        )
        return JsonResponse({'forecast': forecast})


class WorkflowViewSet(viewsets.ModelViewSet):
    """工作流ViewSet"""
    queryset = Workflow.objects.all()
    serializer_class = WorkflowSerializer


class WorkflowInstanceViewSet(viewsets.ModelViewSet):
    """工作流实例ViewSet"""
    queryset = WorkflowInstance.objects.all()
    serializer_class = WorkflowInstanceSerializer
    
    @action(detail=True, methods=['post'])
    def start(self, request, pk=None):
        """启动工作流实例"""
        instance = self.get_object()
        result = workflow_engine.execute_workflow(instance.workflow_id, instance.id)
        
        if result['success']:
            return Response({'message': '工作流启动成功', 'result': result}, status=status.HTTP_200_OK)
        else:
            return Response({'message': '工作流启动失败', 'result': result}, status=status.HTTP_400_BAD_REQUEST)


# ===============================================================================
# 汇率管理视图
# ===============================================================================

class ExchangeRateListView(View):
    """汇率列表页面"""
    def get(self, request):
        # 获取所有货币对
        today = timezone.now().date()
        yesterday = today - timedelta(days=1)
        
        # 今日汇率
        today_rates = ExchangeRate.objects.filter(rate_date=today)
        
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
        
        # 最近7天的汇率历史
        seven_days_ago = today - timedelta(days=6)
        recent_rates = ExchangeRate.objects.filter(
            rate_date__gte=seven_days_ago,
            rate_date__lte=today
        ).order_by('rate_date', 'from_currency', 'to_currency')
        
        # 按日期分组
        history_by_date = {}
        for rate in recent_rates:
            date_str = str(rate.rate_date)
            if date_str not in history_by_date:
                history_by_date[date_str] = []
            history_by_date[date_str].append({
                'from': rate.from_currency,
                'to': rate.to_currency,
                'rate': float(rate.rate),
            })
        
        context = {
            'rates_data': rates_data,
            'history_by_date': history_by_date,
            'today': today,
            'yesterday': yesterday,
        }
        return render(request, 'workflow/exchange_rate_list.html', context)


class ExchangeRateUpdateView(View):
    """更新汇率页面"""
    def get(self, request):
        # 手动触发更新
        try:
            result = update_exchange_rates_auto()
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
