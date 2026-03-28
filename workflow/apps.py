from django.apps import AppConfig
import os


class WorkflowConfig(AppConfig):
    default_auto_field = 'django.db.models.BigAutoField'
    name = 'workflow'
    
    def ready(self):
        """Django应用启动时执行"""
        # 避免在manage.py命令或重载时重复启动
        if os.environ.get('RUN_MAIN') == 'true':
            from .scheduler import start_scheduler
            start_scheduler()
