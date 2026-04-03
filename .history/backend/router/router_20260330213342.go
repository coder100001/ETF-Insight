package router

import (
	"etf-insight/handlers"
	"etf-insight/infrastructure/cache"
	"etf-insight/infrastructure/database"
	"etf-insight/infrastructure/messagequeue"
	"etf-insight/infrastructure/metrics"
	"etf-insight/services"

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
		r.setupETFRoutes(v1)
		r.setupPortfolioRoutes(v1)
		r.setupAsyncRoutes(v1)
		r.setupConfigRoutes(v1)
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

// setupETFRoutes ETF相关路由
func (r *Router) setupETFRoutes(group *gin.RouterGroup) {
	// 初始化服务
	cacheService := services.NewCacheService(nil, r.redis)
	analysisService := services.NewETFAnalysisService(cacheService, nil)

	// 初始化handler
	handler := handlers.NewETFHandler(analysisService)

	etfs := group.Group("/etfs")
	{
		etfs.GET("", handler.GetETFList)
		etfs.GET("/:symbol", handler.GetETFDetail)
		etfs.GET("/:symbol/prices", handler.GetETFPrices)
		etfs.GET("/:symbol/holdings", handler.GetETFHoldings)
		etfs.GET("/:symbol/realtime", handler.GetETFRealtime)
		etfs.POST("/:symbol/refresh", handler.RefreshETFData)
	}

	// 比较分析
	group.GET("/comparison", handler.GetETFComparison)
}

// setupPortfolioRoutes 投资组合路由
func (r *Router) setupPortfolioRoutes(group *gin.RouterGroup) {
	cacheService := services.NewCacheService(nil, r.redis)
	exchangeService := services.NewExchangeRateService()

	// 使用新的Repository模式
	portfolioHandler := handlers.NewPortfolioHandler(cacheService, exchangeService)

	portfolios := group.Group("/portfolios")
	{
		portfolios.POST("/analyze", portfolioHandler.AnalyzePortfolio)
		portfolios.POST("/optimize", portfolioHandler.OptimizePortfolio)
		portfolios.POST("/backtest", portfolioHandler.BacktestPortfolio)
		portfolios.GET("/overlap", portfolioHandler.GetPortfolioOverlap)
	}
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

// setupConfigRoutes 配置路由
func (r *Router) setupConfigRoutes(group *gin.RouterGroup) {
	configHandler := handlers.NewETFConfigHandler(r.db)

	configs := group.Group("/configs")
	{
		configs.GET("", configHandler.GetAllConfigs)
		configs.GET("/:symbol", configHandler.GetConfig)
		configs.POST("", configHandler.CreateConfig)
		configs.PUT("/:symbol", configHandler.UpdateConfig)
		configs.DELETE("/:symbol", configHandler.DeleteConfig)
	}
}

// GetEngine 获取gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run 启动服务器
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
