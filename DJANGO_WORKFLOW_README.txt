================================================================================
Django ETF工作流管理系统 - 完整使用指南
================================================================================

## 系统概述

这是一个基于Django MTV（Model-Template-View）架构的ETF数据分析工作流管理系统。

核心技术栈：
- Django 3.2 - Web框架
- Django REST Framework - API框架
- PyMySQL - MySQL驱动
- Bootstrap 4.6 - 前端UI框架
- Chart.js - 数据可视化

## 项目结构

py_project/
├── etf_workflow_project/     # Django项目配置
│   ├── settings.py           # 项目配置
│   ├── urls.py               # 主URL路由
│   └── wsgi.py               # WSGI配置
├── workflow/                 # 工作流应用（MTV结构）
│   ├── models.py             # Model层 - 数据模型
│   ├── views.py              # View层 - 业务逻辑
│   ├── templates/            # Template层 - HTML模板
│   ├── serializers.py        # REST API序列化器
│   ├── admin.py              # Django Admin配置
│   ├── urls.py               # 应用URL路由
│   └── engine.py             # 工作流执行引擎
├── manage.py                 # Django管理脚本
└── init_django_workflow.sh   # 初始化脚本

## 安装部署

### 步骤1：配置数据库

1.1 创建MySQL数据库：
```sql
CREATE DATABASE etf_workflow DEFAULT CHARSET=utf8mb4;
```

1.2 修改数据库配置：
编辑 etf_workflow_project/settings.py，修改第83-92行：
```python
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.mysql',
        'NAME': 'etf_workflow',
        'USER': 'root',
        'PASSWORD': 'your_password',  # 修改为你的MySQL密码
        'HOST': 'localhost',
        'PORT': '3306',
    }
}
```

### 步骤2：执行数据库迁移

```bash
source venv/bin/activate
python manage.py migrate
```

### 步骤3：创建管理员账号

```bash
python manage.py createsuperuser
```
按提示输入用户名、邮箱和密码

### 步骤4：初始化测试数据（可选）

```bash
chmod +x init_django_workflow.sh
./init_django_workflow.sh
```

或手动初始化：
```bash
python manage.py shell
```
然后复制init_django_workflow.sh中的Python代码执行

### 步骤5：启动服务

```bash
python manage.py runserver 0.0.0.0:8000
```

## 访问系统

1. Web管理界面：http://localhost:8000
   - 仪表板：http://localhost:8000/workflow/
   - 工作流列表：http://localhost:8000/workflow/workflows/
   - 执行记录：http://localhost:8000/workflow/instances/

2. Django Admin后台：http://localhost:8000/admin
   - 使用第3步创建的管理员账号登录
   - 可以直接管理所有数据模型

3. REST API接口：
   - 工作流列表：GET http://localhost:8000/workflow/api/workflows/
   - 执行工作流：POST http://localhost:8000/workflow/api/workflows/{id}/execute/
   - 实例状态：GET http://localhost:8000/workflow/api/instances/{id}/status/
   - 工作流统计：GET http://localhost:8000/workflow/api/workflows/stats/

## 核心功能

### 1. MTV架构

**Model层（models.py）** - 数据模型定义
- Workflow - 工作流定义
- WorkflowStep - 工作流步骤
- WorkflowInstance - 工作流实例
- WorkflowInstanceStep - 实例步骤
- SystemLog - 系统日志
- ETFData - ETF数据
- PortfolioConfig - 投资组合配置
- AnalysisReport - 分析报告
- Notification - 通知记录

**View层（views.py）** - 业务逻辑
- DashboardView - 仪表板页面
- WorkflowListView - 工作流列表
- WorkflowDetailView - 工作流详情
- InstanceListView - 执行记录列表
- InstanceDetailView - 实例详情
- WorkflowViewSet - 工作流API
- WorkflowInstanceViewSet - 实例API

**Template层（templates/）** - 页面展示
- base.html - 基础模板
- dashboard.html - 仪表板
- workflow_list.html - 工作流列表
- workflow_detail.html - 工作流详情
- instance_list.html - 执行记录
- instance_detail.html - 实例详情

### 2. 工作流执行引擎（engine.py）

- 工作流调度和执行
- 步骤级别的重试机制
- 错误处理和日志记录
- 可扩展的处理器注册

### 3. Django Admin管理

访问 http://localhost:8000/admin 可以：
- 查看和编辑所有数据
- 创建新工作流和步骤
- 查看执行历史
- 管理系统配置

### 4. REST API

支持完整的RESTful API：
```bash
# 查看所有工作流
curl http://localhost:8000/workflow/api/workflows/

# 执行工作流
curl -X POST http://localhost:8000/workflow/api/workflows/1/execute/ \
     -H "Content-Type: application/json" \
     -d '{"trigger_by":"api","context":{}}'

# 查询实例状态
curl http://localhost:8000/workflow/api/instances/1/status/
```

## 使用示例

### 示例1：创建新工作流

**方法1：使用Django Admin**
1. 访问 http://localhost:8000/admin
2. 点击"工作流" -> "添加"
3. 填写工作流信息并保存
4. 添加工作流步骤

**方法2：使用Python Shell**
```bash
python manage.py shell
```
```python
from workflow.models import Workflow, WorkflowStep

# 创建工作流
wf = Workflow.objects.create(
    name='我的工作流',
    description='描述',
    category='data_collection',
    status=1
)

# 添加步骤
WorkflowStep.objects.create(
    workflow=wf,
    name='步骤1',
    step_type='fetch_data',
    order_index=1,
    handler_type=2,
    handler_config={'function': 'my_handler'},
    retry_times=3
)
```

### 示例2：执行工作流

**方法1：Web界面**
1. 访问工作流列表页
2. 点击"执行"按钮

**方法2：API调用**
```python
import requests

response = requests.post(
    'http://localhost:8000/workflow/api/workflows/1/execute/',
    json={'trigger_by': 'api', 'context': {}}
)
print(response.json())
```

### 示例3：自定义处理器

编辑 workflow/engine.py：
```python
class WorkflowEngine:
    def _register_default_handlers(self):
        self.handlers['my_custom_handler'] = self.my_custom_handler
    
    def my_custom_handler(self, params):
        # 你的处理逻辑
        symbol = params.get('symbol')
        # ... 处理代码
        return {'status': 'success'}
```

## Django MTV核心概念

### Model（模型）
- 定义数据结构
- ORM映射到数据库表
- 包含业务逻辑方法
- 示例：workflow/models.py

### Template（模板）
- HTML页面渲染
- Django模板语言
- 继承和组件化
- 示例：workflow/templates/

### View（视图）
- 处理请求和响应
- 业务逻辑控制
- 调用Model和渲染Template
- 示例：workflow/views.py

### URL配置
- 路由映射
- URL模式匹配
- 命名空间
- 示例：workflow/urls.py

## 常见问题

### Q1：数据库连接失败
A：检查settings.py中的数据库配置，确保MySQL服务正在运行

### Q2：模板找不到
A：确保workflow应用已添加到INSTALLED_APPS，且模板路径正确

### Q3：静态文件404
A：执行 python manage.py collectstatic（生产环境）
   开发环境Django会自动处理静态文件

### Q4：如何添加新页面
A：
1. 在views.py添加视图类/函数
2. 在templates/创建HTML模板
3. 在urls.py添加URL映射

### Q5：如何修改数据模型
A：
1. 修改models.py
2. 执行：python manage.py makemigrations
3. 执行：python manage.py migrate

## 扩展开发

### 添加新的数据模型
1. 在models.py中定义Model类
2. 执行数据库迁移
3. 在admin.py注册到Admin
4. 创建serializers（如需API）

### 添加新的视图
1. 在views.py创建视图类
2. 在templates/创建模板
3. 在urls.py配置路由
4. 更新base.html添加导航

### 集成现有ETF脚本
修改engine.py的处理器：
```python
def fetch_etf_data_handler(self, params):
    from fetch_etf_data import ETFDataFetcher
    symbol = params.get('symbol')
    fetcher = ETFDataFetcher()
    data = fetcher.fetch_historical_data(symbol)
    return {'symbol': symbol, 'records': len(data)}
```

## Django命令参考

```bash
# 创建迁移文件
python manage.py makemigrations

# 执行迁移
python manage.py migrate

# 创建超级用户
python manage.py createsuperuser

# 启动开发服务器
python manage.py runserver

# 进入Python Shell
python manage.py shell

# 检查项目问题
python manage.py check

# 收集静态文件
python manage.py collectstatic
```

## 技术支持

Django官方文档：https://docs.djangoproject.com/
Django REST Framework：https://www.django-rest-framework.org/

================================================================================
