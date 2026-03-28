# 用户操作日志系统 - 完整设计文档

## 📋 目录

1. [系统概述](#系统概述)
2. [数据库设计](#数据库设计)
3. [服务层实现](#服务层实现)
4. [性能优化](#性能优化)
5. [查询接口](#查询接口)
6. [使用指南](#使用指南)
7. [部署步骤](#部署步骤)
8. [维护建议](#维护建议)

---

## 系统概述

### 功能目标

设计一个完整的用户操作日志系统，能够：
- ✓ 记录所有用户操作
- ✓ 包含完整的操作信息
- ✓ 支持多维度查询筛选
- ✓ 优化大数据量性能
- ✓ 提供统计分析功能

### 核心要素

| 要素 | 说明 | 数据类型 |
|------|------|----------|
| 操作时间 | 操作发生的具体时间 | DateTime |
| 操作类型 | 用户执行的操作类型 | Char(50) |
| 执行用户 | 操作的用户对象 | ForeignKey(User) |
| 用户ID | 用户的唯一标识（支持匿名） | Char(100) |
| 用户名 | 操作时的用户名（保留历史） | Char(150) |
| 系统模块名称 | 操作所属的模块 | Char(50) |
| 操作内容详情 | 操作的详细描述 | TextField |
| 请求信息 | URL、IP、User-Agent等 | 多个字段 |
| 响应信息 | 状态码、执行时间等 | 多个字段 |
| 业务数据 | 操作相关的业务数据 | JSONField |

---

## 数据库设计

### 模型：UserOperationLog

#### 字段说明

**1. 基本信息**
```python
operation_time      # 操作时间 (DateTime, indexed)
operation_type     # 操作类型 (Char, choices, indexed)
user               # 操作用户 (ForeignKey, indexed, null=True)
user_id            # 用户ID (Char, indexed, 支持匿名)
username           # 用户名 (Char, indexed, 保留历史）
module             # 系统模块 (Char, choices, indexed)
operation_name     # 操作名称 (Char)
operation_detail   # 操作详情 (Text)
```

**2. 请求信息**
```python
request_method     # 请求方法 (GET/POST等)
request_url        # 完整URL
request_path       # 请求路径 (indexed, 用于分组）
ip_address        # 客户端IP (indexed)
user_agent        # 浏览器代理
referer           # 来源页面
request_params     # GET参数 (JSON)
request_body       # POST/PUT body (JSON)
```

**3. 响应信息**
```python
response_status    # HTTP状态码
response_size     # 响应大小 (bytes)
execution_time    # 执行时间 (ms)
```

**4. 业务数据**
```python
business_data     # 业务相关数据 (JSON)
entity_type       # 实体类型 (indexed)
entity_id         # 实体ID (indexed)
```

**5. 状态和错误**
```python
status            # 操作状态 (0:进行中, 1:成功, 2:失败, 3:已取消)
error_code        # 错误码
error_message     # 错误信息
stack_trace       # 堆栈信息
```

**6. 审计信息**
```python
session_id        # 会话ID (indexed)
correlation_id    # 关联ID (indexed, 用于跨服务追踪）
parent_log_id    # 父日志ID (支持操作链）
tags             # 标签 (JSON, 用于分类）
```

**7. 元数据**
```python
created_at        # 创建时间 (indexed)
updated_at        # 更新时间
is_deleted        # 是否已删除 (软删除，indexed）
archived_at       # 归档时间 (indexed, null=未归档）
```

#### 操作类型枚举

```python
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
```

#### 系统模块枚举

```python
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
```

#### 数据库索引优化

```python
class Meta:
    db_table = 'user_operation_log'
    ordering = ['-operation_time']
    
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
```

**索引设计原则：**
- ✓ 为所有筛选字段建立索引
- ✓ 复合索引优化多条件查询
- ✓ 时间字段建立倒序索引
- ✓ 避免过多索引影响写入性能

---

## 服务层实现

### UserOperationLogService

#### 1. 创建日志

```python
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
```

#### 2. 记录页面访问

```python
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
        request_params=dict(request.GET),
        session_id=request.session.session_key,
        status=1,
        **kwargs
    )
```

#### 3. 记录API调用

```python
@staticmethod
def log_api_call(request, operation_name, response_data=None, **kwargs):
    """记录API调用"""
    return UserOperationLogService.create_log(
        operation_type='api_call',
        operation_name=operation_name,
        module='api',
        user=request.user if request.user.is_authenticated else None,
        user_id=UserOperationLogService._get_user_id(request, request.user),
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
```

#### 4. 查询日志

```python
@staticmethod
def query_logs(start_date=None, end_date=None, user_id=None, 
              operation_type=None, module=None, status=None,
              entity_type=None, entity_id=None,
              limit=100, offset=0):
    """
    查询操作日志
    
    支持的筛选条件：
    - 时间范围 (start_date, end_date)
    - 用户 (user_id)
    - 操作类型 (operation_type)
    - 系统模块 (module)
    - 操作状态 (status)
    - 实体类型和ID (entity_type, entity_id)
    
    Returns:
        QuerySet: 查询结果（带分页）
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
```

#### 5. 统计分析

```python
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
```

---

## 性能优化

### 1. 数据库层优化

#### 索引策略
- ✓ 复合索引优化常用查询组合
- ✓ 时间字段建立倒序索引
- ✓ 避免过度索引（影响写入性能）

#### 分页查询
```python
# 使用offset和limit
queryset.order_by('-operation_time')[offset:offset+limit]

# 或使用Django Paginator
paginator = Paginator(queryset, per_page)
logs_page = paginator.get_page(page_number)
```

#### 只查询必要字段
```python
# 只选择需要的字段
queryset.only(
    'operation_time', 'operation_type', 'username', 
    'operation_name', 'status', 'ip_address'
)
```

### 2. 应用层优化

#### 异步日志写入
```python
# 使用中间件异步写入，不影响主请求性能
from django.core.cache import cache

cache.set(cache_key, log_data, timeout=60)
# 后台任务处理缓存中的日志
```

#### 缓存优化
```python
# 统计数据缓存1小时
@cache_page(60 * 60)
def get_statistics():
    # ...

# 用户操作历史缓存
cache_key = f"user_ops_{user_id}_{date}"
```

#### 批量操作
```python
# 批量插入而非单条插入
UserOperationLog.objects.bulk_create([log1, log2, ...], batch_size=100)

# 批量更新状态
queryset.update(status=1)
```

### 3. 架构层优化

#### 读写分离
```python
# 查询使用从库
# 配置DATABASE_ROUTERS
# 主库负责写入，从库负责读取
```

#### 数据分区
```sql
-- 按时间范围分区
CREATE TABLE user_operation_log_2025_01 PARTITION OF user_operation_log
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

#### 日志归档
```python
# 定期归档90天前的日志
python manage.py archive_user_operation_logs --days=90

# 归档后标记
archived_at = timezone.now()
is_deleted = False  # 或True，如果要从主表移除
```

### 4. 存储优化

#### 归档策略
- 90天前的日志自动归档
- 归档日志移至历史表
- 主表只保留活跃数据

#### 清理策略
- 归档180天前的日志删除
- 错误日志保留更长时间
- 重要操作日志永久保留

---

## 查询接口

### REST API

#### 1. 获取日志列表
```http
GET /workflow/api/operation-logs/

Query Parameters:
- start_date: 2025-01-01T00:00:00
- end_date: 2025-12-31T23:59:59
- user_id: 123
- operation_type: view
- module: portfolio
- status: 1
- page: 1
- per_page: 50

Response:
{
    "results": [...],
    "count": 1000,
    "page": 1,
    "pages": 20
}
```

#### 2. 获取日志详情
```http
GET /workflow/api/operation-logs/{id}/

Response:
{
    "id": 1,
    "operation_time": "2025-12-29T12:30:45Z",
    "operation_type": "view",
    "operation_type_display": "页面访问",
    "user_id": "123",
    "username": "admin",
    "module": "portfolio",
    "module_display": "投资组合",
    "operation_name": "访问投资组合页面",
    "ip_address": "192.168.1.100",
    "request_path": "/workflow/portfolio/",
    "status": 1,
    "execution_time": 150
}
```

#### 3. 导出日志
```http
GET /workflow/api/operation-logs/export/

Query Parameters: 同列表查询

Response:
Content-Type: application/csv
Content-Disposition: attachment; filename="operation_logs_20251229.csv"
```

#### 4. 获取统计
```http
GET /workflow/api/operation-logs/statistics/

Query Parameters:
- start_date: 2025-01-01T00:00:00
- end_date: 2025-12-31T23:59:59

Response:
{
    "total_operations": 10000,
    "by_type": [
        {"operation_type": "view", "count": 5000},
        {"operation_type": "update", "count": 3000},
        ...
    ],
    "by_module": [
        {"module": "portfolio", "count": 4000},
        ...
    ],
    "by_user": [
        {"username": "admin", "count": 2000},
        ...
    ]
}
```

---

## 使用指南

### 1. 在代码中记录日志

#### 示例1：记录页面访问
```python
# views.py
from workflow.models import UserOperationLogService

def portfolio_view(request):
    # ... 业务逻辑 ...
    
    # 记录访问日志
    UserOperationLogService.log_page_view(
        request=request,
        user=request.user if request.user.is_authenticated else None,
        entity_type='portfolio',
        entity_id='default'
    )
    
    return render(request, 'workflow/portfolio.html')
```

#### 示例2：记录API调用
```python
# views.py
def update_realtime_data(request):
    UserOperationLogService.log_api_call(
        request=request,
        operation_name='更新实时数据',
        entity_type='etf_data',
        business_data={
            'allocation': allocation,
            'total_investment': total_investment
        }
    )
    # ... 业务逻辑 ...
```

#### 示例3：记录数据更新
```python
# services.py
def update_etf_data():
    # ... 更新逻辑 ...
    
    UserOperationLogService.log_data_update(
        user=request.user,
        module='etf_dashboard',
        operation_name='更新ETF数据',
        entity_type='etf_data',
        entity_id='SCHD'
    )
```

#### 示例4：记录配置变更
```python
# admin.py
def save_model(request, obj):
    old_value = get_old_value(obj)
    new_value = get_new_value(obj)
    
    UserOperationLogService.log_config_change(
        user=request.user,
        config_type=obj.__class__.__name__,
        old_value=old_value,
        new_value=new_value,
        entity_id=obj.id
    )
```

### 2. 查询日志

#### 方式1：Web界面
访问 http://127.0.0.1:8002/workflow/user-operation-logs/

#### 方式2：API查询
```python
import requests

response = requests.get(
    'http://127.0.0.1:8002/workflow/api/operation-logs/',
    params={
        'user_id': '123',
        'start_date': '2025-12-01T00:00:00',
        'end_date': '2025-12-31T23:59:59',
        'page': 1,
        'per_page': 50
    }
)

logs = response.json()
```

### 3. 自动记录（中间件）

配置中间件后，所有请求自动记录：

```python
# settings.py
MIDDLEWARE = [
    ...
    'workflow.middleware.UserOperationLogMiddleware',
    ...
]
```

---

## 部署步骤

### 1. 添加模型

将`UserOperationLog`模型添加到`workflow/models.py`

### 2. 创建数据库迁移

```bash
python manage.py makemigrations workflow
python manage.py migrate workflow
```

### 3. 创建服务文件

将`UserOperationLogService`类添加到`workflow/services.py`

### 4. 配置中间件

```python
# workflow/middleware.py
class UserOperationLogMiddleware:
    def __call__(self, request):
        # ... 实现 ...
```

### 5. 注册中间件

```python
# settings.py
MIDDLEWARE = [
    ...
    'workflow.middleware.UserOperationLogMiddleware',
    ...
]
```

### 6. 创建视图和URL

```python
# views.py
class UserOperationLogView(View):
    def get(self, request):
        # ... 实现 ...

# urls.py
path('user-operation-logs/', views.UserOperationLogView.as_view(), name='user_operation_logs'),
```

### 7. 创建模板

创建`workflow/templates/workflow/user_operation_logs.html`

### 8. 配置定时任务

```python
# settings.py
CELERY_BEAT_SCHEDULE = {
    'archive-logs': {
        'task': 'workflow.tasks.archive_operation_logs',
        'schedule': crontab(hour=2, minute=0),  # 每天凌晨2点
    },
}
```

### 9. 验证部署

```bash
# 1. 启动服务器
python manage.py runserver

# 2. 访问日志页面
http://127.0.0.1:8002/workflow/user-operation-logs/

# 3. 测试日志记录
# 访问几个页面，检查日志是否记录

# 4. 测试查询功能
# 使用筛选条件查询日志

# 5. 测试导出功能
# 导出日志CSV文件

# 6. 测试归档功能
python manage.py archive_user_operation_logs --days=90 --dry-run
```

---

## 维护建议

### 1. 日常维护

#### 监控指标
- 日志表大小
- 查询性能
- 写入性能
- 归档执行状态

#### 定期检查
```sql
-- 检查表大小
SELECT 
    table_name,
    ROUND((data_length + index_length) / 1024 / 1024, 2) AS size_mb
FROM information_schema.tables
WHERE table_schema = 'your_database'
AND table_name = 'user_operation_log';
```

### 2. 性能调优

#### 查询优化
- 使用EXPLAIN分析慢查询
- 添加必要的索引
- 避免SELECT *
- 使用只查询必要字段

#### 存储优化
- 定期归档旧数据
- 压缩历史日志
- 清理过期缓存
- 监控磁盘空间

### 3. 容量规划

#### 预估日志量
假设每天10,000条操作：
- 1年：3,650,000条
- 每条约1KB
- 总计：3.6GB

#### 归档策略
- 活跃表：保留90天（约900MB）
- 历史表：保留1年（约2.7GB）
- 超过1年：压缩归档

### 4. 故障排查

#### 日志记录失败
- 检查数据库连接
- 检查中间件配置
- 查看应用错误日志
- 验证模型迁移

#### 查询性能差
- 检查索引是否生效
- 使用EXPLAIN分析查询计划
- 考虑增加缓存
- 评估分区策略

#### 磁盘空间不足
- 立即执行归档
- 压缩历史数据
- 删除过期归档
- 扩容存储

---

## 附录

### A. 日志字段映射表

| 功能需求 | 对应字段 | 说明 |
|----------|----------|------|
| 操作时间 | operation_time | 操作发生的时间 |
| 操作类型 | operation_type | 操作类型枚举 |
| 执行用户 | user, user_id | 支持已登录和匿名用户 |
| 操作内容详情 | operation_name, operation_detail | 名称和详情 |
| 系统模块名称 | module | 模块枚举 |

### B. 操作类型完整列表

| 值 | 显示名称 | 说明 |
|-----|---------|------|
| view | 页面访问 | 用户访问某个页面 |
| create | 创建操作 | 创建新数据 |
| update | 更新操作 | 修改现有数据 |
| delete | 删除操作 | 删除数据 |
| query | 查询操作 | 查询数据 |
| export | 导出操作 | 导出数据 |
| import | 导入操作 | 导入数据 |
| login | 登录操作 | 用户登录 |
| logout | 登出操作 | 用户登出 |
| data_update | 数据更新 | 系统更新数据 |
| portfolio_analysis | 组合分析 | 投资组合分析 |
| etf_comparison | ETF对比 | ETF对比分析 |
| script_execute | 脚本执行 | 执行脚本 |
| scheduled_task | 定时任务 | 定时任务执行 |
| manual_trigger | 手动触发 | 手动触发操作 |
| config_change | 配置变更 | 修改配置 |
| api_call | API调用 | API接口调用 |

### C. 性能基准

| 指标 | 目标值 | 测量方法 |
|--------|---------|----------|
| 单条日志写入 | < 10ms | time.time() |
| 查询响应时间 | < 100ms | Django Debug Toolbar |
| 页面加载影响 | < 5% | 性能测试 |
| 磁盘I/O | < 50% | 系统监控 |
| CPU使用率 | < 30% | 系统监控 |

### D. 监控告警规则

| 指标 | 告警阈值 | 处理方式 |
|--------|----------|----------|
| 日志表大小 | > 10GB | 立即归档 |
| 写入延迟 | > 100ms | 检查数据库 |
| 查询延迟 | > 500ms | 检查索引 |
| 磁盘使用率 | > 80% | 清理空间 |
| 归档失败 | 任何失败 | 人工介入 |

---

## 总结

本设计提供了一个完整的用户操作日志系统，具有以下特点：

✅ **完整性**：记录所有必要的操作信息
✅ **灵活性**：支持多种查询和筛选
✅ **性能**：优化的索引和查询策略
✅ **可扩展**：支持未来功能扩展
✅ **易维护**：清晰的代码结构和文档

系统设计遵循最佳实践，能够满足日志记录、查询分析、性能优化的全方位需求。

---

**文档版本**：1.0
**创建日期**：2025-12-29
**更新日期**：2025-12-29
**维护人员**：开发团队
