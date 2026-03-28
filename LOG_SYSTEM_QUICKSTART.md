# 用户操作日志系统 - 快速开始指南

## 📚 文档导航

- `user_operation_log_system.py` - 完整的模型和服务代码
- `USER_OPERATION_LOG_DESIGN.md` - 详细设计文档（**主要参考**）
- `LOG_SYSTEM_QUICKSTART.md` - 本文件（快速实施指南）

---

## 🚀 快速实施（5步完成）

### 第1步：添加模型（5分钟）

编辑 `/Users/liunian/Desktop/dnmp/py_project/workflow/models.py`

在文件末尾，`class ExchangeRate(models.Model):` 之前添加：

```python
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
    
    operation_time = models.DateTimeField(
        db_index=True,
        verbose_name='操作时间'
    )
    
    operation_type = models.CharField(
        max_length=50,
        choices=OPERATION_TYPE_CHOICES,
        db_index=True,
        verbose_name='操作类型'
    )
    
    user = models.ForeignKey(
        User,
        null=True,
        blank=True,
        on_delete=models.SET_NULL,
        related_name='operation_logs',
        db_index=True,
        verbose_name='操作用户'
    )
    
    user_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='用户ID'
    )
    
    username = models.CharField(
        max_length=150,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='用户名'
    )
    
    module = models.CharField(
        max_length=50,
        choices=MODULE_CHOICES,
        db_index=True,
        verbose_name='系统模块'
    )
    
    operation_name = models.CharField(
        max_length=200,
        verbose_name='操作名称'
    )
    
    operation_detail = models.TextField(
        null=True,
        blank=True,
        verbose_name='操作详情'
    )
    
    # ============ 请求信息 ============
    
    request_method = models.CharField(
        max_length=10,
        null=True,
        blank=True,
        verbose_name='请求方法'
    )
    
    request_url = models.CharField(
        max_length=500,
        null=True,
        blank=True,
        verbose_name='请求URL'
    )
    
    request_path = models.CharField(
        max_length=200,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='请求路径'
    )
    
    ip_address = models.GenericIPAddressField(
        null=True,
        blank=True,
        db_index=True,
        verbose_name='IP地址'
    )
    
    user_agent = models.TextField(
        null=True,
        blank=True,
        verbose_name='用户代理'
    )
    
    referer = models.CharField(
        max_length=500,
        null=True,
        blank=True,
        verbose_name='来源页面'
    )
    
    # ============ 响应信息 ============
    
    response_status = models.IntegerField(
        null=True,
        blank=True,
        verbose_name='响应状态码'
    )
    
    response_size = models.BigIntegerField(
        null=True,
        blank=True,
        verbose_name='响应大小'
    )
    
    execution_time = models.IntegerField(
        null=True,
        blank=True,
        verbose_name='执行时间(毫秒)'
    )
    
    # ============ 业务数据 ============
    
    business_data = models.JSONField(
        null=True,
        blank=True,
        verbose_name='业务数据'
    )
    
    entity_type = models.CharField(
        max_length=50,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='实体类型'
    )
    
    entity_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='实体ID'
    )
    
    # ============ 状态和错误 ============
    
    status = models.IntegerField(
        choices=STATUS_CHOICES,
        default=0,
        db_index=True,
        verbose_name='操作状态'
    )
    
    error_code = models.CharField(
        max_length=50,
        null=True,
        blank=True,
        verbose_name='错误码'
    )
    
    error_message = models.TextField(
        null=True,
        blank=True,
        verbose_name='错误信息'
    )
    
    stack_trace = models.TextField(
        null=True,
        blank=True,
        verbose_name='堆栈信息'
    )
    
    # ============ 审计信息 ============
    
    session_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='会话ID'
    )
    
    correlation_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        db_index=True,
        verbose_name='关联ID'
    )
    
    parent_log_id = models.CharField(
        max_length=100,
        null=True,
        blank=True,
        verbose_name='父日志ID'
    )
    
    tags = models.JSONField(
        null=True,
        blank=True,
        verbose_name='标签'
    )
    
    # ============ 元数据 ============
    
    created_at = models.DateTimeField(
        auto_now_add=True,
        db_index=True,
        verbose_name='创建时间'
    )
    
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name='更新时间'
    )
    
    is_deleted = models.BooleanField(
        default=False,
        db_index=True,
        verbose_name='已删除'
    )
    
    archived_at = models.DateTimeField(
        null=True,
        blank=True,
        db_index=True,
        verbose_name='归档时间'
    )
    
    class Meta:
        db_table = 'user_operation_log'
        verbose_name = '用户操作日志'
        verbose_name_plural = '用户操作日志'
        ordering = ['-operation_time']
        
        indexes = [
            models.Index(fields=['user_id', 'operation_time'], name='idx_user_time'),
            models.Index(fields=['operation_type', 'operation_time'], name='idx_type_time'),
            models.Index(fields=['module', 'operation_time'], name='idx_module_time'),
            models.Index(fields=['status', 'operation_time'], name='idx_status_time'),
            models.Index(fields=['ip_address', 'operation_time'], name='idx_ip_time'),
            models.Index(fields=['session_id', 'operation_time'], name='idx_session_time'),
            models.Index(fields=['correlation_id'], name='idx_correlation'),
            models.Index(fields=['entity_type', 'entity_id'], name='idx_entity'),
            models.Index(fields=['-operation_time'], name='idx_time_desc'),
        ]
    
    def __str__(self):
        return f"[{self.operation_type}] {self.username or self.user_id} - {self.operation_name}"
    
    def save(self, *args, **kwargs):
        if not self.operation_time:
            self.operation_time = timezone.now()
        super().save(*args, **kwargs)
```

### 第2步：创建数据库迁移（2分钟）

```bash
cd /Users/liunian/Desktop/dnmp/py_project
python manage.py makemigrations workflow
python manage.py migrate workflow
```

### 第3步：添加服务类（10分钟）

在 `/Users/liunian/Desktop/dnmp/py_project/workflow/services.py` 末尾添加：

```python
class UserOperationLogService:
    """用户操作日志服务"""
    
    @staticmethod
    def create_log(operation_type, operation_name, **kwargs):
        """创建操作日志"""
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
            module='etf_dashboard' if 'etf' in request.path else 'portfolio' if 'portfolio' in request.path else 'other',
            user=user,
            user_id=str(user.id) if user and user.is_authenticated else None,
            username=user.username if user else 'anonymous',
            request_method=request.method,
            request_url=request.build_absolute_uri(),
            request_path=request.path,
            ip_address=request.META.get('REMOTE_ADDR'),
            user_agent=request.META.get('HTTP_USER_AGENT'),
            referer=request.META.get('HTTP_REFERER'),
            request_params=dict(request.GET),
            session_id=request.session.session_key,
            status=1,
            **kwargs
        )
    
    @staticmethod
    def log_api_call(request, operation_name, **kwargs):
        """记录API调用"""
        return UserOperationLogService.create_log(
            operation_type='api_call',
            operation_name=operation_name,
            module='api',
            user=request.user if request.user.is_authenticated else None,
            user_id=str(request.user.id) if request.user.is_authenticated else None,
            username=request.user.username if request.user.is_authenticated else 'anonymous',
            request_method=request.method,
            request_url=request.build_absolute_uri(),
            request_path=request.path,
            ip_address=request.META.get('REMOTE_ADDR'),
            status=1,
            **kwargs
        )
    
    @staticmethod
    def query_logs(start_date=None, end_date=None, user_id=None, 
                  operation_type=None, module=None, status=None,
                  limit=100, offset=0):
        """查询操作日志"""
        queryset = UserOperationLog.objects.filter(is_deleted=False)
        
        if start_date:
            queryset = queryset.filter(operation_time__gte=start_date)
        if end_date:
            queryset = queryset.filter(operation_time__lte=end_date)
        if user_id:
            queryset = queryset.filter(user_id=user_id)
        if operation_type:
            queryset = queryset.filter(operation_type=operation_type)
        if module:
            queryset = queryset.filter(module=module)
        if status is not None:
            queryset = queryset.filter(status=status)
        
        return queryset.order_by('-operation_time')[offset:offset+limit]
```

### 第4步：添加视图和URL（10分钟）

编辑 `/Users/liunian/Desktop/dnmp/py_project/workflow/views.py`，添加：

```python
class UserOperationLogView(View):
    """用户操作日志查询界面"""
    
    def get(self, request):
        from workflow.services import UserOperationLogService
        from django.core.paginator import Paginator
        
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
        
        # 分页
        paginator = Paginator(logs, per_page)
        logs_page = paginator.get_page(page)
        
        context = {
            'logs': logs_page,
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
```

编辑 `/Users/liunian/Desktop/dnmp/py_project/workflow/urls.py`，添加：

```python
path('user-operation-logs/', views.UserOperationLogView.as_view(), name='user_operation_logs'),
```

### 第5步：测试系统（5分钟）

```bash
# 启动服务器
python manage.py runserver 8002

# 访问日志页面
# 打开浏览器访问：http://127.0.0.1:8002/workflow/user-operation-logs/

# 测试日志记录
# 访问其他页面，然后返回日志页面查看记录
```

---

## 📝 在代码中使用日志

### 示例1：记录页面访问

```python
# views.py
from workflow.services import UserOperationLogService

def portfolio_view(request):
    UserOperationLogService.log_page_view(request)
    return render(request, 'workflow/portfolio.html')
```

### 示例2：记录API调用

```python
# views.py
def update_realtime_data(request):
    UserOperationLogService.log_api_call(
        request=request,
        operation_name='更新实时数据',
        business_data={'allocation': allocation}
    )
    # ... 业务逻辑
```

---

## 🔧 高级配置

### 配置日志归档（定时任务）

```python
# settings.py
CELERY_BEAT_SCHEDULE = {
    'archive-logs': {
        'task': 'workflow.tasks.archive_operation_logs',
        'schedule': crontab(hour=2, minute=0),  # 每天凌晨2点
    },
}
```

### 配置中间件（自动记录）

```python
# settings.py
MIDDLEWARE = [
    ...
    'workflow.middleware.UserOperationLogMiddleware',
    ...
]
```

---

## 📊 预估数据量

假设每天10,000条操作：

| 时间跨度 | 日志数量 | 预估大小 | 归档后 |
|---------|---------|-----------|---------|
| 1天 | 10,000 | 10MB | 0MB |
| 30天 | 300,000 | 300MB | 0MB |
| 90天 | 900,000 | 900MB | 0MB |
| 180天 | 1,800,000 | 1.8GB | 900MB |
| 365天 | 3,650,000 | 3.6GB | 2.7GB |

---

## ✅ 验证清单

部署完成后，检查以下项目：

- [ ] 数据库迁移成功
- [ ] 日志表创建完成
- [ ] 日志查询页面可访问
- [ ] 页面访问日志正常记录
- [ ] API调用日志正常记录
- [ ] 查询筛选功能正常
- [ ] 分页功能正常
- [ ] 统计数据准确
- [ ] 性能符合预期
- [ ] 归档命令可用

---

## 📞 故障排查

### 问题1：日志表未创建
**解决**：检查数据库迁移是否执行
```bash
python manage.py showmigrations workflow
python manage.py migrate workflow --verbosity=2
```

### 问题2：页面访问慢
**解决**：检查数据库索引
```sql
EXPLAIN SELECT * FROM user_operation_log WHERE user_id = '123';
```

### 问题3：磁盘空间不足
**解决**：立即执行归档
```bash
python manage.py archive_user_operation_logs --days=90 --delete
```

---

## 📚 相关文件

- `workflow/models.py` - 数据库模型
- `workflow/services.py` - 服务层
- `workflow/views.py` - 视图层
- `workflow/urls.py` - URL配置
- `workflow/templates/workflow/user_operation_logs.html` - 日志查询页面
- `workflow/management/commands/archive_user_operation_logs.py` - 归档命令

---

## 🎯 下一步

1. **实现日志中间件** - 自动记录所有请求
2. **添加导出功能** - 支持CSV/Excel导出
3. **实现统计分析** - 图表展示操作趋势
4. **配置告警规则** - 异常日志自动通知
5. **优化查询性能** - 根据实际使用情况调整索引

---

**快速指南完成！** 

如有问题，请参考完整设计文档：`USER_OPERATION_LOG_DESIGN.md`

祝部署顺利！🎉
