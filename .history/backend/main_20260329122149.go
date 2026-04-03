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

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	runOnce := flag.Bool("run-once", false, "run update once and exit")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		utils.Fatal("Failed to load config", err)
	}

	utils.InitLogger(cfg.Log.Level)
	utils.Info("Configuration loaded", "path", *configPath)

	if err := models.InitDB(); err != nil {
		utils.Fatal("Failed to initialize database", err)
	}
	utils.Info("Database initialized")

	cacheService := services.NewCacheService(&cfg.ETF.Cache)
	utils.Info("Cache service initialized", "type", "memory")

	analysisService := services.NewETFAnalysisService(cacheService, nil)

	exchangeService := services.NewExchangeRateService()
	scheduler := tasks.NewScheduler(&cfg.Schedule, cacheService, analysisService, exchangeService)

	if *runOnce {
		scheduler.RunOnce()
		return
	}

	scheduler.Start()
	defer scheduler.Stop()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(handlers.LoggerMiddleware())
	router.Use(handlers.CORSMiddleware())

	etfHandler := handlers.NewETFHandler(cacheService, analysisService)
	portfolioHandler := handlers.NewPortfolioHandler(analysisService)

	router.GET("/health", handlers.HealthHandler)

	api := router.Group("/api")
	{
		etf := api.Group("/etf")
		{
			etf.GET("/list", etfHandler.GetETFList)
			etf.GET("/comparison", etfHandler.GetETFComparison)
			etf.GET("/:symbol/realtime", etfHandler.GetETFRealtime)
			etf.GET("/:symbol/history", etfHandler.GetETFHistory)
			etf.GET("/:symbol/metrics", etfHandler.GetETFMetrics)
			etf.GET("/:symbol/forecast", etfHandler.GetETFForecast)
			etf.POST("/update-realtime", etfHandler.UpdateRealtimeData)
			etf.POST("/portfolio", portfolioHandler.AnalyzePortfolio)
		}

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

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	utils.Info("Starting server", "addr", addr)

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
