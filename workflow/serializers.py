"""
ETF工作流系统 - Django Serializer层
"""

from rest_framework import serializers
from .models import (
    Workflow, WorkflowStep, WorkflowInstance, 
    WorkflowInstanceStep, SystemLog, Notification,
    ETFData, PortfolioConfig, AnalysisReport
)


class WorkflowStepSerializer(serializers.ModelSerializer):
    """工作流步骤序列化器"""
    class Meta:
        model = WorkflowStep
        fields = '__all__'


class WorkflowSerializer(serializers.ModelSerializer):
    """工作流序列化器"""
    steps = WorkflowStepSerializer(many=True, read_only=True)
    step_count = serializers.SerializerMethodField()
    
    class Meta:
        model = Workflow
        fields = '__all__'
    
    def get_step_count(self, obj):
        return obj.steps.count()


class WorkflowInstanceStepSerializer(serializers.ModelSerializer):
    """工作流实例步骤序列化器"""
    status_display = serializers.CharField(source='get_status_display', read_only=True)
    
    class Meta:
        model = WorkflowInstanceStep
        fields = '__all__'


class WorkflowInstanceSerializer(serializers.ModelSerializer):
    """工作流实例序列化器"""
    workflow_name = serializers.CharField(source='workflow.name', read_only=True)
    status_display = serializers.CharField(source='get_status_display', read_only=True)
    step_instances = WorkflowInstanceStepSerializer(many=True, read_only=True)
    
    class Meta:
        model = WorkflowInstance
        fields = '__all__'


class SystemLogSerializer(serializers.ModelSerializer):
    """系统日志序列化器"""
    class Meta:
        model = SystemLog
        fields = '__all__'


class NotificationSerializer(serializers.ModelSerializer):
    """通知序列化器"""
    class Meta:
        model = Notification
        fields = '__all__'


class ETFDataSerializer(serializers.ModelSerializer):
    """ETF数据序列化器"""
    class Meta:
        model = ETFData
        fields = '__all__'


class PortfolioConfigSerializer(serializers.ModelSerializer):
    """投资组合配置序列化器"""
    class Meta:
        model = PortfolioConfig
        fields = '__all__'


class AnalysisReportSerializer(serializers.ModelSerializer):
    """分析报告序列化器"""
    class Meta:
        model = AnalysisReport
        fields = '__all__'
