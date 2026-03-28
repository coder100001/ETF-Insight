# 外汇管理功能说明

## 概述
外汇管理模块提供多货币汇率查询、转换、历史数据分析等功能，支持美元(USD)、人民币(CNY)、港币(HKD)之间的实时汇率管理。

## 功能特性

### 1. 货币转换器
- 支持 USD、CNY、HKD 三种货币互转
- 实时汇率计算
- 支持自定义金额
- 自动获取最新汇率

### 2. 汇率列表
- 显示当日所有货币对汇率
- 对比昨日汇率，显示涨跌
- 计算涨跌幅百分比
- 显示数据来源（API/系统/手动）
- 显示更新时间

### 3. 汇率走势图
- Chart.js 绘制最近7天汇率走势
- 支持多条曲线对比
- 交互式图表（鼠标悬停显示详细数据）

### 4. 历史记录查询
- 按日期分组显示历史数据
- 支持按货币对筛选
- 可自定义查询天数（默认30天）
- 表格形式展示，易于阅读

### 5. 汇率更新
- 手动触发更新
- 自动更新（系统默认值）
- API集成（预留接口）
- 批量更新所有货币对

## 使用指南

### 访问页面
```
http://localhost:8000/workflow/exchange-rates/
```

### 货币转换

**方式1：通过Web界面**
1. 在"货币转换器"区域选择源货币
2. 输入要转换的金额
3. 选择目标货币
4. 点击"转换"按钮
5. 查看转换结果和当前汇率

**方式2：通过API**
```bash
curl "http://localhost:8000/workflow/exchange-rates/convert/?from=USD&to=CNY&amount=1000"
```

响应示例：
```json
{
  "success": true,
  "from_currency": "USD",
  "to_currency": "CNY",
  "amount": 1000.0,
  "rate": 7.2,
  "result": 7200.0
}
```

### 查看汇率列表

直接访问汇率管理页面即可查看当日所有货币对汇率，包括：
- USD/CNY（美元对人民币）
- USD/HKD（美元对港币）
- CNY/HKD（人民币对港币）
- 以及反向货币对

### 汇率更新

**通过Web界面**
点击页面上的"立即更新"按钮即可手动触发汇率更新。

**通过API**
```bash
curl "http://localhost:8000/workflow/exchange-rates/update/"
```

响应示例：
```json
{
  "success": true,
  "message": "成功更新 6 条汇率记录",
  "result": {
    "success": true,
    "message": "成功更新 6 条汇率记录",
    "updated_count": 6,
    "errors": []
  }
}
```

### 历史查询

**通过Web界面**
1. 滚动到"汇率历史记录"区域
2. 点击货币对按钮（USD/CNY、USD/HKD、CNY/HKD）
3. 在下方表格中查看详细历史数据

**通过API**
```bash
curl "http://localhost:8000/workflow/exchange-rates/history/?from=USD&to=CNY&days=30"
```

响应示例：
```json
{
  "success": true,
  "from_currency": "USD",
  "to_currency": "CNY",
  "data": [
    {
      "date": "2025-12-27",
      "rate": 7.2,
      "source": "system"
    },
    {
      "date": "2025-12-26",
      "rate": 7.19,
      "source": "system"
    }
  ]
}
```

## 技术架构

### 数据模型
```python
class ExchangeRate(models.Model):
    from_currency = models.CharField(max_length=10)  # 源货币
    to_currency = models.CharField(max_length=10)    # 目标货币
    rate = models.DecimalField(max_digits=15, decimal_places=6)  # 汇率
    rate_date = models.DateField()  # 汇率日期
    data_source = models.CharField(max_length=50)  # 数据来源
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
```

### 服务层
```python
class ExchangeRateService:
    def get_latest_rate(from_currency, to_currency)
    def update_rates(rates, source)
    def get_history(from_currency, to_currency, days)
    def calculate_cross_rate(from_currency, to_currency)
    def convert(amount, from_currency, to_currency)
```

### 视图层
```python
ExchangeRateListView        # 汇率列表页面
ExchangeRateUpdateView      # 更新汇率API
ExchangeRateHistoryView     # 历史查询API
ExchangeRateConvertView    # 货币转换API
```

### 路由配置
```python
path('exchange-rates/', ExchangeRateListView.as_view())
path('exchange-rates/update/', ExchangeRateUpdateView.as_view())
path('exchange-rates/history/', ExchangeRateHistoryView.as_view())
path('exchange-rates/convert/', ExchangeRateConvertView.as_view())
```

## 支持的货币

| 代码 | 名称 | 符号 |
|------|------|------|
| USD | 美元 | $ |
| CNY | 人民币 | ¥ |
| HKD | 港币 | HK$ |

## 汇率数据源

### 系统默认汇率
当API不可用时，使用系统预设的汇率：
- 1 USD = 7.20 CNY
- 1 USD = 7.80 HKD
- 1 CNY = 1.083333 HKD

### API集成（预留）
代码中已预留真实API接口，可集成：
- Fixer.io (https://api.fixer.io)
- ExchangeRate-API (https://v6.exchangerate-api.com)
- 其他汇率数据服务

## 集成到工作流

汇率更新已集成到工作流系统中，可在工作流步骤中使用：

```python
# 在工作流处理器中调用
from workflow.handlers import update_exchange_rates

def my_handler(params):
    # 更新汇率
    result = update_exchange_rates({})
    print(f"更新了 {result['updated_count']} 条汇率记录")
    
    # 使用汇率进行计算
    from workflow.services_exchange_rate import exchange_rate_service
    usd_amount = exchange_rate_service.convert(10000, 'CNY', 'USD')
    print(f"10000 CNY = {usd_amount} USD")
    
    return {'status': 'success'}
```

## 定时任务

建议配置定时任务自动更新汇率：

### 使用Django管理命令
创建 `workflow/management/commands/update_exchange_rates.py`:
```python
from django.core.management.base import BaseCommand
from workflow.services_exchange_rate import update_exchange_rates_auto

class Command(BaseCommand):
    help = '更新汇率数据'

    def handle(self, *args, **options):
        result = update_exchange_rates_auto()
        self.stdout.write(self.style.SUCCESS(result['message']))
```

### 配置Cron（Linux）
```bash
# 每天早上9点更新汇率
0 9 * * * python /path/to/manage.py update_exchange_rates
```

## 错误处理

### 常见错误

**1. 汇率不存在**
- 原因：查询的货币对在数据库中不存在
- 处理：返回默认汇率 1.0
- 建议：先执行汇率更新

**2. API请求失败**
- 原因：网络问题或API密钥无效
- 处理：降级到系统默认汇率
- 建议：检查网络连接和API密钥

**3. 交叉汇率计算错误**
- 原因：缺少中间货币汇率
- 处理：尝试直接汇率或返回1.0
- 建议：确保所有基础货币对都有数据

## 性能优化

### 数据库优化
- 已添加索引：`(from_currency, to_currency)`、`(-rate_date)`
- 查询优化：使用 `select_related` 和 `prefetch_related`

### 缓存策略
- 建议使用Redis缓存最新汇率（TTL: 1小时）
- 历史数据可缓存更长时间（TTL: 24小时）

### 批量操作
- 批量更新：一次性更新所有货币对
- 减少数据库查询次数

## 扩展建议

### 短期（1-2周）
1. 添加更多货币支持（EUR、GBP、JPY等）
2. 集成真实汇率API
3. 添加汇率预警功能
4. 支持自定义汇率设置

### 中期（1-2月）
1. 实现汇率趋势分析
2. 添加汇率预测功能
3. 支持汇率对比分析
4. 添加导出功能（CSV/Excel）

### 长期（3-6月）
1. 集成WebSocket实时推送
2. 实现汇率交易策略回测
3. 添加汇率风险管理功能
4. 支持多语言界面

## 文件清单

### 新增文件
1. `workflow/services_exchange_rate.py` - 汇率服务实现
2. `workflow/FOREX_MANAGEMENT.md` - 本文档

### 修改文件
1. `workflow/views_exchange_rate.py` - 增强汇率视图
2. `workflow/templates/workflow/exchange_rate_list.html` - 汇率管理页面
3. `workflow/templates/workflow/base.html` - 添加导航链接
4. `workflow/urls.py` - 添加路由配置

## 使用示例

### Python代码示例
```python
from workflow.services_exchange_rate import exchange_rate_service

# 获取最新汇率
rate = exchange_rate_service.get_latest_rate('USD', 'CNY')
print(f"1 USD = {rate} CNY")

# 货币转换
cny = exchange_rate_service.convert(100, 'USD', 'CNY')
print(f"100 USD = {cny} CNY")

# 获取历史数据
history = exchange_rate_service.get_history('USD', 'CNY', days=30)
for item in history:
    print(f"{item['date']}: {item['rate']}")

# 计算交叉汇率
cross_rate = exchange_rate_service.calculate_cross_rate('CNY', 'HKD')
print(f"1 CNY = {cross_rate} HKD")
```

### JavaScript示例
```javascript
// 货币转换
async function convertCurrency(from, to, amount) {
    const response = await fetch(
        `/workflow/exchange-rates/convert/?from=${from}&to=${to}&amount=${amount}`
    );
    const data = await response.json();
    return data.result;
}

// 使用
const result = await convertCurrency('USD', 'CNY', 1000);
console.log(`1000 USD = ${result} CNY`);
```

## 总结

外汇管理模块提供了完整的汇率管理功能，包括：

✅ **货币转换器**：支持三种货币互转
✅ **汇率列表**：实时显示所有货币对汇率
✅ **汇率走势**：图表展示历史趋势
✅ **历史查询**：支持按日期和货币对查询
✅ **API接口**：提供RESTful API
✅ **工作流集成**：可在工作流中使用
✅ **服务层封装**：易于扩展和维护

通过这些功能，用户可以方便地进行货币转换、查看汇率走势、分析历史数据，并在工作流中集成汇率相关功能。
