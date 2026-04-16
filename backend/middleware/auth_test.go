package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"etf-insight/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}

	auth := NewAuthMiddleware(cfg)

	assert.NotNil(t, auth)
	assert.Equal(t, "test-secret-key", string(auth.secretKey))
	assert.Equal(t, 24*time.Hour, auth.expiry)
}

func TestAuthMiddleware_GenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	token, err := auth.GenerateToken("user123", "testuser", "admin")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthMiddleware_ValidateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	// 生成token
	token, err := auth.GenerateToken("user123", "testuser", "admin")
	assert.NoError(t, err)

	// 验证token
	claims, err := auth.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
}

func TestAuthMiddleware_ValidateToken_Invalid(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	// 验证无效的token
	claims, err := auth.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthMiddleware_AuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	router := gin.New()
	router.Use(auth.AuthRequired())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试没有Authorization头
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_AuthRequired_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	router := gin.New()
	router.Use(auth.AuthRequired())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试无效的认证格式
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_AuthRequired_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	// 生成有效token
	token, err := auth.GenerateToken("user123", "testuser", "admin")
	assert.NoError(t, err)

	router := gin.New()
	router.Use(auth.AuthRequired())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试有效token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RoleRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	// 生成admin token
	token, err := auth.GenerateToken("user123", "testuser", "admin")
	assert.NoError(t, err)

	router := gin.New()
	router.Use(auth.AuthRequired())
	router.Use(auth.RequireRole("admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RoleRequired_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	// 生成user token (不是admin)
	token, err := auth.GenerateToken("user123", "testuser", "user")
	assert.NoError(t, err)

	router := gin.New()
	router.Use(auth.AuthRequired())
	router.Use(auth.RequireRole("admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOptionalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.JWTConfig{
		SecretKey:   "test-secret-key",
		ExpiryHours: 24,
	}
	auth := NewAuthMiddleware(cfg)

	router := gin.New()
	router.Use(auth.OptionalAuth())
	router.GET("/optional", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{"message": "authenticated", "user_id": userID})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "anonymous"})
		}
	})

	// 测试没有token也能访问
	req := httptest.NewRequest("GET", "/optional", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
