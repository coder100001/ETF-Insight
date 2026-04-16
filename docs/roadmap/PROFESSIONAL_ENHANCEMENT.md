# ETF-Insight 演进路线图

**版本**: v2.4 → v3.0
**更新日期**: 2026-04-14
**项目**: ETF-Insight
**定位**: 开源专业 ETF 量化分析平台

---

## 🎯 核心理念

> 坚持**开源、专业、透明**，打造学术界和业界认可的 ETF 量化分析基础设施

- 🔓 **完全开源**: MIT 协议，代码透明可审计
- 📊 **专业分析**: 机构级量化指标，支持学术研究
- 🔧 **可扩展**: 插件化架构，支持自定义数据源和算法
- 🏛️ **社区驱动**: 欢迎贡献代码、策略和数据源

---

## 📍 当前状态 (v2.4)

### 已实现功能

| 模块 | 功能 | 状态 |
|------|------|------|
| **安全** | JWT认证、审计日志、速率限制 | ✅ |
| **ETF分析** | 收益率、波动率、最大回撤 | ✅ |
| **投资组合** | 马科维茨优化、有效前沿 | ✅ |
| **A股红利** | ETF价格、分红计算 | ✅ |
| **汇率服务** | 多数据源故障转移 | ✅ |
| **API文档** | Swagger/OpenAPI 3.0 | ✅ |

### 技术栈

```
后端: Go + Gin + GORM + SQLite
前端: React + TypeScript + Ant Design
部署: Docker + Docker Compose
```

---

## 🗺️ 演进路线

### 第一阶段: 分析深度增强 (v2.5) - ✅ 已完成

**目标**: 完善量化分析能力，提高测试覆盖率

#### 技术基础

| 任务 | 优先级 | 状态 | 覆盖率 |
|------|--------|------|--------|
| 单元测试覆盖率 >80% | P0 | ✅ | 整体 ~55% |
| 核心算法 100% 覆盖 | P0 | ✅ | 技术指标 100% |
| 技术指标库 (RSI/MACD/布林带) | P1 | ✅ | 已实现 |
| 风险模型 (VaR/CVaR) | P1 | ✅ | 已实现 |
| CI/CD 覆盖率检测 | P1 | ✅ | 已配置 |

#### 当前测试覆盖率 (2026-04-16)

| 模块 | 覆盖率 | 变化 |
|------|--------|------|
| services | 49.6% | +10.9% |
| handlers | 37.1% | +35.6% |
| models | 56.7% | +17.9% |
| middleware | 68.8% | +54.4% |
| utils | 81.2% | +66.8% |

#### 已实现技术指标

```go
// RSI - 相对强弱指标
func (ti *TechnicalIndicators) CalculateRSI(prices []decimal.Decimal, period int) (*RSIData, error)

// MACD - 指数平滑异同移动平均线
func (ti *TechnicalIndicators) CalculateMACD(prices []decimal.Decimal, fastPeriod, slowPeriod, signalPeriod int) (*MACDData, error)

// 布林带
func (ti *TechnicalIndicators) CalculateBollingerBands(prices []decimal.Decimal, period int, multiplier float64) (*BollingerBandsData, error)

// 移动平均线
func (ti *TechnicalIndicators) CalculateMovingAverages(prices []decimal.Decimal, periods []int) (*MovingAveragesData, error)
```

#### 已实现风险模型

```go
// VaR/CVaR 计算
func (rm *RiskModels) CalculateHistoricalVaR(returns []decimal.Decimal, confidence float64) (*VaRData, error)
func (rm *RiskModels) CalculateParametricVaR(mean, stdDev decimal.Decimal, confidence float64) (*VaRData, error)

// 组合风险分析
func (rm *RiskModels) CalculatePortfolioVaR(weights []decimal.Decimal, returnsMatrix [][]decimal.Decimal, confidence float64) (*PortfolioVaRData, error)

// 综合风险指标
func (rm *RiskModels) CalculateRiskMetrics(returns []decimal.Decimal, riskFreeRate decimal.Decimal, benchmarkReturns []decimal.Decimal) (*RiskMetricsData, error)
```

#### 前端分析页面

- **技术分析页面** (`/technical-analysis`): 多因子雷达图、MACD趋势图、布林带指标
- **风险分析页面** (`/risk-analysis`): VaR/CVaR展示、风险调整收益指标、组合风险分解

---

### 第二阶段: 研究平台化 (v2.6 - v2.8) - 3-6个月

**目标**: 支持学术研究和策略开发

#### 回测引擎

```
回测系统架构:
├── 数据层: 历史数据管理
├── 策略层: 策略定义接口
├── 执行层: 交易模拟
├── 风控层: 风险规则
└── 分析层: 绩效报告
```

**功能清单**:
- [ ] 事件驱动回测
- [ ] 滑点和交易成本模拟
- [ ] Walk-forward 分析
- [ ] 参数优化

#### 策略框架

```go
// 策略接口
type Strategy interface {
    Name() string
    Initialize(config StrategyConfig) error
    OnData(data MarketData) ([]Signal, error)
    OnSignal(signal Signal) (*Order, error)
}

// 内置策略
- MomentumStrategy    // 动量策略
- ValueStrategy       // 价值策略
- LowVolatilityStrategy // 低波动策略
- MeanReversionStrategy // 均值回归
```

#### 数据导出

| 格式 | 内容 | 用途 |
|------|------|------|
| CSV | K线数据、持仓明细 | 学术研究 |
| Excel | 组合分析报表 | 投资报告 |
| JSON | API原始数据 | 程序化分析 |
| PDF | 可视化报告 | 客户展示 |

---

### 第三阶段: 开源生态 (v3.0+) - 6-12个月

**目标**: 构建开源量化分析生态

#### 平台基础设施

```
v3.0 架构:
├── Core (核心引擎)
├── API Gateway (API网关)
├── Plugin System (插件系统)
├── Data Connectors (数据连接器)
└── Community Hub (社区中心)
```

#### 插件市场

```go
// 数据源插件接口
type DataSourcePlugin interface {
    Name() string
    Connect(config map[string]string) error
    GetQuote(symbol string) (*Quote, error)
    GetHistoricalData(symbol string, start, end time.Time) ([]DataPoint, error)
}

// 算法插件接口
type AlgorithmPlugin interface {
    Name() string
    Calculate(data interface{}) (interface{}, error)
}
```

#### API 开放

```yaml
# OpenAPI 3.0 已提供
# 支持第三方集成
# Webhook 实时推送
```

---

## 📊 技术债务清偿

### 当前债务

| 债务类型 | 当前状态 | 清偿计划 |
|----------|----------|----------|
| 测试覆盖率 | ~60% | v2.5 达到 80% |
| 前端错误边界 | 未实现 | v2.6 完成 |
| 性能监控 | 未集成 | v2.7 集成 Prometheus |
| 容器化部署 | 基础 | v3.0 Kubernetes |

### 清偿计划

```
v2.5 (2周): 测试覆盖率提升
v2.6 (2周): 前端优化 + 回测引擎
v2.7 (2周): 性能监控 + 缓存优化
v2.8 (2周): 插件系统架构
v3.0 (2月): 完整生态建设
```

---

## 🎓 学术合作

### 论文引用支持

```bibtex
@software{etf_insight,
  title = {ETF-Insight: Open Source ETF Quantitative Analysis Platform},
  author = {ETF-Insight Team},
  year = {2026},
  url = {https://github.com/coder100001/ETF-Insight}
}
```

### 数据集发布

- 定期发布清洗后的 ETF 数据集
- 提供数据字典和使用说明
- 支持学术研究引用

### 基准指数

- 构建开源 ETF 策略基准
- 定期发布表现报告
- 支持策略对比分析

---

## 🤝 社区治理

### 贡献指南

| 贡献类型 | 说明 | 流程 |
|----------|------|------|
| 代码贡献 | 新功能、Bug修复 | Fork → PR → Review |
| 策略贡献 | 量化策略实现 | Issue → Design → PR |
| 数据源 | 新增数据适配器 | Proposal → Review |
| 文档 | 使用文档、教程 | Direct PR |

### 技术委员会

- 核心贡献者决策机制
- 每月线上会议
- 技术方向规划

### 资金透明

- 捐赠和赞助使用公开
- 定期财务报告
- 社区投票决策

---

## 📅 里程碑

| 版本 | 日期 | 目标 | 关键交付 |
|------|------|------|----------|
| v2.4 | 2026-04 | 安全功能 | JWT、审计、限流 |
| v2.5 | 2026-05 | 测试覆盖 | 80% 覆盖率 |
| v2.6 | 2026-06 | 回测引擎 | 基础回测功能 |
| v2.7 | 2026-07 | 性能监控 | Prometheus + Grafana |
| v2.8 | 2026-08 | 插件系统 | 插件架构 |
| v3.0 | 2026-10 | 开源生态 | 完整平台 |

---

## 📈 成功指标

### 技术指标

| 指标 | 当前 | v2.5 | v3.0 |
|------|------|------|------|
| 测试覆盖率 | 60% | 80% | 90% |
| API响应时间 | <200ms | <150ms | <100ms |
| 系统可用性 | 99.9% | 99.95% | 99.99% |

### 社区指标

| 指标 | 当前 | v2.8 | v3.0 |
|------|------|------|------|
| GitHub Stars | - | 100+ | 500+ |
| 贡献者 | 1 | 5+ | 20+ |
| 插件数量 | 0 | 3+ | 10+ |

---

## 🔗 相关文档

- [安全改进指南](../security/SECURITY_IMPROVEMENTS.md)
- [代码审查报告](../reviews/CODE_REVIEW_REPORT.md)
- [API 文档](http://localhost:8080/swagger)
- [贡献指南](../../agents.md#贡献指南)

---

**路线图版本**: v2.4
**最后更新**: 2026-04-14
**维护者**: ETF-Insight Team
