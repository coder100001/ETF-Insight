package services

import (
	"etf-insight/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// UnifiedLog 统一日志模型，用于前端展示
type UnifiedLog struct {
	ID           uint      `json:"id"`
	LogType      string    `json:"log_type"`      // "audit" 或 "operation"
	Timestamp    time.Time `json:"timestamp"`     // CreatedAt 或 StartTime
	User         string    `json:"user"`          // Username 或 Operator
	Module       string    `json:"module"`        // Resource 或 OperationName
	ActionType   string    `json:"action_type"`   // Action 或 OperationType
	Details      string    `json:"details"`       // Path/Method 或 Details
	IP           string    `json:"ip"`            // IP地址
	Status       string    `json:"status"`        // "success" 或 "failure"
	StatusCode   int       `json:"status_code"`   // HTTP状态码或Operation状态
	ErrorMessage string    `json:"error_message"` // 错误信息
	Duration     int       `json:"duration_ms"`   // 操作时长(ms)
}

// LogFilterParams 日志筛选参数
type LogFilterParams struct {
	models.PaginationQuery
	StartTime  *time.Time `json:"start_time" form:"start_time"`
	EndTime    *time.Time `json:"end_time" form:"end_time"`
	User       string     `json:"user" form:"user"`
	ActionType string     `json:"action_type" form:"action_type"`
	Status     string     `json:"status" form:"status"`     // "success" 或 "failure"
	LogType    string     `json:"log_type" form:"log_type"` // "audit" 或 "operation" 或 ""（全部）
}

// LogQueryResult 日志查询结果
type LogQueryResult struct {
	Logs       []UnifiedLog
	TotalAudit int64
	TotalOp    int64
	Total      int64
}

// OperationLogsService 操作日志服务
type OperationLogsService struct {
	db *gorm.DB
}

// NewOperationLogsService 创建操作日志服务实例
func NewOperationLogsService(db *gorm.DB) *OperationLogsService {
	return &OperationLogsService{db: db}
}

// QueryLogs 查询日志
func (s *OperationLogsService) QueryLogs(params LogFilterParams) (*LogQueryResult, error) {
	// 查询AuditLog
	auditLogs, auditTotal, err := s.queryAuditLogs(params)
	if err != nil {
		return nil, fmt.Errorf("查询AuditLog失败: %w", err)
	}

	// 查询OperationLog
	operationLogs, opTotal, err := s.queryOperationLogs(params)
	if err != nil {
		return nil, fmt.Errorf("查询OperationLog失败: %w", err)
	}

	// 合并结果
	allLogs := append(auditLogs, operationLogs...)

	// 根据时间戳排序（新到旧）
	s.sortLogsByTimestampDesc(allLogs)

	// 应用分页
	offset := params.GetOffset()
	limit := params.GetLimit()
	total := auditTotal + opTotal

	if offset >= len(allLogs) {
		return &LogQueryResult{
			Logs:       []UnifiedLog{},
			TotalAudit: auditTotal,
			TotalOp:    opTotal,
			Total:      total,
		}, nil
	}

	end := offset + limit
	if end > len(allLogs) {
		end = len(allLogs)
	}

	pagedLogs := allLogs[offset:end]

	return &LogQueryResult{
		Logs:       pagedLogs,
		TotalAudit: auditTotal,
		TotalOp:    opTotal,
		Total:      total,
	}, nil
}

// queryAuditLogs 查询审计日志
func (s *OperationLogsService) queryAuditLogs(params LogFilterParams) ([]UnifiedLog, int64, error) {
	// 如果指定了log_type且不是"audit"，跳过查询
	if params.LogType != "" && params.LogType != "audit" {
		return []UnifiedLog{}, 0, nil
	}

	query := s.db.Model(&models.AuditLog{})

	// 时间筛选
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}

	// 用户筛选
	if params.User != "" {
		query = query.Where("username LIKE ?", "%"+params.User+"%")
	}

	// 操作类型筛选
	if params.ActionType != "" {
		query = query.Where("action = ?", params.ActionType)
	}

	// 状态筛选
	if params.Status != "" {
		if params.Status == "success" {
			query = query.Where("status_code >= 200 AND status_code < 300")
		} else if params.Status == "failure" {
			query = query.Where("status_code >= 400")
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 如果总数0，直接返回
	if total == 0 {
		return []UnifiedLog{}, 0, nil
	}

	// 查询数据
	var auditLogs []models.AuditLog
	err := query.Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&auditLogs).Error

	if err != nil {
		return nil, 0, err
	}

	// 转换为UnifiedLog
	unifiedLogs := make([]UnifiedLog, 0, len(auditLogs))
	for _, log := range auditLogs {
		status := "failure"
		if log.StatusCode >= 200 && log.StatusCode < 300 {
			status = "success"
		}

		unifiedLogs = append(unifiedLogs, UnifiedLog{
			ID:           log.ID,
			LogType:      "audit",
			Timestamp:    log.CreatedAt,
			User:         log.Username,
			Module:       log.Resource,
			ActionType:   log.Action,
			Details:      fmt.Sprintf("%s %s", log.Method, log.Path),
			IP:           log.IP,
			Status:       status,
			StatusCode:   log.StatusCode,
			ErrorMessage: log.Error,
			Duration:     0, // AuditLog没有时长字段
		})
	}

	return unifiedLogs, total, nil
}

// queryOperationLogs 查询操作日志
func (s *OperationLogsService) queryOperationLogs(params LogFilterParams) ([]UnifiedLog, int64, error) {
	// 如果指定了log_type且不是"operation"，跳过查询
	if params.LogType != "" && params.LogType != "operation" {
		return []UnifiedLog{}, 0, nil
	}

	query := s.db.Model(&models.OperationLog{})

	// 时间筛选
	if params.StartTime != nil {
		query = query.Where("start_time >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("start_time <= ?", *params.EndTime)
	}

	// 用户筛选
	if params.User != "" {
		query = query.Where("operator LIKE ?", "%"+params.User+"%")
	}

	// 操作类型筛选
	if params.ActionType != "" {
		query = query.Where("operation_type = ?", params.ActionType)
	}

	// 状态筛选
	if params.Status != "" {
		var statusValue int
		if params.Status == "success" {
			statusValue = 1 // 成功
		} else if params.Status == "failure" {
			statusValue = 2 // 失败
		}
		query = query.Where("status = ?", statusValue)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 如果总数0，直接返回
	if total == 0 {
		return []UnifiedLog{}, 0, nil
	}

	// 查询数据
	var operationLogs []models.OperationLog
	err := query.Order("start_time DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&operationLogs).Error

	if err != nil {
		return nil, 0, err
	}

	// 转换为UnifiedLog
	unifiedLogs := make([]UnifiedLog, 0, len(operationLogs))
	for _, log := range operationLogs {
		status := "failure"
		if log.Status == 1 {
			status = "success"
		}

		unifiedLogs = append(unifiedLogs, UnifiedLog{
			ID:           log.ID,
			LogType:      "operation",
			Timestamp:    log.StartTime,
			User:         log.Operator,
			Module:       log.OperationName,
			ActionType:   log.OperationType,
			Details:      log.Details,
			IP:           "", // OperationLog没有IP字段
			Status:       status,
			StatusCode:   log.Status,
			ErrorMessage: log.ErrorMessage,
			Duration:     log.DurationMs,
		})
	}

	return unifiedLogs, total, nil
}

// sortLogsByTimestampDesc 按时间戳降序排序
func (s *OperationLogsService) sortLogsByTimestampDesc(logs []UnifiedLog) {
	// 使用简单的冒泡排序（数量不会太大）
	for i := 0; i < len(logs); i++ {
		for j := i + 1; j < len(logs); j++ {
			if logs[i].Timestamp.Before(logs[j].Timestamp) {
				logs[i], logs[j] = logs[j], logs[i]
			}
		}
	}
}

// GetLogTypes 获取日志类型统计
func (s *OperationLogsService) GetLogTypes() (map[string]int64, error) {
	var auditCount int64
	if err := s.db.Model(&models.AuditLog{}).Count(&auditCount).Error; err != nil {
		return nil, err
	}

	var opCount int64
	if err := s.db.Model(&models.OperationLog{}).Count(&opCount).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"audit":     auditCount,
		"operation": opCount,
	}, nil
}

// GetActionTypes 获取操作类型列表
func (s *OperationLogsService) GetActionTypes() ([]string, error) {
	// AuditLog的Action类型
	var auditActions []string
	if err := s.db.Model(&models.AuditLog{}).
		Select("DISTINCT action").
		Where("action != ''").
		Pluck("action", &auditActions).Error; err != nil {
		return nil, err
	}

	// OperationLog的OperationType类型
	var opTypes []string
	if err := s.db.Model(&models.OperationLog{}).
		Select("DISTINCT operation_type").
		Where("operation_type != ''").
		Pluck("operation_type", &opTypes).Error; err != nil {
		return nil, err
	}

	// 合并去重
	allTypes := make(map[string]bool)
	for _, action := range auditActions {
		allTypes[action] = true
	}
	for _, opType := range opTypes {
		allTypes[opType] = true
	}

	result := make([]string, 0, len(allTypes))
	for action := range allTypes {
		result = append(result, action)
	}

	return result, nil
}

// GetUsers 获取用户列表
func (s *OperationLogsService) GetUsers() ([]string, error) {
	// AuditLog的用户
	var auditUsers []string
	if err := s.db.Model(&models.AuditLog{}).
		Select("DISTINCT username").
		Where("username != ''").
		Pluck("username", &auditUsers).Error; err != nil {
		return nil, err
	}

	// OperationLog的操作者
	var opUsers []string
	if err := s.db.Model(&models.OperationLog{}).
		Select("DISTINCT operator").
		Where("operator != ''").
		Pluck("operator", &opUsers).Error; err != nil {
		return nil, err
	}

	// 合并去重
	allUsers := make(map[string]bool)
	for _, user := range auditUsers {
		allUsers[user] = true
	}
	for _, user := range opUsers {
		allUsers[user] = true
	}

	result := make([]string, 0, len(allUsers))
	for user := range allUsers {
		result = append(result, user)
	}

	return result, nil
}
