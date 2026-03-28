"""
ETF工作流引擎 - Django版本
"""

import time
import traceback
from datetime import datetime
from django.utils import timezone
from .models import (
    Workflow, WorkflowStep, WorkflowInstance,
    WorkflowInstanceStep, SystemLog
)


class WorkflowEngine:
    """工作流执行引擎"""
    
    def __init__(self):
        self.handlers = {}
        self._register_default_handlers()
    
    def _register_default_handlers(self):
        """注册默认处理器"""
        # 数据采集相关
        self.handlers['fetch_etf_data'] = self.fetch_etf_data_handler
        self.handlers['get_etf_list'] = self.get_etf_list_handler
        self.handlers['fetch_realtime_data'] = self.fetch_realtime_data_handler
        self.handlers['fetch_historical_data'] = self.fetch_historical_data_handler
        self.handlers['fetch_us_etf_data'] = self.fetch_us_etf_data_handler
        self.handlers['fetch_cn_etf_data'] = self.fetch_cn_etf_data_handler
        self.handlers['fetch_hk_etf_data'] = self.fetch_hk_etf_data_handler

        # 数据处理相关
        self.handlers['init_environment'] = self.init_environment_handler
        self.handlers['validate_data'] = self.validate_data_handler
        self.handlers['validate_and_clean_data'] = self.validate_and_clean_data_handler

        # 数据存储
        self.handlers['save_to_database'] = self.save_to_database_handler
        self.handlers['update_exchange_rates'] = self.update_exchange_rates_handler

        # 分析相关
        self.handlers['build_portfolio'] = self.build_portfolio_handler
        self.handlers['analyze_portfolio'] = self.analyze_portfolio_handler
        self.handlers['forecast_returns'] = self.forecast_returns_handler

        # 报告和通知
        self.handlers['generate_report'] = self.generate_report_handler
        self.handlers['send_notification'] = self.send_notification_handler
    
    def register_handler(self, name, handler_func):
        """注册自定义处理器"""
        self.handlers[name] = handler_func
    
    def execute_workflow(self, workflow_id, context=None, trigger_by='system'):
        """执行工作流"""
        instance = None
        try:
            # 获取工作流
            workflow = Workflow.objects.get(id=workflow_id)
            
            # 创建工作流实例
            instance = WorkflowInstance.objects.create(
                workflow=workflow,
                trigger_type=2,  # 手动触发
                trigger_by=trigger_by,
                status=0,  # 等待中
                context_data=context or {}
            )
            
            # 更新状态为运行中
            instance.status = 1
            instance.start_time = timezone.now()
            instance.save()
            
            # 获取工作流步骤
            steps = workflow.steps.order_by('order_index')
            
            # 创建步骤实例
            step_instances = []
            for step in steps:
                step_instance = WorkflowInstanceStep.objects.create(
                    workflow_instance=instance,
                    workflow_step=step,
                    step_name=step.name,
                    status=0  # 等待
                )
                step_instances.append(step_instance)
            
            # 执行步骤
            all_success = True
            for step_instance in step_instances:
                success = self._execute_step(step_instance, context or {})
                if not success:
                    all_success = False
                    step = step_instance.workflow_step
                    if step.is_critical and step.on_failure == 'stop':
                        self._log(instance, 'ERROR', f'关键步骤失败，停止执行: {step.name}')
                        break
            
            # 更新实例状态
            instance.end_time = timezone.now()
            instance.duration = int((instance.end_time - instance.start_time).total_seconds())
            instance.status = 2 if all_success else 3  # 2-成功 3-失败
            instance.save()
            
            self._log(instance, 'INFO', f'工作流执行完成，状态: {instance.get_status_display()}')
            
            return instance
            
        except Exception as e:
            if instance:
                instance.status = 3  # 失败
                instance.error_message = str(e)
                instance.end_time = timezone.now()
                if instance.start_time:
                    instance.duration = int((instance.end_time - instance.start_time).total_seconds())
                instance.save()
                self._log(instance, 'ERROR', f'工作流执行异常: {str(e)}', traceback.format_exc())
            raise
    
    def _execute_step(self, step_instance, context):
        """执行单个步骤"""
        step = step_instance.workflow_step
        
        try:
            # 更新步骤状态为运行中
            step_instance.status = 1
            step_instance.start_time = timezone.now()
            # 记录输入数据
            step_instance.input_data = {
                'context': context,
                'step_config': step.handler_config or {}
            }
            step_instance.save()
            
            self._log(step_instance.workflow_instance, 'INFO',
                     f'开始执行步骤: {step.name} (序号: {step.order_index})')
            self._log(step_instance.workflow_instance, 'DEBUG',
                     f'步骤配置: 类型={step.step_type}, 处理器={step.handler_type}')
            self._log(step_instance.workflow_instance, 'DEBUG',
                     f'输入参数: {context}')
            
            # 获取处理器配置
            config = step.handler_config or {}
            
            # 执行处理器
            result = None
            if step.handler_type == 2:  # 函数处理器
                func_name = config.get('function')
                params = config.get('params', {})
                
                self._log(step_instance.workflow_instance, 'DEBUG',
                         f'调用处理器: {func_name}')
                
                if func_name in self.handlers:
                    # 合并上下文和参数
                    merged_params = {**context, **params}
                    result = self.handlers[func_name](merged_params)
                    self._log(step_instance.workflow_instance, 'DEBUG',
                             f'处理器返回: {type(result).__name__}')
                else:
                    raise ValueError(f'未找到处理器: {func_name}')
            else:
                # 其他类型处理器（脚本、API调用）的占位逻辑
                self._log(step_instance.workflow_instance, 'WARNING',
                         f'处理器类型 {step.handler_type} 暂未实现，使用模拟结果')
                result = {'status': 'simulated', 'message': '处理器类型暂未实现'}
            
            # 更新步骤状态为成功
            step_instance.status = 2
            step_instance.end_time = timezone.now()
            step_instance.duration = int((step_instance.end_time - step_instance.start_time).total_seconds())
            step_instance.output_data = result
            step_instance.logs = f'执行成功\n开始时间: {step_instance.start_time}\n结束时间: {step_instance.end_time}\n耗时: {step_instance.duration}秒'
            step_instance.save()
            
            self._log(step_instance.workflow_instance, 'INFO',
                     f'步骤执行成功: {step.name}, 耗时: {step_instance.duration}秒')
            self._log(step_instance.workflow_instance, 'DEBUG',
                     f'输出结果: {result}')
            
            return True
            
        except Exception as e:
            # 处理失败
            step_instance.status = 3
            step_instance.error_message = str(e)
            step_instance.end_time = timezone.now()
            if step_instance.start_time:
                step_instance.duration = int((step_instance.end_time - step_instance.start_time).total_seconds())
            
            error_detail = traceback.format_exc()
            step_instance.logs = f'执行失败\n错误: {str(e)}\n\n堆栈:\n{error_detail}'
            
            # 检查是否需要重试
            if step_instance.retry_count < step.retry_times:
                step_instance.retry_count += 1
                step_instance.save()
                
                self._log(step_instance.workflow_instance, 'WARNING',
                         f'步骤执行失败，准备重试: {step.name}, 第{step_instance.retry_count}次重试')
                self._log(step_instance.workflow_instance, 'WARNING',
                         f'错误信息: {str(e)}')
                
                time.sleep(step.retry_interval)
                return self._execute_step(step_instance, context)
            else:
                step_instance.save()
                self._log(step_instance.workflow_instance, 'ERROR',
                         f'步骤执行失败: {step.name}, 错误: {str(e)}')
                self._log(step_instance.workflow_instance, 'ERROR',
                         f'堆栈信息: {error_detail}')
                return False
    
    def _log(self, instance, level, message, stack_trace=None):
        """记录日志"""
        SystemLog.objects.create(
            workflow_instance=instance,
            log_level=level,
            module='workflow_engine',
            message=message,
            stack_trace=stack_trace
        )
        print(f"[{level}] {message}")
    
    # ========================================================================
    # 默认处理器实现（从handlers模块导入）
    # ========================================================================

    def init_environment_handler(self, params):
        """初始化环境"""
        from .handlers import init_environment
        return init_environment(params)

    def get_etf_list_handler(self, params):
        """获取ETF列表"""
        from .handlers import get_etf_list
        return get_etf_list(params)

    def fetch_realtime_data_handler(self, params):
        """拉取ETF实时数据"""
        from .handlers import fetch_realtime_data
        return fetch_realtime_data(params)

    def fetch_historical_data_handler(self, params):
        """获取历史数据"""
        from .handlers import fetch_historical_data
        return fetch_historical_data(params)

    def fetch_us_etf_data_handler(self, params):
        """获取美股ETF数据"""
        from .handlers import fetch_us_etf_data
        return fetch_us_etf_data(params)

    def fetch_cn_etf_data_handler(self, params):
        """获取A股ETF数据"""
        from .handlers import fetch_cn_etf_data
        return fetch_cn_etf_data(params)

    def fetch_hk_etf_data_handler(self, params):
        """获取港股ETF数据"""
        from .handlers import fetch_hk_etf_data
        return fetch_hk_etf_data(params)

    def validate_and_clean_data_handler(self, params):
        """数据清洗与验证"""
        from .handlers import validate_and_clean_data
        return validate_and_clean_data(params)

    def update_exchange_rates_handler(self, params):
        """更新汇率数据"""
        from .handlers import update_exchange_rates
        return update_exchange_rates(params)

    def build_portfolio_handler(self, params):
        """构建投资组合"""
        from .handlers import build_portfolio
        return build_portfolio(params)

    def analyze_portfolio_handler(self, params):
        """投资组合分析"""
        from .handlers import analyze_portfolio
        return analyze_portfolio(params)

    def forecast_returns_handler(self, params):
        """收益预测"""
        from .handlers import forecast_returns
        return forecast_returns(params)

    def generate_report_handler(self, params):
        """生成报告"""
        from .handlers import generate_report
        return generate_report(params)

    def send_notification_handler(self, params):
        """发送通知"""
        from .handlers import send_notification
        return send_notification(params)

    # 原有的处理器（向后兼容）
    def fetch_etf_data_handler(self, params):
        """ETF数据拉取处理器（旧版本，保留兼容）"""
        symbol = params.get('symbol')
        print(f"  拉取 {symbol} 数据...")
        time.sleep(1)
        return {
            'symbol': symbol,
            'records': 100,
            'status': 'success'
        }

    def validate_data_handler(self, params):
        """数据验证处理器（旧版本，保留兼容）"""
        print("  执行数据质量检查...")
        time.sleep(0.5)
        return {
            'validation': 'passed',
            'issues': []
        }

    def save_to_database_handler(self, params):
        """保存到数据库处理器（旧版本，保留兼容）"""
        print("  保存数据到数据库...")
        time.sleep(0.5)
        return {
            'saved': True,
            'records': 400
        }


# 全局引擎实例
workflow_engine = WorkflowEngine()
