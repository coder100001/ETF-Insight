package backtest

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// EventType 事件类型
type EventType string

const (
	EventMarketData EventType = "MARKET_DATA" // 市场数据事件
	EventSignal     EventType = "SIGNAL"      // 交易信号事件
	EventOrder      EventType = "ORDER"       // 订单事件
	EventFill       EventType = "FILL"        // 成交事件
	EventPortfolio  EventType = "PORTFOLIO"   // 组合更新事件
	EventDividend   EventType = "DIVIDEND"    // 股息事件
	EventSplit      EventType = "SPLIT"       // 拆股事件
	EventRebalance  EventType = "REBALANCE"   // 再平衡事件
	EventStopLoss   EventType = "STOP_LOSS"   // 止损事件
	EventTakeProfit EventType = "TAKE_PROFIT" // 止盈事件
)

// Event 事件接口
type Event interface {
	GetType() EventType
	GetTimestamp() time.Time
	GetSymbol() string
}

// BaseEvent 基础事件
type BaseEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Symbol    string    `json:"symbol"`
}

func (e *BaseEvent) GetType() EventType      { return e.Type }
func (e *BaseEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *BaseEvent) GetSymbol() string       { return e.Symbol }

// MarketDataEvent 市场数据事件
type MarketDataEvent struct {
	BaseEvent
	Bar *Bar `json:"bar"`
}

// SignalEvent 信号事件
type SignalEvent struct {
	BaseEvent
	SignalType   SignalType      `json:"signal_type"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	StopPrice    decimal.Decimal `json:"stop_price"`  // 止损价
	LimitPrice   decimal.Decimal `json:"limit_price"` // 限价
	Reason       string          `json:"reason"`
	StrategyName string          `json:"strategy_name"`
}

// OrderType 订单类型
type OrderType string

const (
	OrderMarket    OrderType = "MARKET"     // 市价单
	OrderLimit     OrderType = "LIMIT"      // 限价单
	OrderStop      OrderType = "STOP"       // 止损单
	OrderStopLimit OrderType = "STOP_LIMIT" // 止损限价单
	OrderTrailing  OrderType = "TRAILING"   // 跟踪止损单
)

// OrderSide 订单方向
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"   // 待处理
	OrderStatusSubmitted OrderStatus = "SUBMITTED" // 已提交
	OrderStatusPartial   OrderStatus = "PARTIAL"   // 部分成交
	OrderStatusFilled    OrderStatus = "FILLED"    // 完全成交
	OrderStatusCancelled OrderStatus = "CANCELLED" // 已取消
	OrderStatusRejected  OrderStatus = "REJECTED"  // 已拒绝
)

// Order 订单
type Order struct {
	ID           string          `json:"id"`
	Symbol       string          `json:"symbol"`
	Type         OrderType       `json:"type"`
	Side         OrderSide       `json:"side"`
	Quantity     decimal.Decimal `json:"quantity"`
	FilledQty    decimal.Decimal `json:"filled_qty"`
	Price        decimal.Decimal `json:"price"`       // 委托价格
	StopPrice    decimal.Decimal `json:"stop_price"`  // 止损价格
	LimitPrice   decimal.Decimal `json:"limit_price"` // 限价
	Status       OrderStatus     `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	FilledAt     *time.Time      `json:"filled_at,omitempty"`
	Reason       string          `json:"reason"`
	StrategyName string          `json:"strategy_name"`
}

// OrderEvent 订单事件
type OrderEvent struct {
	BaseEvent
	Order *Order `json:"order"`
}

// Fill 成交记录
type Fill struct {
	OrderID      string          `json:"order_id"`
	Symbol       string          `json:"symbol"`
	Side         OrderSide       `json:"side"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Slippage     decimal.Decimal `json:"slippage"`
	Commission   decimal.Decimal `json:"commission"`
	Timestamp    time.Time       `json:"timestamp"`
	StrategyName string          `json:"strategy_name"`
}

// FillEvent 成交事件
type FillEvent struct {
	BaseEvent
	Fill *Fill `json:"fill"`
}

// PortfolioEvent 组合事件
type PortfolioEvent struct {
	BaseEvent
	Positions     map[string]*Position `json:"positions"`
	Cash          decimal.Decimal      `json:"cash"`
	TotalValue    decimal.Decimal      `json:"total_value"`
	UnrealizedPnL decimal.Decimal      `json:"unrealized_pnl"`
	RealizedPnL   decimal.Decimal      `json:"realized_pnl"`
}

// DividendEvent 股息事件
type DividendEvent struct {
	BaseEvent
	DividendPerShare decimal.Decimal `json:"dividend_per_share"`
	TotalDividend    decimal.Decimal `json:"total_dividend"`
	PositionQty      decimal.Decimal `json:"position_qty"`
}

// SplitEvent 拆股事件
type SplitEvent struct {
	BaseEvent
	Ratio       decimal.Decimal `json:"ratio"` // 拆股比例 (如 2:1 则为 2)
	OldQuantity decimal.Decimal `json:"old_quantity"`
	NewQuantity decimal.Decimal `json:"new_quantity"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(event Event) error
}

// EventBus 事件总线
type EventBus struct {
	handlers map[EventType][]EventHandler
	queue    chan Event
	running  bool
}

// NewEventBus 创建事件总线
func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
		queue:    make(chan Event, bufferSize),
		running:  false,
	}
}

// Register 注册事件处理器
func (eb *EventBus) Register(eventType EventType, handler EventHandler) {
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Emit 发送事件
func (eb *EventBus) Emit(event Event) {
	select {
	case eb.queue <- event:
	default:
		// 队列满，丢弃事件或记录日志
		fmt.Printf("Event queue full, dropping event: %v\n", event.GetType())
	}
}

// EmitSync 同步发送事件（立即处理）
func (eb *EventBus) EmitSync(event Event) error {
	return eb.processEvent(event)
}

// processEvent 处理单个事件
func (eb *EventBus) processEvent(event Event) error {
	handlers := eb.handlers[event.GetType()]
	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			return fmt.Errorf("event handler error: %w", err)
		}
	}
	return nil
}

// Start 启动事件循环
func (eb *EventBus) Start() {
	eb.running = true
	go eb.eventLoop()
}

// Stop 停止事件循环
func (eb *EventBus) Stop() {
	eb.running = false
	close(eb.queue)
}

// eventLoop 事件循环
func (eb *EventBus) eventLoop() {
	for eb.running {
		select {
		case event, ok := <-eb.queue:
			if !ok {
				return
			}
			if err := eb.processEvent(event); err != nil {
				fmt.Printf("Error processing event: %v\n", err)
			}
		}
	}
}

// EventDrivenEngine 事件驱动回测引擎
type EventDrivenEngine struct {
	initialCapital decimal.Decimal
	currentCapital decimal.Decimal
	positions      map[string]*Position
	orders         map[string]*Order
	pendingOrders  []*Order
	fills          []*Fill
	equityCurve    []*EquityPoint

	// 事件系统
	eventBus *EventBus

	// 模型配置
	slippageModel   SlippageModel
	commissionModel CommissionModel
	dividendModel   DividendModel

	// 策略
	strategy Strategy

	// 数据
	dataProvider DataProvider

	// 统计
	startDate   time.Time
	endDate     time.Time
	currentDate time.Time
	peakEquity  decimal.Decimal
	maxDrawdown decimal.Decimal

	// 风控
	stopLossEnabled   bool
	stopLossPercent   decimal.Decimal
	takeProfitEnabled bool
	takeProfitPercent decimal.Decimal

	// 再平衡
	rebalanceEnabled  bool
	rebalanceInterval int // 再平衡间隔天数
	lastRebalanceDate time.Time
}

// NewEventDrivenEngine 创建事件驱动回测引擎
func NewEventDrivenEngine(initialCapital float64, strategy Strategy) *EventDrivenEngine {
	capital := decimal.NewFromFloat(initialCapital)
	return &EventDrivenEngine{
		initialCapital:    capital,
		currentCapital:    capital,
		positions:         make(map[string]*Position),
		orders:            make(map[string]*Order),
		pendingOrders:     make([]*Order, 0),
		fills:             make([]*Fill, 0),
		equityCurve:       make([]*EquityPoint, 0),
		eventBus:          NewEventBus(1000),
		strategy:          strategy,
		slippageModel:     &DefaultSlippageModel{SlippageRate: decimal.NewFromFloat(0.001)},
		commissionModel:   &DefaultCommissionModel{CommissionRate: decimal.NewFromFloat(0.001)},
		dividendModel:     &DefaultDividendModel{},
		peakEquity:        capital,
		stopLossEnabled:   false,
		stopLossPercent:   decimal.NewFromFloat(0.05),
		takeProfitEnabled: false,
		takeProfitPercent: decimal.NewFromFloat(0.10),
		rebalanceEnabled:  false,
		rebalanceInterval: 30,
	}
}

// SetSlippageModel 设置滑点模型
func (e *EventDrivenEngine) SetSlippageModel(model SlippageModel) {
	e.slippageModel = model
}

// SetCommissionModel 设置手续费模型
func (e *EventDrivenEngine) SetCommissionModel(model CommissionModel) {
	e.commissionModel = model
}

// SetDividendModel 设置股息模型
func (e *EventDrivenEngine) SetDividendModel(model DividendModel) {
	e.dividendModel = model
}

// SetDataProvider 设置数据提供者
func (e *EventDrivenEngine) SetDataProvider(provider DataProvider) {
	e.dataProvider = provider
}

// SetStopLoss 设置止损
func (e *EventDrivenEngine) SetStopLoss(enabled bool, percent float64) {
	e.stopLossEnabled = enabled
	e.stopLossPercent = decimal.NewFromFloat(percent)
}

// SetTakeProfit 设置止盈
func (e *EventDrivenEngine) SetTakeProfit(enabled bool, percent float64) {
	e.takeProfitEnabled = enabled
	e.takeProfitPercent = decimal.NewFromFloat(percent)
}

// SetRebalance 设置再平衡
func (e *EventDrivenEngine) SetRebalance(enabled bool, interval int) {
	e.rebalanceEnabled = enabled
	e.rebalanceInterval = interval
}

// Run 运行回测
func (e *EventDrivenEngine) Run(startDate, endDate time.Time) (*EventDrivenBacktestResult, error) {
	e.startDate = startDate
	e.endDate = endDate
	e.currentDate = startDate

	// 启动事件总线
	e.eventBus.Start()
	defer e.eventBus.Stop()

	// 注册事件处理器
	e.registerEventHandlers()

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
		e.currentDate = bar.Date

		// 处理待执行订单
		e.processPendingOrders(bar)

		// 检查止损止盈
		e.checkStopLossTakeProfit(bar)

		// 检查再平衡
		e.checkRebalance(bar)

		// 发送市场数据事件
		marketEvent := &MarketDataEvent{
			BaseEvent: BaseEvent{
				Type:      EventMarketData,
				Timestamp: bar.Date,
				Symbol:    bar.Symbol,
			},
			Bar: bar,
		}
		e.eventBus.EmitSync(marketEvent)

		// 更新持仓价格
		e.updatePositions(bar)

		// 处理股息
		e.processDividends(bar)

		// 处理拆股
		e.processSplits(bar)

		// 策略信号生成
		signals := e.strategy.GenerateSignals(&BacktestEngine{}, bar)

		// 将信号转换为订单
		for _, signal := range signals {
			e.signalToOrder(signal, bar)
		}

		// 记录权益曲线
		e.recordEquity(bar.Date)
	}

	// 计算统计指标
	e.calculateStatistics()

	return e.generateResult(), nil
}

// Run 运行回测 (兼容旧接口)
func (e *EventDrivenEngine) RunBacktest(startDate, endDate time.Time) (*BacktestResult, error) {
	result, err := e.Run(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 转换为旧格式
	return &BacktestResult{
		InitialCapital:   result.InitialCapital,
		FinalEquity:      result.FinalEquity,
		TotalReturn:      result.TotalReturn,
		AnnualizedReturn: result.AnnualizedReturn,
		MaxDrawdown:      result.MaxDrawdown,
		SharpeRatio:      result.SharpeRatio,
		SortinoRatio:     result.SortinoRatio,
		CalmarRatio:      result.CalmarRatio,
		WinRate:          result.WinRate,
		ProfitFactor:     result.ProfitFactor,
		TotalTrades:      result.TotalTrades,
		WinningTrades:    result.WinningTrades,
		LosingTrades:     result.LosingTrades,
		Transactions:     result.Transactions,
		EquityCurve:      result.EquityCurve,
		StartDate:        result.StartDate,
		EndDate:          result.EndDate,
		DurationDays:     result.DurationDays,
	}, nil
}

// registerEventHandlers 注册事件处理器
func (e *EventDrivenEngine) registerEventHandlers() {
	// 订单事件处理器
	e.eventBus.Register(EventOrder, &OrderEventHandler{engine: e})
	// 成交事件处理器
	e.eventBus.Register(EventFill, &FillEventHandler{engine: e})
}

// signalToOrder 将信号转换为订单
func (e *EventDrivenEngine) signalToOrder(signal *Signal, bar *Bar) {
	var side OrderSide
	var orderType OrderType

	switch signal.Type {
	case SignalBuy:
		side = OrderSideBuy
		orderType = OrderMarket
	case SignalSell:
		side = OrderSideSell
		orderType = OrderMarket
	default:
		return
	}

	// 检查是否有足够资金/持仓
	if side == OrderSideBuy {
		estimatedCost := bar.Close.Mul(signal.Quantity)
		if estimatedCost.GreaterThan(e.currentCapital) {
			return // 资金不足
		}
	} else {
		position, exists := e.positions[signal.Symbol]
		if !exists || position.Quantity.LessThan(signal.Quantity) {
			return // 持仓不足
		}
	}

	order := &Order{
		ID:           fmt.Sprintf("%s-%d", signal.Symbol, len(e.orders)+1),
		Symbol:       signal.Symbol,
		Type:         orderType,
		Side:         side,
		Quantity:     signal.Quantity,
		FilledQty:    decimal.Zero,
		Price:        signal.Price,
		Status:       OrderStatusPending,
		CreatedAt:    bar.Date,
		UpdatedAt:    bar.Date,
		Reason:       signal.Reason,
		StrategyName: e.strategy.GetName(),
	}

	e.orders[order.ID] = order
	e.pendingOrders = append(e.pendingOrders, order)

	// 发送订单事件
	orderEvent := &OrderEvent{
		BaseEvent: BaseEvent{
			Type:      EventOrder,
			Timestamp: bar.Date,
			Symbol:    signal.Symbol,
		},
		Order: order,
	}
	e.eventBus.Emit(orderEvent)
}

// processPendingOrders 处理待执行订单
func (e *EventDrivenEngine) processPendingOrders(bar *Bar) {
	remainingOrders := make([]*Order, 0)

	for _, order := range e.pendingOrders {
		if order.Symbol != bar.Symbol {
			remainingOrders = append(remainingOrders, order)
			continue
		}

		// 检查订单是否可以执行
		canExecute := false
		executionPrice := bar.Close

		switch order.Type {
		case OrderMarket:
			canExecute = true
		case OrderLimit:
			if order.Side == OrderSideBuy && bar.Low.LessThanOrEqual(order.LimitPrice) {
				canExecute = true
				executionPrice = order.LimitPrice
			} else if order.Side == OrderSideSell && bar.High.GreaterThanOrEqual(order.LimitPrice) {
				canExecute = true
				executionPrice = order.LimitPrice
			}
		case OrderStop:
			if order.Side == OrderSideBuy && bar.High.GreaterThanOrEqual(order.StopPrice) {
				canExecute = true
				executionPrice = order.StopPrice
			} else if order.Side == OrderSideSell && bar.Low.LessThanOrEqual(order.StopPrice) {
				canExecute = true
				executionPrice = order.StopPrice
			}
		}

		if canExecute {
			e.executeOrder(order, executionPrice, bar.Date)
		} else {
			remainingOrders = append(remainingOrders, order)
		}
	}

	e.pendingOrders = remainingOrders
}

// executeOrder 执行订单
func (e *EventDrivenEngine) executeOrder(order *Order, price decimal.Decimal, timestamp time.Time) {
	// 计算滑点
	isBuy := order.Side == OrderSideBuy
	slippage := e.slippageModel.Calculate(price, order.Quantity, isBuy)

	var executionPrice decimal.Decimal
	if isBuy {
		executionPrice = price.Add(slippage)
	} else {
		executionPrice = price.Sub(slippage)
	}

	// 计算手续费
	commission := e.commissionModel.Calculate(executionPrice, order.Quantity)

	// 更新订单状态
	order.Status = OrderStatusFilled
	order.FilledQty = order.Quantity
	order.UpdatedAt = timestamp
	order.FilledAt = &timestamp

	// 创建成交记录
	fill := &Fill{
		OrderID:      order.ID,
		Symbol:       order.Symbol,
		Side:         order.Side,
		Quantity:     order.Quantity,
		Price:        executionPrice,
		Slippage:     slippage,
		Commission:   commission,
		Timestamp:    timestamp,
		StrategyName: order.StrategyName,
	}
	e.fills = append(e.fills, fill)

	// 更新持仓和资金
	if isBuy {
		e.executeBuy(order.Symbol, order.Quantity, executionPrice, commission, timestamp)
	} else {
		e.executeSell(order.Symbol, order.Quantity, executionPrice, commission, timestamp)
	}

	// 发送成交事件
	fillEvent := &FillEvent{
		BaseEvent: BaseEvent{
			Type:      EventFill,
			Timestamp: timestamp,
			Symbol:    order.Symbol,
		},
		Fill: fill,
	}
	e.eventBus.Emit(fillEvent)
}

// executeBuy 执行买入
func (e *EventDrivenEngine) executeBuy(symbol string, quantity, price, commission decimal.Decimal, timestamp time.Time) {
	totalCost := price.Mul(quantity).Add(commission)

	// 更新持仓
	position, exists := e.positions[symbol]
	if !exists {
		position = &Position{
			Symbol:      symbol,
			Quantity:    decimal.Zero,
			AverageCost: decimal.Zero,
			OpenDate:    timestamp,
		}
		e.positions[symbol] = position
	}

	// 更新平均成本
	totalQuantity := position.Quantity.Add(quantity)
	if totalQuantity.GreaterThan(decimal.Zero) {
		totalCostBasis := position.AverageCost.Mul(position.Quantity).Add(price.Mul(quantity))
		position.AverageCost = totalCostBasis.Div(totalQuantity)
	}
	position.Quantity = totalQuantity

	// 扣除资金
	e.currentCapital = e.currentCapital.Sub(totalCost)
}

// executeSell 执行卖出
func (e *EventDrivenEngine) executeSell(symbol string, quantity, price, commission decimal.Decimal, timestamp time.Time) {
	position, exists := e.positions[symbol]
	if !exists || position.Quantity.LessThan(quantity) {
		return
	}

	totalRevenue := price.Mul(quantity).Sub(commission)

	// 计算已实现盈亏
	costBasis := position.AverageCost.Mul(quantity)
	realizedPnL := totalRevenue.Sub(costBasis)
	position.RealizedPnL = position.RealizedPnL.Add(realizedPnL)

	// 更新持仓
	position.Quantity = position.Quantity.Sub(quantity)
	if position.Quantity.IsZero() {
		delete(e.positions, symbol)
	}

	// 增加资金
	e.currentCapital = e.currentCapital.Add(totalRevenue)
}

// checkStopLossTakeProfit 检查止损止盈
func (e *EventDrivenEngine) checkStopLossTakeProfit(bar *Bar) {
	if position, exists := e.positions[bar.Symbol]; exists {
		if position.AverageCost.IsZero() {
			return
		}

		currentPrice := bar.Close
		costBasis := position.AverageCost

		// 计算当前盈亏比例
		var pnlPercent decimal.Decimal
		if costBasis.GreaterThan(decimal.Zero) {
			pnlPercent = currentPrice.Sub(costBasis).Div(costBasis)
		}

		// 止损检查
		if e.stopLossEnabled && pnlPercent.LessThan(decimal.Zero) {
			lossPercent := pnlPercent.Abs()
			if lossPercent.GreaterThanOrEqual(e.stopLossPercent) {
				// 触发止损
				stopOrder := &Order{
					ID:        fmt.Sprintf("SL-%s-%d", bar.Symbol, len(e.orders)+1),
					Symbol:    bar.Symbol,
					Type:      OrderMarket,
					Side:      OrderSideSell,
					Quantity:  position.Quantity,
					Status:    OrderStatusPending,
					CreatedAt: bar.Date,
					UpdatedAt: bar.Date,
					Reason:    "止损触发",
				}
				e.orders[stopOrder.ID] = stopOrder
				e.pendingOrders = append(e.pendingOrders, stopOrder)
			}
		}

		// 止盈检查
		if e.takeProfitEnabled && pnlPercent.GreaterThan(decimal.Zero) {
			if pnlPercent.GreaterThanOrEqual(e.takeProfitPercent) {
				// 触发止盈
				tpOrder := &Order{
					ID:        fmt.Sprintf("TP-%s-%d", bar.Symbol, len(e.orders)+1),
					Symbol:    bar.Symbol,
					Type:      OrderMarket,
					Side:      OrderSideSell,
					Quantity:  position.Quantity,
					Status:    OrderStatusPending,
					CreatedAt: bar.Date,
					UpdatedAt: bar.Date,
					Reason:    "止盈触发",
				}
				e.orders[tpOrder.ID] = tpOrder
				e.pendingOrders = append(e.pendingOrders, tpOrder)
			}
		}
	}
}

// checkRebalance 检查再平衡
func (e *EventDrivenEngine) checkRebalance(bar *Bar) {
	if !e.rebalanceEnabled {
		return
	}

	if e.lastRebalanceDate.IsZero() {
		e.lastRebalanceDate = bar.Date
		return
	}

	daysSinceLastRebalance := int(bar.Date.Sub(e.lastRebalanceDate).Hours() / 24)
	if daysSinceLastRebalance >= e.rebalanceInterval {
		// 触发再平衡事件
		rebalanceEvent := &BaseEvent{
			Type:      EventRebalance,
			Timestamp: bar.Date,
			Symbol:    "PORTFOLIO",
		}
		e.eventBus.Emit(rebalanceEvent)
		e.lastRebalanceDate = bar.Date
	}
}

// updatePositions 更新持仓价格
func (e *EventDrivenEngine) updatePositions(bar *Bar) {
	if position, exists := e.positions[bar.Symbol]; exists {
		position.CurrentPrice = bar.Close
		position.MarketValue = position.Quantity.Mul(bar.Close)
		position.UnrealizedPnL = position.MarketValue.Sub(position.AverageCost.Mul(position.Quantity))
	}
}

// processDividends 处理股息
func (e *EventDrivenEngine) processDividends(bar *Bar) {
	if bar.Dividend.IsZero() {
		return
	}

	for symbol, position := range e.positions {
		if symbol == bar.Symbol {
			dividendAmount := e.dividendModel.Calculate(bar.Dividend, position.Quantity)
			position.Dividends = position.Dividends.Add(dividendAmount)
			e.currentCapital = e.currentCapital.Add(dividendAmount)

			// 发送股息事件
			dividendEvent := &DividendEvent{
				BaseEvent: BaseEvent{
					Type:      EventDividend,
					Timestamp: bar.Date,
					Symbol:    symbol,
				},
				DividendPerShare: bar.Dividend,
				TotalDividend:    dividendAmount,
				PositionQty:      position.Quantity,
			}
			e.eventBus.Emit(dividendEvent)
		}
	}
}

// processSplits 处理拆股
func (e *EventDrivenEngine) processSplits(bar *Bar) {
	if bar.Split.IsZero() || bar.Split.Equal(decimal.NewFromInt(1)) {
		return
	}

	for symbol, position := range e.positions {
		if symbol == bar.Symbol {
			oldQuantity := position.Quantity
			newQuantity := oldQuantity.Mul(bar.Split)
			position.Quantity = newQuantity
			position.AverageCost = position.AverageCost.Div(bar.Split)

			// 发送拆股事件
			splitEvent := &SplitEvent{
				BaseEvent: BaseEvent{
					Type:      EventSplit,
					Timestamp: bar.Date,
					Symbol:    symbol,
				},
				Ratio:       bar.Split,
				OldQuantity: oldQuantity,
				NewQuantity: newQuantity,
			}
			e.eventBus.Emit(splitEvent)
		}
	}
}

// recordEquity 记录权益曲线
func (e *EventDrivenEngine) recordEquity(date time.Time) {
	positionValue := decimal.Zero
	for _, position := range e.positions {
		positionValue = positionValue.Add(position.MarketValue)
	}

	totalEquity := e.currentCapital.Add(positionValue)

	// 更新峰值权益
	if totalEquity.GreaterThan(e.peakEquity) {
		e.peakEquity = totalEquity
	}

	// 计算回撤
	drawdown := decimal.Zero
	if e.peakEquity.GreaterThan(decimal.Zero) {
		drawdown = e.peakEquity.Sub(totalEquity).Div(e.peakEquity)
	}
	if drawdown.GreaterThan(e.maxDrawdown) {
		e.maxDrawdown = drawdown
	}

	// 计算收益率
	returnValue := decimal.Zero
	if len(e.equityCurve) > 0 {
		prevEquity := e.equityCurve[len(e.equityCurve)-1].Equity
		if prevEquity.GreaterThan(decimal.Zero) {
			returnValue = totalEquity.Sub(prevEquity).Div(prevEquity)
		}
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
func (e *EventDrivenEngine) calculateStatistics() {
	// 已在记录权益曲线时计算
}

// EventDrivenBacktestResult 事件驱动回测结果
type EventDrivenBacktestResult struct {
	InitialCapital   decimal.Decimal `json:"initial_capital"`
	FinalEquity      decimal.Decimal `json:"final_equity"`
	TotalReturn      decimal.Decimal `json:"total_return"`
	AnnualizedReturn decimal.Decimal `json:"annualized_return"`
	MaxDrawdown      decimal.Decimal `json:"max_drawdown"`
	SharpeRatio      decimal.Decimal `json:"sharpe_ratio"`
	SortinoRatio     decimal.Decimal `json:"sortino_ratio"`
	CalmarRatio      decimal.Decimal `json:"calmar_ratio"`
	WinRate          decimal.Decimal `json:"win_rate"`
	ProfitFactor     decimal.Decimal `json:"profit_factor"`
	TotalTrades      int             `json:"total_trades"`
	WinningTrades    int             `json:"winning_trades"`
	LosingTrades     int             `json:"losing_trades"`
	Transactions     []*Transaction  `json:"transactions"`
	EquityCurve      []*EquityPoint  `json:"equity_curve"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	DurationDays     int             `json:"duration_days"`
}

// generateResult 生成回测结果
func (e *EventDrivenEngine) generateResult() *EventDrivenBacktestResult {
	if len(e.equityCurve) == 0 {
		return &EventDrivenBacktestResult{
			InitialCapital: e.initialCapital,
			FinalEquity:    e.currentCapital,
			TotalReturn:    decimal.Zero,
		}
	}

	finalEquity := e.equityCurve[len(e.equityCurve)-1].Equity
	totalReturn := finalEquity.Sub(e.initialCapital).Div(e.initialCapital).Mul(decimal.NewFromInt(100))

	// 计算年化收益率
	durationDays := int(e.endDate.Sub(e.startDate).Hours() / 24)
	var annualizedReturn decimal.Decimal
	if durationDays > 0 && e.initialCapital.GreaterThan(decimal.Zero) {
		years := decimal.NewFromInt(int64(durationDays)).Div(decimal.NewFromInt(365))
		if years.GreaterThan(decimal.Zero) {
			growth := finalEquity.Div(e.initialCapital)
			_ = decimal.NewFromFloat(1).Div(years)
			// 简化计算: (1 + total_return)^(1/years) - 1
			annualizedReturn = growth.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
		}
	}

	// 计算夏普比率等需要收益率序列的指标
	returns := make([]float64, 0)
	for _, point := range e.equityCurve {
		if point.ReturnValue.Abs().GreaterThan(decimal.Zero) {
			returns = append(returns, point.ReturnValue.InexactFloat64())
		}
	}

	sharpeRatio := decimal.NewFromFloat(calculateSharpeRatioFloat(returns))
	sortinoRatio := decimal.NewFromFloat(calculateSortinoRatioFloat(returns))

	// 计算胜率
	winningTrades := 0
	losingTrades := 0
	var totalWin, totalLoss decimal.Decimal

	for _, fill := range e.fills {
		// 简化处理，实际需要根据持仓变化计算
		if fill.Side == OrderSideSell {
			// 假设卖出时计算盈亏
			if position, exists := e.positions[fill.Symbol]; exists {
				if position.RealizedPnL.GreaterThan(decimal.Zero) {
					winningTrades++
					totalWin = totalWin.Add(position.RealizedPnL)
				} else {
					losingTrades++
					totalLoss = totalLoss.Add(position.RealizedPnL.Abs())
				}
			}
		}
	}

	totalTrades := winningTrades + losingTrades
	var winRate decimal.Decimal
	if totalTrades > 0 {
		winRate = decimal.NewFromInt(int64(winningTrades)).Div(decimal.NewFromInt(int64(totalTrades))).Mul(decimal.NewFromInt(100))
	}

	var profitFactor decimal.Decimal
	if totalLoss.GreaterThan(decimal.Zero) {
		profitFactor = totalWin.Div(totalLoss)
	}

	// 计算Calmar比率
	var calmarRatio decimal.Decimal
	if e.maxDrawdown.GreaterThan(decimal.Zero) {
		calmarRatio = annualizedReturn.Div(e.maxDrawdown.Mul(decimal.NewFromInt(100)))
	}

	return &EventDrivenBacktestResult{
		InitialCapital:   e.initialCapital,
		FinalEquity:      finalEquity,
		TotalReturn:      totalReturn,
		AnnualizedReturn: annualizedReturn,
		MaxDrawdown:      e.maxDrawdown.Mul(decimal.NewFromInt(100)),
		SharpeRatio:      sharpeRatio,
		SortinoRatio:     sortinoRatio,
		CalmarRatio:      calmarRatio,
		WinRate:          winRate,
		ProfitFactor:     profitFactor,
		TotalTrades:      totalTrades,
		WinningTrades:    winningTrades,
		LosingTrades:     losingTrades,
		Transactions:     e.convertFillsToTransactions(),
		EquityCurve:      e.equityCurve,
		StartDate:        e.startDate,
		EndDate:          e.endDate,
		DurationDays:     durationDays,
	}
}

// convertFillsToTransactions 将成交记录转换为交易记录
func (e *EventDrivenEngine) convertFillsToTransactions() []*Transaction {
	transactions := make([]*Transaction, 0, len(e.fills))
	for _, fill := range e.fills {
		transType := "BUY"
		if fill.Side == OrderSideSell {
			transType = "SELL"
		}

		transaction := &Transaction{
			ID:         fill.OrderID,
			Symbol:     fill.Symbol,
			Type:       transType,
			Quantity:   fill.Quantity,
			Price:      fill.Price,
			Slippage:   fill.Slippage,
			Commission: fill.Commission,
			NetAmount:  fill.Price.Mul(fill.Quantity).Sub(fill.Commission),
			Timestamp:  fill.Timestamp,
			Reason:     fill.StrategyName,
		}
		transactions = append(transactions, transaction)
	}
	return transactions
}

// OrderEventHandler 订单事件处理器
type OrderEventHandler struct {
	engine *EventDrivenEngine
}

func (h *OrderEventHandler) Handle(event Event) error {
	// 订单事件处理逻辑
	return nil
}

// FillEventHandler 成交事件处理器
type FillEventHandler struct {
	engine *EventDrivenEngine
}

func (h *FillEventHandler) Handle(event Event) error {
	// 成交事件处理逻辑
	return nil
}

// calculateSharpeRatioFloat 计算夏普比率 (float64版本)
func calculateSharpeRatioFloat(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// 计算平均收益率
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// 计算标准差
	var varianceSum float64
	for _, r := range returns {
		diff := r - mean
		varianceSum += diff * diff
	}
	stdDev := sqrtFloat(varianceSum / float64(len(returns)))

	// 年化夏普比率 (假设月度数据)
	if stdDev > 0 {
		annualizedMean := mean * 12
		annualizedStdDev := stdDev * 3.464 // sqrt(12)
		return annualizedMean / annualizedStdDev
	}

	return 0
}

// calculateSortinoRatioFloat 计算索提诺比率 (float64版本)
func calculateSortinoRatioFloat(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// 计算平均收益率
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// 计算下行标准差（只考虑负收益）
	var downsideSum float64
	downsideCount := 0
	for _, r := range returns {
		if r < 0 {
			downsideSum += r * r
			downsideCount++
		}
	}

	if downsideCount == 0 {
		return 0
	}

	downsideStdDev := sqrtFloat(downsideSum / float64(downsideCount))

	// 年化索提诺比率
	if downsideStdDev > 0 {
		annualizedMean := mean * 12
		annualizedDownsideStdDev := downsideStdDev * 3.464
		return annualizedMean / annualizedDownsideStdDev
	}

	return 0
}

// sqrtFloat 计算平方根
func sqrtFloat(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
