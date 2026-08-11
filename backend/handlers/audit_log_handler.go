package handlers

import (
	"net/http"
	"strconv"
	"time"

	"etf-insight/models"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct{}

func NewAuditLogHandler() *AuditLogHandler {
	return &AuditLogHandler{}
}

// OperationLogItem 操作日志响应项（对齐前端 UnifiedLog 契约）
// 除统一字段外保留后端原始字段（operation_type/start_time 等），
// 兼容 Dashboard 的字段映射逻辑。
type OperationLogItem struct {
	ID           uint      `json:"id"`
	LogType      string    `json:"log_type"` // 固定 "operation"
	Timestamp    time.Time `json:"timestamp"`
	User         string    `json:"user"`
	Module       string    `json:"module"`
	ActionType   string    `json:"action_type"`
	Details      string    `json:"details"`
	IP           string    `json:"ip"`
	Status       string    `json:"status"` // processing/success/error
	StatusCode   int       `json:"status_code"`
	ErrorMessage string    `json:"error_message"`
	DurationMs   int       `json:"duration_ms"`

	// 后端原始字段（向后兼容）
	OperationType string     `json:"operation_type"`
	OperationName string     `json:"operation_name"`
	Operator      string     `json:"operator"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
}

// logStatusToInt 前端状态字符串 → OperationLog.Status 数值
// OperationLog.Status: 0:进行中, 1:成功, 2:失败
var logStatusToInt = map[string]int{
	"success":    1,
	"failure":    2,
	"error":      2,
	"processing": 0,
}

// operationLogStatus OperationLog.Status 数值 → 前端状态字符串
func operationLogStatus(status int) string {
	switch status {
	case 0:
		return "processing"
	case 1:
		return "success"
	case 2:
		return "error"
	default:
		return "unknown"
	}
}

// GetOperationLogs 获取操作日志列表
// 响应契约：data 为 OperationLogItem 数组，meta.pagination/meta.summary 供前端分页与统计展示。
// log_type=audit 时仅返回统计信息（审计日志由中间件写入 audit_logs 表，暂未提供查询接口）。
func (h *AuditLogHandler) GetOperationLogs(c *gin.Context) {
	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	logType := c.DefaultQuery("log_type", "operation")

	var (
		items []OperationLogItem
		total int64
	)

	if logType == "audit" {
		models.DB.Model(&models.AuditLog{}).Count(&total)
		items = []OperationLogItem{}
	} else {
		var logs []models.OperationLog
		// OperationLog 模型无 CreatedAt 字段，按 StartTime 排序
		query := models.DB.Order("start_time DESC")

		// 筛选参数（与前端 getLogs 参数对齐）
		if user := c.Query("user"); user != "" {
			query = query.Where("operator = ?", user)
		}
		if operator := c.Query("operator"); operator != "" {
			query = query.Where("operator = ?", operator)
		}
		if actionType := c.Query("action_type"); actionType != "" {
			query = query.Where("operation_type = ?", actionType)
		}
		if status := c.Query("status"); status != "" {
			if code, ok := logStatusToInt[status]; ok {
				query = query.Where("status = ?", code)
			}
		}

		query.Model(&models.OperationLog{}).Count(&total)

		if err := query.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to fetch operation logs",
			})
			return
		}

		items = make([]OperationLogItem, 0, len(logs))
		for _, log := range logs {
			items = append(items, OperationLogItem{
				ID:            log.ID,
				LogType:       "operation",
				Timestamp:     log.StartTime,
				User:          log.Operator,
				ActionType:    log.OperationName,
				Details:       log.Details,
				Status:        operationLogStatus(log.Status),
				ErrorMessage:  log.ErrorMessage,
				DurationMs:    log.DurationMs,
				OperationType: log.OperationType,
				OperationName: log.OperationName,
				Operator:      log.Operator,
				StartTime:     log.StartTime,
				EndTime:       log.EndTime,
			})
		}
	}

	// 统计信息（meta.summary）
	var totalOperation int64
	models.DB.Model(&models.OperationLog{}).Count(&totalOperation)
	var totalAudit int64
	models.DB.Model(&models.AuditLog{}).Count(&totalAudit)

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"count":   len(items),
		"meta": gin.H{
			"pagination": gin.H{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": totalPages,
				"has_next":    page < int(totalPages),
				"has_prev":    page > 1,
			},
			"summary": gin.H{
				"total_logs":      totalOperation + totalAudit,
				"total_audit":     totalAudit,
				"total_operation": totalOperation,
			},
		},
	})
}
