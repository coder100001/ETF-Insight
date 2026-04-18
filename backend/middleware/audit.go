package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = rw

		c.Next()

		duration := time.Since(start)

		username, _ := c.Get("username")
		userID, _ := c.Get("user_id")

		metadata := map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
			"body_size":   len(requestBody),
		}

		if len(requestBody) > 0 {
			var jsonBody map[string]interface{}
			if json.Unmarshal(requestBody, &jsonBody) == nil {
				sensitiveFields := []string{"password", "token", "secret", "api_key", "authorization"}
				for _, field := range sensitiveFields {
					if _, ok := jsonBody[field]; ok {
						jsonBody[field] = "***REDACTED***"
					}
				}
				metadata["request_body"] = jsonBody
			}
		}

		metadataJSON, _ := json.Marshal(metadata)

		auditLog := &models.AuditLog{
			UserID:     toString(userID),
			Username:   toString(username),
			Action:     c.Request.Method,
			Resource:   c.FullPath(),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			RequestID:  requestID,
			Metadata:   string(metadataJSON),
			CreatedAt:  time.Now(),
		}

		if len(auditLog.Resource) == 0 || auditLog.Resource == c.Request.URL.Path {
			auditLog.Resource = extractResource(c.Request.URL.Path)
		}

		// 检查数据库连接是否可用
		if models.DB == nil {
			utils.Warn("Database not initialized, skipping audit log")
			return
		}

		go func(log *models.AuditLog) {
			if err := models.DB.Create(log).Error; err != nil {
				utils.Error("Failed to create audit log", err)
			}
		}(auditLog)
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func extractResource(path string) string {
	if len(path) == 0 {
		return "unknown"
	}
	parts := splitPath(path)
	if len(parts) == 0 {
		return "root"
	}
	return parts[len(parts)-1]
}

func splitPath(path string) []string {
	var parts []string
	var current []byte
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = nil
			}
		} else {
			current = append(current, path[i])
		}
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	return parts
}
