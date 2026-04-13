package datasource

import (
	"fmt"
	"time"
)

// 标准错误定义
var (
	// ErrNoAvailableProvider 无可用数据源错误
	ErrNoAvailableProvider = &ProviderError{
		Code:    "NO_AVAILABLE_PROVIDER",
		Message: "没有可用的数据源提供者",
	}

	// ErrAPILimitExceeded API限制超出错误
	ErrAPILimitExceeded = &ProviderError{
		Code:    "API_LIMIT_EXCEEDED",
		Message: "API调用次数超出限制",
	}

	// ErrInvalidAPIKey 无效API密钥错误
	ErrInvalidAPIKey = &ProviderError{
		Code:    "INVALID_API_KEY",
		Message: "API密钥无效或已过期",
	}

	// ErrNetworkError 网络错误
	ErrNetworkError = &ProviderError{
		Code:    "NETWORK_ERROR",
		Message: "网络连接失败",
	}

	// ErrDataParseError 数据解析错误
	ErrDataParseError = &ProviderError{
		Code:    "DATA_PARSE_ERROR",
		Message: "数据解析失败",
	}

	// ErrCurrencyNotSupported 货币不支持错误
	ErrCurrencyNotSupported = &ProviderError{
		Code:    "CURRENCY_NOT_SUPPORTED",
		Message: "目标货币不支持",
	}

	// ErrRateLimitExceeded 速率限制错误
	ErrRateLimitExceeded = &ProviderError{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "请求速率超出限制",
	}
)

// ProviderError 数据源错误
type ProviderError struct {
	Code      string // 错误代码
	Message   string // 错误消息
	Source    string // 数据源名称
	InnerErr  error  // 内部错误
	Timestamp string // 错误时间戳
}

// Error 实现error接口
func (e *ProviderError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("[%s][%s] %s", e.Source, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 获取内部错误
func (e *ProviderError) Unwrap() error {
	return e.InnerErr
}

// Wrap 包装错误
func Wrap(err error, source, code, message string) *ProviderError {
	return &ProviderError{
		Code:      code,
		Message:   message,
		Source:    source,
		InnerErr:  err,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// IsProviderError 检查错误是否为ProviderError类型
func IsProviderError(err error) bool {
	_, ok := err.(*ProviderError)
	return ok
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	if pe, ok := err.(*ProviderError); ok {
		return pe.Code
	}
	return "UNKNOWN_ERROR"
}

// IsTemporaryError 检查是否为临时错误（可重试）
func IsTemporaryError(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case "NETWORK_ERROR", "RATE_LIMIT_EXCEEDED", "API_LIMIT_EXCEEDED":
		return true
	default:
		return false
	}
}

// IsFatalError 检查是否为致命错误（不可重试）
func IsFatalError(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case "INVALID_API_KEY", "CURRENCY_NOT_SUPPORTED":
		return true
	default:
		return false
	}
}
