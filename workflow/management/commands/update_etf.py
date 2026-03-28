"""
Django管理命令：更新ETF数据

用法:
    python manage.py update_etf                    # 更新所有启用的ETF
    python manage.py update_etf SCHD SPYD          # 更新指定的ETF
    python manage.py update_etf --all              # 更新所有启用的ETF
"""

from django.core.management.base import BaseCommand
import logging

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = '更新ETF数据（支持指定特定ETF）'

    def add_arguments(self, parser):
        parser.add_argument(
            'symbols',
            nargs='*',
            type=str,
            help='要更新的ETF代码（如SCHD、SPYD等），不指定则更新所有'
        )
        parser.add_argument(
            '--all',
            action='store_true',
            help='更新所有启用的ETF'
        )

    def handle(self, *args, **options):
        from workflow.scheduler import update_etf_data
        from workflow.services import etf_service
        
        symbols = options.get('symbols', [])
        update_all = options.get('all', False)
        
        # 确定要更新的ETF列表
        if update_all or not symbols:
            # 更新所有启用的ETF
            symbols_to_update = None  # None表示更新所有
            self.stdout.write(self.style.SUCCESS('准备更新所有启用的ETF...'))
        else:
            # 更新指定的ETF
            symbols_to_update = [s.upper() for s in symbols]
            self.stdout.write(self.style.SUCCESS(f'准备更新指定的ETF: {", ".join(symbols_to_update)}'))
        
        # 执行更新
        try:
            success = update_etf_data(symbols=symbols_to_update)
            
            if success:
                self.stdout.write(self.style.SUCCESS('✓ ETF数据更新成功！'))
            else:
                self.stdout.write(self.style.ERROR('✗ ETF数据更新失败，请查看日志'))
                
        except Exception as e:
            self.stdout.write(self.style.ERROR(f'✗ 更新过程中出错: {str(e)}'))
            logger.exception('ETF数据更新失败')
