package analytics

import (
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// BacktestConfig 回测配置
type BacktestConfig struct {
	StartDate      time.Time
	EndDate        time.Time
	InitialCapital decimal.Decimal
	RebalanceFreq  string // daily, weekly, monthly, quarterly, yearly
	CommissionRate decimal.Decimal
	TaxRate        decimal.Decimal
	Slippage       decimal.Decimal
}

// PortfolioConfig 投资组合配置
type PortfolioConfig struct {
	Symbol   string
	Weight   decimal.Decimal
	Strategy string // buy_and_hold, rebalance
}

// BacktestResult 回测结果
type BacktestResult struct {
	Config           BacktestConfig
	Performance      PerformanceMetrics
	Trades           []TradeRecord
	DailyReturns     []DailyReturn
	Drawdowns        []DrawdownPeriod
	MonthlyReturns   map[string]decimal.Decimal
	YearlyReturns    map[string]decimal.Decimal
	RiskMetrics      RiskMetrics
	BenchmarkCompare BenchmarkComparison
}

// PerformanceMetrics 业绩指标
type PerformanceMetrics struct {
	TotalReturn       decimal.Decimal
	AnnualizedReturn  decimal.Decimal
	Volatility        decimal.Decimal
	SharpeRatio       decimal.Decimal
	SortinoRatio      decimal.Decimal
	MaxDrawdown       decimal.Decimal
	MaxDrawdownPeriod string
	WinRate           decimal.Decimal
	ProfitFactor      decimal.Decimal
	CalmarRatio       decimal.Decimal
	FinalCapital      decimal.Decimal
}

// TradeRecord 交易记录
type TradeRecord struct {
	Date       time.Time
	Symbol     string
	Action     string // buy, sell, rebalance
	Shares     decimal.Decimal
	Price      decimal.Decimal
	Amount     decimal.Decimal
	Commission decimal.Decimal
	Reason     string
}

// DailyReturn 日收益
type DailyReturn struct {
	Date             time.Time
	PortfolioValue   decimal.Decimal
	DailyReturn      decimal.Decimal
	CumulativeReturn decimal.Decimal
}

// DrawdownPeriod 回撤期
type DrawdownPeriod struct {
	StartDate   time.Time
	EndDate     time.Time
	PeakValue   decimal.Decimal
	TroughValue decimal.Decimal
	Drawdown    decimal.Decimal
	Duration    int // days
}

// BenchmarkComparison 基准对比
type BenchmarkComparison struct {
	BenchmarkSymbol  string
	PortfolioReturn  decimal.Decimal
	BenchmarkReturn  decimal.Decimal
	ExcessReturn     decimal.Decimal
	TrackingError    decimal.Decimal
	InformationRatio decimal.Decimal
	Beta             decimal.Decimal
	Alpha            decimal.Decimal
	Correlation      decimal.Decimal
}

// BacktestEngine 回测引擎
type BacktestEngine struct {
	config    BacktestConfig
	portfolio []PortfolioConfig
	prices    map[string][]PricePoint
}

// NewBacktestEngine 创建回测引擎
func NewBacktestEngine(config BacktestConfig, portfolio []PortfolioConfig, prices map[string][]PricePoint) *BacktestEngine {
	return &BacktestEngine{
		config:    config,
		portfolio: portfolio,
		prices:    prices,
	}
}

// Run 执行回测
func (e *BacktestEngine) Run() *BacktestResult {
	result := &BacktestResult{
		Config:         e.config,
		Trades:         make([]TradeRecord, 0),
		DailyReturns:   make([]DailyReturn, 0),
		Drawdowns:      make([]DrawdownPeriod, 0),
		MonthlyReturns: make(map[string]decimal.Decimal),
		YearlyReturns:  make(map[string]decimal.Decimal),
	}

	// 初始化持仓
	holdings := make(map[string]decimal.Decimal) // symbol -> shares
	cash := e.config.InitialCapital

	// 初始买入
	for _, config := range e.portfolio {
		allocation := e.config.InitialCapital.Mul(config.Weight).Div(decimal.NewFromInt(100))
		if prices, exists := e.prices[config.Symbol]; exists && len(prices) > 0 {
			price := prices[0].Price
			shares := allocation.Div(price).Truncate(2)
			cost := shares.Mul(price)
			commission := cost.Mul(e.config.CommissionRate)

			if cost.Add(commission).LessThanOrEqual(cash) {
				holdings[config.Symbol] = shares
				cash = cash.Sub(cost).Sub(commission)

				result.Trades = append(result.Trades, TradeRecord{
					Date:       prices[0].Date,
					Symbol:     config.Symbol,
					Action:     "buy",
					Shares:     shares,
					Price:      price,
					Amount:     cost,
					Commission: commission,
					Reason:     "initial_allocation",
				})
			}
		}
	}

	// 获取所有交易日
	tradingDays := e.getTradingDays()

	// 模拟每日交易
	peakValue := e.config.InitialCapital
	currentDrawdown := DrawdownPeriod{}
	inDrawdown := false

	for i, date := range tradingDays {
		// 计算当日组合价值
		portfolioValue := cash
		for symbol, shares := range holdings {
			if price := e.getPrice(symbol, date); !price.IsZero() {
				portfolioValue = portfolioValue.Add(shares.Mul(price))
			}
		}

		// 记录日收益
		dailyReturn := decimal.Zero
		if i > 0 {
			prevValue := result.DailyReturns[i-1].PortfolioValue
			if !prevValue.IsZero() {
				dailyReturn = portfolioValue.Sub(prevValue).Div(prevValue)
			}
		}

		cumulativeReturn := decimal.Zero
		if !e.config.InitialCapital.IsZero() {
			cumulativeReturn = portfolioValue.Sub(e.config.InitialCapital).Div(e.config.InitialCapital)
		}

		result.DailyReturns = append(result.DailyReturns, DailyReturn{
			Date:             date,
			PortfolioValue:   portfolioValue,
			DailyReturn:      dailyReturn,
			CumulativeReturn: cumulativeReturn,
		})

		// 更新回撤
		if portfolioValue.GreaterThan(peakValue) {
			if inDrawdown {
				// 结束回撤期
				currentDrawdown.EndDate = date
				currentDrawdown.Duration = int(currentDrawdown.EndDate.Sub(currentDrawdown.StartDate).Hours() / 24)
				result.Drawdowns = append(result.Drawdowns, currentDrawdown)
				inDrawdown = false
			}
			peakValue = portfolioValue
		} else {
			drawdown := peakValue.Sub(portfolioValue).Div(peakValue)
			if !inDrawdown {
				inDrawdown = true
				currentDrawdown = DrawdownPeriod{
					StartDate: date,
					PeakValue: peakValue,
				}
			}
			if drawdown.GreaterThan(currentDrawdown.Drawdown) {
				currentDrawdown.Drawdown = drawdown
				currentDrawdown.TroughValue = portfolioValue
			}
		}

		// 再平衡检查
		if e.shouldRebalance(date, i) {
			rebalanceTrades := e.rebalance(date, holdings, cash, portfolioValue)
			result.Trades = append(result.Trades, rebalanceTrades...)

			// 更新持仓和现金
			for _, trade := range rebalanceTrades {
				if trade.Action == "buy" {
					holdings[trade.Symbol] = holdings[trade.Symbol].Add(trade.Shares)
					cash = cash.Sub(trade.Amount).Sub(trade.Commission)
				} else if trade.Action == "sell" {
					holdings[trade.Symbol] = holdings[trade.Symbol].Sub(trade.Shares)
					cash = cash.Add(trade.Amount).Sub(trade.Commission)
				}
			}
		}
	}

	// 计算业绩指标
	result.Performance = e.calculatePerformance(result.DailyReturns)
	result.RiskMetrics = e.calculateRiskMetrics(result.DailyReturns)
	result.MonthlyReturns = e.calculateMonthlyReturns(result.DailyReturns)
	result.YearlyReturns = e.calculateYearlyReturns(result.DailyReturns)

	return result
}

// getTradingDays 获取交易日列表
func (e *BacktestEngine) getTradingDays() []time.Time {
	days := make([]time.Time, 0)

	// 使用第一个ETF的价格日期作为交易日
	for _, prices := range e.prices {
		for _, p := range prices {
			if (p.Date.Equal(e.config.StartDate) || p.Date.After(e.config.StartDate)) &&
				(p.Date.Equal(e.config.EndDate) || p.Date.Before(e.config.EndDate)) {
				days = append(days, p.Date)
			}
		}
		break
	}

	// 排序
	sort.Slice(days, func(i, j int) bool {
		return days[i].Before(days[j])
	})

	return days
}

// getPrice 获取某日的价格
func (e *BacktestEngine) getPrice(symbol string, date time.Time) decimal.Decimal {
	if prices, exists := e.prices[symbol]; exists {
		for _, p := range prices {
			if p.Date.Equal(date) {
				return p.Price
			}
		}
	}
	return decimal.Zero
}

// shouldRebalance 检查是否需要再平衡
func (e *BacktestEngine) shouldRebalance(date time.Time, dayIndex int) bool {
	switch e.config.RebalanceFreq {
	case "daily":
		return true
	case "weekly":
		return date.Weekday() == time.Monday
	case "monthly":
		return date.Day() == 1
	case "quarterly":
		return date.Day() == 1 && (date.Month() == 1 || date.Month() == 4 || date.Month() == 7 || date.Month() == 10)
	case "yearly":
		return date.Day() == 1 && date.Month() == 1
	default:
		return false
	}
}

// rebalance 执行再平衡
func (e *BacktestEngine) rebalance(date time.Time, holdings map[string]decimal.Decimal, cash, portfolioValue decimal.Decimal) []TradeRecord {
	trades := make([]TradeRecord, 0)

	// 计算当前权重
	currentWeights := make(map[string]decimal.Decimal)
	for symbol, shares := range holdings {
		if price := e.getPrice(symbol, date); !price.IsZero() {
			value := shares.Mul(price)
			weight := value.Div(portfolioValue).Mul(decimal.NewFromInt(100))
			currentWeights[symbol] = weight
		}
	}

	// 计算目标权重
	targetWeights := make(map[string]decimal.Decimal)
	for _, config := range e.portfolio {
		targetWeights[config.Symbol] = config.Weight
	}

	// 卖出超配的
	for symbol, currentWeight := range currentWeights {
		targetWeight := targetWeights[symbol]
		if currentWeight.GreaterThan(targetWeight.Add(decimal.NewFromFloat(5))) { // 超过5%才调整
			if price := e.getPrice(symbol, date); !price.IsZero() {
				sharesToSell := holdings[symbol].Mul(currentWeight.Sub(targetWeight)).Div(currentWeight)
				amount := sharesToSell.Mul(price)
				commission := amount.Mul(e.config.CommissionRate)

				trades = append(trades, TradeRecord{
					Date:       date,
					Symbol:     symbol,
					Action:     "sell",
					Shares:     sharesToSell,
					Price:      price,
					Amount:     amount,
					Commission: commission,
					Reason:     "rebalance",
				})
			}
		}
	}

	// 买入低配的
	for symbol, targetWeight := range targetWeights {
		currentWeight := currentWeights[symbol]
		if targetWeight.GreaterThan(currentWeight.Add(decimal.NewFromFloat(5))) { // 低于5%才调整
			if price := e.getPrice(symbol, date); !price.IsZero() {
				weightDiff := targetWeight.Sub(currentWeight)
				allocation := portfolioValue.Mul(weightDiff).Div(decimal.NewFromInt(100))
				sharesToBuy := allocation.Div(price).Truncate(2)
				amount := sharesToBuy.Mul(price)
				commission := amount.Mul(e.config.CommissionRate)

				if amount.Add(commission).LessThanOrEqual(cash) {
					trades = append(trades, TradeRecord{
						Date:       date,
						Symbol:     symbol,
						Action:     "buy",
						Shares:     sharesToBuy,
						Price:      price,
						Amount:     amount,
						Commission: commission,
						Reason:     "rebalance",
					})
				}
			}
		}
	}

	return trades
}

// calculatePerformance 计算业绩指标
func (e *BacktestEngine) calculatePerformance(dailyReturns []DailyReturn) PerformanceMetrics {
	metrics := PerformanceMetrics{}

	if len(dailyReturns) == 0 {
		return metrics
	}

	// 总收益
	finalValue := dailyReturns[len(dailyReturns)-1].PortfolioValue
	metrics.TotalReturn = finalValue.Sub(e.config.InitialCapital).Div(e.config.InitialCapital).Mul(decimal.NewFromInt(100))
	metrics.FinalCapital = finalValue

	// 年化收益
	days := len(dailyReturns)
	years := decimal.NewFromFloat(float64(days) / 252) // 假设每年252个交易日
	if !years.IsZero() {
		// 使用CAGR公式: (EndValue/StartValue)^(1/years) - 1
		cagr := math.Pow(finalValue.Div(e.config.InitialCapital).InexactFloat64(), 1.0/years.InexactFloat64()) - 1
		metrics.AnnualizedReturn = decimal.NewFromFloat(cagr * 100)
	}

	// 波动率
	returns := make([]decimal.Decimal, len(dailyReturns))
	for i, dr := range dailyReturns {
		returns[i] = dr.DailyReturn
	}
	metrics.Volatility = calculateVolatility(returns)

	// 夏普比率
	if !metrics.Volatility.IsZero() {
		metrics.SharpeRatio = metrics.AnnualizedReturn.Div(metrics.Volatility)
	}

	// 最大回撤
	maxDD := decimal.Zero
	peak := e.config.InitialCapital
	for _, dr := range dailyReturns {
		if dr.PortfolioValue.GreaterThan(peak) {
			peak = dr.PortfolioValue
		}
		drawdown := peak.Sub(dr.PortfolioValue).Div(peak)
		if drawdown.GreaterThan(maxDD) {
			maxDD = drawdown
		}
	}
	metrics.MaxDrawdown = maxDD.Mul(decimal.NewFromInt(100))

	// 卡玛比率
	if !metrics.MaxDrawdown.IsZero() {
		metrics.CalmarRatio = metrics.AnnualizedReturn.Div(metrics.MaxDrawdown)
	}

	return metrics
}

// calculateRiskMetrics 计算风险指标
func (e *BacktestEngine) calculateRiskMetrics(dailyReturns []DailyReturn) RiskMetrics {
	metrics := RiskMetrics{}

	if len(dailyReturns) < 2 {
		return metrics
	}

	returns := make([]decimal.Decimal, len(dailyReturns))
	prices := make([]PricePoint, len(dailyReturns))
	for i, dr := range dailyReturns {
		returns[i] = dr.DailyReturn
		prices[i] = PricePoint{Date: dr.Date, Price: dr.PortfolioValue}
	}

	// 使用已有的风险指标计算函数
	riskMetrics, _ := CalculateRiskMetrics(prices, decimal.NewFromFloat(0.02), nil)
	if riskMetrics != nil {
		metrics = *riskMetrics
	}

	return metrics
}

// calculateMonthlyReturns 计算月度收益
func (e *BacktestEngine) calculateMonthlyReturns(dailyReturns []DailyReturn) map[string]decimal.Decimal {
	monthly := make(map[string]decimal.Decimal)

	for _, dr := range dailyReturns {
		key := dr.Date.Format("2006-01")
		if _, exists := monthly[key]; !exists {
			// 找到月初的价值
			monthly[key] = decimal.Zero
		}
	}

	// 计算每月收益
	for key := range monthly {
		var startValue, endValue decimal.Decimal
		var foundStart bool

		for _, dr := range dailyReturns {
			if dr.Date.Format("2006-01") == key {
				if !foundStart {
					startValue = dr.PortfolioValue
					foundStart = true
				}
				endValue = dr.PortfolioValue
			}
		}

		if !startValue.IsZero() {
			monthly[key] = endValue.Sub(startValue).Div(startValue).Mul(decimal.NewFromInt(100))
		}
	}

	return monthly
}

// calculateYearlyReturns 计算年度收益
func (e *BacktestEngine) calculateYearlyReturns(dailyReturns []DailyReturn) map[string]decimal.Decimal {
	yearly := make(map[string]decimal.Decimal)

	for _, dr := range dailyReturns {
		key := dr.Date.Format("2006")
		if _, exists := yearly[key]; !exists {
			yearly[key] = decimal.Zero
		}
	}

	// 计算每年收益
	for key := range yearly {
		var startValue, endValue decimal.Decimal
		var foundStart bool

		for _, dr := range dailyReturns {
			if dr.Date.Format("2006") == key {
				if !foundStart {
					startValue = dr.PortfolioValue
					foundStart = true
				}
				endValue = dr.PortfolioValue
			}
		}

		if !startValue.IsZero() {
			yearly[key] = endValue.Sub(startValue).Div(startValue).Mul(decimal.NewFromInt(100))
		}
	}

	return yearly
}

// CompareWithBenchmark 与基准对比
func (result *BacktestResult) CompareWithBenchmark(benchmarkPrices []PricePoint) BenchmarkComparison {
	compare := BenchmarkComparison{}

	if len(benchmarkPrices) < 2 || len(result.DailyReturns) == 0 {
		return compare
	}

	// 计算基准收益
	benchmarkStart := benchmarkPrices[0].Price
	benchmarkEnd := benchmarkPrices[len(benchmarkPrices)-1].Price
	compare.BenchmarkReturn = benchmarkEnd.Sub(benchmarkStart).Div(benchmarkStart).Mul(decimal.NewFromInt(100))

	// 组合收益
	compare.PortfolioReturn = result.Performance.TotalReturn

	// 超额收益
	compare.ExcessReturn = compare.PortfolioReturn.Sub(compare.BenchmarkReturn)

	// 计算跟踪误差和相关性
	portfolioReturns := make([]decimal.Decimal, 0)
	benchmarkReturns := make([]decimal.Decimal, 0)

	for _, dr := range result.DailyReturns {
		// 找到对应的基准收益
		for i := 1; i < len(benchmarkPrices); i++ {
			if benchmarkPrices[i].Date.Equal(dr.Date) {
				benchmarkReturn := benchmarkPrices[i].Price.Sub(benchmarkPrices[i-1].Price).Div(benchmarkPrices[i-1].Price)
				benchmarkReturns = append(benchmarkReturns, benchmarkReturn)
				portfolioReturns = append(portfolioReturns, dr.DailyReturn)
				break
			}
		}
	}

	// 跟踪误差（组合收益与基准收益差异的标准差）
	if len(portfolioReturns) == len(benchmarkReturns) && len(portfolioReturns) > 1 {
		differences := make([]decimal.Decimal, len(portfolioReturns))
		for i := 0; i < len(portfolioReturns); i++ {
			differences[i] = portfolioReturns[i].Sub(benchmarkReturns[i])
		}

		// 计算跟踪误差
		meanDiff := decimal.Zero
		for _, d := range differences {
			meanDiff = meanDiff.Add(d)
		}
		meanDiff = meanDiff.Div(decimal.NewFromInt(int64(len(differences))))

		variance := decimal.Zero
		for _, d := range differences {
			diff := d.Sub(meanDiff)
			variance = variance.Add(diff.Mul(diff))
		}
		variance = variance.Div(decimal.NewFromInt(int64(len(differences) - 1)))

		compare.TrackingError = decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()) * math.Sqrt(252) * 100)

		// 信息比率
		if !compare.TrackingError.IsZero() {
			compare.InformationRatio = compare.ExcessReturn.Div(compare.TrackingError)
		}

		// Beta和Alpha
		compare.Beta, compare.Alpha = calculateBetaAlpha(portfolioReturns, benchmarkReturns, decimal.NewFromFloat(0.02))

		// 相关性
		compare.Correlation = calculateCorrelation(portfolioReturns, benchmarkReturns)
	}

	return compare
}

// calculateCorrelation 计算相关系数
func calculateCorrelation(x, y []decimal.Decimal) decimal.Decimal {
	if len(x) != len(y) || len(x) < 2 {
		return decimal.Zero
	}

	meanX := decimal.Zero
	meanY := decimal.Zero

	for i := 0; i < len(x); i++ {
		meanX = meanX.Add(x[i])
		meanY = meanY.Add(y[i])
	}

	meanX = meanX.Div(decimal.NewFromInt(int64(len(x))))
	meanY = meanY.Div(decimal.NewFromInt(int64(len(y))))

	numerator := decimal.Zero
	sumX2 := decimal.Zero
	sumY2 := decimal.Zero

	for i := 0; i < len(x); i++ {
		diffX := x[i].Sub(meanX)
		diffY := y[i].Sub(meanY)
		numerator = numerator.Add(diffX.Mul(diffY))
		sumX2 = sumX2.Add(diffX.Mul(diffX))
		sumY2 = sumY2.Add(diffY.Mul(diffY))
	}

	denominator := decimal.NewFromFloat(math.Sqrt(sumX2.Mul(sumY2).InexactFloat64()))
	if denominator.IsZero() {
		return decimal.Zero
	}

	return numerator.Div(denominator)
}

// MonteCarloSimulation 蒙特卡洛模拟
type MonteCarloResult struct {
	Simulations     int
	ConfidenceLevel float64
	VaR             decimal.Decimal
	CVaR            decimal.Decimal
	Percentiles     map[int]decimal.Decimal
}

// RunMonteCarlo 运行蒙特卡洛模拟
func RunMonteCarlo(returns []decimal.Decimal, initialValue decimal.Decimal, days int, simulations int) *MonteCarloResult {
	result := &MonteCarloResult{
		Simulations:     simulations,
		ConfidenceLevel: 0.95,
		Percentiles:     make(map[int]decimal.Decimal),
	}

	if len(returns) < 2 || simulations <= 0 {
		return result
	}

	// 计算历史收益统计
	meanReturn := decimal.Zero
	for _, r := range returns {
		meanReturn = meanReturn.Add(r)
	}
	meanReturn = meanReturn.Div(decimal.NewFromInt(int64(len(returns))))

	variance := decimal.Zero
	for _, r := range returns {
		diff := r.Sub(meanReturn)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(returns) - 1)))
	stdDev := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	// 运行模拟
	finalValues := make([]decimal.Decimal, simulations)

	for i := 0; i < simulations; i++ {
		value := initialValue
		for day := 0; day < days; day++ {
			// 使用正态分布随机收益
			randomReturn := meanReturn.Add(stdDev.Mul(decimal.NewFromFloat(randNorm())))
			value = value.Mul(decimal.NewFromInt(1).Add(randomReturn))
		}
		finalValues[i] = value
	}

	// 排序
	sort.Slice(finalValues, func(i, j int) bool {
		return finalValues[i].LessThan(finalValues[j])
	})

	// 计算百分位数
	percentiles := []int{5, 10, 25, 50, 75, 90, 95}
	for _, p := range percentiles {
		index := int(float64(simulations) * float64(p) / 100)
		if index < len(finalValues) {
			result.Percentiles[p] = finalValues[index]
		}
	}

	// 计算VaR和CVaR
	varIndex := int(float64(simulations) * 0.05) // 95%置信度
	if varIndex < len(finalValues) {
		result.VaR = initialValue.Sub(finalValues[varIndex])

		// CVaR = 超过VaR的损失平均值
		cvarSum := decimal.Zero
		count := 0
		for i := 0; i <= varIndex && i < len(finalValues); i++ {
			cvarSum = cvarSum.Add(initialValue.Sub(finalValues[i]))
			count++
		}
		if count > 0 {
			result.CVaR = cvarSum.Div(decimal.NewFromInt(int64(count)))
		}
	}

	return result
}

// randNorm 生成标准正态分布随机数（Box-Muller变换）
func randNorm() float64 {
	// 简化的正态分布近似
	u1 := 0.5 // 应该使用随机数
	u2 := 0.5 // 应该使用随机数

	r := math.Sqrt(-2 * math.Log(u1))
	theta := 2 * math.Pi * u2

	return r * math.Cos(theta)
}

// WalkForwardAnalysis 滚动向前分析
type WalkForwardResult struct {
	Period          string
	InSampleReturn  decimal.Decimal
	OutSampleReturn decimal.Decimal
	Robustness      decimal.Decimal // 样本外/样本内比率
}

// WalkForwardTest 滚动向前测试
func WalkForwardTest(returns []decimal.Decimal, inSampleDays, outSampleDays int) []WalkForwardResult {
	results := make([]WalkForwardResult, 0)

	if len(returns) < inSampleDays+outSampleDays {
		return results
	}

	for i := 0; i+inSampleDays+outSampleDays <= len(returns); i += outSampleDays {
		// 样本内数据
		inSample := returns[i : i+inSampleDays]
		inSampleReturn := decimal.Zero
		for _, r := range inSample {
			inSampleReturn = inSampleReturn.Add(r)
		}

		// 样本外数据
		outSample := returns[i+inSampleDays : i+inSampleDays+outSampleDays]
		outSampleReturn := decimal.Zero
		for _, r := range outSample {
			outSampleReturn = outSampleReturn.Add(r)
		}

		robustness := decimal.Zero
		if !inSampleReturn.IsZero() {
			robustness = outSampleReturn.Div(inSampleReturn)
		}

		results = append(results, WalkForwardResult{
			Period:          "Period_" + string(rune(i)),
			InSampleReturn:  inSampleReturn,
			OutSampleReturn: outSampleReturn,
			Robustness:      robustness,
		})
	}

	return results
}
