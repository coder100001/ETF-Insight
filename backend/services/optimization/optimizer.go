package optimization

// MPTOptimizerInterface 均值-方差优化器接口
type MPTOptimizerInterface interface {
	Optimize(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint) (*PortfolioResult, error)
	OptimizeMaxSharpe(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint) (*PortfolioResult, error)
	OptimizeMinVolatility(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint) (*PortfolioResult, error)
	OptimizeForTargetReturn(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint, targetReturn float64) (*PortfolioResult, error)
	CalculateEfficientFrontier(returns map[string]float64, covMatrix map[string]map[string]float64, constraint *Constraint, numPoints int) ([]*EfficientFrontierPoint, error)
}
