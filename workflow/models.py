"""
ETF工作流系统 - Django Model层
"""

from django.db import models
from django.utils import timezone


class Workflow(models.Model):
    """工作流定义"""
    STATUS_CHOICES = [
        (0, '禁用'),
        (1, '启用'),
        (2, '已归档'),
    ]
    
    TRIGGER_TYPE_CHOICES = [
        (1, '定时'),
        (2, '手动'),
        (3, '事件'),
    ]
    
    name = models.CharField(max_length=100, verbose_name='工作流名称')
    description = models.TextField(null=True, blank=True, verbose_name='描述')
    category = models.CharField(max_length=50, null=True, blank=True, verbose_name='分类')
    status = models.IntegerField(choices=STATUS_CHOICES, default=1, verbose_name='状态')
    trigger_type = models.IntegerField(choices=TRIGGER_TYPE_CHOICES, null=True, blank=True, verbose_name='触发类型')
    trigger_config = models.JSONField(null=True, blank=True, verbose_name='触发配置')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'workflow'
        verbose_name = '工作流'
        verbose_name_plural = '工作流'
        ordering = ['-id']
    
    def __str__(self):
        return self.name


class WorkflowStep(models.Model):
    """工作流步骤"""
    HANDLER_TYPE_CHOICES = [
        (1, '脚本'),
        (2, '函数'),
        (3, 'API调用'),
    ]
    
    workflow = models.ForeignKey(Workflow, on_delete=models.CASCADE, related_name='steps', verbose_name='工作流')
    name = models.CharField(max_length=100, verbose_name='步骤名称')
    step_type = models.CharField(max_length=50, verbose_name='步骤类型')
    order_index = models.IntegerField(verbose_name='执行顺序')
    handler_type = models.IntegerField(choices=HANDLER_TYPE_CHOICES, verbose_name='处理器类型')
    handler_config = models.JSONField(null=True, blank=True, verbose_name='处理器配置')
    retry_times = models.IntegerField(default=3, verbose_name='重试次数')
    retry_interval = models.IntegerField(default=5, verbose_name='重试间隔(秒)')
    timeout = models.IntegerField(default=300, verbose_name='超时时间(秒)')
    is_critical = models.BooleanField(default=False, verbose_name='是否关键步骤')
    on_failure = models.CharField(max_length=50, null=True, blank=True, verbose_name='失败处理')
    depends_on = models.JSONField(null=True, blank=True, verbose_name='依赖步骤')
    extra_config = models.JSONField(null=True, blank=True, verbose_name='额外配置')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'workflow_step'
        verbose_name = '工作流步骤'
        verbose_name_plural = '工作流步骤'
        ordering = ['workflow', 'order_index']
    
    def __str__(self):
        return f"{self.workflow.name} - {self.name}"


class WorkflowInstance(models.Model):
    """工作流实例"""
    STATUS_CHOICES = [
        (0, '等待中'),
        (1, '运行中'),
        (2, '成功'),
        (3, '失败'),
        (4, '已取消'),
    ]
    
    workflow = models.ForeignKey(Workflow, on_delete=models.CASCADE, related_name='instances', verbose_name='工作流')
    trigger_type = models.IntegerField(null=True, blank=True, verbose_name='触发方式')
    trigger_by = models.CharField(max_length=100, null=True, blank=True, verbose_name='触发人')
    status = models.IntegerField(choices=STATUS_CHOICES, default=0, verbose_name='状态')
    start_time = models.DateTimeField(null=True, blank=True, verbose_name='开始时间')
    end_time = models.DateTimeField(null=True, blank=True, verbose_name='结束时间')
    duration = models.IntegerField(null=True, blank=True, verbose_name='执行时长(秒)')
    context_data = models.JSONField(null=True, blank=True, verbose_name='上下文数据')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误信息')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'workflow_instance'
        verbose_name = '工作流实例'
        verbose_name_plural = '工作流实例'
        ordering = ['-created_at']
    
    def __str__(self):
        return f"实例#{self.id} - {self.workflow.name}"
    
    @property
    def status_display(self):
        return self.get_status_display()
    
    @property
    def status_class(self):
        """返回Bootstrap badge样式类"""
        class_map = {
            0: 'secondary',
            1: 'primary',
            2: 'success',
            3: 'danger',
            4: 'warning',
        }
        return class_map.get(self.status, 'secondary')
    
    @property
    def status_label(self):
        """返回状态标签"""
        return self.get_status_display()
    
    @property
    def workflow_name(self):
        """返回工作流名称（兼容模板）"""
        return self.workflow.name
    
    @property
    def status_class(self):
        """返回Bootstrap badge样式类"""
        class_map = {
            0: 'secondary',
            1: 'primary',
            2: 'success',
            3: 'danger',
            4: 'warning',
        }
        return class_map.get(self.status, 'secondary')
    
    @property
    def status_label(self):
        """返回状态标签"""
        return self.get_status_display()


class WorkflowInstanceStep(models.Model):
    """工作流实例步骤"""
    STATUS_CHOICES = [
        (0, '等待'),
        (1, '运行中'),
        (2, '成功'),
        (3, '失败'),
        (4, '跳过'),
    ]
    
    workflow_instance = models.ForeignKey(WorkflowInstance, on_delete=models.CASCADE, 
                                         related_name='step_instances', verbose_name='工作流实例')
    workflow_step = models.ForeignKey(WorkflowStep, on_delete=models.CASCADE, verbose_name='步骤定义')
    step_name = models.CharField(max_length=100, null=True, blank=True, verbose_name='步骤名称')
    status = models.IntegerField(choices=STATUS_CHOICES, default=0, verbose_name='状态')
    retry_count = models.IntegerField(default=0, verbose_name='已重试次数')
    assigned_to = models.BigIntegerField(null=True, blank=True, verbose_name='分配给')
    start_time = models.DateTimeField(null=True, blank=True, verbose_name='开始时间')
    end_time = models.DateTimeField(null=True, blank=True, verbose_name='结束时间')
    duration = models.IntegerField(null=True, blank=True, verbose_name='执行时长(秒)')
    input_data = models.JSONField(null=True, blank=True, verbose_name='输入数据')
    output_data = models.JSONField(null=True, blank=True, verbose_name='输出结果')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误信息')
    logs = models.TextField(null=True, blank=True, verbose_name='执行日志')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'workflow_instance_step'
        verbose_name = '实例步骤'
        verbose_name_plural = '实例步骤'
        ordering = ['workflow_instance', 'created_at']
    
    def __str__(self):
        return f"{self.workflow_instance} - {self.step_name}"
    
    @property
    def name(self):
        """返回步骤名称（兼容模板）"""
        return self.step_name or self.workflow_step.name
    
    @property
    def status_class(self):
        """返回Bootstrap badge样式类"""
        class_map = {
            0: 'secondary',
            1: 'primary',
            2: 'success',
            3: 'danger',
            4: 'info',
        }
        return class_map.get(self.status, 'secondary')
    
    @property
    def status_label(self):
        """返回状态标签"""
        return self.get_status_display()


class Notification(models.Model):
    """通知记录"""
    TYPE_CHOICES = [
        (1, '邮件'),
        (2, '短信'),
        (3, 'APP推送'),
        (4, 'Webhook'),
    ]
    
    STATUS_CHOICES = [
        (0, '待发送'),
        (1, '已发送'),
        (2, '发送失败'),
    ]
    
    workflow_instance_step = models.ForeignKey(WorkflowInstanceStep, null=True, blank=True,
                                              on_delete=models.CASCADE, related_name='notifications',
                                              verbose_name='实例步骤')
    workflow_instance = models.ForeignKey(WorkflowInstance, null=True, blank=True,
                                         on_delete=models.CASCADE, related_name='notifications',
                                         verbose_name='工作流实例')
    notification_type = models.IntegerField(choices=TYPE_CHOICES, verbose_name='通知类型')
    recipient = models.CharField(max_length=200, null=True, blank=True, verbose_name='接收人')
    title = models.CharField(max_length=200, null=True, blank=True, verbose_name='通知标题')
    content = models.TextField(null=True, blank=True, verbose_name='通知内容')
    status = models.IntegerField(choices=STATUS_CHOICES, default=0, verbose_name='状态')
    server_id = models.CharField(max_length=100, null=True, blank=True, verbose_name='服务器标识')
    send_at = models.DateTimeField(null=True, blank=True, verbose_name='发送时间')
    retry_count = models.IntegerField(default=0, verbose_name='重试次数')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误信息')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    
    class Meta:
        db_table = 'notification'
        verbose_name = '通知记录'
        verbose_name_plural = '通知记录'
        ordering = ['-created_at']
    
    def __str__(self):
        return f"{self.get_notification_type_display()} - {self.title}"


class SystemLog(models.Model):
    """系统日志"""
    LEVEL_CHOICES = [
        ('DEBUG', 'DEBUG'),
        ('INFO', 'INFO'),
        ('WARNING', 'WARNING'),
        ('ERROR', 'ERROR'),
        ('CRITICAL', 'CRITICAL'),
    ]
    
    workflow_instance = models.ForeignKey(WorkflowInstance, null=True, blank=True,
                                         on_delete=models.CASCADE, related_name='logs',
                                         verbose_name='工作流实例')
    workflow_instance_step = models.ForeignKey(WorkflowInstanceStep, null=True, blank=True,
                                              on_delete=models.CASCADE, related_name='system_logs',
                                              verbose_name='实例步骤')
    log_level = models.CharField(max_length=20, choices=LEVEL_CHOICES, verbose_name='日志级别')
    module = models.CharField(max_length=100, null=True, blank=True, verbose_name='模块名称')
    message = models.TextField(verbose_name='日志消息')
    stack_trace = models.TextField(null=True, blank=True, verbose_name='堆栈信息')
    extra_data = models.JSONField(null=True, blank=True, verbose_name='额外数据')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    
    class Meta:
        db_table = 'system_log'
        verbose_name = '系统日志'
        verbose_name_plural = '系统日志'
        ordering = ['-created_at']
    
    def __str__(self):
        return f"[{self.log_level}] {self.message[:50]}"
    
    @property
    def level(self):
        """返回日志级别（兼容模板）"""
        return self.log_level
    
    @property
    def time(self):
        """返回日志时间（兼容模板）"""
        return self.created_at


class ETFData(models.Model):
    """ETF数据"""
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')
    date = models.DateField(verbose_name='日期')
    open_price = models.DecimalField(max_digits=10, decimal_places=4, null=True, blank=True, verbose_name='开盘价')
    close_price = models.DecimalField(max_digits=10, decimal_places=4, null=True, blank=True, verbose_name='收盘价')
    high_price = models.DecimalField(max_digits=10, decimal_places=4, null=True, blank=True, verbose_name='最高价')
    low_price = models.DecimalField(max_digits=10, decimal_places=4, null=True, blank=True, verbose_name='最低价')
    volume = models.BigIntegerField(null=True, blank=True, verbose_name='成交量')
    dividend = models.DecimalField(max_digits=10, decimal_places=4, null=True, blank=True, verbose_name='股息')
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    fetch_instance = models.ForeignKey(WorkflowInstance, null=True, blank=True,
                                      on_delete=models.SET_NULL, verbose_name='采集实例')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    
    class Meta:
        db_table = 'etf_data'
        verbose_name = 'ETF数据'
        verbose_name_plural = 'ETF数据'
        unique_together = [['symbol', 'date']]
        ordering = ['-date', 'symbol']
    
    def __str__(self):
        return f"{self.symbol} - {self.date}"


class PortfolioConfig(models.Model):
    """投资组合配置"""
    STATUS_CHOICES = [
        (0, '禁用'),
        (1, '启用'),
    ]
    
    name = models.CharField(max_length=100, verbose_name='组合名称')
    description = models.TextField(null=True, blank=True, verbose_name='组合描述')
    allocation = models.JSONField(verbose_name='配置比例')
    total_investment = models.DecimalField(max_digits=15, decimal_places=2, null=True, blank=True, verbose_name='投资金额')
    status = models.IntegerField(choices=STATUS_CHOICES, default=1, verbose_name='状态')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'portfolio_config'
        verbose_name = '投资组合配置'
        verbose_name_plural = '投资组合配置'
        ordering = ['-created_at']
    
    def __str__(self):
        return self.name


class AnalysisReport(models.Model):
    """分析报告"""
    STATUS_CHOICES = [
        (1, '正常'),
        (2, '已归档'),
    ]
    
    workflow_instance = models.ForeignKey(WorkflowInstance, null=True, blank=True,
                                         on_delete=models.SET_NULL, verbose_name='工作流实例')
    portfolio_config = models.ForeignKey(PortfolioConfig, null=True, blank=True,
                                        on_delete=models.SET_NULL, verbose_name='组合配置')
    report_type = models.CharField(max_length=50, null=True, blank=True, verbose_name='报告类型')
    report_date = models.DateField(verbose_name='报告日期')
    file_path = models.CharField(max_length=500, null=True, blank=True, verbose_name='报告文件路径')
    metrics = models.JSONField(null=True, blank=True, verbose_name='关键指标')
    status = models.IntegerField(choices=STATUS_CHOICES, default=1, verbose_name='状态')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    
    class Meta:
        db_table = 'analysis_report'
        verbose_name = '分析报告'
        verbose_name_plural = '分析报告'
        ordering = ['-report_date']
    
    def __str__(self):
        return f"{self.report_type} - {self.report_date}"


class OperationLog(models.Model):
    """操作记录 - 记录所有系统操作"""
    OPERATION_TYPE_CHOICES = [
        ('data_update', '数据更新'),
        ('portfolio_analysis', '组合分析'),
        ('etf_comparison', 'ETF对比'),
        ('script_execute', '脚本执行'),
        ('scheduled_task', '定时任务'),
        ('manual_trigger', '手动触发'),
        ('page_view', '页面访问'),
    ]
    
    STATUS_CHOICES = [
        (0, '进行中'),
        (1, '成功'),
        (2, '失败'),
    ]
    
    workflow_instance = models.ForeignKey(WorkflowInstance, null=True, blank=True,
                                         on_delete=models.SET_NULL, related_name='operation_logs',
                                         verbose_name='工作流实例')
    operation_type = models.CharField(max_length=50, choices=OPERATION_TYPE_CHOICES, verbose_name='操作类型')
    operation_name = models.CharField(max_length=200, verbose_name='操作名称')
    operator = models.CharField(max_length=100, null=True, blank=True, verbose_name='操作人')
    status = models.IntegerField(choices=STATUS_CHOICES, default=0, verbose_name='状态')
    start_time = models.DateTimeField(auto_now_add=True, verbose_name='开始时间')
    end_time = models.DateTimeField(null=True, blank=True, verbose_name='结束时间')
    duration_ms = models.IntegerField(null=True, blank=True, verbose_name='耗时(毫秒)')
    input_params = models.JSONField(null=True, blank=True, verbose_name='输入参数')
    output_result = models.JSONField(null=True, blank=True, verbose_name='输出结果')
    error_message = models.TextField(null=True, blank=True, verbose_name='错误信息')
    ip_address = models.CharField(max_length=50, null=True, blank=True, verbose_name='IP地址')
    user_agent = models.CharField(max_length=500, null=True, blank=True, verbose_name='用户代理')
    extra_data = models.JSONField(null=True, blank=True, verbose_name='额外数据')
    
    class Meta:
        db_table = 'operation_log'
        verbose_name = '操作记录'
        verbose_name_plural = '操作记录'
        ordering = ['-start_time']
        indexes = [
            models.Index(fields=['operation_type', 'start_time']),
            models.Index(fields=['status', 'start_time']),
        ]
    
    def __str__(self):
        return f"[{self.get_operation_type_display()}] {self.operation_name}"
    
    def complete(self, success=True, result=None, error=None):
        """完成操作"""
        from django.utils import timezone
        self.end_time = timezone.now()
        self.status = 1 if success else 2
        if result:
            self.output_result = result
        if error:
            self.error_message = str(error)
        if self.start_time:
            self.duration_ms = int((self.end_time - self.start_time).total_seconds() * 1000)
        self.save()


class ExchangeRate(models.Model):
    """汇率表 - 存储人民币、港币、美元之间的汇率（以美元为基准）"""
    CURRENCY_CHOICES = [
        ('USD', '美元'),
        ('CNY', '人民币'),
        ('HKD', '港币'),
    ]
    
    from_currency = models.CharField(max_length=10, choices=CURRENCY_CHOICES, verbose_name='源货币')
    to_currency = models.CharField(max_length=10, choices=CURRENCY_CHOICES, verbose_name='目标货币')
    rate = models.DecimalField(max_digits=15, decimal_places=6, verbose_name='汇率')
    rate_date = models.DateField(verbose_name='汇率日期')
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'exchange_rate'
        verbose_name = '汇率'
        verbose_name_plural = '汇率'
        unique_together = [['from_currency', 'to_currency', 'rate_date']]
        ordering = ['-rate_date', 'from_currency', 'to_currency']
        indexes = [
            models.Index(fields=['from_currency', 'to_currency']),
            models.Index(fields=['-rate_date']),
        ]
    
    def __str__(self):
        return f"1 {self.from_currency} = {self.rate} {self.to_currency}"


class ETFConfig(models.Model):
    """柯ETF配置 - 管理系统支持的ETF列表"""
    MARKET_CHOICES = [
        ('US', '美股'),
        ('CN', 'A股'),
        ('HK', '港股'),
    ]
    
    STATUS_CHOICES = [
        (0, '禁用'),
        (1, '启用'),
    ]
    
    symbol = models.CharField(max_length=20, unique=True, verbose_name='ETF代码')
    name = models.CharField(max_length=200, verbose_name='ETF名称')
    market = models.CharField(max_length=10, choices=MARKET_CHOICES, default='US', verbose_name='市场')
    strategy = models.CharField(max_length=100, null=True, blank=True, verbose_name='策略类型')
    description = models.TextField(null=True, blank=True, verbose_name='描述')
    focus = models.CharField(max_length=50, null=True, blank=True, verbose_name='焦点领域')
    expense_ratio = models.DecimalField(max_digits=5, decimal_places=4, null=True, blank=True, verbose_name='费率(%)')
    status = models.IntegerField(choices=STATUS_CHOICES, default=1, verbose_name='状态')
    sort_order = models.IntegerField(default=0, verbose_name='排序')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')
    updated_at = models.DateTimeField(auto_now=True, verbose_name='更新时间')
    
    class Meta:
        db_table = 'etf_config'
        verbose_name = 'ETF配置'
        verbose_name_plural = 'ETF配置'
        ordering = ['market', 'sort_order', 'symbol']
    
    def __str__(self):
        return f"{self.symbol} - {self.name}"
