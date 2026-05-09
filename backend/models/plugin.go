package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PluginType string

const (
	PluginTypeAlphaGenerator     PluginType = "AlphaGenerator"
	PluginTypePortfolioOptimizer PluginType = "PortfolioOptimizer"
	PluginTypeRiskModel          PluginType = "RiskModel"
)

type PluginStatus string

const (
	PluginStatusActive     PluginStatus = "active"
	PluginStatusDeprecated PluginStatus = "deprecated"
	PluginStatusDisabled   PluginStatus = "disabled"
)

type PluginRegistry struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	PluginName string     `json:"plugin_name" gorm:"uniqueIndex;size:100"`
	PluginType PluginType `json:"plugin_type" gorm:"size:50;index"`
	Version    string     `json:"version" gorm:"size:20"`

	InputSchema  JSONMap `json:"input_schema" gorm:"type:json"`
	OutputSchema JSONMap `json:"output_schema" gorm:"type:json"`

	Dependencies JSONMap `json:"dependencies" gorm:"type:json"`

	Description   string `json:"description" gorm:"size:500"`
	Author        string `json:"author" gorm:"size:100"`
	Documentation string `json:"documentation" gorm:"size:200"`

	Status       PluginStatus `json:"status" gorm:"size:20;default:'active'"`
	RegisteredAt time.Time    `json:"registered_at"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

func (PluginRegistry) TableName() string {
	return "plugin_registries"
}

type PluginConfiguration struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	PluginID    uint `json:"plugin_id" gorm:"index"`
	PortfolioID uint `json:"portfolio_id" gorm:"index"`

	ConfigName string  `json:"config_name" gorm:"size:100"`
	Parameters JSONMap `json:"parameters" gorm:"type:json"`

	IsActive  bool `json:"is_active" gorm:"default:true"`
	IsDefault bool `json:"is_default" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Plugin    PluginRegistry `json:"plugin" gorm:"foreignKey:PluginID"`
	Portfolio Portfolio      `json:"portfolio" gorm:"foreignKey:PortfolioID"`
}

func (PluginConfiguration) TableName() string {
	return "plugin_configurations"
}

type ExecutionStatus string

const (
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusTimeout ExecutionStatus = "timeout"
)

type PluginExecutionLog struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	PluginID uint `json:"plugin_id" gorm:"index"`
	ConfigID uint `json:"config_id" gorm:"index"`

	ExecutionID string    `json:"execution_id" gorm:"uniqueIndex;size:50"`
	StartTime   time.Time `json:"start_time" gorm:"index"`
	EndTime     time.Time `json:"end_time"`
	Duration    int       `json:"duration"`

	InputData  JSONMap `json:"input_data" gorm:"type:json"`
	OutputData JSONMap `json:"output_data" gorm:"type:json"`

	Status       ExecutionStatus `json:"status" gorm:"size:20"`
	ErrorMessage string          `json:"error_message" gorm:"type:text"`

	MemoryUsage int64   `json:"memory_usage"`
	CPUUsage    float64 `json:"cpu_usage"`

	CreatedAt time.Time `json:"created_at"`

	Plugin PluginRegistry      `json:"plugin" gorm:"foreignKey:PluginID"`
	Config PluginConfiguration `json:"config" gorm:"foreignKey:ConfigID"`
}

func (PluginExecutionLog) TableName() string {
	return "plugin_execution_logs"
}

type ModelBenchmarkMatrix struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	ComparisonName string `json:"comparison_name" gorm:"size:100;index"`

	AlphaPluginID     uint `json:"alpha_plugin_id" gorm:"index"`
	OptimizerPluginID uint `json:"optimizer_plugin_id" gorm:"index"`
	RiskPluginID      uint `json:"risk_plugin_id" gorm:"index"`

	BacktestWindow     int             `json:"backtest_window"`
	RebalanceFrequency string          `json:"rebalance_frequency" gorm:"size:20"`
	TransactionCost    decimal.Decimal `json:"transaction_cost" gorm:"type:decimal(5,4)"`

	TotalReturn         decimal.Decimal `json:"total_return" gorm:"type:decimal(10,6)"`
	AnnualReturn        decimal.Decimal `json:"annual_return" gorm:"type:decimal(10,6)"`
	Volatility          decimal.Decimal `json:"volatility" gorm:"type:decimal(10,6)"`
	SharpeRatio         decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,6)"`
	MaxDrawdown         decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,6)"`
	CalmarRatio         decimal.Decimal `json:"calmar_ratio" gorm:"type:decimal(10,6)"`
	Rolling1YWinRate    decimal.Decimal `json:"rolling_1y_win_rate" gorm:"type:decimal(5,2)"`
	TailDependencyIndex decimal.Decimal `json:"tail_dependency_index" gorm:"type:decimal(10,6)"`

	DetailedMetrics JSONMap `json:"detailed_metrics" gorm:"type:json"`

	ComparisonDate time.Time `json:"comparison_date"`
	CreatedAt      time.Time `json:"created_at"`

	AlphaPlugin     PluginRegistry `json:"alpha_plugin" gorm:"foreignKey:AlphaPluginID"`
	OptimizerPlugin PluginRegistry `json:"optimizer_plugin" gorm:"foreignKey:OptimizerPluginID"`
	RiskPlugin      PluginRegistry `json:"risk_plugin" gorm:"foreignKey:RiskPluginID"`
}

func (ModelBenchmarkMatrix) TableName() string {
	return "model_benchmark_matrices"
}

type ExperimentStatus string

const (
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusFailed    ExperimentStatus = "failed"
)

type StrategyExperiment struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	ExperimentName string `json:"experiment_name" gorm:"size:100;index"`
	Description    string `json:"description" gorm:"size:500"`

	BenchmarkMatrixID uint            `json:"benchmark_matrix_id" gorm:"index"`
	AllocationRatio   decimal.Decimal `json:"allocation_ratio" gorm:"type:decimal(5,2)"`

	ExperimentResult JSONMap `json:"experiment_result" gorm:"type:json"`
	IsSuccessful     bool    `json:"is_successful"`
	SuccessCriteria  JSONMap `json:"success_criteria" gorm:"type:json"`

	StartDate time.Time        `json:"start_date"`
	EndDate   time.Time        `json:"end_date"`
	Status    ExperimentStatus `json:"status" gorm:"size:20"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	BenchmarkMatrix ModelBenchmarkMatrix `json:"benchmark_matrix" gorm:"foreignKey:BenchmarkMatrixID"`
}

func (StrategyExperiment) TableName() string {
	return "strategy_experiments"
}

func (p PluginType) IsValid() bool {
	switch p {
	case PluginTypeAlphaGenerator, PluginTypePortfolioOptimizer, PluginTypeRiskModel:
		return true
	default:
		return false
	}
}

func (s PluginStatus) IsValid() bool {
	switch s {
	case PluginStatusActive, PluginStatusDeprecated, PluginStatusDisabled:
		return true
	default:
		return false
	}
}

func (s ExecutionStatus) IsValid() bool {
	switch s {
	case ExecutionStatusSuccess, ExecutionStatusFailed, ExecutionStatusTimeout:
		return true
	default:
		return false
	}
}

func (s ExperimentStatus) IsValid() bool {
	switch s {
	case ExperimentStatusRunning, ExperimentStatusCompleted, ExperimentStatusFailed:
		return true
	default:
		return false
	}
}
