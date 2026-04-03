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
	r.Static("/favicon.svg", "./