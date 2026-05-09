package optimization

// MPTOptimizerInterface 均值-方差优化器接口
type MPTOptimizerInterface interface {
	Optimize(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint) (*PortfolioResult, error)
	OptimizeMaxSharpe(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint) (*PortfolioResult, error)
	OptimizeMinVolatility(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint) (*PortfolioResult, error)
	OptimizeForTargetReturn(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint, targetReturn float64) (*PortfolioResult, error)
	CalculateEfficientFrontier(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint, numPoints int) ([]*EfficientFrontierPoint, error)
}

// RiskParityOptimizerInterface 风险平价优化器接口
type RiskParityOptimizerInterface interface {
	Optimize(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *RiskParityConstraint) (*RiskParityResult, error)
	OptimizeInverseVol(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *RiskParityConstraint) (*RiskParityResult, error)
	OptimizeERC(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *RiskParityConstraint) (*RiskParityResult, error)
	CalculateRiskBudget(returns map[string]float64, covMatrix map[string]map[string]float64, riskBudget map[string]float64, constraint *RiskParityConstraint) (*RiskParityResult, error)
}

// BlackLittermanOptimizerInterface Black-Litterman优化器接口
type BlackLittermanOptimizerInterface interface {
	Optimize(marketWeights map[string]float64, covMatrix map[string]map[string]float64, views []*InvestorView, constraint *BlackLittermanConstraint) (*BlackLittermanResult, error)
	OptimizeWithViews(marketWeights map[string]float64, covMatrix map[string]map[string]float64, absoluteViews map[string]float64, relativeViews []*RelativeView, constraint *BlackLittermanConstraint) (*BlackLittermanResult, error)
	CalculateMarketImpliedReturns(marketWeights map[string]float64, covMatrix map[string]map[string]float64, riskAversion float64) map[string]float64
}
