package services

import (
	"errors"
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidConfidenceLevel = errors.New("confidence level must be between 0 and 1")
	ErrInvalidReturns         = errors.New("returns slice cannot be empty")
)

// RiskModels 风险模型计算服务
type RiskModels struct{}

// NewRiskModels 创建风险模型服务
func NewRiskModels() *RiskModels {
	return &RiskModels{}
}

// VaRData 风险价值数据结构
type VaRData struct {
	VaR           decimal.Decimal `json:"var"`
	CVaR          decimal.Decimal `json:"cvar"`
	Confidence    decimal.Decimal `json:"confidence"`
	Method        string          `json:"method"`
	PeriodDays    int             `json:"period_days"`
	AnnualizedVaR decimal.Decimal `json:"annualized_var"`
}

// CalculateHistoricalVaR 使用历史模拟法计算 VaR 和 CVaR
// returns: 收益率序列（日收益率）
// confidence: 置信水平，如 0.95 表示 95% 置信度
func (rm *RiskModels) CalculateHistoricalVaR(returns []decimal.Decimal, confidence float64) (*VaRData, error) {
	if len(returns) == 0 {
		return nil, ErrInvalidReturns
	}
	if confidence <= 0 || confidence >= 1 {
		return nil, ErrInvalidConfidenceLevel
	}

	// 复制并排序收益率（从低到高）
	sortedReturns := make([]decimal.Decimal, len(returns))
	copy(sortedReturns, returns)
	sort.Slice(sortedReturns, func(i, j int) bool {
		return sortedReturns[i].LessThan(sortedReturns[j])
	})

	// 计算分位数位置
	index := int(math.Floor(float64(len(sortedReturns)) * (1 - confidence)))
	if index >= len(sortedReturns) {
		index = len(sortedReturns) - 1
	}
	if index < 0 {
		index = 0
	}

	// VaR 是第 index 个最小值（负数表示损失）
	varValue := sortedReturns[index]

	// CVaR (Expected Shortfall) 是超过 VaR 阈值的平均损失
	cvarSum := decimal.Zero
	count := 0
	for i := 0; i <= index; i++ {
		cvarSum = cvarSum.Add(sortedReturns[i])
		count++
	}

	cvarValue := decimal.Zero
	if count > 0 {
		cvarValue = cvarSum.Div(decimal.NewFromInt(int64(count)))
	}

	// 年化 VaR (假设 252 个交易日)
	annualizedVaR := varValue.Mul(decimal.NewFromFloat(math.Sqrt(252)))

	return &VaRData{
		VaR:           varValue.Round(4),
		CVaR:          cvarValue.Round(4),
		Confidence:    decimal.NewFromFloat(confidence).Round(4),
		Method:        "historical",
		PeriodDays:    len(returns),
		AnnualizedVaR: annualizedVaR.Round(4),
	}, nil
}

// CalculateParametricVaR 使用参数法（方差-协方差法）计算 VaR
// returns: 收益率序列
// confidence: 置信水平
func (rm *RiskModels) CalculateParametricVaR(returns []decimal.Decimal, confidence float64) (*VaRData, error) {
	if len(returns) == 0 {
		return nil, ErrInvalidReturns
	}
	if confidence <= 0 || confidence >= 1 {
		return nil, ErrInvalidConfidenceLevel
	}

	// 计算均值
	mean := calculateMean(returns)

	// 计算标准差
	stdDev := calculateStdDev(returns, mean)

	// 获取标准正态分布的分位数
	zScore := getZScore(confidence)

	// VaR = -(均值 - z * 标准差)
	// 负号是因为 VaR 通常表示为正数损失
	zDecimal := decimal.NewFromFloat(zScore)
	varValue := mean.Sub(zDecimal.Mul(stdDev)).Neg()

	// CVaR 计算说明：
	// 当前使用近似值 VaR * 1.2，适用于正态分布假设下的快速估算。
	//
	// 精确计算公式（未来可优化）：
	// CVaR = μ - σ * φ(z) / (1 - α)
	// 其中 φ(z) 是标准正态密度函数，α 是置信水平
	//
	// 参考：https://en.wikipedia.org/wiki/Expected_shortfall
	cvarValue := varValue.Mul(decimal.NewFromFloat(1.2)) // 近似值

	// 年化
	annualizedVaR := varValue.Mul(decimal.NewFromFloat(math.Sqrt(252)))

	return &VaRData{
		VaR:           varValue.Round(4),
		CVaR:          cvarValue.Round(4),
		Confidence:    decimal.NewFromFloat(confidence).Round(4),
		Method:        "parametric",
		PeriodDays:    len(returns),
		AnnualizedVaR: annualizedVaR.Round(4),
	}, nil
}

// PortfolioRiskData 投资组合风险数据结构
type PortfolioRiskData struct {
	PortfolioVaR           decimal.Decimal            `json:"portfolio_var"`
	PortfolioCVaR          decimal.Decimal            `json:"portfolio_cvar"`
	DiversificationBenefit decimal.Decimal            `json:"diversification_benefit"`
	ComponentVaR           map[string]decimal.Decimal `json:"component_var"`
	MarginalVaR            map[string]decimal.Decimal `json:"marginal_var"`
}

// CalculatePortfolioVaR 计算投资组合 VaR
// weights: 各资产权重
// returns: 各资产收益率矩阵（每列是一个资产的收益率序列）
// confidence: 置信水平
func (rm *RiskModels) CalculatePortfolioVaR(
	weights map[string]decimal.Decimal,
	returns map[string][]decimal.Decimal,
	confidence float64,
) (*PortfolioRiskData, error) {
	if len(weights) == 0 || len(returns) == 0 {
		return nil, errors.New("weights and returns cannot be empty")
	}

	// 计算各资产的 VaR
	assetVaRs := make(map[string]decimal.Decimal)
	for asset, assetReturns := range returns {
		varData, err := rm.CalculateHistoricalVaR(assetReturns, confidence)
		if err != nil {
			return nil, err
		}
		assetVaRs[asset] = varData.VaR
	}

	// 计算组合收益率序列
	portfolioReturns := make([]decimal.Decimal, 0)
	// 假设所有资产有相同数量的收益率观测值
	numObservations := len(returns[getFirstKey(returns)])

	for i := range numObservations {
		portfolioReturn := decimal.Zero
		for asset, assetReturns := range returns {
			if i < len(assetReturns) {
				weight := weights[asset]
				portfolioReturn = portfolioReturn.Add(assetReturns[i].Mul(weight))
			}
		}
		portfolioReturns = append(portfolioReturns, portfolioReturn)
	}

	// 计算组合 VaR
	portfolioVarData, err := rm.CalculateHistoricalVaR(portfolioReturns, confidence)
	if err != nil {
		return nil, err
	}

	// 计算分散化收益
	undiversifiedVaR := decimal.Zero
	for asset, weight := range weights {
		undiversifiedVaR = undiversifiedVaR.Add(assetVaRs[asset].Mul(weight))
	}
	diversificationBenefit := undiversifiedVaR.Sub(portfolioVarData.VaR)

	// 计算 Component VaR (近似值)
	componentVaR := make(map[string]decimal.Decimal)
	for asset, weight := range weights {
		// Component VaR ≈ Weight * Marginal VaR
		// 简化计算：使用资产 VaR 乘以权重
		componentVaR[asset] = assetVaRs[asset].Mul(weight).Round(4)
	}

	// 计算 Marginal VaR (近似值)
	marginalVaR := make(map[string]decimal.Decimal)
	for asset := range weights {
		// Marginal VaR ≈ Portfolio VaR * Beta
		// 简化计算：使用资产 VaR 与组合 VaR 的比例
		if portfolioVarData.VaR.GreaterThan(decimal.Zero) {
			marginalVaR[asset] = assetVaRs[asset].Div(portfolioVarData.VaR).Round(4)
		} else {
			marginalVaR[asset] = decimal.Zero
		}
	}

	return &PortfolioRiskData{
		PortfolioVaR:           portfolioVarData.VaR,
		PortfolioCVaR:          portfolioVarData.CVaR,
		DiversificationBenefit: diversificationBenefit.Round(4),
		ComponentVaR:           componentVaR,
		MarginalVaR:            marginalVaR,
	}, nil
}

// RiskMetrics 综合风险指标
type RiskMetrics struct {
	Volatility       decimal.Decimal `json:"volatility"`
	SharpeRatio      decimal.Decimal `json:"sharpe_ratio"`
	SortinoRatio     decimal.Decimal `json:"sortino_ratio"`
	MaxDrawdown      decimal.Decimal `json:"max_drawdown"`
	CalmarRatio      decimal.Decimal `json:"calmar_ratio"`
	Beta             decimal.Decimal `json:"beta"`
	Alpha            decimal.Decimal `json:"alpha"`
	TrackingError    decimal.Decimal `json:"tracking_error"`
	InformationRatio decimal.Decimal `json:"information_ratio"`
}

// CalculateRiskMetrics 计算综合风险指标
// returns: 资产收益率序列
// riskFreeRate: 无风险利率（日利率）
// benchmarkReturns: 基准收益率序列（可选，用于计算 Beta 和 Alpha）
func (rm *RiskModels) CalculateRiskMetrics(
	returns []decimal.Decimal,
	riskFreeRate decimal.Decimal,
	benchmarkReturns []decimal.Decimal,
) (*RiskMetrics, error) {
	if len(returns) == 0 {
		return nil, ErrInvalidReturns
	}

	// 计算年化波动率
	mean := calculateMean(returns)
	stdDev := calculateStdDev(returns, mean)
	annualizedVolatility := stdDev.Mul(decimal.NewFromFloat(math.Sqrt(252)))

	// 计算年化收益率
	annualizedReturn := mean.Mul(decimal.NewFromInt(252))

	// 计算夏普比率
	excessReturn := annualizedReturn.Sub(riskFreeRate.Mul(decimal.NewFromInt(252)))
	sharpeRatio := decimal.Zero
	if annualizedVolatility.GreaterThan(decimal.Zero) {
		sharpeRatio = excessReturn.Div(annualizedVolatility)
	}

	// 计算索提诺比率（只考虑下行波动）
	downsideReturns := make([]decimal.Decimal, 0)
	for _, r := range returns {
		if r.LessThan(riskFreeRate) {
			downsideReturns = append(downsideReturns, r)
		}
	}
	downsideStdDev := decimal.Zero
	if len(downsideReturns) > 0 {
		downsideMean := calculateMean(downsideReturns)
		downsideStdDev = calculateStdDev(downsideReturns, downsideMean)
	}
	annualizedDownsideDev := downsideStdDev.Mul(decimal.NewFromFloat(math.Sqrt(252)))
	sortinoRatio := decimal.Zero
	if annualizedDownsideDev.GreaterThan(decimal.Zero) {
		sortinoRatio = excessReturn.Div(annualizedDownsideDev)
	}

	// 计算最大回撤
	maxDrawdown := calculateMaxDrawdownFromReturns(returns)

	// 计算卡尔玛比率
	calmarRatio := decimal.Zero
	if maxDrawdown.GreaterThan(decimal.Zero) {
		calmarRatio = annualizedReturn.Div(maxDrawdown)
	}

	result := &RiskMetrics{
		Volatility:   annualizedVolatility.Round(4),
		SharpeRatio:  sharpeRatio.Round(4),
		SortinoRatio: sortinoRatio.Round(4),
		MaxDrawdown:  maxDrawdown.Round(4),
		CalmarRatio:  calmarRatio.Round(4),
	}

	// 如果有基准数据，计算 Beta 和 Alpha
	if len(benchmarkReturns) > 0 && len(benchmarkReturns) == len(returns) {
		beta, alpha := calculateBetaAlpha(returns, benchmarkReturns, riskFreeRate)
		result.Beta = beta.Round(4)
		result.Alpha = alpha.Round(4)

		// 计算跟踪误差
		trackingError := calculateTrackingError(returns, benchmarkReturns)
		result.TrackingError = trackingError.Round(4)

		// 计算信息比率
		if trackingError.GreaterThan(decimal.Zero) {
			result.InformationRatio = excessReturn.Div(trackingError).Round(4)
		}
	}

	return result, nil
}

// 辅助函数

func calculateMean(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}
	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

func calculateStdDev(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}
	variance := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(values))))
	return decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
}

func getZScore(confidence float64) float64 {
	// 标准正态分布的分位数近似值
	// 95% -> 1.645, 99% -> 2.326
	zScores := map[float64]float64{
		0.90:  1.282,
		0.95:  1.645,
		0.99:  2.326,
		0.999: 3.090,
	}

	if z, ok := zScores[confidence]; ok {
		return z
	}

	// 线性插值
	if confidence < 0.90 {
		return 1.282 * confidence / 0.90
	} else if confidence < 0.95 {
		return 1.282 + (1.645-1.282)*(confidence-0.90)/0.05
	} else if confidence < 0.99 {
		return 1.645 + (2.326-1.645)*(confidence-0.95)/0.04
	}
	return 2.326 + (3.090-2.326)*(confidence-0.99)/0.009
}

func calculateMaxDrawdownFromReturns(returns []decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	// 计算累计收益
	cumulative := make([]decimal.Decimal, len(returns))
	cumulative[0] = decimal.NewFromInt(1).Add(returns[0])
	for i := 1; i < len(returns); i++ {
		cumulative[i] = cumulative[i-1].Mul(decimal.NewFromInt(1).Add(returns[i]))
	}

	maxDrawdown := decimal.Zero
	peak := cumulative[0]

	for i := 1; i < len(cumulative); i++ {
		if cumulative[i].GreaterThan(peak) {
			peak = cumulative[i]
		}
		drawdown := peak.Sub(cumulative[i]).Div(peak)
		if drawdown.GreaterThan(maxDrawdown) {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

func calculateBetaAlpha(returns, benchmarkReturns []decimal.Decimal, riskFreeRate decimal.Decimal) (beta, alpha decimal.Decimal) {
	if len(returns) != len(benchmarkReturns) || len(returns) == 0 {
		return decimal.Zero, decimal.Zero
	}

	// 计算超额收益
	assetExcess := make([]decimal.Decimal, len(returns))
	benchmarkExcess := make([]decimal.Decimal, len(benchmarkReturns))

	for i := range returns {
		assetExcess[i] = returns[i].Sub(riskFreeRate)
		benchmarkExcess[i] = benchmarkReturns[i].Sub(riskFreeRate)
	}

	// 计算协方差和方差
	meanAsset := calculateMean(assetExcess)
	meanBenchmark := calculateMean(benchmarkExcess)

	covariance := decimal.Zero
	variance := decimal.Zero

	for i := range returns {
		diffAsset := assetExcess[i].Sub(meanAsset)
		diffBenchmark := benchmarkExcess[i].Sub(meanBenchmark)
		covariance = covariance.Add(diffAsset.Mul(diffBenchmark))
		variance = variance.Add(diffBenchmark.Mul(diffBenchmark))
	}

	covariance = covariance.Div(decimal.NewFromInt(int64(len(returns))))
	variance = variance.Div(decimal.NewFromInt(int64(len(returns))))

	// Beta = Cov(r, rm) / Var(rm)
	if variance.GreaterThan(decimal.Zero) {
		beta = covariance.Div(variance)
	}

	// Alpha = E[r] - rf - Beta * (E[rm] - rf)
	alpha = meanAsset.Sub(beta.Mul(meanBenchmark))

	return beta, alpha
}

func calculateTrackingError(returns, benchmarkReturns []decimal.Decimal) decimal.Decimal {
	if len(returns) != len(benchmarkReturns) || len(returns) == 0 {
		return decimal.Zero
	}

	differences := make([]decimal.Decimal, len(returns))
	for i := range returns {
		differences[i] = returns[i].Sub(benchmarkReturns[i])
	}

	mean := calculateMean(differences)
	stdDev := calculateStdDev(differences, mean)

	// 年化跟踪误差
	return stdDev.Mul(decimal.NewFromFloat(math.Sqrt(252)))
}

func getFirstKey(m map[string][]decimal.Decimal) string {
	for k := range m {
		return k
	}
	return ""
}
