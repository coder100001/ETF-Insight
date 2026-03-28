package routers

import (
	"etf-insight/config"
	"etf-insight/handlers"
	"etf-insight/services"
	"etf-insight/tasks"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config, cache *services.CacheService, analysis *services.ETFAnalysisService, exchange *services.ExchangeRateService, scheduler *tasks.Scheduler) *gin.Engine {
	// 设置Gin模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	r.Use(cors.New(corsConfig))

	// 创建handler
	etfHandler := handlers.NewETFHandler(cache, analysis)
	portfolioHandler := handlers.NewPortfolioHandler(analysis)
	exchangeHandler := handlers.NewExchangeRateHandler(exchange)
	workflowHandler := handlers.NewWorkflowHandler()
	schedulerHandler := handlers.NewSchedulerHandler(scheduler)
	adminHandler := handlers.NewAdminHandler()

	// API路由组
	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"message": "API is running",
			})
		})

		// ETF相关路由
		etf := api.Group("/etf")
		{
			etf.GET("/list", etfHandler.GetETFList)
			etf.GET("/comparison", etfHandler.GetComparison)
			etf.GET("/portfolio", etfHandler.GetPortfolioAnalysis)
			etf.GET("/:symbol/realtime", etfHandler.GetRealtimeData)
			etf.GET("/:symbol/metrics", etfHandler.GetMetrics)
			etf.GET("/:symbol/history", etfHandler.GetHistory)
			etf.GET("/:symbol/forecast", etfHandler.GetForecast)
			etf.POST("/update-realtime", etfHandler.UpdateRealtimeData)
		}

		// 投资组合配置路由
		portfolio := api.Group("/portfolio-configs")
		{
			portfolio.GET("/", portfolioHandler.GetConfigs)
			portfolio.POST("/", portfolioHandler.CreateConfig)
			portfolio.GET("/:id", portfolioHandler.GetConfigDetail)
			portfolio.PUT("/:id", portfolioHandler.UpdateConfig)
			portfolio.DELETE("/:id", portfolioHandler.DeleteConfig)
			portfolio.POST("/:id/toggle-status", portfolioHandler.ToggleStatus)
			portfolio.POST("/:id/analyze", portfolioHandler.AnalyzeConfig)
		}

		// 汇率相关路由
		exchange := api.Group("/exchange-rates")
		{
			exchange.GET("/", exchangeHandler.GetRates)
			exchange.GET("/history", exchangeHandler.GetHistory)
			exchange.GET("/convert", exchangeHandler.Convert)
			exchange.POST("/update", exchangeHandler.UpdateRates)
		}

		// 工作流相关路由
		workflow := api.Group("/workflows")
		{
			workflow.GET("/", workflowHandler.GetWorkflows)
			workflow.POST("/", workflowHandler.CreateWorkflow)
			workflow.GET("/:id", workflowHandler.GetWorkflow)
			workflow.PUT("/:id", workflowHandler.UpdateWorkflow)
			workflow.DELETE("/:id", workflowHandler.DeleteWorkflow)
			workflow.POST("/:id/start", workflowHandler.StartWorkflow)
		}

		// 工作流实例路由
		instances := api.Group("/instances")
		{
			instances.GET("/", workflowHandler.GetInstances)
			instances.GET("/:id", workflowHandler.GetInstance)
			instances.POST("/:id/retry", workflowHandler.RetryInstance)
		}

		// 定时任务路由
		scheduler := api.Group("/scheduler")
		{
			scheduler.GET("/jobs", schedulerHandler.GetJobs)
			scheduler.POST("/run-once", schedulerHandler.RunOnce)
		}

		// 管理路由
		admin := api.Group("/admin")
		{
			admin.GET("/stats", adminHandler.GetStats)
			admin.GET("/logs", adminHandler.GetLogs)
			admin.POST("/clear-cache", adminHandler.ClearCache)
		}
	}

	return r
}
