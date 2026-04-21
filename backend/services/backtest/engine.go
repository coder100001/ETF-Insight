package backtest

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// BacktestEngine 回测引擎
// 支持事件驱动架构，自定义买卖规则，滑点、手续费、股息再投资
type BacktestEngine struct {
	initialCapital decimal.Decimal
	currentCapital decimal.Decimal
	positions      map[string]*Position
	transactions   []*Transaction
	equityCurve    []*EquityPoint
	factors        map[string]decimal.Decimal // 因子值缓存

	// 模型配置
	slippageModel   SlippageModel
	commissionModel CommissionModel
	dividendModel   DividendModel

	// 策略接口
	strategy Strategy

	// 数据
	dataProvider DataProvider

	// 统计
	startDate        time.Time
	endDate          time.Time
	totalReturn      decimal.Decimal
	annualizedReturn decimal.Decimal
	maxDrawdown      decimal.Decimal
	sharpeRatio      decimal.Decimal
	sortinoRatio     decimal.Decimal
	calmarRatio      decimal.Decimal
	winRate          decimal.Decimal
	profitFactor     decimal.Decimal
}

// Position 持仓
type Position struct {
	Symbol        string          `json:"symbol"`
	Quantity      decimal.Decimal `json:"quantity"`
	AverageCost   decimal.Decimal `json:"average_cost"`
	CurrentPrice  decimal.Decimal `json:"current_price"`
	MarketValue   decimal.Decimal `json:"market_value"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL   decimal.Decimal `json:"realized_pnl"`
	OpenDate      time.Time       `json:"open_date"`
	Dividends     decimal.Decimal `json:"dividends"`
}

// Transaction 交易记录
type Transaction struct {
	ID         string          `json:"id"`
	Symbol     string          `json:"symbol"`
	Type       string          `json:"type"` // BUY, SELL, DIVIDEND
	Quantity   decimal.Decimal `json:"quantity"`
	Price      decimal.Decimal `json:"price"`
	Slippage   decimal.Decimal `json:"slippage"`
	Commission decimal.Decimal `json:"commission"`
	NetAmount  decimal.Decimal `json:"net_amount"`
	Timestamp  time.Time       `json:"timestamp"`
	Reason     string          `json:"reason"`
}

// EquityPoint 权益曲线点
type EquityPoint struct {
	Date          time.Time       `json:"date"`
	Equity        decimal.Decimal `json:"equity"`
	Cash          decimal.Decimal `json:"cash"`
	PositionValue decimal.Decimal `json:"position_value"`
	Drawdown      decimal.Decimal `json:"drawdown"`
	ReturnValue   decimal.Decimal `json:"return"`
}

// BacktestResult 回测结果
type BacktestResult struct {
	InitialCapital   decimal.Decimal    `json:"initial_capital"`
	FinalEquity      decimal.Decimal    `json:"final_equity"`
	TotalReturn      decimal.Decimal    `json:"total_return"`
	AnnualizedReturn decimal.Decimal    `json:"annualized_return"`
	MaxDrawdown      decimal.Decimal    `json:"max_drawdown"`
	SharpeRatio      decimal.Decimal    `json:"sharpe_ratio"`
	SortinoRatio     decimal.Decimal    `json:"sortino_ratio"`
	CalmarRatio      decimal.Decimal    `json:"calmar_ratio"`
	WinRate          decimal.Decimal    `json:"win_rate"`
	ProfitFactor     decimal.Decimal    `json:"profit_factor"`
	TotalTrades      int                `json:"total_trades"`
	WinningTrades    int                `json:"winning_trades"`
	LosingTrades     int                `json:"losing_trades"`
	AvgWin           decimal.Decimal    `json:"avg_win"`
	AvgLoss          decimal.Decimal    `json:"avg_loss"`
	Transactions     []*Transaction     `json:"transactions"`
	EquityCurve      []*EquityPoint     `json:"equity_curve"`
	FactorExposure   map[string]float64 `json:"factor_exposure"`
	StartDate        time.Time          `json:"start_date"`
	EndDate          time.Time          `json:"end_date"`
	DurationDays     int                `json:"duration_days"`
}

// NewBacktestEngine 创建回测引擎
func NewBacktestEngine(initialCapital float64, strategy Strategy) *BacktestEngine {
	return &BacktestEngine{
		initialCapital:  decimal.NewFromFloat(initialCapital),
		currentCapital:  decimal.NewFromFloat(initialCapital),
		positions:       make(map[string]*Position),
		transactions:    make([]*Transaction, 0),
		equityCurve:     make([]*EquityPoint, 0),
		factors:         make(map[string]decimal.Decimal),
		strategy:        strategy,
		slippageModel:   &DefaultSlippageModel{SlippageRate: decimal.NewFromFloat(0.001)},     // 默认0.1%滑点
		commissionModel: &DefaultCommissionModel{CommissionRate: decimal.NewFromFloat(0.001)}, // 默认0.1%手续费
		dividendModel:   &DefaultDividendModel{},
	}
}

// SetSlippageModel 设置滑点模型
func (e *BacktestEngine) SetSlippageModel(model SlippageModel) {
	e.slippageModel = model
}

// SetCommissionModel 设置手续费模型
func (e *BacktestEngine) SetCommissionModel(model CommissionModel) {
	e.commissionModel = model
}

// SetDividendModel 设置股息模型
func (e *BacktestEngine) SetDividendModel(model DividendModel) {
	e.dividendModel = model
}

// SetDataProvider 设置数据提供者
func (e *BacktestEngine) SetDataProvider(provider DataProvider) {
	e.dataProvider = provider
}

// Run 运行回测
func (e *BacktestEngine) Run(startDate, endDate time.Time) (*BacktestResult, error) {
	e.startDate = startDate
	e.endDate = endDate

	// 获取历史数据
	data, err := e.dataProvider.GetData(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取数据失败: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("没有可用的历史数据")
	}

	// 按日期排序
	sortDataByDate(data)

	// 事件驱动循环
	for _, bar := range data {
		// 更新持仓价格
		e.updatePositions(bar)

		// 计算因子
		e.calculateFactors(bar)

		// 策略信号生成
		signals := e.strategy.GenerateSignals(e, bar)

		// 执行交易
		for _, signal := range signals {
			e.executeSignal(signal, bar)
		}

		// 处理股息
		e.processDividends(bar)

		// 记录权益曲线
		e.recordEquity(bar.Date)
	}

	// 计算统计指标
	e.calculateStatistics()

	return e.generateResult(), nil
}

// executeSignal 执行交易信号
func (e *BacktestEngine) executeSignal(signal *Signal, bar *Bar) error {
	switch signal.Type {
	case SignalBuy:
		return e.executeBuy(signal, bar)
	case SignalSell:
		return e.executeSell(signal, bar)
	}
	return nil
}

// executeBuy 执行买入
func (e *BacktestEngine) executeBuy(signal *Signal, bar *Bar) error {
	// 计算滑点
	slippage := e.slippageModel.Calculate(bar.Close, signal.Quantity, true)
	executionPrice := bar.Close.Add(slippage)

	// 计算手续费
	commission := e.commissionModel.Calculate(executionPrice, signal.Quantity)

	// 计算总成本
	totalCost := executionPrice.Mul(signal.Quantity).Add(commission)

	// 检查资金是否足够
	if totalCost.GreaterThan(e.currentCapital) {
		return fmt.Errorf("资金不足: 需要 %s, 可用 %s", totalCost.String(), e.currentCapital.String())
	}

	// 更新持仓
	position, exists := e.positions[signal.Symbol]
	if !exists {
		position = &Position{
			Symbol:      signal.Symbol,
			Quantity:    decimal.Zero,
			AverageCost: decimal.Zero,
			OpenDate:    bar.Date,
		}
		e.positions[signal.Symbol] = position
	}

	// 更新平均成本
	totalQuantity := position.Quantity.Add(signal.Quantity)
	if totalQuantity.GreaterThan(decimal.Zero) {
		totalCostBasis := position.AverageCost.Mul(position.Quantity).Add(executionPrice.Mul(signal.Quantity))
		position.AverageCost = totalCostBasis.Div(totalQuantity)
	}
	position.Quantity = totalQuantity

	// 扣除资金
	e.currentCapital = e.currentCapital.Sub(totalCost)

	// 记录交易
	transaction := &Transaction{
		ID:         fmt.Sprintf("%s-%d", signal.Symbol, len(e.transactions)+1),
		Symbol:     signal.Symbol,
		Type:       "BUY",
		Quantity:   signal.Quantity,
		Price:      executionPrice,
		Slippage:   slippage.Mul(signal.Quantity),
		Commission: commission,
		NetAmount:  totalCost,
		Timestamp:  bar.Date,
		Reason:     signal.Reason,
	}
	e.transactions = append(e.transactions, transaction)

	return nil
}

// executeSell 执行卖出
func (e *BacktestEngine) executeSell(signal *Signal, bar *Bar) error {
	position, exists := e.positions[signal.Symbol]
	if !exists || position.Quantity.LessThan(signal.Quantity) {
		return fmt.Errorf("持仓不足: 需要 %s, 可用 %s", signal.Quantity.String(), position.Quantity.String())
	}

	// 计算滑点
	slippage := e.slippageModel.Calculate(bar.Close, signal.Quantity, false)
	executionPrice := bar.Close.Sub(slippage)

	// 计算手续费
	commission := e.commissionModel.Calculate(executionPrice, signal.Quantity)

	// 计算总收入
	totalRevenue := executionPrice.Mul(signal.Quantity).Sub(commission)

	// 计算已实现盈亏
	costBasis := position.AverageCost.Mul(signal.Quantity)
	realizedPnL := totalRevenue.Sub(costBasis)
	position.RealizedPnL = position.RealizedPnL.Add(realizedPnL)

	// 更新持仓
	position.Quantity = position.Quantity.Sub(signal.Quantity)
	if position.Quantity.IsZero() {
		delete(e.positions, signal.Symbol)
	}

	// 增加资金
	e.currentCapital = e.currentCapital.Add(totalRevenue)

	// 记录交易
	transaction := &Transaction{
		ID:         fmt.Sprintf("%s-%d", signal.Symbol, len(e.transactions)+1),
		Symbol:     signal.Symbol,
		Type:       "SELL",
		Quantity:   signal.Quantity,
		Price:      executionPrice,
		Slippage:   slippage.Mul(signal.Quantity),
		Commission: commission,
		NetAmount:  totalRevenue,
		Timestamp:  bar.Date,
		Reason:     signal.Reason,
	}
	e.transactions = append(e.transactions, transaction)

	return nil
}

// updatePositions 更新持仓价格
func (e *BacktestEngine) updatePositions(bar *Bar) {
	if position, exists := e.positions[bar.Symbol]; exists {
		position.CurrentPrice = bar.Close
		position.MarketValue = position.Quantity.Mul(bar.Close)
		position.UnrealizedPnL = position.MarketValue.Sub(position.AverageCost.Mul(position.Quantity))
	}
}

// processDividends 处理股息
func (e *BacktestEngine) processDividends(bar *Bar) {
	if bar.Dividend.IsZero() {
		return
	}

	for symbol, position := range e.positions {
		if symbol == bar.Symbol {
			dividendAmount := e.dividendModel.Calculate(bar.Dividend, position.Quantity)
			position.Dividends = position.Dividends.Add(dividendAmount)
			e.currentCapital = e.currentCapital.Add(dividendAmount)

			// 记录股息交易
			transaction := &Transaction{
				ID:        fmt.Sprintf("%s-DIV-%d", symbol, len(e.transactions)+1),
				Symbol:    symbol,
				Type:      "DIVIDEND",
				Quantity:  position.Quantity,
				Price:     bar.Dividend,
				NetAmount: dividendAmount,
				Timestamp: bar.Date,
				Reason:    "股息收入",
			}
			e.transactions = append(e.transactions, transaction)
		}
	}
}

// recordEquity 记录权益曲线
func (e *BacktestEngine) recordEquity(date time.Time) {
	positionValue := decimal.Zero
	for _, position := range e.positions {
		positionValue = positionValue.Add(position.MarketValue)
	}

	totalEquity := e.currentCapital.Add(positionValue)

	// 计算收益率和回撤
	returnValue := decimal.Zero
	drawdown := decimal.Zero
	if len(e.equityCurve) > 0 {
		prevEquity := e.equityCurve[len(e.equityCurve)-1].Equity
		if prevEquity.GreaterThan(decimal.Zero) {
			returnValue = totalEquity.Sub(prevEquity).Div(prevEquity)
		}
	}

	// 计算从峰值到当前的回撤
	peak := e.initialCapital
	for _, point := range e.equityCurve {
		if point.Equity.GreaterThan(peak) {
			peak = point.Equity
		}
	}
	if totalEquity.GreaterThan(peak) {
		peak = totalEquity
	}
	if peak.GreaterThan(decimal.Zero) {
		drawdown = peak.Sub(totalEquity).Div(peak)
	}

	point := &EquityPoint{
		Date:          date,
		Equity:        totalEquity,
		Cash:          e.currentCapital,
		PositionValue: positionValue,
		Drawdown:      drawdown,
		ReturnValue:   returnValue,
	}
	e.equityCurve = append(e.equityCurve, point)
}

// calculateStatistics 计算统计指标
func (e *BacktestEngine) calculateStatistics() {
	if len(e.equityCurve) == 0 {
		return
	}

	// 总收益率
	finalEquity := e.equityCurve[len(e.equityCurve)-1].Equity
	e.totalReturn = finalEquity.Sub(e.initialCapital).Div(e.initialCapital).Mul(decimal.NewFromInt(100))

	// 年化收益率
	days := decimal.NewFromFloat(float64(e.equityCurve[len(e.equityCurve)-1].Date.Sub(e.equityCurve[0].Date).Hours() / 24))
	if days.GreaterThan(decimal.Zero) {
		years := days.Div(decimal.NewFromInt(365))
		if years.GreaterThan(decimal.Zero) {
			e.annualizedReturn = decimal.NewFromFloat(math.Pow(finalEquity.Div(e.initialCapital).InexactFloat64(), 1/years.InexactFloat64()) - 1)
			e.annualizedReturn = e.annualizedReturn.Mul(decimal.NewFromInt(100))
		}
	}

	// 最大回撤
	e.maxDrawdown = decimal.Zero
	for _, point := range e.equityCurve {
		if point.Drawdown.GreaterThan(e.maxDrawdown) {
			e.maxDrawdown = point.Drawdown
		}
	}

	// 计算夏普比率、索提诺比率
	returns := make([]float64, len(e.equityCurve)-1)
	for i := 1; i < len(e.equityCurve); i++ {
		returns[i-1] = e.equityCurve[i].ReturnValue.InexactFloat64()
	}
	e.sharpeRatio = calculateSharpeRatio(returns)
	e.sortinoRatio = calculateSortinoRatio(returns)

	// 卡尔玛比率
	if e.maxDrawdown.GreaterThan(decimal.Zero) {
		e.calmarRatio = e.annualizedReturn.Div(e.maxDrawdown.Mul(decimal.NewFromInt(100)))
	}

	// 胜率和盈亏比
	e.calculateTradeStatistics()
}

// calculateTradeStatistics 计算交易统计
func (e *BacktestEngine) calculateTradeStatistics() {
	winningTrades := 0
	losingTrades := 0
	totalWin := decimal.Zero
	totalLoss := decimal.Zero

	for _, tx := range e.transactions {
		if tx.Type == "SELL" {
			// 计算单笔交易盈亏
			// 简化计算：使用净金额与成本的差异
			if tx.NetAmount.GreaterThan(decimal.Zero) {
				winningTrades++
				totalWin = totalWin.Add(tx.NetAmount)
			} else {
				losingTrades++
				totalLoss = totalLoss.Add(tx.NetAmount.Abs())
			}
		}
	}

	totalTrades := winningTrades + losingTrades
	if totalTrades > 0 {
		e.winRate = decimal.NewFromInt(int64(winningTrades)).Div(decimal.NewFromInt(int64(totalTrades))).Mul(decimal.NewFromInt(100))
	}

	if totalLoss.GreaterThan(decimal.Zero) {
		e.profitFactor = totalWin.Div(totalLoss)
	}
}

// calculateFactors 计算因子
func (e *BacktestEngine) calculateFactors(bar *Bar) {
	// 由策略实现具体的因子计算
	if factorStrategy, ok := e.strategy.(FactorStrategy); ok {
		factors := factorStrategy.CalculateFactors(e, bar)
		for name, value := range factors {
			e.factors[name] = value
		}
	}
}

// generateResult 生成回测结果
func (e *BacktestEngine) generateResult() *BacktestResult {
	// 计算因子暴露
	factorExposure := make(map[string]float64)
	for name, value := range e.factors {
		factorExposure[name] = value.InexactFloat64()
	}

	return &BacktestResult{
		InitialCapital:   e.initialCapital,
		FinalEquity:      e.equityCurve[len(e.equityCurve)-1].Equity,
		TotalReturn:      e.totalReturn,
		AnnualizedReturn: e.annualizedReturn,
		MaxDrawdown:      e.maxDrawdown.Mul(decimal.NewFromInt(100)),
		SharpeRatio:      e.sharpeRatio,
		SortinoRatio:     e.sortinoRatio,
		CalmarRatio:      e.calmarRatio,
		WinRate:          e.winRate,
		ProfitFactor:     e.profitFactor,
		TotalTrades:      len(e.transactions),
		Transactions:     e.transactions,
		EquityCurve:      e.equityCurve,
		FactorExposure:   factorExposure,
		StartDate:        e.startDate,
		EndDate:          e.endDate,
		DurationDays:     int(e.endDate.Sub(e.startDate).Hours() / 24),
	}
}

// GetPosition 获取持仓
func (e *BacktestEngine) GetPosition(symbol string) *Position {
	return e.positions[symbol]
}

// GetAllPositions 获取所有持仓
func (e *BacktestEngine) GetAllPositions() map[string]*Position {
	return e.positions
}

// GetCurrentCapital 获取当前资金
func (e *BacktestEngine) GetCurrentCapital() decimal.Decimal {
	return e.currentCapital
}

// GetEquityCurve 获取权益曲线
func (e *BacktestEngine) GetEquityCurve() []*EquityPoint {
	return e.equityCurve
}

// GetFactor 获取因子值
func (e *BacktestEngine) GetFactor(name string) decimal.Decimal {
	return e.factors[name]
}

// 辅助函数

func sortDataByDate(data []*Bar) {
	// 使用标准库的快速排序，性能更好 O(n log n)
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date.Before(data[j].Date)
	})
}

func calculateSharpeRatio(returns []float64) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	// 计算平均收益率
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// 计算标准差
	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	std := math.Sqrt(variance / float64(len(returns)))

	if std == 0 {
		return decimal.Zero
	}

	// 年化夏普比率 (假设252个交易日)
	riskFreeRate := 0.045 / 252.0 // 日化无风险利率
	sharpe := (mean - riskFreeRate) / std * math.Sqrt(252)

	return decimal.NewFromFloat(sharpe)
}

func calculateSortinoRatio(returns []float64) decimal.Decimal {
	if len(returns) == 0 {
		return decimal.Zero
	}

	// 计算平均收益率
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// 计算下行偏差
	targetReturn := 0.0
	downsideSquares := 0.0
	for _, r := range returns {
		if r < targetReturn {
			downsideSquares += (r - targetReturn) * (r - targetReturn)
		}
	}
	downsideStd := math.Sqrt(downsideSquares / float64(len(returns)))

	if downsideStd == 0 {
		return decimal.Zero
	}

	// 年化索提诺比率
	riskFreeRate := 0.045 / 252.0
	sortino := (mean - riskFreeRate) / downsideStd * math.Sqrt(252)

	return decimal.NewFromFloat(sortino)
}
