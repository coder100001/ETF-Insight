package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/tasks"
	"etf-insight/utils"
)

func main() {
	var (
		configPath = flag.String("config", "", "path to config file")
		runOnce    = flag.Bool("run-once", false, "run update once and exit")
	)
	flag.Parse()

	// 设置日志
	utils.InitLogger("info")

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		utils.Fatal("Failed to load config", err)
	}

	// 初始化数据库
	if err := models.InitDB(); err != nil {
		utils.Fatal("Failed to initialize database", err)
	}

	if err := models.AutoMigrate(); err != nil {
		utils.Fatal("Failed to migrate database", err)
	}

	// 初始化服务
	cacheService := services.NewCacheService(&cfg.ETF.Cache)
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

	// 设置HTTP路由
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// API路由
	mux.HandleFunc("/api/etf/quote/", func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Path[len("/api/etf/quote/"):]
		data, err := cacheService.GetRealtimeData(symbol)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/api/etf/analysis", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Symbols []string `json:"symbols"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result := analysisService.GetETFComparison(req.Symbols)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("/api/exchange-rates", func(w http.ResponseWriter, r *http.Request) {
		rates := exchangeService.GetAllRates()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rates)
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	utils.Info("Starting server", "addr", addr)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("Failed to start server", err)
		}
	}()

	<-quit
	utils.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.Error("Server shutdown error", err)
	}
}
