package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/router"
	"etf-insight/services"
	"etf-insight/services/datasource"
	"etf-insight/services/datasource/unified"
	"etf-insight/services/event"
	erdatasource "etf-insight/services/exchange_rate/datasource"
	"etf-insight/tasks"
	"etf-insight/utils"
)

type App struct {
	config           *config.Config
	router           *router.Router
	scheduler        *tasks.Scheduler
	exchangeRateTask *tasks.ExchangeRateTask
	server           *http.Server
	provider         datasource.DataSourceProvider
}

func New(configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	utils.InitLogger(cfg.Log.Level)
	utils.Info("Configuration loaded", "path", configPath)

	if err := models.InitDB(cfg.Database.GetDSN()); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	utils.Info("Database initialized")

	if err := models.InitDefaultData(); err != nil {
		return nil, fmt.Errorf("failed to initialize default data: %w", err)
	}
	utils.Info("Default data initialized")

	event.InitEventBus(models.DB)
	utils.Info("Event bus initialized")

	if err := models.InitExchangeRateTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize exchange rate tables: %w", err)
	}
	utils.Info("Exchange rate tables initialized")

	if err := models.InitDefaultCurrencyPairs(); err != nil {
		return nil, fmt.Errorf("failed to initialize default currency pairs: %w", err)
	}
	utils.Info("Default currency pairs initialized")

	exchangeService := services.NewExchangeRateService()
	analysisService := services.NewETFAnalysisService(exchangeService)
	optimizer := services.NewPortfolioOptimizer(analysisService)

	finageProvider := datasource.NewFinageProvider()
	utils.Info("Finage provider initialized",
		"available", finageProvider.IsAvailable(context.Background()))

	providerFactory := datasource.NewProviderFactory()
	providerFactory.Register("finage", finageProvider)
	providerFactory.Register("fallback", datasource.NewMockDataProvider())

	ctx := context.Background()
	defaultProvider, err := providerFactory.GetDefault(ctx)
	if err != nil {
		utils.Warn("No data source available, using mock provider", "error", err)
		defaultProvider = datasource.NewMockDataProvider()
	} else {
		utils.Info("Using data source", "provider", defaultProvider.GetName())
	}

	// 初始化统一数据源注册表
	initUnifiedRegistry(defaultProvider)

	scheduler := tasks.NewScheduler(&cfg.Schedule, analysisService, exchangeService, defaultProvider)

	exchangeRateConfig := &erdatasource.DataSourceConfig{
		OpenExchangeAPIKey: cfg.ExchangeRate.OpenExchangeAPIKey,
		CurrencyAPIKey:     cfg.ExchangeRate.CurrencyAPIKey,
	}
	exchangeRateTask := tasks.NewExchangeRateTask(exchangeRateConfig)

	r := router.NewRouter(
		cfg,
		analysisService,
		optimizer,
		exchangeService,
		defaultProvider,
		exchangeRateConfig,
		exchangeRateTask,
	)
	r.RegisterRoutes()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: r.GetEngine(),
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

	return &App{
		config:           cfg,
		router:           r,
		scheduler:        scheduler,
		exchangeRateTask: exchangeRateTask,
		server:           server,
		provider:         defaultProvider,
	}, nil
}

func (a *App) RunOnce() {
	a.scheduler.RunOnce()
}

func (a *App) Start() error {
	a.scheduler.Start()
	a.exchangeRateTask.Start()

	utils.Info("Starting server", "addr", a.server.Addr)

	errChan := make(chan error, 1)

	go func() {
		if a.config.Server.CertFile != "" && a.config.Server.KeyFile != "" {
			if _, err := os.Stat(a.config.Server.CertFile); err != nil {
				errChan <- fmt.Errorf("certificate file does not exist: %w", err)
				return
			}
			if _, err := os.Stat(a.config.Server.KeyFile); err != nil {
				errChan <- fmt.Errorf("key file does not exist: %w", err)
				return
			}
			utils.Info("HTTPS enabled", "cert", a.config.Server.CertFile)
			if err := a.server.ListenAndServeTLS(a.config.Server.CertFile, a.config.Server.KeyFile); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("failed to start HTTPS server: %w", err)
			}
		} else {
			utils.Warn("Running in HTTP mode (no TLS certificates provided)")
			if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("failed to start HTTP server: %w", err)
			}
		}
	}()

	if err := a.waitForServerReady(errChan); err != nil {
		a.scheduler.Stop()
		a.exchangeRateTask.Stop()
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return a.Shutdown()
}

func (a *App) waitForServerReady(errChan <-chan error) error {
	protocol := "http"
	httpClient := &http.Client{Timeout: 2 * time.Second}

	if a.config.Server.CertFile != "" && a.config.Server.KeyFile != "" {
		protocol = "https"
		if a.config.Server.InsecureHealthCheck {
			httpClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			utils.Warn("HTTPS health check with InsecureSkipVerify enabled - use only in development")
		}
	}

	healthURL := fmt.Sprintf("%s://localhost:%d/health", protocol, a.config.Server.Port)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-errChan:
			return err
		case <-timeout:
			return fmt.Errorf("server startup timeout: health check failed after 5 seconds")
		case <-ticker.C:
			resp, err := httpClient.Get(healthURL)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					utils.Info("Server started successfully", "addr", a.server.Addr)
					return nil
				}
			}
		}
	}
}

func (a *App) Shutdown() error {
	utils.Info("Shutting down server...")

	a.scheduler.Stop()
	a.exchangeRateTask.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		utils.Error("Server shutdown error", err)
		return err
	}

	utils.Info("Server stopped")
	return nil
}

// initUnifiedRegistry 初始化统一数据源注册表
// 将现有的 ETF 和汇率数据源注册到统一注册表中
func initUnifiedRegistry(etfProvider datasource.DataSourceProvider) {
	registry := unified.GetUnifiedRegistry()
	ctx := context.Background()

	// 注册 ETF 数据源
	if etfProvider != nil {
		adapter := unified.NewETFAdapter(etfProvider)
		registry.Register(etfProvider.GetName(), adapter)
		utils.Info("ETF data source registered to unified registry",
			"name", etfProvider.GetName(),
			"available", etfProvider.IsAvailable(ctx))
	}

	// 注册汇率数据源（使用 Fallback Provider，无需 API Key）
	fxProvider := erdatasource.NewFallbackProvider()
	if fxProvider != nil {
		adapter := unified.NewFXAdapter(fxProvider)
		registry.Register(fxProvider.GetName(), adapter)
		utils.Info("FX data source registered to unified registry",
			"name", fxProvider.GetName(),
			"available", fxProvider.IsAvailable(ctx))
	}

	utils.Info("Unified data source registry initialized",
		"total_providers", registry.Count())
}
