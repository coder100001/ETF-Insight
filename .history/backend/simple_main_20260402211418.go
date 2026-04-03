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
		c.Header("Cache