# 数据层改造实施指南

**版本**: v1.0
**创建日期**: 2026-04-25
**适用阶段**: v2.7 → v2.8

---

## 📋 概述

本文档提供数据层改造的具体实施指南，包括：
1. Go模型定义代码
2. 数据库迁移脚本
3. 服务层接口实现
4. 测试用例示例

---

## 1. Go模型定义

### 1.1 因子数据模型

创建文件：`backend/models/factor.go`

```go
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// FactorType 因子类型
type FactorType string

const (
	FactorTypeMarket FactorType = "Mkt-RF" // 市场因子
	FactorTypeSMB    FactorType = "SMB"    // 市值因子
	FactorTypeHML    FactorType = "HML"    // 价值因子
	FactorTypeRMW    FactorType = "RMW"    // 盈利因子
	FactorTypeCMA    FactorType = "CMA"    // 投资因子
)

// FactorData 因子数据
type FactorData struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	FactorName string          `json:"factor_name" gorm:"size:20;index:idx_factor_date"`
	Date       time.Time       `json:"date" gorm:"index:idx_factor_date"`
	Value      decimal.Decimal `json:"value" gorm:"type:decimal(10,6)"`
	DataSource string          `json:"data_source" gorm:"size:50"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (FactorData) TableName() string {
	return "factor_data"
}

// SignalStrength 信号强度
type SignalStrength string

const (
	SignalStrengthStrongPositive SignalStrength = "strong_positive"
	SignalStrengthWeakPositive   SignalStrength = "weak_positive"
	SignalStrengthNeutral        SignalStrength = "neutral"
	SignalStrengthWeakNegative   SignalStrength = "weak_negative"
	SignalStrengthStrongNegative SignalStrength = "strong_negative"
)

// FactorTimingSignal 因子择时信号
type FactorTimingSignal struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	FactorName    string          `json:"factor_name" gorm:"size:20;index:idx_signal_date"`
	SignalDate    time.Time       `json:"signal_date" gorm:"index:idx_signal_date"`

	// 择时指标
	MASlope60     decimal.Decimal `json:"ma_slope_60" gorm:"type:decimal(10,6)"`
	ZScore        decimal.Decimal `json:"z_score" gorm:"type:decimal(10,6)"`
	Percentile    decimal.Decimal `json:"percentile" gorm:"type:decimal(5,2)"`

	// 信号强度
	SignalStrength SignalStrength `json:"signal_strength" gorm:"size:20"`
	SignalScore    int            `json:"signal_score"`

	// 预期收益
	ExpectedReturn decimal.Decimal `json:"expected_return" gorm:"type:decimal(10,6)"`
	Confidence     decimal.Decimal `json:"confidence" gorm:"type:decimal(5,2)"`

	CreatedAt      time.Time `json:"created_at"`
}

func (FactorTimingSignal) TableName() string {
	return "factor_timing_signals"
}
```

### 1.2 Alpha观点模型

创建文件：`backend/models/alpha_view.go`

```go
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ViewType 观点类型
type ViewType string

const (
	ViewTypeAbsolute ViewType = "absolute" // 绝对观点
	ViewTypeRelative ViewType = "relative" // 相对观点
)

// ViewMethod 观点生成方法
type ViewMethod string

const (
	ViewMethodFactorTiming  ViewMethod = "factor_timing"  // 因子择时
	ViewMethodMomentum      ViewMethod = "momentum"       // 动量
	ViewMethodMeanReversion ViewMethod = "mean_reversion" // 均值回复
)

// ViewStatus 观点状态
type ViewStatus string

const (
	ViewStatusActive    ViewStatus = "active"    // 活跃
	ViewStatusExpired   ViewStatus = "expired"   // 过期
	ViewStatusValidated ViewStatus = "validated" // 已验证
)

// AlphaView Alpha观点
type AlphaView struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	PortfolioID  uint            `json:"portfolio_id" gorm:"index"`

	// 观点内容
	AssetSymbol  string          `json:"asset_symbol" gorm:"size:20;index:idx_view_asset"`
	ViewReturn   decimal.Decimal `json:"view_return" gorm:"type:decimal(10,6)"`
	Confidence   decimal.Decimal `json:"confidence" gorm:"type:decimal(5,2)"`

	// 观点类型
	ViewType     ViewType        `json:"view_type" gorm:"size:20"`
	ViewMethod   ViewMethod      `json:"view_method" gorm:"size:50"`

	// 生成信息
	GeneratedAt  time.Time       `json:"generated_at"`
	ValidUntil   time.Time       `json:"valid_until"`
	Status       ViewStatus      `json:"status" gorm:"size:20;default:'active'"`

	// 因子来源
	SourceFactor string          `json:"source_factor" gorm:"size:20"`
	FactorLoading decimal.Decimal `json:"factor_loading" gorm:"type:decimal(10,6)"`

	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	Portfolio    Portfolio           `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
	Performance  *AlphaViewPerformance `json:"performance,omitempty" gorm:"foreignKey:ViewID"`
}

func (AlphaView) TableName() string {
	return "alpha_views"
}

// AlphaViewPerformance Alpha观点表现
type AlphaViewPerformance struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	ViewID           uint            `json:"view_id" gorm:"index"`

	// 实际表现
	ActualReturn     decimal.Decimal `json:"actual_return" gorm:"type:decimal(10,6)"`
	PredictionError  decimal.Decimal `json:"prediction_error" gorm:"type:decimal(10,6)"`

	// 验证结果
	IsValidated      bool            `json:"is_validated"`
	ValidationDate   time.Time       `json:"validation_date"`
	IsCorrect        bool            `json:"is_correct"`

	// 滚动统计
	RollingWinRate   decimal.Decimal `json:"rolling_win_rate" gorm:"type:decimal(5,2)"`

	CreatedAt        time.Time `json:"created_at"`

	// 关联
	View             AlphaView `json:"view,omitempty" gorm:"foreignKey:ViewID"`
}

func (AlphaViewPerformance) TableName() string {
	return "alpha_view_performances"
}
```

### 1.3 Black-Litterman配置模型

创建文件：`backend/models/black_litterman.go`

```go
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PriorType 先验类型
type PriorType string

const (
	PriorTypeEqualWeight  PriorType = "equal_weight"  // 等权
	PriorTypeMinVariance  PriorType = "min_variance"  // 最小方差
	PriorTypeMarketCap    PriorType = "market_cap"    // 市值加权
)

// OmegaMethod 观点误差矩阵方法
type OmegaMethod string

const (
	OmegaMethodIdzorek     OmegaMethod = "Idzorek"      // Idzorek方法
	OmegaMethodHeLitterman OmegaMethod = "HeLitterman"  // He-Litterman方法
)

// BlackLittermanConfig Black-Litterman模型配置
type BlackLittermanConfig struct {
	ID             uint            `json:"id" gorm:"primaryKey"`
	PortfolioID    uint            `json:"portfolio_id" gorm:"uniqueIndex"`

	// 模型参数
	RiskAversion   decimal.Decimal `json:"risk_aversion" gorm:"type:decimal(10,6)"`
	PriorType      PriorType       `json:"prior_type" gorm:"size:20"`

	// 先验基准
	PriorWeights   string          `json:"prior_weights" gorm:"type:json"`
	ImpliedReturns string          `json:"implied_returns" gorm:"type:json"`

	// 观点误差矩阵
	OmegaMethod    OmegaMethod     `json:"omega_method" gorm:"size:20"`
	OmegaMatrix    string          `json:"omega_matrix" gorm:"type:json"`

	// 配置状态
	IsActive       bool            `json:"is_active" gorm:"default:true"`
	LastCalculated time.Time       `json:"last_calculated"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联
	Portfolio      Portfolio `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
}

func (BlackLittermanConfig) TableName() string {
	return "black_litterman_configs"
}

// BLPosteriorReturn BL后验收益
type BLPosteriorReturn struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ConfigID         uint      `json:"config_id" gorm:"index"`
	CalculationDate  time.Time `json:"calculation_date" gorm:"index"`

	// 后验结果
	PosteriorReturns string          `json:"posterior_returns" gorm:"type:json"`
	PosteriorWeights string          `json:"posterior_weights" gorm:"type:json"`
	PosteriorCov     string          `json:"posterior_cov" gorm:"type:json"`

	// 观点融合信息
	NumViews         int             `json:"num_views"`
	ViewImpact       decimal.Decimal `json:"view_impact" gorm:"type:decimal(10,6)"`

	CreatedAt        time.Time `json:"created_at"`

	// 关联
	Config           BlackLittermanConfig `json:"config,omitempty" gorm:"foreignKey:ConfigID"`
}

func (BLPosteriorReturn) TableName() string {
	return "bl_posterior_returns"
}
```

### 1.4 风险预算模型

创建文件：`backend/models/risk_budget.go`

```go
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// RiskBudgetConfig 风险预算配置
type RiskBudgetConfig struct {
	ID                    uint            `json:"id" gorm:"primaryKey"`
	PortfolioID           uint            `json:"portfolio_id" gorm:"uniqueIndex"`

	// CVaR预算
	StockCVaRBudget       decimal.Decimal `json:"stock_cvar_budget" gorm:"type:decimal(5,2)"`
	BondCVaRBudget        decimal.Decimal `json:"bond_cvar_budget" gorm:"type:decimal(5,2)"`
	CommodityCVaRBudget   decimal.Decimal `json:"commodity_cvar_budget" gorm:"type:decimal(5,2)"`
	CashCVaRBudget        decimal.Decimal `json:"cash_cvar_budget" gorm:"type:decimal(5,2)"`

	// VaR预算（可选）
	UseVaRConstraint      bool            `json:"use_var_constraint" gorm:"default:false"`
	StockVaRBudget        decimal.Decimal `json:"stock_var_budget" gorm:"type:decimal(5,2)"`
	BondVaRBudget         decimal.Decimal `json:"bond_var_budget" gorm:"type:decimal(5,2)"`

	// 其他约束
	MinSkewness           decimal.Decimal `json:"min_skewness" gorm:"type:decimal(10,6)"`
	MaxDrawdown           decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(5,2)"`

	// 置信水平
	CVaRConfidence        decimal.Decimal `json:"cvar_confidence" gorm:"type:decimal(5,4);default:0.95"`
	VaRConfidence         decimal.Decimal `json:"var_confidence" gorm:"type:decimal(5,4);default:0.95"`

	// 配置状态
	IsActive              bool      `json:"is_active" gorm:"default:true"`
	EffectiveDate         time.Time `json:"effective_date"`

	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// 关联
	Portfolio             Portfolio `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
}

func (RiskBudgetConfig) TableName() string {
	return "risk_budget_configs"
}

// MonteCarloSimulation 蒙特卡洛模拟
type MonteCarloSimulation struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	PortfolioID      uint            `json:"portfolio_id" gorm:"index:idx_simulation_portfolio"`
	SimulationDate   time.Time       `json:"simulation_date" gorm:"index:idx_simulation_portfolio"`

	// 模拟参数
	NumPaths         int             `json:"num_paths"`
	TimeSteps        int             `json:"time_steps"`
	TimeHorizon      int             `json:"time_horizon"`
	ConfidenceLevel  decimal.Decimal `json:"confidence_level" gorm:"type:decimal(5,4)"`

	// 风险指标
	VaR95            decimal.Decimal `json:"var_95" gorm:"type:decimal(10,6)"`
	VaR99            decimal.Decimal `json:"var_99" gorm:"type:decimal(10,6)"`
	CVaR95           decimal.Decimal `json:"cvar_95" gorm:"type:decimal(10,6)"`
	CVaR99           decimal.Decimal `json:"cvar_99" gorm:"type:decimal(10,6)"`

	// 分布参数
	MeanReturn       decimal.Decimal `json:"mean_return" gorm:"type:decimal(10,6)"`
	StdDev           decimal.Decimal `json:"std_dev" gorm:"type:decimal(10,6)"`
	Skewness         decimal.Decimal `json:"skewness" gorm:"type:decimal(10,6)"`
	Kurtosis         decimal.Decimal `json:"kurtosis" gorm:"type:decimal(10,6)"`

	// 模拟结果（缓存）
	SimulationResult string          `json:"simulation_result" gorm:"type:json"`
	CacheExpiry      time.Time       `json:"cache_expiry"`

	CreatedAt        time.Time `json:"created_at"`

	// 关联
	Portfolio        Portfolio `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
}

func (MonteCarloSimulation) TableName() string {
	return "monte_carlo_simulations"
}

// RiskContribution 风险贡献
type RiskContribution struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	SimulationID     uint            `json:"simulation_id" gorm:"index"`
	AssetID          uint            `json:"asset_id" gorm:"index"`
	AssetSymbol      string          `json:"asset_symbol" gorm:"size:20"`

	// 风险贡献
	Weight           decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`
	CVaRContribution decimal.Decimal `json:"cvar_contribution" gorm:"type:decimal(10,6)"`
	MarginalCVaR     decimal.Decimal `json:"marginal_cvar" gorm:"type:decimal(10,6)"`
	CVaRPercentage   decimal.Decimal `json:"cvar_percentage" gorm:"type:decimal(5,2)"`

	// 预算对比
	BudgetLimit      decimal.Decimal `json:"budget_limit" gorm:"type:decimal(5,2)"`
	BudgetUsage      decimal.Decimal `json:"budget_usage" gorm:"type:decimal(5,2)"`
	BudgetDeviation  decimal.Decimal `json:"budget_deviation" gorm:"type:decimal(5,2)"`

	CalculationDate  time.Time `json:"calculation_date"`
	CreatedAt        time.Time `json:"created_at"`

	// 关联
	Simulation       MonteCarloSimulation `json:"simulation,omitempty" gorm:"foreignKey:SimulationID"`
	Asset            Asset                `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
}

func (RiskContribution) TableName() string {
	return "risk_contributions"
}

// RiskBudgetExecution 风险预算执行记录
type RiskBudgetExecution struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	PortfolioID       uint            `json:"portfolio_id" gorm:"index"`
	ExecutionDate     time.Time       `json:"execution_date" gorm:"index"`

	// 执行情况
	TotalCVaR         decimal.Decimal `json:"total_cvar" gorm:"type:decimal(10,6)"`
	StockCVaRUsage    decimal.Decimal `json:"stock_cvar_usage" gorm:"type:decimal(5,2)"`
	BondCVaRUsage     decimal.Decimal `json:"bond_cvar_usage" gorm:"type:decimal(5,2)"`

	// 约束满足情况
	IsBudgetSatisfied bool            `json:"is_budget_satisfied"`
	ViolationDetails  string          `json:"violation_details" gorm:"type:json"`

	// 优化信息
	OptimizationMethod string         `json:"optimization_method" gorm:"size:50"`
	Iterations         int            `json:"iterations"`

	CreatedAt         time.Time `json:"created_at"`

	// 关联
	Portfolio         Portfolio `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
}

func (RiskBudgetExecution) TableName() string {
	return "risk_budget_executions"
}
```

### 1.5 插件架构模型

创建文件：`backend/models/plugin.go`

```go
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// PluginType 插件类型
type PluginType string

const (
	PluginTypeAlphaGenerator     PluginType = "AlphaGenerator"
	PluginTypePortfolioOptimizer PluginType = "PortfolioOptimizer"
	PluginTypeRiskModel          PluginType = "RiskModel"
)

// PluginStatus 插件状态
type PluginStatus string

const (
	PluginStatusActive     PluginStatus = "active"
	PluginStatusDeprecated PluginStatus = "deprecated"
	PluginStatusDisabled   PluginStatus = "disabled"
)

// PluginRegistry 插件注册表
type PluginRegistry struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	PluginName    string       `json:"plugin_name" gorm:"uniqueIndex;size:100"`
	PluginType    PluginType   `json:"plugin_type" gorm:"size:50;index"`
	Version       string       `json:"version" gorm:"size:20"`

	// 接口定义
	InputSchema   string       `json:"input_schema" gorm:"type:json"`
	OutputSchema  string       `json:"output_schema" gorm:"type:json"`

	// 依赖关系
	Dependencies  string       `json:"dependencies" gorm:"type:json"`

	// 元数据
	Description   string       `json:"description" gorm:"size:500"`
	Author        string       `json:"author" gorm:"size:100"`
	Documentation string       `json:"documentation" gorm:"size:200"`

	// 状态
	Status        PluginStatus `json:"status" gorm:"size:20;default:'active'"`
	RegisteredAt  time.Time    `json:"registered_at"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func (PluginRegistry) TableName() string {
	return "plugin_registries"
}

// PluginConfiguration 插件配置
type PluginConfiguration struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PluginID    uint      `json:"plugin_id" gorm:"index"`
	PortfolioID uint      `json:"portfolio_id" gorm:"index"`

	// 配置参数
	ConfigName  string    `json:"config_name" gorm:"size:100"`
	Parameters  string    `json:"parameters" gorm:"type:json"`

	// 状态
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Plugin      PluginRegistry `json:"plugin,omitempty" gorm:"foreignKey:PluginID"`
	Portfolio   Portfolio      `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID"`
}

func (PluginConfiguration) TableName() string {
	return "plugin_configurations"
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusTimeout ExecutionStatus = "timeout"
)

// PluginExecutionLog 插件执行日志
type PluginExecutionLog struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	PluginID     uint            `json:"plugin_id" gorm:"index"`
	ConfigID     uint            `json:"config_id" gorm:"index"`

	// 执行信息
	ExecutionID  string          `json:"execution_id" gorm:"uniqueIndex;size:50"`
	StartTime    time.Time       `json:"start_time" gorm:"index"`
	EndTime      time.Time       `json:"end_time"`
	Duration     int             `json:"duration"`

	// 输入输出
	InputData    string          `json:"input_data" gorm:"type:json"`
	OutputData   string          `json:"output_data" gorm:"type:json"`

	// 执行状态
	Status       ExecutionStatus `json:"status" gorm:"size:20"`
	ErrorMessage string          `json:"error_message" gorm:"type:text"`

	// 性能指标
	MemoryUsage  int64           `json:"memory_usage"`
	CPUUsage     float64         `json:"cpu_usage"`

	CreatedAt    time.Time `json:"created_at"`

	// 关联
	Plugin       PluginRegistry      `json:"plugin,omitempty" gorm:"foreignKey:PluginID"`
	Config       PluginConfiguration `json:"config,omitempty" gorm:"foreignKey:ConfigID"`
}

func (PluginExecutionLog) TableName() string {
	return "plugin_execution_logs"
}

// ModelBenchmarkMatrix 模型基准对比矩阵
type ModelBenchmarkMatrix struct {
	ID                   uint            `json:"id" gorm:"primaryKey"`
	ComparisonName       string          `json:"comparison_name" gorm:"size:100;index"`

	// 插件组合
	AlphaPluginID        uint            `json:"alpha_plugin_id" gorm:"index"`
	OptimizerPluginID    uint            `json:"optimizer_plugin_id" gorm:"index"`
	RiskPluginID         uint            `json:"risk_plugin_id" gorm:"index"`

	// 回测配置
	BacktestWindow       int             `json:"backtest_window"`
	RebalanceFrequency   string          `json:"rebalance_frequency" gorm:"size:20"`
	TransactionCost      decimal.Decimal `json:"transaction_cost" gorm:"type:decimal(5,4)"`

	// 性能指标
	TotalReturn          decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`
	AnnualReturn         decimal.Decimal `json:"annual_return" gorm:"type:decimal(10,6)"`
	Volatility           decimal.Decimal `json:"volatility" gorm:"type:decimal(10,6)"`
	SharpeRatio          decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,6)"`
	MaxDrawdown          decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,6)"`
	CalmarRatio          decimal.Decimal `json:"calmar_ratio" gorm:"type:decimal(10,6)"`
	Rolling1YWinRate     decimal.Decimal `json:"rolling_1y_win_rate" gorm:"type:decimal(5,2)"`
	TailDependencyIndex  decimal.Decimal `json:"tail_dependency_index" gorm:"type:decimal(10,6)"`

	// 详细指标
	DetailedMetrics      string          `json:"detailed_metrics" gorm:"type:json"`

	ComparisonDate       time.Time `json:"comparison_date"`
	CreatedAt            time.Time `json:"created_at"`

	// 关联
	AlphaPlugin          PluginRegistry `json:"alpha_plugin,omitempty" gorm:"foreignKey:AlphaPluginID"`
	OptimizerPlugin      PluginRegistry `json:"optimizer_plugin,omitempty" gorm:"foreignKey:OptimizerPluginID"`
	RiskPlugin           PluginRegistry `json:"risk_plugin,omitempty" gorm:"foreignKey:RiskPluginID"`
}

func (ModelBenchmarkMatrix) TableName() string {
	return "model_benchmark_matrices"
}

// ExperimentStatus 实验状态
type ExperimentStatus string

const (
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusFailed    ExperimentStatus = "failed"
)

// StrategyExperiment 策略实验
type StrategyExperiment struct {
	ID                uint             `json:"id" gorm:"primaryKey"`
	ExperimentName    string           `json:"experiment_name" gorm:"size:100;index"`
	Description       string           `json:"description" gorm:"size:500"`

	// 实验配置
	BenchmarkMatrixID uint             `json:"benchmark_matrix_id" gorm:"index"`
	AllocationRatio   decimal.Decimal  `json:"allocation_ratio" gorm:"type:decimal(5,2)"`

	// 实验结果
	ExperimentResult  string           `json:"experiment_result" gorm:"type:json"`
	IsSuccessful      bool             `json:"is_successful"`
	SuccessCriteria   string           `json:"success_criteria" gorm:"type:json"`

	// 时间信息
	StartDate         time.Time        `json:"start_date"`
	EndDate           time.Time        `json:"end_date"`
	Status            ExperimentStatus `json:"status" gorm:"size:20"`

	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// 关联
	BenchmarkMatrix   ModelBenchmarkMatrix `json:"benchmark_matrix,omitempty" gorm:"foreignKey:BenchmarkMatrixID"`
}

func (StrategyExperiment) TableName() string {
	return "strategy_experiments"
}
```

---

## 2. 数据库迁移脚本

### 2.1 更新AutoMigrate

更新文件：`backend/models/db.go`

```go
// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	return DB.AutoMigrate(
		// 原有模型
		&ETFConfig{},
		&ETFData{},
		&OperationLog{},
		&AuditLog{},
		&ExchangeRate{},
		&AShareDividendETF{},
		&AShareETFPortfolio{},
		&ASharePortfolioHolding{},
		&PortfolioConfig{},
		&UniversalETF{},
		&ETFClassification{},
		&ETFConstituent{},
		&ETFHistoricalData{},
		&ETFDividend{},
		&ETFAssetAllocation{},
		// 新增统一数据层模型
		&Asset{},
		&AssetMetadata{},
		&Holding{},
		&HoldingSnapshot{},
		&ETFHoldingsReport{},
		&HoldingChange{},
		&Price{},
		&Portfolio{},
		&PortfolioPosition{},
		// 新增因子数据层模型
		&FactorData{},
		&FactorTimingSignal{},
		// 新增Alpha观点层模型
		&AlphaView{},
		&AlphaViewPerformance{},
		&BlackLittermanConfig{},
		&BLPosteriorReturn{},
		// 新增风险预算层模型
		&RiskBudgetConfig{},
		&MonteCarloSimulation{},
		&RiskContribution{},
		&RiskBudgetExecution{},
		// 新增插件架构层模型
		&PluginRegistry{},
		&PluginConfiguration{},
		&PluginExecutionLog{},
		&ModelBenchmarkMatrix{},
		&StrategyExperiment{},
	)
}
```

### 2.2 手动迁移脚本

创建文件：`backend/migrations/001_add_factor_tables.sql`

```sql
-- 阶段一：因子数据层迁移
-- 执行时间：2026-Q4 第1-2周

-- 1. 创建因子数据表
CREATE TABLE IF NOT EXISTS factor_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factor_name VARCHAR(20) NOT NULL,
    date DATE NOT NULL,
    value DECIMAL(10,6) NOT NULL,
    data_source VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_factor_date ON factor_data(factor_name, date);

-- 2. 创建因子择时信号表
CREATE TABLE IF NOT EXISTS factor_timing_signals (
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

CREATE INDEX IF NOT EXISTS idx_signal_date ON factor_timing_signals(factor_name, signal_date);

-- 3. 插入默认因子数据（示例）
-- 实际数据需要从Fama-French库导入
INSERT INTO factor_data (factor_name, date, value, data_source) VALUES
('Mkt-RF', '2024-01-01', 0.0052, 'Fama-French Library'),
('SMB', '2024-01-01', 0.0018, 'Fama-French Library'),
('HML', '2024-01-01', 0.0023, 'Fama-French Library');
```

创建文件：`backend/migrations/002_add_alpha_view_tables.sql`

```sql
-- 阶段二：Alpha观点层迁移
-- 执行时间：2026-Q4 第3-4周

-- 1. 创建Alpha观点表
CREATE TABLE IF NOT EXISTS alpha_views (
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

CREATE INDEX IF NOT EXISTS idx_view_asset ON alpha_views(asset_symbol, generated_at);

-- 2. 创建Alpha观点表现表
CREATE TABLE IF NOT EXISTS alpha_view_performances (
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
CREATE TABLE IF NOT EXISTS black_litterman_configs (
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
CREATE TABLE IF NOT EXISTS bl_posterior_returns (
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

CREATE INDEX IF NOT EXISTS idx_posterior_date ON bl_posterior_returns(config_id, calculation_date);
```

创建文件：`backend/migrations/003_add_risk_budget_tables.sql`

```sql
-- 阶段三：风险预算层迁移
-- 执行时间：2026-Q4 第5-6周

-- 1. 创建风险预算配置表
CREATE TABLE IF NOT EXISTS risk_budget_configs (
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
CREATE TABLE IF NOT EXISTS monte_carlo_simulations (
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

CREATE INDEX IF NOT EXISTS idx_simulation_portfolio ON monte_carlo_simulations(portfolio_id, simulation_date);

-- 3. 创建风险贡献表
CREATE TABLE IF NOT EXISTS risk_contributions (
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
CREATE TABLE IF NOT EXISTS risk_budget_executions (
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

CREATE INDEX IF NOT EXISTS idx_execution_date ON risk_budget_executions(portfolio_id, execution_date);
```

创建文件：`backend/migrations/004_add_plugin_tables.sql`

```sql
-- 阶段四：插件架构层迁移
-- 执行时间：2027-Q1 第1-2周

-- 1. 创建插件注册表
CREATE TABLE IF NOT EXISTS plugin_registries (
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

CREATE INDEX IF NOT EXISTS idx_plugin_type ON plugin_registries(plugin_type);

-- 2. 创建插件配置表
CREATE TABLE IF NOT EXISTS plugin_configurations (
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
CREATE TABLE IF NOT EXISTS plugin_execution_logs (
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

CREATE INDEX IF NOT EXISTS idx_execution_time ON plugin_execution_logs(start_time);

-- 4. 创建模型基准对比矩阵表
CREATE TABLE IF NOT EXISTS model_benchmark_matrices (
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
CREATE TABLE IF NOT EXISTS strategy_experiments (
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

-- 6. 插入默认插件注册数据
INSERT INTO plugin_registries (plugin_name, plugin_type, version, description, author, status) VALUES
('FamaFrenchAlpha', 'AlphaGenerator', '1.0.0', 'Fama-French因子择时Alpha生成器', 'ETF-Insight Team', 'active'),
('MPTOptimizer', 'PortfolioOptimizer', '1.0.0', '马科维茨均值-方差优化器', 'ETF-Insight Team', 'active'),
('RiskParityOptimizer', 'PortfolioOptimizer', '1.0.0', '风险平价优化器', 'ETF-Insight Team', 'active'),
('BlackLittermanOptimizer', 'PortfolioOptimizer', '1.0.0', 'Black-Litterman优化器', 'ETF-Insight Team', 'active'),
('CVaRRiskModel', 'RiskModel', '1.0.0', 'CVaR风险模型', 'ETF-Insight Team', 'active');
```

---

## 3. 服务层接口实现

### 3.1 因子数据服务

创建文件：`backend/services/factor/factor_data_service.go`

```go
package factor

import (
	"etf-insight/models"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FactorDataService 因子数据服务
type FactorDataService struct {
	db *gorm.DB
}

// NewFactorDataService 创建因子数据服务
func NewFactorDataService(db *gorm.DB) *FactorDataService {
	return &FactorDataService{db: db}
}

// SaveFactorData 保存因子数据
func (s *FactorDataService) SaveFactorData(factorName string, date time.Time, value decimal.Decimal, dataSource string) error {
	factorData := models.FactorData{
		FactorName: factorName,
		Date:       date,
		Value:      value,
		DataSource: dataSource,
		CreatedAt:  time.Now(),
	}
	return s.db.Create(&factorData).Error
}

// GetFactorData 获取因子数据
func (s *FactorDataService) GetFactorData(factorName string, startDate, endDate time.Time) ([]models.FactorData, error) {
	var factorData []models.FactorData
	err := s.db.Where("factor_name = ? AND date BETWEEN ? AND ?", factorName, startDate, endDate).
		Order("date ASC").
		Find(&factorData).Error
	return factorData, err
}

// CalculateTimingSignal 计算因子择时信号
func (s *FactorDataService) CalculateTimingSignal(factorName string, signalDate time.Time) (*models.FactorTimingSignal, error) {
	// 获取过去60个交易日的因子数据
	startDate := signalDate.AddDate(0, 0, -90) // 预留缓冲
	factorData, err := s.GetFactorData(factorName, startDate, signalDate)
	if err != nil {
		return nil, err
	}

	if len(factorData) < 60 {
		return nil, fmt.Errorf("insufficient data for timing signal calculation")
	}

	// 取最近60个数据点
	recentData := factorData[len(factorData)-60:]

	// 计算60日移动平均斜率
	maSlope := calculateMASlope(recentData)

	// 计算Z-score
	zScore := calculateZScore(recentData)

	// 计算历史百分位
	percentile := calculatePercentile(recentData)

	// 确定信号强度
	signalStrength := determineSignalStrength(zScore)
	signalScore := getSignalScore(signalStrength)

	// 计算预期收益和信心水平
	expectedReturn := calculateExpectedReturn(recentData, zScore)
	confidence := calculateConfidence(zScore, signalScore)

	signal := &models.FactorTimingSignal{
		FactorName:     factorName,
		SignalDate:     signalDate,
		MASlope60:      maSlope,
		ZScore:         zScore,
		Percentile:     percentile,
		SignalStrength: signalStrength,
		SignalScore:    signalScore,
		ExpectedReturn: expectedReturn,
		Confidence:     confidence,
		CreatedAt:      time.Now(),
	}

	// 保存信号
	err = s.db.Create(signal).Error
	if err != nil {
		return nil, err
	}

	return signal, nil
}

// 辅助函数实现...
func calculateMASlope(data []models.FactorData) decimal.Decimal {
	// 实现移动平均斜率计算
	// 使用线性回归计算斜率
	return decimal.Zero
}

func calculateZScore(data []models.FactorData) decimal.Decimal {
	// 实现Z-score计算
	return decimal.Zero
}

func calculatePercentile(data []models.FactorData) decimal.Decimal {
	// 实现历史百分位计算
	return decimal.Zero
}

func determineSignalStrength(zScore decimal.Decimal) models.SignalStrength {
	// 根据Z-score确定信号强度
	absZScore := zScore.Abs()

	if absZScore.GreaterThanOrEqual(decimal.NewFromFloat(2.0)) {
		if zScore.GreaterThan(decimal.Zero) {
			return models.SignalStrengthStrongPositive
		}
		return models.SignalStrengthStrongNegative
	} else if absZScore.GreaterThanOrEqual(decimal.NewFromFloat(1.5)) {
		if zScore.GreaterThan(decimal.Zero) {
			return models.SignalStrengthWeakPositive
		}
		return models.SignalStrengthWeakNegative
	}

	return models.SignalStrengthNeutral
}

func getSignalScore(strength models.SignalStrength) int {
	switch strength {
	case models.SignalStrengthStrongPositive:
		return 2
	case models.SignalStrengthWeakPositive:
		return 1
	case models.SignalStrengthNeutral:
		return 0
	case models.SignalStrengthWeakNegative:
		return -1
	case models.SignalStrengthStrongNegative:
		return -2
	default:
		return 0
	}
}

func calculateExpectedReturn(data []models.FactorData, zScore decimal.Decimal) decimal.Decimal {
	// 实现预期收益计算
	return decimal.Zero
}

func calculateConfidence(zScore decimal.Decimal, signalScore int) decimal.Decimal {
	// 根据Z-score和信号分数计算信心水平
	absZScore := zScore.Abs()

	if absZScore.GreaterThanOrEqual(decimal.NewFromFloat(2.0)) {
		return decimal.NewFromFloat(0.80)
	} else if absZScore.GreaterThanOrEqual(decimal.NewFromFloat(1.5)) {
		return decimal.NewFromFloat(0.60)
	}

	return decimal.NewFromFloat(0.40)
}
```

---

## 4. 测试用例示例

### 4.1 因子数据模型测试

创建文件：`backend/models/factor_test.go`

```go
package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFactorData_TableName(t *testing.T) {
	factorData := FactorData{}
	assert.Equal(t, "factor_data", factorData.TableName())
}

func TestFactorTimingSignal_TableName(t *testing.T) {
	signal := FactorTimingSignal{}
	assert.Equal(t, "factor_timing_signals", signal.TableName())
}

func TestFactorData_CRUD(t *testing.T) {
	// 初始化测试数据库
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 创建因子数据
	factorData := FactorData{
		FactorName: "Mkt-RF",
		Date:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Value:      decimal.NewFromFloat(0.0052),
		DataSource: "Fama-French Library",
		CreatedAt:  time.Now(),
	}

	// 测试创建
	err := db.Create(&factorData).Error
	assert.NoError(t, err)
	assert.NotZero(t, factorData.ID)

	// 测试查询
	var retrieved FactorData
	err = db.First(&retrieved, factorData.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Mkt-RF", retrieved.FactorName)
	assert.True(t, retrieved.Value.Equal(decimal.NewFromFloat(0.0052)))

	// 测试更新
	retrieved.Value = decimal.NewFromFloat(0.0060)
	err = db.Save(&retrieved).Error
	assert.NoError(t, err)

	var updated FactorData
	err = db.First(&updated, factorData.ID).Error
	assert.NoError(t, err)
	assert.True(t, updated.Value.Equal(decimal.NewFromFloat(0.0060)))
}

func TestFactorTimingSignal_CRUD(t *testing.T) {
	// 初始化测试数据库
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 创建因子择时信号
	signal := FactorTimingSignal{
		FactorName:     "HML",
		SignalDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MASlope60:      decimal.NewFromFloat(0.0001),
		ZScore:         decimal.NewFromFloat(1.8),
		Percentile:     decimal.NewFromFloat(85.5),
		SignalStrength: SignalStrengthWeakPositive,
		SignalScore:    1,
		ExpectedReturn: decimal.NewFromFloat(0.0025),
		Confidence:     decimal.NewFromFloat(0.60),
		CreatedAt:      time.Now(),
	}

	// 测试创建
	err := db.Create(&signal).Error
	assert.NoError(t, err)
	assert.NotZero(t, signal.ID)

	// 测试查询
	var retrieved FactorTimingSignal
	err = db.First(&retrieved, signal.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "HML", retrieved.FactorName)
	assert.Equal(t, SignalStrengthWeakPositive, retrieved.SignalStrength)
	assert.Equal(t, 1, retrieved.SignalScore)
}

func TestSignalStrength_Determine(t *testing.T) {
	tests := []struct {
		name     string
		zScore   decimal.Decimal
		expected SignalStrength
	}{
		{
			name:     "Strong Positive",
			zScore:   decimal.NewFromFloat(2.5),
			expected: SignalStrengthStrongPositive,
		},
		{
			name:     "Weak Positive",
			zScore:   decimal.NewFromFloat(1.6),
			expected: SignalStrengthWeakPositive,
		},
		{
			name:     "Neutral",
			zScore:   decimal.NewFromFloat(0.5),
			expected: SignalStrengthNeutral,
		},
		{
			name:     "Weak Negative",
			zScore:   decimal.NewFromFloat(-1.7),
			expected: SignalStrengthWeakNegative,
		},
		{
			name:     "Strong Negative",
			zScore:   decimal.NewFromFloat(-2.3),
			expected: SignalStrengthStrongNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineSignalStrength(tt.zScore)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

---

## 5. 执行检查清单

### 5.1 迁移前检查

- [ ] 备份现有数据库
- [ ] 验证数据库连接
- [ ] 确认GORM版本兼容性
- [ ] 检查磁盘空间充足
- [ ] 准备回滚脚本

### 5.2 迁移执行检查

- [ ] 执行迁移脚本
- [ ] 验证表结构创建成功
- [ ] 验证索引创建成功
- [ ] 验证外键约束有效
- [ ] 导入初始数据

### 5.3 迁移后验证

- [ ] 运行单元测试
- [ ] 运行集成测试
- [ ] 执行数据完整性检查
- [ ] 验证性能指标
- [ ] 更新API文档

---

**文档版本**: v1.0
**最后更新**: 2026-04-25
**维护者**: ETF-Insight Team
