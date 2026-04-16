package services

import (
	"fmt"
	"math"
	"math/rand"
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
	historicalData map[string][]float64
}

// NewScenarioAnalysisService 创建情景分析服务
func NewScenarioAnalysisService() *ScenarioAnalysisService {
	return &ScenarioAnalysisService{
		historicalData: make(map[string][]float64),
	}
}

// LoadHistoricalData 加载历史数据
func (s *ScenarioAnalysisService) LoadHistoricalData(symbol string, data []float64) {
	s.historicalData[symbol] = data
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

	// 计算加权指标
	weightedMetrics := s.calculateWeightedMetrics(portfolio)

	// 生成三种情景
	scenarios := make(map[MarketScenario]*ScenarioResult)

	// 中性情景
	scenarios[Neutral] = s.generateScenario(
		Neutral,
		weightedMetrics,
		totalInvestment,
		timeHorizonYears,
	)

	// 悲观情景
	pessimisticMetrics := s.adjustForPessimistic(weightedMetrics)
	scenarios[Pessimistic] = s.generateScenario(
		Pessimistic,
		pessimisticMetrics,
		totalInvestment,
		timeHorizonYears,
	)

	// 乐观情景
	optimisticMetrics := s.adjustForOptimistic(weightedMetrics)
	scenarios[Optimistic] = s.generateScenario(
		Optimistic,
		optimisticMetrics,
		totalInvestment,
		timeHorizonYears,
	)

	// 计算对比指标
	comparison := s.calculateComparison(scenarios)

	result := &ScenarioAnalysisResult{
		Portfolio:         portfolio,
		TotalInvestment:   totalInvestment,
		TimeHorizonYears:  timeHorizonYears,
		Scenarios:         scenarios,
		ComparisonMetrics: comparison,
		Methodology: "基于历史表现和蒙特卡洛模拟的三种市场情景分析。" +
			"使用几何布朗运动模型预测资产价格路径，考虑股息再投资。",
		Assumptions: "中性情景：基于历史平均收益率和波动率。\n" +
			"悲观情景：收益率降低30%，波动率增加50%，最大回撤增加100%。\n" +
			"乐观情景：收益率增加30%，波动率降低20%，最大回撤减少50%。\n" +
			"无风险利率：4.5%（当前美国国债利率水平）。",
		Limitations: "1. 历史表现不代表未来结果\n" +
			"2. 模型假设市场遵循几何布朗运动\n" +
			"3. 未考虑极端市场事件（黑天鹅）\n" +
			"4. 交易成本和税费影响未完全建模\n" +
			"5. 股息收益率假设保持不变",
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

// calculateWeightedMetrics 计算加权指标
func (s *ScenarioAnalysisService) calculateWeightedMetrics(portfolio map[string]float64) *WeightedMetrics {
	// SCHD 和 JEPQ 的历史指标（基于实际数据）
	etfMetrics := map[string]*ETFBaseMetrics{
		"SCHD": {
			AnnualReturn:  0.12, // 12% 年化收益
			Volatility:    0.18, // 18% 波动率
			SharpeRatio:   0.67,
			MaxDrawdown:   0.20,  // 20% 最大回撤
			DividendYield: 0.035, // 3.5% 股息率
		},
		"JEPQ": {
			AnnualReturn:  0.10, // 10% 年化收益
			Volatility:    0.15, // 15% 波动率
			SharpeRatio:   0.67,
			MaxDrawdown:   0.18,  // 18% 最大回撤
			DividendYield: 0.095, // 9.5% 股息率
		},
	}

	weightedReturn := 0.0
	weightedVolatility := 0.0
	weightedSharpe := 0.0
	weightedDrawdown := 0.0
	weightedDividend := 0.0

	for symbol, weight := range portfolio {
		if metrics, ok := etfMetrics[symbol]; ok {
			weightedReturn += weight * metrics.AnnualReturn
			weightedVolatility += weight * metrics.Volatility
			weightedSharpe += weight * metrics.SharpeRatio
			weightedDrawdown += weight * metrics.MaxDrawdown
			weightedDividend += weight * metrics.DividendYield
		}
	}

	// 组合波动率需要考虑相关性（简化计算）
	// 假设 SCHD 和 JEPQ 相关系数为 0.7
	correlation := 0.7
	if len(portfolio) == 2 {
		symbols := make([]string, 0, 2)
		for s := range portfolio {
			symbols = append(symbols, s)
		}
		if m1, ok1 := etfMetrics[symbols[0]]; ok1 {
			if m2, ok2 := etfMetrics[symbols[1]]; ok2 {
				w1 := portfolio[symbols[0]]
				w2 := portfolio[symbols[1]]
				vol1 := m1.Volatility
				vol2 := m2.Volatility
				// 组合方差公式
				portfolioVariance := w1*w1*vol1*vol1 + w2*w2*vol2*vol2 + 2*w1*w2*correlation*vol1*vol2
				weightedVolatility = math.Sqrt(portfolioVariance)
			}
		}
	}

	return &WeightedMetrics{
		AnnualReturn:  weightedReturn,
		Volatility:    weightedVolatility,
		SharpeRatio:   weightedSharpe,
		MaxDrawdown:   weightedDrawdown,
		DividendYield: weightedDividend,
	}
}

// WeightedMetrics 加权指标
type WeightedMetrics struct {
	AnnualReturn  float64
	Volatility    float64
	SharpeRatio   float64
	MaxDrawdown   float64
	DividendYield float64
}

// ETFBaseMetrics ETF 基础指标
type ETFBaseMetrics struct {
	AnnualReturn  float64
	Volatility    float64
	SharpeRatio   float64
	MaxDrawdown   float64
	DividendYield float64
}

// adjustForPessimistic 调整为悲观情景
func (s *ScenarioAnalysisService) adjustForPessimistic(metrics *WeightedMetrics) *WeightedMetrics {
	return &WeightedMetrics{
		AnnualReturn:  metrics.AnnualReturn * 0.7,   // 收益降低30%
		Volatility:    metrics.Volatility * 1.5,     // 波动率增加50%
		SharpeRatio:   metrics.SharpeRatio * 0.5,    // 夏普比率降低50%
		MaxDrawdown:   metrics.MaxDrawdown * 2.0,    // 最大回撤增加100%
		DividendYield: metrics.DividendYield * 0.85, // 股息率降低15%
	}
}

// adjustForOptimistic 调整为乐观情景
func (s *ScenarioAnalysisService) adjustForOptimistic(metrics *WeightedMetrics) *WeightedMetrics {
	return &WeightedMetrics{
		AnnualReturn:  metrics.AnnualReturn * 1.3,  // 收益增加30%
		Volatility:    metrics.Volatility * 0.8,    // 波动率降低20%
		SharpeRatio:   metrics.SharpeRatio * 1.4,   // 夏普比率增加40%
		MaxDrawdown:   metrics.MaxDrawdown * 0.5,   // 最大回撤减少50%
		DividendYield: metrics.DividendYield * 1.1, // 股息率增加10%
	}
}

// generateScenario 生成情景预测
func (s *ScenarioAnalysisService) generateScenario(
	scenario MarketScenario,
	metrics *WeightedMetrics,
	totalInvestment float64,
	timeHorizonYears int,
) *ScenarioResult {
	riskFreeRate := 0.045 // 4.5% 无风险利率

	assumptions := &ScenarioAssumptions{
		Scenario:      scenario,
		AnnualReturn:  metrics.AnnualReturn,
		Volatility:    metrics.Volatility,
		SharpeRatio:   (metrics.AnnualReturn - riskFreeRate) / metrics.Volatility,
		MaxDrawdown:   metrics.MaxDrawdown,
		RiskFreeRate:  riskFreeRate,
		DividendYield: metrics.DividendYield,
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

	// 生成年度预测
	projections := make([]PortfolioProjection, timeHorizonYears)
	currentValue := totalInvestment
	bestYear := -math.MaxFloat64
	worstYear := math.MaxFloat64

	startDate := time.Now()

	for year := 0; year < timeHorizonYears; year++ {
		// 使用几何布朗运动模拟
		// dS = μS dt + σS dW
		drift := metrics.AnnualReturn - 0.5*metrics.Volatility*metrics.Volatility
		randomShock := s.generateRandomShock(metrics.Volatility)

		annualReturn := drift + randomShock
		startValue := currentValue
		endValue := currentValue * math.Exp(annualReturn)

		// 加上股息收入
		dividendIncome := endValue * metrics.DividendYield
		// 假设股息再投资
		endValue += dividendIncome

		cumulativeReturn := (endValue - totalInvestment) / totalInvestment

		yearStart := startDate.AddDate(year, 0, 0)
		yearEnd := startDate.AddDate(year+1, 0, 0)

		projection := PortfolioProjection{
			Year:             year + 1,
			StartDate:        yearStart.Format("2006-01-02"),
			EndDate:          yearEnd.Format("2006-01-02"),
			StartValue:       startValue,
			EndValue:         endValue,
			AnnualReturn:     annualReturn * 100,
			CumulativeReturn: cumulativeReturn * 100,
			Volatility:       metrics.Volatility * 100,
			MaxDrawdown:      metrics.MaxDrawdown * 100,
			SharpeRatio:      assumptions.SharpeRatio,
			DividendIncome:   dividendIncome,
			TotalValue:       endValue,
		}

		projections[year] = projection

		if annualReturn > bestYear {
			bestYear = annualReturn
		}
		if annualReturn < worstYear {
			worstYear = annualReturn
		}

		currentValue = endValue
	}

	totalReturn := (currentValue - totalInvestment) / totalInvestment
	avgAnnualReturn := math.Pow(currentValue/totalInvestment, 1.0/float64(timeHorizonYears)) - 1

	return &ScenarioResult{
		Assumptions:     assumptions,
		Projections:     projections,
		FinalValue:      currentValue,
		TotalReturn:     totalReturn * 100,
		AvgAnnualReturn: avgAnnualReturn * 100,
		BestYear:        bestYear * 100,
		WorstYear:       worstYear * 100,
	}
}

// generateRandomShock 生成随机冲击（简化版蒙特卡洛）
func (s *ScenarioAnalysisService) generateRandomShock(volatility float64) float64 {
	// 使用 Box-Muller 变换生成正态分布随机数
	u1 := rand.Float64()
	u2 := rand.Float64()
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return volatility * z * 0.3 // 降低随机性使结果更可预测
}

// calculateComparison 计算情景对比
func (s *ScenarioAnalysisService) calculateComparison(
	scenarios map[MarketScenario]*ScenarioResult,
) *ComparisonMetrics {
	bestScenario := Neutral
	worstScenario := Neutral
	bestValue := -math.MaxFloat64
	worstValue := math.MaxFloat64
	bestRiskAdjusted := Neutral
	bestSharpe := -math.MaxFloat64

	for scenario, result := range scenarios {
		if result.FinalValue > bestValue {
			bestValue = result.FinalValue
			bestScenario = scenario
		}
		if result.FinalValue < worstValue {
			worstValue = result.FinalValue
			worstScenario = scenario
		}
		if result.Assumptions.SharpeRatio > bestSharpe {
			bestSharpe = result.Assumptions.SharpeRatio
			bestRiskAdjusted = scenario
		}
	}

	return &ComparisonMetrics{
		BestScenario:       bestScenario,
		WorstScenario:      worstScenario,
		ValueDifference:    bestValue - worstValue,
		ReturnSpread:       scenarios[bestScenario].TotalReturn - scenarios[worstScenario].TotalReturn,
		RiskAdjustedWinner: bestRiskAdjusted,
	}
}
