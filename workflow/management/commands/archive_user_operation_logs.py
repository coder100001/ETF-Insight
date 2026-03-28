"""
用户操作日志管理命令 - 归档和清理
"""
from django.core.management.base import BaseCommand
from django.utils import timezone
from datetime import timedelta
import logging

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = '归档旧的操作日志到历史表'

    def add_arguments(self, parser):
        parser.add_argument(
            '--days',
            type=int,
            default=90,
            help='归档多少天前的日志（默认90天）'
        )
        parser.add_argument(
            '--delete',
            action='store_true',
            help='归档后删除主表中的日志'
        )
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='只显示将要归档的日志数量，不实际执行'
        )

    def handle(self, *args, **options):
        from workflow.models import UserOperationLog
        
        days = options['days']
        delete_after_archive = options['delete']
        dry_run = options['dry_run']
        
        # 计算归档日期
        archive_date = timezone.now() - timedelta(days=days)
        
        # 查询需要归档的日志
        logs_to_archive = UserOperationLog.objects.filter(
            operation_time__lt=archive_date,
            is_deleted=False,
            archived_at__isnull=True
        )
        
        count = logs_to_archive.count()
        
        if dry_run:
            self.stdout.write(
                self.style.WARNING(
                    f'[DRY RUN] 将归档 {count} 条日志（操作时间早于 {archive_date.date()}）'
                )
            )
            return
        
        if count == 0:
            self.stdout.write(self.style.SUCCESS('没有需要归档的日志'))
            return
        
        # 执行归档
        self.stdout.write(f'开始归档 {count} 条日志...')
        
        # 更新归档时间
        updated = logs_to_archive.update(archived_at=timezone.now())
        
        self.stdout.write(
            self.style.SUCCESS(
                f'✓ 成功归档 {updated} 条日志'
            )
        )
        
        # 如果需要，删除主表中的已归档日志
        if delete_after_archive:
            deleted_count = logs_to_archive.update(is_deleted=True)
            self.stdout.write(
                self.style.WARNING(
                    f'⚠ 已标记 {deleted_count} 条归档日志为删除'
                )
            )
        
        # 显示统计信息
        self.stdout.write('\n归档统计：')
        self.stdout.write(f'  归档天数：{days}天')
        self.stdout.write(f'  归档日期：{archive_date.date()}')
        self.stdout.write(f'  归档数量：{updated}')
        self.stdout.write(f'  剩余活跃日志：{UserOperationLog.objects.filter(is_deleted=False, archived_at__isnull=True).count()}')
