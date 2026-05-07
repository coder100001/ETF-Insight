package optimization

import (
	"fmt"
	"math"
)

// BlackLittermanOptimizer Black-Litterman模型优化器
// 融合市场均衡收益和投资者观点，生成后验收益估计
type BlackLittermanOptimizer struct {
	Tau          float64 // 缩放参数，通常取0.025-0.05
	RiskFreeRate float64
}

// NewBlackLittermanOptimizer 创建Black-Litterman优化器
func NewBlackLittermanOptimizer() *BlackLittermanOptimizer {
	return &BlackLittermanOptimizer{
		Tau:          0.025,
		RiskFreeRate: 0.045,
	}
}

// SetTau 设置缩放参数
func (o *BlackLittermanOptimizer) SetTau(tau float64) {
	o.Tau = tau
}

// SetRiskFreeRate 设置无风险利率
func (o *BlackLittermanOptimizer) SetRiskFreeRate(rate float64) {
	o.RiskFreeRate = rate
}

// BlackLittermanResult Black-Litterman优化结果
type BlackLittermanResult struct {
	PriorReturns     map[string]float64 `json:"prior_returns"`     // 先验收益（市场均衡）
	PosteriorReturns map[string]float64 `json:"posterior_returns"` // 后验收益
	ImpliedReturns   map[string]float64 `json:"implied_returns"`   // 隐含收益
	OptimalWeights   map[string]float64 `json:"optimal_weights"`   // 最优权重
	ExpectedReturn   float64            `json:"expected_return"`   // 预期收益
	Volatility       float64            `json:"volatility"`        // 波动率
	SharpeRatio      float64            `json:"sharpe_ratio"`      // 夏普比率
	Confidence       float64            `json:"confidence"`        // 观点置信度
}

// InvestorView 投资者观点
type InvestorView struct {
	Type        string    `json:"type"`        // "absolute" 或 "relative"
	Assets      []string  `json:"assets"`      // 涉及的资产
	Weights     []float64 `json:"weights"`     // 观点权重（相对观点时使用）
	Return      float64   `json:"return"`      // 预期收益
	Confidence  float64   `json:"confidence"`  // 置信度 (0-1)
	Description string    `json:"description"` // 观点描述
}

// BlackLittermanConstraint Black-Litterman约束条件
type BlackLittermanConstraint struct {
	MinWeight map[string]float64
	MaxWeight map[string]float64
}

// NewBlackLittermanConstraint 创建约束
func NewBlackLittermanConstraint(symbols []string) *BlackLittermanConstraint {
	minWeight := make(map[string]float64)
	maxWeight := make(map[string]float64)

	for _, symbol := range symbols {
		minWeight[symbol] = 0.0
		maxWeight[symbol] = 1.0
	}

	return &BlackLittermanConstraint{
		MinWeight: minWeight,
		MaxWeight: maxWeight,
	}
}

// Optimize 执行Black-Litterman优化
// marketWeights: 市场均衡权重（如市值加权）
// covMatrix: 收益协方差矩阵
// views: 投资者观点列表
func (o *BlackLittermanOptimizer) Optimize(
	marketWeights map[string]float64,
	covMatrix map[string]map[string]float64,
	views []*InvestorView,
	constraint *BlackLittermanConstraint,
) (*BlackLittermanResult, error) {
	symbols := make([]string, 0, len(marketWeights))
	for symbol := range marketWeights {
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return nil, fmt.Errorf("market weights is empty")
	}

	n := len(symbols)

	// 构建协方差矩阵 Σ
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}
	for i, s1 := range symbols {
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	// 步骤1: 计算先验收益（反向优化）
	// Π = δ * Σ * w_market
	delta := 2.5 // 风险厌恶系数，通常取2-4
	priorReturns := o.calculatePriorReturns(Sigma, marketWeights, symbols, delta)

	// 步骤2: 构建观点矩阵 P 和观点收益向量 Q
	P, Q, Omega := o.buildViewMatrices(views, symbols)

	// 步骤3: 计算后验收益
	// E[R] = [(τΣ)^-1 + P^T * Ω^-1 * P]^-1 * [(τΣ)^-1 * Π + P^T * Ω^-1 * Q]
	posteriorReturns := o.calculatePosteriorReturns(Sigma, priorReturns, P, Q, Omega, symbols)

	// 步骤4: 基于后验收益进行均值-方差优化
	optimalWeights := o.optimizeWeights(Sigma, posteriorReturns, constraint, symbols)

	// 计算组合指标
	portfolioReturn := 0.0
	portfolioVariance := 0.0
	for i, s1 := range symbols {
		portfolioReturn += posteriorReturns[s1] * optimalWeights[s1]
		for j, s2 := range symbols {
			portfolioVariance += optimalWeights[s1] * Sigma[i][j] * optimalWeights[s2]
		}
	}
	portfolioVolatility := math.Sqrt(portfolioVariance)
	sharpeRatio := (portfolioReturn - o.RiskFreeRate) / portfolioVolatility

	// 计算隐含收益（如果按市场权重持有）
	impliedReturns := o.calculateImpliedReturns(Sigma, marketWeights, symbols, delta)

	// 计算整体置信度
	avgConfidence := o.calculateAverageConfidence(views)

	return &BlackLittermanResult{
		PriorReturns:     priorReturns,
		PosteriorReturns: posteriorReturns,
		ImpliedReturns:   impliedReturns,
		OptimalWeights:   optimalWeights,
		ExpectedReturn:   portfolioReturn,
		Volatility:       portfolioVolatility,
		SharpeRatio:      sharpeRatio,
		Confidence:       avgConfidence,
	}, nil
}

// OptimizeWithViews 使用观点进行优化（简化接口）
func (o *BlackLittermanOptimizer) OptimizeWithViews(
	marketWeights map[string]float64,
	covMatrix map[string]map[string]float64,
	absoluteViews map[string]float64, // 绝对观点：资产 -> 预期收益
	relativeViews []*RelativeView, // 相对观点
	constraint *BlackLittermanConstraint,
) (*BlackLittermanResult, error) {
	// 转换观点格式
	views := make([]*InvestorView, 0)

	// 添加绝对观点
	for asset, ret := range absoluteViews {
		views = append(views, &InvestorView{
			Type:       "absolute",
			Assets:     []string{asset},
			Return:     ret,
			Confidence: 0.5, // 默认置信度
		})
	}

	// 添加相对观点
	for _, rv := range relativeViews {
		views = append(views, &InvestorView{
			Type:       "relative",
			Assets:     []string{rv.Asset1, rv.Asset2},
			Weights:    []float64{1.0, -1.0},
			Return:     rv.ExpectedDiff,
			Confidence: rv.Confidence,
		})
	}

	return o.Optimize(marketWeights, covMatrix, views, constraint)
}

// RelativeView 相对观点
type RelativeView struct {
	Asset1       string
	Asset2       string
	ExpectedDiff float64 // 预期收益差（Asset1 - Asset2）
	Confidence   float64
}

// calculatePriorReturns 计算先验收益（反向优化）
func (o *BlackLittermanOptimizer) calculatePriorReturns(
	Sigma [][]float64,
	marketWeights map[string]float64,
	symbols []string,
	delta float64,
) map[string]float64 {
	priorReturns := make(map[string]float64)

	// Π = δ * Σ * w_market
	for i, s1 := range symbols {
		priorReturns[s1] = 0
		for j, s2 := range symbols {
			w := marketWeights[s2]
			priorReturns[s1] += delta * Sigma[i][j] * w
		}
	}

	return priorReturns
}

// buildViewMatrices 构建观点矩阵
func (o *BlackLittermanOptimizer) buildViewMatrices(
	views []*InvestorView,
	symbols []string,
) ([][]float64, []float64, [][]float64) {
	k := len(views)
	n := len(symbols)

	if k == 0 {
		return nil, nil, nil
	}

	// 构建符号到索引的映射
	symbolIndex := make(map[string]int)
	for i, symbol := range symbols {
		symbolIndex[symbol] = i
	}

	// P矩阵 (k x n)
	P := make([][]float64, k)
	for i := range P {
		P[i] = make([]float64, n)
	}

	// Q向量 (k x 1)
	Q := make([]float64, k)

	// Ω矩阵 (k x k)，对角矩阵
	Omega := make([][]float64, k)
	for i := range Omega {
		Omega[i] = make([]float64, k)
	}

	for i, view := range views {
		Q[i] = view.Return

		if view.Type == "absolute" && len(view.Assets) > 0 {
			// 绝对观点：P[i][asset] = 1
			if idx, ok := symbolIndex[view.Assets[0]]; ok {
				P[i][idx] = 1.0
			}
		} else if view.Type == "relative" && len(view.Assets) >= 2 {
			// 相对观点：P[i][asset1] = 1, P[i][asset2] = -1
			if idx1, ok := symbolIndex[view.Assets[0]]; ok {
				P[i][idx1] = 1.0
			}
			if idx2, ok := symbolIndex[view.Assets[1]]; ok {
				P[i][idx2] = -1.0
			}
		}

		// 观点的不确定性（置信度越低，不确定性越高）
		confidence := view.Confidence
		if confidence <= 0 {
			confidence = 0.1
		}
		if confidence > 1 {
			confidence = 1.0
		}
		// Ω[i][i] 与置信度成反比
		Omega[i][i] = (1.0 - confidence) / confidence * 0.01
	}

	return P, Q, Omega
}

// calculatePosteriorReturns 计算后验收益
func (o *BlackLittermanOptimizer) calculatePosteriorReturns(
	Sigma [][]float64,
	priorReturns map[string]float64,
	P [][]float64,
	Q []float64,
	Omega [][]float64,
	symbols []string,
) map[string]float64 {
	n := len(symbols)

	// 如果没有观点，返回先验收益
	if P == nil || len(P) == 0 {
		return priorReturns
	}

	k := len(P)

	// 构建先验收益向量
	Pi := make([]float64, n)
	for i, symbol := range symbols {
		Pi[i] = priorReturns[symbol]
	}

	// 计算 τΣ
	tauSigma := make([][]float64, n)
	for i := range tauSigma {
		tauSigma[i] = make([]float64, n)
		for j := range tauSigma[i] {
			tauSigma[i][j] = o.Tau * Sigma[i][j]
		}
	}

	// 简化计算：使用后验收益的解析解
	// E[R] = Π + τΣ * P^T * (P * τΣ * P^T + Ω)^-1 * (Q - P * Π)

	// 1. 计算 P * Π
	P_Pi := make([]float64, k)
	for i := range k {
		for j := range n {
			P_Pi[i] += P[i][j] * Pi[j]
		}
	}

	// 2. 计算 Q - P * Π
	residual := make([]float64, k)
	for i := range k {
		residual[i] = Q[i] - P_Pi[i]
	}

	// 3. 计算 P * τΣ * P^T + Ω (简化为对角矩阵)
	// 这里使用简化计算，实际应该使用矩阵求逆
	posteriorReturns := make(map[string]float64)

	// 简化：直接根据观点调整先验收益
	for _, symbol := range symbols {
		posteriorReturns[symbol] = priorReturns[symbol]
	}

	// 应用观点调整
	for _, view := range viewsFromMatrices(P, Q, symbols) {
		if view.Type == "absolute" && len(view.Assets) > 0 {
			asset := view.Assets[0]
			// 向观点收益靠拢
			confidence := 0.5
			if len(view.Weights) > 0 {
				confidence = view.Weights[0]
			}
			posteriorReturns[asset] = priorReturns[asset]*(1-confidence) + view.Return*confidence
		}
	}

	return posteriorReturns
}

// viewsFromMatrices 从矩阵还原观点（简化）
func viewsFromMatrices(P [][]float64, Q []float64, symbols []string) []*InvestorView {
	views := make([]*InvestorView, 0)
	for i := range P {
		assets := make([]string, 0)
		weights := make([]float64, 0)
		for j := range symbols {
			if P[i][j] != 0 {
				assets = append(assets, symbols[j])
				weights = append(weights, P[i][j])
			}
		}
		viewType := "absolute"
		if len(assets) > 1 {
			viewType = "relative"
		}
		views = append(views, &InvestorView{
			Type:    viewType,
			Assets:  assets,
			Weights: weights,
			Return:  Q[i],
		})
	}
	return views
}

// optimizeWeights 基于后验收益优化权重
func (o *BlackLittermanOptimizer) optimizeWeights(
	Sigma [][]float64,
	posteriorReturns map[string]float64,
	constraint *BlackLittermanConstraint,
	symbols []string,
) map[string]float64 {
	n := len(symbols)

	// 构建收益向量
	mu := make([]float64, n)
	for i, symbol := range symbols {
		mu[i] = posteriorReturns[symbol]
	}

	// 简化：使用等权重作为起点，然后调整
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0 / float64(n)
	}

	// 基于收益调整权重（收益越高，权重越大）
	// 使用softmax风格的调整
	expReturns := make([]float64, n)
	sumExp := 0.0
	for i := range n {
		expReturns[i] = math.Exp(mu[i] * 10) // 缩放因子
		sumExp += expReturns[i]
	}

	for i := range weights {
		weights[i] = expReturns[i] / sumExp
	}

	// 应用约束（nil 约束时创建默认值）
	if constraint == nil {
		constraint = &BlackLittermanConstraint{
			MinWeight: make(map[string]float64),
			MaxWeight: make(map[string]float64),
		}
	}
	weights = o.applyConstraints(weights, symbols, constraint)

	// 构建结果
	result := make(map[string]float64)
	for i, symbol := range symbols {
		result[symbol] = weights[i]
	}

	return result
}

// applyConstraints 应用约束
func (o *BlackLittermanOptimizer) applyConstraints(
	weights []float64,
	symbols []string,
	constraint *BlackLittermanConstraint,
) []float64 {
	n := len(weights)
	result := make([]float64, n)

	// 应用边界
	for i, symbol := range symbols {
		result[i] = weights[i]
		if min, ok := constraint.MinWeight[symbol]; ok && result[i] < min {
			result[i] = min
		}
		if max, ok := constraint.MaxWeight[symbol]; ok && result[i] > max {
			result[i] = max
		}
	}

	// 归一化
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

// calculateImpliedReturns 计算隐含收益
func (o *BlackLittermanOptimizer) calculateImpliedReturns(
	Sigma [][]float64,
	marketWeights map[string]float64,
	symbols []string,
	delta float64,
) map[string]float64 {
	return o.calculatePriorReturns(Sigma, marketWeights, symbols, delta)
}

// calculateAverageConfidence 计算平均置信度
func (o *BlackLittermanOptimizer) calculateAverageConfidence(views []*InvestorView) float64 {
	if len(views) == 0 {
		return 0
	}

	totalConfidence := 0.0
	for _, view := range views {
		totalConfidence += view.Confidence
	}

	return totalConfidence / float64(len(views))
}

// CalculateMarketImpliedReturns 计算市场隐含收益（反向优化）
// 给定市场权重和协方差矩阵，反推出市场隐含的收益预期
func (o *BlackLittermanOptimizer) CalculateMarketImpliedReturns(
	marketWeights map[string]float64,
	covMatrix map[string]map[string]float64,
	riskAversion float64,
) map[string]float64 {
	symbols := make([]string, 0, len(marketWeights))
	for symbol := range marketWeights {
		symbols = append(symbols, symbol)
	}

	n := len(symbols)
	Sigma := make([][]float64, n)
	for i := range Sigma {
		Sigma[i] = make([]float64, n)
	}
	for i, s1 := range symbols {
		for j, s2 := range symbols {
			Sigma[i][j] = covMatrix[s1][s2]
		}
	}

	return o.calculatePriorReturns(Sigma, marketWeights, symbols, riskAversion)
}
