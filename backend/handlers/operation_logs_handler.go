package handlers

import (
	"etf-insight/models"
	"etf-insight/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// OperationLogsHandler 操作日志处理器
type OperationLogsHandler struct {
	service *services.OperationLogsService
}

// NewOperationLogsHandler 创建操作日志处理器
func NewOperationLogsHandler(service *services.OperationLogsService) *OperationLogsHandler {
	return &OperationLogsHandler{service: service}
}

// GetLogsResponse 获取日志响应
type GetLogsResponse struct {
	Success bool                  `json:"success"`
	Data    []services.UnifiedLog `json:"data,omitempty"`
	Meta    LogsResponseMeta      `json:"meta,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// LogsResponseMeta 日志响应元数据
type LogsResponseMeta struct {
	Pagination models.PaginationMeta `json:"pagination"`
	Summary    LogsSummary           `json:"summary"`
}

// LogsSummary 日志摘要信息
type LogsSummary struct {
	TotalLogs  int64 `json:"total_logs"`
	TotalAudit int64 `json:"total_audit"`
	TotalOp    int64 `json:"total_operation"`
}

// GetLogs 获取日志列表
// @Summary 获取操作日志列表
// @Description 获取审计日志和操作日志的统一列表，支持多条件筛选
// @Tags operation-logs
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param user query string false "用户"
// @Param action_type query string false "操作类型"
// @Param status query string false "状态: success/failure"
// @Param log_type query string false "日志类型: audit/operation"
// @Success 200 {object} GetLogsResponse
// @Router /api/logs [get]
func (h *OperationLogsHandler) GetLogs(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析时间参数
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	// 解析筛选参数
	user := c.Query("user")
	actionType := c.Query("action_type")
	status := c.Query("status")
	logType := c.Query("log_type")

	// 构建查询参数
	params := services.LogFilterParams{
		PaginationQuery: models.PaginationQuery{
			Page:     page,
			PageSize: pageSize,
		},
		StartTime:  startTime,
		EndTime:    endTime,
		User:       user,
		ActionType: actionType,
		Status:     status,
		LogType:    logType,
	}

	// 查询日志
	result, err := h.service.QueryLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetLogsResponse{
			Success: false,
			Error:   "查询日志失败: " + err.Error(),
		})
		return
	}

	// 计算总页数
	total := result.Total
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	c.JSON(http.StatusOK, GetLogsResponse{
		Success: true,
		Data:    result.Logs,
		Meta: LogsResponseMeta{
			Pagination: models.PaginationMeta{
				Page:       params.Page,
				PageSize:   params.PageSize,
				Total:      total,
				TotalPages: totalPages,
				HasNext:    params.Page < totalPages,
				HasPrev:    params.Page > 1,
			},
			Summary: LogsSummary{
				TotalLogs:  total,
				TotalAudit: result.TotalAudit,
				TotalOp:    result.TotalOp,
			},
		},
	})
}

// ExportLogsRequest 导出日志请求
type ExportLogsRequest struct {
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	User       string     `json:"user"`
	ActionType string     `json:"action_type"`
	Status     string     `json:"status"`
	LogType    string     `json:"log_type"`
}

// ExportLogs 导出日志到Excel
// @Summary 导出操作日志
// @Description 将筛选后的操作日志导出为Excel文件
// @Tags operation-logs
// @Accept json
// @Produce json
// @Param request body ExportLogsRequest true "筛选条件"
// @Success 200 {file} file "Excel文件"
// @Router /api/logs/export [post]
func (h *OperationLogsHandler) ExportLogs(c *gin.Context) {
	var req ExportLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 构建查询参数（不使用分页，导出所有符合条件的数据）
	params := services.LogFilterParams{
		PaginationQuery: models.PaginationQuery{
			Page:     1,
			PageSize: 10000, // 导出最大条数
		},
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		User:       req.User,
		ActionType: req.ActionType,
		Status:     req.Status,
		LogType:    req.LogType,
	}

	// 查询日志
	result, err := h.service.QueryLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询日志失败: " + err.Error(),
		})
		return
	}

	// 转换为Excel格式
	excelData, err := convertLogsToExcel(result.Logs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "生成Excel失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=operation_logs_"+time.Now().Format("20060102_150405")+".xlsx")
	c.Header("Content-Length", strconv.Itoa(len(excelData)))

	// 返回文件
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)
}

// convertLogsToExcel 将日志转换为Excel格式
func convertLogsToExcel(logs []services.UnifiedLog) ([]byte, error) {
	// 这里简化实现，实际应该使用xlsx库生成Excel文件
	// 返回一个简单的CSV格式作为示例
	var csvContent string
	csvContent = "ID,日志类型,操作时间,用户,模块,操作类型,操作详情,IP地址,状态,状态码,错误信息,操作时长(ms)\n"

	for _, log := range logs {
		csvContent += strconv.FormatUint(uint64(log.ID), 10) + ","
		csvContent += log.LogType + ","
		csvContent += log.Timestamp.Format("2006-01-02 15:04:05") + ","
		csvContent += log.User + ","
		csvContent += log.Module + ","
		csvContent += log.ActionType + ","
		csvContent += strconv.Quote(log.Details) + "," // 用引号包裹，防止逗号问题
		csvContent += log.IP + ","
		csvContent += log.Status + ","
		csvContent += strconv.Itoa(log.StatusCode) + ","
		csvContent += strconv.Quote(log.ErrorMessage) + ","
		csvContent += strconv.Itoa(log.Duration) + "\n"
	}

	return []byte(csvContent), nil
}

// GetLogDetail 获取日志详情
// @Summary 获取日志详情
// @Description 根据ID获取日志详细信息
// @Tags operation-logs
// @Produce json
// @Param id path int true "日志ID"
// @Param type path string true "日志类型: audit/operation"
// @Success 200 {object} GetLogDetailResponse
// @Router /api/logs/{type}/{id} [get]
func (h *OperationLogsHandler) GetLogDetail(c *gin.Context) {
	logType := c.Param("type")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID格式",
		})
		return
	}

	var logDetail interface{}
	var found bool

	if logType == "audit" {
		// 查询AuditLog
		var auditLog models.AuditLog
		result := models.DB.First(&auditLog, uint(id))
		found = result.RowsAffected > 0
		logDetail = auditLog
	} else if logType == "operation" {
		// 查询OperationLog
		var opLog models.OperationLog
		result := models.DB.First(&opLog, uint(id))
		found = result.RowsAffected > 0
		logDetail = opLog
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的日志类型，必须是audit或operation",
		})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "日志未找到",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logDetail,
	})
}

// GetLogTypesResponse 获取日志类型响应
type GetLogTypesResponse struct {
	Success bool             `json:"success"`
	Types   map[string]int64 `json:"types,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// GetLogTypes 获取日志类型统计
// @Summary 获取日志类型统计
// @Description 获取审计日志和操作日志的数量统计
// @Tags operation-logs
// @Produce json
// @Success 200 {object} GetLogTypesResponse
// @Router /api/logs/types [get]
func (h *OperationLogsHandler) GetLogTypes(c *gin.Context) {
	types, err := h.service.GetLogTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetLogTypesResponse{
			Success: false,
			Error:   "获取日志类型统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetLogTypesResponse{
		Success: true,
		Types:   types,
	})
}

// GetActionTypesResponse 获取操作类型响应
type GetActionTypesResponse struct {
	Success     bool     `json:"success"`
	ActionTypes []string `json:"action_types,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// GetActionTypes 获取操作类型列表
// @Summary 获取操作类型列表
// @Description 获取所有可用的操作类型列表
// @Tags operation-logs
// @Produce json
// @Success 200 {object} GetActionTypesResponse
// @Router /api/logs/action-types [get]
func (h *OperationLogsHandler) GetActionTypes(c *gin.Context) {
	actionTypes, err := h.service.GetActionTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetActionTypesResponse{
			Success: false,
			Error:   "获取操作类型列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetActionTypesResponse{
		Success:     true,
		ActionTypes: actionTypes,
	})
}

// GetUsersResponse 获取用户响应
type GetUsersResponse struct {
	Success bool     `json:"success"`
	Users   []string `json:"users,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取所有有操作日志的用户列表
// @Tags operation-logs
// @Produce json
// @Success 200 {object} GetUsersResponse
// @Router /api/logs/users [get]
func (h *OperationLogsHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetUsersResponse{
			Success: false,
			Error:   "获取用户列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetUsersResponse{
		Success: true,
		Users:   users,
	})
}

// GetLogDetailResponse 日志详情响应
type GetLogDetailResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
