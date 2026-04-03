package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"etf-insight/analytics"
	"etf-insight/infrastructure/cache"
	"etf-insight/infrastructure/database"
	"etf-insight/infrastructure/messagequeue"

	"github.com/shopspring/decimal"
)

// TaskMessage 任务消息
type TaskMessage struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// ResultMessage 结果消息
type ResultMessage struct {
	TaskID    string      `json:"task_id"`
	Type      string      `json:"type"`
	Status    string      `json:"status"` // success, error
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	Duration  int64       `json:"duration_ms"`
	Timestamp time.Time   `json:"timestamp"`
}

// AnalyticsWorker 分析工作器
type AnalyticsWorker struct {
	id       int
	db       *database.DB
	redis    *cache.RedisClient
	rabbitmq *messagequeue.RabbitMQ
	quit     chan bool
}

// NewAnalyticsWorker 创建工作器
func NewAnalyticsWorker(id int, db *database.DB, redis *cache.RedisClient, rabbitmq *messagequeue.RabbitMQ) *AnalyticsWorker {
	return &AnalyticsWorker{
		id:       id,
		db:       db,
		redis:    redis,
		rabbitmq: rabbitmq,
		quit:     make(chan bool),
	}
}

// Start 启动工作器
func (w *AnalyticsWorker) Start() {
	log.Printf("Worker %d started", w.id)

	// 消费任务队列
	msgs, err := w.rabbitmq.Consume(messagequeue.QueueAnalyticsTasks, fmt.Sprintf("worker-%d", w.id))
	if err != nil {
		log.Fatalf("Failed to consume queue: %v", err)
	}

	for {
		select {
		case msg := <-msgs:
			w.processMessage(msg)
		case <-w.quit:
			log.Printf("Worker %d stopped", w.id)
			return
		}
	}
}

// Stop 停止工作器
func (w *AnalyticsWorker) Stop() {
	close(w.quit)
}

// Delivery 消息投递接口
type Delivery interface {
	Body() []byte
	Ack(multiple bool) error
	Nack(multiple bool, requeue bool) error
}

// processMessage 处理消息
func (w *AnalyticsWorker) processMessage(msg interface{}) {
	// 类型断言
	delivery, ok := msg.(Delivery)
	if !ok {
		log.Printf("Invalid message type")
		return
	}

	start := time.Now()

	var task TaskMessage
	if err := json.Unmarshal(delivery.Body(), &task); err != nil {
		log.Printf("Failed to unmarshal task: %v", err)
		delivery.Nack(false, false)
		return
	}

	log.Printf("Worker %d processing task %s (type: %s)", w.id, task.ID, task.Type)

	// 处理任务
	var result interface{}
	var taskErr error

	ctx := context.Background()

	switch task.Type {
	case "risk_analysis":
		result, taskErr = w.handleRiskAnalysis(ctx, task.Payload)
	case "factor_analysis":
		result, taskErr = w.handleFactorAnalysis(ctx, task.Payload)
	case "overlap_analysis":
		result, taskErr = w.handleOverlapAnalysis(ctx, task.Payload)
	case "backtest":
		result, taskErr = w.handleBacktest(ctx, task.Payload)
	default:
		taskErr = fmt.Errorf("unknown task type: %s", task.Type)
	}

	duration := time.Since(start).Milliseconds()

	// 发送结果
	resultMsg := ResultMessage{
		TaskID:    task.ID,
		Type:      task.Type,
		Status:    "success",
		Result:    result,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if taskErr != nil {
		resultMsg.Status = "error"
		resultMsg.Error = taskErr.Error()
	}

	if err := w.rabbitmq.PublishJSON(ctx, messagequeue.ExchangeAnalytics, "analytics.result", resultMsg); err != nil {
		log.Printf("Failed to publish result: %v", err)
	}

	// 缓存结果
	if taskErr == nil {
		cacheKey := fmt.Sprintf("analytics:result:%s", task.ID)
		w.redis.SetJSON(ctx, cacheKey, resultMsg, 24*time.Hour)
	}

	// 确认消息
	delivery.Ack(false)

	log.Printf("Worker %d completed task %s in %dms", w.id, task.ID, duration)
}

// handleRiskAnalysis 处理风险分析
func (w *AnalyticsWorker) handleRiskAnalysis(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	var req struct {
		Symbol    string  `json:"symbol"`
		Period    string  `json:"period"`
		RiskFree  float64 `json:"risk_free_rate"`
		Benchmark string  `json:"benchmark_symbol"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	// 从数据库获取价格数据
	prices, err := w.getPricesFromDB(ctx, req.Symbol, req.Period)
	if err != nil {
		return nil, err
	}

	// 计算风险指标
	riskFreeRate := decimal.NewFromFloat(req.RiskFree)
	metrics, err := analytics.CalculateRiskMetrics(prices, riskFreeRate, nil)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// handleFactorAnalysis 处理因子分析
func (w *AnalyticsWorker) handleFactorAnalysis(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	var req struct {
		Symbol string `json:"symbol"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	// 获取持仓数据
	holdings, err := w.getHoldingsFromDB(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	// 计算因子暴露
	result := analytics.CalculateFactorExposure(req.Symbol, holdings.Stocks, nil)

	return result, nil
}

// handleOverlapAnalysis 处理重叠分析
func (w *AnalyticsWorker) handleOverlapAnalysis(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	var req struct {
		ETF1 string `json:"etf1"`
		ETF2 string `json:"etf2"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	// 获取持仓数据
	holdings1, err := w.getHoldingsFromDB(ctx, req.ETF1)
	if err != nil {
		return nil, err
	}

	holdings2, err := w.getHoldingsFromDB(ctx, req.ETF2)
	if err != nil {
		return nil, err
	}

	// 计算重叠度
	overlap := analytics.CalculateOverlap(holdings1, holdings2)

	return overlap, nil
}

// handleBacktest 处理回测
func (w *AnalyticsWorker) handleBacktest(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	var req struct {
		Portfolio []struct {
			Symbol string  `json:"symbol"`
			Weight float64 `json:"weight"`
		} `json:"portfolio"`
		Config struct {
			StartDate      string  `json:"start_date"`
			EndDate        string  `json:"end_date"`
			InitialCapital float64 `json:"initial_capital"`
			RebalanceFreq  string  `json:"rebalance_freq"`
			CommissionRate float64 `json:"commission_rate"`
		} `json:"config"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	// 获取价格数据
	prices := make(map[string][]analytics.PricePoint)
	for _, p := range req.Portfolio {
		priceData, err := w.getPricesFromDB(ctx, p.Symbol, "5y")
		if err != nil {
			return nil, err
		}
		prices[p.Symbol] = priceData
	}

	// 构建回测配置
	config := analytics.BacktestConfig{
		StartDate:      parseDate(req.Config.StartDate),
		EndDate:        parseDate(req.Config.EndDate),
		InitialCapital: decimal.NewFromFloat(req.Config.InitialCapital),
		RebalanceFreq:  req.Config.RebalanceFreq,
		CommissionRate: decimal.NewFromFloat(req.Config.CommissionRate),
	}

	// 构建投资组合
	portfolio := make([]analytics.PortfolioConfig, len(req.Portfolio))
	for i, p := range req.Portfolio {
		portfolio[i] = analytics.PortfolioConfig{
			Symbol: p.Symbol,
			Weight: decimal.NewFromFloat(p.Weight),
		}
	}

	// 执行回测
	engine := analytics.NewBacktestEngine(config, portfolio, prices)
	result := engine.Run()

	return result, nil
}

// getPricesFromDB 从数据库获取价格数据
func (w *AnalyticsWorker) getPricesFromDB(ctx context.Context, symbol, period string) ([]analytics.PricePoint, error) {
	// 这里应该查询实际的数据库
	// 简化示例
	return []analytics.PricePoint{}, nil
}

// getHoldingsFromDB 从数据库获取持仓数据
func (w *AnalyticsWorker) getHoldingsFromDB(ctx context.Context, symbol string) (analytics.ETFHoldings, error) {
	// 这里应该查询实际的数据库
	// 简化示例
	return analytics.ETFHoldings{Symbol: symbol}, nil
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func main() {
	// 获取工作器数量
	workerCount := 4
	if wc := os.Getenv("WORKER_COUNT"); wc != "" {
		if n, err := strconv.Atoi(wc); err == nil {
			workerCount = n
		}
	}

	log.Printf("Starting Analytics Service with %d workers", workerCount)

	// 初始化数据库连接
	dbConfig := database.DefaultConfig()
	db, err := database.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 初始化Redis
	redisConfig := cache.DefaultConfig()
	redis, err := cache.New(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redis.Close()

	// 初始化RabbitMQ
	rabbitmqConfig := messagequeue.DefaultConfig()
	rabbitmq, err := messagequeue.New(rabbitmqConfig)
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq: %v", err)
	}
	defer rabbitmq.Close()

	// 设置RabbitMQ拓扑
	if err := rabbitmq.SetupTopology(); err != nil {
		log.Fatalf("Failed to setup rabbitmq topology: %v", err)
	}

	// 创建工作器
	workers := make([]*AnalyticsWorker, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = NewAnalyticsWorker(i, db, redis, rabbitmq)
		go workers[i].Start()
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	log.Println("Shutting down Analytics Service...")

	// 停止所有工作器
	for _, worker := range workers {
		worker.Stop()
	}

	// 等待一段时间让工作器完成当前任务
	time.Sleep(2 * time.Second)

	log.Println("Analytics Service stopped")
}
