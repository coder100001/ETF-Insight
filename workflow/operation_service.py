"""
操作记录服务 - 记录系统操作和工作流执行
"""

import logging
from django.utils import timezone
from contextlib import contextmanager

logger = logging.getLogger(__name__)


class OperationLogService:
    """操作记录服务"""
    
    @staticmethod
    def log_operation(operation_type, operation_name, operator=None, 
                      input_params=None, ip_address=None, user_agent=None,
                      workflow_instance=None, extra_data=None):
        """
        记录操作开始
        返回OperationLog对象，可用于后续完成操作
        """
        from .models import OperationLog
        
        log = OperationLog.objects.create(
            workflow_instance=workflow_instance,
            operation_type=operation_type,
            operation_name=operation_name,
            operator=operator,
            status=0,  # 进行中
            input_params=input_params,
            ip_address=ip_address,
            user_agent=user_agent,
            extra_data=extra_data
        )
        logger.info(f'[操作记录] 开始: {operation_name} (ID: {log.id})')
        return log
    
    @staticmethod
    def complete_operation(log, success=True, result=None, error=None):
        """完成操作记录"""
        if log:
            log.complete(success=success, result=result, error=error)
            status = '成功' if success else '失败'
            logger.info(f'[操作记录] {status}: {log.operation_name} (ID: {log.id}, 耗时: {log.duration_ms}ms)')
    
    @staticmethod
    @contextmanager
    def track_operation(operation_type, operation_name, operator=None,
                       input_params=None, ip_address=None, user_agent=None,
                       workflow_instance=None, extra_data=None):
        """
        上下文管理器 - 自动跟踪操作
        用法:
            with operation_service.track_operation('data_update', '更新SPYD数据') as log:
                # 执行操作
                result = do_something()
                log.output_result = result
        """
        log = OperationLogService.log_operation(
            operation_type=operation_type,
            operation_name=operation_name,
            operator=operator,
            input_params=input_params,
            ip_address=ip_address,
            user_agent=user_agent,
            workflow_instance=workflow_instance,
            extra_data=extra_data
        )
        
        try:
            yield log
            OperationLogService.complete_operation(log, success=True)
        except Exception as e:
            OperationLogService.complete_operation(log, success=False, error=str(e))
            raise
    
    @staticmethod
    def create_workflow_instance(workflow_name, trigger_type=1, trigger_by='system'):
        """创建工作流实例"""
        from .models import Workflow, WorkflowInstance
        
        # 获取或创建工作流定义
        workflow, created = Workflow.objects.get_or_create(
            name=workflow_name,
            defaults={
                'description': f'{workflow_name}自动创建',
                'category': 'ETF数据',
                'status': 1,
                'trigger_type': trigger_type,
            }
        )
        
        # 创建实例
        instance = WorkflowInstance.objects.create(
            workflow=workflow,
            trigger_type=trigger_type,
            trigger_by=trigger_by,
            status=1,  # 运行中
            start_time=timezone.now()
        )
        
        logger.info(f'[工作流] 创建实例: {workflow_name} (ID: {instance.id})')
        return instance
    
    @staticmethod
    def complete_workflow_instance(instance, success=True, error_message=None, context_data=None):
        """完成工作流实例"""
        if instance:
            instance.end_time = timezone.now()
            instance.status = 2 if success else 3  # 2=成功, 3=失败
            if instance.start_time:
                instance.duration = int((instance.end_time - instance.start_time).total_seconds())
            if error_message:
                instance.error_message = error_message
            if context_data:
                instance.context_data = context_data
            instance.save()
            
            status = '成功' if success else '失败'
            logger.info(f'[工作流] 完成: {instance.workflow.name} (ID: {instance.id}, 状态: {status})')
    
    @staticmethod
    def get_recent_logs(operation_type=None, limit=50):
        """获取最近的操作记录"""
        from .models import OperationLog
        
        qs = OperationLog.objects.all()
        if operation_type:
            qs = qs.filter(operation_type=operation_type)
        return qs[:limit]
    
    @staticmethod
    def get_operation_stats(days=7):
        """获取操作统计"""
        from .models import OperationLog
        from django.db.models import Count
        from datetime import timedelta
        
        start_date = timezone.now() - timedelta(days=days)
        
        stats = OperationLog.objects.filter(
            start_time__gte=start_date
        ).values('operation_type').annotate(
            count=Count('id')
        ).order_by('-count')
        
        return list(stats)


# 全局服务实例
operation_service = OperationLogService()
