package export

import (
	"etf-insight/models"
	"time"
)

// RecordExport 记录导出操作到审计日志
func RecordExport(userID, username, pageType, format string, statusCode int, err error) {
	auditLog := &models.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     "export",
		Resource:   pageType,
		ResourceID: format,
		Method:     "POST",
		Path:       "/api/export/" + pageType,
		StatusCode: statusCode,
		CreatedAt:  time.Now(),
	}

	if err != nil {
		auditLog.Error = err.Error()
	}

	// 异步写入数据库
	go func() {
		if models.DB != nil {
			models.DB.Create(auditLog)
		}
	}()
}

// RecordExportWithRequestID 记录导出操作到审计日志（包含请求ID）
func RecordExportWithRequestID(userID, username, pageType, format, requestID string, statusCode int, err error) {
	auditLog := &models.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     "export",
		Resource:   pageType,
		ResourceID: format,
		Method:     "POST",
		Path:       "/api/export/" + pageType,
		StatusCode: statusCode,
		RequestID:  requestID,
		CreatedAt:  time.Now(),
	}

	if err != nil {
		auditLog.Error = err.Error()
	}

	// 异步写入数据库
	go func() {
		if models.DB != nil {
			models.DB.Create(auditLog)
		}
	}()
}
