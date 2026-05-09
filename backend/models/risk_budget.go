package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type RiskBudgetConfig struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	PortfolioID uint `json:"portfolio_id" gorm:"uniqueIndex"`

	StockCVaRBudget     decimal.Decimal `json:"stock_cvar_budget" gorm:"type:decimal(5,2)"`
	BondCVaRBudget      decimal.Decimal `json:"bond_cvar_budget" gorm:"type:decimal(5,2)"`
	CommodityCVaRBudget decimal.Decimal `json:"commodity_cvar_budget" gorm:"type:decimal(5,2)"`
	CashCVaRBudget      decimal.Decimal `json:"cash_cvar_budget" gorm:"type:decimal(5,2)"`

	UseVaRConstraint bool            `json:"use_var_constraint" gorm:"default:false"`
	StockVaRBudget   decimal.Decimal `json:"stock_var_budget" gorm:"type:decimal(5,2)"`
	BondVaRBudget    decimal.Decimal `json:"bond_var_budget" gorm:"type:decimal(5,2)"`

	MinSkewness decimal.Decimal `json:"min_skewness" gorm:"type:decimal(10,6)"`
	MaxDrawdown decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(5,2)"`

	CVaRConfidence decimal.Decimal `json:"cvar_confidence" gorm:"type:decimal(5,4);default:0.95"`
	VaRConfidence  decimal.Decimal `json:"var_confidence" gorm:"type:decimal(5,4);default:0.95"`

	IsActive      bool      `json:"is_active" gorm:"default:true"`
	EffectiveDate time.Time `json:"effective_date"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Portfolio Portfolio `json:"portfolio" gorm:"foreignKey:PortfolioID"`
}

func (RiskBudgetConfig) TableName() string {
	return "risk_budget_configs"
}

type MonteCarloSimulation struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	PortfolioID    uint      `json:"portfolio_id" gorm:"index:idx_simulation_portfolio"`
	SimulationDate time.Time `json:"simulation_date" gorm:"index:idx_simulation_portfolio"`

	NumPaths        int             `json:"num_paths"`
	TimeSteps       int             `json:"time_steps"`
	TimeHorizon     int             `json:"time_horizon"`
	ConfidenceLevel decimal.Decimal `json:"confidence_level" gorm:"type:decimal(5,4)"`

	VaR95  decimal.Decimal `json:"var_95" gorm:"type:decimal(10,6)"`
	VaR99  decimal.Decimal `json:"var_99" gorm:"type:decimal(10,6)"`
	CVaR95 decimal.Decimal `json:"cvar_95" gorm:"type:decimal(10,6)"`
	CVaR99 decimal.Decimal `json:"cvar_99" gorm:"type:decimal(10,6)"`

	MeanReturn decimal.Decimal `json:"mean_return" gorm:"type:decimal(10,6)"`
	StdDev     decimal.Decimal `json:"std_dev" gorm:"type:decimal(10,6)"`
	Skewness   decimal.Decimal `json:"skewness" gorm:"type:decimal(10,6)"`
	Kurtosis   decimal.Decimal `json:"kurtosis" gorm:"type:decimal(10,6)"`

	SimulationResult string    `json:"simulation_result" gorm:"type:json"`
	CacheExpiry      time.Time `json:"cache_expiry"`

	CreatedAt time.Time `json:"created_at"`

	Portfolio Portfolio `json:"portfolio" gorm:"foreignKey:PortfolioID"`
}

func (MonteCarloSimulation) TableName() string {
	return "monte_carlo_simulations"
}

type RiskContribution struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	SimulationID uint   `json:"simulation_id" gorm:"index"`
	AssetID      uint   `json:"asset_id" gorm:"index"`
	AssetSymbol  string `json:"asset_symbol" gorm:"size:20"`

	Weight           decimal.Decimal `json:"weight" gorm:"type:decimal(5,2)"`
	CVaRContribution decimal.Decimal `json:"cvar_contribution" gorm:"type:decimal(10,6)"`
	MarginalCVaR     decimal.Decimal `json:"marginal_cvar" gorm:"type:decimal(10,6)"`
	CVaRPercentage   decimal.Decimal `json:"cvar_percentage" gorm:"type:decimal(5,2)"`

	BudgetLimit     decimal.Decimal `json:"budget_limit" gorm:"type:decimal(5,2)"`
	BudgetUsage     decimal.Decimal `json:"budget_usage" gorm:"type:decimal(5,2)"`
	BudgetDeviation decimal.Decimal `json:"budget_deviation" gorm:"type:decimal(5,2)"`

	CalculationDate time.Time `json:"calculation_date"`
	CreatedAt       time.Time `json:"created_at"`

	Simulation MonteCarloSimulation `json:"simulation" gorm:"foreignKey:SimulationID"`
	Asset      Asset                `json:"asset" gorm:"foreignKey:AssetID"`
}

func (RiskContribution) TableName() string {
	return "risk_contributions"
}

type RiskBudgetExecution struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	PortfolioID   uint      `json:"portfolio_id" gorm:"index"`
	ExecutionDate time.Time `json:"execution_date" gorm:"index"`

	TotalCVaR      decimal.Decimal `json:"total_cvar" gorm:"type:decimal(10,6)"`
	StockCVaRUsage decimal.Decimal `json:"stock_cvar_usage" gorm:"type:decimal(5,2)"`
	BondCVaRUsage  decimal.Decimal `json:"bond_cvar_usage" gorm:"type:decimal(5,2)"`

	IsBudgetSatisfied bool   `json:"is_budget_satisfied"`
	ViolationDetails  JSONMap `json:"violation_details" gorm:"type:json"`

	OptimizationMethod string `json:"optimization_method" gorm:"size:50"`
	Iterations         int    `json:"iterations"`

	CreatedAt time.Time `json:"created_at"`

	Portfolio Portfolio `json:"portfolio" gorm:"foreignKey:PortfolioID"`
}

func (RiskBudgetExecution) TableName() string {
	return "risk_budget_executions"
}
