package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etf-insight/config"
	"etf-insight/handlers"
	"etf-insight/middleware"
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

	if err := models.InitDB(cfg.Database.GetDSN()); err != nil {
		utils.Fatal("Failed to initialize database", err)
	}
	utils.Info("Database initialized")

	if err := models.InitDefaultData(); err != nil {
		utils.Fatal("Failed to initialize default data", err)
	}
	utils.Info("Default data initialized")

	// 初始化汇率相关表
	if err := models.InitExchangeRateTables(); err != nil {
		utils.Fatal("Failed to initialize exchange rate tables", err)
	}
	utils.Info("Exchange rate tables initialized")

	if err := models.InitDefaultCurrencyPairs(); err != nil {
		utils.Fatal("Failed to initialize default currency pairs", err)
	}
	utils.Info("Default currency pairs initialized")

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
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimiter())

	etfHandler := handlers.NewETFHandler(cacheService, analysisService)
	portfolioHandler := handlers.NewPortfolioHandler(analysisService)

	router.GET("/health", handlers.HealthHandler)

	router.GET("/api/etf/list", etfHandler.GetETFList)
	router.GET("/api/etf/comparison", etfHandler.GetETFComparison)
	router.POST("/api/etf/update-realtime", etfHandler.UpdateRealtimeData)
	router.POST("/api/etf/portfolio", portfolioHandler.AnalyzePortfolio)
	router.GET("/api/etf/:symbol/realtime", etfHandler.GetETFRealtime)
	router.GET("/api/etf/:symbol/history", etfHandler.GetETFHistory)
	router.GET("/api/etf/:symbol/metrics", etfHandler.GetETFMetrics)
	router.GET("/api/etf/:symbol/forecast", etfHandler.GetETFForecast)

	router.GET("/api/portfolio-configs/", portfolioHandler.GetPortfolioConfigs)
	router.POST("/api/portfolio-configs/", portfolioHandler.CreatePortfolioConfig)
	router.GET("/api/portfolio-configs/:id", portfolioHandler.GetPortfolioConfig)
	router.PUT("/api/portfolio-configs/:id", portfolioHandler.UpdatePortfolioConfig)
	router.DELETE("/api/portfolio-configs/:id", portfolioHandler.DeletePortfolioConfig)
	router.POST("/api/portfolio-configs/:id/toggle-status", portfolioHandler.TogglePortfolioConfigStatus)
	router.POST("/api/portfolio-configs/:id/analyze", portfolioHandler.AnalyzePortfolioConfig)

	// ETF配置路由
	etfConfigHandler := handlers.NewETFConfigHandler()
	router.GET("/api/etf-configs/", etfConfigHandler.GetETFConfigs)
	router.POST("/api/etf-configs/", etfConfigHandler.CreateETFConfig)
	router.GET("/api/etf-configs/:id", etfConfigHandler.GetETFConfig)
	router.PUT("/api/etf-configs/:id", etfConfigHandler.UpdateETFConfig)
	router.DELETE("/api/etf-configs/:id", etfConfigHandler.DeleteETFConfig)
	router.POST("/api/etf-configs/:id/toggle-status", etfConfigHandler.ToggleETFConfigStatus)
	router.POST("/api/etf-configs/:id/auto-update", etfConfigHandler.ToggleETFConfigAutoUpdate)

	// A股红利ETF组合路由
	aShareHandler := handlers.NewASharePortfolioHandler()
	router.GET("/api/a-share/etfs", aShareHandler.GetDefaultETFs)
	router.GET("/api/a-share/portfolio/default", aShareHandler.GetDefaultPortfolio)
	router.POST("/api/a-share/portfolio/analyze", aShareHandler.AnalyzePortfolio)
	router.POST("/api/a-share/portfolio/holding/:symbol", aShareHandler.UpdateHolding)
	router.GET("/api/a-share/dividend/:frequency", aShareHandler.CalculateDividendByFrequency)

	router.GET("/api/exchange-rates", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": []map[string]interface{}{
				{
					"id":             1,
					"from_currency":  "USD",
					"to_currency":    "CNY",
					"rate":           7.2,
					"previous_rate":  7.18,
					"change_percent": 0.28,
					"updated_at":     time.Now().Format(time.RFC3339),
					"source":         "央行中间价",
				},
				{
					"id":             2,
					"from_currency":  "USD",
					"to_currency":    "HKD",
					"rate":           7.8,
					"previous_rate":  7.79,
					"change_percent": 0.13,
					"updated_at":     time.Now().Format(time.RFC3339),
					"source":         "央行中间价",
				},
				{
					"id":             3,
					"from_currency":  "CNY",
					"to_currency":    "USD",
					"rate":           0.139,
					"previous_rate":  0.1392,
					"change_percent": -0.14,
					"updated_at":     time.Now().Format(time.RFC3339),
					"source":         "计算汇率",
				},
				{
					"id":             4,
					"from_currency":  "HKD",
					"to_currency":    "USD",
					"rate":           0.128,
					"previous_rate":  0.1284,
					"change_percent": -0.31,
					"updated_at":     time.Now().Format(time.RFC3339),
					"source":         "计算汇率",
				},
			},
		})
	})

	router.Static("/assets", "../frontend/dist/assets")
	router.StaticFile("/favicon.svg", "../frontend/dist/favicon.svg")
	router.StaticFile("/icons.svg", "../frontend/dist/icons.svg")

	router.NoRoute(func(c *gin.Context) {
		c.File("../frontend/dist/index.html")
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}

	utils.Info("Starting server", "addr", addr)

	go func() {
		if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
			utils.Info("HTTPS enabled", "cert", cfg.Server.CertFile)
			if err := srv.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil && err != http.ErrServerClosed {
				utils.Fatal("Failed to start HTTPS server", err)
			}
		} else {
			utils.Warn("Running in HTTP mode (no TLS certificates provided)")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				utils.Fatal("Failed to start HTTP server", err)
			}
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
