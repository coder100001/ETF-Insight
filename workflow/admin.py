"""
Django Admin管理界面配置
"""

from django.contrib import admin
from .models import (
    Workflow, WorkflowStep, WorkflowInstance,
    WorkflowInstanceStep, SystemLog, Notification,
    ETFData, PortfolioConfig, AnalysisReport
)


@admin.register(Workflow)
class WorkflowAdmin(admin.ModelAdmin):
    list_display = ['id', 'name', 'category', 'status', 'created_at']
    list_filter = ['status', 'category']
    search_fields = ['name', 'description']


@admin.register(WorkflowStep)
class WorkflowStepAdmin(admin.ModelAdmin):
    list_display = ['id', 'workflow', 'name', 'order_index', 'is_critical']
    list_filter = ['is_critical', 'handler_type']
    search_fields = ['name']


@admin.register(WorkflowInstance)
class WorkflowInstanceAdmin(admin.ModelAdmin):
    list_display = ['id', 'workflow', 'status', 'trigger_by', 'start_time', 'duration']
    list_filter = ['status', 'trigger_type']
    search_fields = ['trigger_by']
    date_hierarchy = 'created_at'


@admin.register(WorkflowInstanceStep)
class WorkflowInstanceStepAdmin(admin.ModelAdmin):
    list_display = ['id', 'workflow_instance', 'step_name', 'status', 'duration', 'retry_count']
    list_filter = ['status']


@admin.register(SystemLog)
class SystemLogAdmin(admin.ModelAdmin):
    list_display = ['id', 'log_level', 'module', 'message', 'created_at']
    list_filter = ['log_level', 'module']
    search_fields = ['message']
    date_hierarchy = 'created_at'


@admin.register(Notification)
class NotificationAdmin(admin.ModelAdmin):
    list_display = ['id', 'notification_type', 'recipient', 'status', 'send_at']
    list_filter = ['notification_type', 'status']


@admin.register(ETFData)
class ETFDataAdmin(admin.ModelAdmin):
    list_display = ['id', 'symbol', 'date', 'close_price', 'volume']
    list_filter = ['symbol', 'data_source']
    search_fields = ['symbol']
    date_hierarchy = 'date'


@admin.register(PortfolioConfig)
class PortfolioConfigAdmin(admin.ModelAdmin):
    list_display = ['id', 'name', 'total_investment', 'status', 'created_at']
    list_filter = ['status']
    search_fields = ['name']


@admin.register(AnalysisReport)
class AnalysisReportAdmin(admin.ModelAdmin):
    list_display = ['id', 'report_type', 'report_date', 'status', 'created_at']
    list_filter = ['report_type', 'status']
    date_hierarchy = 'report_date'
