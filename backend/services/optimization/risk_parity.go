package optimization

import (
	"fmt"
	"math"
)

// RiskParityOptimizer 风险平价优化器
// 实现基于风险贡献的资产配置，使各资产对组合风险的贡献相等
type RiskParityOptimizer struct {
	MaxIter   int
	Tolerance float64
}

// NewRiskParityOptimizer 创建风险平价优化器
func NewRiskParityOptimizer() *RiskParityOptimizer {
	return &RiskParityOptimizer{
		MaxIter:   1000,
		Tolerance: 1e-8,
	}
}

// RiskParityResult 风险平价优化结果
type RiskParityResult struct {
	Weights              map[string]float64 `json:"weights"`               // 各资产权重
	RiskContributions    map[string]float64 `json:"risk_contributions"`    // 各资产风险贡献
	ExpectedReturn       float64            `json:"expected_return"`       // 预期年化收益率
	Volatility           float64            `json:"volatility"`            // 年化波动率
	Leverage             float64            `json:"leverage"`              // 杠杆倍数（如使用杠杆）
	TargetVolatility     float64            `json:"target_volatility"`     // 目标波动率
	DiversificationRatio float64            `json:"diversification_ratio"` // 分散化比率
}

// RiskParityConstraint 风险平价约束条件
type RiskParityConstraint struct {
	MinWeight        map[string]float64
	MaxWeight        map[string]float64
	TargetVolatility float64 // 目标波动率（可选）
	UseLeverage      bool    // 是否允许使用杠杆
	MaxLeverage      float64 // 最大杠杆倍数
}

// NewRiskParityConstraint 创建风险平价约束
func NewRiskParityConstraint(symbols []string) *RiskParityConstraint {
	minWeight := make(map[string]float64)
	maxWeight := make(map[string]float64)

	for _, symbol := range symbols {
		minWeight[symbol] = 0.0
		maxWeight[symbol] = 1.0
	}

	return &RiskParityConstraint{
		MinWeight:        minWeight,
		MaxWeight:        maxWeight,
		TargetVolatility: 0,
		UseLeverage:      false,
		MaxLeverage:      2.0,
	}
}

// Optimize 执行风险平价优化
// 目标：使各资产对组合风险的贡献相等
func (o *RiskParityOptimizer) Optimize(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *RiskParityConstraint,
) (*RiskParityResult, error) {
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return nil, fmt.Errorf("returns map is empty")
	}

	n := len(symbols)

	// 构建协方差矩阵
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}

	for i, s1 := range symbols {
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	// 使用牛顿法求解风险平价权重
	weights, err := o.solveRiskParity(Sigma, symbols, constraint)
	if err != nil {
		return nil, fmt.Errorf("risk parity optimization failed: %w", err)
	}

	// 如果设置了目标波动率，调整杠杆
	leverage := 1.0
	if constraint.TargetVolatility > 0 && constraint.UseLeverage {
		currentVol := o.calculatePortfolioVolatility(Sigma, weights)
		if currentVol > 0 {
			leverage = constraint.TargetVolatility / currentVol
			if leverage > constraint.MaxLeverage {
				leverage = constraint.MaxLeverage
			}
			for i := range weights {
				weights[i] *= leverage
			}
		}
	}

	// 构建结果
	weightMap := make(map[string]float64)
	for i, symbol := range symbols {
		weightMap[symbol] = weights[i]
	}

	// 计算风险贡献
	riskContributions := o.calculateRiskContributions(Sigma, weights, symbols)

	// 计算组合指标
	portfolioReturn := 0.0
	for i, symbol := range symbols {
		portfolioReturn += returns[symbol] * weights[i]
	}
	portfolioVolatility := o.calculatePortfolioVolatility(Sigma, weights)

	// 计算分散化比率
	diversificationRatio := o.calculateDiversificationRatio(Sigma, weights)

	return &RiskParityResult{
		Weights:              weightMap,
		RiskContributions:    riskContributions,
		ExpectedReturn:       portfolioReturn,
		Volatility:           portfolioVolatility,
		Leverage:             leverage,
		TargetVolatility:     constraint.TargetVolatility,
		DiversificationRatio: diversificationRatio,
	}, nil
}

// OptimizeERC 等风险贡献优化（Equal Risk Contribution）
// 这是风险平价的另一种表述方式
func (o *RiskParityOptimizer) OptimizeERC(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *RiskParityConstraint,
) (*RiskParityResult, error) {
	return o.Optimize(returns, covMatrix, constraint)
}

// OptimizeInverseVol 逆波动率加权
// 简化版风险平价，仅使用对角线元素（各资产波动率）
func (o *RiskParityOptimizer) OptimizeInverseVol(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *RiskParityConstraint,
) (*RiskParityResult, error) {
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}

	n := len(symbols)

	// 计算逆波动率权重
	invVols := make([]float64, n)
	for i, s1 := range symbols {
		vol := math.Sqrt(covMatrix[s1][s1])
		if vol > 0 {
			invVols[i] = 1.0 / vol
		}
	}

	// 归一化
	sum := 0.0
	for _, v := range invVols {
		sum += v
	}

	weights := make([]float64, n)
	for i := range weights {
		weights[i] = invVols[i] / sum
	}

	// 应用约束
	weights = o.applyConstraints(weights, symbols, constraint)

	// 构建协方差矩阵用于计算指标
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}
	for i, s1 := range symbols {
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	// 构建结果
	weightMap := make(map[string]float64)
	for i, symbol := range symbols {
		weightMap[symbol] = weights[i]
	}

	riskContributions := o.calculateRiskContributions(Sigma, weights, symbols)

	portfolioReturn := 0.0
	for i, symbol := range symbols {
		portfolioReturn += returns[symbol] * weights[i]
	}
	portfolioVolatility := o.calculatePortfolioVolatility(Sigma, weights)

	return &RiskParityResult{
		Weights:              weightMap,
		RiskContributions:    riskContributions,
		ExpectedReturn:       portfolioReturn,
		Volatility:           portfolioVolatility,
		Leverage:             1.0,
		TargetVolatility:     0,
		DiversificationRatio: o.calculateDiversificationRatio(Sigma, weights),
	}, nil
}

// solveRiskParity 求解风险平价问题
// 使用牛顿法迭代求解
func (o *RiskParityOptimizer) solveRiskParity(
	Sigma [][]float64,
	symbols []string,
	constraint *RiskParityConstraint,
) ([]float64, error) {
	n := len(symbols)

	// 初始化：等权重
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0 / float64(n)
	}

	// 迭代求解
	for iter := 0; iter < o.MaxIter; iter++ {
		// 计算风险贡献
		portfolioVol := o.calculatePortfolioVolatility(Sigma, weights)
		if portfolioVol == 0 {
			return weights, nil
		}

		marginalRC := make([]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				marginalRC[i] += Sigma[i][j] * weights[j]
			}
		}

		// 目标：每个资产的风险贡献相等 = portfolioVol^2 / n
		targetRC := portfolioVol * portfolioVol / float64(n)

		// 计算当前风险贡献
		currentRC := make([]float64, n)
		for i := 0; i < n; i++ {
			currentRC[i] = weights[i] * marginalRC[i]
		}

		// 计算梯度（风险贡献与目标的差异）
		gradient := make([]float64, n)
		for i := 0; i < n; i++ {
			gradient[i] = currentRC[i] - targetRC
		}

		// 梯度下降更新
		learningRate := 0.1
		newWeights := make([]float64, n)
		for i := 0; i < n; i++ {
			newWeights[i] = weights[i] - learningRate*gradient[i]/portfolioVol
		}

		// 应用约束
		newWeights = o.applyConstraints(newWeights, symbols, constraint)

		// 检查收敛
		diff := 0.0
		for i := 0; i < n; i++ {
			d := newWeights[i] - weights[i]
			diff += d * d
		}

		weights = newWeights

		if math.Sqrt(diff) < o.Tolerance {
			break
		}
	}

	return weights, nil
}

// applyConstraints 应用约束条件
func (o *RiskParityOptimizer) applyConstraints(
	weights []float64,
	symbols []string,
	constraint *RiskParityConstraint,
) []float64 {
	n := len(weights)
	result := make([]float64, n)

	// 应用权重边界
	for i, symbol := range symbols {
		result[i] = weights[i]

		if min, ok := constraint.MinWeight[symbol]; ok && result[i] < min {
			result[i] = min
		}
		if max, ok := constraint.MaxWeight[symbol]; ok && result[i] > max {
			result[i] = max
		}
	}

	// 归一化到总和为1
	sum := 0.0
	for _, w := range result {
		sum += w
	}

	if sum > 0 {
		for i := range result {
			result[i] /= sum
		}
	}

	return result
}

// calculateRiskContributions 计算各资产的风险贡献
func (o *RiskParityOptimizer) calculateRiskContributions(
	Sigma [][]float64,
	weights []float64,
	symbols []string,
) map[string]float64 {
	n := len(weights)
	portfolioVol := o.calculatePortfolioVolatility(Sigma, weights)

	if portfolioVol == 0 {
		result := make(map[string]float64)
		for _, symbol := range symbols {
			result[symbol] = 1.0 / float64(n)
		}
		return result
	}

	marginalRC := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			marginalRC[i] += Sigma[i][j] * weights[j]
		}
	}

	result := make(map[string]float64)
	totalRC := 0.0
	for i, symbol := range symbols {
		rc := weights[i] * marginalRC[i]
		result[symbol] = rc
		totalRC += rc
	}

	// 归一化为百分比
	if totalRC > 0 {
		for symbol := range result {
			result[symbol] /= totalRC
		}
	}

	return result
}

// calculatePortfolioVolatility 计算组合波动率
func (o *RiskParityOptimizer) calculatePortfolioVolatility(Sigma [][]float64, weights []float64) float64 {
	variance := 0.0
	for i := 0; i < len(weights); i++ {
		for j := 0; j < len(weights); j++ {
			variance += weights[i] * Sigma[i][j] * weights[j]
		}
	}
	return math.Sqrt(variance)
}

// calculateDiversificationRatio 计算分散化比率
func (o *RiskParityOptimizer) calculateDiversificationRatio(Sigma [][]float64, weights []float64) float64 {
	weightedAvgVol := 0.0
	for i := 0; i < len(weights); i++ {
		individualVol := math.Sqrt(Sigma[i][i])
		weightedAvgVol += weights[i] * individualVol
	}

	portfolioVol := o.calculatePortfolioVolatility(Sigma, weights)

	if portfolioVol == 0 {
		return 1.0
	}

	return weightedAvgVol / portfolioVol
}

// CalculateRiskBudget 计算风险预算
// 允许设置不同的风险预算比例（非等风险贡献）
func (o *RiskParityOptimizer) CalculateRiskBudget(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	riskBudget map[string]float64,
	constraint *RiskParityConstraint,
) (*RiskParityResult, error) {
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}

	n := len(symbols)

	// 验证风险预算
	totalBudget := 0.0
	for _, budget := range riskBudget {
		totalBudget += budget
	}

	if math.Abs(totalBudget-1.0) > 0.01 {
		// 归一化风险预算
		for symbol := range riskBudget {
			riskBudget[symbol] /= totalBudget
		}
	}

	// 构建协方差矩阵
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}

	for i, s1 := range symbols {
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	// 使用带风险预算的优化
	weights, err := o.solveRiskBudget(Sigma, symbols, riskBudget, constraint)
	if err != nil {
		return nil, err
	}

	// 构建结果
	weightMap := make(map[string]float64)
	for i, symbol := range symbols {
		weightMap[symbol] = weights[i]
	}

	riskContributions := o.calculateRiskContributions(Sigma, weights, symbols)

	portfolioReturn := 0.0
	for i, symbol := range symbols {
		portfolioReturn += returns[symbol] * weights[i]
	}
	portfolioVolatility := o.calculatePortfolioVolatility(Sigma, weights)

	return &RiskParityResult{
		Weights:              weightMap,
		RiskContributions:    riskContributions,
		ExpectedReturn:       portfolioReturn,
		Volatility:           portfolioVolatility,
		Leverage:             1.0,
		TargetVolatility:     0,
		DiversificationRatio: o.calculateDiversificationRatio(Sigma, weights),
	}, nil
}

// solveRiskBudget 求解风险预算问题
func (o *RiskParityOptimizer) solveRiskBudget(
	Sigma [][]float64,
	symbols []string,
	riskBudget map[string]float64,
	constraint *RiskParityConstraint,
) ([]float64, error) {
	n := len(symbols)

	// 初始化
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0 / float64(n)
	}

	for iter := 0; iter < o.MaxIter; iter++ {
		portfolioVol := o.calculatePortfolioVolatility(Sigma, weights)
		if portfolioVol == 0 {
			return weights, nil
		}

		marginalRC := make([]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				marginalRC[i] += Sigma[i][j] * weights[j]
			}
		}

		// 计算当前风险贡献
		currentRC := make([]float64, n)
		for i := 0; i < n; i++ {
			currentRC[i] = weights[i] * marginalRC[i]
		}

		// 计算梯度（相对于风险预算的差异）
		gradient := make([]float64, n)
		for i, symbol := range symbols {
			targetBudget := riskBudget[symbol]
			if targetBudget == 0 {
				targetBudget = 1.0 / float64(n)
			}
			targetRC := portfolioVol * portfolioVol * targetBudget
			gradient[i] = currentRC[i] - targetRC
		}

		// 梯度下降
		learningRate := 0.1
		newWeights := make([]float64, n)
		for i := 0; i < n; i++ {
			newWeights[i] = weights[i] - learningRate*gradient[i]/portfolioVol
		}

		// 应用约束
		newWeights = o.applyConstraints(newWeights, symbols, constraint)

		// 检查收敛
		diff := 0.0
		for i := 0; i < n; i++ {
			d := newWeights[i] - weights[i]
			diff += d * d
		}

		weights = newWeights

		if math.Sqrt(diff) < o.Tolerance {
			break
		}
	}

	return weights, nil
}
