"""
ETF组合配置管理视图和API
"""

from django.views.generic import TemplateView
from django.http import JsonResponse
from django.views.decorators.http import require_http_methods
from django.views.decorators.csrf import csrf_exempt
from django.utils.decorators import method_decorator
from django.db import transaction
from django.db.models import Count, Avg
import json
import logging

from .models import PortfolioConfig, ETFConfig
from .services import etf_service

logger = logging.getLogger(__name__)


class PortfolioConfigListView(TemplateView):
    """组合配置管理页面"""
    template_name = 'workflow/portfolio_config_list.html'
    
    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        
        # 获取所有组合配置
        configs = PortfolioConfig.objects.all().order_by('-created_at')
        
        # 计算统计数据
        total_configs = configs.count()
        active_configs = configs.filter(status=1).count()
        
        # 获取启用的ETF数量
        total_etfs = ETFConfig.objects.filter(status=1).count()
        
        # 计算平均配置ETF数量
        avg_allocation_count = 0
        if total_configs > 0:
            total_allocations = sum([len(config.allocation) for config in configs])
            avg_allocation_count = round(total_allocations / total_configs, 1)
        
        context.update({
            'configs': configs,
            'total_configs': total_configs,
            'active_configs': active_configs,
            'total_etfs': total_etfs,
            'avg_allocation_count': avg_allocation_count
        })
        
        return context


@require_http_methods(["GET"])
def get_portfolio_configs(request):
    """获取组合配置列表API"""
    try:
        configs = PortfolioConfig.objects.all().order_by('-created_at')
        
        data = [{
            'id': config.id,
            'name': config.name,
            'description': config.description,
            'total_investment': float(config.total_investment) if config.total_investment else 0,
            'allocation': config.allocation,
            'status': config.status,
            'created_at': config.created_at.isoformat(),
            'updated_at': config.updated_at.isoformat()
        } for config in configs]
        
        return JsonResponse({
            'success': True,
            'data': data
        })
    except Exception as e:
        logger.error(f"获取组合配置列表失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@require_http_methods(["GET"])
def get_portfolio_config_detail(request, config_id):
    """获取单个组合配置详情API"""
    try:
        config = PortfolioConfig.objects.get(id=config_id)
        
        # 获取配置中包含的ETF的详细信息
        etf_details = {}
        for symbol, weight in config.allocation.items():
            try:
                etf_config = ETFConfig.objects.get(symbol=symbol, status=1)
                etf_details[symbol] = {
                    'name': etf_config.name,
                    'market': etf_config.market,
                    'strategy': etf_config.strategy,
                    'expense_ratio': float(etf_config.expense_ratio) if etf_config.expense_ratio else 0
                }
            except ETFConfig.DoesNotExist:
                etf_details[symbol] = {'name': symbol, 'market': 'Unknown', 'strategy': '', 'expense_ratio': 0}
        
        data = {
            'id': config.id,
            'name': config.name,
            'description': config.description,
            'total_investment': float(config.total_investment) if config.total_investment else 0,
            'allocation': config.allocation,
            'status': config.status,
            'etf_details': etf_details,
            'created_at': config.created_at.isoformat(),
            'updated_at': config.updated_at.isoformat()
        }
        
        return JsonResponse({
            'success': True,
            'data': data
        })
    except PortfolioConfig.DoesNotExist:
        return JsonResponse({
            'success': False,
            'error': '配置不存在'
        }, status=404)
    except Exception as e:
        logger.error(f"获取组合配置详情失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@csrf_exempt
@require_http_methods(["POST", "PUT", "PATCH"])
def save_portfolio_config(request, config_id=None):
    """保存或更新组合配置API"""
    try:
        data = json.loads(request.body)
        
        # 验证必填字段
        name = data.get('name', '').strip()
        if not name:
            return JsonResponse({
                'success': False,
                'error': '组合名称不能为空'
            }, status=400)
        
        # 验证投资金额
        total_investment = data.get('total_investment', 0)
        if not total_investment or total_investment < 100:
            return JsonResponse({
                'success': False,
                'error': '投资金额必须至少为100美元'
            }, status=400)
        
        # 验证权重配置
        allocation = data.get('allocation', {})
        if not allocation:
            return JsonResponse({
                'success': False,
                'error': '权重配置不能为空'
            }, status=400)
        
        # 验证权重总和
        total_weight = sum(allocation.values())
        if abs(total_weight - 1.0) > 0.01:  # 允许0.01的误差
            return JsonResponse({
                'success': False,
                'error': f'权重总和应为100%，当前为{total_weight * 100:.1f}%'
            }, status=400)
        
        # 验证ETF是否存在
        valid_symbols = set(ETFConfig.objects.filter(status=1).values_list('symbol', flat=True))
        invalid_symbols = set(allocation.keys()) - valid_symbols
        if invalid_symbols:
            return JsonResponse({
                'success': False,
                'error': f'以下ETF不存在或已禁用: {", ".join(invalid_symbols)}'
            }, status=400)
        
        # 创建或更新配置
        with transaction.atomic():
            if config_id:
                # 更新现有配置
                config = PortfolioConfig.objects.get(id=config_id)
                config.name = name
                config.description = data.get('description', '')
                config.total_investment = total_investment
                config.allocation = allocation
                if 'status' in data:
                    config.status = data['status']
                config.save()
                
                logger.info(f"更新组合配置: {config.name} (ID: {config.id})")
            else:
                # 创建新配置
                config = PortfolioConfig.objects.create(
                    name=name,
                    description=data.get('description', ''),
                    total_investment=total_investment,
                    allocation=allocation,
                    status=data.get('status', 1)
                )
                
                logger.info(f"创建组合配置: {config.name} (ID: {config.id})")
            
            # 获取更新后的配置数据
            config_data = {
                'id': config.id,
                'name': config.name,
                'description': config.description,
                'total_investment': float(config.total_investment),
                'allocation': config.allocation,
                'status': config.status,
                'created_at': config.created_at.isoformat(),
                'updated_at': config.updated_at.isoformat()
            }
            
            return JsonResponse({
                'success': True,
                'data': config_data
            })
    
    except PortfolioConfig.DoesNotExist:
        return JsonResponse({
            'success': False,
            'error': '配置不存在'
        }, status=404)
    except Exception as e:
        logger.error(f"保存组合配置失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@csrf_exempt
@require_http_methods(["DELETE"])
def delete_portfolio_config(request, config_id):
    """删除组合配置API"""
    try:
        config = PortfolioConfig.objects.get(id=config_id)
        config_name = config.name
        config.delete()
        
        logger.info(f"删除组合配置: {config_name} (ID: {config_id})")
        
        return JsonResponse({
            'success': True,
            'message': '配置删除成功'
        })
    except PortfolioConfig.DoesNotExist:
        return JsonResponse({
            'success': False,
            'error': '配置不存在'
        }, status=404)
    except Exception as e:
        logger.error(f"删除组合配置失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@csrf_exempt
@require_http_methods(["POST"])
def toggle_portfolio_config_status(request, config_id):
    """切换组合配置状态API"""
    try:
        config = PortfolioConfig.objects.get(id=config_id)
        config.status = 1 if config.status == 0 else 0
        config.save()
        
        status_text = '启用' if config.status == 1 else '禁用'
        logger.info(f"{status_text}组合配置: {config.name} (ID: {config_id})")
        
        return JsonResponse({
            'success': True,
            'data': {
                'id': config.id,
                'status': config.status,
                'status_text': status_text
            }
        })
    except PortfolioConfig.DoesNotExist:
        return JsonResponse({
            'success': False,
            'error': '配置不存在'
        }, status=404)
    except Exception as e:
        logger.error(f"切换组合配置状态失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@csrf_exempt
@require_http_methods(["POST"])
def analyze_portfolio_from_config(request):
    """根据配置分析投资组合API"""
    try:
        data = json.loads(request.body)
        allocation = data.get('allocation', {})
        total_investment = data.get('total_investment', 10000)
        tax_rate = data.get('tax_rate', 0.10)
        
        if not allocation:
            return JsonResponse({
                'success': False,
                'error': '权重配置不能为空'
            }, status=400)
        
        # 调用投资组合分析服务
        result = etf_service.analyze_portfolio(
            allocation=allocation,
            total_investment=total_investment,
            tax_rate=tax_rate
        )
        
        return JsonResponse({
            'success': True,
            'data': result
        })
    
    except Exception as e:
        logger.error(f"分析投资组合失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@require_http_methods(["GET"])
def get_portfolio_performance(request, config_id):
    """获取投资组合性能指标API"""
    try:
        config = PortfolioConfig.objects.get(id=config_id)
        
        # 调用投资组合分析服务
        result = etf_service.analyze_portfolio(
            allocation=config.allocation,
            total_investment=config.total_investment or 10000
        )
        
        # 提取关键性能指标
        performance_metrics = {
            'total_value': result.get('total_value', 0),
            'total_return': result.get('total_return', 0),
            'total_return_percent': result.get('total_return_percent', 0),
            'total_dividend': result.get('total_dividend', 0),
            'weighted_dividend_yield': result.get('weighted_dividend_yield', 0),
            'capital_gains': result.get('capital_gains', 0),
            'tax_amount': result.get('tax_amount', 0),
            'net_return': result.get('net_return', 0),
            'holdings_count': len(result.get('holdings', []))
        }
        
        return JsonResponse({
            'success': True,
            'data': performance_metrics
        })
    
    except PortfolioConfig.DoesNotExist:
        return JsonResponse({
            'success': False,
            'error': '配置不存在'
        }, status=404)
    except Exception as e:
        logger.error(f"获取投资组合性能指标失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@require_http_methods(["GET"])
def get_etf_configs(request):
    """获取ETF配置列表API"""
    try:
        configs = ETFConfig.objects.all().order_by('sort_order', 'symbol')
        
        data = [{
            'id': config.id,
            'symbol': config.symbol,
            'name': config.name,
            'market': config.market,
            'strategy': config.strategy,
            'description': config.description,
            'focus': config.focus,
            'expense_ratio': float(config.expense_ratio) if config.expense_ratio else 0,
            'status': config.status,
            'sort_order': config.sort_order
        } for config in configs]
        
        return JsonResponse({
            'success': True,
            'data': data
        })
    except Exception as e:
        logger.error(f"获取ETF配置列表失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)


@require_http_methods(["GET"])
def get_portfolio_stats(request):
    """获取组合配置统计数据API"""
    try:
        configs = PortfolioConfig.objects.all()
        
        # 基本统计
        total_configs = configs.count()
        active_configs = configs.filter(status=1).count()
        
        # 投资金额统计
        investment_stats = configs.aggregate(
            avg_investment=Avg('total_investment'),
            total_investment=Count('id') * Avg('total_investment')
        )
        
        # 配置ETF数量统计
        allocation_counts = [len(config.allocation) for config in configs]
        avg_etf_per_config = round(sum(allocation_counts) / len(allocation_counts), 1) if allocation_counts else 0
        
        # 最受欢迎的ETF
        etf_popularity = {}
        for config in configs:
            for symbol in config.allocation.keys():
                etf_popularity[symbol] = etf_popularity.get(symbol, 0) + 1
        
        top_etfs = sorted(etf_popularity.items(), key=lambda x: x[1], reverse=True)[:5]
        
        return JsonResponse({
            'success': True,
            'data': {
                'total_configs': total_configs,
                'active_configs': active_configs,
                'inactive_configs': total_configs - active_configs,
                'avg_investment': round(float(investment_stats['avg_investment'] or 0), 2),
                'avg_etf_per_config': avg_etf_per_config,
                'top_etfs': [{'symbol': symbol, 'count': count} for symbol, count in top_etfs]
            }
        })
    except Exception as e:
        logger.error(f"获取组合统计数据失败: {e}")
        return JsonResponse({
            'success': False,
            'error': str(e)
        }, status=500)
