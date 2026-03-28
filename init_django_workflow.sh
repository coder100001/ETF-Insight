#!/bin/bash

echo "================================"
echo "Django ETF工作流系统 - 初始化脚本"
echo "================================"

# 激活虚拟环境
source venv/bin/activate

# 1. 生成数据库迁移文件
echo ">>> 1. 生成数据库迁移文件..."
python manage.py makemigrations

# 2. 执行数据库迁移
echo ">>> 2. 执行数据库迁移..."
python manage.py migrate

# 3. 创建超级用户（可选）
echo ">>> 3. 是否创建管理员账号？(y/n)"
read -p "请输入: " create_admin
if [ "$create_admin" == "y" ]; then
    python manage.py createsuperuser
fi

# 4. 初始化测试数据
echo ">>> 4. 初始化测试数据..."
python manage.py shell <<EOF
from workflow.models import Workflow, WorkflowStep
from django.utils import timezone
import json

# 创建工作流1: ETF数据日度采集
wf1 = Workflow.objects.create(
    name='ETF数据日度采集',
    description='SCHD/SPYD/VYMI/SCHY四只ETF的每日数据采集',
    category='data_collection',
    status=1,
    trigger_type=1,
    trigger_config={'cron': '0 18 * * 1-5', 'timezone': 'America/New_York'}
)

# 添加步骤
WorkflowStep.objects.create(
    workflow=wf1, name='初始化环境', step_type='init', order_index=1,
    handler_type=1, handler_config={'script': 'init_env.py'},
    retry_times=3, timeout=60, is_critical=True, on_failure='stop'
)

for idx, symbol in enumerate(['SCHD', 'SPYD', 'VYMI', 'SCHY'], start=2):
    WorkflowStep.objects.create(
        workflow=wf1, name=f'拉取{symbol}数据', step_type='fetch_data', order_index=idx,
        handler_type=2, handler_config={'function': 'fetch_etf_data', 'params': {'symbol': symbol}},
        retry_times=5, retry_interval=5, timeout=120, is_critical=True, on_failure='retry'
    )

WorkflowStep.objects.create(
    workflow=wf1, name='数据质量检查', step_type='validate', order_index=6,
    handler_type=2, handler_config={'function': 'validate_data'},
    retry_times=3, timeout=60, is_critical=True, on_failure='stop'
)

WorkflowStep.objects.create(
    workflow=wf1, name='保存到数据库', step_type='save', order_index=7,
    handler_type=2, handler_config={'function': 'save_to_database'},
    retry_times=3, timeout=120, is_critical=True, on_failure='stop'
)

# 创建工作流2: 投资组合每日分析
wf2 = Workflow.objects.create(
    name='投资组合每日分析',
    description='基于最新数据进行投资组合收益和风险分析',
    category='analysis',
    status=1,
    trigger_type=1,
    trigger_config={'cron': '0 19 * * 1-5', 'timezone': 'America/New_York'}
)

print('✓ 初始化数据完成！')
print(f'  - 创建工作流: {Workflow.objects.count()} 个')
print(f'  - 创建步骤: {WorkflowStep.objects.count()} 个')
EOF

echo ""
echo "================================"
echo "初始化完成！"
echo "================================"
echo "访问方式："
echo "  - Web界面: http://localhost:8000"
echo "  - 管理后台: http://localhost:8000/admin"
echo ""
echo "启动服务："
echo "  python manage.py runserver 0.0.0.0:8000"
echo "================================"
