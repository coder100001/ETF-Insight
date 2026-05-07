package services

import (
	"errors"
	"math"

	"github.com/shopspring/decimal"
)

var ErrInsufficientData = errors.New("insufficient data for calculation")

// TechnicalIndicators 技术指标计算服务
type TechnicalIndicators struct{}

// NewTechnicalIndicators 创建技术指标服务
func NewTechnicalIndicators() *TechnicalIndicators {
	return &TechnicalIndicators{}
}

// RSIData RSI 数据结构
type RSIData struct {
	RSI        decimal.Decimal `json:"rsi"`
	Overbought bool            `json:"overbought"`
	Oversold   bool            `json:"oversold"`
}

// CalculateRSI 计算 RSI (Relative Strength Index) 相对强弱指标
// period: 计算周期，通常为 14
// prices: 价格序列，从旧到新排列
func (ti *TechnicalIndicators) CalculateRSI(prices []decimal.Decimal, period int) (*RSIData, error) {
	if len(prices) < period+1 {
		return nil, ErrInsufficientData
	}

	// 计算价格变化
	gains := make([]decimal.Decimal, 0)
	losses := make([]decimal.Decimal, 0)

	for i := 1; i < len(prices); i++ {
		change := prices[i].Sub(prices[i-1])
		if change.GreaterThan(decimal.Zero) {
			gains = append(gains, change)
			losses = append(losses, decimal.Zero)
		} else {
			gains = append(gains, decimal.Zero)
			losses = append(losses, change.Abs())
		}
	}

	// 计算平均收益和平均损失
	avgGain := calculateAverage(gains[:period])
	avgLoss := calculateAverage(losses[:period])

	// 使用平滑移动平均计算后续值
	for i := period; i < len(gains); i++ {
		avgGain = avgGain.Mul(decimal.NewFromInt(int64(period - 1))).Add(gains[i]).Div(decimal.NewFromInt(int64(period)))
		avgLoss = avgLoss.Mul(decimal.NewFromInt(int64(period - 1))).Add(losses[i]).Div(decimal.NewFromInt(int64(period)))
	}

	// 计算 RS 和 RSI
	if avgLoss.Equal(decimal.Zero) {
		return &RSIData{
			RSI:        decimal.NewFromInt(100),
			Overbought: true,
			Oversold:   false,
		}, nil
	}

	rs := avgGain.Div(avgLoss)
	rsi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))

	return &RSIData{
		RSI:        rsi.Round(2),
		Overbought: rsi.GreaterThanOrEqual(decimal.NewFromInt(70)),
		Oversold:   rsi.LessThanOrEqual(decimal.NewFromInt(30)),
	}, nil
}

// MACDData MACD 数据结构
type MACDData struct {
	MACD      decimal.Decimal `json:"macd"`
	Signal    decimal.Decimal `json:"signal"`
	Histogram decimal.Decimal `json:"histogram"`
	Bullish   bool            `json:"bullish"`
}

// CalculateMACD 计算 MACD (Moving Average Convergence Divergence) 指数平滑异同移动平均线
// fastPeriod: 快线周期，通常为 12
// slowPeriod: 慢线周期，通常为 26
// signalPeriod: 信号线周期，通常为 9
func (ti *TechnicalIndicators) CalculateMACD(prices []decimal.Decimal, fastPeriod, slowPeriod, signalPeriod int) (*MACDData, error) {
	if len(prices) < slowPeriod+signalPeriod {
		return nil, ErrInsufficientData
	}

	// 计算 EMA
	fastEMA := calculateEMA(prices, fastPeriod)
	slowEMA := calculateEMA(prices, slowPeriod)

	// 确保两个 EMA 数组长度相同，以对齐较短的为准
	minLen := min(len(slowEMA), len(fastEMA))

	// 对齐 EMA 数组（取后 minLen 个元素）
	fastEMA = fastEMA[len(fastEMA)-minLen:]
	slowEMA = slowEMA[len(slowEMA)-minLen:]

	// 计算 MACD 线
	macdLine := make([]decimal.Decimal, minLen)
	for i := 0; i < minLen; i++ {
		macdLine[i] = fastEMA[i].Sub(slowEMA[i])
	}

	// 计算信号线 (MACD 的 EMA)
	signalLine := calculateEMA(macdLine, signalPeriod)

	// 获取最新值
	latestMACD := macdLine[len(macdLine)-1]
	latestSignal := signalLine[len(signalLine)-1]
	histogram := latestMACD.Sub(latestSignal)

	// 判断趋势
	bullish := latestMACD.GreaterThan(latestSignal)

	return &MACDData{
		MACD:      latestMACD.Round(4),
		Signal:    latestSignal.Round(4),
		Histogram: histogram.Round(4),
		Bullish:   bullish,
	}, nil
}

// BollingerBandsData 布林带数据结构
type BollingerBandsData struct {
	Upper     decimal.Decimal `json:"upper"`
	Middle    decimal.Decimal `json:"middle"`
	Lower     decimal.Decimal `json:"lower"`
	Bandwidth decimal.Decimal `json:"bandwidth"`
	PercentB  decimal.Decimal `json:"percent_b"`
}

// CalculateBollingerBands 计算布林带 (Bollinger Bands)
// period: 周期，通常为 20
// stdDev: 标准差倍数，通常为 2
func (ti *TechnicalIndicators) CalculateBollingerBands(prices []decimal.Decimal, period int, stdDev float64) (*BollingerBandsData, error) {
	if len(prices) < period {
		return nil, ErrInsufficientData
	}

	// 计算中轨 (SMA)
	middle := calculateSMA(prices[len(prices)-period:])

	// 计算标准差
	variance := calculateVariance(prices[len(prices)-period:], middle)
	stdDeviation := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	// 计算上下轨
	stdDevDecimal := decimal.NewFromFloat(stdDev)
	upper := middle.Add(stdDeviation.Mul(stdDevDecimal))
	lower := middle.Sub(stdDeviation.Mul(stdDevDecimal))

	// 计算带宽
	bandwidth := upper.Sub(lower).Div(middle).Mul(decimal.NewFromInt(100))

	// 计算 %B
	latestPrice := prices[len(prices)-1]
	percentB := latestPrice.Sub(lower).Div(upper.Sub(lower)).Mul(decimal.NewFromInt(100))

	return &BollingerBandsData{
		Upper:     upper.Round(4),
		Middle:    middle.Round(4),
		Lower:     lower.Round(4),
		Bandwidth: bandwidth.Round(4),
		PercentB:  percentB.Round(4),
	}, nil
}

// MovingAverageData 移动平均线数据结构
type MovingAverageData struct {
	SMA map[int]decimal.Decimal `json:"sma"`
	EMA map[int]decimal.Decimal `json:"ema"`
}

// CalculateMovingAverages 计算多条移动平均线
func (ti *TechnicalIndicators) CalculateMovingAverages(prices []decimal.Decimal, periods []int) (*MovingAverageData, error) {
	maxPeriod := 0
	for _, p := range periods {
		if p > maxPeriod {
			maxPeriod = p
		}
	}

	if len(prices) < maxPeriod {
		return nil, ErrInsufficientData
	}

	result := &MovingAverageData{
		SMA: make(map[int]decimal.Decimal),
		EMA: make(map[int]decimal.Decimal),
	}

	for _, period := range periods {
		// 计算 SMA
		sma := calculateSMA(prices[len(prices)-period:])
		result.SMA[period] = sma.Round(4)

		// 计算 EMA
		ema := calculateEMA(prices, period)
		result.EMA[period] = ema[len(ema)-1].Round(4)
	}

	return result, nil
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

func calculateSMA(prices []decimal.Decimal) decimal.Decimal {
	return calculateAverage(prices)
}

func calculateEMA(prices []decimal.Decimal, period int) []decimal.Decimal {
	if len(prices) < period {
		return nil
	}

	multiplier := decimal.NewFromFloat(2.0 / float64(period+1))
	ema := make([]decimal.Decimal, len(prices))

	// 第一个 EMA 使用 SMA
	ema[period-1] = calculateSMA(prices[:period])

	// 计算后续 EMA
	for i := period; i < len(prices); i++ {
		ema[i] = prices[i].Sub(ema[i-1]).Mul(multiplier).Add(ema[i-1])
	}

	return ema[period-1:]
}

func calculateVariance(prices []decimal.Decimal, mean decimal.Decimal) decimal.Decimal {
	if len(prices) == 0 {
		return decimal.Zero
	}

	sumSquaredDiff := decimal.Zero
	for _, price := range prices {
		diff := price.Sub(mean)
		sumSquaredDiff = sumSquaredDiff.Add(diff.Mul(diff))
	}

	return sumSquaredDiff.Div(decimal.NewFromInt(int64(len(prices))))
}
