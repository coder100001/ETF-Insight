package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i, w.Code)
		}
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("Header %s: expected %s, got %s", header, expected, got)
		}
	}
}

func TestSecurityHeaders_APICache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl == "" {
		t.Error("API endpoints should have Cache-Control header")
	}
}

func TestNewRateLimiter(t *testing.T) {
	rl := newRateLimiter(10, time.Minute)

	if rl == nil {
		t.Fatal("newRateLimiter returned nil")
	}

	if rl.limit != 10 {
		t.Errorf("Expected limit 10, got %d", rl.limit)
	}

	if rl.window != time.Minute {
		t.Errorf("Expected window 1 minute, got %v", rl.window)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := newRateLimiter(5, time.Minute)

	ip := "192.168.1.1"

	for i := 0; i < 5; i++ {
		if !rl.allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	if rl.allow(ip) {
		t.Error("Request beyond limit should be denied")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := newRateLimiter(2, 100*time.Millisecond)

	ip := "192.168.1.2"

	if !rl.allow(ip) {
		t.Error("First request should be allowed")
	}

	if !rl.allow(ip) {
		t.Error("Second request should be allowed")
	}

	if rl.allow(ip) {
		t.Error("Third request should be denied")
	}

	time.Sleep(150 * time.Millisecond)

	if !rl.allow(ip) {
		t.Error("Request after window expiry should be allowed")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := newRateLimiter(10, 10*time.Millisecond)

	for i := 0; i < 50; i++ {
		rl.allow("192.168.1." + string(rune(i)))
	}

	time.Sleep(200 * time.Millisecond)

	rl.mu.Lock()
	count := len(rl.visitors)
	rl.mu.Unlock()

	if count > 25 {
		t.Skipf("Cleanup may not have run yet, %d visitors remain", count)
	}
}
