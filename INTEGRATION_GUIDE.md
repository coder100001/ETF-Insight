# Django + React 整合指南

## 🎯 整合架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户浏览器                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Django (端口8000)                         │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  URL路由                                              │  │
│  │  ├── /api/* → API路由 (workflow.urls)                │  │
│  │  ├── /admin/ → Django Admin                         │  │
│  │  └── /* → React前端 (TemplateView)                   │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  API端点                                              │  │
│  │  ├── /api/workflow/api/etf/list/                     │  │
│  │  ├── /api/workflow/api/etf/portfolio/analyze/        │  │
│  │  ├── /api/workflow/api/update-realtime/              │  │
│  │  └── ...                                             │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  静态文件服务                                         │  │
│  │  ├── /static/index.html → React入口                  │  │
│  │  ├── /static/assets/* → JS/CSS文件                   │  │
│  │  └── /static/favicon.svg                             │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 📁 项目结构

```
py_project/
├── etf_workflow_project/          # Django项目配置
│   ├── settings.py                # 包含静态文件配置
│   ├── urls.py                    # 主路由配置
│   └── ...
├── workflow/                      # Django应用
│   ├── urls.py                    # API路由
│   ├── views.py                   # API视图
│   └── ...
├── frontend/                      # React前端
│   ├── src/                       # React源代码
│   ├── dist/                      # 构建输出
│   │   ├── index.html             # Django模板
│   │   └── assets/                # 静态资源
│   └── vite.config.ts             # 构建配置
├── static/                        # Django静态文件
└── manage.py
```

## 🔧 关键配置说明

### 1. Django设置 (settings.py)

```python
# 静态文件配置
STATIC_URL = '/static/'
STATICFILES_DIRS = [
    BASE_DIR / 'static',
    BASE_DIR / 'frontend' / 'dist',  # React构建目录
]

# 模板配置
TEMPLATES[0]['DIRS'] = [BASE_DIR / 'frontend' / 'dist']
```

### 2. Django路由 (urls.py)

```python
from django.views.generic import TemplateView
from django.views.decorators.csrf import ensure_csrf_cookie

class ReactAppView(TemplateView):
    template_name = 'index.html'
    
    @ensure_csrf_cookie
    def get(self, request, *args, **kwargs):
        return super().get(request, *args, **kwargs)

urlpatterns = [
    # API路由
    path('api/workflow/', include('workflow.urls')),
    
    # React前端路由 - 捕获所有非API路由
    re_path(r'^(?!api/|admin/|static/|media/).*$', ReactAppView.as_view()),
]
```

### 3. React模板 (frontend/dist/index.html)

```html
{% load static %}
<!doctype html>
<html lang="zh-CN">
  <head>
    <!-- Django模板标签 -->
    <meta name="csrf-token" content="{{ csrf_token }}" />
    <script src="{% static 'assets/index.js' %}"></script>
  </head>
  <body>
    <div id="root"></div>
    <script>
      // Django注入的全局变量
      window.CSRF_TOKEN = "{{ csrf_token }}";
      window.API_BASE_URL = "/api/workflow";
    </script>
  </body>
</html>
```

### 4. React API配置 (src/utils/api.ts)

```typescript
// 使用相对路径，自动适配Django路由
const API_BASE_URL = '/api/workflow';

// 自动获取CSRF Token
const getCSRFToken = (): string => {
  // 从Django模板注入的window对象获取
  if ((window as any).CSRF_TOKEN) {
    return (window as any).CSRF_TOKEN;
  }
  // 从cookie获取
  const match = document.cookie.match(/csrftoken=([^;]+)/);
  return match ? match[1] : '';
};

// 自动添加到请求头
apiClient.interceptors.request.use((config) => {
  const csrfToken = getCSRFToken();
  if (csrfToken) {
    config.headers['X-CSRFToken'] = csrfToken;
  }
  return config;
});
```

## 🚀 开发工作流程

### 开发模式 (前后端分离)

1. **启动Django后端**
   ```bash
   cd /Users/liunian/Desktop/dnmp/py_project
   source venv/bin/activate
   python manage.py runserver 8000
   ```

2. **启动React开发服务器**
   ```bash
   cd /Users/liunian/Desktop/dnmp/py_project/frontend
   npm run dev
   ```
   - React运行在 http://localhost:5173
   - API请求通过Vite代理到Django

### 生产模式 (Django统一服务)

1. **构建React应用**
   ```bash
   cd /Users/liunian/Desktop/dnmp/py_project/frontend
   npm run build
   ```

2. **启动Django服务器**
   ```bash
   cd /Users/liunian/Desktop/dnmp/py_project
   python manage.py runserver 8000
   ```
   - 访问 http://localhost:8000
   - Django同时服务API和React前端

## 📡 API端点

### ETF相关
- `GET /api/workflow/api/etf/list/` - 获取ETF列表
- `GET /api/workflow/api/etf/<symbol>/realtime/` - 获取实时数据
- `GET /api/workflow/api/etf/<symbol>/history/` - 获取历史数据
- `POST /api/workflow/api/etf/portfolio/analyze/` - 分析投资组合

### 数据更新
- `POST /api/workflow/api/update-realtime/` - 更新实时数据
- `POST /api/workflow/api/update-exchange-rates/` - 更新汇率

### 配置管理
- `GET /api/workflow/api/portfolio-configs/` - 获取配置列表
- `POST /api/workflow/api/portfolio-configs/<id>/` - 保存配置

## 🔒 安全特性

1. **CSRF保护**
   - Django模板自动注入CSRF token
   - React自动从window对象或cookie获取
   - 所有POST/PUT/DELETE请求自动携带

2. **CORS配置**
   - 开发模式允许所有来源
   - 生产模式应限制为特定域名

3. **会话管理**
   - 使用Django的session框架
   - 支持用户认证和权限控制

## 🎨 前端特性

### 富途风格UI
- 专业金融配色（红涨绿跌）
- 实时行情卡片
- 交互式图表
- 响应式布局

### 技术栈
- React 18 + TypeScript
- Ant Design 组件库
- Styled Components
- ECharts 图表
- Axios HTTP客户端

## 📝 注意事项

1. **构建后不要修改dist/index.html**
   - 每次`npm run build`会覆盖它
   - 使用模板保留Django标签

2. **静态文件收集**
   - 生产环境运行 `python manage.py collectstatic`
   - 确保nginx/apache正确配置静态文件服务

3. **API路径**
   - 始终以 `/api/` 开头
   - 避免与React路由冲突

4. **开发vs生产**
   - 开发: 使用Vite代理
   - 生产: Django直接服务

## 🔧 故障排除

### 问题1: CSRF验证失败
**解决**: 确保访问的是Django服务的URL (localhost:8000)，不是Vite的URL

### 问题2: 静态文件404
**解决**: 检查 `STATICFILES_DIRS` 配置，确保包含 frontend/dist

### 问题3: React路由刷新404
**解决**: Django的catch-all路由已处理，确保正则表达式正确

### 问题4: API请求失败
**解决**: 检查浏览器开发者工具的Network面板，确认请求URL和CSRF头

## 🎉 完成！

现在您拥有一个完整的Django + React整合应用：
- ✅ Django提供API和静态文件服务
- ✅ React提供现代化前端界面
- ✅ 自动CSRF保护
- ✅ 富途风格的专业UI
- ✅ 单域名部署，无需跨域
