package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"etf-insight/services/datasource"
)

func main() {
	fmt.Println("Testing Provider Factory...")
	fmt.Println()

	// 从环境变量获取 API Key
	finageAPIKey := os.Getenv("FINAGE_API_KEY")
	if finageAPIKey == "" {
		fmt.Println("❌ FINAGE_API_KEY 环境变量未设置")
		return
	}

	// 创建工厂
	factory := datasource.NewProviderFactory()

	// 注册 Finage provider
	finageProvider := datasource.NewFinageProvider(
		datasource.FinageConfig{
			APIKey:    finageAPIKey,
			Timeout:   30 * time.Second,
			RateLimit: 100,
		},
	)
	factory.Register("finage", finageProvider)
	fmt.Println("✅ Finage provider registered")

	// 注册 fallback provider
	factory.Register("fallback", datasource.NewFallbackProvider())
	fmt.Println("✅ Fallback provider registered")

	fmt.Println()
	fmt.Println("Registered providers:", factory.ListProviders())
	fmt.Println()

	// 获取默认 provider
	ctx := context.Background()
	fmt.Println("Getting default provider...")
	provider, err := factory.GetDefault(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Selected provider: %s\n", provider.GetName())
	fmt.Println()

	// 测试获取报价
	fmt.Println("Testing GetQuote for SCHD...")
	quote, err := provider.GetQuote(ctx, "SCHD")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Symbol: %s\n", quote.Symbol)
	fmt.Printf("Current Price: %.2f\n", quote.CurrentPrice)
	fmt.Printf("Data Source: %s\n", quote.DataSource)
}
