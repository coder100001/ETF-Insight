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
	"etf-insight/services/datasource"
	"etf-insight/services/event"
	"etf-insight/tasks"
	"etf-insight/utils"
)

// App 应用结构体，持有所有运行时依赖
type App struct {
	config           *config.Config
	router           *router.Router
	scheduler        *tasks.Scheduler
	exchangeRateTask *tasks.ExchangeRateTask
	server           *http.Server
	provider         datasource.DataSourceProvider
}

// Bootstrap 执行应用启动前的所有初始化工作（副作用操作）：
// 加载配置、初始化日志、数据库连接、默认数据、事件总线、汇率表等。
// 返回加载后的配置对象，用于后续依赖注入。
func Bootstrap(configPath string) (*config.Config, error) {
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

	return cfg, nil
}

// NewApp 使用已构造好的依赖组装 App，创建 HTTP Server。
// 所有服务/任务/路由的构造由 wire 依赖注入容器完成，此函数仅负责组装和创建 Server。
func NewApp(
	cfg *config.Config,
	r *router.Router,
	scheduler *tasks.Scheduler,
	exchangeRateTask *tasks.ExchangeRateTask,
	provider datasource.DataSourceProvider,
) *App {
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
		provider:         provider,
	}
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
