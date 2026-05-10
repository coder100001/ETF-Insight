package export

import "fmt"

// ExportError 导出错误结构
type ExportError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *ExportError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewError 创建新的导出错误
func NewError(code, message, details string) *ExportError {
	return &ExportError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 预定义错误
var (
	ErrDataMissing      = NewError("EXPORT_001", "数据缺失", "")
	ErrDataInvalid      = NewError("EXPORT_002", "数据无效", "")
	ErrDataTooLarge     = NewError("EXPORT_003", "数据过大", "")
	ErrFormatNotSupport = NewError("EXPORT_004", "格式不支持", "")
	ErrGenerateFailed   = NewError("EXPORT_005", "生成失败", "")
	ErrTimeout          = NewError("EXPORT_006", "操作超时", "")
)

// WrapError 包装错误
func WrapError(base *ExportError, details string) *ExportError {
	return &ExportError{
		Code:    base.Code,
		Message: base.Message,
		Details: details,
	}
}
