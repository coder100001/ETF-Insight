"""etf_workflow_project URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/3.2/topics/http/urls/
"""
from django.contrib import admin
from django.urls import path, include, re_path
from django.views.generic import TemplateView
from django.views.decorators.csrf import ensure_csrf_cookie
from django.http import JsonResponse
from django.shortcuts import render

# API健康检查
def api_health_check(request):
    return JsonResponse({'status': 'ok', 'message': 'API is running'})

# 确保CSRF cookie被设置的React入口视图
@ensure_csrf_cookie
def react_app_view(request):
    """渲染React前端应用"""
    return render(request, 'index.html')

urlpatterns = [
    # Django Admin
    path('admin/', admin.site.urls),
    
    # API路由
    path('api/health/', api_health_check, name='health_check'),
    path('api/workflow/', include('workflow.urls')),
    
    # React前端路由 - 捕获所有非API路由
    re_path(r'^(?!api/|admin/|static/|media/).*$', react_app_view, name='react_app'),
]