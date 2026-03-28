# 外汇管理功能快速开始指南

## 访问页面

```
http://localhost:8000/workflow/exchange-rates/
```

## 功能说明

### 1. 货币转换器
位于页面顶部，可以快速进行货币转换：
- 选择源货币（USD/CNY/HKD）
- 输入金额
- 选择目标货币
- 点击"转换"按钮
- 查看转换结果

### 2. 今日汇率
显示当日所有货币对汇率，包括：
- USD → CNY、HKD
- CNY → USD、HKD
- HKD → USD、CNY
- 对比昨日涨跌
- 涨跌幅百分比

### 3. 汇率走势图
使用 Chart.js 展示最近7天汇率走势：
- 三条曲线：USD/CNY（绿色）、USD/HKD（红色）、CNY/HKD（蓝色）
- 鼠标悬停查看详细数据

### 4. 历史记录
显示汇率历史数据：
- 按日期分组
- 可点击货币对按钮筛选
- 默认显示所有货币对

### 5. 更新汇率
点击"立即更新"按钮，更新所有汇率：
- 使用系统默认汇率
- 更新今日所有货币对
- 显示更新结果

## API 使用

### 货币转换
```bash
curl "http://localhost:8000/workflow/exchange-rates/convert/?from=USD&to=CNY&amount=1000"
```

### 汇率历史
```bash
curl "http://localhost:8000/workflow/exchange-rates/history/?from=USD&to=CNY&days=30"
```

### 更新汇率
```bash
curl "http://localhost:8000/workflow/exchange-rates/update/"
```

## Python 代码示例

```python
from workflow.services_exchange_rate import exchange_rate_service

# 获取汇率
rate = exchange_rate_service.get_latest_rate('USD', 'CNY')
print(f"1 USD = {rate} CNY")

# 货币转换
result = exchange_rate_service.convert(100, 'USD', 'CNY')
print(f"100 USD = {result} CNY")

# 获取历史
history = exchange_rate_service.get_history('USD', 'CNY', days=7)
for item in history:
    print(f"{item['date']}: {item['rate']}")
```

## 在工作流中使用

### 配置工作流步骤

```python
from workflow.models import WorkflowStep

# 在工作流中添加汇率更新步骤
WorkflowStep.objects.create(
    workflow=workflow,
    name='更新汇率数据',
    step_type='fetch',
    order_index=2,
    handler_type=2,
    handler_config={
        'function': 'update_exchange_rates',
        'params': {}
    },
    is_critical=True,
    retry_times=2
)
```

### 执行工作流
```python
from workflow.engine import workflow_engine
from workflow.models import Workflow

wf = Workflow.objects.get(name='我的工作流')
instance = workflow_engine.execute_workflow(wf.id)
```

## 快速参考

| 功能 | URL | 方法 |
|------|-----|------|
| 汇率列表 | `/workflow/exchange-rates/` | GET |
| 货币转换 | `/workflow/exchange-rates/convert/` | GET |
| 历史查询 | `/workflow/exchange-rates/history/` | GET |
| 更新汇率 | `/workflow/exchange-rates/update/` | GET |

## 支持的货币

| 代码 | 名称 | 符号 |
|------|------|------|
| USD | 美元 | $ |
| CNY | 人民币 | ¥ |
| HKD | 港币 | HK$ |

## 默认汇率

- 1 USD = 7.20 CNY
- 1 USD = 7.80 HKD
- 1 CNY = 1.083333 HKD
- 反向汇率根据上述计算

## 常见问题

**Q: 如何添加新的货币？**
A: 修改 `workflow/models.py` 中的 `ExchangeRate.CURRENCY_CHOICES`，添加新货币代码和名称。

**Q: 如何集成真实汇率API？**
A: 在 `workflow/services_exchange_rate.py` 中实现 `fetch_from_api()` 方法，调用第三方汇率API。

**Q: 如何配置定时更新汇率？**
A: 创建 Django 管理命令，配置 Cron 定时任务调用 `update_exchange_rates_auto()`。

**Q: 如何在工作流中使用汇率？**
A: 在工作流步骤配置中使用 `update_exchange_rates` 处理器，或在自定义处理器中调用 `exchange_rate_service`。

## 文档

完整文档请参阅：
- `workflow/FOREX_MANAGEMENT.md` - 详细功能说明
- `FOREX_FEATURE_SUMMARY.md` - 实现报告

---

**外汇管理功能已完成并可用！** 🎉
