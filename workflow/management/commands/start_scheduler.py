"""
Django management command - 启动定时任务调度器
"""

from django.core.management.base import BaseCommand
from workflow.scheduler import start_exchange_rate_job, list_jobs, get_scheduler


class Command(BaseCommand):
    help = '启动定时任务调度器，用于每天10:30自动更新汇率'

    def handle(self, *args, **options):
        self.stdout.write(self.style.SUCCESS('启动定时任务调度器...'))
        
        try:
            # 启动汇率更新定时任务
            start_exchange_rate_job()
            
            # 列出所有任务
            jobs = list_jobs()
            
            self.stdout.write(self.style.SUCCESS('=' * 60))
            self.stdout.write(self.style.SUCCESS('定时任务已配置:'))
            self.stdout.write(self.style.SUCCESS('=' * 60))
            
            for job in jobs:
                self.stdout.write(
                    f"  - {job['name']} (ID: {job['id']})"
                )
                self.stdout.write(f"    下次执行时间: {job['next_run_time']}")
                self.stdout.write(f"    触发器: {job['trigger']}")
                self.stdout.write("")
            
            self.stdout.write(self.style.SUCCESS('=' * 60))
            self.stdout.write(self.style.SUCCESS('调度器正在运行...'))
            self.stdout.write(self.style.WARNING('按 Ctrl+C 停止调度器'))
            self.stdout.write(self.style.SUCCESS('=' * 60))
            
            # 保持调度器运行
            sched = get_scheduler()
            
            try:
                # 保持主线程运行
                import time
                while True:
                    time.sleep(1)
            except KeyboardInterrupt:
                self.stdout.write("\n")
                self.stdout.write(self.style.WARNING('接收到停止信号...'))
                sched.shutdown()
                self.stdout.write(self.style.SUCCESS('调度器已停止'))
                
        except Exception as e:
            self.stdout.write(self.style.ERROR(f'启动调度器失败: {e}'))
            raise
