# ETF-Insight 安全改进指南

## 📋 概述

本文档记录了 ETF-Insight 项目 v2.4 版本已实现的安全功能和改进建议。

**文档状态**: ✅ 已实施 (v2.4)
**最后更新**: 2026-04-14

---

## ✅ 已实现的安全功能 (v2.4)

### 1. 审计日志

**实现文件**: `backend/middleware/audit.go`

```go
// AuditLogger 审计日志中间件
func AuditLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        requestID := c.GetString("request_id")
        if requestID == "" {
            requestID = uuid.New().String()
            c.Set("request_id", requestID)
        }

        // 读取请求体
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }

        // 包装 ResponseWriter 以捕获响应
        blw := &bodyLogWriter{
            body:           bytes.NewBufferString(""),
            ResponseWriter: c.Writer,
        }
        c.Writer = blw

        c.Next()

        duration := time.Since(start)
        username, _ := c.Get("username")
        userID, _ := c.Get("user_id")

        // 构建审计日志
        auditLog := map[string]interface{}{
            "request_id":  requestID,
            "timestamp":   time.Now().Format(time.RFC3339),
            "method":      c.Request.Method,
            "path":        c.Request.URL.Path,
            "ip":          c.ClientIP(),
            "user_agent":  c.Request.UserAgent(),
            "user_id":     userID,
            "username":    username,
            "status_code": c.Writer.Status(),
            "duration_ms": duration.Milliseconds(),
        }

        // 异步写入日志
        go writeAuditLog(auditLog)
    }
}
```

**特性**:
- ✅ 自动记录所有 API 请求
- ✅ 敏感字段脱敏（password/token/secret/api_key）
- ✅ Request ID 追踪
- ✅ 异步写入，不影响性能

---

### 2. 数据验证

**实现文件**: `backend/middleware/validation.go`

```go
// ValidationRule 验证规则
type ValidationRule struct {
    Field    string
    Type     string  // string, number, email
    Required bool
    Min      float64
    Max      float64
    Pattern  *regexp.Regexp
    Enum     []string
}

// ValidateInput 通用输入验证中间件
func ValidateInput(rules []ValidationRule) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body map[string]interface{}
        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(400, gin.H{"error": "无效的请求体"})
            c.Abort()
            return
        }

        var errors []ValidationError
        for _, rule := range rules {
            value, exists := body[rule.Field]

            if rule.Required && !exists {
                errors = append(errors, ValidationError{
                    Field:   rule.Field,
                    Message: fmt.Sprintf("%s 是必填字段", rule.Field),
                })
                continue
            }

            if exists {
                if err := validateField(rule, value); err != nil {
                    errors = append(errors, ValidationError{
                        Field:   rule.Field,
                        Message: err.Error(),
                    })
                }
            }
        }

        if len(errors) > 0 {
            c.JSON(400, gin.H{
                "error":   "验证失败",
                "details": errors,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// ValidateSymbol 股票代码验证中间件
func ValidateSymbol() gin.HandlerFunc {
    return func(c *gin.Context) {
        symbol := c.Param("symbol")
        if symbol == "" {
            c.JSON(400, gin.H{"error": "股票代码不能为空"})
            c.Abort()
            return
        }

        // 验证股票代码格式
        if !isValidSymbol(symbol) {
            c.JSON(400, gin.H{"error": "无效的股票代码格式"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

### 3. 速率限制

**实现文件**: `backend/middleware/ratelimit.go`

```go
// RateLimiterHandler 速率限制中间件
func RateLimiterHandler(maxRequests int, window time.Duration) gin.HandlerFunc {
    limiter := NewRateLimiter(maxRequests, window)

    return func(c *gin.Context) {
        ip := c.ClientIP()

        if !limiter.Allow(ip) {
            c.JSON(429, gin.H{
                "error": "请求过于频繁，请稍后再试",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// RateLimiter 滑动窗口限流器
type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.RWMutex
    limit    int
    window   time.Duration
}

type visitor struct {
    timestamps []time.Time
}

func (rl *RateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    v, exists := rl.visitors[ip]

    if !exists {
        rl.visitors[ip] = &visitor{
            timestamps: []time.Time{now},
        }
        return true
    }

    // 清理过期时间戳
    cutoff := now.Add(-rl.window)
    valid := make([]time.Time, 0)
    for _, t := range v.timestamps {
        if t.After(cutoff) {
            valid = append(valid, t)
        }
    }

    if len(valid) >= rl.limit {
        return false
    }

    valid = append(valid, now)
    v.timestamps = valid
    return true
}
```

---

### 4. CORS 安全配置

**实现文件**: `backend/handlers/middleware.go`

```go
// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        // 允许的域名列表
        allowedOrigins := []string{
            "http://localhost:3000",
            "http://localhost:5173",
            "http://localhost:8080",
        }

        allowed := false
        for _, allowedOrigin := range allowedOrigins {
            if origin == allowedOrigin {
                allowed = true
                break
            }
        }

        if allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        }

        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

---

### 5. 安全响应头

**实现文件**: `backend/middleware/security.go`

```go
// SecurityHeaders 安全响应头中间件
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // XSS 保护
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")

        // CSP
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")

        // HSTS (生产环境启用)
        // c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

        c.Next()
    }
}
```

---

## 📊 安全功能使用示例

### 在 main.go 中启用所有安全功能

```go
// backend/main.go

func main() {
    // ...
    router := gin.New()

    // 安全中间件（按顺序）
    router.Use(middleware.SecurityHeaders())      // 1. 安全响应头
    router.Use(middleware.RateLimiterHandler(100, time.Minute)) // 2. 速率限制
    router.Use(middleware.AuditLogger())          // 3. 审计日志
    router.Use(handlers.CORSMiddleware())         // 4. CORS

    // 公开路由
    router.GET("/health", healthHandler)

    // API 路由
    api := router.Group("/api")
    {
        api.GET("/etf/list", etfHandler.GetList)
        api.POST("/portfolio/analyze", portfolioHandler.Analyze)
    }

    // ...
}
```

---

## 🔒 安全最佳实践

### 1. API Key 管理

```go
// ✅ 正确：从环境变量读取
apiKey := os.Getenv("FINAGE_API_KEY")

// ❌ 错误：硬编码
apiKey := "your_api_key_here"
```

### 2. 日志脱敏

```go
// ✅ 正确：敏感信息脱敏
utils.Info("API request", "url", sanitizeURL(reqURL))

func sanitizeURL(url string) string {
    re := regexp.MustCompile(`apikey=[^&\s]+`)
    return re.ReplaceAllString(url, "apikey=***")
}
```

### 3. 输入验证

```go
// ✅ 正确：验证股票代码
router.GET("/api/etf/:symbol", middleware.ValidateSymbol(), handler.GetETF)

// ✅ 正确：验证请求体
router.POST("/api/etf", middleware.ValidateInput([]middleware.ValidationRule{
    {Field: "symbol", Type: "string", Required: true, Min: 1, Max: 20},
    {Field: "price", Type: "number", Required: true, Min: 0},
}), handler.CreateETF)
```

---

## 📋 安全检查清单

### 部署前检查
- [x] 审计日志已配置
- [x] 速率限制已启用
- [x] CORS 已配置
- [x] 安全响应头已添加
- [x] 输入验证已完善
- [x] API Key 未硬编码
- [x] 日志脱敏已实现

### 定期审计
- [ ] 依赖包安全扫描（每周）
- [ ] 访问日志审查（每月）
- [ ] 权限配置检查（每月）

---

**文档版本**: v2.4
**维护者**: ETF-Insight Team
