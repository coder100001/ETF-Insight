package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etf-insight/config"
	"etf-insight/handlers"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/tasks"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

// @title ETF-Insight API
// @version 1.0
// @description ETF 数据管理与分析平台 API
// @host localhost:8080
// @BasePath /

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "path to config file")
	runOnce := flag.Bool("run-once", false, "run update once and exit")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		utils.Fatal("Failed to load config", err)
	}

	// 初始化日志
	utils.InitLogger(cfg.Log.Level)
	utils.Info("Configuration loaded", "path", *configPath)

	// 初始化数据库
	if err := models.InitDB(&cfg.Database); err != nil {
		utils.Fatal("Failed to initialize database", err)
	}
	utils.Info("Database initialized")

	// 初始化缓存服务
	cacheService := services.NewCacheService(&cfg.Cache)
	utils.Info("Cache service initialized", "type", "memory")

	// 初始化分析服务
	analysisService := services.NewETFAnalysisService(cacheService, nil)

	// 初始化调度器
	scheduler := tasks.NewScheduler(&cfg.Schedule, cacheService, analysisService, nil)

	// 如果指定了run-once，执行一次更新并退出
	if *runOnce {
		scheduler.RunOnce()
		return
	}

	// 启动调度器
	scheduler.Start()
	defer scheduler.Stop()

	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 使用中间件
	router.Use(gin.Recovery())
	router.Use(handlers.LoggerMiddleware())
	router.Use(handlers.CORSMiddleware())

	// 创建处理器
	etfHandler := handlers.NewETFHandler(cacheService, analysisService)
	portfolioHandler := handlers.NewPortfolioHandler(analysisService)

	// 健康检查
	router.GET("/health", handlers.HealthHandler)

	// API 路由
	api := router.Group("/api")
	{
		// ETF 相关路由
		etf := api.Group("/etf")
		{
			etf.GET("/list", etfHandler.GetETFList)
			etf.GET("/comparison", etfHandler.GetETFComparison)
			etf.GET("/:symbol/realtime", etfHandler.GetETFRealtime)
			etf.GET("/:symbol/history", etfHandler.GetETFHistory)
			etf.GET("/:symbol/metrics", etfHandler.GetETFMetrics)
			etf.GET("/:symbol/forecast", etfHandler.GetETFForecast)
			etf.POST("/update-realtime", etfHandler.UpdateRealtimeData)

			// 投资组合分析
			etf.POST("/portfolio", portfolioHandler.AnalyzePortfolio)
		}

		// 投资组合配置路由
		portfolioConfigs := api.Group("/portfolio-configs")
		{
			portfolioConfigs.GET("/", portfolioHandler.GetPortfolioConfigs)
			portfolioConfigs.GET("/:id", portfolioHandler.GetPortfolioConfig)
			portfolioConfigs.POST("/", portfolioHandler.CreatePortfolioConfig)
			portfolioConfigs.PUT("/:id", portfolioHandler.UpdatePortfolioConfig)
			portfolioConfigs.DELETE("/:id", portfolioHandler.DeletePortfolioConfig)
			portfolioConfigs.POST("/:id/toggle-status", portfolioHandler.TogglePortfolioConfigStatus)
			portfolioConfigs.POST("/:id/analyze", portfolioHandler.AnalyzePortfolioConfig)
		}

		// 汇率路由
		api.GET("/exchange-rates", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": map[string]float64{
					"USD_CNY": 7.2,
					"USD_HKD": 7.8,
					"CNY_USD": 0.139,
					"HKD_USD": 0.128,
				},
			})
		})
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	utils.Info("Starting server", "addr", addr)

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("Failed to start server", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		utils.Error("Server shutdown error", err)
	}
}
