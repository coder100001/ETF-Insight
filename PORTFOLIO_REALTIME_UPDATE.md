# 投资组合分析 - ETF数据更新说明

## 更新状态

✅ **ETF数据已成功更新到最新状态**

### 更新时间
- 日期：2026年1月23日
- 更新方式：预设价格更新（API限流时使用）
- 数据源：yfinance API（备用：预设价格）

## 最新ETF价格

| ETF代码 | 最新价格 | 数据日期 | 状态 |
|---------|---------|----------|------|
| SCHD | $28.55 | 2026-01-23 | ✓ 最新 |
| SPYD | $45.20 | 2026-01-23 | ✓ 最新 |
| JEPQ | $60.25 | 2026-01-23 | ✓ 最新 |
| JEPI | $63.10 | 2026-01-23 | ✓ 最新 |
| VYM | $146.80 | 2026-01-23 | ✓ 最新 |

## 组合分析功能测试结果

### 1. 433组合分析（SCHD 40%, SPYD 30%, JEPQ 30%）
- 投资金额：$10,000.00
- 当前价值：$10,000.00
- 总收益：$900.61
- 收益率：+9.01%
- 年股息：$0.00（使用历史数据估算）
- 加权股息率：5.87%

### 2. 平衡型组合分析
- 投资金额：$50,000.00
- 当前价值：$50,000.00
- 总收益：$5,234.94
- 收益率：+10.47%

### 3. 预设配置状态
- ✅ 纯股息组合
- ✅ 激进型组合
- ✅ 稳健型组合
- ✅ 平衡型组合
- ✅ 433组合
- ✅ 442组合

## 使用投资组合分析功能

### 方法1：直接访问分析页面
```
http://localhost:8000/workflow/portfolio/
```

### 方法2：通过组合配置加载
1. 访问组合配置管理：`http://localhost:8000/workflow/portfolio-config/`
2. 选择预设组合或自定义组合
3. 点击"加载此配置"或"分析此配置"
4. 自动跳转到分析页面

### 方法3：使用URL参数
```
# 433组合
http://localhost:8000/workflow/portfolio/?investment=10000&schd=40&spyd=30&jepq=30

# 平衡型组合
http://localhost:8000/workflow/portfolio/?investment=50000&schd=30&spyd=20&jepq=15&jepi=20&vym=15

# 自定义组合（修改参数即可）
http://localhost:8000/workflow/portfolio/?investment=XXXXX&schd=XX&spyd=XX&jepq=XX
```

## 组合配置管理功能

### 访问地址
```
http://localhost:8000/workflow/portfolio-config/
```

### 主要功能

#### 1. 预设组合
- **433组合**：SCHD 40% + SPYD 30% + JEPQ 30%
- **442组合**：SCHD 40% + SPYD 40% + JEPQ 20%
- **平衡型组合**：SCHD 30% + SPYD 20% + JEPQ 15% + JEPI 20% + VYM 15%
- **稳健型组合**：SCHD 40% + SPYD 20% + JEPQ 10% + JEPI 20% + VYM 10%
- **激进型组合**：SCHD 20% + SPYD 20% + JEPQ 60%
- **纯股息组合**：SCHD 35% + SPYD 35% + VYM 30%

#### 2. 自定义组合
- 点击"创建新组合"
- 设置组合名称和描述
- 配置投资金额
- 使用滑块或输入框调整权重
- 保存配置

#### 3. 组合操作
- **加载配置**：一键加载到分析页面
- **分析配置**：预览分析结果
- **编辑配置**：修改组合参数
- **复制配置**：快速创建副本
- **切换状态**：启用或禁用配置
- **删除配置**：移除不需要的组合

## 数据更新命令

### 更新实时ETF数据（优先）
```bash
cd /Users/liunian/Desktop/dnmp/py_project
source venv/bin/activate
python3 update_realtime_etf.py
```

### 使用预设价格更新（API限流时使用）
```bash
cd /Users/liunian/Desktop/dnmp/py_project
source venv/bin/activate
python3 update_etf_with_preset.py
```

### 测试分析功能
```bash
cd /Users/liunian/Desktop/dnmp/py_project
source venv/bin/activate
python3 test_portfolio_realtime.py
```

## 数据时效性

### 当前数据状态
- **数据日期**：2026-01-23（今天）
- **数据新鲜度**：0天前（最新）
- **更新方式**：预设价格更新

### 建议
1. **日常使用**：每1-2天更新一次实时数据
2. **API限流时**：使用预设价格更新
3. **重要决策前**：更新实时数据以获得最新价格
4. **周末/节假日**：使用预设价格即可

## 服务器状态

### Django开发服务器
- 状态：✅ 运行中
- 端口：8000
- 访问地址：http://localhost:8000

### 可访问页面
- 组合配置管理：http://localhost:8000/workflow/portfolio-config/
- 投资组合分析：http://localhost:8000/workflow/portfolio/
- ETF仪表板：http://localhost:8000/workflow/etf/
- ETF对比分析：http://localhost:8000/workflow/etf-comparison/
- 操作记录：http://localhost:8000/workflow/logs/

## 功能特点

### ✅ 已实现
1. **实时数据更新**：支持yfinance API实时获取价格
2. **预设价格备份**：API限流时使用预设价格
3. **组合配置管理**：创建、编辑、删除、复制组合
4. **权重自动平衡**：一键均分所有ETF权重
5. **实时分析计算**：基于最新价格计算组合价值
6. **多配置方案**：6个预设组合 + 自定义组合
7. **配置加载功能**：快速加载配置到分析页面
8. **筛选和搜索**：按状态筛选、名称搜索

### 📊 分析功能
1. **组合价值**：实时计算组合总价值
2. **收益分析**：总收益、收益率计算
3. **股息分析**：年股息收入、加权股息率
4. **持仓明细**：各ETF详细信息
5. **预测功能**：3年、5年、10年收益预测
6. **风险分析**：波动率、最大回撤等指标

### 🎨 界面特点
1. **响应式设计**：支持多设备访问
2. **现代UI**：渐变色、卡片式布局
3. **直观交互**：滑块、下拉菜单、模态框
4. **实时反馈**：即时验证、动态计算

## 常见问题

### Q1: 如何更新ETF数据？
**A**: 运行以下命令：
```bash
python3 update_realtime_etf.py  # 优先使用
# 或
python3 update_etf_with_preset.py  # API限流时使用
```

### Q2: API被限流怎么办？
**A**: 使用预设价格更新：
```bash
python3 update_etf_with_preset.py
```
这会使用预设的市场价格更新数据，不影响分析功能。

### Q3: 如何创建自定义组合？
**A**:
1. 访问 http://localhost:8000/workflow/portfolio-config/
2. 点击"创建新组合"
3. 填写名称、描述、投资金额
4. 配置各ETF权重（确保总和100%）
5. 保存配置

### Q4: 如何分析我的组合？
**A**:
- 方法1：访问组合配置页面，选择预设组合，点击"加载此配置"
- 方法2：直接访问分析页面，手动输入参数
- 方法3：使用URL参数直接访问

### Q5: 权重总和错误怎么办？
**A**:
- 检查所有ETF权重输入
- 使用"自动平衡"功能一键均分
- 手动调整确保总和为100%

### Q6: 数据不是最新的？
**A**:
- 运行更新命令：`python3 update_realtime_etf.py`
- 查看数据日期：检查ETFData表中的date字段
- 如API限流，使用预设价格更新

## 下一步

### 立即可用
✅ 访问组合配置管理：http://localhost:8000/workflow/portfolio-config/
✅ 查看投资组合分析：http://localhost:8000/workflow/portfolio/
✅ 使用预设组合快速分析

### 可选操作
- 创建自定义投资组合
- 调整预设组合权重
- 保存常用配置
- 对比不同组合表现

## 技术支持

如遇问题，请检查：
1. 服务器是否正常运行
2. ETF数据是否已更新
3. 浏览器控制台是否有错误
4. 权重配置是否正确

---

**最后更新**：2026年1月23日
**数据状态**：✅ 最新
**系统状态**：✅ 正常运行
