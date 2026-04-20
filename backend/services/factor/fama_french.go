package factor

import (
	"math"
	"sort"
)

// FamaFrenchModel Fama-French三因子/五因子模型
type FamaFrenchModel struct {
	// 因子数据
	MarketReturns   []float64 // 市场超额收益 (Rm - Rf)
	SMBReturns      []float64 // 市值因子 (Small - Big)
	HMLReturns      []float64 // 价值因子 (High - Low)
	RMWReturns      []float64 // 盈利因子 (Robust - Weak) - 五因子
	CMAReturns      []float64 // 投资因子 (Conservative - Aggressive) - 五因子
	RiskFreeReturns []float64 // 无风险利率

	// 模型参数
	UseFiveFactor bool // 是否使用五因子模型
}

// NewFamaFrenchModel 创建Fama-French模型
func NewFamaFrenchModel() *FamaFrenchModel {
	return &FamaFrenchModel{
		UseFiveFactor: false, // 默认使用三因子
	}
}

// SetFiveFactor 设置是否使用五因子模型
func (m *FamaFrenchModel) SetFiveFactor(useFiveFactor bool) {
	m.UseFiveFactor = useFiveFactor
}

// FactorExposure 因子暴露
type FactorExposure struct {
	Market        float64 `json:"market"`                  // 市场因子暴露 (Beta)
	Size          float64 `json:"size"`                    // 市值因子暴露 (SMB)
	Value         float64 `json:"value"`                   // 价值因子暴露 (HML)
	Profitability float64 `json:"profitability,omitempty"` // 盈利因子暴露 (RMW) - 五因子
	Investment    float64 `json:"investment,omitempty"`    // 投资因子暴露 (CMA) - 五因子
	Alpha         float64 `json:"alpha"`                   // 截距项 (超额收益)
	R2            float64 `json:"r2"`                      // 拟合优度
	AdjR2         float64 `json:"adj_r2"`                  // 调整R2
}

// FactorAttribution 因子归因结果
type FactorAttribution struct {
	Exposures         *FactorExposure    `json:"exposures"`          // 因子暴露
	Contributions     map[string]float64 `json:"contributions"`      // 各因子贡献
	TotalReturn       float64            `json:"total_return"`       // 总收益
	ExplainedReturn   float64            `json:"explained_return"`   // 因子解释的收益
	UnexplainedReturn float64            `json:"unexplained_return"` // 未解释收益 (Alpha)
	AnnualizedAlpha   float64            `json:"annualized_alpha"`   // 年化Alpha
	TStatistics       map[string]float64 `json:"t_statistics"`       // T统计量
	PValues           map[string]float64 `json:"p_values"`           // P值
}

// FactorReturn 因子收益统计
type FactorReturn struct {
	Name        string  `json:"name"`         // 因子名称
	Annualized  float64 `json:"annualized"`   // 年化收益
	Volatility  float64 `json:"volatility"`   // 年化波动率
	Sharpe      float64 `json:"sharpe"`       // 夏普比率
	MaxDrawdown float64 `json:"max_drawdown"` // 最大回撤
}

// RegressionResult 回归结果
type RegressionResult struct {
	Coefficients map[string]float64 `json:"coefficients"` // 回归系数
	StdErrors    map[string]float64 `json:"std_errors"`   // 标准误
	TStats       map[string]float64 `json:"t_stats"`      // T统计量
	PValues      map[string]float64 `json:"p_values"`     // P值
	R2           float64            `json:"r2"`           // R平方
	AdjR2        float64            `json:"adj_r2"`       // 调整R平方
	Residuals    []float64          `json:"residuals"`    // 残差
}

// AnalyzePortfolio 分析投资组合的因子暴露
func (m *FamaFrenchModel) AnalyzePortfolio(
	portfolioReturns []float64,
	weights map[string]float64,
) (*FactorAttribution, error) {
	// 执行多元线性回归
	// Rp - Rf = α + β1*(Rm - Rf) + β2*SMB + β3*HML + β4*RMW + β5*CMA + ε

	result, err := m.performRegression(portfolioReturns)
	if err != nil {
		return nil, err
	}

	// 构建因子暴露
	exposure := &FactorExposure{
		Market: result.Coefficients["market"],
		Size:   result.Coefficients["smb"],
		Value:  result.Coefficients["hml"],
		Alpha:  result.Coefficients["alpha"],
		R2:     result.R2,
		AdjR2:  result.AdjR2,
	}

	if m.UseFiveFactor {
		exposure.Profitability = result.Coefficients["rmw"]
		exposure.Investment = result.Coefficients["cma"]
	}

	// 计算各因子贡献
	contributions := m.calculateFactorContributions(exposure)

	// 计算收益归因
	totalReturn := 0.0
	for _, r := range portfolioReturns {
		totalReturn += r
	}
	avgReturn := totalReturn / float64(len(portfolioReturns))
	annualizedReturn := avgReturn * 12 // 假设月度数据

	explainedReturn := contributions["market"] + contributions["smb"] +
		contributions["hml"]
	if m.UseFiveFactor {
		explainedReturn += contributions["rmw"] + contributions["cma"]
	}

	return &FactorAttribution{
		Exposures:         exposure,
		Contributions:     contributions,
		TotalReturn:       annualizedReturn,
		ExplainedReturn:   explainedReturn,
		UnexplainedReturn: exposure.Alpha * 12,
		AnnualizedAlpha:   exposure.Alpha * 12,
		TStatistics:       result.TStats,
		PValues:           result.PValues,
	}, nil
}

// AnalyzeETF 分析单个ETF的因子暴露
func (m *FamaFrenchModel) AnalyzeETF(
	etfReturns []float64,
	symbol string,
) (*FactorAttribution, error) {
	return m.AnalyzePortfolio(etfReturns, map[string]float64{symbol: 1.0})
}

// performRegression 执行多元线性回归
func (m *FamaFrenchModel) performRegression(dependent []float64) (*RegressionResult, error) {
	n := len(dependent)
	if n == 0 {
		return nil, nil
	}

	// 构建自变量矩阵 X
	var k int // 变量数
	if m.UseFiveFactor {
		k = 6 // alpha + 5个因子
	} else {
		k = 4 // alpha + 3个因子
	}

	// 截断数据到相同长度
	minLen := n
	if len(m.MarketReturns) < minLen {
		minLen = len(m.MarketReturns)
	}
	if len(m.SMBReturns) < minLen {
		minLen = len(m.SMBReturns)
	}
	if len(m.HMLReturns) < minLen {
		minLen = len(m.HMLReturns)
	}
	if m.UseFiveFactor {
		if len(m.RMWReturns) < minLen {
			minLen = len(m.RMWReturns)
		}
		if len(m.CMAReturns) < minLen {
			minLen = len(m.CMAReturns)
		}
	}

	// 构建X矩阵和Y向量
	X := make([][]float64, minLen)
	Y := make([]float64, minLen)

	for i := 0; i < minLen; i++ {
		X[i] = make([]float64, k)
		X[i][0] = 1.0 // 截距项
		X[i][1] = m.MarketReturns[i]
		X[i][2] = m.SMBReturns[i]
		X[i][3] = m.HMLReturns[i]
		if m.UseFiveFactor {
			X[i][4] = m.RMWReturns[i]
			X[i][5] = m.CMAReturns[i]
		}
		Y[i] = dependent[i]
	}

	// 使用最小二乘法估计系数: β = (X'X)^-1 X'Y
	XtX := m.matrixMultiply(m.transpose(X), X)
	XtY := m.matrixVectorMultiply(m.transpose(X), Y)

	// 求解线性方程组 (X'X)β = X'Y
	beta := m.solveLinearSystem(XtX, XtY)

	// 计算预测值和残差
	predicted := make([]float64, minLen)
	residuals := make([]float64, minLen)
	tss := 0.0 // 总平方和
	rss := 0.0 // 残差平方和

	meanY := 0.0
	for _, y := range Y {
		meanY += y
	}
	meanY /= float64(minLen)

	for i := 0; i < minLen; i++ {
		predicted[i] = 0
		for j := 0; j < k; j++ {
			predicted[i] += X[i][j] * beta[j]
		}
		residuals[i] = Y[i] - predicted[i]
		tss += (Y[i] - meanY) * (Y[i] - meanY)
		rss += residuals[i] * residuals[i]
	}

	// 计算R2
	r2 := 0.0
	if tss > 0 {
		r2 = 1 - rss/tss
	}

	// 调整R2
	adjR2 := 1 - (1-r2)*float64(minLen-1)/float64(minLen-k)

	// 计算标准误
	mse := rss / float64(minLen-k)
	varCovar := m.scalarMatrixMultiply(m.matrixInverse(XtX), mse)

	stdErrors := make([]float64, k)
	for i := 0; i < k; i++ {
		stdErrors[i] = math.Sqrt(varCovar[i][i])
	}

	// 构建结果
	coefficients := map[string]float64{
		"alpha":  beta[0],
		"market": beta[1],
		"smb":    beta[2],
		"hml":    beta[3],
	}
	stdErrMap := map[string]float64{
		"alpha":  stdErrors[0],
		"market": stdErrors[1],
		"smb":    stdErrors[2],
		"hml":    stdErrors[3],
	}
	tStats := map[string]float64{
		"alpha":  beta[0] / stdErrors[0],
		"market": beta[1] / stdErrors[1],
		"smb":    beta[2] / stdErrors[2],
		"hml":    beta[3] / stdErrors[3],
	}

	if m.UseFiveFactor {
		coefficients["rmw"] = beta[4]
		coefficients["cma"] = beta[5]
		stdErrMap["rmw"] = stdErrors[4]
		stdErrMap["cma"] = stdErrors[5]
		tStats["rmw"] = beta[4] / stdErrors[4]
		tStats["cma"] = beta[5] / stdErrors[5]
	}

	// 计算P值 (简化计算，使用正态分布近似)
	pValues := make(map[string]float64)
	for key, tStat := range tStats {
		pValues[key] = 2 * (1 - normalCDF(math.Abs(tStat)))
	}

	return &RegressionResult{
		Coefficients: coefficients,
		StdErrors:    stdErrMap,
		TStats:       tStats,
		PValues:      pValues,
		R2:           r2,
		AdjR2:        adjR2,
		Residuals:    residuals,
	}, nil
}

// calculateFactorContributions 计算各因子贡献
func (m *FamaFrenchModel) calculateFactorContributions(exposure *FactorExposure) map[string]float64 {
	contributions := make(map[string]float64)

	// 计算因子平均收益
	avgMarket := 0.0
	avgSMB := 0.0
	avgHML := 0.0

	for _, r := range m.MarketReturns {
		avgMarket += r
	}
	for _, r := range m.SMBReturns {
		avgSMB += r
	}
	for _, r := range m.HMLReturns {
		avgHML += r
	}

	avgMarket /= float64(len(m.MarketReturns))
	avgSMB /= float64(len(m.SMBReturns))
	avgHML /= float64(len(m.HMLReturns))

	// 年化
	contributions["market"] = exposure.Market * avgMarket * 12
	contributions["smb"] = exposure.Size * avgSMB * 12
	contributions["hml"] = exposure.Value * avgHML * 12
	contributions["alpha"] = exposure.Alpha * 12

	if m.UseFiveFactor {
		avgRMW := 0.0
		avgCMA := 0.0
		for _, r := range m.RMWReturns {
			avgRMW += r
		}
		for _, r := range m.CMAReturns {
			avgCMA += r
		}
		avgRMW /= float64(len(m.RMWReturns))
		avgCMA /= float64(len(m.CMAReturns))

		contributions["rmw"] = exposure.Profitability * avgRMW * 12
		contributions["cma"] = exposure.Investment * avgCMA * 12
	}

	return contributions
}

// GetFactorStatistics 获取因子统计信息
func (m *FamaFrenchModel) GetFactorStatistics() []*FactorReturn {
	factors := []*FactorReturn{
		m.calculateFactorStats("Market (Rm-Rf)", m.MarketReturns),
		m.calculateFactorStats("Size (SMB)", m.SMBReturns),
		m.calculateFactorStats("Value (HML)", m.HMLReturns),
	}

	if m.UseFiveFactor {
		factors = append(factors, m.calculateFactorStats("Profitability (RMW)", m.RMWReturns))
		factors = append(factors, m.calculateFactorStats("Investment (CMA)", m.CMAReturns))
	}

	return factors
}

// calculateFactorStats 计算单个因子的统计信息
func (m *FamaFrenchModel) calculateFactorStats(name string, returns []float64) *FactorReturn {
	if len(returns) == 0 {
		return &FactorReturn{Name: name}
	}

	// 年化收益
	totalReturn := 0.0
	for _, r := range returns {
		totalReturn += r
	}
	avgReturn := totalReturn / float64(len(returns))
	annualizedReturn := avgReturn * 12

	// 年化波动率
	variance := 0.0
	for _, r := range returns {
		variance += (r - avgReturn) * (r - avgReturn)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))
	annualizedVol := stdDev * math.Sqrt(12)

	// 夏普比率 (假设无风险利率为0)
	sharpe := 0.0
	if annualizedVol > 0 {
		sharpe = annualizedReturn / annualizedVol
	}

	// 最大回撤
	maxDrawdown := m.calculateMaxDrawdown(returns)

	return &FactorReturn{
		Name:        name,
		Annualized:  annualizedReturn,
		Volatility:  annualizedVol,
		Sharpe:      sharpe,
		MaxDrawdown: maxDrawdown,
	}
}

// calculateMaxDrawdown 计算最大回撤
func (m *FamaFrenchModel) calculateMaxDrawdown(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	cumulative := 1.0
	peak := 1.0
	maxDrawdown := 0.0

	for _, r := range returns {
		cumulative *= (1 + r)
		if cumulative > peak {
			peak = cumulative
		}
		drawdown := (peak - cumulative) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// ComparePortfolios 比较多个投资组合的因子特征
func (m *FamaFrenchModel) ComparePortfolios(
	portfolios map[string][]float64,
) map[string]*FactorAttribution {
	results := make(map[string]*FactorAttribution)

	for name, returns := range portfolios {
		attribution, err := m.AnalyzePortfolio(returns, nil)
		if err == nil {
			results[name] = attribution
		}
	}

	return results
}

// RiskDecomposition 风险分解
func (m *FamaFrenchModel) RiskDecomposition(
	exposures *FactorExposure,
) map[string]float64 {
	// 计算各因子对组合风险的贡献
	decomposition := make(map[string]float64)

	// 这里简化计算，使用暴露的绝对值作为风险贡献的近似
	totalExposure := math.Abs(exposures.Market) + math.Abs(exposures.Size) + math.Abs(exposures.Value)
	if m.UseFiveFactor {
		totalExposure += math.Abs(exposures.Profitability) + math.Abs(exposures.Investment)
	}

	if totalExposure > 0 {
		decomposition["market"] = math.Abs(exposures.Market) / totalExposure
		decomposition["size"] = math.Abs(exposures.Size) / totalExposure
		decomposition["value"] = math.Abs(exposures.Value) / totalExposure
		if m.UseFiveFactor {
			decomposition["profitability"] = math.Abs(exposures.Profitability) / totalExposure
			decomposition["investment"] = math.Abs(exposures.Investment) / totalExposure
		}
	}

	return decomposition
}

// ==================== 矩阵运算辅助函数 ====================

// transpose 矩阵转置
func (m *FamaFrenchModel) transpose(matrix [][]float64) [][]float64 {
	rows := len(matrix)
	cols := len(matrix[0])

	result := make([][]float64, cols)
	for i := range result {
		result[i] = make([]float64, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[j][i] = matrix[i][j]
		}
	}

	return result
}

// matrixMultiply 矩阵乘法
func (m *FamaFrenchModel) matrixMultiply(a, b [][]float64) [][]float64 {
	rows := len(a)
	cols := len(b[0])
	inner := len(b)

	result := make([][]float64, rows)
	for i := range result {
		result[i] = make([]float64, cols)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			for k := 0; k < inner; k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}

	return result
}

// matrixVectorMultiply 矩阵与向量乘法
func (m *FamaFrenchModel) matrixVectorMultiply(matrix [][]float64, vector []float64) []float64 {
	rows := len(matrix)
	result := make([]float64, rows)

	for i := 0; i < rows; i++ {
		for j := 0; j < len(vector); j++ {
			result[i] += matrix[i][j] * vector[j]
		}
	}

	return result
}

// solveLinearSystem 求解线性方程组 (使用高斯消元法)
func (m *FamaFrenchModel) solveLinearSystem(A [][]float64, b []float64) []float64 {
	n := len(A)

	// 构建增广矩阵
	augmented := make([][]float64, n)
	for i := range augmented {
		augmented[i] = make([]float64, n+1)
		copy(augmented[i], A[i])
		augmented[i][n] = b[i]
	}

	// 前向消元
	for i := 0; i < n; i++ {
		// 主元选择
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(augmented[k][i]) > math.Abs(augmented[maxRow][i]) {
				maxRow = k
			}
		}
		augmented[i], augmented[maxRow] = augmented[maxRow], augmented[i]

		// 消元
		for k := i + 1; k < n; k++ {
			factor := augmented[k][i] / augmented[i][i]
			for j := i; j <= n; j++ {
				augmented[k][j] -= factor * augmented[i][j]
			}
		}
	}

	// 回代
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = augmented[i][n]
		for j := i + 1; j < n; j++ {
			x[i] -= augmented[i][j] * x[j]
		}
		x[i] /= augmented[i][i]
	}

	return x
}

// matrixInverse 矩阵求逆 (简化实现，适用于小矩阵)
func (m *FamaFrenchModel) matrixInverse(matrix [][]float64) [][]float64 {
	n := len(matrix)
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
	}

	// 单位矩阵
	for i := 0; i < n; i++ {
		result[i][i] = 1.0
	}

	// 使用增广矩阵方法求逆
	augmented := make([][]float64, n)
	for i := range augmented {
		augmented[i] = make([]float64, 2*n)
		copy(augmented[i], matrix[i])
		copy(augmented[i][n:], result[i])
	}

	// 高斯消元
	for i := 0; i < n; i++ {
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(augmented[k][i]) > math.Abs(augmented[maxRow][i]) {
				maxRow = k
			}
		}
		augmented[i], augmented[maxRow] = augmented[maxRow], augmented[i]

		pivot := augmented[i][i]
		for j := 0; j < 2*n; j++ {
			augmented[i][j] /= pivot
		}

		for k := 0; k < n; k++ {
			if k != i {
				factor := augmented[k][i]
				for j := 0; j < 2*n; j++ {
					augmented[k][j] -= factor * augmented[i][j]
				}
			}
		}
	}

	// 提取逆矩阵
	for i := 0; i < n; i++ {
		copy(result[i], augmented[i][n:])
	}

	return result
}

// scalarMatrixMultiply 标量与矩阵乘法
func (m *FamaFrenchModel) scalarMatrixMultiply(matrix [][]float64, scalar float64) [][]float64 {
	rows := len(matrix)
	cols := len(matrix[0])

	result := make([][]float64, rows)
	for i := range result {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = matrix[i][j] * scalar
		}
	}

	return result
}

// normalCDF 标准正态分布累积分布函数
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// LoadFactorData 加载因子数据 (简化实现，实际应从数据库或API获取)
func (m *FamaFrenchModel) LoadFactorData(
	marketReturns,
	smbReturns,
	hmlReturns,
	riskFreeReturns []float64,
) {
	m.MarketReturns = marketReturns
	m.SMBReturns = smbReturns
	m.HMLReturns = hmlReturns
	m.RiskFreeReturns = riskFreeReturns
}

// LoadFiveFactorData 加载五因子数据
func (m *FamaFrenchModel) LoadFiveFactorData(
	marketReturns,
	smbReturns,
	hmlReturns,
	rmwReturns,
	cmaReturns,
	riskFreeReturns []float64,
) {
	m.MarketReturns = marketReturns
	m.SMBReturns = smbReturns
	m.HMLReturns = hmlReturns
	m.RMWReturns = rmwReturns
	m.CMAReturns = cmaReturns
	m.RiskFreeReturns = riskFreeReturns
	m.UseFiveFactor = true
}

// GenerateSampleFactorData 生成示例因子数据 (用于测试)
func GenerateSampleFactorData(periods int) (
	marketReturns,
	smbReturns,
	hmlReturns,
	riskFreeReturns []float64,
) {
	// 生成模拟的月度因子数据
	marketReturns = make([]float64, periods)
	smbReturns = make([]float64, periods)
	hmlReturns = make([]float64, periods)
	riskFreeReturns = make([]float64, periods)

	// 历史平均因子收益 (月度)
	// 市场溢价: ~0.5% /月
	// SMB: ~0.2% /月
	// HML: ~0.3% /月
	// 无风险利率: ~0.15% /月

	for i := 0; i < periods; i++ {
		marketReturns[i] = 0.005 + randNorm()*0.045 // 年化约6%，波动率约16%
		smbReturns[i] = 0.002 + randNorm()*0.03     // 年化约2.4%，波动率约10%
		hmlReturns[i] = 0.003 + randNorm()*0.03     // 年化约3.6%，波动率约10%
		riskFreeReturns[i] = 0.0015                 // 年化约1.8%
	}

	return
}

// randNorm 生成标准正态分布随机数 (Box-Muller变换)
func randNorm() float64 {
	// 简化实现，实际应使用更好的随机数生成器
	// 这里使用近似方法
	sum := 0.0
	for i := 0; i < 12; i++ {
		sum += randFloat()
	}
	return sum - 6.0
}

// randFloat 生成[0,1)均匀分布随机数
func randFloat() float64 {
	// 使用简单伪随机数生成器
	// 实际应使用crypto/rand或math/rand
	return 0.5 // 简化返回
}

// SortByFactorExposure 按因子暴露排序
func SortByFactorExposure(attributions map[string]*FactorAttribution, factor string) []string {
	type pair struct {
		name     string
		exposure float64
	}

	pairs := make([]pair, 0, len(attributions))
	for name, attr := range attributions {
		var exposure float64
		switch factor {
		case "market":
			exposure = attr.Exposures.Market
		case "size":
			exposure = attr.Exposures.Size
		case "value":
			exposure = attr.Exposures.Value
		case "profitability":
			exposure = attr.Exposures.Profitability
		case "investment":
			exposure = attr.Exposures.Investment
		}
		pairs = append(pairs, pair{name, exposure})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].exposure > pairs[j].exposure
	})

	result := make([]string, len(pairs))
	for i, p := range pairs {
		result[i] = p.name
	}

	return result
}
