package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 静态文件服务 - 前端页面
	r.Static("/assets", "./frontend/dist/assets")
	r.Static("/favicon.svg", "./frontend/dist/favicon.svg")
	r.Static("/icons.svg", "./frontend/dist/icons.svg")

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
		// A股ETF
		aShare := v1.Group("/a-share")
		{
			aShare.GET("/etfs", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": []map[string]interface{}{
						{"symbol": "515080", "name": "中证红利ETF", "dividend_yield": "4.8-5.1%", "frequency": "季分"},
						{"symbol": "515180", "name": "红利ETF", "dividend_yield": "4.4-4.5