package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/routers"
	"etf-insight/services"
	"etf-insight/tasks"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		configPath = flag.String("config", "", "path to config file")
		initDB     = flag.Bool("init-db", false, "initialize database with default data")
		runOnce    = flag.Bool("run-once", false, "run update once and exit")
	)
	flag.Parse()

	// 设置日志
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load config")
	}

	// 初始化数据库
	if err := models.InitDB(&cfg.Database); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}

	// 自动迁移
	if err := models.AutoMigrate(); err != nil {
		logrus.WithError(err).Fatal("Failed to migrate database")
	}

	// 初始化默认数据
	if *initDB {
		if err := models.InitDefaultData(); err != nil {
			logrus.WithError(err).Fatal("Failed to initialize default data")
		}
		logrus.Info("Database initialized with default data")
		return
	}

	// 初始化服务
	cacheService := services.NewCacheService(&cfg.Redis, &cfg.ETF.Cache)
	defer cacheService.Close()

	exchangeService := services.NewExchangeRateService()
	analysisService := services.NewETFAnalysisService(cacheService, exchangeService)

	// 初始化调度器
	scheduler := tasks.NewScheduler(&cfg.Schedule, cacheService, analysisService, exchangeService)

	// 如果指定了run-once，执行一次更新并退出
	if *runOnce {
		scheduler.RunOnce()
		return
	}

	// 启动调度器
	scheduler.Start()
	defer scheduler.Stop()

	// 设置路由
	router := routers.SetupRouter(cfg, cacheService, analysisService, exchangeService, scheduler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logrus.Infof("Starting server on %s", addr)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Run(addr); err != nil {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	<-quit
	logrus.Info("Shutting down server...")
}
