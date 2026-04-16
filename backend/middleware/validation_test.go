package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidateInput_ValidData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "name",
			Type:     "string",
			Required: true,
		},
		{
			Field:    "age",
			Type:     "int",
			Required: true,
			Min:      0,
			Max:      150,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `{"name": "John", "age": 25}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateInput_MissingRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "name",
			Type:     "string",
			Required: true,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `{}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateInput_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "name",
			Type:     "string",
			Required: true,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `invalid json`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateInput_GETRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "name",
			Type:     "string",
			Required: true,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateInput_NumberOutOfRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "age",
			Type:     "int",
			Required: true,
			Min:      0,
			Max:      150,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `{"age": 200}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateInput_InvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "email",
			Type:     "email",
			Required: true,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `{"email": "invalid-email"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateInput_ValidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "email",
			Type:     "email",
			Required: true,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	body := `{"email": "test@example.com"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateInput_StringMinMaxLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "username",
			Type:     "string",
			Required: true,
			Min:      3,
			Max:      20,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试太短
	body := `{"username": "ab"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateInput_StringPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	rules := []ValidationRule{
		{
			Field:    "code",
			Type:     "string",
			Required: true,
			Pattern:  pattern,
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试不符合模式
	body := `{"code": "abc-123"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateInput_EnumValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rules := []ValidationRule{
		{
			Field:    "status",
			Type:     "string",
			Required: true,
			Enum:     []string{"active", "inactive", "pending"},
		},
	}

	router := gin.New()
	router.Use(ValidateInput(rules))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试无效枚举值
	body := `{"status": "deleted"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		{Field: "name", Message: "name is required"},
		{Field: "email", Message: "email is invalid"},
	}

	errMsg := errors.Error()
	assert.Contains(t, errMsg, "name is required")
	assert.Contains(t, errMsg, "email is invalid")
}
