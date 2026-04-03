package router

import (
	"etf-insight/handlers"
	"etf-insight/infrastructure/cache"
	"etf-insight/infrastructure/database"
	"etf-insight/infrastructure/messagequeue"
	"etf-insight/infrastructure/metrics"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Router 路由管理器
type Router struct {
	engine   *gin.Engine
	db       *gorm.DB
	redis    *cache.RedisClient
	rabbitmq *messagequeue.RabbitMQ
}

// New 创建路由管理器
func New(db *database.DB, redis *cache.RedisClient, rabbitmq *messagequeue.RabbitMQ) *Router {
	engine := gin.New()

	// 使用中间件
	engine.Use(gin.Recovery())
	engine.Use(metrics.MetricsMiddleware())

	return &Router{
		engine:   engine,
		db:       db.DB,
		redis:    redis,
		rabbitmq: rabbitmq,
	}
}

// SetupRoutes 设置所有路由
func (r *Router) SetupRoutes() {
	// 健康检查
	r.setupHealthRoutes()

	// 监控指标
	r.setupMetricsRoutes()

	// API v1
	v1 := r.engine.Group("/api/v1")
	{
		r.setupAsyncRoutes(v1)
		r.setupAShareRoutes(v1)
	}
}

// setupHealthRoutes 健康检查路由
func (r *Router) setupHealthRoutes() {
	handler := handlers.NewHealthHandler(r.db, r.redis, r.rabbitmq)

	r.engine.GET("/health", handler.HealthCheck)
	r.engine.GET("/ready", handler.ReadinessCheck)
	r.engine.GET("/live", handler.LivenessCheck)
}

// setupMetricsRoutes 监控指标路由
func (r *Router) setupMetricsRoutes() {
	r.engine.GET("/metrics", metrics.PrometheusHandler())
}

// setupAsyncRoutes 异步分析路由
func (r *Router) setupAsyncRoutes(group *gin.RouterGroup) {
	asyncHandler := handlers.NewAsyncAnalyticsHandler(r.rabbitmq, r.redis, r.db)

	async := group.Group("/async")
	{
		// 提交任务
		async.POST("/risk-analysis", asyncHandler.SubmitRiskAnalysis)
		async.POST("/factor-analysis", asyncHandler.SubmitFactorAnalysis)
		async.POST("/overlap-analysis", asyncHandler.SubmitOverlapAnalysis)
		async.POST("/backtest", asyncHandler.SubmitBacktest)

		// 任务管理
		async.GET("/tasks/:taskId/status", asyncHandler.GetTaskStatus)
		async.GET("/tasks/:taskId/result", asyncHandler.GetTaskResult)
		async.GET("/tasks/pending", asyncHandler.ListPendingTasks)
		async.DELETE("/tasks/:taskId", asyncHandler.CancelTask)
	}
}

// setupAShareRoutes A 股 ETF 路由
func (r *Router) setupAShareRoutes(group *gin.RouterGroup) {
	aShareHandler := handlers.NewASharePortfolioHandler()

	aShare := group.Group("/a-share")
	{
		aShare.GET("/etfs", aShareHandler.GetDefaultETFs)
		aShare.GET("/portfolio", aShareHandler.GetDefaultPortfolio)
		aShare.POST("/portfolio/analyze", aShareHandler.AnalyzePortfolio)
	}
}

// GetEngine 获取 gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run 启动服务器
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
