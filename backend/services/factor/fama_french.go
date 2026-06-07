package factor

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"etf-insight/models"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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
		return nil, fmt.Errorf("依赖变量数据为空")
	}

	// 构建自变量矩阵 X
	var k int // 变量数
	if m.UseFiveFactor {
		k = 6 // alpha + 5个因子
	} else {
		k = 4 // alpha + 3个因子
	}

	// 截断数据到相同长度
	minLen := min(len(m.MarketReturns), n)
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

	// 验证数据长度足够
	if minLen == 0 {
		return nil, fmt.Errorf("因子数据为空")
	}
	// 至少需要与变量数相同的数据点，但为了更稳定的结果，建议数据点 > 变量数
	if minLen < k {
		return nil, fmt.Errorf("数据点数量(%d)不足，至少需要%d个数据点", minLen, k)
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

	// 求解线性方程组 (X'X)β = X'Y，使用带正则化的方法处理奇异矩阵
	beta := m.solveLinearSystemWithRegularization(XtX, XtY, 1e-6)

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

	// 计算t统计量，避免除以0
	calcTStat := func(b, se float64) float64 {
		if se < 1e-10 {
			return 0
		}
		return b / se
	}

	tStats := map[string]float64{
		"alpha":  calcTStat(beta[0], stdErrors[0]),
		"market": calcTStat(beta[1], stdErrors[1]),
		"smb":    calcTStat(beta[2], stdErrors[2]),
		"hml":    calcTStat(beta[3], stdErrors[3]),
	}

	if m.UseFiveFactor {
		coefficients["rmw"] = beta[4]
		coefficients["cma"] = beta[5]
		stdErrMap["rmw"] = stdErrors[4]
		stdErrMap["cma"] = stdErrors[5]
		tStats["rmw"] = calcTStat(beta[4], stdErrors[4])
		tStats["cma"] = calcTStat(beta[5], stdErrors[5])
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

	// 安全计算平均值，避免除以0
	safeAvg := func(returns []float64) float64 {
		if len(returns) == 0 {
			return 0
		}
		sum := 0.0
		for _, r := range returns {
			sum += r
		}
		return sum / float64(len(returns))
	}

	// 计算因子平均收益
	avgMarket := safeAvg(m.MarketReturns)
	avgSMB := safeAvg(m.SMBReturns)
	avgHML := safeAvg(m.HMLReturns)

	// 年化
	contributions["market"] = exposure.Market * avgMarket * 12
	contributions["smb"] = exposure.Size * avgSMB * 12
	contributions["hml"] = exposure.Value * avgHML * 12
	contributions["alpha"] = exposure.Alpha * 12

	if m.UseFiveFactor {
		avgRMW := safeAvg(m.RMWReturns)
		avgCMA := safeAvg(m.CMAReturns)

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

	// 年化波动率 (使用样本标准差，除以 n-1)
	variance := 0.0
	for _, r := range returns {
		variance += (r - avgReturn) * (r - avgReturn)
	}

	// 使用样本标准差 (n-1)，至少需要2个数据点
	var stdDev float64
	if len(returns) > 1 {
		stdDev = math.Sqrt(variance / float64(len(returns)-1))
	} else {
		stdDev = 0
	}
	annualizedVol := stdDev * math.Sqrt(12)

	// 夏普比率 (假设无风险利率为0)
	// 设置最小波动率阈值，避免除以极小数产生异常大的夏普比率
	const minVolatility = 0.001 // 最小0.1%的年化波动率
	sharpe := 0.0
	if annualizedVol >= minVolatility {
		sharpe = annualizedReturn / annualizedVol
	} else if annualizedVol > 0 {
		// 波动率过低，使用最小阈值计算夏普比率
		sharpe = annualizedReturn / minVolatility
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

	for i := range rows {
		for j := range cols {
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

	for i := range rows {
		for j := range cols {
			for k := range inner {
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

	for i := range rows {
		for j := range vector {
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
	for i := range n {
		// 主元选择
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(augmented[k][i]) > math.Abs(augmented[maxRow][i]) {
				maxRow = k
			}
		}
		augmented[i], augmented[maxRow] = augmented[maxRow], augmented[i]

		// 检查主元是否为0（奇异矩阵）
		if math.Abs(augmented[i][i]) < 1e-10 {
			// 矩阵接近奇异，返回零解作为fallback
			return make([]float64, n)
		}

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
		// 检查主元是否为0
		if math.Abs(augmented[i][i]) < 1e-10 {
			x[i] = 0
			continue
		}
		x[i] = augmented[i][n]
		for j := i + 1; j < n; j++ {
			x[i] -= augmented[i][j] * x[j]
		}
		x[i] /= augmented[i][i]
	}

	return x
}

// solveLinearSystemWithRegularization 使用Tikhonov正则化求解线性方程组
// 这可以处理接近奇异的矩阵，提供更稳定的解
func (m *FamaFrenchModel) solveLinearSystemWithRegularization(A [][]float64, b []float64, lambda float64) []float64 {
	n := len(A)

	// 添加正则化项: (A + λI)x = b
	// 这可以提高矩阵的条件数，使求解更稳定
	regularizedA := make([][]float64, n)
	for i := range regularizedA {
		regularizedA[i] = make([]float64, n)
		copy(regularizedA[i], A[i])
		// 只对对角线元素添加正则化（岭回归）
		regularizedA[i][i] += lambda
	}

	// 使用标准的高斯消元法求解
	return m.solveLinearSystem(regularizedA, b)
}

// matrixInverse 矩阵求逆 (简化实现，适用于小矩阵)
func (m *FamaFrenchModel) matrixInverse(matrix [][]float64) [][]float64 {
	n := len(matrix)
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
	}

	// 单位矩阵
	for i := range n {
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
	for i := range n {
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(augmented[k][i]) > math.Abs(augmented[maxRow][i]) {
				maxRow = k
			}
		}
		augmented[i], augmented[maxRow] = augmented[maxRow], augmented[i]

		pivot := augmented[i][i]
		// 检查主元是否为0（奇异矩阵）
		if math.Abs(pivot) < 1e-10 {
			// 矩阵接近奇异，返回单位矩阵作为fallback
			for ii := range n {
				for jj := range n {
					if ii == jj {
						result[ii][jj] = 1.0
					} else {
						result[ii][jj] = 0.0
					}
				}
			}
			return result
		}
		for j := 0; j < 2*n; j++ {
			augmented[i][j] /= pivot
		}

		for k := range n {
			if k != i {
				factor := augmented[k][i]
				for j := 0; j < 2*n; j++ {
					augmented[k][j] -= factor * augmented[i][j]
				}
			}
		}
	}

	// 提取逆矩阵
	for i := range n {
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
		for j := range cols {
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

// GenerateSampleFactorData 生成示例因子数据 (用于测试或当真实数据不可用时)
// 注意: 生产环境应使用LoadFactorDataFromDB从数据库获取真实因子数据
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

	for i := range periods {
		marketReturns[i] = 0.005 + randNorm()*0.045 // 年化约6%，波动率约16%
		smbReturns[i] = 0.002 + randNorm()*0.03     // 年化约2.4%，波动率约10%
		hmlReturns[i] = 0.003 + randNorm()*0.03     // 年化约3.6%，波动率约10%
		riskFreeReturns[i] = 0.0015                 // 年化约1.8%
	}

	return
}

// GenerateSampleFiveFactorData 生成示例五因子数据 (用于测试或当真实数据不可用时)
func GenerateSampleFiveFactorData(periods int) (
	marketReturns,
	smbReturns,
	hmlReturns,
	rmwReturns,
	cmaReturns,
	riskFreeReturns []float64,
) {
	// 生成模拟的月度因子数据
	marketReturns = make([]float64, periods)
	smbReturns = make([]float64, periods)
	hmlReturns = make([]float64, periods)
	rmwReturns = make([]float64, periods)
	cmaReturns = make([]float64, periods)
	riskFreeReturns = make([]float64, periods)

	// 历史平均因子收益 (月度)
	// 市场溢价: ~0.5% /月
	// SMB: ~0.2% /月
	// HML: ~0.3% /月
	// RMW (盈利因子): ~0.25% /月
	// CMA (投资因子): ~0.2% /月
	// 无风险利率: ~0.15% /月

	for i := range periods {
		marketReturns[i] = 0.005 + randNorm()*0.045 // 年化约6%，波动率约16%
		smbReturns[i] = 0.002 + randNorm()*0.03     // 年化约2.4%，波动率约10%
		hmlReturns[i] = 0.003 + randNorm()*0.03     // 年化约3.6%，波动率约10%
		rmwReturns[i] = 0.0025 + randNorm()*0.025   // 年化约3%，波动率约8.7%
		cmaReturns[i] = 0.002 + randNorm()*0.025    // 年化约2.4%，波动率约8.7%
		riskFreeReturns[i] = 0.0015                 // 年化约1.8%
	}

	return
}

// GenerateSamplePortfolioReturns 生成示例组合收益率数据 (用于测试或当真实数据不可用时)
func GenerateSamplePortfolioReturns(periods int) []float64 {
	returns := make([]float64, periods)

	// 生成模拟的组合收益率
	// 假设组合有适度的市场暴露和一定的超额收益
	for i := range periods {
		// 基于市场收益 + alpha + 噪声
		marketReturn := 0.005 + randNorm()*0.045 // 市场收益
		alpha := 0.001                           // 月度 alpha ~0.1%
		returns[i] = marketReturn + alpha + randNorm()*0.01
	}

	return returns
}

// FactorDataSource 因子数据来源
const (
	FactorSourceKennethFrench = "kenneth_french" // Kenneth French数据库
	FactorSourceAQR           = "aqr"            // AQR因子数据
	FactorSourceLocalDB       = "local_db"       // 本地数据库
)

// frenchDataFactorURLs maps factor model types to Kenneth French data URLs.
var frenchDataFactorURLs = map[string]string{
	"3factor_monthly": "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_Factors_CSV.zip",
	"3factor_daily":   "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_Factors_daily_CSV.zip",
	"5factor_monthly": "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_5_Factors_2x3_CSV.zip",
	"5factor_daily":   "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_5_Factors_2x3_daily_CSV.zip",
}

// FrenchFactorRow represents a single parsed row of factor data from the Kenneth French library.
type FrenchFactorRow struct {
	Date  time.Time
	MktRF float64
	SMB   float64
	HML   float64
	RMW   float64 // Only for 5-factor
	CMA   float64 // Only for 5-factor
	RF    float64
}

// LoadFactorDataFromFrench downloads and parses factor data from the Kenneth French Data Library.
// frequency: "monthly" or "daily"
// fiveFactor: true for 5-factor model, false for 3-factor
// Returns parsed rows filtered by date range.
func LoadFactorDataFromFrench(startDate, endDate time.Time, frequency string, fiveFactor bool) ([]FrenchFactorRow, error) {
	key := "3factor_" + frequency
	if fiveFactor {
		key = "5factor_" + frequency
	}

	url, ok := frenchDataFactorURLs[key]
	if !ok {
		return nil, fmt.Errorf("unsupported factor model/frequency combination: %s", key)
	}

	csvData, err := downloadAndUnzipFrenchData(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download French data: %w", err)
	}

	rows, err := parseFrenchCSVData(csvData, fiveFactor)
	if err != nil {
		return nil, fmt.Errorf("failed to parse French data: %w", err)
	}

	// Filter by date range
	var filtered []FrenchFactorRow
	for _, row := range rows {
		if !row.Date.Before(startDate) && !row.Date.After(endDate) {
			filtered = append(filtered, row)
		}
	}

	return filtered, nil
}

// downloadAndUnzipFrenchData downloads a ZIP file from the given URL and extracts the first CSV file.
func downloadAndUnzipFrenchData(url string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to open ZIP: %w", err)
	}

	// Find the first CSV file in the ZIP
	for _, f := range zipReader.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, ".csv") {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open file %s in ZIP: %w", f.Name, err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s in ZIP: %w", f.Name, err)
			}

			return string(data), nil
		}
	}

	return "", fmt.Errorf("no CSV file found in ZIP archive")
}

// parseFrenchCSVData parses the Kenneth French CSV format.
// The format has:
//   - Header lines with description text
//   - A blank line
//   - Column header line (e.g., "Mkt-RF   SMB   HML   RF")
//   - Data rows with YYYYMM (monthly) or YYYYMMDD (daily) dates and percentage values
//   - Values may be separated by spaces or tabs
//   - Some rows may have footer text after the data
func parseFrenchCSVData(csvData string, fiveFactor bool) ([]FrenchFactorRow, error) {
	lines := strings.Split(csvData, "\n")
	var rows []FrenchFactorRow

	// Find the header line with column names
	headerFound := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Look for the header line containing "Mkt-RF"
		if strings.Contains(trimmed, "Mkt-RF") && strings.Contains(trimmed, "SMB") && strings.Contains(trimmed, "HML") {
			headerFound = true
			// Parse data starting from the next line
			for j := i + 1; j < len(lines); j++ {
				dataLine := strings.TrimSpace(lines[j])
				if dataLine == "" {
					continue
				}

				row, err := parseFrenchCSVLine(dataLine, fiveFactor)
				if err != nil {
					// Skip lines that can't be parsed (likely footer text)
					continue
				}
				rows = append(rows, *row)
			}
			break
		}
	}

	if !headerFound {
		return nil, fmt.Errorf("could not find header line with 'Mkt-RF' in CSV data")
	}

	return rows, nil
}

// parseFrenchCSVLine parses a single data line from the Kenneth French CSV.
// Format: "YYYYMM  value1  value2  value3  [value4  value5]  value6"
// Values are percentages and will be converted to decimals (divided by 100).
func parseFrenchCSVLine(line string, fiveFactor bool) (*FrenchFactorRow, error) {
	// Split by whitespace (spaces and tabs)
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, fmt.Errorf("insufficient fields in line: %s", line)
	}

	dateStr := fields[0]

	// Determine date format based on length
	var date time.Time
	var err error
	if len(dateStr) == 6 {
		// Monthly: YYYYMM
		date, err = time.Parse("200601", dateStr)
	} else if len(dateStr) == 8 {
		// Daily: YYYYMMDD
		date, err = time.Parse("20060102", dateStr)
	} else {
		return nil, fmt.Errorf("unexpected date format: %s", dateStr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	// Parse factor values (percentages -> decimals)
	expectedCols := 5 // Mkt-RF, SMB, HML, RF = 4 values + date = 5
	if fiveFactor {
		expectedCols = 7 // Mkt-RF, SMB, HML, RMW, CMA, RF = 6 values + date = 7
	}

	if len(fields) < expectedCols {
		return nil, fmt.Errorf("expected %d fields for %s, got %d", expectedCols, map[bool]string{true: "5-factor", false: "3-factor"}[fiveFactor], len(fields))
	}

	mktRF, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Mkt-RF: %w", err)
	}

	smb, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SMB: %w", err)
	}

	hml, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HML: %w", err)
	}

	var rmw, cma float64
	rfIdx := 4
	if fiveFactor {
		rmw, err = strconv.ParseFloat(fields[4], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RMW: %w", err)
		}
		cma, err = strconv.ParseFloat(fields[5], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CMA: %w", err)
		}
		rfIdx = 6
	}

	rf, err := strconv.ParseFloat(fields[rfIdx], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RF: %w", err)
	}

	return &FrenchFactorRow{
		Date:  date,
		MktRF: mktRF / 100, // Convert percentage to decimal
		SMB:   smb / 100,
		HML:   hml / 100,
		RMW:   rmw / 100,
		CMA:   cma / 100,
		RF:    rf / 100,
	}, nil
}

// LoadFactorDataFromFrenchLegacy is the legacy interface that returns raw float64 slices.
// Deprecated: Use LoadFactorDataFromFrench instead for new code.
func LoadFactorDataFromFrenchLegacy(startDate, endDate time.Time) (
	marketReturns,
	smbReturns,
	hmlReturns,
	riskFreeReturns []float64,
	err error,
) {
	rows, err := LoadFactorDataFromFrench(startDate, endDate, "monthly", false)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, row := range rows {
		marketReturns = append(marketReturns, row.MktRF)
		smbReturns = append(smbReturns, row.SMB)
		hmlReturns = append(hmlReturns, row.HML)
		riskFreeReturns = append(riskFreeReturns, row.RF)
	}

	return marketReturns, smbReturns, hmlReturns, riskFreeReturns, nil
}

// FrenchRowsToFactorData converts parsed French factor rows to FactorData model instances.
// factorName should be one of: "Mkt-RF", "SMB", "HML", "RMW", "CMA".
func FrenchRowsToFactorData(rows []FrenchFactorRow, factorName string) []models.FactorData {
	data := make([]models.FactorData, 0, len(rows))
	for _, row := range rows {
		var value float64
		switch factorName {
		case "Mkt-RF":
			value = row.MktRF
		case "SMB":
			value = row.SMB
		case "HML":
			value = row.HML
		case "RMW":
			value = row.RMW
		case "CMA":
			value = row.CMA
		default:
			continue
		}
		data = append(data, models.FactorData{
			FactorName: factorName,
			Date:       row.Date,
			Value:      decimal.NewFromFloat(value),
			DataSource: FactorSourceKennethFrench,
			CreatedAt:  time.Now(),
		})
	}
	return data
}

// LoadFactorDataFromDB 从本地数据库加载因子数据
// 需要预先在数据库中存储历史因子数据
func LoadFactorDataFromDB(db *gorm.DB, startDate, endDate time.Time) (
	marketReturns,
	smbReturns,
	hmlReturns,
	riskFreeReturns []float64,
	err error,
) {
	// 查询数据库中的因子数据
	// 假设有一个factor_data表存储历史因子数据
	type FactorData struct {
		Date          time.Time
		MarketPremium float64 // 市场溢价 (Rm - Rf)
		SMB           float64 // 市值因子
		HML           float64 // 价值因子
		RiskFreeRate  float64 // 无风险利率
	}

	var factors []FactorData
	result := db.Table("factor_data").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date ASC").
		Find(&factors)

	if result.Error != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load factor data from DB: %w", result.Error)
	}

	if len(factors) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("no factor data found for the specified period")
	}

	marketReturns = make([]float64, len(factors))
	smbReturns = make([]float64, len(factors))
	hmlReturns = make([]float64, len(factors))
	riskFreeReturns = make([]float64, len(factors))

	for i, f := range factors {
		marketReturns[i] = f.MarketPremium
		smbReturns[i] = f.SMB
		hmlReturns[i] = f.HML
		riskFreeReturns[i] = f.RiskFreeRate
	}

	return
}

// ParseFrenchCSV 解析Kenneth French数据库CSV格式
func ParseFrenchCSV(data string) (
	dates []time.Time,
	marketReturns,
	smbReturns,
	hmlReturns,
	riskFreeReturns []float64,
	err error,
) {
	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// French数据库格式: 第一列是日期(YYYYMM), 后面是因子值(百分比)
	for _, record := range records {
		if len(record) < 4 {
			continue
		}

		// 跳过标题行和非数据行
		dateStr := strings.TrimSpace(record[0])
		if len(dateStr) != 6 {
			continue
		}

		date, err := time.Parse("200601", dateStr)
		if err != nil {
			continue
		}

		mktrf, _ := strconv.ParseFloat(strings.TrimSpace(record[1]), 64)
		smb, _ := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		hml, _ := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		rf, _ := strconv.ParseFloat(strings.TrimSpace(record[4]), 64)

		// 转换为小数形式
		dates = append(dates, date)
		marketReturns = append(marketReturns, mktrf/100)
		smbReturns = append(smbReturns, smb/100)
		hmlReturns = append(hmlReturns, hml/100)
		riskFreeReturns = append(riskFreeReturns, rf/100)
	}

	return
}

// CalculateFactorFromETFs 从ETF组合计算因子暴露
// 使用市值因子ETF和价值因子ETF作为代理
func CalculateFactorFromETFs(
	marketETF string, // 市场ETF (如SPY)
	smallCapETF string, // 小盘ETF (如IWM)
	largeCapETF string, // 大盘ETF (如VV)
	valueETF string, // 价值ETF (如VTV)
	growthETF string, // 成长ETF (如VUG)
	startDate, endDate time.Time,
) (
	marketReturns,
	smbReturns,
	hmlReturns []float64,
	err error,
) {
	// 获取各ETF的历史数据
	getReturns := func(symbol string) ([]decimal.Decimal, error) {
		var data []models.ETFData
		result := models.DB.Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
			Order("date ASC").
			Find(&data)
		if result.Error != nil {
			return nil, result.Error
		}

		returns := make([]decimal.Decimal, len(data)-1)
		for i := 1; i < len(data); i++ {
			ret := data[i].ClosePrice.Sub(data[i-1].ClosePrice).Div(data[i-1].ClosePrice)
			returns[i-1] = ret
		}
		return returns, nil
	}

	marketRets, err := getReturns(marketETF)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get market ETF data: %w", err)
	}

	smallRets, err := getReturns(smallCapETF)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get small cap ETF data: %w", err)
	}

	largeRets, err := getReturns(largeCapETF)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get large cap ETF data: %w", err)
	}

	valueRets, err := getReturns(valueETF)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get value ETF data: %w", err)
	}

	growthRets, err := getReturns(growthETF)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get growth ETF data: %w", err)
	}

	// 计算因子收益
	minLen := min(len(smallRets), len(marketRets))
	if len(largeRets) < minLen {
		minLen = len(largeRets)
	}
	if len(valueRets) < minLen {
		minLen = len(valueRets)
	}
	if len(growthRets) < minLen {
		minLen = len(growthRets)
	}

	marketReturns = make([]float64, minLen)
	smbReturns = make([]float64, minLen)
	hmlReturns = make([]float64, minLen)

	for i := 0; i < minLen; i++ {
		marketReturns[i], _ = marketRets[i].Float64()
		smbReturn := smallRets[i].Sub(largeRets[i])
		smbReturns[i], _ = smbReturn.Float64()
		hmlReturn := valueRets[i].Sub(growthRets[i])
		hmlReturns[i], _ = hmlReturn.Float64()
	}

	return
}

// randNorm 生成标准正态分布随机数 (Box-Muller变换)
func randNorm() float64 {
	// 简化实现，实际应使用更好的随机数生成器
	// 这里使用近似方法
	sum := 0.0
	for range 12 {
		sum += randFloat()
	}
	return sum - 6.0
}

// randFloat 生成[0,1)均匀分布随机数
func randFloat() float64 {
	return rand.Float64()
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
