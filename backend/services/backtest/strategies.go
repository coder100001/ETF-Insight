package backtest

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
)

// BaseStrategy 基础策略
type BaseStrategy struct {
	name        string
	description string
}

func (s *BaseStrategy) GetName() string {
	return s.name
}

func (s *BaseStrategy) GetDescription() string {
	return s.description
}

// MovingAverageCrossStrategy 均线交叉策略
type MovingAverageCrossStrategy struct {
	BaseStrategy
	ShortPeriod  int // 短期均线周期
	LongPeriod   int // 长期均线周期
	priceHistory map[string][]decimal.Decimal
}

// NewMovingAverageCrossStrategy 创建均线交叉策略
func NewMovingAverageCrossStrategy(shortPeriod, longPeriod int) *MovingAverageCrossStrategy {
	return &MovingAverageCrossStrategy{
		BaseStrategy: BaseStrategy{
			name:        "Moving Average Cross",
			description: fmt.Sprintf("MA%d/MA%d crossover strategy", shortPeriod, longPeriod),
		},
		ShortPeriod:  shortPeriod,
		LongPeriod:   longPeriod,
		priceHistory: make(map[string][]decimal.Decimal),
	}
}

// GenerateSignals 生成交易信号
func (s *MovingAverageCrossStrategy) GenerateSignals(engine *BacktestEngine, bar *Bar) []*Signal {
	signals := make([]*Signal, 0)

	// 更新价格历史
	s.priceHistory[bar.Symbol] = append(s.priceHistory[bar.Symbol], bar.Close)

	prices := s.priceHistory[bar.Symbol]
	if len(prices) < s.LongPeriod {
		return signals
	}

	// 计算均线
	shortMA := calculateSMA(prices, s.ShortPeriod)
	longMA := calculateSMA(prices, s.LongPeriod)

	// 需要至少两天的数据来判断交叉
	if len(prices) < s.LongPeriod+1 {
		return signals
	}

	// 前一天的均线
	prevShortMA := calculateSMA(prices[:len(prices)-1], s.ShortPeriod)
	prevLongMA := calculateSMA(prices[:len(prices)-1], s.LongPeriod)

	// 金叉买入信号
	if shortMA.GreaterThan(longMA) && prevShortMA.LessThanOrEqual(prevLongMA) {
		// 计算买入数量 (使用可用资金的50%)
		capital := engine.GetCurrentCapital().Mul(decimal.NewFromFloat(0.5))
		quantity := capital.Div(bar.Close).Round(0)
		if quantity.GreaterThan(decimal.Zero) {
			signals = append(signals, &Signal{
				Type:     SignalBuy,
				Symbol:   bar.Symbol,
				Quantity: quantity,
				Reason:   fmt.Sprintf("金叉买入: MA%d > MA%d", s.ShortPeriod, s.LongPeriod),
			})
		}
	}

	// 死叉卖出信号
	if shortMA.LessThan(longMA) && prevShortMA.GreaterThanOrEqual(prevLongMA) {
		position := engine.GetPosition(bar.Symbol)
		if position != nil && position.Quantity.GreaterThan(decimal.Zero) {
			signals = append(signals, &Signal{
				Type:     SignalSell,
				Symbol:   bar.Symbol,
				Quantity: position.Quantity,
				Reason:   fmt.Sprintf("死叉卖出: MA%d < MA%d", s.ShortPeriod, s.LongPeriod),
			})
		}
	}

	return signals
}

// RSIStrategy RSI策略
type RSIStrategy struct {
	BaseStrategy
	Period       int             // RSI周期
	Oversold     decimal.Decimal // 超卖阈值
	Overbought   decimal.Decimal // 超买阈值
	priceHistory map[string][]decimal.Decimal
}

// NewRSIStrategy 创建RSI策略
func NewRSIStrategy(period int, oversold, overbought float64) *RSIStrategy {
	return &RSIStrategy{
		BaseStrategy: BaseStrategy{
			name:        "RSI Strategy",
			description: fmt.Sprintf("RSI(%d) strategy, oversold=%.1f, overbought=%.1f", period, oversold, overbought),
		},
		Period:       period,
		Oversold:     decimal.NewFromFloat(oversold),
		Overbought:   decimal.NewFromFloat(overbought),
		priceHistory: make(map[string][]decimal.Decimal),
	}
}

// GenerateSignals 生成交易信号
func (s *RSIStrategy) GenerateSignals(engine *BacktestEngine, bar *Bar) []*Signal {
	signals := make([]*Signal, 0)

	// 更新价格历史
	s.priceHistory[bar.Symbol] = append(s.priceHistory[bar.Symbol], bar.Close)

	prices := s.priceHistory[bar.Symbol]
	if len(prices) < s.Period+1 {
		return signals
	}

	// 计算RSI
	rsi := calculateRSI(prices, s.Period)

	// RSI超卖买入
	if rsi.LessThanOrEqual(s.Oversold) {
		capital := engine.GetCurrentCapital().Mul(decimal.NewFromFloat(0.3))
		quantity := capital.Div(bar.Close).Round(0)
		if quantity.GreaterThan(decimal.Zero) {
			signals = append(signals, &Signal{
				Type:     SignalBuy,
				Symbol:   bar.Symbol,
				Quantity: quantity,
				Reason:   fmt.Sprintf("RSI超卖买入: RSI=%.2f", rsi.InexactFloat64()),
			})
		}
	}

	// RSI超买卖出
	if rsi.GreaterThanOrEqual(s.Overbought) {
		position := engine.GetPosition(bar.Symbol)
		if position != nil && position.Quantity.GreaterThan(decimal.Zero) {
			signals = append(signals, &Signal{
				Type:     SignalSell,
				Symbol:   bar.Symbol,
				Quantity: position.Quantity,
				Reason:   fmt.Sprintf("RSI超买卖出: RSI=%.2f", rsi.InexactFloat64()),
			})
		}
	}

	return signals
}

// MomentumStrategy 动量策略
type MomentumStrategy struct {
	BaseStrategy
	LookbackPeriod     int // 回看周期
	TopN               int // 选择前N个动量最强的标的
	RebalanceFreq      int // 调仓频率(天)
	momentumCache      map[string]decimal.Decimal
	priceHistory       map[string][]decimal.Decimal // 价格历史
	daysSinceRebalance int
}

// NewMomentumStrategy 创建动量策略
func NewMomentumStrategy(lookbackPeriod, topN, rebalanceFreq int) *MomentumStrategy {
	return &MomentumStrategy{
		BaseStrategy: BaseStrategy{
			name:        "Momentum Strategy",
			description: fmt.Sprintf("%d-day momentum, top %d, rebalance every %d days", lookbackPeriod, topN, rebalanceFreq),
		},
		LookbackPeriod: lookbackPeriod,
		TopN:           topN,
		RebalanceFreq:  rebalanceFreq,
		momentumCache:  make(map[string]decimal.Decimal),
	}
}

// GenerateSignals 生成交易信号
func (s *MomentumStrategy) GenerateSignals(engine *BacktestEngine, bar *Bar) []*Signal {
	signals := make([]*Signal, 0)

	// 更新价格历史用于动量计算
	if s.priceHistory == nil {
		s.priceHistory = make(map[string][]decimal.Decimal)
	}
	s.priceHistory[bar.Symbol] = append(s.priceHistory[bar.Symbol], bar.Close)

	// 计算并缓存当前标的的动量值
	if len(s.priceHistory[bar.Symbol]) >= s.LookbackPeriod {
		prices := s.priceHistory[bar.Symbol]
		// 动量 = (当前价格 / N天前价格) - 1
		momentum := prices[len(prices)-1].Div(prices[len(prices)-s.LookbackPeriod]).Sub(decimal.NewFromInt(1))
		s.momentumCache[bar.Symbol] = momentum
	}

	s.daysSinceRebalance++

	// 到达调仓日
	if s.daysSinceRebalance >= s.RebalanceFreq {
		s.daysSinceRebalance = 0

		// 按动量值排序所有标的
		type symbolMomentum struct {
			symbol   string
			momentum decimal.Decimal
		}
		var sortedSymbols []symbolMomentum
		for symbol, momentum := range s.momentumCache {
			sortedSymbols = append(sortedSymbols, symbolMomentum{symbol: symbol, momentum: momentum})
		}

		// 按动量降序排序（使用 sort.Slice，O(n log n)）
		sort.Slice(sortedSymbols, func(i, j int) bool {
			return sortedSymbols[i].momentum.GreaterThan(sortedSymbols[j].momentum)
		})

		// 获取前N个标的
		topSymbols := make(map[string]bool)
		for i := 0; i < s.TopN && i < len(sortedSymbols); i++ {
			topSymbols[sortedSymbols[i].symbol] = true
		}

		// 卖出不在前N的持仓
		positions := engine.GetAllPositions()
		for symbol, position := range positions {
			if position.Quantity.GreaterThan(decimal.Zero) && !topSymbols[symbol] {
				signals = append(signals, &Signal{
					Type:     SignalSell,
					Symbol:   symbol,
					Quantity: position.Quantity,
					Reason:   fmt.Sprintf("动量调仓卖出: 动量排名下降"),
				})
			}
		}

		// 买入动量最强的标的（只买入当前bar对应的标的如果在topN中）
		if topSymbols[bar.Symbol] {
			// 检查是否已持有
			position := engine.GetPosition(bar.Symbol)
			if position == nil || position.Quantity.IsZero() {
				// 平均分配资金
				capital := engine.GetCurrentCapital().Mul(decimal.NewFromFloat(0.9)) // 保留10%现金
				quantity := capital.Div(bar.Close).Div(decimal.NewFromInt(int64(s.TopN))).Round(0)
				if quantity.GreaterThan(decimal.Zero) {
					momentum := s.momentumCache[bar.Symbol]
					signals = append(signals, &Signal{
						Type:     SignalBuy,
						Symbol:   bar.Symbol,
						Quantity: quantity,
						Reason:   fmt.Sprintf("动量策略买入: momentum=%.4f", momentum.InexactFloat64()),
					})
				}
			}
		}
	}

	return signals
}

// FactorBasedStrategy 因子策略
type FactorBasedStrategy struct {
	BaseStrategy
	factorWeights map[FactorType]decimal.Decimal // 因子权重
	factorLibrary *FactorLibrary
	priceHistory  map[string][]decimal.Decimal
}

// NewFactorBasedStrategy 创建因子策略
func NewFactorBasedStrategy(weights map[FactorType]float64) *FactorBasedStrategy {
	factorWeights := make(map[FactorType]decimal.Decimal)
	for k, v := range weights {
		factorWeights[k] = decimal.NewFromFloat(v)
	}

	return &FactorBasedStrategy{
		BaseStrategy: BaseStrategy{
			name:        "Factor-Based Strategy",
			description: "Multi-factor quantitative strategy",
		},
		factorWeights: factorWeights,
		factorLibrary: NewFactorLibrary(nil),
		priceHistory:  make(map[string][]decimal.Decimal),
	}
}

// GenerateSignals 生成交易信号
func (s *FactorBasedStrategy) GenerateSignals(engine *BacktestEngine, bar *Bar) []*Signal {
	signals := make([]*Signal, 0)

	// 更新价格历史
	s.priceHistory[bar.Symbol] = append(s.priceHistory[bar.Symbol], bar.Close)

	// 计算综合因子得分
	factorScores := s.CalculateFactors(engine, bar)

	// 计算加权得分
	totalScore := decimal.Zero
	for factorType, weight := range s.factorWeights {
		if score, ok := factorScores[string(factorType)]; ok {
			totalScore = totalScore.Add(score.Mul(weight))
		}
	}

	// 根据因子得分生成信号
	// 简化：得分>0买入，得分<0卖出
	if totalScore.GreaterThan(decimal.Zero) {
		capital := engine.GetCurrentCapital().Mul(totalScore.Abs()) // 根据得分调整仓位
		quantity := capital.Div(bar.Close).Round(0)
		if quantity.GreaterThan(decimal.Zero) {
			signals = append(signals, &Signal{
				Type:     SignalBuy,
				Symbol:   bar.Symbol,
				Quantity: quantity,
				Reason:   fmt.Sprintf("因子得分买入: score=%.4f", totalScore.InexactFloat64()),
			})
		}
	} else if totalScore.LessThan(decimal.Zero) {
		position := engine.GetPosition(bar.Symbol)
		if position != nil && position.Quantity.GreaterThan(decimal.Zero) {
			sellRatio := totalScore.Abs() // 根据负得分调整卖出比例
			if sellRatio.GreaterThan(decimal.NewFromInt(1)) {
				sellRatio = decimal.NewFromInt(1)
			}
			quantity := position.Quantity.Mul(sellRatio).Round(0)
			if quantity.GreaterThan(decimal.Zero) {
				signals = append(signals, &Signal{
					Type:     SignalSell,
					Symbol:   bar.Symbol,
					Quantity: quantity,
					Reason:   fmt.Sprintf("因子得分卖出: score=%.4f", totalScore.InexactFloat64()),
				})
			}
		}
	}

	return signals
}

// CalculateFactors 计算因子值
func (s *FactorBasedStrategy) CalculateFactors(engine *BacktestEngine, bar *Bar) map[string]decimal.Decimal {
	factors := make(map[string]decimal.Decimal)

	prices := s.priceHistory[bar.Symbol]
	if len(prices) < 20 {
		return factors
	}

	// 计算动量因子
	momentum := s.factorLibrary.CalculateMomentumFactor(prices, 12*20) // 约12个月
	factors[string(FactorMomentum)] = momentum

	// 计算低波因子
	returns := calculateReturns(prices)
	lowVol := s.factorLibrary.CalculateLowVolFactor(returns)
	factors[string(FactorLowVol)] = lowVol

	return factors
}

// GetFactorNames 获取因子名称列表
func (s *FactorBasedStrategy) GetFactorNames() []string {
	names := make([]string, 0, len(s.factorWeights))
	for factorType := range s.factorWeights {
		names = append(names, string(factorType))
	}
	return names
}

// BuyAndHoldStrategy 买入持有策略
type BuyAndHoldStrategy struct {
	BaseStrategy
	bought bool
}

// NewBuyAndHoldStrategy 创建买入持有策略
func NewBuyAndHoldStrategy() *BuyAndHoldStrategy {
	return &BuyAndHoldStrategy{
		BaseStrategy: BaseStrategy{
			name:        "Buy and Hold",
			description: "Simple buy and hold strategy",
		},
		bought: false,
	}
}

// GenerateSignals 生成交易信号
func (s *BuyAndHoldStrategy) GenerateSignals(engine *BacktestEngine, bar *Bar) []*Signal {
	signals := make([]*Signal, 0)

	if !s.bought {
		// 全仓买入
		capital := engine.GetCurrentCapital().Mul(decimal.NewFromFloat(0.95)) // 保留5%现金
		quantity := capital.Div(bar.Close).Round(0)
		if quantity.GreaterThan(decimal.Zero) {
			signals = append(signals, &Signal{
				Type:     SignalBuy,
				Symbol:   bar.Symbol,
				Quantity: quantity,
				Reason:   "买入持有策略初始买入",
			})
			s.bought = true
		}
	}

	return signals
}

// 辅助函数

func calculateSMA(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period {
		return decimal.Zero
	}

	sum := decimal.Zero
	for i := len(prices) - period; i < len(prices); i++ {
		sum = sum.Add(prices[i])
	}

	return sum.Div(decimal.NewFromInt(int64(period)))
}

func calculateRSI(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period+1 {
		return decimal.NewFromFloat(50) // 默认返回中性值
	}

	gains := decimal.Zero
	losses := decimal.Zero

	for i := len(prices) - period; i < len(prices); i++ {
		change := prices[i].Sub(prices[i-1])
		if change.GreaterThan(decimal.Zero) {
			gains = gains.Add(change)
		} else {
			losses = losses.Add(change.Abs())
		}
	}

	if losses.IsZero() {
		return decimal.NewFromInt(100)
	}

	avgGain := gains.Div(decimal.NewFromInt(int64(period)))
	avgLoss := losses.Div(decimal.NewFromInt(int64(period)))

	rs := avgGain.Div(avgLoss)
	rsi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))

	return rsi
}

func calculateReturns(prices []decimal.Decimal) []decimal.Decimal {
	if len(prices) < 2 {
		return nil
	}

	returns := make([]decimal.Decimal, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = prices[i].Sub(prices[i-1]).Div(prices[i-1])
	}

	return returns
}
