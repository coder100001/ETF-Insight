package handlers

import (
	"net/http"

	"etf-insight/tasks"

	"github.com/gin-gonic/gin"
)

// SchedulerHandler 调度器处理器
type SchedulerHandler struct {
	scheduler *tasks.Scheduler
}

// NewSchedulerHandler 创建新的调度器处理器
func NewSchedulerHandler(scheduler *tasks.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler: scheduler,
	}
}

// GetJobs 获取所有定时任务
func (h *SchedulerHandler) GetJobs(c *gin.Context) {
	jobs := h.scheduler.GetJobs()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    jobs,
	})
}

// RunOnce 立即执行一次更新
func (h *SchedulerHandler) RunOnce(c *gin.Context) {
	go h.scheduler.RunOnce()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "update task started",
	})
}
