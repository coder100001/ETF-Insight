package backtest

import (
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

// FactorLibrary 因子库
// 包含 Fama-French 三因子、五因子、动量、价值、低波等因子
type FactorLibrary struct {
	dataProvider DataProvider
	cache        map[string][]decimal.Decimal // 因子值缓存
}

// NewFactorLibrary 创建因子库
func NewFactorLibrary(provider DataProvider) *FactorLibrary {
	return &FactorLibrary{
		dataProvider: provider,
		cache:        make(map[string][]decimal.Decimal),
	}
}

// FactorType 因子类型
type FactorType string

const (
	// Fama-French 三因子
	FactorMarket FactorType = "MKT" // 市场因子 (Rm - Rf)
	FactorSMB    FactorType = "SMB" // 规模因子 (Small Minus Big)
	FactorHML    FactorType = "HML" // 价值因子 (High Minus Low)

	// Fama-French 五因子扩展
	FactorRMW FactorType = "RMW" // 盈利能力因子 (Robust Minus Weak)
	FactorCMA FactorType = "CMA" // 投资因子 (Conservative Minus Aggressive)

	// 其他因子
	FactorMomentum FactorType = "MOM"    // 动量因子
	FactorLowVol   FactorType = "LOWVOL" // 低波因子
	FactorQuality  FactorType = "QUAL"   // 质量因子
	FactorValue    FactorType = "VALUE"  // 价值因子
	FactorSize     FactorType = "SIZE"   // 规模因子
)

// FactorDefinition 因子定义
type FactorDefinition struct {
	Name        string     `json:"name"`
	Type        FactorType `json:"type"`
	Description string     `json:"description"`
	Formula     string     `json:"formula"`
	Category    string     `json:"category"` // risk, style, momentum, quality
}

// FactorExposure 因子暴露
type FactorExposure struct {
	FactorType  FactorType      `json:"factor_type"`
	Value       decimal.Decimal `json:"value"`
	Percentile  decimal.Decimal `json:"percentile"` // 在全市场中的分位数
	Description string          `json:"description"`
}

// FamaFrench3Factor Fama-French 三因子模型
type FamaFrench3Factor struct {
	MarketReturn decimal.Decimal `json:"market_return"` // 市场收益率
	SMB          decimal.Decimal `json:"smb"`           // 规模因子
	HML          decimal.Decimal `json:"hml"`           // 价值因子
	RiskFreeRate decimal.Decimal `json:"risk_free_rate"`
}

// FamaFrench5Factor Fama-French 五因子模型
type FamaFrench5Factor struct {
	FamaFrench3Factor
	RMW decimal.Decimal `json:"rmw"` // 盈利能力因子
	CMA decimal.Decimal `json:"cma"` // 投资因子
}

// CalculateFamaFrench3Factor 计算 Fama-French 三因子
// 基于历史收益率数据计算
func (f *FactorLibrary) CalculateFamaFrench3Factor(
	returns []decimal.Decimal,
	marketReturns []decimal.Decimal,
	marketCapRatios []decimal.Decimal, // 市值比例
	bookToMarketRatios []decimal.Decimal, // 账面市值比
) *FamaFrench3Factor {
	if len(returns) == 0 || len(marketReturns) == 0 {
		return nil
	}

	// 市场因子 (MKT): 市场收益率 - 无风险利率
	riskFreeRate := decimal.NewFromFloat(0.045 / 252.0) // 日化无风险利率
	marketFactor := calculateAverage(marketReturns).Sub(riskFreeRate)

	// SMB (Small Minus Big): 小市值股票收益 - 大市值股票收益
	smb := f.calculateSMB(returns, marketCapRatios)

	// HML (High Minus Low): 高账面市值比股票收益 - 低账面市值比股票收益
	hml := f.calculateHML(returns, bookToMarketRatios)

	return &FamaFrench3Factor{
		MarketReturn: marketFactor,
		SMB:          smb,
		HML:          hml,
		RiskFreeRate: riskFreeRate,
	}
}

// CalculateFamaFrench5Factor 计算 Fama-French 五因子
func (f *FactorLibrary) CalculateFamaFrench5Factor(
	returns []decimal.Decimal,
	marketReturns []decimal.Decimal,
	marketCapRatios []decimal.Decimal,
	bookToMarketRatios []decimal.Decimal,
	profitabilityRatios []decimal.Decimal, // 盈利能力比率
	investmentRatios []decimal.Decimal, // 投资比率
) *FamaFrench5Factor {
	threeFactor := f.CalculateFamaFrench3Factor(returns, marketReturns, marketCapRatios, bookToMarketRatios)
	if threeFactor == nil {
		return nil
	}

	// RMW (Robust Minus Weak): 高盈利能力股票收益 - 低盈利能力股票收益
	rmw := f.calculateRMW(returns, profitabilityRatios)

	// CMA (Conservative Minus Aggressive): 保守投资股票收益 - 激进投资股票收益
	cma := f.calculateCMA(returns, investmentRatios)

	return &FamaFrench5Factor{
		FamaFrench3Factor: *threeFactor,
		RMW:               rmw,
		CMA:               cma,
	}
}

// calculateSMB 计算规模因子
func (f *FactorLibrary) calculateSMB(returns []decimal.Decimal, marketCapRatios []decimal.Decimal) decimal.Decimal {
	if len(returns) != len(marketCapRatios) || len(returns) == 0 {
		return decimal.Zero
	}

	// 按市值中位数分组
	median := calculateMedian(marketCapRatios)

	var smallReturns []decimal.Decimal
	var bigReturns []decimal.Decimal

	for i, mc := range marketCapRatios {
		if mc.LessThan(median) {
			smallReturns = append(smallReturns, returns[i])
		} else {
			bigReturns = append(bigReturns, returns[i])
		}
	}

	smallAvg := calculateAverage(smallReturns)
	bigAvg := calculateAverage(bigReturns)

	return smallAvg.Sub(bigAvg)
}

// calculateHML 计算价值因子
func (f *FactorLibrary) calculateHML(returns []decimal.Decimal, bookToMarketRatios []decimal.Decimal) decimal.Decimal {
	if len(returns) != len(bookToMarketRatios) || len(returns) == 0 {
		return decimal.Zero
	}

	// 按账面市值比分为三组: 高(价值)、中、低(成长)
	p30 := calculatePercentile(bookToMarketRatios, 0.30)
	p70 := calculatePercentile(bookToMarketRatios, 0.70)

	var highReturns []decimal.Decimal // 高账面市值比 (价值股)
	var lowReturns []decimal.Decimal  // 低账面市值比 (成长股)

	for i, btm := range bookToMarketRatios {
		if btm.GreaterThanOrEqual(p70) {
			highReturns = append(highReturns, returns[i])
		} else if btm.LessThanOrEqual(p30) {
			lowReturns = append(lowReturns, returns[i])
		}
	}

	highAvg := calculateAverage(highReturns)
	lowAvg := calculateAverage(lowReturns)

	return highAvg.Sub(lowAvg)
}

// calculateRMW 计算盈利能力因子
func (f *FactorLibrary) calculateRMW(returns []decimal.Decimal, profitabilityRatios []decimal.Decimal) decimal.Decimal {
	if len(returns) != len(profitabilityRatios) || len(returns) == 0 {
		return decimal.Zero
	}

	// 按盈利能力分为高、低两组
	median := calculateMedian(profitabilityRatios)

	var robustReturns []decimal.Decimal // 高盈利能力
	var weakReturns []decimal.Decimal   // 低盈利能力

	for i, prof := range profitabilityRatios {
		if prof.GreaterThanOrEqual(median) {
			robustReturns = append(robustReturns, returns[i])
		} else {
			weakReturns = append(weakReturns, returns[i])
		}
	}

	robustAvg := calculateAverage(robustReturns)
	weakAvg := calculateAverage(weakReturns)

	return robustAvg.Sub(weakAvg)
}

// calculateCMA 计算投资因子
func (f *FactorLibrary) calculateCMA(returns []decimal.Decimal, investmentRatios []decimal.Decimal) decimal.Decimal {
	if len(returns) != len(investmentRatios) || len(returns) == 0 {
		return decimal.Zero
	}

	// 按投资比率分为保守、激进两组
	median := calculateMedian(investmentRatios)

	var conservativeReturns []decimal.Decimal // 低投资比率
	var aggressiveReturns []decimal.Decimal   // 高投资比率

	for i, inv := range investmentRatios {
		if inv.LessThanOrEqual(median) {
			conservativeReturns = append(conservativeReturns, returns[i])
		} else {
			aggressiveReturns = append(aggressiveReturns, returns[i])
		}
	}

	conservativeAvg := calculateAverage(conservativeReturns)
	aggressiveAvg := calculateAverage(aggressiveReturns)

	return conservativeAvg.Sub(aggressiveAvg)
}

// CalculateMomentumFactor 计算动量因子
// MOM = 过去12个月收益率 - 过去1个月收益率
func (f *FactorLibrary) CalculateMomentumFactor(prices []decimal.Decimal, lookbackMonths int) decimal.Decimal {
	if len(prices) < lookbackMonths {
		return decimal.Zero
	}

	// 计算12个月收益率
	return12M := prices[len(prices)-1].Div(prices[0]).Sub(decimal.NewFromInt(1))

	// 排除最近1个月 (避免短期反转效应)
	if len(prices) > 20 { // 约1个月交易日
		return1M := prices[len(prices)-1].Div(prices[len(prices)-20]).Sub(decimal.NewFromInt(1))
		return return12M.Sub(return1M)
	}

	return return12M
}

// CalculateLowVolFactor 计算低波因子
// 基于历史波动率，低波动股票往往有超额收益
func (f *FactorLibrary) CalculateLowVolFactor(returns []decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	// 计算波动率
	volatility := calculateVolatility(returns)

	// 低波因子 = 1 / 波动率 (波动率越低，因子值越高)
	if volatility.GreaterThan(decimal.Zero) {
		return decimal.NewFromInt(1).Div(volatility)
	}

	return decimal.Zero
}

// CalculateQualityFactor 计算质量因子
// 基于盈利能力、盈利稳定性、财务杠杆等
func (f *FactorLibrary) CalculateQualityFactor(
	roe decimal.Decimal, // 净资产收益率
	roa decimal.Decimal, // 总资产收益率
	debtToEquity decimal.Decimal, // 资产负债率
	earningsStability decimal.Decimal, // 盈利稳定性
) decimal.Decimal {
	// 质量因子综合评分
	// ROE和ROA越高越好，资产负债率越低越好，盈利稳定性越高越好

	score := decimal.Zero

	// ROE贡献 (假设合理范围0-30%)
	if roe.GreaterThan(decimal.Zero) {
		score = score.Add(roe.Mul(decimal.NewFromInt(3)))
	}

	// ROA贡献 (假设合理范围0-15%)
	if roa.GreaterThan(decimal.Zero) {
		score = score.Add(roa.Mul(decimal.NewFromInt(2)))
	}

	// 财务杠杆惩罚 (资产负债率越高，惩罚越大)
	if debtToEquity.GreaterThan(decimal.Zero) {
		score = score.Sub(debtToEquity.Mul(decimal.NewFromFloat(0.5)))
	}

	// 盈利稳定性奖励
	score = score.Add(earningsStability)

	return score
}

// CalculateValueFactor 计算价值因子
// 基于市盈率、市净率、市销率等估值指标
func (f *FactorLibrary) CalculateValueFactor(
	pe decimal.Decimal, // 市盈率
	pb decimal.Decimal, // 市净率
	ps decimal.Decimal, // 市销率
	dividendYield decimal.Decimal, // 股息率
) decimal.Decimal {
	// 价值因子综合评分
	// 估值指标越低越好，股息率越高越好

	score := decimal.Zero

	// PE倒数 (PE越低，得分越高)
	if pe.GreaterThan(decimal.Zero) {
		score = score.Add(decimal.NewFromInt(1).Div(pe).Mul(decimal.NewFromInt(100)))
	}

	// PB倒数
	if pb.GreaterThan(decimal.Zero) {
		score = score.Add(decimal.NewFromInt(1).Div(pb).Mul(decimal.NewFromInt(50)))
	}

	// PS倒数
	if ps.GreaterThan(decimal.Zero) {
		score = score.Add(decimal.NewFromInt(1).Div(ps).Mul(decimal.NewFromInt(30)))
	}

	// 股息率奖励
	score = score.Add(dividendYield.Mul(decimal.NewFromInt(100)))

	return score
}

// CalculateFactorExposure 计算因子暴露
// 使用回归分析计算投资组合对各因子的敏感度
func (f *FactorLibrary) CalculateFactorExposure(
	portfolioReturns []decimal.Decimal,
	factorReturns map[FactorType][]decimal.Decimal,
) map[FactorType]decimal.Decimal {
	exposures := make(map[FactorType]decimal.Decimal)

	for factorType, factorValues := range factorReturns {
		if len(factorValues) != len(portfolioReturns) {
			continue
		}

		// 简单线性回归计算因子暴露 (beta)
		beta := calculateBeta(portfolioReturns, factorValues)
		exposures[factorType] = beta
	}

	return exposures
}

// 辅助函数

func calculateAverage(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}

	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

func calculateMedian(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	// 复制并排序
	sorted := make([]decimal.Decimal, len(values))
	copy(sorted, values)

	// 使用 sort.Slice，O(n log n)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LessThan(sorted[j])
	})

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return sorted[mid-1].Add(sorted[mid]).Div(decimal.NewFromInt(2))
	}
	return sorted[mid]
}

func calculatePercentile(values []decimal.Decimal, p float64) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	// 复制并排序
	sorted := make([]decimal.Decimal, len(values))
	copy(sorted, values)

	// 使用 sort.Slice，O(n log n)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LessThan(sorted[j])
	})

	// 计算分位数位置
	index := float64(len(sorted)-1) * p
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	// 线性插值
	fraction := decimal.NewFromFloat(index - float64(lower))
	return sorted[lower].Add(sorted[upper].Sub(sorted[lower]).Mul(fraction))
}

func calculateVolatility(returns []decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	// 计算平均收益率
	mean := calculateAverage(returns)

	// 计算方差
	variance := decimal.Zero
	for _, r := range returns {
		diff := r.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(returns))))

	// 年化波动率
	volatility := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64() * 252))

	return volatility
}

func calculateBeta(assetReturns []decimal.Decimal, marketReturns []decimal.Decimal) decimal.Decimal {
	if len(assetReturns) != len(marketReturns) || len(assetReturns) == 0 {
		return decimal.Zero
	}

	// 计算协方差和方差
	assetMean := calculateAverage(assetReturns)
	marketMean := calculateAverage(marketReturns)

	covariance := decimal.Zero
	marketVariance := decimal.Zero

	for i := 0; i < len(assetReturns); i++ {
		assetDiff := assetReturns[i].Sub(assetMean)
		marketDiff := marketReturns[i].Sub(marketMean)

		covariance = covariance.Add(assetDiff.Mul(marketDiff))
		marketVariance = marketVariance.Add(marketDiff.Mul(marketDiff))
	}

	covariance = covariance.Div(decimal.NewFromInt(int64(len(assetReturns))))
	marketVariance = marketVariance.Div(decimal.NewFromInt(int64(len(marketReturns))))

	if marketVariance.IsZero() {
		return decimal.Zero
	}

	return covariance.Div(marketVariance)
}

// GetFactorDefinitions 获取所有因子定义
func (f *FactorLibrary) GetFactorDefinitions() []FactorDefinition {
	return []FactorDefinition{
		{
			Name:        "市场因子",
			Type:        FactorMarket,
			Description: "市场超额收益 (Rm - Rf)",
			Formula:     "R_market - R_f",
			Category:    "risk",
		},
		{
			Name:        "规模因子 (SMB)",
			Type:        FactorSMB,
			Description: "小市值股票收益减去大市值股票收益",
			Formula:     "R_small - R_big",
			Category:    "style",
		},
		{
			Name:        "价值因子 (HML)",
			Type:        FactorHML,
			Description: "高账面市值比股票收益减去低账面市值比股票收益",
			Formula:     "R_high_BM - R_low_BM",
			Category:    "style",
		},
		{
			Name:        "盈利能力因子 (RMW)",
			Type:        FactorRMW,
			Description: "高盈利能力股票收益减去低盈利能力股票收益",
			Formula:     "R_robust - R_weak",
			Category:    "quality",
		},
		{
			Name:        "投资因子 (CMA)",
			Type:        FactorCMA,
			Description: "保守投资股票收益减去激进投资股票收益",
			Formula:     "R_conservative - R_aggressive",
			Category:    "style",
		},
		{
			Name:        "动量因子",
			Type:        FactorMomentum,
			Description: "过去12个月收益率(排除最近1个月)",
			Formula:     "Return_12M - Return_1M",
			Category:    "momentum",
		},
		{
			Name:        "低波因子",
			Type:        FactorLowVol,
			Description: "波动率的倒数，低波动股票有超额收益",
			Formula:     "1 / Volatility",
			Category:    "style",
		},
		{
			Name:        "质量因子",
			Type:        FactorQuality,
			Description: "综合盈利能力、财务稳健性、盈利稳定性",
			Formula:     "f(ROE, ROA, D/E, Stability)",
			Category:    "quality",
		},
		{
			Name:        "价值因子",
			Type:        FactorValue,
			Description: "综合估值指标(PE, PB, PS)和股息率",
			Formula:     "f(PE, PB, PS, Dividend Yield)",
			Category:    "style",
		},
	}
}
