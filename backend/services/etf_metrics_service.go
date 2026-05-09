package services

import (
	"math"

	"etf-insight/models"
)

type ETFMetricsService struct{}

func NewETFMetricsService() *ETFMetricsService {
	return &ETFMetricsService{}
}

// HandlerMetrics 指标数据结构（从 handlers 包迁移）
type HandlerMetrics struct {
	Volatility  float64
	TotalReturn float64
	MaxDrawdown float64
	SharpeRatio float64
}

// CalculateFromPrices 从历史价格计算指标
func (s *ETFMetricsService) CalculateFromPrices(prices []models.ETFData, period string) *HandlerMetrics {
	if len(prices) < 2 {
		return &HandlerMetrics{}
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		prevPrice := prices[i].ClosePrice.InexactFloat64()
		currPrice := prices[i-1].ClosePrice.InexactFloat64()
		if prevPrice > 0 {
			returns[i-1] = (currPrice - prevPrice) / prevPrice
		}
	}

	firstPrice := prices[len(prices)-1].ClosePrice.InexactFloat64()
	lastPrice := prices[0].ClosePrice.InexactFloat64()
	totalReturn := 0.0
	if firstPrice > 0 {
		totalReturn = (lastPrice - firstPrice) / firstPrice * 100
	}

	volatility := s.calculateVolatility(returns)
	maxDrawdown := s.calculateMaxDrawdown(prices)
	sharpeRatio := s.calculateSharpeRatio(returns, 0.02)

	return &HandlerMetrics{
		Volatility:  volatility,
		TotalReturn: totalReturn,
		MaxDrawdown: maxDrawdown,
		SharpeRatio: sharpeRatio,
	}
}

func (s *ETFMetricsService) calculateVolatility(returns []float64) float64 {
	if len(returns) < 10 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	return stdDev * math.Sqrt(252) * 100
}

func (s *ETFMetricsService) calculateMaxDrawdown(prices []models.ETFData) float64 {
	if len(prices) < 10 {
		return 0
	}

	maxDrawdown := 0.0
	peak := prices[len(prices)-1].ClosePrice.InexactFloat64()

	for i := len(prices) - 1; i >= 0; i-- {
		price := prices[i].ClosePrice.InexactFloat64()
		if price > peak {
			peak = price
		}
		drawdown := (peak - price) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown * 100
}

func (s *ETFMetricsService) calculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) < 10 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	meanReturn := sum / float64(len(returns))

	annualizedReturn := meanReturn * 252

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-meanReturn, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))
	annualizedStdDev := stdDev * math.Sqrt(252)

	if annualizedStdDev == 0 {
		return 0
	}

	return (annualizedReturn - riskFreeRate) / annualizedStdDev
}
