package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	// 从环境变量读取允许的源，默认为本地开发环境
	allowedOrigins := getCORSAllowedOrigins()
	allowedSet := make(map[string]bool)
	for _, o := range allowedOrigins {
		allowedSet[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 如果设置了允许的源且请求源在允许列表中，则允许跨域
		if origin != "" && allowedSet[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if len(allowedOrigins) == 0 {
			// 如果没有配置允许的源，为了开发便利，允许任意源（仅限开发环境）
			// 生产环境必须配置 ALLOWED_ORIGINS
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getCORSAllowedOrigins 从环境变量获取允许的跨域源
func getCORSAllowedOrigins() []string {
	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		// 默认允许本地开发环境
		return []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
		}
	}

	origins := strings.Split(originsStr, ",")
	// 清理空格
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		utils.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		)
	}
}
