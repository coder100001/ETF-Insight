package services

import (
	"math"

	"etf-insight/services/statistics"

	"github.com/shopspring/decimal"
)

// DecimalMean 计算 decimal 切片的均值
func DecimalMean(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}
	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

// DecimalVariance 计算样本方差 (N-1)
func DecimalVariance(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	n := len(values)
	if n < 2 {
		return decimal.Zero
	}
	sum := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		sum = sum.Add(diff.Mul(diff))
	}
	return sum.Div(decimal.NewFromInt(int64(n - 1)))
}

// DecimalStdDev 计算标准差
func DecimalStdDev(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	variance := DecimalVariance(values, mean)
	if variance.IsNegative() {
		return decimal.Zero
	}
	// 使用 float64 开方后转回 decimal
	floatVar, _ := variance.Float64()
	return decimal.NewFromFloat(math.Sqrt(floatVar))
}

// DecimalToFloat64Slice 将 decimal 切片转换为 float64 切片
func DecimalToFloat64Slice(values []decimal.Decimal) []float64 {
	result := make([]float64, len(values))
	for i, v := range values {
		result[i], _ = v.Float64()
	}
	return result
}

// Float64ToDecimalSlice 将 float64 切片转换为 decimal 切片
func Float64ToDecimalSlice(values []float64) []decimal.Decimal {
	result := make([]decimal.Decimal, len(values))
	for i, v := range values {
		result[i] = decimal.NewFromFloat(v)
	}
	return result
}

// DecimalLogReturns 计算对数收益率 (decimal 版本)
// r_t = ln(P_t / P_{t-1})
func DecimalLogReturns(prices []decimal.Decimal) []decimal.Decimal {
	if len(prices) < 2 {
		return []decimal.Decimal{}
	}

	returns := make([]decimal.Decimal, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		prevPrice := prices[i-1]
		currPrice := prices[i]

		if prevPrice.IsPositive() {
			// ln(P_t / P_{t-1}) = ln(P_t) - ln(P_{t-1})
			prevFloat, _ := prevPrice.Float64()
			currFloat, _ := currPrice.Float64()
			logReturn := math.Log(currFloat / prevFloat)
			returns[i-1] = decimal.NewFromFloat(logReturn)
		}
	}

	return returns
}

// DecimalMaxDrawdown 计算最大回撤 (decimal 版本)
func DecimalMaxDrawdown(prices []decimal.Decimal) decimal.Decimal {
	if len(prices) == 0 {
		return decimal.Zero
	}

	maxDrawdown := decimal.Zero
	peak := prices[0]

	for _, price := range prices {
		if price.GreaterThan(peak) {
			peak = price
		}

		if peak.IsPositive() {
			drawdown := peak.Sub(price).Div(peak)
			if drawdown.GreaterThan(maxDrawdown) {
				maxDrawdown = drawdown
			}
		}
	}

	return maxDrawdown
}

// DecimalVaR 计算风险价值 (参数法, decimal 版本)
func DecimalVaR(returns []decimal.Decimal, confidence decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	mean := DecimalMean(returns)
	std := DecimalStdDev(returns, mean)

	// 使用 float64 计算 Z 分位数
	confFloat, _ := confidence.Float64()
	zScore := -inverseNormalCDF(confFloat)

	// VaR = 均值 + Z * 标准差
	return mean.Add(decimal.NewFromFloat(zScore).Mul(std))
}

// DecimalCVaR 计算条件风险价值 (参数法, decimal 版本)
func DecimalCVaR(returns []decimal.Decimal, confidence decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	mean := DecimalMean(returns)
	std := DecimalStdDev(returns, mean)

	// 使用 float64 计算 Z 分位数和 PDF/CDF
	confFloat, _ := confidence.Float64()
	zScore := -inverseNormalCDF(confFloat)

	// CVaR = mean - std * phi(Z) / Phi(Z)
	phi := (1.0 / math.Sqrt(2*math.Pi)) * math.Exp(-zScore*zScore/2)
	Phi := statistics.NormalCDF(zScore)
	if Phi < 1e-10 {
		Phi = 1e-10
	}

	cvarAdjustment := std.Mul(decimal.NewFromFloat(phi / Phi)).Neg()
	return mean.Add(cvarAdjustment)
}

// DecimalSortinoRatio 计算索提诺比率 (decimal 版本)
func DecimalSortinoRatio(returns []decimal.Decimal, riskFreeRate decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	meanReturn := DecimalMean(returns)

	// 计算下行标准差
	targetReturn := decimal.Zero
	downsideSum := decimal.Zero
	count := 0

	for _, r := range returns {
		if r.LessThan(targetReturn) {
			diff := r.Sub(targetReturn)
			downsideSum = downsideSum.Add(diff.Mul(diff))
			count++
		}
	}

	if count == 0 {
		return decimal.Zero
	}

	downsideStdFloat, _ := downsideSum.Div(decimal.NewFromInt(int64(len(returns)))).Float64()
	downsideStd := decimal.NewFromFloat(math.Sqrt(downsideStdFloat))

	if downsideStd.IsZero() {
		return decimal.Zero
	}

	// 年化处理
	annualReturn := meanReturn.Mul(decimal.NewFromInt(252))
	annualDownsideStd := downsideStd.Mul(decimal.NewFromFloat(math.Sqrt(252)))

	return annualReturn.Sub(riskFreeRate).Div(annualDownsideStd)
}

// DecimalCalmarRatio 计算卡尔玛比率 (decimal 版本)
func DecimalCalmarRatio(annualReturn, maxDrawdown decimal.Decimal) decimal.Decimal {
	if maxDrawdown.IsZero() {
		return decimal.Zero
	}
	return annualReturn.Div(maxDrawdown)
}

// DecimalSkewness 计算偏度 (decimal 版本)
func DecimalSkewness(returns []decimal.Decimal) decimal.Decimal {
	n := len(returns)
	if n < 3 {
		return decimal.Zero
	}

	mean := DecimalMean(returns)
	std := DecimalStdDev(returns, mean)

	if std.IsZero() {
		return decimal.Zero
	}

	sumCubed := decimal.Zero
	for _, r := range returns {
		z := r.Sub(mean).Div(std)
		sumCubed = sumCubed.Add(z.Mul(z).Mul(z))
	}

	// 样本偏度修正
	nDec := decimal.NewFromInt(int64(n))
	return nDec.Mul(sumCubed).Div(nDec.Sub(decimal.NewFromInt(1)).Mul(nDec.Sub(decimal.NewFromInt(2))))
}

// DecimalKurtosis 计算峰度 (decimal 版本, 超额峰度)
func DecimalKurtosis(returns []decimal.Decimal) decimal.Decimal {
	n := len(returns)
	if n < 4 {
		return decimal.Zero
	}

	mean := DecimalMean(returns)
	std := DecimalStdDev(returns, mean)

	if std.IsZero() {
		return decimal.Zero
	}

	sumFourth := decimal.Zero
	for _, r := range returns {
		z := r.Sub(mean).Div(std)
		z2 := z.Mul(z)
		sumFourth = sumFourth.Add(z2.Mul(z2))
	}

	nDec := decimal.NewFromInt(int64(n))
	kurtosis := sumFourth.Div(nDec)
	return kurtosis.Sub(decimal.NewFromInt(3)) // 超额峰度
}

// DecimalDownsideDeviation 计算下行偏差 (decimal 版本)
func DecimalDownsideDeviation(returns []decimal.Decimal, targetReturn decimal.Decimal) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	downsideSum := decimal.Zero
	count := 0

	for _, r := range returns {
		if r.LessThan(targetReturn) {
			diff := r.Sub(targetReturn)
			downsideSum = downsideSum.Add(diff.Mul(diff))
			count++
		}
	}

	if count == 0 {
		return decimal.Zero
	}

	floatVal, _ := downsideSum.Div(decimal.NewFromInt(int64(len(returns)))).Float64()
	return decimal.NewFromFloat(math.Sqrt(floatVal))
}
