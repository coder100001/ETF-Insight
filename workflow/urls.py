"""
工作流应用 - URL配置
"""

from django.urls import path, include
from rest_framework.routers import DefaultRouter
from . import views
from . import views_exchange_rate
from . import portfolio_views
from .views import ETFApiView

# REST API路由
router = DefaultRouter()
router.register(r'workflows', views.WorkflowViewSet)
router.register(r'instances', views.WorkflowInstanceViewSet)

app_name = 'workflow'

urlpatterns = [
    # 工作流页面路由
    path('', views.DashboardView.as_view(), name='dashboard'),
    path('dashboard/', views.DashboardView.as_view(), name='dashboard'),
    path('workflows/', views.WorkflowListView.as_view(), name='workflow_list'),
    path('workflow/<int:workflow_id>/', views.WorkflowDetailView.as_view(), name='workflow_detail'),
    path('instances/', views.InstanceListView.as_view(), name='instance_list'),
    path('instance/<int:instance_id>/', views.InstanceDetailView.as_view(), name='instance_detail'),
    path('instance/<int:instance_id>/step/<int:step_id>/retry/', views.StepRetryView.as_view(), name='step_retry'),
    path('add-step/', views.AddStepView.as_view(), name='add_step'),
    
    # ETF分析页面路由
    path('etf/', views.ETFDashboardView.as_view(), name='etf_dashboard'),
    path('etf/<str:symbol>/', views.ETFDetailView.as_view(), name='etf_detail'),
    path('etf-comparison/', views.ETFComparisonView.as_view(), name='etf_comparison'),
    path('etf-forecast/', views.ETFForecastView.as_view(), name='etf_forecast'),  # ETF预测页面
    path('portfolio/', views.PortfolioAnalysisView.as_view(), name='portfolio_analysis'),
    # 为兼容性添加同名URL（解决某些地方调用 reverse('portfolio') 的问题）
    path('portfolio/', views.PortfolioAnalysisView.as_view(), name='portfolio'),
    
    # 操作记录页面
    path('logs/', views.OperationLogView.as_view(), name='operation_logs'),
    
    # ETF配置管理
    path('etf-config/', views.ETFConfigListView.as_view(), name='etf_config_list'),
    path('etf-config/<int:config_id>/', views.ETFConfigDetailView.as_view(), name='etf_config_detail'),
    
    # 汇率管理
    path('exchange-rates/', views_exchange_rate.ExchangeRateListView.as_view(), name='exchange_rate_list'),
    path('exchange-rates/update/', views_exchange_rate.ExchangeRateUpdateView.as_view(), name='exchange_rate_update'),
    path('exchange-rates/history/', views_exchange_rate.ExchangeRateHistoryView.as_view(), name='exchange_rate_history'),
    path('exchange-rates/convert/', views_exchange_rate.ExchangeRateConvertView.as_view(), name='exchange_rate_convert'),
    
    # ETF API路由
    path('api/etf/list/', ETFApiView.as_view(), {'action': 'list'}, name='api_etf_list'),
    path('api/etf/portfolio/analyze/', ETFApiView.as_view(), {'action': 'portfolio_analyze'}, name='api_etf_portfolio_analyze'),
    path('api/etf/comparison/', ETFApiView.as_view(), {'action': 'comparison'}, name='api_etf_comparison'),
    path('api/etf/portfolio/', ETFApiView.as_view(), {'action': 'portfolio'}, name='api_etf_portfolio'),
    path('api/etf/<str:symbol>/realtime/', ETFApiView.as_view(), {'action': 'realtime'}, name='api_etf_realtime'),
    path('api/etf/<str:symbol>/metrics/', ETFApiView.as_view(), {'action': 'metrics'}, name='api_etf_metrics'),
    path('api/etf/<str:symbol>/history/', ETFApiView.as_view(), {'action': 'history'}, name='api_etf_history'),
    path('api/etf/<str:symbol>/forecast/', ETFApiView.as_view(), {'action': 'forecast'}, name='api_etf_forecast'),
    path('api/update-realtime/', views.UpdateRealtimeDataView.as_view(), name='api_update_realtime'),
    path('api/update-exchange-rates/', views.UpdateExchangeRatesView.as_view(), name='api_update_exchange_rates'),

    # 投资组合配置管理路由
    path('portfolio-config/', portfolio_views.PortfolioConfigListView.as_view(), name='portfolio_config_list'),

    # 工作流API路由
    path('api/', include(router.urls)),

    # 投资组合配置API路由
    path('api/portfolio-configs/', portfolio_views.get_portfolio_configs),
    path('api/portfolio-configs/<int:config_id>/', portfolio_views.save_portfolio_config),
    path('api/portfolio-configs/<int:config_id>/detail/', portfolio_views.get_portfolio_config_detail),
    path('api/portfolio-configs/<int:config_id>/toggle-status/', portfolio_views.toggle_portfolio_config_status),
    path('api/portfolio-configs/<int:config_id>/performance/', portfolio_views.get_portfolio_performance),
    path('api/analyze-portfolio/', portfolio_views.analyze_portfolio_from_config),
    path('api/etf-configs/', portfolio_views.get_etf_configs),
    path('api/portfolio-stats/', portfolio_views.get_portfolio_stats),
]