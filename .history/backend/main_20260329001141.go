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

	"github.com/shopspring/decimal"
)

// initMockData 初始化模拟数据
func initMockData(cacheService *services.CacheService) {
	mockData := []*services.RealtimeData{
		{
			Symbol:        "SCHD",
			Name:          "Schwab US Dividend Equity ETF",
			CurrentPrice:  30.44,
			PreviousClose: 31.67,
			Change:        -1.23,
			ChangePercent: -3.88,
			OpenPrice:     30.35,
			DayHigh:       30.59,
			DayLow:        30.20,
			Volume:        8500000,
			DividendYield: 3.45,
		},
		{
			Symbol:        "SPYD",
			Name:          "SPDR S&P 500 High Dividend ETF",
			CurrentPrice:  47.85,
			PreviousClose: 48.14,
			Change:        -0.29,
			ChangePercent: -0.60,
			OpenPrice:     47.71,
			DayHigh:       48.09,
			DayLow:        47.47,
			Volume:        6200000,
			DividendYield: 4.12,
		},
		{
			Symbol:        "JEPQ",
			Name:          "JPMorgan Nasdaq Equity Premium Income ETF",
			CurrentPrice:  57.20,
			PreviousClose: 57.51,
			Change:        -0.31,
			ChangePercent: -0.54,
			OpenPrice:     57.03,
			DayHigh:       57.49,
			DayLow:        56.74,
			Volume:        4800000,
			DividendYield: 11.2,
		},
		{
			Symbol:        "JEPI",
			Name:          "JPMorgan Equity Premium Income ETF",
			CurrentPrice:  58.90,
			PreviousClose: 59.31,
			Change:        -0.41,
			ChangePercent: -0.69,
			OpenPrice:     58.72,
			DayHigh:       59.19,
			DayLow:        58.43,
			Volume:        9200000,
			DividendYield: 9.8,
		},
		{
			Symbol:        "VYM",
			Name:          "Vanguard High Dividend Yield ETF",
			CurrentPrice:  154.50,
			PreviousClose: 155.37,
			Change:        -0.87,
			ChangePercent: -0.56,
			OpenPrice:     154.04,
			DayHigh:       155.27,
			DayLow:        153.26,
			Volume:        3800000,
			DividendYield: 2.95,
		},
		{
			Symbol:        "QQQ",
			Name:          "Invesco QQQ Trust",
			CurrentPrice:  385.20,
			PreviousClose: 380.15,
			Change:        5.05,
			ChangePercent: 1.33,
			OpenPrice:     381.50,
			DayHigh:       386.20,
			DayLow:        380.10,
			Volume:        52000000,
			DividendYield: 0.65,
		},
	}

	for _, data := range mockData {
		cacheService.SetRealtimeData(data.Symbol, data)
	}

	utils.Info("Mock data initialized")
}

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

	// 预填充模拟数据
	initMockData(cacheService)

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

	// CORS中间件
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置CORS头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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

		// 解析请求体
		var req struct {
			Allocation      map[string]float64 `json:"allocation"`
			TotalInvestment float64            `json:"total_investment"`
			TaxRate         float64            `json:"tax_rate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// 使用默认数据
			req.Allocation = map[string]float64{"QQQ": 50, "SCHD": 50}
			req.TotalInvestment = 100000
			req.TaxRate = 0.10
		}

		// 调用分析服务
		portfolioResult, err := analysisService.AnalyzePortfolio(
			req.Allocation,
			decimal.NewFromFloat(req.TotalInvestment),
			decimal.NewFromFloat(req.TaxRate),
		)

		if err != nil {
			result := map[string]interface{}{
				"success": false,
				"message": "分析失败: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"total_value":      portfolioResult.TotalValue.InexactFloat64(),
				"total_return":     portfolioResult.TotalReturn.InexactFloat64(),
				"total_return_pct": portfolioResult.TotalReturnPercent.InexactFloat64(),
				"annual_dividend":  portfolioResult.AnnualDividendBeforeTax.InexactFloat64(),
				"dividend_yield":   portfolioResult.WeightedDividendYield.InexactFloat64(),
				"tax_rate":         portfolioResult.TaxRate.InexactFloat64(),
				"after_tax_return": portfolioResult.TotalReturnWithDividend.InexactFloat64(),
				"holdings": func() []map[string]interface{} {
					h := make([]map[string]interface{}, len(portfolioResult.Holdings))
					for i, holding := range portfolioResult.Holdings {
						h[i] = map[string]interface{}{
							"symbol":                     holding.Symbol,
							"weight":                     holding.Weight.InexactFloat64(),
							"value":                      holding.InvestmentUSD.InexactFloat64(),
							"name":                       holding.Name,
							"current_price":              holding.CurrentPrice.InexactFloat64(),
							"shares":                     holding.Shares.InexactFloat64(),
							"current_value":              holding.CurrentValueUSD.InexactFloat64(),
							"capital_gain":               holding.CapitalGain.InexactFloat64(),
							"capital_gain_percent":       holding.CapitalGainPercent.InexactFloat64(),
							"total_return":               holding.TotalReturn.InexactFloat64(),
							"volatility":                 holding.Volatility.InexactFloat64(),
							"dividend_yield":             holding.DividendYield.InexactFloat64(),
							"annual_dividend_before_tax": holding.AnnualDividendBeforeTax.InexactFloat64(),
							"annual_dividend_after_tax":  holding.AnnualDividendAfterTax.InexactFloat64(),
						}
					}
					return h
				}(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// 投资组合配置 API
	mux.HandleFunc("/api/portfolio-configs/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// 返回配置列表
			result := map[string]interface{}{
				"success": true,
				"data": []map[string]interface{}{
					{
						"id":               1,
						"name":             "保守型组合",
						"description":      "低风险稳健配置",
						"allocation":       map[string]float64{"SCHD": 60, "VNQ": 20, "VYM": 20},
						"total_investment": 50000,
						"status":           1,
						"created_at":       time.Now().Format(time.RFC3339),
					},
					{
						"id":               2,
						"name":             "成长型组合",
						"description":      "高风险高收益配置",
						"allocation":       map[string]float64{"QQQ": 70, "SCHD": 30},
						"total_investment": 100000,
						"status":           1,
						"created_at":       time.Now().Format(time.RFC3339),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case http.MethodPost:
			// 创建配置
			result := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id":         3,
					"name":       "新配置",
					"status":     1,
					"created_at": time.Now().Format(time.RFC3339),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
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
		Handler: corsHandler(mux),
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
