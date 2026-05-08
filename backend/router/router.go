package router

import (
	"etf-insight/config"
	"etf-insight/docs"
	"etf-insight/handlers"
	"etf-insight/middleware"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/datasource"
	erdatasource "etf-insight/services/exchange_rate/datasource"
	"etf-insight/tasks"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ETF             *handlers.ETFHandler
	Portfolio       *handlers.PortfolioHandler
	Optimizer       *handlers.PortfolioOptimizerHandler
	Optimization    *handlers.OptimizationHandler
	Factor          *handlers.FactorHandler
	FactorTiming    *handlers.FactorTimingHandler
	AlphaView       *handlers.AlphaViewHandler
	BlackLitterman  *handlers.BlackLittermanHandler
	ETFHolding      *handlers.ETFHoldingHandler
	PortfolioPen    *handlers.PortfolioPenetrationHandler
	Backtest        *handlers.BacktestHandler
	ETFConfig       *handlers.ETFConfigHandler
	ASharePortfolio *handlers.ASharePortfolioHandler
	AShareData      *handlers.AShareDataHandler
	UniversalETF    *handlers.UniversalETFHandler
	ExchangeRate    *handlers.ExchangeRateHandler
	OperationLogs   *handlers.OperationLogsHandler
	Report          *handlers.ReportHandler
	QuantLib        *handlers.QuantLibHandler
	Agent           *handlers.AgentHandler
	Data            *handlers.DataHandler
	Analytics       *handlers.AnalyticsHandler
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
	exchangeService *services.ExchangeRateService,
	provider datasource.DataSourceProvider,
	exchangeRateConfig *erdatasource.DataSourceConfig,
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
		Optimization:    handlers.NewOptimizationHandler(),
		Factor:          handlers.NewFactorHandler(),
		ETFHolding:      handlers.NewETFHoldingHandler(models.DB),
		PortfolioPen:    handlers.NewPortfolioPenetrationHandler(models.DB),
		Backtest:        handlers.NewBacktestHandler(),
		ETFConfig:       handlers.NewETFConfigHandler(),
		ASharePortfolio: handlers.NewASharePortfolioHandler(),
		AShareData:      handlers.NewAShareDataHandler(),
		UniversalETF:    handlers.NewUniversalETFHandler(),
		ExchangeRate:    handlers.NewExchangeRateHandler(exchangeRateConfig, exchangeRateTask),
		OperationLogs:   handlers.NewOperationLogsHandler(services.NewOperationLogsService(models.DB)),
		Report:          handlers.NewReportHandler(services.NewReportService(models.DB)),
	}

	factorDataService := services.NewFactorDataService(models.DB)
	alphaViewService := services.NewAlphaViewService(models.DB, factorDataService)
	blService := services.NewBlackLittermanService(models.DB, alphaViewService)

	h.FactorTiming = handlers.NewFactorTimingHandler(factorDataService)
	h.AlphaView = handlers.NewAlphaViewHandler(alphaViewService)
	h.BlackLitterman = handlers.NewBlackLittermanHandler(blService)
	h.QuantLib = handlers.NewQuantLibHandler()
	h.Agent = handlers.NewAgentHandler()
	h.Data = handlers.NewDataHandler()
	h.Analytics = handlers.NewAnalyticsHandler()

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
	r.registerRiskBudgetRoutes()
	r.registerFactorRoutes()
	r.registerFactorTimingRoutes()
	r.registerAlphaViewRoutes()
	r.registerBlackLittermanRoutes()
	r.registerCacheRoutes()
	r.registerBacktestRoutes()
	r.registerETFConfigRoutes()
	r.registerAShareRoutes()
	r.registerUniversalETFRoutes()
	r.registerExchangeRateRoutes()
	r.registerOperationLogsRoutes()
	r.registerReportRoutes()
	r.registerQuantLibRoutes()
	r.registerAgentRoutes()
	r.registerDataRoutes()
	r.registerAnalyticsRoutes()
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
		etf.GET("/:symbol/realtime", r.handlers.ETF.GetETFRealtime)
		etf.GET("/:symbol/history", r.handlers.ETF.GetETFHistory)
		etf.GET("/:symbol/metrics", r.handlers.ETF.GetETFMetrics)
		etf.GET("/:symbol/forecast", r.handlers.ETF.GetETFForecast)
		etf.GET("/:symbol/risk", r.handlers.ETF.GetETFRisk)
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
		opt.POST("/mpt", r.handlers.Optimization.MPTOptimize)
		opt.POST("/efficient-frontier", r.handlers.Optimization.EfficientFrontier)
		opt.POST("/covariance", r.handlers.Optimization.CalculateCovarianceMatrix)
		opt.POST("/etf-statistics", r.handlers.Optimization.GetETFStatistics)
		opt.POST("/risk-parity", r.handlers.Optimization.RiskParityOptimize)
		opt.POST("/black-litterman", r.handlers.Optimization.BlackLittermanOptimize)
		opt.POST("/market-implied-returns", r.handlers.Optimization.MarketImpliedReturns)
	}
}

func (r *Router) registerRiskBudgetRoutes() {
	rb := r.engine.Group("/api/risk-budget")
	{
		rb.GET("/configs", r.handlers.Optimization.GetRiskBudgetConfigs)
		rb.POST("/configs", r.handlers.Optimization.CreateRiskBudgetConfig)
	}
}

func (r *Router) registerFactorRoutes() {
	factor := r.engine.Group("/api/factor")
	{
		factor.POST("/analyze", r.handlers.Factor.AnalyzeFactorExposure)
		factor.POST("/portfolio", r.handlers.Factor.AnalyzePortfolioFactors)
		factor.POST("/multi-asset", r.handlers.Factor.AnalyzeMultipleAssets)
		factor.GET("/statistics", r.handlers.Factor.GetFactorStatistics)
		factor.POST("/risk-decomposition", r.handlers.Factor.DecomposeRisk)
		factor.POST("/compare", r.handlers.Factor.CompareFactorAttribution)
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

func (r *Router) registerBacktestRoutes() {
	backtest := r.engine.Group("/api/backtest")
	{
		backtest.POST("/run", r.handlers.Backtest.RunBacktest)
		backtest.POST("/event-driven", r.handlers.Backtest.RunEventDrivenBacktest)
		backtest.GET("/strategies", r.handlers.Backtest.ListStrategies)
		backtest.POST("/factors", r.handlers.Backtest.AnalyzeFactors)
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

func (r *Router) registerUniversalETFRoutes() {
	universal := r.engine.Group("/api/universal-etf")
	{
		universal.POST("/initialize", r.handlers.UniversalETF.InitializeDefaultETFs)
		universal.GET("/", r.handlers.UniversalETF.GetAllETFs)
		universal.GET("/:symbol", r.handlers.UniversalETF.GetETFBySymbol)
		universal.GET("/asset-class/:asset_class", r.handlers.UniversalETF.GetETFsByAssetClass)
		universal.GET("/region/:region", r.handlers.UniversalETF.GetETFsByRegion)
		universal.GET("/type/:etf_type", r.handlers.UniversalETF.GetETFsByType)
		universal.GET("/search", r.handlers.UniversalETF.SearchETFs)
		universal.POST("/filter", r.handlers.UniversalETF.FilterETFs)
		universal.GET("/distribution/asset-class", r.handlers.UniversalETF.GetAssetClassDistribution)
		universal.GET("/distribution/region", r.handlers.UniversalETF.GetRegionDistribution)
		universal.POST("/compare", r.handlers.UniversalETF.CompareETFs)
		universal.GET("/portfolio-allocation", r.handlers.UniversalETF.GetPortfolioAllocation)
		universal.GET("/categories", r.handlers.UniversalETF.GetCategories)
		universal.GET("/top-performers", r.handlers.UniversalETF.GetTopPerformers)
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

func (r *Router) registerOperationLogsRoutes() {
	logs := r.engine.Group("/api/logs")
	{
		logs.GET("", r.handlers.OperationLogs.GetLogs)
		logs.GET("/", r.handlers.OperationLogs.GetLogs)
		logs.GET("/types", r.handlers.OperationLogs.GetLogTypes)
		logs.GET("/action-types", r.handlers.OperationLogs.GetActionTypes)
		logs.GET("/users", r.handlers.OperationLogs.GetUsers)
		logs.POST("/export", r.handlers.OperationLogs.ExportLogs)
		logs.GET("/:type/:id", r.handlers.OperationLogs.GetLogDetail)
	}
}

func (r *Router) registerReportRoutes() {
	reports := r.engine.Group("/api/reports")
	{
		// 模板相关路由
		reports.GET("/templates", r.handlers.Report.GetTemplates)
		reports.GET("/templates/default", r.handlers.Report.GetDefaultTemplates)
		reports.GET("/templates/:id", r.handlers.Report.GetTemplate)
		reports.POST("/templates", r.handlers.Report.CreateTemplate)
		reports.PUT("/templates/:id", r.handlers.Report.UpdateTemplate)
		reports.DELETE("/templates/:id", r.handlers.Report.DeleteTemplate)

		// 报告生成相关路由
		reports.POST("/generate", r.handlers.Report.GenerateReport)
		reports.GET("", r.handlers.Report.GetReports)
		reports.GET("/:id", r.handlers.Report.GetReport)
		reports.GET("/:id/download", r.handlers.Report.DownloadReport)
		reports.DELETE("/:id", r.handlers.Report.DeleteReport)
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

func (r *Router) registerFactorTimingRoutes() {
	timing := r.engine.Group("/api/factor/timing")
	{
		timing.POST("/calculate", r.handlers.FactorTiming.CalculateFactorTiming)
		timing.GET("/history/:factor_name", r.handlers.FactorTiming.GetFactorTimingHistory)
		timing.GET("/signal/:factor_name", r.handlers.FactorTiming.GetLatestSignal)
	}
}

func (r *Router) registerAlphaViewRoutes() {
	views := r.engine.Group("/api/alpha-views")
	{
		views.POST("/", r.handlers.AlphaView.CreateView)
		views.GET("/", r.handlers.AlphaView.GetActiveViews)
		views.GET("/active", r.handlers.AlphaView.GetActiveViews)
		views.GET("/:id", r.handlers.AlphaView.GetView)
		views.PUT("/:id", r.handlers.AlphaView.UpdateView)
		views.DELETE("/:id", r.handlers.AlphaView.DeactivateView)
		views.POST("/generate-from-factor", r.handlers.AlphaView.GenerateFromFactor)
	}
}

func (r *Router) registerBlackLittermanRoutes() {
	bl := r.engine.Group("/api/black-litterman")
	{
		bl.POST("/configs", r.handlers.BlackLitterman.CreateConfig)
		bl.GET("/configs/:id", r.handlers.BlackLitterman.GetConfig)
		bl.PUT("/configs/:id", r.handlers.BlackLitterman.UpdateConfig)
		bl.POST("/calculate", r.handlers.BlackLitterman.CalculatePosterior)
		bl.GET("/results/:id", r.handlers.BlackLitterman.GetPosteriorResults)
	}
}

func (r *Router) registerQuantLibRoutes() {
	ql := r.engine.Group("/api/quantlib")
	{
		ql.POST("/options/european", r.handlers.QuantLib.PriceEuropeanOption)
		ql.POST("/options/american", r.handlers.QuantLib.PriceAmericanOption)
		ql.POST("/options/greeks", r.handlers.QuantLib.CalculateGreeks)
		ql.POST("/yield-curve/build", r.handlers.QuantLib.BuildYieldCurve)
		ql.POST("/bonds/price", r.handlers.QuantLib.PriceBond)
		ql.POST("/risk/var", r.handlers.QuantLib.CalculateVaR)
		ql.GET("/reference/:type", r.handlers.QuantLib.GetReferenceData)
	}
}

func (r *Router) registerAgentRoutes() {
	ag := r.engine.Group("/api/agents")
	{
		ag.GET("/health", r.handlers.Agent.Health)
		ag.GET("/discover", r.handlers.Agent.Discover)
		ag.POST("/run", r.handlers.Agent.Run)
		ag.POST("/team", r.handlers.Agent.RunTeam)
	}
}

func (r *Router) registerDataRoutes() {
	d := r.engine.Group("/api/data")
	{
		d.GET("/health", r.handlers.Data.Health)
		d.GET("/fred/series/:series_id", r.handlers.Data.FredSeries)
		d.GET("/yfinance/quote/:symbol", r.handlers.Data.YFinanceQuote)
		d.GET("/akshare/stock/spot", r.handlers.Data.AkShareStockSpot)
	}
}

func (r *Router) registerAnalyticsRoutes() {
	a := r.engine.Group("/api/analytics")
	{
		a.GET("/health", r.handlers.Analytics.Health)
		a.POST("/optimize", r.handlers.Analytics.OptimizePortfolio)
		a.POST("/var", r.handlers.Analytics.CalculateVaR)
		a.POST("/capm", r.handlers.Analytics.CalculateCAPM)
	}
}
