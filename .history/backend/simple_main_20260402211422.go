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

	// 根路径返回前端首页 - 使用 handler 直接读取文件内容，避免重定向缓存问题
	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("../frontend/dist/index.html")
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
		// A股ETF
		aShare := v1.Group("/a-share")
		{
			aShare.GET("/etfs", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": []map[string]interface{}{
						{"symbol": "515080", "name": "中证红利ETF", "dividend_yield": "4.8-5.1%", "frequency": "季分"},
						{"symbol": "515180", "name": "红利ETF", "dividend