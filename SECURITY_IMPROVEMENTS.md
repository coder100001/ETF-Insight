# ETF-Insight 安全改进指南

## 📋 概述

本文档提供了针对 ETF-Insight 项目的安全改进建议和实施方案。所有改进按优先级分类，包含具体的代码示例和实施步骤。

---

## 🔴 P0 - 立即修复（安全风险）

### 1. 修复 CORS 配置

**当前问题**: `Access-Control-Allow-Origin: *` 允许所有域名访问

**安全风险**: CSRF 攻击、数据泄露

**修复方案**:

```go
// backend/handlers/middleware.go

package handlers

import (
    "net/http"
    "os"
    "strings"
    "github.com/gin-gonic/gin"
)

// CORSMiddleware CORS 中间件（安全版本）
func CORSMiddleware() gin.HandlerFunc {
    // 从环境变量读取允许的域名列表
    allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
    if allowedOriginsStr == "" {
        // 默认只允许本地开发环境
        allowedOriginsStr = "http://localhost:3000,http://localhost:8080"
    }
    
    allowedOrigins := strings.Split(allowedOriginsStr, ",")
    
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // 检查请求来源是否在允许列表中
        allowed := false
        for _, allowedOrigin := range allowedOrigins {
            if strings.TrimSpace(allowedOrigin) == origin {
                allowed = true
                break
            }
        }
        
        if allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        }
        
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

**环境变量配置**:
```bash
# .env
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080,https://yourdomain.com
```

---

### 2. 添加 CSRF 保护

**当前问题**: POST/PUT/DELETE 请求缺少 CSRF Token 验证

**安全风险**: CSRF 攻击

**修复方案**:

#### 步骤 1: 安装依赖
```bash
cd backend
go get github.com/utrack/gin-csrf
```

#### 步骤 2: 添加 CSRF 中间件
```go
// backend/middleware/csrf.go

package middleware

import (
    "os"
    "github.com/gin-gonic/gin"
    csrf "github.com/utrack/gin-csrf"
)

// CSRFMiddleware CSRF 保护中间件
func CSRFMiddleware() gin.HandlerFunc {
    secret := os.Getenv("CSRF_SECRET")
    if secret == "" {
        secret = "default-csrf-secret-change-in-production"
    }
    
    return csrf.Middleware(csrf.Options{
        Secret: secret,
        ErrorFunc: func(c *gin.Context) {
            c.JSON(403, gin.H{
                "success": false,
                "error": "CSRF token invalid or missing",
            })
            c.Abort()
        },
    })
}
```

#### 步骤 3: 在 main.go 中启用
```go
// backend/main.go

import "etf-insight/middleware"

func main() {
    // ...
    router.Use(middleware.CSRFMiddleware())
    // ...
}
```

#### 步骤 4: 前端获取和发送 CSRF Token
```typescript
// frontend/src/services/api.ts

// 获取 CSRF Token
const getCSRFToken = (): string => {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute('content') || '' : '';
};

// 在请求中添加 CSRF Token
async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const csrfToken = getCSRFToken();
  
  return fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
      ...options?.headers,
    },
  }).then(response => {
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
  });
}
```

**环境变量配置**:
```bash
# .env
CSRF_SECRET=your-random-secret-key-here-min-32-chars
```

---

### 3. 添加请求体大小限制

**当前问题**: 未限制请求体大小

**安全风险**: DoS 攻击、内存耗尽

**修复方案**:

```go
// backend/middleware/security.go

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// RequestSizeLimiter 请求体大小限制中间件
func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 默认限制 1MB
        if maxSize == 0 {
            maxSize = 1 << 20 // 1MB
        }
        
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
        
        c.Next()
    }
}
```

**在 main.go 中启用**:
```go
// backend/main.go

func main() {
    // ...
    router.Use(middleware.RequestSizeLimiter(1 << 20)) // 1MB
    // ...
}
```

---

### 4. 实现日志脱敏

**当前问题**: 日志可能包含 API Key 等敏感信息

**安全风险**: 敏感信息泄露

**修复方案**:

```go
// backend/utils/logger.go

import (
    "regexp"
    "strings"
)

// 敏感信息正则表达式
var (
    apiKeyRegex = regexp.MustCompile(`(?i)(apikey|api_key|token|secret|password)=([^&\s]+)`)
    emailRegex  = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

// SanitizeLog 脱敏日志内容
func SanitizeLog(msg string) string {
    // 脱敏 API Key
    msg = apiKeyRegex.ReplaceAllString(msg, "$1=***")
    
    // 脱敏邮箱（可选）
    msg = emailRegex.ReplaceAllStringFunc(msg, func(email string) string {
        parts := strings.Split(email, "@")
        if len(parts) == 2 {
            return parts[0][:1] + "***@" + parts[1]
        }
        return "***@***"
    })
    
    return msg
}

// Info 记录信息日志（自动脱敏）
func Info(msg string, keysAndValues ...interface{}) {
    msg = SanitizeLog(msg)
    logger.Info(msg, keysAndValues...)
}

// Error 记录错误日志（自动脱敏）
func Error(msg string, err error, keysAndValues ...interface{}) {
    msg = SanitizeLog(msg)
    if err != nil {
        errMsg := SanitizeLog(err.Error())
        keysAndValues = append(keysAndValues, "error", errMsg)
    }
    logger.Error(msg, keysAndValues...)
}
```

---

## 🟡 P1 - 短期修复（1-2周）

### 5. 配置数据库连接池

**当前问题**: 未配置数据库连接池参数

**性能风险**: 连接耗尽、性能下降

**修复方案**:

```go
// backend/models/db.go

import (
    "time"
    "gorm.io/gorm"
)

func InitDB(dsn string) error {
    var err error
    DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return err
    }

    // 配置连接池
    sqlDB, err := DB.DB()
    if err != nil {
        return err
    }

    // 设置最大空闲连接数
    sqlDB.SetMaxIdleConns(10)
    
    // 设置最大打开连接数
    sqlDB.SetMaxOpenConns(100)
    
    // 设置连接最大生命周期
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    // 设置连接最大空闲时间
    sqlDB.SetConnMaxIdleTime(10 * time.Minute)

    return nil
}
```

---

### 6. 添加输入验证

**当前问题**: 部分 API 缺少严格的输入验证

**安全风险**: SQL 注入、XSS 攻击

**修复方案**:

```go
// backend/handlers/etf_config_handler.go

import (
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

// 创建全局验证器
var validate = validator.New()

// CreateETFConfigRequest 创建 ETF 配置请求
type CreateETFConfigRequest struct {
    Symbol       string  `json:"symbol" binding:"required,min=1,max=10,alphanum"`
    Name         string  `json:"name" binding:"required,min=1,max=100"`
    Description  string  `json:"description" binding:"max=500"`
    ExpenseRatio float64 `json:"expense_ratio" binding:"gte=0,lte=10"`
    Currency     string  `json:"currency" binding:"required,len=3,alpha"`
    Exchange     string  `json:"exchange" binding:"required,max=50"`
}

func (h *ETFConfigHandler) CreateETFConfig(c *gin.Context) {
    var req CreateETFConfigRequest
    
    // 绑定并验证请求
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{
            "success": false,
            "error": "Invalid request: " + err.Error(),
        })
        return
    }
    
    // 额外的自定义验证
    if err := validate.Struct(req); err != nil {
        c.JSON(400, gin.H{
            "success": false,
            "error": "Validation failed: " + err.Error(),
        })
        return
    }
    
    // 处理请求...
}
```

**自定义验证器**:
```go
// backend/utils/validator.go

package utils

import (
    "regexp"
    "github.com/go-playground/validator/v10"
)

// RegisterCustomValidators 注册自定义验证器
func RegisterCustomValidators(v *validator.Validate) {
    // 验证股票代码格式
    v.RegisterValidation("stock_symbol", func(fl validator.FieldLevel) bool {
        symbol := fl.Field().String()
        matched, _ := regexp.MatchString(`^[A-Z]{1,5}$`, symbol)
        return matched
    })
    
    // 验证货币代码格式
    v.RegisterValidation("currency_code", func(fl validator.FieldLevel) bool {
        code := fl.Field().String()
        matched, _ := regexp.MatchString(`^[A-Z]{3}$`, code)
        return matched
    })
}
```

---

### 7. 优化速率限制器

**当前问题**: 速率限制器可能内存泄漏

**性能风险**: 内存耗尽

**修复方案**:

```go
// backend/middleware/security.go

import (
    "sync"
    "time"
    "container/list"
    "github.com/gin-gonic/gin"
)

type visitor struct {
    lastSeen time.Time
    count    int
}

type rateLimiter struct {
    visitors  map[string]*visitor
    lruList   *list.List
    lruMap    map[string]*list.Element
    mu        sync.RWMutex
    limit     int
    window    time.Duration
    maxVisitors int
}

func newRateLimiter(limit int, window time.Duration, maxVisitors int) *rateLimiter {
    if maxVisitors == 0 {
        maxVisitors = 10000 // 默认最多保存 10000 个访客
    }
    
    rl := &rateLimiter{
        visitors:    make(map[string]*visitor),
        lruList:     list.New(),
        lruMap:      make(map[string]*list.Element),
        limit:       limit,
        window:      window,
        maxVisitors: maxVisitors,
    }

    go rl.cleanupVisitors()

    return rl
}

func (rl *rateLimiter) cleanupVisitors() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        
        // 清理过期访客
        for ip, v := range rl.visitors {
            if now.Sub(v.lastSeen) > rl.window {
                delete(rl.visitors, ip)
                if elem, ok := rl.lruMap[ip]; ok {
                    rl.lruList.Remove(elem)
                    delete(rl.lruMap, ip)
                }
            }
        }
        
        // 如果访客数量超过限制，移除最久未使用的
        for len(rl.visitors) > rl.maxVisitors {
            elem := rl.lruList.Back()
            if elem != nil {
                ip := elem.Value.(string)
                delete(rl.visitors, ip)
                delete(rl.lruMap, ip)
                rl.lruList.Remove(elem)
            } else {
                break
            }
        }
        
        rl.mu.Unlock()
    }
}

func (rl *rateLimiter) allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    v, exists := rl.visitors[ip]
    if !exists {
        // 新访客
        rl.visitors[ip] = &visitor{
            lastSeen: time.Now(),
            count:    1,
        }
        
        // 添加到 LRU 列表
        elem := rl.lruList.PushFront(ip)
        rl.lruMap[ip] = elem
        
        return true
    }

    // 更新 LRU 位置
    if elem, ok := rl.lruMap[ip]; ok {
        rl.lruList.MoveToFront(elem)
    }

    if time.Since(v.lastSeen) > rl.window {
        v.count = 1
        v.lastSeen = time.Now()
        return true
    }

    if v.count >= rl.limit {
        return false
    }

    v.count++
    v.lastSeen = time.Now()
    return true
}

var limiter = newRateLimiter(100, time.Minute, 10000)

func RateLimiter() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()

        if !limiter.allow(ip) {
            c.JSON(429, gin.H{
                "success": false,
                "error": "Too many requests. Please try again later.",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

### 8. SQL 注入防护审计

**当前问题**: 需要确保所有查询都使用参数化

**安全风险**: SQL 注入

**审计清单**:

```go
// ✅ 正确：使用参数化查询
models.DB.Where("symbol = ?", symbol).First(&etf)
models.DB.Where("from_currency = ? AND to_currency = ?", from, to).First(&rate)

// ❌ 错误：字符串拼接
models.DB.Where(fmt.Sprintf("symbol = '%s'", symbol)).First(&etf)

// ✅ 正确：使用 GORM 的查询构建器
models.DB.Model(&ETFData{}).
    Where("symbol = ?", symbol).
    Where("date >= ?", startDate).
    Order("date DESC").
    Find(&data)

// ❌ 错误：原始 SQL 字符串拼接
query := fmt.Sprintf("SELECT * FROM etf_data WHERE symbol = '%s'", symbol)
models.DB.Raw(query).Scan(&data)

// ✅ 正确：原始 SQL 使用参数
models.DB.Raw("SELECT * FROM etf_data WHERE symbol = ?", symbol).Scan(&data)
```

**审计脚本**:
```bash
#!/bin/bash
# audit_sql.sh - 检查潜在的 SQL 注入风险

echo "Searching for potential SQL injection vulnerabilities..."

# 查找字符串拼接的 SQL 查询
grep -rn "fmt.Sprintf.*SELECT\|UPDATE\|DELETE\|INSERT" backend/ --include="*.go"

# 查找 Raw 查询
grep -rn "\.Raw(" backend/ --include="*.go"

# 查找 Exec 查询
grep -rn "\.Exec(" backend/ --include="*.go"

echo "Review the results above for potential SQL injection risks."
```

---

## 🟢 P2 - 中期改进（1个月）

### 9. 添加单元测试

**测试覆盖率目标**: >80%

**关键测试用例**:

```go
// backend/services/datasource/finage_provider_test.go

package datasource

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFinageProvider_GetQuote(t *testing.T) {
    provider := NewFinageProvider(FinageConfig{
        APIKey: "test_api_key",
    })
    
    ctx := context.Background()
    
    t.Run("Valid symbol", func(t *testing.T) {
        quote, err := provider.GetQuote(ctx, "AAPL")
        assert.NoError(t, err)
        assert.NotNil(t, quote)
        assert.Equal(t, "AAPL", quote.Symbol)
    })
    
    t.Run("Invalid symbol", func(t *testing.T) {
        _, err := provider.GetQuote(ctx, "")
        assert.Error(t, err)
    })
    
    t.Run("Context cancellation", func(t *testing.T) {
        ctx, cancel := context.WithCancel(ctx)
        cancel()
        
        _, err := provider.GetQuote(ctx, "AAPL")
        assert.Error(t, err)
    })
}
```

```go
// backend/services/exchange_rate/datasource/manager_test.go

package datasource

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/shopspring/decimal"
)

func TestDataSourceManager_Failover(t *testing.T) {
    // 创建模拟数据源
    primary := &MockProvider{
        name: "primary",
        available: false, // 主数据源不可用
    }
    
    backup := &MockProvider{
        name: "backup",
        available: true,
        rate: decimal.NewFromFloat(7.2),
    }
    
    manager := NewDataSourceManager(primary, backup)
    
    ctx := context.Background()
    
    t.Run("Failover to backup", func(t *testing.T) {
        rate, err := manager.GetRate(ctx, "USD", "CNY")
        assert.NoError(t, err)
        assert.Equal(t, decimal.NewFromFloat(7.2), rate)
        
        // 验证当前数据源是备份数据源
        current := manager.GetCurrentProvider()
        assert.Equal(t, "backup", current.GetName())
    })
}
```

---

### 10. 添加前端错误边界

**修复方案**:

```typescript
// frontend/src/components/ErrorBoundary.tsx

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Result, Button } from 'antd';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
    
    // 上报错误到监控系统
    this.reportError(error, errorInfo);
    
    this.setState({
      error,
      errorInfo,
    });
  }

  reportError(error: Error, errorInfo: ErrorInfo) {
    // 集成 Sentry 或其他错误监控服务
    // Sentry.captureException(error, { extra: errorInfo });
    
    // 或发送到自己的错误收集接口
    fetch('/api/errors', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
        timestamp: new Date().toISOString(),
      }),
    }).catch(console.error);
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      return (
        <Result
          status="error"
          title="出错了"
          subTitle="抱歉，页面遇到了一些问题。"
          extra={[
            <Button type="primary" key="reload" onClick={() => window.location.reload()}>
              刷新页面
            </Button>,
            <Button key="reset" onClick={this.handleReset}>
              返回
            </Button>,
          ]}
        >
          {process.env.NODE_ENV === 'development' && this.state.error && (
            <div style={{ textAlign: 'left', marginTop: 20 }}>
              <h4>错误详情：</h4>
              <pre style={{ background: '#f5f5f5', padding: 10, overflow: 'auto' }}>
                {this.state.error.toString()}
                {this.state.errorInfo?.componentStack}
              </pre>
            </div>
          )}
        </Result>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
```

**使用方式**:
```typescript
// frontend/src/App.tsx

import ErrorBoundary from './components/ErrorBoundary';

function App() {
  return (
    <ErrorBoundary>
      <Router>
        {/* 应用内容 */}
      </Router>
    </ErrorBoundary>
  );
}
```

---

## 📊 安全检查清单

### 部署前检查
- [ ] CORS 配置已限制允许的域名
- [ ] CSRF 保护已启用
- [ ] 请求体大小限制已配置
- [ ] 日志脱敏已实现
- [ ] 数据库连接池已配置
- [ ] 输入验证已完善
- [ ] 速率限制器已优化
- [ ] SQL 注入防护已审计
- [ ] HTTPS 已启用（生产环境）
- [ ] API Key 未硬编码
- [ ] 敏感信息已加密存储
- [ ] 错误信息不泄露敏感数据

### 定期安全审计
- [ ] 依赖包安全扫描（每周）
- [ ] 代码安全审查（每次 PR）
- [ ] 渗透测试（每季度）
- [ ] 日志审计（每月）
- [ ] 访问控制审查（每月）

---

## 🛠️ 安全工具推荐

### Go 安全工具
```bash
# 安全扫描
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# 依赖检查
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 静态分析
go vet ./...
```

### 前端安全工具
```bash
# npm 审计
npm audit

# 依赖更新
npm outdated
npm update

# Snyk 扫描
npx snyk test
```

---

## 📞 安全事件响应

### 发现安全漏洞时
1. **立即评估**: 确定漏洞严重程度
2. **隔离影响**: 如果可能，临时禁用受影响功能
3. **修复漏洞**: 按照本文档的修复方案实施
4. **测试验证**: 确保修复有效且不引入新问题
5. **部署更新**: 尽快部署到生产环境
6. **通知用户**: 如果涉及用户数据，及时通知
7. **事后分析**: 总结经验，更新安全流程

---

**文档版本**: v1.0  
**最后更新**: 2026-04-13  
**维护者**: ETF-Insight Security Team
