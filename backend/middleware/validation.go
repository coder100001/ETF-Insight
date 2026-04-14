package middleware

import (
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ValidationRule struct {
	Field    string
	Type     string
	Required bool
	Min      float64
	Max      float64
	Pattern  *regexp.Regexp
	Enum     []string
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

func ValidateInput(rules []ValidationRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "无效的JSON格式",
				"code":    "INVALID_JSON",
			})
			c.Abort()
			return
		}

		errors := validateData(rules, data)
		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errors.Error(),
				"code":    "VALIDATION_ERROR",
				"details": errors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func validateData(rules []ValidationRule, data map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	for _, rule := range rules {
		value, exists := data[rule.Field]

		if !exists || value == nil {
			if rule.Required {
				errors = append(errors, ValidationError{Field: rule.Field, Message: rule.Field + " 是必填字段"})
			}
			continue
		}

		switch rule.Type {
		case "string":
			if err := validateString(value, rule); err != nil {
				errors = append(errors, *err)
			}
		case "number", "int", "float":
			if err := validateNumber(value, rule); err != nil {
				errors = append(errors, *err)
			}
		case "email":
			if err := validateEmail(value); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	return errors
}

func validateString(value interface{}, rule ValidationRule) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 必须是字符串"}
	}

	if rule.Min > 0 && float64(len(str)) < rule.Min {
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 长度不能小于 " + strconv.FormatFloat(rule.Min, 'f', 0, 64)}
	}

	if rule.Max > 0 && float64(len(str)) > rule.Max {
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 长度不能大于 " + strconv.FormatFloat(rule.Max, 'f', 0, 64)}
	}

	if rule.Pattern != nil && !rule.Pattern.MatchString(str) {
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 格式不正确"}
	}

	if len(rule.Enum) > 0 {
		found := false
		for _, e := range rule.Enum {
			if str == e {
				found = true
				break
			}
		}
		if !found {
			return &ValidationError{Field: rule.Field, Message: rule.Field + " 必须是指定的值之一"}
		}
	}

	return nil
}

func validateNumber(value interface{}, rule ValidationRule) *ValidationError {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case float32:
		num = float64(v)
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return &ValidationError{Field: rule.Field, Message: rule.Field + " 必须是数字"}
		}
	default:
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 必须是数字"}
	}

	if rule.Min != 0 && num < rule.Min {
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 不能小于 " + strconv.FormatFloat(rule.Min, 'f', 2, 64)}
	}

	if rule.Max != 0 && num > rule.Max {
		return &ValidationError{Field: rule.Field, Message: rule.Field + " 不能大于 " + strconv.FormatFloat(rule.Max, 'f', 2, 64)}
	}

	return nil
}

func validateEmail(value interface{}) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return &ValidationError{Field: "email", Message: "邮箱格式不正确"}
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return &ValidationError{Field: "email", Message: "邮箱格式不正确"}
	}

	return nil
}

func ValidateSymbol() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol")
		if symbol != "" {
			symbolRegex := regexp.MustCompile(`^[A-Za-z0-9\-\.]{1,20}$`)
			if !symbolRegex.MatchString(symbol) {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "无效的股票代码格式",
					"code":    "INVALID_SYMBOL",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func RateLimiterHandler(maxRequests int, window time.Duration) gin.HandlerFunc {
	type visitor struct {
		count    int
		lastSeen time.Time
	}

	visitors := make(map[string]*visitor)
	var mu sync.RWMutex

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > window {
			v.count = 1
			v.lastSeen = time.Now()
			mu.Unlock()
			c.Next()
			return
		}

		v.count++
		v.lastSeen = time.Now()

		if v.count > maxRequests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "请求过于频繁，请稍后再试",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		mu.Unlock()
		c.Next()
	}
}

func ValidateIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if net.ParseIP(ip) == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "无效的IP地址",
				"code":    "INVALID_IP",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
