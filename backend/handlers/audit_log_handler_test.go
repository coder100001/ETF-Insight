package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// operationLogsResponse 对齐 /api/logs 新响应契约的测试结构
type operationLogsResponse struct {
	Success bool               `json:"success"`
	Data    []OperationLogItem `json:"data"`
	Count   int                `json:"count"`
	Meta    struct {
		Pagination struct {
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			Total      int64 `json:"total"`
			TotalPages int64 `json:"total_pages"`
			HasNext    bool  `json:"has_next"`
			HasPrev    bool  `json:"has_prev"`
		} `json:"pagination"`
		Summary struct {
			TotalLogs      int64 `json:"total_logs"`
			TotalAudit     int64 `json:"total_audit"`
			TotalOperation int64 `json:"total_operation"`
		} `json:"summary"`
	} `json:"meta"`
}

func TestGetOperationLogs(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 插入一条操作日志
	log := models.OperationLog{
		OperationType: "sync",
		OperationName: "测试同步",
		Operator:      "tester",
		Status:        1,
		StartTime:     time.Now(),
	}
	assert.NoError(t, models.DB.Create(&log).Error)

	router := gin.New()
	handler := NewAuditLogHandler()
	router.GET("/api/logs", handler.GetOperationLogs)

	req := httptest.NewRequest("GET", "/api/logs?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp operationLogsResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, 1, resp.Count)
	assert.Equal(t, int64(1), resp.Meta.Pagination.Total)
	assert.Equal(t, "success", resp.Data[0].Status)
	assert.Equal(t, "operation", resp.Data[0].LogType)
	// 统一字段映射
	assert.Equal(t, "测试同步", resp.Data[0].ActionType)
	assert.Equal(t, "tester", resp.Data[0].User)
	assert.Equal(t, "tester", resp.Data[0].Operator)
	// 原始字段向后兼容
	assert.Equal(t, "sync", resp.Data[0].OperationType)
	assert.NotNil(t, resp.Data[0].Timestamp)
	// 派生的统一字段（module/status_code）
	assert.Equal(t, "sync", resp.Data[0].Module)
	assert.Equal(t, 200, resp.Data[0].StatusCode)
	// 统计信息
	assert.Equal(t, int64(1), resp.Meta.Summary.TotalOperation)
	assert.Equal(t, int64(1), resp.Meta.Summary.TotalLogs)
}

func TestGetOperationLogs_WithFilter(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	log := models.OperationLog{
		OperationType: "sync",
		OperationName: "同步任务",
		Operator:      "alice",
		Status:        1,
		StartTime:     time.Now(),
	}
	assert.NoError(t, models.DB.Create(&log).Error)

	router := gin.New()
	handler := NewAuditLogHandler()
	router.GET("/api/logs", handler.GetOperationLogs)

	// user 筛选（映射到 operator）
	req := httptest.NewRequest("GET", "/api/logs?user=alice", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp operationLogsResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Count)

	// 不匹配的筛选
	req2 := httptest.NewRequest("GET", "/api/logs?operator=nobody", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	var resp2 operationLogsResponse
	assert.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp2))
	assert.Equal(t, 0, resp2.Count)
	assert.Equal(t, int64(0), resp2.Meta.Pagination.Total)
}

func TestGetOperationLogs_StatusFilter(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 一条成功(1)一条失败(2)
	assert.NoError(t, models.DB.Create(&models.OperationLog{
		OperationType: "sync",
		Operator:      "alice",
		Status:        1,
		StartTime:     time.Now(),
	}).Error)
	assert.NoError(t, models.DB.Create(&models.OperationLog{
		OperationType: "backtest",
		Operator:      "bob",
		Status:        2,
		StartTime:     time.Now(),
	}).Error)

	router := gin.New()
	handler := NewAuditLogHandler()
	router.GET("/api/logs", handler.GetOperationLogs)

	// 按字符串状态筛选（前端 success/failure → 后端 1/2）
	req := httptest.NewRequest("GET", "/api/logs?status=failure", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp operationLogsResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Count)
	assert.Equal(t, "error", resp.Data[0].Status)
	assert.Equal(t, "backtest", resp.Data[0].OperationType)
	// 失败状态映射为 500
	assert.Equal(t, 500, resp.Data[0].StatusCode)
	assert.Equal(t, "backtest", resp.Data[0].Module)

	// 分页元数据
	assert.Equal(t, 1, resp.Meta.Pagination.Page)
	assert.Equal(t, 20, resp.Meta.Pagination.PageSize)
	assert.Equal(t, int64(1), resp.Meta.Pagination.TotalPages)
	assert.False(t, resp.Meta.Pagination.HasNext)
	assert.False(t, resp.Meta.Pagination.HasPrev)
}

func TestGetOperationLogs_AuditLogType(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 插入一条审计日志
	assert.NoError(t, models.DB.Create(&models.AuditLog{
		Username:   "tester",
		Action:     "GET",
		Resource:   "/api/etf",
		Path:       "/api/etf",
		StatusCode: http.StatusOK,
		CreatedAt:  time.Now(),
	}).Error)

	router := gin.New()
	handler := NewAuditLogHandler()
	router.GET("/api/logs", handler.GetOperationLogs)

	// log_type=audit：暂无查询接口，仅返回统计信息
	req := httptest.NewRequest("GET", "/api/logs?log_type=audit", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp operationLogsResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, 0, resp.Count)
	assert.Equal(t, int64(1), resp.Meta.Summary.TotalAudit)
	assert.Equal(t, int64(1), resp.Meta.Summary.TotalLogs)
}
