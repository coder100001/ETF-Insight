package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	// API路由 - ETF列表
	mux.HandleFunc("/api/etf/list", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{"symbol": "QQQ", "name": "Invesco QQQ Trust", "market": "US", "category": "大盘股"},
				{"symbol": "SCHD", "name": "Schwab US Dividend Equity ETF", "market": "US", "category": "股息"},
				{"symbol": "VNQ", "name": "Vanguard Real Estate ETF", "market": "US", "category": "REITs"},
				{"symbol": "VYM", "name": "Vanguard High Dividend Yield ETF", "market": "US", "category": "股息"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// ETF对比数据
	mux.HandleFunc("/api/etf/comparison", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{"symbol": "QQQ", "name": "Invesco QQQ Trust", "price": 450.20, "change": 2.5, "yield": 0.56},
				{"symbol": "SCHD", "name": "Schwab US Dividend Equity ETF", "price": 26.80, "change": 0.8, "yield": 3.45},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// ETF详情相关路由 (metrics, history, forecast, realtime)
	mux.HandleFunc("/api/etf/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/etf/"):]
		parts := splitPath(path)

		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}

		symbol := parts[0]
		action := parts[1]

		switch action {
		case "realtime":
			data, err := cacheService.GetRealtimeData(symbol)
			if err != nil {
				// 返回模拟数据
				result := map[string]interface{}{
					"success": true,
					"data": map[string]interface{}{
						"symbol":         symbol,
						"name":           symbol + " ETF",
						"current_price":  100.0,
						"change":         1.5,
						"change_percent": 1.5,
						"volume":         1000000,
						"market_cap":     1000000000,
						"pe_ratio":       25.5,
						"dividend_yield": 1.5,
						"currency":       "USD",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(result)
				return
			}
			result := map[string]interface{}{
				"success": true,
				"data":    data,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case "metrics":
			period := r.URL.Query().Get("period")
			if period == "" {
				period = "1y"
			}
			result := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"symbol":        symbol,
					"period":        period,
					"volatility":    15.5,
					"sharpe_ratio":  1.2,
					"max_drawdown":  -12.3,
					"total_return":  25.5,
					"annual_return": 8.5,
					"beta":          1.05,
					"alpha":         2.3,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case "history":
			period := r.URL.Query().Get("period")
			if period == "" {
				period = "1y"
			}
			// 生成模拟历史数据
			var history []map[string]interface{}
			basePrice := 100.0
			for i := 0; i < 30; i++ {
				date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
				price := basePrice + float64(i)*0.5
				history = append([]map[string]interface{}{
					{
						"date":   date,
						"open":   price - 1,
						"high":   price + 2,
						"low":    price - 2,
						"close":  price,
						"volume": 1000000 + i*1000,
					},
				}, history...)
			}
			result := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"symbol":  symbol,
					"period":  period,
					"history": history,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case "forecast":
			initialInvestment := 10000.0
			taxRate := 0.10
			result := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"symbol":             symbol,
					"initial_investment": initialInvestment,
					"tax_rate":           taxRate,
					"forecasts": []map[string]interface{}{
						{"year": 1, "value": 10800, "return": 800},
						{"year": 3, "value": 12500, "return": 2500},
						{"year": 5, "value": 14500, "return": 4500},
						{"year": 10, "value": 21000, "return": 11000},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		default:
			http.NotFound(w, r)
		}
	})

	// 投资组合分析
	mux.HandleFunc("/api/etf/portfolio", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"total_value":  10000.0,
				"total_return": 500.0,
				"holdings": []map[string]interface{}{
					{"symbol": "QQQ", "weight": 50, "value": 5000},
					{"symbol": "SCHD", "weight": 50, "value": 5000},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// 汇率API
	mux.HandleFunc("/api/exchange-rates", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]interface{}{
			"success": true,
			"data": map[string]float64{
				"USD_CNY": 7.2,
				"USD_HKD": 7.8,
				"CNY_USD": 0.139,
				"HKD_USD": 0.128,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
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

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
