package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"etf-insight/infrastructure/cache"
	"etf-insight/infrastructure/messagequeue"
	"etf-insight/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AsyncAnalyticsHandler 异步分析处理器
type AsyncAnalyticsHandler struct {
	rabbitmq *messagequeue.RabbitMQ
	redis    *cache.RedisClient
	db       *models.DB
}

// NewAsyncAnalyticsHandler 创建处理器
func NewAsyncAnalyticsHandler(rabbitmq *messagequeue.RabbitMQ, redis *cache.RedisClient, db *models.DB) *AsyncAnalyticsHandler {
	return &AsyncAnalyticsHandler{
		rabbitmq: rabbitmq,
		redis:    redis,
		db:       db,
	}
}

// SubmitRiskAnalysisRequest 提交风险分析请求
type SubmitRiskAnalysisRequest struct {
	Symbol          string  `json:"symbol" binding:"required"`
	Period          string  `json:"period" default:"1y"`
	RiskFreeRate    float64 `json:"risk_free_rate" default:"0.04"`
	BenchmarkSymbol string  `json:"benchmark_symbol"`
}

// SubmitFactorAnalysisRequest 提交因子分析请求
type SubmitFactorAnalysisRequest struct {
	Symbol string `json:"symbol" binding:"required"`
}

// SubmitOverlapAnalysisRequest 提交重叠分析请求
type SubmitOverlapAnalysisRequest struct {
	ETF1 string `json:"etf1" binding:"required"`
	ETF2 string `json:"etf2" binding:"required"`
}

// SubmitBacktestRequest 提交回测请求
type SubmitBacktestRequest struct {
	Portfolio []struct {
		Symbol string  `json:"symbol" binding:"required"`
		Weight float64 `json:"weight" binding:"required"`
	} `json:"portfolio" binding:"required,min=1"`
	Config struct {
		StartDate      string  `json:"start_date" binding:"required"`
		EndDate        string  `json:"end_date" binding:"required"`
		InitialCapital float64 `json:"initial_capital" default:"100000"`
		RebalanceFreq  string  `json:"rebalance_freq" default:"monthly"`
		CommissionRate float64 `json:"commission_rate" default:"0.001"`
	} `json:"config"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	TaskID    string    `json:"task_id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskResultResponse 任务结果响应
type TaskResultResponse struct {
	TaskID      string          `json:"task_id"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       string          `json:"error,omitempty"`
	Duration    int64           `json:"duration_ms,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// SubmitRiskAnalysis 提交风险分析任务
func (h *AsyncAnalyticsHandler) SubmitRiskAnalysis(c *gin.Context) {
	var req SubmitRiskAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID := uuid.New().String()
	payload, _ := json.Marshal(req)

	task := messagequeue.NewMessage("risk_analysis", payload)
	task.ID = taskID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.rabbitmq.PublishJSON(ctx, messagequeue.ExchangeAnalytics, messagequeue.RoutingKeyRiskAnalysis, task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit task"})
		return
	}

	// 缓存任务状态
	h.redis.SetJSON(ctx, getTaskCacheKey(taskID), TaskResponse{
		TaskID:    taskID,
		Type:      "risk_analysis",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	}, 24*time.Hour)

	c.JSON(http.StatusAccepted, TaskResponse{
		TaskID:    taskID,
		Type:      "risk_analysis",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	})
}

// SubmitFactorAnalysis 提交因子分析任务
func (h *AsyncAnalyticsHandler) SubmitFactorAnalysis(c *gin.Context) {
	var req SubmitFactorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID := uuid.New().String()
	payload, _ := json.Marshal(req)

	task := messagequeue.NewMessage("factor_analysis", payload)
	task.ID = taskID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.rabbitmq.PublishJSON(ctx, messagequeue.ExchangeAnalytics, messagequeue.RoutingKeyFactorAnalysis, task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit task"})
		return
	}

	h.redis.SetJSON(ctx, getTaskCacheKey(taskID), TaskResponse{
		TaskID:    taskID,
		Type:      "factor_analysis",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	}, 24*time.Hour)

	c.JSON(http.StatusAccepted, TaskResponse{
		TaskID:    taskID,
		Type:      "factor_analysis",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	})
}

// SubmitOverlapAnalysis 提交重叠分析任务
func (h *AsyncAnalyticsHandler) SubmitOverlapAnalysis(c *gin.Context) {
	var req SubmitOverlapAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID := uuid.New().String()
	payload, _ := json.Marshal(req)

	task := messagequeue.NewMessage("overlap_analysis", payload)
	task.ID = taskID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.rabbitmq.PublishJSON(ctx, messagequeue.ExchangeAnalytics, "analytics.overlap", task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit task"})
		return
	}

	h.redis.SetJSON(ctx, getTaskCacheKey(taskID), TaskResponse{
		TaskID:    taskID,
		Type:      "overlap_analysis",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	}, 24*time.Hour)

	c.JSON(http.StatusAccepted, TaskResponse{
		TaskID:    taskID,
		Type:      "overlap_analysis",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	})
}

// SubmitBacktest 提交回测任务
func (h *AsyncAnalyticsHandler) SubmitBacktest(c *gin.Context) {
	var req SubmitBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID := uuid.New().String()
	payload, _ := json.Marshal(req)

	task := messagequeue.NewMessage("backtest", payload)
	task.ID = taskID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.rabbitmq.PublishJSON(ctx, messagequeue.ExchangeAnalytics, messagequeue.RoutingKeyBacktest, task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit task"})
		return
	}

	h.redis.SetJSON(ctx, getTaskCacheKey(taskID), TaskResponse{
		TaskID:    taskID,
		Type:      "backtest",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	}, 24*time.Hour)

	c.JSON(http.StatusAccepted, TaskResponse{
		TaskID:    taskID,
		Type:      "backtest",
		Status:    "pending",
		Message:   "Task submitted successfully",
		CreatedAt: time.Now(),
	})
}

// GetTaskStatus 获取任务状态
func (h *AsyncAnalyticsHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	ctx := context.Background()

	// 先从Redis获取
	var result TaskResultResponse
	if err := h.redis.GetJSON(ctx, getResultCacheKey(taskID), &result); err == nil {
		c.JSON(http.StatusOK, result)
		return
	}

	// 如果Redis没有，返回pending状态
	var taskStatus TaskResponse
	if err := h.redis.GetJSON(ctx, getTaskCacheKey(taskID), &taskStatus); err == nil {
		c.JSON(http.StatusOK, taskStatus)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}

// GetTaskResult 获取任务结果
func (h *AsyncAnalyticsHandler) GetTaskResult(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	ctx := context.Background()

	var result TaskResultResponse
	if err := h.redis.GetJSON(ctx, getResultCacheKey(taskID), &result); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Result not found or task not completed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListPendingTasks 列出待处理任务
func (h *AsyncAnalyticsHandler) ListPendingTasks(c *gin.Context) {
	ctx := context.Background()

	// 获取所有任务key
	keys, err := h.redis.Keys(ctx, "analytics:task:*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	tasks := make([]TaskResponse, 0)
	for _, key := range keys {
		var task TaskResponse
		if err := h.redis.GetJSON(ctx, key, &task); err == nil && task.Status == "pending" {
			tasks = append(tasks, task)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// CancelTask 取消任务
func (h *AsyncAnalyticsHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	ctx := context.Background()

	// 更新任务状态为cancelled
	h.redis.SetJSON(ctx, getTaskCacheKey(taskID), TaskResponse{
		TaskID:    taskID,
		Status:    "cancelled",
		Message:   "Task cancelled by user",
		CreatedAt: time.Now(),
	}, 24*time.Hour)

	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled"})
}

// Helper functions
func getTaskCacheKey(taskID string) string {
	return "analytics:task:" + taskID
}

func getResultCacheKey(taskID string) string {
	return "analytics:result:" + taskID
}
