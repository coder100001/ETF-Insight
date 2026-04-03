package tasks

import (
	"context"
	"time"

	"etf-insight/services"
	"etf-insight/utils"

	"github.com/robfig/cron/v3"
)

// ExchangeRateTask 汇率同步定时任务
type ExchangeRateTask struct {
	cron        *cron.Cron
	syncService *services.ExchangeRateSyncService
	isRunning   bool
}

// NewExchangeRateTask 创建汇率同步任务
func NewExchangeRateTask() *ExchangeRateTask {
	return &ExchangeRateTask{
		cron:        cron.New(cron.WithSeconds()),
		syncService: services.NewExchangeRateSyncService(),
	}
}

// Start 启动定时任务
func (t *ExchangeRateTask) Start() {
	if t.isRunning {
		return
	}

	// 每5分钟执行一次汇率更新
	_, err := t.cron.AddFunc("0 */5 * * * *", func() {
		t.runFrequentSync()
	})
	if err != nil {
		utils.Error("Failed to add frequent sync cron job", err)
		return
	}

	// 每日上午10:30执行全量同步
	_, err = t.cron.AddFunc("0 30 10 * * *", func() {
		t.runDailySync()
	})
	if err != nil {
		utils.Error("Failed to add daily sync cron job", err)
		return
	}

	t.cron.Start()
	t.isRunning = true
	utils.Info("Exchange rate sync task started", map[string]interface{}{
		"frequent_interval": "5m",
		"daily_sync":        "10:30",
	})
}

// Stop 停止定时任务
func (t *ExchangeRateTask) Stop() {
	if !t.isRunning {
		return
	}
	t.cron.Stop()
	t.isRunning = false
	utils.Info("Exchange rate sync task stopped", nil)
}

// runFrequentSync 执行高频同步（每5分钟）
func (t *ExchangeRateTask) runFrequentSync() {
	utils.Info("Starting frequent exchange rate sync (5min interval)", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := t.syncService.IncrementalSync(ctx)
	if err != nil {
		utils.Error("Frequent sync failed", err)
		return
	}

	utils.Info("Frequent sync completed", map[string]interface{}{
		"batch_id":    result.BatchID,
		"total":       result.TotalCount,
		"success":     result.SuccessCount,
		"failed":      result.FailedCount,
		"skipped":     result.SkippedCount,
		"duration_ms": result.Duration.Milliseconds(),
	})
}

// runDailySync 执行每日同步
func (t *ExchangeRateTask) runDailySync() {
	utils.Info("Starting daily exchange rate sync", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := t.syncService.FullSync(ctx)
	if err != nil {
		utils.Error("Daily sync failed", err)
		return
	}

	utils.Info("Daily sync completed", map[string]interface{}{
		"batch_id":    result.BatchID,
		"total":       result.TotalCount,
		"success":     result.SuccessCount,
		"failed":      result.FailedCount,
		"skipped":     result.SkippedCount,
		"duration_ms": result.Duration.Milliseconds(),
	})
}

// TriggerManualSync 触发手动同步
func (t *ExchangeRateTask) TriggerManualSync(syncType string) (*services.SyncResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if syncType == "full" {
		return t.syncService.FullSync(ctx)
	}
	return t.syncService.IncrementalSync(ctx)
}

// IsRunning 检查任务是否正在运行
func (t *ExchangeRateTask) IsRunning() bool {
	return t.isRunning
}
