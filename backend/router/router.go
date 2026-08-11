package router

import (
	"etf-insight/config"
	"etf-insight/docs"
	"etf-insight/handlers"
	"etf-insight/middleware"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/datasource"
	"etf-insight/tasks"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ETF             *handlers.ETFHandler
	Portfolio       *handlers.PortfolioHandler
	Optimizer       *handlers.PortfolioOptimizerHandler
	MPT             *handlers.MPTHandler
	ETFHolding      *handlers.ETFHoldingHandler
	PortfolioPen    *handlers.PortfolioPenetrationHandler
	ETFConfig       *handlers.ETFConfigHandler
	ASharePortfolio *handlers.ASharePortfolioHandler
	AShareData      *handlers.AShareDataHandler
	ExchangeRate    *handlers.ExchangeRateHandler
	Export          *handlers.ExportHandler
	AuditLog        *handlers.AuditLogHandler
}

type Router struct {
	engine   *gin.Engine
	handlers *Handlers
	config   *config.Config
}

func NewRouter(
	cfg *config.Config,
	analysisService *services.ETFAnalysisService,
	optimizer *services.PortfolioOptimizer,
	provider datasource.DataSourceProvider,
	exchangeRateTask *tasks.ExchangeRateTask,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(handlers.LoggerMiddleware())
	engine.Use(handlers.CORSMiddleware())
	engine.Use(middleware.SecurityHeaders())
	engine.Use(middleware.RateLimiter())

	h := &Handlers{
		ETF:             handlers.NewETFHandler(analysisService, provider),
		Portfolio:       handlers.NewPortfolioHandler(analysisService),
		Optimizer:       handlers.NewPortfolioOptimizerHandler(optimizer),
		MPT:             handlers.NewMPTHandler(),
		ETFHolding:      handlers.NewETFHoldingHandler(models.DB),
		PortfolioPen:    handlers.NewPortfolioPenetrationHandler(models.DB),
		ETFConfig:       handlers.NewETFConfigHandler(),
		ASharePortfolio: handlers.NewASharePortfolioHandler(),
		AShareData:      handlers.NewAShareDataHandler(),
		ExchangeRate:    handlers.NewExchangeRateHandler(exchangeRateTask),
		Export:          handlers.NewExportHandler(),
		AuditLog:        handlers.NewAuditLogHandler(),
	}

	return &Router{engine: engine, handlers: h, config: cfg}
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

func (r *Router) RegisterRoutes() {
	r.registerHealthRoutes()
	r.registerETFRoutes()
	r.registerPortfolioRoutes()
	r.registerOptimizationRoutes()
	r.registerCacheRoutes()
	r.registerETFConfigRoutes()
	r.registerAShareRoutes()
	r.registerExchangeRateRoutes()
	r.registerExportRoutes()
	r.registerAuditLogRoutes()
	r.registerFinancialConfigRoutes()
	r.registerStaticRoutes()
	docs.RegisterSwaggerRoutes(r.engine)
}

func (r *Router) registerHealthRoutes() {
	r.engine.GET("/health", handlers.HealthHandler)
	r.engine.GET("/ready", handlers.ReadyHandler)
	r.engine.GET("/live", handlers.LiveHandler)
}

func (r *Router) registerETFRoutes() {
	etf := r.engine.Group("/api/etf")
	{
		etf.GET("/list", r.handlers.ETF.GetETFList)
		etf.GET("/comparison", r.handlers.ETF.GetETFComparison)
		etf.POST("/update-realtime", r.handlers.ETF.UpdateRealtimeData)
		etf.POST("/portfolio", r.handlers.Portfolio.AnalyzePortfolio)
		etf.GET("/data-status", r.handlers.ETF.GetDataSourceStatus)
		etf.GET("/:symbol/realtime", r.handlers.ETF.GetETFRealtime)
		etf.GET("/:symbol/history", r.handlers.ETF.GetETFHistory)
		etf.GET("/:symbol/metrics", r.handlers.ETF.GetETFMetrics)
		etf.GET("/:symbol/forecast", r.handlers.ETF.GetETFForecast)
		etf.GET("/:symbol/risk", r.handlers.ETF.GetETFRisk)
		etf.GET("/:symbol/similar", r.handlers.ETF.GetSimilarETFs)
		// ETF持仓相关路由
		etf.GET("/:symbol/holdings", r.handlers.ETFHolding.GetETFHoldings)
		etf.GET("/overlap", r.handlers.ETFHolding.GetETFOverlap)
		etf.GET("/:symbol/top-holdings", r.handlers.ETFHolding.GetTopHoldings)
		etf.GET("/:symbol/sector-allocation", r.handlers.ETFHolding.GetSectorAllocation)
		etf.POST("/holdings/comparison", r.handlers.ETFHolding.GetETFHoldingsComparison)
		etf.POST("/:symbol/holdings", r.handlers.ETFHolding.SaveETFHoldings)
	}
}

func (r *Router) registerPortfolioRoutes() {
	portfolio := r.engine.Group("/api/portfolio")
	{
		portfolio.POST("/scenarios", r.handlers.Portfolio.AnalyzeScenarios)
		portfolio.GET("/default-templates", r.handlers.Portfolio.GetDefaultPortfolioTemplates)
		portfolio.POST("/risk", r.handlers.Portfolio.AnalyzePortfolioRisk)
		portfolio.POST("/optimize", r.handlers.Optimizer.OptimizePortfolio)
		portfolio.POST("/efficient-frontier", r.handlers.Optimizer.GetEfficientFrontier)
		// 投资组合穿透分析相关路由
		portfolio.POST("/penetration", r.handlers.PortfolioPen.AnalyzePortfolioPenetration)
		portfolio.POST("/compare", r.handlers.PortfolioPen.ComparePortfolios)
		portfolio.POST("/sector-exposure", r.handlers.PortfolioPen.GetSectorExposure)
	}

	configs := r.engine.Group("/api/portfolio-configs")
	{
		configs.GET("/", r.handlers.Portfolio.GetPortfolioConfigs)
		configs.POST("/", r.handlers.Portfolio.CreatePortfolioConfig)
		configs.GET("/:id", r.handlers.Portfolio.GetPortfolioConfig)
		configs.PUT("/:id", r.handlers.Portfolio.UpdatePortfolioConfig)
		configs.DELETE("/:id", r.handlers.Portfolio.DeletePortfolioConfig)
		configs.POST("/:id/toggle-status", r.handlers.Portfolio.TogglePortfolioConfigStatus)
		configs.POST("/:id/analyze", r.handlers.Portfolio.AnalyzePortfolioConfig)
	}
}

func (r *Router) registerOptimizationRoutes() {
	opt := r.engine.Group("/api/optimization")
	{
		opt.POST("/mpt", r.handlers.MPT.MPTOptimize)
		opt.POST("/efficient-frontier", r.handlers.MPT.EfficientFrontier)
		opt.POST("/covariance", r.handlers.MPT.CalculateCovarianceMatrix)
	}
}

func (r *Router) registerCacheRoutes() {
	cache := r.engine.Group("/api/cache")
	{
		cache.GET("/overlap/stats", r.handlers.ETFHolding.GetCacheStats)
		cache.POST("/overlap/invalidate", r.handlers.ETFHolding.InvalidateCache)
		cache.POST("/overlap/clean", r.handlers.ETFHolding.CleanExpiredCache)
	}
}

func (r *Router) registerETFConfigRoutes() {
	configs := r.engine.Group("/api/etf-configs")
	{
		configs.GET("/", r.handlers.ETFConfig.GetETFConfigs)
		configs.POST("/", r.handlers.ETFConfig.CreateETFConfig)
		configs.GET("/:id", r.handlers.ETFConfig.GetETFConfig)
		configs.PUT("/:id", r.handlers.ETFConfig.UpdateETFConfig)
		configs.DELETE("/:id", r.handlers.ETFConfig.DeleteETFConfig)
		configs.POST("/:id/toggle-status", r.handlers.ETFConfig.ToggleETFConfigStatus)
		configs.POST("/:id/auto-update", r.handlers.ETFConfig.ToggleETFConfigAutoUpdate)
	}
}

func (r *Router) registerAShareRoutes() {
	aShare := r.engine.Group("/api/a-share")
	{
		aShare.GET("/etfs", r.handlers.ASharePortfolio.GetDefaultETFs)
		aShare.GET("/portfolio/default", r.handlers.ASharePortfolio.GetDefaultPortfolio)
		aShare.POST("/portfolio/analyze", r.handlers.ASharePortfolio.AnalyzePortfolio)
		aShare.POST("/portfolio/holding/:symbol", r.handlers.ASharePortfolio.UpdateHolding)
		aShare.GET("/dividend/:frequency", r.handlers.ASharePortfolio.CalculateDividendByFrequency)
		aShare.GET("/prices", r.handlers.ASharePortfolio.GetETFPrices)
		aShare.GET("/prices/:symbol", r.handlers.ASharePortfolio.GetETFPriceBySymbol)
		aShare.POST("/prices/refresh", r.handlers.ASharePortfolio.RefreshETFPrices)
		aShare.POST("/enable-akshare", r.handlers.AShareData.EnableAKShare)
		aShare.POST("/sync-etf-list", r.handlers.AShareData.SyncETFList)
		aShare.POST("/sync-prices", r.handlers.AShareData.SyncETFPrices)
		aShare.POST("/refresh-all", r.handlers.AShareData.RefreshAllData)
		aShare.GET("/price/:symbol", r.handlers.AShareData.GetETFPrice)
		aShare.GET("/all-prices", r.handlers.AShareData.GetPortfolioETFPrices)
		aShare.POST("/historical/:symbol", r.handlers.AShareData.GetHistoricalData)
		aShare.GET("/search", r.handlers.AShareData.SearchETFs)
		aShare.GET("/by-frequency/:frequency", r.handlers.AShareData.GetETFsByFrequency)
		aShare.GET("/dividend-yield/:symbol", r.handlers.AShareData.CalculateDividendYield)
		aShare.GET("/data-source-status", r.handlers.AShareData.GetDataSourceStatus)
	}
}

func (r *Router) registerExchangeRateRoutes() {
	exchange := r.engine.Group("/api/exchange-rates")
	{
		exchange.GET("/", r.handlers.ExchangeRate.GetExchangeRates)
		exchange.GET("/:from/:to", r.handlers.ExchangeRate.GetExchangeRate)
		exchange.POST("/convert", r.handlers.ExchangeRate.ConvertCurrency)
		exchange.POST("/sync", r.handlers.ExchangeRate.TriggerSync)
		exchange.GET("/summary", r.handlers.ExchangeRate.GetExchangeRatesSummary)
		exchange.GET("/currencies", r.handlers.ExchangeRate.GetSupportedCurrencies)
		exchange.GET("/datasource-status", r.handlers.ExchangeRate.GetDataSourceStatus)
	}
	r.engine.GET("/api/currency-pairs", r.handlers.ExchangeRate.GetCurrencyPairs)
}

func (r *Router) registerExportRoutes() {
	export := r.engine.Group("/api/export")
	{
		export.POST("/:type", r.handlers.Export.Export)
		export.GET("/formats", r.handlers.Export.GetSupportedFormats)
		export.GET("/types", r.handlers.Export.GetSupportedTypes)
	}
}

func (r *Router) registerAuditLogRoutes() {
	r.engine.GET("/api/logs", r.handlers.AuditLog.GetOperationLogs)
}

func (r *Router) registerFinancialConfigRoutes() {
	cfg := r.engine.Group("/api/config")
	{
		cfg.GET("/financial", handlers.GetFinancialConfig)
		cfg.PUT("/financial", handlers.UpdateFinancialConfig)
	}
}

func (r *Router) registerStaticRoutes() {
	r.engine.Static("/assets", "../frontend/dist/assets")
	r.engine.StaticFile("/favicon.svg", "../frontend/dist/favicon.svg")
	r.engine.StaticFile("/icons.svg", "../frontend/dist/icons.svg")
	r.engine.NoRoute(func(c *gin.Context) {
		c.File("../frontend/dist/index.html")
	})
}
