package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ETFHistoricalMetrics ETF历史指标
type ETFHistoricalMetrics struct {
	Symbol        string    `json:"symbol"`
	AnnualReturn  float64   `json:"annual_return"`  // 年化收益率
	Volatility    float64   `json:"volatility"`     // 年化波动率
	SharpeRatio   float64   `json:"sharpe_ratio"`   // 夏普比率
	MaxDrawdown   float64   `json:"max_drawdown"`   // 最大回撤
	DividendYield float64   `json:"dividend_yield"` // 股息率
	DataPoints    int       `json:"data_points"`    // 数据点数量
	StartDate     time.Time `json:"start_date"`     // 数据开始日期
	EndDate       time.Time `json:"end_date"`       // 数据结束日期
}

// PortfolioAnalytics 组合分析指标
type PortfolioAnalytics struct {
	ExpectedReturn    float64                          `json:"expected_return"`    // 预期年化收益率
	Volatility        float64                          `json:"volatility"`         // 年化波动率
	SharpeRatio       float64                          `json:"sharpe_ratio"`       // 夏普比率
	SortinoRatio      float64                          `json:"sortino_ratio"`      // 索提诺比率
	MaxDrawdown       float64                          `json:"max_drawdown"`       // 最大回撤
	VaR95             float64                          `json:"var_95"`             // 95% VaR
	VaR99             float64                          `json:"var_99"`             // 99% VaR
	CVaR95            float64                          `json:"cvar_95"`            // 95% CVaR
	Beta              float64                          `json:"beta"`               // Beta系数
	Alpha             float64                          `json:"alpha"`              // Alpha系数
	CorrelationMatrix map[string]float64               `json:"correlation_matrix"` // 相关系数矩阵
	ETFMetrics        map[string]*ETFHistoricalMetrics `json:"etf_metrics"`
}

// PortfolioAnalyticsService 组合分析服务
type PortfolioAnalyticsService struct {
	db *gorm.DB
}

// NewPortfolioAnalyticsService 创建组合分析服务
func NewPortfolioAnalyticsService() *PortfolioAnalyticsService {
	return &PortfolioAnalyticsService{
		db: models.DB,
	}
}

// GetETFHistoricalData 获取ETF历史价格数据
func (s *PortfolioAnalyticsService) GetETFHistoricalData(symbol string, days int) ([]models.ETFData, error) {
	var data []models.ETFData

	// 计算开始日期
	startDate := time.Now().AddDate(0, 0, -days)

	err := s.db.Where("symbol = ? AND date >= ?", symbol, startDate).
		Order("date ASC").
		Find(&data).Error

	if err != nil {
		return nil, fmt.Errorf("获取ETF %s 历史数据失败: %w", symbol, err)
	}

	if len(data) < 30 {
		return nil, fmt.Errorf("ETF %s 历史数据不足，需要至少30天数据，当前只有%d天", symbol, len(data))
	}

	return data, nil
}

// CalculateLogReturns 计算对数收益率
// r_t = ln(P_t / P_{t-1})
func (s *PortfolioAnalyticsService) CalculateLogReturns(prices []decimal.Decimal) []float64 {
	returns := make([]float64, len(prices)-1)

	for i := 1; i < len(prices); i++ {
		prevPrice, _ := prices[i-1].Float64()
		currPrice, _ := prices[i].Float64()

		if prevPrice > 0 {
			returns[i-1] = math.Log(currPrice / prevPrice)
		}
	}

	return returns
}

// CalculateETFMetrics 计算ETF历史指标
func (s *PortfolioAnalyticsService) CalculateETFMetrics(symbol string, days int) (*ETFHistoricalMetrics, error) {
	data, err := s.GetETFHistoricalData(symbol, days)
	if err != nil {
		return nil, err
	}

	// 提取收盘价
	prices := make([]decimal.Decimal, len(data))
	for i, d := range data {
		prices[i] = d.ClosePrice
	}

	// 计算对数收益率
	returns := s.CalculateLogReturns(prices)

	if len(returns) == 0 {
		return nil, fmt.Errorf("无法计算收益率")
	}

	// 计算年化收益率
	// 使用几何平均: (P_end / P_start)^(252/n) - 1
	startPrice, _ := prices[0].Float64()
	endPrice, _ := prices[len(prices)-1].Float64()
	n := float64(len(returns))

	annualReturn := math.Pow(endPrice/startPrice, 252.0/n) - 1

	// 计算年化波动率
	// σ_annual = σ_daily * sqrt(252)
	meanReturn := s.mean(returns)
	variance := s.variance(returns, meanReturn)
	dailyVolatility := math.Sqrt(variance)
	annualVolatility := dailyVolatility * math.Sqrt(252)

	// 计算夏普比率
	// Sharpe = (R_p - R_f) / σ_p
	riskFreeRate := 0.045 // 4.5% 无风险利率
	sharpeRatio := (annualReturn - riskFreeRate) / annualVolatility
	if annualVolatility == 0 {
		sharpeRatio = 0
	}

	// 计算最大回撤
	maxDrawdown := s.calculateMaxDrawdown(prices)

	// 计算股息率 (简化计算: 假设基于SCHD/JEPQ的已知股息率)
	dividendYield := s.getEstimatedDividendYield(symbol)

	return &ETFHistoricalMetrics{
		Symbol:        symbol,
		AnnualReturn:  annualReturn,
		Volatility:    annualVolatility,
		SharpeRatio:   sharpeRatio,
		MaxDrawdown:   maxDrawdown,
		DividendYield: dividendYield,
		DataPoints:    len(data),
		StartDate:     data[0].Date,
		EndDate:       data[len(data)-1].Date,
	}, nil
}

// CalculateCorrelation 计算两个ETF的相关系数
func (s *PortfolioAnalyticsService) CalculateCorrelation(symbol1, symbol2 string, days int) (float64, error) {
	data1, err := s.GetETFHistoricalData(symbol1, days)
	if err != nil {
		return 0, err
	}

	data2, err := s.GetETFHistoricalData(symbol2, days)
	if err != nil {
		return 0, err
	}

	// 对齐日期
	returns1, returns2 := s.alignReturns(data1, data2)

	if len(returns1) < 10 {
		return 0, fmt.Errorf("对齐后的数据点不足")
	}

	// 计算相关系数
	mean1 := s.mean(returns1)
	mean2 := s.mean(returns2)

	numerator := 0.0
	denom1 := 0.0
	denom2 := 0.0

	for i := 0; i < len(returns1); i++ {
		diff1 := returns1[i] - mean1
		diff2 := returns2[i] - mean2
		numerator += diff1 * diff2
		denom1 += diff1 * diff1
		denom2 += diff2 * diff2
	}

	if denom1 == 0 || denom2 == 0 {
		return 0, nil
	}

	correlation := numerator / (math.Sqrt(denom1) * math.Sqrt(denom2))

	// 限制在 [-1, 1] 范围内
	if correlation > 1 {
		correlation = 1
	} else if correlation < -1 {
		correlation = -1
	}

	return correlation, nil
}

// AnalyzePortfolio 分析投资组合
func (s *PortfolioAnalyticsService) AnalyzePortfolio(
	portfolio map[string]float64,
	days int,
) (*PortfolioAnalytics, error) {
	// 获取每个ETF的指标
	etfMetrics := make(map[string]*ETFHistoricalMetrics)

	for symbol := range portfolio {
		metrics, err := s.CalculateETFMetrics(symbol, days)
		if err != nil {
			return nil, fmt.Errorf("计算ETF %s 指标失败: %w", symbol, err)
		}
		etfMetrics[symbol] = metrics
	}

	// 计算组合预期收益率
	// R_p = Σ w_i * R_i
	expectedReturn := 0.0
	for symbol, weight := range portfolio {
		expectedReturn += weight * etfMetrics[symbol].AnnualReturn
	}

	// 计算组合波动率 (考虑相关性)
	// σ_p² = Σ Σ w_i * w_j * σ_i * σ_j * ρ_ij
	portfolioVariance := 0.0
	symbols := make([]string, 0, len(portfolio))
	for s := range portfolio {
		symbols = append(symbols, s)
	}

	correlationMatrix := make(map[string]float64)

	for i, sym1 := range symbols {
		for j, sym2 := range symbols {
			w1 := portfolio[sym1]
			w2 := portfolio[sym2]
			sigma1 := etfMetrics[sym1].Volatility
			sigma2 := etfMetrics[sym2].Volatility

			var correlation float64
			if i == j {
				correlation = 1.0
			} else {
				corr, err := s.CalculateCorrelation(sym1, sym2, days)
				if err != nil {
					// 如果计算失败，使用默认相关系数
					corr = 0.7
				}
				correlation = corr
			}

			key := fmt.Sprintf("%s_%s", sym1, sym2)
			correlationMatrix[key] = correlation

			portfolioVariance += w1 * w2 * sigma1 * sigma2 * correlation
		}
	}

	portfolioVolatility := math.Sqrt(portfolioVariance)

	// 计算夏普比率
	riskFreeRate := 0.045
	sharpeRatio := (expectedReturn - riskFreeRate) / portfolioVolatility
	if portfolioVolatility == 0 {
		sharpeRatio = 0
	}

	// 计算组合最大回撤 (简化: 取各ETF最大回撤的加权平均)
	maxDrawdown := 0.0
	for symbol, weight := range portfolio {
		maxDrawdown += weight * etfMetrics[symbol].MaxDrawdown
	}

	return &PortfolioAnalytics{
		ExpectedReturn:    expectedReturn,
		Volatility:        portfolioVolatility,
		SharpeRatio:       sharpeRatio,
		MaxDrawdown:       maxDrawdown,
		CorrelationMatrix: correlationMatrix,
		ETFMetrics:        etfMetrics,
	}, nil
}

// 辅助函数

func (s *PortfolioAnalyticsService) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *PortfolioAnalyticsService) variance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}

func (s *PortfolioAnalyticsService) calculateMaxDrawdown(prices []decimal.Decimal) float64 {
	if len(prices) == 0 {
		return 0
	}

	maxDrawdown := 0.0
	peak, _ := prices[0].Float64()

	for _, priceDec := range prices {
		price, _ := priceDec.Float64()

		if price > peak {
			peak = price
		}

		drawdown := (peak - price) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

func (s *PortfolioAnalyticsService) getEstimatedDividendYield(symbol string) float64 {
	// 基于已知ETF的股息率估计
	yields := map[string]float64{
		"SCHD": 0.035,
		"JEPQ": 0.095,
		"QQQ":  0.006,
		"VTI":  0.015,
		"SPY":  0.013,
		"VYM":  0.028,
		"JEPI": 0.075,
		"BND":  0.030,
	}

	if y, ok := yields[symbol]; ok {
		return y
	}
	return 0.02 // 默认2%
}

func (s *PortfolioAnalyticsService) alignReturns(data1, data2 []models.ETFData) ([]float64, []float64) {
	// 创建日期到价格的映射
	priceMap1 := make(map[string]decimal.Decimal)
	for _, d := range data1 {
		priceMap1[d.Date.Format("2006-01-02")] = d.ClosePrice
	}

	priceMap2 := make(map[string]decimal.Decimal)
	for _, d := range data2 {
		priceMap2[d.Date.Format("2006-01-02")] = d.ClosePrice
	}

	// 找到共同日期
	commonDates := make([]string, 0)
	for date := range priceMap1 {
		if _, ok := priceMap2[date]; ok {
			commonDates = append(commonDates, date)
		}
	}

	// 排序日期
	sort.Strings(commonDates)

	// 提取对齐后的价格
	prices1 := make([]decimal.Decimal, len(commonDates))
	prices2 := make([]decimal.Decimal, len(commonDates))

	for i, date := range commonDates {
		prices1[i] = priceMap1[date]
		prices2[i] = priceMap2[date]
	}

	// 计算收益率
	returns1 := s.CalculateLogReturns(prices1)
	returns2 := s.CalculateLogReturns(prices2)

	return returns1, returns2
}

// CalculateVaR 计算风险价值 (参数法)
func (s *PortfolioAnalyticsService) CalculateVaR(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := s.mean(returns)
	std := math.Sqrt(s.variance(returns, mean))

	// 使用标准正态分布的分位数
	// 95% -> -1.645, 99% -> -2.326
	var zScore float64
	switch confidence {
	case 0.95:
		zScore = -1.645
	case 0.99:
		zScore = -2.326
	default:
		zScore = -1.645
	}

	// VaR = 均值 + Z * 标准差
	return mean + zScore*std
}

// CalculateCVaR 计算条件风险价值 (参数法)
func (s *PortfolioAnalyticsService) CalculateCVaR(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := s.mean(returns)
	std := math.Sqrt(s.variance(returns, mean))

	// CVaR = 均值 - std * φ(Z) / Φ(Z)
	// 其中 φ 是标准正态PDF, Φ 是标准正态CDF
	var zScore float64
	switch confidence {
	case 0.95:
		zScore = -1.645
	case 0.99:
		zScore = -2.326
	default:
		zScore = -1.645
	}

	// 标准正态分布在Z处的PDF值
	phi := (1.0 / math.Sqrt(2*math.Pi)) * math.Exp(-zScore*zScore/2)

	// 标准正态分布在Z处的CDF值 (近似)
	Phi := 1 - confidence

	cvar := mean - std*(phi/Phi)
	return cvar
}
