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
	"etf-insight/docs"
	"etf-insight/handlers"
	"etf-insight/middleware"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/datasource"
	"etf-insight/services/event"
	erdatasource "etf-insight/services/exchange_rate/datasource"
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

	// 初始化事件总线
	event.InitEventBus(models.DB)
	utils.Info("Event bus initialized")

	// 初始化汇率相关表
	if err := models.InitExchangeRateTables(); err != nil {
		utils.Fatal("Failed to initialize exchange rate tables", err)
	}
	utils.Info("Exchange rate tables initialized")

	if err := models.InitDefaultCurrencyPairs(); err != nil {
		utils.Fatal("Failed to initialize default currency pairs", err)
	}
	utils.Info("Default currency pairs initialized")

	// 缓存服务已移除，不再需要初始化
	utils.Info("Cache service removed, all data will be fetched directly from database/API")

	analysisService := services.NewETFAnalysisService(nil)
	optimizer := services.NewPortfolioOptimizer(analysisService)
	exchangeService := services.NewExchangeRateService()

	// 初始化 Finage 数据源（主要数据源）
	finageProvider := datasource.NewFinageProvider()
	utils.Info("Finage provider initialized",
		"available", finageProvider.IsAvailable(context.Background()))

	// 创建数据源工厂
	providerFactory := datasource.NewProviderFactory()
	providerFactory.Register("finage", finageProvider)
	providerFactory.Register("fallback", datasource.NewMockDataProvider())

	// 获取可用的数据源
	ctx := context.Background()
	defaultProvider, err := providerFactory.GetDefault(ctx)
	if err != nil {
		utils.Warn("No data source available, using mock provider", "error", err)
		defaultProvider = datasource.NewMockDataProvider()
	} else {
		utils.Info("Using data source", "provider", defaultProvider.GetName())
	}

	scheduler := tasks.NewScheduler(&cfg.Schedule, analysisService, exchangeService, defaultProvider)

	if *runOnce {
		scheduler.RunOnce()
		return
	}

	scheduler.Start()
	defer scheduler.Stop()

	// 启动汇率同步定时任务（使用新的多数据源故障转移系统）
	exchangeRateConfig := &erdatasource.DataSourceConfig{
		OpenExchangeAPIKey: cfg.ExchangeRate.OpenExchangeAPIKey,
		CurrencyAPIKey:     cfg.ExchangeRate.CurrencyAPIKey,
	}
	exchangeRateTask := tasks.NewExchangeRateTask(exchangeRateConfig)
	exchangeRateTask.Start()
	defer exchangeRateTask.Stop()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(handlers.LoggerMiddleware())
	router.Use(handlers.CORSMiddleware())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimiter())

	etfHandler := handlers.NewETFHandler(analysisService, defaultProvider)
	portfolioHandler := handlers.NewPortfolioHandler(analysisService)
	optimizerHandler := handlers.NewPortfolioOptimizerHandler(optimizer)
	optimizationHandler := handlers.NewOptimizationHandler()

	// 初始化操作日志Handler
	operationLogsService := services.NewOperationLogsService(models.DB)
	operationLogsHandler := handlers.NewOperationLogsHandler(operationLogsService)

	router.GET("/health", handlers.HealthHandler)
	router.GET("/ready", handlers.ReadyHandler)
	router.GET("/live", handlers.LiveHandler)

	router.GET("/api/etf/list", etfHandler.GetETFList)
	router.GET("/api/etf/comparison", etfHandler.GetETFComparison)
	router.POST("/api/etf/update-realtime", etfHandler.UpdateRealtimeData)
	router.POST("/api/etf/portfolio", portfolioHandler.AnalyzePortfolio)
	router.GET("/api/etf/:symbol/realtime", etfHandler.GetETFRealtime)
	router.GET("/api/etf/:symbol/history", etfHandler.GetETFHistory)
	router.GET("/api/etf/:symbol/metrics", etfHandler.GetETFMetrics)
	router.GET("/api/etf/:symbol/forecast", etfHandler.GetETFForecast)
	router.GET("/api/etf/:symbol/risk", etfHandler.GetETFRisk)

	router.GET("/api/portfolio-configs/", portfolioHandler.GetPortfolioConfigs)
	router.POST("/api/portfolio-configs/", portfolioHandler.CreatePortfolioConfig)
	router.GET("/api/portfolio-configs/:id", portfolioHandler.GetPortfolioConfig)
	router.PUT("/api/portfolio-configs/:id", portfolioHandler.UpdatePortfolioConfig)
	router.DELETE("/api/portfolio-configs/:id", portfolioHandler.DeletePortfolioConfig)
	router.POST("/api/portfolio-configs/:id/toggle-status", portfolioHandler.TogglePortfolioConfigStatus)
	router.POST("/api/portfolio-configs/:id/analyze", portfolioHandler.AnalyzePortfolioConfig)

	// 投资组合情景分析路由
	router.POST("/api/portfolio/scenarios", portfolioHandler.AnalyzeScenarios)
	router.GET("/api/portfolio/default-templates", portfolioHandler.GetDefaultPortfolioTemplates)
	router.POST("/api/portfolio/risk", portfolioHandler.AnalyzePortfolioRisk)

	// 投资组合优化路由
	router.POST("/api/portfolio/optimize", optimizerHandler.OptimizePortfolio)
	router.POST("/api/portfolio/efficient-frontier", optimizerHandler.GetEfficientFrontier)

	// MPT均值-方差优化路由
	router.POST("/api/optimization/mpt", optimizationHandler.MPTOptimize)
	router.POST("/api/optimization/efficient-frontier", optimizationHandler.EfficientFrontier)
	router.POST("/api/optimization/covariance", optimizationHandler.CalculateCovarianceMatrix)
	router.POST("/api/optimization/etf-statistics", optimizationHandler.GetETFStatistics)

	// 风险平价优化路由
	router.POST("/api/optimization/risk-parity", optimizationHandler.RiskParityOptimize)

	// Black-Litterman优化路由
	router.POST("/api/optimization/black-litterman", optimizationHandler.BlackLittermanOptimize)
	router.POST("/api/optimization/market-implied-returns", optimizationHandler.MarketImpliedReturns)

	// Fama-French因子分析路由
	factorHandler := handlers.NewFactorHandler()
	router.POST("/api/factor/analyze", factorHandler.AnalyzeFactorExposure)
	router.POST("/api/factor/portfolio", factorHandler.AnalyzePortfolioFactors)
	router.POST("/api/factor/multi-asset", factorHandler.AnalyzeMultipleAssets)
	router.GET("/api/factor/statistics", factorHandler.GetFactorStatistics)
	router.POST("/api/factor/risk-decomposition", factorHandler.DecomposeRisk)
	router.POST("/api/factor/compare", factorHandler.CompareFactorAttribution)

	// ETF持仓穿透路由
	etfHoldingHandler := handlers.NewETFHoldingHandler(models.DB)
	router.GET("/api/etf/:symbol/holdings", etfHoldingHandler.GetETFHoldings)
	router.GET("/api/etf/overlap", etfHoldingHandler.GetETFOverlap)
	router.GET("/api/etf/:symbol/top-holdings", etfHoldingHandler.GetTopHoldings)
	router.GET("/api/etf/:symbol/sector-allocation", etfHoldingHandler.GetSectorAllocation)
	router.POST("/api/etf/holdings/comparison", etfHoldingHandler.GetETFHoldingsComparison)
	router.POST("/api/etf/:symbol/holdings", etfHoldingHandler.SaveETFHoldings)

	// 投资组合穿透分析路由
	portfolioPenetrationHandler := handlers.NewPortfolioPenetrationHandler(models.DB)
	router.POST("/api/portfolio/penetration", portfolioPenetrationHandler.AnalyzePortfolioPenetration)
	router.POST("/api/portfolio/compare", portfolioPenetrationHandler.ComparePortfolios)
	router.POST("/api/portfolio/sector-exposure", portfolioPenetrationHandler.GetSectorExposure)

	// 缓存管理路由
	router.GET("/api/cache/overlap/stats", etfHoldingHandler.GetCacheStats)
	router.POST("/api/cache/overlap/invalidate", etfHoldingHandler.InvalidateCache)
	router.POST("/api/cache/overlap/clean", etfHoldingHandler.CleanExpiredCache)

	// 回测路由
	backtestHandler := handlers.NewBacktestHandler()
	router.POST("/api/backtest/run", backtestHandler.RunBacktest)
	router.POST("/api/backtest/event-driven", backtestHandler.RunEventDrivenBacktest)
	router.GET("/api/backtest/strategies", backtestHandler.ListStrategies)
	router.POST("/api/backtest/factors", backtestHandler.AnalyzeFactors)

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

	// A股ETF价格路由
	router.GET("/api/a-share/prices", aShareHandler.GetETFPrices)
	router.GET("/api/a-share/prices/:symbol", aShareHandler.GetETFPriceBySymbol)
	router.POST("/api/a-share/prices/refresh", aShareHandler.RefreshETFPrices)

	// A股数据源路由
	aShareDataHandler := handlers.NewAShareDataHandler()
	router.POST("/api/a-share/enable-akshare", aShareDataHandler.EnableAKShare)
	router.POST("/api/a-share/sync-etf-list", aShareDataHandler.SyncETFList)
	router.POST("/api/a-share/sync-prices", aShareDataHandler.SyncETFPrices)
	router.POST("/api/a-share/refresh-all", aShareDataHandler.RefreshAllData)
	router.GET("/api/a-share/price/:symbol", aShareDataHandler.GetETFPrice)
	router.GET("/api/a-share/all-prices", aShareDataHandler.GetAllETFPrices)
	router.POST("/api/a-share/historical/:symbol", aShareDataHandler.GetHistoricalData)
	router.GET("/api/a-share/search", aShareDataHandler.SearchETFs)
	router.GET("/api/a-share/by-frequency/:frequency", aShareDataHandler.GetETFsByFrequency)
	router.GET("/api/a-share/dividend-yield/:symbol", aShareDataHandler.CalculateDividendYield)
	router.GET("/api/a-share/data-source-status", aShareDataHandler.GetDataSourceStatus)

	// 跨资产类别ETF路由
	universalETFHandler := handlers.NewUniversalETFHandler()
	router.POST("/api/universal-etf/initialize", universalETFHandler.InitializeDefaultETFs)
	router.GET("/api/universal-etf", universalETFHandler.GetAllETFs)
	router.GET("/api/universal-etf/:symbol", universalETFHandler.GetETFBySymbol)
	router.GET("/api/universal-etf/asset-class/:asset_class", universalETFHandler.GetETFsByAssetClass)
	router.GET("/api/universal-etf/region/:region", universalETFHandler.GetETFsByRegion)
	router.GET("/api/universal-etf/type/:etf_type", universalETFHandler.GetETFsByType)
	router.GET("/api/universal-etf/search", universalETFHandler.SearchETFs)
	router.POST("/api/universal-etf/filter", universalETFHandler.FilterETFs)
	router.GET("/api/universal-etf/distribution/asset-class", universalETFHandler.GetAssetClassDistribution)
	router.GET("/api/universal-etf/distribution/region", universalETFHandler.GetRegionDistribution)
	router.POST("/api/universal-etf/compare", universalETFHandler.CompareETFs)
	router.GET("/api/universal-etf/portfolio-allocation", universalETFHandler.GetPortfolioAllocation)
	router.GET("/api/universal-etf/categories", universalETFHandler.GetCategories)
	router.GET("/api/universal-etf/top-performers", universalETFHandler.GetTopPerformers)

	// 汇率管理路由
	exchangeRateHandler := handlers.NewExchangeRateHandler(exchangeRateConfig, exchangeRateTask)
	router.GET("/api/exchange-rates", exchangeRateHandler.GetExchangeRates)
	router.GET("/api/exchange-rates/:from/:to", exchangeRateHandler.GetExchangeRate)
	router.POST("/api/exchange-rates/convert", exchangeRateHandler.ConvertCurrency)
	router.POST("/api/exchange-rates/sync", exchangeRateHandler.TriggerSync)
	router.GET("/api/exchange-rates/summary", exchangeRateHandler.GetExchangeRatesSummary)
	router.GET("/api/exchange-rates/currencies", exchangeRateHandler.GetSupportedCurrencies)
	router.GET("/api/exchange-rates/datasource-status", exchangeRateHandler.GetDataSourceStatus)
	router.GET("/api/currency-pairs", exchangeRateHandler.GetCurrencyPairs)

	// 操作日志路由（需要认证）
	authMiddleware := middleware.NewAuthMiddleware(&cfg.JWT)
	logsGroup := router.Group("/api/logs")
	logsGroup.Use(authMiddleware.AuthRequired())
	logsGroup.Use(authMiddleware.RequirePermission("logs_view")) // 需要logs_view权限
	logsGroup.GET("/", operationLogsHandler.GetLogs)
	logsGroup.GET("/types", operationLogsHandler.GetLogTypes)
	logsGroup.GET("/action-types", operationLogsHandler.GetActionTypes)
	logsGroup.GET("/users", operationLogsHandler.GetUsers)
	logsGroup.POST("/export", operationLogsHandler.ExportLogs)
	logsGroup.GET("/:type/:id", operationLogsHandler.GetLogDetail)

	// Swagger API 文档路由
	docs.RegisterSwaggerRoutes(router)

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
