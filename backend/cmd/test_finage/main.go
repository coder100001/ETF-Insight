package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"etf-insight/services/datasource"
)

func main() {
	fmt.Println("Testing Finage Provider...")
	fmt.Println()

	// 从环境变量获取 API Key
	finageAPIKey := os.Getenv("FINAGE_API_KEY")
	if finageAPIKey == "" {
		fmt.Println("❌ FINAGE_API_KEY 环境变量未设置")
		return
	}

	// 创建 Finage provider
	provider := datasource.NewFinageProvider(
		datasource.FinageConfig{
			APIKey:    finageAPIKey,
			Timeout:   30 * time.Second,
			RateLimit: 100,
		},
	)

	fmt.Printf("Provider Name: %s\n", provider.GetName())
	fmt.Println()

	// 测试可用性
	ctx := context.Background()
	fmt.Printf("Testing availability...\n")
	available := provider.IsAvailable(ctx)
	fmt.Printf("Available: %v\n", available)
	fmt.Println()

	if !available {
		fmt.Println("Provider is not available!")
		return
	}

	// 测试获取报价
	fmt.Printf("Testing GetQuote for SCHD...\n")
	quote, err := provider.GetQuote(ctx, "SCHD")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Symbol: %s\n", quote.Symbol)
	fmt.Printf("Current Price: %.2f\n", quote.CurrentPrice)
	fmt.Printf("Bid: %.2f\n", quote.DayLow)
	fmt.Printf("Ask: %.2f\n", quote.DayHigh)
	fmt.Printf("Timestamp: %v\n", quote.Timestamp)
	fmt.Printf("Data Source: %s\n", quote.DataSource)
}
