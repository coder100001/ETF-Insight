# ETF-Insight 数据层演进改造方案

**版本**: v2.7 → v2.8
**创建日期**: 2026-04-25
**状态**: 规划中

---

## 📊 执行摘要

本文档基于《ETF-Insight 演进路线图 2026》中的模型深度融合方案，制定数据层的全面改造计划。改造遵循**保留现有数据完整性、仅进行功能完善和结构优化**的原则，严禁执行任何数据删除操作。

### 核心目标

1. **支持因子择时与Alpha观点生成**：为闭环一提供数据支撑
2. **支持风险预算与CVaR管理**：为闭环二提供数据支撑
3. **支持插件架构**：为微内核架构提供配置管理
4. **提升性能与可扩展性**：优化数据访问效率

---

## 🔍 现有数据层架构分析

### 架构优势

| 模块 | 优势描述 | 评估 |
|------|----------|------|
| **统一资产模型** | Asset模型支持多种资产类型（股票、ETF、指数、债券） | ⭐⭐⭐⭐⭐ |
| **价格数据模型** | Price模型支持多时间粒度（日/周/月/分钟），包含调整后价格 | ⭐⭐⭐⭐⭐ |
| **组合管理模型** | Portfolio模型设计完善，支持持仓、表现、再平衡全流程 | ⭐⭐⭐⭐⭐ |
| **持仓穿透模型** | Holding/ETFHolding支持ETF成分股分析和穿透 | ⭐⭐⭐⭐ |
| **因子分析服务** | 已实现Fama-French三因子/五因子分析服务 | ⭐⭐⭐⭐ |

### 架构不足

| 模块 | 问题描述 | 影响程度 | 改造优先级 |
|------|----------|----------|------------|
| **因子数据存储** | 缺少因子数据持久化，依赖模拟数据或实时计算 | 🔴 高 | P0 |
| **因子择时信号** | 缺少因子择时信号（Z-score、斜率）存储 | 🔴 高 | P0 |
| **Alpha观点存储** | 缺少Black-Litterman观点存储和管理 | 🔴 高 | P0 |
| **风险预算配置** | 缺少CVaR预算配置和约束存储 | 🔴 高 | P0 |
| **蒙特卡洛缓存** | 缺少模拟结果缓存，重复计算成本高 | 🟡 中 | P1 |
| **插件配置管理** | 缺少插件注册、配置、执行日志存储 | 🟡 中 | P1 |
| **回测结果存储** | 缺少回测结果持久化，无法复现历史分析 | 🟡 中 | P2 |
| **模型基准对比** | 缺少模型基准对比矩阵存储 | 🟡 中 | P2 |

---

## 🎯 演进方案需求映射

### 闭环一：Alpha模型 + Black-Litterman

#### 数据需求分析

| 功能点 | 数据需求 | 数据模型 | 优先级 |
|--------|----------|----------|--------|
| **因子数据管理** | 市场因子、SMB、HML、RMW、CMA的历史数据 | `FactorData` | P0 |
| **因子择时信号** | 60日移动平均斜率、Z-score | `FactorTimingSignal` | P0 |
| **Alpha观点生成** | 资产、观点收益、信心水平、生成时间 | `AlphaView` | P0 |
| **观点历史记录** | 观点的历史表现、胜率统计 | `AlphaViewPerformance` | P0 |
| **BL模型配置** | 风险厌恶系数δ、先验基准、观点误差矩阵Ω | `BlackLittermanConfig` | P0 |
| **后验收益存储** | BL模型计算的后验预期收益向量 | `BLPosteriorReturn` | P1 |

#### 关键数据字段设计

**因子择时信号（FactorTimingSignal）**：
- 因子名称（Mkt-RF、SMB、HML等）
- 计算日期
- 60日移动平均斜率
- 当前Z-score
- 历史百分位
- 信号强度（强正/弱正/中性/弱负/强负）

**Alpha观点（AlphaView）**：
- 观点ID
- 资产代码
- 观点收益（%）
- 信心水平（%）
- 观点类型（绝对/相对）
- 生成方法（因子择时/动量/均值回复）
- 生成时间
- 有效期
- 状态（活跃/过期/已验证）

---

### 闭环二：风险预算 + 风险平价

#### 数据需求分析

| 功能点 | 数据需求 | 数据模型 | 优先级 |
|--------|----------|----------|--------|
| **风险预算配置** | 各资产类别的CVaR预算上限 | `RiskBudgetConfig` | P0 |
| **蒙特卡洛模拟** | 模拟参数、结果分布、VaR/CVaR值 | `MonteCarloSimulation` | P0 |
| **风险贡献分解** | 各资产的CVaR贡献、边际贡献 | `RiskContribution` | P0 |
| **风险预算执行** | 实际风险预算使用情况、偏离度 | `RiskBudgetExecution` | P1 |
| **CVaR历史记录** | 历史CVaR值、极端事件记录 | `CVaRHistory` | P1 |

#### 关键数据字段设计

**风险预算配置（RiskBudgetConfig）**：
- 配置ID
- 组合ID
- 资产类别（股票/债券/商品/现金）
- CVaR预算上限（%）
- VaR预算上限（%）
- 偏度约束（最小值）
- 最大回撤限制
- 配置生效日期
- 配置状态

**蒙特卡洛模拟（MonteCarloSimulation）**：
- 模拟ID
- 组合ID
- 模拟参数（路径数、时间步长、置信水平）
- 模拟日期
- VaR值（95%/99%）
- CVaR值（95%/99%）
- 分布参数（均值、标准差、偏度、峰度）
- 模拟结果（JSON存储）
- 缓存有效期

---

### 微内核+插件架构

#### 数据需求分析

| 功能点 | 数据需求 | 数据模型 | 优先级 |
|--------|----------|----------|--------|
| **插件注册** | 插件名称、版本、类型、接口定义 | `PluginRegistry` | P1 |
| **插件配置** | 插件参数配置、依赖关系 | `PluginConfiguration` | P1 |
| **插件执行日志** | 执行时间、输入输出、性能指标 | `PluginExecutionLog` | P1 |
| **模型基准对比** | 不同插件组合的性能对比矩阵 | `ModelBenchmarkMatrix` | P2 |
| **策略实验记录** | 策略实验的配置、结果、评估 | `StrategyExperiment` | P2 |

#### 关键数据字段设计

**插件注册（PluginRegistry）**：
- 插件ID
- 插件名称
- 插件类型（AlphaGenerator/PortfolioOptimizer/RiskModel）
- 版本号
- 接口定义（JSON Schema）
- 依赖插件列表
- 状态（启用/禁用）
- 注册时间

**模型基准对比（ModelBenchmarkMatrix）**：
- 对比ID
- Alpha插件ID
- Optimizer插件ID
- Risk插件ID
- 回测窗口（3年滚动）
- 再平衡频率
- 交易成本模型
- 性能指标（JSON存储）
  - 最大回撤
  - 夏普比率
  - Calmar比率
  - 滚动1年胜率
  - 尾部依赖指数
- 对比日期

---

## 📐 数据模型设计

### 新增数据模型

#### 1. 因子数据模型

```go
// FactorData 因子数据
type FactorData struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    FactorName string   `json:"factor_name" gorm:"size:20;index:idx_factor_date"` // Mkt-RF, SMB, HML, RMW, CMA
    Date      time.Time `json:"date" gorm:"index:idx_factor_date"`
    Value     decimal.Decimal `json:"value" gorm:"type:decimal(10,6)"`
    DataSource string    `json:"data_source" gorm:"size:50"` // Fama-French Library, ETF Proxy
    CreatedAt time.Time `json:"created_at"`
}

// FactorTimingSignal 因子择时信号
type FactorTimingSignal struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    FactorName      string    `json:"factor_name" gorm:"size:20;index:idx_signal_date"`
    SignalDate      time.Time `json:"signal_date" gorm:"index:idx_signal_date"`

    // 择时指标
    MASlope60       decimal.Decimal `json:"ma_slope_60" gorm:"type:decimal(10,6)"`   // 60日移动平均斜率
    ZScore          decimal.Decimal `json:"z_score" gorm:"type:decimal(10,6)"`        // 当前Z-score
    Percentile      decimal.Decimal `json:"percentile" gorm:"type:decimal(5,2)"`      // 历史百分位

    // 信号强度
    SignalStrength  string    `json:"signal_strength" gorm:"size:20"` // strong_positive, weak_positive, neutral, weak_negative, strong_negative
    SignalScore     int       `json:"signal_score"`                    // -2, -1, 0, 1, 2

    // 预期收益
    ExpectedReturn  decimal.Decimal `json:"expected_return" gorm:"type:decimal(10,6)"`
    Confidence      decimal.Decimal `json:"confidence" gorm:"type:decimal(5,2)"` // 信心水平

    CreatedAt       time.Time `json:"created_at"`
}
```

#### 2. Alpha观点模型

```go
// AlphaView Alpha观点
type AlphaView struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PortfolioID uint      `json:"portfolio_id" gorm:"index"` // 关联组合

    // 观点内容
    AssetSymbol string    `json:"asset_symbol" gorm:"size:20;index:idx_view_asset"` // 资产代码
    ViewReturn  decimal.Decimal `json:"view_return" gorm:"type:decimal(10,6)"`      // 观点收益（%）
    Confidence  decimal.Decimal `json:"confidence" gorm:"type:decimal(5,2)"`        // 信心水平（%）

    // 观点类型
    ViewType    string    `json:"view_type" gorm:"size:20"`           // absolute, relative
    ViewMethod  string    `json:"view_method" gorm:"size:50"`         // factor_timing, momentum, mean_reversion

    // 生成信息
    GeneratedAt time.Time `json:"generated_at"`                      // 生成时间
    ValidUntil  time.Time `json:"valid_until"`                       // 有效期
    Status      string    `json:"status" gorm:"size:20;default:'active'"` // active, expired, validated

    // 因子来源
    SourceFactor string    `json:"source_factor" gorm:"size:20"`       // 来源因子（HML, SMB等）
    FactorLoading decimal.Decimal `json:"factor_loading" gorm:"type:decimal(10,6)"` // 因子载荷

    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// AlphaViewPerformance Alpha观点表现
type AlphaViewPerformance struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    ViewID          uint      `json:"view_id" gorm:"index"` // 关联观点

    // 实际表现
    ActualReturn    decimal.Decimal `json:"actual_return" gorm:"type:decimal(10,6)"` // 实际收益
    PredictionError decimal.Decimal `json:"prediction_error" gorm:"type:decimal(10,6)"` // 预测误差

    // 验证结果
    IsValidated     bool      `json:"is_validated"`           // 是否已验证
    ValidationDate  time.Time `json:"validation_date"`        // 验证日期
    IsCorrect       bool      `json:"is_correct"`             // 方向是否正确

    // 滚动统计
    RollingWinRate  decimal.Decimal `json:"rolling_win_rate" gorm:"type:decimal(5,2)"` // 滚动胜率（3个月）

    CreatedAt       time.Time `json:"created_at"`
}
```

#### 3. Black-Litterman配置模型

```go
// BlackLittermanConfig Black-Litterman模型配置
type BlackLittermanConfig struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PortfolioID uint      `json:"portfolio_id" gorm:"uniqueIndex"` // 关联组合

    // 模型参数
    RiskAversion    decimal.Decimal `json:"risk_aversion" gorm:"type:decimal(10,6)"` // 风险厌恶系数δ
    PriorType       string          `json:"prior_type" gorm:"size:20"`               // equal_weight, min_variance, market_cap

    // 先验基准
    PriorWeights    string          `json:"prior_weights" gorm:"type:json"`          // 先验权重（JSON）
    ImpliedReturns  string          `json:"implied_returns" gorm:"type:json"`        // 隐含均衡收益（JSON）

    // 观点误差矩阵
    OmegaMethod     string          `json:"omega_method" gorm:"size:20"`             // Idzorek, HeLitterman
    OmegaMatrix     string          `json:"omega_matrix" gorm:"type:json"`           // 观点误差矩阵Ω

    // 配置状态
    IsActive        bool            `json:"is_active" gorm:"default:true"`
    LastCalculated  time.Time       `json:"last_calculated"`

    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}

// BLPosteriorReturn BL后验收益
type BLPosteriorReturn struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    ConfigID        uint      `json:"config_id" gorm:"index"` // 关联配置
    CalculationDate time.Time `json:"calculation_date" gorm:"index"`

    // 后验结果
    PosteriorReturns string `json:"posterior_returns" gorm:"type:json"` // 后验预期收益向量
    PosteriorWeights string `json:"posterior_weights" gorm:"type:json"` // 后验权重向量
    PosteriorCov     string `json:"posterior_cov" gorm:"type:json"`     // 后验协方差矩阵

    // 观点融合信息
    NumViews        int       `json:"num_views"`                          // 观点数量
    ViewImpact      decimal.Decimal `json:"view_impact" gorm:"type:decimal(10,6)"` // 观点影响度

    CreatedAt       time.Time `json:"created_at"`
}
```

#### 4. 风险预算模型

```go
// RiskBudgetConfig 风险预算配置
type RiskBudgetConfig struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PortfolioID uint      `json:"portfolio_id" gorm:"uniqueIndex"` // 关联组合

    // CVaR预算
    StockCVaRBudget    decimal.Decimal `json:"stock_cvar_budget" gorm:"type:decimal(5,2)"`    // 股票CVaR预算上限（%）
    BondCVaRBudget     decimal.Decimal `json:"bond_cvar_budget" gorm:"type:decimal(5,2)"`     // 债券CVaR预算上限（%）
    CommodityCVaRBudget decimal.Decimal `json:"commodity_cvar_budget" gorm:"type:decimal(5,2)"` // 商品CVaR预算上限（%）
    CashCVaRBudget     decimal.Decimal `json:"cash_cvar_budget" gorm:"type:decimal(5,2)"`     // 现金CVaR预算上限（%）

    // VaR预算（可选）
    UseVaRConstraint   bool            `json:"use_var_constraint" gorm:"default:false"`
    StockVaRBudget     decimal.Decimal `json:"stock_var_budget" gorm:"type:decimal(5,2)"`
    BondVaRBudget      decimal.Decimal `json:"bond_var_budget" gorm:"type:decimal(5,2)"`

    // 其他约束
    MinSkewness        decimal.Decimal `json:"min_skewness" gorm:"type:decimal(10,6)"`   // 最小偏度约束
    MaxDrawdown        decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(5,2)"`    // 最大回撤限制（%）

    // 置信水平
    CVaRConfidence     decimal.Decimal `json:"cvar_confidence" gorm:"type:decimal(5,4);default:0.95"` // CVaR置信水平
    VaRConfidence      decimal.Decimal `json:"var_confidence" gorm:"type:decimal(5,4);default:0.95"`  // VaR置信水平

    // 配置状态
    IsActive           bool      `json:"is_active" gorm:"default:true"`
    EffectiveDate      time.Time `json:"effective_date"`

    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

// MonteCarloSimulation 蒙特卡洛模拟
type MonteCarloSimulation struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PortfolioID uint      `json:"portfolio_id" gorm:"index:idx_simulation_portfolio"` // 关联组合
    SimulationDate time.Time `json:"simulation_date" gorm:"index:idx_simulation_portfolio"`

    // 模拟参数
    NumPaths        int       `json:"num_paths"`           // 模拟路径数
    TimeSteps       int       `json:"time_steps"`          // 时间步长
    TimeHorizon     int       `json:"time_horizon"`        // 时间范围（天）
    ConfidenceLevel decimal.Decimal `json:"confidence_level" gorm:"type:decimal(5,4)"` // 置信水平

    // 风险指标
    VaR95          decimal.Decimal `json:"var_95" gorm:"type:decimal(10,6)"`   // VaR (95%)
    VaR99          decimal.Decimal `json:"var_99" gorm:"type:decimal(10,6)"`   // VaR (99%)
    CVaR95         decimal.Decimal `json:"cvar_95" gorm:"type:decimal(10,6)"`  // CVaR (95%)
    CVaR99         decimal.Decimal `json:"cvar_99" gorm:"type:decimal(10,6)"`  // CVaR (99%)

    // 分布参数
    MeanReturn     decimal.Decimal `json:"mean_return" gorm:"type:decimal(10,6)"`
    StdDev         decimal.Decimal `json:"std_dev" gorm:"type:decimal(10,6)"`
    Skewness       decimal.Decimal `json:"skewness" gorm:"type:decimal(10,6)"`
    Kurtosis       decimal.Decimal `json:"kurtosis" gorm:"type:decimal(10,6)"`

    // 模拟结果（缓存）
    SimulationResult string `json:"simulation_result" gorm:"type:json"` // 模拟结果（JSON）
    CacheExpiry      time.Time `json:"cache_expiry"`                     // 缓存过期时间

    CreatedAt       time.Time `json:"created_at"`
}

// RiskContribution 风险贡献
type RiskContribution struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    SimulationID    uint      `json:"simulation_id" gorm:"index"` // 关联模拟
    AssetID         uint      `json:"asset_id" gorm:"index"`      // 资产ID
    AssetSymbol     string    `json:"asset_symbol" gorm:"size:20"` // 资产代码

    // 风险贡献
    Weight          decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`        // 权重
    CVaRContribution decimal.Decimal `json:"cvar_contribution" gorm:"type:decimal(10,6)"` // CVaR贡献
    MarginalCVaR    decimal.Decimal `json:"marginal_cvar" gorm:"type:decimal(10,6)"`      // 边际CVaR
    CVaRPercentage  decimal.Decimal `json:"cvar_percentage" gorm:"type:decimal(5,2)"`     // CVaR贡献百分比

    // 预算对比
    BudgetLimit     decimal.Decimal `json:"budget_limit" gorm:"type:decimal(5,2)"`   // 预算上限
    BudgetUsage     decimal.Decimal `json:"budget_usage" gorm:"type:decimal(5,2)"`   // 预算使用率
    BudgetDeviation decimal.Decimal `json:"budget_deviation" gorm:"type:decimal(5,2)"` // 预算偏离度

    CalculationDate time.Time `json:"calculation_date"`
    CreatedAt       time.Time `json:"created_at"`
}

// RiskBudgetExecution 风险预算执行记录
type RiskBudgetExecution struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    PortfolioID     uint      `json:"portfolio_id" gorm:"index"` // 关联组合
    ExecutionDate   time.Time `json:"execution_date" gorm:"index"`

    // 执行情况
    TotalCVaR       decimal.Decimal `json:"total_cvar" gorm:"type:decimal(10,6)"`     // 组合总CVaR
    StockCVaRUsage  decimal.Decimal `json:"stock_cvar_usage" gorm:"type:decimal(5,2)"` // 股票CVaR使用率
    BondCVaRUsage   decimal.Decimal `json:"bond_cvar_usage" gorm:"type:decimal(5,2)"`  // 债券CVaR使用率

    // 约束满足情况
    IsBudgetSatisfied bool      `json:"is_budget_satisfied"` // 是否满足预算约束
    ViolationDetails  string    `json:"violation_details" gorm:"type:json"` // 违规详情

    // 优化信息
    OptimizationMethod string    `json:"optimization_method" gorm:"size:50"` // 优化方法
    Iterations         int       `json:"iterations"`                        // 迭代次数

    CreatedAt          time.Time `json:"created_at"`
}
```

#### 5. 插件架构模型

```go
// PluginRegistry 插件注册表
type PluginRegistry struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PluginName  string    `json:"plugin_name" gorm:"uniqueIndex;size:100"` // 插件名称
    PluginType  string    `json:"plugin_type" gorm:"size:50;index"`        // AlphaGenerator, PortfolioOptimizer, RiskModel
    Version     string    `json:"version" gorm:"size:20"`                  // 版本号

    // 接口定义
    InputSchema  string `json:"input_schema" gorm:"type:json"`  // 输入接口定义（JSON Schema）
    OutputSchema string `json:"output_schema" gorm:"type:json"` // 输出接口定义（JSON Schema）

    // 依赖关系
    Dependencies string `json:"dependencies" gorm:"type:json"` // 依赖插件列表

    // 元数据
    Description  string `json:"description" gorm:"size:500"` // 插件描述
    Author       string `json:"author" gorm:"size:100"`      // 作者
    Documentation string `json:"documentation" gorm:"size:200"` // 文档链接

    // 状态
    Status       string    `json:"status" gorm:"size:20;default:'active'"` // active, deprecated, disabled
    RegisteredAt time.Time `json:"registered_at"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// PluginConfiguration 插件配置
type PluginConfiguration struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PluginID    uint      `json:"plugin_id" gorm:"index"` // 关联插件
    PortfolioID uint      `json:"portfolio_id" gorm:"index"` // 关联组合（可选）

    // 配置参数
    ConfigName  string    `json:"config_name" gorm:"size:100"` // 配置名称
    Parameters  string    `json:"parameters" gorm:"type:json"` // 参数配置（JSON）

    // 状态
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    IsDefault   bool      `json:"is_default" gorm:"default:false"` // 是否为默认配置

    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// PluginExecutionLog 插件执行日志
type PluginExecutionLog struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    PluginID    uint      `json:"plugin_id" gorm:"index"` // 关联插件
    ConfigID    uint      `json:"config_id" gorm:"index"` // 关联配置

    // 执行信息
    ExecutionID string    `json:"execution_id" gorm:"uniqueIndex;size:50"` // 执行ID
    StartTime   time.Time `json:"start_time" gorm:"index"`
    EndTime     time.Time `json:"end_time"`
    Duration    int       `json:"duration"` // 执行时长（毫秒）

    // 输入输出
    InputData   string `json:"input_data" gorm:"type:json"`    // 输入数据
    OutputData  string `json:"output_data" gorm:"type:json"`   // 输出数据

    // 执行状态
    Status      string    `json:"status" gorm:"size:20"` // success, failed, timeout
    ErrorMessage string   `json:"error_message" gorm:"type:text"`

    // 性能指标
    MemoryUsage int64     `json:"memory_usage"` // 内存使用（字节）
    CPUUsage    float64   `json:"cpu_usage"`    // CPU使用率

    CreatedAt   time.Time `json:"created_at"`
}

// ModelBenchmarkMatrix 模型基准对比矩阵
type ModelBenchmarkMatrix struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    ComparisonName  string    `json:"comparison_name" gorm:"size:100;index"` // 对比名称

    // 插件组合
    AlphaPluginID       uint   `json:"alpha_plugin_id" gorm:"index"`       // Alpha插件ID
    OptimizerPluginID   uint   `json:"optimizer_plugin_id" gorm:"index"`   // Optimizer插件ID
    RiskPluginID        uint   `json:"risk_plugin_id" gorm:"index"`        // Risk插件ID

    // 回测配置
    BacktestWindow      int       `json:"backtest_window"`        // 回测窗口（年）
    RebalanceFrequency  string    `json:"rebalance_frequency" gorm:"size:20"` // 再平衡频率
    TransactionCost     decimal.Decimal `json:"transaction_cost" gorm:"type:decimal(5,4)"` // 交易成本

    // 性能指标
    TotalReturn         decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`
    AnnualReturn        decimal.Decimal `json:"annual_return" gorm:"type:decimal(10,6)"`
    Volatility          decimal.Decimal `json:"volatility" gorm:"type:decimal(10,6)"`
    SharpeRatio         decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,6)"`
    MaxDrawdown         decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,6)"`
    CalmarRatio         decimal.Decimal `json:"calmar_ratio" gorm:"type:decimal(10,6)"`
    Rolling1YWinRate    decimal.Decimal `json:"rolling_1y_win_rate" gorm:"type:decimal(5,2)"`
    TailDependencyIndex decimal.Decimal `json:"tail_dependency_index" gorm:"type:decimal(10,6)"` // 尾部依赖指数

    // 详细指标（JSON）
    DetailedMetrics     string    `json:"detailed_metrics" gorm:"type:json"`

    ComparisonDate      time.Time `json:"comparison_date"`
    CreatedAt           time.Time `json:"created_at"`
}

// StrategyExperiment 策略实验
type StrategyExperiment struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    ExperimentName  string    `json:"experiment_name" gorm:"size:100;index"` // 实验名称
    Description     string    `json:"description" gorm:"size:500"`

    // 实验配置
    BenchmarkMatrixID uint   `json:"benchmark_matrix_id" gorm:"index"` // 关联基准对比
    AllocationRatio   decimal.Decimal `json:"allocation_ratio" gorm:"type:decimal(5,2)"` // 资金分配比例（%）

    // 实验结果
    ExperimentResult  string    `json:"experiment_result" gorm:"type:json"` // 实验结果（JSON）
    IsSuccessful      bool      `json:"is_successful"` // 是否成功
    SuccessCriteria   string    `json:"success_criteria" gorm:"type:json"` // 成功标准

    // 时间信息
    StartDate        time.Time `json:"start_date"`
    EndDate          time.Time `json:"end_date"`
    Status           string    `json:"status" gorm:"size:20"` // running, completed, failed

    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

---

## 🔄 数据迁移策略

### 迁移原则

1. **零数据删除**：所有迁移操作仅新增表和字段，不删除任何现有数据
2. **向后兼容**：新增字段设置默认值，确保现有功能不受影响
3. **渐进式迁移**：分阶段执行迁移，每个阶段独立可回滚
4. **数据验证**：迁移后执行数据完整性验证

### 迁移步骤

#### 阶段一：因子数据层（P0，预计2周）

```sql
-- 1. 创建因子数据表
CREATE TABLE factor_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factor_name VARCHAR(20) NOT NULL,
    date DATE NOT NULL,
    value DECIMAL(10,6) NOT NULL,
    data_source VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_factor_date ON factor_data(factor_name, date);

-- 2. 创建因子择时信号表
CREATE TABLE factor_timing_signals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factor_name VARCHAR(20) NOT NULL,
    signal_date DATE NOT NULL,
    ma_slope_60 DECIMAL(10,6),
    z_score DECIMAL(10,6),
    percentile DECIMAL(5,2),
    signal_strength VARCHAR(20),
    signal_score INTEGER,
    expected_return DECIMAL(10,6),
    confidence DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_signal_date ON factor_timing_signals(factor_name, signal_date);

-- 3. 数据初始化：从Fama-French库导入历史因子数据
-- （通过后端服务实现）
```

#### 阶段二：Alpha观点层（P0，预计2周）

```sql
-- 1. 创建Alpha观点表
CREATE TABLE alpha_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER,
    asset_symbol VARCHAR(20) NOT NULL,
    view_return DECIMAL(10,6) NOT NULL,
    confidence DECIMAL(5,2) NOT NULL,
    view_type VARCHAR(20),
    view_method VARCHAR(50),
    generated_at TIMESTAMP NOT NULL,
    valid_until TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    source_factor VARCHAR(20),
    factor_loading DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE INDEX idx_view_asset ON alpha_views(asset_symbol, generated_at);

-- 2. 创建Alpha观点表现表
CREATE TABLE alpha_view_performances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    view_id INTEGER NOT NULL,
    actual_return DECIMAL(10,6),
    prediction_error DECIMAL(10,6),
    is_validated BOOLEAN DEFAULT FALSE,
    validation_date TIMESTAMP,
    is_correct BOOLEAN,
    rolling_win_rate DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (view_id) REFERENCES alpha_views(id)
);

-- 3. 创建Black-Litterman配置表
CREATE TABLE black_litterman_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER UNIQUE,
    risk_aversion DECIMAL(10,6),
    prior_type VARCHAR(20),
    prior_weights TEXT,
    implied_returns TEXT,
    omega_method VARCHAR(20),
    omega_matrix TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    last_calculated TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- 4. 创建BL后验收益表
CREATE TABLE bl_posterior_returns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_id INTEGER NOT NULL,
    calculation_date TIMESTAMP NOT NULL,
    posterior_returns TEXT,
    posterior_weights TEXT,
    posterior_cov TEXT,
    num_views INTEGER,
    view_impact DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (config_id) REFERENCES black_litterman_configs(id)
);

CREATE INDEX idx_posterior_date ON bl_posterior_returns(config_id, calculation_date);
```

#### 阶段三：风险预算层（P0，预计2周）

```sql
-- 1. 创建风险预算配置表
CREATE TABLE risk_budget_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER UNIQUE,
    stock_cvar_budget DECIMAL(5,2) DEFAULT 40.00,
    bond_cvar_budget DECIMAL(5,2) DEFAULT 10.00,
    commodity_cvar_budget DECIMAL(5,2) DEFAULT 20.00,
    cash_cvar_budget DECIMAL(5,2) DEFAULT 5.00,
    use_var_constraint BOOLEAN DEFAULT FALSE,
    stock_var_budget DECIMAL(5,2),
    bond_var_budget DECIMAL(5,2),
    min_skewness DECIMAL(10,6),
    max_drawdown DECIMAL(5,2),
    cvar_confidence DECIMAL(5,4) DEFAULT 0.95,
    var_confidence DECIMAL(5,4) DEFAULT 0.95,
    is_active BOOLEAN DEFAULT TRUE,
    effective_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- 2. 创建蒙特卡洛模拟表
CREATE TABLE monte_carlo_simulations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL,
    simulation_date TIMESTAMP NOT NULL,
    num_paths INTEGER DEFAULT 10000,
    time_steps INTEGER DEFAULT 252,
    time_horizon INTEGER DEFAULT 252,
    confidence_level DECIMAL(5,4) DEFAULT 0.95,
    var_95 DECIMAL(10,6),
    var_99 DECIMAL(10,6),
    cvar_95 DECIMAL(10,6),
    cvar_99 DECIMAL(10,6),
    mean_return DECIMAL(10,6),
    std_dev DECIMAL(10,6),
    skewness DECIMAL(10,6),
    kurtosis DECIMAL(10,6),
    simulation_result TEXT,
    cache_expiry TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE INDEX idx_simulation_portfolio ON monte_carlo_simulations(portfolio_id, simulation_date);

-- 3. 创建风险贡献表
CREATE TABLE risk_contributions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    simulation_id INTEGER NOT NULL,
    asset_id INTEGER NOT NULL,
    asset_symbol VARCHAR(20),
    weight DECIMAL(5,2),
    cvar_contribution DECIMAL(10,6),
    marginal_cvar DECIMAL(10,6),
    cvar_percentage DECIMAL(5,2),
    budget_limit DECIMAL(5,2),
    budget_usage DECIMAL(5,2),
    budget_deviation DECIMAL(5,2),
    calculation_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (simulation_id) REFERENCES monte_carlo_simulations(id),
    FOREIGN KEY (asset_id) REFERENCES assets(id)
);

-- 4. 创建风险预算执行记录表
CREATE TABLE risk_budget_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL,
    execution_date TIMESTAMP NOT NULL,
    total_cvar DECIMAL(10,6),
    stock_cvar_usage DECIMAL(5,2),
    bond_cvar_usage DECIMAL(5,2),
    is_budget_satisfied BOOLEAN,
    violation_details TEXT,
    optimization_method VARCHAR(50),
    iterations INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE INDEX idx_execution_date ON risk_budget_executions(portfolio_id, execution_date);
```

#### 阶段四：插件架构层（P1，预计2周）

```sql
-- 1. 创建插件注册表
CREATE TABLE plugin_registries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_name VARCHAR(100) UNIQUE NOT NULL,
    plugin_type VARCHAR(50) NOT NULL,
    version VARCHAR(20),
    input_schema TEXT,
    output_schema TEXT,
    dependencies TEXT,
    description VARCHAR(500),
    author VARCHAR(100),
    documentation VARCHAR(200),
    status VARCHAR(20) DEFAULT 'active',
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plugin_type ON plugin_registries(plugin_type);

-- 2. 创建插件配置表
CREATE TABLE plugin_configurations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id INTEGER NOT NULL,
    portfolio_id INTEGER,
    config_name VARCHAR(100),
    parameters TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- 3. 创建插件执行日志表
CREATE TABLE plugin_execution_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id INTEGER NOT NULL,
    config_id INTEGER,
    execution_id VARCHAR(50) UNIQUE NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration INTEGER,
    input_data TEXT,
    output_data TEXT,
    status VARCHAR(20),
    error_message TEXT,
    memory_usage INTEGER,
    cpu_usage REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (config_id) REFERENCES plugin_configurations(id)
);

CREATE INDEX idx_execution_time ON plugin_execution_logs(start_time);

-- 4. 创建模型基准对比矩阵表
CREATE TABLE model_benchmark_matrices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comparison_name VARCHAR(100) NOT NULL,
    alpha_plugin_id INTEGER,
    optimizer_plugin_id INTEGER,
    risk_plugin_id INTEGER,
    backtest_window INTEGER DEFAULT 3,
    rebalance_frequency VARCHAR(20),
    transaction_cost DECIMAL(5,4),
    total_return DECIMAL(10,6),
    annual_return DECIMAL(10,6),
    volatility DECIMAL(10,6),
    sharpe_ratio DECIMAL(10,6),
    max_drawdown DECIMAL(10,6),
    calmar_ratio DECIMAL(10,6),
    rolling_1y_win_rate DECIMAL(5,2),
    tail_dependency_index DECIMAL(10,6),
    detailed_metrics TEXT,
    comparison_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alpha_plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (optimizer_plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (risk_plugin_id) REFERENCES plugin_registries(id)
);

-- 5. 创建策略实验表
CREATE TABLE strategy_experiments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experiment_name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    benchmark_matrix_id INTEGER,
    allocation_ratio DECIMAL(5,2) DEFAULT 20.00,
    experiment_result TEXT,
    is_successful BOOLEAN,
    success_criteria TEXT,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'running',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (benchmark_matrix_id) REFERENCES model_benchmark_matrices(id)
);
```

---

## 📊 性能优化方案

### 索引优化

#### 已有索引优化

```sql
-- 优化现有表的索引
-- 1. 价格表优化
CREATE INDEX IF NOT EXISTS idx_price_asset_date_type ON prices(asset_id, date, price_type);

-- 2. 组合持仓表优化
CREATE INDEX IF NOT EXISTS idx_position_portfolio_symbol ON portfolio_positions(portfolio_id, symbol);

-- 3. 持仓关系表优化
CREATE INDEX IF NOT EXISTS idx_holding_parent_date ON holdings(parent_id, effective_date);
```

#### 新增索引

```sql
-- 1. 因子数据索引
CREATE INDEX idx_factor_composite ON factor_data(factor_name, date, value);

-- 2. Alpha观点索引
CREATE INDEX idx_view_portfolio_status ON alpha_views(portfolio_id, status, generated_at);

-- 3. 风险贡献索引
CREATE INDEX idx_risk_contribution_composite ON risk_contributions(simulation_id, asset_id, cvar_percentage);

-- 4. 插件执行日志索引
CREATE INDEX idx_plugin_execution_composite ON plugin_execution_logs(plugin_id, start_time, status);
```

### 查询优化

#### 物化视图

```sql
-- 1. 组合风险指标物化视图
CREATE VIEW portfolio_risk_metrics AS
SELECT
    p.id AS portfolio_id,
    p.name AS portfolio_name,
    mcs.cvar_95,
    mcs.cvar_99,
    mcs.var_95,
    mcs.var_99,
    mcs.simulation_date
FROM portfolios p
LEFT JOIN monte_carlo_simulations mcs ON p.id = mcs.portfolio_id
WHERE mcs.simulation_date = (
    SELECT MAX(simulation_date)
    FROM monte_carlo_simulations
    WHERE portfolio_id = p.id
);

-- 2. Alpha观点胜率统计视图
CREATE VIEW alpha_view_win_rates AS
SELECT
    av.source_factor,
    av.view_method,
    COUNT(*) AS total_views,
    SUM(CASE WHEN avp.is_correct THEN 1 ELSE 0 END) AS correct_predictions,
    AVG(avp.rolling_win_rate) AS avg_rolling_win_rate
FROM alpha_views av
LEFT JOIN alpha_view_performances avp ON av.id = avp.view_id
WHERE avp.is_validated = TRUE
GROUP BY av.source_factor, av.view_method;
```

#### 分区策略

```sql
-- 对大表进行分区（PostgreSQL）
-- 1. 价格表按年分区
CREATE TABLE prices_partitioned (
    LIKE prices INCLUDING DEFAULTS INCLUDING CONSTRAINTS
) PARTITION BY RANGE (date);

CREATE TABLE prices_2024 PARTITION OF prices_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE prices_2025 PARTITION OF prices_partitioned
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

-- 2. 插件执行日志按月分区
CREATE TABLE plugin_execution_logs_partitioned (
    LIKE plugin_execution_logs INCLUDING DEFAULTS INCLUDING CONSTRAINTS
) PARTITION BY RANGE (start_time);
```

### 缓存策略

#### Redis缓存设计

```yaml
# 缓存键设计
cache_keys:
  # 因子数据缓存
  factor_data: "factor:{factor_name}:{date}"
  factor_timing: "factor_timing:{factor_name}:{signal_date}"

  # Alpha观点缓存
  alpha_views: "alpha_views:portfolio:{portfolio_id}:active"
  alpha_performance: "alpha_performance:{view_id}"

  # 风险预算缓存
  monte_carlo: "monte_carlo:portfolio:{portfolio_id}:{date}"
  risk_contribution: "risk_contribution:simulation:{simulation_id}"

  # 插件结果缓存
  plugin_result: "plugin:{plugin_id}:execution:{execution_id}"

# 缓存过期时间
cache_ttl:
  factor_data: 86400      # 1天
  factor_timing: 3600     # 1小时
  alpha_views: 300        # 5分钟
  monte_carlo: 7200       # 2小时
  plugin_result: 1800     # 30分钟
```

---

## 🔐 安全加固方案

### 数据访问控制

#### 角色权限设计

```sql
-- 1. 创建角色
CREATE ROLE factor_analyst;
CREATE ROLE portfolio_manager;
CREATE ROLE risk_manager;
CREATE ROLE plugin_admin;

-- 2. 授予权限
-- 因子分析师
GRANT SELECT ON factor_data TO factor_analyst;
GRANT SELECT ON factor_timing_signals TO factor_analyst;
GRANT INSERT, UPDATE ON alpha_views TO factor_analyst;

-- 组合经理
GRANT SELECT, INSERT, UPDATE ON portfolios TO portfolio_manager;
GRANT SELECT, INSERT, UPDATE ON portfolio_positions TO portfolio_manager;
GRANT SELECT ON alpha_views TO portfolio_manager;

-- 风险经理
GRANT SELECT, INSERT, UPDATE ON risk_budget_configs TO risk_manager;
GRANT SELECT, INSERT ON monte_carlo_simulations TO risk_manager;
GRANT SELECT ON risk_contributions TO risk_manager;

-- 插件管理员
GRANT ALL ON plugin_registries TO plugin_admin;
GRANT ALL ON plugin_configurations TO plugin_admin;
GRANT SELECT ON plugin_execution_logs TO plugin_admin;
```

### 数据加密

#### 敏感字段加密

```go
// 敏感数据加密
type EncryptedField struct {
    Value     []byte `json:"value"`      // 加密后的值
    KeyID     string `json:"key_id"`     // 密钥ID
    Algorithm string `json:"algorithm"`  // 加密算法
    IV        []byte `json:"iv"`         // 初始化向量
}

// 需要加密的字段
// 1. Black-Litterman配置中的先验权重
// 2. 插件配置中的敏感参数
// 3. API密钥等认证信息
```

### 审计日志

#### 数据变更审计

```sql
-- 创建审计日志表
CREATE TABLE data_audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    table_name VARCHAR(50) NOT NULL,
    record_id INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values TEXT,             -- JSON格式
    new_values TEXT,             -- JSON格式
    changed_by VARCHAR(100),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(50),
    user_agent TEXT
);

CREATE INDEX idx_audit_table_record ON data_audit_logs(table_name, record_id);
CREATE INDEX idx_audit_time ON data_audit_logs(changed_at);
```

---

## 📈 实施计划

### 时间表

| 阶段 | 任务 | 预计时间 | 负责人 | 依赖 |
|------|------|----------|--------|------|
| **阶段一** | 因子数据层迁移 | 2周 | Backend | 无 |
| **阶段二** | Alpha观点层迁移 | 2周 | Backend | 阶段一 |
| **阶段三** | 风险预算层迁移 | 2周 | Backend | 无 |
| **阶段四** | 插件架构层迁移 | 2周 | Backend | 阶段一、二、三 |
| **阶段五** | 性能优化 | 1周 | Backend | 阶段四 |
| **阶段六** | 安全加固 | 1周 | Backend | 阶段四 |
| **阶段七** | 测试与验证 | 1周 | QA | 阶段五、六 |

### 里程碑

#### 里程碑1：因子数据层完成（第2周）

- ✅ 因子数据表创建完成
- ✅ 因子择时信号表创建完成
- ✅ Fama-French历史数据导入完成
- ✅ 因子择时服务对接完成
- ✅ 单元测试覆盖率 > 80%

#### 里程碑2：Alpha观点层完成（第4周）

- ✅ Alpha观点表创建完成
- ✅ Black-Litterman配置表创建完成
- ✅ Alpha观点生成服务对接完成
- ✅ BL模型优化器对接完成
- ✅ 集成测试通过

#### 里程碑3：风险预算层完成（第6周）

- ✅ 风险预算配置表创建完成
- ✅ 蒙特卡洛模拟表创建完成
- ✅ 风险贡献分解服务对接完成
- ✅ CVaR预算优化器对接完成
- ✅ 性能测试通过

#### 里程碑4：插件架构层完成（第8周）

- ✅ 插件注册表创建完成
- ✅ 插件配置管理完成
- ✅ 插件执行日志完成
- ✅ 模型基准对比矩阵完成
- ✅ 策略实验台可用

#### 里程碑5：系统上线（第11周）

- ✅ 所有数据迁移完成
- ✅ 性能优化完成
- ✅ 安全加固完成
- ✅ 全量测试通过
- ✅ 文档更新完成

---

## ✅ 验收标准

### 功能验收

| 功能点 | 验收标准 | 测试方法 |
|--------|----------|----------|
| **因子数据管理** | 能存储和查询所有Fama-French因子历史数据 | 数据完整性测试 |
| **因子择时信号** | 能正确计算60日斜率和Z-score | 单元测试 |
| **Alpha观点生成** | 能生成符合BL模型格式的观点 | 集成测试 |
| **Black-Litterman** | 能正确融合观点并输出后验收益 | 回归测试 |
| **风险预算配置** | 能设置和验证CVaR预算约束 | 单元测试 |
| **蒙特卡洛模拟** | 能缓存模拟结果并正确计算CVaR | 性能测试 |
| **插件注册** | 能注册和管理三种类型插件 | 功能测试 |
| **模型基准对比** | 能生成完整的对比矩阵 | 集成测试 |

### 性能验收

| 指标 | 目标值 | 测试方法 |
|------|--------|----------|
| **因子数据查询** | < 50ms | 性能测试 |
| **Alpha观点生成** | < 200ms | 性能测试 |
| **BL模型计算** | < 500ms | 性能测试 |
| **蒙特卡洛模拟** | < 5s（10000路径） | 性能测试 |
| **风险贡献分解** | < 300ms | 性能测试 |
| **插件执行** | < 1s（平均） | 性能测试 |

### 数据完整性验收

| 检查项 | 验收标准 | 测试方法 |
|--------|----------|----------|
| **数据一致性** | 所有外键约束有效 | 完整性测试 |
| **数据完整性** | 无NULL值（除允许字段） | 数据验证 |
| **索引有效性** | 所有索引正常工作 | 查询计划分析 |
| **缓存一致性** | 缓存与数据库一致 | 缓存测试 |

---

## 📚 相关文档

- [ETF-Insight 演进路线图 2026](../roadmap/EVOLUTION_ROADMAP_2026.md)
- [数据库设计文档](../architecture/DATABASE_DESIGN.md)
- [API接口文档](http://localhost:8080/swagger)
- [插件开发指南](../guides/PLUGIN_DEVELOPMENT.md)

---

**文档版本**: v1.0
**最后更新**: 2026-04-25
**维护者**: ETF-Insight Team
