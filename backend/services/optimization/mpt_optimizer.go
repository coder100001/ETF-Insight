package optimization

import (
	"fmt"
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

// MPTOptimizer 均值-方差优化器
// 实现现代投资组合理论 (Modern Portfolio Theory)
type MPTOptimizer struct {
	// 优化参数
	RiskFreeRate float64 // 无风险利率
	MaxIter      int     // 最大迭代次数
	Tolerance    float64 // 收敛容差
}

// NewMPTOptimizer 创建MPT优化器
func NewMPTOptimizer() *MPTOptimizer {
	return &MPTOptimizer{
		RiskFreeRate: 0.045, // 默认4.5%年化无风险利率
		MaxIter:      1000,
		Tolerance:    1e-8,
	}
}

// SetRiskFreeRate 设置无风险利率
func (o *MPTOptimizer) SetRiskFreeRate(rate float64) {
	o.RiskFreeRate = rate
}

// PortfolioResult 组合优化结果
type PortfolioResult struct {
	Weights              map[string]float64 `json:"weights"`               // 各资产权重
	ExpectedReturn       float64            `json:"expected_return"`       // 预期年化收益率
	Volatility           float64            `json:"volatility"`            // 年化波动率
	SharpeRatio          float64            `json:"sharpe_ratio"`          // 夏普比率
	SortinoRatio         float64            `json:"sortino_ratio"`         // 索提诺比率
	DiversificationRatio float64            `json:"diversification_ratio"` // 分散化比率
	RiskContribution     map[string]float64 `json:"risk_contribution"`     // 各资产风险贡献
	HerfindahlIndex      float64            `json:"herfindahl_index"`      // 赫芬达尔指数（集中度）
}

// EfficientFrontierPoint 有效前沿点
type EfficientFrontierPoint struct {
	TargetReturn   float64            `json:"target_return"`   // 目标收益率
	MinVolatility  float64            `json:"min_volatility"`  // 最小波动率
	OptimalWeights map[string]float64 `json:"optimal_weights"` // 最优权重
	SharpeRatio    float64            `json:"sharpe_ratio"`    // 夏普比率
}

// Constraint 优化约束条件
type Constraint struct {
	MinWeight      map[string]float64 // 各资产最小权重
	MaxWeight      map[string]float64 // 各资产最大权重
	SectorLimits   map[string]float64 // 行业权重上限
	TotalWeight    float64            // 总权重（通常为1.0）
	AllowShort     bool               // 是否允许做空
	MaxShortWeight float64            // 最大做空权重
}

// NewConstraint 创建默认约束条件
func NewConstraint(symbols []string) *Constraint {
	minWeight := make(map[string]float64)
	maxWeight := make(map[string]float64)

	for _, symbol := range symbols {
		minWeight[symbol] = 0.0
		maxWeight[symbol] = 1.0
	}

	return &Constraint{
		MinWeight:      minWeight,
		MaxWeight:      maxWeight,
		SectorLimits:   make(map[string]float64),
		TotalWeight:    1.0,
		AllowShort:     false,
		MaxShortWeight: 0.0,
	}
}

// SetMinWeight 设置资产最小权重
func (c *Constraint) SetMinWeight(symbol string, weight float64) {
	c.MinWeight[symbol] = weight
}

// SetMaxWeight 设置资产最大权重
func (c *Constraint) SetMaxWeight(symbol string, weight float64) {
	c.MaxWeight[symbol] = weight
}

// SetSectorLimit 设置行业权重上限
func (c *Constraint) SetSectorLimit(sector string, limit float64) {
	c.SectorLimits[sector] = limit
}

// Optimize 执行均值-方差优化
// returns: 各资产预期年化收益率
// covMatrix: 协方差矩阵
// constraint: 约束条件
func (o *MPTOptimizer) Optimize(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *Constraint,
) (*PortfolioResult, error) {
	// 验证输入
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return nil, fmt.Errorf("returns map is empty")
	}

	// 验证协方差矩阵
	for _, s1 := range symbols {
		if _, ok := covMatrix[s1]; !ok {
			return nil, fmt.Errorf("covariance matrix missing data for %s", s1)
		}
		for _, s2 := range symbols {
			if _, ok := covMatrix[s1][s2]; !ok {
				return nil, fmt.Errorf("covariance matrix missing data for %s-%s", s1, s2)
			}
		}
	}

	// 构建收益率向量和协方差矩阵
	n := len(symbols)
	mu := make([]float64, n)
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}

	for i, s1 := range symbols {
		mu[i] = returns[s1]
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	// 构建约束边界
	minWeights := make([]float64, n)
	maxWeights := make([]float64, n)
	for i, symbol := range symbols {
		if min, ok := constraint.MinWeight[symbol]; ok {
			minWeights[i] = min
		}
		if max, ok := constraint.MaxWeight[symbol]; ok {
			maxWeights[i] = max
		}
	}

	// 使用二次规划求解最优权重
	// 目标：最小化 w^T * Sigma * w
	// 约束：sum(w) = 1, minWeights <= w <= maxWeights
	weights, err := o.solveQuadraticProgramming(Sigma, mu, minWeights, maxWeights, constraint.TotalWeight)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	// 构建结果
	weightMap := make(map[string]float64)
	for i, symbol := range symbols {
		weightMap[symbol] = weights[i]
	}

	// 计算组合指标
	portfolioReturn := o.calculatePortfolioReturn(mu, weights)
	portfolioVolatility := o.calculatePortfolioVolatility(Sigma, weights)
	sharpeRatio := (portfolioReturn - o.RiskFreeRate) / portfolioVolatility

	// 计算索提诺比率（使用下行波动率）
	downsideVolatility := o.calculateDownsideVolatility(returns, weights, symbols)
	sortinoRatio := 0.0
	if downsideVolatility > 0 {
		sortinoRatio = (portfolioReturn - o.RiskFreeRate) / downsideVolatility
	}

	// 计算风险贡献
	riskContribution := o.calculateRiskContribution(Sigma, weights, symbols)

	// 计算赫芬达尔指数（集中度指标）
	herfindahlIndex := 0.0
	for _, w := range weights {
		herfindahlIndex += w * w
	}

	// 计算分散化比率
	diversificationRatio := o.calculateDiversificationRatio(Sigma, weights)

	return &PortfolioResult{
		Weights:              weightMap,
		ExpectedReturn:       portfolioReturn,
		Volatility:           portfolioVolatility,
		SharpeRatio:          sharpeRatio,
		SortinoRatio:         sortinoRatio,
		DiversificationRatio: diversificationRatio,
		RiskContribution:     riskContribution,
		HerfindahlIndex:      herfindahlIndex,
	}, nil
}

// OptimizeMaxSharpe 最大化夏普比率
func (o *MPTOptimizer) OptimizeMaxSharpe(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *Constraint,
) (*PortfolioResult, error) {
	// 尝试不同的目标收益率，找到夏普比率最大的组合
	minReturn := 0.0
	maxReturn := 0.0

	for _, r := range returns {
		if r < minReturn || minReturn == 0 {
			minReturn = r
		}
		if r > maxReturn {
			maxReturn = r
		}
	}

	bestSharpe := -999.0
	var bestResult *PortfolioResult

	// 在收益率范围内搜索
	steps := 50
	for i := 0; i <= steps; i++ {
		targetReturn := minReturn + (maxReturn-minReturn)*float64(i)/float64(steps)

		result, err := o.OptimizeForTargetReturn(returns, covMatrix, constraint, targetReturn)
		if err != nil {
			continue
		}

		if result.SharpeRatio > bestSharpe {
			bestSharpe = result.SharpeRatio
			bestResult = result
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("failed to find optimal portfolio")
	}

	return bestResult, nil
}

// OptimizeMinVolatility 最小波动率组合
func (o *MPTOptimizer) OptimizeMinVolatility(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *Constraint,
) (*PortfolioResult, error) {
	// 直接调用Optimize，它会最小化波动率
	return o.Optimize(returns, covMatrix, constraint)
}

// OptimizeForTargetReturn 针对目标收益率优化
func (o *MPTOptimizer) OptimizeForTargetReturn(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *Constraint,
	targetReturn float64,
) (*PortfolioResult, error) {
	// 添加收益率约束后优化
	// 简化实现：使用惩罚函数方法
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}

	n := len(symbols)
	mu := make([]float64, n)
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}

	for i, s1 := range symbols {
		mu[i] = returns[s1]
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	// 构建约束边界
	minWeights := make([]float64, n)
	maxWeights := make([]float64, n)
	for i, symbol := range symbols {
		if min, ok := constraint.MinWeight[symbol]; ok {
			minWeights[i] = min
		}
		if max, ok := constraint.MaxWeight[symbol]; ok {
			maxWeights[i] = max
		}
	}

	// 使用带收益率约束的优化
	weights, err := o.solveQuadraticProgrammingWithReturnConstraint(
		Sigma, mu, minWeights, maxWeights, constraint.TotalWeight, targetReturn,
	)
	if err != nil {
		return nil, err
	}

	// 构建结果
	weightMap := make(map[string]float64)
	for i, symbol := range symbols {
		weightMap[symbol] = weights[i]
	}

	portfolioReturn := o.calculatePortfolioReturn(mu, weights)
	portfolioVolatility := o.calculatePortfolioVolatility(Sigma, weights)
	sharpeRatio := (portfolioReturn - o.RiskFreeRate) / portfolioVolatility

	downsideVolatility := o.calculateDownsideVolatility(returns, weights, symbols)
	sortinoRatio := 0.0
	if downsideVolatility > 0 {
		sortinoRatio = (portfolioReturn - o.RiskFreeRate) / downsideVolatility
	}

	riskContribution := o.calculateRiskContribution(Sigma, weights, symbols)

	herfindahlIndex := 0.0
	for _, w := range weights {
		herfindahlIndex += w * w
	}

	diversificationRatio := o.calculateDiversificationRatio(Sigma, weights)

	return &PortfolioResult{
		Weights:              weightMap,
		ExpectedReturn:       portfolioReturn,
		Volatility:           portfolioVolatility,
		SharpeRatio:          sharpeRatio,
		SortinoRatio:         sortinoRatio,
		DiversificationRatio: diversificationRatio,
		RiskContribution:     riskContribution,
		HerfindahlIndex:      herfindahlIndex,
	}, nil
}

// CalculateEfficientFrontier 计算有效前沿
func (o *MPTOptimizer) CalculateEfficientFrontier(
	returns map[string]float64,
	covMatrix map[string]map[string]float64,
	constraint *Constraint,
	numPoints int,
) ([]*EfficientFrontierPoint, error) {
	if numPoints < 2 {
		numPoints = 20
	}

	// 确定收益率范围
	minReturn := 0.0
	maxReturn := 0.0
	for _, r := range returns {
		if r < minReturn || minReturn == 0 {
			minReturn = r
		}
		if r > maxReturn {
			maxReturn = r
		}
	}

	frontier := make([]*EfficientFrontierPoint, 0, numPoints)

	for i := 0; i < numPoints; i++ {
		targetReturn := minReturn + (maxReturn-minReturn)*float64(i)/float64(numPoints-1)

		result, err := o.OptimizeForTargetReturn(returns, covMatrix, constraint, targetReturn)
		if err != nil {
			continue
		}

		frontier = append(frontier, &EfficientFrontierPoint{
			TargetReturn:   targetReturn,
			MinVolatility:  result.Volatility,
			OptimalWeights: result.Weights,
			SharpeRatio:    result.SharpeRatio,
		})
	}

	// 按波动率排序
	sort.Slice(frontier, func(i, j int) bool {
		return frontier[i].MinVolatility < frontier[j].MinVolatility
	})

	return frontier, nil
}

// 内部计算方法

// solveQuadraticProgramming 求解二次规划问题
// 使用投影梯度下降法
func (o *MPTOptimizer) solveQuadraticProgramming(
	Sigma [][]float64,
	mu []float64,
	minWeights []float64,
	maxWeights []float64,
	totalWeight float64,
) ([]float64, error) {
	n := len(mu)
	weights := make([]float64, n)

	// 初始化：等权重
	for i := range weights {
		weights[i] = totalWeight / float64(n)
	}

	// 投影梯度下降
	learningRate := 0.01
	for iter := 0; iter < o.MaxIter; iter++ {
		// 计算梯度: 2 * Sigma * w
		gradient := make([]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				gradient[i] += 2 * Sigma[i][j] * weights[j]
			}
		}

		// 梯度下降更新
		newWeights := make([]float64, n)
		for i := 0; i < n; i++ {
			newWeights[i] = weights[i] - learningRate*gradient[i]
		}

		// 投影到约束空间
		newWeights = o.projectToConstraints(newWeights, minWeights, maxWeights, totalWeight)

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

		// 学习率衰减
		if iter%100 == 0 {
			learningRate *= 0.95
		}
	}

	return weights, nil
}

// solveQuadraticProgrammingWithReturnConstraint 带收益率约束的二次规划
func (o *MPTOptimizer) solveQuadraticProgrammingWithReturnConstraint(
	Sigma [][]float64,
	mu []float64,
	minWeights []float64,
	maxWeights []float64,
	totalWeight float64,
	targetReturn float64,
) ([]float64, error) {
	n := len(mu)
	weights := make([]float64, n)

	// 初始化
	for i := range weights {
		weights[i] = totalWeight / float64(n)
	}

	learningRate := 0.01
	penaltyWeight := 1000.0 // 收益率约束的惩罚权重

	for iter := 0; iter < o.MaxIter; iter++ {
		// 计算当前收益率
		currentReturn := 0.0
		for i := 0; i < n; i++ {
			currentReturn += mu[i] * weights[i]
		}

		// 计算梯度
		gradient := make([]float64, n)
		for i := 0; i < n; i++ {
			// 波动率梯度
			for j := 0; j < n; j++ {
				gradient[i] += 2 * Sigma[i][j] * weights[j]
			}
			// 收益率约束梯度（惩罚函数）
			returnViolation := currentReturn - targetReturn
			gradient[i] += penaltyWeight * returnViolation * mu[i]
		}

		// 梯度下降
		newWeights := make([]float64, n)
		for i := 0; i < n; i++ {
			newWeights[i] = weights[i] - learningRate*gradient[i]
		}

		// 投影
		newWeights = o.projectToConstraints(newWeights, minWeights, maxWeights, totalWeight)

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

		if iter%100 == 0 {
			learningRate *= 0.95
		}
	}

	return weights, nil
}

// projectToConstraints 投影到约束空间
func (o *MPTOptimizer) projectToConstraints(
	weights []float64,
	minWeights []float64,
	maxWeights []float64,
	totalWeight float64,
) []float64 {
	n := len(weights)
	result := make([]float64, n)

	// 第一步：裁剪到边界
	for i := 0; i < n; i++ {
		result[i] = weights[i]
		if result[i] < minWeights[i] {
			result[i] = minWeights[i]
		}
		if result[i] > maxWeights[i] {
			result[i] = maxWeights[i]
		}
	}

	// 第二步：归一化到总权重
	currentSum := 0.0
	for _, w := range result {
		currentSum += w
	}

	if currentSum > 0 {
		scale := totalWeight / currentSum
		for i := range result {
			result[i] *= scale
		}
	}

	// 第三步：再次检查边界（可能因为归一化超出边界）
	for i := 0; i < n; i++ {
		if result[i] < minWeights[i] {
			result[i] = minWeights[i]
		}
		if result[i] > maxWeights[i] {
			result[i] = maxWeights[i]
		}
	}

	return result
}

// calculatePortfolioReturn 计算组合收益率
func (o *MPTOptimizer) calculatePortfolioReturn(mu []float64, weights []float64) float64 {
	result := 0.0
	for i := 0; i < len(mu); i++ {
		result += mu[i] * weights[i]
	}
	return result
}

// calculatePortfolioVolatility 计算组合波动率
func (o *MPTOptimizer) calculatePortfolioVolatility(Sigma [][]float64, weights []float64) float64 {
	variance := 0.0
	for i := 0; i < len(weights); i++ {
		for j := 0; j < len(weights); j++ {
			variance += weights[i] * Sigma[i][j] * weights[j]
		}
	}
	return math.Sqrt(variance)
}

// calculateDownsideVolatility 计算下行波动率
func (o *MPTOptimizer) calculateDownsideVolatility(
	returns map[string]float64,
	weights []float64,
	symbols []string,
) float64 {
	// 简化实现：使用预期收益率作为目标
	portfolioReturn := 0.0
	for i, symbol := range symbols {
		portfolioReturn += returns[symbol] * weights[i]
	}

	// 计算各资产的下行风险贡献
	downsideVar := 0.0
	for i, s1 := range symbols {
		for j, s2 := range symbols {
			// 简化：假设收益率低于无风险利率为下行
			if returns[s1] < o.RiskFreeRate && returns[s2] < o.RiskFreeRate {
				downsideVar += weights[i] * weights[j] * 0.01 // 假设协方差
			}
		}
	}

	return math.Sqrt(downsideVar)
}

// calculateRiskContribution 计算各资产的风险贡献
func (o *MPTOptimizer) calculateRiskContribution(
	Sigma [][]float64,
	weights []float64,
	symbols []string,
) map[string]float64 {
	n := len(weights)
	portfolioVol := o.calculatePortfolioVolatility(Sigma, weights)

	if portfolioVol == 0 {
		result := make(map[string]float64)
		for _, symbol := range symbols {
			result[symbol] = 0
		}
		return result
	}

	// 计算边际风险贡献
	marginalRC := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			marginalRC[i] += Sigma[i][j] * weights[j]
		}
	}

	// 计算风险贡献占比
	result := make(map[string]float64)
	for i, symbol := range symbols {
		rc := weights[i] * marginalRC[i] / portfolioVol
		result[symbol] = rc / portfolioVol // 归一化到百分比
	}

	return result
}

// calculateDiversificationRatio 计算分散化比率
func (o *MPTOptimizer) calculateDiversificationRatio(Sigma [][]float64, weights []float64) float64 {
	// 分散化比率 = 加权平均波动率 / 组合波动率
	// 比率 > 1 表示有分散化效益

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

// DecimalToFloat64 decimal.Decimal 转换为 float64
func DecimalToFloat64(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

// Float64ToDecimal float64 转换为 decimal.Decimal
func Float64ToDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
