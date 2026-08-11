package main

import (
	"flag"
	"fmt"
	"os"

	"etf-insight/core"
	"etf-insight/di"
	"etf-insight/utils"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		utils.Warn("Failed to load .env file", "error", err)
	}

	configPath := flag.String("config", "config.yaml", "path to config file")
	runOnce := flag.Bool("run-once", false, "run update once and exit")
	flag.Parse()

	// Phase 1: Bootstrap - 加载配置、初始化数据库等副作用操作
	cfg, err := core.Bootstrap(*configPath)
	if err != nil {
		utils.Fatal("Failed to bootstrap application", err)
	}

	// Phase 2: DI - 使用 wire 容器构造所有服务/任务/路由
	container, err := di.InitializeContainer(cfg)
	if err != nil {
		utils.Fatal("Failed to initialize dependency container", err)
	}

	// 注册路由
	container.Router.RegisterRoutes()

	// Phase 3: 组装 App 并启动
	app := core.NewApp(
		cfg,
		container.Router,
		container.Scheduler,
		container.ExchangeRateTask,
		container.Provider,
	)

	if *runOnce {
		app.RunOnce()
		return
	}

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
