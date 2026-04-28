package utils

import (
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

// CalculateMean 计算均值
func CalculateMean(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

// CalculateVariance 计算方差 (样本方差，除以 n-1)
func CalculateVariance(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	if len(values) < 2 {
		return decimal.Zero
	}

	variance := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	return variance.Div(decimal.NewFromInt(int64(len(values) - 1)))
}

// CalculatePopulationVariance 计算总体方差 (除以 n)
func CalculatePopulationVariance(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	variance := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	return variance.Div(decimal.NewFromInt(int64(len(values))))
}

// CalculateStdDev 计算标准差 (样本标准差)
func CalculateStdDev(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	variance := CalculateVariance(values, mean)
	if variance.IsZero() {
		return decimal.Zero
	}
	return decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
}

// CalculatePopulationStdDev 计算总体标准差
func CalculatePopulationStdDev(values []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	variance := CalculatePopulationVariance(values, mean)
	if variance.IsZero() {
		return decimal.Zero
	}
	return decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
}

// CalculateCovariance 计算协方差
func CalculateCovariance(values1, values2 []decimal.Decimal) decimal.Decimal {
	minLen := int(math.Min(float64(len(values1)), float64(len(values2))))
	if minLen < 2 {
		return decimal.Zero
	}

	mean1 := CalculateMean(values1[:minLen])
	mean2 := CalculateMean(values2[:minLen])

	cov := decimal.Zero
	for i := 0; i < minLen; i++ {
		diff1 := values1[i].Sub(mean1)
		diff2 := values2[i].Sub(mean2)
		cov = cov.Add(diff1.Mul(diff2))
	}
	return cov.Div(decimal.NewFromInt(int64(minLen - 1)))
}

// SortDecimals 对 decimal 数组进行排序 (升序)
func SortDecimals(values []decimal.Decimal) {
	sort.Slice(values, func(i, j int) bool {
		return values[i].LessThan(values[j])
	})
}

// SortDecimalsDesc 对 decimal 数组进行排序 (降序)
func SortDecimalsDesc(values []decimal.Decimal) {
	sort.Slice(values, func(i, j int) bool {
		return values[i].GreaterThan(values[j])
	})
}

// NormalizeWeights 权重归一化，使权重之和为 1
func NormalizeWeights(weights []decimal.Decimal) []decimal.Decimal {
	if len(weights) == 0 {
		return []decimal.Decimal{}
	}

	total := decimal.Zero
	for _, w := range weights {
		total = total.Add(w)
	}

	if total.IsZero() {
		// 如果总和为0，返回等权重
		equalWeight := decimal.NewFromFloat(1.0 / float64(len(weights)))
		result := make([]decimal.Decimal, len(weights))
		for i := range result {
			result[i] = equalWeight
		}
		return result
	}

	result := make([]decimal.Decimal, len(weights))
	for i, w := range weights {
		result[i] = w.Div(total)
	}
	return result
}

// CalculatePercentile 计算百分位数
func CalculatePercentile(values []decimal.Decimal, percentile float64) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}
	if percentile <= 0 {
		return values[0]
	}
	if percentile >= 1 {
		return values[len(values)-1]
	}

	// 复制并排序
	sorted := make([]decimal.Decimal, len(values))
	copy(sorted, values)
	SortDecimals(sorted)

	index := float64(len(sorted)-1) * percentile
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := decimal.NewFromFloat(index - float64(lower))
	return sorted[lower].Add(sorted[upper].Sub(sorted[lower]).Mul(weight))
}

// CalculateSkewness 计算偏度
func CalculateSkewness(values []decimal.Decimal, mean, stdDev decimal.Decimal) decimal.Decimal {
	if len(values) == 0 || stdDev.IsZero() {
		return decimal.Zero
	}

	thirdMoment := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		thirdMoment = thirdMoment.Add(diff.Mul(diff).Mul(diff))
	}
	thirdMoment = thirdMoment.Div(decimal.NewFromInt(int64(len(values))))

	stdDevPow3 := stdDev.Mul(stdDev).Mul(stdDev)
	return thirdMoment.Div(stdDevPow3)
}

// CalculateKurtosis 计算峰度
func CalculateKurtosis(values []decimal.Decimal, mean, stdDev decimal.Decimal) decimal.Decimal {
	if len(values) == 0 || stdDev.IsZero() {
		return decimal.Zero
	}

	fourthMoment := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		fourthMoment = fourthMoment.Add(diff.Mul(diff).Mul(diff).Mul(diff))
	}
	fourthMoment = fourthMoment.Div(decimal.NewFromInt(int64(len(values))))

	variance := stdDev.Mul(stdDev)
	return fourthMoment.Div(variance.Mul(variance)).Sub(decimal.NewFromInt(3))
}
