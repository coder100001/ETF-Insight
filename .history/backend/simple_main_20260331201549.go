package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   "2024",
		})
	})

	// A 股 ETF 接口
	r.GET("/api/v1/a-share/etfs", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": []gin.H{
				{"symbol": "515080", "name": "中证红利ETF", "dividend_yield": "4.8-5.1"},
				{"symbol": "515180", "name": "红利ETF", "dividend_yield": "4.4-4.5"},
				{"symbol": "515300", "name": "红利低波ETF", "dividend_yield": "4.4-4.5"},
			},
		})
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
