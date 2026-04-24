package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) {
	utils.InitLogger("warn")
	models.InitDB(":memory:")
}

func TestToString(t *testing.T) {
	assert.Equal(t, "", toString(nil))
	assert.Equal(t, "", toString(""))
	assert.Equal(t, "test", toString("test"))
	assert.Equal(t, "", toString(123))
}

func TestExtractResource(t *testing.T) {
	assert.Equal(t, "unknown", extractResource(""))
	assert.Equal(t, "root", extractResource("/"))
	assert.Equal(t, "users", extractResource("/api/users"))
	assert.Equal(t, "123", extractResource("/api/users/123"))
	assert.Equal(t, "profile", extractResource("/profile"))
}

func TestSplitPath(t *testing.T) {
	// 空字符串返回 nil
	assert.Nil(t, splitPath(""))
	// 只有斜杠也返回 nil
	assert.Nil(t, splitPath("/"))
	assert.Equal(t, []string{"api", "users"}, splitPath("/api/users"))
	assert.Equal(t, []string{"api", "users", "123"}, splitPath("/api/users/123"))
}

func TestAuditLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	router.Use(AuditLogger())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `{"test": "data"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// 验证 X-Request-ID 头是否被设置
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestAuditLogger_WithSensitiveData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	router.Use(AuditLogger())
	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "logged in"})
	})

	body := `{"username": "test", "password": "secret123", "api_key": "key123"}`
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditLogger_WithUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user123")
		c.Set("username", "testuser")
		c.Next()
	})
	router.Use(AuditLogger())
	router.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "test"})
	})

	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBuffer(nil),
	}

	testData := []byte("test response data")
	n, err := rw.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, "test response data", rw.body.String())
}
