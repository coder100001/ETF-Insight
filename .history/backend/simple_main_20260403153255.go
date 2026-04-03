package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 获取当前工作目录
	currentDir, _ := os.Getwd()
	frontendDir := filepath.Join(currentDir, "..", "frontend", "dist")

	// 静态文件服务 - 前端页面
	r.Static("/assets", filepath.Join(frontendDir, "assets"))
	r.Static("/favicon.svg", filepath.Join(frontendDir, "favicon.svg"))
	r.Static("/icons.svg", filepath.Join(frontendDir, "icons.svg"))

	// 根路径 - 直接返回 index.html
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(filepath.Join(frontendDir, "index.html"))
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "ETF Insight API is running",
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// A 股 ETF
		aShare := v1.Group("/a-share")
		{
			aShare.GET("/etfs", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": []map[string]interface{}{
						{"symbol": "515080", "name": "中证红利 ETF", "dividend_yield": "4.8-5.1%", "frequency": "季分"},
						{"symbol": "515180", "name": "红利 ETF", "dividend_yield": "4.4-4.5%", "frequency": "年分"},
						{"symbol": "515300", "name": "红利低波 ETF", "dividend_yield": "4.4-4.5%", "frequency": "季分"},
						{"symbol": "510720", "name": "红利国企 ETF", "dividend_yield": "3.5-4%", "frequency": "月分"},
					},
				})
			})
		}

		// ETF 列表
		v1.GET("/etfs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": []map[string]interface{}{
					{"symbol": "QQQ", "name": "Invesco QQQ Trust", "currency": "USD"},
					{"symbol": "SCHD", "name": "Schwab US Dividend Equity ETF", "currency": "USD"},
					{"symbol": "VNQ", "name": "Vanguard Real Estate ETF", "currency": "USD"},
					{"symbol": "VYM", "name": "Vanguard High Dividend Yield ETF", "currency": "USD"},
				},
			})
		})
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
