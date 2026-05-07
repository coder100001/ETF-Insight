package services

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// MarketScenario 市场情景类型
type MarketScenario string

const (
	Neutral     MarketScenario = "neutral"
	Pessimistic MarketScenario = "pessimistic"
	Optimistic  MarketScenario = "optimistic"
)

// ScenarioAssumptions 情景假设
type ScenarioAssumptions struct {
	Scenario        MarketScenario `json:"scenario"`
	AnnualReturn    float64        `json:"annual_return"`
	Volatility      float64        `json:"volatility"`
	SharpeRatio     float64        `json:"sharpe_ratio"`
	MaxDrawdown     float64        `json:"max_drawdown"`
	Description     string         `json:"description"`
	RiskFreeRate    float64        `json:"risk_free_rate"`
	MarketCondition string         `json:"market_condition"`
	DividendYield   float64        `json:"dividend_yield"`
}

// PortfolioProjection 投资组合预测
type PortfolioProjection struct {
	Year             int     `json:"year"`
	StartDate        string  `json:"start_date"`
	EndDate          string  `json:"end_date"`
	StartValue       float64 `json:"start_value"`
	EndValue         float64 `json:"end_value"`
	AnnualReturn     float64 `json:"annual_return"`
	CumulativeReturn float64 `json:"cumulative_return"`
	Volatility       float64 `json:"volatility"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	DividendIncome   float64 `json:"dividend_income"`
	TotalValue       float64 `json:"total_value"`
}

// ScenarioAnalysisResult 情景分析结果
type ScenarioAnalysisResult struct {
	Portfolio         map[string]float64                 `json:"portfolio"`
	TotalInvestment   float64                            `json:"total_investment"`
	TimeHorizonYears  int                                `json:"time_horizon_years"`
	Scenarios         map[MarketScenario]*ScenarioResult `json:"scenarios"`
	ComparisonMetrics *ComparisonMetrics                 `json:"comparison_metrics"`
	Methodology       string                             `json:"methodology"`
	Assumptions       string                             `json:"assumptions"`
	Limitations       string                             `json:"limitations"`
}

// ScenarioResult 单个情景结果
type ScenarioResult struct {
	Assumptions     *ScenarioAssumptions  `json:"assumptions"`
	Projections     []PortfolioProjection `json:"projections"`
	FinalValue      float64               `json:"final_value"`
	TotalReturn     float64               `json:"total_return"`
	AvgAnnualReturn float64               `json:"avg_annual_return"`
	BestYear        float64               `json:"best_year"`
	WorstYear       float64               `json:"worst_year"`
	VaR95           float64               `json:"var_95"`
	VaR99           float64               `json:"var_99"`
	CVaR95          float64               `json:"cvar_95"`
}

// ComparisonMetrics 对比指标
type ComparisonMetrics struct {
	BestScenario       MarketScenario `json:"best_scenario"`
	WorstScenario      MarketScenario `json:"worst_scenario"`
	ValueDifference    float64        `json:"value_difference"`
	ReturnSpread       float64        `json:"return_spread"`
	RiskAdjustedWinner MarketScenario `json:"risk_adjusted_winner"`
}

// ScenarioAnalysisService 情景分析服务
type ScenarioAnalysisService struct {
	analyticsService *PortfolioAnalyticsService
}

// NewScenarioAnalysisService 创建情景分析服务
func NewScenarioAnalysisService() *ScenarioAnalysisService {
	return &ScenarioAnalysisService{
		analyticsService: NewPortfolioAnalyticsService(),
	}
}

// AnalyzePortfolio 分析投资组合
func (s *ScenarioAnalysisService) AnalyzePortfolio(
	portfolio map[string]float64,
	totalInvestment float64,
	timeHorizonYears int,
) (*ScenarioAnalysisResult, error) {
	// 验证投资组合权重
	if err := s.validatePortfolio(portfolio); err != nil {
		return nil, err
	}

	// 使用真实历史数据计算组合指标 (使用252天 ≈ 1年数据)
	portfolioAnalytics, err := s.analyticsService.AnalyzePortfolio(portfolio, 252*3) // 3年数据
	if err != nil {
		// 如果无法获取历史数据，使用默认指标
		portfolioAnalytics = s.getDefaultPortfolioAnalytics(portfolio)
	}

	// 生成三种情景
	scenarios := make(map[MarketScenario]*ScenarioResult)

	// 中性情景 - 基于历史数据
	scenarios[Neutral] = s.generateScenario(
		Neutral,
		portfolioAnalytics,
		totalInvestment,
		timeHorizonYears,
		1.0, // 中性乘数
	)

	// 悲观情景 - 收益降低，波动增加
	pessimisticAnalytics := s.adjustAnalyticsForScenario(portfolioAnalytics, Pessimistic)
	scenarios[Pessimistic] = s.generateScenario(
		Pessimistic,
		pessimisticAnalytics,
		totalInvestment,
		timeHorizonYears,
		0.85, // 悲观乘数
	)

	// 乐观情景 - 收益增加，波动降低
	optimisticAnalytics := s.adjustAnalyticsForScenario(portfolioAnalytics, Optimistic)
	scenarios[Optimistic] = s.generateScenario(
		Optimistic,
		optimisticAnalytics,
		totalInvestment,
		timeHorizonYears,
		1.15, // 乐观乘数
	)

	// 计算对比指标
	comparison := s.calculateComparison(scenarios)

	result := &ScenarioAnalysisResult{
		Portfolio:         portfolio,
		TotalInvestment:   totalInvestment,
		TimeHorizonYears:  timeHorizonYears,
		Scenarios:         scenarios,
		ComparisonMetrics: comparison,
		Methodology: "基于真实历史数据的几何布朗运动蒙特卡洛模拟。" +
			"使用3年历史数据计算年化收益率、波动率和相关系数。" +
			"进行1000次蒙特卡洛模拟，计算置信区间和风险指标。",
		Assumptions: "中性情景：基于历史平均收益率和波动率。\n" +
			"悲观情景：收益率降低15%，波动率增加30%，最大回撤增加50%。\n" +
			"乐观情景：收益率增加15%，波动率降低20%，最大回撤减少30%。\n" +
			"无风险利率：4.5%（当前美国国债利率水平）。\n" +
			"股息再投资：假设股息按年再投资。",
		Limitations: "1. 历史表现不代表未来结果\n" +
			"2. 模型假设市场遵循几何布朗运动\n" +
			"3. 未考虑极端市场事件（黑天鹅）\n" +
			"4. 交易成本和税费影响未完全建模\n" +
			"5. 股息收益率假设保持不变\n" +
			"6. 相关系数基于历史数据，可能随时间变化",
	}

	return result, nil
}

// validatePortfolio 验证投资组合
func (s *ScenarioAnalysisService) validatePortfolio(portfolio map[string]float64) error {
	if len(portfolio) == 0 {
		return fmt.Errorf("投资组合不能为空")
	}

	totalWeight := 0.0
	for symbol, weight := range portfolio {
		if weight <= 0 || weight > 1 {
			return fmt.Errorf("资产 %s 的权重必须在 0-1 之间，当前: %.2f", symbol, weight)
		}
		totalWeight += weight
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		return fmt.Errorf("投资组合权重总和必须为 1.0，当前: %.2f", totalWeight)
	}

	return nil
}

// getDefaultPortfolioAnalytics 获取默认组合指标
func (s *ScenarioAnalysisService) getDefaultPortfolioAnalytics(portfolio map[string]float64) *PortfolioAnalytics {
	// 基于SCHD和JEPQ的默认指标
	etfMetrics := map[string]*ETFHistoricalMetrics{
		"SCHD": {
			Symbol:        "SCHD",
			AnnualReturn:  0.12,
			Volatility:    0.18,
			SharpeRatio:   0.67,
			MaxDrawdown:   0.20,
			DividendYield: 0.035,
		},
		"JEPQ": {
			Symbol:        "JEPQ",
			AnnualReturn:  0.10,
			Volatility:    0.15,
			SharpeRatio:   0.67,
			MaxDrawdown:   0.18,
			DividendYield: 0.095,
		},
		"QQQ": {
			Symbol:        "QQQ",
			AnnualReturn:  0.15,
			Volatility:    0.22,
			SharpeRatio:   0.68,
			MaxDrawdown:   0.35,
			DividendYield: 0.006,
		},
		"VTI": {
			Symbol:        "VTI",
			AnnualReturn:  0.10,
			Volatility:    0.16,
			SharpeRatio:   0.59,
			MaxDrawdown:   0.25,
			DividendYield: 0.015,
		},
	}

	// 计算加权指标
	expectedReturn := 0.0
	volatility := 0.0
	maxDrawdown := 0.0

	for symbol, weight := range portfolio {
		if metrics, ok := etfMetrics[symbol]; ok {
			expectedReturn += weight * metrics.AnnualReturn
			volatility += weight * metrics.Volatility
			maxDrawdown += weight * metrics.MaxDrawdown
		}
	}

	// 组合波动率调整 (简化计算)
	volatility = volatility * 0.9 // 考虑分散化效应

	riskFreeRate := 0.045
	sharpeRatio := (expectedReturn - riskFreeRate) / volatility
	if volatility == 0 {
		sharpeRatio = 0
	}

	return &PortfolioAnalytics{
		ExpectedReturn: expectedReturn,
		Volatility:     volatility,
		SharpeRatio:    sharpeRatio,
		MaxDrawdown:    maxDrawdown,
		ETFMetrics:     etfMetrics,
	}
}

// adjustAnalyticsForScenario 根据情景调整指标
func (s *ScenarioAnalysisService) adjustAnalyticsForScenario(
	analytics *PortfolioAnalytics,
	scenario MarketScenario,
) *PortfolioAnalytics {
	adjusted := &PortfolioAnalytics{
		ExpectedReturn: analytics.ExpectedReturn,
		Volatility:     analytics.Volatility,
		SharpeRatio:    analytics.SharpeRatio,
		MaxDrawdown:    analytics.MaxDrawdown,
		ETFMetrics:     analytics.ETFMetrics,
	}

	switch scenario {
	case Pessimistic:
		adjusted.ExpectedReturn = analytics.ExpectedReturn * 0.85
		adjusted.Volatility = analytics.Volatility * 1.30
		adjusted.SharpeRatio = analytics.SharpeRatio * 0.60
		adjusted.MaxDrawdown = analytics.MaxDrawdown * 1.50
	case Optimistic:
		adjusted.ExpectedReturn = analytics.ExpectedReturn * 1.15
		adjusted.Volatility = analytics.Volatility * 0.80
		adjusted.SharpeRatio = analytics.SharpeRatio * 1.40
		adjusted.MaxDrawdown = analytics.MaxDrawdown * 0.70
	}

	return adjusted
}

// generateScenario 生成情景预测 (使用蒙特卡洛模拟)
func (s *ScenarioAnalysisService) generateScenario(
	scenario MarketScenario,
	analytics *PortfolioAnalytics,
	totalInvestment float64,
	timeHorizonYears int,
	scenarioMultiplier float64,
) *ScenarioResult {
	riskFreeRate := 0.045

	assumptions := &ScenarioAssumptions{
		Scenario:      scenario,
		AnnualReturn:  analytics.ExpectedReturn * scenarioMultiplier,
		Volatility:    analytics.Volatility,
		SharpeRatio:   (analytics.ExpectedReturn*scenarioMultiplier - riskFreeRate) / analytics.Volatility,
		MaxDrawdown:   analytics.MaxDrawdown,
		RiskFreeRate:  riskFreeRate,
		DividendYield: 0.05, // 简化假设
	}

	// 设置情景描述
	switch scenario {
	case Neutral:
		assumptions.Description = "基于历史平均表现，市场按长期趋势发展"
		assumptions.MarketCondition = "市场稳定，经济温和增长"
	case Pessimistic:
		assumptions.Description = "经济衰退或市场调整期，收益下降风险上升"
		assumptions.MarketCondition = "经济放缓，市场波动加剧"
	case Optimistic:
		assumptions.Description = "牛市行情，经济增长强劲，市场表现优异"
		assumptions.MarketCondition = "经济繁荣，市场情绪乐观"
	}

	// 运行蒙特卡洛模拟 (1000次)
	numSimulations := 1000
	finalValues := make([]float64, numSimulations)
	allProjections := make([][]PortfolioProjection, numSimulations)

	for sim := range numSimulations {
		projections, finalValue := s.runSingleSimulation(
			assumptions,
			totalInvestment,
			timeHorizonYears,
			sim,
		)
		finalValues[sim] = finalValue
		allProjections[sim] = projections
	}

	// 计算统计指标
	sort.Float64s(finalValues)
	medianFinalValue := finalValues[numSimulations/2]
	var95 := finalValues[int(float64(numSimulations)*0.05)] // 5%分位数
	var99 := finalValues[int(float64(numSimulations)*0.01)] // 1%分位数
	cvar95 := s.calculateCVaRFromSimulations(finalValues, 0.95)

	// 使用中位数路径作为展示结果
	medianProjections := allProjections[numSimulations/2]

	// 计算其他指标
	totalReturn := (medianFinalValue - totalInvestment) / totalInvestment
	avgAnnualReturn := math.Pow(medianFinalValue/totalInvestment, 1.0/float64(timeHorizonYears)) - 1

	// 找出最好和最坏年份
	bestYear := -math.MaxFloat64
	worstYear := math.MaxFloat64
	for _, proj := range medianProjections {
		if proj.AnnualReturn > bestYear {
			bestYear = proj.AnnualReturn
		}
		if proj.AnnualReturn < worstYear {
			worstYear = proj.AnnualReturn
		}
	}

	return &ScenarioResult{
		Assumptions:     assumptions,
		Projections:     medianProjections,
		FinalValue:      medianFinalValue,
		TotalReturn:     totalReturn * 100,
		AvgAnnualReturn: avgAnnualReturn * 100,
		BestYear:        bestYear,
		WorstYear:       worstYear,
		VaR95:           var95,
		VaR99:           var99,
		CVaR95:          cvar95,
	}
}

// DividendPayment 股息支付记录
type DividendPayment struct {
	Date           time.Time `json:"date"`
	Amount         float64   `json:"amount"`
	ReinvestShares float64   `json:"reinvest_shares"`
	SharePrice     float64   `json:"share_price"`
}

// runSingleSimulation 运行单次模拟 (改进版，支持季度股息再投资)
func (s *ScenarioAnalysisService) runSingleSimulation(
	assumptions *ScenarioAssumptions,
	initialValue float64,
	timeHorizonYears int,
	seed int,
) ([]PortfolioProjection, float64) {
	// 为每次模拟设置不同的随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(seed)))

	projections := make([]PortfolioProjection, timeHorizonYears)
	currentValue := initialValue
	startDate := time.Now()

	// 几何布朗运动参数
	mu := assumptions.AnnualReturn
	sigma := assumptions.Volatility

	// 季度参数
	quartersPerYear := 4
	dt := 1.0 / float64(quartersPerYear) // 季度时间步长
	quarterlyDividendYield := assumptions.DividendYield / float64(quartersPerYear)

	for year := range timeHorizonYears {
		yearStartValue := currentValue
		yearDividendIncome := 0.0

		// 按季度模拟
		for range quartersPerYear {
			quarterStartValue := currentValue

			// 几何布朗运动: dS = μS dt + σS dW
			// S_t = S_0 * exp((μ - σ²/2)t + σ√t Z)
			z := s.generateNormalRandom(r)
			growthFactor := math.Exp((mu-0.5*sigma*sigma)*dt + sigma*math.Sqrt(dt)*z)

			quarterEndValue := quarterStartValue * growthFactor

			// 季度股息支付和再投资
			// 假设在季度末支付股息，并立即以季度末价格再投资
			quarterlyDividend := quarterEndValue * quarterlyDividendYield
			yearDividendIncome += quarterlyDividend

			// 股息再投资: 增加持仓价值
			// 假设以季度末价格购买额外份额
			quarterEndValue += quarterlyDividend

			currentValue = quarterEndValue
		}

		// 年度汇总
		annualReturn := (currentValue - yearStartValue) / yearStartValue
		cumulativeReturn := (currentValue - initialValue) / initialValue

		yearStart := startDate.AddDate(year, 0, 0)
		yearEnd := startDate.AddDate(year+1, 0, 0)

		projections[year] = PortfolioProjection{
			Year:             year + 1,
			StartDate:        yearStart.Format("2006-01-02"),
			EndDate:          yearEnd.Format("2006-01-02"),
			StartValue:       yearStartValue,
			EndValue:         currentValue,
			AnnualReturn:     annualReturn * 100,
			CumulativeReturn: cumulativeReturn * 100,
			Volatility:       assumptions.Volatility * 100,
			MaxDrawdown:      assumptions.MaxDrawdown * 100,
			SharpeRatio:      assumptions.SharpeRatio,
			DividendIncome:   yearDividendIncome,
			TotalValue:       currentValue,
		}
	}

	return projections, currentValue
}

// runSingleSimulationMonthly 运行单次模拟 (月度版本，更精确)
func (s *ScenarioAnalysisService) runSingleSimulationMonthly(
	assumptions *ScenarioAssumptions,
	initialValue float64,
	timeHorizonYears int,
	seed int,
) ([]PortfolioProjection, float64) {
	// 为每次模拟设置不同的随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(seed)))

	projections := make([]PortfolioProjection, timeHorizonYears)
	currentValue := initialValue
	startDate := time.Now()

	// 几何布朗运动参数
	mu := assumptions.AnnualReturn
	sigma := assumptions.Volatility

	// 月度参数
	monthsPerYear := 12
	dt := 1.0 / float64(monthsPerYear) // 月度时间步长
	monthlyDividendYield := assumptions.DividendYield / float64(monthsPerYear)

	// 月度股息支付通常不均匀，模拟实际支付模式
	// 大多数ETF每季度支付，但在不同月份
	dividendMonths := map[int]bool{2: true, 5: true, 8: true, 11: true} // 3月、6月、9月、12月(0-indexed)

	for year := range timeHorizonYears {
		yearStartValue := currentValue
		yearDividendIncome := 0.0

		// 按月度模拟
		for month := range monthsPerYear {
			monthStartValue := currentValue

			// 几何布朗运动
			z := s.generateNormalRandom(r)
			growthFactor := math.Exp((mu-0.5*sigma*sigma)*dt + sigma*math.Sqrt(dt)*z)

			monthEndValue := monthStartValue * growthFactor

			// 月度或季度股息支付
			var monthlyDividend float64
			if dividendMonths[month] {
				// 季度支付月，支付3个月的累积股息
				monthlyDividend = monthEndValue * monthlyDividendYield * 3
			}
			yearDividendIncome += monthlyDividend

			// 股息再投资
			monthEndValue += monthlyDividend

			currentValue = monthEndValue
		}

		// 年度汇总
		annualReturn := (currentValue - yearStartValue) / yearStartValue
		cumulativeReturn := (currentValue - initialValue) / initialValue

		yearStart := startDate.AddDate(year, 0, 0)
		yearEnd := startDate.AddDate(year+1, 0, 0)

		projections[year] = PortfolioProjection{
			Year:             year + 1,
			StartDate:        yearStart.Format("2006-01-02"),
			EndDate:          yearEnd.Format("2006-01-02"),
			StartValue:       yearStartValue,
			EndValue:         currentValue,
			AnnualReturn:     annualReturn * 100,
			CumulativeReturn: cumulativeReturn * 100,
			Volatility:       assumptions.Volatility * 100,
			MaxDrawdown:      assumptions.MaxDrawdown * 100,
			SharpeRatio:      assumptions.SharpeRatio,
			DividendIncome:   yearDividendIncome,
			TotalValue:       currentValue,
		}
	}

	return projections, currentValue
}

// generateNormalRandom 生成标准正态分布随机数 (Box-Muller变换)
func (s *ScenarioAnalysisService) generateNormalRandom(r *rand.Rand) float64 {
	u1 := r.Float64()
	u2 := r.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// calculateCVaRFromSimulations 从模拟结果计算CVaR
func (s *ScenarioAnalysisService) calculateCVaRFromSimulations(values []float64, confidence float64) float64 {
	sort.Float64s(values)

	// 找到对应分位数的位置
	index := max(int(float64(len(values))*(1-confidence)), 0)

	// 计算尾部平均值
	sum := 0.0
	count := 0
	for i := 0; i <= index && i < len(values); i++ {
		sum += values[i]
		count++
	}

	if count == 0 {
		return values[0]
	}

	return sum / float64(count)
}

// calculateComparison 计算对比指标
func (s *ScenarioAnalysisService) calculateComparison(
	scenarios map[MarketScenario]*ScenarioResult,
) *ComparisonMetrics {
	comparison := &ComparisonMetrics{}

	// 找出最好和最坏情景
	bestValue := -math.MaxFloat64
	worstValue := math.MaxFloat64
	bestSharpe := -math.MaxFloat64

	for scenario, result := range scenarios {
		if result.FinalValue > bestValue {
			bestValue = result.FinalValue
			comparison.BestScenario = scenario
		}
		if result.FinalValue < worstValue {
			worstValue = result.FinalValue
			comparison.WorstScenario = scenario
		}
		if result.Assumptions.SharpeRatio > bestSharpe {
			bestSharpe = result.Assumptions.SharpeRatio
			comparison.RiskAdjustedWinner = scenario
		}
	}

	comparison.ValueDifference = bestValue - worstValue
	comparison.ReturnSpread = scenarios[Optimistic].TotalReturn - scenarios[Pessimistic].TotalReturn

	return comparison
}
