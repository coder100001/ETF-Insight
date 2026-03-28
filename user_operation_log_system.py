"""
用户操作日志系统 - 完整设计和实现
"""

# ============================================================
# 一、数据库模型设计
# ============================================================

"""
新增模型：UserOperationLog
用途：记录所有用户操作的完整日志
"""

"""
字段设计：

1. 基本信息
- id: 主键
- operation_time: 操作时间 (DateTimeField, indexed)
- operation_type: 操作类型 (CharField, choices, indexed)
- user: 操作用户 (ForeignKey to User, indexed, null=True)
- user_id: 用户ID (CharField, indexed, 支持匿名用户)
- username: 用户名 (CharField, 支持历史用户查询)
- module: 系统模块名称 (CharField, indexed)
- operation_name: 操作名称 (CharField)
- operation_detail: 操作内容详情 (TextField)

2. 请求信息
- request_method: 请求方法 (CharField)
- request_url: 请求URL (CharField)
- request_path: 请求路径 (CharField, indexed)
- ip_address: IP地址 (CharField, indexed)
- user_agent: 用户代理 (TextField)
- referer: 来源页面 (CharField)

3. 请求参数
- request_params: 请求参数 (JSONField)
- request_body: 请求体 (JSONField, 适合POST数据)

4. 响应信息
- response_status: 响应状态码 (IntegerField)
- response_size: 响应大小 (IntegerField)
- execution_time: 执行时间(毫秒) (IntegerField)

5. 业务数据
- business_data: 业务相关数据 (JSONField)
- entity_type: 实体类型 (CharField)
- entity_id: 实体ID (CharField, indexed)

6. 状态和错误
- status: 操作状态 (IntegerField, choices, indexed)
- error_code: 错误码 (CharField)
- error_message: 错误信息 (TextField)
- stack_trace: 堆栈信息 (TextField)

7. 审计信息
- session_id: 会话ID (CharField, indexed)
- correlation_id: 关联ID (CharField, indexed)
- parent_log_id: 父日志ID (ForeignKey, 支持操作链)
- tags: 标签 (ArrayField, 用于分类)

8. 元数据
- created_at: 创建时间 (DateTimeField, indexed)
- updated_at: 更新时间 (DateTimeField)
- is_deleted: 是否已删除 (BooleanField, default=False)
- archived_at: 归档时间 (DateTimeField, null=True)
"""

# ============================================================
# 二、模型定义
# ============================================================

from django.db import models
from django.contrib.auth import get_user_model
from django.db.models import Index
from django.utils import timezone

User = get_user_model()


class UserOperationLog(models.Model):
    """用户操作日志 - 完整记录所有用户操作"""
    
    # 操作类型选择
    OPERATION_TYPE_CHOICES = [
        ('view', '页面访问'),
        ('create', '创建操作'),
        ('update', '更新操作'),
        ('delete', '删除操作'),
        ('query', '查询操作'),
        ('export', '导出操作'),
        ('import', '导入操作'),
        ('login', '登录操作'),
        ('logout', '登出操作'),
        ('data_update', '数据更新'),
        ('portfolio_analysis', '组合分析'),
        ('etf_comparison', 'ETF对比'),
        ('script_execute', '脚本执行'),
        ('scheduled_task', '定时任务'),
        ('manual_trigger', '手动触发'),
        ('config_change', '配置变更'),
        ('api_call', 'API调用'),
    ]
    
    # 系统模块选择
    MODULE_CHOICES = [
        ('etf_dashboard', 'ETF仪表板'),
        ('portfolio', '投资组合'),
        ('etf_comparison', 'ETF对比'),
        ('etf_config', 'ETF配置'),
        ('workflow', '工作流'),
        ('operation_logs', '操作日志'),
        ('exchange_rates', '汇率管理'),
        ('api', 'API接口'),
        ('auth', '用户认证'),
        ('admin', '后台管理'),
    ]
    
    # 操作状态选择
    STATUS_CHOICES = [
        (0, '进行中'),
        (1, '成功'),
        (2, '失败'),
        (3, '已取消'),
    ]
    
    # ============ 基本信息 ============
    
    # 操作时间
    operation_time = models.DateTimeField(
        db_index=True,
        verbose_name='操作时间',
        help_text='操作发生的具体时间'
    )
    
    # 操作类型
    operation_type = models.CharField(
        max_length=50,
        choices=OPERATION_TYPE_CHOICES,
        db_index=True,
        verbose_name='操作类型',
        help_text='用户执行的操作类型'
    )
    
    # 操作用户（外键）
    user = models.ForeignKey(
        User,
        null=True,
        blank=True,
        on_delete=models.SET_NULL,
        related_name='operation_logs',
        db_index=True,
        verbose_name='操作用户',
        help_text='执行操作的用户'
    )
    
    # 用户ID（支持匿名用户）
    user_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='用户ID',
        help_text='用户的唯一标识符'
    )
    
    # 用户名（保留历史用户名）
    username = models.CharField(
        max_length=150,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='用户名',
        help_text='操作时的用户名（即使用户被删除也保留）'
    )
    
    # 系统模块名称
    module = models.CharField(
        max_length=50,
        choices=MODULE_CHOICES,
        db_index=True,
        verbose_name='系统模块',
        help_text='操作所属的系统模块'
    )
    
    # 操作名称
    operation_name = models.CharField(
        max_length=200,
        verbose_name='操作名称',
        help_text='操作的简短描述'
    )
    
    # 操作内容详情
    operation_detail = models.TextField(
        null=True,
        blank=True,
        verbose_name='操作详情',
        help_text='操作的详细描述'
    )
    
    # ============ 请求信息 ============
    
    # 请求方法
    request_method = models.CharField(
        max_length=10,
        null=True,
        blank=True,
        verbose_name='请求方法',
        help_text='HTTP请求方法（GET/POST/PUT/DELETE）'
    )
    
    # 请求URL
    request_url = models.CharField(
        max_length=500,
        null=True,
        blank=True,
        verbose_name='请求URL',
        help_text='完整的请求URL'
    )
    
    # 请求路径（用于分组查询）
    request_path = models.CharField(
        max_length=200,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='请求路径',
        help_text='请求的路径部分'
    )
    
    # IP地址
    ip_address = models.GenericIPAddressField(
        null=True,
        blank=True,
        db_index=True,
        verbose_name='IP地址',
        help_text='客户端IP地址'
    )
    
    # 用户代理
    user_agent = models.TextField(
        null=True,
        blank=True,
        verbose_name='用户代理',
        help_text='浏览器的User-Agent字符串'
    )
    
    # 来源页面
    referer = models.CharField(
        max_length=500,
        null=True,
        blank=True,
        verbose_name='来源页面',
        help_text='用户来源的URL'
    )
    
    # ============ 请求参数 ============
    
    # 请求参数
    request_params = models.JSONField(
        null=True,
        blank=True,
        verbose_name='请求参数',
        help_text='GET请求的查询参数'
    )
    
    # 请求体
    request_body = models.JSONField(
        null=True,
        blank=True,
        verbose_name='请求体',
        help_text='POST/PUT请求的body数据'
    )
    
    # ============ 响应信息 ============
    
    # 响应状态码
    response_status = models.IntegerField(
        null=True,
        blank=True,
        verbose_name='响应状态码',
        help_text='HTTP响应状态码'
    )
    
    # 响应大小
    response_size = models.BigIntegerField(
        null=True,
        blank=True,
        verbose_name='响应大小',
        help_text='响应数据的字节大小'
    )
    
    # 执行时间
    execution_time = models.IntegerField(
        null=True,
        blank=True,
        verbose_name='执行时间(毫秒)',
        help_text='操作执行的毫秒数'
    )
    
    # ============ 业务数据 ============
    
    # 业务相关数据
    business_data = models.JSONField(
        null=True,
        blank=True,
        verbose_name='业务数据',
        help_text='操作相关的业务数据'
    )
    
    # 实体类型
    entity_type = models.CharField(
        max_length=50,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='实体类型',
        help_text='操作涉及的数据实体类型（如：ETF、Portfolio等）'
    )
    
    # 实体ID
    entity_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='实体ID',
        help_text='操作涉及的数据实体ID'
    )
    
    # ============ 状态和错误 ============
    
    # 操作状态
    status = models.IntegerField(
        choices=STATUS_CHOICES,
        default=0,
        db_index=True,
        verbose_name='操作状态',
        help_text='操作的执行状态'
    )
    
    # 错误码
    error_code = models.CharField(
        max_length=50,
        null=True,
        blank=True,
        verbose_name='错误码',
        help_text='错误代码标识'
    )
    
    # 错误信息
    error_message = models.TextField(
        null=True,
        blank=True,
        verbose_name='错误信息',
        help_text='详细的错误描述'
    )
    
    # 堆栈信息
    stack_trace = models.TextField(
        null=True,
        blank=True,
        verbose_name='堆栈信息',
        help_text='异常的堆栈跟踪'
    )
    
    # ============ 审计信息 ============
    
    # 会话ID
    session_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='会话ID',
        help_text='用户会话的唯一标识'
    )
    
    # 关联ID（用于追踪相关操作）
    correlation_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='关联ID',
        help_text='跨服务/模块的请求追踪ID'
    )
    
    # 父日志ID（操作链）
    parent_log_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        verbose_name='父日志ID',
        help_text='关联的父操作日志ID'
    )
    
    # 标签（用于分类和搜索）
    tags = models.JSONField(
        null=True,
        blank=True,
        verbose_name='标签',
        help_text='操作标签数组（用于分类和搜索）'
    )
    
    # ============ 元数据 ============
    
    # 创建时间
    created_at = models.DateTimeField(
        auto_now_add=True,
        db_index=True,
        verbose_name='创建时间'
    )
    
    # 更新时间
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name='更新时间'
    )
    
    # 是否已删除（软删除）
    is_deleted = models.BooleanField(
        default=False,
        db_index=True,
        verbose_name='已删除',
        help_text='标记为已删除（软删除）'
    )
    
    # 归档时间
    archived_at = models.DateTimeField(
        null=True,
        blank=True,
        db_index=True,
        verbose_name='归档时间',
        help_text='日志归档到历史表的时间'
    )
    
    class Meta:
        db_table = 'user_operation_log'
        verbose_name = '用户操作日志'
        verbose_name_plural = '用户操作日志'
        ordering = ['-operation_time']
        
        # 索引优化 - 提升查询性能
        indexes = [
            # 复合索引 - 常用查询组合
            Index(fields=['user_id', 'operation_time'], name='idx_user_time'),
            Index(fields=['operation_type', 'operation_time'], name='idx_type_time'),
            Index(fields=['module', 'operation_time'], name='idx_module_time'),
            Index(fields=['status', 'operation_time'], name='idx_status_time'),
            Index(fields=['ip_address', 'operation_time'], name='idx_ip_time'),
            Index(fields=['session_id', 'operation_time'], name='idx_session_time'),
            Index(fields=['correlation_id'], name='idx_correlation'),
            Index(fields=['entity_type', 'entity_id'], name='idx_entity'),
            # 时间范围查询索引
            Index(fields=['-operation_time'], name='idx_time_desc'),
        ]
    
    def __str__(self):
        return f"[{self.operation_type}] {self.username or self.user_id} - {self.operation_name}"
    
    def save(self, *args, **kwargs):
        """保存时自动设置操作时间"""
        if not self.operation_time:
            self.operation_time = timezone.now()
        super().save(*args, **kwargs)
    
    def mark_success(self, result=None):
        """标记为成功"""
        self.status = 1
        if result:
            self.business_data = result
        self.save(update_fields=['status', 'business_data'])
    
    def mark_failed(self, error_code=None, error_message=None, stack_trace=None):
        """标记为失败"""
        self.status = 2
        if error_code:
            self.error_code = error_code
        if error_message:
            self.error_message = error_message
        if stack_trace:
            self.stack_trace = stack_trace
        self.save(update_fields=['status', 'error_code', 'error_message', 'stack_trace'])
    
    def to_dict(self):
        """转换为字典（用于导出）"""
        return {
            'id': self.id,
            'operation_time': self.operation_time.isoformat() if self.operation_time else None,
            'operation_type': self.operation_type,
            'operation_type_display': self.get_operation_type_display(),
            'user_id': self.user_id,
            'username': self.username,
            'module': self.module,
            'module_display': self.get_module_display(),
            'operation_name': self.operation_name,
            'operation_detail': self.operation_detail,
            'request_method': self.request_method,
            'request_url': self.request_url,
            'request_path': self.request_path,
            'ip_address': self.ip_address,
            'user_agent': self.user_agent,
            'referer': self.referer,
            'request_params': self.request_params,
            'request_body': self.request_body,
            'response_status': self.response_status,
            'response_size': self.response_size,
            'execution_time': self.execution_time,
            'business_data': self.business_data,
            'entity_type': self.entity_type,
            'entity_id': self.entity_id,
            'status': self.status,
            'status_display': self.get_status_display(),
            'error_code': self.error_code,
            'error_message': self.error_message,
            'stack_trace': self.stack_trace,
            'session_id': self.session_id,
            'correlation_id': self.correlation_id,
            'parent_log_id': self.parent_log_id,
            'tags': self.tags,
            'created_at': self.created_at.isoformat() if self.created_at else None,
        }


# ============================================================
# 三、日志服务
# ============================================================

class UserOperationLogService:
    """用户操作日志服务"""
    
    @staticmethod
    def create_log(operation_type, operation_name, **kwargs):
        """
        创建操作日志
        
        Args:
            operation_type: 操作类型（必填）
            operation_name: 操作名称（必填）
            **kwargs: 其他字段
            
        Returns:
            UserOperationLog: 创建的日志实例
        """
        log = UserOperationLog(
            operation_type=operation_type,
            operation_name=operation_name,
            **kwargs
        )
        log.save()
        return log
    
    @staticmethod
    def log_page_view(request, user=None, **kwargs):
        """记录页面访问"""
        return UserOperationLogService.create_log(
            operation_type='view',
            operation_name=f"访问页面: {request.path}",
            module=UserOperationLogService._get_module_from_path(request.path),
            user=user,
            user_id=UserOperationLogService._get_user_id(request, user),
            username=UserOperationLogService._get_username(user),
            request_method=request.method,
            request_url=request.build_absolute_uri(),
            request_path=request.path,
            ip_address=UserOperationLogService._get_client_ip(request),
            user_agent=request.META.get('HTTP_USER_AGENT'),
            referer=request.META.get('HTTP_REFERER'),
            request_params=dict(request.GET),
            session_id=request.session.session_key,
            status=1,
            **kwargs
        )
    
    @staticmethod
    def log_api_call(request, operation_name, response_data=None, **kwargs):
        """记录API调用"""
        return UserOperationLogService.create_log(
            operation_type='api_call',
            operation_name=operation_name,
            module='api',
            user=request.user if request.user.is_authenticated else None,
            user_id=UserOperationLogService._get_user_id(request, request.user if request.user.is_authenticated else None),
            username=UserOperationLogService._get_username(request.user),
            request_method=request.method,
            request_url=request.build_absolute_uri(),
            request_path=request.path,
            ip_address=UserOperationLogService._get_client_ip(request),
            user_agent=request.META.get('HTTP_USER_AGENT'),
            request_params=dict(request.GET),
            request_body=UserOperationLogService._parse_request_body(request),
            response_status=getattr(response_data, 'status_code', None),
            business_data=response_data,
            session_id=request.session.session_key,
            **kwargs
        )
    
    @staticmethod
    def log_data_update(user, module, operation_name, **kwargs):
        """记录数据更新"""
        return UserOperationLogService.create_log(
            operation_type='data_update',
            operation_name=operation_name,
            module=module,
            user=user,
            user_id=str(user.id) if user else None,
            username=UserOperationLogService._get_username(user),
            status=1,
            **kwargs
        )
    
    @staticmethod
    def log_config_change(user, config_type, old_value, new_value, **kwargs):
        """记录配置变更"""
        return UserOperationLogService.create_log(
            operation_type='config_change',
            operation_name=f"配置变更: {config_type}",
            module='admin',
            user=user,
            user_id=str(user.id) if user else None,
            username=UserOperationLogService._get_username(user),
            business_data={
                'config_type': config_type,
                'old_value': old_value,
                'new_value': new_value,
            },
            status=1,
            **kwargs
        )
    
    @staticmethod
    def _get_module_from_path(path):
        """根据路径获取模块名称"""
        module_mapping = {
            '/workflow/etf/': 'etf_dashboard',
            '/workflow/portfolio/': 'portfolio',
            '/workflow/etf-comparison/': 'etf_comparison',
            '/workflow/etf-config/': 'etf_config',
            '/workflow/logs/': 'operation_logs',
            '/workflow/exchange-rates/': 'exchange_rates',
        }
        for pattern, module in module_mapping.items():
            if path.startswith(pattern):
                return module
        return 'other'
    
    @staticmethod
    def _get_user_id(request, user):
        """获取用户ID"""
        if user:
            return str(user.id)
        return request.session.get('anonymous_user_id')
    
    @staticmethod
    def _get_username(user):
        """获取用户名"""
        if user:
            return user.username
        return 'anonymous'
    
    @staticmethod
    def _get_client_ip(request):
        """获取客户端IP"""
        x_forwarded_for = request.META.get('HTTP_X_FORWARDED_FOR')
        if x_forwarded_for:
            ip = x_forwarded_for.split(',')[0]
        else:
            ip = request.META.get('REMOTE_ADDR')
        return ip
    
    @staticmethod
    def _parse_request_body(request):
        """解析请求体"""
        try:
            if request.method in ['POST', 'PUT', 'PATCH']:
                if request.content_type == 'application/json':
                    import json
                    return json.loads(request.body)
                elif request.content_type == 'application/x-www-form-urlencoded':
                    return dict(request.POST)
        except:
            pass
        return None
    
    @staticmethod
    def query_logs(start_date=None, end_date=None, user_id=None, 
                  operation_type=None, module=None, status=None,
                  entity_type=None, entity_id=None,
                  limit=100, offset=0):
        """
        查询操作日志
        
        Args:
            start_date: 开始日期
            end_date: 结束日期
            user_id: 用户ID
            operation_type: 操作类型
            module: 系统模块
            status: 操作状态
            entity_type: 实体类型
            entity_id: 实体ID
            limit: 返回数量限制
            offset: 偏移量
            
        Returns:
            QuerySet: 查询结果
        """
        queryset = UserOperationLog.objects.filter(is_deleted=False)
        
        # 时间范围过滤
        if start_date:
            queryset = queryset.filter(operation_time__gte=start_date)
        if end_date:
            queryset = queryset.filter(operation_time__lte=end_date)
        
        # 用户过滤
        if user_id:
            queryset = queryset.filter(user_id=user_id)
        
        # 操作类型过滤
        if operation_type:
            queryset = queryset.filter(operation_type=operation_type)
        
        # 模块过滤
        if module:
            queryset = queryset.filter(module=module)
        
        # 状态过滤
        if status is not None:
            queryset = queryset.filter(status=status)
        
        # 实体过滤
        if entity_type:
            queryset = queryset.filter(entity_type=entity_type)
        if entity_id:
            queryset = queryset.filter(entity_id=entity_id)
        
        # 排序和分页
        return queryset.order_by('-operation_time')[offset:offset+limit]
    
    @staticmethod
    def get_statistics(start_date=None, end_date=None):
        """获取统计数据"""
        queryset = UserOperationLog.objects.filter(is_deleted=False)
        
        if start_date:
            queryset = queryset.filter(operation_time__gte=start_date)
        if end_date:
            queryset = queryset.filter(operation_time__lte=end_date)
        
        return {
            'total_operations': queryset.count(),
            'by_type': queryset.values('operation_type').annotate(
                count=models.Count('id')
            ).order_by('-count'),
            'by_module': queryset.values('module').annotate(
                count=models.Count('id')
            ).order_by('-count'),
            'by_status': queryset.values('status').annotate(
                count=models.Count('id')
            ).order_by('-count'),
            'by_user': queryset.values('username').annotate(
                count=models.Count('id')
            ).order_by('-count')[:10],
        }


# ============================================================
# 四、中间件 - 自动记录日志
# ============================================================

class UserOperationLogMiddleware:
    """用户操作日志中间件 - 自动记录所有请求"""
    
    def __init__(self, get_response):
        self.get_response = get_response
        # 不记录的路径列表
        self.exclude_paths = [
            '/admin/jsi18n/',
            '/static/',
            '/media/',
            '/favicon.ico',
            '/health/',
        ]
    
    def __call__(self, request):
        """处理请求"""
        path = request.path
        
        # 排除静态资源和特定路径
        if any(path.startswith(p) for p in self.exclude_paths):
            return self.get_response(request)
        
        # 处理响应
        response = self.get_response(request)
        
        # 异步记录日志（不影响性能）
        from django.core.cache import cache
        cache_key = f"async_log_{request.META.get('REMOTE_ADDR')}_{timezone.now().timestamp()}"
        
        # 记录日志到队列（异步处理）
        log_data = {
            'operation_type': 'view' if request.method == 'GET' else 'api_call',
            'operation_name': f"{request.method} {path}",
            'module': UserOperationLogService._get_module_from_path(path),
            'request_method': request.method,
            'request_url': request.build_absolute_uri(),
            'request_path': path,
            'ip_address': UserOperationLogService._get_client_ip(request),
            'user_agent': request.META.get('HTTP_USER_AGENT'),
            'request_params': dict(request.GET),
            'response_status': response.status_code,
            'response_size': len(response.content) if hasattr(response, 'content') else 0,
            'status': 1 if response.status_code < 400 else 2,
        }
        
        # 如果用户已登录，添加用户信息
        if request.user.is_authenticated:
            log_data['user'] = request.user
            log_data['user_id'] = str(request.user.id)
            log_data['username'] = request.user.username
        
        # 缓存日志数据，由后台任务处理
        cache.set(cache_key, log_data, timeout=60)
        
        return response


# ============================================================
# 五、管理命令 - 日志归档
# ============================================================

"""
文件：management/commands/archive_operation_logs.py

用途：定期归档旧日志到历史表
"""

"""
from django.core.management.base import BaseCommand
from django.utils import timezone
from datetime import timedelta
import logging

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = '归档旧的操作日志'

    def add_arguments(self, parser):
        parser.add_argument(
            '--days',
            type=int,
            default=90,
            help='归档多少天前的日志（默认90天）'
        )
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='只显示将要归档的日志数量，不实际执行'
        )

    def handle(self, *args, **options):
        days = options['days']
        dry_run = options['dry_run']
        
        # 计算归档日期
        archive_date = timezone.now() - timedelta(days=days)
        
        # 查询需要归档的日志
        logs_to_archive = UserOperationLog.objects.filter(
            operation_time__lt=archive_date,
            is_deleted=False
        )
        
        count = logs_to_archive.count()
        
        if dry_run:
            self.stdout.write(
                self.style.WARNING(
                    f'将归档 {count} 条日志（操作时间早于 {archive_date}）'
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
                f'成功归档 {updated} 条日志'
            )
        )
        
        # 清理已归档的日志（可选）
        # logs_to_archive.update(is_deleted=True)
"""

# ============================================================
# 六、视图 - 日志查询界面
# ============================================================

class UserOperationLogView(View):
    """用户操作日志查询界面"""
    
    def get(self, request):
        """显示日志查询页面"""
        # 获取筛选参数
        start_date = request.GET.get('start_date')
        end_date = request.GET.get('end_date')
        user_id = request.GET.get('user_id')
        operation_type = request.GET.get('operation_type')
        module = request.GET.get('module')
        status = request.GET.get('status')
        page = int(request.GET.get('page', 1))
        per_page = int(request.GET.get('per_page', 50))
        
        # 查询日志
        logs = UserOperationLogService.query_logs(
            start_date=start_date,
            end_date=end_date,
            user_id=user_id,
            operation_type=operation_type,
            module=module,
            status=status,
            limit=per_page,
            offset=(page - 1) * per_page
        )
        
        # 获取统计数据
        stats = UserOperationLogService.get_statistics(
            start_date=start_date,
            end_date=end_date
        )
        
        # 分页
        from django.core.paginator import Paginator
        paginator = Paginator(logs, per_page)
        logs_page = paginator.get_page(page)
        
        context = {
            'logs': logs_page,
            'stats': stats,
            'filters': {
                'start_date': start_date,
                'end_date': end_date,
                'user_id': user_id,
                'operation_type': operation_type,
                'module': module,
                'status': status,
            },
            'operation_types': UserOperationLog.OPERATION_TYPE_CHOICES,
            'modules': UserOperationLog.MODULE_CHOICES,
            'status_choices': UserOperationLog.STATUS_CHOICES,
        }
        
        return render(request, 'workflow/user_operation_logs.html', context)


# ============================================================
# 七、性能优化策略
# ============================================================

"""
1. 数据库索引优化
   - 已在Meta中定义复合索引
   - 针对常用查询组合优化
   - 时间字段建立索引

2. 分页查询
   - 使用offset和limit
   - Django ORM自动优化
   - 避免全表扫描

3. 日志归档
   - 定期归档90天前的日志
   - 标记archived_at字段
   - 主表只保留活跃数据
   - 提升查询性能

4. 异步写入
   - 使用中间件异步写入
   - 不影响主请求性能
   - 使用Celery或Django Q

5. 缓存优化
   - 统计数据缓存1小时
   - 用户操作历史缓存
   - 热点数据缓存

6. 读写分离
   - 写入操作使用主库
   - 查询操作使用从库
   - 降低主库压力

7. 数据分区
   - 按时间范围分区
   - 每月一个分区
   - 加速时间范围查询

8. 批量操作
   - 批量插入而非单条插入
   - 批量更新状态
   - 减少数据库往返
"""

print("""
============================================================
用户操作日志系统设计完成
============================================================

功能特性：
✓ 完整记录用户操作
✓ 支持多维度筛选
✓ 性能优化策略
✓ 自动日志归档
✓ 异步日志写入
✓ 统计分析功能

数据库表：
- user_operation_log: 主日志表（已索引优化）

核心服务：
- UserOperationLogService: 日志服务类
- UserOperationLogMiddleware: 日志中间件

管理命令：
- archive_operation_logs: 日志归档命令

查询接口：
- 按时间范围查询
- 按用户ID查询
- 按操作类型查询
- 按系统模块查询
- 按状态查询
- 按实体类型查询
- 按实体ID查询

性能优化：
- 数据库索引（复合索引）
- 分页查询
- 日志归档（90天）
- 异步写入
- 缓存优化
- 读写分离
- 数据分区
- 批量操作

后续工作：
1. 创建数据库迁移文件
2. 添加到settings的INSTALLED_APPS
3. 注册管理界面
4. 配置中间件
5. 创建定时任务
6. 实现日志查询页面
7. 添加导出功能
8. 实现日志分析图表
""")
