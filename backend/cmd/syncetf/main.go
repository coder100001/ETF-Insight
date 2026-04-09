package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"etf-insight/models"
	"etf-insight/services/datasource"
	"etf-insight/services/sync"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// SyncApplication 同步应用
type SyncApplication struct {
	syncService *sync.SyncService
	factory     *datasource.ProviderFactory
	ctx         context.Context
}

// NewSyncApplication 创建同步应用
func NewSyncApplication() (*SyncApplication, error) {
	ctx := context.Background()

	// 创建数据源工厂
	factory := datasource.NewProviderFactory()

	// 获取Finage API Key（优先环境变量）
	finageAPIKey := os.Getenv("FINAGE_API_KEY")
	if finageAPIKey != "" {
		// 注册Finage提供者
		finageProvider := datasource.NewFinageProvider(
			datasource.FinageConfig{
				APIKey:    finageAPIKey,
				Timeout:   30 * time.Second,
				RateLimit: 100,
				ProxyURL:  getEnvProxy(),
			},
		)
		factory.Register("finage", finageProvider)
		fmt.Println("✅ 已注册 Finage 数据源")
	}

	// 获取Finnhub API Key（优先环境变量）
	finnhubAPIKey := os.Getenv("FINNHUB_API_KEY")
	if finnhubAPIKey == "" {
		fmt.Println("⚠️ FINNHUB_API_KEY 环境变量未设置，Finnhub 数据源不可用")
	}

	// 注册Finnhub提供者（仅当 API Key 已配置时）
	if finnhubAPIKey != "" {
		finnhubProvider := datasource.NewFinnhubProvider(
			datasource.FinnhubConfig{
				APIKey:    finnhubAPIKey,
				Timeout:   30 * time.Second,
				RateLimit: 60,
				ProxyURL:  getEnvProxy(),
			},
		)
		factory.Register("finnhub", finnhubProvider)
		fmt.Println("✅ 已注册 Finnhub 数据源")
	}

	// 注册后备提供者
	factory.Register("fallback", datasource.NewFallbackProvider())

	// 创建同步服务（自动选择可用数据源）
	syncService, err := sync.NewSyncServiceWithFactory(factory)
	if err != nil {
		fmt.Printf("警告: %v，将使用后备数据源\n", err)
		syncService = sync.NewSyncService(datasource.NewFallbackProvider())
	}

	return &SyncApplication{
		syncService: syncService,
		factory:     factory,
		ctx:         ctx,
	}, nil
}

func main() {
	// 初始化日志
	utils.InitLogger("info")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           ETF 数据同步服务 (微服务架构版)                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 初始化数据库
	if err := models.InitDB("etf_insight.db"); err != nil {
		fmt.Printf("❌ 数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 数据库连接成功")
	fmt.Println()

	// 创建应用
	app, err := NewSyncApplication()
	if err != nil {
		fmt.Printf("❌ 应用初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 显示配置信息
	app.showConfig()
	fmt.Println()

	// 检查数据源
	if err := app.checkDataSource(); err != nil {
		fmt.Printf("⚠️  数据源检查: %v\n", err)
	}
	fmt.Println()

	// 执行同步
	if err := app.runSync(); err != nil {
		fmt.Printf("❌ 同步失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ 所有任务执行完毕")
}

// showConfig 显示配置信息
func (app *SyncApplication) showConfig() {
	fmt.Println("📋 配置信息:")
	fmt.Println("─────────────")
	fmt.Printf("  数据源: %s\n", app.syncService.GetProvider().GetName())
	fmt.Printf("  代理: %s\n", getEnvProxy())
	fmt.Printf("  速率限制: %d calls/sec\n", app.syncService.GetProvider().GetRateLimit())
}

// checkDataSource 检查数据源可用性
func (app *SyncApplication) checkDataSource() error {
	ctx, cancel := context.WithTimeout(app.ctx, 10*time.Second)
	defer cancel()

	fmt.Println("🔍 检查数据源可用性...")

	// 检查Finage
	if provider, ok := app.factory.Get("finage"); ok {
		if provider.IsAvailable(ctx) {
			fmt.Println("  ✅ Finage API: 可用")
			return nil
		}
		fmt.Println("  ❌ Finage API: 不可用")
	}

	// 检查Finnhub
	if provider, ok := app.factory.Get("finnhub"); ok {
		if provider.IsAvailable(ctx) {
			fmt.Println("  ✅ Finnhub API: 可用")
			return nil
		}
		fmt.Println("  ❌ Finnhub API: 不可用")
	}

	// 检查后备数据源
	if provider, ok := app.factory.Get("fallback"); ok {
		if provider.IsAvailable(ctx) {
			fmt.Println("  ✅ 后备数据源: 可用")
			app.syncService.SwitchProvider("fallback")
			return nil
		}
	}

	return fmt.Errorf("没有可用的数据源")
}

// runSync 执行同步
func (app *SyncApplication) runSync() error {
	fmt.Println("🚀 开始同步ETF数据...")
	fmt.Println()

	// 获取ETF列表
	etfList := sync.GetDefaultETFList()
	fmt.Printf("📊 准备同步 %d 只ETF:\n", len(etfList))
	for i, etf := range etfList {
		fmt.Printf("   %2d. %-6s (%s)\n", i+1, etf.Symbol, etf.Category)
	}
	fmt.Println()

	// 执行同步
	ctx, cancel := context.WithTimeout(app.ctx, 5*time.Minute)
	defer cancel()

	result, err := app.syncService.SyncETFs(ctx, etfList)
	if err != nil {
		return err
	}

	// 显示结果
	app.showResult(result)

	return nil
}

// showResult 显示同步结果
func (app *SyncApplication) showResult(result *sync.SyncResult) {
	fmt.Println()
	fmt.Println("📈 同步结果:")
	fmt.Println("═══════════════")
	fmt.Printf("  数据源: %s\n", result.DataSource)
	fmt.Printf("  总数: %d\n", result.TotalCount)
	fmt.Printf("  成功: %d ✅\n", result.SuccessCount)
	fmt.Printf("  失败: %d ❌\n", result.FailCount)
	fmt.Printf("  更新: %d 📝\n", result.UpdatedCount)
	fmt.Printf("  耗时: %s\n", result.Duration.Round(time.Second))
	fmt.Println()

	// 显示详细信息
	if len(result.Details) > 0 {
		fmt.Println("📋 详细结果:")
		fmt.Println("─────────────")
		for _, detail := range result.Details {
			status := "✅"
			if !detail.Success {
				status = "❌"
			}
			fmt.Printf("  %s %-6s $%7.2f (%+.2f%%) [%s]\n",
				status,
				detail.Symbol,
				detail.Price,
				detail.ChangePct,
				detail.DataSource,
			)
		}
	}

	// 显示错误
	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Println("⚠️  错误信息:")
		fmt.Println("─────────────")
		for i, err := range result.Errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}
}

// getEnvProxy 获取环境代理设置
func getEnvProxy() string {
	for _, env := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		if url := os.Getenv(env); url != "" {
			return url
		}
	}
	return "未设置"
}

// decimalNewFromFloat 转换为decimal
type decimalType interface{}

func decimalNewFromFloat(f float64) decimalType {
	return decimal.NewFromFloat(f)
}
