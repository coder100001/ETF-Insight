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
	Symbol         string                        `json:"symbol"`
	AnnualReturn   float64                       `json:"annual_return"`   // 年化收益率
	Volatility     float64                       `json:"volatility"`      // 年化波动率
	SharpeRatio    float64                       `json:"sharpe_ratio"`    // 夏普比率
	MaxDrawdown    float64                       `json:"max_drawdown"`    // 最大回撤
	DividendYield  float64                       `json:"dividend_yield"`  // 股息率
	DataPoints     int                           `json:"data_points"`     // 数据点数量
	StartDate      time.Time                     `json:"start_date"`      // 数据开始日期
	EndDate        time.Time                     `json:"end_date"`        // 数据结束日期
	RollingMetrics map[int]*RollingWindowMetrics `json:"rolling_metrics"` // 滚动窗口指标
	SortinoRatio   float64                       `json:"sortino_ratio"`   // 索提诺比率
	CalmarRatio    float64                       `json:"calmar_ratio"`    // 卡尔玛比率
	Skewness       float64                       `json:"skewness"`        // 偏度
	Kurtosis       float64                       `json:"kurtosis"`        // 峰度
	VaR95          float64                       `json:"var_95"`          // 95% VaR
	CVaR95         float64                       `json:"cvar_95"`         // 95% CVaR
}

// PortfolioAnalytics 组合分析指标
type PortfolioAnalytics struct {
	ExpectedReturn    float64                          `json:"expected_return"`    // 预期年化收益率
	Volatility        float64                          `json:"volatility"`         // 年化波动率
	SharpeRatio       float64                          `json:"sharpe_ratio"`       // 夏普比率
	SortinoRatio      float64                          `json:"sortino_ratio"`      // 索提诺比率
	CalmarRatio       float64                          `json:"calmar_ratio"`       // 卡尔玛比率
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

// SetDB 设置数据库连接(用于测试)
func (s *PortfolioAnalyticsService) SetDB(db *gorm.DB) {
	s.db = db
}

// GetETFHistoricalData 获取ETF历史价格数据
func (s *PortfolioAnalyticsService) GetETFHistoricalData(symbol string, days int) ([]models.ETFData, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

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

	// 计算索提诺比率
	sortinoRatio := s.CalculateSortinoRatio(returns, riskFreeRate)

	// 计算卡尔玛比率
	calmarRatio := s.CalculateCalmarRatio(annualReturn, maxDrawdown)

	// 计算偏度和峰度
	skewness := s.CalculateSkewness(returns)
	kurtosis := s.CalculateKurtosis(returns)

	// 计算VaR和CVaR
	var95 := s.CalculateVaR(returns, 0.95)
	cvar95 := s.CalculateCVaR(returns, 0.95)

	// 计算滚动窗口指标
	rollingMetrics := s.CalculateAllRollingWindows(returns, prices)

	return &ETFHistoricalMetrics{
		Symbol:         symbol,
		AnnualReturn:   annualReturn,
		Volatility:     annualVolatility,
		SharpeRatio:    sharpeRatio,
		MaxDrawdown:    maxDrawdown,
		DividendYield:  dividendYield,
		DataPoints:     len(data),
		StartDate:      data[0].Date,
		EndDate:        data[len(data)-1].Date,
		RollingMetrics: rollingMetrics,
		SortinoRatio:   sortinoRatio,
		CalmarRatio:    calmarRatio,
		Skewness:       skewness,
		Kurtosis:       kurtosis,
		VaR95:          var95,
		CVaR95:         cvar95,
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

	for i := range returns1 {
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
	// 获取每个ETF的指标和历史数据
	etfMetrics := make(map[string]*ETFHistoricalMetrics)
	etfData := make(map[string][]models.ETFData)

	for symbol := range portfolio {
		metrics, err := s.CalculateETFMetrics(symbol, days)
		if err != nil {
			return nil, fmt.Errorf("计算ETF %s 指标失败: %w", symbol, err)
		}
		etfMetrics[symbol] = metrics

		// 同时获取历史数据用于后续计算
		data, err := s.GetETFHistoricalData(symbol, days)
		if err == nil {
			etfData[symbol] = data
		}
	}

	// 计算组合预期收益率
	// R_p = Σ w_i * R_i
	expectedReturn := 0.0
	for symbol, weight := range portfolio {
		expectedReturn += weight * etfMetrics[symbol].AnnualReturn
	}

	// 计算组合波动率 (考虑相关性)
	// σ_p² = Σ_i w_i² * σ_i² + 2 * Σ_i<j w_i * w_j * σ_i * σ_j * ρ_ij
	portfolioVariance := 0.0
	symbols := make([]string, 0, len(portfolio))
	for s := range portfolio {
		symbols = append(symbols, s)
	}

	correlationMatrix := make(map[string]float64)

	// 第一遍: 计算方差项 (i = j)
	for _, sym := range symbols {
		w := portfolio[sym]
		sigma := etfMetrics[sym].Volatility
		// w_i² * σ_i²
		portfolioVariance += w * w * sigma * sigma
		// 对角线相关系数为1
		key := fmt.Sprintf("%s_%s", sym, sym)
		correlationMatrix[key] = 1.0
	}

	// 第二遍: 计算协方差项 (i < j)，乘以2因为矩阵对称
	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			sym1 := symbols[i]
			sym2 := symbols[j]
			w1 := portfolio[sym1]
			w2 := portfolio[sym2]
			sigma1 := etfMetrics[sym1].Volatility
			sigma2 := etfMetrics[sym2].Volatility

			corr, err := s.CalculateCorrelation(sym1, sym2, days)
			if err != nil {
				// 如果计算失败，使用默认相关系数
				corr = 0.7
			}

			// 存储相关系数矩阵 (对称)
			key1 := fmt.Sprintf("%s_%s", sym1, sym2)
			key2 := fmt.Sprintf("%s_%s", sym2, sym1)
			correlationMatrix[key1] = corr
			correlationMatrix[key2] = corr

			// 2 * w_i * w_j * σ_i * σ_j * ρ_ij
			portfolioVariance += 2 * w1 * w2 * sigma1 * sigma2 * corr
		}
	}

	portfolioVolatility := math.Sqrt(portfolioVariance)

	// 计算夏普比率
	riskFreeRate := 0.045
	sharpeRatio := (expectedReturn - riskFreeRate) / portfolioVolatility
	if portfolioVolatility == 0 {
		sharpeRatio = 0
	}

	// 计算组合最大回撤 (基于组合净值序列)
	maxDrawdown := s.calculatePortfolioMaxDrawdown(portfolio, etfMetrics, days)

	// 计算索提诺比率 (需要组合收益率序列)
	portfolioReturns := s.calculatePortfolioReturnsFromData(portfolio, etfData)
	sortinoRatio := s.CalculateSortinoRatio(portfolioReturns, riskFreeRate)

	// 计算卡尔玛比率
	calmarRatio := s.CalculateCalmarRatio(expectedReturn, maxDrawdown)

	return &PortfolioAnalytics{
		ExpectedReturn:    expectedReturn,
		Volatility:        portfolioVolatility,
		SharpeRatio:       sharpeRatio,
		SortinoRatio:      sortinoRatio,
		CalmarRatio:       calmarRatio,
		MaxDrawdown:       maxDrawdown,
		CorrelationMatrix: correlationMatrix,
		ETFMetrics:        etfMetrics,
	}, nil
}

// calculatePortfolioReturnsFromData 从已有数据计算组合收益率序列
func (s *PortfolioAnalyticsService) calculatePortfolioReturnsFromData(
	portfolio map[string]float64,
	etfData map[string][]models.ETFData,
) []float64 {
	// 获取所有ETF的共同日期
	commonDates := s.findCommonDates(etfData)
	if len(commonDates) < 2 {
		return []float64{}
	}

	// 计算每日组合价值
	portfolioValues := make([]float64, len(commonDates))

	for i, date := range commonDates {
		dailyValue := 0.0
		for symbol, weight := range portfolio {
			for _, d := range etfData[symbol] {
				if d.Date.Format("2006-01-02") == date {
					price, _ := d.ClosePrice.Float64()
					// 归一化价格
					if len(etfData[symbol]) > 0 {
						firstPrice, _ := etfData[symbol][0].ClosePrice.Float64()
						if firstPrice > 0 {
							return_rate := (price - firstPrice) / firstPrice
							dailyValue += weight * (1 + return_rate)
						}
					}
					break
				}
			}
		}
		portfolioValues[i] = dailyValue
	}

	// 计算对数收益率
	returns := make([]float64, len(portfolioValues)-1)
	for i := 1; i < len(portfolioValues); i++ {
		if portfolioValues[i-1] > 0 {
			returns[i-1] = math.Log(portfolioValues[i] / portfolioValues[i-1])
		}
	}

	return returns
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

// calculatePortfolioMaxDrawdown 计算组合最大回撤
// 基于组合净值序列计算，而不是简单加权
func (s *PortfolioAnalyticsService) calculatePortfolioMaxDrawdown(
	portfolio map[string]float64,
	etfMetrics map[string]*ETFHistoricalMetrics,
	days int,
) float64 {
	// 获取所有ETF的历史数据
	etfData := make(map[string][]models.ETFData)
	for symbol := range portfolio {
		data, err := s.GetETFHistoricalData(symbol, days)
		if err != nil {
			// 如果无法获取数据，使用ETF自身的最大回撤
			return etfMetrics[symbol].MaxDrawdown
		}
		etfData[symbol] = data
	}

	// 找到所有ETF的共同日期
	commonDates := s.findCommonDates(etfData)
	if len(commonDates) < 2 {
		// 数据不足，返回加权平均作为fallback
		weightedMDD := 0.0
		for symbol, weight := range portfolio {
			weightedMDD += weight * etfMetrics[symbol].MaxDrawdown
		}
		return weightedMDD
	}

	// 计算组合每日净值
	portfolioValues := make([]float64, len(commonDates))
	baseValue := 100.0 // 基准值

	for i, date := range commonDates {
		dailyValue := 0.0
		for symbol, weight := range portfolio {
			// 找到该日期对应的ETF数据
			for _, d := range etfData[symbol] {
				if d.Date.Format("2006-01-02") == date {
					price, _ := d.ClosePrice.Float64()
					// 归一化价格 (以第一天为基准)
					if i == 0 {
						dailyValue += weight * baseValue
					} else {
						// 计算相对于第一天的收益率
						firstPrice, _ := etfData[symbol][0].ClosePrice.Float64()
						if firstPrice > 0 {
							return_rate := (price - firstPrice) / firstPrice
							dailyValue += weight * baseValue * (1 + return_rate)
						}
					}
					break
				}
			}
		}
		portfolioValues[i] = dailyValue
	}

	// 计算最大回撤
	return s.calculateMaxDrawdownFromValues(portfolioValues)
}

// findCommonDates 找到所有ETF的共同交易日
func (s *PortfolioAnalyticsService) findCommonDates(etfData map[string][]models.ETFData) []string {
	if len(etfData) == 0 {
		return []string{}
	}

	// 获取第一个ETF的所有日期
	var firstSymbol string
	for symbol := range etfData {
		firstSymbol = symbol
		break
	}

	dateSet := make(map[string]bool)
	for _, d := range etfData[firstSymbol] {
		dateSet[d.Date.Format("2006-01-02")] = true
	}

	// 检查其他ETF的日期
	for symbol, data := range etfData {
		if symbol == firstSymbol {
			continue
		}
		currentDates := make(map[string]bool)
		for _, d := range data {
			currentDates[d.Date.Format("2006-01-02")] = true
		}

		// 只保留共同日期
		for date := range dateSet {
			if !currentDates[date] {
				delete(dateSet, date)
			}
		}
	}

	// 转换为排序后的切片
	dates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	return dates
}

// calculateMaxDrawdownFromValues 从净值序列计算最大回撤
func (s *PortfolioAnalyticsService) calculateMaxDrawdownFromValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	maxDrawdown := 0.0
	peak := values[0]

	for _, value := range values {
		if value > peak {
			peak = value
		}

		drawdown := (peak - value) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// CalculateSortinoRatio 计算索提诺比率
// Sortino Ratio = (R_p - R_f) / σ_d
// 其中 σ_d 是下行标准差 (只考虑负收益)
func (s *PortfolioAnalyticsService) CalculateSortinoRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// 计算平均收益率
	meanReturn := s.mean(returns)

	// 计算下行标准差 (只考虑低于目标收益率的偏差)
	// 目标收益率通常设为0或风险-free rate
	targetReturn := 0.0 // 使用0作为最小可接受收益率
	downsideDeviations := make([]float64, 0)

	for _, r := range returns {
		if r < targetReturn {
			downsideDeviations = append(downsideDeviations, (r-targetReturn)*(r-targetReturn))
		}
	}

	if len(downsideDeviations) == 0 {
		return 0 // 没有下行风险，无法计算
	}

	// 计算下行标准差
	downsideVariance := 0.0
	for _, d := range downsideDeviations {
		downsideVariance += d
	}
	downsideStd := math.Sqrt(downsideVariance / float64(len(returns)))

	if downsideStd == 0 {
		return 0
	}

	// 年化处理
	annualReturn := meanReturn * 252
	annualDownsideStd := downsideStd * math.Sqrt(252)
	annualRiskFreeRate := riskFreeRate

	return (annualReturn - annualRiskFreeRate) / annualDownsideStd
}

// CalculateCalmarRatio 计算卡尔玛比率
// Calmar Ratio = 年化收益率 / 最大回撤
func (s *PortfolioAnalyticsService) CalculateCalmarRatio(annualReturn, maxDrawdown float64) float64 {
	if maxDrawdown == 0 {
		return 0
	}
	return annualReturn / maxDrawdown
}

// CalculateDownsideDeviation 计算下行偏差
// 只考虑低于目标收益率的波动
func (s *PortfolioAnalyticsService) CalculateDownsideDeviation(returns []float64, targetReturn float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	downsideSum := 0.0
	count := 0

	for _, r := range returns {
		if r < targetReturn {
			downsideSum += (r - targetReturn) * (r - targetReturn)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return math.Sqrt(downsideSum / float64(len(returns)))
}

// CalculateSkewness 计算偏度
// 衡量收益分布的不对称性
func (s *PortfolioAnalyticsService) CalculateSkewness(returns []float64) float64 {
	n := float64(len(returns))
	if n < 3 {
		return 0
	}

	mean := s.mean(returns)
	std := math.Sqrt(s.variance(returns, mean))

	if std == 0 {
		return 0
	}

	sumCubed := 0.0
	for _, r := range returns {
		sumCubed += math.Pow((r-mean)/std, 3)
	}

	// 样本偏度修正
	skewness := (n / ((n - 1) * (n - 2))) * sumCubed
	return skewness
}

// CalculateKurtosis 计算峰度
// 衡量收益分布的尾部厚度
func (s *PortfolioAnalyticsService) CalculateKurtosis(returns []float64) float64 {
	n := float64(len(returns))
	if n < 4 {
		return 0
	}

	mean := s.mean(returns)
	std := math.Sqrt(s.variance(returns, mean))

	if std == 0 {
		return 0
	}

	sumFourth := 0.0
	for _, r := range returns {
		sumFourth += math.Pow((r-mean)/std, 4)
	}

	//  excess kurtosis (减去3得到超额峰度)
	kurtosis := sumFourth / n
	return kurtosis - 3
}

// RollingWindowMetrics 滚动窗口指标
type RollingWindowMetrics struct {
	WindowDays   int     `json:"window_days"`   // 窗口天数
	AnnualReturn float64 `json:"annual_return"` // 年化收益率
	Volatility   float64 `json:"volatility"`    // 年化波动率
	SharpeRatio  float64 `json:"sharpe_ratio"`  // 夏普比率
	MaxDrawdown  float64 `json:"max_drawdown"`  // 最大回撤
	CalmarRatio  float64 `json:"calmar_ratio"`  // 卡尔玛比率
	SortinoRatio float64 `json:"sortino_ratio"` // 索提诺比率
	VaR95        float64 `json:"var_95"`        // 95% VaR
	WinRate      float64 `json:"win_rate"`      // 胜率
	AvgGain      float64 `json:"avg_gain"`      // 平均盈利
	AvgLoss      float64 `json:"avg_loss"`      // 平均亏损
	ProfitFactor float64 `json:"profit_factor"` // 盈亏比
}

// CalculateRollingWindowMetrics 计算滚动窗口指标
// 支持 30日/60日/90日/180日/1年 等窗口
func (s *PortfolioAnalyticsService) CalculateRollingWindowMetrics(
	returns []float64,
	prices []decimal.Decimal,
	windowDays int,
) *RollingWindowMetrics {
	if len(returns) < windowDays || windowDays < 30 {
		return nil
	}

	// 获取最近窗口期的数据
	recentReturns := returns[len(returns)-windowDays:]
	recentPrices := prices[len(prices)-windowDays-1:]

	// 计算年化收益率
	totalReturn := 0.0
	for _, r := range recentReturns {
		totalReturn += r
	}
	annualReturn := totalReturn * (252.0 / float64(windowDays))

	// 计算年化波动率
	mean := s.mean(recentReturns)
	variance := s.variance(recentReturns, mean)
	volatility := math.Sqrt(variance) * math.Sqrt(252)

	// 计算夏普比率
	riskFreeRate := 0.045
	sharpeRatio := 0.0
	if volatility > 0 {
		sharpeRatio = (annualReturn - riskFreeRate) / volatility
	}

	// 计算最大回撤
	maxDrawdown := s.calculateMaxDrawdown(recentPrices)

	// 计算卡尔玛比率
	calmarRatio := 0.0
	if maxDrawdown > 0 {
		calmarRatio = annualReturn / maxDrawdown
	}

	// 计算索提诺比率
	sortinoRatio := s.CalculateSortinoRatio(recentReturns, riskFreeRate)

	// 计算VaR 95%
	var95 := s.CalculateVaR(recentReturns, 0.95)

	// 计算胜率、平均盈亏、盈亏比
	winCount := 0
	lossCount := 0
	gainSum := 0.0
	lossSum := 0.0

	for _, r := range recentReturns {
		if r > 0 {
			winCount++
			gainSum += r
		} else if r < 0 {
			lossCount++
			lossSum += math.Abs(r)
		}
	}

	winRate := 0.0
	if len(recentReturns) > 0 {
		winRate = float64(winCount) / float64(len(recentReturns))
	}

	avgGain := 0.0
	if winCount > 0 {
		avgGain = gainSum / float64(winCount)
	}

	avgLoss := 0.0
	if lossCount > 0 {
		avgLoss = lossSum / float64(lossCount)
	}

	profitFactor := 0.0
	if lossSum > 0 {
		profitFactor = gainSum / lossSum
	}

	return &RollingWindowMetrics{
		WindowDays:   windowDays,
		AnnualReturn: annualReturn,
		Volatility:   volatility,
		SharpeRatio:  sharpeRatio,
		MaxDrawdown:  maxDrawdown,
		CalmarRatio:  calmarRatio,
		SortinoRatio: sortinoRatio,
		VaR95:        var95,
		WinRate:      winRate,
		AvgGain:      avgGain,
		AvgLoss:      avgLoss,
		ProfitFactor: profitFactor,
	}
}

// CalculateAllRollingWindows 计算所有常用滚动窗口指标
// 返回 30日、60日、90日、180日、1年的指标
func (s *PortfolioAnalyticsService) CalculateAllRollingWindows(
	returns []float64,
	prices []decimal.Decimal,
) map[int]*RollingWindowMetrics {
	windows := []int{30, 60, 90, 180, 252} // 252个交易日≈1年
	result := make(map[int]*RollingWindowMetrics)

	for _, window := range windows {
		if len(returns) >= window {
			metrics := s.CalculateRollingWindowMetrics(returns, prices, window)
			if metrics != nil {
				result[window] = metrics
			}
		}
	}

	return result
}
